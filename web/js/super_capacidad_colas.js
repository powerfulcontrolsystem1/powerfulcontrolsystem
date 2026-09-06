(() => {
  const api = "/super/api/capacidad_colas";
  const cards = document.getElementById("queueCards");
  const rows = document.getElementById("summaryRows");
  const status = document.getElementById("statusMsg");
  let configs = [];
  let snapshots = [];
  const esc = (value) => String(value ?? "").replace(/[&<>"']/g, (char) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[char]);
  const number = (value) => Number.isFinite(Number(value)) ? Number(value) : 0;
  const integer = (value) => Math.max(0, Math.round(number(value)));
  const byLane = (list, lane) => (Array.isArray(list) ? list : []).find((item) => item && item.lane === lane) || {};
  const setStatus = (message, error) => { status.textContent = message || ""; status.className = error ? "form-help error queue-status" : "form-help muted queue-status"; };
  const description = (lane) => lane === "printing" ? "Cola durable por empresa y agente de impresión." : lane === "product_add" ? "Admisión síncrona para crear productos o agregarlos al carrito." : "Reintentos durables, fragmentados y aislados por empresa.";
  const field = (name, label, value, min, disabled) => `<div class="queue-config-row"><label>${esc(label)}</label><input class="form-input ${name}" type="number" min="${min}" step="1" value="${esc(value || 0)}" ${disabled ? "disabled" : ""}></div>`;

  function render() {
    cards.innerHTML = configs.map((config) => {
      const snapshot = byLane(snapshots, config.lane);
      const percent = Math.max(0, number(snapshot.saturation_percent));
      const tone = !snapshot.query_ok ? "err" : percent >= 100 ? "err" : percent >= 75 ? "warn" : "ok";
      const state = !snapshot.query_ok ? "Sin medición" : percent >= 100 ? "Saturada" : percent >= 75 ? "Alta" : "Normal";
      const product = config.lane === "product_add";
      return `<article class="queue-card" data-lane="${esc(config.lane)}"><div class="queue-card-head"><div><h2>${esc(config.label)}</h2><div class="queue-help">${esc(description(config.lane))}</div></div><span class="queue-badge ${tone}">${state}</span></div><div class="queue-kpis"><div class="queue-kpi"><span>${product ? "Solicitudes/min" : "Pendientes"}</span><strong>${esc(product ? snapshot.requests_current_minute || 0 : snapshot.pending || 0)}</strong></div><div class="queue-kpi"><span>Mayor empresa</span><strong>${esc(snapshot.busiest_tenant_pending || 0)}</strong></div><div class="queue-kpi"><span>Procesando</span><strong>${esc(snapshot.processing || 0)}</strong></div><div class="queue-kpi"><span>Antigüedad</span><strong>${esc(Math.round(number(snapshot.oldest_seconds)))} s</strong></div></div><div class="queue-meter ${tone}"><span class="queue-meter-fill" data-saturation="${Math.min(100, percent)}"></span></div><div class="queue-help queue-pressure">Presión: ${percent.toFixed(1)}%</div><div class="queue-config"><label class="queue-alert-label"><input class="alerts-enabled" type="checkbox" ${config.alerts_enabled ? "checked" : ""}> Alertar por saturación</label>${field("rate", "Límite por empresa / minuto", config.rate_limit_per_minute, 1)}${field("pending", "Alerta por pendientes globales", config.pending_alert_threshold, 0, product)}${field("tenant", "Máximo pendiente por empresa", config.max_pending_per_tenant, 0, product)}${field("oldest", "Alerta por antigüedad (segundos)", config.oldest_alert_seconds, 0, product)}</div></article>`;
    }).join("");
    cards.querySelectorAll(".queue-meter-fill").forEach((meter) => { meter.style.width = `${meter.dataset.saturation}%`; });
    rows.innerHTML = snapshots.map((snapshot) => {
      const percent = number(snapshot.saturation_percent);
      const tone = !snapshot.query_ok ? "err" : percent >= 100 ? "err" : percent >= 75 ? "warn" : "ok";
      const value = snapshot.lane === "product_add" ? `${snapshot.requests_current_minute || 0} solicitudes/min` : `${snapshot.pending || 0} trabajos`;
      return `<tr><td>${esc(snapshot.label || snapshot.lane)}</td><td>${esc(value)}</td><td>${esc(snapshot.processing || 0)}</td><td>${esc(snapshot.failed || 0)}</td><td>${esc(snapshot.active_tenants || 0)}</td><td>Empresa ${esc(snapshot.busiest_tenant_id || "-")}: ${esc(snapshot.busiest_tenant_pending || 0)}</td><td>${Math.round(number(snapshot.oldest_seconds))} s</td><td><span class="queue-badge ${tone}">${!snapshot.query_ok ? "Error" : percent >= 100 ? "Saturada" : percent >= 75 ? "Alta" : "Normal"}</span></td></tr>`;
    }).join("") || '<tr><td colspan="8" class="muted">Sin mediciones disponibles.</td></tr>';
  }
  async function load() { setStatus("Consultando capacidad…", false); const response = await fetch(api, { credentials: "same-origin" }); const data = await response.json().catch(() => null); if (!data) throw new Error("Respuesta inválida del servidor"); configs = Array.isArray(data.configs) ? data.configs : []; snapshots = Array.isArray(data.snapshots) ? data.snapshots : []; render(); if (!response.ok || data.ok === false) throw new Error(data.error || `HTTP ${response.status}`); setStatus("Capacidad actualizada.", false); }
  function readConfigs() { return Array.from(cards.querySelectorAll(".queue-card")).map((card) => { const previous = byLane(configs, card.dataset.lane); return { lane: card.dataset.lane, label: previous.label || card.dataset.lane, alerts_enabled: card.querySelector(".alerts-enabled").checked, rate_limit_per_minute: Math.max(1, integer(card.querySelector(".rate").value)), pending_alert_threshold: integer(card.querySelector(".pending").value), oldest_alert_seconds: integer(card.querySelector(".oldest").value), max_pending_per_tenant: integer(card.querySelector(".tenant").value) }; }); }
  async function save() { setStatus("Guardando configuración…", false); const response = await fetch(api, { method: "PUT", credentials: "same-origin", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ configs: readConfigs() }) }); const data = await response.json().catch(() => null); if (!response.ok || !data || data.ok === false) throw new Error((data && data.error) || `HTTP ${response.status}`); configs = data.configs || configs; await load(); setStatus("Configuración guardada. Los procesos la aplican en pocos segundos.", false); }
  async function evaluate() { setStatus("Evaluando y enviando alertas activas…", false); const response = await fetch(`${api}?action=evaluate`, { method: "POST", credentials: "same-origin" }); const data = await response.json().catch(() => null); if (data) { configs = data.configs || configs; snapshots = data.snapshots || snapshots; render(); } if (!response.ok || !data || data.ok === false) throw new Error((data && data.error) || `HTTP ${response.status}`); setStatus("Evaluación completada. Solo se envían alertas cuando un umbral está superado.", false); }
  document.getElementById("reloadBtn").addEventListener("click", () => load().catch((error) => setStatus(error.message, true)));
  document.getElementById("saveBtn").addEventListener("click", () => save().catch((error) => setStatus(error.message, true)));
  document.getElementById("evaluateBtn").addEventListener("click", () => evaluate().catch((error) => setStatus(error.message, true)));
  load().catch((error) => setStatus(error.message, true));
})();
