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

// EmpresaImpresoraDispositivo vincula una impresora a un computador/caja detectado por PCS.
type EmpresaImpresoraDispositivo struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	DispositivoID      string `json:"dispositivo_id"`
	Etiqueta           string `json:"etiqueta,omitempty"`
	CajaCodigo         string `json:"caja_codigo,omitempty"`
	EstacionID         int64  `json:"estacion_id,omitempty"`
	Funcionalidad      string `json:"funcionalidad,omitempty"`
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
	DispositivoID string           `json:"dispositivo_id,omitempty"`
	Fuente        string           `json:"fuente"`
	Impresora     EmpresaImpresora `json:"impresora"`
}

// EmpresaImpresoraTrabajo representa un documento en cola para el agente local
// de impresion instalado en una caja o computador de la empresa.
type EmpresaImpresoraTrabajo struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	EstacionID         int64  `json:"estacion_id,omitempty"`
	AgenteID           string `json:"agente_id,omitempty"`
	ImpresoraID        int64  `json:"impresora_id,omitempty"`
	ImpresoraNombre    string `json:"impresora_nombre,omitempty"`
	ImpresoraCodigo    string `json:"impresora_codigo,omitempty"`
	Funcionalidad      string `json:"funcionalidad,omitempty"`
	TipoDocumento      string `json:"tipo_documento,omitempty"`
	ReferenciaTipo     string `json:"referencia_tipo,omitempty"`
	ReferenciaID       int64  `json:"referencia_id,omitempty"`
	TipoItem           string `json:"tipo_item,omitempty"`
	Titulo             string `json:"titulo,omitempty"`
	FormatoImpresion   string `json:"formato_impresion,omitempty"`
	ContenidoTipo      string `json:"contenido_tipo,omitempty"`
	Contenido          string `json:"contenido,omitempty"`
	Copias             int64  `json:"copias,omitempty"`
	Prioridad          int64  `json:"prioridad,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Intentos           int64  `json:"intentos,omitempty"`
	MaxIntentos        int64  `json:"max_intentos,omitempty"`
	TomadoPor          string `json:"tomado_por,omitempty"`
	TomadoEn           string `json:"tomado_en,omitempty"`
	ImpresoEn          string `json:"impreso_en,omitempty"`
	UltimoError        string `json:"ultimo_error,omitempty"`
	MetadataJSON       string `json:"metadata_json,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
}

// EnsureEmpresaImpresorasSchema crea/migra tablas del módulo de impresoras por empresa.
func EnsureEmpresaImpresorasSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
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
		`CREATE TABLE IF NOT EXISTS empresa_impresoras_dispositivos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			dispositivo_id TEXT NOT NULL,
			etiqueta TEXT,
			caja_codigo TEXT,
			estacion_id INTEGER DEFAULT 0,
			funcionalidad TEXT DEFAULT 'general',
			impresora_id INTEGER NOT NULL,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_impresoras_dispositivo_func ON empresa_impresoras_dispositivos(empresa_id, dispositivo_id, funcionalidad);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_impresoras_dispositivo_printer ON empresa_impresoras_dispositivos(empresa_id, impresora_id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_impresoras_dispositivo_estado ON empresa_impresoras_dispositivos(empresa_id, estado);`,
		`CREATE TABLE IF NOT EXISTS empresa_impresoras_cola (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			estacion_id INTEGER DEFAULT 0,
			agente_id TEXT,
			impresora_id INTEGER,
			funcionalidad TEXT,
			tipo_documento TEXT,
			referencia_tipo TEXT,
			referencia_id INTEGER DEFAULT 0,
			tipo_item TEXT,
			titulo TEXT,
			formato_impresion TEXT DEFAULT 'pos',
			contenido_tipo TEXT DEFAULT 'text/plain',
			contenido TEXT NOT NULL,
			copias INTEGER DEFAULT 1,
			prioridad INTEGER DEFAULT 100,
			estado TEXT DEFAULT 'pendiente',
			intentos INTEGER DEFAULT 0,
			max_intentos INTEGER DEFAULT 3,
			tomado_por TEXT,
			tomado_en TEXT,
			impreso_en TEXT,
			ultimo_error TEXT,
			metadata_json TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_impresoras_cola_estado ON empresa_impresoras_cola(empresa_id, estado, prioridad, id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_impresoras_cola_impresora ON empresa_impresoras_cola(empresa_id, impresora_id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_impresoras_cola_agente ON empresa_impresoras_cola(empresa_id, agente_id, estacion_id);`,
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

	// Asociacion por computador/caja detectado en el navegador o agente local.
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_dispositivos", "empresa_id", "INTEGER NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_dispositivos", "dispositivo_id", "TEXT NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_dispositivos", "etiqueta", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_dispositivos", "caja_codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_dispositivos", "estacion_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_dispositivos", "funcionalidad", "TEXT DEFAULT 'general'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_dispositivos", "impresora_id", "INTEGER NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_dispositivos", "fecha_creacion", "TEXT DEFAULT (CURRENT_TIMESTAMP)"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_dispositivos", "fecha_actualizacion", "TEXT DEFAULT (CURRENT_TIMESTAMP)"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_dispositivos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_dispositivos", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_dispositivos", "observaciones", "TEXT"); err != nil {
		return err
	}

	// Cola para agente local de impresion.
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "empresa_id", "INTEGER NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "estacion_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "agente_id", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "impresora_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "funcionalidad", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "tipo_documento", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "referencia_tipo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "referencia_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "tipo_item", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "titulo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "formato_impresion", "TEXT DEFAULT 'pos'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "contenido_tipo", "TEXT DEFAULT 'text/plain'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "contenido", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "copias", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "prioridad", "INTEGER DEFAULT 100"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "estado", "TEXT DEFAULT 'pendiente'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "intentos", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "max_intentos", "INTEGER DEFAULT 3"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "tomado_por", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "tomado_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "impreso_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "ultimo_error", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "metadata_json", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "fecha_creacion", "TEXT DEFAULT (CURRENT_TIMESTAMP)"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "fecha_actualizacion", "TEXT DEFAULT (CURRENT_TIMESTAMP)"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_impresoras_cola", "usuario_creador", "TEXT"); err != nil {
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

func scanEmpresaImpresoraDispositivo(row empresaImpresoraScanner) (*EmpresaImpresoraDispositivo, error) {
	item := EmpresaImpresoraDispositivo{}
	if err := row.Scan(
		&item.ID,
		&item.EmpresaID,
		&item.DispositivoID,
		&item.Etiqueta,
		&item.CajaCodigo,
		&item.EstacionID,
		&item.Funcionalidad,
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
	item.DispositivoID = normalizeEmpresaImpresoraDispositivoID(item.DispositivoID)
	item.Funcionalidad = normalizeEmpresaImpresoraFuncionalidad(item.Funcionalidad)
	item.Estado = normalizeEmpresaImpresoraEstado(item.Estado)
	return &item, nil
}

func scanEmpresaImpresoraTrabajo(row empresaImpresoraScanner) (*EmpresaImpresoraTrabajo, error) {
	item := EmpresaImpresoraTrabajo{}
	if err := row.Scan(
		&item.ID,
		&item.EmpresaID,
		&item.EstacionID,
		&item.AgenteID,
		&item.ImpresoraID,
		&item.ImpresoraNombre,
		&item.ImpresoraCodigo,
		&item.Funcionalidad,
		&item.TipoDocumento,
		&item.ReferenciaTipo,
		&item.ReferenciaID,
		&item.TipoItem,
		&item.Titulo,
		&item.FormatoImpresion,
		&item.ContenidoTipo,
		&item.Contenido,
		&item.Copias,
		&item.Prioridad,
		&item.Estado,
		&item.Intentos,
		&item.MaxIntentos,
		&item.TomadoPor,
		&item.TomadoEn,
		&item.ImpresoEn,
		&item.UltimoError,
		&item.MetadataJSON,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
	); err != nil {
		return nil, err
	}
	item.Funcionalidad = normalizeEmpresaImpresoraFuncionalidad(item.Funcionalidad)
	item.TipoItem = normalizeEmpresaImpresoraTipoItem(item.TipoItem)
	item.FormatoImpresion = normalizeEmpresaImpresoraFormato(item.FormatoImpresion)
	item.ContenidoTipo = normalizeEmpresaImpresoraContenidoTipo(item.ContenidoTipo)
	item.Estado = normalizeEmpresaImpresoraTrabajoEstado(item.Estado)
	return &item, nil
}

func normalizeEmpresaImpresoraEstado(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "inactivo" || value == "desactivado" || value == "off" {
		return "inactivo"
	}
	return "activo"
}

func normalizeEmpresaImpresoraTrabajoEstado(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "tomado", "en_proceso", "procesando":
		return "tomado"
	case "impreso", "completado", "ok":
		return "impreso"
	case "error", "fallido":
		return "error"
	case "cancelado", "anulado":
		return "cancelado"
	default:
		return "pendiente"
	}
}

func normalizeEmpresaImpresoraFormato(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "carta" {
		return "carta"
	}
	return defaultEmpresaImpresoraFormato
}

func normalizeEmpresaImpresoraContenidoTipo(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "text/html", "application/pdf", "application/json", "application/vnd.pcs.escpos-base64":
		return value
	default:
		return "text/plain"
	}
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

func normalizeEmpresaImpresoraTipoItem(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "producto", "receta":
		return value
	default:
		return ""
	}
}

func normalizeEmpresaImpresoraDispositivoID(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(value))
	for i := 0; i < len(value); i++ {
		ch := value[i]
		isLetter := ch >= 'a' && ch <= 'z'
		isDigit := ch >= '0' && ch <= '9'
		if isLetter || isDigit || ch == '_' || ch == '-' || ch == '.' || ch == ':' {
			b.WriteByte(ch)
		}
	}
	out := strings.Trim(b.String(), "_.:-")
	if len(out) > 120 {
		return out[:120]
	}
	return out
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

func normalizeEmpresaImpresoraTrabajoCopias(raw int64) int64 {
	if raw <= 0 {
		return 1
	}
	if raw > 10 {
		return 10
	}
	return raw
}

func normalizeEmpresaImpresoraTrabajoMaxIntentos(raw int64) int64 {
	if raw <= 0 {
		return 3
	}
	if raw > 10 {
		return 10
	}
	return raw
}

func trimEmpresaImpresoraText(raw string, max int) string {
	value := strings.TrimSpace(raw)
	if max > 0 && len(value) > max {
		return value[:max]
	}
	return value
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

func empresaImpresoraTrabajoSelectSQL() string {
	return `SELECT
		c.id,
		c.empresa_id,
		COALESCE(c.estacion_id, 0),
		COALESCE(c.agente_id, ''),
		COALESCE(c.impresora_id, 0),
		COALESCE(i.nombre, ''),
		COALESCE(i.codigo, ''),
		COALESCE(c.funcionalidad, ''),
		COALESCE(c.tipo_documento, ''),
		COALESCE(c.referencia_tipo, ''),
		COALESCE(c.referencia_id, 0),
		COALESCE(c.tipo_item, ''),
		COALESCE(c.titulo, ''),
		COALESCE(c.formato_impresion, 'pos'),
		COALESCE(c.contenido_tipo, 'text/plain'),
		COALESCE(c.contenido, ''),
		COALESCE(c.copias, 1),
		COALESCE(c.prioridad, 100),
		COALESCE(c.estado, 'pendiente'),
		COALESCE(c.intentos, 0),
		COALESCE(c.max_intentos, 3),
		COALESCE(c.tomado_por, ''),
		COALESCE(c.tomado_en, ''),
		COALESCE(c.impreso_en, ''),
		COALESCE(c.ultimo_error, ''),
		COALESCE(c.metadata_json, ''),
		COALESCE(c.fecha_creacion, ''),
		COALESCE(c.fecha_actualizacion, ''),
		COALESCE(c.usuario_creador, '')
	FROM empresa_impresoras_cola c
	LEFT JOIN empresa_impresoras i ON i.id = c.impresora_id AND i.empresa_id = c.empresa_id`
}

// GetEmpresaImpresoraTrabajoByID obtiene un trabajo de cola por empresa e id.
func GetEmpresaImpresoraTrabajoByID(dbConn *sql.DB, empresaID, trabajoID int64) (*EmpresaImpresoraTrabajo, error) {
	row := queryRowSQLCompat(dbConn, empresaImpresoraTrabajoSelectSQL()+`
	WHERE c.empresa_id = ? AND c.id = ?
	LIMIT 1`, empresaID, trabajoID)
	return scanEmpresaImpresoraTrabajo(row)
}

// ListEmpresaImpresoraCola lista trabajos de impresion por empresa y estado.
func ListEmpresaImpresoraCola(dbConn *sql.DB, empresaID int64, estado string, limit int64) ([]EmpresaImpresoraTrabajo, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id requerido")
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	query := empresaImpresoraTrabajoSelectSQL() + `
	WHERE c.empresa_id = ?`
	args := []interface{}{empresaID}
	estado = strings.ToLower(strings.TrimSpace(estado))
	if estado != "" && estado != "todos" {
		query += ` AND COALESCE(c.estado, 'pendiente') = ?`
		args = append(args, normalizeEmpresaImpresoraTrabajoEstado(estado))
	}
	query += ` ORDER BY c.id DESC LIMIT ?`
	args = append(args, limit)
	rows, err := querySQLCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaImpresoraTrabajo, 0)
	for rows.Next() {
		item, errScan := scanEmpresaImpresoraTrabajo(rows)
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

// CrearEmpresaImpresoraTrabajo encola un documento para que lo procese el agente local.
func CrearEmpresaImpresoraTrabajo(dbConn *sql.DB, payload EmpresaImpresoraTrabajo) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id requerido")
	}
	payload.Funcionalidad = normalizeEmpresaImpresoraFuncionalidad(payload.Funcionalidad)
	payload.TipoItem = normalizeEmpresaImpresoraTipoItem(payload.TipoItem)
	payload.FormatoImpresion = normalizeEmpresaImpresoraFormato(payload.FormatoImpresion)
	payload.ContenidoTipo = normalizeEmpresaImpresoraContenidoTipo(payload.ContenidoTipo)
	payload.Copias = normalizeEmpresaImpresoraTrabajoCopias(payload.Copias)
	payload.Prioridad = normalizeEmpresaImpresoraPrioridad(payload.Prioridad)
	payload.MaxIntentos = normalizeEmpresaImpresoraTrabajoMaxIntentos(payload.MaxIntentos)
	payload.Estado = "pendiente"
	payload.AgenteID = trimEmpresaImpresoraText(payload.AgenteID, 120)
	payload.TipoDocumento = trimEmpresaImpresoraText(payload.TipoDocumento, 80)
	payload.ReferenciaTipo = trimEmpresaImpresoraText(payload.ReferenciaTipo, 80)
	payload.Titulo = trimEmpresaImpresoraText(payload.Titulo, 180)
	payload.UsuarioCreador = trimEmpresaImpresoraText(payload.UsuarioCreador, 180)
	payload.MetadataJSON = trimEmpresaImpresoraText(payload.MetadataJSON, 20000)
	payload.Contenido = trimEmpresaImpresoraText(payload.Contenido, 200000)
	if payload.UsuarioCreador == "" {
		payload.UsuarioCreador = "sistema"
	}
	if payload.Contenido == "" {
		return 0, fmt.Errorf("contenido de impresion requerido")
	}
	if payload.ImpresoraID > 0 {
		if err := ensureEmpresaImpresoraExistsAndActive(dbConn, payload.EmpresaID, payload.ImpresoraID); err != nil {
			return 0, err
		}
	} else {
		resolved, err := ResolveEmpresaImpresoraOperacionConDispositivo(dbConn, payload.EmpresaID, payload.Funcionalidad, payload.TipoItem, payload.ReferenciaID, payload.AgenteID)
		if err != nil {
			return 0, err
		}
		if resolved != nil {
			payload.ImpresoraID = resolved.Impresora.ID
			payload.FormatoImpresion = normalizeEmpresaImpresoraFormato(resolved.Impresora.FormatoImpresion)
		}
	}

	return insertSQLCompat(dbConn, `INSERT INTO empresa_impresoras_cola (
		empresa_id,
		estacion_id,
		agente_id,
		impresora_id,
		funcionalidad,
		tipo_documento,
		referencia_tipo,
		referencia_id,
		tipo_item,
		titulo,
		formato_impresion,
		contenido_tipo,
		contenido,
		copias,
		prioridad,
		estado,
		intentos,
		max_intentos,
		metadata_json,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador
	) VALUES (
		?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'pendiente', 0, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?
	)`,
		payload.EmpresaID,
		payload.EstacionID,
		payload.AgenteID,
		payload.ImpresoraID,
		payload.Funcionalidad,
		payload.TipoDocumento,
		payload.ReferenciaTipo,
		payload.ReferenciaID,
		payload.TipoItem,
		payload.Titulo,
		payload.FormatoImpresion,
		payload.ContenidoTipo,
		payload.Contenido,
		payload.Copias,
		payload.Prioridad,
		payload.MaxIntentos,
		payload.MetadataJSON,
		payload.UsuarioCreador,
	)
}

// TomarEmpresaImpresoraTrabajos reclama trabajos pendientes para un agente local.
func TomarEmpresaImpresoraTrabajos(dbConn *sql.DB, empresaID int64, agenteID string, estacionID int64, limit int64) ([]EmpresaImpresoraTrabajo, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id requerido")
	}
	agenteID = trimEmpresaImpresoraText(agenteID, 120)
	if agenteID == "" {
		return nil, fmt.Errorf("agente_id requerido")
	}
	if limit <= 0 {
		limit = 5
	}
	if limit > 25 {
		limit = 25
	}
	rows, err := querySQLCompat(dbConn, `SELECT c.id
	FROM empresa_impresoras_cola c
	WHERE c.empresa_id = ?
		AND COALESCE(c.estado, 'pendiente') = 'pendiente'
		AND COALESCE(c.intentos, 0) < COALESCE(c.max_intentos, 3)
		AND (COALESCE(c.agente_id, '') = '' OR COALESCE(c.agente_id, '') = ?)
		AND (COALESCE(c.estacion_id, 0) = 0 OR COALESCE(c.estacion_id, 0) = ?)
	ORDER BY COALESCE(c.prioridad, 100) ASC, c.id ASC
	LIMIT ?`, empresaID, agenteID, estacionID, limit)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, err
	}
	_ = rows.Close()

	out := make([]EmpresaImpresoraTrabajo, 0, len(ids))
	for _, id := range ids {
		res, err := execSQLCompat(dbConn, `UPDATE empresa_impresoras_cola
			SET estado = 'tomado',
				tomado_por = ?,
				tomado_en = CURRENT_TIMESTAMP,
				intentos = COALESCE(intentos, 0) + 1,
				fecha_actualizacion = CURRENT_TIMESTAMP
			WHERE empresa_id = ? AND id = ? AND COALESCE(estado, 'pendiente') = 'pendiente'`, agenteID, empresaID, id)
		if err != nil {
			return nil, err
		}
		affected, _ := res.RowsAffected()
		if affected != 1 {
			continue
		}
		item, err := GetEmpresaImpresoraTrabajoByID(dbConn, empresaID, id)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, nil
}

// ActualizarEmpresaImpresoraTrabajoEstado cierra o marca error en un trabajo tomado.
func ActualizarEmpresaImpresoraTrabajoEstado(dbConn *sql.DB, empresaID, trabajoID int64, estado, agenteID, ultimoError string) error {
	if empresaID <= 0 || trabajoID <= 0 {
		return fmt.Errorf("empresa_id y trabajo_id requeridos")
	}
	estado = normalizeEmpresaImpresoraTrabajoEstado(estado)
	agenteID = trimEmpresaImpresoraText(agenteID, 120)
	ultimoError = trimEmpresaImpresoraText(ultimoError, 2000)
	if estado != "impreso" && estado != "error" && estado != "cancelado" {
		return fmt.Errorf("estado de cola no permitido")
	}
	var res sql.Result
	var err error
	if estado == "impreso" {
		res, err = execSQLCompat(dbConn, `UPDATE empresa_impresoras_cola
			SET estado = 'impreso',
				impreso_en = CURRENT_TIMESTAMP,
				tomado_por = CASE WHEN TRIM(COALESCE(?, '')) <> '' THEN ? ELSE tomado_por END,
				ultimo_error = '',
				fecha_actualizacion = CURRENT_TIMESTAMP
			WHERE empresa_id = ? AND id = ?`, agenteID, agenteID, empresaID, trabajoID)
	} else {
		res, err = execSQLCompat(dbConn, `UPDATE empresa_impresoras_cola
			SET estado = ?,
				tomado_por = CASE WHEN TRIM(COALESCE(?, '')) <> '' THEN ? ELSE tomado_por END,
				ultimo_error = ?,
				fecha_actualizacion = CURRENT_TIMESTAMP
			WHERE empresa_id = ? AND id = ?`, estado, agenteID, agenteID, ultimoError, empresaID, trabajoID)
	}
	if err != nil {
		return err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return fmt.Errorf("trabajo de impresion no encontrado")
	}
	return nil
}

// ReintentarEmpresaImpresoraTrabajo devuelve a pendiente un trabajo fallido o tomado.
func ReintentarEmpresaImpresoraTrabajo(dbConn *sql.DB, empresaID, trabajoID int64, usuario string) error {
	if empresaID <= 0 || trabajoID <= 0 {
		return fmt.Errorf("empresa_id y trabajo_id requeridos")
	}
	usuario = trimEmpresaImpresoraText(usuario, 180)
	res, err := execSQLCompat(dbConn, `UPDATE empresa_impresoras_cola
		SET estado = 'pendiente',
			tomado_por = '',
			tomado_en = '',
			impreso_en = '',
			ultimo_error = '',
			fecha_actualizacion = CURRENT_TIMESTAMP,
			usuario_creador = CASE WHEN TRIM(COALESCE(?, '')) <> '' THEN ? ELSE usuario_creador END
		WHERE empresa_id = ? AND id = ? AND COALESCE(estado, 'pendiente') IN ('tomado', 'error', 'cancelado')`, usuario, usuario, empresaID, trabajoID)
	if err != nil {
		return err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return fmt.Errorf("trabajo de impresion no encontrado o no reintentable")
	}
	return nil
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

// ListEmpresaImpresoraDispositivosByEmpresa lista impresoras asociadas a computadores/cajas.
func ListEmpresaImpresoraDispositivosByEmpresa(dbConn *sql.DB, empresaID int64) ([]EmpresaImpresoraDispositivo, error) {
	rows, err := querySQLCompat(dbConn, `SELECT
		d.id,
		d.empresa_id,
		COALESCE(d.dispositivo_id, ''),
		COALESCE(d.etiqueta, ''),
		COALESCE(d.caja_codigo, ''),
		COALESCE(d.estacion_id, 0),
		COALESCE(d.funcionalidad, 'general'),
		d.impresora_id,
		COALESCE(i.nombre, ''),
		COALESCE(i.codigo, ''),
		COALESCE(d.fecha_creacion, ''),
		COALESCE(d.fecha_actualizacion, ''),
		COALESCE(d.usuario_creador, ''),
		COALESCE(d.estado, 'activo'),
		COALESCE(d.observaciones, '')
	FROM empresa_impresoras_dispositivos d
	LEFT JOIN empresa_impresoras i ON i.id = d.impresora_id AND i.empresa_id = d.empresa_id
	WHERE d.empresa_id = ?
	ORDER BY d.etiqueta ASC, d.dispositivo_id ASC, d.funcionalidad ASC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaImpresoraDispositivo, 0)
	for rows.Next() {
		item, errScan := scanEmpresaImpresoraDispositivo(rows)
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

// UpsertEmpresaImpresoraDispositivo asocia una impresora activa a un computador detectado.
func UpsertEmpresaImpresoraDispositivo(dbConn *sql.DB, payload EmpresaImpresoraDispositivo) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id requerido")
	}
	payload.DispositivoID = normalizeEmpresaImpresoraDispositivoID(payload.DispositivoID)
	if payload.DispositivoID == "" {
		return 0, fmt.Errorf("dispositivo_id requerido")
	}
	if payload.ImpresoraID <= 0 {
		return 0, fmt.Errorf("impresora_id requerido")
	}
	if err := ensureEmpresaImpresoraExistsAndActive(dbConn, payload.EmpresaID, payload.ImpresoraID); err != nil {
		return 0, err
	}
	payload.Funcionalidad = normalizeEmpresaImpresoraFuncionalidad(payload.Funcionalidad)
	payload.Etiqueta = trimEmpresaImpresoraText(payload.Etiqueta, 120)
	payload.CajaCodigo = trimEmpresaImpresoraText(payload.CajaCodigo, 80)
	payload.UsuarioCreador = trimEmpresaImpresoraText(payload.UsuarioCreador, 180)
	payload.Observaciones = trimEmpresaImpresoraText(payload.Observaciones, 255)
	payload.Estado = normalizeEmpresaImpresoraEstado(payload.Estado)
	if payload.UsuarioCreador == "" {
		payload.UsuarioCreador = "sistema"
	}

	if _, err := execSQLCompat(dbConn, `INSERT INTO empresa_impresoras_dispositivos (
		empresa_id,
		dispositivo_id,
		etiqueta,
		caja_codigo,
		estacion_id,
		funcionalidad,
		impresora_id,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?, ?)
	ON CONFLICT(empresa_id, dispositivo_id, funcionalidad) DO UPDATE SET
		etiqueta = excluded.etiqueta,
		caja_codigo = excluded.caja_codigo,
		estacion_id = excluded.estacion_id,
		impresora_id = excluded.impresora_id,
		fecha_actualizacion = CURRENT_TIMESTAMP,
		usuario_creador = excluded.usuario_creador,
		estado = excluded.estado,
		observaciones = excluded.observaciones`,
		payload.EmpresaID,
		payload.DispositivoID,
		payload.Etiqueta,
		payload.CajaCodigo,
		payload.EstacionID,
		payload.Funcionalidad,
		payload.ImpresoraID,
		payload.UsuarioCreador,
		payload.Estado,
		payload.Observaciones,
	); err != nil {
		return 0, err
	}

	var id int64
	if err := queryRowSQLCompat(dbConn, `SELECT id FROM empresa_impresoras_dispositivos WHERE empresa_id = ? AND dispositivo_id = ? AND funcionalidad = ? LIMIT 1`, payload.EmpresaID, payload.DispositivoID, payload.Funcionalidad).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// DeleteEmpresaImpresoraDispositivo elimina la asociacion de una funcionalidad para un computador.
func DeleteEmpresaImpresoraDispositivo(dbConn *sql.DB, empresaID int64, dispositivoID, funcionalidad string) error {
	if empresaID <= 0 {
		return fmt.Errorf("empresa_id requerido")
	}
	dispositivoID = normalizeEmpresaImpresoraDispositivoID(dispositivoID)
	if dispositivoID == "" {
		return fmt.Errorf("dispositivo_id requerido")
	}
	funcionalidad = normalizeEmpresaImpresoraFuncionalidad(funcionalidad)
	_, err := execSQLCompat(dbConn, `DELETE FROM empresa_impresoras_dispositivos WHERE empresa_id = ? AND dispositivo_id = ? AND funcionalidad = ?`, empresaID, dispositivoID, funcionalidad)
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

func resolveEmpresaImpresoraByDispositivo(dbConn *sql.DB, empresaID int64, dispositivoID, funcionalidad string) (*EmpresaImpresora, string, error) {
	dispositivoID = normalizeEmpresaImpresoraDispositivoID(dispositivoID)
	if dispositivoID == "" {
		return nil, "", sql.ErrNoRows
	}
	funcionalidad = normalizeEmpresaImpresoraFuncionalidad(funcionalidad)
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
		COALESCE(i.observaciones, ''),
		COALESCE(d.funcionalidad, 'general')
	FROM empresa_impresoras_dispositivos d
	INNER JOIN empresa_impresoras i ON i.id = d.impresora_id AND i.empresa_id = d.empresa_id
	WHERE d.empresa_id = ?
		AND d.dispositivo_id = ?
		AND d.funcionalidad IN (?, 'general')
		AND COALESCE(NULLIF(TRIM(d.estado), ''), 'activo') = 'activo'
		AND COALESCE(NULLIF(TRIM(i.estado), ''), 'activo') = 'activo'
	ORDER BY CASE WHEN d.funcionalidad = ? THEN 0 ELSE 1 END, d.id ASC
	LIMIT 1`, empresaID, dispositivoID, funcionalidad, funcionalidad)
	item := EmpresaImpresora{}
	var esPredeterminadaInt int
	var resolvedFunc string
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
		&resolvedFunc,
	); err != nil {
		return nil, "", err
	}
	item.EsPredeterminada = esPredeterminadaInt == 1
	item.FormatoImpresion = normalizeEmpresaImpresoraFormato(item.FormatoImpresion)
	item.TipoConexion = normalizeEmpresaImpresoraTipoConexion(item.TipoConexion)
	item.Estado = normalizeEmpresaImpresoraEstado(item.Estado)
	return &item, normalizeEmpresaImpresoraFuncionalidad(resolvedFunc), nil
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
	return ResolveEmpresaImpresoraOperacionConDispositivo(dbConn, empresaID, funcionalidad, tipoItem, referenciaID, "")
}

// ResolveEmpresaImpresoraOperacionConDispositivo selecciona impresora por item -> computador -> funcionalidad -> predeterminada.
func ResolveEmpresaImpresoraOperacionConDispositivo(dbConn *sql.DB, empresaID int64, funcionalidad string, tipoItem string, referenciaID int64, dispositivoID string) (*EmpresaImpresoraResolucion, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id requerido")
	}
	funcionalidad = normalizeEmpresaImpresoraFuncionalidad(funcionalidad)
	tipoItem = strings.ToLower(strings.TrimSpace(tipoItem))
	dispositivoID = normalizeEmpresaImpresoraDispositivoID(dispositivoID)
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

	if dispositivoID != "" {
		impresora, dispositivoFuncionalidad, err := resolveEmpresaImpresoraByDispositivo(dbConn, empresaID, dispositivoID, funcionalidad)
		if err == nil {
			fuente := "computador"
			if dispositivoFuncionalidad == "general" && funcionalidad != "general" {
				fuente = "computador_general"
			}
			return &EmpresaImpresoraResolucion{
				EmpresaID:     empresaID,
				Funcionalidad: funcionalidad,
				ProductoID:    mapEmpresaImpresoraProductoID(tipoItem, referenciaID),
				RecetaID:      mapEmpresaImpresoraRecetaID(tipoItem, referenciaID),
				TipoItem:      tipoItem,
				DispositivoID: dispositivoID,
				Fuente:        fuente,
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
