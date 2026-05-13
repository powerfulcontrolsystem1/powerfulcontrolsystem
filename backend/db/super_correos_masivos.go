package db

import (
	"database/sql"
	"fmt"
	"strings"
)

const SuperCorreoNotificacionTipoMasivoGlobal = "correo_masivo_global"

// SuperCorreoMasivo representa una campaña global enviada desde super administrador.
type SuperCorreoMasivo struct {
	ID                 int64  `json:"id"`
	Codigo             string `json:"codigo"`
	Categoria          string `json:"categoria"`
	Alcance            string `json:"alcance"`
	Asunto             string `json:"asunto"`
	CuerpoTexto        string `json:"cuerpo_texto,omitempty"`
	CuerpoHTML         string `json:"cuerpo_html,omitempty"`
	TotalDestinatarios int    `json:"total_destinatarios"`
	Enviados           int    `json:"enviados"`
	Fallidos           int    `json:"fallidos"`
	Omitidos           int    `json:"omitidos"`
	EstadoEnvio        string `json:"estado_envio"`
	ModoPrueba         int    `json:"modo_prueba"`
	MetadataJSON       string `json:"metadata_json,omitempty"`
	FechaEnvio         string `json:"fecha_envio,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// SuperCorreoMasivoDestinatario guarda el resultado individual de una campaña.
type SuperCorreoMasivoDestinatario struct {
	ID                 int64  `json:"id"`
	CorreoMasivoID     int64  `json:"correo_masivo_id"`
	Email              string `json:"email"`
	Nombre             string `json:"nombre,omitempty"`
	TipoDestinatario   string `json:"tipo_destinatario"`
	EmpresaID          int64  `json:"empresa_id,omitempty"`
	EmpresaNombre      string `json:"empresa_nombre,omitempty"`
	Rol                string `json:"rol,omitempty"`
	Resultado          string `json:"resultado"`
	ErrorDetalle       string `json:"error_detalle,omitempty"`
	FechaEnvio         string `json:"fecha_envio,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// EnsureSuperCorreosMasivosSchema crea o migra las tablas de correos globales.
func EnsureSuperCorreosMasivosSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("db connection is required")
	}

	campaignID := "INTEGER PRIMARY KEY AUTOINCREMENT"
	recipientID := "INTEGER PRIMARY KEY AUTOINCREMENT"
	fkID := "INTEGER"
	if isPostgresDialect() {
		campaignID = "BIGSERIAL PRIMARY KEY"
		recipientID = "BIGSERIAL PRIMARY KEY"
		fkID = "BIGINT"
	}

	stmts := []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS super_correos_masivos (
			id %s,
			codigo TEXT UNIQUE NOT NULL,
			categoria TEXT NOT NULL,
			alcance TEXT NOT NULL,
			asunto TEXT NOT NULL,
			cuerpo_texto TEXT,
			cuerpo_html TEXT,
			total_destinatarios INTEGER DEFAULT 0,
			enviados INTEGER DEFAULT 0,
			fallidos INTEGER DEFAULT 0,
			omitidos INTEGER DEFAULT 0,
			estado_envio TEXT DEFAULT 'pendiente',
			modo_prueba INTEGER DEFAULT 0,
			metadata_json TEXT,
			fecha_envio TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`, campaignID),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS super_correos_masivos_destinatarios (
			id %s,
			correo_masivo_id %s NOT NULL,
			email TEXT NOT NULL,
			nombre TEXT,
			tipo_destinatario TEXT NOT NULL,
			empresa_id %s,
			empresa_nombre TEXT,
			rol TEXT,
			resultado TEXT DEFAULT 'pendiente',
			error_detalle TEXT,
			fecha_envio TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`, recipientID, fkID, fkID),
		`CREATE INDEX IF NOT EXISTS ix_super_correos_masivos_fecha ON super_correos_masivos(fecha_creacion DESC)`,
		`CREATE INDEX IF NOT EXISTS ix_super_correos_masivos_estado ON super_correos_masivos(estado_envio, categoria)`,
		`CREATE INDEX IF NOT EXISTS ix_super_correos_masivos_destinatarios_campana ON super_correos_masivos_destinatarios(correo_masivo_id, resultado)`,
		`CREATE INDEX IF NOT EXISTS ix_super_correos_masivos_destinatarios_email ON super_correos_masivos_destinatarios(email)`,
	}
	for _, stmt := range stmts {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}

	columnsCampaign := []struct {
		name string
		def  string
	}{
		{"codigo", "TEXT"},
		{"categoria", "TEXT"},
		{"alcance", "TEXT"},
		{"asunto", "TEXT"},
		{"cuerpo_texto", "TEXT"},
		{"cuerpo_html", "TEXT"},
		{"total_destinatarios", "INTEGER DEFAULT 0"},
		{"enviados", "INTEGER DEFAULT 0"},
		{"fallidos", "INTEGER DEFAULT 0"},
		{"omitidos", "INTEGER DEFAULT 0"},
		{"estado_envio", "TEXT DEFAULT 'pendiente'"},
		{"modo_prueba", "INTEGER DEFAULT 0"},
		{"metadata_json", "TEXT"},
		{"fecha_envio", "TEXT"},
		{"fecha_creacion", "TEXT DEFAULT (datetime('now','localtime'))"},
		{"fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"},
		{"usuario_creador", "TEXT"},
		{"estado", "TEXT DEFAULT 'activo'"},
		{"observaciones", "TEXT"},
	}
	for _, column := range columnsCampaign {
		if err := ensureColumnIfMissing(dbConn, "super_correos_masivos", column.name, column.def); err != nil {
			return err
		}
	}

	columnsRecipients := []struct {
		name string
		def  string
	}{
		{"correo_masivo_id", fkID},
		{"email", "TEXT"},
		{"nombre", "TEXT"},
		{"tipo_destinatario", "TEXT"},
		{"empresa_id", fkID},
		{"empresa_nombre", "TEXT"},
		{"rol", "TEXT"},
		{"resultado", "TEXT DEFAULT 'pendiente'"},
		{"error_detalle", "TEXT"},
		{"fecha_envio", "TEXT"},
		{"fecha_creacion", "TEXT DEFAULT (datetime('now','localtime'))"},
		{"fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"},
		{"usuario_creador", "TEXT"},
		{"estado", "TEXT DEFAULT 'activo'"},
		{"observaciones", "TEXT"},
	}
	for _, column := range columnsRecipients {
		if err := ensureColumnIfMissing(dbConn, "super_correos_masivos_destinatarios", column.name, column.def); err != nil {
			return err
		}
	}

	return nil
}

func CreateSuperCorreoMasivo(dbConn *sql.DB, item SuperCorreoMasivo) (int64, error) {
	if err := EnsureSuperCorreosMasivosSchema(dbConn); err != nil {
		return 0, err
	}
	item.Codigo = strings.TrimSpace(item.Codigo)
	item.Categoria = strings.TrimSpace(item.Categoria)
	item.Alcance = strings.TrimSpace(item.Alcance)
	item.Asunto = strings.TrimSpace(item.Asunto)
	item.UsuarioCreador = strings.TrimSpace(item.UsuarioCreador)
	if item.Codigo == "" {
		return 0, fmt.Errorf("codigo is required")
	}
	if item.Categoria == "" {
		item.Categoria = "informacion"
	}
	if item.Alcance == "" {
		item.Alcance = "todos"
	}
	if item.EstadoEnvio == "" {
		item.EstadoEnvio = "pendiente"
	}
	if item.Estado == "" {
		item.Estado = "activo"
	}

	return insertSQLCompat(dbConn, `INSERT INTO super_correos_masivos (
		codigo, categoria, alcance, asunto, cuerpo_texto, cuerpo_html,
		total_destinatarios, enviados, fallidos, omitidos, estado_envio, modo_prueba,
		metadata_json, fecha_envio, usuario_creador, estado, observaciones,
		fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
		item.Codigo,
		item.Categoria,
		item.Alcance,
		item.Asunto,
		item.CuerpoTexto,
		item.CuerpoHTML,
		item.TotalDestinatarios,
		item.Enviados,
		item.Fallidos,
		item.Omitidos,
		item.EstadoEnvio,
		item.ModoPrueba,
		item.MetadataJSON,
		item.FechaEnvio,
		item.UsuarioCreador,
		item.Estado,
		item.Observaciones,
	)
}

func CreateSuperCorreoMasivoDestinatario(dbConn *sql.DB, item SuperCorreoMasivoDestinatario) (int64, error) {
	if err := EnsureSuperCorreosMasivosSchema(dbConn); err != nil {
		return 0, err
	}
	item.Email = strings.TrimSpace(item.Email)
	item.Nombre = strings.TrimSpace(item.Nombre)
	item.TipoDestinatario = strings.TrimSpace(item.TipoDestinatario)
	item.Resultado = strings.TrimSpace(item.Resultado)
	if item.Email == "" {
		return 0, fmt.Errorf("email is required")
	}
	if item.TipoDestinatario == "" {
		item.TipoDestinatario = "usuario"
	}
	if item.Resultado == "" {
		item.Resultado = "pendiente"
	}
	if item.Estado == "" {
		item.Estado = "activo"
	}

	return insertSQLCompat(dbConn, `INSERT INTO super_correos_masivos_destinatarios (
		correo_masivo_id, email, nombre, tipo_destinatario, empresa_id, empresa_nombre,
		rol, resultado, error_detalle, fecha_envio, usuario_creador, estado, observaciones,
		fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
		item.CorreoMasivoID,
		item.Email,
		item.Nombre,
		item.TipoDestinatario,
		item.EmpresaID,
		item.EmpresaNombre,
		item.Rol,
		item.Resultado,
		item.ErrorDetalle,
		item.FechaEnvio,
		item.UsuarioCreador,
		item.Estado,
		item.Observaciones,
	)
}

func UpdateSuperCorreoMasivoResultado(dbConn *sql.DB, id int64, enviados, fallidos, omitidos int, estadoEnvio, observaciones string) error {
	if dbConn == nil {
		return fmt.Errorf("db connection is required")
	}
	estadoEnvio = strings.TrimSpace(estadoEnvio)
	if estadoEnvio == "" {
		estadoEnvio = "enviado"
	}
	_, err := execSQLCompat(dbConn, `UPDATE super_correos_masivos
		SET enviados = ?, fallidos = ?, omitidos = ?, estado_envio = ?, observaciones = ?,
			fecha_envio = datetime('now','localtime'), fecha_actualizacion = datetime('now','localtime')
		WHERE id = ?`,
		enviados, fallidos, omitidos, estadoEnvio, strings.TrimSpace(observaciones), id,
	)
	return err
}

func ListSuperCorreosMasivos(dbConn *sql.DB, limit int) ([]SuperCorreoMasivo, error) {
	if err := EnsureSuperCorreosMasivosSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		id, codigo, categoria, alcance, asunto,
		COALESCE(total_destinatarios, 0), COALESCE(enviados, 0), COALESCE(fallidos, 0), COALESCE(omitidos, 0),
		COALESCE(estado_envio, ''), COALESCE(modo_prueba, 0), COALESCE(fecha_envio, ''),
		COALESCE(fecha_creacion, ''), COALESCE(fecha_actualizacion, ''), COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'), COALESCE(observaciones, '')
	FROM super_correos_masivos
	ORDER BY id DESC
	LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]SuperCorreoMasivo, 0)
	for rows.Next() {
		var item SuperCorreoMasivo
		if err := rows.Scan(
			&item.ID,
			&item.Codigo,
			&item.Categoria,
			&item.Alcance,
			&item.Asunto,
			&item.TotalDestinatarios,
			&item.Enviados,
			&item.Fallidos,
			&item.Omitidos,
			&item.EstadoEnvio,
			&item.ModoPrueba,
			&item.FechaEnvio,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
