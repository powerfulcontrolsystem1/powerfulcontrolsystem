package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaModuloColombiaHandler(dbEmp *sql.DB, modulo string) http.HandlerFunc {
	modulo = dbpkg.NormalizeEmpresaModuloColombia(modulo)
	return func(w http.ResponseWriter, r *http.Request) {
		if modulo == "" {
			http.Error(w, "modulo no soportado", http.StatusBadRequest)
			return
		}
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
			case "", "dashboard":
				row, err := dbpkg.BuildEmpresaModuloColombiaDashboard(dbEmp, empresaID, modulo)
				if err != nil {
					http.Error(w, "No se pudo consultar el modulo empresarial", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "plantilla":
				writeJSON(w, http.StatusOK, dbpkg.GetEmpresaModuloColombiaPlantilla(modulo))
				return
			case "diagnostico", "diagnostico_configuracion":
				row, err := dbpkg.BuildEmpresaModuloColombiaDiagnostico(dbEmp, empresaID, modulo)
				if err != nil {
					http.Error(w, "No se pudo generar el diagnostico del modulo empresarial", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "reporte":
				row, err := dbpkg.BuildEmpresaModuloColombiaReporte(dbEmp, empresaID, modulo)
				if err != nil {
					http.Error(w, "No se pudo generar el reporte del modulo empresarial", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "agenda":
				row, err := dbpkg.BuildEmpresaModuloColombiaAgenda(dbEmp, empresaID, modulo)
				if err != nil {
					http.Error(w, "No se pudo generar la agenda del modulo empresarial", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "responsables":
				rows, err := dbpkg.BuildEmpresaModuloColombiaResponsables(dbEmp, empresaID, modulo)
				if err != nil {
					http.Error(w, "No se pudo generar el tablero de responsables", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "sla":
				row, err := dbpkg.BuildEmpresaModuloColombiaSLA(dbEmp, empresaID, modulo)
				if err != nil {
					http.Error(w, "No se pudo generar el SLA del modulo empresarial", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "riesgo":
				row, err := dbpkg.BuildEmpresaModuloColombiaRiesgo(dbEmp, empresaID, modulo)
				if err != nil {
					http.Error(w, "No se pudo generar la matriz de riesgo del modulo empresarial", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "exportacion":
				row, err := dbpkg.BuildEmpresaModuloColombiaExportacion(dbEmp, empresaID, modulo)
				if err != nil {
					http.Error(w, "No se pudo generar la exportacion del modulo empresarial", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "expediente":
				registroID := int64Query(r, "registro_id")
				row, err := dbpkg.GetEmpresaModuloColombiaExpediente(dbEmp, empresaID, modulo, registroID)
				if err != nil {
					http.Error(w, "No se pudo consultar el expediente del registro", http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "registros":
				rows, err := dbpkg.ListEmpresaModuloColombiaRegistros(dbEmp, empresaID, modulo, strings.TrimSpace(r.URL.Query().Get("estado")), 500)
				if err != nil {
					http.Error(w, "No se pudieron listar registros", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "buscar":
				proximos, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("proximos_dias")))
				row, err := dbpkg.BuscarEmpresaModuloColombiaRegistros(dbEmp, empresaID, modulo, dbpkg.EmpresaModuloColombiaFiltro{
					Texto:        r.URL.Query().Get("texto"),
					Estado:       r.URL.Query().Get("estado"),
					Tipo:         r.URL.Query().Get("tipo"),
					Categoria:    r.URL.Query().Get("categoria"),
					Prioridad:    r.URL.Query().Get("prioridad"),
					Responsable:  r.URL.Query().Get("responsable"),
					Vencidos:     strings.EqualFold(r.URL.Query().Get("vencidos"), "true") || r.URL.Query().Get("vencidos") == "1",
					ProximosDias: proximos,
				}, 500)
				if err != nil {
					http.Error(w, "No se pudo buscar registros del modulo empresarial", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "eventos":
				rows, err := dbpkg.ListEmpresaModuloColombiaEventos(dbEmp, empresaID, modulo, 100)
				if err != nil {
					http.Error(w, "No se pudieron listar eventos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "evidencias":
				registroID := int64Query(r, "registro_id")
				rows, err := dbpkg.ListEmpresaModuloColombiaEvidencias(dbEmp, empresaID, modulo, registroID, 100)
				if err != nil {
					http.Error(w, "No se pudieron listar evidencias", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "aprobaciones":
				registroID := int64Query(r, "registro_id")
				rows, err := dbpkg.ListEmpresaModuloColombiaAprobaciones(dbEmp, empresaID, modulo, registroID, strings.TrimSpace(r.URL.Query().Get("estado")), 100)
				if err != nil {
					http.Error(w, "No se pudieron listar aprobaciones", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "tareas":
				registroID := int64Query(r, "registro_id")
				rows, err := dbpkg.ListEmpresaModuloColombiaTareas(dbEmp, empresaID, modulo, registroID, strings.TrimSpace(r.URL.Query().Get("estado")), 100)
				if err != nil {
					http.Error(w, "No se pudieron listar tareas", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}
		case http.MethodPost, http.MethodPut:
			switch action {
			case "registro":
				var payload dbpkg.EmpresaModuloColombiaRegistro
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.Modulo = modulo
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.UpsertEmpresaModuloColombiaRegistro(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				status := http.StatusCreated
				if r.Method == http.MethodPut {
					status = http.StatusOK
				}
				writeJSON(w, status, map[string]interface{}{"ok": true, "id": id})
				return
			case "estado":
				var payload struct {
					RegistroID int64  `json:"registro_id"`
					Estado     string `json:"estado"`
					Detalle    string `json:"detalle"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if payload.RegistroID <= 0 {
					http.Error(w, "registro_id es requerido", http.StatusBadRequest)
					return
				}
				if err := dbpkg.CambiarEstadoEmpresaModuloColombiaRegistro(dbEmp, empresaID, modulo, payload.RegistroID, payload.Estado, payload.Detalle, adminEmail); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "accion_masiva":
				var payload dbpkg.EmpresaModuloColombiaAccionMasiva
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				result, err := dbpkg.AplicarEmpresaModuloColombiaAccionMasiva(dbEmp, empresaID, modulo, payload, adminEmail)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, result)
				return
			case "cierre_controlado":
				var payload struct {
					RegistroID int64  `json:"registro_id"`
					Detalle    string `json:"detalle"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if err := dbpkg.CerrarEmpresaModuloColombiaRegistroControlado(dbEmp, empresaID, modulo, payload.RegistroID, payload.Detalle, adminEmail); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "evento":
				var payload struct {
					RegistroID     int64  `json:"registro_id"`
					Evento         string `json:"evento"`
					EstadoAnterior string `json:"estado_anterior"`
					EstadoNuevo    string `json:"estado_nuevo"`
					Detalle        string `json:"detalle"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if err := dbpkg.RegistrarEmpresaModuloColombiaEvento(dbEmp, empresaID, modulo, payload.RegistroID, payload.Evento, payload.EstadoAnterior, payload.EstadoNuevo, payload.Detalle, adminEmail); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "evidencia":
				var payload dbpkg.EmpresaModuloColombiaEvidencia
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.Modulo = modulo
				payload.Usuario = adminEmail
				id, err := dbpkg.RegistrarEmpresaModuloColombiaEvidencia(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "aprobacion_solicitar":
				var payload dbpkg.EmpresaModuloColombiaAprobacion
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.Modulo = modulo
				payload.SolicitadoPor = adminEmail
				id, err := dbpkg.SolicitarEmpresaModuloColombiaAprobacion(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "aprobacion_decidir":
				var payload struct {
					AprobacionID int64  `json:"aprobacion_id"`
					Decision     string `json:"decision"`
					Comentario   string `json:"comentario"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if err := dbpkg.DecidirEmpresaModuloColombiaAprobacion(dbEmp, empresaID, modulo, payload.AprobacionID, payload.Decision, payload.Comentario, adminEmail); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "tarea":
				var payload dbpkg.EmpresaModuloColombiaTarea
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.Modulo = modulo
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.CrearEmpresaModuloColombiaTarea(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "tarea_estado":
				var payload struct {
					TareaID    int64  `json:"tarea_id"`
					Estado     string `json:"estado"`
					Comentario string `json:"comentario"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if err := dbpkg.CambiarEstadoEmpresaModuloColombiaTarea(dbEmp, empresaID, modulo, payload.TareaID, payload.Estado, payload.Comentario, adminEmail); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "generar_plan_accion":
				result, err := dbpkg.GenerarEmpresaModuloColombiaPlanAccion(dbEmp, empresaID, modulo, adminEmail)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, result)
				return
			case "importar_registros":
				var payload struct {
					Registros []dbpkg.EmpresaModuloColombiaRegistro `json:"registros"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				result, err := dbpkg.ImportEmpresaModuloColombiaRegistros(dbEmp, empresaID, modulo, payload.Registros, adminEmail)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, result)
				return
			case "seed_demo":
				if err := dbpkg.SeedEmpresaModuloColombiaDemo(dbEmp, empresaID, modulo, adminEmail); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			}
		}

		http.Error(w, "accion o metodo no soportado", http.StatusMethodNotAllowed)
	}
}
