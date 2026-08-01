#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const requirements = [
  ["scripts/rs.ps1", "function Assert-ProductionRevision"],
  ["scripts/rs.ps1", "Actualizar repositorio omitido por DryRun/PreviewOnly"],
  ["scripts/sync_to_vps.ps1", "function Assert-ApprovedProductionRevision"],
  ["scripts/sync_to_vps.ps1", "$effectiveBootstrapServer = $true"],
  ["scripts/sync_to_vps.ps1", "bootstrap remoto limitado a configuración operativa"],
  ["scripts/sync_to_vps.ps1", "exit $script:SyncExitCode"],
  ["deploy/scripts/vps-compose-sidecar-up.sh", "up -d --build --remove-orphans"],
  ["deploy/docker-compose.release.yml", "PCS_FRONTEND_IMAGE_DIGEST"],
  [".github/workflows/release-candidate.yml", "packages: write"],
  [".github/workflows/release-candidate.yml", "docker build --target api"],
  [".github/workflows/release-candidate.yml", "security-artifacts/release-images.env"],
  ["deploy/scripts/vps-staging-digest-up.sh", "--no-build"],
  ["deploy/scripts/vps-p108-empty-migration-drill.sh", "P109_VERIFY_MIGRATION_FAILURES debe ser 0 o 1."],
  ["deploy/scripts/vps-p108-empty-migration-drill.sh", "PCS_PREVIOUS_API_IMAGE_DIGEST debe usar repositorio@sha256:<64 hex>."],
  ["deploy/scripts/vps-p108-empty-migration-drill.sh", "El DDL sobrevivió al rollback transaccional."],
  ["deploy/scripts/vps-p108-empty-migration-drill.sh", "El ledger registró una migración fallida."],
  ["deploy/scripts/vps-p108-empty-migration-drill.sh", "compatibilidad_anterior="],
  ["deploy/scripts/vps-p108-empty-migration-drill.sh", "trap 'exit 143' TERM"],
  ["deploy/scripts/vps-p109-restored-app-drill.sh", "P109_DRILL_ID debe iniciar con p109-restore-app-."],
  ["deploy/scripts/vps-p109-restored-app-drill.sh", "PCS_API_IMAGE_DIGEST debe usar repositorio@sha256:<64 hex>."],
  ["deploy/scripts/vps-p109-restored-app-drill.sh", "PCS_MIGRATE_IMAGE_DIGEST debe usar repositorio@sha256:<64 hex>."],
  ["deploy/scripts/vps-p109-restored-app-drill.sh", "runtime_privilegios=0"],
  ["deploy/scripts/vps-p109-restored-app-drill.sh", "docker network rm"],
  ["deploy/scripts/vps-p109-restored-app-drill.sh", "P109_HOLD_SECONDS no puede superar 900."],
  ["deploy/scripts/vps-p109-restored-app-drill.sh", "trap 'exit 143' TERM"],
  ["deploy/scripts/vps-p109-restored-app-drill.sh", "P109_VERIFY_REPLICA debe ser 0 o 1."],
  ["deploy/scripts/vps-p109-restored-app-drill.sh", "P109_VERIFY_COORDINATED_ROLLBACK debe ser 0 o 1."],
  ["deploy/scripts/vps-p109-restored-app-drill.sh", "El rollback coordinado requiere P109_VERIFY_REPLICA=1."],
  ["deploy/scripts/vps-p109-restored-app-drill.sh", "P109_VERIFY_PRIVATE_INVENTORY debe ser 0 o 1."],
  ["deploy/scripts/vps-p109-restored-app-drill.sh", "El restore privado tiene referencias sin archivo."],
  ["deploy/scripts/vps-p109-restored-app-drill.sh", "huerfanos_privados="],
  ["deploy/scripts/vps-p109-restored-app-drill.sh", "referencias_heredadas="],
  ["deploy/scripts/vps-p109-restored-app-drill.sh", "X-CSRF-Token"],
  ["deploy/scripts/vps-p109-restored-app-drill.sh", "replica_checks="],
  ["deploy/scripts/vps-p109-restored-app-drill.sh", "archivos_hostiles="],
  ["deploy/scripts/vps-p109-restored-app-drill.sh", "dropdb --force"],
  ["deploy/scripts/vps-p109-restored-app-drill.sh", "El inventario privado cambio despues del rollback."],
  ["deploy/scripts/vps-p109-restored-app-drill.sh", "rollback_checks="],
  ["deploy/scripts/vps-p109-restored-app-drill.sh", "rollback_dominios="],
  ["deploy/scripts/vps-p109-restored-app-drill.sh", "Una carga hostil creo una fila empresarial."],
];
const failed = requirements.filter(([file, text]) => !fs.readFileSync(path.join(root, file), "utf8").includes(text));
if (failed.length) { console.error(JSON.stringify({ status: "failed", failed }, null, 2)); process.exit(1); }
console.log(JSON.stringify({ status: "ok", checks: requirements.length }, null, 2));
