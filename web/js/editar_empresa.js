(function () {
  var state = {
    empresa: null,
    impacto: null,
    licencias: null,
    accesos: [],
    invitaciones: [],
    shareScopeCatalog: [],
    shareMeta: {
      is_owner: false,
      requester_email: "",
      principal_email: "",
    },
    deleteDownloadOffered: false,
    deleting: false,
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

  function buildEmpresaDownloadUrl() {
    var empresaId = state.empresa && state.empresa.id ? state.empresa.id : getEmpresaId();
    return "/descargar_informacion_de_la_empresa.html?empresa_id=" + encodeURIComponent(empresaId);
  }

  function removeEmpresaFromSelectorOrderLocal(empresaId) {
    var id = String(empresaId || "").trim();
    if (!id || !window.localStorage) return;
    try {
      for (var index = 0; index < window.localStorage.length; index += 1) {
        var key = window.localStorage.key(index);
        if (!key || key.indexOf("seleccionar_empresa:orden:") !== 0) continue;
        var raw = window.localStorage.getItem(key);
        if (!raw) continue;
        var values = JSON.parse(raw);
        if (!Array.isArray(values)) continue;
        var next = [];
        var changed = false;
        var seen = {};
        values.forEach(function (value) {
          var normalized = String(parseInt(value, 10) || "");
          if (!normalized || seen[normalized]) return;
          seen[normalized] = true;
          if (normalized === id) {
            changed = true;
            return;
          }
          next.push(Number(normalized));
        });
        if (changed) {
          window.localStorage.setItem(key, JSON.stringify(next));
        }
      }
    } catch (e) {}
  }

  function openEmpresaDownload(options) {
    var opts = options && typeof options === "object" ? options : {};
    var sameWindowFallback = opts.sameWindowFallback !== false;
    var showMessage = opts.showMessage !== false;
    state.deleteDownloadOffered = true;
    var url = buildEmpresaDownloadUrl();
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
    if (!opened && sameWindowFallback) {
      window.location.href = url;
      return true;
    }
    if (showMessage) {
      setMessage("empresaDeleteMessage", "Se abrio la pagina de descarga. Cuando termines, vuelve aqui y confirma la eliminacion si deseas continuar.", false);
    }
    return opened;
  }

  function confirmDownloadBeforeEmpresaDelete() {
    var wantsDownload = window.confirm(
      "Antes de eliminar esta empresa, deseas descargar toda su informacion?\n\n" +
      "Aceptar: abre la descarga en una nueva pestana y luego continua la eliminacion.\n" +
      "Cancelar: continua la eliminacion sin descargar ahora."
    );
    if (!wantsDownload) return true;
    if (!openEmpresaDownload({ sameWindowFallback: false, showMessage: false })) {
      setMessage("empresaDeleteMessage", "No se pudo abrir la descarga automaticamente. Usa el boton Descargar informacion antes de eliminar y vuelve a confirmar.", true);
      return false;
    }
    setMessage("empresaDeleteMessage", "Se abrio la descarga. Conserva el archivo; la eliminacion continuara.", false);
    return true;
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

  function setNodeState(id, ok, text) {
    var node = $(id);
    if (!node) return;
    node.textContent = text || "";
    node.classList.toggle("is-ok", !!ok);
    node.classList.toggle("is-pending", !ok);
  }

  function getDeleteValidationState() {
    var expectedName = String(state.empresa && state.empresa.nombre || "").trim();
    var nameOk = String($("empresaDeleteConfirm") && $("empresaDeleteConfirm").value || "").trim() === expectedName && expectedName !== "";
    var phraseOk = String($("empresaDeletePhrase") && $("empresaDeletePhrase").value || "").trim().toUpperCase() === "ELIMINAR";
    var riskOk = !!($("empresaDeleteAcknowledge") && $("empresaDeleteAcknowledge").checked);
    return {
      nameOk: nameOk,
      phraseOk: phraseOk,
      riskOk: riskOk,
      ready: nameOk && phraseOk && riskOk && !isSharedEmpresa()
    };
  }

  function updateDeleteChecklist() {
    if (!state.empresa) return;
    var validation = getDeleteValidationState();
    var nameOk = validation.nameOk;
    var phraseOk = validation.phraseOk;
    var riskOk = validation.riskOk;
    var btn = $("empresaDeleteBtn");
    if (btn) {
      btn.disabled = state.deleting || !validation.ready;
    }
    setNodeState("empresaDeleteNameCheck", nameOk, nameOk ? "Nombre exacto validado" : "Nombre exacto pendiente");
    setNodeState("empresaDeletePhraseCheck", phraseOk, phraseOk ? "Frase ELIMINAR validada" : "Frase de seguridad pendiente");
    setNodeState("empresaDeleteRiskCheck", riskOk, riskOk ? "Riesgo aceptado conscientemente" : "Aceptacion de riesgo pendiente");
  }

  function setDeleteBusy(busy, text) {
    state.deleting = !!busy;
    var btn = $("empresaDeleteBtn");
    var progress = $("empresaDeleteProgress");
    var progressText = $("empresaDeleteProgressText");
    var controls = ["empresaDeleteConfirm", "empresaDeletePhrase", "empresaDeleteAcknowledge", "empresaDownloadBeforeDeleteBtn"];
    if (btn) {
      btn.disabled = !!busy || isSharedEmpresa();
      btn.textContent = busy ? "Eliminando empresa..." : "Eliminar empresa definitivamente";
    }
    if (progress) progress.hidden = !busy;
    if (progressText && text) progressText.textContent = text;
    controls.forEach(function (id) {
      var el = $(id);
      if (el) el.disabled = !!busy || isSharedEmpresa();
    });
    if (!busy) {
      updateDeleteChecklist();
    }
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

  function canManageShares() {
    return !!(state.shareMeta && (state.shareMeta.can_manage_shares || state.shareMeta.is_owner || state.shareMeta.is_super_admin));
  }

  function canRevokeAccess(item) {
    if (!item) return false;
    if (canManageShares()) return true;
    var requester = normalizeEmail(state.shareMeta && state.shareMeta.requester_email);
    return requester && (
      requester === normalizeEmail(item.admin_email) ||
      requester === normalizeEmail(item.compartido_por_email)
    );
  }

  function canRevokeInvitation(item) {
    if (!item) return false;
    if (canManageShares()) return true;
    var requester = normalizeEmail(state.shareMeta && state.shareMeta.requester_email);
    return requester && (
      requester === normalizeEmail(item.admin_email) ||
      requester === normalizeEmail(item.invitado_por_email)
    );
  }

  function displayPerson(name, email) {
    return String(name || email || "Administrador").trim();
  }

  function normalizeShareNivel(value) {
    value = String(value || "").trim().toLowerCase();
    if (value === "solo_ver" || value === "solo_lectura") return "solo_ver";
    if (value === "modulos" || value === "solo_modulos") return "modulos";
    return "acceso_total";
  }

  function shareNivelLabel(value) {
    switch (normalizeShareNivel(value)) {
      case "solo_ver": return "Solo ver";
      case "modulos": return "Solo ciertos modulos";
      default: return "Acceso total";
    }
  }

  function parseShareModules(value) {
    if (Array.isArray(value)) {
      return value.map(function (item) { return String(item || "").trim(); }).filter(Boolean);
    }
    return String(value || "").split(",").map(function (item) { return item.trim(); }).filter(Boolean);
  }

  function shareModuleLabel(modulo) {
    modulo = String(modulo || "").trim();
    var found = state.shareScopeCatalog.find(function (item) {
      return String(item && item.modulo || "") === modulo;
    });
    return found && found.label ? found.label : modulo;
  }

  function shareScopeText(item) {
    var nivel = normalizeShareNivel(item && item.nivel_acceso);
    if (item && item.scope_label) return String(item.scope_label);
    if (nivel !== "modulos") return shareNivelLabel(nivel);
    var modules = parseShareModules(item && item.modulos_permitidos);
    if (!modules.length) return "Solo ciertos modulos";
    return "Solo modulos: " + modules.map(shareModuleLabel).join(", ");
  }

  function renderShareModuleSelector() {
    var box = $("empresaShareModules");
    if (!box) return;
    var catalog = state.shareScopeCatalog.length ? state.shareScopeCatalog : [
      { modulo: "ventas", label: "Ventas" },
      { modulo: "inventario", label: "Inventario" },
      { modulo: "finanzas", label: "Finanzas, caja y reportes" },
      { modulo: "facturacion", label: "Facturacion electronica" },
      { modulo: "reportes", label: "Reportes" },
      { modulo: "documentos_onlyoffice", label: "Documentos" }
    ];
    box.innerHTML = catalog.map(function (item) {
      var modulo = String(item && item.modulo || "").trim();
      if (!modulo) return "";
      return '<label class="empresa-share-module-option">'
        + '<input type="checkbox" value="' + escapeHtml(modulo) + '" data-share-module>'
        + '<span>' + escapeHtml(item.label || modulo) + '</span>'
        + '</label>';
    }).join("");
  }

  function updateShareScopeVisibility() {
    var select = $("empresaShareNivel");
    var wrap = $("empresaShareModulesWrap");
    if (!select || !wrap) return;
    wrap.hidden = normalizeShareNivel(select.value) !== "modulos";
  }

  function getSelectedShareModules() {
    return Array.prototype.map.call(document.querySelectorAll("#empresaShareModules [data-share-module]:checked"), function (input) {
      return String(input.value || "").trim();
    }).filter(Boolean);
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
    if ($("empresaDeletePhrase")) {
      $("empresaDeletePhrase").disabled = !!isShared;
      $("empresaDeletePhrase").value = "";
    }
    if ($("empresaDeleteAcknowledge")) {
      $("empresaDeleteAcknowledge").disabled = !!isShared;
      $("empresaDeleteAcknowledge").checked = false;
    }
    if (saveButton) {
      saveButton.disabled = !!isShared;
    }
    if ($("empresaDeleteBtn")) {
      $("empresaDeleteBtn").disabled = true;
    }
    if ($("empresaDownloadBeforeDeleteBtn")) {
      $("empresaDownloadBeforeDeleteBtn").disabled = !!isShared;
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
    updateDeleteChecklist();
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
    var ownerCanManage = canManageShares();
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
            ? buildShareActionButton(isSharedEmpresa() && normalizeEmail(item.admin_email) === normalizeEmail(state.shareMeta.requester_email) ? 'Eliminar mi acceso' : 'Dejar de compartir', 'revoke', item.id, 'access')
            : '';
          return '<article class="empresa-share-item">'
            + '<div><strong>' + escapeHtml(sharedTo) + '</strong><div class="muted">' + escapeHtml(item.admin_email || '') + '</div>'
            + '<div class="muted">Compartido por: ' + escapeHtml(sharedBy) + (accepted ? ' - Desde: ' + escapeHtml(accepted) : '') + '</div>'
            + '<div class="muted">Alcance: ' + escapeHtml(shareScopeText(item)) + '</div></div>'
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
            actions += buildShareActionButton('Dejar de compartir', 'revoke', item.id, 'invitation');
          }
          return '<article class="empresa-share-item">'
            + '<div><strong>' + escapeHtml(invitedTo) + '</strong><div class="muted">' + escapeHtml(item.admin_email || '') + '</div>'
            + '<div class="muted">Invitado por: ' + escapeHtml(invitedBy) + (expira ? ' - Expira: ' + escapeHtml(expira) : '') + '</div>'
            + '<div class="muted">Alcance: ' + escapeHtml(shareScopeText(item)) + '</div></div>'
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
    state.shareScopeCatalog = Array.isArray(data && data.scope_catalog) ? data.scope_catalog : state.shareScopeCatalog;
    state.shareMeta = {
      is_owner: !!(data && data.is_owner),
      is_super_admin: !!(data && data.is_super_admin),
      can_manage_shares: !!(data && data.can_manage_shares),
      requester_email: String(data && data.requester_email ? data.requester_email : '').trim(),
      principal_email: String(data && data.principal_email ? data.principal_email : '').trim(),
    };
    renderShareModuleSelector();
    updateShareScopeVisibility();
    renderShares();
  }

  function formatCurrency(value) {
    var amount = Number(value || 0);
    return new Intl.NumberFormat("es-CO", {
      style: "currency",
      currency: "COP",
      maximumFractionDigits: 0,
    }).format(amount);
  }

  function formatDate(value) {
    var raw = String(value || "").trim();
    if (!raw) return "Sin fecha";
    var normalized = raw.length === 10 ? (raw + "T00:00:00") : raw.replace(" ", "T");
    var dt = new Date(normalized);
    if (Number.isNaN(dt.getTime())) return raw;
    return new Intl.DateTimeFormat("es-CO", { dateStyle: "medium" }).format(dt);
  }

  function buildLicenciaCheckoutUrl(mode, addonLicenciaIds) {
    var empresaId = state.empresa && state.empresa.id ? state.empresa.id : getEmpresaId();
    var licencias = state.licencias || {};
    var base = licencias.base_licencia || null;
    if (!empresaId || !base || !base.id) return "";
    var url = new URL("/pagar_licencia.html", window.location.origin);
    url.searchParams.set("empresa_id", String(empresaId));
    url.searchParams.set("licencia_id", String(base.id));
    if (state.empresa && state.empresa.tipo_id) url.searchParams.set("tipo_id", String(state.empresa.tipo_id));
    if (state.empresa && state.empresa.tipo_nombre) url.searchParams.set("tipo_nombre", String(state.empresa.tipo_nombre));
    if (mode) url.searchParams.set("checkout_mode", String(mode));
    if (addonLicenciaIds && addonLicenciaIds.length) {
      url.searchParams.set("addon_licencia_ids", addonLicenciaIds.join(","));
    }
    return url.pathname + url.search;
  }

  function renderLicencias() {
    var licencias = state.licencias || {};
    var resumenNode = $("empresaLicenciasResumen");
    var baseNode = $("empresaLicenciasBase");
    var activasNode = $("empresaLicenciasActivas");
    var catalogoNode = $("empresaLicenciasCatalogo");
    if (!resumenNode || !baseNode || !activasNode || !catalogoNode) return;

    var base = licencias.base_licencia || null;
    var addons = Array.isArray(licencias.licencias_adicionales) ? licencias.licencias_adicionales : [];
    var catalogo = Array.isArray(licencias.catalogo_adicionales) ? licencias.catalogo_adicionales : [];
    var bundle = licencias.bundle_summary || null;

    if (!base) {
      resumenNode.textContent = "La empresa todavía no tiene una licencia base activa. Primero debe existir una licencia principal antes de agregar módulos adicionales.";
      baseNode.innerHTML = '<p class="muted">Sin licencia base activa.</p>';
      activasNode.innerHTML = '<p class="muted">No se pueden activar adicionales sin una licencia base activa.</p>';
      catalogoNode.innerHTML = '<p class="muted">El catálogo adicional se habilita cuando la empresa tenga una licencia base vigente.</p>';
      return;
    }

    var siguienteTotal = bundle && typeof bundle.total_periodico_siguiente === "number" ? bundle.total_periodico_siguiente : Number(base.valor || 0);
    var checkoutAgrupado = buildLicenciaCheckoutUrl("empresa_bundle", []);
    resumenNode.innerHTML = ''
      + '<strong>Resumen agrupado:</strong> '
      + 'la siguiente renovación sumará la licencia base y los adicionales activos en un solo cobro de '
      + '<strong>' + escapeHtml(formatCurrency(siguienteTotal)) + '</strong>'
      + (bundle && bundle.fecha_corte_base ? ' con corte base el ' + escapeHtml(formatDate(bundle.fecha_corte_base)) + '.' : '.')
      + (checkoutAgrupado ? ' <a href="' + escapeHtml(checkoutAgrupado) + '">Renovar todo en un solo pago</a>.' : '');

    baseNode.innerHTML = ''
      + '<article class="empresa-share-item">'
      + '<div><strong>' + escapeHtml(base.nombre || "Licencia base") + '</strong>'
      + '<div class="muted">' + escapeHtml(base.descripcion || "Licencia principal de la empresa.") + '</div>'
      + '<div class="muted">Valor periódico: ' + escapeHtml(formatCurrency(base.valor || 0)) + ' - vence: ' + escapeHtml(formatDate(base.fecha_fin)) + '</div>'
      + '</div>'
      + '<div class="empresa-share-item-actions"><span class="empresa-share-state is-activo">Base</span></div>'
      + '</article>';

    if (!addons.length) {
      activasNode.innerHTML = '<p class="muted">No hay licencias adicionales activas o históricas para esta empresa.</p>';
    } else {
      activasNode.innerHTML = addons.map(function(item) {
        var activo = Number(item.activo || 0) === 1;
        var autoRenovar = Number(item.auto_renovar || 0) === 1;
        var status = activo ? "activo" : "inactivo";
        return '<article class="empresa-share-item">'
          + '<div><strong>' + escapeHtml(item.licencia_nombre || item.nombre || "Licencia adicional") + '</strong>'
          + '<div class="muted">' + escapeHtml(item.codigo_funcion || item.licencia_codigo_funcion || "Módulo adicional") + '</div>'
          + '<div class="muted">Valor periódico: ' + escapeHtml(formatCurrency(item.valor || 0)) + ' - vigente hasta: ' + escapeHtml(formatDate(item.fecha_fin)) + '</div>'
          + '</div>'
          + '<div class="empresa-share-item-actions">'
          + '<span class="empresa-share-state is-' + escapeHtml(status) + '">' + escapeHtml(activo ? "Activo" : "Inactivo") + '</span>'
          + '<button type="button" class="btn secondary empresa-addon-toggle-renew" data-licencia-id="' + escapeHtml(item.licencia_id) + '" data-auto-renovar="' + escapeHtml(autoRenovar ? "0" : "1") + '">' + (autoRenovar ? "No renovar" : "Reactivar renovación") + '</button>'
          + '<button type="button" class="btn danger empresa-addon-toggle-active" data-licencia-id="' + escapeHtml(item.licencia_id) + '" data-action="' + (activo ? "desactivar_adicional" : "activar_adicional") + '">' + (activo ? "Desactivar" : "Activar") + '</button>'
          + '</div>'
          + '</article>';
      }).join("");
    }

    if (!catalogo.length) {
      catalogoNode.innerHTML = '<p class="muted">No hay licencias adicionales disponibles para este tipo de empresa en este momento.</p>';
    } else {
      catalogoNode.innerHTML = catalogo.map(function(item) {
        var buyUrl = buildLicenciaCheckoutUrl("empresa_addons", [Number(item.id)]);
        return '<article class="empresa-share-item">'
          + '<div><strong>' + escapeHtml(item.nombre || "Licencia adicional") + '</strong>'
          + '<div class="muted">' + escapeHtml(item.descripcion || "Extiende módulos o funciones específicas sobre la licencia base.") + '</div>'
          + '<div class="muted">Código: ' + escapeHtml(item.codigo_funcion || "Sin código") + ' - valor: ' + escapeHtml(formatCurrency(item.valor || 0)) + ' - vigencia: ' + escapeHtml(String(item.duracion_dias || 0)) + ' días</div>'
          + '</div>'
          + '<div class="empresa-share-item-actions">'
          + (buyUrl ? '<a class="btn" href="' + escapeHtml(buyUrl) + '">Comprar adicional</a>' : '<span class="muted">No disponible</span>')
          + '</div>'
          + '</article>';
      }).join("");
    }
    updateDeleteChecklist();
  }

  async function loadLicencias() {
    if (!state.empresa || !state.empresa.id) return;
    var data = await fetchJSON('/super/api/empresa_licencias_adicionales?empresa_id=' + encodeURIComponent(state.empresa.id) + '&action=resumen', { credentials: 'same-origin' });
    state.licencias = data || null;
    renderLicencias();
  }

  async function toggleLicenciaAdicional(action, licenciaId, payload) {
    if (!state.empresa || !licenciaId) return;
    await fetchJSON('/super/api/empresa_licencias_adicionales?empresa_id=' + encodeURIComponent(state.empresa.id) + '&action=' + encodeURIComponent(action), {
      method: 'POST',
      credentials: 'same-origin',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(Object.assign({ licencia_id: Number(licenciaId) }, payload || {}))
    });
    await loadLicencias();
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
      await loadLicencias();
    } catch (err) {
      setMessage("empresaLicenciasMessage", err.message || "No se pudo cargar el estado de licencias de la empresa.", true);
    }
    try {
      if ($("empresaShareNivel")) $("empresaShareNivel").value = "solo_ver";
      Array.prototype.forEach.call(document.querySelectorAll("#empresaShareModules [data-share-module]"), function (input) {
        input.checked = false;
      });
      updateShareScopeVisibility();
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
    if (!state.empresa || state.deleting) return;
    var confirmacion = $("empresaDeleteConfirm").value.trim();
    var phrase = $("empresaDeletePhrase") ? $("empresaDeletePhrase").value.trim().toUpperCase() : "";
    var riskAccepted = !!($("empresaDeleteAcknowledge") && $("empresaDeleteAcknowledge").checked);
    updateDeleteChecklist();
    if (!confirmacion) {
      setMessage("empresaDeleteMessage", "Debes escribir el nombre de la empresa.", true);
      return;
    }
    if (confirmacion !== (state.empresa.nombre || "")) {
      setMessage("empresaDeleteMessage", "El nombre digitado no coincide exactamente.", true);
      return;
    }
    if (phrase !== "ELIMINAR") {
      setMessage("empresaDeleteMessage", "Debes escribir ELIMINAR para confirmar el borrado irreversible.", true);
      return;
    }
    if (!riskAccepted) {
      setMessage("empresaDeleteMessage", "Marca la aceptacion de riesgo antes de eliminar la empresa.", true);
      return;
    }
    if (!confirmDownloadBeforeEmpresaDelete()) {
      return;
    }
    try {
      setDeleteBusy(true, "Actualizando impacto operativo antes de eliminar...");
      var impactoData = await fetchJSON("/super/api/empresas?id=" + encodeURIComponent(state.empresa.id) + "&action=impacto_desactivacion", { credentials: "same-origin" });
      state.impacto = impactoData && impactoData.impacto ? impactoData.impacto : state.impacto;
      renderSummary();
      setDeleteBusy(true, "Eliminando registros, accesos y archivos asociados...");
      var data = await fetchJSON(
        "/super/api/empresas?id=" + encodeURIComponent(state.empresa.id) + "&action=eliminar_total",
        {
          method: "DELETE",
          credentials: "same-origin",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            confirmacion_nombre: confirmacion,
            confirmacion_accion: "ELIMINAR",
            confirmacion_riesgo: true,
            descarga_ofrecida: !!state.deleteDownloadOffered
          }),
        }
      );
      var result = data && data.result ? data.result : null;
      var archivos = data && data.archivos ? data.archivos : {};
      var pathsEliminados = Array.isArray(archivos.paths_eliminados) ? archivos.paths_eliminados.length : 0;
      var erroresArchivos = Array.isArray(archivos.errores) ? archivos.errores.length : 0;
      var message = result
        ? "Empresa eliminada. Tablas afectadas: " + result.tablas_afectadas + ". Registros eliminados: " + result.registros_eliminados + ". Carpetas eliminadas: " + pathsEliminados + "."
        : "Empresa eliminada correctamente. Carpetas eliminadas: " + pathsEliminados + ".";
      if (erroresArchivos) {
        message += " Advertencias al limpiar archivos: " + erroresArchivos + ".";
      }
      removeEmpresaFromSelectorOrderLocal(state.empresa.id);
      setMessage("empresaDeleteMessage", message, false);
      setDeleteBusy(true, "Empresa eliminada. Redirigiendo al selector...");
      window.setTimeout(function () {
        redirectToSeleccionarEmpresa();
      }, 900);
    } catch (err) {
      setDeleteBusy(false);
      setMessage("empresaDeleteMessage", err.message || "No se pudo eliminar la empresa.", true);
    }
  }

  async function inviteAdmin(ev) {
    ev.preventDefault();
    if (!state.empresa) return;
    if (!canManageShares()) {
      setMessage("empresaShareMessageBox", "Solo el propietario o un super administrador puede invitar nuevos administradores.", true);
      return;
    }

    var email = $("empresaShareEmail").value.trim();
    var mensaje = $("empresaShareMessage").value.trim();
    var nivelAcceso = normalizeShareNivel($("empresaShareNivel") ? $("empresaShareNivel").value : "solo_ver");
    var modulosPermitidos = nivelAcceso === "modulos" ? getSelectedShareModules() : [];
    if (!email) {
      setMessage("empresaShareMessageBox", "Debes indicar el correo del administrador.", true);
      return;
    }
    if (nivelAcceso === "modulos" && !modulosPermitidos.length) {
      setMessage("empresaShareMessageBox", "Elige al menos un modulo para compartir solo ciertos modulos.", true);
      return;
    }

    try {
      clearMessageActions();
      var data = await fetchJSON('/super/api/empresas/compartidos', {
        method: 'POST',
        credentials: 'same-origin',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          empresa_id: state.empresa.id,
          email: email,
          mensaje: mensaje,
          nivel_acceso: nivelAcceso,
          modulos_permitidos: modulosPermitidos
        })
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
    var downloadBtn = $("empresaDownloadBeforeDeleteBtn");
    if (downloadBtn) {
      downloadBtn.addEventListener("click", openEmpresaDownload);
    }
    ["empresaDeleteConfirm", "empresaDeletePhrase", "empresaDeleteAcknowledge"].forEach(function (id) {
      var el = $(id);
      if (!el) return;
      el.addEventListener(el.type === "checkbox" ? "change" : "input", updateDeleteChecklist);
    });
    var shareForm = $("empresaShareForm");
    if (shareForm) {
      shareForm.addEventListener("submit", inviteAdmin);
    }
    renderShareModuleSelector();
    updateShareScopeVisibility();
    var shareNivel = $("empresaShareNivel");
    if (shareNivel) {
      shareNivel.addEventListener("change", updateShareScopeVisibility);
    }
    document.addEventListener("click", function (ev) {
      var renewBtn = ev.target && ev.target.closest ? ev.target.closest(".empresa-addon-toggle-renew") : null;
      if (renewBtn) {
        toggleLicenciaAdicional("auto_renovar", renewBtn.getAttribute("data-licencia-id"), {
          auto_renovar: renewBtn.getAttribute("data-auto-renovar") === "1"
        }).then(function () {
          setMessage("empresaLicenciasMessage", "La preferencia de renovación quedó actualizada.", false);
        }).catch(function (err) {
          setMessage("empresaLicenciasMessage", err.message || "No se pudo actualizar la renovación del adicional.", true);
        });
        return;
      }
      var activeBtn = ev.target && ev.target.closest ? ev.target.closest(".empresa-addon-toggle-active") : null;
      if (activeBtn) {
        var action = activeBtn.getAttribute("data-action") || "desactivar_adicional";
        toggleLicenciaAdicional(action, activeBtn.getAttribute("data-licencia-id")).then(function () {
          setMessage("empresaLicenciasMessage", action === "desactivar_adicional" ? "La licencia adicional quedó desactivada." : "La licencia adicional quedó activa otra vez.", false);
        }).catch(function (err) {
          setMessage("empresaLicenciasMessage", err.message || "No se pudo actualizar la licencia adicional.", true);
        });
      }
    });
    loadEmpresa().catch(function (err) {
      setMessage("empresaEditMessage", err.message || "No se pudo cargar la empresa.", true);
      setMessage("empresaDeleteMessage", err.message || "No se pudo cargar la empresa.", true);
      setMessage("empresaShareMessageBox", err.message || "No se pudo cargar el estado de empresas compartidas.", true);
      setMessage("empresaLicenciasMessage", err.message || "No se pudo cargar el estado de licencias.", true);
    });
  });
})();
