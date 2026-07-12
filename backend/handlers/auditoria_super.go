package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

type superAuditoriaUIPayload struct {
	Accion        string                 `json:"accion"`
	Modulo        string                 `json:"modulo"`
	Recurso       string                 `json:"recurso"`
	EmpresaID     int64                  `json:"empresa_id"`
	RecursoID     int64                  `json:"recurso_id"`
	Endpoint      string                 `json:"endpoint"`
	Observaciones string                 `json:"observaciones"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// SuperAuditoriaHandler expone la auditoria global/super y permite eventos UI explicitos.
func SuperAuditoriaHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requesterAdmin, principalEmail, err := resolveRequesterAdminScope(dbSuper, r)
		if err != nil {
			http.Error(w, "failed to resolve admin scope", http.StatusInternalServerError)
			return
		}
		if requesterAdmin == nil {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		superPanelScope := superAuditoriaRequestUsesSuperPanelScope(r)
		if superPanelScope && !utils.IsSuperPanelRole(requesterAdmin.Role) {
			http.Error(w, "auditoria super solo disponible para super administradores", http.StatusForbidden)
			return
		}

		switch r.Method {
		case http.MethodGet:
			filter, err := buildSuperAuditoriaFilterFromRequest(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			effectivePrincipal := superAuditoriaEffectivePrincipalScope(r, requesterAdmin, principalEmail)
			if effectivePrincipal != "" {
				filter.PrincipalEmail = effectivePrincipal
			}

			total, err := dbpkg.CountSuperAuditoriaEventos(dbSuper, filter)
			if err != nil {
				http.Error(w, "No se pudo consultar el total de auditoria global", http.StatusInternalServerError)
				return
			}
			rows, err := dbpkg.ListSuperAuditoriaEventos(dbSuper, filter)
			if err != nil {
				http.Error(w, "No se pudo consultar la auditoria global", http.StatusInternalServerError)
				return
			}
			limit, offset := normalizeAuditoriaPage(filter.Limit, filter.Offset)
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":        true,
				"items":     rows,
				"total":     total,
				"limit":     limit,
				"offset":    offset,
				"scope":     superAuditoriaScopeLabel(requesterAdmin, effectivePrincipal),
				"auditor":   strings.TrimSpace(requesterAdmin.Email),
				"principal": strings.TrimSpace(effectivePrincipal),
			})
			return
		case http.MethodPost:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action != "ui_event" && action != "evento_ui" {
				http.Error(w, "action invalida", http.StatusBadRequest)
				return
			}
			var payload superAuditoriaUIPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID > 0 {
				ok, err := dbpkg.CanAdminAccessEmpresaIA(dbEmp, dbSuper, strings.TrimSpace(requesterAdmin.Email), payload.EmpresaID)
				if err != nil {
					http.Error(w, "failed to validate empresa scope", http.StatusInternalServerError)
					return
				}
				if !ok {
					http.Error(w, "empresa fuera del alcance del administrador autenticado", http.StatusForbidden)
					return
				}
			}
			effectivePrincipal := superAuditoriaEffectivePrincipalScope(r, requesterAdmin, principalEmail)
			eventID, err := dbpkg.CreateSuperAuditoriaEvento(dbSuper, dbpkg.SuperAuditoriaEvento{
				EmpresaID:      payload.EmpresaID,
				PrincipalEmail: effectivePrincipal,
				Modulo:         firstNonBlank(payload.Modulo, "selector_empresa_ui"),
				Accion:         firstNonBlank(payload.Accion, "interaccion"),
				Recurso:        firstNonBlank(payload.Recurso, "seleccionar_empresa"),
				RecursoID:      payload.RecursoID,
				MetodoHTTP:     r.Method,
				Endpoint:       firstNonBlank(payload.Endpoint, r.URL.Path),
				Resultado:      "ok",
				CodigoHTTP:     http.StatusOK,
				RequestID:      resolveAuditoriaRequestID(r),
				IPOrigen:       resolveAuditoriaIP(r),
				UserAgent:      r.UserAgent(),
				MetadataJSON:   superAuditoriaMetadataJSON(payload.Metadata),
				UsuarioCreador: strings.TrimSpace(requesterAdmin.Email),
				Observaciones:  firstNonBlank(payload.Observaciones, "evento visual del selector de empresas"),
			})
			if err != nil {
				http.Error(w, "No se pudo registrar auditoria global", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": eventID})
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

// WithSuperAuditoria registra trazabilidad de endpoints globales usados por el selector.
func WithSuperAuditoria(dbSuper *sql.DB, modulo string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper); !ok {
			return
		}
		start := time.Now()
		rw := &auditCaptureResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		if rw.status == 0 {
			rw.status = http.StatusOK
		}
		registrarSuperAuditoriaNoBloqueante(dbSuper, r, modulo, rw.status, time.Since(start))
	}
}

func registrarSuperAuditoriaNoBloqueante(dbSuper *sql.DB, r *http.Request, modulo string, statusCode int, elapsed time.Duration) {
	if dbSuper == nil || r == nil {
		return
	}
	if !superAuditoriaDebeRegistrar(r, statusCode) {
		return
	}
	adminEmail := strings.TrimSpace(adminEmailFromRequest(r))
	principalEmail := adminEmail
	if admin, principal, err := resolveRequesterAdminScope(dbSuper, r); err == nil {
		if admin != nil && strings.TrimSpace(admin.Email) != "" {
			adminEmail = strings.TrimSpace(admin.Email)
		}
		if principal != "" {
			principalEmail = principal
		} else if admin != nil && utils.IsSuperPanelRole(admin.Role) && !administradoresRequestUsesPrincipalScope(r) {
			principalEmail = ""
		}
	}
	audit := dbpkg.SuperAuditoriaEvento{
		EmpresaID:      resolveSuperAuditoriaEmpresaID(r),
		PrincipalEmail: principalEmail,
		Modulo:         firstNonBlank(modulo, "super"),
		Accion:         resolveSuperAuditoriaAccion(r),
		Recurso:        resolveSuperAuditoriaRecurso(r),
		RecursoID:      resolveSuperAuditoriaRecursoID(r),
		MetodoHTTP:     r.Method,
		Endpoint:       r.URL.Path,
		Resultado:      resolveAuditoriaResultado(statusCode),
		CodigoHTTP:     int64(statusCode),
		RequestID:      resolveAuditoriaRequestID(r),
		IPOrigen:       resolveAuditoriaIP(r),
		UserAgent:      r.UserAgent(),
		MetadataJSON: superAuditoriaMetadataJSON(map[string]interface{}{
			"query":       superAuditoriaSafeQuery(r),
			"elapsed_ms":  elapsed.Milliseconds(),
			"referer":     sanitizeSuperAuditoriaString(r.Referer(), 260),
			"content_len": r.ContentLength,
		}),
		UsuarioCreador: firstNonBlank(adminEmail, "anonimo"),
		Observaciones:  "auditoria automatica de operacion global",
	}
	go func() {
		if _, err := dbpkg.CreateSuperAuditoriaEvento(dbSuper, audit); err != nil {
			log.Printf("[auditoria_super] no se pudo registrar evento modulo=%s accion=%s status=%d error=%v", audit.Modulo, audit.Accion, statusCode, err)
		}
	}()
}

func buildSuperAuditoriaFilterFromRequest(r *http.Request) (dbpkg.SuperAuditoriaEventoFilter, error) {
	limit, err := parseIntQueryOptional(r, "limit")
	if err != nil {
		return dbpkg.SuperAuditoriaEventoFilter{}, err
	}
	offset, err := parseIntQueryOptional(r, "offset")
	if err != nil {
		return dbpkg.SuperAuditoriaEventoFilter{}, err
	}
	empresaID, err := parseInt64QueryOptional(r, "empresa_id")
	if err != nil {
		return dbpkg.SuperAuditoriaEventoFilter{}, err
	}
	recursoID, err := parseInt64QueryOptional(r, "recurso_id")
	if err != nil {
		return dbpkg.SuperAuditoriaEventoFilter{}, err
	}
	codigoHTTP, err := parseInt64QueryOptional(r, "codigo_http")
	if err != nil {
		return dbpkg.SuperAuditoriaEventoFilter{}, err
	}
	desde, err := normalizeAuditoriaDateTime(strings.TrimSpace(r.URL.Query().Get("desde")), false)
	if err != nil {
		return dbpkg.SuperAuditoriaEventoFilter{}, err
	}
	hasta, err := normalizeAuditoriaDateTime(strings.TrimSpace(r.URL.Query().Get("hasta")), true)
	if err != nil {
		return dbpkg.SuperAuditoriaEventoFilter{}, err
	}
	return dbpkg.SuperAuditoriaEventoFilter{
		EmpresaID:       empresaID,
		Modulo:          strings.TrimSpace(r.URL.Query().Get("modulo")),
		Accion:          strings.TrimSpace(r.URL.Query().Get("accion")),
		MetodoHTTP:      strings.TrimSpace(r.URL.Query().Get("metodo_http")),
		Recurso:         strings.TrimSpace(r.URL.Query().Get("recurso")),
		Endpoint:        strings.TrimSpace(r.URL.Query().Get("endpoint")),
		Search:          strings.TrimSpace(r.URL.Query().Get("search")),
		RecursoID:       recursoID,
		CodigoHTTP:      codigoHTTP,
		Resultado:       strings.TrimSpace(r.URL.Query().Get("resultado")),
		UsuarioCreador:  strings.TrimSpace(r.URL.Query().Get("usuario")),
		RequestID:       strings.TrimSpace(r.URL.Query().Get("request_id")),
		Desde:           desde,
		Hasta:           hasta,
		IncludeInactive: queryBool(r, "include_inactive"),
		Limit:           limit,
		Offset:          offset,
		PrincipalEmail:  strings.TrimSpace(r.URL.Query().Get("principal_email")),
	}, nil
}

func superAuditoriaEffectivePrincipalScope(r *http.Request, admin *dbpkg.Admin, principalEmail string) string {
	principalEmail = strings.ToLower(strings.TrimSpace(principalEmail))
	if principalEmail != "" {
		return principalEmail
	}
	if admin == nil {
		return ""
	}
	if utils.IsSuperPanelRole(admin.Role) && (superAuditoriaRequestUsesSuperPanelScope(r) || !administradoresRequestUsesPrincipalScope(r)) {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(admin.Email))
}

func superAuditoriaScopeLabel(admin *dbpkg.Admin, principalEmail string) string {
	if admin != nil && utils.IsSuperPanelRole(admin.Role) && strings.TrimSpace(principalEmail) == "" {
		return "global"
	}
	return "principal"
}

func superAuditoriaRequestUsesSuperPanelScope(r *http.Request) bool {
	if r == nil {
		return false
	}
	scope := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scope")))
	switch scope {
	case "super", "super_panel", "super_admin", "panel_super":
		return true
	default:
		return false
	}
}

func superAuditoriaDebeRegistrar(r *http.Request, statusCode int) bool {
	_ = r
	_ = statusCode
	return true
}

func resolveSuperAuditoriaAccion(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	if action != "" {
		return action
	}
	switch strings.ToUpper(strings.TrimSpace(r.Method)) {
	case http.MethodGet:
		return "consultar"
	case http.MethodPost:
		return "crear"
	case http.MethodPut:
		return "actualizar"
	case http.MethodDelete:
		return "eliminar"
	default:
		return strings.ToLower(strings.TrimSpace(r.Method))
	}
}

func resolveSuperAuditoriaRecurso(r *http.Request) string {
	path := strings.Trim(strings.ToLower(strings.TrimSpace(r.URL.Path)), "/")
	path = strings.ReplaceAll(path, "/", "_")
	if path == "" {
		return "super"
	}
	return path
}

func resolveSuperAuditoriaEmpresaID(r *http.Request) int64 {
	for _, key := range []string{"empresa_id", "id"} {
		if raw := strings.TrimSpace(r.URL.Query().Get(key)); raw != "" {
			if id, err := strconv.ParseInt(raw, 10, 64); err == nil && id > 0 {
				if strings.Contains(strings.ToLower(r.URL.Path), "empresas") || key == "empresa_id" {
					return id
				}
			}
		}
	}
	return 0
}

func resolveSuperAuditoriaRecursoID(r *http.Request) int64 {
	for _, key := range []string{"id", "recurso_id", "licencia_id", "administrador_id", "invitation_id"} {
		if raw := strings.TrimSpace(r.URL.Query().Get(key)); raw != "" {
			if id, err := strconv.ParseInt(raw, 10, 64); err == nil && id > 0 {
				return id
			}
		}
	}
	return 0
}

func superAuditoriaSafeQuery(r *http.Request) map[string]string {
	out := map[string]string{}
	if r == nil || r.URL == nil {
		return out
	}
	for key, values := range r.URL.Query() {
		lower := strings.ToLower(strings.TrimSpace(key))
		if lower == "" {
			continue
		}
		if strings.Contains(lower, "token") || strings.Contains(lower, "password") || strings.Contains(lower, "clave") || strings.Contains(lower, "secret") {
			out[lower] = "[redacted]"
			continue
		}
		if len(values) == 0 {
			out[lower] = ""
			continue
		}
		out[lower] = sanitizeSuperAuditoriaString(values[0], 180)
	}
	return out
}

func superAuditoriaMetadataJSON(in map[string]interface{}) string {
	clean := sanitizeSuperAuditoriaMap(in)
	raw, err := json.Marshal(clean)
	if err != nil || !json.Valid(raw) {
		return "{}"
	}
	return string(raw)
}

func sanitizeSuperAuditoriaMap(in map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	for key, value := range in {
		k := strings.ToLower(sanitizeSuperAuditoriaString(key, 80))
		if k == "" {
			continue
		}
		if strings.Contains(k, "token") || strings.Contains(k, "password") || strings.Contains(k, "clave") || strings.Contains(k, "secret") {
			out[k] = "[redacted]"
			continue
		}
		switch v := value.(type) {
		case string:
			out[k] = sanitizeSuperAuditoriaString(v, 240)
		case bool, int, int64, float64, float32, nil:
			out[k] = v
		case map[string]string:
			nested := map[string]interface{}{}
			for nk, nv := range v {
				nested[nk] = nv
			}
			out[k] = sanitizeSuperAuditoriaMap(nested)
		case map[string]interface{}:
			out[k] = sanitizeSuperAuditoriaMap(v)
		default:
			out[k] = sanitizeSuperAuditoriaString(strings.TrimSpace(toString(v)), 240)
		}
	}
	return out
}

func sanitizeSuperAuditoriaString(value string, limit int) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	if limit > 0 && len([]rune(value)) > limit {
		return string([]rune(value)[:limit])
	}
	return value
}

func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(raw)
	}
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
