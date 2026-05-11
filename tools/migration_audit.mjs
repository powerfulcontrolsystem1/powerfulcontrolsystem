#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";

const repoRoot = process.cwd();
const outArgIndex = process.argv.indexOf("--out");
const strict = process.argv.includes("--strict");
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
const hasMigrationTable = /schema_migrations/.test(migrationSource) && /ApplySchemaMigration/.test(migrationSource);
const tests = files.filter((file) => file.endsWith("_test.go")).length;

const report = {
  generated_at: new Date().toISOString(),
  status: hasMigrationTable && tests >= 20 ? "ok" : "warning",
  checks: [
    { name: "schema_migrations_table", ok: hasMigrationTable },
    { name: "schema_touching_files", ok: schemaFiles.length > 0, count: schemaFiles.length, examples: schemaFiles.slice(0, 25).map((file) => path.relative(repoRoot, file).replace(/\\/g, "/")) },
    { name: "db_tests_present", ok: tests >= 20, count: tests },
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
