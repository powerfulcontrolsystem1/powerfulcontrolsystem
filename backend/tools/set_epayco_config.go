package main

import (
    "database/sql"
    "flag"
    "fmt"
    "log"
    "os"

    _ "modernc.org/sqlite"

    dbpkg "github.com/you/pos-backend/db"
    "github.com/you/pos-backend/utils"
)

func ensureConfiguracionesTable(db *sql.DB) error {
    create := `CREATE TABLE IF NOT EXISTS configuraciones (
        config_key TEXT PRIMARY KEY,
        value TEXT,
        encrypted INTEGER DEFAULT 0,
        fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
        fecha_actualizacion TEXT,
        usuario_creador TEXT,
        estado TEXT DEFAULT 'activo',
        observaciones TEXT
    );` 
    _, err := db.Exec(create)
    return err
}

func main() {
    dbPath := flag.String("db", "db/superadministrador.db", "ruta al archivo de DB superadministrador")
    cust := flag.String("cust", "", "P_CUST_ID_CLIENTE (epayco cust id)")
    key := flag.String("key", "", "P_KEY (epayco secret key)")
    enable := flag.Bool("enable", true, "marcar epayco.enabled = 1")
    flag.Parse()

    if *cust == "" && *key == "" {
        log.Fatalf("se requiere --cust o --key para guardar")
    }

    db, err := sql.Open("sqlite", *dbPath)
    if err != nil {
        log.Fatalf("open db: %v", err)
    }
    defer db.Close()

    if err := ensureConfiguracionesTable(db); err != nil {
        log.Fatalf("ensure configuraciones table: %v", err)
    }

    if *cust != "" {
        if err := dbpkg.SetConfigValue(db, "epayco.cust_id", *cust, false); err != nil {
            log.Fatalf("failed to save epayco.cust_id: %v", err)
        }
        fmt.Println("epayco.cust_id guardado")
    }

    if *key != "" {
        if !utils.EncryptionAvailable() {
            log.Fatalf("encryption required: CONFIG_ENC_KEY not set in environment")
        }
        enc, err := utils.EncryptString(*key)
        if err != nil {
            log.Fatalf("encrypt failed: %v", err)
        }
        if err := dbpkg.SetConfigValue(db, "epayco.key", enc, true); err != nil {
            log.Fatalf("failed to save epayco.key: %v", err)
        }
        fmt.Println("epayco.key guardado (cifrado)")
    }

    if enable != nil {
        v := "0"
        if *enable { v = "1" }
        if err := dbpkg.SetConfigValue(db, "epayco.enabled", v, false); err != nil {
            log.Fatalf("failed to save epayco.enabled: %v", err)
        }
        fmt.Println("epayco.enabled actualizado a:", v)
    }

    fmt.Println("Operación completada.")
    os.Exit(0)
}
