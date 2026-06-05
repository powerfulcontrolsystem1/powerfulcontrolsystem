(function () {
  var installPromptEvent = window.__pcsInstallPromptEvent || null;
  var installButton = document.getElementById("installPwaBtn");
  var messageBox = document.getElementById("installPwaMessage");
  var installButtonLabel = installButton ? installButton.querySelector(".pwa-install-label") : null;
  var serviceWorkerReady = false;
  var serviceWorkerSupported = "serviceWorker" in navigator;
  var installAttemptInProgress = false;
  var promptReadyAt = installPromptEvent ? Date.now() : 0;
  var controllerReloadKey = "pcs_pwa_controller_reload_at";

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
    installButton.setAttribute("aria-busy", installAttemptInProgress ? "true" : "false");
    setInstallButtonLabel("Instalar app");
    exposeState();
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
    promptReadyAt = Date.now();
    syncInstallButton();
    setMessage("", false);
  }

  function registerServiceWorker() {
    if (!serviceWorkerSupported) {
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

  function exposeState() {
    window.__pcsPwaInstallState = {
      serviceWorkerSupported: serviceWorkerSupported,
      serviceWorkerReady: serviceWorkerReady,
      hasPrompt: !!installPromptEvent,
      promptReadyAt: promptReadyAt,
      standalone: isStandalone(),
      inProgress: installAttemptInProgress
    };
  }

  function currentURLWithPwaReadyFlag() {
    try {
      var url = new URL(window.location.href);
      url.searchParams.set("pwa_ready", String(Date.now()));
      return url.toString();
    } catch (e) {
      return window.location.href;
    }
  }

  function shouldReloadForServiceWorkerController() {
    if (!serviceWorkerSupported || !serviceWorkerReady || installPromptEvent) {
      return false;
    }
    if (hasTypedLoginData()) {
      return false;
    }
    if (navigator.serviceWorker && navigator.serviceWorker.controller) {
      return false;
    }
    try {
      var last = Number(sessionStorage.getItem(controllerReloadKey) || "0");
      if (last && Date.now() - last < 30000) {
        return false;
      }
      sessionStorage.setItem(controllerReloadKey, String(Date.now()));
    } catch (e) {
      return false;
    }
    return true;
  }

  function hasTypedLoginData() {
    var fields = document.querySelectorAll("input[type='email'], input[type='password'], input[type='text']");
    for (var i = 0; i < fields.length; i += 1) {
      if (String(fields[i].value || "").trim()) {
        return true;
      }
    }
    return false;
  }

  function installMockPromptIfRequested() {
    var params = null;
    try {
      params = new URLSearchParams(window.location.search || "");
    } catch (e) {
      params = null;
    }
    if (!params || !(params.has("qa_pwa") || params.has("qa_pwa_codex"))) {
      return;
    }
    if (installPromptEvent || isStandalone()) {
      return;
    }
    var resolveChoice = null;
    var userChoice = new Promise(function (resolve) {
      resolveChoice = resolve;
    });
    var mockEvent = {
      preventDefault: function () {},
      prompt: function () {
        window.setTimeout(function () {
          if (resolveChoice) {
            resolveChoice({ outcome: "accepted", platform: "qa" });
          }
        }, 250);
        return Promise.resolve();
      },
      userChoice: userChoice
    };
    rememberInstallPrompt(mockEvent);
  }

  function openInstallPrompt(promptEvent) {
    if (!promptEvent) {
      return false;
    }
    var promptResult = null;
    setMessage("Abriendo instalacion de la app...", false);
    try {
      promptResult = promptEvent.prompt();
    } catch (error) {
      setMessage("Chrome no permitio abrir la instalacion todavia. Espera unos segundos y presiona Instalar app nuevamente.", true);
      installPromptEvent = null;
      window.__pcsInstallPromptEvent = null;
      syncInstallButton();
      return true;
    }
    Promise.resolve(promptResult).catch(function () {
      setMessage("Chrome no permitio abrir la instalacion todavia. Espera unos segundos y presiona Instalar app nuevamente.", true);
      installPromptEvent = null;
      window.__pcsInstallPromptEvent = null;
      syncInstallButton();
    });
    if (!promptEvent.userChoice) {
      setMessage("Si aparece la ventana de instalacion, confirma para crear el acceso de la app.", false);
      installPromptEvent = null;
      window.__pcsInstallPromptEvent = null;
      syncInstallButton();
      return true;
    }
    promptEvent.userChoice.then(function (choice) {
      if (choice && choice.outcome === "accepted") {
        setMessage("Instalacion iniciada. Chrome creara el acceso de la aplicacion.", false);
      } else {
        setMessage("Instalacion cancelada.", true);
      }
      installPromptEvent = null;
      window.__pcsInstallPromptEvent = null;
      syncInstallButton();
    }).catch(function () {
      setMessage("Chrome no confirmo la instalacion. Presiona Instalar app nuevamente.", true);
      installPromptEvent = null;
      window.__pcsInstallPromptEvent = null;
      syncInstallButton();
    });
    return true;
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
    window.__pcsInstallPromptEvent = null;
    syncInstallButton();
    setMessage("App instalada correctamente.", false);
  });

  window.addEventListener("load", function () {
    installMockPromptIfRequested();
  });

  if (installButton) {
    installButton.addEventListener("click", function () {
      if (isStandalone()) {
        setMessage("La app ya esta instalada.", false);
        return;
      }
      if (installPromptEvent) {
        openInstallPrompt(installPromptEvent);
        return;
      }
      if (isiOS()) {
        setMessage("En iPhone o iPad, usa Compartir y luego Agregar a pantalla de inicio.", false);
        return;
      }
      if (!serviceWorkerSupported) {
        setMessage("Este navegador no permite instalar la app desde el boton. Abrela en Chrome o Edge y presiona Instalar app.", true);
        return;
      }
      installAttemptInProgress = true;
      syncInstallButton();
      setMessage("Preparando instalacion de la app. Espera unos segundos...", false);
      registerServiceWorker().then(function () {
        installAttemptInProgress = false;
        if (installPromptEvent) {
          syncInstallButton();
          setMessage("Listo. Presiona Instalar app nuevamente para crear el acceso.", false);
          return;
        }
        if (shouldReloadForServiceWorkerController()) {
          setMessage("Activando instalador de la app. La pagina se recargara una vez.", false);
          window.setTimeout(function () {
            window.location.replace(currentURLWithPwaReadyFlag());
          }, 650);
          return;
        }
        syncInstallButton();
        setMessage("La app esta preparada. Si Chrome no abre la ventana, espera unos segundos y presiona Instalar app otra vez.", false);
      });
    });
  }

  syncInstallButton();
  installMockPromptIfRequested();
})();
