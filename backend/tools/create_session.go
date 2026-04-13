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
    email := flag.String("email", "admin_test@example.com", "email del administrador (se creará si no existe)")
    name := flag.String("name", "Admin Test", "nombre del administrador")
    role := flag.String("role", "super_administrador", "rol del administrador")
    token := flag.String("token", "test-session-token-123", "token de sesión a crear")
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
    if err := dbpkg.CreateSession(db, *email, "127.0.0.1", "create_session_tool", *token); err != nil {
        log.Fatalf("failed to create session: %v", err)
    }
    fmt.Printf("created session token=%s for admin=%s\n", *token, *email)
}
