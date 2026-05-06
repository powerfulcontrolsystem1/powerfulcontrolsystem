package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaCobranzaHandler(dbEmp *sql.DB) http.HandlerFunc {
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
		usuario := strings.TrimSpace(adminEmailFromRequest(r))
		if usuario == "" {
			usuario = "sistema"
		}

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "dashboard":
				row, err := dbpkg.BuildEmpresaCobranzaDashboard(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar gestion de cobranza", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "cuentas":
				rows, err := dbpkg.ListEmpresaCobranzaCuentas(dbEmp, empresaID, dbpkg.EmpresaCobranzaCuentaFiltro{
					Estado:  r.URL.Query().Get("estado"),
					Query:   r.URL.Query().Get("q"),
					MoraMin: intQuery(r, "mora_min"),
					Limit:   300,
				})
				if err != nil {
					http.Error(w, "No se pudieron listar cuentas por cobrar", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "plantillas":
				rows, err := dbpkg.ListEmpresaCobranzaPlantillas(dbEmp, empresaID, 300)
				if err != nil {
					http.Error(w, "No se pudieron listar plantillas de cobranza", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "campanas":
				rows, err := dbpkg.ListEmpresaCobranzaCampanas(dbEmp, empresaID, 300)
				if err != nil {
					http.Error(w, "No se pudieron listar campanas de cobranza", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "gestiones":
				rows, err := dbpkg.ListEmpresaCobranzaGestiones(dbEmp, empresaID, 300)
				if err != nil {
					http.Error(w, "No se pudieron listar gestiones de cobranza", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "promesas":
				rows, err := dbpkg.ListEmpresaCobranzaPromesas(dbEmp, empresaID, r.URL.Query().Get("estado"), 300)
				if err != nil {
					http.Error(w, "No se pudieron listar promesas de pago", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}

		case http.MethodPost, http.MethodPut:
			switch action {
			case "plantilla":
				var payload dbpkg.EmpresaCobranzaPlantilla
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.Usuario = usuario
				id, err := dbpkg.UpsertEmpresaCobranzaPlantilla(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			case "campana":
				var payload dbpkg.EmpresaCobranzaCampana
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.Usuario = usuario
				id, err := dbpkg.UpsertEmpresaCobranzaCampana(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			case "gestion", "simular_envio":
				var payload dbpkg.EmpresaCobranzaGestion
				_ = json.NewDecoder(r.Body).Decode(&payload)
				payload.EmpresaID = empresaID
				payload.Usuario = usuario
				if action == "simular_envio" {
					payload.Resultado = "enviado_simulado"
					if payload.Mensaje == "" {
						payload.Mensaje = "Mensaje de cobranza programado desde simulacion interna. No se envio por proveedor externo."
					}
				}
				row, promesa, err := dbpkg.CreateEmpresaCobranzaGestion(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "gestion": row, "promesa": promesa})
				return
			case "promesa":
				var payload dbpkg.EmpresaCobranzaPromesa
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.Usuario = usuario
				id, err := dbpkg.UpsertEmpresaCobranzaPromesa(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			case "marcar_promesa":
				var payload struct {
					ID            int64  `json:"id"`
					Estado        string `json:"estado"`
					Observaciones string `json:"observaciones"`
				}
				_ = json.NewDecoder(r.Body).Decode(&payload)
				if payload.ID <= 0 {
					payload.ID = int64Query(r, "id")
				}
				if payload.Estado == "" {
					payload.Estado = r.URL.Query().Get("estado")
				}
				row, err := dbpkg.UpdateEmpresaCobranzaPromesaEstado(dbEmp, empresaID, payload.ID, payload.Estado, usuario, payload.Observaciones)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "promesa": row})
				return
			case "seed_demo":
				if err := dbpkg.SeedEmpresaCobranzaDemo(dbEmp, empresaID, usuario); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			}
		}
		http.Error(w, fmt.Sprintf("Metodo o accion no permitida: %s", action), http.StatusMethodNotAllowed)
	}
}
