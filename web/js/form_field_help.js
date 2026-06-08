(function() {
  "use strict";

  var STYLE_ID = "pcs-field-help-style";
  var POPOVER_ID = "pcs-field-help-popover";
  var activeButton = null;

  function ensureStyles() {
    if (document.getElementById(STYLE_ID)) return;
    var style = document.createElement("style");
    style.id = STYLE_ID;
    style.textContent = [
      ".pcs-field-help-label{display:inline-flex;align-items:center;gap:6px;flex-wrap:wrap}",
      ".pcs-field-help-button{width:22px;height:22px;min-width:22px;border-radius:50%;border:1px solid var(--border,#cbd5e1);background:var(--surface,#fff);color:var(--text,#0f172a);font:800 14px/1 Arial,sans-serif;display:inline-flex;align-items:center;justify-content:center;cursor:pointer;padding:0;vertical-align:middle}",
      ".pcs-field-help-button:hover,.pcs-field-help-button:focus{border-color:var(--primary,#2563eb);box-shadow:0 0 0 3px rgba(37,99,235,.14);outline:none}",
      ".pcs-field-help-popover{position:absolute;z-index:2147483000;max-width:min(290px,calc(100vw - 20px));padding:10px 12px;border:1px solid var(--border,#cbd5e1);border-radius:8px;background:var(--card,#fff);color:var(--text,#0f172a);box-shadow:0 14px 34px rgba(15,23,42,.18);font:400 14px/1.42 Arial,sans-serif;letter-spacing:0}",
      ".pcs-field-help-popover strong{display:block;margin:0 0 4px;color:var(--text,#0f172a);font-size:14px;line-height:1.25}",
      ".pcs-field-help-popover p{margin:0;color:var(--muted,#475569);font-size:14px;line-height:1.42}",
      "@media (max-width:520px){.pcs-field-help-popover{max-width:calc(100vw - 16px)}}"
    ].join("");
    document.head.appendChild(style);
  }

  function escapeSelector(value) {
    if (window.CSS && typeof window.CSS.escape === "function") {
      return window.CSS.escape(value);
    }
    return String(value || "").replace(/[^a-zA-Z0-9_-]/g, "\\$&");
  }

  function text(value) {
    return String(value == null ? "" : value).replace(/\s+/g, " ").trim();
  }

  function configFrom(value) {
    if (value && typeof value === "object") {
      return {
        title: text(value.title),
        text: text(value.text || value.body || value.description)
      };
    }
    return { title: "", text: text(value) };
  }

  function labelText(label) {
    if (!label) return "Campo";
    var clone = label.cloneNode(true);
    clone.querySelectorAll(".pcs-field-help-button").forEach(function(node) {
      node.remove();
    });
    return text(clone.textContent) || "Campo";
  }

  function previousLabelIn(container, control) {
    if (!container || !control) return null;
    var labels = Array.prototype.slice.call(container.querySelectorAll("label"));
    var match = null;
    labels.forEach(function(label) {
      if (label.contains(control)) {
        match = label;
        return;
      }
      if (label.compareDocumentPosition(control) & Node.DOCUMENT_POSITION_FOLLOWING) {
        if (!match || match.compareDocumentPosition(label) & Node.DOCUMENT_POSITION_FOLLOWING) {
          match = label;
        }
      }
    });
    return match;
  }

  function findLabel(control) {
    if (!control || !control.id) return null;
    var selector = 'label[for="' + escapeSelector(control.id) + '"]';
    var explicit = document.querySelector(selector);
    if (explicit) return explicit;
    var wrapped = control.closest("label");
    if (wrapped) return wrapped;
    return previousLabelIn(control.closest(".form-col") || control.parentElement, control);
  }

  function closePopover() {
    var popover = document.getElementById(POPOVER_ID);
    if (popover) popover.remove();
    if (activeButton) {
      activeButton.setAttribute("aria-expanded", "false");
      activeButton = null;
    }
  }

  function positionPopover(popover, button) {
    var rect = button.getBoundingClientRect();
    var margin = 8;
    popover.style.left = "0px";
    popover.style.top = "0px";
    document.body.appendChild(popover);
    var width = popover.offsetWidth || 280;
    var left = rect.left + window.scrollX - 12;
    left = Math.max(margin + window.scrollX, Math.min(left, window.scrollX + window.innerWidth - width - margin));
    popover.style.left = left + "px";
    popover.style.top = (rect.bottom + window.scrollY + 7) + "px";
  }

  function openPopover(button, cfg) {
    closePopover();
    var popover = document.createElement("div");
    popover.id = POPOVER_ID;
    popover.className = "pcs-field-help-popover";
    popover.setAttribute("role", "dialog");
    popover.setAttribute("aria-live", "polite");
    var title = document.createElement("strong");
    title.textContent = cfg.title || button.getAttribute("data-pcs-help-title") || "Ayuda";
    var body = document.createElement("p");
    body.textContent = cfg.text || "Completa este dato segun la informacion real de la empresa.";
    popover.appendChild(title);
    popover.appendChild(body);
    activeButton = button;
    button.setAttribute("aria-expanded", "true");
    positionPopover(popover, button);
  }

  function addHelp(controlId, value) {
    var control = document.getElementById(controlId);
    if (!control) return false;
    var cfg = configFrom(value);
    if (!cfg.text) return false;
    var label = findLabel(control);
    if (!label || label.querySelector(".pcs-field-help-button")) return false;
    ensureStyles();
    var title = cfg.title || labelText(label);
    label.classList.add("pcs-field-help-label");
    var button = document.createElement("button");
    button.type = "button";
    button.className = "pcs-field-help-button";
    button.textContent = "?";
    button.setAttribute("aria-label", "Ayuda: " + title);
    button.setAttribute("aria-expanded", "false");
    button.setAttribute("data-pcs-help-title", title);
    button.addEventListener("click", function(ev) {
      ev.preventDefault();
      ev.stopPropagation();
      if (activeButton === button) {
        closePopover();
      } else {
        openPopover(button, { title: title, text: cfg.text });
      }
    });
    label.appendChild(button);
    return true;
  }

  function install(map) {
    if (!map || typeof map !== "object") return 0;
    ensureStyles();
    var total = 0;
    Object.keys(map).forEach(function(controlId) {
      if (addHelp(controlId, map[controlId])) total += 1;
    });
    return total;
  }

  document.addEventListener("click", function(ev) {
    if (!activeButton) return;
    var popover = document.getElementById(POPOVER_ID);
    if (popover && (popover.contains(ev.target) || activeButton.contains(ev.target))) return;
    closePopover();
  });
  document.addEventListener("keydown", function(ev) {
    if (ev.key === "Escape") closePopover();
  });
  window.addEventListener("resize", closePopover);
  window.addEventListener("scroll", closePopover, true);

  window.PCSFieldHelp = {
    install: install,
    close: closePopover
  };
})();
