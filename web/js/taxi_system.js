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
  var empresaId = q("empresa_id") || (window.__resolveEmpresaIdContext ? window.__resolveEmpresaIdContext() : "");
  var base = "/api/empresa/taxi_system?empresa_id=" + encodeURIComponent(empresaId);
  var map = L.map("taxiMap").setView([4.711, -74.0721], 12);
  L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", { attribution: "&copy; OpenStreetMap" }).addTo(map);
  var layer = L.layerGroup().addTo(map);
  var state = { cfg: null, requests: [], drivers: [] };

  function setMsg(id, text, bad) {
    var el = document.getElementById(id);
    if (!el) return;
    el.textContent = text || "";
    el.style.color = bad ? "#ffb4b4" : "";
  }

  function fillKpis(d) {
    document.getElementById("kpiPendientes").textContent = d.solicitudes_pendientes || 0;
    document.getElementById("kpiActivos").textContent = d.servicios_activos || 0;
    document.getElementById("kpiOnline").textContent = d.conductores_online || 0;
    document.getElementById("kpiDisponibles").textContent = d.conductores_disponibles || 0;
    document.getElementById("kpiClientes").textContent = d.clientes_registrados || 0;
  }

  function renderRequests(items) {
    var box = document.getElementById("requestsList");
    if (!items.length) {
      box.innerHTML = '<div class="taxi-item">No hay servicios en cola.</div>';
      return;
    }
    box.innerHTML = items.map(function (x) {
      return '<article class="taxi-item">' +
        '<strong>#' + x.id + ' · ' + esc(x.cliente_nombre) + '</strong>' +
        '<div>' + esc(x.recoger_texto || "-") + '</div>' +
        '<div class="taxi-badges">' +
        '<span class="taxi-badge">Estado: ' + esc(x.estado || "-") + '</span>' +
        '<span class="taxi-badge">Canal: ' + esc(x.canal || "-") + '</span>' +
        (x.tarifa_estimada ? '<span class="taxi-badge">Estimado: $' + esc(Number(x.tarifa_estimada).toLocaleString("es-CO")) + '</span>' : '') +
        (x.conductor_nombre ? '<span class="taxi-badge">Conductor: ' + esc(x.conductor_nombre) + '</span>' : '') +
        '</div>' +
        '<div class="taxi-actions" style="margin-top:10px;">' +
        '<button class="btn secondary" data-dispatch="' + x.id + '">Redespachar</button>' +
        '<button class="btn secondary" data-route="' + x.id + '">Ver ruta</button>' +
        (x.estado !== "cancelado" && x.estado !== "completado" ? '<button class="btn secondary" data-cancel="' + x.id + '">Cancelar</button>' : '') +
        '</div>' +
        '</article>';
    }).join("");
  }

  function renderDrivers(items) {
    var box = document.getElementById("driversList");
    if (!items.length) {
      box.innerHTML = '<div class="taxi-item">No hay conductores registrados.</div>';
      return;
    }
    box.innerHTML = items.map(function (x) {
      return '<article class="taxi-item">' +
        '<strong>' + esc(x.nombre) + ' · ' + esc(x.vehiculo_placa || x.codigo || "") + '</strong>' +
        '<div>' + esc(x.vehiculo_modelo || x.vehiculo_tipo || "Sin vehículo definido") + '</div>' +
        '<div class="taxi-badges">' +
        '<span class="taxi-badge">' + (x.online ? "Online" : "Offline") + '</span>' +
        '<span class="taxi-badge">' + (x.disponible ? "Disponible" : "Ocupado") + '</span>' +
        (x.ultimo_reporte_en ? '<span class="taxi-badge">GPS: ' + esc(x.ultimo_reporte_en) + '</span>' : '') +
        '</div>' +
        '</article>';
    }).join("");
  }

  function renderOffers(items) {
    var box = document.getElementById("offersList");
    if (!items.length) {
      box.innerHTML = '<div class="taxi-item">No hay ofertas recientes.</div>';
      return;
    }
    box.innerHTML = items.map(function (x) {
      return '<article class="taxi-item"><strong>#' + x.request_id + ' · ' + esc(x.conductor_nombre || "-") + '</strong><div class="taxi-badges"><span class="taxi-badge">Estado: ' + esc(x.estado) + '</span><span class="taxi-badge">Distancia: ' + esc(x.distancia_km) + ' km</span><span class="taxi-badge">ETA: ' + esc(x.tiempo_aproximado_min) + ' min</span></div></article>';
    }).join("");
  }

  function drawMap() {
    layer.clearLayers();
    var bounds = [];
    state.drivers.forEach(function (d) {
      if (!num(d.ultima_latitud) && !num(d.ultima_longitud)) return;
      var marker = L.circleMarker([num(d.ultima_latitud), num(d.ultima_longitud)], { radius: 8, color: d.disponible ? "#27ae60" : "#f39c12", weight: 2, fillOpacity: 0.85 }).addTo(layer);
      marker.bindPopup("<strong>" + esc(d.nombre) + "</strong><br>" + esc(d.vehiculo_placa || d.vehiculo_modelo || ""));
      bounds.push([num(d.ultima_latitud), num(d.ultima_longitud)]);
    });
    state.requests.forEach(function (r) {
      var lat = num(r.recoger_latitud), lng = num(r.recoger_longitud);
      if (!lat && !lng) return;
      var marker = L.marker([lat, lng]).addTo(layer);
      marker.bindPopup("<strong>Servicio #" + r.id + "</strong><br>" + esc(r.cliente_nombre) + "<br>" + esc(r.recoger_texto || ""));
      bounds.push([lat, lng]);
    });
    if (state.cfg && num(state.cfg.latitud_base) && num(state.cfg.longitud_base)) {
      bounds.push([num(state.cfg.latitud_base), num(state.cfg.longitud_base)]);
    }
    if (bounds.length) {
      map.fitBounds(bounds, { padding: [20, 20] });
    }
  }

  async function loadConfig() {
    var cfg = await j(base + "&action=config");
    state.cfg = cfg;
    document.getElementById("cfgNombreSistema").value = cfg.nombre_sistema || "Taxi system";
    document.getElementById("cfgNombrePortal").value = cfg.nombre_portal || "Solicita tu servicio";
    document.getElementById("cfgRadio").value = cfg.radio_busqueda_km || 7;
    document.getElementById("cfgDriversRound").value = cfg.conductores_por_ronda || 5;
    document.getElementById("cfgTimeout").value = cfg.timeout_oferta_segundos || 25;
    document.getElementById("cfgLatBase").value = cfg.latitud_base || "";
    document.getElementById("cfgLngBase").value = cfg.longitud_base || "";
    document.getElementById("cfgRegistroCliente").checked = !!cfg.permitir_registro_cliente;
    document.getElementById("cfgUbicacionCliente").checked = !!cfg.permitir_ubicacion_cliente;
    document.getElementById("cfgAutoDispatch").checked = !!cfg.permitir_despacho_automatico;
    document.getElementById("openTaxiPortalBtn").href = "/taxi_system.html?empresa_id=" + encodeURIComponent(empresaId);
    document.getElementById("openTaxiDriverBtn").href = "/taxi_system_conductor.html?empresa_id=" + encodeURIComponent(empresaId);
  }

  async function loadDashboard() {
    var d = await j(base + "&action=dashboard");
    fillKpis(d);
    state.requests = Array.isArray(d.requests) ? d.requests : [];
    state.drivers = Array.isArray(d.drivers) ? d.drivers : [];
    renderRequests(state.requests);
    renderDrivers(state.drivers);
    renderOffers(Array.isArray(d.offers) ? d.offers : []);
    drawMap();
  }

  async function drawRoute(requestId) {
    try {
      var points = await j(base + "&action=route&request_id=" + encodeURIComponent(requestId));
      drawMap();
      if (!Array.isArray(points) || !points.length) {
        setMsg("mapMsg", "Aún no hay puntos GPS para este servicio.", false);
        return;
      }
      var latlngs = points.map(function (p) { return [num(p.latitud), num(p.longitud)]; });
      L.polyline(latlngs, { color: "#5dade2", weight: 4 }).addTo(layer);
      map.fitBounds(latlngs, { padding: [18, 18] });
      setMsg("mapMsg", "Ruta mostrada para el servicio #" + requestId + ".", false);
    } catch (e) {
      setMsg("mapMsg", e.message, true);
    }
  }

  document.getElementById("taxiConfigForm").addEventListener("submit", async function (ev) {
    ev.preventDefault();
    try {
      await j(base + "&action=config", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({
        nombre_sistema: document.getElementById("cfgNombreSistema").value,
        nombre_portal: document.getElementById("cfgNombrePortal").value,
        radio_busqueda_km: num(document.getElementById("cfgRadio").value),
        conductores_por_ronda: num(document.getElementById("cfgDriversRound").value),
        timeout_oferta_segundos: num(document.getElementById("cfgTimeout").value),
        latitud_base: num(document.getElementById("cfgLatBase").value),
        longitud_base: num(document.getElementById("cfgLngBase").value),
        permitir_registro_cliente: document.getElementById("cfgRegistroCliente").checked,
        permitir_ubicacion_cliente: document.getElementById("cfgUbicacionCliente").checked,
        permitir_despacho_automatico: document.getElementById("cfgAutoDispatch").checked
      }) });
      setMsg("configMsg", "Configuración guardada.");
      await loadConfig();
    } catch (e) { setMsg("configMsg", e.message, true); }
  });

  document.getElementById("driverForm").addEventListener("submit", async function (ev) {
    ev.preventDefault();
    try {
      await j(base + "&action=drivers", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({
        codigo: document.getElementById("drvCodigo").value,
        nombre: document.getElementById("drvNombre").value,
        documento: document.getElementById("drvDocumento").value,
        telefono: document.getElementById("drvTelefono").value,
        vehiculo_placa: document.getElementById("drvPlaca").value,
        vehiculo_modelo: document.getElementById("drvModelo").value,
        vehiculo_tipo: document.getElementById("drvTipo").value,
        pin: document.getElementById("drvPin").value
      }) });
      setMsg("driverMsg", "Conductor creado correctamente.");
      ev.target.reset();
      await loadDashboard();
    } catch (e) { setMsg("driverMsg", e.message, true); }
  });

  document.getElementById("requestsList").addEventListener("click", async function (ev) {
    var dispatchId = ev.target.getAttribute("data-dispatch");
    var routeId = ev.target.getAttribute("data-route");
    var cancelId = ev.target.getAttribute("data-cancel");
    try {
      if (dispatchId) {
        await j(base + "&action=dispatch&request_id=" + encodeURIComponent(dispatchId), { method: "POST" });
        setMsg("mapMsg", "Oferta relanzada para el servicio #" + dispatchId + ".");
        await loadDashboard();
      } else if (routeId) {
        await drawRoute(routeId);
      } else if (cancelId) {
        await j(base + "&action=request_state", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ request_id: Number(cancelId), state: "cancelado", notes: "cancelado desde central" }) });
        await loadDashboard();
      }
    } catch (e) {
      setMsg("mapMsg", e.message, true);
    }
  });

  (async function init() {
    try {
      await loadConfig();
      await loadDashboard();
      setInterval(loadDashboard, 15000);
    } catch (e) {
      setMsg("mapMsg", e.message, true);
    }
  })();
})();
