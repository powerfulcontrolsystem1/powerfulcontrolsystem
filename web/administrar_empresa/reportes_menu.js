(function(){
      function applySubpageFromQuery() {
        try {
          var params = new URLSearchParams(window.location.search);
          var sub = params.get('subpage');
          if (!sub) return;
          var iframe = document.getElementById('reportesContentFrame');
          if (!iframe) return;
          var src = sub.indexOf('/') === 0 ? sub : ('/administrar_empresa/' + sub);
          if (window.__empresaModuleGuard && typeof window.__empresaModuleGuard.withEmpresa === 'function') {
            src = window.__empresaModuleGuard.withEmpresa(src);
          }
          iframe.src = src;
        } catch (_) {}
      }
      if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', applySubpageFromQuery);
      } else {
        applySubpageFromQuery();
      }
    })();
