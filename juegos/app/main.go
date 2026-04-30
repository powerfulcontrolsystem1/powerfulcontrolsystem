package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
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
	savesDir    string
	core        string
}

type romItem struct {
	Name    string `json:"name"`
	File    string `json:"file"`
	URL     string `json:"url"`
	Core    string `json:"core"`
	System  string `json:"system"`
	Size    int64  `json:"size"`
	ModTime string `json:"mod_time"`
}

type romListResponse struct {
	Core string    `json:"core"`
	ROMs []romItem `json:"roms"`
}

type romKind struct {
	core   string
	system string
}

type saveUploadRequest struct {
	EmpresaID        string `json:"empresa_id"`
	ROM              string `json:"rom"`
	Core             string `json:"core"`
	DataBase64       string `json:"data_base64"`
	ScreenshotBase64 string `json:"screenshot_base64"`
	Source           string `json:"source"`
}

type saveMetadata struct {
	EmpresaID string `json:"empresa_id"`
	ROM       string `json:"rom"`
	Core      string `json:"core"`
	Kind      string `json:"kind"`
	Bytes     int    `json:"bytes"`
	Source    string `json:"source"`
	UpdatedAt string `json:"updated_at"`
}

type saveLatestResponse struct {
	OK        bool   `json:"ok"`
	Exists    bool   `json:"exists"`
	EmpresaID string `json:"empresa_id"`
	ROM       string `json:"rom"`
	StateURL  string `json:"state_url,omitempty"`
	SaveURL   string `json:"save_url,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

var allowedROMExtensions = map[string]romKind{
	".sfc": {core: "snes", system: "SNES"},
	".smc": {core: "snes", system: "SNES"},
	".fig": {core: "snes", system: "SNES"},
	".swc": {core: "snes", system: "SNES"},
	".nes": {core: "nes", system: "NES"},
	".gb":  {core: "gb", system: "Game Boy"},
	".gbc": {core: "gb", system: "Game Boy Color"},
	".gba": {core: "gba", system: "Game Boy Advance"},
	".gen": {core: "segaMD", system: "Mega Drive"},
	".bin": {core: "segaMD", system: "Mega Drive"},
	".zip": {core: "", system: "ZIP"},
}

func main() {
	cfg := loadConfig()
	mustRegisterMIMETypes()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/roms", cfg.handleListROMs)
	mux.HandleFunc("/api/saves/latest", cfg.handleSaveLatest)
	mux.HandleFunc("/api/saves/state", cfg.handleSaveState)
	mux.HandleFunc("/api/saves/file", cfg.handleSaveFile)
	mux.HandleFunc("/health", handleHealth)
	mux.Handle("/emulator/", http.StripPrefix("/emulator/", secureStaticFileServer(cfg.emulatorDir)))
	mux.HandleFunc("/roms/", cfg.handleROMFile)
	mux.Handle("/", secureSPAFileServer(cfg.publicDir))

	server := &http.Server{
		Addr:              cfg.addr,
		Handler:           securityHeaders(mux),
		ReadHeaderTimeout: 8 * time.Second,
	}

	log.Printf("juegos emulator server listening on %s", cfg.addr)
	log.Printf("public=%s emulator=%s roms=%s saves=%s core=%s", cfg.publicDir, cfg.emulatorDir, cfg.romsDir, cfg.savesDir, cfg.core)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func loadConfig() serverConfig {
	addr := flag.String("addr", envDefault("JUEGOS_ADDR", ":8099"), "HTTP bind address")
	publicDir := flag.String("public", envDefault("JUEGOS_PUBLIC_DIR", "./public"), "public static directory")
	emulatorDir := flag.String("emulator", envDefault("JUEGOS_EMULATOR_DIR", "./emulator"), "EmulatorJS static directory")
	romsDir := flag.String("roms", envDefault("JUEGOS_ROMS_DIR", "./roms"), "read-only ROM directory")
	savesDir := flag.String("saves", envDefault("JUEGOS_SAVES_DIR", "./empresas"), "per-company save directory")
	core := flag.String("core", envDefault("JUEGOS_CORE", "snes"), "EmulatorJS core")
	flag.Parse()

	return serverConfig{
		addr:        *addr,
		publicDir:   cleanRoot(*publicDir),
		emulatorDir: cleanRoot(*emulatorDir),
		romsDir:     cleanRoot(*romsDir),
		savesDir:    cleanRoot(*savesDir),
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
	_ = mime.AddExtensionType(".js", "application/javascript; charset=utf-8")
	_ = mime.AddExtensionType(".mjs", "application/javascript; charset=utf-8")
	_ = mime.AddExtensionType(".css", "text/css; charset=utf-8")
	_ = mime.AddExtensionType(".wasm", "application/wasm")
	_ = mime.AddExtensionType(".data", "application/octet-stream")
	_ = mime.AddExtensionType(".sfc", "application/octet-stream")
	_ = mime.AddExtensionType(".smc", "application/octet-stream")
	_ = mime.AddExtensionType(".nes", "application/octet-stream")
	_ = mime.AddExtensionType(".gb", "application/octet-stream")
	_ = mime.AddExtensionType(".gbc", "application/octet-stream")
	_ = mime.AddExtensionType(".gba", "application/octet-stream")
	_ = mime.AddExtensionType(".md", "application/octet-stream")
	_ = mime.AddExtensionType(".gen", "application/octet-stream")
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
		kind, ok := romKindForFile(name, cfg.core)
		if !ok {
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
			Core:    kind.core,
			System:  kind.system,
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

func (cfg serverConfig) handleSaveLatest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	empresaID, rom, ok := saveContextFromQuery(r)
	if !ok {
		http.Error(w, "invalid save context", http.StatusBadRequest)
		return
	}

	statePath, ok := cfg.saveFilePath(empresaID, rom, "latest.state")
	if !ok {
		http.Error(w, "invalid save context", http.StatusBadRequest)
		return
	}
	savePath, _ := cfg.saveFilePath(empresaID, rom, "latest.save")

	resp := saveLatestResponse{OK: true, EmpresaID: empresaID, ROM: rom}
	if info, err := os.Stat(statePath); err == nil && !info.IsDir() {
		resp.Exists = true
		resp.StateURL = saveDownloadURL("/api/saves/state", empresaID, rom)
		resp.UpdatedAt = info.ModTime().UTC().Format(time.RFC3339)
	}
	if info, err := os.Stat(savePath); err == nil && !info.IsDir() {
		resp.Exists = true
		resp.SaveURL = saveDownloadURL("/api/saves/file", empresaID, rom)
		if resp.UpdatedAt == "" || info.ModTime().UTC().Format(time.RFC3339) > resp.UpdatedAt {
			resp.UpdatedAt = info.ModTime().UTC().Format(time.RFC3339)
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

func (cfg serverConfig) handleSaveState(w http.ResponseWriter, r *http.Request) {
	cfg.handleSaveFileKind(w, r, "state", "latest.state", "application/octet-stream")
}

func (cfg serverConfig) handleSaveFile(w http.ResponseWriter, r *http.Request) {
	cfg.handleSaveFileKind(w, r, "save", "latest.save", "application/octet-stream")
}

func (cfg serverConfig) handleSaveFileKind(w http.ResponseWriter, r *http.Request, kind, filename, contentType string) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		empresaID, rom, ok := saveContextFromQuery(r)
		if !ok {
			http.NotFound(w, r)
			return
		}
		fullPath, ok := cfg.saveFilePath(empresaID, rom, filename)
		if !ok {
			http.NotFound(w, r)
			return
		}
		info, err := os.Stat(fullPath)
		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "no-store")
		http.ServeFile(w, r, fullPath)
	case http.MethodPost:
		cfg.handleSaveUpload(w, r, kind, filename)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (cfg serverConfig) handleSaveUpload(w http.ResponseWriter, r *http.Request, kind, filename string) {
	r.Body = http.MaxBytesReader(w, r.Body, 96*1024*1024)
	defer r.Body.Close()

	var payload saveUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json payload", http.StatusBadRequest)
		return
	}

	empresaID, ok := sanitizeEmpresaID(payload.EmpresaID)
	if !ok {
		http.Error(w, "invalid empresa_id", http.StatusBadRequest)
		return
	}
	rom, ok := sanitizeROMName(payload.ROM)
	if !ok || !isAllowedROMFile(rom) {
		http.Error(w, "invalid rom", http.StatusBadRequest)
		return
	}
	data, err := decodeBase64Payload(payload.DataBase64, 64*1024*1024)
	if err != nil {
		http.Error(w, "invalid save data", http.StatusBadRequest)
		return
	}

	fullPath, ok := cfg.saveFilePath(empresaID, rom, filename)
	if !ok {
		http.Error(w, "invalid save path", http.StatusBadRequest)
		return
	}
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o750); err != nil {
		http.Error(w, "could not create save directory", http.StatusInternalServerError)
		return
	}
	if err := writeFileAtomic(fullPath, data, 0o640); err != nil {
		http.Error(w, "could not write save", http.StatusInternalServerError)
		return
	}

	if payload.ScreenshotBase64 != "" {
		if shot, err := decodeBase64Payload(payload.ScreenshotBase64, 10*1024*1024); err == nil && len(shot) > 0 {
			shotPath, ok := cfg.saveFilePath(empresaID, rom, "latest.png")
			if ok {
				_ = writeFileAtomic(shotPath, shot, 0o640)
			}
		}
	}

	meta := saveMetadata{
		EmpresaID: empresaID,
		ROM:       rom,
		Core:      sanitizeSmallText(payload.Core, 32),
		Kind:      kind,
		Bytes:     len(data),
		Source:    sanitizeSmallText(payload.Source, 64),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	metaPath, ok := cfg.saveFilePath(empresaID, rom, "meta.json")
	if ok {
		if encoded, err := json.MarshalIndent(meta, "", "  "); err == nil {
			_ = writeFileAtomic(metaPath, encoded, 0o640)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":         true,
		"empresa_id": empresaID,
		"rom":        rom,
		"kind":       kind,
		"bytes":      len(data),
		"updated_at": meta.UpdatedAt,
	})
}

func secureStaticFileServer(root string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		cleanPath := path.Clean("/" + strings.TrimPrefix(r.URL.Path, "/"))
		if strings.Contains(cleanPath, "..") {
			http.NotFound(w, r)
			return
		}
		fullPath, ok := safeJoin(root, strings.TrimPrefix(cleanPath, "/"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		info, err := os.Stat(fullPath)
		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=604800")
		http.ServeFile(w, r, fullPath)
	})
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
		if cleanPath == "/" {
			http.ServeFile(w, r, filepath.Join(root, "index.html"))
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
	_, ok := allowedROMExtensions[ext]
	return ok
}

func romKindForFile(name, fallbackCore string) (romKind, bool) {
	ext := strings.ToLower(filepath.Ext(name))
	kind, ok := allowedROMExtensions[ext]
	if !ok {
		return romKind{}, false
	}
	if kind.core == "" {
		kind.core = strings.TrimSpace(fallbackCore)
	}
	if kind.core == "" {
		kind.core = "snes"
	}
	return kind, true
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
	if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", false
	}
	return fullAbs, true
}

func saveContextFromQuery(r *http.Request) (string, string, bool) {
	empresaID, ok := sanitizeEmpresaID(r.URL.Query().Get("empresa_id"))
	if !ok {
		return "", "", false
	}
	rom, ok := sanitizeROMName(r.URL.Query().Get("rom"))
	if !ok || !isAllowedROMFile(rom) {
		return "", "", false
	}
	return empresaID, rom, true
}

func (cfg serverConfig) saveFilePath(empresaID, rom, filename string) (string, bool) {
	empresaID, ok := sanitizeEmpresaID(empresaID)
	if !ok {
		return "", false
	}
	rom, ok = sanitizeROMName(rom)
	if !ok || !isAllowedROMFile(rom) {
		return "", false
	}
	file, ok := sanitizeSaveFilename(filename)
	if !ok {
		return "", false
	}
	romFolder := sanitizeStorageSegment(strings.TrimSuffix(rom, filepath.Ext(rom)))
	if romFolder == "" {
		return "", false
	}
	rel := filepath.Join("empresa_"+empresaID, "emulador", romFolder, file)
	return safeJoin(cfg.savesDir, rel)
}

func sanitizeEmpresaID(raw string) (string, bool) {
	value := strings.TrimSpace(raw)
	if value == "" || value == "0" {
		return "publico", true
	}
	if strings.EqualFold(value, "publico") || strings.EqualFold(value, "public") {
		return "publico", true
	}
	if len(value) > 20 {
		return "", false
	}
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return "", false
		}
	}
	return value, true
}

func sanitizeSaveFilename(raw string) (string, bool) {
	name := strings.TrimSpace(raw)
	if name == "" || name != filepath.Base(name) || strings.ContainsAny(name, `/\`) || strings.Contains(name, "..") {
		return "", false
	}
	switch name {
	case "latest.state", "latest.save", "latest.png", "meta.json":
		return name, true
	default:
		return "", false
	}
}

func sanitizeStorageSegment(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	var builder strings.Builder
	lastDash := false
	for _, ch := range value {
		allowed := (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')
		if allowed {
			builder.WriteRune(ch)
			lastDash = false
			continue
		}
		if !lastDash {
			builder.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(builder.String(), "-")
}

func sanitizeSmallText(raw string, max int) string {
	value := strings.TrimSpace(raw)
	value = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '\t' || r == 0 {
			return -1
		}
		return r
	}, value)
	if len(value) > max {
		return value[:max]
	}
	return value
}

func decodeBase64Payload(raw string, maxBytes int) ([]byte, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, fmt.Errorf("empty payload")
	}
	if idx := strings.Index(value, ","); strings.HasPrefix(strings.ToLower(value), "data:") && idx >= 0 {
		value = value[idx+1:]
	}
	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		data, err = base64.RawStdEncoding.DecodeString(value)
	}
	if err != nil {
		return nil, err
	}
	if len(data) == 0 || len(data) > maxBytes {
		return nil, fmt.Errorf("payload size out of range")
	}
	return data, nil
}

func writeFileAtomic(target string, data []byte, perm os.FileMode) error {
	tmp := target + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	return os.Rename(tmp, target)
}

func saveDownloadURL(endpoint, empresaID, rom string) string {
	values := url.Values{}
	values.Set("empresa_id", empresaID)
	values.Set("rom", rom)
	return endpoint + "?" + values.Encode()
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("json encode error: %v", err)
	}
}
