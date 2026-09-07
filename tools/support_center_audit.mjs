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
  return exists(rel) ? fs.readFileSync(path.join(repoRoot, rel), "utf8") : "";
}

const mainGo = read("backend/main.go");
const checks = [
  { name: "soporte_remoto_empresa", ok: exists("web/administrar_empresa/soporte_remoto.html") && /soporte_remoto/.test(mainGo) },
  { name: "soporte_remoto_super", ok: exists("web/super/soporte_remoto.html") && /super.*soporte_remoto|soporte_remoto/i.test(mainGo) },
  { name: "errores_sistema_super", ok: exists("backend/db/super_errores_sistema.go") && /\/super\/api\/errores/.test(mainGo) },
  { name: "alertas_sistema_super", ok: exists("backend/db/super_alertas.go") && exists("web/super/alertas_sistema.html") },
  { name: "auditoria_empresa", ok: exists("backend/db/auditoria_empresa.go") && exists("backend/handlers/auditoria_empresa.go") },
  { name: "runbook_release", ok: exists("documentos/gobernanza_tecnica/runbooks/runbook_release_profesional.md") },
];

const report = {
  generated_at: new Date().toISOString(),
  status: checks.every((item) => item.ok) ? "ok" : "warning",
  checks,
};

fs.mkdirSync(outDir, { recursive: true });
const stamp = report.generated_at.replace(/[-:]/g, "").replace(/\..+$/, "").replace("T", "_");
const jsonPath = path.join(outDir, `support_center_audit_${stamp}.json`);
const mdPath = path.join(outDir, `support_center_audit_${stamp}.md`);
fs.writeFileSync(jsonPath, JSON.stringify(report, null, 2), "utf8");
fs.writeFileSync(mdPath, [
  "# Centro de soporte interno",
  "",
  `Fecha: ${report.generated_at}`,
  `Estado: ${report.status}`,
  "",
  ...checks.flatMap((item) => [
    `## ${item.ok ? "OK" : "REVISAR"} - ${item.name}`,
    "```json",
    JSON.stringify(item, null, 2),
    "```",
    "",
  ]),
].join("\n"), "utf8");

console.log(JSON.stringify({ status: report.status, json: jsonPath, markdown: mdPath }, null, 2));
if (process.argv.includes("--strict") && report.status !== "ok") process.exitCode = 2;
