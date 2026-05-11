#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";

const repoRoot = process.cwd();
const outArgIndex = process.argv.indexOf("--out");
const checkOnly = process.argv.includes("--check");
const outFile = outArgIndex >= 0 && process.argv[outArgIndex + 1]
  ? path.resolve(repoRoot, process.argv[outArgIndex + 1])
  : path.join(repoRoot, "documentos", "api", "openapi.generated.yaml");

const mainGoPath = path.join(repoRoot, "backend", "main.go");
const mainGo = fs.readFileSync(mainGoPath, "utf8");
const routes = [...mainGo.matchAll(/http\.HandleFunc\("([^"]+)"/g)]
  .map((m) => m[1])
  .filter((route) => route.startsWith("/"))
  .sort((a, b) => a.localeCompare(b));

function yamlString(value) {
  return JSON.stringify(String(value));
}

const lines = [
  "openapi: 3.0.3",
  "info:",
  "  title: Powerful Control System API",
  "  version: \"generated\"",
  "  description: Inventario automatico de rutas registradas en backend/main.go. Completar contratos detallados por modulo en documentos/gobernanza_tecnica/contratos.",
  "servers:",
  "  - url: https://powerfulcontrolsystem.com",
  "  - url: https://staging.powerfulcontrolsystem.com",
  "paths:",
];

for (const route of routes) {
  const tag = route.startsWith("/super/") ? "super-administrador"
    : route.startsWith("/api/empresa/") ? "empresa"
    : route.startsWith("/auth/") ? "autenticacion"
    : "publico";
  lines.push(`  ${yamlString(route)}:`);
  lines.push("    get:");
  lines.push(`      tags: [${tag}]`);
  lines.push(`      summary: Ruta inventariada ${route}`);
  lines.push("      responses:");
  lines.push("        \"200\":");
  lines.push("          description: Respuesta exitosa o manejada por el handler real.");
  lines.push("    post:");
  lines.push(`      tags: [${tag}]`);
  lines.push(`      summary: Ruta inventariada ${route}`);
  lines.push("      responses:");
  lines.push("        \"200\":");
  lines.push("          description: Respuesta exitosa o manejada por el handler real.");
}

const next = lines.join("\n") + "\n";
if (checkOnly && fs.existsSync(outFile)) {
  const current = fs.readFileSync(outFile, "utf8");
  if (current !== next) {
    console.error(`OpenAPI desactualizado: ${path.relative(repoRoot, outFile)}`);
    process.exit(2);
  }
  console.log(`OpenAPI actualizado: ${path.relative(repoRoot, outFile)}`);
} else {
  fs.mkdirSync(path.dirname(outFile), { recursive: true });
  fs.writeFileSync(outFile, next, "utf8");
  console.log(`OpenAPI generado: ${path.relative(repoRoot, outFile)} (${routes.length} rutas)`);
}
