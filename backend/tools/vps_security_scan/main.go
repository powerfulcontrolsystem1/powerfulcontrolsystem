package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/you/pos-backend/vpssecurity"
	"github.com/you/pos-backend/vpssecurity/config"
	storepkg "github.com/you/pos-backend/vpssecurity/logstore"
	"github.com/you/pos-backend/vpssecurity/scanner"
)

func main() {
	var configPath string
	var targetHost string
	var portList string
	var profile string
	var trigger string
	var triggeredBy string
	var dataDir string
	flag.StringVar(&configPath, "config", "", "Ruta al archivo de configuracion del escaner VPS")
	flag.StringVar(&targetHost, "target", "", "Host o IP objetivo")
	flag.StringVar(&portList, "ports", "", "Lista de puertos separados por coma")
	flag.StringVar(&profile, "profile", "", "Perfil quick o full")
	flag.StringVar(&trigger, "trigger", "cli", "Origen del disparo")
	flag.StringVar(&triggeredBy, "triggered-by", "cli", "Actor que lanza el escaneo")
	flag.StringVar(&dataDir, "data-dir", "", "Directorio para historial y reportes")
	flag.Parse()

	manager := config.NewManager(configPath)
	settings, err := manager.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if strings.TrimSpace(targetHost) != "" {
		settings.TargetHost = strings.TrimSpace(targetHost)
	}
	if strings.TrimSpace(portList) != "" {
		settings.PortList = strings.TrimSpace(portList)
	}
	if strings.TrimSpace(profile) != "" {
		settings.Profile = strings.ToLower(strings.TrimSpace(profile))
	}
	if strings.TrimSpace(dataDir) != "" {
		settings.DataDir = strings.TrimSpace(dataDir)
	}
	store := storepkg.NewStore(settings.DataDir)
	if err := store.Ensure(); err != nil {
		log.Fatalf("prepare store: %v", err)
	}
	report, err := vpssecurity.RunOnce(context.Background(), settings, store, scanner.SystemExecutor{}, trigger, triggeredBy, "")
	if err != nil {
		log.Fatalf("run scan: %v", err)
	}
	reportsDir := filepath.Join(store.Root(), "runs", report.ScanID, "reports")
	fmt.Printf("SCAN_ID=%s\n", report.ScanID)
	fmt.Printf("GENERATED_AT=%s\n", report.GeneratedAt)
	fmt.Printf("TARGET=%s\n", report.TargetHost)
	fmt.Printf("PROFILE=%s\n", report.Profile)
	fmt.Printf("HEALTH=%s\n", report.Summary.Health)
	fmt.Printf("TOTAL_FINDINGS=%d\n", report.Summary.TotalFindings)
	fmt.Printf("REPORT_JSON=%s\n", filepath.Join(reportsDir, "report.json"))
	fmt.Printf("REPORT_TXT=%s\n", filepath.Join(reportsDir, "report.txt"))
	fmt.Printf("REPORT_HTML=%s\n", filepath.Join(reportsDir, "report.html"))
	fmt.Printf("REPORT_PDF=%s\n", filepath.Join(reportsDir, "report.pdf"))
	fmt.Printf("REPORT_XLS=%s\n", filepath.Join(reportsDir, "report.xls"))
}
