package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaOdontologiaHandler(dbEmp *sql.DB) http.HandlerFunc {
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
			case "", "dashboard":
				row, err := dbpkg.BuildEmpresaOdontologiaDashboard(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar el dashboard de odontologia", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "pacientes":
				rows, err := dbpkg.ListEmpresaOdontologiaPacientes(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudieron listar los pacientes", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "profesionales":
				rows, err := dbpkg.ListEmpresaOdontologiaProfesionales(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudieron listar los profesionales", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "consultorios":
				rows, err := dbpkg.ListEmpresaOdontologiaConsultorios(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudieron listar los consultorios", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "citas":
				rows, err := dbpkg.ListEmpresaOdontologiaCitasByFecha(dbEmp, empresaID, strings.TrimSpace(r.URL.Query().Get("fecha")))
				if err != nil {
					http.Error(w, "No se pudieron listar las citas", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "historias":
				rows, err := dbpkg.ListEmpresaOdontologiaHistorias(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudieron listar las historias", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "odontogramas":
				rows, err := dbpkg.ListEmpresaOdontologiaOdontogramas(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudieron listar los odontogramas", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "tratamientos":
				rows, err := dbpkg.ListEmpresaOdontologiaTratamientos(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudieron listar los tratamientos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "presupuestos":
				rows, err := dbpkg.ListEmpresaOdontologiaPresupuestos(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudieron listar los presupuestos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "pagos":
				rows, err := dbpkg.ListEmpresaOdontologiaPagos(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudieron listar los pagos", http.StatusInternalServerError)
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
			case "sincronizar_nucleo":
				resumen, err := dbpkg.SyncEmpresaOdontologiaNucleo(dbEmp, empresaID, adminEmail)
				if err != nil {
					http.Error(w, "No se pudo sincronizar odontologia con el nucleo", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "integracion": resumen})
				return
			case "pacientes":
				var payload dbpkg.EmpresaOdontologiaPaciente
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.CreateEmpresaOdontologiaPaciente(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "profesionales":
				var payload dbpkg.EmpresaOdontologiaProfesional
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.CreateEmpresaOdontologiaProfesional(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "consultorios":
				var payload dbpkg.EmpresaOdontologiaConsultorio
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.CreateEmpresaOdontologiaConsultorio(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "citas":
				var payload dbpkg.EmpresaOdontologiaCita
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.CreateEmpresaOdontologiaCita(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "historias":
				var payload dbpkg.EmpresaOdontologiaHistoria
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.CreateEmpresaOdontologiaHistoria(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "odontogramas":
				var payload dbpkg.EmpresaOdontologiaOdontograma
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.CreateEmpresaOdontologiaOdontograma(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "tratamientos":
				var payload dbpkg.EmpresaOdontologiaTratamiento
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.CreateEmpresaOdontologiaTratamiento(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "presupuestos":
				var payload dbpkg.EmpresaOdontologiaPresupuesto
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.CreateEmpresaOdontologiaPresupuesto(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "pagos":
				var payload dbpkg.EmpresaOdontologiaPago
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.CreateEmpresaOdontologiaPago(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			default:
				http.Error(w, "action no soportada", http.StatusBadRequest)
				return
			}

		case http.MethodPut:
			switch action {
			case "estado_paciente":
				id, _ := parseInt64QueryOptional(r, "id")
				estado := strings.TrimSpace(r.URL.Query().Get("estado"))
				if id <= 0 || estado == "" {
					http.Error(w, "id y estado son obligatorios", http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetEmpresaOdontologiaPacienteEstado(dbEmp, empresaID, id, estado); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			case "estado_profesional":
				id, _ := parseInt64QueryOptional(r, "id")
				estado := strings.TrimSpace(r.URL.Query().Get("estado"))
				if id <= 0 || estado == "" {
					http.Error(w, "id y estado son obligatorios", http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetEmpresaOdontologiaProfesionalEstado(dbEmp, empresaID, id, estado); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			case "estado_consultorio":
				id, _ := parseInt64QueryOptional(r, "id")
				estado := strings.TrimSpace(r.URL.Query().Get("estado"))
				if id <= 0 || estado == "" {
					http.Error(w, "id y estado son obligatorios", http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetEmpresaOdontologiaConsultorioEstado(dbEmp, empresaID, id, estado); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			case "estado_cita":
				id, _ := parseInt64QueryOptional(r, "id")
				estado := strings.TrimSpace(r.URL.Query().Get("estado"))
				if id <= 0 || estado == "" {
					http.Error(w, "id y estado son obligatorios", http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetEmpresaOdontologiaCitaEstado(dbEmp, empresaID, id, estado); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			case "estado_tratamiento":
				id, _ := parseInt64QueryOptional(r, "id")
				estado := strings.TrimSpace(r.URL.Query().Get("estado"))
				if id <= 0 || estado == "" {
					http.Error(w, "id y estado son obligatorios", http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetEmpresaOdontologiaTratamientoEstado(dbEmp, empresaID, id, estado); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			case "estado_presupuesto":
				id, _ := parseInt64QueryOptional(r, "id")
				estado := strings.TrimSpace(r.URL.Query().Get("estado"))
				if id <= 0 || estado == "" {
					http.Error(w, "id y estado son obligatorios", http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetEmpresaOdontologiaPresupuestoEstado(dbEmp, empresaID, id, estado); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
			default:
				http.Error(w, "action no soportada", http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}
