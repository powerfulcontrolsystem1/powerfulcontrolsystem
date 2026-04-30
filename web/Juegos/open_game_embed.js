(function () {
  var THEME_VALUES = {
    dark: true,
    "dark-violet": true,
    "dark-emerald": true,
    light: true,
    "light-rose": true,
    "light-gold": true
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
}());
