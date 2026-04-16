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
    move: [
      { frequency: 260, duration: 0.045, volume: 0.03, type: 'triangle' }
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
    unlockAudio: unlockAudio,
    playEffect: playEffect
  };
})(window);