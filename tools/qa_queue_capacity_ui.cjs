#!/usr/bin/env node
"use strict";

const fs = require("fs");
const path = require("path");
const Module = require("module");

function loadBundledPlaywright() {
  const candidates = [];
  if (process.env.NODE_PATH) candidates.push(...process.env.NODE_PATH.split(path.delimiter));
  if (process.env.USERPROFILE) {
    candidates.push(path.join(process.env.USERPROFILE, ".cache", "codex-runtimes", "codex-primary-runtime", "dependencies", "node", "node_modules"));
  }
  for (const candidate of candidates) {
    if (!candidate || !fs.existsSync(path.join(candidate, "playwright"))) continue;
    process.env.NODE_PATH = process.env.NODE_PATH ? process.env.NODE_PATH + path.delimiter + candidate : candidate;
    Module._initPaths();
    return;
  }
}

loadBundledPlaywright();
const { chromium } = require("playwright");
const ROOT = path.resolve(__dirname, "..");
const PAGE = path.join(ROOT, "web", "super", "capacidad_colas.html");
const CSS = path.join(ROOT, "web", "estilos.css");
const OUT = process.env.PCS_QA_QUEUE_OUT_DIR || path.join(ROOT, "test_runs", "qa_queue_capacity_ui_" + new Date().toISOString().replace(/[:.]/g, "-").slice(0, 19));
const CHROME = process.env.PCS_QA_CHROME_EXECUTABLE || "";

const fixture = {
  ok: true,
  configs: [
    { lane: "printing", label: "Impresiones", alerts_enabled: true, rate_limit_per_minute: 120, pending_alert_threshold: 200, oldest_alert_seconds: 120, max_pending_per_tenant: 100 },
    { lane: "product_add", label: "Agregar productos", alerts_enabled: true, rate_limit_per_minute: 240, pending_alert_threshold: 0, oldest_alert_seconds: 0, max_pending_per_tenant: 0 },
    { lane: "fiscal", label: "Emision de facturas", alerts_enabled: true, rate_limit_per_minute: 30, pending_alert_threshold: 100, oldest_alert_seconds: 300, max_pending_per_tenant: 25 }
  ],
  snapshots: [
    { lane: "printing", label: "Impresiones", pending: 42, processing: 8, failed: 1, active_tenants: 21, busiest_tenant_id: 17, busiest_tenant_pending: 9, oldest_seconds: 35, saturation_percent: 29.2, query_ok: true },
    { lane: "product_add", label: "Agregar productos", requests_current_minute: 1730, active_tenants: 96, busiest_tenant_id: 24, busiest_tenant_pending: 188, saturation_percent: 78.3, query_ok: true },
    { lane: "fiscal", label: "Emision de facturas", pending: 124, processing: 12, failed: 3, active_tenants: 48, busiest_tenant_id: 31, busiest_tenant_pending: 18, oldest_seconds: 410, saturation_percent: 136.7, query_ok: true }
  ]
};

async function render(browser, name, viewport) {
  const page = await browser.newPage({ viewport });
  await page.route("**/*", async (route) => {
    const url = new URL(route.request().url());
    if (url.pathname === "/super/api/capacidad_colas") {
      await route.fulfill({ status: 200, contentType: "application/json", body: JSON.stringify(fixture) });
    } else if (url.pathname === "/estilos.css") {
      await route.fulfill({ status: 200, contentType: "text/css", body: fs.readFileSync(CSS, "utf8") });
    } else if (url.pathname === "/super/capacidad_colas.html") {
      await route.fulfill({ status: 200, contentType: "text/html", body: fs.readFileSync(PAGE, "utf8") });
    } else {
      await route.fulfill({ status: 200, contentType: "application/javascript", body: "" });
    }
  });
  await page.goto("http://pcs.local/super/capacidad_colas.html", { waitUntil: "networkidle" });
  await page.waitForSelector(".queue-card:nth-child(3) .queue-badge.err");
  const cardCount = await page.locator(".queue-card").count();
  if (cardCount !== 3) throw new Error("se esperaban tres carriles, got " + cardCount);
  const overflow = await page.evaluate(() => document.documentElement.scrollWidth > document.documentElement.clientWidth + 1);
  if (overflow) throw new Error("la pagina desborda horizontalmente en " + name);
  const screenshot = path.join(OUT, name + ".png");
  await page.screenshot({ path: screenshot, fullPage: name === "desktop" });
  let detailScreenshot = "";
  if (name === "mobile") {
    await page.locator(".queue-card").nth(2).scrollIntoViewIfNeeded();
    detailScreenshot = path.join(OUT, "mobile_fiscal.png");
    await page.screenshot({ path: detailScreenshot });
  }
  await page.close();
  return { screenshot, detailScreenshot };
}

(async () => {
  fs.mkdirSync(OUT, { recursive: true });
  const browser = await chromium.launch({ headless: true, ...(CHROME ? { executablePath: CHROME } : {}) });
  try {
    const desktop = await render(browser, "desktop", { width: 1440, height: 1000 });
    const mobile = await render(browser, "mobile", { width: 390, height: 844 });
    process.stdout.write(JSON.stringify({ ok: true, desktop, mobile }) + "\n");
  } finally {
    await browser.close();
  }
})().catch((error) => {
  process.stderr.write(String(error && error.stack ? error.stack : error) + "\n");
  process.exitCode = 1;
});
