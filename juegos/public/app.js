(() => {
  const allowed = {
    dark: true,
    "dark-violet": true,
    "dark-emerald": true,
    light: true,
    "light-rose": true,
    "light-gold": true
  };
  function cookie(name) {
    const match = String(document.cookie || "").match(`(^|;)\\s*${name}\\s*=\\s*([^;]+)`);
    return match ? decodeURIComponent(match.pop()) : "";
  }
  function theme() {
    let value = "";
    try {
      value = localStorage.getItem("theme") || cookie("pcs_theme") || "";
    } catch (_) {
      value = cookie("pcs_theme") || "";
    }
    value = String(value || "").trim().toLowerCase();
    if (value === "dark-protect") value = "dark";
    return allowed[value] ? value : "dark";
  }
  document.documentElement.setAttribute("data-theme", theme());
  window.addEventListener("storage", (event) => {
    if (event && event.key === "theme") {
      document.documentElement.setAttribute("data-theme", theme());
    }
  });
})();

const state = {
  core: "snes",
  roms: [],
  selected: "",
  loaded: false,
  empresaId: resolveEmpresaId(),
  latestSave: null,
  autoSaveTimer: null,
  basePath: normalizeBasePath(window.JUEGOS_BASE_PATH || inferBasePath()),
  loaderPath: "/emulator/data/loader.js",
  dataPath: "/emulator/data/"
};

const els = {
  romSelect: document.getElementById("romSelect"),
  romGrid: document.getElementById("romGrid"),
  playButton: document.getElementById("playButton"),
  saveNowButton: document.getElementById("saveNowButton"),
  loadSaveButton: document.getElementById("loadSaveButton"),
  refreshRoms: document.getElementById("refreshRoms"),
  fullscreenButton: document.getElementById("fullscreenButton"),
  reloadButton: document.getElementById("reloadButton"),
  currentTitle: document.getElementById("currentTitle"),
  currentSystem: document.getElementById("currentSystem"),
  coreBadge: document.getElementById("coreBadge"),
  gamepadBadge: document.getElementById("gamepadBadge"),
  message: document.getElementById("message"),
  saveStatus: document.getElementById("saveStatus"),
  playerShell: document.getElementById("playerShell"),
  emptyState: document.getElementById("emptyState"),
  game: document.getElementById("game")
};

function inferBasePath() {
  const currentPath = window.location.pathname || "/";
  if (currentPath === "/" || currentPath === "/index.html") return "";
  const withoutSlash = currentPath.endsWith("/") ? currentPath.slice(0, -1) : currentPath;
  const base = withoutSlash.slice(0, withoutSlash.lastIndexOf("/"));
  return base || withoutSlash;
}

function normalizeBasePath(path) {
  const value = String(path || "").trim();
  if (!value || value === "/") return "";
  const withSlash = value.startsWith("/") ? value : `/${value}`;
  return withSlash.replace(/\/+$/, "");
}

function appPath(path) {
  const clean = String(path || "");
  if (/^https?:\/\//i.test(clean)) return clean;
  return `${state.basePath}${clean.startsWith("/") ? clean : `/${clean}`}`;
}

function setMessage(text, isError = false) {
  els.message.textContent = text || "";
  els.message.classList.toggle("is-error", Boolean(isError));
}

function setSaveStatus(text, isError = false) {
  els.saveStatus.textContent = text || "";
  els.saveStatus.classList.toggle("is-error", Boolean(isError));
}

function parsePositiveInt(value) {
  const parsed = Number.parseInt(String(value || "").trim(), 10);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : 0;
}

function resolveEmpresaId() {
  const params = new URLSearchParams(window.location.search);
  const fromUrl = parsePositiveInt(params.get("empresa_id") || params.get("id"));
  if (fromUrl) return String(fromUrl);
  const keys = ["active_empresa_id", "empresa_id", "admin_empresa_id"];
  const stores = [window.sessionStorage, window.localStorage];
  for (const store of stores) {
    for (const key of keys) {
      try {
        const value = parsePositiveInt(store.getItem(key));
        if (value) return String(value);
      } catch (_) {
        /* Storage can be blocked. */
      }
    }
  }
  return "publico";
}

function formatBytes(bytes) {
  const value = Number(bytes || 0);
  if (value < 1024) return `${value} B`;
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`;
  return `${(value / 1024 / 1024).toFixed(1)} MB`;
}

function getRomFromUrl() {
  return new URLSearchParams(window.location.search).get("rom") || "";
}

function persistSelectedRom(file) {
  try {
    localStorage.setItem(`juegos.selectedRom.${state.empresaId}`, file || "");
  } catch (_) {
    /* localStorage can be unavailable in restricted browsers. */
  }
}

function getPersistedRom() {
  try {
    return localStorage.getItem(`juegos.selectedRom.${state.empresaId}`) || localStorage.getItem("juegos.selectedRom") || "";
  } catch (_) {
    return "";
  }
}

function selectRom(file) {
  state.selected = file || "";
  els.romSelect.value = state.selected;
  state.latestSave = null;
  persistSelectedRom(state.selected);
  renderROMCards();
  refreshLatestSaveStatus().catch(() => {});
}

async function loadROMs() {
  setMessage("Cargando ROMs...");
  const response = await fetch(appPath("/api/roms"), { credentials: "same-origin" });
  if (!response.ok) throw new Error("No fue posible cargar /api/roms");
  const data = await response.json();
  state.core = data.core || "snes";
  state.roms = Array.isArray(data.roms) ? data.roms : [];
  els.coreBadge.textContent = `Cores: ${summarizeCores(state.roms) || state.core}`;
  if (state.empresaId !== "publico") {
    els.coreBadge.textContent += ` | Empresa ${state.empresaId}`;
  }
  renderROMSelect();
  renderROMCards();

  const requested = getRomFromUrl();
  const remembered = getPersistedRom();
  const fallback = state.roms[0] ? state.roms[0].file : "";
  const next = findROM(requested) ? requested : (findROM(remembered) ? remembered : fallback);
  selectRom(next);
  setMessage(state.roms.length ? "ROMs listas." : "No hay ROMs disponibles. Copia archivos legales/homebrew en /roms.");
}

function findROM(file) {
  return state.roms.find((rom) => rom.file === file);
}

function saveURL(endpoint, rom) {
  const params = new URLSearchParams();
  params.set("empresa_id", state.empresaId || "publico");
  params.set("rom", rom.file || state.selected || "");
  return appPath(`${endpoint}?${params.toString()}`);
}

function renderROMSelect() {
  els.romSelect.innerHTML = "";
  if (!state.roms.length) {
    const option = document.createElement("option");
    option.value = "";
    option.textContent = "Sin ROMs";
    els.romSelect.appendChild(option);
    els.playButton.disabled = true;
    return;
  }

  const grouped = groupROMsBySystem(state.roms);
  for (const group of grouped) {
    const optgroup = document.createElement("optgroup");
    optgroup.label = `${group.system} (${group.roms.length})`;
    for (const rom of group.roms) {
      const option = document.createElement("option");
      option.value = rom.file;
      option.textContent = rom.name;
      optgroup.appendChild(option);
    }
    els.romSelect.appendChild(optgroup);
  }
  els.playButton.disabled = false;
}

function groupROMsBySystem(roms) {
  const order = ["NES", "SNES", "Nintendo 64", "Game Boy", "Game Boy Color", "Game Boy Advance", "Mega Drive", "ZIP", "ROM"];
  const groups = new Map();
  for (const rom of roms) {
    const system = rom.system || rom.core || "ROM";
    if (!groups.has(system)) groups.set(system, []);
    groups.get(system).push(rom);
  }
  return Array.from(groups.entries())
    .map(([system, items]) => ({
      system,
      roms: items.slice().sort((a, b) => String(a.name || "").localeCompare(String(b.name || ""), "es", { sensitivity: "base" }))
    }))
    .sort((a, b) => {
      const ai = order.indexOf(a.system);
      const bi = order.indexOf(b.system);
      if (ai >= 0 || bi >= 0) return (ai >= 0 ? ai : 999) - (bi >= 0 ? bi : 999);
      return a.system.localeCompare(b.system, "es", { sensitivity: "base" });
    });
}

function renderROMCards() {
  if (!els.romGrid) return;
  els.romGrid.innerHTML = "";
  els.romGrid.hidden = true;
}

function configureEmulator(rom, latestSave) {
  window.EJS_player = "#game";
  window.EJS_core = rom.core || state.core || "snes";
  window.EJS_gameUrl = appPath(rom.url);
  window.EJS_gameName = rom.name;
  window.EJS_gameID = `${state.empresaId || "publico"}:${rom.core || state.core || "core"}:${rom.file}`;
  window.EJS_pathtodata = appPath(state.dataPath);
  window.EJS_startOnLoaded = true;
  window.EJS_AdUrl = "";
  window.EJS_color = "#38bdf8";
  window.EJS_backgroundColor = "#000000";
  window.EJS_fullscreenOnLoaded = false;
  window.EJS_volume = 0.65;
  window.EJS_language = "es";
  window.EJS_fixedSaveInterval = 10000;
  if (latestSave && latestSave.state_url) {
    window.EJS_loadStateURL = `${appPath(latestSave.state_url)}&t=${Date.now()}`;
  } else {
    try {
      delete window.EJS_loadStateURL;
    } catch (_) {
      window.EJS_loadStateURL = "";
    }
  }
  window.EJS_onSaveState = (event) => handleSaveStateEvent(event, "emulator-button");
  window.EJS_onSaveUpdate = (event) => handleSaveUpdateEvent(event, "emulator-save-update");
  window.EJS_ready = () => {
    setSaveStatus(latestSave && latestSave.state_url ? "Partida anterior cargada desde la carpeta de la empresa." : "Emulador listo. Usa Guardar avance para conservar la partida.");
    startAutoSaveState();
  };
}

function stopAutoSaveState() {
  if (state.autoSaveTimer) {
    window.clearInterval(state.autoSaveTimer);
    state.autoSaveTimer = null;
  }
}

function startAutoSaveState() {
  stopAutoSaveState();
  state.autoSaveTimer = window.setInterval(() => {
    captureAndUploadCurrentState("auto").catch(() => {});
  }, 60000);
}

async function fetchLatestSave(rom) {
  const response = await fetch(saveURL("/api/saves/latest", rom), { credentials: "same-origin", cache: "no-store" });
  if (!response.ok) throw new Error("No fue posible consultar partidas guardadas.");
  return response.json();
}

async function refreshLatestSaveStatus() {
  const rom = findROM(state.selected);
  if (!rom) {
    setSaveStatus("");
    if (els.loadSaveButton) els.loadSaveButton.disabled = true;
    return null;
  }
  try {
    const latest = await fetchLatestSave(rom);
    state.latestSave = latest;
    const hasState = Boolean(latest && latest.state_url);
    if (els.loadSaveButton) els.loadSaveButton.disabled = !hasState;
    if (hasState) {
      setSaveStatus(`Partida guardada disponible para esta empresa: ${formatDateTime(latest.updated_at)}.`);
    } else {
      setSaveStatus(state.empresaId === "publico" ? "Sin empresa activa: se guardara en carpeta publica." : "Sin partida guardada para esta empresa.");
    }
    return latest;
  } catch (error) {
    if (els.loadSaveButton) els.loadSaveButton.disabled = true;
    setSaveStatus(error.message || "No se pudo consultar la partida.", true);
    return null;
  }
}

function formatDateTime(value) {
  if (!value) return "sin fecha";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString("es-CO", { dateStyle: "short", timeStyle: "short" });
}

function bytesToBase64(bytes) {
  let binary = "";
  const chunkSize = 0x8000;
  for (let i = 0; i < bytes.length; i += chunkSize) {
    binary += String.fromCharCode.apply(null, bytes.subarray(i, i + chunkSize));
  }
  return window.btoa(binary);
}

async function toBase64Payload(value) {
  if (!value) return "";
  if (typeof value === "string") {
    if (value.startsWith("data:")) {
      const comma = value.indexOf(",");
      return comma >= 0 ? value.slice(comma + 1) : "";
    }
    return window.btoa(unescape(encodeURIComponent(value)));
  }
  if (value instanceof Blob) {
    value = await value.arrayBuffer();
  }
  if (value instanceof ArrayBuffer) {
    return bytesToBase64(new Uint8Array(value));
  }
  if (ArrayBuffer.isView(value)) {
    return bytesToBase64(new Uint8Array(value.buffer, value.byteOffset, value.byteLength));
  }
  if (Array.isArray(value)) {
    return bytesToBase64(new Uint8Array(value));
  }
  if (value.data) {
    return toBase64Payload(value.data);
  }
  if (value.state) {
    return toBase64Payload(value.state);
  }
  return "";
}

function extractSaveStatePayload(event) {
  if (Array.isArray(event)) {
    return { screenshot: event[0], data: event[1] };
  }
  if (event && typeof event === "object") {
    return {
      screenshot: event.screenshot || event.image || event.preview || "",
      data: event.state || event.saveState || event.data || event.save || ""
    };
  }
  return { screenshot: "", data: event };
}

async function uploadSave(kind, data, screenshot, source) {
  const rom = findROM(state.selected);
  if (!rom) throw new Error("Selecciona un juego antes de guardar.");
  const dataBase64 = await toBase64Payload(data);
  if (!dataBase64) throw new Error("El emulador no entrego datos de partida para guardar.");
  const screenshotBase64 = screenshot ? await toBase64Payload(screenshot) : "";
  const endpoint = kind === "state" ? "/api/saves/state" : "/api/saves/file";
  const response = await fetch(appPath(endpoint), {
    method: "POST",
    credentials: "same-origin",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      empresa_id: state.empresaId || "publico",
      rom: rom.file,
      core: rom.core || state.core || "",
      data_base64: dataBase64,
      screenshot_base64: screenshotBase64,
      source: source || "manual"
    })
  });
  if (!response.ok) {
    const text = await response.text().catch(() => "");
    throw new Error(text || "No fue posible guardar la partida en el servidor.");
  }
  const result = await response.json();
  await refreshLatestSaveStatus();
  return result;
}

async function handleSaveStateEvent(event, source) {
  try {
    const payload = extractSaveStatePayload(event);
    const result = await uploadSave("state", payload.data, payload.screenshot, source);
    setSaveStatus(`Partida guardada en carpeta de empresa (${formatBytes(result.bytes)}).`);
  } catch (error) {
    setSaveStatus(error.message || "No se pudo guardar la partida.", true);
  }
}

async function handleSaveUpdateEvent(event, source) {
  try {
    const data = event && typeof event === "object" ? (event.save || event.data) : event;
    const screenshot = event && typeof event === "object" ? event.screenshot : "";
    if (!data) return;
    const result = await uploadSave("file", data, screenshot, source);
    setSaveStatus(`Guardado interno sincronizado (${formatBytes(result.bytes)}).`);
  } catch (error) {
    setSaveStatus(error.message || "No se pudo sincronizar el guardado interno.", true);
  }
}

async function captureAndUploadCurrentState(source) {
  const manager = window.EJS_emulator && window.EJS_emulator.gameManager;
  if (!manager || typeof manager.getState !== "function") {
    if (source === "manual") {
      setSaveStatus("Esta version del emulador no expone guardado directo. Abre el menu del emulador y usa Guardar estado.", true);
    }
    return null;
  }
  const currentState = await Promise.resolve(manager.getState());
  if (!currentState) {
    if (source === "manual") setSaveStatus("No se obtuvo estado del juego para guardar.", true);
    return null;
  }
  const result = await uploadSave("state", currentState, "", source || "manual");
  if (source === "manual") {
    setSaveStatus(`Avance guardado en carpeta de empresa (${formatBytes(result.bytes)}).`);
  }
  return result;
}

function summarizeCores(roms) {
  const systems = Array.from(new Set(roms.map((rom) => rom.system || rom.core).filter(Boolean)));
  return systems.slice(0, 4).join(", ") + (systems.length > 4 ? ` +${systems.length - 4}` : "");
}

function resetPlayerDOM() {
  els.game.innerHTML = "";
  els.emptyState.style.display = "none";
}

function injectLoader() {
  return new Promise((resolve, reject) => {
    const script = document.createElement("script");
    script.src = `${appPath(state.loaderPath)}?v=${Date.now()}`;
    script.async = true;
    script.onload = resolve;
    script.onerror = () => reject(new Error("No se pudo cargar EmulatorJS. Verifica /emulator/data/loader.js"));
    document.body.appendChild(script);
  });
}

async function startSelectedGame() {
  const rom = findROM(state.selected);
  if (!rom) {
    setMessage("Selecciona una ROM valida.", true);
    return;
  }

  const url = new URL(window.location.href);
  url.searchParams.set("rom", rom.file);
  if (state.empresaId && state.empresaId !== "publico") {
    url.searchParams.set("empresa_id", state.empresaId);
  }
  window.history.replaceState({}, "", url.toString());

  if (state.loaded) {
    window.location.href = url.toString();
    return;
  }

  try {
    setMessage(`Cargando ${rom.name}...`);
    const latestSave = await fetchLatestSave(rom).catch((error) => {
      setSaveStatus(error.message || "No se pudo consultar la partida guardada.", true);
      return null;
    });
    state.latestSave = latestSave;
    resetPlayerDOM();
    configureEmulator(rom, latestSave);
    await injectLoader();
    state.loaded = true;
    if (els.currentSystem) els.currentSystem.textContent = rom.system || rom.core || "Consola";
    els.currentTitle.textContent = rom.name;
    setMessage("Juego cargado. En movil, usa el boton tactil del emulador para controles y menu.");
    if (els.loadSaveButton) els.loadSaveButton.disabled = !(latestSave && latestSave.state_url);
  } catch (error) {
    els.emptyState.style.display = "";
    setMessage(error.message || "No se pudo iniciar el juego.", true);
  }
}

async function enterFullscreen() {
  const target = els.playerShell;
  if (!target.requestFullscreen) {
    setMessage("Pantalla completa no soportada por este navegador.", true);
    return;
  }
  await target.requestFullscreen();
}

function updateGamepadStatus() {
  const pads = typeof navigator.getGamepads === "function" ? Array.from(navigator.getGamepads()).filter(Boolean) : [];
  els.gamepadBadge.textContent = pads.length ? `Gamepad: ${pads[0].id || "detectado"}` : "Gamepad: no detectado";
}

function bindEvents() {
  els.romSelect.addEventListener("change", () => {
    selectRom(els.romSelect.value);
    if (els.romSelect.value) {
      startSelectedGame();
    }
  });
  els.playButton.addEventListener("click", startSelectedGame);
  els.saveNowButton.addEventListener("click", () => {
    captureAndUploadCurrentState("manual").catch((error) => setSaveStatus(error.message || "No se pudo guardar el avance.", true));
  });
  els.loadSaveButton.addEventListener("click", async () => {
    const rom = findROM(state.selected);
    if (!rom) return;
    const latest = await fetchLatestSave(rom).catch((error) => {
      setSaveStatus(error.message || "No se pudo consultar la partida.", true);
      return null;
    });
    if (!latest || !latest.state_url) {
      setSaveStatus("No hay una partida guardada para retomar.", true);
      return;
    }
    const url = new URL(window.location.href);
    url.searchParams.set("rom", rom.file);
    if (state.empresaId && state.empresaId !== "publico") {
      url.searchParams.set("empresa_id", state.empresaId);
    }
    window.location.href = url.toString();
  });
  els.refreshRoms.addEventListener("click", () => loadROMs().catch((err) => setMessage(err.message, true)));
  els.fullscreenButton.addEventListener("click", () => enterFullscreen().catch((err) => setMessage(err.message, true)));
  els.reloadButton.addEventListener("click", () => {
    if (state.selected) window.location.reload();
  });
  window.addEventListener("gamepadconnected", updateGamepadStatus);
  window.addEventListener("gamepaddisconnected", updateGamepadStatus);
}

async function bootstrap() {
  bindEvents();
  updateGamepadStatus();
  if (els.loadSaveButton) els.loadSaveButton.disabled = true;
  await loadROMs();
  const requested = getRomFromUrl();
  if (requested && findROM(requested)) {
    await startSelectedGame();
  }
}

bootstrap().catch((err) => setMessage(err.message || "Error inicializando Juegos.", true));
