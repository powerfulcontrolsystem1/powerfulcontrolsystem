#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";

const repoRoot = process.cwd();
const outIndex = process.argv.indexOf("--out");
const outDir = outIndex >= 0 && process.argv[outIndex + 1]
  ? path.resolve(repoRoot, process.argv[outIndex + 1])
  : path.join(repoRoot, "documentos", "reportes_profesionales");

const scriptPath = path.join(repoRoot, "deploy/scripts/vps-hardening-audit.sh");
const docPath = path.join(repoRoot, "documentos/gobernanza_tecnica/runbooks/runbook_hardening_vps.md");
const script = fs.existsSync(scriptPath) ? fs.readFileSync(scriptPath, "utf8") : "";
const doc = fs.existsSync(docPath) ? fs.readFileSync(docPath, "utf8") : "";

const checks = [
  { name: "ssh_hardening_checked", ok: /PermitRootLogin|PasswordAuthentication/.test(script + doc) },
  { name: "firewall_checked", ok: /ufw|firewall/i.test(script + doc) },
  { name: "fail2ban_checked", ok: /fail2ban/i.test(script + doc) },
  { name: "docker_runtime_checked", ok: /docker compose|docker ps/i.test(script + doc) }
];

const report = { generated_at: new Date().toISOString(), status: checks.every((check) => check.ok) ? "ok" : "warning", checks };
fs.mkdirSync(outDir, { recursive: true });
const stamp = report.generated_at.replace(/[-:]/g, "").replace(/\..+$/, "").replace("T", "_");
fs.writeFileSync(path.join(outDir, `vps_hardening_${stamp}.json`), JSON.stringify(report, null, 2), "utf8");
console.log(JSON.stringify(report, null, 2));
if (process.argv.includes("--strict") && report.status !== "ok") process.exitCode = 2;
