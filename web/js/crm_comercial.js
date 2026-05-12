(function () {
  "use strict";

  const LEAD_TRANSITIONS = {
    nuevo: ["contactado", "descalificado"],
    contactado: ["calificado", "descalificado"],
    calificado: ["propuesta", "descalificado"],
    propuesta: ["negociacion", "ganado", "perdido", "descalificado"],
    negociacion: ["ganado", "perdido", "descalificado"],
    perdido: ["reactivado"],
    reactivado: ["contactado", "calificado", "descalificado"],
    ganado: ["postventa"],
    postventa: ["cerrado"],
    descalificado: [],
    cerrado: []
  };

  const INTERACTION_TRANSITIONS = {
    abierta: ["en_progreso", "cerrada", "cancelada"],
    en_progreso: ["cerrada", "cancelada"],
    cerrada: ["reabierta"],
    reabierta: ["en_progreso", "cerrada", "cancelada"],
    cancelada: ["reabierta"]
  };

  const CAMPAIGN_TRANSITIONS = {
    planificada: ["activa", "cancelada"],
    activa: ["pausada", "finalizada", "cancelada"],
    pausada: ["activa", "finalizada", "cancelada"],
    finalizada: ["archivada"],
    archivada: [],
    cancelada: []
  };

  const QUOTE_TRANSITIONS = {
    borrador: ["emitida", "anulada"],
    emitida: ["aprobada", "rechazada", "vencida", "anulada"],
    aprobada: ["convertida", "anulada"],
    rechazada: ["borrador", "anulada"],
    vencida: ["borrador", "anulada"],
    convertida: [],
    anulada: []
  };

  const TAB_PANEL = {
    tablero: "crmPanelTablero",
    leads: "crmPanelLeads",
    interacciones: "crmPanelInteracciones",
    cotizaciones: "crmPanelCotizaciones",
    campanas: "crmPanelCampanas",
    forecast: "crmPanelForecast",
    metas: "crmPanelMetas",
    embudo: "crmPanelEmbudo",
    ayuda: "crmPanelAyuda"
  };

  const state = {
    empresaID: 0,
    q: "",
    includeInactive: false,
    periodo: "",
    activeTab: "tablero",
    leads: [],
    interacciones: [],
    campanas: [],
    cotizaciones: [],
    embudo: { summary: {}, items: [], alertas: [] },
    advanced: { embudo: [], agenda: [], top_leads: [], metas: [], alertas: [], responsables: [], canales: [], acciones_prioritarias: [] }
  };

  function $(id) {
    return document.getElementById(id);
  }

  function normalize(value) {
    return String(value == null ? "" : value).trim();
  }

  function asNumber(value) {
    const n = Number(value || 0);
    return Number.isFinite(n) ? n : 0;
  }

  function esc(value) {
    return String(value == null ? "" : value)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#39;");
  }

  function queryParam(name) {
    return new URLSearchParams(window.location.search || "").get(name);
  }

  function currentMonth() {
    const d = new Date();
    return d.getFullYear() + "-" + String(d.getMonth() + 1).padStart(2, "0");
  }

  function resolveEmpresaID() {
    if (typeof window.__resolveEmpresaIdContext === "function") {
      const resolved = Number(window.__resolveEmpresaIdContext() || 0);
      if (resolved > 0) return resolved;
    }
    const fromUrl = Number(queryParam("empresa_id") || queryParam("id") || 0);
    if (fromUrl > 0) return fromUrl;
    try {
      return Number(sessionStorage.getItem("active_empresa_id") || sessionStorage.getItem("empresa_id") || localStorage.getItem("active_empresa_id") || localStorage.getItem("empresa_id") || 0);
    } catch (_) {
      return 0;
    }
  }

  function safeArray(value) {
    if (Array.isArray(value)) return value;
    if (value && Array.isArray(value.items)) return value.items;
    if (value && Array.isArray(value.data)) return value.data;
    return [];
  }

  function formatMoney(value) {
    return new Intl.NumberFormat("es-CO", { style: "currency", currency: "COP", maximumFractionDigits: 0 }).format(asNumber(value));
  }

  function formatNumber(value, digits) {
    return new Intl.NumberFormat("es-CO", { minimumFractionDigits: digits || 0, maximumFractionDigits: digits || 0 }).format(asNumber(value));
  }

  function formatDate(value) {
    const raw = normalize(value);
    if (!raw) return "-";
    return raw.length >= 10 ? raw.slice(0, 10) : raw;
  }

  function formatDateTime(value) {
    const raw = normalize(value);
    if (!raw) return "-";
    const date = new Date(raw.replace(" ", "T"));
    if (Number.isNaN(date.getTime())) return raw;
    return new Intl.DateTimeFormat("es-CO", { dateStyle: "medium", timeStyle: "short" }).format(date);
  }

  function toDateInputValue(value) {
    const raw = normalize(value);
    return raw.length >= 10 ? raw.slice(0, 10) : "";
  }

  function toDateTimeInputValue(value) {
    const raw = normalize(value);
    if (!raw) return "";
    const normalized = raw.replace(" ", "T");
    const date = new Date(normalized);
    if (!Number.isNaN(date.getTime())) {
      const year = date.getFullYear();
      const month = String(date.getMonth() + 1).padStart(2, "0");
      const day = String(date.getDate()).padStart(2, "0");
      const hour = String(date.getHours()).padStart(2, "0");
      const minute = String(date.getMinutes()).padStart(2, "0");
      return year + "-" + month + "-" + day + "T" + hour + ":" + minute;
    }
    return normalized.slice(0, 16);
  }

  function todayInput() {
    return new Date().toISOString().slice(0, 10);
  }

  function setValue(id, value) {
    const el = $(id);
    if (el) el.value = value == null ? "" : String(value);
  }

  function setMessage(id, text, kind) {
    const el = $(id);
    if (!el) return;
    el.textContent = text || "";
    el.className = "form-help";
    if (id === "crmMsg") el.classList.add("crm-msg");
    if (kind === "error") el.classList.add("value-negative");
    if (kind === "success") el.classList.add("value-positive");
  }

  async function fetchJSON(url, options) {
    const response = await fetch(url, Object.assign({ credentials: "same-origin" }, options || {}));
    const contentType = String(response.headers.get("Content-Type") || "").toLowerCase();
    if (!response.ok) {
      const text = await response.text();
      throw new Error(text || ("HTTP " + response.status));
    }
    if (contentType.indexOf("application/json") >= 0) return response.json();
    const text = await response.text();
    return text ? JSON.parse(text) : {};
  }

  async function safeFetch(url, fallback) {
    try {
      return await fetchJSON(url);
    } catch (err) {
      console.warn("[CRM]", err);
      return fallback;
    }
  }

  function buildURL(path, extraParams) {
    const url = new URL(path, window.location.origin);
    url.searchParams.set("empresa_id", String(state.empresaID));
    if (state.includeInactive) url.searchParams.set("include_inactive", "1");
    if (state.q) url.searchParams.set("q", state.q);
    Object.keys(extraParams || {}).forEach(function (key) {
      const value = extraParams[key];
      if (value == null || value === "") return;
      url.searchParams.set(key, String(value));
    });
    return url.pathname + url.search;
  }

  function postJSON(path, payload) {
    return fetchJSON(path, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(Object.assign({ empresa_id: state.empresaID }, payload || {}))
    });
  }

  function putJSON(path, payload) {
    return fetchJSON(path, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(Object.assign({ empresa_id: state.empresaID }, payload || {}))
    });
  }

  function badgeClassForStatus(status) {
    const value = normalize(status).toLowerCase();
    if (["ganado", "aprobada", "activa", "cerrada", "convertida", "postventa", "finalizada"].indexOf(value) >= 0) return "crm-pill is-success";
    if (["perdido", "descalificado", "cancelada", "anulada", "rechazada", "vencida"].indexOf(value) >= 0) return "crm-pill is-danger";
    if (["negociacion", "propuesta", "contactado", "emitida", "pausada", "en_progreso", "calificado"].indexOf(value) >= 0) return "crm-pill is-warning";
    return "crm-pill";
  }

  function statusBadge(status) {
    return "<span class=\"" + badgeClassForStatus(status) + "\">" + esc(normalize(status) || "-") + "</span>";
  }

  function allowedTransitions(map, current) {
    return map[normalize(current).toLowerCase()] || [];
  }

  function transitionSelect(kind, id, map, current) {
    const options = allowedTransitions(map, current);
    if (!options.length) return "<span class=\"form-help\">Sin transicion</span>";
    return "<select class=\"form-input crm-state-select\" data-transition-kind=\"" + esc(kind) + "\" data-transition-id=\"" + Number(id || 0) + "\">" +
      "<option value=\"\">Cambiar estado</option>" +
      options.map(function (item) { return "<option value=\"" + esc(item) + "\">" + esc(item) + "</option>"; }).join("") +
      "</select>";
  }

  function exportRowsToCSV(filename, rows) {
    if (!Array.isArray(rows) || !rows.length) throw new Error("No hay registros para exportar.");
    const keys = Array.from(rows.reduce(function (acc, row) {
      Object.keys(row || {}).forEach(function (key) { acc.add(key); });
      return acc;
    }, new Set()));
    const lines = [keys.join(",")];
    rows.forEach(function (row) {
      lines.push(keys.map(function (key) {
        const raw = row == null ? "" : row[key];
        return "\"" + String(raw == null ? "" : raw).replace(/"/g, "\"\"") + "\"";
      }).join(","));
    });
    const blob = new Blob([lines.join("\n")], { type: "text/csv;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  }

  function setActiveTab(tab) {
    if (!TAB_PANEL[tab]) tab = "tablero";
    state.activeTab = tab;
    document.querySelectorAll(".crm-tab-btn").forEach(function (btn) {
      const active = btn.getAttribute("data-tab") === tab;
      btn.classList.toggle("is-active", active);
      btn.classList.toggle("secondary", !active);
    });
    Object.keys(TAB_PANEL).forEach(function (key) {
      const panel = $(TAB_PANEL[key]);
      if (panel) panel.classList.toggle("is-hidden", key !== tab);
    });
    try {
      const next = new URL(window.location.href);
      next.searchParams.set("tab", tab);
      window.history.replaceState({}, "", next.toString());
    } catch (_) {}
  }

  function leadName(lead) {
    return normalize(lead.nombre) || normalize(lead.empresa_origen) || normalize(lead.codigo) || ("Lead #" + (lead.id || ""));
  }

  function appendLeadOptions(selectedValue) {
    const options = ["<option value=\"\">Sin lead asociado</option>"];
    state.leads.forEach(function (lead) {
      const id = Number(lead.id || 0);
      if (id <= 0) return;
      options.push("<option value=\"" + id + "\">" + esc((normalize(lead.codigo) || ("LEAD-" + id)) + " - " + leadName(lead)) + "</option>");
    });
    ["interaccionLeadId", "leadConvertir"].forEach(function (id) {
      const select = $(id);
      if (!select) return;
      select.innerHTML = options.join("");
      select.value = selectedValue ? String(selectedValue) : "";
    });
  }

  function computeAgenda() {
    const items = [];
    const now = Date.now();
    state.leads.forEach(function (lead) {
      const due = normalize(lead.proximo_contacto);
      if (!due) return;
      const time = new Date(due.replace(" ", "T")).getTime();
      items.push({
        tipo: "Lead",
        prioridad: Number.isNaN(time) || time >= now ? 1 : 0,
        titulo: leadName(lead),
        detalle: "Proximo contacto",
        fecha: due,
        estado: normalize(lead.estado_lead) || "nuevo",
        valor: asNumber(lead.valor_potencial)
      });
    });
    state.interacciones.forEach(function (item) {
      const next = normalize(item.proxima_accion);
      if (!next) return;
      items.push({
        tipo: "Seguimiento",
        prioridad: 1,
        titulo: normalize(item.resumen) || normalize(item.codigo) || ("Seguimiento #" + item.id),
        detalle: normalize(item.resultado) || "Accion pendiente",
        fecha: next,
        estado: normalize(item.estado_interaccion) || "abierta",
        valor: 0
      });
    });
    return items.sort(function (a, b) {
      if (a.prioridad !== b.prioridad) return a.prioridad - b.prioridad;
      return normalize(a.fecha).localeCompare(normalize(b.fecha));
    }).slice(0, 10);
  }

  function renderMiniList(id, items, emptyText) {
    const el = $(id);
    if (!el) return;
    if (!items || !items.length) {
      el.innerHTML = "<p class=\"form-help\">" + esc(emptyText || "Sin registros.") + "</p>";
      return;
    }
    el.innerHTML = items.map(function (item) {
      return "<div class=\"crm-mini-item\">" + item + "</div>";
    }).join("");
  }

  function renderKPIs() {
    const adv = state.advanced || {};
    const activeLeads = asNumber(adv.leads_activos) || state.leads.filter(function (x) {
      return ["ganado", "perdido", "descalificado", "cerrado"].indexOf(normalize(x.estado_lead).toLowerCase()) < 0;
    }).length;
    const pipeline = asNumber(adv.valor_pipeline) || state.leads.reduce(function (sum, x) { return sum + asNumber(x.valor_potencial); }, 0);
    const forecast = asNumber(adv.forecast_ponderado) || state.leads.reduce(function (sum, x) { return sum + asNumber(x.valor_potencial) * (asNumber(x.probabilidad) / 100); }, 0);
    const meta = asNumber(adv.meta_valor);
    const conversion = asNumber(adv.conversion_pct);
    const health = asNumber(adv.salud_comercial_pct);
    const riskValue = asNumber(adv.valor_riesgo);
    const kpis = [
      ["Salud CRM", formatNumber(health, 0) + "%", health >= 80 ? "Operacion comercial controlada." : "Revisa alertas y acciones priorizadas."],
      ["Leads activos", formatNumber(activeLeads), "Oportunidades abiertas en seguimiento."],
      ["Pipeline", formatMoney(pipeline), "Valor bruto del embudo comercial."],
      ["Forecast", formatMoney(forecast), "Pipeline ponderado por probabilidad."],
      ["Valor en riesgo", formatMoney(riskValue), "Oportunidades estancadas o sin avance reciente."],
      ["Meta del periodo", formatMoney(meta), meta > 0 ? formatNumber(adv.cumplimiento_meta_pct, 0) + "% de cumplimiento esperado." : "Configura metas por asesor o canal."],
      ["Cotizaciones abiertas", formatNumber(adv.cotizaciones_abiertas || state.cotizaciones.length), formatMoney(adv.cotizaciones_valor)],
      ["Conversion ganada", formatNumber(conversion, 0) + "%", "Ganados frente a perdidos."],
      ["Agenda hoy", formatNumber(adv.agenda_hoy), "Acciones comerciales para hoy."],
      ["Campanas activas", formatNumber(adv.campanas_activas || state.campanas.filter(function (x) { return normalize(x.estado_campana) === "activa"; }).length), "Fuentes de demanda vigentes."]
    ];
    const el = $("crmKpiGrid");
    if (!el) return;
    el.innerHTML = kpis.map(function (item) {
      return "<article class=\"crm-kpi-card\"><span>" + esc(item[0]) + "</span><strong>" + esc(item[1]) + "</strong><small>" + esc(item[2]) + "</small></article>";
    }).join("");
  }

  function renderDashboard() {
    const adv = state.advanced || {};
    const maxValue = Math.max.apply(null, [1].concat((adv.embudo || []).map(function (x) { return asNumber(x.valor); })));
    renderMiniList("crmPipelineStages", (adv.embudo || []).map(function (item) {
      const pct = Math.max(3, Math.min(100, (asNumber(item.valor) / maxValue) * 100));
      return "<strong>" + esc(item.estado || "sin_estado") + "</strong><br><small>" + formatNumber(item.leads) + " leads | " + formatMoney(item.valor) + " | forecast " + formatMoney(item.forecast) + "</small><div class=\"crm-progress\"><span style=\"width:" + pct + "%\"></span></div>";
    }), "Todavia no hay etapas con valor comercial.");

    renderMiniList("crmAlerts", (adv.alertas || []).map(function (text) {
      return "<strong>Atencion</strong><br><small>" + esc(text) + "</small>";
    }), "Sin alertas criticas del CRM.");

    const agenda = safeArray(adv.agenda).length ? adv.agenda : computeAgenda();
    renderMiniList("crmAgendaDashboard", agenda.slice(0, 8).map(function (item) {
      return "<strong>" + esc(normalize(item.tipo) + ": " + (normalize(item.nombre) || normalize(item.titulo) || normalize(item.referencia))) + "</strong><br><small>" + esc(formatDateTime(item.fecha)) + " | " + esc(normalize(item.responsable) || normalize(item.estado) || "-") + "</small>";
    }), "No hay acciones programadas.");

    renderMiniList("crmActionPlan", safeArray(adv.acciones_prioritarias).map(function (item) {
      return "<strong>" + esc(normalize(item.titulo) || "Accion comercial") + "</strong><br><small>" + esc(normalize(item.detalle) || "-") + "</small><br><span class=\"" + badgeClassForStatus(normalize(item.severidad) === "alta" ? "perdido" : (normalize(item.severidad) === "media" ? "propuesta" : "activa")) + "\">" + esc(normalize(item.severidad) || "baja") + "</span> <small>" + esc(normalize(item.accion) || "-") + (asNumber(item.valor) > 0 ? " | " + formatMoney(item.valor) : "") + "</small>";
    }), "Sin acciones criticas para el periodo.");

    renderMiniList("crmResponsablesWrap", safeArray(adv.responsables).map(function (item) {
      return "<strong>" + esc(normalize(item.responsable) || "Sin asignar") + "</strong><br><small>" + formatNumber(item.leads_activos) + " leads | " + formatMoney(item.forecast_ponderado) + " forecast | " + formatNumber(item.probabilidad_promedio) + "% prob.</small>";
    }), "Sin responsables con pipeline activo.");

    renderMiniList("crmCanalesWrap", safeArray(adv.canales).map(function (item) {
      return "<strong>" + esc(normalize(item.canal) || "Sin canal") + "</strong><br><small>" + formatNumber(item.leads) + " leads | " + formatMoney(item.forecast_ponderado) + " forecast | conversion " + formatNumber(item.conversion_pct) + "%</small>";
    }), "Sin canales comerciales registrados.");

    renderScoresTable("crmTopLeadsDashboard", safeArray(adv.top_leads), 8);
  }

  function renderScoresTable(id, rows, limit) {
    const el = $(id);
    if (!el) return;
    const list = safeArray(rows).slice(0, limit || 50);
    if (!list.length) {
      el.innerHTML = "<p class=\"form-help\">No hay scoring disponible.</p>";
      return;
    }
    el.innerHTML = "<table class=\"table\"><thead><tr><th>Lead</th><th>Estado</th><th>Valor</th><th>Prob.</th><th>Score</th><th>Recomendacion</th><th>Proximo contacto</th></tr></thead><tbody>" +
      list.map(function (item) {
        return "<tr><td><strong>" + esc(normalize(item.codigo) || ("#" + item.id)) + "</strong><br><small>" + esc(normalize(item.nombre) || normalize(item.empresa_origen) || "-") + "</small></td>" +
          "<td>" + statusBadge(item.estado_lead) + "</td><td>" + formatMoney(item.valor_potencial) + "</td><td>" + formatNumber(item.probabilidad) + "%</td><td><strong>" + formatNumber(item.score) + "</strong></td><td>" + esc(item.recomendacion || "-") + "</td><td>" + esc(formatDateTime(item.proximo_contacto)) + "</td></tr>";
      }).join("") +
      "</tbody></table>";
  }

  function renderLeads() {
    const pill = $("leadCounterPill");
    if (pill) pill.textContent = state.leads.length + " leads";
    appendLeadOptions();
    const el = $("leadsTableWrap");
    if (!el) return;
    if (!state.leads.length) {
      el.innerHTML = "<p class=\"form-help\">No hay leads con los filtros actuales.</p>";
      return;
    }
    el.innerHTML = "<table class=\"table\"><thead><tr><th>Prospecto</th><th>Estado</th><th>Canal</th><th>Valor</th><th>Prob.</th><th>Contacto</th><th>Responsable</th><th>Acciones</th></tr></thead><tbody>" +
      state.leads.map(function (item) {
        return "<tr><td><strong>" + esc(leadName(item)) + "</strong><br><small>" + esc(normalize(item.codigo) || ("#" + item.id)) + " | " + esc(normalize(item.email) || normalize(item.telefono) || "-") + "</small></td>" +
          "<td>" + statusBadge(item.estado_lead) + "</td><td>" + esc(normalize(item.canal_origen) || "-") + "</td><td>" + formatMoney(item.valor_potencial) + "</td><td>" + formatNumber(item.probabilidad) + "%</td><td>" + esc(formatDateTime(item.proximo_contacto)) + "</td><td>" + esc(normalize(item.propietario) || "-") + "</td>" +
          "<td class=\"crm-actions-cell\"><button class=\"btn secondary small\" type=\"button\" data-edit-lead=\"" + Number(item.id || 0) + "\">Editar</button><button class=\"btn secondary small\" type=\"button\" data-follow-lead=\"" + Number(item.id || 0) + "\">Seguimiento</button><button class=\"btn secondary small\" type=\"button\" data-quote-lead=\"" + Number(item.id || 0) + "\">Cotizar</button>" + transitionSelect("lead", item.id, LEAD_TRANSITIONS, item.estado_lead) + "</td></tr>";
      }).join("") +
      "</tbody></table>";
  }

  function renderInteracciones() {
    const pill = $("interaccionCounterPill");
    if (pill) pill.textContent = state.interacciones.length + " seguimientos";
    const el = $("interaccionesTableWrap");
    if (!el) return;
    if (!state.interacciones.length) {
      el.innerHTML = "<p class=\"form-help\">No hay seguimientos registrados.</p>";
      return;
    }
    el.innerHTML = "<table class=\"table\"><thead><tr><th>Seguimiento</th><th>Estado</th><th>Tipo</th><th>Fecha</th><th>Responsable</th><th>Resultado</th><th>Acciones</th></tr></thead><tbody>" +
      state.interacciones.map(function (item) {
        return "<tr><td><strong>" + esc(normalize(item.codigo) || ("#" + item.id)) + "</strong><br><small>" + esc(normalize(item.resumen) || "-") + "</small></td><td>" + statusBadge(item.estado_interaccion) + "</td><td>" + esc(normalize(item.tipo_interaccion) || "-") + "</td><td>" + esc(formatDateTime(item.fecha_interaccion)) + "</td><td>" + esc(normalize(item.usuario_responsable) || "-") + "</td><td>" + esc(normalize(item.resultado) || "-") + "</td><td class=\"crm-actions-cell\"><button class=\"btn secondary small\" type=\"button\" data-edit-interaccion=\"" + Number(item.id || 0) + "\">Editar</button>" + transitionSelect("interaccion", item.id, INTERACTION_TRANSITIONS, item.estado_interaccion) + "</td></tr>";
      }).join("") +
      "</tbody></table>";
  }

  function renderCotizaciones() {
    const pill = $("cotizacionCounterPill");
    if (pill) pill.textContent = state.cotizaciones.length + " cotizaciones";
    const el = $("cotizacionesTableWrap");
    if (!el) return;
    if (!state.cotizaciones.length) {
      el.innerHTML = "<p class=\"form-help\">No hay cotizaciones comerciales.</p>";
      return;
    }
    el.innerHTML = "<table class=\"table\"><thead><tr><th>Cotizacion</th><th>Estado</th><th>Cliente</th><th>Fecha</th><th>Vigencia</th><th>Total</th><th>Acciones</th></tr></thead><tbody>" +
      state.cotizaciones.map(function (item) {
        const id = Number(item.id || 0);
        return "<tr><td><strong>" + esc(normalize(item.codigo) || ("#" + id)) + "</strong><br><small>" + esc(normalize(item.origen) || "-") + "</small></td><td>" + statusBadge(item.estado_documento) + "</td><td>" + esc(normalize(item.cliente_nombre) || "-") + "</td><td>" + esc(formatDate(item.fecha_documento)) + "</td><td>" + esc(formatDate(item.vigencia_hasta)) + "</td><td>" + formatMoney(item.total) + "</td><td class=\"crm-actions-cell\"><button class=\"btn secondary small\" type=\"button\" data-edit-cotizacion=\"" + id + "\">Editar</button><button class=\"btn secondary small\" type=\"button\" data-convert-quote=\"convertir_pedido\" data-quote-id=\"" + id + "\">Pedido</button><button class=\"btn secondary small\" type=\"button\" data-convert-quote=\"convertir_documento_final\" data-quote-id=\"" + id + "\">Factura</button>" + transitionSelect("cotizacion", id, QUOTE_TRANSITIONS, item.estado_documento) + "</td></tr>";
      }).join("") +
      "</tbody></table>";
  }

  function renderCampanas() {
    const pill = $("campanaCounterPill");
    if (pill) pill.textContent = state.campanas.length + " campanas";
    const el = $("campanasTableWrap");
    if (!el) return;
    if (!state.campanas.length) {
      el.innerHTML = "<p class=\"form-help\">No hay campanas comerciales.</p>";
      return;
    }
    el.innerHTML = "<table class=\"table\"><thead><tr><th>Campana</th><th>Estado</th><th>Canal</th><th>Fechas</th><th>Presupuesto</th><th>KPI</th><th>Acciones</th></tr></thead><tbody>" +
      state.campanas.map(function (item) {
        return "<tr><td><strong>" + esc(normalize(item.nombre) || normalize(item.codigo) || ("#" + item.id)) + "</strong><br><small>" + esc(normalize(item.objetivo) || "-") + "</small></td><td>" + statusBadge(item.estado_campana) + "</td><td>" + esc(normalize(item.canal) || "-") + "</td><td>" + esc(formatDate(item.fecha_inicio)) + " - " + esc(formatDate(item.fecha_fin)) + "</td><td>" + formatMoney(item.presupuesto) + "</td><td>" + esc(normalize(item.kpi_objetivo) || "-") + "</td><td class=\"crm-actions-cell\"><button class=\"btn secondary small\" type=\"button\" data-edit-campana=\"" + Number(item.id || 0) + "\">Editar</button>" + transitionSelect("campana", item.id, CAMPAIGN_TRANSITIONS, item.estado_campana) + "</td></tr>";
      }).join("") +
      "</tbody></table>";
  }

  function renderForecast() {
    const adv = state.advanced || {};
    const maxValue = Math.max.apply(null, [1].concat((adv.embudo || []).map(function (x) { return asNumber(x.forecast); })));
    renderMiniList("forecastEmbudoWrap", (adv.embudo || []).map(function (item) {
      const pct = Math.max(3, Math.min(100, (asNumber(item.forecast) / maxValue) * 100));
      return "<strong>" + esc(item.estado || "sin_estado") + "</strong><br><small>" + formatNumber(item.leads) + " leads | prob. " + formatNumber(item.probabilidad_promedio) + "% | " + formatMoney(item.forecast) + "</small><div class=\"crm-progress\"><span style=\"width:" + pct + "%\"></span></div>";
    }), "Sin forecast calculado.");

    const agenda = safeArray(adv.agenda).length ? safeArray(adv.agenda) : computeAgenda();
    renderMiniList("forecastAgendaWrap", agenda.slice(0, 10).map(function (item) {
      return "<strong>" + esc(normalize(item.nombre) || normalize(item.titulo) || normalize(item.referencia) || "-") + "</strong><br><small>" + esc(formatDateTime(item.fecha)) + " | " + esc(normalize(item.estado) || "-") + "</small>";
    }), "Sin agenda comercial.");

    renderMiniList("forecastAlertsWrap", safeArray(adv.alertas).map(function (text) {
      return "<strong>Alerta</strong><br><small>" + esc(text) + "</small>";
    }), "Forecast sin alertas.");

    renderScoresTable("forecastScoresWrap", safeArray(adv.top_leads), 30);
  }

  function renderMetas() {
    const metas = safeArray(state.advanced.metas);
    const el = $("metasTableWrap");
    if (!el) return;
    if (!metas.length) {
      el.innerHTML = "<p class=\"form-help\">No hay metas configuradas para este periodo.</p>";
      return;
    }
    el.innerHTML = "<table class=\"table\"><thead><tr><th>Periodo</th><th>Propietario</th><th>Canal</th><th>Meta valor</th><th>Leads</th><th>Conversion</th><th>Estado</th></tr></thead><tbody>" +
      metas.map(function (item) {
        return "<tr><td>" + esc(item.periodo || "-") + "</td><td>" + esc(item.propietario || "Equipo") + "</td><td>" + esc(item.canal || "todos") + "</td><td>" + formatMoney(item.meta_valor) + "</td><td>" + formatNumber(item.meta_leads) + "</td><td>" + formatNumber(item.meta_conversion_pct) + "%</td><td>" + statusBadge(item.estado) + "</td></tr>";
      }).join("") +
      "</tbody></table>";
  }

  function renderEmbudo() {
    const summary = state.embudo && state.embudo.summary ? state.embudo.summary : {};
    const items = safeArray(state.embudo && state.embudo.items);
    const alertas = safeArray(state.embudo && state.embudo.alertas);
    const conversion = asNumber(summary.conversion_pct);
    const badge = $("embudoConversionBadge");
    if (badge) badge.textContent = formatNumber(conversion, 0) + "%";
    const alertBadge = $("embudoAlertBadge");
    if (alertBadge) alertBadge.textContent = String(alertas.length);

    renderMiniList("embudoResumen", [
      "<strong>Cotizaciones</strong><br><small>" + formatNumber(summary.cotizaciones || state.cotizaciones.length) + " documentos en seguimiento.</small>",
      "<strong>Valor convertido</strong><br><small>" + formatMoney(summary.valor_convertido || 0) + "</small>",
      "<strong>Tiempo promedio</strong><br><small>" + formatNumber(summary.horas_promedio_conversion || 0, 1) + " horas</small>"
    ], "Sin resumen de embudo.");

    renderMiniList("embudoAlertas", alertas.map(function (alerta) {
      return "<strong>" + esc(normalize(alerta.cotizacion_codigo) || ("Cotizacion #" + alerta.cotizacion_id)) + "</strong><br><small>" + esc(normalize(alerta.alerta) || "Alerta comercial") + "</small>";
    }), "Sin alertas de SLA o vigencia.");

    const agenda = computeAgenda();
    renderMiniList("crmAgendaWrap", agenda.map(function (entry) {
      return "<strong>" + esc(entry.tipo + ": " + entry.titulo) + "</strong><br><small>" + esc(entry.detalle) + " | " + esc(formatDateTime(entry.fecha)) + "</small>";
    }), "No hay proximas acciones registradas.");

    const table = $("embudoTableWrap");
    if (!table) return;
    if (!items.length) {
      table.innerHTML = "<p class=\"form-help\">No hay trazabilidad documental disponible.</p>";
      return;
    }
    table.innerHTML = "<table class=\"table\"><thead><tr><th>Cotizacion</th><th>Estado cot.</th><th>Pedido</th><th>Estado ped.</th><th>Documento final</th><th>Etapa</th><th>Tiempo</th><th>Alerta</th></tr></thead><tbody>" +
      items.map(function (item) {
        return "<tr><td>" + esc(normalize(item.cotizacion_codigo) || "-") + "</td><td>" + esc(normalize(item.estado_cotizacion) || "-") + "</td><td>" + esc(normalize(item.pedido_codigo) || "-") + "</td><td>" + esc(normalize(item.estado_pedido) || "-") + "</td><td>" + esc(normalize(item.documento_final_codigo) || "-") + "</td><td>" + esc(normalize(item.conversion_etapa) || "-") + "</td><td>" + formatNumber(item.horas_desde_cotizacion) + " h</td><td>" + esc(normalize(item.alerta) || "-") + "</td></tr>";
      }).join("") +
      "</tbody></table>";
  }

  function renderAll() {
    renderKPIs();
    renderDashboard();
    renderLeads();
    renderInteracciones();
    renderCotizaciones();
    renderCampanas();
    renderForecast();
    renderMetas();
    renderEmbudo();
    bindDynamicActions();
  }

  function fillLeadForm(item) {
    item = item || {};
    setValue("leadId", item.id || "");
    setValue("leadCodigo", item.codigo || "");
    setValue("leadEstado", normalize(item.estado_lead) || "nuevo");
    setValue("leadNombre", item.nombre || "");
    setValue("leadEmpresaOrigen", item.empresa_origen || "");
    setValue("leadEmail", item.email || "");
    setValue("leadTelefono", item.telefono || "");
    setValue("leadCanal", item.canal_origen || "");
    setValue("leadPropietario", item.propietario || "");
    setValue("leadValor", item.valor_potencial || "");
    setValue("leadProbabilidad", item.probabilidad || "");
    setValue("leadProximoContacto", toDateTimeInputValue(item.proximo_contacto));
    setValue("leadNotas", item.notas || "");
    setValue("leadObservaciones", item.observaciones || "");
    if ($("leadSaveBtn")) $("leadSaveBtn").textContent = item.id ? "Actualizar lead" : "Guardar lead";
    setMessage("leadMsg", "", "");
  }

  function fillInteraccionForm(item) {
    item = item || {};
    setValue("interaccionId", item.id || "");
    setValue("interaccionCodigo", item.codigo || "");
    setValue("interaccionEstado", normalize(item.estado_interaccion) || "abierta");
    appendLeadOptions(item.lead_id || "");
    setValue("interaccionClienteId", item.cliente_id || "");
    setValue("interaccionTipo", normalize(item.tipo_interaccion) || "seguimiento");
    setValue("interaccionFecha", item.id ? toDateTimeInputValue(item.fecha_interaccion) : toDateTimeInputValue(new Date().toISOString()));
    setValue("interaccionResponsable", item.usuario_responsable || "");
    setValue("interaccionResultado", item.resultado || "");
    setValue("interaccionResumen", item.resumen || "");
    setValue("interaccionProximaAccion", item.proxima_accion || "");
    setValue("interaccionObservaciones", item.observaciones || "");
    if ($("interaccionSaveBtn")) $("interaccionSaveBtn").textContent = item.id ? "Actualizar seguimiento" : "Guardar seguimiento";
    setMessage("interaccionMsg", "", "");
  }

  function fillCotizacionForm(item) {
    item = item || {};
    setValue("cotizacionId", item.id || "");
    setValue("cotizacionCodigo", item.codigo || "");
    setValue("cotizacionEstado", normalize(item.estado_documento) || "borrador");
    setValue("cotizacionClienteNombre", item.cliente_nombre || "");
    setValue("cotizacionClienteId", item.cliente_id || "");
    setValue("cotizacionFecha", item.id ? toDateInputValue(item.fecha_documento) : todayInput());
    setValue("cotizacionVigencia", toDateInputValue(item.vigencia_hasta));
    setValue("cotizacionSubtotal", item.subtotal || "");
    setValue("cotizacionDescuento", item.descuento_total || "");
    setValue("cotizacionImpuesto", item.impuesto_total || "");
    setValue("cotizacionTotal", item.total || "");
    setValue("cotizacionMoneda", normalize(item.moneda) || "COP");
    setValue("cotizacionOrigen", item.origen || "");
    setValue("cotizacionNotas", item.notas || "");
    setValue("cotizacionObservaciones", item.observaciones || "");
    if ($("cotizacionSaveBtn")) $("cotizacionSaveBtn").textContent = item.id ? "Actualizar cotizacion" : "Guardar cotizacion";
    setMessage("cotizacionMsg", "", "");
  }

  function fillCampanaForm(item) {
    item = item || {};
    setValue("campanaId", item.id || "");
    setValue("campanaCodigo", item.codigo || "");
    setValue("campanaEstado", normalize(item.estado_campana) || "planificada");
    setValue("campanaNombre", item.nombre || "");
    setValue("campanaCanal", normalize(item.canal) || "email");
    setValue("campanaObjetivo", item.objetivo || "");
    setValue("campanaPresupuesto", item.presupuesto || "");
    setValue("campanaAudiencia", item.audiencia || "");
    setValue("campanaFechaInicio", toDateInputValue(item.fecha_inicio));
    setValue("campanaFechaFin", toDateInputValue(item.fecha_fin));
    setValue("campanaKPI", item.kpi_objetivo || "");
    setValue("campanaResultados", item.resultado_json || "");
    setValue("campanaObservaciones", item.observaciones || "");
    if ($("campanaSaveBtn")) $("campanaSaveBtn").textContent = item.id ? "Actualizar campana" : "Guardar campana";
    setMessage("campanaMsg", "", "");
  }

  function prepareInteractionFromLead(id) {
    const lead = state.leads.find(function (item) { return Number(item.id || 0) === id; }) || {};
    fillInteraccionForm({});
    appendLeadOptions(id);
    setValue("interaccionTipo", "seguimiento");
    setValue("interaccionResumen", "Seguimiento comercial de " + leadName(lead));
    setValue("interaccionResponsable", lead.propietario || "");
    setActiveTab("interacciones");
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  function prepareQuoteFromLead(id) {
    const lead = state.leads.find(function (item) { return Number(item.id || 0) === id; }) || {};
    fillCotizacionForm({});
    setValue("cotizacionClienteNombre", leadName(lead));
    setValue("cotizacionOrigen", normalize(lead.codigo) || "crm");
    setValue("cotizacionNotas", lead.notas || "");
    setValue("cotizacionObservaciones", "Preparada desde lead " + (normalize(lead.codigo) || ("#" + id)));
    setActiveTab("cotizaciones");
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  async function transitionRecord(kind, id, targetState) {
    const map = {
      lead: ["/api/empresa/crm/leads?action=transicionar", "estado_lead"],
      interaccion: ["/api/empresa/crm/interacciones?action=transicionar", "estado_interaccion"],
      cotizacion: ["/api/empresa/ventas/cotizaciones?action=transicionar", "estado_documento"],
      campana: ["/api/empresa/crm/campanas?action=transicionar", "estado_campana"]
    };
    const cfg = map[kind];
    if (!cfg || !id || !targetState) return;
    const payload = { id: id, nuevo_estado: targetState };
    payload[cfg[1]] = targetState;
    await postJSON(cfg[0], payload);
    await loadAllData();
  }

  async function convertLeadToQuote() {
    const leadID = Number(($("leadConvertir") || {}).value || 0);
    if (leadID <= 0) {
      setMessage("crmMsg", "Selecciona un lead para convertir.", "error");
      return;
    }
    await postJSON("/api/empresa/crm_avanzado?action=cotizacion_desde_lead", {
      action: "cotizacion_desde_lead",
      lead_id: leadID,
      codigo: normalize(($("cotCodigo") || {}).value)
    });
    setValue("cotCodigo", "");
    setMessage("crmMsg", "Cotizacion generada desde el lead.", "success");
    await loadAllData();
    setActiveTab("cotizaciones");
  }

  async function convertQuoteAction(id, action) {
    await postJSON("/api/empresa/ventas/cotizaciones?action=" + encodeURIComponent(action), { id: id });
    setMessage("crmMsg", action === "convertir_pedido" ? "Cotizacion convertida a pedido." : "Documento final generado desde cotizacion.", "success");
    await loadAllData();
    setActiveTab("cotizaciones");
  }

  async function seedDemo() {
    await postJSON("/api/empresa/crm_avanzado?action=seed_demo", { action: "seed_demo" });
    setMessage("crmMsg", "Datos demo creados para validar el CRM.", "success");
    await loadAllData();
  }

  async function loadAllData() {
    if (!state.empresaID) {
      setMessage("crmMsg", "Selecciona una empresa para usar el CRM.", "error");
      return;
    }
    const periodoInput = $("crmPeriodo");
    state.periodo = normalize(periodoInput && periodoInput.value) || currentMonth();
    if (periodoInput && !periodoInput.value) periodoInput.value = state.periodo;
    if ($("metaPeriodo") && !$("metaPeriodo").value) $("metaPeriodo").value = state.periodo;

    const results = await Promise.all([
      safeFetch(buildURL("/api/empresa/crm/leads"), []),
      safeFetch(buildURL("/api/empresa/crm/interacciones"), []),
      safeFetch(buildURL("/api/empresa/crm/campanas"), []),
      safeFetch(buildURL("/api/empresa/ventas/cotizaciones"), []),
      safeFetch(buildURL("/api/empresa/ventas/cotizaciones", { action: "embudo", limit: 40 }), { summary: {}, items: [], alertas: [] }),
      safeFetch(buildURL("/api/empresa/crm_avanzado", { action: "dashboard", periodo: state.periodo }), { embudo: [], agenda: [], top_leads: [], metas: [], alertas: [], responsables: [], canales: [], acciones_prioritarias: [] })
    ]);

    state.leads = safeArray(results[0]);
    state.interacciones = safeArray(results[1]);
    state.campanas = safeArray(results[2]);
    state.cotizaciones = safeArray(results[3]);
    state.embudo = results[4] || { summary: {}, items: [], alertas: [] };
    state.advanced = results[5] || { embudo: [], agenda: [], top_leads: [], metas: [], alertas: [], responsables: [], canales: [], acciones_prioritarias: [] };
    renderAll();
  }

  function bindDynamicActions() {
    document.querySelectorAll("[data-edit-lead]").forEach(function (btn) {
      btn.onclick = function () {
        const id = Number(btn.getAttribute("data-edit-lead") || 0);
        fillLeadForm(state.leads.find(function (item) { return Number(item.id || 0) === id; }) || {});
        setActiveTab("leads");
        window.scrollTo({ top: 0, behavior: "smooth" });
      };
    });
    document.querySelectorAll("[data-follow-lead]").forEach(function (btn) {
      btn.onclick = function () { prepareInteractionFromLead(Number(btn.getAttribute("data-follow-lead") || 0)); };
    });
    document.querySelectorAll("[data-quote-lead]").forEach(function (btn) {
      btn.onclick = function () { prepareQuoteFromLead(Number(btn.getAttribute("data-quote-lead") || 0)); };
    });
    document.querySelectorAll("[data-edit-interaccion]").forEach(function (btn) {
      btn.onclick = function () {
        const id = Number(btn.getAttribute("data-edit-interaccion") || 0);
        fillInteraccionForm(state.interacciones.find(function (item) { return Number(item.id || 0) === id; }) || {});
        setActiveTab("interacciones");
        window.scrollTo({ top: 0, behavior: "smooth" });
      };
    });
    document.querySelectorAll("[data-edit-cotizacion]").forEach(function (btn) {
      btn.onclick = function () {
        const id = Number(btn.getAttribute("data-edit-cotizacion") || 0);
        fillCotizacionForm(state.cotizaciones.find(function (item) { return Number(item.id || 0) === id; }) || {});
        setActiveTab("cotizaciones");
        window.scrollTo({ top: 0, behavior: "smooth" });
      };
    });
    document.querySelectorAll("[data-edit-campana]").forEach(function (btn) {
      btn.onclick = function () {
        const id = Number(btn.getAttribute("data-edit-campana") || 0);
        fillCampanaForm(state.campanas.find(function (item) { return Number(item.id || 0) === id; }) || {});
        setActiveTab("campanas");
        window.scrollTo({ top: 0, behavior: "smooth" });
      };
    });
    document.querySelectorAll("[data-convert-quote]").forEach(function (btn) {
      btn.onclick = async function () {
        try {
          await convertQuoteAction(Number(btn.getAttribute("data-quote-id") || 0), btn.getAttribute("data-convert-quote"));
        } catch (err) {
          setMessage("crmMsg", err.message || "No se pudo convertir la cotizacion.", "error");
        }
      };
    });
    document.querySelectorAll("[data-transition-kind]").forEach(function (select) {
      select.onchange = async function () {
        const target = normalize(select.value);
        if (!target) return;
        try {
          await transitionRecord(select.getAttribute("data-transition-kind"), Number(select.getAttribute("data-transition-id") || 0), target);
          setMessage("crmMsg", "Estado actualizado.", "success");
        } catch (err) {
          setMessage("crmMsg", err.message || "No se pudo cambiar el estado.", "error");
          select.value = "";
        }
      };
    });
  }

  function bindEvents() {
    document.querySelectorAll(".crm-tab-btn").forEach(function (btn) {
      btn.addEventListener("click", function () { setActiveTab(btn.getAttribute("data-tab")); });
    });

    $("crmSearchBtn").addEventListener("click", async function () {
      state.q = normalize($("crmSearch").value);
      state.includeInactive = !!$("crmIncludeInactive").checked;
      await loadAllData();
      setMessage("crmMsg", "Filtros aplicados.", "success");
    });
    $("crmRefreshBtn").addEventListener("click", async function () {
      await loadAllData();
      setMessage("crmMsg", "CRM actualizado.", "success");
    });
    $("crmSeedBtn").addEventListener("click", async function () {
      try {
        await seedDemo();
      } catch (err) {
        setMessage("crmMsg", err.message || "No se pudo crear el demo.", "error");
      }
    });
    $("crmExportBtn").addEventListener("click", function () {
      try {
        const map = {
          leads: ["crm_leads.csv", state.leads],
          interacciones: ["crm_seguimientos.csv", state.interacciones],
          cotizaciones: ["crm_cotizaciones.csv", state.cotizaciones],
          campanas: ["crm_campanas.csv", state.campanas],
          forecast: ["crm_forecast.csv", safeArray(state.advanced.top_leads)],
          metas: ["crm_metas.csv", safeArray(state.advanced.metas)],
          embudo: ["crm_embudo.csv", safeArray(state.embudo.items)]
        };
        const entry = map[state.activeTab] || ["crm_tablero.csv", safeArray(state.advanced.top_leads)];
        exportRowsToCSV(entry[0], entry[1]);
        setMessage("crmMsg", "Exportacion generada.", "success");
      } catch (err) {
        setMessage("crmMsg", err.message || "No se pudo exportar.", "error");
      }
    });

    $("crmPeriodo").addEventListener("change", async function () {
      state.periodo = normalize($("crmPeriodo").value) || currentMonth();
      setValue("metaPeriodo", state.periodo);
      await loadAllData();
    });

    $("leadCancelBtn").addEventListener("click", function () { fillLeadForm({}); });
    $("interaccionCancelBtn").addEventListener("click", function () { fillInteraccionForm({}); });
    $("cotizacionCancelBtn").addEventListener("click", function () { fillCotizacionForm({}); });
    $("campanaCancelBtn").addEventListener("click", function () { fillCampanaForm({}); });

    $("leadForm").addEventListener("submit", async function (event) {
      event.preventDefault();
      const id = Number($("leadId").value || 0);
      const payload = {
        id: id,
        estado_lead: normalize($("leadEstado").value) || "nuevo",
        nombre: normalize($("leadNombre").value),
        empresa_origen: normalize($("leadEmpresaOrigen").value),
        email: normalize($("leadEmail").value),
        telefono: normalize($("leadTelefono").value),
        canal_origen: normalize($("leadCanal").value),
        valor_potencial: asNumber($("leadValor").value),
        probabilidad: asNumber($("leadProbabilidad").value),
        propietario: normalize($("leadPropietario").value),
        proximo_contacto: normalize($("leadProximoContacto").value).replace("T", " "),
        notas: normalize($("leadNotas").value),
        observaciones: normalize($("leadObservaciones").value)
      };
      if (!payload.nombre) {
        setMessage("leadMsg", "El nombre del prospecto es obligatorio.", "error");
        return;
      }
      try {
        if (id > 0) await putJSON("/api/empresa/crm/leads", payload);
        else await postJSON("/api/empresa/crm/leads", payload);
        fillLeadForm({});
        await loadAllData();
        setMessage("leadMsg", "Lead guardado correctamente.", "success");
      } catch (err) {
        setMessage("leadMsg", err.message || "No se pudo guardar el lead.", "error");
      }
    });

    $("interaccionForm").addEventListener("submit", async function (event) {
      event.preventDefault();
      const id = Number($("interaccionId").value || 0);
      const payload = {
        id: id,
        lead_id: Number($("interaccionLeadId").value || 0),
        cliente_id: Number($("interaccionClienteId").value || 0),
        estado_interaccion: normalize($("interaccionEstado").value) || "abierta",
        tipo_interaccion: normalize($("interaccionTipo").value) || "seguimiento",
        fecha_interaccion: normalize($("interaccionFecha").value).replace("T", " "),
        usuario_responsable: normalize($("interaccionResponsable").value),
        resultado: normalize($("interaccionResultado").value),
        resumen: normalize($("interaccionResumen").value),
        proxima_accion: normalize($("interaccionProximaAccion").value),
        observaciones: normalize($("interaccionObservaciones").value)
      };
      if (!payload.resumen) {
        setMessage("interaccionMsg", "El resumen del seguimiento es obligatorio.", "error");
        return;
      }
      try {
        if (id > 0) await putJSON("/api/empresa/crm/interacciones", payload);
        else await postJSON("/api/empresa/crm/interacciones", payload);
        fillInteraccionForm({});
        await loadAllData();
        setMessage("interaccionMsg", "Seguimiento guardado correctamente.", "success");
      } catch (err) {
        setMessage("interaccionMsg", err.message || "No se pudo guardar el seguimiento.", "error");
      }
    });

    $("cotizacionForm").addEventListener("submit", async function (event) {
      event.preventDefault();
      const id = Number($("cotizacionId").value || 0);
      const payload = {
        id: id,
        codigo: normalize($("cotizacionCodigo").value),
        estado_documento: normalize($("cotizacionEstado").value) || "borrador",
        cliente_nombre: normalize($("cotizacionClienteNombre").value),
        cliente_id: Number($("cotizacionClienteId").value || 0),
        fecha_documento: normalize($("cotizacionFecha").value) || todayInput(),
        vigencia_hasta: normalize($("cotizacionVigencia").value),
        subtotal: asNumber($("cotizacionSubtotal").value),
        descuento_total: asNumber($("cotizacionDescuento").value),
        impuesto_total: asNumber($("cotizacionImpuesto").value),
        total: asNumber($("cotizacionTotal").value),
        moneda: normalize($("cotizacionMoneda").value) || "COP",
        origen: normalize($("cotizacionOrigen").value),
        notas: normalize($("cotizacionNotas").value),
        observaciones: normalize($("cotizacionObservaciones").value)
      };
      if (!payload.total) payload.total = Math.max(0, payload.subtotal - payload.descuento_total + payload.impuesto_total);
      if (!payload.cliente_nombre) {
        setMessage("cotizacionMsg", "El cliente es obligatorio.", "error");
        return;
      }
      try {
        if (id > 0) await putJSON("/api/empresa/ventas/cotizaciones", payload);
        else await postJSON("/api/empresa/ventas/cotizaciones", payload);
        fillCotizacionForm({});
        await loadAllData();
        setMessage("cotizacionMsg", "Cotizacion guardada correctamente.", "success");
      } catch (err) {
        setMessage("cotizacionMsg", err.message || "No se pudo guardar la cotizacion.", "error");
      }
    });

    $("campanaForm").addEventListener("submit", async function (event) {
      event.preventDefault();
      const id = Number($("campanaId").value || 0);
      const payload = {
        id: id,
        estado_campana: normalize($("campanaEstado").value) || "planificada",
        nombre: normalize($("campanaNombre").value),
        canal: normalize($("campanaCanal").value) || "email",
        objetivo: normalize($("campanaObjetivo").value),
        presupuesto: asNumber($("campanaPresupuesto").value),
        audiencia: normalize($("campanaAudiencia").value),
        fecha_inicio: normalize($("campanaFechaInicio").value),
        fecha_fin: normalize($("campanaFechaFin").value),
        kpi_objetivo: normalize($("campanaKPI").value),
        resultado_json: normalize($("campanaResultados").value),
        observaciones: normalize($("campanaObservaciones").value)
      };
      if (!payload.nombre) {
        setMessage("campanaMsg", "El nombre de la campana es obligatorio.", "error");
        return;
      }
      try {
        if (id > 0) await putJSON("/api/empresa/crm/campanas", payload);
        else await postJSON("/api/empresa/crm/campanas", payload);
        fillCampanaForm({});
        await loadAllData();
        setMessage("campanaMsg", "Campana guardada correctamente.", "success");
      } catch (err) {
        setMessage("campanaMsg", err.message || "No se pudo guardar la campana.", "error");
      }
    });

    $("metaForm").addEventListener("submit", async function (event) {
      event.preventDefault();
      try {
        await postJSON("/api/empresa/crm_avanzado?action=meta", {
          action: "meta",
          meta: {
            periodo: normalize($("metaPeriodo").value) || state.periodo || currentMonth(),
            propietario: normalize($("metaPropietario").value),
            canal: normalize($("metaCanal").value),
            meta_valor: asNumber($("metaValor").value),
            meta_leads: Number($("metaLeads").value || 0),
            meta_conversion_pct: asNumber($("metaConv").value),
            estado: "activo"
          }
        });
        setMessage("crmMsg", "Meta comercial guardada.", "success");
        await loadAllData();
      } catch (err) {
        setMessage("crmMsg", err.message || "No se pudo guardar la meta.", "error");
      }
    });

    $("btnConvertLead").addEventListener("click", async function () {
      try {
        await convertLeadToQuote();
      } catch (err) {
        setMessage("crmMsg", err.message || "No se pudo convertir el lead.", "error");
      }
    });
  }

  async function init() {
    state.empresaID = resolveEmpresaID();
    state.periodo = currentMonth();
    const periodInput = $("crmPeriodo");
    if (periodInput) periodInput.value = state.periodo;
    if ($("metaPeriodo")) $("metaPeriodo").value = state.periodo;
    bindEvents();
    fillLeadForm({});
    fillInteraccionForm({});
    fillCotizacionForm({});
    fillCampanaForm({});
    const initialTab = normalize(queryParam("tab")) || "tablero";
    setActiveTab(initialTab);
    try {
      await loadAllData();
    } catch (err) {
      setMessage("crmMsg", err.message || "No se pudo cargar el CRM.", "error");
    }
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
