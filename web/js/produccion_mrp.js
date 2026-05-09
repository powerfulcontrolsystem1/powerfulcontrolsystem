(function () {
  "use strict";

  var empresaId = "";
  try {
    empresaId = (window.__resolveEmpresaIdContext && window.__resolveEmpresaIdContext()) || "";
  } catch (e) {}
  if (!empresaId) {
    empresaId = new URLSearchParams(location.search).get("empresa_id") || new URLSearchParams(location.search).get("id") || "";
  }

  var state = {
    recetas: [],
    ordenes: [],
    consumos: [],
    calidad: [],
    plan: [],
    dashboard: {},
    config: {},
    filters: { recetas: "", ordenes: "", search: "" },
    loading: false
  };
  var fmt = new Intl.NumberFormat("es-CO", { style: "currency", currency: "COP", maximumFractionDigits: 0 });
  var qtyFmt = new Intl.NumberFormat("es-CO", { maximumFractionDigits: 4 });
  var stateLabels = {
    activo: "Activo",
    inactivo: "Inactivo",
    borrador: "Borrador",
    programada: "Programada",
    en_proceso: "En proceso",
    calidad: "Calidad",
    cerrada: "Cerrada",
    cancelada: "Cancelada",
    aprobado: "Aprobado",
    rechazado: "Rechazado",
    reproceso: "Reproceso",
    pendiente: "Pendiente",
    reemplazado: "Reemplazado",
    urgente: "Urgente"
  };

  function $(id) { return document.getElementById(id); }
  function esc(v) { return String(v == null ? "" : v).replace(/[&<>"]/g, function (c) { return ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" })[c]; }); }
  function num(v) { var n = Number(v); return Number.isFinite(n) ? n : 0; }
  function qty(v) { return qtyFmt.format(num(v)); }
  function money(v) { try { return fmt.format(num(v)); } catch (e) { return "$" + num(v).toLocaleString("es-CO"); } }
  function label(v) { return stateLabels[String(v || "").toLowerCase()] || String(v || "Sin estado"); }
  function pct(done, total) { return total > 0 ? Math.max(0, Math.min(100, Math.round((done / total) * 100))) : 0; }
  function todayMonth() { return new Date().toISOString().slice(0, 7); }
  function msg(text, isError) {
    var el = $("prodMsg");
    if (!el) return;
    el.textContent = text || "";
    el.classList.toggle("is-error", !!isError);
  }
  function setBusy(on, text) {
    state.loading = !!on;
    document.querySelectorAll(".prod-btn").forEach(function (btn) {
      if (btn.dataset.allowBusy === "1") return;
      btn.disabled = !!on;
    });
    if (text) msg(text);
  }

  function requestedTab() {
    var tab = new URLSearchParams(location.search || "").get("tab") || "dashboard";
    var exists = false;
    document.querySelectorAll(".prod-tab").forEach(function (btn) {
      if (btn.dataset.tab === tab) exists = true;
    });
    return exists ? tab : "dashboard";
  }

  function setTab(name, skipUrl) {
    var selected = null;
    document.querySelectorAll(".prod-tab").forEach(function (btn) {
      var active = btn.dataset.tab === name;
      btn.classList.toggle("is-active", active);
      if (active) selected = btn;
    });
    if (!selected) return false;
    document.querySelectorAll(".prod-panel").forEach(function (panel) {
      panel.classList.toggle("is-active", panel.id === "tab-" + name);
    });
    if (!skipUrl) {
      var url = new URL(location.href);
      url.searchParams.set("tab", name);
      history.replaceState(null, "", url.pathname + url.search + url.hash);
    }
    return true;
  }

  async function api(action, opts) {
    if (!empresaId) throw new Error("No se encontro empresa_id para Produccion/MRP.");
    var url = "/api/empresa/produccion_mrp?empresa_id=" + encodeURIComponent(empresaId) + (action ? "&action=" + encodeURIComponent(action) : "");
    var res = await fetch(url, Object.assign({ credentials: "same-origin" }, opts || {}));
    var txt = await res.text();
    var data = {};
    try { data = txt ? JSON.parse(txt) : {}; } catch (e) { data = { raw: txt }; }
    if (!res.ok) throw new Error(data.error || data.raw || txt || ("HTTP " + res.status));
    return data;
  }
  function body(payload, method) {
    return { method: method || "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(payload || {}) };
  }

  function chip(value) {
    var raw = String(value || "").toLowerCase();
    return '<span class="prod-chip" data-state="' + esc(raw) + '">' + esc(label(raw)) + '</span>';
  }

  function fillKpis(d) {
    $("kpiRecetas").textContent = d.recetas_activas || 0;
    $("kpiAbiertas").textContent = d.ordenes_abiertas || 0;
    $("kpiProgramadas").textContent = d.ordenes_programadas || 0;
    $("kpiProceso").textContent = d.ordenes_proceso || 0;
    $("kpiCalidad").textContent = d.ordenes_calidad || 0;
    $("kpiUrgentes").textContent = d.ordenes_urgentes || 0;
    $("kpiCostoAbierto").textContent = money(d.costo_estimado_abierto || 0);
    $("kpiCostoMes").textContent = money(d.costo_real_mes || 0);
    if ($("checkRecetas")) $("checkRecetas").textContent = d.recetas_activas || 0;
    if ($("checkUrgentes")) $("checkUrgentes").textContent = d.ordenes_urgentes || 0;
  }

  function fillConfig(c) {
    state.config = c || {};
    $("cfgNombre").value = state.config.nombre_sistema || "Produccion / MRP";
    $("cfgMoneda").value = state.config.moneda || "COP";
    $("cfgCosteo").value = state.config.costeo_modo || "estandar";
    $("cfgAprobar").checked = !!state.config.aprobar_ordenes;
    $("cfgConsumir").checked = !!state.config.consumir_inventario_al_iniciar;
    $("cfgCalidad").checked = !!state.config.cerrar_con_calidad;
  }

  function renderDashboard() {
    var d = state.dashboard || {};
    var counts = { borrador: d.ordenes_borrador || 0, programada: d.ordenes_programadas || 0, en_proceso: d.ordenes_proceso || 0, calidad: d.ordenes_calidad || 0, cerrada: d.ordenes_cerradas || 0 };
    if (!counts.borrador && !counts.programada && !counts.en_proceso && !counts.calidad && !counts.cerrada) {
      state.ordenes.forEach(function (o) { if (counts[o.estado] != null) counts[o.estado] += 1; });
    }
    var stages = [
      ["borrador", "Borrador", "Pendientes de programar"],
      ["programada", "Programadas", "Listas para iniciar"],
      ["en_proceso", "En proceso", "En piso de produccion"],
      ["calidad", "Calidad", "En revision"],
      ["cerrada", "Cerradas", "Terminadas"]
    ];
    $("prodPipeline").innerHTML = stages.map(function (s) {
      return '<article class="prod-stage">' +
        '<span>' + esc(s[1]) + '</span><strong>' + (counts[s[0]] || 0) + '</strong><small>' + esc(s[2]) + '</small>' +
      '</article>';
    }).join("");

    var alerts = (d.alertas || []).map(function (text) {
      return [/vencid|urgente|no hay|componentes|requieren/i.test(text) ? "is-danger" : "", text, "Revisa recetas, ordenes, consumo real, calidad y el plan MRP antes de cerrar la produccion."];
    });
    if (!state.recetas.length) alerts.push(["is-danger", "No hay recetas/BOM", "Registra una receta para poder crear ordenes y generar MRP real."]);
    var sinComponentes = state.recetas.filter(function (r) { return r.estado === "activo" && !(r.componentes || []).length; }).length;
    if (sinComponentes) alerts.push(["", "Recetas sin componentes", sinComponentes + " receta(s) activas no tienen materiales definidos."]);
    var vencidas = state.ordenes.filter(function (o) {
      return o.fecha_programada && o.estado !== "cerrada" && o.estado !== "cancelada" && o.fecha_programada.slice(0, 10) < new Date().toISOString().slice(0, 10);
    }).length;
    if (vencidas) alerts.push(["is-danger", "Ordenes vencidas", vencidas + " orden(es) superaron su fecha programada."]);
    var riesgoMRP = state.plan.filter(function (p) { return num(p.disponible_proyectado) < 0 || num(p.cantidad_sugerida_compra) > 0 || num(p.cantidad_sugerida_producir) > 0; }).length;
    if ($("checkMRP")) $("checkMRP").textContent = riesgoMRP;
    if (riesgoMRP) alerts.push(["", "MRP con requerimientos", riesgoMRP + " linea(s) requieren compra o produccion."]);
    if (!alerts.length) alerts.push(["", "Operacion estable", "No hay alertas criticas con los datos cargados."]);
    $("prodAlerts").innerHTML = alerts.map(function (a) {
      return '<article class="prod-alert ' + a[0] + '"><strong>' + esc(a[1]) + '</strong><p>' + esc(a[2]) + '</p></article>';
    }).join("");

    var recent = state.consumos.slice(0, 8).map(function (c) {
      return ["Consumo", c.producto_nombre, qty(c.cantidad_consumida), money(c.costo_total), c.fecha_consumo || ""];
    }).concat(state.calidad.slice(0, 8).map(function (c) {
      return ["Calidad", "Orden #" + c.orden_id + " - " + label(c.resultado), qty(c.cantidad_aprobada), qty(c.cantidad_rechazada), c.fecha_revision || ""];
    })).slice(0, 12);
    $("prodRecent").innerHTML = '<table class="prod-table"><thead><tr><th>Tipo</th><th>Detalle</th><th>Cantidad</th><th>Valor/Resultado</th><th>Fecha</th></tr></thead><tbody>' +
      (recent.map(function (r) { return '<tr><td>' + esc(r[0]) + '</td><td>' + esc(r[1]) + '</td><td class="num">' + esc(r[2]) + '</td><td class="num">' + esc(r[3]) + '</td><td>' + esc(r[4]) + '</td></tr>'; }).join("") || '<tr><td colspan="5">Sin actividad reciente.</td></tr>') +
      '</tbody></table>';
  }

  function recetaToComponentLines(r) {
    return (r.componentes || []).map(function (c) {
      return [c.producto_nombre || "", c.cantidad || 0, c.unidad || "und", c.costo_unitario || 0, c.merma_porcentaje || 0, c.etapa || "produccion"].join("|");
    }).join("\n") || "Materia prima principal|1|und|0|0|produccion";
  }

  function clearRecetaForm() {
    $("recetaForm").reset();
    $("recId").value = "";
    $("recVersion").value = "1.0";
    $("recUnidad").value = "und";
    $("recCantidad").value = "1";
    $("recComponentes").value = "Materia prima principal|1|und|0|0|produccion";
    $("recSubmit").textContent = "Guardar receta";
  }

  function editReceta(id) {
    var r = state.recetas.find(function (x) { return String(x.id) === String(id); });
    if (!r) return;
    $("recId").value = r.id || "";
    $("recCodigo").value = r.codigo || "";
    $("recVersion").value = r.version || "1.0";
    $("recNombre").value = r.nombre || "";
    $("recProducto").value = r.producto_terminado_nombre || r.nombre || "";
    $("recUnidad").value = r.unidad || "und";
    $("recCantidad").value = r.cantidad_base || 1;
    $("recCosto").value = r.costo_estandar || 0;
    $("recMerma").value = r.merma_porcentaje || 0;
    $("recTiempo").value = r.tiempo_estimado_min || 0;
    $("recEstado").value = r.estado || "activo";
    $("recComponentes").value = recetaToComponentLines(r);
    $("recSubmit").textContent = "Actualizar receta";
    document.querySelector('[data-tab="recetas"]').click();
  }

  function renderRecetas() {
    var list = $("recetasList");
    var filtered = state.recetas.filter(function (r) {
      return !state.filters.recetas || r.estado === state.filters.recetas;
    });
    list.innerHTML = filtered.length ? filtered.map(function (r) {
      var comps = (r.componentes || []).map(function (c) { return esc(c.producto_nombre) + " x " + esc(qty(c.cantidad)) + " " + esc(c.unidad); }).join(", ");
      return '<article class="prod-item">' +
        '<div class="prod-item-top"><strong>' + esc(r.codigo) + " - " + esc(r.nombre) + '</strong>' + chip(r.estado) + '</div>' +
        '<div class="prod-meta-grid">' +
          '<div class="prod-meta"><span>Producto</span><strong>' + esc(r.producto_terminado_nombre) + '</strong></div>' +
          '<div class="prod-meta"><span>Base</span><strong>' + esc(qty(r.cantidad_base)) + " " + esc(r.unidad) + '</strong></div>' +
          '<div class="prod-meta"><span>Costo</span><strong>' + money(r.costo_estandar) + '</strong></div>' +
        '</div>' +
        '<div class="prod-muted">' + (comps || "Sin componentes registrados") + '</div>' +
        '<div class="prod-row-actions">' +
          '<button class="prod-btn small" type="button" data-receta="' + r.id + '">Usar en orden</button>' +
          '<button class="prod-btn small" type="button" data-edit-receta="' + r.id + '">Editar</button>' +
        '</div>' +
      '</article>';
    }).join("") : '<div class="prod-empty">No hay recetas para el filtro seleccionado.</div>';

    var opts = state.recetas.filter(function (r) { return r.estado !== "inactivo"; }).map(function (r) {
      return '<option value="' + r.id + '">' + esc(r.codigo) + " - " + esc(r.nombre) + '</option>';
    }).join("");
    $("ordReceta").innerHTML = opts || '<option value="">Sin recetas</option>';
  }

  function clearOrdenForm() {
    $("ordenForm").reset();
    $("ordId").value = "";
    $("ordCantidad").value = "1";
    $("ordPrioridad").value = "normal";
    $("ordEstado").value = "programada";
    $("ordenFormTitle").textContent = "Orden de produccion";
    $("ordSubmit").textContent = "Crear orden";
  }

  function editOrden(id) {
    var o = state.ordenes.find(function (x) { return String(x.id) === String(id); });
    if (!o) return;
    $("ordId").value = o.id || "";
    $("ordReceta").value = o.receta_id || "";
    $("ordCantidad").value = o.cantidad_planificada || 1;
    $("ordPrioridad").value = o.prioridad || "normal";
    $("ordFecha").value = (o.fecha_programada || "").slice(0, 10);
    $("ordResponsable").value = o.responsable || "";
    $("ordEstado").value = o.estado || "programada";
    $("ordObs").value = o.observaciones || "";
    $("ordenFormTitle").textContent = "Editar orden " + (o.codigo || "");
    $("ordSubmit").textContent = "Actualizar orden";
    document.querySelector('[data-tab="ordenes"]').click();
  }

  function orderActions(o) {
    var actions = ['<button class="prod-btn small" type="button" data-edit-orden="' + o.id + '">Editar</button>'];
    if (o.estado === "borrador" || o.estado === "programada") actions.push('<button class="prod-btn small" type="button" data-state="en_proceso" data-order="' + o.id + '">Iniciar</button>');
    if (o.estado === "en_proceso") actions.push('<button class="prod-btn small" type="button" data-state="calidad" data-order="' + o.id + '">Enviar a calidad</button>');
    if ((o.estado === "en_proceso" && !state.config.cerrar_con_calidad) || o.estado === "calidad") actions.push('<button class="prod-btn small" type="button" data-state="cerrada" data-order="' + o.id + '">Cerrar</button>');
    if (o.estado !== "cerrada" && o.estado !== "cancelada") actions.push('<button class="prod-btn small danger" type="button" data-state="cancelada" data-order="' + o.id + '">Cancelar</button>');
    return actions.join("");
  }

  function renderOrdenes() {
    var list = $("ordenesList");
    var q = String(state.filters.search || "").toLowerCase();
    var filtered = state.ordenes.filter(function (o) {
      if (state.filters.ordenes && o.estado !== state.filters.ordenes) return false;
      if (!q) return true;
      return [o.codigo, o.producto_terminado_nombre, o.responsable, o.prioridad].join(" ").toLowerCase().indexOf(q) >= 0;
    });
    list.innerHTML = filtered.length ? filtered.map(function (o) {
      var progress = pct(num(o.cantidad_producida), num(o.cantidad_planificada));
      var variance = num(o.costo_real) - num(o.costo_estimado);
      return '<article class="prod-item">' +
        '<div class="prod-item-top"><strong>' + esc(o.codigo) + " - " + esc(o.producto_terminado_nombre) + '</strong>' + chip(o.estado) + '</div>' +
        '<div class="prod-progress" title="Avance de produccion"><span style="width:' + progress + '%"></span></div>' +
        '<div class="prod-meta-grid">' +
          '<div class="prod-meta"><span>Cantidad</span><strong>' + esc(qty(o.cantidad_producida)) + " / " + esc(qty(o.cantidad_planificada)) + '</strong></div>' +
          '<div class="prod-meta"><span>Costo</span><strong>' + money(o.costo_real || o.costo_estimado) + '</strong></div>' +
          '<div class="prod-meta"><span>Variacion</span><strong>' + money(variance) + '</strong></div>' +
        '</div>' +
        '<div class="prod-muted">' + esc(o.responsable || "Sin responsable") + " - " + esc(label(o.prioridad || "normal")) + (o.fecha_programada ? " - " + esc(o.fecha_programada.slice(0, 10)) : "") + '</div>' +
        '<div class="prod-row-actions">' + orderActions(o) + '</div>' +
      '</article>';
    }).join("") : '<div class="prod-empty">No hay ordenes para el filtro seleccionado.</div>';

    var opts = state.ordenes.filter(function (o) { return o.estado !== "cerrada" && o.estado !== "cancelada"; }).map(function (o) {
      return '<option value="' + o.id + '">' + esc(o.codigo) + " - " + esc(o.producto_terminado_nombre) + '</option>';
    }).join("");
    $("consOrden").innerHTML = opts || '<option value="">Sin ordenes abiertas</option>';
    $("calOrden").innerHTML = opts || '<option value="">Sin ordenes abiertas</option>';
  }

  function renderMovimientos() {
    var consumoRows = state.consumos.map(function (c) {
      return '<tr><td>Consumo</td><td>' + esc(c.producto_nombre) + '</td><td class="num">' + esc(qty(c.cantidad_consumida)) + '</td><td class="num">' + money(c.costo_total) + '</td><td>' + esc(c.fecha_consumo || "") + '</td></tr>';
    });
    var calidadRows = state.calidad.map(function (c) {
      return '<tr><td>' + chip(c.resultado) + '</td><td>Orden #' + esc(c.orden_id) + '</td><td class="num">' + esc(qty(c.cantidad_aprobada)) + '</td><td class="num">' + esc(qty(c.cantidad_rechazada)) + '</td><td>' + esc(c.responsable || "") + '</td><td>' + esc(c.fecha_revision || "") + '</td></tr>';
    });
    $("consumosTable").innerHTML = '<table class="prod-table"><thead><tr><th>Tipo</th><th>Producto</th><th>Cantidad</th><th>Costo</th><th>Fecha</th></tr></thead><tbody>' + (consumoRows.join("") || '<tr><td colspan="5">Sin consumos registrados.</td></tr>') + '</tbody></table>';
    $("calidadTable").innerHTML = '<table class="prod-table"><thead><tr><th>Resultado</th><th>Orden</th><th>Aprobada</th><th>Rechazada</th><th>Responsable</th><th>Fecha</th></tr></thead><tbody>' + (calidadRows.join("") || '<tr><td colspan="6">Sin revisiones de calidad.</td></tr>') + '</tbody></table>';
  }

  function renderMRP() {
    var rows = state.plan.map(function (p) {
      var risk = num(p.disponible_proyectado) < 0 || num(p.cantidad_sugerida_compra) > 0 || num(p.cantidad_sugerida_producir) > 0;
      return '<tr class="' + (risk ? "is-risk" : "") + '">' +
        '<td>' + esc(p.periodo) + '</td><td>' + esc(p.producto_nombre) + '</td><td>' + esc(p.origen || "") + '</td>' +
        '<td class="num">' + esc(qty(p.demanda_estimada)) + '</td><td class="num">' + esc(qty(p.requerido_bruto)) + '</td>' +
        '<td class="num">' + esc(qty(p.stock_seguridad)) + '</td><td class="num">' + esc(qty(p.disponible_proyectado)) + '</td>' +
        '<td class="num">' + esc(qty(p.cantidad_sugerida_compra)) + '</td><td class="num">' + esc(qty(p.cantidad_sugerida_producir)) + '</td><td>' + chip(p.estado) + '</td></tr>';
    });
    $("mrpTable").innerHTML = '<table class="prod-table"><thead><tr><th>Periodo</th><th>Producto/Material</th><th>Origen</th><th>Demanda</th><th>Requerido</th><th>Stock seguridad</th><th>Disponible</th><th>Comprar</th><th>Producir</th><th>Estado</th></tr></thead><tbody>' + (rows.join("") || '<tr><td colspan="10">Sin plan MRP generado para ordenes abiertas.</td></tr>') + '</tbody></table>';
  }

  function parseComponentes(raw) {
    return String(raw || "").split(/\n+/).map(function (line, idx) {
      var parts = line.split("|").map(function (p) { return p.trim(); });
      return {
        producto_nombre: parts[0] || "",
        cantidad: num(parts[1] || 0),
        unidad: parts[2] || "und",
        costo_unitario: num(parts[3] || 0),
        merma_porcentaje: num(parts[4] || 0),
        obligatoria: true,
        etapa: parts[5] || "produccion",
        orden: idx + 1
      };
    }).filter(function (x) { return x.producto_nombre && x.cantidad > 0; });
  }

  function renderAll() {
    renderDashboard();
    renderRecetas();
    renderOrdenes();
    renderMovimientos();
    renderMRP();
  }

  async function load() {
    try {
      setBusy(true, "Actualizando Produccion/MRP...");
      var d = await api("dashboard");
      state.dashboard = d || {};
      fillKpis(d);
      state.recetas = d.recetas || [];
      state.ordenes = d.ordenes || [];
      state.plan = d.plan || [];
      state.consumos = d.consumos_recientes || [];
      state.calidad = d.revisiones_calidad || [];
      fillConfig(d.config || {});
      renderAll();
      msg("Produccion/MRP actualizado.");
    } catch (e) {
      msg(e.message, true);
    } finally {
      setBusy(false);
    }
  }

  async function generateMRP() {
    try {
      setBusy(true, "Generando MRP...");
      var period = ($("mrpPeriodo") && $("mrpPeriodo").value) || todayMonth();
      state.plan = await api("generar_mrp", body({ periodo: period }));
      renderMRP();
      renderDashboard();
      msg("Plan MRP generado para " + period + ".");
      document.querySelector('[data-tab="mrp"]').click();
    } catch (e) {
      msg(e.message, true);
    } finally {
      setBusy(false);
    }
  }

  function downloadCSV(filename, rows) {
    var csv = rows.map(function (row) {
      return row.map(function (cell) {
        var text = String(cell == null ? "" : cell);
        return /[",\n;]/.test(text) ? '"' + text.replace(/"/g, '""') + '"' : text;
      }).join(";");
    }).join("\n");
    var blob = new Blob([csv], { type: "text/csv;charset=utf-8" });
    var url = URL.createObjectURL(blob);
    var a = document.createElement("a");
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
  }

  function exportCSV() {
    var rows = [["Tipo", "Codigo/Periodo", "Producto", "Estado", "Cantidad/Demanda", "Comprar", "Producir", "Costo"]];
    state.ordenes.forEach(function (o) {
      rows.push(["Orden", o.codigo, o.producto_terminado_nombre, label(o.estado), qty(o.cantidad_planificada), "", "", money(o.costo_estimado)]);
    });
    state.plan.forEach(function (p) {
      rows.push(["MRP", p.periodo, p.producto_nombre, label(p.estado), qty(p.demanda_estimada), qty(p.cantidad_sugerida_compra), qty(p.cantidad_sugerida_producir), ""]);
    });
    downloadCSV("produccion_mrp_" + todayMonth() + ".csv", rows);
  }

  document.addEventListener("click", async function (ev) {
    var tab = ev.target.closest("[data-tab]");
    if (tab) {
      setTab(tab.dataset.tab);
      return;
    }
    var recetaBtn = ev.target.closest("[data-receta]");
    if (recetaBtn) {
      $("ordReceta").value = recetaBtn.dataset.receta;
      document.querySelector('[data-tab="ordenes"]').click();
      return;
    }
    var editRecetaBtn = ev.target.closest("[data-edit-receta]");
    if (editRecetaBtn) {
      editReceta(editRecetaBtn.dataset.editReceta);
      return;
    }
    var editOrdenBtn = ev.target.closest("[data-edit-orden]");
    if (editOrdenBtn) {
      editOrden(editOrdenBtn.dataset.editOrden);
      return;
    }
    var stateBtn = ev.target.closest("[data-state][data-order]");
    if (stateBtn) {
      try {
        setBusy(true, "Actualizando estado...");
        await api("orden_estado", body({ orden_id: Number(stateBtn.dataset.order), estado: stateBtn.dataset.state }));
        await load();
      } catch (e) {
        msg(e.message, true);
      } finally {
        setBusy(false);
      }
    }
  });

  $("prodRefresh").onclick = load;
  $("prodSeed").onclick = async function () { try { setBusy(true, "Cargando demo..."); await api("seed_demo", body({})); await load(); } catch (e) { msg(e.message, true); } finally { setBusy(false); } };
  $("prodGenerateMRP").onclick = generateMRP;
  $("mrpGenerate").onclick = generateMRP;
  $("prodExport").onclick = exportCSV;
  $("recCancel").onclick = clearRecetaForm;
  $("ordCancel").onclick = clearOrdenForm;
  $("recFilter").onchange = function () { state.filters.recetas = this.value; renderRecetas(); };
  $("ordenFilter").onchange = function () { state.filters.ordenes = this.value; renderOrdenes(); };
  $("ordenSearch").oninput = function () { state.filters.search = this.value; renderOrdenes(); };
  $("mrpPeriodo").value = todayMonth();
  setTab(requestedTab(), true);

  $("recetaForm").onsubmit = async function (ev) {
    ev.preventDefault();
    try {
      var payload = {
        id: Number($("recId").value || 0),
        codigo: $("recCodigo").value,
        nombre: $("recNombre").value,
        producto_terminado_nombre: $("recProducto").value,
        version: $("recVersion").value,
        unidad: $("recUnidad").value,
        cantidad_base: num($("recCantidad").value),
        costo_estandar: num($("recCosto").value),
        merma_porcentaje: num($("recMerma").value),
        tiempo_estimado_min: Number($("recTiempo").value || 0),
        estado: $("recEstado").value,
        componentes: parseComponentes($("recComponentes").value)
      };
      setBusy(true, payload.id ? "Actualizando receta..." : "Guardando receta...");
      await api("recetas", body(payload, payload.id ? "PUT" : "POST"));
      clearRecetaForm();
      await load();
    } catch (e) {
      msg(e.message, true);
    } finally {
      setBusy(false);
    }
  };

  $("ordenForm").onsubmit = async function (ev) {
    ev.preventDefault();
    try {
      var rec = state.recetas.find(function (r) { return String(r.id) === String($("ordReceta").value); }) || {};
      var id = Number($("ordId").value || 0);
      var payload = {
        id: id,
        receta_id: Number($("ordReceta").value),
        producto_terminado_nombre: rec.producto_terminado_nombre || rec.nombre || "",
        cantidad_planificada: num($("ordCantidad").value),
        prioridad: $("ordPrioridad").value,
        fecha_programada: $("ordFecha").value,
        responsable: $("ordResponsable").value,
        estado: $("ordEstado").value,
        observaciones: $("ordObs").value
      };
      setBusy(true, id ? "Actualizando orden..." : "Creando orden...");
      await api("ordenes", body(payload, id ? "PUT" : "POST"));
      clearOrdenForm();
      await load();
    } catch (e) {
      msg(e.message, true);
    } finally {
      setBusy(false);
    }
  };

  $("consumoForm").onsubmit = async function (ev) {
    ev.preventDefault();
    try {
      setBusy(true, "Registrando consumo...");
      await api("consumos", body({ orden_id: Number($("consOrden").value), producto_nombre: $("consProducto").value, cantidad_consumida: num($("consCantidad").value), costo_unitario: num($("consCosto").value), lote_codigo: $("consLote").value }));
      ev.target.reset();
      $("consCantidad").value = "1";
      await load();
    } catch (e) {
      msg(e.message, true);
    } finally {
      setBusy(false);
    }
  };

  $("calidadForm").onsubmit = async function (ev) {
    ev.preventDefault();
    try {
      setBusy(true, "Registrando calidad...");
      await api("calidad", body({ orden_id: Number($("calOrden").value), resultado: $("calResultado").value, cantidad_aprobada: num($("calAprobada").value), cantidad_rechazada: num($("calRechazada").value), responsable: $("calResponsable").value, observaciones: $("calObs").value }));
      ev.target.reset();
      $("calResultado").value = "pendiente";
      $("calAprobada").value = "0";
      $("calRechazada").value = "0";
      await load();
    } catch (e) {
      msg(e.message, true);
    } finally {
      setBusy(false);
    }
  };

  $("configForm").onsubmit = async function (ev) {
    ev.preventDefault();
    try {
      setBusy(true, "Guardando configuracion...");
      await api("config", body({ nombre_sistema: $("cfgNombre").value, moneda: $("cfgMoneda").value, costeo_modo: $("cfgCosteo").value, aprobar_ordenes: $("cfgAprobar").checked, consumir_inventario_al_iniciar: $("cfgConsumir").checked, cerrar_con_calidad: $("cfgCalidad").checked }));
      await load();
    } catch (e) {
      msg(e.message, true);
    } finally {
      setBusy(false);
    }
  };

  clearRecetaForm();
  clearOrdenForm();
  load();
})();
