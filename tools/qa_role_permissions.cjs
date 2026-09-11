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
  { id: 3, nombre: 'jefe_bodega' }, { id: 4, nombre: 'responsable_bodega' },
  { id: 5, nombre: 'perfil_futuro' }, { id: 6, nombre: 'super_administrador', asignable: false },
  { id: 7, nombre: 'Cajero de auditoría', empresa_id: 12, rol_base_id: 1, personalizado: true },
  { id: 8, nombre: 'Responsable de inventario nocturno', empresa_id: 12, rol_base_id: 2, personalizado: true }
].map(function (role) { return { estado: 'activo', asignable: true, ...role }; });
function matrix(roleID) {
  return {
    empresa_id: 12, rol_id: roleID, rol_nombre: roleFixture.find(r => r.id === roleID)?.nombre || 'Perfil QA', rol_base_id: 1, rol_base_nombre: 'Cajero',
    acciones_etiqueta: { R: 'Leer', C: 'Crear', U: 'Actualizar', D: 'Eliminar', A: 'Aprobar' },
    modulos_etiqueta: { ventas: 'Ventas', inventario: 'Inventario', finanzas: 'Finanzas', seguridad: 'Seguridad' },
    modulos: modules.map((modulo, index) => ({ modulo, read: true, create: index < 2, update: index < 2 || modulo === 'seguridad', delete: false, approve: false })),
    paginas: Array.from({ length: 80 }, (_, index) => ({ pagina_clave: 'linkFixture' + index, titulo: 'Página de prueba ' + (index + 1), modulo: modules[index % modules.length], grupo: 'Grupo ' + (index % 8 + 1), permitido: index < 12 }))
  };
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
  const names = ['isCustomRole', 'normalizeRoleKey', 'canonicalRoleKey', 'selectableRoles', 'dedupeRoles'];
  const source = names.map(function (name) {
    const match = html.match(new RegExp('    function ' + name + '\\([^]*?(?=\\n    (?:async )?function |\\n    document\\.)'));
    assert.ok(match, name + ' found');
    return match[0];
  }).join('\n');
  const context = vm.createContext({ state: { roles: roleFixture } });
  vm.runInContext(source, context);
  assert.deepEqual(Array.from(context.selectableRoles(), r => r.id), [1, 2, 3, 4, 5, 7, 8], 'catalog supports future profiles and distinct warehouse role IDs');
  assert.deepEqual(Array.from(context.dedupeRoles([...roleFixture, roleFixture[2]]), r => r.id), roleFixture.map(r => r.id), 'only repeated IDs are removed');
}
function fakeElement() {
  const listeners = {};
  return {
    listeners, hidden: false, disabled: false, value: '', dataset: {}, textContent: '', innerHTML: '',
    addEventListener(name, callback) { listeners[name] = callback; },
    querySelectorAll() { return []; }, querySelector() { return null; }, scrollIntoView() {}, appendChild() {}, classList: { add() {}, remove() {} },
    matches(selector) { return selector === 'input[data-role-module]' && this.dataset.roleModule !== undefined; }
  };
}
async function controllerChecks() {
  const elements = new Map();
  const events = {};
  const requests = [];
  let failSave = false;
  let readOnly = false;
  let holdLoad = null;
  const document = {
    getElementById(id) { if (!elements.has(id)) elements.set(id, fakeElement()); return elements.get(id); },
    addEventListener(name, callback) { events[name] = callback; }
  };
  const context = vm.createContext({ document, window: { confirm: () => true, addEventListener() {} }, fetch: async (url, options = {}) => {
    requests.push({ url, options });
    if (!options.method && holdLoad) await holdLoad;
    const ok = !(options.method === 'PUT' && failSave);
    const data = matrix(Number(new URL(url, 'http://fixture').searchParams.get('rol_id')));
    if (readOnly && url.includes('/permisos_contexto')) data.modulos.find(m => m.modulo === 'seguridad').update = false;
    return { ok, status: ok ? 200 : 403, text: async () => JSON.stringify(ok ? options.method === 'PUT' ? { ok: true } : data : { error: 'Permiso denegado' }) };
  } });
  vm.runInContext(controller, context);
  const get = id => document.getElementById(id);
  const open = id => events.click({ target: { closest: () => ({ disabled: false, dataset: { empresaId: '12', roleId: String(id), roleName: 'Perfil ' + id } }) } });
  const flush = () => new Promise(resolve => setImmediate(resolve));
  open(7);
  await flush();
  assert.equal(get('rolePermissionsSave').disabled, true, 'unchanged matrix cannot be submitted');
  const input = fakeElement();
  input.dataset = { roleModule: '0', roleAction: 'D' };
  input.checked = true;
  get('rolePermissionsContent').listeners.change({ target: input });
  assert.equal(get('rolePermissionsSave').disabled, false, 'changing an action enables save');
  get('rolePermissionsSearch').value = 'inventario';
  get('rolePermissionsSearch').listeners.input();
  failSave = true;
  await get('rolePermissionsSave').listeners.click();
  assert.equal(get('rolePermissionsSave').disabled, false, 'failed save retains changes for retry');
  assert.match(get('rolePermissionsMsg').textContent, /denegado/);
  failSave = false;
  await get('rolePermissionsSave').listeners.click();
  const saved = JSON.parse(requests.at(-1).options.body);
  assert.equal(saved.empresa_id, 12);
  assert.equal(saved.rol_id, 7);
  assert.equal(saved.permisos_modulo.length, modules.length * 5, 'search does not remove hidden permissions from payload');
  assert.equal(saved.permisos_pagina.length, 80);
  assert.equal(saved.permisos_modulo.find(p => p.modulo === 'ventas' && p.accion === 'D').permitido, true);
  assert.equal(get('rolePermissionsSave').disabled, true);
  let release;
  holdLoad = new Promise(resolve => { release = resolve; });
  open(8);
  const count = requests.length;
  await get('rolePermissionsSave').listeners.click();
  assert.equal(requests.length, count, 'loading a new role cannot save the previous role matrix');
  release();
  await flush();
  assert.match(get('rolePermissionsTitle').textContent, /Perfil 8/);
  holdLoad = null;
  readOnly = true;
  open(7);
  await flush();
  get('rolePermissionsContent').listeners.change({ target: input });
  assert.equal(get('rolePermissionsSave').disabled, true, 'security read-only users cannot edit the role matrix');
  assert.match(get('rolePermissionsMsg').textContent, /Modo consulta/);
}
async function superRaceChecks() {
  const elements = new Map(), requests = [];
  const get = id => { if (!elements.has(id)) elements.set(id, fakeElement()); return elements.get(id); };
  let releaseFirst;
  const first = new Promise(resolve => { releaseFirst = resolve; });
  const context = vm.createContext({
    URLSearchParams, window: { location: { search: '?rol_id=7' }, confirm: () => true, addEventListener() {} },
    document: { getElementById: get, createElement: fakeElement, querySelectorAll: () => [] },
    fetch: async (url, options = {}) => {
      requests.push({ url, options });
      if (url.endsWith('rol_id=7')) await first;
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
  await get('saveBtn').listeners.click();
  const mutation = requests.find(request => request.options.method === 'PUT');
  assert.equal(JSON.parse(mutation.options.body).rol_id, 8, 'late role A response cannot overwrite selected role B');
  assert.match(get('msg').textContent, /guardados correctamente/, 'save success survives reload');
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
          if (req.method === 'PUT') { apply(matrices.get(id), data); send({ ok: true }); } else send(matrices.get(id));
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
  await superRaceChecks();
  process.stdout.write('PASS: sintaxis, catálogo escalable por ID, tenant/rol del payload, guardado fallido, búsqueda, protección durante carga y respuesta super fuera de orden.\n');
  if (process.argv.includes('--serve')) await serve();
})().catch(error => { process.stderr.write(error.stack + '\n'); process.exitCode = 1; });
