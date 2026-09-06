(function (global, document) {
  "use strict";

  global.__pcsInstallPromptEvent = null;
  global.addEventListener("beforeinstallprompt", function (event) {
    event.preventDefault();
    global.__pcsInstallPromptEvent = event;
    global.dispatchEvent(new CustomEvent("pcs:beforeinstallprompt"));
  });

  if (!document.documentElement.hasAttribute("data-login-theme-bootstrap")) return;

  var allowedThemes = {
    light: true,
    "light-rose": true,
    "light-gold": true,
    "light-wood": true,
    dark: true,
    "dark-violet": true,
    "dark-emerald": true,
    "dark-corporate": true,
    "dark-absolute": true,
    "dark-obsidian": true,
    "dark-neon": true
  };

  function cookieTheme() {
    var match = String(document.cookie || "").match(/(?:^|;\s*)pcs_theme=([^;]+)/);
    return match ? decodeURIComponent(match[1] || "") : "";
  }

  function normalize(theme) {
    var value = String(theme || "").trim().toLowerCase();
    if (value === "dark-protect") value = "dark";
    return allowedThemes[value] ? value : "light";
  }

  function storedTheme() {
    var value = cookieTheme();
    try {
      value = value || global.localStorage.getItem("theme") || "";
    } catch (_) {}
    return normalize(value);
  }

  function apply(theme) {
    var normalized = normalize(theme);
    var root = document.documentElement;
    root.setAttribute("data-theme", normalized);
    root.classList.toggle("theme-light", normalized.indexOf("light") === 0);
    root.classList.toggle("theme-dark", normalized.indexOf("light") !== 0);
    return normalized;
  }

  global.__pcsLoginTheme = { normalize: normalize, apply: apply, current: storedTheme };
  apply(storedTheme());
})(window, document);
