#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const requirements = [
  ["scripts/rs.ps1", "function Assert-ProductionRevision"],
  ["scripts/rs.ps1", "Actualizar repositorio omitido por DryRun/PreviewOnly"],
  ["scripts/sync_to_vps.ps1", "function Assert-ApprovedProductionRevision"],
  ["scripts/sync_to_vps.ps1", "$effectiveBootstrapServer = $false"],
  ["scripts/sync_to_vps.ps1", "exit $script:SyncExitCode"],
  ["deploy/scripts/vps-compose-sidecar-up.sh", "up -d --build --remove-orphans"],
];
const failed = requirements.filter(([file, text]) => !fs.readFileSync(path.join(root, file), "utf8").includes(text));
if (failed.length) { console.error(JSON.stringify({ status: "failed", failed }, null, 2)); process.exit(1); }
console.log(JSON.stringify({ status: "ok", checks: requirements.length }, null, 2));
