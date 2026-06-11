package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaProduccionMRPHandler(dbEmp *sql.DB) http.HandlerFunc {
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
				row, err := dbpkg.BuildEmpresaProduccionMRPDashboard(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar produccion/MRP", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "config":
				row, err := dbpkg.GetEmpresaProduccionMRPConfig(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar configuracion", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "recetas":
				rows, err := dbpkg.ListEmpresaProduccionRecetas(dbEmp, empresaID, strings.TrimSpace(r.URL.Query().Get("estado")), 250)
				if err != nil {
					http.Error(w, "No se pudieron listar recetas", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "componentes":
				recetaID, err := parseInt64Query(r, "receta_id")
				if err != nil {
					http.Error(w, "receta_id es obligatorio", http.StatusBadRequest)
					return
				}
				rows, err := dbpkg.ListEmpresaProduccionComponentes(dbEmp, empresaID, recetaID)
				if err != nil {
					http.Error(w, "No se pudieron listar componentes", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "ordenes":
				rows, err := dbpkg.ListEmpresaProduccionOrdenes(dbEmp, empresaID, strings.TrimSpace(r.URL.Query().Get("estado")), 250)
				if err != nil {
					http.Error(w, "No se pudieron listar ordenes", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "orden":
				id, err := parseInt64Query(r, "orden_id")
				if err != nil {
					http.Error(w, "orden_id es obligatorio", http.StatusBadRequest)
					return
				}
				row, err := dbpkg.GetEmpresaProduccionOrdenByID(dbEmp, empresaID, id)
				if err != nil {
					http.Error(w, "Orden no encontrada", http.StatusNotFound)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "consumos":
				ordenID, _ := parseInt64QueryOptional(r, "orden_id")
				rows, err := dbpkg.ListEmpresaProduccionConsumos(dbEmp, empresaID, ordenID, 250)
				if err != nil {
					http.Error(w, "No se pudieron listar consumos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "calidad":
				ordenID, _ := parseInt64QueryOptional(r, "orden_id")
				rows, err := dbpkg.ListEmpresaProduccionCalidad(dbEmp, empresaID, ordenID, 200)
				if err != nil {
					http.Error(w, "No se pudo listar calidad", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "mrp_plan":
				rows, err := dbpkg.ListEmpresaProduccionMRPPlan(dbEmp, empresaID, strings.TrimSpace(r.URL.Query().Get("periodo")), 250)
				if err != nil {
					http.Error(w, "No se pudo listar plan MRP", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "catalogo_recetas_vendibles":
				rows, err := dbpkg.GetRecetasProductosByEmpresa(dbEmp, empresaID, strings.TrimSpace(r.URL.Query().Get("q")), "activo", false, 100, 0)
				if err != nil {
					http.Error(w, "No se pudieron listar recetas vendibles", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}

		case http.MethodPost:
			switch action {
			case "config":
				var payload dbpkg.EmpresaProduccionMRPConfig
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, adminEmail
				if err := dbpkg.UpsertEmpresaProduccionMRPConfig(dbEmp, payload); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "recetas":
				var payload dbpkg.EmpresaProduccionReceta
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, adminEmail
				id, err := dbpkg.UpsertEmpresaProduccionReceta(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "componentes":
				var payload struct {
					RecetaID    int64                               `json:"receta_id"`
					Componentes []dbpkg.EmpresaProduccionComponente `json:"componentes"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if err := dbpkg.ReplaceEmpresaProduccionComponentes(dbEmp, empresaID, payload.RecetaID, payload.Componentes); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "ordenes":
				var payload dbpkg.EmpresaProduccionOrden
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, adminEmail
				row, err := dbpkg.CreateEmpresaProduccionOrden(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, row)
				return
			case "orden_estado":
				var payload struct {
					OrdenID int64  `json:"orden_id"`
					Estado  string `json:"estado"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				row, err := dbpkg.CambiarEstadoEmpresaProduccionOrden(dbEmp, empresaID, payload.OrdenID, payload.Estado, adminEmail)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "consumos":
				var payload dbpkg.EmpresaProduccionConsumo
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, adminEmail
				id, err := dbpkg.RegistrarEmpresaProduccionConsumo(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "calidad":
				var payload dbpkg.EmpresaProduccionCalidad
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				id, err := dbpkg.RegistrarEmpresaProduccionCalidad(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "generar_mrp":
				var payload struct {
					Periodo string `json:"periodo"`
				}
				_ = json.NewDecoder(r.Body).Decode(&payload)
				rows, err := dbpkg.GenerarEmpresaProduccionMRPPlan(dbEmp, empresaID, payload.Periodo, adminEmail)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "seed_demo":
				if err := dbpkg.SeedEmpresaProduccionMRPDemo(dbEmp, empresaID, adminEmail); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "import_receta_producto":
				var payload struct {
					RecetaProductoID int64 `json:"receta_producto_id"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				row, err := dbpkg.ImportEmpresaProduccionRecetaFromRecetaProducto(dbEmp, empresaID, payload.RecetaProductoID, adminEmail)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, row)
				return
			}

		case http.MethodPut:
			switch action {
			case "recetas":
				var payload dbpkg.EmpresaProduccionReceta
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				id, err := dbpkg.UpsertEmpresaProduccionReceta(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			case "ordenes":
				var payload dbpkg.EmpresaProduccionOrden
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				if err := dbpkg.UpdateEmpresaProduccionOrden(dbEmp, payload); err != nil {
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
