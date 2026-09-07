package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

// EmpresaDBAdminHandler expone CRUD controlado sobre tablas por empresa_id.
// Reglas:
// - Solo rol administrador o super_administrador
// - Toda mutación requiere cabecera X-PCS-Confirmed=1
// - DELETE requiere adicionalmente X-PCS-Confirm-Text=ELIMINAR
// - Solo opera sobre tablas con columna empresa_id (scoping obligatorio)
func EmpresaDBAdminHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil || empresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}

		role := strings.TrimSpace(adminRoleFromRequest(r))
		if !strings.EqualFold(role, "administrador") && !strings.EqualFold(role, "super_administrador") {
			http.Error(w, "forbidden: requiere rol administrador", http.StatusForbidden)
			return
		}

		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "" {
			action = "schema"
		}

		requireConfirmed := func(isDelete bool) bool {
			if strings.TrimSpace(r.Header.Get("X-PCS-Confirmed")) != "1" {
				http.Error(w, "confirmacion requerida (X-PCS-Confirmed=1)", http.StatusPreconditionRequired)
				return false
			}
			if isDelete {
				if strings.TrimSpace(strings.ToUpper(r.Header.Get("X-PCS-Confirm-Text"))) != "ELIMINAR" {
					http.Error(w, "confirmacion adicional requerida para DELETE (X-PCS-Confirm-Text=ELIMINAR)", http.StatusPreconditionRequired)
					return false
				}
			}
			return true
		}

		start := time.Now()
		audit := func(ok bool, code int, meta interface{}, obs string) {
			result := "ok"
			if !ok {
				result = "error"
			}
			_, _ = dbpkg.CreateEmpresaAuditoriaEvento(dbEmp, dbpkg.EmpresaAuditoriaEvento{
				EmpresaID:      empresaID,
				Modulo:         "chat_ia_db",
				Accion:         action,
				Recurso:        "db_admin",
				RecursoID:      0,
				MetodoHTTP:     r.Method,
				Endpoint:       r.URL.Path,
				Resultado:      result,
				CodigoHTTP:     int64(code),
				RequestID:      utils.RequestIDFromContext(r.Context()),
				IPOrigen:       strings.TrimSpace(r.RemoteAddr),
				UserAgent:      r.UserAgent(),
				MetadataJSON:   dbpkg.DBAdminMarshalSafeMetadata(meta),
				UsuarioCreador: strings.TrimSpace(adminEmailFromRequest(r)),
				Observaciones:  strings.TrimSpace(obs),
			})
			_ = start // placeholder: disponible para futuras métricas
		}

		switch r.Method {
		case http.MethodGet:
			if action != "schema" && action != "tables" && action != "columns" {
				http.Error(w, "action invalida", http.StatusBadRequest)
				return
			}
			if action == "schema" || action == "tables" {
				tables, err := dbpkg.DBAdminListEmpresaTables(dbEmp)
				if err != nil {
					audit(false, 500, map[string]interface{}{"error": err.Error()}, "list tables")
					http.Error(w, "No se pudo listar tablas", http.StatusInternalServerError)
					return
				}
				audit(true, 200, map[string]interface{}{"tables_count": len(tables)}, "ok")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "empresa_id": empresaID, "tables": tables})
				return
			}
			if action == "columns" {
				table := strings.TrimSpace(r.URL.Query().Get("table"))
				cols, err := dbpkg.DBAdminGetTableColumns(dbEmp, table)
				if err != nil {
					audit(false, 400, map[string]interface{}{"table": table, "error": err.Error()}, "get columns")
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				audit(true, 200, map[string]interface{}{"table": table, "columns": len(cols)}, "ok")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "empresa_id": empresaID, "table": table, "columns": cols})
				return
			}
			http.Error(w, "action invalida", http.StatusBadRequest)
			return

		case http.MethodPost:
			if action != "select" && action != "insert" && action != "update" && action != "delete" {
				http.Error(w, "action invalida", http.StatusBadRequest)
				return
			}
			if !requireConfirmed(action != "select" && action == "delete") {
				return
			}

			if action == "select" {
				var payload dbpkg.DBAdminSelectRequest
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				rows, err := dbpkg.DBAdminSelect(dbEmp, empresaID, payload)
				if err != nil {
					audit(false, 400, map[string]interface{}{"table": payload.Table, "error": err.Error()}, "select")
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				audit(true, 200, map[string]interface{}{"table": payload.Table, "rows": len(rows)}, "ok")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "empresa_id": empresaID, "rows": rows})
				return
			}

			var payload dbpkg.DBAdminMutationRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}

			switch action {
			case "insert":
				id, err := dbpkg.DBAdminInsert(dbEmp, empresaID, payload)
				if err != nil {
					audit(false, 400, map[string]interface{}{"table": payload.Table, "error": err.Error()}, "insert")
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				audit(true, 200, map[string]interface{}{"table": payload.Table, "id": id}, "ok")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "empresa_id": empresaID, "id": id})
				return
			case "update":
				affected, err := dbpkg.DBAdminUpdate(dbEmp, empresaID, payload)
				if err != nil {
					audit(false, 400, map[string]interface{}{"table": payload.Table, "error": err.Error()}, "update")
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				audit(true, 200, map[string]interface{}{"table": payload.Table, "affected": affected}, "ok")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "empresa_id": empresaID, "affected": affected})
				return
			case "delete":
				if !requireConfirmed(true) {
					return
				}
				affected, err := dbpkg.DBAdminDelete(dbEmp, empresaID, payload)
				if err != nil {
					audit(false, 400, map[string]interface{}{"table": payload.Table, "error": err.Error()}, "delete")
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				audit(true, 200, map[string]interface{}{"table": payload.Table, "affected": affected}, "ok")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "empresa_id": empresaID, "affected": affected})
				return
			}

			http.Error(w, "action invalida", http.StatusBadRequest)
			return
		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
	}
}

