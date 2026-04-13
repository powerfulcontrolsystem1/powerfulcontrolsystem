(function () {
  var cb = document.getElementById("rememberAccount");
  var googleBtn = document.querySelector(".google-btn");

  function clearRememberState() {
    try {
      localStorage.removeItem("rememberAccount");
      localStorage.removeItem("rememberedEmail");
    } catch (e) {}
  }

  try {
    if (cb && localStorage.getItem("rememberAccount") === "1") {
      cb.checked = true;
    }

    // Si el usuario previamente indicó recordar cuenta y existe un email recordado,
    // incluimos el login_hint y opcionalmente redirigimos automáticamente a Google.
    try {
      var rememberedEmail = localStorage.getItem("rememberedEmail") || "";
      if (rememberedEmail && localStorage.getItem("rememberAccount") === "1") {
        var loginUrl = "/auth/google/login?login_hint=" + encodeURIComponent(rememberedEmail);
        if (googleBtn) {
          googleBtn.setAttribute('href', loginUrl);
          // Navegación automática para agilizar acceso. Añadir ?no_auto_login=1 a la URL
          // para evitar el auto-redirect en pruebas.
          if (!window.location.search.includes('no_auto_login=1')) {
            window.location.href = loginUrl;
          }
        }
      }
    } catch (e) {}
  } catch (e) {}

  if (googleBtn) {
    googleBtn.addEventListener("click", function (ev) {
      ev.preventDefault();
      try {
        if (cb && cb.checked) {
          localStorage.setItem("rememberAccount", "1");
        } else {
          clearRememberState();
        }
      } catch (e) {}

      // El contrato/reCAPTCHA se resuelve después del callback OAuth en /accept.html.
      // Si existe un email recordado, anexarlo como login_hint para facilitar la selección de cuenta.
      var target = "/auth/google/login";
      try {
        var em = localStorage.getItem('rememberedEmail');
        if (em) {
          target += '?login_hint=' + encodeURIComponent(em);
        }
      } catch (e) {}
      window.location.href = target;
    });
  }

  var logoutLink = document.querySelector('.fm-item[href="/auth/logout"]');
  if (logoutLink) {
    logoutLink.addEventListener("click", function () {
      clearRememberState();
    });
  }
})();
