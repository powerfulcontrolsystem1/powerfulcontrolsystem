package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestJuegosSimsongMigrationPreservesLegacy(t *testing.T) {
	dsn := os.Getenv("PCS_JUEGOS_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL PCS_JUEGOS_TEST_DSN")
	}
	conn, err := sql.Open(PostgresCompatDriverName(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	ctx := context.Background()
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	// Transaction-local schema avoids modifying other QA fixtures; rollback removes it.
	schema := fmt.Sprintf("simsong_test_%d", time.Now().UnixNano())
	if _, err := tx.ExecContext(ctx, `CREATE SCHEMA `+schema+`; SET LOCAL search_path TO `+schema); err != nil {
		t.Fatal(err)
	}
	if err := applyJuegosSchemaTx(ctx, tx); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO juegos_partidas(propietario,juego,version,estado) VALUES('legacy','selva',7,'{"score":250}')`); err != nil {
		t.Fatal(err)
	}
	if err := applySimsongSchemaTx(ctx, tx); err != nil {
		t.Fatal(err)
	}
	var version, score int
	if err := tx.QueryRowContext(ctx, `SELECT version,(estado->>'score')::int FROM juegos_partidas WHERE propietario='legacy' AND juego='selva'`).Scan(&version, &score); err != nil {
		t.Fatal(err)
	}
	if version != 7 || score != 250 {
		t.Fatal("legacy save was modified")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO juegos_partidas(propietario,juego) VALUES('new-player','simsong')`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO juegos_partidas(propietario,juego) VALUES('unknown','unsupported')`); err == nil {
		t.Fatal("unknown game passed constraint")
	}
}
