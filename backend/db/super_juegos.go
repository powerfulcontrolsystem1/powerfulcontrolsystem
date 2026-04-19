package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

type SuperJuegoRecord struct {
	ID            int    `json:"id"`
	Juego         string `json:"juego"`
	NombreJugador string `json:"nombre_jugador"`
	EmpresaID     string `json:"empresa_id"` // "Publico" or an actual ID as string
	Puntaje       int    `json:"puntaje"`
	Nivel         int    `json:"nivel"`
	Fecha         string `json:"fecha_creacion"`
}

func EnsureSuperJuegosSchema(db *sql.DB) error {
	createTable := `
	CREATE TABLE IF NOT EXISTS super_juegos_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		juego TEXT NOT NULL,
		nombre_jugador TEXT NOT NULL,
		empresa_id TEXT DEFAULT 'Publico',
		puntaje INTEGER DEFAULT 0,
		nivel INTEGER DEFAULT 1,
		fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
		fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`

	// En PostgreSQL CURRENT_TIMESTAMP es estándar, al igual que en SQLite.
	// Haremos el switch genérico para la tabla y secuencias si PostgreSQL lo requiere (AUTOINCREMENT -> SERIAL)
	isPostgres := false
	if err := db.QueryRow("SELECT 1 FROM pg_class LIMIT 1").Scan(new(int)); err == nil {
		isPostgres = true
	}

	if isPostgres {
		createTable = `
		CREATE TABLE IF NOT EXISTS super_juegos_records (
			id SERIAL PRIMARY KEY,
			juego TEXT NOT NULL,
			nombre_jugador TEXT NOT NULL,
			empresa_id TEXT DEFAULT 'Publico',
			puntaje INTEGER DEFAULT 0,
			nivel INTEGER DEFAULT 1,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`
	}

	if _, err := db.Exec(createTable); err != nil {
		return fmt.Errorf("failed to create super_juegos_records table: %w", err)
	}

	// Helpers para evolución de esquema (en caso de campos futuros, usando SQLite/Postgres handlers)
	addIfMissing := func(colDef string, name string) {
		var exists bool
		if isPostgres {
			q := "SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name='super_juegos_records' AND column_name=$1);"
			_ = db.QueryRow(q, name).Scan(&exists)
		} else {
			rows, err := db.Query("PRAGMA table_info(super_juegos_records);")
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var cid int
					var cname, ctype string
					var notnull int
					var dflt sql.NullString
					var pk int
					if rows.Scan(&cid, &cname, &ctype, &notnull, &dflt, &pk) == nil && cname == name {
						exists = true
						break
					}
				}
			}
		}

		if !exists {
			// En postgres IF NOT EXISTS es seguro para agregar columnas
			q := fmt.Sprintf("ALTER TABLE super_juegos_records ADD COLUMN IF NOT EXISTS %s;", colDef)
			if !isPostgres {
				q = fmt.Sprintf("ALTER TABLE super_juegos_records ADD COLUMN %s;", colDef)
			}
			if _, err := db.Exec(q); err != nil {
				if !strings.Contains(err.Error(), "already exists") && !strings.Contains(err.Error(), "duplicate column") {
					log.Printf("failed to add column %s to super_juegos_records: %v", name, err)
				}
			} else {
				log.Printf("added column %s to super_juegos_records", name)
			}
		}
	}

	addIfMissing("fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP", "fecha_actualizacion")
	addIfMissing("usuario_creador TEXT", "usuario_creador")
	addIfMissing("estado TEXT DEFAULT 'activo'", "estado")
	addIfMissing("observaciones TEXT", "observaciones")

	// Crear índices para optimizar los tops por juego
	if isPostgres {
		db.Exec("CREATE INDEX IF NOT EXISTS idx_super_juegos_records_top ON super_juegos_records(juego, puntaje DESC, fecha_creacion ASC);")
	} else {
		db.Exec("CREATE INDEX IF NOT EXISTS idx_super_juegos_records_top ON super_juegos_records(juego, puntaje DESC, fecha_creacion ASC);")
	}

	return nil
}

func SaveSuperJuegoRecord(db *sql.DB, rec SuperJuegoRecord) (int64, error) {
	q := `INSERT INTO super_juegos_records (juego, nombre_jugador, empresa_id, puntaje, nivel)
	      VALUES ($1, $2, $3, $4, $5) RETURNING id`

	isPostgres := false
	if err := db.QueryRow("SELECT 1 FROM pg_class LIMIT 1").Scan(new(int)); err == nil {
		isPostgres = true
	}

	var id int64
	if isPostgres {
		err := db.QueryRow(q, rec.Juego, rec.NombreJugador, rec.EmpresaID, rec.Puntaje, rec.Nivel).Scan(&id)
		if err != nil {
			return 0, err
		}
		return id, nil
	} else {
		qLite := `INSERT INTO super_juegos_records (juego, nombre_jugador, empresa_id, puntaje, nivel)
	      VALUES (?, ?, ?, ?, ?)`
		res, err := db.Exec(qLite, rec.Juego, rec.NombreJugador, rec.EmpresaID, rec.Puntaje, rec.Nivel)
		if err != nil {
			return 0, err
		}
		return res.LastInsertId()
	}
}

func GetTopSuperJuegoRecords(db *sql.DB, juego string, limit int) ([]SuperJuegoRecord, error) {
	q := `SELECT id, juego, nombre_jugador, empresa_id, puntaje, nivel, fecha_creacion
	      FROM super_juegos_records
	      WHERE juego = $1 AND estado = 'activo'
	      ORDER BY puntaje DESC, fecha_creacion ASC
	      LIMIT $2`

	isPostgres := false
	if err := db.QueryRow("SELECT 1 FROM pg_class LIMIT 1").Scan(new(int)); err == nil {
		isPostgres = true
	}

	var rows *sql.Rows
	var err error

	if isPostgres {
		rows, err = db.Query(q, juego, limit)
	} else {
		qLite := `SELECT id, juego, nombre_jugador, empresa_id, puntaje, nivel, fecha_creacion
	      FROM super_juegos_records
	      WHERE juego = ? AND estado = 'activo'
	      ORDER BY puntaje DESC, fecha_creacion ASC
	      LIMIT ?`
		rows, err = db.Query(qLite, juego, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tops []SuperJuegoRecord
	for rows.Next() {
		var r SuperJuegoRecord
		if err := rows.Scan(&r.ID, &r.Juego, &r.NombreJugador, &r.EmpresaID, &r.Puntaje, &r.Nivel, &r.Fecha); err != nil {
			return nil, err
		}
		tops = append(tops, r)
	}
	return tops, nil
}
