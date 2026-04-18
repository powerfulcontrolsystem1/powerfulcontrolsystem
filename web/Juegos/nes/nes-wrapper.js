(function () {
  var SLOT_COUNT = 5;
  var browser = null;
  var currentSlot = null;
  var running = false;
  var pressedKeys = new Set();

  var stage = document.getElementById('nes-stage');
  var statusBar = document.getElementById('statusBar');
  var pauseBtn = document.getElementById('pauseBtn');
  var saveStateBtn = document.getElementById('saveStateBtn');
  var loadStateBtn = document.getElementById('loadStateBtn');
  var fullscreenBtn = document.getElementById('fullscreenBtn');

  function slotRomKey(slot) {
    return 'pcs_nes_rom_slot_' + slot;
  }

  function slotMetaKey(slot) {
    return 'pcs_nes_rom_meta_' + slot;
  }

  function slotStateKey(slot) {
    return 'pcs_nes_state_slot_' + slot;
  }

  function setStatus(message, isError) {
    statusBar.textContent = message;
    statusBar.classList.toggle('error', !!isError);
  }

  function updatePauseLabel() {
    pauseBtn.textContent = running ? 'Pausar' : 'Reanudar';
  }

  function decodeBase64ToUint8Array(base64) {
    var binary = atob(base64);
    var bytes = new Uint8Array(binary.length);
    for (var index = 0; index < binary.length; index += 1) {
      bytes[index] = binary.charCodeAt(index);
    }
    return bytes;
  }

  function encodeArrayBufferToBase64(buffer) {
    var bytes = new Uint8Array(buffer);
    var binary = '';
    var chunk = 0x8000;
    for (var index = 0; index < bytes.length; index += chunk) {
      binary += String.fromCharCode.apply(null, bytes.subarray(index, index + chunk));
    }
    return btoa(binary);
  }

  function getSlotMeta(slot) {
    try {
      return JSON.parse(localStorage.getItem(slotMetaKey(slot)) || 'null');
    } catch (_) {
      return null;
    }
  }

  function renderSlotCards() {
    document.querySelectorAll('.slot').forEach(function (slotEl) {
      var slot = slotEl.getAttribute('data-slot');
      var metaEl = slotEl.querySelector('.slot-meta');
      var hasRom = !!localStorage.getItem(slotRomKey(slot));
      var meta = getSlotMeta(slot);
      var hasState = !!localStorage.getItem(slotStateKey(slot));
      if (!hasRom) {
        metaEl.textContent = 'Vacía';
      } else {
        var parts = [];
        parts.push(meta && meta.name ? meta.name : 'ROM guardada');
        if (hasState) {
          parts.push('partida lista');
        }
        metaEl.textContent = parts.join(' · ');
      }
      slotEl.classList.toggle('is-active-slot', String(currentSlot || '') === String(slot));
    });
  }

  function fitBrowser() {
    if (browser && typeof browser.fitInParent === 'function') {
      browser.fitInParent();
    }
  }

  function destroyBrowser() {
    if (browser && typeof browser.destroy === 'function') {
      browser.destroy();
    }
    browser = null;
    stage.innerHTML = '';
    running = false;
    updatePauseLabel();
  }

  function buildBrowser(romData, label) {
    destroyBrowser();
    browser = new jsnes.Browser({
      container: stage,
      romData: romData,
      onError: function (error) {
        setStatus('Fallo del emulador: ' + error, true);
      },
      onBatteryRamWrite: function () {
        if (currentSlot != null) {
          setStatus('Guardado interno del cartucho actualizado en la ranura ' + currentSlot + '.');
        }
      }
    });
    running = true;
    updatePauseLabel();
    fitBrowser();
    setStatus('ROM cargada: ' + label + '. Usa Guardar partida para snapshot local.');
  }

  function loadSlot(slot) {
    var base64 = localStorage.getItem(slotRomKey(slot));
    if (!base64) {
      setStatus('La ranura ' + slot + ' está vacía.', true);
      return;
    }
    currentSlot = slot;
    var meta = getSlotMeta(slot);
    var romData = decodeBase64ToUint8Array(base64);
    buildBrowser(romData, meta && meta.name ? meta.name : 'Ranura ' + slot);
    renderSlotCards();
  }

  function saveRomToSlot(slot, file) {
    var reader = new FileReader();
    reader.onload = function () {
      localStorage.setItem(slotRomKey(slot), encodeArrayBufferToBase64(reader.result));
      localStorage.setItem(slotMetaKey(slot), JSON.stringify({
        name: file.name,
        size: file.size,
        updatedAt: new Date().toISOString()
      }));
      setStatus('ROM guardada en la ranura ' + slot + ': ' + file.name + '.');
      renderSlotCards();
    };
    reader.onerror = function () {
      setStatus('No se pudo leer la ROM de la ranura ' + slot + '.', true);
    };
    reader.readAsArrayBuffer(file);
  }

  function clearSlot(slot) {
    localStorage.removeItem(slotRomKey(slot));
    localStorage.removeItem(slotMetaKey(slot));
    localStorage.removeItem(slotStateKey(slot));
    if (String(currentSlot || '') === String(slot)) {
      currentSlot = null;
      destroyBrowser();
      setStatus('La ranura activa fue borrada. Carga otra ROM para continuar.');
    } else {
      setStatus('Se borró la ranura ' + slot + '.');
    }
    renderSlotCards();
  }

  function saveState() {
    if (!browser || !browser.nes) {
      setStatus('Primero carga una ROM para poder guardar una partida.', true);
      return;
    }
    if (currentSlot == null) {
      setStatus('La partida se guarda por ranura. Carga una ROM desde la biblioteca primero.', true);
      return;
    }
    try {
      localStorage.setItem(slotStateKey(currentSlot), JSON.stringify(browser.nes.toJSON()));
      renderSlotCards();
      setStatus('Partida guardada en la ranura ' + currentSlot + '.');
    } catch (error) {
      setStatus('No se pudo guardar la partida: ' + error, true);
    }
  }

  function restoreState() {
    if (!browser || !browser.nes) {
      setStatus('Primero carga una ROM para restaurar una partida.', true);
      return;
    }
    if (currentSlot == null) {
      setStatus('La restauración depende de una ranura activa.', true);
      return;
    }
    var raw = localStorage.getItem(slotStateKey(currentSlot));
    if (!raw) {
      setStatus('La ranura ' + currentSlot + ' no tiene una partida guardada.', true);
      return;
    }
    try {
      browser.nes.fromJSON(JSON.parse(raw));
      setStatus('Partida restaurada desde la ranura ' + currentSlot + '.');
    } catch (error) {
      setStatus('No se pudo restaurar la partida: ' + error, true);
    }
  }

  function togglePause() {
    if (!browser) {
      setStatus('Primero carga una ROM.', true);
      return;
    }
    if (running) {
      browser.stop();
      running = false;
      setStatus('Emulación en pausa.');
    } else {
      browser.start();
      running = true;
      setStatus('Emulación reanudada.');
    }
    updatePauseLabel();
  }

  function requestFullscreen() {
    var target = stage;
    if (!document.fullscreenElement && target.requestFullscreen) {
      target.requestFullscreen();
    } else if (document.exitFullscreen) {
      document.exitFullscreen();
    }
  }

  function keyCodeFor(key) {
    var map = {
      ArrowUp: 38,
      ArrowDown: 40,
      ArrowLeft: 37,
      ArrowRight: 39,
      Enter: 13,
      Control: 17,
      x: 88,
      z: 90
    };
    return map[key] || 0;
  }

  function codeFor(key) {
    var map = {
      ArrowUp: 'ArrowUp',
      ArrowDown: 'ArrowDown',
      ArrowLeft: 'ArrowLeft',
      ArrowRight: 'ArrowRight',
      Enter: 'Enter',
      Control: 'ControlRight',
      x: 'KeyX',
      z: 'KeyZ'
    };
    return map[key] || key;
  }

  function dispatchVirtualKey(key, isDown) {
    var eventName = isDown ? 'keydown' : 'keyup';
    var keyboardEvent = new KeyboardEvent(eventName, {
      key: key,
      code: codeFor(key),
      bubbles: true,
      cancelable: true
    });
    var keyCode = keyCodeFor(key);
    Object.defineProperty(keyboardEvent, 'keyCode', { get: function () { return keyCode; } });
    Object.defineProperty(keyboardEvent, 'which', { get: function () { return keyCode; } });
    window.dispatchEvent(keyboardEvent);
    document.dispatchEvent(keyboardEvent);
  }

  function releaseAllTouchKeys() {
    Array.from(pressedKeys).forEach(function (key) {
      dispatchVirtualKey(key, false);
      pressedKeys.delete(key);
    });
    document.querySelectorAll('.touch-key.is-active').forEach(function (button) {
      button.classList.remove('is-active');
    });
  }

  document.querySelectorAll('.slot').forEach(function (slotEl) {
    var slot = slotEl.getAttribute('data-slot');
    var input = slotEl.querySelector('.slot-file');
    var loadBtn = slotEl.querySelector('.load-slot');
    var clearBtn = slotEl.querySelector('.clear-slot');

    input.addEventListener('change', function (event) {
      var file = event.target.files && event.target.files[0];
      if (!file) {
        return;
      }
      saveRomToSlot(slot, file);
      input.value = '';
    });

    loadBtn.addEventListener('click', function () {
      loadSlot(slot);
    });

    clearBtn.addEventListener('click', function () {
      clearSlot(slot);
    });
  });

  document.querySelectorAll('#touchControls [data-key]').forEach(function (button) {
    button.addEventListener('pointerdown', function (event) {
      event.preventDefault();
      var key = button.getAttribute('data-key');
      if (!key || pressedKeys.has(key)) {
        return;
      }
      pressedKeys.add(key);
      button.classList.add('is-active');
      dispatchVirtualKey(key, true);
    });

    ['pointerup', 'pointercancel', 'pointerleave'].forEach(function (name) {
      button.addEventListener(name, function (event) {
        event.preventDefault();
        var key = button.getAttribute('data-key');
        if (!key || !pressedKeys.has(key)) {
          return;
        }
        pressedKeys.delete(key);
        button.classList.remove('is-active');
        dispatchVirtualKey(key, false);
      });
    });
  });

  document.addEventListener('pointerup', releaseAllTouchKeys);
  window.addEventListener('blur', releaseAllTouchKeys);
  window.addEventListener('resize', fitBrowser);

  pauseBtn.addEventListener('click', togglePause);
  saveStateBtn.addEventListener('click', saveState);
  loadStateBtn.addEventListener('click', restoreState);
  fullscreenBtn.addEventListener('click', requestFullscreen);

  renderSlotCards();
  updatePauseLabel();
})();