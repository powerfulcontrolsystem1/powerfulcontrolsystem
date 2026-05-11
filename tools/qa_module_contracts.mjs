#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";

const repoRoot = process.cwd();
const outArgIndex = process.argv.indexOf("--out");
const strict = process.argv.includes("--strict");
const outDir = outArgIndex >= 0 && process.argv[outArgIndex + 1]
  ? path.resolve(repoRoot, process.argv[outArgIndex + 1])
  : path.join(repoRoot, "documentos", "reportes_profesionales");

const required = [
  "web/login.html",
  "web/elegir_licencia.html",
  "web/pagar_licencia.html",
  "web/super_administrador.html",
  "web/administrar_empresa.html",
  "web/js/print_documents.js",
  "tools/qa_e2e_buttons.cjs",
  "tools/qa_print_formats.cjs",
  "documentos/gobernanza_tecnica/contratos/contrato_checkout_licencias_publico.md",
  "documentos/gobernanza_tecnica/contratos/contrato_facturacion_electronica_y_documentos_transaccionales.md",
  "documentos/gobernanza_tecnica/contratos/contrato_permisos_contexto_y_wrappers_api_empresa.md",
];

const missing = required.filter((rel) => !fs.existsSync(path.join(repoRoot, rel)));
const printSource = fs.existsSync(path.join(repoRoot, "web/js/print_documents.js"))
  ? fs.readFileSync(path.join(repoRoot, "web/js/print_documents.js"), "utf8")
  : "";
const printKinds = ["factura", "recibo", "comprobante"].filter((kind) => printSource.includes(kind));

const report = {
  generated_at: new Date().toISOString(),
  status: missing.length === 0 && printKinds.length >= 3 ? "ok" : "warning",
  checks: [
    { name: "critical_files", ok: missing.length === 0, missing },
    { name: "print_contracts", ok: printKinds.length >= 3, detected: printKinds },
  ],
};

fs.mkdirSync(outDir, { recursive: true });
const stamp = report.generated_at.replace(/[-:]/g, "").replace(/\..+$/, "").replace("T", "_");
const jsonPath = path.join(outDir, `qa_module_contracts_${stamp}.json`);
const mdPath = path.join(outDir, `qa_module_contracts_${stamp}.md`);
fs.writeFileSync(jsonPath, JSON.stringify(report, null, 2), "utf8");
fs.writeFileSync(mdPath, `# QA funcional por modulos criticos\n\nEstado: ${report.status}\n\n\`\`\`json\n${JSON.stringify(report, null, 2)}\n\`\`\`\n`, "utf8");
console.log(JSON.stringify({ status: report.status, json: jsonPath, markdown: mdPath }, null, 2));
if (strict && report.status !== "ok") process.exitCode = 2;
