#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";

const repoRoot = process.cwd();
const outIndex = process.argv.indexOf("--out");
const outDir = outIndex >= 0 && process.argv[outIndex + 1]
  ? path.resolve(repoRoot, process.argv[outIndex + 1])
  : path.join(repoRoot, "documentos", "reportes_profesionales");
const docPath = path.join(repoRoot, "documentos/gobernanza_tecnica/slo_sla_operativo.md");
const doc = fs.existsSync(docPath) ? fs.readFileSync(docPath, "utf8") : "";

const checks = [
  { name: "slo_availability_defined", ok: /disponibilidad/i.test(doc) && /99\.5|99,5|99\.9|99,9/.test(doc) },
  { name: "rto_rpo_defined", ok: /\bRTO\b/i.test(doc) && /\bRPO\b/i.test(doc) },
  { name: "incident_severity_defined", ok: /Severidad|P1|P2|P3/i.test(doc) },
  { name: "release_gate_referenced", ok: /release_gate|preflight/i.test(doc) }
];

const report = { generated_at: new Date().toISOString(), status: checks.every((check) => check.ok) ? "ok" : "warning", checks };
fs.mkdirSync(outDir, { recursive: true });
const stamp = report.generated_at.replace(/[-:]/g, "").replace(/\..+$/, "").replace("T", "_");
fs.writeFileSync(path.join(outDir, `slo_sla_${stamp}.json`), JSON.stringify(report, null, 2), "utf8");
console.log(JSON.stringify(report, null, 2));
if (process.argv.includes("--strict") && report.status !== "ok") process.exitCode = 2;
