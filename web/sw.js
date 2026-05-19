const PCS_CACHE = "pcs-shell-v1";
const PCS_SHELL = [
  "/login.html",
  "/estilos.css",
  "/img/logo.png",
  "/img/pwa-icon-192.png",
  "/img/pwa-icon-512.png"
];

self.addEventListener("install", function (event) {
  event.waitUntil(
    caches.open(PCS_CACHE)
      .then(function (cache) { return cache.addAll(PCS_SHELL); })
      .catch(function () {})
      .then(function () { return self.skipWaiting(); })
  );
});

self.addEventListener("activate", function (event) {
  event.waitUntil(
    caches.keys()
      .then(function (keys) {
        return Promise.all(keys.filter(function (key) { return key !== PCS_CACHE; }).map(function (key) { return caches.delete(key); }));
      })
      .then(function () { return self.clients.claim(); })
  );
});

self.addEventListener("fetch", function (event) {
  var request = event.request;
  if (!request || request.method !== "GET") {
    return;
  }

  var url = new URL(request.url);
  if (url.origin !== self.location.origin) {
    return;
  }
  if (url.pathname.indexOf("/api/") === 0 || url.pathname.indexOf("/super/api/") === 0 || url.pathname.indexOf("/auth/") === 0) {
    return;
  }

  if (request.mode === "navigate") {
    event.respondWith(fetch(request).catch(function () { return caches.match("/login.html"); }));
    return;
  }

  event.respondWith(
    caches.match(request).then(function (cached) {
      if (cached) {
        return cached;
      }
      return fetch(request).then(function (response) {
        var cacheable = response && response.ok && /\.(?:css|js|png|jpg|jpeg|svg|webp|ico|webmanifest)$/i.test(url.pathname);
        if (cacheable) {
          var copy = response.clone();
          caches.open(PCS_CACHE).then(function (cache) { cache.put(request, copy); }).catch(function () {});
        }
        return response;
      });
    })
  );
});
