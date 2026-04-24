(function () {
  var empresasPanel = document.getElementById("empresasPanel");
  var contentFrame = document.getElementById("contentFrame");
  var navLinks = Array.from(document.querySelectorAll(".admin-sidebar .nav a"));
  var storage = null;
  var viewKey = "seleccionar_empresa:view";
  var currentEmpresas = [];
  var currentAccount = null;
  var shareNoticeEl = document.getElementById("selectorShareNotice");

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
      throw new Error((data && (data.error || data.message)) || raw || "Solicitud fallida");
    }
    return data;
  }

  function setShareNotice(text, isError) {
    if (!shareNoticeEl) return;
    shareNoticeEl.classList.toggle("is-hidden", !text);
    shareNoticeEl.textContent = text || "";
    shareNoticeEl.classList.toggle("error", !!isError);
    shareNoticeEl.classList.toggle("success", !isError && !!text);
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
    var linkAdministradores = document.getElementById("linkAdministradores");
    var linkReportes = document.getElementById("linkReportesGlobales");
    var principalSuper = isPrincipalSuperAccount(account);
    setElementVisible(linkLicencias, canManageScopedLicencias(account));
    setElementVisible(linkAdministradores, principalSuper);
    setElementVisible(linkReportes, principalSuper);
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
  }

  function normalizeCompanyTypeName(value) {
    var normalized = String(value || "").trim().toLowerCase();
    if (typeof normalized.normalize === "function") {
      normalized = normalized.normalize("NFD").replace(/[\u0300-\u036f]/g, "");
    }
    return normalized;
  }

  function getEmpresaTypeVisual(empresa) {
    var tipoNombre = String(empresa && empresa.tipo_nombre ? empresa.tipo_nombre : "").trim();
    var normalized = normalizeCompanyTypeName(tipoNombre);
    var visualRules = [
      {
        pattern: /(restaurante|restaurant|bar|cafe|cafeteria|panaderia|pasteleria|comida|pizzeria|licoreria|gastro)/,
        tone: "food",
        icon: "/img/restaurante.png",
        alt: "Icono de restaurante",
        eyebrow: "Atencion gastronomica",
        activeCopy: "Operacion lista para atender clientes, registrar consumos y administrar cobros del negocio.",
        pendingCopy: "Configura la licencia para activar una operacion agil de mesas, pedidos y facturacion del local."
      },
      {
        pattern: /(hotel|hostal|hosped|motel|apartahotel|resort|alojamiento)/,
        tone: "lodging",
        icon: "/img/motel.png",
        alt: "Icono de hotel o motel",
        eyebrow: "Operacion de hospedaje",
        activeCopy: "Gestion preparada para reservas, recepcion, habitaciones y seguimiento operativo por estancia.",
        pendingCopy: "Activa la licencia para gestionar hospedaje, recepcion y trazabilidad comercial por habitacion."
      },
      {
        pattern: /(tienda|almacen|supermercado|market|boutique|farmacia|drogueria|minimercado|retail|comercio|ferreteria|papeleria|pos|punto de venta)/,
        tone: "retail",
        icon: "/img/punto_venta.png",
        alt: "Icono de punto de venta",
        eyebrow: "Comercio y mostrador",
        activeCopy: "Empresa lista para ventas de mostrador, control comercial e interaccion directa con clientes.",
        pendingCopy: "Habilita la licencia para operar catalogo, facturacion y flujo comercial en punto de venta."
      },
      {
        pattern: /(bodega|distribuidora|logistica|almacenamiento|inventario|deposito|suministros|mayorista|warehouse)/,
        tone: "logistics",
        icon: "/img/warehouse-color.svg",
        alt: "Icono de bodega o logistica",
        eyebrow: "Control de inventario",
        activeCopy: "Preparada para movimientos de bodega, control de existencias y operacion logistica por empresa.",
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
      eyebrow: "Operacion empresarial",
      activeCopy: "Empresa disponible para continuar la gestion administrativa y operativa desde el panel principal.",
      pendingCopy: "Configura la licencia para habilitar la operacion completa de esta empresa dentro del sistema."
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
    return "Empresa propia";
  }

  function isSharedEmpresa(empresa) {
    return String(empresa && empresa.access_source ? empresa.access_source : "owner").toLowerCase() === "shared";
  }

  function navigateToEmpresa(empresa, hasLicense) {
    persistEmpresaContext(empresa.id);
    if (hasLicense) {
      var adminURL =
        "/administrar_empresa.html?id=" + encodeURIComponent(empresa.id) +
        "&empresa_id=" + encodeURIComponent(empresa.id);
      window.location.href = adminURL;
      return;
    }
    var params = new URLSearchParams();
    params.set("empresa_id", empresa.id);
    params.set("id", empresa.id);
    if (empresa.tipo_id) params.set("tipo_id", empresa.tipo_id);
    if (empresa.tipo_nombre) params.set("tipo_nombre", empresa.tipo_nombre);
    window.location.href = "/elegir_licencia.html?" + params.toString();
  }

  function buildEmpresaShareButton(empresa) {
    var disabled = isSharedEmpresa(empresa);
    var title = disabled
      ? "Solo el administrador propietario puede compartir una empresa que recibió por invitación"
      : "Compartir empresa con otro administrador";
    return '' +
      '<button type="button" class="empresa-card-icon-action empresa-share-toggle' + (disabled ? ' is-disabled' : '') + '" data-empresa-id="' + escapeHtml(String(empresa.id || '')) + '" data-share-disabled="' + (disabled ? '1' : '0') + '" aria-label="' + escapeHtml(title) + '" title="' + escapeHtml(title) + '">' +
      '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="18" height="18" aria-hidden="true" focusable="false"><path fill="currentColor" d="M15 8a3 3 0 1 0-2.83-4H12a3 3 0 0 0 .17 1L8.91 6.63a3 3 0 0 0-1.91-.68A3 3 0 1 0 10 9a2.9 2.9 0 0 0-.17-.98l3.26-1.63A3 3 0 0 0 15 8Zm-8 3a3 3 0 0 0 1.91-.68l3.26 1.63A3 3 0 0 0 12 13a3 3 0 1 0 1.09-2.31l-3.26-1.63A2.9 2.9 0 0 0 10 8a2.9 2.9 0 0 0-.09-.69l3.26-1.63A3 3 0 0 0 15 6a3 3 0 1 0-1.09-2.31L10.65 5.32A3 3 0 1 0 7 11Zm8 8a3 3 0 0 0-2.83-4H12a3 3 0 0 0 .17 1l-3.26 1.63A3 3 0 1 0 10 20a2.9 2.9 0 0 0-.17-.98l3.26-1.63A3 3 0 1 0 15 19Z"/></svg>' +
      '</button>';
  }

  function buildEmpresaSharePanel(empresa) {
    if (isSharedEmpresa(empresa)) {
      return '' +
        '<div class="empresa-card-share-panel" data-empresa-id="' + escapeHtml(String(empresa.id || '')) + '" hidden>' +
        '<p class="empresa-card-share-feedback is-error">Solo el administrador propietario puede enviar nuevas invitaciones para esta empresa.</p>' +
        '</div>';
    }
    return '' +
      '<div class="empresa-card-share-panel" data-empresa-id="' + escapeHtml(String(empresa.id || '')) + '" hidden>' +
      '<form class="empresa-card-share-form" data-empresa-id="' + escapeHtml(String(empresa.id || '')) + '">' +
      '<label class="empresa-card-share-label" for="share-email-' + escapeHtml(String(empresa.id || '')) + '">Compartir con otro administrador</label>' +
      '<div class="empresa-card-share-row">' +
      '<input id="share-email-' + escapeHtml(String(empresa.id || '')) + '" class="form-input empresa-card-share-input" data-share-email type="email" placeholder="correo@ejemplo.com" required>' +
      '<button type="submit" class="btn empresa-card-share-submit">Enviar</button>' +
      '</div>' +
      '<p class="empresa-card-share-feedback" data-share-feedback></p>' +
      '</form>' +
      '</div>';
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
    feedback.textContent = text || '';
    feedback.classList.toggle('is-error', !!isError && !!text);
    feedback.classList.toggle('is-success', !isError && !!text);
  }

  function closeAllEmpresaSharePanels(exceptEmpresaId) {
    Array.prototype.forEach.call(document.querySelectorAll('.empresa-card-share-panel'), function (panel) {
      var panelEmpresaId = parseInt(panel.getAttribute('data-empresa-id') || '0', 10);
      var shouldKeepOpen = exceptEmpresaId > 0 && panelEmpresaId === exceptEmpresaId;
      panel.hidden = !shouldKeepOpen;
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
    var email = String(emailInput && emailInput.value ? emailInput.value : '').trim();
    if (!email) {
      setEmpresaShareFeedback(empresaId, 'Debes escribir el correo del otro administrador.', true);
      return;
    }
    setEmpresaShareFeedback(empresaId, 'Enviando invitación...', false);
    try {
      var data = await fetchJSON('/super/api/empresas/compartidos', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ empresa_id: empresaId, email: email, mensaje: '' })
      });
      var message = data && data.message ? data.message : 'Invitación enviada correctamente.';
      setEmpresaShareFeedback(empresaId, message, false);
      setShareNotice(message, false);
      if (emailInput) {
        emailInput.value = '';
      }
    } catch (err) {
      var errorMessage = err && err.message ? err.message : 'No se pudo enviar la invitación.';
      setEmpresaShareFeedback(empresaId, errorMessage, true);
      setShareNotice(errorMessage, true);
    }
  }

  function buildEmpresaCard(empresa, hasLicense) {
    var visual = getEmpresaTypeVisual(empresa);
    var descripcion = buildEmpresaCardDescription(empresa, visual, hasLicense);
    var cardLink = document.createElement("div");
    cardLink.className = "card-link empresa-card-link";
    cardLink.tabIndex = 0;
    cardLink.setAttribute("role", "button");
    cardLink.setAttribute("aria-label", (hasLicense ? "Administrar " : "Elegir licencia para ") + String(empresa.nombre || "empresa"));
    cardLink.addEventListener("click", function (evt) {
      if (evt.target.closest && evt.target.closest('.empresa-share-toggle, .empresa-card-share-panel, button.download-data')) {
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
    div.className = "portal-card warm empresa-card empresa-tone-" + visual.tone;
    div.setAttribute('data-tone', visual.tone || 'generic');
    div.innerHTML =
      '<span class="empresa-card-badge">' +
      escapeHtml(visual.label || "Empresa") +
      "</span>" +
      '<span class="empresa-card-watermark" aria-hidden="true">' +
      '<img src="' + escapeHtml(visual.icon || "/img/company-briefcase-color.svg") + '" alt="">' +
      "</span>" +
      '<div class="card-body">' +
      '<h3 class="card-title">' +
      escapeHtml(empresa.nombre || "--") +
      "</h3>" +
      '<p class="empresa-shared-note">' +
      escapeHtml(buildEmpresaAccessLabel(empresa)) +
      "</p>" +
      '<p class="card-desc muted">' +
      escapeHtml(descripcion || "") +
      "</p>" +
      '<div class="card-actions">' +
      '<button class="license-indicator ' +
      (hasLicense ? "active" : "inactive") +
      '" type="button" aria-hidden="true">' +
      (hasLicense ? "Licencia activa" : "Sin licencia") +
      "</button>" +
      '<div class="empresa-card-quick-actions">' +
      buildEmpresaShareButton(empresa) +
      '</div>' +
      "</div>" +
      buildEmpresaSharePanel(empresa) +
      "</div>";

    if (!hasLicense) {
      var dlBtn = document.createElement("button");
      dlBtn.type = "button";
      dlBtn.className = "license-indicator active download-data";
      dlBtn.setAttribute("data-empresa-id", String(empresa.id || ""));
      dlBtn.setAttribute("data-empresa-name", String(empresa.nombre || ""));
      dlBtn.setAttribute("aria-label", "Descargar informacion de la empresa " + String(empresa.nombre || ""));
      dlBtn.setAttribute("title", "Descargar informacion de la empresa");
      dlBtn.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="16" height="16" aria-hidden="true" focusable="false"><path fill="currentColor" d="M12 3v10l4-4-1.4-1.4L13 9.2V3h-2zM5 18v2h14v-2H5z"/></svg>';
      var actions = div.querySelector(".card-actions");
      if (actions) {
        actions.appendChild(dlBtn);
      }
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
      throw new Error((data && (data.error || data.message)) || raw || "No se pudo obtener impacto de desactivación");
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

  async function setEmpresaEstado(empresa, estadoObjetivo) {
    var empresaId = Number(empresa && empresa.id ? empresa.id : 0);
    if (!empresaId) {
      throw new Error("empresa_id inválido");
    }

    if (estadoObjetivo === "inactivo") {
      var impacto = await fetchEmpresaImpacto(empresaId);
      var resumen = formatImpactoTexto(impacto);
      var force = false;

      if (resumen) {
        force = window.confirm(
          "La empresa tiene impacto operativo activo:\n" + resumen + "\n\n¿Deseas desactivarla de todas formas?"
        );
        if (!force) {
          return;
        }
      } else {
        var confirmar = window.confirm("¿Confirmas desactivar la empresa '" + (empresa.nombre || "") + "'?");
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

  function appendEmpresasGroup(container, title, empresas, activeByEmpresa) {
    if (!empresas.length) return;
    var section = document.createElement("section");
    section.className = "card empresa-section";

    var header = document.createElement("div");
    header.className = "empresa-section-header";
    header.innerHTML = "<h2>" + title + "</h2><span class=\"form-help empresa-group-total\">Total: " + empresas.length + "</span>";

    var grid = document.createElement("div");
    grid.className = "portal-grid empresas-grid";
    empresas.forEach(function (empresa) {
      var hasLicense = !!activeByEmpresa[empresa.id];
      grid.appendChild(buildEmpresaCard(empresa, hasLicense));
    });

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

      var container = document.getElementById("cards");
      container.innerHTML = "";

      var tipoSelect = document.getElementById("tipo_id");
      if (tipoSelect) {
        tipoSelect.innerHTML = '<option value="">-- Seleccionar --</option>';
        tipos.forEach(function (t) {
          var opt = document.createElement("option");
          opt.value = t.nombre;
          opt.text = t.nombre;
          opt.dataset.id = t.id;
          tipoSelect.appendChild(opt);
        });
      }

      var list = empresas.slice().sort(function (left, right) {
        var leftShared = String(left && left.access_source ? left.access_source : "owner").toLowerCase() === "shared";
        var rightShared = String(right && right.access_source ? right.access_source : "owner").toLowerCase() === "shared";
        if (leftShared !== rightShared) {
          return leftShared ? 1 : -1;
        }
        return String(left && left.nombre ? left.nombre : "").localeCompare(String(right && right.nombre ? right.nombre : ""), "es", { sensitivity: "base" });
      });
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

      appendEmpresasGroup(container, "Empresas con licencia activa", conLicenciaActiva, activeByEmpresa);
      appendEmpresasGroup(container, "Empresas sin licencia activa", sinLicenciaActiva, activeByEmpresa);

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

    if (!view || !view.mode) {
      showEmpresasPanel();
      setActiveNav(linkAgregar);
      return;
    }

    if (view.mode === "frame" && view.href) {
      var targetLink = findLinkByHref(view.href);
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
    var linkEditarEmpresaMenu = document.getElementById("linkEditarEmpresaMenu");
    var linkLicencias = document.getElementById("linkLicencias");
    var linkAdministradores = document.getElementById("linkAdministradores");
    var linkReportes = document.getElementById("linkReportesGlobales");

    if (linkAgregar) {
      linkAgregar.addEventListener("click", function (ev) {
        ev.preventDefault();
        showEmpresasPanel();
        setActiveNav(linkAgregar);
      });
    }

    if (linkEditarEmpresaMenu) {
      linkEditarEmpresaMenu.addEventListener("click", function (ev) {
        ev.preventDefault();
        var empresaId = resolveEmpresaIdForMenu();
        if (!empresaId) {
          window.alert("Primero crea o selecciona una empresa para editarla.");
          return;
        }
        var empresa = getEmpresaFromCurrentList(empresaId);
        if (empresa && String(empresa.access_source || "owner").toLowerCase() === "shared") {
          window.alert("Solo el administrador propietario puede editar o eliminar la ficha de una empresa compartida.");
          return;
        }
        persistEmpresaContext(empresaId);
        setActiveNav(linkEditarEmpresaMenu);
        window.location.href = "/editar_empresa.html?id=" + encodeURIComponent(String(empresaId)) + "&empresa_id=" + encodeURIComponent(String(empresaId));
      });
    }

    [linkLicencias, linkAdministradores, linkReportes].forEach(function (link) {
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

  document.addEventListener("DOMContentLoaded", function () {
    applySidebarPermissions(null);
    wireSidebarFrameLinks();
    fetchCurrentAccount().finally(async function () {
      await render();
      await processSharedInvitationToken();
      await render();
      restoreLastView();
    });

    var form = document.getElementById("form");
    if (!form) return;

    form.onsubmit = async function (e) {
      e.preventDefault();
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
      try {
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
        if (!createRes.ok) {
          var errorText = await createRes.text().catch(function () {
            return "";
          });
          throw new Error(errorText || "No se pudo crear la empresa");
        }
        hideForm();
        render();
      } catch (err) {
        document.getElementById("msg").innerText = err.message;
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

  document.addEventListener('click', function(ev){
    var shareBtn = ev.target.closest && ev.target.closest('button.empresa-share-toggle');
    if (shareBtn) {
      ev.preventDefault();
      ev.stopPropagation();
      var disabled = shareBtn.getAttribute('data-share-disabled') === '1';
      var empresaIdShare = parseInt(shareBtn.getAttribute('data-empresa-id') || '0', 10);
      if (disabled) {
        setShareNotice('Solo el administrador propietario puede enviar invitaciones para una empresa compartida.', true);
        return;
      }
      if (empresaIdShare > 0) {
        setShareNotice('', false);
        toggleEmpresaSharePanel(empresaIdShare);
      }
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
      persistEmpresaContext(id);
      window.location.href = '/descargar_informacion_de_la_empresa.html?' + params.toString();
      return;
    }

    if (!ev.target.closest || !ev.target.closest('.empresa-card-share-panel')) {
      closeAllEmpresaSharePanels(0);
    }
  });
})();
