package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
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

type ServicioEstado struct {
	ID          string            `json:"id"`
	Nombre      string            `json:"nombre"`
	Estado      string            `json:"estado"`
	Detalle     string            `json:"detalle"`
	Habilitado  bool              `json:"habilitado,omitempty"`
	Componentes map[string]string `json:"componentes,omitempty"`
	Prueba      *ServicioPrueba   `json:"prueba,omitempty"`
}

type ServicioPrueba struct {
	OK        bool              `json:"ok"`
	Resumen   string            `json:"resumen"`
	Revisado  string            `json:"revisado"`
	Puertos   map[string]string `json:"puertos,omitempty"`
	Servicios map[string]string `json:"servicios,omitempty"`
}

type rustDeskRemoteConfig struct {
	Host     string
	User     string
	KeyPath  string
	ExecPath string
	UsePlink bool
}

var rustDeskHBBSCandidates = []string{"rustdesk-hbbs", "hbbs"}
var rustDeskHBBRCandidates = []string{"rustdesk-hbbr", "hbbr"}

type rustDeskPanelConfig struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
	User    string `json:"user"`
	KeyPath string `json:"key_path"`
}

func loadRustDeskPanelConfig(dbSuper *sql.DB) rustDeskPanelConfig {
	cfg := rustDeskPanelConfig{}
	if dbSuper == nil {
		return cfg
	}
	enabled, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, "rustdesk.vps_ssh_enabled")
	switch strings.ToLower(strings.TrimSpace(enabled)) {
	case "1", "true", "on", "activo", "enabled":
		cfg.Enabled = true
	}
	host, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, "rustdesk.vps_ssh_host")
	user, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, "rustdesk.vps_ssh_user")
	keyPath, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, "rustdesk.vps_ssh_key_path")
	cfg.Host = strings.TrimSpace(host)
	cfg.User = strings.TrimSpace(user)
	cfg.KeyPath = strings.TrimSpace(keyPath)
	return cfg
}

func buildRustDeskServiceState(dbSuper *sql.DB, includeProbe bool) ServicioEstado {
	panelCfg := loadRustDeskPanelConfig(dbSuper)
	hbbsService := resolveRustDeskServiceName(dbSuper, rustDeskHBBSCandidates)
	hbbrService := resolveRustDeskServiceName(dbSuper, rustDeskHBBRCandidates)
	hbbsStatus := checkSystemctlStatus(dbSuper, hbbsService)
	hbbrStatus := checkSystemctlStatus(dbSuper, hbbrService)
	overall := "inactive"
	switch {
	case hbbsStatus == "active" && hbbrStatus == "active":
		overall = "active"
	case hbbsStatus == "error" || hbbrStatus == "error":
		overall = "error"
	case hbbsStatus == "active" || hbbrStatus == "active":
		overall = "degraded"
	}

	if runtime.GOOS == "windows" && !shouldUseRustDeskRemoteExec(dbSuper) {
		overall = "unavailable"
	}

	detalle := "Servidor ID/Relay para soporte remoto de clientes a traves de VPS."
	if overall == "unavailable" {
		detalle = "Este backend corre en Windows. Para gestionar RustDesk en el VPS, activa 'Control por SSH' y configura host/usuario/llave en esta misma pantalla."
	}

	state := ServicioEstado{
		ID:         "rustdesk",
		Nombre:     "RustDesk (Soporte Remoto)",
		Estado:     overall,
		Detalle:    detalle,
		Habilitado: panelCfg.Enabled,
		Componentes: map[string]string{
			"rustdesk-hbbs": hbbsStatus,
			"rustdesk-hbbr": hbbrStatus,
		},
	}
	if includeProbe {
		probe := probeRustDeskService(dbSuper)
		state.Prueba = &probe
	}
	return state
}

func SuperServidoresListHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		servicios := []ServicioEstado{
			buildRustDeskServiceState(dbSuper, false),
			buildOnlyOfficeServiceState(dbSuper),
			buildPCSBackendServiceState(dbSuper),
			buildNginxServiceState(dbSuper),
			buildPostgresServiceState(dbSuper),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "servicios": servicios})
	}
}

func readLinuxProcessRSSKBByGrep(expr string) (int64, error) {
	if runtime.GOOS == "windows" {
		return 0, fmt.Errorf("unavailable on windows")
	}
	// sum RSS KB of matching processes
	cmd := exec.Command("bash", "-lc", "ps -eo rss,args | grep -E "+shellEscapeForPOSIX(expr)+" | grep -v grep | awk '{s+=$1} END{print s+0}'")
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	v := strings.TrimSpace(string(out))
	if v == "" {
		return 0, nil
	}
	n, _ := strconv.ParseInt(v, 10, 64)
	return n, nil
}

func buildOnlyOfficeServiceState(dbSuper *sql.DB) ServicioEstado {
	enabled := isOnlyOfficeEnabled(dbSuper)
	state := ServicioEstado{
		ID:         "onlyoffice",
		Nombre:     "OnlyOffice Document Server",
		Estado:     "inactive",
		Detalle:    "Editor de documentos (Docker) para el módulo Documentos.",
		Habilitado: enabled,
		Componentes: map[string]string{
			"docker": "unknown",
		},
	}
	if runtime.GOOS == "windows" {
		state.Estado = "unavailable"
		state.Detalle = "Backend en Windows: no se puede inspeccionar el contenedor OnlyOffice localmente."
		return state
	}
	// Check docker container running
	out, _ := exec.Command("bash", "-lc", "docker ps --format '{{.Names}}' 2>/dev/null | grep -x 'pcs-onlyoffice-documentserver' || true").Output()
	if strings.TrimSpace(string(out)) != "" {
		state.Estado = "active"
		state.Componentes["docker"] = "active"
		rss, _ := readLinuxProcessRSSKBByGrep("onlyoffice|documentserver|ds\\/run-document-server")
		if rss > 0 {
			state.Componentes["mem_rss_kb"] = fmt.Sprintf("%d", rss)
		}
		return state
	}
	state.Componentes["docker"] = "inactive"
	return state
}

func buildPCSBackendServiceState(dbSuper *sql.DB) ServicioEstado {
	state := ServicioEstado{
		ID:         "pcs_backend",
		Nombre:     "Backend PCS (systemd)",
		Estado:     "unknown",
		Detalle:    "Servicio principal del backend.",
		Habilitado: true,
		Componentes: map[string]string{
			"powerfulcontrolsystem.service": checkSystemctlStatus(dbSuper, "powerfulcontrolsystem.service"),
		},
	}
	if state.Componentes["powerfulcontrolsystem.service"] == "active" {
		state.Estado = "active"
	} else {
		state.Estado = state.Componentes["powerfulcontrolsystem.service"]
	}
	if runtime.GOOS != "windows" {
		rss, _ := readLinuxProcessRSSKBByGrep("server_linux_amd64|pos-backend|powerfulcontrolsystem")
		if rss > 0 {
			state.Componentes["mem_rss_kb"] = fmt.Sprintf("%d", rss)
		}
	}
	return state
}

func buildNginxServiceState(dbSuper *sql.DB) ServicioEstado {
	state := ServicioEstado{
		ID:         "nginx",
		Nombre:     "Nginx (reverse proxy)",
		Estado:     checkSystemctlStatus(dbSuper, "nginx"),
		Detalle:    "Proxy HTTPS y rutas públicas.",
		Habilitado: true,
		Componentes: map[string]string{
			"nginx": checkSystemctlStatus(dbSuper, "nginx"),
		},
	}
	if state.Componentes["nginx"] == "active" {
		state.Estado = "active"
	}
	if runtime.GOOS != "windows" {
		rss, _ := readLinuxProcessRSSKBByGrep("nginx: master|nginx: worker")
		if rss > 0 {
			state.Componentes["mem_rss_kb"] = fmt.Sprintf("%d", rss)
		}
	}
	return state
}

func buildPostgresServiceState(dbSuper *sql.DB) ServicioEstado {
	state := ServicioEstado{
		ID:          "postgres",
		Nombre:      "PostgreSQL",
		Estado:      "unknown",
		Detalle:     "Base de datos del sistema (VPS).",
		Habilitado:  true,
		Componentes: map[string]string{},
	}
	if runtime.GOOS == "windows" {
		state.Estado = "unavailable"
		return state
	}
	// postgresql service name varies; try generic.
	pgStatus := checkSystemctlStatus(dbSuper, "postgresql")
	state.Componentes["postgresql"] = pgStatus
	state.Estado = pgStatus
	rss, _ := readLinuxProcessRSSKBByGrep("postgres:|postgresql")
	if rss > 0 {
		state.Componentes["mem_rss_kb"] = fmt.Sprintf("%d", rss)
	}
	return state
}

func SuperServidoresToggleHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			ID     string `json:"id"`
			Accion string `json:"accion"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		if payload.ID == "rustdesk" {
			action := strings.ToLower(strings.TrimSpace(payload.Accion))
			if action == "enable" || action == "disable" {
				if err := setRustDeskPanelEnabled(dbSuper, action == "enable"); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			if action == "start" || action == "stop" || action == "restart" {
				if runtime.GOOS == "windows" && !shouldUseRustDeskRemoteExec(dbSuper) {
					http.Error(w, "Control local no disponible en Windows. Activa 'Control por SSH' y configura la conexión al VPS.", http.StatusBadRequest)
					return
				}
				err1 := runSystemctl(dbSuper, action, resolveRustDeskServiceName(dbSuper, rustDeskHBBSCandidates))
				err2 := runSystemctl(dbSuper, action, resolveRustDeskServiceName(dbSuper, rustDeskHBBRCandidates))
				if err1 != nil {
					http.Error(w, err1.Error(), http.StatusInternalServerError)
					return
				}
				if err2 != nil {
					http.Error(w, err2.Error(), http.StatusInternalServerError)
					return
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "servicio": buildRustDeskServiceState(dbSuper, false)})
	}
}

func SuperServidoresProbeHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		service := strings.TrimSpace(r.URL.Query().Get("id"))
		if service == "" {
			service = "rustdesk"
		}
		if service != "rustdesk" {
			http.Error(w, "servicio no soportado", http.StatusBadRequest)
			return
		}
		state := buildRustDeskServiceState(dbSuper, true)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "servicio": state})
	}
}

func checkSystemctlStatus(dbSuper *sql.DB, service string) string {
	if strings.TrimSpace(service) == "" {
		return "missing"
	}
	if shouldUseRustDeskRemoteExec(dbSuper) {
		out, err := runRustDeskRemoteShell(dbSuper, fmt.Sprintf("systemctl is-active %s 2>/dev/null || true", shellEscapeForPOSIX(service)))
		if err != nil && strings.TrimSpace(out) == "" {
			return "error"
		}
		res := strings.TrimSpace(firstNonEmptyLine(out))
		switch res {
		case "active":
			return "active"
		case "inactive", "failed", "activating", "deactivating":
			return res
		case "":
			return "error"
		default:
			return "error"
		}
	}
	if runtime.GOOS == "windows" {
		return "inactive"
	}
	cmd := exec.Command("systemctl", "is-active", service)
	out, err := cmd.Output()
	if err != nil {
		return "error"
	}
	res := strings.TrimSpace(string(out))
	if res == "active" {
		return "active"
	}
	return "inactive"
}

func runSystemctl(dbSuper *sql.DB, accion string, service string) error {
	if strings.TrimSpace(service) == "" {
		return fmt.Errorf("no se encontro ninguna unidad systemd compatible con RustDesk en el VPS")
	}
	if shouldUseRustDeskRemoteExec(dbSuper) {
		_, err := runRustDeskRemoteShell(dbSuper, fmt.Sprintf("sudo -n systemctl %s %s", shellEscapeForPOSIX(accion), shellEscapeForPOSIX(service)))
		return err
	}
	if runtime.GOOS == "windows" {
		return fmt.Errorf("control local de RustDesk no disponible en Windows sin configuracion SSH al VPS")
	}
	cmd := exec.Command("sudo", "systemctl", accion, service)
	err := cmd.Run()
	return err
}

func probeRustDeskService(dbSuper *sql.DB) ServicioPrueba {
	probe := ServicioPrueba{
		OK:       false,
		Resumen:  "Comprobacion no disponible en este entorno.",
		Revisado: time.Now().In(time.Local).Format("2006-01-02 15:04:05"),
		Puertos:  map[string]string{},
		Servicios: map[string]string{
			"rustdesk-hbbs": checkSystemctlStatus(dbSuper, resolveRustDeskServiceName(dbSuper, rustDeskHBBSCandidates)),
			"rustdesk-hbbr": checkSystemctlStatus(dbSuper, resolveRustDeskServiceName(dbSuper, rustDeskHBBRCandidates)),
		},
	}
	if runtime.GOOS == "windows" {
		if !shouldUseRustDeskRemoteExec(dbSuper) {
			probe.Resumen = "Entorno local Windows: configura DB_VPS_SSH_HOST/USER/KEY_PATH para ejecutar la prueba real en el VPS Linux."
			return probe
		}
	}

	ports := []int{21114, 21115, 21116, 21117, 21118, 21119}
	openPorts := 0
	if shouldUseRustDeskRemoteExec(dbSuper) {
		ssOutput, err := runRustDeskRemoteShell(dbSuper, "ss -ltn 2>/dev/null || true")
		if err != nil && strings.TrimSpace(ssOutput) == "" {
			probe.Resumen = "No se pudo ejecutar la prueba remota de RustDesk en el VPS."
			return probe
		}
		for _, port := range ports {
			portText := fmt.Sprintf(":%d", port)
			if strings.Contains(ssOutput, portText) {
				probe.Puertos[fmt.Sprintf("%d", port)] = "abierto"
				openPorts++
				continue
			}
			probe.Puertos[fmt.Sprintf("%d", port)] = "cerrado"
		}
	} else {
		for _, port := range ports {
			address := fmt.Sprintf("127.0.0.1:%d", port)
			conn, err := net.DialTimeout("tcp", address, 700*time.Millisecond)
			if err != nil {
				probe.Puertos[fmt.Sprintf("%d", port)] = "cerrado"
				continue
			}
			_ = conn.Close()
			probe.Puertos[fmt.Sprintf("%d", port)] = "abierto"
			openPorts++
		}
	}
	if probe.Servicios["rustdesk-hbbs"] == "active" && probe.Servicios["rustdesk-hbbr"] == "active" && openPorts > 0 {
		probe.OK = true
		if shouldUseRustDeskRemoteExec(dbSuper) {
			probe.Resumen = fmt.Sprintf("RustDesk responde en el VPS: hbbs/hbbr activos y %d puerto(s) abiertos.", openPorts)
		} else {
			probe.Resumen = fmt.Sprintf("RustDesk responde: hbbs/hbbr activos y %d puerto(s) locales abiertos.", openPorts)
		}
		return probe
	}
	if shouldUseRustDeskRemoteExec(dbSuper) {
		probe.Resumen = fmt.Sprintf("RustDesk en el VPS con alertas: hbbs=%s, hbbr=%s, puertos abiertos=%d.", probe.Servicios["rustdesk-hbbs"], probe.Servicios["rustdesk-hbbr"], openPorts)
	} else {
		probe.Resumen = fmt.Sprintf("RustDesk con alertas: hbbs=%s, hbbr=%s, puertos abiertos=%d.", probe.Servicios["rustdesk-hbbs"], probe.Servicios["rustdesk-hbbr"], openPorts)
	}
	return probe
}

func resolveRustDeskServiceName(dbSuper *sql.DB, candidates []string) string {
	for _, candidate := range candidates {
		if rustDeskServiceExists(dbSuper, candidate) {
			return candidate
		}
	}
	return ""
}

func rustDeskServiceExists(dbSuper *sql.DB, service string) bool {
	if strings.TrimSpace(service) == "" {
		return false
	}
	command := fmt.Sprintf("systemctl cat %s >/dev/null 2>&1", shellEscapeForPOSIX(service))
	if shouldUseRustDeskRemoteExec(dbSuper) {
		_, err := runRustDeskRemoteShell(dbSuper, command)
		return err == nil
	}
	if runtime.GOOS == "windows" {
		return false
	}
	cmd := exec.Command("sh", "-lc", command)
	return cmd.Run() == nil
}

func shouldUseRustDeskRemoteExec(dbSuper *sql.DB) bool {
	panelCfg := loadRustDeskPanelConfig(dbSuper)
	if panelCfg.Enabled && panelCfg.Host != "" && panelCfg.User != "" {
		return true
	}
	return strings.TrimSpace(os.Getenv("DB_VPS_SSH_HOST")) != "" && strings.TrimSpace(os.Getenv("DB_VPS_SSH_USER")) != ""
}

func runRustDeskRemoteShell(dbSuper *sql.DB, script string) (string, error) {
	cfg, err := resolveRustDeskRemoteConfig(dbSuper)
	if err != nil {
		return "", err
	}
	target := fmt.Sprintf("%s@%s", cfg.User, cfg.Host)
	var cmd *exec.Cmd
	if cfg.UsePlink {
		cmd = exec.Command(cfg.ExecPath, "-batch", "-i", cfg.KeyPath, target, "sh", "-lc", script)
	} else {
		cmd = exec.Command(cfg.ExecPath, "-o", "BatchMode=yes", "-i", cfg.KeyPath, target, "sh", "-lc", script)
	}
	out, runErr := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if runErr != nil {
		if output != "" {
			return output, fmt.Errorf("ssh/plink rustdesk command failed: %w: %s", runErr, output)
		}
		return "", fmt.Errorf("ssh/plink rustdesk command failed: %w", runErr)
	}
	return output, nil
}

func resolveRustDeskRemoteConfig(dbSuper *sql.DB) (rustDeskRemoteConfig, error) {
	panelCfg := loadRustDeskPanelConfig(dbSuper)
	host := strings.TrimSpace(panelCfg.Host)
	user := strings.TrimSpace(panelCfg.User)
	keyPath := strings.TrimSpace(panelCfg.KeyPath)
	if host == "" {
		host = strings.TrimSpace(os.Getenv("DB_VPS_SSH_HOST"))
	}
	if user == "" {
		user = strings.TrimSpace(os.Getenv("DB_VPS_SSH_USER"))
	}
	if keyPath == "" {
		keyPath = strings.TrimSpace(os.Getenv("DB_VPS_SSH_KEY_PATH"))
	}
	if host == "" || user == "" {
		return rustDeskRemoteConfig{}, fmt.Errorf("faltan DB_VPS_SSH_HOST o DB_VPS_SSH_USER para gestionar RustDesk en el VPS")
	}
	if keyPath == "" {
		keyPath = filepath.Join("..", "clave privada ssh.ppk")
	}
	resolvedKeyPath := keyPath
	if !filepath.IsAbs(resolvedKeyPath) {
		if absPath, err := filepath.Abs(resolvedKeyPath); err == nil {
			resolvedKeyPath = absPath
		}
	}
	if _, err := os.Stat(resolvedKeyPath); err != nil {
		return rustDeskRemoteConfig{}, fmt.Errorf("no se encontro la llave SSH de RustDesk: %s", resolvedKeyPath)
	}
	usePlink := strings.EqualFold(filepath.Ext(resolvedKeyPath), ".ppk")
	execPath := ""
	if usePlink {
		execPath = resolvePlinkPath()
		if execPath == "" {
			return rustDeskRemoteConfig{}, fmt.Errorf("no se encontro plink.exe para usar la llave SSH .ppk")
		}
	} else {
		execPath = resolveSSHPath()
		if execPath == "" {
			return rustDeskRemoteConfig{}, fmt.Errorf("no se encontro ssh.exe para ejecutar comandos remotos de RustDesk")
		}
	}
	return rustDeskRemoteConfig{
		Host:     host,
		User:     user,
		KeyPath:  resolvedKeyPath,
		ExecPath: execPath,
		UsePlink: usePlink,
	}, nil
}

func setRustDeskPanelEnabled(dbSuper *sql.DB, enabled bool) error {
	if dbSuper == nil {
		return fmt.Errorf("configuracion no disponible (dbSuper nil)")
	}
	val := "0"
	if enabled {
		val = "1"
	}
	return dbpkg.SetConfigValue(dbSuper, "rustdesk.vps_ssh_enabled", val, false)
}

func resolvePlinkPath() string {
	if path, err := exec.LookPath("plink.exe"); err == nil {
		return path
	}
	candidates := []string{
		`D:\Program Files\PuTTY\plink.exe`,
		`C:\Program Files\PuTTY\plink.exe`,
		`C:\Program Files (x86)\PuTTY\plink.exe`,
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func resolveSSHPath() string {
	if path, err := exec.LookPath("ssh.exe"); err == nil {
		return path
	}
	candidates := []string{
		`C:\Windows\System32\OpenSSH\ssh.exe`,
		`C:\Program Files\Git\usr\bin\ssh.exe`,
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func shellEscapeForPOSIX(value string) string {
	if value == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}

func firstNonEmptyLine(raw string) string {
	for _, line := range strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
