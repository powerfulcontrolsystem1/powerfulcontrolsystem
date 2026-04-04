package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
	_ "modernc.org/sqlite"
)

func truncate(s string, n int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= n {
		return string(r)
	}
	return string(r[:n])
}

func main() {
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
	v, enc, err := dbpkg.GetConfigValue(db, cfgKey)
	if err != nil {
		fmt.Fprintln(os.Stderr, "get config error:", err)
		os.Exit(1)
	}
	if v == "" {
		fmt.Fprintln(os.Stderr, "config value empty")
		os.Exit(1)
	}
	apiKey := v
	if enc {
		if !utils.EncryptionAvailable() {
			fmt.Fprintln(os.Stderr, "CONFIG_ENC_KEY not set or invalid for decryption")
			os.Exit(1)
		}
		dec, derr := utils.DecryptString(v)
		if derr != nil {
			fmt.Fprintln(os.Stderr, "decrypt error:", derr)
			os.Exit(1)
		}
		apiKey = dec
	}

	endpoint := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"
	endpoint = endpoint + "?key=" + url.QueryEscape(apiKey)
	bodyMap := map[string]interface{}{
		"system_instruction": map[string]interface{}{
			"parts": []map[string]string{{"text": "Eres un asistente de prueba. Responde breve en Español."}},
		},
		"contents": []map[string]interface{}{
			{"role": "user", "parts": []map[string]string{{"text": "Prueba breve"}}},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0.0,
			"maxOutputTokens": 100,
		},
	}
	b, _ := json.Marshal(bodyMap)
	client := &http.Client{Timeout: 20 * time.Second}
	req, _ := http.NewRequest("POST", endpoint, strings.NewReader(string(b)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "request error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		fmt.Fprintf(os.Stderr, "provider error (%d): %s\n", resp.StatusCode, truncate(string(raw), 600))
		os.Exit(1)
	}
	fmt.Println("provider response (truncated):")
	fmt.Println(truncate(string(raw), 2000))
}
