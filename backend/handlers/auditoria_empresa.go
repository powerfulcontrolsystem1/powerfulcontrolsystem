package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/csv"
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

type auditoriaForenseBase struct {
	ID             int64  `json:"id"`
	FechaEvento    string `json:"fecha_evento"`
	Modulo         string `json:"modulo"`
	Accion         string `json:"accion"`
	Recurso        string `json:"recurso"`
	RecursoID      int64  `json:"recurso_id"`
	MetodoHTTP     string `json:"metodo_http"`
	Endpoint       string `json:"endpoint"`
	Resultado      string `json:"resultado"`
	CodigoHTTP     int64  `json:"codigo_http"`
	RequestID      string `json:"request_id"`
	UsuarioCreador string `json:"usuario_creador"`
	IPOrigen       string `json:"ip_origen"`
	Observaciones  string `json:"observaciones"`
	MetadataJSON   string `json:"metadata_json"`
}

type auditoriaForenseRegistro struct {
	Indice             int                  `json:"indice"`
	Base               auditoriaForenseBase `json:"base"`
	HashRegistro       string               `json:"hash_registro"`
	HashCadenaAnterior string               `json:"hash_cadena_anterior"`
	HashCadena         string               `json:"hash_cadena"`
}

type auditoriaForenseManifest struct {
	EmpresaID          int64                  `json:"empresa_id"`
	GeneradoEn         string                 `json:"generado_en"`
	AlgoritmoHash      string                 `json:"algoritmo_hash"`
	TotalCoincidencias int64                  `json:"total_coincidencias"`
	TotalRegistros     int                    `json:"total_registros"`
	HashGlobal         string                 `json:"hash_global"`
	HashCadenaFinal    string                 `json:"hash_cadena_final"`
	Filtros            map[string]interface{} `json:"filtros"`
}

type auditoriaForenseExportPayload struct {
	OK        bool                       `json:"ok"`
	Manifest  auditoriaForenseManifest   `json:"manifest"`
	Registros []auditoriaForenseRegistro `json:"registros"`
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
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
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
			if action == "export_forense" || action == "forense_export" || action == "cadena_custodia" {
				format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
				if format == "" {
					format = "json"
				}
				if format != "json" && format != "csv" {
					http.Error(w, "format invalido (use json o csv)", http.StatusBadRequest)
					return
				}

				payload, err := buildAuditoriaForenseExportPayload(empresaID, filter, total, rows)
				if err != nil {
					http.Error(w, "No se pudo construir la exportacion forense", http.StatusInternalServerError)
					return
				}

				if format == "csv" {
					if err := writeAuditoriaForenseCSV(w, payload); err != nil {
						http.Error(w, "No se pudo generar la exportacion forense CSV", http.StatusInternalServerError)
						return
					}
					return
				}

				writeJSON(w, http.StatusOK, payload)
				return
			}
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

func buildAuditoriaForenseExportPayload(empresaID int64, filter dbpkg.EmpresaAuditoriaEventoFilter, total int64, rows []dbpkg.EmpresaAuditoriaEvento) (auditoriaForenseExportPayload, error) {
	registros := make([]auditoriaForenseRegistro, 0, len(rows))
	chainPrev := "GENESIS"
	chainStream := strings.Builder{}

	for idx, row := range rows {
		base := auditoriaForenseBase{
			ID:             row.ID,
			FechaEvento:    strings.TrimSpace(row.FechaEvento),
			Modulo:         strings.TrimSpace(row.Modulo),
			Accion:         strings.TrimSpace(row.Accion),
			Recurso:        strings.TrimSpace(row.Recurso),
			RecursoID:      row.RecursoID,
			MetodoHTTP:     strings.TrimSpace(row.MetodoHTTP),
			Endpoint:       strings.TrimSpace(row.Endpoint),
			Resultado:      strings.TrimSpace(row.Resultado),
			CodigoHTTP:     row.CodigoHTTP,
			RequestID:      strings.TrimSpace(row.RequestID),
			UsuarioCreador: strings.TrimSpace(row.UsuarioCreador),
			IPOrigen:       strings.TrimSpace(row.IPOrigen),
			Observaciones:  strings.TrimSpace(row.Observaciones),
			MetadataJSON:   strings.TrimSpace(row.MetadataJSON),
		}
		baseJSON, err := json.Marshal(base)
		if err != nil {
			return auditoriaForenseExportPayload{}, err
		}
		hashRegistro := sha256Hex(baseJSON)
		hashCadena := sha256Hex([]byte(chainPrev + "|" + hashRegistro))

		registros = append(registros, auditoriaForenseRegistro{
			Indice:             idx + 1,
			Base:               base,
			HashRegistro:       hashRegistro,
			HashCadenaAnterior: chainPrev,
			HashCadena:         hashCadena,
		})

		chainPrev = hashCadena
		chainStream.WriteString(hashCadena)
		chainStream.WriteString("\n")
	}

	filtros := map[string]interface{}{
		"modulo":           strings.TrimSpace(filter.Modulo),
		"accion":           strings.TrimSpace(filter.Accion),
		"metodo_http":      strings.TrimSpace(filter.MetodoHTTP),
		"recurso":          strings.TrimSpace(filter.Recurso),
		"endpoint":         strings.TrimSpace(filter.Endpoint),
		"search":           strings.TrimSpace(filter.Search),
		"recurso_id":       filter.RecursoID,
		"codigo_http":      filter.CodigoHTTP,
		"resultado":        strings.TrimSpace(filter.Resultado),
		"usuario":          strings.TrimSpace(filter.UsuarioCreador),
		"request_id":       strings.TrimSpace(filter.RequestID),
		"desde":            strings.TrimSpace(filter.Desde),
		"hasta":            strings.TrimSpace(filter.Hasta),
		"include_inactive": filter.IncludeInactive,
		"limit":            filter.Limit,
		"offset":           filter.Offset,
	}

	manifest := auditoriaForenseManifest{
		EmpresaID:          empresaID,
		GeneradoEn:         time.Now().Format("2006-01-02 15:04:05"),
		AlgoritmoHash:      "sha256",
		TotalCoincidencias: total,
		TotalRegistros:     len(registros),
		HashGlobal:         sha256Hex([]byte(chainStream.String())),
		HashCadenaFinal:    chainPrev,
		Filtros:            filtros,
	}

	return auditoriaForenseExportPayload{
		OK:        true,
		Manifest:  manifest,
		Registros: registros,
	}, nil
}

func writeAuditoriaForenseCSV(w http.ResponseWriter, payload auditoriaForenseExportPayload) error {
	filename := fmt.Sprintf("auditoria_forense_empresa_%d_%s.csv", payload.Manifest.EmpresaID, time.Now().Format("20060102_150405"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)

	writer := csv.NewWriter(w)
	if err := writer.Write([]string{"manifest_empresa_id", strconv.FormatInt(payload.Manifest.EmpresaID, 10)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"manifest_generado_en", payload.Manifest.GeneradoEn}); err != nil {
		return err
	}
	if err := writer.Write([]string{"manifest_algoritmo_hash", payload.Manifest.AlgoritmoHash}); err != nil {
		return err
	}
	if err := writer.Write([]string{"manifest_total_coincidencias", strconv.FormatInt(payload.Manifest.TotalCoincidencias, 10)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"manifest_total_registros", strconv.Itoa(payload.Manifest.TotalRegistros)}); err != nil {
		return err
	}
	if err := writer.Write([]string{"manifest_hash_global", payload.Manifest.HashGlobal}); err != nil {
		return err
	}
	if err := writer.Write([]string{"manifest_hash_cadena_final", payload.Manifest.HashCadenaFinal}); err != nil {
		return err
	}
	if err := writer.Write([]string{}); err != nil {
		return err
	}
	if err := writer.Write([]string{
		"indice",
		"id",
		"fecha_evento",
		"modulo",
		"accion",
		"recurso",
		"recurso_id",
		"metodo_http",
		"endpoint",
		"resultado",
		"codigo_http",
		"request_id",
		"usuario_creador",
		"ip_origen",
		"observaciones",
		"metadata_json",
		"hash_registro",
		"hash_cadena_anterior",
		"hash_cadena",
	}); err != nil {
		return err
	}

	for _, row := range payload.Registros {
		if err := writer.Write([]string{
			strconv.Itoa(row.Indice),
			strconv.FormatInt(row.Base.ID, 10),
			row.Base.FechaEvento,
			row.Base.Modulo,
			row.Base.Accion,
			row.Base.Recurso,
			strconv.FormatInt(row.Base.RecursoID, 10),
			row.Base.MetodoHTTP,
			row.Base.Endpoint,
			row.Base.Resultado,
			strconv.FormatInt(row.Base.CodigoHTTP, 10),
			row.Base.RequestID,
			row.Base.UsuarioCreador,
			row.Base.IPOrigen,
			row.Base.Observaciones,
			row.Base.MetadataJSON,
			row.HashRegistro,
			row.HashCadenaAnterior,
			row.HashCadena,
		}); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}

func sha256Hex(input []byte) string {
	sum := sha256.Sum256(input)
	return fmt.Sprintf("%x", sum[:])
}

func registrarAuditoriaOperacionNoBloqueante(dbEmp *sql.DB, r *http.Request, empresaID int64, modulo, permissionAction string, statusCode int, elapsed time.Duration) {
	startedAt := time.Now()
	defer func() {
		dbpkg.PerfLogf("[perf][auditoria] empresa=%d modulo=%s accion=%s dur=%s", empresaID, modulo, permissionAction, time.Since(startedAt))
	}()
	if dbEmp == nil {
		return
	}
	if !accionDebeAuditarse(permissionAction) {
		return
	}
	if empresaID <= 0 {
		return
	}

	metadata := map[string]interface{}{
		"permission_action": strings.ToUpper(strings.TrimSpace(permissionAction)),
		"duracion_ms":       elapsed.Milliseconds(),
		"content_length":    r.ContentLength,
	}
	if role := strings.TrimSpace(r.Header.Get("X-Admin-Role")); role != "" {
		metadata["admin_role"] = role
	}
	if roleEfectivo := strings.TrimSpace(r.Header.Get("X-Admin-Role-Efectivo")); roleEfectivo != "" {
		metadata["admin_role_efectivo"] = roleEfectivo
	}
	if len(r.URL.Query()) > 0 {
		keys := make([]string, 0, len(r.URL.Query()))
		for key := range r.URL.Query() {
			key = strings.TrimSpace(key)
			if key != "" {
				keys = append(keys, key)
			}
		}
		metadata["query_keys"] = keys
	}
	if src := strings.TrimSpace(r.Header.Get("X-PCS-Source")); src != "" {
		metadata["source"] = strings.ToLower(src)
	}
	if cid := strings.TrimSpace(r.Header.Get("X-PCS-Chat-Conversation-ID")); cid != "" {
		metadata["chat_conversation_id"] = cid
	}
	if strings.TrimSpace(r.Header.Get(permissionApprovalHeaderRequired)) == "1" {
		metadata["permission_approval_required"] = true
	}
	if approvedBy := strings.TrimSpace(r.Header.Get(permissionApprovalHeaderBy)); approvedBy != "" {
		metadata["permission_approved_by"] = approvedBy
	}
	if approvalCode := strings.TrimSpace(r.Header.Get(permissionApprovalHeaderCode)); approvalCode != "" {
		metadata["permission_approval_code"] = approvalCode
	}
	if approvalReason := strings.TrimSpace(r.Header.Get(permissionApprovalHeaderReason)); approvalReason != "" {
		metadata["permission_approval_reason"] = approvalReason
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
		Observaciones:  "auditoria automatica de operacion empresarial",
	}

	go func(audit dbpkg.EmpresaAuditoriaEvento, eid int64, mod string) {
		if _, err := dbpkg.CreateEmpresaAuditoriaEvento(dbEmp, audit); err != nil {
			log.Printf("[auditoria] no se pudo registrar evento empresa_id=%d modulo=%s accion=%s error=%v", eid, mod, audit.Accion, err)
		}
	}(auditoria, empresaID, modulo)
}

func accionDebeAuditarse(permissionAction string) bool {
	switch strings.ToUpper(strings.TrimSpace(permissionAction)) {
	case "R", "C", "U", "D", "A":
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
	case "R":
		return "leer"
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
