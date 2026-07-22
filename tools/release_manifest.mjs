#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";
import { execFileSync } from "node:child_process";

const repoRoot = process.cwd();
const versionArg = process.argv.find((arg) => arg.startsWith("--version="));
const baseRefArg = process.argv.find((arg) => arg.startsWith("--base-ref="));
const version = versionArg ? versionArg.split("=")[1].trim() : new Date().toISOString().slice(0, 10).replace(/-/g, ".");
const baseRef = baseRefArg ? baseRefArg.split("=")[1].trim() : "origin/main";
const checkOnly = process.argv.includes("--check");
const strict = process.argv.includes("--strict");
const outDir = path.join(repoRoot, "documentos", "releases");
const digestPattern = /^[^@\s]+@sha256:[a-fA-F0-9]{64}$/;

function git(args, fallback = "") {
  try {
    return execFileSync("git", args, { cwd: repoRoot, encoding: "utf8" }).trim();
  } catch {
    return fallback;
  }
}

function gitSucceeds(args) {
  try {
    execFileSync("git", args, { cwd: repoRoot, stdio: "ignore" });
    return true;
  } catch {
    return false;
  }
}

function imageDigestStatus(name) {
  const value = (process.env[name] || "").trim();
  return {
    present: value !== "",
    valid: digestPattern.test(value),
    reference: digestPattern.test(value) ? value : "",
  };
}

const branch = git(["branch", "--show-current"], "unknown");
const commit = git(["rev-parse", "HEAD"], "unknown");
const status = git(["status", "--short"], "");
const recent = git(["log", "--oneline", "-10"], "").split(/\r?\n/).filter(Boolean);
const generatedAt = new Date().toISOString();
const baseCommit = git(["rev-parse", "--verify", baseRef], "");
const upstream = git(["rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}"], "");
const upstreamCommit = upstream ? git(["rev-parse", "--verify", upstream], "") : "";
const baseIsAncestor = baseCommit !== "" && gitSucceeds(["merge-base", "--is-ancestor", baseRef, "HEAD"]);
const imageDigests = {
  PCS_API_IMAGE_DIGEST: imageDigestStatus("PCS_API_IMAGE_DIGEST"),
  PCS_MIGRATE_IMAGE_DIGEST: imageDigestStatus("PCS_MIGRATE_IMAGE_DIGEST"),
  PCS_WORKER_IMAGE_DIGEST: imageDigestStatus("PCS_WORKER_IMAGE_DIGEST"),
};
const releaseGaps = [];
if (status.trim() !== "") releaseGaps.push("working_tree_dirty");
if (!baseCommit) releaseGaps.push("base_ref_unresolved");
if (!baseIsAncestor) releaseGaps.push("candidate_not_based_on_base_ref");
if (!upstream) releaseGaps.push("branch_without_upstream");
for (const [name, item] of Object.entries(imageDigests)) {
  if (!item.present) releaseGaps.push(`${name}_missing`);
  else if (!item.valid) releaseGaps.push(`${name}_invalid`);
}

const manifest = {
  version,
  generated_at: generatedAt,
  branch,
  commit,
  working_tree_clean: status.trim() === "",
  candidate: {
    base_ref: baseRef,
    base_commit: baseCommit,
    base_is_ancestor: baseIsAncestor,
    upstream: upstream || null,
    upstream_commit: upstreamCommit || null,
    image_digests: imageDigests,
    release_gaps: releaseGaps,
  },
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

if (checkOnly) {
  console.log(JSON.stringify(manifest, null, 2));
  if (strict && releaseGaps.length > 0) process.exitCode = 1;
} else {
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
    `Base: ${baseRef}${baseCommit ? ` (${baseCommit})` : " (no resuelta)"}`,
    `Base ancestro del candidato: ${baseIsAncestor ? "si" : "no"}`,
    `Upstream: ${upstream || "no configurado"}`,
    `Bloqueos de release: ${releaseGaps.length ? releaseGaps.join(", ") : "ninguno"}`,
    "",
    "## Imagenes inmutables",
    "",
    ...Object.entries(imageDigests).map(([name, item]) => `- ${name}: ${item.valid ? item.reference : item.present ? "invalida" : "pendiente"}`),
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

  console.log(JSON.stringify({ json: jsonPath, markdown: mdPath, version, commit, release_gaps: releaseGaps }, null, 2));
  if (strict && releaseGaps.length > 0) process.exitCode = 1;
}
