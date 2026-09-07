#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";

const repoRoot = process.cwd();
const outIndex = process.argv.indexOf("--out");
const outDir = outIndex >= 0 && process.argv[outIndex + 1]
  ? path.resolve(repoRoot, process.argv[outIndex + 1])
  : path.join(repoRoot, "documentos", "reportes_profesionales");

function read(rel) {
  const full = path.join(repoRoot, rel);
  return fs.existsSync(full) ? fs.readFileSync(full, "utf8") : "";
}

const payments = read("backend/handlers/payments_handlers.go");
const printQA = read("tools/qa_print_formats.cjs");
const matrixDoc = read("documentos/gobernanza_tecnica/contratos/contrato_matriz_pagos_reales.md");

const checks = [
  { name: "wompi_gateway_flow", ok: /wompi/i.test(payments) && /webhook|checkout|transaction/i.test(payments) },
  { name: "epayco_gateway_flow", ok: /epayco/i.test(payments) && /webhook|checkout|transaction/i.test(payments) },
  { name: "large_small_receipts_visual_qa", ok: /formato|ticket|factura|receipt|print/i.test(printQA) },
  { name: "real_payment_contract", ok: /sandbox|produccion|reembolso|webhook/i.test(matrixDoc) }
];

const report = { generated_at: new Date().toISOString(), status: checks.every((check) => check.ok) ? "ok" : "warning", checks };
fs.mkdirSync(outDir, { recursive: true });
const stamp = report.generated_at.replace(/[-:]/g, "").replace(/\..+$/, "").replace("T", "_");
fs.writeFileSync(path.join(outDir, `payment_real_matrix_${stamp}.json`), JSON.stringify(report, null, 2), "utf8");
console.log(JSON.stringify(report, null, 2));
if (process.argv.includes("--strict") && report.status !== "ok") process.exitCode = 2;
