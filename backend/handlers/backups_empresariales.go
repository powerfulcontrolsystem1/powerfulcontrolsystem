package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type empresaBackupCreatePayload struct {
	EmpresaID      int64    `json:"empresa_id"`
	Nombre         string   `json:"nombre"`
	Descripcion    string   `json:"descripcion"`
	IncludeTables  []string `json:"include_tables"`
	ExcludeTables  []string `json:"exclude_tables"`
	UsuarioCreador string   `json:"usuario_creador"`
	Observaciones  string   `json:"observaciones"`
}

type empresaBackupRestorePayload struct {
	EmpresaID      int64  `json:"empresa_id"`
	BackupID       int64  `json:"backup_id"`
	UsuarioCreador string `json:"usuario_creador"`
	Observaciones  string `json:"observaciones"`
}

type empresaBackupPurgePayload struct {
	EmpresaID          int64    `json:"empresa_id"`
	FechaCorte         string   `json:"fecha_corte"`
	IncludeTables      []string `json:"include_tables"`
	ExcludeTables      []string `json:"exclude_tables"`
	UsuarioCreador     string   `json:"usuario_creador"`
	Observaciones      string   `json:"observaciones"`
	CrearBackupPrevio  *bool    `json:"crear_backup_previo"`
	NombreBackupPrevio string   `json:"nombre_backup_previo"`
}

type empresaBackupImportPayload struct {
	EmpresaID      int64                       `json:"empresa_id"`
	Nombre         string                      `json:"nombre"`
	Descripcion    string                      `json:"descripcion"`
	UsuarioCreador string                      `json:"usuario_creador"`
	Payload        *dbpkg.EmpresaBackupPayload `json:"payload"`
}

type empresaBackupSendEmailPayload struct {
	EmpresaID int64  `json:"empresa_id"`
	ID        int64  `json:"id"`
	ToEmail   string `json:"to_email"`
	Format    string `json:"format"`
	Subject   string `json:"subject,omitempty"`
	Message   string `json:"message,omitempty"`
}

func empresaBackupsNormalizeAction(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "listar", "list":
		return "listar"
	case "crear", "create", "snapshot":
		return "crear"
	case "detalle", "get", "backup":
		return "detalle"
	case "export", "exportar", "descargar", "download":
		return "export"
	case "exportar_local", "export_local", "descargar_local", "download_local":
		return "exportar_local"
	case "restaurar", "restore":
		return "restaurar"
	case "exportar_configuracion", "export_config", "descargar_configuracion":
		return "exportar_configuracion"
	case "exportar_configuracion_local", "export_config_local", "descargar_configuracion_local":
		return "exportar_configuracion_local"
	case "importar_configuracion", "import_config", "restaurar_configuracion":
		return "importar_configuracion"
	case "depurar_fecha", "purgar_fecha", "eliminar_hasta_fecha", "depurar_hasta_fecha":
		return "depurar_fecha"
	case "enviar_email", "email", "send_email":
		return "enviar_email"
	case "activar":
		return "activar"
	case "desactivar", "eliminar", "delete":
		return "desactivar"
	default:
		return ""
	}
}

func empresaBackupsDefaultActionByMethod(method string) string {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case http.MethodGet:
		return "listar"
	case http.MethodPost:
		return "crear"
	case http.MethodPut, http.MethodPatch:
		return "restaurar"
	case http.MethodDelete:
		return "desactivar"
	default:
		return "listar"
	}
}

func empresaBackupsDecodeBodyJSON(r *http.Request, dst interface{}) error {
	if r == nil || r.Body == nil {
		return io.EOF
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}

func empresaBackupsUsuarioFromRequest(r *http.Request, fallback string) string {
	if raw := strings.TrimSpace(fallback); raw != "" {
		return raw
	}
	if v := strings.TrimSpace(adminEmailFromRequest(r)); v != "" {
		return v
	}
	if v := strings.TrimSpace(r.Header.Get("X-Usuario")); v != "" {
		return v
	}
	return "sistema"
}

func empresaBackupsNormalizeFechaCorte(raw string) (string, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return "", fmt.Errorf("fecha_corte es obligatoria")
	}
	if t, err := time.ParseInLocation("2006-01-02", v, time.Local); err == nil {
		return t.Format("2006-01-02") + " 23:59:59", nil
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, v, time.Local); err == nil {
			return t.Format("2006-01-02 15:04:05"), nil
		}
	}
	return "", fmt.Errorf("fecha_corte invalida")
}

func parseCSVStrings(raw string) []string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		clean := strings.TrimSpace(part)
		if clean == "" {
			continue
		}
		out = append(out, clean)
	}
	return out
}

// EmpresaBackupsHandler gestiona snapshots y restauraciones de datos empresariales.
func EmpresaBackupsHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		action := empresaBackupsNormalizeAction(r.URL.Query().Get("action"))
		if action == "" {
			action = empresaBackupsDefaultActionByMethod(r.Method)
		}

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "listar":
				empresaBackupsHandleList(w, r, dbEmp)
				return
			case "detalle":
				empresaBackupsHandleDetail(w, r, dbEmp)
				return
			case "export":
				empresaBackupsHandleExport(w, r, dbEmp)
				return
			case "exportar_local":
				empresaBackupsHandleExportLocal(w, r, dbEmp)
				return
			case "exportar_configuracion":
				empresaBackupsHandleExportConfig(w, r, dbEmp)
				return
			case "exportar_configuracion_local":
				empresaBackupsHandleExportConfigLocal(w, r, dbEmp)
				return
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
				return
			}
		case http.MethodPost:
			switch action {
			case "crear":
				empresaBackupsHandleCreate(w, r, dbEmp)
				return
			case "restaurar":
				empresaBackupsHandleRestore(w, r, dbEmp)
				return
			case "exportar_configuracion":
				empresaBackupsHandleExportConfig(w, r, dbEmp)
				return
			case "exportar_local":
				empresaBackupsHandleExportLocal(w, r, dbEmp)
				return
			case "exportar_configuracion_local":
				empresaBackupsHandleExportConfigLocal(w, r, dbEmp)
				return
			case "importar_configuracion":
				empresaBackupsHandleImportConfig(w, r, dbEmp)
				return
			case "depurar_fecha":
				empresaBackupsHandlePurgeByDate(w, r, dbEmp)
				return
			case "enviar_email":
				empresaBackupsHandleSendEmail(w, r, dbEmp, dbSuper)
				return
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
				return
			}
		case http.MethodPut, http.MethodPatch:
			switch action {
			case "restaurar":
				empresaBackupsHandleRestore(w, r, dbEmp)
				return
			case "depurar_fecha":
				empresaBackupsHandlePurgeByDate(w, r, dbEmp)
				return
			case "activar", "desactivar":
				empresaBackupsHandleToggle(w, r, dbEmp, action)
				return
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
				return
			}
		case http.MethodDelete:
			empresaBackupsHandleToggle(w, r, dbEmp, "desactivar")
			return
		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
	}
}

func empresaBackupsHandleSendEmail(w http.ResponseWriter, r *http.Request, dbEmp, dbSuper *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if dbSuper == nil {
		http.Error(w, "configuracion de correo no disponible", http.StatusInternalServerError)
		return
	}

	var payload empresaBackupSendEmailPayload
	if err := empresaBackupsDecodeBodyJSON(r, &payload); err != nil {
		http.Error(w, "payload JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID > 0 && payload.EmpresaID != empresaID {
		http.Error(w, "empresa_id no coincide con el contexto", http.StatusBadRequest)
		return
	}
	if payload.ID <= 0 {
		http.Error(w, "id es obligatorio", http.StatusBadRequest)
		return
	}

	format := strings.ToLower(strings.TrimSpace(payload.Format))
	if format == "" {
		format = "pdf"
	}
	if format == "json.gz" || format == "gz" || format == "gzip" {
		http.Error(w, "format invalido para correo (use json, csv, txt, xls o pdf)", http.StatusBadRequest)
		return
	}

	backupMeta, exportPayload, err := dbpkg.GetEmpresaBackupPayloadByID(dbEmp, empresaID, payload.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "backup no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo consultar export del backup", http.StatusInternalServerError)
		return
	}

	subject := strings.TrimSpace(payload.Subject)
	if subject == "" {
		subject = fmt.Sprintf("Backup empresa %d: %s", empresaID, strings.TrimSpace(backupMeta.Codigo))
	}
	body := strings.TrimSpace(payload.Message)
	if body == "" {
		body = "Adjunto encontrarás el respaldo solicitado."
	}

	var fileName, contentType string
	var content []byte

	if format == "json" {
		fileName = fmt.Sprintf("backup_empresa_%d_%s.json", empresaID, strings.TrimSpace(backupMeta.Codigo))
		raw, jerr := json.Marshal(exportPayload)
		if jerr != nil {
			http.Error(w, "no se pudo serializar el backup", http.StatusInternalServerError)
			return
		}
		contentType = "application/json"
		content = raw
	} else {
		// Reusar el dataset resumen (mismo que export) para formatos tabulares/PDF/TXT.
		rows := make([]map[string]interface{}, 0, len(exportPayload.Tables))
		for _, table := range exportPayload.Tables {
			rows = append(rows, map[string]interface{}{
				"tabla":            table.Table,
				"columnas":         len(table.Columns),
				"registros":        len(table.Rows),
				"columnas_detalle": strings.Join(table.Columns, ","),
			})
		}
		ds := empresaReporteDataset{
			Key:         "seguridad_backups_empresariales",
			Title:       "Backups empresariales",
			Level:       "seguridad",
			Description: "Resumen de tablas incluidas en snapshot empresarial para respaldo y restauracion.",
			EmpresaID:   empresaID,
			GeneratedAt: time.Now().In(time.Local).Format("2006-01-02 15:04:05"),
			Columns: []string{
				"tabla",
				"columnas",
				"registros",
				"columnas_detalle",
			},
			Rows:     rows,
			RowCount: len(rows),
			Summary: map[string]interface{}{
				"backup_id":     payload.ID,
				"codigo_backup": backupMeta.Codigo,
				"version":       exportPayload.Version,
				"tablas":        exportPayload.TotalTables,
				"registros":     exportPayload.TotalRows,
				"creado_en":     exportPayload.CreatedAt,
				"creado_por":    exportPayload.CreatedBy,
			},
		}
		fileName, contentType, content, err = reportesBuildExportBytes(ds, format)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	metaJSON := fmt.Sprintf(`{"scope":"empresa_backups","empresa_id":%d,"backup_id":%d,"backup_code":%q,"format":%q}`, empresaID, payload.ID, strings.TrimSpace(backupMeta.Codigo), format)
	if err := sendReportesEmailWithAttachment(r, dbSuper, empresaID, payload.ToEmail, subject, body, fileName, contentType, content, metaJSON); err != nil {
		http.Error(w, "no se pudo enviar el correo: "+err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":       true,
		"to_email": strings.TrimSpace(payload.ToEmail),
		"filename": fileName,
		"format":   format,
	})
}

func empresaBackupsHandleList(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
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

	rows, total, err := dbpkg.ListEmpresaBackups(dbEmp, empresaID, dbpkg.EmpresaBackupFilter{
		IncludeInactive: queryBool(r, "include_inactive"),
		Q:               strings.TrimSpace(r.URL.Query().Get("q")),
		Limit:           limit,
		Offset:          offset,
	})
	if err != nil {
		http.Error(w, "No se pudo consultar backups empresariales", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"total":      total,
		"rows":       rows,
	})
}

func empresaBackupsHandleCreate(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var payload empresaBackupCreatePayload
	if err := empresaBackupsDecodeBodyJSON(r, &payload); err != nil && err != io.EOF {
		http.Error(w, "payload JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID > 0 && payload.EmpresaID != empresaID {
		http.Error(w, "empresa_id no coincide con el contexto", http.StatusBadRequest)
		return
	}

	usuario := empresaBackupsUsuarioFromRequest(r, payload.UsuarioCreador)
	backupID, err := dbpkg.CreateEmpresaBackupSnapshot(dbEmp, empresaID, payload.Nombre, payload.Descripcion, usuario, dbpkg.EmpresaBackupBuildOptions{
		IncludeTables: payload.IncludeTables,
		ExcludeTables: payload.ExcludeTables,
		CreatedBy:     usuario,
	})
	if err != nil {
		http.Error(w, "No se pudo generar backup empresarial", http.StatusInternalServerError)
		return
	}

	created, err := dbpkg.GetEmpresaBackupByID(dbEmp, empresaID, backupID, false)
	if err != nil {
		http.Error(w, "backup generado pero no se pudo consultar", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"backup":     created,
	})
}

func empresaBackupsHandleDetail(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	backupID, err := parseInt64QueryOptional(r, "id")
	if err != nil || backupID <= 0 {
		http.Error(w, "id es obligatorio", http.StatusBadRequest)
		return
	}
	includeSnapshot := queryBool(r, "include_snapshot")

	if includeSnapshot {
		backupMeta, payload, err := dbpkg.GetEmpresaBackupPayloadByID(dbEmp, empresaID, backupID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "backup no encontrado", http.StatusNotFound)
				return
			}
			http.Error(w, "No se pudo consultar detalle del backup", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":         true,
			"empresa_id": empresaID,
			"backup":     backupMeta,
			"payload":    payload,
		})
		return
	}

	backup, err := dbpkg.GetEmpresaBackupByID(dbEmp, empresaID, backupID, false)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "backup no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo consultar detalle del backup", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"backup":     backup,
	})
}

func empresaBackupsHandleExport(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	backupID, err := parseInt64QueryOptional(r, "id")
	if err != nil || backupID <= 0 {
		http.Error(w, "id es obligatorio", http.StatusBadRequest)
		return
	}

	backupMeta, payload, err := dbpkg.GetEmpresaBackupPayloadByID(dbEmp, empresaID, backupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "backup no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo consultar export del backup", http.StatusInternalServerError)
		return
	}

	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	if format == "" {
		format = "json"
	}
	if format == "json" {
		fileName := fmt.Sprintf("backup_empresa_%d_%s.json", empresaID, strings.TrimSpace(backupMeta.Codigo))
		w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
		writeJSON(w, http.StatusOK, payload)
		return
	}

	rows := make([]map[string]interface{}, 0, len(payload.Tables))
	for _, table := range payload.Tables {
		rows = append(rows, map[string]interface{}{
			"tabla":            table.Table,
			"columnas":         len(table.Columns),
			"registros":        len(table.Rows),
			"columnas_detalle": strings.Join(table.Columns, ","),
		})
	}

	ds := empresaReporteDataset{
		Key:         "seguridad_backups_empresariales",
		Title:       "Backups empresariales",
		Level:       "seguridad",
		Description: "Resumen de tablas incluidas en snapshot empresarial para respaldo y restauracion.",
		EmpresaID:   empresaID,
		Desde:       "",
		Hasta:       "",
		GeneratedAt: time.Now().In(time.Local).Format("2006-01-02 15:04:05"),
		Columns: []string{
			"tabla",
			"columnas",
			"registros",
			"columnas_detalle",
		},
		Rows:     rows,
		RowCount: len(rows),
		Summary: map[string]interface{}{
			"backup_id":     backupID,
			"codigo_backup": backupMeta.Codigo,
			"version":       payload.Version,
			"tablas":        payload.TotalTables,
			"registros":     payload.TotalRows,
			"creado_en":     payload.CreatedAt,
			"creado_por":    payload.CreatedBy,
		},
	}

	if err := writeReportesDatasetExport(w, ds, format); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func empresaBackupLocalFilename(prefix string, empresaID int64) string {
	cleanPrefix := strings.TrimSpace(prefix)
	if cleanPrefix == "" {
		cleanPrefix = "backup_empresa"
	}
	return fmt.Sprintf("%s_%d_%s.json", cleanPrefix, empresaID, time.Now().In(time.Local).Format("20060102_150405"))
}

func empresaBackupsDecodeLocalExportPayload(r *http.Request) (empresaBackupCreatePayload, error) {
	var payload empresaBackupCreatePayload
	if r == nil {
		return payload, nil
	}
	if r.Method == http.MethodGet {
		payload.IncludeTables = parseCSVStrings(r.URL.Query().Get("include_tables"))
		payload.ExcludeTables = parseCSVStrings(r.URL.Query().Get("exclude_tables"))
		return payload, nil
	}
	if err := empresaBackupsDecodeBodyJSON(r, &payload); err != nil && err != io.EOF {
		return payload, err
	}
	return payload, nil
}

func empresaBackupsHandleExportLocal(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	payloadReq, err := empresaBackupsDecodeLocalExportPayload(r)
	if err != nil {
		http.Error(w, "payload JSON invalido", http.StatusBadRequest)
		return
	}
	if payloadReq.EmpresaID > 0 && payloadReq.EmpresaID != empresaID {
		http.Error(w, "empresa_id no coincide con el contexto", http.StatusBadRequest)
		return
	}
	usuario := empresaBackupsUsuarioFromRequest(r, payloadReq.UsuarioCreador)
	payload, err := dbpkg.BuildEmpresaBackupPayload(dbEmp, empresaID, dbpkg.EmpresaBackupBuildOptions{
		IncludeTables: payloadReq.IncludeTables,
		ExcludeTables: payloadReq.ExcludeTables,
		CreatedBy:     usuario,
	})
	if err != nil {
		http.Error(w, "No se pudo generar backup local empresarial", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=\""+empresaBackupLocalFilename("backup_empresa_local", empresaID)+"\"")
	w.Header().Set("X-PCS-Storage", "cliente")
	writeJSON(w, http.StatusOK, payload)
}

func empresaBackupsConfigDefaultName(prefix string, empresaID int64) string {
	base := strings.TrimSpace(prefix)
	if base == "" {
		base = "Configuracion empresa"
	}
	return fmt.Sprintf("%s %d %s", base, empresaID, time.Now().In(time.Local).Format("2006-01-02 15:04:05"))
}

func empresaBackupsHandleExportConfigLocal(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var payloadReq empresaBackupCreatePayload
	if r.Method != http.MethodGet {
		if err := empresaBackupsDecodeBodyJSON(r, &payloadReq); err != nil && err != io.EOF {
			http.Error(w, "payload JSON invalido", http.StatusBadRequest)
			return
		}
	}
	if payloadReq.EmpresaID > 0 && payloadReq.EmpresaID != empresaID {
		http.Error(w, "empresa_id no coincide con el contexto", http.StatusBadRequest)
		return
	}
	usuario := empresaBackupsUsuarioFromRequest(r, payloadReq.UsuarioCreador)
	payload, _, err := dbpkg.BuildEmpresaConfigBackupPayload(dbEmp, empresaID, usuario)
	if err != nil {
		http.Error(w, "No se pudo exportar configuracion empresarial local", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=\""+empresaBackupLocalFilename("configuracion_empresa_local", empresaID)+"\"")
	w.Header().Set("X-PCS-Storage", "cliente")
	writeJSON(w, http.StatusOK, payload)
}

func empresaBackupsHandleExportConfig(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var payload empresaBackupCreatePayload
	if err := empresaBackupsDecodeBodyJSON(r, &payload); err != nil && err != io.EOF {
		http.Error(w, "payload JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID > 0 && payload.EmpresaID != empresaID {
		http.Error(w, "empresa_id no coincide con el contexto", http.StatusBadRequest)
		return
	}

	usuario := empresaBackupsUsuarioFromRequest(r, payload.UsuarioCreador)
	nombre := strings.TrimSpace(payload.Nombre)
	if nombre == "" {
		nombre = empresaBackupsConfigDefaultName("Configuracion empresa", empresaID)
	}
	descripcion := strings.TrimSpace(payload.Descripcion)
	if descripcion == "" {
		descripcion = "exportacion de configuracion por empresa"
	}

	backupID, err := dbpkg.CreateEmpresaConfigBackupSnapshot(dbEmp, empresaID, nombre, descripcion, usuario)
	if err != nil {
		http.Error(w, "No se pudo exportar configuracion empresarial", http.StatusInternalServerError)
		return
	}

	backupMeta, exportPayload, err := dbpkg.GetEmpresaBackupPayloadByID(dbEmp, empresaID, backupID)
	if err != nil {
		http.Error(w, "configuracion exportada pero no se pudo consultar", http.StatusInternalServerError)
		return
	}

	fileName := fmt.Sprintf("configuracion_empresa_%d_%s.json", empresaID, strings.TrimSpace(backupMeta.Codigo))
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	w.Header().Set("X-Backup-Id", strconv.FormatInt(backupID, 10))
	w.Header().Set("X-Backup-Code", strings.TrimSpace(backupMeta.Codigo))
	writeJSON(w, http.StatusOK, exportPayload)
}

func empresaBackupsDecodeImportedPayload(r *http.Request) (*empresaBackupImportPayload, *dbpkg.EmpresaBackupPayload, error) {
	if r == nil || r.Body == nil {
		return nil, nil, io.EOF
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, err
	}
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return nil, nil, io.EOF
	}

	var wrapped empresaBackupImportPayload
	if err := json.Unmarshal(raw, &wrapped); err == nil && wrapped.Payload != nil {
		return &wrapped, wrapped.Payload, nil
	}

	var payload dbpkg.EmpresaBackupPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, nil, err
	}
	return &empresaBackupImportPayload{}, &payload, nil
}

func empresaBackupsSanitizeConfigPayload(targetEmpresaID int64, payload *dbpkg.EmpresaBackupPayload, usuario string) (*dbpkg.EmpresaBackupPayload, int64, error) {
	if payload == nil {
		return nil, 0, fmt.Errorf("payload de configuracion invalido")
	}
	allowed := map[string]struct{}{}
	for _, table := range dbpkg.EmpresaConfigBackupDefaultTables() {
		allowed[table] = struct{}{}
	}

	sourceEmpresaID := payload.EmpresaID
	tables := make([]dbpkg.EmpresaBackupTablePayload, 0, len(payload.Tables))
	totalRows := int64(0)
	for _, table := range payload.Tables {
		name := strings.ToLower(strings.TrimSpace(table.Table))
		if _, ok := allowed[name]; !ok {
			continue
		}
		table.Table = name
		tables = append(tables, table)
		totalRows += int64(len(table.Rows))
	}
	if len(tables) == 0 {
		return nil, sourceEmpresaID, fmt.Errorf("el archivo no contiene tablas de configuracion compatibles")
	}

	createdAt := strings.TrimSpace(payload.CreatedAt)
	if createdAt == "" {
		createdAt = time.Now().In(time.Local).Format(time.RFC3339)
	}
	createdBy := strings.TrimSpace(payload.CreatedBy)
	if createdBy == "" {
		createdBy = strings.TrimSpace(usuario)
	}

	return &dbpkg.EmpresaBackupPayload{
		Version:     "empresa-backup.v1",
		Scope:       "configuracion_empresa",
		EmpresaID:   targetEmpresaID,
		CreatedAt:   createdAt,
		CreatedBy:   createdBy,
		TotalTables: len(tables),
		TotalRows:   totalRows,
		Tables:      tables,
	}, sourceEmpresaID, nil
}

func empresaBackupsHandleImportConfig(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	wrapped, payload, err := empresaBackupsDecodeImportedPayload(r)
	if err != nil {
		http.Error(w, "payload JSON invalido", http.StatusBadRequest)
		return
	}
	if wrapped != nil && wrapped.EmpresaID > 0 && wrapped.EmpresaID != empresaID {
		http.Error(w, "empresa_id no coincide con el contexto", http.StatusBadRequest)
		return
	}

	usuarioFallback := ""
	if wrapped != nil {
		usuarioFallback = wrapped.UsuarioCreador
	}
	usuario := empresaBackupsUsuarioFromRequest(r, usuarioFallback)
	sanitized, sourceEmpresaID, err := empresaBackupsSanitizeConfigPayload(empresaID, payload, usuario)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	nombre := strings.TrimSpace(wrapped.Nombre)
	if nombre == "" {
		nombre = empresaBackupsConfigDefaultName("Importacion configuracion", empresaID)
	}
	descripcion := strings.TrimSpace(wrapped.Descripcion)
	if descripcion == "" {
		descripcion = "importacion de configuracion por empresa"
	}

	backupID, err := dbpkg.CreateEmpresaConfigBackupSnapshotFromPayload(dbEmp, empresaID, nombre, descripcion, usuario, sanitized, sourceEmpresaID)
	if err != nil {
		http.Error(w, "No se pudo registrar la importacion de configuracion", http.StatusInternalServerError)
		return
	}

	result, err := dbpkg.RestoreEmpresaBackupByID(dbEmp, empresaID, backupID, usuario, "restauracion de configuracion importada")
	if err != nil {
		http.Error(w, "No se pudo aplicar la configuracion importada", http.StatusInternalServerError)
		return
	}

	updated, _ := dbpkg.GetEmpresaBackupByID(dbEmp, empresaID, backupID, false)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                true,
		"empresa_id":        empresaID,
		"source_empresa_id": sourceEmpresaID,
		"resultado":         result,
		"backup":            updated,
	})
}

func empresaBackupsHandleRestore(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	backupID, _ := parseInt64QueryOptional(r, "id")
	var payload empresaBackupRestorePayload
	if err := empresaBackupsDecodeBodyJSON(r, &payload); err != nil && err != io.EOF {
		http.Error(w, "payload JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID > 0 && payload.EmpresaID != empresaID {
		http.Error(w, "empresa_id no coincide con el contexto", http.StatusBadRequest)
		return
	}
	if payload.BackupID > 0 {
		backupID = payload.BackupID
	}
	if backupID <= 0 {
		http.Error(w, "backup_id es obligatorio", http.StatusBadRequest)
		return
	}

	usuario := empresaBackupsUsuarioFromRequest(r, payload.UsuarioCreador)
	result, err := dbpkg.RestoreEmpresaBackupByID(dbEmp, empresaID, backupID, usuario, payload.Observaciones)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "backup no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo restaurar backup empresarial", http.StatusInternalServerError)
		return
	}

	updated, _ := dbpkg.GetEmpresaBackupByID(dbEmp, empresaID, backupID, false)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"resultado":  result,
		"backup":     updated,
	})
}

func empresaBackupsHandlePurgeByDate(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var payload empresaBackupPurgePayload
	if err := empresaBackupsDecodeBodyJSON(r, &payload); err != nil && err != io.EOF {
		http.Error(w, "payload JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID > 0 && payload.EmpresaID != empresaID {
		http.Error(w, "empresa_id no coincide con el contexto", http.StatusBadRequest)
		return
	}

	fechaCorteRaw := strings.TrimSpace(payload.FechaCorte)
	if fechaCorteRaw == "" {
		fechaCorteRaw = strings.TrimSpace(r.URL.Query().Get("fecha_corte"))
	}
	fechaCorte, err := empresaBackupsNormalizeFechaCorte(fechaCorteRaw)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(payload.IncludeTables) == 0 {
		payload.IncludeTables = parseCSVStrings(r.URL.Query().Get("include_tables"))
	}
	if len(payload.ExcludeTables) == 0 {
		payload.ExcludeTables = parseCSVStrings(r.URL.Query().Get("exclude_tables"))
	}

	usuario := empresaBackupsUsuarioFromRequest(r, payload.UsuarioCreador)
	crearBackupPrevio := true
	if payload.CrearBackupPrevio != nil {
		crearBackupPrevio = *payload.CrearBackupPrevio
	}

	var backupPrevio *dbpkg.EmpresaBackup
	if crearBackupPrevio {
		nombreBackup := strings.TrimSpace(payload.NombreBackupPrevio)
		if nombreBackup == "" {
			nombreBackup = "Backup previo depuracion hasta " + fechaCorte
		}
		obsBackup := strings.TrimSpace(payload.Observaciones)
		if obsBackup == "" {
			obsBackup = "backup previo antes de depuracion por fecha"
		}
		backupID, backupErr := dbpkg.CreateEmpresaBackupSnapshot(dbEmp, empresaID, nombreBackup, obsBackup, usuario, dbpkg.EmpresaBackupBuildOptions{
			IncludeTables: payload.IncludeTables,
			ExcludeTables: payload.ExcludeTables,
			CreatedBy:     usuario,
		})
		if backupErr != nil {
			http.Error(w, "No se pudo crear backup previo a la depuracion", http.StatusInternalServerError)
			return
		}
		backupPrevio, _ = dbpkg.GetEmpresaBackupByID(dbEmp, empresaID, backupID, false)
	}

	result, err := dbpkg.PurgeEmpresaDataByDateCorte(dbEmp, empresaID, fechaCorte, payload.IncludeTables, payload.ExcludeTables)
	if err != nil {
		http.Error(w, "No se pudo depurar informacion por fecha", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                   true,
		"empresa_id":           empresaID,
		"action":               "depurar_fecha",
		"fecha_corte":          fechaCorte,
		"resultado":            result,
		"backup_previo":        backupPrevio,
		"backup_previo_creado": backupPrevio != nil,
	})
}

func empresaBackupsHandleToggle(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, action string) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	backupID, err := parseInt64QueryOptional(r, "id")
	if err != nil || backupID <= 0 {
		http.Error(w, "id es obligatorio", http.StatusBadRequest)
		return
	}

	state := "activo"
	if action == "desactivar" {
		state = "inactivo"
	}
	if err := dbpkg.SetEmpresaBackupEstadoByID(dbEmp, empresaID, backupID, state); err != nil {
		http.Error(w, "No se pudo actualizar estado del backup", http.StatusInternalServerError)
		return
	}

	updated, _ := dbpkg.GetEmpresaBackupByID(dbEmp, empresaID, backupID, false)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"id":         backupID,
		"estado":     state,
		"backup":     updated,
	})
}

func empresaBackupsBuildTogglePath(empresaID, backupID int64, action string) string {
	if strings.TrimSpace(action) == "" {
		action = "desactivar"
	}
	return "/api/empresa/backups?empresa_id=" + strconv.FormatInt(empresaID, 10) + "&id=" + strconv.FormatInt(backupID, 10) + "&action=" + action
}
