package handlers

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

const (
	nextcloudEnabledKey         = "nextcloud.enabled"
	nextcloudBaseURLKey         = "nextcloud.base_url"
	nextcloudAdminUserKey       = "nextcloud.admin_user"
	nextcloudAdminCredentialKey = "nextcloud.admin_secret" // #nosec G101 -- config key; value is encrypted.
	nextcloudDefaultQuotaMBKey  = "nextcloud.default_quota_mb"
	nextcloudDefaultQuotaMB     = int64(1024)
	nextcloudMaxQuotaMB         = int64(1024 * 1024)
)

type nextcloudOCSMeta struct {
	Status     string `json:"status"`
	StatusCode int    `json:"statuscode"`
	Message    string `json:"message"`
}

type nextcloudOCSEnvelope struct {
	OCS struct {
		Meta nextcloudOCSMeta `json:"meta"`
		Data json.RawMessage  `json:"data"`
	} `json:"ocs"`
}

type nextcloudCompanyAccount struct {
	User          string
	QuotaMB       int64
	Active        bool
	Provisioned   bool
	ProvisionedAt sql.NullTime
}

func nextcloudConfig(dbConn *sql.DB, key string) string {
	if dbConn == nil {
		return ""
	}
	value, _, _, _, err := dbpkg.GetConfigEntry(dbConn, key)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(value)
}

func nextcloudAdminCredential(dbConn *sql.DB) string {
	value, err := getDecryptedConfigValue(dbConn, nextcloudAdminCredentialKey)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(value)
}

func nextcloudEnabled(dbConn *sql.DB) bool {
	value := nextcloudConfig(dbConn, nextcloudEnabledKey)
	return value == "1" || strings.EqualFold(value, "true")
}

func nextcloudQuotaMB(dbConn *sql.DB) int64 {
	value, err := strconv.ParseInt(nextcloudConfig(dbConn, nextcloudDefaultQuotaMBKey), 10, 64)
	if err != nil || value <= 0 || value > nextcloudMaxQuotaMB {
		return nextcloudDefaultQuotaMB
	}
	return value
}

func validateNextcloudBaseURL(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("la URL de Nextcloud debe ser HTTPS y no incluir credenciales, query ni fragmento")
	}
	host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	if host == "" || host == "localhost" {
		return "", fmt.Errorf("el host de Nextcloud no es valido")
	}
	if ip := net.ParseIP(host); ip != nil && (ip.IsLoopback() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsPrivate()) && !strings.EqualFold(strings.TrimSpace(os.Getenv("PCS_NEXTCLOUD_ALLOW_PRIVATE_HOSTS")), "true") {
		return "", fmt.Errorf("las direcciones privadas requieren PCS_NEXTCLOUD_ALLOW_PRIVATE_HOSTS=true")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	return strings.TrimRight(parsed.String(), "/"), nil
}

func newNextcloudHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	transport.ResponseHeaderTimeout = 15 * time.Second
	return &http.Client{
		Transport: transport,
		Timeout:   20 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func nextcloudOCS(ctx context.Context, client *http.Client, method, baseURL, endpoint, adminUser, adminCredential string, form url.Values) (nextcloudOCSMeta, error) {
	var meta nextcloudOCSMeta
	if client == nil {
		client = newNextcloudHTTPClient()
	}
	baseURL, err := validateNextcloudBaseURL(baseURL)
	if err != nil || baseURL == "" {
		return meta, fmt.Errorf("configuracion Nextcloud invalida")
	}
	if !strings.HasPrefix(endpoint, "/ocs/") {
		return meta, fmt.Errorf("endpoint OCS invalido")
	}
	separator := "?"
	if strings.Contains(endpoint, "?") {
		separator = "&"
	}
	requestURL := baseURL + endpoint + separator + "format=json"
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return meta, fmt.Errorf("no se pudo preparar la solicitud Nextcloud")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("OCS-APIRequest", "true")
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.SetBasicAuth(strings.TrimSpace(adminUser), adminCredential)

	response, err := client.Do(req)
	if err != nil {
		return meta, fmt.Errorf("Nextcloud no responde")
	}
	defer response.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(response.Body, 1<<20))
	var envelope nextcloudOCSEnvelope
	if err := decoder.Decode(&envelope); err != nil {
		return meta, fmt.Errorf("Nextcloud devolvio una respuesta invalida")
	}
	meta = envelope.OCS.Meta
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return meta, fmt.Errorf("Nextcloud rechazo la operacion")
	}
	return meta, nil
}

func nextcloudOCSSuccess(meta nextcloudOCSMeta) bool {
	return meta.StatusCode == 100 && strings.EqualFold(strings.TrimSpace(meta.Status), "ok")
}

func newNextcloudTemporaryPassword() (string, error) {
	return utils.GenerateSecureToken(32)
}

func nextcloudUserExists(ctx context.Context, client *http.Client, baseURL, user, adminUser, adminCredential string) (bool, error) {
	meta, err := nextcloudOCS(ctx, client, http.MethodGet, baseURL, "/ocs/v1.php/cloud/users/"+url.PathEscape(user), adminUser, adminCredential, nil)
	if err != nil {
		if meta.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}
	if meta.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if !nextcloudOCSSuccess(meta) {
		return false, fmt.Errorf("Nextcloud no pudo consultar la cuenta")
	}
	return true, nil
}

func ensureNextcloudCompanyAccount(dbEmp *sql.DB, empresaID, quotaMB int64) (nextcloudCompanyAccount, error) {
	var account nextcloudCompanyAccount
	if dbEmp == nil || empresaID <= 0 {
		return account, fmt.Errorf("empresa invalida")
	}
	if err := dbpkg.EnsureEmpresaNextcloudSchema(dbEmp); err != nil {
		return account, err
	}
	user := "pcs_empresa_" + strconv.FormatInt(empresaID, 10)
	if _, err := dbEmp.Exec(`INSERT INTO empresa_nextcloud_accounts (empresa_id, nextcloud_user, quota_mb)
		VALUES ($1,$2,$3) ON CONFLICT (empresa_id) DO NOTHING`, empresaID, user, quotaMB); err != nil {
		return account, err
	}
	err := dbEmp.QueryRow(`SELECT nextcloud_user, quota_mb, activo, provisioned, provisioned_at
		FROM empresa_nextcloud_accounts WHERE empresa_id=$1 LIMIT 1`, empresaID).
		Scan(&account.User, &account.QuotaMB, &account.Active, &account.Provisioned, &account.ProvisionedAt)
	return account, err
}

func auditNextcloudCompanyAction(dbEmp *sql.DB, r *http.Request, empresaID int64, action, result string, status int) {
	_, _ = dbpkg.CreateEmpresaAuditoriaEvento(dbEmp, dbpkg.EmpresaAuditoriaEvento{
		EmpresaID:      empresaID,
		Modulo:         "gestion_documental",
		Accion:         action,
		Recurso:        "nextcloud_account",
		MetodoHTTP:     r.Method,
		Endpoint:       "/api/empresa/nextcloud",
		Resultado:      result,
		CodigoHTTP:     int64(status),
		RequestID:      resolveAuditoriaRequestID(r),
		IPOrigen:       resolveAuditoriaIP(r),
		UserAgent:      strings.TrimSpace(r.UserAgent()),
		UsuarioCreador: adminEmailFromRequest(r),
		Estado:         "activo",
		Observaciones:  "gestion de cuenta Nextcloud empresarial",
	})
}

func nextcloudConfiguration(dbSuper *sql.DB) (baseURL, adminUser, adminCredential string, configured bool) {
	baseURL, _ = validateNextcloudBaseURL(nextcloudConfig(dbSuper, nextcloudBaseURLKey))
	adminUser = nextcloudConfig(dbSuper, nextcloudAdminUserKey)
	adminCredential = nextcloudAdminCredential(dbSuper)
	configured = nextcloudEnabled(dbSuper) && baseURL != "" && adminUser != "" && adminCredential != ""
	return
}

// DeleteNextcloudCompanyAccount removes the remote user and therefore its
// files before the company cascade deletes the local assignment row.
func DeleteNextcloudCompanyAccount(ctx context.Context, dbEmp, dbSuper *sql.DB, empresaID int64) error {
	if dbEmp == nil || empresaID <= 0 {
		return fmt.Errorf("empresa invalida")
	}
	var user string
	var provisioned bool
	err := dbEmp.QueryRow(`SELECT nextcloud_user, provisioned FROM empresa_nextcloud_accounts WHERE empresa_id=$1`, empresaID).Scan(&user, &provisioned)
	if err == sql.ErrNoRows || !provisioned {
		return nil
	}
	if err != nil {
		return err
	}
	baseURL, adminUser, credential, _ := nextcloudConfiguration(dbSuper)
	if baseURL == "" || adminUser == "" || credential == "" {
		return fmt.Errorf("Nextcloud no esta configurado para eliminar la cuenta empresarial")
	}
	meta, err := nextcloudOCS(ctx, newNextcloudHTTPClient(), http.MethodDelete, baseURL, "/ocs/v1.php/cloud/users/"+url.PathEscape(user), adminUser, credential, nil)
	if err != nil && meta.StatusCode != http.StatusNotFound {
		return fmt.Errorf("Nextcloud no pudo eliminar la cuenta empresarial")
	}
	if err == nil && !nextcloudOCSSuccess(meta) {
		return fmt.Errorf("Nextcloud no pudo eliminar la cuenta empresarial")
	}
	return nil
}

// EmpresaNextcloudHandler never accepts company authority from outside the
// validated middleware context and never persists an end-user password.
func EmpresaNextcloudHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID := parseEmpresaIDFromContext(r)
		if empresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		account, err := ensureNextcloudCompanyAccount(dbEmp, empresaID, nextcloudQuotaMB(dbSuper))
		if err != nil {
			http.Error(w, "No se pudo preparar la cuenta Nextcloud", http.StatusInternalServerError)
			return
		}
		baseURL, adminUser, adminCredential, configured := nextcloudConfiguration(dbSuper)
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		temporaryPassword := ""

		switch r.Method {
		case http.MethodGet:
			if action != "" {
				http.Error(w, "accion invalida", http.StatusBadRequest)
				return
			}
		case http.MethodPost:
			switch action {
			case "activate", "deactivate":
				active := action == "activate"
				if _, err := dbEmp.Exec(`UPDATE empresa_nextcloud_accounts SET activo=$1, updated_at=CURRENT_TIMESTAMP WHERE empresa_id=$2`, active, empresaID); err != nil {
					http.Error(w, "No se pudo cambiar el estado de Nextcloud", http.StatusInternalServerError)
					return
				}
				account.Active = active
				auditNextcloudCompanyAction(dbEmp, r, empresaID, action, "ok", http.StatusOK)
			case "provision", "reset_password":
				if !configured {
					http.Error(w, "Nextcloud pendiente de configuracion", http.StatusConflict)
					return
				}
			default:
				http.Error(w, "accion invalida", http.StatusBadRequest)
				return
			}
			if action == "activate" || action == "deactivate" {
				break
			}
			if !configured {
				http.Error(w, "Nextcloud pendiente de configuracion", http.StatusConflict)
				return
			}
			client := newNextcloudHTTPClient()
			switch action {
			case "provision":
				exists, checkErr := nextcloudUserExists(r.Context(), client, baseURL, account.User, adminUser, adminCredential)
				if checkErr != nil {
					auditNextcloudCompanyAction(dbEmp, r, empresaID, "aprovisionar", "error", http.StatusBadGateway)
					http.Error(w, checkErr.Error(), http.StatusBadGateway)
					return
				}
				if !exists {
					temporaryPassword, err = newNextcloudTemporaryPassword()
					if err != nil {
						http.Error(w, "No se pudo generar la credencial temporal", http.StatusInternalServerError)
						return
					}
					meta, createErr := nextcloudOCS(r.Context(), client, http.MethodPost, baseURL, "/ocs/v1.php/cloud/users", adminUser, adminCredential, url.Values{
						"userid": {account.User}, "password": {temporaryPassword}, "quota": {fmt.Sprintf("%d MB", account.QuotaMB)},
					})
					if createErr != nil || !nextcloudOCSSuccess(meta) {
						auditNextcloudCompanyAction(dbEmp, r, empresaID, "aprovisionar", "error", http.StatusBadGateway)
						http.Error(w, "Nextcloud no pudo crear la cuenta", http.StatusBadGateway)
						return
					}
				}
				meta, quotaErr := nextcloudOCS(r.Context(), client, http.MethodPut, baseURL, "/ocs/v1.php/cloud/users/"+url.PathEscape(account.User), adminUser, adminCredential, url.Values{
					"key": {"quota"}, "value": {fmt.Sprintf("%d MB", account.QuotaMB)},
				})
				if quotaErr != nil || !nextcloudOCSSuccess(meta) {
					auditNextcloudCompanyAction(dbEmp, r, empresaID, "aprovisionar", "error", http.StatusBadGateway)
					http.Error(w, "Nextcloud no pudo aplicar la cuota", http.StatusBadGateway)
					return
				}
				result, updateErr := dbEmp.Exec(`UPDATE empresa_nextcloud_accounts
					SET provisioned=TRUE, provisioned_at=COALESCE(provisioned_at,CURRENT_TIMESTAMP), updated_at=CURRENT_TIMESTAMP
					WHERE empresa_id=$1`, empresaID)
				if updateErr != nil {
					http.Error(w, "No se pudo confirmar el aprovisionamiento", http.StatusInternalServerError)
					return
				}
				if rows, rowsErr := result.RowsAffected(); rowsErr != nil || rows != 1 {
					http.Error(w, "No se pudo confirmar la cuenta empresarial", http.StatusConflict)
					return
				}
				account.Provisioned = true
				auditNextcloudCompanyAction(dbEmp, r, empresaID, "aprovisionar", "ok", http.StatusOK)
			case "reset_password":
				if !account.Provisioned {
					http.Error(w, "La cuenta aun no esta aprovisionada", http.StatusConflict)
					return
				}
				temporaryPassword, err = newNextcloudTemporaryPassword()
				if err != nil {
					http.Error(w, "No se pudo generar la credencial temporal", http.StatusInternalServerError)
					return
				}
				meta, resetErr := nextcloudOCS(r.Context(), client, http.MethodPut, baseURL, "/ocs/v1.php/cloud/users/"+url.PathEscape(account.User), adminUser, adminCredential, url.Values{
					"key": {"password"}, "value": {temporaryPassword},
				})
				if resetErr != nil || !nextcloudOCSSuccess(meta) {
					auditNextcloudCompanyAction(dbEmp, r, empresaID, "restablecer_password", "error", http.StatusBadGateway)
					http.Error(w, "Nextcloud no pudo restablecer la credencial", http.StatusBadGateway)
					return
				}
				auditNextcloudCompanyAction(dbEmp, r, empresaID, "restablecer_password", "ok", http.StatusOK)
			default:
				http.Error(w, "accion invalida", http.StatusBadRequest)
				return
			}
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}

		response := map[string]interface{}{
			"ok": true, "empresa_id": empresaID, "nextcloud_user": account.User,
			"quota_mb": account.QuotaMB, "provisioned": account.Provisioned,
			"active":  account.Active,
			"enabled": nextcloudEnabled(dbSuper), "configured": configured,
			"web_url": baseURL,
		}
		if account.Provisioned && baseURL != "" {
			response["webdav_url"] = baseURL + "/remote.php/dav/files/" + url.PathEscape(account.User) + "/"
		}
		if temporaryPassword != "" {
			response["temporary_password"] = temporaryPassword
			response["temporary_password_once"] = true
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, http.StatusOK, response)
	}
}

func NextcloudConfigHandler(dbSuper, dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method == http.MethodGet {
			_, encrypted, _, _, _ := dbpkg.GetConfigEntry(dbSuper, nextcloudAdminCredentialKey)
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok": true, "enabled": nextcloudEnabled(dbSuper),
				"base_url":         nextcloudConfig(dbSuper, nextcloudBaseURLKey),
				"admin_user":       nextcloudConfig(dbSuper, nextcloudAdminUserKey),
				"admin_secret_set": encrypted && nextcloudAdminCredential(dbSuper) != "",
				"default_quota_mb": nextcloudQuotaMB(dbSuper),
			})
			return
		}
		if r.Method != http.MethodPut && r.Method != http.MethodPost {
			w.Header().Set("Allow", "GET, PUT, POST")
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("action")), "test") {
			baseURL, adminUser, credential, configured := nextcloudConfiguration(dbSuper)
			if !configured {
				http.Error(w, "Nextcloud no esta configurado", http.StatusConflict)
				return
			}
			meta, err := nextcloudOCS(r.Context(), newNextcloudHTTPClient(), http.MethodGet, baseURL, "/ocs/v1.php/cloud/capabilities", adminUser, credential, nil)
			if err != nil || !nextcloudOCSSuccess(meta) {
				http.Error(w, "No se pudo validar la conexion con Nextcloud", http.StatusBadGateway)
				return
			}
			writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
			return
		}

		var payload struct {
			Enabled        bool   `json:"enabled"`
			BaseURL        string `json:"base_url"`
			AdminUser      string `json:"admin_user"`
			AdminSecret    string `json:"admin_secret"`
			DefaultQuotaMB int64  `json:"default_quota_mb"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&payload); err != nil {
			http.Error(w, "payload invalido", http.StatusBadRequest)
			return
		}
		baseURL, err := validateNextcloudBaseURL(payload.BaseURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		payload.AdminUser = strings.TrimSpace(payload.AdminUser)
		if payload.DefaultQuotaMB <= 0 {
			payload.DefaultQuotaMB = nextcloudDefaultQuotaMB
		}
		if payload.DefaultQuotaMB > nextcloudMaxQuotaMB {
			http.Error(w, "cuota invalida", http.StatusBadRequest)
			return
		}
		secretAlreadySet := nextcloudAdminCredential(dbSuper) != ""
		if payload.Enabled && (baseURL == "" || payload.AdminUser == "" || (strings.TrimSpace(payload.AdminSecret) == "" && !secretAlreadySet)) {
			http.Error(w, "URL, usuario y secreto son obligatorios para activar Nextcloud", http.StatusBadRequest)
			return
		}
		values := []struct{ key, value string }{
			{nextcloudEnabledKey, strconv.FormatBool(payload.Enabled)},
			{nextcloudBaseURLKey, baseURL},
			{nextcloudAdminUserKey, payload.AdminUser},
			{nextcloudDefaultQuotaMBKey, strconv.FormatInt(payload.DefaultQuotaMB, 10)},
		}
		for _, item := range values {
			if err := dbpkg.SetConfigValue(dbSuper, item.key, item.value, false); err != nil {
				http.Error(w, "No se pudo guardar la configuracion", http.StatusInternalServerError)
				return
			}
		}
		if secret := strings.TrimSpace(payload.AdminSecret); secret != "" {
			if err := dbpkg.SetConfigValue(dbSuper, nextcloudAdminCredentialKey, secret, true); err != nil {
				http.Error(w, "No se pudo cifrar el secreto de Nextcloud", http.StatusInternalServerError)
				return
			}
		}
		if payload.Enabled {
			if _, err := dbpkg.EnsureEmpresaNextcloudAssignmentsForAll(dbEmp, payload.DefaultQuotaMB); err != nil {
				http.Error(w, "No se pudieron asignar los espacios Nextcloud a las empresas", http.StatusInternalServerError)
				return
			}
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}
