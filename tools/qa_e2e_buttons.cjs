#!/usr/bin/env node
"use strict";

const fs = require("fs");
const path = require("path");
const Module = require("module");

function bundledPlaywrightPath() {
  const candidates = [];
  if (process.env.NODE_PATH) candidates.push(process.env.NODE_PATH);
  if (process.env.USERPROFILE) {
    candidates.push(path.join(process.env.USERPROFILE, ".cache", "codex-runtimes", "codex-primary-runtime", "dependencies", "node", "node_modules"));
  }
  for (const candidate of candidates) {
    if (!candidate || !fs.existsSync(path.join(candidate, "playwright"))) continue;
    if (!process.env.NODE_PATH) process.env.NODE_PATH = candidate;
    else if (!process.env.NODE_PATH.split(path.delimiter).includes(candidate)) process.env.NODE_PATH += path.delimiter + candidate;
    Module._initPaths();
    return candidate;
  }
  return "";
}

bundledPlaywrightPath();
let chromium;
try {
  ({ chromium } = require("playwright"));
} catch (error) {
  throw new Error("No se encontro Playwright. Use el runtime Codex o defina NODE_PATH a su node_modules; no instale dependencias para ejecutar esta auditoria. Causa: " + String(error.message || error));
}

const ROOT = path.resolve(__dirname, "..");
const WEB_ROOT = path.join(ROOT, "web");
const OUT_DIR = process.env.PCS_QA_OUT_DIR || path.join(ROOT, "test_runs", "qa_e2e_buttons_" + new Date().toISOString().replace(/[:.]/g, "-").slice(0, 19));
// A real authenticated audit must always choose its target explicitly. Keeping
// a public production URL as a fallback made it too easy to run the checker
// against the wrong environment when credentials were present.
const BASE_URL = (process.env.PCS_QA_BASE_URL || "").replace(/\/+$/, "");
const EMAIL = process.env.PCS_QA_EMAIL || "";
const PASSWORD = process.env.PCS_QA_PASSWORD || "";
const EMPRESA_ID = process.env.PCS_QA_EMPRESA_ID || "";
const MAX_PAGES = Number(process.env.PCS_QA_MAX_PAGES || "0");
const ROUTE_OFFSET = Number(process.env.PCS_QA_ROUTE_OFFSET || "0");
const ROUTE_BATCH_SIZE = Number(process.env.PCS_QA_ROUTE_BATCH_SIZE || "0");
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
const CHROME_EXECUTABLE = process.env.PCS_QA_CHROME_EXECUTABLE || "";
const VALIDATE_RUNTIME_ONLY = process.env.PCS_QA_VALIDATE_RUNTIME === "1";
const MUTATING_METHODS = new Set(["POST", "PUT", "PATCH", "DELETE"]);
const NON_OPERATIONAL_MUTATION_PATHS = new Set(["/api/public/portal_visitas"]);

// "Cerrar" y "Cancelar" no son universalmente inocuos: pueden cerrar caja,
// anular un flujo operativo o descartar un formulario. El auditor solo pulsa
// acciones cuyo texto sea inequívocamente de consulta o navegación.
const SAFE_TEXT = /^(abrir|volver|limpiar|buscar|filtrar|ver|mostrar|ocultar|detalle|detalles|actualizar vista|refrescar vista|seleccionar|escuchar|sonando|copiar|expandir|minimizar|siguiente|anterior)$/i;
const AMBIGUOUS_OPERATION_TEXT = /(^|\s)(cerrar|cancelar)(\s|$)/i;
const AI_ACTION_TEXT = /(^|[^\p{L}\p{N}_])(ia|ai|gpt|openai|asistente)([^\p{L}\p{N}_]|$)/iu;
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
    const normalized = ROUTES_FILTER.map((route) => {
      const parsed = new URL(route, BASE_URL);
      if (parsed.pathname.startsWith("/administrar_empresa/") || parsed.pathname === "/administrar_empresa.html") {
        if (!parsed.searchParams.has("empresa_id")) parsed.searchParams.set("empresa_id", EMPRESA_ID);
        if (!parsed.searchParams.has("id")) parsed.searchParams.set("id", EMPRESA_ID);
      }
      return parsed.pathname + parsed.search + parsed.hash;
    });
    return sliceRouteBatch(normalized);
  }
  const files = walk(WEB_ROOT)
    .map(routeForFile)
    .filter(Boolean)
    .sort((a, b) => {
      const score = (r) => (r.startsWith("/administrar_empresa") ? 0 : r.startsWith("/super") ? 1 : 2);
      return score(a) - score(b) || a.localeCompare(b);
    });
  return sliceRouteBatch(files);
}

// El inventario completo puede superar el límite de un ejecutor remoto. El
// desplazamiento y el tamaño de lote permiten recorrerlo de forma determinista
// sin ocultar rutas ni mezclar resultados de distintas corridas.
function sliceRouteBatch(routes) {
  if (!Number.isInteger(ROUTE_OFFSET) || ROUTE_OFFSET < 0) {
    throw new Error("PCS_QA_ROUTE_OFFSET debe ser un entero mayor o igual a cero.");
  }
  if (!Number.isInteger(ROUTE_BATCH_SIZE) || ROUTE_BATCH_SIZE < 0) {
    throw new Error("PCS_QA_ROUTE_BATCH_SIZE debe ser un entero mayor o igual a cero.");
  }
  const effectiveSize = ROUTE_BATCH_SIZE > 0 ? ROUTE_BATCH_SIZE : MAX_PAGES;
  return effectiveSize > 0 ? routes.slice(ROUTE_OFFSET, ROUTE_OFFSET + effectiveSize) : routes.slice(ROUTE_OFFSET);
}

function ensureDir(dir) {
  fs.mkdirSync(dir, { recursive: true });
}

function validateExplicitTarget() {
  if (!BASE_URL) {
    throw new Error("Defina PCS_QA_BASE_URL explícitamente; la auditoría no tiene destino predeterminado.");
  }
  let parsed;
  try {
    parsed = new URL(BASE_URL);
  } catch (error) {
    throw new Error("PCS_QA_BASE_URL debe ser una URL http(s) válida.");
  }
  if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
    throw new Error("PCS_QA_BASE_URL debe usar http o https.");
  }
  if (!/^\d+$/.test(EMPRESA_ID) || Number(EMPRESA_ID) <= 0) {
    throw new Error("Defina PCS_QA_EMPRESA_ID con una empresa positiva autorizada.");
  }
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
  if (AMBIGUOUS_OPERATION_TEXT.test(haystack)) return "review";
  if (AI_ACTION_TEXT.test(haystack)) return "review";
  if (button.type === "submit") return "unsafe";
  if (button.href && !button.href.startsWith(BASE_URL) && !button.href.startsWith("/")) return "unsafe";
  if (button.dataset && Object.keys(button.dataset).some((key) => /^(tab|toggle|go|section|filter|view|modal|close|target)$/i.test(key))) return "safe";
  if (SAFE_TEXT.test(String(button.text || "").trim())) return "safe";
  return "review";
}

function cssAttributeValue(value) {
  return String(value || "").replace(/\\/g, "\\\\").replace(/"/g, '\\"');
}

// Los indices DOM solo sirven para inventario. Varias paginas insertan roles,
// menús o controles al iniciar y el mismo indice puede apuntar a otro boton
// despues de una recarga. Solo se ejecutan controles con identidad estable.
function stableButtonSelector(button) {
  if (button.id) return '[id="' + cssAttributeValue(button.id) + '"]';
  if (button.dataset && button.dataset.target) return '[data-target="' + cssAttributeValue(button.dataset.target) + '"]';
  if (button.href) return button.tag + '[href="' + cssAttributeValue(button.href) + '"]';
  return "";
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
      const closedDetails = el.closest("details:not([open])");
      const hiddenByDetails = Boolean(closedDetails && !el.closest("summary"));
      const visible = !hiddenByDetails && rect.width > 0 && rect.height > 0 && style.visibility !== "hidden" && style.display !== "none" && (!el.checkVisibility || el.checkVisibility({ checkOpacity: true, checkVisibilityCSS: true }));
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
      const closedDetails = el.closest("details:not([open])");
      const hiddenByDetails = Boolean(closedDetails && !el.closest("summary"));
      const visible = !hiddenByDetails && rect.width > 0 && rect.height > 0 && style.visibility !== "hidden" && style.display !== "none" && (!el.checkVisibility || el.checkVisibility({ checkOpacity: true, checkVisibilityCSS: true }));
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
  const blockedMutations = [];
  const telemetryMutations = [];
  let navigatingForAudit = false;
  // Esta auditoria solo valida navegacion y presentacion. Incluso si una
  // etiqueta, un dataset o un indice dinamico se clasifican mal, la red es la
  // ultima frontera: ninguna accion autenticada puede modificar staging.
  await page.route("**/*", async (routeHandler) => {
    const request = routeHandler.request();
    const method = request.method().toUpperCase();
    if (MUTATING_METHODS.has(method)) {
      const requestPath = new URL(request.url()).pathname;
      if (NON_OPERATIONAL_MUTATION_PATHS.has(requestPath)) {
        telemetryMutations.push({ method, url: request.url(), resourceType: request.resourceType() });
        await routeHandler.abort("blockedbyclient");
        return;
      }
      blockedMutations.push({ method, url: request.url(), resourceType: request.resourceType() });
      await routeHandler.abort("blockedbyclient");
      return;
    }
    await routeHandler.continue();
  });
  page.on("console", (msg) => {
    const message = msg.text();
    if (/Service Worker registration blocked by Playwright/i.test(message)) return;
    if (/ERR_BLOCKED_BY_CLIENT\.Inspector/i.test(message)) return;
    if (["error", "warning"].includes(msg.type())) consoleErrors.push({ type: msg.type(), text: message.slice(0, 600) });
  });
  page.on("pageerror", (err) => pageErrors.push(String(err && err.message ? err.message : err).slice(0, 600)));
  page.on("requestfailed", (req) => {
    if (MUTATING_METHODS.has(req.method().toUpperCase())) return;
    const failure = req.failure() ? req.failure().errorText : "request failed";
    if (navigatingForAudit && failure === "net::ERR_ABORTED") return;
    requestFailures.push({ url: req.url(), failure });
  });
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
  const result = { route, viewport: viewport.name, url, status: "ok", buttons: [], clicked: [], skipped: [], blockedMutations, telemetryMutations, issues: [], consoleErrors, pageErrors, requestFailures, responseErrors, dialogs, screenshot: "" };
  const resetAuditPage = async () => {
    navigatingForAudit = true;
    try {
      await page.goto(url, { waitUntil: "domcontentloaded", timeout: 18000 }).catch(() => null);
      await page.waitForTimeout(180);
    } finally {
      navigatingForAudit = false;
    }
  };
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
          const selector = stableButtonSelector(button);
          if (!selector) {
            result.skipped.push({
              index: button.index,
              text: button.text,
              id: button.id,
              reason: "safe-control-without-stable-selector"
            });
            continue;
          }
          const target = page.locator(selector);
          const targetCount = await target.count();
          if (targetCount !== 1 || !(await target.isVisible().catch(() => false))) {
            result.skipped.push({
              index: button.index,
              text: button.text,
              id: button.id,
              selector,
              reason: targetCount !== 1 ? "safe-selector-not-unique" : "safe-control-not-visible-after-state-change"
            });
            continue;
          }
          const beforeUrl = page.url();
          const blockedBefore = blockedMutations.length;
          await target.click({ timeout: 1800, force: false });
          await page.waitForTimeout(350);
          const blockedByClick = blockedMutations.slice(blockedBefore);
          result.clicked.push({
            index: button.index,
            text: button.text || button.ariaLabel || button.title || button.id || button.className,
            mutationBlocked: blockedByClick.length > 0
          });
          if (blockedByClick.length) {
            result.issues.push({
              type: "safe-button-attempted-mutation",
              button: { index: button.index, text: button.text, id: button.id },
              requests: blockedByClick.slice(0, 8)
            });
          }
          const afterUrl = page.url();
          if (afterUrl !== beforeUrl) {
            if (afterUrl.startsWith(BASE_URL)) {
              await resetAuditPage();
            } else {
              result.issues.push({ type: "external-navigation", from: beforeUrl, to: afterUrl });
              break;
            }
          }
          if (afterUrl === beforeUrl) {
            await resetAuditPage();
          }
        } catch (err) {
          result.issues.push({ type: "safe-button-click-failed", button, message: String(err.message || err).slice(0, 500) });
        }
      }
    }
    if (blockedMutations.length && !result.issues.some((issue) => issue.type === "safe-button-attempted-mutation")) {
      result.issues.push({ type: "page-attempted-mutation", requests: blockedMutations.slice(0, 8) });
    }
    result.skipped = result.skipped.concat(result.buttons.filter((button) => button.classification === "unsafe").map((button) => ({ index: button.index, text: button.text, id: button.id, className: button.className, dataset: button.dataset }))).slice(0, 80);
    if (result.pageErrors.length || result.consoleErrors.length || result.responseErrors.some((x) => ![401, 403, 404].includes(x.status)) || result.securityBlock || result.blockedMutations.length) {
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
  const blockedMutations = results.reduce((n, item) => n + item.blockedMutations.length, 0);
  const telemetryMutations = results.reduce((n, item) => n + (item.telemetryMutations || []).length, 0);
  const pagesWithErrors = results.filter((item) => item.status !== "ok" || item.pageErrors.length || item.consoleErrors.length || item.requestFailures.length || item.responseErrors.length || item.issues.length);
  return { totalPages: results.length, byStatus, totalButtons, clicked, unsafe, blockedMutations, telemetryMutations, pagesWithErrors: pagesWithErrors.length };
}

function writeMarkdown(results, summary) {
  const lines = [];
  lines.push("# QA E2E botones y visual");
  lines.push("");
  lines.push("- Base URL: `" + BASE_URL + "`");
  lines.push("- Empresa: `" + EMPRESA_ID + "`");
  lines.push("- Paginas recorridas: `" + summary.totalPages + "`");
  lines.push("- Desplazamiento de rutas: `" + ROUTE_OFFSET + "`");
  lines.push("- Tamano de lote: `" + (ROUTE_BATCH_SIZE || MAX_PAGES || "completo") + "`");
  lines.push("- Botones detectados: `" + summary.totalButtons + "`");
  lines.push("- Clicks seguros ejecutados: `" + summary.clicked + "`");
  lines.push("- Acciones riesgosas omitidas: `" + summary.unsafe + "`");
  lines.push("- Mutaciones bloqueadas por la guardia: `" + summary.blockedMutations + "`");
  lines.push("- Telemetria no operativa bloqueada: `" + summary.telemetryMutations + "`");
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
    if (item.blockedMutations.length) lines.push("- Mutaciones bloqueadas: `" + item.blockedMutations.slice(0, 4).map((x) => x.method + " " + x.url.replace(BASE_URL, "")).join(" | ").replace(/`/g, "'") + "`");
    if (item.issues.length) lines.push("- Visual/interaccion: `" + item.issues.slice(0, 4).map((x) => x.type).join(", ") + "`");
    if (item.screenshot) lines.push("- Captura: `" + path.relative(ROOT, item.screenshot).replace(/\\/g, "/") + "`");
  }
  fs.writeFileSync(path.join(OUT_DIR, "reporte.md"), lines.join("\n"), "utf8");
}

async function main() {
  if (VALIDATE_RUNTIME_ONLY) {
    process.stdout.write(JSON.stringify({ playwright: "ready", chromeExecutable: CHROME_EXECUTABLE || "bundled/default", runtimeOnly: true }) + "\n");
    return;
  }
  validateExplicitTarget();
  ensureDir(OUT_DIR);
  ensureDir(path.join(OUT_DIR, "screenshots"));
  const jsonlPath = path.join(OUT_DIR, "results.jsonl");
  fs.writeFileSync(jsonlPath, "", "utf8");
  const routes = discoverRoutes();
  const browser = await chromium.launch({
    headless: HEADLESS,
    ...(CHROME_EXECUTABLE ? { executablePath: CHROME_EXECUTABLE } : {})
  });
  const allResults = [];
  try {
    for (const viewport of VIEWPORTS) {
      const context = await browser.newContext({
        viewport: { width: viewport.width, height: viewport.height },
        isMobile: Boolean(viewport.isMobile),
        deviceScaleFactor: viewport.isMobile ? 2 : 1,
        ignoreHTTPSErrors: true,
        serviceWorkers: "block"
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
  process.stdout.write(JSON.stringify({ outDir: OUT_DIR, routeOffset: ROUTE_OFFSET, routeBatchSize: ROUTE_BATCH_SIZE || MAX_PAGES || 0, summary }, null, 2) + "\n");
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
