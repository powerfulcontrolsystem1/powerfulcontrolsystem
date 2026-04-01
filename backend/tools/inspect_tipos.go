package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	dbpkg "github.com/you/pos-backend/db"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "db/superadministrador.db")
	if err != nil {
		log.Fatalf("failed to open superadministrador.db: %v", err)
	}
	defer db.Close()

	tipos, err := dbpkg.GetTiposDeUsuario(db, 0, true)
	if err != nil {
		log.Fatalf("GetTiposDeUsuario: %v", err)
	}

	// Si no hay tipos, crear uno de prueba para verificar si el nombre del rol se guarda/recupera
	if len(tipos) == 0 {
		fmt.Println("No hay tipos_de_usuario, creando uno de prueba con rol_id=1, tipo_empresa_id=1...")
		if id, err := dbpkg.CreateTipoDeUsuario(db, 1, 1, "prueba", "tipo de prueba", "inspector"); err != nil {
			log.Printf("CreateTipoDeUsuario error: %v", err)
		} else {
			fmt.Printf("Creado tipo_de_usuario id=%d\n", id)
		}
		// recargar
		tipos, err = dbpkg.GetTiposDeUsuario(db, 0, true)
		if err != nil {
			log.Fatalf("GetTiposDeUsuario after create: %v", err)
		}
	}
	roles, err := dbpkg.GetRolesDeUsuario(db, 0, true)
	if err != nil {
		log.Fatalf("GetRolesDeUsuario: %v", err)
	}

	b1, _ := json.MarshalIndent(tipos, "", "  ")
	fmt.Println("TIPOS_DE_USUARIO:")
	fmt.Println(string(b1))

	b2, _ := json.MarshalIndent(roles, "", "  ")
	fmt.Println("ROLES_DE_USUARIO:")
	fmt.Println(string(b2))
}
