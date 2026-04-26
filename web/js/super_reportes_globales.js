(function () {
  'use strict';

  var state = {
    empresas: [],
    datasets: [],
    selectionMode: 'multiple',
    selectedEmpresaIDs: [],
    tablero: null,
    dataset: null
  };

  function toNumber(value) {
    var num = Number(value);
    return Number.isFinite(num) ? num : 0;
  }

  function byId(id) {
    return document.getElementById(id);
  }

  function normalize(value) {
    return String(value == null ? '' : value).trim();
  }

  function escapeHtml(value) {
    return String(value == null ? '' : value)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;');
  }

  function fmtMoney(value) {
    return new Intl.NumberFormat('es-CO', { style: 'currency', currency: 'COP', maximumFractionDigits: 0 }).format(toNumber(value));
  }

  function fmtCompactNumber(value) {
    return new Intl.NumberFormat('es-CO', { notation: 'compact', maximumFractionDigits: 1 }).format(toNumber(value));
  }

  function pct(value, total) {
    if (!total) return 0;
    return Math.round((toNumber(value) * 1000) / toNumber(total)) / 10;
  }

  function setMsg(text, isError) {
    var el = byId('rgMsg');
    if (!el) return;
    el.textContent = text || '';
    el.style.color = isError ? '#ef5350' : '';
  }

  function getSelectedEmpresaIDs() {
    var ids = state.selectedEmpresaIDs.slice().sort(function (a, b) { return a - b; });
    if (state.selectionMode === 'single') {
      return ids.length ? [ids[0]] : [];
    }
    return ids;
  }

  function syncSelectedFromDOM() {
    if (state.selectionMode === 'single') {
      syncSelectedFromSingleSelect();
      return;
    }
    var inputs = document.querySelectorAll('.reportes-globales-empresa-item input[type="checkbox"]');
    var next = [];
    Array.prototype.forEach.call(inputs, function (input) {
      if (input.checked) next.push(Number(input.value || 0));
    });
    state.selectedEmpresaIDs = next.filter(function (id) { return Number.isFinite(id) && id > 0; });
  }

  function syncSelectedFromSingleSelect() {
    var select = byId('rgEmpresaUnica');
    if (!select) {
      state.selectedEmpresaIDs = [];
      return;
    }
    var value = Number(select.value || 0);
    state.selectedEmpresaIDs = value > 0 ? [value] : [];
  }

  function updateSelectionSummary() {
    var resume = byId('rgSeleccionResumen');
    var help = byId('rgSeleccionAyuda');
    var ids = getSelectedEmpresaIDs();
    if (resume) {
      if (state.selectionMode === 'single') {
        var empresa = state.empresas.find(function (item) { return Number(item.id) === Number(ids[0] || 0); });
        resume.textContent = empresa ? (empresa.nombre || ('Empresa ' + empresa.id)) : 'Sin empresa';
      } else {
        resume.textContent = ids.length + ' empresa' + (ids.length === 1 ? '' : 's');
      }
    }
    if (help) {
      help.textContent = state.selectionMode === 'single'
        ? 'Elige una empresa puntual para ver su consolidado y sus datasets sin mezclarla con las demás.'
        : 'Marca una o varias empresas y luego aplica filtros. Puedes usar Todas o Activas para armar el bloque rápido.';
    }
  }

  function renderSingleEmpresaOptions() {
    var wrap = byId('rgEmpresaUnicaWrap');
    var select = byId('rgEmpresaUnica');
    var empresasWrap = byId('rgEmpresasLista');
    var allBtn = byId('rgSeleccionarTodas');
    var activeBtn = byId('rgSoloActivas');
    var clearBtn = byId('rgLimpiarEmpresas');
    if (!select) return;

    if (!state.empresas.length) {
      select.innerHTML = '<option value="">Sin empresas</option>';
    } else {
      select.innerHTML = state.empresas.map(function (empresa) {
        return '<option value="' + escapeHtml(empresa.id) + '">' + escapeHtml(empresa.nombre || ('Empresa ' + empresa.id)) + '</option>';
      }).join('');
    }

    if (state.selectionMode === 'single') {
      var currentId = Number(state.selectedEmpresaIDs[0] || 0);
      if (!currentId && state.empresas.length) {
        currentId = Number(state.empresas[0].id);
        state.selectedEmpresaIDs = [currentId];
      }
      if (currentId) {
        select.value = String(currentId);
      }
      if (wrap) wrap.hidden = false;
      if (empresasWrap) empresasWrap.hidden = true;
      if (allBtn) allBtn.disabled = true;
      if (activeBtn) activeBtn.disabled = true;
      if (clearBtn) clearBtn.disabled = true;
    } else {
      if (wrap) wrap.hidden = true;
      if (empresasWrap) empresasWrap.hidden = false;
      if (allBtn) allBtn.disabled = false;
      if (activeBtn) activeBtn.disabled = false;
      if (clearBtn) clearBtn.disabled = false;
    }

    updateSelectionSummary();
  }

  async function refreshAfterSelectionChange() {
    if (state.selectionMode !== 'single') {
      updateSelectionSummary();
      return;
    }
    updateSelectionSummary();
    try {
      await refreshAll();
    } catch (err) {
      setMsg(err.message || 'No se pudo refrescar la empresa seleccionada.', true);
    }
  }

  function buildBaseParams(action) {
    var params = new URLSearchParams();
    params.set('action', action);
    params.set('modo', normalize(byId('rgModo') && byId('rgModo').value) || 'consolidado');
    params.set('max_rows', '300');

    var dataset = normalize(byId('rgDataset') && byId('rgDataset').value);
    if (dataset) params.set('dataset', dataset);

    var desde = normalize(byId('rgFechaDesde') && byId('rgFechaDesde').value);
    var hasta = normalize(byId('rgFechaHasta') && byId('rgFechaHasta').value);
    if (desde) params.set('desde', desde);
    if (hasta) params.set('hasta', hasta);

    var selected = getSelectedEmpresaIDs();
    if (selected.length) {
      if (state.selectionMode === 'single') params.set('empresa_id', String(selected[0]));
      else params.set('empresa_ids', selected.join(','));
    }
    return params;
  }

  async function requestJSON(url) {
    var res = await fetch(url, { credentials: 'same-origin' });
    if (!res.ok) {
      var body = await res.text();
      throw new Error(body || ('HTTP ' + res.status));
    }
    return res.json();
  }

  function renderEmpresas() {
    var wrap = byId('rgEmpresasLista');
    if (!wrap) return;
    if (!state.empresas.length) {
      wrap.innerHTML = '<div class="reportes-globales-empty">No hay empresas creadas por este administrador.</div>';
      return;
    }
    wrap.innerHTML = state.empresas.map(function (empresa) {
      var checked = state.selectedEmpresaIDs.indexOf(Number(empresa.id)) >= 0 ? ' checked' : '';
      var estado = normalize(empresa.estado || 'activo') || 'activo';
      return '' +
        '<label class="reportes-globales-empresa-item">' +
          '<input type="checkbox" value="' + escapeHtml(empresa.id) + '"' + checked + '>' +
          '<span class="reportes-globales-empresa-copy">' +
            '<strong>' + escapeHtml(empresa.nombre || ('Empresa ' + empresa.id)) + '</strong>' +
            '<span>NIT: ' + escapeHtml(empresa.nit || '-') + '</span>' +
            '<span>Estado: ' + escapeHtml(estado) + '</span>' +
          '</span>' +
        '</label>';
    }).join('');

    Array.prototype.forEach.call(wrap.querySelectorAll('input[type="checkbox"]'), function (input) {
      input.addEventListener('change', function () {
        if (state.selectionMode === 'single') {
          state.selectedEmpresaIDs = [Number(input.value || 0)];
          renderEmpresas();
        }
        syncSelectedFromDOM();
        updateSelectionSummary();
      });
    });

    renderSingleEmpresaOptions();
  }

  function renderDatasetOptions() {
    var select = byId('rgDataset');
    if (!select) return;
    if (!state.datasets.length) {
      select.innerHTML = '<option value="">Sin datasets</option>';
      return;
    }
    select.innerHTML = state.datasets.map(function (item) {
      var label = '[' + normalize(item.level || 'operativo') + '] ' + normalize(item.title || item.key || 'dataset');
      return '<option value="' + escapeHtml(item.key) + '">' + escapeHtml(label) + '</option>';
    }).join('');
    if (!normalize(select.value)) {
      select.value = normalize(state.datasets[0].key);
    }
  }

  function renderTablero() {
    var tbody = document.querySelector('#rgTablaEmpresasResumen tbody');
    var tablero = state.tablero;
    if (!tablero || !tbody) {
      return;
    }
    var totales = tablero.totales || {};
    var operativo = totales.operativo || {};
    var financiero = totales.financiero || {};
    var estadoResultados = totales.estado_resultados || {};

    byId('rgKpiEmpresas').textContent = String(totales.empresas_seleccionadas || 0);
    byId('rgKpiEmpresasActivas').textContent = String(totales.empresas_activas || 0);
    byId('rgKpiVentas').textContent = String(operativo.ventas_cerradas || 0);
    byId('rgKpiIngresos').textContent = fmtMoney(financiero.ingresos || operativo.ingresos_ventas || 0);
    byId('rgKpiEgresos').textContent = fmtMoney(financiero.egresos || 0);
    byId('rgKpiBalance').textContent = fmtMoney(financiero.balance || 0);
    byId('rgKpiClientes').textContent = String(operativo.clientes_activos || 0);
    byId('rgKpiProductos').textContent = String(operativo.productos_activos || 0);
    byId('rgKpiUtilidad').textContent = fmtMoney(estadoResultados.utilidad_operacional || 0);

    var items = Array.isArray(tablero.por_empresa) ? tablero.por_empresa : [];
    renderExecutiveSummary(items, totales);
    renderCharts(items, totales);
    if (!items.length) {
      tbody.innerHTML = '<tr><td colspan="8">No hay empresas seleccionadas.</td></tr>';
      return;
    }
    tbody.innerHTML = items.map(function (item) {
      var empresa = item.empresa || {};
      var empresaTablero = item.tablero || {};
      var op = empresaTablero.operativo || {};
      var fin = empresaTablero.financiero || {};
      return '' +
        '<tr>' +
          '<td>' + escapeHtml(empresa.nombre || ('Empresa ' + empresa.id)) + '</td>' +
          '<td>' + escapeHtml(empresa.estado || 'activo') + '</td>' +
          '<td>' + escapeHtml(op.ventas_cerradas || 0) + '</td>' +
          '<td>' + escapeHtml(fmtMoney(fin.ingresos || op.ingresos_ventas || 0)) + '</td>' +
          '<td>' + escapeHtml(fmtMoney(fin.egresos || 0)) + '</td>' +
          '<td>' + escapeHtml(fmtMoney(fin.balance || 0)) + '</td>' +
          '<td>' + escapeHtml(op.clientes_activos || 0) + '</td>' +
          '<td>' + escapeHtml(op.productos_activos || 0) + '</td>' +
        '</tr>';
    }).join('');
  }

  function renderExecutiveSummary(items, totales) {
    var topEl = byId('rgExecutiveTopEmpresa');
    var topMetaEl = byId('rgExecutiveTopEmpresaMeta');
    var riskEl = byId('rgExecutiveRiskEmpresa');
    var riskMetaEl = byId('rgExecutiveRiskEmpresaMeta');
    var concentrationEl = byId('rgExecutiveConcentration');
    var concentrationMetaEl = byId('rgExecutiveConcentrationMeta');
    var narrativeEl = byId('rgExecutiveNarrative');
    if (!topEl || !topMetaEl || !riskEl || !riskMetaEl || !concentrationEl || !concentrationMetaEl || !narrativeEl) return;

    if (!items.length) {
      topEl.textContent = 'Sin datos';
      topMetaEl.textContent = 'Selecciona empresas para ver la lectura ejecutiva.';
      riskEl.textContent = 'Sin datos';
      riskMetaEl.textContent = 'No hay presión operativa que analizar.';
      concentrationEl.textContent = '0%';
      concentrationMetaEl.textContent = 'No hay participación acumulada.';
      narrativeEl.innerHTML = '<li>No hay información suficiente para construir el reporte ejecutivo.</li>';
      return;
    }

    var topItem = items.slice().sort(function (a, b) {
      return toNumber((b.tablero || {}).financiero && (b.tablero || {}).financiero.ingresos) - toNumber((a.tablero || {}).financiero && (a.tablero || {}).financiero.ingresos);
    })[0];
    var riskItem = items.slice().sort(function (a, b) {
      return toNumber((a.tablero || {}).financiero && (a.tablero || {}).financiero.balance) - toNumber((b.tablero || {}).financiero && (b.tablero || {}).financiero.balance);
    })[0];
    var totalIngresos = toNumber((totales.financiero || {}).ingresos || (totales.operativo || {}).ingresos_ventas || 0);
    var topIngresos = toNumber(((topItem || {}).tablero || {}).financiero && ((topItem || {}).tablero || {}).financiero.ingresos);
    var topShare = pct(topIngresos, totalIngresos);

    topEl.textContent = ((topItem || {}).empresa || {}).nombre || 'Sin datos';
    topMetaEl.textContent = 'Ingresos ' + fmtMoney(topIngresos) + ' | Balance ' + fmtMoney((((topItem || {}).tablero || {}).financiero || {}).balance || 0) + ' | Participación ' + topShare + '%';

    riskEl.textContent = ((riskItem || {}).empresa || {}).nombre || 'Sin datos';
    riskMetaEl.textContent = 'Balance ' + fmtMoney((((riskItem || {}).tablero || {}).financiero || {}).balance || 0) + ' | Egresos ' + fmtMoney((((riskItem || {}).tablero || {}).financiero || {}).egresos || 0) + ' | Ventas ' + String((((riskItem || {}).tablero || {}).operativo || {}).ventas_cerradas || 0);

    concentrationEl.textContent = topShare + '%';
    concentrationMetaEl.textContent = 'La empresa líder concentra ' + topShare + '% de los ingresos del bloque seleccionado.';

    var activas = toNumber(totales.empresas_activas);
    var seleccionadas = toNumber(totales.empresas_seleccionadas);
    var margen = pct((totales.estado_resultados || {}).utilidad_operacional || 0, totalIngresos || 0);
    var narrativas = [
      'Se analizan ' + seleccionadas + ' empresas, con ' + activas + ' activas dentro del grupo filtrado.',
      'El ingreso consolidado es ' + fmtMoney(totalIngresos) + ' y la utilidad operacional representa ' + margen + '% del ingreso total.',
      (((riskItem || {}).empresa || {}).nombre ? ('La empresa a vigilar es ' + ((riskItem || {}).empresa || {}).nombre + ', con balance de ' + fmtMoney((((riskItem || {}).tablero || {}).financiero || {}).balance || 0) + '.') : 'No se detectaron riesgos relevantes.'),
      (((topItem || {}).empresa || {}).nombre ? ('La empresa líder es ' + ((topItem || {}).empresa || {}).nombre + ', con ' + String((((topItem || {}).tablero || {}).operativo || {}).ventas_cerradas || 0) + ' ventas cerradas en el período.') : 'No hay líder destacado por ahora.')
    ];
    narrativeEl.innerHTML = narrativas.map(function (text) {
      return '<li>' + escapeHtml(text) + '</li>';
    }).join('');
  }

  function renderCharts(items, totales) {
    renderIngresosChart(items);
    renderBalanceChart(items);
    renderPortfolioChart(items, totales);
  }

  function renderIngresosChart(items) {
    var host = byId('rgChartIngresos');
    if (!host) return;
    if (!items.length) {
      host.innerHTML = '<div class="reportes-globales-empty">Sin datos para graficar ingresos.</div>';
      return;
    }
    var data = items.slice().sort(function (a, b) {
      return toNumber((b.tablero || {}).financiero && (b.tablero || {}).financiero.ingresos) - toNumber((a.tablero || {}).financiero && (a.tablero || {}).financiero.ingresos);
    }).slice(0, 6);
    var maxValue = data.reduce(function (acc, item) {
      return Math.max(acc, toNumber(((item.tablero || {}).financiero || {}).ingresos || 0));
    }, 0) || 1;
    host.innerHTML = data.map(function (item) {
      var empresa = item.empresa || {};
      var value = toNumber(((item.tablero || {}).financiero || {}).ingresos || 0);
      var width = Math.max(6, Math.round((value * 100) / maxValue));
      return '' +
        '<div class="reportes-globales-bar-row">' +
          '<div class="reportes-globales-bar-copy"><strong>' + escapeHtml(empresa.nombre || ('Empresa ' + empresa.id)) + '</strong><span>' + escapeHtml(fmtMoney(value)) + '</span></div>' +
          '<div class="reportes-globales-bar-track"><span class="reportes-globales-bar-fill fill-income" style="width:' + width + '%"></span></div>' +
        '</div>';
    }).join('');
  }

  function renderBalanceChart(items) {
    var host = byId('rgChartBalance');
    if (!host) return;
    if (!items.length) {
      host.innerHTML = '<div class="reportes-globales-empty">Sin datos para graficar balance.</div>';
      return;
    }
    var data = items.slice(0, 6);
    var maxValue = data.reduce(function (acc, item) {
      var fin = (item.tablero || {}).financiero || {};
      return Math.max(acc, Math.abs(toNumber(fin.balance || 0)), toNumber(fin.egresos || 0));
    }, 0) || 1;
    host.innerHTML = data.map(function (item) {
      var empresa = item.empresa || {};
      var fin = item.tablero && item.tablero.financiero ? item.tablero.financiero : {};
      var balance = toNumber(fin.balance || 0);
      var egresos = toNumber(fin.egresos || 0);
      var balanceWidth = Math.max(4, Math.round((Math.abs(balance) * 100) / maxValue));
      var egresoWidth = Math.max(4, Math.round((egresos * 100) / maxValue));
      return '' +
        '<div class="reportes-globales-balance-row">' +
          '<div class="reportes-globales-bar-copy"><strong>' + escapeHtml(empresa.nombre || ('Empresa ' + empresa.id)) + '</strong><span>Balance ' + escapeHtml(fmtMoney(balance)) + '</span></div>' +
          '<div class="reportes-globales-compare-track">' +
            '<span class="reportes-globales-bar-fill fill-expense" style="width:' + egresoWidth + '%"></span>' +
            '<span class="reportes-globales-bar-fill ' + (balance >= 0 ? 'fill-balance-positive' : 'fill-balance-negative') + '" style="width:' + balanceWidth + '%"></span>' +
          '</div>' +
          '<div class="reportes-globales-legend-inline"><span>Egresos ' + escapeHtml(fmtMoney(egresos)) + '</span></div>' +
        '</div>';
    }).join('');
  }

  function renderPortfolioChart(items, totales) {
    var host = byId('rgChartPortfolio');
    if (!host) return;
    if (!items.length) {
      host.innerHTML = '<div class="reportes-globales-empty">Sin datos para graficar el portafolio.</div>';
      return;
    }
    var activeCount = 0;
    var inactiveCount = 0;
    items.forEach(function (item) {
      var estado = normalize(((item.empresa || {}).estado || 'activo')).toLowerCase();
      if (estado === 'activo') activeCount += 1;
      else inactiveCount += 1;
    });
    var total = activeCount + inactiveCount || 1;
    var activePct = pct(activeCount, total);
    var inactivePct = Math.max(0, 100 - activePct);
    var ingresos = toNumber(((totales || {}).financiero || {}).ingresos || 0);
    var ventas = toNumber(((totales || {}).operativo || {}).ventas_cerradas || 0);
    host.innerHTML = '' +
      '<div class="reportes-globales-portfolio-ring" data-active-pct="' + activePct + '">' +
        '<div class="reportes-globales-portfolio-core"><strong>' + activePct + '%</strong><span>activas</span></div>' +
      '</div>' +
      '<div class="reportes-globales-portfolio-meta">' +
        '<div><strong>' + activeCount + '</strong><span>Activas</span></div>' +
        '<div><strong>' + inactiveCount + '</strong><span>Inactivas</span></div>' +
        '<div><strong>' + fmtCompactNumber(ventas) + '</strong><span>Ventas</span></div>' +
        '<div><strong>' + fmtCompactNumber(ingresos) + '</strong><span>Ingresos</span></div>' +
      '</div>' +
      '<div class="reportes-globales-legend-inline"><span class="legend-income">Activas ' + activePct + '%</span><span class="legend-risk">Inactivas ' + inactivePct + '%</span></div>';

    // Apply dynamic gradient using CSS variables for colors (keeps color tokens in stylesheets)
    var ring = host.querySelector('.reportes-globales-portfolio-ring');
    if (ring) {
      ring.style.background = 'conic-gradient(var(--status-success) 0 ' + activePct + '%, var(--status-danger) ' + activePct + '% 100%)';
    }
  }

  function renderDatasetTable() {
    var thead = document.querySelector('#rgTablaDataset thead');
    var tbody = document.querySelector('#rgTablaDataset tbody');
    var meta = byId('rgDatasetMeta');
    var resumen = byId('rgDatasetResumen');
    if (!thead || !tbody || !meta || !resumen) return;

    var payload = state.dataset;
    var combinado = payload && payload.combinado ? payload.combinado : null;
    if (!combinado || !Array.isArray(combinado.columns)) {
      thead.innerHTML = '';
      tbody.innerHTML = '<tr><td>No hay dataset cargado.</td></tr>';
      meta.textContent = 'Seleccione un dataset y empresas para cargar la información.';
      resumen.textContent = '';
      return;
    }

    meta.textContent = 'Dataset: ' + normalize(payload.dataset_title || payload.dataset_key) + ' | Modo: ' + normalize(payload.modo) + ' | Empresas: ' + String((payload.empresas || []).length) + ' | Filas: ' + String(combinado.row_count || 0);
    var summary = combinado.summary || {};
    var summaryKeys = Object.keys(summary);
    resumen.textContent = summaryKeys.length ? ('Resumen: ' + summaryKeys.map(function (key) {
      var value = summary[key];
      if (typeof value === 'number' && (key.indexOf('ingres') >= 0 || key.indexOf('total') >= 0 || key.indexOf('ticket') >= 0 || key.indexOf('costo') >= 0 || key.indexOf('balance') >= 0)) {
        return key + ': ' + fmtMoney(value);
      }
      return key + ': ' + String(value);
    }).join(' | ')) : '';

    var cols = combinado.columns.slice();
    thead.innerHTML = '<tr>' + cols.map(function (col) { return '<th>' + escapeHtml(col) + '</th>'; }).join('') + '</tr>';
    if (!Array.isArray(combinado.rows) || !combinado.rows.length) {
      tbody.innerHTML = '<tr><td colspan="' + cols.length + '">No hay filas para el filtro actual.</td></tr>';
      return;
    }
    tbody.innerHTML = combinado.rows.map(function (row) {
      return '<tr>' + cols.map(function (col) {
        var value = row[col];
        if (typeof value === 'number' && (col.indexOf('total') >= 0 || col.indexOf('ingres') >= 0 || col.indexOf('costo') >= 0 || col.indexOf('balance') >= 0)) {
          value = fmtMoney(value);
        }
        return '<td>' + escapeHtml(value == null ? '-' : value) + '</td>';
      }).join('') + '</tr>';
    }).join('');
  }

  function renderIndividuales() {
    var wrap = byId('rgIndividuales');
    if (!wrap) return;
    var payload = state.dataset;
    if (!payload || normalize(payload.modo) !== 'individual') {
      wrap.innerHTML = '<div class="reportes-globales-empty">Cambia el modo a separado por empresa para ver cada reporte individual.</div>';
      return;
    }
    var items = Array.isArray(payload.individuales) ? payload.individuales : [];
    if (!items.length) {
      wrap.innerHTML = '<div class="reportes-globales-empty">No hay datasets individuales para el filtro seleccionado.</div>';
      return;
    }
    wrap.innerHTML = items.map(function (item) {
      var empresa = item.empresa || {};
      var dataset = item.dataset || {};
      var cols = Array.isArray(dataset.columns) ? dataset.columns : [];
      var rows = Array.isArray(dataset.rows) ? dataset.rows : [];
      var summary = dataset.summary || {};
      var summaryKeys = Object.keys(summary);
      var summaryText = summaryKeys.length ? summaryKeys.map(function (key) { return key + ': ' + String(summary[key]); }).join(' | ') : 'Sin resumen adicional';
      var table = rows.length ? (
        '<div class="table-wrap"><table class="table"><thead><tr>' + cols.map(function (col) { return '<th>' + escapeHtml(col) + '</th>'; }).join('') + '</tr></thead><tbody>' +
        rows.map(function (row) {
          return '<tr>' + cols.map(function (col) {
            var value = row[col];
            if (typeof value === 'number' && (col.indexOf('total') >= 0 || col.indexOf('ingres') >= 0 || col.indexOf('costo') >= 0 || col.indexOf('balance') >= 0)) {
              value = fmtMoney(value);
            }
            return '<td>' + escapeHtml(value == null ? '-' : value) + '</td>';
          }).join('') + '</tr>';
        }).join('') + '</tbody></table></div>'
      ) : '<div class="reportes-globales-empty">Sin filas para esta empresa.</div>';
      return '' +
        '<article class="reportes-globales-individual-card">' +
          '<h3 >' + escapeHtml(empresa.nombre || ('Empresa ' + empresa.id)) + '</h3>' +
          '<p class="form-help">NIT: ' + escapeHtml(empresa.nit || '-') + ' | Estado: ' + escapeHtml(empresa.estado || 'activo') + ' | Filas: ' + escapeHtml(dataset.row_count || 0) + '</p>' +
          '<p class="form-help">' + escapeHtml(summaryText) + '</p>' +
          table +
        '</article>';
    }).join('');
  }

  async function loadCatalog() {
    var data = await requestJSON('/super/api/reportes_globales?action=catalogo');
    state.empresas = Array.isArray(data.empresas) ? data.empresas : [];
    state.datasets = Array.isArray(data.datasets) ? data.datasets : [];
    state.selectedEmpresaIDs = state.empresas.map(function (empresa) { return Number(empresa.id); });
    renderEmpresas();
    renderDatasetOptions();
    updateSelectionSummary();
  }

  async function refreshAll() {
    if (!state.empresas.length) {
      renderTablero();
      renderDatasetTable();
      renderIndividuales();
      return;
    }
    syncSelectedFromDOM();
    if (!state.selectedEmpresaIDs.length) {
      setMsg('Seleccione al menos una empresa para construir los reportes.', true);
      state.tablero = { por_empresa: [], totales: {} };
      state.dataset = null;
      renderTablero();
      renderDatasetTable();
      renderIndividuales();
      return;
    }
    setMsg('Cargando reportes globales...', false);
    var tableroParams = buildBaseParams('tablero');
    var datasetParams = buildBaseParams('dataset');
    var result = await Promise.all([
      requestJSON('/super/api/reportes_globales?' + tableroParams.toString()),
      requestJSON('/super/api/reportes_globales?' + datasetParams.toString())
    ]);
    state.tablero = result[0];
    state.dataset = result[1];
    renderTablero();
    renderDatasetTable();
    renderIndividuales();
    setMsg('Reportes globales actualizados.', false);
  }

  async function exportDataset() {
    syncSelectedFromDOM();
    if (!state.selectedEmpresaIDs.length) {
      throw new Error('Seleccione al menos una empresa para exportar.');
    }
    var params = buildBaseParams('export');
    params.set('format', normalize(byId('rgFormato') && byId('rgFormato').value) || 'json');
    var res = await fetch('/super/api/reportes_globales?' + params.toString(), { credentials: 'same-origin' });
    if (!res.ok) {
      var body = await res.text();
      throw new Error(body || 'No se pudo exportar el dataset global.');
    }
    var blob = await res.blob();
    var disposition = normalize(res.headers.get('Content-Disposition'));
    var filename = 'reportes_globales.' + normalize(byId('rgFormato') && byId('rgFormato').value || 'json');
    var match = disposition.match(/filename="?([^";]+)"?/i);
    if (match && match[1]) filename = match[1];
    var link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    setTimeout(function () { URL.revokeObjectURL(link.href); }, 0);
  }

  async function sendDatasetByEmail() {
    syncSelectedFromDOM();
    if (!state.selectedEmpresaIDs.length) {
      throw new Error('Seleccione al menos una empresa para enviar el reporte.');
    }
    var toEmail = normalize(byId('rgEmailTo') && byId('rgEmailTo').value);
    if (!toEmail) {
      throw new Error('Ingrese el correo de destino.');
    }
    var params = buildBaseParams('enviar_email');
    var format = normalize(byId('rgFormato') && byId('rgFormato').value) || 'pdf';
    var body = {
      to_email: toEmail,
      format: format,
      dataset: normalize(byId('rgDataset') && byId('rgDataset').value),
      modo: normalize(byId('rgModo') && byId('rgModo').value) || 'consolidado',
      desde: normalize(byId('rgFechaDesde') && byId('rgFechaDesde').value),
      hasta: normalize(byId('rgFechaHasta') && byId('rgFechaHasta').value)
    };
    var res = await fetch('/super/api/reportes_globales?' + params.toString(), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'same-origin',
      body: JSON.stringify(body)
    });
    if (!res.ok) {
      var errText = await res.text();
      throw new Error(errText || 'No se pudo enviar el reporte por correo.');
    }
    return res.json();
  }

  function applyDateDefaults() {
    var today = new Date();
    var from = new Date(today.getTime() - (30 * 24 * 60 * 60 * 1000));
    if (byId('rgFechaDesde')) byId('rgFechaDesde').value = from.toISOString().slice(0, 10);
    if (byId('rgFechaHasta')) byId('rgFechaHasta').value = today.toISOString().slice(0, 10);
  }

  function selectEmpresas(mode) {
    var ids = [];
    if (state.selectionMode === 'single') {
      if (mode === 'active') {
        var activeEmpresa = state.empresas.find(function (empresa) {
          return normalize(empresa.estado || 'activo').toLowerCase() === 'activo';
        });
        ids = activeEmpresa ? [Number(activeEmpresa.id)] : [];
      } else if (state.empresas.length) {
        ids = [Number(state.empresas[0].id)];
      }
      state.selectedEmpresaIDs = ids;
      renderEmpresas();
      refreshAfterSelectionChange();
      return;
    }
    state.empresas.forEach(function (empresa) {
      var estado = normalize(empresa.estado || 'activo').toLowerCase();
      if (mode === 'all') {
        ids.push(Number(empresa.id));
      } else if (mode === 'active' && estado === 'activo') {
        ids.push(Number(empresa.id));
      }
    });
    state.selectedEmpresaIDs = ids;
    renderEmpresas();
    updateSelectionSummary();
  }

  function applySelectionMode(mode) {
    state.selectionMode = mode === 'single' ? 'single' : 'multiple';
    if (state.selectionMode === 'single') {
      if (!state.selectedEmpresaIDs.length && state.empresas.length) {
        state.selectedEmpresaIDs = [Number(state.empresas[0].id)];
      } else {
        state.selectedEmpresaIDs = [Number(state.selectedEmpresaIDs[0] || 0)].filter(function (id) { return id > 0; });
      }
    } else if (!state.selectedEmpresaIDs.length && state.empresas.length) {
      state.selectedEmpresaIDs = state.empresas.map(function (empresa) { return Number(empresa.id); });
    }
    renderEmpresas();
  }

  document.addEventListener('DOMContentLoaded', async function () {
    applyDateDefaults();
    try {
      await loadCatalog();
      await refreshAll();
    } catch (err) {
      setMsg(err.message || 'No se pudo cargar la vista de reportes globales.', true);
    }

    if (byId('rgSeleccionarTodas')) {
      byId('rgSeleccionarTodas').addEventListener('click', function () { selectEmpresas('all'); });
    }
    if (byId('rgSoloActivas')) {
      byId('rgSoloActivas').addEventListener('click', function () { selectEmpresas('active'); });
    }
    if (byId('rgLimpiarEmpresas')) {
      byId('rgLimpiarEmpresas').addEventListener('click', function () {
        state.selectedEmpresaIDs = [];
        renderEmpresas();
        updateSelectionSummary();
      });
    }
    if (byId('rgTipoSeleccion')) {
      byId('rgTipoSeleccion').addEventListener('change', async function () {
        applySelectionMode(byId('rgTipoSeleccion').value);
        if (state.selectionMode === 'single') {
          await refreshAfterSelectionChange();
        } else {
          updateSelectionSummary();
        }
      });
    }
    if (byId('rgEmpresaUnica')) {
      byId('rgEmpresaUnica').addEventListener('change', async function () {
        syncSelectedFromSingleSelect();
        await refreshAfterSelectionChange();
      });
    }
    if (byId('rgAplicar')) {
      byId('rgAplicar').addEventListener('click', async function () {
        try {
          await refreshAll();
        } catch (err) {
          setMsg(err.message || 'No se pudieron aplicar los filtros.', true);
        }
      });
    }
    if (byId('rgActualizar')) {
      byId('rgActualizar').addEventListener('click', async function () {
        try {
          await refreshAll();
        } catch (err) {
          setMsg(err.message || 'No se pudieron actualizar los reportes.', true);
        }
      });
    }
    if (byId('rgVerDataset')) {
      byId('rgVerDataset').addEventListener('click', async function () {
        try {
          await refreshAll();
        } catch (err) {
          setMsg(err.message || 'No se pudo cargar el dataset.', true);
        }
      });
    }
    if (byId('rgExportar')) {
      byId('rgExportar').addEventListener('click', async function () {
        try {
          await exportDataset();
          setMsg('Exportación generada correctamente.', false);
        } catch (err) {
          setMsg(err.message || 'No se pudo exportar el dataset.', true);
        }
      });
    }
    if (byId('rgEnviarEmail')) {
      byId('rgEnviarEmail').addEventListener('click', async function () {
        try {
          var resp = await sendDatasetByEmail();
          setMsg('Reporte enviado a ' + normalize(resp.to_email) + ' (' + normalize(resp.filename) + ').', false);
        } catch (err) {
          setMsg(err.message || 'No se pudo enviar el reporte por correo.', true);
        }
      });
    }
    if (byId('rgImprimir')) {
      byId('rgImprimir').addEventListener('click', function () {
        window.print();
      });
    }
    if (byId('rgModo')) {
      byId('rgModo').addEventListener('change', async function () {
        try {
          await refreshAll();
        } catch (err) {
          setMsg(err.message || 'No se pudo cambiar el modo del reporte.', true);
        }
      });
    }
    if (byId('rgDataset')) {
      byId('rgDataset').addEventListener('change', async function () {
        try {
          await refreshAll();
        } catch (err) {
          setMsg(err.message || 'No se pudo cambiar el dataset.', true);
        }
      });
    }
  });
})();

