// pcs-migrate is the only deployable role allowed to execute runtime schema
// changes. It intentionally exits after a successful, idempotent run.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"

	dbpkg "github.com/you/pos-backend/db"
)

func open(name, dsn string) (*sql.DB, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, fmt.Errorf("%s is required", name)
	}
	dbConn, err := sql.Open(dbpkg.PostgresCompatDriverName(), dsn)
	if err != nil {
		return nil, err
	}
	pool, err := dbpkg.LoadPostgresPoolConfig(os.Getenv, "migrate")
	if err != nil {
		_ = dbConn.Close()
		return nil, err
	}
	if err := dbpkg.ConfigurePostgresPool(dbConn, pool); err != nil {
		_ = dbConn.Close()
		return nil, err
	}
	if err := dbConn.Ping(); err != nil {
		_ = dbConn.Close()
		return nil, err
	}
	return dbConn, nil
}

func main() {
	if err := os.Setenv("DB_DIALECT", "postgres"); err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	empresas, err := open("DB_EMPRESAS_DSN", os.Getenv("DB_EMPRESAS_DSN"))
	if err != nil {
		log.Fatal(err)
	}
	defer empresas.Close()
	super, err := open("DB_SUPERADMIN_DSN", os.Getenv("DB_SUPERADMIN_DSN"))
	if err != nil {
		log.Fatal(err)
	}
	defer super.Close()

	for _, target := range []struct {
		name  string
		scope string
		db    *sql.DB
	}{
		{name: "empresas", scope: dbpkg.MigrationTargetEmpresas, db: empresas},
		{name: "superadministrador", scope: dbpkg.MigrationTargetSuper, db: super},
	} {
		// This compatibility layer is still required by legacy PostgreSQL queries.
		// It runs in the migration role, never in the API or pcs-worker. The legacy
		// baseline is registered by the migration catalog before runtime roles start.
		if err := dbpkg.EnsurePostgresRuntimeCompat(target.db); err != nil {
			log.Fatalf("%s PostgreSQL compatibility: %v", target.name, err)
		}
		report, err := dbpkg.ApplyPlatformMigrations(ctx, target.db, target.scope, "pcs-migrate")
		if err != nil {
			log.Fatalf("%s migration failed: %v", target.name, err)
		}
		log.Printf("%s migrations: applied=%d existing=%d legacy_marked=%d", target.name, len(report.Applied), len(report.AlreadyKnown), len(report.LegacyMarked))
	}
	log.Print("migrations completed: checksummed platform foundation")
}
