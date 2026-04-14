package db

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

const empresaBackupsSchemaVersion = "empresa-backup.v1"

// EmpresaBackup representa un snapshot empresarial persistido para restauracion.
type EmpresaBackup struct {
	ID                 int64             `json:"id"`
	EmpresaID          int64             `json:"empresa_id"`
	Codigo             string            `json:"codigo"`
	Nombre             string            `json:"nombre"`
	Descripcion        string            `json:"descripcion,omitempty"`
	VersionSchema      string            `json:"version_schema"`
	Alcance            string            `json:"alcance"`
	TipoBackup         string            `json:"tipo_backup"`
	IncludeTables      []string          `json:"include_tables,omitempty"`
	ExcludeTables      []string          `json:"exclude_tables,omitempty"`
	TotalTablas        int               `json:"total_tablas"`
	TotalRegistros     int64             `json:"total_registros"`
	TamanoBytes        int64             `json:"tamano_bytes"`
	HashContenido      string            `json:"hash_contenido,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
	FechaCreacion      string            `json:"fecha_creacion,omitempty"`
	FechaActualizacion string            `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string            `json:"usuario_creador,omitempty"`
	Estado             string            `json:"estado,omitempty"`
	Observaciones      string            `json:"observaciones,omitempty"`
	RestauradoEn       string            `json:"restaurado_en,omitempty"`
	RestauradoPor      string            `json:"restaurado_por,omitempty"`
	SnapshotJSON       string            `json:"snapshot_json,omitempty"`
}

// EmpresaBackupTablePayload representa una tabla incluida dentro del snapshot.
type EmpresaBackupTablePayload struct {
	Table   string                   `json:"table"`
	Columns []string                 `json:"columns"`
	Rows    []map[string]interface{} `json:"rows"`
}

// EmpresaBackupPayload representa un respaldo empresarial portable en JSON.
type EmpresaBackupPayload struct {
	Version     string                      `json:"version"`
	Scope       string                      `json:"scope"`
	EmpresaID   int64                       `json:"empresa_id"`
	CreatedAt   string                      `json:"created_at"`
	CreatedBy   string                      `json:"created_by,omitempty"`
	TotalTables int                         `json:"total_tables"`
	TotalRows   int64                       `json:"total_rows"`
	Tables      []EmpresaBackupTablePayload `json:"tables"`
}

// EmpresaBackupBuildOptions define filtros de construccion para snapshot.
type EmpresaBackupBuildOptions struct {
	IncludeTables []string
	ExcludeTables []string
	CreatedBy     string
}

// EmpresaBackupFilter aplica filtros de consulta para historial de backups.
type EmpresaBackupFilter struct {
	IncludeInactive bool
	Q               string
	Limit           int
	Offset          int
}

// EmpresaBackupRestoreResult resume el resultado de restaurar un backup.
type EmpresaBackupRestoreResult struct {
	EmpresaID            int64    `json:"empresa_id"`
	BackupID             int64    `json:"backup_id"`
	CodigoBackup         string   `json:"codigo_backup"`
	TablasRestauradas    int      `json:"tablas_restauradas"`
	RegistrosRestaurados int64    `json:"registros_restaurados"`
	TablasOmitidas       []string `json:"tablas_omitidas,omitempty"`
	EjecutadoEn          string   `json:"ejecutado_en"`
	EjecutadoPor         string   `json:"ejecutado_por,omitempty"`
}

// EmpresaBackupPurgeTableResult resume eliminacion por tabla en una depuracion por fecha.
type EmpresaBackupPurgeTableResult struct {
	Table      string `json:"table"`
	DateColumn string `json:"date_column"`
	Deleted    int64  `json:"deleted"`
}

// EmpresaBackupPurgeResult resume una depuracion de datos empresariales por fecha de corte.
type EmpresaBackupPurgeResult struct {
	EmpresaID           int64                           `json:"empresa_id"`
	FechaCorte          string                          `json:"fecha_corte"`
	TablasEvaluadas     int                             `json:"tablas_evaluadas"`
	TablasDepuradas     int                             `json:"tablas_depuradas"`
	RegistrosEliminados int64                           `json:"registros_eliminados"`
	TablasSinFecha      []string                        `json:"tablas_sin_fecha,omitempty"`
	Detalle             []EmpresaBackupPurgeTableResult `json:"detalle,omitempty"`
}

func normalizeEmpresaBackupEstado(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "inactivo") {
		return "inactivo"
	}
	return "activo"
}

func empresaBackupNormalizeLimitOffset(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func empresaBackupLikePattern(raw string) string {
	value := strings.TrimSpace(raw)
	value = strings.ReplaceAll(value, "!", "!!")
	value = strings.ReplaceAll(value, "%", "!%")
	value = strings.ReplaceAll(value, "_", "!_")
	return "%" + value + "%"
}

func empresaBackupNormalizeTables(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, item := range values {
		clean := strings.TrimSpace(item)
		if clean == "" || !isSafeSQLIdentifier(clean) {
			continue
		}
		clean = strings.ToLower(clean)
		if seen[clean] {
			continue
		}
		seen[clean] = true
		out = append(out, clean)
	}
	sort.Strings(out)
	return out
}

func empresaBackupEncodeTablesJSON(values []string) string {
	normalized := empresaBackupNormalizeTables(values)
	if len(normalized) == 0 {
		return "[]"
	}
	raw, err := json.Marshal(normalized)
	if err != nil {
		return "[]"
	}
	return string(raw)
}

func empresaBackupDecodeTablesJSON(raw string) []string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return []string{}
	}
	out := make([]string, 0)
	if err := json.Unmarshal([]byte(value), &out); err != nil {
		return []string{}
	}
	return empresaBackupNormalizeTables(out)
}

func empresaBackupEncodeMetadataJSON(metadata map[string]string) string {
	if len(metadata) == 0 {
		return "{}"
	}
	raw, err := json.Marshal(metadata)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func empresaBackupDecodeMetadataJSON(raw string) map[string]string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return map[string]string{}
	}
	out := map[string]string{}
	if err := json.Unmarshal([]byte(value), &out); err != nil {
		return map[string]string{}
	}
	return out
}

func empresaBackupHash(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func empresaBackupGenerateCode(empresaID int64) string {
	now := time.Now().In(time.Local)
	return fmt.Sprintf("BKP-%d-%s-%d", empresaID, now.Format("20060102150405"), now.UnixNano())
}

func empresaBackupExcludedInternalTable(table string) bool {
	table = strings.ToLower(strings.TrimSpace(table))
	switch table {
	case "empresa_backups", "empresa_backups_restauraciones":
		return true
	default:
		return false
	}
}

func empresaBackupResolveDateColumn(columns []string) string {
	candidates := []string{
		"fecha_creacion",
		"fecha_evento",
		"fecha_movimiento",
		"pagado_en",
		"activado_en",
		"fecha",
	}
	for _, candidate := range candidates {
		if empresaBackupHasColumn(columns, candidate) {
			return candidate
		}
	}
	return ""
}

type empresaBackupQueryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

func empresaBackupGetTableColumns(q empresaBackupQueryer, table string) ([]string, error) {
	if !isSafeSQLIdentifier(table) {
		return nil, fmt.Errorf("tabla invalida: %s", table)
	}

	if isPostgresDialect() {
		rows, err := q.Query(`
			SELECT column_name
			FROM information_schema.columns
			WHERE table_schema = ANY (current_schemas(false))
			  AND table_name = $1
			ORDER BY ordinal_position
		`, table)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		cols := make([]string, 0)
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return nil, err
			}
			clean := strings.TrimSpace(strings.ToLower(name))
			if clean == "" || !isSafeSQLIdentifier(clean) {
				continue
			}
			cols = append(cols, clean)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return cols, nil
	}

	rows, err := q.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols := make([]string, 0)
	for rows.Next() {
		var (
			cid      int
			name     string
			colType  string
			notNull  int
			defaultV sql.NullString
			pk       int
		)
		if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultV, &pk); err != nil {
			return nil, err
		}
		clean := strings.TrimSpace(strings.ToLower(name))
		if clean == "" || !isSafeSQLIdentifier(clean) {
			continue
		}
		cols = append(cols, clean)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return cols, nil
}

func empresaBackupHasColumn(columns []string, column string) bool {
	for _, item := range columns {
		if strings.EqualFold(strings.TrimSpace(item), strings.TrimSpace(column)) {
			return true
		}
	}
	return false
}

func empresaBackupListCandidateTables(dbConn *sql.DB, includeTables, excludeTables []string) ([]string, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}

	include := empresaBackupNormalizeTables(includeTables)
	exclude := empresaBackupNormalizeTables(excludeTables)
	includeSet := map[string]bool{}
	excludeSet := map[string]bool{}
	for _, item := range include {
		includeSet[item] = true
	}
	for _, item := range exclude {
		excludeSet[item] = true
	}

	tablesQuery := `SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`
	if isPostgresDialect() {
		tablesQuery = `
			SELECT table_name AS name
			FROM information_schema.tables
			WHERE table_schema = ANY (current_schemas(false))
			  AND table_type = 'BASE TABLE'
			ORDER BY table_name
		`
	}

	rows, err := dbConn.Query(tablesQuery)
	if err != nil {
		return nil, err
	}
	tableNames := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return nil, err
		}
		tableNames = append(tableNames, name)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	out := make([]string, 0)
	for _, name := range tableNames {
		table := strings.ToLower(strings.TrimSpace(name))
		if table == "" || !isSafeSQLIdentifier(table) {
			continue
		}
		if empresaBackupExcludedInternalTable(table) {
			continue
		}
		if len(includeSet) > 0 && !includeSet[table] {
			continue
		}
		if excludeSet[table] {
			continue
		}

		cols, colErr := empresaBackupGetTableColumns(dbConn, table)
		if colErr != nil {
			continue
		}
		if !empresaBackupHasColumn(cols, "empresa_id") {
			continue
		}
		out = append(out, table)
	}
	sort.Strings(out)
	return out, nil
}

func empresaBackupFetchTableSnapshot(dbConn *sql.DB, table string, empresaID int64) (EmpresaBackupTablePayload, error) {
	if dbConn == nil {
		return EmpresaBackupTablePayload{}, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return EmpresaBackupTablePayload{}, errors.New("empresa_id invalido")
	}
	if table == "" || !isSafeSQLIdentifier(table) {
		return EmpresaBackupTablePayload{}, fmt.Errorf("tabla invalida: %s", table)
	}

	columns, err := empresaBackupGetTableColumns(dbConn, table)
	if err != nil {
		return EmpresaBackupTablePayload{}, err
	}
	if !empresaBackupHasColumn(columns, "empresa_id") {
		return EmpresaBackupTablePayload{}, fmt.Errorf("tabla %s no contiene empresa_id", table)
	}

	query := "SELECT * FROM " + table + " WHERE empresa_id = ?"
	if empresaBackupHasColumn(columns, "id") {
		query += " ORDER BY id ASC"
	}
	rows, err := dbConn.Query(query, empresaID)
	if err != nil {
		return EmpresaBackupTablePayload{}, err
	}
	defer rows.Close()

	items, err := rowsToMapSlice(rows)
	if err != nil {
		return EmpresaBackupTablePayload{}, err
	}

	return EmpresaBackupTablePayload{
		Table:   table,
		Columns: columns,
		Rows:    items,
	}, nil
}

// EnsureEmpresaBackupsSchema crea y migra tablas para snapshots y restauraciones por empresa.
func EnsureEmpresaBackupsSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_backups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			descripcion TEXT,
			version_schema TEXT NOT NULL DEFAULT 'empresa-backup.v1',
			alcance TEXT NOT NULL DEFAULT 'empresa',
			tipo_backup TEXT NOT NULL DEFAULT 'full',
			include_tables_json TEXT DEFAULT '[]',
			exclude_tables_json TEXT DEFAULT '[]',
			total_tablas INTEGER DEFAULT 0,
			total_registros INTEGER DEFAULT 0,
			tamano_bytes INTEGER DEFAULT 0,
			hash_contenido TEXT,
			snapshot_json TEXT NOT NULL,
			metadata_json TEXT DEFAULT '{}',
			restaurado_en TEXT,
			restaurado_por TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_backups_empresa_fecha ON empresa_backups(empresa_id, fecha_creacion DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_backups_empresa_estado ON empresa_backups(empresa_id, estado, id DESC);`,
		`CREATE TABLE IF NOT EXISTS empresa_backups_restauraciones (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			backup_id INTEGER NOT NULL,
			codigo_backup TEXT,
			tablas_restauradas INTEGER DEFAULT 0,
			registros_restaurados INTEGER DEFAULT 0,
			tablas_omitidas_json TEXT DEFAULT '[]',
			resultado TEXT DEFAULT 'ok',
			detalle_json TEXT DEFAULT '{}',
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_backups_restauraciones_empresa_fecha ON empresa_backups_restauraciones(empresa_id, fecha_creacion DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_backups_restauraciones_backup ON empresa_backups_restauraciones(empresa_id, backup_id, id DESC);`,
	}

	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_backups", "descripcion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_backups", "include_tables_json", "TEXT DEFAULT '[]'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_backups", "exclude_tables_json", "TEXT DEFAULT '[]'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_backups", "total_tablas", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_backups", "total_registros", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_backups", "tamano_bytes", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_backups", "hash_contenido", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_backups", "metadata_json", "TEXT DEFAULT '{}'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_backups", "restaurado_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_backups", "restaurado_por", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_backups_restauraciones", "tablas_omitidas_json", "TEXT DEFAULT '[]'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_backups_restauraciones", "detalle_json", "TEXT DEFAULT '{}'"); err != nil {
		return err
	}

	return nil
}

// BuildEmpresaBackupPayload construye el snapshot JSON de la empresa para tablas con empresa_id.
func BuildEmpresaBackupPayload(dbConn *sql.DB, empresaID int64, options EmpresaBackupBuildOptions) (*EmpresaBackupPayload, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, errors.New("empresa_id invalido")
	}

	tables, err := empresaBackupListCandidateTables(dbConn, options.IncludeTables, options.ExcludeTables)
	if err != nil {
		return nil, err
	}
	if len(tables) == 0 {
		return nil, fmt.Errorf("no se encontraron tablas elegibles para backup")
	}

	payload := &EmpresaBackupPayload{
		Version:     empresaBackupsSchemaVersion,
		Scope:       "empresa",
		EmpresaID:   empresaID,
		CreatedAt:   time.Now().In(time.Local).Format(time.RFC3339),
		CreatedBy:   strings.TrimSpace(options.CreatedBy),
		TotalTables: 0,
		TotalRows:   0,
		Tables:      make([]EmpresaBackupTablePayload, 0, len(tables)),
	}

	for _, table := range tables {
		snapshot, snapshotErr := empresaBackupFetchTableSnapshot(dbConn, table, empresaID)
		if snapshotErr != nil {
			continue
		}
		payload.Tables = append(payload.Tables, snapshot)
		payload.TotalTables++
		payload.TotalRows += int64(len(snapshot.Rows))
	}

	if payload.TotalTables == 0 {
		return nil, fmt.Errorf("no se pudo construir el snapshot empresarial")
	}

	return payload, nil
}

// CreateEmpresaBackupSnapshot genera y persiste un backup empresarial completo.
func CreateEmpresaBackupSnapshot(dbConn *sql.DB, empresaID int64, nombre, descripcion, usuario string, options EmpresaBackupBuildOptions) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return 0, errors.New("empresa_id invalido")
	}

	payload, err := BuildEmpresaBackupPayload(dbConn, empresaID, EmpresaBackupBuildOptions{
		IncludeTables: options.IncludeTables,
		ExcludeTables: options.ExcludeTables,
		CreatedBy:     firstNonBlankString(strings.TrimSpace(options.CreatedBy), strings.TrimSpace(usuario)),
	})
	if err != nil {
		return 0, err
	}

	rawSnapshot, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	cleanName := strings.TrimSpace(nombre)
	if cleanName == "" {
		cleanName = fmt.Sprintf("Backup empresa %d %s", empresaID, time.Now().In(time.Local).Format("2006-01-02 15:04:05"))
	}
	codigo := empresaBackupGenerateCode(empresaID)
	includeTables := empresaBackupNormalizeTables(options.IncludeTables)
	excludeTables := empresaBackupNormalizeTables(options.ExcludeTables)
	metadata := map[string]string{
		"scope":   "empresa",
		"version": empresaBackupsSchemaVersion,
	}
	if creator := strings.TrimSpace(payload.CreatedBy); creator != "" {
		metadata["created_by"] = creator
	}

	_, err = dbConn.Exec(`
		INSERT INTO empresa_backups (
			empresa_id,
			codigo,
			nombre,
			descripcion,
			version_schema,
			alcance,
			tipo_backup,
			include_tables_json,
			exclude_tables_json,
			total_tablas,
			total_registros,
			tamano_bytes,
			hash_contenido,
			snapshot_json,
			metadata_json,
			usuario_creador,
			estado,
			observaciones
		) VALUES (?, ?, ?, ?, ?, 'empresa', 'full', ?, ?, ?, ?, ?, ?, ?, ?, ?, 'activo', ?)
	`,
		empresaID,
		codigo,
		cleanName,
		strings.TrimSpace(descripcion),
		empresaBackupsSchemaVersion,
		empresaBackupEncodeTablesJSON(includeTables),
		empresaBackupEncodeTablesJSON(excludeTables),
		payload.TotalTables,
		payload.TotalRows,
		int64(len(rawSnapshot)),
		empresaBackupHash(rawSnapshot),
		string(rawSnapshot),
		empresaBackupEncodeMetadataJSON(metadata),
		strings.TrimSpace(firstNonBlankString(usuario, payload.CreatedBy, "sistema")),
		strings.TrimSpace(descripcion),
	)
	if err != nil {
		return 0, err
	}

	row := dbConn.QueryRow(`SELECT id FROM empresa_backups WHERE empresa_id = ? AND codigo = ? LIMIT 1`, empresaID, codigo)
	var id int64
	if scanErr := row.Scan(&id); scanErr != nil {
		return 0, scanErr
	}
	return id, nil
}

func scanEmpresaBackupRow(scanner interface {
	Scan(dest ...interface{}) error
}, includeSnapshot bool) (*EmpresaBackup, error) {
	var (
		row               EmpresaBackup
		includeTablesJSON string
		excludeTablesJSON string
		metadataJSON      string
		snapshotJSON      string
	)

	if includeSnapshot {
		if err := scanner.Scan(
			&row.ID,
			&row.EmpresaID,
			&row.Codigo,
			&row.Nombre,
			&row.Descripcion,
			&row.VersionSchema,
			&row.Alcance,
			&row.TipoBackup,
			&includeTablesJSON,
			&excludeTablesJSON,
			&row.TotalTablas,
			&row.TotalRegistros,
			&row.TamanoBytes,
			&row.HashContenido,
			&snapshotJSON,
			&metadataJSON,
			&row.RestauradoEn,
			&row.RestauradoPor,
			&row.FechaCreacion,
			&row.FechaActualizacion,
			&row.UsuarioCreador,
			&row.Estado,
			&row.Observaciones,
		); err != nil {
			return nil, err
		}
		row.SnapshotJSON = strings.TrimSpace(snapshotJSON)
	} else {
		if err := scanner.Scan(
			&row.ID,
			&row.EmpresaID,
			&row.Codigo,
			&row.Nombre,
			&row.Descripcion,
			&row.VersionSchema,
			&row.Alcance,
			&row.TipoBackup,
			&includeTablesJSON,
			&excludeTablesJSON,
			&row.TotalTablas,
			&row.TotalRegistros,
			&row.TamanoBytes,
			&row.HashContenido,
			&metadataJSON,
			&row.RestauradoEn,
			&row.RestauradoPor,
			&row.FechaCreacion,
			&row.FechaActualizacion,
			&row.UsuarioCreador,
			&row.Estado,
			&row.Observaciones,
		); err != nil {
			return nil, err
		}
	}

	row.IncludeTables = empresaBackupDecodeTablesJSON(includeTablesJSON)
	row.ExcludeTables = empresaBackupDecodeTablesJSON(excludeTablesJSON)
	row.Metadata = empresaBackupDecodeMetadataJSON(metadataJSON)
	if row.VersionSchema == "" {
		row.VersionSchema = empresaBackupsSchemaVersion
	}
	if row.Alcance == "" {
		row.Alcance = "empresa"
	}
	if row.TipoBackup == "" {
		row.TipoBackup = "full"
	}
	return &row, nil
}

// ListEmpresaBackups lista snapshots por empresa con paginacion y busqueda.
func ListEmpresaBackups(dbConn *sql.DB, empresaID int64, filter EmpresaBackupFilter) ([]EmpresaBackup, int64, error) {
	if dbConn == nil {
		return nil, 0, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, 0, errors.New("empresa_id invalido")
	}

	limit, offset := empresaBackupNormalizeLimitOffset(filter.Limit, filter.Offset)
	where := " WHERE empresa_id = ?"
	args := []interface{}{empresaID}

	if !filter.IncludeInactive {
		where += " AND LOWER(COALESCE(estado, 'activo')) = 'activo'"
	}

	q := strings.TrimSpace(filter.Q)
	if q != "" {
		pattern := empresaBackupLikePattern(q)
		where += " AND (LOWER(COALESCE(codigo, '')) LIKE LOWER(?) ESCAPE '!' OR LOWER(COALESCE(nombre, '')) LIKE LOWER(?) ESCAPE '!' OR LOWER(COALESCE(descripcion, '')) LIKE LOWER(?) ESCAPE '!')"
		args = append(args, pattern, pattern, pattern)
	}

	countQuery := "SELECT COUNT(1) FROM empresa_backups" + where
	var total int64
	if err := dbConn.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT
		id,
		empresa_id,
		COALESCE(codigo, ''),
		COALESCE(nombre, ''),
		COALESCE(descripcion, ''),
		COALESCE(version_schema, ''),
		COALESCE(alcance, ''),
		COALESCE(tipo_backup, ''),
		COALESCE(include_tables_json, '[]'),
		COALESCE(exclude_tables_json, '[]'),
		COALESCE(total_tablas, 0),
		COALESCE(total_registros, 0),
		COALESCE(tamano_bytes, 0),
		COALESCE(hash_contenido, ''),
		COALESCE(metadata_json, '{}'),
		COALESCE(restaurado_en, ''),
		COALESCE(restaurado_por, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_backups` + where + ` ORDER BY id DESC LIMIT ? OFFSET ?`

	args = append(args, limit, offset)
	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]EmpresaBackup, 0)
	for rows.Next() {
		row, rowErr := scanEmpresaBackupRow(rows, false)
		if rowErr != nil {
			return nil, 0, rowErr
		}
		out = append(out, *row)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return out, total, nil
}

// GetEmpresaBackupByID obtiene un backup puntual por empresa e id.
func GetEmpresaBackupByID(dbConn *sql.DB, empresaID, backupID int64, includeSnapshot bool) (*EmpresaBackup, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 || backupID <= 0 {
		return nil, errors.New("empresa_id o id invalido")
	}

	query := `SELECT
		id,
		empresa_id,
		COALESCE(codigo, ''),
		COALESCE(nombre, ''),
		COALESCE(descripcion, ''),
		COALESCE(version_schema, ''),
		COALESCE(alcance, ''),
		COALESCE(tipo_backup, ''),
		COALESCE(include_tables_json, '[]'),
		COALESCE(exclude_tables_json, '[]'),
		COALESCE(total_tablas, 0),
		COALESCE(total_registros, 0),
		COALESCE(tamano_bytes, 0),
		COALESCE(hash_contenido, ''),`
	if includeSnapshot {
		query += `
		COALESCE(snapshot_json, ''),`
	}
	query += `
		COALESCE(metadata_json, '{}'),
		COALESCE(restaurado_en, ''),
		COALESCE(restaurado_por, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_backups
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`

	row := dbConn.QueryRow(query, empresaID, backupID)
	result, err := scanEmpresaBackupRow(row, includeSnapshot)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetEmpresaBackupPayloadByID devuelve metadata + payload JSON parseado de un backup.
func GetEmpresaBackupPayloadByID(dbConn *sql.DB, empresaID, backupID int64) (*EmpresaBackup, *EmpresaBackupPayload, error) {
	backup, err := GetEmpresaBackupByID(dbConn, empresaID, backupID, true)
	if err != nil {
		return nil, nil, err
	}
	raw := strings.TrimSpace(backup.SnapshotJSON)
	if raw == "" {
		return nil, nil, fmt.Errorf("snapshot_json vacio")
	}

	payload := &EmpresaBackupPayload{}
	if err := json.Unmarshal([]byte(raw), payload); err != nil {
		return nil, nil, err
	}
	if payload.EmpresaID <= 0 {
		payload.EmpresaID = empresaID
	}
	if strings.TrimSpace(payload.Version) == "" {
		payload.Version = empresaBackupsSchemaVersion
	}
	if strings.TrimSpace(payload.Scope) == "" {
		payload.Scope = "empresa"
	}
	return backup, payload, nil
}

func empresaBackupPrepareInsertColumns(snapshot EmpresaBackupTablePayload, row map[string]interface{}, currentColumns []string, empresaID int64) ([]string, []interface{}) {
	currentSet := map[string]bool{}
	for _, col := range currentColumns {
		clean := strings.TrimSpace(strings.ToLower(col))
		if clean == "" || !isSafeSQLIdentifier(clean) {
			continue
		}
		currentSet[clean] = true
	}

	ordered := make([]string, 0)
	if len(snapshot.Columns) > 0 {
		for _, col := range snapshot.Columns {
			clean := strings.TrimSpace(strings.ToLower(col))
			if clean == "" || !isSafeSQLIdentifier(clean) {
				continue
			}
			ordered = append(ordered, clean)
		}
	} else {
		for key := range row {
			clean := strings.TrimSpace(strings.ToLower(key))
			if clean == "" || !isSafeSQLIdentifier(clean) {
				continue
			}
			ordered = append(ordered, clean)
		}
		sort.Strings(ordered)
	}

	if len(ordered) == 0 {
		return nil, nil
	}

	seen := map[string]bool{}
	columns := make([]string, 0, len(ordered)+1)
	values := make([]interface{}, 0, len(ordered)+1)
	for _, col := range ordered {
		if seen[col] || !currentSet[col] {
			continue
		}
		seen[col] = true
		if col == "empresa_id" {
			columns = append(columns, col)
			values = append(values, empresaID)
			continue
		}
		columns = append(columns, col)
		values = append(values, normalizeGenericValue(row[col]))
	}

	if !seen["empresa_id"] && currentSet["empresa_id"] {
		columns = append(columns, "empresa_id")
		values = append(values, empresaID)
	}

	if len(columns) == 0 {
		return nil, nil
	}
	return columns, values
}

func empresaBackupInsertRow(tx *sql.Tx, table string, columns []string, values []interface{}) error {
	if len(columns) == 0 || len(columns) != len(values) {
		return errors.New("columns/values invalidos")
	}
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	query := "INSERT INTO " + table + " (" + strings.Join(columns, ", ") + ") VALUES (" + strings.Join(placeholders, ", ") + ")"
	_, err := tx.Exec(query, values...)
	return err
}

// RestoreEmpresaBackupByID aplica un snapshot previo y registra trazabilidad de restauracion.
func RestoreEmpresaBackupByID(dbConn *sql.DB, empresaID, backupID int64, usuario, observaciones string) (*EmpresaBackupRestoreResult, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 || backupID <= 0 {
		return nil, errors.New("empresa_id o backup_id invalido")
	}

	backup, payload, err := GetEmpresaBackupPayloadByID(dbConn, empresaID, backupID)
	if err != nil {
		return nil, err
	}
	if payload.EmpresaID > 0 && payload.EmpresaID != empresaID {
		return nil, fmt.Errorf("empresa_id del backup no coincide")
	}

	tx, err := dbConn.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	restoredTables := 0
	restoredRows := int64(0)
	skippedTables := make([]string, 0)

	for _, tableSnapshot := range payload.Tables {
		table := strings.ToLower(strings.TrimSpace(tableSnapshot.Table))
		if table == "" || !isSafeSQLIdentifier(table) || empresaBackupExcludedInternalTable(table) {
			skippedTables = append(skippedTables, table)
			continue
		}

		currentColumns, colErr := empresaBackupGetTableColumns(tx, table)
		if colErr != nil || !empresaBackupHasColumn(currentColumns, "empresa_id") {
			skippedTables = append(skippedTables, table)
			continue
		}

		if _, delErr := tx.Exec("DELETE FROM "+table+" WHERE empresa_id = ?", empresaID); delErr != nil {
			return nil, delErr
		}

		for _, item := range tableSnapshot.Rows {
			columns, values := empresaBackupPrepareInsertColumns(tableSnapshot, item, currentColumns, empresaID)
			if len(columns) == 0 {
				continue
			}
			if insErr := empresaBackupInsertRow(tx, table, columns, values); insErr != nil {
				return nil, insErr
			}
			restoredRows++
		}
		restoredTables++
	}

	if restoredTables == 0 {
		return nil, fmt.Errorf("no se restauraron tablas del backup")
	}

	now := time.Now().In(time.Local).Format("2006-01-02 15:04:05")
	executor := strings.TrimSpace(firstNonBlankString(usuario, "sistema"))
	obs := strings.TrimSpace(observaciones)
	if obs == "" {
		obs = fmt.Sprintf("restaurado desde backup %s", strings.TrimSpace(backup.Codigo))
	}

	detailJSON := empresaBackupEncodeMetadataJSON(map[string]string{
		"resultado":             "ok",
		"version":               strings.TrimSpace(payload.Version),
		"tablas_restauradas":    fmt.Sprintf("%d", restoredTables),
		"registros_restaurados": fmt.Sprintf("%d", restoredRows),
	})

	_, err = tx.Exec(`
		INSERT INTO empresa_backups_restauraciones (
			empresa_id,
			backup_id,
			codigo_backup,
			tablas_restauradas,
			registros_restaurados,
			tablas_omitidas_json,
			resultado,
			detalle_json,
			usuario_creador,
			estado,
			observaciones
		) VALUES (?, ?, ?, ?, ?, ?, 'ok', ?, ?, 'activo', ?)
	`,
		empresaID,
		backupID,
		strings.TrimSpace(backup.Codigo),
		restoredTables,
		restoredRows,
		empresaBackupEncodeTablesJSON(skippedTables),
		detailJSON,
		executor,
		obs,
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`
		UPDATE empresa_backups
		SET restaurado_en = ?,
			restaurado_por = ?,
			fecha_actualizacion = datetime('now','localtime')
		WHERE empresa_id = ? AND id = ?
	`, now, executor, empresaID, backupID)
	if err != nil {
		return nil, err
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return nil, commitErr
	}
	tx = nil

	return &EmpresaBackupRestoreResult{
		EmpresaID:            empresaID,
		BackupID:             backupID,
		CodigoBackup:         strings.TrimSpace(backup.Codigo),
		TablasRestauradas:    restoredTables,
		RegistrosRestaurados: restoredRows,
		TablasOmitidas:       skippedTables,
		EjecutadoEn:          now,
		EjecutadoPor:         executor,
	}, nil
}

// PurgeEmpresaDataByDateCorte elimina informacion de la empresa desde el origen hasta la fecha de corte (inclusive).
// Solo depura tablas que tengan `empresa_id` y alguna columna de fecha soportada.
func PurgeEmpresaDataByDateCorte(dbConn *sql.DB, empresaID int64, fechaCorte string, includeTables, excludeTables []string) (*EmpresaBackupPurgeResult, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, errors.New("empresa_id invalido")
	}
	fechaCorte = strings.TrimSpace(fechaCorte)
	if fechaCorte == "" {
		return nil, errors.New("fecha_corte es obligatoria")
	}

	tables, err := empresaBackupListCandidateTables(dbConn, includeTables, excludeTables)
	if err != nil {
		return nil, err
	}
	if len(tables) == 0 {
		return &EmpresaBackupPurgeResult{
			EmpresaID:       empresaID,
			FechaCorte:      fechaCorte,
			TablasEvaluadas: 0,
			TablasDepuradas: 0,
		}, nil
	}

	tx, err := dbConn.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	result := &EmpresaBackupPurgeResult{
		EmpresaID:       empresaID,
		FechaCorte:      fechaCorte,
		TablasEvaluadas: len(tables),
		TablasDepuradas: 0,
		TablasSinFecha:  make([]string, 0),
		Detalle:         make([]EmpresaBackupPurgeTableResult, 0, len(tables)),
	}

	for _, table := range tables {
		columns, colErr := empresaBackupGetTableColumns(tx, table)
		if colErr != nil {
			result.TablasSinFecha = append(result.TablasSinFecha, table)
			continue
		}
		dateColumn := empresaBackupResolveDateColumn(columns)
		if dateColumn == "" {
			result.TablasSinFecha = append(result.TablasSinFecha, table)
			continue
		}

		query := "DELETE FROM " + table + " WHERE empresa_id = ? AND COALESCE(" + dateColumn + ", '') <> '' AND datetime(" + dateColumn + ") <= datetime(?)"
		execResult, execErr := tx.Exec(query, empresaID, fechaCorte)
		if execErr != nil {
			return nil, execErr
		}
		affected, affErr := execResult.RowsAffected()
		if affErr != nil {
			return nil, affErr
		}
		result.RegistrosEliminados += affected
		result.TablasDepuradas++
		result.Detalle = append(result.Detalle, EmpresaBackupPurgeTableResult{
			Table:      table,
			DateColumn: dateColumn,
			Deleted:    affected,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil

	return result, nil
}

// SetEmpresaBackupEstadoByID aplica estado activo/inactivo al backup.
func SetEmpresaBackupEstadoByID(dbConn *sql.DB, empresaID, backupID int64, estado string) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if empresaID <= 0 || backupID <= 0 {
		return errors.New("empresa_id o backup_id invalido")
	}
	_, err := dbConn.Exec(`
		UPDATE empresa_backups
		SET estado = ?, fecha_actualizacion = datetime('now','localtime')
		WHERE empresa_id = ? AND id = ?
	`, normalizeEmpresaBackupEstado(estado), empresaID, backupID)
	return err
}
