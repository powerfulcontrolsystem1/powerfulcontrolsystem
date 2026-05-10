package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaHorariosTrabajadoresHandler expone un modulo profesional de programacion laboral.
func EmpresaHorariosTrabajadoresHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		adminEmail := strings.TrimSpace(adminEmailFromRequest(r))

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "config", "configuracion":
				cfg, err := dbpkg.GetHorarioTrabajadorConfig(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar la configuracion de horarios", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, cfg)
				return

			case "dashboard", "resumen":
				desde := strings.TrimSpace(r.URL.Query().Get("desde"))
				hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
				dashboard, err := dbpkg.BuildHorarioTrabajadorDashboard(dbEmp, empresaID, desde, hasta)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, dashboard)
				return

			case "by_user", "por_usuario":
				usuarioID, err := parseInt64QueryOptional(r, "usuario_id")
				if err != nil || usuarioID <= 0 {
					http.Error(w, "usuario_id invalido", http.StatusBadRequest)
					return
				}
				desde := strings.TrimSpace(r.URL.Query().Get("desde"))
				hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
				items, err := dbpkg.GetHorariosTrabajadorByUsuario(dbEmp, empresaID, usuarioID, desde, hasta)
				if err != nil {
					http.Error(w, "No se pudo consultar la programacion del usuario", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "items": items})
				return
			}

			desde := strings.TrimSpace(r.URL.Query().Get("desde"))
			hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
			q := strings.TrimSpace(r.URL.Query().Get("q"))
			area := strings.TrimSpace(r.URL.Query().Get("area"))
			sede := strings.TrimSpace(r.URL.Query().Get("sede"))
			estado := strings.TrimSpace(r.URL.Query().Get("estado"))
			publishedOnly := queryBool(r, "published_only")
			limit, err := parseIntQueryOptional(r, "limit")
			if err != nil {
				http.Error(w, "limit invalido", http.StatusBadRequest)
				return
			}
			items, err := dbpkg.ListHorariosTrabajadores(dbEmp, empresaID, desde, hasta, q, area, sede, estado, publishedOnly, limit)
			if err != nil {
				http.Error(w, "No se pudo listar la programacion laboral", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "items": items})
			return

		case http.MethodPost:
			switch action {
			case "config", "configuracion", "save_config":
				var payload dbpkg.HorarioTrabajadorConfig
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				if err := dbpkg.UpsertHorarioTrabajadorConfig(dbEmp, payload); err != nil {
					http.Error(w, "No se pudo guardar la configuracion de horarios", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return

			case "bulk_create", "programar_semana", "programar_rango":
				var payload dbpkg.HorarioTrabajadorBulkInput
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				created, warnings, err := dbpkg.CreateHorariosTrabajadoresBulk(dbEmp, payload)
				if err != nil {
					if errors.Is(err, dbpkg.ErrHorarioTrabajadorConflict) {
						http.Error(w, err.Error(), http.StatusConflict)
						return
					}
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{
					"ok":       true,
					"creados":  created,
					"warnings": warnings,
				})
				return

			case "publish", "publicar":
				var payload dbpkg.HorarioTrabajadorPublishInput
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				rows, err := dbpkg.PublishHorariosTrabajadores(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "actualizados": rows})
				return
			}

			var payload dbpkg.HorarioTrabajador
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			payload.EmpresaID = empresaID
			payload.UsuarioCreador = adminEmail
			id, err := dbpkg.CreateHorarioTrabajador(dbEmp, &payload)
			if err != nil {
				if errors.Is(err, dbpkg.ErrHorarioTrabajadorConflict) {
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]interface{}{
				"ok":                    true,
				"id":                    id,
				"conflictos_detectados": payload.ConflictosDetectados,
			})
			return

		case http.MethodPut:
			var payload dbpkg.HorarioTrabajador
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			payload.EmpresaID = empresaID
			payload.UsuarioCreador = adminEmail
			if err := dbpkg.UpdateHorarioTrabajador(dbEmp, &payload); err != nil {
				if errors.Is(err, dbpkg.ErrHorarioTrabajadorConflict) {
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":                    true,
				"conflictos_detectados": payload.ConflictosDetectados,
			})
			return

		case http.MethodDelete:
			id, err := parseInt64QueryOptional(r, "id")
			if err != nil || id <= 0 {
				http.Error(w, "id invalido", http.StatusBadRequest)
				return
			}
			if err := dbpkg.DeleteHorarioTrabajador(dbEmp, id, empresaID); err != nil {
				http.Error(w, "No se pudo eliminar el turno", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

// EmpresaMiHorarioUsuarioHandler expone la programacion publicada del usuario autenticado.
func EmpresaMiHorarioUsuarioHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		adminEmail := strings.ToLower(strings.TrimSpace(adminEmailFromRequest(r)))
		if adminEmail == "" || adminEmail == "sistema" {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		usuario, err := dbpkg.GetEmpresaUsuarioByEmailScoped(dbEmp, adminEmail, empresaID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				desde := strings.TrimSpace(r.URL.Query().Get("desde"))
				hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
				if desde == "" || hasta == "" {
					today := time.Now()
					if desde == "" {
						desde = today.Format("2006-01-02")
					}
					if hasta == "" {
						hasta = today.AddDate(0, 0, 14).Format("2006-01-02")
					}
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":         true,
					"empresa_id": empresaID,
					"desde":      desde,
					"hasta":      hasta,
					"usuario": map[string]interface{}{
						"id":     0,
						"email":  adminEmail,
						"nombre": adminEmail,
						"rol":    "administrador",
					},
					"resumen": buildMiHorarioResumen(nil, time.Now()),
					"items":   []dbpkg.HorarioTrabajador{},
					"warning": "No hay un usuario operativo asociado a este correo en la empresa.",
				})
				return
			}
			http.Error(w, "No se pudo validar el usuario operativo", http.StatusInternalServerError)
			return
		}
		desde := strings.TrimSpace(r.URL.Query().Get("desde"))
		hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
		if desde == "" || hasta == "" {
			today := time.Now()
			if desde == "" {
				desde = today.Format("2006-01-02")
			}
			if hasta == "" {
				hasta = today.AddDate(0, 0, 14).Format("2006-01-02")
			}
		}
		items, err := dbpkg.GetHorariosTrabajadorByUsuarioPerfil(dbEmp, empresaID, usuario.ID, usuario.Email, desde, hasta, true, 300)
		if err != nil {
			http.Error(w, "No se pudo consultar tu horario", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":         true,
			"empresa_id": empresaID,
			"desde":      desde,
			"hasta":      hasta,
			"usuario": map[string]interface{}{
				"id":     usuario.ID,
				"email":  usuario.Email,
				"nombre": usuario.Nombre,
				"rol":    usuario.RolNombre,
			},
			"resumen": buildMiHorarioResumen(items, time.Now()),
			"items":   items,
		})
	}
}

func buildMiHorarioResumen(items []dbpkg.HorarioTrabajador, now time.Time) map[string]interface{} {
	today := now.Format("2006-01-02")
	var horas float64
	var turnosHoy int
	var proximos int
	for _, item := range items {
		horas += item.HorasProgramadas
		if strings.TrimSpace(item.Fecha) == today {
			turnosHoy++
		}
		if strings.TrimSpace(item.Fecha) >= today {
			proximos++
		}
	}
	return map[string]interface{}{
		"turnos":         len(items),
		"turnos_hoy":     turnosHoy,
		"proximos":       proximos,
		"horas":          horas,
		"actualizado_en": now.Format("2006-01-02 15:04:05"),
	}
}
