#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";

const repoRoot = process.cwd();
const outputPath = path.join(repoRoot, "documentos", "arquitectura", "inventario_runtime_ensure.md");
const checkOnly = process.argv.includes("--check");

function normalizeLineEndings(value) {
  return String(value).replace(/\r\n/g, "\n").replace(/\r/g, "\n");
}

function compareText(left, right) {
  return left < right ? -1 : left > right ? 1 : 0;
}

function walk(relativeDir) {
  const root = path.join(repoRoot, relativeDir);
  const files = [];
  const pending = [root];
  while (pending.length) {
    const current = pending.pop();
    for (const entry of fs.readdirSync(current, { withFileTypes: true })) {
      const full = path.join(current, entry.name);
      if (entry.isDirectory()) pending.push(full);
      else if (entry.isFile() && entry.name.endsWith(".go") && !entry.name.endsWith("_test.go")) files.push(full);
    }
  }
  return files.sort(compareText);
}

function riskFor(relativePath) {
  if (relativePath === "backend/main.go") return "bootstrap legado; solo autoridad migrate en produccion";
  if (relativePath.startsWith("backend/handlers/")) return "trafico HTTP; priorizar reemplazo por verificacion";
  if (relativePath.startsWith("backend/cmd/pcs-migrate/")) return "migrador dedicado; autoridad de esquema permitida";
  if (relativePath.startsWith("backend/cmd/")) return "proceso de plataforma; revisar rol";
  return "servicio interno; revisar rol y transaccion";
}

const entries = [];
for (const fullPath of [...walk("backend/handlers"), ...walk("backend/internal"), ...walk("backend/cmd"), path.join(repoRoot, "backend", "main.go")]) {
  const source = fs.readFileSync(fullPath, "utf8");
  const relativePath = path.relative(repoRoot, fullPath).replace(/\\/g, "/");
  const pattern = /(?:dbpkg\.)?(Ensure[A-Za-z0-9_]+)\s*\(/g;
  for (const match of source.matchAll(pattern)) {
    const before = source.slice(Math.max(0, (match.index ?? 0) - 12), match.index ?? 0);
    if (/func\s+$/.test(before)) continue;
    entries.push({
      name: match[1],
      path: relativePath,
      line: source.slice(0, match.index ?? 0).split("\n").length,
      risk: riskFor(relativePath),
    });
  }
}
entries.sort((a, b) => compareText(a.path, b.path) || a.line - b.line);
const byRisk = new Map();
for (const entry of entries) byRisk.set(entry.risk, (byRisk.get(entry.risk) ?? 0) + 1);

const lines = [
  "# Inventario de llamadas Ensure en procesos ejecutables",
  "",
  "Estado: generado. Actualizar con `node tools/runtime_ensure_inventory.mjs`.",
  "",
  "Las llamadas listadas permiten vigilar la autoridad de esquema. En produccion, API y worker deben llegar a verificar esquema versionado, no crear o alterar tablas. Las llamadas de `backend/main.go` estan dentro de `LegacySchemaBootstrap`: `runtimeconfig` solo permite activarlas con rol `migrate` y una decision explicita. El binario `pcs-migrate` es la autoridad dedicada.",
  "",
  "## Resumen",
  "",
  `- Llamadas inventariadas: ${entries.length}.`,
  ...[...byRisk.entries()].sort(([a], [b]) => compareText(a, b)).map(([risk, count]) => `- ${risk}: ${count}.`),
  "",
  "## Registro",
  "",
  "| Funcion Ensure | Llamador | Riesgo / prioridad |",
  "| --- | --- | --- |",
  ...entries.map((entry) => `| \`${entry.name}\` | [${entry.path}:${entry.line}](../../${entry.path}#L${entry.line}) | ${entry.risk} |`),
  "",
  "## Gate de retiro",
  "",
  "1. No agregar nuevas filas: el preflight exige que este inventario coincida con el codigo.",
  "2. El conteo aceptable de trafico HTTP es cero; cualquier nueva fila HTTP bloquea el preflight.",
  "3. Cada extraccion debe incluir prueba de base actualizada y de esquema faltante que falle cerrado, sin DDL desde la solicitud.",
  "4. Retirar gradualmente el bootstrap legado de `backend/main.go` cuando el catalogo inmutable cubra cada esquema; `pcs-migrate` conserva la autoridad de esquema.",
];

const httpEnsureCount = entries.filter((entry) => entry.path.startsWith("backend/handlers/")).length;
if (httpEnsureCount !== 0) {
  console.error(`bootstrap Ensure detectado en trafico HTTP: ${httpEnsureCount}`);
  process.exitCode = 3;
}
const rendered = `${lines.join("\n")}\n`;

if (checkOnly) {
  const current = fs.existsSync(outputPath) ? fs.readFileSync(outputPath, "utf8") : "";
  if (normalizeLineEndings(current) !== rendered) {
    console.error("inventario runtime Ensure desactualizado; ejecuta node tools/runtime_ensure_inventory.mjs");
    process.exitCode = 2;
  } else {
    console.log(`inventario runtime Ensure vigente: ${entries.length} llamadas`);
  }
} else {
  fs.mkdirSync(path.dirname(outputPath), { recursive: true });
  fs.writeFileSync(outputPath, rendered, "utf8");
  console.log(`inventario runtime Ensure generado: ${entries.length} llamadas`);
}
