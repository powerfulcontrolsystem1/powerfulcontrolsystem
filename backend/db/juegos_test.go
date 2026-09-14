package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"testing"
	"time"
)

func TestJuegosPersistencePostgres(t *testing.T) {
	dsn := os.Getenv("PCS_JUEGOS_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL PCS_JUEGOS_TEST_DSN")
	}
	conn, err := sql.Open(PostgresCompatDriverName(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	ctx := context.Background()
	// Apply the exact new migration in this explicitly isolated test database.
	var ready bool
	if err := conn.QueryRow(`SELECT to_regclass('juegos_partidas') IS NOT NULL`).Scan(&ready); err != nil {
		t.Fatal(err)
	}
	if !ready {
		tx, err := conn.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := applyJuegosSchemaTx(ctx, tx); err != nil {
			tx.Rollback()
			t.Fatal(err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}
	}
	prefix := "juegos-test-" + time.Now().Format("20060102150405.000000000")
	owners := []string{prefix + ":admin:1", prefix + ":empresa:1:usuario:1", prefix + ":empresa:2:usuario:1"}
	t.Cleanup(func() {
		for _, owner := range owners {
			conn.Exec(`DELETE FROM juegos_partidas WHERE propietario=$1`, owner)
			conn.Exec(`DELETE FROM juegos_records WHERE propietario=$1`, owner)
		}
	})
	state := json.RawMessage(`{"schema":1,"game":"pacman","score":100,"data":{"level":2}}`)
	version, err := SaveJuegoPartida(ctx, conn, owners[0], "QA Uno", "pacman", 0, 100, state)
	if err != nil || version != 1 {
		t.Fatalf("initial save: %d %v", version, err)
	}
	for _, owner := range owners[1:] {
		p, err := GetJuegoPartida(ctx, conn, owner, "pacman")
		if err != nil || p.Version != 0 || string(p.Estado) != "null" {
			t.Fatalf("private save leaked: %v", err)
		}
	}
	var first time.Time
	conn.QueryRow(`SELECT fecha FROM juegos_records WHERE propietario=$1 AND juego='pacman'`, owners[0]).Scan(&first)
	for _, score := range []int64{80, 100} {
		version, err = SaveJuegoPartida(ctx, conn, owners[0], "QA Uno", "pacman", version, score, state)
		if err != nil {
			t.Fatal(err)
		}
	}
	var best int64
	var date time.Time
	conn.QueryRow(`SELECT puntaje,fecha FROM juegos_records WHERE propietario=$1 AND juego='pacman'`, owners[0]).Scan(&best, &date)
	if best != 100 || !date.Equal(first) {
		t.Fatal("lower/equal score changed record")
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	success, conflict := 0, 0
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := SaveJuegoPartida(ctx, conn, owners[0], "QA Uno", "pacman", version, 200, state)
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				success++
			} else if errors.Is(err, ErrJuegoConflict) {
				conflict++
			} else {
				t.Errorf("concurrent save: %v", err)
			}
		}()
	}
	wg.Wait()
	if success != 1 || conflict != 7 {
		t.Fatalf("CAS success=%d conflicts=%d", success, conflict)
	}
	_, err = SaveJuegoPartida(ctx, conn, owners[2], "QA Otra Empresa", "pacman", 0, 300, state)
	if err != nil {
		t.Fatal(err)
	}
	records, err := GetJuegoRecords(ctx, conn, "pacman")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, record := range records {
		if record.Nombre == "QA Otra Empresa" && record.Puntaje == 300 {
			found = true
		}
	}
	if !found {
		t.Fatal("cross-company public projection missing")
	}
	p, err := GetJuegoPartida(ctx, conn, owners[0], "pacman")
	if err != nil || p.Version != 4 {
		t.Fatalf("durable state version: %d %v", p.Version, err)
	}
}
