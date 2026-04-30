(function (global) {
  'use strict';

  var scriptPromise = null;
  var onloadCallbackName = '__pcsRecaptchaOnload';
  var scriptSelector = 'script[data-pcs-recaptcha-script="1"]';

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

  function normalizeProvider(raw) {
    var value = String(raw || '').trim().toLowerCase();
    if (!value) return 'google-recaptcha-v2';
    if (value.indexOf('enterprise') !== -1) return 'google-recaptcha-enterprise';
    if (value.indexOf('v3') !== -1) return 'google-recaptcha-v3';
    return 'google-recaptcha-v2';
  }

  function resolveRecaptchaApiV3() {
    var g = global.grecaptcha;
    if (!g) return null;
    if (typeof g.ready === 'function' && typeof g.execute === 'function') {
      return g;
    }
    if (g.enterprise && typeof g.enterprise.ready === 'function' && typeof g.enterprise.execute === 'function') {
      return g.enterprise;
    }
    return null;
  }

  function resolveLoadedApi() {
    var prov = normalizeProvider(global.RECAPTCHA_PROVIDER);
    if (prov === 'google-recaptcha-v3' || prov === 'google-recaptcha-enterprise') {
      var apiV3 = resolveRecaptchaApiV3();
      if (apiV3) {
        return apiV3;
      }
    }
    var apiWidget = resolveRecaptchaApi();
    if (apiWidget) {
      return apiWidget;
    }
    return resolveRecaptchaApiV3();
  }

  function loadScript() {
    if (scriptPromise) {
      return scriptPromise;
    }
    var providerEarly = normalizeProvider(global.RECAPTCHA_PROVIDER);
    if (providerEarly === 'google-recaptcha-v3' || providerEarly === 'google-recaptcha-enterprise') {
      var v3Cached = resolveRecaptchaApiV3();
      if (v3Cached) {
        scriptPromise = Promise.resolve(v3Cached);
        return scriptPromise;
      }
    } else {
      var readyApi = resolveRecaptchaApi();
      if (readyApi) {
        scriptPromise = Promise.resolve(readyApi);
        return scriptPromise;
      }
      var readyV3 = resolveRecaptchaApiV3();
      if (readyV3) {
        scriptPromise = Promise.resolve(readyV3);
        return scriptPromise;
      }
    }
    scriptPromise = new Promise(function (resolve, reject) {
      try {
        global[onloadCallbackName] = function () {
          var api = resolveLoadedApi();
          if (api) {
            resolve(api);
            return;
          }
          reject(new Error('Google reCAPTCHA cargó, pero no expuso las funciones esperadas (render o execute).'));
        };
      } catch (e) {}

      var script = document.createElement('script');
      script.setAttribute('data-pcs-recaptcha-script', '1');
      var provider = normalizeProvider(global.RECAPTCHA_PROVIDER);
      var base = (provider === 'google-recaptcha-enterprise')
        ? 'https://www.google.com/recaptcha/enterprise.js'
        : 'https://www.google.com/recaptcha/api.js';

      if (provider === 'google-recaptcha-v3' || provider === 'google-recaptcha-enterprise') {
        // v3: no widget. Se usa execute(sitekey, {action}) luego de ready().
        script.src = base + '?render=' + encodeURIComponent(String(global.RECAPTCHA_SITE_KEY || '').trim());
      } else {
        // v2 checkbox: widget explícito.
        script.src = base + '?onload=' + encodeURIComponent(onloadCallbackName) + '&render=explicit';
      }
      script.async = true;
      script.defer = true;
      script.onload = function () {
        function tryResolve(attempt) {
          var api = resolveLoadedApi();
          if (api) {
            resolve(api);
            return;
          }
          if (attempt < 12) {
            setTimeout(function () {
              tryResolve(attempt + 1);
            }, 50);
            return;
          }
          reject(new Error('Google reCAPTCHA cargó, pero no expuso las funciones esperadas (render o execute).'));
        }
        tryResolve(0);
      };
      script.onerror = function () { reject(new Error('No se pudo cargar Google reCAPTCHA.')); };
      document.head.appendChild(script);
    });
    return scriptPromise;
  }

  function resetScriptLoadState() {
    scriptPromise = null;
    try {
      delete global[onloadCallbackName];
    } catch (e) {
      global[onloadCallbackName] = undefined;
    }
    if (resolveLoadedApi()) {
      return;
    }
    try {
      var scripts = document.querySelectorAll(scriptSelector);
      Array.prototype.forEach.call(scripts, function (script) {
        if (script && script.parentNode) {
          script.parentNode.removeChild(script);
        }
      });
    } catch (e) {}
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

    function showRetry(message) {
      var container = getContainer();
      if (!container) {
        return;
      }
      setContainerVisible(true);
      container.style.minHeight = '0';
      container.innerHTML = '';

      var box = document.createElement('div');
      box.className = 'recaptcha-retry-box';
      box.style.display = 'flex';
      box.style.flexDirection = 'column';
      box.style.alignItems = 'flex-start';
      box.style.gap = '8px';

      var text = document.createElement('span');
      text.textContent = message || 'No se pudo cargar la verificacion de seguridad.';

      var button = document.createElement('button');
      button.type = 'button';
      button.className = 'btn secondary small';
      button.textContent = 'Reintentar reCAPTCHA';
      button.addEventListener('click', function () {
        widgetId = null;
        initializing = null;
        resetScriptLoadState();
        container.innerHTML = 'Reintentando verificacion de seguridad...';
        init().catch(function (error) {
          showRetry((error && error.message) ? error.message : 'No se pudo cargar la verificacion de seguridad.');
        });
      });

      box.appendChild(text);
      box.appendChild(button);
      container.appendChild(box);
    }

    function init() {
      if (!isEnabled() || isBypassEnabled()) {
        setContainerVisible(false);
        return Promise.resolve(false);
      }
      var provider = normalizeProvider(global.RECAPTCHA_PROVIDER);
      if (provider === 'google-recaptcha-v3' || provider === 'google-recaptcha-enterprise') {
        // v3: no hay widget que renderizar
        setContainerVisible(false);
        return loadScript().then(function () { return true; }).catch(function (error) {
          resetScriptLoadState();
          showRetry((error && error.message) ? error.message : 'No se pudo cargar la verificacion de seguridad.');
          throw error;
        });
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
        resetScriptLoadState();
        showRetry((error && error.message) ? error.message : 'No se pudo cargar la verificacion de seguridad.');
        throw error;
      });
      return initializing;
    }

    function ensureToken() {
      if (!isEnabled() || isBypassEnabled()) {
        return Promise.resolve({ ok: true, token: '' });
      }
      var provider = normalizeProvider(global.RECAPTCHA_PROVIDER);
      return init().then(function () {
        if (provider === 'google-recaptcha-v3' || provider === 'google-recaptcha-enterprise') {
          return loadScript().then(function () {
            var api = resolveRecaptchaApiV3();
            if (!api || typeof api.ready !== 'function' || typeof api.execute !== 'function') {
              showRetry('Google reCAPTCHA no expuso execute() para este tipo de clave.');
              return {
                ok: false,
                message: 'Google reCAPTCHA no expuso execute() para este tipo de clave. Comprueba en configuración avanzada que el tipo (v2 / v3 / Enterprise) coincida con la clave creada en Google.'
              };
            }
            return new Promise(function (resolve) {
              var settled = false;
              var timer = setTimeout(function () {
                if (settled) {
                  return;
                }
                settled = true;
                showRetry('La verificacion de seguridad tardo demasiado.');
                resolve({ ok: false, message: 'La verificación de seguridad tardó demasiado. Usa el botón Reintentar reCAPTCHA e inténtalo de nuevo.' });
              }, 18000);
              function finish(result) {
                if (settled) {
                  return;
                }
                settled = true;
                clearTimeout(timer);
                resolve(result);
              }
              try {
                api.ready(function () {
                  var action = String(settings.action || 'submit').trim() || 'submit';
                  api.execute(String(global.RECAPTCHA_SITE_KEY || '').trim(), { action: action }).then(function (token) {
                    finish({ ok: true, token: String(token || '').trim() });
                  }).catch(function (e) {
                    showRetry((e && e.message) ? e.message : 'No se pudo ejecutar la verificacion de seguridad.');
                    finish({ ok: false, message: (e && e.message) ? e.message : 'No se pudo ejecutar la verificación de seguridad.' });
                  });
                });
              } catch (e) {
                showRetry((e && e.message) ? e.message : 'No se pudo ejecutar la verificacion de seguridad.');
                finish({ ok: false, message: (e && e.message) ? e.message : 'No se pudo ejecutar la verificación de seguridad.' });
              }
            });
          });
        }

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
        resetScriptLoadState();
        showRetry((error && error.message) ? error.message : 'No se pudo cargar la verificacion de seguridad.');
        return {
          ok: false,
          message: (error && error.message) ? error.message : 'No se pudo cargar la verificación de seguridad.'
        };
      });
    }

    function reset() {
      var provider = normalizeProvider(global.RECAPTCHA_PROVIDER);
      if (provider === 'google-recaptcha-v3' || provider === 'google-recaptcha-enterprise') {
        return;
      }
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
