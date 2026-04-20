// menu.js - inyecta y gestiona el menú flotante centralizado
(function(){
  var SESSION_STATE_COOKIE = 'browser_session_active';
  var THEME_STORAGE_KEY = 'theme';
  var THEME_COOKIE_NAME = 'pcs_theme';
  var THEME_VALUES = {
    dark: true,
    'dark-violet': true,
    'dark-emerald': true,
    light: true,
    'light-rose': true,
    'light-gold': true
  };

  function getCookie(name){
    var match = String(document.cookie || '').match('(^|;)\\s*' + name + '\\s*=\\s*([^;]+)');
    return match ? match.pop() : '';
  }

  function hasBrowserSessionCookie(){
    return getCookie(SESSION_STATE_COOKIE) === '1';
  }

  function normalizeTheme(theme){
    var value = String(theme || '').trim().toLowerCase();
    return THEME_VALUES[value] ? value : 'dark';
  }

  function isLightTheme(theme){
    return normalizeTheme(theme).indexOf('light') === 0;
  }

  function readStoredTheme(){
    var stored = '';
    try {
      stored = window.localStorage.getItem(THEME_STORAGE_KEY) || '';
    } catch (error) {}
    if (!stored) {
      stored = getCookie(THEME_COOKIE_NAME) || '';
    }
    return normalizeTheme(stored);
  }

  function persistThemeLocally(theme){
    var normalized = normalizeTheme(theme);
    try {
      window.localStorage.setItem(THEME_STORAGE_KEY, normalized);
    } catch (error) {}
    try {
      document.cookie = THEME_COOKIE_NAME + '=' + encodeURIComponent(normalized) + '; Path=/; Max-Age=31536000; SameSite=Lax';
    } catch (error) {}
    return normalized;
  }

  function applyThemeToDocument(doc, theme){
    if (!doc || !doc.documentElement) return;
    var normalized = normalizeTheme(theme);
    var root = doc.documentElement;
    root.setAttribute('data-theme', normalized);
    root.classList.toggle('theme-light', isLightTheme(normalized));
    root.classList.toggle('theme-dark', !isLightTheme(normalized));
  }

  function syncThemeIntoFrames(doc, theme){
    if (!doc || !doc.querySelectorAll) return;
    var frames = doc.querySelectorAll('iframe');
    for (var i = 0; i < frames.length; i += 1) {
      var frame = frames[i];
      try {
        if (frame.contentDocument) {
          applyThemeToDocument(frame.contentDocument, theme);
          syncThemeIntoFrames(frame.contentDocument, theme);
        }
      } catch (error) {}
    }
  }

  function bindThemeToFrames(themeManager){
    function bindFrame(frame){
      if (!frame || frame.dataset.pcsThemeBound === '1') return;
      frame.dataset.pcsThemeBound = '1';
      frame.addEventListener('load', function(){
        themeManager.applyTheme(themeManager.getTheme());
      });
      try {
        if (frame.contentDocument && frame.contentDocument.readyState !== 'loading') {
          applyThemeToDocument(frame.contentDocument, themeManager.getTheme());
          syncThemeIntoFrames(frame.contentDocument, themeManager.getTheme());
        }
      } catch (error) {}
    }

    function bindExistingFrames(){
      if (!document.querySelectorAll) return;
      var frames = document.querySelectorAll('iframe');
      for (var i = 0; i < frames.length; i += 1) {
        bindFrame(frames[i]);
      }
    }

    bindExistingFrames();

    try {
      var observer = new MutationObserver(function(mutations){
        for (var i = 0; i < mutations.length; i += 1) {
          var mutation = mutations[i];
          if (!mutation.addedNodes) continue;
          for (var j = 0; j < mutation.addedNodes.length; j += 1) {
            var node = mutation.addedNodes[j];
            if (!node) continue;
            if (node.tagName === 'IFRAME') {
              bindFrame(node);
            }
            if (node.querySelectorAll) {
              var nested = node.querySelectorAll('iframe');
              for (var k = 0; k < nested.length; k += 1) {
                bindFrame(nested[k]);
              }
            }
          }
        }
      });
      observer.observe(document.documentElement || document.body, { childList: true, subtree: true });
    } catch (error) {}
  }

  function createThemeManager(){
    var currentTheme = persistThemeLocally(readStoredTheme());

    function updateToggleState(theme){
      var toggle = document.getElementById('themeToggle');
      var popup = document.getElementById('themeSelectorPopup');
      if (toggle) {
        toggle.innerHTML = 'Cambiar apariencia \u25BC';
      }
      if (popup && popup.querySelectorAll) {
        var options = popup.querySelectorAll('.theme-option');
        for (var i = 0; i < options.length; i += 1) {
          var option = options[i];
          option.classList.toggle('active', option.getAttribute('data-theme-value') === theme);
        }
      }
    }

    function broadcastTheme(theme){
      try {
        window.dispatchEvent(new CustomEvent('pcs:theme-changed', { detail: { theme: theme } }));
      } catch (error) {}
    }

    function applyTheme(theme){
      currentTheme = persistThemeLocally(theme);
      applyThemeToDocument(document, currentTheme);
      syncThemeIntoFrames(document, currentTheme);
      updateToggleState(currentTheme);
      broadcastTheme(currentTheme);
      return currentTheme;
    }

    function saveThemeToDB(theme){
      if (!hasBrowserSessionCookie()) return;
      fetch('/api/user/configuracion', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'same-origin',
        body: JSON.stringify({ apariencia: theme })
      }).catch(function(err){
        console.log('Error saving theme', err);
      });
    }

    function refreshFromServer(){
      if (!hasBrowserSessionCookie()) return Promise.resolve(currentTheme);
      return fetch('/api/user/configuracion', { credentials: 'same-origin' })
        .then(function(response){
          if (!response.ok) throw new Error('theme unavailable');
          return response.json();
        })
        .then(function(data){
          if (data && data.ok && data.apariencia) {
            applyTheme(data.apariencia);
          }
          return currentTheme;
        })
        .catch(function(err){
          console.log('Error reading theme', err);
          return currentTheme;
        });
    }

    applyTheme(currentTheme);
    bindThemeToFrames({ getTheme: function(){ return currentTheme; }, applyTheme: applyTheme });

    window.addEventListener('storage', function(event){
      if (event && event.key === THEME_STORAGE_KEY && event.newValue) {
        applyTheme(event.newValue);
      }
    });

    window.addEventListener('pcs:theme-request-sync', function(){
      applyTheme(currentTheme);
    });

    return {
      getTheme: function(){ return currentTheme; },
      applyTheme: applyTheme,
      setTheme: function(theme){
        var nextTheme = applyTheme(theme);
        saveThemeToDB(nextTheme);
        return nextTheme;
      },
      refreshFromServer: refreshFromServer,
      normalizeTheme: normalizeTheme,
      isLightTheme: isLightTheme
    };
  }

  var themeManager = createThemeManager();
  window.__pcsThemeManager = themeManager;

  function initAdminMobileSidebarToggle(){
    var sidebars = document.querySelectorAll ? document.querySelectorAll('.admin-sidebar.admin-sidebar-mobile-collapsible') : [];
    if (!sidebars || !sidebars.length) return;

    sidebars.forEach(function(sidebar){
      var nav = sidebar.querySelector('.nav');
      var toggleBtn = sidebar.querySelector('.admin-menu-visibility-toggle');
      if (!nav || !toggleBtn) return;

      var mq = null;
      try {
        mq = window.matchMedia('(max-width: 640px)');
      } catch (e) {
        mq = null;
      }

      var showLabel = toggleBtn.getAttribute('data-show-label') || 'Mostrar menú';
      var hideLabel = toggleBtn.getAttribute('data-hide-label') || 'Ocultar menú';
      var label = toggleBtn.querySelector('.admin-menu-visibility-label');

      function isMobile(){
        return !!(mq && mq.matches);
      }

      function applySidebarState(collapsed){
        var shouldCollapse = !!collapsed && isMobile();
        sidebar.classList.toggle('is-collapsed', shouldCollapse);
        toggleBtn.setAttribute('aria-expanded', shouldCollapse ? 'false' : 'true');
        if (label) {
          label.textContent = shouldCollapse ? showLabel : hideLabel;
        }
      }

      function resetForViewport(){
        applySidebarState(false);
      }

      toggleBtn.addEventListener('click', function(){
        if (!isMobile()) return;
        applySidebarState(!sidebar.classList.contains('is-collapsed'));
      });

      nav.addEventListener('click', function(event){
        var target = event.target && event.target.closest ? event.target.closest('a') : null;
        if (!target || !isMobile()) return;
        window.setTimeout(function(){
          applySidebarState(true);
        }, 0);
      });

      if (mq) {
        if (typeof mq.addEventListener === 'function') {
          mq.addEventListener('change', resetForViewport);
        } else if (typeof mq.addListener === 'function') {
          mq.addListener(resetForViewport);
        }
      }

      resetForViewport();
    });
  }

  function injectMenu(){
    try {
      if (window.top !== window.self) return;
      if (window.__pcsFloatingMenuInjected || document.documentElement.dataset.pcsFloatingMenuInjected === '1') return;
    } catch (e) {
      return;
    }

    try {
      var path = (window.location && window.location.pathname) ? (window.location.pathname || '').toLowerCase() : '';
      if (path === '/' || path === '/index.html' || path.endsWith('/index.html')) {
        return;
      }
      if (path === '/login.html' || path.endsWith('/login.html') || path === '/login_usuario.html' || path.endsWith('/login_usuario.html')) {
        return;
      }
      if (!hasBrowserSessionCookie()) {
        return;
      }
    } catch (e) {
      return;
    }

    if (document.querySelector('.floating-menu')) return;

    var wrapper = document.createElement('div');
    wrapper.className = 'floating-menu';
    wrapper.setAttribute('aria-hidden','false');
    wrapper.innerHTML = '<button class="fm-toggle" aria-label="Abrir menú"><span class="fm-toggle-icon">☰</span></button>' +
      '<div class="fm-panel" role="menu">' +
        '<a class="fm-item" href="/index.html">Portal</a>' +
        '<a class="fm-item" href="/red_social_comercial.html">Red social comercial</a>' +
        '<a class="fm-item" href="/configuracion_de_la_cuenta.html">Configuración de la cuenta</a>' +
        '<a class="fm-item" href="/Juegos/menu_juegos.html">Juegos</a>' +
        '<div class="theme-selector-item" id="themeToggleWrapper" style="position:relative;">' +
          '<button id="themeToggle" class="fm-item theme-toggle-btn" type="button" aria-expanded="false" aria-haspopup="true" aria-label="Cambiar apariencia">Cambiar apariencia \u25BC</button>' +
          '<div id="themeSelectorPopup" class="theme-selector-popup" aria-hidden="true" role="menu">' +
            '<div class="theme-opt-group">Oscuros</div>' +
            '<button class="theme-option" type="button" data-theme-value="dark">Azul Elegante</button>' +
            '<button class="theme-option" type="button" data-theme-value="dark-violet">Morado Midnight</button>' +
            '<button class="theme-option" type="button" data-theme-value="dark-emerald">Negro Esmeralda</button>' +
            '<div class="theme-opt-group mt-1">Claros</div>' +
            '<button class="theme-option" type="button" data-theme-value="light">Blanco Corporativo</button>' +
            '<button class="theme-option" type="button" data-theme-value="light-rose">Rosa Pastel</button>' +
            '<button class="theme-option" type="button" data-theme-value="light-gold">Blanco Dorado</button>' +
          '</div>' +
        '</div>' +
        '<a id="sessionLink" class="fm-item" href="/login.html">Iniciar sesión</a>' +
        '<a class="fm-item" href="/ayuda/ayuda.html">Ayuda</a>' +
      '</div>';

    if (document.body && document.body.firstChild) document.body.insertBefore(wrapper, document.body.firstChild);
    else if (document.body) document.body.appendChild(wrapper);

    try {
      if (document.body) document.body.classList.add('has-floating-menu');
      document.documentElement.classList.add('has-floating-menu');
    } catch (e) {}

    var toggle = wrapper.querySelector('.fm-toggle');
    var panel = wrapper.querySelector('.fm-panel');
    var sessionLink = wrapper.querySelector('#sessionLink');
    var themeToggle = wrapper.querySelector('#themeToggle');
    var themeSelectorPopup = wrapper.querySelector('#themeSelectorPopup');

    function setPanelOpen(isOpen){
      if (!panel || !toggle) return;
      panel.classList.toggle('open', !!isOpen);
      toggle.setAttribute('aria-expanded', isOpen ? 'true' : 'false');
    }

    function closePanel(){
      setPanelOpen(false);
    }

    function closeThemePopup(){
      if (!themeToggle || !themeSelectorPopup) return;
      themeToggle.setAttribute('aria-expanded', 'false');
      themeSelectorPopup.setAttribute('aria-hidden', 'true');
      themeSelectorPopup.classList.remove('show');
    }

    function setSessionLinkAuthenticated(isAuthenticated){
      if (!sessionLink) return;
      sessionLink.onclick = null;
      if (isAuthenticated) {
        sessionLink.textContent = 'Cerrar sesión';
        sessionLink.href = '/auth/logout';
        return;
      }
      sessionLink.textContent = 'Iniciar sesión';
      sessionLink.href = '/login.html';
    }

    setSessionLinkAuthenticated(hasBrowserSessionCookie());
    themeManager.applyTheme(themeManager.getTheme());

    if (toggle && panel) {
      toggle.setAttribute('aria-expanded', 'false');
      toggle.addEventListener('click', function(event){
        event.stopPropagation();
        setPanelOpen(!panel.classList.contains('open'));
      });

      panel.addEventListener('click', function(event){
        event.stopPropagation();
        var item = event.target && event.target.closest ? event.target.closest('.fm-item') : null;
        if (item && item.id !== 'themeToggle' && !item.classList.contains('fm-country')) {
          closeThemePopup();
          closePanel();
        }
      });

      document.addEventListener('click', function(event){
        var clickInsideMenu = wrapper.contains(event.target);
        if (!clickInsideMenu) {
          closeThemePopup();
          closePanel();
          return;
        }
        var clickInsideTheme = themeSelectorPopup && themeSelectorPopup.contains(event.target);
        var clickThemeButton = themeToggle && themeToggle.contains(event.target);
        if (!clickInsideTheme && !clickThemeButton) {
          closeThemePopup();
        }
      });

      document.addEventListener('keydown', function(event){
        if (event.key === 'Escape') {
          closeThemePopup();
          closePanel();
        }
      });
    }

    if (themeToggle && themeSelectorPopup) {
      themeToggle.addEventListener('click', function(event){
        event.stopPropagation();
        var isExpanded = themeToggle.getAttribute('aria-expanded') === 'true';
        themeToggle.setAttribute('aria-expanded', isExpanded ? 'false' : 'true');
        themeSelectorPopup.setAttribute('aria-hidden', isExpanded ? 'true' : 'false');
        themeSelectorPopup.classList.toggle('show', !isExpanded);
      });

      var options = themeSelectorPopup.querySelectorAll('.theme-option');
      for (var i = 0; i < options.length; i += 1) {
        options[i].addEventListener('click', function(event){
          event.stopPropagation();
          var selectedTheme = this.getAttribute('data-theme-value');
          themeManager.setTheme(selectedTheme);
          closeThemePopup();
        });
      }
    }

    try {
      window.__pcsFloatingMenuInjected = true;
      document.documentElement.dataset.pcsFloatingMenuInjected = '1';
    } catch (e) {}

    themeManager.refreshFromServer();

    (function loadAvatar(){
      function setAvatarUrl(url, name){
        try {
          if (!toggle) return;
          toggle.innerHTML = '';
          var img = document.createElement('img');
          img.className = 'fm-avatar';
          img.src = url;
          img.alt = name || 'Perfil';
          img.onerror = function(){
            toggle.innerHTML = '<span class="fm-toggle-icon">☰</span>';
          };
          toggle.appendChild(img);
          if (name) toggle.title = name;
        } catch (error) {}
      }

      function fallbackIcon(){
        try {
          if (!toggle) return;
          toggle.innerHTML = '<span class="fm-toggle-icon">☰</span>';
        } catch (error) {}
      }

      try {
        if (!hasBrowserSessionCookie()) {
          setSessionLinkAuthenticated(false);
          fallbackIcon();
          return;
        }
        fetch('/me', { credentials: 'same-origin' })
          .then(function(res){
            if (!res.ok) throw new Error('no-auth');
            return res.json();
          })
          .then(function(data){
            setSessionLinkAuthenticated(true);
            var photo = (data && (data.photo || data.avatar)) || '';
            var name = (data && (data.name || data.email)) || '';
            if (photo) setAvatarUrl(photo, name); else fallbackIcon();
          })
          .catch(function(){
            setSessionLinkAuthenticated(false);
            fallbackIcon();
          });
      } catch (error) {
        setSessionLinkAuthenticated(false);
        fallbackIcon();
      }
    })();

    (function ensureIconFallback(){
      var defaultIcon = '/img/analytics-color.svg';

      function setFallback(img){
        try {
          if (!img) return;
          if (!img.getAttribute('src') || img.getAttribute('src').trim() === '') img.src = defaultIcon;
          img.addEventListener('error', function(){
            if (img.src !== defaultIcon) img.src = defaultIcon;
          });
        } catch (error) {}
      }

      if (document.querySelectorAll) {
        document.querySelectorAll('.admin-sidebar .nav a img.icon').forEach(setFallback);
      }

      try {
        var observer = new MutationObserver(function(mutations){
          mutations.forEach(function(mutation){
            if (!mutation.addedNodes) return;
            mutation.addedNodes.forEach(function(node){
              if (node && node.querySelectorAll) node.querySelectorAll('.admin-sidebar .nav a img.icon').forEach(setFallback);
              if (node && node.matches && node.matches('.admin-sidebar .nav a img.icon')) setFallback(node);
            });
          });
        });
        observer.observe(document.body, { childList: true, subtree: true });
      } catch (error) {}
    })();
  }

  function initAcceptModalFromQuery(){
    return;
  }

  function init(){
    themeManager.applyTheme(themeManager.getTheme());
    initAcceptModalFromQuery();
    initAdminMobileSidebarToggle();
    injectMenu();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
