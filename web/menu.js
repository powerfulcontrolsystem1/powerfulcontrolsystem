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

      // Nota: no mostramos avatar/usuario en el menú flotante por simplicidad.
  }

  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', injectMenu); else injectMenu();
})();
