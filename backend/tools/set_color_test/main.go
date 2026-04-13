package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"

	dbpkg "github.com/you/pos-backend/db"
)

func main() {
	dbPath := "backend/empresas.db"
	if p := os.Getenv("DB_EMPRESAS_PATH"); p != "" {
		dbPath = p
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	payload := dbpkg.EmpresaConfiguracionAvanzada{
		EmpresaID:            1,
		ColorCarritoActivo:   "#123abc",
		ColorCarritoInactivo: "#abcdef",
		UsuarioCreador:       "script-test",
	}

	id, err := dbpkg.UpsertEmpresaConfiguracionAvanzada(db, payload)
	if err != nil {
		log.Fatalf("upsert error: %v", err)
	}
	fmt.Printf("Upsert OK, id=%d\n", id)

	cfg, err := dbpkg.GetEmpresaConfiguracionAvanzada(db, payload.EmpresaID)
	if err != nil {
		log.Fatalf("get config error: %v", err)
	}
	fmt.Printf("color_carrito_activo=%s\n", cfg.ColorCarritoActivo)
	fmt.Printf("color_carrito_inactivo=%s\n", cfg.ColorCarritoInactivo)
}
