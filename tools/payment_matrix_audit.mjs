#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";

const repoRoot = process.cwd();
const outArgIndex = process.argv.indexOf("--out");
const outDir = outArgIndex >= 0 && process.argv[outArgIndex + 1]
  ? path.resolve(repoRoot, process.argv[outArgIndex + 1])
  : path.join(repoRoot, "documentos", "reportes_profesionales");

function read(rel) {
  return fs.existsSync(path.join(repoRoot, rel)) ? fs.readFileSync(path.join(repoRoot, rel), "utf8") : "";
}

const mainGo = read("backend/main.go");
const handlers = read("backend/handlers/payments_handlers.go");
const page = read("web/pagar_licencia.html");
const printQA = read("tools/qa_print_formats.cjs");
const contract = read("documentos/gobernanza_tecnica/contratos/contrato_checkout_licencias_publico.md");

const cases = [
  { id: "wompi_create", ok: /\/wompi\/create|Wompi/i.test(mainGo + handlers), expected: "creacion de transaccion Wompi" },
  { id: "wompi_status", ok: /\/wompi\/status|transaction_status|reference/i.test(mainGo + handlers), expected: "consulta por transaccion o referencia Wompi" },
  { id: "wompi_webhook", ok: /\/wompi\/webhook|webhook/i.test(mainGo + handlers), expected: "webhook Wompi idempotente" },
  { id: "epayco_create", ok: /\/epayco\/create|Epayco/i.test(mainGo + handlers), expected: "creacion de transaccion Epayco" },
  { id: "epayco_status", ok: /\/epayco\/status|transaction_status|reference/i.test(mainGo + handlers), expected: "consulta por transaccion o referencia Epayco" },
  { id: "epayco_webhook", ok: /\/epayco\/webhook|webhook/i.test(mainGo + handlers), expected: "webhook Epayco idempotente" },
  { id: "return_flow", ok: /epayco\/respuesta|reference|provider/i.test(page), expected: "retorno de pasarela sin perder referencia" },
  { id: "print_formats", ok: /factura|recibo|comprobante_ingreso|comprobante_egreso/i.test(printQA), expected: "evidencia visual de comprobantes grandes y POS" },
  { id: "contract_doc", ok: /Wompi|Epayco|webhook|referencia/i.test(contract), expected: "contrato de checkout documentado" },
];

const report = {
  generated_at: new Date().toISOString(),
  status: cases.every((item) => item.ok) ? "ok" : "warning",
  cases,
};

fs.mkdirSync(outDir, { recursive: true });
const stamp = report.generated_at.replace(/[-:]/g, "").replace(/\..+$/, "").replace("T", "_");
const jsonPath = path.join(outDir, `payment_matrix_audit_${stamp}.json`);
const mdPath = path.join(outDir, `payment_matrix_audit_${stamp}.md`);
fs.writeFileSync(jsonPath, JSON.stringify(report, null, 2), "utf8");
fs.writeFileSync(mdPath, [
  "# Matriz de pagos y comprobantes",
  "",
  `Fecha: ${report.generated_at}`,
  `Estado: ${report.status}`,
  "",
  ...cases.flatMap((item) => [
    `## ${item.ok ? "OK" : "REVISAR"} - ${item.id}`,
    item.expected,
    "",
  ]),
].join("\n"), "utf8");

console.log(JSON.stringify({ status: report.status, json: jsonPath, markdown: mdPath }, null, 2));
if (process.argv.includes("--strict") && report.status !== "ok") process.exitCode = 2;
