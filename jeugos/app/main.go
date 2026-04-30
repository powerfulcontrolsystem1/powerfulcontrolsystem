package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type serverConfig struct {
	addr        string
	publicDir   string
	emulatorDir string
	romsDir     string
	core        string
}

type romItem struct {
	Name    string `json:"name"`
	File    string `json:"file"`
	URL     string `json:"url"`
	Size    int64  `json:"size"`
	ModTime string `json:"mod_time"`
}

type romListResponse struct {
	Core string    `json:"core"`
	ROMs []romItem `json:"roms"`
}

var allowedROMExtensions = map[string]bool{
	".sfc": true,
	".smc": true,
	".fig": true,
	".swc": true,
	".zip": true,
}

func main() {
	cfg := loadConfig()
	mustRegisterMIMETypes()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/roms", cfg.handleListROMs)
	mux.HandleFunc("/health", handleHealth)
	mux.Handle("/emulator/", http.StripPrefix("/emulator/", secureStaticFileServer(cfg.emulatorDir)))
	mux.HandleFunc("/roms/", cfg.handleROMFile)
	mux.Handle("/", secureSPAFileServer(cfg.publicDir))

	server := &http.Server{
		Addr:              cfg.addr,
		Handler:           securityHeaders(mux),
		ReadHeaderTimeout: 8 * time.Second,
	}

	log.Printf("jeugos emulator server listening on %s", cfg.addr)
	log.Printf("public=%s emulator=%s roms=%s core=%s", cfg.publicDir, cfg.emulatorDir, cfg.romsDir, cfg.core)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func loadConfig() serverConfig {
	addr := flag.String("addr", envDefault("JEUGOS_ADDR", ":8099"), "HTTP bind address")
	publicDir := flag.String("public", envDefault("JEUGOS_PUBLIC_DIR", "./public"), "public static directory")
	emulatorDir := flag.String("emulator", envDefault("JEUGOS_EMULATOR_DIR", "./emulator"), "EmulatorJS static directory")
	romsDir := flag.String("roms", envDefault("JEUGOS_ROMS_DIR", "./roms"), "read-only ROM directory")
	core := flag.String("core", envDefault("JEUGOS_CORE", "snes"), "EmulatorJS core")
	flag.Parse()

	return serverConfig{
		addr:        *addr,
		publicDir:   cleanRoot(*publicDir),
		emulatorDir: cleanRoot(*emulatorDir),
		romsDir:     cleanRoot(*romsDir),
		core:        strings.TrimSpace(*core),
	}
}

func envDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func cleanRoot(root string) string {
	abs, err := filepath.Abs(root)
	if err != nil {
		return filepath.Clean(root)
	}
	return abs
}

func mustRegisterMIMETypes() {
	_ = mime.AddExtensionType(".wasm", "application/wasm")
	_ = mime.AddExtensionType(".data", "application/octet-stream")
	_ = mime.AddExtensionType(".sfc", "application/octet-stream")
	_ = mime.AddExtensionType(".smc", "application/octet-stream")
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (cfg serverConfig) handleListROMs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	items := make([]romItem, 0)
	entries, err := os.ReadDir(cfg.romsDir)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusOK, romListResponse{Core: cfg.core, ROMs: items})
			return
		}
		http.Error(w, "could not read roms", http.StatusInternalServerError)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !isAllowedROMFile(name) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		safeName, ok := sanitizeROMName(name)
		if !ok {
			continue
		}
		items = append(items, romItem{
			Name:    displayROMName(safeName),
			File:    safeName,
			URL:     "/roms/" + url.PathEscape(safeName),
			Size:    info.Size(),
			ModTime: info.ModTime().UTC().Format(time.RFC3339),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})
	writeJSON(w, http.StatusOK, romListResponse{Core: cfg.core, ROMs: items})
}

func (cfg serverConfig) handleROMFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawName := strings.TrimPrefix(r.URL.Path, "/roms/")
	rawName, err := url.PathUnescape(rawName)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	name, ok := sanitizeROMName(rawName)
	if !ok || !isAllowedROMFile(name) {
		http.NotFound(w, r)
		return
	}

	fullPath, ok := safeJoin(cfg.romsDir, name)
	if !ok {
		http.NotFound(w, r)
		return
	}
	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeFile(w, r, fullPath)
}

func secureStaticFileServer(root string) http.Handler {
	return http.FileServer(http.Dir(root))
}

func secureSPAFileServer(root string) http.HandlerFunc {
	fs := http.FileServer(http.Dir(root))
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		cleanPath := path.Clean("/" + r.URL.Path)
		if strings.Contains(cleanPath, "..") {
			http.NotFound(w, r)
			return
		}
		fullPath, ok := safeJoin(root, strings.TrimPrefix(cleanPath, "/"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(root, "index.html"))
	}
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}

func sanitizeROMName(raw string) (string, bool) {
	name := strings.TrimSpace(raw)
	if name == "" || strings.Contains(name, "\x00") {
		return "", false
	}
	if name != filepath.Base(name) || name != path.Base(name) {
		return "", false
	}
	if strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) {
		return "", false
	}
	return name, true
}

func isAllowedROMFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return allowedROMExtensions[ext]
}

func displayROMName(file string) string {
	name := strings.TrimSuffix(file, filepath.Ext(file))
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")
	return strings.TrimSpace(name)
}

func safeJoin(root, name string) (string, bool) {
	rootAbs := cleanRoot(root)
	full := filepath.Join(rootAbs, filepath.Clean(name))
	fullAbs := cleanRoot(full)
	rel, err := filepath.Rel(rootAbs, fullAbs)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", false
	}
	return fullAbs, true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("json encode error: %v", err)
	}
}
