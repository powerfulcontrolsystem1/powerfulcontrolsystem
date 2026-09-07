(function (global) {
  'use strict';

  var existing = global.PCS_PUBLIC_CONFIG && typeof global.PCS_PUBLIC_CONFIG === 'object'
    ? global.PCS_PUBLIC_CONFIG
    : {};

  var recaptcha = existing.recaptcha && typeof existing.recaptcha === 'object'
    ? existing.recaptcha
    : {};

  global.PCS_PUBLIC_CONFIG = Object.assign({}, existing, {
    recaptcha: Object.assign({
      enabled: false,
      siteKey: '',
      provider: 'google-recaptcha-v2',
      devBypass: false
    }, recaptcha)
  });

  global.RECAPTCHA_ENABLED = Boolean(global.PCS_PUBLIC_CONFIG.recaptcha.enabled);
  global.RECAPTCHA_SITE_KEY = String(global.PCS_PUBLIC_CONFIG.recaptcha.siteKey || '');
  global.RECAPTCHA_PROVIDER = String(global.PCS_PUBLIC_CONFIG.recaptcha.provider || 'google-recaptcha-v2');
  global.RECAPTCHA_DEV_BYPASS = Boolean(global.PCS_PUBLIC_CONFIG.recaptcha.devBypass);
})(window);
