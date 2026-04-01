(function () {
  var cb = document.getElementById("rememberAccount");
  try {
    if (localStorage.getItem("rememberAccount") === "1") cb.checked = true;
  } catch (e) {}

  // Auto-login if remember flag is set and no explicit opt-out in URL (?no_auto=1)
  try {
    var params = new URLSearchParams(window.location.search);
    var noAuto = params.get("no_auto");
    if (localStorage.getItem("rememberAccount") === "1" && noAuto !== "1") {
      // Si tenemos un email guardado, pasar login_hint para que Google lo seleccione automaticamente.
      var remembered = null;
      try {
        remembered = localStorage.getItem("rememberedEmail");
      } catch (e) {}
      var target = "/auth/google/login";
      if (remembered) {
        target += "?login_hint=" + encodeURIComponent(remembered);
      }
      setTimeout(function () {
        window.location.href = target;
      }, 200);
      return;
    }
  } catch (e) {}

  var googleBtn = document.querySelector(".google-btn");
  if (googleBtn) {
    googleBtn.addEventListener("click", function () {
      try {
        if (cb && cb.checked) {
          localStorage.setItem("rememberAccount", "1");
        } else {
          localStorage.removeItem("rememberAccount");
          localStorage.removeItem("rememberedEmail");
        }
      } catch (e) {}
    });
  }

  var logoutLink = document.querySelector('.fm-item[href="/auth/logout"]');
  if (logoutLink) {
    logoutLink.addEventListener("click", function () {
      try {
        localStorage.removeItem("rememberAccount");
        localStorage.removeItem("rememberedEmail");
      } catch (e) {}
    });
  }
})();
