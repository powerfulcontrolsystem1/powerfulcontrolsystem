// menu.js - inyecta y gestiona el menú flotante centralizado
(function(){
  function injectMenu(){
    try {
      // Solo inyectar en la ventana top-level (evitar iframes/subframes)
      if (window.top !== window.self) return;
      // Evitar reinyecciones en cargas dinámicas: marca global y atributo DOM
      if (window.__pcsFloatingMenuInjected || document.documentElement.dataset.pcsFloatingMenuInjected === '1') return;
    } catch (e) {}
    // Ocultar menú flotante en la página principal (index)
    try {
      var _p = (window.location && window.location.pathname) ? (window.location.pathname || '').toLowerCase() : '';
      if (_p === '/' || _p === '/index.html' || _p.endsWith('/index.html')) {
        return;
      }
    } catch (e) {}
    if (document.querySelector('.floating-menu')) return;
    const wrapper = document.createElement('div');
    wrapper.className = 'floating-menu';
    wrapper.setAttribute('aria-hidden','false');
    wrapper.innerHTML = '<button class="fm-toggle" aria-label="Abrir menú">☰</button>' +
      '<div class="fm-panel" role="menu">' +
        '<a class="fm-item" href="/index.html">Portal</a>' +
        '<a class="fm-item" href="/ayuda/ayuda.html">Ayuda</a>' +
        '<a class="fm-item" href="/ultimas_mejoras.html">Últimas mejoras</a>' +
        '<a class="fm-item" href="/venta_digital.html">Venta digital</a>' +
          '<a class="fm-item" href="/Juegos/menu_juegos.html">Juegos</a>' +
        '<a id="calculatorLink" class="fm-item" href="/administrar_empresa/calculadora.html">Calculadora</a>' +
        '<button id="themeToggle" class="fm-item" type="button" aria-label="Cambiar tema"></button>' +
        '<div id="countryFlagItem" class="fm-item fm-country" style="display:none"></div>' +
        '<a id="sessionLink" class="fm-item" href="/login.html">Iniciar sesión</a>' +
      '</div>';
    // insertar al inicio del body
    if (document.body && document.body.firstChild) document.body.insertBefore(wrapper, document.body.firstChild);
    else if (document.body) document.body.appendChild(wrapper);

    const toggle = wrapper.querySelector('.fm-toggle');
    const panel = wrapper.querySelector('.fm-panel');
    if (toggle && panel) {
      function setPanelOpen(isOpen){
        panel.classList.toggle('open', !!isOpen);
        toggle.setAttribute('aria-expanded', isOpen ? 'true' : 'false');
      }

      function closePanel(){
        setPanelOpen(false);
      }

      toggle.setAttribute('aria-expanded', 'false');
      toggle.addEventListener('click', function(e){
        e.stopPropagation();
        setPanelOpen(!panel.classList.contains('open'));
      });
      panel.addEventListener('click', function(e){
        e.stopPropagation();
        var item = e.target.closest && e.target.closest('.fm-item');
        if (item && !item.classList.contains('fm-country')) {
          closePanel();
        }
      });
      document.addEventListener('click', closePanel);
      document.addEventListener('keydown', function(e){
        if (e.key === 'Escape') {
          closePanel();
        }
      });
    }

      var SESSION_STATE_COOKIE = 'browser_session_active';

    function getCookie(name){ const v = document.cookie.match('(^|;)\\s*'+name+'\\s*=\\s*([^;]+)'); return v ? v.pop() : ''; }

    function isPlausibleEmail(value){
      return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(String(value || '').trim());
    }

      function hasBrowserSessionCookie(){
        return getCookie(SESSION_STATE_COOKIE) === '1';
      }

    function clearRememberOnLogoutIfNeeded(){
      try {
        if (localStorage.getItem('rememberAccount') === '1') return;
        localStorage.removeItem('rememberAccount');
        localStorage.removeItem('rememberedEmail');
      } catch(e) {}
    }

    const sessionLink = wrapper.querySelector('#sessionLink');
      function setSessionLinkAuthenticated(isAuthenticated){
        if (!sessionLink) return;
        sessionLink.onclick = null;
        if (isAuthenticated) {
          sessionLink.textContent = 'Cerrar sesión';
          sessionLink.href = '/auth/logout';
          sessionLink.onclick = function(){ clearRememberOnLogoutIfNeeded(); };
          return;
        }
        sessionLink.textContent = 'Iniciar sesión';
        sessionLink.href = '/login.html';
      }
      setSessionLinkAuthenticated(hasBrowserSessionCookie());

    // Delegación: asegurarse de limpiar rememberAccount aunque otros scripts no encontraran el enlace
    document.addEventListener('click', function(e){
      try{
        var a = e.target.closest && e.target.closest('.fm-item[href="/auth/logout"]');
        if (a) { clearRememberOnLogoutIfNeeded(); }
      }catch(ee){}
    });

    // Theme toggle (iconos sol / luna)
    const themeToggle = wrapper.querySelector('#themeToggle');
    var currentTheme = (function(){ try{ return localStorage.getItem('theme') || 'dark' }catch(e){ return 'dark' } })();
    function applyTheme(t){ if (t === 'light') document.documentElement.classList.add('theme-light'); else document.documentElement.classList.remove('theme-light'); }
    function updateThemeBtn(){
      if (!themeToggle) return;
      var isLight = document.documentElement.classList.contains('theme-light');
      if (isLight) {
        // si estamos en tema claro, mostrar ícono luna (acción: activar modo oscuro)
        themeToggle.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="18" height="18" aria-hidden="true"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" fill="currentColor"/></svg>';
        themeToggle.setAttribute('aria-label','Activar modo oscuro');
      } else {
        // si estamos en tema oscuro, mostrar ícono sol (acción: activar modo claro)
        themeToggle.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="18" height="18" aria-hidden="true"><path d="M6.76 4.84l-1.8-1.79L3.17 4.84l1.79 1.79L6.76 4.84zM1 13h3v-2H1v2zm10 9h2v-3h-2v3zm7-1.76l1.79 1.79 1.79-1.79-1.79-1.79-1.79 1.79zM17.24 4.84l1.79-1.79L17.24 1.26 15.45 3.05 17.24 4.84zM12 7a5 5 0 1 0 0 10 5 5 0 0 0 0-10z" fill="currentColor"/></svg>';
        themeToggle.setAttribute('aria-label','Activar modo claro');
      }
    }
    applyTheme(currentTheme); updateThemeBtn();
    if (themeToggle) themeToggle.addEventListener('click', function(){
      currentTheme = document.documentElement.classList.contains('theme-light') ? 'dark' : 'light';
      try{ localStorage.setItem('theme', currentTheme) }catch(e){}
      applyTheme(currentTheme);
      updateThemeBtn();
    });

    // (Modo ventana eliminado por petición)

    function detectCountryFromBrowserSignals(){
      var tz = '';
      try { tz = (Intl.DateTimeFormat().resolvedOptions().timeZone || '').toLowerCase(); } catch(e) { tz = ''; }
      var lang = ((navigator && navigator.language) ? navigator.language : '').toLowerCase();
      if (tz.indexOf('panama') >= 0 || lang.indexOf('es-pa') === 0) return { code:'PA', name:'Panamá', flag:'🇵🇦', source:'navegador' };
      if (tz.indexOf('guayaquil') >= 0 || tz.indexOf('quito') >= 0 || lang.indexOf('es-ec') === 0) return { code:'EC', name:'Ecuador', flag:'🇪🇨', source:'navegador' };
      return { code:'CO', name:'Colombia', flag:'🇨🇴', source:'navegador' };
    }

    function resolveEmpresaId(){
      try {
        var p = new URLSearchParams(window.location.search || '');
        var id = p.get('empresa_id') || p.get('id') || '';
        if (id) {
          try { localStorage.setItem('active_empresa_id', String(id)); } catch(e) {}
          return id;
        }
        return localStorage.getItem('active_empresa_id') || '';
      } catch(e) {
        try { return localStorage.getItem('active_empresa_id') || ''; } catch(ee) { return ''; }
      }
    }

    function updateCalculatorLink(){
      var calcLink = wrapper.querySelector('#calculatorLink');
      if (!calcLink) return;
      var empresaId = resolveEmpresaId();
      var url = new URL('/administrar_empresa/calculadora.html', window.location.origin);
      if (empresaId) {
        url.searchParams.set('empresa_id', empresaId);
      }
      calcLink.setAttribute('href', url.pathname + url.search);
    }

    function renderCountryItem(info){
      var item = wrapper.querySelector('#countryFlagItem');
      if (!item || !info) return;
      var label = (info.flag || '🌐') + ' ' + (info.name || info.code || 'País');
      item.textContent = 'País: ' + label;
      item.title = 'Detección: ' + (info.source || 'desconocida');
      item.style.display = '';
    }

    function loadCountryItem(){
      var empresaId = resolveEmpresaId();
      var tz = '';
      try { tz = Intl.DateTimeFormat().resolvedOptions().timeZone || ''; } catch(e) { tz = ''; }
      var lang = (navigator && navigator.language) ? navigator.language : '';

      if (!empresaId) {
        renderCountryItem(detectCountryFromBrowserSignals());
        return;
      }

      var url = '/api/empresa/facturacion_electronica/pais_detectado?empresa_id=' + encodeURIComponent(empresaId) + '&tz=' + encodeURIComponent(tz) + '&lang=' + encodeURIComponent(lang);
      fetch(url, { credentials: 'same-origin' })
        .then(function(res){
          if (!res.ok) throw new Error('HTTP ' + res.status);
          return res.json();
        })
        .then(function(data){
          renderCountryItem({
            code: data.pais_codigo || '',
            name: data.pais_nombre || '',
            flag: data.bandera || '🌐',
            source: data.source || 'api'
          });
        })
        .catch(function(){
          renderCountryItem(detectCountryFromBrowserSignals());
        });
    }

    updateCalculatorLink();
    loadCountryItem();

    // Marcar inyección para evitar duplicados en futuras cargas dinámicas
    try {
      window.__pcsFloatingMenuInjected = true;
      document.documentElement.dataset.pcsFloatingMenuInjected = '1';
    } catch (e) {}

    // Cargar foto de perfil desde /me y usarla como icono del botón (fallback a símbolo)
    (function loadAvatar(){
      function setAvatarUrl(url, name){
        try {
          if (!toggle) return;
          toggle.innerHTML = '';
          var img = document.createElement('img');
          img.className = 'fm-avatar';
          img.src = url;
          img.alt = name || 'Perfil';
          img.onerror = function(){ toggle.innerHTML = '<span class="fm-toggle-icon">☰</span>'; };
          toggle.appendChild(img);
          if (name) toggle.title = name;
        } catch (e) {}
      }

      function fallbackIcon(){ try { if (!toggle) return; toggle.innerHTML = '<span class="fm-toggle-icon">☰</span>'; } catch(e){} }

      function syncRememberedEmailFromProfile(data){
        try {
          if (!data) return;
          if (localStorage.getItem('rememberAccount') !== '1') return;
          var email = String(data.email || '').trim();
          if (!isPlausibleEmail(email)) return;
          var remembered = String(localStorage.getItem('rememberedEmail') || '').trim();
          if (!remembered || !isPlausibleEmail(remembered)) {
            localStorage.setItem('rememberedEmail', email);
          }
        } catch (e) {}
      }

      try {
          if (!hasBrowserSessionCookie()) {
            setSessionLinkAuthenticated(false);
          fallbackIcon();
          return;
        }
        fetch('/me', { credentials: 'same-origin' }).then(function(res){
          if (!res.ok) throw new Error('no-auth');
          return res.json();
        }).then(function(data){
            setSessionLinkAuthenticated(true);
          syncRememberedEmailFromProfile(data);
          var photo = (data && (data.photo || data.avatar)) || '';
          var name = (data && (data.name || data.email)) || '';
          if (photo) setAvatarUrl(photo, name); else fallbackIcon();
          }).catch(function(){ setSessionLinkAuthenticated(false); fallbackIcon(); });
        } catch (e) { setSessionLinkAuthenticated(false); fallbackIcon(); }
    })();

    // Icon fallback: asegurar que todo img.icon tenga una fuente válida
    (function ensureIconFallback(){
      var defaultIcon = '/img/analytics-color.svg';
      function setFallback(img){
        try{
          if(!img) return;
          if(!img.getAttribute('src') || img.getAttribute('src').trim() === '') img.src = defaultIcon;
          img.addEventListener('error', function(){ if(img.src !== defaultIcon) img.src = defaultIcon; });
        }catch(e){}
      }
      // aplicar inicialmente
      document.querySelectorAll && document.querySelectorAll('.admin-sidebar .nav a img.icon').forEach(setFallback);
      // observar cambios dinámicos en caso de que menús se modifiquen
      try{
        var obs = new MutationObserver(function(mutations){
          mutations.forEach(function(m){
            m.addedNodes && m.addedNodes.forEach(function(node){
              if(node && node.querySelectorAll){ node.querySelectorAll('.admin-sidebar .nav a img.icon').forEach(setFallback); }
              if(node && node.matches && node.matches('.admin-sidebar .nav a img.icon')) setFallback(node);
            });
          });
        });
        obs.observe(document.body, { childList:true, subtree:true });
      }catch(e){}
    })();

    // Ahora mostramos avatar (si está disponible) en el botón del menú flotante.
  }

  function initAcceptModalFromQuery(){
    // La aceptación de contrato se gestiona ahora en /accept.html.
    // Esta función se mantiene como no-op para compatibilidad con versiones antiguas.
    return;
  }

  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', function(){ injectMenu(); initAcceptModalFromQuery(); }); else { injectMenu(); initAcceptModalFromQuery(); }
})();
