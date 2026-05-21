(function () {
  "use strict";

  var currentCache = "pcs-shell-v4";
  var reloadKey = "pcs_sw_cache_refresh_v4";

  function clearOldCaches() {
    if (!("caches" in window)) {
      return Promise.resolve(false);
    }
    return caches.keys().then(function (keys) {
      var stale = keys.filter(function (key) {
        return /^pcs-shell-/.test(key) && key !== currentCache;
      });
      return Promise.all(stale.map(function (key) {
        return caches.delete(key);
      })).then(function () {
        return stale.length > 0;
      });
    }).catch(function () {
      return false;
    });
  }

  function refreshOnceIfNeeded(cleared) {
    if (!cleared || !navigator.serviceWorker || !navigator.serviceWorker.controller) {
      return;
    }
    try {
      if (sessionStorage.getItem(reloadKey) === "1") {
        return;
      }
      sessionStorage.setItem(reloadKey, "1");
      window.location.reload();
    } catch (e) {}
  }

  if ("serviceWorker" in navigator) {
    navigator.serviceWorker.getRegistration("/").then(function (registration) {
      if (registration && typeof registration.update === "function") {
        registration.update().catch(function () {});
      }
    }).catch(function () {});
  }

  clearOldCaches().then(refreshOnceIfNeeded);
})();
