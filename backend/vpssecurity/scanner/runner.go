package scanner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/you/pos-backend/vpssecurity/config"
	"github.com/you/pos-backend/vpssecurity/parser"
	"github.com/you/pos-backend/vpssecurity/reports"
)

type Executor interface {
	Run(ctx context.Context, command string, args []string, workDir string) ([]byte, error)
}

type SystemExecutor struct{}

func (SystemExecutor) Run(ctx context.Context, command string, args []string, workDir string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	if strings.TrimSpace(workDir) != "" {
		cmd.Dir = workDir
	}
	if filepath.Base(command) == "trivy" {
		cacheDir := strings.TrimSpace(os.Getenv("PCS_VPS_SECURITY_TRIVY_CACHE_DIR"))
		if cacheDir == "" {
			cacheDir = filepath.Join(workDir, "trivy-cache")
		}
		if err := os.MkdirAll(cacheDir, 0o700); err != nil {
			return nil, fmt.Errorf("preparar cache de Trivy: %w", err)
		}
		cmd.Env = append(os.Environ(), "TRIVY_CACHE_DIR="+cacheDir, "TMPDIR="+workDir)
	}
	return cmd.CombinedOutput()
}

type Artifact struct {
	Name    string
	Content []byte
}

type RunResult struct {
	Tools          []reports.ToolResult
	Findings       []reports.Finding
	Artifacts      []Artifact
	SystemInfo     reports.SystemInfo
	Notes          []string
	Errors         []string
	HardeningIndex int
}

func RunAudit(ctx context.Context, settings config.Settings, executor Executor) RunResult {
	if executor == nil {
		executor = SystemExecutor{}
	}
	result := RunResult{
		Tools:      make([]reports.ToolResult, 0, 4),
		Findings:   make([]reports.Finding, 0, 64),
		Artifacts:  make([]Artifact, 0, 8),
		Notes:      make([]string, 0, 8),
		Errors:     make([]string, 0, 4),
		SystemInfo: collectBaseSystemInfo(),
	}
	lynisTool, lynisFindings, lynisArtifacts, hardeningIndex := runLynis(ctx, settings, executor)
	result.Tools = append(result.Tools, lynisTool)
	result.Findings = append(result.Findings, limitFindings(lynisFindings, settings.MaxFindingsPerTool, &result.Notes, "Lynis")...)
	result.Artifacts = append(result.Artifacts, lynisArtifacts...)
	result.HardeningIndex = hardeningIndex

	nmapTool, nmapFindings, nmapArtifacts := runNmap(ctx, settings, executor)
	result.Tools = append(result.Tools, nmapTool)
	result.Findings = append(result.Findings, limitFindings(nmapFindings, settings.MaxFindingsPerTool, &result.Notes, "Nmap")...)
	result.Artifacts = append(result.Artifacts, nmapArtifacts...)

	vulnTool, vulnFindings, vulnArtifacts := runVulnerabilityScanner(ctx, settings, executor)
	result.Tools = append(result.Tools, vulnTool)
	result.Findings = append(result.Findings, limitFindings(vulnFindings, settings.MaxFindingsPerTool, &result.Notes, "Escaner de vulnerabilidades")...)
	result.Artifacts = append(result.Artifacts, vulnArtifacts...)

	customTool, customFindings, customArtifacts, customInfo := runCustomChecks(ctx, settings, executor)
	result.Tools = append(result.Tools, customTool)
	result.Findings = append(result.Findings, limitFindings(customFindings, settings.MaxFindingsPerTool, &result.Notes, "Chequeos custom")...)
	result.Artifacts = append(result.Artifacts, customArtifacts...)
	mergeSystemInfo(&result.SystemInfo, customInfo)
	if result.HardeningIndex > 0 {
		result.Notes = append(result.Notes, fmt.Sprintf("Lynis calculo hardening index %d", result.HardeningIndex))
	}
	return result
}

func collectBaseSystemInfo() reports.SystemInfo {
	hostname, _ := os.Hostname()
	info := reports.SystemInfo{
		Hostname: strings.TrimSpace(hostname),
		OS:       runtime.GOOS,
		Kernel:   strings.TrimSpace(readFirstLine("/proc/version")),
	}
	if release := readOSRelease(); release != "" {
		info.OS = release
	}
	return info
}

func readFirstLine(path string) string {
	// #nosec G304 -- path is normalized and constrained to a server-controlled root before this operation.
	raw, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n")
	if len(lines) == 0 {
		return ""
	}
	return strings.TrimSpace(lines[0])
}

func readOSRelease() string {
	raw, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			return strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "PRETTY_NAME=")), "\"")
		}
	}
	return ""
}

func runLynis(ctx context.Context, settings config.Settings, executor Executor) (reports.ToolResult, []reports.Finding, []Artifact, int) {
	tool := reports.ToolResult{Name: "lynis", DisplayName: "Lynis", Available: false, Executed: false, Status: "omitido"}
	if !settings.Lynis.Enabled {
		tool.Summary = "deshabilitado en configuracion"
		return tool, nil, nil, 0
	}
	if runtime.GOOS != "linux" {
		tool.Summary = "solo se ejecuta en Linux"
		return tool, nil, nil, 0
	}
	if _, err := exec.LookPath(settings.Lynis.Command); err != nil {
		tool.Status = "no_disponible"
		tool.Summary = "Lynis no esta instalado"
		tool.Error = err.Error()
		return tool, []reports.Finding{missingToolFinding("lynis", settings.Lynis.Command)}, nil, 0
	}
	start := time.Now()
	workDir, cleanup := makeTempDir("vpssecurity-lynis-")
	defer cleanup()
	reportFile := filepath.Join(workDir, "lynis-report.dat")
	logFile := filepath.Join(workDir, "lynis.log")
	args := []string{"audit", "system", "--quick", "--quiet", "--report-file", reportFile, "--logfile", logFile}
	tool.Command = settings.Lynis.Command + " " + strings.Join(args, " ")
	runCtx, cancel := context.WithTimeout(ctx, time.Duration(settings.Lynis.TimeoutSeconds)*time.Second)
	defer cancel()
	output, err := executor.Run(runCtx, settings.Lynis.Command, args, workDir)
	tool.Available = true
	tool.DurationMs = time.Since(start).Milliseconds()
	artifacts := make([]Artifact, 0, 2)
	// #nosec G304 -- path is normalized and constrained to a server-controlled root before this operation.
	if reportRaw, readErr := os.ReadFile(reportFile); readErr == nil {
		artifacts = append(artifacts, Artifact{Name: "raw/lynis-report.dat", Content: reportRaw})
		findings, hardeningIndex, summary := parser.ParseLynisReport(reportRaw, settings.TargetHost)
		tool.Executed = err == nil
		if err != nil {
			tool.Status = "parcial"
			tool.Error = strings.TrimSpace(string(output))
		} else {
			tool.Status = "ok"
		}
		tool.Summary = summary
		// #nosec G304 -- path is normalized and constrained to a server-controlled root before this operation.
		if logRaw, logErr := os.ReadFile(logFile); logErr == nil {
			artifacts = append(artifacts, Artifact{Name: "raw/lynis.log", Content: logRaw})
		}
		return tool, findings, artifacts, hardeningIndex
	}
	tool.Status = "error"
	if err != nil {
		tool.Error = strings.TrimSpace(string(output))
	}
	tool.Summary = "Lynis no genero archivo de reporte"
	artifacts = append(artifacts, Artifact{Name: "raw/lynis-output.txt", Content: output})
	return tool, nil, artifacts, 0
}

func runNmap(ctx context.Context, settings config.Settings, executor Executor) (reports.ToolResult, []reports.Finding, []Artifact) {
	tool := reports.ToolResult{Name: "nmap", DisplayName: "Nmap", Available: false, Executed: false, Status: "omitido"}
	if !settings.Nmap.Enabled {
		tool.Summary = "deshabilitado en configuracion"
		return tool, nil, nil
	}
	if runtime.GOOS != "linux" {
		tool.Summary = "solo se ejecuta en Linux"
		return tool, nil, nil
	}
	if _, err := exec.LookPath(settings.Nmap.Command); err != nil {
		tool.Status = "no_disponible"
		tool.Summary = "Nmap no esta instalado"
		tool.Error = err.Error()
		return tool, []reports.Finding{missingToolFinding("nmap", settings.Nmap.Command)}, nil
	}
	start := time.Now()
	workDir, cleanup := makeTempDir("vpssecurity-nmap-")
	defer cleanup()
	outputFile := filepath.Join(workDir, "nmap.xml")
	ports := strings.TrimSpace(settings.PortList)
	if settings.Profile == "quick" {
		ports = quickPorts(ports)
	}
	args := []string{"-Pn", "-sV", "-T4", "-p", ports, "-oX", outputFile}
	if settings.Profile == "full" {
		args = append(args, "-sC")
	}
	args = append(args, settings.TargetHost)
	tool.Command = settings.Nmap.Command + " " + strings.Join(args, " ")
	runCtx, cancel := context.WithTimeout(ctx, time.Duration(settings.Nmap.TimeoutSeconds)*time.Second)
	defer cancel()
	output, err := executor.Run(runCtx, settings.Nmap.Command, args, workDir)
	tool.Available = true
	tool.DurationMs = time.Since(start).Milliseconds()
	artifacts := make([]Artifact, 0, 2)
	// #nosec G304 -- path is normalized and constrained to a server-controlled root before this operation.
	if raw, readErr := os.ReadFile(outputFile); readErr == nil {
		artifacts = append(artifacts, Artifact{Name: "raw/nmap.xml", Content: raw})
		findings, openPorts, summary, parseErr := parser.ParseNmapXML(raw, settings.TargetHost)
		if parseErr != nil {
			tool.Status = "error"
			tool.Summary = "Nmap genero XML invalido"
			tool.Error = parseErr.Error()
			return tool, nil, artifacts
		}
		tool.Executed = err == nil
		if err != nil {
			tool.Status = "parcial"
			tool.Error = strings.TrimSpace(string(output))
		} else {
			tool.Status = "ok"
		}
		tool.Summary = fmt.Sprintf("%s (%s)", summary, joinPorts(openPorts))
		artifacts = append(artifacts, Artifact{Name: "raw/nmap-output.txt", Content: output})
		return tool, findings, artifacts
	}
	tool.Status = "error"
	tool.Summary = "Nmap no genero archivo XML"
	tool.Error = strings.TrimSpace(string(output))
	artifacts = append(artifacts, Artifact{Name: "raw/nmap-output.txt", Content: output})
	return tool, nil, artifacts
}

func quickPorts(portList string) string {
	if strings.TrimSpace(portList) == "" {
		return "49222,80,443"
	}
	parts := strings.Split(portList, ",")
	selected := make([]string, 0, 3)
	allowed := map[string]struct{}{"22": {}, "80": {}, "443": {}}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if _, ok := allowed[part]; ok {
			selected = append(selected, part)
		}
	}
	if len(selected) == 0 {
		return "49222,80,443"
	}
	return strings.Join(selected, ",")
}

func joinPorts(ports []int) string {
	if len(ports) == 0 {
		return "sin puertos abiertos"
	}
	values := make([]string, 0, len(ports))
	for _, port := range ports {
		values = append(values, strconv.Itoa(port))
	}
	return "puertos abiertos: " + strings.Join(values, ",")
}

func runVulnerabilityScanner(ctx context.Context, settings config.Settings, executor Executor) (reports.ToolResult, []reports.Finding, []Artifact) {
	tool := reports.ToolResult{Name: "vulnerability_scan", DisplayName: "Escaner de vulnerabilidades", Available: false, Executed: false, Status: "omitido"}
	if !settings.VulnerabilityScan.Enabled {
		tool.Summary = "deshabilitado en configuracion"
		return tool, nil, nil
	}
	if runtime.GOOS != "linux" {
		tool.Summary = "solo se ejecuta en Linux"
		return tool, nil, nil
	}
	provider := strings.ToLower(strings.TrimSpace(settings.VulnerabilityScan.Provider))
	if provider != "trivy" {
		tool.Status = "no_soportado"
		tool.Summary = "la implementacion actual soporta Trivy como alternativa ligera a OpenVAS"
		return tool, nil, nil
	}
	if _, err := exec.LookPath(settings.VulnerabilityScan.Command); err != nil {
		tool.Status = "no_disponible"
		tool.Summary = "Trivy no esta instalado"
		tool.Error = err.Error()
		return tool, []reports.Finding{missingToolFinding("trivy", settings.VulnerabilityScan.Command)}, nil
	}
	start := time.Now()
	workDir, cleanup := makeTempDir("vpssecurity-trivy-")
	defer cleanup()
	outputFile := filepath.Join(workDir, "trivy.json")
	targetPath := strings.TrimSpace(settings.VulnerabilityScan.TargetPath)
	if targetPath == "" {
		targetPath = "/"
	}
	args := trivyRootfsArgs(outputFile, targetPath)
	tool.Command = settings.VulnerabilityScan.Command + " " + strings.Join(args, " ")
	runCtx, cancel := context.WithTimeout(ctx, time.Duration(settings.VulnerabilityScan.TimeoutSeconds)*time.Second)
	defer cancel()
	output, err := executor.Run(runCtx, settings.VulnerabilityScan.Command, args, workDir)
	tool.Available = true
	tool.DurationMs = time.Since(start).Milliseconds()
	artifacts := make([]Artifact, 0, 2)
	// #nosec G304 -- path is normalized and constrained to a server-controlled root before this operation.
	if raw, readErr := os.ReadFile(outputFile); readErr == nil {
		artifacts = append(artifacts, Artifact{Name: "raw/trivy.json", Content: raw})
		findings, summary, parseErr := parser.ParseTrivyJSON(raw, settings.TargetHost)
		if parseErr != nil {
			tool.Status = "error"
			tool.Summary = "Trivy genero JSON invalido"
			tool.Error = parseErr.Error()
			return tool, nil, artifacts
		}
		tool.Executed = err == nil
		if err != nil {
			tool.Status = "parcial"
			tool.Error = strings.TrimSpace(string(output))
		} else {
			tool.Status = "ok"
		}
		tool.Summary = summary
		artifacts = append(artifacts, Artifact{Name: "raw/trivy-output.txt", Content: output})
		return tool, findings, artifacts
	}
	tool.Status = "error"
	tool.Summary = "Trivy no genero archivo JSON"
	tool.Error = strings.TrimSpace(string(output))
	artifacts = append(artifacts, Artifact{Name: "raw/trivy-output.txt", Content: output})
	return tool, nil, artifacts
}

func trivyRootfsArgs(outputFile, targetPath string) []string {
	return []string{
		"rootfs",
		"--format", "json",
		"--quiet",
		"--scanners", "vuln,misconfig",
		"--severity", "CRITICAL,HIGH,MEDIUM,LOW",
		"--skip-dirs", "/proc",
		"--skip-dirs", "/sys",
		"--skip-dirs", "/dev",
		"--skip-dirs", "/run",
		"--skip-dirs", "/var/lib/docker",
		"--skip-files", "/etc/shadow",
		"--skip-files", "/etc/shadow-",
		"--skip-files", "/etc/gshadow",
		"--skip-files", "/etc/gshadow-",
		"--output", outputFile,
		targetPath,
	}
}

func limitFindings(findings []reports.Finding, limit int, notes *[]string, label string) []reports.Finding {
	if limit <= 0 || len(findings) <= limit {
		return findings
	}
	if notes != nil {
		*notes = append(*notes, fmt.Sprintf("%s se truncó de %d a %d hallazgos para mantener reportes manejables", label, len(findings), limit))
	}
	return findings[:limit]
}

func mergeSystemInfo(target *reports.SystemInfo, incoming reports.SystemInfo) {
	if target == nil {
		return
	}
	if strings.TrimSpace(incoming.Firewall) != "" {
		target.Firewall = incoming.Firewall
	}
	if incoming.NginxDetected {
		target.NginxDetected = true
	}
	if incoming.SSHConfigChecked {
		target.SSHConfigChecked = true
	}
	if incoming.UpgradablePackages > 0 {
		target.UpgradablePackages = incoming.UpgradablePackages
	}
	if len(incoming.RunningServices) > 0 {
		target.RunningServices = append([]string(nil), incoming.RunningServices...)
	}
}

func missingToolFinding(tool, command string) reports.Finding {
	return reports.Finding{
		Tool:           tool,
		Category:       "tooling",
		Severity:       reports.SeverityMedium,
		Title:          fmt.Sprintf("%s no esta disponible en el VPS", strings.ToUpper(tool)),
		Description:    "El modulo de seguridad no pudo ejecutar una herramienta requerida porque el comando no existe en el host.",
		Recommendation: fmt.Sprintf("Instale %s con scripts/install_vps_security_tools.sh o ajuste el comando en la configuracion del panel.", command),
		Target:         "localhost",
		Evidence:       command,
	}
}

func makeTempDir(prefix string) (string, func()) {
	baseDir := strings.TrimSpace(os.Getenv("PCS_VPS_SECURITY_TMP_DIR"))
	if baseDir != "" {
		if err := os.MkdirAll(baseDir, 0o700); err != nil {
			return "", func() {}
		}
	}
	workDir, err := os.MkdirTemp(baseDir, prefix)
	if err != nil {
		return "", func() {}
	}
	return workDir, func() { _ = os.RemoveAll(workDir) }
}

func isCommandUnavailable(err error) bool {
	if err == nil {
		return false
	}
	var execErr *exec.Error
	return errors.As(err, &execErr)
}
