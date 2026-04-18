(function () {
  var state = {
    empresa: null,
    impacto: null,
  };

  function $(id) {
    return document.getElementById(id);
  }

  function getEmpresaId() {
    var params = new URLSearchParams(window.location.search || "");
    var raw = params.get("id") || params.get("empresa_id") || "";
    var id = parseInt(raw, 10);
    return Number.isFinite(id) && id > 0 ? id : 0;
  }

  async function fetchJSON(url, options) {
    var res = await fetch(url, options || { credentials: "same-origin" });
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

  function setMessage(id, text, isError) {
    var node = $(id);
    if (!node) return;
    node.textContent = text || "";
    node.classList.toggle("error", !!isError);
    node.classList.toggle("success", !isError && !!text);
  }

  function buildImpactoTexto(impacto) {
    if (!impacto) {
      return "No se detectaron bloqueos operativos previos a la eliminacion.";
    }
    var mensajes = [];
    if ((impacto.usuarios_activos || 0) > 0) mensajes.push("Hay usuarios activos vinculados a esta empresa.");
    if ((impacto.carritos_abiertos || 0) > 0) mensajes.push("Existen carritos abiertos que tambien seran purgados.");
    if ((impacto.reservas_vigentes || 0) > 0) mensajes.push("Se detectaron reservas vigentes dentro del alcance de borrado.");
    if ((impacto.licencias_activas || 0) > 0) mensajes.push("La empresa conserva licencias activas que se eliminaran junto con la empresa.");
    return mensajes.length ? mensajes.join(" ") : "No se detectaron bloqueos operativos previos a la eliminacion.";
  }

  function renderEmpresa() {
    var empresa = state.empresa;
    if (!empresa) return;
    $("empresaEditTitle").textContent = empresa.nombre || "Editar empresa";
    $("empresaEditSubtitle").textContent = "Gestiona el nombre y la descripcion de " + (empresa.nombre || "la empresa") + ", o elimínala por completo desde este mismo panel.";
    $("empresaNombre").value = empresa.nombre || "";
    $("empresaTipo").value = empresa.tipo_nombre || "No definido";
    $("empresaObservaciones").value = empresa.observaciones || "";
    $("empresaDeleteConfirm").placeholder = empresa.nombre || "";
    $("empresaTipoMeta").textContent = empresa.tipo_nombre || "No definido";
    $("empresaNitMeta").textContent = empresa.nit || "Sin NIT";
    $("empresaEstadoMeta").textContent = empresa.estado || "activo";
  }

  function renderSummary() {
    var impacto = state.impacto || {};
    $("empresaEditUsers").textContent = impacto.usuarios_activos != null ? String(impacto.usuarios_activos) : "0";
    $("empresaEditCarritos").textContent = impacto.carritos_abiertos != null ? String(impacto.carritos_abiertos) : "0";
    $("empresaEditReservas").textContent = impacto.reservas_vigentes != null ? String(impacto.reservas_vigentes) : "0";
    $("empresaEditLicencias").textContent = impacto.licencias_activas != null ? String(impacto.licencias_activas) : "0";
    $("empresaEditImpacto").textContent = buildImpactoTexto(impacto);
  }

  async function loadEmpresa() {
    var empresaId = getEmpresaId();
    if (!empresaId) {
      throw new Error("Empresa invalida");
    }
    var empresa = await fetchJSON("/super/api/empresas?id=" + encodeURIComponent(empresaId), { credentials: "same-origin" });
    var impacto = await fetchJSON("/super/api/empresas?id=" + encodeURIComponent(empresaId) + "&action=impacto_desactivacion", { credentials: "same-origin" });

    state.empresa = empresa;
    state.impacto = impacto && impacto.impacto ? impacto.impacto : null;
    renderEmpresa();
    renderSummary();
  }

  async function saveEmpresa(ev) {
    ev.preventDefault();
    if (!state.empresa) return;

    var payload = {
      tipo_id: state.empresa.tipo_id || 0,
      tipo_nombre: state.empresa.tipo_nombre || "",
      nombre: $("empresaNombre").value.trim(),
      nit: state.empresa.nit || "",
      observaciones: $("empresaObservaciones").value.trim(),
    };

    if (!payload.nombre) {
      setMessage("empresaEditMessage", "El nombre es obligatorio.", true);
      return;
    }

    try {
      await fetchJSON("/super/api/empresas?id=" + encodeURIComponent(state.empresa.id), {
        method: "PUT",
        credentials: "same-origin",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      state.empresa.nombre = payload.nombre;
      state.empresa.observaciones = payload.observaciones;
      renderEmpresa();
      setMessage("empresaEditMessage", "Cambios guardados correctamente.", false);
    } catch (err) {
      setMessage("empresaEditMessage", err.message || "No se pudo guardar la empresa.", true);
    }
  }

  async function deleteEmpresa() {
    if (!state.empresa) return;
    var confirmacion = $("empresaDeleteConfirm").value.trim();
    if (!confirmacion) {
      setMessage("empresaDeleteMessage", "Debes escribir el nombre de la empresa.", true);
      return;
    }
    if (confirmacion !== (state.empresa.nombre || "")) {
      setMessage("empresaDeleteMessage", "El nombre digitado no coincide exactamente.", true);
      return;
    }
    if (!window.confirm("Esta accion eliminara todos los datos de la empresa. ¿Deseas continuar?")) {
      return;
    }

    try {
      var data = await fetchJSON(
        "/super/api/empresas?id=" + encodeURIComponent(state.empresa.id) + "&action=eliminar_total",
        {
          method: "DELETE",
          credentials: "same-origin",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ confirmacion_nombre: confirmacion }),
        }
      );
      var result = data && data.result ? data.result : null;
      setMessage(
        "empresaDeleteMessage",
        result
          ? "Empresa eliminada. Tablas afectadas: " + result.tablas_afectadas + ". Registros eliminados: " + result.registros_eliminados + "."
          : "Empresa eliminada correctamente.",
        false
      );
      window.setTimeout(function () {
        window.location.href = "/seleccionar_empresa.html";
      }, 900);
    } catch (err) {
      setMessage("empresaDeleteMessage", err.message || "No se pudo eliminar la empresa.", true);
    }
  }

  document.addEventListener("DOMContentLoaded", function () {
    var form = $("empresaEditForm");
    if (form) {
      form.addEventListener("submit", saveEmpresa);
    }
    var deleteBtn = $("empresaDeleteBtn");
    if (deleteBtn) {
      deleteBtn.addEventListener("click", deleteEmpresa);
    }
    loadEmpresa().catch(function (err) {
      setMessage("empresaEditMessage", err.message || "No se pudo cargar la empresa.", true);
      setMessage("empresaDeleteMessage", err.message || "No se pudo cargar la empresa.", true);
    });
  });
})();