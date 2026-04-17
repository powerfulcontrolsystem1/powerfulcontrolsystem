//go:build tools
// +build tools

package main

import (
    "database/sql"
    "flag"
    "fmt"
    "log"

    dbpkg "github.com/you/pos-backend/db"
    _ "modernc.org/sqlite"
)

func main() {
    email := flag.String("email", "licencia_test@example.com", "email del administrador propietario de la licencia")
    plan := flag.String("plan", "free", "plan de licencia")
    dbPath := flag.String("db", "superadministrador.db", "ruta a la BD superadministrador")
    flag.Parse()

    db, err := sql.Open("sqlite", *dbPath)
    if err != nil {
        log.Fatalf("failed to open db: %v", err)
    }
    defer db.Close()

    if err := dbpkg.EnsureLicenciasSchema(db); err != nil {
        log.Fatalf("failed to ensure licencias schema: %v", err)
    }
    id, err := dbpkg.CreateLicenciaForAdmin(db, *email, *plan)
    if err != nil {
        log.Fatalf("failed to create licencia: %v", err)
    }
    fmt.Printf("created licencia id=%d for admin=%s plan=%s\n", id, *email, *plan)
}
