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

	// Query configuraciones for Mercado Pago keys
	rows3, err := dbSuper.Query("SELECT config_key, value, encrypted, fecha_actualizacion FROM configuraciones WHERE config_key LIKE 'mercadopago.%'")
	if err != nil {
		fmt.Println("No se pudo leer configuraciones:", err)
	} else {
		defer rows3.Close()
		fmt.Println("\nConfiguraciones (mercadopago.*) en superadministrador.db:")
		for rows3.Next() {
			var key string
			var value sql.NullString
			var encrypted sql.NullInt64
			var fecha sql.NullString
			if err := rows3.Scan(&key, &value, &encrypted, &fecha); err != nil {
				log.Println("scan error:", err)
				continue
			}
			enc := 0
			if encrypted.Valid {
				enc = int(encrypted.Int64)
			}
			val := "(null)"
			if value.Valid {
				val = value.String
			}
			ts := ""
			if fecha.Valid {
				ts = fecha.String
			}
			fmt.Printf(" - %s | encrypted=%d | fecha_actualizacion=%s | value=%s\n", key, enc, ts, val)
		}

		// Query pagos_mercadopago to verify preferences created
		rows4, err := dbSuper.Query("SELECT id, licencia_id, empresa_id, preference_id, payment_id, status, fecha_creacion FROM pagos_mercadopago ORDER BY id DESC")
		if err != nil {
			fmt.Println("No se pudo leer pagos_mercadopago:", err)
		} else {
			defer rows4.Close()
			fmt.Println("\nPagos Mercado Pago (pagos_mercadopago):")
			for rows4.Next() {
				var id int64
				var licenciaID sql.NullInt64
				var empresaID sql.NullInt64
				var prefID sql.NullString
				var paymentID sql.NullString
				var status sql.NullString
				var fecha sql.NullString
				if err := rows4.Scan(&id, &licenciaID, &empresaID, &prefID, &paymentID, &status, &fecha); err != nil {
					log.Println("scan error:", err)
					continue
				}
				lid := ""
				if licenciaID.Valid {
					lid = fmt.Sprint(licenciaID.Int64)
				}
				eid := ""
				if empresaID.Valid {
					eid = fmt.Sprint(empresaID.Int64)
				}
				p := "(null)"
				if prefID.Valid {
					p = prefID.String
				}
				pay := "(null)"
				if paymentID.Valid {
					pay = paymentID.String
				}
				st := "(null)"
				if status.Valid {
					st = status.String
				}
				ts := ""
				if fecha.Valid {
					ts = fecha.String
				}
				fmt.Printf(" - id=%d | licencia=%s | empresa=%s | pref=%s | payment=%s | status=%s | fecha=%s\n", id, lid, eid, p, pay, st, ts)
			}
		}
	}
}
