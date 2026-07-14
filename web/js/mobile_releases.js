(function () {
  'use strict';

  function safeOfficialUrl(raw) {
    if (!raw || typeof raw !== 'string') return '';
    try {
      var url = new URL(raw, window.location.origin);
      if (url.protocol !== 'https:') return '';
      var allowedHosts = [window.location.hostname, 'play.google.com', 'apps.apple.com', 'testflight.apple.com'];
      if (allowedHosts.indexOf(url.hostname) === -1) return '';
      return url.href;
    } catch (_) { return ''; }
  }

  function setText(root, selector, value) {
    var node = root.querySelector(selector);
    if (node) node.textContent = value || '';
  }

  function setLink(root, selector, url, label, disabledLabel) {
    var node = root.querySelector(selector);
    if (!node) return;
    if (url) {
      node.href = url;
      node.hidden = false;
      node.removeAttribute('aria-disabled');
      node.textContent = label;
    } else {
      node.removeAttribute('href');
      node.hidden = false;
      node.setAttribute('aria-disabled', 'true');
      node.textContent = disabledLabel;
    }
  }

  function render(releases) {
    document.querySelectorAll('[data-mobile-release-card]').forEach(function (root) {
      var platform = root.getAttribute('data-mobile-release-card');
      var item = releases && releases[platform] ? releases[platform] : {};
      var available = item.available === true;
      setText(root, '[data-mobile-release-version]', item.version || 'Próximamente');
      setText(root, '[data-mobile-release-requirements]', item.minimum_requirements || 'Pendiente de publicación');
      setText(root, '[data-mobile-release-published]', item.published_at ? ('Publicado: ' + item.published_at) : 'Sin versión pública todavía');
      setText(root, '[data-mobile-release-checksum]', item.sha256 ? ('SHA-256: ' + item.sha256) : 'Checksum disponible con la versión oficial');
      setText(root, '[data-mobile-release-instructions]', item.instructions || 'La publicación oficial se anunciará aquí.');
      var primary = platform === 'android' ? safeOfficialUrl(item.download_url) : safeOfficialUrl(item.app_store_url);
      var label = platform === 'android' ? 'Descargar aplicación oficial' : 'Abrir App Store';
      setLink(root, '[data-mobile-release-primary]', available ? primary : '', label, 'Próximamente');
      var secondary = platform === 'android' ? safeOfficialUrl(item.play_store_url) : safeOfficialUrl(item.testflight_url);
      setLink(root, '[data-mobile-release-secondary]', secondary, platform === 'android' ? 'Ver en Google Play' : 'Abrir TestFlight', 'No disponible');
      root.dataset.available = available ? 'true' : 'false';
    });
  }

  fetch('/assets/data/mobile-releases.json', { cache: 'no-store', credentials: 'same-origin' })
    .then(function (response) { if (!response.ok) throw new Error('release metadata unavailable'); return response.json(); })
    .then(render)
    .catch(function () { render({}); });
})();
