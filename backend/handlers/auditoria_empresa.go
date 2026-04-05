package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	utilspkg "github.com/you/pos-backend/utils"
)

type auditCaptureResponseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *auditCaptureResponseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *auditCaptureResponseWriter) Write(p []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	return rw.ResponseWriter.Write(p)
}

// EmpresaAuditoriaEventosHandler expone consulta y depuracion manual de auditoria por empresa.
func EmpresaAuditoriaEventosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			limit, err := parseIntQueryOptional(r, "limit")
			if err != nil {
				http.Error(w, "limit invalido", http.StatusBadRequest)
				return
			}
			offset, err := parseIntQueryOptional(r, "offset")
			if err != nil {
				http.Error(w, "offset invalido", http.StatusBadRequest)
				return
			}
			if limit < 0 {
				http.Error(w, "limit invalido", http.StatusBadRequest)
				return
			}
			if offset < 0 {
				http.Error(w, "offset invalido", http.StatusBadRequest)
				return
			}
			recursoID, err := parseInt64QueryOptional(r, "recurso_id")
			if err != nil {
				http.Error(w, "recurso_id invalido", http.StatusBadRequest)
				return
			}
			if recursoID < 0 {
				http.Error(w, "recurso_id invalido", http.StatusBadRequest)
				return
			}
			codigoHTTP, err := parseInt64QueryOptional(r, "codigo_http")
			if err != nil {
				http.Error(w, "codigo_http invalido", http.StatusBadRequest)
				return
			}
			if codigoHTTP < 0 || (codigoHTTP > 0 && (codigoHTTP < 100 || codigoHTTP > 599)) {
				http.Error(w, "codigo_http invalido", http.StatusBadRequest)
				return
			}

			desde, err := normalizeAuditoriaDateTime(strings.TrimSpace(r.URL.Query().Get("desde")), false)
			if err != nil {
				http.Error(w, "desde invalido", http.StatusBadRequest)
				return
			}
			hasta, err := normalizeAuditoriaDateTime(strings.TrimSpace(r.URL.Query().Get("hasta")), true)
			if err != nil {
				http.Error(w, "hasta invalido", http.StatusBadRequest)
				return
			}
			if desde != "" && hasta != "" {
				desdeTime, _ := time.ParseInLocation("2006-01-02 15:04:05", desde, time.Local)
				hastaTime, _ := time.ParseInLocation("2006-01-02 15:04:05", hasta, time.Local)
				if desdeTime.After(hastaTime) {
					http.Error(w, "rango de fechas invalido", http.StatusBadRequest)
					return
				}
			}

			filter := dbpkg.EmpresaAuditoriaEventoFilter{
				Modulo:          strings.TrimSpace(r.URL.Query().Get("modulo")),
				Accion:          strings.TrimSpace(r.URL.Query().Get("accion")),
				MetodoHTTP:      strings.TrimSpace(r.URL.Query().Get("metodo_http")),
				Recurso:         strings.TrimSpace(r.URL.Query().Get("recurso")),
				Endpoint:        strings.TrimSpace(r.URL.Query().Get("endpoint")),
				Search:          strings.TrimSpace(r.URL.Query().Get("search")),
				RecursoID:       recursoID,
				CodigoHTTP:      codigoHTTP,
				Resultado:       strings.TrimSpace(r.URL.Query().Get("resultado")),
				UsuarioCreador:  strings.TrimSpace(r.URL.Query().Get("usuario")),
				RequestID:       strings.TrimSpace(r.URL.Query().Get("request_id")),
				Desde:           desde,
				Hasta:           hasta,
				IncludeInactive: queryBool(r, "include_inactive"),
				Limit:           limit,
				Offset:          offset,
			}

			total, err := dbpkg.CountEmpresaAuditoriaEventos(dbEmp, empresaID, filter)
			if err != nil {
				http.Error(w, "No se pudo consultar el total de auditoria", http.StatusInternalServerError)
				return
			}

			rows, err := dbpkg.ListEmpresaAuditoriaEventos(dbEmp, empresaID, filter)
			if err != nil {
				http.Error(w, "No se pudo consultar la auditoria", http.StatusInternalServerError)
				return
			}
			pageLimit, pageOffset := normalizeAuditoriaPage(limit, offset)
			w.Header().Set("X-Total-Count", strconv.FormatInt(total, 10))
			w.Header().Set("X-Page-Limit", strconv.Itoa(pageLimit))
			w.Header().Set("X-Page-Offset", strconv.Itoa(pageOffset))
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPut, http.MethodPost:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "" {
				action = "retener"
			}
			if action != "retener" && action != "purgar" {
				http.Error(w, "action invalida", http.StatusBadRequest)
				return
			}
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			retencionDias, err := parseInt64QueryOptional(r, "retencion_dias")
			if err != nil {
				http.Error(w, "retencion_dias invalido", http.StatusBadRequest)
				return
			}
			if retencionDias <= 0 {
				retencionDias, _ = parseInt64QueryOptional(r, "dias")
			}
			eliminados, err := dbpkg.PurgeEmpresaAuditoriaEventos(dbEmp, empresaID, retencionDias)
			if err != nil {
				http.Error(w, "No se pudo depurar la auditoria", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":             true,
				"empresa_id":     empresaID,
				"action":         action,
				"eliminados":     eliminados,
				"retencion_dias": normalizeRetencionDiasForHandler(retencionDias),
			})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func registrarAuditoriaOperacionNoBloqueante(dbEmp *sql.DB, r *http.Request, empresaID int64, modulo, permissionAction string, statusCode int, elapsed time.Duration) {
	if dbEmp == nil {
		return
	}
	if !accionCriticaParaAuditoria(permissionAction) {
		return
	}
	if empresaID <= 0 {
		return
	}

	metadata := map[string]interface{}{
		"permission_action": strings.ToUpper(strings.TrimSpace(permissionAction)),
		"duracion_ms":       elapsed.Milliseconds(),
	}
	if queryAction := strings.TrimSpace(r.URL.Query().Get("action")); queryAction != "" {
		metadata["query_action"] = strings.ToLower(queryAction)
	}
	if rid, err := parseInt64QueryOptional(r, "id"); err == nil && rid > 0 {
		metadata["recurso_id_query"] = rid
	}
	if carritoID, err := parseInt64QueryOptional(r, "carrito_id"); err == nil && carritoID > 0 {
		metadata["carrito_id"] = carritoID
	}
	if proveedorID, err := parseInt64QueryOptional(r, "proveedor_id"); err == nil && proveedorID > 0 {
		metadata["proveedor_id"] = proveedorID
	}
	if entidadID, err := parseInt64QueryOptional(r, "entidad_id"); err == nil && entidadID > 0 {
		metadata["entidad_id"] = entidadID
	}
	if documentoCodigo := strings.TrimSpace(r.URL.Query().Get("documento_codigo")); documentoCodigo != "" {
		metadata["documento_codigo"] = documentoCodigo
	}
	metaJSON, err := json.Marshal(metadata)
	if err != nil {
		metaJSON = []byte(`{"marshal_error":"metadata"}`)
	}

	auditoria := dbpkg.EmpresaAuditoriaEvento{
		EmpresaID:      empresaID,
		Modulo:         strings.TrimSpace(modulo),
		Accion:         resolveAuditoriaAccion(r, permissionAction),
		Recurso:        resolveAuditoriaRecursoDesdePath(r.URL.Path),
		RecursoID:      resolveAuditoriaRecursoID(r),
		MetodoHTTP:     strings.ToUpper(strings.TrimSpace(r.Method)),
		Endpoint:       strings.TrimSpace(r.URL.Path),
		Resultado:      resolveAuditoriaResultado(statusCode),
		CodigoHTTP:     int64(statusCode),
		RequestID:      resolveAuditoriaRequestID(r),
		IPOrigen:       resolveAuditoriaIP(r),
		UserAgent:      strings.TrimSpace(r.UserAgent()),
		MetadataJSON:   string(metaJSON),
		RetencionDias:  normalizeRetencionDiasForHandler(0),
		UsuarioCreador: strings.TrimSpace(adminEmailFromRequest(r)),
		Estado:         "activo",
		Observaciones:  "auditoria automatica de accion critica",
	}

	if _, err := dbpkg.CreateEmpresaAuditoriaEvento(dbEmp, auditoria); err != nil {
		log.Printf("[auditoria] no se pudo registrar evento empresa_id=%d modulo=%s accion=%s error=%v", empresaID, modulo, auditoria.Accion, err)
	}
}

func accionCriticaParaAuditoria(permissionAction string) bool {
	switch strings.ToUpper(strings.TrimSpace(permissionAction)) {
	case "C", "U", "D", "A":
		return true
	default:
		return false
	}
}

func resolveAuditoriaAccion(r *http.Request, permissionAction string) string {
	if q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action"))); q != "" {
		q = strings.ReplaceAll(q, "-", "_")
		q = strings.ReplaceAll(q, " ", "_")
		return q
	}
	switch strings.ToUpper(strings.TrimSpace(permissionAction)) {
	case "C":
		return "crear"
	case "U":
		return "actualizar"
	case "D":
		return "eliminar"
	case "A":
		return "aprobar"
	default:
		return "accion_critica"
	}
}

func resolveAuditoriaRecursoDesdePath(path string) string {
	v := strings.TrimSpace(path)
	v = strings.TrimPrefix(v, "/")
	v = strings.TrimPrefix(v, "api/")
	v = strings.TrimPrefix(v, "empresa/")
	v = strings.Trim(v, "/")
	if v == "" {
		return "empresa"
	}
	return v
}

func resolveAuditoriaRecursoID(r *http.Request) int64 {
	keys := []string{"id", "carrito_id", "item_id", "proveedor_id", "entidad_id", "sucursal_id"}
	for _, key := range keys {
		if id, err := parseInt64QueryOptional(r, key); err == nil && id > 0 {
			return id
		}
	}
	return 0
}

func resolveAuditoriaResultado(statusCode int) string {
	if statusCode >= 400 {
		return "error"
	}
	return "ok"
}

func resolveAuditoriaRequestID(r *http.Request) string {
	if v := strings.TrimSpace(utilspkg.RequestIDFromContext(r.Context())); v != "" {
		return v
	}
	if v := strings.TrimSpace(r.Header.Get("X-Request-ID")); v != "" {
		return v
	}
	return ""
}

func resolveAuditoriaIP(r *http.Request) string {
	if xfwd := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xfwd != "" {
		parts := strings.Split(xfwd, ",")
		for _, p := range parts {
			if ip := strings.TrimSpace(p); ip != "" {
				return ip
			}
		}
	}
	remote := strings.TrimSpace(r.RemoteAddr)
	if remote == "" {
		return ""
	}
	host, _, err := net.SplitHostPort(remote)
	if err == nil {
		return host
	}
	return remote
}

func normalizeRetencionDiasForHandler(days int64) int64 {
	if days <= 0 {
		return 180
	}
	if days > 3650 {
		return 3650
	}
	return days
}

func normalizeAuditoriaPage(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}
	if offset < 0 {
		offset = 0
	}
	if offset > 500000 {
		offset = 500000
	}
	return limit, offset
}

func normalizeAuditoriaDateTime(raw string, endOfDay bool) (string, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return "", nil
	}

	if t, err := time.ParseInLocation("2006-01-02", v, time.Local); err == nil {
		if endOfDay {
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
		return t.Format("2006-01-02 15:04:05"), nil
	}

	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, v, time.Local); err == nil {
			return t.Format("2006-01-02 15:04:05"), nil
		}
	}

	if t, err := time.Parse(time.RFC3339, v); err == nil {
		return t.Local().Format("2006-01-02 15:04:05"), nil
	}

	return "", fmt.Errorf("formato de fecha invalido")
}
