package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaGimnasioHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "" {
			action = "dashboard"
		}

		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			switch action {
			case "dashboard":
				row, err := dbpkg.GetEmpresaGimnasioDashboard(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar el dashboard de gimnasio", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "socios":
				rows, err := dbpkg.ListEmpresaGimnasioSocios(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar los socios", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "planes":
				rows, err := dbpkg.ListEmpresaGimnasioPlanes(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar los planes", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "entrenadores":
				rows, err := dbpkg.ListEmpresaGimnasioEntrenadores(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar los entrenadores", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "clases":
				rows, err := dbpkg.ListEmpresaGimnasioClases(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar las clases", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "inscripciones":
				rows, err := dbpkg.ListEmpresaGimnasioInscripciones(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar las inscripciones", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "asistencias":
				rows, err := dbpkg.ListEmpresaGimnasioAsistencias(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar las asistencias", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "pagos":
				rows, err := dbpkg.ListEmpresaGimnasioPagos(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar los pagos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "acceso_config":
				cfg, err := dbpkg.GetEmpresaGimnasioAccesoConfig(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar la configuracion de acceso", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, cfg)
				return
			case "credenciales":
				rows, err := dbpkg.ListEmpresaGimnasioCredenciales(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar las credenciales", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "dispositivos":
				rows, err := dbpkg.ListEmpresaGimnasioDispositivosAcceso(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar los dispositivos de acceso", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "eventos_acceso":
				rows, err := dbpkg.ListEmpresaGimnasioEventosAcceso(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar la bitacora de acceso", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			default:
				http.Error(w, "action no soportada", http.StatusBadRequest)
				return
			}

		case http.MethodPost:
			switch action {
			case "socios":
				var payload dbpkg.EmpresaGimnasioSocio
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				id, err := dbpkg.CreateEmpresaGimnasioSocio(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "planes":
				var payload dbpkg.EmpresaGimnasioPlan
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				id, err := dbpkg.CreateEmpresaGimnasioPlan(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "entrenadores":
				var payload dbpkg.EmpresaGimnasioEntrenador
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				id, err := dbpkg.CreateEmpresaGimnasioEntrenador(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "clases":
				var payload dbpkg.EmpresaGimnasioClase
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				id, err := dbpkg.CreateEmpresaGimnasioClase(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "inscripciones":
				var payload dbpkg.EmpresaGimnasioInscripcion
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				id, err := dbpkg.CreateEmpresaGimnasioInscripcion(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "asistencias", "checkin":
				var payload dbpkg.EmpresaGimnasioAsistencia
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				id, err := dbpkg.CreateEmpresaGimnasioAsistencia(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "pagos":
				var payload dbpkg.EmpresaGimnasioPago
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				id, err := dbpkg.CreateEmpresaGimnasioPago(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "credenciales":
				var payload dbpkg.EmpresaGimnasioCredencial
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				id, err := dbpkg.CreateEmpresaGimnasioCredencial(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "dispositivos":
				var payload dbpkg.EmpresaGimnasioDispositivoAcceso
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				id, err := dbpkg.CreateEmpresaGimnasioDispositivoAcceso(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "validar_acceso":
				var payload struct {
					EmpresaID        int64  `json:"empresa_id"`
					CodigoCredencial string `json:"codigo_credencial"`
					MetodoAcceso     string `json:"metodo_acceso"`
					DispositivoID    int64  `json:"dispositivo_id"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				evento, err := dbpkg.ValidarEmpresaGimnasioAcceso(dbEmp, payload.EmpresaID, payload.CodigoCredencial, payload.MetodoAcceso, payload.DispositivoID, strings.TrimSpace(adminEmailFromRequest(r)))
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": evento.Resultado == "aprobado", "evento": evento})
				return
			default:
				http.Error(w, "action no soportada", http.StatusBadRequest)
				return
			}

		case http.MethodPut:
			switch action {
			case "socios":
				var payload dbpkg.EmpresaGimnasioSocio
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if err := dbpkg.UpdateEmpresaGimnasioSocio(dbEmp, payload); err != nil {
					handleGimnasioUpdateError(w, err, "socio")
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "planes":
				var payload dbpkg.EmpresaGimnasioPlan
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if err := dbpkg.UpdateEmpresaGimnasioPlan(dbEmp, payload); err != nil {
					handleGimnasioUpdateError(w, err, "plan")
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "entrenadores":
				var payload dbpkg.EmpresaGimnasioEntrenador
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if err := dbpkg.UpdateEmpresaGimnasioEntrenador(dbEmp, payload); err != nil {
					handleGimnasioUpdateError(w, err, "entrenador")
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "clases":
				var payload dbpkg.EmpresaGimnasioClase
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if err := dbpkg.UpdateEmpresaGimnasioClase(dbEmp, payload); err != nil {
					handleGimnasioUpdateError(w, err, "clase")
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "acceso_config":
				var payload dbpkg.EmpresaGimnasioAccesoConfig
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				id, err := dbpkg.UpsertEmpresaGimnasioAccesoConfig(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				cfg, _ := dbpkg.GetEmpresaGimnasioAccesoConfig(dbEmp, payload.EmpresaID)
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id, "configuracion": cfg})
				return
			case "credenciales":
				var payload dbpkg.EmpresaGimnasioCredencial
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if err := dbpkg.UpdateEmpresaGimnasioCredencial(dbEmp, payload); err != nil {
					handleGimnasioUpdateError(w, err, "credencial")
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "dispositivos":
				var payload dbpkg.EmpresaGimnasioDispositivoAcceso
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if err := dbpkg.UpdateEmpresaGimnasioDispositivoAcceso(dbEmp, payload); err != nil {
					handleGimnasioUpdateError(w, err, "dispositivo")
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "cancelar_inscripcion":
				empresaID, err := parseInt64Query(r, "empresa_id")
				if err != nil {
					http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
					return
				}
				id, err := parseInt64Query(r, "id")
				if err != nil {
					http.Error(w, "id es obligatorio", http.StatusBadRequest)
					return
				}
				if err := dbpkg.UpdateEmpresaGimnasioInscripcionEstado(dbEmp, empresaID, id, "cancelada"); err != nil {
					handleGimnasioUpdateError(w, err, "inscripcion")
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": "cancelada"})
				return
			default:
				http.Error(w, "action no soportada", http.StatusBadRequest)
				return
			}

		case http.MethodDelete:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			id, err := parseInt64Query(r, "id")
			if err != nil {
				http.Error(w, "id es obligatorio", http.StatusBadRequest)
				return
			}
			switch action {
			case "socios":
				err = dbpkg.DeleteEmpresaGimnasioSocio(dbEmp, empresaID, id)
			case "planes":
				err = dbpkg.DeleteEmpresaGimnasioPlan(dbEmp, empresaID, id)
			case "entrenadores":
				err = dbpkg.DeleteEmpresaGimnasioEntrenador(dbEmp, empresaID, id)
			case "clases":
				err = dbpkg.DeleteEmpresaGimnasioClase(dbEmp, empresaID, id)
			case "credenciales":
				err = dbpkg.DeleteEmpresaGimnasioCredencial(dbEmp, empresaID, id)
			case "dispositivos":
				err = dbpkg.DeleteEmpresaGimnasioDispositivoAcceso(dbEmp, empresaID, id)
			default:
				http.Error(w, "action no soportada", http.StatusBadRequest)
				return
			}
			if err != nil {
				handleGimnasioUpdateError(w, err, "registro")
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func handleGimnasioUpdateError(w http.ResponseWriter, err error, label string) {
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, label+" no encontrado", http.StatusNotFound)
		return
	}
	http.Error(w, err.Error(), http.StatusBadRequest)
}
