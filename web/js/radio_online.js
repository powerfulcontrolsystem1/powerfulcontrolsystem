(function () {
  "use strict";

  var container = document.getElementById("radioCatalogPage");
  var stations = window.__pcsRadioStations || [];
  if (!container) return;

  function escapeHTML(value) {
    return String(value || "").replace(/[&<>\"']/g, function (c) { return ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "\"": "&quot;", "'": "&#39;" })[c]; });
  }

  container.innerHTML = stations.map(function (station) {
    return '' +
      '<article class="radio-station-card">' +
      '  <div class="radio-station-badge">' + escapeHTML(station.country) + '</div>' +
      '  <h3>' + escapeHTML(station.name) + '</h3>' +
      '  <p>' + escapeHTML(station.tagline) + '</p>' +
      '  <div class="radio-station-meta"><span>' + escapeHTML(station.genre) + '</span></div>' +
      '  <div class="radio-station-actions">' +
      '    <button type="button" class="btn small" data-radio-page-play="' + escapeHTML(station.id) + '">Escuchar</button>' +
      '    <a href="' + escapeHTML(station.sourceUrl) + '" target="_blank" rel="noopener" class="btn secondary small">Fuente</a>' +
      '  </div>' +
      '</article>';
  }).join("");

  container.addEventListener("click", function (ev) {
    var button = ev.target.closest("[data-radio-page-play]");
    if (!button) return;
    var id = button.getAttribute("data-radio-page-play");
    if (window.parent && typeof window.parent.__pcsRadioPlayerOpenStation === "function") {
      window.parent.__pcsRadioPlayerOpenStation(id);
      return;
    }
  });
})();
