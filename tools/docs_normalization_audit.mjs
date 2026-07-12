#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";

const repoRoot = process.cwd();
const outArgIndex = process.argv.indexOf("--out");
const outDir = outArgIndex >= 0 && process.argv[outArgIndex + 1]
  ? path.resolve(repoRoot, process.argv[outArgIndex + 1])
  : path.join(repoRoot, "documentos", "reportes_profesionales");

const targets = [
  "README.md",
  "RESUMEN_DEL_PROYECTO.md",
  "CHANGELOG.md",
  "documentos/descripcion_del_proyecto",
  "documentos/descripcion_de_archivos",
  "documentos/descripcion_de_modulos",
  "documentos/historial_de_cambios",
  "documentos/docker_vps_operacion.md",
  "documentos/plan_profesional_12_puntos.md",
];

const badPatterns = [/Ã./g, /Â./g, /\?n/g, /\?a/g, /\?o/g, /\uFFFD/g];
const findings = [];
const refinedBadPatterns = [
  /\u00c3\u0192./g,
  /\u00c3\u201a./g,
  /\u00c3[\u0080-\u00bf]/g,
  /\u00c2[\u0080-\u00bf]/g,
  /\u00e2\u20ac./g,
  /\u00ef\u00bf\u00bd/g,
  /\uFFFD/g,
];

for (const rel of targets) {
  const abs = path.join(repoRoot, rel);
  if (!fs.existsSync(abs)) continue;
  const text = fs.readFileSync(abs, "utf8");
  const matches = refinedBadPatterns.reduce((total, pattern) => total + (text.match(pattern) || []).length, 0);
  if (matches > 0) findings.push({ file: rel, suspicious_sequences: matches });
}

const report = {
  generated_at: new Date().toISOString(),
  status: findings.length ? "warning" : "ok",
  findings,
  recommendation: "Normalizar documentos historicos por lotes pequenos para no perder trazabilidad.",
};

fs.mkdirSync(outDir, { recursive: true });
const stamp = report.generated_at.replace(/[-:]/g, "").replace(/\..+$/, "").replace("T", "_");
const jsonPath = path.join(outDir, `docs_normalization_${stamp}.json`);
const mdPath = path.join(outDir, `docs_normalization_${stamp}.md`);
fs.writeFileSync(jsonPath, JSON.stringify(report, null, 2), "utf8");
fs.writeFileSync(mdPath, [
  "# Auditoria de normalizacion documental",
  "",
  `Fecha: ${report.generated_at}`,
  `Estado: ${report.status}`,
  "",
  ...findings.map((item) => `- ${item.file}: ${item.suspicious_sequences} secuencias sospechosas`),
  findings.length ? "" : "Sin secuencias sospechosas en documentos clave.",
  "",
].join("\n"), "utf8");

console.log(JSON.stringify({ status: report.status, json: jsonPath, markdown: mdPath }, null, 2));
if (process.argv.includes("--strict") && report.status !== "ok") process.exitCode = 2;
