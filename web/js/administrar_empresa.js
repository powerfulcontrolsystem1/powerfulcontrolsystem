function getQueryParam(name) {
  var params = new URLSearchParams(window.location.search);
  return params.get(name);
}

(function () {
  var id = getQueryParam("id") || getQueryParam("empresa_id");
  var title = document.getElementById("empresaTitle");
  var frame = document.getElementById("contentFrame");
  var permsEvidence = document.getElementById("menuPermsEvidence");
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
    document.getElementById("linkVentaPublica"),
    document.getElementById("linkProductos"),
    document.getElementById("linkCombosProductos"),
    document.getElementById("linkCodigosDescuento"),
    document.getElementById("linkCompras"),
    document.getElementById("linkConfiguracion"),
    document.getElementById("linkUsuarios"),
    document.getElementById("linkAsistenciaEmpleados"),
    document.getElementById("linkNominaSueldos"),
    document.getElementById("linkVehiculosRegistro"),
    document.getElementById("linkAuditoria"),
    document.getElementById("linkChatTareas"),
    document.getElementById("linkClientes"),
    document.getElementById("linkFacturacionElectronica"),
    document.getElementById("linkFacturasElectronicas"),
    document.getElementById("linkERPExtendido"),
    document.getElementById("linkChatIA"),
    document.getElementById("linkFinanzas"),
    document.getElementById("linkCreditos"),
    document.getElementById("linkBackups"),
    document.getElementById("linkSoporteRemoto"),
    document.getElementById("linkPropinas"),
    document.getElementById("linkComisiones"),
    document.getElementById("linkUbicacionGPS"),
    document.getElementById("linkConfigEstaciones"),
    document.getElementById("linkTarifasPorMinutos"),
    document.getElementById("linkTarifasPorDia"),
    document.getElementById("linkEstaciones"),
    document.getElementById("linkReservasHotel"),
    document.getElementById("linkReportes"),
    document.getElementById("linkCalculadora"),
    document.getElementById("linkGraficosEstadisticas"),
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
    linkEstaciones: { module: permModuleVentas, action: permActionUpdate },
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
