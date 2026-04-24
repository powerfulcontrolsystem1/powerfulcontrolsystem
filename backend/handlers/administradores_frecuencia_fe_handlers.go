package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/mail"
	"sort"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

const superFrecuenciaFEAdminsConfigKey = "super.frecuencia_fe.admin_emails"

func normalizeEmailLower(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	v = strings.ReplaceAll(v, " ", "")
	return v
}

func isSuperAdminRequest(r *http.Request) bool {
	return strings.EqualFold(strings.TrimSpace(adminRoleFromRequest(r)), "super_administrador")
}

func loadFrecuenciaFEAdminEmails(dbSuper *sql.DB) ([]string, string, string, error) {
	raw, _, updatedBy, updatedAt, err := dbpkg.GetConfigEntry(dbSuper, superFrecuenciaFEAdminsConfigKey)
	if err != nil {
		return nil, "", "", err
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{}, strings.TrimSpace(updatedAt), strings.TrimSpace(updatedBy), nil
	}
	var list []string
	if uerr := json.Unmarshal([]byte(raw), &list); uerr != nil {
		// Si el contenido quedó como string simple (legado), intentar separarlo.
		parts := strings.Split(raw, ",")
		list = make([]string, 0, len(parts))
		for _, p := range parts {
			if e := normalizeEmailLower(p); e != "" {
				list = append(list, e)
			}
		}
	}
	uniq := make(map[string]struct{}, len(list))
	out := make([]string, 0, len(list))
	for _, it := range list {
		e := normalizeEmailLower(it)
		if e == "" {
			continue
		}
		if _, ok := uniq[e]; ok {
			continue
		}
		uniq[e] = struct{}{}
		out = append(out, e)
	}
	sort.Strings(out)
	return out, strings.TrimSpace(updatedAt), strings.TrimSpace(updatedBy), nil
}

func saveFrecuenciaFEAdminEmails(dbSuper *sql.DB, emails []string, actor string) error {
	sort.Strings(emails)
	raw, _ := json.Marshal(emails)
	if err := dbpkg.SetConfigValue(dbSuper, superFrecuenciaFEAdminsConfigKey, string(raw), false); err != nil {
		return err
	}
	_ = dbpkg.SetConfigValue(dbSuper, superFrecuenciaFEAdminsConfigKey+".updated_by", strings.TrimSpace(actor), false)
	return nil
}

// SuperAdministradoresFrecuenciaFEHandler gestiona la lista de emails autorizados para ver la página frecuencia_fp.html.
func SuperAdministradoresFrecuenciaFEHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminEmail := strings.TrimSpace(adminEmailFromRequest(r))
		if adminEmail == "" || adminEmail == "sistema" {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		if !isSuperAdminRequest(r) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		switch r.Method {
		case http.MethodGet:
			list, updatedAt, updatedBy, err := loadFrecuenciaFEAdminEmails(dbSuper)
			if err != nil {
				http.Error(w, "failed to load: "+err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":         true,
				"emails":     list,
				"updated_at": updatedAt,
				"updated_by": updatedBy,
			})
			return

		case http.MethodPut, http.MethodPost:
			var payload struct {
				Emails []string `json:"emails"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
				return
			}
			uniq := make(map[string]struct{}, len(payload.Emails))
			emails := make([]string, 0, len(payload.Emails))
			for _, it := range payload.Emails {
				e := normalizeEmailLower(it)
				if e == "" {
					continue
				}
				if _, err := mail.ParseAddress(e); err != nil {
					http.Error(w, "email inválido: "+e, http.StatusBadRequest)
					return
				}
				if _, ok := uniq[e]; ok {
					continue
				}
				uniq[e] = struct{}{}
				emails = append(emails, e)
			}

			if err := saveFrecuenciaFEAdminEmails(dbSuper, emails, adminEmail); err != nil {
				http.Error(w, "failed to save: "+err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":     true,
				"emails": emails,
			})
			return

		default:
			http.Error(w, "Método no permitido.", http.StatusMethodNotAllowed)
			return
		}
	}
}

// EmpresaFrecuenciaFPAllowedHandler devuelve si el admin actual está autorizado.
func EmpresaFrecuenciaFPAllowedHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminEmail := normalizeEmailLower(adminEmailFromRequest(r))
		if adminEmail == "" || adminEmail == "sistema" {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		list, _, _, err := loadFrecuenciaFEAdminEmails(dbSuper)
		if err != nil {
			http.Error(w, "failed to load: "+err.Error(), http.StatusInternalServerError)
			return
		}
		allowed := false
		for _, e := range list {
			if normalizeEmailLower(e) == adminEmail {
				allowed = true
				break
			}
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":      true,
			"allowed": allowed,
			"email":   adminEmail,
		})
	}
}

