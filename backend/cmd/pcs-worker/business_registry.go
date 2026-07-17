package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/handlers"
	platformworker "github.com/you/pos-backend/internal/platform/worker"
	ventasdomain "github.com/you/pos-backend/internal/ventas"
	"github.com/you/pos-backend/metrics"
)

const (
	jobSuperAlerts        = "maintenance.super-alerts"
	jobAuditRetention     = "maintenance.audit-retention"
	jobLicenseState       = "licenses.state-sync"
	jobLicenseAlerts      = "licenses.expiry-alerts"
	jobVPSSnapshot        = "maintenance.vps-snapshot"
	jobDIANNewsAgent      = "integrations.dian-news-agent"
	jobLegalParameters    = "compliance.legal-parameters"
	jobCollections        = "notifications.collections"
	jobAccounting         = "accounting.pending-events"
	jobElectricalSchedule = "integrations.electrical-schedule"
	jobCommerceSalePaid   = "commerce.sale-paid"
	jobSystemMetrics      = "maintenance.system-metrics"
)

func businessRegistry(dbEmp, dbSuper *sql.DB) map[string]platformworker.HandlerSpec {
	registry := make(map[string]platformworker.HandlerSpec)
	ventasService := ventasdomain.Service{DB: dbEmp}
	add := func(kind string, timeout time.Duration, maxAttempts int, fn func(context.Context) error) {
		registry[kind] = platformworker.HandlerSpec{Kind: kind, Version: 1, Timeout: timeout, MaxAttempts: maxAttempts, Enabled: true, Handle: func(ctx context.Context, _ dbpkg.AsyncJob) error { return fn(ctx) }}
	}
	add(jobSuperAlerts, 2*time.Minute, 5, func(context.Context) error { handlers.EvaluateSuperAlertasSistema(dbSuper, false); return nil })
	add(jobAuditRetention, 10*time.Minute, 5, func(context.Context) error { _, err := dbpkg.PurgeExpiredEmpresaAuditoriaEventos(dbEmp); return err })
	add(jobLicenseState, 10*time.Minute, 8, func(context.Context) error { _, err := dbpkg.SyncEmpresasEstadoPorLicencia(dbEmp, dbSuper); return err })
	add(jobLicenseAlerts, 20*time.Minute, 8, func(context.Context) error { return handlers.RunLicenciaVencimientoScheduled(dbSuper, dbEmp) })
	add(jobVPSSnapshot, 30*time.Minute, 3, func(context.Context) error { return handlers.RunSuperVPSSnapshotScheduled(dbSuper) })
	add(jobDIANNewsAgent, 20*time.Minute, 5, func(context.Context) error { return handlers.RunSuperMaintenanceAgentsScheduled(dbSuper) })
	add(jobLegalParameters, 30*time.Minute, 5, func(context.Context) error {
		_, err := dbpkg.CheckAndApplyEmpresaParametrosLegalesAuto(dbEmp, "sistema.pcs-worker")
		return err
	})
	add(jobCollections, 25*time.Minute, 8, func(context.Context) error { return handlers.RunEmpresaCobranzaRecordatoriosScheduled(dbEmp, dbSuper) })
	add(jobAccounting, 25*time.Minute, 10, func(context.Context) error {
		result, err := dbpkg.RunEmpresaAsientosContablesWorkerCycle(dbEmp, "pcs-worker", envInt("ASIENTOS_WORKER_BATCH_SIZE", 100), envInt("ASIENTOS_WORKER_MAX_RETRIES", 5))
		if err == nil && result.Fallidos > 0 {
			return fmt.Errorf("accounting cycle left %d failed event(s)", result.Fallidos)
		}
		return err
	})
	add(jobElectricalSchedule, 10*time.Minute, 5, func(context.Context) error {
		_, err := handlers.EjecutarControlElectricoProgramacionPendiente(dbEmp, time.Now())
		return err
	})
	add(jobSystemMetrics, 2*time.Minute, 5, func(context.Context) error {
		return metrics.CollectOnce(dbSuper)
	})
	registry[jobCommerceSalePaid] = platformworker.HandlerSpec{
		Kind: jobCommerceSalePaid, Version: 1, Timeout: 5 * time.Minute, MaxAttempts: 10, Enabled: true,
		Handle: ventasService.RecoverPaidSaleAccounting,
	}
	return registry
}

func businessSchedules() []platformworker.ScheduleSpec {
	return []platformworker.ScheduleSpec{
		{Kind: jobSuperAlerts, Version: 1, Interval: time.Minute, MaxAttempts: 5, Priority: 80},
		{Kind: jobAuditRetention, Version: 1, Interval: 12 * time.Hour, MaxAttempts: 5, Priority: 150},
		{Kind: jobLicenseState, Version: 1, Interval: time.Hour, MaxAttempts: 8, Priority: 60},
		{Kind: jobLicenseAlerts, Version: 1, Interval: 12 * time.Hour, MaxAttempts: 8, Priority: 70},
		{Kind: jobVPSSnapshot, Version: 1, Interval: time.Hour, MaxAttempts: 3, Priority: 200},
		{Kind: jobDIANNewsAgent, Version: 1, Interval: time.Minute, MaxAttempts: 5, Priority: 120},
		{Kind: jobLegalParameters, Version: 1, Interval: 24 * time.Hour, MaxAttempts: 5, Priority: 140},
		{Kind: jobCollections, Version: 1, Interval: time.Hour, MaxAttempts: 8, Priority: 90},
		{Kind: jobAccounting, Version: 1, Interval: time.Duration(envInt("ASIENTOS_WORKER_INTERVAL_MINUTES", 15)) * time.Minute, MaxAttempts: 10, Priority: 50},
		{Kind: jobElectricalSchedule, Version: 1, Interval: time.Minute, MaxAttempts: 5, Priority: 40},
		{Kind: jobSystemMetrics, Version: 1, Interval: time.Duration(metrics.DefaultIntervalSeconds()) * time.Second, MaxAttempts: 5, Priority: 160},
	}
}

func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv(key)))
	if err != nil || value < 1 {
		return fallback
	}
	return value
}
