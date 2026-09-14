package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

// Personal platform data. Only the explicitly public projection is cross-company.
const juegosSchemaSQL = `CREATE TABLE juegos_partidas (
 propietario TEXT NOT NULL, juego TEXT NOT NULL,
 version BIGINT NOT NULL DEFAULT 0, estado JSONB NOT NULL DEFAULT '{}',
 actualizado_en TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
 PRIMARY KEY (propietario, juego),
 CHECK (juego IN ('pacman','tetris','buscaminas','solitario','selva','sorpresa')),
 CHECK (octet_length(estado::text) <= 524288)
);
CREATE TABLE juegos_records (
 propietario TEXT NOT NULL, juego TEXT NOT NULL,
 nombre TEXT NOT NULL, puntaje BIGINT NOT NULL CHECK (puntaje BETWEEN 1 AND 100000000),
 fecha TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
 PRIMARY KEY (propietario, juego)
);
CREATE INDEX juegos_records_ranking ON juegos_records (juego, puntaje DESC, fecha ASC, propietario);`

func applyJuegosSchemaTx(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, juegosSchemaSQL)
	return err
}

type JuegoPartida struct {
	Version       int64           `json:"version"`
	Estado        json.RawMessage `json:"estado"`
	ActualizadoEn time.Time       `json:"actualizado_en"`
}
type JuegoRecord struct {
	Nombre  string    `json:"nombre"`
	Puntaje int64     `json:"puntaje"`
	Fecha   time.Time `json:"fecha"`
}

var ErrJuegoConflict = errors.New("la partida cambió en otro dispositivo")

func GetJuegoPartida(ctx context.Context, conn *sql.DB, owner, game string) (JuegoPartida, error) {
	p := JuegoPartida{Estado: json.RawMessage(`null`)}
	err := conn.QueryRowContext(ctx, `SELECT version, estado, actualizado_en FROM juegos_partidas WHERE propietario=$1 AND juego=$2`, owner, game).Scan(&p.Version, &p.Estado, &p.ActualizadoEn)
	if errors.Is(err, sql.ErrNoRows) {
		return p, nil
	}
	return p, err
}

// Compare-and-swap prevents stale tabs/devices from overwriting progress. Record
// update and save are one transaction. Equal scores retain the first timestamp.
func SaveJuegoPartida(ctx context.Context, conn *sql.DB, owner, name, game string, version, score int64, state json.RawMessage) (int64, error) {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var next int64
	err = tx.QueryRowContext(ctx, `INSERT INTO juegos_partidas (propietario,juego,version,estado)
 SELECT $1,$2,1,$4::jsonb WHERE $3::bigint=0
 ON CONFLICT (propietario,juego) DO NOTHING RETURNING version`, owner, game, version, string(state)).Scan(&next)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRowContext(ctx, `UPDATE juegos_partidas SET estado=$4::jsonb, version=version+1, actualizado_en=CURRENT_TIMESTAMP
   WHERE propietario=$1 AND juego=$2 AND version=$3 RETURNING version`, owner, game, version, string(state)).Scan(&next)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrJuegoConflict
	}
	if err != nil {
		return 0, err
	}
	if score > 0 {
		_, err = tx.ExecContext(ctx, `INSERT INTO juegos_records(propietario,juego,nombre,puntaje) VALUES($1,$2,$3,$4)
   ON CONFLICT(propietario,juego) DO UPDATE SET nombre=EXCLUDED.nombre,puntaje=EXCLUDED.puntaje,fecha=CURRENT_TIMESTAMP
   WHERE EXCLUDED.puntaje>juegos_records.puntaje`, owner, game, name, score)
		if err != nil {
			return 0, err
		}
	}
	return next, tx.Commit()
}

func GetJuegoRecords(ctx context.Context, conn *sql.DB, game string) ([]JuegoRecord, error) {
	rows, err := conn.QueryContext(ctx, `SELECT nombre,puntaje,fecha FROM juegos_records WHERE juego=$1 ORDER BY puntaje DESC,fecha ASC,propietario ASC LIMIT 20`, game)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := make([]JuegoRecord, 0)
	for rows.Next() {
		var item JuegoRecord
		if err = rows.Scan(&item.Nombre, &item.Puntaje, &item.Fecha); err != nil {
			return nil, err
		}
		records = append(records, item)
	}
	return records, rows.Err()
}
