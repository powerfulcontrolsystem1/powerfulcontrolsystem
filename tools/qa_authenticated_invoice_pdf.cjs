#!/usr/bin/env node
"use strict";

const fs = require("fs");
const path = require("path");
const Module = require("module");

const modules = process.env.NODE_PATH || path.join(process.env.USERPROFILE || "", ".cache", "codex-runtimes", "codex-primary-runtime", "dependencies", "node", "node_modules");
process.env.NODE_PATH = modules;
Module._initPaths();
const { chromium } = require("playwright");

const baseURL = String(process.env.PCS_PRINT_BASE_URL || "").replace(/\/+$/, "");
const email = String(process.env.PCS_PRINT_EMAIL || "");
const password = String(process.env.PCS_PRINT_PASSWORD || "");
const companyID = Number(process.env.PCS_PRINT_EMPRESA_ID || 0);
const legalNumber = String(process.env.PCS_PRINT_DOCUMENT || "");
const output = path.resolve(process.env.PCS_PRINT_OUTPUT || "test_runs/p109_invoice.pdf");
const executablePath = String(process.env.PCS_PRINT_CHROME_PATH || "");
const pageOverride = String(process.env.PCS_PRINT_PAGE_OVERRIDE || "");

if (!baseURL || !email || !password || !companyID || !legalNumber || !executablePath) throw new Error("Falta configuración segura de impresión autenticada.");

(async () => {
  const browser = await chromium.launch({ headless: true, executablePath });
  try {
    const context = await browser.newContext({ ignoreHTTPSErrors: true, serviceWorkers: "block" });
    const page = await context.newPage();
    await page.goto(baseURL + "/login.html", { waitUntil: "domcontentloaded" });
    await page.locator("#adminEmail").fill(email);
    await page.locator("#adminPassword").fill(password);
    await page.locator("#emailLoginBtn").click();
    await page.waitForURL(/super_administrador|seleccionar_empresa|administrar_empresa/, { timeout: 30000 });
    if (pageOverride) {
      const html = fs.readFileSync(path.resolve(pageOverride), "utf8");
      await page.route("**/administrar_empresa/facturas_electronicas.html*", async (route) => {
        await route.fulfill({ status: 200, contentType: "text/html; charset=utf-8", body: html });
      });
    }
    await page.goto(baseURL + "/administrar_empresa/facturas_electronicas.html?empresa_id=" + companyID, { waitUntil: "networkidle" });
    if (pageOverride) {
      const candidateLoaded = await page.locator("body").evaluate((body) => body.innerHTML.includes("systemLogoUrl"));
      if (!candidateLoaded) throw new Error("La página candidata local no reemplazó la vista remota.");
    }
    const row = page.getByRole("row", { name: new RegExp(legalNumber.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")) });
    await row.waitFor({ state: "visible", timeout: 20000 });
    const popupPromise = context.waitForEvent("page", { timeout: 10000 }).catch(() => null);
    await row.getByRole("button", { name: "Visualizar", exact: true }).click();
    const popup = await popupPromise;
    const printable = popup || page;
    if (popup) await popup.waitForLoadState("networkidle", { timeout: 20000 }).catch(() => null);
    await printable.waitForFunction(() => Array.from(document.images).every((image) => image.complete), null, { timeout: 10000 }).catch(() => null);
    const brokenImages = await printable.locator("img").evaluateAll((images) => images.filter((image) => !image.naturalWidth).map((image) => image.alt || image.src));
    fs.mkdirSync(path.dirname(output), { recursive: true });
    await printable.pdf({ path: output, format: "A4", printBackground: true, preferCSSPageSize: true });
    const stats = fs.statSync(output);
    process.stdout.write(JSON.stringify({ status: brokenImages.length ? "review" : "ok", output, bytes: stats.size, document: legalNumber, brokenImages }) + "\n");
    await context.close();
  } finally {
    await browser.close();
  }
})().catch((error) => { console.error(String(error.message || error)); process.exit(1); });
