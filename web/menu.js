// menu.js - inyecta y gestiona el menú flotante centralizado
(function(){
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
      // Solo inyectar en la ventana top-level (evitar iframes/subframes)
      if (window.top !== window.self) return;
      // Evitar reinyecciones en cargas dinámicas: marca global y atributo DOM
      if (window.__pcsFloatingMenuInjected || document.documentElement.dataset.pcsFloatingMenuInjected === '1') return;
    } catch (e) {}
    // Ocultar menú flotante en la página principal (index) y en páginas de login.
    try {
      var _p = (window.location && window.location.pathname) ? (window.location.pathname || '').toLowerCase() : '';
      if (_p === '/' || _p === '/index.html' || _p.endsWith('/index.html')) {
        return;
      }
      // No mostrar en páginas de login (login de administradores / login_usuario)
      if (_p === '/login.html' || _p.endsWith('/login.html') || _p === '/login_usuario.html' || _p.endsWith('/login_usuario.html')) {
        return;
      }
      // Solo inyectar menú flotante cuando exista cookie de sesión activa
      var cookieStr = String(document.cookie || '');
      if (cookieStr.indexOf('browser_session_active=1') === -1) {
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
        '<a class="fm-item" href="/red_social_comercial.html">Red social comercial</a>' +
        '<a class="fm-item" href="/configuracion_de_la_cuenta.html">ConfiguraciÃ³n de la cuenta</a>' +
        '<div class="theme-selector-item" id="themeToggleWrapper" style="position:relative;">' +
          '<button id="themeToggle" class="fm-item theme-toggle-btn" type="button" aria-expanded="false" aria-haspopup="true" aria-label="Cambiar apariencia">Apariencia</button>' +
          '<div id="themeSelectorPopup" class="theme-selector-popup" aria-hidden="true" role="menu">' +
            '<div class="theme-opt-group">Oscuros</div>' +
            '<button class="theme-option" data-theme-value="dark">Azul Elegante</button>' +
            '<button class="theme-option" data-theme-value="dark-violet">Morado Midnight</button>' +
            '<button class="theme-option" data-theme-value="dark-emerald">Negro Esmeralda</button>' +
            '<div class="theme-opt-group mt-1">Claros</div>' +
            '<button class="theme-option" data-theme-value="light">Blanco Corporativo</button>' +
            '<button class="theme-option" data-theme-value="light-rose">Rosa Pastel</button>' +
            '<button class="theme-option" data-theme-value="light-gold">Blanco Dorado</button>' +
          '</div>' +
        '</div>' +
        '<a id="sessionLink" class="fm-item" href="/login.html">Iniciar sesión</a>' +
        '<a class="fm-item" href="/ayuda/ayuda.html">Ayuda</a>' +
      '</div>';
    // insertar al inicio del body
    if (document.body && document.body.firstChild) document.body.insertBefore(wrapper, document.body.firstChild);
    else if (document.body) document.body.appendChild(wrapper);
    try {
      if (document.body) document.body.classList.add('has-floating-menu');
      document.documentElement.classList.add('has-floating-menu');
    } catch (e) {}

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
      // Asegurar cierre del panel en interacción directa con cada item (touch/click)
      try {
        if (panel.querySelectorAll) {
          panel.querySelectorAll('.fm-item').forEach(function(it){
            it.addEventListener('click', function(ev){
              try{ closePanel(); }catch(e){}
            });
          });
        }
      } catch(e) {}
      document.addEventListener('click', closePanel);
      document.addEventListener('keydown', function(e){
        if (e.key === 'Escape') {
          closePanel();
        }
      });
    }

      var SESSION_STATE_COOKIE = 'browser_session_active';

    function getCookie(name){ const v = document.cookie.match('(^|;)\\s*'+name+'\\s*=\\s*([^;]+)'); return v ? v.pop() : ''; }

      function hasBrowserSessionCookie(){
        return getCookie(SESSION_STATE_COOKIE) === '1';
      }

    const sessionLink = wrapper.querySelector('#sessionLink');
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

    // Theme toggle Logic
    const themeToggleWrapper = wrapper.querySelector('#themeToggleWrapper');
    const themeToggle = wrapper.querySelector('#themeToggle');
    const themeSelectorPopup = wrapper.querySelector('#themeSelectorPopup');
    
    var currentTheme = (function(){ try{ return localStorage.getItem('theme') || 'dark' }catch(e){ return 'dark' } })();
    
    function applyTheme(t){
      // Apply theme attribute
      if (t === 'dark') {
          // If 'dark' is default we can just set it empty or explicitly to 'dark' for css to match html[data-theme="dark"]
          document.documentElement.setAttribute('data-theme', t);
      } else {
          document.documentElement.setAttribute('data-theme', t);
      }
      
      // Update label
      if (themeToggle) {
          themeToggle.innerHTML = 'Apariencia';
      }
      
      // Update active state in popup
      if (themeSelectorPopup) {
         themeSelectorPopup.querySelectorAll('.theme-option').forEach(opt => {
            if (opt.getAttribute('data-theme-value') === t) {
               opt.classList.add('active');
            } else {
               opt.classList.remove('active');
            }
         });
      }
    }
    
    applyTheme(currentTheme); 
    
    if (themeToggleWrapper && themeToggle && themeSelectorPopup) {
        themeToggle.addEventListener('click', function(e) {
            e.stopPropagation();
            const isExpanded = themeToggle.getAttribute('aria-expanded') === 'true';
            
            // Toggle visibility
            themeToggle.setAttribute('aria-expanded', !isExpanded);
            themeSelectorPopup.setAttribute('aria-hidden', isExpanded);
            if (!isExpanded) {
                themeSelectorPopup.classList.add('show');
            } else {
                themeSelectorPopup.classList.remove('show');
            }
        });
        
        themeSelectorPopup.querySelectorAll('.theme-option').forEach(opt => {
            opt.addEventListener('click', function(e) {
                e.stopPropagation();
                const selectedTheme = this.getAttribute('data-theme-value');
                currentTheme = selectedTheme;
                try{ localStorage.setItem('theme', selectedTheme) }catch(e){}
                applyTheme(selectedTheme);
                
                // Close popup
                themeToggle.setAttribute('aria-expanded', 'false');
                themeSelectorPopup.setAttribute('aria-hidden', 'true');
                themeSelectorPopup.classList.remove('show');
                
                // Optional: completely close menu immediately after picking theme? 
                // Or leave menu open? Usually better to leave it open so they see result.
            });
        });
        
        // click outside popup to close it (handled by panel event listener mostly, but let's ensure it doesn't leave the popup stranding if menu closes)
        document.addEventListener('click', function() {
            themeToggle.setAttribute('aria-expanded', 'false');
            themeSelectorPopup.setAttribute('aria-hidden', 'true');
            themeSelectorPopup.classList.remove('show');
        });
        
        // stop propagation inside popup so clicking it doesn't close the parent panel
        themeSelectorPopup.addEventListener('click', function(e) {
           e.stopPropagation(); 
        });
    }

    // (Modo ventana eliminado por petición)

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
    // El portal publico deja un unico acceso directo al emulador N64.
    // No se muestran accesos adicionales para experiencias retiradas.
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

  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', function(){ injectMenu(); initAcceptModalFromQuery(); initAdminMobileSidebarToggle(); }); else { injectMenu(); initAcceptModalFromQuery(); initAdminMobileSidebarToggle(); }
})();
