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
)

// rewriteRuntimePostgresDSNForTunnel mirrors the logic used by main.go to rewrite
// a Postgres DSN when DB_VPS_TUNNEL_ENABLED=1 and DB_VPS_LOCAL_PORT is set.
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

func main() {
    var dsnFlag string
    var provider string
    var enabled string
    var actor string

    flag.StringVar(&dsnFlag, "dsn", "", "Postgres DSN for superadmin DB (overrides DB_SUPERADMIN_DSN env)")
    flag.StringVar(&provider, "provider", "google", "provider slug (google)")
    flag.StringVar(&enabled, "enabled", "0", "0 or 1")
    flag.StringVar(&actor, "actor", "cli-helper", "actor name to write in updated_by config")
    flag.Parse()

    dsn := dsnFlag
    if dsn == "" {
        dsn = os.Getenv("DB_SUPERADMIN_DSN")
    }
    if dsn == "" {
        log.Fatalf("DB_SUPERADMIN_DSN not set and -dsn not provided")
    }

    // Optionally rewrite DSN for local tunnel (matches main.go behavior)
    dsn = rewriteRuntimePostgresDSNForTunnel(dsn)

    db, err := sql.Open("pgx", dsn)
    if err != nil {
        log.Fatalf("open db: %v", err)
    }
    defer db.Close()
    if err := db.Ping(); err != nil {
        log.Fatalf("ping db: %v", err)
    }

    key := fmt.Sprintf("ai.provider.%s.enabled", provider)
    if err := dbpkg.SetConfigValue(db, key, enabled, false); err != nil {
        log.Fatalf("SetConfigValue %s: %v", key, err)
    }
    updatedByKey := key + ".updated_by"
    if err := dbpkg.SetConfigValue(db, updatedByKey, actor, false); err != nil {
        log.Fatalf("SetConfigValue %s: %v", updatedByKey, err)
    }

    fmt.Printf("OK: set %s=%s (actor=%s)\n", key, enabled, actor)
}
