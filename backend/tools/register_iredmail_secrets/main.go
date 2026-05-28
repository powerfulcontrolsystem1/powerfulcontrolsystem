package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/handlers"
)

func loadEnvFile(path string, overwrite bool, preserveExisting map[string]bool) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if key == "" {
			continue
		}
		if strings.TrimSpace(os.Getenv(key)) != "" && (preserveExisting[key] || !overwrite) {
			continue
		}
		_ = os.Setenv(key, value)
	}
	return scanner.Err()
}

func main() {
	log.SetFlags(0)
	preserveRuntime := map[string]bool{
		"CONFIG_ENC_KEY":    true,
		"DB_SUPERADMIN_DSN": true,
		"DB_EMPRESAS_DSN":   true,
		"DB_DIALECT":        true,
	}
	_ = loadEnvFile("backend/.env.local", false, nil)
	_ = loadEnvFile("deploy/.env.platform", true, preserveRuntime)
	_ = loadEnvFile("/root/powerfulcontrolsystem/backend/.env.local", false, nil)
	_ = loadEnvFile("/root/powerfulcontrolsystem/deploy/.env.platform", true, preserveRuntime)

	dsn := strings.TrimSpace(os.Getenv("DB_SUPERADMIN_DSN"))
	if dsn == "" {
		log.Fatal("DB_SUPERADMIN_DSN no esta definido")
	}
	if strings.TrimSpace(os.Getenv("CONFIG_ENC_KEY")) == "" {
		log.Fatal("CONFIG_ENC_KEY no esta definido; no se pueden cifrar secretos")
	}
	dbSuper, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer dbSuper.Close()
	if err := dbSuper.Ping(); err != nil {
		log.Fatal(err)
	}
	if err := dbpkg.EnsurePostgresRuntimeCompat(dbSuper); err != nil {
		log.Fatal(err)
	}
	if err := dbpkg.EnsureEmpresaEmailCorporativoSchema(dbSuper); err != nil {
		log.Fatal(err)
	}
	if err := handlers.EnsureCorporateEmailConfigFromEnv(dbSuper); err != nil {
		log.Fatal(err)
	}
	keys := []string{
		"email_corporativo.enabled",
		"email_corporativo.auto_create",
		"email_corporativo.domain",
		"email_corporativo.webmail_url",
		"email_corporativo.provision_mode",
		"email_corporativo.iredadmin_api_base_url",
		"email_corporativo.iredadmin_admin",
		"email_corporativo.iredadmin_password",
		"email_corporativo.quota_mb",
	}
	for _, key := range keys {
		value, encrypted, err := dbpkg.GetConfigValue(dbSuper, key)
		if err != nil {
			log.Fatalf("%s: %v", key, err)
		}
		if strings.TrimSpace(value) == "" {
			log.Fatalf("%s quedo vacio", key)
		}
		if key == "email_corporativo.iredadmin_password" && !encrypted {
			log.Fatalf("%s no quedo cifrado", key)
		}
	}
	fmt.Println("OK: configuracion iRedMail registrada en base de datos; secreto iRedAdmin cifrado")
}
