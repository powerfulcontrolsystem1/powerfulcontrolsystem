const state = {
  core: "snes",
  roms: [],
  selected: "",
  loaded: false,
  basePath: normalizeBasePath(window.JUEGOS_BASE_PATH || inferBasePath()),
  loaderPath: "/emulator/data/loader.js",
  dataPath: "/emulator/data/"
};

const els = {
  romSelect: document.getElementById("romSelect"),
  romGrid: document.getElementById("romGrid"),
  playButton: document.getElementById("playButton"),
  refreshRoms: document.getElementById("refreshRoms"),
  fullscreenButton: document.getElementById("fullscreenButton"),
  reloadButton: document.getElementById("reloadButton"),
  currentTitle: document.getElementById("currentTitle"),
  coreBadge: document.getElementById("coreBadge"),
  gamepadBadge: document.getElementById("gamepadBadge"),
  message: document.getElementById("message"),
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
    localStorage.setItem("juegos.selectedRom", file || "");
  } catch (_) {
    /* localStorage can be unavailable in restricted browsers. */
  }
}

function getPersistedRom() {
  try {
    return localStorage.getItem("juegos.selectedRom") || "";
  } catch (_) {
    return "";
  }
}

function selectRom(file) {
  state.selected = file || "";
  els.romSelect.value = state.selected;
  persistSelectedRom(state.selected);
  renderROMCards();
}

async function loadROMs() {
  setMessage("Cargando ROMs...");
  const response = await fetch(appPath("/api/roms"), { credentials: "same-origin" });
  if (!response.ok) throw new Error("No fue posible cargar /api/roms");
  const data = await response.json();
  state.core = data.core || "snes";
  state.roms = Array.isArray(data.roms) ? data.roms : [];
  els.coreBadge.textContent = `Core: ${state.core}`;
  renderROMSelect();
  renderROMCards();

  const requested = getRomFromUrl();
  const remembered = getPersistedRom();
  const fallback = state.roms[0] ? state.roms[0].file : "";
  const next = findROM(requested) ? requested : (findROM(remembered) ? remembered : fallback);
  selectRom(next);
  setMessage(state.roms.length ? "ROMs listas." : "No hay ROMs disponibles. Copia archivos .sfc o .smc en /roms.");
}

function findROM(file) {
  return state.roms.find((rom) => rom.file === file);
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

  for (const rom of state.roms) {
    const option = document.createElement("option");
    option.value = rom.file;
    option.textContent = rom.name;
    els.romSelect.appendChild(option);
  }
  els.playButton.disabled = false;
}

function renderROMCards() {
  els.romGrid.innerHTML = "";
  if (!state.roms.length) {
    const empty = document.createElement("p");
    empty.className = "message";
    empty.textContent = "Sin juegos detectados.";
    els.romGrid.appendChild(empty);
    return;
  }

  for (const rom of state.roms) {
    const button = document.createElement("button");
    button.type = "button";
    button.className = `rom-card${rom.file === state.selected ? " is-selected" : ""}`;
    button.innerHTML = `<strong></strong><small></small>`;
    button.querySelector("strong").textContent = rom.name;
    button.querySelector("small").textContent = `${formatBytes(rom.size)} - ${rom.file}`;
    button.addEventListener("click", () => {
      selectRom(rom.file);
      startSelectedGame();
    });
    els.romGrid.appendChild(button);
  }
}

function configureEmulator(rom) {
  window.EJS_player = "#game";
  window.EJS_core = state.core || "snes";
  window.EJS_gameUrl = appPath(rom.url);
  window.EJS_gameName = rom.name;
  window.EJS_pathtodata = appPath(state.dataPath);
  window.EJS_startOnLoaded = true;
  window.EJS_AdUrl = "";
  window.EJS_color = "#38bdf8";
  window.EJS_backgroundColor = "#000000";
  window.EJS_fullscreenOnLoaded = false;
  window.EJS_volume = 0.65;
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
  window.history.replaceState({}, "", url.toString());

  if (state.loaded) {
    window.location.href = url.toString();
    return;
  }

  try {
    setMessage(`Cargando ${rom.name}...`);
    resetPlayerDOM();
    configureEmulator(rom);
    await injectLoader();
    state.loaded = true;
    els.currentTitle.textContent = rom.name;
    setMessage("Juego cargado. En movil, usa el boton tactil del emulador para controles y menu.");
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
  els.romSelect.addEventListener("change", () => selectRom(els.romSelect.value));
  els.playButton.addEventListener("click", startSelectedGame);
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
  await loadROMs();
  const requested = getRomFromUrl();
  if (requested && findROM(requested)) {
    await startSelectedGame();
  }
}

bootstrap().catch((err) => setMessage(err.message || "Error inicializando Juegos.", true));
