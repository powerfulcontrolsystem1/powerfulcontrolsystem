(function () {
  var form = document.getElementById('adminRegisterForm');
  if (!form) {
    return;
  }

  var emailInput = document.getElementById('registerEmail');
  var nameInput = document.getElementById('registerName');
  var phoneInput = document.getElementById('registerPhone');
  var passwordInput = document.getElementById('registerPassword');
  var passwordConfirmInput = document.getElementById('registerPasswordConfirm');
  var submitButton = document.getElementById('adminRegisterBtn');
  var messageBox = document.getElementById('adminRegisterMessage');

  function normalize(value) {
    return String(value || '').trim();
  }

  function showMessage(text, isError) {
    if (!messageBox) {
      return;
    }
    messageBox.style.display = text ? 'block' : 'none';
    messageBox.textContent = text || '';
    messageBox.style.color = isError ? 'crimson' : 'green';
  }

  function setBusy(isBusy) {
    if (!submitButton) {
      return;
    }
    if (!submitButton.dataset.defaultText) {
      submitButton.dataset.defaultText = submitButton.textContent;
    }
    submitButton.disabled = !!isBusy;
    submitButton.textContent = isBusy ? 'Registrando...' : submitButton.dataset.defaultText;
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
    var response = await fetch(url, {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify(body),
      credentials: 'same-origin'
    });
    var text = await response.text();
    try {
      return {ok: response.ok, status: response.status, json: JSON.parse(text)};
    } catch (error) {
      return {ok: response.ok, status: response.status, text: text};
    }
  }

  form.addEventListener('submit', async function (event) {
    event.preventDefault();

    var email = normalize(emailInput && emailInput.value);
    var name = normalize(nameInput && nameInput.value);
    var telefono = normalize(phoneInput && phoneInput.value);
    var password = passwordInput && passwordInput.value ? passwordInput.value : '';
    var passwordConfirm = passwordConfirmInput && passwordConfirmInput.value ? passwordConfirmInput.value : '';

    if (!email || !name || !telefono || !password || !passwordConfirm) {
      showMessage('Debes completar todos los campos del registro.', true);
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
    showMessage('', false);
    try {
      var response = await postJson('/super/api/administradores/register', {
        email: email,
        name: name,
        telefono: telefono,
        password: password
      });
      if (!response.ok) {
        showMessage(getResponseMessage(response, 'No se pudo completar el registro.'), true);
        return;
      }
      showMessage(getResponseMessage(response, 'Registro exitoso. Revisa tu correo para confirmar la cuenta.'), false);
      window.setTimeout(function () {
        window.location.href = '/login.html?email=' + encodeURIComponent(email);
      }, 1800);
    } catch (error) {
      showMessage(error && error.message ? error.message : 'No se pudo completar el registro.', true);
    } finally {
      setBusy(false);
    }
  });
})();