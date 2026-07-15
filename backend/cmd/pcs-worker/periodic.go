package main

import (
	"context"
	"database/sql"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/handlers"
	"github.com/you/pos-backend/internal/platform/worker"
	"github.com/you/pos-backend/metrics"
	"github.com/you/pos-backend/utils"
)

// startPeriodicWorkers is the single home for recurring PCS work. API replicas
// never invoke these loops, so scaling HTTP cannot duplicate fiscal reminders,
// snapshots, retention or accounting work.
func startPeriodicWorkers(ctx context.Context, empresasDB, superDB *sql.DB) {
	startLoop(ctx, "metrics.collector", func(stop <-chan struct{}) {
		metrics.StartCollector(superDB, metrics.DefaultIntervalSeconds(), stop)
	})
	startLoop(ctx, "super.alertas_worker", func(stop <-chan struct{}) {
		handlers.StartSuperAlertasWorker(superDB, time.Minute, stop)
	})
	startLoop(ctx, "auditoria.retention_worker", func(stop <-chan struct{}) {
		dbpkg.StartEmpresaAuditoriaRetentionWorker(empresasDB, 12*time.Hour, stop)
	})
	startLoop(ctx, "licencias.estado_empresas_worker", func(stop <-chan struct{}) {
		dbpkg.StartLicenciaEmpresaEstadoWorker(empresasDB, superDB, time.Hour, stop)
	})
	startLoop(ctx, "licencias.vencimiento_alertas_worker", func(stop <-chan struct{}) {
		handlers.StartLicenciaVencimientoAlertasWorker(superDB, empresasDB, 12*time.Hour, stop)
	})
	startLoop(ctx, "super.vps_snapshot_worker", func(stop <-chan struct{}) {
		handlers.StartSuperVPSSnapshotWorker(superDB, time.Hour, stop)
	})
	startLoop(ctx, "super.mantenimiento_agentes_worker", func(stop <-chan struct{}) {
		handlers.StartSuperMantenimientoAgentesWorker(superDB, time.Minute, stop)
	})
	startLoop(ctx, "parametros_legales.worker", func(stop <-chan struct{}) {
		dbpkg.StartEmpresaParametrosLegalesWorker(empresasDB, 24*time.Hour, stop)
	})
	startLoop(ctx, "cobranza.recordatorios_worker", func(stop <-chan struct{}) {
		handlers.StartEmpresaCobranzaRecordatoriosWorker(empresasDB, superDB, time.Hour, stop)
	})
	startLoop(ctx, "finanzas.asientos_worker", func(stop <-chan struct{}) {
		dbpkg.StartEmpresaAsientosContablesWorker(empresasDB, 5*time.Minute, 100, 5, stop)
	})
	startLoop(ctx, "control_electrico.programacion_worker", func(stop <-chan struct{}) {
		handlers.StartControlElectricoProgramacionWorker(empresasDB, time.Minute, stop)
	})
}

func startLoop(ctx context.Context, module string, run func(stop <-chan struct{})) {
	stop := make(chan struct{})
	go func() {
		<-ctx.Done()
		close(stop)
	}()
	go utils.RunProtectedProcess(module, map[string]interface{}{"role": "pcs-worker"}, func() { run(stop) })
}

// The registry is intentionally closed. Domain handlers are added only when a
// business flow writes a matching transactional outbox event and regression
// test; unknown work is retried/dead-lettered rather than silently discarded.
func productionAsyncJobRegistry(empresasDB, superDB *sql.DB) []worker.HandlerSpec {
	return nil
}

func productionOutboxHandlers(superDB *sql.DB) map[string]worker.OutboxHandler {
	return map[string]worker.OutboxHandler{}
}
