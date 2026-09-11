package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

const loginAuditModule = "autenticacion"

type loginAuditAttempt struct {
	EmpresaID     int64
	Email         string
	PrincipalType string
	AuthMethod    string
	Role          string
	Authenticated bool
	Reason        string
}

func (attempt *loginAuditAttempt) markAuthenticated(role string) {
	if attempt == nil {
		return
	}
	attempt.Authenticated = true
	attempt.Role = strings.TrimSpace(role)
	attempt.Reason = "autenticado"
}

func (attempt *loginAuditAttempt) record(dbSuper *sql.DB, r *http.Request, statusCode int) {
	if attempt == nil || dbSuper == nil || r == nil || (r.Method != http.MethodPost && r.Method != http.MethodGet) {
		return
	}
	event := attempt.buildEvent(r, statusCode)
	if _, err := dbpkg.InsertSuperAuditoriaEventoPrepared(dbSuper, event); err != nil {
		log.Printf("[auditoria_login] no se pudo registrar intento tipo=%s empresa_id=%d resultado=%s status=%d error=%v", event.Recurso, event.EmpresaID, event.Resultado, event.CodigoHTTP, err)
	}
}

func (attempt *loginAuditAttempt) buildEvent(r *http.Request, statusCode int) dbpkg.SuperAuditoriaEvento {
	if statusCode <= 0 {
		statusCode = http.StatusOK
	}
	result := "rechazado"
	action := "inicio_sesion_fallido"
	reason := strings.ToLower(strings.TrimSpace(attempt.Reason))
	if attempt.Authenticated {
		result = "ok"
		action = "inicio_sesion_exitoso"
		reason = "autenticado"
	} else if statusCode >= http.StatusInternalServerError {
		result = "error"
		if reason == "" {
			reason = "error_interno"
		}
	} else if reason == "" {
		switch statusCode {
		case http.StatusBadRequest:
			reason = "solicitud_invalida"
		case http.StatusUnauthorized:
			reason = "credenciales_invalidas"
		case http.StatusForbidden:
			reason = "acceso_denegado"
		case http.StatusTooManyRequests:
			reason = "bloqueo_temporal"
		default:
			reason = "accion_adicional_requerida"
		}
	}

	email := strings.ToLower(strings.TrimSpace(attempt.Email))
	actor := email
	if actor == "" {
		actor = "anonimo"
	}
	return dbpkg.SuperAuditoriaEvento{
		EmpresaID:      attempt.EmpresaID,
		PrincipalEmail: "",
		Modulo:         loginAuditModule,
		Accion:         action,
		Recurso:        firstNonBlank(strings.TrimSpace(attempt.PrincipalType), "cuenta"),
		MetodoHTTP:     r.Method,
		Endpoint:       r.URL.Path,
		Resultado:      result,
		CodigoHTTP:     int64(statusCode),
		RequestID:      resolveAuditoriaRequestID(r),
		IPOrigen:       resolveAuditoriaIP(r),
		UserAgent:      r.UserAgent(),
		MetadataJSON: superAuditoriaMetadataJSON(map[string]interface{}{
			"metodo_autenticacion": firstNonBlank(attempt.AuthMethod, "password"),
			"tipo_principal":       firstNonBlank(attempt.PrincipalType, "cuenta"),
			"motivo":               reason,
			"rol":                  strings.TrimSpace(attempt.Role),
		}),
		UsuarioCreador: actor,
		Observaciones:  "auditoria de inicio de sesion; no almacena contrasenas, tokens ni credenciales",
	}
}
