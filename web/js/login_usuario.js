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
      empresaContextHint.textContent = "Empresa activa: " + state.empresaID + ". Este dato se enviara en todos los formularios del portal.";
    } else {
      empresaContextHint.textContent = "Todavia no hay empresa seleccionada. Ingresa el ID o abre este portal desde el boton publico de administrar empresa.";
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
    target.style.color = isError ? "#ef5350" : "";
  }

  function setContractStatus(text, isWarning) {
    if (!contractStatus) return;
    contractStatus.textContent = text || "";
    contractStatus.style.color = isWarning ? "#f7cd7c" : "";
  }

  function setContractPanelAttention(required) {
    if (!contractPanel) return;
    contractPanel.classList.toggle("requires-attention", !!required);
  }

  function applyContract(contract) {
    state.contract = contract || null;
    if (!contract) {
      contractTitle.textContent = "Contrato no disponible";
      contractVersion.textContent = "Version --";
      contractSummary.textContent = "No fue posible cargar el contrato vigente. Intenta recargar esta pagina.";
      contractNote.textContent = "";
      setContractStatus("Si el problema persiste, informa al administrador.", true);
      return;
    }

    contractTitle.textContent = contract.titulo || "Contrato vigente";
    contractVersion.textContent = "Version " + String(contract.version || "--");
    contractSummary.textContent = contract.resumen || "Lee el contrato antes de continuar con el acceso.";
    contractNote.textContent = contract.nota_aceptacion || "";
    contractLink.href = "/contrato.html" + (contract.version ? "?version=" + encodeURIComponent(contract.version) : "");
    setContractStatus("La aceptacion se registrara cuando completes el flujo de acceso.", false);
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
      forms[key].style.display = key === selected ? "" : "none";
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
        message: payload.message || "Completa tu contrasena inicial para continuar."
      });
      return;
    }

    if (payload && payload.password_rotation_required) {
      showForm("change", {
        email: payload.email,
        message: payload.message || "Debes cambiar tu contrasena antes de continuar."
      });
      return;
    }

    throw new Error(normalizeErrorMessage(payload, "No fue posible completar el acceso."));
  }

  function submitJSON(url, payload) {
    return fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "same-origin",
      body: JSON.stringify(payload)
    }).then(function (response) {
      return readResponsePayload(response).then(function (body) {
        if (!response.ok) {
          throw new Error(normalizeErrorMessage(body, "HTTP " + response.status));
        }
        return body;
      });
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

    submitJSON("/api/empresa/usuarios/login", withBasePayload({
      email: document.getElementById("email").value,
      password: document.getElementById("password").value
    }))
      .then(function (payload) {
        handleAuthResult(payload, "login");
      })
      .catch(function (error) {
        setMessage("login", error.message || "No fue posible iniciar sesion.", true);
      });
  }

  function onSetupSubmit(event) {
    event.preventDefault();
    if (!requireEmpresaIDIfNeeded("setup")) return;

    submitJSON("/api/empresa/usuarios/establecer_password", withBasePayload({
      email: document.getElementById("setupEmail").value,
      documento_identidad: document.getElementById("setupDocumento").value,
      password: document.getElementById("setupPassword").value,
      password_confirm: document.getElementById("setupPasswordConfirm").value
    }))
      .then(function (payload) {
        handleAuthResult(payload, "setup");
      })
      .catch(function (error) {
        setMessage("setup", error.message || "No fue posible crear la contrasena.", true);
      });
  }

  function onRecoverySubmit(event) {
    event.preventDefault();
    if (!requireEmpresaIDIfNeeded("recovery")) return;

    submitJSON("/api/empresa/usuarios/solicitar_recuperacion_password", withBasePayload({
      email: document.getElementById("recoveryEmail").value
    }))
      .then(function (payload) {
        prefillEmail(document.getElementById("recoveryEmail").value);
        setMessage("recovery", normalizeErrorMessage(payload, "Si el correo existe, enviaremos instrucciones para recuperar la contrasena."), false);
      })
      .catch(function (error) {
        setMessage("recovery", error.message || "No fue posible iniciar la recuperacion.", true);
      });
  }

  function onResetSubmit(event) {
    event.preventDefault();
    if (!requireEmpresaIDIfNeeded("reset")) return;

    submitJSON("/api/empresa/usuarios/restablecer_password", withBasePayload({
      email: document.getElementById("resetEmail").value,
      token: document.getElementById("resetToken").value,
      password: document.getElementById("resetPassword").value,
      password_confirm: document.getElementById("resetPasswordConfirm").value
    }))
      .then(function (payload) {
        handleAuthResult(payload, "reset");
      })
      .catch(function (error) {
        setMessage("reset", error.message || "No fue posible restablecer la contrasena.", true);
      });
  }

  function onChangePasswordSubmit(event) {
    event.preventDefault();
    if (!requireEmpresaIDIfNeeded("change")) return;

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
        setMessage("change", error.message || "No fue posible cambiar la contrasena.", true);
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
        setContractStatus("La aceptacion se registrara cuando completes el flujo de acceso.", false);
      }
    });
  }

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
      message: "Token detectado en la URL. Define tu nueva contrasena para continuar."
    });
  } else {
    showForm("login", { email: urlEmail });
  }
})();
function setMessage(targetId, text, isError) {
  var el = document.getElementById(targetId);
  el.textContent = text || "";
  el.style.color = isError ? "#ef5350" : "";
}

function getQueryParam(name) {
  var params = new URLSearchParams(window.location.search || "");
  return params.get(name);
}

function parsePositiveInt(value) {
  var n = Number(String(value == null ? "" : value).trim());
  if (!Number.isFinite(n)) return 0;
  n = Math.trunc(n);
  return n > 0 ? n : 0;
}

function readEmpresaIDFromStorage() {
  var keys = ["active_empresa_id", "empresa_id", "admin_empresa_id"];
  var stores = [];
  try {
    stores.push(window.sessionStorage);
  } catch (e) {}
  try {
    stores.push(window.localStorage);
  } catch (e) {}

  for (var s = 0; s < stores.length; s += 1) {
    var store = stores[s];
    if (!store) continue;
    for (var i = 0; i < keys.length; i += 1) {
      var raw = "";
      try {
        raw = store.getItem(keys[i]) || "";
      } catch (e) {
        raw = "";
      }
      var parsed = parsePositiveInt(raw);
      if (parsed > 0) return parsed;
    }
  }
  return 0;
}

function resolveEmpresaID() {
  var queryID = parsePositiveInt(getQueryParam("empresa_id") || getQueryParam("id"));
  if (queryID > 0) return queryID;
  return readEmpresaIDFromStorage();
}

function persistEmpresaIDContext(value) {
  var parsed = parsePositiveInt(value);
  if (parsed <= 0) {
    return;
  }

  var asText = String(parsed);
  var stores = [];
  try {
    stores.push(window.sessionStorage);
  } catch (e) {}
  try {
    stores.push(window.localStorage);
  } catch (e) {}

  for (var i = 0; i < stores.length; i += 1) {
    var store = stores[i];
    if (!store) continue;
    try {
      store.setItem("active_empresa_id", asText);
      store.setItem("empresa_id", asText);
      store.setItem("admin_empresa_id", asText);
    } catch (e) {}
  }
}

function readEmpresaIDFromInput() {
  var input = document.getElementById("empresaID");
  if (!input) {
    return 0;
  }
  return parsePositiveInt(input.value);
}

function refreshEmpresaIDInput() {
  var input = document.getElementById("empresaID");
  if (!input) {
    return;
  }
  input.value = empresaID > 0 ? String(empresaID) : "";
}

function setEmpresaID(value) {
  var parsed = parsePositiveInt(value);
  if (parsed <= 0) {
    return 0;
  }
  empresaID = parsed;
  persistEmpresaIDContext(empresaID);
  refreshEmpresaIDInput();
  return empresaID;
}

var empresaID = resolveEmpresaID();
var tokenRecuperacionPrefill = (getQueryParam("token_recuperacion") || "").trim();
var emailRecuperacionPrefill = (getQueryParam("email") || "").trim();

function ensureEmpresaScope(targetId) {
  if (empresaID <= 0) {
    var typedEmpresaID = readEmpresaIDFromInput();
    if (typedEmpresaID > 0) {
      setEmpresaID(typedEmpresaID);
    }
  }
  if (empresaID > 0) {
    persistEmpresaIDContext(empresaID);
    return true;
  }
  setMessage(targetId, "Falta empresa_id. Ingresa el Empresa ID para iniciar sesion de empresa.", true);
  return false;
}

function resolveRedirectURLForEmpresa(rawURL) {
  var fallback = "/administrar_empresa.html";
  var candidate = String(rawURL || "").trim();
  if (!candidate) {
    candidate = fallback;
  }

  try {
    var parsed = new URL(candidate, window.location.origin);
    var existingEmpresa = parsePositiveInt(parsed.searchParams.get("empresa_id") || parsed.searchParams.get("id"));
    if (empresaID > 0 && existingEmpresa <= 0) {
      parsed.searchParams.set("empresa_id", String(empresaID));
      if (!parsed.searchParams.get("id")) {
        parsed.searchParams.set("id", String(empresaID));
      }
    }
    return parsed.pathname + parsed.search + parsed.hash;
  } catch (e) {
    return candidate;
  }
}

async function getErrorMessage(res) {
  var t = await res.text();
  return t || "HTTP " + res.status;
}

function showSetupForm(email, msg) {
  document.getElementById("loginUsuarioForm").style.display = "none";
  document.getElementById("recoveryRequestForm").style.display = "none";
  document.getElementById("resetPasswordForm").style.display = "none";
  document.getElementById("changePasswordForm").style.display = "none";
  document.getElementById("setupPasswordForm").style.display = "block";
  document.getElementById("setupEmail").value = email || "";
  setMessage("setupMsg", msg || "Primer ingreso detectado. Define tu contrasena para continuar.", false);
  document.getElementById("setupDocumento").focus();
}

function showLoginForm() {
  document.getElementById("setupPasswordForm").style.display = "none";
  document.getElementById("recoveryRequestForm").style.display = "none";
  document.getElementById("resetPasswordForm").style.display = "none";
  document.getElementById("changePasswordForm").style.display = "none";
  document.getElementById("loginUsuarioForm").style.display = "block";
  setMessage("setupMsg", "", false);
  setMessage("recoveryMsg", "", false);
  setMessage("resetMsg", "", false);
  setMessage("changePasswordMsg", "", false);
  setMessage("msg", "", false);
  document.getElementById("password").focus();
}

function showRecoveryRequestForm(email) {
  document.getElementById("loginUsuarioForm").style.display = "none";
  document.getElementById("setupPasswordForm").style.display = "none";
  document.getElementById("resetPasswordForm").style.display = "none";
  document.getElementById("changePasswordForm").style.display = "none";
  document.getElementById("recoveryRequestForm").style.display = "block";
  document.getElementById("recoveryEmail").value = (email || "").trim();
  setMessage("recoveryMsg", "", false);
  document.getElementById("recoveryEmail").focus();
}

function showResetForm(email, token, msg) {
  document.getElementById("loginUsuarioForm").style.display = "none";
  document.getElementById("setupPasswordForm").style.display = "none";
  document.getElementById("recoveryRequestForm").style.display = "none";
  document.getElementById("changePasswordForm").style.display = "none";
  document.getElementById("resetPasswordForm").style.display = "block";
  document.getElementById("resetEmail").value = (email || "").trim();
  document.getElementById("resetToken").value = (token || "").trim();
  setMessage("resetMsg", msg || "", false);
  if ((token || "").trim()) {
    document.getElementById("resetPassword").focus();
  } else {
    document.getElementById("resetToken").focus();
  }
}

function showChangePasswordForm(email, msg) {
  document.getElementById("loginUsuarioForm").style.display = "none";
  document.getElementById("setupPasswordForm").style.display = "none";
  document.getElementById("recoveryRequestForm").style.display = "none";
  document.getElementById("resetPasswordForm").style.display = "none";
  document.getElementById("changePasswordForm").style.display = "block";
  document.getElementById("changeEmail").value = (email || "").trim();
  setMessage("changePasswordMsg", msg || "", false);
  document.getElementById("changeCurrentPassword").focus();
}

document.getElementById("loginUsuarioForm").addEventListener("submit", async function (ev) {
  ev.preventDefault();

  if (!ensureEmpresaScope("msg")) {
    return;
  }

  var btn = document.getElementById("btnIngresar");
  var email = (document.getElementById("email").value || "").trim();
  var password = document.getElementById("password").value || "";

  if (!email) {
    setMessage("msg", "Debes completar el correo.", true);
    return;
  }

  btn.disabled = true;
  var prevText = btn.textContent;
  btn.textContent = "Ingresando...";
  setMessage("msg", "", false);

  try {
    var loginURL = "/api/empresa/usuarios/login?empresa_id=" + encodeURIComponent(String(empresaID));
    var res = await fetch(loginURL, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ empresa_id: empresaID, email: email, password: password }),
    });

    if (!res.ok) {
      throw new Error(await getErrorMessage(res));
    }

    var body = await res.json();
    if (body && body.password_setup_required) {
      showSetupForm(email, body.message || "Primer ingreso detectado. Define tu contrasena para continuar.");
      return;
    }
    if (body && body.password_rotation_required) {
      showChangePasswordForm(
        (body.email || email || "").trim(),
        body.message || "Debes actualizar tu contrasena para continuar."
      );
      return;
    }

    var redirectURL = body && body.redirect_url ? String(body.redirect_url) : "/administrar_empresa.html";
    window.location.href = resolveRedirectURLForEmpresa(redirectURL);
  } catch (err) {
    setMessage("msg", err.message || "No se pudo iniciar sesion.", true);
  } finally {
    btn.disabled = false;
    btn.textContent = prevText;
  }
});

document.getElementById("setupPasswordForm").addEventListener("submit", async function (ev) {
  ev.preventDefault();

  if (!ensureEmpresaScope("setupMsg")) {
    return;
  }

  var btn = document.getElementById("btnCrearPassword");
  var email = (document.getElementById("setupEmail").value || "").trim();
  var documento = (document.getElementById("setupDocumento").value || "").trim();
  var password = document.getElementById("setupPassword").value || "";
  var passwordConfirm = document.getElementById("setupPasswordConfirm").value || "";

  if (!email || !documento || !password || !passwordConfirm) {
    setMessage("setupMsg", "Debes completar todos los campos.", true);
    return;
  }
  if (password !== passwordConfirm) {
    setMessage("setupMsg", "Las contrasenas no coinciden.", true);
    return;
  }

  btn.disabled = true;
  var prevText = btn.textContent;
  btn.textContent = "Guardando...";
  setMessage("setupMsg", "", false);

  try {
    var setupURL = "/api/empresa/usuarios/establecer_password?empresa_id=" + encodeURIComponent(String(empresaID));
    var res = await fetch(setupURL, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        empresa_id: empresaID,
        email: email,
        documento_identidad: documento,
        password: password,
        password_confirm: passwordConfirm,
      }),
    });

    if (!res.ok) {
      throw new Error(await getErrorMessage(res));
    }

    var body = await res.json();
    var redirectURL = body && body.redirect_url ? String(body.redirect_url) : "/administrar_empresa.html";
    window.location.href = resolveRedirectURLForEmpresa(redirectURL);
  } catch (err) {
    setMessage("setupMsg", err.message || "No se pudo crear la contrasena.", true);
  } finally {
    btn.disabled = false;
    btn.textContent = prevText;
  }
});

document.getElementById("recoveryRequestForm").addEventListener("submit", async function (ev) {
  ev.preventDefault();

  if (!ensureEmpresaScope("recoveryMsg")) {
    return;
  }

  var btn = document.getElementById("btnSolicitarRecuperacion");
  var email = (document.getElementById("recoveryEmail").value || "").trim();
  if (!email) {
    setMessage("recoveryMsg", "Debes completar el correo.", true);
    return;
  }

  btn.disabled = true;
  var prevText = btn.textContent;
  btn.textContent = "Enviando...";
  setMessage("recoveryMsg", "", false);

  try {
    var recoveryURL = "/api/empresa/usuarios/solicitar_recuperacion_password?empresa_id=" + encodeURIComponent(String(empresaID));
    var res = await fetch(recoveryURL, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ empresa_id: empresaID, email: email }),
    });

    if (!res.ok) {
      throw new Error(await getErrorMessage(res));
    }

    var body = await res.json();
    var delivery = body && body.delivery ? String(body.delivery) : "masked";
    var msg = body && body.message ? String(body.message) : "Si el correo existe, enviaremos instrucciones para recuperar la contraseña.";
    if (delivery === "manual") {
      msg += " Si no recibes correo, solicita soporte al administrador de la empresa.";
    }
    setMessage("recoveryMsg", msg, false);
    showResetForm(email, "", "Cuando recibas el token, ingrésalo aquí para continuar.");
  } catch (err) {
    setMessage("recoveryMsg", err.message || "No se pudo solicitar la recuperación.", true);
  } finally {
    btn.disabled = false;
    btn.textContent = prevText;
  }
});

document.getElementById("resetPasswordForm").addEventListener("submit", async function (ev) {
  ev.preventDefault();

  if (!ensureEmpresaScope("resetMsg")) {
    return;
  }

  var btn = document.getElementById("btnRestablecerPassword");
  var email = (document.getElementById("resetEmail").value || "").trim();
  var token = (document.getElementById("resetToken").value || "").trim();
  var password = document.getElementById("resetPassword").value || "";
  var passwordConfirm = document.getElementById("resetPasswordConfirm").value || "";

  if (!email || !token || !password || !passwordConfirm) {
    setMessage("resetMsg", "Debes completar todos los campos.", true);
    return;
  }
  if (password !== passwordConfirm) {
    setMessage("resetMsg", "Las contrasenas no coinciden.", true);
    return;
  }

  btn.disabled = true;
  var prevText = btn.textContent;
  btn.textContent = "Restableciendo...";
  setMessage("resetMsg", "", false);

  try {
    var resetURL = "/api/empresa/usuarios/restablecer_password?empresa_id=" + encodeURIComponent(String(empresaID));
    var res = await fetch(resetURL, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        empresa_id: empresaID,
        email: email,
        token: token,
        password: password,
        password_confirm: passwordConfirm,
      }),
    });

    if (!res.ok) {
      throw new Error(await getErrorMessage(res));
    }

    var body = await res.json();
    var redirectURL = body && body.redirect_url ? String(body.redirect_url) : "/administrar_empresa.html";
    window.location.href = resolveRedirectURLForEmpresa(redirectURL);
  } catch (err) {
    setMessage("resetMsg", err.message || "No se pudo restablecer la contrasena.", true);
  } finally {
    btn.disabled = false;
    btn.textContent = prevText;
  }
});

document.getElementById("changePasswordForm").addEventListener("submit", async function (ev) {
  ev.preventDefault();

  if (!ensureEmpresaScope("changePasswordMsg")) {
    return;
  }

  var btn = document.getElementById("btnCambiarPassword");
  var email = (document.getElementById("changeEmail").value || "").trim();
  var currentPassword = document.getElementById("changeCurrentPassword").value || "";
  var newPassword = document.getElementById("changeNewPassword").value || "";
  var newPasswordConfirm = document.getElementById("changeNewPasswordConfirm").value || "";

  if (!email || !currentPassword || !newPassword || !newPasswordConfirm) {
    setMessage("changePasswordMsg", "Debes completar todos los campos.", true);
    return;
  }
  if (newPassword !== newPasswordConfirm) {
    setMessage("changePasswordMsg", "Las contrasenas no coinciden.", true);
    return;
  }

  btn.disabled = true;
  var prevText = btn.textContent;
  btn.textContent = "Actualizando...";
  setMessage("changePasswordMsg", "", false);

  try {
    var changeURL = "/api/empresa/usuarios/cambiar_password?empresa_id=" + encodeURIComponent(String(empresaID));
    var res = await fetch(changeURL, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        empresa_id: empresaID,
        email: email,
        current_password: currentPassword,
        new_password: newPassword,
        new_password_confirm: newPasswordConfirm,
      }),
    });

    if (!res.ok) {
      throw new Error(await getErrorMessage(res));
    }

    var body = await res.json();
    var redirectURL = body && body.redirect_url ? String(body.redirect_url) : "/administrar_empresa.html";
    window.location.href = resolveRedirectURLForEmpresa(redirectURL);
  } catch (err) {
    setMessage("changePasswordMsg", err.message || "No se pudo cambiar la contrasena.", true);
  } finally {
    btn.disabled = false;
    btn.textContent = prevText;
  }
});

document.getElementById("btnVolverLogin").addEventListener("click", function () {
  showLoginForm();
});

document.getElementById("btnMostrarRecuperacion").addEventListener("click", function () {
  var email = (document.getElementById("email").value || "").trim();
  showRecoveryRequestForm(email);
});

document.getElementById("btnMostrarCambioClave").addEventListener("click", function () {
  var email = (document.getElementById("email").value || "").trim();
  showChangePasswordForm(email, "");
});

document.getElementById("btnVolverDesdeRecuperacion").addEventListener("click", function () {
  showLoginForm();
});

document.getElementById("btnVolverDesdeReset").addEventListener("click", function () {
  showLoginForm();
});

document.getElementById("btnVolverDesdeCambio").addEventListener("click", function () {
  showLoginForm();
});

(function initializeLoginUsuario() {
  var emailInput = document.getElementById("email");
  var empresaInput = document.getElementById("empresaID");

  if (!empresaInput) {
    return;
  }

  refreshEmpresaIDInput();

  if (emailInput && emailRecuperacionPrefill) {
    emailInput.value = emailRecuperacionPrefill;
  }

  var syncEmpresaFromInput = function () {
    var raw = String(empresaInput.value || "").trim();
    if (!raw) {
      return;
    }
    var parsed = parsePositiveInt(raw);
    if (parsed <= 0) {
      setMessage("msg", "Empresa ID invalido.", true);
      return;
    }
    setEmpresaID(parsed);
    setMessage("msg", "", false);
  };

  empresaInput.addEventListener("change", syncEmpresaFromInput);
  empresaInput.addEventListener("blur", syncEmpresaFromInput);
})();

if (tokenRecuperacionPrefill) {
  showResetForm(emailRecuperacionPrefill, tokenRecuperacionPrefill, "Token detectado en la URL. Completa tu nueva contrasena para continuar.");
}
