(function() {
  'use strict';

  var HELP_URL = '/ayuda/ayuda_contextual.html';
  var mountedAttr = 'data-context-help-mounted';
  var targetAttr = 'data-context-help-target';
  var pageMounted = false;
  var mutationTimer = 0;

  function ready(fn) {
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', fn, { once: true });
    } else {
      fn();
    }
  }

  function normalize(value) {
    return String(value == null ? '' : value).replace(/\s+/g, ' ').trim();
  }

  function slugify(value) {
    return normalize(value)
      .toLowerCase()
      .normalize('NFD')
      .replace(/[\u0300-\u036f]/g, '')
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-+|-+$/g, '')
      .slice(0, 80) || 'seccion';
  }

  function pageTitle() {
    var h1 = document.querySelector('h1');
    return normalize(h1 && h1.textContent) || normalize(document.title) || 'Pagina';
  }

  function originPath() {
    return window.location.pathname + window.location.search;
  }

  function buildHelpHref(sectionTitle, kind, element) {
    var params = new URLSearchParams();
    params.set('pagina', pageTitle());
    params.set('seccion', sectionTitle || pageTitle());
    params.set('tipo', kind || 'seccion');
    params.set('origen', originPath());
    if (element && element.id) params.set('id', element.id);
    params.set('clave', slugify((window.location.pathname || '') + ' ' + (sectionTitle || '')));
    return HELP_URL + '?' + params.toString();
  }

  function createHelpLink(sectionTitle, kind, element) {
    var link = document.createElement('a');
    link.className = 'context-help-link context-help-link--' + (kind || 'section');
    link.href = buildHelpHref(sectionTitle, kind, element);
    link.target = '_self';
    link.rel = 'help';
    link.textContent = '?';
    link.title = 'Ayuda: ' + (sectionTitle || pageTitle());
    link.setAttribute('aria-label', 'Abrir ayuda de ' + (sectionTitle || pageTitle()));
    return link;
  }

  function directHeadingText(element) {
    if (!element) return '';
    var selectors = [
      ':scope > .empresa-section-header h1',
      ':scope > .empresa-section-header h2',
      ':scope > .empresa-section-header h3',
      ':scope > .page-header h1',
      ':scope > .page-header h2',
      ':scope > header h1',
      ':scope > header h2',
      ':scope > h1',
      ':scope > h2',
      ':scope > h3',
      ':scope > .section-heading'
    ];
    for (var i = 0; i < selectors.length; i += 1) {
      try {
        var node = element.querySelector(selectors[i]);
        var text = normalize(node && node.textContent);
        if (text) return text;
      } catch (_) {}
    }
    var labelled = normalize(element.getAttribute('aria-label') || element.getAttribute('data-help-title') || element.getAttribute('title'));
    if (labelled) return labelled;
    var id = normalize(element.id);
    if (id) return id.replace(/[-_]+/g, ' ');
    return '';
  }

  function shouldSkip(element) {
    if (!element || element.nodeType !== 1) return true;
    if (element.hasAttribute(mountedAttr)) return true;
    if (element.getAttribute('data-context-help') === 'off') return true;
    if (element.closest('[data-context-help="off"]')) return true;
    if (element.closest('.floating-menu, .ai-chat-drawer, .context-help-link, script, style')) return true;
    return false;
  }

  function findHeaderHost(element) {
    if (!element) return null;
    var selectors = [
      ':scope > .empresa-section-header',
      ':scope > .page-header',
      ':scope > .card-header',
      ':scope > .electric-header',
      ':scope > .cart-electric-header',
      ':scope > header',
      ':scope > .device-title-row'
    ];
    for (var i = 0; i < selectors.length; i += 1) {
      try {
        var host = element.querySelector(selectors[i]);
        if (host) return host;
      } catch (_) {}
    }
    return null;
  }

  function addToPageHeader() {
    if (pageMounted) return;
    var host = document.querySelector('.page-header, .electric-header, .cart-electric-header, .help-hero, main > header, .container > h1');
    if (!host || shouldSkip(host)) return;
    var title = pageTitle();
    var link = createHelpLink(title, 'page', host);
    host.classList.add('context-help-host', 'context-help-page-host');
    host.appendChild(link);
    host.setAttribute(mountedAttr, '1');
    pageMounted = true;
  }

  function addToTarget(element) {
    if (shouldSkip(element)) return;
    var title = directHeadingText(element);
    if (!title) return;
    var host = findHeaderHost(element);
    var link = createHelpLink(title, 'section', element);
    element.setAttribute(mountedAttr, '1');
    element.setAttribute(targetAttr, '1');
    if (host && !host.querySelector(':scope > .context-help-link')) {
      host.classList.add('context-help-host');
      host.appendChild(link);
      return;
    }
    element.classList.add('context-help-floating-host');
    link.classList.add('context-help-link--floating');
    element.appendChild(link);
  }

  function scan() {
    addToPageHeader();
    var selector = [
      '.card',
      '.empresa-section',
      'main > section',
      '.container > section',
      'article.portal-card',
      'article.home-offer-card',
      'article.device-card',
      '.advanced-card',
      '.panel-soft'
    ].join(',');
    var nodes = document.querySelectorAll(selector);
    for (var i = 0; i < nodes.length; i += 1) {
      addToTarget(nodes[i]);
    }
  }

  function scheduleScan() {
    window.clearTimeout(mutationTimer);
    mutationTimer = window.setTimeout(scan, 180);
  }

  ready(function() {
    if (document.body && document.body.getAttribute('data-context-help') === 'off') return;
    scan();
    if (window.MutationObserver) {
      var observer = new MutationObserver(function(records) {
        for (var i = 0; i < records.length; i += 1) {
          if (records[i].addedNodes && records[i].addedNodes.length) {
            scheduleScan();
            return;
          }
        }
      });
      observer.observe(document.body || document.documentElement, { childList: true, subtree: true });
    }
  });
})();
