package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

type empresaNextcloudProvisionResp struct {
	OK            bool   `json:"ok"`
	EmpresaID     int64  `json:"empresa_id"`
	BaseURL       string `json:"base_url"`
	NextcloudUser string `json:"nextcloud_user"`
	Password      string `json:"password,omitempty"` // solo en provision/reset
	QuotaGB       int64  `json:"quota_gb"`
	WebURL        string `json:"web_url"`
	WebDAVURL     string `json:"webdav_url"`
	UpdatedAt     string `json:"updated_at,omitempty"`
}

func nextcloudQuotaGBForEmpresa(dbSuper *sql.DB) int64 {
	v, _, _, err := getLimitacionInt64(dbSuper, superEmpresaLimitNextcloudMaxGBKey, defaultEmpresaNextcloudMaxGB)
	if err != nil {
		return defaultNextcloudQuotaGBPerEmpresa
	}
	if v < 0 {
		return 0
	}
	if v == 0 {
		// 0 significa "sin cuota" pero el usuario pidió restringir por defecto; mantenemos 1.
		return defaultNextcloudQuotaGBPerEmpresa
	}
	return v
}

func nextcloudBuildEmpresaUserID(empresaID int64) string {
	return fmt.Sprintf("empresa_%d", empresaID)
}

func nextcloudQuotaStringGB(gb int64) string {
	if gb <= 0 {
		return "1 GB"
	}
	return fmt.Sprintf("%d GB", gb)
}

func nextcloudOCSRequest(method, baseURL, path string, adminUser, adminSecret string, form url.Values) (*http.Response, []byte, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	full := baseURL + path
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req, _ := http.NewRequest(method, full, body)
	req.Header.Set("OCS-APIRequest", "true")
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if adminUser != "" || adminSecret != "" {
		req.SetBasicAuth(adminUser, adminSecret)
	}
	client := &http.Client{Timeout: 18 * time.Second}
	res, err := client.Do(req)
	if err != nil || res == nil {
		return nil, nil, fmt.Errorf("no se pudo conectar a Nextcloud")
	}
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 2<<20))
	_ = res.Body.Close()
	return res, raw, nil
}

func nextcloudEnsureUserAndQuota(dbSuper *sql.DB, baseURL, adminUser, adminSecret, userID, password string, quotaGB int64) error {
	quotaStr := nextcloudQuotaStringGB(quotaGB)

	// 1) Crear usuario (si ya existe, Nextcloud devuelve error; luego intentamos set quota)
	form := url.Values{}
	form.Set("userid", userID)
	form.Set("password", password)
	form.Set("quota", quotaStr)
	_, _, _ = nextcloudOCSRequest(http.MethodPost, baseURL, "/ocs/v1.php/cloud/users", adminUser, adminSecret, form)

	// 2) Forzar cuota (idempotente)
	form2 := url.Values{}
	form2.Set("key", "quota")
	form2.Set("value", quotaStr)
	res, raw, err := nextcloudOCSRequest(http.MethodPut, baseURL, "/ocs/v1.php/cloud/users/"+url.PathEscape(userID), adminUser, adminSecret, form2)
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		// Nextcloud suele devolver XML OCS; devolvemos algo útil sin filtrar secretos
		msg := strings.TrimSpace(string(raw))
		if msg == "" {
			msg = "respuesta no OK al ajustar cuota"
		}
		return fmt.Errorf("nextcloud: no se pudo ajustar cuota (http %d): %s", res.StatusCode, msg)
	}
	return nil
}

func EmpresaNextcloudHandler(dbEmpresas *sql.DB, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "" {
			action = "context"
		}

		enabled := true
		if dbSuper != nil {
			enabled = isNextcloudEnabled(dbSuper)
		}

		baseURL, _ := nextcloudResolveBaseURL(dbSuper)
		adminUser, _ := nextcloudResolveAdminUser(dbSuper)
		adminSecret, _ := nextcloudResolveAdminSecret(dbSuper)
		quotaGB := nextcloudQuotaGBForEmpresa(dbSuper)

		// "configured" para la operación real (provision); el context puede funcionar parcial.
		configured := strings.TrimSpace(baseURL) != "" && strings.TrimSpace(adminUser) != "" && strings.TrimSpace(adminSecret) != ""

		switch action {
		case "context":
			acc, ok, _ := dbpkg.GetEmpresaNextcloudAccount(dbEmpresas, empresaID)
			userID := nextcloudBuildEmpresaUserID(empresaID)
			if ok && strings.TrimSpace(acc.NextcloudUser) != "" {
				userID = strings.TrimSpace(acc.NextcloudUser)
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"ok": true,
				"enabled": enabled,
				"configured": configured,
				"empresa_id": empresaID,
				"base_url": baseURL,
				"nextcloud_user": userID,
				"has_credentials": ok,
				"quota_gb": quotaGB,
				"web_url": strings.TrimRight(baseURL, "/"),
				"webdav_url": strings.TrimRight(baseURL, "/") + "/remote.php/dav/files/" + url.PathEscape(userID) + "/",
				"updated_at": acc.UpdatedAt,
			})
			return

		case "provision":
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			if !enabled {
				writeJSON(w, http.StatusOK, map[string]any{
					"ok": false,
					"enabled": false,
					"configured": configured,
					"error": "Nextcloud está desactivado por super administrador.",
				})
				return
			}
			if baseURL == "" || adminUser == "" || adminSecret == "" {
				writeJSON(w, http.StatusOK, map[string]any{
					"ok": false,
					"enabled": enabled,
					"configured": false,
					"error": "Nextcloud no está configurado en super (base_url/admin_user/admin_secret).",
				})
				return
			}
			if !utils.EncryptionAvailable() {
				writeJSON(w, http.StatusOK, map[string]any{
					"ok": false,
					"enabled": enabled,
					"configured": configured,
					"error": "Cifrado requerido: CONFIG_ENC_KEY no está disponible.",
				})
				return
			}

			var payload struct {
				ResetPassword bool `json:"reset_password"`
			}
			if r.Body != nil {
				_ = json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&payload)
			}

			userID := nextcloudBuildEmpresaUserID(empresaID)
			acc, has, _ := dbpkg.GetEmpresaNextcloudAccount(dbEmpresas, empresaID)
			if has && strings.TrimSpace(acc.NextcloudUser) != "" {
				userID = strings.TrimSpace(acc.NextcloudUser)
			}

			plainPass := ""
			if !has || payload.ResetPassword {
				gen, gerr := nextcloudGeneratePassword()
				if gerr != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "No se pudo generar contraseña."})
					return
				}
				plainPass = gen
			}

			// Si ya existe y no pedimos reset, no tocamos contraseña, solo cuota.
			passForCreate := plainPass
			if passForCreate == "" {
				// para "create" necesitamos password; pero si el usuario ya existe, no importa,
				// igual hacemos set cuota. En caso de que NO exista, este flujo fallaría;
				// por eso si no hay cuenta local, siempre generamos.
				passForCreate = "TempPass-NotUsed"
			}

			if err := nextcloudEnsureUserAndQuota(dbSuper, baseURL, adminUser, adminSecret, userID, passForCreate, quotaGB); err != nil {
				writeJSON(w, http.StatusOK, map[string]any{
					"ok": false,
					"enabled": enabled,
					"configured": configured,
					"error": err.Error(),
				})
				return
			}

			updatedAt := ""
			if plainPass != "" {
				enc, encErr := utils.EncryptString(plainPass)
				if encErr != nil {
					writeJSON(w, http.StatusOK, map[string]any{
						"ok": false,
						"enabled": enabled,
						"configured": configured,
						"error": "No se pudo cifrar la contraseña.",
					})
					return
				}
				acc2, err := dbpkg.UpsertEmpresaNextcloudAccount(dbEmpresas, empresaID, userID, enc)
				if err != nil {
					writeJSON(w, http.StatusOK, map[string]any{
						"ok": false,
						"enabled": enabled,
						"configured": configured,
						"error": "No se pudo guardar credenciales de Nextcloud.",
					})
					return
				}
				updatedAt = acc2.UpdatedAt
			}

			resp := empresaNextcloudProvisionResp{
				OK:            true,
				EmpresaID:     empresaID,
				BaseURL:       baseURL,
				NextcloudUser: userID,
				Password:      plainPass,
				QuotaGB:       quotaGB,
				WebURL:        strings.TrimRight(baseURL, "/"),
				WebDAVURL:     strings.TrimRight(baseURL, "/") + "/remote.php/dav/files/" + url.PathEscape(userID) + "/",
				UpdatedAt:     updatedAt,
			}
			writeJSON(w, http.StatusOK, resp)
			return

		default:
			http.Error(w, "action invalida (context, provision)", http.StatusBadRequest)
			return
		}
	}
}

