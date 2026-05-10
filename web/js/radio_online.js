(function () {
  "use strict";

  var container = document.getElementById("radioCatalogPage");
  var enabledToggle = document.getElementById("radioOnlineEnabled");
  var statusEl = document.getElementById("radioOnlineStatus");
  var ENABLED_KEY = "pcs_radio_online_enabled";
  var stations = window.__pcsRadioStations || [];
  if (!container) return;

  function escapeHTML(value) {
    return String(value || "").replace(/[&<>\"']/g, function (c) { return ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "\"": "&quot;", "'": "&#39;" })[c]; });
  }

  function readEnabled() {
    try {
      return window.localStorage.getItem(ENABLED_KEY) !== "0";
    } catch (_) {
      return true;
    }
  }

  function writeEnabled(enabled) {
    try {
      window.localStorage.setItem(ENABLED_KEY, enabled ? "1" : "0");
    } catch (_) {}
  }

  function notifyParent(enabled) {
    try {
      if (window.parent && typeof window.parent.__pcsRadioPlayerSetEnabled === "function") {
        window.parent.__pcsRadioPlayerSetEnabled(enabled);
      }
    } catch (_) {}
  }

  function applyEnabled(enabled) {
    if (enabledToggle) enabledToggle.checked = !!enabled;
    if (statusEl) statusEl.textContent = enabled ? "Musica latina online activa." : "Musica latina online desactivada.";
    container.classList.toggle("is-disabled", !enabled);
    container.querySelectorAll("[data-radio-page-play]").forEach(function (button) {
      button.disabled = !enabled;
      button.textContent = enabled ? "Escuchar" : "Desactivada";
    });
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

  applyEnabled(readEnabled());

  if (enabledToggle) {
    enabledToggle.addEventListener("change", function () {
      var enabled = !!enabledToggle.checked;
      writeEnabled(enabled);
      applyEnabled(enabled);
      notifyParent(enabled);
    });
  }

  container.addEventListener("click", function (ev) {
    var button = ev.target.closest("[data-radio-page-play]");
    if (!button) return;
    if (!readEnabled()) {
      applyEnabled(false);
      return;
    }
    var id = button.getAttribute("data-radio-page-play");
    if (window.parent && typeof window.parent.__pcsRadioPlayerOpenStation === "function") {
      window.parent.__pcsRadioPlayerOpenStation(id);
      return;
    }
  });
})();
