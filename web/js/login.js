(function () {
  var rememberCheckbox = document.getElementById("rememberAccount");
  var googleBtn = document.querySelector(".google-btn");
  var hintBlock = document.getElementById("rememberedAccountHint");
  var hintEmail = document.getElementById("rememberedAccountEmail");
  var clearRememberLink = document.getElementById("clearRememberedAccount");

  var KEY_REMEMBER_FLAG = "rememberAccount";
  var KEY_REMEMBER_EMAIL = "rememberedEmail";

  function safeGetItem(key) {
    try { return localStorage.getItem(key) || ""; } catch (e) { return ""; }
  }

  function safeSetItem(key, value) {
    try { localStorage.setItem(key, value); } catch (e) {}
  }

  function setRememberedEmail(email) {
    var normalized = String(email || "").trim();
    if (!normalized) {
      return;
    }
    safeSetItem(KEY_REMEMBER_EMAIL, normalized);
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
    return String(safeGetItem(KEY_REMEMBER_EMAIL) || "").trim();
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
        setRememberedEmail(email);
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
})();
