package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

func backupFile(src string) (string, error) {
	now := time.Now().Format("20060102-150405")
	dir := filepath.Dir(src)
	base := filepath.Base(src)
	backup := filepath.Join(dir, base+"."+now+".bak")
	in, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer in.Close()
	out, err := os.Create(backup)
	if err != nil {
		return "", err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return "", err
	}
	return backup, nil
}

func ensureIndices(db *sql.DB) error {
	stmts := []string{
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_administradores_email ON administradores(email);",
		"CREATE INDEX IF NOT EXISTS idx_administradores_role ON administradores(role);",
		"CREATE INDEX IF NOT EXISTS idx_sesiones_token ON sesiones(token);",
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	// VACUUM to optimize
	if _, err := db.Exec("VACUUM;"); err != nil {
		// VACUUM can fail in some contexts; log and continue
		log.Println("vacuum warning:", err)
	}
	return nil
}

func listAdmins(db *sql.DB) error {
	rows, err := db.Query("SELECT id, email, COALESCE(role,'') FROM administradores ORDER BY id DESC")
	if err != nil {
		return err
	}
	defer rows.Close()
	count := 0
	fmt.Println("Administradores en superadministrador.db:")
	for rows.Next() {
		var id int
		var email, role string
		if err := rows.Scan(&id, &email, &role); err != nil {
			return err
		}
		count++
		fmt.Printf(" - %s (%s)\n", email, role)
	}
	fmt.Printf("Total administradores: %d\n", count)
	return nil
}

func listUsers(empDB *sql.DB) error {
	rows, err := empDB.Query("SELECT id, email, COALESCE(role,'') FROM users ORDER BY id DESC")
	if err != nil {
		return err
	}
	defer rows.Close()
	count := 0
	fmt.Println("\nUsuarios en empresas.db (tabla users):")
	for rows.Next() {
		var id int
		var email, role string
		if err := rows.Scan(&id, &email, &role); err != nil {
			return err
		}
		count++
		fmt.Printf(" - %s (%s)\n", email, role)
	}
	fmt.Printf("Total usuarios tabla users: %d\n", count)
	return nil
}

func main() {
	superPath := os.Getenv("DB_SUPERADMIN_PATH")
	if superPath == "" {
		superPath = "db/superadministrador.db"
	}
	empPath := os.Getenv("DB_EMPRESAS_PATH")
	if empPath == "" {
		empPath = "db/empresas.db"
	}

	fmt.Println("Super DB:", superPath)
	fmt.Println("Empresas DB:", empPath)

	// Backup super DB
	if _, err := os.Stat(superPath); err == nil {
		b, err := backupFile(superPath)
		if err != nil {
			log.Fatalf("backup failed: %v", err)
		}
		fmt.Println("Backup creado:", b)
	} else {
		fmt.Println("Advertencia: archivo superadministrador.db no encontrado en:", superPath)
	}

	// Open DBs
	dbSuper, err := sql.Open("sqlite", superPath)
	if err != nil {
		log.Fatalf("open super db: %v", err)
	}
	defer dbSuper.Close()
	dbEmp, err := sql.Open("sqlite", empPath)
	if err != nil {
		log.Fatalf("open empresas db: %v", err)
	}
	defer dbEmp.Close()

	// Ensure indices
	if err := ensureIndices(dbSuper); err != nil {
		log.Fatalf("ensure indices failed: %v", err)
	}
	fmt.Println("Indices creados/asegurados.")

	// List admins and users
	if err := listAdmins(dbSuper); err != nil {
		log.Println("list admins error:", err)
	}
	if err := listUsers(dbEmp); err != nil {
		log.Println("list users error:", err)
	}

	fmt.Println("Operación completada.")
}
