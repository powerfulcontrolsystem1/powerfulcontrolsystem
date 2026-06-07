(function() {
  'use strict';

  function normalizeTheme(value) {
    const allowed = {
      light: true,
      'light-rose': true,
      'light-gold': true,
      dark: true,
      'dark-violet': true,
      'dark-emerald': true,
      'dark-corporate': true,
      'dark-absolute': true,
      'dark-obsidian': true,
      'dark-neon': true,
      'light-wood': true
    };
    let theme = String(value || '').trim().toLowerCase();
    if (theme === 'dark-protect') theme = 'dark';
    return allowed[theme] ? theme : 'light';
  }

  function readCookieTheme() {
    const match = String(document.cookie || '').match(/(?:^|;\s*)pcs_theme=([^;]+)/);
    return match ? decodeURIComponent(match[1] || '') : '';
  }

  function resolveTheme() {
    try {
      if (window.parent && window.parent !== window) {
        const parentTheme = window.parent.document.documentElement.getAttribute('data-theme') || '';
        if (parentTheme) return normalizeTheme(parentTheme);
      }
    } catch (_) {}
    try {
      return normalizeTheme(window.localStorage.getItem('theme') || readCookieTheme());
    } catch (_) {
      return normalizeTheme(readCookieTheme());
    }
  }

  function applyThemeContext() {
    const theme = resolveTheme();
    const root = document.documentElement;
    root.setAttribute('data-theme', theme);
    root.classList.toggle('theme-light', theme.indexOf('light') === 0);
    root.classList.toggle('theme-dark', theme.indexOf('light') !== 0);
  }

  function applySubmenuModeContext() {
    let submenu = false;
    try {
      submenu = (new URLSearchParams(window.location.search || '')).get('submenu') === '1';
    } catch (_) {}
    document.documentElement.classList.toggle('empresa-module-submenu-content', submenu);
    if (document.body) document.body.classList.toggle('empresa-module-submenu-content', submenu);
  }

  applySubmenuModeContext();
  applyThemeContext();

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

  function isEmpresaModulePage() {
    const body = document.body;
    if (!body) return false;
    return body.classList.contains('empresa-subpage')
      || body.classList.contains('modulo-colombia-page')
      || body.classList.contains('admin-subpage');
  }

  function safeMessage(raw) {
    const text = String(raw || '').replace(/\s+/g, ' ').trim();
    if (!text) return '';
    return text.length > 220 ? text.slice(0, 217) + '...' : text;
  }

  function findModuleMount() {
    return document.getElementById('moduloColombiaApp')
      || document.querySelector('main.container')
      || document.querySelector('main')
      || document.querySelector('.container')
      || document.body;
  }

  function renderMissingEmpresaContext() {
    if (!isEmpresaModulePage() || resolveEmpresaId() > 0) return;
    if (document.getElementById('empresaContextWarning')) return;
    const mount = findModuleMount();
    if (!mount) return;

    const warning = document.createElement('section');
    warning.id = 'empresaContextWarning';
    warning.className = 'empresa-context-warning';
    warning.setAttribute('role', 'status');
    warning.innerHTML = [
      '<div>',
      '<span class="empresa-context-warning-kicker">Contexto requerido</span>',
      '<h2>Selecciona una empresa para continuar</h2>',
      '<p>Este modulo necesita una empresa activa para cargar permisos, datos y operaciones sin inconsistencias.</p>',
      '</div>',
      '<div class="empresa-context-warning-actions">',
      '<a class="btn primary" href="/seleccionar_empresa.html" target="_top">Seleccionar empresa</a>',
      '<button class="btn secondary" type="button" data-empresa-context-retry>Reintentar</button>',
      '</div>'
    ].join('');
    const retry = warning.querySelector('[data-empresa-context-retry]');
    if (retry) {
      retry.addEventListener('click', function() {
        window.location.reload();
      });
    }
    mount.insertBefore(warning, mount.firstChild || null);
  }

  function showRuntimeAlert(title, detail) {
    if (!isEmpresaModulePage()) return;
    const old = document.getElementById('empresaRuntimeAlert');
    if (old && old.parentNode) old.parentNode.removeChild(old);

    const alert = document.createElement('aside');
    alert.id = 'empresaRuntimeAlert';
    alert.className = 'empresa-runtime-alert';
    alert.setAttribute('role', 'alert');
    alert.innerHTML = [
      '<div>',
      '<strong></strong>',
      '<p></p>',
      '</div>',
      '<div class="empresa-runtime-alert-actions">',
      '<button class="btn secondary small" type="button" data-runtime-retry>Reintentar</button>',
      '<button class="empresa-runtime-close" type="button" aria-label="Cerrar aviso" data-runtime-close>&times;</button>',
      '</div>'
    ].join('');
    const strong = alert.querySelector('strong');
    const paragraph = alert.querySelector('p');
    if (strong) strong.textContent = safeMessage(title) || 'No se pudo completar la operacion';
    if (paragraph) paragraph.textContent = safeMessage(detail) || 'Revisa la conexion o intenta recargar este modulo.';
    const close = alert.querySelector('[data-runtime-close]');
    const retry = alert.querySelector('[data-runtime-retry]');
    if (close) {
      close.addEventListener('click', function() {
        if (alert.parentNode) alert.parentNode.removeChild(alert);
      });
    }
    if (retry) {
      retry.addEventListener('click', function() {
        window.location.reload();
      });
    }
    document.body.appendChild(alert);
  }

  let connectivityHideTimer = null;
  const CONNECTIVITY_AUDIT_QUEUE_PREFIX = 'pcs_connectivity_audit_queue_';
  const CONNECTIVITY_LAST_STATE_KEY = 'pcs_connectivity_last_state';

  function clearConnectivityHideTimer() {
    if (connectivityHideTimer) {
      clearTimeout(connectivityHideTimer);
      connectivityHideTimer = null;
    }
  }

  function isTopLevelWindow() {
    try {
      return window.top === window;
    } catch (_) {
      return true;
    }
  }

  function connectivityAuditQueueKey() {
    const empresaId = resolveEmpresaId();
    return CONNECTIVITY_AUDIT_QUEUE_PREFIX + (empresaId > 0 ? String(empresaId) : 'sin_empresa');
  }

  function readConnectivityAuditQueue() {
    try {
      const raw = localStorage.getItem(connectivityAuditQueueKey()) || '[]';
      const parsed = JSON.parse(raw);
      return Array.isArray(parsed) ? parsed : [];
    } catch (_) {
      return [];
    }
  }

  function writeConnectivityAuditQueue(items) {
    try {
      localStorage.setItem(connectivityAuditQueueKey(), JSON.stringify(items.slice(-40)));
    } catch (_) {}
  }

  function queueConnectivityAuditEvent(eventData) {
    if (!isTopLevelWindow()) return;
    const queue = readConnectivityAuditQueue();
    queue.push(eventData);
    writeConnectivityAuditQueue(queue);
  }

  function readConnectivityLastState() {
    try {
      return sessionStorage.getItem(CONNECTIVITY_LAST_STATE_KEY) || '';
    } catch (_) {
      return '';
    }
  }

  function writeConnectivityLastState(state) {
    try {
      sessionStorage.setItem(CONNECTIVITY_LAST_STATE_KEY, state);
    } catch (_) {}
  }

  function normalizeConnectivityContext(raw) {
    if (!raw || typeof raw !== 'object') return null;
    return {
      canContinueOffline: !!(raw.canContinueOffline || raw.offlineSalesEnabled || raw.facturacionOfflineHabilitada),
      module: String(raw.module || raw.modulo || '').trim(),
      label: String(raw.label || raw.etiqueta || '').trim()
    };
  }

  function readConnectivityContextFrom(win) {
    if (!win) return null;
    try {
      if (typeof win.__PCSConnectivityContext === 'function') {
        return normalizeConnectivityContext(win.__PCSConnectivityContext());
      }
      return normalizeConnectivityContext(win.__PCSConnectivityContext);
    } catch (_) {
      return null;
    }
  }

  function resolveConnectivityContext() {
    const own = readConnectivityContextFrom(window);
    if (own) return own;
    try {
      const active = normalizeConnectivityContext(window.__PCSActiveModuleConnectivityContext);
      if (active) return active;
    } catch (_) {}
    try {
      const frame = document.getElementById('contentFrame') || document.querySelector('iframe.admin-empresa-frame');
      if (frame && frame.contentWindow) {
        const child = readConnectivityContextFrom(frame.contentWindow);
        if (child) return child;
      }
    } catch (_) {}
    return { canContinueOffline: false, module: '', label: '' };
  }

  function connectivityAlertCopy(online) {
    const ctx = resolveConnectivityContext();
    if (online) {
      return {
        title: 'Conexion a internet restablecida',
        detail: ctx.canContinueOffline
          ? 'El sistema volvio a estar en linea. Se sincronizaran las ventas pendientes y puedes continuar.'
          : 'El sistema volvio a estar en linea. Ya puedes continuar con este modulo.'
      };
    }
    if (ctx.canContinueOffline) {
      return {
        title: 'Modo sin internet para caja activo',
        detail: 'Puedes seguir vendiendo en esta caja. Las ventas se imprimen provisionalmente y se sincronizan cuando vuelva internet.'
      };
    }
    return {
      title: 'Se perdio la conexion a internet',
      detail: 'Este modulo necesita internet para operar. Por favor espere a que vuelva la conexion; manten esta ventana abierta.'
    };
  }

  function buildConnectivityAuditEvent(online) {
    return {
      estado: online ? 'online' : 'offline',
      online: !!online,
      fecha_evento: new Date().toISOString(),
      evento_id: String(Date.now()) + '-' + Math.random().toString(36).slice(2, 10),
      origen: 'browser',
      path: String(window.location.pathname || '').slice(0, 220),
      pending_count: readConnectivityAuditQueue().length
    };
  }

  function sendConnectivityAuditEvent(eventData) {
    const empresaId = resolveEmpresaId();
    if (!empresaId || typeof fetch !== 'function') {
      return Promise.reject(new Error('empresa_id no disponible'));
    }
    return fetch('/api/empresa/auditoria/eventos?empresa_id=' + encodeURIComponent(String(empresaId)) + '&action=conexion', {
      method: 'POST',
      credentials: 'same-origin',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(eventData || {})
    }).then(function(response) {
      if (!response.ok) throw new Error('auditoria conectividad HTTP ' + response.status);
      return response;
    });
  }

  function flushConnectivityAuditQueue() {
    if (!isTopLevelWindow() || navigator.onLine === false) return Promise.resolve(false);
    const queue = readConnectivityAuditQueue();
    if (!queue.length) return Promise.resolve(true);
    let chain = Promise.resolve();
    const sent = [];
    queue.forEach(function(item) {
      chain = chain.then(function() {
        return sendConnectivityAuditEvent(item).then(function() {
          sent.push(item);
        });
      });
    });
    return chain.then(function() {
      writeConnectivityAuditQueue([]);
      return true;
    }).catch(function() {
      const remaining = queue.slice(sent.length);
      writeConnectivityAuditQueue(remaining);
      return false;
    });
  }

  function registerConnectivityState(online) {
    if (!isTopLevelWindow()) return;
    const state = online ? 'online' : 'offline';
    const previous = readConnectivityLastState();
    if (previous === state) {
      if (online) flushConnectivityAuditQueue();
      return;
    }
    writeConnectivityLastState(state);
    queueConnectivityAuditEvent(buildConnectivityAuditEvent(online));
    if (online) flushConnectivityAuditQueue();
  }

  function renderConnectivityAlert(online) {
    if (!document.body) return;
    clearConnectivityHideTimer();

    const old = document.getElementById('empresaConnectivityAlert');
    if (old && old.parentNode) old.parentNode.removeChild(old);

    const alert = document.createElement('aside');
    alert.id = 'empresaConnectivityAlert';
    alert.className = 'empresa-runtime-alert empresa-connectivity-alert ' + (online ? 'is-online' : 'is-offline');
    alert.setAttribute('role', 'alert');
    alert.setAttribute('aria-live', 'assertive');
    alert.innerHTML = [
      '<div>',
      '<strong></strong>',
      '<p></p>',
      '</div>',
      '<div class="empresa-runtime-alert-actions">',
      '<button class="empresa-runtime-close" type="button" aria-label="Cerrar aviso" data-connectivity-close>&times;</button>',
      '</div>'
    ].join('');

    const strong = alert.querySelector('strong');
    const paragraph = alert.querySelector('p');
    const copy = connectivityAlertCopy(online);
    if (strong) strong.textContent = copy.title;
    if (paragraph) paragraph.textContent = copy.detail;

    const close = alert.querySelector('[data-connectivity-close]');
    if (close) {
      close.addEventListener('click', function() {
        clearConnectivityHideTimer();
        if (alert.parentNode) alert.parentNode.removeChild(alert);
      });
    }
    document.body.appendChild(alert);

    if (online) {
      connectivityHideTimer = setTimeout(function() {
        if (alert.parentNode) alert.parentNode.removeChild(alert);
      }, 8000);
    }
  }

  function syncConnectivityState() {
    if (navigator.onLine === false) {
      renderConnectivityAlert(false);
      registerConnectivityState(false);
    } else {
      flushConnectivityAuditQueue();
    }
  }

  function installConnectivityMonitor() {
    if (window.__empresaConnectivityMonitorInstalled) return;
    window.__empresaConnectivityMonitorInstalled = true;
    window.addEventListener('offline', function() {
      renderConnectivityAlert(false);
      registerConnectivityState(false);
    });
    window.addEventListener('online', function() {
      renderConnectivityAlert(true);
      registerConnectivityState(true);
    });
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', syncConnectivityState);
    } else {
      syncConnectivityState();
    }
  }

  window.__empresaModuleGuard = window.__empresaModuleGuard || {};
  window.__empresaModuleGuard.resolveEmpresaId = window.__empresaModuleGuard.resolveEmpresaId || resolveEmpresaId;
  window.__empresaModuleGuard.withEmpresa = window.__empresaModuleGuard.withEmpresa || withEmpresa;
  window.__empresaModuleGuard.applyThemeContext = window.__empresaModuleGuard.applyThemeContext || applyThemeContext;
  window.__empresaModuleGuard.applyEmpresaContext = window.__empresaModuleGuard.applyEmpresaContext || applyEmpresaContext;
  window.__empresaModuleGuard.showRuntimeAlert = window.__empresaModuleGuard.showRuntimeAlert || showRuntimeAlert;
  window.__empresaModuleGuard.showConnectivityAlert = renderConnectivityAlert;
  window.__empresaModuleGuard.refreshConnectivityAlert = function() {
    if (document.getElementById('empresaConnectivityAlert') && navigator.onLine === false) {
      renderConnectivityAlert(false);
    }
  };

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
    document.querySelectorAll('a[href^="/administrar_empresa/"]').forEach(function(link) {
      link.setAttribute('href', withEmpresa(link.getAttribute('href') || ''));
    });
    document.querySelectorAll('iframe.admin-empresa-frame[src^="/administrar_empresa/"]').forEach(function(frame) {
      frame.setAttribute('src', withEmpresa(frame.getAttribute('src') || ''));
    });
  }

  function bootModuleGuard() {
    applySubmenuModeContext();
    applyThemeContext();
    applyEmpresaContext();
    renderMissingEmpresaContext();
    ensureHelpAIBridge();
  }

  function ensureHelpAIBridge() {
    try {
      if (window.__pcsHelpAIBridgeInstalled) return;
      if (document.querySelector('script[data-pcs-help-ai-bridge]')) return;
      const script = document.createElement('script');
      script.src = '/js/help_ai_bridge.js?v=20260607-help-ai4';
      script.defer = true;
      script.dataset.pcsHelpAiBridge = '1';
      (document.head || document.documentElement).appendChild(script);
    } catch (_) {}
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', function() {
      bootModuleGuard();
    });
  } else {
    bootModuleGuard();
  }

  window.addEventListener('storage', applyThemeContext);
  window.addEventListener('pageshow', applyThemeContext);
  installConnectivityMonitor();
  window.addEventListener('error', function(event) {
    const message = event && (event.message || (event.error && event.error.message));
    showRuntimeAlert('Error inesperado en el modulo', message);
  });
  window.addEventListener('unhandledrejection', function(event) {
    const reason = event && event.reason;
    const message = reason && (reason.message || reason.statusText || reason.toString && reason.toString());
    showRuntimeAlert('Operacion incompleta', message);
  });
})();
