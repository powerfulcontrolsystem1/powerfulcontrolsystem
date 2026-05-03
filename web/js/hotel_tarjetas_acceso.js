(function () {
  "use strict";

  var state = { empresaID: 0, stations: [], cards: [] };
  function $(id) { return document.getElementById(id); }
  function text(v) { return String(v == null ? "" : v).trim(); }
  function number(v) { var n = Number(v); return Number.isFinite(n) ? n : 0; }
  function esc(v) { return String(v == null ? "" : v).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/\"/g, "&quot;").replace(/'/g, "&#39;"); }

  function getQueryParam(name) {
    try { return new URLSearchParams(location.search || "").get(name); } catch (e) { return ""; }
  }

  function resolveEmpresaID() {
    var id = number(getQueryParam("empresa_id") || getQueryParam("id"));
    if (id <= 0 && parent && parent !== window) {
      try {
        if (typeof parent.__resolveEmpresaIdContext === "function") id = number(parent.__resolveEmpresaIdContext());
      } catch (e) {}
    }
    if (id <= 0) {
      try { id = number(sessionStorage.getItem("active_empresa_id") || localStorage.getItem("active_empresa_id") || sessionStorage.getItem("empresa_id")); } catch (e) {}
    }
    if (id > 0) {
      try { sessionStorage.setItem("active_empresa_id", String(id)); sessionStorage.setItem("empresa_id", String(id)); } catch (e) {}
    }
    return id;
  }

  function defaultStationCode(id) { return state.empresaID > 0 && id > 0 ? "EST-" + state.empresaID + "-" + id : ""; }
  function parseStationID(code) {
    var m = /^EST-(\d+)-(\d+)$/i.exec(text(code));
    if (!m) return 0;
    if (state.empresaID > 0 && number(m[1]) !== state.empresaID) return 0;
    return number(m[2]);
  }

  function setMsg(message, isError) {
    $("msg").textContent = message || "";
    $("msg").classList.toggle("is-error", !!isError);
  }

  async function ensureOk(resp) {
    if (resp.ok) return;
    var raw = await resp.text().catch(function () { return ""; });
    throw new Error(raw || "HTTP " + resp.status);
  }

  async function fetchJSON(url, options) {
    var resp = await fetch(url, Object.assign({ credentials: "same-origin" }, options || {}));
    await ensureOk(resp);
    return resp.status === 204 ? null : resp.json();
  }

  function endpoint(path, extra) {
    var params = new URLSearchParams(extra || {});
    params.set("empresa_id", String(state.empresaID));
    return path + "?" + params.toString();
  }

  async function loadStations() {
    var byID = new Map();
    try {
      var raw = localStorage.getItem("estaciones_config_empresa_" + state.empresaID);
      var parsed = raw ? JSON.parse(raw) : null;
      (parsed && Array.isArray(parsed.estaciones) ? parsed.estaciones : []).forEach(function (item, idx) {
        var id = number(item.id || idx + 1);
        if (id > 0) byID.set(id, { id: id, code: defaultStationCode(id), name: text(item.nombre) || "Habitación " + id });
      });
    } catch (e) {}
    try {
      var carts = await fetchJSON(endpoint("/api/empresa/carritos_compra", { include_inactive: "1" }));
      (Array.isArray(carts) ? carts : []).forEach(function (item) {
        var id = parseStationID(item.codigo);
        if (id > 0) byID.set(id, { id: id, code: text(item.codigo).toUpperCase(), name: text(item.nombre) || "Habitación " + id });
      });
    } catch (e) {}
    state.stations = Array.from(byID.values()).sort(function (a, b) { return a.id - b.id; });
    $("stationSelect").innerHTML = '<option value="">Selecciona habitación</option>' + state.stations.map(function (s) {
      return '<option value="' + esc(s.id) + '">' + esc(s.name + " (#" + s.id + ")") + "</option>";
    }).join("");
  }

  function syncStation(id) {
    var sid = number(id);
    if (sid <= 0) return;
    var station = state.stations.find(function (s) { return s.id === sid; });
    $("stationId").value = String(sid);
    $("stationCode").value = station && station.code ? station.code : defaultStationCode(sid);
    $("stationName").value = station && station.name ? station.name : "Habitación " + sid;
  }

  function toLocalInput(date) {
    var pad = function (n) { return String(n).padStart(2, "0"); };
    return date.getFullYear() + "-" + pad(date.getMonth() + 1) + "-" + pad(date.getDate()) + "T" + pad(date.getHours()) + ":" + pad(date.getMinutes());
  }

  function resetForm() {
    $("cardId").value = "";
    $("stationSelect").value = "";
    $("stationId").value = "";
    $("stationCode").value = "";
    $("stationName").value = "";
    $("cardCode").value = "CARD-" + Date.now().toString().slice(-6);
    $("cardUid").value = "";
    $("guestName").value = "";
    $("reservationId").value = "";
    $("maxUses").value = "0";
    var now = new Date();
    var checkout = new Date(now.getTime() + 22 * 60 * 60 * 1000);
    $("validFrom").value = toLocalInput(now);
    $("validTo").value = toLocalInput(checkout);
    $("status").value = "activo";
    $("notes").value = "";
    setMsg("", false);
  }

  function payload() {
    var sid = number($("stationId").value);
    return {
      id: number($("cardId").value),
      empresa_id: state.empresaID,
      estacion_id: sid,
      estacion_codigo: text($("stationCode").value) || defaultStationCode(sid),
      estacion_nombre: text($("stationName").value) || "Habitación " + sid,
      codigo_tarjeta: text($("cardCode").value),
      card_uid: text($("cardUid").value),
      huesped_nombre: text($("guestName").value),
      reserva_id: number($("reservationId").value),
      vigente_desde: text($("validFrom").value),
      vigente_hasta: text($("validTo").value),
      max_usos: number($("maxUses").value),
      estado: text($("status").value) || "activo",
      observaciones: text($("notes").value)
    };
  }

  async function saveCard(event) {
    event.preventDefault();
    var data = payload();
    if (data.estacion_id <= 0 || !data.codigo_tarjeta || !data.vigente_desde || !data.vigente_hasta) {
      setMsg("Completa habitación, código de tarjeta y vigencia.", true);
      return;
    }
    setMsg("Programando tarjeta...", false);
    try {
      var saved = await fetchJSON("/api/empresa/hotel_tarjetas_acceso", {
        method: data.id > 0 ? "PUT" : "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data)
      });
      var code = saved && saved.access_code ? " Código temporal: " + saved.access_code : "";
      setMsg("Tarjeta guardada." + code, false);
      await loadCards();
    } catch (e) {
      setMsg(e.message || "No se pudo guardar la tarjeta.", true);
    }
  }

  function renderCards() {
    if (!state.cards.length) {
      $("list").innerHTML = '<div class="form-help" style="padding:16px">No hay tarjetas programadas.</div>';
      return;
    }
    $("list").innerHTML = '<table class="access-table"><thead><tr><th>Tarjeta</th><th>Habitación</th><th>Huésped</th><th>Vigencia</th><th>Uso</th><th>Estado</th><th>Acciones</th></tr></thead><tbody>' +
      state.cards.map(function (item) {
        var active = text(item.estado || "activo").toLowerCase() === "activo";
        return "<tr><td><strong>" + esc(item.codigo_tarjeta) + "</strong><br><small>ID " + esc(item.id) + "</small></td>" +
          "<td>" + esc(item.estacion_nombre || ("Habitación " + item.estacion_id)) + "<br><small>" + esc(item.estacion_codigo || defaultStationCode(item.estacion_id)) + "</small></td>" +
          "<td>" + esc(item.huesped_nombre || "-") + "<br><small>Reserva " + esc(item.reserva_id || "-") + "</small></td>" +
          "<td>" + esc(item.vigente_desde) + "<br>" + esc(item.vigente_hasta) + "</td>" +
          "<td>" + esc(item.usos_realizados || 0) + (item.max_usos > 0 ? " / " + esc(item.max_usos) : " / sin límite") + "<br><small>" + esc(item.ultimo_uso_en || "Sin uso") + "</small></td>" +
          "<td><span class=\"access-badge " + (active ? "" : "off") + "\">" + esc(item.estado || "activo") + "</span></td>" +
          "<td><button class=\"btn secondary\" data-edit=\"" + esc(item.id) + "\" type=\"button\">Editar</button> " +
          "<button class=\"btn secondary\" data-toggle=\"" + esc(item.id) + "\" data-action=\"" + (active ? "desactivar" : "activar") + "\" type=\"button\">" + (active ? "Bloquear" : "Activar") + "</button></td></tr>";
      }).join("") + "</tbody></table>";
    $("list").querySelectorAll("[data-edit]").forEach(function (btn) {
      btn.addEventListener("click", function () { editCard(number(btn.getAttribute("data-edit"))); });
    });
    $("list").querySelectorAll("[data-toggle]").forEach(function (btn) {
      btn.addEventListener("click", function () { toggleCard(number(btn.getAttribute("data-toggle")), btn.getAttribute("data-action")); });
    });
  }

  async function loadCards() {
    var rows = await fetchJSON(endpoint("/api/empresa/hotel_tarjetas_acceso", { include_inactive: $("includeInactive").checked ? "1" : "0", limit: "800" }));
    state.cards = Array.isArray(rows) ? rows : [];
    renderCards();
  }

  function editCard(id) {
    var item = state.cards.find(function (row) { return number(row.id) === id; });
    if (!item) return;
    $("cardId").value = item.id || "";
    $("stationSelect").value = item.estacion_id || "";
    $("stationId").value = item.estacion_id || "";
    $("stationCode").value = item.estacion_codigo || defaultStationCode(item.estacion_id);
    $("stationName").value = item.estacion_nombre || "";
    $("cardCode").value = item.codigo_tarjeta || "";
    $("cardUid").value = "";
    $("guestName").value = item.huesped_nombre || "";
    $("reservationId").value = item.reserva_id || "";
    $("maxUses").value = item.max_usos || 0;
    $("validFrom").value = String(item.vigente_desde || "").replace(" ", "T").slice(0, 16);
    $("validTo").value = String(item.vigente_hasta || "").replace(" ", "T").slice(0, 16);
    $("status").value = item.estado || "activo";
    $("notes").value = item.observaciones || "";
    setMsg("Editando tarjeta #" + id + ". El UID físico solo se cambia si escribes uno nuevo.", false);
  }

  async function toggleCard(id, action) {
    try {
      await fetchJSON(endpoint("/api/empresa/hotel_tarjetas_acceso", { id: String(id), action: action }), { method: "PUT" });
      await loadCards();
    } catch (e) {
      setMsg(e.message || "No se pudo cambiar el estado.", true);
    }
  }

  document.addEventListener("DOMContentLoaded", async function () {
    state.empresaID = resolveEmpresaID();
    resetForm();
    $("cardForm").addEventListener("submit", saveCard);
    $("resetBtn").addEventListener("click", resetForm);
    $("reloadBtn").addEventListener("click", loadCards);
    $("includeInactive").addEventListener("change", loadCards);
    $("stationSelect").addEventListener("change", function () { syncStation(this.value); });
    $("stationId").addEventListener("input", function () { if (number(this.value) > 0 && !text($("stationCode").value)) syncStation(this.value); });
    if (state.empresaID <= 0) {
      setMsg("Abre esta página desde Administrar empresa para recibir empresa_id.", true);
      return;
    }
    await loadStations();
    await loadCards();
  });
})();
