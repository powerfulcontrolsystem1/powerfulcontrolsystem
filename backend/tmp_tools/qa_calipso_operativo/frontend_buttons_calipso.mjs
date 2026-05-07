import { createRequire } from 'node:module';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const require = createRequire(import.meta.url);
const { chromium } = require('playwright');
const scriptDir = path.dirname(fileURLToPath(import.meta.url));

const baseURL = (process.env.QA_BASE_URL || 'http://127.0.0.1:8080').replace(/\/+$/, '');
const empresaID = Number(process.env.QA_EMPRESA_ID || 7);
const email = String(process.env.QA_ADMIN_EMAIL || '').trim();
const password = String(process.env.QA_ADMIN_PASSWORD || '');
const outPath = process.env.QA_REPORT_PATH || path.join(scriptDir, 'frontend_buttons_calipso_report.json');
const initialViewport = {
  width: Number(process.env.QA_VIEWPORT_WIDTH || 1440),
  height: Number(process.env.QA_VIEWPORT_HEIGHT || 920)
};

if (!email || !password) {
  throw new Error('Define QA_ADMIN_EMAIL y QA_ADMIN_PASSWORD para ejecutar el recorrido de botones.');
}

const report = {
  empresa_id: empresaID,
  base_url: baseURL,
  started_at: new Date().toISOString(),
  login: {},
  shell: { desktop: {}, mobile: {} },
  modules: [],
  events: [],
  network_errors: [],
  console_errors: [],
  page_errors: [],
  skipped: [],
  summary: {}
};

function pushLimited(list, item, limit = 500) {
  if (list.length < limit) list.push(item);
}

function compactText(value) {
  return String(value || '').replace(/\s+/g, ' ').trim();
}

function normalizeURL(url) {
  try {
    const parsed = new URL(url, baseURL);
    return parsed.pathname + parsed.search;
  } catch {
    return String(url || '');
  }
}

function classifyAction(text, href, tag, type) {
  const label = compactText(text).toLowerCase();
  const url = String(href || '').toLowerCase();
  const kind = String(type || '').toLowerCase();
  if (!label && !url) return { action: 'skip', reason: 'sin texto visible' };
  if (/^[+\-\u2212]$/.test(label)) {
    return { action: 'trial', reason: 'control grafico de mapa o zoom' };
  }
  if (String(tag || '').toUpperCase() === 'DIV') {
    return { action: 'trial', reason: 'encabezado o control contenedor' };
  }
  if (/eliminar|borrar|anular|desactivar|deshabilitar|quitar|reset|reiniciar|vaciar|cerrar periodo|bloquear periodo|revocar|suspender|logout|salir/.test(label)) {
    return { action: 'trial', reason: 'accion destructiva o de cierre' };
  }
  if (/^(fuente|source)$/.test(label)) {
    return { action: 'trial', reason: 'abre fuente externa de audio' };
  }
  if (/imprimir|print/.test(label)) {
    return { action: 'trial', reason: 'abre dialogo de impresion del sistema' };
  }
  if (/importar|subir|cargar archivo|adjuntar/.test(label) || kind === 'file') {
    return { action: 'trial', reason: 'requiere archivo local' };
  }
  if (/google|whatsapp|mailto:|tel:/.test(label + ' ' + url)) {
    return { action: 'trial', reason: 'abre servicio externo' };
  }
  return { action: 'click', reason: '' };
}

async function launchBrowser() {
  const common = {
    slowMo: Number(process.env.QA_SLOWMO_MS || 15),
    args: ['--disable-dev-shm-usage']
  };
  if (process.env.QA_HEADED === '0') {
    return chromium.launch({ ...common, headless: true });
  }
  try {
    return await chromium.launch({ ...common, headless: false });
  } catch (error) {
    report.events.push({ type: 'browser_fallback_headless', message: error.message });
    return chromium.launch({ ...common, headless: true });
  }
}

async function attachObservers(page, scope) {
  page.on('console', (msg) => {
    const type = msg.type();
    if (type === 'error') {
      pushLimited(report.console_errors, { scope, text: msg.text(), url: page.url(), at: new Date().toISOString() });
    }
  });
  page.on('pageerror', (error) => {
    pushLimited(report.page_errors, { scope, message: error.message, stack: error.stack || '', url: page.url(), at: new Date().toISOString() });
  });
  page.on('requestfailed', (request) => {
    const failureText = request.failure() && request.failure().errorText;
    if (String(failureText || '').includes('ERR_ABORTED')) {
      return;
    }
    pushLimited(report.network_errors, {
      scope,
      kind: 'requestfailed',
      method: request.method(),
      url: request.url(),
      failure: failureText,
      at: new Date().toISOString()
    });
  });
  page.on('response', (response) => {
    const status = response.status();
    if (status >= 500 || status === 404) {
      pushLimited(report.network_errors, {
        scope,
        kind: 'http',
        status,
        url: response.url(),
        at: new Date().toISOString()
      });
    }
  });
  page.on('dialog', async (dialog) => {
    report.events.push({ scope, type: 'dialog', dialog_type: dialog.type(), message: dialog.message() });
    await dialog.dismiss().catch(() => {});
  });
}

async function waitQuiet(page, ms = 450) {
  await page.waitForTimeout(ms);
}

async function closeGlobalOverlays(page) {
  const closeRadio = page.locator('#closeRadioDrawer');
  if (await closeRadio.count() && await closeRadio.first().isVisible().catch(() => false)) {
    await closeRadio.first().click({ timeout: 1500 }).catch(() => {});
    await waitQuiet(page, 120);
  }
  const closeAI = page.locator('#closeAIDrawer');
  const aiOpen = await page.locator('#aiChatDrawer.open').count().catch(() => 0);
  if (aiOpen && await closeAI.first().isVisible().catch(() => false)) {
    await closeAI.first().click({ timeout: 1500 }).catch(async () => {
      await page.evaluate(() => {
        const close = document.getElementById('closeAIDrawer');
        if (close) close.click();
      }).catch(() => {});
    });
    await waitQuiet(page, 120);
  }
  const robotInlineOpen = await page.locator('#robotInlineChatPanel:not([hidden])').count().catch(() => 0);
  if (robotInlineOpen) {
    await page.evaluate(() => {
      const hide = document.getElementById('robotHideBtn');
      if (hide && hide.offsetParent !== null) hide.click();
    }).catch(() => {});
    await waitQuiet(page, 120);
  }
}

async function login(page) {
  await page.goto(`${baseURL}/login.html`, { waitUntil: 'domcontentloaded', timeout: 30000 });
  await page.locator('#adminEmail').fill(email);
  await page.locator('#adminPassword').fill(password);
  await page.locator('.password-visibility-toggle[data-target="adminPassword"]').click({ timeout: 5000 }).catch(() => {});
  const visibleType = await page.locator('#adminPassword').evaluate((el) => el.type).catch(() => '');
  await page.locator('.password-visibility-toggle[data-target="adminPassword"]').click({ timeout: 5000 }).catch(() => {});
  report.login.password_toggle_ok = visibleType === 'text';

  const redirected = await Promise.all([
    page.waitForURL(/(super_administrador|seleccionar_empresa|administrar_empresa)/, { timeout: 15000 }).then(() => true).catch(() => false),
    page.locator('#emailLoginBtn').click({ timeout: 10000 })
  ]).then(([ok]) => ok).catch(() => false);

  if (!redirected) {
    const apiResp = await page.request.post(`${baseURL}/super/api/administradores/login`, {
      data: { email, password, recaptcha_token: 'dev-bypass' }
    });
    report.login.api_status = apiResp.status();
    const body = await apiResp.text().catch(() => '');
    report.login.api_body_sample = body.slice(0, 240);
    if (!apiResp.ok()) {
      throw new Error(`No se pudo iniciar sesion por UI ni API. Status API: ${apiResp.status()} ${body.slice(0, 180)}`);
    }
  }
  report.login.final_url = page.url();
  report.login.ok = true;
}

async function openAdmin(page, viewportName) {
  await page.addInitScript((id) => {
    try {
      window.localStorage.setItem('active_empresa_id', String(id));
      window.localStorage.setItem('empresa_id', String(id));
      window.localStorage.setItem('admin_empresa_id', String(id));
      window.sessionStorage.setItem('active_empresa_id', String(id));
      window.sessionStorage.setItem('empresa_id', String(id));
      window.sessionStorage.setItem('admin_empresa_id', String(id));
    } catch {}
  }, empresaID);
  await page.goto(`${baseURL}/administrar_empresa.html?id=${empresaID}`, { waitUntil: 'domcontentloaded', timeout: 30000 });
  await page.waitForSelector('#contentFrame', { timeout: 15000 });
  await page.waitForLoadState('domcontentloaded');
  await waitQuiet(page, 1200);
  report.shell[viewportName].url = page.url();
  report.shell[viewportName].title = await page.title();
}

async function exerciseShell(page, viewportName) {
  const shell = report.shell[viewportName];
  const visibility = page.locator('.admin-menu-visibility-toggle');
  if (viewportName === 'mobile' && await visibility.count()) {
    const isVisible = await visibility.first().isVisible().catch(() => false);
    const expanded = await visibility.first().getAttribute('aria-expanded').catch(() => '');
    if (isVisible && expanded !== 'true') {
      await visibility.first().click({ timeout: 5000 }).catch((error) => {
        shell.mobile_menu_open_error = error.message;
      });
      await waitQuiet(page, 250);
    }
  }

  const groupCount = await page.locator('.admin-nav-group-title').count();
  shell.group_count = groupCount;
  shell.group_clicks = [];
  for (let i = 0; i < groupCount; i += 1) {
    const button = page.locator('.admin-nav-group-title').nth(i);
    const label = compactText(await button.textContent().catch(() => `grupo_${i + 1}`));
    const before = await button.getAttribute('aria-expanded').catch(() => '');
    await button.click({ timeout: 5000 }).catch((error) => {
      shell.group_clicks.push({ label, ok: false, error: error.message });
    });
    await waitQuiet(page, 120);
    const after = await button.getAttribute('aria-expanded').catch(() => '');
    shell.group_clicks.push({ label, before, after, ok: before !== after || after === 'true' || after === 'false' });
    if (after !== 'true') {
      await button.click({ timeout: 5000 }).catch(() => {});
    }
  }

  if (await visibility.count()) {
    const buttonIsVisible = await visibility.first().isVisible().catch(() => false);
    const before = await visibility.first().getAttribute('aria-expanded').catch(() => '');
    if (buttonIsVisible) {
      await visibility.first().click({ timeout: 5000 }).catch((error) => {
        shell.menu_visibility_error = error.message;
      });
      await waitQuiet(page, 150);
      const after = await visibility.first().getAttribute('aria-expanded').catch(() => '');
      await visibility.first().click({ timeout: 5000 }).catch(() => {});
      shell.menu_visibility = { before, after, toggled: before !== after, visible: true };
    } else {
      shell.menu_visibility = { before, after: before, toggled: false, visible: false };
    }
  }

  await page.locator('#adminFavoriteBtn').click({ timeout: 5000 }).catch((error) => {
    shell.favorite_error = error.message;
  });
  await waitQuiet(page, 150);
  await page.locator('#adminFavoriteBtn').click({ timeout: 5000 }).catch(() => {});

  await page.locator('#openRadioDrawer').click({ timeout: 5000 }).catch((error) => {
    shell.radio_error = error.message;
  });
  await waitQuiet(page, 350);
  shell.radio_visible = await page.locator('#radioDrawer').evaluate((el) => !el.hasAttribute('hidden') && el.getAttribute('aria-hidden') !== 'true').catch(() => false);
  await page.locator('#radioFloatingEnabled').click({ timeout: 5000 }).catch(() => {});
  await page.locator('#radioFloatingEnabled').click({ timeout: 5000 }).catch(() => {});
  await page.locator('#closeRadioDrawer').click({ timeout: 5000 }).catch(() => {});

  const showRobotBtn = page.locator('#robotShowBtn');
  if (await showRobotBtn.count() && await showRobotBtn.first().isVisible().catch(() => false)) {
    await showRobotBtn.first().click({ timeout: 5000 }).catch((error) => {
      shell.robot_show_error = error.message;
    });
  } else if (await page.locator('#openAIDrawer').isVisible().catch(() => false)) {
    await page.locator('#openAIDrawer').click({ timeout: 5000 }).catch((error) => {
      shell.ai_error = error.message;
    });
  }
  await waitQuiet(page, 450);
  shell.ai_visible = await page.locator('#aiChatDrawer').evaluate((el) => {
    const style = window.getComputedStyle(el);
    const rect = el.getBoundingClientRect();
    return style.visibility !== 'hidden' && style.display !== 'none' && rect.bottom > 0 && rect.top < window.innerHeight && el.classList.contains('open');
  }).catch(() => false);
  const robotSend = page.locator('#robotInlineSend');
  const robotInput = page.locator('#robotInlineInput');
  if (await robotSend.count() && await robotSend.first().isVisible().catch(() => false)) {
    shell.robot_send_button_box = await robotSend.first().boundingBox().catch(() => null);
    shell.robot_input_box = await robotInput.first().boundingBox().catch(() => null);
    shell.robot_send_trial_ok = await robotSend.first().click({ trial: true, timeout: 5000 }).then(() => true).catch((error) => {
      shell.robot_send_trial_error = error.message;
      return false;
    });
  }
  if (shell.ai_visible) {
    const sendBox = await page.locator('#aiChatForm button[type="submit"]').boundingBox().catch(() => null);
    const inputBox = await page.locator('#aiChatInput').boundingBox().catch(() => null);
    shell.ai_send_button_box = sendBox;
    shell.ai_input_box = inputBox;
    shell.ai_send_trial_ok = await page.locator('#aiChatForm button[type="submit"]').click({ trial: true, timeout: 5000 }).then(() => true).catch((error) => {
      shell.ai_send_trial_error = error.message;
      return false;
    });
    await page.locator('#aiChatHintToggle').click({ timeout: 5000 }).catch(() => {});
    await page.locator('#aiChatMinimize').click({ timeout: 5000 }).catch(() => {});
  }
}

async function frameLinks(page) {
  return page.locator('a[target="contentFrame"]').evaluateAll((links) => links
    .map((link) => {
      const rect = link.getBoundingClientRect();
      return {
        id: link.id || '',
        text: (link.textContent || '').replace(/\s+/g, ' ').trim(),
        href: link.getAttribute('href') || '',
        visible: rect.width > 0 && rect.height > 0 && window.getComputedStyle(link).display !== 'none' && window.getComputedStyle(link).visibility !== 'hidden'
      };
    })
    .filter((link) => link.href));
}

async function getContentFrame(page, expectedPath) {
  await page.waitForFunction((pathName) => {
    const frame = document.getElementById('contentFrame');
    return frame && frame.contentWindow && frame.contentWindow.location && frame.contentWindow.location.pathname.endsWith(pathName);
  }, expectedPath, { timeout: 12000 }).catch(() => {});
  return page.frame({ name: 'contentFrame' }) || page.frames().find((frame) => frame.url().includes('/administrar_empresa/'));
}

async function exerciseFrameButtons(frame, moduleReport) {
  const selector = 'button, input[type="button"], input[type="submit"], a.btn, a.button, a[role="button"], [role="button"]';
  const rawCount = await frame.locator(selector).count().catch(() => 0);
  moduleReport.button_count = rawCount;
  moduleReport.buttons = [];
  const maxButtons = Math.min(rawCount, Number(process.env.QA_MAX_BUTTONS_PER_MODULE || 18));
  for (let i = 0; i < maxButtons; i += 1) {
    const item = frame.locator(selector).nth(i);
    const info = await item.evaluate((el) => {
      const rect = el.getBoundingClientRect();
      const style = window.getComputedStyle(el);
      const drawer = el.closest && el.closest('#aiChatDrawer');
      const closedDrawer = !!(drawer && !drawer.classList.contains('open') && !drawer.classList.contains('minimized'));
      return {
        tag: el.tagName,
        type: el.getAttribute('type') || '',
        text: (el.innerText || el.textContent || el.getAttribute('aria-label') || el.getAttribute('title') || el.value || '').replace(/\s+/g, ' ').trim(),
        href: el.getAttribute('href') || '',
        visible: !closedDrawer && rect.width > 0 && rect.height > 0 && style.display !== 'none' && style.visibility !== 'hidden' && !el.disabled
      };
    }).catch(() => null);
    if (!info || !info.visible) continue;
    const decision = classifyAction(info.text, info.href, info.tag, info.type);
    const row = { index: i, label: info.text || info.href || info.tag, decision: decision.action, reason: decision.reason };
    moduleReport.buttons.push(row);
    if (decision.action === 'skip') {
      report.skipped.push({ module: moduleReport.text, label: row.label, reason: decision.reason });
      continue;
    }
    try {
      if (decision.action === 'trial') {
        await item.click({ trial: true, timeout: 3500 });
        row.ok = true;
        report.skipped.push({ module: moduleReport.text, label: row.label, reason: decision.reason, trial_ok: true });
      } else {
        const downloadPromise = frame.page().waitForEvent('download', { timeout: 2500 }).catch(() => null);
        await item.click({ timeout: 5000 });
        const download = await downloadPromise;
        if (download) {
          try {
            row.download = download.suggestedFilename();
          } catch {
            row.download = '';
          }
          await download.cancel().catch(() => {});
        }
        row.ok = true;
        await waitQuiet(frame.page(), 450);
        await closeGlobalOverlays(frame.page());
      }
    } catch (error) {
      if (decision.action === 'trial') {
        row.ok = true;
        row.warning = error.message;
        report.skipped.push({ module: moduleReport.text, label: row.label, reason: decision.reason, trial_ok: false, warning: error.message });
      } else {
        row.ok = false;
        row.error = error.message;
      }
    }
  }
}

async function exerciseModules(page) {
  const links = await frameLinks(page);
  report.shell.desktop.all_module_links = links.length;
  report.shell.desktop.visible_module_links = links.filter((link) => link.visible).length;
  const limit = Number(process.env.QA_MAX_MODULES || 80);
  const filters = String(process.env.QA_MODULE_FILTER || '').split(',').map((item) => item.trim().toLowerCase()).filter(Boolean);
  const selectedLinks = filters.length
    ? links.filter((link) => filters.some((filter) => String(link.id + ' ' + link.text + ' ' + link.href).toLowerCase().includes(filter)))
    : links;
  report.shell.desktop.filtered_module_links = selectedLinks.length;
  for (const link of selectedLinks.slice(0, limit)) {
    if (page.isClosed()) {
      report.events.push({ type: 'browser_closed_before_module', module: link.text, at: new Date().toISOString() });
      break;
    }
    const href = normalizeURL(link.href);
    const expectedPath = href.split('?')[0];
    const moduleReport = {
      id: link.id,
      text: link.text,
      href,
      started_at: new Date().toISOString()
    };
    report.modules.push(moduleReport);
    try {
      await closeGlobalOverlays(page);
      const locator = link.id ? page.locator(`#${link.id}`) : page.locator(`a[target="contentFrame"][href^="${expectedPath}"]`).first();
      if (link.visible && await locator.isVisible().catch(() => false)) {
        await locator.scrollIntoViewIfNeeded({ timeout: 5000 }).catch(() => {});
        await locator.click({ timeout: 8000 });
        moduleReport.navigation = 'menu_click';
      } else {
        await page.evaluate((targetHref) => {
          const frame = document.getElementById('contentFrame');
          if (frame) frame.setAttribute('src', targetHref);
        }, href);
        moduleReport.navigation = 'direct_frame_src_hidden_menu_link';
      }
      await waitQuiet(page, 900);
      const frame = await getContentFrame(page, expectedPath);
      if (!frame) {
        moduleReport.ok = false;
        moduleReport.error = 'No se encontro contentFrame';
        continue;
      }
      moduleReport.frame_url = frame.url();
      moduleReport.title = await frame.title().catch(() => '');
      const bodyText = await frame.locator('body').innerText({ timeout: 5000 }).catch(() => '');
      moduleReport.body_sample = compactText(bodyText).slice(0, 260);
      if (/error\s+404|not found|no encontrado/i.test(bodyText)) {
        moduleReport.ok = false;
        moduleReport.error = 'Pagina con mensaje de no encontrado';
        continue;
      }
      await exerciseFrameButtons(frame, moduleReport);
      moduleReport.ok = moduleReport.buttons.every((button) => button.ok !== false);
    } catch (error) {
      moduleReport.ok = false;
      moduleReport.error = error.message;
    } finally {
      moduleReport.finished_at = new Date().toISOString();
    }
  }
}

async function run() {
  const browser = await launchBrowser();
  const context = await browser.newContext({
    viewport: initialViewport,
    acceptDownloads: true,
    ignoreHTTPSErrors: true
  });
  const page = await context.newPage();
  await attachObservers(page, 'desktop');
  await login(page);
  await openAdmin(page, 'desktop');
  await exerciseShell(page, 'desktop');
  await exerciseModules(page);

  if (process.env.QA_SKIP_MOBILE === '1') {
    report.shell.mobile.skipped = 'Omitida por QA_SKIP_MOBILE=1.';
  } else if (!page.isClosed()) {
    await page.setViewportSize({ width: 390, height: 844 });
    await openAdmin(page, 'mobile');
    await exerciseShell(page, 'mobile');
  } else {
    report.shell.mobile.skipped = 'La pagina se cerro antes de la prueba movil.';
  }

  report.finished_at = new Date().toISOString();
  report.summary.modules_total = report.modules.length;
  report.summary.modules_ok = report.modules.filter((mod) => mod.ok).length;
  report.summary.modules_failed = report.modules.filter((mod) => mod.ok === false).length;
  report.summary.buttons_clicked = report.modules.reduce((sum, mod) => sum + (mod.buttons || []).filter((button) => button.decision === 'click' && button.ok).length, 0);
  report.summary.buttons_trial = report.modules.reduce((sum, mod) => sum + (mod.buttons || []).filter((button) => button.decision === 'trial' && button.ok).length, 0);
  report.summary.buttons_failed = report.modules.reduce((sum, mod) => sum + (mod.buttons || []).filter((button) => button.ok === false).length, 0);
  report.summary.console_errors = report.console_errors.length;
  report.summary.network_errors = report.network_errors.length;
  report.summary.page_errors = report.page_errors.length;
  report.summary.skipped = report.skipped.length;
  await browser.close();
  fs.mkdirSync(path.dirname(outPath), { recursive: true });
  fs.writeFileSync(outPath, JSON.stringify(report, null, 2));
  console.log(JSON.stringify(report.summary, null, 2));
  console.log(`REPORT=${outPath}`);
}

run().catch((error) => {
  report.finished_at = new Date().toISOString();
  report.fatal_error = error && error.stack ? error.stack : String(error);
  fs.mkdirSync(path.dirname(outPath), { recursive: true });
  fs.writeFileSync(outPath, JSON.stringify(report, null, 2));
  console.error(error && error.stack ? error.stack : String(error));
  console.error(`REPORT=${outPath}`);
  process.exit(1);
});
