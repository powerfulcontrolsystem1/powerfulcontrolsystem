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
    nightRates: [],
    dayUseRates: [],
    motelRates: [],
    rules: null
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
      .replace(/\"/g, "&quot;")
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
    if (a === b) return a;
    return a + " - " + b;
  }

  function peopleRange(from, to) {
    var a = number(from, 1);
    var b = number(to);
    if (a <= 0) a = 1;
    if (b <= 0) return a + "+ personas";
    if (a === b) return a + " persona" + (a === 1 ? "" : "s");
    return a + " - " + b + " personas";
  }

  function fillDaySelect(id, includeAll) {
    var select = $(id);
    if (!select) return;
    var html = includeAll ? '<option value="">Todos</option>' : "";
    DAYS.forEach(function (day) {
      html += '<option value="' + day.id + '">' + esc(day.name) + '</option>';
    });
    select.innerHTML = html;
  }

  function fillStationSelect(id, includeEmpty) {
    var select = $(id);
    if (!select) return;
    var html = includeEmpty ? '<option value="">Selecciona una estacion</option>' : "";
    state.stations.forEach(function (station) {
      html += '<option value="' + station.id + '">' + esc(station.name + " (#" + station.id + ")") + '</option>';
    });
    if (!state.stations.length) {
      html += '<option value="">Sin estaciones detectadas, usa el ID manual</option>';
    }
    select.innerHTML = html;
  }

  function syncStation(kind, id) {
    var stationID = number(id);
    if (stationID <= 0) return;
    var station = findStation(stationID);
    var prefix = kind === "night" ? "night" : (kind === "motel" ? "motel" : "dayUse");
    $(prefix + "StationId").value = String(stationID);
    $(prefix + "StationCode").value = station && station.code ? station.code : defaultStationCode(state.empresaID, stationID);
    $(prefix + "StationName").value = station && station.name ? station.name : "Estacion " + stationID;
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
          byID.set(id, { id: id, code: defaultStationCode(state.empresaID, id), name: text(item && item.nombre) || "Estacion " + id });
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
          byID.set(id, { id: id, code: code || defaultStationCode(state.empresaID, id), name: text(item && item.nombre) || "Estacion " + id });
        });
      }
    } catch (e) {
      console.warn("No se pudieron cargar estaciones desde carritos", e);
    }

    state.stations = Array.from(byID.values()).sort(function (a, b) { return a.id - b.id; });
    ["nightStationSelect", "dayUseStationSelect", "motelStationSelect", "simNightStation", "simDayUseStation"].forEach(function (id) { fillStationSelect(id, true); });
  }

  function endpoint(path, extra) {
    var params = new URLSearchParams(extra || {});
    params.set("empresa_id", String(state.empresaID));
    return path + "?" + params.toString();
  }

  async function fetchJSON(url, options) {
    var resp = await fetch(url, Object.assign({ credentials: "same-origin" }, options || {}));
    await ensureOk(resp);
    return resp.status === 204 ? null : resp.json();
  }

  async function loadNightRates() {
    var rows = await fetchJSON(endpoint("/api/empresa/tarifas_por_dia", { include_inactive: "1", limit: "800" }));
    state.nightRates = Array.isArray(rows) ? rows : [];
    renderNightRates();
  }

  async function loadDayUseRates() {
    var rows = await fetchJSON(endpoint("/api/empresa/tarifas_por_minutos", { include_inactive: "1", limit: "800" }));
    state.dayUseRates = Array.isArray(rows) ? rows : [];
    renderDayUseRates();
  }

  async function loadMotelRates() {
    var rows = await fetchJSON(endpoint("/api/empresa/tarifas_motel", { include_inactive: "1", limit: "1000" }));
    state.motelRates = Array.isArray(rows) ? rows : [];
    renderMotelRates();
  }

  async function loadRules() {
    var cfg = await fetchJSON(endpoint("/api/empresa/tarifas_por_minutos", { action: "config" }));
    state.rules = cfg || {};
    fillRulesForm();
  }

  function updateKpis() {
    var activeNight = state.nightRates.filter(function (item) { return text(item.estado || "activo").toLowerCase() === "activo"; });
    var activeDayUse = state.dayUseRates.filter(function (item) { return text(item.estado || "activo").toLowerCase() === "activo"; });
    var activeMotel = state.motelRates.filter(function (item) { return text(item.estado || "activo").toLowerCase() === "activo"; });
    var avg = activeNight.length ? activeNight.reduce(function (sum, item) { return sum + number(item.valor_dia); }, 0) / activeNight.length : 0;
    $("kpiNightActive").textContent = String(activeNight.length);
    $("kpiDayUseActive").textContent = String(activeDayUse.length);
    $("kpiMotelActive").textContent = String(activeMotel.length);
    $("kpiNightAverage").textContent = money(avg, (activeNight[0] && activeNight[0].moneda) || "COP");
    $("kpiDailyCap").textContent = number(state.rules && state.rules.monto_maximo_diario) > 0 ? money(state.rules.monto_maximo_diario, "COP") : "Sin tope";
  }

  function renderNightRates() {
    var root = $("nightList");
    if (!state.nightRates.length) {
      root.innerHTML = '<div class="hotel-empty">Todavia no hay tarifas por noche. Crea una tarifa para cada estacion o aplica una tarifa base a todas.</div>';
      updateKpis();
      return;
    }
    var rows = state.nightRates.map(function (item) {
      var status = text(item.estado || "activo").toLowerCase();
      var toggle = status === "activo" ? "desactivar" : "activar";
      return "<tr>" +
        "<td><strong>" + esc(item.estacion_nombre || ("Estacion " + item.estacion_id)) + "</strong><br><small>" + esc(item.estacion_codigo || defaultStationCode(state.empresaID, item.estacion_id)) + "</small></td>" +
        "<td>" + money(item.valor_dia, item.moneda) + "<br><small>" + esc(item.nombre_tarifa || item.servicio_nombre || "hospedaje") + "</small></td>" +
        "<td>" + esc(peopleRange(item.personas_desde, item.personas_hasta)) + "</td>" +
        "<td>" + esc(item.hora_check_in || "15:00") + " / " + esc(item.hora_check_out || "12:00") + "</td>" +
        "<td><span class=\"hotel-badge\">" + esc(status) + "</span><br><small>Prioridad " + esc(item.prioridad || 1) + "</small></td>" +
        "<td>" + (item.aplicar_automaticamente ? "Automatico" : "Manual") + "</td>" +
        "<td><button class=\"btn secondary\" type=\"button\" data-edit-night=\"" + esc(item.id) + "\">Editar</button> " +
        "<button class=\"btn secondary\" type=\"button\" data-toggle-night=\"" + esc(item.id) + "\" data-action=\"" + toggle + "\">" + (toggle === "activar" ? "Activar" : "Desactivar") + "</button> " +
        "<button class=\"btn danger\" type=\"button\" data-delete-night=\"" + esc(item.id) + "\">Eliminar</button></td>" +
      "</tr>";
    }).join("");
    root.innerHTML = '<table class="hotel-table"><thead><tr><th>Estacion</th><th>Valor</th><th>Ocupacion</th><th>Horario</th><th>Estado</th><th>Aplicacion</th><th>Acciones</th></tr></thead><tbody>' + rows + '</tbody></table>';
    bindTableActions(root);
    updateKpis();
  }

  function renderDayUseRates() {
    var root = $("dayUseList");
    if (!state.dayUseRates.length) {
      root.innerHTML = '<div class="hotel-empty">Todavia no hay reglas day-use o por fracciones. Crea una regla para permanencias cortas.</div>';
      updateKpis();
      return;
    }
    var rows = state.dayUseRates.map(function (item) {
      var status = text(item.estado || "activo").toLowerCase();
      var toggle = status === "activo" ? "desactivar" : "activar";
      return "<tr>" +
        "<td><strong>" + esc(item.estacion_nombre || ("Estacion " + item.estacion_id)) + "</strong><br><small>" + esc(item.estacion_codigo || defaultStationCode(state.empresaID, item.estacion_id)) + "</small></td>" +
        "<td>" + esc(dayRange(item.dia_semana_desde, item.dia_semana_hasta)) + "</td>" +
        "<td>" + esc(item.minutos_base || 0) + " min / " + money(item.valor_base, item.moneda) + "</td>" +
        "<td>" + esc(item.minutos_extra || 0) + " min / " + money(item.valor_extra, item.moneda) + "<br><small>" + (item.cobrar_por_fraccion ? "Por fraccion" : "Bloque completo") + "</small></td>" +
        "<td><span class=\"hotel-badge warn\">" + esc(status) + "</span><br><small>Prioridad " + esc(item.prioridad || 1) + "</small></td>" +
        "<td><button class=\"btn secondary\" type=\"button\" data-edit-dayuse=\"" + esc(item.id) + "\">Editar</button> " +
        "<button class=\"btn secondary\" type=\"button\" data-toggle-dayuse=\"" + esc(item.id) + "\" data-action=\"" + toggle + "\">" + (toggle === "activar" ? "Activar" : "Desactivar") + "</button> " +
        "<button class=\"btn danger\" type=\"button\" data-delete-dayuse=\"" + esc(item.id) + "\">Eliminar</button></td>" +
      "</tr>";
    }).join("");
    root.innerHTML = '<table class="hotel-table"><thead><tr><th>Estacion</th><th>Dias</th><th>Base</th><th>Extra</th><th>Estado</th><th>Acciones</th></tr></thead><tbody>' + rows + '</tbody></table>';
    bindTableActions(root);
    updateKpis();
  }

  function motelTypeLabel(value) {
    var type = text(value || "express").toLowerCase();
    var labels = {
      express: "Express",
      day_use: "Day-use",
      nocturno: "Nocturno",
      amanecida: "Amanecida",
      suite: "Suite",
      vip: "VIP",
      promocion: "Promocion"
    };
    return labels[type] || "Express";
  }

  function fillMotelPlanSelect() {
    var select = $("simMotelPlan");
    if (!select) return;
    var active = state.motelRates.filter(function (item) {
      return text(item.estado || "activo").toLowerCase() === "activo";
    });
    var html = '<option value="">Selecciona un plan motel</option>';
    active.forEach(function (item) {
      html += '<option value="' + esc(item.id) + '">' + esc((item.nombre_plan || "Plan motel") + " - " + (item.estacion_nombre || ("Estacion " + item.estacion_id))) + '</option>';
    });
    select.innerHTML = html;
  }

  function renderMotelRates() {
    var root = $("motelList");
    if (!root) return;
    if (!state.motelRates.length) {
      root.innerHTML = '<div class="hotel-empty">Todavia no hay planes motel. Usa los presets express o amanecida para iniciar una matriz profesional por estacion.</div>';
      fillMotelPlanSelect();
      updateKpis();
      return;
    }
    var rows = state.motelRates.map(function (item) {
      var status = text(item.estado || "activo").toLowerCase();
      var toggle = status === "activo" ? "desactivar" : "activar";
      return "<tr>" +
        "<td><strong>" + esc(item.nombre_plan || "Plan motel") + "</strong><br><small>" + esc(motelTypeLabel(item.tipo_plan)) + " | " + esc(item.categoria_habitacion || "General") + "</small></td>" +
        "<td><strong>" + esc(item.estacion_nombre || ("Estacion " + item.estacion_id)) + "</strong><br><small>" + esc(item.estacion_codigo || defaultStationCode(state.empresaID, item.estacion_id)) + "</small></td>" +
        "<td>" + esc(dayRange(item.dia_semana_desde, item.dia_semana_hasta)) + "<br><small>" + esc((item.hora_inicio || "00:00") + " - " + (item.hora_fin || "23:59")) + "</small></td>" +
        "<td>" + esc(item.minutos_incluidos || 0) + " min / " + money(item.valor_base, item.moneda) + "<br><small>Tolerancia " + esc(item.tolerancia_minutos || 0) + " min</small></td>" +
        "<td>" + esc(item.minutos_extra || 0) + " min / " + money(item.valor_extra, item.moneda) + "<br><small>" + (item.cobrar_por_fraccion ? "Fraccion parcial" : "Solo bloque completo") + "</small></td>" +
        "<td><span class=\"hotel-badge warn\">" + esc(status) + "</span><br><small>Prioridad " + esc(item.prioridad || 1) + "</small></td>" +
        "<td><button class=\"btn secondary\" type=\"button\" data-edit-motel=\"" + esc(item.id) + "\">Editar</button> " +
        "<button class=\"btn secondary\" type=\"button\" data-toggle-motel=\"" + esc(item.id) + "\" data-action=\"" + toggle + "\">" + (toggle === "activar" ? "Activar" : "Desactivar") + "</button> " +
        "<button class=\"btn danger\" type=\"button\" data-delete-motel=\"" + esc(item.id) + "\">Eliminar</button></td>" +
      "</tr>";
    }).join("");
    root.innerHTML = '<table class="hotel-table"><thead><tr><th>Plan</th><th>Estacion</th><th>Vigencia</th><th>Base</th><th>Extra</th><th>Estado</th><th>Acciones</th></tr></thead><tbody>' + rows + '</tbody></table>';
    bindTableActions(root);
    fillMotelPlanSelect();
    updateKpis();
  }

  function bindTableActions(root) {
    root.querySelectorAll("[data-edit-night]").forEach(function (btn) {
      btn.addEventListener("click", function () { editNight(number(btn.getAttribute("data-edit-night"))); });
    });
    root.querySelectorAll("[data-delete-night]").forEach(function (btn) {
      btn.addEventListener("click", function () { removeRate("night", number(btn.getAttribute("data-delete-night"))); });
    });
    root.querySelectorAll("[data-toggle-night]").forEach(function (btn) {
      btn.addEventListener("click", function () { toggleRate("night", number(btn.getAttribute("data-toggle-night")), btn.getAttribute("data-action")); });
    });
    root.querySelectorAll("[data-edit-dayuse]").forEach(function (btn) {
      btn.addEventListener("click", function () { editDayUse(number(btn.getAttribute("data-edit-dayuse"))); });
    });
    root.querySelectorAll("[data-delete-dayuse]").forEach(function (btn) {
      btn.addEventListener("click", function () { removeRate("dayuse", number(btn.getAttribute("data-delete-dayuse"))); });
    });
    root.querySelectorAll("[data-toggle-dayuse]").forEach(function (btn) {
      btn.addEventListener("click", function () { toggleRate("dayuse", number(btn.getAttribute("data-toggle-dayuse")), btn.getAttribute("data-action")); });
    });
    root.querySelectorAll("[data-edit-motel]").forEach(function (btn) {
      btn.addEventListener("click", function () { editMotel(number(btn.getAttribute("data-edit-motel"))); });
    });
    root.querySelectorAll("[data-delete-motel]").forEach(function (btn) {
      btn.addEventListener("click", function () { removeRate("motel", number(btn.getAttribute("data-delete-motel"))); });
    });
    root.querySelectorAll("[data-toggle-motel]").forEach(function (btn) {
      btn.addEventListener("click", function () { toggleRate("motel", number(btn.getAttribute("data-toggle-motel")), btn.getAttribute("data-action")); });
    });
  }

  function resetNight() {
    $("nightId").value = "";
    $("nightName").value = "tarifa_hotel_cama_doble";
    $("nightStationSelect").value = "";
    $("nightStationId").value = "";
    $("nightStationCode").value = "";
    $("nightStationName").value = "";
    $("nightService").value = "hospedaje";
    $("nightValue").value = "";
    $("nightPeopleFrom").value = "1";
    $("nightPeopleTo").value = "0";
    $("nightCurrency").value = "COP";
    $("nightCheckIn").value = "15:00";
    $("nightCheckOut").value = "12:00";
    $("nightPriority").value = "1";
    $("nightStatus").value = "activo";
    $("nightAutomatic").checked = true;
    $("nightNotes").value = "";
    setMessage("nightMsg", "", false);
  }

  function resetDayUse() {
    $("dayUseId").value = "";
    $("dayUseStationSelect").value = "";
    $("dayUseStationId").value = "";
    $("dayUseStationCode").value = "";
    $("dayUseStationName").value = "";
    $("dayUseCurrency").value = "COP";
    $("dayUseDayFrom").value = "1";
    $("dayUseDayTo").value = "7";
    $("dayUseBaseMinutes").value = "180";
    $("dayUseBaseValue").value = "";
    $("dayUseExtraMinutes").value = "60";
    $("dayUseExtraValue").value = "";
    $("dayUsePriority").value = "1";
    $("dayUseStatus").value = "activo";
    $("dayUseFraction").checked = false;
    $("dayUseNotes").value = "";
    setMessage("dayUseMsg", "", false);
  }

  function buildNightPayload() {
    var stationID = number($("nightStationId").value);
    return {
      id: number($("nightId").value),
      empresa_id: state.empresaID,
      nombre_tarifa: text($("nightName").value) || "tarifa_hotel_cama_doble",
      estacion_id: stationID,
      estacion_codigo: text($("nightStationCode").value) || defaultStationCode(state.empresaID, stationID),
      estacion_nombre: text($("nightStationName").value) || "Estacion " + stationID,
      servicio_nombre: text($("nightService").value) || "hospedaje",
      valor_dia: number($("nightValue").value),
      personas_desde: number($("nightPeopleFrom").value, 1),
      personas_hasta: number($("nightPeopleTo").value),
      hora_check_in: text($("nightCheckIn").value) || "15:00",
      hora_check_out: text($("nightCheckOut").value) || "12:00",
      moneda: text($("nightCurrency").value) || "COP",
      prioridad: number($("nightPriority").value, 1),
      aplicar_automaticamente: !!$("nightAutomatic").checked,
      estado: text($("nightStatus").value) || "activo",
      observaciones: text($("nightNotes").value)
    };
  }

  function buildDayUsePayload() {
    var stationID = number($("dayUseStationId").value);
    return {
      id: number($("dayUseId").value),
      empresa_id: state.empresaID,
      estacion_id: stationID,
      estacion_codigo: text($("dayUseStationCode").value) || defaultStationCode(state.empresaID, stationID),
      estacion_nombre: text($("dayUseStationName").value) || "Estacion " + stationID,
      dia_semana_desde: number($("dayUseDayFrom").value, 1),
      dia_semana_hasta: number($("dayUseDayTo").value, 7),
      minutos_base: number($("dayUseBaseMinutes").value, 180),
      valor_base: number($("dayUseBaseValue").value),
      minutos_extra: number($("dayUseExtraMinutes").value, 60),
      valor_extra: number($("dayUseExtraValue").value),
      cobrar_por_fraccion: !!$("dayUseFraction").checked,
      moneda: text($("dayUseCurrency").value) || "COP",
      prioridad: number($("dayUsePriority").value, 1),
      estado: text($("dayUseStatus").value) || "activo",
      observaciones: text($("dayUseNotes").value)
    };
  }

  function validateStationPayload(payload, msgID) {
    if (state.empresaID <= 0) {
      setMessage(msgID, "No se encontro empresa_id para guardar.", true);
      return false;
    }
    if (payload.estacion_id <= 0) {
      setMessage(msgID, "Selecciona una estacion o escribe el ID de estacion.", true);
      return false;
    }
    return true;
  }

  async function saveNight(event) {
    event.preventDefault();
    var payload = buildNightPayload();
    if (!validateStationPayload(payload, "nightMsg")) return;
    if (payload.personas_desde <= 0) return setMessage("nightMsg", "Personas desde debe ser mayor a cero.", true);
    if (payload.personas_hasta > 0 && payload.personas_hasta < payload.personas_desde) return setMessage("nightMsg", "Personas hasta no puede ser menor que personas desde.", true);
    setMessage("nightMsg", "Guardando tarifa por noche...", false);
    try {
      await fetchJSON("/api/empresa/tarifas_por_dia", {
        method: payload.id > 0 ? "PUT" : "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });
      setMessage("nightMsg", "Tarifa guardada correctamente.", false);
      resetNight();
      await loadNightRates();
    } catch (e) {
      setMessage("nightMsg", e.message || "No se pudo guardar la tarifa.", true);
    }
  }

  async function saveDayUse(event) {
    event.preventDefault();
    var payload = buildDayUsePayload();
    if (!validateStationPayload(payload, "dayUseMsg")) return;
    setMessage("dayUseMsg", "Guardando regla day-use...", false);
    try {
      await fetchJSON("/api/empresa/tarifas_por_minutos", {
        method: payload.id > 0 ? "PUT" : "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });
      setMessage("dayUseMsg", "Regla guardada correctamente.", false);
      resetDayUse();
      await loadDayUseRates();
    } catch (e) {
      setMessage("dayUseMsg", e.message || "No se pudo guardar la regla.", true);
    }
  }

  async function applyAll(kind) {
    var isNight = kind === "night";
    var payload = isNight ? buildNightPayload() : buildDayUsePayload();
    var msgID = isNight ? "nightMsg" : "dayUseMsg";
    if (state.empresaID <= 0) return setMessage(msgID, "No se encontro empresa_id para guardar.", true);
    if (isNight && payload.personas_hasta > 0 && payload.personas_hasta < payload.personas_desde) return setMessage(msgID, "Personas hasta no puede ser menor que personas desde.", true);
    setMessage(msgID, "Aplicando regla a todas las estaciones...", false);
    try {
      await fetchJSON(endpoint(isNight ? "/api/empresa/tarifas_por_dia" : "/api/empresa/tarifas_por_minutos", { action: "aplicar_todas_estaciones" }), {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });
      setMessage(msgID, "Regla aplicada a todas las estaciones disponibles.", false);
      if (isNight) await loadNightRates(); else await loadDayUseRates();
    } catch (e) {
      setMessage(msgID, e.message || "No se pudo aplicar a todas.", true);
    }
  }

  function editNight(id) {
    var item = state.nightRates.find(function (row) { return number(row.id) === id; });
    if (!item) return;
    showTab("nightSection");
    $("nightId").value = item.id || "";
    $("nightName").value = item.nombre_tarifa || item.servicio_nombre || "tarifa_hotel_cama_doble";
    $("nightStationSelect").value = item.estacion_id || "";
    $("nightStationId").value = item.estacion_id || "";
    $("nightStationCode").value = item.estacion_codigo || defaultStationCode(state.empresaID, item.estacion_id);
    $("nightStationName").value = item.estacion_nombre || "";
    $("nightService").value = item.servicio_nombre || "hospedaje";
    $("nightValue").value = item.valor_dia || 0;
    $("nightPeopleFrom").value = item.personas_desde || 1;
    $("nightPeopleTo").value = item.personas_hasta || 0;
    $("nightCurrency").value = item.moneda || "COP";
    $("nightCheckIn").value = (item.hora_check_in || "15:00").slice(0, 5);
    $("nightCheckOut").value = (item.hora_check_out || "12:00").slice(0, 5);
    $("nightPriority").value = item.prioridad || 1;
    $("nightStatus").value = item.estado || "activo";
    $("nightAutomatic").checked = item.aplicar_automaticamente !== false;
    $("nightNotes").value = item.observaciones || "";
    setMessage("nightMsg", "Editando tarifa #" + id + ".", false);
    $("nightForm").scrollIntoView({ behavior: "smooth", block: "start" });
  }

  function editDayUse(id) {
    var item = state.dayUseRates.find(function (row) { return number(row.id) === id; });
    if (!item) return;
    showTab("dayUseSection");
    $("dayUseId").value = item.id || "";
    $("dayUseStationSelect").value = item.estacion_id || "";
    $("dayUseStationId").value = item.estacion_id || "";
    $("dayUseStationCode").value = item.estacion_codigo || defaultStationCode(state.empresaID, item.estacion_id);
    $("dayUseStationName").value = item.estacion_nombre || "";
    $("dayUseCurrency").value = item.moneda || "COP";
    $("dayUseDayFrom").value = item.dia_semana_desde || 1;
    $("dayUseDayTo").value = item.dia_semana_hasta || 7;
    $("dayUseBaseMinutes").value = item.minutos_base || 180;
    $("dayUseBaseValue").value = item.valor_base || 0;
    $("dayUseExtraMinutes").value = item.minutos_extra || 60;
    $("dayUseExtraValue").value = item.valor_extra || 0;
    $("dayUsePriority").value = item.prioridad || 1;
    $("dayUseStatus").value = item.estado || "activo";
    $("dayUseFraction").checked = !!item.cobrar_por_fraccion;
    $("dayUseNotes").value = item.observaciones || "";
    setMessage("dayUseMsg", "Editando regla #" + id + ".", false);
    $("dayUseForm").scrollIntoView({ behavior: "smooth", block: "start" });
  }

  function resetMotel() {
    $("motelId").value = "";
    $("motelPlanName").value = "";
    $("motelPlanType").value = "express";
    $("motelStationSelect").value = "";
    $("motelStationId").value = "";
    $("motelStationCode").value = "";
    $("motelStationName").value = "";
    $("motelCategory").value = "";
    $("motelDayFrom").value = "1";
    $("motelDayTo").value = "7";
    $("motelStartTime").value = "00:00";
    $("motelEndTime").value = "23:59";
    $("motelIncludedMinutes").value = "180";
    $("motelBaseValue").value = "";
    $("motelExtraMinutes").value = "60";
    $("motelExtraValue").value = "";
    $("motelTolerance").value = "10";
    $("motelCurrency").value = "COP";
    $("motelPriority").value = "1";
    $("motelStatus").value = "activo";
    $("motelFraction").checked = true;
    $("motelAutomatic").checked = true;
    $("motelNotes").value = "";
    setMessage("motelMsg", "", false);
  }

  function applyMotelPreset(type) {
    if (type === "night") {
      $("motelPlanName").value = "Amanecida";
      $("motelPlanType").value = "amanecida";
      $("motelStartTime").value = "21:00";
      $("motelEndTime").value = "08:00";
      $("motelIncludedMinutes").value = "660";
      $("motelBaseValue").value = $("motelBaseValue").value || "120000";
      $("motelExtraMinutes").value = "60";
      $("motelExtraValue").value = $("motelExtraValue").value || "20000";
      $("motelTolerance").value = "15";
      $("motelNotes").value = "Preset amanecida con cobro por hora adicional.";
      return;
    }
    $("motelPlanName").value = "Express 3 horas";
    $("motelPlanType").value = "express";
    $("motelStartTime").value = "00:00";
    $("motelEndTime").value = "23:59";
    $("motelIncludedMinutes").value = "180";
    $("motelBaseValue").value = $("motelBaseValue").value || "50000";
    $("motelExtraMinutes").value = "60";
    $("motelExtraValue").value = $("motelExtraValue").value || "15000";
    $("motelTolerance").value = "10";
    $("motelNotes").value = "Preset express para permanencias cortas.";
  }

  function buildMotelPayload() {
    var stationID = number($("motelStationId").value);
    return {
      id: number($("motelId").value),
      empresa_id: state.empresaID,
      estacion_id: stationID,
      estacion_codigo: text($("motelStationCode").value) || defaultStationCode(state.empresaID, stationID),
      estacion_nombre: text($("motelStationName").value) || "Estacion " + stationID,
      nombre_plan: text($("motelPlanName").value) || "Plan express",
      tipo_plan: text($("motelPlanType").value) || "express",
      categoria_habitacion: text($("motelCategory").value),
      dia_semana_desde: number($("motelDayFrom").value, 1),
      dia_semana_hasta: number($("motelDayTo").value, 7),
      hora_inicio: text($("motelStartTime").value) || "00:00",
      hora_fin: text($("motelEndTime").value) || "23:59",
      minutos_incluidos: number($("motelIncludedMinutes").value, 180),
      valor_base: number($("motelBaseValue").value),
      minutos_extra: number($("motelExtraMinutes").value, 60),
      valor_extra: number($("motelExtraValue").value),
      cobrar_por_fraccion: !!$("motelFraction").checked,
      tolerancia_minutos: number($("motelTolerance").value),
      moneda: text($("motelCurrency").value) || "COP",
      prioridad: number($("motelPriority").value, 1),
      aplicar_automaticamente: !!$("motelAutomatic").checked,
      estado: text($("motelStatus").value) || "activo",
      observaciones: text($("motelNotes").value)
    };
  }

  async function saveMotel(event) {
    event.preventDefault();
    var payload = buildMotelPayload();
    if (!validateStationPayload(payload, "motelMsg")) return;
    setMessage("motelMsg", "Guardando plan motel...", false);
    try {
      await fetchJSON("/api/empresa/tarifas_motel", {
        method: payload.id > 0 ? "PUT" : "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });
      setMessage("motelMsg", "Plan motel guardado correctamente.", false);
      resetMotel();
      await loadMotelRates();
    } catch (e) {
      setMessage("motelMsg", e.message || "No se pudo guardar el plan motel.", true);
    }
  }

  function editMotel(id) {
    var item = state.motelRates.find(function (row) { return number(row.id) === id; });
    if (!item) return;
    showTab("motelSection");
    $("motelId").value = item.id || "";
    $("motelPlanName").value = item.nombre_plan || "";
    $("motelPlanType").value = item.tipo_plan || "express";
    $("motelStationSelect").value = item.estacion_id || "";
    $("motelStationId").value = item.estacion_id || "";
    $("motelStationCode").value = item.estacion_codigo || defaultStationCode(state.empresaID, item.estacion_id);
    $("motelStationName").value = item.estacion_nombre || "";
    $("motelCategory").value = item.categoria_habitacion || "";
    $("motelDayFrom").value = item.dia_semana_desde || 1;
    $("motelDayTo").value = item.dia_semana_hasta || 7;
    $("motelStartTime").value = (item.hora_inicio || "00:00").slice(0, 5);
    $("motelEndTime").value = (item.hora_fin || "23:59").slice(0, 5);
    $("motelIncludedMinutes").value = item.minutos_incluidos || 180;
    $("motelBaseValue").value = item.valor_base || 0;
    $("motelExtraMinutes").value = item.minutos_extra || 60;
    $("motelExtraValue").value = item.valor_extra || 0;
    $("motelTolerance").value = item.tolerancia_minutos || 0;
    $("motelCurrency").value = item.moneda || "COP";
    $("motelPriority").value = item.prioridad || 1;
    $("motelStatus").value = item.estado || "activo";
    $("motelFraction").checked = item.cobrar_por_fraccion !== false;
    $("motelAutomatic").checked = item.aplicar_automaticamente !== false;
    $("motelNotes").value = item.observaciones || "";
    setMessage("motelMsg", "Editando plan motel #" + id + ".", false);
    $("motelForm").scrollIntoView({ behavior: "smooth", block: "start" });
  }

  async function toggleRate(kind, id, action) {
    if (!id || !action) return;
    var isNight = kind === "night";
    var isMotel = kind === "motel";
    var msgID = isNight ? "nightMsg" : (isMotel ? "motelMsg" : "dayUseMsg");
    try {
      await fetchJSON(endpoint(isNight ? "/api/empresa/tarifas_por_dia" : (isMotel ? "/api/empresa/tarifas_motel" : "/api/empresa/tarifas_por_minutos"), { id: String(id), action: action }), { method: "PUT" });
      setMessage(msgID, "Estado actualizado.", false);
      if (isNight) await loadNightRates(); else if (isMotel) await loadMotelRates(); else await loadDayUseRates();
    } catch (e) {
      setMessage(msgID, e.message || "No se pudo actualizar el estado.", true);
    }
  }

  async function removeRate(kind, id) {
    if (!id) return;
    if (!window.confirm("Eliminar esta tarifa?")) return;
    var isNight = kind === "night";
    var isMotel = kind === "motel";
    var msgID = isNight ? "nightMsg" : (isMotel ? "motelMsg" : "dayUseMsg");
    try {
      await fetchJSON(endpoint(isNight ? "/api/empresa/tarifas_por_dia" : (isMotel ? "/api/empresa/tarifas_motel" : "/api/empresa/tarifas_por_minutos"), { id: String(id) }), { method: "DELETE" });
      setMessage(msgID, "Tarifa eliminada.", false);
      if (isNight) await loadNightRates(); else if (isMotel) await loadMotelRates(); else await loadDayUseRates();
    } catch (e) {
      setMessage(msgID, e.message || "No se pudo eliminar la tarifa.", true);
    }
  }

  function fillRulesForm() {
    var cfg = state.rules || {};
    $("ruleRoundingMode").value = cfg.redondeo_modo || "ninguno";
    $("ruleRoundingUnit").value = cfg.redondeo_unidad || 100;
    $("ruleDailyMin").value = cfg.monto_minimo_diario || 0;
    $("ruleDailyMax").value = cfg.monto_maximo_diario || 0;
    $("ruleTolerance").value = cfg.margen_tolerancia_entrada_minutos || 0;
    $("ruleCancelMinutes").value = cfg.margen_desactivacion_minutos || 0;
    $("ruleSensorAuto").checked = !!cfg.sensor_auto_activar_estacion;
    $("ruleCancelEnabled").checked = !!cfg.margen_desactivacion_habilitado;
    $("ruleNotes").value = cfg.observaciones || "";
    updateKpis();
  }

  async function saveRules(event) {
    event.preventDefault();
    setMessage("rulesMsg", "Guardando reglas globales...", false);
    var payload = {
      empresa_id: state.empresaID,
      redondeo_modo: text($("ruleRoundingMode").value) || "ninguno",
      redondeo_unidad: number($("ruleRoundingUnit").value, 100),
      monto_minimo_diario: number($("ruleDailyMin").value),
      monto_maximo_diario: number($("ruleDailyMax").value),
      margen_tolerancia_entrada_minutos: number($("ruleTolerance").value),
      sensor_auto_activar_estacion: !!$("ruleSensorAuto").checked,
      margen_desactivacion_habilitado: !!$("ruleCancelEnabled").checked,
      margen_desactivacion_minutos: number($("ruleCancelMinutes").value),
      estado: "activo",
      observaciones: text($("ruleNotes").value)
    };
    try {
      state.rules = await fetchJSON(endpoint("/api/empresa/tarifas_por_minutos", { action: "config" }), {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });
      setMessage("rulesMsg", "Reglas guardadas correctamente.", false);
      fillRulesForm();
    } catch (e) {
      setMessage("rulesMsg", e.message || "No se pudieron guardar las reglas.", true);
    }
  }

  function toDateTimeLocal(date) {
    var pad = function (n) { return String(n).padStart(2, "0"); };
    return date.getFullYear() + "-" + pad(date.getMonth() + 1) + "-" + pad(date.getDate()) + "T" + pad(date.getHours()) + ":" + pad(date.getMinutes());
  }

  async function simulateNight(event) {
    event.preventDefault();
    var stationID = number($("simNightStation").value);
    var people = number($("simNightPeople").value, 1);
    var start = text($("simNightStart").value);
    var end = text($("simNightEnd").value);
    if (stationID <= 0 || !start || !end) {
      $("nightSimResult").textContent = "Selecciona estacion y fechas validas.";
      return;
    }
    try {
      var data = await fetchJSON(endpoint("/api/empresa/tarifas_por_dia", { action: "calcular", estacion_id: stationID, personas: people, activado_en: start, fecha_corte: end }));
      $("nightSimResult").innerHTML = "<strong>Total estimado: " + esc(money(data.monto_total, data.moneda || (data.tarifa && data.tarifa.moneda))) + "</strong><br>" +
        "Dias cobrados: " + esc(data.dias_cobrados || 0) + " | Dias equivalentes: " + esc(data.dias_equivalentes || 0) + "<br>" +
        "Tarifa base: " + esc(money(data.valor_dia, data.moneda || "COP")) + " | Ocupacion: " + esc(peopleRange(data.personas_desde, data.personas_hasta));
    } catch (e) {
      $("nightSimResult").textContent = e.message || "No se pudo simular.";
    }
  }

  async function simulateDayUse(event) {
    event.preventDefault();
    var stationID = number($("simDayUseStation").value);
    var day = number($("simDayUseDay").value, 1);
    var minutes = number($("simDayUseMinutes").value);
    if (stationID <= 0 || minutes <= 0) {
      $("dayUseSimResult").textContent = "Selecciona estacion y minutos validos.";
      return;
    }
    try {
      var data = await fetchJSON(endpoint("/api/empresa/tarifas_por_minutos", { action: "calcular", estacion_id: stationID, dia_semana: day, minutos_consumidos: minutes }));
      $("dayUseSimResult").innerHTML = "<strong>Total estimado: " + esc(money(data.monto_total, data.moneda || (data.tarifa && data.tarifa.moneda))) + "</strong><br>" +
        "Base: " + esc(money(data.monto_base, data.moneda || "COP")) + " | Extra: " + esc(money(data.monto_extra, data.moneda || "COP")) + "<br>" +
        "Bloques extra: " + esc(data.bloques_extra || 0) + " | Minutos facturables: " + esc(data.minutos_facturables || minutes);
    } catch (e) {
      $("dayUseSimResult").textContent = e.message || "No se pudo simular.";
    }
  }

  async function simulateMotel(event) {
    event.preventDefault();
    var planID = number($("simMotelPlan").value);
    var minutes = number($("simMotelMinutes").value);
    if (planID <= 0 || minutes <= 0) {
      $("motelSimResult").textContent = "Selecciona un plan motel y minutos validos.";
      return;
    }
    try {
      var data = await fetchJSON(endpoint("/api/empresa/tarifas_motel", { action: "calcular", id: planID, minutos_consumidos: minutes }));
      var detail = data && data.detalle ? data.detalle : {};
      $("motelSimResult").innerHTML = "<strong>Total estimado: " + esc(money(detail.monto_total, detail.moneda || "COP")) + "</strong><br>" +
        "Plan: " + esc(detail.nombre_plan || "Plan motel") + " | Tipo: " + esc(motelTypeLabel(detail.tipo_plan)) + "<br>" +
        "Base: " + esc(money(detail.monto_base, detail.moneda || "COP")) + " | Extra: " + esc(money(detail.monto_extra, detail.moneda || "COP")) + "<br>" +
        "Bloques extra: " + esc(detail.bloques_extra || 0) + " | Minutos facturables: " + esc(detail.minutos_facturables || minutes);
    } catch (e) {
      $("motelSimResult").textContent = e.message || "No se pudo simular el plan motel.";
    }
  }

  function showTab(id) {
    document.querySelectorAll(".hotel-section").forEach(function (section) {
      section.classList.toggle("is-active", section.id === id);
    });
    document.querySelectorAll(".hotel-tab").forEach(function (tab) {
      tab.classList.toggle("is-active", tab.getAttribute("data-tab") === id);
    });
  }

  function bindEvents() {
    document.querySelectorAll(".hotel-tab").forEach(function (tab) {
      tab.addEventListener("click", function () { showTab(tab.getAttribute("data-tab")); });
    });
    document.querySelectorAll("[data-scroll-target]").forEach(function (btn) {
      btn.addEventListener("click", function () {
        var id = btn.getAttribute("data-scroll-target");
        showTab(id);
        var el = $(id);
        if (el) el.scrollIntoView({ behavior: "smooth", block: "start" });
      });
    });
    $("nightStationSelect").addEventListener("change", function () { syncStation("night", this.value); });
    $("dayUseStationSelect").addEventListener("change", function () { syncStation("dayuse", this.value); });
    $("motelStationSelect").addEventListener("change", function () { syncStation("motel", this.value); });
    $("nightStationId").addEventListener("input", function () { if (number(this.value) > 0 && !text($("nightStationCode").value)) syncStation("night", this.value); });
    $("dayUseStationId").addEventListener("input", function () { if (number(this.value) > 0 && !text($("dayUseStationCode").value)) syncStation("dayuse", this.value); });
    $("motelStationId").addEventListener("input", function () { if (number(this.value) > 0 && !text($("motelStationCode").value)) syncStation("motel", this.value); });
    $("nightForm").addEventListener("submit", saveNight);
    $("dayUseForm").addEventListener("submit", saveDayUse);
    $("motelForm").addEventListener("submit", saveMotel);
    $("rulesForm").addEventListener("submit", saveRules);
    $("nightReset").addEventListener("click", resetNight);
    $("dayUseReset").addEventListener("click", resetDayUse);
    $("motelReset").addEventListener("click", resetMotel);
    $("motelPresetExpress").addEventListener("click", function () { applyMotelPreset("express"); });
    $("motelPresetNight").addEventListener("click", function () { applyMotelPreset("night"); });
    $("nightApplyAll").addEventListener("click", function () { applyAll("night"); });
    $("dayUseApplyAll").addEventListener("click", function () { applyAll("dayuse"); });
    $("reloadAll").addEventListener("click", loadAll);
    $("nightSimForm").addEventListener("submit", simulateNight);
    $("dayUseSimForm").addEventListener("submit", simulateDayUse);
    $("motelSimForm").addEventListener("submit", simulateMotel);
  }

  async function loadAll() {
    if (state.empresaID <= 0) {
      setMessage("nightMsg", "No se pudo iniciar tarifas_de_hotel sin empresa_id.", true);
      $("nightList").innerHTML = '<div class="hotel-empty">Abre esta pagina desde Administrar empresa para recibir el contexto de empresa.</div>';
      $("dayUseList").innerHTML = '<div class="hotel-empty">Sin contexto de empresa.</div>';
      $("motelList").innerHTML = '<div class="hotel-empty">Sin contexto de empresa.</div>';
      return;
    }
    setMessage("rulesMsg", "Cargando configuracion tarifaria...", false);
    try {
      await loadStations();
      await Promise.all([loadNightRates(), loadDayUseRates(), loadMotelRates(), loadRules()]);
      setMessage("rulesMsg", "", false);
    } catch (e) {
      setMessage("rulesMsg", e.message || "No se pudo cargar la configuracion tarifaria.", true);
    }
  }

  function initDefaults() {
    fillDaySelect("dayUseDayFrom", false);
    fillDaySelect("dayUseDayTo", false);
    fillDaySelect("simDayUseDay", false);
    fillDaySelect("motelDayFrom", false);
    fillDaySelect("motelDayTo", false);
    resetNight();
    resetDayUse();
    resetMotel();
    var now = new Date();
    var yesterday = new Date(now.getTime() - 24 * 60 * 60 * 1000);
    $("simNightStart").value = toDateTimeLocal(yesterday);
    $("simNightEnd").value = toDateTimeLocal(now);
  }

  document.addEventListener("DOMContentLoaded", function () {
    state.empresaID = resolveEmpresaID();
    initDefaults();
    bindEvents();
    loadAll();
  });
}());
