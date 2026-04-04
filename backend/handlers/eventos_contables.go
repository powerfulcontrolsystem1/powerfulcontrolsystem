package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

// registrarEventoContableNoBloqueante registra eventos contables sin interrumpir el flujo HTTP.
func registrarEventoContableNoBloqueante(dbEmp *sql.DB, r *http.Request, scope string, evento dbpkg.EmpresaEventoContable, payload map[string]interface{}) {
	if dbEmp == nil || evento.EmpresaID <= 0 {
		return
	}
	evento.Modulo = strings.ToLower(strings.TrimSpace(evento.Modulo))
	evento.Evento = strings.ToLower(strings.TrimSpace(evento.Evento))
	evento.Entidad = strings.TrimSpace(evento.Entidad)
	if evento.Modulo == "" || evento.Evento == "" || evento.Entidad == "" {
		return
	}
	if strings.TrimSpace(evento.UsuarioCreador) == "" {
		evento.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
	}
	if strings.TrimSpace(evento.Estado) == "" {
		evento.Estado = "activo"
	}
	if strings.TrimSpace(evento.Observaciones) != "" {
		evento.Observaciones = strings.TrimSpace(evento.Observaciones)
	}
	if strings.TrimSpace(evento.PayloadJSON) == "" && payload != nil {
		if b, err := json.Marshal(payload); err != nil {
			if strings.TrimSpace(scope) == "" {
				scope = "eventos_contables"
			}
			log.Printf("[%s] no se pudo serializar payload empresa_id=%d modulo=%s evento=%s error: %v", scope, evento.EmpresaID, evento.Modulo, evento.Evento, err)
		} else {
			evento.PayloadJSON = string(b)
		}
	}
	if _, err := dbpkg.CreateEmpresaEventoContable(dbEmp, evento); err != nil {
		if strings.TrimSpace(scope) == "" {
			scope = "eventos_contables"
		}
		log.Printf("[%s] evento contable omitido empresa_id=%d modulo=%s evento=%s entidad=%s entidad_id=%d error: %v", scope, evento.EmpresaID, evento.Modulo, evento.Evento, evento.Entidad, evento.EntidadID, err)
	}
}
