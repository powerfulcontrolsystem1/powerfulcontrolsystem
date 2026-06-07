(function () {
  'use strict';

  var HELP_MESSAGE_TYPE = 'pcs-help-ai-open';
  var INSTALLED_FLAG = '__pcsHelpAIBridgeInstalled';
  var BRIDGE_SRC = '/js/help_ai_bridge.js?v=20260607-help-ai4';

  if (window[INSTALLED_FLAG]) return;
  window[INSTALLED_FLAG] = true;

  function normalize(value) {
    return String(value == null ? '' : value).replace(/\s+/g, ' ').trim();
  }

  function parsePositiveInt(raw) {
    var n = Number(String(raw || '').trim());
    if (!Number.isFinite(n)) return 0;
    n = Math.trunc(n);
    return n > 0 ? n : 0;
  }

  function resolveEmpresaId() {
    try {
      if (typeof window.__resolveEmpresaIdContext === 'function') {
        var resolved = parsePositiveInt(window.__resolveEmpresaIdContext());
        if (resolved > 0) return resolved;
      }
    } catch (error) {}

    try {
      var params = new URLSearchParams(window.location.search || '');
      var own = parsePositiveInt(params.get('empresa_id') || params.get('id'));
      if (own > 0) return own;
    } catch (error) {}

    try {
      if (window.parent && window.parent !== window && typeof window.parent.__resolveEmpresaIdContext === 'function') {
        var parentResolved = parsePositiveInt(window.parent.__resolveEmpresaIdContext());
        if (parentResolved > 0) return parentResolved;
      }
    } catch (error) {}

    var keys = ['active_empresa_id', 'empresa_id', 'admin_empresa_id'];
    var stores = [];
    try { stores.push(window.sessionStorage); } catch (error) {}
    try { stores.push(window.localStorage); } catch (error) {}
    for (var s = 0; s < stores.length; s += 1) {
      var store = stores[s];
      if (!store) continue;
      for (var i = 0; i < keys.length; i += 1) {
        try {
          var parsed = parsePositiveInt(store.getItem(keys[i]));
          if (parsed > 0) return parsed;
        } catch (error) {}
      }
    }
    return 0;
  }

  function isModifiedClick(event) {
    return !!(event && (event.ctrlKey || event.metaKey || event.shiftKey || event.altKey || event.button === 1));
  }

  function isHelpHref(href) {
    var value = normalize(href).toLowerCase();
    if (!value) return false;
    if (value.charAt(0) === '#') return false;
    return value.indexOf('/ayuda/') >= 0
      || value.indexOf('ayuda_contextual.html') >= 0
      || value.indexOf('tutorial_nomina.html') >= 0
      || value.indexOf('facturacion_electronica_tutorial_dian.html') >= 0;
  }

  function pointsToNamedFrame(link) {
    if (!link) return false;
    var target = normalize(link.getAttribute('target'));
    if (!target || target === '_self' || target === '_top' || target === '_parent' || target === '_blank') return false;
    try {
      return !!document.querySelector('iframe[name="' + CSS.escape(target) + '"]');
    } catch (error) {
      return !!document.querySelector('iframe[name="' + target.replace(/"/g, '\\"') + '"]');
    }
  }

  function getClosestHelpLauncher(target) {
    if (!target || !target.closest) return null;
    var explicit = target.closest('[data-pcs-help-ai]');
    if (explicit) return explicit;
    var link = target.closest('a[href]');
    if (link && !pointsToNamedFrame(link) && isHelpHref(link.getAttribute('href') || link.href || '')) return link;
    var button = target.closest('button[data-help-url],button[data-help-href]');
    return button || null;
  }

  function readMainHeading() {
    var selectors = [
      '.empresa-module-header h1',
      '.empresa-section-header h1',
      '.empresa-section-header h2',
      'main h1',
      'h1',
      'h2'
    ];
    for (var i = 0; i < selectors.length; i += 1) {
      var el = document.querySelector(selectors[i]);
      var text = normalize(el && el.textContent);
      if (text) return text;
    }
    return normalize(document.title) || 'Pantalla actual';
  }

  function readSectionTitle(source) {
    try {
      var section = source && source.closest && source.closest('.empresa-section,.card,section,article');
      if (section) {
        var heading = section.querySelector('h1,h2,h3,[data-help-section-title]');
        var sectionText = normalize(heading && heading.textContent);
        if (sectionText) return sectionText;
      }
    } catch (error) {}
    return readMainHeading();
  }

  function resolveHelpUrl(source) {
    var raw = '';
    if (source) {
      raw = source.getAttribute('data-help-url')
        || source.getAttribute('data-help-href')
        || source.getAttribute('href')
        || '';
    }
    if (!raw) raw = '/ayuda/ayuda_contextual.html';
    try {
      var url = new URL(raw, window.location.origin);
      var empresaId = resolveEmpresaId();
      if (empresaId > 0 && url.origin === window.location.origin && !url.searchParams.get('empresa_id')) {
        url.searchParams.set('empresa_id', String(empresaId));
      }
      return url.pathname + url.search + url.hash;
    } catch (error) {
      return raw;
    }
  }

  function buildPayload(source) {
    var helpUrl = resolveHelpUrl(source);
    var pageTitle = normalize((source && source.getAttribute('data-help-page')) || readMainHeading());
    var sectionTitle = normalize((source && source.getAttribute('data-help-section')) || readSectionTitle(source));
    var detail = normalize(source && source.getAttribute('data-help-detail'));
    var prompt = normalize(source && source.getAttribute('data-help-prompt'));
    var type = normalize(source && source.getAttribute('data-help-type')) || 'ayuda';
    var empresaId = resolveEmpresaId();
    var originPath = '';
    try {
      originPath = window.location.pathname + window.location.search + window.location.hash;
    } catch (error) {}
    return {
      type: type,
      title: sectionTitle || pageTitle,
      page: pageTitle,
      section: sectionTitle,
      detail: detail,
      prompt: prompt,
      helpUrl: helpUrl,
      origin: originPath,
      empresa_id: empresaId > 0 ? empresaId : null
    };
  }

  function postToParent(payload) {
    try {
      if (!window.parent || window.parent === window) return false;
      if (typeof window.parent.PCSAIChatHelp === 'function') {
        return window.parent.PCSAIChatHelp(payload) !== false;
      }
      window.parent.postMessage({ type: HELP_MESSAGE_TYPE, payload: payload }, window.location.origin);
      return true;
    } catch (error) {
      return false;
    }
  }

  function openLocally(payload) {
    try {
      if (typeof window.PCSAIChatHelp === 'function') {
        return window.PCSAIChatHelp(payload) !== false;
      }
    } catch (error) {}
    try {
      if (typeof window.PCSAIChatOpen === 'function') {
        return window.PCSAIChatOpen({
          mode: 'ayudante',
          preferRobot: true,
          prompt: payload.prompt || ('Ayudame con ' + (payload.section || payload.page || 'esta pantalla') + '.')
        }) !== false;
      }
    } catch (error) {}
    return false;
  }

  function openHelpFallback(payload) {
    var url = normalize(payload && payload.helpUrl) || '/ayuda/ayuda_contextual.html';
    try {
      if (window.top && window.top !== window) {
        window.open(url, '_blank', 'noopener,noreferrer');
      } else {
        window.location.href = url;
      }
    } catch (error) {
      window.location.href = url;
    }
  }

  function requestHelp(payload) {
    payload = payload || buildPayload(null);
    if (openLocally(payload)) return true;
    if (postToParent(payload)) return true;
    return false;
  }

  function handleClick(event) {
    if (isModifiedClick(event)) return;
    var launcher = getClosestHelpLauncher(event.target);
    if (!launcher) return;
    if (launcher.getAttribute('data-pcs-help-ai') === 'off') return;
    var payload = buildPayload(launcher);
    if (requestHelp(payload)) {
      event.preventDefault();
      event.stopPropagation();
      return;
    }
    if (launcher.hasAttribute('data-pcs-help-ai')) {
      event.preventDefault();
      event.stopPropagation();
      openHelpFallback(payload);
    }
  }

  function injectBridgeIntoFrame(frame) {
    if (!frame) return;
    try {
      var doc = frame.contentDocument;
      var childWindow = frame.contentWindow;
      if (!doc || !childWindow) return;
      if (childWindow[INSTALLED_FLAG]) return;
      if (doc.querySelector('script[data-pcs-help-ai-bridge]')) return;
      var script = doc.createElement('script');
      script.src = BRIDGE_SRC;
      script.defer = true;
      script.dataset.pcsHelpAiBridge = '1';
      (doc.head || doc.documentElement).appendChild(script);
    } catch (error) {}
  }

  function scanFramesForBridge() {
    var frames = [];
    try {
      frames = Array.prototype.slice.call(document.querySelectorAll('iframe'));
    } catch (error) {
      frames = [];
    }
    frames.forEach(function (frame) {
      injectBridgeIntoFrame(frame);
      if (!frame.dataset.pcsHelpAiBridgeLoadBound) {
        frame.dataset.pcsHelpAiBridgeLoadBound = '1';
        frame.addEventListener('load', function () {
          injectBridgeIntoFrame(frame);
        });
      }
    });
  }

  function startFrameBridgeScanner() {
    scanFramesForBridge();
    var attempts = 0;
    var timer = window.setInterval(function () {
      attempts += 1;
      scanFramesForBridge();
      if (attempts >= 12) {
        window.clearInterval(timer);
      }
    }, 800);
    try {
      var observer = new MutationObserver(function () {
        scanFramesForBridge();
      });
      observer.observe(document.documentElement || document.body, { childList: true, subtree: true });
      window.setTimeout(function () {
        try { observer.disconnect(); } catch (error) {}
      }, 15000);
    } catch (error) {}
  }

  window.PCSHelpAI = {
    buildPayload: buildPayload,
    request: requestHelp,
    fallback: openHelpFallback
  };

  document.addEventListener('click', handleClick, true);
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', startFrameBridgeScanner);
  } else {
    startFrameBridgeScanner();
  }
})();
