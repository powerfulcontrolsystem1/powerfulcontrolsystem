#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";

const repoRoot = process.cwd();
const outputPath = path.join(repoRoot, "documentos", "arquitectura", "matriz_rutas_multiempresa.md");
const checkOnly = process.argv.includes("--check");

function walk(relativeDir) {
  const root = path.join(repoRoot, relativeDir);
  const files = [];
  const pending = [root];
  while (pending.length) {
    const current = pending.pop();
    for (const entry of fs.readdirSync(current, { withFileTypes: true })) {
      const full = path.join(current, entry.name);
      if (entry.isDirectory()) {
        pending.push(full);
      } else if (entry.isFile() && entry.name.endsWith(".go") && !entry.name.endsWith("_test.go")) {
        files.push(full);
      }
    }
  }
  return files.sort();
}

function sourceLocation(fullPath, source, index) {
  return {
    file: path.relative(repoRoot, fullPath).replace(/\\/g, "/"),
    line: source.slice(0, index).split("\n").length,
  };
}

function callEnd(source, openIndex) {
  let depth = 0;
  let quoted = false;
  let escaped = false;
  for (let index = openIndex; index < source.length; index += 1) {
    const char = source[index];
    if (quoted) {
      if (escaped) escaped = false;
      else if (char === "\\") escaped = true;
      else if (char === '"') quoted = false;
      continue;
    }
    if (char === '"') {
      quoted = true;
      continue;
    }
    if (char === "(") depth += 1;
    if (char === ")") {
      depth -= 1;
      if (depth === 0) return index;
    }
  }
  return -1;
}

const routes = [];
for (const fullPath of walk("backend")) {
  const source = fs.readFileSync(fullPath, "utf8");
  const pattern = /http\.HandleFunc\(\s*"(\/api\/empresa\/[^\"]+)"\s*,/g;
  for (const match of source.matchAll(pattern)) {
    const start = match.index ?? 0;
    const open = source.indexOf("(", start);
    const end = callEnd(source, open);
    if (open < 0 || end < 0) {
      throw new Error(`could not parse route registration ${match[1]} in ${fullPath}`);
    }
    const routeEnd = source.indexOf('"', source.indexOf('"', open) + 1) + 1;
    const comma = source.indexOf(",", routeEnd);
    const expression = source.slice(comma + 1, end).replace(/\s+/g, " ").trim();
    const location = sourceLocation(fullPath, source, match.index ?? 0);
    const wrapperMatch = expression.match(/\b(WithEmpresa[A-Za-z0-9_]*Permissions|WithEmpresaPublicScope|WithEmpresaRolePermissions)\s*\(/);
    routes.push({
      path: match[1],
      ...location,
      wrapper: wrapperMatch?.[1] ?? "SIN_WRAPPER_AUTORITATIVO_DETECTADO",
      status: wrapperMatch ? "protegida" : "requiere revision manual",
    });
  }
}

routes.sort((a, b) => a.path.localeCompare(b.path) || a.file.localeCompare(b.file) || a.line - b.line);
const duplicateRoutes = routes.filter((route, index) => index > 0 && routes[index - 1].path === route.path);
const protectedRoutes = routes.filter((route) => route.status === "protegida").length;
const manualRoutes = routes.length - protectedRoutes;

const lines = [
  "# Matriz de rutas multiempresa",
  "",
  "Estado: generado. Actualizar con `node tools/tenant_route_inventory.mjs`.",
  "",
  "Este inventario detecta registros HTTP bajo `/api/empresa/` y exige que cada uno tenga una evidencia de wrapper autoritativo. Es un control de cobertura, no sustituye las pruebas negativas A/B ni el filtro `empresa_id` en SQL, archivos, cache y jobs.",
  "",
  "## Resumen",
  "",
  `- Rutas empresariales inventariadas: ${routes.length}.`,
  `- Con wrapper autoritativo detectado: ${protectedRoutes}.`,
  `- Requieren revision manual: ${manualRoutes}.`,
  `- Duplicados de ruta detectados: ${duplicateRoutes.length}.`,
  "",
  "## Registro",
  "",
  "| Ruta | Archivo | Wrapper detectado | Estado |",
  "| --- | --- | --- | --- |",
  ...routes.map((route) => `| \`${route.path}\` | [${route.file}:${route.line}](../../${route.file}#L${route.line}) | \`${route.wrapper}\` | ${route.status} |`),
  "",
  "## Gate de cambios",
  "",
  "1. Una ruta nueva bajo `/api/empresa/` debe usar un wrapper que cree `TenantContext` despues de validar sesion, pertenencia, rol y permiso.",
  "2. Una fila `requiere revision manual` bloquea declarar cobertura completa hasta documentar su excepcion o corregirla.",
  "3. El handler debe tomar el `empresa_id` desde `TenantContext`; parametros de URL, JSON o cabecera nunca son fuente de autoridad.",
  "4. Los cambios de lectura, escritura, exportacion, descarga, cache o job requieren prueba negativa entre empresa A y empresa B.",
];
const rendered = `${lines.join("\n")}\n`;

if (checkOnly) {
  const current = fs.existsSync(outputPath) ? fs.readFileSync(outputPath, "utf8") : "";
  if (current !== rendered) {
    console.error("matriz de rutas multiempresa desactualizada; ejecuta node tools/tenant_route_inventory.mjs");
    process.exitCode = 2;
  } else {
    console.log(`matriz multiempresa vigente: ${routes.length} rutas, ${manualRoutes} requieren revision manual`);
  }
} else {
  fs.mkdirSync(path.dirname(outputPath), { recursive: true });
  fs.writeFileSync(outputPath, rendered, "utf8");
  console.log(`matriz multiempresa generada: ${routes.length} rutas, ${manualRoutes} requieren revision manual`);
}
