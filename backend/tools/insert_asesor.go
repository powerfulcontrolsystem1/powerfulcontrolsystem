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
    email := flag.String("email", "asesor_test@example.com", "email del asesor (se creará si no existe)")
    name := flag.String("name", "Asesor Test", "nombre del asesor")
    role := flag.String("role", "asesor", "rol del asesor")
    dbPath := flag.String("db", "superadministrador.db", "ruta a la BD superadministrador")
    flag.Parse()

    db, err := sql.Open("sqlite", *dbPath)
    if err != nil {
        log.Fatalf("failed to open db: %v", err)
    }
    defer db.Close()

    if err := dbpkg.UpsertAdministrador(db, *email, *name, *role, ""); err != nil {
        log.Fatalf("failed to upsert administrador: %v", err)
    }
    fmt.Printf("upserted asesor email=%s name=%s role=%s\n", *email, *name, *role)
}
