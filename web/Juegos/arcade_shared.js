(function (global) {
  var PROFILE_KEY = 'pcs_arcade_profile_v1';
  var SCORES_KEY = 'pcs_arcade_scores_v1';
  var SETTINGS_KEY = 'pcs_arcade_settings_v1';
  var MAX_NAME_LENGTH = 28;
  var MAX_DETAIL_LENGTH = 84;
  var MAX_ENTRIES_PER_GAME = 5;
  var audioContext = null;

  function readJSON(key, fallback) {
    try {
      var raw = global.localStorage.getItem(key);
      if (!raw) return fallback;
      var parsed = JSON.parse(raw);
      return parsed && typeof parsed === 'object' ? parsed : fallback;
    } catch (_) {
      return fallback;
    }
  }

  function writeJSON(key, value) {
    try {
      global.localStorage.setItem(key, JSON.stringify(value));
      return true;
    } catch (_) {
      return false;
    }
  }

  function sanitizeText(value, maxLength) {
    var normalized = String(value == null ? '' : value)
      .replace(/\s+/g, ' ')
      .trim();
    if (!normalized) return '';
    return normalized.slice(0, maxLength || 200);
  }

  function dispatchUpdate(reason, detail) {
    try {
      global.dispatchEvent(new CustomEvent('pcs-arcade-store-changed', {
        detail: {
          reason: reason || '',
          payload: detail || null
        }
      }));
    } catch (_) {}
  }

  function sortEntries(entries) {
    return entries.slice().sort(function (left, right) {
      var scoreDiff = Number(right.score || 0) - Number(left.score || 0);
      if (scoreDiff !== 0) return scoreDiff;
      var leftTime = String(left.recordedAt || '');
      var rightTime = String(right.recordedAt || '');
      if (leftTime === rightTime) return 0;
      return leftTime < rightTime ? 1 : -1;
    });
  }

  function getProfile() {
    var data = readJSON(PROFILE_KEY, {});
    return {
      playerName: sanitizeText(data.playerName, MAX_NAME_LENGTH)
    };
  }

  function getPlayerName(fallback) {
    var playerName = getProfile().playerName;
    return playerName || fallback || 'Invitado';
  }

  function setPlayerName(name) {
    var playerName = sanitizeText(name, MAX_NAME_LENGTH);
    writeJSON(PROFILE_KEY, { playerName: playerName });
    dispatchUpdate('profile', { playerName: playerName });
    return getProfile();
  }

  function promptPlayerName(message) {
    var currentName = getProfile().playerName;
    var nextValue = global.prompt(message || 'Escribe el nombre del jugador', currentName || '');
    if (nextValue == null) {
      return getProfile();
    }
    return setPlayerName(nextValue);
  }

  function ensurePlayerName(message) {
    var profile = getProfile();
    if (profile.playerName) return profile;
    profile = promptPlayerName(message || 'Escribe el nombre que quieres usar en los marcadores');
    if (profile.playerName) return profile;
    return setPlayerName('Invitado');
  }

  function getSettings() {
    var data = readJSON(SETTINGS_KEY, {});
    return {
      soundEnabled: data.soundEnabled !== false
    };
  }

  function isSoundEnabled() {
    return getSettings().soundEnabled;
  }

  function setSoundEnabled(enabled) {
    var nextValue = !!enabled;
    writeJSON(SETTINGS_KEY, { soundEnabled: nextValue });
    dispatchUpdate('settings', { soundEnabled: nextValue });
    return getSettings();
  }

  function toggleSoundEnabled() {
    return setSoundEnabled(!isSoundEnabled());
  }

  function getAllScores() {
    return readJSON(SCORES_KEY, {});
  }

  function normalizeEntries(entries) {
    if (!Array.isArray(entries)) return [];
    return sortEntries(entries.filter(function (entry) {
      return entry && typeof entry === 'object';
    }).map(function (entry) {
      return {
        playerName: sanitizeText(entry.playerName, MAX_NAME_LENGTH) || 'Invitado',
        score: Math.max(0, Math.round(Number(entry.score) || 0)),
        detail: sanitizeText(entry.detail, MAX_DETAIL_LENGTH),
        recordedAt: sanitizeText(entry.recordedAt, 64) || new Date().toISOString()
      };
    })).slice(0, MAX_ENTRIES_PER_GAME);
  }

  function getGameScores(gameSlug) {
    var allScores = getAllScores();
    var normalizedSlug = sanitizeText(gameSlug, 80);
    var current = normalizedSlug ? allScores[normalizedSlug] : null;
    var entries = normalizeEntries(current && current.entries);
    return {
      slug: normalizedSlug,
      entries: entries,
      best: entries[0] || null
    };
  }

  function recordScore(gameSlug, score, options) {
    var normalizedSlug = sanitizeText(gameSlug, 80);
    if (!normalizedSlug) {
      return { slug: '', entries: [], best: null };
    }

    var numericScore = Math.max(0, Math.round(Number(score) || 0));
    var meta = options && typeof options === 'object' ? options : {};
    var allScores = getAllScores();
    var currentScores = getGameScores(normalizedSlug);
    var entry = {
      playerName: sanitizeText(meta.playerName, MAX_NAME_LENGTH) || getPlayerName('Invitado'),
      score: numericScore,
      detail: sanitizeText(meta.detail, MAX_DETAIL_LENGTH),
      recordedAt: new Date().toISOString()
    };

    var nextEntries = currentScores.entries.slice();
    nextEntries.push(entry);
    allScores[normalizedSlug] = { entries: normalizeEntries(nextEntries) };
    writeJSON(SCORES_KEY, allScores);
    dispatchUpdate('scores', { slug: normalizedSlug, entry: entry });
    return getGameScores(normalizedSlug);
  }

  function getGlobalSummary() {
    var allScores = getAllScores();
    var games = Object.keys(allScores).map(function (slug) {
      var gameScores = getGameScores(slug);
      return {
        slug: slug,
        best: gameScores.best,
        entries: gameScores.entries
      };
    }).filter(function (item) {
      return item.best;
    });

    games.sort(function (left, right) {
      return Number(right.best.score || 0) - Number(left.best.score || 0);
    });

    return {
      totalGamesWithScores: games.length,
      totalEntries: games.reduce(function (count, item) {
        return count + item.entries.length;
      }, 0),
      overallBest: games[0] || null
    };
  }

  function formatScoreEntry(entry) {
    if (!entry) return 'Sin marcas guardadas';
    var label = entry.playerName + ' · ' + String(entry.score);
    return entry.detail ? label + ' · ' + entry.detail : label;
  }

  function formatClock(totalSeconds) {
    var safe = Math.max(0, Math.floor(Number(totalSeconds) || 0));
    var minutes = Math.floor(safe / 60);
    var seconds = safe % 60;
    return String(minutes).padStart(2, '0') + ':' + String(seconds).padStart(2, '0');
  }

  function clamp(value, min, max) {
    return Math.max(min, Math.min(max, value));
  }

  function randomItem(list) {
    if (!Array.isArray(list) || !list.length) return null;
    return list[Math.floor(Math.random() * list.length)];
  }

  var POWER_LIBRARY = [
    { id: 'shield', label: 'Escudo', detail: 'Bloquea un golpe duro.', cooldown: 16 },
    { id: 'chrono', label: 'Crono', detail: 'Ralentiza el ritmo enemigo.', cooldown: 18 },
    { id: 'heart', label: 'Vida', detail: 'Regala una vida adicional.', cooldown: 26 },
    { id: 'jackpot', label: 'Jackpot', detail: 'Suma puntos y bonus.', cooldown: 18 },
    { id: 'magnet', label: 'Magneto', detail: 'Atrae premios y bonus.', cooldown: 20 },
    { id: 'combo', label: 'Combo', detail: 'Duplica el impulso del marcador.', cooldown: 18 },
    { id: 'pulse', label: 'Pulso', detail: 'Limpia riesgos inmediatos.', cooldown: 22 },
    { id: 'ghost', label: 'Fantasma', detail: 'Da unos segundos de intangibilidad.', cooldown: 24 },
    { id: 'radar', label: 'Radar', detail: 'Entrega una ayuda tactica.', cooldown: 16 },
    { id: 'repair', label: 'Repair', detail: 'Recarga recursos especiales.', cooldown: 18 },
    { id: 'overdrive', label: 'Boost', detail: 'Activa un turbo ofensivo.', cooldown: 20 },
    { id: 'storm', label: 'Storm', detail: 'Golpea varias amenazas a la vez.', cooldown: 24 }
  ];

  var PRIZE_LIBRARY = [
    'Ficha doble',
    'Bono relampago',
    'Caja arcade',
    'Ticket premium',
    'Racha de campeon',
    'Cofre movil'
  ];

  function createPowerSystem(options) {
    var config = options && typeof options === 'object' ? options : {};
    var doc = global.document;
    var root = config.root || doc.querySelector('.arcade-window');
    var chipRow = config.chipRow || (root ? root.querySelector('.arcade-chip-row') : null);
    var stage = config.stage || (root ? root.querySelector('.arcade-stage') : null);
    var onChange = typeof config.onChange === 'function' ? config.onChange : function () {};
    var setStatus = typeof config.setStatus === 'function' ? config.setStatus : function () {};
    var stateRef = config.state && typeof config.state === 'object' ? config.state : null;
    var handlers = config.handlers && typeof config.handlers === 'object' ? config.handlers : {};
    var powerState = {
      energy: 0,
      tickets: 0,
      streak: 0,
      prize: 'Sin premio',
      charges: {},
      cooldowns: {},
      mounted: false,
      hidden: false
    };
    var ui = {
      wrap: null,
      energyValue: null,
      ticketValue: null,
      prizeValue: null,
      buttons: {}
    };

    POWER_LIBRARY.forEach(function (power) {
      powerState.charges[power.id] = 0;
      powerState.cooldowns[power.id] = 0;
    });

    function setPowerLabel(label) {
      if (typeof handlers.setPowerLabel === 'function') {
        handlers.setPowerLabel({ label: label });
      } else if (stateRef && typeof stateRef.powerLabel !== 'undefined') {
        stateRef.powerLabel = label;
      }
    }

    function invokeHandler(name, payload) {
      var handler = handlers[name];
      if (typeof handler !== 'function') return false;
      handler(payload || {});
      return true;
    }

    function addScore(value) {
      var safe = Math.max(0, Math.round(Number(value) || 0));
      if (!safe) return;
      if (!invokeHandler('addScore', { value: safe }) && stateRef && typeof stateRef.score === 'number') {
        stateRef.score += safe;
      }
    }

    function addBonus(value) {
      var safe = Math.max(0, Math.round(Number(value) || 0));
      if (!safe) return;
      if (!invokeHandler('addBonus', { value: safe }) && stateRef && typeof stateRef.bonus === 'number') {
        stateRef.bonus += safe;
      }
    }

    function addLife(value) {
      var safe = Math.max(1, Math.round(Number(value) || 1));
      if (!invokeHandler('grantLife', { value: safe, label: 'Vida' }) && stateRef && typeof stateRef.lives === 'number') {
        stateRef.lives += safe;
      }
    }

    function ensureMounted() {
      if (powerState.mounted || !root || !chipRow || !stage) return;
      var wrap = doc.createElement('section');
      wrap.className = 'arcade-power-rack';
      wrap.innerHTML = '' +
        '<div class="arcade-power-summary">' +
          '<div class="arcade-power-heading">' +
            '<strong>Poderes y premios</strong>' +
            '<span>12 clases tactiles para cada juego</span>' +
          '</div>' +
          '<div class="arcade-power-stats">' +
            '<span class="arcade-power-stat">Energia <strong data-arcade-energy>0%</strong></span>' +
            '<span class="arcade-power-stat">Tickets <strong data-arcade-tickets>0</strong></span>' +
            '<span class="arcade-power-stat arcade-power-prize">Premio <strong data-arcade-prize>Sin premio</strong></span>' +
          '</div>' +
        '</div>' +
        '<div class="arcade-power-grid" data-arcade-power-grid></div>';
      root.insertBefore(wrap, stage);
      var grid = wrap.querySelector('[data-arcade-power-grid]');
      POWER_LIBRARY.forEach(function (power) {
        var button = doc.createElement('button');
        button.type = 'button';
        button.className = 'arcade-power-btn';
        button.setAttribute('data-power-id', power.id);
        button.innerHTML = '' +
          '<span class="arcade-power-btn-top">' +
            '<strong>' + power.label + '</strong>' +
            '<span data-power-count>0</span>' +
          '</span>' +
          '<span class="arcade-power-btn-detail">' + power.detail + '</span>' +
          '<span class="arcade-power-btn-cd" data-power-cooldown>Listo</span>';
        grid.appendChild(button);
        ui.buttons[power.id] = button;
      });
      ui.wrap = wrap;
      ui.energyValue = wrap.querySelector('[data-arcade-energy]');
      ui.ticketValue = wrap.querySelector('[data-arcade-tickets]');
      ui.prizeValue = wrap.querySelector('[data-arcade-prize]');
      wrap.addEventListener('click', function (event) {
        var button = event.target.closest('button[data-power-id]');
        if (!button) return;
        activate(button.getAttribute('data-power-id'));
      });
      powerState.mounted = true;
      render();
    }

    function render() {
      ensureMounted();
      if (!powerState.mounted) return;
      ui.energyValue.textContent = String(Math.round(powerState.energy)) + '%';
      ui.ticketValue.textContent = String(powerState.tickets);
      ui.prizeValue.textContent = powerState.prize;
      POWER_LIBRARY.forEach(function (power) {
        var button = ui.buttons[power.id];
        if (!button) return;
        var count = powerState.charges[power.id] || 0;
        var cooldown = powerState.cooldowns[power.id] || 0;
        button.querySelector('[data-power-count]').textContent = 'x' + String(count);
        button.querySelector('[data-power-cooldown]').textContent = cooldown > 0 ? ('CD ' + String(Math.ceil(cooldown)) + 's') : 'Listo';
        button.disabled = count <= 0 || cooldown > 0 || powerState.hidden;
        button.classList.toggle('active', count > 0 && cooldown <= 0 && !powerState.hidden);
      });
    }

    function setPrize(label, ticketGain) {
      powerState.prize = sanitizeText(label, 40) || randomItem(PRIZE_LIBRARY) || 'Premio arcade';
      powerState.tickets += Math.max(1, Math.round(Number(ticketGain) || 1));
      setStatus('Premio desbloqueado: ' + powerState.prize + '.');
      render();
    }

    function grantCharge(powerId, count, silent) {
      if (!powerState.charges.hasOwnProperty(powerId)) return;
      powerState.charges[powerId] += Math.max(1, Math.round(Number(count) || 1));
      if (!silent) {
        setStatus('Poder listo: ' + (POWER_LIBRARY.find(function (item) { return item.id === powerId; }) || {}).label + '.');
      }
      render();
    }

    function grantRandomPower(amount) {
      var count = Math.max(1, Math.round(Number(amount) || 1));
      while (count > 0) {
        var next = randomItem(POWER_LIBRARY);
        if (next) grantCharge(next.id, 1, true);
        count -= 1;
      }
      render();
    }

    function rewardEnergy(amount, reason) {
      var gain = clamp(Math.round(Number(amount) || 0), 0, 100);
      if (!gain) return;
      powerState.energy += gain;
      while (powerState.energy >= 100) {
        powerState.energy -= 100;
        powerState.streak += 1;
        grantRandomPower(powerState.streak % 3 === 0 ? 2 : 1);
        setPrize(reason || randomItem(PRIZE_LIBRARY), powerState.streak % 2 === 0 ? 3 : 2);
      }
      render();
    }

    function activate(powerId) {
      var def = POWER_LIBRARY.find(function (item) { return item.id === powerId; });
      if (!def) return false;
      if ((powerState.charges[powerId] || 0) <= 0 || (powerState.cooldowns[powerId] || 0) > 0 || powerState.hidden) return false;
      powerState.charges[powerId] -= 1;
      powerState.cooldowns[powerId] = def.cooldown;
      setPowerLabel(def.label);
      if (powerId === 'shield') {
        invokeHandler('grantShield', { value: 1, label: def.label });
      } else if (powerId === 'chrono') {
        invokeHandler('slowTime', { duration: 12, label: def.label });
      } else if (powerId === 'heart') {
        addLife(1);
        addBonus(18);
      } else if (powerId === 'jackpot') {
        addScore(70);
        addBonus(40);
        setPrize('Jackpot movil', 4);
      } else if (powerId === 'magnet') {
        addScore(26);
        addBonus(26);
        invokeHandler('magnet', { duration: 12, label: def.label });
      } else if (powerId === 'combo') {
        addScore(54);
        addBonus(22);
        invokeHandler('combo', { duration: 10, label: def.label });
      } else if (powerId === 'pulse') {
        invokeHandler('pulse', { strength: 1, label: def.label });
        addBonus(18);
      } else if (powerId === 'ghost') {
        invokeHandler('ghost', { duration: 10, label: def.label });
        addBonus(14);
      } else if (powerId === 'radar') {
        invokeHandler('radar', { charges: 1, label: def.label });
        addBonus(10);
      } else if (powerId === 'repair') {
        invokeHandler('repair', { charges: 1, label: def.label });
        addBonus(12);
      } else if (powerId === 'overdrive') {
        invokeHandler('overdrive', { duration: 12, label: def.label });
        addScore(34);
      } else if (powerId === 'storm') {
        invokeHandler('storm', { strength: 1, label: def.label });
        addScore(38);
      }
      playEffect('powerUp');
      onChange();
      render();
      return true;
    }

    function setHidden(hidden) {
      powerState.hidden = !!hidden;
      render();
    }

    function reset() {
      powerState.energy = 16;
      powerState.tickets = 0;
      powerState.streak = 0;
      powerState.prize = 'Sin premio';
      Object.keys(powerState.charges).forEach(function (key) {
        powerState.charges[key] = 0;
        powerState.cooldowns[key] = 0;
      });
      grantCharge('shield', 1, true);
      grantCharge('chrono', 1, true);
      grantCharge('radar', 1, true);
      setHidden(false);
      render();
      onChange();
    }

    function tick(delta) {
      var step = Number(delta) || 0;
      if (step > 10) step = step / 1000;
      if (step <= 0) step = 1 / 60;
      Object.keys(powerState.cooldowns).forEach(function (key) {
        if (powerState.cooldowns[key] > 0) {
          powerState.cooldowns[key] = Math.max(0, powerState.cooldowns[key] - step);
        }
      });
      render();
    }

    ensureMounted();

    return {
      reset: reset,
      tick: tick,
      activate: activate,
      pause: function () { setHidden(true); },
      resume: function () { setHidden(false); },
      finish: function () { setHidden(true); },
      noteScore: function (amount, reason) {
        rewardEnergy(Math.max(2, Math.round((Number(amount) || 0) / 12)), reason || 'Racha de puntos');
      },
      noteBonus: function (amount, reason) {
        rewardEnergy(Math.max(2, Math.round((Number(amount) || 0) / 14)), reason || 'Bonus movil');
      },
      noteLevel: function (level) {
        var safeLevel = Math.max(1, Math.round(Number(level) || 1));
        grantRandomPower(Math.min(2, 1 + Math.floor(safeLevel / 4)));
        setPrize('Nivel ' + String(level) + ' superado', 2);
      },
      notePrize: function (label, tickets) {
        setPrize(label, tickets);
      },
      grantPower: grantCharge,
      render: render
    };
  }

  function getAudioContext() {
    var AudioCtor = global.AudioContext || global.webkitAudioContext;
    if (!AudioCtor) return null;
    if (!audioContext) {
      try {
        audioContext = new AudioCtor();
      } catch (_) {
        audioContext = null;
      }
    }
    if (audioContext && audioContext.state === 'suspended') {
      audioContext.resume().catch(function () {});
    }
    return audioContext;
  }

  function unlockAudio() {
    var ctx = getAudioContext();
    if (!ctx) return;
    try {
      var buffer = ctx.createBuffer(1, 1, 22050);
      var source = ctx.createBufferSource();
      source.buffer = buffer;
      source.connect(ctx.destination);
      source.start(0);
    } catch (_) {}
  }

  function playSequence(sequence) {
    if (!isSoundEnabled()) return;
    var ctx = getAudioContext();
    if (!ctx || !Array.isArray(sequence) || !sequence.length) return;

    var cursor = ctx.currentTime + 0.015;
    sequence.forEach(function (note) {
      var duration = Math.max(0.04, Number(note.duration) || 0.12);
      var volume = Math.max(0.001, Number(note.volume) || 0.06);
      var attack = Math.min(0.02, duration * 0.45);
      var osc = ctx.createOscillator();
      var gain = ctx.createGain();

      osc.type = note.type || 'sine';
      osc.frequency.setValueAtTime(Number(note.frequency) || 440, cursor);
      if (note.slideTo) {
        osc.frequency.linearRampToValueAtTime(Number(note.slideTo), cursor + duration);
      }

      gain.gain.setValueAtTime(0.0001, cursor);
      gain.gain.exponentialRampToValueAtTime(volume, cursor + attack);
      gain.gain.exponentialRampToValueAtTime(0.0001, cursor + duration);

      osc.connect(gain);
      gain.connect(ctx.destination);
      osc.start(cursor);
      osc.stop(cursor + duration + 0.025);
      cursor += duration;
    });
  }

  var effects = {
    start: [
      { frequency: 392, duration: 0.08, volume: 0.06, type: 'triangle' },
      { frequency: 523.25, duration: 0.11, volume: 0.07, type: 'triangle' },
      { frequency: 659.25, duration: 0.12, volume: 0.07, type: 'triangle' }
    ],
    countdownTick: [
      { frequency: 740, duration: 0.05, volume: 0.04, type: 'square' },
      { frequency: 587.33, duration: 0.07, volume: 0.03, type: 'triangle' }
    ],
    countdownGo: [
      { frequency: 523.25, duration: 0.06, volume: 0.05, type: 'triangle' },
      { frequency: 659.25, duration: 0.06, volume: 0.06, type: 'triangle' },
      { frequency: 783.99, duration: 0.08, volume: 0.07, type: 'triangle' },
      { frequency: 1046.5, duration: 0.16, volume: 0.08, type: 'triangle' }
    ],
    point: [
      { frequency: 784, duration: 0.06, volume: 0.06, type: 'square' },
      { frequency: 1046.5, duration: 0.08, volume: 0.07, type: 'square' }
    ],
    fail: [
      { frequency: 320, slideTo: 180, duration: 0.18, volume: 0.08, type: 'sawtooth' },
      { frequency: 160, slideTo: 90, duration: 0.14, volume: 0.05, type: 'triangle' }
    ],
    win: [
      { frequency: 523.25, duration: 0.08, volume: 0.05, type: 'triangle' },
      { frequency: 659.25, duration: 0.08, volume: 0.06, type: 'triangle' },
      { frequency: 783.99, duration: 0.08, volume: 0.06, type: 'triangle' },
      { frequency: 1046.5, duration: 0.16, volume: 0.08, type: 'triangle' }
    ],
    launch: [
      { frequency: 210, slideTo: 530, duration: 0.12, volume: 0.06, type: 'sawtooth' }
    ],
    bounce: [
      { frequency: 240, duration: 0.05, volume: 0.045, type: 'square' },
      { frequency: 180, duration: 0.05, volume: 0.04, type: 'square' }
    ],
    hit: [
      { frequency: 160, duration: 0.04, volume: 0.05, type: 'triangle' },
      { frequency: 120, duration: 0.06, volume: 0.045, type: 'triangle' }
    ],
    lift: [
      { frequency: 520, slideTo: 760, duration: 0.05, volume: 0.035, type: 'triangle' }
    ],
    flip: [
      { frequency: 600, duration: 0.04, volume: 0.04, type: 'triangle' }
    ],
    match: [
      { frequency: 660, duration: 0.06, volume: 0.05, type: 'square' },
      { frequency: 880, duration: 0.08, volume: 0.05, type: 'square' }
    ],
    powerUp: [
      { frequency: 392, duration: 0.06, volume: 0.04, type: 'triangle' },
      { frequency: 523.25, duration: 0.06, volume: 0.05, type: 'triangle' },
      { frequency: 783.99, duration: 0.12, volume: 0.06, type: 'square' }
    ],
    move: [
      { frequency: 260, duration: 0.045, volume: 0.03, type: 'triangle' }
    ],
    paddle: [
      { frequency: 180, duration: 0.04, volume: 0.04, type: 'square' },
      { frequency: 260, duration: 0.05, volume: 0.04, type: 'square' }
    ],
    drop: [
      { frequency: 420, slideTo: 180, duration: 0.11, volume: 0.05, type: 'sawtooth' }
    ],
    capture: [
      { frequency: 330, duration: 0.05, volume: 0.04, type: 'triangle' },
      { frequency: 494, duration: 0.07, volume: 0.05, type: 'triangle' },
      { frequency: 659.25, duration: 0.09, volume: 0.06, type: 'square' }
    ],
    shoot: [
      { frequency: 920, duration: 0.03, volume: 0.035, type: 'square' },
      { frequency: 700, duration: 0.04, volume: 0.03, type: 'triangle' }
    ],
    eat: [
      { frequency: 480, duration: 0.05, volume: 0.05, type: 'square' },
      { frequency: 720, duration: 0.07, volume: 0.06, type: 'square' }
    ]
  };

  function playEffect(name) {
    var sequence = effects[name];
    if (!sequence) return;
    playSequence(sequence);
  }

  function createGameSession(options) {
    var config = options && typeof options === 'object' ? options : {};
    var overlay = config.overlay || null;
    var overlayTitle = config.overlayTitle || null;
    var overlayText = config.overlayText || null;
    var timerElement = config.timerElement || null;
    var startButton = config.startButton || null;
    var countdownSeconds = Math.max(1, Math.round(Number(config.countdownSeconds) || 5));
    var state = {
      elapsed: 0,
      running: false,
      paused: false,
      countdownHandle: 0,
      timerHandle: 0,
      countdownActive: false
    };

    function renderTimer() {
      if (timerElement) {
        timerElement.textContent = formatClock(state.elapsed);
      }
    }

    function clearCountdown() {
      if (state.countdownHandle) {
        global.clearInterval(state.countdownHandle);
        state.countdownHandle = 0;
      }
      state.countdownActive = false;
      if (overlay) {
        overlay.classList.remove('arcade-overlay-countdown');
      }
      if (startButton) {
        startButton.disabled = false;
      }
    }

    function stopTimer() {
      if (state.timerHandle) {
        global.clearInterval(state.timerHandle);
        state.timerHandle = 0;
      }
      state.running = false;
      state.paused = false;
    }

    function startTimer() {
      stopTimer();
      state.running = true;
      renderTimer();
      state.timerHandle = global.setInterval(function () {
        if (!state.running || state.paused) return;
        state.elapsed += 1;
        renderTimer();
      }, 1000);
    }

    function reset() {
      clearCountdown();
      stopTimer();
      state.elapsed = 0;
      renderTimer();
    }

    function pause() {
      state.paused = true;
    }

    function resume() {
      if (state.running) {
        state.paused = false;
      }
    }

    function finish() {
      clearCountdown();
      stopTimer();
    }

    function startCountdown(onDone, countdownOptions) {
      var settings = countdownOptions && typeof countdownOptions === 'object' ? countdownOptions : {};
      var seconds = Math.max(1, Math.round(Number(settings.seconds) || countdownSeconds));
      var label = sanitizeText(settings.label, 60) || 'Comienza en';
      reset();
      state.countdownActive = true;
      if (startButton) {
        startButton.disabled = true;
      }
      if (overlay) {
        overlay.hidden = false;
        overlay.classList.add('arcade-overlay-countdown');
      }
      var current = seconds;

      function renderCountdown() {
        if (overlayTitle) overlayTitle.textContent = String(current);
        if (overlayText) overlayText.textContent = label;
      }

      renderCountdown();
      playEffect('countdownTick');
      state.countdownHandle = global.setInterval(function () {
        current -= 1;
        if (current <= 0) {
          clearCountdown();
          if (overlayTitle) overlayTitle.textContent = 'YA';
          if (overlayText) overlayText.textContent = 'A jugar';
          if (overlay) overlay.hidden = false;
          playEffect('countdownGo');
          global.setTimeout(function () {
            if (overlay) overlay.hidden = true;
            startTimer();
            if (typeof onDone === 'function') onDone();
          }, 520);
          return;
        }
        renderCountdown();
        playEffect('countdownTick');
      }, 1000);
    }

    renderTimer();

    return {
      reset: reset,
      pause: pause,
      resume: resume,
      finish: finish,
      startCountdown: startCountdown,
      getElapsedSeconds: function () { return state.elapsed; },
      isCountdownActive: function () { return state.countdownActive; }
    };
  }

  ['pointerdown', 'keydown', 'touchstart'].forEach(function (eventName) {
    global.addEventListener(eventName, unlockAudio, { capture: true, passive: true });
  });

  global.PCSArcade = {
    getProfile: getProfile,
    getPlayerName: getPlayerName,
    setPlayerName: setPlayerName,
    promptPlayerName: promptPlayerName,
    ensurePlayerName: ensurePlayerName,
    getSettings: getSettings,
    isSoundEnabled: isSoundEnabled,
    setSoundEnabled: setSoundEnabled,
    toggleSoundEnabled: toggleSoundEnabled,
    getGameScores: getGameScores,
    recordScore: recordScore,
    getGlobalSummary: getGlobalSummary,
    formatScoreEntry: formatScoreEntry,
    formatClock: formatClock,
    createPowerSystem: createPowerSystem,
    unlockAudio: unlockAudio,
    playEffect: playEffect,
    createGameSession: createGameSession
  };
})(window);