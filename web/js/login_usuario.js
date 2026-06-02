(function () {
  "use strict";

  var state = {
    empresaID: 0,
    activeForm: "login",
    contract: null,
    invitationToken: ""
  };

  var forms = {
    login: document.getElementById("loginUsuarioForm"),
    setup: document.getElementById("setupPasswordForm"),
    recovery: document.getElementById("recoveryRequestForm"),
    inviteRecovery: document.getElementById("inviteRecoveryForm"),
    reset: document.getElementById("resetPasswordForm"),
    change: document.getElementById("changePasswordForm")
  };

  var messageTargets = {
    login: document.getElementById("msg"),
    setup: document.getElementById("setupMsg"),
    recovery: document.getElementById("recoveryMsg"),
    inviteRecovery: document.getElementById("inviteRecoveryMsg"),
    reset: document.getElementById("resetMsg"),
    change: document.getElementById("changePasswordMsg")
  };

  var empresaInput = document.getElementById("empresaID");
  var empresaContextHint = document.getElementById("empresaContextHint");
  var contractPanel = document.getElementById("contractPanel");
  var contractCheckbox = document.getElementById("contractAcceptCheckbox");
  var contractTitle = document.getElementById("contractTitle");
  var contractVersion = document.getElementById("contractVersion");
  var contractSummary = document.getElementById("contractSummary");
  var contractNote = document.getElementById("contractNote");
  var contractStatus = document.getElementById("contractStatus");
  var contractLink = document.getElementById("contractLink");
  var contractDialog = document.getElementById("contractDialog");
  var contractDialogTitle = document.getElementById("contractDialogTitle");
  var contractDialogSummary = document.getElementById("contractDialogSummary");
  var contractDialogContent = document.getElementById("contractDialogContent");
  var contractDialogClose = document.getElementById("contractDialogClose");
  var googleUsuarioBtn = document.getElementById("googleUsuarioBtn");
  var cashRegisterLoginDialog = document.getElementById("cashRegisterLoginDialog");
  var cashRegisterLoginSelect = document.getElementById("cashRegisterLoginSelect");
  var cashRegisterLoginMsg = document.getElementById("cashRegisterLoginMsg");
  var cashRegisterLoginContinue = document.getElementById("cashRegisterLoginContinue");
  var pendingCashRegisterAuthPayload = null;
  var recaptchaManagers = {
    login: window.PCSRecaptcha ? window.PCSRecaptcha.createManager({ containerId: "empresaLoginRecaptcha", action: "empresa_login" }) : null,
    setup: window.PCSRecaptcha ? window.PCSRecaptcha.createManager({ containerId: "empresaSetupRecaptcha", action: "empresa_setup_password" }) : null,
    recovery: window.PCSRecaptcha ? window.PCSRecaptcha.createManager({ containerId: "empresaRecoveryRecaptcha", action: "empresa_recovery" }) : null,
    inviteRecovery: window.PCSRecaptcha ? window.PCSRecaptcha.createManager({ containerId: "empresaInviteRecoveryRecaptcha", action: "empresa_invitation_recovery" }) : null,
    reset: window.PCSRecaptcha ? window.PCSRecaptcha.createManager({ containerId: "empresaResetRecaptcha", action: "empresa_reset_password" }) : null
  };

  function parsePositiveInt(raw) {
    var value = Number(String(raw || "").trim());
    if (!Number.isFinite(value)) return 0;
    value = Math.trunc(value);
    return value > 0 ? value : 0;
  }

  function getQueryParam(name) {
    try {
      return new URLSearchParams(window.location.search || "").get(name) || "";
    } catch (error) {
      return "";
    }
  }

  function persistEmpresaID(rawEmpresaID) {
    var parsed = parsePositiveInt(rawEmpresaID);
    if (!parsed) return 0;
    var value = String(parsed);
    try {
      window.sessionStorage.setItem("active_empresa_id", value);
      window.sessionStorage.setItem("empresa_id", value);
      window.sessionStorage.setItem("admin_empresa_id", value);
    } catch (error) {}
    try {
      window.localStorage.setItem("active_empresa_id", value);
      window.localStorage.setItem("empresa_id", value);
      window.localStorage.setItem("admin_empresa_id", value);
    } catch (error) {}
    return parsed;
  }

  function resolveEmpresaID() {
    var fromURL = parsePositiveInt(getQueryParam("empresa_id") || getQueryParam("id"));
    if (fromURL > 0) {
      return persistEmpresaID(fromURL);
    }
    return 0;
  }

  function updateEmpresaContextHint() {
    if (empresaInput && state.empresaID > 0) {
      empresaInput.value = String(state.empresaID);
    }
    if (!empresaContextHint) return;
    if (state.empresaID > 0) {
      empresaContextHint.textContent = "Referencia de empresa detectada desde un enlace anterior. Tambien puedes entrar solo con tu correo y clave.";
    } else {
      empresaContextHint.textContent = "El sistema abrirá automáticamente la empresa asociada a tu correo y cargará tus roles asignados.";
    }
  }

  function syncEmpresaIDFromInput() {
    var parsed = parsePositiveInt(empresaInput && empresaInput.value);
    if (!parsed) {
      updateEmpresaContextHint();
      updateGoogleUsuarioHref();
      return;
    }
    state.empresaID = persistEmpresaID(parsed);
    clearMessages();
    updateEmpresaContextHint();
    updateGoogleUsuarioHref();
  }

  function clearMessages() {
    Object.keys(messageTargets).forEach(function (key) {
      setMessage(key, "", false);
    });
  }

  function setMessage(key, text, isError) {
    var target = messageTargets[key];
    if (!target) return;
    target.textContent = text || "";
    target.classList.toggle("is-hidden", !text);
    target.classList.toggle("is-visible", !!text);
    target.classList.toggle("error", !!text && !!isError);
    target.classList.toggle("success", !!text && !isError);
  }

  function setPasswordVisibility(toggleBtn, input, isVisible) {
    if (!toggleBtn || !input) {
      return;
    }
    input.type = isVisible ? "text" : "password";
    toggleBtn.setAttribute("aria-pressed", isVisible ? "true" : "false");
    toggleBtn.setAttribute("aria-label", isVisible ? "Ocultar contrasena" : "Mostrar contrasena");
    toggleBtn.setAttribute("title", isVisible ? "Ocultar contrasena" : "Mostrar contrasena");
    toggleBtn.classList.toggle("is-visible", !!isVisible);
  }

  function initPasswordVisibilityToggles() {
    var toggles = document.querySelectorAll(".password-visibility-toggle[data-target]");
    Array.prototype.forEach.call(toggles, function (toggleBtn) {
      var targetId = toggleBtn.getAttribute("data-target");
      var input = targetId ? document.getElementById(targetId) : null;
      setPasswordVisibility(toggleBtn, input, false);
      toggleBtn.addEventListener("click", function () {
        if (!input) {
          return;
        }
        setPasswordVisibility(toggleBtn, input, input.type === "password");
        input.focus();
      });
    });
  }

  function setContractStatus(text, isWarning) {
    if (!contractStatus) return;
    contractStatus.textContent = text || "";
    contractStatus.classList.toggle("text-warning", !!text && !!isWarning);
  }

  function setContractPanelAttention(required) {
    if (!contractPanel) return;
    contractPanel.classList.toggle("requires-attention", !!required);
  }

  function getInvitationTokenFromURL() {
    return String(
      getQueryParam("token_invitacion") ||
      getQueryParam("invitation_token") ||
      getQueryParam("token_confirmacion") ||
      ""
    ).trim();
  }

  function setSetupInvitationToken(token) {
    state.invitationToken = String(token || "").trim();
    var input = document.getElementById("setupInvitationToken");
    if (input) input.value = state.invitationToken;
    updateGoogleUsuarioHref();
  }

  function invitationRequiredMessage() {
    return "El registro solo se completa desde la invitacion enviada por el administrador. Si borraste ese correo, usa Recuperar email de invitacion.";
  }

  function showInvitationRequiredOnLogin() {
    showForm("login", { email: document.getElementById("email").value || getQueryParam("email") });
    setMessage("login", invitationRequiredMessage(), true);
  }

  function googleErrorMessage(code) {
    switch (String(code || "").trim()) {
      case "sin_invitacion":
        return "Este correo de Google no tiene una invitacion activa de una empresa. Pide al administrador que cree tu usuario y te envie la invitacion.";
      case "invitacion_pendiente":
        return "Tu usuario existe, pero debes abrir la invitacion enviada por el administrador para completar el primer acceso.";
      case "correo_ambiguo":
        return "Este correo esta asociado a mas de una empresa. Abre el enlace de invitacion de la empresa correcta para entrar con Google.";
      case "contrato_requerido":
        return "Para entrar con Google desde la invitacion primero debes aceptar el contrato vigente en esta pantalla.";
      case "usuario_inactivo":
        return "Tu usuario esta inactivo. Solicita al administrador que revise tu acceso.";
      case "email_no_verificado":
        return "Google no confirmo que ese correo este verificado. Usa una cuenta de Google verificada.";
      case "sesion_error":
        return "Google valido tu correo, pero no fue posible abrir la sesion. Intenta nuevamente.";
      default:
        return "No se pudo iniciar con Google. Verifica que hayas recibido una invitacion de la empresa.";
    }
  }

  function updateGoogleUsuarioHref() {
    if (!googleUsuarioBtn) return;
    try {
      var target = new URL("/auth/google/usuario/login", window.location.origin);
      if (state.empresaID > 0) {
        target.searchParams.set("empresa_id", String(state.empresaID));
      }
      if (state.invitationToken) {
        target.searchParams.set("token_invitacion", state.invitationToken);
      }
      if (contractCheckbox && contractCheckbox.checked) {
        target.searchParams.set("accept_contract", "1");
      }
      googleUsuarioBtn.href = target.pathname + target.search;
    } catch (error) {
      googleUsuarioBtn.href = "/auth/google/usuario/login";
    }
  }

  function formatContractContent(content) {
    if (!content) {
      return "<p>No hay contenido contractual disponible en este momento.</p>";
    }
    var normalized = String(content).trim();
    if (!normalized) {
      return "<p>No hay contenido contractual disponible en este momento.</p>";
    }
    if (/<\/?[a-z][\s\S]*>/i.test(normalized)) {
      return normalized;
    }
    return normalized
      .split(/\n{2,}/)
      .map(function (block) {
        var safeText = block
          .split(/\n/)
          .map(function (line) { return line.trim(); })
          .filter(Boolean)
          .join("<br>");
        return safeText ? "<p>" + safeText + "</p>" : "";
      })
      .join("");
  }

  function closeContractDialog() {
    if (!contractDialog) return;
    contractDialog.hidden = true;
    contractDialog.setAttribute("aria-hidden", "true");
  }

  function openContractDialog() {
    if (!contractDialog) return;
    if (state.contract) {
      contractDialogTitle.textContent = state.contract.titulo || "Contrato vigente";
      contractDialogSummary.textContent = state.contract.resumen || "Lee el contrato vigente antes de continuar.";
      contractDialogContent.innerHTML = formatContractContent(state.contract.contenido);
    } else {
      contractDialogTitle.textContent = "Contrato vigente";
      contractDialogSummary.textContent = "No fue posible cargar el contenido del contrato.";
      contractDialogContent.innerHTML = "<p>Recarga esta página o intenta de nuevo en unos segundos.</p>";
    }
    contractDialog.hidden = false;
    contractDialog.setAttribute("aria-hidden", "false");
  }

  function applyContract(contract) {
    state.contract = contract || null;
    if (!contract) {
      contractTitle.textContent = "Contrato no disponible";
      contractVersion.textContent = "Versión --";
      contractSummary.textContent = "No fue posible cargar el contrato vigente. Intenta recargar esta página.";
      contractNote.textContent = "";
      if (contractDialogTitle) contractDialogTitle.textContent = "Contrato no disponible";
      if (contractDialogSummary) contractDialogSummary.textContent = "No fue posible cargar el contenido contractual.";
      if (contractDialogContent) contractDialogContent.innerHTML = "<p>Intenta recargar esta página para volver a consultar el contrato vigente.</p>";
      setContractStatus("Si el problema persiste, informa al administrador.", true);
      return;
    }

    contractTitle.textContent = contract.titulo || "Contrato vigente";
    contractVersion.textContent = "Versión " + String(contract.version || "--");
    contractSummary.textContent = contract.resumen || "Lee el contrato antes de continuar con el acceso.";
    contractNote.textContent = contract.nota_aceptacion || "";
    if (contractLink) {
      contractLink.href = "#contrato-interno";
    }
    if (contractDialogTitle) contractDialogTitle.textContent = contract.titulo || "Contrato vigente";
    if (contractDialogSummary) contractDialogSummary.textContent = contract.resumen || "Lee el contrato vigente antes de continuar.";
    if (contractDialogContent) contractDialogContent.innerHTML = formatContractContent(contract.contenido);
    setContractStatus("La aceptación se registrará cuando completes tu registro.", false);
  }

  function loadContract() {
    return fetch("/api/public/contrato", { credentials: "same-origin" })
      .then(function (response) {
        return response.json();
      })
      .then(function (payload) {
        applyContract(payload && payload.contrato ? payload.contrato : null);
      })
      .catch(function () {
        applyContract(null);
      });
  }

  function getActiveFormKey() {
    return state.activeForm || "login";
  }

  function showForm(name, options) {
    var selected = name || "login";
    Object.keys(forms).forEach(function (key) {
      if (!forms[key]) return;
      forms[key].classList.toggle("is-hidden", key !== selected);
    });
    state.activeForm = selected;
    clearMessages();

    options = options || {};
    if (options.email) {
      prefillEmail(options.email);
    }
    if (options.invitationToken) {
      setSetupInvitationToken(options.invitationToken);
    }
    if (options.token) {
      var resetToken = document.getElementById("resetToken");
      if (resetToken) resetToken.value = options.token;
    }
    if (options.message) {
      setMessage(selected, options.message, false);
    }
    if (recaptchaManagers[selected]) {
      recaptchaManagers[selected].init().catch(function (error) {
        setMessage(selected, error && error.message ? error.message : 'No se pudo cargar la verificación de seguridad.', true);
      });
    }
  }

  function ensureRecaptcha(formKey) {
    var manager = recaptchaManagers[formKey];
    if (!manager) {
      return Promise.resolve('');
    }
    return manager.ensureToken().then(function (result) {
      if (!result.ok) {
        setMessage(formKey, result.message || 'Completa el reCAPTCHA que aparece debajo del formulario para continuar.', true);
        return null;
      }
      return result.token || '';
    }).catch(function (error) {
      setMessage(formKey, error && error.message ? error.message : 'No se pudo cargar la verificación de seguridad.', true);
      return null;
    });
  }

  function prefillEmail(email) {
    var normalized = String(email || "").trim();
    if (!normalized) return;
    ["email", "setupEmail", "recoveryEmail", "inviteRecoveryEmail", "resetEmail", "changeEmail"].forEach(function (id) {
      var input = document.getElementById(id);
      if (input) input.value = normalized;
    });
  }

  function withBasePayload(payload) {
    var out = payload || {};
    if (state.empresaID > 0) {
      out.empresa_id = state.empresaID;
    }
    return out;
  }

  function readResponsePayload(response) {
    var contentType = String(response.headers.get("Content-Type") || "").toLowerCase();
    if (contentType.indexOf("application/json") >= 0) {
      return response.json();
    }
    return response.text().then(function (text) {
      return { message: text || ("HTTP " + response.status) };
    });
  }

  function normalizeErrorMessage(payload, fallback) {
    if (!payload) return fallback;
    if (typeof payload === "string") return payload;
    if (payload.message) return String(payload.message);
    if (payload.error) return String(payload.error);
    return fallback;
  }

  function persistThemePreference(theme) {
    var normalized = String(theme || "").trim();
    if (!normalized) return;
    try {
      window.localStorage.setItem("theme", normalized);
    } catch (error) {}
    try {
      document.cookie = "pcs_theme=" + encodeURIComponent(normalized) + "; Path=/; Max-Age=31536000; SameSite=Lax";
    } catch (error) {}
  }

  function handleContractRequirement(payload, formKey) {
    setContractPanelAttention(true);
    if (contractCheckbox) {
      contractCheckbox.checked = false;
    }
    if (payload && payload.contract) {
      applyContract(payload.contract);
    }
    setContractStatus("Debes aceptar el contrato antes de completar tu registro.", true);
    setMessage(formKey, normalizeErrorMessage(payload, "Debes aceptar el contrato vigente para completar tu registro."), true);
    if (contractPanel && typeof contractPanel.scrollIntoView === "function") {
      contractPanel.scrollIntoView({ behavior: "smooth", block: "center" });
    }
  }

  function redirectAfterAuth(payload) {
    var redirectURL = String((payload && payload.redirect_url) || "").trim();
    if (!redirectURL && state.empresaID > 0) {
      redirectURL = "/administrar_empresa.html?id=" + encodeURIComponent(state.empresaID);
    }
    if (!redirectURL) {
      redirectURL = "/administrar_empresa.html";
    }
    window.location.href = redirectURL;
  }

  function normalizeCashCode(value) {
    return String(value == null ? "" : value)
      .trim()
      .toUpperCase()
      .replace(/\s+/g, "_")
      .replace(/[^A-Z0-9_-]/g, "_")
      .replace(/_+/g, "_")
      .replace(/^_+|_+$/g, "");
  }

  function normalizeCashierRole(value) {
    var raw = String(value || "").trim().toLowerCase();
    return raw.replace(/\s+/g, "_");
  }

  function isCashierAuthPayload(payload) {
    var role = normalizeCashierRole((payload && (payload.rol_nombre || payload.rol || payload.role)) || "");
    return role === "cajero" || role === "caja" || role === "cashier";
  }

  function parseStationConfigValue(raw) {
    var current = raw;
    for (var i = 0; i < 8; i += 1) {
      if (typeof current !== "string") break;
      var trimmed = current.trim();
      if (!trimmed) return null;
      try {
        current = JSON.parse(trimmed);
      } catch (error) {
        return null;
      }
    }
    return current && typeof current === "object" ? current : null;
  }

  function defaultCashRegisterRow(index) {
    var pos = Math.max(1, Number(index || 1));
    return {
      codigo: "CAJA-" + pos,
      nombre: pos === 1 ? "Caja principal" : ("Caja " + pos),
      descripcion: "",
      activa: true
    };
  }

  function normalizeCashRegisterRows(raw) {
    var rows = Array.isArray(raw) ? raw : [];
    var out = [];
    var seen = {};
    rows.forEach(function (item, index) {
      if (!item || typeof item !== "object") return;
      var fallback = defaultCashRegisterRow(index + 1);
      var codigo = normalizeCashCode(item.codigo || item.caja_codigo || item.code || fallback.codigo) || fallback.codigo;
      if (seen[codigo]) return;
      seen[codigo] = true;
      var nombre = String(item.nombre || item.caja_nombre || item.name || fallback.nombre).trim() || fallback.nombre;
      var descripcion = String(item.descripcion || item.caja_descripcion || item.description || "").trim();
      var activa = item.activa === undefined && item.active === undefined ? true : !!(item.activa !== undefined ? item.activa : item.active);
      if (activa !== false) {
        out.push({ codigo: codigo, nombre: nombre, descripcion: descripcion, activa: true });
      }
    });
    if (!out.length) out.push(defaultCashRegisterRow(1));
    return out.slice(0, 50);
  }

  function cashRegisterLoginStorageKey(payload) {
    var empresaID = Number((payload && payload.empresa_id) || state.empresaID || 0);
    var email = String((payload && payload.email) || document.getElementById("email").value || "").trim().toLowerCase();
    return "pcs_cajero_caja_login_" + String(empresaID || 0) + "_" + email.replace(/[^a-z0-9_.@-]/g, "_");
  }

  function readLastCashRegisterCode(payload) {
    try {
      return normalizeCashCode(window.localStorage.getItem(cashRegisterLoginStorageKey(payload)));
    } catch (error) {
      return "";
    }
  }

  function persistSelectedCashRegister(payload, caja) {
    if (!caja || !caja.codigo) return;
    try {
      window.localStorage.setItem(cashRegisterLoginStorageKey(payload), caja.codigo);
      window.sessionStorage.setItem("pcs_caja_trabajo_codigo", caja.codigo);
      window.sessionStorage.setItem("pcs_caja_trabajo_nombre", caja.nombre || "");
      window.sessionStorage.setItem("pcs_caja_trabajo_descripcion", caja.descripcion || "");
      window.sessionStorage.setItem("pcs_caja_trabajo_empresa_id", String((payload && payload.empresa_id) || state.empresaID || 0));
    } catch (error) {}
  }

  function redirectWithCashRegister(payload, caja) {
    persistSelectedCashRegister(payload, caja);
    var next = Object.assign({}, payload || {});
    var redirectURL = String(next.redirect_url || "").trim() || "/administrar_empresa.html";
    try {
      var target = new URL(redirectURL, window.location.origin);
      target.searchParams.set("caja_codigo", caja.codigo);
      target.searchParams.set("caja_nombre", caja.nombre || "");
      if (caja.descripcion) target.searchParams.set("caja_descripcion", caja.descripcion);
      next.redirect_url = target.pathname + target.search + target.hash;
    } catch (error) {}
    redirectAfterAuth(next);
  }

  function setCashRegisterLoginMessage(text, isError) {
    if (!cashRegisterLoginMsg) return;
    cashRegisterLoginMsg.textContent = text || "";
    cashRegisterLoginMsg.classList.toggle("value-negative", !!isError);
  }

  function closeCashRegisterLoginDialog() {
    if (!cashRegisterLoginDialog) return;
    cashRegisterLoginDialog.hidden = true;
    cashRegisterLoginDialog.setAttribute("aria-hidden", "true");
  }

  function openCashRegisterLoginDialog(payload, rows) {
    if (!cashRegisterLoginDialog || !cashRegisterLoginSelect) {
      redirectAfterAuth(payload);
      return;
    }
    pendingCashRegisterAuthPayload = payload || null;
    var lastCode = readLastCashRegisterCode(payload);
    cashRegisterLoginSelect.innerHTML = "";
    rows.forEach(function (caja) {
      var label = caja.codigo + (caja.nombre ? (" - " + caja.nombre) : "");
      if (caja.descripcion && label.toLowerCase().indexOf(caja.descripcion.toLowerCase()) === -1) {
        label += " · " + caja.descripcion;
      }
      cashRegisterLoginSelect.appendChild(new Option(label, caja.codigo));
    });
    if (lastCode && rows.some(function (item) { return item.codigo === lastCode; })) {
      cashRegisterLoginSelect.value = lastCode;
    }
    setCashRegisterLoginMessage(lastCode ? "Ultima caja usada preseleccionada." : "Elige la caja fisica que usaras en este turno.", false);
    cashRegisterLoginDialog.hidden = false;
    cashRegisterLoginDialog.setAttribute("aria-hidden", "false");
    cashRegisterLoginSelect.focus();
  }

  function loadCashRegisterLoginConfig(empresaID) {
    if (!empresaID) {
      return Promise.resolve({ enabled: true, rows: normalizeCashRegisterRows([]) });
    }
    return fetch("/api/empresa/estacion_prefs?empresa_id=" + encodeURIComponent(empresaID), { credentials: "same-origin" })
      .then(function (response) {
        if (!response.ok) throw new Error("No se pudo cargar la configuracion de cajas.");
        return response.json();
      })
      .then(function (data) {
        var prefs = data && Array.isArray(data.prefs) ? data.prefs : [];
        var pref = prefs.find(function (item) {
          return item && String(item.clave || "") === "estaciones_config" && Number(item.estacion_id || 0) === 0;
        });
        var cfg = parseStationConfigValue(pref && pref.valor) || {};
        return {
          enabled: cfg.solicitar_caja_login_cajero === undefined ? true : !!cfg.solicitar_caja_login_cajero,
          rows: normalizeCashRegisterRows(cfg.cajas_config || cfg.cajas_fisicas || cfg.cajas || cfg.caja_catalogo)
        };
      })
      .catch(function () {
        return { enabled: true, rows: normalizeCashRegisterRows([]) };
      });
  }

  function maybePromptCashRegisterBeforeRedirect(payload) {
    if (!isCashierAuthPayload(payload)) {
      redirectAfterAuth(payload);
      return;
    }
    loadCashRegisterLoginConfig(Number((payload && payload.empresa_id) || state.empresaID || 0)).then(function (config) {
      if (!config || config.enabled === false) {
        redirectAfterAuth(payload);
        return;
      }
      var rows = normalizeCashRegisterRows(config.rows);
      if (rows.length <= 1) {
        redirectWithCashRegister(payload, rows[0]);
        return;
      }
      openCashRegisterLoginDialog(payload, rows);
    });
  }

  function handleAuthResult(payload, formKey) {
    if (payload && payload.ok) {
      setContractPanelAttention(false);
      if (payload.empresa_id) {
        state.empresaID = persistEmpresaID(payload.empresa_id);
        updateEmpresaContextHint();
      }
      persistThemePreference(payload.apariencia);
      maybePromptCashRegisterBeforeRedirect(payload);
      return;
    }

    if (payload && payload.contract_acceptance_required) {
      handleContractRequirement(payload, formKey);
      return;
    }

    if (payload && payload.password_setup_required) {
      if (payload.empresa_id) {
        state.empresaID = persistEmpresaID(payload.empresa_id);
        updateEmpresaContextHint();
      }
      if (!state.invitationToken) {
        showForm("login", { email: payload.email });
        setMessage("login", payload.message || invitationRequiredMessage(), true);
        return;
      }
      showForm("setup", {
        email: payload.email,
        invitationToken: state.invitationToken,
        message: payload.message || "Completa tu contraseña inicial para continuar."
      });
      return;
    }

    if (payload && payload.password_rotation_required) {
      showForm("change", {
        email: payload.email,
        message: payload.message || "Debes cambiar tu contraseña antes de continuar."
      });
      return;
    }

    throw new Error(normalizeErrorMessage(payload, "No fue posible completar el acceso."));
  }

  function submitJSON(url, payload) {
    var controller = new AbortController();
    var tid = setTimeout(function () {
      controller.abort();
    }, 45000);
    return fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "same-origin",
      body: JSON.stringify(payload),
      signal: controller.signal
    }).then(function (response) {
      return readResponsePayload(response).then(function (body) {
        clearTimeout(tid);
        if (!response.ok) {
          throw new Error(normalizeErrorMessage(body, "HTTP " + response.status));
        }
        return body;
      });
    }).catch(function (err) {
      clearTimeout(tid);
      if (err && (err.name === "AbortError" || String(err.message || "").toLowerCase().indexOf("abort") >= 0)) {
        throw new Error("Tiempo de espera agotado al contactar el servidor.");
      }
      throw err;
    });
  }

  function onLoginSubmit(event) {
    event.preventDefault();

    var btn = document.getElementById("btnIngresar");
    var prevText = btn ? String(btn.textContent || "") : "";
    function setLoginBusy(isBusy) {
      if (!btn) return;
      btn.disabled = !!isBusy;
      btn.textContent = isBusy ? "Ingresando..." : prevText;
    }
    setMessage("login", "", false);
    setLoginBusy(true);

    ensureRecaptcha("login").then(function (token) {
      if (token === null) return;
      return submitJSON("/api/empresa/usuarios/login", withBasePayload({
        email: document.getElementById("email").value,
        password: document.getElementById("password").value,
        recaptcha_token: token
      }));
    })
      .then(function (payload) {
        if (!payload) return;
        handleAuthResult(payload, "login");
      })
      .catch(function (error) {
        setMessage("login", error.message || "No fue posible iniciar sesión.", true);
      })
      .finally(function () {
        setLoginBusy(false);
        if (recaptchaManagers.login) recaptchaManagers.login.reset();
      });
  }

  function onSetupSubmit(event) {
    event.preventDefault();

    var btn = document.getElementById("btnCrearPassword");
    var prevText = btn ? String(btn.textContent || "") : "";
    function setSetupBusy(isBusy) {
      if (!btn) return;
      btn.disabled = !!isBusy;
      btn.textContent = isBusy ? "Guardando..." : prevText;
    }
    setMessage("setup", "", false);
    var invitationToken = state.invitationToken || (document.getElementById("setupInvitationToken") && document.getElementById("setupInvitationToken").value) || "";
    invitationToken = String(invitationToken || "").trim();
    if (!invitationToken) {
      setMessage("setup", invitationRequiredMessage(), true);
      return;
    }
    setSetupBusy(true);

    ensureRecaptcha("setup").then(function (token) {
      if (token === null) return;
      return submitJSON("/api/empresa/usuarios/establecer_password", withBasePayload({
        email: document.getElementById("setupEmail").value,
        documento_identidad: document.getElementById("setupDocumento").value,
        password: document.getElementById("setupPassword").value,
        password_confirm: document.getElementById("setupPasswordConfirm").value,
        token_invitacion: invitationToken,
        accept_contract: !!(contractCheckbox && contractCheckbox.checked),
        recaptcha_token: token
      }));
    })
      .then(function (payload) {
        if (!payload) return;
        handleAuthResult(payload, "setup");
      })
      .catch(function (error) {
        setMessage("setup", error.message || "No fue posible crear la contraseña.", true);
      })
      .finally(function () {
        setSetupBusy(false);
        if (recaptchaManagers.setup) recaptchaManagers.setup.reset();
      });
  }

  function onRecoverySubmit(event) {
    event.preventDefault();

    var btn = document.getElementById("btnSolicitarRecuperacion");
    var prevText = btn ? String(btn.textContent || "") : "";
    function setRecoveryBusy(isBusy) {
      if (!btn) return;
      btn.disabled = !!isBusy;
      btn.textContent = isBusy ? "Enviando..." : prevText;
    }
    setMessage("recovery", "", false);
    setRecoveryBusy(true);

    ensureRecaptcha("recovery").then(function (token) {
      if (token === null) return;
      return submitJSON("/api/empresa/usuarios/solicitar_recuperacion_password", withBasePayload({
        email: document.getElementById("recoveryEmail").value,
        recaptcha_token: token
      }));
    })
      .then(function (payload) {
        if (!payload) return;
        prefillEmail(document.getElementById("recoveryEmail").value);
        setMessage("recovery", normalizeErrorMessage(payload, "Si el correo existe, enviaremos instrucciones para recuperar la contraseña."), false);
      })
      .catch(function (error) {
        setMessage("recovery", error.message || "No fue posible iniciar la recuperación.", true);
      })
      .finally(function () {
        setRecoveryBusy(false);
        if (recaptchaManagers.recovery) recaptchaManagers.recovery.reset();
      });
  }

  function onInviteRecoverySubmit(event) {
    event.preventDefault();

    var btn = document.getElementById("btnRecuperarInvitacion");
    var prevText = btn ? String(btn.textContent || "") : "";
    function setInviteRecoveryBusy(isBusy) {
      if (!btn) return;
      btn.disabled = !!isBusy;
      btn.textContent = isBusy ? "Enviando..." : prevText;
    }
    setMessage("inviteRecovery", "", false);
    setInviteRecoveryBusy(true);

    ensureRecaptcha("inviteRecovery").then(function (token) {
      if (token === null) return;
      return submitJSON("/api/empresa/usuarios/recuperar_invitacion", withBasePayload({
        email: document.getElementById("inviteRecoveryEmail").value,
        recaptcha_token: token
      }));
    })
      .then(function (payload) {
        if (!payload) return;
        prefillEmail(document.getElementById("inviteRecoveryEmail").value);
        setMessage("inviteRecovery", normalizeErrorMessage(payload, "Si ese correo tiene una invitacion pendiente, enviaremos nuevamente el email de invitacion."), false);
      })
      .catch(function (error) {
        setMessage("inviteRecovery", error.message || "No fue posible solicitar la invitacion.", true);
      })
      .finally(function () {
        setInviteRecoveryBusy(false);
        if (recaptchaManagers.inviteRecovery) recaptchaManagers.inviteRecovery.reset();
      });
  }

  function onResetSubmit(event) {
    event.preventDefault();

    var btn = document.getElementById("btnRestablecerPassword");
    var prevText = btn ? String(btn.textContent || "") : "";
    function setResetBusy(isBusy) {
      if (!btn) return;
      btn.disabled = !!isBusy;
      btn.textContent = isBusy ? "Restableciendo..." : prevText;
    }
    setMessage("reset", "", false);
    setResetBusy(true);

    ensureRecaptcha("reset").then(function (token) {
      if (token === null) return;
      return submitJSON("/api/empresa/usuarios/restablecer_password", withBasePayload({
        email: document.getElementById("resetEmail").value,
        token: document.getElementById("resetToken").value,
        password: document.getElementById("resetPassword").value,
        password_confirm: document.getElementById("resetPasswordConfirm").value,
        recaptcha_token: token
      }));
    })
      .then(function (payload) {
        if (!payload) return;
        handleAuthResult(payload, "reset");
      })
      .catch(function (error) {
        setMessage("reset", error.message || "No fue posible restablecer la contraseña.", true);
      })
      .finally(function () {
        setResetBusy(false);
        if (recaptchaManagers.reset) recaptchaManagers.reset.reset();
      });
  }

  function onChangePasswordSubmit(event) {
    event.preventDefault();

    var btn = document.getElementById("btnCambiarPassword");
    var prevText = btn ? String(btn.textContent || "") : "";
    function setChangeBusy(isBusy) {
      if (!btn) return;
      btn.disabled = !!isBusy;
      btn.textContent = isBusy ? "Actualizando..." : prevText;
    }
    setMessage("change", "", false);
    setChangeBusy(true);

    submitJSON("/api/empresa/usuarios/cambiar_password", withBasePayload({
      email: document.getElementById("changeEmail").value,
      current_password: document.getElementById("changeCurrentPassword").value,
      new_password: document.getElementById("changeNewPassword").value,
      new_password_confirm: document.getElementById("changeNewPasswordConfirm").value
    }))
      .then(function (payload) {
        handleAuthResult(payload, "change");
      })
      .catch(function (error) {
        setMessage("change", error.message || "No fue posible cambiar la contraseña.", true);
      })
      .finally(function () {
        setChangeBusy(false);
      });
  }

  function bindButton(id, handler) {
    var element = document.getElementById(id);
    if (!element) return;
    element.addEventListener("click", function (event) {
      event.preventDefault();
      handler();
    });
  }

  state.empresaID = resolveEmpresaID();
  updateEmpresaContextHint();
  loadContract();

  var urlEmail = String(getQueryParam("email") || "").trim();
  var urlToken = String(getQueryParam("token_recuperacion") || "").trim();
  var urlInvitationToken = getInvitationTokenFromURL();
  if (urlInvitationToken) {
    setSetupInvitationToken(urlInvitationToken);
  }
  if (urlEmail) {
    prefillEmail(urlEmail);
  }

  if (empresaInput) {
    empresaInput.addEventListener("change", syncEmpresaIDFromInput);
    empresaInput.addEventListener("blur", syncEmpresaIDFromInput);
  }

  if (contractCheckbox) {
    contractCheckbox.addEventListener("change", function () {
      if (contractCheckbox.checked) {
        setContractPanelAttention(false);
        setContractStatus("La aceptación se registrará cuando completes tu registro.", false);
      }
      updateGoogleUsuarioHref();
    });
  }

  if (googleUsuarioBtn) {
    googleUsuarioBtn.addEventListener("click", function (event) {
      updateGoogleUsuarioHref();
      if (state.invitationToken && contractCheckbox && !contractCheckbox.checked) {
        event.preventDefault();
        showForm("setup", {
          email: document.getElementById("email").value || urlEmail,
          invitationToken: state.invitationToken
        });
        setContractPanelAttention(true);
        setContractStatus("Acepta el contrato para continuar con Google.", true);
        setMessage("setup", googleErrorMessage("contrato_requerido"), true);
      }
    });
  }

  if (contractLink) {
    contractLink.addEventListener("click", function (event) {
      event.preventDefault();
      openContractDialog();
    });
  }

  if (contractDialogClose) {
    contractDialogClose.addEventListener("click", function (event) {
      event.preventDefault();
      closeContractDialog();
    });
  }

  if (contractDialog) {
    contractDialog.addEventListener("click", function (event) {
      if (event.target && event.target.hasAttribute("data-close-contract")) {
        closeContractDialog();
      }
    });
  }

  if (cashRegisterLoginContinue) {
    cashRegisterLoginContinue.addEventListener("click", function (event) {
      event.preventDefault();
      var payload = pendingCashRegisterAuthPayload || {};
      var code = normalizeCashCode(cashRegisterLoginSelect && cashRegisterLoginSelect.value);
      var selected = null;
      if (cashRegisterLoginSelect) {
        selected = Array.prototype.slice.call(cashRegisterLoginSelect.options || []).map(function (option) {
          return {
            codigo: normalizeCashCode(option.value),
            nombre: String(option.textContent || "").replace(normalizeCashCode(option.value), "").replace(/^[\s\-·]+/, "").split("·")[0].trim(),
            descripcion: String(option.textContent || "").split("·").slice(1).join("·").trim()
          };
        }).find(function (item) { return item.codigo === code; });
      }
      if (!selected || !selected.codigo) {
        setCashRegisterLoginMessage("Selecciona una caja para continuar.", true);
        return;
      }
      closeCashRegisterLoginDialog();
      redirectWithCashRegister(payload, selected);
    });
  }

  document.addEventListener("keydown", function (event) {
    if (event.key === "Escape" && contractDialog && !contractDialog.hidden) {
      closeContractDialog();
    }
  });

  if (forms.login) forms.login.addEventListener("submit", onLoginSubmit);
  if (forms.setup) forms.setup.addEventListener("submit", onSetupSubmit);
  if (forms.recovery) forms.recovery.addEventListener("submit", onRecoverySubmit);
  if (forms.inviteRecovery) forms.inviteRecovery.addEventListener("submit", onInviteRecoverySubmit);
  if (forms.reset) forms.reset.addEventListener("submit", onResetSubmit);
  if (forms.change) forms.change.addEventListener("submit", onChangePasswordSubmit);

  bindButton("btnMostrarRegistro", function () {
    if (!state.invitationToken) {
      showInvitationRequiredOnLogin();
      return;
    }
    showForm("setup", { email: document.getElementById("email").value || urlEmail, invitationToken: state.invitationToken });
  });
  bindButton("btnVolverALoginHero", function () {
    showForm("login");
  });
  bindButton("btnMostrarRecuperacionHero", function () {
    showForm("recovery", { email: document.getElementById("email").value || urlEmail });
  });
  bindButton("linkGoSetup", function () {
    if (!state.invitationToken) {
      showInvitationRequiredOnLogin();
      return;
    }
    showForm("setup", { email: document.getElementById("email").value || urlEmail, invitationToken: state.invitationToken });
  });
  bindButton("linkGoRecovery", function () {
    showForm("recovery", { email: document.getElementById("email").value || urlEmail });
  });
  bindButton("linkGoInviteRecovery", function () {
    showForm("inviteRecovery", { email: document.getElementById("email").value || urlEmail });
  });
  bindButton("linkGoChangePassword", function () {
    showForm("change", { email: document.getElementById("email").value || urlEmail });
  });
  bindButton("btnVolverLogin", function () {
    showForm("login", { email: document.getElementById("setupEmail").value });
  });
  bindButton("btnVolverDesdeRecuperacion", function () {
    showForm("login", { email: document.getElementById("recoveryEmail").value });
  });
  bindButton("btnVolverDesdeInvitacion", function () {
    showForm("login", { email: document.getElementById("inviteRecoveryEmail").value });
  });
  bindButton("btnIrAReset", function () {
    showForm("reset", { email: document.getElementById("recoveryEmail").value });
  });
  bindButton("btnVolverDesdeReset", function () {
    showForm("login", { email: document.getElementById("resetEmail").value });
  });
  bindButton("btnVolverDesdeCambio", function () {
    showForm("login", { email: document.getElementById("changeEmail").value });
  });

  if (urlInvitationToken) {
    showForm("setup", {
      email: urlEmail,
      invitationToken: urlInvitationToken,
      message: "Invitacion detectada. Crea tu contrasena para entrar al panel de la empresa."
    });
  } else if (urlToken) {
    showForm("reset", {
      email: urlEmail,
      token: urlToken,
      message: "Token detectado en la URL. Define tu nueva contraseña para continuar."
    });
  } else {
    showForm("login", { email: urlEmail });
  }
  initPasswordVisibilityToggles();
  updateGoogleUsuarioHref();

  var googleError = String(getQueryParam("google_error") || "").trim();
  if (googleError) {
    if (googleError === "contrato_requerido" && state.invitationToken) {
      showForm("setup", { email: urlEmail, invitationToken: state.invitationToken });
      setContractPanelAttention(true);
      setContractStatus("Acepta el contrato para continuar con Google.", true);
    }
    setMessage(getActiveFormKey(), googleErrorMessage(googleError), true);
  }
})();
