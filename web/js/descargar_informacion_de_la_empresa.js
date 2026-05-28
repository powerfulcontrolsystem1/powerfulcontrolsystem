(function () {
  var params = new URLSearchParams(window.location.search || '');
  var empresaID = parseInt(params.get('empresa_id') || params.get('id') || '0', 10);
  var statusEl = document.getElementById('empresaExportStatus');
  var titleEl = document.getElementById('empresaExportTitle');
  var subtitleEl = document.getElementById('empresaExportSubtitle');
  var summaryEl = document.getElementById('empresaExportSummary');
  var formatsEl = document.getElementById('empresaExportFormats');
  var tablesEl = document.getElementById('empresaExportTables');
  var refreshBtn = document.getElementById('empresaExportRefresh');
  var isDownloading = false;

  var formatCatalog = [
    {
      id: 'pdf',
      title: 'PDF ejecutivo',
      description: 'Entrega un consolidado listo para compartir con gerencia, auditoria o archivo documental.',
      badge: 'Informe formal'
    },
    {
      id: 'xls',
      title: 'Excel empresarial',
      description: 'Descarga tabular compatible con hojas de calculo y procesos contables de oficina.',
      badge: 'Analisis y office'
    },
    {
      id: 'csv',
      title: 'CSV interoperable',
      description: 'Formato simple para BI, integraciones, ETL o intercambio con software externo.',
      badge: 'Integracion de datos'
    },
    {
      id: 'json',
      title: 'JSON completo',
      description: 'Snapshot completo con filas por tabla, ideal para respaldo tecnico y revisiones detalladas.',
      badge: 'Snapshot estructurado'
    },
    {
      id: 'txt',
      title: 'TXT operativo',
      description: 'Version legible para revision rapida, soporte o adjuntos livianos de trabajo.',
      badge: 'Lectura rapida'
    }
  ];

  function setStatus(message, isError) {
    if (!statusEl) return;
    statusEl.textContent = message || '';
    statusEl.classList.toggle('is-error', !!isError);
    statusEl.classList.toggle('is-success', !isError && !!message);
  }

  function getFilenameFromHeaders(response, fallbackBaseName, format) {
    var disposition = response && response.headers ? response.headers.get('Content-Disposition') : '';
    var match = disposition && disposition.match(/filename\*=UTF-8''([^;]+)|filename="?([^";]+)"?/i);
    var raw = match ? (match[1] || match[2] || '') : '';
    if (raw) {
      try {
        return decodeURIComponent(raw);
      } catch (err) {
        return raw;
      }
    }
    return fallbackBaseName + '.' + format;
  }

  function buildSafeEmpresaSlug() {
    var rawName = titleEl && titleEl.textContent ? titleEl.textContent : 'empresa';
    return String(rawName)
      .toLowerCase()
      .replace(/descargar informacion de\s+/i, '')
      .replace(/[^a-z0-9]+/g, '_')
      .replace(/^_+|_+$/g, '') || 'empresa';
  }

  function updateDownloadButtonsState() {
    var buttons = document.querySelectorAll('[data-export-format]');
    buttons.forEach(function (button) {
      button.disabled = isDownloading;
      button.classList.toggle('is-loading', isDownloading);
    });
  }

  async function handleDownload(format) {
    if (!empresaID || isDownloading) {
      return;
    }
    isDownloading = true;
    updateDownloadButtonsState();
    setStatus('Preparando descarga ' + String(format || '').toUpperCase() + '...', false);
    try {
      var response = await fetch('/super/api/empresas?id=' + encodeURIComponent(String(empresaID)) + '&action=exportar_informacion&format=' + encodeURIComponent(format), {
        credentials: 'same-origin'
      });
      if (!response.ok) {
        var errorText = await response.text();
        throw new Error(errorText || ('HTTP ' + response.status));
      }
      var blob = await response.blob();
      var fileName = getFilenameFromHeaders(response, 'empresa_' + buildSafeEmpresaSlug() + '_export', format);
      var blobURL = window.URL.createObjectURL(blob);
      var anchor = document.createElement('a');
      anchor.href = blobURL;
      anchor.download = fileName;
      document.body.appendChild(anchor);
      anchor.click();
      anchor.remove();
      window.URL.revokeObjectURL(blobURL);
      setStatus('Descarga lista: ' + fileName, false);
    } catch (err) {
      setStatus('No fue posible descargar el archivo: ' + (err && err.message ? err.message : err), true);
    } finally {
      isDownloading = false;
      updateDownloadButtonsState();
    }
  }

  function escapeHtml(value) {
    return String(value == null ? '' : value).replace(/[&<>"']/g, function (match) {
      return {
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;',
        "'": '&#39;'
      }[match];
    });
  }

  function formatValue(value) {
    if (value == null || value === '') return 'No disponible';
    return String(value);
  }

  function buildSummaryCard(title, value, detail) {
    return '' +
      '<article class="home-offer-card empresa-export-info-card">' +
        '<span class="empresa-export-chip">' + escapeHtml(title) + '</span>' +
        '<h3>' + escapeHtml(value) + '</h3>' +
        '<p>' + escapeHtml(detail) + '</p>' +
      '</article>';
  }

  function buildFormatCard(item) {
    return '' +
      '<article class="home-offer-card empresa-export-format-card">' +
        '<span class="empresa-export-chip">' + escapeHtml(item.badge) + '</span>' +
        '<h3>' + escapeHtml(item.title) + '</h3>' +
        '<p>' + escapeHtml(item.description) + '</p>' +
        '<div class="card-actions">' +
          '<button class="home-offer-btn empresa-export-download-btn" type="button" data-export-format="' + escapeHtml(item.id) + '">Descargar ' + escapeHtml(item.id.toUpperCase()) + '</button>' +
        '</div>' +
      '</article>';
  }

  function buildTableCard(table) {
    var rows = Array.isArray(table.rows) ? table.rows : [];
    var preview = rows.slice(0, 2).map(function (row) {
      return '<pre>' + escapeHtml(JSON.stringify(row, null, 2)) + '</pre>';
    }).join('');
    var meta = [table.source, table.table, String(table.row_count || 0) + ' registros'].join(' · ');
    var truncated = table.truncated ? '<p class="empresa-export-note">La vista previa fue recortada. La descarga incluye la totalidad de registros.</p>' : '';
    return '' +
      '<article class="portal-card empresa-export-table-card">' +
        '<span class="empresa-export-chip">' + escapeHtml(meta) + '</span>' +
        '<h3>' + escapeHtml(table.table || 'Tabla') + '</h3>' +
        '<p>Columnas: ' + escapeHtml((table.columns || []).join(', ')) + '</p>' +
        truncated +
        '<div class="empresa-export-preview">' + preview + '</div>' +
      '</article>';
  }

  function renderSnapshot(snapshot) {
    var empresa = snapshot && snapshot.empresa ? snapshot.empresa : {};
    var warnings = Array.isArray(snapshot.warnings) ? snapshot.warnings : [];
    titleEl.textContent = empresa.nombre ? 'Descargar informacion de ' + empresa.nombre : 'Descargar informacion empresarial';
    subtitleEl.textContent = 'Empresa ID ' + formatValue(empresa.id) + ' · Tipo ' + formatValue(empresa.tipo_nombre || 'General') + ' · Estado ' + formatValue(empresa.estado || 'activo');

    var summaryCards = [
      buildSummaryCard('Empresa', formatValue(empresa.nombre || 'Empresa seleccionada'), 'NIT: ' + formatValue(empresa.nit || 'Sin NIT registrado')),
      buildSummaryCard('Tablas detectadas', formatValue(snapshot.total_tables || 0), 'Se incluyen tablas operativas y tablas super asociadas por empresa_id.'),
      buildSummaryCard('Registros consolidados', formatValue(snapshot.total_rows || 0), 'Total de filas detectadas en el snapshot actual.'),
      buildSummaryCard('Fuentes', 'Operativa ' + formatValue(snapshot.source_totals && snapshot.source_totals.empresas || 0) + ' / Super ' + formatValue(snapshot.source_totals && snapshot.source_totals.super || 0), 'Consolidacion cruzada entre base empresarial y base de licencias o administracion.')
    ];
    if (warnings.length) {
      summaryCards.push(buildSummaryCard('Advertencias', String(warnings.length), warnings.slice(0, 2).join(' | ')));
    }
    summaryEl.innerHTML = summaryCards.join('');

    formatsEl.innerHTML = formatCatalog.map(buildFormatCard).join('');
    formatsEl.querySelectorAll('[data-export-format]').forEach(function (button) {
      button.addEventListener('click', function () {
        handleDownload(button.getAttribute('data-export-format') || '');
      });
    });
    updateDownloadButtonsState();

    var tables = Array.isArray(snapshot.tables) ? snapshot.tables : [];
    if (!tables.length) {
      tablesEl.innerHTML = '<article class="portal-card empresa-export-table-card"><h3>Sin tablas asociadas</h3><p>No se encontraron registros ligados a esta empresa en el momento del resumen.</p></article>';
    } else {
      tablesEl.innerHTML = tables.map(buildTableCard).join('');
    }

    setStatus(warnings.length ? 'Resumen actualizado con advertencias. Puedes descargar la informacion disponible.' : 'Resumen actualizado. Puedes descargar el formato que necesites.', false);
  }

  async function loadSnapshot() {
    if (!empresaID) {
      setStatus('Falta empresa_id en la URL. Abre esta vista desde seleccionar_empresa.html para descargar una empresa concreta.', true);
      return;
    }
    setStatus('Consolidando informacion empresarial...', false);
    try {
      var response = await fetch('/super/api/empresas?id=' + encodeURIComponent(String(empresaID)) + '&action=resumen_descarga', {
        credentials: 'same-origin'
      });
      var text = await response.text();
      var data = {};
      if (text) {
        try {
          data = JSON.parse(text);
        } catch (err) {
          throw new Error('Respuesta invalida del servidor');
        }
      }
      if (!response.ok) {
        throw new Error((data && (data.error || data.message)) || text || ('HTTP ' + response.status));
      }
      renderSnapshot(data.snapshot || {});
    } catch (err) {
      setStatus('No fue posible cargar la informacion: ' + (err && err.message ? err.message : err), true);
      summaryEl.innerHTML = '';
      formatsEl.innerHTML = '';
      tablesEl.innerHTML = '';
    }
  }

  if (refreshBtn) {
    refreshBtn.addEventListener('click', loadSnapshot);
  }

  loadSnapshot();
})();
