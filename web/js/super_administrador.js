(function () {
  var sidebar = document.querySelector(".admin-sidebar .nav");
  var links = sidebar ? Array.from(sidebar.querySelectorAll("a")) : [];
  var iframe = document.getElementById("contentFrame");
  var favoriteBtn = document.getElementById("superFavoriteBtn");
  var storage = null;
  var localStore = null;
  var lastPageKey = "super_admin:last_page";
  var favoritesKey = "super_admin:favorites";

  try {
    storage = window.sessionStorage;
  } catch (e) {
    storage = null;
  }

  try {
    localStore = window.localStorage;
  } catch (e) {
    localStore = null;
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

  var coreSuperPages = {
    "/super/licencias_resumen.html": true,
    "/super/auditoria_global.html": true,
    "/super/diagramas/documentacion_tecnica_completa.html": true,
    "/super/diagramas/arquitectura_general.html": true,
    "/super/diagramas/mapa_de_modulos.html": true,
    "/super/diagramas/mapa_de_navegacion.html": true,
    "/super/diagramas/erd_global_resumido_por_dominios.html": true,
    "/super/diagramas/casos_de_uso.html": true,
    "/super/diagramas/diagrama_de_clases_uml.html": true,
    "/super/diagramas/login_y_resolucion_de_empresa.html": true,
    "/super/diagramas/venta_pos_con_inventario_y_caja.html": true,
    "/super/diagramas/facturacion_electronica_dian.html": true,
    "/super/diagramas/webhook_rappi_separado_por_empresa.html": true,
    "/super/diagramas/cierre_de_venta.html": true,
    "/super/diagramas/despliegue_rs.html": true,
    "/super/diagramas/diagramas_de_estados.html": true,
    "/super/diagramas/diagramas_de_estados_2.html": true,
    "/super/diagramas/diagramas_de_estados_3.html": true,
    "/super/diagramas/diagrama_de_componentes.html": true,
    "/super/diagramas/diagrama_de_despliegue.html": true,
    "/super/diagramas/diagrama_de_paquetes.html": true,
    "/super/diagramas/diagrama_de_flujo_de_datos.html": true,
    "/super/pagina_principal.html": true,
    "/super/informacion_de_modulos.html": true,
    "/super/noticias.html": true,
    "/super/informacion_de_la_empresa_y_de_los_sistemas_para_ia.html": true,
    "/super/tipos_empresas.html": true,
    "/super/preconfiguracion_tipos_empresa.html": true,
    "/super/plantillas_produccion_masiva.html": true,
    "/super/licencias.html": true,
    "/super/licencias_codigos_descuento.html": true,
    "/super/empresas.html": true,
    "/super/administradores.html": true,
    "/super/administradores_frecuencia_fe.html": true,
    "/super/contrato.html": true,
    "/super/roles_de_usuario.html": true,
    "/super/permisos_rol.html": true,
    "/super/auditoria_super_admin.html": true,
    "/super/integracion_ia.html": true,
    "/super/chat_con_ia_global.html": true,
    "/super/voz_streaming_ia.html": true,
    "/super/configuracion_logica_del_chat_con_ia.html": true,
    "/super/contexto_ia_logica_negocio.html": true,
    "/super/alertas_sistema.html": true,
    "/super/seguridad.html": true,
    "/super/vps2.html": true,
    "/super/servidores.html": true,
    "/super/docker_portabilidad.html": true,
    "/super/soporte_remoto.html": true,
    "/super/domotica_storage.html": true,
    "/super/tickets_ayuda.html": true,
    "/super/email_corporativo.html": true,
    "/super/correos_masivos.html": true,
    "/super/recordatorios_infraestructura.html": true,
    "/super/agentes_de_mantenimiento_qutomatico.html": true,
    "/super/formato_para_emviar_email.html": true,
    "/super/mantenimiento_sistema.html": true,
    "/super/explorador_archivos.html": true,
    "/super/administrar_base_de_datos.html": true,
    "/super/configuracion_avanzada.html": true,
    "/super/configuracion/consumos.html": true,
    "/super/configuracion/rustdesk_vps.html": true,
    "/super/configuracion/limitaciones.html": true,
    "/super/configuracion/onlyoffice.html": true,
    "/super/configuracion/voz_ia.html": true,
    "/super/configuracion/epayco.html": true,
    "/super/configuracion/wompi_nequi.html": true,
    "/super/configuracion/alertas_licencia.html": true,
    "/super/configuracion/whatsapp_notificaciones.html": true,
    "/super/configuracion/whatsapp_portal.html": true,
    "/super/configuracion/recaptcha.html": true,
    "/super/configuracion/login_2fa.html": true,
    "/super/configuracion/ia_global.html": true,
    "/super/configuracion/respaldo.html": true
  };

  function auditSuperPanelEvent(action, payload) {
    try {
      var data = Object.assign({
        accion: action || "interaccion",
        modulo: "super_panel_ui",
        recurso: "super_administrador",
        endpoint: iframe ? iframe.getAttribute("src") : window.location.pathname,
        observaciones: "evento visual del panel super",
        metadata: {}
      }, payload || {});
      if (!data.metadata || typeof data.metadata !== "object") data.metadata = {};
      fetch("/super/api/auditoria?action=ui_event&scope=super_panel", {
        method: "POST",
        credentials: "same-origin",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data)
      }).catch(function () {});
    } catch (e) {}
  }

  function isAllowedSuperHref(href) {
    var normalized = normalizeHref(href).split("?")[0];
    if (normalized === "/ayuda/ayuda.html") return true;
    return !!coreSuperPages[normalized];
  }

  function findMenuLinkByHref(href) {
    var current = normalizeHref(href);
    var currentPath = current.split("?")[0];
    return links.find(function (a) {
      var linkHref = normalizeHref(a.getAttribute("href"));
      if (!linkHref) return false;
      if (linkHref === current) return true;
      return linkHref.split("?")[0] === currentPath;
    }) || null;
  }

  function getCurrentFrameHref() {
    if (!iframe) return "";
    try {
      return iframe.contentWindow.location.pathname + iframe.contentWindow.location.search;
    } catch (e) {
      return iframe.getAttribute("src") || "";
    }
  }

  function readFavorites() {
    if (!localStore) return [];
    try {
      var raw = localStore.getItem(favoritesKey) || "[]";
      var parsed = JSON.parse(raw);
      return Array.isArray(parsed) ? parsed.filter(function (item) {
        return item && isAllowedSuperHref(item.href);
      }) : [];
    } catch (e) {
      return [];
    }
  }

  function writeFavorites(favorites) {
    if (!localStore) return;
    try {
      localStore.setItem(favoritesKey, JSON.stringify(favorites.slice(0, 24)));
    } catch (e) {}
  }

  function favoriteTitleFromFrame(href) {
    var link = findMenuLinkByHref(href);
    var menuText = link ? String(link.textContent || "").trim() : "";
    if (menuText) return menuText;
    try {
      var doc = iframe && iframe.contentDocument;
      var h1 = doc && doc.querySelector("h1");
      var title = h1 ? String(h1.textContent || "").trim() : "";
      if (title) return title;
      title = doc && doc.title ? String(doc.title || "").trim() : "";
      if (title) return title;
    } catch (e) {}
    var normalized = normalizeHref(href).split("?")[0].split("/").pop() || "Pagina";
    return normalized.replace(/\.html$/i, "").replace(/[_-]+/g, " ").replace(/\b\w/g, function (c) {
      return c.toUpperCase();
    });
  }

  function favoriteIconFromMenu(href) {
    var link = findMenuLinkByHref(href);
    var img = link ? link.querySelector("img.icon") : null;
    if (img && img.getAttribute("src")) {
      return { type: "img", src: img.getAttribute("src") };
    }
    return { type: "text", value: "*" };
  }

  function isFavoriteHref(href) {
    var normalized = normalizeHref(href);
    if (!normalized) return false;
    return readFavorites().some(function (item) {
      return normalizeHref(item.href) === normalized;
    });
  }

  function notifyFavoritesChanged() {
    try {
      window.dispatchEvent(new CustomEvent("pcs-super-favorites-changed"));
    } catch (e) {}
    try {
      if (iframe && iframe.contentWindow) {
        iframe.contentWindow.postMessage({ type: "pcs-super-favorites-changed" }, window.location.origin);
      }
    } catch (e) {}
  }

  function updateFavoriteButton(href) {
    if (!favoriteBtn) return;
    var normalized = normalizeHref(href || getCurrentFrameHref());
    var available = !!normalized && isAllowedSuperHref(normalized);
    favoriteBtn.hidden = !available;
    favoriteBtn.disabled = !available;
    if (!available) {
      favoriteBtn.setAttribute("aria-pressed", "false");
      return;
    }
    var active = isFavoriteHref(normalized);
    favoriteBtn.setAttribute("aria-pressed", active ? "true" : "false");
    favoriteBtn.title = active ? "Quitar de favoritos" : "Agregar a favoritos";
    favoriteBtn.setAttribute("aria-label", active ? "Quitar pagina de super administrador de favoritos" : "Agregar pagina de super administrador a favoritos");
  }

  function toggleCurrentFavorite() {
    var href = normalizeHref(getCurrentFrameHref());
    if (!href || !isAllowedSuperHref(href)) return;
    var favorites = readFavorites();
    var exists = favorites.some(function (item) {
      return normalizeHref(item.href) === href;
    });
    if (exists) {
      favorites = favorites.filter(function (item) {
        return normalizeHref(item.href) !== href;
      });
      auditSuperPanelEvent("favorito_super_quitar", {
        recurso: href.split("?")[0],
        endpoint: href,
        metadata: { href: href }
      });
    } else {
      favorites.unshift({
        href: href,
        title: favoriteTitleFromFrame(href),
        icon: favoriteIconFromMenu(href),
        added_at: new Date().toISOString()
      });
      auditSuperPanelEvent("favorito_super_agregar", {
        recurso: href.split("?")[0],
        endpoint: href,
        metadata: { href: href }
      });
    }
    writeFavorites(favorites);
    updateFavoriteButton(href);
    notifyFavoritesChanged();
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
    updateFavoriteButton(current);
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
        updateFavoriteButton(href);
        auditSuperPanelEvent("abrir_modulo_super", {
          recurso: normalizeHref(href).split("?")[0] || "modulo_super",
          endpoint: href,
          metadata: {
            texto: (a.textContent || "").trim(),
            href: normalizeHref(href)
          }
        });
      } else {
        window.location.href = href;
      }
    });
  });

  if (iframe) {
    var defaultIframeSrc = iframe.getAttribute("src") || "/super/licencias_resumen.html";
    var initialIframeSrc = normalizeHref(defaultIframeSrc) || "/super/licencias_resumen.html";
    iframe.setAttribute("src", initialIframeSrc);
    setActiveByHref(initialIframeSrc);
    updateFavoriteButton(initialIframeSrc);
  }

  if (favoriteBtn) {
    favoriteBtn.addEventListener("click", function () {
      toggleCurrentFavorite();
    });
  }

  if (iframe) {
    iframe.addEventListener("load", function () {
      try {
        var src = iframe.contentWindow.location.pathname + iframe.contentWindow.location.search;
        persistLastPage(src);
        setActiveByHref(src);
        updateFavoriteButton(src);
      } catch (e) {
        var src2 = iframe.getAttribute("src");
        persistLastPage(src2);
        setActiveByHref(src2);
        updateFavoriteButton(src2);
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
