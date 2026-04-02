package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	email := flag.String("email", "", "Email del administrador")
	role := flag.String("role", "administrador", "Nuevo rol")
	flag.Parse()

	if *email == "" {
		log.Fatal("-email es obligatorio")
	}

	dbPath := os.Getenv("DB_SUPERADMIN_PATH")
	if dbPath == "" {
		dbPath = "db/superadministrador.db"
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("error abriendo DB: %v", err)
	}
	defer db.Close()

	// Intentar actualizar
	res, err := db.Exec("UPDATE administradores SET role = ? WHERE email = ?", *role, *email)
	if err != nil {
		log.Fatalf("error actualizando role: %v", err)
	}
	ra, _ := res.RowsAffected()
	if ra == 0 {
		// Si no existe, insertar un registro mínimo
		_, err = db.Exec("INSERT INTO administradores (email, name, role, fecha_creacion, estado) VALUES (?, ?, ?, datetime('now','localtime'), 'activo')", *email, *email, *role)
		if err != nil {
			log.Fatalf("error insertando administrador: %v", err)
		}
		fmt.Printf("Usuario %s no existía: insertado con role=%s\n", *email, *role)
		return
	}

	fmt.Printf("Actualizado %d fila(s): %s -> role=%s\n", ra, *email, *role)
}
