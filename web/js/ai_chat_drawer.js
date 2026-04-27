(function () {
  var DRAWER_ID = 'aiChatDrawer';
  var TOGGLE_ID = 'openAIDrawer';
  var CLOSE_ID = 'closeAIDrawer';
  var FORM_ID = 'aiChatForm';
  var INPUT_ID = 'aiChatInput';
  var MESSAGES_ID = 'aiChatMessages';
  var NOTICE_ID = 'aiChatNotice';
  var HINT_TOGGLE_ID = 'aiChatHintToggle';
  var HINTS_ID = 'aiChatHints';

  var state = {
    proposals: [],
    loading: false
  };

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

  function buildEndpoint() {
    if (isSuperContext()) {
      return '/super/api/chat_con_ia_global/consultar';
    }
    return '/api/empresa/chat_con_inteligencia_artificial/consultar';
  }

  function getEndpointLabel() {
    return isSuperContext() ? 'chat global de super administrador' : 'chat empresarial';
  }

  function normalize(text) {
    return String(text || '').trim();
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
    title.textContent = 'Acciones sugeridas (requieren confirmación)';
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
      titleText.textContent = normalize(act.title) || ('Acción ' + (index + 1));
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
    messagesEl.scrollTop = messagesEl.scrollHeight;
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

  function sendQuery(query) {
    var endpoint = buildEndpoint();
    var body = { pregunta: query };
    var pageContext = String(window.location.pathname || '') + String(window.location.search || '');
    if (pageContext) {
      body.pagina_contexto = pageContext;
    }
    if (!isSuperContext()) {
      var empresaId = getCurrentEmpresaId();
      if (!empresaId) {
        throw new Error('No se encontró una empresa activa. Ingresa desde el contexto de una empresa para usar el chat IA empresarial.');
      }
      body.empresa_id = parsePositiveInt(empresaId);
    }

    return fetch(endpoint, {
      method: 'POST',
      credentials: 'same-origin',
      headers: {
        'Content-Type': 'application/json',
        'X-PCS-Source': 'ai_drawer'
      },
      body: JSON.stringify(body)
    })
      .then(function (resp) {
        if (!resp.ok) {
          return resp.json().catch(function () { return null; }).then(function (data) {
            var errorMessage = (data && data.error) ? String(data.error) : resp.statusText || 'Error desconocido';
            if (resp.status === 401 || resp.status === 403) {
              errorMessage = 'No tienes permiso para usar el asistente IA. Pídele a un administrador que habilite el acceso de rol para este usuario.';
            }
            throw new Error(errorMessage);
          });
        }
        return resp.json();
      })
      .then(function (data) {
        if (!data || data.ok === false) {
          var detail = (data && data.error) ? String(data.error) : 'No se pudo obtener respuesta de IA.';
          throw new Error(detail);
        }
        var answer = String(data.respuesta || data.answer || data.message || 'La IA respondió sin contenido.');
        return extractPCSActionBlock(answer);
      });
  }

  function handleSubmit(event) {
    event.preventDefault();
    if (state.loading) return;
    var input = document.getElementById(INPUT_ID);
    if (!input) return;
    var query = String(input.value || '').trim();
    if (!query) return;
    input.value = '';
    appendMessage('user', query);
    setNotice('Procesando tu consulta…');
    state.loading = true;
    sendQuery(query).then(function (result) {
      appendMessage('assistant', result.clean, null, result.proposal);
      setNotice('Respuesta lista. Puedes seguir escribiendo otra consulta.');
    }).catch(function (err) {
      appendMessage('assistant', err.message || 'Error al procesar la consulta.', 'error');
      setNotice('No se pudo completar la solicitud. ' + String(err.message || ''), true);
    }).finally(function () {
      state.loading = false;
    });
  }

  function executeActionProposal(msgIdx) {
    var proposal = state.proposals[msgIdx];
    if (!proposal || !Array.isArray(proposal.actions) || !proposal.actions.length) return;
    if (state.loading) return;
    state.loading = true;
    setNotice('Ejecutando acciones sugeridas…');

    var messagesEl = document.getElementById(MESSAGES_ID);
    var messageEl = messagesEl && messagesEl.querySelector('[data-proposal-index="' + msgIdx + '"]');

    return Promise.resolve().then(function () {
      var chain = Promise.resolve();
      proposal.actions.forEach(function (act) {
        chain = chain.then(function () {
          var endpoint = normalize(act.endpoint);
          var method = normalize(act.method).toUpperCase() || 'POST';
          if (method === 'DELETE') {
            throw new Error('Acción bloqueada: DELETE no está permitida desde el chat IA.');
          }
          if (method === 'OPEN') {
            if (!endpoint || endpoint[0] !== '/') {
              throw new Error('Acción OPEN bloqueada: la URL debe ser relativa.');
            }
            window.open(endpoint, '_blank', 'noopener,noreferrer');
            appendMessage('assistant', 'Acción ejecutada: abrir vista ' + endpoint + '.');
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
                throw new Error('Fallo al ejecutar acción: HTTP ' + res.status + ' — ' + detail);
              }
              appendMessage('assistant', 'Acción ejecutada: ' + normalize(act.title || act.endpoint) + '. Respuesta: ' + text);
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

  function initDrawer() {
    var toggle = document.getElementById(TOGGLE_ID);
    var drawer = document.getElementById(DRAWER_ID);
    var closeBtn = document.getElementById(CLOSE_ID);
    var form = document.getElementById(FORM_ID);
    var messagesEl = document.getElementById(MESSAGES_ID);
    var hintToggle = document.getElementById(HINT_TOGGLE_ID);
    var hints = document.getElementById(HINTS_ID);

    if (!toggle || !drawer || !closeBtn || !form || !messagesEl) return;

    toggle.addEventListener('click', function () {
      var expanded = !drawer.classList.contains('open');
      drawer.classList.toggle('open');
      toggle.setAttribute('aria-expanded', expanded ? 'true' : 'false');
      if (expanded) {
        window.setTimeout(function () {
          var input = document.getElementById(INPUT_ID);
          if (input) input.focus();
        }, 50);
      }
    });

    closeBtn.addEventListener('click', function () {
      drawer.classList.remove('open');
      toggle.setAttribute('aria-expanded', 'false');
    });

    form.addEventListener('submit', handleSubmit);
    if (hintToggle && hints) {
      hintToggle.addEventListener('click', function () {
        hints.classList.toggle('is-hidden');
        hintToggle.textContent = hints.classList.contains('is-hidden') ? 'Ver ejemplos' : 'Ocultar ejemplos';
      });
    }

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
      if (event.key === 'Escape' && drawer.classList.contains('open')) {
        drawer.classList.remove('open');
        toggle.setAttribute('aria-expanded', 'false');
      }
    });

    if (!messagesEl.querySelector('.ai-chat-message')) {
      appendMessage('assistant', 'Asistente IA disponible para ' + getEndpointLabel() + '. ' + (isSuperContext() ? 'Consulta datos globales de super administrador.' : 'Consulta datos de la empresa actual y solicita acciones administrativas. Puedes asignar tareas, enviar mensajes, crear pedidos y terminar ventas según tu rol.'));
    }

    getCurrentRole().then(updateAccessInfo);
  }

  document.addEventListener('DOMContentLoaded', initDrawer);
})();
