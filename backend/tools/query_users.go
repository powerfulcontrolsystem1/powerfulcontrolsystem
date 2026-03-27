package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	dbEmpPath := os.Getenv("DB_EMPRESAS_PATH")
	if dbEmpPath == "" {
		dbEmpPath = "db/empresas.db"
	}
	dbSuperPath := os.Getenv("DB_SUPERADMIN_PATH")
	if dbSuperPath == "" {
		dbSuperPath = "db/superadministrador.db"
	}

	fmt.Println("Usando DB_EMPRESAS_PATH=", dbEmpPath)
	fmt.Println("Usando DB_SUPERADMIN_PATH=", dbSuperPath)

	dbEmp, err := sql.Open("sqlite", dbEmpPath)
	if err != nil {
		log.Fatalf("error abriendo empresas db: %v", err)
	}
	defer dbEmp.Close()

	dbSuper, err := sql.Open("sqlite", dbSuperPath)
	if err != nil {
		log.Fatalf("error abriendo superadministrador db: %v", err)
	}
	defer dbSuper.Close()

	// Query administradores in super DB
	rows, err := dbSuper.Query("SELECT email, role FROM administradores")
	if err != nil {
		fmt.Println("No se pudo leer administradores:", err)
	} else {
		defer rows.Close()
		count := 0
		fmt.Println("Administradores en superadministrador.db:")
		for rows.Next() {
			var email, role string
			if err := rows.Scan(&email, &role); err != nil {
				log.Println("scan error:", err)
				continue
			}
			count++
			fmt.Printf(" - %s (%s)\n", email, role)
		}
		fmt.Printf("Total administradores: %d\n", count)
	}

	// Query users in empresas db
	rows2, err := dbEmp.Query("SELECT email, role FROM users")
	if err != nil {
		fmt.Println("No se pudo leer users:", err)
	} else {
		defer rows2.Close()
		count2 := 0
		fmt.Println("\nUsuarios en empresas.db (tabla users):")
		for rows2.Next() {
			var email, role sql.NullString
			if err := rows2.Scan(&email, &role); err != nil {
				log.Println("scan error:", err)
				continue
			}
			count2++
			r := "(sin role)"
			if role.Valid {
				r = role.String
			}
			fmt.Printf(" - %s (%s)\n", email.String, r)
		}
		fmt.Printf("Total usuarios tabla users: %d\n", count2)
	}
}
