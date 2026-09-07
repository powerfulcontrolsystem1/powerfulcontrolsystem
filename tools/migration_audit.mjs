#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";
import { execFileSync } from "node:child_process";

const repoRoot = process.cwd();
const outArgIndex = process.argv.indexOf("--out");
const strict = process.argv.includes("--strict");
const baseRefArgIndex = process.argv.indexOf("--base-ref");
const baseRef = baseRefArgIndex >= 0 && process.argv[baseRefArgIndex + 1]
  ? process.argv[baseRefArgIndex + 1]
  : "origin/main";
const outDir = outArgIndex >= 0 && process.argv[outArgIndex + 1]
  ? path.resolve(repoRoot, process.argv[outArgIndex + 1])
  : path.join(repoRoot, "documentos", "reportes_profesionales");

function walk(dir) {
  const root = path.join(repoRoot, dir);
  const out = [];
  const stack = [root];
  while (stack.length) {
    const current = stack.pop();
    for (const entry of fs.readdirSync(current, { withFileTypes: true })) {
      const full = path.join(current, entry.name);
      if (entry.isDirectory()) stack.push(full);
      else if (entry.isFile() && entry.name.endsWith(".go")) out.push(full);
    }
  }
  return out;
}

const files = walk("backend/db");
const schemaFiles = files.filter((file) => /CREATE TABLE|ALTER TABLE|CREATE INDEX|DROP TABLE/i.test(fs.readFileSync(file, "utf8")));
const migrationSource = fs.readFileSync(path.join(repoRoot, "backend/db/migrations.go"), "utf8");
const workerSource = fs.readFileSync(path.join(repoRoot, "backend/cmd/pcs-worker/main.go"), "utf8");
const inventoryPath = path.join(repoRoot, "documentos", "arquitectura", "inventario_bootstrap_ensure.md");
const hasMigrationTable = /schema_migrations/.test(migrationSource) && /RunMigrations/.test(migrationSource) && /pg_advisory_xact_lock/.test(migrationSource) && /MigrationChecksum/.test(migrationSource);
const workerCreatesSchema = /Ensure(?:AsyncJobs|Outbox)Schema\s*\(/.test(workerSource);
const inventoryPresent = fs.existsSync(inventoryPath) && fs.readFileSync(inventoryPath, "utf8").includes("Inventario de bootstrap Ensure");
const tests = files.filter((file) => file.endsWith("_test.go")).length;

function git(args, { allowFailure = false } = {}) {
  try {
    return execFileSync("git", args, { cwd: repoRoot, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"] }).trim();
  } catch (error) {
    if (allowFailure) return "";
    const detail = String(error?.stderr ?? error?.message ?? "git fallo").trim();
    throw new Error(`git ${args.join(" ")} fallo: ${detail}`);
  }
}

function normalizeExpression(value) {
  return String(value).replace(/\s+/g, " ").trim();
}

function catalogEntries(source) {
  const targets = [
    { name: "empresas", start: source.indexOf("case MigrationTargetEmpresas:"), end: source.indexOf("case MigrationTargetSuper:") },
    { name: "superadministrador", start: source.indexOf("case MigrationTargetSuper:"), end: source.indexOf("default:", source.indexOf("case MigrationTargetSuper:")) },
  ];
  const entries = [];
  for (const target of targets) {
    if (target.start < 0 || target.end < 0 || target.end <= target.start) continue;
    const section = source.slice(target.start, target.end);
    const pattern = /Version:\s*("[^"]+"|[A-Za-z0-9_]+)\s*,[\s\S]{0,500}?Description:\s*"([^"]*)"\s*,[\s\S]{0,500}?Body:\s*([^,\r\n}]+)/g;
    for (const match of section.matchAll(pattern)) {
      entries.push({
        target: target.name,
        version: normalizeExpression(match[1]),
        description: match[2],
        body: normalizeExpression(match[3]),
      });
    }
  }
  return entries;
}

function historicalMigrationChecks() {
  const baseExists = git(["rev-parse", "--verify", baseRef], { allowFailure: true });
  if (!baseExists) {
    return { base_ref: baseRef, base_available: false, historical_entries_immutable: false, modified_historical_files: [], details: [`no existe ${baseRef}`] };
  }
  const catalogPath = "backend/db/platform_migrations.go";
  const baseSource = git(["show", `${baseRef}:${catalogPath}`]);
  const currentSource = fs.readFileSync(path.join(repoRoot, catalogPath), "utf8");
  const baseEntries = catalogEntries(baseSource);
  const currentEntries = catalogEntries(currentSource);
  const currentByKey = new Map(currentEntries.map((entry) => [`${entry.target}:${entry.version}`, entry]));
  const changedEntries = [];
  for (const entry of baseEntries) {
    const key = `${entry.target}:${entry.version}`;
    const current = currentByKey.get(key);
    if (!current) {
      changedEntries.push(`${key}: eliminada`);
    } else if (current.description !== entry.description || current.body !== entry.body) {
      changedEntries.push(`${key}: descripcion o Body modificado`);
    }
  }
  const changedRows = git(["diff", "--name-status", baseRef, "--", "backend/db"], { allowFailure: true })
    .split(/\r?\n/)
    .filter(Boolean);
  const modifiedHistoricalFiles = changedRows.filter((row) => {
    const [status, file] = row.split(/\s+/, 2);
    if (!file || status === "A" || file.endsWith("_test.go") || file === catalogPath) return false;
    return /(?:^|\/)(?:migrations\.go|[^/]*migration[^/]*\.go)$/i.test(file);
  });
  return {
    base_ref: baseRef,
    base_available: true,
    historical_entries_immutable: changedEntries.length === 0 && modifiedHistoricalFiles.length === 0,
    base_entries: baseEntries.length,
    current_entries: currentEntries.length,
    new_entries: Math.max(0, currentEntries.length - baseEntries.length),
    modified_historical_files: modifiedHistoricalFiles,
    details: changedEntries,
  };
}

const historical = historicalMigrationChecks();
const checksOk = hasMigrationTable && !workerCreatesSchema && inventoryPresent && tests >= 20 && historical.base_available && historical.historical_entries_immutable;

const report = {
  generated_at: new Date().toISOString(),
  status: checksOk ? "ok" : "warning",
  base_ref: baseRef,
  checks: [
    { name: "checksummed_locked_migration_runner", ok: hasMigrationTable },
    { name: "worker_has_no_schema_ddl_calls", ok: !workerCreatesSchema },
    { name: "ensure_bootstrap_inventory", ok: inventoryPresent },
    { name: "schema_touching_files", ok: schemaFiles.length > 0, count: schemaFiles.length, examples: schemaFiles.slice(0, 25).map((file) => path.relative(repoRoot, file).replace(/\\/g, "/")) },
    { name: "db_tests_present", ok: tests >= 20, count: tests },
    { name: "migration_base_ref_available", ok: historical.base_available, base_ref: baseRef },
    { name: "historical_migration_entries_immutable", ok: historical.historical_entries_immutable, ...historical },
  ],
};

fs.mkdirSync(outDir, { recursive: true });
const stamp = report.generated_at.replace(/[-:]/g, "").replace(/\..+$/, "").replace("T", "_");
const jsonPath = path.join(outDir, `migration_audit_${stamp}.json`);
const mdPath = path.join(outDir, `migration_audit_${stamp}.md`);
fs.writeFileSync(jsonPath, JSON.stringify(report, null, 2), "utf8");
fs.writeFileSync(mdPath, `# Auditoria de migraciones\n\nEstado: ${report.status}\n\n\`\`\`json\n${JSON.stringify(report, null, 2)}\n\`\`\`\n`, "utf8");
console.log(JSON.stringify({ status: report.status, json: jsonPath, markdown: mdPath }, null, 2));
if (strict && report.status !== "ok") process.exitCode = 2;
