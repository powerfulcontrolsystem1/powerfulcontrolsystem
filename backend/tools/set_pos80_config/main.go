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
	var actor string
	var printerCode string
	var printerName string
	flag.Int64Var(&empresaID, "empresa", 33, "empresa_id a configurar")
	flag.StringVar(&actor, "actor", "codex-pos80-config", "usuario_creador/actor operativo")
	flag.StringVar(&printerCode, "printer-code", "POS_80MM", "codigo de impresora POS")
	flag.StringVar(&printerName, "printer-name", "Impresora POS 80mm", "nombre visible de impresora POS")
	flag.Parse()

	if empresaID <= 0 {
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

	printerCode = strings.ToUpper(strings.TrimSpace(printerCode))
	if printerCode == "" {
		printerCode = "POS_80MM"
	}
	printerName = strings.TrimSpace(printerName)
	if printerName == "" {
		printerName = "Impresora POS 80mm"
	}

	if _, err := dbConn.Exec(`
		INSERT INTO empresa_corte_caja_configuracion (
			empresa_id, formato_impresion, usuario_creador, estado, observaciones
		) VALUES (
			?, 'pos', ?, 'activo', 'Reporte de turno configurado para impresora POS 80mm'
		)
		ON CONFLICT (empresa_id) DO UPDATE SET
			formato_impresion = 'pos',
			fecha_actualizacion = datetime('now','localtime'),
			usuario_creador = EXCLUDED.usuario_creador,
			estado = 'activo',
			observaciones = EXCLUDED.observaciones
	`, empresaID, actor); err != nil {
		log.Fatalf("guardar formato reporte POS: %v", err)
	}

	if _, err := dbConn.Exec(`
		UPDATE empresa_impresoras
		SET es_predeterminada = 0, fecha_actualizacion = datetime('now','localtime')
		WHERE empresa_id = ? AND codigo <> ?
	`, empresaID, printerCode); err != nil {
		log.Fatalf("limpiar impresoras predeterminadas: %v", err)
	}

	if _, err := dbConn.Exec(`
		INSERT INTO empresa_impresoras (
			empresa_id, codigo, nombre, tipo_conexion, direccion, area_operativa,
			formato_impresion, es_predeterminada, usuario_creador, estado, observaciones
		) VALUES (
			?, ?, ?, 'windows', 'POS 80mm', 'caja',
			'pos', 1, ?, 'activo', 'Impresora POS 80mm activa para pruebas de reporte de turno'
		)
		ON CONFLICT (empresa_id, codigo) DO UPDATE SET
			nombre = EXCLUDED.nombre,
			tipo_conexion = EXCLUDED.tipo_conexion,
			direccion = EXCLUDED.direccion,
			area_operativa = EXCLUDED.area_operativa,
			formato_impresion = 'pos',
			es_predeterminada = 1,
			fecha_actualizacion = datetime('now','localtime'),
			usuario_creador = EXCLUDED.usuario_creador,
			estado = 'activo',
			observaciones = EXCLUDED.observaciones
	`, empresaID, printerCode, printerName, actor); err != nil {
		log.Fatalf("guardar impresora POS: %v", err)
	}

	funcionalidades := []string{"general", "corte_caja", "turno_reporte", "cajon_monedero"}
	for _, funcionalidad := range funcionalidades {
		if _, err := dbConn.Exec(`
			INSERT INTO empresa_impresoras_funcionalidades (
				empresa_id, funcionalidad, impresora_id, prioridad, usuario_creador, estado, observaciones
			)
			SELECT ?, ?, id, 10, ?, 'activo', 'Asignado a POS 80mm para pruebas de turno'
			FROM empresa_impresoras
			WHERE empresa_id = ? AND codigo = ?
			LIMIT 1
			ON CONFLICT (empresa_id, funcionalidad) DO UPDATE SET
				impresora_id = EXCLUDED.impresora_id,
				prioridad = EXCLUDED.prioridad,
				fecha_actualizacion = datetime('now','localtime'),
				usuario_creador = EXCLUDED.usuario_creador,
				estado = 'activo',
				observaciones = EXCLUDED.observaciones
		`, empresaID, funcionalidad, actor, empresaID, printerCode); err != nil {
			log.Fatalf("asignar funcionalidad %s: %v", funcionalidad, err)
		}
	}

	var formato string
	var impresora string
	err = dbConn.QueryRow(`
		SELECT COALESCE(c.formato_impresion, ''), COALESCE(i.nombre, '')
		FROM empresa_corte_caja_configuracion c
		LEFT JOIN empresa_impresoras i ON i.empresa_id = c.empresa_id AND i.codigo = ?
		WHERE c.empresa_id = ?
	`, printerCode, empresaID).Scan(&formato, &impresora)
	if err != nil {
		log.Fatalf("verificar configuracion: %v", err)
	}

	fmt.Printf("empresa_id=%d reporte_formato=%s impresora_predeterminada=%s funcionalidades=%d\n", empresaID, formato, impresora, len(funcionalidades))
}
