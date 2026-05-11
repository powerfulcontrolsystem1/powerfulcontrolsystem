#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";

const repoRoot = process.cwd();
const outIndex = process.argv.indexOf("--out");
const outDir = outIndex >= 0 && process.argv[outIndex + 1]
  ? path.resolve(repoRoot, process.argv[outIndex + 1])
  : path.join(repoRoot, "documentos", "reportes_profesionales");

function read(rel) {
  return fs.existsSync(path.join(repoRoot, rel)) ? fs.readFileSync(path.join(repoRoot, rel), "utf8") : "";
}

const checks = [
  {
    name: "refresh_staging_runs_anonymizer",
    ok: /vps-anonymize-staging\.sh/.test(read("deploy/scripts/vps-refresh-staging-from-production.sh")),
    evidence: "El refresco de staging invoca anonimizacion por defecto."
  },
  {
    name: "anonymizer_masks_admin_data",
    ok: /information_schema\.columns/i.test(read("deploy/scripts/vps-anonymize-staging.sh")) && /email|correo|telefono|documento|direccion/i.test(read("deploy/scripts/vps-anonymize-staging.sh")) && /staging\.local|example\.test/i.test(read("deploy/scripts/vps-anonymize-staging.sh")),
    evidence: "El anonimizador cubre columnas sensibles por introspeccion y dominios de prueba."
  },
  {
    name: "verification_script_exists",
    ok: fs.existsSync(path.join(repoRoot, "deploy/scripts/vps-verify-staging-anonymization.sh")),
    evidence: "Existe script de verificacion post-refresh."
  }
];

const report = {
  generated_at: new Date().toISOString(),
  status: checks.every((check) => check.ok) ? "ok" : "warning",
  checks
};

fs.mkdirSync(outDir, { recursive: true });
const stamp = report.generated_at.replace(/[-:]/g, "").replace(/\..+$/, "").replace("T", "_");
fs.writeFileSync(path.join(outDir, `staging_anonymization_${stamp}.json`), JSON.stringify(report, null, 2), "utf8");
fs.writeFileSync(path.join(outDir, `staging_anonymization_${stamp}.md`), [
  "# Auditoria de anonimizacion staging",
  "",
  `Fecha: ${report.generated_at}`,
  `Estado: ${report.status}`,
  "",
  ...checks.map((check) => `- ${check.ok ? "OK" : "REVISAR"} ${check.name}: ${check.evidence}`)
].join("\n"), "utf8");

console.log(JSON.stringify({ status: report.status, checks }, null, 2));
if (process.argv.includes("--strict") && report.status !== "ok") process.exitCode = 2;
