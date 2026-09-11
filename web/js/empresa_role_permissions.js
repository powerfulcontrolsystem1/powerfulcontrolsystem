(function () {
  'use strict';

  const panel = document.getElementById('rolePermissionsPanel');
  if (!panel) return;
  const content = document.getElementById('rolePermissionsContent');
  const search = document.getElementById('rolePermissionsSearch');
  const save = document.getElementById('rolePermissionsSave');
  const close = document.getElementById('rolePermissionsClose');
  const reload = document.getElementById('rolePermissionsReload');
  const message = document.getElementById('rolePermissionsMsg');
  const summary = document.getElementById('rolePermissionsSummary');
  const actions = [['R', 'read', 'Leer'], ['C', 'create', 'Crear'], ['U', 'update', 'Actualizar'], ['D', 'delete', 'Eliminar'], ['A', 'approve', 'Aprobar']];
  const state = { empresaID: 0, roleID: 0, name: '', revision: '', sequence: 0, loading: false, saving: false, canEdit: false, refreshRequired: false, snapshot: '', modules: [], pages: [], moduleOverrides: new Map(), pageOverrides: new Map(), loadController: null };

  function escape(value) {
    return String(value == null ? '' : value).replace(/[&<>"']/g, function (ch) {
      return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[ch];
    });
  }
  function normalize(value) {
    return String(value || '').normalize('NFD').replace(/[\u0300-\u036f]/g, '').toLowerCase();
  }
  function endpoint() {
    return '/api/empresa/roles_de_usuario?empresa_id=' + encodeURIComponent(state.empresaID) + '&action=permisos&rol_id=' + encodeURIComponent(state.roleID);
  }
  function definition() {
    return {
      permisos_modulo: Array.from(state.moduleOverrides).sort(function (a, b) { return a[0].localeCompare(b[0]); }).map(function (entry) {
        const separator = entry[0].lastIndexOf(':');
        return { modulo: entry[0].slice(0, separator), accion: entry[0].slice(separator + 1), permitido: entry[1] };
      }),
      permisos_pagina: Array.from(state.pageOverrides).sort(function (a, b) { return a[0].localeCompare(b[0]); }).map(function (entry) {
        return { pagina_clave: entry[0], permitido: entry[1] };
      })
    };
  }
  function payload() { return Object.assign({ empresa_id: state.empresaID, rol_id: state.roleID, revision: state.revision }, definition()); }
  function dirty() { return !!state.snapshot && JSON.stringify(definition()) !== state.snapshot; }
  function setMessage(text, error) {
    message.textContent = text;
    message.className = 'form-help' + (error ? ' value-negative' : '');
  }
  function updateStatus() {
    const changed = dirty();
    const overrides = state.moduleOverrides.size + state.pageOverrides.size;
    const inherited = state.modules.length * actions.length + state.pages.length - overrides;
    summary.textContent = state.modules.length + ' módulos · ' + state.pages.length + ' páginas · ' + overrides + ' excepciones propias · ' + inherited + ' permisos heredados.' + (changed ? ' Hay cambios sin guardar.' : '');
    save.disabled = state.loading || state.saving || !state.canEdit || !state.snapshot || !state.revision || state.refreshRequired || !changed;
    close.disabled = state.saving;
    reload.disabled = state.loading || state.saving || !state.roleID;
    content.querySelectorAll('select').forEach(function (input) { input.disabled = state.loading || state.saving || !state.canEdit; });
  }
  async function readResponse(res) {
    const text = await res.text();
    let data;
    try { data = text ? JSON.parse(text) : {}; } catch (_) { data = null; }
    if (!res.ok) {
      const error = new Error(data && (data.error || data.message) || text || 'HTTP ' + res.status);
      error.status = res.status;
      throw error;
    }
    if (!data) throw new Error('La respuesta de permisos no es válida. Actualiza e inténtalo de nuevo.');
    return data;
  }
  async function request(url, options, controller) {
    controller = controller || new AbortController();
    const timer = setTimeout(function () { controller.pcsTimedOut = true; controller.abort(); }, 30000);
    try {
      return await readResponse(await fetch(url, Object.assign({ credentials: 'same-origin' }, options || {}, { signal: controller.signal })));
    } catch (error) {
      if (controller.pcsTimedOut) throw new Error('La solicitud superó los 30 segundos.');
      throw error;
    } finally { clearTimeout(timer); }
  }
  function choices(map, key, inherited) {
    const selected = map.has(key) ? (map.get(key) ? 'allow' : 'deny') : 'inherit';
    const inheritedLabel = 'Heredar' + (typeof inherited === 'boolean' ? (inherited ? ' (permitido)' : ' (denegado)') : ' del rol base');
    return [['inherit', inheritedLabel], ['allow', 'Permitir'], ['deny', 'Denegar']].map(function (choice) {
      return '<option value="' + choice[0] + '"' + (selected === choice[0] ? ' selected' : '') + '>' + escape(choice[1]) + '</option>';
    }).join('');
  }
  function applyFilter() {
    const query = normalize(search.value.trim());
    let visible = 0;
    content.querySelectorAll('[data-role-permission-search]').forEach(function (item) {
      item.hidden = !!query && item.dataset.rolePermissionSearch.indexOf(query) < 0;
      if (!item.hidden) visible += 1;
    });
    content.querySelectorAll('[data-role-page-group]').forEach(function (group) {
      group.hidden = !group.querySelector('[data-role-permission-search]:not([hidden])');
      if (query && !group.hidden) group.open = true;
    });
    const empty = document.getElementById('rolePermissionsEmpty');
    if (empty) empty.hidden = visible > 0;
  }
  function render(data) {
    const moduleLabels = data.modulos_etiqueta || {};
    const actionLabels = data.acciones_etiqueta || {};
    const inheritedModules = new Map((data.modulos_heredados || []).map(function (item) { return [item.modulo, item]; }));
    const inheritedPages = new Map((data.paginas_heredadas || []).map(function (item) { return [item.pagina_clave, item.permitido]; }));
    const modules = state.modules.map(function (item, index) {
      const title = moduleLabels[item.modulo] || item.modulo;
      const blob = normalize([title, item.modulo].concat(actions.map(function (a) { return actionLabels[a[0]] || a[2]; })).join(' '));
      return '<details open data-role-permission-search="' + escape(blob) + '"><summary>' + escape(title) + '</summary><div class="role-permissions-actions">' + actions.map(function (action) {
        const label = actionLabels[action[0]] || action[2];
        const base = inheritedModules.get(item.modulo);
        return '<label><span>' + escape(label) + '</span><select data-role-module="' + index + '" data-role-action="' + action[0] + '" aria-label="' + escape(label + ' en ' + title) + '">' + choices(state.moduleOverrides, item.modulo + ':' + action[0], base && base[action[1]]) + '</select><small>Vigente al cargar: ' + (item[action[1]] ? 'permitido' : 'denegado') + '</small></label>';
      }).join('') + '</div></details>';
    }).join('');
    const groups = new Map();
    state.pages.forEach(function (page, index) {
      const group = page.grupo || 'Otras páginas';
      if (!groups.has(group)) groups.set(group, []);
      groups.get(group).push({ page: page, index: index });
    });
    const pages = Array.from(groups).sort(function (a, b) { return a[0].localeCompare(b[0], 'es'); }).map(function (entry) {
      return '<details data-role-page-group><summary>' + escape(entry[0]) + ' (' + entry[1].length + ')</summary>' + entry[1].map(function (row) {
        const page = row.page;
        const title = page.titulo || page.pagina_clave;
        const moduleTitle = (Array.isArray(page.any_modules) && page.any_modules.length ? page.any_modules : [page.modulo]).filter(Boolean).map(function (mod) { return moduleLabels[mod] || mod; }).join(', ');
        const blob = normalize([entry[0], title, page.pagina_clave, moduleTitle].join(' '));
        return '<label class="role-permissions-page" data-role-permission-search="' + escape(blob) + '"><span>' + escape(title) + '<small>' + escape(moduleTitle) + ' · Vigente al cargar: ' + (page.permitido ? 'visible' : 'oculta') + '</small></span><select data-role-page="' + row.index + '" aria-label="' + escape('Visibilidad de ' + title) + '">' + choices(state.pageOverrides, page.pagina_clave, inheritedPages.get(page.pagina_clave)) + '</select></label>';
      }).join('') + '</details>';
    }).join('');
    content.innerHTML = '<h3>Acciones por módulo</h3>' + modules + '<h3>Visibilidad de páginas</h3>' + pages + '<p id="rolePermissionsEmpty" class="form-help" hidden>No hay coincidencias para esta búsqueda.</p>';
    applyFilter();
    updateStatus();
  }
  function applyData(data, context) {
    if (Number(data.empresa_id) !== state.empresaID || Number(data.rol_id) !== state.roleID) throw new Error('El contexto recibido no corresponde al rol solicitado.');
    if (typeof data.revision !== 'string' || !data.revision || !Array.isArray(data.permisos_modulo) || !Array.isArray(data.permisos_pagina)) throw new Error('La API no devolvió la versión y la herencia del rol. Recarga antes de editar.');
    state.revision = data.revision;
    state.modules = Array.isArray(data.modulos) ? data.modulos : [];
    state.pages = Array.isArray(data.paginas) ? data.paginas : [];
    state.moduleOverrides = new Map(data.permisos_modulo.map(function (item) { return [item.modulo + ':' + item.accion, !!item.permitido]; }));
    state.pageOverrides = new Map(data.permisos_pagina.map(function (item) { return [item.pagina_clave, !!item.permitido]; }));
    state.canEdit = Array.isArray(context.modulos) && context.modulos.some(function (mod) { return mod.modulo === 'seguridad' && mod.update; });
    state.snapshot = JSON.stringify(definition());
    state.refreshRequired = false;
    document.getElementById('rolePermissionsInfo').textContent = 'Empresa ' + state.empresaID + ' · Basado en ' + (data.rol_base_nombre || 'rol operativo') + '. Solo las excepciones se guardan en este perfil.';
    render(data);
  }
  async function fetchCurrent(controller) {
    try {
      return await Promise.all([
        request(endpoint(), {}, controller),
        request('/api/empresa/permisos_contexto?empresa_id=' + encodeURIComponent(state.empresaID), {}, controller)
      ]);
    } finally { controller.abort(); }
  }
  async function load(empresaID, roleID, name) {
    if (state.saving || (dirty() && !window.confirm('Hay un borrador sin guardar. ¿Descartarlo y cargar la versión vigente del rol?'))) return;
    if (state.loadController) state.loadController.abort();
    const sequence = ++state.sequence;
    const sameRole = state.empresaID === empresaID && state.roleID === roleID;
    const controller = new AbortController();
    state.loadController = controller;
    state.empresaID = empresaID;
    state.roleID = roleID;
    state.name = name;
    state.loading = true;
    if (!sameRole) {
      state.snapshot = '';
      state.revision = '';
      state.canEdit = false;
      state.refreshRequired = false;
      state.modules = [];
      state.pages = [];
      state.moduleOverrides.clear();
      state.pageOverrides.clear();
      content.innerHTML = '';
      search.value = '';
    }
    panel.hidden = false;
    document.getElementById('rolePermissionsTitle').textContent = 'Permisos: ' + name;
    document.getElementById('rolePermissionsInfo').textContent = 'Empresa ' + empresaID + ' · Rol ' + roleID;
    setMessage('Cargando permisos del rol…', false);
    updateStatus();
    panel.scrollIntoView({ behavior: 'smooth', block: 'start' });
    try {
      const [data, context] = await fetchCurrent(controller);
      if (sequence !== state.sequence) return;
      applyData(data, context);
      setMessage(state.canEdit ? 'Permisos cargados. Heredar sigue el rol base; Permitir y Denegar definen excepciones propias.' : 'Modo consulta: tu perfil no permite actualizar los permisos de seguridad.', false);
    } catch (error) {
      if (sequence === state.sequence) {
        state.refreshRequired = true;
        setMessage((error.message || 'No se pudieron cargar los permisos.') + (dirty() ? ' El borrador se conserva. Recarga para obtener la versión vigente.' : ''), true);
      }
    } finally {
      if (sequence === state.sequence) { state.loading = false; state.loadController = null; updateStatus(); }
    }
  }

  document.addEventListener('click', function (event) {
    const button = event.target.closest('.custom-role-permissions');
    if (!button || button.disabled) return;
    const empresaID = Number(button.dataset.empresaId), roleID = Number(button.dataset.roleId);
    if (Number.isSafeInteger(empresaID) && empresaID > 0 && Number.isSafeInteger(roleID) && roleID > 0) load(empresaID, roleID, button.dataset.roleName || 'Rol personalizado');
  });
  content.addEventListener('change', function (event) {
    const input = event.target;
    if (state.loading || state.saving || !state.canEdit || !['inherit', 'allow', 'deny'].includes(input.value)) return;
    let map, key;
    if (input.matches('select[data-role-module]')) {
      const action = actions.find(function (a) { return a[0] === input.dataset.roleAction; });
      const item = state.modules[Number(input.dataset.roleModule)];
      if (item && action) { map = state.moduleOverrides; key = item.modulo + ':' + action[0]; }
    } else if (input.matches('select[data-role-page]')) {
      const item = state.pages[Number(input.dataset.rolePage)];
      if (item) { map = state.pageOverrides; key = item.pagina_clave; }
    }
    if (map) { if (input.value === 'inherit') map.delete(key); else map.set(key, input.value === 'allow'); }
    updateStatus();
  });
  search.addEventListener('input', applyFilter);
  reload.addEventListener('click', function () { if (!reload.disabled) load(state.empresaID, state.roleID, state.name); });
  close.addEventListener('click', function () {
    if (state.saving || (dirty() && !window.confirm('Hay permisos sin guardar. ¿Descartar los cambios?'))) return;
    ++state.sequence;
    if (state.loadController) state.loadController.abort();
    state.snapshot = '';
    panel.hidden = true;
  });
  save.addEventListener('click', async function () {
    if (save.disabled || state.loading || state.saving || !dirty()) return;
    state.saving = true;
    updateStatus();
    setMessage('Guardando permisos del rol…', false);
    let confirmed = false;
    try {
      const body = payload();
      body.aprobado_por = 'administrador_empresa';
      body.codigo_aprobacion = 'permisos-rol-' + Date.now();
      body.motivo_aprobacion = 'Configuración de permisos del rol personalizado de empresa';
      await request(endpoint(), { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
      confirmed = true;
      state.snapshot = JSON.stringify(definition());
      state.refreshRequired = true;
      const [data, context] = await fetchCurrent(new AbortController());
      applyData(data, context);
      setMessage('Excepciones del rol guardadas. Los demás permisos siguen heredando los cambios del rol base, dentro de la empresa y de su licencia.', false);
    } catch (error) {
      if (confirmed) {
        state.refreshRequired = true;
        setMessage('El guardado se confirmó, pero no se pudo cargar la nueva versión. Recarga los permisos antes de volver a guardar.', true);
      } else if (error.status === 409 || error.status === 428) {
        state.refreshRequired = true;
        setMessage('El rol o su base cambió desde la carga. Tu borrador sigue en pantalla. Recarga la versión vigente antes de guardar; se pedirá confirmar antes de descartar el borrador.', true);
      } else if (!error.status || error.status >= 500) {
        state.refreshRequired = true;
        setMessage((error.message || 'La conexión falló.') + ' El resultado del guardado es incierto. Tu borrador se conserva; recarga para comprobar el estado antes de enviar de nuevo.', true);
      } else {
        setMessage((error.message || 'No se pudieron guardar los permisos.') + ' El borrador se conserva.', true);
      }
    } finally { state.saving = false; updateStatus(); }
  });
  window.addEventListener('beforeunload', function (event) {
    if (!dirty()) return;
    event.preventDefault();
    event.returnValue = '';
  });
})();
