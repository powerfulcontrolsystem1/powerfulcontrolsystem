(function() {
  'use strict';

  const OFFLINE_QUEUE_VERSION = 1;

  const state = {
    empresaID: 0,
    stationID: 0,
    stationName: '',
    carritoCode: '',
    carritoID: 0,
    carrito: null,
    items: [],
    productos: [],
    productosCache: [],
    searchTimer: null,
    lastTotalRendered: 0,
    syncInProgress: false,
    offlineQueue: null
  };

  function normalize(v) {
    return String(v == null ? '' : v).trim();
  }

  function normalizedCode(v) {
    return normalize(v).toUpperCase();
  }

  function sanitize(v) {
    return String(v == null ? '' : v)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/\"/g, '&quot;')
      .replace(/'/g, '&#39;');
  }

  function toNumber(v, fallback) {
    const n = Number(v);
    return Number.isFinite(n) ? n : fallback;
  }

  function round2(v) {
    return Math.round(toNumber(v, 0) * 100) / 100;
  }

  function formatLocalDateTime(d) {
    const date = d instanceof Date ? d : new Date();
    const pad = function(n) { return n < 10 ? '0' + n : String(n); };
    return date.getFullYear() + '-' +
      pad(date.getMonth() + 1) + '-' +
      pad(date.getDate()) + ' ' +
      pad(date.getHours()) + ':' +
      pad(date.getMinutes()) + ':' +
      pad(date.getSeconds());
  }

  function money(value, currency) {
    const code = normalize(currency).toUpperCase() || 'COP';
    const amount = toNumber(value, 0);
    try {
      return amount.toLocaleString('es-CO', {
        style: 'currency',
        currency: code,
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
      });
    } catch (_) {
      return code + ' ' + amount.toFixed(2);
    }
  }

  function getCurrency() {
    return normalize(state.carrito && state.carrito.moneda).toUpperCase() || 'COP';
  }

  function setMessage(id, text, type) {
    const el = document.getElementById(id);
    if (!el) return;
    el.textContent = text || '';
    if (type === 'error') {
      el.className = 'error';
      return;
    }
    if (type === 'success') {
      el.className = 'success';
      return;
    }
    el.className = 'form-help';
  }

  async function requestJSON(url, options) {
    const response = await fetch(url, Object.assign({ credentials: 'same-origin' }, options || {}));
    const raw = await response.text();
    let data = null;

    if (raw) {
      try {
        data = JSON.parse(raw);
      } catch (_) {
        data = raw;
      }
    }

    if (!response.ok) {
      const msg = typeof data === 'string'
        ? data
        : ((data && (data.error || data.message)) || ('HTTP ' + response.status));
      const err = new Error(msg || ('HTTP ' + response.status));
      err.statusCode = response.status;
      err.responseData = data;
      throw err;
    }

    return data;
  }

  function isLikelyNetworkError(err) {
    const msg = normalize(err && err.message).toLowerCase();
    return msg.indexOf('failed to fetch') >= 0 ||
      msg.indexOf('networkerror') >= 0 ||
      msg.indexOf('network request failed') >= 0 ||
      msg.indexOf('load failed') >= 0;
  }

  function isOnline() {
    return navigator.onLine !== false;
  }

  function getCarritoCodeForStation(stationID) {
    return 'EST-' + state.empresaID + '-' + stationID;
  }

  function isCarritoOperativoActivo(carrito) {
    if (!carrito) return false;
    const estado = normalize(carrito.estado).toLowerCase() || 'activo';
    const estadoCarrito = normalize(carrito.estado_carrito).toLowerCase() || 'abierto';
    return estado === 'activo' && estadoCarrito !== 'cerrado';
  }

  function isCarritoPagado(carrito) {
    return normalize(carrito && carrito.pagado_en) !== '';
  }

  function getActiveItems() {
    return (Array.isArray(state.items) ? state.items : []).filter(function(item) {
      return normalize(item && item.estado).toLowerCase() !== 'inactivo';
    });
  }

  function getOfflineStorageKey() {
    const scope = state.stationID > 0 ? String(state.stationID) : (state.carritoCode || 'global');
    return 'ventas_simple_offline_v' + OFFLINE_QUEUE_VERSION + ':' + state.empresaID + ':' + scope;
  }

  function clonePlain(value) {
    return JSON.parse(JSON.stringify(value));
  }

  function createEmptyOfflineQueue() {
    return {
      version: OFFLINE_QUEUE_VERSION,
      empresa_id: state.empresaID,
      estacion_id: state.stationID,
      carrito_id: state.carritoID,
      carrito_codigo: state.carritoCode,
      shadow_items: [],
      pending_activate_reset: false,
      needs_sync: false,
      carrito_snapshot: null,
      updated_at: formatLocalDateTime(new Date())
    };
  }

  function ensureOfflineQueue() {
    if (!state.offlineQueue) {
      state.offlineQueue = createEmptyOfflineQueue();
    }
    state.offlineQueue.empresa_id = state.empresaID;
    state.offlineQueue.estacion_id = state.stationID;
    state.offlineQueue.carrito_id = state.carritoID;
    state.offlineQueue.carrito_codigo = state.carritoCode;
    return state.offlineQueue;
  }

  function shadowItemFromCarritoItem(item) {
    return {
      referencia_id: Number(item.referencia_id || 0),
      codigo_item: normalize(item.codigo_item),
      descripcion: normalize(item.descripcion) || 'Producto',
      unidad_medida: normalize(item.unidad_medida) || 'unidad',
      cantidad: round2(toNumber(item.cantidad, 0)),
      precio_unitario: round2(toNumber(item.precio_unitario, 0)),
      descuento_porcentaje: round2(toNumber(item.descuento_porcentaje, 0)),
      impuesto_porcentaje: round2(toNumber(item.impuesto_porcentaje, 0)),
      impuesto_codigo: normalize(item.impuesto_codigo) || 'IVA'
    };
  }

  function shadowItemFromProduct(producto, cantidad) {
    return {
      referencia_id: Number(producto.id || 0),
      codigo_item: normalize(producto.codigo_barras) || normalize(producto.sku),
      descripcion: normalize(producto.nombre) || 'Producto',
      unidad_medida: normalize(producto.unidad_medida) || 'unidad',
      cantidad: round2(cantidad),
      precio_unitario: round2(toNumber(producto.precio, 0)),
      descuento_porcentaje: 0,
      impuesto_porcentaje: round2(toNumber(producto.impuesto_porcentaje, 0)),
      impuesto_codigo: 'IVA'
    };
  }

  function lineTotalsFromShadow(shadow) {
    const cantidad = Math.max(0, toNumber(shadow.cantidad, 0));
    const precio = Math.max(0, toNumber(shadow.precio_unitario, 0));
    const descuentoPct = Math.max(0, Math.min(100, toNumber(shadow.descuento_porcentaje, 0)));
    const impuestoPct = Math.max(0, toNumber(shadow.impuesto_porcentaje, 0));
    const base = cantidad * precio;
    const descuento = base * (descuentoPct / 100);
    const baseGravable = Math.max(0, base - descuento);
    const impuesto = baseGravable * (impuestoPct / 100);
    const totalLinea = baseGravable + impuesto;
    return {
      base_gravable: round2(baseGravable),
      valor_descuento: round2(descuento),
      valor_impuesto: round2(impuesto),
      subtotal_linea: round2(baseGravable),
      total_linea: round2(totalLinea)
    };
  }

  function carritoItemFromShadow(shadow, index) {
    const totals = lineTotalsFromShadow(shadow);
    return {
      id: -100000 - index,
      empresa_id: state.empresaID,
      carrito_id: state.carritoID,
      tipo_item: 'producto',
      referencia_id: Number(shadow.referencia_id || 0),
      codigo_item: normalize(shadow.codigo_item),
      descripcion: normalize(shadow.descripcion) || 'Producto',
      unidad_medida: normalize(shadow.unidad_medida) || 'unidad',
      cantidad: round2(toNumber(shadow.cantidad, 0)),
      precio_unitario: round2(toNumber(shadow.precio_unitario, 0)),
      descuento_porcentaje: round2(toNumber(shadow.descuento_porcentaje, 0)),
      impuesto_porcentaje: round2(toNumber(shadow.impuesto_porcentaje, 0)),
      impuesto_codigo: normalize(shadow.impuesto_codigo) || 'IVA',
      base_gravable: totals.base_gravable,
      valor_descuento: totals.valor_descuento,
      valor_impuesto: totals.valor_impuesto,
      subtotal_linea: totals.subtotal_linea,
      total_linea: totals.total_linea,
      estado: 'activo',
      observaciones: 'item local pendiente de sincronizacion'
    };
  }

  function rebuildStateItemsFromShadow() {
    const queue = ensureOfflineQueue();
    const shadow = Array.isArray(queue.shadow_items) ? queue.shadow_items : [];
    state.items = shadow.filter(function(item) {
      return toNumber(item && item.cantidad, 0) > 0;
    }).map(function(item, idx) {
      return carritoItemFromShadow(item, idx);
    });
  }

  function recalculateLocalCarritoTotals() {
    const activeItems = getActiveItems();
    let subtotal = 0;
    let impuesto = 0;
    let total = 0;

    activeItems.forEach(function(item) {
      const totals = lineTotalsFromShadow(item);
      subtotal += totals.subtotal_linea;
      impuesto += totals.valor_impuesto;
      total += totals.total_linea;
    });

    if (!state.carrito) {
      state.carrito = {
        estado: 'activo',
        estado_carrito: 'abierto',
        moneda: getCurrency(),
        metodo_pago: 'efectivo',
        activado_en: formatLocalDateTime(new Date())
      };
    }

    state.carrito.subtotal = round2(subtotal);
    state.carrito.impuesto_total = round2(impuesto);
    state.carrito.total = round2(total);
    if (!isCarritoPagado(state.carrito)) {
      state.carrito.total_pagado = 0;
      state.carrito.devolucion_total = 0;
    }
  }

  function captureCarritoSnapshot() {
    const carrito = state.carrito || {};
    return {
      estado: normalize(carrito.estado) || 'activo',
      estado_carrito: normalize(carrito.estado_carrito) || 'abierto',
      moneda: normalize(carrito.moneda) || getCurrency(),
      subtotal: round2(toNumber(carrito.subtotal, 0)),
      impuesto_total: round2(toNumber(carrito.impuesto_total, 0)),
      total: round2(toNumber(carrito.total, 0)),
      total_pagado: round2(toNumber(carrito.total_pagado, 0)),
      devolucion_total: round2(toNumber(carrito.devolucion_total, 0)),
      pagado_en: normalize(carrito.pagado_en),
      activado_en: normalize(carrito.activado_en),
      metodo_pago: normalize(carrito.metodo_pago) || 'efectivo'
    };
  }

  function applyCarritoSnapshot(snapshot) {
    if (!snapshot) return;
    if (!state.carrito) {
      state.carrito = {};
    }
    state.carrito.estado = normalize(snapshot.estado) || 'activo';
    state.carrito.estado_carrito = normalize(snapshot.estado_carrito) || 'abierto';
    state.carrito.moneda = normalize(snapshot.moneda) || 'COP';
    state.carrito.subtotal = round2(toNumber(snapshot.subtotal, 0));
    state.carrito.impuesto_total = round2(toNumber(snapshot.impuesto_total, 0));
    state.carrito.total = round2(toNumber(snapshot.total, 0));
    state.carrito.total_pagado = round2(toNumber(snapshot.total_pagado, 0));
    state.carrito.devolucion_total = round2(toNumber(snapshot.devolucion_total, 0));
    state.carrito.pagado_en = normalize(snapshot.pagado_en);
    state.carrito.activado_en = normalize(snapshot.activado_en);
    state.carrito.metodo_pago = normalize(snapshot.metodo_pago) || 'efectivo';
  }

  function syncQueueFromCurrentState(needsSync) {
    const queue = ensureOfflineQueue();
    queue.carrito_id = state.carritoID;
    queue.carrito_codigo = state.carritoCode;
    queue.shadow_items = getActiveItems().map(shadowItemFromCarritoItem);
    queue.carrito_snapshot = captureCarritoSnapshot();
    if (typeof needsSync === 'boolean') {
      queue.needs_sync = needsSync;
    }
    queue.updated_at = formatLocalDateTime(new Date());
  }

  async function computeSHA256Hex(text) {
    if (!window.crypto || !window.crypto.subtle || typeof TextEncoder === 'undefined') {
      return '';
    }
    const data = new TextEncoder().encode(String(text || ''));
    const digest = await window.crypto.subtle.digest('SHA-256', data);
    const hashArray = Array.from(new Uint8Array(digest));
    return hashArray.map(function(b) { return b.toString(16).padStart(2, '0'); }).join('');
  }

  async function persistOfflineQueue() {
    const queue = ensureOfflineQueue();
    queue.version = OFFLINE_QUEUE_VERSION;
    queue.empresa_id = state.empresaID;
    queue.estacion_id = state.stationID;
    queue.carrito_id = state.carritoID;
    queue.carrito_codigo = state.carritoCode;
    queue.updated_at = formatLocalDateTime(new Date());

    const payload = { queue: clonePlain(queue) };
    const rawQueue = JSON.stringify(payload.queue);
    const checksum = await computeSHA256Hex(rawQueue + '|' + state.empresaID + '|' + state.stationID + '|' + state.carritoCode);
    payload.checksum = checksum;
    localStorage.setItem(getOfflineStorageKey(), JSON.stringify(payload));
    updateSyncStatusUI();
  }

  async function loadOfflineQueue() {
    state.offlineQueue = createEmptyOfflineQueue();
    const raw = localStorage.getItem(getOfflineStorageKey());
    if (!raw) {
      return;
    }
    try {
      const parsed = JSON.parse(raw);
      if (!parsed || typeof parsed !== 'object' || !parsed.queue) {
        return;
      }
      const queue = parsed.queue;
      if (Number(queue.version || 0) !== OFFLINE_QUEUE_VERSION) {
        return;
      }
      if (Number(queue.empresa_id || 0) !== Number(state.empresaID || 0)) {
        return;
      }
      if (Number(queue.estacion_id || 0) !== Number(state.stationID || 0)) {
        return;
      }
      const expectedChecksum = normalize(parsed.checksum);
      const rawQueue = JSON.stringify(queue);
      const computedChecksum = await computeSHA256Hex(rawQueue + '|' + state.empresaID + '|' + state.stationID + '|' + state.carritoCode);
      if (expectedChecksum && computedChecksum && expectedChecksum !== computedChecksum) {
        localStorage.removeItem(getOfflineStorageKey());
        setMessage('mainMsg', 'Se descarto cola offline por verificacion de integridad.', 'error');
        return;
      }

      state.offlineQueue = queue;
      if (Number(state.offlineQueue.carrito_id || 0) > 0) {
        state.carritoID = Number(state.offlineQueue.carrito_id || 0);
      }
      if (!state.carritoCode && normalize(state.offlineQueue.carrito_codigo)) {
        state.carritoCode = normalize(state.offlineQueue.carrito_codigo);
      }
      if (state.offlineQueue.carrito_snapshot) {
        applyCarritoSnapshot(state.offlineQueue.carrito_snapshot);
      }
      if (Array.isArray(state.offlineQueue.shadow_items) && state.offlineQueue.shadow_items.length > 0) {
        rebuildStateItemsFromShadow();
        recalculateLocalCarritoTotals();
      }
    } catch (_) {
      localStorage.removeItem(getOfflineStorageKey());
    }
  }

  function pendingSyncCount() {
    const queue = state.offlineQueue || createEmptyOfflineQueue();
    let count = 0;
    if (queue.needs_sync) {
      count += Array.isArray(queue.shadow_items) ? queue.shadow_items.length : 0;
      if (queue.pending_activate_reset) {
        count += 1;
      }
    }
    return count;
  }

  function updateSyncStatusUI(customText) {
    const tag = document.getElementById('syncStatusTag');
    const text = document.getElementById('syncStatusText');
    const btn = document.getElementById('btnSyncNow');
    if (!tag || !text || !btn) return;

    const online = isOnline();
    const pending = pendingSyncCount();

    tag.className = 'ventas-simple-sync-tag';
    if (state.syncInProgress) {
      tag.textContent = 'Sincronizando';
      tag.classList.add('is-syncing');
      text.textContent = customText || 'Aplicando cambios offline en el servidor...';
      btn.disabled = true;
      return;
    }

    if (!online) {
      tag.textContent = 'Modo offline';
      tag.classList.add('is-offline');
      text.textContent = customText || (pending > 0
        ? ('Cambios pendientes por sincronizar: ' + pending + '.')
        : 'Sin conexion. Puedes seguir operando y sincronizar despues.');
      btn.disabled = true;
      return;
    }

    tag.textContent = 'En linea';
    tag.classList.add('is-online');
    text.textContent = customText || (pending > 0
      ? ('Cambios pendientes por sincronizar: ' + pending + '.')
      : 'Sin cambios pendientes.');
    btn.disabled = pending <= 0;
  }

  async function loadEmpresaInfo() {
    if (!state.empresaID) return;
    if (!isOnline()) {
      document.getElementById('empresaInfo').textContent = 'Empresa: modo offline';
      return;
    }
    try {
      const empresa = await requestJSON('/super/api/empresas?id=' + encodeURIComponent(state.empresaID));
      const nombre = normalize(empresa && empresa.nombre) || ('Empresa #' + state.empresaID);
      document.getElementById('empresaInfo').textContent = 'Empresa: ' + nombre;
    } catch (_) {
      document.getElementById('empresaInfo').textContent = 'Empresa: no disponible';
    }
  }

  async function listCarritos() {
    return await requestJSON('/api/empresa/carritos_compra?empresa_id=' + encodeURIComponent(state.empresaID) + '&include_inactive=1');
  }

  function findStationCarrito(rows) {
    const list = Array.isArray(rows) ? rows : [];

    if (state.carritoID > 0) {
      const byID = list.find(function(item) { return Number(item.id || 0) === Number(state.carritoID); });
      if (byID) return byID;
    }

    if (state.carritoCode) {
      const targetCode = normalizedCode(state.carritoCode);
      const byCode = list.find(function(item) {
        return normalizedCode(item && item.codigo) === targetCode;
      });
      if (byCode) return byCode;
    }

    if (state.stationID > 0) {
      const expectedCode = normalizedCode(getCarritoCodeForStation(state.stationID));
      state.carritoCode = expectedCode;
      const byStationCode = list.find(function(item) {
        return normalizedCode(item && item.codigo) === expectedCode;
      });
      if (byStationCode) return byStationCode;
    }

    return null;
  }

  async function refreshCarritoFromServer() {
    const rows = await listCarritos();
    const found = findStationCarrito(rows);
    state.carrito = found || state.carrito;
    state.carritoID = found ? Number(found.id || 0) : state.carritoID;
    if (found && !state.carritoCode) {
      state.carritoCode = normalizedCode(found.codigo);
    }
    return found;
  }

  async function createStationCarrito() {
    if (!state.empresaID) {
      throw new Error('empresa_id requerido para crear carrito de estacion.');
    }
    const code = state.carritoCode || (state.stationID > 0 ? getCarritoCodeForStation(state.stationID) : '');
    if (!code) {
      throw new Error('No se pudo determinar el codigo del carrito de estacion.');
    }

    const payload = {
      empresa_id: Number(state.empresaID),
      codigo: code,
      nombre: state.stationName || (state.stationID > 0 ? ('Estacion ' + state.stationID) : 'Estacion'),
      canal_venta: 'mostrador',
      moneda: 'COP',
      referencia_externa: state.stationID > 0 ? ('ESTACION_' + state.stationID) : '',
      observaciones: 'Carrito automatico de estacion para venta simple'
    };

    const body = await requestJSON('/api/empresa/carritos_compra', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });

    const createdID = Number(body && body.id ? body.id : 0);
    if (createdID > 0) {
      state.carritoID = createdID;
    }
  }

  async function activateStationSession(resetItems) {
    if (!state.empresaID || !state.carritoID) {
      throw new Error('No se pudo activar la venta: carrito no encontrado.');
    }
    await requestJSON('/api/empresa/carritos_compra?empresa_id=' + encodeURIComponent(state.empresaID) +
      '&id=' + encodeURIComponent(state.carritoID) +
      '&action=activar_estacion' +
      '&reset_items=' + (resetItems ? '1' : '0'), {
      method: 'PUT'
    });
  }

  async function ensureStationCarritoLifecycle() {
    let carrito = await refreshCarritoFromServer();
    if (!carrito) {
      await createStationCarrito();
      carrito = await refreshCarritoFromServer();
    }
    if (!carrito) {
      throw new Error('No se pudo crear o encontrar el carrito de la estacion.');
    }

    const activo = isCarritoOperativoActivo(carrito);
    const activadoEn = normalize(carrito.activado_en);

    if (!activo) {
      await activateStationSession(true);
      carrito = await refreshCarritoFromServer();
    } else if (!activadoEn) {
      await activateStationSession(false);
      carrito = await refreshCarritoFromServer();
    }

    if (!carrito) {
      throw new Error('No se pudo preparar la sesion de venta de la estacion.');
    }
  }

  async function loadItemsFromServer() {
    if (!state.empresaID || !state.carritoID) {
      state.items = [];
      return;
    }
    const rows = await requestJSON('/api/empresa/carritos_compra/items?empresa_id=' + encodeURIComponent(state.empresaID) +
      '&carrito_id=' + encodeURIComponent(state.carritoID) + '&include_inactive=1');
    state.items = Array.isArray(rows) ? rows : [];
  }

  function renderHeader() {
    const stationText = state.stationName || (state.stationID > 0 ? ('Estacion ' + state.stationID) : 'Estacion');
    document.getElementById('pageTitle').textContent = 'Venta simple - ' + stationText;

    const carritoCode = state.carritoCode || (state.stationID > 0 ? getCarritoCodeForStation(state.stationID) : 'sin codigo');
    const estado = isCarritoOperativoActivo(state.carrito) ? 'sesion activa' : 'sesion cerrada';
    const activadoEn = normalize(state.carrito && state.carrito.activado_en);
    const activadoTexto = activadoEn ? (' | Entrada: ' + activadoEn) : '';

    document.getElementById('stationInfo').textContent =
      'Estacion: ' + stationText + ' | Carrito: ' + carritoCode + ' | Estado: ' + estado + activadoTexto;
  }

  function renderSummary() {
    const activeItems = getActiveItems();
    const subtotal = toNumber(state.carrito && state.carrito.subtotal, 0);
    const impuesto = toNumber(state.carrito && state.carrito.impuesto_total, 0);
    const total = toNumber(state.carrito && state.carrito.total, 0);
    const currency = getCurrency();

    document.getElementById('kpiItems').textContent = String(activeItems.length);
    document.getElementById('kpiSubtotal').textContent = money(subtotal, currency);
    document.getElementById('kpiImpuesto').textContent = money(impuesto, currency);
    document.getElementById('kpiTotal').textContent = money(total, currency);

    const carritoEstado = isCarritoOperativoActivo(state.carrito) ? 'activo' : 'inactivo/cerrado';
    document.getElementById('itemsSummary').textContent =
      'Estado del carrito: ' + carritoEstado + ' | Items activos: ' + activeItems.length + ' | Total actual: ' + money(total, currency);

    document.getElementById('payTotalLabel').textContent = 'Total a cobrar: ' + money(total, currency);

    const payAmount = document.getElementById('payAmount');
    const currentAmount = toNumber(payAmount.value, 0);
    if (!normalize(payAmount.value) || Math.abs(currentAmount - state.lastTotalRendered) < 0.01) {
      payAmount.value = total.toFixed(2);
    }
    state.lastTotalRendered = total;

    const canCharge = !!state.carritoID && isCarritoOperativoActivo(state.carrito) && total > 0 && isOnline();
    document.getElementById('btnCobrar').disabled = !canCharge;
    document.getElementById('btnNuevaVenta').disabled = !state.carritoID;

    const canCorrect = !!state.carritoID && isCarritoPagado(state.carrito) && isOnline();
    document.getElementById('btnCorreccionRapida').disabled = !canCorrect;
  }

  function renderProducts() {
    const tbody = document.querySelector('#productosTable tbody');
    const rows = Array.isArray(state.productos) ? state.productos : [];
    const currency = getCurrency();
    const canAdd = !!state.carritoID && isCarritoOperativoActivo(state.carrito);

    if (!rows.length) {
      tbody.innerHTML = '<tr><td colspan="5">No hay productos para mostrar con el filtro actual.</td></tr>';
      return;
    }

    tbody.innerHTML = rows.map(function(producto) {
      const codigo = normalize(producto.codigo_barras) || normalize(producto.sku) || 'sin codigo';
      const precio = money(producto.precio, currency);
      const stock = toNumber(producto.stock_total, 0).toFixed(2);
      const productoID = Number(producto.id || 0);
      return '<tr>' +
        '<td>' + sanitize(normalize(producto.nombre) || 'Producto') + '</td>' +
        '<td>' + sanitize(codigo) + '</td>' +
        '<td>' + sanitize(precio) + '</td>' +
        '<td>' + sanitize(stock) + '</td>' +
        '<td><button class="btn secondary" type="button" data-product-id="' + productoID + '" ' + (canAdd ? '' : 'disabled') + '>Agregar</button></td>' +
        '</tr>';
    }).join('');

    tbody.querySelectorAll('button[data-product-id]').forEach(function(btn) {
      btn.addEventListener('click', function() {
        const productID = Number(btn.getAttribute('data-product-id') || 0);
        addProductToCart(productID).catch(function(err) {
          setMessage('catalogMsg', err.message || 'No se pudo agregar el producto.', 'error');
        });
      });
    });
  }

  function buildItemPayloadFromProduct(producto, cantidad, existingItem) {
    const payload = {
      id: existingItem ? Number(existingItem.id || 0) : 0,
      empresa_id: Number(state.empresaID),
      carrito_id: Number(state.carritoID),
      tipo_item: 'producto',
      referencia_id: Number(producto.id || 0),
      codigo_item: normalize(producto.codigo_barras) || normalize(producto.sku),
      descripcion: normalize(producto.nombre) || 'Producto',
      unidad_medida: normalize(producto.unidad_medida) || 'unidad',
      cantidad: Number(cantidad),
      precio_unitario: toNumber(producto.precio, 0),
      descuento_porcentaje: 0,
      impuesto_porcentaje: toNumber(producto.impuesto_porcentaje, 0),
      impuesto_codigo: 'IVA',
      observaciones: 'agregado desde ventas_simple'
    };

    if (existingItem) {
      payload.cantidad = toNumber(existingItem.cantidad, 0) + Number(cantidad);
      payload.precio_unitario = toNumber(existingItem.precio_unitario, payload.precio_unitario);
      payload.descuento_porcentaje = toNumber(existingItem.descuento_porcentaje, 0);
      payload.impuesto_porcentaje = toNumber(existingItem.impuesto_porcentaje, payload.impuesto_porcentaje);
      payload.impuesto_codigo = normalize(existingItem.impuesto_codigo) || 'IVA';
      payload.descripcion = normalize(existingItem.descripcion) || payload.descripcion;
      payload.unidad_medida = normalize(existingItem.unidad_medida) || payload.unidad_medida;
    }

    return payload;
  }

  function buildItemPayloadFromShadow(shadow) {
    return {
      empresa_id: Number(state.empresaID),
      carrito_id: Number(state.carritoID),
      tipo_item: 'producto',
      referencia_id: Number(shadow.referencia_id || 0),
      codigo_item: normalize(shadow.codigo_item),
      descripcion: normalize(shadow.descripcion) || 'Producto',
      unidad_medida: normalize(shadow.unidad_medida) || 'unidad',
      cantidad: round2(toNumber(shadow.cantidad, 0)),
      precio_unitario: round2(toNumber(shadow.precio_unitario, 0)),
      descuento_porcentaje: round2(toNumber(shadow.descuento_porcentaje, 0)),
      impuesto_porcentaje: round2(toNumber(shadow.impuesto_porcentaje, 0)),
      impuesto_codigo: normalize(shadow.impuesto_codigo) || 'IVA',
      observaciones: 'sincronizacion segura desde modo offline'
    };
  }

  function findActiveItemByProductID(productID) {
    return getActiveItems().find(function(item) {
      return normalize(item.tipo_item || 'producto').toLowerCase() === 'producto' &&
        Number(item.referencia_id || 0) === Number(productID);
    }) || null;
  }

  function upsertShadowItem(producto, qtyDelta) {
    const queue = ensureOfflineQueue();
    if (!Array.isArray(queue.shadow_items)) {
      queue.shadow_items = [];
    }

    const refID = Number(producto.id || 0);
    if (refID <= 0) {
      throw new Error('Producto invalido para modo offline.');
    }

    let shadow = queue.shadow_items.find(function(item) {
      return Number(item.referencia_id || 0) === refID;
    });
    if (!shadow) {
      shadow = shadowItemFromProduct(producto, 0);
      queue.shadow_items.push(shadow);
    }

    shadow.cantidad = round2(toNumber(shadow.cantidad, 0) + qtyDelta);
    if (shadow.cantidad <= 0) {
      queue.shadow_items = queue.shadow_items.filter(function(item) {
        return Number(item.referencia_id || 0) !== refID;
      });
    }

    queue.needs_sync = true;
    queue.carrito_id = state.carritoID;
    queue.carrito_codigo = state.carritoCode;
    queue.carrito_snapshot = captureCarritoSnapshot();
  }

  async function addProductToCart(productID) {
    if (!state.carritoID) {
      throw new Error('No hay carrito de estacion activo.');
    }
    if (!isCarritoOperativoActivo(state.carrito)) {
      throw new Error('El carrito esta cerrado. Inicia una nueva venta.');
    }

    const qty = toNumber(document.getElementById('productoCantidadRapida').value, 1);
    if (qty <= 0) {
      throw new Error('La cantidad debe ser mayor a cero.');
    }

    const producto = (Array.isArray(state.productos) ? state.productos : []).find(function(item) {
      return Number(item.id || 0) === Number(productID);
    });
    if (!producto) {
      throw new Error('Producto no encontrado en el catalogo actual.');
    }

    if (isOnline()) {
      try {
        const existing = findActiveItemByProductID(productID);
        const payload = buildItemPayloadFromProduct(producto, qty, existing);
        const method = existing ? 'PUT' : 'POST';

        await requestJSON('/api/empresa/carritos_compra/items', {
          method: method,
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });

        setMessage('catalogMsg', existing ? 'Cantidad actualizada en carrito.' : 'Producto agregado al carrito.', 'success');
        await refreshAndRender();
        return;
      } catch (err) {
        if (!isLikelyNetworkError(err)) {
          throw err;
        }
      }
    }

    upsertShadowItem(producto, qty);
    rebuildStateItemsFromShadow();
    recalculateLocalCarritoTotals();
    syncQueueFromCurrentState(true);
    await persistOfflineQueue();
    renderSummary();
    renderItems();
    updateSyncStatusUI('Cambio guardado en modo offline.');
    setMessage('catalogMsg', 'Producto agregado offline. Se sincronizara cuando haya conexion.', 'success');
  }

  function renderItems() {
    const tbody = document.querySelector('#itemsTable tbody');
    const rows = getActiveItems();
    const currency = getCurrency();
    const canEdit = isCarritoOperativoActivo(state.carrito);

    if (!rows.length) {
      tbody.innerHTML = '<tr><td colspan="5">No hay items activos en esta venta.</td></tr>';
      return;
    }

    tbody.innerHTML = rows.map(function(item) {
      const itemID = Number(item.id || 0);
      const cantidad = toNumber(item.cantidad, 0);
      return '<tr>' +
        '<td>' + sanitize(normalize(item.descripcion) || 'Item') + '</td>' +
        '<td>' + sanitize(cantidad.toFixed(2)) + '</td>' +
        '<td>' + sanitize(money(item.precio_unitario, currency)) + '</td>' +
        '<td>' + sanitize(money(item.total_linea, currency)) + '</td>' +
        '<td>' +
          '<div class="ventas-simple-qty-actions">' +
            '<button class="btn secondary ventas-simple-qty-btn" type="button" data-item-id="' + itemID + '" data-action="decrease" ' + (canEdit ? '' : 'disabled') + '>-</button>' +
            '<span class="ventas-simple-qty-value">' + sanitize(cantidad.toFixed(2)) + '</span>' +
            '<button class="btn secondary ventas-simple-qty-btn" type="button" data-item-id="' + itemID + '" data-action="increase" ' + (canEdit ? '' : 'disabled') + '>+</button>' +
          '</div>' +
          '<button class="btn danger" type="button" data-item-id="' + itemID + '" data-action="remove" ' + (canEdit ? '' : 'disabled') + '>Quitar</button>' +
        '</td>' +
        '</tr>';
    }).join('');

    tbody.querySelectorAll('button[data-item-id][data-action]').forEach(function(btn) {
      btn.addEventListener('click', function() {
        const itemID = Number(btn.getAttribute('data-item-id') || 0);
        const action = normalize(btn.getAttribute('data-action')).toLowerCase();
        handleItemAction(itemID, action).catch(function(err) {
          setMessage('itemsMsg', err.message || 'No se pudo actualizar el item.', 'error');
        });
      });
    });
  }

  function buildItemPayloadFromExisting(item, cantidad) {
    return {
      id: Number(item.id || 0),
      empresa_id: Number(state.empresaID),
      carrito_id: Number(state.carritoID),
      tipo_item: normalize(item.tipo_item) || 'producto',
      referencia_id: Number(item.referencia_id || 0),
      codigo_item: normalize(item.codigo_item),
      descripcion: normalize(item.descripcion) || 'Item',
      unidad_medida: normalize(item.unidad_medida) || 'unidad',
      cantidad: Number(cantidad),
      precio_unitario: toNumber(item.precio_unitario, 0),
      descuento_porcentaje: toNumber(item.descuento_porcentaje, 0),
      impuesto_porcentaje: toNumber(item.impuesto_porcentaje, 0),
      impuesto_codigo: normalize(item.impuesto_codigo) || 'IVA',
      observaciones: normalize(item.observaciones)
    };
  }

  function adjustShadowByItem(item, action) {
    const refID = Number(item.referencia_id || 0);
    if (refID <= 0) {
      throw new Error('El item no es compatible con sincronizacion offline.');
    }

    const queue = ensureOfflineQueue();
    if (!Array.isArray(queue.shadow_items)) {
      queue.shadow_items = [];
    }

    const shadow = queue.shadow_items.find(function(entry) {
      return Number(entry.referencia_id || 0) === refID;
    }) || shadowItemFromCarritoItem(item);

    if (!queue.shadow_items.some(function(entry) { return Number(entry.referencia_id || 0) === refID; })) {
      queue.shadow_items.push(shadow);
    }

    if (action === 'remove') {
      shadow.cantidad = 0;
    } else if (action === 'increase') {
      shadow.cantidad = round2(toNumber(shadow.cantidad, 0) + 1);
    } else if (action === 'decrease') {
      shadow.cantidad = round2(toNumber(shadow.cantidad, 0) - 1);
    }

    queue.shadow_items = queue.shadow_items.filter(function(entry) {
      return toNumber(entry.cantidad, 0) > 0;
    });
    queue.needs_sync = true;
    queue.carrito_snapshot = captureCarritoSnapshot();
  }

  async function handleItemAction(itemID, action) {
    const item = getActiveItems().find(function(entry) {
      return Number(entry.id || 0) === Number(itemID);
    });
    if (!item) {
      throw new Error('No se encontro el item seleccionado.');
    }

    if (!isCarritoOperativoActivo(state.carrito)) {
      throw new Error('El carrito esta cerrado. Inicia una nueva venta para editar items.');
    }

    if (isOnline()) {
      try {
        if (action === 'remove') {
          await requestJSON('/api/empresa/carritos_compra/items?empresa_id=' + encodeURIComponent(state.empresaID) +
            '&carrito_id=' + encodeURIComponent(state.carritoID) +
            '&id=' + encodeURIComponent(itemID), {
            method: 'DELETE'
          });
          setMessage('itemsMsg', 'Item eliminado del carrito.', 'success');
          await refreshAndRender();
          return;
        }

        let nextQty = toNumber(item.cantidad, 0);
        if (action === 'increase') {
          nextQty += 1;
        } else if (action === 'decrease') {
          nextQty -= 1;
        }

        if (nextQty <= 0) {
          await requestJSON('/api/empresa/carritos_compra/items?empresa_id=' + encodeURIComponent(state.empresaID) +
            '&carrito_id=' + encodeURIComponent(state.carritoID) +
            '&id=' + encodeURIComponent(itemID), {
            method: 'DELETE'
          });
          setMessage('itemsMsg', 'Item eliminado del carrito.', 'success');
          await refreshAndRender();
          return;
        }

        const payload = buildItemPayloadFromExisting(item, nextQty);
        await requestJSON('/api/empresa/carritos_compra/items', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });
        setMessage('itemsMsg', 'Cantidad actualizada.', 'success');
        await refreshAndRender();
        return;
      } catch (err) {
        if (!isLikelyNetworkError(err)) {
          throw err;
        }
      }
    }

    adjustShadowByItem(item, action);
    rebuildStateItemsFromShadow();
    recalculateLocalCarritoTotals();
    syncQueueFromCurrentState(true);
    await persistOfflineQueue();
    renderSummary();
    renderItems();
    updateSyncStatusUI('Cambio guardado en modo offline.');
    setMessage('itemsMsg', 'Cambio guardado offline. Se sincronizara cuando haya conexion.', 'success');
  }

  async function searchProducts() {
    if (!state.empresaID) return;
    const q = normalize(document.getElementById('productoSearch').value).toLowerCase();

    if (!isOnline()) {
      const base = Array.isArray(state.productosCache) ? state.productosCache : [];
      state.productos = base.filter(function(item) {
        if (!q) return true;
        const haystack = [
          normalize(item && item.nombre).toLowerCase(),
          normalize(item && item.sku).toLowerCase(),
          normalize(item && item.codigo_barras).toLowerCase()
        ].join(' | ');
        return haystack.indexOf(q) >= 0;
      });
      setMessage('catalogMsg', 'Catalogo filtrado en cache local (modo offline).', null);
      renderProducts();
      return;
    }

    const params = new URLSearchParams();
    params.set('empresa_id', String(state.empresaID));
    params.set('estado', 'activo');
    params.set('limit', '80');
    if (q) {
      params.set('q', q);
    }

    const rows = await requestJSON('/api/empresa/productos?' + params.toString());
    state.productos = Array.isArray(rows) ? rows : [];
    state.productosCache = clonePlain(state.productos);
    setMessage('catalogMsg', state.productos.length ? ('Productos encontrados: ' + state.productos.length) : 'Sin resultados para el filtro actual.', null);
    renderProducts();
  }

  function updateReferenceFieldVisibility() {
    const method = normalize(document.getElementById('payMethod').value).toLowerCase();
    const requiresReference = method === 'tarjeta_credito' || method === 'tarjeta_debito' || method === 'transferencia_bancaria';
    const col = document.getElementById('payReferenceCol');
    const input = document.getElementById('payReference');
    col.classList.toggle('ventas-simple-hidden', !requiresReference);
    if (!requiresReference) {
      input.value = '';
    }
  }

  async function cobrarCarrito() {
    if (!state.empresaID || !state.carritoID) {
      setMessage('payMsg', 'No hay carrito disponible para cobrar.', 'error');
      return;
    }
    if (!isCarritoOperativoActivo(state.carrito)) {
      setMessage('payMsg', 'El carrito ya esta cerrado. Inicia una nueva venta.', 'error');
      return;
    }
    if (!isOnline()) {
      setMessage('payMsg', 'Para cerrar cobro se requiere conexion. Sincroniza y vuelve a intentar.', 'error');
      return;
    }

    const total = toNumber(state.carrito && state.carrito.total, 0);
    if (total <= 0) {
      setMessage('payMsg', 'El total debe ser mayor a cero para cobrar.', 'error');
      return;
    }

    const metodo = normalize(document.getElementById('payMethod').value).toLowerCase() || 'efectivo';
    const referencia = normalize(document.getElementById('payReference').value);
    const totalPagado = toNumber(document.getElementById('payAmount').value, total);

    if ((metodo === 'tarjeta_credito' || metodo === 'tarjeta_debito' || metodo === 'transferencia_bancaria') && referencia.length < 4) {
      setMessage('payMsg', 'La referencia es obligatoria (minimo 4 caracteres) para este metodo.', 'error');
      return;
    }

    const payload = {
      metodo_pago: metodo,
      total_pagado: totalPagado
    };
    if (referencia) {
      payload.referencia_pago = referencia;
    }

    try {
      setMessage('payMsg', 'Procesando cobro...', null);
      await requestJSON('/api/empresa/carritos_compra?empresa_id=' + encodeURIComponent(state.empresaID) +
        '&id=' + encodeURIComponent(state.carritoID) + '&action=pagar_estacion', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });

      setMessage('payMsg', 'Cobro aplicado y carrito cerrado.', 'success');
      setMessage('mainMsg', 'Venta cerrada correctamente. Usa "Iniciar nueva venta" para continuar en esta estacion.', 'success');
      await refreshAndRender();
    } catch (err) {
      setMessage('payMsg', err.message || 'No se pudo completar el cobro.', 'error');
    }
  }

  async function applyQuickCorrection() {
    if (!state.empresaID || !state.carritoID) {
      setMessage('correctionMsg', 'No hay carrito disponible para corregir.', 'error');
      return;
    }
    if (!isCarritoPagado(state.carrito)) {
      setMessage('correctionMsg', 'Solo puedes corregir una venta ya pagada.', 'error');
      return;
    }
    if (!isOnline()) {
      setMessage('correctionMsg', 'Para correccion post-cobro se requiere conexion.', 'error');
      return;
    }

    const monto = toNumber(document.getElementById('correctionAmount').value, 0);
    if (monto <= 0) {
      setMessage('correctionMsg', 'El monto a corregir debe ser mayor a cero.', 'error');
      return;
    }

    let motivo = normalize(document.getElementById('correctionReason').value);
    if (!motivo) {
      motivo = 'correccion rapida post-cobro';
    }

    try {
      setMessage('correctionMsg', 'Aplicando correccion...', null);
      await requestJSON('/api/empresa/carritos_compra?empresa_id=' + encodeURIComponent(state.empresaID) +
        '&id=' + encodeURIComponent(state.carritoID) + '&action=anular_cierre_parcial', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          monto_anulado: monto,
          motivo: motivo
        })
      });

      setMessage('correctionMsg', 'Correccion aplicada con trazabilidad.', 'success');
      setMessage('mainMsg', 'Se aplico correccion post-cobro en la venta de estacion.', 'success');
      await refreshAndRender();
    } catch (err) {
      setMessage('correctionMsg', err.message || 'No se pudo aplicar la correccion.', 'error');
    }
  }

  async function startNewSale() {
    try {
      if (!state.carritoID) {
        throw new Error('No se encontro carrito de estacion.');
      }

      if (isOnline()) {
        await activateStationSession(true);
        await refreshAndRender();
        setMessage('mainMsg', 'Nueva venta iniciada para la estacion.', 'success');
        setMessage('payMsg', '', null);
        setMessage('correctionMsg', '', null);
        return;
      }

      const queue = ensureOfflineQueue();
      queue.pending_activate_reset = true;
      queue.needs_sync = true;
      queue.shadow_items = [];

      if (!state.carrito) {
        state.carrito = {};
      }
      state.carrito.estado = 'activo';
      state.carrito.estado_carrito = 'abierto';
      state.carrito.pagado_en = '';
      state.carrito.activado_en = formatLocalDateTime(new Date());
      state.carrito.total_pagado = 0;
      state.carrito.devolucion_total = 0;
      state.carrito.metodo_pago = 'efectivo';
      state.items = [];
      recalculateLocalCarritoTotals();
      syncQueueFromCurrentState(true);
      await persistOfflineQueue();

      renderSummary();
      renderItems();
      updateSyncStatusUI('Nueva venta preparada offline. Sincroniza para aplicar en servidor.');
      setMessage('mainMsg', 'Nueva venta preparada en modo offline. Se aplicara en servidor al sincronizar.', 'success');
      setMessage('payMsg', '', null);
      setMessage('correctionMsg', '', null);
    } catch (err) {
      setMessage('mainMsg', err.message || 'No se pudo iniciar una nueva venta.', 'error');
    }
  }

  async function reconcileServerItemsWithShadow() {
    const queue = ensureOfflineQueue();
    const shadowItems = (Array.isArray(queue.shadow_items) ? queue.shadow_items : []).filter(function(item) {
      return Number(item && item.referencia_id || 0) > 0 && toNumber(item && item.cantidad, 0) > 0;
    });

    const targetByProduct = new Map();
    shadowItems.forEach(function(item) {
      targetByProduct.set(Number(item.referencia_id || 0), item);
    });

    const serverActiveItems = getActiveItems().filter(function(item) {
      return normalize(item.tipo_item || 'producto').toLowerCase() === 'producto' && Number(item.referencia_id || 0) > 0;
    });

    for (let i = 0; i < serverActiveItems.length; i++) {
      const item = serverActiveItems[i];
      const refID = Number(item.referencia_id || 0);
      const target = targetByProduct.get(refID);

      if (!target) {
        await requestJSON('/api/empresa/carritos_compra/items?empresa_id=' + encodeURIComponent(state.empresaID) +
          '&carrito_id=' + encodeURIComponent(state.carritoID) +
          '&id=' + encodeURIComponent(item.id), {
          method: 'DELETE'
        });
        continue;
      }

      targetByProduct.delete(refID);

      const currentQty = round2(toNumber(item.cantidad, 0));
      const targetQty = round2(toNumber(target.cantidad, 0));
      const currentPrice = round2(toNumber(item.precio_unitario, 0));
      const targetPrice = round2(toNumber(target.precio_unitario, 0));

      if (Math.abs(currentQty - targetQty) <= 0.009 && Math.abs(currentPrice - targetPrice) <= 0.009) {
        continue;
      }

      const payload = buildItemPayloadFromExisting(item, targetQty);
      payload.precio_unitario = targetPrice;
      payload.impuesto_porcentaje = round2(toNumber(target.impuesto_porcentaje, payload.impuesto_porcentaje));
      payload.codigo_item = normalize(target.codigo_item) || payload.codigo_item;
      payload.descripcion = normalize(target.descripcion) || payload.descripcion;
      await requestJSON('/api/empresa/carritos_compra/items', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
    }

    const remaining = Array.from(targetByProduct.values());
    for (let i = 0; i < remaining.length; i++) {
      const payload = buildItemPayloadFromShadow(remaining[i]);
      await requestJSON('/api/empresa/carritos_compra/items', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
    }
  }

  async function syncOfflineQueue(reason) {
    if (state.syncInProgress) {
      return;
    }
    if (!isOnline()) {
      updateSyncStatusUI('Sin conexion. La sincronizacion se reintentara cuando vuelva la red.');
      return;
    }

    const queue = ensureOfflineQueue();
    if (!queue.needs_sync && !queue.pending_activate_reset) {
      updateSyncStatusUI();
      return;
    }

    state.syncInProgress = true;
    updateSyncStatusUI('Sincronizando cambios offline...');

    try {
      await ensureStationCarritoLifecycle();
      await refreshCarritoFromServer();
      await loadItemsFromServer();

      if (queue.pending_activate_reset) {
        await activateStationSession(true);
        queue.pending_activate_reset = false;
        await refreshCarritoFromServer();
        await loadItemsFromServer();
      }

      await reconcileServerItemsWithShadow();
      await refreshCarritoFromServer();
      await loadItemsFromServer();

      queue.needs_sync = false;
      queue.shadow_items = getActiveItems().map(shadowItemFromCarritoItem);
      queue.carrito_snapshot = captureCarritoSnapshot();
      await persistOfflineQueue();

      renderHeader();
      renderSummary();
      renderItems();
      renderProducts();
      await loadStationMetrics();
      setMessage('mainMsg', 'Sincronizacion completada. Cambios offline aplicados en servidor.', 'success');
      updateSyncStatusUI('Sincronizacion completada.');
    } catch (err) {
      setMessage('mainMsg', (err && err.message) || 'No se pudo sincronizar cambios offline.', 'error');
      updateSyncStatusUI('Sincronizacion pendiente por error. Reintenta al recuperar conexion.');
    } finally {
      state.syncInProgress = false;
      updateSyncStatusUI();
    }

    if (reason === 'manual') {
      setMessage('catalogMsg', '', null);
      setMessage('itemsMsg', '', null);
    }
  }

  async function loadStationMetrics() {
    const defaultCurrency = getCurrency();
    const setDefaults = function(text) {
      document.getElementById('metricVentas7d').textContent = '0';
      document.getElementById('metricCorrecciones7d').textContent = '0';
      document.getElementById('metricTiempoPromedio').textContent = '0 min';
      document.getElementById('metricMonto7d').textContent = money(0, defaultCurrency);
      document.getElementById('metricasInfo').textContent = text || 'Metricas por estacion (ultimos 7 dias).';
    };

    if (!isOnline() || !state.empresaID || !state.stationID) {
      setDefaults('Metricas no disponibles en modo offline.');
      return;
    }

    try {
      const params = new URLSearchParams();
      params.set('empresa_id', String(state.empresaID));
      params.set('action', 'metricas_estacion');
      params.set('estacion_id', String(state.stationID));
      params.set('days', '7');
      params.set('limit', '5');

      const payload = await requestJSON('/api/empresa/carritos_compra?' + params.toString());
      const rows = Array.isArray(payload && payload.rows) ? payload.rows : [];
      const row = rows.length ? rows[0] : null;

      if (!row) {
        setDefaults('Sin operaciones recientes para esta estacion en los ultimos 7 dias.');
        return;
      }

      document.getElementById('metricVentas7d').textContent = String(toNumber(row.ventas_pagadas, 0));
      document.getElementById('metricCorrecciones7d').textContent = String(toNumber(row.correcciones, 0));

      const mins = Math.max(0, toNumber(row.tiempo_promedio_segundos, 0)) / 60;
      document.getElementById('metricTiempoPromedio').textContent = mins.toFixed(1) + ' min';
      document.getElementById('metricMonto7d').textContent = money(toNumber(row.monto_vendido, 0), defaultCurrency);

      const ultima = normalize(row.ultima_operacion);
      document.getElementById('metricasInfo').textContent = ultima
        ? ('Metricas actualizadas. Ultima operacion: ' + ultima)
        : 'Metricas por estacion (ultimos 7 dias).';
    } catch (err) {
      setDefaults('No se pudieron cargar metricas de estacion.');
      setMessage('mainMsg', err.message || 'No se pudieron consultar metricas de estacion.', 'error');
    }
  }

  async function refreshAndRender() {
    if (!isOnline()) {
      if (state.offlineQueue && Array.isArray(state.offlineQueue.shadow_items) && state.offlineQueue.shadow_items.length) {
        rebuildStateItemsFromShadow();
        recalculateLocalCarritoTotals();
      } else if (state.offlineQueue && state.offlineQueue.carrito_snapshot) {
        applyCarritoSnapshot(state.offlineQueue.carrito_snapshot);
      }
      renderHeader();
      renderSummary();
      renderItems();
      renderProducts();
      await loadStationMetrics();
      updateSyncStatusUI();
      return;
    }

    await refreshCarritoFromServer();
    await loadItemsFromServer();

    const queue = ensureOfflineQueue();
    if (!queue.needs_sync && !queue.pending_activate_reset) {
      syncQueueFromCurrentState(false);
      await persistOfflineQueue();
    }

    renderHeader();
    renderSummary();
    renderItems();
    renderProducts();
    await loadStationMetrics();
    updateSyncStatusUI();
  }

  function getStationsPageURL() {
    const params = new URLSearchParams();
    if (state.empresaID) {
      params.set('empresa_id', String(state.empresaID));
    }
    return '/administrar_empresa/estaciones.html' + (params.toString() ? ('?' + params.toString()) : '');
  }

  function wireEvents() {
    document.getElementById('btnBackToStations').addEventListener('click', function() {
      window.location.href = getStationsPageURL();
    });

    document.getElementById('btnRecargar').addEventListener('click', function() {
      refreshAndRender().catch(function(err) {
        setMessage('mainMsg', err.message || 'No se pudo actualizar la vista.', 'error');
      });
    });

    document.getElementById('btnBuscarProductos').addEventListener('click', function() {
      searchProducts().catch(function(err) {
        setMessage('catalogMsg', err.message || 'No se pudo consultar productos.', 'error');
      });
    });

    document.getElementById('productoSearch').addEventListener('input', function() {
      if (state.searchTimer) {
        clearTimeout(state.searchTimer);
      }
      state.searchTimer = setTimeout(function() {
        searchProducts().catch(function(err) {
          setMessage('catalogMsg', err.message || 'No se pudo consultar productos.', 'error');
        });
      }, 250);
    });

    document.getElementById('productoSearch').addEventListener('keydown', function(ev) {
      if (ev.key !== 'Enter') return;
      ev.preventDefault();
      searchProducts().catch(function(err) {
        setMessage('catalogMsg', err.message || 'No se pudo consultar productos.', 'error');
      });
    });

    document.getElementById('payMethod').addEventListener('change', function() {
      updateReferenceFieldVisibility();
    });

    document.getElementById('btnCobrar').addEventListener('click', function() {
      cobrarCarrito();
    });

    document.getElementById('btnNuevaVenta').addEventListener('click', function() {
      startNewSale();
    });

    document.getElementById('btnCorreccionRapida').addEventListener('click', function() {
      applyQuickCorrection();
    });

    document.getElementById('btnSyncNow').addEventListener('click', function() {
      syncOfflineQueue('manual');
    });

    window.addEventListener('offline', function() {
      updateSyncStatusUI('Modo offline activo. Los cambios se guardaran localmente.');
      setMessage('mainMsg', 'Conexion perdida. Operando en modo offline con sincronizacion diferida.', 'error');
    });

    window.addEventListener('online', function() {
      updateSyncStatusUI('Conexion restablecida.');
      syncOfflineQueue('online');
      refreshAndRender().catch(function(err) {
        setMessage('mainMsg', err.message || 'No se pudo refrescar tras recuperar conexion.', 'error');
      });
    });
  }

  async function bootstrap() {
    const params = new URLSearchParams(window.location.search);
    state.empresaID = Number(params.get('empresa_id') || 0);
    state.stationID = Number(params.get('estacion_id') || 0);
    state.stationName = normalize(params.get('estacion_nombre') || '');
    state.carritoCode = normalizedCode(params.get('carrito_codigo') || '');
    state.carritoID = Number(params.get('carrito_id') || 0);

    if (!state.empresaID) {
      setMessage('mainMsg', 'Falta empresa_id en la URL.', 'error');
      return;
    }

    if (!state.stationName && state.stationID > 0) {
      state.stationName = 'Estacion ' + state.stationID;
    }

    if (!state.carritoCode && state.stationID > 0) {
      state.carritoCode = normalizedCode(getCarritoCodeForStation(state.stationID));
    }

    if (!state.stationID && !state.carritoCode) {
      setMessage('mainMsg', 'Debes abrir esta vista desde la pantalla de Estaciones para indicar estacion_id.', 'error');
      return;
    }

    wireEvents();
    updateReferenceFieldVisibility();
    renderHeader();

    await loadOfflineQueue();
    updateSyncStatusUI();
    await loadEmpresaInfo();

    if (!isOnline()) {
      if (state.offlineQueue && state.offlineQueue.carrito_snapshot) {
        applyCarritoSnapshot(state.offlineQueue.carrito_snapshot);
      }
      if (state.offlineQueue && Array.isArray(state.offlineQueue.shadow_items) && state.offlineQueue.shadow_items.length) {
        rebuildStateItemsFromShadow();
        recalculateLocalCarritoTotals();
      }
      renderHeader();
      renderSummary();
      renderItems();
      renderProducts();
      await loadStationMetrics();
      setMessage('mainMsg', 'Modo offline activo. Puedes operar y sincronizar luego.', 'error');
      return;
    }

    try {
      await ensureStationCarritoLifecycle();
      await loadItemsFromServer();
      await searchProducts();

      const queue = ensureOfflineQueue();
      if (queue.needs_sync || queue.pending_activate_reset) {
        await syncOfflineQueue('bootstrap');
      } else {
        syncQueueFromCurrentState(false);
        await persistOfflineQueue();
      }

      renderHeader();
      renderSummary();
      renderItems();
      await loadStationMetrics();
    } catch (err) {
      setMessage('mainMsg', err.message || 'No se pudo iniciar la venta simple por estacion.', 'error');
    }
  }

  bootstrap();
})();
