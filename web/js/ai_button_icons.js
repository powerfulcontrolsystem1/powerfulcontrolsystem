(function () {
  "use strict";

  var ICON_URL = "/img/ia_mark.svg";
  var IA_TEXT = /\b(ia|i\.a\.|inteligencia artificial|gpt|chatgpt)\b/i;
  var IA_ACTIONS = /(analizar|generar|extraer|diagnosticar|consultar|preparar|borrador|asistente|centro).*(ia|gpt)|(ia|gpt).*(analizar|generar|extraer|diagnosticar|consultar|preparar|borrador|asistente|centro)/i;

  function textOf(el) {
    return String((el && (el.textContent || el.getAttribute("aria-label") || el.title)) || "").replace(/\s+/g, " ").trim();
  }

  function isAIButton(el) {
    if (!el || el.nodeType !== 1) return false;
    if (el.getAttribute("data-ai-button") === "true" || el.getAttribute("data-ai-action") === "true") return true;
    if (el.querySelector && (el.querySelector('img[src*="gpt.svg"]') || el.querySelector('img[src*="ia_mark.svg"]'))) return true;
    var text = textOf(el);
    return IA_TEXT.test(text) && IA_ACTIONS.test(text);
  }

  function hasIcon(el) {
    return !!(el && el.querySelector && el.querySelector(".ai-button-icon, img[src*='gpt.svg'], img[src*='ia_mark.svg']"));
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
      img.src = ICON_URL;
      img.alt = "";
      img.setAttribute("aria-hidden", "true");
      el.insertBefore(img, el.firstChild);
    }
  }

  function applyAIButtonIcons(root) {
    var scope = root && root.querySelectorAll ? root : document;
    Array.prototype.slice.call(scope.querySelectorAll("button, a.btn, .btn, .capture-btn, .fin-center-chip")).forEach(enhance);
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
      observer.observe(document.documentElement, { childList: true, subtree: true });
    }
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", boot);
  } else {
    boot();
  }
})();
