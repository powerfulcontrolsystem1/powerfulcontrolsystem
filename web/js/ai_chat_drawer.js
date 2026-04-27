(function () {
  var DRAWER_ID = 'aiChatDrawer';
  var TOGGLE_ID = 'openAIDrawer';
  var CLOSE_ID = 'closeAIDrawer';
  var FORM_ID = 'aiChatForm';
  var INPUT_ID = 'aiChatInput';
  var MODE_ID = 'aiChatMode';
  var ATTACHMENT_INPUT_ID = 'aiChatAttachment';
  var ATTACH_BTN_ID = 'aiChatAttachBtn';
  var CLEAR_ATTACHMENT_ID = 'aiChatClearAttachment';
  var ATTACHMENT_NAME_ID = 'aiChatAttachmentName';
  var MIC_ID = 'aiChatMicBtn';
  var VOICE_ID = 'aiChatVoiceBtn';
  var CONV_ID = 'aiChatConvBtn';
  var BACKDROP_ID = 'aiChatBackdrop';
  var MINIMIZE_ID = 'aiChatMinimize';
  var MINIBAR_ID = 'aiChatMinibar';
  var MINIBAR_EXPAND_ID = 'aiChatMinibarExpand';
  var MESSAGES_ID = 'aiChatMessages';
  var NOTICE_ID = 'aiChatNotice';
  var HINT_TOGGLE_ID = 'aiChatHintToggle';
  var HINTS_ID = 'aiChatHints';
  var MAX_ATTACHMENT_BYTES = 8 * 1024 * 1024;

  var state = {
    proposals: [],
    loading: false,
    selectedAttachment: null,
    voiceEnabled: false,
    listening: false,
    conversationMode: false
  };

  var ICON_MIC = '<svg viewBox="0 0 24 24" width="22" height="22" aria-hidden="true"><path fill="currentColor" d="M12 14c1.66 0 3-1.34 3-3V5c0-1.66-1.34-3-3-3S9 3.34 9 5v6c0 1.66 1.34 3 3 3zm5-3c0 2.76-2.24 5-5 5s-5-2.24-5-5H5c0 3.53 2.61 6.43 6 6.92V21h2v-3.08c3.39-.49 6-3.39 6-6.92h-2z"/></svg>';
  var ICON_SPK = '<svg viewBox="0 0 24 24" width="22" height="22" aria-hidden="true"><path fill="currentColor" d="M3 9v6h4l5 5V4L7 9H3zm13.5 3c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.25 2.5-4.02zM14 3.23v2.06c2.89.86 5 3.54 5 6.71s-2.11 5.85-5 6.71v2.06c4.01-.91 7-4.49 7-8.77s-2.99-7.86-7-8.77z"/></svg>';
  var ICON_CONV = '<svg viewBox="0 0 24 24" width="22" height="22" aria-hidden="true"><path fill="currentColor" d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm0 14H6l-2 2V4h16v12z"/></svg>';

  function isMobileChatViewport() {
    try {
      return window.matchMedia('(max-width: 860px)').matches;
    } catch (e) {
      return false;
    }
  }

  function scrollChatToBottom() {
    var messagesEl = document.getElementById(MESSAGES_ID);
    var host = messagesEl && messagesEl.closest('.ai-chat-body-scroll');
    if (host) {
      host.scrollTop = host.scrollHeight;
    } else if (messagesEl) {
      messagesEl.scrollTop = messagesEl.scrollHeight;
    }
  }

  function parsePositiveInt(raw) {
    var value = Number(String(raw || '').trim());
    if (!Number.isFinite(value)) return 0;
    value = Math.trunc(value);
    return value > 0 ? value : 0;
  }

  function getCurrentEmpresaId() {
    if (typeof window.__resolveEmpresaIdContext === 'function') {
      try {
        var resolved = window.__resolveEmpresaIdContext();
        if (parsePositiveInt(resolved) > 0) {
          return String(parsePositiveInt(resolved));
        }
      } catch (error) {
        // no-op
      }
    }

    var params = new URLSearchParams(window.location.search || '');
    var id = parsePositiveInt(params.get('empresa_id') || params.get('id') || '');
    if (id > 0) {
      return String(id);
    }

    var keys = ['active_empresa_id', 'empresa_id', 'admin_empresa_id'];
    var stores = [];
    try { stores.push(window.sessionStorage); } catch (error) {}
    try { stores.push(window.localStorage); } catch (error) {}

    for (var s = 0; s < stores.length; s += 1) {
      var store = stores[s];
      if (!store) continue;
      for (var i = 0; i < keys.length; i += 1) {
        try {
          var raw = store.getItem(keys[i]) || '';
          var parsed = parsePositiveInt(raw);
          if (parsed > 0) {
            return String(parsed);
          }
        } catch (error) {
          continue;
        }
      }
    }

    return '';
  }

  function isSuperContext() {
    var path = String(window.location.pathname || '').toLowerCase();
    return path.indexOf('/seleccionar_empresa.html') >= 0 || path.indexOf('/super_administrador.html') >= 0 || path.indexOf('/super/') === 0;
  }

  function buildTextEndpoint() {
    if (isSuperContext()) {
      return '/super/api/chat_con_ia_global/consultar';
    }
    return '/api/empresa/chat_con_inteligencia_artificial/consultar';
  }

  function buildAttachmentEndpoint() {
    if (isSuperContext()) {
      return '/super/api/chat_con_ia_global/consultar_con_adjunto';
    }
    return '/api/empresa/chat_con_inteligencia_artificial/consultar_con_adjunto';
  }

  function getEndpointLabel() {
    return isSuperContext() ? 'chat global de super administrador' : 'chat empresarial';
  }

  function normalize(text) {
    return String(text || '').trim();
  }

  function getAssistantMode() {
    var modeEl = document.getElementById(MODE_ID);
    var value = normalize(modeEl && modeEl.value);
    if (value === 'reportes') return 'reportes';
    return value === 'ayudante' ? 'ayudante' : 'operativo';
  }

  function isReportMode() {
    return getAssistantMode() === 'reportes';
  }

  function buildReportesEndpoint() {
    return '/api/empresa/reportes_ia_chat';
  }

  function getCurrentAttachment() {
    return state.selectedAttachment || null;
  }

  function describeAttachment(file) {
    if (!file) return '';
    var kb = Math.max(1, Math.round((Number(file.size) || 0) / 1024));
    return String(file.name || 'archivo') + ' (' + kb + ' KB)';
  }

  function renderAttachmentState() {
    var labelEl = document.getElementById(ATTACHMENT_NAME_ID);
    var clearBtn = document.getElementById(CLEAR_ATTACHMENT_ID);
    var file = getCurrentAttachment();

    if (labelEl) {
      if (file) {
        labelEl.textContent = 'Adjunto listo: ' + describeAttachment(file);
        labelEl.classList.remove('is-hidden');
      } else {
        labelEl.textContent = '';
        labelEl.classList.add('is-hidden');
      }
    }
    if (clearBtn) {
      clearBtn.classList.toggle('is-hidden', !file);
    }
  }

  function isSpeechRecognitionSupported() {
    return !!(window.SpeechRecognition || window.webkitSpeechRecognition);
  }

  function isSpeechSynthesisSupported() {
    return !!(window.speechSynthesis && typeof window.SpeechSynthesisUtterance === 'function');
  }

  function updateVoiceButtons(micBtn, voiceBtn, convBtn) {
    if (micBtn) {
      micBtn.innerHTML = ICON_MIC;
      micBtn.title = state.listening ? 'Detener dictado' : 'Dictar con el micrófono';
      micBtn.setAttribute('aria-label', state.listening ? 'Detener dictado' : 'Dictar mensaje');
      micBtn.setAttribute('aria-pressed', state.listening ? 'true' : 'false');
      micBtn.disabled = state.loading || !isSpeechRecognitionSupported();
      if (!isSpeechRecognitionSupported()) {
        micBtn.title = 'Dictado no disponible en este navegador';
      }
    }
    if (voiceBtn) {
      voiceBtn.innerHTML = ICON_SPK;
      voiceBtn.title = state.voiceEnabled ? 'Desactivar voz del asistente' : 'Activar voz del asistente';
      voiceBtn.setAttribute('aria-label', state.voiceEnabled ? 'Voz del asistente activada' : 'Activar voz del asistente');
      voiceBtn.setAttribute('aria-pressed', state.voiceEnabled ? 'true' : 'false');
      voiceBtn.disabled = state.loading || !isSpeechSynthesisSupported();
      if (!isSpeechSynthesisSupported()) {
        voiceBtn.title = 'Texto a voz no disponible en este navegador';
      }
    }
    if (convBtn) {
      convBtn.innerHTML = ICON_CONV;
      convBtn.title = state.conversationMode ? 'Modo conversación activo' : 'Modo conversación (dictado y voz del asistente)';
      convBtn.setAttribute('aria-label', 'Modo conversación');
      convBtn.setAttribute('aria-pressed', state.conversationMode ? 'true' : 'false');
    }
  }

  function speakAssistantText(text) {
    var readAloud = state.voiceEnabled || state.conversationMode;
    if (!readAloud || !isSpeechSynthesisSupported() || !text) return;
    try {
      window.speechSynthesis.cancel();
      var utterance = new SpeechSynthesisUtterance(String(text));
      utterance.lang = 'es-CO';
      utterance.rate = 1;
      utterance.pitch = 1;
      window.speechSynthesis.speak(utterance);
    } catch (err) {
      console.warn('No se pudo reproducir voz:', err);
    }
  }

  function setupSpeechRecognition(input, micBtn, voiceBtn, convBtn) {
    if (!micBtn || !input || !isSpeechRecognitionSupported()) return;
    var SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
    var recognition = new SpeechRecognition();
    recognition.lang = 'es-CO';
    recognition.interimResults = true;
    recognition.continuous = false;
    var finalText = '';

    function setListening(on) {
      state.listening = !!on;
      updateVoiceButtons(micBtn, voiceBtn || document.getElementById(VOICE_ID), convBtn || document.getElementById(CONV_ID));
    }

    recognition.onresult = function (event) {
      var interimText = '';
      for (var i = event.resultIndex; i < event.results.length; i += 1) {
        var result = event.results[i];
        var transcript = String((result[0] && result[0].transcript) || '');
        if (result.isFinal) {
          finalText += transcript;
        } else {
          interimText += transcript;
        }
      }
      var current = String(input.value || '').trim();
      input.value = (current ? current + ' ' : '') + String(finalText + interimText).trim();
    };

    recognition.onerror = function () {
      setListening(false);
      setNotice('Error de micrófono.');
    };

    recognition.onend = function () {
      setListening(false);
      finalText = '';
    };

    micBtn.addEventListener('click', function () {
      if (state.loading) return;
      if (state.listening) {
        try { recognition.stop(); } catch (e) { }
        setListening(false);
        return;
      }
      finalText = '';
      try {
        recognition.start();
        setListening(true);
      } catch (err) {
        setListening(false);
        setNotice('No se pudo iniciar dictado.');
      }
    });
  }

  function setupSpeechControls(input, micBtn, voiceBtn, convBtn) {
    updateVoiceButtons(micBtn, voiceBtn, convBtn);
    if (voiceBtn) {
      voiceBtn.addEventListener('click', function () {
        if (state.loading) return;
        state.voiceEnabled = !state.voiceEnabled;
        if (!state.voiceEnabled) {
          state.conversationMode = false;
        }
        updateVoiceButtons(micBtn, voiceBtn, convBtn);
        setNotice(state.voiceEnabled ? 'Respuestas de voz activadas (API del navegador: síntesis).' : 'Respuestas de voz desactivadas.');
      });
    }
    if (convBtn) {
      convBtn.addEventListener('click', function () {
        if (state.loading) return;
        state.conversationMode = !state.conversationMode;
        if (state.conversationMode) {
          state.voiceEnabled = true;
          setNotice('Modo conversación: lectura automática de respuestas. Dictado y voz usan la Web Speech API del navegador (sin coste extra).');
        } else {
          setNotice('Modo conversación desactivado.');
        }
        updateVoiceButtons(micBtn, voiceBtn, convBtn);
      });
    }
    setupSpeechRecognition(input, micBtn, voiceBtn, convBtn);
  }

  function syncModeUI() {
    var modeEl = document.getElementById(MODE_ID);
    var attachBtn = document.getElementById(ATTACH_BTN_ID);
    var clearBtn = document.getElementById(CLEAR_ATTACHMENT_ID);
    var attachName = document.getElementById(ATTACHMENT_NAME_ID);
    var reportOption = modeEl && modeEl.querySelector('option[value="reportes"]');
    var reportMode = isReportMode();
    var superContext = isSuperContext();

    if (reportOption) {
      reportOption.hidden = superContext;
      reportOption.disabled = superContext;
      if (superContext && normalize(modeEl.value) === 'reportes') {
        modeEl.value = 'operativo';
        reportMode = false;
      }
    }

    if (attachBtn) attachBtn.disabled = reportMode;
    if (clearBtn) clearBtn.disabled = reportMode;
    if (attachName) {
      if (reportMode) {
        attachName.textContent = 'Modo reportes: el asistente usara el flujo centralizado de reportes y exportaciones.';
        attachName.classList.remove('is-hidden');
      } else if (!getCurrentAttachment()) {
        attachName.textContent = '';
        attachName.classList.add('is-hidden');
      }
    }
    if (reportMode && getCurrentAttachment()) {
      clearAttachmentSelection();
    }
  }

  function clearAttachmentSelection() {
    state.selectedAttachment = null;
    var input = document.getElementById(ATTACHMENT_INPUT_ID);
    if (input) {
      input.value = '';
    }
    renderAttachmentState();
  }

  function extractPCSActionBlock(text) {
    var raw = normalize(text);
    if (!raw) {
      return { clean: raw, proposal: null };
    }
    var marker = '\nPCS_ACTION\n';
    var idx = raw.lastIndexOf(marker);
    if (idx < 0) {
      return { clean: raw, proposal: null };
    }
    var before = raw.slice(0, idx).trim();
    var after = raw.slice(idx + marker.length).trim();
    if (!after) {
      return { clean: before, proposal: null };
    }
    try {
      var proposal = JSON.parse(after);
      if (!proposal || proposal.version !== 1 || !Array.isArray(proposal.actions)) {
        return { clean: raw, proposal: null };
      }
      return { clean: before, proposal: proposal };
    } catch (e) {
      return { clean: raw, proposal: null };
    }
  }

  function createActionProposalElement(proposal, proposalIndex) {
    var section = document.createElement('section');
    section.className = 'ai-action-card';
    section.dataset.actionMsgIdx = String(proposalIndex);

    var title = document.createElement('div');
    title.className = 'ai-action-title';
    title.textContent = 'Acciones sugeridas (requieren confirmacion)';
    section.appendChild(title);

    if (proposal.note) {
      var note = document.createElement('div');
      note.className = 'ai-action-note';
      note.textContent = normalize(proposal.note);
      section.appendChild(note);
    }

    (Array.isArray(proposal.actions) ? proposal.actions : []).forEach(function (act, index) {
      var actionItem = document.createElement('div');
      actionItem.className = 'ai-action-item';

      var head = document.createElement('div');
      head.className = 'ai-action-head';

      var titleText = document.createElement('b');
      titleText.textContent = normalize(act.title) || ('Accion ' + (index + 1));
      head.appendChild(titleText);

      var mini = document.createElement('span');
      mini.className = 'ai-action-mini';
      var method = normalize(act.method).toUpperCase() || 'POST';
      mini.textContent = method + ' ' + normalize(act.endpoint);
      head.appendChild(mini);

      actionItem.appendChild(head);

      if (act.body != null) {
        var bodyPre = document.createElement('pre');
        bodyPre.className = 'ai-action-body';
        try {
          bodyPre.textContent = JSON.stringify(act.body, null, 2);
        } catch (e) {
          bodyPre.textContent = String(act.body);
        }
        actionItem.appendChild(bodyPre);
      }

      section.appendChild(actionItem);
    });

    var actionsBar = document.createElement('div');
    actionsBar.className = 'ai-action-actions';

    var confirm = document.createElement('button');
    confirm.type = 'button';
    confirm.className = 'btn';
    confirm.dataset.actionConfirm = String(proposalIndex);
    confirm.textContent = 'Confirmar y ejecutar';
    actionsBar.appendChild(confirm);

    var cancel = document.createElement('button');
    cancel.type = 'button';
    cancel.className = 'btn secondary';
    cancel.dataset.actionCancel = String(proposalIndex);
    cancel.textContent = 'Cancelar';
    actionsBar.appendChild(cancel);

    section.appendChild(actionsBar);
    return section;
  }

  function appendMessage(author, text, messageType, actionProposal) {
    var messagesEl = document.getElementById(MESSAGES_ID);
    if (!messagesEl || !text) return;
    var item = document.createElement('div');
    item.className = 'ai-chat-message ' + author;
    if (messageType === 'error') {
      item.classList.add('error');
    }

    var textNode = document.createElement('div');
    textNode.textContent = String(text);
    item.appendChild(textNode);

    if (actionProposal && Array.isArray(actionProposal.actions) && actionProposal.actions.length) {
      var proposalIndex = state.proposals.length;
      state.proposals.push(actionProposal);
      item.dataset.proposalIndex = String(proposalIndex);
      item.appendChild(createActionProposalElement(actionProposal, proposalIndex));
    }

    messagesEl.appendChild(item);
    scrollChatToBottom();
  }

  function setNotice(message, isWarning) {
    var noticeEl = document.getElementById(NOTICE_ID);
    if (!noticeEl) return;
    if (!message) {
      noticeEl.classList.add('is-hidden');
      noticeEl.textContent = '';
      return;
    }
    noticeEl.textContent = String(message);
    noticeEl.classList.remove('is-hidden');
    noticeEl.classList.toggle('ai-chat-notice-warning', !!isWarning);
  }

  function getCurrentRole() {
    return fetch('/me', { credentials: 'same-origin' })
      .then(function (resp) {
        if (!resp.ok) return '';
        return resp.json().catch(function () { return {}; });
      })
      .then(function (data) {
        if (!data || typeof data !== 'object') return '';
        return String(data.role || data.Role || '').trim();
      })
      .catch(function () {
        return '';
      });
  }

  function updateAccessInfo(role) {
    var info = role ? 'Rol actual: ' + role + '.' : 'Rol no detectado.';
    var hint = isSuperContext()
      ? 'Usa el asistente IA para preguntas globales, reportes y tareas administrativas del super administrador.'
      : 'Usa el asistente IA para reportes, productos, configuraciones y acciones de la empresa actual.';
    setNotice(info + ' ' + hint);
  }

  function parseErrorResponse(resp) {
    return resp.text().then(function (text) {
      var msg = normalize(text);
      if (msg) {
        try {
          var data = JSON.parse(msg);
          if (data && typeof data === 'object' && normalize(data.error)) {
            msg = normalize(data.error);
          }
        } catch (error) {
          // Mensaje plano.
        }
      }
      if (!msg) {
        msg = resp.statusText || 'Error desconocido';
      }
      if (resp.status === 401 || resp.status === 403) {
        msg = 'No tienes permiso para usar el asistente IA. Pidele a un administrador que habilite el acceso de rol para este usuario.';
      }
      throw new Error(msg);
    });
  }

  function sendQuery(query, attachment) {
    if (isReportMode()) {
      if (isSuperContext()) {
        throw new Error('El modo reportes centralizado aplica al contexto de empresa. En super administrador usa el asistente global en modo operativo o ayudante.');
      }
      var empresaIdForReports = getCurrentEmpresaId();
      if (!empresaIdForReports) {
        throw new Error('No se encontro una empresa activa para generar reportes.');
      }
      return fetch(buildReportesEndpoint(), {
        method: 'POST',
        credentials: 'same-origin',
        headers: {
          'Content-Type': 'application/json',
          'X-PCS-Source': 'ai_drawer'
        },
        body: JSON.stringify({
          empresa_id: parsePositiveInt(empresaIdForReports),
          modo: 'reporte',
          pregunta: query,
          historial: []
        })
      }).then(function (resp) {
        if (!resp.ok) {
          return parseErrorResponse(resp);
        }
        return resp.json();
      }).then(function (data) {
        if (!data || data.ok === false) {
          throw new Error((data && data.error) ? String(data.error) : 'No se pudo obtener respuesta de reportes IA.');
        }
        var text = normalize(data.respuesta || 'Reporte listo.');
        if (normalize(data.title)) {
          text += '\n\nReporte: ' + normalize(data.title);
        }
        if (normalize(data.format)) {
          text += '\nFormato: ' + String(data.format).toUpperCase();
        }
        if (normalize(data.export_url)) {
          text += '\nEnlace: ' + normalize(data.export_url);
        }
        return { clean: text, proposal: null };
      });
    }

    var endpoint = attachment ? buildAttachmentEndpoint() : buildTextEndpoint();
    var mode = getAssistantMode();
    var pageContext = String(window.location.pathname || '') + String(window.location.search || '');
    var body = {
      pregunta: query,
      modo_asistente: mode
    };

    if (pageContext) {
      body.pagina_contexto = pageContext;
    }
    if (!isSuperContext()) {
      var empresaId = getCurrentEmpresaId();
      if (!empresaId) {
        throw new Error('No se encontro una empresa activa. Ingresa desde el contexto de una empresa para usar el chat IA empresarial.');
      }
      body.empresa_id = parsePositiveInt(empresaId);
    }

    var options = {
      method: 'POST',
      credentials: 'same-origin',
      headers: {
        'X-PCS-Source': 'ai_drawer'
      }
    };

    if (attachment) {
      var formData = new FormData();
      formData.set('pregunta', query);
      formData.set('modo_asistente', mode);
      if (pageContext) {
        formData.set('pagina_contexto', pageContext);
      }
      if (!isSuperContext()) {
        formData.set('empresa_id', String(body.empresa_id));
        formData.set('use_gpt55', '1');
      }
      formData.set('file', attachment, attachment.name || 'adjunto');
      options.body = formData;
    } else {
      options.headers['Content-Type'] = 'application/json';
      options.body = JSON.stringify(body);
    }

    return fetch(endpoint, options)
      .then(function (resp) {
        if (!resp.ok) {
          return parseErrorResponse(resp);
        }
        return resp.json();
      })
      .then(function (data) {
        if (!data || data.ok === false) {
          var detail = (data && data.error) ? String(data.error) : 'No se pudo obtener respuesta de IA.';
          throw new Error(detail);
        }
        var answer = String(data.respuesta || data.answer || data.message || 'La IA respondio sin contenido.');
        return extractPCSActionBlock(answer);
      });
  }

  function handleSubmit(event) {
    event.preventDefault();
    if (state.loading) return;
    var input = document.getElementById(INPUT_ID);
    if (!input) return;

    var query = String(input.value || '').trim();
    var attachment = getCurrentAttachment();
    if (!query) return;

    input.value = '';
    appendMessage('user', attachment ? (query + '\n\n[Adjunto: ' + describeAttachment(attachment) + ']') : query);
    setNotice(attachment ? 'Procesando consulta con adjunto...' : 'Procesando tu consulta...');
    state.loading = true;
    updateVoiceButtons(document.getElementById(MIC_ID), document.getElementById(VOICE_ID), document.getElementById(CONV_ID));

    sendQuery(query, attachment).then(function (result) {
      appendMessage('assistant', result.clean, null, result.proposal);
      speakAssistantText(result.clean);
      setNotice('Respuesta lista. Puedes seguir escribiendo otra consulta.');
      clearAttachmentSelection();
    }).catch(function (err) {
      appendMessage('assistant', err.message || 'Error al procesar la consulta.', 'error');
      setNotice('No se pudo completar la solicitud. ' + String(err.message || ''), true);
    }).finally(function () {
      state.loading = false;
      updateVoiceButtons(document.getElementById(MIC_ID), document.getElementById(VOICE_ID), document.getElementById(CONV_ID));
    });
  }

  function executeActionProposal(msgIdx) {
    var proposal = state.proposals[msgIdx];
    if (!proposal || !Array.isArray(proposal.actions) || !proposal.actions.length) return;
    if (state.loading) return;
    state.loading = true;
    updateVoiceButtons(document.getElementById(MIC_ID), document.getElementById(VOICE_ID), document.getElementById(CONV_ID));
    setNotice('Ejecutando acciones sugeridas...');

    var messagesEl = document.getElementById(MESSAGES_ID);
    var messageEl = messagesEl && messagesEl.querySelector('[data-proposal-index="' + msgIdx + '"]');

    return Promise.resolve().then(function () {
      var chain = Promise.resolve();
      proposal.actions.forEach(function (act) {
        chain = chain.then(function () {
          var endpoint = normalize(act.endpoint);
          var method = normalize(act.method).toUpperCase() || 'POST';
          if (method === 'DELETE') {
            throw new Error('Accion bloqueada: DELETE no esta permitida desde el chat IA.');
          }
          if (method === 'OPEN') {
            if (!endpoint || endpoint[0] !== '/') {
              throw new Error('Accion OPEN bloqueada: la URL debe ser relativa.');
            }
            window.open(endpoint, '_blank', 'noopener,noreferrer');
            appendMessage('assistant', 'Accion ejecutada: abrir vista ' + endpoint + '.');
            return;
          }
          var payload = act.body != null ? act.body : null;
          return fetch(endpoint, {
            method: method,
            credentials: 'same-origin',
            headers: {
              'Content-Type': 'application/json',
              'X-PCS-Source': 'ai_drawer'
            },
            body: payload ? JSON.stringify(payload) : null
          }).then(function (res) {
            return res.text().then(function (text) {
              if (!res.ok) {
                var detail = text || res.statusText || 'Error inesperado';
                throw new Error('Fallo al ejecutar accion: HTTP ' + res.status + ' - ' + detail);
              }
              appendMessage('assistant', 'Accion ejecutada: ' + normalize(act.title || act.endpoint) + '. Respuesta: ' + text);
            });
          });
        });
      });
      return chain;
    }).then(function () {
      if (messageEl) {
        var actionCard = messageEl.querySelector('.ai-action-card');
        if (actionCard && actionCard.parentNode) {
          actionCard.parentNode.removeChild(actionCard);
        }
      }
      state.proposals[msgIdx] = null;
      setNotice('Acciones ejecutadas correctamente.');
    }).catch(function (err) {
      appendMessage('assistant', err.message || 'Error al ejecutar las acciones.', 'error');
      setNotice('Error al ejecutar acciones: ' + String(err.message || ''), true);
    }).finally(function () {
      state.loading = false;
      updateVoiceButtons(document.getElementById(MIC_ID), document.getElementById(VOICE_ID), document.getElementById(CONV_ID));
    });
  }

  function cancelActionProposal(msgIdx) {
    var proposal = state.proposals[msgIdx];
    if (!proposal) return;
    var messagesEl = document.getElementById(MESSAGES_ID);
    var messageEl = messagesEl && messagesEl.querySelector('[data-proposal-index="' + msgIdx + '"]');
    if (messageEl) {
      var actionCard = messageEl.querySelector('.ai-action-card');
      if (actionCard && actionCard.parentNode) {
        actionCard.parentNode.removeChild(actionCard);
      }
    }
    state.proposals[msgIdx] = null;
    setNotice('Acciones canceladas.');
  }

  function setChatBackdropVisible(on) {
    var el = document.getElementById(BACKDROP_ID);
    if (!el) return;
    el.classList.toggle('is-visible', !!on);
    el.setAttribute('aria-hidden', on ? 'false' : 'true');
  }

  function setChatBodyScrollLock(on) {
    if (isMobileChatViewport()) {
      document.body.style.overflow = on ? 'hidden' : '';
    }
  }

  function closeChatDrawerFully() {
    var drawer = document.getElementById(DRAWER_ID);
    var toggle = document.getElementById(TOGGLE_ID);
    var minibar = document.getElementById(MINIBAR_ID);
    if (!drawer || !toggle) return;
    drawer.classList.remove('open');
    drawer.classList.remove('minimized');
    if (minibar) minibar.hidden = true;
    toggle.classList.remove('is-drawer-open');
    toggle.setAttribute('aria-expanded', 'false');
    setChatBackdropVisible(false);
    setChatBodyScrollLock(false);
  }

  function openChatDrawerFromUser() {
    var drawer = document.getElementById(DRAWER_ID);
    var toggle = document.getElementById(TOGGLE_ID);
    var minibar = document.getElementById(MINIBAR_ID);
    if (!drawer || !toggle) return;
    drawer.classList.remove('minimized');
    if (minibar) minibar.hidden = true;
    drawer.classList.add('open');
    toggle.classList.add('is-drawer-open');
    toggle.setAttribute('aria-expanded', 'true');
    setChatBackdropVisible(true);
    setChatBodyScrollLock(true);
    window.setTimeout(function () {
      var inp = document.getElementById(INPUT_ID);
      if (inp) inp.focus();
    }, 50);
  }

  function minimizeChatDrawer() {
    var drawer = document.getElementById(DRAWER_ID);
    var toggle = document.getElementById(TOGGLE_ID);
    var minibar = document.getElementById(MINIBAR_ID);
    if (!drawer || !toggle) return;
    if (isMobileChatViewport()) {
      closeChatDrawerFully();
      return;
    }
    drawer.classList.remove('open');
    drawer.classList.add('minimized');
    if (minibar) minibar.hidden = false;
    toggle.classList.remove('is-drawer-open');
    toggle.setAttribute('aria-expanded', 'false');
    setChatBackdropVisible(false);
    setChatBodyScrollLock(false);
  }

  function initDrawer() {
    var toggle = document.getElementById(TOGGLE_ID);
    var drawer = document.getElementById(DRAWER_ID);
    var closeBtn = document.getElementById(CLOSE_ID);
    var minimizeBtn = document.getElementById(MINIMIZE_ID);
    var minibarExpand = document.getElementById(MINIBAR_EXPAND_ID);
    var backdrop = document.getElementById(BACKDROP_ID);
    var form = document.getElementById(FORM_ID);
    var messagesEl = document.getElementById(MESSAGES_ID);
    var hintToggle = document.getElementById(HINT_TOGGLE_ID);
    var hints = document.getElementById(HINTS_ID);
    var attachInput = document.getElementById(ATTACHMENT_INPUT_ID);
    var attachBtn = document.getElementById(ATTACH_BTN_ID);
    var clearAttachBtn = document.getElementById(CLEAR_ATTACHMENT_ID);
    var modeEl = document.getElementById(MODE_ID);
    var input = document.getElementById(INPUT_ID);

    if (!toggle || !drawer || !closeBtn || !form || !messagesEl) return;

    toggle.addEventListener('click', function () {
      if (drawer.classList.contains('open')) {
        closeChatDrawerFully();
        return;
      }
      openChatDrawerFromUser();
    });

    closeBtn.addEventListener('click', function () {
      closeChatDrawerFully();
    });

    if (minimizeBtn) {
      minimizeBtn.addEventListener('click', function () {
        minimizeChatDrawer();
      });
    }

    if (minibarExpand) {
      minibarExpand.addEventListener('click', function () {
        openChatDrawerFromUser();
      });
    }

    if (backdrop) {
      backdrop.addEventListener('click', function () {
        closeChatDrawerFully();
      });
    }

    form.addEventListener('submit', handleSubmit);
    if (modeEl) {
      modeEl.addEventListener('change', function () {
        syncModeUI();
        setNotice(isReportMode()
          ? 'Modo reportes activo. Este chat central usara el flujo de reportes y exportaciones de la empresa.'
          : 'Modo actualizado. Puedes seguir consultando normalmente.');
      });
    }
    if (hintToggle && hints) {
      hintToggle.addEventListener('click', function () {
        hints.classList.toggle('is-hidden');
        hintToggle.textContent = hints.classList.contains('is-hidden') ? 'Ver ejemplos' : 'Ocultar ejemplos';
      });
    }

    if (attachBtn && attachInput) {
      attachBtn.addEventListener('click', function () {
        if (state.loading) return;
        if (isReportMode()) {
          setNotice('El modo reportes no admite adjuntos en este flujo.', true);
          return;
        }
        attachInput.click();
      });
      attachInput.addEventListener('change', function () {
        var file = attachInput.files && attachInput.files[0] ? attachInput.files[0] : null;
        if (!file) {
          clearAttachmentSelection();
          return;
        }
        if (Number(file.size) > MAX_ATTACHMENT_BYTES) {
          clearAttachmentSelection();
          setNotice('El archivo supera el maximo permitido de 8 MB.', true);
          return;
        }
        state.selectedAttachment = file;
        renderAttachmentState();
        setNotice('Adjunto listo para enviar: ' + describeAttachment(file));
      });
    }

    if (clearAttachBtn) {
      clearAttachBtn.addEventListener('click', function () {
        if (state.loading) return;
        clearAttachmentSelection();
        setNotice('Adjunto removido.');
      });
    }

    setupSpeechControls(input, document.getElementById(MIC_ID), document.getElementById(VOICE_ID), document.getElementById(CONV_ID));

    messagesEl.addEventListener('click', function (event) {
      var target = event.target;
      if (!target) return;
      var confirmButton = target.closest('button[data-action-confirm]');
      var cancelButton = target.closest('button[data-action-cancel]');
      if (confirmButton) {
        executeActionProposal(parseInt(confirmButton.dataset.actionConfirm, 10));
      } else if (cancelButton) {
        cancelActionProposal(parseInt(cancelButton.dataset.actionCancel, 10));
      }
    });

    document.addEventListener('keydown', function (event) {
      if (event.key !== 'Escape') return;
      var d = document.getElementById(DRAWER_ID);
      if (!d || (!d.classList.contains('open') && !d.classList.contains('minimized'))) return;
      closeChatDrawerFully();
    });
    if (input) {
      input.addEventListener('keydown', function (event) {
        if (event.key === 'Enter' && !event.shiftKey) {
          event.preventDefault();
          form.requestSubmit();
        }
      });
    }

    window.addEventListener('message', function (event) {
      var data = event && event.data;
      if (!data || data.type !== 'pcs-ai-drawer-open') return;
      openChatDrawerFromUser();
      if (modeEl && normalize(data.mode)) {
        modeEl.value = normalize(data.mode);
        syncModeUI();
      }
      if (input && normalize(data.prompt)) {
        input.value = normalize(data.prompt);
      }
      window.setTimeout(function () {
        if (input) input.focus();
      }, 50);
    });

    if (!messagesEl.querySelector('.ai-chat-message')) {
      appendMessage('assistant', 'Asistente IA disponible para ' + getEndpointLabel() + '. ' + (isSuperContext() ? 'Consulta datos globales de super administrador.' : 'Consulta datos de la empresa actual y solicita acciones administrativas. Puedes asignar tareas, enviar mensajes, crear pedidos y terminar ventas segun tu rol.'));
    }

    renderAttachmentState();
    syncModeUI();
    getCurrentRole().then(updateAccessInfo);
  }

  document.addEventListener('DOMContentLoaded', initDrawer);
})();
