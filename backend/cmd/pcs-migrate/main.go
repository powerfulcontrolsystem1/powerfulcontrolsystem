package main

import (
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
		name string
		db   *sql.DB
	}{
		{name: "empresas", db: empresas},
		{name: "superadministrador", db: super},
	} {
		if err := dbpkg.EnsurePostgresRuntimeCompat(target.db); err != nil {
			log.Fatalf("%s compatibility: %v", target.name, err)
		}
		if err := dbpkg.EnsureSchemaMigrationsTable(target.db); err != nil {
			log.Fatalf("%s migration ledger: %v", target.name, err)
		}
		if err := dbpkg.RegisterSchemaMigration(target.db, "platform", "20260714-runtime-foundation", "runtime roles, durable queue and outbox"); err != nil {
			log.Fatalf("%s migration ledger record: %v", target.name, err)
		}
	}
	if err := dbpkg.EnsureMobileAPIIdempotencySchema(empresas); err != nil {
		log.Fatal(err)
	}
	if err := dbpkg.EnsureAsyncJobsSchema(super); err != nil {
		log.Fatal(err)
	}
	if err := dbpkg.EnsureOutboxSchema(super); err != nil {
		log.Fatal(err)
	}
	log.Print("migrations completed: runtime foundation")
}
