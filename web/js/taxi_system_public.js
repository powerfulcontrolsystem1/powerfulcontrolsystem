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
  var map = L.map("taxiPublicMap").setView([4.711, -74.0721], 13);
  L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", { attribution: "&copy; OpenStreetMap" }).addTo(map);
  var layer = L.layerGroup().addTo(map);
  var state = { cfg: null, customerToken: "", currentRequestId: 0, currentPos: null };
  var tokenKey = "taxi_system:customer_token:" + empresaId;

  function setMsg(id, text, bad) {
    var el = document.getElementById(id);
    if (!el) return;
    el.textContent = text || "";
    el.style.color = bad ? "#ffb4b4" : "";
  }

  function persistToken(v) {
    state.customerToken = v || "";
    try { localStorage.setItem(tokenKey, state.customerToken); } catch (e) {}
    document.getElementById("customerSessionChip").textContent = state.customerToken ? "Cliente registrado activo" : "Modo invitado";
  }

  function drawMap(requestData, routePoints) {
    layer.clearLayers();
    var bounds = [];
    if (state.currentPos) {
      L.circleMarker([state.currentPos.lat, state.currentPos.lng], { radius: 8, color: "#3ddc97", weight: 2 }).addTo(layer).bindPopup("Tu ubicación");
      bounds.push([state.currentPos.lat, state.currentPos.lng]);
    }
    if (requestData && num(requestData.recoger_latitud) && num(requestData.recoger_longitud)) {
      L.marker([num(requestData.recoger_latitud), num(requestData.recoger_longitud)]).addTo(layer).bindPopup("Recogida");
      bounds.push([num(requestData.recoger_latitud), num(requestData.recoger_longitud)]);
    }
    if (requestData && num(requestData.destino_latitud) && num(requestData.destino_longitud)) {
      L.marker([num(requestData.destino_latitud), num(requestData.destino_longitud)]).addTo(layer).bindPopup("Destino");
      bounds.push([num(requestData.destino_latitud), num(requestData.destino_longitud)]);
    }
    if (Array.isArray(routePoints) && routePoints.length) {
      var latlngs = routePoints.map(function (p) { return [num(p.latitud), num(p.longitud)]; });
      L.polyline(latlngs, { color: "#5dade2", weight: 4 }).addTo(layer);
      bounds = bounds.concat(latlngs);
    }
    if (bounds.length) {
      map.fitBounds(bounds, { padding: [18, 18] });
    }
  }

  function renderRequestDetails(item) {
    var box = document.getElementById("requestDetails");
    if (!item) {
      box.innerHTML = '<span class="taxi-chip">Aún no has creado una solicitud.</span>';
      document.getElementById("requestStateChip").textContent = "Sin solicitud activa";
      return;
    }
    document.getElementById("requestStateChip").textContent = "Estado: " + (item.estado || "-");
    box.innerHTML = [
      '<span class="taxi-chip">Servicio #' + item.id + '</span>',
      '<span class="taxi-chip">Estado: ' + esc(item.estado || "-") + '</span>',
      item.conductor_nombre ? '<span class="taxi-chip">Conductor: ' + esc(item.conductor_nombre) + '</span>' : '',
      item.vehiculo_placa ? '<span class="taxi-chip">Placa: ' + esc(item.vehiculo_placa) + '</span>' : '',
      item.tarifa_estimada ? '<span class="taxi-chip">Tarifa estimada: $' + Number(item.tarifa_estimada).toLocaleString("es-CO") + '</span>' : '',
      item.recoger_texto ? '<span class="taxi-chip">Recogida: ' + esc(item.recoger_texto) + '</span>' : '',
      item.destino_texto ? '<span class="taxi-chip">Destino: ' + esc(item.destino_texto) + '</span>' : ''
    ].filter(Boolean).join("");
  }

  async function loadConfig() {
    var cfg = await j(base + "&action=config");
    state.cfg = cfg;
    document.getElementById("portalTitle").textContent = cfg.nombre_sistema || "Taxi system";
    document.getElementById("portalDesc").textContent = (cfg.nombre_portal || "Solicita tu servicio") + ". Puedes pedir un vehículo como invitado o registrarte para guardar tus datos y seguir el viaje con mayor comodidad.";
  }

  async function useCurrentLocation() {
    return new Promise(function (resolve, reject) {
      if (!navigator.geolocation) {
        reject(new Error("Tu navegador no soporta geolocalización"));
        return;
      }
      navigator.geolocation.getCurrentPosition(function (pos) {
        state.currentPos = { lat: pos.coords.latitude, lng: pos.coords.longitude, accuracy: pos.coords.accuracy || 0, speed: pos.coords.speed || 0, heading: pos.coords.heading || 0 };
        drawMap(null, []);
        resolve(state.currentPos);
      }, function (err) {
        reject(new Error(err.message || "No se pudo obtener tu ubicación"));
      }, { enableHighAccuracy: true, maximumAge: 5000, timeout: 12000 });
    });
  }

  async function loadRequest() {
    if (!state.currentRequestId) return;
    try {
      var req = await j(base + "&action=request&request_id=" + encodeURIComponent(state.currentRequestId));
      var route = await j(base + "&action=route&request_id=" + encodeURIComponent(state.currentRequestId));
      renderRequestDetails(req);
      drawMap(req, Array.isArray(route) ? route : []);
      if (state.customerToken && state.currentPos && req.comparte_ubicacion_cliente) {
        await j(base + "&action=share_location", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({
          request_id: state.currentRequestId,
          customer_token: state.customerToken,
          latitud: state.currentPos.lat,
          longitud: state.currentPos.lng,
          precision_metros: state.currentPos.accuracy,
          velocidad_kmh: state.currentPos.speed || 0,
          rumbo_grados: state.currentPos.heading || 0
        }) });
      }
    } catch (e) {
      setMsg("requestMsg", e.message, true);
    }
  }

  document.getElementById("btnUseCurrentLocation").addEventListener("click", async function () {
    try {
      await useCurrentLocation();
      setMsg("requestMsg", "Ubicación cargada correctamente.");
    } catch (e) { setMsg("requestMsg", e.message, true); }
  });

  document.getElementById("customerRegisterForm").addEventListener("submit", async function (ev) {
    ev.preventDefault();
    try {
      await j(base + "&action=register_customer", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({
        nombre: document.getElementById("regNombre").value,
        telefono: document.getElementById("regTelefono").value,
        email: document.getElementById("regEmail").value,
        pin: document.getElementById("regPin").value
      }) });
      setMsg("registerMsg", "Cliente registrado. Ahora puedes iniciar sesión.");
      ev.target.reset();
    } catch (e) { setMsg("registerMsg", e.message, true); }
  });

  document.getElementById("customerLoginForm").addEventListener("submit", async function (ev) {
    ev.preventDefault();
    try {
      var data = await j(base + "&action=login_customer", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({
        telefono: document.getElementById("loginTelefono").value,
        pin: document.getElementById("loginPin").value
      }) });
      persistToken(data.token_sesion || "");
      document.getElementById("reqNombre").value = data.nombre || "";
      document.getElementById("reqTelefono").value = data.telefono || "";
      document.getElementById("reqDocumento").value = data.documento || "";
      setMsg("loginMsg", "Sesión iniciada correctamente.");
    } catch (e) { setMsg("loginMsg", e.message, true); }
  });

  document.getElementById("requestForm").addEventListener("submit", async function (ev) {
    ev.preventDefault();
    try {
      if (document.getElementById("shareLocation").checked && !state.currentPos) {
        await useCurrentLocation();
      }
      var payload = {
        cliente_nombre: document.getElementById("reqNombre").value,
        cliente_telefono: document.getElementById("reqTelefono").value,
        cliente_documento: document.getElementById("reqDocumento").value,
        recoger_texto: document.getElementById("reqPickup").value,
        destino_texto: document.getElementById("reqDestination").value,
        notas: document.getElementById("reqNotes").value,
        metodo_solicitud: document.getElementById("reqMetodo").value || "Taxi",
        comparte_ubicacion_cliente: document.getElementById("shareLocation").checked
      };
      if (state.currentPos) {
        payload.recoger_latitud = state.currentPos.lat;
        payload.recoger_longitud = state.currentPos.lng;
      }
      var req = await j(base + "&action=request_service", { method: "POST", headers: Object.assign({ "Content-Type": "application/json" }, state.customerToken ? { "X-Taxi-Customer-Token": state.customerToken } : {}), body: JSON.stringify(payload) });
      state.currentRequestId = req.id;
      setMsg("requestMsg", "Solicitud creada. Estamos buscando la unidad más cercana.");
      await loadRequest();
    } catch (e) { setMsg("requestMsg", e.message, true); }
  });

  (async function init() {
    try {
      persistToken((function(){ try { return localStorage.getItem(tokenKey) || ""; } catch (e) { return ""; } })());
      await loadConfig();
      drawMap(null, []);
      setInterval(loadRequest, 12000);
    } catch (e) {
      setMsg("requestMsg", e.message, true);
    }
  })();
})();
