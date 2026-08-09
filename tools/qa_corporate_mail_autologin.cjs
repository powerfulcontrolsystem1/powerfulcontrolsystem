#!/usr/bin/env node
"use strict";

const path = require("path");
const Module = require("module");

process.env.NODE_PATH = process.env.NODE_PATH || path.join(process.env.USERPROFILE || "", ".cache", "codex-runtimes", "codex-primary-runtime", "dependencies", "node", "node_modules");
Module._initPaths();
const { chromium } = require("playwright");

const baseURL = String(process.env.PCS_QA_BASE_URL || "").replace(/\/+$/, "");
const email = String(process.env.PCS_QA_EMAIL || "");
const password = String(process.env.PCS_QA_PASSWORD || "");
const companyID = Number(process.env.PCS_QA_EMPRESA_ID || 0);
const executablePath = String(process.env.PCS_QA_CHROME_PATH || "");
const stripSSOTheme = String(process.env.PCS_QA_STRIP_SSO_THEME || "") === "1";

function cookieMetadata(headers) {
  const raw = headers["set-cookie"] || "";
  return String(raw).split(/,(?=\s*[^;,=]+=[^;,]+)/).filter(Boolean).map((cookie) => ({
    name: String(cookie).split("=", 1)[0].trim(),
    domain: ((String(cookie).match(/;\s*Domain=([^;]+)/i) || [])[1] || "host-only").trim(),
    path: ((String(cookie).match(/;\s*Path=([^;]+)/i) || [])[1] || "default").trim(),
    secure: /;\s*Secure(?:;|$)/i.test(cookie),
    sameSite: ((String(cookie).match(/;\s*SameSite=([^;]+)/i) || [])[1] || "unspecified").trim()
  }));
}

if (!baseURL || !email || !password || !companyID || !executablePath) throw new Error("Falta configuración segura de QA de correo.");

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
    await page.goto(baseURL + "/administrar_empresa/email_corporativo.html?empresa_id=" + companyID, { waitUntil: "networkidle" });
    const frame = page.locator("#emailFrame");
    await frame.waitFor({ state: "attached", timeout: 20000 });
    await page.waitForFunction(() => {
      const element = document.querySelector("#emailFrame");
      return Boolean(element && element.getAttribute("src"));
    }, null, { timeout: 20000 });
    const autologinURL = await frame.getAttribute("src");
    const response = await context.request.get(autologinURL, { maxRedirects: 0, failOnStatusCode: false });
    const location = response.headers().location || "";
    const result = {
      status: response.status(),
      redirect: Boolean(location),
      redirectHost: location ? new URL(location, autologinURL).host : "",
      cookies: cookieMetadata(response.headers()),
      companyID
    };
    if (response.status() >= 300 && response.status() < 400 && location) {
      let target = new URL(location, autologinURL).toString();
      if (stripSSOTheme) {
        const parsedTarget = new URL(target);
        const hash = parsedTarget.searchParams.get("hash") || "";
        parsedTarget.search = hash ? "?sso&hash=" + encodeURIComponent(hash) : parsedTarget.search;
        target = parsedTarget.toString();
      }
      let currentURL = target;
      const redirectChain = [];
      for (let step = 0; step < 6; step += 1) {
        const currentResponse = await context.request.get(currentURL, { maxRedirects: 0, failOnStatusCode: false });
        const currentLocation = currentResponse.headers().location || "";
        const parsed = new URL(currentURL);
        redirectChain.push({ host: parsed.host, path: parsed.pathname, status: currentResponse.status(), redirect: Boolean(currentLocation), cookies: cookieMetadata(currentResponse.headers()) });
        if (!(currentResponse.status() >= 300 && currentResponse.status() < 400 && currentLocation)) break;
        currentURL = new URL(currentLocation, currentURL).toString();
      }
      result.redirectChain = redirectChain;
      const webmail = await context.newPage();
      let navigationError = "";
      const finalResponse = await webmail.goto(currentURL, { waitUntil: "domcontentloaded", timeout: 30000 }).catch((error) => {
        navigationError = String(error && error.message || "").split("\n")[0].replace(/https?:\/\/\S+/g, "[url]");
        return null;
      });
      result.webmailStatus = finalResponse ? finalResponse.status() : 0;
      result.webmailTitle = await webmail.title().catch(() => "");
      result.webmailURLHost = /^https?:/i.test(webmail.url()) ? new URL(webmail.url()).host : "";
      result.navigationError = navigationError;
    }
    process.stdout.write(JSON.stringify(result) + "\n");
    await context.close();
  } finally {
    await browser.close();
  }
})().catch((error) => { console.error(String(error.message || error)); process.exit(1); });
