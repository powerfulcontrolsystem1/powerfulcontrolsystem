(function () {
  var sidebar = document.querySelector(".admin-sidebar .nav");
  var links = sidebar ? Array.from(sidebar.querySelectorAll("a")) : [];
  var iframe = document.getElementById("contentFrame");
  var storage = null;
  var lastPageKey = "super_admin:last_page";

  try {
    storage = window.sessionStorage;
  } catch (e) {
    storage = null;
  }

  function normalizeHref(href) {
    var raw = String(href || "").trim();
    if (!raw) return "";
    try {
      var u = new URL(raw, window.location.origin);
      return u.pathname + u.search;
    } catch (e) {
      return "";
    }
  }

  function isAllowedSuperHref(href) {
    return normalizeHref(href).indexOf("/super/") === 0;
  }

  function persistLastPage(href) {
    if (!storage) return;
    var normalized = normalizeHref(href);
    if (!isAllowedSuperHref(normalized)) return;
    try {
      storage.setItem(lastPageKey, normalized);
    } catch (e) {}
  }

  function restoreLastPage(defaultHref) {
    var fallback = normalizeHref(defaultHref) || "/super/licencias_resumen.html";
    if (!storage) return fallback;
    try {
      var raw = storage.getItem(lastPageKey) || "";
      var normalized = normalizeHref(raw);
      if (!isAllowedSuperHref(normalized)) return fallback;
      return normalized;
    } catch (e) {
      return fallback;
    }
  }

  function clearActive() {
    links.forEach(function (a) {
      a.classList.remove("active");
    });
  }

  function setActiveByHref(href) {
    var current = normalizeHref(href);
    var currentPath = current.split("?")[0];
    clearActive();
    var found = links.find(function (a) {
      var linkHref = normalizeHref(a.getAttribute("href"));
      if (!linkHref) return false;
      if (linkHref === current) return true;
      return linkHref.split("?")[0] === currentPath;
    });
    if (found) found.classList.add("active");
  }

  links.forEach(function (a) {
    a.addEventListener("click", function (e) {
      var targetAttr = a.getAttribute("target");
      if (targetAttr === "_blank" || a.classList.contains("select-company")) {
        return;
      }

      e.preventDefault();
      clearActive();
      this.classList.add("active");

      var href = a.getAttribute("href");
      if (!href) return;

      if (iframe) {
        iframe.setAttribute("src", href);
        persistLastPage(href);
      } else {
        window.location.href = href;
      }
    });
  });

  if (iframe) {
    var defaultIframeSrc = iframe.getAttribute("src") || "/super/licencias_resumen.html";
    var initialIframeSrc = restoreLastPage(defaultIframeSrc);
    iframe.setAttribute("src", initialIframeSrc);
    setActiveByHref(initialIframeSrc);
  }

  if (iframe) {
    iframe.addEventListener("load", function () {
      try {
        var src = iframe.contentWindow.location.pathname + iframe.contentWindow.location.search;
        persistLastPage(src);
        setActiveByHref(src);
      } catch (e) {
        var src2 = iframe.getAttribute("src");
        persistLastPage(src2);
        setActiveByHref(src2);
      }
    });
  }
})();

(function () {
  try {
    if (localStorage.getItem("rememberAccount") === "1") {
      fetch("/me")
        .then(function (res) {
          if (!res.ok) throw new Error("unauth");
          return res.json();
        })
        .then(function (admin) {
          if (admin && admin.email) {
            try {
              localStorage.setItem("rememberedEmail", admin.email);
            } catch (e) {}
          }
        })
        .catch(function () {});
    }
  } catch (e) {}
})();
