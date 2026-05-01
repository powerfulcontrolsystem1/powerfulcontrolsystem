package main

import (
	"encoding/base64"
	"encoding/json"
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

type localEmulatorROMKind struct {
	core   string
	system string
}

type localEmulatorROMItem struct {
	Name    string `json:"name"`
	File    string `json:"file"`
	URL     string `json:"url"`
	Core    string `json:"core"`
	System  string `json:"system"`
	Size    int64  `json:"size"`
	ModTime string `json:"mod_time"`
}

type localEmulatorROMListResponse struct {
	Core string                 `json:"core"`
	ROMs []localEmulatorROMItem `json:"roms"`
}

type localEmulatorSaveUploadRequest struct {
	EmpresaID        string `json:"empresa_id"`
	ROM              string `json:"rom"`
	Core             string `json:"core"`
	DataBase64       string `json:"data_base64"`
	ScreenshotBase64 string `json:"screenshot_base64"`
	Source           string `json:"source"`
}

var localEmulatorAllowedROMExtensions = map[string]localEmulatorROMKind{
	".sfc": {core: "snes", system: "SNES"},
	".smc": {core: "snes", system: "SNES"},
	".fig": {core: "snes", system: "SNES"},
	".swc": {core: "snes", system: "SNES"},
	".nes": {core: "nes", system: "NES"},
	".n64": {core: "n64", system: "Nintendo 64"},
	".z64": {core: "n64", system: "Nintendo 64"},
	".v64": {core: "n64", system: "Nintendo 64"},
	".gb":  {core: "gb", system: "Game Boy"},
	".gbc": {core: "gb", system: "Game Boy Color"},
	".gba": {core: "gba", system: "Game Boy Advance"},
	".gen": {core: "segaMD", system: "Mega Drive"},
	".bin": {core: "segaMD", system: "Mega Drive"},
	".zip": {core: "snes", system: "ZIP"},
}

func registerLocalEmulatorRoutes(backendDir, webDir string) {
	juegosDir := resolveJuegosDir(backendDir, webDir)
	if juegosDir == "" {
		log.Printf("warning: carpeta juegos no encontrada; /emulador/ mostrara pagina no disponible")
	}
	mustRegisterLocalEmulatorMIMETypes()
	handler := localEmulatorHandler(juegosDir)
	http.HandleFunc("/emulador", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/emulador/", http.StatusMovedPermanently)
	})
	http.Handle("/emulador/", http.StripPrefix("/emulador", handler))
}

func resolveJuegosDir(backendDir, webDir string) string {
	candidates := []string{
		os.Getenv("PCS_JUEGOS_DIR"),
		filepath.Join(backendDir, "..", "juegos"),
		filepath.Join(webDir, "..", "juegos"),
		"juegos",
		filepath.Join("..", "juegos"),
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, "juegos"), filepath.Join(wd, "..", "juegos"))
	}
	seen := map[string]bool{}
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		abs, err := filepath.Abs(candidate)
		if err != nil {
			abs = filepath.Clean(candidate)
		}
		if seen[abs] {
			continue
		}
		seen[abs] = true
		if info, err := os.Stat(filepath.Join(abs, "public", "index.html")); err == nil && !info.IsDir() {
			return abs
		}
	}
	return ""
}

func mustRegisterLocalEmulatorMIMETypes() {
	_ = mime.AddExtensionType(".js", "application/javascript; charset=utf-8")
	_ = mime.AddExtensionType(".mjs", "application/javascript; charset=utf-8")
	_ = mime.AddExtensionType(".css", "text/css; charset=utf-8")
	_ = mime.AddExtensionType(".wasm", "application/wasm")
	_ = mime.AddExtensionType(".data", "application/octet-stream")
	for ext := range localEmulatorAllowedROMExtensions {
		_ = mime.AddExtensionType(ext, "application/octet-stream")
	}
}

func localEmulatorHandler(juegosDir string) http.Handler {
	publicDir := filepath.Join(juegosDir, "public")
	emulatorDir := filepath.Join(juegosDir, "emulator")
	romsDir := filepath.Join(juegosDir, "roms")
	savesDir := filepath.Join(juegosDir, "empresas")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if juegosDir == "" {
			localEmulatorUnavailable(w)
			return
		}

		cleanPath := path.Clean("/" + strings.TrimPrefix(r.URL.Path, "/"))
		switch {
		case cleanPath == "/" || cleanPath == "/index.html":
			if r.Method == http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			http.ServeFile(w, r, filepath.Join(publicDir, "index.html"))
		case cleanPath == "/api/roms":
			if r.Method == http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleLocalEmulatorListROMs(w, r, romsDir)
		case cleanPath == "/api/saves/latest":
			handleLocalEmulatorSaveLatest(w, r, savesDir)
		case cleanPath == "/api/saves/state":
			handleLocalEmulatorSaveBlob(w, r, savesDir, "state", "latest.state")
		case cleanPath == "/api/saves/file":
			handleLocalEmulatorSaveBlob(w, r, savesDir, "save", "latest.save")
		case strings.HasPrefix(cleanPath, "/emulator/"):
			if r.Method == http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			serveLocalEmulatorFile(w, r, emulatorDir, strings.TrimPrefix(cleanPath, "/emulator/"), 7*24*time.Hour)
		case strings.HasPrefix(cleanPath, "/roms/"):
			if r.Method == http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			name, err := url.PathUnescape(strings.TrimPrefix(cleanPath, "/roms/"))
			if err != nil || !localEmulatorIsAllowedROMFile(name) {
				http.NotFound(w, r)
				return
			}
			serveLocalEmulatorFile(w, r, romsDir, name, 24*time.Hour)
		default:
			if r.Method == http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			serveLocalEmulatorFile(w, r, publicDir, strings.TrimPrefix(cleanPath, "/"), time.Hour)
		}
	})
}

func handleLocalEmulatorListROMs(w http.ResponseWriter, r *http.Request, romsDir string) {
	items := make([]localEmulatorROMItem, 0)
	entries, err := os.ReadDir(romsDir)
	if err != nil {
		writeLocalEmulatorJSON(w, http.StatusOK, localEmulatorROMListResponse{Core: "snes", ROMs: items})
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		kind, ok := localEmulatorROMKindForFile(name)
		if !ok {
			continue
		}
		safeName, ok := localEmulatorSanitizeFilename(name)
		if !ok {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		items = append(items, localEmulatorROMItem{
			Name:    localEmulatorDisplayName(safeName),
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
	writeLocalEmulatorJSON(w, http.StatusOK, localEmulatorROMListResponse{Core: "snes", ROMs: items})
}

func serveLocalEmulatorFile(w http.ResponseWriter, r *http.Request, root, rawName string, maxAge time.Duration) {
	name, ok := localEmulatorSanitizePath(rawName)
	if !ok {
		http.NotFound(w, r)
		return
	}
	full, ok := localEmulatorSafeJoin(root, name)
	if !ok {
		http.NotFound(w, r)
		return
	}
	info, err := os.Stat(full)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}
	if maxAge > 0 {
		w.Header().Set("Cache-Control", "public, max-age="+strconvItoa(int(maxAge.Seconds())))
	}
	http.ServeFile(w, r, full)
}

func handleLocalEmulatorSaveLatest(w http.ResponseWriter, r *http.Request, savesDir string) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	empresaID, rom, ok := localEmulatorSaveContext(r)
	if !ok {
		http.Error(w, "invalid save context", http.StatusBadRequest)
		return
	}
	statePath, ok := localEmulatorSaveFilePath(savesDir, empresaID, rom, "latest.state")
	if !ok {
		http.Error(w, "invalid save context", http.StatusBadRequest)
		return
	}
	savePath, _ := localEmulatorSaveFilePath(savesDir, empresaID, rom, "latest.save")
	response := map[string]any{"ok": true, "exists": false, "empresa_id": empresaID, "rom": rom}
	if info, err := os.Stat(statePath); err == nil && !info.IsDir() {
		response["exists"] = true
		response["state_url"] = "/api/saves/state?empresa_id=" + url.QueryEscape(empresaID) + "&rom=" + url.QueryEscape(rom)
		response["updated_at"] = info.ModTime().UTC().Format(time.RFC3339)
	}
	if info, err := os.Stat(savePath); err == nil && !info.IsDir() {
		response["exists"] = true
		response["save_url"] = "/api/saves/file?empresa_id=" + url.QueryEscape(empresaID) + "&rom=" + url.QueryEscape(rom)
		if response["updated_at"] == nil {
			response["updated_at"] = info.ModTime().UTC().Format(time.RFC3339)
		}
	}
	writeLocalEmulatorJSON(w, http.StatusOK, response)
}

func handleLocalEmulatorSaveBlob(w http.ResponseWriter, r *http.Request, savesDir, kind, filename string) {
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		empresaID, rom, ok := localEmulatorSaveContext(r)
		if !ok {
			http.NotFound(w, r)
			return
		}
		full, ok := localEmulatorSaveFilePath(savesDir, empresaID, rom, filename)
		if !ok {
			http.NotFound(w, r)
			return
		}
		info, err := os.Stat(full)
		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/octet-stream")
		http.ServeFile(w, r, full)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 96*1024*1024)
	defer r.Body.Close()
	var payload localEmulatorSaveUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json payload", http.StatusBadRequest)
		return
	}
	empresaID, ok := localEmulatorSanitizeEmpresaID(payload.EmpresaID)
	if !ok {
		http.Error(w, "invalid empresa_id", http.StatusBadRequest)
		return
	}
	rom, ok := localEmulatorSanitizeFilename(payload.ROM)
	if !ok || !localEmulatorIsAllowedROMFile(rom) {
		http.Error(w, "invalid rom", http.StatusBadRequest)
		return
	}
	data, err := localEmulatorDecodeBase64Payload(payload.DataBase64, 64*1024*1024)
	if err != nil {
		http.Error(w, "invalid save data", http.StatusBadRequest)
		return
	}
	full, ok := localEmulatorSaveFilePath(savesDir, empresaID, rom, filename)
	if !ok {
		http.Error(w, "invalid save path", http.StatusBadRequest)
		return
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o750); err != nil {
		http.Error(w, "could not create save directory", http.StatusInternalServerError)
		return
	}
	if err := localEmulatorWriteFileAtomic(full, data, 0o640); err != nil {
		http.Error(w, "could not write save", http.StatusInternalServerError)
		return
	}
	metaPath, ok := localEmulatorSaveFilePath(savesDir, empresaID, rom, "meta.json")
	if ok {
		meta := map[string]any{
			"empresa_id": empresaID,
			"rom":        rom,
			"core":       strings.TrimSpace(payload.Core),
			"kind":       kind,
			"bytes":      len(data),
			"source":     strings.TrimSpace(payload.Source),
			"updated_at": time.Now().UTC().Format(time.RFC3339),
		}
		if encoded, err := json.MarshalIndent(meta, "", "  "); err == nil {
			_ = localEmulatorWriteFileAtomic(metaPath, encoded, 0o640)
		}
	}
	writeLocalEmulatorJSON(w, http.StatusOK, map[string]any{"ok": true, "empresa_id": empresaID, "rom": rom, "kind": kind, "bytes": len(data)})
}

func localEmulatorSafeJoin(root, name string) (string, bool) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", false
	}
	full := filepath.Join(rootAbs, filepath.Clean(name))
	fullAbs, err := filepath.Abs(full)
	if err != nil {
		return "", false
	}
	rel, err := filepath.Rel(rootAbs, fullAbs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", false
	}
	return fullAbs, true
}

func localEmulatorSanitizeFilename(raw string) (string, bool) {
	name := strings.TrimSpace(raw)
	if name == "" || name != filepath.Base(name) || name != path.Base(name) || strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) {
		return "", false
	}
	return name, true
}

func localEmulatorSanitizePath(raw string) (string, bool) {
	name := strings.TrimSpace(raw)
	if name == "" || strings.Contains(name, "\x00") {
		return "", false
	}
	clean := path.Clean("/" + strings.TrimPrefix(name, "/"))
	if clean == "/" || strings.Contains(clean, "..") {
		return "", false
	}
	return strings.TrimPrefix(clean, "/"), true
}

func localEmulatorIsAllowedROMFile(name string) bool {
	_, ok := localEmulatorROMKindForFile(name)
	return ok
}

func localEmulatorROMKindForFile(name string) (localEmulatorROMKind, bool) {
	kind, ok := localEmulatorAllowedROMExtensions[strings.ToLower(filepath.Ext(name))]
	return kind, ok
}

func localEmulatorDisplayName(file string) string {
	name := strings.TrimSuffix(file, filepath.Ext(file))
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")
	return strings.TrimSpace(name)
}

func localEmulatorSaveContext(r *http.Request) (string, string, bool) {
	empresaID, ok := localEmulatorSanitizeEmpresaID(r.URL.Query().Get("empresa_id"))
	if !ok {
		return "", "", false
	}
	rom, ok := localEmulatorSanitizeFilename(r.URL.Query().Get("rom"))
	if !ok || !localEmulatorIsAllowedROMFile(rom) {
		return "", "", false
	}
	return empresaID, rom, true
}

func localEmulatorSanitizeEmpresaID(raw string) (string, bool) {
	value := strings.TrimSpace(raw)
	if value == "" || value == "0" || strings.EqualFold(value, "publico") || strings.EqualFold(value, "public") {
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

func localEmulatorSaveFilePath(savesDir, empresaID, rom, filename string) (string, bool) {
	empresaID, ok := localEmulatorSanitizeEmpresaID(empresaID)
	if !ok {
		return "", false
	}
	rom, ok = localEmulatorSanitizeFilename(rom)
	if !ok || !localEmulatorIsAllowedROMFile(rom) {
		return "", false
	}
	if filename != "latest.state" && filename != "latest.save" && filename != "meta.json" {
		return "", false
	}
	romFolder := localEmulatorStorageSegment(strings.TrimSuffix(rom, filepath.Ext(rom)))
	if romFolder == "" {
		return "", false
	}
	return localEmulatorSafeJoin(savesDir, filepath.Join("empresa_"+empresaID, "emulador", romFolder, filename))
}

func localEmulatorStorageSegment(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	var builder strings.Builder
	lastDash := false
	for _, ch := range value {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
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

func localEmulatorDecodeBase64Payload(raw string, maxBytes int) ([]byte, error) {
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

func localEmulatorWriteFileAtomic(target string, data []byte, perm os.FileMode) error {
	tmp := target + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	return os.Rename(tmp, target)
}

func writeLocalEmulatorJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("emulador json encode error: %v", err)
	}
}

func localEmulatorUnavailable(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte(`<!doctype html><html lang="es"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Emulador no disponible</title><body style="font-family:system-ui;margin:0;min-height:100vh;display:grid;place-items:center;background:#0c1420;color:#eef6ff"><main style="max-width:520px;padding:24px"><h1>Emulador no disponible</h1><p>No se encontro la carpeta <code>juegos/public</code> en el servidor. Sincroniza el proyecto completo y vuelve a intentar.</p><a style="color:#38bdf8" href="/Juegos/menu_juegos.html">Volver a juegos</a></main></body></html>`))
}

func strconvItoa(value int) string {
	if value <= 0 {
		return "0"
	}
	var digits [20]byte
	i := len(digits)
	for value > 0 {
		i--
		digits[i] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[i:])
}
