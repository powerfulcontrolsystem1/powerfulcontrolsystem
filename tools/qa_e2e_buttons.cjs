#!/usr/bin/env node
"use strict";

const fs = require("fs");
const path = require("path");
const { chromium } = require("playwright");

const ROOT = path.resolve(__dirname, "..");
const WEB_ROOT = path.join(ROOT, "web");
const OUT_DIR = process.env.PCS_QA_OUT_DIR || path.join(ROOT, "test_runs", "qa_e2e_buttons_" + new Date().toISOString().replace(/[:.]/g, "-").slice(0, 19));
const BASE_URL = (process.env.PCS_QA_BASE_URL || "https://powerfulcontrolsystem.com").replace(/\/+$/, "");
const EMAIL = process.env.PCS_QA_EMAIL || "";
const PASSWORD = process.env.PCS_QA_PASSWORD || "";
const EMPRESA_ID = process.env.PCS_QA_EMPRESA_ID || "7";
const MAX_PAGES = Number(process.env.PCS_QA_MAX_PAGES || "0");
const MAX_SAFE_CLICKS_PER_PAGE = Number(process.env.PCS_QA_MAX_SAFE_CLICKS_PER_PAGE || "8");
const SETTLE_MS = Number(process.env.PCS_QA_SETTLE_MS || "450");
const NETWORK_IDLE_TIMEOUT_MS = Number(process.env.PCS_QA_NETWORK_IDLE_TIMEOUT_MS || "3500");
const HEADLESS = process.env.PCS_QA_HEADLESS !== "0";
const CLICK_SAFE_BUTTONS = process.env.PCS_QA_CLICK_SAFE_BUTTONS !== "0";
const ROUTES_FILTER = (process.env.PCS_QA_ROUTES || "")
  .split(",")
  .map((route) => route.trim())
  .filter(Boolean);
const ALL_VIEWPORTS = [
  { name: "desktop", width: 1366, height: 900 },
  { name: "mobile", width: 390, height: 844, isMobile: true }
];
const VIEWPORTS = (process.env.PCS_QA_VIEWPORTS || "desktop,mobile")
  .split(",")
  .map((name) => name.trim())
  .filter(Boolean)
  .map((name) => ALL_VIEWPORTS.find((viewport) => viewport.name === name))
  .filter(Boolean);

const SAFE_TEXT = /^(abrir|cerrar|volver|cancelar|limpiar|buscar|filtrar|ver|mostrar|ocultar|editar|gestionar|detalle|detalles|actualizar vista|refrescar vista|nuevo|nueva|agregar|seleccionar|escuchar|sonando|copiar|expandir|minimizar|siguiente|anterior)$/i;
const UNSAFE_TEXT = /(eliminar|borrar|desactivar|activar|guardar|crear|registrar|enviar|pagar|comprar|checkout|confirmar|aprobar|rechazar|anular|cancelar pedido|cancelar servicio|cerrar caja|cobrar|emitir|facturar|despachar|publicar|subir|descargar|exportar|imprimir|reset|restablecer|reenviar|aceptar|generar|sincronizar|escanear|iniciar|completar|atender|llamar|re-llamar|listo|vencido|devolver|entregar)/i;
const UNSAFE_ATTR = /(delete|del|remove|destroy|save|submit|pay|checkout|purchase|send|confirm|approve|reject|cancel|close-sale|cobrar|emitir|facturar|dispatch|state|print|download|export|upload|scan|sync|publish|accept|resend|generate|crear|guardar|eliminar|pagar)/i;

function walk(dir, files = []) {
  if (!fs.existsSync(dir)) return files;
  for (const item of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, item.name);
    if (item.isDirectory()) walk(full, files);
    else if (item.isFile() && item.name.endsWith(".html")) files.push(full);
  }
  return files;
}

function routeForFile(file) {
  const rel = path.relative(WEB_ROOT, file).replace(/\\/g, "/");
  if (rel.includes("/source/")) return null;
  const url = "/" + rel;
  if (url.startsWith("/administrar_empresa/") || url === "/administrar_empresa.html") {
    const joiner = url.includes("?") ? "&" : "?";
    return url + joiner + "empresa_id=" + encodeURIComponent(EMPRESA_ID) + "&id=" + encodeURIComponent(EMPRESA_ID);
  }
  return url;
}

function discoverRoutes() {
  if (ROUTES_FILTER.length) {
    return MAX_PAGES > 0 ? ROUTES_FILTER.slice(0, MAX_PAGES) : ROUTES_FILTER;
  }
  const files = walk(WEB_ROOT)
    .map(routeForFile)
    .filter(Boolean)
    .sort((a, b) => {
      const score = (r) => (r.startsWith("/administrar_empresa") ? 0 : r.startsWith("/super") ? 1 : 2);
      return score(a) - score(b) || a.localeCompare(b);
    });
  return MAX_PAGES > 0 ? files.slice(0, MAX_PAGES) : files;
}

function ensureDir(dir) {
  fs.mkdirSync(dir, { recursive: true });
}

function slug(route) {
  return route.replace(/^\/+/, "").replace(/[^\w.-]+/g, "_").replace(/^_+|_+$/g, "").slice(0, 120) || "home";
}

function classifyButton(button) {
  const haystack = [
    button.text,
    button.id,
    button.name,
    button.className,
    button.type,
    button.href,
    JSON.stringify(button.dataset || {})
  ].join(" ");
  if (button.disabled || !button.visible) return "skip";
  if (UNSAFE_TEXT.test(haystack) || UNSAFE_ATTR.test(haystack)) return "unsafe";
  if (button.type === "submit") return "unsafe";
  if (button.href && !button.href.startsWith(BASE_URL) && !button.href.startsWith("/")) return "unsafe";
  if (button.dataset && Object.keys(button.dataset).some((key) => /tab|toggle|go|section|filter|view|modal|close/i.test(key))) return "safe";
  if (SAFE_TEXT.test(String(button.text || "").trim())) return "safe";
  if (button.ariaLabel || button.title) return "safe";
  return "review";
}

async function login(page) {
  if (!EMAIL || !PASSWORD) {
    throw new Error("Faltan PCS_QA_EMAIL y PCS_QA_PASSWORD en el entorno.");
  }
  let authenticated = false;
  for (let attempt = 1; attempt <= 2; attempt += 1) {
    await page.goto(BASE_URL + "/login.html", { waitUntil: "domcontentloaded", timeout: 45000 });
    await page.locator("#adminEmail").fill(EMAIL, { timeout: 15000 });
    await page.locator("#adminPassword").fill(PASSWORD, { timeout: 15000 });
    await Promise.all([
      page.waitForLoadState("networkidle", { timeout: 45000 }).catch(() => null),
      page.locator("#emailLoginBtn").click({ timeout: 15000 })
    ]);
    await page.waitForTimeout(900);
    const check = await page.request.get(BASE_URL + "/me", { timeout: 15000 }).catch(() => null);
    authenticated = Boolean(check && check.ok());
    if (authenticated) break;
    await page.waitForTimeout(700);
  }
  if (!authenticated) {
    throw new Error("No se pudo validar la sesion autenticada despues del login.");
  }
  await page.evaluate((empresaID) => {
    try {
      localStorage.setItem("active_empresa_id", empresaID);
      localStorage.setItem("empresa_id", empresaID);
      sessionStorage.setItem("active_empresa_id", empresaID);
      sessionStorage.setItem("empresa_id", empresaID);
      sessionStorage.setItem("admin_empresa_id", empresaID);
    } catch (e) {}
  }, EMPRESA_ID).catch(() => null);
}

async function collectButtons(page) {
  return page.evaluate(() => {
    const nodes = Array.from(document.querySelectorAll("button, [role='button'], input[type='button'], input[type='submit'], input[type='reset'], a.btn, a.button, [onclick]"));
    return nodes.map((el, index) => {
      const rect = el.getBoundingClientRect();
      const style = getComputedStyle(el);
      const visible = rect.width > 0 && rect.height > 0 && style.visibility !== "hidden" && style.display !== "none";
      const dataset = {};
      for (const key of Object.keys(el.dataset || {})) dataset[key] = el.dataset[key];
      el.setAttribute("data-qa-button-index", String(index));
      return {
        index,
        tag: el.tagName.toLowerCase(),
        text: (el.innerText || el.value || el.getAttribute("aria-label") || el.getAttribute("title") || "").trim().replace(/\s+/g, " ").slice(0, 120),
        ariaLabel: el.getAttribute("aria-label") || "",
        title: el.getAttribute("title") || "",
        id: el.id || "",
        name: el.getAttribute("name") || "",
        className: el.className ? String(el.className).slice(0, 180) : "",
        type: el.getAttribute("type") || "",
        href: el.getAttribute("href") || "",
        disabled: Boolean(el.disabled || el.getAttribute("aria-disabled") === "true"),
        visible,
        dataset,
        rect: { x: Math.round(rect.x), y: Math.round(rect.y), width: Math.round(rect.width), height: Math.round(rect.height) }
      };
    });
  });
}

async function collectVisualIssues(page) {
  return page.evaluate(() => {
    const issues = [];
    const viewportWidth = document.documentElement.clientWidth || innerWidth;
    const bodyWidth = Math.max(document.body.scrollWidth, document.documentElement.scrollWidth);
    if (bodyWidth > viewportWidth + 24) {
      issues.push({ type: "horizontal-overflow", viewportWidth, bodyWidth });
    }
    for (const el of Array.from(document.querySelectorAll("button, [role='button'], input[type='button'], input[type='submit'], a.btn, a.button"))) {
      const rect = el.getBoundingClientRect();
      const style = getComputedStyle(el);
      const visible = rect.width > 0 && rect.height > 0 && style.visibility !== "hidden" && style.display !== "none";
      if (!visible) continue;
      const label = (el.innerText || el.value || el.getAttribute("aria-label") || el.getAttribute("title") || "").trim();
      if (!label) {
        issues.push({ type: "button-without-label", selector: el.id ? "#" + el.id : el.tagName.toLowerCase(), rect: { x: rect.x, y: rect.y, width: rect.width, height: rect.height } });
      }
      if (el.scrollWidth > el.clientWidth + 4 || el.scrollHeight > el.clientHeight + 4) {
        issues.push({ type: "button-content-overflow", label: label.slice(0, 80), clientWidth: el.clientWidth, scrollWidth: el.scrollWidth, clientHeight: el.clientHeight, scrollHeight: el.scrollHeight });
      }
    }
    return issues.slice(0, 30);
  });
}

async function auditRoute(context, route, viewport) {
  const page = await context.newPage();
  const consoleErrors = [];
  const pageErrors = [];
  const requestFailures = [];
  const responseErrors = [];
  const dialogs = [];
  page.on("console", (msg) => {
    if (["error", "warning"].includes(msg.type())) consoleErrors.push({ type: msg.type(), text: msg.text().slice(0, 600) });
  });
  page.on("pageerror", (err) => pageErrors.push(String(err && err.message ? err.message : err).slice(0, 600)));
  page.on("requestfailed", (req) => requestFailures.push({ url: req.url(), failure: req.failure() ? req.failure().errorText : "request failed" }));
  page.on("response", (res) => {
    const status = res.status();
    const url = res.url();
    if (status >= 400 && !url.includes("/favicon.ico")) responseErrors.push({ status, url });
  });
  page.on("dialog", async (dialog) => {
    dialogs.push({ type: dialog.type(), message: dialog.message().slice(0, 300) });
    await dialog.dismiss().catch(() => null);
  });

  const url = BASE_URL + route;
  const result = { route, viewport: viewport.name, url, status: "ok", buttons: [], clicked: [], skipped: [], issues: [], consoleErrors, pageErrors, requestFailures, responseErrors, dialogs, screenshot: "" };
  try {
    await page.goto(url, { waitUntil: "domcontentloaded", timeout: 45000 });
    await page.waitForLoadState("networkidle", { timeout: NETWORK_IDLE_TIMEOUT_MS }).catch(() => null);
    await page.waitForTimeout(SETTLE_MS);
    const bodyText = await page.locator("body").innerText({ timeout: 8000 }).catch(() => "");
    result.unauthorized = /unauthorized|no autorizado|iniciar sesion|iniciar sesión/i.test(bodyText);
    result.securityBlock = /completa la verificaci[oó]n de seguridad|verify you are human|captcha challenge|hcaptcha|cf-turnstile|recaptcha-checkbox/i.test(bodyText);
    result.buttons = (await collectButtons(page)).map((button) => ({ ...button, classification: classifyButton(button) }));
    result.issues = await collectVisualIssues(page);
    const shotName = viewport.name + "_" + slug(route) + ".png";
    result.screenshot = path.join(OUT_DIR, "screenshots", shotName);
    await page.screenshot({ path: result.screenshot, fullPage: false }).catch((err) => {
      result.issues.push({ type: "screenshot-failed", message: String(err.message || err).slice(0, 300) });
    });

    if (CLICK_SAFE_BUTTONS) {
      const safeButtons = result.buttons.filter((button) => button.classification === "safe").slice(0, MAX_SAFE_CLICKS_PER_PAGE);
      for (const button of safeButtons) {
        try {
          const beforeUrl = page.url();
          await page.locator('[data-qa-button-index="' + button.index + '"]').click({ timeout: 1800, force: false });
          await page.waitForTimeout(180);
          result.clicked.push({ index: button.index, text: button.text || button.ariaLabel || button.title || button.id || button.className });
          const afterUrl = page.url();
          if (afterUrl !== beforeUrl) {
            if (afterUrl.startsWith(BASE_URL)) {
              await page.goto(url, { waitUntil: "domcontentloaded", timeout: 18000 }).catch(() => null);
              await page.waitForTimeout(180);
            } else {
              result.issues.push({ type: "external-navigation", from: beforeUrl, to: afterUrl });
              break;
            }
          }
        } catch (err) {
          result.issues.push({ type: "safe-button-click-failed", button, message: String(err.message || err).slice(0, 500) });
        }
      }
    }
    result.skipped = result.buttons.filter((button) => button.classification === "unsafe").map((button) => ({ index: button.index, text: button.text, id: button.id, className: button.className, dataset: button.dataset })).slice(0, 80);
    if (result.pageErrors.length || result.consoleErrors.length || result.responseErrors.some((x) => ![401, 403, 404].includes(x.status)) || result.securityBlock) {
      result.status = "review";
    }
  } catch (err) {
    result.status = "error";
    result.error = String(err.message || err).slice(0, 900);
  } finally {
    await page.close().catch(() => null);
  }
  return result;
}

function summarize(results) {
  const byStatus = results.reduce((acc, item) => {
    acc[item.status] = (acc[item.status] || 0) + 1;
    return acc;
  }, {});
  const totalButtons = results.reduce((n, item) => n + item.buttons.length, 0);
  const clicked = results.reduce((n, item) => n + item.clicked.length, 0);
  const unsafe = results.reduce((n, item) => n + item.skipped.length, 0);
  const pagesWithErrors = results.filter((item) => item.status !== "ok" || item.pageErrors.length || item.consoleErrors.length || item.requestFailures.length || item.responseErrors.length || item.issues.length);
  return { totalPages: results.length, byStatus, totalButtons, clicked, unsafe, pagesWithErrors: pagesWithErrors.length };
}

function writeMarkdown(results, summary) {
  const lines = [];
  lines.push("# QA E2E botones y visual");
  lines.push("");
  lines.push("- Base URL: `" + BASE_URL + "`");
  lines.push("- Empresa: `" + EMPRESA_ID + "`");
  lines.push("- Paginas recorridas: `" + summary.totalPages + "`");
  lines.push("- Botones detectados: `" + summary.totalButtons + "`");
  lines.push("- Clicks seguros ejecutados: `" + summary.clicked + "`");
  lines.push("- Acciones riesgosas omitidas: `" + summary.unsafe + "`");
  lines.push("- Paginas con hallazgos: `" + summary.pagesWithErrors + "`");
  lines.push("");
  lines.push("## Hallazgos");
  const findings = results.filter((item) => item.status !== "ok" || item.pageErrors.length || item.consoleErrors.length || item.requestFailures.length || item.responseErrors.length || item.issues.length).slice(0, 120);
  if (!findings.length) lines.push("Sin hallazgos en el barrido automatizado.");
  for (const item of findings) {
    lines.push("");
    lines.push("### " + item.viewport + " " + item.route);
    lines.push("- Estado: `" + item.status + "`");
    if (item.error) lines.push("- Error: `" + item.error.replace(/`/g, "'") + "`");
    if (item.unauthorized) lines.push("- Autorizacion: posible pantalla o respuesta no autenticada.");
    if (item.securityBlock) lines.push("- Seguridad: bloqueo/captcha detectado.");
    if (item.pageErrors.length) lines.push("- Page errors: `" + item.pageErrors.slice(0, 2).join(" | ").replace(/`/g, "'") + "`");
    if (item.consoleErrors.length) lines.push("- Consola: `" + item.consoleErrors.slice(0, 3).map((x) => x.text).join(" | ").replace(/`/g, "'") + "`");
    if (item.responseErrors.length) lines.push("- HTTP >=400: `" + item.responseErrors.slice(0, 4).map((x) => x.status + " " + x.url.replace(BASE_URL, "")).join(" | ").replace(/`/g, "'") + "`");
    if (item.requestFailures.length) lines.push("- Requests fallidos: `" + item.requestFailures.slice(0, 3).map((x) => x.failure + " " + x.url.replace(BASE_URL, "")).join(" | ").replace(/`/g, "'") + "`");
    if (item.issues.length) lines.push("- Visual/interaccion: `" + item.issues.slice(0, 4).map((x) => x.type).join(", ") + "`");
    if (item.screenshot) lines.push("- Captura: `" + path.relative(ROOT, item.screenshot).replace(/\\/g, "/") + "`");
  }
  fs.writeFileSync(path.join(OUT_DIR, "reporte.md"), lines.join("\n"), "utf8");
}

async function main() {
  ensureDir(OUT_DIR);
  ensureDir(path.join(OUT_DIR, "screenshots"));
  const jsonlPath = path.join(OUT_DIR, "results.jsonl");
  fs.writeFileSync(jsonlPath, "", "utf8");
  const routes = discoverRoutes();
  const browser = await chromium.launch({ headless: HEADLESS });
  const allResults = [];
  try {
    for (const viewport of VIEWPORTS) {
      const context = await browser.newContext({
        viewport: { width: viewport.width, height: viewport.height },
        isMobile: Boolean(viewport.isMobile),
        deviceScaleFactor: viewport.isMobile ? 2 : 1,
        ignoreHTTPSErrors: true
      });
      const loginPage = await context.newPage();
      await login(loginPage);
      await loginPage.close();
      for (let i = 0; i < routes.length; i += 1) {
        const result = await auditRoute(context, routes[i], viewport);
        allResults.push(result);
        fs.appendFileSync(jsonlPath, JSON.stringify(result) + "\n", "utf8");
        process.stdout.write(JSON.stringify({ done: allResults.length, total: routes.length * VIEWPORTS.length, viewport: viewport.name, route: routes[i], status: result.status, buttons: result.buttons.length, clicked: result.clicked.length }) + "\n");
      }
      await context.close();
    }
  } finally {
    await browser.close();
  }
  const summary = summarize(allResults);
  fs.writeFileSync(path.join(OUT_DIR, "results.json"), JSON.stringify({ summary, results: allResults }, null, 2), "utf8");
  writeMarkdown(allResults, summary);
  process.stdout.write(JSON.stringify({ outDir: OUT_DIR, summary }, null, 2) + "\n");
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
