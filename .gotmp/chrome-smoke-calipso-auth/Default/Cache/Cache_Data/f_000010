(function () {
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

  const state = {
    empresaID: 0,
    q: "",
    includeInactive: false,
    activeTab: "leads",
    leads: [],
    interacciones: [],
    campanas: [],
    cotizaciones: [],
    embudo: { summary: {}, items: [], alertas: [] }
  };

  function queryParam(name) {
    return new URLSearchParams(window.location.search).get(name);
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

  function formatMoney(value) {
    return new Intl.NumberFormat("es-CO", {
      style: "currency",
      currency: "COP",
      maximumFractionDigits: 0
    }).format(asNumber(value));
  }

  function formatNumber(value, digits) {
    return new Intl.NumberFormat("es-CO", {
      minimumFractionDigits: digits || 0,
      maximumFractionDigits: digits || 0
    }).format(asNumber(value));
  }

  function formatDateTime(value) {
    const raw = normalize(value);
    if (!raw) return "-";
    const normalized = raw.replace(" ", "T");
    const date = new Date(normalized);
    if (!Number.isNaN(date.getTime())) {
      return new Intl.DateTimeFormat("es-CO", { dateStyle: "medium", timeStyle: "short" }).format(date);
    }
    return raw;
  }

  function formatDate(value) {
    const raw = normalize(value);
    if (!raw) return "-";
    return raw.length >= 10 ? raw.slice(0, 10) : raw;
  }

  function toDateInputValue(value) {
    const raw = normalize(value);
    if (!raw) return "";
    return raw.length >= 10 ? raw.slice(0, 10) : raw;
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
      const hours = String(date.getHours()).padStart(2, "0");
      const minutes = String(date.getMinutes()).padStart(2, "0");
      return year + "-" + month + "-" + day + "T" + hours + ":" + minutes;
    }
    return normalized.length >= 16 ? normalized.slice(0, 16) : "";
  }

  function nowDateInput() {
    return new Date().toISOString().slice(0, 10);
  }

  function setMessage(id, text, kind) {
    const el = document.getElementById(id);
    if (!el) return;
    el.textContent = text || "";
    el.className = "form-help";
    if (kind === "error") el.classList.add("value-negative");
    if (kind === "success") el.classList.add("value-positive");
  }

  async function fetchJSON(url, options) {
    const res = await fetch(url, Object.assign({ credentials: "same-origin" }, options || {}));
    const contentType = String(res.headers.get("Content-Type") || "").toLowerCase();
    if (!res.ok) {
      const text = await res.text();
      throw new Error(text || ("HTTP " + res.status));
    }
    if (contentType.indexOf("application/json") >= 0) {
      return res.json();
    }
    const text = await res.text();
    return text ? JSON.parse(text) : {};
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

  function badgeClassForStatus(status) {
    const value = normalize(status).toLowerCase();
    if (["ganado", "aprobada", "activa", "cerrada", "convertida", "postventa", "finalizada"].indexOf(value) >= 0) return "crm-pill is-success";
    if (["perdido", "descalificado", "cancelada", "anulada", "rechazada", "vencida"].indexOf(value) >= 0) return "crm-pill is-danger";
    if (["negociacion", "propuesta", "contactado", "emitida", "pausada", "en_progreso"].indexOf(value) >= 0) return "crm-pill is-warning";
    return "crm-pill";
  }

  function exportRowsToCSV(filename, rows) {
    if (!Array.isArray(rows) || !rows.length) {
      throw new Error("No hay registros para exportar.");
    }
    const keys = Array.from(rows.reduce(function (acc, row) {
      Object.keys(row || {}).forEach(function (key) { acc.add(key); });
      return acc;
    }, new Set()));
    const lines = [keys.join(",")];
    rows.forEach(function (row) {
      lines.push(keys.map(function (key) {
        const raw = row == null ? "" : row[key];
        const value = raw == null ? "" : String(raw);
        return "\"" + value.replace(/"/g, "\"\"") + "\"";
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

  function capitalize(value) {
    const raw = String(value || "");
    return raw.charAt(0).toUpperCase() + raw.slice(1);
  }

  function setActiveTab(tab) {
    state.activeTab = tab;
    document.querySelectorAll(".crm-tab-btn").forEach(function (btn) {
      const isActive = btn.getAttribute("data-tab") === tab;
      btn.classList.toggle("is-active", isActive);
      btn.classList.toggle("secondary", !isActive);
    });
    document.querySelectorAll(".crm-panel").forEach(function (panel) {
      panel.classList.toggle("is-hidden", panel.id !== "crmPanel" + capitalize(tab));
    });
  }

  function allowedTransitions(map, current) {
    const key = normalize(current).toLowerCase();
    return Array.isArray(map[key]) ? map[key] : [];
  }

  function appendLeadOptions(selectedValue) {
    const select = document.getElementById("interaccionLeadId");
    const options = ["<option value=\"\">Sin lead asociado</option>"];
    state.leads.forEach(function (lead) {
      const id = Number(lead.id || 0);
      if (id <= 0) return;
      const label = (normalize(lead.codigo) || ("LEAD-" + id)) + " - " + (normalize(lead.nombre) || "Prospecto");
      options.push("<option value=\"" + id + "\">" + esc(label) + "</option>");
    });
    select.innerHTML = options.join("");
    select.value = selectedValue ? String(selectedValue) : "";
  }

  function computeAgenda() {
    const items = [];
    const now = Date.now();
    state.leads.forEach(function (lead) {
      const due = normalize(lead.proximo_contacto);
      if (!due) return;
      const time = new Date(due.replace(" ", "T")).getTime();
      if (Number.isNaN(time)) return;
      items.push({
        tipo: "Lead",
        prioridad: time < now ? 0 : 1,
        titulo: normalize(lead.nombre) || normalize(lead.codigo) || ("Lead #" + lead.id),
        detalle: "Próximo contacto",
        fecha: due,
        estado: normalize(lead.estado_lead) || "nuevo"
      });
    });
    state.interacciones.forEach(function (item) {
      const next = normalize(item.proxima_accion);
      if (!next) return;
      items.push({
        tipo: "Seguimiento",
        prioridad: 2,
        titulo: normalize(item.tipo_interaccion) || normalize(item.codigo) || ("Interacción #" + item.id),
        detalle: next,
        fecha: normalize(item.fecha_interaccion),
        estado: normalize(item.estado_interaccion) || "abierta"
      });
    });
    return items.sort(function (a, b) {
      if (a.prioridad !== b.prioridad) return a.prioridad - b.prioridad;
      return normalize(a.fecha).localeCompare(normalize(b.fecha));
    }).slice(0, 10);
  }

  function renderKPIs() {
    const activeLeads = state.leads.filter(function (lead) { return normalize(lead.estado) !== "inactivo"; });
    const pipelineLeads = activeLeads.filter(function (lead) {
      return ["nuevo", "contactado", "calificado", "propuesta", "negociacion", "reactivado"].indexOf(normalize(lead.estado_lead).toLowerCase()) >= 0;
    });
    const wonLeads = activeLeads.filter(function (lead) { return normalize(lead.estado_lead).toLowerCase() === "ganado"; });
    const overdueLeads = activeLeads.filter(function (lead) {
      const due = normalize(lead.proximo_contacto);
      if (!due) return false;
      const time = new Date(due.replace(" ", "T")).getTime();
      return !Number.isNaN(time) && time < Date.now();
    });
    const openInteractions = state.interacciones.filter(function (row) {
      return ["abierta", "en_progreso", "reabierta"].indexOf(normalize(row.estado_interaccion).toLowerCase()) >= 0;
    });
    const activeCampaigns = state.campanas.filter(function (row) { return normalize(row.estado_campana).toLowerCase() === "activa"; });
    const summary = state.embudo.summary || {};
    const cards = [
      { title: "Leads activos", value: activeLeads.length, note: pipelineLeads.length + " en gestión comercial" },
      { title: "Valor potencial", value: formatMoney(activeLeads.reduce(function (sum, item) { return sum + asNumber(item.valor_potencial); }, 0)), note: "Acumulado del pipeline" },
      { title: "Seguimientos abiertos", value: openInteractions.length, note: overdueLeads.length + " leads con contacto vencido" },
      { title: "Campañas activas", value: activeCampaigns.length, note: state.campanas.length + " campañas registradas" },
      { title: "Cotizaciones", value: state.cotizaciones.length, note: (summary.cotizaciones_convertidas_pedido || 0) + " convertidas a pedido" },
      { title: "Conversión total", value: formatNumber(summary.conversion_total_pct || 0, 1) + "%", note: wonLeads.length + " leads ganados" }
    ];
    document.getElementById("crmKpiGrid").innerHTML = cards.map(function (card) {
      return "<div class=\"crm-kpi-card\"><span>" + esc(card.title) + "</span><strong>" + esc(card.value) + "</strong><small>" + esc(card.note) + "</small></div>";
    }).join("");
  }

  function renderLeads() {
    document.getElementById("leadCounterPill").textContent = state.leads.length + " leads";
    appendLeadOptions(document.getElementById("interaccionLeadId").value);
    if (!state.leads.length) {
      document.getElementById("leadsTableWrap").innerHTML = "<p class=\"form-help\">No hay leads para este filtro.</p>";
      return;
    }
    const rows = state.leads.map(function (item) {
      const stateValue = normalize(item.estado_lead).toLowerCase() || "nuevo";
      const options = ["<option value=\"\">Mover etapa...</option>"].concat(allowedTransitions(LEAD_TRANSITIONS, stateValue).map(function (next) {
        return "<option value=\"" + esc(next) + "\">" + esc(next) + "</option>";
      }));
      return "<tr>" +
        "<td>" + Number(item.id || 0) + "</td>" +
        "<td><strong>" + esc(normalize(item.nombre) || "-") + "</strong><br><small>" + esc(normalize(item.codigo) || "") + "</small></td>" +
        "<td>" + esc(normalize(item.empresa_origen) || "-") + "</td>" +
        "<td>" + esc(normalize(item.canal_origen) || "-") + "</td>" +
        "<td><span class=\"" + badgeClassForStatus(stateValue) + "\">" + esc(stateValue) + "</span></td>" +
        "<td>" + formatMoney(item.valor_potencial) + "</td>" +
        "<td>" + formatNumber(item.probabilidad || 0, 0) + "%</td>" +
        "<td>" + esc(normalize(item.propietario) || "-") + "</td>" +
        "<td>" + esc(formatDateTime(item.proximo_contacto)) + "</td>" +
        "<td class=\"crm-actions-cell\">" +
          "<button type=\"button\" class=\"btn secondary\" data-lead-edit=\"" + Number(item.id || 0) + "\">Editar</button>" +
          "<button type=\"button\" class=\"btn secondary\" data-lead-follow=\"" + Number(item.id || 0) + "\">Seguimiento</button>" +
          "<button type=\"button\" class=\"btn secondary\" data-lead-quote=\"" + Number(item.id || 0) + "\">Cotizar</button>" +
          "<button type=\"button\" class=\"btn secondary\" data-lead-client=\"" + Number(item.id || 0) + "\">Crear cliente</button>" +
          "<select class=\"form-input crm-state-select\" data-lead-transition=\"" + Number(item.id || 0) + "\">" + options.join("") + "</select>" +
        "</td>" +
      "</tr>";
    }).join("");
    document.getElementById("leadsTableWrap").innerHTML = "<table class=\"table\"><thead><tr><th>ID</th><th>Lead</th><th>Empresa</th><th>Canal</th><th>Estado</th><th>Valor</th><th>Prob.</th><th>Propietario</th><th>Próximo contacto</th><th>Acciones</th></tr></thead><tbody>" + rows + "</tbody></table>";

    document.querySelectorAll("[data-lead-edit]").forEach(function (btn) {
      btn.onclick = function () { editLead(Number(btn.getAttribute("data-lead-edit"))); };
    });
    document.querySelectorAll("[data-lead-follow]").forEach(function (btn) {
      btn.onclick = function () { prepareInteractionFromLead(Number(btn.getAttribute("data-lead-follow"))); };
    });
    document.querySelectorAll("[data-lead-quote]").forEach(function (btn) {
      btn.onclick = function () { prepareQuoteFromLead(Number(btn.getAttribute("data-lead-quote"))); };
    });
    document.querySelectorAll("[data-lead-client]").forEach(function (btn) {
      btn.onclick = function () { convertLeadToClient(Number(btn.getAttribute("data-lead-client"))); };
    });
    document.querySelectorAll("[data-lead-transition]").forEach(function (select) {
      select.onchange = async function () {
        const target = normalize(select.value);
        if (!target) return;
        try {
          await transitionRecord("/api/empresa/crm/leads?action=transicionar", Number(select.getAttribute("data-lead-transition")), target, "estado_lead");
          setMessage("crmMsg", "Lead movido a " + target + ".", "success");
          select.value = "";
        } catch (err) {
          setMessage("crmMsg", err.message || "No se pudo mover el lead.", "error");
        }
      };
    });
  }

  function renderInteracciones() {
    document.getElementById("interaccionCounterPill").textContent = state.interacciones.length + " seguimientos";
    if (!state.interacciones.length) {
      document.getElementById("interaccionesTableWrap").innerHTML = "<p class=\"form-help\">No hay seguimientos registrados.</p>";
      return;
    }
    const rows = state.interacciones.map(function (item) {
      const stateValue = normalize(item.estado_interaccion).toLowerCase() || "abierta";
      const options = ["<option value=\"\">Mover estado...</option>"].concat(allowedTransitions(INTERACTION_TRANSITIONS, stateValue).map(function (next) {
        return "<option value=\"" + esc(next) + "\">" + esc(next) + "</option>";
      }));
      return "<tr>" +
        "<td>" + Number(item.id || 0) + "</td>" +
        "<td><strong>" + esc(normalize(item.tipo_interaccion) || "-") + "</strong><br><small>" + esc(normalize(item.codigo) || "") + "</small></td>" +
        "<td>" + esc(formatDateTime(item.fecha_interaccion)) + "</td>" +
        "<td>" + esc(normalize(item.usuario_responsable) || "-") + "</td>" +
        "<td>" + esc(normalize(item.resumen) || "-") + "</td>" +
        "<td>" + esc(normalize(item.proxima_accion) || "-") + "</td>" +
        "<td><span class=\"" + badgeClassForStatus(stateValue) + "\">" + esc(stateValue) + "</span></td>" +
        "<td class=\"crm-actions-cell\">" +
          "<button type=\"button\" class=\"btn secondary\" data-int-edit=\"" + Number(item.id || 0) + "\">Editar</button>" +
          "<select class=\"form-input crm-state-select\" data-int-transition=\"" + Number(item.id || 0) + "\">" + options.join("") + "</select>" +
        "</td>" +
      "</tr>";
    }).join("");
    document.getElementById("interaccionesTableWrap").innerHTML = "<table class=\"table\"><thead><tr><th>ID</th><th>Interacción</th><th>Fecha</th><th>Responsable</th><th>Resumen</th><th>Próxima acción</th><th>Estado</th><th>Acciones</th></tr></thead><tbody>" + rows + "</tbody></table>";

    document.querySelectorAll("[data-int-edit]").forEach(function (btn) {
      btn.onclick = function () { editInteraccion(Number(btn.getAttribute("data-int-edit"))); };
    });
    document.querySelectorAll("[data-int-transition]").forEach(function (select) {
      select.onchange = async function () {
        const target = normalize(select.value);
        if (!target) return;
        try {
          await transitionRecord("/api/empresa/crm/interacciones?action=transicionar", Number(select.getAttribute("data-int-transition")), target, "estado_interaccion");
          setMessage("crmMsg", "Seguimiento movido a " + target + ".", "success");
          select.value = "";
        } catch (err) {
          setMessage("crmMsg", err.message || "No se pudo mover el seguimiento.", "error");
        }
      };
    });
  }

  function renderCotizaciones() {
    document.getElementById("cotizacionCounterPill").textContent = state.cotizaciones.length + " cotizaciones";
    if (!state.cotizaciones.length) {
      document.getElementById("cotizacionesTableWrap").innerHTML = "<p class=\"form-help\">No hay cotizaciones para este filtro.</p>";
      return;
    }
    const rows = state.cotizaciones.map(function (item) {
      const stateValue = normalize(item.estado_documento).toLowerCase() || "borrador";
      const options = ["<option value=\"\">Mover estado...</option>"].concat(allowedTransitions(QUOTE_TRANSITIONS, stateValue).map(function (next) {
        return "<option value=\"" + esc(next) + "\">" + esc(next) + "</option>";
      }));
      return "<tr>" +
        "<td>" + Number(item.id || 0) + "</td>" +
        "<td><strong>" + esc(normalize(item.codigo) || "-") + "</strong><br><small>" + esc(normalize(item.cliente_nombre) || "-") + "</small></td>" +
        "<td>" + esc(formatDate(item.fecha_documento)) + "</td>" +
        "<td>" + esc(formatDate(item.vigencia_hasta)) + "</td>" +
        "<td><span class=\"" + badgeClassForStatus(stateValue) + "\">" + esc(stateValue) + "</span></td>" +
        "<td>" + formatMoney(item.total) + "</td>" +
        "<td>" + esc(normalize(item.origen) || "-") + "</td>" +
        "<td class=\"crm-actions-cell\">" +
          "<button type=\"button\" class=\"btn secondary\" data-cot-edit=\"" + Number(item.id || 0) + "\">Editar</button>" +
          "<button type=\"button\" class=\"btn secondary\" data-cot-pedido=\"" + Number(item.id || 0) + "\">A pedido</button>" +
          "<button type=\"button\" class=\"btn secondary\" data-cot-doc=\"" + Number(item.id || 0) + "\">Doc. final</button>" +
          "<select class=\"form-input crm-state-select\" data-cot-transition=\"" + Number(item.id || 0) + "\">" + options.join("") + "</select>" +
        "</td>" +
      "</tr>";
    }).join("");
    document.getElementById("cotizacionesTableWrap").innerHTML = "<table class=\"table\"><thead><tr><th>ID</th><th>Cotización</th><th>Fecha</th><th>Vigencia</th><th>Estado</th><th>Total</th><th>Origen</th><th>Acciones</th></tr></thead><tbody>" + rows + "</tbody></table>";

    document.querySelectorAll("[data-cot-edit]").forEach(function (btn) {
      btn.onclick = function () { editCotizacion(Number(btn.getAttribute("data-cot-edit"))); };
    });
    document.querySelectorAll("[data-cot-pedido]").forEach(function (btn) {
      btn.onclick = function () { convertQuoteAction(Number(btn.getAttribute("data-cot-pedido")), "convertir_pedido"); };
    });
    document.querySelectorAll("[data-cot-doc]").forEach(function (btn) {
      btn.onclick = function () { convertQuoteAction(Number(btn.getAttribute("data-cot-doc")), "convertir_documento_final"); };
    });
    document.querySelectorAll("[data-cot-transition]").forEach(function (select) {
      select.onchange = async function () {
        const target = normalize(select.value);
        if (!target) return;
        try {
          await transitionRecord("/api/empresa/ventas/cotizaciones?action=transicionar", Number(select.getAttribute("data-cot-transition")), target, "estado_documento");
          setMessage("crmMsg", "Cotización movida a " + target + ".", "success");
          select.value = "";
        } catch (err) {
          setMessage("crmMsg", err.message || "No se pudo mover la cotización.", "error");
        }
      };
    });
  }

  function renderCampanas() {
    document.getElementById("campanaCounterPill").textContent = state.campanas.length + " campañas";
    if (!state.campanas.length) {
      document.getElementById("campanasTableWrap").innerHTML = "<p class=\"form-help\">No hay campañas registradas.</p>";
      return;
    }
    const rows = state.campanas.map(function (item) {
      const stateValue = normalize(item.estado_campana).toLowerCase() || "planificada";
      const options = ["<option value=\"\">Mover estado...</option>"].concat(allowedTransitions(CAMPAIGN_TRANSITIONS, stateValue).map(function (next) {
        return "<option value=\"" + esc(next) + "\">" + esc(next) + "</option>";
      }));
      return "<tr>" +
        "<td>" + Number(item.id || 0) + "</td>" +
        "<td><strong>" + esc(normalize(item.nombre) || "-") + "</strong><br><small>" + esc(normalize(item.codigo) || "") + "</small></td>" +
        "<td>" + esc(normalize(item.canal) || "-") + "</td>" +
        "<td>" + esc(formatDate(item.fecha_inicio)) + "</td>" +
        "<td>" + esc(formatDate(item.fecha_fin)) + "</td>" +
        "<td>" + formatMoney(item.presupuesto) + "</td>" +
        "<td><span class=\"" + badgeClassForStatus(stateValue) + "\">" + esc(stateValue) + "</span></td>" +
        "<td class=\"crm-actions-cell\">" +
          "<button type=\"button\" class=\"btn secondary\" data-camp-edit=\"" + Number(item.id || 0) + "\">Editar</button>" +
          "<select class=\"form-input crm-state-select\" data-camp-transition=\"" + Number(item.id || 0) + "\">" + options.join("") + "</select>" +
        "</td>" +
      "</tr>";
    }).join("");
    document.getElementById("campanasTableWrap").innerHTML = "<table class=\"table\"><thead><tr><th>ID</th><th>Campaña</th><th>Canal</th><th>Inicio</th><th>Fin</th><th>Presupuesto</th><th>Estado</th><th>Acciones</th></tr></thead><tbody>" + rows + "</tbody></table>";

    document.querySelectorAll("[data-camp-edit]").forEach(function (btn) {
      btn.onclick = function () { editCampana(Number(btn.getAttribute("data-camp-edit"))); };
    });
    document.querySelectorAll("[data-camp-transition]").forEach(function (select) {
      select.onchange = async function () {
        const target = normalize(select.value);
        if (!target) return;
        try {
          await transitionRecord("/api/empresa/crm/campanas?action=transicionar", Number(select.getAttribute("data-camp-transition")), target, "estado_campana");
          setMessage("crmMsg", "Campaña movida a " + target + ".", "success");
          select.value = "";
        } catch (err) {
          setMessage("crmMsg", err.message || "No se pudo mover la campaña.", "error");
        }
      };
    });
  }

  function renderEmbudo() {
    const summary = state.embudo.summary || {};
    const alertas = Array.isArray(state.embudo.alertas) ? state.embudo.alertas : [];
    const items = Array.isArray(state.embudo.items) ? state.embudo.items : [];
    document.getElementById("embudoConversionBadge").textContent = "Conversión " + formatNumber(summary.conversion_total_pct || 0, 1) + "%";
    document.getElementById("embudoAlertBadge").textContent = (summary.alertas_total || 0) + " alertas";
    document.getElementById("embudoResumen").innerHTML = "<div class=\"crm-mini-list\">" +
      "<div class=\"crm-mini-item\"><strong>Cotizaciones:</strong> " + Number(summary.cotizaciones_total || 0) + " | <strong>A pedido:</strong> " + Number(summary.cotizaciones_convertidas_pedido || 0) + " | <strong>Con documento final:</strong> " + Number(summary.cotizaciones_documento_final || 0) + "</div>" +
      "<div class=\"crm-mini-item\"><strong>Conversión cotización a pedido:</strong> " + formatNumber(summary.conversion_cotizacion_pedido_pct || 0, 1) + "% | <strong>Pedido a documento:</strong> " + formatNumber(summary.conversion_pedido_documento_pct || 0, 1) + "%</div>" +
      "<div class=\"crm-mini-item\"><strong>SLA cotización:</strong> " + Number(summary.sla_cotizacion_horas || 0) + " h | <strong>SLA pedido:</strong> " + Number(summary.sla_pedido_horas || 0) + " h</div>" +
      "</div>";

    document.getElementById("embudoAlertas").innerHTML = alertas.length ? alertas.map(function (alerta) {
      return "<div class=\"crm-mini-item\"><strong>" + esc(normalize(alerta.cotizacion_codigo) || ("Cotización #" + alerta.cotizacion_id)) + "</strong><br>" + esc(normalize(alerta.alerta) || "Alerta comercial") + "<br><small>Etapa: " + esc(normalize(alerta.conversion_etapa) || "-") + " | Estado cotización: " + esc(normalize(alerta.estado_cotizacion) || "-") + "</small></div>";
    }).join("") : "<p class=\"form-help\">Sin alertas de SLA o vigencia en el rango actual.</p>";

    const agenda = computeAgenda();
    document.getElementById("crmAgendaWrap").innerHTML = agenda.length ? agenda.map(function (entry) {
      return "<div class=\"crm-mini-item\"><strong>" + esc(entry.tipo + ": " + entry.titulo) + "</strong><br>" + esc(entry.detalle) + "<br><small>" + esc(formatDateTime(entry.fecha)) + " | " + esc(entry.estado) + "</small></div>";
    }).join("") : "<p class=\"form-help\">No hay próximas acciones registradas.</p>";

    if (!items.length) {
      document.getElementById("embudoTableWrap").innerHTML = "<p class=\"form-help\">No hay datos del embudo comercial.</p>";
      return;
    }
    const rows = items.map(function (item) {
      return "<tr>" +
        "<td>" + esc(normalize(item.cotizacion_codigo) || "-") + "</td>" +
        "<td>" + esc(normalize(item.estado_cotizacion) || "-") + "</td>" +
        "<td>" + esc(normalize(item.pedido_codigo) || "-") + "</td>" +
        "<td>" + esc(normalize(item.estado_pedido) || "-") + "</td>" +
        "<td>" + esc(normalize(item.documento_final_codigo) || "-") + "</td>" +
        "<td>" + esc(normalize(item.conversion_etapa) || "-") + "</td>" +
        "<td>" + Number(item.horas_desde_cotizacion || 0) + " h</td>" +
        "<td>" + esc(normalize(item.alerta) || "-") + "</td>" +
      "</tr>";
    }).join("");
    document.getElementById("embudoTableWrap").innerHTML = "<table class=\"table\"><thead><tr><th>Cotización</th><th>Estado cot.</th><th>Pedido</th><th>Estado ped.</th><th>Doc. final</th><th>Etapa</th><th>Tiempo</th><th>Alerta</th></tr></thead><tbody>" + rows + "</tbody></table>";
  }

  function fillLeadForm(item) {
    document.getElementById("leadId").value = item ? String(item.id || "") : "";
    document.getElementById("leadCodigo").value = item ? normalize(item.codigo) : "";
    document.getElementById("leadEstado").value = item ? (normalize(item.estado_lead) || "nuevo") : "nuevo";
    document.getElementById("leadNombre").value = item ? normalize(item.nombre) : "";
    document.getElementById("leadEmpresaOrigen").value = item ? normalize(item.empresa_origen) : "";
    document.getElementById("leadEmail").value = item ? normalize(item.email) : "";
    document.getElementById("leadTelefono").value = item ? normalize(item.telefono) : "";
    document.getElementById("leadCanal").value = item ? normalize(item.canal_origen) : "";
    document.getElementById("leadPropietario").value = item ? normalize(item.propietario) : "";
    document.getElementById("leadValor").value = item ? String(asNumber(item.valor_potencial) || "") : "";
    document.getElementById("leadProbabilidad").value = item ? String(asNumber(item.probabilidad) || "") : "";
    document.getElementById("leadProximoContacto").value = item ? toDateTimeInputValue(item.proximo_contacto) : "";
    document.getElementById("leadNotas").value = item ? normalize(item.notas) : "";
    document.getElementById("leadObservaciones").value = item ? normalize(item.observaciones) : "";
    document.getElementById("leadSaveBtn").textContent = item ? "Actualizar lead" : "Guardar lead";
    setMessage("leadMsg", "", "");
  }

  function fillInteraccionForm(item) {
    document.getElementById("interaccionId").value = item ? String(item.id || "") : "";
    document.getElementById("interaccionCodigo").value = item ? normalize(item.codigo) : "";
    appendLeadOptions(item ? item.lead_id : "");
    document.getElementById("interaccionEstado").value = item ? (normalize(item.estado_interaccion) || "abierta") : "abierta";
    document.getElementById("interaccionClienteId").value = item ? String(asNumber(item.cliente_id) || "") : "";
    document.getElementById("interaccionTipo").value = item ? (normalize(item.tipo_interaccion) || "seguimiento") : "seguimiento";
    document.getElementById("interaccionFecha").value = item ? toDateTimeInputValue(item.fecha_interaccion) : toDateTimeInputValue(new Date().toISOString());
    document.getElementById("interaccionResponsable").value = item ? normalize(item.usuario_responsable) : "";
    document.getElementById("interaccionResultado").value = item ? normalize(item.resultado) : "";
    document.getElementById("interaccionResumen").value = item ? normalize(item.resumen) : "";
    document.getElementById("interaccionProximaAccion").value = item ? normalize(item.proxima_accion) : "";
    document.getElementById("interaccionObservaciones").value = item ? normalize(item.observaciones) : "";
    document.getElementById("interaccionSaveBtn").textContent = item ? "Actualizar seguimiento" : "Guardar seguimiento";
    setMessage("interaccionMsg", "", "");
  }

  function fillCotizacionForm(item) {
    document.getElementById("cotizacionId").value = item ? String(item.id || "") : "";
    document.getElementById("cotizacionCodigo").value = item ? normalize(item.codigo) : "";
    document.getElementById("cotizacionEstado").value = item ? (normalize(item.estado_documento) || "borrador") : "borrador";
    document.getElementById("cotizacionClienteNombre").value = item ? normalize(item.cliente_nombre) : "";
    document.getElementById("cotizacionClienteId").value = item ? String(asNumber(item.cliente_id) || "") : "";
    document.getElementById("cotizacionFecha").value = item ? toDateInputValue(item.fecha_documento) : nowDateInput();
    document.getElementById("cotizacionVigencia").value = item ? toDateInputValue(item.vigencia_hasta) : "";
    document.getElementById("cotizacionSubtotal").value = item ? String(asNumber(item.subtotal) || "") : "";
    document.getElementById("cotizacionDescuento").value = item ? String(asNumber(item.descuento_total) || "") : "";
    document.getElementById("cotizacionImpuesto").value = item ? String(asNumber(item.impuesto_total) || "") : "";
    document.getElementById("cotizacionTotal").value = item ? String(asNumber(item.total) || "") : "";
    document.getElementById("cotizacionMoneda").value = item ? (normalize(item.moneda) || "COP") : "COP";
    document.getElementById("cotizacionOrigen").value = item ? normalize(item.origen) : "";
    document.getElementById("cotizacionNotas").value = item ? normalize(item.notas) : "";
    document.getElementById("cotizacionObservaciones").value = item ? normalize(item.observaciones) : "";
    document.getElementById("cotizacionSaveBtn").textContent = item ? "Actualizar cotización" : "Guardar cotización";
    setMessage("cotizacionMsg", "", "");
  }

  function fillCampanaForm(item) {
    document.getElementById("campanaId").value = item ? String(item.id || "") : "";
    document.getElementById("campanaCodigo").value = item ? normalize(item.codigo) : "";
    document.getElementById("campanaEstado").value = item ? (normalize(item.estado_campana) || "planificada") : "planificada";
    document.getElementById("campanaNombre").value = item ? normalize(item.nombre) : "";
    document.getElementById("campanaCanal").value = item ? (normalize(item.canal) || "email") : "email";
    document.getElementById("campanaObjetivo").value = item ? normalize(item.objetivo) : "";
    document.getElementById("campanaPresupuesto").value = item ? String(asNumber(item.presupuesto) || "") : "";
    document.getElementById("campanaAudiencia").value = item ? normalize(item.audiencia) : "";
    document.getElementById("campanaFechaInicio").value = item ? toDateInputValue(item.fecha_inicio) : "";
    document.getElementById("campanaFechaFin").value = item ? toDateInputValue(item.fecha_fin) : "";
    document.getElementById("campanaKPI").value = item ? normalize(item.kpi_objetivo) : "";
    document.getElementById("campanaResultados").value = item ? normalize(item.resultado_json) : "";
    document.getElementById("campanaObservaciones").value = item ? normalize(item.observaciones) : "";
    document.getElementById("campanaSaveBtn").textContent = item ? "Actualizar campaña" : "Guardar campaña";
    setMessage("campanaMsg", "", "");
  }

  function editLead(id) {
    fillLeadForm(state.leads.find(function (item) { return Number(item.id || 0) === id; }) || null);
    setActiveTab("leads");
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  function editInteraccion(id) {
    fillInteraccionForm(state.interacciones.find(function (item) { return Number(item.id || 0) === id; }) || null);
    setActiveTab("interacciones");
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  function editCotizacion(id) {
    fillCotizacionForm(state.cotizaciones.find(function (item) { return Number(item.id || 0) === id; }) || null);
    setActiveTab("cotizaciones");
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  function editCampana(id) {
    fillCampanaForm(state.campanas.find(function (item) { return Number(item.id || 0) === id; }) || null);
    setActiveTab("campanas");
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  function prepareInteractionFromLead(id) {
    const lead = state.leads.find(function (item) { return Number(item.id || 0) === id; });
    fillInteraccionForm(null);
    appendLeadOptions(id);
    if (lead) {
      document.getElementById("interaccionTipo").value = "seguimiento";
      document.getElementById("interaccionResumen").value = "Seguimiento comercial de " + (normalize(lead.nombre) || normalize(lead.codigo) || "lead");
      document.getElementById("interaccionResponsable").value = normalize(lead.propietario);
      document.getElementById("interaccionProximaAccion").value = normalize(lead.notas);
    }
    setActiveTab("interacciones");
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  function prepareQuoteFromLead(id) {
    const lead = state.leads.find(function (item) { return Number(item.id || 0) === id; });
    fillCotizacionForm(null);
    if (lead) {
      document.getElementById("cotizacionClienteNombre").value = normalize(lead.nombre) || normalize(lead.empresa_origen);
      document.getElementById("cotizacionOrigen").value = normalize(lead.codigo) || normalize(lead.canal_origen);
      document.getElementById("cotizacionNotas").value = normalize(lead.notas);
      document.getElementById("cotizacionObservaciones").value = "Preparada desde lead " + (normalize(lead.codigo) || ("#" + id));
    }
    setActiveTab("cotizaciones");
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  async function transitionRecord(path, id, targetState, stateColumn) {
    await fetchJSON(path, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ empresa_id: state.empresaID, id: id, nuevo_estado: targetState, [stateColumn]: targetState })
    });
    await loadAllData();
  }

  async function convertLeadToClient(id) {
    const lead = state.leads.find(function (item) { return Number(item.id || 0) === id; });
    if (!lead) return;
    const provisionalDoc = normalize(lead.codigo) || ("LEAD-" + id);
    const payload = {
      empresa_id: state.empresaID,
      tipo_documento: "OTRO",
      numero_documento: provisionalDoc,
      nombre_razon_social: normalize(lead.nombre) || normalize(lead.empresa_origen) || ("Lead " + provisionalDoc),
      nombre_comercial: normalize(lead.empresa_origen),
      email: normalize(lead.email),
      telefono: normalize(lead.telefono),
      observaciones: [
        "Cliente provisional generado desde CRM comercial.",
        normalize(lead.notas),
        normalize(lead.observaciones)
      ].filter(Boolean).join(" | ")
    };
    try {
      const response = await fetchJSON("/api/empresa/clientes", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });
      const nextObservaciones = [normalize(lead.observaciones), "cliente_creado_desde_lead=" + String(response.id || "")]
        .filter(Boolean)
        .join(" | ");
      await fetchJSON("/api/empresa/crm/leads", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ empresa_id: state.empresaID, id: id, observaciones: nextObservaciones })
      });
      setMessage("crmMsg", "Cliente creado desde el lead #" + id + ".", "success");
      await loadAllData();
    } catch (err) {
      setMessage("crmMsg", err.message || "No se pudo crear el cliente.", "error");
    }
  }

  async function convertQuoteAction(id, action) {
    try {
      const response = await fetchJSON("/api/empresa/ventas/cotizaciones?action=" + encodeURIComponent(action), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ empresa_id: state.empresaID, id: id })
      });
      const label = action === "convertir_pedido" ? "pedido" : "documento final";
      setMessage("crmMsg", "Cotización convertida a " + label + " correctamente.", "success");
      if (response && response.documento_final && response.documento_final.documento_codigo) {
        setMessage("cotizacionMsg", "Documento generado: " + response.documento_final.documento_codigo, "success");
      }
      await loadAllData();
      setActiveTab("cotizaciones");
    } catch (err) {
      setMessage("crmMsg", err.message || "No se pudo convertir la cotización.", "error");
    }
  }

  async function loadAllData() {
    if (!state.empresaID) {
      throw new Error("empresa_id es obligatorio para usar el CRM.");
    }
    const results = await Promise.all([
      fetchJSON(buildURL("/api/empresa/crm/leads")),
      fetchJSON(buildURL("/api/empresa/crm/interacciones")),
      fetchJSON(buildURL("/api/empresa/crm/campanas")),
      fetchJSON(buildURL("/api/empresa/ventas/cotizaciones")),
      fetchJSON(buildURL("/api/empresa/ventas/cotizaciones", { action: "embudo", limit: 30 }))
    ]);

    state.leads = Array.isArray(results[0]) ? results[0] : [];
    state.interacciones = Array.isArray(results[1]) ? results[1] : [];
    state.campanas = Array.isArray(results[2]) ? results[2] : [];
    state.cotizaciones = Array.isArray(results[3]) ? results[3] : [];
    state.embudo = results[4] || { summary: {}, items: [], alertas: [] };

    renderKPIs();
    renderLeads();
    renderInteracciones();
    renderCotizaciones();
    renderCampanas();
    renderEmbudo();
  }

  function bindEvents() {
    document.querySelectorAll(".crm-tab-btn").forEach(function (btn) {
      btn.addEventListener("click", function () { setActiveTab(btn.getAttribute("data-tab")); });
    });

    document.getElementById("crmSearchBtn").addEventListener("click", async function () {
      state.q = normalize(document.getElementById("crmSearch").value);
      state.includeInactive = !!document.getElementById("crmIncludeInactive").checked;
      try {
        await loadAllData();
        setMessage("crmMsg", "Filtros aplicados.", "success");
      } catch (err) {
        setMessage("crmMsg", err.message || "No se pudieron aplicar los filtros.", "error");
      }
    });

    document.getElementById("crmRefreshBtn").addEventListener("click", async function () {
      try {
        await loadAllData();
        setMessage("crmMsg", "CRM actualizado.", "success");
      } catch (err) {
        setMessage("crmMsg", err.message || "No se pudo actualizar el CRM.", "error");
      }
    });

    document.getElementById("crmExportBtn").addEventListener("click", function () {
      try {
        if (state.activeTab === "leads") exportRowsToCSV("crm_leads.csv", state.leads);
        else if (state.activeTab === "interacciones") exportRowsToCSV("crm_interacciones.csv", state.interacciones);
        else if (state.activeTab === "cotizaciones") exportRowsToCSV("crm_cotizaciones.csv", state.cotizaciones);
        else if (state.activeTab === "campanas") exportRowsToCSV("crm_campanas.csv", state.campanas);
        else exportRowsToCSV("crm_embudo.csv", state.embudo.items || []);
        setMessage("crmMsg", "Exportación generada para la vista actual.", "success");
      } catch (err) {
        setMessage("crmMsg", err.message || "No se pudo exportar la vista actual.", "error");
      }
    });

    document.getElementById("leadForm").addEventListener("submit", async function (event) {
      event.preventDefault();
      const id = Number(document.getElementById("leadId").value || 0);
      const payload = {
        empresa_id: state.empresaID,
        id: id,
        estado_lead: normalize(document.getElementById("leadEstado").value) || "nuevo",
        nombre: normalize(document.getElementById("leadNombre").value),
        empresa_origen: normalize(document.getElementById("leadEmpresaOrigen").value),
        email: normalize(document.getElementById("leadEmail").value),
        telefono: normalize(document.getElementById("leadTelefono").value),
        canal_origen: normalize(document.getElementById("leadCanal").value),
        valor_potencial: asNumber(document.getElementById("leadValor").value),
        probabilidad: asNumber(document.getElementById("leadProbabilidad").value),
        propietario: normalize(document.getElementById("leadPropietario").value),
        proximo_contacto: normalize(document.getElementById("leadProximoContacto").value).replace("T", " "),
        notas: normalize(document.getElementById("leadNotas").value),
        observaciones: normalize(document.getElementById("leadObservaciones").value)
      };
      if (!payload.nombre) {
        setMessage("leadMsg", "El nombre del prospecto es obligatorio.", "error");
        return;
      }
      try {
        await fetchJSON("/api/empresa/crm/leads", {
          method: id > 0 ? "PUT" : "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload)
        });
        fillLeadForm(null);
        setMessage("leadMsg", id > 0 ? "Lead actualizado." : "Lead creado.", "success");
        setMessage("crmMsg", "Lead guardado correctamente.", "success");
        await loadAllData();
      } catch (err) {
        setMessage("leadMsg", err.message || "No se pudo guardar el lead.", "error");
      }
    });

    document.getElementById("interaccionForm").addEventListener("submit", async function (event) {
      event.preventDefault();
      const id = Number(document.getElementById("interaccionId").value || 0);
      const payload = {
        empresa_id: state.empresaID,
        id: id,
        lead_id: Number(document.getElementById("interaccionLeadId").value || 0),
        cliente_id: Number(document.getElementById("interaccionClienteId").value || 0),
        tipo_interaccion: normalize(document.getElementById("interaccionTipo").value),
        fecha_interaccion: normalize(document.getElementById("interaccionFecha").value).replace("T", " "),
        resumen: normalize(document.getElementById("interaccionResumen").value),
        resultado: normalize(document.getElementById("interaccionResultado").value),
        usuario_responsable: normalize(document.getElementById("interaccionResponsable").value),
        proxima_accion: normalize(document.getElementById("interaccionProximaAccion").value),
        estado_interaccion: normalize(document.getElementById("interaccionEstado").value) || "abierta",
        observaciones: normalize(document.getElementById("interaccionObservaciones").value)
      };
      if (!payload.tipo_interaccion || !payload.resumen) {
        setMessage("interaccionMsg", "Tipo de interacción y resumen son obligatorios.", "error");
        return;
      }
      try {
        await fetchJSON("/api/empresa/crm/interacciones", {
          method: id > 0 ? "PUT" : "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload)
        });
        fillInteraccionForm(null);
        setMessage("interaccionMsg", id > 0 ? "Seguimiento actualizado." : "Seguimiento creado.", "success");
        setMessage("crmMsg", "Seguimiento comercial guardado.", "success");
        await loadAllData();
      } catch (err) {
        setMessage("interaccionMsg", err.message || "No se pudo guardar el seguimiento.", "error");
      }
    });

    document.getElementById("cotizacionForm").addEventListener("submit", async function (event) {
      event.preventDefault();
      const id = Number(document.getElementById("cotizacionId").value || 0);
      const payload = {
        empresa_id: state.empresaID,
        id: id,
        cliente_id: Number(document.getElementById("cotizacionClienteId").value || 0),
        cliente_nombre: normalize(document.getElementById("cotizacionClienteNombre").value),
        fecha_documento: normalize(document.getElementById("cotizacionFecha").value),
        vigencia_hasta: normalize(document.getElementById("cotizacionVigencia").value),
        estado_documento: normalize(document.getElementById("cotizacionEstado").value) || "borrador",
        subtotal: asNumber(document.getElementById("cotizacionSubtotal").value),
        descuento_total: asNumber(document.getElementById("cotizacionDescuento").value),
        impuesto_total: asNumber(document.getElementById("cotizacionImpuesto").value),
        total: asNumber(document.getElementById("cotizacionTotal").value),
        moneda: normalize(document.getElementById("cotizacionMoneda").value) || "COP",
        origen: normalize(document.getElementById("cotizacionOrigen").value),
        notas: normalize(document.getElementById("cotizacionNotas").value),
        observaciones: normalize(document.getElementById("cotizacionObservaciones").value)
      };
      if (!payload.cliente_nombre) {
        setMessage("cotizacionMsg", "El cliente o prospecto es obligatorio.", "error");
        return;
      }
      if (!payload.total) {
        payload.total = payload.subtotal - payload.descuento_total + payload.impuesto_total;
      }
      try {
        await fetchJSON("/api/empresa/ventas/cotizaciones", {
          method: id > 0 ? "PUT" : "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload)
        });
        fillCotizacionForm(null);
        setMessage("cotizacionMsg", id > 0 ? "Cotización actualizada." : "Cotización creada.", "success");
        setMessage("crmMsg", "Cotización guardada correctamente.", "success");
        await loadAllData();
      } catch (err) {
        setMessage("cotizacionMsg", err.message || "No se pudo guardar la cotización.", "error");
      }
    });

    document.getElementById("campanaForm").addEventListener("submit", async function (event) {
      event.preventDefault();
      const id = Number(document.getElementById("campanaId").value || 0);
      const payload = {
        empresa_id: state.empresaID,
        id: id,
        nombre: normalize(document.getElementById("campanaNombre").value),
        canal: normalize(document.getElementById("campanaCanal").value),
        objetivo: normalize(document.getElementById("campanaObjetivo").value),
        presupuesto: asNumber(document.getElementById("campanaPresupuesto").value),
        fecha_inicio: normalize(document.getElementById("campanaFechaInicio").value),
        fecha_fin: normalize(document.getElementById("campanaFechaFin").value),
        estado_campana: normalize(document.getElementById("campanaEstado").value) || "planificada",
        audiencia: normalize(document.getElementById("campanaAudiencia").value),
        kpi_objetivo: normalize(document.getElementById("campanaKPI").value),
        resultado_json: normalize(document.getElementById("campanaResultados").value),
        observaciones: normalize(document.getElementById("campanaObservaciones").value)
      };
      if (!payload.nombre || !payload.canal) {
        setMessage("campanaMsg", "Nombre y canal son obligatorios.", "error");
        return;
      }
      try {
        await fetchJSON("/api/empresa/crm/campanas", {
          method: id > 0 ? "PUT" : "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload)
        });
        fillCampanaForm(null);
        setMessage("campanaMsg", id > 0 ? "Campaña actualizada." : "Campaña creada.", "success");
        setMessage("crmMsg", "Campaña guardada correctamente.", "success");
        await loadAllData();
      } catch (err) {
        setMessage("campanaMsg", err.message || "No se pudo guardar la campaña.", "error");
      }
    });

    document.getElementById("leadResetBtn").addEventListener("click", function () { fillLeadForm(null); });
    document.getElementById("leadCancelBtn").addEventListener("click", function () { fillLeadForm(null); });
    document.getElementById("interaccionResetBtn").addEventListener("click", function () { fillInteraccionForm(null); });
    document.getElementById("interaccionCancelBtn").addEventListener("click", function () { fillInteraccionForm(null); });
    document.getElementById("cotizacionResetBtn").addEventListener("click", function () { fillCotizacionForm(null); });
    document.getElementById("cotizacionCancelBtn").addEventListener("click", function () { fillCotizacionForm(null); });
    document.getElementById("campanaResetBtn").addEventListener("click", function () { fillCampanaForm(null); });
    document.getElementById("campanaCancelBtn").addEventListener("click", function () { fillCampanaForm(null); });
  }

  async function init() {
    state.empresaID = Number(queryParam("empresa_id") || queryParam("id") || 0);
    bindEvents();
    fillLeadForm(null);
    fillInteraccionForm(null);
    fillCotizacionForm(null);
    fillCampanaForm(null);
    setActiveTab("leads");
    try {
      await loadAllData();
      setMessage("crmMsg", "CRM comercial listo para operar.", "success");
    } catch (err) {
      setMessage("crmMsg", err.message || "No se pudo iniciar el CRM comercial.", "error");
    }
  }

  init();
})();
