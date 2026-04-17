//go:build tools
// +build tools

package main

import (
    "database/sql"
    "flag"
    "fmt"
    "log"

    _ "modernc.org/sqlite"
)

func main() {
    dbPath := flag.String("db", "superadministrador.db", "ruta a la BD superadministrador")
    licenciaID := flag.Int64("licencia_id", 0, "licencia_id a buscar")
    empresaID := flag.Int64("empresa_id", 0, "empresa_id a buscar")
    flag.Parse()

    db, err := sql.Open("sqlite", *dbPath)
    if err != nil {
        log.Fatalf("failed to open db: %v", err)
    }
    defer db.Close()

    fmt.Println("--- pagos_wompi (últimos) ---")
    q := "SELECT id, licencia_id, empresa_id, transaction_id, reference, status, discount_code, asesor_id, fecha_creacion FROM pagos_wompi WHERE licencia_id = ? AND empresa_id = ? ORDER BY id DESC LIMIT 10"
    rows, err := db.Query(q, *licenciaID, *empresaID)
    if err != nil {
        log.Fatalf("query pagos_wompi: %v", err)
    }
    defer rows.Close()
    for rows.Next() {
        var id int64
        var lic sql.NullInt64
        var emp sql.NullInt64
        var tx sql.NullString
        var ref sql.NullString
        var status sql.NullString
        var discount sql.NullString
        var asesor sql.NullString
        var fc sql.NullString
        if err := rows.Scan(&id, &lic, &emp, &tx, &ref, &status, &discount, &asesor, &fc); err != nil {
            log.Fatalf("scan pagos_wompi: %v", err)
        }
        fmt.Printf("id=%d tx=%s ref=%s status=%s discount=%s asesor=%s fecha=%s\n", id, tx.String, ref.String, status.String, discount.String, asesor.String, fc.String)
    }

    fmt.Println("--- asesor_comisiones (últimos) ---")
    q2 := "SELECT id, asesor_id, empresa_id, licencia_id, pago_id, transaction_id, monto_total, porcentaje, monto_comision, referencia, observaciones, programado_para, pagado, fecha_creacion FROM asesor_comisiones WHERE licencia_id = ? AND empresa_id = ? ORDER BY id DESC LIMIT 20"
    rows2, err := db.Query(q2, *licenciaID, *empresaID)
    if err != nil {
        log.Fatalf("query asesor_comisiones: %v", err)
    }
    defer rows2.Close()
    for rows2.Next() {
        var id int64
        var asesor sql.NullString
        var emp sql.NullInt64
        var lic sql.NullInt64
        var pagoID sql.NullInt64
        var tx sql.NullString
        var montoTotal sql.NullFloat64
        var porcentaje sql.NullFloat64
        var montoCom sql.NullFloat64
        var referencia sql.NullString
        var obs sql.NullString
        var programado sql.NullString
        var pagado int
        var fc sql.NullString
        if err := rows2.Scan(&id, &asesor, &emp, &lic, &pagoID, &tx, &montoTotal, &porcentaje, &montoCom, &referencia, &obs, &programado, &pagado, &fc); err != nil {
            log.Fatalf("scan asesor_comisiones: %v", err)
        }
        fmt.Printf("id=%d asesor=%s pago_id=%v tx=%s monto_total=%.2f pct=%.2f monto_com=%.2f prog=%s pagado=%d obs=%s fecha=%s\n", id, asesor.String, pagoID.Int64, tx.String, montoTotal.Float64, porcentaje.Float64, montoCom.Float64, programado.String, pagado, obs.String, fc.String)
    }
}
