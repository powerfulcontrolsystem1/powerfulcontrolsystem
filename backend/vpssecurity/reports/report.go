package reports

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Severity string

const (
	SeverityCritical Severity = "CRITICO"
	SeverityHigh     Severity = "ALTO"
	SeverityMedium   Severity = "MEDIO"
	SeverityLow      Severity = "BAJO"
	SeverityInfo     Severity = "INFO"
)

type Finding struct {
	ID             string   `json:"id"`
	Tool           string   `json:"tool"`
	Category       string   `json:"category"`
	Severity       Severity `json:"severity"`
	Title          string   `json:"title"`
	Description    string   `json:"description,omitempty"`
	Recommendation string   `json:"recommendation,omitempty"`
	Target         string   `json:"target,omitempty"`
	Port           int      `json:"port,omitempty"`
	Service        string   `json:"service,omitempty"`
	Reference      string   `json:"reference,omitempty"`
	Evidence       string   `json:"evidence,omitempty"`
}

type ToolResult struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Available   bool   `json:"available"`
	Executed    bool   `json:"executed"`
	Status      string `json:"status"`
	Summary     string `json:"summary,omitempty"`
	Error       string `json:"error,omitempty"`
	Command     string `json:"command,omitempty"`
	DurationMs  int64  `json:"duration_ms,omitempty"`
	Artifact    string `json:"artifact,omitempty"`
}

type SystemInfo struct {
	Hostname           string   `json:"hostname,omitempty"`
	OS                 string   `json:"os,omitempty"`
	Kernel             string   `json:"kernel,omitempty"`
	Firewall           string   `json:"firewall,omitempty"`
	NginxDetected      bool     `json:"nginx_detected,omitempty"`
	SSHConfigChecked   bool     `json:"ssh_config_checked,omitempty"`
	UpgradablePackages int      `json:"upgradable_packages,omitempty"`
	RunningServices    []string `json:"running_services,omitempty"`
}

type Summary struct {
	Critical         int      `json:"critical"`
	High             int      `json:"high"`
	Medium           int      `json:"medium"`
	Low              int      `json:"low"`
	Info             int      `json:"info"`
	TotalFindings    int      `json:"total_findings"`
	HighestSeverity  string   `json:"highest_severity"`
	OpenPorts        []int    `json:"open_ports,omitempty"`
	HardeningIndex   int      `json:"hardening_index,omitempty"`
	Health           string   `json:"health,omitempty"`
	CoverageComplete bool     `json:"coverage_complete"`
	CoverageStatus   string   `json:"coverage_status"`
	IncompleteTools  []string `json:"incomplete_tools,omitempty"`
}

type ConfigSnapshot struct {
	Scope                string   `json:"scope"`
	TargetHost           string   `json:"target_host"`
	PortList             string   `json:"port_list"`
	Profile              string   `json:"profile"`
	Cron                 string   `json:"cron,omitempty"`
	EnabledTools         []string `json:"enabled_tools,omitempty"`
	VulnerabilityScanner string   `json:"vulnerability_scanner,omitempty"`
}

type Comparison struct {
	PreviousScanID      string         `json:"previous_scan_id,omitempty"`
	PreviousGeneratedAt string         `json:"previous_generated_at,omitempty"`
	NewFindings         int            `json:"new_findings"`
	ResolvedFindings    int            `json:"resolved_findings"`
	SeverityDelta       map[string]int `json:"severity_delta,omitempty"`
	NewOpenPorts        []int          `json:"new_open_ports,omitempty"`
	ClosedPorts         []int          `json:"closed_ports,omitempty"`
	Summary             string         `json:"summary,omitempty"`
}

type ScanReport struct {
	ScanID      string            `json:"scan_id"`
	Status      string            `json:"status"`
	Scope       string            `json:"scope"`
	GeneratedAt string            `json:"generated_at"`
	StartedAt   string            `json:"started_at,omitempty"`
	CompletedAt string            `json:"completed_at,omitempty"`
	DurationMs  int64             `json:"duration_ms,omitempty"`
	Trigger     string            `json:"trigger,omitempty"`
	TriggeredBy string            `json:"triggered_by,omitempty"`
	TargetHost  string            `json:"target_host"`
	Profile     string            `json:"profile"`
	Config      ConfigSnapshot    `json:"config"`
	Tools       []ToolResult      `json:"tools,omitempty"`
	SystemInfo  SystemInfo        `json:"system_info,omitempty"`
	Summary     Summary           `json:"summary"`
	Findings    []Finding         `json:"findings,omitempty"`
	Notes       []string          `json:"notes,omitempty"`
	Errors      []string          `json:"errors,omitempty"`
	Comparison  Comparison        `json:"comparison,omitempty"`
	Reports     map[string]string `json:"reports,omitempty"`
}

type HistoryEntry struct {
	ScanID           string            `json:"scan_id"`
	GeneratedAt      string            `json:"generated_at"`
	Scope            string            `json:"scope"`
	TargetHost       string            `json:"target_host"`
	Profile          string            `json:"profile"`
	Status           string            `json:"status"`
	CoverageStatus   string            `json:"coverage_status"`
	CoverageComplete bool              `json:"coverage_complete"`
	TotalFindings    int               `json:"total_findings"`
	HighestSeverity  string            `json:"highest_severity"`
	NewFindings      int               `json:"new_findings,omitempty"`
	ResolvedFindings int               `json:"resolved_findings,omitempty"`
	Reports          map[string]string `json:"reports,omitempty"`
}

func severityRank(severity Severity) int {
	switch normalizeSeverity(severity) {
	case SeverityCritical:
		return 5
	case SeverityHigh:
		return 4
	case SeverityMedium:
		return 3
	case SeverityLow:
		return 2
	default:
		return 1
	}
}

func normalizeSeverity(severity Severity) Severity {
	switch strings.ToUpper(strings.TrimSpace(string(severity))) {
	case "CRITICAL", "CRITICO", "CRITICA":
		return SeverityCritical
	case "HIGH", "ALTO":
		return SeverityHigh
	case "MEDIUM", "MEDIA", "MEDIO":
		return SeverityMedium
	case "LOW", "BAJO", "BAJA":
		return SeverityLow
	default:
		return SeverityInfo
	}
}

func NormalizeFindings(findings []Finding) []Finding {
	normalized := make([]Finding, 0, len(findings))
	for _, finding := range findings {
		finding.Severity = normalizeSeverity(finding.Severity)
		finding.Tool = strings.TrimSpace(finding.Tool)
		finding.Category = strings.TrimSpace(finding.Category)
		finding.Title = strings.TrimSpace(finding.Title)
		finding.Target = strings.TrimSpace(finding.Target)
		finding.Service = strings.TrimSpace(finding.Service)
		finding.Reference = strings.TrimSpace(finding.Reference)
		finding.Description = strings.TrimSpace(finding.Description)
		finding.Recommendation = strings.TrimSpace(finding.Recommendation)
		finding.Evidence = strings.TrimSpace(finding.Evidence)
		if finding.ID == "" {
			hash := sha256.Sum256([]byte(strings.Join([]string{
				string(finding.Severity),
				finding.Tool,
				finding.Category,
				finding.Title,
				finding.Target,
				strconv.Itoa(finding.Port),
				finding.Service,
				finding.Reference,
			}, "|")))
			finding.ID = hex.EncodeToString(hash[:16])
		}
		normalized = append(normalized, finding)
	}
	sort.Slice(normalized, func(i, j int) bool {
		if severityRank(normalized[i].Severity) != severityRank(normalized[j].Severity) {
			return severityRank(normalized[i].Severity) > severityRank(normalized[j].Severity)
		}
		if normalized[i].Tool != normalized[j].Tool {
			return normalized[i].Tool < normalized[j].Tool
		}
		if normalized[i].Category != normalized[j].Category {
			return normalized[i].Category < normalized[j].Category
		}
		return normalized[i].Title < normalized[j].Title
	})
	return normalized
}

func ApplySummary(report *ScanReport) {
	if report == nil {
		return
	}
	report.Findings = NormalizeFindings(report.Findings)
	summary := report.Summary
	summary.Critical = 0
	summary.High = 0
	summary.Medium = 0
	summary.Low = 0
	summary.Info = 0
	summary.TotalFindings = len(report.Findings)
	portSet := make(map[int]struct{})
	highest := SeverityInfo
	for _, finding := range report.Findings {
		switch normalizeSeverity(finding.Severity) {
		case SeverityCritical:
			summary.Critical++
		case SeverityHigh:
			summary.High++
		case SeverityMedium:
			summary.Medium++
		case SeverityLow:
			summary.Low++
		default:
			summary.Info++
		}
		if severityRank(finding.Severity) > severityRank(highest) {
			highest = normalizeSeverity(finding.Severity)
		}
		if finding.Port > 0 {
			portSet[finding.Port] = struct{}{}
		}
	}
	ports := make([]int, 0, len(portSet))
	for port := range portSet {
		ports = append(ports, port)
	}
	sort.Ints(ports)
	summary.OpenPorts = ports
	summary.HighestSeverity = string(highest)
	switch {
	case summary.Critical > 0:
		summary.Health = "critico"
	case summary.High > 0:
		summary.Health = "alto_riesgo"
	case summary.Medium > 0:
		summary.Health = "riesgo_medio"
	case summary.TotalFindings > 0:
		summary.Health = "bajo_riesgo"
	default:
		summary.Health = "estable"
	}
	report.Summary = summary
	ApplyCoverage(report)
	if strings.TrimSpace(report.GeneratedAt) == "" {
		report.GeneratedAt = time.Now().UTC().Format(time.RFC3339)
	}
}

func ApplyCoverage(report *ScanReport) {
	if report == nil {
		return
	}
	required := make([]string, 0, len(report.Config.EnabledTools))
	seenRequired := make(map[string]struct{}, len(report.Config.EnabledTools))
	for _, name := range report.Config.EnabledTools {
		name = canonicalToolName(name)
		if name == "" {
			continue
		}
		if _, exists := seenRequired[name]; exists {
			continue
		}
		seenRequired[name] = struct{}{}
		required = append(required, name)
	}
	if len(required) == 0 {
		report.Summary.CoverageComplete = false
		report.Summary.CoverageStatus = "desconocida"
		report.Summary.IncompleteTools = nil
		return
	}
	results := make(map[string]ToolResult, len(report.Tools))
	for _, tool := range report.Tools {
		results[canonicalToolName(tool.Name)] = tool
	}
	incomplete := make([]string, 0)
	for _, name := range required {
		tool, exists := results[name]
		if !exists || !tool.Available || !tool.Executed || strings.ToLower(strings.TrimSpace(tool.Status)) != "ok" {
			incomplete = append(incomplete, name)
		}
	}
	sort.Strings(incomplete)
	report.Summary.IncompleteTools = incomplete
	report.Summary.CoverageComplete = len(incomplete) == 0
	if report.Summary.CoverageComplete {
		report.Summary.CoverageStatus = "completa"
	} else {
		report.Summary.CoverageStatus = "incompleta"
	}
}

func canonicalToolName(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "trivy", "vulnerability_scan":
		return "vulnerability_scan"
	case "lynis":
		return "lynis"
	case "nmap":
		return "nmap"
	default:
		return strings.ToLower(strings.TrimSpace(name))
	}
}

func Compare(current, previous *ScanReport) Comparison {
	if current == nil || previous == nil {
		return Comparison{}
	}
	comparison := Comparison{
		PreviousScanID:      strings.TrimSpace(previous.ScanID),
		PreviousGeneratedAt: strings.TrimSpace(previous.GeneratedAt),
		SeverityDelta:       map[string]int{},
	}
	currentKeys := make(map[string]Finding, len(current.Findings))
	for _, finding := range NormalizeFindings(current.Findings) {
		currentKeys[finding.ID] = finding
	}
	previousKeys := make(map[string]Finding, len(previous.Findings))
	for _, finding := range NormalizeFindings(previous.Findings) {
		previousKeys[finding.ID] = finding
	}
	for key, finding := range currentKeys {
		if _, exists := previousKeys[key]; !exists {
			comparison.NewFindings++
			comparison.SeverityDelta[strings.ToLower(string(normalizeSeverity(finding.Severity)))]++
		}
	}
	for key, finding := range previousKeys {
		if _, exists := currentKeys[key]; !exists {
			comparison.ResolvedFindings++
			comparison.SeverityDelta[strings.ToLower(string(normalizeSeverity(finding.Severity)))]--
		}
	}
	currentPorts := make(map[int]struct{}, len(current.Summary.OpenPorts))
	for _, port := range current.Summary.OpenPorts {
		currentPorts[port] = struct{}{}
	}
	previousPorts := make(map[int]struct{}, len(previous.Summary.OpenPorts))
	for _, port := range previous.Summary.OpenPorts {
		previousPorts[port] = struct{}{}
	}
	for port := range currentPorts {
		if _, exists := previousPorts[port]; !exists {
			comparison.NewOpenPorts = append(comparison.NewOpenPorts, port)
		}
	}
	for port := range previousPorts {
		if _, exists := currentPorts[port]; !exists {
			comparison.ClosedPorts = append(comparison.ClosedPorts, port)
		}
	}
	sort.Ints(comparison.NewOpenPorts)
	sort.Ints(comparison.ClosedPorts)
	summaryBits := make([]string, 0, 3)
	if comparison.NewFindings > 0 {
		summaryBits = append(summaryBits, fmt.Sprintf("%d hallazgos nuevos", comparison.NewFindings))
	}
	if comparison.ResolvedFindings > 0 {
		summaryBits = append(summaryBits, fmt.Sprintf("%d hallazgos resueltos", comparison.ResolvedFindings))
	}
	if len(comparison.NewOpenPorts) > 0 || len(comparison.ClosedPorts) > 0 {
		summaryBits = append(summaryBits, fmt.Sprintf("puertos nuevos=%d cerrados=%d", len(comparison.NewOpenPorts), len(comparison.ClosedPorts)))
	}
	comparison.Summary = strings.Join(summaryBits, "; ")
	return comparison
}

func HistoryFromReport(report *ScanReport) HistoryEntry {
	if report == nil {
		return HistoryEntry{}
	}
	return HistoryEntry{
		ScanID:           report.ScanID,
		GeneratedAt:      report.GeneratedAt,
		Scope:            report.Scope,
		TargetHost:       report.TargetHost,
		Profile:          report.Profile,
		Status:           report.Status,
		CoverageStatus:   report.Summary.CoverageStatus,
		CoverageComplete: report.Summary.CoverageComplete,
		TotalFindings:    report.Summary.TotalFindings,
		HighestSeverity:  report.Summary.HighestSeverity,
		NewFindings:      report.Comparison.NewFindings,
		ResolvedFindings: report.Comparison.ResolvedFindings,
		Reports:          report.Reports,
	}
}

func GenerateArtifacts(report ScanReport) (map[string][]byte, error) {
	ApplySummary(&report)
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, err
	}
	artifacts := map[string][]byte{
		"json": jsonData,
		"txt":  BuildText(report),
		"csv":  BuildCSV(report),
		"html": BuildHTML(report),
		"xls":  BuildExcel(report),
	}
	pdfData, err := BuildPDF(report)
	if err != nil {
		return nil, err
	}
	artifacts["pdf"] = pdfData
	return artifacts, nil
}

func BuildText(report ScanReport) []byte {
	ApplySummary(&report)
	lines := textLines(report)
	return []byte(strings.Join(lines, "\n") + "\n")
}

func textLines(report ScanReport) []string {
	lines := []string{
		"REPORTE DE SEGURIDAD VPS",
		"Scan ID: " + safeText(report.ScanID),
		"Generado: " + safeText(report.GeneratedAt),
		"Alcance: " + safeText(report.Scope),
		"Objetivo: " + safeText(report.TargetHost),
		"Perfil: " + safeText(report.Profile),
		"Estado: " + safeText(report.Status),
		"Cobertura: " + safeText(report.Summary.CoverageStatus),
		fmt.Sprintf("Resumen: critico=%d alto=%d medio=%d bajo=%d info=%d total=%d", report.Summary.Critical, report.Summary.High, report.Summary.Medium, report.Summary.Low, report.Summary.Info, report.Summary.TotalFindings),
	}
	if len(report.Summary.IncompleteTools) > 0 {
		lines = append(lines, "Herramientas incompletas: "+strings.Join(report.Summary.IncompleteTools, ", "))
	}
	if report.Summary.HardeningIndex > 0 {
		lines = append(lines, fmt.Sprintf("Hardening index: %d", report.Summary.HardeningIndex))
	}
	if len(report.Summary.OpenPorts) > 0 {
		portText := make([]string, 0, len(report.Summary.OpenPorts))
		for _, port := range report.Summary.OpenPorts {
			portText = append(portText, strconv.Itoa(port))
		}
		lines = append(lines, "Puertos abiertos: "+strings.Join(portText, ", "))
	}
	if report.SystemInfo.Firewall != "" {
		lines = append(lines, "Firewall: "+safeText(report.SystemInfo.Firewall))
	}
	if len(report.Notes) > 0 {
		lines = append(lines, "", "Notas:")
		for _, note := range report.Notes {
			lines = append(lines, "- "+safeText(note))
		}
	}
	lines = append(lines, "", "Herramientas:")
	for _, tool := range report.Tools {
		lines = append(lines, fmt.Sprintf("- %s: %s (%s)", safeText(tool.DisplayName), safeText(tool.Status), safeText(tool.Summary)))
	}
	lines = append(lines, "", "Hallazgos:")
	if len(report.Findings) == 0 {
		lines = append(lines, "- Sin hallazgos reportados")
	} else {
		for idx, finding := range report.Findings {
			lines = append(lines, fmt.Sprintf("%d. [%s] %s - %s", idx+1, finding.Severity, safeText(finding.Tool), safeText(finding.Title)))
			if finding.Target != "" || finding.Port > 0 || finding.Service != "" {
				context := make([]string, 0, 3)
				if finding.Target != "" {
					context = append(context, "target="+safeText(finding.Target))
				}
				if finding.Port > 0 {
					context = append(context, "port="+strconv.Itoa(finding.Port))
				}
				if finding.Service != "" {
					context = append(context, "service="+safeText(finding.Service))
				}
				lines = append(lines, "   Contexto: "+strings.Join(context, " | "))
			}
			if finding.Reference != "" {
				lines = append(lines, "   Referencia: "+safeText(finding.Reference))
			}
			if finding.Description != "" {
				lines = append(lines, "   Descripcion: "+safeText(finding.Description))
			}
			if finding.Evidence != "" {
				lines = append(lines, "   Evidencia: "+safeText(finding.Evidence))
			}
			if finding.Recommendation != "" {
				lines = append(lines, "   Accion: "+safeText(finding.Recommendation))
			}
		}
	}
	if report.Comparison.PreviousScanID != "" {
		lines = append(lines, "", "Comparacion:")
		lines = append(lines, fmt.Sprintf("- Scan previo: %s (%s)", safeText(report.Comparison.PreviousScanID), safeText(report.Comparison.PreviousGeneratedAt)))
		lines = append(lines, fmt.Sprintf("- Nuevos: %d | Resueltos: %d", report.Comparison.NewFindings, report.Comparison.ResolvedFindings))
		if report.Comparison.Summary != "" {
			lines = append(lines, "- Resumen: "+safeText(report.Comparison.Summary))
		}
	}
	return lines
}

func BuildCSV(report ScanReport) []byte {
	ApplySummary(&report)
	buffer := &bytes.Buffer{}
	writer := csv.NewWriter(buffer)
	_ = writer.Write([]string{"scan_id", "generated_at", "scope", "target_host", "profile", "coverage_status", "coverage_complete", "severity", "tool", "category", "title", "target", "port", "service", "reference", "recommendation", "evidence"})
	if len(report.Findings) == 0 {
		_ = writer.Write([]string{
			report.ScanID,
			report.GeneratedAt,
			report.Scope,
			report.TargetHost,
			report.Profile,
			report.Summary.CoverageStatus,
			strconv.FormatBool(report.Summary.CoverageComplete),
			"", "", "", "", "", "", "", "", "", "",
		})
	}
	for _, finding := range report.Findings {
		_ = writer.Write([]string{
			report.ScanID,
			report.GeneratedAt,
			report.Scope,
			report.TargetHost,
			report.Profile,
			report.Summary.CoverageStatus,
			strconv.FormatBool(report.Summary.CoverageComplete),
			string(finding.Severity),
			finding.Tool,
			finding.Category,
			finding.Title,
			finding.Target,
			strconv.Itoa(finding.Port),
			finding.Service,
			finding.Reference,
			finding.Recommendation,
			finding.Evidence,
		})
	}
	writer.Flush()
	return buffer.Bytes()
}

func BuildHTML(report ScanReport) []byte {
	ApplySummary(&report)
	buffer := &bytes.Buffer{}
	buffer.WriteString("<!doctype html><html lang=\"es\"><head><meta charset=\"utf-8\"><title>Reporte Seguridad VPS</title><style>")
	buffer.WriteString("body{font-family:Segoe UI,Arial,sans-serif;margin:24px;color:#14213d;background:#f7f4ea}h1,h2{margin:0 0 12px}section{background:#fff;border:1px solid #d6d0c3;border-radius:14px;padding:18px;margin-bottom:18px}.summary{display:grid;grid-template-columns:repeat(auto-fit,minmax(140px,1fr));gap:12px}.metric{padding:12px;border-radius:12px;background:#f1ede1}.badge{display:inline-block;padding:4px 8px;border-radius:999px;font-weight:700}.critico{background:#8c1c13;color:#fff}.alto{background:#c4502d;color:#fff}.medio{background:#f4a259;color:#1a1a1a}.bajo{background:#72b01d;color:#fff}.info{background:#3d5a80;color:#fff}table{width:100%;border-collapse:collapse}th,td{padding:10px;border-bottom:1px solid #ece7db;text-align:left;vertical-align:top}th{background:#f7f4ea}small{color:#5f6b7a}</style></head><body>")
	buffer.WriteString("<section><h1>Reporte de seguridad VPS</h1>")
	buffer.WriteString("<p><strong>Scan ID:</strong> " + htmlEscape(report.ScanID) + "<br>")
	buffer.WriteString("<strong>Generado:</strong> " + htmlEscape(report.GeneratedAt) + "<br>")
	buffer.WriteString("<strong>Alcance:</strong> " + htmlEscape(report.Scope) + "<br>")
	buffer.WriteString("<strong>Objetivo:</strong> " + htmlEscape(report.TargetHost) + "<br>")
	buffer.WriteString("<strong>Perfil:</strong> " + htmlEscape(report.Profile) + "</p></section>")
	buffer.WriteString("<section><h2>Resumen</h2><div class=\"summary\">")
	buffer.WriteString(summaryMetric("Critico", strconv.Itoa(report.Summary.Critical), "critico"))
	buffer.WriteString(summaryMetric("Alto", strconv.Itoa(report.Summary.High), "alto"))
	buffer.WriteString(summaryMetric("Medio", strconv.Itoa(report.Summary.Medium), "medio"))
	buffer.WriteString(summaryMetric("Bajo", strconv.Itoa(report.Summary.Low), "bajo"))
	buffer.WriteString(summaryMetric("Info", strconv.Itoa(report.Summary.Info), "info"))
	buffer.WriteString(summaryMetric("Total", strconv.Itoa(report.Summary.TotalFindings), "info"))
	buffer.WriteString(summaryMetric("Cobertura", htmlEscape(report.Summary.CoverageStatus), "info"))
	if report.Summary.HardeningIndex > 0 {
		buffer.WriteString(summaryMetric("Hardening", strconv.Itoa(report.Summary.HardeningIndex), "info"))
	}
	buffer.WriteString(summaryMetric("Salud", htmlEscape(report.Summary.Health), "info"))
	buffer.WriteString("</div></section>")
	buffer.WriteString("<section><h2>Herramientas</h2><table><thead><tr><th>Herramienta</th><th>Estado</th><th>Resumen</th><th>Error</th></tr></thead><tbody>")
	for _, tool := range report.Tools {
		buffer.WriteString("<tr><td>" + htmlEscape(tool.DisplayName) + "</td><td>" + htmlEscape(tool.Status) + "</td><td>" + htmlEscape(tool.Summary) + "</td><td><small>" + htmlEscape(tool.Error) + "</small></td></tr>")
	}
	buffer.WriteString("</tbody></table></section>")
	buffer.WriteString("<section><h2>Hallazgos</h2><table><thead><tr><th>Severidad</th><th>Herramienta</th><th>Titulo</th><th>Contexto</th><th>Accion</th></tr></thead><tbody>")
	if len(report.Findings) == 0 {
		buffer.WriteString("<tr><td colspan=\"5\">Sin hallazgos</td></tr>")
	}
	for _, finding := range report.Findings {
		context := make([]string, 0, 3)
		if finding.Target != "" {
			context = append(context, "Objetivo: "+finding.Target)
		}
		if finding.Port > 0 {
			context = append(context, "Puerto: "+strconv.Itoa(finding.Port))
		}
		if finding.Service != "" {
			context = append(context, "Servicio: "+finding.Service)
		}
		if finding.Reference != "" {
			context = append(context, "Ref: "+finding.Reference)
		}
		buffer.WriteString("<tr><td><span class=\"badge " + strings.ToLower(string(normalizeSeverity(finding.Severity))) + "\">" + htmlEscape(string(finding.Severity)) + "</span></td><td>" + htmlEscape(finding.Tool) + "</td><td><strong>" + htmlEscape(finding.Title) + "</strong><br><small>" + htmlEscape(finding.Description) + "</small></td><td>" + htmlEscape(strings.Join(context, " | ")) + "<br><small>" + htmlEscape(finding.Evidence) + "</small></td><td>" + htmlEscape(finding.Recommendation) + "</td></tr>")
	}
	buffer.WriteString("</tbody></table></section>")
	if report.Comparison.PreviousScanID != "" {
		buffer.WriteString("<section><h2>Comparacion</h2>")
		buffer.WriteString("<p><strong>Previo:</strong> " + htmlEscape(report.Comparison.PreviousScanID) + "<br>")
		buffer.WriteString("<strong>Nuevos:</strong> " + strconv.Itoa(report.Comparison.NewFindings) + "<br>")
		buffer.WriteString("<strong>Resueltos:</strong> " + strconv.Itoa(report.Comparison.ResolvedFindings) + "<br>")
		buffer.WriteString("<strong>Resumen:</strong> " + htmlEscape(report.Comparison.Summary) + "</p></section>")
	}
	buffer.WriteString("</body></html>")
	return buffer.Bytes()
}

func summaryMetric(label, value, className string) string {
	return "<div class=\"metric\"><strong>" + htmlEscape(label) + "</strong><br><span class=\"badge " + htmlEscape(className) + "\">" + htmlEscape(value) + "</span></div>"
}

func BuildExcel(report ScanReport) []byte {
	ApplySummary(&report)
	buffer := &bytes.Buffer{}
	buffer.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	buffer.WriteString("<?mso-application progid=\"Excel.Sheet\"?>")
	buffer.WriteString("<Workbook xmlns=\"urn:schemas-microsoft-com:office:spreadsheet\" xmlns:ss=\"urn:schemas-microsoft-com:office:spreadsheet\">")
	buffer.WriteString("<Worksheet ss:Name=\"Resumen\"><Table>")
	writeXMLRow(buffer, []string{"Campo", "Valor"})
	writeXMLRow(buffer, []string{"Scan ID", report.ScanID})
	writeXMLRow(buffer, []string{"Generado", report.GeneratedAt})
	writeXMLRow(buffer, []string{"Alcance", report.Scope})
	writeXMLRow(buffer, []string{"Objetivo", report.TargetHost})
	writeXMLRow(buffer, []string{"Perfil", report.Profile})
	writeXMLRow(buffer, []string{"Estado", report.Status})
	writeXMLRow(buffer, []string{"Cobertura", report.Summary.CoverageStatus})
	writeXMLRow(buffer, []string{"Cobertura completa", strconv.FormatBool(report.Summary.CoverageComplete)})
	writeXMLRow(buffer, []string{"Herramientas incompletas", strings.Join(report.Summary.IncompleteTools, ", ")})
	writeXMLRow(buffer, []string{"Critico", strconv.Itoa(report.Summary.Critical)})
	writeXMLRow(buffer, []string{"Alto", strconv.Itoa(report.Summary.High)})
	writeXMLRow(buffer, []string{"Medio", strconv.Itoa(report.Summary.Medium)})
	writeXMLRow(buffer, []string{"Bajo", strconv.Itoa(report.Summary.Low)})
	writeXMLRow(buffer, []string{"Info", strconv.Itoa(report.Summary.Info)})
	writeXMLRow(buffer, []string{"Total", strconv.Itoa(report.Summary.TotalFindings)})
	buffer.WriteString("</Table></Worksheet>")
	buffer.WriteString("<Worksheet ss:Name=\"Hallazgos\"><Table>")
	writeXMLRow(buffer, []string{"Severidad", "Herramienta", "Categoria", "Titulo", "Objetivo", "Puerto", "Servicio", "Referencia", "Accion", "Evidencia"})
	for _, finding := range report.Findings {
		writeXMLRow(buffer, []string{
			string(finding.Severity),
			finding.Tool,
			finding.Category,
			finding.Title,
			finding.Target,
			strconv.Itoa(finding.Port),
			finding.Service,
			finding.Reference,
			finding.Recommendation,
			finding.Evidence,
		})
	}
	buffer.WriteString("</Table></Worksheet></Workbook>")
	return buffer.Bytes()
}

func writeXMLRow(buffer *bytes.Buffer, cells []string) {
	buffer.WriteString("<Row>")
	for _, cell := range cells {
		buffer.WriteString("<Cell><Data ss:Type=\"String\">" + xmlEscape(cell) + "</Data></Cell>")
	}
	buffer.WriteString("</Row>")
}

func BuildPDF(report ScanReport) ([]byte, error) {
	ApplySummary(&report)
	lines := wrapLines(textLines(report), 92)
	if len(lines) == 0 {
		lines = []string{"Sin datos"}
	}
	pages := chunkLines(lines, 44)
	objects := make([]string, 0, 3+len(pages)*2)
	objects = append(objects, "<< /Type /Catalog /Pages 2 0 R >>")
	pageRefs := make([]string, 0, len(pages))
	for idx := range pages {
		pageObjNumber := 5 + idx*2
		pageRefs = append(pageRefs, fmt.Sprintf("%d 0 R", pageObjNumber))
	}
	objects = append(objects, "<< /Type /Pages /Kids ["+strings.Join(pageRefs, " ")+"] /Count "+strconv.Itoa(len(pageRefs))+" >>")
	objects = append(objects, "<< /Type /Font /Subtype /Type1 /BaseFont /Courier >>")
	for idx, pageLines := range pages {
		contentObj := buildPDFContent(pageLines)
		objects = append(objects, fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(contentObj), contentObj))
		contentObjNumber := 4 + idx*2
		objects = append(objects, fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 3 0 R >> >> /Contents %d 0 R >>", contentObjNumber))
	}
	buffer := &bytes.Buffer{}
	buffer.WriteString("%PDF-1.4\n")
	offsets := make([]int, 0, len(objects)+1)
	offsets = append(offsets, 0)
	for idx, obj := range objects {
		offsets = append(offsets, buffer.Len())
		buffer.WriteString(strconv.Itoa(idx+1) + " 0 obj\n")
		buffer.WriteString(obj)
		buffer.WriteString("\nendobj\n")
	}
	xrefPos := buffer.Len()
	buffer.WriteString("xref\n")
	buffer.WriteString("0 " + strconv.Itoa(len(objects)+1) + "\n")
	buffer.WriteString("0000000000 65535 f \n")
	for idx := 1; idx < len(offsets); idx++ {
		buffer.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[idx]))
	}
	buffer.WriteString("trailer\n<< /Size " + strconv.Itoa(len(objects)+1) + " /Root 1 0 R >>\n")
	buffer.WriteString("startxref\n")
	buffer.WriteString(strconv.Itoa(xrefPos) + "\n%%EOF")
	return buffer.Bytes(), nil
}

func buildPDFContent(lines []string) string {
	builder := &strings.Builder{}
	builder.WriteString("BT\n/F1 10 Tf\n36 760 Td\n12 TL\n")
	for idx, line := range lines {
		escaped := pdfEscape(line)
		if idx == 0 {
			builder.WriteString("(" + escaped + ") Tj\n")
		} else {
			builder.WriteString("T*\n(" + escaped + ") Tj\n")
		}
	}
	builder.WriteString("ET")
	return builder.String()
}

func chunkLines(lines []string, size int) [][]string {
	if size <= 0 {
		return [][]string{lines}
	}
	chunks := make([][]string, 0, (len(lines)/size)+1)
	for start := 0; start < len(lines); start += size {
		end := start + size
		if end > len(lines) {
			end = len(lines)
		}
		chunks = append(chunks, lines[start:end])
	}
	return chunks
}

func wrapLines(lines []string, width int) []string {
	wrapped := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimRight(line, "\r\n")
		if len(line) <= width {
			wrapped = append(wrapped, line)
			continue
		}
		current := line
		for len(current) > width {
			split := strings.LastIndex(current[:width], " ")
			if split <= 0 {
				split = width
			}
			wrapped = append(wrapped, strings.TrimSpace(current[:split]))
			current = strings.TrimSpace(current[split:])
		}
		if current != "" {
			wrapped = append(wrapped, current)
		}
	}
	return wrapped
}

func pdfEscape(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "(", "\\(", ")", "\\)")
	return replacer.Replace(value)
}

func htmlEscape(value string) string {
	return xmlEscape(value)
}

func xmlEscape(value string) string {
	var buffer bytes.Buffer
	_ = xml.EscapeText(&buffer, []byte(value))
	return buffer.String()
}

func safeText(value string) string {
	return strings.TrimSpace(value)
}
