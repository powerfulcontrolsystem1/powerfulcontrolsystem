#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";

const repoRoot = process.cwd();
const outputPath = path.join(repoRoot, "documentos", "arquitectura", "inventario_bootstrap_ensure.md");
const checkOnly = process.argv.includes("--check");

function walk(relativeDir) {
  const result = [];
  const root = path.join(repoRoot, relativeDir);
  const stack = [root];
  while (stack.length) {
    const current = stack.pop();
    for (const entry of fs.readdirSync(current, { withFileTypes: true })) {
      const full = path.join(current, entry.name);
      if (entry.isDirectory()) stack.push(full);
      else if (entry.isFile() && entry.name.endsWith(".go") && !entry.name.endsWith("_test.go")) result.push(full);
    }
  }
  return result.sort();
}

function functionBody(source, offset) {
  const open = source.indexOf("{", offset);
  if (open < 0) return "";
  let depth = 0;
  for (let i = open; i < source.length; i += 1) {
    if (source[i] === "{") depth += 1;
    else if (source[i] === "}") {
      depth -= 1;
      if (depth === 0) return source.slice(open, i + 1);
    }
  }
  return source.slice(open);
}

function classify(name, body, relativePath) {
  if (/^(EnsureAsyncJobsSchema|EnsureOutboxSchema|EnsureMobileAPIIdempotencySchema)$/.test(name)) return "DDL catalogado de plataforma";
  if (/PostgresRuntimeCompat|PrimaryKeySequences|Compatibility|Compat/i.test(name)) return "compatibilidad PostgreSQL";
  if (/CREATE\s+(?:OR\s+REPLACE\s+)?(?:TABLE|INDEX|FUNCTION)|ALTER\s+TABLE|DROP\s+TABLE/i.test(body)) return "DDL / indice / funcion";
  if (/Seed|Default|Provision|Assignment|RowsForExisting|PowerfulSystem|TipoEmpresa|Catalogo/i.test(name) || /\bINSERT\s+INTO\b/i.test(body)) return "seed o provisionamiento idempotente";
  if (/Schema/i.test(name)) return "posible DDL, requiere extraccion";
  if (/handlers\//.test(relativePath)) return "provisionamiento de integracion";
  return "regla auxiliar o verificacion";
}

function target(name, relativePath) {
  if (name === "EnsureAsyncJobsSchema" || name === "EnsureOutboxSchema") return "superadministrador";
  if (name === "EnsureMobileAPIIdempotencySchema") return "empresas";
  if (/super|administrador|licencia|paymentgateway|contrato/i.test(`${name} ${relativePath}`)) return "superadministrador o por confirmar";
  if (/main\.go$/.test(relativePath)) return "ambas bases";
  return "empresas o por confirmar";
}

const files = [...walk("backend/db"), ...walk("backend/handlers"), path.join(repoRoot, "backend", "main.go")];
const entries = [];
for (const file of files) {
  const source = fs.readFileSync(file, "utf8");
  const relative = path.relative(repoRoot, file).replace(/\\/g, "/");
  const pattern = /func\s+(?:\([^)]*\)\s*)?(Ensure[A-Za-z0-9_]+)\s*\(/g;
  for (const match of source.matchAll(pattern)) {
    const body = functionBody(source, match.index ?? 0);
    const line = source.slice(0, match.index).split("\n").length;
    entries.push({
      name: match[1],
      path: relative,
      line,
      classification: classify(match[1], body, relative),
      target: target(match[1], relative),
    });
  }
}
entries.sort((a, b) => a.path.localeCompare(b.path) || a.line - b.line);
const counts = new Map();
for (const entry of entries) counts.set(entry.classification, (counts.get(entry.classification) ?? 0) + 1);

const lines = [
  "# Inventario de bootstrap Ensure",
  "",
  "Estado: generado. Ultima actualizacion: 2026-07-16.",
  "",
  "Este archivo se genera con `node tools/ensure_bootstrap_inventory.mjs`. Inventaria las funciones `Ensure*` de backend y es la base obligatoria para retirar el bootstrap historico. Una clasificacion `por confirmar` no autoriza desactivar `PCS_RUNTIME_SCHEMA_BOOTSTRAP`; debe convertirse en una migracion catalogada, seed programado o verificacion sin DDL.",
  "",
  "## Resumen",
  "",
  `- Funciones inventariadas: ${entries.length}.`,
  ...[...counts.entries()].sort(([a], [b]) => a.localeCompare(b)).map(([key, value]) => `- ${key}: ${value}.`),
  "- Fuente: `backend/db`, `backend/handlers` y `backend/main.go`; excluye pruebas.",
  "",
  "## Registro",
  "",
  "| Funcion | Archivo | Clase inferida | Base objetivo inferida |",
  "| --- | --- | --- | --- |",
  ...entries.map((entry) => `| \`${entry.name}\` | [${entry.path}:${entry.line}](../../${entry.path}#L${entry.line}) | ${entry.classification} | ${entry.target} |`),
  "",
  "## Gate de retiro",
  "",
  "1. Catalogar cada fila DDL en `db.PlatformMigrations` o una migracion de dominio equivalente con checksum.",
  "2. Mover seeds/provisionamientos a jobs versionados y explicitos, no al arranque de API.",
  "3. Repetir migraciones en staging, comparar esquema y ejecutar pruebas operativas antes de cambiar `PCS_RUNTIME_SCHEMA_BOOTSTRAP=0`.",
  "4. Mantener este inventario sincronizado mediante el preflight.",
];
const rendered = `${lines.join("\n")}\n`;
if (checkOnly) {
  const current = fs.existsSync(outputPath) ? fs.readFileSync(outputPath, "utf8") : "";
  if (current !== rendered) {
    console.error("inventario Ensure desactualizado; ejecuta node tools/ensure_bootstrap_inventory.mjs");
    process.exitCode = 2;
  } else {
    console.log(`inventario Ensure vigente: ${entries.length} funciones`);
  }
} else {
  fs.mkdirSync(path.dirname(outputPath), { recursive: true });
  fs.writeFileSync(outputPath, rendered, "utf8");
  console.log(`inventario Ensure generado: ${entries.length} funciones`);
}
