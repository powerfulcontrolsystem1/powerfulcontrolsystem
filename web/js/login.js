(function () {
  var googleBtn = document.querySelector(".google-btn");
  var rememberedEmailStorageKey = 'pcs_admin_login_remembered_email';

  function readCookie(name) {
    var match = String(document.cookie || '').match('(^|;)\\s*' + name + '\\s*=\\s*([^;]+)');
    return match ? match.pop() : '';
  }

  function normalizeEmail(value) {
    return String(value || "").trim();
  }

  function getStoredRememberedEmail() {
    try {
      return normalizeEmail(window.localStorage.getItem(rememberedEmailStorageKey));
    } catch (error) {
      return '';
    }
  }

  function setStoredRememberedEmail(email) {
    try {
      window.localStorage.setItem(rememberedEmailStorageKey, normalizeEmail(email));
    } catch (error) {}
  }

  function clearStoredRememberedEmail() {
    try {
      window.localStorage.removeItem(rememberedEmailStorageKey);
    } catch (error) {}
  }

  if (googleBtn) {
    googleBtn.setAttribute("href", "/auth/google/login");
    googleBtn.addEventListener("click", function (event) {
      event.preventDefault();
      window.location.href = "/auth/google/login";
    });
  }

  var emailLoginForm = document.getElementById('emailLoginForm');
  var forgotPasswordForm = document.getElementById('forgotPasswordForm');
  var resetPasswordForm = document.getElementById('resetPasswordForm');
  var resetLinkToken = '';
  var emailInput = document.getElementById('adminEmail');
  var passInput = document.getElementById('adminPassword');
  var otpRow = document.getElementById('adminOtpRow');
  var otpInput = document.getElementById('adminOtpCode');
  var rememberAdminEmailCheckbox = document.getElementById('rememberAdminEmailCheckbox');
  var forgotEmailInput = document.getElementById('forgotEmail');
  var resetEmailInput = document.getElementById('resetEmail');
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
  var recaptchaManagers = {
    login: window.PCSRecaptcha ? window.PCSRecaptcha.createManager({ containerId: 'adminLoginRecaptcha', action: 'admin_login' }) : null,
    forgot: window.PCSRecaptcha ? window.PCSRecaptcha.createManager({ containerId: 'adminForgotRecaptcha', action: 'admin_forgot_password' }) : null,
    reset: window.PCSRecaptcha ? window.PCSRecaptcha.createManager({ containerId: 'adminResetRecaptcha', action: 'admin_reset_password' }) : null
  };

  function setPasswordVisibility(toggleBtn, input, isVisible) {
    if (!toggleBtn || !input) {
      return;
    }
    input.type = isVisible ? 'text' : 'password';
    toggleBtn.setAttribute('aria-pressed', isVisible ? 'true' : 'false');
    toggleBtn.setAttribute('aria-label', isVisible ? 'Ocultar contraseña' : 'Mostrar contraseña');
    toggleBtn.setAttribute('title', isVisible ? 'Ocultar contraseña' : 'Mostrar contraseña');
    toggleBtn.classList.toggle('is-visible', !!isVisible);
  }

  function initPasswordVisibilityToggles() {
    var toggles = document.querySelectorAll('.password-visibility-toggle[data-target]');
    Array.prototype.forEach.call(toggles, function (toggleBtn) {
      var targetId = toggleBtn.getAttribute('data-target');
      var input = targetId ? document.getElementById(targetId) : null;
      setPasswordVisibility(toggleBtn, input, false);
      toggleBtn.addEventListener('click', function () {
        if (!input) {
          return;
        }
        setPasswordVisibility(toggleBtn, input, input.type === 'password');
        input.focus();
      });
    });
  }

  function isAdmin2FALoginEnabled() {
    return !!window.ADMIN_2FA_LOGIN_ENABLED;
  }

  function syncAdmin2FAFieldVisibility() {
    var enabled = isAdmin2FALoginEnabled();
    if (otpRow) {
      otpRow.classList.toggle('is-hidden', !enabled);
      otpRow.style.display = enabled ? '' : 'none';
    }
    if (otpInput) {
      otpInput.disabled = !enabled;
      if (!enabled) {
        otpInput.value = '';
      }
    }
  }

  function showMsg(target, text, isError) {
    if (!target) {
      return;
    }
    target.classList.toggle('is-hidden', !text);
    target.classList.toggle('is-visible', !!text);
    target.textContent = text || '';
    target.classList.toggle('error', !!text && !!isError);
    target.classList.toggle('success', !!text && !isError);
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

  function syncRememberedEmailWithUI() {
    var email = normalizeEmail(emailInput && emailInput.value);
    if (rememberAdminEmailCheckbox && rememberAdminEmailCheckbox.checked && email) {
      setStoredRememberedEmail(email);
      return;
    }
    clearStoredRememberedEmail();
  }

  function hydrateRememberedEmail() {
    var rememberedEmail = getStoredRememberedEmail();
    if (!rememberedEmail) {
      return;
    }
    if (rememberAdminEmailCheckbox) {
      rememberAdminEmailCheckbox.checked = true;
    }
    prefillEmailFields(rememberedEmail);
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

  function getSharedInvitationTokenFromQuery() {
    try {
      var params = new URLSearchParams(window.location.search);
      return normalizeEmail(params.get('shared_invitation_token')).replace(/\s+/g, '');
    } catch (e) {
      return '';
    }
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
    if (recaptchaManagers[view]) {
      recaptchaManagers[view].init().catch(function (error) {
        var target = view === 'login' ? loginMessageDiv : (view === 'forgot' ? forgotMessageDiv : resetMessageDiv);
        showMsg(target, error && error.message ? error.message : 'No se pudo cargar la verificación de seguridad.', true);
      });
    }
  }

  async function ensureRecaptcha(view, target) {
    var manager = recaptchaManagers[view];
    if (!manager) {
      return '';
    }
    var result;
    try {
      result = await manager.ensureToken();
    } catch (error) {
      showMsg(target, error && error.message ? error.message : 'No se pudo cargar la verificación de seguridad.', true);
      return null;
    }
    if (!result.ok) {
      showMsg(target, result.message || 'Completa el reCAPTCHA que aparece debajo del formulario para continuar.', true);
      return null;
    }
    return result.token || '';
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
    var controller = new AbortController();
    var tid = setTimeout(function () {
      controller.abort();
    }, 45000);
    var res;
    try {
      var headers = {'Content-Type': 'application/json'};
      var csrfToken = readCookie('pcs_csrf');
      if (csrfToken) {
        headers['X-CSRF-Token'] = csrfToken;
      }
      res = await fetch(url, {
        method: 'POST',
        headers: headers,
        body: JSON.stringify(body),
        credentials: 'same-origin',
        signal: controller.signal
      });
    } catch (e) {
      clearTimeout(tid);
      if (e && (e.name === 'AbortError' || (String(e.message || '').toLowerCase().indexOf('abort') >= 0))) {
        return { ok: false, status: 0, text: 'Tiempo de espera agotado al contactar el servidor.', json: null };
      }
      throw e;
    }
    clearTimeout(tid);
    var txt = await res.text();
    try {
      return {ok: res.ok, status: res.status, json: JSON.parse(txt)};
    } catch (e) {
      return {ok: res.ok, status: res.status, text: txt};
    }
  }

  function persistThemePreference(theme) {
    var normalized = String(theme || '').trim();
    if (!normalized) {
      return;
    }
    try {
      window.localStorage.setItem('theme', normalized);
    } catch (error) {}
    try {
      document.cookie = 'pcs_theme=' + encodeURIComponent(normalized) + '; Path=/; Max-Age=31536000; SameSite=Lax';
    } catch (error) {}
  }
  function openForgotPasswordView(event) {
    if (event) {
      event.preventDefault();
    }
    prefillEmailFields(emailInput && emailInput.value);
    showAuthView('forgot');
  }

  if (rememberAdminEmailCheckbox) {
    rememberAdminEmailCheckbox.addEventListener('change', function () {
      syncRememberedEmailWithUI();
    });
  }

  if (emailInput) {
    emailInput.addEventListener('input', function () {
      if (rememberAdminEmailCheckbox && rememberAdminEmailCheckbox.checked) {
        syncRememberedEmailWithUI();
      }
    });
  }

  if (emailLoginForm) {
    emailLoginForm.addEventListener('submit', async function (event) {
      event.preventDefault();
      var email = normalizeEmail(emailInput && emailInput.value);
      var password = passInput && passInput.value ? passInput.value : '';
      var otpCode = isAdmin2FALoginEnabled() && otpInput && otpInput.value ? otpInput.value.replace(/\D/g, '').slice(0, 6) : '';
      if (!email || !password) {
        showMsg(loginMessageDiv, 'Debes ingresar correo y contraseña.', true);
        return;
      }

      if (rememberAdminEmailCheckbox && rememberAdminEmailCheckbox.checked) {
        setStoredRememberedEmail(email);
      } else {
        clearStoredRememberedEmail();
      }

      setButtonBusy(loginBtn, 'Ingresando...', true);
      showMsg(loginMessageDiv, '', false);
      try {
        var loginToken = await ensureRecaptcha('login', loginMessageDiv);
        if (loginToken === null) {
          return;
        }
        var response = await postJson('/super/api/administradores/login', {email: email, password: password, otp_code: otpCode, recaptcha_token: loginToken});
        if (response.ok && response.json && response.json.redirect_url) {
          persistThemePreference(response.json.apariencia);
          var sharedInvitationToken = getSharedInvitationTokenFromQuery();
          if (sharedInvitationToken) {
            window.location.href = '/seleccionar_empresa.html?shared_invitation_token=' + encodeURIComponent(sharedInvitationToken);
            return;
          }
          window.location.href = response.json.redirect_url;
          return;
        }
        if (response.json && response.json.password_setup_required) {
          showMsg(loginMessageDiv, getResponseMessage(response, 'Tu cuenta todavía no tiene una contraseña activa.'), true);
          return;
        }
        if (response.json && response.json.two_factor_required) {
          showMsg(loginMessageDiv, getResponseMessage(response, 'Ingresa el codigo 2FA de tu aplicacion autenticadora.'), true);
          syncAdmin2FAFieldVisibility();
          if (otpInput) otpInput.focus();
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
        if (recaptchaManagers.login) {
          recaptchaManagers.login.reset();
        }
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
      resetLinkToken = '';
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
        var forgotToken = await ensureRecaptcha('forgot', forgotMessageDiv);
        if (forgotToken === null) {
          return;
        }
        var response = await postJson('/super/api/administradores/solicitar_recuperacion', {email: email, recaptcha_token: forgotToken});
        if (!response.ok) {
          showMsg(forgotMessageDiv, getResponseMessage(response, 'No se pudo iniciar la recuperación de contraseña.'), true);
          return;
        }
        prefillEmailFields(email);
        showMsg(forgotMessageDiv, getResponseMessage(response, 'Si la cuenta existe y ya fue confirmada, enviaremos instrucciones para restablecer la contraseña.'), false);
      } catch (error) {
        showMsg(forgotMessageDiv, error && error.message ? error.message : 'No se pudo solicitar la recuperación de contraseña.', true);
      } finally {
        if (recaptchaManagers.forgot) {
          recaptchaManagers.forgot.reset();
        }
        setButtonBusy(forgotBtn, 'Enviando...', false);
      }
    });
  }

  if (resetPasswordForm) {
    resetPasswordForm.addEventListener('submit', async function (event) {
      event.preventDefault();
      var email = normalizeEmail(resetEmailInput && resetEmailInput.value);
      var token = normalizeEmail(resetLinkToken).replace(/\s+/g, '');
      var password = resetPasswordInput && resetPasswordInput.value ? resetPasswordInput.value : '';
      var passwordConfirm = resetPasswordConfirmInput && resetPasswordConfirmInput.value ? resetPasswordConfirmInput.value : '';
      if (!email || !password || !passwordConfirm) {
        showMsg(resetMessageDiv, 'Debes completar el correo y las dos contraseñas.', true);
        return;
      }
      if (!token) {
        showMsg(resetMessageDiv, 'El enlace de recuperación no es válido o ya no contiene el código de seguridad.', true);
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
        var resetTokenValue = await ensureRecaptcha('reset', resetMessageDiv);
        if (resetTokenValue === null) {
          return;
        }
        var response = await postJson('/super/api/administradores/restablecer_password', {
          email: email,
          token: token,
          password: password,
          recaptcha_token: resetTokenValue
        });
        if (response.ok && response.json && response.json.redirect_url) {
          window.location.href = response.json.redirect_url;
          return;
        }
        showMsg(resetMessageDiv, getResponseMessage(response, 'No se pudo restablecer la contraseña.'), true);
      } catch (error) {
        showMsg(resetMessageDiv, error && error.message ? error.message : 'No se pudo restablecer la contraseña.', true);
      } finally {
        if (recaptchaManagers.reset) {
          recaptchaManagers.reset.reset();
        }
        setButtonBusy(resetBtn, 'Restableciendo...', false);
      }
    });
  }

  initPasswordVisibilityToggles();
  syncAdmin2FAFieldVisibility();

  (function handleRecoveryFromQuery() {
    try {
      hydrateRememberedEmail();
      var params = new URLSearchParams(window.location.search);
      var queryEmail = normalizeEmail(params.get('email'));
      var token = normalizeEmail(params.get('token_recuperacion') || params.get('token')).replace(/\s+/g, '');
      var view = normalizeEmail(params.get('view') || params.get('modo')).toLowerCase();

      if (queryEmail) {
        prefillEmailFields(queryEmail);
      }
      resetLinkToken = token;

      if (queryEmail && token) {
        showAuthView('reset');
        showMsg(resetMessageDiv, 'Define tu nueva contraseña para completar el restablecimiento.', false);
        return;
      }
      if (view === 'reset') {
        showAuthView('reset');
        showMsg(resetMessageDiv, 'Abre el enlace de recuperación enviado al correo para poder definir una nueva contraseña.', true);
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
