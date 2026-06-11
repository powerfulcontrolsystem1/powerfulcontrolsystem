package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp *sql.DB, r *http.Request, empresaID int64, modulo, accion, recurso string, recursoID int64, statusCode int, metadata map[string]interface{}, observaciones string) {
	if dbEmp == nil || r == nil || empresaID <= 0 {
		return
	}
	modulo = strings.TrimSpace(modulo)
	accion = strings.TrimSpace(accion)
	if modulo == "" || accion == "" {
		return
	}
	if metadata == nil {
		metadata = map[string]interface{}{}
	}
	metadata["auditoria_especifica"] = true
	metadata["endpoint"] = r.URL.Path
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		metadataJSON = []byte(`{"auditoria_especifica":true}`)
	}
	if statusCode <= 0 {
		statusCode = http.StatusOK
	}
	go func() {
		_, err := dbpkg.CreateEmpresaAuditoriaEvento(dbEmp, dbpkg.EmpresaAuditoriaEvento{
			EmpresaID:      empresaID,
			Modulo:         modulo,
			Accion:         accion,
			Recurso:        recurso,
			RecursoID:      recursoID,
			MetodoHTTP:     r.Method,
			Endpoint:       r.URL.Path,
			Resultado:      resolveAuditoriaResultado(statusCode),
			CodigoHTTP:     int64(statusCode),
			RequestID:      resolveAuditoriaRequestID(r),
			IPOrigen:       resolveAuditoriaIP(r),
			UserAgent:      r.UserAgent(),
			MetadataJSON:   string(metadataJSON),
			RetencionDias:  0,
			UsuarioCreador: adminEmailFromRequest(r),
			Estado:         "activo",
			Observaciones:  sanitizeAuditMetadataText(observaciones, 500),
		})
		if err != nil {
			log.Printf("[auditoria] modulo_especifico empresa_id=%d modulo=%s accion=%s error: %v", empresaID, modulo, accion, err)
		}
	}()
}
