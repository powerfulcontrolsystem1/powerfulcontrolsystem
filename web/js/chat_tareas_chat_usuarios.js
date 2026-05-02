(function () {
  "use strict";

  var POLL_INTERVAL_MS = 3000;
  var state = {
    empresaId: 0,
    conversacionId: 0,
    mensajes: [],
    participantes: [],
    pollingHandle: 0,
    sending: false,
    lastMessageId: 0,
    bootstrapped: false
  };

  function n(v) {
    var x = Number(String(v || "").trim());
    return Number.isFinite(x) && x > 0 ? Math.trunc(x) : 0;
  }

  function txt(v) {
    return String(v == null ? "" : v).trim();
  }

  function esc(v) {
    return String(v == null ? "" : v)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#39;");
  }

  function formatDateTime(value) {
    var raw = txt(value);
    if (!raw) return "";
    var iso = raw.replace(" ", "T");
    var d = new Date(iso);
    if (Number.isNaN(d.getTime())) return raw;
    try {
      return new Intl.DateTimeFormat("es-CO", {
        dateStyle: "short",
        timeStyle: "short"
      }).format(d);
    } catch (_) {
      return raw;
    }
  }

  function setStatus(message, isError) {
    var node = document.getElementById("chatUsuariosStatus");
    if (!node) return;
    node.textContent = message || "";
    node.classList.toggle("is-error", !!isError);
  }

  function resolveEmpresaId() {
    try {
      var params = new URLSearchParams(window.location.search || "");
      var fromQuery = n(params.get("empresa_id") || params.get("id"));
      if (fromQuery) return fromQuery;
    } catch (_) {}
    try {
      if (window.parent && typeof window.parent.__resolveEmpresaIdContext === "function") {
        var parentId = n(window.parent.__resolveEmpresaIdContext());
        if (parentId) return parentId;
      }
    } catch (_) {}
    try {
      var stored = n(window.sessionStorage.getItem("empresa_id") || window.localStorage.getItem("empresa_id"));
      if (stored) return stored;
    } catch (_) {}
    return 0;
  }

  async function api(path, opts) {
    var res = await fetch(path, Object.assign({ credentials: "same-origin" }, opts || {}));
    var text = await res.text();
    var data = null;
    try {
      data = text ? JSON.parse(text) : null;
    } catch (_) {
      data = text;
    }
    if (!res.ok) {
      var message = data && typeof data === "object" ? (data.error || data.message) : "";
      throw new Error(message || text || ("HTTP " + res.status));
    }
    return data;
  }

  async function ensureGeneralConversation() {
    var payload = await api("/api/empresa/chat_tareas/conversaciones?empresa_id=" + encodeURIComponent(String(state.empresaId)) + "&action=chat_usuarios");
    var item = payload && payload.conversacion ? payload.conversacion : null;
    state.conversacionId = n(item && item.id);
    if (!state.conversacionId) {
      throw new Error("No se pudo preparar el chat general.");
    }
    var titleNode = document.getElementById("chatUsuariosConversationTitle");
    if (titleNode) {
      titleNode.textContent = txt(item.titulo) || "Chat usuarios";
    }
    var descNode = document.getElementById("chatUsuariosConversationDesc");
    if (descNode) {
      descNode.textContent = txt(item.descripcion) || "Canal general interno de la empresa.";
    }
  }

  async function loadParticipantes() {
    if (!state.conversacionId) return;
    var rows = await api("/api/empresa/chat_tareas/participantes?empresa_id=" + encodeURIComponent(String(state.empresaId)) + "&conversacion_id=" + encodeURIComponent(String(state.conversacionId)));
    state.participantes = Array.isArray(rows) ? rows : [];
    renderParticipantes();
  }

  async function loadMensajes(isSoftRefresh) {
    if (!state.conversacionId) return;
    var rows = await api("/api/empresa/chat_tareas/mensajes?empresa_id=" + encodeURIComponent(String(state.empresaId)) + "&conversacion_id=" + encodeURIComponent(String(state.conversacionId)) + "&limit=250");
    var next = Array.isArray(rows) ? rows : [];
    var prevLastId = state.lastMessageId;
    state.mensajes = next;
    state.lastMessageId = next.length ? n(next[next.length - 1].id) : 0;
    renderMensajes();
    if (isSoftRefresh && state.lastMessageId && state.lastMessageId !== prevLastId) {
      renderPulse();
    }
  }

  function renderParticipantes() {
    var node = document.getElementById("chatUsuariosParticipants");
    if (!node) return;
    if (!state.participantes.length) {
      node.innerHTML = '<li class="chat-usuarios-empty-pill">Sin participantes visibles todavía.</li>';
      return;
    }
    node.innerHTML = state.participantes.map(function (item) {
      var nombre = txt(item.nombre) || txt(item.email) || "Usuario";
      var tipo = txt(item.participante_tipo) || "usuario";
      return '<li class="chat-usuarios-participant-pill">' +
        '<span class="chat-usuarios-participant-name">' + esc(nombre) + '</span>' +
        '<span class="chat-usuarios-participant-type">' + esc(tipo) + '</span>' +
      '</li>';
    }).join("");
    var counter = document.getElementById("chatUsuariosParticipantsCount");
    if (counter) {
      counter.textContent = String(state.participantes.length);
    }
  }

  function renderMensajes() {
    var node = document.getElementById("chatUsuariosMessages");
    if (!node) return;
    if (!state.mensajes.length) {
      node.innerHTML = '<div class="chat-usuarios-empty-state"><strong>Sin mensajes todavía.</strong><span>Escribe el primer mensaje para abrir la conversación de equipo.</span></div>';
      return;
    }
    node.innerHTML = state.mensajes.map(function (item) {
      var autor = txt(item.autor_nombre) || txt(item.autor_email) || "Usuario";
      var contenido = txt(item.contenido);
      var fecha = formatDateTime(item.fecha_envio || item.fecha_creacion);
      return '<article class="chat-usuarios-message-card">' +
        '<header class="chat-usuarios-message-head">' +
          '<strong>' + esc(autor) + '</strong>' +
          '<time>' + esc(fecha) + '</time>' +
        '</header>' +
        '<div class="chat-usuarios-message-body">' + esc(contenido).replace(/\n/g, "<br>") + '</div>' +
      '</article>';
    }).join("");
    node.scrollTop = node.scrollHeight;
    var counter = document.getElementById("chatUsuariosMessagesCount");
    if (counter) {
      counter.textContent = String(state.mensajes.length);
    }
  }

  function renderPulse() {
    var badge = document.getElementById("chatUsuariosLiveBadge");
    if (!badge) return;
    badge.classList.remove("is-pulse");
    void badge.offsetWidth;
    badge.classList.add("is-pulse");
  }

  async function sendMessage() {
    if (state.sending || !state.conversacionId) return;
    var input = document.getElementById("chatUsuariosInput");
    var value = txt(input && input.value);
    if (!value) {
      setStatus("Escribe un mensaje antes de enviarlo.", true);
      return;
    }
    state.sending = true;
    var button = document.getElementById("chatUsuariosSendBtn");
    if (button) button.disabled = true;
    try {
      await api("/api/empresa/chat_tareas/mensajes", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          empresa_id: state.empresaId,
          conversacion_id: state.conversacionId,
          contenido: value,
          tipo_mensaje: "texto"
        })
      });
      if (input) input.value = "";
      setStatus("Mensaje enviado. El equipo ya lo está viendo.", false);
      await loadMensajes(true);
      await loadParticipantes();
    } catch (err) {
      setStatus(err && err.message ? err.message : "No se pudo enviar el mensaje.", true);
    } finally {
      state.sending = false;
      if (button) button.disabled = false;
    }
  }

  function startPolling() {
    stopPolling();
    state.pollingHandle = window.setInterval(function () {
      loadMensajes(true).catch(function () {});
      loadParticipantes().catch(function () {});
    }, POLL_INTERVAL_MS);
  }

  function stopPolling() {
    if (state.pollingHandle) {
      window.clearInterval(state.pollingHandle);
      state.pollingHandle = 0;
    }
  }

  async function init() {
    state.empresaId = resolveEmpresaId();
    if (!state.empresaId) {
      setStatus("No se pudo resolver la empresa activa para este chat.", true);
      return;
    }
    try {
      await ensureGeneralConversation();
      await loadParticipantes();
      await loadMensajes(false);
      startPolling();
      state.bootstrapped = true;
      setStatus("Chat general activo y sincronizado.", false);
    } catch (err) {
      setStatus(err && err.message ? err.message : "No se pudo iniciar el chat de usuarios.", true);
    }
  }

  window.addEventListener("beforeunload", stopPolling);
  document.addEventListener("visibilitychange", function () {
    if (!state.bootstrapped) return;
    if (document.hidden) {
      stopPolling();
      return;
    }
    startPolling();
    loadMensajes(true).catch(function () {});
    loadParticipantes().catch(function () {});
  });

  document.addEventListener("DOMContentLoaded", function () {
    var form = document.getElementById("chatUsuariosComposer");
    if (form) {
      form.addEventListener("submit", function (ev) {
        ev.preventDefault();
        sendMessage();
      });
    }
    var refresh = document.getElementById("chatUsuariosRefreshBtn");
    if (refresh) {
      refresh.addEventListener("click", function () {
        Promise.all([loadMensajes(true), loadParticipantes()]).then(function () {
          setStatus("Chat actualizado manualmente.", false);
        }).catch(function (err) {
          setStatus(err && err.message ? err.message : "No se pudo actualizar.", true);
        });
      });
    }
    init();
  });
})();
