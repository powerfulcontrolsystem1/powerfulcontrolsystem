package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

func backup(src string) (string, error) {
	if _, err := os.Stat(src); err != nil {
		return "", err
	}
	now := time.Now().Format("20060102-150405")
	dst := src + "." + now + ".bak"
	in, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return "", err
	}
	return dst, nil
}

func main() {
	empPath := os.Getenv("DB_EMPRESAS_PATH")
	if empPath == "" {
		empPath = "db/empresas.db"
	}
	superPath := os.Getenv("DB_SUPERADMIN_PATH")
	if superPath == "" {
		superPath = "db/superadministrador.db"
	}

	fmt.Println("Empresas DB:", empPath)
	fmt.Println("Super DB:", superPath)

	// backups
	if b, err := backup(empPath); err != nil {
		log.Fatalf("backup empresas failed: %v", err)
	} else {
		fmt.Println("Backup empresas creado:", b)
	}
	if b, err := backup(superPath); err != nil {
		log.Fatalf("backup super failed: %v", err)
	} else {
		fmt.Println("Backup super creado:", b)
	}

	// open DBs
	dbEmp, err := sql.Open("sqlite", empPath)
	if err != nil {
		log.Fatalf("open emp db: %v", err)
	}
	defer dbEmp.Close()
	dbSuper, err := sql.Open("sqlite", superPath)
	if err != nil {
		log.Fatalf("open super db: %v", err)
	}
	defer dbSuper.Close()

	// Read users from empresas.db
	rows, err := dbEmp.Query("SELECT id, email, name, role FROM users")
	if err != nil {
		log.Fatalf("query users failed: %v", err)
	}
	defer rows.Close()

	tx, err := dbSuper.Begin()
	if err != nil {
		log.Fatalf("begin tx super failed: %v", err)
	}
	defer tx.Rollback()

	insertStmt, err := tx.Prepare("INSERT OR IGNORE INTO administradores (email, name, role, fecha_creacion, fecha_actualizacion, estado) VALUES (?, ?, ?, datetime('now','localtime'), datetime('now','localtime'), 'activo')")
	if err != nil {
		log.Fatalf("prepare insert failed: %v", err)
	}
	defer insertStmt.Close()

	updateStmt, err := tx.Prepare("UPDATE administradores SET name = ?, role = ?, fecha_actualizacion = datetime('now','localtime') WHERE email = ?")
	if err != nil {
		log.Fatalf("prepare update failed: %v", err)
	}
	defer updateStmt.Close()

	migrated := 0
	for rows.Next() {
		var id int
		var email, name, role sql.NullString
		if err := rows.Scan(&id, &email, &name, &role); err != nil {
			log.Println("scan user error:", err)
			continue
		}
		r := "administrador"
		if role.Valid && role.String != "" {
			r = role.String
		}
		emailStr := ""
		nameStr := ""
		if email.Valid {
			emailStr = email.String
		}
		if name.Valid {
			nameStr = name.String
		}
		if emailStr == "" {
			continue
		}
		if _, err := insertStmt.Exec(emailStr, nameStr, r); err != nil {
			log.Println("insert admin error:", err)
		}
		if _, err := updateStmt.Exec(nameStr, r, emailStr); err != nil {
			log.Println("update admin error:", err)
		}
		migrated++
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("commit failed: %v", err)
	}

	fmt.Printf("Migrated %d users to administradores.\n", migrated)

	// Delete users from empresas.db
	if migrated > 0 {
		if _, err := dbEmp.Exec("DELETE FROM users"); err != nil {
			log.Fatalf("failed to delete users from empresas.db: %v", err)
		}
		fmt.Println("Eliminados registros de users en empresas.db (tabla vaciada)")
	} else {
		fmt.Println("No se migró ningún usuario; no se borraron registros en empresas.db")
	}

	// show counts
	var cntSuper int
	if err := dbSuper.QueryRow("SELECT COUNT(*) FROM administradores").Scan(&cntSuper); err != nil {
		log.Println("count super failed:", err)
	} else {
		fmt.Printf("Administradores ahora: %d\n", cntSuper)
	}
	var cntEmp int
	if err := dbEmp.QueryRow("SELECT COUNT(*) FROM users").Scan(&cntEmp); err != nil {
		log.Println("count users failed:", err)
	} else {
		fmt.Printf("Usuarios en empresas.db (users): %d\n", cntEmp)
	}

	fmt.Println("Sincronización completada.")
}
