(function () {
  "use strict";

  if (!window.PCSRadioCatalog || !document.getElementById("radioMiniAudio")) {
    return;
  }

  var STORAGE_KEY = "pcs_radio_player_state";
  var ENABLED_KEY = "pcs_radio_online_enabled";
  var PREFS_ENDPOINT = "/api/chat_flotante/preferencias";
  var countryTools = window.PCSRadioCatalog;
  var drawer = document.getElementById("radioDrawer");
  var openBtn = document.getElementById("openRadioDrawer");
  var closeBtn = document.getElementById("closeRadioDrawer");
  var closeBtnBottom = document.getElementById("closeRadioDrawerBottom");
  var grid = document.getElementById("radioStationGrid");
  var mini = document.getElementById("radioMiniPlayer");
  var miniAudio = document.getElementById("radioMiniAudio");
  var miniName = document.getElementById("radioMiniName");
  var miniTagline = document.getElementById("radioMiniTagline");
  var miniPlayPause = document.getElementById("radioMiniPlayPause");
  var miniVolume = document.getElementById("radioMiniVolume");
  var miniClose = document.getElementById("radioMiniClose");
  var enabledToggle = document.getElementById("radioFloatingEnabled");
  var enabledStatus = document.getElementById("radioFloatingStatus");
  var countrySelect = document.getElementById("radioCountrySelect");
  var countryStatus = document.getElementById("radioCountryStatus");
  var customForm = document.getElementById("radioCustomForm");
  var customName = document.getElementById("radioCustomName");
  var customGenre = document.getElementById("radioCustomGenre");
  var customStream = document.getElementById("radioCustomStream");
  var customSource = document.getElementById("radioCustomSource");
  var customCountry = document.getElementById("radioCustomCountry");

  var state = {
    stationId: "",
    playing: false,
    volume: 0.7,
    enabled: false,
    countryCode: "",
    countrySource: "",
    customStations: [],
    configLoaded: false
  };

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

  function buildPrefsEndpoint() {
    var empresaId = getCurrentEmpresaId();
    if (!empresaId) return PREFS_ENDPOINT;
    return PREFS_ENDPOINT + "?empresa_id=" + encodeURIComponent(String(empresaId));
  }

  function buildCountryEndpoint() {
    var empresaId = getCurrentEmpresaId();
    var tz = "";
    try { tz = Intl.DateTimeFormat().resolvedOptions().timeZone || ""; } catch (_) {}
    var lang = "";
    try { lang = window.navigator.language || ""; } catch (_) {}
    return "/api/empresa/facturacion_electronica/pais_detectado?empresa_id=" + encodeURIComponent(String(empresaId || 0)) + "&tz=" + encodeURIComponent(tz) + "&lang=" + encodeURIComponent(lang);
  }

  function saveState() {
    try {
      window.localStorage.setItem(STORAGE_KEY, JSON.stringify({
        stationId: state.stationId,
        playing: state.playing,
        volume: state.volume
      }));
    } catch (_) {}
  }

  function loadState() {
    try {
      var raw = window.localStorage.getItem(STORAGE_KEY) || "";
      if (!raw) return;
      var parsed = JSON.parse(raw);
      if (parsed && typeof parsed === "object") {
        state.stationId = String(parsed.stationId || "");
        state.playing = !!parsed.playing;
        state.volume = Number(parsed.volume || 0.7);
      }
    } catch (_) {}
    state.enabled = false;
  }

  function escapeHTML(value) {
    return String(value || "").replace(/[&<>\"']/g, function (c) { return ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "\"": "&quot;", "'": "&#39;" })[c]; });
  }

  function currentStations() {
    return countryTools.stationsForCountry(state.countryCode, state.customStations);
  }

  function stationById(id) {
    return currentStations().find(function (item) { return item.id === id; }) || null;
  }

  function countryLabel(code) {
    return (countryTools.labels && countryTools.labels[code]) || code || "país no soportado";
  }

  function updateCountryControls() {
    if (countrySelect) countrySelect.value = state.countryCode || "";
    if (customCountry && !customCountry.value && state.countryCode) customCountry.value = state.countryCode;
    if (!countryStatus) return;
    if (state.countryCode) {
      countryStatus.textContent = "Pais de emisoras: " + countryLabel(state.countryCode) + ". Se muestran 10 principales y las personalizadas de la empresa.";
    } else {
      countryStatus.textContent = "La detección automática no encontró Panamá o Ecuador. Puedes escoger un país o agregar emisoras personalizadas.";
    }
  }

  function setDrawerOpen(open) {
    if (!drawer || !openBtn) return;
    drawer.classList.toggle("is-open", !!open);
    drawer.setAttribute("aria-hidden", open ? "false" : "true");
    openBtn.setAttribute("aria-expanded", open ? "true" : "false");
  }

  function renderGrid() {
    if (!grid) return;
    var stations = currentStations();
    grid.classList.toggle("is-disabled", !state.enabled);
    if (!stations.length) {
      grid.innerHTML = '<div class="radio-station-empty">Este módulo muestra emisoras principales solo para Panamá y Ecuador. Selecciona uno de esos países o agrega una emisora personalizada para esta empresa.</div>';
      updateCountryControls();
      return;
    }
    grid.innerHTML = stations.map(function (station) {
      var active = state.stationId === station.id;
      return '' +
        '<article class="radio-station-card' + (active ? ' is-active' : '') + (station.custom ? ' is-custom' : '') + '">' +
        '  <div class="radio-station-badge">' + escapeHTML(station.custom ? 'Personalizada' : station.country) + '</div>' +
        '  <h3>' + escapeHTML(station.name) + '</h3>' +
        '  <p>' + escapeHTML(station.tagline) + '</p>' +
        '  <div class="radio-station-meta">' +
        '    <span>' + escapeHTML(station.genre) + '</span>' +
        '  </div>' +
        '  <div class="radio-station-actions">' +
        '    <button type="button" class="btn' + (active ? '' : ' secondary') + ' small" data-radio-play="' + escapeHTML(station.id) + '">' + (active && state.playing ? 'Sonando' : 'Reproducir') + '</button>' +
        (station.sourceUrl ? '    <a href="' + escapeHTML(station.sourceUrl) + '" target="_blank" rel="noopener" class="btn secondary small">Fuente</a>' : '') +
        (station.custom ? '    <button type="button" class="btn danger small" data-radio-delete="' + escapeHTML(station.id) + '">Eliminar</button>' : '') +
        '  </div>' +
        '</article>';
    }).join("");
    updateCountryControls();
  }

  function updateMiniPlayer() {
    if (!state.enabled) {
      mini.hidden = true;
      renderGrid();
      return;
    }
    var station = stationById(state.stationId);
    if (!station) {
      mini.hidden = true;
      renderGrid();
      return;
    }
    mini.hidden = false;
    miniName.textContent = station.name;
    miniTagline.textContent = station.tagline;
    miniVolume.value = String(state.volume);
    miniPlayPause.textContent = state.playing ? "Pausar" : "Reanudar";
    renderGrid();
  }

  function playStation(id, autoplay) {
    if (!state.enabled) {
      setRadioEnabled(true);
      persistRadioConfig();
    }
    var station = stationById(id);
    if (!station) return;
    state.stationId = station.id;
    state.volume = Number(miniVolume.value || state.volume || 0.7);
    miniAudio.volume = state.volume;
    if (miniAudio.src !== station.streamUrl) {
      miniAudio.src = station.streamUrl;
    }
    if (autoplay !== false) {
      var playPromise = miniAudio.play();
      state.playing = true;
      if (playPromise && typeof playPromise.catch === "function") {
        playPromise.catch(function () {
          state.playing = false;
          updateMiniPlayer();
          saveState();
        });
      }
    } else {
      state.playing = false;
    }
    setDrawerOpen(false);
    updateMiniPlayer();
    saveState();
  }

  function stopPlayback() {
    miniAudio.pause();
    miniAudio.removeAttribute("src");
    miniAudio.load();
    state.stationId = "";
    state.playing = false;
    saveState();
    updateMiniPlayer();
    renderGrid();
  }

  function closeRadioCompletely() {
    setDrawerOpen(false);
    state.enabled = false;
    try {
      window.localStorage.setItem(ENABLED_KEY, "0");
    } catch (_) {}
    stopPlayback();
    if (enabledToggle) enabledToggle.checked = false;
    if (enabledStatus) enabledStatus.textContent = "Reproductor apagado";
    if (openBtn) {
      openBtn.hidden = true;
      openBtn.setAttribute("aria-hidden", "true");
      openBtn.setAttribute("aria-pressed", "false");
      openBtn.classList.add("is-disabled");
    }
    persistRadioConfig();
  }

  function setRadioEnabled(enabled) {
    state.enabled = !!enabled;
    try {
      window.localStorage.setItem(ENABLED_KEY, state.enabled ? "1" : "0");
    } catch (_) {}
    if (!state.enabled) {
      setDrawerOpen(false);
      stopPlayback();
    }
    if (enabledToggle) enabledToggle.checked = state.enabled;
    if (enabledStatus) enabledStatus.textContent = state.enabled ? "Lista para reproducir" : "Reproductor apagado";
    if (openBtn) {
      openBtn.hidden = !state.enabled;
      openBtn.classList.toggle("is-disabled", !state.enabled);
      openBtn.setAttribute("aria-hidden", state.enabled ? "false" : "true");
      openBtn.setAttribute("aria-pressed", state.enabled ? "true" : "false");
      openBtn.setAttribute("title", "Abrir emisoras online");
      var label = openBtn.querySelector(".ai-chat-toggle-label");
      if (label) label.textContent = "Emisoras";
    }
    renderGrid();
    updateMiniPlayer();
  }

  function persistRadioConfig() {
    return fetch(buildPrefsEndpoint(), {
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
      applyPreferencePayload(data || {});
    }).catch(function (err) {
      if (enabledStatus) {
        enabledStatus.textContent = state.enabled ? "Activa localmente, sin guardar" : "Apagada localmente, sin guardar";
      }
      console.warn("No se pudo guardar la configuracion de emisoras:", err);
    });
  }

  function applyPreferencePayload(data) {
    if (data && typeof data.radio_online_enabled === "boolean") {
      state.enabled = data.radio_online_enabled;
      try { window.localStorage.setItem(ENABLED_KEY, state.enabled ? "1" : "0"); } catch (_) {}
      if (!state.enabled) {
        state.stationId = "";
        state.playing = false;
        saveState();
      }
    }
    if (data && Array.isArray(data.radio_custom_stations)) {
      state.customStations = countryTools.normalizeCustomList(data.radio_custom_stations);
    }
    if (data && Object.prototype.hasOwnProperty.call(data, "radio_country")) {
      var savedCountry = countryTools.normalizeCountry(data.radio_country);
      if (savedCountry) {
        state.countryCode = savedCountry;
        state.countrySource = "empresa";
      }
    }
    setRadioEnabled(state.enabled);
    updateCountryControls();
  }

  function fetchDetectedCountry() {
    return fetch(buildCountryEndpoint(), { credentials: "same-origin" })
      .then(function (res) {
        if (!res.ok) throw new Error("sin país empresa");
        return res.json();
      })
      .then(function (data) {
        return {
          country: countryTools.normalizeCountry(data && (data.pais_codigo || data.country_code || data.country)),
          source: data && data.source ? String(data.source) : "empresa"
        };
      })
      .catch(function () {
        return fetch("/api/public/geo", { credentials: "same-origin" })
          .then(function (res) {
            if (!res.ok) throw new Error("sin geo");
            return res.json();
          })
          .then(function (data) {
            return {
              country: countryTools.normalizeCountry(data && (data.pais_codigo || data.country_code || data.country)),
              source: data && data.source ? String(data.source) : "ip"
            };
          });
      });
  }

  function loadCompanyRadioPreference() {
    fetch(buildPrefsEndpoint(), { credentials: "same-origin" })
      .then(function (res) {
        if (!res.ok) return {};
        return res.json();
      })
      .then(function (data) {
        applyPreferencePayload(data || {});
        if (state.countryCode) return null;
        return fetchDetectedCountry().then(function (detected) {
          if (detected && detected.country) {
            state.countryCode = detected.country;
            state.countrySource = detected.source || "detectado";
            renderGrid();
            updateCountryControls();
          }
          return null;
        });
      })
      .catch(function () {
        renderGrid();
      });
  }

  function togglePlayback() {
    if (!state.enabled || !state.stationId) return;
    if (miniAudio.paused) {
      miniAudio.play().then(function () {
        state.playing = true;
        updateMiniPlayer();
        saveState();
      }).catch(function () {});
    } else {
      miniAudio.pause();
      state.playing = false;
      updateMiniPlayer();
      saveState();
    }
  }

  function addCustomStation() {
    var station = countryTools.normalizeCustomList([{
      name: customName ? customName.value : "",
      genre: customGenre ? customGenre.value : "",
      streamUrl: customStream ? customStream.value : "",
      sourceUrl: customSource ? customSource.value : "",
      countryCode: customCountry ? customCountry.value : state.countryCode,
      country: countryLabel(customCountry ? customCountry.value : state.countryCode),
      tagline: "Emisora personalizada de esta empresa."
    }])[0];
    if (!station) {
      if (countryStatus) countryStatus.textContent = "Escribe nombre y URL http/https valida para agregar la emisora.";
      return;
    }
    state.customStations = state.customStations.filter(function (item) { return item.id !== station.id; });
    state.customStations.push(station);
    if (customForm) customForm.reset();
    if (customCountry && state.countryCode) customCountry.value = state.countryCode;
    renderGrid();
    persistRadioConfig();
  }

  function deleteCustomStation(id) {
    state.customStations = state.customStations.filter(function (item) { return item.id !== id; });
    if (state.stationId === id) stopPlayback();
    renderGrid();
    persistRadioConfig();
  }

  function wireEvents() {
    if (openBtn) openBtn.addEventListener("click", function () {
      if (!drawer) return;
      if (!state.enabled) {
        setRadioEnabled(true);
        persistRadioConfig();
      }
      setDrawerOpen(!drawer.classList.contains("is-open"));
    });
    if (closeBtn) closeBtn.addEventListener("click", closeRadioCompletely);
    if (closeBtnBottom) closeBtnBottom.addEventListener("click", closeRadioCompletely);
    if (enabledToggle) enabledToggle.addEventListener("change", function () {
      setRadioEnabled(!!enabledToggle.checked);
      persistRadioConfig();
    });
    if (countrySelect) countrySelect.addEventListener("change", function () {
      var selectedCountry = countryTools.normalizeCountry(countrySelect.value);
      if (!selectedCountry) {
        fetchDetectedCountry().then(function (detected) {
          state.countryCode = detected && detected.country ? detected.country : "";
          state.countrySource = detected && detected.source ? detected.source : "detectado";
          renderGrid();
          persistRadioConfig();
        }).catch(function () {
          state.countryCode = "";
          renderGrid();
          persistRadioConfig();
        });
        return;
      }
      state.countryCode = selectedCountry;
      state.countrySource = "empresa";
      renderGrid();
      persistRadioConfig();
    });
    if (customForm) customForm.addEventListener("submit", function (ev) {
      ev.preventDefault();
      addCustomStation();
    });
    if (miniClose) miniClose.addEventListener("click", closeRadioCompletely);
    if (miniPlayPause) miniPlayPause.addEventListener("click", togglePlayback);
    if (miniVolume) miniVolume.addEventListener("input", function () {
      state.volume = Number(miniVolume.value || 0.7);
      miniAudio.volume = state.volume;
      saveState();
    });
    if (grid) {
      grid.addEventListener("click", function (ev) {
        var deleteButton = ev.target.closest("[data-radio-delete]");
        if (deleteButton) {
          deleteCustomStation(deleteButton.getAttribute("data-radio-delete"));
          return;
        }
        var button = ev.target.closest("[data-radio-play]");
        if (!button) return;
        playStation(button.getAttribute("data-radio-play"), true);
      });
    }
    miniAudio.addEventListener("pause", function () {
      state.playing = false;
      updateMiniPlayer();
      saveState();
    });
    miniAudio.addEventListener("play", function () {
      state.playing = true;
      updateMiniPlayer();
      saveState();
    });
    document.addEventListener("keydown", function (ev) {
      if (ev.key === "Escape" && drawer && drawer.classList.contains("is-open")) {
        setDrawerOpen(false);
      }
    });
  }

  window.__pcsRadioPlayerOpenStation = function (id) {
    if (!state.enabled) return;
    setDrawerOpen(true);
    playStation(id, true);
  };

  window.__pcsRadioPlayerOpenDrawer = function () {
    if (!state.enabled) {
      setRadioEnabled(true);
      persistRadioConfig();
    }
    if (openBtn) {
      openBtn.hidden = false;
      openBtn.setAttribute("aria-hidden", "false");
    }
    setDrawerOpen(true);
    renderGrid();
    updateCountryControls();
  };

  window.__pcsRadioPlayerSetEnabled = setRadioEnabled;
  window.__pcsRadioPlayerIsEnabled = function () {
    return !!state.enabled;
  };
  window.__pcsRadioPlayerReloadConfig = loadCompanyRadioPreference;

  loadState();
  wireEvents();
  renderGrid();
  setRadioEnabled(state.enabled);
  loadCompanyRadioPreference();
  if (state.enabled && state.stationId) {
    playStation(state.stationId, false);
    updateMiniPlayer();
  }
})();
