// menu.js - inyecta y gestiona el menú flotante centralizado
(function(){
  function injectMenu(){
    if (document.querySelector('.floating-menu')) return;
    const wrapper = document.createElement('div');
    wrapper.className = 'floating-menu';
    wrapper.setAttribute('aria-hidden','false');
    wrapper.innerHTML = '<button class="fm-toggle" aria-label="Abrir menú">☰</button>' +
      '<div class="fm-panel" role="menu">' +
        '<a class="fm-item" href="/index.html">Portal</a>' +
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
      toggle.addEventListener('click', function(e){ e.stopPropagation(); panel.classList.toggle('open'); });
      document.addEventListener('click', function(){ panel.classList.remove('open'); });
    }

    function getCookie(name){ const v = document.cookie.match('(^|;)\\s*'+name+'\\s*=\\s*([^;]+)'); return v ? v.pop() : ''; }

    const sessionLink = wrapper.querySelector('#sessionLink');
    if (sessionLink){
      if (getCookie('session_token')){
        sessionLink.textContent = 'Cerrar sesión';
        sessionLink.href = '/auth/logout';
        sessionLink.addEventListener('click', function(){ try { localStorage.removeItem('rememberAccount'); localStorage.removeItem('rememberedEmail'); } catch(e){} });
      } else {
        sessionLink.textContent = 'Iniciar sesión';
        sessionLink.href = '/login.html';
      }
    }

    // Delegación: asegurarse de limpiar rememberAccount aunque otros scripts no encontraran el enlace
    document.addEventListener('click', function(e){
      try{
        var a = e.target.closest && e.target.closest('.fm-item[href="/auth/logout"]');
        if (a) { try { localStorage.removeItem('rememberAccount'); localStorage.removeItem('rememberedEmail'); } catch(e){} }
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

    loadCountryItem();

    // Nota: no mostramos avatar/usuario en el menú flotante por simplicidad.
  }

  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', injectMenu); else injectMenu();
})();
