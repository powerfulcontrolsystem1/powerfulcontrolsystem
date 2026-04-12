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
      window.location.href = "/auth/google/login";
    });
  }

  var logoutLink = document.querySelector('.fm-item[href="/auth/logout"]');
  if (logoutLink) {
    logoutLink.addEventListener("click", function () {
      clearRememberState();
    });
  }
})();
