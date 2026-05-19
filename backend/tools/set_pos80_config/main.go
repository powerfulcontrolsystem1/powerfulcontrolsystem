package main

import (
	"database/sql"
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

func importDotEnvValues(path string) map[string]string {
	values := map[string]string{}
	raw, err := os.ReadFile(path)
	if err != nil {
		return values
	}
	for _, line := range strings.Split(string(raw), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		idx := strings.Index(trimmed, "=")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(trimmed[:idx])
		value := strings.TrimSpace(trimmed[idx+1:])
		values[key] = strings.Trim(value, "\"'")
	}
	return values
}

func ensureEnvFromLocalFile() {
	candidates := []string{
		filepath.Join("backend", ".env.local"),
		filepath.Join(".", ".env.local"),
	}
	for _, path := range candidates {
		values := importDotEnvValues(path)
		if len(values) == 0 {
			continue
		}
		for key, value := range values {
			if strings.TrimSpace(os.Getenv(key)) == "" {
				_ = os.Setenv(key, value)
			}
		}
		return
	}
}

func rewriteRuntimePostgresDSNForTunnel(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.TrimSpace(os.Getenv("DB_VPS_TUNNEL_ENABLED")) != "1" {
		return raw
	}
	localPort := strings.TrimSpace(os.Getenv("DB_VPS_LOCAL_PORT"))
	if localPort == "" {
		return raw
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	hostname := parsed.Hostname()
	if hostname == "" {
		hostname = "127.0.0.1"
	}
	if hostname != "127.0.0.1" && hostname != "localhost" {
		return raw
	}
	parsed.Host = net.JoinHostPort("127.0.0.1", localPort)
	return parsed.String()
}

func main() {
	ensureEnvFromLocalFile()

	var empresaID int64
	var allEmpresas bool
	var actor string
	flag.Int64Var(&empresaID, "empresa", 33, "empresa_id a configurar")
	flag.BoolVar(&allEmpresas, "all", false, "configura todas las empresas activas")
	flag.StringVar(&actor, "actor", "codex-pos80-config", "usuario_creador/actor operativo")
	flag.Parse()

	if !allEmpresas && empresaID <= 0 {
		log.Fatal("empresa debe ser mayor que cero")
	}
	dsn := rewriteRuntimePostgresDSNForTunnel(os.Getenv("DB_EMPRESAS_DSN"))
	if strings.TrimSpace(dsn) == "" {
		log.Fatal("DB_EMPRESAS_DSN no esta configurado")
	}
	_ = os.Setenv("DB_DIALECT", "postgres")

	dbConn, err := sql.Open(dbpkg.PostgresCompatDriverName(), dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer dbConn.Close()
	if err := dbConn.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	if err := dbpkg.EnsurePostgresRuntimeCompat(dbConn); err != nil {
		log.Fatalf("compat postgres: %v", err)
	}
	if err := dbpkg.EnsureEmpresaCorteCajaConfiguracionSchema(dbConn); err != nil {
		log.Fatalf("schema corte caja configuracion: %v", err)
	}
	if err := dbpkg.EnsureEmpresaImpresorasSchema(dbConn); err != nil {
		log.Fatalf("schema impresoras: %v", err)
	}

	if allEmpresas {
		count, err := dbpkg.EnsureAllEmpresasPOS80Defaults(dbConn, actor)
		if err != nil {
			log.Fatalf("configurar empresas POS 80mm: %v", err)
		}
		fmt.Printf("empresas_configuradas=%d reporte_formato=pos impresora_predeterminada=%s funcionalidades=%d\n", count, dbpkg.DefaultEmpresaPOS80PrinterCode, len(dbpkg.DefaultEmpresaPOS80Funcionalidades))
		return
	}

	if _, err := dbpkg.EnsureEmpresaPOS80Defaults(dbConn, empresaID, actor); err != nil {
		log.Fatalf("configurar empresa POS 80mm: %v", err)
	}
	var formato string
	var impresora string
	err = dbConn.QueryRow(`
		SELECT COALESCE(c.formato_impresion, ''), COALESCE(i.nombre, '')
		FROM empresa_corte_caja_configuracion c
		LEFT JOIN empresa_impresoras i ON i.empresa_id = c.empresa_id AND i.codigo = ?
		WHERE c.empresa_id = ?
	`, dbpkg.DefaultEmpresaPOS80PrinterCode, empresaID).Scan(&formato, &impresora)
	if err != nil {
		log.Fatalf("verificar configuracion: %v", err)
	}

	fmt.Printf("empresa_id=%d reporte_formato=%s impresora_predeterminada=%s funcionalidades=%d\n", empresaID, formato, impresora, len(dbpkg.DefaultEmpresaPOS80Funcionalidades))
}
