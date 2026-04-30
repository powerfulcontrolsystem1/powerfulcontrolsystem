package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func main() {
	var dsnFlag string
	var usuario string
	var overwrite bool
	flag.StringVar(&dsnFlag, "dsn", "", "DSN PostgreSQL de superadministrador; si se omite usa DB_SUPERADMIN_DSN")
	flag.StringVar(&usuario, "usuario", "herramienta.preconfiguracion", "usuario_creador para auditoria de la siembra")
	flag.BoolVar(&overwrite, "overwrite", false, "reemplaza plantillas existentes; por defecto respeta personalizaciones")
	flag.Parse()

	loadEnvDefaults()
	dsn := strings.TrimSpace(dsnFlag)
	if dsn == "" {
		dsn = strings.TrimSpace(os.Getenv("DB_SUPERADMIN_DSN"))
	}
	if dsn == "" {
		log.Fatal("DB_SUPERADMIN_DSN no esta configurado")
	}
	dsn = rewritePostgresDSNForTunnel(dsn)

	dbConn, err := sql.Open(dbpkg.PostgresCompatDriverName(), dsn)
	if err != nil {
		log.Fatalf("open db superadministrador: %v", err)
	}
	defer dbConn.Close()
	if err := dbConn.Ping(); err != nil {
		log.Fatalf("ping db superadministrador: %v", err)
	}
	dbpkg.SetDefaultDB(dbConn)
	if err := dbpkg.EnsurePostgresRuntimeCompat(dbConn); err != nil {
		log.Fatalf("ensure postgres compat: %v", err)
	}

	result, err := dbpkg.SeedDefaultTipoEmpresaPreconfiguraciones(dbConn, usuario, overwrite)
	if err != nil {
		log.Fatalf("registrar preconfiguraciones: %v", err)
	}
	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("serializar resultado: %v", err)
	}
	fmt.Println(string(out))
}

func loadEnvDefaults() {
	for _, path := range envCandidates() {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
			raw := strings.TrimSpace(line)
			if raw == "" || strings.HasPrefix(raw, "#") {
				continue
			}
			idx := strings.Index(raw, "=")
			if idx <= 0 {
				continue
			}
			key := strings.TrimSpace(raw[:idx])
			value := strings.TrimSpace(raw[idx+1:])
			if key == "" || strings.TrimSpace(os.Getenv(key)) != "" {
				continue
			}
			if len(value) >= 2 {
				if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
					value = value[1 : len(value)-1]
				}
			}
			_ = os.Setenv(key, value)
		}
		return
	}
}

func envCandidates() []string {
	wd, _ := os.Getwd()
	candidates := []string{
		".env.local",
		".env",
		filepath.Join("backend", ".env.local"),
		filepath.Join("backend", ".env"),
		filepath.Join("..", "backend", ".env.local"),
		filepath.Join("..", "backend", ".env"),
		filepath.Join("..", ".env.local"),
		filepath.Join("..", ".env"),
	}
	if wd != "" {
		candidates = append(candidates,
			filepath.Join(wd, ".env.local"),
			filepath.Join(wd, ".env"),
			filepath.Join(wd, "backend", ".env.local"),
			filepath.Join(wd, "backend", ".env"),
		)
	}
	return candidates
}

func rewritePostgresDSNForTunnel(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.TrimSpace(os.Getenv("DB_VPS_TUNNEL_ENABLED")) != "1" {
		return raw
	}
	localPort := strings.TrimSpace(os.Getenv("DB_VPS_LOCAL_PORT"))
	if localPort == "" {
		return raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	hostname := u.Hostname()
	if hostname == "" {
		hostname = "127.0.0.1"
	}
	if hostname != "127.0.0.1" && hostname != "localhost" {
		return raw
	}
	u.Host = net.JoinHostPort("127.0.0.1", localPort)
	return u.String()
}
