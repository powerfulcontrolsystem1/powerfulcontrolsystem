//go:build tools
// +build tools

package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func main() {
	since := flag.Int("since", 24, "horas hacia atras para consultar runtime")
	dbPath := flag.String("db", "epayco_runtime.db", "ruta a la BD runtime")
	flag.Parse()

	// placeholder: consultar runtime epayco
	fmt.Printf("Simulando query epayco runtime desde hace %d horas en %s (now=%v)\n", *since, *dbPath, time.Now())
	_ = dbpkg // evitar unused import en ejemplo
	log.Println("done")
}
package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func loadEnvFile(path string) map[string]string {
	out := map[string]string{}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return out
	}
	lines := strings.Split(string(b), "\n")
	for _, l := range lines {
		raw := strings.TrimSpace(l)
		if raw == "" || strings.HasPrefix(raw, "#") {
			continue
		}
		idx := strings.Index(raw, "=")
		if idx <= 0 {
			continue
		}
		k := strings.TrimSpace(raw[:idx])
		v := strings.TrimSpace(raw[idx+1:])
		if len(v) >= 2 && ((v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'')) {
			v = v[1 : len(v)-1]
		}
		out[k] = v
	}
	return out
}

func main() {
	// Try several likely locations for .env.local (repo root and backend/)
	envPaths := []string{"backend/.env.local", ".env.local", "backend/.env", ".env"}
	env := map[string]string{}
	for _, p := range envPaths {
		if _, err := os.Stat(p); err == nil {
			env = loadEnvFile(p)
			break
		}
	}
	// fallback to process env
	for _, k := range []string{"DB_SUPERADMIN_DSN", "DB_EMPRESAS_DSN"} {
		if _, ok := env[k]; !ok {
			if v := strings.TrimSpace(os.Getenv(k)); v != "" {
				env[k] = v
			}
		}
	}

	dsn := env["DB_SUPERADMIN_DSN"]
	if dsn == "" {
		log.Fatalf("DB_SUPERADMIN_DSN no encontrado en backend/.env.local ni en variables de entorno")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	fmt.Println("Conectado a DB superadministrador.")

	// Query epayco configs
	rows, err := db.Query("SELECT config_key, value, encrypted FROM configuraciones WHERE config_key LIKE 'epayco.%' ORDER BY config_key")
	if err != nil {
		fmt.Printf("error consultando configuraciones epayco: %v\n", err)
	} else {
		defer rows.Close()
		fmt.Println("\nConfiguraciones epayco:")
		for rows.Next() {
			var key string
			var val sql.NullString
			var enc sql.NullInt64
			if err := rows.Scan(&key, &val, &enc); err != nil {
				fmt.Printf(" scan err: %v\n", err)
				continue
			}
			v := "(NULL)"
			if val.Valid { v = val.String }
			e := 0
			if enc.Valid { e = int(enc.Int64) }
			fmt.Printf(" - %s | encrypted=%d | value=%s\n", key, e, v)
		}
	}

	// Query latest sessions
	rows2, err := db.Query("SELECT id, admin_email, token, activo, fecha_inicio FROM sesiones ORDER BY id DESC LIMIT 10")
	if err != nil {
		fmt.Printf("error consultando sesiones: %v\n", err)
		return
	}
	defer rows2.Close()

	fmt.Println("\nSesiones recientes (id | admin_email | token (masked) | activo | fecha_inicio):")
	for rows2.Next() {
		var id int64
		var email sql.NullString
		var token sql.NullString
		var activo sql.NullInt64
		var fecha sql.NullString
		if err := rows2.Scan(&id, &email, &token, &activo, &fecha); err != nil {
			fmt.Printf(" scan sesion err: %v\n", err)
			continue
		}
		tok := "(NULL)"
		if token.Valid {
			if len(token.String) > 8 {
				tok = token.String[:4] + "..." + token.String[len(token.String)-4:]
			} else {
				tok = token.String
			}
		}
		em := "(NULL)"
		if email.Valid { em = email.String }
		ac := 0
		if activo.Valid { ac = int(activo.Int64) }
		fs := ""
		if fecha.Valid { fs = fecha.String }
		fmt.Printf(" - %d | %s | %s | %d | %s\n", id, em, tok, ac, fs)
	}

	// Helpful hint for the user
	fmt.Println("\nSi hay una sesión activa, puedes copiar el token (comienzo/final mostrado) y usarlo como cookie 'session_token' para llamadas HTTP al backend.")
}
