package db

import (
	"database/sql"
	"fmt"
	"time"
)

type PublicacionRedSocial struct {
	ID            int       `json:"id"`
	EmpresaID     int       `json:"empresa_id"`
	Nombre        string    `json:"nombre"`
	Descripcion   string    `json:"descripcion"`
	FotoURL       string    `json:"foto_url"`
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
		fecha_creacion DATETIME DEFAULT CURRENT_TIMESTAMP,
		estado TEXT DEFAULT 'activo'
	);`
	if isPostgresDialect() {
		query = `
		CREATE TABLE IF NOT EXISTS empresa_publicaciones_red_social (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			nombre TEXT NOT NULL,
			descripcion TEXT NOT NULL,
			foto_url TEXT,
			fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			estado TEXT DEFAULT 'activo'
		);`
	}
	_, err := execSQLCompat(db, query)
	if err != nil {
		return fmt.Errorf("error creando empresa_publicaciones_red_social: %v", err)
	}
	return nil
}

func GetPublicacionesRedSocialActivas(db *sql.DB) ([]PublicacionRedSocial, error) {
	if err := EnsureEmpresaPublicacionesRedSocialSchema(db); err != nil {
		return nil, err
	}
	query := `SELECT id, empresa_id, nombre, descripcion, COALESCE(foto_url,''), fecha_creacion, estado 
	          FROM empresa_publicaciones_red_social WHERE estado = 'activo' ORDER BY fecha_creacion DESC LIMIT 50`
	rows, err := querySQLCompat(db, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pubs []PublicacionRedSocial
	for rows.Next() {
		var p PublicacionRedSocial
		if err := rows.Scan(&p.ID, &p.EmpresaID, &p.Nombre, &p.Descripcion, &p.FotoURL, &p.FechaCreacion, &p.Estado); err != nil {
			return nil, err
		}
		pubs = append(pubs, p)
	}
	if pubs == nil {
		pubs = []PublicacionRedSocial{}
	}
	return pubs, nil
}

func GetPublicacionesRedSocialByEmpresa(db *sql.DB, empresaID int) ([]PublicacionRedSocial, error) {
	if err := EnsureEmpresaPublicacionesRedSocialSchema(db); err != nil {
		return nil, err
	}
	query := `SELECT id, empresa_id, nombre, descripcion, COALESCE(foto_url,''), fecha_creacion, estado 
	          FROM empresa_publicaciones_red_social WHERE empresa_id = ? ORDER BY fecha_creacion DESC`
	rows, err := querySQLCompat(db, query, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pubs []PublicacionRedSocial
	for rows.Next() {
		var p PublicacionRedSocial
		if err := rows.Scan(&p.ID, &p.EmpresaID, &p.Nombre, &p.Descripcion, &p.FotoURL, &p.FechaCreacion, &p.Estado); err != nil {
			return nil, err
		}
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
	query := `INSERT INTO empresa_publicaciones_red_social (empresa_id, nombre, descripcion, foto_url, estado) 
	          VALUES (?, ?, ?, ?, ?)`
	id, err := insertSQLCompat(db, query, p.EmpresaID, p.Nombre, p.Descripcion, p.FotoURL, p.Estado)
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
	query := `UPDATE empresa_publicaciones_red_social SET nombre=?, descripcion=?, foto_url=?, estado=? WHERE id=? AND empresa_id=?`
	_, err := execSQLCompat(db, query, p.Nombre, p.Descripcion, p.FotoURL, p.Estado, p.ID, p.EmpresaID)
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
