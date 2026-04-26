package db

import (
	"database/sql"
	"fmt"
	"time"
)

type PublicacionRedSocial struct {
	ID            int       `json:"id"`
	EmpresaID     int       `json:"empresa_id"`
	EmpresaNombre string    `json:"empresa_nombre,omitempty"`
	Nombre        string    `json:"nombre"`
	Descripcion   string    `json:"descripcion"`
	FotoURL       string    `json:"foto_url"`
	YoutubeURL    string    `json:"youtube_url"`
	FechaCreacion time.Time `json:"fecha_creacion"`
	Estado        string    `json:"estado"`
}

func EnsureEmpresaPublicacionesRedSocialSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS empresa_publicaciones_red_social (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER NOT NULL,
		nombre TEXT NOT NULL,
		descripcion TEXT NOT NULL,
		foto_url TEXT,
		youtube_url TEXT,
		fecha_creacion DATETIME DEFAULT CURRENT_TIMESTAMP,
		estado TEXT DEFAULT 'activo'
	);`
	if shouldUsePostgresCompat(db) {
		query = `
		CREATE TABLE IF NOT EXISTS empresa_publicaciones_red_social (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			nombre TEXT NOT NULL,
			descripcion TEXT NOT NULL,
			foto_url TEXT,
			youtube_url TEXT,
			fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			estado TEXT DEFAULT 'activo'
		);`
	}
	_, err := execSQLCompat(db, query)
	if err != nil {
		return fmt.Errorf("error creando empresa_publicaciones_red_social: %v", err)
	}
	// Migración suave de columnas nuevas.
	_ = ensureColumnIfMissing(db, "empresa_publicaciones_red_social", "foto_url", "TEXT")
	_ = ensureColumnIfMissing(db, "empresa_publicaciones_red_social", "youtube_url", "TEXT")
	return nil
}

func clampInt(v, def, min, max int) int {
	if v <= 0 {
		v = def
	}
	if v < min {
		v = min
	}
	if v > max {
		v = max
	}
	return v
}

func GetPublicacionesRedSocialActivas(db *sql.DB, limit, offset int) ([]PublicacionRedSocial, error) {
	if err := EnsureEmpresaPublicacionesRedSocialSchema(db); err != nil {
		return nil, err
	}
	limit = clampInt(limit, 20, 1, 50)
	if offset < 0 {
		offset = 0
	}
	query := `SELECT p.id, p.empresa_id, COALESCE(e.nombre, ''), p.nombre, p.descripcion, COALESCE(p.foto_url,''), COALESCE(p.youtube_url,''), p.fecha_creacion, p.estado 
	          FROM empresa_publicaciones_red_social p
	          LEFT JOIN empresas e ON e.id = p.empresa_id OR COALESCE(e.empresa_id, 0) = p.empresa_id
	          WHERE p.estado = 'activo' ORDER BY p.fecha_creacion DESC LIMIT ? OFFSET ?`
	rows, err := querySQLCompat(db, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pubs []PublicacionRedSocial
	for rows.Next() {
		var p PublicacionRedSocial
		var youtube string
		if err := rows.Scan(&p.ID, &p.EmpresaID, &p.EmpresaNombre, &p.Nombre, &p.Descripcion, &p.FotoURL, &youtube, &p.FechaCreacion, &p.Estado); err != nil {
			return nil, err
		}
		p.YoutubeURL = youtube
		pubs = append(pubs, p)
	}
	if pubs == nil {
		pubs = []PublicacionRedSocial{}
	}
	return pubs, nil
}

func GetPublicacionesRedSocialByEmpresa(db *sql.DB, empresaID int, limit, offset int) ([]PublicacionRedSocial, error) {
	if err := EnsureEmpresaPublicacionesRedSocialSchema(db); err != nil {
		return nil, err
	}
	limit = clampInt(limit, 50, 1, 200)
	if offset < 0 {
		offset = 0
	}
	query := `SELECT p.id, p.empresa_id, COALESCE(e.nombre, ''), p.nombre, p.descripcion, COALESCE(p.foto_url,''), COALESCE(p.youtube_url,''), p.fecha_creacion, p.estado 
	          FROM empresa_publicaciones_red_social p
	          LEFT JOIN empresas e ON e.id = p.empresa_id OR COALESCE(e.empresa_id, 0) = p.empresa_id
	          WHERE p.empresa_id = ? ORDER BY p.fecha_creacion DESC LIMIT ? OFFSET ?`
	rows, err := querySQLCompat(db, query, empresaID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pubs []PublicacionRedSocial
	for rows.Next() {
		var p PublicacionRedSocial
		var youtube string
		if err := rows.Scan(&p.ID, &p.EmpresaID, &p.EmpresaNombre, &p.Nombre, &p.Descripcion, &p.FotoURL, &youtube, &p.FechaCreacion, &p.Estado); err != nil {
			return nil, err
		}
		p.YoutubeURL = youtube
		pubs = append(pubs, p)
	}
	if pubs == nil {
		pubs = []PublicacionRedSocial{}
	}
	return pubs, nil
}

func InsertPublicacionRedSocial(db *sql.DB, p *PublicacionRedSocial) error {
	if err := EnsureEmpresaPublicacionesRedSocialSchema(db); err != nil {
		return err
	}
	query := `INSERT INTO empresa_publicaciones_red_social (empresa_id, nombre, descripcion, foto_url, youtube_url, estado)
	          VALUES (?, ?, ?, ?, ?, ?)`
	id, err := insertSQLCompat(db, query, p.EmpresaID, p.Nombre, p.Descripcion, p.FotoURL, p.YoutubeURL, p.Estado)
	if err != nil {
		return err
	}
	p.ID = int(id)
	return nil
}

func UpdatePublicacionRedSocial(db *sql.DB, p *PublicacionRedSocial) error {
	if err := EnsureEmpresaPublicacionesRedSocialSchema(db); err != nil {
		return err
	}
	query := `UPDATE empresa_publicaciones_red_social SET nombre=?, descripcion=?, foto_url=?, youtube_url=?, estado=? WHERE id=? AND empresa_id=?`
	_, err := execSQLCompat(db, query, p.Nombre, p.Descripcion, p.FotoURL, p.YoutubeURL, p.Estado, p.ID, p.EmpresaID)
	return err
}

func DeletePublicacionRedSocial(db *sql.DB, id, empresaID int) error {
	if err := EnsureEmpresaPublicacionesRedSocialSchema(db); err != nil {
		return err
	}
	query := `DELETE FROM empresa_publicaciones_red_social WHERE id=? AND empresa_id=?`
	_, err := execSQLCompat(db, query, id, empresaID)
	return err
}
