#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";

const repoRoot = process.cwd();
const outArgIndex = process.argv.indexOf("--out");
const outDir = outArgIndex >= 0 && process.argv[outArgIndex + 1]
  ? path.resolve(repoRoot, process.argv[outArgIndex + 1])
  : path.join(repoRoot, "documentos", "reportes_profesionales");

const roles = [
  {
    id: "super_administrador",
    name: "Super administrador",
    requiredPages: ["web/super_administrador.html", "web/super/alertas_sistema.html", "web/super/seguridad.html", "web/super/servidores.html", "web/super/licencias_resumen.html"],
    requiredApis: ["/super/api/alertas_sistema", "/super/api/errores", "/super/api/servidores"],
  },
  {
    id: "admin_empresa",
    name: "Administrador de empresa",
    requiredPages: ["web/administrar_empresa.html", "web/administrar_empresa/administrar_productos_menu.html", "web/administrar_empresa/configuracion_permisos.html"],
    requiredApis: ["/api/empresa/permisos", "/api/empresa/productos"],
  },
  {
    id: "cajero",
    name: "Cajero",
    requiredPages: ["web/administrar_empresa/corte_de_caja.html", "web/administrar_empresa/facturacion_electronica.html"],
    requiredApis: ["/api/empresa/corte_caja", "/api/empresa/facturacion_electronica"],
  },
  {
    id: "vendedor",
    name: "Vendedor",
    requiredPages: ["web/administrar_empresa/venta_publica.html", "web/administrar_empresa/administrar_clientes.html"],
    requiredApis: ["/api/public/venta_digital", "/api/empresa/clientes"],
  },
  {
    id: "asesor_comercial",
    name: "Asesor comercial",
    requiredPages: ["web/super/asesor_comercial.html"],
    requiredApis: ["/super/api/asesor_comercial"],
  },
  {
    id: "soporte",
    name: "Soporte operativo",
    requiredPages: ["web/administrar_empresa/soporte_remoto.html", "web/super/soporte_remoto.html"],
    requiredApis: ["/api/empresa/soporte_remoto", "/super/api/soporte_remoto"],
  },
];

function exists(rel) {
  return fs.existsSync(path.join(repoRoot, rel));
}

function readIf(rel) {
  const abs = path.join(repoRoot, rel);
  return fs.existsSync(abs) ? fs.readFileSync(abs, "utf8") : "";
}

const mainGo = readIf("backend/main.go");
const workflow = readIf(".github/workflows/e2e-visual.yml");

const checks = roles.map((role) => {
  const missingPages = role.requiredPages.filter((rel) => !exists(rel));
  const missingApis = role.requiredApis.filter((api) => !mainGo.includes(api));
  return {
    role: role.id,
    name: role.name,
    ok: missingPages.length === 0 && missingApis.length === 0,
    missing_pages: missingPages,
    missing_apis: missingApis,
  };
});

const workflowReady = /PCS_QA_EMAIL/.test(workflow) && /PCS_QA_PASSWORD/.test(workflow) && /PCS_QA_VIEWPORTS/.test(workflow);
const report = {
  generated_at: new Date().toISOString(),
  status: checks.every((item) => item.ok) && workflowReady ? "ok" : "warning",
  workflow_ready: workflowReady,
  roles: checks,
  recommendation: "Para pruebas reales por rol, configurar credenciales PCS_QA_ROLE_MATRIX_JSON en GitHub Actions y ejecutar qa_e2e_buttons por rol.",
};

fs.mkdirSync(outDir, { recursive: true });
const stamp = report.generated_at.replace(/[-:]/g, "").replace(/\..+$/, "").replace("T", "_");
const jsonPath = path.join(outDir, `qa_roles_matrix_${stamp}.json`);
const mdPath = path.join(outDir, `qa_roles_matrix_${stamp}.md`);
fs.writeFileSync(jsonPath, JSON.stringify(report, null, 2), "utf8");
fs.writeFileSync(mdPath, [
  "# QA por roles",
  "",
  `Fecha: ${report.generated_at}`,
  `Estado: ${report.status}`,
  "",
  `Workflow E2E listo: ${workflowReady ? "si" : "no"}`,
  "",
  ...checks.flatMap((role) => [
    `## ${role.ok ? "OK" : "REVISAR"} - ${role.name}`,
    "```json",
    JSON.stringify(role, null, 2),
    "```",
    "",
  ]),
].join("\n"), "utf8");

console.log(JSON.stringify({ status: report.status, json: jsonPath, markdown: mdPath }, null, 2));
if (process.argv.includes("--strict") && report.status !== "ok") process.exitCode = 2;
