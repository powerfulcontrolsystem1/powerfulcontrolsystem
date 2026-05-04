(function() {
  'use strict';

  function removeHelpIcons() {
    try {
      document.querySelectorAll('.context-help-link, .login-help-icon').forEach(function(node) {
        node.remove();
      });
    } catch (_) {}
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', removeHelpIcons, { once: true });
  } else {
    removeHelpIcons();
  }
})();
