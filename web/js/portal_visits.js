(function initPortalVisits() {
  "use strict";

  var widgets = Array.prototype.slice.call(document.querySelectorAll("[data-portal-visits-widget]"));
  var legacyRoot = document.getElementById("portalVisitCounter");
  if (!widgets.length && legacyRoot) widgets = [legacyRoot];
  if (!widgets.length) return;

  var localKey = "pcs_portal_visit_counted_at_v1";
  var countIntervalMs = 30 * 60 * 1000;
  var countedInPage = false;
  var sharedCountryPromise = null;

  var countryMeta = {
    CO: { name: "Colombia", x: 282, y: 285 },
    PA: { name: "Panama", x: 255, y: 270 },
    EC: { name: "Ecuador", x: 265, y: 315 },
    US: { name: "Estados Unidos", x: 205, y: 170 },
    CA: { name: "Canada", x: 190, y: 115 },
    MX: { name: "Mexico", x: 185, y: 230 },
    CR: { name: "Costa Rica", x: 238, y: 270 },
    DO: { name: "Republica Dominicana", x: 312, y: 245 },
    VE: { name: "Venezuela", x: 315, y: 280 },
    PE: { name: "Peru", x: 292, y: 335 },
    CL: { name: "Chile", x: 320, y: 410 },
    AR: { name: "Argentina", x: 350, y: 430 },
    BR: { name: "Brasil", x: 380, y: 350 },
    ES: { name: "Espana", x: 475, y: 205 },
    FR: { name: "Francia", x: 468, y: 185 },
    GB: { name: "Reino Unido", x: 455, y: 160 },
    DE: { name: "Alemania", x: 500, y: 175 },
    IT: { name: "Italia", x: 505, y: 205 },
    PT: { name: "Portugal", x: 455, y: 210 },
    MA: { name: "Marruecos", x: 465, y: 240 },
    ZA: { name: "Sudafrica", x: 560, y: 420 },
    NG: { name: "Nigeria", x: 525, y: 310 },
    EG: { name: "Egipto", x: 555, y: 255 },
    IN: { name: "India", x: 690, y: 285 },
    CN: { name: "China", x: 760, y: 230 },
    JP: { name: "Japon", x: 875, y: 235 },
    AU: { name: "Australia", x: 835, y: 400 }
  };

  function esc(value) {
    return String(value == null ? "" : value)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#39;");
  }

  function normalizeCountry(raw) {
    var value = String(raw || "").trim().toUpperCase();
    return /^[A-Z]{2}$/.test(value) ? value : "";
  }

  function formatNumber(value) {
    try {
      return new Intl.NumberFormat("es-CO").format(Number(value || 0));
    } catch (err) {
      return String(Number(value || 0));
    }
  }

  function shouldCountVisit() {
    try {
      var last = Number(window.localStorage.getItem(localKey) || 0);
      return !last || Date.now() - last > countIntervalMs;
    } catch (err) {
      return true;
    }
  }

  function markVisitCounted() {
    try {
      window.localStorage.setItem(localKey, String(Date.now()));
    } catch (err) {}
  }

  function widgetShouldCount(widget) {
    var raw = widget.getAttribute("data-count-visit");
    if (raw == null || raw === "") return widget.id === "portalVisitCounter";
    return /^(1|true|si|yes)$/i.test(raw);
  }

  function partsFor(widget) {
    return {
      totalEl: widget.querySelector("[data-portal-visits-total]") || document.getElementById("portalVisitTotal"),
      statusEl: widget.querySelector("[data-portal-visits-status]") || document.getElementById("portalVisitStatus"),
      mapEl: widget.querySelector("[data-portal-visits-map]") || document.getElementById("portalVisitMap"),
      listEl: widget.querySelector("[data-portal-visits-countries]") || document.getElementById("portalVisitCountries")
    };
  }

  async function detectCountry() {
    try {
      var res = await fetch("/api/public/geo", { credentials: "same-origin" });
      if (!res.ok) throw new Error("geo " + res.status);
      var data = await res.json();
      return normalizeCountry(data && (data.pais_codigo || data.country_code || data.country));
    } catch (err) {
      return "";
    }
  }

  async function loadStats(country, countVisit) {
    var method = countVisit && !countedInPage && shouldCountVisit() ? "POST" : "GET";
    var options = { method: method, credentials: "same-origin", headers: {} };
    if (method === "POST") {
      countedInPage = true;
      options.headers["Content-Type"] = "application/json";
      options.body = JSON.stringify({ pais_codigo: country || "" });
    }
    var res = await fetch("/api/public/portal_visitas", options);
    var data = await res.json().catch(function() { return null; });
    if (!res.ok || !data || data.ok === false) {
      throw new Error(data && data.error ? data.error : "No se pudo cargar el contador");
    }
    if (method === "POST") markVisitCounted();
    return data;
  }

  function metaFor(code) {
    code = normalizeCountry(code);
    if (countryMeta[code]) return countryMeta[code];
    var seed = 0;
    for (var i = 0; i < code.length; i += 1) seed += code.charCodeAt(i) * (i + 7);
    return {
      name: code || "Pais no detectado",
      x: 430 + (seed % 440),
      y: 130 + (seed % 260)
    };
  }

  function markerColor(index, count, max) {
    var ratio = max > 0 ? count / max : 0;
    if (index === 0 || ratio >= 0.76) return "#ef4444";
    if (ratio >= 0.46) return "#f59e0b";
    if (ratio >= 0.22) return "#14b8a6";
    return "#2563eb";
  }

  function renderMap(mapEl, rows) {
    if (!mapEl) return;
    var max = rows.reduce(function(acc, item) { return Math.max(acc, Number(item.visitas || 0)); }, 0);
    var graticule = [120, 240, 360, 480, 600, 720, 840].map(function(x) {
      return '<path class="portal-visit-graticule" d="M' + x + ' 52 C' + (x - 18) + ' 180 ' + (x + 18) + ' 340 ' + x + ' 468"></path>';
    }).join("") + [120, 200, 280, 360, 440].map(function(y) {
      return '<path class="portal-visit-graticule" d="M54 ' + y + ' C250 ' + (y - 18) + ' 750 ' + (y - 18) + ' 946 ' + y + '"></path>';
    }).join("");
    var mapLayer = '<image class="portal-visit-real-map" href="/img/world-map-natural-earth-public-domain.svg" x="42" y="44" width="916" height="432" preserveAspectRatio="xMidYMid meet"></image>';
    var markers = rows.map(function(item, index) {
      var code = normalizeCountry(item.pais_codigo);
      var meta = metaFor(code);
      var count = Number(item.visitas || 0);
      var radius = Math.max(7, Math.min(24, 7 + Math.sqrt(count || 1) * 3));
      var color = markerColor(index, count, max);
      return '<g class="portal-visit-marker" tabindex="0" role="listitem" style="color:' + color + '" aria-label="' + esc(meta.name) + ': ' + esc(formatNumber(count)) + ' visitas">' +
        '<circle class="portal-visit-marker-halo" cx="' + meta.x + '" cy="' + meta.y + '" r="' + (radius + 8) + '"></circle>' +
        '<circle cx="' + meta.x + '" cy="' + meta.y + '" r="' + radius + '" fill="' + color + '"></circle>' +
        '<text x="' + meta.x + '" y="' + (meta.y + 4) + '">' + esc(code) + '</text>' +
        '<title>' + esc(meta.name) + ': ' + esc(formatNumber(count)) + ' visitas</title>' +
        '</g>';
    }).join("");

    mapEl.innerHTML =
      '<svg class="portal-visit-world" viewBox="0 0 1000 520" role="img" aria-label="Mapa de visitas por pais">' +
      '<defs><radialGradient id="portalVisitOcean" cx="50%" cy="35%" r="70%"><stop offset="0%" stop-color="rgba(20,184,166,.20)"></stop><stop offset="100%" stop-color="rgba(37,99,235,.08)"></stop></radialGradient></defs>' +
      '<rect class="portal-visit-map-frame" x="20" y="26" width="960" height="468" rx="22" fill="url(#portalVisitOcean)"></rect>' +
      '<g aria-hidden="true">' + graticule + mapLayer + '</g>' +
      '<g role="list">' + markers + '</g>' +
      '</svg>';
  }

  function renderList(listEl, rows) {
    if (!listEl) return;
    if (!rows.length) {
      listEl.innerHTML = '<div class="portal-visit-empty">Aun no hay visitas registradas.</div>';
      return;
    }
    var max = rows.reduce(function(acc, item) { return Math.max(acc, Number(item.visitas || 0)); }, 0);
    listEl.innerHTML = rows.slice(0, 10).map(function(item, index) {
      var code = normalizeCountry(item.pais_codigo);
      var meta = metaFor(code);
      var count = Number(item.visitas || 0);
      var pct = max > 0 ? Math.max(8, Math.round((count / max) * 100)) : 8;
      var color = markerColor(index, count, max);
      return '<article class="portal-visit-country">' +
        '<div><strong>' + esc(meta.name) + '</strong><span>' + esc(code) + '</span></div>' +
        '<div class="portal-visit-bar" aria-hidden="true"><i style="width:' + pct + '%;background:' + color + '"></i></div>' +
        '<b>' + esc(formatNumber(count)) + '</b>' +
        '</article>';
    }).join("");
  }

  function render(widget, data) {
    var parts = partsFor(widget);
    var rows = Array.isArray(data && data.paises) ? data.paises : [];
    if (parts.totalEl) parts.totalEl.textContent = formatNumber(data && data.total_visitas);
    renderMap(parts.mapEl, rows);
    renderList(parts.listEl, rows);
    if (parts.statusEl) {
      var country = normalizeCountry(data && data.pais_registrado);
      parts.statusEl.textContent = country ? "Tu visita cuenta para " + metaFor(country).name + "." : "Conteo agregado por pais, sin guardar IP.";
    }
  }

  async function runWidget(widget) {
    var parts = partsFor(widget);
    try {
      if (parts.statusEl) parts.statusEl.textContent = "Cargando contador...";
      if (!sharedCountryPromise) sharedCountryPromise = detectCountry();
      var country = await sharedCountryPromise;
      var data = await loadStats(country, widgetShouldCount(widget));
      render(widget, data);
    } catch (err) {
      if (parts.statusEl) parts.statusEl.textContent = "Contador temporalmente no disponible.";
      render(widget, { total_visitas: 0, paises: [] });
    }
  }

  widgets.forEach(runWidget);
})();
