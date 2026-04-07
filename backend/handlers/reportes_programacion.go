package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type reportesHTTPError struct {
	status  int
	message string
}

func (e reportesHTTPError) Error() string {
	return e.message
}

func newReportesHTTPError(status int, message string) error {
	return reportesHTTPError{status: status, message: strings.TrimSpace(message)}
}

func writeReportesHTTPError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	if httpErr, ok := err.(reportesHTTPError); ok {
		http.Error(w, httpErr.message, httpErr.status)
		return
	}
	http.Error(w, "Error interno de reportes: "+strings.TrimSpace(err.Error()), http.StatusInternalServerError)
}

func reportesDatasetExists(key string) (empresaReporteCatalogoItem, bool) {
	normalized := strings.ToLower(strings.TrimSpace(key))
	for _, item := range reportesCatalogo {
		if strings.ToLower(strings.TrimSpace(item.Key)) == normalized {
			return item, true
		}
	}
	return empresaReporteCatalogoItem{}, false
}

func reportesNormalizeFormat(format string) string {
	v := strings.ToLower(strings.TrimSpace(format))
	switch v {
	case "excel", "tsv":
		return "xls"
	case "json", "csv", "txt", "xls", "pdf":
		return v
	default:
		return ""
	}
}

func reportesNormalizeFormats(formats []string, fallback []string) []string {
	uniq := make(map[string]struct{})
	result := make([]string, 0, len(formats))
	for _, raw := range formats {
		for _, part := range reportesSplitCommaList(raw) {
			norm := reportesNormalizeFormat(part)
			if norm == "" {
				continue
			}
			if _, exists := uniq[norm]; exists {
				continue
			}
			uniq[norm] = struct{}{}
			result = append(result, norm)
		}
	}
	if len(result) > 0 {
		return result
	}
	for _, raw := range fallback {
		norm := reportesNormalizeFormat(raw)
		if norm == "" {
			continue
		}
		if _, exists := uniq[norm]; exists {
			continue
		}
		uniq[norm] = struct{}{}
		result = append(result, norm)
	}
	if len(result) == 0 {
		result = append(result, "json")
	}
	return result
}

func reportesSplitCommaList(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}

func reportesDecodeJSONArray(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	var arr []string
	if err := json.Unmarshal([]byte(trimmed), &arr); err == nil {
		return arr
	}
	return reportesSplitCommaList(trimmed)
}

func reportesDecodeJSONMap(raw string) map[string]interface{} {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return map[string]interface{}{}
	}
	var out map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &out); err != nil || out == nil {
		return map[string]interface{}{}
	}
	return out
}

func reportesMarshalJSON(v interface{}, fallback string) string {
	raw, err := json.Marshal(v)
	if err != nil {
		return fallback
	}
	return string(raw)
}

func reportesHashBytes(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func reportesHashString(raw string) string {
	return reportesHashBytes([]byte(raw))
}

func reportesBoolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func reportesToBool(v interface{}, fallback bool) bool {
	switch t := v.(type) {
	case bool:
		return t
	case float64:
		return t > 0
	case int:
		return t > 0
	case int64:
		return t > 0
	case string:
		value := strings.ToLower(strings.TrimSpace(t))
		if value == "1" || value == "true" || value == "si" || value == "yes" {
			return true
		}
		if value == "0" || value == "false" || value == "no" {
			return false
		}
	}
	return fallback
}

func reportesToInt64(v interface{}, fallback int64) int64 {
	switch t := v.(type) {
	case int:
		return int64(t)
	case int32:
		return int64(t)
	case int64:
		return t
	case float32:
		return int64(t)
	case float64:
		return int64(t)
	case json.Number:
		i, err := t.Int64()
		if err == nil {
			return i
		}
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64)
		if err == nil {
			return parsed
		}
	}
	return fallback
}

func reportesCurrentTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func reportesNullString(v sql.NullString) string {
	if v.Valid {
		return v.String
	}
	return ""
}

func reportesNullInt64(v sql.NullInt64) int64 {
	if v.Valid {
		return v.Int64
	}
	return 0
}

func reportesNormalizeHoraEnvio(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "08:00", nil
	}
	parsed, err := time.Parse("15:04", trimmed)
	if err != nil {
		return "", err
	}
	return parsed.Format("15:04"), nil
}

func reportesNormalizeFrecuencia(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "diario", "daily", "dia":
		return "diario"
	case "semanal", "weekly", "semana":
		return "semanal"
	case "mensual", "monthly", "mes":
		return "mensual"
	case "manual":
		return "manual"
	default:
		return "diario"
	}
}

func reportesComputeNextExecution(now time.Time, frecuencia, hora string) string {
	if strings.TrimSpace(frecuencia) == "manual" {
		return ""
	}
	hhmm, err := reportesNormalizeHoraEnvio(hora)
	if err != nil {
		hhmm = "08:00"
	}
	clock, _ := time.Parse("15:04", hhmm)
	next := time.Date(now.Year(), now.Month(), now.Day(), clock.Hour(), clock.Minute(), 0, 0, now.Location())

	switch reportesNormalizeFrecuencia(frecuencia) {
	case "semanal":
		if !next.After(now) {
			next = next.AddDate(0, 0, 7)
		}
	case "mensual":
		if !next.After(now) {
			next = next.AddDate(0, 1, 0)
		}
	default:
		if !next.After(now) {
			next = next.AddDate(0, 0, 1)
		}
	}

	return next.Format("2006-01-02 15:04:05")
}

func reportesDecodeBodyJSON(r *http.Request, dst interface{}) error {
	if r == nil || r.Body == nil {
		return io.EOF
	}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}

type reportePlantillaPayload struct {
	ID             int64                  `json:"id"`
	EmpresaID      int64                  `json:"empresa_id"`
	Codigo         string                 `json:"codigo"`
	Nombre         string                 `json:"nombre"`
	DatasetKey     string                 `json:"dataset_key"`
	Formato        string                 `json:"formato"`
	Version        int64                  `json:"version"`
	Columnas       []string               `json:"columnas"`
	Config         map[string]interface{} `json:"config"`
	UsuarioCreador string                 `json:"usuario_creador"`
	Observaciones  string                 `json:"observaciones"`
	MarcarVigente  *bool                  `json:"marcar_vigente"`
}

type reporteProgramacionPayload struct {
	ID                  int64                  `json:"id"`
	EmpresaID           int64                  `json:"empresa_id"`
	Nombre              string                 `json:"nombre"`
	Dataset             string                 `json:"dataset"`
	DatasetKey          string                 `json:"dataset_key"`
	Frecuencia          string                 `json:"frecuencia"`
	HoraEnvio           string                 `json:"hora_envio"`
	Timezone            string                 `json:"timezone"`
	Formatos            []string               `json:"formatos"`
	Destinatarios       []string               `json:"destinatarios"`
	DestinatariosRaw    string                 `json:"destinatarios_raw"`
	TemplateCodigo      string                 `json:"template_codigo"`
	TemplateVersion     int64                  `json:"template_version"`
	Parametros          map[string]interface{} `json:"parametros"`
	ValidarConsistencia *bool                  `json:"validar_consistencia"`
	Activa              *bool                  `json:"activa"`
	UsuarioCreador      string                 `json:"usuario_creador"`
	Observaciones       string                 `json:"observaciones"`
}

type reporteExecuteProgramacionPayload struct {
	ProgramacionID int64    `json:"programacion_id"`
	Formatos       []string `json:"formatos"`
	UsuarioCreador string   `json:"usuario_creador"`
}

type reporteConsistenciaPayload struct {
	Dataset         string   `json:"dataset"`
	DatasetKey      string   `json:"dataset_key"`
	Formatos        []string `json:"formatos"`
	TemplateCodigo  string   `json:"template_codigo"`
	TemplateVersion int64    `json:"template_version"`
}

func handleEmpresaReportesPlantillasAction(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID int64) error {
	switch r.Method {
	case http.MethodGet:
		query := `SELECT
			id, empresa_id, codigo, nombre, dataset_key, version, formato,
			columnas_json, config_json, vigente, hash_contenido,
			fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
		FROM empresa_reportes_plantillas
		WHERE empresa_id = ?`
		args := []interface{}{empresaID}

		codigo := strings.TrimSpace(r.URL.Query().Get("codigo"))
		if codigo != "" {
			query += " AND lower(codigo) = ?"
			args = append(args, strings.ToLower(codigo))
		}

		dataset := strings.TrimSpace(r.URL.Query().Get("dataset"))
		if dataset != "" {
			query += " AND lower(dataset_key) = ?"
			args = append(args, strings.ToLower(dataset))
		}

		if queryBool(r, "solo_vigente") || queryBool(r, "vigente") {
			query += " AND vigente = 1"
		}
		query += " ORDER BY codigo ASC, version DESC, id DESC LIMIT 500"

		rows, err := dbEmp.Query(query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		items := make([]map[string]interface{}, 0)
		for rows.Next() {
			var (
				id, empID, version                    int64
				codigo, nombre, datasetKey, formato   string
				columnasJSON, configJSON              string
				vigente                               int
				hashContenido                         sql.NullString
				fechaCreacion, fechaActualizacion     string
				usuarioCreador, estado, observaciones sql.NullString
			)
			if err := rows.Scan(
				&id, &empID, &codigo, &nombre, &datasetKey, &version, &formato,
				&columnasJSON, &configJSON, &vigente, &hashContenido,
				&fechaCreacion, &fechaActualizacion, &usuarioCreador, &estado, &observaciones,
			); err != nil {
				return err
			}

			item := map[string]interface{}{
				"id":                  id,
				"empresa_id":          empID,
				"codigo":              codigo,
				"nombre":              nombre,
				"dataset_key":         datasetKey,
				"version":             version,
				"formato":             formato,
				"columnas":            reportesDecodeJSONArray(columnasJSON),
				"config":              reportesDecodeJSONMap(configJSON),
				"vigente":             vigente > 0,
				"hash_contenido":      reportesNullString(hashContenido),
				"fecha_creacion":      fechaCreacion,
				"fecha_actualizacion": fechaActualizacion,
				"usuario_creador":     reportesNullString(usuarioCreador),
				"estado":              reportesFirstNonBlank(reportesNullString(estado), "activo"),
				"observaciones":       reportesNullString(observaciones),
			}
			items = append(items, item)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"empresa_id": empresaID,
			"total":      len(items),
			"plantillas": items,
		})
		return nil

	case http.MethodPost, http.MethodPut:
		var payload reportePlantillaPayload
		if err := reportesDecodeBodyJSON(r, &payload); err != nil && err != io.EOF {
			return newReportesHTTPError(http.StatusBadRequest, "payload JSON invalido")
		}
		if payload.EmpresaID > 0 && payload.EmpresaID != empresaID {
			return newReportesHTTPError(http.StatusBadRequest, "empresa_id no coincide con el contexto")
		}

		codigo := strings.TrimSpace(payload.Codigo)
		if codigo == "" {
			return newReportesHTTPError(http.StatusBadRequest, "codigo es obligatorio")
		}
		datasetKey := strings.ToLower(strings.TrimSpace(reportesFirstNonBlank(payload.DatasetKey, r.URL.Query().Get("dataset"))))
		if datasetKey == "" {
			return newReportesHTTPError(http.StatusBadRequest, "dataset_key es obligatorio")
		}
		if _, ok := reportesDatasetExists(datasetKey); !ok {
			return newReportesHTTPError(http.StatusBadRequest, "dataset_key no soportado")
		}

		formato := reportesNormalizeFormat(reportesFirstNonBlank(payload.Formato, "json"))
		if formato == "" {
			return newReportesHTTPError(http.StatusBadRequest, "formato de plantilla no soportado")
		}

		nombre := strings.TrimSpace(payload.Nombre)
		if nombre == "" {
			nombre = "Plantilla " + strings.ToUpper(codigo)
		}

		columnas := make([]string, 0, len(payload.Columnas))
		seen := make(map[string]struct{})
		for _, col := range payload.Columnas {
			trimmed := strings.TrimSpace(col)
			if trimmed == "" {
				continue
			}
			key := strings.ToLower(trimmed)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			columnas = append(columnas, trimmed)
		}

		config := payload.Config
		if config == nil {
			config = map[string]interface{}{}
		}
		columnasJSON := reportesMarshalJSON(columnas, "[]")
		configJSON := reportesMarshalJSON(config, "{}")
		hashContenido := reportesHashString(strings.Join([]string{
			codigo,
			datasetKey,
			formato,
			columnasJSON,
			configJSON,
		}, "|"))
		usuario := strings.TrimSpace(payload.UsuarioCreador)
		if usuario == "" {
			usuario = "sistema_reportes"
		}
		vigente := true
		if payload.MarcarVigente != nil {
			vigente = *payload.MarcarVigente
		}
		ahora := reportesCurrentTimestamp()

		targetID := payload.ID
		if targetID <= 0 && r.Method == http.MethodPut && payload.Version > 0 {
			err := dbEmp.QueryRow(`
				SELECT id
				FROM empresa_reportes_plantillas
				WHERE empresa_id = ? AND lower(codigo) = lower(?) AND version = ?
				ORDER BY id DESC LIMIT 1
			`, empresaID, codigo, payload.Version).Scan(&targetID)
			if err != nil && err != sql.ErrNoRows {
				return err
			}
		}

		if targetID > 0 {
			if vigente {
				if _, err := dbEmp.Exec(`
					UPDATE empresa_reportes_plantillas
					SET vigente = 0, fecha_actualizacion = ?
					WHERE empresa_id = ? AND lower(codigo) = lower(?) AND id <> ?
				`, ahora, empresaID, codigo, targetID); err != nil {
					return err
				}
			}

			if _, err := dbEmp.Exec(`
				UPDATE empresa_reportes_plantillas
				SET
					nombre = ?,
					dataset_key = ?,
					formato = ?,
					columnas_json = ?,
					config_json = ?,
					vigente = ?,
					hash_contenido = ?,
					fecha_actualizacion = ?,
					usuario_creador = ?,
					observaciones = ?
				WHERE empresa_id = ? AND id = ?
			`, nombre, datasetKey, formato, columnasJSON, configJSON, reportesBoolToInt(vigente), hashContenido, ahora, usuario, strings.TrimSpace(payload.Observaciones), empresaID, targetID); err != nil {
				return err
			}

			item, err := getReportePlantillaByID(dbEmp, empresaID, targetID)
			if err != nil {
				return err
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":         true,
				"empresa_id": empresaID,
				"plantilla":  item,
			})
			return nil
		}

		var maxVersion int64
		if err := dbEmp.QueryRow(`
			SELECT COALESCE(MAX(version), 0)
			FROM empresa_reportes_plantillas
			WHERE empresa_id = ? AND lower(codigo) = lower(?)
		`, empresaID, codigo).Scan(&maxVersion); err != nil {
			return err
		}
		version := maxVersion + 1
		if payload.Version > version {
			version = payload.Version
		}

		if vigente {
			if _, err := dbEmp.Exec(`
				UPDATE empresa_reportes_plantillas
				SET vigente = 0, fecha_actualizacion = ?
				WHERE empresa_id = ? AND lower(codigo) = lower(?)
			`, ahora, empresaID, codigo); err != nil {
				return err
			}
		}

		res, err := dbEmp.Exec(`
			INSERT INTO empresa_reportes_plantillas (
				empresa_id, codigo, nombre, dataset_key, version, formato,
				columnas_json, config_json, vigente, hash_contenido,
				fecha_actualizacion, usuario_creador, estado, observaciones
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'activo', ?)
		`, empresaID, codigo, nombre, datasetKey, version, formato, columnasJSON, configJSON, reportesBoolToInt(vigente), hashContenido, ahora, usuario, strings.TrimSpace(payload.Observaciones))
		if err != nil {
			return err
		}
		insertID, err := res.LastInsertId()
		if err != nil {
			return err
		}
		item, err := getReportePlantillaByID(dbEmp, empresaID, insertID)
		if err != nil {
			return err
		}
		writeJSON(w, http.StatusCreated, map[string]interface{}{
			"ok":         true,
			"empresa_id": empresaID,
			"plantilla":  item,
		})
		return nil

	default:
		return newReportesHTTPError(http.StatusMethodNotAllowed, "method not allowed")
	}
}

func getReportePlantillaByID(dbEmp *sql.DB, empresaID, id int64) (map[string]interface{}, error) {
	var (
		rowID, empID, version                 int64
		codigo, nombre, datasetKey, formato   string
		columnasJSON, configJSON              string
		vigente                               int
		hashContenido                         sql.NullString
		fechaCreacion, fechaActualizacion     string
		usuarioCreador, estado, observaciones sql.NullString
	)
	err := dbEmp.QueryRow(`
		SELECT
			id, empresa_id, codigo, nombre, dataset_key, version, formato,
			columnas_json, config_json, vigente, hash_contenido,
			fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
		FROM empresa_reportes_plantillas
		WHERE empresa_id = ? AND id = ?
	`, empresaID, id).Scan(
		&rowID, &empID, &codigo, &nombre, &datasetKey, &version, &formato,
		&columnasJSON, &configJSON, &vigente, &hashContenido,
		&fechaCreacion, &fechaActualizacion, &usuarioCreador, &estado, &observaciones,
	)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id":                  rowID,
		"empresa_id":          empID,
		"codigo":              codigo,
		"nombre":              nombre,
		"dataset_key":         datasetKey,
		"version":             version,
		"formato":             formato,
		"columnas":            reportesDecodeJSONArray(columnasJSON),
		"config":              reportesDecodeJSONMap(configJSON),
		"vigente":             vigente > 0,
		"hash_contenido":      reportesNullString(hashContenido),
		"fecha_creacion":      fechaCreacion,
		"fecha_actualizacion": fechaActualizacion,
		"usuario_creador":     reportesNullString(usuarioCreador),
		"estado":              reportesFirstNonBlank(reportesNullString(estado), "activo"),
		"observaciones":       reportesNullString(observaciones),
	}, nil
}

func resolveReportePlantilla(dbEmp *sql.DB, empresaID int64, codigo string, version int64) (map[string]interface{}, error) {
	codigo = strings.TrimSpace(codigo)
	if codigo == "" {
		return nil, sql.ErrNoRows
	}

	query := `
		SELECT
			id, empresa_id, codigo, nombre, dataset_key, version, formato,
			columnas_json, config_json, vigente, hash_contenido,
			fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
		FROM empresa_reportes_plantillas
		WHERE empresa_id = ? AND lower(codigo) = lower(?)`
	args := []interface{}{empresaID, codigo}
	if version > 0 {
		query += " AND version = ?"
		args = append(args, version)
	} else {
		query += " AND vigente = 1"
	}
	query += " ORDER BY version DESC, id DESC LIMIT 1"

	var (
		id, empID, rowVersion                  int64
		rowCodigo, nombre, datasetKey, formato string
		columnasJSON, configJSON               string
		vigente                                int
		hashContenido                          sql.NullString
		fechaCreacion, fechaActualizacion      string
		usuarioCreador, estado, observaciones  sql.NullString
	)
	err := dbEmp.QueryRow(query, args...).Scan(
		&id, &empID, &rowCodigo, &nombre, &datasetKey, &rowVersion, &formato,
		&columnasJSON, &configJSON, &vigente, &hashContenido,
		&fechaCreacion, &fechaActualizacion, &usuarioCreador, &estado, &observaciones,
	)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":                  id,
		"empresa_id":          empID,
		"codigo":              rowCodigo,
		"nombre":              nombre,
		"dataset_key":         datasetKey,
		"version":             rowVersion,
		"formato":             formato,
		"columnas":            reportesDecodeJSONArray(columnasJSON),
		"config":              reportesDecodeJSONMap(configJSON),
		"vigente":             vigente > 0,
		"hash_contenido":      reportesNullString(hashContenido),
		"fecha_creacion":      fechaCreacion,
		"fecha_actualizacion": fechaActualizacion,
		"usuario_creador":     reportesNullString(usuarioCreador),
		"estado":              reportesFirstNonBlank(reportesNullString(estado), "activo"),
		"observaciones":       reportesNullString(observaciones),
	}, nil
}

func reportesApplyTemplate(ds empresaReporteDataset, plantilla map[string]interface{}) empresaReporteDataset {
	if plantilla == nil {
		return ds
	}
	columnasRaw, _ := plantilla["columnas"].([]string)
	if len(columnasRaw) == 0 {
		if values, ok := plantilla["columnas"].([]interface{}); ok {
			columnasRaw = make([]string, 0, len(values))
			for _, value := range values {
				columnasRaw = append(columnasRaw, strings.TrimSpace(reportesStringValue(value)))
			}
		}
	}
	columnas := make([]string, 0, len(columnasRaw))
	for _, col := range columnasRaw {
		if strings.TrimSpace(col) == "" {
			continue
		}
		columnas = append(columnas, col)
	}
	if len(columnas) > 0 {
		filteredRows := make([]map[string]interface{}, 0, len(ds.Rows))
		for _, row := range ds.Rows {
			filtered := make(map[string]interface{}, len(columnas))
			for _, col := range columnas {
				filtered[col] = row[col]
			}
			filteredRows = append(filteredRows, filtered)
		}
		ds.Columns = columnas
		ds.Rows = filteredRows
		ds.RowCount = len(filteredRows)
	}

	if ds.Summary == nil {
		ds.Summary = make(map[string]interface{})
	}
	ds.Summary["template_codigo"] = plantilla["codigo"]
	ds.Summary["template_version"] = plantilla["version"]
	return ds
}

func reportesApplyTemplateFromRequest(dbEmp *sql.DB, empresaID int64, r *http.Request, ds empresaReporteDataset) (empresaReporteDataset, error) {
	templateCodigo := strings.TrimSpace(r.URL.Query().Get("template_codigo"))
	if templateCodigo == "" {
		return ds, nil
	}
	templateVersion, err := parseInt64QueryOptional(r, "template_version")
	if err != nil {
		return ds, newReportesHTTPError(http.StatusBadRequest, "template_version invalido")
	}
	plantilla, err := resolveReportePlantilla(dbEmp, empresaID, templateCodigo, templateVersion)
	if err != nil {
		if err == sql.ErrNoRows {
			return ds, newReportesHTTPError(http.StatusBadRequest, "plantilla de exportacion no encontrada")
		}
		return ds, err
	}
	return reportesApplyTemplate(ds, plantilla), nil
}

func handleEmpresaReportesProgramacionAction(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID int64) error {
	switch r.Method {
	case http.MethodGet:
		query := `SELECT
			id, empresa_id, nombre, dataset_key, nivel, formatos, parametros_json,
			template_codigo, template_version, frecuencia, hora_envio, timezone,
			destinatarios, ultimo_ejecutado_en, proximo_ejecutado_en,
			activa, validacion_consistencia, hash_ultima_ejecucion,
			fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
		FROM empresa_reportes_programaciones
		WHERE empresa_id = ?`
		args := []interface{}{empresaID}

		dataset := strings.TrimSpace(r.URL.Query().Get("dataset"))
		if dataset != "" {
			query += " AND lower(dataset_key) = ?"
			args = append(args, strings.ToLower(dataset))
		}

		if r.URL.Query().Has("activa") {
			if queryBool(r, "activa") {
				query += " AND activa = 1"
			} else {
				query += " AND activa = 0"
			}
		}

		limit, err := parseIntQueryOptional(r, "limit")
		if err != nil {
			return newReportesHTTPError(http.StatusBadRequest, "limit invalido")
		}
		if limit <= 0 || limit > 500 {
			limit = 200
		}
		query += " ORDER BY id DESC LIMIT ?"
		args = append(args, limit)

		rows, err := dbEmp.Query(query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		items := make([]map[string]interface{}, 0)
		for rows.Next() {
			var (
				id, empID, templateVersion                       int64
				nombre, datasetKey, nivel                        string
				formatosJSON, parametrosJSON                     string
				templateCodigo, frecuencia, horaEnvio, timezone  sql.NullString
				destinatarios, ultimoEjecutado, proximoEjecutado sql.NullString
				activa, validarConsistencia                      int
				hashUltimaEjecucion                              sql.NullString
				fechaCreacion, fechaActualizacion                string
				usuarioCreador, estado, observaciones            sql.NullString
			)
			if err := rows.Scan(
				&id, &empID, &nombre, &datasetKey, &nivel, &formatosJSON, &parametrosJSON,
				&templateCodigo, &templateVersion, &frecuencia, &horaEnvio, &timezone,
				&destinatarios, &ultimoEjecutado, &proximoEjecutado,
				&activa, &validarConsistencia, &hashUltimaEjecucion,
				&fechaCreacion, &fechaActualizacion, &usuarioCreador, &estado, &observaciones,
			); err != nil {
				return err
			}

			item := map[string]interface{}{
				"id":                    id,
				"empresa_id":            empID,
				"nombre":                nombre,
				"dataset_key":           datasetKey,
				"nivel":                 nivel,
				"formatos":              reportesDecodeJSONArray(formatosJSON),
				"parametros":            reportesDecodeJSONMap(parametrosJSON),
				"template_codigo":       reportesNullString(templateCodigo),
				"template_version":      templateVersion,
				"frecuencia":            reportesFirstNonBlank(reportesNullString(frecuencia), "diario"),
				"hora_envio":            reportesFirstNonBlank(reportesNullString(horaEnvio), "08:00"),
				"timezone":              reportesFirstNonBlank(reportesNullString(timezone), "America/Bogota"),
				"destinatarios":         reportesSplitCommaList(reportesNullString(destinatarios)),
				"ultimo_ejecutado_en":   reportesNullString(ultimoEjecutado),
				"proximo_ejecutado_en":  reportesNullString(proximoEjecutado),
				"activa":                activa > 0,
				"validar_consistencia":  validarConsistencia > 0,
				"hash_ultima_ejecucion": reportesNullString(hashUltimaEjecucion),
				"fecha_creacion":        fechaCreacion,
				"fecha_actualizacion":   fechaActualizacion,
				"usuario_creador":       reportesNullString(usuarioCreador),
				"estado":                reportesFirstNonBlank(reportesNullString(estado), "activo"),
				"observaciones":         reportesNullString(observaciones),
			}
			items = append(items, item)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"empresa_id":     empresaID,
			"total":          len(items),
			"programaciones": items,
		})
		return nil

	case http.MethodPost, http.MethodPut:
		var payload reporteProgramacionPayload
		if err := reportesDecodeBodyJSON(r, &payload); err != nil && err != io.EOF {
			return newReportesHTTPError(http.StatusBadRequest, "payload JSON invalido")
		}
		if payload.EmpresaID > 0 && payload.EmpresaID != empresaID {
			return newReportesHTTPError(http.StatusBadRequest, "empresa_id no coincide con el contexto")
		}

		nombre := strings.TrimSpace(payload.Nombre)
		if nombre == "" {
			nombre = "Programacion de reportes"
		}
		datasetKey := strings.ToLower(strings.TrimSpace(reportesFirstNonBlank(payload.DatasetKey, payload.Dataset, r.URL.Query().Get("dataset"))))
		if datasetKey == "" {
			return newReportesHTTPError(http.StatusBadRequest, "dataset es obligatorio")
		}
		meta, ok := reportesDatasetExists(datasetKey)
		if !ok {
			return newReportesHTTPError(http.StatusBadRequest, "dataset no soportado")
		}

		horaEnvio, err := reportesNormalizeHoraEnvio(payload.HoraEnvio)
		if err != nil {
			return newReportesHTTPError(http.StatusBadRequest, "hora_envio invalida (use HH:MM)")
		}

		frecuencia := reportesNormalizeFrecuencia(payload.Frecuencia)
		timezone := strings.TrimSpace(payload.Timezone)
		if timezone == "" {
			timezone = "America/Bogota"
		}

		formatos := reportesNormalizeFormats(payload.Formatos, []string{"json", "csv", "xls"})
		destinatarios := make([]string, 0, len(payload.Destinatarios)+4)
		destinatarios = append(destinatarios, payload.Destinatarios...)
		destinatarios = append(destinatarios, reportesSplitCommaList(payload.DestinatariosRaw)...)
		cleanDestinatarios := make([]string, 0, len(destinatarios))
		seenDestinatarios := make(map[string]struct{})
		for _, d := range destinatarios {
			trimmed := strings.TrimSpace(d)
			if trimmed == "" {
				continue
			}
			key := strings.ToLower(trimmed)
			if _, exists := seenDestinatarios[key]; exists {
				continue
			}
			seenDestinatarios[key] = struct{}{}
			cleanDestinatarios = append(cleanDestinatarios, trimmed)
		}

		templateCodigo := strings.TrimSpace(payload.TemplateCodigo)
		templateVersion := payload.TemplateVersion
		if templateCodigo != "" {
			plantilla, err := resolveReportePlantilla(dbEmp, empresaID, templateCodigo, templateVersion)
			if err != nil {
				if err == sql.ErrNoRows {
					return newReportesHTTPError(http.StatusBadRequest, "plantilla de exportacion no encontrada")
				}
				return err
			}
			resolvedVersion := reportesToInt64(plantilla["version"], 0)
			templateVersion = resolvedVersion
		}

		validarConsistencia := true
		if payload.ValidarConsistencia != nil {
			validarConsistencia = *payload.ValidarConsistencia
		}
		activa := true
		if payload.Activa != nil {
			activa = *payload.Activa
		}

		parametros := payload.Parametros
		if parametros == nil {
			parametros = map[string]interface{}{}
		}
		parametrosJSON := reportesMarshalJSON(parametros, "{}")
		formatosJSON := reportesMarshalJSON(formatos, "[\"json\"]")
		usuario := strings.TrimSpace(payload.UsuarioCreador)
		if usuario == "" {
			usuario = "sistema_reportes"
		}
		ahora := reportesCurrentTimestamp()
		proximo := reportesComputeNextExecution(time.Now(), frecuencia, horaEnvio)

		targetID := payload.ID
		if targetID <= 0 && r.URL.Query().Has("id") {
			parsedID, err := parseInt64QueryOptional(r, "id")
			if err == nil {
				targetID = parsedID
			}
		}

		if targetID > 0 {
			if _, err := dbEmp.Exec(`
				UPDATE empresa_reportes_programaciones
				SET
					nombre = ?,
					dataset_key = ?,
					nivel = ?,
					formatos = ?,
					parametros_json = ?,
					template_codigo = ?,
					template_version = ?,
					frecuencia = ?,
					hora_envio = ?,
					timezone = ?,
					destinatarios = ?,
					proximo_ejecutado_en = ?,
					activa = ?,
					validacion_consistencia = ?,
					fecha_actualizacion = ?,
					usuario_creador = ?,
					observaciones = ?
				WHERE empresa_id = ? AND id = ?
			`,
				nombre,
				datasetKey,
				meta.Level,
				formatosJSON,
				parametrosJSON,
				templateCodigo,
				templateVersion,
				frecuencia,
				horaEnvio,
				timezone,
				strings.Join(cleanDestinatarios, ","),
				proximo,
				reportesBoolToInt(activa),
				reportesBoolToInt(validarConsistencia),
				ahora,
				usuario,
				strings.TrimSpace(payload.Observaciones),
				empresaID,
				targetID,
			); err != nil {
				return err
			}
			item, err := getReporteProgramacionByID(dbEmp, empresaID, targetID)
			if err != nil {
				return err
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":           true,
				"empresa_id":   empresaID,
				"programacion": item,
			})
			return nil
		}

		res, err := dbEmp.Exec(`
			INSERT INTO empresa_reportes_programaciones (
				empresa_id, nombre, dataset_key, nivel, formatos, parametros_json,
				template_codigo, template_version, frecuencia, hora_envio, timezone,
				destinatarios, proximo_ejecutado_en, activa, validacion_consistencia,
				fecha_actualizacion, usuario_creador, estado, observaciones
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'activo', ?)
		`,
			empresaID,
			nombre,
			datasetKey,
			meta.Level,
			formatosJSON,
			parametrosJSON,
			templateCodigo,
			templateVersion,
			frecuencia,
			horaEnvio,
			timezone,
			strings.Join(cleanDestinatarios, ","),
			proximo,
			reportesBoolToInt(activa),
			reportesBoolToInt(validarConsistencia),
			ahora,
			usuario,
			strings.TrimSpace(payload.Observaciones),
		)
		if err != nil {
			return err
		}
		insertID, err := res.LastInsertId()
		if err != nil {
			return err
		}
		item, err := getReporteProgramacionByID(dbEmp, empresaID, insertID)
		if err != nil {
			return err
		}
		writeJSON(w, http.StatusCreated, map[string]interface{}{
			"ok":           true,
			"empresa_id":   empresaID,
			"programacion": item,
		})
		return nil

	default:
		return newReportesHTTPError(http.StatusMethodNotAllowed, "method not allowed")
	}
}

func getReporteProgramacionByID(dbEmp *sql.DB, empresaID, id int64) (map[string]interface{}, error) {
	var (
		rowID, empID, templateVersion                       int64
		nombre, datasetKey, nivel                           string
		formatosJSON, parametrosJSON                        string
		templateCodigo, frecuencia, horaEnvio, timezone     sql.NullString
		destinatarios, ultimoEjecutado, proximoEjecutado    sql.NullString
		activa, validarConsistencia                         int
		hashUltimaEjecucion                                 sql.NullString
		fechaCreacion, fechaActualizacion                   string
		usuarioCreador, estado, observaciones              sql.NullString
	)
	err := dbEmp.QueryRow(`
		SELECT
			id, empresa_id, nombre, dataset_key, nivel, formatos, parametros_json,
			template_codigo, template_version, frecuencia, hora_envio, timezone,
			destinatarios, ultimo_ejecutado_en, proximo_ejecutado_en,
			activa, validacion_consistencia, hash_ultima_ejecucion,
			fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
		FROM empresa_reportes_programaciones
		WHERE empresa_id = ? AND id = ?
	`, empresaID, id).Scan(
		&rowID, &empID, &nombre, &datasetKey, &nivel, &formatosJSON, &parametrosJSON,
		&templateCodigo, &templateVersion, &frecuencia, &horaEnvio, &timezone,
		&destinatarios, &ultimoEjecutado, &proximoEjecutado,
		&activa, &validarConsistencia, &hashUltimaEjecucion,
		&fechaCreacion, &fechaActualizacion, &usuarioCreador, &estado, &observaciones,
	)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id":                    rowID,
		"empresa_id":            empID,
		"nombre":                nombre,
		"dataset_key":           datasetKey,
		"nivel":                 nivel,
		"formatos":              reportesDecodeJSONArray(formatosJSON),
		"parametros":            reportesDecodeJSONMap(parametrosJSON),
		"template_codigo":       reportesNullString(templateCodigo),
		"template_version":      templateVersion,
		"frecuencia":            reportesFirstNonBlank(reportesNullString(frecuencia), "diario"),
		"hora_envio":            reportesFirstNonBlank(reportesNullString(horaEnvio), "08:00"),
		"timezone":              reportesFirstNonBlank(reportesNullString(timezone), "America/Bogota"),
		"destinatarios":         reportesSplitCommaList(reportesNullString(destinatarios)),
		"ultimo_ejecutado_en":   reportesNullString(ultimoEjecutado),
		"proximo_ejecutado_en":  reportesNullString(proximoEjecutado),
		"activa":                activa > 0,
		"validar_consistencia":  validarConsistencia > 0,
		"hash_ultima_ejecucion": reportesNullString(hashUltimaEjecucion),
		"fecha_creacion":        fechaCreacion,
		"fecha_actualizacion":   fechaActualizacion,
		"usuario_creador":       reportesNullString(usuarioCreador),
		"estado":                reportesFirstNonBlank(reportesNullString(estado), "activo"),
		"observaciones":         reportesNullString(observaciones),
	}, nil
}

func handleEmpresaReportesEjecucionesAction(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID int64) error {
	if r.Method != http.MethodGet {
		return newReportesHTTPError(http.StatusMethodNotAllowed, "method not allowed")
	}

	query := `SELECT
		id, empresa_id, programacion_id, dataset_key, referencia,
		formato_principal, formatos_json, estado_ejecucion, ejecutado_en,
		consistencia_estado, consistencia_detalle_json, salida_resumen_json,
		error_detalle, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
	FROM empresa_reportes_ejecuciones
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	programacionID, err := parseInt64QueryOptional(r, "programacion_id")
	if err != nil {
		return newReportesHTTPError(http.StatusBadRequest, "programacion_id invalido")
	}
	if programacionID > 0 {
		query += " AND programacion_id = ?"
		args = append(args, programacionID)
	}

	dataset := strings.TrimSpace(r.URL.Query().Get("dataset"))
	if dataset != "" {
		query += " AND lower(dataset_key) = ?"
		args = append(args, strings.ToLower(dataset))
	}

	limit, err := parseIntQueryOptional(r, "limit")
	if err != nil {
		return newReportesHTTPError(http.StatusBadRequest, "limit invalido")
	}
	if limit <= 0 || limit > 500 {
		limit = 200
	}

	query += " ORDER BY id DESC LIMIT ?"
	args = append(args, limit)

	rows, err := dbEmp.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var (
			id, empID                                         int64
			rowProgramacionID                                 sql.NullInt64
			datasetKey                                        string
			referencia, formatoPrincipal, formatosJSON        sql.NullString
			estadoEjecucion, ejecutadoEn                      sql.NullString
			consistenciaEstado, consistenciaDetalleJSON       sql.NullString
			salidaResumenJSON, errorDetalle                   sql.NullString
			fechaCreacion, fechaActualizacion                 sql.NullString
			usuarioCreador, estado, observaciones            sql.NullString
		)
		if err := rows.Scan(
			&id, &empID, &rowProgramacionID, &datasetKey, &referencia,
			&formatoPrincipal, &formatosJSON, &estadoEjecucion, &ejecutadoEn,
			&consistenciaEstado, &consistenciaDetalleJSON, &salidaResumenJSON,
			&errorDetalle, &fechaCreacion, &fechaActualizacion, &usuarioCreador, &estado, &observaciones,
		); err != nil {
			return err
		}

		item := map[string]interface{}{
			"id":                   id,
			"empresa_id":           empID,
			"programacion_id":      reportesNullInt64(rowProgramacionID),
			"dataset_key":          datasetKey,
			"referencia":           reportesNullString(referencia),
			"formato_principal":    reportesFirstNonBlank(reportesNullString(formatoPrincipal), "json"),
			"formatos":             reportesDecodeJSONArray(reportesNullString(formatosJSON)),
			"estado_ejecucion":     reportesFirstNonBlank(reportesNullString(estadoEjecucion), "completado"),
			"ejecutado_en":         reportesNullString(ejecutadoEn),
			"consistencia_estado":  reportesFirstNonBlank(reportesNullString(consistenciaEstado), "pendiente"),
			"consistencia_detalle": reportesDecodeJSONMap(reportesNullString(consistenciaDetalleJSON)),
			"salida_resumen":       reportesDecodeJSONMap(reportesNullString(salidaResumenJSON)),
			"error_detalle":        reportesNullString(errorDetalle),
			"fecha_creacion":       reportesNullString(fechaCreacion),
			"fecha_actualizacion":  reportesNullString(fechaActualizacion),
			"usuario_creador":      reportesNullString(usuarioCreador),
			"estado":               reportesFirstNonBlank(reportesNullString(estado), "activo"),
			"observaciones":        reportesNullString(observaciones),
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"empresa_id":  empresaID,
		"total":       len(items),
		"ejecuciones": items,
	})
	return nil
}

func handleEmpresaReportesExecuteProgramacionAction(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, baseBuilder *reportesBuilder) error {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		return newReportesHTTPError(http.StatusMethodNotAllowed, "method not allowed")
	}

	var payload reporteExecuteProgramacionPayload
	if err := reportesDecodeBodyJSON(r, &payload); err != nil && err != io.EOF {
		return newReportesHTTPError(http.StatusBadRequest, "payload JSON invalido")
	}

	programacionID := payload.ProgramacionID
	if programacionID <= 0 {
		parsedID, err := parseInt64QueryOptional(r, "programacion_id")
		if err != nil {
			return newReportesHTTPError(http.StatusBadRequest, "programacion_id invalido")
		}
		programacionID = parsedID
	}
	if programacionID <= 0 {
		return newReportesHTTPError(http.StatusBadRequest, "programacion_id es obligatorio")
	}

	var (
		datasetKey, formatosJSON, parametrosJSON             string
		templateCodigo, frecuencia, horaEnvio, destinatarios string
		templateVersion                                      int64
		validarConsistencia                                  int
	)
	err := dbEmp.QueryRow(`
		SELECT
			dataset_key, formatos, parametros_json,
			template_codigo, template_version,
			frecuencia, hora_envio, destinatarios,
			validacion_consistencia
		FROM empresa_reportes_programaciones
		WHERE empresa_id = ? AND id = ?
	`, baseBuilder.empresaID, programacionID).Scan(
		&datasetKey, &formatosJSON, &parametrosJSON,
		&templateCodigo, &templateVersion,
		&frecuencia, &horaEnvio, &destinatarios,
		&validarConsistencia,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return newReportesHTTPError(http.StatusNotFound, "programacion no encontrada")
		}
		return err
	}

	scheduleBuilder := &reportesBuilder{
		db:               baseBuilder.db,
		empresaID:        baseBuilder.empresaID,
		desde:            baseBuilder.desde,
		hasta:            baseBuilder.hasta,
		cierreID:         baseBuilder.cierreID,
		empleadoNominaID: baseBuilder.empleadoNominaID,
		cajaCodigo:       baseBuilder.cajaCodigo,
		turno:            baseBuilder.turno,
		usuario:          baseBuilder.usuario,
		categoria:        baseBuilder.categoria,
		metodoPago:       baseBuilder.metodoPago,
		maxRows:          baseBuilder.maxRows,
		includeInactive:  baseBuilder.includeInactive,
		itemsCache:       make(map[int64][]dbpkg.CarritoCompraItem),
	}

	params := reportesDecodeJSONMap(parametrosJSON)
	if v, ok := params["desde"].(string); ok && strings.TrimSpace(scheduleBuilder.desde) == "" {
		scheduleBuilder.desde = strings.TrimSpace(v)
	}
	if v, ok := params["hasta"].(string); ok && strings.TrimSpace(scheduleBuilder.hasta) == "" {
		scheduleBuilder.hasta = strings.TrimSpace(v)
	}
	if v, ok := params["usuario"].(string); ok && strings.TrimSpace(scheduleBuilder.usuario) == "" {
		scheduleBuilder.usuario = strings.TrimSpace(v)
	}
	if v, ok := params["caja_codigo"].(string); ok && strings.TrimSpace(scheduleBuilder.cajaCodigo) == "" {
		scheduleBuilder.cajaCodigo = strings.TrimSpace(v)
	}
	if v, ok := params["turno"].(string); ok && strings.TrimSpace(scheduleBuilder.turno) == "" {
		scheduleBuilder.turno = strings.TrimSpace(v)
	}
	if v, ok := params["categoria"].(string); ok && strings.TrimSpace(scheduleBuilder.categoria) == "" {
		scheduleBuilder.categoria = strings.TrimSpace(v)
	}
	if v, ok := params["metodo_pago"].(string); ok && strings.TrimSpace(scheduleBuilder.metodoPago) == "" {
		scheduleBuilder.metodoPago = strings.TrimSpace(v)
	}
	if v, ok := params["max_rows"]; ok {
		parsed := int(reportesToInt64(v, int64(scheduleBuilder.maxRows)))
		if parsed > 0 {
			scheduleBuilder.maxRows = parsed
		}
	}
	if v, ok := params["cierre_id"]; ok && scheduleBuilder.cierreID <= 0 {
		scheduleBuilder.cierreID = reportesToInt64(v, 0)
	}
	if v, ok := params["empleado_nomina_id"]; ok && scheduleBuilder.empleadoNominaID <= 0 {
		scheduleBuilder.empleadoNominaID = reportesToInt64(v, 0)
	}
	if v, ok := params["include_inactive"]; ok {
		scheduleBuilder.includeInactive = reportesToBool(v, scheduleBuilder.includeInactive)
	}

	ds, err := scheduleBuilder.buildDataset(strings.ToLower(strings.TrimSpace(datasetKey)))
	if err != nil {
		return newReportesHTTPError(http.StatusBadRequest, err.Error())
	}

	var plantilla map[string]interface{}
	if strings.TrimSpace(templateCodigo) != "" {
		plantilla, err = resolveReportePlantilla(dbEmp, baseBuilder.empresaID, templateCodigo, templateVersion)
		if err != nil {
			if err == sql.ErrNoRows {
				return newReportesHTTPError(http.StatusBadRequest, "plantilla asociada no encontrada")
			}
			return err
		}
		ds = reportesApplyTemplate(ds, plantilla)
	}

	formats := reportesNormalizeFormats(reportesDecodeJSONArray(formatosJSON), []string{"json"})
	if len(payload.Formatos) > 0 {
		formats = reportesNormalizeFormats(payload.Formatos, formats)
	}

	consistencia := map[string]interface{}{}
	if validarConsistencia > 0 {
		result, err := reportesValidateDatasetConsistency(ds, formats)
		if err != nil {
			return err
		}
		consistencia = result
	}

	now := reportesCurrentTimestamp()
	referencia := fmt.Sprintf("RP-%d-%s", programacionID, time.Now().Format("20060102150405"))
	usuario := strings.TrimSpace(payload.UsuarioCreador)
	if usuario == "" {
		usuario = "sistema_reportes"
	}

	estadoConsistencia := "no_validada"
	if validarConsistencia > 0 {
		if reportesToBool(consistencia["consistente"], false) {
			estadoConsistencia = "consistente"
		} else {
			estadoConsistencia = "inconsistente"
		}
	}
	estadoEjecucion := "completado"
	if estadoConsistencia == "inconsistente" {
		estadoEjecucion = "completado_con_alertas"
	}

	salidaResumen := map[string]interface{}{
		"dataset_key":   ds.Key,
		"row_count":     ds.RowCount,
		"formatos":      formats,
		"destinatarios": reportesSplitCommaList(destinatarios),
	}
	if plantilla != nil {
		salidaResumen["plantilla_codigo"] = plantilla["codigo"]
		salidaResumen["plantilla_version"] = plantilla["version"]
	}

	consistenciaJSON := reportesMarshalJSON(consistencia, "{}")
	salidaResumenJSON := reportesMarshalJSON(salidaResumen, "{}")
	formatosPersist := reportesMarshalJSON(formats, "[\"json\"]")

	res, err := dbEmp.Exec(`
		INSERT INTO empresa_reportes_ejecuciones (
			empresa_id, programacion_id, dataset_key, referencia,
			formato_principal, formatos_json, estado_ejecucion, ejecutado_en,
			consistencia_estado, consistencia_detalle_json, salida_resumen_json,
			fecha_actualizacion, usuario_creador, estado, observaciones
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'activo', ?)
	`,
		baseBuilder.empresaID,
		programacionID,
		ds.Key,
		referencia,
		formats[0],
		formatosPersist,
		estadoEjecucion,
		now,
		estadoConsistencia,
		consistenciaJSON,
		salidaResumenJSON,
		now,
		usuario,
		"Ejecucion programada de reportes",
	)
	if err != nil {
		return err
	}
	ejecucionID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	hashUltimaEjecucion := ""
	if validarConsistencia > 0 {
		hashUltimaEjecucion = strings.TrimSpace(reportesStringValue(consistencia["estructura_hash"]))
	}
	if hashUltimaEjecucion == "" {
		hashUltimaEjecucion = reportesHashString(ds.Key + "|" + strconv.Itoa(ds.RowCount) + "|" + now)
	}
	proximo := reportesComputeNextExecution(time.Now(), frecuencia, horaEnvio)

	if _, err := dbEmp.Exec(`
		UPDATE empresa_reportes_programaciones
		SET
			ultimo_ejecutado_en = ?,
			proximo_ejecutado_en = ?,
			hash_ultima_ejecucion = ?,
			fecha_actualizacion = ?
		WHERE empresa_id = ? AND id = ?
	`, now, proximo, hashUltimaEjecucion, now, baseBuilder.empresaID, programacionID); err != nil {
		return err
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                   true,
		"empresa_id":           baseBuilder.empresaID,
		"programacion_id":      programacionID,
		"ejecucion_id":         ejecucionID,
		"referencia":           referencia,
		"dataset_key":          ds.Key,
		"row_count":            ds.RowCount,
		"formatos":             formats,
		"consistencia":         consistencia,
		"estado_ejecucion":     estadoEjecucion,
		"estado_consistencia":  estadoConsistencia,
		"destinatarios":        reportesSplitCommaList(destinatarios),
		"ultimo_ejecutado_en":  now,
		"proximo_ejecutado_en": proximo,
	})
	return nil
}

func handleEmpresaReportesConsistenciaAction(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, baseBuilder *reportesBuilder) error {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		return newReportesHTTPError(http.StatusMethodNotAllowed, "method not allowed")
	}

	var payload reporteConsistenciaPayload
	if r.Method == http.MethodPost {
		if err := reportesDecodeBodyJSON(r, &payload); err != nil && err != io.EOF {
			return newReportesHTTPError(http.StatusBadRequest, "payload JSON invalido")
		}
	}

	dataset := strings.ToLower(strings.TrimSpace(reportesFirstNonBlank(
		payload.Dataset,
		payload.DatasetKey,
		r.URL.Query().Get("dataset"),
	)))
	if dataset == "" {
		return newReportesHTTPError(http.StatusBadRequest, "dataset es obligatorio")
	}
	if _, ok := reportesDatasetExists(dataset); !ok {
		return newReportesHTTPError(http.StatusBadRequest, "dataset no soportado")
	}

	formats := make([]string, 0)
	if len(payload.Formatos) > 0 {
		formats = append(formats, payload.Formatos...)
	}
	formats = append(formats, reportesSplitCommaList(r.URL.Query().Get("formatos"))...)
	if format := strings.TrimSpace(r.URL.Query().Get("format")); format != "" {
		formats = append(formats, format)
	}
	formats = reportesNormalizeFormats(formats, []string{"json", "csv", "txt", "xls", "pdf"})

	ds, err := baseBuilder.buildDataset(dataset)
	if err != nil {
		return newReportesHTTPError(http.StatusBadRequest, err.Error())
	}

	templateCodigo := strings.TrimSpace(reportesFirstNonBlank(payload.TemplateCodigo, r.URL.Query().Get("template_codigo")))
	templateVersion := payload.TemplateVersion
	if templateVersion <= 0 {
		parsedVersion, err := parseInt64QueryOptional(r, "template_version")
		if err == nil {
			templateVersion = parsedVersion
		}
	}
	if templateCodigo != "" {
		plantilla, err := resolveReportePlantilla(dbEmp, baseBuilder.empresaID, templateCodigo, templateVersion)
		if err != nil {
			if err == sql.ErrNoRows {
				return newReportesHTTPError(http.StatusBadRequest, "plantilla no encontrada")
			}
			return err
		}
		ds = reportesApplyTemplate(ds, plantilla)
	}

	result, err := reportesValidateDatasetConsistency(ds, formats)
	if err != nil {
		return err
	}
	result["empresa_id"] = baseBuilder.empresaID
	result["dataset_key"] = ds.Key
	result["generated_at"] = ds.GeneratedAt
	writeJSON(w, http.StatusOK, result)
	return nil
}

func reportesDatasetContentForFormat(ds empresaReporteDataset, format string) ([]byte, error) {
	norm := reportesNormalizeFormat(format)
	switch norm {
	case "json":
		raw, err := json.Marshal(ds)
		if err != nil {
			return nil, err
		}
		return raw, nil
	case "csv":
		content, err := reportesDatasetCSVContent(ds)
		if err != nil {
			return nil, err
		}
		return []byte(content), nil
	case "txt":
		return []byte(reportesDatasetTXTContent(ds)), nil
	case "xls":
		return []byte(reportesDatasetTSVContent(ds)), nil
	case "pdf":
		return reportesDatasetPDFContent(ds), nil
	default:
		return nil, fmt.Errorf("formato no soportado")
	}
}

func reportesCountDelimitedRows(content []byte) int {
	trimmed := strings.TrimSpace(string(content))
	if trimmed == "" {
		return 0
	}
	lines := strings.Split(trimmed, "\n")
	if len(lines) <= 1 {
		return 0
	}
	return len(lines) - 1
}

func reportesNormalizeSummary(summary map[string]interface{}) map[string]interface{} {
	if summary == nil {
		return map[string]interface{}{}
	}
	keys := make([]string, 0, len(summary))
	for key := range summary {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make(map[string]interface{}, len(summary))
	for _, key := range keys {
		out[key] = summary[key]
	}
	return out
}

func reportesValidateDatasetConsistency(ds empresaReporteDataset, formats []string) (map[string]interface{}, error) {
	normalizedFormats := reportesNormalizeFormats(formats, []string{"json", "csv", "txt", "xls", "pdf"})
	summaryNormalized := reportesNormalizeSummary(ds.Summary)
	structurePayload := map[string]interface{}{
		"dataset_key": ds.Key,
		"columns":     ds.Columns,
		"row_count":   ds.RowCount,
		"summary":     summaryNormalized,
	}
	structureRaw, err := json.Marshal(structurePayload)
	if err != nil {
		return nil, err
	}
	structureHash := reportesHashBytes(structureRaw)

	alerts := make([]string, 0)
	formatResults := make([]map[string]interface{}, 0, len(normalizedFormats))
	consistente := true

	for _, format := range normalizedFormats {
		content, err := reportesDatasetContentForFormat(ds, format)
		if err != nil {
			consistente = false
			alerts = append(alerts, "No se pudo generar formato "+format)
			formatResults = append(formatResults, map[string]interface{}{
				"format":          format,
				"ok":              false,
				"error":           "generacion_fallida",
				"estructura_hash": structureHash,
			})
			continue
		}

		rowCountDetectado := ds.RowCount
		if format == "csv" || format == "xls" {
			rowCountDetectado = reportesCountDelimitedRows(content)
			if rowCountDetectado != ds.RowCount {
				consistente = false
				alerts = append(alerts, "Conteo de filas inconsistente en formato "+format)
			}
		}

		formatResults = append(formatResults, map[string]interface{}{
			"format":              format,
			"ok":                  true,
			"bytes":               len(content),
			"contenido_hash":      reportesHashBytes(content),
			"row_count_detectado": rowCountDetectado,
			"estructura_hash":     structureHash,
		})
	}

	result := map[string]interface{}{
		"dataset_key":     ds.Key,
		"row_count":       ds.RowCount,
		"columns_count":   len(ds.Columns),
		"summary_keys":    len(summaryNormalized),
		"formatos":        formatResults,
		"consistente":     consistente,
		"alertas":         alerts,
		"estructura_hash": structureHash,
		"validado_en":     reportesCurrentTimestamp(),
	}
	return result, nil
}
