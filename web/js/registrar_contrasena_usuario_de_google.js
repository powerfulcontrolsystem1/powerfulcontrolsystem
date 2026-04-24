(function () {
  var form = document.getElementById('googlePasswordSetupForm');
  var intro = document.getElementById('googlePasswordSetupIntro');
  var message = document.getElementById('googlePasswordSetupMessage');
  var emailInput = document.getElementById('googleAccountEmail');
  var passwordInput = document.getElementById('googlePassword');
  var confirmInput = document.getElementById('googlePasswordConfirm');
  var submitBtn = document.getElementById('googlePasswordSetupBtn');
  var skipLink = document.getElementById('googlePasswordSkipLink');

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

  function showMessage(text, isError) {
    if (!message) {
      return;
    }
    message.classList.toggle('is-hidden', !text);
    message.classList.toggle('is-visible', !!text);
    message.textContent = text || '';
    message.classList.toggle('error', !!text && !!isError);
    message.classList.toggle('success', !!text && !isError);
  }

  function setBusy(isBusy) {
    if (!submitBtn) {
      return;
    }
    if (!submitBtn.dataset.defaultText) {
      submitBtn.dataset.defaultText = submitBtn.textContent;
    }
    submitBtn.disabled = !!isBusy;
    submitBtn.textContent = isBusy ? 'Guardando...' : submitBtn.dataset.defaultText;
  }

  function resolveRedirect(account) {
    if (account && account.is_super) {
      return '/super_administrador.html';
    }
    if (account && account.admin && String(account.admin.role || '').trim().toLowerCase() === 'super_administrador') {
      return '/super_administrador.html';
    }
    return '/seleccionar_empresa.html';
  }

  async function loadAccount() {
    try {
      var response = await fetch('/api/account', { credentials: 'same-origin' });
      if (!response.ok) {
        window.location.href = '/login.html';
        return;
      }
      var account = await response.json();
      if (!account || !account.email) {
        window.location.href = '/login.html';
        return;
      }
      if (emailInput) {
        emailInput.value = account.email;
      }
      if (skipLink) {
        skipLink.href = resolveRedirect(account);
      }
      if (account.admin && Number(account.admin.password_set || 0) === 1 && String(account.admin.password_hash || '').trim()) {
        window.location.href = resolveRedirect(account);
        return;
      }
      if (intro) {
        intro.textContent = 'Tu cuenta de Google ya está activa. Registra una contraseña para usar también el acceso por correo.';
      }
      if (form) {
        form.classList.remove('is-hidden');
      }
    } catch (error) {
      if (intro) {
        intro.textContent = 'No se pudo validar la sesión actual.';
      }
      showMessage(error && error.message ? error.message : 'No se pudo validar la cuenta.', true);
    }
  }

  async function postJson(url, body) {
    var response = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
      credentials: 'same-origin'
    });
    var text = await response.text();
    try {
      return { ok: response.ok, status: response.status, json: JSON.parse(text) };
    } catch (error) {
      return { ok: response.ok, status: response.status, text: text };
    }
  }

  if (form) {
    form.addEventListener('submit', async function (event) {
      event.preventDefault();
      showMessage('', false);

      var password = passwordInput && passwordInput.value ? passwordInput.value : '';
      var passwordConfirm = confirmInput && confirmInput.value ? confirmInput.value : '';
      if (!password || !passwordConfirm) {
        showMessage('Debes escribir y confirmar la nueva contraseña.', true);
        return;
      }
      if (password.length < 8) {
        showMessage('La contraseña debe tener mínimo 8 caracteres.', true);
        return;
      }
      if (password !== passwordConfirm) {
        showMessage('Las contraseñas no coinciden.', true);
        return;
      }

      setBusy(true);
      try {
        var result = await postJson('/api/account/set_google_password', {
          password: password,
          password_confirm: passwordConfirm
        });
        if (!result.ok) {
          var text = result.json && (result.json.message || result.json.error) ? (result.json.message || result.json.error) : (result.text || 'No se pudo guardar la contraseña.');
          showMessage(String(text), true);
          return;
        }
        var redirectUrl = result.json && result.json.redirect_url ? String(result.json.redirect_url) : '/seleccionar_empresa.html';
        showMessage('Contraseña registrada correctamente. Redirigiendo...', false);
        window.location.href = redirectUrl;
      } catch (error) {
        showMessage(error && error.message ? error.message : 'No se pudo guardar la contraseña.', true);
      } finally {
        setBusy(false);
      }
    });
  }

  loadAccount();
  initPasswordVisibilityToggles();
})();