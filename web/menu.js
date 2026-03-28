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
        '<button id="themeToggle" class="fm-item" type="button">Modo claro</button>' +
        '<button id="openWindowBtn" class="fm-item" type="button">Abrir como ventana</button>' +
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
        sessionLink.addEventListener('click', function(){ try { localStorage.removeItem('rememberAccount'); } catch(e){} });
      } else {
        sessionLink.textContent = 'Iniciar sesión';
        sessionLink.href = '/login.html';
      }
    }

    // Delegación: asegurarse de limpiar rememberAccount aunque otros scripts no encontraran el enlace
    document.addEventListener('click', function(e){
      try{
        var a = e.target.closest && e.target.closest('.fm-item[href="/auth/logout"]');
        if (a) { try { localStorage.removeItem('rememberAccount'); } catch(e){} }
      }catch(ee){}
    });

    // Theme toggle
    const themeToggle = wrapper.querySelector('#themeToggle');
    const openWindowBtn = wrapper.querySelector('#openWindowBtn');
    var currentTheme = (function(){ try{ return localStorage.getItem('theme') || 'dark' }catch(e){ return 'dark' } })();
    function applyTheme(t){ if (t === 'light') document.documentElement.classList.add('theme-light'); else document.documentElement.classList.remove('theme-light'); }
    function updateThemeBtn(){ if (!themeToggle) return; themeToggle.textContent = document.documentElement.classList.contains('theme-light') ? 'Modo oscuro' : 'Modo claro'; }
    applyTheme(currentTheme); updateThemeBtn();
    if (themeToggle) themeToggle.addEventListener('click', function(){ currentTheme = document.documentElement.classList.contains('theme-light') ? 'dark' : 'light'; try{ localStorage.setItem('theme', currentTheme) }catch(e){} applyTheme(currentTheme); updateThemeBtn(); });

    if (openWindowBtn) openWindowBtn.addEventListener('click', function(){
      var url = window.location.href;
      var w = Math.min(screen.width - 40, 1200); var h = Math.min(screen.height - 60, 900);
      var feats = 'toolbar=no,location=no,directories=no,status=no,menubar=no,scrollbars=yes,resizable=yes,top=10,left=10,width='+w+',height='+h;
      var nw = window.open(url, '_blank', feats);
      if (!nw) alert('El navegador bloqueó la apertura de la ventana sin barra de direcciones. Permite ventanas emergentes o instala como aplicación (PWA).');
    });
  }

  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', injectMenu); else injectMenu();
})();
