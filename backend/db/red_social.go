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
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("error creando empresa_publicaciones_red_social: %v", err)
	}
	return nil
}

func GetPublicacionesRedSocialActivas(db *sql.DB) ([]PublicacionRedSocial, error) {
	query := `SELECT id, empresa_id, nombre, descripcion, COALESCE(foto_url,''), fecha_creacion, estado 
	          FROM empresa_publicaciones_red_social WHERE estado = 'activo' ORDER BY fecha_creacion DESC LIMIT 50`
	rows, err := db.Query(query)
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
	query := `SELECT id, empresa_id, nombre, descripcion, COALESCE(foto_url,''), fecha_creacion, estado 
	          FROM empresa_publicaciones_red_social WHERE empresa_id = ? ORDER BY fecha_creacion DESC`
	rows, err := db.Query(query, empresaID)
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
	query := `INSERT INTO empresa_publicaciones_red_social (empresa_id, nombre, descripcion, foto_url, estado) 
	          VALUES (?, ?, ?, ?, ?)`
	res, err := db.Exec(query, p.EmpresaID, p.Nombre, p.Descripcion, p.FotoURL, p.Estado)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	p.ID = int(id)
	return nil
}

func UpdatePublicacionRedSocial(db *sql.DB, p *PublicacionRedSocial) error {
	query := `UPDATE empresa_publicaciones_red_social SET nombre=?, descripcion=?, foto_url=?, estado=? WHERE id=? AND empresa_id=?`
	_, err := db.Exec(query, p.Nombre, p.Descripcion, p.FotoURL, p.Estado, p.ID, p.EmpresaID)
	return err
}

func DeletePublicacionRedSocial(db *sql.DB, id, empresaID int) error {
	query := `DELETE FROM empresa_publicaciones_red_social WHERE id=? AND empresa_id=?`
	_, err := db.Exec(query, id, empresaID)
	return err
}
