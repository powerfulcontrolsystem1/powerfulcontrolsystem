package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/handlers"
)

func main() {
	apply := flag.Bool("apply", false, "aplica la migracion; sin este indicador solo simula")
	confirm := flag.String("confirm", "", "confirmacion requerida para aplicar")
	webRoot := flag.String("web-root", strings.TrimSpace(os.Getenv("PCS_WEB_ROOT")), "raiz web heredada")
	flag.Parse()

	if *apply && *confirm != "MIGRATE_PRIVATE_UPLOADS" {
		exitError(errors.New("para aplicar use --confirm=MIGRATE_PRIVATE_UPLOADS"))
	}
	if strings.TrimSpace(*webRoot) == "" {
		resolved, err := filepath.Abs(filepath.Join("..", "web"))
		if err != nil {
			exitError(errors.New("no se pudo resolver la raiz web"))
		}
		*webRoot = resolved
	}
	dsn := firstNonEmptyEnv("DB_EMPRESAS_DSN", "DB_EMPRESAS_URL", "PCS_DB_EMPRESAS_DSN")
	if dsn == "" {
		exitError(errors.New("falta el DSN empresarial"))
	}
	dbConn, err := sql.Open(dbpkg.PostgresCompatDriverName(), dsn)
	if err != nil {
		exitError(errors.New("no se pudo abrir la base empresarial"))
	}
	defer func() { _ = dbConn.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := dbConn.PingContext(ctx); err != nil {
		exitError(errors.New("no se pudo validar la base empresarial"))
	}
	result, err := handlers.MigrateLegacyPrivateUploads(dbConn, *webRoot, *apply)
	if err != nil {
		exitError(err)
	}
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		exitError(errors.New("no se pudo emitir el resultado"))
	}
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func exitError(err error) {
	fmt.Fprintln(os.Stderr, "migracion de archivos privados no completada")
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	os.Exit(1)
}
