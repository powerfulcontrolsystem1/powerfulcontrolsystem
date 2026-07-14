package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaAIOpenAIProveedorConfigHandler manages an optional customer-owned
// OpenAI key. It is wrapped with seguridad permissions by the route registrar.
func EmpresaAIOpenAIProveedorConfigHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
				return
			}
			cfg, err := dbpkg.GetEmpresaAIOpenAIProviderConfig(dbEmp, empresaID)
			if err != nil {
				http.Error(w, "No se pudo consultar la configuracion IA empresarial", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok": true, "empresa_id": empresaID, "openai_propio": cfg,
				"nota": "La clave se cifra en servidor y nunca se devuelve al navegador.",
			})
		case http.MethodPut:
			var payload struct {
				EmpresaID  int64  `json:"empresa_id"`
				Habilitado bool   `json:"habilitado"`
				APIKey     string `json:"api_key"`
				Reemplazar bool   `json:"reemplazar_clave"`
				Eliminar   bool   `json:"eliminar_clave"`
			}
			if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 {
				payload.EmpresaID, _ = parseInt64QueryOptional(r, "empresa_id")
			}
			if payload.EmpresaID <= 0 {
				http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
				return
			}
			payload.APIKey = strings.TrimSpace(payload.APIKey)
			cfg, err := dbpkg.UpsertEmpresaAIOpenAIProviderConfig(dbEmp, payload.EmpresaID, payload.Habilitado, payload.APIKey, adminEmailFromRequest(r), payload.Reemplazar || payload.APIKey != "", payload.Eliminar)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok": true, "empresa_id": payload.EmpresaID, "openai_propio": cfg,
				"nota": "Configuracion guardada. La clave no se muestra ni se registra en auditoria.",
			})
		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		}
	}
}
