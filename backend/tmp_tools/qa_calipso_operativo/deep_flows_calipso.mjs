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
const outPath = process.env.QA_REPORT_PATH || path.join(scriptDir, 'deep_flows_calipso_report.json');

if (!email || !password) {
  throw new Error('Define QA_ADMIN_EMAIL y QA_ADMIN_PASSWORD.');
}

const stamp = new Date().toISOString().replace(/[-:.TZ]/g, '').slice(0, 14);
const qaPrefix = `QA${stamp}`;
const fixturesDir = path.join(scriptDir, 'fixtures');

const report = {
  empresa_id: empresaID,
  base_url: baseURL,
  qa_prefix: qaPrefix,
  started_at: new Date().toISOString(),
  login: {},
  steps: [],
  console_errors: [],
  page_errors: [],
  network_errors: [],
  residual_risks: [],
  summary: {}
};

function ensureFixtures() {
  fs.mkdirSync(fixturesDir, { recursive: true });
  const png = path.join(fixturesDir, 'qa_producto.png');
  const csv = path.join(fixturesDir, 'qa_import.csv');
  if (!fs.existsSync(png)) {
    fs.writeFileSync(png, Buffer.from(
      'iVBORw0KGgoAAAANSUhEUgAAAGQAAABkCAYAAABw4pVUAAAACXBIWXMAAAsTAAALEwEAmpwYAAABFElEQVR4nO3csQ3CMBQAQWn+Qx2EwAGYgRHYgTnYhKmo6B4tqkT8c0jvVQBA8j3v9wF8JjYxwzZm2zX2g2r7vV9VnC9m3m8wIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg4xIg7xGl7tBGR0QHq/AAAAAElFTkSuQmCC',
      'base64'
    ));
  }
  if (!fs.existsSync(csv)) {
    fs.writeFileSync(csv, 'codigo,nombre,valor\nQA-001,Producto QA,12000\n', 'utf8');
  }
  return { png, csv };
}

function pushLimited(list, item, limit = 400) {
  if (list.length < limit) list.push(item);
}

function okStep(name, data = {}) {
  report.steps.push({ name, ok: true, ...data });
}

function failStep(name, error, data = {}) {
  report.steps.push({ name, ok: false, error: error && error.message ? error.message : String(error), ...data });
}

async function launchBrowser() {
  return chromium.launch({
    headless: process.env.QA_HEADED === '1' ? false : true,
    slowMo: Number(process.env.QA_SLOWMO_MS || 10),
    args: ['--disable-dev-shm-usage']
  });
}

async function attachObservers(page, scope) {
  page.on('console', (msg) => {
    if (msg.type() === 'error') {
      if (msg.text().includes('status of 503') && page.url().includes('/super_administrador.html')) return;
      pushLimited(report.console_errors, { scope, text: msg.text(), url: page.url(), at: new Date().toISOString() });
    }
  });
  page.on('pageerror', (error) => {
    pushLimited(report.page_errors, { scope, message: error.message, stack: error.stack || '', url: page.url(), at: new Date().toISOString() });
  });
  page.on('requestfailed', (request) => {
    const failure = request.failure() && request.failure().errorText;
    if (String(failure || '').includes('ERR_ABORTED')) return;
    pushLimited(report.network_errors, { scope, kind: 'requestfailed', method: request.method(), url: request.url(), failure, at: new Date().toISOString() });
  });
  page.on('response', (response) => {
    const status = response.status();
    if (status === 503 && response.url().includes('/super/api/vps/procesos')) return;
    if (status >= 500 || status === 404) {
      pushLimited(report.network_errors, { scope, kind: 'http', status, url: response.url(), at: new Date().toISOString() });
    }
  });
  page.on('dialog', async (dialog) => {
    report.steps.push({ name: 'dialog', ok: true, scope, dialog_type: dialog.type(), message: dialog.message() });
    await dialog.accept().catch(() => {});
  });
}

async function login(page) {
  await page.goto(`${baseURL}/login.html`, { waitUntil: 'domcontentloaded', timeout: 30000 });
  await page.locator('#adminEmail').fill(email);
  await page.locator('#adminPassword').fill(password);
  const redirected = await Promise.all([
    page.waitForURL(/(super_administrador|seleccionar_empresa|administrar_empresa)/, { timeout: 15000 }).then(() => true).catch(() => false),
    page.locator('#emailLoginBtn').click({ timeout: 10000 })
  ]).then(([ok]) => ok).catch(() => false);
  if (!redirected) {
    const res = await page.request.post(`${baseURL}/super/api/administradores/login`, {
      data: { email, password, recaptcha_token: 'dev-bypass' }
    });
    const text = await res.text().catch(() => '');
    report.login.api_status = res.status();
    if (!res.ok()) throw new Error(`Login API fallo: ${res.status()} ${text.slice(0, 200)}`);
  }
  await page.addInitScript((id) => {
    try {
      localStorage.setItem('empresa_id', String(id));
      localStorage.setItem('active_empresa_id', String(id));
      sessionStorage.setItem('empresa_id', String(id));
      sessionStorage.setItem('active_empresa_id', String(id));
    } catch {}
  }, empresaID);
  report.login.ok = true;
  report.login.final_url = page.url();
}

async function api(page, endpoint, action, { method = 'GET', data = undefined, extra = '' } = {}) {
  const sep = endpoint.includes('?') ? '&' : '?';
  const url = `${baseURL}${endpoint}${sep}empresa_id=${encodeURIComponent(empresaID)}${action ? `&action=${encodeURIComponent(action)}` : ''}${extra || ''}`;
  const res = await page.request.fetch(url, {
    method,
    data,
    headers: data ? { 'Content-Type': 'application/json' } : undefined
  });
  const text = await res.text().catch(() => '');
  let body = null;
  try { body = text ? JSON.parse(text) : null; } catch { body = text; }
  if (!res.ok()) {
    const error = new Error(`${method} ${endpoint} action=${action || ''} HTTP ${res.status()} ${String(text).slice(0, 220)}`);
    error.status = res.status();
    error.body = body;
    throw error;
  }
  return body;
}

async function flowParqueadero(page) {
  const plate = `${qaPrefix.slice(0, 2)}${stamp.slice(-4)}`.slice(0, 6).toUpperCase();
  const ticket = await api(page, '/api/empresa/parqueadero', 'entrada', {
    method: 'POST',
    data: {
      placa: plate,
      tipo_vehiculo: 'carro',
      cliente_nombre: 'Cliente QA Motel Calipso',
      cliente_documento: `CC-${stamp}`,
      observaciones: 'Ticket QA automatizado'
    }
  });
  if (!ticket || !ticket.id || !ticket.qr_token) throw new Error('Ticket de parqueadero sin id o token QR.');
  const calculo = await api(page, '/api/empresa/parqueadero', 'calcular', {
    method: 'POST',
    data: { ticket_id: ticket.id }
  });
  if (!calculo || !calculo.cobro) throw new Error('No se obtuvo cobro automatico.');
  const publico = await page.request.get(`${baseURL}/api/public/parqueadero?empresa_id=${empresaID}&action=validar_salida&token=${encodeURIComponent(ticket.qr_token)}`);
  if (!publico.ok()) throw new Error(`Validacion publica QR parqueadero HTTP ${publico.status()}`);
  const cierre = await api(page, '/api/empresa/parqueadero', 'cobrar_salida', {
    method: 'POST',
    data: { token: ticket.qr_token, metodo_pago: 'efectivo' }
  });
  if (!cierre || !cierre.ticket || !['salido', 'cerrado'].includes(String(cierre.ticket.estado || '').toLowerCase())) throw new Error('El ticket QA no cerro correctamente.');

  const ticketAnular = await api(page, '/api/empresa/parqueadero', 'entrada', {
    method: 'POST',
    data: { placa: `${plate}A`.slice(0, 6), tipo_vehiculo: 'moto', cliente_nombre: 'Cliente QA anulacion' }
  });
  await api(page, '/api/empresa/parqueadero', 'anular', {
    method: 'POST',
    data: { ticket_id: ticketAnular.id, motivo: 'Anulacion QA controlada' }
  });
  okStep('parqueadero_ticket_qr_cobro_anulacion', {
    ticket_id: ticket.id,
    ticket_anulado_id: ticketAnular.id,
    total: cierre.cobro && cierre.cobro.total
  });
}

async function flowWMS(page) {
  await api(page, '/api/empresa/logistica_wms', 'seed_demo', { method: 'POST', data: {} });
  const ubicacionCode = `QA-WMS-${stamp.slice(-6)}`;
  await api(page, '/api/empresa/logistica_wms', 'ubicacion', {
    method: 'POST',
    data: { codigo: ubicacionCode, bodega: 'Calipso QA', zona: 'QA', pasillo: 'P1', rack: 'R1', nivel: 'N1', posicion: '01', tipo: 'picking', capacidad: 100, ocupacion: 0, estado: 'activa', observaciones: 'Ubicacion QA automatizada' }
  });
  const ordenCode = `WMS-QA-${stamp}`;
  const orden = await api(page, '/api/empresa/logistica_wms', 'orden', {
    method: 'POST',
    data: { codigo: ordenCode, tipo: 'picking', origen_documento: `QA-${stamp}`, cliente: 'Motel Calipso QA', tercero: 'Motel Calipso QA', fecha_compromiso: '2026-05-07', prioridad: 'alta', responsable: email, estado: 'liberada', observaciones: 'Orden QA' }
  });
  const item = await api(page, '/api/empresa/logistica_wms', 'item', {
    method: 'POST',
    data: { orden_id: orden.id, producto_nombre: 'Kit QA amenities', sku: `SKU-${stamp}`, ubicacion_origen: ubicacionCode, ubicacion_destino: 'PACK-QA', lote: `L-${stamp}`, serial: `S-${stamp}`, cantidad_solicitada: 2, estado: 'pendiente' }
  });
  await api(page, '/api/empresa/logistica_wms', 'avance_item', {
    method: 'POST',
    data: { id: item.id, cantidad_pickeada: 2, cantidad_empacada: 2, estado: 'completado' }
  });
  const despacho = await api(page, '/api/empresa/logistica_wms', 'despacho', {
    method: 'POST',
    data: { orden_id: orden.id, codigo: `DSP-QA-${stamp}`, transportadora: 'Mensajeria QA', guia: `GUIA-${stamp}`, conductor: 'Conductor QA', vehiculo: 'QA-001', ruta: 'Ruta Calipso', estado: 'en_ruta', fecha_salida: '2026-05-07', costo_flete: 12000 }
  });
  okStep('wms_ubicacion_orden_item_avance_despacho', { ubicacion: ubicacionCode, orden_id: orden.id, item_id: item.id, despacho_id: despacho.id });
}

async function flowCentrosCosto(page) {
  await api(page, '/api/empresa/centros_costo', 'seed_demo', { method: 'POST', data: {}, extra: '&periodo=2026-05' });
  const code = `CC-QA-${stamp.slice(-6)}`;
  const centro = await api(page, '/api/empresa/centros_costo', 'centro', {
    method: 'POST',
    extra: '&periodo=2026-05',
    data: { codigo: code, nombre: 'Centro QA Motel Calipso', tipo: 'proyecto', sucursal: 'Motel Calipso', area: 'QA', unidad_negocio: 'Pruebas', responsable: email, meta_margen_pct: 35, estado: 'activo', observaciones: 'Centro creado por QA automatizado' }
  });
  const presupuesto = await api(page, '/api/empresa/centros_costo', 'presupuesto', {
    method: 'POST',
    extra: '&periodo=2026-05',
    data: { centro_costo_codigo: code, periodo: '2026-05', escenario: 'base', ingresos_presupuesto: 2500000, egresos_presupuesto: 1200000, meta_margen_pct: 35, responsable: email, estado: 'activo' }
  });
  const regla = await api(page, '/api/empresa/centros_costo', 'regla', {
    method: 'POST',
    extra: '&periodo=2026-05',
    data: { centro_costo_codigo: code, origen_modulo: 'general', nombre: 'Regla QA Motel Calipso', categoria: 'QA', cuenta_patron: '4135', tercero_patron: 'Motel Calipso', porcentaje: 100, prioridad: 10, activa: true, estado: 'activo' }
  });
  const dashboard = await api(page, '/api/empresa/centros_costo', 'dashboard', { extra: '&periodo=2026-05' });
  const found = (dashboard.centros || []).some((x) => x.codigo === code);
  if (!found) throw new Error('El centro QA no aparece en dashboard.');
  okStep('centros_costo_centro_presupuesto_regla_dashboard', { centro_id: centro.id, presupuesto_id: presupuesto.id, regla_id: regla.id, codigo: code });
}

async function flowActivosFijos(page) {
  const code = `AF-QA-${stamp.slice(-6)}`;
  const activo = await api(page, '/api/empresa/activos_fijos_niif_fiscal', 'activo', {
    method: 'POST',
    data: {
      codigo: code,
      nombre: 'Activo QA Motel Calipso',
      categoria: 'equipo_computo',
      serial: `SER-${stamp}`,
      placa: `PL-${stamp.slice(-6)}`,
      fecha_compra: '2026-05-07',
      costo: 1800000,
      valor_residual: 100000,
      vida_util_meses: 36,
      vida_util_fiscal_meses: 36,
      metodo_depreciacion: 'linea_recta',
      metodo_depreciacion_fiscal: 'linea_recta',
      cuenta_activo: '152805',
      cuenta_depreciacion: '159205',
      cuenta_gasto: '516020',
      ubicacion: 'Recepcion QA',
      responsable: email,
      centro_costo: 'CC-QA',
      proveedor: 'Proveedor QA'
    }
  });
  await api(page, '/api/empresa/activos_fijos_niif_fiscal', 'depreciacion', { method: 'POST', data: { periodo: '2026-05' } });
  await api(page, '/api/empresa/activos_fijos_niif_fiscal', 'evento', {
    method: 'POST',
    data: { activo_id: activo.id, tipo: 'traslado', fecha_evento: '2026-05-07', valor: 0, ubicacion_destino: 'Bodega QA', responsable_destino: email, detalle: 'Evento QA automatizado', estado: 'cerrado' }
  });
  okStep('activos_fijos_alta_depreciacion_evento', { activo_id: activo.id, codigo: code });
}

async function flowRedSocialArchivo(page, fixtures) {
  await page.goto(`${baseURL}/administrar_empresa/publicar_red_social.html?empresa_id=${empresaID}`, { waitUntil: 'domcontentloaded', timeout: 30000 });
  await page.locator('#pubFotoFile').setInputFiles(fixtures.png);
  await page.locator('#btnSubirFoto').click({ timeout: 10000 });
  await page.waitForFunction(() => document.querySelector('#pubFoto') && document.querySelector('#pubFoto').value.includes('/uploads/red_social/'), null, { timeout: 15000 });
  const uploadedURL = await page.locator('#pubFoto').inputValue();
  await page.locator('#pubNombre').fill(`Publicacion QA ${stamp}`);
  await page.locator('#pubDesc').fill('Publicacion automatizada de prueba para Motel Calipso.');
  await page.locator('button[onclick="crearPublicacion()"]').click({ timeout: 10000 });
  await page.waitForTimeout(1000);
  okStep('red_social_upload_imagen_y_publicacion', { uploaded_url: uploadedURL });
}

async function flowIntegraciones(page) {
  const catalog = await page.request.get(`${baseURL}/visualizar_productos_y_precios_publico.html?empresa_slug=motel-calipso`);
  if (!catalog.ok()) throw new Error(`Carta publica Motel Calipso HTTP ${catalog.status()}`);
  const venta = await page.request.get(`${baseURL}/venta_publica.html?empresa_slug=motel-calipso`);
  if (!venta.ok()) throw new Error(`Venta publica Motel Calipso HTTP ${venta.status()}`);
  const parqueaderoPublicoSinToken = await page.request.get(`${baseURL}/api/public/parqueadero?empresa_id=${empresaID}&action=validar_salida&token=qa-token-inexistente`);
  if (![400, 404].includes(parqueaderoPublicoSinToken.status())) {
    throw new Error(`Validacion de QR inexistente respondio ${parqueaderoPublicoSinToken.status()}, se esperaba rechazo controlado.`);
  }
  await page.goto(`${baseURL}/administrar_empresa/taxi_system.html?empresa_id=${empresaID}`, { waitUntil: 'domcontentloaded', timeout: 30000 });
  await page.waitForSelector('body', { timeout: 10000 });
  const taxiBody = await page.locator('body').innerText({ timeout: 10000 });
  if (!/gps|mapa|conductor|veh/i.test(taxiBody)) throw new Error('Taxi system no muestra elementos de mapa/GPS/conductores.');
  report.residual_risks.push('No se valido hardware fisico de impresora, sensores electricos, GPS real ni pasarelas/DIAN en produccion porque requieren credenciales/dispositivos externos.');
  okStep('integraciones_publicas_qr_mapa_riesgos', { catalog_status: catalog.status(), venta_status: venta.status(), qr_invalido_status: parqueaderoPublicoSinToken.status() });
}

async function run() {
  const fixtures = ensureFixtures();
  const browser = await launchBrowser();
  const context = await browser.newContext({ viewport: { width: 1440, height: 920 }, acceptDownloads: true });
  const page = await context.newPage();
  await attachObservers(page, 'deep');
  try {
    await login(page);
    for (const [name, fn] of [
      ['parqueadero', flowParqueadero],
      ['wms', flowWMS],
      ['centros_costo', flowCentrosCosto],
      ['activos_fijos', flowActivosFijos],
      ['red_social_archivo', (p) => flowRedSocialArchivo(p, fixtures)],
      ['integraciones', flowIntegraciones]
    ]) {
      try {
        await fn(page);
      } catch (error) {
        failStep(name, error);
      }
    }
  } finally {
    await browser.close().catch(() => {});
  }
  report.finished_at = new Date().toISOString();
  report.summary = {
    steps_total: report.steps.filter((s) => s.name !== 'dialog').length,
    steps_ok: report.steps.filter((s) => s.name !== 'dialog' && s.ok).length,
    steps_failed: report.steps.filter((s) => s.name !== 'dialog' && !s.ok).length,
    console_errors: report.console_errors.length,
    page_errors: report.page_errors.length,
    network_errors: report.network_errors.length,
    residual_risks: report.residual_risks.length
  };
  fs.writeFileSync(outPath, JSON.stringify(report, null, 2));
  console.log(JSON.stringify(report.summary, null, 2));
  console.log(`REPORT=${outPath}`);
  if (report.summary.steps_failed > 0 || report.summary.page_errors > 0 || report.summary.network_errors > 0) {
    process.exitCode = 1;
  }
}

run().catch((error) => {
  report.finished_at = new Date().toISOString();
  report.fatal = error.message;
  fs.writeFileSync(outPath, JSON.stringify(report, null, 2));
  console.error(error);
  process.exit(1);
});
