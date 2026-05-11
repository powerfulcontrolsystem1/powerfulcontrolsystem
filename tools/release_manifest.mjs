#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";
import { execFileSync } from "node:child_process";

const repoRoot = process.cwd();
const versionArg = process.argv.find((arg) => arg.startsWith("--version="));
const version = versionArg ? versionArg.split("=")[1].trim() : new Date().toISOString().slice(0, 10).replace(/-/g, ".");
const outDir = path.join(repoRoot, "documentos", "releases");

function git(args, fallback = "") {
  try {
    return execFileSync("git", args, { cwd: repoRoot, encoding: "utf8" }).trim();
  } catch {
    return fallback;
  }
}

const branch = git(["branch", "--show-current"], "unknown");
const commit = git(["rev-parse", "--short", "HEAD"], "unknown");
const status = git(["status", "--short"], "");
const recent = git(["log", "--oneline", "-10"], "").split(/\r?\n/).filter(Boolean);
const generatedAt = new Date().toISOString();

const manifest = {
  version,
  generated_at: generatedAt,
  branch,
  commit,
  working_tree_clean: status.trim() === "",
  recent_commits: recent,
  required_checks: [
    "scripts/profesional_preflight.ps1 -Full",
    "scripts/vps_backup_operacion.ps1",
    "scripts/vps_restore_validation.ps1 -ExecuteDrill",
    "tools/qa_e2e_buttons.cjs against staging",
    "tools/qa_print_formats.cjs",
    "tools/load_smoke_test.mjs against staging"
  ],
};

fs.mkdirSync(outDir, { recursive: true });
const safe = version.replace(/[^\w.-]+/g, "_");
const jsonPath = path.join(outDir, `release_${safe}.json`);
const mdPath = path.join(outDir, `release_${safe}.md`);
fs.writeFileSync(jsonPath, JSON.stringify(manifest, null, 2), "utf8");
fs.writeFileSync(mdPath, [
  `# Release ${version}`,
  "",
  `Fecha: ${generatedAt}`,
  `Rama: ${branch}`,
  `Commit: ${commit}`,
  `Working tree limpio: ${manifest.working_tree_clean ? "si" : "no"}`,
  "",
  "## Checks requeridos",
  "",
  ...manifest.required_checks.map((item) => `- ${item}`),
  "",
  "## Commits recientes",
  "",
  ...recent.map((item) => `- ${item}`),
  "",
].join("\n"), "utf8");

console.log(JSON.stringify({ json: jsonPath, markdown: mdPath, version, commit }, null, 2));
