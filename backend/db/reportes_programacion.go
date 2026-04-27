package db

import (
	"database/sql"
	"fmt"
)

// EnsureEmpresaReportesProgramacionSchema ensures scheduling, template versioning,
// and execution trace tables for reportes module 31.
func EnsureEmpresaReportesProgramacionSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("db connection is nil")
	}

	isPostgres := isPostgresDialect()

	createProgramaciones := `CREATE TABLE IF NOT EXISTS empresa_reportes_programaciones (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER NOT NULL,
		nombre TEXT NOT NULL,
		dataset_key TEXT NOT NULL,
		nivel TEXT DEFAULT 'operativo',
		formatos TEXT DEFAULT '["json"]',
		parametros_json TEXT DEFAULT '{}',
		template_codigo TEXT,
		template_version INTEGER DEFAULT 0,
		frecuencia TEXT DEFAULT 'diario',
		hora_envio TEXT DEFAULT '08:00',
		timezone TEXT DEFAULT 'America/Bogota',
		destinatarios TEXT,
		ultimo_ejecutado_en TEXT,
		proximo_ejecutado_en TEXT,
		activa INTEGER DEFAULT 1,
		validacion_consistencia INTEGER DEFAULT 1,
		hash_ultima_ejecucion TEXT,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`
	if isPostgres {
		createProgramaciones = `CREATE TABLE IF NOT EXISTS empresa_reportes_programaciones (
		id SERIAL PRIMARY KEY,
		empresa_id INTEGER NOT NULL,
		nombre TEXT NOT NULL,
		dataset_key TEXT NOT NULL,
		nivel TEXT DEFAULT 'operativo',
		formatos TEXT DEFAULT '["json"]',
		parametros_json TEXT DEFAULT '{}',
		template_codigo TEXT,
		template_version INTEGER DEFAULT 0,
		frecuencia TEXT DEFAULT 'diario',
		hora_envio TEXT DEFAULT '08:00',
		timezone TEXT DEFAULT 'America/Bogota',
		destinatarios TEXT,
		ultimo_ejecutado_en TEXT,
		proximo_ejecutado_en TEXT,
		activa INTEGER DEFAULT 1,
		validacion_consistencia INTEGER DEFAULT 1,
		hash_ultima_ejecucion TEXT,
		fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		fecha_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`
	}
	if _, err := dbConn.Exec(createProgramaciones); err != nil {
		return fmt.Errorf("create empresa_reportes_programaciones: %w", err)
	}

	createPlantillas := `CREATE TABLE IF NOT EXISTS empresa_reportes_plantillas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER NOT NULL,
		codigo TEXT NOT NULL,
		nombre TEXT NOT NULL,
		dataset_key TEXT NOT NULL,
		version INTEGER NOT NULL DEFAULT 1,
		formato TEXT DEFAULT 'json',
		columnas_json TEXT DEFAULT '[]',
		config_json TEXT DEFAULT '{}',
		vigente INTEGER DEFAULT 1,
		hash_contenido TEXT,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`
	if isPostgres {
		createPlantillas = `CREATE TABLE IF NOT EXISTS empresa_reportes_plantillas (
		id SERIAL PRIMARY KEY,
		empresa_id INTEGER NOT NULL,
		codigo TEXT NOT NULL,
		nombre TEXT NOT NULL,
		dataset_key TEXT NOT NULL,
		version INTEGER NOT NULL DEFAULT 1,
		formato TEXT DEFAULT 'json',
		columnas_json TEXT DEFAULT '[]',
		config_json TEXT DEFAULT '{}',
		vigente INTEGER DEFAULT 1,
		hash_contenido TEXT,
		fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		fecha_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`
	}
	if _, err := dbConn.Exec(createPlantillas); err != nil {
		return fmt.Errorf("create empresa_reportes_plantillas: %w", err)
	}

	createEjecuciones := `CREATE TABLE IF NOT EXISTS empresa_reportes_ejecuciones (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER NOT NULL,
		programacion_id INTEGER,
		dataset_key TEXT NOT NULL,
		referencia TEXT,
		formato_principal TEXT DEFAULT 'json',
		formatos_json TEXT DEFAULT '["json"]',
		estado_ejecucion TEXT DEFAULT 'completado',
		ejecutado_en TEXT DEFAULT (datetime('now','localtime')),
		consistencia_estado TEXT DEFAULT 'pendiente',
		consistencia_detalle_json TEXT,
		salida_resumen_json TEXT,
		error_detalle TEXT,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`
	if isPostgres {
		createEjecuciones = `CREATE TABLE IF NOT EXISTS empresa_reportes_ejecuciones (
		id SERIAL PRIMARY KEY,
		empresa_id INTEGER NOT NULL,
		programacion_id INTEGER,
		dataset_key TEXT NOT NULL,
		referencia TEXT,
		formato_principal TEXT DEFAULT 'json',
		formatos_json TEXT DEFAULT '["json"]',
		estado_ejecucion TEXT DEFAULT 'completado',
		ejecutado_en TEXT DEFAULT CURRENT_TIMESTAMP,
		consistencia_estado TEXT DEFAULT 'pendiente',
		consistencia_detalle_json TEXT,
		salida_resumen_json TEXT,
		error_detalle TEXT,
		fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		fecha_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`
	}
	if _, err := dbConn.Exec(createEjecuciones); err != nil {
		return fmt.Errorf("create empresa_reportes_ejecuciones: %w", err)
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_programaciones", "empresa_id", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_programaciones.empresa_id: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_programaciones", "dataset_key", "TEXT"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_programaciones.dataset_key: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_programaciones", "frecuencia", "TEXT DEFAULT 'diario'"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_programaciones.frecuencia: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_programaciones", "hora_envio", "TEXT DEFAULT '08:00'"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_programaciones.hora_envio: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_programaciones", "timezone", "TEXT DEFAULT 'America/Bogota'"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_programaciones.timezone: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_programaciones", "formatos", "TEXT DEFAULT '[\"json\"]'"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_programaciones.formatos: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_programaciones", "parametros_json", "TEXT DEFAULT '{}'"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_programaciones.parametros_json: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_programaciones", "template_codigo", "TEXT"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_programaciones.template_codigo: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_programaciones", "template_version", "INTEGER DEFAULT 0"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_programaciones.template_version: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_programaciones", "ultimo_ejecutado_en", "TEXT"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_programaciones.ultimo_ejecutado_en: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_programaciones", "proximo_ejecutado_en", "TEXT"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_programaciones.proximo_ejecutado_en: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_programaciones", "activa", "INTEGER DEFAULT 1"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_programaciones.activa: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_programaciones", "validacion_consistencia", "INTEGER DEFAULT 1"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_programaciones.validacion_consistencia: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_programaciones", "hash_ultima_ejecucion", "TEXT"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_programaciones.hash_ultima_ejecucion: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_programaciones", "fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_programaciones.fecha_actualizacion: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_programaciones", "usuario_creador", "TEXT"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_programaciones.usuario_creador: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_programaciones", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_programaciones.estado: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_programaciones", "observaciones", "TEXT"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_programaciones.observaciones: %w", err)
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_plantillas", "empresa_id", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_plantillas.empresa_id: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_plantillas", "codigo", "TEXT"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_plantillas.codigo: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_plantillas", "dataset_key", "TEXT"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_plantillas.dataset_key: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_plantillas", "version", "INTEGER DEFAULT 1"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_plantillas.version: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_plantillas", "formato", "TEXT DEFAULT 'json'"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_plantillas.formato: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_plantillas", "columnas_json", "TEXT DEFAULT '[]'"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_plantillas.columnas_json: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_plantillas", "config_json", "TEXT DEFAULT '{}'"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_plantillas.config_json: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_plantillas", "vigente", "INTEGER DEFAULT 1"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_plantillas.vigente: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_plantillas", "hash_contenido", "TEXT"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_plantillas.hash_contenido: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_plantillas", "fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_plantillas.fecha_actualizacion: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_plantillas", "usuario_creador", "TEXT"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_plantillas.usuario_creador: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_plantillas", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_plantillas.estado: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_plantillas", "observaciones", "TEXT"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_plantillas.observaciones: %w", err)
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_ejecuciones", "empresa_id", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_ejecuciones.empresa_id: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_ejecuciones", "programacion_id", "INTEGER"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_ejecuciones.programacion_id: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_ejecuciones", "dataset_key", "TEXT"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_ejecuciones.dataset_key: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_ejecuciones", "referencia", "TEXT"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_ejecuciones.referencia: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_ejecuciones", "formato_principal", "TEXT DEFAULT 'json'"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_ejecuciones.formato_principal: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_ejecuciones", "formatos_json", "TEXT DEFAULT '[\"json\"]'"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_ejecuciones.formatos_json: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_ejecuciones", "estado_ejecucion", "TEXT DEFAULT 'completado'"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_ejecuciones.estado_ejecucion: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_ejecuciones", "ejecutado_en", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_ejecuciones.ejecutado_en: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_ejecuciones", "consistencia_estado", "TEXT DEFAULT 'pendiente'"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_ejecuciones.consistencia_estado: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_ejecuciones", "consistencia_detalle_json", "TEXT"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_ejecuciones.consistencia_detalle_json: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_ejecuciones", "salida_resumen_json", "TEXT"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_ejecuciones.salida_resumen_json: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_ejecuciones", "error_detalle", "TEXT"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_ejecuciones.error_detalle: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_ejecuciones", "fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_ejecuciones.fecha_actualizacion: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_ejecuciones", "usuario_creador", "TEXT"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_ejecuciones.usuario_creador: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_ejecuciones", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_ejecuciones.estado: %w", err)
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_reportes_ejecuciones", "observaciones", "TEXT"); err != nil {
		return fmt.Errorf("ensure empresa_reportes_ejecuciones.observaciones: %w", err)
	}

	if _, err := dbConn.Exec(`CREATE INDEX IF NOT EXISTS idx_reportes_programaciones_empresa_proximo
		ON empresa_reportes_programaciones(empresa_id, activa, proximo_ejecutado_en)`); err != nil {
		return fmt.Errorf("create idx_reportes_programaciones_empresa_proximo: %w", err)
	}
	if _, err := dbConn.Exec(`CREATE INDEX IF NOT EXISTS idx_reportes_programaciones_empresa_dataset
		ON empresa_reportes_programaciones(empresa_id, dataset_key)`); err != nil {
		return fmt.Errorf("create idx_reportes_programaciones_empresa_dataset: %w", err)
	}
	if _, err := dbConn.Exec(`CREATE INDEX IF NOT EXISTS idx_reportes_plantillas_empresa_codigo
		ON empresa_reportes_plantillas(empresa_id, codigo, version DESC)`); err != nil {
		return fmt.Errorf("create idx_reportes_plantillas_empresa_codigo: %w", err)
	}
	if _, err := dbConn.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_reportes_plantillas_empresa_codigo_version
		ON empresa_reportes_plantillas(empresa_id, codigo, version)`); err != nil {
		return fmt.Errorf("create ux_reportes_plantillas_empresa_codigo_version: %w", err)
	}
	if _, err := dbConn.Exec(`CREATE INDEX IF NOT EXISTS idx_reportes_ejecuciones_empresa_programacion
		ON empresa_reportes_ejecuciones(empresa_id, programacion_id, ejecutado_en DESC)`); err != nil {
		return fmt.Errorf("create idx_reportes_ejecuciones_empresa_programacion: %w", err)
	}
	if _, err := dbConn.Exec(`CREATE INDEX IF NOT EXISTS idx_reportes_ejecuciones_empresa_dataset
		ON empresa_reportes_ejecuciones(empresa_id, dataset_key, ejecutado_en DESC)`); err != nil {
		return fmt.Errorf("create idx_reportes_ejecuciones_empresa_dataset: %w", err)
	}

	return nil
}
