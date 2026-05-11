#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";

const repoRoot = process.cwd();
const outArgIndex = process.argv.indexOf("--out");
const outDir = outArgIndex >= 0 && process.argv[outArgIndex + 1]
  ? path.resolve(repoRoot, process.argv[outArgIndex + 1])
  : path.join(repoRoot, "documentos", "reportes_profesionales");

function exists(rel) {
  return fs.existsSync(path.join(repoRoot, rel));
}

function read(rel) {
  return fs.readFileSync(path.join(repoRoot, rel), "utf8");
}

function count(pattern, text) {
  return (text.match(pattern) || []).length;
}

const compose = exists("deploy/docker-compose.platform.yml") ? read("deploy/docker-compose.platform.yml") : "";
const mainGo = exists("backend/main.go") ? read("backend/main.go") : "";

const checks = [
  {
    name: "docker_healthchecks",
    ok: count(/healthcheck:/g, compose) >= 2,
    count: count(/healthcheck:/g, compose),
  },
  {
    name: "persistent_backend_logs",
    ok: /pcs_backend_logs/.test(compose),
  },
  {
    name: "alertas_sistema_module",
    ok: exists("web/super/alertas_sistema.html") && /alertas_sistema/.test(mainGo),
  },
  {
    name: "vps_security_scanner",
    ok: exists("scripts/run_vps_security_scan.sh") && exists("backend/vpssecurity/service.go"),
  },
  {
    name: "professional_reports_ignored",
    ok: exists(".gitignore") && /documentos\/reportes_profesionales\//.test(read(".gitignore")),
  },
  {
    name: "runtime_state_log",
    ok: exists("backend/logs/server_runtime_state.json") || /server_runtime_state/.test(mainGo),
  },
];

const report = {
  generated_at: new Date().toISOString(),
  status: checks.every((c) => c.ok) ? "ok" : "warning",
  checks,
};

fs.mkdirSync(outDir, { recursive: true });
const stamp = report.generated_at.replace(/[-:]/g, "").replace(/\..+$/, "").replace("T", "_");
const jsonPath = path.join(outDir, `observability_report_${stamp}.json`);
const mdPath = path.join(outDir, `observability_report_${stamp}.md`);
fs.writeFileSync(jsonPath, JSON.stringify(report, null, 2), "utf8");
fs.writeFileSync(mdPath, [
  "# Reporte de observabilidad",
  "",
  `Fecha: ${report.generated_at}`,
  `Estado: ${report.status}`,
  "",
  ...checks.flatMap((check) => [
    `## ${check.ok ? "OK" : "REVISAR"} - ${check.name}`,
    "```json",
    JSON.stringify(check, null, 2),
    "```",
    "",
  ]),
].join("\n"), "utf8");

console.log(JSON.stringify({ status: report.status, json: jsonPath, markdown: mdPath }, null, 2));
