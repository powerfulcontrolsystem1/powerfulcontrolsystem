(function () {
  var installPromptEvent = null;
  var installButton = document.getElementById("installPwaBtn");
  var messageBox = document.getElementById("installPwaMessage");
  var installButtonLabel = installButton ? installButton.querySelector(".pwa-install-label") : null;

  function setMessage(text, isError) {
    if (!messageBox) {
      return;
    }
    messageBox.textContent = text || "";
    messageBox.classList.toggle("error", !!text && !!isError);
    messageBox.classList.toggle("success", !!text && !isError);
  }

  function isStandalone() {
    return window.matchMedia && window.matchMedia("(display-mode: standalone)").matches || window.navigator.standalone === true;
  }

  function isiOS() {
    return /iphone|ipad|ipod/i.test(window.navigator.userAgent || "");
  }

  function syncInstallButton() {
    if (!installButton) {
      return;
    }
    if (isStandalone()) {
      installButton.disabled = true;
      setInstallButtonLabel("App instalada");
      setMessage("", false);
      return;
    }
    installButton.disabled = false;
    setInstallButtonLabel("Instalar app");
  }

  function setInstallButtonLabel(text) {
    if (installButtonLabel) {
      installButtonLabel.textContent = text;
      return;
    }
    installButton.textContent = text;
  }

  if ("serviceWorker" in navigator) {
    window.addEventListener("load", function () {
      navigator.serviceWorker.register("/sw.js", { scope: "/" }).then(function (registration) {
        if (registration && typeof registration.update === "function") {
          registration.update().catch(function () {});
        }
      }).catch(function () {});
    });
  }

  window.addEventListener("beforeinstallprompt", function (event) {
    event.preventDefault();
    installPromptEvent = event;
    syncInstallButton();
    setMessage("", false);
  });

  window.addEventListener("appinstalled", function () {
    installPromptEvent = null;
    syncInstallButton();
    setMessage("App instalada correctamente.", false);
  });

  if (installButton) {
    installButton.addEventListener("click", function () {
      if (isStandalone()) {
        setMessage("La app ya esta instalada.", false);
        return;
      }
      if (installPromptEvent) {
        installPromptEvent.prompt();
        installPromptEvent.userChoice.then(function (choice) {
          if (choice && choice.outcome === "accepted") {
            setMessage("Instalacion iniciada.", false);
          } else {
            setMessage("Instalacion cancelada.", true);
          }
          installPromptEvent = null;
          syncInstallButton();
        }).catch(function () {
          setMessage("No se pudo abrir la instalacion. Usa el menu del navegador para instalar la app.", true);
        });
        return;
      }
      if (isiOS()) {
        setMessage("En iPhone o iPad, usa Compartir y luego Agregar a pantalla de inicio.", false);
        return;
      }
      setMessage("Si no aparece la ventana, usa el menu del navegador y elige Instalar app o Agregar a pantalla de inicio.", false);
    });
  }

  syncInstallButton();
})();
