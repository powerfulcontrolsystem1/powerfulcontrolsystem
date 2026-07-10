package scanner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/you/pos-backend/vpssecurity/config"
	"github.com/you/pos-backend/vpssecurity/reports"
)

func runCustomChecks(ctx context.Context, settings config.Settings, executor Executor) (reports.ToolResult, []reports.Finding, []Artifact, reports.SystemInfo) {
	tool := reports.ToolResult{Name: "custom_checks", DisplayName: "Chequeos del servidor", Available: true, Executed: true, Status: "ok"}
	if runtime.GOOS != "linux" {
		tool.Status = "omitido"
		tool.Summary = "chequeos completos solo disponibles en Linux"
		return tool, nil, nil, reports.SystemInfo{}
	}
	start := time.Now()
	findings := make([]reports.Finding, 0)
	artifacts := make([]Artifact, 0)
	info := reports.SystemInfo{}

	firewallFinding, firewallText := checkFirewall(ctx, executor)
	info.Firewall = firewallText
	if firewallFinding.Title != "" {
		findings = append(findings, firewallFinding)
	}
	artifacts = append(artifacts, Artifact{Name: "raw/firewall.txt", Content: []byte(firewallText + "\n")})

	nginxFindings, nginxText, nginxDetected := checkNginxConfig()
	info.NginxDetected = nginxDetected
	findings = append(findings, nginxFindings...)
	if nginxText != "" {
		artifacts = append(artifacts, Artifact{Name: "raw/nginx-check.txt", Content: []byte(nginxText)})
	}

	sshFindings, sshText, sshChecked := checkSSHConfig()
	info.SSHConfigChecked = sshChecked
	findings = append(findings, sshFindings...)
	if sshText != "" {
		artifacts = append(artifacts, Artifact{Name: "raw/ssh-check.txt", Content: []byte(sshText)})
	}

	permissionFindings, permissionText := checkCriticalPermissions()
	findings = append(findings, permissionFindings...)
	if permissionText != "" {
		artifacts = append(artifacts, Artifact{Name: "raw/permissions-check.txt", Content: []byte(permissionText)})
	}

	serviceFindings, services, servicesText := checkRunningServices(ctx, executor)
	info.RunningServices = services
	findings = append(findings, serviceFindings...)
	if servicesText != "" {
		artifacts = append(artifacts, Artifact{Name: "raw/services-check.txt", Content: []byte(servicesText)})
	}

	updateFindings, upgradablePackages, updatesText := checkUpdates(ctx, executor)
	info.UpgradablePackages = upgradablePackages
	findings = append(findings, updateFindings...)
	if updatesText != "" {
		artifacts = append(artifacts, Artifact{Name: "raw/updates-check.txt", Content: []byte(updatesText)})
	}

	tool.DurationMs = time.Since(start).Milliseconds()
	tool.Summary = fmt.Sprintf("%d hallazgos, firewall=%s, actualizaciones pendientes=%d", len(findings), defaultSummaryValue(info.Firewall, "desconocido"), info.UpgradablePackages)
	if len(findings) == 0 {
		tool.Summary = "sin hallazgos en chequeos custom"
	}
	return tool, findings, artifacts, info
}

func checkFirewall(ctx context.Context, executor Executor) (reports.Finding, string) {
	status := "sin datos"
	cmds := []struct {
		command string
		args    []string
	}{
		{command: "ufw", args: []string{"status"}},
		{command: "nft", args: []string{"list", "ruleset"}},
		{command: "iptables", args: []string{"-S"}},
	}
	for _, candidate := range cmds {
		if _, err := execLookPath(candidate.command); err != nil {
			continue
		}
		runCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		output, err := executor.Run(runCtx, candidate.command, candidate.args, "")
		cancel()
		text := strings.TrimSpace(string(output))
		if text == "" && err != nil {
			continue
		}
		status = text
		lower := strings.ToLower(text)
		switch candidate.command {
		case "ufw":
			if strings.Contains(lower, "status: inactive") {
				return reports.Finding{Tool: "custom", Category: "firewall", Severity: reports.SeverityHigh, Title: "UFW inactivo", Description: "El firewall UFW aparece desactivado en el VPS.", Recommendation: "Active UFW y permita solo puertos realmente necesarios.", Target: settingsTargetLocalhost(), Evidence: text}, status
			}
			if strings.Contains(lower, "default: deny (incoming)") {
				return reports.Finding{}, "UFW activo con politica deny incoming"
			}
			return reports.Finding{Tool: "custom", Category: "firewall", Severity: reports.SeverityMedium, Title: "Firewall sin politica clara de entrada", Description: "UFW esta presente pero no muestra una politica endurecida de entrada.", Recommendation: "Asegure una politica default deny incoming y documente las aperturas necesarias.", Target: settingsTargetLocalhost(), Evidence: text}, status
		case "nft":
			if strings.TrimSpace(text) == "" {
				return reports.Finding{Tool: "custom", Category: "firewall", Severity: reports.SeverityHigh, Title: "nftables sin reglas cargadas", Description: "No se encontraron reglas activas en nftables.", Recommendation: "Cree un ruleset minimo con politica restrictiva y servicios permitidos.", Target: settingsTargetLocalhost()}, status
			}
			return reports.Finding{}, "nftables con reglas activas"
		case "iptables":
			if strings.Contains(lower, "-p input accept") || strings.Contains(lower, "-a input -j accept") {
				return reports.Finding{Tool: "custom", Category: "firewall", Severity: reports.SeverityHigh, Title: "iptables acepta trafico de entrada sin restriccion", Description: "Las reglas observadas sugieren una politica permisiva de entrada.", Recommendation: "Cambie la politica de INPUT a DROP/DENY y permita explicitamente solo puertos autorizados.", Target: settingsTargetLocalhost(), Evidence: text}, status
			}
			return reports.Finding{}, "iptables con reglas activas"
		}
	}
	return reports.Finding{Tool: "custom", Category: "firewall", Severity: reports.SeverityMedium, Title: "No se pudo determinar el estado del firewall", Description: "El sistema no encontro UFW, nftables ni iptables accesibles durante el escaneo.", Recommendation: "Instale o habilite una capa de firewall administrada y valide su estado desde el panel super.", Target: settingsTargetLocalhost()}, status
}

func checkNginxConfig() ([]reports.Finding, string, bool) {
	if runtime.GOOS != "linux" {
		return nil, "", false
	}
	if _, err := os.Stat("/etc/nginx"); err != nil {
		return nil, "nginx no detectado", false
	}
	configs := make([]string, 0)
	_ = filepath.Walk("/etc/nginx", func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(info.Name()), ".conf") {
			configs = append(configs, path)
		}
		return nil
	})
	combined := &strings.Builder{}
	findings := make([]reports.Finding, 0)
	hasTLS := false
	hasHSTS := false
	hasFrameOptions := false
	hasNoSniff := false
	for _, cfg := range configs {
		raw, err := os.ReadFile(cfg)
		if err != nil {
			continue
		}
		text := string(raw)
		combined.WriteString("# " + cfg + "\n")
		combined.WriteString(text)
		combined.WriteString("\n")
		lower := strings.ToLower(text)
		if strings.Contains(lower, "listen 443") || strings.Contains(lower, "ssl_certificate") {
			hasTLS = true
		}
		if strings.Contains(lower, "strict-transport-security") {
			hasHSTS = true
		}
		if strings.Contains(lower, "x-frame-options") {
			hasFrameOptions = true
		}
		if strings.Contains(lower, "x-content-type-options") {
			hasNoSniff = true
		}
		if strings.Contains(lower, "server_tokens on") {
			findings = append(findings, reports.Finding{Tool: "custom", Category: "nginx", Severity: reports.SeverityMedium, Title: "Nginx expone server_tokens on", Description: "La configuracion expone version o firma del servidor web.", Recommendation: "Cambie server_tokens a off para reducir fingerprinting.", Target: settingsTargetLocalhost(), Evidence: cfg})
		}
		if strings.Contains(lower, "autoindex on") {
			findings = append(findings, reports.Finding{Tool: "custom", Category: "nginx", Severity: reports.SeverityHigh, Title: "Nginx tiene autoindex habilitado", Description: "El listado de directorios puede exponer archivos o rutas sensibles.", Recommendation: "Desactive autoindex o restrinja el bloque afectado.", Target: settingsTargetLocalhost(), Evidence: cfg})
		}
		if strings.Contains(lower, "tlsv1;") || strings.Contains(lower, "tlsv1.1") {
			findings = append(findings, reports.Finding{Tool: "custom", Category: "nginx", Severity: reports.SeverityHigh, Title: "Nginx permite protocolos TLS obsoletos", Description: "Se detectaron versiones TLS legacy en la configuracion del servidor web.", Recommendation: "Permita solo TLSv1.2 y TLSv1.3 en ssl_protocols.", Target: settingsTargetLocalhost(), Evidence: cfg})
		}
	}
	if hasTLS && !hasHSTS {
		findings = append(findings, reports.Finding{Tool: "custom", Category: "nginx", Severity: reports.SeverityMedium, Title: "Nginx sin cabecera HSTS", Description: "Se detecto TLS en Nginx pero no una cabecera Strict-Transport-Security visible en la configuracion.", Recommendation: "Agregue HSTS en los hosts HTTPS una vez validado el dominio y la redirección segura.", Target: settingsTargetLocalhost()})
	}
	if !hasFrameOptions {
		findings = append(findings, reports.Finding{Tool: "custom", Category: "nginx", Severity: reports.SeverityLow, Title: "Falta X-Frame-Options en Nginx", Description: "No se detecto una cabecera X-Frame-Options en los archivos revisados.", Recommendation: "Agregue X-Frame-Options DENY o SAMEORIGIN segun corresponda.", Target: settingsTargetLocalhost()})
	}
	if !hasNoSniff {
		findings = append(findings, reports.Finding{Tool: "custom", Category: "nginx", Severity: reports.SeverityLow, Title: "Falta X-Content-Type-Options en Nginx", Description: "No se detecto la cabecera de endurecimiento X-Content-Type-Options.", Recommendation: "Agregue X-Content-Type-Options nosniff a los bloques HTTP/HTTPS.", Target: settingsTargetLocalhost()})
	}
	return findings, combined.String(), true
}

func checkSSHConfig() ([]reports.Finding, string, bool) {
	raw, err := os.ReadFile("/etc/ssh/sshd_config")
	if err != nil {
		return nil, "", false
	}
	text := string(raw)
	findings := make([]reports.Finding, 0)
	lowerLines := strings.Split(strings.ToLower(text), "\n")
	for _, rawLine := range lowerLines {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		switch {
		case strings.HasPrefix(line, "permitrootlogin yes"):
			findings = append(findings, reports.Finding{Tool: "custom", Category: "ssh", Severity: reports.SeverityHigh, Title: "SSH permite root login", Description: "PermitRootLogin esta activo en sshd_config.", Recommendation: "Cambie PermitRootLogin a no o prohibit-password y use sudo con cuentas nominativas.", Target: settingsTargetLocalhost(), Evidence: line})
		case strings.HasPrefix(line, "passwordauthentication yes"):
			findings = append(findings, reports.Finding{Tool: "custom", Category: "ssh", Severity: reports.SeverityMedium, Title: "SSH permite autenticacion por password", Description: "La autenticacion por password sigue habilitada en SSH.", Recommendation: "Use llaves publicas y deshabilite PasswordAuthentication si la operacion lo permite.", Target: settingsTargetLocalhost(), Evidence: line})
		case strings.HasPrefix(line, "protocol 1"):
			findings = append(findings, reports.Finding{Tool: "custom", Category: "ssh", Severity: reports.SeverityCritical, Title: "SSH aun permite protocolo 1", Description: "Se detecto una configuracion insegura del protocolo SSH.", Recommendation: "Mantenga solo protocolo SSH 2.", Target: settingsTargetLocalhost(), Evidence: line})
		}
	}
	return findings, text, true
}

func checkCriticalPermissions() ([]reports.Finding, string) {
	paths := []struct {
		Path          string
		ForbiddenMask os.FileMode
		Severity      reports.Severity
		Title         string
		Action        string
	}{
		// Ubuntu permite root:shadow 640: la lectura del grupo shadow es necesaria
		// para PAM, pero no se permiten escritura/ejecucion de grupo ni acceso otros.
		{Path: "/etc/shadow", ForbiddenMask: 0o037, Severity: reports.SeverityHigh, Title: "Permisos inseguros en /etc/shadow", Action: "Ajuste permisos a 640 o mas restrictivos y valide propietario root:shadow."},
		{Path: "/etc/sudoers", ForbiddenMask: 0o022, Severity: reports.SeverityHigh, Title: "Permisos inseguros en /etc/sudoers", Action: "Mantenga /etc/sudoers solo editable por root y valide sintaxis con visudo."},
		{Path: "/root/.ssh", ForbiddenMask: 0o077, Severity: reports.SeverityMedium, Title: "Permisos inseguros en /root/.ssh", Action: "Ajuste el directorio a 700 y los archivos sensibles a 600."},
		{Path: "/root/.ssh/authorized_keys", ForbiddenMask: 0o077, Severity: reports.SeverityMedium, Title: "Permisos inseguros en authorized_keys", Action: "Mantenga authorized_keys con permisos 600 y propietario root."},
	}
	findings := make([]reports.Finding, 0)
	builder := &strings.Builder{}
	for _, item := range paths {
		info, err := os.Stat(item.Path)
		if err != nil {
			continue
		}
		mode := info.Mode().Perm()
		builder.WriteString(item.Path + " -> " + mode.String() + "\n")
		if mode&item.ForbiddenMask != 0 {
			findings = append(findings, reports.Finding{Tool: "custom", Category: "permisos", Severity: item.Severity, Title: item.Title, Description: "Se detectaron permisos mas amplios de lo recomendado sobre un archivo/directorio sensible.", Recommendation: item.Action, Target: settingsTargetLocalhost(), Evidence: fmt.Sprintf("%s=%s", item.Path, mode.String())})
		}
	}
	return findings, builder.String()
}

func checkRunningServices(ctx context.Context, executor Executor) ([]reports.Finding, []string, string) {
	if _, err := execLookPath("systemctl"); err != nil {
		return nil, nil, "systemctl no disponible"
	}
	runCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	output, err := executor.Run(runCtx, "systemctl", []string{"list-units", "--type=service", "--state=running", "--no-legend", "--no-pager"}, "")
	if err != nil && strings.TrimSpace(string(output)) == "" {
		return nil, nil, ""
	}
	text := string(output)
	services := make([]string, 0)
	findings := make([]reports.Finding, 0)
	suspicious := map[string]reports.Severity{
		"telnet.service":       reports.SeverityCritical,
		"vsftpd.service":       reports.SeverityHigh,
		"avahi-daemon.service": reports.SeverityMedium,
		"cups.service":         reports.SeverityMedium,
		"rpcbind.service":      reports.SeverityHigh,
		"smbd.service":         reports.SeverityHigh,
	}
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) == 0 {
			continue
		}
		service := fields[0]
		services = append(services, service)
		if severity, ok := suspicious[service]; ok {
			findings = append(findings, reports.Finding{Tool: "custom", Category: "servicios", Severity: severity, Title: "Servicio sensible en ejecucion: " + service, Description: "Se encontro un servicio que suele requerir validacion expresa en servidores productivos.", Recommendation: "Confirme si el servicio es necesario y restrinja su exposicion o deshabilitelo.", Target: settingsTargetLocalhost(), Evidence: service})
		}
	}
	sort.Strings(services)
	return findings, services, text
}

func checkUpdates(ctx context.Context, executor Executor) ([]reports.Finding, int, string) {
	if _, err := execLookPath("apt"); err != nil {
		return nil, 0, "apt no disponible"
	}
	runCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	output, err := executor.Run(runCtx, "apt", []string{"list", "--upgradable"}, "")
	if err != nil && strings.TrimSpace(string(output)) == "" {
		return nil, 0, ""
	}
	text := strings.TrimSpace(string(output))
	if text == "" {
		return nil, 0, ""
	}
	count := 0
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(strings.ToLower(line), "listing") {
			continue
		}
		count++
	}
	if count == 0 {
		return nil, 0, text
	}
	severity := reports.SeverityMedium
	if count > 40 {
		severity = reports.SeverityHigh
	}
	return []reports.Finding{{Tool: "custom", Category: "actualizaciones", Severity: severity, Title: fmt.Sprintf("%d paquetes pendientes de actualizacion", count), Description: "El host reporta paquetes actualizables mediante APT.", Recommendation: "Revise los paquetes pendientes y aplique ventana de mantenimiento para actualizar el sistema.", Target: settingsTargetLocalhost(), Evidence: firstNLines(text, 20)}}, count, text
}

func defaultSummaryValue(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func firstNLines(text string, limit int) string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if len(lines) > limit {
		lines = lines[:limit]
	}
	return strings.Join(lines, "\n")
}

func settingsTargetLocalhost() string {
	return "127.0.0.1"
}

func execLookPath(command string) (string, error) {
	return exec.LookPath(command)
}
