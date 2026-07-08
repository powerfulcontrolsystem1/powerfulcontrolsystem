(function () {
  var favoritesKey = "super_admin:favorites";
  var panelHref = "/super/licencias_resumen.html";
  var shellHref = "/super_administrador.html";

  function normalizeHref(href) {
    try {
      var url = new URL(href || window.location.href, window.location.origin);
      return url.pathname + url.search;
    } catch (e) {
      return "";
    }
  }

  function currentHref() {
    return normalizeHref(window.location.pathname + window.location.search);
  }

  function esc(value) {
    return String(value || "").replace(/[&<>"']/g, function (ch) {
      return ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[ch];
    });
  }

  function readFavorites() {
    try {
      var parsed = JSON.parse(window.localStorage.getItem(favoritesKey) || "[]");
      return Array.isArray(parsed) ? parsed : [];
    } catch (e) {
      return [];
    }
  }

  function writeFavorites(items) {
    try {
      window.localStorage.setItem(favoritesKey, JSON.stringify(items.slice(0, 24)));
    } catch (e) {}
  }

  function pageTitle() {
    var h1 = document.querySelector("h1");
    var title = h1 ? String(h1.textContent || "").trim() : "";
    if (title) return title;
    title = String(document.title || "").trim();
    if (title) return title.replace(/\s+-\s+Super Administrador$/i, "");
    return "Pagina de super administrador";
  }

  function isFavorite(href) {
    var normalized = normalizeHref(href);
    return readFavorites().some(function (item) {
      return normalizeHref(item && item.href) === normalized;
    });
  }

  function notifyShell() {
    try {
      window.dispatchEvent(new CustomEvent("pcs-super-favorites-changed"));
    } catch (e) {}
    try {
      if (window.parent && window.parent !== window) {
        window.parent.postMessage({ type: "pcs-super-favorites-changed" }, window.location.origin);
      }
    } catch (e) {}
  }

  function toggleFavorite(button) {
    var href = currentHref();
    if (!href || href === panelHref) return;
    var favorites = readFavorites();
    if (isFavorite(href)) {
      favorites = favorites.filter(function (item) {
        return normalizeHref(item && item.href) !== href;
      });
    } else {
      favorites.unshift({
        href: href,
        title: pageTitle(),
        icon: { type: "text", value: "*" },
        added_at: new Date().toISOString()
      });
    }
    writeFavorites(favorites);
    syncFavoriteButton(button);
    notifyShell();
  }

  function syncFavoriteButton(button) {
    if (!button) return;
    var active = isFavorite(currentHref());
    button.setAttribute("aria-pressed", active ? "true" : "false");
    button.title = active ? "Quitar de favoritos" : "Agregar a favoritos";
    button.textContent = active ? "Favorito" : "Favorito";
  }

  function goToPanel() {
    try {
      if (window.parent && window.parent !== window && window.parent.document) {
        var frame = window.parent.document.getElementById("contentFrame");
        if (frame) {
          frame.setAttribute("src", panelHref);
          return;
        }
      }
    } catch (e) {}
    window.location.href = shellHref;
  }

  function installStyles() {
    if (document.getElementById("superPageToolsStyles")) return;
    var style = document.createElement("style");
    style.id = "superPageToolsStyles";
    style.textContent = [
      ".super-page-tools{position:fixed;right:14px;bottom:14px;z-index:9998;display:flex;gap:8px;align-items:center;max-width:calc(100vw - 28px)}",
      ".super-page-tools button{min-height:36px;border:1px solid var(--border,#cbd5e1);border-radius:8px;background:var(--surface,#fff);color:var(--text,#111827);box-shadow:0 10px 24px rgba(15,23,42,.16);font:inherit;font-size:.82rem;font-weight:850;padding:8px 10px;cursor:pointer}",
      ".super-page-tools button:hover{background:var(--focus,#e0f2fe)}",
      ".super-page-tools button[aria-pressed='true']{background:color-mix(in srgb,var(--accent,#0f6fcb) 14%,var(--surface,#fff));color:var(--accent,#0f6fcb);border-color:color-mix(in srgb,var(--accent,#0f6fcb) 42%,var(--border,#cbd5e1))}",
      "@media(max-width:560px){.super-page-tools{right:8px;bottom:8px}.super-page-tools button{font-size:.76rem;padding:7px 8px}}"
    ].join("\n");
    document.head.appendChild(style);
  }

  function install() {
    if (!document.body || document.getElementById("superPageTools")) return;
    installStyles();
    var wrap = document.createElement("div");
    wrap.id = "superPageTools";
    wrap.className = "super-page-tools";
    wrap.setAttribute("aria-label", "Acciones de pagina super administrador");
    var fav = document.createElement("button");
    fav.type = "button";
    fav.setAttribute("aria-pressed", "false");
    fav.addEventListener("click", function () { toggleFavorite(fav); });
    syncFavoriteButton(fav);
    var panel = document.createElement("button");
    panel.type = "button";
    panel.textContent = "Panel super";
    panel.title = "Ir al panel de super administrador";
    panel.addEventListener("click", goToPanel);
    if (currentHref() !== panelHref) wrap.appendChild(fav);
    wrap.appendChild(panel);
    document.body.appendChild(wrap);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", install);
  } else {
    install();
  }
})();
