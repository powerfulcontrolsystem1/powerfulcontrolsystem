package parser

import (
	"encoding/xml"
	"fmt"
	"net"
	"strings"

	"github.com/you/pos-backend/vpssecurity/reports"
)

type nmapRun struct {
	Hosts []nmapHost `xml:"host"`
}

type nmapHost struct {
	Addresses []nmapAddress `xml:"address"`
	Ports     []nmapPort    `xml:"ports>port"`
}

type nmapAddress struct {
	Addr string `xml:"addr,attr"`
}

type nmapPort struct {
	PortID  int         `xml:"portid,attr"`
	State   nmapState   `xml:"state"`
	Service nmapService `xml:"service"`
}

type nmapState struct {
	State string `xml:"state,attr"`
}

type nmapService struct {
	Name      string `xml:"name,attr"`
	Product   string `xml:"product,attr"`
	Version   string `xml:"version,attr"`
	ExtraInfo string `xml:"extrainfo,attr"`
}

func ParseNmapXML(data []byte, fallbackTarget string) ([]reports.Finding, []int, string, error) {
	var doc nmapRun
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, nil, "", err
	}
	findings := make([]reports.Finding, 0)
	openPorts := make([]int, 0)
	for _, host := range doc.Hosts {
		target := fallbackTarget
		if len(host.Addresses) > 0 && strings.TrimSpace(host.Addresses[0].Addr) != "" {
			target = strings.TrimSpace(host.Addresses[0].Addr)
		}
		for _, port := range host.Ports {
			if strings.ToLower(strings.TrimSpace(port.State.State)) != "open" {
				continue
			}
			openPorts = append(openPorts, port.PortID)
			severity, recommendation := nmapSeverityForPort(port.PortID, port.Service.Name)
			serviceLabel := strings.TrimSpace(strings.Join([]string{port.Service.Name, port.Service.Product, port.Service.Version}, " "))
			serviceLabel = strings.Join(strings.Fields(serviceLabel), " ")
			if serviceLabel == "" {
				serviceLabel = strings.TrimSpace(port.Service.Name)
			}
			title := fmt.Sprintf("Puerto %d expuesto por %s", port.PortID, defaultString(serviceLabel, "servicio no identificado"))
			description := "Nmap detecto el puerto como abierto y accesible en el host auditado."
			if isLoopbackTarget(target) {
				severity = reports.SeverityInfo
				title = fmt.Sprintf("Puerto %d interno en loopback por %s", port.PortID, defaultString(serviceLabel, "servicio no identificado"))
				description = "Nmap detecto el puerto dentro del loopback del runtime auditado; este resultado no demuestra exposicion publica."
				recommendation = "Confirme que el servicio sea esperado y que Docker o el host no publiquen este puerto en una interfaz externa."
			}
			findings = append(findings, reports.Finding{
				Tool:           "nmap",
				Category:       "puertos",
				Severity:       severity,
				Title:          title,
				Description:    description,
				Recommendation: recommendation,
				Target:         target,
				Port:           port.PortID,
				Service:        serviceLabel,
				Evidence:       strings.TrimSpace(strings.Join([]string{port.Service.Product, port.Service.Version, port.Service.ExtraInfo}, " ")),
			})
		}
	}
	summary := fmt.Sprintf("%d puertos abiertos detectados", len(openPorts))
	return findings, openPorts, summary, nil
}

func isLoopbackTarget(target string) bool {
	target = strings.TrimSpace(strings.ToLower(target))
	if target == "localhost" {
		return true
	}
	ip := net.ParseIP(target)
	return ip != nil && ip.IsLoopback()
}

func nmapSeverityForPort(port int, service string) (reports.Severity, string) {
	switch port {
	case 21, 23, 445, 6379, 9200, 11211:
		return reports.SeverityCritical, "Cierre o restrinja inmediatamente este puerto a redes de administracion; si el servicio es imprescindible, apliquelo detras de firewall/VPN y autenticacion fuerte."
	case 25, 3306, 5432, 27017, 5900:
		return reports.SeverityHigh, "Revise si el servicio debe estar expuesto; limite origenes, active firewall y exija credenciales robustas o tuneles privados."
	case 22, 80, 443:
		return reports.SeverityMedium, "Valide que la exposicion sea intencional y que el servicio tenga hardening, parches y control de acceso adecuados."
	default:
		lower := strings.ToLower(strings.TrimSpace(service))
		if strings.Contains(lower, "telnet") || strings.Contains(lower, "ftp") || strings.Contains(lower, "redis") {
			return reports.SeverityHigh, "El servicio detectado suele requerir aislamiento estricto o sustitucion por alternativas mas seguras."
		}
		return reports.SeverityLow, "Confirme la necesidad operativa del puerto y documente su apertura en la configuracion del VPS."
	}
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}
