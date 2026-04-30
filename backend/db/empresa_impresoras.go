package db

import (
	"database/sql"
	"fmt"
	"strings"
)

const (
	defaultEmpresaImpresoraFormato      = "pos"
	defaultEmpresaImpresoraTipoConexion = "red"
)

// EmpresaImpresora representa una impresora registrada por empresa.
type EmpresaImpresora struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Codigo             string `json:"codigo"`
	Nombre             string `json:"nombre"`
	TipoConexion       string `json:"tipo_conexion,omitempty"`
	Direccion          string `json:"direccion,omitempty"`
	AreaOperativa      string `json:"area_operativa,omitempty"`
	FormatoImpresion   string `json:"formato_impresion,omitempty"`
	EsPredeterminada   bool   `json:"es_predeterminada"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// EmpresaImpresoraFuncionalidad define la impresora asignada por funcionalidad operativa.
type EmpresaImpresoraFuncionalidad struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Funcionalidad      string `json:"funcionalidad"`
	ImpresoraID        int64  `json:"impresora_id"`
	ImpresoraNombre    string `json:"impresora_nombre,omitempty"`
	ImpresoraCodigo    string `json:"impresora_codigo,omitempty"`
	Prioridad          int64  `json:"prioridad,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// EmpresaImpresoraProducto define la impresora asignada por producto.
type EmpresaImpresoraProducto struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	ProductoID         int64  `json:"producto_id"`
	ProductoNombre     string `json:"producto_nombre,omitempty"`
	ImpresoraID        int64  `json:"impresora_id"`
	ImpresoraNombre    string `json:"impresora_nombre,omitempty"`
	ImpresoraCodigo    string `json:"impresora_codigo,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// EmpresaImpresoraResolucion representa la impresora seleccionada para una ejecución operativa.
type EmpresaImpresoraResolucion struct {
	EmpresaID     int64            `json:"empresa_id"`
	Funcionalidad string           `json:"funcionalidad,omitempty"`
	ProductoID    int64            `json:"producto_id,omitempty"`
	Fuente        string           `json:"fuente"`
	Impresora     EmpresaImpresora `json:"impresora"`
}

// EnsureEmpresaImpresorasSchema crea/migra tablas del módulo de impresoras por empresa.
func EnsureEmpresaImpresorasSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_impresoras (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			tipo_conexion TEXT DEFAULT 'red',
			direccion TEXT,
			area_operativa TEXT,
			formato_impresion TEXT DEFAULT 'pos',
			es_predeterminada INTEGER DEFAULT 0,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_impresoras_empresa_codigo ON empresa_impresoras(empresa_id, codigo);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_impresoras_empresa_estado ON empresa_impresoras(empresa_id, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_impresoras_empresa_default ON empresa_impresoras(empresa_id, es_predeterminada);`,
		`CREATE TABLE IF NOT EXISTS empresa_impresoras_funcionalidades (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			funcionalidad TEXT NOT NULL,
			impresora_id INTEGER NOT NULL,
			prioridad INTEGER DEFAULT 100,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_impresoras_funcionalidad ON empresa_impresoras_funcionalidades(empresa_id, funcionalidad);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_impresoras_funcionalidades_printer ON empresa_impresoras_funcionalidades(empresa_id, impresora_id);`,
		`CREATE TABLE IF NOT EXISTS empresa_impresoras_productos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			producto_id INTEGER NOT NULL,
			impresora_id INTEGER NOT NULL,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_impresoras_producto ON empresa_impresoras_productos(empresa_id, producto_id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_impresoras_productos_printer ON empresa_impresoras_productos(empresa_id, impresora_id);`,
	}

	for _, stmt := range stmts {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}

	// Tabla principal: columnas evolutivas
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras", "empresa_id", "INTEGER NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras", "codigo", "TEXT NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras", "nombre", "TEXT NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras", "tipo_conexion", "TEXT DEFAULT 'red'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras", "direccion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras", "area_operativa", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras", "formato_impresion", "TEXT DEFAULT 'pos'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras", "es_predeterminada", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras", "fecha_creacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras", "fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras", "observaciones", "TEXT"); err != nil {
		return err
	}

	// Tabla por funcionalidad
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_funcionalidades", "empresa_id", "INTEGER NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_funcionalidades", "funcionalidad", "TEXT NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_funcionalidades", "impresora_id", "INTEGER NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_funcionalidades", "prioridad", "INTEGER DEFAULT 100"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_funcionalidades", "fecha_creacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_funcionalidades", "fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_funcionalidades", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_funcionalidades", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_funcionalidades", "observaciones", "TEXT"); err != nil {
		return err
	}

	// Tabla por producto
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_productos", "empresa_id", "INTEGER NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_productos", "producto_id", "INTEGER NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_productos", "impresora_id", "INTEGER NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_productos", "fecha_creacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_productos", "fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_productos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_productos", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_productos", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

type empresaImpresoraScanner interface {
	Scan(dest ...interface{}) error
}

func empresaImpresoraDefaultSelectExpr(alias string) string {
	prefix := ""
	if strings.TrimSpace(alias) != "" {
		prefix = strings.TrimSpace(alias) + "."
	}
	return "CASE WHEN lower(COALESCE(CAST(" + prefix + "es_predeterminada AS TEXT), '')) IN ('1', 'true', 't', 'yes', 'si') THEN 1 ELSE 0 END"
}

func empresaImpresoraDefaultWhereExpr(alias string) string {
	return empresaImpresoraDefaultSelectExpr(alias) + " = 1"
}

func scanEmpresaImpresora(row empresaImpresoraScanner) (*EmpresaImpresora, error) {
	item := EmpresaImpresora{}
	var esPredeterminadaInt int
	if err := row.Scan(
		&item.ID,
		&item.EmpresaID,
		&item.Codigo,
		&item.Nombre,
		&item.TipoConexion,
		&item.Direccion,
		&item.AreaOperativa,
		&item.FormatoImpresion,
		&esPredeterminadaInt,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	); err != nil {
		return nil, err
	}
	item.EsPredeterminada = esPredeterminadaInt == 1
	item.FormatoImpresion = normalizeEmpresaImpresoraFormato(item.FormatoImpresion)
	item.TipoConexion = normalizeEmpresaImpresoraTipoConexion(item.TipoConexion)
	item.Estado = normalizeEmpresaImpresoraEstado(item.Estado)
	return &item, nil
}

func normalizeEmpresaImpresoraEstado(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "inactivo" || value == "desactivado" || value == "off" {
		return "inactivo"
	}
	return "activo"
}

func normalizeEmpresaImpresoraFormato(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "carta" {
		return "carta"
	}
	return defaultEmpresaImpresoraFormato
}

func normalizeEmpresaImpresoraTipoConexion(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "usb", "red", "windows", "bluetooth":
		return value
	default:
		return defaultEmpresaImpresoraTipoConexion
	}
}

func normalizeEmpresaImpresoraFuncionalidad(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return "general"
	}
	var b strings.Builder
	b.Grow(len(value))
	for i := 0; i < len(value); i++ {
		ch := value[i]
		isLetter := ch >= 'a' && ch <= 'z'
		isDigit := ch >= '0' && ch <= '9'
		if isLetter || isDigit {
			b.WriteByte(ch)
			continue
		}
		if ch == '_' || ch == '-' || ch == ' ' {
			b.WriteByte('_')
		}
	}
	norm := strings.Trim(b.String(), "_")
	if norm == "" {
		return "general"
	}
	return norm
}

func normalizeEmpresaImpresoraCodigo(raw, nombre string) string {
	input := strings.TrimSpace(raw)
	if input == "" {
		input = strings.TrimSpace(nombre)
	}
	if input == "" {
		return "IMPRESORA"
	}
	input = strings.ToUpper(input)
	var b strings.Builder
	b.Grow(len(input))
	lastUnderscore := false
	for i := 0; i < len(input); i++ {
		ch := input[i]
		isLetter := ch >= 'A' && ch <= 'Z'
		isDigit := ch >= '0' && ch <= '9'
		if isLetter || isDigit {
			b.WriteByte(ch)
			lastUnderscore = false
			continue
		}
		if ch == '_' || ch == '-' || ch == ' ' {
			if !lastUnderscore {
				b.WriteByte('_')
				lastUnderscore = true
			}
		}
	}
	code := strings.Trim(b.String(), "_")
	if code == "" {
		return "IMPRESORA"
	}
	if len(code) > 60 {
		return code[:60]
	}
	return code
}

func normalizeEmpresaImpresoraPrioridad(raw int64) int64 {
	if raw <= 0 {
		return 100
	}
	if raw > 99999 {
		return 99999
	}
	return raw
}

// ListEmpresaImpresorasByEmpresa lista impresoras por empresa.
func ListEmpresaImpresorasByEmpresa(dbConn *sql.DB, empresaID int64, includeInactive bool) ([]EmpresaImpresora, error) {
	query := `SELECT
		id,
		empresa_id,
		COALESCE(codigo, ''),
		COALESCE(nombre, ''),
		COALESCE(tipo_conexion, 'red'),
		COALESCE(direccion, ''),
		COALESCE(area_operativa, ''),
		COALESCE(formato_impresion, 'pos'),
		` + empresaImpresoraDefaultSelectExpr("") + `,
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_impresoras
	WHERE empresa_id = ?`
	if !includeInactive {
		query += ` AND COALESCE(NULLIF(TRIM(estado), ''), 'activo') = 'activo'`
	}
	query += ` ORDER BY ` + empresaImpresoraDefaultSelectExpr("") + ` DESC, nombre ASC, id ASC`

	rows, err := querySQLCompat(dbConn, query, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaImpresora, 0)
	for rows.Next() {
		item, errScan := scanEmpresaImpresora(rows)
		if errScan != nil {
			return nil, errScan
		}
		out = append(out, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// GetEmpresaImpresoraByID obtiene una impresora puntual por empresa e id.
func GetEmpresaImpresoraByID(dbConn *sql.DB, empresaID, impresoraID int64) (*EmpresaImpresora, error) {
	row := queryRowSQLCompat(dbConn, `SELECT
		id,
		empresa_id,
		COALESCE(codigo, ''),
		COALESCE(nombre, ''),
		COALESCE(tipo_conexion, 'red'),
		COALESCE(direccion, ''),
		COALESCE(area_operativa, ''),
		COALESCE(formato_impresion, 'pos'),
		`+empresaImpresoraDefaultSelectExpr("")+`,
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_impresoras
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, impresoraID)
	return scanEmpresaImpresora(row)
}

func ensureEmpresaImpresoraDefaultConsistency(dbConn *sql.DB, empresaID int64) error {
	var defaultCount int64
	if err := queryRowSQLCompat(dbConn, `SELECT COUNT(1) FROM empresa_impresoras WHERE empresa_id = ? AND COALESCE(NULLIF(TRIM(estado), ''), 'activo') = 'activo' AND `+empresaImpresoraDefaultWhereExpr(""), empresaID).Scan(&defaultCount); err != nil {
		return err
	}

	if defaultCount > 1 {
		var keepID int64
		if err := queryRowSQLCompat(dbConn, `SELECT id FROM empresa_impresoras WHERE empresa_id = ? AND COALESCE(NULLIF(TRIM(estado), ''), 'activo') = 'activo' AND `+empresaImpresoraDefaultWhereExpr("")+` ORDER BY id ASC LIMIT 1`, empresaID).Scan(&keepID); err != nil {
			return err
		}
		if _, err := execSQLCompat(dbConn, `UPDATE empresa_impresoras SET es_predeterminada = CASE WHEN id = ? THEN 1 ELSE 0 END, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ? AND COALESCE(NULLIF(TRIM(estado), ''), 'activo') = 'activo'`, keepID, empresaID); err != nil {
			return err
		}
		return nil
	}

	if defaultCount == 0 {
		var firstActiveID int64
		err := queryRowSQLCompat(dbConn, `SELECT id FROM empresa_impresoras WHERE empresa_id = ? AND COALESCE(NULLIF(TRIM(estado), ''), 'activo') = 'activo' ORDER BY id ASC LIMIT 1`, empresaID).Scan(&firstActiveID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil
			}
			return err
		}
		if _, err := execSQLCompat(dbConn, `UPDATE empresa_impresoras SET es_predeterminada = CASE WHEN id = ? THEN 1 ELSE 0 END, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ?`, firstActiveID, empresaID); err != nil {
			return err
		}
	}
	return nil
}

// UpsertEmpresaImpresora crea o actualiza una impresora por empresa.
func UpsertEmpresaImpresora(dbConn *sql.DB, payload EmpresaImpresora) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id requerido")
	}

	payload.Nombre = strings.TrimSpace(payload.Nombre)
	if payload.Nombre == "" {
		return 0, fmt.Errorf("nombre de impresora requerido")
	}
	payload.Codigo = normalizeEmpresaImpresoraCodigo(payload.Codigo, payload.Nombre)
	payload.TipoConexion = normalizeEmpresaImpresoraTipoConexion(payload.TipoConexion)
	payload.FormatoImpresion = normalizeEmpresaImpresoraFormato(payload.FormatoImpresion)
	payload.Estado = normalizeEmpresaImpresoraEstado(payload.Estado)
	payload.UsuarioCreador = strings.TrimSpace(payload.UsuarioCreador)
	if payload.UsuarioCreador == "" {
		payload.UsuarioCreador = "sistema"
	}
	payload.Direccion = strings.TrimSpace(payload.Direccion)
	payload.AreaOperativa = strings.TrimSpace(payload.AreaOperativa)
	payload.Observaciones = strings.TrimSpace(payload.Observaciones)

	esPredeterminadaInt := 0
	if payload.EsPredeterminada {
		esPredeterminadaInt = 1
	}

	if payload.ID > 0 {
		if _, err := execSQLCompat(dbConn, `UPDATE empresa_impresoras SET
			codigo = ?,
			nombre = ?,
			tipo_conexion = ?,
			direccion = ?,
			area_operativa = ?,
			formato_impresion = ?,
			es_predeterminada = ?,
			fecha_actualizacion = datetime('now','localtime'),
			usuario_creador = ?,
			estado = ?,
			observaciones = ?
		WHERE empresa_id = ? AND id = ?`,
			payload.Codigo,
			payload.Nombre,
			payload.TipoConexion,
			payload.Direccion,
			payload.AreaOperativa,
			payload.FormatoImpresion,
			esPredeterminadaInt,
			payload.UsuarioCreador,
			payload.Estado,
			payload.Observaciones,
			payload.EmpresaID,
			payload.ID,
		); err != nil {
			return 0, err
		}
	} else {
		var duplicateID int64
		err := queryRowSQLCompat(dbConn, `SELECT id FROM empresa_impresoras WHERE empresa_id = ? AND codigo = ? LIMIT 1`, payload.EmpresaID, payload.Codigo).Scan(&duplicateID)
		if err == nil && duplicateID > 0 {
			return 0, fmt.Errorf("ya existe una impresora con codigo %q", payload.Codigo)
		}
		if err != nil && err != sql.ErrNoRows {
			return 0, err
		}

		insertedID, errInsert := insertSQLCompat(dbConn, `INSERT INTO empresa_impresoras (
			empresa_id,
			codigo,
			nombre,
			tipo_conexion,
			direccion,
			area_operativa,
			formato_impresion,
			es_predeterminada,
			fecha_creacion,
			fecha_actualizacion,
			usuario_creador,
			estado,
			observaciones
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'), ?, ?, ?)`,
			payload.EmpresaID,
			payload.Codigo,
			payload.Nombre,
			payload.TipoConexion,
			payload.Direccion,
			payload.AreaOperativa,
			payload.FormatoImpresion,
			esPredeterminadaInt,
			payload.UsuarioCreador,
			payload.Estado,
			payload.Observaciones,
		)
		if errInsert != nil {
			return 0, errInsert
		}
		payload.ID = insertedID
	}

	if payload.EsPredeterminada {
		if _, err := execSQLCompat(dbConn, `UPDATE empresa_impresoras SET es_predeterminada = CASE WHEN id = ? THEN 1 ELSE 0 END, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ?`, payload.ID, payload.EmpresaID); err != nil {
			return 0, err
		}
	}
	if err := ensureEmpresaImpresoraDefaultConsistency(dbConn, payload.EmpresaID); err != nil {
		return 0, err
	}
	return payload.ID, nil
}

// SetEmpresaImpresoraPredeterminada marca una impresora activa como predeterminada.
func SetEmpresaImpresoraPredeterminada(dbConn *sql.DB, empresaID, impresoraID int64, usuario string) error {
	if empresaID <= 0 || impresoraID <= 0 {
		return fmt.Errorf("empresa_id e impresora_id requeridos")
	}
	var estado string
	if err := queryRowSQLCompat(dbConn, `SELECT COALESCE(estado, 'activo') FROM empresa_impresoras WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, impresoraID).Scan(&estado); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("impresora no encontrada")
		}
		return err
	}
	if normalizeEmpresaImpresoraEstado(estado) != "activo" {
		return fmt.Errorf("solo se puede seleccionar como predeterminada una impresora activa")
	}
	if strings.TrimSpace(usuario) == "" {
		usuario = "sistema"
	}
	if _, err := execSQLCompat(dbConn, `UPDATE empresa_impresoras SET es_predeterminada = CASE WHEN id = ? THEN 1 ELSE 0 END, fecha_actualizacion = datetime('now','localtime'), usuario_creador = ? WHERE empresa_id = ?`, impresoraID, strings.TrimSpace(usuario), empresaID); err != nil {
		return err
	}
	return ensureEmpresaImpresoraDefaultConsistency(dbConn, empresaID)
}

// SetEmpresaImpresoraEstado activa o desactiva una impresora.
func SetEmpresaImpresoraEstado(dbConn *sql.DB, empresaID, impresoraID int64, estado, usuario string) error {
	if empresaID <= 0 || impresoraID <= 0 {
		return fmt.Errorf("empresa_id e impresora_id requeridos")
	}
	normEstado := normalizeEmpresaImpresoraEstado(estado)
	if strings.TrimSpace(usuario) == "" {
		usuario = "sistema"
	}
	esPredeterminada := 1
	if normEstado != "activo" {
		esPredeterminada = 0
	}
	if _, err := execSQLCompat(dbConn, `UPDATE empresa_impresoras SET estado = ?, es_predeterminada = CASE WHEN ? = 1 THEN es_predeterminada ELSE 0 END, fecha_actualizacion = datetime('now','localtime'), usuario_creador = ? WHERE empresa_id = ? AND id = ?`, normEstado, esPredeterminada, strings.TrimSpace(usuario), empresaID, impresoraID); err != nil {
		return err
	}
	return ensureEmpresaImpresoraDefaultConsistency(dbConn, empresaID)
}

// ListEmpresaImpresoraFuncionalidadesByEmpresa lista asignaciones por funcionalidad.
func ListEmpresaImpresoraFuncionalidadesByEmpresa(dbConn *sql.DB, empresaID int64) ([]EmpresaImpresoraFuncionalidad, error) {
	rows, err := querySQLCompat(dbConn, `SELECT
		f.id,
		f.empresa_id,
		COALESCE(f.funcionalidad, ''),
		f.impresora_id,
		COALESCE(p.nombre, ''),
		COALESCE(p.codigo, ''),
		COALESCE(f.prioridad, 100),
		COALESCE(f.fecha_creacion, ''),
		COALESCE(f.fecha_actualizacion, ''),
		COALESCE(f.usuario_creador, ''),
		COALESCE(f.estado, 'activo'),
		COALESCE(f.observaciones, '')
	FROM empresa_impresoras_funcionalidades f
	LEFT JOIN empresa_impresoras p ON p.id = f.impresora_id AND p.empresa_id = f.empresa_id
	WHERE f.empresa_id = ?
	ORDER BY f.funcionalidad ASC, f.id ASC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaImpresoraFuncionalidad, 0)
	for rows.Next() {
		item := EmpresaImpresoraFuncionalidad{}
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.Funcionalidad,
			&item.ImpresoraID,
			&item.ImpresoraNombre,
			&item.ImpresoraCodigo,
			&item.Prioridad,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		item.Funcionalidad = normalizeEmpresaImpresoraFuncionalidad(item.Funcionalidad)
		item.Estado = normalizeEmpresaImpresoraEstado(item.Estado)
		item.Prioridad = normalizeEmpresaImpresoraPrioridad(item.Prioridad)
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func ensureEmpresaImpresoraExistsAndActive(dbConn *sql.DB, empresaID, impresoraID int64) error {
	row := queryRowSQLCompat(dbConn, `SELECT COALESCE(estado, 'activo') FROM empresa_impresoras WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, impresoraID)
	var estado string
	if err := row.Scan(&estado); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("impresora no encontrada")
		}
		return err
	}
	if normalizeEmpresaImpresoraEstado(estado) != "activo" {
		return fmt.Errorf("la impresora seleccionada está inactiva")
	}
	return nil
}

// UpsertEmpresaImpresoraFuncionalidad crea/actualiza asignación funcionalidad -> impresora.
func UpsertEmpresaImpresoraFuncionalidad(dbConn *sql.DB, payload EmpresaImpresoraFuncionalidad) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id requerido")
	}
	if payload.ImpresoraID <= 0 {
		return 0, fmt.Errorf("impresora_id requerido")
	}
	if err := ensureEmpresaImpresoraExistsAndActive(dbConn, payload.EmpresaID, payload.ImpresoraID); err != nil {
		return 0, err
	}
	payload.Funcionalidad = normalizeEmpresaImpresoraFuncionalidad(payload.Funcionalidad)
	payload.Prioridad = normalizeEmpresaImpresoraPrioridad(payload.Prioridad)
	payload.UsuarioCreador = strings.TrimSpace(payload.UsuarioCreador)
	if payload.UsuarioCreador == "" {
		payload.UsuarioCreador = "sistema"
	}
	payload.Estado = normalizeEmpresaImpresoraEstado(payload.Estado)
	payload.Observaciones = strings.TrimSpace(payload.Observaciones)

	if _, err := execSQLCompat(dbConn, `INSERT INTO empresa_impresoras_funcionalidades (
		empresa_id,
		funcionalidad,
		impresora_id,
		prioridad,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'), ?, ?, ?)
	ON CONFLICT(empresa_id, funcionalidad) DO UPDATE SET
		impresora_id = excluded.impresora_id,
		prioridad = excluded.prioridad,
		fecha_actualizacion = datetime('now','localtime'),
		usuario_creador = excluded.usuario_creador,
		estado = excluded.estado,
		observaciones = excluded.observaciones`,
		payload.EmpresaID,
		payload.Funcionalidad,
		payload.ImpresoraID,
		payload.Prioridad,
		payload.UsuarioCreador,
		payload.Estado,
		payload.Observaciones,
	); err != nil {
		return 0, err
	}

	var id int64
	if err := queryRowSQLCompat(dbConn, `SELECT id FROM empresa_impresoras_funcionalidades WHERE empresa_id = ? AND funcionalidad = ? LIMIT 1`, payload.EmpresaID, payload.Funcionalidad).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// DeleteEmpresaImpresoraFuncionalidad elimina asignación por funcionalidad.
func DeleteEmpresaImpresoraFuncionalidad(dbConn *sql.DB, empresaID int64, funcionalidad string) error {
	if empresaID <= 0 {
		return fmt.Errorf("empresa_id requerido")
	}
	funcionalidad = normalizeEmpresaImpresoraFuncionalidad(funcionalidad)
	_, err := execSQLCompat(dbConn, `DELETE FROM empresa_impresoras_funcionalidades WHERE empresa_id = ? AND funcionalidad = ?`, empresaID, funcionalidad)
	return err
}

// ListEmpresaImpresoraProductosByEmpresa lista asignaciones por producto.
func ListEmpresaImpresoraProductosByEmpresa(dbConn *sql.DB, empresaID int64) ([]EmpresaImpresoraProducto, error) {
	rows, err := querySQLCompat(dbConn, `SELECT
		a.id,
		a.empresa_id,
		a.producto_id,
		COALESCE(p.nombre, ''),
		a.impresora_id,
		COALESCE(i.nombre, ''),
		COALESCE(i.codigo, ''),
		COALESCE(a.fecha_creacion, ''),
		COALESCE(a.fecha_actualizacion, ''),
		COALESCE(a.usuario_creador, ''),
		COALESCE(a.estado, 'activo'),
		COALESCE(a.observaciones, '')
	FROM empresa_impresoras_productos a
	LEFT JOIN productos p ON p.id = a.producto_id AND p.empresa_id = a.empresa_id
	LEFT JOIN empresa_impresoras i ON i.id = a.impresora_id AND i.empresa_id = a.empresa_id
	WHERE a.empresa_id = ?
	ORDER BY p.nombre ASC, a.producto_id ASC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaImpresoraProducto, 0)
	for rows.Next() {
		item := EmpresaImpresoraProducto{}
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.ProductoID,
			&item.ProductoNombre,
			&item.ImpresoraID,
			&item.ImpresoraNombre,
			&item.ImpresoraCodigo,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		item.Estado = normalizeEmpresaImpresoraEstado(item.Estado)
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func ensureEmpresaProductoExists(dbConn *sql.DB, empresaID, productoID int64) error {
	var id int64
	if err := queryRowSQLCompat(dbConn, `SELECT id FROM productos WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, productoID).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("producto no encontrado")
		}
		return err
	}
	return nil
}

// UpsertEmpresaImpresoraProducto crea/actualiza asignación producto -> impresora.
func UpsertEmpresaImpresoraProducto(dbConn *sql.DB, payload EmpresaImpresoraProducto) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id requerido")
	}
	if payload.ProductoID <= 0 {
		return 0, fmt.Errorf("producto_id requerido")
	}
	if payload.ImpresoraID <= 0 {
		return 0, fmt.Errorf("impresora_id requerido")
	}
	if err := ensureEmpresaProductoExists(dbConn, payload.EmpresaID, payload.ProductoID); err != nil {
		return 0, err
	}
	if err := ensureEmpresaImpresoraExistsAndActive(dbConn, payload.EmpresaID, payload.ImpresoraID); err != nil {
		return 0, err
	}

	payload.UsuarioCreador = strings.TrimSpace(payload.UsuarioCreador)
	if payload.UsuarioCreador == "" {
		payload.UsuarioCreador = "sistema"
	}
	payload.Estado = normalizeEmpresaImpresoraEstado(payload.Estado)
	payload.Observaciones = strings.TrimSpace(payload.Observaciones)

	if _, err := execSQLCompat(dbConn, `INSERT INTO empresa_impresoras_productos (
		empresa_id,
		producto_id,
		impresora_id,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, datetime('now','localtime'), datetime('now','localtime'), ?, ?, ?)
	ON CONFLICT(empresa_id, producto_id) DO UPDATE SET
		impresora_id = excluded.impresora_id,
		fecha_actualizacion = datetime('now','localtime'),
		usuario_creador = excluded.usuario_creador,
		estado = excluded.estado,
		observaciones = excluded.observaciones`,
		payload.EmpresaID,
		payload.ProductoID,
		payload.ImpresoraID,
		payload.UsuarioCreador,
		payload.Estado,
		payload.Observaciones,
	); err != nil {
		return 0, err
	}

	var id int64
	if err := queryRowSQLCompat(dbConn, `SELECT id FROM empresa_impresoras_productos WHERE empresa_id = ? AND producto_id = ? LIMIT 1`, payload.EmpresaID, payload.ProductoID).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// DeleteEmpresaImpresoraProducto elimina asignación por producto.
func DeleteEmpresaImpresoraProducto(dbConn *sql.DB, empresaID, productoID int64) error {
	if empresaID <= 0 {
		return fmt.Errorf("empresa_id requerido")
	}
	if productoID <= 0 {
		return fmt.Errorf("producto_id requerido")
	}
	_, err := execSQLCompat(dbConn, `DELETE FROM empresa_impresoras_productos WHERE empresa_id = ? AND producto_id = ?`, empresaID, productoID)
	return err
}

func resolveEmpresaImpresoraByProducto(dbConn *sql.DB, empresaID, productoID int64) (*EmpresaImpresora, error) {
	row := queryRowSQLCompat(dbConn, `SELECT
		i.id,
		i.empresa_id,
		COALESCE(i.codigo, ''),
		COALESCE(i.nombre, ''),
		COALESCE(i.tipo_conexion, 'red'),
		COALESCE(i.direccion, ''),
		COALESCE(i.area_operativa, ''),
		COALESCE(i.formato_impresion, 'pos'),
		`+empresaImpresoraDefaultSelectExpr("i")+`,
		COALESCE(i.fecha_creacion, ''),
		COALESCE(i.fecha_actualizacion, ''),
		COALESCE(i.usuario_creador, ''),
		COALESCE(i.estado, 'activo'),
		COALESCE(i.observaciones, '')
	FROM empresa_impresoras_productos p
	INNER JOIN empresa_impresoras i ON i.id = p.impresora_id AND i.empresa_id = p.empresa_id
	WHERE p.empresa_id = ?
		AND p.producto_id = ?
		AND COALESCE(NULLIF(TRIM(p.estado), ''), 'activo') = 'activo'
		AND COALESCE(NULLIF(TRIM(i.estado), ''), 'activo') = 'activo'
	LIMIT 1`, empresaID, productoID)
	item, err := scanEmpresaImpresora(row)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func resolveEmpresaImpresoraByFuncionalidad(dbConn *sql.DB, empresaID int64, funcionalidad string) (*EmpresaImpresora, error) {
	row := queryRowSQLCompat(dbConn, `SELECT
		i.id,
		i.empresa_id,
		COALESCE(i.codigo, ''),
		COALESCE(i.nombre, ''),
		COALESCE(i.tipo_conexion, 'red'),
		COALESCE(i.direccion, ''),
		COALESCE(i.area_operativa, ''),
		COALESCE(i.formato_impresion, 'pos'),
		`+empresaImpresoraDefaultSelectExpr("i")+`,
		COALESCE(i.fecha_creacion, ''),
		COALESCE(i.fecha_actualizacion, ''),
		COALESCE(i.usuario_creador, ''),
		COALESCE(i.estado, 'activo'),
		COALESCE(i.observaciones, '')
	FROM empresa_impresoras_funcionalidades f
	INNER JOIN empresa_impresoras i ON i.id = f.impresora_id AND i.empresa_id = f.empresa_id
	WHERE f.empresa_id = ?
		AND f.funcionalidad = ?
		AND COALESCE(NULLIF(TRIM(f.estado), ''), 'activo') = 'activo'
		AND COALESCE(NULLIF(TRIM(i.estado), ''), 'activo') = 'activo'
	ORDER BY COALESCE(f.prioridad, 100) ASC, f.id ASC
	LIMIT 1`, empresaID, funcionalidad)
	item, err := scanEmpresaImpresora(row)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func resolveEmpresaImpresoraPredeterminada(dbConn *sql.DB, empresaID int64) (*EmpresaImpresora, error) {
	row := queryRowSQLCompat(dbConn, `SELECT
		id,
		empresa_id,
		COALESCE(codigo, ''),
		COALESCE(nombre, ''),
		COALESCE(tipo_conexion, 'red'),
		COALESCE(direccion, ''),
		COALESCE(area_operativa, ''),
		COALESCE(formato_impresion, 'pos'),
		`+empresaImpresoraDefaultSelectExpr("")+`,
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_impresoras
	WHERE empresa_id = ?
		AND COALESCE(NULLIF(TRIM(estado), ''), 'activo') = 'activo'
	ORDER BY `+empresaImpresoraDefaultSelectExpr("")+` DESC, id ASC
	LIMIT 1`, empresaID)
	item, err := scanEmpresaImpresora(row)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// ResolveEmpresaImpresora selecciona impresora por producto -> funcionalidad -> predeterminada.
func ResolveEmpresaImpresora(dbConn *sql.DB, empresaID int64, funcionalidad string, productoID int64) (*EmpresaImpresoraResolucion, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id requerido")
	}
	funcionalidad = normalizeEmpresaImpresoraFuncionalidad(funcionalidad)

	if productoID > 0 {
		impresora, err := resolveEmpresaImpresoraByProducto(dbConn, empresaID, productoID)
		if err == nil {
			return &EmpresaImpresoraResolucion{
				EmpresaID:     empresaID,
				Funcionalidad: funcionalidad,
				ProductoID:    productoID,
				Fuente:        "producto",
				Impresora:     *impresora,
			}, nil
		}
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	if funcionalidad != "" {
		impresora, err := resolveEmpresaImpresoraByFuncionalidad(dbConn, empresaID, funcionalidad)
		if err == nil {
			return &EmpresaImpresoraResolucion{
				EmpresaID:     empresaID,
				Funcionalidad: funcionalidad,
				ProductoID:    productoID,
				Fuente:        "funcionalidad",
				Impresora:     *impresora,
			}, nil
		}
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	impresora, err := resolveEmpresaImpresoraPredeterminada(dbConn, empresaID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &EmpresaImpresoraResolucion{
		EmpresaID:     empresaID,
		Funcionalidad: funcionalidad,
		ProductoID:    productoID,
		Fuente:        "predeterminada",
		Impresora:     *impresora,
	}, nil
}
