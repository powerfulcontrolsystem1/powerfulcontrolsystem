package db

import (
	"database/sql"
	"strings"
)

type SuperVPSSnapshotLog struct {
	ID                 int64  `json:"id"`
	Codigo             string `json:"codigo"`
	FileName           string `json:"file_name"`
	FilePath           string `json:"file_path,omitempty"`
	Estado             string `json:"estado"`
	TamanoBytes        int64  `json:"tamano_bytes"`
	HashSHA256         string `json:"hash_sha256,omitempty"`
	CloudProvider      string `json:"cloud_provider,omitempty"`
	CloudDestino       string `json:"cloud_destino,omitempty"`
	CloudEstado        string `json:"cloud_estado,omitempty"`
	CloudMensaje       string `json:"cloud_mensaje,omitempty"`
	Automatico         int    `json:"automatico"`
	IncluyeSecretos    int    `json:"incluye_secretos"`
	IncluyeImagenes    int    `json:"incluye_imagenes"`
	ManifestJSON       string `json:"manifest_json,omitempty"`
	Error              string `json:"error,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	FechaInicio        string `json:"fecha_inicio,omitempty"`
	FechaFin           string `json:"fecha_fin,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
}

func EnsureSuperVPSSnapshotSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return nil
	}
	statements := []string{
		`CREATE TABLE IF NOT EXISTS super_vps_snapshots (
			id BIGSERIAL PRIMARY KEY,
			codigo TEXT NOT NULL,
			file_name TEXT,
			file_path TEXT,
			estado TEXT DEFAULT 'pendiente',
			tamano_bytes BIGINT DEFAULT 0,
			hash_sha256 TEXT,
			cloud_provider TEXT,
			cloud_destino TEXT,
			cloud_estado TEXT,
			cloud_mensaje TEXT,
			automatico INTEGER DEFAULT 0,
			incluye_secretos INTEGER DEFAULT 0,
			incluye_imagenes INTEGER DEFAULT 0,
			manifest_json TEXT,
			error TEXT,
			usuario_creador TEXT,
			fecha_inicio TEXT,
			fecha_fin TEXT,
			fecha_creacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT),
			fecha_actualizacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT)
		)`,
		`ALTER TABLE super_vps_snapshots ADD COLUMN IF NOT EXISTS codigo TEXT`,
		`ALTER TABLE super_vps_snapshots ADD COLUMN IF NOT EXISTS file_name TEXT`,
		`ALTER TABLE super_vps_snapshots ADD COLUMN IF NOT EXISTS file_path TEXT`,
		`ALTER TABLE super_vps_snapshots ADD COLUMN IF NOT EXISTS estado TEXT DEFAULT 'pendiente'`,
		`ALTER TABLE super_vps_snapshots ADD COLUMN IF NOT EXISTS tamano_bytes BIGINT DEFAULT 0`,
		`ALTER TABLE super_vps_snapshots ADD COLUMN IF NOT EXISTS hash_sha256 TEXT`,
		`ALTER TABLE super_vps_snapshots ADD COLUMN IF NOT EXISTS cloud_provider TEXT`,
		`ALTER TABLE super_vps_snapshots ADD COLUMN IF NOT EXISTS cloud_destino TEXT`,
		`ALTER TABLE super_vps_snapshots ADD COLUMN IF NOT EXISTS cloud_estado TEXT`,
		`ALTER TABLE super_vps_snapshots ADD COLUMN IF NOT EXISTS cloud_mensaje TEXT`,
		`ALTER TABLE super_vps_snapshots ADD COLUMN IF NOT EXISTS automatico INTEGER DEFAULT 0`,
		`ALTER TABLE super_vps_snapshots ADD COLUMN IF NOT EXISTS incluye_secretos INTEGER DEFAULT 0`,
		`ALTER TABLE super_vps_snapshots ADD COLUMN IF NOT EXISTS incluye_imagenes INTEGER DEFAULT 0`,
		`ALTER TABLE super_vps_snapshots ADD COLUMN IF NOT EXISTS manifest_json TEXT`,
		`ALTER TABLE super_vps_snapshots ADD COLUMN IF NOT EXISTS error TEXT`,
		`ALTER TABLE super_vps_snapshots ADD COLUMN IF NOT EXISTS usuario_creador TEXT`,
		`ALTER TABLE super_vps_snapshots ADD COLUMN IF NOT EXISTS fecha_inicio TEXT`,
		`ALTER TABLE super_vps_snapshots ADD COLUMN IF NOT EXISTS fecha_fin TEXT`,
		`ALTER TABLE super_vps_snapshots ADD COLUMN IF NOT EXISTS fecha_creacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT)`,
		`ALTER TABLE super_vps_snapshots ADD COLUMN IF NOT EXISTS fecha_actualizacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_super_vps_snapshots_codigo ON super_vps_snapshots(codigo)`,
		`CREATE INDEX IF NOT EXISTS ix_super_vps_snapshots_estado ON super_vps_snapshots(estado, fecha_creacion DESC)`,
	}
	for _, stmt := range statements {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func InsertSuperVPSSnapshotLog(dbConn *sql.DB, item SuperVPSSnapshotLog) (int64, error) {
	if dbConn == nil {
		return 0, sql.ErrConnDone
	}
	if err := EnsureSuperVPSSnapshotSchema(dbConn); err != nil {
		return 0, err
	}
	nowExpr := sqlNowExpr()
	return insertSQLCompat(dbConn, `INSERT INTO super_vps_snapshots (
		codigo, file_name, file_path, estado, tamano_bytes, hash_sha256, cloud_provider, cloud_destino, cloud_estado, cloud_mensaje, automatico, incluye_secretos, incluye_imagenes, manifest_json, error, usuario_creador, fecha_inicio, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CAST(`+nowExpr+` AS TEXT), CAST(`+nowExpr+` AS TEXT), CAST(`+nowExpr+` AS TEXT))`,
		strings.TrimSpace(item.Codigo), strings.TrimSpace(item.FileName), strings.TrimSpace(item.FilePath), normalizeSuperVPSSnapshotEstado(item.Estado), item.TamanoBytes, strings.TrimSpace(item.HashSHA256), strings.TrimSpace(item.CloudProvider), strings.TrimSpace(item.CloudDestino), strings.TrimSpace(item.CloudEstado), strings.TrimSpace(item.CloudMensaje), item.Automatico, item.IncluyeSecretos, item.IncluyeImagenes, strings.TrimSpace(item.ManifestJSON), truncateSuperVPSSnapshotText(item.Error, 1200), strings.TrimSpace(item.UsuarioCreador))
}

func UpdateSuperVPSSnapshotLog(dbConn *sql.DB, item SuperVPSSnapshotLog) error {
	if dbConn == nil {
		return nil
	}
	if item.ID <= 0 {
		return nil
	}
	if err := EnsureSuperVPSSnapshotSchema(dbConn); err != nil {
		return err
	}
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, `UPDATE super_vps_snapshots
		SET file_name=?, file_path=?, estado=?, tamano_bytes=?, hash_sha256=?, cloud_provider=?, cloud_destino=?, cloud_estado=?, cloud_mensaje=?, automatico=?, incluye_secretos=?, incluye_imagenes=?, manifest_json=?, error=?, fecha_fin=CASE WHEN ? IN ('completado','error') THEN CAST(`+nowExpr+` AS TEXT) ELSE fecha_fin END, fecha_actualizacion=CAST(`+nowExpr+` AS TEXT)
		WHERE id=?`,
		strings.TrimSpace(item.FileName), strings.TrimSpace(item.FilePath), normalizeSuperVPSSnapshotEstado(item.Estado), item.TamanoBytes, strings.TrimSpace(item.HashSHA256), strings.TrimSpace(item.CloudProvider), strings.TrimSpace(item.CloudDestino), strings.TrimSpace(item.CloudEstado), strings.TrimSpace(item.CloudMensaje), item.Automatico, item.IncluyeSecretos, item.IncluyeImagenes, strings.TrimSpace(item.ManifestJSON), truncateSuperVPSSnapshotText(item.Error, 1200), normalizeSuperVPSSnapshotEstado(item.Estado), item.ID)
	return err
}

func GetSuperVPSSnapshotLog(dbConn *sql.DB, id int64) (*SuperVPSSnapshotLog, error) {
	if dbConn == nil || id <= 0 {
		return nil, sql.ErrNoRows
	}
	if err := EnsureSuperVPSSnapshotSchema(dbConn); err != nil {
		return nil, err
	}
	var item SuperVPSSnapshotLog
	err := queryRowSQLCompat(dbConn, `SELECT id, COALESCE(codigo,''), COALESCE(file_name,''), COALESCE(file_path,''), COALESCE(estado,''), COALESCE(tamano_bytes,0), COALESCE(hash_sha256,''), COALESCE(cloud_provider,''), COALESCE(cloud_destino,''), COALESCE(cloud_estado,''), COALESCE(cloud_mensaje,''), COALESCE(automatico,0), COALESCE(incluye_secretos,0), COALESCE(incluye_imagenes,0), COALESCE(manifest_json,''), COALESCE(error,''), COALESCE(usuario_creador,''), COALESCE(fecha_inicio,''), COALESCE(fecha_fin,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,'')
		FROM super_vps_snapshots WHERE id=? LIMIT 1`, id).Scan(&item.ID, &item.Codigo, &item.FileName, &item.FilePath, &item.Estado, &item.TamanoBytes, &item.HashSHA256, &item.CloudProvider, &item.CloudDestino, &item.CloudEstado, &item.CloudMensaje, &item.Automatico, &item.IncluyeSecretos, &item.IncluyeImagenes, &item.ManifestJSON, &item.Error, &item.UsuarioCreador, &item.FechaInicio, &item.FechaFin, &item.FechaCreacion, &item.FechaActualizacion)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func ListSuperVPSSnapshotLogs(dbConn *sql.DB, limit int) ([]SuperVPSSnapshotLog, error) {
	if dbConn == nil {
		return []SuperVPSSnapshotLog{}, nil
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if err := EnsureSuperVPSSnapshotSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT id, COALESCE(codigo,''), COALESCE(file_name,''), COALESCE(file_path,''), COALESCE(estado,''), COALESCE(tamano_bytes,0), COALESCE(hash_sha256,''), COALESCE(cloud_provider,''), COALESCE(cloud_destino,''), COALESCE(cloud_estado,''), COALESCE(cloud_mensaje,''), COALESCE(automatico,0), COALESCE(incluye_secretos,0), COALESCE(incluye_imagenes,0), COALESCE(manifest_json,''), COALESCE(error,''), COALESCE(usuario_creador,''), COALESCE(fecha_inicio,''), COALESCE(fecha_fin,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,'')
		FROM super_vps_snapshots ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]SuperVPSSnapshotLog, 0)
	for rows.Next() {
		var item SuperVPSSnapshotLog
		if err := rows.Scan(&item.ID, &item.Codigo, &item.FileName, &item.FilePath, &item.Estado, &item.TamanoBytes, &item.HashSHA256, &item.CloudProvider, &item.CloudDestino, &item.CloudEstado, &item.CloudMensaje, &item.Automatico, &item.IncluyeSecretos, &item.IncluyeImagenes, &item.ManifestJSON, &item.Error, &item.UsuarioCreador, &item.FechaInicio, &item.FechaFin, &item.FechaCreacion, &item.FechaActualizacion); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func normalizeSuperVPSSnapshotEstado(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "pendiente", "en_proceso", "completado", "error":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return "pendiente"
	}
}

func truncateSuperVPSSnapshotText(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 || len(value) <= max {
		return value
	}
	return value[:max]
}
