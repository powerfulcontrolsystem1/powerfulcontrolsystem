package metrics

import (
	"database/sql"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"

	dbpkg "github.com/you/pos-backend/db"
)

// StartCollector inicia la recolección periódica de métricas.
// intervalSeconds configura la periodicidad en segundos.
// stopCh permite detener el collector.
func StartCollector(dbConn *sql.DB, intervalSeconds int, stopCh <-chan struct{}) {
	if intervalSeconds <= 0 {
		intervalSeconds = 10
	}
	ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
	defer ticker.Stop()

	// Primera recolección inmediata. Esta función se conserva solo para
	// herramientas locales; producción programa CollectOnce desde pcs-worker.
	if err := CollectOnce(dbConn); err != nil {
		log.Println("metrics: failed to store metric:", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := CollectOnce(dbConn); err != nil {
				log.Println("metrics: failed to store metric:", err)
			}
		case <-stopCh:
			log.Println("metrics: collector stopped")
			return
		}
	}
}

// CollectOnce gathers one operational sample and stores it. It is safe to
// invoke from the durable worker and deliberately does not create schema.
func CollectOnce(dbConn *sql.DB) error {
	// CPU
	percents, err := cpu.Percent(0, false)
	var cpuPercent float64
	if err == nil && len(percents) > 0 {
		cpuPercent = percents[0]
	} else if err != nil {
		log.Println("metrics: cpu.Percent error:", err)
	}

	// Mem
	vm, err := mem.VirtualMemory()
	var memTotal uint64
	var memUsed uint64
	var memPercent float64
	if err == nil && vm != nil {
		memTotal = vm.Total
		memUsed = vm.Used
		memPercent = vm.UsedPercent
	} else if err != nil {
		log.Println("metrics: mem.VirtualMemory error:", err)
	}

	// Disco principal del VPS/contenedor. En Linux "/" representa el filesystem
	// montado para la app; en Windows local se mantiene compatible para pruebas.
	du, err := disk.Usage("/")
	var diskTotal uint64
	var diskUsed uint64
	var diskPercent float64
	if err == nil && du != nil {
		diskTotal = du.Total
		diskUsed = du.Used
		diskPercent = du.UsedPercent
	} else if err != nil {
		log.Println("metrics: disk.Usage error:", err)
	}

	// Network (agregado)
	netIOs, err := net.IOCounters(false)
	var netRecv uint64
	var netSent uint64
	if err == nil && len(netIOs) > 0 {
		netRecv = netIOs[0].BytesRecv
		netSent = netIOs[0].BytesSent
	} else if err != nil {
		log.Println("metrics: net.IOCounters error:", err)
	}

	if err := dbpkg.InsertMetric(dbConn, cpuPercent, memTotal, memUsed, memPercent, diskTotal, diskUsed, diskPercent, netRecv, netSent); err != nil {
		return err
	}
	log.Printf("metrics: stored cpu=%.2f mem=%.2f%% disk=%.2f%% recv=%d sent=%d", cpuPercent, memPercent, diskPercent, netRecv, netSent)
	return nil
}

// Read configuration helper (env METRICS_INTERVAL_SECONDS)
func DefaultIntervalSeconds() int {
	if s := os.Getenv("METRICS_INTERVAL_SECONDS"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			return v
		}
	}
	return 10
}
