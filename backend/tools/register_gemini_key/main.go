package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
	_ "modernc.org/sqlite"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		fmt.Fprintln(os.Stderr, "failed reading input:", err)
		os.Exit(1)
	}
	key := strings.TrimSpace(input)
	if key == "" {
		fmt.Fprintln(os.Stderr, "empty API key on stdin")
		os.Exit(1)
	}

	if !utils.EncryptionAvailable() {
		fmt.Fprintln(os.Stderr, "CONFIG_ENC_KEY not set or invalid")
		os.Exit(1)
	}

	enc, err := utils.EncryptString(key)
	if err != nil {
		fmt.Fprintln(os.Stderr, "encrypt error:", err)
		os.Exit(1)
	}

	dbPath := "db/superadministrador.db"
	if p := os.Getenv("DB_SUPERADMIN_PATH"); p != "" {
		dbPath = p
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "db open error:", err)
		os.Exit(1)
	}
	defer db.Close()

	cfgKey := "ai.model.google.gemini_2_0_flash.api_key"
	if err := dbpkg.SetConfigValue(db, cfgKey, enc, true); err != nil {
		fmt.Fprintln(os.Stderr, "db set error:", err)
		os.Exit(1)
	}

	fmt.Println("ok")
}
