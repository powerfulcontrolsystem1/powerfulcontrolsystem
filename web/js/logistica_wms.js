(function () {
  "use strict";

  var state = {
    empresaId: resolveEmpresaId(),
    dashboard: {},
    selectedOrden: null,
    filters: { tipo: "", estado: "", search: "", ubicacion: "" },
    loading: false
  };

  var labelMap = {
    borrador: "Borrador",
    liberada: "Liberada",
    en_picking: "En picking",
    en_packing: "En packing",
    lista_despacho: "Lista despacho",
    despachada: "Despachada",
    cerrada: "Cerrada",
    cancelada: "Cancelada",
    activa: "Activa",
    inactiva: "Inactiva",
    programado: "Programado",
    en_ruta: "En ruta",
    entregado: "Entregado",
    devuelto: "Devuelto",
    pendiente: "Pendiente",
    pickeado: "Pickeado",
    empacado: "Empacado",
    completado: "Completado",
    urgente: "Urgente"
  };
  var numFmt = new Intl.NumberFormat("es-CO", { maximumFractionDigits: 2 });
  var moneyFmt = new Intl.NumberFormat("es-CO", { style: "currency", currency: "COP", maximumFractionDigits: 0 });

  function el(id) { return document.getElementById(id); }
  function val(id) { return (el(id) && el(id).value || "").trim(); }
  function num(id) { var n = Number(val(id)); return Number.isFinite(n) ? n : 0; }
  function fmt(v) { return numFmt.format(Number(v) || 0); }
  function money(v) { try { return moneyFmt.format(Number(v) || 0); } catch (_) { return "$" + fmt(v); } }
  function esc(v) { return String(v == null ? "" : v).replace(/[&<>"']/g, function (ch) { return { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[ch]; }); }
  function label(v) { var key = String(v || "").toLowerCase(); return labelMap[key] || String(v || "-").replace(/_/g, " "); }

  function resolveEmpresaId() {
    try {
      if (window.__resolveEmpresaIdContext) return String(window.__resolveEmpresaIdContext() || "");
    } catch (_) {}
    var sources = [location.search];
    try { if (window.parent && window.parent !== window) sources.push(window.parent.location.search); } catch (_) {}
    for (var i = 0; i < sources.length; i += 1) {
      var id = new URLSearchParams(sources[i]).get("empresa_id") || new URLSearchParams(sources[i]).get("id");
      if (id) return id;
    }
    try { return sessionStorage.getItem("empresa_id") || localStorage.getItem("empresa_id") || ""; } catch (_) { return ""; }
  }

  function msg(text, isError) {
    var target = el("wmsMsg");
    if (!target) return;
    target.textContent = text || "";
    target.classList.toggle("is-error", !!isError);
  }

  function setBusy(on, text) {
    state.loading = !!on;
    document.querySelectorAll(".wms-btn").forEach(function (btn) { btn.disabled = !!on; });
    if (text) msg(text);
  }

  function requestedTab() {
    var tab = new URLSearchParams(location.search || "").get("tab") || "dashboard";
    var exists = false;
    document.querySelectorAll(".wms-tab").forEach(function (btn) {
      if (btn.dataset.tab === tab) exists = true;
    });
    return exists ? tab : "dashboard";
  }

  function setTab(name, skipUrl) {
    var selected = null;
    document.querySelectorAll(".wms-tab").forEach(function (btn) {
      var active = btn.dataset.tab === name;
      btn.classList.toggle("is-active", active);
      if (active) selected = btn;
    });
    if (!selected) return false;
    document.querySelectorAll(".wms-panel").forEach(function (panel) {
      panel.classList.toggle("is-active", panel.id === "tab-" + name);
    });
    if (!skipUrl) {
      var url = new URL(location.href);
      url.searchParams.set("tab", name);
      history.replaceState(null, "", url.pathname + url.search + url.hash);
    }
    return true;
  }

  function api(action, options, extra) {
    if (!state.empresaId) return Promise.reject(new Error("empresa_id no disponible"));
    var url = "/api/empresa/logistica_wms?empresa_id=" + encodeURIComponent(state.empresaId) + "&action=" + encodeURIComponent(action) + (extra || "");
    return fetch(url, Object.assign({ credentials: "same-origin", headers: { "Content-Type": "application/json" } }, options || {})).then(function (res) {
      return res.text().then(function (txt) {
        var data = {};
        try { data = txt ? JSON.parse(txt) : {}; } catch (_) { data = { raw: txt }; }
        if (!res.ok) throw new Error(data.error || data.raw || txt || res.statusText);
        return data;
      });
    });
  }

  function chip(value, className) {
    return '<span class="wms-chip ' + (className || estadoClass(value)) + '">' + esc(label(value)) + "</span>";
  }

  function estadoClass(value) {
    var v = String(value || "").toLowerCase();
    if (["cerrada", "entregado", "activa", "completado", "despachada", "lista_despacho"].indexOf(v) >= 0) return "wms-good";
    if (["cancelada", "devuelto", "inactiva", "cancelado"].indexOf(v) >= 0) return "wms-bad";
    return "wms-warn";
  }

  function progressBar(value) {
    var pct = Math.max(0, Math.min(100, Number(value) || 0));
    return '<div class="wms-progress"><span style="width:' + pct.toFixed(2) + '%"></span></div><span class="wms-muted">' + pct.toFixed(1) + "%</span>";
  }

  function rowsOrdenes() {
    var rows = (state.dashboard.ordenes_recientes || []).slice();
    var q = state.filters.search.toLowerCase();
    return rows.filter(function (r) {
      if (state.filters.tipo && r.tipo !== state.filters.tipo) return false;
      if (state.filters.estado && r.estado !== state.filters.estado) return false;
      if (!q) return true;
      return [r.codigo, r.cliente, r.tercero, r.responsable, r.origen_documento].join(" ").toLowerCase().indexOf(q) >= 0;
    });
  }

  function fillKpis(d) {
    el("kpiUbicaciones").textContent = d.ubicaciones_activas || 0;
    el("kpiOrdenes").textContent = d.ordenes_abiertas || 0;
    el("kpiPicking").textContent = d.ordenes_picking || 0;
    el("kpiPacking").textContent = d.ordenes_packing || 0;
    el("kpiListas").textContent = d.ordenes_listas || 0;
    el("kpiRuta").textContent = d.despachos_en_ruta || 0;
    el("kpiPendientes").textContent = fmt(d.unidades_pendientes || 0);
    el("kpiOcupacion").textContent = fmt(d.ocupacion_pct || 0) + "%";
    if (el("readyOrdenes")) el("readyOrdenes").textContent = d.ordenes_abiertas || 0;
    if (el("readyPendientes")) el("readyPendientes").textContent = fmt(d.unidades_pendientes || 0);
    if (el("readyOcupacion")) el("readyOcupacion").textContent = fmt(d.ocupacion_pct || 0) + "%";
  }

  function renderPipeline() {
    var rows = state.dashboard.ordenes_recientes || [];
    var counts = { liberada: 0, en_picking: 0, en_packing: 0, lista_despacho: 0, despachada: 0 };
    rows.forEach(function (r) { if (counts[r.estado] != null) counts[r.estado] += 1; });
    var stages = [
      ["liberada", "Liberadas", "Listas para asignar bodega"],
      ["en_picking", "Picking", "Separacion en proceso"],
      ["en_packing", "Packing", "Empaque y verificacion"],
      ["lista_despacho", "Lista despacho", "Puede salir a ruta"],
      ["despachada", "Despachadas", "Salida registrada"]
    ];
    el("wmsPipeline").innerHTML = stages.map(function (s) {
      return '<article class="wms-stage"><span>' + esc(s[1]) + '</span><strong>' + (counts[s[0]] || 0) + '</strong><small>' + esc(s[2]) + '</small></article>';
    }).join("");
  }

  function renderAlerts() {
    var alerts = (state.dashboard.alertas || []).slice();
    if ((state.dashboard.ordenes_vencidas || 0) > 0) alerts.unshift(state.dashboard.ordenes_vencidas + " orden(es) vencidas por SLA.");
    if ((state.dashboard.ordenes_urgentes || 0) > 0) alerts.unshift(state.dashboard.ordenes_urgentes + " orden(es) urgentes activas.");
    if (!alerts.length) alerts.push("Operacion WMS sin alertas criticas.");
    el("alertasList").innerHTML = alerts.map(function (a) {
      var danger = /vencid|sin ubicaciones|supera/i.test(a) ? " is-danger" : "";
      return '<article class="wms-alert' + danger + '"><strong>' + esc(a) + '</strong><p>Revisa la cola operativa, capacidad, responsables y eventos recientes.</p></article>';
    }).join("");
  }

  function renderOrdenes() {
    var rows = rowsOrdenes();
    var html = rows.length ? '<table class="wms-table"><thead><tr><th>ID</th><th>Orden</th><th>Tipo</th><th>Cliente</th><th>Compromiso</th><th>Items</th><th>Picking</th><th>Packing</th><th>Estado</th><th>Acciones</th></tr></thead><tbody>' +
      rows.map(function (r) {
        return '<tr><td>' + r.id + '</td><td><strong>' + esc(r.codigo) + '</strong><br><span class="wms-muted">' + esc(r.origen_documento || "-") + '</span></td><td>' + chip(r.tipo, "") + '</td><td>' + esc(r.cliente || r.tercero || "-") + '</td><td>' + esc(r.fecha_compromiso || "-") + '</td><td class="num">' + (r.total_items || 0) + " / " + fmt(r.total_unidades || 0) + '</td><td>' + progressBar(r.progreso_picking || 0) + '</td><td>' + progressBar(r.progreso_packing || 0) + '</td><td>' + chip(r.estado) + '</td><td><div class="wms-row-actions"><button class="wms-btn small" data-edit-orden="' + r.id + '" type="button">Gestionar</button><button class="wms-btn small" data-select-orden="' + r.id + '" type="button">Items</button></div></td></tr>';
      }).join("") + '</tbody></table>' : '<div class="wms-empty">No hay ordenes para los filtros seleccionados.</div>';
    el("ordenesTable").innerHTML = html;
    el("ordenesResumen").innerHTML = html;
    fillOrdenSelects();
  }

  function renderUbicaciones() {
    var rows = (state.dashboard.ubicaciones || []).filter(function (r) { return !state.filters.ubicacion || r.estado === state.filters.ubicacion; });
    el("ubicacionesTable").innerHTML = rows.length ? '<table class="wms-table"><thead><tr><th>Codigo</th><th>Bodega</th><th>Posicion</th><th>Tipo</th><th>Capacidad</th><th>Ocupacion</th><th>%</th><th>Estado</th><th></th></tr></thead><tbody>' +
      rows.map(function (r) {
        var pct = r.capacidad > 0 ? Math.min(100, (r.ocupacion / r.capacidad) * 100) : 0;
        return '<tr><td><strong>' + esc(r.codigo) + '</strong></td><td>' + esc(r.bodega) + '</td><td>' + esc([r.zona, r.pasillo, r.rack, r.nivel, r.posicion].filter(Boolean).join(" / ")) + '</td><td>' + chip(r.tipo, "") + '</td><td class="num">' + fmt(r.capacidad) + '</td><td class="num">' + fmt(r.ocupacion) + '</td><td>' + progressBar(pct) + '</td><td>' + chip(r.estado) + '</td><td><button class="wms-btn small" data-edit-ubi="' + esc(r.codigo) + '" type="button">Editar</button></td></tr>';
      }).join("") + '</tbody></table>' : '<div class="wms-empty">Sin ubicaciones WMS para el filtro seleccionado.</div>';
  }

  function renderItems() {
    var rows = (state.selectedOrden && state.selectedOrden.items) || [];
    if (!rows.length) {
      el("itemsTable").innerHTML = '<div class="wms-empty">Selecciona una orden para ver sus items reales, o agrega items a una orden abierta.</div>';
      return;
    }
    el("itemsTable").innerHTML = '<table class="wms-table"><thead><tr><th>ID</th><th>Producto</th><th>Ubicaciones</th><th>Lote/Serial</th><th>Solicitado</th><th>Pickeado</th><th>Empacado</th><th>Estado</th><th>Acciones</th></tr></thead><tbody>' +
      rows.map(function (i) {
        return '<tr><td>' + i.id + '</td><td><strong>' + esc(i.producto_nombre) + '</strong><br><span class="wms-muted">' + esc(i.sku || "-") + '</span></td><td>' + esc((i.ubicacion_origen || "-") + " -> " + (i.ubicacion_destino || "-")) + '</td><td>' + esc([i.lote, i.serial].filter(Boolean).join(" / ") || "-") + '</td><td class="num">' + fmt(i.cantidad_solicitada) + '</td><td class="num">' + fmt(i.cantidad_pickeada) + '</td><td class="num">' + fmt(i.cantidad_empacada) + '</td><td>' + chip(i.estado) + '</td><td><button class="wms-btn small" data-avance-item="' + i.id + '" data-avance-pick="' + (i.cantidad_pickeada || 0) + '" data-avance-pack="' + (i.cantidad_empacada || 0) + '" type="button">Avance</button></td></tr>';
      }).join("") + '</tbody></table>';
  }

  function renderDespachos() {
    var rows = state.dashboard.despachos_recientes || [];
    el("despachosTable").innerHTML = rows.length ? '<table class="wms-table"><thead><tr><th>ID</th><th>Codigo</th><th>Orden</th><th>Transportadora</th><th>Guia</th><th>Ruta</th><th>Flete</th><th>Salida</th><th>Estado</th></tr></thead><tbody>' +
      rows.map(function (r) { return '<tr><td>' + r.id + '</td><td><strong>' + esc(r.codigo) + '</strong></td><td>' + esc(r.orden_id) + '</td><td>' + esc(r.transportadora || "-") + '</td><td>' + esc(r.guia || "-") + '</td><td>' + esc(r.ruta || "-") + '</td><td class="num">' + money(r.costo_flete || 0) + '</td><td>' + esc(r.fecha_salida || "-") + '</td><td>' + chip(r.estado) + '</td></tr>'; }).join("") + '</tbody></table>' : '<div class="wms-empty">Sin despachos registrados.</div>';
  }

  function renderEventos() {
    var rows = state.dashboard.eventos_recientes || [];
    el("eventosTable").innerHTML = rows.length ? '<table class="wms-table"><thead><tr><th>Fecha</th><th>Referencia</th><th>Evento</th><th>Estado</th><th>Detalle</th><th>Usuario</th></tr></thead><tbody>' +
      rows.map(function (r) { return '<tr><td>' + esc(r.fecha_creacion || "-") + '</td><td>' + esc(r.referencia_tipo + " #" + r.referencia_id) + '</td><td>' + chip(r.evento, "") + '</td><td>' + esc([r.estado_anterior, r.estado_nuevo].filter(Boolean).join(" -> ")) + '</td><td>' + esc(r.detalle || "-") + '</td><td>' + esc(r.usuario || "-") + '</td></tr>'; }).join("") + '</tbody></table>' : '<div class="wms-empty">Sin eventos WMS.</div>';
  }

  function renderAll() {
    fillKpis(state.dashboard || {});
    renderPipeline();
    renderAlerts();
    renderOrdenes();
    renderUbicaciones();
    renderDespachos();
    renderEventos();
    renderItems();
  }

  function fillOrdenSelects() {
    var open = (state.dashboard.ordenes_recientes || []).filter(function (o) { return o.estado !== "cerrada" && o.estado !== "cancelada"; });
    var opts = open.map(function (o) { return '<option value="' + o.id + '">' + esc(o.codigo) + " - " + esc(o.cliente || o.tercero || o.tipo) + '</option>'; }).join("");
    el("itemOrdenId").innerHTML = opts || '<option value="">Sin ordenes abiertas</option>';
    el("desOrdenId").innerHTML = opts || '<option value="">Sin ordenes abiertas</option>';
    if (state.selectedOrden) {
      el("itemOrdenId").value = state.selectedOrden.id || "";
      el("desOrdenId").value = state.selectedOrden.id || "";
    }
  }

  function clearOrdenForm() {
    el("ordenForm").reset();
    el("ordenId").value = "";
    el("ordenTitle").textContent = "Orden WMS";
    el("ordenTipo").value = "picking";
    el("ordenPrioridad").value = "normal";
    el("ordenEstado").value = "borrador";
  }

  function fillOrden(r) {
    el("ordenId").value = r.id || 0;
    el("ordenTitle").textContent = "Orden WMS " + (r.codigo || "");
    el("ordenCodigo").value = r.codigo || "";
    el("ordenTipo").value = r.tipo || "picking";
    el("ordenOrigen").value = r.origen_documento || "";
    el("ordenCliente").value = r.cliente || r.tercero || "";
    el("ordenFecha").value = (r.fecha_compromiso || "").slice(0, 10);
    el("ordenPrioridad").value = r.prioridad || "normal";
    el("ordenEstado").value = r.estado || "borrador";
    el("ordenResponsable").value = r.responsable || "";
    el("ordenObs").value = r.observaciones || "";
    el("itemOrdenId").value = r.id || "";
    el("desOrdenId").value = r.id || "";
  }

  function clearUbicacionForm() {
    el("ubicacionForm").reset();
    el("ubiBodega").value = "Principal";
    el("ubiTipo").value = "almacenamiento";
    el("ubiEstado").value = "activa";
    el("ubiCapacidad").value = "0";
    el("ubiOcupacion").value = "0";
  }

  function loadDetalle(id) {
    return api("detalle", null, "&id=" + encodeURIComponent(id)).then(function (row) {
      state.selectedOrden = row || null;
      if (row) fillOrden(row);
      renderItems();
      fillOrdenSelects();
      msg("Detalle de orden cargado.");
      return row;
    });
  }

  function load() {
    setBusy(true, "Cargando WMS...");
    return api("dashboard").then(function (d) {
      state.dashboard = d || {};
      renderAll();
      msg("WMS actualizado.");
    }).catch(function (e) {
      msg(e.message, true);
    }).finally(function () {
      setBusy(false);
    });
  }

  function exportCSV() {
    var rows = [["ID", "Orden", "Tipo", "Estado", "Cliente", "Items", "Unidades", "Picking", "Packing", "Compromiso"]].concat((state.dashboard.ordenes_recientes || []).map(function (r) {
      return [r.id, r.codigo, r.tipo, r.estado, r.cliente || r.tercero || "", r.total_items, r.total_unidades, r.progreso_picking, r.progreso_packing, r.fecha_compromiso || ""];
    }));
    var csv = rows.map(function (row) {
      return row.map(function (v) { return '"' + String(v == null ? "" : v).replace(/"/g, '""') + '"'; }).join(";");
    }).join("\n");
    var blob = new Blob([csv], { type: "text/csv;charset=utf-8" });
    var url = URL.createObjectURL(blob);
    var a = document.createElement("a");
    a.href = url;
    a.download = "logistica_wms_ordenes.csv";
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
    var oid = ev.target.getAttribute("data-edit-orden") || ev.target.getAttribute("data-select-orden");
    if (oid) {
      var o = (state.dashboard.ordenes_recientes || []).find(function (x) { return String(x.id) === String(oid); });
      if (o) fillOrden(o);
      loadDetalle(oid).then(function () { document.querySelector('[data-tab="items"]').click(); }).catch(function (e) { msg(e.message, true); });
      return;
    }
    var aid = ev.target.getAttribute("data-avance-item");
    if (aid) {
      el("avanceItemId").value = aid;
      el("avancePick").value = ev.target.getAttribute("data-avance-pick") || 0;
      el("avancePack").value = ev.target.getAttribute("data-avance-pack") || 0;
      document.querySelector('[data-tab="items"]').click();
      return;
    }
    var uc = ev.target.getAttribute("data-edit-ubi");
    if (uc) {
      var u = (state.dashboard.ubicaciones || []).find(function (x) { return x.codigo === uc; });
      if (u) {
        el("ubiCodigo").value = u.codigo || "";
        el("ubiBodega").value = u.bodega || "Principal";
        el("ubiZona").value = u.zona || "";
        el("ubiPasillo").value = u.pasillo || "";
        el("ubiRack").value = u.rack || "";
        el("ubiNivel").value = u.nivel || "";
        el("ubiPosicion").value = u.posicion || "";
        el("ubiTipo").value = u.tipo || "almacenamiento";
        el("ubiCapacidad").value = u.capacidad || 0;
        el("ubiOcupacion").value = u.ocupacion || 0;
        el("ubiEstado").value = u.estado || "activa";
        el("ubiObs").value = u.observaciones || "";
      }
    }
  });

  el("wmsRefresh").addEventListener("click", load);
  el("wmsSeed").addEventListener("click", function () {
    setBusy(true, "Cargando demo WMS...");
    api("seed_demo", { method: "POST", body: "{}" }).then(load).catch(function (e) { msg(e.message, true); }).finally(function () { setBusy(false); });
  });
  el("wmsExport").addEventListener("click", exportCSV);
  el("ordenClear").addEventListener("click", clearOrdenForm);
  el("ubiClear").addEventListener("click", clearUbicacionForm);
  el("ubicacionFilter").addEventListener("change", function () { state.filters.ubicacion = this.value; renderUbicaciones(); });
  el("ordenTipoFilter").addEventListener("change", function () { state.filters.tipo = this.value; renderOrdenes(); });
  el("ordenEstadoFilter").addEventListener("change", function () { state.filters.estado = this.value; renderOrdenes(); });
  el("ordenSearch").addEventListener("input", function () { state.filters.search = this.value || ""; renderOrdenes(); });

  el("ubicacionForm").addEventListener("submit", function (ev) {
    ev.preventDefault();
    api("ubicacion", { method: "POST", body: JSON.stringify({ codigo: val("ubiCodigo"), bodega: val("ubiBodega"), zona: val("ubiZona"), pasillo: val("ubiPasillo"), rack: val("ubiRack"), nivel: val("ubiNivel"), posicion: val("ubiPosicion"), tipo: val("ubiTipo"), capacidad: num("ubiCapacidad"), ocupacion: num("ubiOcupacion"), estado: val("ubiEstado"), observaciones: val("ubiObs") }) }).then(load).catch(function (e) { msg(e.message, true); });
  });
  el("ordenForm").addEventListener("submit", function (ev) {
    ev.preventDefault();
    api("orden", { method: num("ordenId") ? "PUT" : "POST", body: JSON.stringify({ id: num("ordenId"), codigo: val("ordenCodigo"), tipo: val("ordenTipo"), origen_documento: val("ordenOrigen"), cliente: val("ordenCliente"), tercero: val("ordenCliente"), fecha_compromiso: val("ordenFecha"), prioridad: val("ordenPrioridad"), responsable: val("ordenResponsable"), estado: val("ordenEstado"), observaciones: val("ordenObs") }) }).then(load).catch(function (e) { msg(e.message, true); });
  });
  el("itemForm").addEventListener("submit", function (ev) {
    ev.preventDefault();
    var orderID = Number(val("itemOrdenId"));
    api("item", { method: "POST", body: JSON.stringify({ orden_id: orderID, producto_nombre: val("itemProducto"), sku: val("itemSku"), ubicacion_origen: val("itemOrigen"), ubicacion_destino: val("itemDestino"), lote: val("itemLote"), serial: val("itemSerial"), cantidad_solicitada: num("itemCantidad"), estado: "pendiente" }) }).then(function () { return load().then(function () { if (orderID) return loadDetalle(orderID); }); }).catch(function (e) { msg(e.message, true); });
  });
  el("avanceForm").addEventListener("submit", function (ev) {
    ev.preventDefault();
    var orderID = Number(val("itemOrdenId"));
    api("avance_item", { method: "POST", body: JSON.stringify({ id: num("avanceItemId"), cantidad_pickeada: num("avancePick"), cantidad_empacada: num("avancePack"), estado: val("avanceEstado") }) }).then(function () { return load().then(function () { if (orderID) return loadDetalle(orderID); }); }).catch(function (e) { msg(e.message, true); });
  });
  el("despachoForm").addEventListener("submit", function (ev) {
    ev.preventDefault();
    api("despacho", { method: "POST", body: JSON.stringify({ orden_id: Number(val("desOrdenId")), codigo: val("desCodigo"), transportadora: val("desTransportadora"), guia: val("desGuia"), conductor: val("desConductor"), vehiculo: val("desVehiculo"), ruta: val("desRuta"), estado: val("desEstado"), fecha_salida: val("desSalida"), fecha_entrega: val("desEntrega"), costo_flete: num("desFlete") }) }).then(load).catch(function (e) { msg(e.message, true); });
  });

  clearOrdenForm();
  clearUbicacionForm();
  setTab(requestedTab(), true);
  load();
})();
