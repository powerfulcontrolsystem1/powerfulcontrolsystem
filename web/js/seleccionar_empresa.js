(function () {
  var empresasPanel = document.getElementById("empresasPanel");
  var contentFrame = document.getElementById("contentFrame");
  var navLinks = Array.from(document.querySelectorAll(".admin-sidebar .nav a"));
  var storage = null;
  var viewKey = "seleccionar_empresa:view";
  var currentEmpresas = [];
  var currentAccount = null;
  var currentActiveByEmpresa = {};
  var selectorEmpresasOrder = [];
  var selectorOrderSaveTimer = null;
  var selectorDragState = null;
  var suppressEmpresaCardClickUntil = 0;
  var shareNoticeEl = document.getElementById("selectorShareNotice");
  var shareInvitesPanel = null;
  var nuevasPlantillasCatalog = Array.isArray(window.PCS_NUEVAS_PLANTILLAS) ? window.PCS_NUEVAS_PLANTILLAS.slice() : [];
  var nuevasPlantillasByModule = {};
  nuevasPlantillasCatalog.forEach(function (item) {
    var module = String(item && item.module ? item.module : "").trim().toLowerCase();
    if (module) nuevasPlantillasByModule[module] = item;
  });
  var empresaShareModuleCatalog = [
    ["ventas", "Ventas"],
    ["inventario", "Inventario"],
    ["finanzas", "Finanzas, caja y reportes"],
    ["facturacion", "Facturacion electronica"],
    ["clientes", "Clientes"],
    ["crm_unificado", "CRM"],
    ["compras", "Compras"],
    ["reportes", "Reportes"],
    ["documentos_onlyoffice", "Documentos"]
  ];
  var deleteModalState = {
    empresa: null,
    impacto: null,
    descargaOfrecida: false,
    deleting: false
  };

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

  function isStandaloneFrameHref(href) {
    var normalized = normalizeHref(href);
    if (!normalized) return false;
    var path = normalized.split("?")[0].toLowerCase();
    return path === "/editar_empresa.html" ||
      path === "/editar_empresa.htm" ||
      path === "/descargar_informacion_de_la_empresa.html" ||
      path === "/descargar_informacion_de_la_empresa.htm";
  }

  function escapeHtml(value) {
    return String(value == null ? "" : value).replace(/[&<>"']/g, function (match) {
      return {
        "&": "&amp;",
        "<": "&lt;",
        ">": "&gt;",
        '"': "&quot;",
        "'": "&#39;"
      }[match];
    });
  }

  async function fetchJSON(url, options) {
    var requestOptions = options ? Object.assign({}, options) : {};
    if (!requestOptions.credentials) {
      requestOptions.credentials = "same-origin";
    }
    var res = await fetch(url, requestOptions);
    var raw = await res.text();
    var data = null;
    try {
      data = raw ? JSON.parse(raw) : null;
    } catch (e) {
      data = null;
    }
    if (!res.ok) {
      var err = new Error((data && (data.error || data.message)) || raw || "Solicitud fallida");
      err.status = res.status;
      err.payload = data;
      throw err;
    }
    return data;
  }

  function normalizeEmpresaOrderIDs(values) {
    var seen = {};
    var out = [];
    if (!Array.isArray(values)) return out;
    values.forEach(function (value) {
      var id = Number(value || 0);
      if (!Number.isFinite(id) || id <= 0) return;
      id = Math.trunc(id);
      if (seen[id]) return;
      seen[id] = true;
      out.push(id);
    });
    return out;
  }

  function selectorOrderStorageKey() {
    var email = "";
    try {
      email = normalizeEmail((currentAccount && (currentAccount.email || (currentAccount.admin && currentAccount.admin.email))) || "");
    } catch (e) {
      email = "";
    }
    return "seleccionar_empresa:orden:" + (email || "anonimo");
  }

  function readSelectorOrderLocal() {
    try {
      var raw = window.localStorage.getItem(selectorOrderStorageKey());
      if (!raw) return [];
      return normalizeEmpresaOrderIDs(JSON.parse(raw));
    } catch (e) {
      return [];
    }
  }

  function writeSelectorOrderLocal(order) {
    try {
      window.localStorage.setItem(selectorOrderStorageKey(), JSON.stringify(normalizeEmpresaOrderIDs(order)));
    } catch (e) {}
  }

  async function loadSelectorEmpresasOrder() {
    selectorEmpresasOrder = readSelectorOrderLocal();
    try {
      var data = await fetchJSON("/api/user/configuracion", { credentials: "same-origin" });
      var remoteOrder = data && (
        data.selector_empresas_orden ||
        data.selector_empresas_order ||
        data.selector_companies_order
      );
      if (Array.isArray(remoteOrder)) {
        var order = normalizeEmpresaOrderIDs(remoteOrder);
        selectorEmpresasOrder = order;
        writeSelectorOrderLocal(order);
      }
    } catch (err) {
      // La preferencia local conserva la experiencia si la configuracion remota no responde.
    }
    return selectorEmpresasOrder.slice();
  }

  function saveSelectorEmpresasOrder(order, immediate) {
    var normalized = normalizeEmpresaOrderIDs(order);
    selectorEmpresasOrder = normalized;
    writeSelectorOrderLocal(normalized);
    if (selectorOrderSaveTimer) {
      window.clearTimeout(selectorOrderSaveTimer);
      selectorOrderSaveTimer = null;
    }
    var persist = function () {
      return fetchJSON("/api/user/configuracion", {
        method: "POST",
        credentials: "same-origin",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ selector_empresas_orden: normalized })
      }).then(function () {
        setShareNotice("Orden de empresas guardado.", false);
      }).catch(function () {
        setShareNotice("El orden se guardo en este navegador, pero no se pudo sincronizar con tu cuenta.", true);
      });
    };
    if (immediate) {
      return persist();
    } else {
      selectorOrderSaveTimer = window.setTimeout(persist, 500);
    }
    return Promise.resolve({ ok: true });
  }

  function collectRenderedEmpresaOrder() {
    return normalizeEmpresaOrderIDs(Array.prototype.map.call(
      document.querySelectorAll(".empresa-card-link[data-empresa-id]"),
      function (node) {
        return node.getAttribute("data-empresa-id");
      }
    ));
  }

  function sortEmpresasBySelectorOrder(empresas) {
    var base = Array.isArray(empresas) ? empresas.slice() : [];
    var indexByID = {};
    selectorEmpresasOrder.forEach(function (id, index) {
      indexByID[String(id)] = index;
    });
    return base.sort(function (left, right) {
      var leftID = String(left && left.id ? left.id : "");
      var rightID = String(right && right.id ? right.id : "");
      var leftHas = Object.prototype.hasOwnProperty.call(indexByID, leftID);
      var rightHas = Object.prototype.hasOwnProperty.call(indexByID, rightID);
      if (leftHas && rightHas) {
        return indexByID[leftID] - indexByID[rightID];
      }
      if (leftHas !== rightHas) {
        return leftHas ? -1 : 1;
      }
      var leftShared = String(left && left.access_source ? left.access_source : "owner").toLowerCase() === "shared";
      var rightShared = String(right && right.access_source ? right.access_source : "owner").toLowerCase() === "shared";
      if (leftShared !== rightShared) {
        return leftShared ? 1 : -1;
      }
      return String(left && left.nombre ? left.nombre : "").localeCompare(String(right && right.nombre ? right.nombre : ""), "es", { sensitivity: "base" });
    });
  }

  function setShareNotice(text, isError) {
    if (!shareNoticeEl) return;
    shareNoticeEl.classList.toggle("is-hidden", !text);
    shareNoticeEl.textContent = text || "";
    shareNoticeEl.classList.toggle("error", !!isError);
    shareNoticeEl.classList.toggle("success", !isError && !!text);
  }

  function ensureShareInvitesPanel() {
    if (shareInvitesPanel) return shareInvitesPanel;
    if (!shareNoticeEl || !shareNoticeEl.parentNode) return null;
    var panel = document.getElementById("selectorShareInvitesPanel");
    if (!panel) {
      panel = document.createElement("div");
      panel.id = "selectorShareInvitesPanel";
      panel.className = "card";
      panel.style.marginTop = "12px";
      panel.style.padding = "14px 16px";
      panel.style.borderRadius = "14px";
      panel.style.display = "none";
      shareNoticeEl.parentNode.insertBefore(panel, shareNoticeEl.nextSibling);
    }
    shareInvitesPanel = panel;
    return panel;
  }

  function renderPendingShareInvites(items) {
    var panel = ensureShareInvitesPanel();
    if (!panel) return;
    if (!Array.isArray(items) || items.length === 0) {
      panel.style.display = "none";
      panel.innerHTML = "";
      return;
    }
    panel.style.display = "";
    panel.innerHTML =
      '<div style="display:flex;align-items:center;justify-content:space-between;gap:12px;flex-wrap:wrap;">' +
      '<div><strong>Invitaciones de empresas pendientes</strong><div class="muted" style="margin-top:4px;">Acepta para que aparezcan en tu lista y se abran automáticamente.</div></div>' +
      '<button type="button" class="btn secondary" data-action="refresh-share-invites">Actualizar</button>' +
      "</div>" +
      '<div style="margin-top:10px;display:grid;gap:10px;">' +
      items.map(function (it) {
        var empresaNombre = String(it.empresa_nombre || "").trim() || ("Empresa #" + String(it.empresa_id || ""));
        var invitadoPor = String(it.invitado_por || "").trim();
        var expira = String(it.expira_en || "").trim();
        var msg = String(it.mensaje || "").trim();
        return '' +
          '<article class="card" style="padding:12px 14px;border-radius:14px;border:1px solid rgba(148,163,184,.25);background:rgba(15,23,42,.35)">' +
          '<div style="display:flex;align-items:flex-start;justify-content:space-between;gap:12px;flex-wrap:wrap;">' +
          '<div style="min-width:220px;">' +
          '<div><strong>' + escapeHtml(empresaNombre) + "</strong></div>" +
          (invitadoPor ? '<div class="muted">Compartida por: ' + escapeHtml(invitadoPor) + "</div>" : "") +
          (expira ? '<div class="muted">Expira: ' + escapeHtml(expira) + "</div>" : "") +
          (msg ? '<div class="muted" style="margin-top:6px;">Mensaje: ' + escapeHtml(msg) + "</div>" : "") +
          "</div>" +
          '<div style="display:flex;gap:8px;align-items:center;">' +
          '<button type="button" class="btn" data-action="accept-share-invite" data-invitation-id="' + escapeHtml(String(it.id || "")) + '" data-empresa-id="' + escapeHtml(String(it.empresa_id || "")) + '">Aceptar</button>' +
          '<button type="button" class="btn secondary empresa-share-resend" data-empresa-id="' + escapeHtml(String(it.empresa_id || "")) + '" data-invitation-id="' + escapeHtml(String(it.id || "")) + '">Reenviar email</button>' +
          "</div>" +
          "</div>" +
          "</article>";
      }).join("") +
      "</div>";
  }

  async function loadPendingShareInvites() {
    try {
      var data = await fetchJSON("/super/api/empresas/compartidos?action=pendientes_mias", { credentials: "same-origin" });
      var items = data && Array.isArray(data.items) ? data.items : [];
      renderPendingShareInvites(items);
    } catch (err) {
      renderPendingShareInvites([]);
    }
  }

  function setHidden(element, hidden) {
    if (!element) return;
    element.classList.toggle("is-hidden", !!hidden);
  }

  function getQueryParam(name) {
    try {
      var params = new URLSearchParams(window.location.search || "");
      return String(params.get(name) || "").trim();
    } catch (e) {
      return "";
    }
  }

  function clearQueryParam(name) {
    try {
      var url = new URL(window.location.href);
      url.searchParams.delete(name);
      window.history.replaceState({}, document.title, url.pathname + url.search);
    } catch (e) {}
  }

  function persistView(view) {
    if (!storage) return;
    try {
      storage.setItem(viewKey, JSON.stringify(view || {}));
    } catch (e) {}
  }

  function readView() {
    if (!storage) return null;
    try {
      var raw = storage.getItem(viewKey);
      if (!raw) return null;
      return JSON.parse(raw);
    } catch (e) {
      return null;
    }
  }

  function persistEmpresaContext(empresaID) {
    var id = Number(empresaID || 0);
    if (!Number.isFinite(id) || id <= 0) {
      return;
    }
    var value = String(Math.trunc(id));
    try {
      sessionStorage.setItem("active_empresa_id", value);
      sessionStorage.setItem("empresa_id", value);
      sessionStorage.setItem("admin_empresa_id", value);
    } catch (e) {}
    try {
      localStorage.setItem("active_empresa_id", value);
      localStorage.setItem("empresa_id", value);
      localStorage.setItem("admin_empresa_id", value);
    } catch (e) {}
  }

  function recordSelectorAuditEvent(action, metadata) {
    try {
      var body = Object.assign({
        accion: action,
        modulo: "selector_empresa_ui",
        recurso: "seleccionar_empresa",
        endpoint: window.location.pathname,
        metadata: Object.assign({
          path: window.location.pathname,
          view: (metadata && metadata.view) || ""
        }, metadata || {})
      }, {});
      var empresaId = Number((metadata && metadata.empresa_id) || 0);
      if (Number.isFinite(empresaId) && empresaId > 0) {
        body.empresa_id = Math.trunc(empresaId);
      }
      fetch("/super/api/auditoria?action=ui_event&scope=principal", {
        method: "POST",
        credentials: "same-origin",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body)
      }).catch(function () {});
    } catch (e) {}
  }

  function clearEmpresaContextIfMatches(empresaID) {
    var id = String(Math.trunc(Number(empresaID || 0)));
    if (!id || id === "0") return;
    var keys = ["active_empresa_id", "empresa_id", "admin_empresa_id"];
    var sources = [];
    try { sources.push(window.sessionStorage); } catch (e) {}
    try { sources.push(window.localStorage); } catch (e) {}
    sources.forEach(function (source) {
      try {
        keys.forEach(function (key) {
          if (String(source.getItem(key) || "") === id) {
            source.removeItem(key);
          }
        });
      } catch (e) {}
    });
  }

  function clearConfigurationAssistantPending(empresaId) {
    if (!empresaId) return;
    var key = "pcs_config_assistant_pending_" + String(empresaId);
    try { window.localStorage.removeItem(key); } catch (e) {}
    try { window.sessionStorage.removeItem(key); } catch (e) {}
  }

  function setConfigurationAssistantPending(empresaId, metadata) {
    if (!empresaId) return;
    var key = "pcs_config_assistant_pending_" + String(empresaId);
    var payload = JSON.stringify({
      empresa_id: Number(empresaId) || empresaId,
      metadata: metadata || {},
      created_at: new Date().toISOString()
    });
    try { window.localStorage.setItem(key, payload); } catch (e) {}
    try { window.sessionStorage.setItem(key, payload); } catch (e) {}
  }

  function readEmpresaContext() {
    var sources = [window.sessionStorage, window.localStorage];
    for (var i = 0; i < sources.length; i += 1) {
      var source = sources[i];
      if (!source) continue;
      try {
        var value = parseInt(source.getItem("active_empresa_id") || source.getItem("empresa_id") || "0", 10);
        if (Number.isFinite(value) && value > 0) {
          return value;
        }
      } catch (e) {}
    }
    return 0;
  }

  function resolveEmpresaIdForMenu() {
    var stored = readEmpresaContext();
    if (stored > 0) {
      return stored;
    }
    if (Array.isArray(currentEmpresas) && currentEmpresas.length) {
      return Number(currentEmpresas[0].id || 0);
    }
    return 0;
  }

  function getEmpresaFromCurrentList(empresaId) {
    var normalizedId = Number(empresaId || 0);
    if (!normalizedId) {
      return null;
    }
    for (var i = 0; i < currentEmpresas.length; i += 1) {
      if (Number(currentEmpresas[i].id || 0) === normalizedId) {
        return currentEmpresas[i];
      }
    }
    return null;
  }

  function normalizeEmail(value) {
    return String(value || "").trim().toLowerCase();
  }

  function getAccountAdmin(account) {
    if (!account) {
      return null;
    }
    if (account.admin && typeof account.admin === "object") {
      return account.admin;
    }
    return account;
  }

  function isPrincipalSuperAccount(account) {
    var admin = getAccountAdmin(account);
    if (!admin) {
      return false;
    }
    var email = normalizeEmail(account.email || admin.email);
    var creator = normalizeEmail(admin.usuario_creador);
    var role = normalizeEmail(account.role || admin.role);
    if (role !== "super_administrador") {
      return false;
    }
    return !creator || creator === email;
  }

  function canManageScopedLicencias(account) {
    var admin = getAccountAdmin(account);
    if (!admin) {
      return false;
    }
    var role = normalizeEmail(account.role || admin.role);
    return role === "super_administrador" || role === "administrador";
  }

  function canManageScopedAdministradores(account) {
    var admin = getAccountAdmin(account);
    if (!admin) {
      return false;
    }
    if (isPrincipalSuperAccount(account)) {
      return true;
    }
    var role = normalizeEmail(account.role || admin.role);
    var email = normalizeEmail(account.email || admin.email);
    var creator = normalizeEmail(admin.usuario_creador);
    return role === "administrador" && email && (!creator || creator === email);
  }

  function isSidebarLinkVisible(link) {
    if (!link) {
      return false;
    }
    var listItem = link.closest ? link.closest("li") : null;
    var target = listItem || link;
    return target.style.display !== "none";
  }

  function setElementVisible(element, visible) {
    if (!element) return;
    var listItem = element.closest ? element.closest("li") : null;
    if (listItem) {
      listItem.style.display = visible ? "" : "none";
    } else {
      element.style.display = visible ? "" : "none";
    }
  }

  function applySidebarPermissions(account) {
    var linkLicencias = document.getElementById("linkLicencias");
    var linkMisClientes = document.getElementById("linkMisClientes");
    var linkAdministradores = document.getElementById("linkAdministradores");
    var linkAuditoriaGlobal = document.getElementById("linkAuditoriaGlobal");
    var linkReportes = document.getElementById("linkReportesGlobales");
    var principalSuper = isPrincipalSuperAccount(account);
    setElementVisible(linkLicencias, canManageScopedLicencias(account));
    setElementVisible(linkMisClientes, false);
    setElementVisible(linkAdministradores, canManageScopedAdministradores(account));
    setElementVisible(linkAuditoriaGlobal, canManageScopedAdministradores(account) || principalSuper);
    setElementVisible(linkReportes, principalSuper);
  }

  function refreshAsesorComercialLink() {
    var linkMisClientes = document.getElementById("linkMisClientes");
    if (!linkMisClientes) return Promise.resolve(false);
    return fetchJSON("/api/asesor_comercial/mis_clientes", { credentials: "same-origin" })
      .then(function (data) {
        var isAsesor = !!(data && data.is_asesor);
        setElementVisible(linkMisClientes, isAsesor);
        return isAsesor;
      })
      .catch(function () {
        setElementVisible(linkMisClientes, false);
        return false;
      });
  }

  function fetchCurrentAccount() {
    return fetch("/me", { credentials: "same-origin" })
      .then(function (res) {
        if (!res.ok) {
          throw new Error("No se pudo obtener la cuenta actual");
        }
        return res.json();
      })
      .then(function (data) {
        currentAccount = data || null;
        applySidebarPermissions(currentAccount);
        refreshAsesorComercialLink();
        return currentAccount;
      })
      .catch(function () {
        currentAccount = null;
        applySidebarPermissions(null);
        return null;
      });
  }

  function setActiveNav(activeLink) {
    navLinks.forEach(function (link) {
      link.classList.remove("active");
    });
    if (activeLink) activeLink.classList.add("active");
  }

  function openInRightFrame(href, link) {
    if (!href) return;
    var normalized = normalizeHref(href);
    if (!normalized) return;
    if (link && !isSidebarLinkVisible(link)) {
      return;
    }
    if (!contentFrame || !empresasPanel) {
      window.location.href = normalized;
      return;
    }
    setHidden(empresasPanel, true);
    setHidden(contentFrame, false);
    contentFrame.setAttribute("src", normalized);
    persistView({ mode: "frame", href: normalized });
    setActiveNav(link);
    recordSelectorAuditEvent("abrir_panel_lateral", { href: normalized, link_id: link && link.id ? link.id : "", view: "frame" });
  }

  function normalizeCompanyTypeName(value) {
    var normalized = String(value || "").trim().toLowerCase();
    if (typeof normalized.normalize === "function") {
      normalized = normalized.normalize("NFD").replace(/[\u0300-\u036f]/g, "");
    }
    return normalized;
  }

  function getVerticalTone(item) {
    var module = String(item && item.module ? item.module : "").trim().toLowerCase();
    if (/(agencia_viajes|operador_turistico)/.test(module)) return "lodging";
    if (/(eventos_boleteria|parque_recreativo)/.test(module)) return "digital";
    if (/(clinica_consultorios|laboratorio_clinico|veterinaria_petshop|salon_spa)/.test(module)) return "services";
    if (/(colegio_academia|guarderia_infantil|capacitacion_empresarial|club_deportivo)/.test(module)) return "generic";
    if (/(transporte_carga_tms)/.test(module)) return "logistics";
    if (/(lavanderia_tintoreria|taller_mecanico|servicios_tecnicos|seguridad_privada|funeraria_exequial)/.test(module)) return "services";
    if (/(inmobiliaria_comercial|cooperativa_fondo)/.test(module)) return "retail";
    return "generic";
  }

  function getVerticalByTypeName(tipoNombre) {
    var normalized = normalizeCompanyTypeName(tipoNombre);
    if (!normalized) return null;
    for (var i = 0; i < nuevasPlantillasCatalog.length; i += 1) {
      var item = nuevasPlantillasCatalog[i];
      var moduleText = normalizeCompanyTypeName(String(item && item.module ? item.module : "").replace(/_/g, " "));
      var titleText = normalizeCompanyTypeName([
        item && item.title,
        item && item.fullTitle,
        item && item.summary,
        item && item.lead
      ].join(" "));
      if ((moduleText && normalized.indexOf(moduleText) >= 0) || (titleText && (titleText.indexOf(normalized) >= 0 || normalized.indexOf(titleText.split(" ").slice(0, 2).join(" ")) >= 0))) {
        return item;
      }
    }
    return null;
  }

  function getVerticalSections(item) {
    return Array.isArray(item && item.sections)
      ? item.sections.map(function (section) { return String(section || "").trim(); }).filter(Boolean)
      : [];
  }

  function renderTipoEmpresaPreview() {
    var preview = document.getElementById("tipoEmpresaPreview");
    var select = document.getElementById("tipo_id");
    if (!preview || !select) return;
    var option = select.options ? select.options[select.selectedIndex] : null;
    var tipoNombre = option ? String(option.text || "").trim() : "";
    var plantilla = getVerticalByTypeName(tipoNombre);
    if (!plantilla) {
      preview.hidden = true;
      preview.innerHTML = "";
      return;
    }
    var sections = getVerticalSections(plantilla).slice(0, 4);
    preview.hidden = false;
    preview.innerHTML =
      '<div class="selector-type-preview__media">' +
      '<img src="' + escapeHtml(plantilla.icon || "/img/company-briefcase-color.svg") + '" alt="">' +
      "</div>" +
      '<div class="selector-type-preview__body">' +
      '<div class="selector-type-preview__kicker">Plantilla profesional</div>' +
      '<strong>' + escapeHtml(plantilla.fullTitle || plantilla.title || tipoNombre) + "</strong>" +
      '<p>' + escapeHtml(plantilla.lead || plantilla.summary || "Tipo de empresa con preconfiguracion, licencias y modulo operativo propio.") + "</p>" +
      (sections.length ? '<div class="selector-type-preview__chips">' + sections.map(function (section) {
        return '<span>' + escapeHtml(section) + '</span>';
      }).join("") + "</div>" : "") +
      "</div>";
  }

  function getEmpresaTypeVisual(empresa) {
    var tipoNombre = String(empresa && empresa.tipo_nombre ? empresa.tipo_nombre : "").trim();
    var plantilla = getVerticalByTypeName(tipoNombre);
    if (plantilla) {
      return {
        tone: getVerticalTone(plantilla),
        icon: plantilla.icon || "/img/company-briefcase-color.svg",
        alt: "Icono de " + (plantilla.title || "plantilla empresarial"),
        eyebrow: "Plantilla profesional",
        activeCopy: "Empresa lista para operar " + (plantilla.fullTitle || plantilla.title || tipoNombre) + " con flujo, roles, permisos, licencias y reportes del motor de plantilla.",
        pendingCopy: "Activa una licencia de plantilla para habilitar " + (plantilla.fullTitle || plantilla.title || tipoNombre) + " con configuracion inicial, roles y ruta de trabajo.",
        label: plantilla.fullTitle || plantilla.title || tipoNombre || "Plantilla"
      };
    }
    var normalized = normalizeCompanyTypeName(tipoNombre);
    var visualRules = [
      {
        pattern: /(hotel|hostal|hosped|apartahotel|resort|alojamiento)/,
        tone: "lodging",
        icon: "/img/hotel-logo.svg",
        alt: "Logo de hotel",
        eyebrow: "Operacion hotelera",
        activeCopy: "Gestion preparada para reservas, recepcion, estaciones, tarifas por dia y seguimiento operativo.",
        pendingCopy: "Activa la licencia para gestionar hospedaje, recepcion y trazabilidad por estacion."
      },
      {
        pattern: /(\bbar\b|\bpub\b|licoreria|discoteca|coctel|bebida|barra)/,
        tone: "food",
        icon: "/img/bar-logo.svg",
        alt: "Logo de bar",
        eyebrow: "Barra y eventos",
        activeCopy: "Operacion lista para bebidas, mesas, barra, eventos, consumos y caja.",
        pendingCopy: "Configura la licencia para activar control de barra, mesas, inventario y cobros."
      },
      {
        pattern: /(gimnasio|gym|fitness|entrenamiento|deporte)/,
        tone: "health",
        icon: "/img/gym-logo.svg",
        alt: "Logo de gimnasio",
        eyebrow: "Planes y socios",
        activeCopy: "Gestion lista para membresias, socios, entrenadores, clases y control de accesos.",
        pendingCopy: "Activa la licencia para operar planes, renovaciones, clases y recaudo de socios."
      },
      {
        pattern: /(odontologia|odontologico|dental|dentista|consultorio dental)/,
        tone: "health",
        icon: "/img/dental-logo.svg",
        alt: "Logo de odontologia",
        eyebrow: "Atencion clinica",
        activeCopy: "Gestion lista para pacientes, agenda, tratamientos, presupuestos y recaudo por consulta.",
        pendingCopy: "Activa la licencia para organizar consultorios, profesionales y tratamientos odontologicos."
      },
      {
        pattern: /(turno|turnos|fila|colas|ticket|llamado|atencion al cliente)/,
        tone: "queue",
        icon: "/img/turnos-logo.svg",
        alt: "Logo de manejo de turnos",
        eyebrow: "Turnos y llamados",
        activeCopy: "Operacion preparada para emitir turnos, llamar clientes y monitorear puestos de atencion.",
        pendingCopy: "Activa la licencia para ordenar servicios, puestos, pantalla publica y flujo de atencion."
      },
      {
        pattern: /(vehiculo|vehiculos|flota|flotas|parqueadero|transporte|automotor)/,
        tone: "fleet",
        icon: "/img/vehiculos-flotas-logo.svg",
        alt: "Logo de vehiculos y flotas",
        eyebrow: "Vehiculos y flotas",
        activeCopy: "Empresa lista para registro de vehiculos, hoja de vida, mantenimientos, alertas y permanencia.",
        pendingCopy: "Activa la licencia para gestionar flota, historiales, mantenimientos y control operativo."
      },
      {
        pattern: /(pyme|pymes|empresa general|microempresa|negocio general)/,
        tone: "generic",
        icon: "/img/pymes-logo.svg",
        alt: "Logo de pymes",
        eyebrow: "Gestion empresarial",
        activeCopy: "Empresa lista para venta directa, inventario, servicios, usuarios y control administrativo.",
        pendingCopy: "Activa la licencia para administrar ventas, catalogo, caja y operaciones de la pyme."
      },
      {
        pattern: /(sensor|sensores|monitoreo|raspberry|iot|acceso|control electrico|domotica|puerta|alarma)/,
        tone: "sensor",
        icon: "/img/sensores-logo.svg",
        alt: "Logo de sensores y monitoreo",
        eyebrow: "Sensores y monitoreo",
        activeCopy: "Operacion lista para sensores, accesos, dispositivos, alertas y monitoreo en tiempo real.",
        pendingCopy: "Activa la licencia para conectar dispositivos, controlar accesos y auditar eventos."
      },
      {
        pattern: /(agencia|marketing|publicidad|digital|red social|redes sociales|contenido|media|estudio creativo|creador)/,
        tone: "digital",
        icon: "/img/redes-sociales-logo.svg",
        alt: "Logo de redes sociales",
        eyebrow: "Canales y servicios digitales",
        activeCopy: "Negocio listo para organizar clientes, tareas, cobros y seguimiento comercial de servicios digitales.",
        pendingCopy: "Configura la licencia para convertir esta cuenta en un centro operativo de servicios digitales."
      },
      {
        pattern: /(salon|belleza|spa|estetica|peluqueria|barberia|manicure|cosmetica)/,
        tone: "services",
        icon: "/img/salon-belleza-logo.svg",
        alt: "Logo de salon de belleza",
        eyebrow: "Agenda y servicios",
        activeCopy: "Empresa lista para servicios, agenda, estilistas, comisiones y cobro por atencion.",
        pendingCopy: "Activa la licencia para organizar sillas, agenda, servicios y comisiones del salon."
      },
      {
        pattern: /(lavadero|lavado|autolavado|car wash|lavanderia de autos)/,
        tone: "services",
        icon: "/img/lavadero-autos-logo.svg",
        alt: "Logo de lavadero de autos",
        eyebrow: "Lavado y vehiculos",
        activeCopy: "Operacion lista para bahias, servicios de lavado, vehiculos, tiempos y comisiones.",
        pendingCopy: "Activa la licencia para controlar bahias, servicios, tiempos y recaudo de lavadero."
      },
      {
        pattern: /(taller|mecanico|mecanica|reparacion|mantenimiento automotriz)/,
        tone: "services",
        icon: "/img/taller-mecanico-logo.svg",
        alt: "Logo de taller mecanico",
        eyebrow: "Ordenes de servicio",
        activeCopy: "Empresa lista para bahias, tecnicos, ordenes, repuestos, comisiones y trazabilidad.",
        pendingCopy: "Activa la licencia para gestionar servicios mecanicos, vehiculos, tecnicos y cobros."
      },
      {
        pattern: /(restaurante|restaurant|bar|cafe|cafeteria|panaderia|pasteleria|comida|pizzeria|licoreria|gastro)/,
        tone: "food",
        icon: "/img/restaurante.png",
        alt: "Icono de restaurante",
        eyebrow: "Atencion gastronomica",
        activeCopy: "Operación lista para atender clientes, registrar consumos y administrar cobros del negocio.",
        pendingCopy: "Configura la licencia para activar una operación ágil de mesas, pedidos y facturación del local."
      },
      {
        pattern: /(hotel|hostal|hosped|motel|apartahotel|resort|alojamiento)/,
        tone: "lodging",
        icon: "/img/motel.png",
        alt: "Icono de hotel o motel",
        eyebrow: "Operación de hospedaje",
        activeCopy: "Gestion preparada para reservas, recepcion, estaciones y seguimiento operativo por estancia.",
        pendingCopy: "Activa la licencia para gestionar hospedaje, recepción y trazabilidad comercial por habitación."
      },
      {
        pattern: /(tienda|almacen|supermercado|market|boutique|farmacia|drogueria|minimercado|retail|comercio|ferreteria|papeleria|pos|punto de venta)/,
        tone: "retail",
        icon: "/img/punto_venta.png",
        alt: "Icono de punto de venta",
        eyebrow: "Comercio y mostrador",
        activeCopy: "Empresa lista para ventas de mostrador, control comercial e interaccion directa con clientes.",
        pendingCopy: "Habilita la licencia para operar catálogo, facturación y flujo comercial en punto de venta."
      },
      {
        pattern: /(bodega|distribuidora|logistica|almacenamiento|inventario|deposito|suministros|mayorista|warehouse)/,
        tone: "logistics",
        icon: "/img/warehouse-color.svg",
        alt: "Icono de bodega o logistica",
        eyebrow: "Control de inventario",
        activeCopy: "Preparada para movimientos de bodega, control de existencias y operación logística por empresa.",
        pendingCopy: "Activa la licencia para orquestar inventario, entradas, salidas y control de almacenamiento."
      },
      {
        pattern: /(agencia|marketing|publicidad|digital|red social|contenido|media|estudio creativo|creador)/,
        tone: "digital",
        icon: "/img/red%20social.png",
        alt: "Icono de negocio digital",
        eyebrow: "Canales y servicios digitales",
        activeCopy: "Negocio listo para organizar clientes, tareas, cobros y seguimiento comercial de servicios digitales.",
        pendingCopy: "Configura la licencia para convertir esta cuenta en un centro operativo de servicios digitales."
      },
      {
        pattern: /(tecnico|tecnica|independiente|servicio|servicios|consultoria|asesoria|salud|belleza|spa|lavanderia|taller|mantenimiento|soporte|laboratorio)/,
        tone: "services",
        icon: "/img/tecnico%20independiente.png",
        alt: "Icono de servicio profesional",
        eyebrow: "Servicio profesional",
        activeCopy: "Empresa lista para agenda, atencion al cliente, seguimiento del servicio y control de cobro.",
        pendingCopy: "Activa la licencia para centralizar clientes, sesiones y trazabilidad del servicio profesional."
      }
    ];

    var fallback = {
      tone: "generic",
      icon: "/img/company-briefcase-color.svg",
      alt: "Icono de empresa",
      eyebrow: "Operación empresarial",
      activeCopy: "Empresa disponible para continuar la gestion administrativa y operativa desde el panel principal.",
      pendingCopy: "Configura la licencia para habilitar la operación completa de esta empresa dentro del sistema."
    };

    for (var i = 0; i < visualRules.length; i += 1) {
      if (visualRules[i].pattern.test(normalized)) {
        return {
          tone: visualRules[i].tone,
          icon: visualRules[i].icon,
          alt: visualRules[i].alt,
          eyebrow: visualRules[i].eyebrow,
          activeCopy: visualRules[i].activeCopy,
          pendingCopy: visualRules[i].pendingCopy,
          label: tipoNombre || "Empresa"
        };
      }
    }

    return {
      tone: fallback.tone,
      icon: fallback.icon,
      alt: fallback.alt,
      eyebrow: fallback.eyebrow,
      activeCopy: fallback.activeCopy,
      pendingCopy: fallback.pendingCopy,
      label: tipoNombre || "Empresa general"
    };
  }

  function buildEmpresaCardDescription(empresa, visual, hasLicense) {
    var observaciones = String(empresa && empresa.observaciones ? empresa.observaciones : "").trim();
    if (observaciones) return observaciones;
    return hasLicense ? visual.activeCopy : visual.pendingCopy;
  }

  function buildEmpresaAccessLabel(empresa) {
    var accessSource = String(empresa && empresa.access_source ? empresa.access_source : "owner").toLowerCase();
    if (accessSource === "shared") {
      var compartidaPor = String(empresa && empresa.compartida_por ? empresa.compartida_por : "").trim();
      return compartidaPor ? "Compartida por " + compartidaPor : "Empresa compartida contigo";
    }
    if (accessSource === "delegated") {
      var principal = String(empresa && empresa.compartida_por ? empresa.compartida_por : "").trim();
      return principal ? "Administracion delegada por " + principal : "Administracion delegada";
    }
    return "Empresa propia";
  }

  function getEmpresaAccessSource(empresa) {
    return String(empresa && empresa.access_source ? empresa.access_source : "owner").toLowerCase();
  }

  function isSharedEmpresa(empresa) {
    return getEmpresaAccessSource(empresa) === "shared";
  }

  function isOwnerEmpresa(empresa) {
    return getEmpresaAccessSource(empresa) === "owner";
  }

  function selectorTruthy(value) {
    return value === true || value === 1 || value === "1" || String(value || "").toLowerCase() === "true";
  }

  function canSharedEmpresaReshare(empresa) {
    return isSharedEmpresa(empresa) && selectorTruthy(empresa && (empresa.shared_puede_compartir || empresa.puede_compartir));
  }

  function canShareEmpresa(empresa) {
    return isOwnerEmpresa(empresa) || isPrincipalSuperAccount(currentAccount) || canSharedEmpresaReshare(empresa);
  }

  function navigateToEmpresa(empresa, hasLicense) {
    persistEmpresaContext(empresa.id);
    recordSelectorAuditEvent(hasLicense ? "abrir_empresa" : "abrir_licencias_empresa", {
      empresa_id: Number(empresa && empresa.id ? empresa.id : 0),
      empresa_nombre: empresa && empresa.nombre ? empresa.nombre : "",
      access_source: empresa && empresa.access_source ? empresa.access_source : "",
      licencia_activa: !!hasLicense,
      view: "empresa_card"
    });
    if (hasLicense) {
      var adminURL =
        "/administrar_empresa.html?id=" + encodeURIComponent(empresa.id) +
        "&empresa_id=" + encodeURIComponent(empresa.id);
      window.location.href = adminURL;
      return;
    }
    navigateToLicenciasEmpresa(empresa);
  }

  function buildEmpresaDownloadURL(empresa) {
    var params = new URLSearchParams();
    if (empresa && empresa.id) {
      params.set("empresa_id", String(empresa.id));
      params.set("id", String(empresa.id));
    }
    if (empresa && empresa.nombre) params.set("empresa_nombre", empresa.nombre);
    return "/descargar_informacion_de_la_empresa.html?" + params.toString();
  }

  function openSelectorDeleteDownload(empresa, allowSameWindowFallback) {
    if (!empresa) return false;
    deleteModalState.descargaOfrecida = true;
    var url = buildEmpresaDownloadURL(empresa);
    var opened = false;
    try {
      var popup = window.open(url, "_blank");
      if (popup) {
        try { popup.opener = null; } catch (openerError) {}
        opened = true;
      }
    } catch (e) {
      opened = false;
    }
    if (!opened && allowSameWindowFallback) {
      window.location.href = url;
      return true;
    }
    return opened;
  }

  function confirmSelectorDownloadBeforeDelete(empresa) {
    var wantsDownload = window.confirm(
      "Antes de eliminar esta empresa, deseas descargar toda su informacion?\n\n" +
      "Aceptar: abre la descarga en una nueva pestana y luego continua la eliminacion.\n" +
      "Cancelar: continua la eliminacion sin descargar ahora."
    );
    if (!wantsDownload) return true;
    if (!openSelectorDeleteDownload(empresa, false)) {
      setSelectorDeleteMessage("No se pudo abrir la descarga automaticamente. Usa el boton Descargar informacion antes de eliminar y vuelve a confirmar.", true);
      return false;
    }
    setSelectorDeleteMessage("Se abrio la descarga. Conserva el archivo; la eliminacion continuara.", false);
    return true;
  }

  function buildLicenciasEmpresaURL(empresa) {
    var params = new URLSearchParams();
    if (empresa && empresa.id) {
      params.set("empresa_id", empresa.id);
      params.set("id", empresa.id);
    }
    if (empresa && empresa.tipo_id) params.set("tipo_id", empresa.tipo_id);
    if (empresa && empresa.tipo_nombre) params.set("tipo_nombre", empresa.tipo_nombre);
    return "/elegir_licencia.html?" + params.toString();
  }

  function navigateToLicenciasEmpresa(empresa) {
    var empresaId = empresa && empresa.id ? empresa.id : "";
    if (!empresaId) {
      window.alert("No se encontro la empresa para elegir licencia.");
      return;
    }
    persistEmpresaContext(empresaId);
    window.location.href = buildLicenciasEmpresaURL(empresa);
  }

  function buildEmpresaShareButton(empresa) {
    var disabled = !canShareEmpresa(empresa);
    var title = disabled
      ? "Solo el propietario, super administrador o administrador compartido autorizado puede compartir esta empresa"
      : "Compartir empresa con otro administrador (correo)";
    return '' +
      '<button type="button" class="empresa-card-icon-action empresa-share-toggle' + (disabled ? ' is-disabled' : '') + '" data-empresa-id="' + escapeHtml(String(empresa.id || '')) + '" data-share-disabled="' + (disabled ? '1' : '0') + '" aria-label="' + escapeHtml(title) + '" title="' + escapeHtml(title) + '">' +
      '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="18" height="18" aria-hidden="true" focusable="false">' +
      '<circle cx="18" cy="5" r="2.6" fill="none" stroke="currentColor" stroke-width="2"></circle>' +
      '<circle cx="6" cy="12" r="2.6" fill="none" stroke="currentColor" stroke-width="2"></circle>' +
      '<circle cx="18" cy="19" r="2.6" fill="none" stroke="currentColor" stroke-width="2"></circle>' +
      '<path d="M8.3 11.1L15.7 6.9M8.3 12.9L15.7 17.1" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"></path>' +
      '</svg>' +
      '</button>';
  }

  function buildEmpresaSharePanel(empresa) {
    if (!canShareEmpresa(empresa)) {
      return '' +
        '<div class="empresa-card-share-panel" data-empresa-id="' + escapeHtml(String(empresa.id || '')) + '" hidden>' +
        '<div class="empresa-card-share-feedback is-error">Solo el propietario, super administrador o administrador compartido autorizado puede enviar nuevas invitaciones para esta empresa.</div>' +
        '</div>';
    }
    return '' +
      '<div class="empresa-card-share-panel" data-empresa-id="' + escapeHtml(String(empresa.id || '')) + '" hidden>' +
      '<form class="empresa-card-share-form" data-empresa-id="' + escapeHtml(String(empresa.id || '')) + '">' +
      '<label class="empresa-card-share-label" for="share-email-' + escapeHtml(String(empresa.id || '')) + '">Enviar correo para compartir empresa</label>' +
      '<div class="empresa-card-share-row">' +
      '<input id="share-email-' + escapeHtml(String(empresa.id || '')) + '" class="form-input empresa-card-share-input" data-share-email type="email" placeholder="correo@ejemplo.com" required>' +
      '<button type="submit" class="btn empresa-card-share-submit">Enviar</button>' +
      '</div>' +
      '<label class="empresa-card-share-label" for="share-level-' + escapeHtml(String(empresa.id || '')) + '">Rol de acceso</label>' +
      '<select id="share-level-' + escapeHtml(String(empresa.id || '')) + '" class="form-input empresa-card-share-level" data-share-level>' +
      '<option value="solo_ver">Solo ver</option>' +
      '<option value="acceso_total">Acceso total</option>' +
      '<option value="modulos">Solo ciertos modulos</option>' +
      '</select>' +
      '<div class="empresa-card-share-modules" data-share-modules hidden>' + buildEmpresaCardShareModules() + '</div>' +
      '<label class="empresa-card-share-option">' +
      '<input type="checkbox" data-share-can-reshare>' +
      '<span>Permitir que este administrador tambien pueda compartir esta empresa</span>' +
      '</label>' +
      '<div class="empresa-card-share-hint">Si lo activas, el nuevo administrador podra invitar a otros administradores a esta misma empresa. Si queda apagado, solo tendra el acceso asignado.</div>' +
      '<div class="empresa-card-share-feedback" data-share-feedback role="status"></div>' +
      '</form>' +
      '</div>';
  }

  function buildEmpresaCardShareModules() {
    return empresaShareModuleCatalog.map(function (item) {
      return '<label class="empresa-card-share-module">'
        + '<input type="checkbox" value="' + escapeHtml(item[0]) + '" data-share-module>'
        + '<span>' + escapeHtml(item[1]) + '</span>'
        + '</label>';
    }).join('');
  }

  function normalizeEmpresaShareNivel(value) {
    value = String(value || '').trim().toLowerCase();
    if (value === 'modulos') return 'modulos';
    if (value === 'acceso_total') return 'acceso_total';
    return 'solo_ver';
  }

  function updateEmpresaCardShareScope(panel) {
    if (!panel) return;
    var level = normalizeEmpresaShareNivel(panel.querySelector('[data-share-level]') && panel.querySelector('[data-share-level]').value);
    var modules = panel.querySelector('[data-share-modules]');
    if (modules) modules.hidden = level !== 'modulos';
  }

  function getEmpresaCardSelectedShareModules(form) {
    return Array.prototype.map.call(form.querySelectorAll('[data-share-module]:checked'), function (input) {
      return String(input.value || '').trim();
    }).filter(Boolean);
  }

  function findEmpresaSharePanel(empresaId) {
    return document.querySelector('.empresa-card-share-panel[data-empresa-id="' + String(empresaId || '') + '"]');
  }

  function findEmpresaShareButton(empresaId) {
    return document.querySelector('.empresa-share-toggle[data-empresa-id="' + String(empresaId || '') + '"]');
  }

  function setEmpresaShareFeedback(empresaId, text, isError) {
    var panel = findEmpresaSharePanel(empresaId);
    if (!panel) return;
    var feedback = panel.querySelector('[data-share-feedback]');
    if (!feedback) return;
    // limpiar acciones previas (ej. bot?n reenviar)
    Array.prototype.forEach.call(feedback.querySelectorAll('button.empresa-share-resend'), function (btn) {
      try { btn.remove(); } catch (e) {}
    });
    feedback.textContent = text || '';
    feedback.classList.toggle('is-error', !!isError && !!text);
    feedback.classList.toggle('is-success', !isError && !!text);
  }

  function showEmpresaShareResendAction(empresaId, invitationId) {
    var panel = findEmpresaSharePanel(empresaId);
    if (!panel) return;
    var feedback = panel.querySelector('[data-share-feedback]');
    if (!feedback) return;
    // asegurar separaci?n visual si ya hay texto
    if (feedback.textContent) {
      feedback.textContent = String(feedback.textContent).trim() + ' ';
    }
    var btn = document.createElement('button');
    btn.type = 'button';
    btn.className = 'btn secondary empresa-share-resend';
    btn.textContent = 'Reenviar invitación';
    btn.setAttribute('data-empresa-id', String(empresaId || ''));
    btn.setAttribute('data-invitation-id', String(invitationId || ''));
    feedback.appendChild(btn);
  }

  function setEmpresaCardShareOpen(panelEl, open) {
    var card = panelEl && panelEl.closest ? panelEl.closest('.portal-card.empresa-card') : null;
    if (card) {
      card.classList.toggle('empresa-card--share-open', !!open);
    }
  }

  function closeAllEmpresaSharePanels(exceptEmpresaId) {
    Array.prototype.forEach.call(document.querySelectorAll('.empresa-card-share-panel'), function (panel) {
      var panelEmpresaId = parseInt(panel.getAttribute('data-empresa-id') || '0', 10);
      var shouldKeepOpen = exceptEmpresaId > 0 && panelEmpresaId === exceptEmpresaId;
      panel.hidden = !shouldKeepOpen;
      setEmpresaCardShareOpen(panel, shouldKeepOpen);
      var btn = findEmpresaShareButton(panelEmpresaId);
      if (btn) {
        btn.classList.toggle('is-open', shouldKeepOpen);
      }
    });
  }

  function toggleEmpresaSharePanel(empresaId) {
    var panel = findEmpresaSharePanel(empresaId);
    var btn = findEmpresaShareButton(empresaId);
    if (!panel || !btn) return;
    var willOpen = panel.hidden;
    closeAllEmpresaSharePanels(willOpen ? empresaId : 0);
    if (willOpen) {
      updateEmpresaCardShareScope(panel);
      var emailInput = panel.querySelector('[data-share-email]');
      if (emailInput) {
        window.setTimeout(function () {
          try {
            emailInput.focus();
          } catch (e) {}
        }, 30);
      }
    }
  }

  async function submitEmpresaShareInvitation(form) {
    var empresaId = parseInt(form.getAttribute('data-empresa-id') || '0', 10);
    if (!empresaId) {
      return;
    }
    var emailInput = form.querySelector('[data-share-email]');
    var levelInput = form.querySelector('[data-share-level]');
    var reshareInput = form.querySelector('[data-share-can-reshare]');
    var email = String(emailInput && emailInput.value ? emailInput.value : '').trim();
    var nivelAcceso = normalizeEmpresaShareNivel(levelInput && levelInput.value);
    var modulosPermitidos = nivelAcceso === 'modulos' ? getEmpresaCardSelectedShareModules(form) : [];
    var puedeCompartir = !!(reshareInput && reshareInput.checked);
    if (!email) {
      setEmpresaShareFeedback(empresaId, 'Debes escribir el correo del otro administrador.', true);
      return;
    }
    if (nivelAcceso === 'modulos' && !modulosPermitidos.length) {
      setEmpresaShareFeedback(empresaId, 'Elige al menos un modulo para compartir solo ciertos modulos.', true);
      return;
    }
    setEmpresaShareFeedback(empresaId, 'Enviando invitación...', false);
    try {
      var data = await fetchJSON('/super/api/empresas/compartidos', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          empresa_id: empresaId,
          email: email,
          mensaje: '',
          nivel_acceso: nivelAcceso,
          modulos_permitidos: modulosPermitidos,
          puede_compartir: puedeCompartir
        })
      });
      var message = data && data.message ? data.message : 'Invitación enviada correctamente.';
      setEmpresaShareFeedback(empresaId, message, false);
      setShareNotice(message, false);
      if (emailInput) {
        emailInput.value = '';
      }
      if (levelInput) levelInput.value = 'solo_ver';
      Array.prototype.forEach.call(form.querySelectorAll('[data-share-module]'), function (input) {
        input.checked = false;
      });
      if (reshareInput) reshareInput.checked = false;
      updateEmpresaCardShareScope(form.closest('.empresa-card-share-panel'));
    } catch (err) {
      var payload = err && err.payload ? err.payload : null;
      if (payload && String(payload.code || '') === 'invitation_pending' && payload.invitation_id) {
        var pendingMsg = (payload.error || err.message || 'Ya existe una invitación pendiente para ese administrador.') + ' ';
        setEmpresaShareFeedback(empresaId, pendingMsg + 'Puedes reenviarla.', true);
        showEmpresaShareResendAction(empresaId, payload.invitation_id);
        setShareNotice(pendingMsg + 'Puedes reenviarla.', true);
        return;
      }
      var errorMessage = err && err.message ? err.message : 'No se pudo enviar la invitación.';
      setEmpresaShareFeedback(empresaId, errorMessage, true);
      setShareNotice(errorMessage, true);
    }
  }

  async function handleEmpresaPreconfigDecision(createData) {
    if (!createData || !createData.preconfiguracion_aplicada || !createData.id) {
      return;
    }
    var preconfig = createData.preconfiguracion || {};
    var estaciones = Number(preconfig.estaciones_creadas || 0);
    var productos = Number(preconfig.productos_creados || 0);
    var usuarios = Number(preconfig.usuarios_creados || 0);
    var tipo = String(preconfig.tipo_empresa_nombre || "").trim();
    setConfigurationAssistantPending(createData.id, preconfig);
    setShareNotice("Empresa creada" + (tipo ? " para " + tipo : "") + " con preconfiguracion inicial: " + estaciones + " estaciones, " + productos + " productos guia y " + usuarios + " usuarios guia. Al entrar se abrira el asistente interactivo para ajustar los datos reales.", false);
  }

  function isEmpresaDragInteractiveTarget(target) {
    if (!target || !target.closest) return false;
    if (target.closest(".empresa-reorder-handle")) return false;
    return !!target.closest("button, a, input, select, textarea, .empresa-card-share-panel, .empresa-share-toggle, .empresa-license-action, .edit-empresa, .delete-empresa, .download-data");
  }

  function beginEmpresaDrag(state) {
    if (!state || !state.card || state.dragging) return;
    state.dragging = true;
    suppressEmpresaCardClickUntil = Date.now() + 700;
    document.body.classList.add("selector-empresa-drag-active");
    state.card.classList.add("is-dragging");
    state.card.setAttribute("aria-grabbed", "true");
    state.grid.classList.add("is-sorting");
  }

  function finishEmpresaDrag(save) {
    var state = selectorDragState;
    if (!state) return;
    try {
      if (state.card) {
        state.card.classList.remove("is-dragging");
        state.card.setAttribute("aria-grabbed", "false");
      }
      if (state.grid) {
        state.grid.classList.remove("is-sorting");
      }
      document.body.classList.remove("selector-empresa-drag-active");
      if (save && state.dragging) {
        saveSelectorEmpresasOrder(collectRenderedEmpresaOrder(), false);
        recordSelectorAuditEvent("ordenar_tarjetas_empresas", {
          view: "empresas",
          empresa_id: Number(state.card && state.card.getAttribute("data-empresa-id") || 0) || 0
        });
      }
    } finally {
      selectorDragState = null;
    }
  }

  function moveEmpresaCardFromPointer(evt) {
    var state = selectorDragState;
    if (!state || !state.card || !state.grid) return;
    var dx = Number(evt.clientX || 0) - state.startX;
    var dy = Number(evt.clientY || 0) - state.startY;
    if (!state.dragging && Math.sqrt(dx * dx + dy * dy) >= 10) {
      beginEmpresaDrag(state);
    }
    if (!state.dragging) return;
    evt.preventDefault();
    var element = document.elementFromPoint(Number(evt.clientX || 0), Number(evt.clientY || 0));
    var targetCard = element && element.closest ? element.closest(".empresa-card-link[data-empresa-id]") : null;
    if (!targetCard || targetCard === state.card || targetCard.parentNode !== state.grid) {
      return;
    }
    var rect = targetCard.getBoundingClientRect();
    var insertAfter = Number(evt.clientY || 0) > rect.top + rect.height / 2;
    if (insertAfter) {
      state.grid.insertBefore(state.card, targetCard.nextSibling);
    } else {
      state.grid.insertBefore(state.card, targetCard);
    }
  }

  function initializeEmpresaDragAndDrop(grid) {
    if (!grid || grid.dataset.sortableEmpresas === "1") return;
    grid.dataset.sortableEmpresas = "1";
    grid.addEventListener("pointerdown", function (evt) {
      var card = evt.target && evt.target.closest ? evt.target.closest(".empresa-card-link[data-empresa-id]") : null;
      if (!card || card.parentNode !== grid) return;
      var isHandle = !!(evt.target.closest && evt.target.closest(".empresa-reorder-handle"));
      if (evt.pointerType === "touch" && !isHandle) return;
      if (isEmpresaDragInteractiveTarget(evt.target)) return;
      selectorDragState = {
        card: card,
        grid: grid,
        startX: Number(evt.clientX || 0),
        startY: Number(evt.clientY || 0),
        dragging: false
      };
      try {
        card.setPointerCapture(evt.pointerId);
      } catch (e) {}
    });
    if (document.documentElement.dataset.selectorEmpresaDragDocumentBound !== "1") {
      document.documentElement.dataset.selectorEmpresaDragDocumentBound = "1";
      document.addEventListener("pointermove", moveEmpresaCardFromPointer, { passive: false });
      document.addEventListener("pointerup", function () {
        finishEmpresaDrag(true);
      });
      document.addEventListener("pointercancel", function () {
        finishEmpresaDrag(false);
      });
    }
  }

  function buildEmpresaCard(empresa, hasLicense) {
    var visual = getEmpresaTypeVisual(empresa);
    var descripcion = buildEmpresaCardDescription(empresa, visual, hasLicense);
    var cardLink = document.createElement("div");
    cardLink.className = "card-link empresa-card-link";
    cardLink.dataset.empresaId = String(empresa.id || "");
    cardLink.tabIndex = 0;
    cardLink.setAttribute("role", "button");
    cardLink.setAttribute("aria-grabbed", "false");
    cardLink.setAttribute("aria-label", (hasLicense ? "Administrar " : "Elegir licencia para ") + String(empresa.nombre || "empresa"));
    cardLink.addEventListener("click", function (evt) {
      if (Date.now() < suppressEmpresaCardClickUntil) {
        evt.preventDefault();
        evt.stopPropagation();
        return;
      }
      if (evt.target.closest && evt.target.closest('.empresa-reorder-handle, .empresa-share-toggle, .empresa-card-share-panel, button.download-data, .empresa-license-action, .edit-empresa, .delete-empresa')) {
        return;
      }
      try {
        navigateToEmpresa(empresa, hasLicense);
      } catch (err) {
        console.error(err);
      }
    });
    cardLink.addEventListener("keydown", function (evt) {
      if (evt.key !== 'Enter' && evt.key !== ' ') {
        return;
      }
      if (evt.target !== cardLink) {
        return;
      }
      evt.preventDefault();
      try {
        navigateToEmpresa(empresa, hasLicense);
      } catch (err) {
        console.error(err);
      }
    });

    var div = document.createElement("div");
    div.className = "portal-card warm empresa-card empresa-tone-" + visual.tone + (hasLicense ? " empresa-card--license-active" : " empresa-card--license-inactive");
    div.setAttribute('data-tone', visual.tone || 'generic');
    var licenseLabel = hasLicense ? "Elegir otra licencia" : "Elegir licencia";
    var licenseTitle = hasLicense
      ? "Elegir otra licencia para " + String(empresa.nombre || "esta empresa")
      : "Elegir licencia para " + String(empresa.nombre || "esta empresa");
    var licenseCellHTML =
      '<button type="button" class="license-indicator empresa-license-action ' + (hasLicense ? "active" : "inactive") + '" ' +
      'data-empresa-id="' + escapeHtml(String(empresa.id || "")) + '" ' +
      'aria-label="' + escapeHtml(licenseTitle) + '" title="' + escapeHtml(licenseTitle) + '">' +
      '<span>' + escapeHtml(licenseLabel) + '</span>' +
      '</button>';
    div.innerHTML =
      '<div class="empresa-card-topline">' +
      '<button type="button" class="empresa-reorder-handle" aria-label="Mover tarjeta de ' + escapeHtml(String(empresa.nombre || "empresa")) + '" title="Mover tarjeta">' +
      '<span aria-hidden="true">⋮⋮</span>' +
      "</button>" +
      '<span class="empresa-card-badge">' +
      escapeHtml(visual.label || "Empresa") +
      "</span>" +
      "</div>" +
      '<span class="empresa-card-watermark" aria-hidden="true">' +
      '<img src="' + escapeHtml(visual.icon || "/img/company-briefcase-color.svg") + '" alt="">' +
      "</span>" +
      '<div class="card-body empresa-card-body">' +
      '<h3 class="card-title">' +
      escapeHtml(empresa.nombre || "--") +
      "</h3>" +
      '<p class="empresa-shared-note">' +
      escapeHtml(buildEmpresaAccessLabel(empresa)) +
      "</p>" +
      '<p class="card-desc muted">' +
      escapeHtml(descripcion || "") +
      "</p>" +
      '<div class="empresa-card-footer-bar" role="group" aria-label="Acciones de la empresa">' +
      '<div class="empresa-card-footer-bar__cell empresa-card-footer-bar__cell--license">' +
      licenseCellHTML +
      "</div>" +
      '<div class="empresa-card-footer-bar__cell empresa-card-footer-bar__cell--download"></div>' +
      '<div class="empresa-card-footer-bar__cell empresa-card-footer-bar__cell--share">' +
      buildEmpresaShareButton(empresa) +
      "</div>" +
      "</div>" +
      buildEmpresaSharePanel(empresa) +
      "</div>";

    var dlBtn = document.createElement("button");
    dlBtn.type = "button";
    dlBtn.className = "license-indicator active download-data";
    dlBtn.setAttribute("data-empresa-id", String(empresa.id || ""));
    dlBtn.setAttribute("data-empresa-name", String(empresa.nombre || ""));
    dlBtn.setAttribute("aria-label", "Descargar información de la empresa " + String(empresa.nombre || ""));
    dlBtn.setAttribute("title", "Descargar información de la empresa");
    dlBtn.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="16" height="16" aria-hidden="true" focusable="false"><path fill="currentColor" d="M12 3v10l4-4-1.4-1.4L13 9.2V3h-2zM5 18v2h14v-2H5z"/></svg>';
    var dlCell = div.querySelector(".empresa-card-footer-bar__cell--download");
    if (dlCell) {
      dlCell.appendChild(dlBtn);
    }

    var licenseBtn = div.querySelector(".empresa-license-action");
    if (licenseBtn) {
      licenseBtn.addEventListener("click", function (ev) {
        ev.preventDefault();
        ev.stopPropagation();
        navigateToLicenciasEmpresa(empresa);
      });
    }

    var editBtn = document.createElement("button");
    editBtn.type = "button";
    editBtn.className = "license-indicator edit-empresa " + (hasLicense ? "active" : "inactive");
    editBtn.setAttribute("data-empresa-id", String(empresa.id || ""));
    editBtn.setAttribute("aria-label", (isOwnerEmpresa(empresa) ? "Editar empresa " : "Ver empresa ") + String(empresa.nombre || ""));
    editBtn.setAttribute("title", isOwnerEmpresa(empresa) ? "Editar empresa" : "Ver empresa");
    editBtn.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="16" height="16" aria-hidden="true" focusable="false"><path fill="currentColor" d="M3 17.25V21h3.75L17.8 9.94l-3.75-3.75L3 17.25zm2.92 2.83H5v-.92l9.06-9.06.92.92L5.92 20.08zM20.71 7.04a1.003 1.003 0 0 0 0-1.42l-2.34-2.34a1.003 1.003 0 0 0-1.42 0l-1.83 1.83 3.75 3.75 1.84-1.82z"/></svg>';
    editBtn.addEventListener("click", function (ev) {
      ev.preventDefault();
      ev.stopPropagation();
      var empresaId = Number(empresa && empresa.id ? empresa.id : 0);
      if (!empresaId) {
        window.alert("Primero crea o selecciona una empresa para editarla.");
        return;
      }
      persistEmpresaContext(empresaId);
      openInRightFrame(
        "/editar_empresa.html?id=" + encodeURIComponent(String(empresaId)) + "&empresa_id=" + encodeURIComponent(String(empresaId)),
        null
      );
    });
    if (dlCell) {
      dlCell.appendChild(editBtn);
    }

    cardLink.appendChild(div);
    return cardLink;
  }

  async function fetchEmpresaImpacto(empresaId) {
    var res = await fetch(
      "/super/api/empresas?id=" + encodeURIComponent(empresaId) + "&action=impacto_desactivacion",
      { credentials: "same-origin" }
    );
    var raw = await res.text();
    var data = null;
    try {
      data = raw ? JSON.parse(raw) : null;
    } catch (e) {
      data = null;
    }
    if (!res.ok) {
      throw new Error((data && (data.error || data.message)) || raw || "No se pudo obtener impacto de desactivaci?n");
    }
    return data && data.impacto ? data.impacto : null;
  }

  function formatImpactoTexto(impacto) {
    if (!impacto) return "";
    var rows = [];
    if ((impacto.usuarios_activos || 0) > 0) rows.push("- Usuarios activos: " + impacto.usuarios_activos);
    if ((impacto.carritos_abiertos || 0) > 0) rows.push("- Carritos abiertos: " + impacto.carritos_abiertos);
    if ((impacto.reservas_vigentes || 0) > 0) rows.push("- Reservas vigentes: " + impacto.reservas_vigentes);
    if ((impacto.licencias_activas || 0) > 0) rows.push("- Licencias activas: " + impacto.licencias_activas);
    return rows.join("\n");
  }

  function ensureSelectorDeleteModal() {
    var existing = document.getElementById("selectorDeleteModal");
    if (existing) return existing;
    var modal = document.createElement("div");
    modal.id = "selectorDeleteModal";
    modal.className = "selector-delete-modal";
    modal.hidden = true;
    modal.innerHTML =
      '<div class="selector-delete-modal__backdrop" data-selector-delete-close></div>' +
      '<section class="selector-delete-modal__panel" role="dialog" aria-modal="true" aria-labelledby="selectorDeleteTitle">' +
      '  <div class="selector-delete-modal__header">' +
      '    <div><p class="selector-delete-kicker">Eliminacion irreversible</p><h2 id="selectorDeleteTitle">Eliminar empresa</h2><p id="selectorDeleteSubtitle" class="form-help"></p></div>' +
      '    <button type="button" class="selector-delete-close" data-selector-delete-close aria-label="Cerrar">×</button>' +
      '  </div>' +
      '  <div id="selectorDeleteImpact" class="selector-delete-impact" aria-live="polite"></div>' +
      '  <div class="selector-delete-download-actions">' +
      '    <button id="selectorDeleteDownloadBtn" type="button" class="btn secondary">Descargar informacion antes de eliminar</button>' +
      '  </div>' +
      '  <label class="form-label" for="selectorDeleteNameInput">Escribe el nombre exacto de la empresa</label>' +
      '  <input id="selectorDeleteNameInput" class="form-input" autocomplete="off">' +
      '  <label class="form-label" for="selectorDeletePhraseInput">Escribe ELIMINAR para autorizar el borrado</label>' +
      '  <input id="selectorDeletePhraseInput" class="form-input" autocomplete="off" autocapitalize="characters" spellcheck="false">' +
      '  <label class="selector-delete-ack" for="selectorDeleteRiskInput">' +
      '    <input id="selectorDeleteRiskInput" type="checkbox">' +
      '    <span>Entiendo que se eliminaran registros, accesos compartidos, licencias, archivos y datos asociados a esta empresa.</span>' +
      '  </label>' +
      '  <div class="selector-delete-checklist" aria-label="Validaciones de eliminacion">' +
      '    <div id="selectorDeleteNameCheck" class="selector-delete-check-item">Nombre exacto pendiente</div>' +
      '    <div id="selectorDeletePhraseCheck" class="selector-delete-check-item">Frase ELIMINAR pendiente</div>' +
      '    <div id="selectorDeleteRiskCheck" class="selector-delete-check-item">Aceptacion de riesgo pendiente</div>' +
      '  </div>' +
      '  <div id="selectorDeleteProgress" class="selector-delete-progress" hidden><span class="empresa-delete-spinner" aria-hidden="true"></span><span id="selectorDeleteProgressText">Preparando eliminacion...</span></div>' +
      '  <div id="selectorDeleteMessage" class="form-help selector-delete-message" role="status"></div>' +
      '  <div class="selector-delete-actions">' +
      '    <button type="button" class="btn secondary" data-selector-delete-close>Cancelar</button>' +
      '    <button id="selectorDeleteConfirmBtn" type="button" class="btn danger">Eliminar definitivamente</button>' +
      '  </div>' +
      '</section>';
    document.body.appendChild(modal);

    Array.prototype.forEach.call(modal.querySelectorAll("[data-selector-delete-close]"), function (btn) {
      btn.addEventListener("click", closeSelectorDeleteModal);
    });
    ["selectorDeleteNameInput", "selectorDeletePhraseInput", "selectorDeleteRiskInput"].forEach(function (id) {
      var input = document.getElementById(id);
      if (!input) return;
      input.addEventListener(input.type === "checkbox" ? "change" : "input", updateSelectorDeleteChecklist);
    });
    var downloadBtn = document.getElementById("selectorDeleteDownloadBtn");
    if (downloadBtn) {
      downloadBtn.addEventListener("click", function () {
        if (!deleteModalState.empresa) return;
        openSelectorDeleteDownload(deleteModalState.empresa, true);
        setSelectorDeleteMessage("Se abrio la descarga. Conserva el archivo antes de continuar si lo necesitas.", false);
        updateSelectorDeleteChecklist();
      });
    }
    var confirmBtn = document.getElementById("selectorDeleteConfirmBtn");
    if (confirmBtn) {
      confirmBtn.addEventListener("click", deleteEmpresaFromSelector);
    }
    return modal;
  }

  function setSelectorDeleteMessage(text, isError) {
    var msg = document.getElementById("selectorDeleteMessage");
    if (!msg) return;
    msg.textContent = text || "";
    msg.classList.toggle("error", !!isError);
    msg.classList.toggle("success", !isError && !!text);
  }

  function setSelectorDeleteBusy(busy, text) {
    deleteModalState.deleting = !!busy;
    var progress = document.getElementById("selectorDeleteProgress");
    var progressText = document.getElementById("selectorDeleteProgressText");
    if (progress) progress.hidden = !busy;
    if (progressText && text) progressText.textContent = text;
    ["selectorDeleteNameInput", "selectorDeletePhraseInput", "selectorDeleteRiskInput", "selectorDeleteDownloadBtn"].forEach(function (id) {
      var el = document.getElementById(id);
      if (el) el.disabled = !!busy;
    });
    updateSelectorDeleteChecklist();
  }

  function renderSelectorDeleteImpact(impacto) {
    var target = document.getElementById("selectorDeleteImpact");
    if (!target) return;
    var data = impacto || {};
    var items = [
      ["Usuarios activos", data.usuarios_activos || 0],
      ["Carritos abiertos", data.carritos_abiertos || 0],
      ["Reservas vigentes", data.reservas_vigentes || 0],
      ["Licencias activas", data.licencias_activas || 0]
    ];
    target.innerHTML = items.map(function (item) {
      var value = Number(item[1] || 0);
      return '<div class="selector-delete-impact-chip' + (value > 0 ? ' has-risk' : '') + '"><span>' + escapeHtml(item[0]) + '</span><strong>' + escapeHtml(String(value)) + '</strong></div>';
    }).join("");
  }

  function setSelectorDeleteCheck(id, ok, okText, pendingText) {
    var el = document.getElementById(id);
    if (!el) return;
    el.classList.toggle("is-ok", !!ok);
    el.textContent = ok ? okText : pendingText;
  }

  function updateSelectorDeleteChecklist() {
    var empresa = deleteModalState.empresa || {};
    var expectedName = String(empresa.nombre || "").trim();
    var nameValue = String((document.getElementById("selectorDeleteNameInput") || {}).value || "").trim();
    var phraseValue = String((document.getElementById("selectorDeletePhraseInput") || {}).value || "").trim().toUpperCase();
    var riskOk = !!((document.getElementById("selectorDeleteRiskInput") || {}).checked);
    var nameOk = expectedName !== "" && nameValue === expectedName;
    var phraseOk = phraseValue === "ELIMINAR";
    setSelectorDeleteCheck("selectorDeleteNameCheck", nameOk, "Nombre exacto validado", "Nombre exacto pendiente");
    setSelectorDeleteCheck("selectorDeletePhraseCheck", phraseOk, "Frase ELIMINAR validada", "Frase ELIMINAR pendiente");
    setSelectorDeleteCheck("selectorDeleteRiskCheck", riskOk, "Riesgo aceptado conscientemente", "Aceptacion de riesgo pendiente");
    var btn = document.getElementById("selectorDeleteConfirmBtn");
    if (btn) {
      btn.disabled = deleteModalState.deleting || !nameOk || !phraseOk || !riskOk || isSharedEmpresa(empresa);
    }
  }

  async function openSelectorDeleteModal(empresa) {
    if (!empresa || !empresa.id) {
      setShareNotice("No se encontro la empresa para eliminar.", true);
      return;
    }
    if (isSharedEmpresa(empresa)) {
      setShareNotice("La eliminacion total solo esta disponible para el administrador propietario.", true);
      return;
    }
    var modal = ensureSelectorDeleteModal();
    deleteModalState.empresa = empresa;
    deleteModalState.impacto = null;
    deleteModalState.descargaOfrecida = false;
    deleteModalState.deleting = false;
    modal.hidden = false;
    document.body.classList.add("selector-delete-modal-open");
    document.getElementById("selectorDeleteTitle").textContent = "Eliminar " + (empresa.nombre || "empresa");
    document.getElementById("selectorDeleteSubtitle").textContent = "Esta accion elimina la empresa de forma permanente. Valida el impacto y confirma con doble seguridad.";
    document.getElementById("selectorDeleteNameInput").value = "";
    document.getElementById("selectorDeleteNameInput").placeholder = empresa.nombre || "";
    document.getElementById("selectorDeletePhraseInput").value = "";
    document.getElementById("selectorDeleteRiskInput").checked = false;
    setSelectorDeleteMessage("", false);
    renderSelectorDeleteImpact(null);
    updateSelectorDeleteChecklist();
    setSelectorDeleteBusy(true, "Consultando impacto operativo...");
    try {
      var impactoData = await fetchEmpresaImpacto(empresa.id);
      deleteModalState.impacto = impactoData || null;
      renderSelectorDeleteImpact(deleteModalState.impacto);
      setSelectorDeleteBusy(false);
      document.getElementById("selectorDeleteNameInput").focus();
    } catch (err) {
      setSelectorDeleteBusy(false);
      setSelectorDeleteMessage(err && err.message ? err.message : "No se pudo consultar el impacto de la empresa.", true);
    }
  }

  function closeSelectorDeleteModal() {
    if (deleteModalState.deleting) return;
    var modal = document.getElementById("selectorDeleteModal");
    if (modal) modal.hidden = true;
    document.body.classList.remove("selector-delete-modal-open");
    deleteModalState.empresa = null;
    deleteModalState.impacto = null;
    deleteModalState.descargaOfrecida = false;
  }

  async function deleteEmpresaFromSelector() {
    var empresa = deleteModalState.empresa;
    if (!empresa || deleteModalState.deleting) return;
    var nameValue = String((document.getElementById("selectorDeleteNameInput") || {}).value || "").trim();
    var phraseValue = String((document.getElementById("selectorDeletePhraseInput") || {}).value || "").trim().toUpperCase();
    var riskOk = !!((document.getElementById("selectorDeleteRiskInput") || {}).checked);
    updateSelectorDeleteChecklist();
    if (nameValue !== String(empresa.nombre || "").trim()) {
      setSelectorDeleteMessage("El nombre digitado no coincide exactamente.", true);
      return;
    }
    if (phraseValue !== "ELIMINAR") {
      setSelectorDeleteMessage("Debes escribir ELIMINAR para confirmar el borrado irreversible.", true);
      return;
    }
    if (!riskOk) {
      setSelectorDeleteMessage("Marca la aceptacion de riesgo antes de eliminar la empresa.", true);
      return;
    }
    if (!confirmSelectorDownloadBeforeDelete(empresa)) {
      return;
    }
    try {
      setSelectorDeleteBusy(true, "Actualizando impacto antes de eliminar...");
      var impactoData = await fetchEmpresaImpacto(empresa.id);
      deleteModalState.impacto = impactoData || deleteModalState.impacto;
      renderSelectorDeleteImpact(deleteModalState.impacto);
      setSelectorDeleteBusy(true, "Eliminando registros, licencias, accesos y archivos...");
      var data = await fetchJSON("/super/api/empresas?id=" + encodeURIComponent(empresa.id) + "&action=eliminar_total", {
        method: "DELETE",
        credentials: "same-origin",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          confirmacion_nombre: nameValue,
          confirmacion_accion: "ELIMINAR",
          confirmacion_riesgo: true,
          descarga_ofrecida: !!deleteModalState.descargaOfrecida
        })
      });
      var result = data && data.result ? data.result : {};
      var archivos = data && data.archivos ? data.archivos : {};
      var erroresArchivos = Array.isArray(archivos.errores) ? archivos.errores.length : 0;
      clearEmpresaContextIfMatches(empresa.id);
      selectorEmpresasOrder = normalizeEmpresaOrderIDs(selectorEmpresasOrder).filter(function (id) {
        return String(id) !== String(empresa.id);
      });
      writeSelectorOrderLocal(selectorEmpresasOrder);
      setSelectorDeleteMessage("Empresa eliminada. Registros eliminados: " + String(result.registros_eliminados || 0) + (erroresArchivos ? ". Advertencias al limpiar archivos: " + erroresArchivos + "." : "."), false);
      setSelectorDeleteBusy(true, "Actualizando selector...");
      await render();
      await loadPendingShareInvites();
      setSelectorDeleteBusy(false);
      closeSelectorDeleteModal();
      showEmpresasPanel();
      setShareNotice("Empresa eliminada definitivamente: " + (empresa.nombre || ("#" + empresa.id)), false);
    } catch (err) {
      setSelectorDeleteBusy(false);
      setSelectorDeleteMessage(err && err.message ? err.message : "No se pudo eliminar la empresa.", true);
    }
  }

  async function setEmpresaEstado(empresa, estadoObjetivo) {
    var empresaId = Number(empresa && empresa.id ? empresa.id : 0);
    if (!empresaId) {
      throw new Error("empresa_id inv?lido");
    }

    if (estadoObjetivo === "inactivo") {
      var impacto = await fetchEmpresaImpacto(empresaId);
      var resumen = formatImpactoTexto(impacto);
      var force = false;

      if (resumen) {
        force = window.confirm(
          "La empresa tiene impacto operativo activo:\n" + resumen + "\n\n?Deseas desactivarla de todas formas?"
        );
        if (!force) {
          return;
        }
      } else {
        var confirmar = window.confirm("?Confirmas desactivar la empresa '" + (empresa.nombre || "") + "'?");
        if (!confirmar) {
          return;
        }
      }

      var disableURL = "/super/api/empresas?id=" + encodeURIComponent(empresaId) + "&action=desactivar";
      if (force) {
        disableURL += "&force=1";
      }
      var disableRes = await fetch(disableURL, {
        method: "PUT",
        credentials: "same-origin",
      });
      var disableRaw = await disableRes.text();
      if (!disableRes.ok) {
        throw new Error(disableRaw || "No se pudo desactivar la empresa");
      }
      await render();
      return;
    }

    var activateRes = await fetch(
      "/super/api/empresas?id=" + encodeURIComponent(empresaId) + "&action=activar&activo=1",
      {
        method: "PUT",
        credentials: "same-origin",
      }
    );
    var activateRaw = await activateRes.text();
    if (!activateRes.ok) {
      throw new Error(activateRaw || "No se pudo reactivar la empresa");
    }
    await render();
  }

  function appendEmpresasGroup(container, title, empresas, activeByEmpresa, options) {
    if (!empresas.length) return;
    var opts = options && typeof options === "object" ? options : {};
    var section = document.createElement("section");
    section.className = "card empresa-section selector-company-group" + (opts.variant ? " selector-company-group--" + opts.variant : "");

    var header = document.createElement("div");
    header.className = "empresa-section-header";
    header.innerHTML = "<h2>" + title + "</h2><span class=\"form-help empresa-group-total\">Total: " + empresas.length + "</span>";

    var grid = document.createElement("div");
    grid.className = "portal-grid empresas-grid";
    empresas.forEach(function (empresa) {
      var hasLicense = !!activeByEmpresa[empresa.id];
      grid.appendChild(buildEmpresaCard(empresa, hasLicense));
    });
    initializeEmpresaDragAndDrop(grid);

    section.appendChild(header);
    section.appendChild(grid);
    container.appendChild(section);
  }

  async function render() {
    try {
      setShareNotice("", false);
      var meRes = await fetch("/me");
      if (!meRes.ok) {
        window.location.href = "/login.html";
        return;
      }
      var me = await meRes.json();
      currentAccount = me || currentAccount;

      var licenciasURL = "/super/api/licencias?scope=mine&con_empresa=1&activo=1";
      var empPromise = fetch("/super/api/empresas");
      var tiposPromise = fetch("/super/api/tipos_empresas");

      var licRes = await fetch(licenciasURL);
      if (!licRes.ok) {
        console.warn("Fallo consulta de licencias con scope=mine, usando fallback global filtrado por activas.");
        licRes = await fetch("/super/api/licencias?con_empresa=1&activo=1");
      }

      var empRes = await empPromise;
      var tiposRes = await tiposPromise;

      if (!empRes.ok) {
        var txt = await empRes.text().catch(function () {
          return "";
        });
        throw new Error("failed to query empresas: " + (txt || String(empRes.status)));
      }

      var empresas = await empRes.json();
      if (!Array.isArray(empresas)) empresas = empresas ? [empresas] : [];

      var licencias = licRes.ok ? await licRes.json() : [];
      if (!Array.isArray(licencias)) licencias = licencias ? [licencias] : [];

      var tipos = tiposRes.ok ? await tiposRes.json() : [];
      if (!Array.isArray(tipos)) tipos = tipos ? [tipos] : [];

      var activeByEmpresa = {};
      licencias.forEach(function (l) {
        if (l.empresa_id && (l.activo === 1 || l.activo === "1" || l.activo === "activo")) {
          activeByEmpresa[l.empresa_id] = true;
        }
      });
      currentActiveByEmpresa = activeByEmpresa;

      var container = document.getElementById("cards");
      container.innerHTML = "";

      var tipoSelect = document.getElementById("tipo_id");
      if (tipoSelect) {
        tipoSelect.innerHTML = '<option value="">-- Seleccionar --</option>';
        tipos.slice().reverse().forEach(function (t) {
          var opt = document.createElement("option");
          opt.value = t.nombre;
          opt.text = t.nombre;
          opt.dataset.id = t.id;
          var plantilla = getVerticalByTypeName(t.nombre);
          if (plantilla) {
            opt.dataset.verticalModule = plantilla.module || "";
            opt.dataset.verticalTitle = plantilla.fullTitle || plantilla.title || "";
          }
          tipoSelect.appendChild(opt);
        });
        renderTipoEmpresaPreview();
        if (!tipoSelect.dataset.verticalPreviewBound) {
          tipoSelect.dataset.verticalPreviewBound = "1";
          tipoSelect.addEventListener("change", renderTipoEmpresaPreview);
        }
      }

      await loadSelectorEmpresasOrder();
      var list = sortEmpresasBySelectorOrder(empresas);
      currentEmpresas = list.slice();
      if (!readEmpresaContext() && list.length > 0) {
        persistEmpresaContext(list[0].id);
      }
      if (list.length === 0) {
        showForm();
        try {
          var msgEl = document.getElementById("msg");
          if (msgEl) msgEl.textContent = "Agrega una empresa para continuar";
        } catch (e) {}
      } else {
        try {
          var msgEl = document.getElementById("msg");
          if (msgEl) msgEl.textContent = "";
        } catch (e) {}
        try { hideForm(); } catch (e) {}
      }

      var conLicenciaActiva = list.filter(function (e) {
        return !!activeByEmpresa[e.id];
      });
      var sinLicenciaActiva = list.filter(function (e) {
        return !activeByEmpresa[e.id];
      });

      appendEmpresasGroup(container, "Empresas con licencia activa", conLicenciaActiva, activeByEmpresa, { variant: "active" });
      appendEmpresasGroup(container, "Empresas sin licencia activa", sinLicenciaActiva, activeByEmpresa, { variant: "inactive" });

      document.getElementById("addBtn").onclick = function () {
        showForm();
        setActiveNav(document.getElementById("linkAgregarEmpresa"));
      };
    } catch (err) {
      console.error(err);
      var target = document.getElementById("cards");
      target.innerText = "Error cargando empresas: " + (err && err.message ? err.message : String(err));
    }
  }

  function showForm() {
    setHidden(empresasPanel, false);
    if (contentFrame) {
      setHidden(contentFrame, true);
      contentFrame.setAttribute("src", "about:blank");
    }
    setHidden(document.getElementById("form"), false);
    setHidden(document.getElementById("addBtn"), true);
    persistView({ mode: "form" });
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  function hideForm() {
    setHidden(document.getElementById("form"), true);
    setHidden(document.getElementById("addBtn"), false);
    document.getElementById("itemId").value = "";
    document.getElementById("nombre").value = "";
    document.getElementById("nit").value = "";
    document.getElementById("observaciones").value = "";
    persistView({ mode: "empresas" });
  }

  function showEmpresasPanel() {
    setHidden(empresasPanel, false);
    if (contentFrame) {
      setHidden(contentFrame, true);
      contentFrame.setAttribute("src", "about:blank");
    }
    hideForm();
    persistView({ mode: "empresas" });
  }

  function findLinkByHref(href) {
    var normalized = normalizeHref(href);
    if (!normalized) return null;
    var normalizedPath = normalized.split("?")[0];
    for (var i = 0; i < navLinks.length; i++) {
      var link = navLinks[i];
      var linkHref = normalizeHref(link.getAttribute("href"));
      if (!linkHref) continue;
      if (linkHref === normalized) return link;
      if (linkHref.split("?")[0] === normalizedPath) return link;
    }
    return null;
  }

  function restoreLastView() {
    var view = readView();
    var linkAgregar = document.getElementById("linkAgregarEmpresa");

    // Si el usuario volvi? con bot?n "atr?s" desde otra p?gina (bfcache / back-forward),
    // priorizamos mostrar la lista de empresas. Esto evita que quede oculta por un view
    // previo en modo "frame" (licencias/reportes) y parezca que "desaparecieron" las tarjetas.
    var forceEmpresas = false;
    try {
      var nav = performance && performance.getEntriesByType ? performance.getEntriesByType('navigation') : [];
      var navType = nav && nav[0] ? String(nav[0].type || '') : '';
      if (navType === 'back_forward') {
        forceEmpresas = true;
      }
    } catch (e) {}
    try {
      if (!forceEmpresas) {
        var ref = String(document.referrer || '').trim();
        if (ref && ref.indexOf(window.location.origin) === 0) {
          // si venimos desde otra p?gina del mismo sitio (no desde seleccionar_empresa),
          // mostramos tarjetas por defecto.
          var refPath = '';
          try { refPath = (new URL(ref)).pathname || ''; } catch (e) { refPath = ''; }
          if (refPath && refPath !== window.location.pathname) {
            forceEmpresas = true;
          }
        }
      }
    } catch (e) {}

    if (!view || !view.mode) {
      showEmpresasPanel();
      setActiveNav(linkAgregar);
      return;
    }

    if (forceEmpresas) {
      showEmpresasPanel();
      setActiveNav(linkAgregar);
      return;
    }

    if (view.mode === "frame" && view.href) {
      var targetLink = findLinkByHref(view.href);
      if (!targetLink && isStandaloneFrameHref(view.href)) {
        openInRightFrame(view.href, null);
        return;
      }
      if (!targetLink || !isSidebarLinkVisible(targetLink)) {
        showEmpresasPanel();
        setActiveNav(linkAgregar);
        return;
      }
      openInRightFrame(view.href, targetLink);
      if (targetLink) setActiveNav(targetLink);
      return;
    }

    if (view.mode === "form") {
      showForm();
      setActiveNav(linkAgregar);
      return;
    }

    showEmpresasPanel();
    setActiveNav(linkAgregar);
  }

  function wireSidebarFrameLinks() {
    var linkAgregar = document.getElementById("linkAgregarEmpresa");
    var linkLicencias = document.getElementById("linkLicencias");
    var linkAdministradores = document.getElementById("linkAdministradores");
    var linkAuditoriaGlobal = document.getElementById("linkAuditoriaGlobal");
    var linkReportes = document.getElementById("linkReportesGlobales");

    if (linkAgregar) {
      linkAgregar.addEventListener("click", function (ev) {
        ev.preventDefault();
        showEmpresasPanel();
        setActiveNav(linkAgregar);
      });
    }

    [linkLicencias, linkAdministradores, linkAuditoriaGlobal, linkReportes].forEach(function (link) {
      if (!link) return;
      link.addEventListener("click", function (ev) {
        ev.preventDefault();
        openInRightFrame(link.getAttribute("href"), link);
      });
    });
  }

  async function processSharedInvitationToken() {
    var token = getQueryParam("shared_invitation_token");
    if (!token) {
      return;
    }
    try {
      var res = await fetch("/super/api/empresas/compartidos/aceptar", {
        method: "POST",
        credentials: "same-origin",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ token: token })
      });
      var raw = await res.text();
      var data = null;
      try {
        data = raw ? JSON.parse(raw) : null;
      } catch (e) {
        data = null;
      }
      if (!res.ok) {
        throw new Error((data && (data.message || data.error)) || raw || "No se pudo aceptar la invitación compartida.");
      }
      if (data && data.empresa_id) {
        persistEmpresaContext(data.empresa_id);
      }
      setShareNotice((data && data.message) || "La empresa compartida ya está disponible en tu selector.", false);
      clearQueryParam("shared_invitation_token");
      showEmpresasPanel();
    } catch (err) {
      setShareNotice(err && err.message ? err.message : "No se pudo aceptar la invitación compartida.", true);
      clearQueryParam("shared_invitation_token");
    }
  }

  function processSharedInvitationAcceptedNotice() {
    if (getQueryParam("shared_invitation_accepted") !== "1") {
      return;
    }
    var empresaId = parseInt(getQueryParam("empresa_id") || "0", 10);
    if (Number.isFinite(empresaId) && empresaId > 0) {
      persistEmpresaContext(empresaId);
    }
    setShareNotice("Invitación aceptada correctamente. La empresa compartida ya aparece en tu lista.", false);
    clearQueryParam("shared_invitation_accepted");
  }

  function wireSelectorEmpresaOrderControls() {
    var resetBtn = document.getElementById("resetEmpresaOrderBtn");
    if (!resetBtn || resetBtn.dataset.bound === "1") return;
    resetBtn.dataset.bound = "1";
    resetBtn.addEventListener("click", async function () {
      resetBtn.disabled = true;
      try {
        selectorEmpresasOrder = [];
        writeSelectorOrderLocal([]);
        await saveSelectorEmpresasOrder([], true);
        await render();
        setShareNotice("Orden restablecido.", false);
      } catch (err) {
        setShareNotice(err && err.message ? err.message : "No se pudo restablecer el orden.", true);
      } finally {
        resetBtn.disabled = false;
      }
    });
  }

  document.addEventListener("DOMContentLoaded", function () {
    applySidebarPermissions(null);
    wireSidebarFrameLinks();
    wireSelectorEmpresaOrderControls();
    fetchCurrentAccount().finally(async function () {
      await render();
      await processSharedInvitationToken();
      await render();
      await loadPendingShareInvites();
      processSharedInvitationAcceptedNotice();
      restoreLastView();
    });

    // Cuando el navegador restaura la p?gina desde el cache de "atr?s/adelante",
    // aseguramos que el panel de empresas quede visible y el listado se refresque.
    window.addEventListener('pageshow', function (ev) {
      try {
        if (ev && ev.persisted) {
          showEmpresasPanel();
          render();
        }
      } catch (e) {}
    });

    window.addEventListener("message", function (ev) {
      if (ev.origin !== window.location.origin) {
        return;
      }
      var data = ev.data || {};
      if (!data || data.type !== "pcs:selector-show-empresas") {
        return;
      }
      showEmpresasPanel();
      setActiveNav(document.getElementById("linkAgregarEmpresa"));
    });

    var form = document.getElementById("form");
    if (!form) return;
    var createEmpresaSubmitting = false;

    form.onsubmit = async function (e) {
      e.preventDefault();
      if (createEmpresaSubmitting) {
        var currentMsg = document.getElementById("msg");
        if (currentMsg) currentMsg.innerText = "La empresa se esta creando. Espera un momento.";
        return;
      }
      var tipoSelect = document.getElementById("tipo_id");
      var selectedOption = tipoSelect && tipoSelect.options ? tipoSelect.options[tipoSelect.selectedIndex] : null;
      var tipoID = 0;
      var tipoNombre = "";
      if (selectedOption) {
      tipoID = parseInt(selectedOption.dataset.id || "0", 10) || 0;
      tipoNombre = selectedOption.text || "";
      }
      var payload = {
      tipo_id: tipoID,
      tipo_nombre: tipoNombre,
        nombre: document.getElementById("nombre").value.trim(),
        nit: document.getElementById("nit").value.trim(),
        observaciones: document.getElementById("observaciones").value.trim(),
        usuario_creador: "",
      };
      var saveBtn = document.getElementById("saveBtn");
      var originalSaveText = saveBtn ? saveBtn.textContent : "";
      try {
        createEmpresaSubmitting = true;
        if (saveBtn) {
          saveBtn.disabled = true;
          saveBtn.setAttribute("aria-busy", "true");
          saveBtn.textContent = "Creando...";
        }
        var meRes = await fetch("/me");
        if (meRes.ok) {
          var me = await meRes.json();
          payload.usuario_creador = me.email || "";
        }
        if (!payload.nombre) {
          document.getElementById("msg").innerText = "Nombre requerido";
          return;
        }
        var createRes = await fetch("/super/api/empresas", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload),
        });
        var createRaw = await createRes.text();
        var createData = null;
        try {
          createData = createRaw ? JSON.parse(createRaw) : null;
        } catch (parseErr) {
          createData = null;
        }
        if (!createRes.ok) {
          throw new Error(createRaw || "No se pudo crear la empresa");
        }
        hideForm();
        await handleEmpresaPreconfigDecision(createData);
        if (createData && createData.id) {
          recordSelectorAuditEvent("empresa_creada_desde_selector", {
            empresa_id: Number(createData.id || 0),
            empresa_nombre: payload.nombre,
            tipo_nombre: payload.tipo_nombre,
            preconfiguracion_aplicada: !!createData.preconfiguracion_aplicada,
            view: "form_agregar_empresa"
          });
        }
        await render();
      } catch (err) {
        document.getElementById("msg").innerText = err.message;
      } finally {
        createEmpresaSubmitting = false;
        if (saveBtn) {
          saveBtn.disabled = false;
          saveBtn.removeAttribute("aria-busy");
          saveBtn.textContent = originalSaveText || "Guardar";
        }
      }
    };

    document.getElementById("cancelBtn").onclick = hideForm;
  });

  document.addEventListener('submit', function (ev) {
    var shareForm = ev.target.closest && ev.target.closest('form.empresa-card-share-form');
    if (!shareForm) return;
    ev.preventDefault();
    ev.stopPropagation();
    submitEmpresaShareInvitation(shareForm);
  });

  document.addEventListener('change', function (ev) {
    var levelInput = ev.target && ev.target.closest ? ev.target.closest('[data-share-level]') : null;
    if (!levelInput) return;
    updateEmpresaCardShareScope(levelInput.closest('.empresa-card-share-panel'));
  });

  document.addEventListener('click', function(ev){
    var shareBtn = ev.target.closest && ev.target.closest('button.empresa-share-toggle');
    if (shareBtn) {
      ev.preventDefault();
      ev.stopPropagation();
      var disabled = shareBtn.getAttribute('data-share-disabled') === '1';
      var empresaIdShare = parseInt(shareBtn.getAttribute('data-empresa-id') || '0', 10);
      if (disabled) {
        setShareNotice('Solo el propietario, super administrador o administrador compartido autorizado puede enviar invitaciones.', true);
        return;
      }
      if (empresaIdShare > 0) {
        setShareNotice('', false);
        toggleEmpresaSharePanel(empresaIdShare);
      }
      return;
    }

    var resendBtn = ev.target.closest && ev.target.closest('button.empresa-share-resend');
    if (resendBtn) {
      ev.preventDefault();
      ev.stopPropagation();
      var empresaIdResend = parseInt(resendBtn.getAttribute('data-empresa-id') || '0', 10);
      var invitationId = parseInt(resendBtn.getAttribute('data-invitation-id') || '0', 10);
      if (!empresaIdResend || !invitationId) {
        return;
      }
      setEmpresaShareFeedback(empresaIdResend, 'Reenviando invitación...', false);
      fetchJSON('/super/api/empresas/compartidos?id=' + encodeURIComponent(invitationId) + '&action=reenviar', {
        method: 'PUT',
        credentials: 'same-origin'
      }).then(function (data) {
        var msg = data && data.message ? data.message : 'Invitación reenviada.';
        setEmpresaShareFeedback(empresaIdResend, msg, false);
        setShareNotice(msg, false);
      }).catch(function (err) {
        var msg = err && err.message ? err.message : 'No se pudo reenviar la invitación.';
        setEmpresaShareFeedback(empresaIdResend, msg, true);
        setShareNotice(msg, true);
      });
      return;
    }

    var acceptInviteBtn = ev.target.closest && ev.target.closest('button[data-action="accept-share-invite"]');
    if (acceptInviteBtn) {
      ev.preventDefault();
      ev.stopPropagation();
      var invitationIdAccept = parseInt(acceptInviteBtn.getAttribute("data-invitation-id") || "0", 10);
      var empresaIdAccept = parseInt(acceptInviteBtn.getAttribute("data-empresa-id") || "0", 10);
      if (!invitationIdAccept || !empresaIdAccept) {
        return;
      }
      setShareNotice("Aceptando invitación...", false);
      fetchJSON("/super/api/empresas/compartidos?id=" + encodeURIComponent(invitationIdAccept) + "&action=aceptar", {
        method: "PUT",
        credentials: "same-origin",
      }).then(async function (data) {
        var empresaId = data && data.empresa_id ? Number(data.empresa_id) : empresaIdAccept;
        if (empresaId) {
          persistEmpresaContext(empresaId);
        }
        await render();
        await loadPendingShareInvites();
        var empresa = getEmpresaFromCurrentList(empresaId);
        if (empresa) {
          var hasLicense = !!currentActiveByEmpresa[empresa.id];
          navigateToEmpresa(empresa, hasLicense);
          return;
        }
        setShareNotice("Invitación aceptada. La empresa ya está en tu lista.", false);
      }).catch(function (err) {
        var msg = err && err.message ? err.message : "No se pudo aceptar la invitación.";
        setShareNotice(msg, true);
      });
      return;
    }

    var refreshInvitesBtn = ev.target.closest && ev.target.closest('button[data-action="refresh-share-invites"]');
    if (refreshInvitesBtn) {
      ev.preventDefault();
      ev.stopPropagation();
      loadPendingShareInvites();
      return;
    }

    var btn = ev.target.closest && ev.target.closest('button.download-data');
    if (btn) {
      ev.preventDefault();
      ev.stopPropagation();
      var id = parseInt(btn.getAttribute('data-empresa-id') || '0', 10);
      var name = btn.getAttribute('data-empresa-name') || '';
      if (!id) return;
      var params = new URLSearchParams();
      params.set('empresa_id', String(id));
      params.set('id', String(id));
      if (name) params.set('empresa_nombre', name);
      params.set('embedded', '1');
      persistEmpresaContext(id);
      openInRightFrame('/descargar_informacion_de_la_empresa.html?' + params.toString(), null);
      return;
    }

    if (!ev.target.closest || !ev.target.closest('.empresa-card-share-panel')) {
      closeAllEmpresaSharePanels(0);
    }
  });
})();
