package handlers

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type ServicioEstado struct {
	ID          string            `json:"id"`
	Nombre      string            `json:"nombre"`
	Estado      string            `json:"estado"`
	Detalle     string            `json:"detalle"`
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

func buildRustDeskServiceState(includeProbe bool) ServicioEstado {
	hbbsService := resolveRustDeskServiceName(rustDeskHBBSCandidates)
	hbbrService := resolveRustDeskServiceName(rustDeskHBBRCandidates)
	hbbsStatus := checkSystemctlStatus(hbbsService)
	hbbrStatus := checkSystemctlStatus(hbbrService)
	overall := "inactive"
	switch {
	case hbbsStatus == "active" && hbbrStatus == "active":
		overall = "active"
	case hbbsStatus == "error" || hbbrStatus == "error":
		overall = "error"
	case hbbsStatus == "active" || hbbrStatus == "active":
		overall = "degraded"
	}
	state := ServicioEstado{
		ID:      "rustdesk",
		Nombre:  "RustDesk (Soporte Remoto)",
		Estado:  overall,
		Detalle: "Servidor ID/Relay para soporte remoto de clientes a traves de VPS.",
		Componentes: map[string]string{
			"rustdesk-hbbs": hbbsStatus,
			"rustdesk-hbbr": hbbrStatus,
		},
	}
	if includeProbe {
		probe := probeRustDeskService()
		state.Prueba = &probe
	}
	return state
}

func SuperServidoresListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		servicios := []ServicioEstado{buildRustDeskServiceState(false)}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "servicios": servicios})
	}
}

func SuperServidoresToggleHandler() http.HandlerFunc {
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
			if payload.Accion == "start" || payload.Accion == "stop" || payload.Accion == "restart" {
				err1 := runSystemctl(payload.Accion, resolveRustDeskServiceName(rustDeskHBBSCandidates))
				err2 := runSystemctl(payload.Accion, resolveRustDeskServiceName(rustDeskHBBRCandidates))
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
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "servicio": buildRustDeskServiceState(false)})
	}
}

func SuperServidoresProbeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		service := strings.TrimSpace(r.URL.Query().Get("id"))
		if service == "" {
			service = "rustdesk"
		}
		if service != "rustdesk" {
			http.Error(w, "servicio no soportado", http.StatusBadRequest)
			return
		}
		state := buildRustDeskServiceState(true)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "servicio": state})
	}
}

func checkSystemctlStatus(service string) string {
	if strings.TrimSpace(service) == "" {
		return "missing"
	}
	if shouldUseRustDeskRemoteExec() {
		out, err := runRustDeskRemoteShell(fmt.Sprintf("systemctl is-active %s 2>/dev/null || true", shellEscapeForPOSIX(service)))
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

func runSystemctl(accion string, service string) error {
	if strings.TrimSpace(service) == "" {
		return fmt.Errorf("no se encontro ninguna unidad systemd compatible con RustDesk en el VPS")
	}
	if shouldUseRustDeskRemoteExec() {
		_, err := runRustDeskRemoteShell(fmt.Sprintf("sudo -n systemctl %s %s", shellEscapeForPOSIX(accion), shellEscapeForPOSIX(service)))
		return err
	}
	if runtime.GOOS == "windows" {
		return fmt.Errorf("control local de RustDesk no disponible en Windows sin configuracion SSH al VPS")
	}
	cmd := exec.Command("sudo", "systemctl", accion, service)
	err := cmd.Run()
	return err
}

func probeRustDeskService() ServicioPrueba {
	probe := ServicioPrueba{
		OK:       false,
		Resumen:  "Comprobacion no disponible en este entorno.",
		Revisado: time.Now().In(time.Local).Format("2006-01-02 15:04:05"),
		Puertos:  map[string]string{},
		Servicios: map[string]string{
			"rustdesk-hbbs": checkSystemctlStatus(resolveRustDeskServiceName(rustDeskHBBSCandidates)),
			"rustdesk-hbbr": checkSystemctlStatus(resolveRustDeskServiceName(rustDeskHBBRCandidates)),
		},
	}
	if runtime.GOOS == "windows" {
		if !shouldUseRustDeskRemoteExec() {
			probe.Resumen = "Entorno local Windows: configura DB_VPS_SSH_HOST/USER/KEY_PATH para ejecutar la prueba real en el VPS Linux."
			return probe
		}
	}

	ports := []int{21114, 21115, 21116, 21117, 21118, 21119}
	openPorts := 0
	if shouldUseRustDeskRemoteExec() {
		ssOutput, err := runRustDeskRemoteShell("ss -ltn 2>/dev/null || true")
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
		if shouldUseRustDeskRemoteExec() {
			probe.Resumen = fmt.Sprintf("RustDesk responde en el VPS: hbbs/hbbr activos y %d puerto(s) abiertos.", openPorts)
		} else {
			probe.Resumen = fmt.Sprintf("RustDesk responde: hbbs/hbbr activos y %d puerto(s) locales abiertos.", openPorts)
		}
		return probe
	}
	if shouldUseRustDeskRemoteExec() {
		probe.Resumen = fmt.Sprintf("RustDesk en el VPS con alertas: hbbs=%s, hbbr=%s, puertos abiertos=%d.", probe.Servicios["rustdesk-hbbs"], probe.Servicios["rustdesk-hbbr"], openPorts)
	} else {
		probe.Resumen = fmt.Sprintf("RustDesk con alertas: hbbs=%s, hbbr=%s, puertos abiertos=%d.", probe.Servicios["rustdesk-hbbs"], probe.Servicios["rustdesk-hbbr"], openPorts)
	}
	return probe
}

func resolveRustDeskServiceName(candidates []string) string {
	for _, candidate := range candidates {
		if rustDeskServiceExists(candidate) {
			return candidate
		}
	}
	return ""
}

func rustDeskServiceExists(service string) bool {
	if strings.TrimSpace(service) == "" {
		return false
	}
	command := fmt.Sprintf("systemctl cat %s >/dev/null 2>&1", shellEscapeForPOSIX(service))
	if shouldUseRustDeskRemoteExec() {
		_, err := runRustDeskRemoteShell(command)
		return err == nil
	}
	if runtime.GOOS == "windows" {
		return false
	}
	cmd := exec.Command("sh", "-lc", command)
	return cmd.Run() == nil
}

func shouldUseRustDeskRemoteExec() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	return strings.TrimSpace(os.Getenv("DB_VPS_SSH_HOST")) != "" && strings.TrimSpace(os.Getenv("DB_VPS_SSH_USER")) != ""
}

func runRustDeskRemoteShell(script string) (string, error) {
	cfg, err := resolveRustDeskRemoteConfig()
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

func resolveRustDeskRemoteConfig() (rustDeskRemoteConfig, error) {
	host := strings.TrimSpace(os.Getenv("DB_VPS_SSH_HOST"))
	user := strings.TrimSpace(os.Getenv("DB_VPS_SSH_USER"))
	keyPath := strings.TrimSpace(os.Getenv("DB_VPS_SSH_KEY_PATH"))
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
