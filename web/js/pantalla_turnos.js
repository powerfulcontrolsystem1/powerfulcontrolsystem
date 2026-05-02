(function () {
  "use strict";
  function q(name) { return (new URLSearchParams(window.location.search)).get(name) || ""; }
  function esc(v) { return String(v || "").replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;"); }
  async function j(url) {
    var res = await fetch(url, { credentials: "same-origin" });
    var text = await res.text();
    var data = {};
    try { data = text ? JSON.parse(text) : {}; } catch (e) { data = { raw: text }; }
    if (!res.ok) throw new Error(data.error || data.message || data.raw || ("HTTP " + res.status));
    return data;
  }
  var empresaId = q("empresa_id") || q("id");
  var base = "/api/public/turnos_atencion?empresa_id=" + encodeURIComponent(empresaId) + "&action=display";
  var screenTitle = document.getElementById("screenTitle");
  var screenClock = document.getElementById("screenClock");
  var callingNow = document.getElementById("callingNow");
  var recentCalls = document.getElementById("recentCalls");
  var serviceSummary = document.getElementById("serviceSummary");

  function render() {
    screenClock.textContent = new Date().toLocaleTimeString("es-CO", { hour: "2-digit", minute: "2-digit", second: "2-digit" });
  }

  function renderList(el, items, builder, emptyText) {
    el.innerHTML = items && items.length ? items.map(builder).join("") : '<div class="tv-ticket"><strong>' + esc(emptyText) + '</strong></div>';
  }

  async function refresh() {
    var data = await j(base);
    screenTitle.textContent = data.titulo || "Pantalla de turnos";
    renderList(callingNow, data.tickets_llamando || [], function (x) {
      return '<div class="tv-ticket"><div class="tv-code">' + esc(x.codigo_turno) + '</div><strong>' + esc(x.servicio_nombre) + '</strong><span>Puesto: ' + esc(x.puesto_nombre || "Por asignar") + '</span></div>';
    }, "Sin tickets llamando en este momento.");
    renderList(recentCalls, data.llamados_recientes || [], function (x) {
      return '<div class="tv-ticket"><strong>' + esc(x.codigo_turno) + '</strong><span>' + esc(x.servicio_nombre) + ' · ' + esc(x.puesto_nombre || "Puesto") + ' · ' + esc(x.estado) + '</span></div>';
    }, "Sin llamados recientes.");
    renderList(serviceSummary, data.resumen_servicios || [], function (x) {
      return '<div class="tv-ticket"><strong>' + esc(x.etiqueta) + '</strong><span>Espera: ' + esc(x.en_espera) + ' · Atención: ' + esc(x.en_atencion) + '</span></div>';
    }, "Sin servicios activos.");
  }

  render();
  refresh().catch(function () {});
  setInterval(render, 1000);
  setInterval(function () { refresh().catch(function () {}); }, 5000);
})();
