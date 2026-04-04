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
    document.getElementById("linkAuditoria"),
    document.getElementById("linkChatTareas"),
    document.getElementById("linkClientes"),
    document.getElementById("linkConfigAvanzada"),
    document.getElementById("linkFacturacionElectronica"),
    document.getElementById("linkChatIA"),
    document.getElementById("linkFinanzas"),
    document.getElementById("linkUbicacionGPS"),
    document.getElementById("linkConfigEstaciones"),
    document.getElementById("linkEstaciones"),
    document.getElementById("linkReportes"),
  ];

  var permActionRead = "R";
  var permActionCreate = "C";
  var permActionUpdate = "U";
  var permActionApprove = "A";

  var permModuleVentas = "ventas";
  var permModuleInventario = "inventario";
  var permModuleFinanzas = "finanzas";
  var permModuleClientes = "clientes";
  var permModuleFacturacion = "facturacion";
  var permModuleSeguridad = "seguridad";

  var menuPermissionCatalog = {
    linkInicio: { alwaysVisible: true },
    linkVentas: { module: permModuleVentas, action: permActionRead },
    linkCarritoCompras: { module: permModuleVentas, action: permActionCreate },
    linkProductos: { module: permModuleInventario, action: permActionCreate },
    linkConfiguracion: { module: permModuleSeguridad, action: permActionUpdate },
    linkUsuarios: { module: permModuleSeguridad, action: permActionUpdate },
    linkAuditoria: { module: permModuleSeguridad, action: permActionRead },
    linkChatTareas: { module: permModuleVentas, action: permActionCreate },
    linkClientes: { module: permModuleClientes, action: permActionCreate },
    linkConfigAvanzada: { module: permModuleSeguridad, action: permActionUpdate },
    linkFacturacionElectronica: { module: permModuleFacturacion, action: permActionCreate },
    linkChatIA: { module: permModuleVentas, action: permActionRead },
    linkFinanzas: { module: permModuleFinanzas, action: permActionCreate },
    linkUbicacionGPS: { module: permModuleInventario, action: permActionCreate },
    linkConfigEstaciones: { module: permModuleVentas, action: permActionApprove },
    linkEstaciones: { module: permModuleVentas, action: permActionUpdate },
    linkReportes: { module: permModuleFinanzas, action: permActionRead },
  };

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

  function normalizePermissionRole(raw) {
    var value = String(raw || "").trim().toLowerCase();
    switch (value) {
      case "super_administrador":
      case "superadmin":
      case "super":
        return "super_administrador";
      case "administrador":
      case "admin":
      case "admin_empresa":
        return "admin_empresa";
      case "supervisor":
      case "supervisor_sucursal":
        return "supervisor_sucursal";
      case "cajero":
        return "cajero";
      case "inventario":
        return "inventario";
      case "compras":
        return "compras";
      case "contabilidad":
      case "contador":
        return "contabilidad";
      case "auditor":
        return "auditor";
      default:
        return value;
    }
  }

  function roleIn(role, allowedRoles) {
    var normalized = String(role || "").trim().toLowerCase();
    if (!normalized) return false;
    for (var i = 0; i < allowedRoles.length; i += 1) {
      if (normalized === String(allowedRoles[i] || "").trim().toLowerCase()) {
        return true;
      }
    }
    return false;
  }

  function roleAllowsModuleAction(role, module, action) {
    var normalizedRole = normalizePermissionRole(role);
    var normalizedModule = String(module || "").trim().toLowerCase();
    var normalizedAction = String(action || permActionRead).trim().toUpperCase();
    var allReadRoles = ["admin_empresa", "supervisor_sucursal", "cajero", "inventario", "compras", "contabilidad", "auditor"];

    if (normalizedRole === "super_administrador") {
      return true;
    }

    switch (normalizedModule) {
      case permModuleVentas:
        if (normalizedAction === permActionRead) return roleIn(normalizedRole, allReadRoles);
        if (normalizedAction === permActionCreate || normalizedAction === permActionUpdate || normalizedAction === "D" || normalizedAction === permActionApprove) {
          return roleIn(normalizedRole, ["admin_empresa", "supervisor_sucursal", "cajero"]);
        }
        break;

      case permModuleInventario:
        if (normalizedAction === permActionRead) return roleIn(normalizedRole, allReadRoles);
        if (normalizedAction === permActionCreate || normalizedAction === permActionUpdate || normalizedAction === "D" || normalizedAction === permActionApprove) {
          return roleIn(normalizedRole, ["admin_empresa", "supervisor_sucursal", "inventario"]);
        }
        break;

      case permModuleFinanzas:
        if (normalizedAction === permActionRead) return roleIn(normalizedRole, allReadRoles);
        if (normalizedAction === permActionCreate || normalizedAction === permActionUpdate || normalizedAction === permActionApprove) {
          return roleIn(normalizedRole, ["admin_empresa", "contabilidad"]);
        }
        if (normalizedAction === "D") {
          return roleIn(normalizedRole, ["contabilidad"]);
        }
        break;

      case permModuleClientes:
        if (normalizedAction === permActionRead) return roleIn(normalizedRole, allReadRoles);
        if (normalizedAction === permActionCreate || normalizedAction === permActionUpdate || normalizedAction === permActionApprove) {
          return roleIn(normalizedRole, ["admin_empresa", "supervisor_sucursal", "cajero"]);
        }
        if (normalizedAction === "D") {
          return false;
        }
        break;

      case permModuleFacturacion:
        if (normalizedAction === permActionRead) return roleIn(normalizedRole, allReadRoles);
        if (normalizedAction === permActionCreate || normalizedAction === permActionUpdate || normalizedAction === permActionApprove) {
          return roleIn(normalizedRole, ["admin_empresa", "cajero"]);
        }
        if (normalizedAction === "D") {
          return false;
        }
        break;

      case permModuleSeguridad:
        if (normalizedAction === permActionRead) return roleIn(normalizedRole, allReadRoles);
        if (normalizedAction === permActionCreate || normalizedAction === permActionUpdate || normalizedAction === "D" || normalizedAction === permActionApprove) {
          return roleIn(normalizedRole, ["admin_empresa"]);
        }
        break;
    }

    return false;
  }

  function setMenuLinkVisible(link, visible) {
    if (!link) return;
    var item = null;
    if (typeof link.closest === "function") {
      item = link.closest("li");
    }
    if (!item) {
      item = link.parentElement;
    }
    if (item) {
      item.style.display = visible ? "" : "none";
    }
    link.setAttribute("data-menu-visible", visible ? "1" : "0");
    if (!visible) {
      link.classList.remove("active");
    }
  }

  function isMenuLinkVisible(link) {
    if (!link) return false;
    return link.getAttribute("data-menu-visible") !== "0";
  }

  function canRoleAccessLink(role, link) {
    if (!link) return false;
    var rule = menuPermissionCatalog[link.id || ""];
    if (!rule || rule.alwaysVisible) {
      return true;
    }
    return roleAllowsModuleAction(role, rule.module, rule.action);
  }

  function applyMenuPermissionsByRole(rawRole) {
    var normalizedRole = normalizePermissionRole(rawRole);
    links.forEach(function (link) {
      setMenuLinkVisible(link, true);
    });
    if (!normalizedRole) {
      return;
    }
    links.forEach(function (link) {
      setMenuLinkVisible(link, canRoleAccessLink(normalizedRole, link));
    });
  }

  function isVisibleMenuHref(href) {
    var current = normalizeHref(href).split("?")[0];
    if (!current) return false;
    for (var i = 0; i < links.length; i += 1) {
      var link = links[i];
      if (!isMenuLinkVisible(link)) continue;
      var linkHref = normalizeHref(link.getAttribute("href")).split("?")[0];
      if (linkHref && linkHref === current) {
        return true;
      }
    }
    return false;
  }

  function firstVisibleFrameSrc(empresaId) {
    for (var i = 0; i < links.length; i += 1) {
      var link = links[i];
      if (!isMenuLinkVisible(link)) continue;
      var href = withEmpresaParam(link.getAttribute("href"), empresaId);
      if (isAllowedFrameHref(href)) {
        return href;
      }
    }
    return defaultFrameSrc(empresaId);
  }

  function resolveInitialFrameSrc(empresaId) {
    var restored = getStoredFrameSrc(empresaId);
    if (restored && isVisibleMenuHref(restored)) {
      return restored;
    }
    return firstVisibleFrameSrc(empresaId);
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

  function fetchCurrentAdminRole() {
    return fetch("/me", { credentials: "same-origin" })
      .then(function (resp) {
        if (!resp.ok) return null;
        return resp.json();
      })
      .then(function (data) {
        if (!data || typeof data !== "object") return "";
        return String(data.role || data.Role || "").trim();
      })
      .catch(function () {
        return "";
      });
  }

  function loadEmpresaTitle(empresaId) {
    return fetch("/super/api/empresas?id=" + encodeURIComponent(empresaId), { credentials: "same-origin" })
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
  }

  function initializeMenuAndFrame(empresaId) {
    setLinksWithEmpresa(empresaId);
    if (!frame) return;
    var initialSrc = resolveInitialFrameSrc(empresaId);
    frame.src = initialSrc;
    setActiveByHref(initialSrc);
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

  fetchCurrentAdminRole()
    .then(function (role) {
      applyMenuPermissionsByRole(role);
      if (id) {
        initializeMenuAndFrame(id);
        loadEmpresaTitle(id);
        return;
      }
      initializeMenuAndFrame("");
      title.textContent = "Administrar Empresa";
    })
    .catch(function () {
      if (id) {
        initializeMenuAndFrame(id);
        loadEmpresaTitle(id);
        return;
      }
      initializeMenuAndFrame("");
      title.textContent = "Administrar Empresa";
    });
})();
