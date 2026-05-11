(function () {
  var statusEl = document.getElementById("totpStatus");
  var setupBtn = document.getElementById("setupTotpBtn");
  var disableBtn = document.getElementById("disableTotpBtn");
  var setupCard = document.getElementById("totpSetupCard");
  var secretInput = document.getElementById("totpSecret");
  var uriEl = document.getElementById("totpProvisioningUri");
  var codeInput = document.getElementById("totpCode");
  var confirmBtn = document.getElementById("confirmTotpBtn");
  var msgEl = document.getElementById("totpMessage");

  function setMessage(text, isError) {
    if (!msgEl) return;
    msgEl.textContent = text || "";
    msgEl.classList.toggle("form-error", !!isError);
  }

  async function request2FA(method, payload) {
    var response = await fetch("/super/api/administradores/2fa", {
      method: method || "GET",
      headers: {"Content-Type": "application/json"},
      credentials: "same-origin",
      body: payload ? JSON.stringify(payload) : undefined
    });
    var data = null;
    try { data = await response.json(); } catch (e) { data = {}; }
    if (!response.ok) {
      throw new Error(data && data.message ? data.message : "No se pudo completar la operación 2FA.");
    }
    return data;
  }

  async function loadStatus() {
    try {
      var data = await request2FA("GET");
      statusEl.textContent = data.enabled
        ? "Activo. El login del super administrador exige código 2FA."
        : "Inactivo. Genera un secreto y confirma un código para activar la protección.";
      if (disableBtn) disableBtn.disabled = !data.enabled;
    } catch (error) {
      statusEl.textContent = error.message || "No se pudo cargar el estado 2FA.";
    }
  }

  if (setupBtn) {
    setupBtn.addEventListener("click", async function () {
      setMessage("", false);
      setupBtn.disabled = true;
      try {
        var data = await request2FA("POST", {action: "setup"});
        if (setupCard) setupCard.classList.remove("is-hidden");
        if (secretInput) secretInput.value = data.secret || "";
        if (uriEl) uriEl.textContent = data.provisioning_uri || "";
        if (codeInput) codeInput.focus();
      } catch (error) {
        setMessage(error.message, true);
      } finally {
        setupBtn.disabled = false;
      }
    });
  }

  if (confirmBtn) {
    confirmBtn.addEventListener("click", async function () {
      var code = codeInput && codeInput.value ? codeInput.value.replace(/\D/g, "").slice(0, 6) : "";
      if (code.length !== 6) {
        setMessage("Ingresa el código de 6 dígitos.", true);
        return;
      }
      confirmBtn.disabled = true;
      try {
        await request2FA("POST", {action: "confirm", code: code});
        setMessage("2FA activado correctamente.", false);
        await loadStatus();
      } catch (error) {
        setMessage(error.message, true);
      } finally {
        confirmBtn.disabled = false;
      }
    });
  }

  if (disableBtn) {
    disableBtn.addEventListener("click", async function () {
      var code = window.prompt("Código 2FA para desactivar");
      if (code === null) return;
      disableBtn.disabled = true;
      try {
        await request2FA("POST", {action: "disable", code: String(code || "")});
        setMessage("2FA desactivado.", false);
        await loadStatus();
      } catch (error) {
        setMessage(error.message, true);
      } finally {
        disableBtn.disabled = false;
      }
    });
  }

  loadStatus();
})();
