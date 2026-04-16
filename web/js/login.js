(function () {
  var rememberCheckbox = document.getElementById("rememberAccount");
  var googleBtn = document.querySelector(".google-btn");
  var hintBlock = document.getElementById("rememberedAccountHint");
  var hintEmail = document.getElementById("rememberedAccountEmail");
  var clearRememberLink = document.getElementById("clearRememberedAccount");

  var KEY_REMEMBER_FLAG = "rememberAccount";
  var KEY_REMEMBER_EMAIL = "rememberedEmail";
  var SESSION_STATE_COOKIE = "browser_session_active";

  function safeGetItem(key) {
    try { return localStorage.getItem(key) || ""; } catch (e) { return ""; }
  }

  function safeSetItem(key, value) {
    try { localStorage.setItem(key, value); } catch (e) {}
  }

  function normalizeEmail(value) {
    return String(value || "").trim();
  }

  function isPlausibleEmail(value) {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(normalizeEmail(value));
  }

  function clearRememberedEmail() {
    try { localStorage.removeItem(KEY_REMEMBER_EMAIL); } catch (e) {}
  }

  function setRememberedEmail(email) {
    var normalized = normalizeEmail(email);
    if (!isPlausibleEmail(normalized)) {
      return false;
    }
    safeSetItem(KEY_REMEMBER_EMAIL, normalized);
    return true;
  }

  function clearRememberState() {
    try {
      localStorage.removeItem(KEY_REMEMBER_FLAG);
      localStorage.removeItem(KEY_REMEMBER_EMAIL);
    } catch (e) {}
  }

  function clearRememberOnLogoutIfNeeded() {
    if (isRememberEnabled()) {
      return;
    }
    clearRememberState();
  }

  function isRememberEnabled() {
    return safeGetItem(KEY_REMEMBER_FLAG) === "1";
  }

  function getRememberedEmail() {
    var normalized = normalizeEmail(safeGetItem(KEY_REMEMBER_EMAIL));
    if (!normalized) {
      return "";
    }
    if (!isPlausibleEmail(normalized)) {
      clearRememberedEmail();
      return "";
    }
    return normalized;
  }

  function hasBrowserSessionCookie() {
    try {
      return new RegExp("(?:^|;\\s*)" + SESSION_STATE_COOKIE + "=1(?:;|$)").test(document.cookie || "");
    } catch (e) {
      return false;
    }
  }

  function buildLoginUrl() {
    var url = "/auth/google/login";
    var rememberedEmail = getRememberedEmail();
    if (isRememberEnabled() && rememberedEmail) {
      url += "?login_hint=" + encodeURIComponent(rememberedEmail);
    }
    return url;
  }

  function syncRememberedEmailFromSession() {
    if (!isRememberEnabled()) {
      return;
    }
    if (!hasBrowserSessionCookie()) {
      return;
    }
    if (getRememberedEmail()) {
      return;
    }

    fetch("/me", { credentials: "same-origin" })
      .then(function (res) {
        if (!res.ok) {
          throw new Error("unauth");
        }
        return res.json();
      })
      .then(function (admin) {
        var email = admin && admin.email ? String(admin.email).trim() : "";
        if (!email) {
          return;
        }
        if (!setRememberedEmail(email)) {
          clearRememberedEmail();
        }
        refreshRememberUI();
      })
      .catch(function () {});
  }

  function refreshRememberUI() {
    var rememberedEmail = getRememberedEmail();
    var enabled = isRememberEnabled();

    if (rememberCheckbox) {
      rememberCheckbox.checked = enabled;
    }

    if (googleBtn) {
      googleBtn.setAttribute("href", buildLoginUrl());
    }

    if (hintBlock && hintEmail) {
      if (enabled && rememberedEmail) {
        hintEmail.textContent = rememberedEmail;
        hintBlock.hidden = false;
      } else {
        hintEmail.textContent = "";
        hintBlock.hidden = true;
      }
    }
  }

  if (rememberCheckbox) {
    rememberCheckbox.addEventListener("change", function () {
      if (rememberCheckbox.checked) {
        safeSetItem(KEY_REMEMBER_FLAG, "1");
        syncRememberedEmailFromSession();
      } else {
        clearRememberState();
      }
      refreshRememberUI();
    });
  }

  if (clearRememberLink) {
    clearRememberLink.addEventListener("click", function (event) {
      event.preventDefault();
      clearRememberState();
      if (rememberCheckbox) {
        rememberCheckbox.checked = false;
      }
      refreshRememberUI();
    });
  }

  if (googleBtn) {
    googleBtn.addEventListener("click", function (event) {
      event.preventDefault();
      if (rememberCheckbox && rememberCheckbox.checked) {
        safeSetItem(KEY_REMEMBER_FLAG, "1");
      } else {
        clearRememberState();
      }

      var target = buildLoginUrl();
      window.location.href = target;
    });
  }

  var logoutLink = document.querySelector('.fm-item[href="/auth/logout"]');
  if (logoutLink) {
    logoutLink.addEventListener("click", function () {
      clearRememberOnLogoutIfNeeded();
    });
  }

  refreshRememberUI();
  syncRememberedEmailFromSession();
  // --- Admin email/password flows ---
  var emailInput = document.getElementById('adminEmail');
  var nameInput = document.getElementById('adminName');
  var passInput = document.getElementById('adminPassword');
  var loginBtn = document.getElementById('emailLoginBtn');
  var registerBtn = document.getElementById('registerBtn');
  var forgotLinkElem = document.getElementById('forgotLink');
  var messageDiv = document.getElementById('emailLoginMessage');

  function showMsg(text, isError) {
    if (!messageDiv) return;
    messageDiv.style.display = 'block';
    messageDiv.textContent = text;
    messageDiv.style.color = isError ? 'crimson' : 'green';
    setTimeout(function(){ messageDiv.style.display = 'none'; }, 8000);
  }

  async function postJson(url, body) {
    var res = await fetch(url, {
      method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(body), credentials: 'same-origin'
    });
    var txt = await res.text();
    try { return {ok: res.ok, status: res.status, json: JSON.parse(txt)}; } catch (e) { return {ok: res.ok, status: res.status, text: txt}; }
  }

  if (loginBtn) {
    loginBtn.addEventListener('click', async function(e){
      e.preventDefault();
      var email = (emailInput && emailInput.value) ? emailInput.value.trim() : '';
      var password = (passInput && passInput.value) ? passInput.value : '';
      if (!email || !password) { showMsg('Email y contraseña requeridos', true); return; }
      var r = await postJson('/super/api/administradores/login', {email: email, password: password});
      if (r.ok && r.json && r.json.redirect_url) {
        window.location.href = r.json.redirect_url;
        return;
      }
      if (r.status === 403) { showMsg('Debes confirmar tu correo. Revisa el email de confirmación.', true); return; }
      if (r.json && r.json.password_setup_required) { showMsg('Tu cuenta requiere establecer contraseña. Revisa tu correo.', true); return; }
      showMsg('Credenciales inválidas', true);
    });
  }

  if (registerBtn) {
    registerBtn.addEventListener('click', async function(e){
      e.preventDefault();
      var email = (emailInput && emailInput.value) ? emailInput.value.trim() : '';
      var name = (nameInput && nameInput.value) ? nameInput.value.trim() : '';
      var password = (passInput && passInput.value) ? passInput.value : '';
      if (!email || !password) { showMsg('Email y contraseña requeridos para registrarse', true); return; }
      var r = await postJson('/super/api/administradores/register', {email: email, name: name, password: password});
      if (r.ok && r.json && r.json.ok) { showMsg('Registro exitoso. Revisa tu correo para confirmar.', false); return; }
      showMsg('Error en el registro', true);
    });
  }

  if (forgotLinkElem) {
    forgotLinkElem.addEventListener('click', async function(e){
      e.preventDefault();
      var email = (emailInput && emailInput.value) ? emailInput.value.trim() : '';
      if (!email) { email = prompt('Introduce tu correo para recuperación:'); }
      if (!email) { return; }
      var r = await postJson('/super/api/administradores/solicitar_recuperacion', {email: email});
      if (r.ok && r.json && r.json.ok) { showMsg('Si existe la cuenta, recibirás un correo con instrucciones.', false); return; }
      showMsg('Error al solicitar recuperación', true);
    });
  }

  // Si venimos con token de recuperación en query string, lanzar flujo de restablecimiento
  (function handleRecoveryFromQuery(){
    try {
      var params = new URLSearchParams(window.location.search);
      var token = params.get('token_recuperacion');
      var email = params.get('email');
      if (token && email) {
        var pwd = prompt('Token detectado. Ingresa nueva contraseña:');
        if (!pwd) { return; }
        var pwd2 = prompt('Confirma nueva contraseña:');
        if (pwd !== pwd2) { alert('Las contraseñas no coinciden'); return; }
        postJson('/super/api/administradores/restablecer_password', {email: email, token: token, password: pwd}).then(function(res){
          if (res.ok && res.json && res.json.redirect_url) {
            window.location.href = res.json.redirect_url;
            return;
          }
          alert('Error al restablecer contraseña');
        });
      }
    } catch (e) {}
  })();
})();
