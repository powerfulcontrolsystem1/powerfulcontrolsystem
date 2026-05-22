(function initPortalVisits() {
  "use strict";

  var root = document.getElementById("portalVisitCounter");
  if (!root) return;

  var totalEl = document.getElementById("portalVisitTotal");
  var statusEl = document.getElementById("portalVisitStatus");
  var mapEl = document.getElementById("portalVisitMap");
  var listEl = document.getElementById("portalVisitCountries");
  var localKey = "pcs_portal_visit_counted_at_v1";
  var countIntervalMs = 30 * 60 * 1000;

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

  async function loadStats(country) {
    var method = shouldCountVisit() ? "POST" : "GET";
    var url = "/api/public/portal_visitas";
    var options = { method: method, credentials: "same-origin", headers: {} };
    if (method === "POST") {
      options.headers["Content-Type"] = "application/json";
      options.body = JSON.stringify({ pais_codigo: country || "" });
    }
    var res = await fetch(url, options);
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

  function renderMap(rows) {
    if (!mapEl) return;
    var max = rows.reduce(function(acc, item) { return Math.max(acc, Number(item.visitas || 0)); }, 0);
    var markers = rows.map(function(item, index) {
      var code = normalizeCountry(item.pais_codigo);
      var meta = metaFor(code);
      var count = Number(item.visitas || 0);
      var radius = Math.max(7, Math.min(24, 7 + Math.sqrt(count || 1) * 3));
      var color = markerColor(index, count, max);
      return '<g class="portal-visit-marker" tabindex="0" role="listitem" aria-label="' + esc(meta.name) + ': ' + esc(formatNumber(count)) + ' visitas">' +
        '<circle cx="' + meta.x + '" cy="' + meta.y + '" r="' + radius + '" fill="' + color + '"></circle>' +
        '<text x="' + meta.x + '" y="' + (meta.y + 4) + '">' + esc(code) + '</text>' +
        '<title>' + esc(meta.name) + ': ' + esc(formatNumber(count)) + ' visitas</title>' +
        '</g>';
    }).join("");

    mapEl.innerHTML =
      '<svg class="portal-visit-world" viewBox="0 0 1000 520" role="img" aria-label="Mapa de visitas por pais">' +
      '<rect class="portal-visit-ocean" x="0" y="0" width="1000" height="520" rx="18"></rect>' +
      '<path class="portal-visit-land" d="M90 140 C125 82 230 78 285 130 C318 162 282 202 322 232 C362 260 332 315 282 300 C230 285 225 235 175 232 C120 229 70 195 90 140Z"></path>' +
      '<path class="portal-visit-land" d="M260 300 C318 318 378 360 370 430 C362 485 315 505 285 455 C260 412 238 345 260 300Z"></path>' +
      '<path class="portal-visit-land" d="M430 155 C520 105 650 122 710 180 C765 235 700 272 750 325 C792 370 710 420 625 382 C552 350 565 285 505 260 C445 235 380 190 430 155Z"></path>' +
      '<path class="portal-visit-land" d="M505 265 C575 255 630 318 615 395 C602 465 535 470 500 410 C470 358 465 292 505 265Z"></path>' +
      '<path class="portal-visit-land" d="M692 170 C782 112 900 155 930 225 C960 292 865 310 815 278 C760 244 670 235 692 170Z"></path>' +
      '<path class="portal-visit-land" d="M780 370 C842 340 930 365 945 420 C958 468 878 485 825 462 C775 440 742 395 780 370Z"></path>' +
      '<g role="list">' + markers + '</g>' +
      '</svg>';
  }

  function renderList(rows) {
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

  function render(data) {
    var rows = Array.isArray(data && data.paises) ? data.paises : [];
    if (totalEl) totalEl.textContent = formatNumber(data && data.total_visitas);
    renderMap(rows);
    renderList(rows);
    if (statusEl) {
      var country = normalizeCountry(data && data.pais_registrado);
      statusEl.textContent = country ? "Tu visita cuenta para " + metaFor(country).name + "." : "Conteo agregado por pais, sin guardar IP.";
    }
  }

  (async function run() {
    try {
      if (statusEl) statusEl.textContent = "Cargando contador...";
      var country = await detectCountry();
      var data = await loadStats(country);
      render(data);
    } catch (err) {
      if (statusEl) statusEl.textContent = "Contador temporalmente no disponible.";
      render({ total_visitas: 0, paises: [] });
    }
  })();
})();
