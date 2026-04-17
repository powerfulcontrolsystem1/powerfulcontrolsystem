function getQueryParam(name) {
  var params = new URLSearchParams(window.location.search);
  var value = params.get(name);
  if (value) {
    return value;
  }
  try {
    if (window.parent && window.parent !== window && window.parent.location) {
      var parentParams = new URLSearchParams(window.parent.location.search || "");
      var parentValue = parentParams.get(name);
      if (parentValue) {
        return parentValue;
      }
    }
  } catch (e) {
    // no-op: acceso a parent puede fallar en algunos contextos
  }
  return "";
}

function parsePositiveInt(raw) {
  var n = Number(String(raw || "").trim());
  if (!Number.isFinite(n)) return 0;
  n = Math.trunc(n);
  return n > 0 ? n : 0;
}

function readEmpresaIdFromStorage() {
  var keys = ["active_empresa_id", "empresa_id", "admin_empresa_id"];
  var stores = [];
  try { stores.push(window.sessionStorage); } catch (e) {}
  try { stores.push(window.localStorage); } catch (e) {}

  for (var s = 0; s < stores.length; s += 1) {
    var store = stores[s];
    if (!store) continue;
    for (var i = 0; i < keys.length; i += 1) {
      var key = keys[i];
      var raw = "";
      try {
        raw = store.getItem(key) || "";
      } catch (e) {
        raw = "";
      }
      var id = parsePositiveInt(raw);
      if (id > 0) {
        return String(id);
      }
    }
  }
  return "";
}

function persistEmpresaIdInStorage(rawEmpresaId) {
  var id = parsePositiveInt(rawEmpresaId);
  if (!id) return "";
  var value = String(id);
  try {
    window.sessionStorage.setItem("active_empresa_id", value);
    window.sessionStorage.setItem("empresa_id", value);
    window.sessionStorage.setItem("admin_empresa_id", value);
  } catch (e) {}
  try {
    window.localStorage.setItem("active_empresa_id", value);
    window.localStorage.setItem("empresa_id", value);
    window.localStorage.setItem("admin_empresa_id", value);
  } catch (e) {}
  return value;
}

function resolveEmpresaIdContext() {
  var fromUrl = parsePositiveInt(getQueryParam("empresa_id") || getQueryParam("id"));
  if (fromUrl > 0) {
    return persistEmpresaIdInStorage(fromUrl);
  }
  var fromStorage = readEmpresaIdFromStorage();
  if (fromStorage) {
    return persistEmpresaIdInStorage(fromStorage);
  }
  return "";
}

try {
  window.__resolveEmpresaIdContext = function () {
    return resolveEmpresaIdContext();
  };
} catch (e) {
  // no-op
}

(function () {
  var id = persistEmpresaIdInStorage(getQueryParam("id") || getQueryParam("empresa_id"));
  if (!id) {
    id = resolveEmpresaIdContext();
  }
  var titleMenu = document.getElementById("empresaTitleMenu");
  var title = titleMenu || document.getElementById("empresaTitle");
  var frame = document.getElementById("contentFrame") || document.querySelector("iframe.admin-empresa-frame");
  var portalUsuariosLink = document.getElementById("linkPortalUsuarios");
  var permsEvidence = document.getElementById("menuPermsEvidence");
  var storage = null;
  try {
    storage = window.sessionStorage;
  } catch (e) {
    storage = null;
  }
  var links = [
    document.getElementById("linkInicio"),
    document.getElementById("linkEstaciones"),
    document.getElementById("linkVentas"),
    document.getElementById("linkVentaPublica"),
    document.getElementById("linkProductos"),
    document.getElementById("linkConfiguracion"),
    document.getElementById("linkFacturacionElectronica"),
    document.getElementById("linkChatIA"),
    document.getElementById("linkFinanzas"),
    document.getElementById("linkBackups"),
    document.getElementById("linkSoporteRemoto"),
    document.getElementById("linkUbicacionGPS"),
    document.getElementById("linkReservasHotel"),
    document.getElementById("linkReportes"),
    document.getElementById("linkUsuarios"),
    document.getElementById("linkCodigosDescuento"),
    document.getElementById("linkAsistenciaEmpleados"),
    document.getElementById("linkVehiculosRegistro"),
    document.getElementById("linkAuditoria"),
    document.getElementById("linkChatTareas"),
    document.getElementById("linkClientes"),
    document.getElementById("linkCalculadora"),
  ];

  var permActionRead = "R";
  var permActionCreate = "C";
  var permActionUpdate = "U";
  var permActionApprove = "A";

  var permModuleVentas = "ventas";
  var permModuleInventario = "inventario";
  var permModuleCompras = "compras";
  var permModuleFinanzas = "finanzas";
  var permModuleClientes = "clientes";
  var permModuleFacturacion = "facturacion";
  var permModuleSeguridad = "seguridad";

  var menuPermissionCatalog = {
    linkInicio: { alwaysVisible: true },
    linkVentas: { module: permModuleVentas, action: permActionRead },
    linkCarritoCompras: { module: permModuleVentas, action: permActionCreate },
    linkVentaPublica: { module: permModuleVentas, action: permActionCreate },
    linkProductos: { module: permModuleInventario, action: permActionCreate },
    linkCombosProductos: { module: permModuleInventario, action: permActionCreate },
    linkCodigosDescuento: { module: permModuleVentas, action: permActionCreate },
    linkCompras: { module: permModuleCompras, action: permActionCreate },
    linkConfiguracion: { module: permModuleSeguridad, action: permActionUpdate },
    linkUsuarios: { module: permModuleSeguridad, action: permActionUpdate },
    linkAsistenciaEmpleados: { module: permModuleSeguridad, action: permActionUpdate },
    linkNominaSueldos: { module: permModuleFinanzas, action: permActionCreate },
    linkVehiculosRegistro: { module: permModuleSeguridad, action: permActionCreate },
    linkAuditoria: { module: permModuleSeguridad, action: permActionRead },
    linkChatTareas: { module: permModuleVentas, action: permActionCreate },
    linkClientes: { module: permModuleClientes, action: permActionCreate },
    linkFacturacionElectronica: { module: permModuleFacturacion, action: permActionCreate },
    linkFacturasElectronicas: { module: permModuleFacturacion, action: permActionRead },
    linkERPExtendido: { module: permModuleSeguridad, action: permActionUpdate },
    linkChatIA: { module: permModuleVentas, action: permActionRead },
    linkFinanzas: { module: permModuleFinanzas, action: permActionCreate },
    linkCreditos: { module: permModuleFinanzas, action: permActionCreate },
    linkBackups: { module: permModuleSeguridad, action: permActionApprove },
    linkSoporteRemoto: { module: permModuleSeguridad, action: permActionApprove },
    linkPropinas: { module: permModuleFinanzas, action: permActionCreate },
    linkComisiones: { module: permModuleFinanzas, action: permActionCreate },
    linkUbicacionGPS: { module: permModuleInventario, action: permActionCreate },
    linkConfigEstaciones: { module: permModuleVentas, action: permActionApprove },
    linkTarifasPorMinutos: { module: permModuleVentas, action: permActionCreate },
    linkTarifasPorDia: { module: permModuleVentas, action: permActionCreate },
    linkEstaciones: { alwaysVisible: true },
    
    linkReservasHotel: { module: permModuleVentas, action: permActionCreate },
    linkReportes: { module: permModuleFinanzas, action: permActionRead },
    linkCalculadora: { module: permModuleFinanzas, action: permActionRead },
    linkGraficosEstadisticas: { module: permModuleFinanzas, action: permActionRead },
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

  function buildPortalUsuariosURL(empresaId, config) {
  var fallback = "/login_usuario.html";
  if (empresaId) {
    fallback += "?empresa_id=" + encodeURIComponent(String(empresaId));
  }
  var cfg = config || {};
  var targetEmpresaId = Number(empresaId || 0);
  var customDomain = String(cfg.dominio_publico || "").trim();
  if (customDomain) {
    try {
    if (customDomain.indexOf("://") === -1) {
      customDomain = "https://" + customDomain;
    }
    var customURL = new URL(customDomain);
    customURL.pathname = "/login_usuario.html";
    customURL.search = "";
    if (targetEmpresaId > 0) {
      customURL.searchParams.set("empresa_id", String(targetEmpresaId));
    }
    return customURL.toString();
    } catch (e) {
    return fallback;
    }
  }
  var slug = String(cfg.empresa_slug || "").trim().toLowerCase();
  if (!slug) return fallback;
  try {
    var url = new URL(window.location.origin);
    var host = String(url.hostname || "").toLowerCase();
    if (host === "powerfulcontrolsystem.com" || host === "www.powerfulcontrolsystem.com" || host.endsWith(".powerfulcontrolsystem.com")) {
    url.protocol = "https:";
    url.hostname = slug + ".powerfulcontrolsystem.com";
    url.pathname = "/login_usuario.html";
    url.search = "";
    if (targetEmpresaId > 0) {
      url.searchParams.set("empresa_id", String(targetEmpresaId));
    }
    return url.toString();
    }
  } catch (e) {
    return fallback;
  }
  return fallback;
  }

  async function resolvePortalUsuariosURL(empresaId) {
  var fallback = buildPortalUsuariosURL(empresaId, null);
  if (!empresaId) return fallback;
  try {
    var res = await fetch("/api/empresa/venta_publica?empresa_id=" + encodeURIComponent(String(empresaId)) + "&action=config", { credentials: "same-origin" });
    if (!res.ok) return fallback;
    var body = await res.json();
    return buildPortalUsuariosURL(empresaId, body && body.config ? body.config : null);
  } catch (e) {
    return fallback;
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

  if (portalUsuariosLink) {
  resolvePortalUsuariosURL(id).then(function (url) {
    portalUsuariosLink.href = url;
  }).catch(function () {
    portalUsuariosLink.href = buildPortalUsuariosURL(id, null);
  });
  portalUsuariosLink.addEventListener("click", function (event) {
    event.preventDefault();
    resolvePortalUsuariosURL(id).then(function (url) {
    portalUsuariosLink.href = url;
    window.open(url, "_blank", "noopener");
    }).catch(function () {
    window.open(buildPortalUsuariosURL(id, null), "_blank", "noopener");
    });
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

  function normalizePermissionAction(raw) {
    var value = String(raw || "").trim().toUpperCase();
    if (!value) return permActionRead;
    return value;
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
    var normalizedAction = normalizePermissionAction(action);
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

      case permModuleCompras:
        if (normalizedAction === permActionRead) return roleIn(normalizedRole, allReadRoles);
        if (normalizedAction === permActionCreate || normalizedAction === permActionUpdate || normalizedAction === "D" || normalizedAction === permActionApprove) {
          return roleIn(normalizedRole, ["admin_empresa", "supervisor_sucursal", "compras"]);
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

  function setMenuPermissionsEvidence(text, isFallback) {
    if (!permsEvidence) return;
    permsEvidence.textContent = text || "";
    if (isFallback) {
      permsEvidence.style.opacity = "0.85";
      return;
    }
    permsEvidence.style.opacity = "1";
  }

  function getPermissionContextModuleRow(permissionContext, moduleName) {
    if (!permissionContext || !Array.isArray(permissionContext.modulos)) {
      return null;
    }
    var target = String(moduleName || "").trim().toLowerCase();
    if (!target) {
      return null;
    }
    for (var i = 0; i < permissionContext.modulos.length; i += 1) {
      var row = permissionContext.modulos[i];
      var rowModule = String(row && row.modulo || "").trim().toLowerCase();
      if (rowModule && rowModule === target) {
        return row;
      }
    }
    return null;
  }

  function boolFromActionMap(actionMap, actionKey) {
    if (!actionMap || typeof actionMap !== "object") {
      return false;
    }
    if (Object.prototype.hasOwnProperty.call(actionMap, actionKey)) {
      return !!actionMap[actionKey];
    }
    var lowerKey = String(actionKey || "").toLowerCase();
    if (Object.prototype.hasOwnProperty.call(actionMap, lowerKey)) {
      return !!actionMap[lowerKey];
    }
    return false;
  }

  function isContextModuleActionAllowed(moduleRow, action) {
    if (!moduleRow || typeof moduleRow !== "object") {
      return false;
    }
    var actionKey = normalizePermissionAction(action);
    if (actionKey === permActionRead && typeof moduleRow.read !== "undefined") {
      return !!moduleRow.read;
    }
    if (actionKey === permActionCreate && typeof moduleRow.create !== "undefined") {
      return !!moduleRow.create;
    }
    if (actionKey === permActionUpdate && typeof moduleRow.update !== "undefined") {
      return !!moduleRow.update;
    }
    if (actionKey === "D" && typeof moduleRow.delete !== "undefined") {
      return !!moduleRow.delete;
    }
    if (actionKey === permActionApprove && typeof moduleRow.approve !== "undefined") {
      return !!moduleRow.approve;
    }
    return boolFromActionMap(moduleRow.acciones, actionKey);
  }

  function canPermissionContextAccessLink(permissionContext, link) {
    if (!link) return false;
    var pageKey = link.id || "";
    var pages = permissionContext && permissionContext.paginas;
    if (pageKey && pages && typeof pages === "object" && Object.prototype.hasOwnProperty.call(pages, pageKey)) {
      return !!pages[pageKey];
    }
    var rule = menuPermissionCatalog[link.id || ""];
    if (!rule || rule.alwaysVisible) {
      return true;
    }
    var moduleRow = getPermissionContextModuleRow(permissionContext, rule.module);
    return isContextModuleActionAllowed(moduleRow, rule.action);
  }

  function applyMenuPermissionsByContext(permissionContext) {
    links.forEach(function (link) {
      setMenuLinkVisible(link, true);
    });
    if (!permissionContext) {
      return;
    }
    links.forEach(function (link) {
      setMenuLinkVisible(link, canPermissionContextAccessLink(permissionContext, link));
    });
  }

  function describePermissionContext(permissionContext) {
    if (!permissionContext || typeof permissionContext !== "object") {
      return "Permisos de menú: sin contexto disponible.";
    }
    var role = normalizePermissionRole(permissionContext.rol || "sin_rol") || "sin_rol";
    var summary = permissionContext.resumen || {};
    var modulesTotal = Number(summary.modulos_total || 0);
    var modulesRead = Number(summary.modulos_lectura || 0);
    var modulesApprove = Number(summary.modulos_aprobacion || 0);
    var enabledActions = Number(summary.acciones_habilitadas || 0);
    return "Permisos de menú: rol " + role +
      " | lectura " + modulesRead + "/" + modulesTotal +
      " | aprobación " + modulesApprove +
      " | acciones habilitadas " + enabledActions +
      " | fuente: /api/empresa/permisos_contexto";
  }

  function fetchEmpresaPermisosContexto(empresaId) {
    if (!empresaId) {
      return Promise.resolve(null);
    }
    var url = "/api/empresa/permisos_contexto?empresa_id=" + encodeURIComponent(empresaId);
    return fetch(url, { credentials: "same-origin" })
      .then(function (resp) {
        if (!resp.ok) return null;
        return resp.json();
      })
      .then(function (data) {
        if (!data || typeof data !== "object") return null;
        if (!Array.isArray(data.modulos)) return null;
        return data;
      })
      .catch(function () {
        return null;
      });
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

  function applyMenuPermissionsWithSource(empresaId, role) {
    var normalizedRole = normalizePermissionRole(role);
    return fetchEmpresaPermisosContexto(empresaId)
      .then(function (permissionContext) {
        if (permissionContext) {
          applyMenuPermissionsByContext(permissionContext);
          setMenuPermissionsEvidence(describePermissionContext(permissionContext), false);
          return;
        }
        applyMenuPermissionsByRole(normalizedRole);
        if (normalizedRole) {
          setMenuPermissionsEvidence("Permisos de menú: rol " + normalizedRole + " | fuente local de respaldo.", true);
        } else {
          setMenuPermissionsEvidence("Permisos de menú: sin rol detectado | fuente local de respaldo.", true);
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
          if (titleMenu) titleMenu.textContent = "Administrar Empresa";
          else if (title) title.textContent = "Administrar Empresa";
          throw new Error("empresa no encontrada");
        }
        return resp.json();
      })
      .then(function (data) {
        var nombre = data && (data.nombre || data.Nombre);
        if (nombre) {
          if (titleMenu) {
            titleMenu.textContent = "Administrar Empresa - " + nombre;
            document.title = titleMenu.textContent;
          } else if (title) {
            var cur = String(title.textContent || "").trim();
            // Only override if the current title is the generic placeholder
            if (!cur || cur === "Administrar Empresa" || cur.indexOf("Administrar Empresa -") === 0) {
              title.textContent = "Administrar Empresa - " + nombre;
              document.title = title.textContent;
            }
          }
        } else {
          if (titleMenu) titleMenu.textContent = "Administrar Empresa";
          else if (title) {
            var cur2 = String(title.textContent || "").trim();
            if (!cur2 || cur2.indexOf("Administrar Empresa -") === 0) {
              title.textContent = "Administrar Empresa";
            }
          }
        }
      })
      .catch(function (err) {
        console.warn("No se pudo cargar empresa:", err);
        if (titleMenu) titleMenu.textContent = "Administrar Empresa";
        else if (title) {
          var cur3 = String(title.textContent || "").trim();
          if (!cur3 || cur3.indexOf("Administrar Empresa -") === 0 || cur3 === "Administrar Empresa") {
            title.textContent = "Administrar Empresa";
          }
        }
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

      // Si una navegación interna del iframe pierde empresa_id,
      // se corrige automáticamente usando el contexto activo.
      if (id) {
        try {
          var normalizedCurrent = normalizeHref(currentHref);
          var currentURL = new URL(normalizedCurrent || currentHref, window.location.origin);
          var hasEmpresaID = parsePositiveInt(currentURL.searchParams.get("empresa_id")) > 0;
          if (!hasEmpresaID) {
            var correctedHref = withEmpresaParam(currentURL.pathname + currentURL.search, id);
            if (correctedHref && correctedHref !== normalizedCurrent) {
              frame.setAttribute("src", correctedHref);
              return;
            }
          }
        } catch (e) {
          // no-op
        }
      }

      persistFrameSrc(currentHref, id);
      setActiveByHref(currentHref);
    });
    // Interceptar F5 / Ctrl+R para recargar solo el iframe y mantener la subpágina activa.
    // Si el foco está en un campo editable (input/textarea/contentEditable) se respeta el comportamiento por defecto.
    document.addEventListener('keydown', function (ev) {
      try {
        var isF5 = ev.key === 'F5' || ev.keyCode === 116;
        var isCtrlR = (ev.ctrlKey || ev.metaKey) && (ev.key === 'r' || ev.keyCode === 82);
        if (!isF5 && !isCtrlR) return;

        var active = document.activeElement;
        var tag = (active && active.tagName) ? active.tagName.toLowerCase() : '';
        var isEditable = tag === 'input' || tag === 'textarea' || (active && active.isContentEditable);
        if (isEditable && !active.readOnly) {
          // permitir refresco normal cuando el usuario está editando
          return;
        }

        ev.preventDefault();
        if (frame && frame.contentWindow) {
          try {
            frame.contentWindow.location.reload();
            return;
          } catch (e) {
            // si por alguna razón no es posible acceder al contentWindow, forzamos reload asignando src
            try {
              var src = frame.getAttribute('src') || frame.src;
              frame.setAttribute('src', src);
              return;
            } catch (e2) {
              // fallback al reload global
            }
          }
        }
        window.location.reload();
      } catch (e) {
        // no-op
      }
    });
  }

  fetchCurrentAdminRole()
    .then(function (role) {
      if (id) {
        return applyMenuPermissionsWithSource(id, role)
          .then(function () {
            initializeMenuAndFrame(id);
            loadEmpresaTitle(id);
          });
      }
      applyMenuPermissionsByRole(role);
      initializeMenuAndFrame("");
      title.textContent = "Administrar Empresa";
      return null;
    })
    .catch(function () {
      if (id) {
        applyMenuPermissionsByRole("");
        setMenuPermissionsEvidence("Permisos de menú: no se pudo resolver contexto, se mantiene visibilidad base.", true);
        initializeMenuAndFrame(id);
        loadEmpresaTitle(id);
        return;
      }
      applyMenuPermissionsByRole("");
      initializeMenuAndFrame("");
      title.textContent = "Administrar Empresa";
    });
})();
