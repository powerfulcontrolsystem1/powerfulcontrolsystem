package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	dbPath := os.Getenv("DB_SUPERADMIN_PATH")
	if dbPath == "" {
		dbPath = "db/superadministrador.db"
	}
	fmt.Println("Conectando a:", dbPath)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("error abriendo DB: %v", err)
	}
	defer db.Close()

	// comprobar si columna 'role' existe
	rows, err := db.Query("PRAGMA table_info(administradores)")
	if err != nil {
		log.Fatalf("PRAGMA error: %v", err)
	}
	defer rows.Close()

	colExists := false
	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dfltValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			log.Fatalf("scan pragma: %v", err)
		}
		if name == "role" {
			colExists = true
			break
		}
	}

	if colExists {
		fmt.Println("La columna 'role' ya existe en administradores. No se realizan cambios.")
	} else {
		fmt.Println("Añadiendo columna 'role' a administradores...")
		if _, err := db.Exec("ALTER TABLE administradores ADD COLUMN role TEXT DEFAULT 'administrador'"); err != nil {
			log.Fatalf("failed to add column role: %v", err)
		}
		fmt.Println("Columna añadida. Actualizando filas existentes...")
		if _, err := db.Exec("UPDATE administradores SET role = 'administrador' WHERE role IS NULL OR role = ''"); err != nil {
			log.Fatalf("failed to update existing rows: %v", err)
		}
		fmt.Println("Actualización completada.")
	}

	// mostrar conteo
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM administradores").Scan(&count); err != nil {
		log.Printf("no se pudo contar administradores: %v", err)
	} else {
		fmt.Printf("Total administradores: %d\n", count)
	}

	// listar muestra de registros
	fmt.Println("Muestra de administradores (email, role):")
	r, err := db.Query("SELECT email, COALESCE(role,'') FROM administradores LIMIT 20")
	if err != nil {
		log.Fatalf("query sample failed: %v", err)
	}
	defer r.Close()
	for r.Next() {
		var email, role string
		if err := r.Scan(&email, &role); err != nil {
			log.Fatal(err)
		}
		fmt.Printf(" - %s (%s)\n", email, role)
	}

	fmt.Println("Hecho.")
}
