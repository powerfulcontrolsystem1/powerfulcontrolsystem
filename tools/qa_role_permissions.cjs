#!/usr/bin/env node
'use strict';

// node tools/qa_role_permissions.cjs runs controller tests without a browser.
// --serve [--port 8874] serves synthetic data for manual CUA desktop/mobile QA.
// This fixture never contacts PCS, PostgreSQL, email or another network host.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const http = require('node:http');
const path = require('node:path');
const vm = require('node:vm');
const root = path.resolve(__dirname, '..');
const webRoot = path.join(root, 'web');
const htmlPaths = ['administrar_empresa/administrar_usuarios.html', 'administrar_empresa/configuracion_permisos.html', 'super/permisos_rol.html'];
const controller = fs.readFileSync(path.join(webRoot, 'js/empresa_role_permissions.js'), 'utf8');
const actions = [['R', 'read'], ['C', 'create'], ['U', 'update'], ['D', 'delete'], ['A', 'approve']];
const modules = ['ventas', 'inventario', 'finanzas', 'clientes', 'compras', 'facturacion', 'seguridad', 'reportes', 'reservas', 'domotica'];
const roleFixture = [
  { id: 1, nombre: 'cajero' }, { id: 2, nombre: 'inventario' },
  { id: 3, nombre: 'jefe_bodega', nombre_visible: 'Jefe de bodega — Hotel (ID 3)' }, { id: 4, nombre: 'responsable_bodega' },
  { id: 5, nombre: 'perfil_futuro' }, { id: 6, nombre: 'super_administrador', asignable: false },
  { id: 7, nombre: 'Cajero de auditoría', empresa_id: 12, rol_base_id: 1, personalizado: true },
  { id: 8, nombre: 'Responsable de inventario nocturno', empresa_id: 12, rol_base_id: 2, personalizado: true }
].map(function (role) { return { estado: 'activo', asignable: true, ...role }; });
function matrix(roleID) {
  const data = {
    empresa_id: 12, rol_id: roleID, rol_nombre: roleFixture.find(r => r.id === roleID)?.nombre || 'Perfil QA', rol_base_id: 1, rol_base_nombre: 'Cajero',
    revision: 'fixture-1', permisos_modulo: [], permisos_pagina: [],
    acciones_etiqueta: { R: 'Leer', C: 'Crear', U: 'Actualizar', D: 'Eliminar', A: 'Aprobar' },
    modulos_etiqueta: { ventas: 'Ventas', inventario: 'Inventario', finanzas: 'Finanzas', seguridad: 'Seguridad' },
    modulos: modules.map((modulo, index) => ({ modulo, read: true, create: index < 2, update: index < 2 || modulo === 'seguridad', delete: false, approve: false })),
    paginas: Array.from({ length: 80 }, (_, index) => ({ pagina_clave: 'linkFixture' + index, titulo: 'Página de prueba ' + (index + 1), modulo: modules[index % modules.length], grupo: 'Grupo ' + (index % 8 + 1), permitido: index < 12 }))
  };
  data.modulos_heredados = structuredClone(data.modulos);
  data.paginas_heredadas = structuredClone(data.paginas);
  applyDefinition(data, { permisos_modulo: [{ modulo: 'reportes', accion: 'R', permitido: false }], permisos_pagina: [{ pagina_clave: 'linkFixture0', permitido: false }] }, false);
  return data;
}
function applyDefinition(data, input, advance = true) {
  data.modulos = structuredClone(data.modulos_heredados);
  data.paginas = structuredClone(data.paginas_heredadas);
  data.permisos_modulo = structuredClone(input.permisos_modulo || []);
  data.permisos_pagina = structuredClone(input.permisos_pagina || []);
  for (const item of data.permisos_modulo) {
    const mod = data.modulos.find(m => m.modulo === item.modulo), action = actions.find(a => a[0] === item.accion);
    if (mod && action) mod[action[1]] = item.permitido;
  }
  for (const item of data.permisos_pagina) { const page = data.paginas.find(p => p.pagina_clave === item.pagina_clave); if (page) page.permitido = item.permitido; }
  if (advance) data.revision = 'fixture-' + (Number(data.revision.split('-').at(-1)) + 1);
}
function syntaxChecks() {
  new vm.Script(controller, { filename: 'empresa_role_permissions.js' });
  for (const relative of htmlPaths) {
    const html = fs.readFileSync(path.join(webRoot, relative), 'utf8');
    let count = 0;
    for (const script of html.matchAll(/<script\b[^>]*>([\s\S]*?)<\/script>/gi)) {
      if (script[1].trim()) new vm.Script(script[1], { filename: relative + ':inline:' + (++count) });
    }
  }
}
function catalogChecks() {
  const html = fs.readFileSync(path.join(webRoot, htmlPaths[0]), 'utf8');
  const names = ['isCustomRole', 'normalizeRoleKey', 'canonicalRoleKey', 'selectableRoles', 'dedupeRoles', 'roleLabel', 'rolNombreById'];
  const source = names.map(function (name) {
    const match = html.match(new RegExp('    function ' + name + '\\([^]*?(?=\\n    (?:async )?function |\\n    document\\.)'));
    assert.ok(match, name + ' found');
    return match[0];
  }).join('\n');
  const context = vm.createContext({ state: { roles: roleFixture } });
  vm.runInContext(source, context);
  assert.deepEqual(Array.from(context.selectableRoles(), r => r.id), [1, 2, 3, 4, 5, 7, 8], 'catalog supports future profiles and distinct warehouse role IDs');
  assert.deepEqual(Array.from(context.dedupeRoles([...roleFixture, roleFixture[2]]), r => r.id), roleFixture.map(r => r.id), 'only repeated IDs are removed');
  assert.equal(context.roleLabel(roleFixture[2]), roleFixture[2].nombre_visible, 'backend label retains company type and role ID');
  assert.equal(context.rolNombreById(3), roleFixture[2].nombre_visible, 'assigned user rows keep the same unambiguous role label');
}
function fakeElement() {
  const listeners = {};
  return {
    listeners, hidden: false, disabled: false, value: '', dataset: {}, textContent: '', innerHTML: '',
    addEventListener(name, callback) { listeners[name] = callback; },
    querySelectorAll() { return []; }, querySelector() { return null; }, scrollIntoView() {}, appendChild() {}, classList: { add() {}, remove() {} },
    matches(selector) { return selector === 'select[data-role-module]' && this.dataset.roleModule !== undefined || selector === 'select[data-role-page]' && this.dataset.rolePage !== undefined; }
  };
}
function controllerHarness() {
  const elements = new Map(), events = {}, requests = [], timers = new Map();
  const matrices = new Map([[7, matrix(7)], [8, matrix(8)]]);
  const flags = { saveStatus: 0, readOnly: false, holdLoad: null, holdSave: false, failRead: false, confirm: true, confirmations: 0, badTenant: false, missingRevision: false };
  let timerID = 0;
  const document = {
    getElementById(id) { if (!elements.has(id)) elements.set(id, fakeElement()); return elements.get(id); },
    addEventListener(name, callback) { events[name] = callback; }
  };
  const response = (status, data) => ({ ok: status < 400, status, text: async () => status === 204 ? '' : JSON.stringify(data) });
  const context = vm.createContext({
    document, AbortController,
    setTimeout(callback, ms) { const id = ++timerID; timers.set(id, { callback, ms }); return id; },
    clearTimeout(id) { timers.delete(id); },
    window: { confirm() { flags.confirmations++; return flags.confirm; }, addEventListener() {} },
    fetch: async (url, options = {}) => {
      requests.push({ url, options });
      if (!options.method && flags.holdLoad) await flags.holdLoad;
      if (!options.method && flags.failRead) throw new Error('Lectura interrumpida');
      if (options.method === 'PUT' && flags.holdSave) return new Promise((_, reject) => {
        options.signal.addEventListener('abort', () => reject(new Error('Aborted')), { once: true });
      });
      if (options.method === 'PUT' && flags.saveStatus) return response(flags.saveStatus, { error: 'Permiso denegado' });
      if (options.signal.aborted) throw new Error('Aborted');
      const id = Number(new URL(url, 'http://fixture').searchParams.get('rol_id'));
      const data = matrices.get(id) || matrix(0);
      if (options.method === 'PUT') {
        const input = JSON.parse(options.body);
        if (!input.revision) return response(428, { error: 'Falta revisión' });
        if (input.revision !== data.revision) return response(409, { error: 'Revisión obsoleta' });
        applyDefinition(data, input);
        return response(204);
      }
      const copy = structuredClone(data);
      if (flags.badTenant) copy.empresa_id = 13;
      if (flags.missingRevision) delete copy.revision;
      if (flags.readOnly && url.includes('/permisos_contexto')) copy.modulos.find(m => m.modulo === 'seguridad').update = false;
      return response(200, copy);
    }
  });
  vm.runInContext(controller, context);
  const get = id => document.getElementById(id);
  const open = id => events.click({ target: { closest: () => ({ disabled: false, dataset: { empresaId: '12', roleId: String(id), roleName: 'Perfil ' + id } }) } });
  const flush = () => new Promise(resolve => setImmediate(resolve));
  const change = (dataset, value) => { const input = fakeElement(); input.dataset = dataset; input.value = value; get('rolePermissionsContent').listeners.change({ target: input }); };
  const save = () => get('rolePermissionsSave').listeners.click();
  const reload = () => get('rolePermissionsReload').listeners.click();
  const puts = () => requests.filter(req => req.options.method === 'PUT');
  return { get, open, flush, change, save, reload, puts, requests, matrices, timers, flags };
}
async function controllerChecks() {
  const { get, open, flush, change, save, reload, puts, requests, matrices, timers, flags } = controllerHarness();
  open(7);
  await flush();
  assert.equal(get('rolePermissionsSave').disabled, true, 'unchanged matrix cannot be submitted');
  assert.match(get('rolePermissionsContent').innerHTML, /Heredar \(permitido\)/, 'base value is shown separately from the override');
  assert.match(get('rolePermissionsContent').innerHTML, /value="deny" selected/, 'stored denial remains an explicit exception');
  change({ roleModule: '0', roleAction: 'D' }, 'allow');
  change({ rolePage: '1' }, 'deny');
  assert.equal(get('rolePermissionsSave').disabled, false, 'changing an action enables save');
  get('rolePermissionsSearch').value = 'inventario';
  get('rolePermissionsSearch').listeners.input();
  flags.saveStatus = 403;
  await save();
  assert.equal(get('rolePermissionsSave').disabled, false, 'failed save retains changes for retry');
  assert.match(get('rolePermissionsMsg').textContent, /denegado/);
  flags.saveStatus = 0;
  await save();
  const saved = JSON.parse(puts().at(-1).options.body);
  assert.equal(saved.empresa_id, 12);
  assert.equal(saved.rol_id, 7);
  assert.equal(saved.revision, 'fixture-1');
  assert.deepEqual(saved.permisos_modulo, [{ modulo: 'reportes', accion: 'R', permitido: false }, { modulo: 'ventas', accion: 'D', permitido: true }], 'one action change preserves the existing override without freezing inherited actions, including filtered modules');
  assert.deepEqual(saved.permisos_pagina, [{ pagina_clave: 'linkFixture0', permitido: false }, { pagina_clave: 'linkFixture1', permitido: false }], 'page edit stores only existing and changed exceptions');
  assert.match(get('rolePermissionsMsg').textContent, /Excepciones del rol guardadas/, 'empty HTTP 204 is accepted and a fresh revision loaded');
  assert.equal(get('rolePermissionsSave').disabled, true);
  change({ roleModule: '0', roleAction: 'D' }, 'inherit');
  change({ rolePage: '1' }, 'inherit');
  await save();
  const inherited = JSON.parse(puts().at(-1).options.body);
  assert.equal(inherited.revision, 'fixture-2', 'next mutation uses the revision returned by the confirmation GET');
  assert.equal(inherited.permisos_modulo.length, 1, 'Heredar deletes the module override');
  assert.equal(inherited.permisos_pagina.length, 1, 'Heredar deletes the page override');
  const data = matrices.get(7);
  data.modulos_heredados[0].delete = true;
  data.paginas_heredadas[1].permitido = false;
  applyDefinition(data, data);
  reload();
  await flush();
  assert.match(get('rolePermissionsContent').innerHTML, /data-role-action="D"[^>]*><option value="inherit" selected>Heredar \(permitido\)/, 'a later base permission change flows into an inherited action');
  assert.match(get('rolePermissionsContent').innerHTML, /data-role-page="1"[^>]*><option value="inherit" selected>Heredar \(denegado\)/, 'a later base visibility change flows into an inherited page');
  assert.equal(get('rolePermissionsSave').disabled, true, 'base propagation does not create a draft');
  let release;
  flags.holdLoad = new Promise(resolve => { release = resolve; });
  open(8);
  const count = requests.length;
  await save();
  assert.equal(requests.length, count, 'loading a new role cannot save the previous role matrix');
  open(7);
  assert.equal(requests[count - 1].options.signal.aborted, true, 'switching role aborts the previous load');
  release();
  await flush();
  assert.match(get('rolePermissionsTitle').textContent, /Perfil 7/);
  assert.doesNotMatch(get('rolePermissionsMsg').textContent, /Aborted/, 'obsolete errors cannot replace the selected role');
  flags.holdLoad = null;
  flags.readOnly = true;
  open(7);
  await flush();
  change({ roleModule: '0', roleAction: 'D' }, 'deny');
  assert.equal(get('rolePermissionsSave').disabled, true, 'security read-only users cannot edit the role matrix');
  assert.match(get('rolePermissionsMsg').textContent, /Modo consulta/);
  assert.equal(timers.size, 0, 'completed requests clear their timeout');
  assert.ok(requests.every(req => req.options.signal), 'every API request can be aborted');
}
async function conflictChecks(status) {
  const { get, open, flush, change, save, reload, puts, matrices, flags } = controllerHarness();
  open(7);
  await flush();
  change({ roleModule: '0', roleAction: 'D' }, 'allow');
  if (status === 409) {
    applyDefinition(matrices.get(7), { permisos_modulo: [{ modulo: 'ventas', accion: 'R', permitido: false }], permisos_pagina: [] });
  } else flags.saveStatus = status;
  await save();
  assert.equal(get('rolePermissionsSave').disabled, true, status + ': conflict prevents overwriting another administrator');
  assert.match(get('rolePermissionsMsg').textContent, /borrador sigue en pantalla/);
  assert.match(get('rolePermissionsSummary').textContent, /cambios sin guardar/, 'conflict retains the draft');
  const putCount = puts().length;
  await save();
  assert.equal(puts().length, putCount, 'save cannot be retried with a stale revision');
  flags.confirm = false;
  reload();
  await flush();
  assert.equal(flags.confirmations, 1, 'refresh requires an explicit decision to discard the draft');
  assert.match(get('rolePermissionsSummary').textContent, /cambios sin guardar/);
  flags.confirm = true;
  flags.failRead = true;
  reload();
  await flush();
  assert.match(get('rolePermissionsSummary').textContent, /cambios sin guardar/, 'failed reload also preserves the draft');
  assert.equal(get('rolePermissionsSave').disabled, true);
  flags.failRead = false;
  flags.saveStatus = 0;
  reload();
  await flush();
  assert.doesNotMatch(get('rolePermissionsSummary').textContent, /cambios sin guardar/);
  change({ roleModule: '0', roleAction: 'D' }, 'allow');
  await save();
  assert.equal(JSON.parse(puts().at(-1).options.body).revision, matrices.get(7).revision === 'fixture-3' ? 'fixture-2' : 'fixture-1');
  if (status === 409) assert.ok(matrices.get(7).permisos_modulo.some(item => item.modulo === 'ventas' && item.accion === 'R' && !item.permitido), 'concurrent denial survives the renewed draft');
}
async function uncertainSaveChecks(mode) {
  const { get, open, flush, change, save, reload, puts, timers, flags } = controllerHarness();
  open(7);
  await flush();
  change({ roleModule: '0', roleAction: 'D' }, 'allow');
  flags.holdSave = mode === 'timeout';
  flags.saveStatus = mode === 'server_error' ? 500 : 0;
  if (mode === 'confirmation_failed') flags.failRead = true;
  const saving = save();
  await flush();
  if (mode === 'timeout') {
    const timeout = Array.from(timers.values()).find(timer => timer.ms === 30000);
    assert.ok(timeout, 'write has a 30 second deadline');
    timeout.callback();
  }
  await saving;
  assert.equal(get('rolePermissionsSave').disabled, true, mode + ': uncertain state cannot be submitted again');
  assert.match(get('rolePermissionsMsg').textContent, mode === 'confirmation_failed' ? /guardado se confirmó/ : /resultado del guardado es incierto/);
  const count = puts().length;
  await save();
  assert.equal(puts().length, count, 'no automatic retry or duplicate mutation');
  if (mode !== 'confirmation_failed') assert.match(get('rolePermissionsSummary').textContent, /cambios sin guardar/);
  flags.holdSave = false; flags.saveStatus = 0; flags.failRead = false;
  reload();
  await flush();
  change({ roleModule: '0', roleAction: 'D' }, mode === 'confirmation_failed' ? 'inherit' : 'allow');
  assert.equal(get('rolePermissionsSave').disabled, false, 'successful explicit reload allows editing with current revision');
  assert.equal(timers.size, 0);
}
async function invalidContractChecks(mode) {
  const { get, open, flush, change, flags } = controllerHarness();
  flags[mode] = true;
  open(7);
  await flush();
  change({ roleModule: '0', roleAction: 'D' }, 'allow');
  assert.equal(get('rolePermissionsSave').disabled, true, mode + ': tenant mismatch or missing revision fails closed');
  assert.match(get('rolePermissionsMsg').textContent, mode === 'badTenant' ? /no corresponde/ : /API no devolvió/);
}
async function superRaceChecks(firstOutcome) {
  const elements = new Map(), requests = [];
  const get = id => { if (!elements.has(id)) elements.set(id, fakeElement()); return elements.get(id); };
  let releaseFirst;
  const first = new Promise(resolve => { releaseFirst = resolve; });
  const context = vm.createContext({
    URLSearchParams, window: { location: { search: '?rol_id=7' }, confirm: () => true, addEventListener() {} },
    document: { getElementById: get, createElement: fakeElement, querySelectorAll: () => [] },
    fetch: async (url, options = {}) => {
      requests.push({ url, options });
      if (url.endsWith('rol_id=7')) {
        await first;
        if (firstOutcome === 'network_error') throw new Error('Obsolete request connection failed');
        if (firstOutcome === 'http_error') return { ok: false, status: 403, text: async () => 'Obsolete role is no longer available' };
      }
      const data = url.endsWith('/tipos_empresas') ? [] : url.endsWith('/roles_de_usuario') ? roleFixture : matrix(Number(new URL(url, 'http://fixture').searchParams.get('rol_id')));
      return { ok: true, json: async () => data, text: async () => JSON.stringify(data) };
    }
  });
  const scripts = Array.from(fs.readFileSync(path.join(webRoot, 'super/permisos_rol.html'), 'utf8').matchAll(/<script\b[^>]*>([\s\S]*?)<\/script>/gi));
  vm.runInContext(scripts[1][1], context);
  await new Promise(resolve => setImmediate(resolve));
  assert.equal(get('saveBtn').disabled, true, 'super matrix remains disabled while initial role loads');
  get('rol_id').value = '8';
  await get('rol_id').listeners.change();
  assert.equal(get('saveBtn').disabled, false);
  releaseFirst();
  await new Promise(resolve => setImmediate(resolve));
  assert.equal(get('saveBtn').disabled, false, firstOutcome + ': obsolete response must not disable the newer role');
  assert.equal(get('panelPermisos').hidden, false, firstOutcome + ': obsolete response must not hide the newer role');
  assert.equal(get('msg').textContent, '', firstOutcome + ': obsolete response must not replace the current status');
  await get('saveBtn').listeners.click();
  const mutation = requests.find(request => request.options.method === 'PUT');
  assert.equal(JSON.parse(mutation.options.body).rol_id, 8, 'late role A response cannot overwrite selected role B');
  assert.match(get('msg').textContent, /guardados correctamente/, 'save success survives reload');
}
function superFilterChecks() {
  const html = fs.readFileSync(path.join(webRoot, 'super/permisos_rol.html'), 'utf8');
  const source = ['applyPageFilter', 'collectPayload'].map(name => {
    const match = html.match(new RegExp('      function ' + name + '\\([^]*?(?=\\n      (?:async )?function )'));
    assert.ok(match, name + ' found');
    return match[0];
  }).join('\n');
  const cards = ['Ventas', 'Inventario', 'Finanzas'].map(textContent => ({ textContent, hidden: false }));
  const moduleInputs = cards.flatMap(card => actions.map(action => ({ checked: action[0] === 'R', getAttribute: name => name === 'data-modulo' ? card.textContent.toLowerCase() : action[0] })));
  const rows = cards.map(card => {
    const classes = new Set();
    return { getAttribute: () => card.textContent.toLowerCase(), classList: { add: value => classes.add(value), remove: value => classes.delete(value) }, classes };
  });
  const pageInputs = cards.map((card, index) => ({ checked: index < 2, getAttribute: () => 'link' + card.textContent }));
  const prSearch = { value: '' };
  const context = vm.createContext({
    prSearch, state: { loadedRoleID: 7 },
    modulosBox: { querySelectorAll: selector => selector === '.pcs-pr-mod-card' ? cards : moduleInputs },
    paginasBox: { querySelectorAll: selector => selector === 'tr.pcs-pr-pag-row' ? rows : selector === 'details.pcs-pr-group' ? [] : pageInputs }
  });
  vm.runInContext(source, context);
  const snapshot = JSON.stringify(context.collectPayload());
  prSearch.value = 'ventas';
  context.applyPageFilter();
  assert.deepEqual(cards.map(card => card.hidden), [false, true, true], 'panel search hides unrelated modules');
  assert.deepEqual(rows.map(row => row.classes.has('pcs-pr-pag-row--hidden')), [false, true, true], 'the same search filters pages');
  prSearch.value = 'sin coincidencias';
  context.applyPageFilter();
  assert.ok(cards.every(card => card.hidden), 'unmatched search hides every module');
  assert.ok(rows.every(row => row.classes.has('pcs-pr-pag-row--hidden')), 'unmatched search hides every page');
  assert.equal(JSON.stringify(context.collectPayload()), snapshot, 'filtering preserves all hidden module/page decisions in the save payload');
  prSearch.value = '';
  context.applyPageFilter();
  assert.ok(cards.every(card => !card.hidden));
  assert.ok(rows.every(row => !row.classes.has('pcs-pr-pag-row--hidden')));
}
async function serve() {
  const index = process.argv.indexOf('--port');
  const port = index >= 0 ? Number(process.argv[index + 1]) : 8874;
  assert.ok(Number.isInteger(port) && port > 1024 && port < 65536);
  const matrices = new Map([[7, matrix(7)], [8, matrix(8)]]);
  let ceiling = matrix(0);
  const users = [{ id: 10, empresa_id: 12, nombre: 'Persona de prueba A', email: 'qa-a@example.invalid', rol_usuario_id: 7, rol_usuario_nombre: 'Cajero de auditoría', activo: 1 }, { id: 11, empresa_id: 12, nombre: 'Persona de prueba B', email: 'qa-b@example.invalid', rol_usuario_id: 3, rol_usuario_nombre: 'jefe_bodega', activo: 1 }];
  function apply(data, input) {
    for (const item of input.permisos_modulo || []) {
      const mod = data.modulos.find(m => m.modulo === item.modulo), action = actions.find(a => a[0] === item.accion);
      if (mod && action) mod[action[1]] = item.permitido;
    }
    for (const item of input.permisos_pagina || []) { const page = data.paginas.find(p => p.pagina_clave === item.pagina_clave); if (page) page.permitido = item.permitido; }
  }
  const server = http.createServer(async (req, res) => {
    try {
      const url = new URL(req.url, 'http://127.0.0.1');
      const send = (data, status = 200) => { res.writeHead(status, { 'Content-Type': 'application/json', 'Cache-Control': 'no-store' }); res.end(JSON.stringify(data)); };
      if (url.pathname.startsWith('/api/') || url.pathname.startsWith('/super/api/')) {
        let body = '';
        for await (const chunk of req) { body += chunk; if (body.length > 1000000) { send({ error: 'Payload demasiado grande' }, 413); return; } }
        const data = body ? JSON.parse(body) : {};
        if (url.pathname.includes('roles_de_usuario') && (url.searchParams.get('action') === 'permisos' || url.pathname.endsWith('/permisos'))) {
          const id = Number(url.searchParams.get('rol_id') || data.rol_id);
          if (!matrices.has(id)) matrices.set(id, matrix(id));
          if (req.method === 'PUT') {
            const current = matrices.get(id);
            if (url.searchParams.get('action') === 'permisos') {
              if (!data.revision) { send({ error: 'Recarga para obtener la revisión' }, 428); return; }
              if (data.revision !== current.revision) { send({ error: 'Otro administrador cambió el rol o su base' }, 409); return; }
              applyDefinition(current, data);
              res.writeHead(204, { 'Cache-Control': 'no-store' }); res.end();
            } else { apply(current, data); send({ ok: true }); }
          } else send(matrices.get(id));
          return;
        }
        if (url.pathname.endsWith('/roles_de_usuario')) { send(roleFixture); return; }
        if (url.pathname.endsWith('/usuarios')) { send(users); return; }
        if (url.pathname.endsWith('/empresas')) { send({ id: 12, nombre: 'PSC · datos de prueba aislados' }); return; }
        if (url.pathname.endsWith('/tipos_empresas')) { send([{ id: 1, nombre: 'Empresa de prueba' }]); return; }
        if (url.pathname.endsWith('/permisos_contexto')) { send({ ...matrix(0), rol: 'admin_empresa', rol_efectivo: 'admin_empresa', resumen: {}, matriz_roles: roleFixture.map(r => ({ rol: r.nombre, resumen: {} })) }); return; }
        if (url.pathname.endsWith('/permisos_empresa')) { if (req.method === 'PUT') apply(ceiling, data); send(ceiling); return; }
        if (url.pathname.endsWith('/estacion_prefs')) { send([]); return; }
        send({}); return;
      }
      const file = path.resolve(webRoot, '.' + decodeURIComponent(url.pathname));
      if (!file.startsWith(webRoot + path.sep)) { res.writeHead(403); res.end(); return; }
      const ext = path.extname(file);
      let body = fs.readFileSync(file);
      if (ext === '.html') body = body.toString().replace(/(<body\b[^>]*>)/i, '$1<div style="padding:8px;background:#ffef9e;color:#292300;text-align:center" role="note">Entorno de prueba aislado · datos sintéticos · sin conexión a producción</div>');
      res.writeHead(200, { 'Content-Type': ({ '.html': 'text/html; charset=utf-8', '.js': 'text/javascript; charset=utf-8', '.css': 'text/css; charset=utf-8', '.svg': 'image/svg+xml' })[ext] || 'application/octet-stream', 'Cache-Control': 'no-store' });
      res.end(body);
    } catch (_) { res.writeHead(404); res.end('Fixture: recurso no disponible'); }
  });
  server.listen(port, '127.0.0.1', () => process.stdout.write('Fixture local: http://127.0.0.1:' + port + '/administrar_empresa/administrar_usuarios.html?empresa_id=12\n'));
}
(async function () {
  syntaxChecks();
  catalogChecks();
  await controllerChecks();
  for (const status of [409, 428]) await conflictChecks(status);
  for (const mode of ['timeout', 'server_error', 'confirmation_failed']) await uncertainSaveChecks(mode);
  for (const mode of ['badTenant', 'missingRevision']) await invalidContractChecks(mode);
  for (const outcome of ['success', 'http_error', 'network_error']) await superRaceChecks(outcome);
  superFilterChecks();
  process.stdout.write('PASS: sintaxis, catálogo por ID/tipo, herencia explícita y overrides dispersos, revisión/409/428, borrador conservado, timeout30s sin duplicación, tenant/consulta, aborto de carga y respuestas super fuera de orden.\n');
  if (process.argv.includes('--serve')) await serve();
})().catch(error => { process.stderr.write(error.stack + '\n'); process.exitCode = 1; });
