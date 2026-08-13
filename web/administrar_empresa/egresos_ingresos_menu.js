(function() {
      function parseEmpresaId(raw) {
        const n = Number(String(raw || '').trim());
        if (!Number.isFinite(n)) return '';
        const v = Math.trunc(n);
        return v > 0 ? String(v) : '';
      }

      window.__resolveEmpresaIdContext = function() {
        try {
          const params = new URLSearchParams(window.location.search || '');
          const own = parseEmpresaId(params.get('empresa_id') || params.get('id'));
          if (own) return own;
        } catch (_) {}
        try {
          let ctx = window.parent;
          let depth = 0;
          while (ctx && ctx !== window && depth < 4) {
            try {
              if (typeof ctx.__resolveEmpresaIdContext === 'function') {
                const resolved = parseEmpresaId(ctx.__resolveEmpresaIdContext());
                if (resolved) return resolved;
              }
            } catch (_) {}
            try {
              const parentParams = new URLSearchParams(ctx.location.search || '');
              const fromParent = parseEmpresaId(parentParams.get('empresa_id') || parentParams.get('id'));
              if (fromParent) return fromParent;
            } catch (_) {}
            try {
              if (!ctx.parent || ctx.parent === ctx) break;
              ctx = ctx.parent;
            } catch (_) {
              break;
            }
            depth += 1;
          }
        } catch (_) {}
        try {
          return parseEmpresaId(sessionStorage.getItem('active_empresa_id') || sessionStorage.getItem('empresa_id') || localStorage.getItem('active_empresa_id') || localStorage.getItem('empresa_id'));
        } catch (_) {
          return '';
        }
      };

      function withEmpresa(rawUrl) {
        try {
          const url = new URL(rawUrl, window.location.origin);
          const empresaId = window.__resolveEmpresaIdContext();
          if (empresaId) url.searchParams.set('empresa_id', empresaId);
          url.searchParams.set('v', '20260501-egresos-ingresos');
          return url.pathname + url.search + url.hash;
        } catch (_) {
          return rawUrl;
        }
      }

      document.querySelectorAll('#egresosIngresosSidebarNav a[href]').forEach(function(link) {
        link.setAttribute('href', withEmpresa(link.getAttribute('href')));
      });
      const frame = document.getElementById('egresosIngresosContentFrame');
      if (frame) frame.setAttribute('src', withEmpresa(frame.getAttribute('src') || '/administrar_empresa/egresos.html'));
    })();
