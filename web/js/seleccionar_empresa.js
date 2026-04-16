(function () {
  var empresasPanel = document.getElementById("empresasPanel");
  var contentFrame = document.getElementById("contentFrame");
  var navLinks = Array.from(document.querySelectorAll(".admin-sidebar .nav a"));
  var storage = null;
  var viewKey = "seleccionar_empresa:view";

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
    if (!contentFrame || !empresasPanel) {
      window.location.href = normalized;
      return;
    }
    empresasPanel.style.display = "none";
    contentFrame.style.display = "";
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

  function buildEmpresaCard(empresa, hasLicense) {
    var estadoRaw = String(empresa && empresa.estado ? empresa.estado : "activo").toLowerCase();
    var empresaActiva = estadoRaw !== "inactivo";
    var visual = getEmpresaTypeVisual(empresa);
    var nitLabel = String(empresa && empresa.nit ? empresa.nit : "").trim() || "Sin NIT registrado";
    var accessLabel = empresaActiva ? (hasLicense ? "Panel habilitado" : "Licencia pendiente") : "Operacion pausada";
    var subtitle = empresaActiva ? (hasLicense ? "Acceso operativo disponible" : "Preparada para activar licencia") : "Empresa en pausa administrativa";
    var ctaLabel = hasLicense ? "Abrir administracion" : "Elegir licencia";
    var statusLabel = empresaActiva ? "Empresa activa" : "Empresa inactiva";
    var description = buildEmpresaCardDescription(empresa, visual, hasLicense);

    var a = document.createElement("a");
    a.href = "#";
    a.className = "card-link";
    a.addEventListener("click", function (evt) {
      evt.preventDefault();
      persistEmpresaContext(empresa.id);
      try {
        if (hasLicense) {
          var adminURL =
            "/administrar_empresa.html?id=" + encodeURIComponent(empresa.id) +
            "&empresa_id=" + encodeURIComponent(empresa.id);
          var opened = window.open(adminURL, "_blank");
          if (!opened) {
            window.location.href = adminURL;
          }
        } else {
          var params = new URLSearchParams();
          params.set("empresa_id", empresa.id);
          params.set("id", empresa.id);
          if (empresa.tipo_id) params.set("tipo_id", empresa.tipo_id);
          if (empresa.tipo_nombre) params.set("tipo_nombre", empresa.tipo_nombre);
          window.location.href = "/elegir_licencia.html?" + params.toString();
        }
      } catch (err) {
        console.error(err);
      }
    });

    var downloadHTML = "";
    if (!hasLicense) {
      downloadHTML =
        '<div class="card-download">' +
        '<button class="license-indicator active download-data" type="button" data-empresa-id="' + escapeHtml(String(empresa.id || "")) + '" data-empresa-name="' + escapeHtml(String(empresa.nombre || "")) + '" aria-label="Descargar datos de ' + escapeHtml(String(empresa.nombre || "")) + '">' +
        '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="16" height="16" aria-hidden="true" focusable="false"><path fill="currentColor" d="M12 3v10l4-4-1.4-1.4L13 9.2V3h-2zM5 18v2h14v-2H5z"/></svg><span class="download-label">Descargar datos</span>' +
        "</button>" +
        "</div>";
    }

    var div = document.createElement("div");
    div.className = "portal-card warm empresa-card";
    div.setAttribute("data-tone", visual.tone);
    div.innerHTML =
      '<div class="empresa-card-shell">' +
      '<div class="empresa-card-top">' +
      '<div class="empresa-card-icon-shell">' +
      '<img class="empresa-card-icon" src="' + escapeHtml(visual.icon) + '" alt="' + escapeHtml(visual.alt) + '">' +
      "</div>" +
      '<div class="empresa-card-badge-stack">' +
      '<span class="empresa-card-chip empresa-card-chip--type">' + escapeHtml(visual.label) + "</span>" +
      '<span class="empresa-card-chip empresa-card-chip--status ' + (empresaActiva ? "is-active" : "is-inactive") + '">' + escapeHtml(statusLabel) + "</span>" +
      "</div>" +
      "</div>" +
      '<div class="card-body empresa-card-body">' +
      '<div class="empresa-card-heading">' +
      '<p class="empresa-card-eyebrow">' + escapeHtml(visual.eyebrow) + "</p>" +
      '<h3 class="card-title">' + escapeHtml(empresa.nombre || "--") + "</h3>" +
      '<p class="empresa-card-subtitle">' + escapeHtml(subtitle) + "</p>" +
      "</div>" +
      '<p class="card-desc muted">' + escapeHtml(description) + "</p>" +
      '<div class="empresa-card-meta">' +
      '<span class="empresa-card-meta-item"><strong>NIT</strong><span>' + escapeHtml(nitLabel) + "</span></span>" +
      '<span class="empresa-card-meta-item"><strong>Acceso</strong><span>' + escapeHtml(accessLabel) + "</span></span>" +
      "</div>" +
      '<div class="card-actions empresa-card-actions">' +
      '<span class="empresa-card-cta">' + escapeHtml(ctaLabel) + "</span>" +
      '<span class="license-indicator ' + (hasLicense ? "active" : "inactive") + '" aria-hidden="true">' + (hasLicense ? "Licencia activa" : "Sin licencia") + "</span>" +
      "</div>" +
      "</div>" +
      downloadHTML +
      "</div>";

    a.appendChild(div);
    return a;
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
    header.innerHTML = "<h2>" + title + "</h2><span class=\"form-help\">Total: " + empresas.length + "</span>";

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
      var meRes = await fetch("/me");
      if (!meRes.ok) {
        window.location.href = "/login.html";
        return;
      }
      var me = await meRes.json();
      try {
        var rememberedEmail = me && me.email ? String(me.email).trim() : "";
        if (localStorage.getItem("rememberAccount") === "1" && /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(rememberedEmail)) {
          localStorage.setItem("rememberedEmail", rememberedEmail);
        }
      } catch (e) {}

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

      var myEmpresas = empresas.filter(function (e) {
        if (!e.usuario_creador) return false;
        return e.usuario_creador.toLowerCase() === (me.email || "").toLowerCase();
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

      var list = myEmpresas.length > 0 ? myEmpresas : empresas;
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
    if (empresasPanel) empresasPanel.style.display = "";
    if (contentFrame) {
      contentFrame.style.display = "none";
      contentFrame.setAttribute("src", "about:blank");
    }
    document.getElementById("form").style.display = "";
    document.getElementById("addBtn").style.display = "none";
    persistView({ mode: "form" });
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  function hideForm() {
    document.getElementById("form").style.display = "none";
    document.getElementById("addBtn").style.display = "";
    document.getElementById("itemId").value = "";
    document.getElementById("nombre").value = "";
    document.getElementById("nit").value = "";
    document.getElementById("observaciones").value = "";
    persistView({ mode: "empresas" });
  }

  function showEmpresasPanel() {
    if (empresasPanel) empresasPanel.style.display = "";
    if (contentFrame) {
      contentFrame.style.display = "none";
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
    var linkReportes = document.getElementById("linkReportesGlobales");

    if (linkAgregar) {
      linkAgregar.addEventListener("click", function (ev) {
        ev.preventDefault();
        showEmpresasPanel();
        setActiveNav(linkAgregar);
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

  document.addEventListener("DOMContentLoaded", function () {
    wireSidebarFrameLinks();
    restoreLastView();

    var form = document.getElementById("form");
    if (!form) return;

    form.onsubmit = async function (e) {
      e.preventDefault();
      var payload = {
        tipo_id: 0,
        tipo_nombre: document.getElementById("tipo_id").value || "",
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

  function escapeHtml(s) {
    return String(s).replace(/[&<>"']/g, function (m) {
      return {
        "&": "&amp;",
        "<": "&lt;",
        ">": "&gt;",
        '"': "&quot;",
        "'": "&#39;",
      }[m];
    });
  }

  function sanitizeFilename(name) {
    if (!name) return '';
    return String(name).replace(/[^a-z0-9\-\_\.]/gi, '_');
  }

  function downloadBlob(blob, filename) {
    var url = URL.createObjectURL(blob);
    var a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }

  async function performEmpresaDownload(empresa, format) {
    if (!empresa || !empresa.id) throw new Error('empresa inválida');
    var payload = {
      empresa_id: empresa.id,
      nombre: 'Backup export ' + (empresa.nombre || ''),
      descripcion: 'Exportado desde UI',
      include_tables: [],
      exclude_tables: []
    };

    var createUrl = '/api/empresa/backups?empresa_id=' + encodeURIComponent(empresa.id) + '&action=crear';
    var createRes = await fetch(createUrl, {
      method: 'POST',
      credentials: 'same-origin',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    var createText = await createRes.text();
    var createData = {};
    if (createText) {
      try { createData = JSON.parse(createText); } catch (e) { /* ignore */ }
    }
    if (!createRes.ok) {
      throw new Error((createData && (createData.error || createData.message)) || createText || ('HTTP ' + createRes.status));
    }
    var backup = createData.backup || {};
    var backupId = backup.id || 0;
    if (!backupId) throw new Error('No se pudo crear backup');

    var exportFormat = format;
    var needGzip = false;
    if (format === 'json.gz') { exportFormat = 'json'; needGzip = true; }

    var exportUrl = '/api/empresa/backups?action=export&id=' + encodeURIComponent(backupId) + '&empresa_id=' + encodeURIComponent(empresa.id) + '&format=' + encodeURIComponent(exportFormat);
    var expRes = await fetch(exportUrl, { credentials: 'same-origin' });
    if (!expRes.ok) {
      var tx = await expRes.text().catch(function(){return '';});
      throw new Error(tx || ('HTTP ' + expRes.status));
    }

    var blob = await expRes.blob();
    var ext = exportFormat === 'json' ? 'json' : (exportFormat === 'csv' ? 'csv' : (exportFormat === 'xls' ? 'xls' : exportFormat));
    var nameSafe = sanitizeFilename(empresa.nombre || ('empresa_' + String(empresa.id)));
    var now = new Date();
    var stamp = now.getFullYear().toString() + String(now.getMonth()+1).padStart(2,'0') + String(now.getDate()).padStart(2,'0') + '_' + String(now.getHours()).padStart(2,'0') + String(now.getMinutes()).padStart(2,'0') + String(now.getSeconds()).padStart(2,'0');
    var filename = nameSafe + '_' + stamp + '.' + ext;

    if (needGzip) {
      if (typeof CompressionStream === 'function') {
        try {
          var cs = new CompressionStream('gzip');
          var compressedStream = blob.stream().pipeThrough(cs);
          var compressedBlob = await new Response(compressedStream).blob();
          downloadBlob(compressedBlob, nameSafe + '_' + stamp + '.json.gz');
          return;
        } catch (e) {
          // fallback to plain json
        }
      }
      // fallback: download original json
      downloadBlob(blob, filename);
      return;
    }

    downloadBlob(blob, filename);
  }

  function showDownloadDialog(empresa) {
    var overlay = document.createElement('div');
    overlay.className = 'modal-overlay';
    overlay.innerHTML = '' +
      '<div class="modal" role="dialog" aria-modal="true" style="max-width:420px;margin:60px auto;padding:18px;background:var(--card);border-radius:10px;box-shadow:0 20px 50px rgba(0,0,0,0.6);">' +
        '<h3 style="margin-top:0;margin-bottom:8px;color:var(--text)">Descargar datos — ' + escapeHtml(empresa.nombre || '') + '</h3>' +
        '<p class="form-help">Selecciona el formato para descargar los datos completos de la empresa.</p>' +
        '<div style="margin-top:12px;display:flex;gap:8px">' +
          '<select class="format" style="flex:1;padding:8px;border-radius:8px">' +
            '<option value="json">JSON (snapshot completo)</option>' +
            '<option value="json.gz">JSON (comprimido .gz)</option>' +
            '<option value="csv">CSV (resumen)</option>' +
            '<option value="xls">Excel (resumen)</option>' +
          '</select>' +
        '</div>' +
        '<div style="margin-top:14px;display:flex;justify-content:flex-end;gap:8px">' +
          '<button class="btn secondary cancel">Cancelar</button>' +
          '<button class="btn confirm">Descargar</button>' +
        '</div>' +
      '</div>';
    overlay.addEventListener('click', function(ev){ if (ev.target === overlay) overlay.remove(); });
    document.body.appendChild(overlay);

    overlay.querySelector('.cancel').addEventListener('click', function(){ overlay.remove(); });
    overlay.querySelector('.confirm').addEventListener('click', async function(){
      try {
        var fmt = overlay.querySelector('select.format').value || 'json';
        overlay.querySelector('.confirm').disabled = true;
        overlay.querySelector('.cancel').disabled = true;
        await performEmpresaDownload(empresa, fmt);
      } catch (err) {
        alert('Error: ' + (err && err.message ? err.message : err));
      } finally {
        overlay.remove();
      }
    });
  }

  document.addEventListener('click', function(ev){
    var btn = ev.target.closest && ev.target.closest('button.download-data');
    if (!btn) return;
    ev.preventDefault();
    ev.stopPropagation();
    var id = parseInt(btn.getAttribute('data-empresa-id') || '0', 10);
    var name = btn.getAttribute('data-empresa-name') || '';
    showDownloadDialog({ id: id, nombre: name });
  });

  render();
})();
