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
	case "restaurar", "restore":
		return "restaurar"
	case "depurar_fecha", "purgar_fecha", "eliminar_hasta_fecha", "depurar_hasta_fecha":
		return "depurar_fecha"
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
func EmpresaBackupsHandler(dbEmp *sql.DB) http.HandlerFunc {
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
			case "depurar_fecha":
				empresaBackupsHandlePurgeByDate(w, r, dbEmp)
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
