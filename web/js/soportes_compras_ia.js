(function () {
  "use strict";

  var state = {
    empresaId: resolveEmpresaId(),
    dashboard: {},
    soportes: [],
    selected: null,
    filters: { estado: "", tipo: "", search: "" },
    loading: false
  };

  var labels = {
    radicado: "Radicado",
    extraido: "Extraido",
    en_revision: "En revision",
    aprobado: "Aprobado",
    contabilizado: "Contabilizado",
    rechazado: "Rechazado",
    duplicado: "Duplicado",
    gasto: "Gasto",
    compra: "Compra",
    documento_soporte: "Documento soporte",
    servicio: "Servicio",
    recibo: "Recibo",
    factura_compra: "Factura de compra",
    cuenta_cobro: "Cuenta de cobro",
    recibo_caja: "Recibo de caja",
    otro: "Otro"
  };
  var moneyFmt = new Intl.NumberFormat("es-CO", { style: "currency", currency: "COP", maximumFractionDigits: 0 });

  function el(id) { return document.getElementById(id); }
  function val(id) { return (el(id) && el(id).value || "").trim(); }
  function esc(v) { return String(v == null ? "" : v).replace(/[&<>"']/g, function (ch) { return { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[ch]; }); }
  function label(v) { return labels[String(v || "").toLowerCase()] || String(v || "-").replace(/_/g, " "); }
  function money(v) { try { return moneyFmt.format(Number(v) || 0); } catch (_) { return "$" + (Number(v) || 0).toFixed(0); } }
  function pct(v) { return Math.round((Number(v) || 0) * 100) + "%"; }

  function resolveEmpresaId() {
    try {
      if (window.__resolveEmpresaIdContext) return String(window.__resolveEmpresaIdContext() || "");
    } catch (_) {}
    var sources = [location.search];
    try { if (window.parent && window.parent !== window) sources.push(window.parent.location.search); } catch (_) {}
    for (var i = 0; i < sources.length; i += 1) {
      var params = new URLSearchParams(sources[i] || "");
      var id = params.get("empresa_id") || params.get("id");
      if (id) return id;
    }
    try { return sessionStorage.getItem("empresa_id") || localStorage.getItem("empresa_id") || ""; } catch (_) { return ""; }
  }

  function msg(text, isError) {
    var target = el("captureMsg");
    if (!target) return;
    target.textContent = text || "";
    target.classList.toggle("is-error", !!isError);
  }

  function setBusy(on, text) {
    state.loading = !!on;
    document.querySelectorAll(".capture-btn").forEach(function (btn) { btn.disabled = !!on; });
    if (text) msg(text);
  }

  function api(action, options, extra) {
    if (!state.empresaId) return Promise.reject(new Error("empresa_id no disponible"));
    var url = "/api/empresa/soportes_compras_ia?empresa_id=" + encodeURIComponent(state.empresaId) + "&action=" + encodeURIComponent(action) + (extra || "");
    var requestOptions = Object.assign({ credentials: "same-origin" }, options || {});
    if (!(requestOptions.body instanceof FormData)) {
      requestOptions.headers = Object.assign({ "Content-Type": "application/json" }, requestOptions.headers || {});
    }
    return fetch(url, requestOptions).then(function (res) {
      return res.text().then(function (txt) {
        var data = {};
        try { data = txt ? JSON.parse(txt) : {}; } catch (_) { data = { raw: txt }; }
        if (!res.ok || data.ok === false) throw new Error(data.error || data.raw || txt || res.statusText);
        return data;
      });
    });
  }

  function chip(value, extra) {
    return '<span class="capture-chip ' + (extra || estadoClass(value)) + '">' + esc(label(value)) + "</span>";
  }

  function estadoClass(value) {
    var v = String(value || "").toLowerCase();
    if (v === "aprobado" || v === "contabilizado") return "capture-good";
    if (v === "rechazado" || v === "duplicado") return "capture-bad";
    return "capture-warn";
  }

  function progress(value) {
    var n = Math.max(0, Math.min(100, Math.round((Number(value) || 0) * 100)));
    return '<div class="capture-progress"><span style="width:' + n + '%"></span></div><span class="capture-muted">' + n + "%</span>";
  }

  function rowList() {
    var q = state.filters.search.toLowerCase();
    return state.soportes.filter(function (s) {
      if (state.filters.estado && s.estado_soporte !== state.filters.estado) return false;
      if (state.filters.tipo && s.tipo_soporte !== state.filters.tipo) return false;
      if (!q) return true;
      return [s.codigo, s.proveedor_nombre, s.proveedor_nit, s.documento_numero, s.documento_tipo, s.archivo_nombre].join(" ").toLowerCase().indexOf(q) >= 0;
    });
  }

  function fillKpis(d) {
    el("kpiPendientes").textContent = d.pendientes || 0;
    el("kpiRevision").textContent = d.en_revision || 0;
    el("kpiAprobados").textContent = d.aprobados || 0;
    el("kpiContabilizados").textContent = d.contabilizados || 0;
    el("kpiDuplicados").textContent = d.duplicados || 0;
    el("kpiTotalPendiente").textContent = money(d.total_pendiente || 0);
    el("kpiTotalAprobado").textContent = money(d.total_aprobado || 0);
    el("kpiConfianza").textContent = pct(d.confianza_promedio || 0);
  }

  function renderPipeline() {
    var d = state.dashboard || {};
    var stages = [
      ["pendientes", "Pendientes", "Radicados o extraidos sin aprobar"],
      ["en_revision", "Revision", "Datos por validar manualmente"],
      ["aprobados", "Aprobados", "Listos para contabilizar"],
      ["contabilizados", "Contabilizados", "Convertidos en CXP"],
      ["duplicados", "Duplicados", "Bloqueados por control"],
      ["rechazados", "Rechazados", "Descartados por usuario"]
    ];
    el("capturePipeline").innerHTML = stages.map(function (s) {
      return '<article class="capture-stage"><span>' + esc(s[1]) + '</span><strong>' + (d[s[0]] || 0) + '</strong><small>' + esc(s[2]) + '</small></article>';
    }).join("");
  }

  function renderAlerts() {
    var alerts = (state.dashboard.alertas || []).slice();
    if (!alerts.length) alerts.push("Bandeja sin alertas criticas.");
    el("captureAlerts").innerHTML = alerts.map(function (a) {
      var danger = /duplicad|vencid|menor|requieren/i.test(a) ? " is-danger" : "";
      return '<article class="capture-alert' + danger + '"><strong>' + esc(a) + '</strong><p>Revisa el soporte, el documento, proveedor, impuestos y aprobacion antes de contabilizar.</p></article>';
    }).join("");
  }

  function tableHtml(rows) {
    if (!rows.length) return '<div class="capture-empty">No hay soportes para los filtros seleccionados.</div>';
    return '<table class="capture-table"><thead><tr><th>Codigo</th><th>Estado</th><th>Proveedor</th><th>Documento</th><th>Fecha</th><th>Total</th><th>Confianza</th><th>Control</th><th>Archivo</th></tr></thead><tbody>' +
      rows.map(function (s) {
        var selected = state.selected && String(state.selected.id) === String(s.id);
        var archivo = safeHref(s.archivo_url);
        var control = s.duplicado_soporte_id ? "Duplicado #" + s.duplicado_soporte_id : (s.requiere_revision_humana ? "Revision humana" : "OK");
        return '<tr data-id="' + esc(s.id) + '" class="' + (selected ? "is-selected" : "") + '"><td><strong>' + esc(s.codigo || "-") + '</strong><br><span class="capture-muted">' + esc(label(s.tipo_soporte)) + '</span></td><td>' + chip(s.estado_soporte) + '</td><td>' + esc(s.proveedor_nombre || "Sin proveedor") + '<br><span class="capture-muted">' + esc(s.proveedor_nit || "-") + '</span></td><td>' + esc(label(s.documento_tipo)) + '<br><span class="capture-muted">' + esc(s.documento_numero || "-") + '</span></td><td>' + esc(s.fecha_documento || "-") + '<br><span class="capture-muted">Vence ' + esc(s.fecha_vencimiento || "-") + '</span></td><td class="num"><strong>' + money(s.total || 0) + '</strong></td><td>' + progress(s.confianza_ia || 0) + '</td><td>' + esc(control) + '</td><td>' + (archivo ? '<a href="' + esc(archivo) + '" target="_blank" rel="noopener">Ver</a>' : '<span class="capture-muted">Manual</span>') + '</td></tr>';
      }).join("") + '</tbody></table>';
  }

  function renderTables() {
    var rows = rowList();
    el("captureTable").innerHTML = tableHtml(rows);
    el("captureRecentTable").innerHTML = tableHtml((state.dashboard.soportes_recientes || []).slice(0, 8));
  }

  function renderDetail() {
    var s = state.selected;
    if (!s) {
      el("captureDetail").innerHTML = '<div class="capture-empty">Selecciona un soporte en la bandeja para revisar datos extraidos, totales y controles.</div>';
      el("captureEvents").innerHTML = '<div class="capture-empty">Sin soporte seleccionado.</div>';
      return;
    }
    var fields = [
      ["Codigo", s.codigo],
      ["Estado", label(s.estado_soporte)],
      ["Tipo", label(s.tipo_soporte)],
      ["Proveedor", s.proveedor_nombre],
      ["NIT", s.proveedor_nit],
      ["Documento", [label(s.documento_tipo), s.documento_numero].filter(Boolean).join(" ")],
      ["Fecha", s.fecha_documento],
      ["Vencimiento", s.fecha_vencimiento],
      ["Subtotal", money(s.subtotal || 0)],
      ["IVA", money(s.impuesto_iva || 0)],
      ["Retenciones", money((Number(s.retencion_fuente) || 0) + (Number(s.retencion_ica) || 0) + (Number(s.retencion_iva) || 0))],
      ["Total", money(s.total || 0)],
      ["Categoria", s.categoria_contable],
      ["Centro costo", s.centro_costo],
      ["Inventario", s.impacta_inventario ? "Si impacta" : "No impacta"],
      ["Confianza", pct(s.confianza_ia || 0)],
      ["Modelo IA", s.modelo_ia],
      ["Aprobado por", s.aprobado_por],
      ["Observaciones", s.observaciones]
    ];
    el("captureDetail").innerHTML = fields.map(function (f) {
      return '<div class="capture-detail-item"><span>' + esc(f[0]) + '</span><strong>' + esc(f[1] || "-") + '</strong></div>';
    }).join("");
  }

  function renderEvents(events) {
    events = events || [];
    el("captureEvents").innerHTML = events.length ? events.map(function (e) {
      return '<div class="capture-detail-item"><span>' + esc(label(e.evento)) + '</span><strong>' + esc([label(e.estado_anterior), label(e.estado_nuevo)].filter(Boolean).join(" -> ") || "-") + '<br><span class="capture-muted">' + esc(e.fecha_creacion || "") + " - " + esc(e.usuario_creador || e.usuario || "-") + '</span></strong></div>';
    }).join("") : '<div class="capture-empty">Sin eventos registrados para este soporte.</div>';
  }

  function renderChecklist() {
    var d = state.dashboard || {};
    var items = [
      ["Empresa activa", state.empresaId ? "Contexto #" + state.empresaId : "Falta seleccionar empresa"],
      ["IA recomendada", d.modelo_recomendado || "openai:gpt-5.5"],
      ["Revision humana", (d.requieren_revision || 0) + " soporte(s)"],
      ["Inventario pendiente", (d.inventario_pendiente || 0) + " soporte(s)"],
      ["Vencidos / por vencer", (d.vencidos || 0) + " / " + (d.por_vencer || 0)]
    ];
    el("captureChecklist").innerHTML = items.map(function (i) {
      return '<div class="capture-detail-item"><span>' + esc(i[0]) + '</span><strong>' + esc(i[1]) + '</strong></div>';
    }).join("");
  }

  function renderAll() {
    fillKpis(state.dashboard || {});
    renderPipeline();
    renderAlerts();
    renderTables();
    renderDetail();
    renderChecklist();
  }

  function selectSoporte(id) {
    state.selected = state.soportes.find(function (s) { return String(s.id) === String(id); }) || null;
    renderTables();
    renderDetail();
    if (!state.selected) {
      renderEvents([]);
      return Promise.resolve();
    }
    return api("eventos", null, "&soporte_id=" + encodeURIComponent(state.selected.id)).then(function (data) {
      renderEvents(data.eventos || []);
    }).catch(function (e) {
      el("captureEvents").innerHTML = '<div class="capture-empty">' + esc(e.message) + '</div>';
    });
  }

  function load() {
    setBusy(true, "Cargando captura inteligente...");
    return Promise.all([
      api("dashboard"),
      api("soportes", null, state.filters.estado ? "&estado=" + encodeURIComponent(state.filters.estado) : "")
    ]).then(function (res) {
      state.dashboard = res[0].dashboard || {};
      state.soportes = res[1].soportes || [];
      if (state.selected) {
        var current = state.soportes.find(function (s) { return String(s.id) === String(state.selected.id); });
        state.selected = current || null;
      }
      if (!state.selected && state.soportes.length) state.selected = state.soportes[0];
      renderAll();
      if (state.selected) {
        return selectSoporte(state.selected.id).then(function () {
          msg("Captura inteligente actualizada.");
        });
      }
      msg("Captura inteligente actualizada.");
    }).catch(function (e) {
      msg(e.message, true);
    }).finally(function () {
      setBusy(false);
    });
  }

  function submitForm(ev) {
    ev.preventDefault();
    setBusy(true, "Radicando soporte...");
    api("radicar", { method: "POST", body: new FormData(el("formSoporte")) }).then(function () {
      el("formSoporte").reset();
      msg("Soporte radicado correctamente.");
      document.querySelector('[data-tab="bandeja"]').click();
      return load();
    }).catch(function (e) {
      msg(e.message, true);
    }).finally(function () {
      setBusy(false);
    });
  }

  function actionSelected(action, text) {
    if (!state.selected) {
      msg("Selecciona un soporte primero.", true);
      return;
    }
    setBusy(true, text);
    api(action, { method: "POST", body: JSON.stringify({ soporte_id: state.selected.id }) }).then(function () {
      msg("Accion completada.");
      return load();
    }).catch(function (e) {
      msg(e.message, true);
    }).finally(function () {
      setBusy(false);
    });
  }

  function seedDemo() {
    setBusy(true, "Cargando soporte de demostracion...");
    api("seed_demo", { method: "POST", body: "{}" }).then(function () {
      msg("Soporte demo creado.");
      return load();
    }).catch(function (e) {
      msg(e.message, true);
    }).finally(function () {
      setBusy(false);
    });
  }

  function exportCSV() {
    var rows = [["ID", "Codigo", "Estado", "Tipo", "Proveedor", "NIT", "Documento", "Fecha", "Vencimiento", "Total", "Confianza"]].concat(rowList().map(function (s) {
      return [s.id, s.codigo, s.estado_soporte, s.tipo_soporte, s.proveedor_nombre, s.proveedor_nit, s.documento_numero, s.fecha_documento, s.fecha_vencimiento, s.total, s.confianza_ia];
    }));
    var csv = rows.map(function (row) {
      return row.map(function (v) { return '"' + String(v == null ? "" : v).replace(/"/g, '""') + '"'; }).join(";");
    }).join("\n");
    var blob = new Blob([csv], { type: "text/csv;charset=utf-8" });
    var url = URL.createObjectURL(blob);
    var a = document.createElement("a");
    a.href = url;
    a.download = "soportes_compras_ia.csv";
    document.body.appendChild(a);
    a.click();
    a.remove();
    setTimeout(function () { URL.revokeObjectURL(url); }, 1000);
  }

  function safeHref(raw) {
    raw = String(raw || "").trim();
    if (!raw) return "";
    try {
      var url = new URL(raw, window.location.origin);
      if ((url.protocol === "http:" || url.protocol === "https:") && (url.origin === window.location.origin || raw.indexOf("/") !== 0)) return url.href;
      if (raw.indexOf("/") === 0 && url.origin === window.location.origin) return url.pathname + url.search + url.hash;
    } catch (_) {}
    return "";
  }

  document.addEventListener("click", function (ev) {
    var tab = ev.target.closest("[data-tab]");
    if (tab) {
      document.querySelectorAll(".capture-tab").forEach(function (b) { b.classList.toggle("is-active", b === tab); });
      document.querySelectorAll(".capture-panel").forEach(function (p) { p.classList.toggle("is-active", p.id === "tab-" + tab.dataset.tab); });
      return;
    }
    var row = ev.target.closest("tr[data-id]");
    if (row) {
      selectSoporte(row.getAttribute("data-id"));
      document.querySelector('[data-tab="detalle"]').click();
    }
  });

  el("formSoporte").addEventListener("submit", submitForm);
  el("btnLimpiar").addEventListener("click", function () { el("formSoporte").reset(); });
  el("captureRefresh").addEventListener("click", load);
  el("captureSeed").addEventListener("click", seedDemo);
  el("captureExport").addEventListener("click", exportCSV);
  el("btnExtraer").addEventListener("click", function () { actionSelected("extraer_ia", "Extrayendo datos con IA..."); });
  el("btnAprobar").addEventListener("click", function () { actionSelected("aprobar", "Aprobando soporte..."); });
  el("btnRechazar").addEventListener("click", function () { actionSelected("rechazar", "Rechazando soporte..."); });
  el("btnContabilizar").addEventListener("click", function () { actionSelected("contabilizar", "Generando cuenta por pagar..."); });
  el("estadoFilter").addEventListener("change", function () { state.filters.estado = this.value; load(); });
  el("tipoFilter").addEventListener("change", function () { state.filters.tipo = this.value; renderTables(); });
  el("searchFilter").addEventListener("input", function () { state.filters.search = this.value || ""; renderTables(); });

  load();
})();
