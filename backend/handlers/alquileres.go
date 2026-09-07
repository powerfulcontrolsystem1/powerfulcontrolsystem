package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaAlquileresHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "" {
			action = "dashboard"
		}
		adminEmail := strings.TrimSpace(adminEmailFromRequest(r))

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "dashboard":
				row, err := dbpkg.BuildEmpresaAlquilerDashboard(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar el dashboard de alquileres", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "config":
				row, err := dbpkg.GetEmpresaAlquilerConfig(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar la configuracion", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "categorias":
				rows, err := dbpkg.ListEmpresaAlquilerCategorias(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar las categorias", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "activos":
				rows, err := dbpkg.ListEmpresaAlquilerActivos(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar los activos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "tarifas":
				rows, err := dbpkg.ListEmpresaAlquilerTarifas(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar las tarifas", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "contratos":
				rows, err := dbpkg.ListEmpresaAlquilerContratos(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar los contratos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "mantenimientos":
				rows, err := dbpkg.ListEmpresaAlquilerMantenimientos(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar los mantenimientos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "ubicaciones":
				contratoID, _ := parseInt64QueryOptional(r, "contrato_id")
				rows, err := dbpkg.ListEmpresaAlquilerUbicaciones(dbEmp, empresaID, contratoID)
				if err != nil {
					http.Error(w, "No se pudo consultar el mapa operativo", http.StatusInternalServerError)
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
			case "config":
				var payload dbpkg.EmpresaAlquilerConfig
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				if err := dbpkg.UpsertEmpresaAlquilerConfig(dbEmp, payload); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "categorias":
				var payload dbpkg.EmpresaAlquilerCategoria
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.CreateEmpresaAlquilerCategoria(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "activos":
				var payload dbpkg.EmpresaAlquilerActivo
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.CreateEmpresaAlquilerActivo(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "tarifas":
				var payload dbpkg.EmpresaAlquilerTarifa
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.CreateEmpresaAlquilerTarifa(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "contratos":
				var payload dbpkg.EmpresaAlquilerContrato
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.CreateEmpresaAlquilerContrato(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "mantenimientos":
				var payload dbpkg.EmpresaAlquilerMantenimiento
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.CreateEmpresaAlquilerMantenimiento(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "ubicaciones":
				var payload dbpkg.EmpresaAlquilerUbicacion
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.CreateEmpresaAlquilerUbicacion(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "cambiar_estado":
				var payload struct {
					ContratoID    int64  `json:"contrato_id"`
					Estado        string `json:"estado"`
					Observaciones string `json:"observaciones"`
					Responsable   string `json:"responsable"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if payload.ContratoID <= 0 {
					http.Error(w, "contrato_id es obligatorio", http.StatusBadRequest)
					return
				}
				responsable := strings.TrimSpace(payload.Responsable)
				if responsable == "" {
					responsable = adminEmail
				}
				if err := dbpkg.UpdateEmpresaAlquilerContratoEstado(dbEmp, empresaID, payload.ContratoID, payload.Estado, responsable, payload.Observaciones); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "seed_demo":
				if err := dbpkg.SeedEmpresaAlquilerProfesionalData(dbEmp, empresaID, adminEmail); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			default:
				http.Error(w, "action no soportada", http.StatusBadRequest)
				return
			}
		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
	}
}
