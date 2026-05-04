(function () {
  'use strict';

  function ready(fn) {
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', fn, { once: true });
    } else {
      fn();
    }
  }

  function safePath(value) {
    value = String(value || '').trim();
    if (!value || value.charAt(0) !== '/') return '';
    if (value.indexOf('//') === 0) return '';
    return value;
  }

  function fallbackTarget() {
    var params = new URLSearchParams(window.location.search || '');
    var origin = safePath(params.get('origen'));
    if (origin) return origin;
    if (window.location.pathname.indexOf('/ayuda/ayuda.html') === -1) return '/ayuda/ayuda.html';
    return '/index.html';
  }

  function goBack() {
    if (window.history && window.history.length > 1) {
      window.history.back();
      return;
    }
    window.location.href = fallbackTarget();
  }

  ready(function () {
    if (document.querySelector('.help-back-bar')) return;
    var bar = document.createElement('div');
    bar.className = 'help-back-bar';

    var button = document.createElement('button');
    button.type = 'button';
    button.className = 'btn secondary help-back-btn';
    button.textContent = 'Atras';
    button.setAttribute('aria-label', 'Volver a la pantalla anterior');
    button.addEventListener('click', goBack);

    bar.appendChild(button);
    document.body.insertBefore(bar, document.body.firstChild);
  });
})();
