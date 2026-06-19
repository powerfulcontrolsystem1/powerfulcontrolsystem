(function() {
  'use strict';

  const tipoMovimiento = String(document.body.dataset.finanzasTipo || 'egreso').trim().toLowerCase() === 'ingreso' ? 'ingreso' : 'egreso';
  let empresaId = resolveEmpresaId();
  let movimientos = [];
  let empresaNombre = '';
  let finanzasConfig = { formato_impresion: 'carta' };
  let advancedPrintConfig = {};

  function el(id) { return document.getElementById(id); }

  function parsePositiveInt(raw) {
    const n = Number(String(raw || '').trim());
    if (!Number.isFinite(n)) return 0;
    const v = Math.trunc(n);
    return v > 0 ? v : 0;
  }

  function resolveEmpresaId() {
    try {
      const params = new URLSearchParams(window.location.search || '');
      const own = parsePositiveInt(params.get('empresa_id') || params.get('id'));
      if (own > 0) return own;
    } catch (_) {}
    try {
      let ctx = window.parent;
      let depth = 0;
      while (ctx && ctx !== window && depth < 4) {
        try {
          if (typeof ctx.__resolveEmpresaIdContext === 'function') {
            const resolved = parsePositiveInt(ctx.__resolveEmpresaIdContext());
            if (resolved > 0) return resolved;
          }
        } catch (_) {}
        try {
          const parentParams = new URLSearchParams(ctx.location.search || '');
          const fromParent = parsePositiveInt(parentParams.get('empresa_id') || parentParams.get('id'));
          if (fromParent > 0) return fromParent;
        } catch (_) {}
        try {
          if (!ctx.parent || ctx.parent === ctx) break;
          ctx = ctx.parent;
        } catch (_) {
          break;
        }
        depth += 1;
      }
    } catch (_) {}
    try {
      const candidates = [
        sessionStorage.getItem('active_empresa_id'),
        sessionStorage.getItem('empresa_id'),
        localStorage.getItem('active_empresa_id'),
        localStorage.getItem('empresa_id')
      ];
      for (let i = 0; i < candidates.length; i += 1) {
        const parsed = parsePositiveInt(candidates[i]);
        if (parsed > 0) return parsed;
      }
    } catch (_) {}
    return 0;
  }

  function escapeHTML(value) {
    return String(value == null ? '' : value).replace(/[&<>"']/g, ch => ({
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;',
      '"': '&quot;',
      "'": '&#39;'
    }[ch]));
  }

  function num(id) {
    const n = Number(String((el(id) && el(id).value) || '').replace(',', '.'));
    return Number.isFinite(n) ? n : 0;
  }

  function normalize(value) {
    return String(value == null ? '' : value).trim();
  }

  function todayLocalDateTime() {
    const d = new Date();
    d.setMinutes(d.getMinutes() - d.getTimezoneOffset());
    return d.toISOString().slice(0, 16);
  }

  function periodoFromDateTime(value) {
    const v = normalize(value);
    if (/^\d{4}-\d{2}/.test(v)) return v.slice(0, 7);
    const d = new Date();
    return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0');
  }

  function formatMoney(value, moneda) {
    const currency = normalize(moneda) || normalize(el('moneda') && el('moneda').value) || 'COP';
    const n = Number(value || 0);
    try {
      return new Intl.NumberFormat('es-CO', { style: 'currency', currency: currency, minimumFractionDigits: 2 }).format(Number.isFinite(n) ? n : 0);
    } catch (_) {
      return currency + ' ' + (Number.isFinite(n) ? n.toFixed(2) : '0.00');
    }
  }

  function setMessage(text, type) {
    const box = el('messageBox');
    if (!box) return;
    box.textContent = text || '';
    box.className = type === 'error' ? 'error' : (type === 'success' ? 'success' : 'form-help');
  }

  async function requestJSON(url, options) {
    const resp = await fetch(url, Object.assign({ credentials: 'same-origin' }, options || {}));
    const raw = await resp.text();
    let data = null;
    try { data = raw ? JSON.parse(raw) : null; } catch (_) { data = null; }
    if (!resp.ok) {
      throw new Error((data && data.error) || raw || ('HTTP ' + resp.status));
    }
    return data;
  }

  async function loadEmpresaInfo() {
    if (!empresaId) {
      el('empresaInfo').textContent = 'Empresa no definida';
      setMessage('No se encontro empresa_id. Vuelve a seleccionar la empresa.', 'error');
      return;
    }
    try {
      const empresa = await requestJSON('/super/api/empresas?id=' + encodeURIComponent(String(empresaId)));
      empresaNombre = normalize(empresa.nombre || empresa.Nombre || 'Sin nombre');
      el('empresaInfo').textContent = 'Empresa: ' + empresaNombre;
    } catch (_) {
      empresaNombre = '#' + empresaId;
      el('empresaInfo').textContent = 'Empresa: #' + empresaId;
    }
  }

  function normalizePrintFormat(value) {
    const v = normalize(value).toLowerCase();
    return v === 'pos' || v === 'ticket' || v === 'pequena' || v === 'pequeña' ? 'pos' : 'carta';
  }

  async function loadFinanzasConfig() {
    if (!empresaId) return;
    try {
      const cfg = await requestJSON('/api/empresa/finanzas/configuracion?empresa_id=' + encodeURIComponent(String(empresaId)));
      finanzasConfig = cfg && typeof cfg === 'object' ? cfg : finanzasConfig;
    } catch (_) {
      finanzasConfig = { formato_impresion: 'carta' };
    }
    try {
      const cfg = await requestJSON('/api/empresa/configuracion_avanzada?empresa_id=' + encodeURIComponent(String(empresaId)));
      advancedPrintConfig = cfg && typeof cfg === 'object' ? cfg : {};
    } catch (_) {
      advancedPrintConfig = {};
    }
    updatePrintFormatInfo();
  }

  function updatePrintFormatInfo() {
    const info = el('printFormatInfo');
    if (!info) return;
    const format = normalizePrintFormat(finanzasConfig.formato_impresion);
    const font = format === 'pos'
      ? normalizePrintFontSize(advancedPrintConfig.impresion_reporte_fuente_pos, 11, 8, 16)
      : normalizePrintFontSize(advancedPrintConfig.impresion_reporte_fuente_carta, 13, 10, 22);
    info.textContent = 'Formato de impresion: ' + (format === 'pos' ? 'POS / ticket pequeno' : 'Carta') + '. Fuente: ' + font + ' px. Se toma de la configuracion empresarial.';
  }

  function normalizePrintFontSize(value, fallback, min, max) {
    const n = Number(value);
    const base = Number.isFinite(Number(fallback)) ? Math.trunc(Number(fallback)) : min;
    if (!Number.isFinite(n)) return Math.max(min, Math.min(max, base));
    const intValue = Math.trunc(n);
    if (intValue < min) return min;
    if (intValue > max) return max;
    return intValue;
  }

  function printStorageKey() {
    return 'finanzas_' + tipoMovimiento + '_imprimir_al_guardar_' + String(empresaId || 'global');
  }

  function getAutoPrintEnabled() {
    const input = el('imprimirAlGuardar');
    return !!(input && input.checked);
  }

  function restoreAutoPrintPreference() {
    const input = el('imprimirAlGuardar');
    if (!input) return;
    try {
      input.checked = localStorage.getItem(printStorageKey()) === '1';
    } catch (_) {
      input.checked = false;
    }
    input.addEventListener('change', () => {
      try {
        localStorage.setItem(printStorageKey(), input.checked ? '1' : '0');
      } catch (_) {}
    });
  }

  function recalcTotals() {
    const monto = Math.max(0, num('monto'));
    const impuesto = Math.max(0, num('impuesto'));
    const retenciones = Math.max(0, num('totalRetenciones'));
    const total = monto + impuesto;
    const neto = Math.max(0, total - retenciones);
    el('totalNeto').value = neto.toFixed(2);
  }

  function dateToInputValue(value) {
    const raw = normalize(value);
    if (/^\d{4}-\d{2}-\d{2}$/.test(raw)) return raw + 'T12:00';
    if (/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}/.test(raw)) return raw.slice(0, 16);
    return '';
  }

  function setValueIfPresent(id, value) {
    const node = el(id);
    if (!node) return;
    const textValue = normalize(value);
    if (textValue !== '') node.value = textValue;
  }

  async function analizarComprobanteIA() {
    if (!empresaId) {
      setMessage('No se encontro empresa_id para analizar el soporte.', 'error');
      return;
    }
    const input = el('comprobanteFile');
    const file = input && input.files && input.files[0] ? input.files[0] : null;
    if (!file) {
      setMessage('Selecciona primero una foto, imagen o PDF del comprobante.', 'error');
      return;
    }
    const btn = el('btnAnalizarComprobanteIA');
    const original = btn ? btn.textContent : '';
    if (btn) {
      btn.disabled = true;
      btn.textContent = 'Analizando...';
    }
    setMessage('Radicando soporte y analizando con GPT-5.5...', '');
    try {
      const formData = new FormData();
      formData.append('empresa_id', String(empresaId));
      formData.append('archivo', file, file.name || 'soporte');
      const radicado = await requestJSON('/api/empresa/soportes_compras_ia?empresa_id=' + encodeURIComponent(String(empresaId)) + '&action=radicar', {
        method: 'POST',
        body: formData
      });
      const soporteId = radicado && radicado.soporte && radicado.soporte.id ? radicado.soporte.id : 0;
      if (!soporteId) throw new Error('No se pudo radicar el soporte para IA.');
      const extraido = await requestJSON('/api/empresa/soportes_compras_ia?empresa_id=' + encodeURIComponent(String(empresaId)) + '&action=extraer_ia', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ soporte_id: soporteId })
      });
      const soporte = extraido && extraido.soporte ? extraido.soporte : {};
      const fecha = dateToInputValue(soporte.fecha_documento);
      if (fecha) setValueIfPresent('fechaMovimiento', fecha);
      setValueIfPresent('numeroComprobante', soporte.documento_numero);
      setValueIfPresent('referenciaExterna', soporte.codigo || soporte.documento_numero);
      setValueIfPresent('terceroNombre', soporte.proveedor_nombre);
      setValueIfPresent('terceroDocumento', soporte.proveedor_nit);
      setValueIfPresent('moneda', soporte.moneda || 'COP');
      setValueIfPresent('categoriaMovimiento', soporte.categoria_contable || (tipoMovimiento === 'ingreso' ? 'ingresos_operativos' : 'compras_gastos'));
      setValueIfPresent('concepto', soporte.documento_tipo || soporte.tipo_soporte || (tipoMovimiento === 'ingreso' ? 'Ingreso con soporte IA' : 'Egreso con soporte IA'));
      setValueIfPresent('descripcion', soporte.observaciones || ('Soporte IA ' + normalize(soporte.codigo || '')));
      setValueIfPresent('comprobanteUrl', soporte.archivo_url);
      const subtotal = Number(soporte.subtotal || 0);
      const total = Number(soporte.total || 0);
      const iva = Number(soporte.impuesto_iva || 0);
      const retenciones = Number(soporte.retencion_fuente || 0) + Number(soporte.retencion_ica || 0) + Number(soporte.retencion_iva || 0);
      if (subtotal > 0) el('monto').value = subtotal.toFixed(2);
      else if (total > 0) el('monto').value = Math.max(0, total - iva).toFixed(2);
      if (iva >= 0) el('impuesto').value = iva.toFixed(2);
      if (retenciones > 0) el('totalRetenciones').value = retenciones.toFixed(2);
      recalcTotals();
      setMessage('Datos cargados desde IA. Revisa y guarda el movimiento cuando este correcto.', 'success');
    } catch (err) {
      setMessage(err && err.message ? err.message : 'No se pudo analizar el comprobante con IA.', 'error');
    } finally {
      if (btn) {
        btn.disabled = false;
        btn.textContent = original || 'Analizar con IA';
      }
    }
  }

  function resetForm() {
    const autoPrint = getAutoPrintEnabled();
    el('movimientoForm').reset();
    const printCheck = el('imprimirAlGuardar');
    if (printCheck) printCheck.checked = autoPrint;
    el('movimientoId').value = '';
    el('codigoMovimiento').value = '';
    el('comprobanteUrl').value = '';
    el('fechaMovimiento').value = todayLocalDateTime();
    el('moneda').value = 'COP';
    el('impuesto').value = '0';
    el('totalRetenciones').value = '0';
    recalcTotals();
    setMessage('', '');
  }

  function buildPayload() {
    const fecha = normalize(el('fechaMovimiento').value) || todayLocalDateTime();
    const monto = Math.max(0, num('monto'));
    const impuesto = Math.max(0, num('impuesto'));
    const retenciones = Math.max(0, num('totalRetenciones'));
    const total = monto + impuesto;
    const totalNeto = Math.max(0, total - retenciones);
    return {
      id: parsePositiveInt(el('movimientoId').value),
      empresa_id: empresaId,
      tipo_movimiento: tipoMovimiento,
      codigo: normalize(el('codigoMovimiento').value),
      fecha_movimiento: fecha.replace('T', ' '),
      periodo_contable: periodoFromDateTime(fecha),
      categoria: normalize(el('categoriaMovimiento').value),
      subcategoria: normalize(el('subcategoria').value),
      concepto: normalize(el('concepto').value),
      descripcion: normalize(el('descripcion').value),
      metodo_pago: normalize(el('metodoPago').value),
      moneda: normalize(el('moneda').value) || 'COP',
      monto: monto,
      impuesto: impuesto,
      retencion_fuente: 0,
      retencion_ica: 0,
      retencion_iva: retenciones,
      total_retenciones: retenciones,
      total: total,
      total_neto: totalNeto,
      tercero_nombre: normalize(el('terceroNombre').value),
      tercero_documento: normalize(el('terceroDocumento').value),
      tipo_comprobante: normalize(el('tipoComprobante').value),
      numero_comprobante: normalize(el('numeroComprobante').value),
      comprobante_url: normalize(el('comprobanteUrl').value),
      referencia_externa: normalize(el('referenciaExterna').value),
      observaciones: normalize(el('observaciones').value),
      estado: 'activo'
    };
  }

  async function uploadComprobante(movimientoId) {
    const input = el('comprobanteFile');
    if (!input || !input.files || input.files.length === 0) return null;
    const fd = new FormData();
    fd.append('empresa_id', String(empresaId));
    fd.append('movimiento_id', String(movimientoId));
    fd.append('archivo', input.files[0]);
    return requestJSON('/api/empresa/finanzas/movimientos/comprobante', { method: 'POST', body: fd });
  }

  async function saveMovimiento(ev) {
    ev.preventDefault();
    if (!empresaId) {
      setMessage('No se encontro empresa activa para guardar.', 'error');
      return;
    }
    const payload = buildPayload();
    if (!payload.concepto) {
      setMessage('El concepto es obligatorio.', 'error');
      el('concepto').focus();
      return;
    }
    if (payload.monto <= 0) {
      setMessage('El monto debe ser mayor que cero.', 'error');
      el('monto').focus();
      return;
    }

    const button = el('btnGuardarMovimiento');
    const previousText = button ? button.textContent : '';
    try {
      if (button) {
        button.disabled = true;
        button.textContent = 'Guardando...';
      }
      const editing = payload.id > 0;
      const url = '/api/empresa/finanzas/movimientos' + (editing ? '' : '?empresa_id=' + encodeURIComponent(String(empresaId)));
      const saved = await requestJSON(url, {
        method: editing ? 'PUT' : 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      const id = editing ? payload.id : Number(saved && saved.id || 0);
      let uploadWarning = '';
      let uploadedComprobanteURL = '';
      if (id > 0) {
        try {
          const upload = await uploadComprobante(id);
          if (upload && upload.comprobante_url) {
            uploadedComprobanteURL = upload.comprobante_url;
            el('comprobanteUrl').value = uploadedComprobanteURL;
          }
        } catch (uploadErr) {
          uploadWarning = ' El registro quedo guardado, pero no se pudo subir el comprobante: ' + (uploadErr.message || 'error desconocido') + '.';
        }
      }
      await loadMovimientos();
      const savedItem = movimientos.find(row => Number(row.id) === id) || Object.assign({}, payload, {
        id: id,
        comprobante_url: uploadedComprobanteURL || payload.comprobante_url
      });
      if (getAutoPrintEnabled() && savedItem && id > 0) {
        printMovimiento(savedItem);
      }
      resetForm();
      setMessage((tipoMovimiento === 'egreso' ? 'Egreso' : 'Ingreso') + ' guardado correctamente.' + uploadWarning, uploadWarning ? 'error' : 'success');
    } catch (err) {
      setMessage(err.message || 'No se pudo guardar el movimiento.', 'error');
    } finally {
      if (button) {
        button.disabled = false;
        button.textContent = previousText;
      }
    }
  }

  async function loadMovimientos() {
    if (!empresaId) return;
    const params = new URLSearchParams();
    params.set('empresa_id', String(empresaId));
    params.set('tipo', tipoMovimiento);
    params.set('limit', '200');
    const q = normalize(el('filtroQ').value);
    const desde = normalize(el('filtroDesde').value);
    const hasta = normalize(el('filtroHasta').value);
    if (q) params.set('q', q);
    if (desde) params.set('desde', desde);
    if (hasta) params.set('hasta', hasta);

    movimientos = await requestJSON('/api/empresa/finanzas/movimientos?' + params.toString());
    renderTable();
  }

  function renderTable() {
    const tbody = el('movimientosTbody');
    const rows = Array.isArray(movimientos) ? movimientos : [];
    el('movCount').textContent = String(rows.length);
    const total = rows.reduce((acc, item) => acc + Number(item.total_neto || item.total || item.monto || 0), 0);
    el('movTotal').textContent = formatMoney(total, rows[0] && rows[0].moneda);
    if (!rows.length) {
      tbody.innerHTML = '<tr><td colspan="8">No hay ' + (tipoMovimiento === 'egreso' ? 'egresos' : 'ingresos') + ' registrados.</td></tr>';
      return;
    }
    tbody.innerHTML = rows.map(item => {
      const comprobante = normalize(item.comprobante_url)
        ? '<a class="btn secondary" href="' + escapeHTML(item.comprobante_url) + '" target="_blank" rel="noopener">Ver</a>'
        : '<span class="muted">Sin archivo</span>';
      return '<tr>' +
        '<td>' + escapeHTML(item.fecha_movimiento || '') + '</td>' +
        '<td>' + escapeHTML(item.codigo || '') + '</td>' +
        '<td>' + escapeHTML(item.categoria || '') + '</td>' +
        '<td>' + escapeHTML(item.concepto || '') + '</td>' +
        '<td>' + escapeHTML(item.tercero_nombre || '') + '</td>' +
        '<td>' + escapeHTML(formatMoney(item.total_neto || item.total || item.monto, item.moneda)) + '</td>' +
        '<td>' + comprobante + '</td>' +
        '<td class="actions">' +
          '<button type="button" class="btn secondary" data-action="edit" data-id="' + Number(item.id || 0) + '">Editar</button>' +
          '<button type="button" class="btn secondary" data-action="print" data-id="' + Number(item.id || 0) + '">Imprimir</button>' +
          '<button type="button" class="btn secondary" data-action="share_whatsapp" data-id="' + Number(item.id || 0) + '">WhatsApp</button>' +
          '<button type="button" class="btn secondary" data-action="share_email" data-id="' + Number(item.id || 0) + '">Correo</button>' +
          '<button type="button" class="btn danger" data-action="anular" data-id="' + Number(item.id || 0) + '">Anular</button>' +
        '</td>' +
      '</tr>';
    }).join('');
  }

  function fillForm(item) {
    el('movimientoId').value = item ? Number(item.id || 0) : '';
    el('codigoMovimiento').value = item ? normalize(item.codigo) : '';
    el('comprobanteUrl').value = item ? normalize(item.comprobante_url) : '';
    el('fechaMovimiento').value = item && item.fecha_movimiento ? String(item.fecha_movimiento).replace(' ', 'T').slice(0, 16) : todayLocalDateTime();
    el('categoriaMovimiento').value = item ? normalize(item.categoria) : el('categoriaMovimiento').value;
    el('subcategoria').value = item ? normalize(item.subcategoria) : '';
    el('concepto').value = item ? normalize(item.concepto) : '';
    el('descripcion').value = item ? normalize(item.descripcion) : '';
    el('metodoPago').value = item ? normalize(item.metodo_pago) || 'efectivo' : 'efectivo';
    el('moneda').value = item ? normalize(item.moneda) || 'COP' : 'COP';
    el('monto').value = item ? Number(item.monto || 0) : '';
    el('impuesto').value = item ? Number(item.impuesto || 0) : 0;
    el('totalRetenciones').value = item ? Number(item.total_retenciones || 0) : 0;
    el('terceroNombre').value = item ? normalize(item.tercero_nombre) : '';
    el('terceroDocumento').value = item ? normalize(item.tercero_documento) : '';
    el('tipoComprobante').value = item ? normalize(item.tipo_comprobante) || 'recibo_interno' : 'recibo_interno';
    el('numeroComprobante').value = item ? normalize(item.numero_comprobante) : '';
    el('referenciaExterna').value = item ? normalize(item.referencia_externa) : '';
    el('observaciones').value = item ? normalize(item.observaciones) : '';
    const file = el('comprobanteFile');
    if (file) file.value = '';
    recalcTotals();
    el('concepto').focus();
  }

  async function accionTabla(action, id) {
    id = parsePositiveInt(id);
    if (!id) return;
    if (action === 'edit') {
      const item = movimientos.find(row => Number(row.id) === id);
      if (item) fillForm(item);
      return;
    }
    if (action === 'print') {
      const item = movimientos.find(row => Number(row.id) === id);
      if (item) printMovimiento(item);
      return;
    }
    if (action === 'share_whatsapp' || action === 'share_email') {
      const item = movimientos.find(row => Number(row.id) === id);
      if (item) shareMovimiento(item, action === 'share_whatsapp' ? 'whatsapp' : 'email');
      return;
    }
    if (action === 'anular') {
      if (!confirm('Anular este ' + tipoMovimiento + '?')) return;
      await requestJSON('/api/empresa/finanzas/movimientos?empresa_id=' + encodeURIComponent(String(empresaId)) + '&id=' + encodeURIComponent(String(id)) + '&action=anular', { method: 'PUT' });
      setMessage((tipoMovimiento === 'egreso' ? 'Egreso' : 'Ingreso') + ' anulado.', 'success');
      await loadMovimientos();
    }
  }

  function formatDateTime(raw) {
    const value = normalize(raw);
    if (!value) return '';
    const d = new Date(value.includes('T') ? value : value.replace(' ', 'T'));
    if (Number.isNaN(d.getTime())) return value;
    return d.toLocaleString('es-CO', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function printableRows(item) {
    return [
      ['Empresa', empresaNombre || ('#' + empresaId)],
      ['Tipo', tipoMovimiento === 'egreso' ? 'Egreso' : 'Ingreso'],
      ['Codigo', item.codigo || ('MOV-' + (item.id || ''))],
      ['Fecha', formatDateTime(item.fecha_movimiento)],
      ['Categoria', item.categoria || ''],
      ['Subcategoria', item.subcategoria || ''],
      ['Concepto', item.concepto || ''],
      ['Tercero', item.tercero_nombre || ''],
      ['Documento', item.tercero_documento || ''],
      ['Metodo de pago', item.metodo_pago || ''],
      ['Comprobante', [item.tipo_comprobante, item.numero_comprobante].filter(Boolean).join(' ')],
      ['Referencia', item.referencia_externa || ''],
      ['Monto base', formatMoney(item.monto, item.moneda)],
      ['Impuesto', formatMoney(item.impuesto, item.moneda)],
      ['Retenciones', formatMoney(item.total_retenciones, item.moneda)],
      ['Total neto', formatMoney(item.total_neto || item.total || item.monto, item.moneda)],
      ['Observaciones', item.observaciones || '']
    ].filter(row => normalize(row[1]) !== '');
  }

  function buildPrintHTML(item) {
    const format = normalizePrintFormat(finanzasConfig.formato_impresion);
    const title = tipoMovimiento === 'egreso' ? 'Comprobante de egreso' : 'Comprobante de ingreso';
    const badge = format === 'pos' ? 'POS' : 'Carta';
    if (window.PCSPrint) {
      const rowsForPrint = printableRows(item).map(row => [row[0], { value: row[1], number: false }]);
      const totalValue = formatMoney(item.total_neto || item.total || item.monto, item.moneda);
      return window.PCSPrint.buildDocument({
        title,
        kind: 'reporte',
        format,
        printConfig: advancedPrintConfig,
        company: empresaNombre || ('Empresa #' + empresaId),
        subtitle: 'Generado: ' + formatDateTime(new Date().toISOString()),
        badge,
        tableHeaders: ['Campo', 'Detalle'],
        rows: rowsForPrint,
        totalLabel: 'Total neto',
        totalValue,
        note: [normalize(item.descripcion), normalize(item.observaciones), normalize(item.comprobante_url) ? ('Comprobante adjunto: ' + normalize(item.comprobante_url)) : 'Sin archivo comprobante adjunto.'].filter(Boolean).join('\n'),
        signatures: format === 'carta',
        closeAfterPrint: false
      });
    }
    const rows = printableRows(item).map(row => (
      '<tr><th>' + escapeHTML(row[0]) + '</th><td>' + escapeHTML(row[1]) + '</td></tr>'
    )).join('');
    const description = normalize(item.descripcion)
      ? '<section class="notes"><h2>Detalle</h2><p>' + escapeHTML(item.descripcion) + '</p></section>'
      : '';
    const comprobante = normalize(item.comprobante_url)
      ? '<p class="file-ref">Comprobante adjunto: ' + escapeHTML(item.comprobante_url) + '</p>'
      : '<p class="file-ref">Sin archivo comprobante adjunto.</p>';
    const pageCSS = format === 'pos'
      ? '@page{size:80mm auto;margin:4mm;} body{width:72mm;}'
      : '@page{size:letter;margin:12mm;} body{max-width:190mm;}';
    const compactCSS = format === 'pos'
      ? 'body{font-size:' + normalizePrintFontSize(advancedPrintConfig.impresion_reporte_fuente_pos, 11, 8, 16) + 'px}.receipt{border:0;padding:0}.brand h1{font-size:15px}.brand p,.meta{font-size:10px}.badge{display:none}th,td{padding:4px 0;border-bottom:1px dashed #999}.totals{font-size:13px}'
      : 'body{font-size:' + normalizePrintFontSize(advancedPrintConfig.impresion_reporte_fuente_carta, 13, 10, 22) + 'px}.receipt{border:1px solid #d7dde8;border-radius:10px;padding:18px}.brand h1{font-size:22px}.brand p,.meta{font-size:12px}th,td{padding:8px;border-bottom:1px solid #e5e7eb}.totals{font-size:18px}';
    return '<!doctype html><html lang="es"><head><meta charset="utf-8"><title>' + escapeHTML(title) + '</title>' +
      '<style>' + pageCSS + 'html,body{background:#fff;color:#111;margin:0 auto;font-family:Arial,Helvetica,sans-serif}' +
      '.receipt{box-sizing:border-box}.brand{display:flex;justify-content:space-between;align-items:flex-start;gap:12px;border-bottom:2px solid #111;padding-bottom:10px;margin-bottom:12px}' +
      '.brand h1{margin:0 0 4px;font-weight:800}.brand p{margin:2px 0;color:#333}.badge{border:1px solid #111;border-radius:6px;padding:6px 8px;font-weight:800}' +
      'table{width:100%;border-collapse:collapse}th{text-align:left;width:34%;vertical-align:top;color:#333}td{text-align:right;font-weight:700;word-break:break-word}' +
      '.totals{margin-top:12px;padding:10px;border:2px solid #111;font-weight:900;display:flex;justify-content:space-between;gap:10px}' +
      '.notes{margin-top:12px}.notes h2{font-size:13px;margin:0 0 4px}.notes p{margin:0;white-space:pre-wrap}.file-ref,.meta{color:#555;word-break:break-word}.signatures{display:grid;grid-template-columns:1fr 1fr;gap:18px;margin-top:26px}.signatures div{border-top:1px solid #111;padding-top:6px;text-align:center;color:#333}' +
      '@media print{.no-print{display:none!important}}' + compactCSS + '</style></head><body>' +
      '<article class="receipt"><header class="brand"><div><h1>' + escapeHTML(title) + '</h1><p>' + escapeHTML(empresaNombre || ('Empresa #' + empresaId)) + '</p><p class="meta">Generado: ' + escapeHTML(formatDateTime(new Date().toISOString())) + '</p></div><div class="badge">' + escapeHTML(badge) + '</div></header>' +
      '<table><tbody>' + rows + '</tbody></table><div class="totals"><span>Total neto</span><span>' + escapeHTML(formatMoney(item.total_neto || item.total || item.monto, item.moneda)) + '</span></div>' +
      description + comprobante + '<section class="signatures"><div>Recibe</div><div>Entrega / registra</div></section></article>' +
      '<script>window.addEventListener("load",function(){setTimeout(function(){window.focus();window.print();},120);});<\/script></body></html>';
  }

  function printMovimiento(item) {
    if (!item) return;
    const frame = document.createElement('iframe');
    frame.title = 'Impresion de ' + tipoMovimiento;
    frame.style.position = 'fixed';
    frame.style.right = '0';
    frame.style.bottom = '0';
    frame.style.width = '0';
    frame.style.height = '0';
    frame.style.border = '0';
    frame.style.opacity = '0';
    frame.setAttribute('aria-hidden', 'true');
    document.body.appendChild(frame);
    const doc = frame.contentWindow && frame.contentWindow.document;
    if (!doc) {
      frame.remove();
      return;
    }
    doc.open();
    doc.write(buildPrintHTML(item));
    doc.close();
    setTimeout(() => {
      try { frame.remove(); } catch (_) {}
    }, 3000);
  }

  function shareMovimiento(item, channel) {
    if (!item) return;
    const title = tipoMovimiento === 'egreso' ? 'Comprobante de egreso' : 'Comprobante de ingreso';
    let url = '';
    try {
      const shareUrl = new URL(window.location.pathname || '', window.location.origin);
      shareUrl.searchParams.set('empresa_id', String(empresaId));
      if (item.codigo) shareUrl.searchParams.set('q', normalize(item.codigo));
      url = shareUrl.toString();
    } catch (_) {
      url = window.location.href;
    }
    const message = [
      'Concepto: ' + normalize(item.concepto),
      'Tercero: ' + (normalize(item.tercero_nombre) || 'No registrado'),
      'Fecha: ' + formatDateTime(item.fecha_movimiento)
    ].join('\n');
    if (window.PCSPrint && typeof window.PCSPrint.shareDocument === 'function') {
      window.PCSPrint.shareDocument({
        channel: channel,
        title: title,
        code: normalize(item.codigo) || ('MOV-' + Number(item.id || 0)),
        total: formatMoney(item.total_neto || item.total || item.monto, item.moneda),
        company: empresaNombre || ('Empresa #' + empresaId),
        message: message,
        url: url
      });
      return;
    }
    const body = encodeURIComponent(title + '\nCodigo: ' + (normalize(item.codigo) || ('MOV-' + Number(item.id || 0))) + '\nTotal: ' + formatMoney(item.total_neto || item.total || item.monto, item.moneda) + '\n' + message + '\nEnlace: ' + url);
    const href = channel === 'whatsapp' ? ('https://wa.me/?text=' + body) : ('mailto:?subject=' + encodeURIComponent(title) + '&body=' + body);
    window.open(href, '_blank', 'noopener,noreferrer');
  }

  async function init() {
    await loadEmpresaInfo();
    await loadFinanzasConfig();
    restoreAutoPrintPreference();
    resetForm();
    ['monto', 'impuesto', 'totalRetenciones'].forEach(id => {
      const node = el(id);
      if (node) node.addEventListener('input', recalcTotals);
    });
    el('movimientoForm').addEventListener('submit', saveMovimiento);
    el('btnNuevoMovimiento').addEventListener('click', resetForm);
    el('btnCancelarMovimiento').addEventListener('click', resetForm);
    const aiBtn = el('btnAnalizarComprobanteIA');
    if (aiBtn) aiBtn.addEventListener('click', () => {
      analizarComprobanteIA().catch(err => setMessage(err.message || 'No se pudo analizar el comprobante.', 'error'));
    });
    el('btnBuscarMovimientos').addEventListener('click', () => {
      loadMovimientos().catch(err => setMessage(err.message || 'No se pudieron cargar los movimientos.', 'error'));
    });
    el('movimientosTbody').addEventListener('click', ev => {
      const btn = ev.target && ev.target.closest ? ev.target.closest('button[data-action]') : null;
      if (!btn) return;
      accionTabla(btn.dataset.action, btn.dataset.id).catch(err => setMessage(err.message || 'No se pudo ejecutar la accion.', 'error'));
    });
    await loadMovimientos();
  }

  init().catch(err => setMessage(err.message || 'No se pudo iniciar el modulo.', 'error'));
})();
