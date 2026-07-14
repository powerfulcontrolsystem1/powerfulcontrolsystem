// csrf_fetch.js centraliza el token synchronizer para mutaciones same-origin.
// No modifica llamadas bearer, cross-origin ni solicitudes que ya declaran token.
(function () {
  "use strict";

  if (!window.fetch || window.__pcsCSRFFetchInstalled) return;
  window.__pcsCSRFFetchInstalled = true;

  function readCookie(name) {
    var prefix = String(name || "") + "=";
    var parts = String(document.cookie || "").split(";");
    for (var i = 0; i < parts.length; i++) {
      var part = parts[i].trim();
      if (part.indexOf(prefix) === 0) {
        try { return decodeURIComponent(part.slice(prefix.length)); } catch (error) { return part.slice(prefix.length); }
      }
    }
    return "";
  }

  function isMutation(method) {
    method = String(method || "GET").toUpperCase();
    return method === "POST" || method === "PUT" || method === "PATCH" || method === "DELETE";
  }

  var originalFetch = window.fetch.bind(window);
  window.fetch = function (input, init) {
    var options = init ? Object.assign({}, init) : {};
    var method = options.method || (input && input.method) || "GET";
    if (!isMutation(method) || options.credentials === "omit") {
      return originalFetch(input, options);
    }

    var rawURL = typeof input === "string" || input instanceof URL ? String(input) : (input && input.url);
    var target;
    try { target = new URL(rawURL || window.location.href, window.location.href); } catch (error) { return originalFetch(input, options); }
    if (target.origin !== window.location.origin) {
      return originalFetch(input, options);
    }

    var headers = new Headers(options.headers || (input && input.headers) || undefined);
    if (headers.has("Authorization") || headers.has("X-CSRF-Token")) {
      return originalFetch(input, options);
    }
    var token = readCookie("pcs_csrf");
    if (token) {
      headers.set("X-CSRF-Token", token);
      options.headers = headers;
    }
    return originalFetch(input, options);
  };
})();
