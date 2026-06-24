// menu.js - inyecta y gestiona el menú flotante centralizado
(function(){
  var SESSION_STATE_COOKIE = 'browser_session_active';
  var THEME_STORAGE_KEY = 'theme';
  var THEME_COOKIE_NAME = 'pcs_theme';
  var THEME_VALUES = {
    dark: true,
    'dark-violet': true,
    'dark-emerald': true,
    'dark-corporate': true,
    'dark-absolute': true,
    'dark-obsidian': true,
    'dark-neon': true,
    light: true,
    'light-rose': true,
    'light-gold': true,
    'light-wood': true
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
    if (value === 'dark-protect') {
      return 'dark';
    }
    return THEME_VALUES[value] ? value : 'light';
  }

  function isLightTheme(theme){
    return normalizeTheme(theme).indexOf('light') === 0;
  }

  function readStoredTheme(){
    var stored = '';
    try {
      stored = decodeURIComponent(getCookie(THEME_COOKIE_NAME) || '');
    } catch (error) {
      stored = getCookie(THEME_COOKIE_NAME) || '';
    }
    try {
      stored = stored || window.localStorage.getItem(THEME_STORAGE_KEY) || '';
    } catch (error) {}
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
      var userChangedSidebarState = false;

      function isMobile(){
        return !!(mq && mq.matches);
      }

      function shouldAutoCollapseOnEntry(){
        try {
          var body = document.body;
          return !!(body && body.getAttribute('data-auto-collapse-menu') === 'true');
        } catch (e) {
          return false;
        }
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
        applySidebarState(!userChangedSidebarState && shouldAutoCollapseOnEntry());
      }

      toggleBtn.dataset.pcsBound = '1';
      toggleBtn.addEventListener('click', function(){
        if (!isMobile()) return;
        userChangedSidebarState = true;
        applySidebarState(!sidebar.classList.contains('is-collapsed'));
      });

      nav.addEventListener('click', function(event){
        var target = event.target && event.target.closest ? event.target.closest('a, button[data-target]') : null;
        if (!target || target.classList.contains('admin-menu-visibility-toggle') || !isMobile()) return;
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

  function clearBrowserSessionStateCookie(){
    try {
      document.cookie = SESSION_STATE_COOKIE + '=; Path=/; Max-Age=0; SameSite=Lax';
    } catch (error) {}
  }

  function clearFloatingMenuInjectionState(){
    try {
      window.__pcsFloatingMenuInjected = false;
      window.__pcsFloatingMenuPending = false;
      if (document.documentElement && document.documentElement.dataset) {
        delete document.documentElement.dataset.pcsFloatingMenuInjected;
        delete document.documentElement.dataset.pcsFloatingMenuPending;
      }
    } catch (error) {}
  }

  function removeFloatingMenu(){
    try {
      var existingMenu = document.querySelector('.floating-menu');
      if (existingMenu && existingMenu.parentNode) {
        existingMenu.parentNode.removeChild(existingMenu);
      }
      if (document.body) {
        document.body.classList.remove('has-floating-menu');
      }
      if (document.documentElement) {
        document.documentElement.classList.remove('has-floating-menu');
      }
    } catch (error) {}
    clearFloatingMenuInjectionState();
  }

  function shouldSkipFloatingMenuForCurrentPath(){
    try {
      var path = (window.location && window.location.pathname) ? (window.location.pathname || '').toLowerCase() : '';
      if (path === '/' || path === '/index.html' || path.endsWith('/index.html')) {
        return true;
      }
      if (path === '/login.html' || path.endsWith('/login.html') || path === '/login_usuario.html' || path.endsWith('/login_usuario.html')) {
        return true;
      }
      // Páginas públicas donde NO debe mostrarse el menú flotante (perfil/sesión)
      if (path === '/descripcion_de_los_sistemas' || path === '/descripcion_de_los_sistemas.ht' || path.indexOf('/descripcion_de_los_sistemas.') === 0) {
        return true;
      }
    } catch (error) {
      return true;
    }
    return false;
  }

  function injectMenu(){
    try {
      if (window.top !== window.self) return;
      if (shouldSkipFloatingMenuForCurrentPath()) {
        removeFloatingMenu();
        return;
      }
      if (!hasBrowserSessionCookie()) {
        removeFloatingMenu();
        return;
      }
      if (window.__pcsFloatingMenuInjected || document.documentElement.dataset.pcsFloatingMenuInjected === '1' || window.__pcsFloatingMenuPending) {
        return;
      }
      window.__pcsFloatingMenuPending = true;
      document.documentElement.dataset.pcsFloatingMenuPending = '1';
    } catch (e) {
      return;
    }

    fetch('/me', { credentials: 'same-origin' })
      .then(function(res){
        if (!res.ok) throw new Error('no-auth');
        return res.json().catch(function(){ return {}; });
      })
      .then(function(data){
        try {
          window.__pcsFloatingMenuPending = false;
          delete document.documentElement.dataset.pcsFloatingMenuPending;
        } catch (error) {}
        mountFloatingMenu(data);
      })
      .catch(function(){
        clearBrowserSessionStateCookie();
        removeFloatingMenu();
      });
  }

  function mountFloatingMenu(authData){
    try {
      if (window.top !== window.self) return;
      if (window.__pcsFloatingMenuInjected || document.documentElement.dataset.pcsFloatingMenuInjected === '1') return;
    } catch (e) {
      return;
    }

    try {
      if (shouldSkipFloatingMenuForCurrentPath() || !hasBrowserSessionCookie()) {
        return;
      }
    } catch (e) {
      return;
    }

    if (document.querySelector('.floating-menu')) return;

    var wrapper = document.createElement('div');
    wrapper.className = 'floating-menu';
    wrapper.setAttribute('aria-hidden','false');
    wrapper.innerHTML = '<button class="fm-toggle" aria-label="Abrir menú"><span class="fm-toggle-icon">☰</span><span id="floatingMenuNotificationBadge" class="fm-toggle-badge" hidden>0</span></button>' +
      '<div class="fm-panel" role="menu">' +
        '<button id="floatingNotificationBell" class="fm-item fm-action-item fm-icon-item fm-notification-item" type="button" aria-expanded="false" aria-controls="floatingNotificationPanel" aria-label="Notificaciones"><span class="fm-notification-icon" aria-hidden="true">&#128276;</span><span class="fm-notification-label">Notificaciones</span><span id="floatingNotificationCount" class="fm-notification-count" hidden>0</span></button>' +
        '<div id="floatingNotificationPanel" class="fm-notification-panel" hidden>' +
          '<div class="fm-notification-panel-head"><strong>Buz&oacute;n de usuario</strong><button id="floatingNotificationRefresh" type="button">Actualizar</button></div>' +
          '<div id="floatingNotificationList" class="fm-notification-list"><p class="fm-notification-empty">Sin mensajes pendientes.</p></div>' +
        '</div>' +
        '<a class="fm-item fm-icon-item" href="/index.html" data-admin-frame-url="/index.html"><img class="fm-item-icon" src="/img/company-briefcase-color.svg" alt="">Portal</a>' +
        '<a class="fm-item fm-icon-item" href="/red_social_comercial.html" data-admin-frame-url="/red_social_comercial.html"><img class="fm-item-icon" src="/img/social.svg" alt="">Red social comercial</a>' +
        '<a class="fm-item fm-icon-item" href="/administrar_empresa/noticias.html" data-admin-frame-url="/administrar_empresa/noticias.html"><img class="fm-item-icon" src="/img/report.svg" alt="">Noticias</a>' +
        '<button id="createHelpTicketLink" class="fm-item fm-action-item fm-icon-item" type="button"><img class="fm-item-icon" src="/img/shield-security-color.svg" alt="">Crear ticket de ayuda</button>' +
        '<button id="openFloatingAILink" class="fm-item fm-action-item fm-icon-item" type="button"><img class="fm-item-icon" data-ai-logo="true" src="/img/pcs_ia_logo.svg" alt="">Robot / IA</button>' +
        '<button id="openFloatingRadioLink" class="fm-item fm-action-item fm-icon-item" type="button"><img class="fm-item-icon" src="/img/play.svg" alt="">Emisoras</button>' +
        '<div class="fm-submenu" id="utilitiesMenuWrapper">' +
          '<button id="utilitiesMenuToggle" class="fm-item fm-submenu-toggle fm-icon-item" type="button" aria-expanded="false" aria-haspopup="true"><img class="fm-item-icon" src="/img/settings-color.svg" alt="">Utilidades \u25BC</button>' +
          '<div id="utilitiesMenuPopup" class="fm-submenu-popup" aria-hidden="true" role="menu">' +
            '<a class="fm-item fm-subitem fm-icon-item" href="/calculadora.html?compact=1" data-open-calculator="1"><img class="fm-item-icon" src="/img/analytics-color.svg" alt="">Calculadora</a>' +
            '<button class="fm-item fm-subitem fm-action-item fm-icon-item" type="button" data-share-current="email"><img class="fm-item-icon" src="/img/network-color.svg" alt="">Compartir por correo</button>' +
            '<a class="fm-item fm-subitem fm-icon-item" href="/Juegos/menu_juegos.html" data-admin-frame-url="/Juegos/menu_juegos.html"><img class="fm-item-icon" src="/img/play.svg" alt="">Juegos</a>' +
            '<a class="fm-item fm-subitem fm-icon-item" href="/emulador/" data-admin-frame-url="/emulador/"><img class="fm-item-icon" src="/img/settings-color.svg" alt="">Emulador</a>' +
          '</div>' +
        '</div>' +
        '<a class="fm-item" href="/configuracion_de_la_cuenta.html" data-admin-frame-url="/configuracion_de_la_cuenta.html">Configuración de la cuenta</a>' +
        '' +
        '<div class="theme-selector-item" id="themeToggleWrapper" style="position:relative;">' +
          '<button id="themeToggle" class="fm-item theme-toggle-btn" type="button" aria-expanded="false" aria-haspopup="true" aria-label="Cambiar apariencia">Cambiar apariencia \u25BC</button>' +
          '<div id="themeSelectorPopup" class="theme-selector-popup" aria-hidden="true" role="menu">' +
            '<div class="theme-opt-group">Oscuros</div>' +
            '<button class="theme-option" type="button" data-theme-value="dark">Azul Elegante</button>' +
            '<button class="theme-option" type="button" data-theme-value="dark-violet">Morado Midnight</button>' +
            '<button class="theme-option" type="button" data-theme-value="dark-emerald">Negro Esmeralda</button>' +
            '<button class="theme-option" type="button" data-theme-value="dark-corporate">Corporativo Oscuro</button>' +
            '<button class="theme-option" type="button" data-theme-value="dark-absolute">Negro Absoluto</button>' +
            '<button class="theme-option" type="button" data-theme-value="dark-obsidian">Obsidiana Profesional</button>' +
            '<button class="theme-option" type="button" data-theme-value="dark-neon">Neon Nocturno</button>' +
            '<div class="theme-opt-group mt-1">Claros</div>' +
            '<button class="theme-option" type="button" data-theme-value="light">Blanco Corporativo</button>' +
            '<button class="theme-option" type="button" data-theme-value="light-rose">Rosa Pastel</button>' +
            '<button class="theme-option" type="button" data-theme-value="light-gold">Blanco Dorado</button>' +
            '<button class="theme-option" type="button" data-theme-value="light-wood">Madera Clara</button>' +
          '</div>' +
        '</div>' +
        '<a id="sessionLink" class="fm-item" href="/login.html">Iniciar sesión</a>' +
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
    var utilitiesToggle = wrapper.querySelector('#utilitiesMenuToggle');
    var utilitiesPopup = wrapper.querySelector('#utilitiesMenuPopup');
    var calculatorLauncher = wrapper.querySelector('[data-open-calculator]');
    var shareLaunchers = wrapper.querySelectorAll('[data-share-current]');
    var helpTicketLauncher = wrapper.querySelector('#createHelpTicketLink');
    var floatingAILauncher = wrapper.querySelector('#openFloatingAILink');
    var floatingRadioLauncher = wrapper.querySelector('#openFloatingRadioLink');
    var floatingNotificationBell = wrapper.querySelector('#floatingNotificationBell');
    var floatingNotificationCount = wrapper.querySelector('#floatingNotificationCount');
    var floatingMenuNotificationBadge = wrapper.querySelector('#floatingMenuNotificationBadge');
    var floatingNotificationPanel = wrapper.querySelector('#floatingNotificationPanel');
    var floatingNotificationList = wrapper.querySelector('#floatingNotificationList');
    var floatingNotificationRefresh = wrapper.querySelector('#floatingNotificationRefresh');

    function ensureFloatingMenuIcons(){
      var fallback = {
        sessionLink: '/img/shield-security-color.svg',
        themeToggle: '/img/settings-color.svg',
        utilitiesMenuToggle: '/img/settings-color.svg',
        createHelpTicketLink: '/img/shield-security-color.svg'
      };
      wrapper.querySelectorAll('.fm-item').forEach(function(item){
        if (item.querySelector('.fm-item-icon') || item.querySelector('.fm-notification-icon')) return;
        var src = fallback[item.id] || '/img/report.svg';
        if (item.matches('a[href*="configuracion"]')) src = '/img/settings-color.svg';
        if (item.matches('a[href*="index"]')) src = '/img/company-briefcase-color.svg';
        if (item.matches('a[href*="red_social"]')) src = '/img/social.svg';
        if (item.matches('a[href*="calculadora"]')) src = '/img/analytics-color.svg';
        if (item.matches('a[href*="Juegos"],a[href*="emulador"]')) src = '/img/play.svg';
        var img = document.createElement('img');
        img.className = 'fm-item-icon';
        img.src = src;
        img.alt = '';
        item.classList.add('fm-icon-item');
        item.insertBefore(img, item.firstChild);
      });
    }
    ensureFloatingMenuIcons();

    function normalizeNotificationCount(value){
      if (typeof value === 'string' && value.indexOf('+') !== -1) return 100;
      var count = Number(value || 0);
      if (!Number.isFinite(count) || count < 0) return 0;
      return Math.floor(count);
    }

    function ensureFloatingMenuBadge(){
      if (!toggle) return null;
      floatingMenuNotificationBadge = toggle.querySelector('#floatingMenuNotificationBadge');
      if (!floatingMenuNotificationBadge) {
        floatingMenuNotificationBadge = document.createElement('span');
        floatingMenuNotificationBadge.id = 'floatingMenuNotificationBadge';
        floatingMenuNotificationBadge.className = 'fm-toggle-badge';
        floatingMenuNotificationBadge.hidden = true;
        floatingMenuNotificationBadge.textContent = '0';
        toggle.appendChild(floatingMenuNotificationBadge);
      }
      return floatingMenuNotificationBadge;
    }

    function setFloatingNotificationCount(value){
      var unread = normalizeNotificationCount(value);
      var label = unread > 99 ? '99+' : String(unread);
      var hasUnread = unread > 0;
      if (floatingNotificationCount) {
        floatingNotificationCount.textContent = label;
        floatingNotificationCount.hidden = !hasUnread;
      }
      var toggleBadge = ensureFloatingMenuBadge();
      if (toggleBadge) {
        toggleBadge.textContent = label;
        toggleBadge.hidden = !hasUnread;
      }
      wrapper.classList.toggle('has-notifications', hasUnread);
      if (toggle) {
        toggle.setAttribute('aria-label', hasUnread ? ('Abrir menú, ' + label + ' notificaciones') : 'Abrir menú');
      }
      if (floatingNotificationBell) {
        floatingNotificationBell.setAttribute('aria-label', hasUnread ? ('Notificaciones, ' + label + ' pendientes') : 'Notificaciones');
        floatingNotificationBell.classList.toggle('has-unread', hasUnread);
      }
    }

    function syncFloatingNotificationFromAdminBell(){
      try {
        var adminBadge = document.getElementById('adminNotificationBadge');
        var adminBell = document.getElementById('adminNotificationBell');
        if (!adminBadge && !adminBell) return;
        var text = adminBadge ? String(adminBadge.textContent || '').trim() : '0';
        var unread = normalizeNotificationCount(text);
        if (adminBell && adminBell.classList.contains('has-unread') && unread <= 0) unread = 1;
        setFloatingNotificationCount(unread);
      } catch (error) {}
    }

    function renderFloatingNotifications(data){
      if (!floatingNotificationList) return;
      var unread = normalizeNotificationCount(data && data.unread);
      setFloatingNotificationCount(unread);
      var messages = Array.isArray(data && data.mensajes) ? data.mensajes.slice(0, 8) : [];
      floatingNotificationList.innerHTML = '';
      if (!messages.length) {
        var empty = document.createElement('p');
        empty.className = 'fm-notification-empty';
        empty.textContent = 'Sin mensajes pendientes.';
        floatingNotificationList.appendChild(empty);
        return;
      }
      messages.forEach(function(msg){
        var item = document.createElement('button');
        item.type = 'button';
        item.className = 'fm-notification-row';
        var title = document.createElement('strong');
        title.textContent = String((msg && msg.titulo) || 'Mensaje');
        var body = document.createElement('span');
        body.textContent = String((msg && msg.mensaje) || '').slice(0, 140);
        item.appendChild(title);
        item.appendChild(body);
        item.addEventListener('click', function(event){
          event.preventDefault();
          event.stopPropagation();
          markFloatingNotificationReadAndOpen(msg);
        });
        floatingNotificationList.appendChild(item);
      });
    }

    function loadFloatingNotifications(){
      var empresaID = getActiveEmpresaID();
      if (!empresaID || !floatingNotificationList) {
        renderFloatingNotifications({ unread: 0, mensajes: [] });
        return Promise.resolve(null);
      }
      floatingNotificationList.innerHTML = '<p class="fm-notification-empty">Cargando notificaciones...</p>';
      return fetch('/api/empresa/buzon?empresa_id=' + encodeURIComponent(empresaID) + '&action=resumen', { credentials: 'same-origin' })
        .then(function(response){
          if (!response.ok) throw new Error('HTTP ' + response.status);
          return response.json();
        })
        .then(function(data){
          renderFloatingNotifications(data || {});
          return data;
        })
        .catch(function(){
          floatingNotificationList.innerHTML = '<p class="fm-notification-empty">No se pudieron cargar las notificaciones.</p>';
          syncFloatingNotificationFromAdminBell();
          return null;
        });
    }

    function markFloatingNotificationReadAndOpen(msg){
      var empresaID = getActiveEmpresaID();
      if (!msg || !msg.id || !empresaID) return;
      fetch('/api/empresa/buzon?empresa_id=' + encodeURIComponent(empresaID) + '&action=leer', {
        method: 'PUT',
        credentials: 'same-origin',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: msg.id })
      }).catch(function(){}).then(function(){
        var href = String(msg.enlace_url || '').trim() || '/administrar_empresa/panel.html';
        if (href.indexOf('/administrar_empresa/') === 0) {
          try {
            var target = new URL(href, window.location.origin);
            target.searchParams.set('empresa_id', String(empresaID));
            href = target.pathname + target.search + target.hash;
          } catch (error) {
            href += (href.indexOf('?') === -1 ? '?' : '&') + 'empresa_id=' + encodeURIComponent(empresaID);
          }
          openPathInContentFrame(href);
        } else {
          window.location.href = href;
        }
        if (floatingNotificationPanel) floatingNotificationPanel.hidden = true;
        if (floatingNotificationBell) floatingNotificationBell.setAttribute('aria-expanded', 'false');
        closePanel();
        loadFloatingNotifications();
      });
    }

    function openAdminNotificationsFromFloating(){
      if (floatingNotificationPanel) {
        var willOpen = floatingNotificationPanel.hidden;
        floatingNotificationPanel.hidden = !willOpen;
        if (floatingNotificationBell) {
          floatingNotificationBell.setAttribute('aria-expanded', willOpen ? 'true' : 'false');
        }
        if (willOpen) {
          loadFloatingNotifications();
        }
        return;
      }
      try {
        var adminBell = document.getElementById('adminNotificationBell');
        if (adminBell && typeof adminBell.click === 'function') {
          adminBell.click();
          closePanel();
          return;
        }
      } catch (error) {}
      var empresaID = getActiveEmpresaID();
      var fallback = '/administrar_empresa/panel.html' + (empresaID ? ('?empresa_id=' + encodeURIComponent(empresaID)) : '');
      if (!openPathInContentFrame(fallback)) {
        window.location.href = fallback;
      }
      closePanel();
    }

    function setPanelOpen(isOpen){
      if (!panel || !toggle) return;
      panel.classList.toggle('open', !!isOpen);
      toggle.setAttribute('aria-expanded', isOpen ? 'true' : 'false');
    }

    function closePanel(){
      if (floatingNotificationPanel) floatingNotificationPanel.hidden = true;
      if (floatingNotificationBell) floatingNotificationBell.setAttribute('aria-expanded', 'false');
      setPanelOpen(false);
    }

    function closeThemePopup(){
      if (!themeToggle || !themeSelectorPopup) return;
      themeToggle.setAttribute('aria-expanded', 'false');
      themeSelectorPopup.setAttribute('aria-hidden', 'true');
      themeSelectorPopup.classList.remove('show');
      themeSelectorPopup.style.maxHeight = '';
      themeSelectorPopup.style.top = '';
      themeSelectorPopup.style.bottom = '';
      themeSelectorPopup.style.left = '';
      themeSelectorPopup.style.right = '';
    }

    function closeUtilitiesPopup(){
      if (!utilitiesToggle || !utilitiesPopup) return;
      utilitiesToggle.setAttribute('aria-expanded', 'false');
      utilitiesPopup.setAttribute('aria-hidden', 'true');
      utilitiesPopup.classList.remove('show');
      utilitiesPopup.style.maxHeight = '';
      utilitiesPopup.style.top = '';
      utilitiesPopup.style.bottom = '';
      utilitiesPopup.style.left = '';
      utilitiesPopup.style.right = '';
    }

    function fitPopupToViewport(popup, anchor){
      if (!popup || !anchor) return;
      var viewportH = window.innerHeight || document.documentElement.clientHeight || 0;
      var viewportW = window.innerWidth || document.documentElement.clientWidth || 0;
      var rect = popup.getBoundingClientRect();
      var anchorRect = anchor.getBoundingClientRect();
      var margin = 12;
      popup.style.maxHeight = Math.max(160, viewportH - margin * 2) + 'px';
      popup.style.overflowY = 'auto';
      popup.style.overflowX = 'hidden';
      if (rect.bottom > viewportH - margin || anchorRect.bottom + rect.height > viewportH) {
        popup.style.top = 'auto';
        popup.style.bottom = '0';
      }
      if (rect.top < margin) {
        popup.style.top = '0';
        popup.style.bottom = 'auto';
      }
      if (rect.right > viewportW - margin) {
        popup.style.left = 'auto';
        popup.style.right = '0';
      }
      if (rect.left < margin) {
        popup.style.left = 'auto';
        popup.style.right = 'calc(100% + 8px)';
      }
    }

    function openCalculatorWindow(){
      var width = 360;
      var height = 520;
      var left = Math.max(0, Math.round((window.screenX || window.screenLeft || 0) + ((window.outerWidth || window.innerWidth || width) - width) / 2));
      var top = Math.max(0, Math.round((window.screenY || window.screenTop || 0) + ((window.outerHeight || window.innerHeight || height) - height) / 2));
      var features = [
        'popup=yes',
        'width=' + width,
        'height=' + height,
        'left=' + left,
        'top=' + top,
        'resizable=yes',
        'scrollbars=no',
        'menubar=no',
        'toolbar=no',
        'location=no',
        'status=no'
      ].join(',');
      var win = window.open('/calculadora.html?compact=1', 'pcs_calculadora', features);
      if (win && typeof win.focus === 'function') {
        win.focus();
      } else {
        window.location.href = '/calculadora.html?compact=1';
      }
    }

    function openPathInContentFrame(path){
      try {
        var frame = document.getElementById('contentFrame') || document.querySelector('iframe.admin-empresa-frame');
        if (frame) {
          frame.src = path;
          return true;
        }
      } catch (error) {}
      return false;
    }

    function openFloatingAI(){
      try {
        if (window.PCSAIChatOpen && typeof window.PCSAIChatOpen === 'function') {
          if (window.PCSAIChatOpen({ source: 'menu_flotante', preferRobot: true }) !== false) {
            return;
          }
        }
      } catch (error) {}
      try {
        var aiBtn = document.getElementById('openAIDrawer');
        if (aiBtn) {
          window.postMessage({ type: 'pcs-ai-drawer-open', source: 'menu_flotante' }, window.location.origin);
          return;
        }
      } catch (error) {}
      var empresaID = getActiveEmpresaID();
      var fallback = '/administrar_empresa/chat_con_inteligencia_artificial.html' + (empresaID ? ('?empresa_id=' + encodeURIComponent(empresaID)) : '');
      if (!openPathInContentFrame(fallback)) {
        window.location.href = fallback;
      }
    }

    function openFloatingRadio(){
      try {
        if (window.__pcsRadioPlayerOpenDrawer && typeof window.__pcsRadioPlayerOpenDrawer === 'function') {
          window.__pcsRadioPlayerOpenDrawer({ source: 'menu_flotante' });
          return;
        }
      } catch (error) {}
      try {
        var radioBtn = document.getElementById('openRadioDrawer');
        if (radioBtn) {
          radioBtn.hidden = false;
          radioBtn.removeAttribute('aria-hidden');
          radioBtn.click();
          return;
        }
      } catch (error) {}
      var empresaID = getActiveEmpresaID();
      var fallback = '/administrar_empresa/radio_online.html' + (empresaID ? ('?empresa_id=' + encodeURIComponent(empresaID)) : '');
      if (!openPathInContentFrame(fallback)) {
        window.location.href = fallback;
      }
    }

    function getStoredEmpresaID(){
      var keys = ['active_empresa_id', 'empresa_id', 'admin_empresa_id'];
      var stores = [];
      try { stores.push(window.sessionStorage); } catch (error) {}
      try { stores.push(window.localStorage); } catch (error) {}
      for (var s = 0; s < stores.length; s += 1) {
        var store = stores[s];
        if (!store) continue;
        for (var i = 0; i < keys.length; i += 1) {
          try {
            var raw = store.getItem(keys[i]) || '';
            var id = parseInt(String(raw).trim(), 10);
            if (Number.isFinite(id) && id > 0) return String(id);
          } catch (error) {}
        }
      }
      return '';
    }

    function getActiveEmpresaID(){
      try {
        if (typeof window.__resolveEmpresaIdContext === 'function') {
          var resolved = window.__resolveEmpresaIdContext();
          if (resolved) return String(resolved);
        }
      } catch (error) {}
      try {
        var params = new URLSearchParams(window.location.search || '');
        var id = parseInt(params.get('empresa_id') || params.get('id') || '', 10);
        if (Number.isFinite(id) && id > 0) return String(id);
      } catch (error) {}
      return getStoredEmpresaID();
    }

    function getActiveSystemPath(){
      try {
        var frame = document.getElementById('contentFrame') || document.querySelector('iframe.admin-empresa-frame');
        if (frame) {
          try {
            var nested = frame.contentDocument && frame.contentDocument.querySelector ? frame.contentDocument.querySelector('#configuracionContentFrame, iframe[src]') : null;
            var nestedSrc = nested ? (nested.getAttribute('src') || nested.src || '') : '';
            if (nestedSrc) return nestedSrc;
          } catch (nestedError) {}
          var src = frame.getAttribute('src') || frame.src || '';
          if (src) return src;
        }
      } catch (error) {}
      try {
        return window.location.pathname + window.location.search;
      } catch (error) {
        return '';
      }
    }

    function getActiveSystemTitle(){
      try {
        var frame = document.getElementById('contentFrame') || document.querySelector('iframe.admin-empresa-frame');
        if (frame && frame.contentDocument) {
          var nested = frame.contentDocument.querySelector('#configuracionContentFrame, iframe[src]');
          if (nested && nested.contentDocument && nested.contentDocument.title) return nested.contentDocument.title;
          if (frame.contentDocument.title) return frame.contentDocument.title;
        }
      } catch (error) {}
      return document.title || 'Powerful Control System';
    }

    function buildActiveShareURL(){
      var raw = getActiveSystemPath();
      try {
        var u = new URL(raw || '/', window.location.origin);
        return u.toString();
      } catch (error) {
        try {
          return window.location.href;
        } catch (e) {
          return '';
        }
      }
    }

    function shareCurrentSystem(channel){
      var title = getActiveSystemTitle() || 'Powerful Control System';
      var url = buildActiveShareURL();
      var moduleName = getActiveSystemModule();
      if (window.PCSPrint && typeof window.PCSPrint.shareDocument === 'function') {
        window.PCSPrint.shareDocument({
          channel: channel,
          title: title,
          code: moduleName,
          message: 'Documento, reporte o pantalla compartida desde Powerful Control System.',
          url: url
        });
        return;
      }
      var body = encodeURIComponent(title + '\nModulo: ' + moduleName + '\nEnlace: ' + url);
      var href = channel === 'whatsapp' ? ('https://wa.me/?text=' + body) : ('mailto:?subject=' + encodeURIComponent(title) + '&body=' + body);
      try {
        window.open(href, '_blank', 'noopener,noreferrer');
      } catch (error) {
        window.location.href = href;
      }
    }

    function getActiveSystemModule(){
      var raw = getActiveSystemPath();
      try {
        var u = new URL(raw, window.location.origin);
        var moduleParam = u.searchParams.get('module') || '';
        if (moduleParam) return moduleParam;
        var file = (u.pathname.split('/').pop() || '').replace(/\.html$/i, '');
        return file || 'sistema';
      } catch (error) {
        return 'sistema';
      }
    }

    function collectHelpTicketContext(){
      var theme = '';
      try {
        theme = document.documentElement.getAttribute('data-theme') || document.body.getAttribute('data-theme') || '';
      } catch (error) {}
      return {
        titulo: getActiveSystemTitle(),
        ruta: getActiveSystemPath(),
        modulo: getActiveSystemModule(),
        viewport: String(window.innerWidth || '') + 'x' + String(window.innerHeight || ''),
        screen: window.screen ? String(window.screen.width || '') + 'x' + String(window.screen.height || '') : '',
        user_agent: navigator.userAgent || '',
        idioma: navigator.language || '',
        tema: theme,
        online: navigator.onLine ? 'true' : 'false',
        hora_local: new Date().toString()
      };
    }

    function ensureHelpTicketDialog(){
      var existing = document.getElementById('pcsHelpTicketBackdrop');
      if (existing) return existing;
      var backdrop = document.createElement('div');
      backdrop.id = 'pcsHelpTicketBackdrop';
      backdrop.className = 'pcs-help-ticket-backdrop';
      backdrop.setAttribute('aria-hidden', 'true');
      backdrop.innerHTML =
        '<section class="pcs-help-ticket-dialog" role="dialog" aria-modal="true" aria-labelledby="pcsHelpTicketTitle">' +
          '<div class="pcs-help-ticket-header">' +
            '<div><h2 id="pcsHelpTicketTitle">Crear ticket de ayuda</h2><p>El equipo de soporte recibira tu solicitud con la empresa, pagina activa y datos tecnicos basicos.</p></div>' +
            '<button id="pcsHelpTicketClose" class="pcs-help-ticket-close" type="button" aria-label="Cerrar">x</button>' +
          '</div>' +
          '<form id="pcsHelpTicketForm" class="pcs-help-ticket-form">' +
            '<div id="pcsHelpTicketContext" class="pcs-help-ticket-context">Pagina activa: sistema</div>' +
            '<label class="form-label" for="pcsHelpTicketSubject">Asunto</label>' +
            '<input id="pcsHelpTicketSubject" class="form-input" type="text" maxlength="180" required placeholder="Ej: No puedo cerrar caja">' +
            '<div class="pcs-help-ticket-grid">' +
              '<label class="form-col"><span class="form-label">Categoria</span><select id="pcsHelpTicketCategory" class="form-input"><option value="general">General</option><option value="tecnico">Tecnico</option><option value="operacion">Operacion</option><option value="configuracion">Configuracion</option><option value="facturacion">Facturacion</option><option value="pagos">Pagos</option><option value="licencias">Licencias</option><option value="usuarios">Usuarios</option><option value="seguridad">Seguridad</option></select></label>' +
              '<label class="form-col"><span class="form-label">Prioridad</span><select id="pcsHelpTicketPriority" class="form-input"><option value="media">Media</option><option value="baja">Baja</option><option value="alta">Alta</option><option value="critica">Critica</option></select></label>' +
            '</div>' +
            '<div class="pcs-help-ticket-grid">' +
              '<label class="form-col"><span class="form-label">Contacto preferido</span><select id="pcsHelpTicketContactPreference" class="form-input"><option value="email">Email</option><option value="whatsapp">WhatsApp</option><option value="telefono">Telefono</option></select></label>' +
              '<label class="form-col"><span class="form-label">Telefono o WhatsApp</span><input id="pcsHelpTicketContactPhone" class="form-input" type="tel" maxlength="80" placeholder="Opcional"></label>' +
            '</div>' +
            '<label class="form-label" for="pcsHelpTicketMessage">Mensaje</label>' +
            '<textarea id="pcsHelpTicketMessage" class="form-textarea" maxlength="4000" required placeholder="Describe que intentabas hacer, que viste y que necesitas resolver."></textarea>' +
            '<label class="pcs-help-ticket-toggle"><input id="pcsHelpTicketIncludeContext" type="checkbox" checked> Incluir contexto tecnico de esta pantalla</label>' +
            '<section class="pcs-help-ticket-recent" aria-label="Tickets recientes">' +
              '<div class="pcs-help-ticket-recent-title">Tickets recientes de esta empresa</div>' +
              '<div id="pcsHelpTicketRecent" class="pcs-help-ticket-recent-list">Cargando...</div>' +
            '</section>' +
            '<div id="pcsHelpTicketStatus" class="pcs-help-ticket-status" aria-live="polite"></div>' +
            '<div class="pcs-help-ticket-actions"><button type="button" id="pcsHelpTicketCancel" class="btn secondary">Cancelar</button><button id="pcsHelpTicketSubmit" type="submit" class="btn primary">Enviar ticket</button></div>' +
          '</form>' +
        '</section>';
      document.body.appendChild(backdrop);

      function closeDialog(){
        backdrop.classList.remove('is-open');
        backdrop.setAttribute('aria-hidden', 'true');
      }
      backdrop.querySelector('#pcsHelpTicketClose').addEventListener('click', closeDialog);
      backdrop.querySelector('#pcsHelpTicketCancel').addEventListener('click', closeDialog);
      backdrop.addEventListener('click', function(event){
        if (event.target === backdrop) closeDialog();
      });
      backdrop.querySelector('#pcsHelpTicketForm').addEventListener('submit', function(event){
        event.preventDefault();
        submitHelpTicket(backdrop);
      });
      return backdrop;
    }

    function setHelpTicketStatus(dialog, text, isError){
      var status = dialog ? dialog.querySelector('#pcsHelpTicketStatus') : null;
      if (!status) return;
      status.textContent = text || '';
      status.classList.toggle('is-error', !!isError);
    }

    function setHelpTicketBusy(dialog, busy){
      var submit = dialog ? dialog.querySelector('#pcsHelpTicketSubmit') : null;
      var cancel = dialog ? dialog.querySelector('#pcsHelpTicketCancel') : null;
      if (submit) {
        submit.disabled = !!busy;
        submit.textContent = busy ? 'Enviando...' : 'Enviar ticket';
      }
      if (cancel) cancel.disabled = !!busy;
    }

    function formatHelpTicketDate(value){
      var raw = String(value || '').trim();
      if (!raw) return '';
      return raw.length > 19 ? raw.slice(0, 19).replace('T', ' ') : raw.replace('T', ' ');
    }

    function escapeHtml(value){
      return String(value == null ? '' : value).replace(/[&<>"']/g, function(ch){
        return ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[ch];
      });
    }

    function renderRecentHelpTickets(dialog, tickets){
      var box = dialog ? dialog.querySelector('#pcsHelpTicketRecent') : null;
      if (!box) return;
      if (!Array.isArray(tickets) || !tickets.length) {
        box.innerHTML = '<div class="pcs-help-ticket-empty">Aun no hay tickets recientes para esta empresa.</div>';
        return;
      }
      box.innerHTML = tickets.slice(0, 5).map(function(ticket){
        var code = escapeHtml(ticket.codigo || ('#' + ticket.id));
        var status = escapeHtml(String(ticket.estado || 'nuevo').replace('_', ' '));
        var subject = escapeHtml(ticket.asunto || 'Sin asunto');
        var date = escapeHtml(formatHelpTicketDate(ticket.fecha_actualizacion || ticket.fecha_creacion));
        return '<article class="pcs-help-ticket-recent-item"><strong>' + code + '</strong><span>' + status + '</span><p>' + subject + '</p><small>' + date + '</small></article>';
      }).join('');
    }

    function loadRecentHelpTickets(dialog, empresaID){
      var box = dialog ? dialog.querySelector('#pcsHelpTicketRecent') : null;
      if (!box) return;
      if (!empresaID) {
        box.innerHTML = '<div class="pcs-help-ticket-empty">Abre una empresa para ver sus tickets recientes.</div>';
        return;
      }
      box.textContent = 'Cargando...';
      fetch('/api/empresa/tickets_ayuda?empresa_id=' + encodeURIComponent(empresaID) + '&limit=5', { credentials: 'same-origin' })
        .then(function(res){
          return res.json().catch(function(){ return {}; }).then(function(data){ return { ok: res.ok, data: data }; });
        })
        .then(function(result){
          if (!result.ok || !result.data || !result.data.ok) throw new Error('No disponible');
          renderRecentHelpTickets(dialog, result.data.tickets || []);
        })
        .catch(function(){
          box.innerHTML = '<div class="pcs-help-ticket-empty">No se pudieron cargar los tickets recientes.</div>';
        });
    }

    function openHelpTicketDialog(){
      var dialog = ensureHelpTicketDialog();
      var empresaID = getActiveEmpresaID();
      setHelpTicketStatus(dialog, empresaID ? '' : 'Abre una empresa antes de crear el ticket para asociarlo correctamente.', !empresaID);
      setHelpTicketBusy(dialog, false);
      var contextLabel = dialog.querySelector('#pcsHelpTicketContext');
      if (contextLabel) {
        contextLabel.textContent = 'Pagina activa: ' + getActiveSystemTitle() + ' - ' + getActiveSystemModule();
      }
      var subject = dialog.querySelector('#pcsHelpTicketSubject');
      var message = dialog.querySelector('#pcsHelpTicketMessage');
      var phone = dialog.querySelector('#pcsHelpTicketContactPhone');
      if (subject && !subject.value) subject.value = '';
      if (message && !message.value) message.value = '';
      if (phone && !phone.value) phone.value = '';
      loadRecentHelpTickets(dialog, empresaID);
      dialog.classList.add('is-open');
      dialog.setAttribute('aria-hidden', 'false');
      window.setTimeout(function(){ if (subject) subject.focus(); }, 20);
    }

    function submitHelpTicket(dialog){
      var empresaID = getActiveEmpresaID();
      if (!empresaID) {
        setHelpTicketStatus(dialog, 'No pude detectar la empresa activa. Entra desde Administrar empresa e intenta de nuevo.', true);
        return;
      }
      var subject = dialog.querySelector('#pcsHelpTicketSubject');
      var category = dialog.querySelector('#pcsHelpTicketCategory');
      var priority = dialog.querySelector('#pcsHelpTicketPriority');
      var message = dialog.querySelector('#pcsHelpTicketMessage');
      var contactPreference = dialog.querySelector('#pcsHelpTicketContactPreference');
      var contactPhone = dialog.querySelector('#pcsHelpTicketContactPhone');
      var includeContext = dialog.querySelector('#pcsHelpTicketIncludeContext');
      var payload = {
        empresa_id: Number(empresaID),
        asunto: subject ? subject.value.trim() : '',
        categoria: category ? category.value : 'general',
        prioridad: priority ? priority.value : 'media',
        mensaje: message ? message.value.trim() : '',
        modulo: getActiveSystemModule(),
        ruta: getActiveSystemPath(),
        origen: 'menu_flotante',
        contacto_preferido: contactPreference ? contactPreference.value : 'email',
        contacto_telefono: contactPhone ? contactPhone.value.trim() : '',
        contexto: includeContext && includeContext.checked ? collectHelpTicketContext() : {}
      };
      if (!payload.asunto || !payload.mensaje) {
        setHelpTicketStatus(dialog, 'Completa el asunto y el mensaje para enviar el ticket.', true);
        return;
      }
      if (payload.mensaje.length < 12) {
        setHelpTicketStatus(dialog, 'Agrega un poco mas de detalle para que soporte pueda ayudarte mejor.', true);
        return;
      }
      setHelpTicketStatus(dialog, 'Enviando ticket...', false);
      setHelpTicketBusy(dialog, true);
      fetch('/api/empresa/tickets_ayuda?empresa_id=' + encodeURIComponent(empresaID), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'same-origin',
        body: JSON.stringify(payload)
      })
        .then(function(res){
          return res.json().catch(function(){ return {}; }).then(function(data){ return { ok: res.ok, status: res.status, data: data }; });
        })
        .then(function(result){
          if (!result.ok || !result.data || !result.data.ok) {
            throw new Error((result.data && result.data.error) || ('HTTP ' + result.status));
          }
          var code = result.data.ticket && result.data.ticket.codigo ? result.data.ticket.codigo : 'creado';
          setHelpTicketStatus(dialog, 'Ticket ' + code + ' enviado correctamente.', false);
          if (subject) subject.value = '';
          if (message) message.value = '';
          loadRecentHelpTickets(dialog, empresaID);
          window.setTimeout(function(){
            dialog.classList.remove('is-open');
            dialog.setAttribute('aria-hidden', 'true');
          }, 1200);
        })
        .catch(function(error){
          setHelpTicketStatus(dialog, error && error.message ? error.message : 'No se pudo enviar el ticket.', true);
        })
        .finally(function(){
          setHelpTicketBusy(dialog, false);
        });
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

    function openInsideAdminFrame(anchor, event){
      if (!anchor || !document.body || !document.body.classList.contains('admin-empresa-shell')) return false;
      var targetUrl = anchor.getAttribute('data-admin-frame-url') || '';
      if (!targetUrl) return false;
      var frame = document.querySelector('iframe[name="contentFrame"], #contentFrame');
      if (!frame) return false;
      event.preventDefault();
      frame.src = targetUrl;
      closeThemePopup();
      closeUtilitiesPopup();
      closePanel();
      return true;
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
        if (item && openInsideAdminFrame(item, event)) {
          return;
        }
        if (item && item.id !== 'themeToggle' && item.id !== 'utilitiesMenuToggle' && !item.classList.contains('fm-country')) {
          closeThemePopup();
          closeUtilitiesPopup();
          closePanel();
        }
      });

      document.addEventListener('click', function(event){
        var clickInsideMenu = wrapper.contains(event.target);
        if (!clickInsideMenu) {
          closeThemePopup();
          closeUtilitiesPopup();
          closePanel();
          return;
        }
        var clickInsideTheme = themeSelectorPopup && themeSelectorPopup.contains(event.target);
        var clickThemeButton = themeToggle && themeToggle.contains(event.target);
        if (!clickInsideTheme && !clickThemeButton) {
          closeThemePopup();
        }
        var clickInsideUtilities = utilitiesPopup && utilitiesPopup.contains(event.target);
        var clickUtilitiesButton = utilitiesToggle && utilitiesToggle.contains(event.target);
        if (!clickInsideUtilities && !clickUtilitiesButton) {
          closeUtilitiesPopup();
        }
      });

      document.addEventListener('keydown', function(event){
        if (event.key === 'Escape') {
          closeThemePopup();
          closeUtilitiesPopup();
          closePanel();
        }
      });
    }

    if (utilitiesToggle && utilitiesPopup) {
      utilitiesToggle.addEventListener('click', function(event){
        event.stopPropagation();
        var isExpanded = utilitiesToggle.getAttribute('aria-expanded') === 'true';
        utilitiesToggle.setAttribute('aria-expanded', isExpanded ? 'false' : 'true');
        utilitiesPopup.setAttribute('aria-hidden', isExpanded ? 'true' : 'false');
        utilitiesPopup.classList.toggle('show', !isExpanded);
        if (!isExpanded) {
          closeThemePopup();
        }
      });
    }

    if (calculatorLauncher) {
      calculatorLauncher.addEventListener('click', function(event){
        event.preventDefault();
        event.stopPropagation();
        closeThemePopup();
        closeUtilitiesPopup();
        closePanel();
        openCalculatorWindow();
      });
    }

    if (shareLaunchers && shareLaunchers.length) {
      for (var shareIndex = 0; shareIndex < shareLaunchers.length; shareIndex += 1) {
        shareLaunchers[shareIndex].addEventListener('click', function(event){
          event.preventDefault();
          event.stopPropagation();
          var channel = this.getAttribute('data-share-current') || 'email';
          closeThemePopup();
          closeUtilitiesPopup();
          closePanel();
          shareCurrentSystem(channel);
        });
      }
    }

    if (helpTicketLauncher) {
      helpTicketLauncher.addEventListener('click', function(event){
        event.preventDefault();
        event.stopPropagation();
        closeThemePopup();
        closeUtilitiesPopup();
        closePanel();
        openHelpTicketDialog();
      });
    }

    if (floatingAILauncher) {
      floatingAILauncher.addEventListener('click', function(event){
        event.preventDefault();
        event.stopPropagation();
        closeThemePopup();
        closeUtilitiesPopup();
        closePanel();
        openFloatingAI();
      });
    }

    if (floatingRadioLauncher) {
      floatingRadioLauncher.addEventListener('click', function(event){
        event.preventDefault();
        event.stopPropagation();
        closeThemePopup();
        closeUtilitiesPopup();
        closePanel();
        openFloatingRadio();
      });
    }

    if (floatingNotificationBell) {
      floatingNotificationBell.addEventListener('click', function(event){
        event.preventDefault();
        event.stopPropagation();
        closeThemePopup();
        closeUtilitiesPopup();
        openAdminNotificationsFromFloating();
      });
    }

    if (floatingNotificationRefresh) {
      floatingNotificationRefresh.addEventListener('click', function(event){
        event.preventDefault();
        event.stopPropagation();
        loadFloatingNotifications();
      });
    }

    window.addEventListener('pcs:notifications-updated', function(event){
      var detail = event && event.detail ? event.detail : {};
      setFloatingNotificationCount(detail.unread);
    });

    try {
      syncFloatingNotificationFromAdminBell();
      var observedAdminBadge = document.getElementById('adminNotificationBadge');
      var observedAdminBell = document.getElementById('adminNotificationBell');
      if (window.MutationObserver && (observedAdminBadge || observedAdminBell)) {
        var notificationObserver = new MutationObserver(function(){
          syncFloatingNotificationFromAdminBell();
        });
        if (observedAdminBadge) {
          notificationObserver.observe(observedAdminBadge, { childList: true, characterData: true, subtree: true });
        }
        if (observedAdminBell) {
          notificationObserver.observe(observedAdminBell, { attributes: true, attributeFilter: ['class'] });
        }
      }
    } catch (error) {}

    if (themeToggle && themeSelectorPopup) {
      themeToggle.addEventListener('click', function(event){
        event.stopPropagation();
        var isExpanded = themeToggle.getAttribute('aria-expanded') === 'true';
        themeToggle.setAttribute('aria-expanded', isExpanded ? 'false' : 'true');
        themeSelectorPopup.setAttribute('aria-hidden', isExpanded ? 'true' : 'false');
        themeSelectorPopup.classList.toggle('show', !isExpanded);
        if (!isExpanded) {
          closeUtilitiesPopup();
          fitPopupToViewport(themeSelectorPopup, themeToggle);
        }
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
      window.__pcsFloatingMenuPending = false;
      document.documentElement.dataset.pcsFloatingMenuInjected = '1';
      delete document.documentElement.dataset.pcsFloatingMenuPending;
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
            ensureFloatingMenuBadge();
            syncFloatingNotificationFromAdminBell();
          };
          toggle.appendChild(img);
          ensureFloatingMenuBadge();
          syncFloatingNotificationFromAdminBell();
          if (name) toggle.title = name;
        } catch (error) {}
      }

      function fallbackIcon(){
        try {
          if (!toggle) return;
          toggle.innerHTML = '<span class="fm-toggle-icon">☰</span>';
          ensureFloatingMenuBadge();
          syncFloatingNotificationFromAdminBell();
        } catch (error) {}
      }

      try {
        if (!hasBrowserSessionCookie()) {
          clearBrowserSessionStateCookie();
          removeFloatingMenu();
          return;
        }
        if (authData && typeof authData === 'object') {
          setSessionLinkAuthenticated(true);
          var photo = authData.photo || authData.avatar || '';
          var name = authData.name || authData.email || '';
          if (photo) setAvatarUrl(photo, name); else fallbackIcon();
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
            clearBrowserSessionStateCookie();
            removeFloatingMenu();
          });
      } catch (error) {
        clearBrowserSessionStateCookie();
        removeFloatingMenu();
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
