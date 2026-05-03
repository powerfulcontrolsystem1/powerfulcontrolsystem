(function () {
  "use strict";

  var DAYS = [
    { id: 1, name: "Lunes" },
    { id: 2, name: "Martes" },
    { id: 3, name: "Miercoles" },
    { id: 4, name: "Jueves" },
    { id: 5, name: "Viernes" },
    { id: 6, name: "Sabado" },
    { id: 7, name: "Domingo" }
  ];

  var state = {
    empresaID: 0,
    stations: [],
    rates: []
  };

  function $(id) { return document.getElementById(id); }

  function text(value) {
    return String(value == null ? "" : value).trim();
  }

  function number(value, fallback) {
    var n = Number(value);
    return Number.isFinite(n) ? n : (fallback || 0);
  }

  function esc(value) {
    return String(value == null ? "" : value)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#39;");
  }

  function money(value, currency) {
    var curr = text(currency || "COP") || "COP";
    try {
      return new Intl.NumberFormat("es-CO", { style: "currency", currency: curr, maximumFractionDigits: 0 }).format(number(value));
    } catch (e) {
      return curr + " " + number(value).toLocaleString("es-CO");
    }
  }

  function getQueryParam(name) {
    try { return new URLSearchParams(window.location.search || "").get(name); } catch (e) { return ""; }
  }

  function readStoredEmpresaID() {
    var keys = ["active_empresa_id", "empresa_id", "admin_empresa_id"];
    var stores = [];
    try { stores.push(window.sessionStorage); } catch (e) {}
    try { stores.push(window.localStorage); } catch (e) {}
    for (var s = 0; s < stores.length; s += 1) {
      for (var k = 0; k < keys.length; k += 1) {
        var id = number(stores[s].getItem(keys[k]));
        if (id > 0) return id;
      }
    }
    return 0;
  }

  function resolveEmpresaID() {
    var id = number(getQueryParam("empresa_id") || getQueryParam("id"));
    if (id <= 0 && window.parent && window.parent !== window) {
      try {
        if (typeof window.parent.__resolveEmpresaIdContext === "function") {
          id = number(window.parent.__resolveEmpresaIdContext());
        }
      } catch (e) {}
    }
    if (id <= 0) id = readStoredEmpresaID();
    if (id > 0) {
      try {
        sessionStorage.setItem("active_empresa_id", String(id));
        sessionStorage.setItem("empresa_id", String(id));
      } catch (e) {}
    }
    return id;
  }

  function setMessage(id, message, isError) {
    var el = $(id);
    if (!el) return;
    el.textContent = message || "";
    el.classList.toggle("is-error", !!isError);
  }

  async function ensureOk(resp) {
    if (resp.ok) return;
    var raw = await resp.text().catch(function () { return ""; });
    throw new Error(raw || ("HTTP " + resp.status));
  }

  async function fetchJSON(url, options) {
    var resp = await fetch(url, Object.assign({ credentials: "same-origin" }, options || {}));
    await ensureOk(resp);
    return resp.status === 204 ? null : resp.json();
  }

  function endpoint(extra) {
    var params = new URLSearchParams(extra || {});
    params.set("empresa_id", String(state.empresaID));
    return "/api/empresa/tarifas_motel?" + params.toString();
  }

  function defaultStationCode(empresaID, stationID) {
    if (!empresaID || !stationID) return "";
    return "EST-" + empresaID + "-" + stationID;
  }

  function stationStorageKey(empresaID) {
    return "estaciones_config_empresa_" + empresaID;
  }

  function parseStationIDFromCode(code, empresaID) {
    var match = /^EST-(\d+)-(\d+)$/i.exec(text(code));
    if (!match) return 0;
    var emp = number(match[1]);
    var station = number(match[2]);
    if (empresaID > 0 && emp !== empresaID) return 0;
    return station;
  }

  function findStation(id) {
    var sid = number(id);
    return state.stations.find(function (item) { return item.id === sid; }) || null;
  }

  function dayName(id) {
    var found = DAYS.find(function (day) { return day.id === number(id); });
    return found ? found.name : "-";
  }

  function dayRange(from, to) {
    var a = dayName(from);
    var b = dayName(to);
    return a === b ? a : a + " - " + b;
  }

  function typeLabel(value) {
    var labels = {
      express: "Express",
      day_use: "Day-use",
      nocturno: "Nocturno",
      amanecida: "Amanecida",
      suite: "Suite",
      vip: "VIP",
      promocion: "Promocion"
    };
    return labels[text(value || "express").toLowerCase()] || "Express";
  }

  function fillDaySelect(id) {
    var select = $(id);
    if (!select) return;
    select.innerHTML = DAYS.map(function (day) {
      return '<option value="' + day.id + '">' + esc(day.name) + '</option>';
    }).join("");
  }

  function fillStationSelect(id, includeAll) {
    var select = $(id);
    if (!select) return;
    var html = includeAll ? '<option value="">Todas</option>' : '<option value="">Selecciona una habitacion</option>';
    state.stations.forEach(function (station) {
      html += '<option value="' + station.id + '">' + esc(station.name + " (#" + station.id + ")") + '</option>';
    });
    if (!state.stations.length) {
      html += '<option value="">Sin estaciones detectadas, usa ID manual</option>';
    }
    select.innerHTML = html;
  }

  function syncStation(id) {
    var stationID = number(id);
    if (stationID <= 0) return;
    var station = findStation(stationID);
    $("stationId").value = String(stationID);
    $("stationCode").value = station && station.code ? station.code : defaultStationCode(state.empresaID, stationID);
    $("stationName").value = station && station.name ? station.name : "Habitacion " + stationID;
  }

  async function loadStations() {
    var byID = new Map();
    try {
      var raw = localStorage.getItem(stationStorageKey(state.empresaID));
      if (raw) {
        var parsed = JSON.parse(raw);
        var rows = parsed && Array.isArray(parsed.estaciones) ? parsed.estaciones : [];
        rows.forEach(function (item, index) {
          var id = number(item && item.id ? item.id : index + 1);
          if (id <= 0) return;
          byID.set(id, { id: id, code: defaultStationCode(state.empresaID, id), name: text(item && item.nombre) || "Habitacion " + id });
        });
      }
    } catch (e) {
      console.warn("No se pudo leer estaciones locales", e);
    }

    try {
      var params = new URLSearchParams();
      params.set("empresa_id", String(state.empresaID));
      params.set("include_inactive", "1");
      var resp = await fetch("/api/empresa/carritos_compra?" + params.toString(), { credentials: "same-origin" });
      if (resp.ok) {
        var carts = await resp.json();
        (Array.isArray(carts) ? carts : []).forEach(function (item) {
          var code = text(item && item.codigo).toUpperCase();
          var id = parseStationIDFromCode(code, state.empresaID);
          if (id <= 0) return;
          byID.set(id, { id: id, code: code || defaultStationCode(state.empresaID, id), name: text(item && item.nombre) || "Habitacion " + id });
        });
      }
    } catch (e) {
      console.warn("No se pudieron cargar estaciones desde carritos", e);
    }

    state.stations = Array.from(byID.values()).sort(function (a, b) { return a.id - b.id; });
    fillStationSelect("stationSelect", false);
    fillStationSelect("filterStation", true);
  }

  function buildPayload(stationOverride) {
    var stationID = stationOverride ? number(stationOverride.id) : number($("stationId").value);
    return {
      id: stationOverride ? 0 : number($("rateId").value),
      empresa_id: state.empresaID,
      estacion_id: stationID,
      estacion_codigo: stationOverride ? (stationOverride.code || defaultStationCode(state.empresaID, stationID)) : (text($("stationCode").value) || defaultStationCode(state.empresaID, stationID)),
      estacion_nombre: stationOverride ? (stationOverride.name || ("Habitacion " + stationID)) : (text($("stationName").value) || "Habitacion " + stationID),
      nombre_plan: text($("planName").value) || "Express 3 horas",
      tipo_plan: text($("planType").value) || "express",
      categoria_habitacion: text($("roomCategory").value),
      dia_semana_desde: number($("dayFrom").value, 1),
      dia_semana_hasta: number($("dayTo").value, 7),
      hora_inicio: text($("startTime").value) || "00:00",
      hora_fin: text($("endTime").value) || "23:59",
      minutos_incluidos: number($("includedMinutes").value, 180),
      valor_base: number($("baseValue").value),
      minutos_extra: number($("extraMinutes").value, 60),
      valor_extra: number($("extraValue").value),
      cobrar_por_fraccion: !!$("chargeFraction").checked,
      tolerancia_minutos: number($("toleranceMinutes").value),
      moneda: text($("currency").value) || "COP",
      prioridad: number($("priority").value, 1),
      aplicar_automaticamente: !!$("autoApply").checked,
      estado: text($("status").value) || "activo",
      observaciones: text($("notes").value)
    };
  }

  function validatePayload(payload) {
    if (state.empresaID <= 0) return "No se encontro empresa_id para guardar.";
    if (payload.estacion_id <= 0) return "Selecciona una habitacion o escribe el ID de estacion.";
    if (payload.valor_base < 0 || payload.valor_extra < 0) return "Los valores no pueden ser negativos.";
    if (payload.minutos_incluidos <= 0 || payload.minutos_extra <= 0) return "Los minutos incluidos y extra deben ser mayores a cero.";
    return "";
  }

  function findExistingFor(payload) {
    return state.rates.find(function (item) {
      return number(item.estacion_id) === number(payload.estacion_id) &&
        text(item.nombre_plan).toLowerCase() === text(payload.nombre_plan).toLowerCase() &&
        text(item.tipo_plan).toLowerCase() === text(payload.tipo_plan).toLowerCase();
    }) || null;
  }

  async function savePayload(payload) {
    var existing = payload.id > 0 ? null : findExistingFor(payload);
    if (existing) payload.id = number(existing.id);
    return fetchJSON("/api/empresa/tarifas_motel", {
      method: payload.id > 0 ? "PUT" : "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload)
    });
  }

  async function saveRate(event) {
    event.preventDefault();
    var payload = buildPayload();
    var error = validatePayload(payload);
    if (error) {
      setMessage("formMsg", error, true);
      return;
    }
    setMessage("formMsg", "Guardando tarifa de motel...", false);
    try {
      await savePayload(payload);
      setMessage("formMsg", "Plan guardado correctamente.", false);
      resetForm();
      await loadRates();
    } catch (e) {
      setMessage("formMsg", e.message || "No se pudo guardar el plan.", true);
    }
  }

  async function applyAllRooms() {
    if (!state.stations.length) {
      setMessage("formMsg", "No hay estaciones detectadas para aplicar la tarifa.", true);
      return;
    }
    var template = buildPayload();
    var error = validatePayload(Object.assign({}, template, { estacion_id: state.stations[0].id }));
    if (error && error.indexOf("habitacion") < 0) {
      setMessage("formMsg", error, true);
      return;
    }
    setMessage("formMsg", "Aplicando plan a " + state.stations.length + " habitaciones...", false);
    var saved = 0;
    try {
      for (var i = 0; i < state.stations.length; i += 1) {
        var payload = buildPayload(state.stations[i]);
        var itemError = validatePayload(payload);
        if (itemError) throw new Error(itemError);
        await savePayload(payload);
        saved += 1;
      }
      setMessage("formMsg", "Plan aplicado a " + saved + " habitaciones.", false);
      await loadRates();
    } catch (e) {
      setMessage("formMsg", e.message || "No se pudo aplicar a todas las habitaciones.", true);
    }
  }

  function resetForm() {
    $("rateId").value = "";
    $("planName").value = "Express 3 horas";
    $("planType").value = "express";
    $("stationSelect").value = "";
    $("stationId").value = "";
    $("stationCode").value = "";
    $("stationName").value = "";
    $("roomCategory").value = "";
    $("dayFrom").value = "1";
    $("dayTo").value = "7";
    $("startTime").value = "00:00";
    $("endTime").value = "23:59";
    $("includedMinutes").value = "180";
    $("baseValue").value = "50000";
    $("extraMinutes").value = "60";
    $("extraValue").value = "15000";
    $("toleranceMinutes").value = "10";
    $("currency").value = "COP";
    $("priority").value = "1";
    $("status").value = "activo";
    $("chargeFraction").checked = true;
    $("autoApply").checked = true;
    $("notes").value = "";
    setMessage("formMsg", "", false);
  }

  function applyPreset(type) {
    var presets = {
      express: {
        name: "Express 3 horas", type: "express", category: "Estandar", from: "1", to: "7",
        start: "00:00", end: "23:59", included: "180", base: "50000", extraMinutes: "60",
        extraValue: "15000", tolerance: "10", notes: "Plan express para permanencias cortas."
      },
      day_use: {
        name: "Day-use 6 horas", type: "day_use", category: "Estandar", from: "1", to: "7",
        start: "08:00", end: "20:00", included: "360", base: "80000", extraMinutes: "60",
        extraValue: "18000", tolerance: "15", notes: "Plan diurno con hora adicional."
      },
      amanecida: {
        name: "Amanecida", type: "amanecida", category: "Estandar", from: "1", to: "7",
        start: "21:00", end: "08:00", included: "660", base: "120000", extraMinutes: "60",
        extraValue: "20000", tolerance: "15", notes: "Plan nocturno con cobro por hora adicional."
      },
      vip: {
        name: "VIP 4 horas", type: "vip", category: "VIP", from: "1", to: "7",
        start: "00:00", end: "23:59", included: "240", base: "95000", extraMinutes: "60",
        extraValue: "25000", tolerance: "10", notes: "Plan VIP con tarifa premium por habitacion."
      },
      promo: {
        name: "Promo lunes a jueves", type: "promocion", category: "Estandar", from: "1", to: "4",
        start: "00:00", end: "23:59", included: "180", base: "42000", extraMinutes: "60",
        extraValue: "12000", tolerance: "10", notes: "Promocion valida de lunes a jueves."
      }
    };
    var p = presets[type] || presets.express;
    $("planName").value = p.name;
    $("planType").value = p.type;
    $("roomCategory").value = p.category;
    $("dayFrom").value = p.from;
    $("dayTo").value = p.to;
    $("startTime").value = p.start;
    $("endTime").value = p.end;
    $("includedMinutes").value = p.included;
    $("baseValue").value = p.base;
    $("extraMinutes").value = p.extraMinutes;
    $("extraValue").value = p.extraValue;
    $("toleranceMinutes").value = p.tolerance;
    $("currency").value = "COP";
    $("chargeFraction").checked = true;
    $("autoApply").checked = true;
    $("notes").value = p.notes;
    setMessage("presetMsg", "Preset cargado: " + p.name + ".", false);
  }

  function getFilteredRates() {
    var needle = text($("filterText").value).toLowerCase();
    var type = text($("filterType").value).toLowerCase();
    var status = text($("filterStatus").value).toLowerCase();
    var stationID = number($("filterStation").value);
    return state.rates.filter(function (item) {
      var haystack = [
        item.nombre_plan,
        item.tipo_plan,
        item.categoria_habitacion,
        item.estacion_nombre,
        item.estacion_codigo,
        item.observaciones
      ].map(text).join(" ").toLowerCase();
      if (needle && haystack.indexOf(needle) === -1) return false;
      if (type && text(item.tipo_plan).toLowerCase() !== type) return false;
      if (status && text(item.estado || "activo").toLowerCase() !== status) return false;
      if (stationID > 0 && number(item.estacion_id) !== stationID) return false;
      return true;
    });
  }

  function updateKpis() {
    var active = state.rates.filter(function (item) { return text(item.estado || "activo").toLowerCase() === "activo"; });
    var rooms = new Set(active.map(function (item) { return number(item.estacion_id); }).filter(Boolean));
    var types = new Set(active.map(function (item) { return text(item.tipo_plan || "express").toLowerCase(); }).filter(Boolean));
    var avg = active.length ? active.reduce(function (sum, item) { return sum + number(item.valor_base); }, 0) / active.length : 0;
    var max = active.length ? active.reduce(function (best, item) { return Math.max(best, number(item.valor_base)); }, 0) : 0;
    $("kpiActive").textContent = String(active.length);
    $("kpiRooms").textContent = String(rooms.size);
    $("kpiAverage").textContent = money(avg, (active[0] && active[0].moneda) || "COP");
    $("kpiMax").textContent = money(max, (active[0] && active[0].moneda) || "COP");
    $("kpiTypes").textContent = String(types.size);
  }

  function fillSimPlanSelect() {
    var select = $("simPlan");
    var active = state.rates.filter(function (item) { return text(item.estado || "activo").toLowerCase() === "activo"; });
    var html = '<option value="">Selecciona un plan</option>';
    active.forEach(function (item) {
      html += '<option value="' + esc(item.id) + '">' + esc((item.nombre_plan || "Plan motel") + " - " + (item.estacion_nombre || ("Habitacion " + item.estacion_id))) + '</option>';
    });
    select.innerHTML = html;
  }

  function renderRates() {
    var root = $("ratesList");
    updateKpis();
    fillSimPlanSelect();
    var rows = getFilteredRates();
    if (!rows.length) {
      root.innerHTML = '<div class="motel-empty">No hay tarifas de motel para los filtros actuales.</div>';
      return;
    }
    root.innerHTML = '<table class="motel-table"><thead><tr><th>Plan</th><th>Habitacion</th><th>Vigencia</th><th>Base</th><th>Extra</th><th>Estado</th><th>Acciones</th></tr></thead><tbody>' +
      rows.map(function (item) {
        var status = text(item.estado || "activo").toLowerCase();
        var toggle = status === "activo" ? "desactivar" : "activar";
        var badgeClass = status === "activo" ? "motel-badge" : "motel-badge off";
        return "<tr>" +
          "<td><strong>" + esc(item.nombre_plan || "Plan motel") + "</strong><br><small>" + esc(typeLabel(item.tipo_plan)) + " | " + esc(item.categoria_habitacion || "General") + "</small></td>" +
          "<td><strong>" + esc(item.estacion_nombre || ("Habitacion " + item.estacion_id)) + "</strong><br><small>" + esc(item.estacion_codigo || defaultStationCode(state.empresaID, item.estacion_id)) + "</small></td>" +
          "<td>" + esc(dayRange(item.dia_semana_desde, item.dia_semana_hasta)) + "<br><small>" + esc((item.hora_inicio || "00:00") + " - " + (item.hora_fin || "23:59")) + "</small></td>" +
          "<td>" + esc(item.minutos_incluidos || 0) + " min<br><strong>" + esc(money(item.valor_base, item.moneda)) + "</strong><br><small>Tolerancia " + esc(item.tolerancia_minutos || 0) + " min</small></td>" +
          "<td>" + esc(item.minutos_extra || 0) + " min<br><strong>" + esc(money(item.valor_extra, item.moneda)) + "</strong><br><small>" + (item.cobrar_por_fraccion ? "Fraccion parcial" : "Bloque completo") + "</small></td>" +
          "<td><span class=\"" + badgeClass + "\">" + esc(status) + "</span><br><small>Prioridad " + esc(item.prioridad || 1) + "</small></td>" +
          "<td><button class=\"btn secondary\" type=\"button\" data-edit=\"" + esc(item.id) + "\">Editar</button> " +
          "<button class=\"btn secondary\" type=\"button\" data-toggle=\"" + esc(item.id) + "\" data-action=\"" + toggle + "\">" + (toggle === "activar" ? "Activar" : "Desactivar") + "</button> " +
          "<button class=\"btn danger\" type=\"button\" data-delete=\"" + esc(item.id) + "\">Eliminar</button></td>" +
        "</tr>";
      }).join("") + "</tbody></table>";
    bindRowActions(root);
  }

  function bindRowActions(root) {
    root.querySelectorAll("[data-edit]").forEach(function (btn) {
      btn.addEventListener("click", function () { editRate(number(btn.getAttribute("data-edit"))); });
    });
    root.querySelectorAll("[data-toggle]").forEach(function (btn) {
      btn.addEventListener("click", function () {
        toggleRate(number(btn.getAttribute("data-toggle")), btn.getAttribute("data-action"));
      });
    });
    root.querySelectorAll("[data-delete]").forEach(function (btn) {
      btn.addEventListener("click", function () { deleteRate(number(btn.getAttribute("data-delete"))); });
    });
  }

  function editRate(id) {
    var item = state.rates.find(function (row) { return number(row.id) === id; });
    if (!item) return;
    $("rateId").value = item.id || "";
    $("planName").value = item.nombre_plan || "";
    $("planType").value = item.tipo_plan || "express";
    $("stationSelect").value = item.estacion_id || "";
    $("stationId").value = item.estacion_id || "";
    $("stationCode").value = item.estacion_codigo || defaultStationCode(state.empresaID, item.estacion_id);
    $("stationName").value = item.estacion_nombre || "";
    $("roomCategory").value = item.categoria_habitacion || "";
    $("dayFrom").value = item.dia_semana_desde || 1;
    $("dayTo").value = item.dia_semana_hasta || 7;
    $("startTime").value = (item.hora_inicio || "00:00").slice(0, 5);
    $("endTime").value = (item.hora_fin || "23:59").slice(0, 5);
    $("includedMinutes").value = item.minutos_incluidos || 180;
    $("baseValue").value = item.valor_base || 0;
    $("extraMinutes").value = item.minutos_extra || 60;
    $("extraValue").value = item.valor_extra || 0;
    $("toleranceMinutes").value = item.tolerancia_minutos || 0;
    $("currency").value = item.moneda || "COP";
    $("priority").value = item.prioridad || 1;
    $("status").value = item.estado || "activo";
    $("chargeFraction").checked = item.cobrar_por_fraccion !== false;
    $("autoApply").checked = item.aplicar_automaticamente !== false;
    $("notes").value = item.observaciones || "";
    setMessage("formMsg", "Editando plan #" + id + ".", false);
    $("rateForm").scrollIntoView({ behavior: "smooth", block: "start" });
  }

  async function toggleRate(id, action) {
    if (!id || !action) return;
    try {
      await fetchJSON(endpoint({ id: String(id), action: action }), { method: "PUT" });
      setMessage("formMsg", "Estado actualizado.", false);
      await loadRates();
    } catch (e) {
      setMessage("formMsg", e.message || "No se pudo actualizar el estado.", true);
    }
  }

  async function deleteRate(id) {
    if (!id) return;
    if (!window.confirm("Eliminar esta tarifa de motel?")) return;
    try {
      await fetchJSON(endpoint({ id: String(id) }), { method: "DELETE" });
      setMessage("formMsg", "Tarifa eliminada.", false);
      await loadRates();
    } catch (e) {
      setMessage("formMsg", e.message || "No se pudo eliminar la tarifa.", true);
    }
  }

  async function simulateRate(event) {
    event.preventDefault();
    var planID = number($("simPlan").value);
    var minutes = number($("simMinutes").value);
    if (planID <= 0 || minutes <= 0) {
      $("simResult").textContent = "Selecciona un plan y minutos validos.";
      return;
    }
    try {
      var data = await fetchJSON(endpoint({ action: "calcular", id: String(planID), minutos_consumidos: String(minutes) }));
      var detail = data && data.detalle ? data.detalle : {};
      $("simResult").innerHTML = "<strong>Total estimado: " + esc(money(detail.monto_total, detail.moneda || "COP")) + "</strong><br>" +
        "Plan: " + esc(detail.nombre_plan || "Plan motel") + " | Tipo: " + esc(typeLabel(detail.tipo_plan)) + "<br>" +
        "Base: " + esc(money(detail.monto_base, detail.moneda || "COP")) + " | Extra: " + esc(money(detail.monto_extra, detail.moneda || "COP")) + "<br>" +
        "Bloques extra: " + esc(detail.bloques_extra || 0) + " | Minutos facturables: " + esc(detail.minutos_facturables || minutes);
    } catch (e) {
      $("simResult").textContent = e.message || "No se pudo simular.";
    }
  }

  async function loadRates() {
    if (state.empresaID <= 0) {
      $("ratesList").innerHTML = '<div class="motel-empty">Abre este modulo desde Administrar empresa para recibir el contexto de empresa.</div>';
      setMessage("formMsg", "No se pudo iniciar tarifas_de_motel sin empresa_id.", true);
      return;
    }
    try {
      state.rates = await fetchJSON(endpoint({ include_inactive: "1", limit: "2000" }));
      if (!Array.isArray(state.rates)) state.rates = [];
      renderRates();
    } catch (e) {
      $("ratesList").innerHTML = '<div class="motel-empty">' + esc(e.message || "No se pudieron cargar las tarifas.") + '</div>';
    }
  }

  async function loadAll() {
    if (state.empresaID <= 0) {
      await loadRates();
      return;
    }
    await loadStations();
    await loadRates();
  }

  function clearFilters() {
    $("filterText").value = "";
    $("filterType").value = "";
    $("filterStatus").value = "";
    $("filterStation").value = "";
    renderRates();
  }

  function bindEvents() {
    $("rateForm").addEventListener("submit", saveRate);
    $("simForm").addEventListener("submit", simulateRate);
    $("stationSelect").addEventListener("change", function () { syncStation(this.value); });
    $("stationId").addEventListener("input", function () {
      if (number(this.value) > 0 && !text($("stationCode").value)) syncStation(this.value);
    });
    $("resetForm").addEventListener("click", resetForm);
    $("applyAllRooms").addEventListener("click", applyAllRooms);
    $("reloadRates").addEventListener("click", loadAll);
    $("quickExpress").addEventListener("click", function () { applyPreset("express"); });
    $("quickNight").addEventListener("click", function () { applyPreset("amanecida"); });
    $("clearFilters").addEventListener("click", clearFilters);
    ["filterText", "filterType", "filterStatus", "filterStation"].forEach(function (id) {
      $(id).addEventListener(id === "filterText" ? "input" : "change", renderRates);
    });
    document.querySelectorAll("[data-preset]").forEach(function (btn) {
      btn.addEventListener("click", function () { applyPreset(btn.getAttribute("data-preset")); });
    });
  }

  function initDefaults() {
    fillDaySelect("dayFrom");
    fillDaySelect("dayTo");
    resetForm();
    $("simMinutes").value = "210";
  }

  document.addEventListener("DOMContentLoaded", function () {
    state.empresaID = resolveEmpresaID();
    initDefaults();
    bindEvents();
    loadAll();
  });
}());
