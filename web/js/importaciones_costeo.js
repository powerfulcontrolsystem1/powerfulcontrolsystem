(function () {
  "use strict";

  var state = {
    empresaId: resolveEmpresaId(),
    dashboard: {},
    importaciones: [],
    selected: null,
    detalle: null,
    filters: { estado: "", search: "" },
    loading: false
  };

  var labels = {
    borrador: "Borrador",
    en_transito: "En transito",
    costeado: "Costeado",
    cerrado: "Cerrado",
    contabilizado: "Contabilizado",
    anulado: "Anulado",
    flete: "Flete",
    seguro: "Seguro",
    arancel: "Arancel",
    iva: "IVA",
    aduana: "Aduana",
    bodega: "Bodega",
    transporte: "Transporte interno",
    nacionalizacion: "Nacionalizacion",
    valor: "Valor",
    peso: "Peso",
    volumen: "Volumen",
    cantidad: "Cantidad"
  };
  var moneyFmt = new Intl.NumberFormat("es-CO", { style: "currency", currency: "COP", maximumFractionDigits: 0 });
  var numFmt = new Intl.NumberFormat("es-CO", { maximumFractionDigits: 2 });

  function el(id) { return document.getElementById(id); }
  function val(id) { return (el(id) && el(id).value || "").trim(); }
  function num(id) { var n = Number(val(id)); return Number.isFinite(n) ? n : 0; }
  function money(v) { try { return moneyFmt.format(Number(v) || 0); } catch (_) { return "$" + (Number(v) || 0).toFixed(0); } }
  function fmt(v) { return numFmt.format(Number(v) || 0); }
  function esc(v) { return String(v == null ? "" : v).replace(/[&<>"']/g, function (ch) { return { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[ch]; }); }
  function label(v) { return labels[String(v || "").toLowerCase()] || String(v || "-").replace(/_/g, " "); }
  function q(name) { return new URLSearchParams(location.search || "").get(name) || ""; }

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

  function today() {
    return new Date().toISOString().slice(0, 10);
  }

  function msg(text, isError) {
    var target = el("impMsg");
    if (!target) return;
    target.textContent = text || "";
    target.classList.toggle("is-error", !!isError);
  }

  function setBusy(on, text) {
    state.loading = !!on;
    document.querySelectorAll(".imp-btn").forEach(function (btn) { btn.disabled = !!on; });
    if (text) msg(text);
  }

  function api(action, options, extra) {
    if (!state.empresaId) return Promise.reject(new Error("empresa_id no disponible"));
    var url = "/api/empresa/importaciones_costeo?empresa_id=" + encodeURIComponent(state.empresaId) + "&action=" + encodeURIComponent(action) + (extra || "");
    return fetch(url, Object.assign({ credentials: "same-origin", headers: { "Content-Type": "application/json" } }, options || {})).then(function (res) {
      return res.text().then(function (txt) {
        var data = {};
        try { data = txt ? JSON.parse(txt) : {}; } catch (_) { data = { raw: txt }; }
        if (!res.ok || data.ok === false) throw new Error(data.error || data.raw || txt || res.statusText);
        return data;
      });
    });
  }

  function chip(value) {
    return '<span class="imp-chip ' + estadoClass(value) + '">' + esc(label(value)) + "</span>";
  }

  function estadoClass(value) {
    var v = String(value || "").toLowerCase();
    if (v === "cerrado" || v === "contabilizado" || v === "costeado") return "imp-good";
    if (v === "anulado") return "imp-bad";
    return "imp-warn";
  }

  function progressPct(value) {
    var pct = Math.max(0, Math.min(100, Number(value) || 0));
    return '<div class="imp-progress"><span style="width:' + pct.toFixed(2) + '%"></span></div><span class="imp-muted">' + pct.toFixed(1) + "%</span>";
  }

  function rowsImportaciones() {
    var q = state.filters.search.toLowerCase();
    return state.importaciones.filter(function (r) {
      if (state.filters.estado && r.estado !== state.filters.estado) return false;
      if (!q) return true;
      return [r.codigo, r.proveedor, r.pais_origen, r.documento_referencia, r.incoterm, r.moneda_origen].join(" ").toLowerCase().indexOf(q) >= 0;
    });
  }

  function fillKpis(d) {
    el("kpiAbiertas").textContent = d.importaciones_abiertas || 0;
    el("kpiTransito").textContent = d.en_transito || 0;
    el("kpiCosteadas").textContent = d.costeadas || 0;
    el("kpiCerradas").textContent = d.importaciones_cerradas || 0;
    el("kpiSubtotal").textContent = money(d.subtotal_cop || 0);
    el("kpiNacionalizacion").textContent = money(d.costos_pendientes_cop || 0);
    el("kpiTotal").textContent = money(d.costo_total_cop || 0);
    el("kpiPct").textContent = fmt(d.nacionalizacion_pct || 0) + "%";
    if (el("miniSinItems")) el("miniSinItems").textContent = d.importaciones_sin_items || 0;
    if (el("miniSinCostos")) el("miniSinCostos").textContent = d.importaciones_sin_costos || 0;
    if (el("miniBorradores")) el("miniBorradores").textContent = d.borradores || 0;
    if (el("miniAnuladas")) el("miniAnuladas").textContent = d.anuladas || 0;
  }

  function setTab(tab) {
    var allowed = { tablero: 1, importaciones: 1, items: 1, costos: 1, costeo: 1 };
    tab = allowed[tab] ? tab : "tablero";
    document.querySelectorAll(".imp-tab").forEach(function (b) { b.classList.toggle("is-active", b.dataset.tab === tab); });
    document.querySelectorAll(".imp-panel").forEach(function (p) { p.classList.toggle("is-active", p.id === "tab-" + tab); });
    try {
      var next = new URL(window.location.href);
      next.searchParams.set("tab", tab);
      window.history.replaceState({}, "", next.toString());
    } catch (_) {}
  }

  function renderPipeline() {
    var d = state.dashboard || {};
    var stages = [
      ["borradores", "Borrador", "Pendiente de completar"],
      ["en_transito", "En transito", "Embarque en proceso"],
      ["costeadas", "Costeada", "Costos distribuidos"],
      ["importaciones_cerradas", "Cerrada", "Lista para inventario"],
      ["contabilizadas", "Contabilizada", "Impacto contable"]
    ];
    el("impPipeline").innerHTML = stages.map(function (s) {
      return '<article class="imp-stage"><span>' + esc(s[1]) + '</span><strong>' + (d[s[0]] || 0) + '</strong><small>' + esc(s[2]) + '</small></article>';
    }).join("");
  }

  function renderAlerts() {
    var alerts = (state.dashboard.alertas || []).slice();
    if (!alerts.length) alerts.push("Importaciones sin alertas criticas.");
    el("impAlerts").innerHTML = alerts.map(function (a) {
      var danger = /superaron|no tienen|sin/i.test(a) ? " is-danger" : "";
      return '<article class="imp-alert' + danger + '"><strong>' + esc(a) + '</strong><p>Revisa items, costos, TRM, base de distribucion y fecha estimada antes de cerrar la importacion.</p></article>';
    }).join("");
  }

  function tableImportaciones(rows) {
    if (!rows.length) return '<div class="imp-empty">No hay importaciones para los filtros seleccionados.</div>';
    return '<table class="imp-table"><thead><tr><th>Codigo</th><th>Proveedor</th><th>Origen</th><th>Fechas</th><th>Subtotal</th><th>Nacionalizacion</th><th>Total</th><th>Estado</th><th>Acciones</th></tr></thead><tbody>' +
      rows.map(function (r) {
        var selected = state.selected && String(state.selected.id) === String(r.id);
        return '<tr data-id="' + esc(r.id) + '" class="' + (selected ? "is-selected" : "") + '"><td><strong>' + esc(r.codigo || "-") + '</strong><br><span class="imp-muted">' + esc(r.documento_referencia || "-") + '</span></td><td>' + esc(r.proveedor || "-") + '</td><td>' + esc(r.pais_origen || "-") + '<br><span class="imp-muted">' + esc(r.incoterm || "FOB") + " / " + esc(r.moneda_origen || "USD") + '</span></td><td>' + esc(r.fecha_documento || "-") + '<br><span class="imp-muted">ETA ' + esc(r.fecha_estimada_llegada || "-") + '</span></td><td class="num">' + money(r.subtotal_cop) + '</td><td class="num">' + money(r.costos_nacionalizacion_cop) + '</td><td class="num"><strong>' + money(r.costo_total_cop) + '</strong></td><td>' + chip(r.estado) + '</td><td><button class="imp-btn small" data-manage="' + esc(r.id) + '" type="button">Gestionar</button></td></tr>';
      }).join("") + '</tbody></table>';
  }

  function renderImportaciones() {
    var rows = rowsImportaciones();
    el("importacionesTable").innerHTML = tableImportaciones(rows);
    el("recentTable").innerHTML = tableImportaciones((state.dashboard.ultimas_importaciones || []).slice(0, 8));
    fillImportacionSelects();
  }

  function renderItems() {
    var rows = (state.detalle && state.detalle.items) || [];
    el("itemsTable").innerHTML = rows.length ? '<table class="imp-table"><thead><tr><th>Producto</th><th>Cantidad</th><th>Peso/Volumen</th><th>Costo origen</th><th>Base COP</th><th>Distribuido</th><th>Unitario final</th></tr></thead><tbody>' +
      rows.map(function (r) {
        return '<tr><td><strong>' + esc(r.producto_nombre || "-") + '</strong><br><span class="imp-muted">' + esc(r.sku || "-") + '</span></td><td class="num">' + fmt(r.cantidad) + " " + esc(r.unidad || "und") + '</td><td>' + fmt(r.peso_kg) + ' kg<br><span class="imp-muted">' + fmt(r.volumen_m3) + ' m3</span></td><td class="num">' + fmt(r.costo_origen) + " " + esc((state.detalle && state.detalle.moneda_origen) || "") + '</td><td class="num">' + money(r.costo_base_cop) + '</td><td class="num">' + money(r.costo_distribuido_cop) + '</td><td class="num"><strong>' + money(r.costo_unitario_final_cop) + '</strong></td></tr>';
      }).join("") + '</tbody></table>' : '<div class="imp-empty">Selecciona una importacion y agrega items para calcular costo base COP.</div>';
  }

  function renderCostos() {
    var rows = (state.detalle && state.detalle.costos) || [];
    el("costosTable").innerHTML = rows.length ? '<table class="imp-table"><thead><tr><th>Tipo</th><th>Concepto</th><th>Base</th><th>Valor</th><th>Tercero</th><th>Documento</th><th>Cuenta</th></tr></thead><tbody>' +
      rows.map(function (r) {
        return '<tr><td>' + chip(r.tipo) + '</td><td><strong>' + esc(r.concepto || "-") + '</strong></td><td>' + esc(label(r.base_distribucion)) + '</td><td class="num">' + money(r.valor_cop) + '</td><td>' + esc(r.tercero || "-") + '</td><td>' + esc(r.documento || "-") + '</td><td>' + esc(r.cuenta_contable || "-") + '</td></tr>';
      }).join("") + '</tbody></table>' : '<div class="imp-empty">Registra costos como flete, seguro, aranceles, aduana o transporte interno.</div>';
  }

  function renderLanded() {
    var rows = (state.detalle && state.detalle.items) || [];
    var total = (state.detalle && Number(state.detalle.costo_total_cop)) || 0;
    el("landedTable").innerHTML = rows.length ? '<table class="imp-table"><thead><tr><th>SKU</th><th>Producto</th><th>Base COP</th><th>Costos distribuidos</th><th>% del embarque</th><th>Costo unitario final</th></tr></thead><tbody>' +
      rows.map(function (r) {
        var landed = (Number(r.costo_base_cop) || 0) + (Number(r.costo_distribuido_cop) || 0);
        var share = total > 0 ? landed / total * 100 : 0;
        return '<tr><td><strong>' + esc(r.sku || "-") + '</strong></td><td>' + esc(r.producto_nombre || "-") + '</td><td class="num">' + money(r.costo_base_cop) + '</td><td class="num">' + money(r.costo_distribuido_cop) + '</td><td>' + progressPct(share) + '</td><td class="num"><strong>' + money(r.costo_unitario_final_cop) + '</strong></td></tr>';
      }).join("") + '</tbody></table>' : '<div class="imp-empty">Agrega items y costos, luego distribuye para ver el costo aterrizado por producto.</div>';
  }

  function renderDetail() {
    var d = state.detalle || state.selected;
    if (!d) {
      el("detailBox").innerHTML = '<div class="imp-empty">Selecciona una importacion para ver el resumen.</div>';
      return;
    }
    var fields = [
      ["Codigo", d.codigo],
      ["Proveedor", d.proveedor],
      ["Pais origen", d.pais_origen],
      ["Incoterm", d.incoterm],
      ["Moneda / TRM", (d.moneda_origen || "-") + " / " + fmt(d.trm || 0)],
      ["Documento", d.documento_referencia],
      ["Fecha", d.fecha_documento],
      ["Llegada estimada", d.fecha_estimada_llegada],
      ["Estado", label(d.estado)],
      ["Subtotal origen", fmt(d.subtotal_origen || 0)],
      ["Subtotal COP", money(d.subtotal_cop || 0)],
      ["Nacionalizacion", money(d.costos_nacionalizacion_cop || 0)],
      ["Costo total", money(d.costo_total_cop || 0)]
    ];
    el("detailBox").innerHTML = fields.map(function (f) {
      return '<div class="imp-detail-item"><span>' + esc(f[0]) + '</span><strong>' + esc(f[1] || "-") + '</strong></div>';
    }).join("");
  }

  function renderAll() {
    fillKpis(state.dashboard || {});
    renderPipeline();
    renderAlerts();
    renderImportaciones();
    renderItems();
    renderCostos();
    renderLanded();
    renderDetail();
  }

  function fillImportacionSelects() {
    var options = state.importaciones.map(function (r) {
      return '<option value="' + esc(r.id) + '">' + esc(r.codigo || ("#" + r.id)) + " - " + esc(r.proveedor || label(r.estado)) + '</option>';
    }).join("");
    el("itemImportacion").innerHTML = options || '<option value="">Sin importaciones</option>';
    el("costImportacion").innerHTML = options || '<option value="">Sin importaciones</option>';
    if (state.selected) {
      el("itemImportacion").value = state.selected.id || "";
      el("costImportacion").value = state.selected.id || "";
    }
  }

  function fillImportacionForm(row) {
    if (!row) return;
    el("impCodigo").value = row.codigo || "";
    el("impProveedor").value = row.proveedor || "";
    el("impPais").value = row.pais_origen || "";
    el("impIncoterm").value = row.incoterm || "FOB";
    el("impMoneda").value = row.moneda_origen || "USD";
    el("impTRM").value = row.trm || 1;
    el("impFecha").value = (row.fecha_documento || "").slice(0, 10);
    el("impEta").value = (row.fecha_estimada_llegada || "").slice(0, 10);
    el("impRef").value = row.documento_referencia || "";
    el("impEstado").value = row.estado || "borrador";
  }

  function clearImportacionForm() {
    el("importacionForm").reset();
    el("impIncoterm").value = "FOB";
    el("impMoneda").value = "USD";
    el("impTRM").value = "3900";
    el("impFecha").value = today();
    el("impEstado").value = "en_transito";
  }

  function selectedImportacionID(selectId) {
    var n = Number(val(selectId));
    if (n > 0) return n;
    return state.selected ? Number(state.selected.id) : 0;
  }

  function loadDetalle(id) {
    if (!id) {
      state.selected = null;
      state.detalle = null;
      renderAll();
      return Promise.resolve();
    }
    return api("detalle", null, "&id=" + encodeURIComponent(id)).then(function (row) {
      state.detalle = row || null;
      state.selected = state.importaciones.find(function (r) { return String(r.id) === String(id); }) || row || null;
      fillImportacionForm(state.selected);
      renderAll();
      return row;
    });
  }

  function load() {
    setBusy(true, "Cargando importaciones...");
    var extra = state.filters.estado ? "&estado=" + encodeURIComponent(state.filters.estado) : "";
    return Promise.all([api("dashboard"), api("importaciones", null, extra)]).then(function (res) {
      state.dashboard = res[0] || {};
      state.importaciones = Array.isArray(res[1]) ? res[1] : (res[1].importaciones || []);
      if (state.selected) {
        var current = state.importaciones.find(function (r) { return String(r.id) === String(state.selected.id); });
        state.selected = current || null;
      }
      if (!state.selected && state.importaciones.length) state.selected = state.importaciones[0];
      renderAll();
      if (state.selected) {
        return loadDetalle(state.selected.id).then(function () { msg("Importaciones actualizadas."); });
      }
      msg("Importaciones actualizadas.");
    }).catch(function (e) {
      msg(e.message, true);
    }).finally(function () {
      setBusy(false);
    });
  }

  function post(action, payload) {
    return api(action, { method: "POST", body: JSON.stringify(payload || {}) });
  }

  function saveImportacion(ev) {
    ev.preventDefault();
    setBusy(true, "Guardando importacion...");
    post("importacion", {
      codigo: val("impCodigo"),
      proveedor: val("impProveedor"),
      pais_origen: val("impPais"),
      incoterm: val("impIncoterm"),
      moneda_origen: val("impMoneda"),
      trm: num("impTRM"),
      fecha_documento: val("impFecha"),
      fecha_estimada_llegada: val("impEta"),
      documento_referencia: val("impRef"),
      estado: val("impEstado")
    }).then(function (res) {
      msg("Importacion guardada.");
      return load().then(function () {
        if (res && res.id) return loadDetalle(res.id);
      });
    }).catch(function (e) {
      msg(e.message, true);
    }).finally(function () {
      setBusy(false);
    });
  }

  function saveItem(ev) {
    ev.preventDefault();
    var id = selectedImportacionID("itemImportacion");
    if (!id) {
      msg("Selecciona una importacion para agregar items.", true);
      return;
    }
    setBusy(true, "Agregando item...");
    post("item", {
      importacion_id: id,
      producto_nombre: val("itemNombre"),
      sku: val("itemSKU"),
      unidad: val("itemUnidad") || "und",
      cantidad: num("itemCantidad"),
      costo_unitario_origen: num("itemCosto"),
      peso_kg: num("itemPeso"),
      volumen_m3: num("itemVol")
    }).then(function () {
      el("itemForm").reset();
      el("itemUnidad").value = "und";
      msg("Item agregado.");
      return load().then(function () { return loadDetalle(id); });
    }).catch(function (e) {
      msg(e.message, true);
    }).finally(function () {
      setBusy(false);
    });
  }

  function saveCosto(ev) {
    ev.preventDefault();
    var id = selectedImportacionID("costImportacion");
    if (!id) {
      msg("Selecciona una importacion para agregar costos.", true);
      return;
    }
    setBusy(true, "Agregando costo...");
    post("costo", {
      importacion_id: id,
      tipo: val("costTipo"),
      concepto: val("costConcepto"),
      base_distribucion: val("costBase"),
      valor_cop: num("costValor"),
      tercero: val("costTercero"),
      documento: val("costDocumento"),
      cuenta_contable: val("costCuenta")
    }).then(function () {
      el("costoForm").reset();
      el("costTipo").value = "flete";
      el("costBase").value = "valor";
      msg("Costo agregado.");
      return load().then(function () { return loadDetalle(id); });
    }).catch(function (e) {
      msg(e.message, true);
    }).finally(function () {
      setBusy(false);
    });
  }

  function distribuir() {
    var id = state.selected ? Number(state.selected.id) : 0;
    if (!id) {
      msg("Selecciona una importacion para distribuir costos.", true);
      return;
    }
    setBusy(true, "Distribuyendo costos por base configurada...");
    post("distribuir", { importacion_id: id }).then(function () {
      msg("Costos distribuidos.");
      return load().then(function () { return loadDetalle(id); });
    }).catch(function (e) {
      msg(e.message, true);
    }).finally(function () {
      setBusy(false);
    });
  }

  function seedDemo() {
    setBusy(true, "Cargando demo...");
    post("seed_demo", {}).then(function () {
      msg("Demo cargado.");
      return load();
    }).catch(function (e) {
      msg(e.message, true);
    }).finally(function () {
      setBusy(false);
    });
  }

  function exportCSV() {
    var rows = [["ID", "Codigo", "Estado", "Proveedor", "Pais", "Incoterm", "Moneda", "TRM", "Fecha", "ETA", "Subtotal COP", "Nacionalizacion", "Costo total"]].concat(rowsImportaciones().map(function (r) {
      return [r.id, r.codigo, r.estado, r.proveedor, r.pais_origen, r.incoterm, r.moneda_origen, r.trm, r.fecha_documento, r.fecha_estimada_llegada, r.subtotal_cop, r.costos_nacionalizacion_cop, r.costo_total_cop];
    }));
    var csv = rows.map(function (row) {
      return row.map(function (v) { return '"' + String(v == null ? "" : v).replace(/"/g, '""') + '"'; }).join(";");
    }).join("\n");
    var blob = new Blob([csv], { type: "text/csv;charset=utf-8" });
    var url = URL.createObjectURL(blob);
    var a = document.createElement("a");
    a.href = url;
    a.download = "importaciones_costeo.csv";
    document.body.appendChild(a);
    a.click();
    a.remove();
    setTimeout(function () { URL.revokeObjectURL(url); }, 1000);
  }

  document.addEventListener("click", function (ev) {
    var tab = ev.target.closest("[data-tab]");
    if (tab) {
      setTab(tab.dataset.tab);
      return;
    }
    var go = ev.target.closest("[data-imp-go]");
    if (go) {
      setTab(go.dataset.impGo);
      return;
    }
    var manage = ev.target.getAttribute("data-manage");
    var row = ev.target.closest("tr[data-id]");
    var id = manage || (row && row.getAttribute("data-id"));
    if (id) {
      loadDetalle(id).then(function () {
        if (manage) document.querySelector('[data-tab="costeo"]').click();
      }).catch(function (e) { msg(e.message, true); });
    }
  });

  el("importacionForm").addEventListener("submit", saveImportacion);
  el("itemForm").addEventListener("submit", saveItem);
  el("costoForm").addEventListener("submit", saveCosto);
  el("impClear").addEventListener("click", clearImportacionForm);
  el("impRefresh").addEventListener("click", load);
  el("impSeed").addEventListener("click", seedDemo);
  el("impExport").addEventListener("click", exportCSV);
  el("btnDistribuir").addEventListener("click", distribuir);
  el("estadoFilter").addEventListener("change", function () { state.filters.estado = this.value; load(); });
  el("searchFilter").addEventListener("input", function () { state.filters.search = this.value || ""; renderImportaciones(); });

  clearImportacionForm();
  setTab(q("tab") || q("panel") || "tablero");
  load();
})();
