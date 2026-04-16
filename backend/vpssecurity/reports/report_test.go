package reports

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGenerateArtifactsIncludesAllFormats(t *testing.T) {
	report := ScanReport{
		ScanID:      "scan-1",
		GeneratedAt: "2026-04-16T12:00:00Z",
		TargetHost:  "127.0.0.1",
		Profile:     "full",
		Status:      "completed",
		Findings: []Finding{
			{Tool: "nmap", Category: "puertos", Severity: SeverityHigh, Title: "Puerto 5432 expuesto", Target: "127.0.0.1", Port: 5432, Recommendation: "Restringir acceso"},
			{Tool: "custom", Category: "firewall", Severity: SeverityMedium, Title: "UFW inactivo", Recommendation: "Activar UFW"},
		},
		Tools: []ToolResult{{Name: "nmap", DisplayName: "Nmap", Status: "ok", Summary: "1 puerto abierto"}},
	}
	artifacts, err := GenerateArtifacts(report)
	if err != nil {
		t.Fatalf("generate artifacts: %v", err)
	}
	formats := []string{"json", "txt", "html", "csv", "pdf", "xls"}
	for _, format := range formats {
		content := artifacts[format]
		if len(content) == 0 {
			t.Fatalf("format %s not generated", format)
		}
	}
	var decoded ScanReport
	if err := json.Unmarshal(artifacts["json"], &decoded); err != nil {
		t.Fatalf("decode json artifact: %v", err)
	}
	if decoded.Summary.TotalFindings != 2 {
		t.Fatalf("expected total findings 2, got %d", decoded.Summary.TotalFindings)
	}
	if !strings.Contains(string(artifacts["txt"]), "Puerto 5432 expuesto") {
		t.Fatalf("txt artifact missing finding")
	}
	if !strings.Contains(string(artifacts["html"]), "Reporte de seguridad VPS") {
		t.Fatalf("html artifact missing title")
	}
}

func TestCompareDetectsNewResolvedAndPortChanges(t *testing.T) {
	previous := &ScanReport{
		ScanID:      "scan-prev",
		GeneratedAt: "2026-04-15T12:00:00Z",
		Summary:     Summary{OpenPorts: []int{22, 80}},
		Findings: []Finding{
			{ID: "a", Tool: "nmap", Severity: SeverityHigh, Title: "Puerto 22"},
			{ID: "b", Tool: "custom", Severity: SeverityMedium, Title: "UFW inactivo"},
		},
	}
	current := &ScanReport{
		ScanID:      "scan-current",
		GeneratedAt: "2026-04-16T12:00:00Z",
		Summary:     Summary{OpenPorts: []int{22, 443}},
		Findings: []Finding{
			{ID: "a", Tool: "nmap", Severity: SeverityHigh, Title: "Puerto 22"},
			{ID: "c", Tool: "nmap", Severity: SeverityHigh, Title: "Puerto 443"},
		},
	}
	comparison := Compare(current, previous)
	if comparison.NewFindings != 1 {
		t.Fatalf("expected 1 new finding, got %d", comparison.NewFindings)
	}
	if comparison.ResolvedFindings != 1 {
		t.Fatalf("expected 1 resolved finding, got %d", comparison.ResolvedFindings)
	}
	if len(comparison.NewOpenPorts) != 1 || comparison.NewOpenPorts[0] != 443 {
		t.Fatalf("expected new open port 443, got %+v", comparison.NewOpenPorts)
	}
	if len(comparison.ClosedPorts) != 1 || comparison.ClosedPorts[0] != 80 {
		t.Fatalf("expected closed port 80, got %+v", comparison.ClosedPorts)
	}
}