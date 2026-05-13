(function () {
  var endpoint = "/api/public/juegos/records";
  var bestSent = Object.create(null);

  function cleanText(value, fallback) {
    var text = String(value || "").replace(/\s+/g, " ").trim();
    return text || fallback || "";
  }

  function parsePositiveInt(value) {
    var match = String(value || "").replace(/[^\d-]+/g, " ").trim().match(/-?\d+/g);
    if (!match || !match.length) return 0;
    var parsed = parseInt(match[match.length - 1], 10);
    return Number.isFinite(parsed) && parsed > 0 ? parsed : 0;
  }

  function slugFromPath(pathname) {
    var path = String(pathname || window.location.pathname || "").replace(/\\/g, "/");
    var parts = path.split("/").filter(Boolean);
    var juegosIndex = parts.indexOf("Juegos");
    if (juegosIndex >= 0 && parts[juegosIndex + 1]) {
      var next = parts[juegosIndex + 1];
      return next.replace(/\.html$/i, "");
    }
    if (parts.length) return parts[parts.length - 1].replace(/\.html$/i, "") || "juego";
    return "juego";
  }

  function resolveEmpresaId() {
    var params = new URLSearchParams(window.location.search || "");
    var fromUrl = parsePositiveInt(params.get("empresa_id") || params.get("id"));
    if (fromUrl) return String(fromUrl);
    var keys = ["active_empresa_id", "empresa_id", "admin_empresa_id"];
    var stores = [window.sessionStorage, window.localStorage];
    for (var s = 0; s < stores.length; s += 1) {
      for (var k = 0; k < keys.length; k += 1) {
        try {
          var value = parsePositiveInt(stores[s].getItem(keys[k]));
          if (value) return String(value);
        } catch (error) {}
      }
    }
    return "Publico";
  }

  function playerName() {
    var keys = ["pcs_juegos_player_name", "usuario_nombre", "admin_nombre", "usuario_email", "admin_email"];
    var stores = [window.localStorage, window.sessionStorage];
    for (var s = 0; s < stores.length; s += 1) {
      for (var k = 0; k < keys.length; k += 1) {
        try {
          var value = cleanText(stores[s].getItem(keys[k]), "");
          if (value) return value.slice(0, 80);
        } catch (error) {}
      }
    }
    return "Jugador";
  }

  function readScoreFromDocument(doc) {
    if (!doc) return null;
    var selectors = ["[data-score]", "#score", "#scoreBoard", "#finalScore", ".score", "[id*='score' i]", "[class*='score' i]"];
    for (var i = 0; i < selectors.length; i += 1) {
      var node = null;
      try {
        node = doc.querySelector(selectors[i]);
      } catch (error) {}
      if (!node) continue;
      var value = node.getAttribute && node.getAttribute("data-score");
      var score = parsePositiveInt(value || node.textContent || "");
      if (score > 0) return score;
    }
    return null;
  }

  function readLevelFromDocument(doc) {
    if (!doc) return 1;
    var node = null;
    try {
      node = doc.querySelector("[data-level], #level, [id*='level' i]");
    } catch (error) {}
    return Math.max(1, parsePositiveInt((node && (node.getAttribute("data-level") || node.textContent)) || "1") || 1);
  }

  function loadTop(juego, limit) {
    var url = endpoint + "?juego=" + encodeURIComponent(juego || slugFromPath()) + "&limit=" + encodeURIComponent(limit || 5);
    return fetch(url, { credentials: "same-origin", cache: "no-store" }).then(function (response) {
      if (!response.ok) throw new Error("No fue posible cargar records.");
      return response.json();
    });
  }

  function submitRecord(payload, options) {
    payload = payload || {};
    options = options || {};
    var juego = cleanText(payload.juego || slugFromPath(), "juego");
    var puntaje = parsePositiveInt(payload.puntaje);
    if (puntaje <= 0) return Promise.resolve(null);
    if (!options.force && bestSent[juego] && puntaje <= bestSent[juego]) {
      return Promise.resolve(null);
    }
    bestSent[juego] = puntaje;
    return fetch(endpoint, {
      method: "POST",
      credentials: "same-origin",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        juego: juego,
        nombre_jugador: cleanText(payload.nombre_jugador || playerName(), "Jugador"),
        empresa_id: cleanText(payload.empresa_id || resolveEmpresaId(), "Publico"),
        puntaje: puntaje,
        nivel: Math.max(1, parsePositiveInt(payload.nivel) || 1)
      })
    }).then(function (response) {
      if (!response.ok) throw new Error("No fue posible guardar el record.");
      return response.json();
    }).catch(function (error) {
      console.warn("Record de juego no guardado", error);
      return null;
    });
  }

  function ensureRecordPanel(root, title) {
    if (!root || root.querySelector(".arcade-records-panel")) return root && root.querySelector(".arcade-records-panel");
    var panel = document.createElement("aside");
    panel.className = "arcade-records-panel";
    panel.innerHTML =
      "<strong>Records</strong>" +
      "<span class=\"arcade-records-status\">Sin puntaje todavía</span>" +
      "<ol class=\"arcade-records-list\"></ol>";
    var header = document.querySelector(".arcade-header, .pacman-header") || root;
    header.appendChild(panel);
    if (title) panel.setAttribute("aria-label", "Records de " + title);
    return panel;
  }

  function renderTop(panel, records) {
    if (!panel) return;
    var list = panel.querySelector(".arcade-records-list");
    var status = panel.querySelector(".arcade-records-status");
    records = Array.isArray(records) ? records : [];
    if (!records.length) {
      status.textContent = "Aun no hay records guardados";
      list.innerHTML = "";
      return;
    }
    status.textContent = "Top " + records.length + " global";
    list.innerHTML = records.map(function (item) {
      return "<li><span>" + cleanText(item.nombre_jugador, "Jugador") + "</span><b>" + parsePositiveInt(item.puntaje) + "</b></li>";
    }).join("");
  }

  function enhanceWrapper(config) {
    config = config || {};
    var frame = config.frame || document.querySelector("[data-arcade-frame]");
    var juego = cleanText(config.juego || slugFromPath(), "juego");
    var title = cleanText(config.title || (frame && frame.title), juego);
    var panel = ensureRecordPanel(document.body, title);
    var lastScore = 0;

    function refreshTop() {
      loadTop(juego, 5).then(function (records) {
        renderTop(panel, records);
      }).catch(function () {});
    }

    function recordScore(score, level, final) {
      score = parsePositiveInt(score);
      if (score <= 0 || score < lastScore) return;
      lastScore = score;
      if (panel) {
        var status = panel.querySelector(".arcade-records-status");
        if (status) status.textContent = "Puntaje actual: " + score;
      }
      submitRecord({ juego: juego, puntaje: score, nivel: level || 1 }, { force: Boolean(final) }).then(refreshTop);
    }

    window.addEventListener("message", function (event) {
      var data = event && event.data;
      if (!data || data.type !== "pcs-game-score") return;
      recordScore(data.puntaje || data.score, data.nivel || data.level, data.final);
    });

    if (frame) {
      window.setInterval(function () {
        try {
          var doc = frame.contentDocument;
          var win = frame.contentWindow;
          var score = readScoreFromDocument(doc);
          if (!score && win) {
            score = parsePositiveInt(
              (win.Game && win.Game.score) ||
              win.PCSGameScore ||
              win.PCSTetrisScore ||
              win.PCSPongScore ||
              win.PCSPacmanScore
            );
          }
          if (score) recordScore(score, readLevelFromDocument(doc), false);
        } catch (error) {}
      }, 1600);
    }

    refreshTop();
    return { recordScore: recordScore, refreshTop: refreshTop };
  }

  function installAutoReporter(config) {
    config = config || {};
    var juego = cleanText(config.juego || slugFromPath(), "juego");
    var last = 0;
    function report(final) {
      var score = readScoreFromDocument(document);
      if (!score) {
        score = parsePositiveInt(window.PCSGameScore || window.PCSTetrisScore || window.PCSPongScore || window.PCSPacmanScore || 0);
      }
      if (!score || score < last) return;
      last = score;
      try {
        if (window.parent && window.parent !== window) {
          window.parent.postMessage({
            type: "pcs-game-score",
            juego: juego,
            puntaje: score,
            nivel: readLevelFromDocument(document),
            final: Boolean(final)
          }, window.location.origin);
        } else {
          submitRecord({ juego: juego, puntaje: score, nivel: readLevelFromDocument(document) }, { force: Boolean(final) });
        }
      } catch (error) {}
    }
    window.setInterval(function () { report(false); }, 1400);
    window.addEventListener("beforeunload", function () { report(true); });
    document.addEventListener("visibilitychange", function () {
      if (document.hidden) report(true);
    });
    return report;
  }

  window.PCSJuegos = {
    endpoint: endpoint,
    slugFromPath: slugFromPath,
    resolveEmpresaId: resolveEmpresaId,
    playerName: playerName,
    parsePositiveInt: parsePositiveInt,
    readScoreFromDocument: readScoreFromDocument,
    readLevelFromDocument: readLevelFromDocument,
    submitRecord: submitRecord,
    loadTop: loadTop,
    enhanceWrapper: enhanceWrapper,
    installAutoReporter: installAutoReporter
  };
}());
