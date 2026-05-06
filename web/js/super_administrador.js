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

  function setAdminNavGroupOpen(group, open) {
    if (!group) return;
    if (open && group.parentElement) {
      var siblings = Array.from(group.parentElement.querySelectorAll(".admin-nav-group"));
      siblings.forEach(function (other) {
        if (other !== group) {
          other.classList.remove("is-open");
          var otherTitle = other.querySelector(".admin-nav-group-title");
          if (otherTitle) otherTitle.setAttribute("aria-expanded", "false");
        }
      });
    }
    group.classList.toggle("is-open", !!open);
    var title = group.querySelector(".admin-nav-group-title");
    if (title) title.setAttribute("aria-expanded", open ? "true" : "false");
  }

  function openMenuGroupForLink(link) {
    if (!link || typeof link.closest !== "function") return;
    var group = link.closest(".admin-nav-group");
    if (!group) return;
    setAdminNavGroupOpen(group, true);
  }

  function setupAdminNavGroups() {
    var groups = Array.from(document.querySelectorAll(".admin-sidebar .admin-nav-group"));
    groups.forEach(function (group, index) {
      var title = group.querySelector(".admin-nav-group-title");
      if (!title) return;
      if (title.tagName && title.tagName.toLowerCase() !== "button") {
        title.setAttribute("role", "button");
        title.setAttribute("tabindex", "0");
      }
      var defaultOpen = group.classList.contains("is-open") || index === 0;
      setAdminNavGroupOpen(group, defaultOpen);
      var toggle = function () {
        setAdminNavGroupOpen(group, !group.classList.contains("is-open"));
      };
      title.addEventListener("click", toggle);
      title.addEventListener("keydown", function (event) {
        if (event.key === "Enter" || event.key === " ") {
          event.preventDefault();
          toggle();
        }
      });
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
    if (found) {
      found.classList.add("active");
      openMenuGroupForLink(found);
    }
  }

  function applySuperRoleNavigation(role) {
    role = String(role || "").trim().toLowerCase();
    if (role !== "control_super_administrador") {
      if (role && role !== "super_administrador") {
        window.location.href = "/seleccionar_empresa.html";
      }
      return;
    }
    var allowed = {
      "/super/licencias_resumen.html": true,
      "/super/administradores.html": true,
      "/super/seguridad.html": true,
      "/super/errores.html": true,
      "/super/reportes_globales.html": true
    };
    links.forEach(function (a) {
      var normalized = normalizeHref(a.getAttribute("href")).split("?")[0];
      var visible = !!allowed[normalized] || a.classList.contains("select-company");
      var item = a.closest ? a.closest("li") : null;
      if (item) item.hidden = !visible;
    });
    document.querySelectorAll(".admin-nav-group").forEach(function (group) {
      var visibleLinks = group.querySelectorAll("li:not([hidden]) a").length;
      group.hidden = visibleLinks === 0;
      if (group.hidden) {
        setAdminNavGroupOpen(group, false);
      }
    });
    var current = normalizeHref(iframe ? iframe.getAttribute("src") : "");
    if (!allowed[current.split("?")[0]] && iframe) {
      iframe.setAttribute("src", "/super/licencias_resumen.html");
      persistLastPage("/super/licencias_resumen.html");
      setActiveByHref("/super/licencias_resumen.html");
    } else {
      setActiveByHref(current || "/super/licencias_resumen.html");
    }
  }

  setupAdminNavGroups();

  links.forEach(function (a) {
    a.addEventListener("click", function (e) {
      var targetAttr = a.getAttribute("target");
      if (targetAttr === "_blank" || a.classList.contains("select-company")) {
        return;
      }

      e.preventDefault();
      clearActive();
      this.classList.add("active");
      openMenuGroupForLink(this);

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

  fetch("/me", { credentials: "same-origin" })
    .then(function (res) {
      if (!res.ok) throw new Error("no-auth");
      return res.json();
    })
    .then(function (admin) {
      applySuperRoleNavigation(admin && admin.role);
    })
    .catch(function () {
      window.location.href = "/login.html";
    });
})();
