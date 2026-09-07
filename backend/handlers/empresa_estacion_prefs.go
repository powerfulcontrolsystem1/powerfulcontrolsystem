package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaEstacionPrefsHandler maneja GET (listar) y PUT/POST (upsert) de prefs por estacion
func EmpresaEstacionPrefsHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			estacionID := int64(0)
			if s := strings.TrimSpace(r.URL.Query().Get("estacion_id")); s != "" {
				// parse optional
				if v, perr := parseInt64Query(r, "estacion_id"); perr == nil {
					estacionID = v
				}
			}
			prefs, err := dbpkg.ListEmpresaEstacionPrefs(dbEmp, empresaID, estacionID, false)
			if err != nil {
				log.Printf("[estacion_prefs] list empresa_id=%d error: %v", empresaID, err)
				http.Error(w, "No se pudieron obtener preferencias", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "prefs": prefs})
			return

		case http.MethodPut, http.MethodPost:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			var payload struct {
				EstacionID     int64  `json:"estacion_id"`
				Clave          string `json:"clave"`
				Valor          string `json:"valor"`
				UsuarioCreador string `json:"usuario_creador,omitempty"`
				Estado         string `json:"estado,omitempty"`
				Observaciones  string `json:"observaciones,omitempty"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.Clave) == "" {
				http.Error(w, "clave es obligatoria", http.StatusBadRequest)
				return
			}
			p := dbpkg.EmpresaEstacionPref{
				EmpresaID:      empresaID,
				EstacionID:     payload.EstacionID,
				Clave:          strings.TrimSpace(payload.Clave),
				Valor:          strings.TrimSpace(payload.Valor),
				UsuarioCreador: strings.TrimSpace(payload.UsuarioCreador),
				Estado:         payload.Estado,
				Observaciones:  strings.TrimSpace(payload.Observaciones),
			}
			if p.UsuarioCreador == "" {
				p.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			}
			id, err := dbpkg.UpsertEmpresaEstacionPref(dbEmp, p)
			if err != nil {
				log.Printf("[estacion_prefs] upsert empresa_id=%d estacion=%d clave=%s error: %v", empresaID, p.EstacionID, p.Clave, err)
				http.Error(w, "No se pudieron guardar preferencias", http.StatusInternalServerError)
				return
			}
			if auditModulo, auditAccion := resolveEstacionPrefAuditModuleAction(p.Clave, p.Valor); auditModulo != "" {
				registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, auditModulo, auditAccion, "empresa_estacion_prefs", id, http.StatusOK, map[string]interface{}{
					"clave":       p.Clave,
					"estacion_id": p.EstacionID,
					"valor_bytes": len([]byte(p.Valor)),
				}, "preferencia empresarial actualizada")
			}
			response := map[string]interface{}{"ok": true, "id": id}
			if p.EstacionID > 0 && p.Clave == "estacion_estado_sucia" && dbpkg.IsEmpresaEstacionDirtyValue(p.Valor) {
				stationName := dbpkg.ResolveEmpresaEstacionNombre(dbEmp, empresaID, p.EstacionID)
				eventID, startErr := dbpkg.StartEmpresaEstacionAseoEvento(dbEmp, empresaID, p.EstacionID, stationName, "", p.UsuarioCreador)
				if startErr != nil {
					log.Printf("[estacion_prefs] start aseo empresa_id=%d estacion=%d error: %v", empresaID, p.EstacionID, startErr)
					response["aseo_error"] = "No se pudo iniciar el control de aseo"
				} else {
					response["aseo_evento_id"] = eventID
				}
			}
			if p.EstacionID == 0 && p.Clave == "estaciones_config" {
				syncResult, syncErr := dbpkg.SyncEmpresaEstacionCarritos(dbEmp, empresaID, p.Valor, p.UsuarioCreador)
				if syncErr != nil {
					log.Printf("[estacion_prefs] sync carritos empresa_id=%d clave=%s error: %v", empresaID, p.Clave, syncErr)
					// La preferencia ya quedó guardada. No bloqueamos el guardado por un fallo de sincronización,
					// para evitar que la UI revierta checks/flags por un error de carritos (que es un paso secundario).
					response["sync_error"] = "No se pudieron sincronizar los carritos de estaciones"
					response["sync_error_detail"] = syncErr.Error()
				} else {
					response["sync"] = syncResult
				}
			}
			writeJSON(w, http.StatusOK, response)
			return
		}
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func resolveEstacionPrefAuditModuleAction(clave, valor string) (string, string) {
	key := strings.ToLower(strings.TrimSpace(clave))
	value := strings.ToLower(strings.TrimSpace(valor))
	switch key {
	case "menu_visual_config":
		return "menu_visible", "configuracion_guardada"
	case "estaciones_config":
		switch {
		case strings.Contains(value, "atajos_pos"):
			return "atajos_pos", "configuracion_guardada"
		case strings.Contains(value, "pago_qr") || strings.Contains(value, "breb"):
			return "breb_qr", "configuracion_guardada"
		default:
			return "estaciones", "configuracion_guardada"
		}
	case "email_corporativo_config":
		return "email_corporativo", "configuracion_guardada"
	default:
		if strings.HasPrefix(key, "chat_flotante.") {
			return "centro_ia_empresarial", "preferencia_guardada"
		}
	}
	return "", ""
}
