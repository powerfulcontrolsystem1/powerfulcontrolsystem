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
  var enableSoundBtn = document.getElementById("enableSoundBtn");
  var fullScreenBtn = document.getElementById("fullScreenBtn");
  var soundStatus = document.getElementById("soundStatus");
  var audioCtx = null;
  var soundEnabled = false;
  var lastCallingSignature = "";
  var didInitialRefresh = false;

  function render() {
    screenClock.textContent = new Date().toLocaleTimeString("es-CO", { hour: "2-digit", minute: "2-digit", second: "2-digit" });
  }

  function renderList(el, items, builder, emptyText) {
    el.innerHTML = items && items.length ? items.map(builder).join("") : '<div class="tv-ticket"><strong>' + esc(emptyText) + '</strong></div>';
  }

  function callingSignature(items) {
    return (items || []).map(function (x) {
      return [x.id, x.codigo_turno, x.puesto_nombre, x.estado, x.fecha_llamado, x.fecha_inicio_atencion].join(":");
    }).join("|");
  }

  function ensureAudio() {
    if (!audioCtx) {
      var Ctx = window.AudioContext || window.webkitAudioContext;
      if (Ctx) audioCtx = new Ctx();
    }
    if (audioCtx && audioCtx.state === "suspended") audioCtx.resume().catch(function () {});
    soundEnabled = !!audioCtx;
    if (enableSoundBtn) enableSoundBtn.classList.toggle("is-on", soundEnabled);
    if (soundStatus) soundStatus.textContent = soundEnabled ? "Alarmas activas" : "Alarmas visuales";
  }

  function beep() {
    if (!soundEnabled || !audioCtx) return;
    var now = audioCtx.currentTime;
    [0, .18, .36].forEach(function (offset) {
      var osc = audioCtx.createOscillator();
      var gain = audioCtx.createGain();
      osc.type = "sine";
      osc.frequency.setValueAtTime(880, now + offset);
      gain.gain.setValueAtTime(.0001, now + offset);
      gain.gain.exponentialRampToValueAtTime(.28, now + offset + .025);
      gain.gain.exponentialRampToValueAtTime(.0001, now + offset + .14);
      osc.connect(gain).connect(audioCtx.destination);
      osc.start(now + offset);
      osc.stop(now + offset + .16);
    });
  }

  function speakTurn(items) {
    if (!soundEnabled || !("speechSynthesis" in window) || !items || !items.length) return;
    var first = items[0];
    var text = "Turno " + (first.codigo_turno || "") + ". " + (first.puesto_nombre ? "Puesto " + first.puesto_nombre : "Acercarse al puesto indicado");
    try {
      window.speechSynthesis.cancel();
      window.speechSynthesis.speak(new SpeechSynthesisUtterance(text));
    } catch (_) {}
  }

  function alertTurn(items) {
    document.body.classList.remove("turn-alert");
    void document.body.offsetWidth;
    document.body.classList.add("turn-alert");
    beep();
    speakTurn(items);
  }

  async function refresh() {
    var data = await j(base);
    var callingItems = data.tickets_llamando || [];
    var signature = callingSignature(callingItems);
    if (didInitialRefresh && signature && signature !== lastCallingSignature) {
      alertTurn(callingItems);
    }
    lastCallingSignature = signature;
    didInitialRefresh = true;
    screenTitle.textContent = data.titulo || "Pantalla de turnos";
    renderList(callingNow, callingItems, function (x) {
      return '<div class="tv-ticket is-current"><div class="tv-code">' + esc(x.codigo_turno) + '</div><strong>' + esc(x.servicio_nombre) + '</strong><span class="tv-puesto">' + esc(x.puesto_nombre || "Por asignar") + '</span><span class="tv-meta">' + esc(x.estado || "llamando") + '</span></div>';
    }, "Sin tickets llamando en este momento.");
    renderList(recentCalls, data.llamados_recientes || [], function (x) {
      return '<div class="tv-ticket"><strong>' + esc(x.codigo_turno) + '</strong><span>' + esc(x.servicio_nombre) + ' · ' + esc(x.puesto_nombre || "Puesto") + ' · ' + esc(x.estado) + '</span></div>';
    }, "Sin llamados recientes.");
    renderList(serviceSummary, data.resumen_servicios || [], function (x) {
      return '<div class="tv-ticket"><strong>' + esc(x.etiqueta) + '</strong><span>Espera: ' + esc(x.en_espera) + ' · Atención: ' + esc(x.en_atencion) + '</span></div>';
    }, "Sin servicios activos.");
  }

  if (enableSoundBtn) {
    enableSoundBtn.addEventListener("click", ensureAudio);
  }
  if (fullScreenBtn) {
    fullScreenBtn.addEventListener("click", function () {
      var el = document.documentElement;
      if (!document.fullscreenElement && el.requestFullscreen) el.requestFullscreen().catch(function () {});
      if (document.fullscreenElement && document.exitFullscreen) document.exitFullscreen().catch(function () {});
    });
  }
  if (q("sound") === "1") {
    if (soundStatus) soundStatus.textContent = "Pulsa activar sonido para alarmas";
  }

  render();
  if (!empresaId) {
    screenTitle.textContent = "Pantalla de turnos";
    renderList(callingNow, [], function () { return ""; }, "Abre esta pantalla desde el enlace publico de una empresa.");
    renderList(recentCalls, [], function () { return ""; }, "Falta empresa_id en el enlace publico.");
    renderList(serviceSummary, [], function () { return ""; }, "Sin empresa asociada.");
    return;
  }
  refresh().catch(function () {});
  setInterval(render, 1000);
  setInterval(function () { refresh().catch(function () {}); }, 5000);
})();
