#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";
import { performance } from "node:perf_hooks";

const repoRoot = process.cwd();
const baseUrl = (process.env.PCS_LOAD_BASE_URL || process.argv.find((arg) => arg.startsWith("--base-url="))?.split("=")[1] || "https://staging.powerfulcontrolsystem.com").replace(/\/+$/, "");
const concurrency = Number(process.env.PCS_LOAD_CONCURRENCY || "20");
const requests = Number(process.env.PCS_LOAD_REQUESTS || "120");
const thresholdMs = Number(process.env.PCS_LOAD_P95_THRESHOLD_MS || "2500");
const maxErrorRate = Number(process.env.PCS_LOAD_MAX_ERROR_RATE || "0.05");
const outDir = path.join(repoRoot, "documentos", "reportes_profesionales");
const paths = (process.env.PCS_LOAD_PATHS || "/,/login.html,/elegir_licencia.html,/pagar_licencia.html")
  .split(",")
  .map((item) => item.trim())
  .filter(Boolean);

async function hit(url) {
  const started = performance.now();
  try {
    const res = await fetch(url, { redirect: "manual" });
    const elapsed = performance.now() - started;
    return { ok: res.status < 500, status: res.status, elapsed };
  } catch (error) {
    return { ok: false, status: 0, elapsed: performance.now() - started, error: String(error.message || error) };
  }
}

const results = [];
let cursor = 0;
async function worker() {
  while (cursor < requests) {
    const current = cursor++;
    const route = paths[current % paths.length];
    results.push(await hit(baseUrl + route));
  }
}

await Promise.all(Array.from({ length: concurrency }, () => worker()));
const sorted = results.map((item) => item.elapsed).sort((a, b) => a - b);
const p95 = sorted[Math.max(0, Math.ceil(sorted.length * 0.95) - 1)] || 0;
const errors = results.filter((item) => !item.ok).length;
const errorRate = results.length ? errors / results.length : 1;
const report = {
  generated_at: new Date().toISOString(),
  base_url: baseUrl,
  paths,
  concurrency,
  requests,
  p95_ms: Math.round(p95),
  errors,
  error_rate: Number(errorRate.toFixed(4)),
  status: p95 <= thresholdMs && errorRate <= maxErrorRate ? "ok" : "warning",
};

fs.mkdirSync(outDir, { recursive: true });
const stamp = report.generated_at.replace(/[-:]/g, "").replace(/\..+$/, "").replace("T", "_");
const jsonPath = path.join(outDir, `load_smoke_${stamp}.json`);
const mdPath = path.join(outDir, `load_smoke_${stamp}.md`);
fs.writeFileSync(jsonPath, JSON.stringify({ ...report, samples: results.slice(0, 20) }, null, 2), "utf8");
fs.writeFileSync(mdPath, [
  "# Prueba de carga smoke",
  "",
  `Fecha: ${report.generated_at}`,
  `Estado: ${report.status}`,
  `Base URL: ${report.base_url}`,
  `Concurrencia: ${report.concurrency}`,
  `Requests: ${report.requests}`,
  `P95 ms: ${report.p95_ms}`,
  `Errores: ${report.errors}`,
  `Error rate: ${report.error_rate}`,
  "",
].join("\n"), "utf8");

console.log(JSON.stringify({ status: report.status, json: jsonPath, markdown: mdPath, p95_ms: report.p95_ms, error_rate: report.error_rate }, null, 2));
if (process.argv.includes("--strict") && report.status !== "ok") process.exitCode = 2;
