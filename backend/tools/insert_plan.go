//go:build tools
// +build tools

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
    asesorID := flag.String("asesor_id", "", "identificador del asesor (email o codigo)")
    asesorEmail := flag.String("asesor_email", "", "email del asesor")
    empresaID := flag.Int64("empresa_id", 0, "empresa_id opcional")
    comision := flag.Float64("comision", 10.0, "porcentaje comision venta")
    meses := flag.Int("meses", 1, "meses de renovacion")
    notas := flag.String("notas", "", "notas")
    dbPath := flag.String("db", "superadministrador.db", "ruta a la BD superadministrador")
    flag.Parse()

    if strings.TrimSpace(*asesorID) == "" {
        log.Fatalf("asesor_id is required")
    }

    db, err := sql.Open("sqlite", *dbPath)
    if err != nil {
        log.Fatalf("failed to open db: %v", err)
    }
    defer db.Close()

    id, err := dbpkg.CreateAsesorComercialPlan(db, *asesorID, *asesorEmail, *empresaID, *comision, 0.0, *meses, *notas)
    if err != nil {
        log.Fatalf("failed to create plan: %v", err)
    }
    fmt.Printf("created plan id=%d\n", id)
}
