#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";

const repoRoot = process.cwd();
const outArgIndex = process.argv.indexOf("--out");
const strict = process.argv.includes("--strict");
const outDir = outArgIndex >= 0 && process.argv[outArgIndex + 1]
  ? path.resolve(repoRoot, process.argv[outArgIndex + 1])
  : path.join(repoRoot, "documentos", "reportes_profesionales");

function read(rel) {
  return fs.readFileSync(path.join(repoRoot, rel), "utf8");
}

function extractAll(text, regex, group = 1) {
  return [...text.matchAll(regex)].map((m) => String(m[group] || "").trim()).filter(Boolean);
}

const permisosGo = read("backend/handlers/empresa_permisos.go");
const licenciasHtml = read("web/super/licencias.html");
const adminJs = read("web/js/administrar_empresa.js");

const backendModules = [...new Set(extractAll(permisosGo, /permModule[A-Za-z0-9_]+\s+=\s+"([^"]+)"/g))].sort();
const wrappers = [...new Set(extractAll(permisosGo, /\bfunc\s+(WithEmpresa[A-Za-z0-9]+Permissions)\b/g))].sort();
const licenseMentions = [...new Set(extractAll(licenciasHtml, /value=["']([^"']+)["']/g).filter((value) => backendModules.includes(value)))].sort();
const menuModules = [...new Set(extractAll(adminJs, /module["']?\s*:\s*["']([^"']+)["']/g))].sort();
const jsLiteralModules = [...new Set(extractAll(adminJs, /["']([a-z0-9_]+)["']/g).filter((value) => backendModules.includes(value)))].sort();

const missingInLicenses = backendModules.filter((module) => !licenseMentions.includes(module) && !licenciasHtml.includes(module));
const missingInFrontend = backendModules.filter((module) => !jsLiteralModules.includes(module) && !menuModules.includes(module));

const checks = [
  {
    name: "backend_permission_modules",
    ok: backendModules.length >= 40,
    count: backendModules.length,
  },
  {
    name: "permission_wrappers",
    ok: wrappers.length >= 40,
    count: wrappers.length,
  },
  {
    name: "license_module_coverage",
    ok: missingInLicenses.length === 0,
    missing: missingInLicenses,
  },
  {
    name: "frontend_permission_coverage",
    ok: missingInFrontend.length <= 10,
    missing_warning: missingInFrontend.slice(0, 60),
    missing_total: missingInFrontend.length,
  },
];

const report = {
  generated_at: new Date().toISOString(),
  status: checks.every((c) => c.ok) ? "ok" : "warning",
  modules: backendModules,
  checks,
};

fs.mkdirSync(outDir, { recursive: true });
const stamp = report.generated_at.replace(/[-:]/g, "").replace(/\..+$/, "").replace("T", "_");
const jsonPath = path.join(outDir, `permissions_license_audit_${stamp}.json`);
const mdPath = path.join(outDir, `permissions_license_audit_${stamp}.md`);
fs.writeFileSync(jsonPath, JSON.stringify(report, null, 2), "utf8");
fs.writeFileSync(mdPath, [
  "# Auditoria de permisos y licencias",
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
if (strict && report.status !== "ok") process.exitCode = 2;
