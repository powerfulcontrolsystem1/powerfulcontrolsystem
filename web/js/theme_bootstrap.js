(function () {
  "use strict";

  var storageKey = "theme";
  var cookieName = "pcs_theme";
  var themes = {
    dark: true,
    "dark-violet": true,
    "dark-emerald": true,
    "dark-corporate": true,
    "dark-absolute": true,
    "dark-obsidian": true,
    "dark-neon": true,
    light: true,
    "light-rose": true,
    "light-gold": true,
    "light-wood": true
  };

  function normalize(value) {
    var theme = String(value || "").trim().toLowerCase();
    return themes[theme] ? theme : "light";
  }

  function cookieTheme() {
    var prefix = cookieName + "=";
    var parts = String(document.cookie || "").split(";");
    for (var i = 0; i < parts.length; i += 1) {
      var part = parts[i].trim();
      if (part.indexOf(prefix) === 0) {
        return decodeURIComponent(part.slice(prefix.length));
      }
    }
    return "";
  }

  function storedTheme() {
    try {
      return localStorage.getItem(storageKey) || "";
    } catch (_) {
      return "";
    }
  }

  function apply(value) {
    var theme = normalize(value);
    var root = document.documentElement;
    var light = theme.indexOf("light") === 0;
    root.setAttribute("data-theme", theme);
    root.setAttribute("data-appearance-mode", light ? "light" : "dark");
    root.classList.toggle("theme-light", light);
    root.classList.toggle("theme-dark", !light);
    root.style.colorScheme = light ? "light" : "dark";
    return theme;
  }

  var current = apply(cookieTheme() || storedTheme());
  window.PCSThemeBootstrap = { apply: apply, current: current, normalize: normalize };

  window.addEventListener("storage", function (event) {
    if (event.key === storageKey && event.newValue) apply(event.newValue);
  });
  window.addEventListener("pcs:theme-changed", function (event) {
    if (event.detail && event.detail.theme) apply(event.detail.theme);
  });
})();
