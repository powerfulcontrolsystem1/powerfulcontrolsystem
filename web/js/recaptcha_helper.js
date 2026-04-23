(function (global) {
  'use strict';

  var scriptPromise = null;
  var onloadCallbackName = '__pcsRecaptchaOnload';

  function resolveRecaptchaApi() {
    var g = global.grecaptcha;
    if (!g) return null;
    if (g.enterprise && typeof g.enterprise.render === 'function') {
      return g.enterprise;
    }
    if (typeof g.render === 'function') {
      return g;
    }
    return null;
  }

  function isEnabled() {
    if (typeof global.RECAPTCHA_REQUESTED_ENABLED !== 'undefined' && !global.RECAPTCHA_REQUESTED_ENABLED) {
      return false;
    }
    return !!global.RECAPTCHA_ENABLED && !!String(global.RECAPTCHA_SITE_KEY || '').trim();
  }

  function isBypassEnabled() {
	return !!global.RECAPTCHA_DEV_BYPASS;
  }

  function loadScript() {
    if (scriptPromise) {
      return scriptPromise;
    }
    var readyApi = resolveRecaptchaApi();
    if (readyApi) {
      scriptPromise = Promise.resolve(readyApi);
      return scriptPromise;
    }
    scriptPromise = new Promise(function (resolve, reject) {
      try {
        global[onloadCallbackName] = function () {
          var api = resolveRecaptchaApi();
          if (api) {
            resolve(api);
            return;
          }
          reject(new Error('Google reCAPTCHA cargó, pero no expuso la función render().'));
        };
      } catch (e) {}

      var script = document.createElement('script');
      script.src = 'https://www.google.com/recaptcha/api.js?onload=' + encodeURIComponent(onloadCallbackName) + '&render=explicit';
      script.async = true;
      script.defer = true;
      script.onload = function () {
        // Si por alguna razón el callback onload no se ejecutó, reintentar aquí.
        var api = resolveRecaptchaApi();
        if (api) {
          resolve(api);
          return;
        }
      };
      script.onerror = function () { reject(new Error('No se pudo cargar Google reCAPTCHA.')); };
      document.head.appendChild(script);
    });
    return scriptPromise;
  }

  function createManager(options) {
    var settings = options || {};
    var widgetId = null;
    var initializing = null;

    function getContainer() {
      if (!settings.containerId) {
        return null;
      }
      return document.getElementById(settings.containerId);
    }

    function setContainerVisible(isVisible) {
      var container = getContainer();
      if (!container) {
        return;
      }
      if (isVisible) {
        container.style.display = 'block';
        container.style.minHeight = '78px';
        container.style.overflow = 'visible';
      } else {
        container.style.display = 'none';
      }
    }

    function init() {
      if (!isEnabled() || isBypassEnabled()) {
        setContainerVisible(false);
        return Promise.resolve(false);
      }
      if (widgetId !== null) {
        setContainerVisible(true);
        return Promise.resolve(true);
      }
      if (initializing) {
        return initializing;
      }
      initializing = loadScript().then(function (grecaptcha) {
        var container = getContainer();
        if (!container) {
          throw new Error('No se encontró el contenedor de reCAPTCHA.');
        }
        setContainerVisible(true);
        // Evita fallos al reusar el mismo contenedor entre vistas.
        container.innerHTML = '';
        widgetId = grecaptcha.render(container, {
          sitekey: String(global.RECAPTCHA_SITE_KEY || '').trim(),
          theme: settings.theme || 'light'
        });
        if (!container.querySelector('iframe')) {
          throw new Error('No se pudo mostrar el widget de reCAPTCHA. Recarga la página e inténtalo de nuevo.');
        }
        initializing = null;
        return true;
      }).catch(function (error) {
        initializing = null;
        throw error;
      });
      return initializing;
    }

    function ensureToken() {
      if (!isEnabled() || isBypassEnabled()) {
        return Promise.resolve({ ok: true, token: '' });
      }
      return init().then(function () {
        var grecaptcha = resolveRecaptchaApi();
        if (!grecaptcha || widgetId === null) {
          return { ok: false, message: 'No se pudo inicializar la verificación de seguridad.' };
        }
        var token = String(grecaptcha.getResponse(widgetId) || '').trim();
        if (!token) {
          setContainerVisible(true);
          var container = getContainer();
          if (container && typeof container.scrollIntoView === 'function') {
            container.scrollIntoView({ behavior: 'smooth', block: 'center' });
          }
          return { ok: false, message: 'Completa el reCAPTCHA que aparece debajo del formulario para continuar.' };
        }
        return { ok: true, token: token };
      }).catch(function (error) {
        setContainerVisible(true);
        return {
          ok: false,
          message: (error && error.message) ? error.message : 'No se pudo cargar la verificación de seguridad.'
        };
      });
    }

    function reset() {
      var grecaptcha = resolveRecaptchaApi();
      if (!grecaptcha || widgetId === null) {
        return;
      }
      grecaptcha.reset(widgetId);
    }

    return {
      init: init,
      ensureToken: ensureToken,
      reset: reset,
      isEnabled: isEnabled
    };
  }

  global.PCSRecaptcha = {
    createManager: createManager,
    isEnabled: isEnabled,
    isBypassEnabled: isBypassEnabled
  };
})(window);