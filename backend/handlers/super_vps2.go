package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const superVPS2ConfigKey = "super.vps2.config"

type superVPS2Config struct {
	Host              string `json:"host"`
	Port              int    `json:"port"`
	User              string `json:"user"`
	RemotePath        string `json:"remote_path"`
	NextcloudDataPath string `json:"nextcloud_data_path"`
	HostKey           string `json:"host_key"`
	IdentityFile      string `json:"identity_file,omitempty"`
	HasPassword       bool   `json:"has_password"`
}

type superVPS2Status struct {
	OK          bool                   `json:"ok"`
	CheckedAt   string                 `json:"checked_at"`
	Config      superVPS2Config        `json:"config"`
	Reachable   map[string]bool        `json:"reachable"`
	System      map[string]string      `json:"system"`
	Resources   map[string]interface{} `json:"resources"`
	Docker      map[string]interface{} `json:"docker"`
	Services    map[string]string      `json:"services"`
	Nextcloud   []map[string]string    `json:"nextcloud"`
	FileBrowser superVPS2FileBrowser   `json:"file_browser,omitempty"`
	LastAction  string                 `json:"last_action,omitempty"`
	LastMessage string                 `json:"last_message,omitempty"`
	Errors      []string               `json:"errors,omitempty"`
	Raw         map[string][]string    `json:"raw,omitempty"`
}

type superVPS2FileBrowser struct {
	RootPath    string                         `json:"root_path"`
	GeneratedAt string                         `json:"generated_at,omitempty"`
	Mode        string                         `json:"mode,omitempty"`
	Directories map[string][]superVPS2FileItem `json:"directories,omitempty"`
}

type superVPS2FileItem struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Type     string `json:"type"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
}

type superVPS2FilesResponse struct {
	OK          bool                `json:"ok"`
	Mode        string              `json:"mode"`
	RootPath    string              `json:"root_path"`
	Path        string              `json:"path"`
	Parent      string              `json:"parent"`
	Entries     []superVPS2FileItem `json:"entries"`
	LastMessage string              `json:"last_message,omitempty"`
	Errors      []string            `json:"errors,omitempty"`
}

// SuperVPS2Handler administra solo acciones cerradas sobre el VPS2. No acepta comandos libres.
func SuperVPS2Handler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper); !ok {
			return
		}
		cfg := loadSuperVPS2Config(dbSuper)
		action := strings.TrimSpace(r.URL.Query().Get("action"))
		if action == "" {
			action = "status"
		}

		switch r.Method {
		case http.MethodGet:
			if action == "files" {
				statusCode, payload := buildSuperVPS2FilesResponse(cfg, strings.TrimSpace(r.URL.Query().Get("path")))
				writeJSON(w, statusCode, payload)
				return
			}
			status := buildSuperVPS2Status(cfg)
			writeJSON(w, http.StatusOK, status)
			return
		case http.MethodPost:
			if action == "config" {
				statusCode, payload := saveSuperVPS2Config(r, dbSuper, cfg)
				writeJSON(w, statusCode, payload)
				return
			}
			resultAction, err := executeSuperVPS2Action(cfg, action)
			status := buildSuperVPS2Status(cfg)
			status.LastAction = action
			status.LastMessage = resultAction
			if err != nil {
				status.Errors = append(status.Errors, err.Error())
				writeJSON(w, http.StatusBadGateway, status)
				return
			}
			writeJSON(w, http.StatusOK, status)
			return
		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"ok": false, "error": "metodo no permitido"})
			return
		}
	}
}

func loadSuperVPS2Config(dbSuper *sql.DB) superVPS2Config {
	cfg := superVPS2Config{
		Host:              firstNonEmptyVPS2(os.Getenv("PCS_VPS2_HOST"), "192.168.1.188"),
		Port:              parseIntDefault(os.Getenv("PCS_VPS2_PORT"), 22),
		User:              firstNonEmptyVPS2(os.Getenv("PCS_VPS2_USER"), "admin1"),
		RemotePath:        firstNonEmptyVPS2(os.Getenv("PCS_VPS2_REMOTE_PATH"), "/home/admin1/powerfulcontrolsystem"),
		NextcloudDataPath: firstNonEmptyVPS2(os.Getenv("PCS_VPS2_NEXTCLOUD_DATA_PATH"), "/srv/data/nextcloud/data"),
		HostKey:           os.Getenv("PCS_VPS2_HOSTKEY"),
	}
	cfg.IdentityFile = os.Getenv("PCS_VPS2_IDENTITY_FILE")
	password := os.Getenv("PCS_VPS2_PASSWORD")

	localCfg := filepath.Join(projectRootFromHandlers(), "scripts", "pcs_deployment.local.ps1")
	if values := parseSimplePowerShellConfig(localCfg); len(values) > 0 {
		cfg.Host = firstNonEmptyVPS2(values["PcsVps2Host"], cfg.Host)
		cfg.Port = parseIntDefault(values["PcsVps2Port"], cfg.Port)
		cfg.User = firstNonEmptyVPS2(values["PcsVps2User"], cfg.User)
		cfg.RemotePath = firstNonEmptyVPS2(values["PcsVps2RemotePath"], cfg.RemotePath)
		cfg.NextcloudDataPath = firstNonEmptyVPS2(values["PcsVps2NextcloudDataPath"], cfg.NextcloudDataPath)
		cfg.HostKey = firstNonEmptyVPS2(values["PcsVps2HostKey"], cfg.HostKey)
		cfg.IdentityFile = firstNonEmptyVPS2(values["PcsVps2IdentityFile"], cfg.IdentityFile)
		password = firstNonEmptyVPS2(values["PcsVps2Password"], password)
	}
	if cfg.HostKey == "" {
		cfg.HostKey = "SHA256:QQmT0ZjCVNNxw7ICwV7FKwrzzzfWrOrtZ9zTrEGkwH0"
	}
	if dbSuper != nil {
		if raw, _, err := dbpkg.GetConfigValue(dbSuper, superVPS2ConfigKey); err == nil && strings.TrimSpace(raw) != "" {
			var saved superVPS2Config
			if json.Unmarshal([]byte(raw), &saved) == nil {
				cfg.Host = firstNonEmptyVPS2(saved.Host, cfg.Host)
				if saved.Port > 0 {
					cfg.Port = saved.Port
				}
				cfg.User = firstNonEmptyVPS2(saved.User, cfg.User)
				cfg.RemotePath = firstNonEmptyVPS2(saved.RemotePath, cfg.RemotePath)
				cfg.NextcloudDataPath = firstNonEmptyVPS2(saved.NextcloudDataPath, cfg.NextcloudDataPath)
				cfg.HostKey = firstNonEmptyVPS2(saved.HostKey, cfg.HostKey)
			}
		}
	}
	cfg.HasPassword = password != ""
	return cfg
}

func saveSuperVPS2Config(r *http.Request, dbSuper *sql.DB, current superVPS2Config) (int, map[string]interface{}) {
	if dbSuper == nil {
		return http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "base de datos super no disponible"}
	}
	var payload superVPS2Config
	if err := json.NewDecoder(io.LimitReader(r.Body, 64*1024)).Decode(&payload); err != nil {
		return http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "payload de VPS2 invalido"}
	}
	cfg := current
	cfg.Host = firstNonEmptyVPS2(payload.Host, current.Host)
	if payload.Port > 0 {
		cfg.Port = payload.Port
	}
	cfg.User = firstNonEmptyVPS2(payload.User, current.User)
	cfg.RemotePath = firstNonEmptyVPS2(payload.RemotePath, current.RemotePath)
	cfg.NextcloudDataPath = firstNonEmptyVPS2(payload.NextcloudDataPath, current.NextcloudDataPath)
	cfg.HostKey = firstNonEmptyVPS2(payload.HostKey, current.HostKey)
	cfg.IdentityFile = ""
	cfg.HasPassword = current.HasPassword
	if cfg.Host == "" || strings.ContainsAny(cfg.Host, " \t\r\n") {
		return http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "IP o host VPS2 invalido"}
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "puerto VPS2 invalido"}
	}
	if cfg.User == "" || strings.ContainsAny(cfg.User, " \t\r\n") {
		return http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "usuario VPS2 invalido"}
	}
	if !strings.HasPrefix(cfg.RemotePath, "/") || !strings.HasPrefix(cfg.NextcloudDataPath, "/") {
		return http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "las rutas remotas deben ser absolutas"}
	}
	stored := cfg
	stored.IdentityFile = ""
	stored.HasPassword = false
	raw, err := json.Marshal(stored)
	if err != nil {
		return http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "no se pudo preparar configuracion VPS2"}
	}
	if err := dbpkg.SetConfigValue(dbSuper, superVPS2ConfigKey, string(raw), false); err != nil {
		return http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "no se pudo guardar configuracion VPS2"}
	}
	_ = dbpkg.SetConfigValue(dbSuper, superVPS2ConfigKey+".updated_by", adminEmailFromRequest(r), false)
	return http.StatusOK, map[string]interface{}{"ok": true, "config": cfg, "last_message": "Configuracion VPS2 guardada."}
}

func buildSuperVPS2Status(cfg superVPS2Config) superVPS2Status {
	status := superVPS2Status{
		OK:        true,
		CheckedAt: time.Now().Format(time.RFC3339),
		Config:    cfg,
		Reachable: map[string]bool{
			"ssh": tcpReachable(cfg.Host, cfg.Port, 2*time.Second),
			"vnc": tcpReachable(cfg.Host, 5901, 2*time.Second),
		},
		System:    map[string]string{},
		Resources: map[string]interface{}{},
		Docker:    map[string]interface{}{},
		Services:  map[string]string{},
		Raw:       map[string][]string{},
	}
	if !status.Reachable["ssh"] {
		if snapshot, ok := loadSuperVPS2Snapshot(); ok {
			snapshot.LastMessage = "Datos cargados desde el ultimo snapshot publicado por sync_to_vps2."
			snapshot.Reachable = status.Reachable
			snapshot.Config = cfg
			return snapshot
		}
		status.OK = false
		status.Errors = append(status.Errors, "SSH no responde en VPS2 desde este servidor. Si VPS2 esta en una red privada, ejecuta sync_to_vps2 para publicar un snapshot.")
		return status
	}

	out, err := runSuperVPS2SSH(cfg, superVPS2StatusCommand(cfg), 12*time.Second)
	if err != nil {
		status.OK = false
		status.Errors = append(status.Errors, err.Error())
		return status
	}

	lines := parseKeyValueLines(out)
	status.Raw = lines
	status.System["hostname"] = firstLine(lines, "hostname")
	status.System["uptime"] = firstLine(lines, "uptime")
	status.System["kernel"] = firstLine(lines, "kernel")
	status.System["default_target"] = firstLine(lines, "default_target")
	status.System["ip"] = strings.Join(lines["ip"], ", ")
	status.Resources["load"] = firstLine(lines, "load")
	status.Resources["cpu_cores"] = firstLine(lines, "cpu_cores")
	status.Resources["temperature_c"] = firstLine(lines, "temperature_c")
	status.Resources["memory"] = parseTripleBytes(firstLine(lines, "memory"))
	status.Resources["disk_root"] = parseDiskLine(firstLine(lines, "disk_root"))
	status.Resources["disk_nextcloud"] = parseDiskLine(firstLine(lines, "disk_nextcloud"))
	status.Services["docker"] = firstLine(lines, "service_docker")
	status.Services["ssh"] = firstLine(lines, "service_ssh")
	status.Docker["version"] = firstLine(lines, "docker_version")
	status.Docker["containers_total"] = firstLine(lines, "docker_total")
	status.Docker["containers_running"] = firstLine(lines, "docker_running")
	status.Nextcloud = parsePipeRows(lines["nextcloud"])
	return status
}

func loadSuperVPS2Snapshot() (superVPS2Status, bool) {
	candidates := []string{
		os.Getenv("PCS_VPS2_STATUS_FILE"),
		filepath.Join(projectRootFromHandlers(), "backup", "vps2_status.json"),
		"/app/backup/vps2_status.json",
	}
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		raw, err := os.ReadFile(candidate)
		if err != nil || len(raw) == 0 {
			continue
		}
		raw = bytes.TrimPrefix(raw, []byte{0xEF, 0xBB, 0xBF})
		var status superVPS2Status
		if err := json.Unmarshal(raw, &status); err != nil {
			continue
		}
		if status.CheckedAt == "" {
			status.CheckedAt = time.Now().Format(time.RFC3339)
		}
		return status, true
	}
	return superVPS2Status{}, false
}

func executeSuperVPS2Action(cfg superVPS2Config, action string) (string, error) {
	var command string
	switch action {
	case "restart_nextcloud":
		command = "docker ps -a --format '{{.Names}}' | grep -Ei 'nextcloud|cloud' | xargs -r docker update --restart unless-stopped >/dev/null && docker ps -a --format '{{.Names}}' | grep -Ei 'nextcloud|cloud' | xargs -r docker restart"
	case "reboot":
		command = superVPS2SudoPrefix(cfg) + " systemctl reboot"
	case "shutdown":
		command = superVPS2SudoPrefix(cfg) + " systemctl poweroff"
	case "status":
		return "Estado actualizado.", nil
	default:
		return "", fmt.Errorf("accion VPS2 no permitida")
	}
	if action == "reboot" || action == "shutdown" {
		_, err := runSuperVPS2SSH(cfg, command, 3*time.Second)
		if err != nil && !strings.Contains(strings.ToLower(err.Error()), "connection") {
			return "", err
		}
		return "Accion enviada al VPS2.", nil
	}
	out, err := runSuperVPS2SSH(cfg, command, 20*time.Second)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func buildSuperVPS2FilesResponse(cfg superVPS2Config, rawPath string) (int, superVPS2FilesResponse) {
	relPath := cleanVPS2FilePath(rawPath)
	response := superVPS2FilesResponse{
		OK:       true,
		Mode:     "snapshot",
		RootPath: cfg.NextcloudDataPath,
		Path:     relPath,
		Parent:   parentVPS2FilePath(relPath),
		Entries:  []superVPS2FileItem{},
	}
	if tcpReachable(cfg.Host, cfg.Port, 2*time.Second) {
		entries, err := listSuperVPS2FilesLive(cfg, relPath)
		if err != nil {
			response.OK = false
			response.Errors = append(response.Errors, err.Error())
			return http.StatusBadGateway, response
		}
		response.Mode = "live"
		response.Entries = entries
		response.LastMessage = "Archivos listados en vivo por SSH."
		return http.StatusOK, response
	}
	snapshot, ok := loadSuperVPS2Snapshot()
	if !ok || snapshot.FileBrowser.Directories == nil {
		response.OK = false
		response.Errors = append(response.Errors, "No hay indice de archivos VPS2 publicado. Ejecuta sync_to_vps2 para actualizarlo.")
		return http.StatusBadGateway, response
	}
	response.RootPath = firstNonEmptyVPS2(snapshot.FileBrowser.RootPath, cfg.NextcloudDataPath)
	if entries, ok := snapshot.FileBrowser.Directories[relPath]; ok {
		response.Entries = entries
	} else if relPath == "" {
		response.Entries = snapshot.FileBrowser.Directories[""]
	} else {
		response.Errors = append(response.Errors, "Carpeta no incluida en el snapshot actual.")
	}
	response.LastMessage = "Archivos cargados desde el snapshot publicado por sync_to_vps2."
	return http.StatusOK, response
}

func listSuperVPS2FilesLive(cfg superVPS2Config, relPath string) ([]superVPS2FileItem, error) {
	base := strings.TrimRight(cfg.NextcloudDataPath, "/")
	target := base
	if relPath != "" {
		target += "/" + relPath
	}
	command := "base=" + shellQuote(target) + `; [ -d "$base" ] || exit 8; find "$base" -mindepth 1 -maxdepth 1 -printf 'file=%f\t%y\t%s\t%TY-%Tm-%Td %TH:%TM\t%p\n' 2>/dev/null | sort | head -n 500`
	out, err := runSuperVPS2SSH(cfg, command, 15*time.Second)
	if err != nil {
		return nil, fmt.Errorf("no se pudo listar la carpeta de Nextcloud en VPS2")
	}
	return parseVPS2FileRows(out, strings.TrimRight(cfg.NextcloudDataPath, "/")), nil
}

func parseVPS2FileRows(out, basePath string) []superVPS2FileItem {
	basePath = strings.TrimRight(basePath, "/")
	items := []superVPS2FileItem{}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "file=") {
			continue
		}
		parts := strings.SplitN(strings.TrimPrefix(line, "file="), "\t", 5)
		if len(parts) < 5 {
			continue
		}
		fullPath := strings.TrimSpace(parts[4])
		relPath := strings.TrimPrefix(strings.TrimPrefix(fullPath, basePath), "/")
		itemType := "file"
		if parts[1] == "d" {
			itemType = "dir"
		}
		items = append(items, superVPS2FileItem{
			Name:     strings.TrimSpace(parts[0]),
			Path:     cleanVPS2FilePath(relPath),
			Type:     itemType,
			Size:     parseInt64Default(parts[2], 0),
			Modified: strings.TrimSpace(parts[3]),
		})
	}
	return items
}

func cleanVPS2FilePath(raw string) string {
	raw = strings.ReplaceAll(strings.TrimSpace(raw), "\\", "/")
	raw = strings.TrimPrefix(raw, "/")
	clean := filepath.ToSlash(filepath.Clean("/" + raw))
	clean = strings.TrimPrefix(clean, "/")
	if clean == "." || strings.HasPrefix(clean, "../") || clean == ".." {
		return ""
	}
	return clean
}

func parentVPS2FilePath(relPath string) string {
	relPath = cleanVPS2FilePath(relPath)
	if relPath == "" {
		return ""
	}
	parent := filepath.ToSlash(filepath.Dir(relPath))
	if parent == "." {
		return ""
	}
	return cleanVPS2FilePath(parent)
}

func superVPS2StatusCommand(cfg superVPS2Config) string {
	nextcloudPath := shellQuote(firstNonEmptyVPS2(cfg.NextcloudDataPath, "/srv/data/nextcloud/data"))
	return `printf 'hostname=%s\n' "$(hostname)";
printf 'uptime=%s\n' "$(uptime -p 2>/dev/null || uptime)";
printf 'kernel=%s\n' "$(uname -srmo)";
printf 'default_target=%s\n' "$(systemctl get-default 2>/dev/null || true)";
printf 'load=%s\n' "$(cat /proc/loadavg 2>/dev/null | awk '{print $1" "$2" "$3}')";
printf 'cpu_cores=%s\n' "$(nproc 2>/dev/null || echo 0)";
temp="$(for f in /sys/class/thermal/thermal_zone*/temp; do [ -r "$f" ] && awk '{printf "%.1f", $1/1000}' "$f" && break; done 2>/dev/null)";
printf 'temperature_c=%s\n' "$temp";
printf 'memory=%s\n' "$(free -b 2>/dev/null | awk '/Mem:/ {print $2" "$3" "$7}')";
printf 'disk_root=%s\n' "$(df -PB1 / 2>/dev/null | awk 'NR==2 {print $2" "$3" "$4" "$5}')";
nextcloud_data_path=` + nextcloudPath + `;
printf 'disk_nextcloud=%s\n' "$(df -PB1 "$nextcloud_data_path" 2>/dev/null | awk 'NR==2 {print $2" "$3" "$4" "$5}')";
printf 'service_docker=%s\n' "$(systemctl is-active docker 2>/dev/null || true)";
printf 'service_ssh=%s\n' "$(systemctl is-active ssh 2>/dev/null || systemctl is-active sshd 2>/dev/null || true)";
printf 'docker_version=%s\n' "$(docker --version 2>/dev/null || true)";
printf 'docker_total=%s\n' "$(docker ps -a -q 2>/dev/null | wc -l | tr -d " ")";
printf 'docker_running=%s\n' "$(docker ps -q 2>/dev/null | wc -l | tr -d " ")";
ip -4 -o addr show scope global 2>/dev/null | awk '{print "ip="$2" "$4}';
docker ps -a --format '{{.Names}}|{{.Status}}|{{.Ports}}' 2>/dev/null | grep -Ei '(^|[-_])(nextcloud|cloud)([-_]|$)|nextcloud' | sed 's/^/nextcloud=/' || true`
}

func runSuperVPS2SSH(cfg superVPS2Config, remoteCommand string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if runtime.GOOS == "windows" {
		plink := resolveWindowsPlink()
		if plink == "" {
			return "", fmt.Errorf("plink.exe no disponible en el servidor")
		}
		args := []string{"-batch", "-ssh", "-P", strconv.Itoa(cfg.Port)}
		if cfg.HostKey != "" {
			args = append(args, "-hostkey", cfg.HostKey)
		}
		if cfg.IdentityFile != "" {
			args = append(args, "-i", cfg.IdentityFile)
		} else if password := loadSuperVPS2Password(); password != "" {
			args = append(args, "-pw", password)
		} else {
			return "", fmt.Errorf("VPS2 sin llave SSH ni password configurado en entorno privado")
		}
		args = append(args, cfg.User+"@"+cfg.Host, remoteCommand)
		out, err := exec.CommandContext(ctx, plink, args...).CombinedOutput() // #nosec G204 -- executable is resolved locally; arguments are separate and remoteCommand is built only by closed server-side actions with shell-quoted paths.
		return string(out), err
	}

	args := []string{"-p", strconv.Itoa(cfg.Port), "-o", "BatchMode=yes", "-o", "ConnectTimeout=5"}
	if cfg.IdentityFile != "" {
		args = append(args, "-i", cfg.IdentityFile)
	}
	args = append(args, cfg.User+"@"+cfg.Host, remoteCommand)
	out, err := exec.CommandContext(ctx, "ssh", args...).CombinedOutput() // #nosec G204 -- executable and options are fixed; the authenticated target and closed remote action are passed as separate arguments.
	return string(out), err
}

func superVPS2SudoPrefix(cfg superVPS2Config) string {
	if password := loadSuperVPS2Password(); password != "" {
		return "printf " + shellQuote(password+"\\n") + " | sudo -S -p ''"
	}
	_ = cfg
	return "sudo -n"
}

func loadSuperVPS2Password() string {
	if password := os.Getenv("PCS_VPS2_PASSWORD"); password != "" {
		return password
	}
	localCfg := filepath.Join(projectRootFromHandlers(), "scripts", "pcs_deployment.local.ps1")
	values := parseSimplePowerShellConfig(localCfg)
	return values["PcsVps2Password"]
}

func parseSimplePowerShellConfig(path string) map[string]string {
	values := map[string]string{}
	raw, err := os.ReadFile(path)
	if err != nil {
		return values
	}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || !strings.Contains(line, "$script:") || !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(strings.TrimPrefix(parts[0], "$script:"))
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)
		if key != "" && !strings.Contains(value, "$script:") {
			values[key] = value
		}
	}
	return values
}

func resolveWindowsPlink() string {
	candidates := []string{
		os.Getenv("PCS_PLINK_PATH"),
		`D:\Program Files\PuTTY\plink.exe`,
		`C:\Program Files\PuTTY\plink.exe`,
		`C:\Program Files (x86)\PuTTY\plink.exe`,
	}
	for _, candidate := range candidates {
		if candidate != "" {
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
		}
	}
	return ""
}

func projectRootFromHandlers() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	if strings.HasSuffix(filepath.ToSlash(wd), "/backend") {
		return filepath.Dir(wd)
	}
	return wd
}

func tcpReachable(host string, port int, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(port)), timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func parseKeyValueLines(out string) map[string][]string {
	values := map[string][]string{}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		values[key] = append(values[key], value)
	}
	return values
}

func firstLine(values map[string][]string, key string) string {
	if list := values[key]; len(list) > 0 {
		return list[0]
	}
	return ""
}

func parseTripleBytes(line string) map[string]int64 {
	parts := strings.Fields(line)
	return map[string]int64{
		"total":     parseInt64At(parts, 0),
		"used":      parseInt64At(parts, 1),
		"available": parseInt64At(parts, 2),
	}
}

func parseDiskLine(line string) map[string]interface{} {
	parts := strings.Fields(line)
	return map[string]interface{}{
		"total":     parseInt64At(parts, 0),
		"used":      parseInt64At(parts, 1),
		"available": parseInt64At(parts, 2),
		"percent":   stringAt(parts, 3),
	}
}

func parsePipeRows(rows []string) []map[string]string {
	out := make([]map[string]string, 0, len(rows))
	for _, row := range rows {
		parts := strings.SplitN(row, "|", 3)
		item := map[string]string{"name": stringAt(parts, 0), "status": stringAt(parts, 1), "ports": stringAt(parts, 2)}
		out = append(out, item)
	}
	return out
}

func parseIntDefault(raw string, fallback int) int {
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func parseInt64At(parts []string, idx int) int64 {
	if idx < 0 || idx >= len(parts) {
		return 0
	}
	n, _ := strconv.ParseInt(strings.TrimSpace(parts[idx]), 10, 64)
	return n
}

func parseInt64Default(raw string, fallback int64) int64 {
	n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return fallback
	}
	return n
}

func stringAt(parts []string, idx int) string {
	if idx < 0 || idx >= len(parts) {
		return ""
	}
	return strings.TrimSpace(parts[idx])
}

func firstNonEmptyVPS2(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
