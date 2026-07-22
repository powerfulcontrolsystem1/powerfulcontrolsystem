const securityState = {
  activeScanId: '',
  pollTimer: null,
  lastReport: null,
};

function qs(id) {
  return document.getElementById(id);
}

function readCSRFCookie() {
  const match = String(document.cookie || '').match(/(?:^|;\s*)pcs_csrf=([^;]+)/);
  if (!match) return '';
  try {
    return decodeURIComponent(match[1]);
  } catch (error) {
    return '';
  }
}

async function fetchJSON(url, options) {
  const requestOptions = Object.assign({}, options || {});
  const method = String(requestOptions.method || 'GET').toUpperCase();
  if (method !== 'GET' && method !== 'HEAD') {
    const token = readCSRFCookie();
    if (token) {
      const headers = new Headers(requestOptions.headers || {});
      if (!headers.has('X-CSRF-Token')) headers.set('X-CSRF-Token', token);
      requestOptions.headers = headers;
    }
  }
  const response = await fetch(url, requestOptions);
  let payload = null;
  try {
    payload = await response.json();
  } catch (error) {
    if (!response.ok) {
      throw new Error('HTTP ' + response.status);
    }
  }
  if (!response.ok) {
    const message = payload && payload.error ? payload.error : 'HTTP ' + response.status;
    throw new Error(message);
  }
  return payload || {};
}

function setConfigState(text, variant) {
  const badge = qs('configState');
  if (!badge) return;
  badge.textContent = text;
  badge.className = 'state-pill ' + (variant || 'neutral');
}

function setRunState(text, variant) {
  const badge = qs('runStateBadge');
  if (!badge) return;
  badge.textContent = text;
  badge.className = 'state-pill ' + (variant || 'neutral');
}

function normalizeVariant(value) {
  const key = String(value || '').toLowerCase();
  if (key.includes('fail') || key.includes('error')) return 'failed';
  if (key.includes('run')) return 'running';
  if (key.includes('complete') || key.includes('ok')) return 'completed';
  return 'neutral';
}

function escapeHTML(value) {
  return String(value || '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

function severityClass(value) {
  const normalized = String(value || '').toLowerCase();
  if (normalized.includes('crit')) return 'critico';
  if (normalized.includes('alto') || normalized.includes('high')) return 'alto';
  if (normalized.includes('medio') || normalized.includes('medium')) return 'medio';
  if (normalized.includes('bajo') || normalized.includes('low')) return 'bajo';
  return 'info';
}

function formatDate(value) {
  if (!value) return 'Sin fecha';
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleString('es-CO');
}

function fillConfigForm(config) {
  qs('targetHostInput').value = config.target_host || '127.0.0.1';
  qs('portListInput').value = config.port_list || '22,80,443';
  qs('profileSelect').value = config.profile || 'full';
  qs('cronInput').value = config.schedule && config.schedule.cron ? config.schedule.cron : '0 2 * * *';
  qs('maxHistoryInput').value = config.max_history || 60;
  qs('maxFindingsInput').value = config.max_findings_per_tool || 150;
  qs('lynisEnabledInput').checked = !!(config.lynis && config.lynis.enabled);
  qs('lynisCommandInput').value = config.lynis && config.lynis.command ? config.lynis.command : 'lynis';
  qs('nmapEnabledInput').checked = !!(config.nmap && config.nmap.enabled);
  qs('nmapCommandInput').value = config.nmap && config.nmap.command ? config.nmap.command : 'nmap';
  qs('vulnEnabledInput').checked = !!(config.vulnerability_scan && config.vulnerability_scan.enabled);
  qs('vulnProviderInput').value = config.vulnerability_scan && config.vulnerability_scan.provider ? config.vulnerability_scan.provider : 'trivy';
  qs('vulnCommandInput').value = config.vulnerability_scan && config.vulnerability_scan.command ? config.vulnerability_scan.command : 'trivy';
  qs('vulnTargetInput').value = config.vulnerability_scan && config.vulnerability_scan.target_path ? config.vulnerability_scan.target_path : '/';
}

function collectConfigPayload() {
  return {
    target_host: qs('targetHostInput').value.trim(),
    port_list: qs('portListInput').value.trim(),
    profile: qs('profileSelect').value,
    max_history: Number(qs('maxHistoryInput').value || 60),
    max_findings_per_tool: Number(qs('maxFindingsInput').value || 150),
    schedule: {
      enabled: true,
      cron: qs('cronInput').value.trim(),
    },
    lynis: {
      enabled: qs('lynisEnabledInput').checked,
      command: qs('lynisCommandInput').value.trim(),
    },
    nmap: {
      enabled: qs('nmapEnabledInput').checked,
      command: qs('nmapCommandInput').value.trim(),
    },
    vulnerability_scan: {
      enabled: qs('vulnEnabledInput').checked,
      provider: qs('vulnProviderInput').value,
      command: qs('vulnCommandInput').value.trim(),
      target_path: qs('vulnTargetInput').value.trim(),
    },
  };
}

async function loadConfig() {
  setConfigState('Cargando', 'neutral');
  try {
    const payload = await fetchJSON('/super/api/security/vps/config');
    fillConfigForm(payload.config || {});
    setConfigState('Listo', 'completed');
    if (payload.status) {
      renderStatus(payload.status);
    }
  } catch (error) {
    console.error(error);
    setConfigState('Error', 'failed');
    showStatusMessage('No se pudo cargar la configuración: ' + error.message, 'failed');
  }
}

async function saveConfig(event) {
  event.preventDefault();
  setConfigState('Guardando', 'running');
  try {
    await fetchJSON('/super/api/security/vps/config', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(collectConfigPayload()),
    });
    setConfigState('Guardado', 'completed');
    showStatusMessage('Configuración guardada correctamente.', 'completed');
  } catch (error) {
    console.error(error);
    setConfigState('Error', 'failed');
    showStatusMessage('No se pudo guardar la configuración: ' + error.message, 'failed');
  }
}

function showStatusMessage(message, variant) {
  const box = qs('statusBox');
  if (!box) return;
  box.hidden = false;
  box.textContent = message;
  box.style.borderColor = variant === 'failed' ? 'rgba(143, 45, 31, 0.22)' : variant === 'completed' ? 'rgba(42, 127, 98, 0.22)' : 'rgba(51, 101, 138, 0.14)';
  box.style.background = variant === 'failed' ? 'rgba(143, 45, 31, 0.08)' : variant === 'completed' ? 'rgba(42, 127, 98, 0.08)' : 'rgba(51, 101, 138, 0.08)';
}

function renderSummary(report) {
  const grid = qs('summaryGrid');
  if (!grid) return;
  if (!report) {
    grid.innerHTML = '<p class="empty-state">Todavía no hay resultados guardados.</p>';
    return;
  }
  const summary = report.summary || {};
  const metrics = [
    { label: 'Crítico', value: summary.critical || 0 },
    { label: 'Alto', value: summary.high || 0 },
    { label: 'Medio', value: summary.medium || 0 },
    { label: 'Bajo', value: summary.low || 0 },
    { label: 'Total', value: summary.total_findings || 0 },
    { label: 'Salud', value: summary.health || 'estable' },
  ];
  if (summary.hardening_index) {
    metrics.push({ label: 'Hardening', value: summary.hardening_index });
  }
  if (Array.isArray(summary.open_ports) && summary.open_ports.length) {
    metrics.push({ label: 'Puertos', value: summary.open_ports.join(', ') });
  }
  grid.innerHTML = metrics.map(function(metric) {
    return '<div class="summary-tile"><strong>' + escapeHTML(metric.label) + '</strong><div class="summary-value">' + escapeHTML(metric.value) + '</div></div>';
  }).join('');
}

function renderTools(report) {
  const grid = qs('toolGrid');
  if (!grid) return;
  const tools = report && Array.isArray(report.tools) ? report.tools : [];
  if (!tools.length) {
    grid.innerHTML = '<p class="empty-state">Sin herramientas ejecutadas todavía.</p>';
    return;
  }
  grid.innerHTML = tools.map(function(tool) {
    return [
      '<div class="tool-tile">',
      '<strong>' + escapeHTML(tool.display_name || tool.name || 'Herramienta') + '</strong>',
      '<span class="tool-pill neutral">' + escapeHTML(tool.status || 'sin estado') + '</span>',
      '<p>' + escapeHTML(tool.summary || 'Sin resumen') + '</p>',
      tool.error ? '<p class="mini-note">' + escapeHTML(tool.error) + '</p>' : '',
      '</div>'
    ].join('');
  }).join('');
}

function renderFindings(report) {
  const list = qs('findingsList');
  const badge = qs('findingsCountBadge');
  if (!list || !badge) return;
  const findings = report && Array.isArray(report.findings) ? report.findings : [];
  badge.textContent = findings.length + ' hallazgos';
  badge.className = 'state-pill ' + (findings.length ? 'running' : 'neutral');
  if (!findings.length) {
    list.innerHTML = '<p class="empty-state">El reporte actual no tiene hallazgos registrados.</p>';
    return;
  }
  list.innerHTML = findings.map(function(finding) {
    const severity = severityClass(finding.severity);
    const meta = [];
    if (finding.tool) meta.push('Herramienta: ' + finding.tool);
    if (finding.category) meta.push('Categoría: ' + finding.category);
    if (finding.target) meta.push('Objetivo: ' + finding.target);
    if (finding.port) meta.push('Puerto: ' + finding.port);
    if (finding.service) meta.push('Servicio: ' + finding.service);
    if (finding.reference) meta.push('Ref: ' + finding.reference);
    return [
      '<article class="finding-card ' + severity + '">',
      '<div class="finding-header">',
      '<div>',
      '<h3 class="finding-title">' + escapeHTML(finding.title || 'Hallazgo') + '</h3>',
      '<div class="finding-meta">' + meta.map(escapeHTML).join('<span>•</span>') + '</div>',
      '</div>',
      '<span class="severity-pill ' + severity + '">' + escapeHTML(finding.severity || 'INFO') + '</span>',
      '</div>',
      '<div class="finding-body">',
      finding.description ? '<p><strong>Descripción:</strong> ' + escapeHTML(finding.description) + '</p>' : '',
      finding.evidence ? '<p><strong>Evidencia:</strong> ' + escapeHTML(finding.evidence) + '</p>' : '',
      finding.recommendation ? '<p><strong>Acción sugerida:</strong> ' + escapeHTML(finding.recommendation) + '</p>' : '',
      '</div>',
      '</article>'
    ].join('');
  }).join('');
}

function renderReportActions(report) {
  const container = qs('reportActions');
  if (!container) return;
  const reportsMap = report && report.reports ? report.reports : null;
  if (!reportsMap) {
    container.innerHTML = '';
    return;
  }
  const order = ['json', 'txt', 'html', 'csv', 'pdf', 'xls'];
  const labels = { json: 'JSON', txt: 'TXT', html: 'HTML', csv: 'CSV', pdf: 'PDF', xls: 'Excel' };
  container.innerHTML = order.filter(function(format) {
    return reportsMap[format];
  }).map(function(format) {
    return '<a class="export-link" href="' + escapeHTML(reportsMap[format]) + '">' + labels[format] + '</a>';
  }).join('');
}

function renderComparison(report, comparisonOverride) {
  const box = qs('comparisonBox');
  if (!box) return;
  const comparison = comparisonOverride || (report ? report.comparison : null);
  if (!comparison || !comparison.previous_scan_id) {
    box.hidden = true;
    box.innerHTML = '';
    return;
  }
  box.hidden = false;
  box.innerHTML = [
    '<strong>Comparación contra ' + escapeHTML(comparison.previous_scan_id) + '</strong><br>',
    'Fecha previa: ' + escapeHTML(comparison.previous_generated_at || 'N/D') + '<br>',
    'Nuevos: ' + escapeHTML(comparison.new_findings || 0) + ' · Resueltos: ' + escapeHTML(comparison.resolved_findings || 0) + '<br>',
    escapeHTML(comparison.summary || 'Sin resumen adicional'),
    Array.isArray(comparison.new_open_ports) && comparison.new_open_ports.length ? '<br>Puertos nuevos: ' + escapeHTML(comparison.new_open_ports.join(', ')) : '',
    Array.isArray(comparison.closed_ports) && comparison.closed_ports.length ? '<br>Puertos cerrados: ' + escapeHTML(comparison.closed_ports.join(', ')) : ''
  ].join('');
}

function renderStatus(status) {
  if (!status) return;
  securityState.activeScanId = status.active ? status.scan_id || '' : '';
  setRunState(status.active ? 'Escaneo en curso' : (status.status || 'Sin escaneos'), normalizeVariant(status.active ? 'running' : status.status));
  if (status.report) {
    securityState.lastReport = status.report;
    renderSummary(status.report);
    renderTools(status.report);
    renderFindings(status.report);
    renderReportActions(status.report);
    renderComparison(status.report);
  }
  if (!status.active && securityState.pollTimer) {
    window.clearInterval(securityState.pollTimer);
    securityState.pollTimer = null;
    loadHistory();
  }
}

async function startScan(event) {
  if (event) event.preventDefault();
  setRunState('Lanzando escaneo', 'running');
  showStatusMessage('Solicitando ejecución del escaneo...', 'neutral');
  try {
    const payload = {
      target_host: qs('targetHostInput').value.trim(),
      port_list: qs('portListInput').value.trim(),
      profile: qs('profileSelect').value,
      trigger: 'super_panel',
    };
    const response = await fetchJSON('/super/api/security/vps/run', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    const status = response.status || {};
    renderStatus(status);
    showStatusMessage('Escaneo iniciado: ' + (status.scan_id || ''), 'completed');
    if (status.scan_id) {
      pollStatus(status.scan_id);
    }
  } catch (error) {
    console.error(error);
    setRunState('Error', 'failed');
    showStatusMessage('No se pudo iniciar el escaneo: ' + error.message, 'failed');
  }
}

function pollStatus(scanId) {
  if (!scanId) return;
  if (securityState.pollTimer) {
    window.clearInterval(securityState.pollTimer);
  }
  securityState.pollTimer = window.setInterval(async function() {
    try {
      const payload = await fetchJSON('/super/api/security/vps/status?scan_id=' + encodeURIComponent(scanId));
      renderStatus(payload.status || {});
    } catch (error) {
      console.error(error);
      window.clearInterval(securityState.pollTimer);
      securityState.pollTimer = null;
      showStatusMessage('No se pudo consultar el estado del escaneo: ' + error.message, 'failed');
    }
  }, 3000);
}

async function loadCurrentStatus() {
  try {
    const payload = await fetchJSON('/super/api/security/vps/status');
    renderStatus(payload.status || {});
  } catch (error) {
    console.error(error);
    showStatusMessage('No se pudo cargar el estado actual: ' + error.message, 'failed');
  }
}

async function loadHistory() {
  const tbody = document.querySelector('#historyTable tbody');
  if (!tbody) return;
  try {
    const payload = await fetchJSON('/super/api/security/vps/history?limit=12');
    const items = Array.isArray(payload.items) ? payload.items : [];
    if (!items.length) {
      tbody.innerHTML = '<tr><td colspan="7">Sin historial todavía.</td></tr>';
      return;
    }
    tbody.innerHTML = items.map(function(item) {
      const severity = severityClass(item.highest_severity);
      const reportsMap = item.reports || {};
      const exportButtons = ['json', 'txt', 'pdf', 'xls'].filter(function(format) {
        return reportsMap[format];
      }).map(function(format) {
        const labels = { json: 'JSON', txt: 'TXT', pdf: 'PDF', xls: 'Excel' };
        return '<a class="export-link" href="' + escapeHTML(reportsMap[format]) + '">' + labels[format] + '</a>';
      }).join('');
      return [
        '<tr>',
        '<td>' + escapeHTML(formatDate(item.generated_at)) + '</td>',
        '<td class="mono">' + escapeHTML(item.scan_id) + '</td>',
        '<td>' + escapeHTML(item.target_host || 'N/D') + '<br><small>' + escapeHTML(item.profile || '') + '</small></td>',
        '<td><span class="severity-pill ' + severity + '">' + escapeHTML(item.highest_severity || 'INFO') + '</span></td>',
        '<td>' + escapeHTML(item.total_findings || 0) + '</td>',
        '<td>' + escapeHTML((item.new_findings || 0) + ' / ' + (item.resolved_findings || 0)) + '</td>',
        '<td class="actions-cell"><div class="inline-actions">',
        '<button type="button" class="subtle-button" data-action="load-report" data-scan-id="' + escapeHTML(item.scan_id) + '">Ver</button>',
        '<button type="button" class="subtle-button" data-action="compare" data-scan-id="' + escapeHTML(item.scan_id) + '">Comparar</button>',
        exportButtons,
        '</div></td>',
        '</tr>'
      ].join('');
    }).join('');
  } catch (error) {
    console.error(error);
    tbody.innerHTML = '<tr><td colspan="7">Error cargando historial: ' + escapeHTML(error.message) + '</td></tr>';
  }
}

async function loadReportByScanId(scanId) {
  try {
    const payload = await fetchJSON('/super/api/security/vps/status?scan_id=' + encodeURIComponent(scanId));
    renderStatus(payload.status || {});
  } catch (error) {
    console.error(error);
    showStatusMessage('No se pudo cargar el reporte solicitado: ' + error.message, 'failed');
  }
}

async function loadComparison(scanId) {
  try {
    const payload = await fetchJSON('/super/api/security/vps/compare?scan_id=' + encodeURIComponent(scanId));
    renderComparison(securityState.lastReport, payload.comparison || {});
  } catch (error) {
    console.error(error);
    showStatusMessage('No se pudo cargar la comparación: ' + error.message, 'failed');
  }
}

async function scanPorts() {
  const ip = qs('ipInput').value.trim() || '127.0.0.1';
  const ports = qs('portsInput').value.trim() || '22,80,443';
  const button = qs('scanPortsBtn');
  button.disabled = true;
  button.textContent = 'Escaneando...';
  try {
    const response = await fetch('/super/api/security/ports?ip=' + encodeURIComponent(ip) + '&ports=' + encodeURIComponent(ports));
    if (!response.ok) throw new Error('HTTP ' + response.status);
    const data = await response.json();
    const tbody = document.querySelector('#portsTable tbody');
    tbody.innerHTML = (Array.isArray(data) ? data : []).map(function(item) {
      const variant = item.estado === 'abierto' ? 'running' : 'completed';
      return '<tr><td>' + escapeHTML(item.puerto) + '</td><td><span class="state-pill ' + variant + '">' + escapeHTML(item.estado) + '</span></td><td>' + escapeHTML(item.ip || '') + '</td><td>' + escapeHTML(item.firewall || 'Desconocido') + '</td></tr>';
    }).join('');
  } catch (error) {
    console.error(error);
    showStatusMessage('Error escaneando puertos: ' + error.message, 'failed');
  } finally {
    button.disabled = false;
    button.textContent = 'Escanear puertos';
  }
}

async function loadProcesses() {
  const button = qs('refreshProcessesBtn');
  button.disabled = true;
  button.textContent = 'Cargando...';
  try {
    const response = await fetch('/super/api/security/processes?limit=200');
    if (!response.ok) throw new Error('HTTP ' + response.status);
    const data = await response.json();
    const tbody = document.querySelector('#processesTable tbody');
    tbody.innerHTML = (Array.isArray(data) ? data : []).map(function(item) {
      return '<tr><td>' + escapeHTML(item.pid) + '</td><td>' + escapeHTML(item.name || '') + '</td><td>' + escapeHTML(item.memory_kb || 0) + '</td></tr>';
    }).join('');
  } catch (error) {
    console.error(error);
    showStatusMessage('Error cargando procesos: ' + error.message, 'failed');
  } finally {
    button.disabled = false;
    button.textContent = 'Actualizar procesos';
  }
}

function wireHistoryActions() {
  const tbody = document.querySelector('#historyTable tbody');
  if (!tbody) return;
  tbody.addEventListener('click', function(event) {
    const target = event.target.closest('[data-action]');
    if (!target) return;
    const action = target.getAttribute('data-action');
    const scanId = target.getAttribute('data-scan-id');
    if (!scanId) return;
    if (action === 'load-report') {
      loadReportByScanId(scanId);
    }
    if (action === 'compare') {
      loadComparison(scanId);
    }
  });
}

document.addEventListener('DOMContentLoaded', function() {
  wireHistoryActions();
  qs('configForm').addEventListener('submit', saveConfig);
  qs('runFullScanBtn').addEventListener('click', startScan);
  qs('runQuickScanBtn').addEventListener('click', startScan);
  qs('refreshHistoryBtn').addEventListener('click', loadHistory);
  qs('scanPortsBtn').addEventListener('click', scanPorts);
  qs('refreshProcessesBtn').addEventListener('click', loadProcesses);
  loadConfig();
  loadCurrentStatus();
  loadHistory();
  scanPorts();
  loadProcesses();
});
