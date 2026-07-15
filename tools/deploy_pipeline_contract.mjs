#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const read = (relativePath) => fs.readFileSync(path.join(root, relativePath), "utf8");
const checks = [
  { file: "scripts/rs.ps1", text: "function Assert-ProductionRevision", reason: "rs must verify the approved production revision before deployment" },
  { file: "scripts/rs.ps1", text: "Actualizar repositorio omitido por DryRun/PreviewOnly", reason: "rs previews must not update Git or deploy" },
  { file: "scripts/sync_to_vps.ps1", text: "[switch]$AllowLegacySecretBootstrap", reason: "legacy secret bootstrap requires an explicit opt-in" },
  { file: "scripts/sync_to_vps.ps1", text: "$effectiveBootstrapServer = $false", reason: "Docker deployment must retain secrets on the remote host" },
  { file: "scripts/sync_to_vps.ps1", text: "exit $script:SyncExitCode", reason: "sync failures must reach the rs process" },
  { file: "deploy/scripts/vps-compose-sidecar-up.sh", text: "up -d --build --remove-orphans", reason: "deployment must remove stale Compose workers" },
];

const failures = checks.filter(({ file, text }) => !read(file).includes(text)).map(({ file, reason }) => `${file}: ${reason}`);
if (failures.length) {
  console.error(JSON.stringify({ status: "failed", failures }, null, 2));
  process.exit(1);
}
console.log(JSON.stringify({ status: "ok", checks: checks.length }, null, 2));
