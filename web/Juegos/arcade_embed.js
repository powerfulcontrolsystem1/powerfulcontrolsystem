(function () {
  var KEY_META = {
    13: { key: "Enter", code: "Enter" },
    27: { key: "Escape", code: "Escape" },
    32: { key: " ", code: "Space" },
    37: { key: "ArrowLeft", code: "ArrowLeft" },
    38: { key: "ArrowUp", code: "ArrowUp" },
    39: { key: "ArrowRight", code: "ArrowRight" },
    40: { key: "ArrowDown", code: "ArrowDown" },
    48: { key: "0", code: "Digit0" },
    49: { key: "1", code: "Digit1" },
    50: { key: "2", code: "Digit2" },
    65: { key: "a", code: "KeyA" },
    70: { key: "f", code: "KeyF" },
    76: { key: "l", code: "KeyL" },
    80: { key: "p", code: "KeyP" },
    81: { key: "q", code: "KeyQ" },
    82: { key: "r", code: "KeyR" },
    83: { key: "s", code: "KeyS" }
  };

  var frame = document.querySelector("[data-arcade-frame]");
  if (!frame) {
    return;
  }

  function makeEvent(win, type, keyCode) {
    var meta = KEY_META[keyCode] || { key: String.fromCharCode(keyCode), code: "" };
    var event = new win.KeyboardEvent(type, {
      key: meta.key,
      code: meta.code,
      bubbles: true,
      cancelable: true,
      which: keyCode,
      keyCode: keyCode
    });
    Object.defineProperty(event, "keyCode", { get: function () { return keyCode; } });
    Object.defineProperty(event, "which", { get: function () { return keyCode; } });
    return event;
  }

  function sendKey(keyCode, type) {
    var win = frame.contentWindow;
    if (!win || !win.document) {
      return;
    }
    try {
      win.focus();
      frame.focus();
      win.dispatchEvent(makeEvent(win, type, keyCode));
      win.document.dispatchEvent(makeEvent(win, type, keyCode));
      if (win.document.body) {
        win.document.body.dispatchEvent(makeEvent(win, type, keyCode));
      }
    } catch (error) {
      console.warn("No se pudo enviar control arcade", error);
    }
  }

  function press(keyCode, hold) {
    sendKey(keyCode, "keydown");
    sendKey(keyCode, "keypress");
    if (!hold) {
      window.setTimeout(function () {
        sendKey(keyCode, "keyup");
      }, 80);
    }
  }

  document.querySelectorAll("[data-key]").forEach(function (button) {
    var keyCode = Number(button.dataset.key);
    var hold = button.dataset.mode !== "tap";

    button.addEventListener("pointerdown", function (event) {
      event.preventDefault();
      press(keyCode, hold);
      if (navigator.vibrate) {
        navigator.vibrate(14);
      }
    });

    ["pointerup", "pointercancel", "pointerleave"].forEach(function (type) {
      button.addEventListener(type, function () {
        if (hold) {
          sendKey(keyCode, "keyup");
        }
      });
    });

    button.addEventListener("click", function (event) {
      event.preventDefault();
    });
  });

  frame.addEventListener("load", function () {
    frame.focus();
  });
}());
