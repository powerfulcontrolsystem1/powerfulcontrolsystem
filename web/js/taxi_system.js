(function () {
  "use strict";

  var tabMeta = {
    configuracion: {
      title: "Configuración",
      summary: "Define identidad comercial, radio de búsqueda y reglas del portal de clientes y conductores."
    },
    conductores: {
      title: "Conductores",
      summary: "Registra la flota, administra acceso móvil y controla disponibilidad y último GPS."
    },
    gps: {
      title: "GPS y telemetria",
      summary: "Conecta apps moviles, trackers, OBD2, celulares y proveedores externos a la operacion taxi."
    },
    despacho: {
      title: "Despacho",
      summary: "Opera solicitudes, relanza ofertas y decide cancelaciones desde la central."
    },
    seguimiento: {
      title: "Seguimiento",
      summary: "Consulta rutas, ofertas recientes y trazabilidad GPS de la operación en vivo."
    }
  };

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
  var tileLayers = {
    osm: L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", { attribution: "&copy; OpenStreetMap" }),
    carto: L.tileLayer("https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png", { attribution: "&copy; OpenStreetMap &copy; CARTO" }),
    hot: L.tileLayer("https://{s}.tile.openstreetmap.fr/hot/{z}/{x}/{y}.png", { attribution: "&copy; OpenStreetMap, HOT" })
  };
  var activeTileLayer = tileLayers.osm.addTo(map);
  var layer = L.layerGroup().addTo(map);
  var state = { cfg: null, requests: [], drivers: [], gpsDevices: [], mapFilter: "all" };

  function setMsg(id, text, bad) {
    var el = document.getElementById(id);
    if (!el) return;
    el.textContent = text || "";
    el.style.color = bad ? "#b91c1c" : "";
  }

  function setPageMsg(text, bad) {
    var el = document.getElementById("taxiPageMsg");
    if (!el) return;
    el.textContent = text || "";
    el.style.color = bad ? "#b91c1c" : "";
    el.style.background = bad ? "rgba(254,242,242,.92)" : "";
    el.style.borderColor = bad ? "rgba(185,28,28,.22)" : "";
  }

  function setTab(tab) {
    var meta = tabMeta[tab] || tabMeta.configuracion;
    tab = tabMeta[tab] ? tab : "configuracion";
    document.querySelectorAll("[data-taxi-tab]").forEach(function (button) {
      var active = button.getAttribute("data-taxi-tab") === tab;
      button.classList.toggle("is-active", active);
      button.classList.toggle("secondary", !active);
    });
    document.querySelectorAll(".taxi-panel").forEach(function (panel) {
      panel.hidden = panel.id !== ("taxiPanel-" + tab);
    });
    var titleEl = document.getElementById("taxiSectionTitle");
    var summaryEl = document.getElementById("taxiSectionSummary");
    if (titleEl) titleEl.textContent = meta.title;
    if (summaryEl) summaryEl.textContent = meta.summary;
    if (tab === "seguimiento") {
      window.setTimeout(function () {
        map.invalidateSize();
        drawMap();
      }, 80);
    }
  }

  function initialTabFromURL() {
    var requested = q("tab") || q("view") || q("section");
    return tabMeta[requested] ? requested : "configuracion";
  }

  function fillKpis(d) {
    document.getElementById("kpiPendientes").textContent = d.solicitudes_pendientes || 0;
    document.getElementById("kpiActivos").textContent = d.servicios_activos || 0;
    document.getElementById("kpiOnline").textContent = d.conductores_online || 0;
    document.getElementById("kpiDisponibles").textContent = d.conductores_disponibles || 0;
    document.getElementById("kpiClientes").textContent = d.clientes_registrados || 0;
  }

  function gpsOnlineCount() {
    return state.gpsDevices.filter(function (x) { return !!(x.ultima_latitud || x.ultima_longitud || x.ultimo_reporte_en); }).length;
  }

  function fillGpsKpi() {
    var el = document.getElementById("kpiGps");
    if (el) el.textContent = String(gpsOnlineCount());
  }

  function renderOperatorStrip() {
    var box = document.getElementById("operatorStrip");
    if (!box) return;
    var active = state.requests.filter(function (x) { return ["aceptada", "en_camino", "abordo"].indexOf(String(x.estado || "")) >= 0; }).length;
    var pending = state.requests.filter(function (x) { return ["pendiente", "ofertado"].indexOf(String(x.estado || "")) >= 0; }).length;
    var available = state.drivers.filter(function (x) { return x.online && x.disponible; }).length;
    box.innerHTML = [
      '<div class="taxi-item"><strong>' + pending + '</strong><span class="form-help">Solicitudes en cola</span></div>',
      '<div class="taxi-item"><strong>' + active + '</strong><span class="form-help">Viajes activos</span></div>',
      '<div class="taxi-item"><strong>' + available + '</strong><span class="form-help">Unidades libres</span></div>',
      '<div class="taxi-item"><strong>' + gpsOnlineCount() + '</strong><span class="form-help">GPS reportando</span></div>'
    ].join("");
  }

  function renderRequests(items) {
    var box = document.getElementById("requestsList");
    if (!items.length) {
      box.innerHTML = '<div class="taxi-item">No hay servicios en cola.</div>';
      return;
    }
    box.innerHTML = items.map(function (x) {
      return '<article class="taxi-item">' +
        '<strong>#' + x.id + ' | ' + esc(x.cliente_nombre) + '</strong>' +
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
        '<strong>' + esc(x.nombre) + ' | ' + esc(x.vehiculo_placa || x.codigo || "") + '</strong>' +
        '<div>' + esc(x.vehiculo_modelo || x.vehiculo_tipo || "Sin vehículo definido") + '</div>' +
        '<div class="taxi-badges">' +
        '<span class="taxi-badge">' + (x.online ? "Online" : "Offline") + '</span>' +
        '<span class="taxi-badge">' + (x.disponible ? "Disponible" : "Ocupado") + '</span>' +
        (x.gps_tipo ? '<span class="taxi-badge">GPS: ' + esc(x.gps_tipo) + '</span>' : '') +
        (x.gps_codigo ? '<span class="taxi-badge">' + esc(x.gps_codigo) + '</span>' : '') +
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
      return '<article class="taxi-item"><strong>#' + x.request_id + ' | ' + esc(x.conductor_nombre || "-") + '</strong><div class="taxi-badges"><span class="taxi-badge">Estado: ' + esc(x.estado) + '</span><span class="taxi-badge">Distancia: ' + esc(x.distancia_km) + ' km</span><span class="taxi-badge">ETA: ' + esc(x.tiempo_aproximado_min) + ' min</span></div></article>';
    }).join("");
  }

  function renderGpsDevices(items) {
    var box = document.getElementById("gpsDevicesList");
    if (!box) return;
    if (!items.length) {
      box.innerHTML = '<div class="taxi-item">No hay dispositivos GPS registrados para taxi.</div>';
      return;
    }
    box.innerHTML = items.map(function (x) {
      var online = !!(x.ultima_latitud || x.ultima_longitud || x.ultimo_reporte_en);
      return '<article class="taxi-item taxi-device-card">' +
        '<div><strong><span class="taxi-device-dot ' + (online ? "is-online" : "") + '"></span> ' + esc(x.nombre || x.codigo || "GPS") + '</strong>' +
        '<div>' + esc([x.proveedor, x.marca, x.modelo].filter(Boolean).join(" / ") || "Sin proveedor definido") + '</div>' +
        '<div class="taxi-badges">' +
        '<span class="taxi-badge">' + esc(x.tipo_dispositivo || "gps_tracker") + '</span>' +
        '<span class="taxi-badge">' + esc(x.protocolo || "manual") + '</span>' +
        (x.placa_activo ? '<span class="taxi-badge">Placa: ' + esc(x.placa_activo) + '</span>' : '') +
        (x.ultimo_reporte_en ? '<span class="taxi-badge">Ultimo: ' + esc(x.ultimo_reporte_en) + '</span>' : '') +
        '</div></div>' +
        '<span class="taxi-badge">' + (online ? "Reportando" : "Configurado") + '</span>' +
        '</article>';
    }).join("");
  }

  function fillDriverGpsSelect() {
    var select = document.getElementById("drvGpsDevice");
    if (!select) return;
    select.innerHTML = '<option value="">Sin equipo dedicado / usar app movil</option>' + state.gpsDevices.map(function (x) {
      return '<option value="' + esc(x.id) + '" data-code="' + esc(x.codigo || "") + '" data-type="' + esc(x.tipo_dispositivo || "") + '" data-provider="' + esc(x.proveedor || "") + '" data-protocol="' + esc(x.protocolo || "") + '">' + esc((x.codigo ? x.codigo + " - " : "") + (x.nombre || "GPS")) + '</option>';
    }).join("");
  }

  async function loadGpsDevices() {
    try {
      var rows = await j(base + "&action=gps_devices&include_inactive=1");
      state.gpsDevices = Array.isArray(rows) ? rows : [];
      renderGpsDevices(state.gpsDevices);
      fillDriverGpsSelect();
      fillGpsKpi();
      renderOperatorStrip();
      drawMap();
    } catch (e) {
      setMsg("gpsMsg", e.message, true);
    }
  }

  function drawMap() {
    layer.clearLayers();
    var bounds = [];
    var filter = state.mapFilter || "all";
    if (state.cfg && num(state.cfg.latitud_base) && num(state.cfg.longitud_base) && filter === "all") {
      var baseMarker = L.circleMarker([num(state.cfg.latitud_base), num(state.cfg.longitud_base)], { radius: 10, color: "#111827", weight: 2, fillColor: "#ffffff", fillOpacity: 1 }).addTo(layer);
      baseMarker.bindPopup("<strong>Base de operacion</strong><br>Central taxi");
      bounds.push([num(state.cfg.latitud_base), num(state.cfg.longitud_base)]);
    }
    state.drivers.forEach(function (d) {
      if (!num(d.ultima_latitud) && !num(d.ultima_longitud)) return;
      if (filter === "requests" || filter === "gps") return;
      if (filter === "available" && !(d.online && d.disponible)) return;
      if (filter === "busy" && d.disponible) return;
      var color = d.disponible ? "#16a34a" : "#f59e0b";
      var marker = L.circleMarker([num(d.ultima_latitud), num(d.ultima_longitud)], { radius: 9, color: color, weight: 3, fillColor: color, fillOpacity: 0.82 }).addTo(layer);
      marker.bindPopup("<strong>" + esc(d.nombre) + "</strong><br>" + esc(d.vehiculo_placa || d.vehiculo_modelo || "") + "<br>GPS: " + esc(d.gps_tipo || "app_movil") + (d.ultimo_reporte_en ? "<br>Ultimo reporte: " + esc(d.ultimo_reporte_en) : ""));
      bounds.push([num(d.ultima_latitud), num(d.ultima_longitud)]);
    });
    state.requests.forEach(function (r) {
      var lat = num(r.recoger_latitud), lng = num(r.recoger_longitud);
      if (!lat && !lng) return;
      if (filter === "available" || filter === "busy" || filter === "gps") return;
      var marker = L.circleMarker([lat, lng], { radius: 8, color: "#dc2626", weight: 3, fillColor: "#fee2e2", fillOpacity: 0.9 }).addTo(layer);
      marker.bindPopup("<strong>Servicio #" + r.id + "</strong><br>" + esc(r.cliente_nombre) + "<br>" + esc(r.recoger_texto || ""));
      bounds.push([lat, lng]);
      if (num(r.destino_latitud) || num(r.destino_longitud)) {
        var dest = [num(r.destino_latitud), num(r.destino_longitud)];
        var destMarker = L.circleMarker(dest, { radius: 7, color: "#2563eb", weight: 3, fillColor: "#dbeafe", fillOpacity: 0.9 }).addTo(layer);
        destMarker.bindPopup("<strong>Destino #" + r.id + "</strong><br>" + esc(r.destino_texto || ""));
        L.polyline([[lat, lng], dest], { color: "#2563eb", weight: 3, dashArray: "8 7" }).addTo(layer);
        bounds.push(dest);
      }
    });
    state.gpsDevices.forEach(function (g) {
      var lat = num(g.ultima_latitud), lng = num(g.ultima_longitud);
      if (!lat && !lng) return;
      if (filter !== "all" && filter !== "gps") return;
      var marker = L.circleMarker([lat, lng], { radius: 7, color: "#7c3aed", weight: 2, fillColor: "#ede9fe", fillOpacity: 0.95 }).addTo(layer);
      marker.bindPopup("<strong>" + esc(g.nombre || g.codigo || "GPS") + "</strong><br>" + esc(g.tipo_dispositivo || "") + " / " + esc(g.protocolo || "") + (g.ultima_velocidad_kmh ? "<br>Velocidad: " + esc(g.ultima_velocidad_kmh) + " km/h" : ""));
      bounds.push([lat, lng]);
    });
    if (bounds.length) {
      map.fitBounds(bounds, { padding: [20, 20] });
    }
    renderOperatorStrip();
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
    renderOperatorStrip();
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
      setPageMsg("Ruta cargada en el mapa para seguimiento del servicio.");
    } catch (e) {
      setMsg("mapMsg", e.message, true);
      setPageMsg(e.message, true);
    }
  }

  document.querySelectorAll("[data-taxi-tab]").forEach(function (button) {
    button.addEventListener("click", function () {
      setTab(button.getAttribute("data-taxi-tab"));
    });
  });

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
      setPageMsg("Configuración actualizada correctamente.");
      await loadConfig();
    } catch (e) {
      setMsg("configMsg", e.message, true);
      setPageMsg(e.message, true);
    }
  });

  document.getElementById("driverForm").addEventListener("submit", async function (ev) {
    ev.preventDefault();
    try {
      var gpsSelect = document.getElementById("drvGpsDevice");
      var gpsOption = gpsSelect && gpsSelect.selectedOptions && gpsSelect.selectedOptions[0] ? gpsSelect.selectedOptions[0] : null;
      await j(base + "&action=drivers", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({
        codigo: document.getElementById("drvCodigo").value,
        nombre: document.getElementById("drvNombre").value,
        documento: document.getElementById("drvDocumento").value,
        telefono: document.getElementById("drvTelefono").value,
        vehiculo_placa: document.getElementById("drvPlaca").value,
        vehiculo_modelo: document.getElementById("drvModelo").value,
        vehiculo_tipo: document.getElementById("drvTipo").value,
        gps_dispositivo_id: gpsSelect && gpsSelect.value ? Number(gpsSelect.value) : 0,
        gps_codigo: gpsOption ? (gpsOption.getAttribute("data-code") || "") : "",
        gps_tipo: document.getElementById("drvGpsTipo").value,
        gps_proveedor: document.getElementById("drvGpsProveedor").value || (gpsOption ? (gpsOption.getAttribute("data-provider") || "") : ""),
        gps_protocolo: document.getElementById("drvGpsProtocolo").value,
        pin: document.getElementById("drvPin").value
      }) });
      setMsg("driverMsg", "Conductor creado correctamente.");
      setPageMsg("Conductor registrado y disponible para la operación.");
      ev.target.reset();
      await loadDashboard();
    } catch (e) {
      setMsg("driverMsg", e.message, true);
      setPageMsg(e.message, true);
    }
  });

  document.getElementById("drvGpsDevice").addEventListener("change", function (ev) {
    var option = ev.target.selectedOptions && ev.target.selectedOptions[0] ? ev.target.selectedOptions[0] : null;
    if (!option || !option.value) return;
    document.getElementById("drvGpsTipo").value = option.getAttribute("data-type") || "gps_tracker";
    document.getElementById("drvGpsProveedor").value = option.getAttribute("data-provider") || "";
    document.getElementById("drvGpsProtocolo").value = option.getAttribute("data-protocol") || "manual";
  });

  document.getElementById("gpsDeviceForm").addEventListener("submit", async function (ev) {
    ev.preventDefault();
    try {
      await j(base + "&action=gps_devices", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({
        codigo: document.getElementById("gpsCodigo").value,
        nombre: document.getElementById("gpsNombre").value,
        tipo_dispositivo: document.getElementById("gpsTipo").value,
        protocolo: document.getElementById("gpsProtocolo").value,
        proveedor: document.getElementById("gpsProveedor").value,
        identificador_hardware: document.getElementById("gpsHardware").value,
        placa_activo: document.getElementById("gpsPlaca").value,
        intervalo_reporte_segundos: num(document.getElementById("gpsIntervalo").value) || 10,
        marca: document.getElementById("gpsMarca").value,
        modelo: document.getElementById("gpsModelo").value,
        estado: "activo"
      }) });
      setMsg("gpsMsg", "Dispositivo GPS agregado correctamente.");
      setPageMsg("GPS listo para vincular a conductores y visualizar en el mapa.");
      ev.target.reset();
      document.getElementById("gpsIntervalo").value = "10";
      await loadGpsDevices();
    } catch (e) {
      setMsg("gpsMsg", e.message, true);
      setPageMsg(e.message, true);
    }
  });

  document.getElementById("mapLayerSelect").addEventListener("change", function (ev) {
    var next = tileLayers[ev.target.value] || tileLayers.osm;
    if (activeTileLayer) map.removeLayer(activeTileLayer);
    activeTileLayer = next.addTo(map);
  });

  document.getElementById("mapFilterSelect").addEventListener("change", function (ev) {
    state.mapFilter = ev.target.value || "all";
    drawMap();
  });

  document.getElementById("centerMapBtn").addEventListener("click", function () {
    drawMap();
    setPageMsg("Mapa centrado con la operacion visible.");
  });

  document.getElementById("useMyBaseBtn").addEventListener("click", function () {
    if (!navigator.geolocation) {
      setMsg("mapMsg", "El navegador no permite geolocalizacion.", true);
      return;
    }
    navigator.geolocation.getCurrentPosition(function (pos) {
      document.getElementById("cfgLatBase").value = pos.coords.latitude.toFixed(8);
      document.getElementById("cfgLngBase").value = pos.coords.longitude.toFixed(8);
      setTab(initialTabFromURL());
      setMsg("configMsg", "Ubicacion cargada. Guarda la configuracion para fijarla como base.");
    }, function () {
      setMsg("mapMsg", "No se pudo obtener la ubicacion del navegador.", true);
    }, { enableHighAccuracy: true, timeout: 10000 });
  });

  document.getElementById("requestsList").addEventListener("click", async function (ev) {
    var dispatchId = ev.target.getAttribute("data-dispatch");
    var routeId = ev.target.getAttribute("data-route");
    var cancelId = ev.target.getAttribute("data-cancel");
    try {
      if (dispatchId) {
        await j(base + "&action=dispatch&request_id=" + encodeURIComponent(dispatchId), { method: "POST" });
        setMsg("mapMsg", "Oferta relanzada para el servicio #" + dispatchId + ".");
        setPageMsg("Oferta relanzada para mejorar la cobertura del servicio.");
        await loadDashboard();
      } else if (routeId) {
        await drawRoute(routeId);
      } else if (cancelId) {
        await j(base + "&action=request_state", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ request_id: Number(cancelId), state: "cancelado", notes: "cancelado desde central" }) });
        setPageMsg("Servicio cancelado desde la central.");
        await loadDashboard();
      }
    } catch (e) {
      setMsg("mapMsg", e.message, true);
      setPageMsg(e.message, true);
    }
  });

  (async function init() {
    try {
      setTab("configuracion");
      setPageMsg("Listo para configurar el despacho, registrar conductores y monitorear la operación.");
      await loadConfig();
      await loadGpsDevices();
      await loadDashboard();
      setInterval(loadDashboard, 15000);
      setInterval(loadGpsDevices, 30000);
    } catch (e) {
      setMsg("mapMsg", e.message, true);
      setPageMsg(e.message, true);
    }
  })();
})();
