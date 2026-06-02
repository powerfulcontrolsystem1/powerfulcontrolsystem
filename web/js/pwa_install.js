(function () {
  var installPromptEvent = window.__pcsInstallPromptEvent || null;
  var installButton = document.getElementById("installPwaBtn");
  var messageBox = document.getElementById("installPwaMessage");
  var installButtonLabel = installButton ? installButton.querySelector(".pwa-install-label") : null;
  var serviceWorkerReady = false;

  function setMessage(text, isError) {
    if (!messageBox) {
      return;
    }
    messageBox.textContent = text || "";
    messageBox.classList.toggle("error", !!text && !!isError);
    messageBox.classList.toggle("success", !!text && !isError);
  }

  function isStandalone() {
    return (window.matchMedia && window.matchMedia("(display-mode: standalone)").matches) || window.navigator.standalone === true;
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
    installButton.classList.toggle("is-install-ready", !!installPromptEvent);
    setInstallButtonLabel("Instalar app");
  }

  function setInstallButtonLabel(text) {
    if (installButtonLabel) {
      installButtonLabel.textContent = text;
      return;
    }
    installButton.textContent = text;
  }

  function rememberInstallPrompt(event) {
    if (!event) {
      return;
    }
    if (typeof event.preventDefault === "function") {
      event.preventDefault();
    }
    installPromptEvent = event;
    window.__pcsInstallPromptEvent = event;
    syncInstallButton();
    setMessage("", false);
  }

  function registerServiceWorker() {
    if (!("serviceWorker" in navigator)) {
      return Promise.resolve(false);
    }
    return navigator.serviceWorker.register("/sw.js", { scope: "/" }).then(function (registration) {
      serviceWorkerReady = true;
      if (registration && typeof registration.update === "function") {
        registration.update().catch(function () {});
      }
      return navigator.serviceWorker.ready.catch(function () { return registration; });
    }).then(function () {
      serviceWorkerReady = true;
      syncInstallButton();
      return true;
    }).catch(function () {
      serviceWorkerReady = false;
      syncInstallButton();
      return false;
    });
  }

  registerServiceWorker();

  window.addEventListener("beforeinstallprompt", function (event) {
    rememberInstallPrompt(event);
  });

  window.addEventListener("pcs:beforeinstallprompt", function () {
    rememberInstallPrompt(window.__pcsInstallPromptEvent);
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
        var promptResult = null;
        setMessage("Abriendo instalacion de la app...", false);
        try {
          promptResult = installPromptEvent.prompt();
        } catch (error) {
          setMessage("No se pudo abrir la instalacion. Usa el menu del navegador para instalar la app.", true);
          installPromptEvent = null;
          syncInstallButton();
          return;
        }
        Promise.resolve(promptResult).catch(function () {
          setMessage("No se pudo abrir la instalacion. Usa el menu del navegador para instalar la app.", true);
          installPromptEvent = null;
          syncInstallButton();
        });
        if (!installPromptEvent || !installPromptEvent.userChoice) {
          setMessage("Si no aparece la ventana, usa el menu del navegador y elige Instalar app o Agregar a pantalla de inicio.", false);
          installPromptEvent = null;
          syncInstallButton();
          return;
        }
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
      setMessage(serviceWorkerReady ? "Chrome todavia no habilito la ventana automatica. Usa el icono de instalar de la barra del navegador o el menu y elige Instalar app." : "Preparando la app para instalacion. Espera unos segundos y vuelve a presionar Instalar app.", false);
      if (!serviceWorkerReady) {
        registerServiceWorker();
      }
    });
  }

  syncInstallButton();
})();
