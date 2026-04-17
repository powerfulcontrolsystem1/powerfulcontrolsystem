//go:build tools
// +build tools

package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	superPath := "superadministrador.db"
	empPath := "empresas.db"

	dbs, err := sql.Open("sqlite", superPath)
	if err != nil {
		log.Fatalf("open super db: %v", err)
	}
	defer dbs.Close()

	rows, err := dbs.Query("SELECT COALESCE(email,''), COALESCE(name,''), COALESCE(role,'') FROM administradores LIMIT 10")
	if err != nil {
		log.Fatalf("query administradores: %v", err)
	}
	defer rows.Close()
	fmt.Println("ADMINS:")
	for rows.Next() {
		var email, name, role string
		if err := rows.Scan(&email, &name, &role); err != nil {
			log.Fatalf("scan admin row: %v", err)
		}
		fmt.Printf("%s|%s|%s\n", email, name, role)
	}

	dbe, err := sql.Open("sqlite", empPath)
	if err != nil {
		log.Fatalf("open empresas db: %v", err)
	}
	defer dbe.Close()

	rows2, err := dbe.Query("SELECT COALESCE(id,0), COALESCE(nombre,'') FROM empresas LIMIT 10")
	if err != nil {
		log.Fatalf("query empresas: %v", err)
	}
	defer rows2.Close()
	fmt.Println("EMPRESAS:")
	for rows2.Next() {
		var id int64
		var nombre string
		if err := rows2.Scan(&id, &nombre); err != nil {
			log.Fatalf("scan empresa row: %v", err)
		}
		fmt.Printf("%d|%s\n", id, nombre)
	}
}
