package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const nextcloudEnabledKey = "nextcloud.enabled"
const nextcloudBaseURLKey = "nextcloud.base_url"
const nextcloudAdminUserKey = "nextcloud.admin_user"
const nextcloudAdminSecretKey = "nextcloud.admin_secret"
const nextcloudDefaultQuotaMBKey = "nextcloud.default_quota_mb"

func nextcloudConfig(db *sql.DB, key string) string {
	if db == nil {
		return ""
	}
	v, _, _, _, _ := dbpkg.GetConfigEntry(db, key)
	return strings.TrimSpace(v)
}
func nextcloudSecret(db *sql.DB) string {
	v, err := getDecryptedConfigValue(db, nextcloudAdminSecretKey)
	if err == nil && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return ""
}
func nextcloudEnabled(db *sql.DB) bool {
	return strings.EqualFold(nextcloudConfig(db, nextcloudEnabledKey), "true") || nextcloudConfig(db, nextcloudEnabledKey) == "1"
}
func nextcloudQuotaMB(db *sql.DB) int64 {
	v, _ := strconv.ParseInt(nextcloudConfig(db, nextcloudDefaultQuotaMBKey), 10, 64)
	if v <= 0 {
		return 1024
	}
	return v
}

func nextcloudOCS(method, baseURL, path, user, secret string, form url.Values) error {
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequest(method, strings.TrimRight(baseURL, "/")+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("OCS-APIRequest", "true")
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.SetBasicAuth(user, secret)
	res, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return fmt.Errorf("Nextcloud no responde")
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return fmt.Errorf("Nextcloud no acepto la operacion")
	}
	return nil
}

// EmpresaNextcloudHandler exposes only the company assignment; credentials never
// leave the backend. Provisioning is idempotent and always uses the resolved scope.
func EmpresaNextcloudHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, "empresa_id invalido", 400)
			return
		}
		if err = dbpkg.EnsureEmpresaNextcloudSchema(dbEmp); err != nil {
			http.Error(w, "No se pudo preparar Nextcloud", 500)
			return
		}
		var user string
		var quota int64
		var provisioned bool
		err = dbEmp.QueryRow(`SELECT nextcloud_user, quota_mb, provisioned FROM empresa_nextcloud_accounts WHERE empresa_id=$1`, empresaID).Scan(&user, &quota, &provisioned)
		if err != nil {
			http.Error(w, "No se encontro la asignacion empresarial", 404)
			return
		}
		baseURL := strings.TrimRight(nextcloudConfig(dbSuper, nextcloudBaseURLKey), "/")
		configured := nextcloudEnabled(dbSuper) && baseURL != "" && nextcloudConfig(dbSuper, nextcloudAdminUserKey) != "" && nextcloudSecret(dbSuper) != ""
		if strings.EqualFold(r.URL.Query().Get("action"), "provision") {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", 405)
				return
			}
			if !configured {
				http.Error(w, "Nextcloud pendiente de configuracion", 409)
				return
			}
			secret := nextcloudSecret(dbSuper)
			form := url.Values{"userid": {user}, "password": {"Cambiar-antes-de-usar-" + strconv.FormatInt(empresaID, 10)}, "quota": {fmt.Sprintf("%d MB", quota)}}
			_ = nextcloudOCS(http.MethodPost, baseURL, "/ocs/v1.php/cloud/users", nextcloudConfig(dbSuper, nextcloudAdminUserKey), secret, form)
			if err := nextcloudOCS(http.MethodPut, baseURL, "/ocs/v1.php/cloud/users/"+url.PathEscape(user), nextcloudConfig(dbSuper, nextcloudAdminUserKey), secret, url.Values{"key": {"quota"}, "value": {fmt.Sprintf("%d MB", quota)}}); err != nil {
				http.Error(w, err.Error(), 502)
				return
			}
			_, _ = dbEmp.Exec(`UPDATE empresa_nextcloud_accounts SET provisioned=TRUE,updated_at=CURRENT_TIMESTAMP WHERE empresa_id=$1`, empresaID)
			provisioned = true
		}
		writeJSON(w, 200, map[string]interface{}{"ok": true, "empresa_id": empresaID, "nextcloud_user": user, "quota_mb": quota, "provisioned": provisioned, "enabled": nextcloudEnabled(dbSuper), "configured": configured, "web_url": baseURL, "webdav_url": baseURL + "/remote.php/dav/files/" + url.PathEscape(user) + "/"})
	}
}

// NextcloudConfigHandler is super-admin-only through the route wrapper.
func NextcloudConfigHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_, encrypted, _, updated, _ := dbpkg.GetConfigEntry(dbSuper, nextcloudAdminSecretKey)
			writeJSON(w, 200, map[string]interface{}{"ok": true, "enabled": nextcloudEnabled(dbSuper), "base_url": nextcloudConfig(dbSuper, nextcloudBaseURLKey), "admin_user": nextcloudConfig(dbSuper, nextcloudAdminUserKey), "admin_secret_set": encrypted || updated != "", "default_quota_mb": nextcloudQuotaMB(dbSuper)})
			return
		}
		if r.Method != http.MethodPut && r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}
		var p struct {
			Enabled        bool   `json:"enabled"`
			BaseURL        string `json:"base_url"`
			AdminUser      string `json:"admin_user"`
			AdminSecret    string `json:"admin_secret"`
			DefaultQuotaMB int64  `json:"default_quota_mb"`
		}
		if json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&p) != nil {
			http.Error(w, "payload invalido", 400)
			return
		}
		if p.DefaultQuotaMB <= 0 {
			p.DefaultQuotaMB = 1024
		}
		if p.DefaultQuotaMB > 1024*1024 {
			http.Error(w, "cuota invalida", 400)
			return
		}
		if p.BaseURL != "" {
			u, err := url.Parse(p.BaseURL)
			if err != nil || u.Scheme != "https" || u.Host == "" {
				http.Error(w, "La URL debe ser HTTPS valida", 400)
				return
			}
		}
		for k, v := range map[string]string{nextcloudEnabledKey: strconv.FormatBool(p.Enabled), nextcloudBaseURLKey: strings.TrimRight(strings.TrimSpace(p.BaseURL), "/"), nextcloudAdminUserKey: strings.TrimSpace(p.AdminUser), nextcloudDefaultQuotaMBKey: strconv.FormatInt(p.DefaultQuotaMB, 10)} {
			if err := dbpkg.SetConfigValue(dbSuper, k, v, false); err != nil {
				http.Error(w, "No se pudo guardar configuracion", 500)
				return
			}
		}
		if strings.TrimSpace(p.AdminSecret) != "" {
			if err := dbpkg.SetConfigValue(dbSuper, nextcloudAdminSecretKey, strings.TrimSpace(p.AdminSecret), true); err != nil {
				http.Error(w, "No se pudo guardar secreto", 500)
				return
			}
		}
		writeJSON(w, 200, map[string]bool{"ok": true})
	}
}
