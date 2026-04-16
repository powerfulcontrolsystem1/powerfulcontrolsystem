(function () {
  var form = document.getElementById('googlePasswordSetupForm');
  var intro = document.getElementById('googlePasswordSetupIntro');
  var message = document.getElementById('googlePasswordSetupMessage');
  var emailInput = document.getElementById('googleAccountEmail');
  var passwordInput = document.getElementById('googlePassword');
  var confirmInput = document.getElementById('googlePasswordConfirm');
  var submitBtn = document.getElementById('googlePasswordSetupBtn');
  var skipLink = document.getElementById('googlePasswordSkipLink');

  function showMessage(text, isError) {
    if (!message) {
      return;
    }
    message.style.display = text ? 'block' : 'none';
    message.textContent = text || '';
    message.style.color = isError ? 'crimson' : 'green';
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
        intro.textContent = 'Tu cuenta de Google ya esta activa. Registra una contrasena para usar tambien el acceso por correo.';
      }
      if (form) {
        form.style.display = 'block';
      }
    } catch (error) {
      if (intro) {
        intro.textContent = 'No se pudo validar la sesion actual.';
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
        showMessage('Debes escribir y confirmar la nueva contrasena.', true);
        return;
      }
      if (password.length < 8) {
        showMessage('La contrasena debe tener minimo 8 caracteres.', true);
        return;
      }
      if (password !== passwordConfirm) {
        showMessage('Las contrasenas no coinciden.', true);
        return;
      }

      setBusy(true);
      try {
        var result = await postJson('/api/account/set_google_password', {
          password: password,
          password_confirm: passwordConfirm
        });
        if (!result.ok) {
          var text = result.json && (result.json.message || result.json.error) ? (result.json.message || result.json.error) : (result.text || 'No se pudo guardar la contrasena.');
          showMessage(String(text), true);
          return;
        }
        var redirectUrl = result.json && result.json.redirect_url ? String(result.json.redirect_url) : '/seleccionar_empresa.html';
        showMessage('Contrasena registrada correctamente. Redirigiendo...', false);
        window.location.href = redirectUrl;
      } catch (error) {
        showMessage(error && error.message ? error.message : 'No se pudo guardar la contrasena.', true);
      } finally {
        setBusy(false);
      }
    });
  }

  loadAccount();
})();