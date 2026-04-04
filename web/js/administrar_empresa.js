function getQueryParam(name) {
  var params = new URLSearchParams(window.location.search);
  return params.get(name);
}

(function () {
  var id = getQueryParam("id");
  var title = document.getElementById("empresaTitle");
  var frame = document.getElementById("contentFrame");
  var storage = null;
  try {
    storage = window.sessionStorage;
  } catch (e) {
    storage = null;
  }
  var links = [
    document.getElementById("linkInicio"),
    document.getElementById("linkVentas"),
    document.getElementById("linkCarritoCompras"),
    document.getElementById("linkProductos"),
    document.getElementById("linkConfiguracion"),
    document.getElementById("linkUsuarios"),
    document.getElementById("linkChatTareas"),
    document.getElementById("linkClientes"),
    document.getElementById("linkConfigAvanzada"),
    document.getElementById("linkFacturacionElectronica"),
    document.getElementById("linkFinanzas"),
    document.getElementById("linkUbicacionGPS"),
    document.getElementById("linkConfigEstaciones"),
    document.getElementById("linkEstaciones"),
    document.getElementById("linkReportes"),
  ];

  function storageKey(empresaId) {
    return "admin_empresa:last_page:" + String(empresaId || "global");
  }

  function normalizeHref(href) {
    var raw = String(href || "").trim();
    if (!raw) return "";
    try {
      var url = new URL(raw, window.location.origin);
      return url.pathname + url.search;
    } catch (e) {
      return "";
    }
  }

  function isAllowedFrameHref(href) {
    var normalized = normalizeHref(href);
    return normalized.indexOf("/administrar_empresa/") === 0;
  }

  function defaultFrameSrc(empresaId) {
    var base = new URL("/administrar_empresa/inicio.html", window.location.origin);
    if (empresaId) {
      base.searchParams.set("empresa_id", empresaId);
    }
    return base.pathname + base.search;
  }

  function withEmpresaParam(href, empresaId) {
    var normalized = normalizeHref(href);
    if (!normalized) return "";
    try {
      var url = new URL(normalized, window.location.origin);
      if (empresaId) {
        url.searchParams.set("empresa_id", empresaId);
      }
      return url.pathname + url.search;
    } catch (e) {
      return "";
    }
  }

  function persistFrameSrc(href, empresaId) {
    if (!storage) return;
    var normalized = withEmpresaParam(href, empresaId);
    if (!isAllowedFrameHref(normalized)) return;
    try {
      storage.setItem(storageKey(empresaId), normalized);
    } catch (e) {}
  }

  function getStoredFrameSrc(empresaId) {
    if (!storage) return "";
    try {
      var raw = storage.getItem(storageKey(empresaId)) || "";
      var normalized = withEmpresaParam(raw, empresaId);
      if (!isAllowedFrameHref(normalized)) return "";
      return normalized;
    } catch (e) {
      return "";
    }
  }

  function clearActive() {
    links.forEach(function (link) {
      if (!link) return;
      link.classList.remove("active");
    });
  }

  function setActiveByHref(href) {
    var current = normalizeHref(href).split("?")[0];
    clearActive();
    links.forEach(function (link) {
      if (!link) return;
      var linkHref = normalizeHref(link.getAttribute("href")).split("?")[0];
      if (linkHref && linkHref === current) {
        link.classList.add("active");
      }
    });
  }

  function setLinksWithEmpresa(empresaId) {
    links.forEach(function (link) {
      if (!link) return;
      var href = link.getAttribute("href");
      if (!href) return;
      var target = new URL(href, window.location.origin);
      if (empresaId) {
        target.searchParams.set("empresa_id", empresaId);
      }
      link.setAttribute("href", target.pathname + target.search);

      link.addEventListener("click", function (ev) {
        ev.preventDefault();
        var linkHref = link.getAttribute("href");
        if (!frame || !linkHref) {
          window.location.href = linkHref;
          return;
        }
        frame.setAttribute("src", linkHref);
        persistFrameSrc(linkHref, empresaId);
        setActiveByHref(linkHref);
      });
    });
  }

  if (frame) {
    frame.addEventListener("load", function () {
      var currentHref = "";
      try {
        currentHref = frame.contentWindow.location.pathname + frame.contentWindow.location.search;
      } catch (e) {
        currentHref = frame.getAttribute("src") || "";
      }
      if (!currentHref) return;
      persistFrameSrc(currentHref, id);
      setActiveByHref(currentHref);
    });
  }

  if (id) {
    setLinksWithEmpresa(id);
    if (frame) {
      var restored = getStoredFrameSrc(id);
      var initialSrc = restored || defaultFrameSrc(id);
      frame.src = initialSrc;
      setActiveByHref(initialSrc);
    }
    fetch("/super/api/empresas?id=" + encodeURIComponent(id), { credentials: "same-origin" })
      .then(function (resp) {
        if (!resp.ok) {
          title.textContent = "Administrar Empresa";
          throw new Error("empresa no encontrada");
        }
        return resp.json();
      })
      .then(function (data) {
        var nombre = data && (data.nombre || data.Nombre);
        if (nombre) {
          title.textContent = "Administrar Empresa - " + nombre;
          document.title = title.textContent;
        } else {
          title.textContent = "Administrar Empresa";
        }
      })
      .catch(function (err) {
        console.warn("No se pudo cargar empresa:", err);
        title.textContent = "Administrar Empresa";
      });
    return;
  }

  setLinksWithEmpresa("");
  if (frame) {
    var restoredGlobal = getStoredFrameSrc("");
    var initialGlobal = restoredGlobal || defaultFrameSrc("");
    frame.src = initialGlobal;
    setActiveByHref(initialGlobal);
  }
  title.textContent = "Administrar Empresa";
})();
