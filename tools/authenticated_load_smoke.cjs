#!/usr/bin/env node
"use strict";

// Smoke autenticado de solo lectura. Requiere un destino, empresa y rutas GET
// explícitos para impedir que una prueba de capacidad alcance producción o
// ejecute operaciones de negocio por accidente.
const fs = require("fs");
const path = require("path");
const Module = require("module");
const { performance } = require("perf_hooks");

function loadPlaywright() {
  const candidates = [process.env.NODE_PATH, process.env.USERPROFILE && path.join(process.env.USERPROFILE, ".cache", "codex-runtimes", "codex-primary-runtime", "dependencies", "node", "node_modules")].filter(Boolean);
  for (const candidate of candidates) {
    if (!fs.existsSync(path.join(candidate, "playwright"))) continue;
    process.env.NODE_PATH = process.env.NODE_PATH ? process.env.NODE_PATH + path.delimiter + candidate : candidate;
    Module._initPaths();
    return require("playwright");
  }
  throw new Error("No se encontro Playwright en el runtime Codex; no instale dependencias para esta prueba.");
}

const { chromium } = loadPlaywright();
const baseURL = (process.env.PCS_AUTH_LOAD_BASE_URL || "").replace(/\/+$/, "");
const email = process.env.PCS_AUTH_LOAD_EMAIL || "";
const password = process.env.PCS_AUTH_LOAD_PASSWORD || "";
const empresaID = process.env.PCS_AUTH_LOAD_EMPRESA_ID || "";
const routes = (process.env.PCS_AUTH_LOAD_PATHS || "").split(",").map((item) => item.trim()).filter(Boolean);
const concurrency = Number(process.env.PCS_AUTH_LOAD_CONCURRENCY || "5");
const requests = Number(process.env.PCS_AUTH_LOAD_REQUESTS || "30");
const thresholdMS = Number(process.env.PCS_AUTH_LOAD_P95_THRESHOLD_MS || "2500");
const maxErrorRate = Number(process.env.PCS_AUTH_LOAD_MAX_ERROR_RATE || "0.05");
const chromePath = process.env.PCS_AUTH_LOAD_CHROME_PATH || "";
const outDir = process.env.PCS_AUTH_LOAD_OUT_DIR || path.join(process.cwd(), "test_runs", "authenticated_load_" + new Date().toISOString().replace(/[:.]/g, "-").slice(0, 19));

function ensureConfiguration() {
  if (!/^https?:\/\//.test(baseURL)) throw new Error("Defina PCS_AUTH_LOAD_BASE_URL explícitamente.");
  if (!email || !password) throw new Error("Defina PCS_AUTH_LOAD_EMAIL y PCS_AUTH_LOAD_PASSWORD solo en variables de entorno.");
  if (!/^\d+$/.test(empresaID) || Number(empresaID) <= 0) throw new Error("Defina PCS_AUTH_LOAD_EMPRESA_ID con una empresa autorizada.");
  if (!routes.length || routes.some((route) => !route.startsWith("/") || /[\r\n]/.test(route))) throw new Error("Defina PCS_AUTH_LOAD_PATHS con rutas GET relativas explícitas.");
  if (!Number.isInteger(concurrency) || concurrency < 1 || concurrency > 20 || !Number.isInteger(requests) || requests < 1 || requests > 500) throw new Error("Concurrencia debe ser 1..20 y solicitudes 1..500.");
}

async function login(page) {
  await page.goto(baseURL + "/login.html", { waitUntil: "domcontentloaded", timeout: 45000 });
  await page.locator("#adminEmail").fill(email, { timeout: 15000 });
  await page.locator("#adminPassword").fill(password, { timeout: 15000 });
  await Promise.all([page.waitForLoadState("networkidle", { timeout: 45000 }).catch(() => null), page.locator("#emailLoginBtn").click({ timeout: 15000 })]);
  const session = await page.request.get(baseURL + "/me", { timeout: 15000 });
  if (!session.ok()) throw new Error("No se pudo validar la sesión autenticada.");
}

function percentile(values, ratio) {
  const sorted = values.slice().sort((a, b) => a - b);
  return Math.round(sorted[Math.max(0, Math.ceil(sorted.length * ratio) - 1)] || 0);
}

async function main() {
  ensureConfiguration();
  const browser = await chromium.launch({ headless: true, ...(chromePath ? { executablePath: chromePath } : {}) });
  try {
    const context = await browser.newContext({ ignoreHTTPSErrors: true });
    const page = await context.newPage();
    await login(page);
    const samples = [];
    let cursor = 0;
    async function worker() {
      while (cursor < requests) {
        const index = cursor++;
        const route = routes[index % routes.length];
        const started = performance.now();
        try {
          const response = await page.request.get(baseURL + route, { timeout: 30000 });
          samples.push({ route, status: response.status(), elapsedMS: Math.round(performance.now() - started) });
        } catch (error) {
          samples.push({ route, status: 0, elapsedMS: Math.round(performance.now() - started), error: "request failed" });
        }
      }
    }
    await Promise.all(Array.from({ length: concurrency }, () => worker()));
    const failures = samples.filter((sample) => sample.status < 200 || sample.status >= 400).length;
    const statusCounts = samples.reduce((counts, sample) => {
      const key = String(sample.status);
      counts[key] = (counts[key] || 0) + 1;
      return counts;
    }, {});
    const report = {
      generated_at: new Date().toISOString(), base_url: baseURL, empresa_id: Number(empresaID), routes,
      concurrency, requests, p50_ms: percentile(samples.map((sample) => sample.elapsedMS), 0.5),
      p95_ms: percentile(samples.map((sample) => sample.elapsedMS), 0.95),
      p99_ms: percentile(samples.map((sample) => sample.elapsedMS), 0.99), failures,
      error_rate: Number((failures / samples.length).toFixed(4)), status_counts: statusCounts,
      status: failures / samples.length <= maxErrorRate && percentile(samples.map((sample) => sample.elapsedMS), 0.95) <= thresholdMS ? "ok" : "warning",
      samples: samples.slice(0, 40)
    };
    fs.mkdirSync(outDir, { recursive: true });
    fs.writeFileSync(path.join(outDir, "authenticated_load_summary.json"), JSON.stringify(report, null, 2), "utf8");
    process.stdout.write(JSON.stringify({ status: report.status, out_dir: outDir, requests: report.requests, p50_ms: report.p50_ms, p95_ms: report.p95_ms, p99_ms: report.p99_ms, failures: report.failures, error_rate: report.error_rate, status_counts: report.status_counts }) + "\n");
    await context.close();
    if (report.status !== "ok") process.exitCode = 2;
  } finally {
    await browser.close();
  }
}

main().catch((error) => { console.error(String(error.message || error)); process.exit(1); });
