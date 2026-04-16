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
