package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/you/pos-backend/vpssecurity/reports"
)

func ParseLynisReport(data []byte, target string) ([]reports.Finding, int, string) {
	findings := make([]reports.Finding, 0)
	hardeningIndex := 0
	warnings := 0
	suggestions := 0
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		switch {
		case strings.HasPrefix(line, "hardening_index="):
			value := strings.TrimSpace(strings.TrimPrefix(line, "hardening_index="))
			if parsed, err := strconv.Atoi(value); err == nil {
				hardeningIndex = parsed
			}
		case strings.HasPrefix(line, "warning[]="):
			warnings++
			value := strings.TrimSpace(strings.TrimPrefix(line, "warning[]="))
			reference, title := splitLynisItem(value)
			findings = append(findings, reports.Finding{
				Tool:           "lynis",
				Category:       detectLynisCategory(title),
				Severity:       detectLynisSeverity(title, true),
				Title:          title,
				Description:    "Lynis reporto una advertencia operativa sobre el host auditado.",
				Recommendation: lynisRecommendation(title),
				Target:         target,
				Reference:      reference,
				Evidence:       value,
			})
		case strings.HasPrefix(line, "suggestion[]="):
			suggestions++
			value := strings.TrimSpace(strings.TrimPrefix(line, "suggestion[]="))
			reference, title := splitLynisItem(value)
			findings = append(findings, reports.Finding{
				Tool:           "lynis",
				Category:       detectLynisCategory(title),
				Severity:       detectLynisSeverity(title, false),
				Title:          title,
				Description:    "Lynis detecto una mejora recomendada para endurecer el servidor.",
				Recommendation: lynisRecommendation(title),
				Target:         target,
				Reference:      reference,
				Evidence:       value,
			})
		}
	}
	summary := fmt.Sprintf("%d advertencias, %d sugerencias, hardening index %d", warnings, suggestions, hardeningIndex)
	return findings, hardeningIndex, summary
}

func splitLynisItem(value string) (string, string) {
	parts := strings.Split(value, "|")
	if len(parts) == 0 {
		return "", strings.TrimSpace(value)
	}
	reference := strings.TrimSpace(parts[0])
	title := strings.TrimSpace(value)
	if len(parts) > 1 {
		title = strings.TrimSpace(parts[1])
	}
	if title == "" {
		title = strings.TrimSpace(value)
	}
	return reference, title
}

func detectLynisCategory(title string) string {
	lower := strings.ToLower(title)
	switch {
	case strings.Contains(lower, "ssh"):
		return "ssh"
	case strings.Contains(lower, "firewall"):
		return "firewall"
	case strings.Contains(lower, "nginx"), strings.Contains(lower, "apache"), strings.Contains(lower, "http"):
		return "web"
	case strings.Contains(lower, "password"), strings.Contains(lower, "auth"):
		return "autenticacion"
	case strings.Contains(lower, "kernel"), strings.Contains(lower, "boot"):
		return "sistema"
	case strings.Contains(lower, "file"), strings.Contains(lower, "permission"):
		return "permisos"
	default:
		return "hardening"
	}
}

func detectLynisSeverity(title string, warning bool) reports.Severity {
	lower := strings.ToLower(title)
	if strings.Contains(lower, "critical") || strings.Contains(lower, "vulnerab") {
		return reports.SeverityCritical
	}
	if strings.Contains(lower, "firewall") || strings.Contains(lower, "root login") || strings.Contains(lower, "password") {
		if warning {
			return reports.SeverityHigh
		}
		return reports.SeverityMedium
	}
	if warning {
		return reports.SeverityHigh
	}
	return reports.SeverityMedium
}

func lynisRecommendation(title string) string {
	lower := strings.ToLower(title)
	switch {
	case strings.Contains(lower, "firewall"):
		return "Revise UFW o nftables y asegure una politica por defecto de entrada en deny con reglas minimas necesarias."
	case strings.Contains(lower, "ssh"):
		return "Ajuste /etc/ssh/sshd_config, limite autenticacion por password y deshabilite root login directo si no es imprescindible."
	case strings.Contains(lower, "password"):
		return "Refuerce la politica PAM y la complejidad/rotacion de credenciales administrativas."
	case strings.Contains(lower, "nginx"), strings.Contains(lower, "http"):
		return "Endurezca la configuracion TLS y cabeceras del servidor web, y elimine modulos o sitios no utilizados."
	default:
		return "Aplique la recomendacion de hardening reportada por Lynis y documente el cambio en el VPS."
	}
}
