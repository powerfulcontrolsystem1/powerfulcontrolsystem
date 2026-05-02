(function () {
  "use strict";
  function q(name) { return (new URLSearchParams(window.location.search)).get(name) || ""; }
  function esc(v) { return String(v == null ? "" : v).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;"); }
  function num(v) { var n = Number(v); return Number.isFinite(n) ? n : 0; }
  async function j(url, opts) {
    var res = await fetch(url, Object.assign({ credentials: "same-origin" }, opts || {}));
    var text = await res.text();
    var data = {};
    try { data = text ? JSON.parse(text) : {}; } catch (e) { data = { raw: text }; }
    if (!res.ok) throw new Error(data.error || data.message || data.raw || ("HTTP " + res.status));
    return data;
  }
  var empresaId = q("empresa_id") || q("id");
  var base = "/api/public/taxi_system?empresa_id=" + encodeURIComponent(empresaId);
  var tokenKey = "taxi_system:driver_token:" + empresaId;
  var state = { token: "", currentPos: null, activeRequestId: 0 };
  var map = L.map("driverMap").setView([4.711, -74.0721], 13);
  L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", { attribution: "&copy; OpenStreetMap" }).addTo(map);
  var layer = L.layerGroup().addTo(map);

  function setMsg(id, text, bad) {
    var el = document.getElementById(id);
    if (!el) return;
    el.textContent = text || "";
    el.style.color = bad ? "#ffb4b4" : "";
  }

  function persistToken(v) {
    state.token = v || "";
    try { localStorage.setItem(tokenKey, state.token); } catch (e) {}
  }

  async function useLocation() {
    return new Promise(function (resolve, reject) {
      if (!navigator.geolocation) return reject(new Error("Tu navegador no soporta geolocalización"));
      navigator.geolocation.getCurrentPosition(function (pos) {
        state.currentPos = { lat: pos.coords.latitude, lng: pos.coords.longitude, accuracy: pos.coords.accuracy || 0, speed: pos.coords.speed || 0, heading: pos.coords.heading || 0 };
        layer.clearLayers();
        L.circleMarker([state.currentPos.lat, state.currentPos.lng], { radius: 8, color: "#27ae60", weight: 2 }).addTo(layer).bindPopup("Tu posición");
        map.setView([state.currentPos.lat, state.currentPos.lng], 15);
        resolve(state.currentPos);
      }, function (err) { reject(new Error(err.message || "No se pudo obtener GPS")); }, { enableHighAccuracy: true, maximumAge: 5000, timeout: 12000 });
    });
  }

  async function loadOffers() {
    if (!state.token) return;
    try {
      var rows = await j(base + "&action=driver_offers", { method: "POST", headers: { "Content-Type": "application/json", "X-Taxi-Driver-Token": state.token }, body: JSON.stringify({ token: state.token }) });
      var box = document.getElementById("driverOffers");
      if (!rows.length) {
        box.innerHTML = '<div class="driver-offer">No hay ofertas pendientes en este momento.</div>';
        return;
      }
      box.innerHTML = rows.map(function (x) {
        return '<article class="driver-offer"><strong>Servicio #' + x.request_id + '</strong><div style="margin-top:8px;">Distancia aprox: ' + esc(x.distancia_km) + ' km · ETA: ' + esc(x.tiempo_aproximado_min) + ' min</div><div class="form-actions" style="margin-top:12px;"><button class="btn add" data-accept="' + x.id + '" data-request="' + x.request_id + '">Aceptar</button><button class="btn secondary" data-reject="' + x.id + '">Rechazar</button></div></article>';
      }).join("");
    } catch (e) {
      setMsg("driverPresenceMsg", e.message, true);
    }
  }

  async function sendDriverLocation() {
    if (!state.token) throw new Error("Debes iniciar sesión como conductor");
    if (!state.currentPos) await useLocation();
    await j(base + "&action=driver_location", { method: "POST", headers: { "Content-Type": "application/json", "X-Taxi-Driver-Token": state.token }, body: JSON.stringify({
      token: state.token,
      request_id: state.activeRequestId || 0,
      latitud: state.currentPos.lat,
      longitud: state.currentPos.lng,
      precision_metros: state.currentPos.accuracy,
      velocidad_kmh: state.currentPos.speed || 0,
      rumbo_grados: state.currentPos.heading || 0
    }) });
  }

  document.getElementById("driverLoginForm").addEventListener("submit", async function (ev) {
    ev.preventDefault();
    try {
      var data = await j(base + "&action=login_driver", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({
        documento: document.getElementById("drvDocumentoLogin").value,
        pin: document.getElementById("drvPinLogin").value
      }) });
      persistToken(data.token_sesion || "");
      document.getElementById("driverOnline").checked = true;
      document.getElementById("driverAvailable").checked = true;
      setMsg("driverLoginMsg", "Conductor autenticado: " + (data.nombre || ""));
      setMsg("driverTripBox", "Vehículo: " + (data.vehiculo_placa || data.vehiculo_modelo || "sin definir"));
      await useLocation();
      await sendDriverLocation();
      await loadOffers();
    } catch (e) { setMsg("driverLoginMsg", e.message, true); }
  });

  document.getElementById("btnSendPresence").addEventListener("click", async function () {
    try {
      if (!state.token) throw new Error("Debes iniciar sesión primero");
      await j(base + "&action=driver_presence", { method: "POST", headers: { "Content-Type": "application/json", "X-Taxi-Driver-Token": state.token }, body: JSON.stringify({
        token: state.token,
        online: document.getElementById("driverOnline").checked,
        disponible: document.getElementById("driverAvailable").checked
      }) });
      setMsg("driverPresenceMsg", "Presencia actualizada.");
    } catch (e) { setMsg("driverPresenceMsg", e.message, true); }
  });

  document.getElementById("btnSendDriverLocation").addEventListener("click", async function () {
    try {
      await useLocation();
      await sendDriverLocation();
      setMsg("driverPresenceMsg", "GPS enviado.");
    } catch (e) { setMsg("driverPresenceMsg", e.message, true); }
  });

  document.getElementById("driverOffers").addEventListener("click", async function (ev) {
    var acceptId = ev.target.getAttribute("data-accept");
    var rejectId = ev.target.getAttribute("data-reject");
    var requestId = ev.target.getAttribute("data-request");
    try {
      if (acceptId) {
        var row = await j(base + "&action=respond_offer", { method: "POST", headers: { "Content-Type": "application/json", "X-Taxi-Driver-Token": state.token }, body: JSON.stringify({ token: state.token, offer_id: Number(acceptId), accept: true }) });
        state.activeRequestId = row.id || Number(requestId || 0);
        document.getElementById("tripRequestId").value = state.activeRequestId || "";
        setMsg("driverTripBox", "Servicio aceptado. Cliente: " + (row.cliente_nombre || "") + " · Recogida: " + (row.recoger_texto || ""));
      } else if (rejectId) {
        await j(base + "&action=respond_offer", { method: "POST", headers: { "Content-Type": "application/json", "X-Taxi-Driver-Token": state.token }, body: JSON.stringify({ token: state.token, offer_id: Number(rejectId), accept: false }) });
      }
      await loadOffers();
    } catch (e) {
      setMsg("driverPresenceMsg", e.message, true);
    }
  });

  document.getElementById("btnTripState").addEventListener("click", async function () {
    try {
      var requestId = Number(document.getElementById("tripRequestId").value || state.activeRequestId || 0);
      if (!requestId) throw new Error("Indica el servicio");
      var row = await j(base + "&action=driver_request_state", { method: "POST", headers: { "Content-Type": "application/json", "X-Taxi-Driver-Token": state.token }, body: JSON.stringify({
        token: state.token,
        request_id: requestId,
        state: document.getElementById("tripState").value,
        notes: document.getElementById("tripNotes").value
      }) });
      state.activeRequestId = (row.estado === "completado" || row.estado === "cancelado") ? 0 : row.id;
      setMsg("tripStateMsg", "Estado actualizado a " + (row.estado || "-") + ".");
    } catch (e) { setMsg("tripStateMsg", e.message, true); }
  });

  (async function init() {
    persistToken((function(){ try { return localStorage.getItem(tokenKey) || ""; } catch (e) { return ""; } })());
    setInterval(loadOffers, 12000);
  })();
})();
