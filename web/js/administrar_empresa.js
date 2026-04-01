function getQueryParam(name) {
  var params = new URLSearchParams(window.location.search);
  return params.get(name);
}

(function () {
  var id = getQueryParam("id");
  var title = document.getElementById("empresaTitle");
  var frame = document.getElementById("contentFrame");
  var links = [
    document.getElementById("linkInicio"),
    document.getElementById("linkVentas"),
    document.getElementById("linkCarritoCompras"),
    document.getElementById("linkProductos"),
    document.getElementById("linkConfiguracion"),
    document.getElementById("linkUsuarios"),
    document.getElementById("linkClientes"),
    document.getElementById("linkConfigAvanzada"),
    document.getElementById("linkConfigEstaciones"),
    document.getElementById("linkEstaciones"),
    document.getElementById("linkReportes"),
  ];

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
    });
  }

  if (id) {
    setLinksWithEmpresa(id);
    if (frame) {
      frame.src = "/administrar_empresa/administrar_productos.html?empresa_id=" + encodeURIComponent(id);
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
    frame.src = "/administrar_empresa/inicio.html";
  }
  title.textContent = "Administrar Empresa";
})();
