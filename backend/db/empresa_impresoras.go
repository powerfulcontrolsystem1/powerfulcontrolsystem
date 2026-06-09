package db

import (
	"database/sql"
	"fmt"
	"strings"
)

const (
	defaultEmpresaImpresoraFormato      = "pos"
	defaultEmpresaImpresoraTipoConexion = "red"
	DefaultEmpresaPOS80PrinterCode      = "POS_80MM"
)

var DefaultEmpresaPOS80Funcionalidades = []string{"general", "corte_caja", "turno_reporte", "cajon_monedero"}

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

// EmpresaImpresoraProductoRegla define una regla masiva de impresora para productos.
type EmpresaImpresoraProductoRegla struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Alcance            string `json:"alcance"`
	CategoriaID        int64  `json:"categoria_id,omitempty"`
	CategoriaNombre    string `json:"categoria_nombre,omitempty"`
	ImpresoraID        int64  `json:"impresora_id"`
	ImpresoraNombre    string `json:"impresora_nombre,omitempty"`
	ImpresoraCodigo    string `json:"impresora_codigo,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// EmpresaImpresoraReceta define la impresora asignada por receta.
type EmpresaImpresoraReceta struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	RecetaID           int64  `json:"receta_id"`
	RecetaNombre       string `json:"receta_nombre,omitempty"`
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
	CategoriaID   int64            `json:"categoria_id,omitempty"`
	RecetaID      int64            `json:"receta_id,omitempty"`
	TipoItem      string           `json:"tipo_item,omitempty"`
	Fuente        string           `json:"fuente"`
	Impresora     EmpresaImpresora `json:"impresora"`
}

// EnsureEmpresaImpresorasSchema crea/migra tablas del módulo de impresoras por empresa.
func EnsureEmpresaImpresorasSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_impresoras (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			tipo_conexion TEXT DEFAULT 'red',
			direccion TEXT,
			area_operativa TEXT,
			formato_impresion TEXT DEFAULT 'pos',
			es_predeterminada INTEGER DEFAULT 0,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_impresoras_empresa_codigo ON empresa_impresoras(empresa_id, codigo);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_impresoras_empresa_estado ON empresa_impresoras(empresa_id, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_impresoras_empresa_default ON empresa_impresoras(empresa_id, es_predeterminada);`,
		`CREATE TABLE IF NOT EXISTS empresa_impresoras_funcionalidades (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			funcionalidad TEXT NOT NULL,
			impresora_id INTEGER NOT NULL,
			prioridad INTEGER DEFAULT 100,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_impresoras_funcionalidad ON empresa_impresoras_funcionalidades(empresa_id, funcionalidad);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_impresoras_funcionalidades_printer ON empresa_impresoras_funcionalidades(empresa_id, impresora_id);`,
		`CREATE TABLE IF NOT EXISTS empresa_impresoras_productos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			producto_id INTEGER NOT NULL,
			impresora_id INTEGER NOT NULL,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_impresoras_producto ON empresa_impresoras_productos(empresa_id, producto_id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_impresoras_productos_printer ON empresa_impresoras_productos(empresa_id, impresora_id);`,
		`CREATE TABLE IF NOT EXISTS empresa_impresoras_productos_reglas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			alcance TEXT NOT NULL DEFAULT 'todos',
			categoria_id INTEGER NOT NULL DEFAULT 0,
			impresora_id INTEGER NOT NULL,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_impresoras_productos_regla ON empresa_impresoras_productos_reglas(empresa_id, alcance, categoria_id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_impresoras_productos_reglas_printer ON empresa_impresoras_productos_reglas(empresa_id, impresora_id);`,
		`CREATE TABLE IF NOT EXISTS empresa_impresoras_recetas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			receta_id INTEGER NOT NULL,
			impresora_id INTEGER NOT NULL,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_impresoras_receta ON empresa_impresoras_recetas(empresa_id, receta_id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_impresoras_recetas_printer ON empresa_impresoras_recetas(empresa_id, impresora_id);`,
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
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras", "fecha_creacion", "TEXT DEFAULT (CURRENT_TIMESTAMP)"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras", "fecha_actualizacion", "TEXT DEFAULT (CURRENT_TIMESTAMP)"); err != nil {
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
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_funcionalidades", "fecha_creacion", "TEXT DEFAULT (CURRENT_TIMESTAMP)"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_funcionalidades", "fecha_actualizacion", "TEXT DEFAULT (CURRENT_TIMESTAMP)"); err != nil {
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
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_productos", "fecha_creacion", "TEXT DEFAULT (CURRENT_TIMESTAMP)"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_productos", "fecha_actualizacion", "TEXT DEFAULT (CURRENT_TIMESTAMP)"); err != nil {
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

	// Reglas masivas por todos los productos o por categoria
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_productos_reglas", "empresa_id", "INTEGER NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_productos_reglas", "alcance", "TEXT NOT NULL DEFAULT 'todos'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_productos_reglas", "categoria_id", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_productos_reglas", "impresora_id", "INTEGER NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_productos_reglas", "fecha_creacion", "TEXT DEFAULT (CURRENT_TIMESTAMP)"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_productos_reglas", "fecha_actualizacion", "TEXT DEFAULT (CURRENT_TIMESTAMP)"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_productos_reglas", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_productos_reglas", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_productos_reglas", "observaciones", "TEXT"); err != nil {
		return err
	}

	// Tabla por receta
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_recetas", "empresa_id", "INTEGER NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_recetas", "receta_id", "INTEGER NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_recetas", "impresora_id", "INTEGER NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_recetas", "fecha_creacion", "TEXT DEFAULT (CURRENT_TIMESTAMP)"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_recetas", "fecha_actualizacion", "TEXT DEFAULT (CURRENT_TIMESTAMP)"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_recetas", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_recetas", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_recetas", "observaciones", "TEXT"); err != nil {
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

// EnsureEmpresaPOS80Defaults deja la empresa preparada para imprimir reportes y
// operaciones de caja en ticket POS 80mm por defecto.
func EnsureEmpresaPOS80Defaults(dbConn *sql.DB, empresaID int64, usuario string) (int64, error) {
	if dbConn == nil {
		return 0, fmt.Errorf("db connection is nil")
	}
	if empresaID <= 0 {
		return 0, fmt.Errorf("empresa_id requerido")
	}
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = "sistema-pos80"
	}
	if err := EnsureEmpresaImpresorasSchema(dbConn); err != nil {
		return 0, err
	}
	if err := EnsureEmpresaCorteCajaConfiguracionSchema(dbConn); err != nil {
		return 0, err
	}
	if _, err := execSQLCompat(dbConn, `INSERT INTO empresa_corte_caja_configuracion (
		empresa_id, formato_impresion, usuario_creador, estado, observaciones
	) VALUES (
		?, 'pos', ?, 'activo', 'Reporte de turno configurado para impresora POS 80mm por defecto'
	)
	ON CONFLICT(empresa_id) DO UPDATE SET
		formato_impresion = 'pos',
		fecha_actualizacion = CURRENT_TIMESTAMP,
		usuario_creador = excluded.usuario_creador,
		estado = 'activo',
		observaciones = excluded.observaciones`, empresaID, usuario); err != nil {
		return 0, err
	}

	var existingID int64
	err := queryRowSQLCompat(dbConn, `SELECT id FROM empresa_impresoras WHERE empresa_id = ? AND codigo = ? LIMIT 1`, empresaID, DefaultEmpresaPOS80PrinterCode).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	if err == sql.ErrNoRows {
		existingID = 0
	}

	printerID, err := UpsertEmpresaImpresora(dbConn, EmpresaImpresora{
		ID:               existingID,
		EmpresaID:        empresaID,
		Codigo:           DefaultEmpresaPOS80PrinterCode,
		Nombre:           "Impresora POS 80mm",
		TipoConexion:     "windows",
		Direccion:        "POS 80mm",
		AreaOperativa:    "caja",
		FormatoImpresion: "pos",
		EsPredeterminada: true,
		UsuarioCreador:   usuario,
		Estado:           "activo",
		Observaciones:    "Impresora POS 80mm activa por defecto para caja, reportes y turno",
	})
	if err != nil {
		return 0, err
	}
	for _, funcionalidad := range DefaultEmpresaPOS80Funcionalidades {
		if _, err := UpsertEmpresaImpresoraFuncionalidad(dbConn, EmpresaImpresoraFuncionalidad{
			EmpresaID:      empresaID,
			Funcionalidad:  funcionalidad,
			ImpresoraID:    printerID,
			Prioridad:      10,
			UsuarioCreador: usuario,
			Estado:         "activo",
			Observaciones:  "Asignado a impresora POS 80mm por defecto",
		}); err != nil {
			return 0, err
		}
	}
	return printerID, nil
}

// EnsureAllEmpresasPOS80Defaults aplica la configuracion POS 80mm a todas las
// empresas activas registradas.
func EnsureAllEmpresasPOS80Defaults(dbConn *sql.DB, usuario string) (int, error) {
	if dbConn == nil {
		return 0, fmt.Errorf("db connection is nil")
	}
	rows, err := querySQLCompat(dbConn, `SELECT COALESCE(empresa_id, id)
		FROM empresas
		WHERE LOWER(COALESCE(estado, 'activo')) <> 'inactivo'
		ORDER BY COALESCE(empresa_id, id) ASC`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	ids := make([]int64, 0)
	seen := map[int64]bool{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
		if id > 0 && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	count := 0
	for _, id := range ids {
		if _, err := EnsureEmpresaPOS80Defaults(dbConn, id, usuario); err != nil {
			return count, fmt.Errorf("empresa_id %d: %w", id, err)
		}
		count++
	}
	return count, nil
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
		if _, err := execSQLCompat(dbConn, `UPDATE empresa_impresoras SET es_predeterminada = CASE WHEN id = ? THEN 1 ELSE 0 END, fecha_actualizacion = CURRENT_TIMESTAMP WHERE empresa_id = ? AND COALESCE(NULLIF(TRIM(estado), ''), 'activo') = 'activo'`, keepID, empresaID); err != nil {
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
		if _, err := execSQLCompat(dbConn, `UPDATE empresa_impresoras SET es_predeterminada = CASE WHEN id = ? THEN 1 ELSE 0 END, fecha_actualizacion = CURRENT_TIMESTAMP WHERE empresa_id = ?`, firstActiveID, empresaID); err != nil {
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
			fecha_actualizacion = CURRENT_TIMESTAMP,
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
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?, ?)`,
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
		if _, err := execSQLCompat(dbConn, `UPDATE empresa_impresoras SET es_predeterminada = CASE WHEN id = ? THEN 1 ELSE 0 END, fecha_actualizacion = CURRENT_TIMESTAMP WHERE empresa_id = ?`, payload.ID, payload.EmpresaID); err != nil {
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
	if _, err := execSQLCompat(dbConn, `UPDATE empresa_impresoras SET es_predeterminada = CASE WHEN id = ? THEN 1 ELSE 0 END, fecha_actualizacion = CURRENT_TIMESTAMP, usuario_creador = ? WHERE empresa_id = ?`, impresoraID, strings.TrimSpace(usuario), empresaID); err != nil {
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
	if _, err := execSQLCompat(dbConn, `UPDATE empresa_impresoras SET estado = ?, es_predeterminada = CASE WHEN ? = 1 THEN es_predeterminada ELSE 0 END, fecha_actualizacion = CURRENT_TIMESTAMP, usuario_creador = ? WHERE empresa_id = ? AND id = ?`, normEstado, esPredeterminada, strings.TrimSpace(usuario), empresaID, impresoraID); err != nil {
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
	) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?, ?)
	ON CONFLICT(empresa_id, funcionalidad) DO UPDATE SET
		impresora_id = excluded.impresora_id,
		prioridad = excluded.prioridad,
		fecha_actualizacion = CURRENT_TIMESTAMP,
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
	) VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?, ?)
	ON CONFLICT(empresa_id, producto_id) DO UPDATE SET
		impresora_id = excluded.impresora_id,
		fecha_actualizacion = CURRENT_TIMESTAMP,
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

func normalizeEmpresaImpresoraProductoReglaAlcance(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "categoria", "category":
		return "categoria"
	default:
		return "todos"
	}
}

// ListEmpresaImpresoraProductoReglasByEmpresa lista reglas masivas por todos los productos o por categoria.
func ListEmpresaImpresoraProductoReglasByEmpresa(dbConn *sql.DB, empresaID int64) ([]EmpresaImpresoraProductoRegla, error) {
	rows, err := querySQLCompat(dbConn, `SELECT
		r.id,
		r.empresa_id,
		COALESCE(r.alcance, 'todos'),
		COALESCE(r.categoria_id, 0),
		COALESCE(c.nombre, ''),
		r.impresora_id,
		COALESCE(i.nombre, ''),
		COALESCE(i.codigo, ''),
		COALESCE(r.fecha_creacion, ''),
		COALESCE(r.fecha_actualizacion, ''),
		COALESCE(r.usuario_creador, ''),
		COALESCE(r.estado, 'activo'),
		COALESCE(r.observaciones, '')
	FROM empresa_impresoras_productos_reglas r
	LEFT JOIN categorias_productos c ON c.id = r.categoria_id AND c.empresa_id = r.empresa_id
	LEFT JOIN empresa_impresoras i ON i.id = r.impresora_id AND i.empresa_id = r.empresa_id
	WHERE r.empresa_id = ?
	ORDER BY CASE WHEN COALESCE(r.alcance, 'todos') = 'todos' THEN 0 ELSE 1 END, c.nombre ASC, r.id ASC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaImpresoraProductoRegla, 0)
	for rows.Next() {
		item := EmpresaImpresoraProductoRegla{}
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.Alcance,
			&item.CategoriaID,
			&item.CategoriaNombre,
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
		item.Alcance = normalizeEmpresaImpresoraProductoReglaAlcance(item.Alcance)
		item.Estado = normalizeEmpresaImpresoraEstado(item.Estado)
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func ensureEmpresaCategoriaProductoExists(dbConn *sql.DB, empresaID, categoriaID int64) error {
	var id int64
	if err := queryRowSQLCompat(dbConn, `SELECT id FROM categorias_productos WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, categoriaID).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("categoria no encontrada")
		}
		return err
	}
	return nil
}

// UpsertEmpresaImpresoraProductoRegla crea/actualiza una regla masiva de producto -> impresora.
func UpsertEmpresaImpresoraProductoRegla(dbConn *sql.DB, payload EmpresaImpresoraProductoRegla) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id requerido")
	}
	if payload.ImpresoraID <= 0 {
		return 0, fmt.Errorf("impresora_id requerido")
	}
	payload.Alcance = normalizeEmpresaImpresoraProductoReglaAlcance(payload.Alcance)
	if payload.Alcance == "todos" {
		payload.CategoriaID = 0
	} else if payload.CategoriaID <= 0 {
		return 0, fmt.Errorf("categoria_id requerido")
	}
	if payload.Alcance == "categoria" {
		if err := ensureEmpresaCategoriaProductoExists(dbConn, payload.EmpresaID, payload.CategoriaID); err != nil {
			return 0, err
		}
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

	if _, err := execSQLCompat(dbConn, `INSERT INTO empresa_impresoras_productos_reglas (
		empresa_id,
		alcance,
		categoria_id,
		impresora_id,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?, ?)
	ON CONFLICT(empresa_id, alcance, categoria_id) DO UPDATE SET
		impresora_id = excluded.impresora_id,
		fecha_actualizacion = CURRENT_TIMESTAMP,
		usuario_creador = excluded.usuario_creador,
		estado = excluded.estado,
		observaciones = excluded.observaciones`,
		payload.EmpresaID,
		payload.Alcance,
		payload.CategoriaID,
		payload.ImpresoraID,
		payload.UsuarioCreador,
		payload.Estado,
		payload.Observaciones,
	); err != nil {
		return 0, err
	}

	var id int64
	if err := queryRowSQLCompat(dbConn, `SELECT id FROM empresa_impresoras_productos_reglas WHERE empresa_id = ? AND alcance = ? AND categoria_id = ? LIMIT 1`, payload.EmpresaID, payload.Alcance, payload.CategoriaID).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// DeleteEmpresaImpresoraProductoRegla elimina una regla masiva por todos los productos o por categoria.
func DeleteEmpresaImpresoraProductoRegla(dbConn *sql.DB, empresaID int64, alcance string, categoriaID int64) error {
	if empresaID <= 0 {
		return fmt.Errorf("empresa_id requerido")
	}
	alcance = normalizeEmpresaImpresoraProductoReglaAlcance(alcance)
	if alcance == "todos" {
		categoriaID = 0
	} else if categoriaID <= 0 {
		return fmt.Errorf("categoria_id requerido")
	}
	_, err := execSQLCompat(dbConn, `DELETE FROM empresa_impresoras_productos_reglas WHERE empresa_id = ? AND alcance = ? AND categoria_id = ?`, empresaID, alcance, categoriaID)
	return err
}

// ListEmpresaImpresoraRecetasByEmpresa lista asignaciones por receta.
func ListEmpresaImpresoraRecetasByEmpresa(dbConn *sql.DB, empresaID int64) ([]EmpresaImpresoraReceta, error) {
	rows, err := querySQLCompat(dbConn, `SELECT
		a.id,
		a.empresa_id,
		a.receta_id,
		COALESCE(c.nombre, ''),
		a.impresora_id,
		COALESCE(i.nombre, ''),
		COALESCE(i.codigo, ''),
		COALESCE(a.fecha_creacion, ''),
		COALESCE(a.fecha_actualizacion, ''),
		COALESCE(a.usuario_creador, ''),
		COALESCE(a.estado, 'activo'),
		COALESCE(a.observaciones, '')
	FROM empresa_impresoras_recetas a
	LEFT JOIN recetas_productos c ON c.id = a.receta_id AND c.empresa_id = a.empresa_id
	LEFT JOIN empresa_impresoras i ON i.id = a.impresora_id AND i.empresa_id = a.empresa_id
	WHERE a.empresa_id = ?
	ORDER BY c.nombre ASC, a.receta_id ASC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaImpresoraReceta, 0)
	for rows.Next() {
		item := EmpresaImpresoraReceta{}
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.RecetaID,
			&item.RecetaNombre,
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

func ensureEmpresaRecetaExists(dbConn *sql.DB, empresaID, recetaID int64) error {
	var id int64
	if err := queryRowSQLCompat(dbConn, `SELECT id FROM recetas_productos WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, recetaID).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("receta no encontrada")
		}
		return err
	}
	return nil
}

// UpsertEmpresaImpresoraReceta crea/actualiza asignacion receta -> impresora.
func UpsertEmpresaImpresoraReceta(dbConn *sql.DB, payload EmpresaImpresoraReceta) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id requerido")
	}
	if payload.RecetaID <= 0 {
		return 0, fmt.Errorf("receta_id requerido")
	}
	if payload.ImpresoraID <= 0 {
		return 0, fmt.Errorf("impresora_id requerido")
	}
	if err := ensureEmpresaRecetaExists(dbConn, payload.EmpresaID, payload.RecetaID); err != nil {
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

	if _, err := execSQLCompat(dbConn, `INSERT INTO empresa_impresoras_recetas (
		empresa_id,
		receta_id,
		impresora_id,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?, ?)
	ON CONFLICT(empresa_id, receta_id) DO UPDATE SET
		impresora_id = excluded.impresora_id,
		fecha_actualizacion = CURRENT_TIMESTAMP,
		usuario_creador = excluded.usuario_creador,
		estado = excluded.estado,
		observaciones = excluded.observaciones`,
		payload.EmpresaID,
		payload.RecetaID,
		payload.ImpresoraID,
		payload.UsuarioCreador,
		payload.Estado,
		payload.Observaciones,
	); err != nil {
		return 0, err
	}

	var id int64
	if err := queryRowSQLCompat(dbConn, `SELECT id FROM empresa_impresoras_recetas WHERE empresa_id = ? AND receta_id = ? LIMIT 1`, payload.EmpresaID, payload.RecetaID).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// DeleteEmpresaImpresoraReceta elimina asignacion por receta.
func DeleteEmpresaImpresoraReceta(dbConn *sql.DB, empresaID, recetaID int64) error {
	if empresaID <= 0 {
		return fmt.Errorf("empresa_id requerido")
	}
	if recetaID <= 0 {
		return fmt.Errorf("receta_id requerido")
	}
	_, err := execSQLCompat(dbConn, `DELETE FROM empresa_impresoras_recetas WHERE empresa_id = ? AND receta_id = ?`, empresaID, recetaID)
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

func getEmpresaProductoCategoriaID(dbConn *sql.DB, empresaID, productoID int64) (int64, error) {
	var categoriaID sql.NullInt64
	if err := queryRowSQLCompat(dbConn, `SELECT categoria_id FROM productos WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, productoID).Scan(&categoriaID); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	if categoriaID.Valid && categoriaID.Int64 > 0 {
		return categoriaID.Int64, nil
	}
	return 0, nil
}

func resolveEmpresaImpresoraByProductoRegla(dbConn *sql.DB, empresaID int64, alcance string, categoriaID int64) (*EmpresaImpresora, error) {
	alcance = normalizeEmpresaImpresoraProductoReglaAlcance(alcance)
	if alcance == "todos" {
		categoriaID = 0
	}
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
	FROM empresa_impresoras_productos_reglas r
	INNER JOIN empresa_impresoras i ON i.id = r.impresora_id AND i.empresa_id = r.empresa_id
	WHERE r.empresa_id = ?
		AND r.alcance = ?
		AND COALESCE(r.categoria_id, 0) = ?
		AND COALESCE(NULLIF(TRIM(r.estado), ''), 'activo') = 'activo'
		AND COALESCE(NULLIF(TRIM(i.estado), ''), 'activo') = 'activo'
	LIMIT 1`, empresaID, alcance, categoriaID)
	item, err := scanEmpresaImpresora(row)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func resolveEmpresaImpresoraByReceta(dbConn *sql.DB, empresaID, recetaID int64) (*EmpresaImpresora, error) {
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
	FROM empresa_impresoras_recetas c
	INNER JOIN empresa_impresoras i ON i.id = c.impresora_id AND i.empresa_id = c.empresa_id
	WHERE c.empresa_id = ?
		AND c.receta_id = ?
		AND COALESCE(NULLIF(TRIM(c.estado), ''), 'activo') = 'activo'
		AND COALESCE(NULLIF(TRIM(i.estado), ''), 'activo') = 'activo'
	LIMIT 1`, empresaID, recetaID)
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

// ResolveEmpresaImpresoraOperacion selecciona impresora por item -> categoria/todos -> funcionalidad -> predeterminada.
func ResolveEmpresaImpresoraOperacion(dbConn *sql.DB, empresaID int64, funcionalidad string, tipoItem string, referenciaID int64) (*EmpresaImpresoraResolucion, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id requerido")
	}
	funcionalidad = normalizeEmpresaImpresoraFuncionalidad(funcionalidad)
	tipoItem = strings.ToLower(strings.TrimSpace(tipoItem))
	if tipoItem == "" && referenciaID > 0 {
		tipoItem = "producto"
	}

	if tipoItem == "receta" && referenciaID > 0 {
		impresora, err := resolveEmpresaImpresoraByReceta(dbConn, empresaID, referenciaID)
		if err == nil {
			return &EmpresaImpresoraResolucion{
				EmpresaID:     empresaID,
				Funcionalidad: funcionalidad,
				RecetaID:      referenciaID,
				TipoItem:      "receta",
				Fuente:        "receta",
				Impresora:     *impresora,
			}, nil
		}
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	if tipoItem == "producto" && referenciaID > 0 {
		impresora, err := resolveEmpresaImpresoraByProducto(dbConn, empresaID, referenciaID)
		if err == nil {
			return &EmpresaImpresoraResolucion{
				EmpresaID:     empresaID,
				Funcionalidad: funcionalidad,
				ProductoID:    referenciaID,
				TipoItem:      "producto",
				Fuente:        "producto",
				Impresora:     *impresora,
			}, nil
		}
		if err != sql.ErrNoRows {
			return nil, err
		}

		categoriaID, err := getEmpresaProductoCategoriaID(dbConn, empresaID, referenciaID)
		if err != nil {
			return nil, err
		}
		if categoriaID > 0 {
			impresora, err := resolveEmpresaImpresoraByProductoRegla(dbConn, empresaID, "categoria", categoriaID)
			if err == nil {
				return &EmpresaImpresoraResolucion{
					EmpresaID:     empresaID,
					Funcionalidad: funcionalidad,
					ProductoID:    referenciaID,
					CategoriaID:   categoriaID,
					TipoItem:      "producto",
					Fuente:        "categoria_producto",
					Impresora:     *impresora,
				}, nil
			}
			if err != sql.ErrNoRows {
				return nil, err
			}
		}

		impresora, err = resolveEmpresaImpresoraByProductoRegla(dbConn, empresaID, "todos", 0)
		if err == nil {
			return &EmpresaImpresoraResolucion{
				EmpresaID:     empresaID,
				Funcionalidad: funcionalidad,
				ProductoID:    referenciaID,
				CategoriaID:   categoriaID,
				TipoItem:      "producto",
				Fuente:        "todos_productos",
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
				ProductoID:    mapEmpresaImpresoraProductoID(tipoItem, referenciaID),
				RecetaID:      mapEmpresaImpresoraRecetaID(tipoItem, referenciaID),
				TipoItem:      tipoItem,
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
		ProductoID:    mapEmpresaImpresoraProductoID(tipoItem, referenciaID),
		RecetaID:      mapEmpresaImpresoraRecetaID(tipoItem, referenciaID),
		TipoItem:      tipoItem,
		Fuente:        "predeterminada",
		Impresora:     *impresora,
	}, nil
}

func mapEmpresaImpresoraProductoID(tipoItem string, referenciaID int64) int64 {
	if tipoItem == "producto" {
		return referenciaID
	}
	return 0
}

func mapEmpresaImpresoraRecetaID(tipoItem string, referenciaID int64) int64 {
	if tipoItem == "receta" {
		return referenciaID
	}
	return 0
}

// ResolveEmpresaImpresora mantiene compatibilidad con la resolucion historica por producto.
func ResolveEmpresaImpresora(dbConn *sql.DB, empresaID int64, funcionalidad string, productoID int64) (*EmpresaImpresoraResolucion, error) {
	return ResolveEmpresaImpresoraOperacion(dbConn, empresaID, funcionalidad, "producto", productoID)
}
