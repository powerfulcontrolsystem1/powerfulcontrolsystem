function setMessage(targetId, text, isError) {
  var el = document.getElementById(targetId);
  el.textContent = text || "";
  el.style.color = isError ? "#ef5350" : "";
}

async function getErrorMessage(res) {
  var t = await res.text();
  return t || "HTTP " + res.status;
}

function showSetupForm(email, msg) {
  document.getElementById("loginUsuarioForm").style.display = "none";
  document.getElementById("setupPasswordForm").style.display = "block";
  document.getElementById("setupEmail").value = email || "";
  setMessage("setupMsg", msg || "Primer ingreso detectado. Define tu contrasena para continuar.", false);
  document.getElementById("setupDocumento").focus();
}

function showLoginForm() {
  document.getElementById("setupPasswordForm").style.display = "none";
  document.getElementById("loginUsuarioForm").style.display = "block";
  setMessage("setupMsg", "", false);
  setMessage("msg", "", false);
  document.getElementById("password").focus();
}

document.getElementById("loginUsuarioForm").addEventListener("submit", async function (ev) {
  ev.preventDefault();

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
    var res = await fetch("/api/empresa/usuarios/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: email, password: password }),
    });

    if (!res.ok) {
      throw new Error(await getErrorMessage(res));
    }

    var body = await res.json();
    if (body && body.password_setup_required) {
      showSetupForm(email, body.message || "Primer ingreso detectado. Define tu contrasena para continuar.");
      return;
    }

    var redirectURL = body && body.redirect_url ? String(body.redirect_url) : "/administrar_empresa.html";
    window.location.href = redirectURL;
  } catch (err) {
    setMessage("msg", err.message || "No se pudo iniciar sesion.", true);
  } finally {
    btn.disabled = false;
    btn.textContent = prevText;
  }
});

document.getElementById("setupPasswordForm").addEventListener("submit", async function (ev) {
  ev.preventDefault();

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
    var res = await fetch("/api/empresa/usuarios/establecer_password", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
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
    window.location.href = redirectURL;
  } catch (err) {
    setMessage("setupMsg", err.message || "No se pudo crear la contrasena.", true);
  } finally {
    btn.disabled = false;
    btn.textContent = prevText;
  }
});

document.getElementById("btnVolverLogin").addEventListener("click", function () {
  showLoginForm();
});
