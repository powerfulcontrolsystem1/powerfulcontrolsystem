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
  var emailLoginForm = document.getElementById('emailLoginForm');
  var forgotPasswordForm = document.getElementById('forgotPasswordForm');
  var resetPasswordForm = document.getElementById('resetPasswordForm');
  var emailInput = document.getElementById('adminEmail');
  var passInput = document.getElementById('adminPassword');
  var forgotEmailInput = document.getElementById('forgotEmail');
  var resetEmailInput = document.getElementById('resetEmail');
  var resetTokenInput = document.getElementById('resetToken');
  var resetPasswordInput = document.getElementById('resetPassword');
  var resetPasswordConfirmInput = document.getElementById('resetPasswordConfirm');
  var loginBtn = document.getElementById('emailLoginBtn');
  var forgotBtn = document.getElementById('forgotPasswordBtn');
  var resetBtn = document.getElementById('resetPasswordBtn');
  var forgotLinkElem = document.getElementById('forgotLink');
  var backToLoginBtn = document.getElementById('backToLoginLink');
  var backFromResetBtn = document.getElementById('backFromResetLink');
  var loginMessageDiv = document.getElementById('emailLoginMessage');
  var forgotMessageDiv = document.getElementById('forgotPasswordMessage');
  var resetMessageDiv = document.getElementById('resetPasswordMessage');

  function showMsg(target, text, isError) {
    if (!target) {
      return;
    }
    target.style.display = text ? 'block' : 'none';
    target.textContent = text || '';
    target.style.color = isError ? 'crimson' : 'green';
  }

  function clearMsgs() {
    showMsg(loginMessageDiv, '', false);
    showMsg(forgotMessageDiv, '', false);
    showMsg(resetMessageDiv, '', false);
  }

  function setButtonBusy(button, busyText, isBusy) {
    if (!button) {
      return;
    }
    if (!button.dataset.defaultText) {
      button.dataset.defaultText = button.textContent;
    }
    button.disabled = !!isBusy;
    button.textContent = isBusy ? busyText : button.dataset.defaultText;
  }

  function prefillEmailFields(email) {
    var normalized = normalizeEmail(email);
    if (!normalized) {
      return;
    }
    if (emailInput) {
      emailInput.value = normalized;
    }
    if (forgotEmailInput) {
      forgotEmailInput.value = normalized;
    }
    if (resetEmailInput) {
      resetEmailInput.value = normalized;
    }
  }

  function clearRecoveryQueryState() {
    try {
      var nextUrl = new URL(window.location.href);
      nextUrl.searchParams.delete('token_recuperacion');
      nextUrl.searchParams.delete('token');
      nextUrl.searchParams.delete('view');
      nextUrl.searchParams.delete('modo');
      if (window.history && window.history.replaceState) {
        var nextPath = nextUrl.pathname;
        var nextSearch = nextUrl.searchParams.toString();
        window.history.replaceState({}, document.title, nextPath + (nextSearch ? '?' + nextSearch : ''));
      }
    } catch (e) {}
  }

  function showAuthView(view) {
    clearMsgs();
    if (emailLoginForm) {
      emailLoginForm.style.display = view === 'login' ? 'block' : 'none';
    }
    if (forgotPasswordForm) {
      forgotPasswordForm.style.display = view === 'forgot' ? 'block' : 'none';
    }
    if (resetPasswordForm) {
      resetPasswordForm.style.display = view === 'reset' ? 'block' : 'none';
    }
    if (view === 'login' && passInput) {
      passInput.focus();
    }
    if (view === 'forgot' && forgotEmailInput) {
      forgotEmailInput.focus();
    }
    if (view === 'reset' && resetPasswordInput) {
      resetPasswordInput.focus();
    }
  }

  function syncRememberPreference(email) {
    if (rememberCheckbox && rememberCheckbox.checked) {
      safeSetItem(KEY_REMEMBER_FLAG, '1');
      setRememberedEmail(email);
    } else {
      clearRememberState();
    }
    refreshRememberUI();
  }

  function getResponseMessage(response, fallback) {
    if (response && response.json) {
      if (response.json.message) {
        return String(response.json.message);
      }
      if (response.json.error) {
        return String(response.json.error);
      }
    }
    if (response && response.text) {
      return String(response.text);
    }
    return fallback;
  }

  async function postJson(url, body) {
    var res = await fetch(url, {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify(body),
      credentials: 'same-origin'
    });
    var txt = await res.text();
    try {
      return {ok: res.ok, status: res.status, json: JSON.parse(txt)};
    } catch (e) {
      return {ok: res.ok, status: res.status, text: txt};
    }
  }

  function openForgotPasswordView(event) {
    if (event) {
      event.preventDefault();
    }
    prefillEmailFields(emailInput && emailInput.value);
    showAuthView('forgot');
  }

  if (emailLoginForm) {
    emailLoginForm.addEventListener('submit', async function (event) {
      event.preventDefault();
      var email = normalizeEmail(emailInput && emailInput.value);
      var password = passInput && passInput.value ? passInput.value : '';
      if (!email || !password) {
        showMsg(loginMessageDiv, 'Debes ingresar correo y contraseña.', true);
        return;
      }

      setButtonBusy(loginBtn, 'Ingresando...', true);
      showMsg(loginMessageDiv, '', false);
      try {
        var response = await postJson('/super/api/administradores/login', {email: email, password: password});
        if (response.ok && response.json && response.json.redirect_url) {
          syncRememberPreference(email);
          window.location.href = response.json.redirect_url;
          return;
        }
        if (response.json && response.json.password_setup_required) {
          showMsg(loginMessageDiv, getResponseMessage(response, 'Tu cuenta todavía no tiene una contraseña activa.'), true);
          return;
        }
        if (response.status === 403) {
          showMsg(loginMessageDiv, getResponseMessage(response, 'Debes confirmar tu correo. Revisa el mensaje de confirmación.'), true);
          return;
        }
        showMsg(loginMessageDiv, getResponseMessage(response, 'Credenciales inválidas.'), true);
      } catch (error) {
        showMsg(loginMessageDiv, error && error.message ? error.message : 'No se pudo iniciar sesión por correo.', true);
      } finally {
        setButtonBusy(loginBtn, 'Ingresando...', false);
      }
    });
  }

  if (forgotLinkElem) {
    forgotLinkElem.addEventListener('click', openForgotPasswordView);
  }

  if (backToLoginBtn) {
    backToLoginBtn.addEventListener('click', function () {
      clearRecoveryQueryState();
      showAuthView('login');
    });
  }

  if (backFromResetBtn) {
    backFromResetBtn.addEventListener('click', function () {
      clearRecoveryQueryState();
      if (resetTokenInput) {
        resetTokenInput.value = '';
      }
      if (resetPasswordInput) {
        resetPasswordInput.value = '';
      }
      if (resetPasswordConfirmInput) {
        resetPasswordConfirmInput.value = '';
      }
      showAuthView('login');
    });
  }

  if (forgotPasswordForm) {
    forgotPasswordForm.addEventListener('submit', async function (event) {
      event.preventDefault();
      var email = normalizeEmail(forgotEmailInput && forgotEmailInput.value);
      if (!email) {
        showMsg(forgotMessageDiv, 'Debes indicar el correo de la cuenta administrativa.', true);
        return;
      }
      setButtonBusy(forgotBtn, 'Enviando...', true);
      showMsg(forgotMessageDiv, '', false);
      try {
        var response = await postJson('/super/api/administradores/solicitar_recuperacion', {email: email});
        if (!response.ok) {
          showMsg(forgotMessageDiv, getResponseMessage(response, 'No se pudo iniciar la recuperación de contraseña.'), true);
          return;
        }
        prefillEmailFields(email);
        showMsg(forgotMessageDiv, getResponseMessage(response, 'Si la cuenta existe y ya fue confirmada, enviaremos instrucciones para restablecer la contraseña.'), false);
      } catch (error) {
        showMsg(forgotMessageDiv, error && error.message ? error.message : 'No se pudo solicitar la recuperación de contraseña.', true);
      } finally {
        setButtonBusy(forgotBtn, 'Enviando...', false);
      }
    });
  }

  if (resetPasswordForm) {
    resetPasswordForm.addEventListener('submit', async function (event) {
      event.preventDefault();
      var email = normalizeEmail(resetEmailInput && resetEmailInput.value);
      var token = normalizeEmail(resetTokenInput && resetTokenInput.value).replace(/\s+/g, '');
      var password = resetPasswordInput && resetPasswordInput.value ? resetPasswordInput.value : '';
      var passwordConfirm = resetPasswordConfirmInput && resetPasswordConfirmInput.value ? resetPasswordConfirmInput.value : '';
      if (!email || !token || !password || !passwordConfirm) {
        showMsg(resetMessageDiv, 'Debes completar correo, token y las dos contraseñas.', true);
        return;
      }
      if (password.length < 8) {
        showMsg(resetMessageDiv, 'La nueva contraseña debe tener mínimo 8 caracteres.', true);
        return;
      }
      if (password !== passwordConfirm) {
        showMsg(resetMessageDiv, 'Las contraseñas no coinciden.', true);
        return;
      }

      setButtonBusy(resetBtn, 'Restableciendo...', true);
      showMsg(resetMessageDiv, '', false);
      try {
        var response = await postJson('/super/api/administradores/restablecer_password', {
          email: email,
          token: token,
          password: password
        });
        if (response.ok && response.json && response.json.redirect_url) {
          syncRememberPreference(email);
          window.location.href = response.json.redirect_url;
          return;
        }
        showMsg(resetMessageDiv, getResponseMessage(response, 'No se pudo restablecer la contraseña.'), true);
      } catch (error) {
        showMsg(resetMessageDiv, error && error.message ? error.message : 'No se pudo restablecer la contraseña.', true);
      } finally {
        setButtonBusy(resetBtn, 'Restableciendo...', false);
      }
    });
  }

  (function handleRecoveryFromQuery() {
    try {
      var params = new URLSearchParams(window.location.search);
      var queryEmail = normalizeEmail(params.get('email'));
      var token = normalizeEmail(params.get('token_recuperacion') || params.get('token')).replace(/\s+/g, '');
      var view = normalizeEmail(params.get('view') || params.get('modo')).toLowerCase();

      if (queryEmail) {
        prefillEmailFields(queryEmail);
      }
      if (token && resetTokenInput) {
        resetTokenInput.value = token;
      }

      if (queryEmail && token) {
        showAuthView('reset');
        showMsg(resetMessageDiv, 'Define tu nueva contraseña para completar el restablecimiento.', false);
        return;
      }
      if (view === 'reset') {
        showAuthView('reset');
        return;
      }
      if (view === 'forgot' || view === 'recuperacion') {
        showAuthView('forgot');
        return;
      }
      showAuthView('login');
    } catch (e) {
      showAuthView('login');
    }
  })();
})();
