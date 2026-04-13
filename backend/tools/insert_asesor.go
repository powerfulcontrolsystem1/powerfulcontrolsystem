package main

import (
    "database/sql"
    "flag"
    "fmt"
    "log"
    "strings"

    dbpkg "github.com/you/pos-backend/db"
    _ "modernc.org/sqlite"
)

func main() {
    email := flag.String("email", "", "email del asesor (required)")
    nombre := flag.String("nombre", "", "nombre del asesor")
    rol := flag.String("rol", "vendedor", "rol del asesor")
    notas := flag.String("notas", "", "notas")
    dbPath := flag.String("db", "superadministrador.db", "ruta a la BD superadministrador")
    flag.Parse()

    if strings.TrimSpace(*email) == "" {
        log.Fatalf("email is required")
    }

    db, err := sql.Open("sqlite", *dbPath)
    if err != nil {
        log.Fatalf("failed to open db: %v", err)
    }
    defer db.Close()

    id, err := dbpkg.CreateAsesor(db, *email, *nombre, *rol, *notas)
    if err != nil {
        log.Fatalf("failed to create asesor: %v", err)
    }
    fmt.Printf("created asesor id=%d\n", id)
}
