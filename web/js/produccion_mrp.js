(function () {
  "use strict";

  var empresaId = "";
  try {
    empresaId = (window.__resolveEmpresaIdContext && window.__resolveEmpresaIdContext()) || "";
  } catch (e) {}
  if (!empresaId) {
    empresaId = new URLSearchParams(location.search).get("empresa_id") || new URLSearchParams(location.search).get("id") || "";
  }

  var state = { recetas: [], ordenes: [], consumos: [], calidad: [], plan: [], config: {} };
  var fmt = new Intl.NumberFormat("es-CO", { style: "currency", currency: "COP", maximumFractionDigits: 0 });

  function $(id) { return document.getElementById(id); }
  function esc(v) { return String(v == null ? "" : v).replace(/[&<>"]/g, function (c) { return ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" })[c]; }); }
  function num(v) { var n = Number(v); return Number.isFinite(n) ? n : 0; }
  function money(v) { try { return fmt.format(num(v)); } catch (e) { return "$" + num(v).toLocaleString("es-CO"); } }
  function msg(text, isError) { var el = $("prodMsg"); if (!el) return; el.textContent = text || ""; el.classList.toggle("is-error", !!isError); }

  async function api(action, opts) {
    if (!empresaId) throw new Error("No se encontró empresa_id para producción/MRP.");
    var url = "/api/empresa/produccion_mrp?empresa_id=" + encodeURIComponent(empresaId) + (action ? "&action=" + encodeURIComponent(action) : "");
    var res = await fetch(url, Object.assign({ credentials: "same-origin" }, opts || {}));
    var txt = await res.text();
    var data = {};
    try { data = txt ? JSON.parse(txt) : {}; } catch (e) { data = { raw: txt }; }
    if (!res.ok) throw new Error(data.error || data.raw || txt || ("HTTP " + res.status));
    return data;
  }

  function body(payload) {
    return { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(payload || {}) };
  }

  function fillKpis(d) {
    $("kpiRecetas").textContent = d.recetas_activas || 0;
    $("kpiAbiertas").textContent = d.ordenes_abiertas || 0;
    $("kpiCalidad").textContent = d.ordenes_calidad || 0;
    $("kpiCerradas").textContent = d.ordenes_cerradas || 0;
    $("kpiCostoAbierto").textContent = money(d.costo_estimado_abierto || 0);
    $("kpiCostoMes").textContent = money(d.costo_real_mes || 0);
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

  function renderRecetas() {
    var list = $("recetasList");
    list.innerHTML = state.recetas.length ? state.recetas.map(function (r) {
      var comps = (r.componentes || []).map(function (c) { return esc(c.producto_nombre) + " x " + esc(c.cantidad) + " " + esc(c.unidad); }).join(", ");
      return '<article class="prod-item">' +
        '<div class="prod-item-top"><strong>' + esc(r.codigo) + " · " + esc(r.nombre) + '</strong><span class="prod-chip">' + esc(r.estado) + '</span></div>' +
        '<div class="prod-muted">' + esc(r.producto_terminado_nombre) + " · base " + esc(r.cantidad_base) + " " + esc(r.unidad) + " · " + money(r.costo_estandar) + '</div>' +
        '<div class="prod-muted">' + (comps || "Sin componentes registrados") + '</div>' +
        '<button class="prod-btn" type="button" data-receta="' + r.id + '">Usar en orden</button>' +
      '</article>';
    }).join("") : '<div class="prod-item">Sin recetas. Carga demo o registra una BOM.</div>';

    var opts = state.recetas.map(function (r) { return '<option value="' + r.id + '">' + esc(r.codigo) + " · " + esc(r.nombre) + '</option>'; }).join("");
    $("ordReceta").innerHTML = opts || '<option value="">Sin recetas</option>';
  }

  function renderOrdenes() {
    var list = $("ordenesList");
    list.innerHTML = state.ordenes.length ? state.ordenes.map(function (o) {
      return '<article class="prod-item">' +
        '<div class="prod-item-top"><strong>' + esc(o.codigo) + " · " + esc(o.producto_terminado_nombre) + '</strong><span class="prod-chip">' + esc(o.estado) + '</span></div>' +
        '<div class="prod-muted">Cantidad ' + esc(o.cantidad_planificada) + " · producido " + esc(o.cantidad_producida) + " · " + money(o.costo_estimado) + '</div>' +
        '<div class="prod-muted">' + esc(o.responsable || "Sin responsable") + " · " + esc(o.prioridad || "normal") + '</div>' +
        '<div class="prod-actions">' +
          '<button class="prod-btn" type="button" data-state="en_proceso" data-order="' + o.id + '">Iniciar</button>' +
          '<button class="prod-btn" type="button" data-state="calidad" data-order="' + o.id + '">Calidad</button>' +
          '<button class="prod-btn" type="button" data-state="cerrada" data-order="' + o.id + '">Cerrar</button>' +
        '</div>' +
      '</article>';
    }).join("") : '<div class="prod-item">Sin órdenes registradas.</div>';

    var opts = state.ordenes.map(function (o) { return '<option value="' + o.id + '">' + esc(o.codigo) + " · " + esc(o.producto_terminado_nombre) + '</option>'; }).join("");
    $("consOrden").innerHTML = opts || '<option value="">Sin órdenes</option>';
    $("calOrden").innerHTML = opts || '<option value="">Sin órdenes</option>';
  }

  function renderMovimientos() {
    var rows = state.consumos.map(function (c) {
      return '<tr><td>Consumo</td><td>' + esc(c.producto_nombre) + '</td><td>' + esc(c.cantidad_consumida) + '</td><td>' + money(c.costo_total) + '</td><td>' + esc(c.fecha_consumo || "") + '</td></tr>';
    }).concat(state.calidad.map(function (c) {
      return '<tr><td>Calidad</td><td>Orden #' + esc(c.orden_id) + " · " + esc(c.resultado) + '</td><td>' + esc(c.cantidad_aprobada) + '</td><td>' + esc(c.cantidad_rechazada) + '</td><td>' + esc(c.fecha_revision || "") + '</td></tr>';
    }));
    $("movimientosTable").innerHTML = '<table class="prod-table"><thead><tr><th>Tipo</th><th>Detalle</th><th>Cantidad</th><th>Valor</th><th>Fecha</th></tr></thead><tbody>' + (rows.join("") || '<tr><td colspan="5">Sin movimientos.</td></tr>') + '</tbody></table>';
  }

  function renderMRP() {
    var rows = state.plan.map(function (p) {
      return '<tr><td>' + esc(p.periodo) + '</td><td>' + esc(p.producto_nombre) + '</td><td>' + esc(p.demanda_estimada) + '</td><td>' + esc(p.stock_seguridad) + '</td><td>' + esc(p.cantidad_sugerida_producir) + '</td><td>' + esc(p.estado) + '</td></tr>';
    });
    $("mrpTable").innerHTML = '<table class="prod-table"><thead><tr><th>Periodo</th><th>Producto</th><th>Demanda</th><th>Stock seguridad</th><th>Sugerido producir</th><th>Estado</th></tr></thead><tbody>' + (rows.join("") || '<tr><td colspan="6">Sin plan MRP generado.</td></tr>') + '</tbody></table>';
  }

  function parseComponentes(raw) {
    return String(raw || "").split(/\n+/).map(function (line, idx) {
      var parts = line.split("|").map(function (p) { return p.trim(); });
      return { producto_nombre: parts[0] || "", cantidad: num(parts[1] || 0), unidad: parts[2] || "und", costo_unitario: num(parts[3] || 0), obligatoria: true, etapa: "produccion", orden: idx + 1 };
    }).filter(function (x) { return x.producto_nombre && x.cantidad > 0; });
  }

  async function load() {
    try {
      var d = await api("dashboard");
      fillKpis(d);
      state.recetas = d.recetas || [];
      state.ordenes = d.ordenes || [];
      state.plan = d.plan || [];
      state.consumos = d.consumos_recientes || [];
      state.calidad = d.revisiones_calidad || [];
      fillConfig(d.config || {});
      renderRecetas();
      renderOrdenes();
      renderMovimientos();
      renderMRP();
      msg("Producción/MRP actualizado.");
    } catch (e) {
      msg(e.message, true);
    }
  }

  document.addEventListener("click", async function (ev) {
    var tab = ev.target.closest(".prod-tab");
    if (tab) {
      document.querySelectorAll(".prod-tab").forEach(function (b) { b.classList.toggle("is-active", b === tab); });
      document.querySelectorAll(".prod-panel").forEach(function (p) { p.classList.toggle("is-active", p.id === "tab-" + tab.dataset.tab); });
      return;
    }
    var recetaBtn = ev.target.closest("[data-receta]");
    if (recetaBtn) {
      $("ordReceta").value = recetaBtn.dataset.receta;
      document.querySelector('[data-tab="ordenes"]').click();
      return;
    }
    var stateBtn = ev.target.closest("[data-state][data-order]");
    if (stateBtn) {
      try {
        await api("orden_estado", body({ orden_id: Number(stateBtn.dataset.order), estado: stateBtn.dataset.state }));
        await load();
      } catch (e) { msg(e.message, true); }
    }
  });

  $("prodRefresh").onclick = load;
  $("prodSeed").onclick = async function () { try { await api("seed_demo", body({})); await load(); } catch (e) { msg(e.message, true); } };
  $("prodGenerateMRP").onclick = async function () { try { state.plan = await api("generar_mrp", body({ periodo: new Date().toISOString().slice(0, 7) })); renderMRP(); msg("Plan MRP generado."); } catch (e) { msg(e.message, true); } };

  $("recetaForm").onsubmit = async function (ev) {
    ev.preventDefault();
    try {
      await api("recetas", body({
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
      }));
      ev.target.reset();
      $("recVersion").value = "1.0"; $("recUnidad").value = "und"; $("recCantidad").value = "1";
      await load();
    } catch (e) { msg(e.message, true); }
  };

  $("ordenForm").onsubmit = async function (ev) {
    ev.preventDefault();
    try {
      var rec = state.recetas.find(function (r) { return String(r.id) === String($("ordReceta").value); }) || {};
      await api("ordenes", body({ receta_id: Number($("ordReceta").value), producto_terminado_nombre: rec.producto_terminado_nombre || rec.nombre || "", cantidad_planificada: num($("ordCantidad").value), prioridad: $("ordPrioridad").value, fecha_programada: $("ordFecha").value, responsable: $("ordResponsable").value, estado: $("ordEstado").value, observaciones: $("ordObs").value }));
      await load();
    } catch (e) { msg(e.message, true); }
  };

  $("consumoForm").onsubmit = async function (ev) {
    ev.preventDefault();
    try {
      await api("consumos", body({ orden_id: Number($("consOrden").value), producto_nombre: $("consProducto").value, cantidad_consumida: num($("consCantidad").value), costo_unitario: num($("consCosto").value), lote_codigo: $("consLote").value }));
      await load();
    } catch (e) { msg(e.message, true); }
  };

  $("calidadForm").onsubmit = async function (ev) {
    ev.preventDefault();
    try {
      await api("calidad", body({ orden_id: Number($("calOrden").value), resultado: $("calResultado").value, cantidad_aprobada: num($("calAprobada").value), cantidad_rechazada: num($("calRechazada").value), responsable: $("calResponsable").value, observaciones: $("calObs").value }));
      await load();
    } catch (e) { msg(e.message, true); }
  };

  $("configForm").onsubmit = async function (ev) {
    ev.preventDefault();
    try {
      await api("config", body({ nombre_sistema: $("cfgNombre").value, moneda: $("cfgMoneda").value, costeo_modo: $("cfgCosteo").value, aprobar_ordenes: $("cfgAprobar").checked, consumir_inventario_al_iniciar: $("cfgConsumir").checked, cerrar_con_calidad: $("cfgCalidad").checked }));
      await load();
    } catch (e) { msg(e.message, true); }
  };

  load();
})();
