(function () {
  "use strict";

  var state = {
    empresaID: 0,
    activeForm: "login",
    contract: null
  };

  var forms = {
    login: document.getElementById("loginUsuarioForm"),
    setup: document.getElementById("setupPasswordForm"),
    recovery: document.getElementById("recoveryRequestForm"),
    reset: document.getElementById("resetPasswordForm"),
    change: document.getElementById("changePasswordForm")
  };

  var messageTargets = {
    login: document.getElementById("msg"),
    setup: document.getElementById("setupMsg"),
    recovery: document.getElementById("recoveryMsg"),
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
  var recaptchaManagers = {
    login: window.PCSRecaptcha ? window.PCSRecaptcha.createManager({ containerId: "empresaLoginRecaptcha", action: "empresa_login" }) : null,
    setup: window.PCSRecaptcha ? window.PCSRecaptcha.createManager({ containerId: "empresaSetupRecaptcha", action: "empresa_setup_password" }) : null,
    recovery: window.PCSRecaptcha ? window.PCSRecaptcha.createManager({ containerId: "empresaRecoveryRecaptcha", action: "empresa_recovery" }) : null,
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

  function readEmpresaIDFromStorage() {
    var keys = ["active_empresa_id", "empresa_id", "admin_empresa_id"];
    var stores = [];
    try { stores.push(window.sessionStorage); } catch (error) {}
    try { stores.push(window.localStorage); } catch (error) {}

    for (var storeIndex = 0; storeIndex < stores.length; storeIndex += 1) {
      var store = stores[storeIndex];
      if (!store) continue;
      for (var keyIndex = 0; keyIndex < keys.length; keyIndex += 1) {
        var parsed = parsePositiveInt(store.getItem(keys[keyIndex]));
        if (parsed > 0) {
          return parsed;
        }
      }
    }
    return 0;
  }

  function resolveEmpresaID() {
    var fromURL = parsePositiveInt(getQueryParam("empresa_id") || getQueryParam("id"));
    if (fromURL > 0) {
      return persistEmpresaID(fromURL);
    }
    return readEmpresaIDFromStorage();
  }

  function updateEmpresaContextHint() {
    if (empresaInput && state.empresaID > 0) {
      empresaInput.value = String(state.empresaID);
    }
    if (!empresaContextHint) return;
    if (state.empresaID > 0) {
      empresaContextHint.textContent = "Empresa activa: " + state.empresaID + ". Este dato se enviará en todos los formularios del portal.";
    } else {
      empresaContextHint.textContent = "Todavía no hay empresa seleccionada. Ingresa el ID o abre este portal desde el botón público de administrar empresa.";
    }
  }

  function syncEmpresaIDFromInput() {
    var parsed = parsePositiveInt(empresaInput && empresaInput.value);
    if (!parsed) {
      updateEmpresaContextHint();
      return;
    }
    state.empresaID = persistEmpresaID(parsed);
    clearMessages();
    updateEmpresaContextHint();
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

  function setContractStatus(text, isWarning) {
    if (!contractStatus) return;
    contractStatus.textContent = text || "";
    contractStatus.classList.toggle("text-warning", !!text && !!isWarning);
  }

  function setContractPanelAttention(required) {
    if (!contractPanel) return;
    contractPanel.classList.toggle("requires-attention", !!required);
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
    setContractStatus("La aceptación se registrará cuando completes el flujo de acceso.", false);
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
    ["email", "setupEmail", "recoveryEmail", "resetEmail", "changeEmail"].forEach(function (id) {
      var input = document.getElementById(id);
      if (input) input.value = normalized;
    });
  }

  function withBasePayload(payload) {
    var out = payload || {};
    if (state.empresaID > 0) {
      out.empresa_id = state.empresaID;
    }
    if (contractCheckbox && contractCheckbox.checked) {
      out.accept_contract = true;
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
    setContractStatus("Debes aceptar el contrato antes de continuar.", true);
    setMessage(formKey, normalizeErrorMessage(payload, "Debes aceptar el contrato vigente para continuar."), true);
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

  function handleAuthResult(payload, formKey) {
    if (payload && payload.ok) {
      setContractPanelAttention(false);
      if (payload.empresa_id) {
        state.empresaID = persistEmpresaID(payload.empresa_id);
        updateEmpresaContextHint();
      }
      persistThemePreference(payload.apariencia);
      redirectAfterAuth(payload);
      return;
    }

    if (payload && payload.contract_acceptance_required) {
      handleContractRequirement(payload, formKey);
      return;
    }

    if (payload && payload.password_setup_required) {
      showForm("setup", {
        email: payload.email,
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

  function requireEmpresaIDIfNeeded(formKey) {
    if (state.empresaID > 0) return true;
    setMessage(formKey, "Debes indicar la empresa antes de continuar.", true);
    if (empresaInput && typeof empresaInput.focus === "function") {
      empresaInput.focus();
    }
    return false;
  }

  function onLoginSubmit(event) {
    event.preventDefault();
    if (!requireEmpresaIDIfNeeded("login")) return;

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
    if (!requireEmpresaIDIfNeeded("setup")) return;

    var btn = document.getElementById("btnCrearPassword");
    var prevText = btn ? String(btn.textContent || "") : "";
    function setSetupBusy(isBusy) {
      if (!btn) return;
      btn.disabled = !!isBusy;
      btn.textContent = isBusy ? "Guardando..." : prevText;
    }
    setMessage("setup", "", false);
    setSetupBusy(true);

    ensureRecaptcha("setup").then(function (token) {
      if (token === null) return;
      return submitJSON("/api/empresa/usuarios/establecer_password", withBasePayload({
        email: document.getElementById("setupEmail").value,
        documento_identidad: document.getElementById("setupDocumento").value,
        password: document.getElementById("setupPassword").value,
        password_confirm: document.getElementById("setupPasswordConfirm").value,
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
    if (!requireEmpresaIDIfNeeded("recovery")) return;

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

  function onResetSubmit(event) {
    event.preventDefault();
    if (!requireEmpresaIDIfNeeded("reset")) return;

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
    if (!requireEmpresaIDIfNeeded("change")) return;

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
        setContractStatus("La aceptación se registrará cuando completes el flujo de acceso.", false);
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

  document.addEventListener("keydown", function (event) {
    if (event.key === "Escape" && contractDialog && !contractDialog.hidden) {
      closeContractDialog();
    }
  });

  if (forms.login) forms.login.addEventListener("submit", onLoginSubmit);
  if (forms.setup) forms.setup.addEventListener("submit", onSetupSubmit);
  if (forms.recovery) forms.recovery.addEventListener("submit", onRecoverySubmit);
  if (forms.reset) forms.reset.addEventListener("submit", onResetSubmit);
  if (forms.change) forms.change.addEventListener("submit", onChangePasswordSubmit);

  bindButton("btnMostrarRegistro", function () {
    showForm("setup", { email: document.getElementById("email").value || urlEmail });
  });
  bindButton("btnVolverALoginHero", function () {
    showForm("login");
  });
  bindButton("btnMostrarRecuperacionHero", function () {
    showForm("recovery", { email: document.getElementById("email").value || urlEmail });
  });
  bindButton("linkGoSetup", function () {
    showForm("setup", { email: document.getElementById("email").value || urlEmail });
  });
  bindButton("linkGoRecovery", function () {
    showForm("recovery", { email: document.getElementById("email").value || urlEmail });
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
  bindButton("btnIrAReset", function () {
    showForm("reset", { email: document.getElementById("recoveryEmail").value });
  });
  bindButton("btnVolverDesdeReset", function () {
    showForm("login", { email: document.getElementById("resetEmail").value });
  });
  bindButton("btnVolverDesdeCambio", function () {
    showForm("login", { email: document.getElementById("changeEmail").value });
  });

  if (urlToken) {
    showForm("reset", {
      email: urlEmail,
      token: urlToken,
      message: "Token detectado en la URL. Define tu nueva contraseña para continuar."
    });
  } else {
    showForm("login", { email: urlEmail });
  }
})();
