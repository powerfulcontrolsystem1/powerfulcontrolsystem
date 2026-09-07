#!/usr/bin/env node
import { spawnSync } from "node:child_process";
import fs from "node:fs";
import path from "node:path";

const repoRoot = process.cwd();
const outIndex = process.argv.indexOf("--out");
const outDir = outIndex >= 0 && process.argv[outIndex + 1]
  ? path.resolve(repoRoot, process.argv[outIndex + 1])
  : path.join(repoRoot, "documentos", "reportes_profesionales");
const baseUrlIndex = process.argv.indexOf("--base-url");
const baseUrl = baseUrlIndex >= 0 && process.argv[baseUrlIndex + 1] ? process.argv[baseUrlIndex + 1] : "http://127.0.0.1:8080";
const stages = [
  { name: "baseline", concurrency: 5, requests: 25 },
  { name: "operational", concurrency: 20, requests: 100 },
  { name: "growth", concurrency: 50, requests: 250 }
];

fs.mkdirSync(outDir, { recursive: true });
const results = [];
for (const stage of stages) {
  const args = ["tools/load_smoke_test.mjs", `--base-url=${baseUrl}`];
  const env = {
    ...process.env,
    PCS_LOAD_CONCURRENCY: String(stage.concurrency),
    PCS_LOAD_REQUESTS: String(stage.requests)
  };
  const result = spawnSync(process.execPath, args, { cwd: repoRoot, encoding: "utf8", env });
  results.push({
    ...stage,
    status: result.status === 0 ? "ok" : "warning",
    exit_code: result.status,
    stdout_tail: result.stdout.split(/\r?\n/).slice(-8).join("\n"),
    stderr_tail: result.stderr.split(/\r?\n/).slice(-8).join("\n")
  });
  if (result.status !== 0 && process.argv.includes("--stop-on-fail")) break;
}

const report = {
  generated_at: new Date().toISOString(),
  base_url: baseUrl,
  status: results.every((item) => item.status === "ok") ? "ok" : "warning",
  results
};
const stamp = report.generated_at.replace(/[-:]/g, "").replace(/\..+$/, "").replace("T", "_");
fs.writeFileSync(path.join(outDir, `load_capacity_plan_${stamp}.json`), JSON.stringify(report, null, 2), "utf8");
console.log(JSON.stringify(report, null, 2));
if (process.argv.includes("--strict") && report.status !== "ok") process.exitCode = 2;
