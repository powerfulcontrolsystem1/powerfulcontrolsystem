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
    tipoID := flag.Int64("tipo_id", 0, "tipo_id opcional")
    nombre := flag.String("nombre", "LICENCIA PRUEBA", "nombre")
    descripcion := flag.String("descripcion", "licencia para pruebas", "descripcion")
    valor := flag.Float64("valor", 10000.0, "valor")
    duracion := flag.Int("duracion", 30, "duracion en dias")
    modulos := flag.String("modulos", "", "modulos habilitados")
    superRol := flag.Int("super_rol", 0, "super rol habilitado")
    dbPath := flag.String("db", "superadministrador.db", "ruta a la BD superadministrador")
    flag.Parse()

    db, err := sql.Open("sqlite", *dbPath)
    if err != nil {
        log.Fatalf("failed to open db: %v", err)
    }
    defer db.Close()

    id, err := dbpkg.CreateLicencia(db, *tipoID, *nombre, *descripcion, *valor, *duracion, *modulos, *superRol)
    if err != nil {
        log.Fatalf("failed to create licencia: %v", err)
    }
    fmt.Printf("created licencia id=%d\n", id)
}
