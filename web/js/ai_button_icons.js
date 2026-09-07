(function () {
  "use strict";

  var ICON_URL = "/img/pcs_ia_logo.svg";
  var IA_TEXT = /\b(ia|i\.a\.|inteligencia artificial|gpt|chatgpt)\b/i;
  var IA_ACTIONS = /(analizar|generar|extraer|diagnosticar|consultar|preparar|borrador|asistente|centro|cargar|leer|buscar|capturar).*(ia|gpt)|(ia|gpt).*(analizar|generar|extraer|diagnosticar|consultar|preparar|borrador|asistente|centro|cargar|leer|buscar|capturar)/i;
  var STYLE_ID = "pcs-ai-button-icon-styles";

  function ensureAIButtonStyles(doc) {
    doc = doc || document;
    if (!doc || doc.getElementById(STYLE_ID)) return;
    var style = doc.createElement("style");
    style.id = STYLE_ID;
    style.textContent = [
      ".ai-action-button{display:inline-flex!important;align-items:center!important;justify-content:center!important;gap:8px!important;min-height:44px!important;padding:10px 14px!important;font-weight:900!important;line-height:1.15!important;}",
      ".ai-action-button .ai-button-icon,.ai-action-button img[data-ai-logo='true']{width:18px!important;height:18px!important;object-fit:contain!important;flex:0 0 auto!important;}",
      ".ai-action-button.capture-btn,.capture-btn.ai-action-button{min-height:48px!important;padding:11px 15px!important;}",
      ".ai-action-button.small,.ai-action-button.btn-sm{min-height:38px!important;padding:8px 11px!important;}",
      ".ai-action-button.small .ai-button-icon,.ai-action-button.btn-sm .ai-button-icon{width:15px!important;height:15px!important;}"
    ].join("");
    (doc.head || doc.documentElement).appendChild(style);
  }

  function textOf(el) {
    return String((el && (el.textContent || el.getAttribute("aria-label") || el.title)) || "").replace(/\s+/g, " ").trim();
  }

  function currentIconUrl(doc) {
    return ICON_URL;
  }

  function syncAILogos(root) {
    var scope = root && root.querySelectorAll ? root : document;
    var doc = scope.ownerDocument || document;
    Array.prototype.slice.call(scope.querySelectorAll("img[data-ai-logo='true']")).forEach(function (img) {
      img.src = currentIconUrl(doc);
    });
  }

  function isAIButton(el) {
    if (!el || el.nodeType !== 1) return false;
    if (el.getAttribute("data-ai-button") === "true" || el.getAttribute("data-ai-action") === "true") return true;
    if (el.querySelector && el.querySelector('img[src*="pcs_ia_logo"]')) return true;
    var text = textOf(el);
    return IA_TEXT.test(text) && IA_ACTIONS.test(text);
  }

  function hasIcon(el) {
    return !!(el && el.querySelector && el.querySelector(".ai-button-icon, img[src*='pcs_ia_logo']"));
  }

  function enhance(el) {
    if (!isAIButton(el)) return;
    el.classList.add("ai-action-button");
    el.setAttribute("data-ai-enhanced", "1");
    if (!el.title) {
      el.title = "Funcion con inteligencia artificial";
    }
    if (!hasIcon(el)) {
      var img = (el.ownerDocument || document).createElement("img");
      img.className = "icon ai-button-icon";
      img.src = currentIconUrl(el.ownerDocument || document);
      img.alt = "";
      img.setAttribute("data-ai-logo", "true");
      img.setAttribute("aria-hidden", "true");
      el.insertBefore(img, el.firstChild);
    }
  }

  function applyAIButtonIcons(root) {
    var scope = root && root.querySelectorAll ? root : document;
    ensureAIButtonStyles(scope.ownerDocument || scope);
    Array.prototype.slice.call(scope.querySelectorAll("button, a.btn, .btn, .capture-btn, .fin-center-chip, .ai-action-btn")).forEach(enhance);
    syncAILogos(scope);
  }

  function enhanceFrame(frame) {
    if (!frame) return;
    try {
      if (frame.contentDocument) applyAIButtonIcons(frame.contentDocument);
    } catch (_) {}
  }

  function observeFrames() {
    Array.prototype.slice.call(document.querySelectorAll("iframe")).forEach(function (frame) {
      enhanceFrame(frame);
      if (frame.getAttribute("data-ai-frame-listener") === "1") return;
      frame.setAttribute("data-ai-frame-listener", "1");
      frame.addEventListener("load", function () {
        setTimeout(function () { enhanceFrame(frame); }, 60);
      });
    });
  }

  window.PCSApplyAIButtonIcons = applyAIButtonIcons;

  function boot() {
    applyAIButtonIcons(document);
    observeFrames();
    if (window.MutationObserver) {
      var observer = new MutationObserver(function () {
        applyAIButtonIcons(document);
        observeFrames();
      });
      var target = document.documentElement || document.body;
      if (target && target.nodeType === 1) {
        observer.observe(target, { childList: true, subtree: true, attributes: true, attributeFilter: ["data-theme"] });
      }
    }
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", boot);
  } else {
    boot();
  }
})();
