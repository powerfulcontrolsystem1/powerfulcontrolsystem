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

  function submenuKeyFromUrl(rawUrl) {
    try {
      var url = new URL(rawUrl, window.location.origin);
      var hash = String(url.hash || "");
      if (hash.indexOf("#tab-") === 0) return hash.slice(5);
    } catch (_) {}
    return "";
  }

  function hashTargetFromUrl(rawUrl) {
    try {
      var url = new URL(rawUrl, window.location.origin);
      return String(url.hash || "").replace(/^#/, "");
    } catch (_) {
      return "";
    }
  }

  function notifyFrameSelection(rawUrl) {
    var key = submenuKeyFromUrl(rawUrl);
    var hash = hashTargetFromUrl(rawUrl);
    if (!key && !hash) return;
    var frame = document.querySelector("iframe[data-submenu-frame]");
    if (!frame || !frame.contentWindow) return;
    window.setTimeout(function () {
      try {
        frame.contentWindow.postMessage({ type: "pcs-submenu-select", key: key, hash: hash }, window.location.origin);
      } catch (_) {}
    }, 80);
  }

  function boot() {
    var links = Array.prototype.slice.call(document.querySelectorAll("[data-submenu-link]"));
    links.forEach(function (link, index) {
      link.setAttribute("href", withContext(link.getAttribute("href") || ""));
      link.addEventListener("click", function () {
        markActive(link);
        notifyFrameSelection(link.getAttribute("href") || "");
      });
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
