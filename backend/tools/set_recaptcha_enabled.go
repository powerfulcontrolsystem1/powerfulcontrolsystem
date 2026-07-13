//go:build tools
// +build tools

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

func rewriteRuntimePostgresDSNForTunnel(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}
	if strings.TrimSpace(os.Getenv("DB_VPS_TUNNEL_ENABLED")) != "1" {
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
	var dsnFlag string
	var enabled string
	var actor string
	var provider string
	var siteKey string
	var secretKey string
	var apply bool
	var confirm string

	flag.StringVar(&dsnFlag, "dsn", "", "Postgres DSN for superadmin DB (overrides DB_SUPERADMIN_DSN env)")
	flag.StringVar(&enabled, "enabled", "1", "0 or 1")
	flag.StringVar(&actor, "actor", "cli-recaptcha", "actor name to write in updated_by config")
	flag.StringVar(&provider, "provider", "", "recaptcha provider: google-recaptcha-v2 or google-recaptcha-enterprise")
	flag.StringVar(&siteKey, "site-key", "", "Google reCAPTCHA site key to persist in super config")
	flag.StringVar(&secretKey, "secret-key", "", "Google reCAPTCHA secret key to persist encrypted in super config")
	flag.BoolVar(&apply, "apply", false, "aplica el cambio; sin este indicador solo simula")
	flag.StringVar(&confirm, "confirm", "", "confirmacion requerida para aplicar")
	flag.Parse()
	if !apply {
		fmt.Printf("DRY-RUN: recaptcha enabled=%s provider=%t site_key=%t secret_key=%t; no se conecto a una base de datos\n", enabled, strings.TrimSpace(provider) != "", strings.TrimSpace(siteKey) != "", strings.TrimSpace(secretKey) != "")
		return
	}
	if confirm != "SET_RECAPTCHA_CONFIG" {
		log.Fatal("para aplicar use -apply -confirm=SET_RECAPTCHA_CONFIG")
	}

	dsn := strings.TrimSpace(dsnFlag)
	if dsn == "" {
		dsn = strings.TrimSpace(os.Getenv("DB_SUPERADMIN_DSN"))
	}
	if dsn == "" {
		log.Fatalf("DB_SUPERADMIN_DSN not set and -dsn not provided")
	}
	dsn = rewriteRuntimePostgresDSNForTunnel(dsn)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	if err := dbpkg.SetConfigValue(db, "security.recaptcha.enabled", enabled, false); err != nil {
		log.Fatalf("SetConfigValue security.recaptcha.enabled: %v", err)
	}
	if err := dbpkg.SetConfigValue(db, "security.recaptcha.enabled.updated_by", actor, false); err != nil {
		log.Fatalf("SetConfigValue security.recaptcha.enabled.updated_by: %v", err)
	}

	trimmedProvider := strings.TrimSpace(provider)
	if trimmedProvider != "" {
		if err := dbpkg.SetConfigValue(db, "security.recaptcha.provider", trimmedProvider, false); err != nil {
			log.Fatalf("SetConfigValue security.recaptcha.provider: %v", err)
		}
		if err := dbpkg.SetConfigValue(db, "security.recaptcha.provider.updated_by", actor, false); err != nil {
			log.Fatalf("SetConfigValue security.recaptcha.provider.updated_by: %v", err)
		}
	}

	trimmedSiteKey := strings.TrimSpace(siteKey)
	if trimmedSiteKey != "" {
		if err := dbpkg.SetConfigValue(db, "security.recaptcha.site_key", trimmedSiteKey, false); err != nil {
			log.Fatalf("SetConfigValue security.recaptcha.site_key: %v", err)
		}
		if err := dbpkg.SetConfigValue(db, "security.recaptcha.site_key.updated_by", actor, false); err != nil {
			log.Fatalf("SetConfigValue security.recaptcha.site_key.updated_by: %v", err)
		}
	}

	trimmedSecretKey := strings.TrimSpace(secretKey)
	if trimmedSecretKey != "" {
		encryptedSecretKey, err := utils.EncryptString(trimmedSecretKey)
		if err != nil {
			log.Fatalf("EncryptString secret-key: %v", err)
		}
		if err := dbpkg.SetConfigValue(db, "security.recaptcha.secret_key", encryptedSecretKey, true); err != nil {
			log.Fatalf("SetConfigValue security.recaptcha.secret_key: %v", err)
		}
		if err := dbpkg.SetConfigValue(db, "security.recaptcha.secret_key.updated_by", actor, false); err != nil {
			log.Fatalf("SetConfigValue security.recaptcha.secret_key.updated_by: %v", err)
		}
	}

	fmt.Printf("OK: set security.recaptcha.enabled=%s provider=%t site_key=%t secret_key=%t (actor=%s)\n", enabled, trimmedProvider != "", trimmedSiteKey != "", trimmedSecretKey != "", actor)
}
