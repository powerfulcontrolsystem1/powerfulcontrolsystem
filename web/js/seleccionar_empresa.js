(function () {
  var empresasPanel = document.getElementById("empresasPanel");
  var contentFrame = document.getElementById("contentFrame");
  var navLinks = Array.from(document.querySelectorAll(".admin-sidebar .nav a"));

  function setActiveNav(activeLink) {
    navLinks.forEach(function (link) {
      link.classList.remove("active");
    });
    if (activeLink) activeLink.classList.add("active");
  }

  function openInRightFrame(href, link) {
    if (!href) return;
    if (!contentFrame || !empresasPanel) {
      window.location.href = href;
      return;
    }
    empresasPanel.style.display = "none";
    contentFrame.style.display = "";
    contentFrame.setAttribute("src", href);
    setActiveNav(link);
  }

  function buildEmpresaCard(empresa, hasLicense) {
    var a = document.createElement("a");
    a.href = "#";
    a.className = "card-link";
    a.addEventListener("click", function (evt) {
      evt.preventDefault();
      try {
        if (hasLicense) {
          window.open("/administrar_empresa.html?id=" + encodeURIComponent(empresa.id), "_blank");
        } else {
          var params = new URLSearchParams();
          params.set("empresa_id", empresa.id);
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
    a.appendChild(div);
    return a;
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
        if (localStorage.getItem("rememberAccount") === "1" && me && me.email) {
          localStorage.setItem("rememberedEmail", me.email);
        }
      } catch (e) {}

      var responses = await Promise.all([fetch("/super/api/empresas"), fetch("/super/api/licencias"), fetch("/super/api/tipos_empresas")]);
      var empRes = responses[0];
      var licRes = responses[1];
      var tiposRes = responses[2];

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
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  function hideForm() {
    document.getElementById("form").style.display = "none";
    document.getElementById("addBtn").style.display = "";
    document.getElementById("itemId").value = "";
    document.getElementById("nombre").value = "";
    document.getElementById("nit").value = "";
    document.getElementById("observaciones").value = "";
  }

  function wireSidebarFrameLinks() {
    var linkAgregar = document.getElementById("linkAgregarEmpresa");
    var linkLicencias = document.getElementById("linkLicencias");
    var linkAdministradores = document.getElementById("linkAdministradores");
    var linkReportes = document.getElementById("linkReportesGlobales");

    if (linkAgregar) {
      linkAgregar.addEventListener("click", function (ev) {
        ev.preventDefault();
        if (empresasPanel) empresasPanel.style.display = "";
        if (contentFrame) {
          contentFrame.style.display = "none";
          contentFrame.setAttribute("src", "about:blank");
        }
        showForm();
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
    setActiveNav(document.getElementById("linkAgregarEmpresa"));

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
        await fetch("/super/api/empresas", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload),
        });
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

  render();
})();
