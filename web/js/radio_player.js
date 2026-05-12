(function () {
  "use strict";

  if (!window.__pcsRadioStations || !document.getElementById("radioMiniAudio")) {
    return;
  }

  var STORAGE_KEY = "pcs_radio_player_state";
  var ENABLED_KEY = "pcs_radio_online_enabled";
  var PREFS_ENDPOINT = "/api/chat_flotante/preferencias";
  var stations = window.__pcsRadioStations.slice();
  var drawer = document.getElementById("radioDrawer");
  var openBtn = document.getElementById("openRadioDrawer");
  var closeBtn = document.getElementById("closeRadioDrawer");
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

  var state = {
    stationId: "",
    playing: false,
    volume: 0.7,
    enabled: true
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

  function saveState() {
    try {
      window.localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
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
    try {
      state.enabled = window.localStorage.getItem(ENABLED_KEY) !== "0";
    } catch (_) {
      state.enabled = true;
    }
  }

  function stationById(id) {
    return stations.find(function (item) { return item.id === id; }) || null;
  }

  function setDrawerOpen(open) {
    if (!drawer || !openBtn) return;
    if (open && !state.enabled) {
      open = false;
    }
    drawer.classList.toggle("is-open", !!open);
    drawer.setAttribute("aria-hidden", open ? "false" : "true");
    openBtn.setAttribute("aria-expanded", open ? "true" : "false");
  }

  function renderGrid() {
    if (!grid) return;
    grid.classList.toggle("is-disabled", !state.enabled);
    grid.innerHTML = stations.map(function (station) {
      var active = state.stationId === station.id;
      return '' +
        '<article class="radio-station-card' + (active ? ' is-active' : '') + '">' +
        '  <div class="radio-station-badge">' + escapeHTML(station.country) + '</div>' +
        '  <h3>' + escapeHTML(station.name) + '</h3>' +
        '  <p>' + escapeHTML(station.tagline) + '</p>' +
        '  <div class="radio-station-meta">' +
        '    <span>' + escapeHTML(station.genre) + '</span>' +
        '  </div>' +
        '  <div class="radio-station-actions">' +
        '    <button type="button" class="btn' + (active ? '' : ' secondary') + ' small" data-radio-play="' + escapeHTML(station.id) + '"' + (!state.enabled ? ' disabled' : '') + '>' + (!state.enabled ? 'Desactivada' : (active && state.playing ? 'Sonando' : 'Escuchar')) + '</button>' +
        '    <a href="' + escapeHTML(station.sourceUrl) + '" target="_blank" rel="noopener" class="btn secondary small">Fuente</a>' +
        '  </div>' +
        '</article>';
    }).join("");
  }

  function escapeHTML(value) {
    return String(value || "").replace(/[&<>\"']/g, function (c) { return ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "\"": "&quot;", "'": "&#39;" })[c]; });
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
    if (!state.enabled) return;
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
      openBtn.setAttribute("title", "Abrir musica latina online");
      var label = openBtn.querySelector(".ai-chat-toggle-label");
      if (label) label.textContent = "Musica latina";
    }
    renderGrid();
    updateMiniPlayer();
  }

  function persistRadioEnabled(enabled) {
    return fetch(buildPrefsEndpoint(), {
      method: "PUT",
      credentials: "same-origin",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ radio_online_enabled: !!enabled })
    }).then(function (res) {
      if (!res.ok) throw new Error("No se pudo guardar la emisora.");
      return res.json();
    }).then(function (data) {
      if (data && typeof data.radio_online_enabled === "boolean") {
        setRadioEnabled(data.radio_online_enabled);
      }
    }).catch(function (err) {
      if (enabledStatus) {
        enabledStatus.textContent = state.enabled ? "Activa localmente, sin guardar" : "Apagada localmente, sin guardar";
      }
      console.warn("No se pudo guardar la preferencia de emisora:", err);
    });
  }

  function loadCompanyRadioPreference() {
    fetch(buildPrefsEndpoint(), { credentials: "same-origin" })
      .then(function (res) {
        if (!res.ok) return null;
        return res.json();
      })
      .then(function (data) {
        if (data && typeof data.radio_online_enabled === "boolean") {
          setRadioEnabled(data.radio_online_enabled);
        }
      })
      .catch(function () {});
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

  function wireEvents() {
    if (openBtn) openBtn.addEventListener("click", function () {
      if (!drawer) return;
      if (!state.enabled) return;
      setDrawerOpen(!drawer.classList.contains("is-open"));
    });
    if (closeBtn) closeBtn.addEventListener("click", function () { setDrawerOpen(false); });
    if (enabledToggle) enabledToggle.addEventListener("change", function () {
      setRadioEnabled(!!enabledToggle.checked);
      persistRadioEnabled(state.enabled);
    });
    if (miniClose) miniClose.addEventListener("click", stopPlayback);
    if (miniPlayPause) miniPlayPause.addEventListener("click", togglePlayback);
    if (miniVolume) miniVolume.addEventListener("input", function () {
      state.volume = Number(miniVolume.value || 0.7);
      miniAudio.volume = state.volume;
      saveState();
    });
    if (grid) {
      grid.addEventListener("click", function (ev) {
        var button = ev.target.closest("[data-radio-play]");
        if (!button) return;
        if (!state.enabled) return;
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

  window.__pcsRadioPlayerSetEnabled = setRadioEnabled;
  window.__pcsRadioPlayerIsEnabled = function () {
    return !!state.enabled;
  };

  loadState();
  wireEvents();
  renderGrid();
  setRadioEnabled(state.enabled);
  loadCompanyRadioPreference();
  if (state.stationId) {
    playStation(state.stationId, false);
    updateMiniPlayer();
  }
})();
