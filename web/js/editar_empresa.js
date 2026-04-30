(function () {
  var state = {
    empresa: null,
    impacto: null,
    accesos: [],
    invitaciones: [],
    shareMeta: {
      is_owner: false,
      requester_email: "",
      principal_email: "",
    },
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

  function redirectToSeleccionarEmpresa() {
    try {
      if (window.parent && window.parent !== window) {
        window.parent.location.href = "/seleccionar_empresa.html";
        return;
      }
    } catch (e) {}
    window.location.href = "/seleccionar_empresa.html";
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
      var err = new Error((data && (data.error || data.message)) || raw || "Solicitud fallida");
      err.status = res.status;
      err.payload = data;
      throw err;
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

  function escapeHtml(value) {
    return String(value == null ? "" : value).replace(/[&<>"']/g, function (match) {
      return {
        "&": "&amp;",
        "<": "&lt;",
        ">": "&gt;",
        '"': "&quot;",
        "'": "&#39;",
      }[match];
    });
  }

  function normalizeEmail(value) {
    return String(value || "").trim().toLowerCase();
  }

  function isSharedEmpresa() {
    return String(state.empresa && state.empresa.access_source ? state.empresa.access_source : "owner").toLowerCase() === "shared";
  }

  function canRevokeAccess(item) {
    if (!item) return false;
    if (state.shareMeta && state.shareMeta.is_owner) return true;
    var requester = normalizeEmail(state.shareMeta && state.shareMeta.requester_email);
    return requester && (
      requester === normalizeEmail(item.admin_email) ||
      requester === normalizeEmail(item.compartido_por_email)
    );
  }

  function canRevokeInvitation(item) {
    if (!item) return false;
    if (state.shareMeta && state.shareMeta.is_owner) return true;
    var requester = normalizeEmail(state.shareMeta && state.shareMeta.requester_email);
    return requester && (
      requester === normalizeEmail(item.admin_email) ||
      requester === normalizeEmail(item.invitado_por_email)
    );
  }

  function displayPerson(name, email) {
    return String(name || email || "Administrador").trim();
  }

  function clearMessageActions() {
    var node = $("empresaShareMessageBox");
    if (!node) return;
    var existing = node.querySelector(".empresa-share-msg-actions");
    if (existing) existing.remove();
  }

  function showResendPendingInvitationAction(invitationId) {
    var node = $("empresaShareMessageBox");
    if (!node) return;
    clearMessageActions();
    var wrap = document.createElement("div");
    wrap.className = "empresa-share-msg-actions";
    wrap.style.marginTop = "10px";
    var btn = document.createElement("button");
    btn.type = "button";
    btn.className = "btn secondary";
    btn.textContent = "Reenviar invitación";
    btn.onclick = function () {
      handleShareAction("resend", invitationId, "invitation");
    };
    wrap.appendChild(btn);
    node.appendChild(wrap);
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
    var isShared = String(empresa.access_source || "owner").toLowerCase() === "shared";
    var shareCard = document.querySelector('.empresa-edit-share-card');
    var shareTitle = shareCard ? shareCard.querySelector('h2') : null;
    var shareCopy = shareCard ? shareCard.querySelector('.empresa-edit-panel-copy') : null;
    var shareForm = $("empresaShareForm");
    var saveButton = document.querySelector('#empresaEditForm button[type="submit"]');
    $("empresaEditTitle").textContent = empresa.nombre || "Editar empresa";
    $("empresaEditSubtitle").textContent = isShared
      ? "Tienes acceso compartido a esta empresa. Puedes consultarla, pero solo el administrador propietario puede modificar la ficha o administrar invitaciones."
      : "Gestiona el nombre y la descripcion de " + (empresa.nombre || "la empresa") + ", o elimínala por completo desde este mismo panel.";
    $("empresaNombre").value = empresa.nombre || "";
    $("empresaTipo").value = empresa.tipo_nombre || "No definido";
    $("empresaObservaciones").value = empresa.observaciones || "";
    $("empresaDeleteConfirm").placeholder = empresa.nombre || "";
    $("empresaTipoMeta").textContent = empresa.tipo_nombre || "No definido";
    $("empresaNitMeta").textContent = empresa.nit || "Sin NIT";
    $("empresaEstadoMeta").textContent = empresa.estado || "activo";
    $("empresaNombre").disabled = !!isShared;
    $("empresaObservaciones").disabled = !!isShared;
    $("empresaDeleteConfirm").disabled = !!isShared;
    if (saveButton) {
      saveButton.disabled = !!isShared;
    }
    if ($("empresaDeleteBtn")) {
      $("empresaDeleteBtn").disabled = !!isShared;
    }
    if (shareCard) {
      shareCard.style.display = '';
    }
    if (shareTitle) {
      shareTitle.textContent = isShared ? "Administradores con acceso" : "Compartir con otro administrador";
    }
    if (shareCopy) {
      shareCopy.textContent = isShared
        ? "Consulta quien compartio esta empresa y quien tiene acceso administrativo. Puedes quitar tu propio acceso compartido desde esta vista."
        : "Envia una invitacion por correo a otro administrador ya registrado para que, despues de aceptarla, tambien vea esta empresa en su selector.";
    }
    if (shareForm) {
      shareForm.style.display = isShared ? 'none' : '';
    }
    if (isShared) {
      setMessage("empresaEditMessage", "Esta empresa fue compartida contigo. La edición estructural solo está disponible para el propietario.", true);
      setMessage("empresaDeleteMessage", "La eliminación total está deshabilitada para accesos compartidos.", true);
    }
  }

  function renderSummary() {
    var impacto = state.impacto || {};
    $("empresaEditUsers").textContent = impacto.usuarios_activos != null ? String(impacto.usuarios_activos) : "0";
    $("empresaEditCarritos").textContent = impacto.carritos_abiertos != null ? String(impacto.carritos_abiertos) : "0";
    $("empresaEditReservas").textContent = impacto.reservas_vigentes != null ? String(impacto.reservas_vigentes) : "0";
    $("empresaEditLicencias").textContent = impacto.licencias_activas != null ? String(impacto.licencias_activas) : "0";
    $("empresaEditImpacto").textContent = buildImpactoTexto(impacto);
  }

  function normalizeInviteStatus(item) {
    var status = String(item && item.estado ? item.estado : "pendiente").trim().toLowerCase();
    return status || "pendiente";
  }

  function buildShareActionButton(label, action, id, kind) {
    return '<button type="button" class="btn secondary empresa-share-action" data-action="' + action + '" data-id="' + String(id || '') + '" data-kind="' + kind + '">' + label + '</button>';
  }

  function bindShareActions() {
    Array.prototype.forEach.call(document.querySelectorAll('.empresa-share-action'), function (btn) {
      btn.onclick = function () {
        handleShareAction(btn.getAttribute('data-action'), btn.getAttribute('data-id'), btn.getAttribute('data-kind'));
      };
    });
  }

  function renderShares() {
    var accessList = $("empresaShareAccessList");
    var inviteList = $("empresaShareInviteList");
    var ownerCanManage = !!(state.shareMeta && state.shareMeta.is_owner);
    if (accessList) {
      if (!state.accesos.length) {
        accessList.innerHTML = '<p class="muted">No hay administradores con acceso compartido activo.</p>';
      } else {
        accessList.innerHTML = state.accesos.map(function (item) {
          var sharedTo = displayPerson(item.admin_name, item.admin_email);
          var sharedBy = displayPerson(item.compartido_por_name, item.compartido_por_email);
          var accepted = String(item.fecha_aceptada || item.fecha_creacion || '').trim();
          var status = String(item.estado || 'activo').trim().toLowerCase();
          var action = canRevokeAccess(item)
            ? buildShareActionButton(isSharedEmpresa() && normalizeEmail(item.admin_email) === normalizeEmail(state.shareMeta.requester_email) ? 'Eliminar mi acceso' : 'Revocar', 'revoke', item.id, 'access')
            : '';
          return '<article class="empresa-share-item">'
            + '<div><strong>' + escapeHtml(sharedTo) + '</strong><div class="muted">' + escapeHtml(item.admin_email || '') + '</div>'
            + '<div class="muted">Compartido por: ' + escapeHtml(sharedBy) + (accepted ? ' - Desde: ' + escapeHtml(accepted) : '') + '</div></div>'
            + '<div class="empresa-share-item-actions"><span class="empresa-share-state is-' + escapeHtml(status || 'activo') + '">' + escapeHtml(status || 'activo') + '</span>'
            + action
            + '</div></article>';
        }).join('');
      }
    }
    if (inviteList) {
      if (!state.invitaciones.length) {
        inviteList.innerHTML = '<p class="muted">No hay invitaciones registradas para esta empresa.</p>';
      } else {
        inviteList.innerHTML = state.invitaciones.map(function (item) {
          var status = normalizeInviteStatus(item);
          var invitedTo = displayPerson(item.admin_name, item.admin_email);
          var invitedBy = displayPerson(item.invitado_por_name, item.invitado_por_email);
          var expira = String(item.expira_en || '').trim();
          var actions = '';
          if (ownerCanManage && (status === 'pendiente' || status === 'expirada')) {
            actions += buildShareActionButton('Reenviar', 'resend', item.id, 'invitation');
          }
          if (canRevokeInvitation(item)) {
            actions += buildShareActionButton('Revocar', 'revoke', item.id, 'invitation');
          }
          return '<article class="empresa-share-item">'
            + '<div><strong>' + escapeHtml(invitedTo) + '</strong><div class="muted">' + escapeHtml(item.admin_email || '') + '</div>'
            + '<div class="muted">Invitado por: ' + escapeHtml(invitedBy) + (expira ? ' - Expira: ' + escapeHtml(expira) : '') + '</div></div>'
            + '<div class="empresa-share-item-actions"><span class="empresa-share-state is-' + escapeHtml(status) + '">' + escapeHtml(status) + '</span>'
            + actions
            + '</div></article>';
        }).join('');
      }
    }
    bindShareActions();
  }

  async function loadShares() {
    if (!state.empresa) return;
    var data = await fetchJSON('/super/api/empresas/compartidos?empresa_id=' + encodeURIComponent(state.empresa.id), { credentials: 'same-origin' });
    state.accesos = Array.isArray(data && data.accesos) ? data.accesos : [];
    state.invitaciones = Array.isArray(data && data.invitaciones) ? data.invitaciones : [];
    state.shareMeta = {
      is_owner: !!(data && data.is_owner),
      requester_email: String(data && data.requester_email ? data.requester_email : '').trim(),
      principal_email: String(data && data.principal_email ? data.principal_email : '').trim(),
    };
    renderShares();
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
    try {
      await loadShares();
    } catch (err) {
      setMessage("empresaShareMessageBox", err.message || "No se pudo cargar el estado de empresas compartidas.", true);
    }
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
        redirectToSeleccionarEmpresa();
      }, 900);
    } catch (err) {
      setMessage("empresaDeleteMessage", err.message || "No se pudo eliminar la empresa.", true);
    }
  }

  async function inviteAdmin(ev) {
    ev.preventDefault();
    if (!state.empresa) return;
    if (isSharedEmpresa()) {
      setMessage("empresaShareMessageBox", "Solo el propietario puede invitar nuevos administradores.", true);
      return;
    }

    var email = $("empresaShareEmail").value.trim();
    var mensaje = $("empresaShareMessage").value.trim();
    if (!email) {
      setMessage("empresaShareMessageBox", "Debes indicar el correo del administrador.", true);
      return;
    }

    try {
      clearMessageActions();
      var data = await fetchJSON('/super/api/empresas/compartidos', {
        method: 'POST',
        credentials: 'same-origin',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ empresa_id: state.empresa.id, email: email, mensaje: mensaje })
      });
      setMessage("empresaShareMessageBox", data && data.message ? data.message : 'Invitación creada correctamente.', false);
      $("empresaShareEmail").value = '';
      $("empresaShareMessage").value = '';
      await loadShares();
    } catch (err) {
      var payload = err && err.payload ? err.payload : null;
      if (payload && String(payload.code || "") === "invitation_pending" && payload.invitation_id) {
        setMessage("empresaShareMessageBox", (payload.error || err.message || "Ya existe una invitación pendiente.") + " Puedes reenviarla.", true);
        showResendPendingInvitationAction(payload.invitation_id);
        return;
      }
      clearMessageActions();
      setMessage("empresaShareMessageBox", err.message || 'No se pudo enviar la invitación.', true);
    }
  }

  async function handleShareAction(action, id, kind) {
    if (!state.empresa || !id) return;
    try {
      if (action === 'resend') {
        var resendData = await fetchJSON('/super/api/empresas/compartidos?id=' + encodeURIComponent(id) + '&action=reenviar', {
          method: 'PUT',
          credentials: 'same-origin'
        });
        setMessage('empresaShareMessageBox', resendData && resendData.message ? resendData.message : 'Invitación reenviada.', false);
      } else if (action === 'revoke') {
        await fetchJSON('/super/api/empresas/compartidos?id=' + encodeURIComponent(id) + '&kind=' + encodeURIComponent(kind), {
          method: 'DELETE',
          credentials: 'same-origin'
        });
        setMessage('empresaShareMessageBox', kind === 'access' ? 'Acceso compartido revocado.' : 'Invitación revocada.', false);
        if (kind === 'access' && isSharedEmpresa()) {
          window.setTimeout(function () {
            redirectToSeleccionarEmpresa();
          }, 900);
          return;
        }
      }
      await loadShares();
    } catch (err) {
      setMessage('empresaShareMessageBox', err.message || 'No se pudo completar la acción sobre el acceso compartido.', true);
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
    var shareForm = $("empresaShareForm");
    if (shareForm) {
      shareForm.addEventListener("submit", inviteAdmin);
    }
    loadEmpresa().catch(function (err) {
      setMessage("empresaEditMessage", err.message || "No se pudo cargar la empresa.", true);
      setMessage("empresaDeleteMessage", err.message || "No se pudo cargar la empresa.", true);
      setMessage("empresaShareMessageBox", err.message || "No se pudo cargar el estado de empresas compartidas.", true);
    });
  });
})();
