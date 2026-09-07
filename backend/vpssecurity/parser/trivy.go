package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/you/pos-backend/vpssecurity/reports"
)

type trivyDocument struct {
	Results []trivyResult `json:"Results"`
}

type trivyResult struct {
	Target            string                 `json:"Target"`
	Class             string                 `json:"Class"`
	Type              string                 `json:"Type"`
	Vulnerabilities   []trivyVulnerability   `json:"Vulnerabilities"`
	Misconfigurations []trivyMisconfiguration `json:"Misconfigurations"`
}

type trivyVulnerability struct {
	VulnerabilityID string `json:"VulnerabilityID"`
	PkgName         string `json:"PkgName"`
	InstalledVersion string `json:"InstalledVersion"`
	FixedVersion    string `json:"FixedVersion"`
	Title           string `json:"Title"`
	Severity        string `json:"Severity"`
	PrimaryURL      string `json:"PrimaryURL"`
	Description     string `json:"Description"`
}

type trivyMisconfiguration struct {
	ID          string `json:"ID"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
	Severity    string `json:"Severity"`
	Resolution  string `json:"Resolution"`
}

func ParseTrivyJSON(data []byte, fallbackTarget string) ([]reports.Finding, string, error) {
	var doc trivyDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, "", err
	}
	findings := make([]reports.Finding, 0)
	vulnCount := 0
	misconfigCount := 0
	for _, result := range doc.Results {
		target := strings.TrimSpace(result.Target)
		if target == "" {
			target = fallbackTarget
		}
		for _, vulnerability := range result.Vulnerabilities {
			vulnCount++
			title := strings.TrimSpace(vulnerability.Title)
			if title == "" {
				title = fmt.Sprintf("%s en paquete %s", vulnerability.VulnerabilityID, vulnerability.PkgName)
			}
			recommendation := "Actualice o reemplace el paquete afectado y vuelva a ejecutar el escaneo."
			if strings.TrimSpace(vulnerability.FixedVersion) != "" {
				recommendation = fmt.Sprintf("Actualice %s desde %s hacia %s o superior.", vulnerability.PkgName, vulnerability.InstalledVersion, vulnerability.FixedVersion)
			}
			findings = append(findings, reports.Finding{
				Tool:           "trivy",
				Category:       "vulnerabilidades",
				Severity:       reports.Severity(strings.ToUpper(strings.TrimSpace(vulnerability.Severity))),
				Title:          title,
				Description:    strings.TrimSpace(vulnerability.Description),
				Recommendation: recommendation,
				Target:         target,
				Reference:      strings.TrimSpace(vulnerability.VulnerabilityID),
				Evidence:       strings.TrimSpace(strings.Join([]string{vulnerability.PkgName, vulnerability.InstalledVersion, vulnerability.PrimaryURL}, " | ")),
			})
		}
		for _, misconfiguration := range result.Misconfigurations {
			misconfigCount++
			findings = append(findings, reports.Finding{
				Tool:           "trivy",
				Category:       "misconfiguracion",
				Severity:       reports.Severity(strings.ToUpper(strings.TrimSpace(misconfiguration.Severity))),
				Title:          strings.TrimSpace(misconfiguration.Title),
				Description:    strings.TrimSpace(misconfiguration.Description),
				Recommendation: defaultString(strings.TrimSpace(misconfiguration.Resolution), "Aplique la correccion sugerida por Trivy y valide el servicio afectado."),
				Target:         target,
				Reference:      strings.TrimSpace(misconfiguration.ID),
				Evidence:       strings.TrimSpace(result.Type),
			})
		}
	}
	summary := fmt.Sprintf("%d vulnerabilidades y %d misconfiguraciones detectadas", vulnCount, misconfigCount)
	return findings, summary, nil
}