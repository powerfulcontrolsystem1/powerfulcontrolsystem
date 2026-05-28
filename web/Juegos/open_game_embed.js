(function () {
  var THEME_VALUES = {
    dark: true,
    "dark-violet": true,
    "dark-emerald": true,
    "dark-corporate": true,
    "dark-neon": true,
    light: true,
    "light-rose": true,
    "light-gold": true,
    "light-wood": true
  };

  function normalizeTheme(theme) {
    var value = String(theme || "").trim().toLowerCase();
    if (value === "dark-protect") value = "dark";
    return THEME_VALUES[value] ? value : "dark";
  }

  function readCookie(name) {
    var match = String(document.cookie || "").match("(^|;)\\s*" + name + "\\s*=\\s*([^;]+)");
    return match ? decodeURIComponent(match.pop()) : "";
  }

  function resolveTheme() {
    try {
      if (window.parent && window.parent !== window && window.parent.__pcsThemeManager) {
        return normalizeTheme(window.parent.__pcsThemeManager.getTheme());
      }
    } catch (error) {}
    try {
      return normalizeTheme(window.localStorage.getItem("theme") || readCookie("pcs_theme") || "");
    } catch (error) {
      return normalizeTheme(readCookie("pcs_theme") || "");
    }
  }

  function applyTheme(theme) {
    if (!document.documentElement) return;
    document.documentElement.setAttribute("data-theme", normalizeTheme(theme));
  }

  var audioCtx = null;
  var lastAudioAt = 0;

  function audioContext() {
    var Ctx = window.AudioContext || window.webkitAudioContext;
    if (!Ctx) return null;
    if (!audioCtx) audioCtx = new Ctx();
    if (audioCtx.state === "suspended") {
      audioCtx.resume().catch(function () {});
    }
    return audioCtx;
  }

  function blip(freq, duration, gainValue) {
    var nowMs = Date.now();
    if (nowMs - lastAudioAt < 55) return;
    lastAudioAt = nowMs;
    var ctx = audioContext();
    if (!ctx) return;
    var osc = ctx.createOscillator();
    var gain = ctx.createGain();
    osc.type = "square";
    osc.frequency.value = freq || 520;
    gain.gain.setValueAtTime(0.0001, ctx.currentTime);
    gain.gain.exponentialRampToValueAtTime(gainValue || 0.032, ctx.currentTime + 0.01);
    gain.gain.exponentialRampToValueAtTime(0.0001, ctx.currentTime + (duration || 0.08));
    osc.connect(gain);
    gain.connect(ctx.destination);
    osc.start();
    osc.stop(ctx.currentTime + (duration || 0.08) + 0.02);
  }

  function parseScore(text) {
    var match = String(text || "").replace(/[^\d-]+/g, " ").trim().match(/-?\d+/g);
    if (!match || !match.length) return 0;
    var value = parseInt(match[match.length - 1], 10);
    return Number.isFinite(value) && value > 0 ? value : 0;
  }

  function readScore() {
    var selectors = ["[data-score]", "#score", "#scoreBoard", "#finalScore", ".score", "[id*='score' i]", "[class*='score' i]"];
    for (var i = 0; i < selectors.length; i += 1) {
      var node = null;
      try {
        node = document.querySelector(selectors[i]);
      } catch (error) {}
      if (!node) continue;
      var score = parseScore((node.getAttribute && node.getAttribute("data-score")) || node.textContent);
      if (score > 0) return score;
    }
    return parseScore(window.PCSGameScore || window.PCSTetrisScore || window.PCSPongScore || window.PCSPacmanScore || 0);
  }

  function readLevel() {
    var node = null;
    try {
      node = document.querySelector("[data-level], #level, [id*='level' i]");
    } catch (error) {}
    return Math.max(1, parseScore((node && ((node.getAttribute && node.getAttribute("data-level")) || node.textContent)) || "1") || 1);
  }

  function slugFromPath() {
    var parts = String(window.location.pathname || "").split("/").filter(Boolean);
    var sourceIndex = parts.indexOf("source");
    if (sourceIndex > 0) return parts[sourceIndex - 1];
    var juegosIndex = parts.indexOf("Juegos");
    if (juegosIndex >= 0 && parts[juegosIndex + 1]) return parts[juegosIndex + 1].replace(/\.html$/i, "");
    return "juego";
  }

  function installScoreReporter() {
    var last = 0;
    function report(final) {
      var score = readScore();
      if (!score || score < last) return;
      if (score > last) blip(760, 0.06, 0.022);
      last = score;
      try {
        if (window.parent && window.parent !== window) {
          window.parent.postMessage({
            type: "pcs-game-score",
            juego: slugFromPath(),
            puntaje: score,
            nivel: readLevel(),
            final: Boolean(final)
          }, window.location.origin);
        }
      } catch (error) {}
    }
    window.setInterval(function () { report(false); }, 1200);
    window.addEventListener("beforeunload", function () { report(true); });
    document.addEventListener("visibilitychange", function () {
      if (document.hidden) report(true);
    });
  }

  applyTheme(resolveTheme());

  if (window.top !== window.self) {
    document.documentElement.classList.add("is-embedded");
  }

  window.addEventListener("message", function (event) {
    var data = event && event.data;
    if (data && data.type === "pcs-theme") {
      applyTheme(data.theme);
    }
  });

  window.addEventListener("storage", function (event) {
    if (event && event.key === "theme") {
      applyTheme(event.newValue);
    }
  });

  ["pointerdown", "keydown"].forEach(function (type) {
    window.addEventListener(type, function () {
      blip(type === "keydown" ? 420 : 560, 0.045, 0.018);
    }, { passive: true });
  });

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", installScoreReporter);
  } else {
    installScoreReporter();
  }
}());
