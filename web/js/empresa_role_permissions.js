(function () {
  'use strict';

  const panel = document.getElementById('rolePermissionsPanel');
  if (!panel) return;
  const content = document.getElementById('rolePermissionsContent');
  const search = document.getElementById('rolePermissionsSearch');
  const save = document.getElementById('rolePermissionsSave');
  const close = document.getElementById('rolePermissionsClose');
  const message = document.getElementById('rolePermissionsMsg');
  const summary = document.getElementById('rolePermissionsSummary');
  const actions = [['R', 'read', 'Leer'], ['C', 'create', 'Crear'], ['U', 'update', 'Actualizar'], ['D', 'delete', 'Eliminar'], ['A', 'approve', 'Aprobar']];
  const state = { empresaID: 0, roleID: 0, sequence: 0, loading: false, saving: false, canEdit: false, snapshot: '', modules: [], pages: [] };

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
  function payload() {
    return {
      empresa_id: state.empresaID,
      rol_id: state.roleID,
      permisos_modulo: state.modules.flatMap(function (item) {
        return actions.map(function (action) { return { modulo: item.modulo, accion: action[0], permitido: !!item[action[1]] }; });
      }),
      permisos_pagina: state.pages.map(function (item) { return { pagina_clave: item.pagina_clave, permitido: !!item.permitido }; })
    };
  }
  function dirty() { return !!state.snapshot && JSON.stringify(payload()) !== state.snapshot; }
  function setMessage(text, error) {
    message.textContent = text;
    message.className = 'form-help' + (error ? ' value-negative' : '');
  }
  function updateStatus() {
    const changed = dirty();
    const enabled = state.modules.reduce(function (n, item) { return n + actions.filter(function (a) { return item[a[1]]; }).length; }, 0);
    summary.textContent = state.modules.length + ' módulos · ' + enabled + ' acciones activas · ' + state.pages.filter(function (p) { return p.permitido; }).length + ' páginas visibles.' + (changed ? ' Hay cambios sin guardar.' : '');
    save.disabled = state.loading || state.saving || !state.canEdit || !state.snapshot || !changed;
    close.disabled = state.saving;
    content.querySelectorAll('input').forEach(function (input) { input.disabled = state.loading || state.saving || !state.canEdit; });
  }
  async function readResponse(res) {
    const text = await res.text();
    let data;
    try { data = text ? JSON.parse(text) : {}; } catch (_) { data = null; }
    if (!res.ok) throw new Error(data && (data.error || data.message) || text || 'HTTP ' + res.status);
    if (!data) throw new Error('La respuesta de permisos no es válida. Actualiza e inténtalo de nuevo.');
    return data;
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
    const modules = state.modules.map(function (item, index) {
      const title = moduleLabels[item.modulo] || item.modulo;
      const blob = normalize([title, item.modulo].concat(actions.map(function (a) { return actionLabels[a[0]] || a[2]; })).join(' '));
      return '<details open data-role-permission-search="' + escape(blob) + '"><summary>' + escape(title) + '</summary><div class="role-permissions-actions">' + actions.map(function (action) {
        const label = actionLabels[action[0]] || action[2];
        return '<label><input type="checkbox" data-role-module="' + index + '" data-role-action="' + action[0] + '" aria-label="' + escape(label + ' en ' + title) + '"' + (item[action[1]] ? ' checked' : '') + '><span>' + escape(label) + '</span></label>';
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
        return '<label class="role-permissions-page" data-role-permission-search="' + escape(blob) + '"><input type="checkbox" data-role-page="' + row.index + '" aria-label="' + escape('Mostrar ' + title) + '"' + (page.permitido ? ' checked' : '') + '><span>' + escape(title) + '<small>' + escape(moduleTitle) + '</small></span></label>';
      }).join('') + '</details>';
    }).join('');
    content.innerHTML = '<h3>Acciones por módulo</h3>' + modules + '<h3>Visibilidad de páginas</h3>' + pages + '<p id="rolePermissionsEmpty" class="form-help" hidden>No hay coincidencias para esta búsqueda.</p>';
    applyFilter();
    updateStatus();
  }
  async function load(empresaID, roleID, name) {
    if (state.saving || (dirty() && !window.confirm('Hay permisos sin guardar. ¿Descartar estos cambios y abrir otro rol?'))) return;
    const sequence = ++state.sequence;
    state.empresaID = empresaID;
    state.roleID = roleID;
    state.snapshot = '';
    state.loading = true;
    state.canEdit = false;
    state.modules = [];
    state.pages = [];
    panel.hidden = false;
    content.innerHTML = '';
    search.value = '';
    document.getElementById('rolePermissionsTitle').textContent = 'Permisos: ' + name;
    document.getElementById('rolePermissionsInfo').textContent = 'Empresa ' + empresaID + ' · Rol ' + roleID;
    setMessage('Cargando permisos del rol…', false);
    updateStatus();
    panel.scrollIntoView({ behavior: 'smooth', block: 'start' });
    try {
      const [data, context] = await Promise.all([
        fetch(endpoint(), { credentials: 'same-origin' }).then(readResponse),
        fetch('/api/empresa/permisos_contexto?empresa_id=' + encodeURIComponent(empresaID), { credentials: 'same-origin' }).then(readResponse)
      ]);
      if (sequence !== state.sequence) return;
      if ((data.empresa_id && Number(data.empresa_id) !== empresaID) || (data.rol_id && Number(data.rol_id) !== roleID)) throw new Error('El contexto recibido no corresponde al rol solicitado.');
      state.modules = Array.isArray(data.modulos) ? data.modulos : [];
      state.pages = Array.isArray(data.paginas) ? data.paginas : [];
      state.canEdit = Array.isArray(context.modulos) && context.modulos.some(function (mod) { return mod.modulo === 'seguridad' && mod.update; });
      state.snapshot = JSON.stringify(payload());
      document.getElementById('rolePermissionsInfo').textContent = 'Empresa ' + empresaID + ' · Basado en ' + (data.rol_base_nombre || 'rol operativo') + '. Los cambios se guardan únicamente en este perfil.';
      render(data);
      setMessage(state.canEdit ? 'Permisos cargados. Selecciona las acciones y páginas que necesita este rol.' : 'Modo consulta: tu perfil no permite actualizar los permisos de seguridad.', false);
    } catch (error) {
      if (sequence === state.sequence) setMessage(error.message || 'No se pudieron cargar los permisos.', true);
    } finally {
      if (sequence === state.sequence) { state.loading = false; updateStatus(); }
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
    if (state.loading || state.saving || !state.canEdit) return;
    if (input.matches('input[data-role-module]')) {
      const action = actions.find(function (a) { return a[0] === input.dataset.roleAction; });
      const item = state.modules[Number(input.dataset.roleModule)];
      if (item && action) item[action[1]] = input.checked;
    } else if (input.matches('input[data-role-page]')) {
      const item = state.pages[Number(input.dataset.rolePage)];
      if (item) item.permitido = input.checked;
    }
    updateStatus();
  });
  search.addEventListener('input', applyFilter);
  close.addEventListener('click', function () {
    if (state.saving || (dirty() && !window.confirm('Hay permisos sin guardar. ¿Descartar los cambios?'))) return;
    ++state.sequence;
    state.snapshot = '';
    panel.hidden = true;
  });
  save.addEventListener('click', async function () {
    if (save.disabled || state.loading || state.saving || !dirty()) return;
    state.saving = true;
    updateStatus();
    setMessage('Guardando permisos del rol…', false);
    try {
      const body = payload();
      body.aprobado_por = 'administrador_empresa';
      body.codigo_aprobacion = 'permisos-rol-' + Date.now();
      body.motivo_aprobacion = 'Configuración de permisos del rol personalizado de empresa';
      await readResponse(await fetch(endpoint(), { method: 'PUT', credentials: 'same-origin', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) }));
      state.snapshot = JSON.stringify(payload());
      setMessage('Permisos del rol guardados correctamente. Se aplican dentro de la empresa y de su licencia.', false);
    } catch (error) {
      setMessage(error.message || 'No se pudieron guardar los permisos. Los cambios se conservan para reintentar.', true);
    } finally { state.saving = false; updateStatus(); }
  });
  window.addEventListener('beforeunload', function (event) {
    if (!dirty()) return;
    event.preventDefault();
    event.returnValue = '';
  });
})();
