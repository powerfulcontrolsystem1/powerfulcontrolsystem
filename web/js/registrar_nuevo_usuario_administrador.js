(function () {
  var form = document.getElementById('adminRegisterForm');
  if (!form) {
    return;
  }

  var emailInput = document.getElementById('registerEmail');
  var nameInput = document.getElementById('registerName');
  var phoneInput = document.getElementById('registerPhone');
  var countryInput = document.getElementById('registerCountry');
  var cityInput = document.getElementById('registerCity');
  var passwordInput = document.getElementById('registerPassword');
  var passwordConfirmInput = document.getElementById('registerPasswordConfirm');
  var submitButton = document.getElementById('adminRegisterBtn');
  var messageBox = document.getElementById('adminRegisterMessage');
  var recaptchaManager = window.PCSRecaptcha ? window.PCSRecaptcha.createManager({ containerId: 'adminRegisterRecaptcha', action: 'admin_register' }) : null;

  var countries = [
    { code: 'AR', name: 'Argentina' },
    { code: 'BO', name: 'Bolivia' },
    { code: 'BR', name: 'Brasil' },
    { code: 'CA', name: 'Canada' },
    { code: 'CL', name: 'Chile' },
    { code: 'CO', name: 'Colombia' },
    { code: 'CR', name: 'Costa Rica' },
    { code: 'CU', name: 'Cuba' },
    { code: 'DO', name: 'Republica Dominicana' },
    { code: 'EC', name: 'Ecuador' },
    { code: 'SV', name: 'El Salvador' },
    { code: 'ES', name: 'Espana' },
    { code: 'US', name: 'Estados Unidos' },
    { code: 'GT', name: 'Guatemala' },
    { code: 'HN', name: 'Honduras' },
    { code: 'MX', name: 'Mexico' },
    { code: 'NI', name: 'Nicaragua' },
    { code: 'PA', name: 'Panama' },
    { code: 'PY', name: 'Paraguay' },
    { code: 'PE', name: 'Peru' },
    { code: 'PR', name: 'Puerto Rico' },
    { code: 'UY', name: 'Uruguay' },
    { code: 'VE', name: 'Venezuela' },
    { code: 'DE', name: 'Alemania' },
    { code: 'FR', name: 'Francia' },
    { code: 'GB', name: 'Reino Unido' },
    { code: 'IT', name: 'Italia' },
    { code: 'PT', name: 'Portugal' },
    { code: 'NL', name: 'Paises Bajos' },
    { code: 'JP', name: 'Japon' },
    { code: 'AU', name: 'Australia' }
  ];
  var timezoneCountryMap = {
    'America/Argentina/Buenos_Aires': 'AR',
    'America/Bogota': 'CO',
    'America/Caracas': 'VE',
    'America/Costa_Rica': 'CR',
    'America/El_Salvador': 'SV',
    'America/Guatemala': 'GT',
    'America/Guayaquil': 'EC',
    'America/Havana': 'CU',
    'America/Lima': 'PE',
    'America/Managua': 'NI',
    'America/Mexico_City': 'MX',
    'America/Montevideo': 'UY',
    'America/Panama': 'PA',
    'America/Port-au-Prince': 'DO',
    'America/Puerto_Rico': 'PR',
    'America/Santiago': 'CL',
    'America/Santo_Domingo': 'DO',
    'America/Sao_Paulo': 'BR',
    'America/Tegucigalpa': 'HN',
    'America/Asuncion': 'PY',
    'Europe/Madrid': 'ES',
    'Europe/Lisbon': 'PT',
    'Europe/London': 'GB',
    'Europe/Paris': 'FR',
    'Europe/Berlin': 'DE',
    'Europe/Rome': 'IT',
    'Europe/Amsterdam': 'NL',
    'Asia/Tokyo': 'JP',
    'Australia/Sydney': 'AU'
  };

  function normalize(value) {
    return String(value || '').trim();
  }

  function hasCountry(code) {
    for (var index = 0; index < countries.length; index += 1) {
      if (countries[index].code === code) {
        return true;
      }
    }
    return false;
  }

  function populateCountryOptions() {
    if (!countryInput) {
      return;
    }
    var options = [];
    for (var index = 0; index < countries.length; index += 1) {
      var item = countries[index];
      options.push('<option value="' + item.code + '">' + item.name + '</option>');
    }
    countryInput.innerHTML = options.join('');
  }

  function detectCountryCode() {
    var languages = [];
    if (navigator.languages && navigator.languages.length) {
      languages = navigator.languages.slice(0, 5);
    } else if (navigator.language) {
      languages = [navigator.language];
    }
    for (var index = 0; index < languages.length; index += 1) {
      var raw = String(languages[index] || '');
      var match = raw.match(/[-_]([A-Za-z]{2})$/);
      if (match) {
        var candidate = match[1].toUpperCase();
        if (hasCountry(candidate)) {
          return candidate;
        }
      }
    }
    try {
      var timezone = Intl.DateTimeFormat().resolvedOptions().timeZone || '';
      if (timezoneCountryMap[timezone]) {
        return timezoneCountryMap[timezone];
      }
    } catch (error) {
      // Sin soporte; usar default.
    }
    return 'CO';
  }

  function getSelectedCountryName() {
    if (!countryInput) {
      return '';
    }
    var selectedIndex = countryInput.selectedIndex;
    if (selectedIndex >= 0 && countryInput.options[selectedIndex]) {
      return normalize(countryInput.options[selectedIndex].textContent);
    }
    return normalize(countryInput.value);
  }

  populateCountryOptions();
  if (countryInput) {
    countryInput.value = detectCountryCode();
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

  async function ensureRecaptcha() {
    if (!recaptchaManager) {
      return '';
    }
    var result = await recaptchaManager.ensureToken();
    if (!result.ok) {
      showMessage(result.message || 'Completa la verificación de seguridad para continuar.', true);
      return null;
    }
    return result.token || '';
  }

  if (recaptchaManager) {
    recaptchaManager.init().catch(function () {});
  }

  form.addEventListener('submit', async function (event) {
    event.preventDefault();

    var email = normalize(emailInput && emailInput.value);
    var name = normalize(nameInput && nameInput.value);
    var telefono = normalize(phoneInput && phoneInput.value);
    var pais = getSelectedCountryName();
    var ciudad = normalize(cityInput && cityInput.value);
    var password = passwordInput && passwordInput.value ? passwordInput.value : '';
    var passwordConfirm = passwordConfirmInput && passwordConfirmInput.value ? passwordConfirmInput.value : '';

    if (!email || !name || !telefono || !pais || !ciudad || !password || !passwordConfirm) {
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
      var registerToken = await ensureRecaptcha();
      if (registerToken === null) {
        return;
      }
      var response = await postJson('/super/api/administradores/register', {
        email: email,
        name: name,
        telefono: telefono,
        pais: pais,
        ciudad: ciudad,
		password: password,
		recaptcha_token: registerToken
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
      if (recaptchaManager) {
        recaptchaManager.reset();
      }
      setBusy(false);
    }
  });
})();