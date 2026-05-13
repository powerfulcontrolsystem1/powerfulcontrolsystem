// menu.js - inyecta y gestiona el menú flotante centralizado
(function(){
  var SESSION_STATE_COOKIE = 'browser_session_active';
  var THEME_STORAGE_KEY = 'theme';
  var THEME_COOKIE_NAME = 'pcs_theme';
  var THEME_VALUES = {
    dark: true,
    'dark-violet': true,
    'dark-emerald': true,
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
    wrapper.innerHTML = '<button class="fm-toggle" aria-label="Abrir menú"><span class="fm-toggle-icon">☰</span></button>' +
      '<div class="fm-panel" role="menu">' +
        '<a class="fm-item" href="/index.html">Portal</a>' +
        '<a class="fm-item" href="/red_social_comercial.html" target="_blank" rel="noopener">Red social comercial</a>' +
        '<button id="createHelpTicketLink" class="fm-item fm-action-item" type="button">Crear ticket de ayuda</button>' +
        '<div class="fm-submenu" id="utilitiesMenuWrapper">' +
          '<button id="utilitiesMenuToggle" class="fm-item fm-submenu-toggle" type="button" aria-expanded="false" aria-haspopup="true">Utilidades \u25BC</button>' +
          '<div id="utilitiesMenuPopup" class="fm-submenu-popup" aria-hidden="true" role="menu">' +
            '<a class="fm-item fm-subitem" href="/calculadora.html?compact=1" data-open-calculator="1">Calculadora</a>' +
            '<a class="fm-item fm-subitem" href="/Juegos/menu_juegos.html" target="_blank" rel="noopener">Juegos</a>' +
            '<a class="fm-item fm-subitem" href="/emulador/" target="_blank" rel="noopener">Emulador</a>' +
          '</div>' +
        '</div>' +
        '<a class="fm-item" href="/configuracion_de_la_cuenta.html">Configuración de la cuenta</a>' +
        '' +
        '<div class="theme-selector-item" id="themeToggleWrapper" style="position:relative;">' +
          '<button id="themeToggle" class="fm-item theme-toggle-btn" type="button" aria-expanded="false" aria-haspopup="true" aria-label="Cambiar apariencia">Cambiar apariencia \u25BC</button>' +
          '<div id="themeSelectorPopup" class="theme-selector-popup" aria-hidden="true" role="menu">' +
            '<div class="theme-opt-group">Oscuros</div>' +
            '<button class="theme-option" type="button" data-theme-value="dark">Azul Elegante</button>' +
            '<button class="theme-option" type="button" data-theme-value="dark-violet">Morado Midnight</button>' +
            '<button class="theme-option" type="button" data-theme-value="dark-emerald">Negro Esmeralda</button>' +
            '<button class="theme-option" type="button" data-theme-value="dark-neon">Neon Nocturno</button>' +
            '<div class="theme-opt-group mt-1">Claros</div>' +
            '<button class="theme-option" type="button" data-theme-value="light">Blanco Corporativo</button>' +
            '<button class="theme-option" type="button" data-theme-value="light-rose">Rosa Pastel</button>' +
            '<button class="theme-option" type="button" data-theme-value="light-gold">Blanco Dorado</button>' +
            '<button class="theme-option" type="button" data-theme-value="light-wood">Madera Clara</button>' +
          '</div>' +
        '</div>' +
        '<a id="sessionLink" class="fm-item" href="/login.html">Iniciar sesión</a>' +
        '<a id="adminHelpLink" class="fm-item" href="/ayuda/ayuda.html" hidden>Ayuda administrador</a>' +
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
    var adminHelpLink = wrapper.querySelector('#adminHelpLink');
    var themeToggle = wrapper.querySelector('#themeToggle');
    var themeSelectorPopup = wrapper.querySelector('#themeSelectorPopup');
    var utilitiesToggle = wrapper.querySelector('#utilitiesMenuToggle');
    var utilitiesPopup = wrapper.querySelector('#utilitiesMenuPopup');
    var calculatorLauncher = wrapper.querySelector('[data-open-calculator]');
    var helpTicketLauncher = wrapper.querySelector('#createHelpTicketLink');

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

    function closeUtilitiesPopup(){
      if (!utilitiesToggle || !utilitiesPopup) return;
      utilitiesToggle.setAttribute('aria-expanded', 'false');
      utilitiesPopup.setAttribute('aria-hidden', 'true');
      utilitiesPopup.classList.remove('show');
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
            '<div><h2 id="pcsHelpTicketTitle">Crear ticket de ayuda</h2><p>El equipo de soporte recibira tu solicitud con la pagina donde estas trabajando.</p></div>' +
            '<button id="pcsHelpTicketClose" class="pcs-help-ticket-close" type="button" aria-label="Cerrar">x</button>' +
          '</div>' +
          '<form id="pcsHelpTicketForm" class="pcs-help-ticket-form">' +
            '<label class="form-label" for="pcsHelpTicketSubject">Asunto</label>' +
            '<input id="pcsHelpTicketSubject" class="form-input" type="text" maxlength="180" required placeholder="Ej: No puedo cerrar caja">' +
            '<div class="pcs-help-ticket-grid">' +
              '<label class="form-col"><span class="form-label">Categoria</span><select id="pcsHelpTicketCategory" class="form-input"><option value="general">General</option><option value="tecnico">Tecnico</option><option value="operacion">Operacion</option><option value="configuracion">Configuracion</option><option value="facturacion">Facturacion</option><option value="pagos">Pagos</option><option value="licencias">Licencias</option><option value="usuarios">Usuarios</option><option value="seguridad">Seguridad</option></select></label>' +
              '<label class="form-col"><span class="form-label">Prioridad</span><select id="pcsHelpTicketPriority" class="form-input"><option value="media">Media</option><option value="baja">Baja</option><option value="alta">Alta</option><option value="critica">Critica</option></select></label>' +
            '</div>' +
            '<label class="form-label" for="pcsHelpTicketMessage">Mensaje</label>' +
            '<textarea id="pcsHelpTicketMessage" class="form-textarea" maxlength="4000" required placeholder="Describe que intentabas hacer, que viste y que necesitas resolver."></textarea>' +
            '<div id="pcsHelpTicketStatus" class="pcs-help-ticket-status" aria-live="polite"></div>' +
            '<div class="pcs-help-ticket-actions"><button type="button" id="pcsHelpTicketCancel" class="btn secondary">Cancelar</button><button type="submit" class="btn primary">Enviar ticket</button></div>' +
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

    function openHelpTicketDialog(){
      var dialog = ensureHelpTicketDialog();
      var empresaID = getActiveEmpresaID();
      setHelpTicketStatus(dialog, empresaID ? '' : 'Abre una empresa antes de crear el ticket para asociarlo correctamente.', !empresaID);
      var subject = dialog.querySelector('#pcsHelpTicketSubject');
      var message = dialog.querySelector('#pcsHelpTicketMessage');
      if (subject && !subject.value) subject.value = '';
      if (message && !message.value) message.value = '';
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
      var payload = {
        empresa_id: Number(empresaID),
        asunto: subject ? subject.value.trim() : '',
        categoria: category ? category.value : 'general',
        prioridad: priority ? priority.value : 'media',
        mensaje: message ? message.value.trim() : '',
        modulo: getActiveSystemModule(),
        ruta: getActiveSystemPath(),
        origen: 'menu_flotante'
      };
      if (!payload.asunto || !payload.mensaje) {
        setHelpTicketStatus(dialog, 'Completa el asunto y el mensaje para enviar el ticket.', true);
        return;
      }
      setHelpTicketStatus(dialog, 'Enviando ticket...', false);
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
          window.setTimeout(function(){
            dialog.classList.remove('is-open');
            dialog.setAttribute('aria-hidden', 'true');
          }, 1200);
        })
        .catch(function(error){
          setHelpTicketStatus(dialog, error && error.message ? error.message : 'No se pudo enviar el ticket.', true);
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

    if (themeToggle && themeSelectorPopup) {
      themeToggle.addEventListener('click', function(event){
        event.stopPropagation();
        var isExpanded = themeToggle.getAttribute('aria-expanded') === 'true';
        themeToggle.setAttribute('aria-expanded', isExpanded ? 'false' : 'true');
        themeSelectorPopup.setAttribute('aria-hidden', isExpanded ? 'true' : 'false');
        themeSelectorPopup.classList.toggle('show', !isExpanded);
        if (!isExpanded) {
          closeUtilitiesPopup();
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
          clearBrowserSessionStateCookie();
          removeFloatingMenu();
          return;
        }
        if (authData && typeof authData === 'object') {
          setSessionLinkAuthenticated(true);
          if (adminHelpLink) {
            adminHelpLink.hidden = String(authData.role || '').trim().toLowerCase() !== 'super_administrador';
          }
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
            if (adminHelpLink) {
              adminHelpLink.hidden = !(data && String(data.role || '').trim().toLowerCase() === 'super_administrador');
            }
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
