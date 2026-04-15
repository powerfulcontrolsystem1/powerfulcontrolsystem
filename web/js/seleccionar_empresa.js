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

  function buildEmpresaCard(empresa, hasLicense) {
    var estadoRaw = String(empresa && empresa.estado ? empresa.estado : "activo").toLowerCase();
    var empresaActiva = estadoRaw !== "inactivo";

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

    var div = document.createElement("div");
    div.className = "portal-card warm";
    div.innerHTML =
      '<div class="card-body">' +
      '<h3 class="card-title">' +
      escapeHtml(empresa.nombre || "--") +
      "</h3>" +
      '<p class="card-desc muted">' +
      escapeHtml(empresa.observaciones || "") +
      "</p>" +
      '<div class="card-actions">' +
      '<button class="license-indicator ' +
      (hasLicense ? "active" : "inactive") +
      '" type="button" aria-hidden="true">' +
      (hasLicense ? "Licencia activa" : "Sin licencia") +
      "</button>" +
      "</div>" +
      "</div>";

    if (!hasLicense) {
      var dlDiv = document.createElement('div');
      dlDiv.className = 'card-download';
      var dlBtn = document.createElement('button');
      dlBtn.type = 'button';
      // Reuse license visual style so download button coincides with 'Licencia activa'
      dlBtn.className = 'license-indicator active download-data';
      dlBtn.setAttribute('data-empresa-id', String(empresa.id || ''));
      dlBtn.setAttribute('data-empresa-name', String(empresa.nombre || ''));
      dlBtn.setAttribute('aria-label', 'Descargar datos de ' + String(empresa.nombre || ''));
      dlBtn.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="16" height="16" aria-hidden="true" focusable="false"><path fill="currentColor" d="M12 3v10l4-4-1.4-1.4L13 9.2V3h-2zM5 18v2h14v-2H5z"/></svg><span class="download-label">Descargar</span>';
      dlDiv.appendChild(dlBtn);
      div.appendChild(dlDiv);
    }

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
