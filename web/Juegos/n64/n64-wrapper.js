(function () {
  var REMOTE_INDEX_URL = 'https://hulkholden.github.io/n64js/index.html';
  var REMOTE_BASE_URL = 'https://hulkholden.github.io/n64js/';
  var BRIDGE_CHANNEL = 'pcs-n64-bridge';
  var DB_NAME = 'pcs_arcade_n64';
  var DB_VERSION = 1;
  var ROM_STORE = 'roms';
  var LAST_ROM_KEY = 'last-rom';

  var romInput = document.getElementById('romInput');
  var saveBtn = document.getElementById('saveBtn');
  var loadBtn = document.getElementById('loadBtn');
  var fullscreenBtn = document.getElementById('fullscreenBtn');
  var iframe = document.getElementById('n64Frame');
  var statusEl = document.getElementById('n64Status');
  var loadedRomEl = document.getElementById('n64LoadedRom');

  var iframeReady = false;
  var currentRomMeta = null;
  var activeKeys = new Set();

  function setStatus(text, isError) {
    statusEl.textContent = text;
    statusEl.classList.toggle('error', !!isError);
  }

  function setLoadedRom(meta) {
    if (!meta || !meta.name) {
      loadedRomEl.textContent = 'ROM activa: ninguna';
      return;
    }
    var label = meta.name;
    if (meta.romId) {
      label += ' · id ' + meta.romId;
    }
    loadedRomEl.textContent = 'ROM activa: ' + label;
  }

  function openDatabase() {
    return new Promise(function (resolve, reject) {
      var request = indexedDB.open(DB_NAME, DB_VERSION);
      request.onupgradeneeded = function (event) {
        var db = event.target.result;
        if (!db.objectStoreNames.contains(ROM_STORE)) {
          db.createObjectStore(ROM_STORE, { keyPath: 'key' });
        }
      };
      request.onsuccess = function () {
        resolve(request.result);
      };
      request.onerror = function () {
        reject(request.error || new Error('No se pudo abrir IndexedDB.'));
      };
    });
  }

  function runStore(mode, executor) {
    return openDatabase().then(function (db) {
      return new Promise(function (resolve, reject) {
        var transaction = db.transaction(ROM_STORE, mode);
        var store = transaction.objectStore(ROM_STORE);
        executor(store, resolve, reject);
        transaction.oncomplete = function () { db.close(); };
        transaction.onerror = function () { reject(transaction.error || new Error('Falló la transacción IndexedDB.')); };
      });
    });
  }

  function saveLastRomRecord(record) {
    return runStore('readwrite', function (store, resolve, reject) {
      var request = store.put(record);
      request.onsuccess = function () { resolve(); };
      request.onerror = function () { reject(request.error || new Error('No se pudo guardar la ROM.')); };
    });
  }

  function loadLastRomRecord() {
    return runStore('readonly', function (store, resolve, reject) {
      var request = store.get(LAST_ROM_KEY);
      request.onsuccess = function () { resolve(request.result || null); };
      request.onerror = function () { reject(request.error || new Error('No se pudo leer la ROM guardada.')); };
    });
  }

  function backupKey(romId) {
    return 'pcs_n64_backup_' + romId;
  }

  function postToFrame(type, payload, transferList) {
    if (!iframe.contentWindow || !iframeReady) {
      throw new Error('El emulador todavía no está listo.');
    }
    iframe.contentWindow.postMessage({ channel: BRIDGE_CHANNEL, type: type, payload: payload || null }, '*', transferList || []);
  }

  function cloneArrayBuffer(buffer) {
    return buffer.slice(0);
  }

  function buildBridgeScript() {
    return [
      '<script>',
      '(function () {',
      '  var CHANNEL = ' + JSON.stringify(BRIDGE_CHANNEL) + ';',
      '  function post(type, payload) { window.parent.postMessage({ channel: CHANNEL, type: type, payload: payload || null }, "*"); }',
      '  function simplifyUi() {',
      '    var container = document.querySelector(".container-fluid");',
      '    if (!container) return;',
      '    Array.prototype.forEach.call(container.children, function (child) {',
      '      if (child.querySelector && child.querySelector("#display")) {',
      '        child.classList.add("pcs-n64-keep");',
      '      } else {',
      '        child.style.display = "none";',
      '      }',
      '    });',
      '    var display = document.getElementById("display");',
      '    if (display) {',
      '      display.style.display = "block";',
      '      display.style.width = "100%";',
      '      display.style.height = "100%";',
      '      display.style.maxWidth = "100%";',
      '      display.style.backgroundColor = "#000";',
      '    }',
      '    document.body.style.margin = "0";',
      '    document.body.style.background = "#000";',
      '    document.body.style.overflow = "hidden";',
      '    container.style.padding = "0";',
      '  }',
      '  function dispatchKey(key, isDown) {',
      '    var name = isDown ? "keydown" : "keyup";',
      '    var evt = new KeyboardEvent(name, { key: key, code: key, bubbles: true, cancelable: true });',
      '    window.dispatchEvent(evt);',
      '    document.dispatchEvent(evt);',
      '    var display = document.getElementById("display");',
      '    if (display) display.dispatchEvent(evt);',
      '  }',
      '  function romInfo() {',
      '    try {',
      '      var hardware = window.n64js && window.n64js.hardware ? window.n64js.hardware() : null;',
      '      var info = hardware && hardware.rominfo ? hardware.rominfo : null;',
      '      if (!info || !info.id) return null;',
      '      return { romId: info.id, name: info.name || "ROM N64", saveType: info.save || "desconocido" };',
      '    } catch (error) {',
      '      return null;',
      '    }',
      '  }',
      '  function exportSave() {',
      '    var hardware = window.n64js.hardware();',
      '    if (hardware && typeof hardware.flushSaveData === "function") {',
      '      hardware.flushSaveData();',
      '    }',
      '    var info = romInfo();',
      '    if (!info) throw new Error("No hay ROM cargada.");',
      '    var payload = {',
      '      romId: info.romId,',
      '      name: info.name,',
      '      saveType: info.saveType,',
      '      save: window.n64js.getLocalStorageItem("save") || null,',
      '      mempacks: [0, 1, 2, 3].map(function (index) { return window.n64js.getLocalStorageItem("mempack" + index) || null; })',
      '    };',
      '    post("save-export", payload);',
      '  }',
      '  function importSave(payload) {',
      '    if (!payload) return;',
      '    if (payload.save) window.n64js.setLocalStorageItem("save", payload.save);',
      '    if (payload.mempacks && payload.mempacks.length) {',
      '      payload.mempacks.forEach(function (item, index) {',
      '        if (item) window.n64js.setLocalStorageItem("mempack" + index, item);',
      '      });',
      '    }',
      '  }',
      '  function waitForBridgeReady() {',
      '    if (window.n64js && typeof window.n64js.loadRomAndStartRunning === "function") {',
      '      simplifyUi();',
      '      post("ready");',
      '      return;',
      '    }',
      '    window.setTimeout(waitForBridgeReady, 60);',
      '  }',
      '  window.addEventListener("message", function (event) {',
      '    var data = event.data;',
      '    if (!data || data.channel !== CHANNEL) return;',
      '    if (!window.n64js || typeof window.n64js.loadRomAndStartRunning !== "function") return;',
      '    try {',
      '      if (data.type === "load-rom") {',
      '        window.n64js.loadRomAndStartRunning(data.payload.buffer);',
      '        post("rom-meta", romInfo());',
      '      } else if (data.type === "dispatch-key") {',
      '        dispatchKey(data.payload.key, !!data.payload.isDown);',
      '      } else if (data.type === "fullscreen") {',
      '        window.n64js.toggleFullscreen();',
      '      } else if (data.type === "export-save") {',
      '        exportSave();',
      '      } else if (data.type === "restore-save") {',
      '        importSave(data.payload.backup || null);',
      '        window.n64js.loadRomAndStartRunning(data.payload.buffer);',
      '        post("rom-meta", romInfo());',
      '      }',
      '    } catch (error) {',
      '      post("error", String(error && error.message ? error.message : error));',
      '    }',
      '  });',
      '  document.addEventListener("DOMContentLoaded", waitForBridgeReady);',
      '})();',
      '<\/script>'
    ].join('');
  }

  function injectBridge(html) {
    var withBase = html.replace('<head>', '<head><base href="' + REMOTE_BASE_URL + '">');
    return withBase.replace('</body>', buildBridgeScript() + '</body>');
  }

  function loadEmulatorShell() {
    setStatus('Cargando núcleo N64 para móvil...');
    return fetch(REMOTE_INDEX_URL).then(function (response) {
      if (!response.ok) {
        throw new Error('No se pudo descargar n64js.');
      }
      return response.text();
    }).then(function (html) {
      iframe.srcdoc = injectBridge(html);
    });
  }

  function readFileAsArrayBuffer(file) {
    return new Promise(function (resolve, reject) {
      var reader = new FileReader();
      reader.onload = function () { resolve(reader.result); };
      reader.onerror = function () { reject(reader.error || new Error('No se pudo leer la ROM.')); };
      reader.readAsArrayBuffer(file);
    });
  }

  function sendRomRecord(record, message) {
    if (!record || !record.buffer) {
      setStatus('No hay ROM guardada todavía.', true);
      return Promise.resolve();
    }
    if (!iframeReady) {
      setStatus('El emulador aún se está preparando. Intenta de nuevo en un momento.', true);
      return Promise.resolve();
    }
    currentRomMeta = { name: record.name || 'ROM N64' };
    setLoadedRom(currentRomMeta);
    setStatus(message || ('Cargando ' + (record.name || 'ROM N64') + '...'));
    var buffer = cloneArrayBuffer(record.buffer);
    postToFrame('load-rom', { buffer: buffer }, [buffer]);
    return Promise.resolve();
  }

  function saveUploadedRom(file) {
    return readFileAsArrayBuffer(file).then(function (buffer) {
      var record = {
        key: LAST_ROM_KEY,
        name: file.name,
        size: file.size,
        updatedAt: new Date().toISOString(),
        buffer: buffer
      };
      return saveLastRomRecord(record).then(function () {
        setStatus('ROM legal guardada en este navegador: ' + file.name + '.');
        return sendRomRecord(record, 'Abriendo ' + file.name + '...');
      });
    });
  }

  function restoreLastRom() {
    return loadLastRomRecord().then(function (record) {
      if (!record) {
        setStatus('Sube tu ROM legal de Mario 64 o cualquier otra ROM propia para comenzar.');
        return null;
      }
      return sendRomRecord(record, 'Recuperando la última ROM guardada: ' + record.name + '...');
    }).catch(function (error) {
      setStatus('No se pudo recuperar la ROM guardada: ' + error.message, true);
      return null;
    });
  }

  function triggerSaveBackup() {
    if (!iframeReady) {
      setStatus('El emulador todavía no está listo para guardar.', true);
      return;
    }
    if (!currentRomMeta || !currentRomMeta.romId) {
      setStatus('Primero carga una ROM y guarda dentro del juego.', true);
      return;
    }
    setStatus('Extrayendo memoria del cartucho para guardar el progreso...');
    postToFrame('export-save');
  }

  function triggerLoadBackup() {
    if (!currentRomMeta || !currentRomMeta.romId) {
      setStatus('Primero carga una ROM para poder restaurar un guardado.', true);
      return;
    }
    var raw = localStorage.getItem(backupKey(currentRomMeta.romId));
    if (!raw) {
      setStatus('No existe un respaldo local para esta ROM. Guarda dentro del juego y luego toca Guardar.', true);
      return;
    }
    loadLastRomRecord().then(function (record) {
      if (!record || !record.buffer) {
        setStatus('No se encontró la ROM activa en IndexedDB. Vuelve a subirla.', true);
        return;
      }
      var backup = JSON.parse(raw);
      setStatus('Restaurando memoria del cartucho y reiniciando la ROM...');
      var buffer = cloneArrayBuffer(record.buffer);
      postToFrame('restore-save', {
        backup: backup.payload,
        buffer: buffer
      }, [buffer]);
    }).catch(function (error) {
      setStatus('No se pudo restaurar el guardado: ' + error.message, true);
    });
  }

  function handleBridgeMessage(event) {
    var data = event.data;
    if (!data || data.channel !== BRIDGE_CHANNEL) {
      return;
    }
    if (data.type === 'ready') {
      iframeReady = true;
      setStatus('Emulador listo. Sube tu ROM legal o espera la restauración automática.');
      restoreLastRom();
      return;
    }
    if (data.type === 'rom-meta') {
      currentRomMeta = data.payload || null;
      setLoadedRom(currentRomMeta);
      if (currentRomMeta && currentRomMeta.name) {
        setStatus('ROM cargada: ' + currentRomMeta.name + '. Si es Mario 64, guarda dentro del juego y luego usa Guardar.');
      }
      return;
    }
    if (data.type === 'save-export') {
      var payload = data.payload || null;
      if (!payload || !payload.romId) {
        setStatus('El emulador no devolvió un guardado válido.', true);
        return;
      }
      localStorage.setItem(backupKey(payload.romId), JSON.stringify({
        savedAt: new Date().toISOString(),
        payload: payload
      }));
      setStatus('Respaldo local guardado para ' + payload.name + '. Ya puedes tocar Cargar cuando quieras restaurarlo.');
      return;
    }
    if (data.type === 'error') {
      setStatus('Error del emulador: ' + data.payload, true);
    }
  }

  function bindTouchControls() {
    document.querySelectorAll('#touchGuide [data-key]').forEach(function (button) {
      function press(event) {
        event.preventDefault();
        var key = button.getAttribute('data-key');
        if (!key || activeKeys.has(key) || !iframeReady) {
          return;
        }
        activeKeys.add(key);
        button.classList.add('is-active');
        postToFrame('dispatch-key', { key: key, isDown: true });
      }

      function release(event) {
        if (event) {
          event.preventDefault();
        }
        var key = button.getAttribute('data-key');
        if (!key || !activeKeys.has(key) || !iframeReady) {
          return;
        }
        activeKeys.delete(key);
        button.classList.remove('is-active');
        postToFrame('dispatch-key', { key: key, isDown: false });
      }

      button.addEventListener('pointerdown', press);
      button.addEventListener('pointerup', release);
      button.addEventListener('pointercancel', release);
      button.addEventListener('pointerleave', release);
    });

    window.addEventListener('blur', function () {
      activeKeys.forEach(function (key) {
        try {
          postToFrame('dispatch-key', { key: key, isDown: false });
        } catch (_) {}
      });
      activeKeys.clear();
      document.querySelectorAll('#touchGuide .is-active').forEach(function (button) {
        button.classList.remove('is-active');
      });
    });
  }

  window.addEventListener('message', handleBridgeMessage);

  romInput.addEventListener('change', function (event) {
    var file = event.target.files && event.target.files[0];
    if (!file) {
      return;
    }
    saveUploadedRom(file).catch(function (error) {
      setStatus('No se pudo cargar la ROM: ' + error.message, true);
    }).finally(function () {
      romInput.value = '';
    });
  });

  saveBtn.addEventListener('click', triggerSaveBackup);
  loadBtn.addEventListener('click', triggerLoadBackup);
  fullscreenBtn.addEventListener('click', function () {
    try {
      postToFrame('fullscreen');
    } catch (error) {
      setStatus('La pantalla completa aún no está disponible: ' + error.message, true);
    }
  });

  bindTouchControls();

  loadEmulatorShell().catch(function (error) {
    setStatus('No se pudo preparar el emulador N64: ' + error.message, true);
  });
})();
