(function () {
  "use strict";

  var DEFAULT_VERSION = "20260508-empresa-submenus";

  function currentVersion() {
    var script = document.currentScript;
    return (script && script.getAttribute("data-version")) || DEFAULT_VERSION;
  }

  function resolveEmpresaId() {
    try {
      if (typeof window.__resolveEmpresaIdContext === "function") {
        var fromContext = Number(window.__resolveEmpresaIdContext() || 0);
        if (Number.isFinite(fromContext) && fromContext > 0) return Math.trunc(fromContext);
      }
    } catch (_) {}
    try {
      var params = new URLSearchParams(window.location.search || "");
      var own = Number(params.get("empresa_id") || params.get("id") || 0);
      if (Number.isFinite(own) && own > 0) return Math.trunc(own);
    } catch (_) {}
    return 0;
  }

  function withContext(rawUrl) {
    try {
      var url = new URL(rawUrl, window.location.origin);
      var empresaId = resolveEmpresaId();
      if (empresaId > 0) url.searchParams.set("empresa_id", String(empresaId));
      url.searchParams.set("submenu", "1");
      if (!url.searchParams.get("v")) url.searchParams.set("v", currentVersion());
      return url.pathname + url.search + url.hash;
    } catch (_) {
      return rawUrl;
    }
  }

  function markActive(link) {
    document.querySelectorAll("[data-submenu-link]").forEach(function (item) {
      item.classList.toggle("active", item === link);
    });
  }

  function boot() {
    var links = Array.prototype.slice.call(document.querySelectorAll("[data-submenu-link]"));
    links.forEach(function (link, index) {
      link.setAttribute("href", withContext(link.getAttribute("href") || ""));
      link.addEventListener("click", function () { markActive(link); });
      if (index === 0) markActive(link);
    });

    document.querySelectorAll("iframe[data-submenu-frame]").forEach(function (frame) {
      var src = frame.getAttribute("src") || frame.getAttribute("data-src") || "";
      frame.setAttribute("src", withContext(src));
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", boot);
  } else {
    boot();
  }
})();
