(function () {
  "use strict";

  var container = document.getElementById("radioCatalogPage");
  var enabledToggle = document.getElementById("radioOnlineEnabled");
  var statusEl = document.getElementById("radioOnlineStatus");
  var countrySelect = document.getElementById("radioCountrySelect");
  var countryStatus = document.getElementById("radioCountryStatus");
  var customForm = document.getElementById("radioCustomForm");
  var customName = document.getElementById("radioCustomName");
  var customGenre = document.getElementById("radioCustomGenre");
  var customStream = document.getElementById("radioCustomStream");
  var customSource = document.getElementById("radioCustomSource");
  var customCountry = document.getElementById("radioCustomCountry");
  var ENABLED_KEY = "pcs_radio_online_enabled";
  var PREFS_ENDPOINT = "/api/chat_flotante/preferencias";
  var catalog = window.PCSRadioCatalog;
  if (!container || !catalog) return;

  var state = {
    enabled: false,
    countryCode: "",
    customStations: []
  };

  function escapeHTML(value) {
    return String(value || "").replace(/[&<>\"']/g, function (c) { return ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "\"": "&quot;", "'": "&#39;" })[c]; });
  }

  function parsePositiveInt(raw) {
    var value = Number(String(raw || "").trim());
    if (!Number.isFinite(value)) return 0;
    value = Math.trunc(value);
    return value > 0 ? value : 0;
  }

  function getCurrentEmpresaId() {
    if (typeof window.__resolveEmpresaIdContext === "function") {
      try {
        var resolved = parsePositiveInt(window.__resolveEmpresaIdContext());
        if (resolved > 0) return resolved;
      } catch (_) {}
    }
    var params = new URLSearchParams(window.location.search || "");
    var fromUrl = parsePositiveInt(params.get("empresa_id") || params.get("id") || "");
    if (fromUrl > 0) return fromUrl;
    var keys = ["active_empresa_id", "empresa_id", "admin_empresa_id"];
    var stores = [];
    try { stores.push(window.sessionStorage); } catch (_) {}
    try { stores.push(window.localStorage); } catch (_) {}
    for (var s = 0; s < stores.length; s += 1) {
      var store = stores[s];
      if (!store) continue;
      for (var i = 0; i < keys.length; i += 1) {
        try {
          var parsed = parsePositiveInt(store.getItem(keys[i]) || "");
          if (parsed > 0) return parsed;
        } catch (_) {}
      }
    }
    return 0;
  }

  function prefsEndpoint() {
    var empresaId = getCurrentEmpresaId();
    return PREFS_ENDPOINT + (empresaId ? "?empresa_id=" + encodeURIComponent(String(empresaId)) : "");
  }

  function detectCountryEndpoint() {
    var empresaId = getCurrentEmpresaId();
    var tz = "";
    try { tz = Intl.DateTimeFormat().resolvedOptions().timeZone || ""; } catch (_) {}
    var lang = "";
    try { lang = window.navigator.language || ""; } catch (_) {}
    return "/api/empresa/facturacion_electronica/pais_detectado?empresa_id=" + encodeURIComponent(String(empresaId || 0)) + "&tz=" + encodeURIComponent(tz) + "&lang=" + encodeURIComponent(lang);
  }

  function countryLabel(code) {
    return (catalog.labels && catalog.labels[code]) || code || "país no soportado";
  }

  function writeEnabled(enabled) {
    try {
      window.localStorage.setItem(ENABLED_KEY, enabled ? "1" : "0");
    } catch (_) {}
  }

  function notifyParent() {
    try {
      if (window.parent && typeof window.parent.__pcsRadioPlayerSetEnabled === "function") {
        window.parent.__pcsRadioPlayerSetEnabled(state.enabled);
      }
      if (window.parent && typeof window.parent.__pcsRadioPlayerReloadConfig === "function") {
        window.parent.__pcsRadioPlayerReloadConfig();
      }
    } catch (_) {}
  }

  function setStatus(text) {
    if (statusEl) statusEl.textContent = text;
  }

  function updateCountryControls() {
    if (countrySelect) countrySelect.value = state.countryCode || "";
    if (customCountry && !customCountry.value && state.countryCode) customCountry.value = state.countryCode;
    if (!countryStatus) return;
    if (state.countryCode) {
      countryStatus.textContent = "Pais de emisoras: " + countryLabel(state.countryCode) + ". El sistema muestra las 10 principales y las emisoras personalizadas de esta empresa.";
    } else {
      countryStatus.textContent = "Este módulo muestra emisoras principales solo para Panamá y Ecuador. Puedes escoger un país o agregar emisoras propias.";
    }
  }

  function render() {
    var stations = catalog.stationsForCountry(state.countryCode, state.customStations);
    container.classList.toggle("is-disabled", !state.enabled);
    if (!stations.length) {
      container.innerHTML = '<div class="radio-station-empty">No hay emisoras para mostrar. Selecciona Panamá o Ecuador, o agrega una emisora personalizada para esta empresa.</div>';
    } else {
      container.innerHTML = stations.map(function (station) {
        return '' +
          '<article class="radio-station-card' + (station.custom ? ' is-custom' : '') + '">' +
          '  <div class="radio-station-badge">' + escapeHTML(station.custom ? 'Personalizada' : station.country) + '</div>' +
          '  <h3>' + escapeHTML(station.name) + '</h3>' +
          '  <p>' + escapeHTML(station.tagline) + '</p>' +
          '  <div class="radio-station-meta"><span>' + escapeHTML(station.genre) + '</span></div>' +
          '  <div class="radio-station-actions">' +
          '    <button type="button" class="btn small" data-radio-page-play="' + escapeHTML(station.id) + '"' + (!state.enabled ? ' disabled' : '') + '>' + (state.enabled ? 'Escuchar' : 'Desactivada') + '</button>' +
          (station.sourceUrl ? '    <a href="' + escapeHTML(station.sourceUrl) + '" target="_blank" rel="noopener" class="btn secondary small">Fuente</a>' : '') +
          (station.custom ? '    <button type="button" class="btn danger small" data-radio-page-delete="' + escapeHTML(station.id) + '">Eliminar</button>' : '') +
          '  </div>' +
          '</article>';
      }).join("");
    }
    if (enabledToggle) enabledToggle.checked = !!state.enabled;
    setStatus(state.enabled ? "Emisora online activa." : "Emisora online desactivada.");
    updateCountryControls();
  }

  function applyPayload(data) {
    if (data && typeof data.radio_online_enabled === "boolean") {
      state.enabled = data.radio_online_enabled;
      writeEnabled(state.enabled);
    }
    if (data && Array.isArray(data.radio_custom_stations)) {
      state.customStations = catalog.normalizeCustomList(data.radio_custom_stations);
    }
    if (data && Object.prototype.hasOwnProperty.call(data, "radio_country")) {
      var savedCountry = catalog.normalizeCountry(data.radio_country);
      if (savedCountry) state.countryCode = savedCountry;
    }
    render();
  }

  function persist() {
    return fetch(prefsEndpoint(), {
      method: "PUT",
      credentials: "same-origin",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        radio_online_enabled: !!state.enabled,
        radio_country: state.countryCode || "",
        radio_custom_stations: state.customStations
      })
    }).then(function (res) {
      if (!res.ok) throw new Error("No se pudo guardar la emisora.");
      return res.json();
    }).then(function (data) {
      applyPayload(data || {});
      notifyParent();
      setStatus("Configuracion de emisoras guardada.");
    }).catch(function (err) {
      setStatus("No se pudo guardar. Se mantiene el cambio local hasta reintentar.");
      console.warn("No se pudo guardar la radio online:", err);
    });
  }

  function detectCountry() {
    return fetch(detectCountryEndpoint(), { credentials: "same-origin" })
      .then(function (res) {
        if (!res.ok) throw new Error("país empresa no disponible");
        return res.json();
      })
      .then(function (data) {
        return catalog.normalizeCountry(data && (data.pais_codigo || data.country_code || data.country));
      })
      .catch(function () {
        return fetch("/api/public/geo", { credentials: "same-origin" })
          .then(function (res) {
            if (!res.ok) throw new Error("geo no disponible");
            return res.json();
          })
          .then(function (data) {
            return catalog.normalizeCountry(data && (data.pais_codigo || data.country_code || data.country));
          });
      });
  }

  function load() {
    state.enabled = false;
    render();
    fetch(prefsEndpoint(), { credentials: "same-origin" })
      .then(function (res) {
        if (!res.ok) return {};
        return res.json();
      })
      .then(function (data) {
        applyPayload(data || {});
        if (state.countryCode) return null;
        return detectCountry().then(function (countryCode) {
          if (countryCode) {
            state.countryCode = countryCode;
            render();
          }
          return null;
        });
      })
      .catch(function () {
        render();
      });
  }

  function addCustomStation() {
    var station = catalog.normalizeCustomList([{
      name: customName ? customName.value : "",
      genre: customGenre ? customGenre.value : "",
      streamUrl: customStream ? customStream.value : "",
      sourceUrl: customSource ? customSource.value : "",
      countryCode: customCountry ? customCountry.value : state.countryCode,
      country: countryLabel(customCountry ? customCountry.value : state.countryCode),
      tagline: "Emisora personalizada de esta empresa."
    }])[0];
    if (!station) {
      setStatus("Escribe nombre y URL http/https valida para agregar la emisora.");
      return;
    }
    state.customStations = state.customStations.filter(function (item) { return item.id !== station.id; });
    state.customStations.push(station);
    if (customForm) customForm.reset();
    if (customCountry && state.countryCode) customCountry.value = state.countryCode;
    render();
    persist();
  }

  function deleteCustomStation(id) {
    state.customStations = state.customStations.filter(function (item) { return item.id !== id; });
    render();
    persist();
  }

  if (enabledToggle) {
    enabledToggle.addEventListener("change", function () {
      state.enabled = !!enabledToggle.checked;
      writeEnabled(state.enabled);
      render();
      notifyParent();
      persist();
    });
  }

  if (countrySelect) {
    countrySelect.addEventListener("change", function () {
      var selectedCountry = catalog.normalizeCountry(countrySelect.value);
      if (!selectedCountry) {
        detectCountry().then(function (countryCode) {
          state.countryCode = countryCode || "";
          render();
          persist();
        }).catch(function () {
          state.countryCode = "";
          render();
          persist();
        });
        return;
      }
      state.countryCode = selectedCountry;
      render();
      persist();
    });
  }

  if (customForm) {
    customForm.addEventListener("submit", function (ev) {
      ev.preventDefault();
      addCustomStation();
    });
  }

  container.addEventListener("click", function (ev) {
    var deleteButton = ev.target.closest("[data-radio-page-delete]");
    if (deleteButton) {
      deleteCustomStation(deleteButton.getAttribute("data-radio-page-delete"));
      return;
    }
    var button = ev.target.closest("[data-radio-page-play]");
    if (!button) return;
    if (!state.enabled) {
      render();
      return;
    }
    var id = button.getAttribute("data-radio-page-play");
    if (window.parent && typeof window.parent.__pcsRadioPlayerOpenStation === "function") {
      window.parent.__pcsRadioPlayerOpenStation(id);
    }
  });

  load();
})();
