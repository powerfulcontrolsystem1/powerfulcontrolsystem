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

  var KEY_META = {
    13: { key: "Enter", code: "Enter" },
    27: { key: "Escape", code: "Escape" },
    32: { key: " ", code: "Space" },
    37: { key: "ArrowLeft", code: "ArrowLeft" },
    38: { key: "ArrowUp", code: "ArrowUp" },
    39: { key: "ArrowRight", code: "ArrowRight" },
    40: { key: "ArrowDown", code: "ArrowDown" },
    48: { key: "0", code: "Digit0" },
    49: { key: "1", code: "Digit1" },
    50: { key: "2", code: "Digit2" },
    65: { key: "a", code: "KeyA" },
    70: { key: "f", code: "KeyF" },
    76: { key: "l", code: "KeyL" },
    80: { key: "p", code: "KeyP" },
    81: { key: "q", code: "KeyQ" },
    82: { key: "r", code: "KeyR" },
    83: { key: "s", code: "KeyS" }
  };

  var frame = document.querySelector("[data-arcade-frame]");
  if (!frame) {
    return;
  }
  var gameSlug = (window.PCSJuegos && window.PCSJuegos.slugFromPath) ? window.PCSJuegos.slugFromPath() : "juego";
  var audioCtx = null;

  function normalizeTheme(theme) {
    var value = String(theme || "").trim().toLowerCase();
    if (value === "dark-protect") value = "dark";
    return THEME_VALUES[value] ? value : "light";
  }

  function currentTheme() {
    try {
      if (window.__pcsThemeManager) {
        return normalizeTheme(window.__pcsThemeManager.getTheme());
      }
    } catch (error) {}
    try {
      return normalizeTheme(window.localStorage.getItem("theme") || "");
    } catch (error) {
      return "light";
    }
  }

  function injectSourceTheme() {
    var doc;
    try {
      doc = frame.contentDocument;
    } catch (error) {
      return;
    }
    if (!doc || !doc.documentElement) return;
    var theme = currentTheme();
    doc.documentElement.setAttribute("data-theme", theme);
    try {
      frame.contentWindow.postMessage({ type: "pcs-theme", theme: theme }, window.location.origin);
    } catch (error) {}
    if (!doc.getElementById("pcsOpenGameSourceTheme")) {
      var link = doc.createElement("link");
      link.id = "pcsOpenGameSourceTheme";
      link.rel = "stylesheet";
      link.href = "/Juegos/open_game_source.css";
      (doc.head || doc.documentElement).appendChild(link);
    }
    if (!doc.getElementById("pcsOpenGameEmbedRuntime") && !doc.querySelector('script[src*="/Juegos/open_game_embed.js"]')) {
      var script = doc.createElement("script");
      script.id = "pcsOpenGameEmbedRuntime";
      script.src = "/Juegos/open_game_embed.js";
      (doc.body || doc.documentElement).appendChild(script);
    }
  }

  function resizeFrameToViewport() {
    var controls = document.querySelector(".arcade-controls");
    var viewportHeight = window.innerHeight || document.documentElement.clientHeight || 720;
    var controlsHeight = controls ? controls.getBoundingClientRect().height : 0;
    var isMobile = window.matchMedia && window.matchMedia("(max-width: 720px)").matches;
    var configuredAspect = Number(frame.getAttribute("data-arcade-aspect"));
    var aspect = Number.isFinite(configuredAspect) && configuredAspect > 0 ? configuredAspect : (isMobile ? 1.05 : 0.62);
    var minHeight = isMobile ? 300 : 360;
    var frameRect = frame.getBoundingClientRect();
    var available = Math.max(260, viewportHeight - frameRect.top - controlsHeight - 10);
    var maxByWidth = Math.max(minHeight, Math.round(frameRect.width * aspect));
    var maxViewport = isMobile ? Math.max(260, viewportHeight - 98) : Math.min(620, Math.max(360, viewportHeight - 106));
    var next = Math.min(available, maxByWidth, maxViewport);
    frame.style.setProperty("--arcade-frame-height", Math.max(Math.min(minHeight, available), Math.round(next)) + "px");
  }

  function makeEvent(win, type, keyCode) {
    var meta = KEY_META[keyCode] || { key: String.fromCharCode(keyCode), code: "" };
    var event = new win.KeyboardEvent(type, {
      key: meta.key,
      code: meta.code,
      bubbles: true,
      cancelable: true,
      which: keyCode,
      keyCode: keyCode
    });
    Object.defineProperty(event, "keyCode", { get: function () { return keyCode; } });
    Object.defineProperty(event, "which", { get: function () { return keyCode; } });
    return event;
  }

  function sendKey(keyCode, type) {
    var win = frame.contentWindow;
    if (!win || !win.document) {
      return;
    }
    try {
      win.focus();
      frame.focus();
      win.dispatchEvent(makeEvent(win, type, keyCode));
      win.document.dispatchEvent(makeEvent(win, type, keyCode));
      if (win.document.body) {
        win.document.body.dispatchEvent(makeEvent(win, type, keyCode));
      }
    } catch (error) {
      console.warn("No se pudo enviar control arcade", error);
    }
  }

  function press(keyCode, hold) {
    arcadeTone(hold ? 420 : 620, 0.055);
    sendKey(keyCode, "keydown");
    sendKey(keyCode, "keypress");
    if (!hold) {
      window.setTimeout(function () {
        sendKey(keyCode, "keyup");
      }, 80);
    }
  }

  function arcadeTone(freq, duration) {
    var Ctx = window.AudioContext || window.webkitAudioContext;
    if (!Ctx) return;
    if (!audioCtx) audioCtx = new Ctx();
    if (audioCtx.state === "suspended") {
      audioCtx.resume().catch(function () {});
    }
    var osc = audioCtx.createOscillator();
    var gain = audioCtx.createGain();
    osc.type = "square";
    osc.frequency.value = freq || 520;
    gain.gain.setValueAtTime(0.0001, audioCtx.currentTime);
    gain.gain.exponentialRampToValueAtTime(0.035, audioCtx.currentTime + 0.01);
    gain.gain.exponentialRampToValueAtTime(0.0001, audioCtx.currentTime + (duration || 0.06));
    osc.connect(gain);
    gain.connect(audioCtx.destination);
    osc.start();
    osc.stop(audioCtx.currentTime + (duration || 0.06) + 0.02);
  }

  document.querySelectorAll("[data-key]").forEach(function (button) {
    var keyCode = Number(button.dataset.key);
    var hold = button.dataset.mode !== "tap";

    button.addEventListener("pointerdown", function (event) {
      event.preventDefault();
      press(keyCode, hold);
      if (navigator.vibrate) {
        navigator.vibrate(14);
      }
    });

    ["pointerup", "pointercancel", "pointerleave"].forEach(function (type) {
      button.addEventListener(type, function () {
        if (hold) {
          sendKey(keyCode, "keyup");
        }
      });
    });

    button.addEventListener("click", function (event) {
      event.preventDefault();
    });
  });

  frame.addEventListener("load", function () {
    injectSourceTheme();
    resizeFrameToViewport();
    window.setTimeout(resizeFrameToViewport, 120);
    frame.focus();
  });

  if (window.PCSJuegos && typeof window.PCSJuegos.enhanceWrapper === "function") {
    window.PCSJuegos.enhanceWrapper({
      frame: frame,
      juego: gameSlug,
      title: frame.getAttribute("title") || document.title || gameSlug
    });
  }

  window.addEventListener("resize", resizeFrameToViewport);
  window.addEventListener("orientationchange", function () {
    window.setTimeout(resizeFrameToViewport, 180);
  });
  window.addEventListener("pcs:theme-changed", function (event) {
    injectSourceTheme();
    try {
      frame.contentWindow.postMessage({ type: "pcs-theme", theme: event.detail && event.detail.theme }, window.location.origin);
    } catch (error) {}
  });

  injectSourceTheme();
  resizeFrameToViewport();
  window.setTimeout(resizeFrameToViewport, 120);
}());
