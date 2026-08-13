(function () {
  "use strict";

  function replaceBrokenImage(img) {
    if (!img || img.dataset.fallbackApplied === "1") return;
    var fallback = String(img.dataset.fallbackSrc || "").trim();
    if (fallback) {
      img.dataset.fallbackApplied = "1";
      img.src = fallback;
      return;
    }
    if (img.dataset.imageErrorMode === "empty-photo") {
      var empty = document.createElement("div");
      empty.className = "relay-device-photo relay-photo-preview empty";
      empty.textContent = "Sin foto";
      img.replaceWith(empty);
    }
  }

  document.addEventListener("error", function (event) {
    var target = event.target;
    if (target && target.tagName === "IMG") replaceBrokenImage(target);
  }, true);

  document.addEventListener("click", function (event) {
    var button = event.target && event.target.closest ? event.target.closest("[data-window-action]") : null;
    if (!button) return;
    var action = String(button.dataset.windowAction || "").toLowerCase();
    if (action === "print") window.print();
    if (action === "close") window.close();
  });

  function scheduleAutoPrint() {
    var delay = Number(document.body && document.body.dataset.autoPrintDelay || 0);
    if (!Number.isFinite(delay) || delay <= 0) return;
    window.setTimeout(function () { window.print(); }, Math.min(delay, 5000));
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", scheduleAutoPrint, { once: true });
  } else {
    scheduleAutoPrint();
  }
})();
