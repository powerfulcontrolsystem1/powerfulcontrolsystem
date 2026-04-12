package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
	_ "modernc.org/sqlite"
)

func loadEnvLocal(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	for _, l := range lines {
		if strings.TrimSpace(l) == "" || strings.HasPrefix(strings.TrimSpace(l), "#") {
			continue
		}
		if idx := strings.Index(l, "="); idx > 0 {
			k := strings.TrimSpace(l[:idx])
			v := strings.TrimSpace(l[idx+1:])
			if strings.HasPrefix(v, "\"") && strings.HasSuffix(v, "\"") {
				v = v[1 : len(v)-1]
			}
			if os.Getenv(k) == "" && v != "" {
				os.Setenv(k, v)
			}
		}
	}
}

func maskKey(k string) string {
	if k == "" {
		return ""
	}
	if len(k) <= 12 {
		return "********"
	}
	return k[:4] + "..." + k[len(k)-4:]
}

func main() {
	// cargar CONFIG_ENC_KEY desde .env.local si existe (ejecucion en backend/)
	loadEnvLocal(".env.local")

	dbPath := "superadministrador.db"
	if p := strings.TrimSpace(os.Getenv("DB_SUPERADMIN_PATH")); p != "" {
		dbPath = p
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		fmt.Printf("ERROR: no se pudo abrir DB %s: %v\n", dbPath, err)
		os.Exit(2)
	}
	defer db.Close()

	keys := []string{"ai.model.deepseek.deepseek_chat.api_key", "ai.provider.deepseek.api_key"}
	var apiKey string
	for _, k := range keys {
		v, enc, err := dbpkg.GetConfigValue(db, k)
		if err != nil || strings.TrimSpace(v) == "" {
			continue
		}
		if enc {
			dec, derr := utils.DecryptString(v)
			if derr != nil {
				fmt.Printf("found %s but decrypt failed: %v\n", k, derr)
				continue
			}
			apiKey = strings.TrimSpace(dec)
			fmt.Printf("Found key in DB %s (encrypted).\n", k)
			break
		}
		apiKey = strings.TrimSpace(v)
		fmt.Printf("Found key in DB %s (plain).\n", k)
		break
	}

	if apiKey == "" {
		// fallback env
		if ev := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")); ev != "" {
			apiKey = ev
			fmt.Printf("Found key in ENV DEEPSEEK_API_KEY.\n")
		}
	}

	if apiKey == "" {
		fmt.Println("No se encontro clave DeepSeek en DB ni en DEEPSEEK_API_KEY")
		fmt.Println("Revisando claves relacionadas en la tabla configuraciones:")
		rows, qerr := db.Query("SELECT config_key, encrypted FROM configuraciones WHERE config_key LIKE '%deepseek%' OR config_key LIKE 'ai.%' LIMIT 200")
		if qerr == nil {
			defer rows.Close()
			for rows.Next() {
				var k string
				var enc sql.NullInt64
				if err := rows.Scan(&k, &enc); err == nil {
					e := 0
					if enc.Valid {
						e = int(enc.Int64)
					}
					fmt.Printf(" - %s (encrypted=%d)\n", k, e)
				}
			}
		} else {
			fmt.Printf("  (error listando configuraciones: %v)\n", qerr)
		}

		if ev := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")); ev != "" {
			fmt.Printf("DEEPSEEK_API_KEY en entorno: %s\n", maskKey(ev))
		}

		os.Exit(3)
	}

	fmt.Printf("Validando clave DeepSeek: %s\n", maskKey(apiKey))

	// Preparar solicitud de verificacion al endpoint de DeepSeek
	body := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": "Verificacion de clave - responder brevemente."},
			{"role": "user", "content": "Hola"},
		},
		"temperature": 0.2,
		"max_tokens":  10,
		"stream":      false,
	}
	b, _ := json.Marshal(body)

	client := &http.Client{Timeout: 20 * time.Second}
	req, _ := http.NewRequest(http.MethodPost, "https://api.deepseek.com/chat/completions", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("ERROR: no se pudo contactar DeepSeek: %v\n", err)
		os.Exit(4)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)

	code := resp.StatusCode
	switch code {
	case 200:
		fmt.Println("OK: clave valida (200). Respuesta truncada:")
		fmt.Println(string(raw)[:min(len(raw), 800)])
		os.Exit(0)
	case 402:
		fmt.Println("INSUFFICIENT BALANCE (402): la clave es valida pero la cuenta no tiene saldo suficiente.")
		fmt.Printf("Respuesta proveedor: %s\n", truncate(string(raw), 800))
		os.Exit(0)
	case 401:
		fmt.Println("UNAUTHORIZED (401): clave invalida o no autorizada.")
		fmt.Printf("Respuesta proveedor: %s\n", truncate(string(raw), 800))
		os.Exit(5)
	default:
		fmt.Printf("Respuesta HTTP %d del proveedor: %s\n", code, truncate(string(raw), 1000))
		os.Exit(6)
	}
}

func truncate(s string, max int) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
