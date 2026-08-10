#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";

const repoRoot = process.cwd();
const outIndex = process.argv.indexOf("--out");
const outDir = outIndex >= 0 && process.argv[outIndex + 1]
  ? path.resolve(repoRoot, process.argv[outIndex + 1])
  : path.join(repoRoot, "documentos", "reportes_profesionales");

function exists(rel) {
  return fs.existsSync(path.join(repoRoot, rel));
}

function read(rel) {
  return exists(rel) ? fs.readFileSync(path.join(repoRoot, rel), "utf8") : "";
}

const dashboard = read("deploy/monitoring/grafana/dashboards/pcs-operacion.json");
const alertRules = read("deploy/monitoring/alert_rules.yml");
const checks = [
  { name: "grafana_dashboard_vps", ok: /CPU VPS|Memoria usada|Disco usado/.test(dashboard) },
  { name: "backend_health_visible", ok: /pcs-backend|pcs-staging-backend/.test(dashboard) },
  { name: "database_worker_queue_visible", ok: /pcs_postgres_ready|pcs_worker_heartbeat_age_seconds|pcs_async_jobs_ready_total|pcs_outbox_ready_total|pcs_async_jobs_expired_leases_total/.test(dashboard) },
  { name: "capacity_alerts", ok: /PCSDiscoVPSAlto|PCSMemoriaAlta|PCSBackendCaido/.test(alertRules) },
  { name: "database_worker_queue_alerts", ok: /PCSPostgreSQLNoDisponible|PCSWorkerSinLatido|PCSColaDurableAcumulada|PCSLeaseDurableVencido|PCSConsultaObservabilidadFallida/.test(alertRules) },
  { name: "email_alert_module_exists", ok: exists("web/super/alertas_sistema.html") && /alertas/i.test(read("backend/handlers/super_alertas.go")) }
];

const report = {
  generated_at: new Date().toISOString(),
  status: checks.every((check) => check.ok) ? "ok" : "warning",
  checks
};

fs.mkdirSync(outDir, { recursive: true });
const stamp = report.generated_at.replace(/[-:]/g, "").replace(/\..+$/, "").replace("T", "_");
fs.writeFileSync(path.join(outDir, `business_observability_${stamp}.json`), JSON.stringify(report, null, 2), "utf8");
console.log(JSON.stringify(report, null, 2));
if (process.argv.includes("--strict") && report.status !== "ok") process.exitCode = 2;
