package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func hashPassword(password, salt string) string {
	sum := sha256.Sum256([]byte(salt + ":" + password))
	return hex.EncodeToString(sum[:])
}

func main() {
	dsn := strings.TrimSpace(os.Getenv("DB_SUPERADMIN_DSN"))
	if dsn == "" {
		log.Fatal("DB_SUPERADMIN_DSN is required")
	}
	email := strings.TrimSpace(os.Getenv("ADMIN_EMAIL"))
	if email == "" {
		email = "powerfulcontrolsystem@gmail.com"
	}
	candidatesEnv := strings.TrimSpace(os.Getenv("ADMIN_PASSWORD_CANDIDATES"))
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	admin, err := dbpkg.GetAdminByEmailFull(db, email)
	if err != nil {
		log.Fatalf("query admin via dbpkg: %v", err)
	}

	fmt.Printf("email=%s\n", email)
	fmt.Printf("id=%d name=%s role=%s estado=%s email_confirmado=%d password_set=%d salt_len=%d hash_len=%d\n", admin.ID, admin.Name, admin.Role, admin.Estado, admin.EmailConfirmado, admin.PasswordSet, len(admin.PasswordSalt), len(admin.PasswordHash))

	if candidatesEnv == "" {
		fmt.Println("ADMIN_PASSWORD_CANDIDATES is empty; not testing any password candidates.")
		fmt.Println("Set ADMIN_PASSWORD_CANDIDATES as comma-separated values to test candidates (do not commit real passwords).")
		return
	}
	for _, candidate := range strings.Split(candidatesEnv, ",") {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		match := hashPassword(candidate, admin.PasswordSalt) == admin.PasswordHash
		fmt.Printf("candidate=%q match=%t\n", candidate, match)
	}
	if admin.PasswordSalt == "" {
		fmt.Println("password_salt is empty")
	}
}