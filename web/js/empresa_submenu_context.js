(function() {
  'use strict';

  function parsePositiveInt(raw) {
    const n = Number(String(raw || '').trim());
    if (!Number.isFinite(n)) return 0;
    const v = Math.trunc(n);
    return v > 0 ? v : 0;
  }

  function resolveEmpresaId() {
    try {
      const params = new URLSearchParams(window.location.search || '');
      const own = parsePositiveInt(params.get('empresa_id') || params.get('id'));
      if (own > 0) return own;
    } catch (_) {}

    try {
      let ctx = window.parent;
      let depth = 0;
      while (ctx && ctx !== window && depth < 5) {
        try {
          if (typeof ctx.__resolveEmpresaIdContext === 'function') {
            const resolved = parsePositiveInt(ctx.__resolveEmpresaIdContext());
            if (resolved > 0) return resolved;
          }
        } catch (_) {}
        try {
          const parentParams = new URLSearchParams(ctx.location.search || '');
          const fromParent = parsePositiveInt(parentParams.get('empresa_id') || parentParams.get('id'));
          if (fromParent > 0) return fromParent;
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
      const candidates = [
        sessionStorage.getItem('active_empresa_id'),
        sessionStorage.getItem('empresa_id'),
        localStorage.getItem('active_empresa_id'),
        localStorage.getItem('empresa_id')
      ];
      for (let i = 0; i < candidates.length; i += 1) {
        const parsed = parsePositiveInt(candidates[i]);
        if (parsed > 0) return parsed;
      }
    } catch (_) {}
    return 0;
  }

  window.__resolveEmpresaIdContext = window.__resolveEmpresaIdContext || resolveEmpresaId;

  function withEmpresa(rawUrl) {
    const empresaId = resolveEmpresaId();
    if (!empresaId) return rawUrl;
    try {
      const url = new URL(rawUrl, window.location.origin);
      if (url.origin === window.location.origin && url.pathname.indexOf('/administrar_empresa/') === 0) {
        url.searchParams.set('empresa_id', String(empresaId));
      }
      return url.pathname + url.search + url.hash;
    } catch (_) {
      return rawUrl;
    }
  }

  function applyEmpresaContext() {
    document.querySelectorAll('a[target$="ContentFrame"][href^="/administrar_empresa/"]').forEach(function(link) {
      link.setAttribute('href', withEmpresa(link.getAttribute('href') || ''));
    });
    document.querySelectorAll('iframe.admin-empresa-frame[src^="/administrar_empresa/"]').forEach(function(frame) {
      frame.setAttribute('src', withEmpresa(frame.getAttribute('src') || ''));
    });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', applyEmpresaContext);
  } else {
    applyEmpresaContext();
  }
})();
