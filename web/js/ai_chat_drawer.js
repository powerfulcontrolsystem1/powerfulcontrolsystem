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
  var CLEAR_CHAT_ID = 'aiChatNewBtn';
  var BACKDROP_ID = 'aiChatBackdrop';
  var MINIMIZE_ID = 'aiChatMinimize';
  var MINIBAR_ID = 'aiChatMinibar';
  var MINIBAR_EXPAND_ID = 'aiChatMinibarExpand';
  var MESSAGES_ID = 'aiChatMessages';
  var NOTICE_ID = 'aiChatNotice';
  var HINT_TOGGLE_ID = 'aiChatHintToggle';
  var HINTS_ID = 'aiChatHints';
  var CONFIG_PANEL_ID = 'aiChatCompactConfig';
  var CONFIG_CLOSE_ID = 'aiChatCompactConfigClose';
  var CONFIG_SAVE_ID = 'aiChatCompactConfigSave';
  var CONFIG_CHAT_ENABLED_ID = 'aiChatCompactConfigEnabled';
  var CONFIG_ROBOT_ENABLED_ID = 'aiChatCompactConfigRobotEnabled';
  var CONFIG_VOICE_ID = 'aiChatCompactConfigVoice';
  var CONFIG_ROBOT_VOICE_ID = 'aiChatCompactConfigRobotVoice';
  var DOCUMENT_TOOLS_ID = 'aiChatDocumentTools';
  var DOCUMENT_FORMAT_ID = 'aiChatDocumentFormat';
  var DOCUMENT_DOWNLOAD_ID = 'aiChatDocumentDownload';
  var DOCUMENT_EMAIL_ID = 'aiChatDocumentEmail';
  var DOCUMENT_WHATSAPP_ID = 'aiChatDocumentWhatsApp';
  var MAX_ATTACHMENT_BYTES = 8 * 1024 * 1024;
  var CHAT_PREFS_ENDPOINT = '/api/chat_flotante/preferencias';

  var state = {
    proposals: [],
    exportables: [],
    loading: false,
    selectedAttachment: null,
    chatEnabled: true,
    robotEnabled: true,
    voiceEnabled: false,
    listening: false,
    conversationMode: false,
    voiceServerAvailable: false,
    voiceServerChecked: false,
    voiceOutputMode: 'computer',
    computerVoiceGender: 'female',
    voiceServerAudio: null,
    voicePlaybackVersion: 0,
    activeSpeechRecognition: null,
    activeSpeechSource: '',
    preferredConversationMicId: MIC_ID,
    conversationResumeTimer: null,
    voiceQueuePromise: Promise.resolve(),
    voiceQueueVersion: 0,
    streamingSpeechBuffer: '',
    robotVoice: 'es-CO',
    robotAssistantVisible: false,
    robotMoodTimer: null,
    lastResponseModelMeta: null,
    generatedDocument: null,
    shareArtifact: null,
    setupWizard: null
  };

  var ICON_MIC = '<svg viewBox="0 0 24 24" width="22" height="22" aria-hidden="true"><path fill="currentColor" d="M12 14c1.66 0 3-1.34 3-3V5c0-1.66-1.34-3-3-3S9 3.34 9 5v6c0 1.66 1.34 3 3 3zm5-3c0 2.76-2.24 5-5 5s-5-2.24-5-5H5c0 3.53 2.61 6.43 6 6.92V21h2v-3.08c3.39-.49 6-3.39 6-6.92h-2z"/></svg>';
  var ICON_SPK = '<svg viewBox="0 0 24 24" width="22" height="22" aria-hidden="true"><path fill="currentColor" d="M3 9v6h4l5 5V4L7 9H3zm13.5 3c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.25 2.5-4.02zM14 3.23v2.06c2.89.86 5 3.54 5 6.71s-2.11 5.85-5 6.71v2.06c4.01-.91 7-4.49 7-8.77s-2.99-7.86-7-8.77z"/></svg>';
  var ICON_CONV = '<svg viewBox="0 0 24 24" width="22" height="22" aria-hidden="true"><path fill="currentColor" d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm0 14H6l-2 2V4h16v12z"/></svg>';
  var ICON_STOP = '<svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true"><path fill="currentColor" d="M6 6h12v12H6z"/></svg>';
  var ROBOT_PANEL_ID = 'robotInlineChatPanel';
  var ROBOT_STATUS_ID = 'robotInlineStatus';
  var ROBOT_ASSISTANT_BUBBLE_ID = 'robotAssistantBubble';
  var ROBOT_USER_BUBBLE_ID = 'robotUserBubble';
  var ROBOT_INLINE_FORM_ID = 'robotInlineForm';
  var ROBOT_INLINE_INPUT_ID = 'robotInlineInput';
  var ROBOT_INLINE_SEND_ID = 'robotInlineSend';
  var ROBOT_INLINE_STOP_VOICE_ID = 'robotInlineStopVoice';
  var ROBOT_INLINE_MIC_ID = 'robotInlineMic';
  var ROBOT_ACTIONS_ID = 'robotAssistantActions';
  var ROBOT_HIDE_ID = 'robotHideBtn';
  var ROBOT_SHOW_ID = 'robotShowBtn';
  var ROBOT_SVG = '<div id="robotAvatarGraphic" class="robot-3d-avatar robot-mood-idle" aria-hidden="true" data-mood="idle">' +
    '<div class="robot-3d-stage">' +
    '<span class="robot-3d-shadow"></span>' +
    '<span class="robot-3d-signal robot-3d-signal-a"></span>' +
    '<span class="robot-3d-signal robot-3d-signal-b"></span>' +
    '<span class="robot-3d-antenna"><span></span></span>' +
    '<div class="robot-3d-head">' +
    '<span class="robot-3d-ear robot-3d-ear-left"></span>' +
    '<span class="robot-3d-ear robot-3d-ear-right"></span>' +
    '<div class="robot-3d-face">' +
    '<span class="robot-3d-eye robot-3d-eye-left"></span>' +
    '<span class="robot-3d-eye robot-3d-eye-right"></span>' +
    '<span class="robot-3d-mouth"></span>' +
    '</div>' +
    '</div>' +
    '<div class="robot-3d-body">' +
    '<span class="robot-3d-core"></span>' +
    '<span class="robot-3d-light robot-3d-light-a"></span>' +
    '<span class="robot-3d-light robot-3d-light-b"></span>' +
    '<span class="robot-3d-light robot-3d-light-c"></span>' +
    '</div>' +
    '<span class="robot-3d-arm robot-3d-arm-left"></span>' +
    '<span class="robot-3d-arm robot-3d-arm-right"></span>' +
    '<span class="robot-3d-leg robot-3d-leg-left"></span>' +
    '<span class="robot-3d-leg robot-3d-leg-right"></span>' +
    '</div>' +
    '</div>';
  var SECRETARY_SVG = '<div id="robotAvatarGraphic" class="secretary-3d-avatar robot-mood-idle" aria-hidden="true" data-mood="idle">' +
    '<div class="secretary-3d-stage">' +
    '<span class="secretary-3d-shadow"></span>' +
    '<span class="secretary-3d-side-light secretary-3d-side-light-a"></span>' +
    '<span class="secretary-3d-side-light secretary-3d-side-light-b"></span>' +
    '<div class="secretary-3d-hair secretary-3d-hair-back"></div>' +
    '<span class="secretary-3d-hair-side secretary-3d-hair-side-left"></span>' +
    '<span class="secretary-3d-hair-side secretary-3d-hair-side-right"></span>' +
    '<div class="secretary-3d-head">' +
    '<span class="secretary-3d-hair secretary-3d-bang-a"></span>' +
    '<span class="secretary-3d-hair secretary-3d-bang-b"></span>' +
    '<div class="secretary-3d-face">' +
    '<span class="secretary-3d-lash secretary-3d-lash-left"></span>' +
    '<span class="secretary-3d-lash secretary-3d-lash-right"></span>' +
    '<span class="secretary-3d-eye secretary-3d-eye-left"></span>' +
    '<span class="secretary-3d-eye secretary-3d-eye-right"></span>' +
    '<span class="secretary-3d-cheek secretary-3d-cheek-left"></span>' +
    '<span class="secretary-3d-cheek secretary-3d-cheek-right"></span>' +
    '<span class="secretary-3d-mouth"></span>' +
    '</div>' +
    '<span class="secretary-3d-ear secretary-3d-ear-left"></span>' +
    '<span class="secretary-3d-ear secretary-3d-ear-right"></span>' +
    '</div>' +
    '<span class="secretary-3d-neck"></span>' +
    '<div class="secretary-3d-body">' +
    '<span class="secretary-3d-blazer"></span>' +
    '<span class="secretary-3d-shirt"></span>' +
    '<span class="secretary-3d-lapel secretary-3d-lapel-left"></span>' +
    '<span class="secretary-3d-lapel secretary-3d-lapel-right"></span>' +
    '<span class="secretary-3d-scarf"></span>' +
    '<span class="secretary-3d-badge"></span>' +
    '<span class="secretary-3d-nameplate"></span>' +
    '<span class="secretary-3d-button secretary-3d-button-a"></span>' +
    '<span class="secretary-3d-button secretary-3d-button-b"></span>' +
    '</div>' +
    '<span class="secretary-3d-arm secretary-3d-arm-left"></span>' +
    '<span class="secretary-3d-arm secretary-3d-arm-right"></span>' +
    '<span class="secretary-3d-tablet"></span>' +
    '<span class="secretary-3d-skirt"></span>' +
    '<span class="secretary-3d-leg secretary-3d-leg-left"></span>' +
    '<span class="secretary-3d-leg secretary-3d-leg-right"></span>' +
    '<span class="secretary-3d-shoe secretary-3d-shoe-left"></span>' +
    '<span class="secretary-3d-shoe secretary-3d-shoe-right"></span>' +
    '</div>' +
    '</div>';

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

  function getPublicPortalContextConfig() {
    if (window && window.__pcsPublicChatContext && typeof window.__pcsPublicChatContext === 'object') {
      return window.__pcsPublicChatContext;
    }
    return null;
  }

  function inferPublicStoreSlugFromPath() {
    var path = String(window.location.pathname || '').trim();
    if (!path) return '';
    var match = path.match(/^\/([^\/]+)\/venta_publica\.html$/i);
    if (match && match[1]) {
      return normalize(match[1]).toLowerCase();
    }
    return '';
  }

  function getPublicEmpresaSlug() {
    var cfg = getPublicPortalContextConfig() || {};
    var params = new URLSearchParams(window.location.search || '');
    var slug = normalize(cfg.empresa_slug || cfg.slug || params.get('empresa_slug') || params.get('slug') || inferPublicStoreSlugFromPath());
    return slug.toLowerCase();
  }

  function getPublicPortalScope() {
    var cfg = getPublicPortalContextConfig() || {};
    var raw = normalize(cfg.scope).toLowerCase();
    if (raw === 'venta_publica') return 'venta_publica';
    var path = String(window.location.pathname || '').toLowerCase();
    if (path === '/venta_publica.html') return 'venta_publica';
    if (/^\/[^\/]+\/venta_publica\.html$/i.test(path)) return 'venta_publica';
    return 'portal';
  }

  function isPublicPortalContext() {
    if (isSuperContext()) return false;
    var cfg = getPublicPortalContextConfig();
    if (cfg && normalize(cfg.scope)) return true;
    var path = String(window.location.pathname || '').toLowerCase();
    return path === '/' || path === '/index.html' || path === '/venta_publica.html' || /^\/[^\/]+\/venta_publica\.html$/i.test(path);
  }

  function isPublicStoreContext() {
    return isPublicPortalContext() && getPublicPortalScope() === 'venta_publica';
  }

  function shouldAutoInjectDrawerShell() {
    try {
      return !!(window && window.__pcsAutoInjectChatShell);
    } catch (error) {
      return false;
    }
  }

  function buildPublicDrawerExamplesMarkup() {
    if (isPublicStoreContext()) {
      return '<p>Preguntas recomendadas:</p>' +
        '<ul>' +
        '<li>¿Qué productos o servicios tiene esta empresa?</li>' +
        '<li>¿Qué promociones públicas están activas?</li>' +
        '<li>¿Cuáles son los precios visibles hoy?</li>' +
        '<li>Muéstrame las páginas públicas disponibles.</li>' +
        '</ul>';
    }
    return '<p>Ejemplos de preguntas útiles:</p>' +
      '<ul>' +
      '<li>¿Qué planes manejan?</li>' +
      '<li>¿Qué módulos incluye la plataforma?</li>' +
      '<li>¿Cómo puedo empezar una prueba gratis?</li>' +
      '<li>¿Por dónde los contacto para una demostración?</li>' +
      '</ul>';
  }

  function buildPublicDrawerTitle() {
    return isPublicStoreContext() ? 'Asistente público de tienda' : 'Asistente público IA';
  }

  function buildPublicDrawerPlaceholder() {
    return isPublicStoreContext()
      ? 'Pregunta por productos, servicios, precios o promociones...'
      : 'Escribe tu consulta aquí...';
  }

  function ensureDrawerShell() {
    if (document.getElementById(TOGGLE_ID) && document.getElementById(DRAWER_ID) && document.getElementById(BACKDROP_ID)) {
      return true;
    }
    if (!shouldAutoInjectDrawerShell()) {
      return false;
    }
    if (!document.body) {
      return false;
    }
    var title = buildPublicDrawerTitle();
    var placeholder = buildPublicDrawerPlaceholder();
    var hintsMarkup = buildPublicDrawerExamplesMarkup();
    document.body.insertAdjacentHTML('beforeend',
      '<button id="' + TOGGLE_ID + '" class="ai-chat-toggle-button" aria-haspopup="dialog" aria-expanded="false" type="button">' +
        '<img class="icon" src="/img/gpt.svg" alt=""><span class="ai-chat-toggle-label">Asistente IA</span>' +
      '</button>' +
      '<div id="' + BACKDROP_ID + '" class="ai-chat-backdrop" aria-hidden="true"></div>' +
      '<section id="' + DRAWER_ID + '" class="ai-chat-drawer" role="dialog" aria-modal="true" aria-labelledby="aiChatTitle">' +
        '<div class="ai-chat-minibar" id="' + MINIBAR_ID + '" hidden>' +
          '<span class="ai-chat-minibar-label">Asistente IA</span>' +
          '<button type="button" id="' + MINIBAR_EXPAND_ID + '" class="ai-chat-minibar-btn" aria-label="Abrir asistente IA">' +
            '<svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true"><path fill="currentColor" d="M7 14l5-5 5 5H7z"/></svg>' +
          '</button>' +
        '</div>' +
        '<div class="ai-chat-drawer-surface">' +
          '<div class="ai-chat-header">' +
            '<div class="ai-chat-header-title-row">' +
              '<span class="ai-chat-header-icon" aria-hidden="true">' +
                '<svg viewBox="0 0 24 24" width="22" height="22"><path fill="currentColor" d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm0 14H6l-2 2V4h16v12z"/></svg>' +
              '</span>' +
              '<div><h2 id="aiChatTitle">' + title + '</h2></div>' +
            '</div>' +
            '<div class="ai-chat-header-actions">' +
              '<button id="' + HINT_TOGGLE_ID + '" type="button" class="btn secondary small">Ver ejemplos</button>' +
              '<button id="aiChatConfigBtn" type="button" class="ai-chat-header-icon-btn" aria-label="Configurar chat flotante" title="Configuración chat flotante">' +
                '<svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true"><path fill="currentColor" d="M12 8a4 4 0 1 0 0 8 4 4 0 0 0 0-8zm7.4 4.5c.04-.3.06-.6.06-.9s-.02-.6-.06-.9l2-1.6a.5.5 0 0 0 .1-.6l-1.8-3.1a.5.5 0 0 0-.6-.2l-2.4 1a7 7 0 0 0-1.6-1l-.4-2.6A.5.5 0 0 0 13.5 2h-3a.5.5 0 0 0-.5.4l-.4 2.6a7 7 0 0 0-1.6 1l-2.4-1a.5.5 0 0 0-.6.2L4.5 7a.5.5 0 0 0 .1.6l2 1.6c-.1.3-.1.6-.1.9s.02.6.06.9l-2 1.6a.5.5 0 0 0-.1.6l1.8 3.1a.5.5 0 0 0 .6.2l2.4-1c.5.4 1.1.7 1.6 1l.4 2.6a.5.5 0 0 0 .5.4h3a.5.5 0 0 0 .5-.4l.4-2.6c.6-.2 1.1-.5 1.6-1l2.4 1a.5.5 0 0 0 .6-.2l1.8-3.1a.5.5 0 0 0-.1-.6l-2-1.6z"/></svg>' +
              '</button>' +
              '<button id="' + MINIMIZE_ID + '" type="button" class="ai-chat-header-icon-btn" aria-label="Minimizar chat" title="Minimizar">' +
                '<svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true"><path fill="currentColor" d="M5 12h14v2H5z"/></svg>' +
              '</button>' +
              '<button id="' + CLOSE_ID + '" type="button" class="ai-chat-close" aria-label="Cerrar asistente IA">×</button>' +
            '</div>' +
          '</div>' +
          '<div id="' + NOTICE_ID + '" class="ai-chat-notice"></div>' +
          '<div class="ai-chat-body-scroll">' +
            '<div id="' + MESSAGES_ID + '" class="ai-chat-messages"></div>' +
            '<div id="' + HINTS_ID + '" class="ai-chat-hints is-hidden">' + hintsMarkup + '</div>' +
          '</div>' +
          '<form id="' + FORM_ID + '" class="ai-chat-form">' +
            '<div class="ai-chat-toolbar-row">' +
              '<div class="ai-chat-voice-toolbar" role="toolbar" aria-label="Voz y conversación">' +
                '<button id="' + CLEAR_CHAT_ID + '" type="button" class="ai-chat-icon-btn" aria-label="Nuevo chat" title="Nuevo chat">' +
                  '<svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true"><path fill="currentColor" d="M19 11h-6V5h-2v6H5v2h6v6h2v-6h6v-2z"/></svg>' +
                '</button>' +
                '<button type="button" id="' + CONV_ID + '" class="ai-chat-icon-btn" aria-pressed="false" aria-label="Modo conversación" title="Modo conversación"></button>' +
                '<button type="button" id="' + MIC_ID + '" class="ai-chat-icon-btn" aria-pressed="false" aria-label="Dictar mensaje" title="Dictar"></button>' +
                '<button type="button" id="' + VOICE_ID + '" class="ai-chat-icon-btn" aria-pressed="false" aria-label="Voz del asistente" title="Leer respuestas"></button>' +
              '</div>' +
              '<div class="ai-chat-controls">' +
                '<label class="ai-chat-control-field" for="' + MODE_ID + '">' +
                  '<span>Modo</span>' +
                  '<select id="' + MODE_ID + '" class="form-input" aria-label="Modo del asistente IA">' +
                    '<option value="operativo">Operativo</option>' +
                    '<option value="ayudante">Ayudante por pasos</option>' +
                  '</select>' +
                '</label>' +
                '<div class="ai-chat-control-field">' +
                  '<span>Adjunto</span>' +
                  '<div class="ai-chat-attachment-row">' +
                    '<input id="' + ATTACHMENT_INPUT_ID + '" type="file" class="ai-chat-file-input" aria-label="Adjuntar archivo para IA" />' +
                    '<button id="' + ATTACH_BTN_ID + '" type="button" class="ai-chat-icon-btn" aria-label="Adjuntar archivo" title="Adjuntar">' +
                      '<svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true"><path fill="currentColor" d="M16.5 6.5v9a4.5 4.5 0 0 1-9 0v-10a3 3 0 0 1 6 0v9a1.5 1.5 0 0 1-3 0v-8h2v8a.5.5 0 0 0 1 0v-9a2 2 0 1 0-4 0v10a2.5 2.5 0 0 0 5 0v-9h2z"/></svg>' +
                    '</button>' +
                    '<button id="' + CLEAR_ATTACHMENT_ID + '" type="button" class="ai-chat-icon-btn is-hidden" aria-label="Quitar adjunto" title="Quitar adjunto">×</button>' +
                  '</div>' +
                  '<div id="' + ATTACHMENT_NAME_ID + '" class="ai-chat-attachment-name is-hidden"></div>' +
                '</div>' +
              '</div>' +
            '</div>' +
            '<textarea id="' + INPUT_ID + '" placeholder="' + placeholder + '" aria-label="Mensaje al asistente IA"></textarea>' +
            '<button type="submit" class="btn primary">Enviar</button>' +
          '</form>' +
        '</div>' +
      '</section>');
    return true;
  }

  function buildTextEndpoint() {
    if (isPublicPortalContext()) {
      return '/api/public/chat_portal';
    }
    if (isSuperContext()) {
      return '/super/api/chat_con_ia_global/consultar';
    }
    return '/api/empresa/chat_con_inteligencia_artificial/consultar';
  }

  function buildStreamEndpoint() {
    if (isPublicPortalContext()) {
      return '/api/public/chat_portal_stream';
    }
    if (isSuperContext()) {
      return '/super/api/chat_con_ia_global/consultar_stream';
    }
    return '/api/empresa/chat_con_inteligencia_artificial/consultar_stream';
  }

  function buildAttachmentEndpoint() {
    if (isPublicPortalContext()) {
      return '/api/public/chat_portal';
    }
    if (isSuperContext()) {
      return '/super/api/chat_con_ia_global/consultar_con_adjunto';
    }
    return '/api/empresa/chat_con_inteligencia_artificial/consultar_con_adjunto';
  }

  var CHAT_PERSONALITY_STORAGE_KEY = 'pcs_ai_chat_personality';
  var CHAT_ENABLED_STORAGE_KEY = 'pcs_ai_chat_enabled';
  var ROBOT_ENABLED_STORAGE_KEY = 'pcs_ai_robot_enabled';
  var VOICE_COMMAND_STORAGE_KEY = 'pcs_ai_chat_voice_enabled';
  var ROBOT_VOICE_STORAGE_KEY = 'pcs_ai_chat_robot_voice';

  function normalizeChatPersonalityMode(value) {
    var mode = normalize(value).toLowerCase();
    if ((mode === 'robot' || mode === 'secretary' || mode === 'secretaria') && state.robotEnabled) {
      return mode === 'secretaria' ? 'secretary' : mode;
    }
    return 'normal';
  }

  function normalizeRobotVoice(value) {
    var raw = normalize(value);
    var lower = raw.toLowerCase();
    if (lower === 'es-co-female' || lower === 'femenina' || lower === 'mujer') return 'es-CO-female';
    if (lower === 'es-co-male' || lower === 'masculina' || lower === 'hombre') return 'es-CO-male';
    if (lower === 'es-mx' || lower === 'mexico' || lower === 'mexicana') return 'es-MX';
    if (lower === 'es-es' || lower === 'espana' || lower === 'españa' || lower === 'castellano') return 'es-ES';
    return 'es-CO';
  }

  function labelForRobotVoice(value) {
    switch (normalizeRobotVoice(value)) {
      case 'es-CO-female': return 'Colombiana femenina';
      case 'es-CO-male': return 'Colombiana masculina';
      case 'es-MX': return 'Español latino';
      case 'es-ES': return 'Español castellano';
      default: return 'Colombiana natural';
    }
  }

  function robotVoiceLang(value) {
    var voice = normalizeRobotVoice(value);
    if (voice === 'es-MX') return 'es-MX';
    if (voice === 'es-ES') return 'es-ES';
    return 'es-CO';
  }

  function getEffectiveRobotVoice() {
    if (getChatPersonalityMode() === 'secretary') return 'es-CO-female';
    if (state.voiceOutputMode === 'computer') {
      return state.computerVoiceGender === 'male' ? 'es-CO-male' : 'es-CO-female';
    }
    return getChatPersonalityMode() === 'secretary' ? 'es-CO-female' : normalizeRobotVoice(state.robotVoice);
  }

  function readEnabledPreference(key, fallback) {
    try {
      var raw = window.localStorage.getItem(key);
      if (raw === null || raw === '') return !!fallback;
      return raw !== '0' && raw !== 'false';
    } catch (error) {
      return !!fallback;
    }
  }

  function writeEnabledPreference(key, enabled) {
    var value = !!enabled;
    try {
      window.localStorage.setItem(key, value ? '1' : '0');
    } catch (error) {}
    return value;
  }

  function setChatEnabledPreference(enabled) {
    var next = isPublicPortalContext() ? true : !!enabled;
    state.chatEnabled = writeEnabledPreference(CHAT_ENABLED_STORAGE_KEY, next);
    if (!state.chatEnabled) {
      closeChatDrawerFully();
      setRobotInlineVisible(false);
    }
    applyChatPersonalityMode();
    return state.chatEnabled;
  }

  function setRobotEnabledPreference(enabled) {
    state.robotEnabled = writeEnabledPreference(ROBOT_ENABLED_STORAGE_KEY, enabled);
    if (!state.robotEnabled) {
      try {
        window.localStorage.setItem(CHAT_PERSONALITY_STORAGE_KEY, 'normal');
      } catch (error) {}
    }
    applyChatPersonalityMode();
    return state.robotEnabled;
  }

  function setRobotVoicePreference(value) {
    state.robotVoice = normalizeRobotVoice(value);
    try {
      window.localStorage.setItem(ROBOT_VOICE_STORAGE_KEY, state.robotVoice);
    } catch (error) {}
    return state.robotVoice;
  }

  function setChatPersonalityMode(value) {
    var mode = normalizeChatPersonalityMode(value);
    try {
      window.localStorage.setItem(CHAT_PERSONALITY_STORAGE_KEY, mode);
    } catch (error) {}
    applyChatPersonalityMode();
    return mode;
  }

  function persistChatPersonalityPreference(value) {
    var mode = setChatPersonalityMode(value);
    return fetch(CHAT_PREFS_ENDPOINT, {
      method: 'PUT',
      credentials: 'same-origin',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ personality_mode: mode })
    }).catch(function (err) {
      console.warn('No se pudo guardar la apariencia del chat:', err);
      return null;
    });
  }

  function normalizeVoiceCommandText(text) {
    return String(text || '')
      .toLowerCase()
      .normalize('NFD')
      .replace(/[\u0300-\u036f]/g, '')
      .replace(/[¿?¡!.,;:]+/g, ' ')
      .replace(/\s+/g, ' ')
      .trim();
  }

  function shouldAutoEnableVoice(text) {
    var normalized = normalizeVoiceCommandText(text);
    if (!normalized) return false;
    return /\b(activa|activar|enciende|prende|habilita|pon)\b.*\b(tu\s+)?voz\b/.test(normalized) ||
      /\bvoz\b.*\b(activa|activada|encendida|prendida|habilitada)\b/.test(normalized);
  }

  function shouldAutoEnableRobot(text) {
    var normalized = normalizeVoiceCommandText(text);
    if (!normalized) return false;
    return /\b(activa|activar|enciende|prende|habilita|pon)\b.*\b(el\s+)?robot\b/.test(normalized) ||
      /\brobot\b.*\b(activo|activado|encendido|prendido|habilitado)\b/.test(normalized);
  }

  function shouldAutoEnableSecretary(text) {
    var normalized = normalizeVoiceCommandText(text);
    if (!normalized) return false;
    return /\b(activa|activar|enciende|prende|habilita|pon)\b.*\b(la\s+)?secretaria\b/.test(normalized) ||
      /\bsecretaria\b.*\b(activa|activada|encendida|prendida|habilitada)\b/.test(normalized);
  }

  function isOnlyLocalPreferenceCommand(text) {
    var normalized = normalizeVoiceCommandText(text);
    normalized = normalized
      .replace(/\b(por favor|porfa|porfis)\b/g, '')
      .replace(/\b(de la ia|del asistente|del chat|de ia|ia)\b/g, '')
      .replace(/\b(el|la|los|las|tu|mi|modo|chat)\b/g, '')
      .replace(/\b(y|e|tambien|ademas)\b/g, ' ')
      .replace(/\s+/g, ' ')
      .trim();
    return /^(activa|activar|enciende|prende|habilita|pon)(\s+(robot|secretaria|voz))+$/.test(normalized);
  }

  function isOnlyVoiceEnableCommand(text) {
    var normalized = normalizeVoiceCommandText(text);
    normalized = normalized.replace(/\b(por favor|porfa|porfis)\b/g, '').replace(/\s+/g, ' ').trim();
    return normalized === 'activa tu voz' ||
      normalized === 'activa voz' ||
      normalized === 'activar voz' ||
      normalized === 'enciende tu voz' ||
      normalized === 'enciende voz' ||
      normalized === 'prende tu voz' ||
      normalized === 'prende voz' ||
      normalized === 'habilita tu voz' ||
      normalized === 'habilita voz' ||
      normalized === 'pon tu voz' ||
      normalized === 'pon voz';
  }

  function persistVoicePreference(enabled) {
    var value = !!enabled;
    try {
      window.localStorage.setItem(VOICE_COMMAND_STORAGE_KEY, value ? '1' : '0');
    } catch (error) {}
    return fetch(CHAT_PREFS_ENDPOINT, {
      method: 'PUT',
      credentials: 'same-origin',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ voice_enabled: value })
    }).catch(function (err) {
      console.warn('No se pudo guardar la preferencia de voz del chat:', err);
      return null;
    });
  }

  function persistChatEnabledPreference(enabled) {
    var value = setChatEnabledPreference(enabled);
    return fetch(CHAT_PREFS_ENDPOINT, {
      method: 'PUT',
      credentials: 'same-origin',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ chat_enabled: value })
    }).catch(function (err) {
      console.warn('No se pudo guardar el estado del chat IA:', err);
      return null;
    });
  }

  function persistRobotEnabledPreference(enabled) {
    var value = setRobotEnabledPreference(enabled);
    return fetch(CHAT_PREFS_ENDPOINT, {
      method: 'PUT',
      credentials: 'same-origin',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ robot_enabled: value })
    }).catch(function (err) {
      console.warn('No se pudo guardar el estado del robot IA:', err);
      return null;
    });
  }

  function persistRobotVoicePreference(value) {
    var voice = setRobotVoicePreference(value);
    return fetch(CHAT_PREFS_ENDPOINT, {
      method: 'PUT',
      credentials: 'same-origin',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ robot_voice: voice })
    }).catch(function (err) {
      console.warn('No se pudo guardar la voz del robot:', err);
      return null;
    });
  }

  function applyVoicePreference(enabled) {
    state.voiceEnabled = !!enabled;
    if (!state.voiceEnabled) {
      state.conversationMode = false;
    }
    try {
      window.localStorage.setItem(VOICE_COMMAND_STORAGE_KEY, state.voiceEnabled ? '1' : '0');
    } catch (error) {}
    updateVoiceButtons(document.getElementById(MIC_ID), document.getElementById(VOICE_ID), document.getElementById(CONV_ID));
  }

  function loadVoicePreference(micBtn, voiceBtn, convBtn) {
    try {
      state.chatEnabled = isPublicPortalContext() ? true : readEnabledPreference(CHAT_ENABLED_STORAGE_KEY, true);
      state.robotEnabled = readEnabledPreference(ROBOT_ENABLED_STORAGE_KEY, true);
      state.voiceEnabled = window.localStorage.getItem(VOICE_COMMAND_STORAGE_KEY) === '1';
      setRobotVoicePreference(window.localStorage.getItem(ROBOT_VOICE_STORAGE_KEY) || state.robotVoice);
    } catch (error) {}
    applyChatPersonalityMode();
    updateVoiceButtons(micBtn, voiceBtn, convBtn);
    fetch(CHAT_PREFS_ENDPOINT, { credentials: 'same-origin' })
      .then(function (res) {
        if (!res.ok) return null;
        return res.json();
      })
      .then(function (data) {
        if (!data) return;
        if (typeof data.chat_enabled === 'boolean') {
          setChatEnabledPreference(data.chat_enabled);
        }
        if (typeof data.robot_enabled === 'boolean') {
          setRobotEnabledPreference(data.robot_enabled);
        }
        if (typeof data.voice_enabled === 'boolean') {
          applyVoicePreference(data.voice_enabled);
        }
        if (data.personality_mode) {
          setChatPersonalityMode(data.personality_mode);
        }
        if (data.robot_voice) {
          setRobotVoicePreference(data.robot_voice);
        }
        try {
          window.localStorage.setItem(VOICE_COMMAND_STORAGE_KEY, state.voiceEnabled ? '1' : '0');
        } catch (error) {}
        updateVoiceButtons(micBtn || document.getElementById(MIC_ID), voiceBtn || document.getElementById(VOICE_ID), convBtn || document.getElementById(CONV_ID));
      })
      .catch(function () {});
  }

  function activateVoiceFromCommand(micBtn, voiceBtn, convBtn) {
    state.voiceEnabled = true;
    persistVoicePreference(true);
    updateVoiceButtons(micBtn || document.getElementById(MIC_ID), voiceBtn || document.getElementById(VOICE_ID), convBtn || document.getElementById(CONV_ID));
    setNotice('Voz del asistente activada y guardada.');
  }

  function activateRobotFromCommand() {
    persistRobotEnabledPreference(true);
    persistChatPersonalityPreference('robot');
    setNotice('Robot IA activado y guardado.');
  }

  function activateSecretaryFromCommand() {
    persistRobotEnabledPreference(true);
    persistChatPersonalityPreference('secretary');
    persistRobotVoicePreference('es-CO-female');
    setNotice('Secretaria IA 3D activada con voz femenina y guardada.');
  }

  function buildPreferenceCommandMessage(wantsRobot, wantsVoice) {
    if (wantsRobot && wantsVoice) {
      return 'Listo. Active el robot IA y la voz del asistente. Guarde estas preferencias para los proximos reinicios.';
    }
    if (wantsRobot) {
      return 'Listo. Active el robot IA y guarde esta preferencia para los proximos reinicios.';
    }
    return 'Listo. Active mi voz y guarde esta preferencia para los proximos reinicios.';
  }

  function getEndpointLabel() {
    if (isPublicStoreContext()) {
      return 'chat publico de venta';
    }
    if (isPublicPortalContext()) {
      return 'chat publico del portal';
    }
    return isSuperContext() ? 'chat global de super administrador' : 'chat empresarial';
  }

  function getChatPersonalityMode() {
    if (isPublicPortalContext()) {
      return 'normal';
    }
    var raw = '';
    try {
      raw = window.localStorage.getItem(CHAT_PERSONALITY_STORAGE_KEY) || '';
    } catch (error) {
      raw = '';
    }
    raw = normalize(raw).toLowerCase();
    if ((raw === 'robot' || raw === 'clippy') && state.robotEnabled) {
      return 'robot';
    }
    if ((raw === 'secretary' || raw === 'secretaria' || raw === 'recepcionista') && state.robotEnabled) {
      return 'secretary';
    }
    return 'normal';
  }

  function isAvatarPersonalityMode(mode) {
    return mode === 'robot' || mode === 'secretary';
  }

  function getAvatarLabel(mode) {
    return mode === 'secretary' ? 'secretaria IA 3D' : 'robot IA 3D';
  }

  function getAvatarMarkup(mode) {
    return mode === 'secretary' ? SECRETARY_SVG : ROBOT_SVG;
  }

  function getRobotInlineElements() {
    return {
      host: document.getElementById(TOGGLE_ID),
      avatar: document.getElementById('robotAvatarGraphic'),
      panel: document.getElementById(ROBOT_PANEL_ID),
      status: document.getElementById(ROBOT_STATUS_ID),
      assistantBubble: document.getElementById(ROBOT_ASSISTANT_BUBBLE_ID),
      actions: document.getElementById(ROBOT_ACTIONS_ID),
      userBubble: document.getElementById(ROBOT_USER_BUBBLE_ID),
      form: document.getElementById(ROBOT_INLINE_FORM_ID),
      input: document.getElementById(ROBOT_INLINE_INPUT_ID),
      send: document.getElementById(ROBOT_INLINE_SEND_ID),
      stopVoice: document.getElementById(ROBOT_INLINE_STOP_VOICE_ID),
      mic: document.getElementById(ROBOT_INLINE_MIC_ID),
      hideBtn: document.getElementById(ROBOT_HIDE_ID),
      showBtn: document.getElementById(ROBOT_SHOW_ID)
    };
  }

  function normalizeRobotMood(mood) {
    mood = normalize(mood).toLowerCase();
    switch (mood) {
      case 'listening':
      case 'thinking':
      case 'speaking':
      case 'happy':
      case 'error':
      case 'action':
      case 'hidden':
        return mood;
      default:
        return 'idle';
    }
  }

  function getRobotStatusText(mood) {
    var modelLabel = buildResponseModelLabel(state.lastResponseModelMeta);
    var suffix = modelLabel ? (' · ' + modelLabel) : '';
    switch (normalizeRobotMood(mood)) {
      case 'listening':
        return 'Escuchando tu voz' + suffix;
      case 'thinking':
        return 'Pensando la mejor respuesta' + suffix;
      case 'speaking':
        return 'Hablando contigo' + suffix;
      case 'happy':
        return 'Lista para ayudarte' + suffix;
      case 'error':
        return 'Necesito que lo intentemos de nuevo' + suffix;
      case 'action':
        return 'Acciones listas para confirmar' + suffix;
      case 'hidden':
        return 'Asistente oculto';
      default:
        return 'Lista para ayudarte' + suffix;
    }
  }

  function setLastResponseModelMeta(meta) {
    state.lastResponseModelMeta = normalizeResponseModelMeta(meta);
    syncRobotStatus(state.loading ? 'thinking' : 'idle');
  }

  function syncRobotStatus(mood) {
    var els = getRobotInlineElements();
    if (!els.status) return;
    var next = normalizeRobotMood(mood);
    els.status.textContent = getRobotStatusText(next);
    els.status.setAttribute('data-status', next);
    ['idle', 'listening', 'thinking', 'speaking', 'happy', 'error', 'action', 'hidden'].forEach(function (name) {
      els.status.classList.remove('is-' + name);
    });
    els.status.classList.add('is-' + next);
  }

  function setRobotMood(mood, durationMs) {
    var els = getRobotInlineElements();
    var next = normalizeRobotMood(mood);
    var classNames = ['idle', 'listening', 'thinking', 'speaking', 'happy', 'error', 'action', 'hidden'];
    if (state.robotMoodTimer) {
      window.clearTimeout(state.robotMoodTimer);
      state.robotMoodTimer = null;
    }
    [els.host, els.avatar].forEach(function (node) {
      if (!node || !node.classList) return;
      classNames.forEach(function (name) {
        node.classList.remove('robot-mood-' + name);
      });
      node.classList.add('robot-mood-' + next);
      if (node.setAttribute) node.setAttribute('data-mood', next);
    });
    syncRobotStatus(next);
    if (durationMs && Number(durationMs) > 0) {
      state.robotMoodTimer = window.setTimeout(function () {
        state.robotMoodTimer = null;
        setRobotMood(state.loading ? 'thinking' : 'idle');
      }, Number(durationMs));
    }
  }

  function setRobotInlineVisible(on) {
    var els = getRobotInlineElements();
    state.robotAssistantVisible = !!on;
    if (els.host && isAvatarPersonalityMode(getChatPersonalityMode())) {
      els.host.style.display = on ? 'inline-flex' : 'none';
      els.host.setAttribute('aria-hidden', on ? 'false' : 'true');
    }
    if (els.panel) {
      els.panel.hidden = !on;
      els.panel.setAttribute('aria-hidden', on ? 'false' : 'true');
    }
    if (els.hideBtn) {
      els.hideBtn.style.display = on ? 'inline-flex' : 'none';
    }
    if (els.showBtn) {
      els.showBtn.style.display = on ? 'none' : 'inline-flex';
    }
    setRobotMood(on ? 'idle' : 'hidden');
  }

  function setRobotAssistantText(text, isError) {
    var els = getRobotInlineElements();
    if (!els.assistantBubble) return;
    var value = normalize(text) || getDefaultAssistantGreeting();
    els.assistantBubble.textContent = value;
    els.assistantBubble.classList.toggle('is-error', !!isError);
    els.assistantBubble.classList.remove('is-thinking');
    setRobotMood(isError ? 'error' : 'speaking', isError ? 2800 : 2200);
  }

  function clearRobotActionChips() {
    var els = getRobotInlineElements();
    if (!els.actions) return;
    els.actions.innerHTML = '';
    els.actions.hidden = true;
  }

  function renderRobotActionChips(actions) {
    var els = getRobotInlineElements();
    if (!els.actions) return;
    var list = Array.isArray(actions) ? actions : [];
    els.actions.innerHTML = '';
    if (!list.length) {
      els.actions.hidden = true;
      return;
    }
    list.slice(0, 8).forEach(function (item) {
      var label = normalize(item && item.label);
      var prompt = normalize(item && item.prompt);
      var url = normalize(item && item.url);
      if (!label || (!prompt && !url)) return;
      var btn = document.createElement('button');
      btn.type = 'button';
      btn.className = 'robot-assistant-action-chip';
      btn.textContent = label;
      btn.addEventListener('click', function (event) {
        event.preventDefault();
        if (url) {
          window.location.href = url;
          return;
        }
        sendRobotPrompt(prompt);
      });
      els.actions.appendChild(btn);
    });
    els.actions.hidden = !els.actions.children.length;
  }

  function renderRobotProposalActions(proposal) {
    if (!proposal || !Array.isArray(proposal.actions) || !proposal.actions.length) {
      clearRobotActionChips();
      return;
    }
    var idx = state.proposals.length;
    state.proposals.push(proposal);
    renderRobotActionChips([
      {
        label: 'Confirmar acciones',
        prompt: '__PCS_ROBOT_CONFIRM_ACTIONS__' + idx
      },
      {
        label: 'Cancelar acciones',
        prompt: '__PCS_ROBOT_CANCEL_ACTIONS__' + idx
      }
    ]);
  }

  function renderRobotDocumentExportActions(text) {
    var els = getRobotInlineElements();
    if (!els.actions) return;
    if (!shouldShowDocumentExports(text)) {
      clearRobotActionChips();
      return;
    }
    var item = {
      content: String(text || ''),
      document_type: inferDocumentExportType(text),
      source_module: inferCurrentSourceModule(),
      title: inferDocumentExportTitle(text)
    };
    els.actions.innerHTML = '';
    [
      ['pdf', 'PDF'],
      ['docx', 'Word'],
      ['xlsx', 'Excel'],
      ['txt', 'TXT'],
      ['json', 'JSON']
    ].forEach(function (entry) {
      var btn = document.createElement('button');
      btn.type = 'button';
      btn.className = 'robot-assistant-action-chip';
      btn.textContent = entry[1];
      btn.addEventListener('click', function (event) {
        event.preventDefault();
        exportChatDocumentContent(item, entry[0], btn);
      });
      els.actions.appendChild(btn);
    });
    els.actions.hidden = false;
  }

  function renderRobotGeneratedDocumentActions(doc) {
    var els = getRobotInlineElements();
    if (!els.actions || !doc) return;
    var urls = doc.download_urls || {};
    els.actions.innerHTML = '';
    [
      ['pdf', 'PDF'],
      ['docx', 'Word'],
      ['xlsx', 'Excel'],
      ['txt', 'TXT'],
      ['json', 'JSON']
    ].forEach(function (entry) {
      var url = normalize(urls[entry[0]]) || ('/download?id=' + encodeURIComponent(doc.document_id || '') + '&type=' + entry[0]);
      var btn = document.createElement('button');
      btn.type = 'button';
      btn.className = 'robot-assistant-action-chip';
      btn.textContent = 'Descargar ' + entry[1];
      btn.addEventListener('click', function (event) {
        event.preventDefault();
        if (url.indexOf('/download?') === 0) {
          window.location.href = url;
        }
      });
      els.actions.appendChild(btn);
    });
    els.actions.hidden = false;
  }

  function setRobotUserText(text) {
    var els = getRobotInlineElements();
    if (!els.userBubble) return;
    var value = normalize(text);
    els.userBubble.textContent = value;
    els.userBubble.hidden = !value;
    if (value) {
      setRobotMood('listening', 1100);
    }
  }

  function setRobotLoading(on) {
    var els = getRobotInlineElements();
    if (els.assistantBubble) {
      els.assistantBubble.classList.toggle('is-thinking', !!on);
      if (on) {
        els.assistantBubble.textContent = 'Pensando...';
        els.assistantBubble.classList.remove('is-error');
      }
    }
    if (els.input) els.input.disabled = !!on;
    if (els.send) els.send.disabled = !!on;
    setRobotMood(on ? 'thinking' : 'idle');
  }

  function focusRobotInput() {
    var input = document.getElementById(ROBOT_INLINE_INPUT_ID);
    if (!input) return;
    window.setTimeout(function () {
      input.focus();
    }, 40);
  }

  function hideRobotAssistant(toggleBtn) {
    closeChatDrawerFully();
    setRobotInlineVisible(false);
    if (toggleBtn) {
      toggleBtn.style.display = 'none';
      toggleBtn.setAttribute('aria-hidden', 'true');
    }
    var showBtn = document.getElementById(ROBOT_SHOW_ID);
    if (showBtn) showBtn.style.display = 'inline-flex';
  }

  function showRobotAssistant(toggleBtn) {
    if (!state.chatEnabled || !state.robotEnabled) return false;
    if (toggleBtn) {
      toggleBtn.style.display = 'inline-flex';
      toggleBtn.setAttribute('aria-hidden', 'false');
      toggleBtn.classList.remove('robot-appear');
      void toggleBtn.offsetWidth;
      toggleBtn.classList.add('robot-appear');
    }
    var showBtn = document.getElementById(ROBOT_SHOW_ID);
    if (showBtn) showBtn.style.display = 'none';
    setRobotInlineVisible(true);
    setRobotMood('happy', 1400);
    focusRobotInput();
    return true;
  }

  function ensureRobotInlineUI(toggleBtn) {
    if (!state.chatEnabled || !state.robotEnabled) return null;
    var panel = document.getElementById(ROBOT_PANEL_ID);
    if (!panel) {
      panel = document.createElement('section');
      panel.id = ROBOT_PANEL_ID;
      panel.className = 'robot-inline-chat-panel';
      panel.setAttribute('aria-label', 'Conversación con robot IA');
      panel.innerHTML =
        '<div id="' + ROBOT_STATUS_ID + '" class="robot-inline-status" data-status="idle">Lista para ayudarte</div>' +
        '<div id="' + ROBOT_ASSISTANT_BUBBLE_ID + '" class="robot-cloud robot-cloud-assistant"></div>' +
        '<div id="' + ROBOT_ACTIONS_ID + '" class="robot-assistant-actions" hidden></div>' +
        '<div id="' + ROBOT_USER_BUBBLE_ID + '" class="robot-cloud robot-cloud-user" hidden></div>' +
        '<form id="' + ROBOT_INLINE_FORM_ID + '" class="robot-cloud robot-cloud-input">' +
        '<textarea id="' + ROBOT_INLINE_INPUT_ID + '" rows="1" maxlength="2000"></textarea>' +
      '<button id="' + ROBOT_INLINE_STOP_VOICE_ID + '" class="robot-inline-stop-voice" type="button" aria-label="Detener voz del avatar" title="Detener voz"></button>' +
        '<button id="' + ROBOT_INLINE_MIC_ID + '" class="robot-inline-mic" type="button" aria-label="Dictar mensaje al robot"></button>' +
        '<button id="' + ROBOT_INLINE_SEND_ID + '" type="submit" aria-label="Enviar al robot">Enviar</button>' +
        '</form>';
      document.body.appendChild(panel);

      var form = document.getElementById(ROBOT_INLINE_FORM_ID);
      var input = document.getElementById(ROBOT_INLINE_INPUT_ID);
      var stopVoiceBtn = document.getElementById(ROBOT_INLINE_STOP_VOICE_ID);
      var micBtn = document.getElementById(ROBOT_INLINE_MIC_ID);
      if (form) {
        form.addEventListener('submit', handleRobotInlineSubmit);
      }
      if (input) {
        input.addEventListener('keydown', function (event) {
          if (event.key === 'Enter' && !event.shiftKey) {
            event.preventDefault();
            submitFormSafely(form, handleRobotInlineSubmit);
          }
        });
        input.addEventListener('input', function () {
          input.style.height = 'auto';
          input.style.height = Math.min(input.scrollHeight, 96) + 'px';
        });
      }
      if (stopVoiceBtn) {
        stopVoiceBtn.innerHTML = '';
        stopVoiceBtn.addEventListener('click', function (event) {
          event.preventDefault();
          event.stopPropagation();
          stopAssistantVoiceForMoment();
          focusRobotInput();
        });
      }
      var sendBtn = document.getElementById(ROBOT_INLINE_SEND_ID);
      if (sendBtn) {
        sendBtn.addEventListener('click', function (event) {
          event.preventDefault();
          event.stopPropagation();
          submitFormSafely(form, handleRobotInlineSubmit);
        });
      }
      setupSpeechRecognition(input, micBtn, document.getElementById(VOICE_ID), document.getElementById(CONV_ID));
      updateVoiceButtons(micBtn, document.getElementById(VOICE_ID), document.getElementById(CONV_ID));
    }

    var hideBtn = document.getElementById(ROBOT_HIDE_ID);
    if (!hideBtn) {
      hideBtn = document.createElement('button');
      hideBtn.id = ROBOT_HIDE_ID;
      hideBtn.type = 'button';
      hideBtn.setAttribute('aria-label', 'Ocultar robot IA');
      hideBtn.title = 'Ocultar robot IA';
      hideBtn.addEventListener('click', function (event) {
        event.stopPropagation();
        hideRobotAssistant(toggleBtn || document.getElementById(TOGGLE_ID));
      });
      document.body.appendChild(hideBtn);
    }

    var showBtn = document.getElementById(ROBOT_SHOW_ID);
    if (!showBtn) {
      showBtn = document.createElement('button');
      showBtn.id = ROBOT_SHOW_ID;
      showBtn.type = 'button';
      showBtn.setAttribute('aria-label', 'Mostrar robot IA');
      showBtn.title = 'Mostrar robot IA';
      showBtn.style.display = 'none';
      showBtn.addEventListener('click', function (event) {
        event.stopPropagation();
        showRobotAssistant(toggleBtn || document.getElementById(TOGGLE_ID));
      });
      document.body.appendChild(showBtn);
    }

    var avatarLabel = getAvatarLabel(getChatPersonalityMode());
      if (hideBtn) hideBtn.textContent = 'Ocultar ' + avatarLabel;
      if (showBtn) showBtn.textContent = 'Mostrar ' + avatarLabel;

      setRobotAssistantText(getDefaultAssistantGreeting());
      setRobotInlineVisible(state.robotAssistantVisible);
      setRobotMood(state.robotAssistantVisible ? 'happy' : 'hidden', state.robotAssistantVisible ? 1600 : 0);
  }

  function applyChatPersonalityMode() {
    var drawer = document.getElementById(DRAWER_ID);
    var input = document.getElementById(INPUT_ID);
    var titleEl = document.getElementById('aiChatTitle');
    var toggleBtn = document.getElementById(TOGGLE_ID);
    var mode = getChatPersonalityMode();

    if (!state.chatEnabled) {
      closeChatDrawerFully();
      if (drawer) drawer.classList.remove('robot-mode', 'secretary-mode');
      setRobotInlineVisible(false);
      if (toggleBtn) {
        toggleBtn.style.display = 'none';
        toggleBtn.setAttribute('aria-hidden', 'true');
      }
      var hiddenRobotBtn = document.getElementById(ROBOT_HIDE_ID);
      var hiddenShowBtn = document.getElementById(ROBOT_SHOW_ID);
      if (hiddenRobotBtn) hiddenRobotBtn.style.display = 'none';
      if (hiddenShowBtn) hiddenShowBtn.style.display = 'none';
      return;
    }

    if (toggleBtn) {
      toggleBtn.setAttribute('aria-hidden', 'false');
    }

    if (drawer) {
      drawer.classList.toggle('robot-mode', isAvatarPersonalityMode(mode));
      drawer.classList.toggle('secretary-mode', mode === 'secretary');
    }
    if (titleEl) {
      titleEl.textContent = isAvatarPersonalityMode(mode) ? (mode === 'secretary' ? 'Secretaria IA 3D' : 'Robot IA 3D') : 'Asistente IA';
    }
    if (input) {
      input.placeholder = 'Escribele al asistente IA...';
    }

    if (input && isAvatarPersonalityMode(mode)) {
      input.placeholder = mode === 'secretary' ? 'Escribele a la secretaria IA...' : 'Escribele al robot IA...';
    }
    if (!isAvatarPersonalityMode(mode)) {
      setRobotInlineVisible(false);
    }

    if (toggleBtn) {
       if (isAvatarPersonalityMode(mode)) {
          toggleBtn.classList.add('is-robot-avatar');
          toggleBtn.classList.toggle('is-secretary-avatar', mode === 'secretary');
          if (typeof toggleBtn.dataset.originalHtml === 'undefined') {
             toggleBtn.dataset.originalHtml = toggleBtn.innerHTML;
          }
           toggleBtn.innerHTML = getAvatarMarkup(mode);
           toggleBtn.setAttribute('aria-label', mode === 'secretary' ? 'Abrir secretaria IA' : 'Abrir robot IA');
           closeChatDrawerFully();
           ensureRobotInlineUI(toggleBtn);
          setRobotInlineVisible(state.robotAssistantVisible);
          return;
       }
       toggleBtn.classList.remove('is-robot-avatar', 'is-secretary-avatar');
       if (mode === 'robot') {
          if (!document.getElementById('robotAvatarGraphic')) {
             if (typeof toggleBtn.dataset.originalHtml === 'undefined') {
                toggleBtn.dataset.originalHtml = toggleBtn.innerHTML;
             }
             var robotHtml = '<div id="robotAvatarGraphic"><svg viewBox="0 0 100 100" style="width:100%; height:100%; filter: drop-shadow(0 6px 10px rgba(0,0,0,0.4));">' +
               '<!-- Sparkles -->' +
               '<g class="exec-sparkles">' +
               '  <path d="M 15 5 L 18 15 L 28 18 L 18 21 L 15 31 L 12 21 L 2 18 L 12 15 Z" fill="#facc15" />' +
               '  <path d="M 85 20 L 87 27 L 94 29 L 87 31 L 85 38 L 83 31 L 76 29 L 83 27 Z" fill="#facc15" />' +
               '  <path d="M 75 75 L 77 82 L 84 84 L 77 86 L 75 93 L 73 86 L 66 84 L 73 82 Z" fill="#facc15" />' +
               '</g>' +
               '<!-- Suit -->' +
               '<path d="M 20 100 Q 50 60 80 100 Z" fill="#1e293b" />' +
               '<!-- Shirt -->' +
               '<path d="M 40 100 L 50 75 L 60 100 Z" fill="#ffffff" />' +
               '<!-- Tie -->' +
               '<path d="M 48 75 L 52 75 L 54 95 L 50 100 L 46 95 Z" fill="#0284c7" />' +
               '<g class="exec-head-group">' +
               '  <!-- Head -->' +
               '  <circle cx="50" cy="45" r="22" fill="#fcd5ce" />' +
               '  <!-- Hair -->' +
               '  <path d="M 28 45 Q 25 15 50 20 Q 75 15 72 45 Q 75 35 68 25 Q 50 15 32 25 Q 25 35 28 45 Z" fill="#334155" />' +
               '  <!-- Eyes -->' +
               '  <circle cx="42" cy="42" r="2.5" fill="#0f172a" class="exec-eye" />' +
               '  <circle cx="58" cy="42" r="2.5" fill="#0f172a" class="exec-eye" />' +
               '  <!-- Glasses -->' +
               '  <rect x="35" y="38" width="14" height="8" rx="2" fill="none" stroke="#0f172a" stroke-width="2" />' +
               '  <rect x="51" y="38" width="14" height="8" rx="2" fill="none" stroke="#0f172a" stroke-width="2" />' +
               '  <line x1="49" y1="42" x2="51" y2="42" stroke="#0f172a" stroke-width="2" />' +
               '  <!-- Mouth -->' +
               '  <path d="M 45 54 Q 50 57 55 54" stroke="#0f172a" stroke-width="2" fill="none" class="exec-mouth" />' +
               '</g>' +
               '</svg></div>';
             toggleBtn.innerHTML = robotHtml;
             
             var hideBtn = document.createElement('button');
             hideBtn.id = 'robotHideBtn';
             hideBtn.innerHTML = 'Ocultar Ejecutivo';
             hideBtn.onclick = function(e) {
                e.stopPropagation();
                toggleBtn.style.display = 'none';
                hideBtn.style.display = 'none';
                document.getElementById('robotShowBtn').style.display = 'block';
             };
             document.body.appendChild(hideBtn);
             
             var showBtn = document.createElement('button');
             showBtn.id = 'robotShowBtn';
             showBtn.innerHTML = 'Aparecer Ejecutivo';
             showBtn.style.display = 'none';
             showBtn.onclick = function(e) {
                e.stopPropagation();
                toggleBtn.style.display = 'inline-flex';
                hideBtn.style.display = 'block';
                showBtn.style.display = 'none';
                
                // Trigger appear animation
                toggleBtn.classList.remove('exec-appear');
                void toggleBtn.offsetWidth; // trigger reflow
                toggleBtn.classList.add('exec-appear');
             };
             document.body.appendChild(showBtn);
          }
          toggleBtn.style.display = 'inline-flex';
          var hb = document.getElementById('robotHideBtn');
          if (hb) hb.style.display = 'block';
          var sb = document.getElementById('robotShowBtn');
          if (sb) sb.style.display = 'none';
       } else {
          if (typeof toggleBtn.dataset.originalHtml !== 'undefined') {
             toggleBtn.innerHTML = toggleBtn.dataset.originalHtml;
          }
          var hb = document.getElementById('robotHideBtn');
          if (hb) hb.style.display = 'none';
          var sb = document.getElementById('robotShowBtn');
          if (sb) sb.style.display = 'none';
          toggleBtn.style.display = 'inline-flex';
       }
    }
  }

  function getDefaultAssistantGreeting() {
    var mode = getChatPersonalityMode();
    if (isPublicStoreContext()) {
      return 'Hola. Soy el asistente publico de esta tienda. Puedo ayudarte con productos, servicios, precios y paginas publicas de esta empresa.';
    }
    if (isPublicPortalContext()) {
      return 'Hola. Soy el asistente publico de Powerful Control System. Puedo ayudarte con planes, modulos, precios, contacto y como empezar.';
    }
    if (mode === 'secretary') {
      return 'Hola. Soy tu secretaria IA 3D, lista para ayudarte a organizar tareas, ventas y configuraciones.';
    }
    if (mode === 'robot') {
      return 'Hola. Soy tu robot IA 3D, listo para ayudarte en este panel.';
    }
    return 'Hola. Soy tu Asistente IA, listo para ayudarte en el panel.';
  }

  function getCompactConfigMode() {
    var selected = document.querySelector('input[name="aiChatCompactMode"]:checked');
    return normalizeChatPersonalityMode(selected && selected.value);
  }

  function setCompactConfigState(mode, voiceEnabled, robotVoice, chatEnabled, robotEnabled) {
    if (typeof chatEnabled === 'boolean') {
      state.chatEnabled = chatEnabled;
    }
    if (typeof robotEnabled === 'boolean') {
      state.robotEnabled = robotEnabled;
    }
    var normalizedMode = normalizeChatPersonalityMode(mode);
    var chatInput = document.getElementById(CONFIG_CHAT_ENABLED_ID);
    var robotInput = document.getElementById(CONFIG_ROBOT_ENABLED_ID);
    var modeInput = document.querySelector('input[name="aiChatCompactMode"][value="' + normalizedMode + '"]');
    var voiceInput = document.getElementById(CONFIG_VOICE_ID);
    var robotVoiceInput = document.getElementById(CONFIG_ROBOT_VOICE_ID);
    var modeInputs = Array.prototype.slice.call(document.querySelectorAll('input[name="aiChatCompactMode"]'));
    if (chatInput) {
      chatInput.checked = !!state.chatEnabled;
    }
    if (robotInput) {
      robotInput.checked = !!state.robotEnabled;
      robotInput.disabled = !state.chatEnabled;
    }
    modeInputs.forEach(function (input) {
      input.disabled = !state.chatEnabled || ((input.value === 'robot' || input.value === 'secretary') && !state.robotEnabled);
    });
    if (modeInput) {
      modeInput.checked = true;
    }
    if (voiceInput && typeof voiceEnabled === 'boolean') {
      voiceInput.checked = voiceEnabled;
      voiceInput.disabled = !state.chatEnabled;
    }
    if (robotVoiceInput) {
      robotVoiceInput.value = normalizedMode === 'secretary' ? 'es-CO-female' : normalizeRobotVoice(robotVoice || state.robotVoice);
      robotVoiceInput.disabled = !state.chatEnabled || !state.robotEnabled;
    }
  }

  function ensureCompactConfigPanel() {
    var panel = document.getElementById(CONFIG_PANEL_ID);
    if (panel) return panel;

    panel = document.createElement('div');
    panel.id = CONFIG_PANEL_ID;
    panel.className = 'ai-chat-compact-config';
    panel.hidden = true;
    panel.innerHTML =
      '<div class="ai-chat-compact-config-card" role="dialog" aria-modal="false" aria-labelledby="aiChatCompactConfigTitle">' +
      '<div class="ai-chat-compact-config-header">' +
      '<strong id="aiChatCompactConfigTitle">Configuración del chat</strong>' +
      '<button id="' + CONFIG_CLOSE_ID + '" type="button" class="ai-chat-header-icon-btn" aria-label="Cerrar configuración">×</button>' +
      '</div>' +
      '<div class="ai-chat-compact-config-body">' +
      '<label class="ai-chat-compact-option"><input id="' + CONFIG_CHAT_ENABLED_ID + '" type="checkbox"><span><b>Activar chat IA</b><small>Muestra u oculta el chat flotante completo.</small></span></label>' +
      '<label class="ai-chat-compact-option"><input id="' + CONFIG_ROBOT_ENABLED_ID + '" type="checkbox"><span><b>Activar robot IA</b><small>Permite el avatar 3D, la guia inicial y avisos de recordatorios.</small></span></label>' +
      '<label class="ai-chat-compact-option"><input type="radio" name="aiChatCompactMode" value="normal"><span><b>Chat cuadrado</b><small>Ventana lateral tradicional con historial y controles completos.</small></span></label>' +
      '<label class="ai-chat-compact-option"><input type="radio" name="aiChatCompactMode" value="robot"><span><b>Robot IA</b><small>Avatar 3D con conversación en globos sobre el robot.</small></span></label>' +
      '<label class="ai-chat-compact-option"><input type="radio" name="aiChatCompactMode" value="secretary"><span><b>Secretaria IA 3D</b><small>Avatar estilo caricatura ejecutiva joven con voz femenina.</small></span></label>' +
      '<label class="ai-chat-compact-option ai-chat-compact-option-voice"><input id="' + CONFIG_VOICE_ID + '" type="checkbox"><span><b>Activar modo voz</b><small>Lee las respuestas con el servicio de voz o la voz del navegador.</small></span></label>' +
      '<label class="ai-chat-compact-option"><span><b>Voz del avatar</b><small>La secretaria usa automáticamente voz femenina.</small><select id="' + CONFIG_ROBOT_VOICE_ID + '" class="form-input"><option value="es-CO">Colombiana natural</option><option value="es-CO-female">Colombiana femenina</option><option value="es-CO-male">Colombiana masculina</option><option value="es-MX">Español latino</option><option value="es-ES">Español castellano</option></select></span></label>' +
      '</div>' +
      '<div class="ai-chat-compact-config-actions">' +
      '<button id="' + CONFIG_SAVE_ID + '" type="button" class="btn primary small">Guardar</button>' +
      '</div>' +
      '<div class="ai-chat-compact-config-status" aria-live="polite"></div>' +
      '</div>';
    document.body.appendChild(panel);

    var closeBtn = document.getElementById(CONFIG_CLOSE_ID);
    var saveBtn = document.getElementById(CONFIG_SAVE_ID);
    var statusEl = panel.querySelector('.ai-chat-compact-config-status');

    function setConfigStatus(message, isError) {
      if (!statusEl) return;
      statusEl.textContent = String(message || '');
      statusEl.classList.toggle('is-error', !!isError);
    }

    function applyCompactPreview() {
      var chatInput = document.getElementById(CONFIG_CHAT_ENABLED_ID);
      var robotInput = document.getElementById(CONFIG_ROBOT_ENABLED_ID);
      var chatOn = setChatEnabledPreference(chatInput ? chatInput.checked : state.chatEnabled);
      var robotOn = setRobotEnabledPreference(robotInput ? robotInput.checked : state.robotEnabled);
      var mode = setChatPersonalityMode(getCompactConfigMode());
      applyVoicePreference(!!document.getElementById(CONFIG_VOICE_ID).checked);
      setRobotVoicePreference(mode === 'secretary' ? 'es-CO-female' : document.getElementById(CONFIG_ROBOT_VOICE_ID).value);
      setCompactConfigState(getChatPersonalityMode(), state.voiceEnabled, state.robotVoice, chatOn, robotOn);
      setConfigStatus('Vista previa aplicada. Chat: ' + (chatOn ? 'activo' : 'desactivado') + '. Robot: ' + (robotOn ? 'activo' : 'desactivado') + '. Voz del avatar: ' + labelForRobotVoice(getEffectiveRobotVoice()) + '. Presiona Guardar para persistirla.');
    }

    var chatInput = document.getElementById(CONFIG_CHAT_ENABLED_ID);
    if (chatInput) {
      chatInput.addEventListener('change', applyCompactPreview);
    }
    var robotInput = document.getElementById(CONFIG_ROBOT_ENABLED_ID);
    if (robotInput) {
      robotInput.addEventListener('change', applyCompactPreview);
    }
    Array.prototype.slice.call(panel.querySelectorAll('input[name="aiChatCompactMode"]')).forEach(function (input) {
      input.addEventListener('change', applyCompactPreview);
    });
    var voiceInput = document.getElementById(CONFIG_VOICE_ID);
    if (voiceInput) {
      voiceInput.addEventListener('change', applyCompactPreview);
    }
    var robotVoiceInput = document.getElementById(CONFIG_ROBOT_VOICE_ID);
    if (robotVoiceInput) {
      robotVoiceInput.addEventListener('change', applyCompactPreview);
    }
    if (closeBtn) {
      closeBtn.addEventListener('click', function () {
        panel.hidden = true;
      });
    }
    panel.addEventListener('click', function (event) {
      if (event.target === panel) {
        panel.hidden = true;
      }
    });
    if (saveBtn) {
      saveBtn.addEventListener('click', function () {
        var chatOn = setChatEnabledPreference(!!document.getElementById(CONFIG_CHAT_ENABLED_ID).checked);
        var robotOn = setRobotEnabledPreference(!!document.getElementById(CONFIG_ROBOT_ENABLED_ID).checked);
        var mode = setChatPersonalityMode(getCompactConfigMode());
        var voice = !!document.getElementById(CONFIG_VOICE_ID).checked;
        var robotVoice = setRobotVoicePreference(mode === 'secretary' ? 'es-CO-female' : document.getElementById(CONFIG_ROBOT_VOICE_ID).value);
        applyVoicePreference(voice);
        setConfigStatus('Guardando configuración...');
        fetch(CHAT_PREFS_ENDPOINT, {
          method: 'PUT',
          credentials: 'same-origin',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ chat_enabled: chatOn, robot_enabled: robotOn, personality_mode: mode, voice_enabled: voice, robot_voice: robotVoice })
        }).then(function (res) {
          if (!res.ok) throw new Error('No se pudo guardar en servidor.');
          return res.json();
        }).then(function (data) {
          var savedChat = typeof (data && data.chat_enabled) === 'boolean' ? setChatEnabledPreference(data.chat_enabled) : chatOn;
          var savedRobot = typeof (data && data.robot_enabled) === 'boolean' ? setRobotEnabledPreference(data.robot_enabled) : robotOn;
          var savedMode = setChatPersonalityMode(data && data.personality_mode ? data.personality_mode : mode);
          var savedVoice = typeof (data && data.voice_enabled) === 'boolean' ? data.voice_enabled : voice;
          var savedRobotVoice = setRobotVoicePreference(data && data.robot_voice ? data.robot_voice : robotVoice);
          applyVoicePreference(savedVoice);
          setCompactConfigState(savedMode, savedVoice, savedRobotVoice, savedChat, savedRobot);
          setConfigStatus('Configuración guardada. Chat: ' + (savedChat ? 'activo' : 'desactivado') + '. Robot: ' + (savedRobot ? 'activo' : 'desactivado') + '. Voz del avatar: ' + labelForRobotVoice(savedMode === 'secretary' ? 'es-CO-female' : savedRobotVoice) + '.');
        }).catch(function (err) {
          setConfigStatus('Configuración aplicada localmente, pero no se pudo guardar. ' + String(err && err.message ? err.message : ''), true);
        });
      });
    }
    return panel;
  }

  function openChatConfigPage() {
    if (isPublicPortalContext()) {
      setNotice(isPublicStoreContext()
        ? 'Este chat publico ya viene restringido al catalogo de esta empresa y no permite configuracion administrativa desde aqui.'
        : 'Este chat publico ya viene restringido al portal y no permite configuracion administrativa desde aqui.');
      return;
    }
    var panel = ensureCompactConfigPanel();
    setCompactConfigState(getChatPersonalityMode(), state.voiceEnabled, state.robotVoice, state.chatEnabled, state.robotEnabled);
    panel.hidden = false;
    fetch(CHAT_PREFS_ENDPOINT, { credentials: 'same-origin' })
      .then(function (res) {
        if (!res.ok) return null;
        return res.json();
      })
      .then(function (data) {
        if (!data) return;
        if (typeof data.chat_enabled === 'boolean') {
          setChatEnabledPreference(data.chat_enabled);
        }
        if (typeof data.robot_enabled === 'boolean') {
          setRobotEnabledPreference(data.robot_enabled);
        }
        if (data.personality_mode) {
          setChatPersonalityMode(data.personality_mode);
        }
        if (typeof data.voice_enabled === 'boolean') {
          applyVoicePreference(data.voice_enabled);
        }
        if (data.robot_voice) {
          setRobotVoicePreference(data.robot_voice);
        }
        setCompactConfigState(getChatPersonalityMode(), state.voiceEnabled, state.robotVoice, state.chatEnabled, state.robotEnabled);
      })
      .catch(function () {});
  }

  function normalize(text) {
    return String(text || '').trim();
  }

  function getAssistantMode() {
    var modeEl = document.getElementById(MODE_ID);
    var value = normalize(modeEl && modeEl.value);
    if (isPublicPortalContext()) {
      return value === 'ayudante' ? 'ayudante' : 'operativo';
    }
    if (value === 'reportes') return 'reportes';
    if (value === 'documentos') return 'documentos';
    return value === 'ayudante' ? 'ayudante' : 'operativo';
  }

  function isReportMode() {
    return getAssistantMode() === 'reportes';
  }

  function isDocumentMode() {
    return getAssistantMode() === 'documentos';
  }

  function shouldAutoUseDocumentMode(query) {
    if (isSuperContext() || isPublicPortalContext()) return false;
    var text = normalize(String(query || ''));
    if (!text) return false;
    var hasAction = /\b(genera|generar|crea|crear|haz|hacer|redacta|redactar|prepara|preparar|exporta|exportar)\b/.test(text);
    var hasDocument = /\b(documento|contrato|factura|reporte|informe|acta|cotizacion|presupuesto|excel|xlsx|word|docx|pdf|tabla|listado)\b/.test(text);
    return hasAction && hasDocument;
  }

  function isImageFileForAI(file) {
    if (!file) return false;
    var type = String(file.type || '').toLowerCase();
    if (type.indexOf('image/') === 0) return true;
    return /\.(jpe?g|png|webp|gif|bmp|heic|heif)$/i.test(String(file.name || ''));
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

  function ensureDocumentModeUI() {
    if (isPublicPortalContext()) {
      return null;
    }
    var modeEl = document.getElementById(MODE_ID);
    if (modeEl && !modeEl.querySelector('option[value="documentos"]')) {
      var option = document.createElement('option');
      option.value = 'documentos';
      option.textContent = 'Documentos IA';
      var reportOption = modeEl.querySelector('option[value="reportes"]');
      if (reportOption && reportOption.parentNode === modeEl) {
        modeEl.insertBefore(option, reportOption);
      } else {
        modeEl.appendChild(option);
      }
    }

    var tools = document.getElementById(DOCUMENT_TOOLS_ID);
    if (tools) return tools;
    var controls = modeEl && modeEl.closest('.ai-chat-controls');
    if (!controls) return null;

    tools = document.createElement('div');
      tools.id = DOCUMENT_TOOLS_ID;
      tools.className = 'ai-chat-control-field ai-chat-document-tools is-hidden';
      tools.innerHTML =
        '<span>Archivo IA</span>' +
        '<div class="ai-chat-document-tool-row">' +
        '<select id="' + DOCUMENT_FORMAT_ID + '" class="form-input" aria-label="Formato de descarga del documento">' +
      '<option value="pdf">PDF</option>' +
      '<option value="docx">Word</option>' +
      '<option value="xlsx">Excel</option>' +
        '<option value="txt">TXT</option>' +
        '<option value="json">JSON</option>' +
        '</select>' +
        '<button id="' + DOCUMENT_DOWNLOAD_ID + '" type="button" class="btn secondary small ai-chat-document-download" disabled>Descargar</button>' +
        '<button id="' + DOCUMENT_EMAIL_ID + '" type="button" class="btn secondary small ai-chat-document-download" disabled>Correo</button>' +
        '<button id="' + DOCUMENT_WHATSAPP_ID + '" type="button" class="btn secondary small ai-chat-document-download" disabled>WhatsApp</button>' +
        '</div>';
    controls.appendChild(tools);

      var downloadBtn = document.getElementById(DOCUMENT_DOWNLOAD_ID);
      if (downloadBtn) {
        downloadBtn.addEventListener('click', function (event) {
          event.preventDefault();
          downloadCurrentGeneratedDocument();
        });
      }
      var emailBtn = document.getElementById(DOCUMENT_EMAIL_ID);
      if (emailBtn) {
        emailBtn.addEventListener('click', function (event) {
          event.preventDefault();
          shareCurrentArtifactByEmail();
        });
      }
      var whatsAppBtn = document.getElementById(DOCUMENT_WHATSAPP_ID);
      if (whatsAppBtn) {
        whatsAppBtn.addEventListener('click', function (event) {
          event.preventDefault();
          shareCurrentArtifactByWhatsApp();
        });
      }
      return tools;
    }

  function setGeneratedDocument(doc) {
    state.generatedDocument = doc || null;
    if (doc) {
      var selectedFormat = getSelectedDocumentFormat();
      var urls = doc.download_urls || {};
      var url = normalize(urls[selectedFormat]) || normalize(urls.pdf) || normalize(urls.docx) || normalize(urls.xlsx) || normalize(urls.txt) || normalize(urls.json);
      setShareArtifact({
        kind: 'document',
        title: normalize(doc.title || 'Documento IA'),
        format: selectedFormat,
        url: url,
        summary: normalize(doc.preview_text || '')
      });
    } else {
      setShareArtifact(null);
    }
    updateDocumentDownloadButton();
  }

  function getSelectedDocumentFormat() {
    var select = document.getElementById(DOCUMENT_FORMAT_ID);
    var value = normalize(select && select.value).toLowerCase();
    return /^(pdf|docx|xlsx|txt|json)$/.test(value) ? value : 'pdf';
  }

  function updateDocumentDownloadButton() {
    var btn = document.getElementById(DOCUMENT_DOWNLOAD_ID);
    var emailBtn = document.getElementById(DOCUMENT_EMAIL_ID);
    var whatsAppBtn = document.getElementById(DOCUMENT_WHATSAPP_ID);
    var hasArtifact = !!(state.generatedDocument || state.shareArtifact);
    if (btn) {
      btn.disabled = !state.generatedDocument && !state.shareArtifact;
      btn.textContent = hasArtifact ? 'Descargar' : 'Genera primero';
    }
    if (emailBtn) emailBtn.disabled = !hasArtifact;
    if (whatsAppBtn) whatsAppBtn.disabled = !hasArtifact;
  }

  function toAbsoluteURL(url) {
    var raw = normalize(url);
    if (!raw) return '';
    if (/^https?:\/\//i.test(raw)) return raw;
    try {
      return new URL(raw, window.location.origin).toString();
    } catch (error) {
      return raw;
    }
  }

  function setShareArtifact(artifact) {
    state.shareArtifact = artifact || null;
    if (artifact && artifact.format) {
      var select = document.getElementById(DOCUMENT_FORMAT_ID);
      if (select) {
        var normalized = normalize(String(artifact.format)).toLowerCase();
        if (select.querySelector('option[value="' + normalized + '"]')) {
          select.value = normalized;
        }
      }
    }
    updateDocumentDownloadButton();
  }

  function getCurrentShareArtifact() {
    if (state.generatedDocument) {
      var format = getSelectedDocumentFormat();
      var urls = state.generatedDocument.download_urls || {};
      var picked = normalize(urls[format]) || normalize(urls.pdf) || normalize(urls.docx) || normalize(urls.xlsx) || normalize(urls.txt) || normalize(urls.json);
      return {
        kind: 'document',
        title: normalize(state.generatedDocument.title || 'Documento IA'),
        format: format,
        url: picked,
        summary: normalize(state.generatedDocument.preview_text || '')
      };
    }
    return state.shareArtifact || null;
  }

  function buildArtifactShareMessage(artifact) {
    if (!artifact) return '';
    var title = normalize(artifact.title || 'Archivo generado desde chat IA');
    var url = toAbsoluteURL(artifact.url);
    var parts = [title];
    if (artifact.format) {
      parts.push('Formato: ' + String(artifact.format).toUpperCase());
    }
    if (artifact.summary) {
      parts.push(normalize(String(artifact.summary)).slice(0, 280));
    }
    if (url) {
      parts.push('Enlace: ' + url);
    }
    return parts.join('\n');
  }

  function parseArtifactExportParams(artifact) {
    var rawUrl = normalize(artifact && artifact.url);
    if (!rawUrl) return null;
    try {
      var parsed = new URL(rawUrl, window.location.origin);
      return parsed;
    } catch (error) {
      return null;
    }
  }

  function shareCurrentArtifactByWhatsApp() {
    var artifact = getCurrentShareArtifact();
    if (!artifact) {
      setNotice('Primero genera un documento o reporte para compartirlo.', true);
      return;
    }
    var text = buildArtifactShareMessage(artifact);
    if (!text) {
      setNotice('No se pudo preparar el contenido para WhatsApp.', true);
      return;
    }
    window.open('https://wa.me/?text=' + encodeURIComponent(text), '_blank', 'noopener,noreferrer');
    setNotice('Se abrio WhatsApp para compartir el archivo generado.');
  }

  function shareCurrentArtifactByEmail() {
    var artifact = getCurrentShareArtifact();
    if (!artifact) {
      setNotice('Primero genera un documento o reporte para compartirlo.', true);
      return;
    }
    var target = window.prompt('Correo destino:', '');
    if (target === null) return;
    var email = normalize(target);
    if (!email) {
      setNotice('No se ingreso un correo de destino.', true);
      return;
    }
    var subject = artifact.title || 'Archivo generado desde chat IA';
    var body = buildArtifactShareMessage(artifact);
    var empresaId = getCurrentEmpresaId();
    if (!empresaId) {
      window.location.href = 'mailto:' + encodeURIComponent(email) + '?subject=' + encodeURIComponent(subject) + '&body=' + encodeURIComponent(body);
      setNotice('Se preparo el correo localmente porque no hay empresa activa.');
      return;
    }
    if (artifact.kind === 'report') {
      var exportParams = parseArtifactExportParams(artifact);
      var dataset = exportParams ? normalize(exportParams.searchParams.get('dataset')) : '';
      var format = exportParams ? normalize(exportParams.searchParams.get('format')) : normalize(artifact.format);
      if (!dataset) {
        window.location.href = 'mailto:' + encodeURIComponent(email) + '?subject=' + encodeURIComponent(subject) + '&body=' + encodeURIComponent(body);
        setNotice('No se pudo resolver el dataset del reporte. Se preparo correo local como respaldo.', true);
        return;
      }
      fetch('/api/empresa/reportes?action=enviar_email&empresa_id=' + encodeURIComponent(String(parsePositiveInt(empresaId))), {
        method: 'POST',
        credentials: 'same-origin',
        headers: {
          'Content-Type': 'application/json',
          'X-PCS-Source': 'ai_drawer'
        },
        body: JSON.stringify({
          to_email: email,
          subject: subject,
          message: body,
          dataset: dataset,
          format: format || 'pdf'
        })
      }).then(function (resp) {
        if (!resp.ok) return parseErrorResponse(resp);
        return resp.json();
      }).then(function () {
        setNotice('Reporte enviado por correo desde el servidor.');
      }).catch(function (err) {
        setNotice('No se pudo enviar el reporte por correo: ' + String(err && err.message ? err.message : err), true);
      });
      return;
    }

    var documentId = normalize(state.generatedDocument && (state.generatedDocument.document_id || state.generatedDocument.id));
    if (!documentId) {
      var fallbackMail = 'mailto:' + encodeURIComponent(email) + '?subject=' + encodeURIComponent(subject) + '&body=' + encodeURIComponent(body);
      window.location.href = fallbackMail;
      setNotice('Se preparo el correo localmente porque no se encontro el documento activo.');
      return;
    }
    fetch('/api/empresa/chat_documentos/compartir_email', {
      method: 'POST',
      credentials: 'same-origin',
      headers: {
        'Content-Type': 'application/json',
        'X-PCS-Source': 'ai_drawer'
      },
      body: JSON.stringify({
        empresa_id: parsePositiveInt(empresaId),
        document_id: documentId,
        format: normalize(artifact.format) || getSelectedDocumentFormat(),
        to_email: email,
        subject: subject,
        message: body
      })
    }).then(function (resp) {
      if (!resp.ok) return parseErrorResponse(resp);
      return resp.json();
    }).then(function () {
      setNotice('Documento enviado por correo desde el servidor.');
    }).catch(function (err) {
      setNotice('No se pudo enviar el documento por correo: ' + String(err && err.message ? err.message : err), true);
    });
  }

  function downloadCurrentGeneratedDocument() {
    var artifact = getCurrentShareArtifact();
    if (!artifact) {
      setNotice('Primero genera un documento o reporte para descargarlo.', true);
      return;
    }
    var url = normalize(artifact.url);
    if (!url || url.indexOf('/download?') !== 0) {
      if (/^https?:\/\//i.test(url)) {
        window.location.href = url;
        return;
      }
      setNotice('No se pudo resolver la descarga del archivo.', true);
      return;
    }
    window.location.href = url;
  }

  function isSpeechRecognitionSupported() {
    return !!(window.SpeechRecognition || window.webkitSpeechRecognition);
  }

  function isSpeechSynthesisSupported() {
    return !!(window.speechSynthesis && typeof window.SpeechSynthesisUtterance === 'function');
  }

  function isVoiceOutputSupported() {
    return isSpeechSynthesisSupported() || state.voiceServerAvailable;
  }

  function sanitizeTextForSpeech(text) {
    var raw = String(text || '');
    if (!raw) return '';
    return raw
      .replace(/```[\s\S]*?```/g, ' bloque de codigo omitido. ')
      .replace(/`([^`]+)`/g, '$1')
      .replace(/!\[([^\]]*)\]\([^)]+\)/g, '$1')
      .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '$1')
      .replace(/https?:\/\/\S+/gi, ' enlace ')
      .replace(/^\s{0,3}#{1,6}\s+/gm, '')
      .replace(/^\s{0,3}>\s?/gm, '')
      .replace(/^\s*[-*+]\s+/gm, '')
      .replace(/^\s*\d+[.)]\s+/gm, '')
      .replace(/\*\*([^*]+)\*\*/g, '$1')
      .replace(/__([^_]+)__/g, '$1')
      .replace(/\*([^*]+)\*/g, '$1')
      .replace(/_([^_]+)_/g, '$1')
      .replace(/~~([^~]+)~~/g, '$1')
      .replace(/[*_~`#>|{}\[\]\\]/g, ' ')
      .replace(/[\u2022\u00b7]/g, ' ')
      .replace(/\s+([,.;:!?])/g, '$1')
      .replace(/([,.;:!?]){2,}/g, '$1')
      .replace(/\s+/g, ' ')
      .trim();
  }

  function stopAssistantVoiceForMoment() {
    clearConversationResumeTimer();
    resetQueuedAssistantSpeech();
    state.voicePlaybackVersion += 1;
    try {
      if (state.voiceServerAudio) {
        state.voiceServerAudio.pause();
        state.voiceServerAudio.currentTime = 0;
        state.voiceServerAudio = null;
      }
    } catch (err) {}
    try {
      if (window.speechSynthesis && typeof window.speechSynthesis.cancel === 'function') {
        window.speechSynthesis.cancel();
      }
    } catch (err) {}
    if (isAvatarPersonalityMode(getChatPersonalityMode())) {
      setRobotMood('idle', 900);
    }
    setNotice('Voz detenida por ahora. La siguiente respuesta volvera a hablar si el modo voz sigue activo.');
  }

  function clearConversationResumeTimer() {
    if (state.conversationResumeTimer) {
      window.clearTimeout(state.conversationResumeTimer);
      state.conversationResumeTimer = null;
    }
  }

  function getPreferredConversationMicButton() {
    var preferredId = state.preferredConversationMicId || MIC_ID;
    var preferred = document.getElementById(preferredId);
    if (preferred) return preferred;
    return document.getElementById(MIC_ID) || document.getElementById(ROBOT_INLINE_MIC_ID) || null;
  }

  function attemptConversationResume(playbackVersion, triesLeft) {
    clearConversationResumeTimer();
    if (!state.conversationMode) return;
    if (playbackVersion !== undefined && state.voicePlaybackVersion !== playbackVersion) return;
    if (state.loading || state.listening) {
      if (triesLeft > 0) {
        state.conversationResumeTimer = window.setTimeout(function () {
          attemptConversationResume(playbackVersion, triesLeft - 1);
        }, 650);
      }
      return;
    }
    var micBtn = getPreferredConversationMicButton();
    if (!micBtn || micBtn.disabled) return;
    micBtn.click();
  }

  function scheduleConversationMicResume(playbackVersion) {
    clearConversationResumeTimer();
    if (!state.conversationMode) return;
    var version = playbackVersion !== undefined ? playbackVersion : state.voicePlaybackVersion;
    Promise.resolve(state.voiceQueuePromise || Promise.resolve()).finally(function () {
      if (!state.conversationMode) return;
      if (state.voicePlaybackVersion !== version) return;
      state.conversationResumeTimer = window.setTimeout(function () {
        attemptConversationResume(version, 5);
      }, 520);
    });
  }

  function updateVoiceButtons(micBtn, voiceBtn, convBtn) {
    if (micBtn) {
      micBtn.innerHTML = ICON_MIC;
      micBtn.title = state.listening ? 'Detener dictado' : 'Dictar con el micrófono';
      micBtn.setAttribute('aria-label', state.listening ? 'Detener dictado' : 'Dictar mensaje');
      micBtn.setAttribute('aria-pressed', state.listening ? 'true' : 'false');
      micBtn.disabled = state.loading || !isSpeechRecognitionSupported();
      micBtn.classList.toggle('is-listening', state.listening);
      if (!isSpeechRecognitionSupported()) {
        micBtn.title = 'Dictado no disponible en este navegador';
      }
    }
    var robotMicBtn = document.getElementById(ROBOT_INLINE_MIC_ID);
    if (robotMicBtn && robotMicBtn !== micBtn) {
      robotMicBtn.innerHTML = ICON_MIC;
      robotMicBtn.title = state.listening ? 'Detener dictado' : 'Dictar con el micrófono';
      robotMicBtn.setAttribute('aria-label', state.listening ? 'Detener dictado' : 'Dictar mensaje al robot');
      robotMicBtn.setAttribute('aria-pressed', state.listening ? 'true' : 'false');
      robotMicBtn.disabled = state.loading || !isSpeechRecognitionSupported();
      robotMicBtn.classList.toggle('is-listening', state.listening);
      if (!isSpeechRecognitionSupported()) {
        robotMicBtn.title = 'Dictado no disponible en este navegador';
      }
    }
    var robotStopVoiceBtn = document.getElementById(ROBOT_INLINE_STOP_VOICE_ID);
    if (robotStopVoiceBtn) {
      robotStopVoiceBtn.innerHTML = ICON_STOP;
      robotStopVoiceBtn.title = 'Detener voz del avatar por ahora';
      robotStopVoiceBtn.setAttribute('aria-label', 'Detener voz del avatar por ahora');
      robotStopVoiceBtn.disabled = !isVoiceOutputSupported();
    }
    if (voiceBtn) {
      voiceBtn.innerHTML = ICON_SPK;
      voiceBtn.title = state.voiceEnabled ? 'Desactivar voz del asistente' : 'Activar voz del asistente';
      voiceBtn.setAttribute('aria-label', state.voiceEnabled ? 'Voz del asistente activada' : 'Activar voz del asistente');
      voiceBtn.setAttribute('aria-pressed', state.voiceEnabled ? 'true' : 'false');
      voiceBtn.disabled = state.loading || !isVoiceOutputSupported();
      if (!isVoiceOutputSupported()) {
        voiceBtn.title = 'Texto a voz no disponible';
      }
    }
    if (convBtn) {
      convBtn.innerHTML = ICON_CONV;
      convBtn.title = state.conversationMode ? 'Modo conversación activo' : 'Modo conversación (dictado y voz del asistente)';
      convBtn.setAttribute('aria-label', 'Modo conversación');
      convBtn.setAttribute('aria-pressed', state.conversationMode ? 'true' : 'false');
    }
  }

  function stopActiveSpeechRecognition(silent) {
    clearConversationResumeTimer();
    var active = state.activeSpeechRecognition;
    if (active) {
      try {
        if (silent && typeof active.abort === 'function') {
          active.abort();
        } else {
          active.stop();
        }
      } catch (err) {
        try {
          if (typeof active.abort === 'function') active.abort();
        } catch (abortErr) {}
      }
    }
    state.activeSpeechRecognition = null;
    state.activeSpeechSource = '';
    state.listening = false;
    updateVoiceButtons(document.getElementById(MIC_ID), document.getElementById(VOICE_ID), document.getElementById(CONV_ID));
    if (isAvatarPersonalityMode(getChatPersonalityMode())) {
      setRobotMood('idle', 700);
    }
    if (!silent) {
      setNotice('Dictado detenido.');
    }
  }

  function submitFormSafely(form, fallbackSubmit) {
    if (state.listening) {
      stopActiveSpeechRecognition(true);
    }
    if (form && typeof form.requestSubmit === 'function') {
      form.requestSubmit();
      return;
    }
    if (typeof fallbackSubmit === 'function') {
      fallbackSubmit({ preventDefault: function () {} });
    }
  }

  function normalizeSpeechCommandText(text) {
    return String(text || '')
      .normalize('NFD')
      .replace(/[\u0300-\u036f]/g, '')
      .toLowerCase()
      .replace(/[^\p{L}\p{N}\s]/gu, ' ')
      .replace(/\s+/g, ' ')
      .trim();
  }

  function splitTextForFastSpeech(text) {
    var clean = String(text || '').replace(/\s+/g, ' ').trim();
    if (!clean) return [];
    var maxChunkLength = 200;
    var chunks = [];

    function pushChunk(value) {
      var chunk = String(value || '').replace(/\s+/g, ' ').trim();
      if (!chunk) return;
      if (chunk.length <= maxChunkLength) {
        chunks.push(chunk);
        return;
      }
      var words = chunk.split(' ');
      var current = '';
      words.forEach(function (word) {
        var next = current ? (current + ' ' + word) : word;
        if (next.length > maxChunkLength && current) {
          chunks.push(current);
          current = word;
        } else {
          current = next;
        }
      });
      if (current) chunks.push(current);
    }

    var sentences = clean.match(/[^.!?;:]+[.!?;:]?/g) || [clean];
    var current = '';
    sentences.forEach(function (sentence) {
      var part = String(sentence || '').trim();
      if (!part) return;
      var next = current ? (current + ' ' + part) : part;
      if (next.length > maxChunkLength && current) {
        pushChunk(current);
        current = part;
      } else {
        current = next;
      }
    });
    if (current) pushChunk(current);
    if (!chunks.length) pushChunk(clean);
    return chunks.slice(0, 18);
  }

  function resetQueuedAssistantSpeech() {
    state.voiceQueueVersion += 1;
    state.streamingSpeechBuffer = '';
    state.voiceQueuePromise = Promise.resolve();
  }

  function extractQueuedSpeechSegments(text, force) {
    var value = String(text || '').replace(/\s+/g, ' ').trim();
    var segments = [];
    if (!value) {
      return { segments: segments, rest: '' };
    }
    var rest = value;
    var lastCut = 0;
    var match;
    var sentenceRegex = /(.+?[.!?;:])(?=\s|$)/g;
    while ((match = sentenceRegex.exec(value))) {
      var part = String(match[1] || '').trim();
      if (part) {
        segments.push(part);
      }
      lastCut = sentenceRegex.lastIndex;
    }
    rest = value.slice(lastCut).trim();
    if (!force && rest.length > 220) {
      var splitAt = rest.lastIndexOf(' ', 180);
      if (splitAt < 80) splitAt = 180;
      var early = rest.slice(0, splitAt).trim();
      if (early) {
        segments.push(early);
        rest = rest.slice(splitAt).trim();
      }
    }
    if (force && rest) {
      segments.push(rest);
      rest = '';
    }
    return { segments: segments, rest: rest };
  }

  function speakAssistantSegmentWithBrowser(text, playbackVersion) {
    if (!isSpeechSynthesisSupported() || !text) {
      return Promise.resolve(false);
    }
    return new Promise(function (resolve) {
      try {
        if (playbackVersion !== undefined && state.voicePlaybackVersion !== playbackVersion) {
          resolve(false);
          return;
        }
        var utterance = new SpeechSynthesisUtterance(String(text));
        var effectiveVoice = getEffectiveRobotVoice();
        utterance.lang = robotVoiceLang(effectiveVoice);
        utterance.rate = 1;
        utterance.pitch = getChatPersonalityMode() === 'secretary' ? 1.08 : 1;
        var desiredVoice = pickBrowserSpeechVoice(effectiveVoice);
        if (desiredVoice) {
          utterance.voice = desiredVoice;
        }
        utterance.onstart = function () {
          if (playbackVersion !== undefined && state.voicePlaybackVersion !== playbackVersion) {
            try { window.speechSynthesis.cancel(); } catch (e) {}
            resolve(false);
            return;
          }
          if (isAvatarPersonalityMode(getChatPersonalityMode())) setRobotMood('speaking');
        };
        utterance.onend = function () {
          resolve(true);
        };
        utterance.onerror = function () {
          if (isAvatarPersonalityMode(getChatPersonalityMode())) setRobotMood('error', 1600);
          resolve(false);
        };
        window.speechSynthesis.speak(utterance);
      } catch (err) {
        console.warn('No se pudo reproducir segmento de voz:', err);
        resolve(false);
      }
    });
  }

  function playAssistantSpeechSegment(text, playbackVersion) {
    var segment = sanitizeTextForSpeech(text);
    if (!segment) {
      return Promise.resolve(false);
    }
    if (state.voiceOutputMode !== 'computer' && state.voiceServerAvailable) {
      return playVoiceStreamAudio(segment, playbackVersion).then(function (played) {
        if (state.voicePlaybackVersion !== playbackVersion) {
          return false;
        }
        if (played) {
          return true;
        }
        return speakAssistantSegmentWithBrowser(segment, playbackVersion);
      });
    }
    return speakAssistantSegmentWithBrowser(segment, playbackVersion);
  }

  function queueAssistantSpeechSegment(text, playbackVersion) {
    var segment = sanitizeTextForSpeech(text);
    if (!segment) return;
    var version = playbackVersion !== undefined ? playbackVersion : state.voicePlaybackVersion;
    state.voiceQueuePromise = (state.voiceQueuePromise || Promise.resolve())
      .then(function () {
        if (state.voicePlaybackVersion !== version) {
          return false;
        }
        return playAssistantSpeechSegment(segment, version);
      })
      .catch(function () {
        return false;
      });
  }

  function beginStreamingSpeechPlayback() {
    resetQueuedAssistantSpeech();
    state.lastResponseModelMeta = null;
    if (!state.voiceEnabled && !state.conversationMode) {
      return null;
    }
    return state.voicePlaybackVersion;
  }

  function pushStreamingSpeechDelta(text, playbackVersion, force) {
    if ((!state.voiceEnabled && !state.conversationMode) || !text) {
      return;
    }
    var sanitized = sanitizeTextForSpeech(text);
    if (!sanitized) {
      return;
    }
    state.streamingSpeechBuffer = (state.streamingSpeechBuffer ? (state.streamingSpeechBuffer + ' ') : '') + sanitized;
    var extracted = extractQueuedSpeechSegments(state.streamingSpeechBuffer, !!force);
    state.streamingSpeechBuffer = extracted.rest;
    extracted.segments.forEach(function (segment) {
      queueAssistantSpeechSegment(segment, playbackVersion);
    });
  }

  function stripSendVoiceCommand(text) {
    var raw = String(text || '').trim();
    if (!raw) {
      return { text: '', shouldSend: false };
    }
    var normalized = normalizeSpeechCommandText(raw);
    var words = normalized ? normalized.split(' ') : [];
    var last = words.length ? words[words.length - 1] : '';
    if (last !== 'enviar' && last !== 'envia') {
      return { text: raw, shouldSend: false };
    }
    var withoutCommand = raw.replace(/(?:^|\s)(enviar|envia|envía)[\s.!?¿¡,;:]*$/i, '').trim();
    return { text: withoutCommand, shouldSend: true };
  }

  function getSubmitFallbackForInput(input) {
    if (input && input.id === ROBOT_INLINE_INPUT_ID) {
      return handleRobotInlineSubmit;
    }
    return handleSubmit;
  }

  function speakAssistantText(text) {
    var readAloud = state.voiceEnabled || state.conversationMode;
    if (!text) return;
    resetQueuedAssistantSpeech();
    if (isAvatarPersonalityMode(getChatPersonalityMode())) {
      setRobotMood('speaking', readAloud ? 0 : 2200);
    }
    if (!readAloud) return;
    var spokenText = sanitizeTextForSpeech(text);
    if (!spokenText) return;
    var playbackVersion = state.voicePlaybackVersion;
    if (state.voiceOutputMode !== 'computer' && state.voiceServerAvailable) {
      playVoiceStreamAudio(spokenText, playbackVersion).then(function (played) {
        if (state.voicePlaybackVersion !== playbackVersion) {
          return;
        }
        if (!played) {
          speakAssistantTextWithBrowser(spokenText, playbackVersion);
          return;
        }
        scheduleConversationMicResume(playbackVersion);
      });
      return;
    }
    speakAssistantTextWithBrowser(spokenText, playbackVersion);
  }

  function speakRobotAnnouncement(text) {
    var spokenText = sanitizeTextForSpeech(text);
    if (!spokenText) return;
    var playbackVersion = state.voicePlaybackVersion;
    if (state.voiceOutputMode !== 'computer' && state.voiceServerAvailable) {
      playVoiceStreamAudio(spokenText, playbackVersion).then(function (played) {
        if (state.voicePlaybackVersion !== playbackVersion) return;
        if (!played) speakAssistantTextWithBrowser(spokenText, playbackVersion);
      });
      return;
    }
    speakAssistantTextWithBrowser(spokenText, playbackVersion);
  }

  function speakAssistantTextWithBrowser(text, playbackVersion) {
    if (!isSpeechSynthesisSupported() || !text) return;
    try {
      if (playbackVersion !== undefined && state.voicePlaybackVersion !== playbackVersion) return;
      window.speechSynthesis.cancel();
      speakAssistantSegmentWithBrowser(text, playbackVersion).then(function (played) {
        if (played && isAvatarPersonalityMode(getChatPersonalityMode())) {
          setRobotMood('happy', 1200);
        }
        if (played) {
          scheduleConversationMicResume(playbackVersion);
        }
      });
    } catch (err) {
      console.warn('No se pudo reproducir voz:', err);
      if (isAvatarPersonalityMode(getChatPersonalityMode())) setRobotMood('error', 1600);
    }
  }

  function pickBrowserSpeechVoice(robotVoice) {
    if (!window.speechSynthesis || typeof window.speechSynthesis.getVoices !== 'function') return null;
    var voices = window.speechSynthesis.getVoices() || [];
    if (!voices.length) return null;
    var preferredLang = robotVoiceLang(robotVoice).toLowerCase();
    var wantsFemale = normalizeRobotVoice(robotVoice) === 'es-CO-female';
    var wantsMale = normalizeRobotVoice(robotVoice) === 'es-CO-male';
    var spanish = voices.filter(function (voice) {
      return String(voice.lang || '').toLowerCase().indexOf('es') === 0;
    });
    var exact = spanish.filter(function (voice) {
      return String(voice.lang || '').toLowerCase() === preferredLang;
    });
    var pool = exact.length ? exact : spanish;
    if (!pool.length) return null;
    if (wantsFemale) {
      var female = pool.find(function (voice) {
        return /female|mujer|maria|paulina|helena|sabina|monica|laura|sofia|lucia/i.test(String(voice.name || ''));
      });
      if (female) return female;
    }
    if (wantsMale) {
      var male = pool.find(function (voice) {
        return /male|hombre|carlos|jorge|diego|juan|pablo|miguel|raul/i.test(String(voice.name || ''));
      });
      if (male) return male;
    }
    return pool[0];
  }

  function playVoiceStreamAudio(text, playbackVersion) {
    if (!state.voiceServerAvailable || !text) {
      return Promise.resolve(false);
    }
    try {
      if (playbackVersion !== undefined && state.voicePlaybackVersion !== playbackVersion) {
        return Promise.resolve(false);
      }
      if (state.voiceServerAudio) {
        state.voiceServerAudio.pause();
        state.voiceServerAudio = null;
      }
      if (window.speechSynthesis) {
        window.speechSynthesis.cancel();
      }
    } catch (e) {}

    var chunks = splitTextForFastSpeech(String(text).slice(0, 4000));
    if (!chunks.length) {
      return Promise.resolve(false);
    }

    var firstChunkPlayed = false;
    var nextBlobPromise = null;

    function fetchVoiceChunk(chunk, timeoutMs) {
      if (playbackVersion !== undefined && state.voicePlaybackVersion !== playbackVersion) {
        return Promise.resolve(null);
      }
      var controller = window.AbortController ? new AbortController() : null;
      var timer = controller ? window.setTimeout(function () {
        try { controller.abort(); } catch (e) {}
      }, timeoutMs || 12000) : null;

      return fetch('/api/voice_stream/tts', {
        method: 'POST',
        credentials: 'same-origin',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ text: String(chunk).slice(0, 800), voice: getEffectiveRobotVoice() }),
        signal: controller ? controller.signal : undefined
      }).then(function (res) {
        if (timer) window.clearTimeout(timer);
        if (!res.ok) {
          if (res.status === 503 || res.status === 502 || res.status === 504) {
            state.voiceServerAvailable = false;
            updateVoiceButtons(document.getElementById(MIC_ID), document.getElementById(VOICE_ID), document.getElementById(CONV_ID));
          }
          return null;
        }
        return res.blob();
      }).then(function (blob) {
        if (!blob || typeof blob.size !== 'number' || blob.size <= 0) return null;
        return blob;
      }).catch(function (err) {
        if (timer) window.clearTimeout(timer);
        if (err && err.name !== 'AbortError') {
          console.warn('No se pudo usar el servidor de voz:', err);
        }
        return null;
      });
    }

    function playVoiceChunk(blob, isLastChunk) {
      if (!blob || playbackVersion !== undefined && state.voicePlaybackVersion !== playbackVersion) {
        return Promise.resolve(false);
      }
      return new Promise(function (resolve) {
        var url = URL.createObjectURL(blob);
        var audio = new Audio(url);
        var settled = false;
        state.voiceServerAudio = audio;

        function finish(played) {
          if (settled) return;
          settled = true;
          URL.revokeObjectURL(url);
          if (state.voiceServerAudio === audio) state.voiceServerAudio = null;
          resolve(!!played);
        }

        audio.onended = function () {
          if (isLastChunk && isAvatarPersonalityMode(getChatPersonalityMode())) {
            setRobotMood('happy', 1200);
          }
          finish(true);
        };
        audio.onerror = function () {
          if (isAvatarPersonalityMode(getChatPersonalityMode())) setRobotMood('error', 1600);
          finish(false);
        };
        if (playbackVersion !== undefined && state.voicePlaybackVersion !== playbackVersion) {
          finish(false);
          return;
        }
        audio.play().then(function () {
          if (isAvatarPersonalityMode(getChatPersonalityMode())) setRobotMood('speaking');
        }).catch(function () {
          if (isAvatarPersonalityMode(getChatPersonalityMode())) setRobotMood('error', 1600);
          finish(false);
        });
      });
    }

    function playChunkAt(index, blob) {
      if (playbackVersion !== undefined && state.voicePlaybackVersion !== playbackVersion) {
        return Promise.resolve(firstChunkPlayed);
      }
      if (!blob) {
        return Promise.resolve(firstChunkPlayed);
      }
      if (index + 1 < chunks.length) {
        nextBlobPromise = fetchVoiceChunk(chunks[index + 1], 12000);
      } else {
        nextBlobPromise = null;
      }
      return playVoiceChunk(blob, index + 1 >= chunks.length).then(function (played) {
        if (index === 0 && played) {
          firstChunkPlayed = true;
        }
        if (!played || index + 1 >= chunks.length) {
          return firstChunkPlayed;
        }
        return (nextBlobPromise || fetchVoiceChunk(chunks[index + 1], 12000)).then(function (nextBlob) {
          return playChunkAt(index + 1, nextBlob);
        });
      });
    }

    return fetchVoiceChunk(chunks[0], chunks.length > 1 ? 6500 : 15000).then(function (blob) {
      return playChunkAt(0, blob);
    }).catch(function () {
      return false;
    });
  }

  function refreshVoiceStreamStatus(micBtn, voiceBtn, convBtn) {
    fetch('/api/voice_stream/status', { credentials: 'same-origin' })
      .then(function (res) {
        if (!res.ok) return null;
        return res.json();
      })
      .then(function (data) {
        state.voiceServerChecked = true;
        state.voiceOutputMode = data && data.mode === 'natural' ? 'natural' : 'computer';
        state.computerVoiceGender = data && data.computer_voice_gender === 'male' ? 'male' : 'female';
        state.voiceServerAvailable = state.voiceOutputMode === 'natural' && !!(data && data.enabled && data.service_ok !== false);
        updateVoiceButtons(micBtn || document.getElementById(MIC_ID), voiceBtn || document.getElementById(VOICE_ID), convBtn || document.getElementById(CONV_ID));
      })
      .catch(function () {
        state.voiceServerChecked = true;
        state.voiceOutputMode = 'computer';
        state.voiceServerAvailable = false;
        updateVoiceButtons(micBtn || document.getElementById(MIC_ID), voiceBtn || document.getElementById(VOICE_ID), convBtn || document.getElementById(CONV_ID));
      });
  }

  function setupSpeechRecognition(input, micBtn, voiceBtn, convBtn) {
    if (!micBtn || !input || !isSpeechRecognitionSupported()) return;
    if (micBtn.dataset && micBtn.dataset.pcsSpeechBound === '1') {
      return;
    }
    if (micBtn.dataset) {
      micBtn.dataset.pcsSpeechBound = '1';
    }
    var SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
    var recognition = null;
    var finalText = '';
    var baseText = String(input.value || '').trim();
    var isRobotMic = micBtn.id === ROBOT_INLINE_MIC_ID;
    var sendCommandQueued = false;

    function setListening(on) {
      state.listening = !!on;
      if (on) {
        state.activeSpeechRecognition = recognition;
        state.activeSpeechSource = micBtn.id || '';
      } else if (state.activeSpeechRecognition === recognition) {
        state.activeSpeechRecognition = null;
        state.activeSpeechSource = '';
      }
      updateVoiceButtons(micBtn, voiceBtn || document.getElementById(VOICE_ID), convBtn || document.getElementById(CONV_ID));
      if (isRobotMic) {
        setRobotMood(on ? 'listening' : 'idle', on ? 0 : 800);
      }
    }

    function createRecognition() {
      var instance = new SpeechRecognition();
      instance.lang = 'es-CO';
      instance.interimResults = true;
      instance.continuous = false;
      instance.maxAlternatives = 1;

      instance.onstart = function () {
        setListening(true);
        state.preferredConversationMicId = micBtn.id || MIC_ID;
        clearConversationResumeTimer();
      };

      instance.onresult = function (event) {
        var interimText = '';
        var updatedFinalText = '';
        for (var i = event.resultIndex; i < event.results.length; i += 1) {
          var result = event.results[i];
          var transcript = String((result[0] && result[0].transcript) || '');
          if (result.isFinal) {
            updatedFinalText += transcript;
          } else {
            interimText += transcript;
          }
        }
        if (updatedFinalText) {
          finalText += updatedFinalText;
        }
        var dictatedFinal = String(finalText || '').trim();
        var sendCommand = stripSendVoiceCommand(dictatedFinal);
        if (sendCommand.shouldSend) {
          finalText = sendCommand.text;
          interimText = '';
        }
        input.value = (baseText ? baseText + ' ' : '') + String((finalText + interimText).trim());
        try {
          input.dispatchEvent(new Event('input', { bubbles: true }));
        } catch (e) {}
        if (sendCommand.shouldSend && !sendCommandQueued) {
          sendCommandQueued = true;
          setNotice('Comando de voz recibido: enviando mensaje.');
          window.setTimeout(function () {
            var form = input.form || (input.closest ? input.closest('form') : null);
            submitFormSafely(form, getSubmitFallbackForInput(input));
            sendCommandQueued = false;
          }, 80);
        }
      };

      instance.onerror = function (event) {
        var errorCode = event && event.error ? String(event.error) : '';
        setListening(false);
        if (errorCode === 'not-allowed' || errorCode === 'service-not-allowed') {
          setNotice('Permiso de micrófono denegado. Revisa los permisos del navegador.');
        } else if (errorCode === 'no-speech') {
          setNotice('No se detectó voz. Intenta hablar más cerca del micrófono.');
        } else if (errorCode === 'audio-capture') {
          setNotice('No se detectó un micrófono disponible en este equipo.');
        } else {
          setNotice(errorCode ? 'Error de micrófono: ' + errorCode + '.' : 'Error de micrófono.');
        }
      };

      instance.onend = function () {
        setListening(false);
        finalText = '';
      };

      return instance;
    }

    micBtn.addEventListener('click', function (event) {
      if (event && typeof event.preventDefault === 'function') event.preventDefault();
      if (event && typeof event.stopPropagation === 'function') event.stopPropagation();
      if (state.loading) return;
      if (state.listening && state.activeSpeechRecognition && state.activeSpeechRecognition !== recognition) {
        stopActiveSpeechRecognition(true);
      }
      if (state.listening && state.activeSpeechRecognition === recognition) {
        try { recognition.stop(); } catch (e) { }
        setListening(false);
        return;
      }
      finalText = '';
      sendCommandQueued = false;
      baseText = String(input.value || '').trim();
      recognition = createRecognition();
      try {
        recognition.start();
        setListening(true);
        if (isRobotMic) {
          setNotice('Escuchando desde el micrófono del robot...');
        }
      } catch (err) {
        setListening(false);
        setNotice('No se pudo iniciar dictado.');
      }
    });
  }

  function setupSpeechControls(input, micBtn, voiceBtn, convBtn) {
    updateVoiceButtons(micBtn, voiceBtn, convBtn);
    loadVoicePreference(micBtn, voiceBtn, convBtn);
    if (voiceBtn) {
      voiceBtn.addEventListener('click', function () {
        if (state.loading) return;
        state.voiceEnabled = !state.voiceEnabled;
        if (!state.voiceEnabled) {
          state.conversationMode = false;
          clearConversationResumeTimer();
        }
        persistVoicePreference(state.voiceEnabled);
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
          state.preferredConversationMicId = (micBtn && micBtn.id) || state.preferredConversationMicId || MIC_ID;
          persistVoicePreference(true);
          setNotice('Modo conversación: lectura automática de respuestas. Cuando la IA termine de hablar, volverá a escuchar por micrófono.');
          if (!state.loading && !state.listening && micBtn && !micBtn.disabled) {
            window.setTimeout(function () {
              if (state.conversationMode && !state.loading && !state.listening) {
                micBtn.click();
              }
            }, 180);
          }
        } else {
          clearConversationResumeTimer();
          setNotice('Modo conversación desactivado.');
        }
        updateVoiceButtons(micBtn, voiceBtn, convBtn);
      });
    }
    setupSpeechRecognition(input, micBtn, voiceBtn, convBtn);
    refreshVoiceStreamStatus(micBtn, voiceBtn, convBtn);
  }

  function syncModeUI() {
    ensureDocumentModeUI();
    var modeEl = document.getElementById(MODE_ID);
    var attachBtn = document.getElementById(ATTACH_BTN_ID);
    var clearBtn = document.getElementById(CLEAR_ATTACHMENT_ID);
    var attachName = document.getElementById(ATTACHMENT_NAME_ID);
    var reportOption = modeEl && modeEl.querySelector('option[value="reportes"]');
    var documentOption = modeEl && modeEl.querySelector('option[value="documentos"]');
    var documentTools = document.getElementById(DOCUMENT_TOOLS_ID);
    var documentFormatSelect = document.getElementById(DOCUMENT_FORMAT_ID);
    var configBtn = document.getElementById('aiChatConfigBtn');
    var attachField = attachBtn && attachBtn.closest('.ai-chat-control-field');
    var reportMode = isReportMode();
    var documentMode = isDocumentMode();
    var superContext = isSuperContext();
    var publicContext = isPublicPortalContext();

    if (modeEl && publicContext) {
      Array.prototype.slice.call(modeEl.querySelectorAll('option')).forEach(function (option) {
        var value = normalize(option && option.value);
        var allowed = value === 'operativo' || value === 'ayudante';
        option.hidden = !allowed;
        option.disabled = !allowed;
      });
      if (normalize(modeEl.value) !== 'operativo' && normalize(modeEl.value) !== 'ayudante') {
        modeEl.value = 'operativo';
      }
      reportMode = false;
      documentMode = false;
    }

    if (reportOption) {
      reportOption.hidden = superContext || publicContext;
      reportOption.disabled = superContext || publicContext;
      if ((superContext || publicContext) && normalize(modeEl.value) === 'reportes') {
        modeEl.value = 'operativo';
        reportMode = false;
      }
    }
    if (documentOption) {
      documentOption.hidden = superContext || publicContext;
      documentOption.disabled = superContext || publicContext;
      if ((superContext || publicContext) && normalize(modeEl.value) === 'documentos') {
        modeEl.value = 'operativo';
        documentMode = false;
      }
    }

    if (configBtn) {
      configBtn.hidden = publicContext;
    }
    if (attachField) {
      attachField.hidden = publicContext;
    }

    if (documentTools) {
      documentTools.classList.toggle('is-hidden', publicContext || !(documentMode || reportMode));
    }
    if (documentFormatSelect) {
      documentFormatSelect.disabled = publicContext || reportMode;
    }
    if (!documentMode) {
      updateDocumentDownloadButton();
    }
    if (!reportMode && !documentMode && !state.generatedDocument) {
      setShareArtifact(null);
    }

    if (attachBtn) attachBtn.disabled = publicContext || reportMode || documentMode;
    if (clearBtn) clearBtn.disabled = publicContext || reportMode || documentMode;
    if (attachName) {
      if (publicContext) {
        attachName.textContent = isPublicStoreContext()
          ? 'Este chat publico esta restringido a preguntas sobre los productos, precios y paginas visibles de esta empresa.'
          : 'Este chat publico responde sobre planes, modulos, precios, contacto y licencias de Powerful Control System.';
        attachName.classList.remove('is-hidden');
      } else if (reportMode) {
        attachName.textContent = 'Modo reportes: el asistente usará el flujo centralizado de reportes y exportaciones.';
        attachName.classList.remove('is-hidden');
      } else if (documentMode) {
        attachName.textContent = 'Modo Documentos IA: el sistema seleccionara automaticamente el mejor modelo disponible para generar el documento. Los adjuntos quedan desactivados; el analisis visual se reserva para fotos.';
        attachName.classList.remove('is-hidden');
      } else if (!getCurrentAttachment()) {
        attachName.textContent = '';
        attachName.classList.add('is-hidden');
      }
    }
    if ((reportMode || documentMode) && getCurrentAttachment()) {
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

  function resetChatConversation() {
    var messagesEl = document.getElementById(MESSAGES_ID);
    if (messagesEl) {
      messagesEl.innerHTML = '';
    }
    state.proposals = [];
    state.exportables = [];
    setGeneratedDocument(null);
    setShareArtifact(null);
    clearAttachmentSelection();
    var input = document.getElementById(INPUT_ID);
    if (input) {
      input.value = '';
    }
    appendMessage('assistant', getDefaultAssistantGreeting());
    setNotice('Chat reiniciado. Escribe tu nueva consulta.');
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

  function inferDocumentExportType(text) {
    var value = normalizeVoiceCommandText(text);
    if (/\bfactura\b/.test(value)) return 'factura';
    if (/\bcontrato\b/.test(value)) return 'contrato';
    if (/\bcotizacion\b|\bcotizacion\b|\bcotización\b/.test(value)) return 'cotizacion';
    if (/\bacta\b/.test(value)) return 'acta';
    if (/\breporte\b|\binforme\b/.test(value)) return 'reporte';
    if (/\btabla\b|\bexcel\b/.test(value) || String(text || '').indexOf('|') >= 0) return 'tabla';
    return 'documento';
  }

  function shouldShowDocumentExports(text) {
    var raw = String(text || '').trim();
    if (!raw) return false;
    var value = normalizeVoiceCommandText(raw);
    var hasDocumentKeyword = /\b(documento|contrato|factura|reporte|informe|acta|cotizacion|cotización|tabla|excel|presupuesto|propuesta|certificado)\b/.test(value);
    if (raw.length < 120 && raw.indexOf('|') < 0 && !hasDocumentKeyword) return false;
    return raw.indexOf('|') >= 0 ||
      hasDocumentKeyword ||
      raw.length > 650;
  }

  function createDocumentExportElement(text, sourceModule) {
    var exportIndex = state.exportables.length;
    state.exportables.push({
      content: String(text || ''),
      document_type: inferDocumentExportType(text),
      source_module: sourceModule || inferCurrentSourceModule(),
      title: inferDocumentExportTitle(text)
    });

    var section = document.createElement('section');
    section.className = 'ai-document-export-card';
    section.dataset.exportIndex = String(exportIndex);

    var title = document.createElement('div');
    title.className = 'ai-action-title';
    title.textContent = 'Exportar documento';
    section.appendChild(title);

    var note = document.createElement('div');
    note.className = 'ai-action-note';
    note.textContent = 'Descarga esta respuesta como archivo profesional generado por el sistema.';
    section.appendChild(note);

    var actionsBar = document.createElement('div');
    actionsBar.className = 'ai-document-export-actions';
    [
      ['pdf', 'PDF'],
      ['docx', 'Word'],
      ['xlsx', 'Excel'],
      ['txt', 'TXT'],
      ['json', 'JSON']
    ].forEach(function (item) {
      var btn = document.createElement('button');
      btn.type = 'button';
      btn.className = 'btn secondary ai-document-export-btn';
      btn.dataset.documentExport = String(exportIndex);
      btn.dataset.exportFormat = item[0];
      btn.textContent = item[1];
      actionsBar.appendChild(btn);
    });
    section.appendChild(actionsBar);
    return section;
  }

  function inferCurrentSourceModule() {
    var path = String(window.location.pathname || '').toLowerCase();
    if (path.indexOf('reportes') >= 0) return 'reportes';
    if (path.indexOf('chat_tareas') >= 0 || path.indexOf('chat_y_tareas') >= 0) {
      if (path.indexOf('agenda') >= 0) return 'agenda';
      if (path.indexOf('tareas') >= 0) return 'tareas';
      return 'chat_tareas';
    }
    return isSuperContext() ? 'chat_ia_global' : 'chat_ia';
  }

  function inferDocumentExportTitle(text) {
    var type = inferDocumentExportType(text);
    var firstLine = String(text || '').split(/\r?\n/).map(function (line) {
      return normalize(line).replace(/^#+\s*/, '');
    }).filter(Boolean)[0] || '';
    if (firstLine && firstLine.length <= 90) return firstLine;
    switch (type) {
      case 'factura': return 'Factura generada desde chat IA';
      case 'contrato': return 'Contrato generado desde chat IA';
      case 'cotizacion': return 'Cotizacion generada desde chat IA';
      case 'acta': return 'Acta generada desde chat IA';
      case 'reporte': return 'Reporte generado desde chat IA';
      case 'tabla': return 'Tabla generada desde chat IA';
      default: return 'Documento generado desde chat IA';
    }
  }

  function exportChatDocumentByIndex(exportIndex, format, button) {
    var item = state.exportables[Number(exportIndex)];
    if (!item) return Promise.resolve();
    return exportChatDocumentContent(item, format, button);
  }

  function exportChatDocumentContent(item, format, button) {
    if (isSuperContext()) {
      setNotice('La exportacion documental desde el chat requiere una empresa activa.', true);
      return Promise.resolve();
    }
    var empresaId = getCurrentEmpresaId();
    if (!empresaId) {
      setNotice('No se encontro empresa activa para exportar el documento.', true);
      return Promise.resolve();
    }
    var originalText = button ? button.textContent : '';
    if (button) {
      button.disabled = true;
      button.textContent = 'Generando...';
    }
    return fetch('/api/empresa/chat_documentos/exportar', {
      method: 'POST',
      credentials: 'same-origin',
      headers: {
        'Content-Type': 'application/json',
        'X-PCS-Source': 'ai_drawer'
      },
      body: JSON.stringify({
        empresa_id: parsePositiveInt(empresaId),
        title: item.title || inferDocumentExportTitle(item.content),
        content: item.content,
        input_format: String(item.content || '').indexOf('|') >= 0 ? 'markdown' : 'text',
        format: format,
        document_type: item.document_type || inferDocumentExportType(item.content),
        source_module: item.source_module || inferCurrentSourceModule(),
        metadata: {
          page_context: String(window.location.pathname || '') + String(window.location.search || ''),
          origin: 'chat_ia'
        }
      })
    }).then(function (resp) {
      if (!resp.ok) return parseErrorResponse(resp);
      return resp.json();
    }).then(function (data) {
      if (!data || data.ok === false) {
        throw new Error((data && data.error) ? String(data.error) : 'No se pudo exportar el documento.');
      }
      setShareArtifact({
        kind: 'document',
        title: normalize(data.title || item.title || 'Documento IA'),
        format: normalize(data.format || format),
        url: normalize(data.download_url),
        summary: normalize(item.content || '')
      });
      if (data.warning) {
        setNotice(String(data.warning), true);
      } else {
        setNotice('Documento generado. Iniciando descarga...');
      }
      var url = normalize(data.download_url);
      if (url) {
        window.location.href = url;
      }
    }).catch(function (err) {
      setNotice('No se pudo exportar: ' + String(err && err.message ? err.message : err), true);
    }).finally(function () {
      if (button) {
        button.disabled = false;
        button.textContent = originalText || String(format || '').toUpperCase();
      }
    });
  }

  function appendMessage(author, text, messageType, actionProposal, meta) {
    var messagesEl = document.getElementById(MESSAGES_ID);
    if (!messagesEl || !text) return;
    var item = document.createElement('div');
    item.className = 'ai-chat-message ' + author;
    if (messageType === 'error') {
      item.classList.add('error');
    }
    if (author === 'assistant') {
      var badge = createResponseModelBadge(meta);
      if (badge) {
        item.appendChild(badge);
      }
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
    if (author === 'assistant' && messageType !== 'error' && messageType !== 'document' && shouldShowDocumentExports(text)) {
      item.appendChild(createDocumentExportElement(text, inferCurrentSourceModule()));
    }

    messagesEl.appendChild(item);
    scrollChatToBottom();
    if (isAvatarPersonalityMode(getChatPersonalityMode())) {
      if (author === 'assistant') {
        setRobotAssistantText(text, messageType === 'error');
      } else if (author === 'user') {
        setRobotUserText(text);
      }
    }
  }

  function normalizeResponseModelMeta(data) {
    if (!data || typeof data !== 'object') return null;
    var modelId = normalize(data.model_id || data.modelId);
    var provider = normalize(data.provider);
    var displayName = normalize(data.display_name || data.displayName);
    var upstreamModel = normalize(data.upstream_model || data.upstreamModel);
    if (!modelId && !displayName && !upstreamModel) return null;
    return {
      model_id: modelId,
      provider: provider,
      display_name: displayName,
      upstream_model: upstreamModel
    };
  }

  function buildResponseModelLabel(meta) {
    if (!meta) return '';
    var display = normalize(meta.display_name);
    var upstream = normalize(meta.upstream_model);
    var modelId = normalize(meta.model_id);
    var provider = normalize(meta.provider).toUpperCase();
    if (display) return display;
    if (upstream && provider) return provider + ' ' + upstream;
    if (upstream) return upstream;
    if (modelId) return modelId;
    return '';
  }

  function createResponseModelBadge(meta) {
    var label = buildResponseModelLabel(meta);
    if (!label) return null;
    var badge = document.createElement('div');
    badge.className = 'ai-chat-model-badge';
    badge.textContent = 'Modelo: ' + label;
    return badge;
  }

  function ensureMessageModelBadge(container, meta) {
    if (!container) return;
    var label = buildResponseModelLabel(meta);
    var existing = container.querySelector('.ai-chat-model-badge');
    if (!label) {
      if (existing && existing.parentNode) {
        existing.parentNode.removeChild(existing);
      }
      return;
    }
    if (existing) {
      existing.textContent = 'Modelo: ' + label;
      return;
    }
    var badge = createResponseModelBadge(meta);
    if (!badge) return;
    container.insertBefore(badge, container.firstChild || null);
  }

  function appendStreamingAssistantMessage(initialText, meta) {
    var messagesEl = document.getElementById(MESSAGES_ID);
    if (!messagesEl) return null;
    var item = document.createElement('div');
    item.className = 'ai-chat-message assistant';
    item.classList.add('is-streaming');
    ensureMessageModelBadge(item, meta);
    var textNode = document.createElement('div');
    textNode.textContent = String(initialText || 'Pensando...');
    item.appendChild(textNode);
    messagesEl.appendChild(item);
    scrollChatToBottom();
    if (isAvatarPersonalityMode(getChatPersonalityMode())) {
      setRobotAssistantText(textNode.textContent);
    }
    return {
      item: item,
      textNode: textNode
    };
  }

  function updateStreamingAssistantMessage(ref, text) {
    if (!ref || !ref.textNode) return;
    var value = String(text || '').trim() || 'Pensando...';
    ref.textNode.textContent = value;
    scrollChatToBottom();
    if (isAvatarPersonalityMode(getChatPersonalityMode())) {
      setRobotAssistantText(value);
    }
  }

  function finalizeStreamingAssistantMessage(ref, text, actionProposal) {
    if (!ref || !ref.item || !ref.textNode) return;
    var value = String(text || '').trim() || 'Respuesta lista.';
    ref.item.classList.remove('is-streaming');
    ref.textNode.textContent = value;
    if (actionProposal && Array.isArray(actionProposal.actions) && actionProposal.actions.length) {
      var proposalIndex = state.proposals.length;
      state.proposals.push(actionProposal);
      ref.item.dataset.proposalIndex = String(proposalIndex);
      ref.item.appendChild(createActionProposalElement(actionProposal, proposalIndex));
    }
    if (shouldShowDocumentExports(value)) {
      ref.item.appendChild(createDocumentExportElement(value, inferCurrentSourceModule()));
    }
    scrollChatToBottom();
    if (isAvatarPersonalityMode(getChatPersonalityMode())) {
      setRobotAssistantText(value);
    }
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

  function shouldUseStreamingForTextQuery(attachment) {
    if (attachment || isDocumentMode() || isReportMode()) return false;
    if (!window.fetch || !window.TextDecoder) return false;
    return true;
  }

  function streamTextQuery(body, callbacks) {
    var endpoint = buildStreamEndpoint();
    var onStreamStart = callbacks && typeof callbacks.onStreamStart === 'function' ? callbacks.onStreamStart : null;
    var onStreamDelta = callbacks && typeof callbacks.onStreamDelta === 'function' ? callbacks.onStreamDelta : null;
    var onStreamMeta = callbacks && typeof callbacks.onStreamMeta === 'function' ? callbacks.onStreamMeta : null;
    var speechPlaybackVersion = beginStreamingSpeechPlayback();
    return fetch(endpoint, {
      method: 'POST',
      credentials: 'same-origin',
      headers: {
        'Content-Type': 'application/json',
        'X-PCS-Source': 'ai_drawer_stream'
      },
      body: JSON.stringify(body)
    }).then(function (resp) {
      if (!resp.ok) {
        return parseErrorResponse(resp);
      }
      if (!resp.body || typeof resp.body.getReader !== 'function') {
        var streamSupportError = new Error('El navegador no soporta respuestas en tiempo real para este chat.');
        streamSupportError.pcsStreamFallback = true;
        throw streamSupportError;
      }
      if (onStreamStart) {
        onStreamStart();
      }
      var reader = resp.body.getReader();
      var decoder = new TextDecoder('utf-8');
      var streamBuffer = '';
      var finalText = '';
      var doneSeen = false;
      var modelMeta = null;

      function processEventChunk(chunk) {
        var lines = String(chunk || '').split(/\r?\n/);
        lines.forEach(function (line) {
          if (!line || line.indexOf('data:') !== 0) return;
          var payload = line.slice(5).trim();
          if (!payload) return;
          var evt = null;
          try {
            evt = JSON.parse(payload);
          } catch (err) {
            return;
          }
          if (!evt || typeof evt !== 'object') return;
          if (evt.error) {
            throw new Error(String(evt.error));
          }
          var evtMeta = normalizeResponseModelMeta(evt);
          if (evtMeta) {
            modelMeta = evtMeta;
            if (onStreamMeta) {
              onStreamMeta(evtMeta);
            }
          }
          if (evt.delta) {
            finalText += String(evt.delta);
            if (onStreamDelta) {
              onStreamDelta(finalText);
            }
            pushStreamingSpeechDelta(String(evt.delta), speechPlaybackVersion, false);
          }
          if (evt.done) {
            doneSeen = true;
          }
        });
      }

      function pump() {
        return reader.read().then(function (result) {
          if (!result) {
            return;
          }
          if (result.done) {
            streamBuffer += decoder.decode();
            if (streamBuffer.trim()) {
              processEventChunk(streamBuffer);
            }
            pushStreamingSpeechDelta('', speechPlaybackVersion, true);
            scheduleConversationMicResume(speechPlaybackVersion);
            var extracted = extractPCSActionBlock(finalText);
            extracted.streamed = true;
            extracted.meta = modelMeta;
            return extracted;
          }
          streamBuffer += decoder.decode(result.value, { stream: true });
          var events = streamBuffer.split('\n\n');
          streamBuffer = events.pop() || '';
          for (var i = 0; i < events.length; i += 1) {
            processEventChunk(events[i]);
          }
          if (doneSeen) {
            pushStreamingSpeechDelta('', speechPlaybackVersion, true);
            scheduleConversationMicResume(speechPlaybackVersion);
            var extractedDone = extractPCSActionBlock(finalText);
            extractedDone.streamed = true;
            extractedDone.meta = modelMeta;
            try { reader.cancel(); } catch (e) {}
            return extractedDone;
          }
          return pump();
        });
      }

      return pump();
    });
  }

  function generateDocumentFromPrompt(query) {
    if (isSuperContext() || isPublicPortalContext()) {
      throw new Error('El modo Documentos IA requiere una empresa activa.');
    }
    var empresaId = getCurrentEmpresaId();
    if (!empresaId) {
      throw new Error('No se encontro una empresa activa para generar el documento.');
    }
    var pageContext = String(window.location.pathname || '') + String(window.location.search || '');
    return fetch('/api/empresa/chat_documentos/generar', {
      method: 'POST',
      credentials: 'same-origin',
      headers: {
        'Content-Type': 'application/json',
        'X-PCS-Source': 'ai_drawer_document_mode'
      },
      body: JSON.stringify({
        empresa_id: parsePositiveInt(empresaId),
        title: inferDocumentExportTitle(query),
        prompt: query,
        input_format: 'markdown',
        template_name: inferDocumentExportType(query),
        formats: ['pdf', 'docx', 'xlsx', 'txt', 'json'],
        metadata: {
          origin: 'chat_ia',
          source_module: inferCurrentSourceModule(),
          document_mode: true,
          page_context: pageContext
        }
      })
    }).then(function (resp) {
      if (!resp.ok) return parseErrorResponse(resp);
      return resp.json();
    }).then(function (data) {
      if (!data || data.ok === false) {
        throw new Error((data && data.error) ? String(data.error) : 'No se pudo generar el documento.');
      }
      setGeneratedDocument(data);
      var selectedFormat = getSelectedDocumentFormat().toUpperCase();
      var preview = normalize(data.preview_text);
      var text = 'Documento generado por el modelo IA seleccionado automaticamente: ' + normalize(data.title || 'Documento IA') + '.';
      text += '\nFormato seleccionado para descarga: ' + selectedFormat + '.';
      text += '\nUsa el boton Descargar para obtenerlo como PDF, Word, Excel, TXT o JSON.';
      if (preview) {
        text += '\n\nVista previa:\n' + preview;
      }
      return { clean: text, proposal: null, document: data };
    });
  }

  function sendQuery(query, attachment, callbacks) {
    var useDocumentMode = isDocumentMode() || (!attachment && shouldAutoUseDocumentMode(query));
    if (useDocumentMode) {
      if (attachment) {
        throw new Error('El modo Documentos IA no usa adjuntos. Para fotos cambia a modo operativo y adjunta la imagen.');
      }
      return generateDocumentFromPrompt(query);
    }
    if (isReportMode()) {
      if (isPublicPortalContext()) {
        throw new Error('El modo reportes no esta disponible en el chat publico.');
      }
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
        setGeneratedDocument(null);
        setShareArtifact({
          kind: 'report',
          title: normalize(data.title || 'Reporte IA'),
          format: normalize(data.format || 'pdf'),
          url: normalize(data.export_url),
          summary: normalize(data.respuesta || '')
        });
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
      modo_asistente: mode,
    };

    if (pageContext) {
      body.pagina_contexto = pageContext;
    }
    if (isPublicPortalContext()) {
      if (attachment) {
        throw new Error('El chat publico no admite adjuntos. Usa preguntas de texto sobre el portal o el catalogo visible.');
      }
      body.scope = getPublicPortalScope();
      if (body.scope === 'venta_publica') {
        var publicSlug = getPublicEmpresaSlug();
        if (!publicSlug) {
          throw new Error('No se pudo identificar la empresa publica de esta pagina para usar el chat.');
        }
        body.empresa_slug = publicSlug;
      }
    } else if (!isSuperContext()) {
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
      if (!isImageFileForAI(attachment)) {
        throw new Error('GPT-5.5 solo se usara para subir y analizar fotos o imagenes. Para documentos de texto usa el modo Documentos IA.');
      }
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

    if (!attachment && shouldUseStreamingForTextQuery(attachment)) {
      return streamTextQuery(body, callbacks).catch(function (err) {
        var canFallback = !!(err && (err.pcsStreamFallback || err.name === 'TypeError'));
        if (!canFallback) {
          throw err;
        }
        if (callbacks && typeof callbacks.onStreamFallback === 'function') {
          callbacks.onStreamFallback(err);
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
              var detailFallback = (data && data.error) ? String(data.error) : 'No se pudo obtener respuesta de IA.';
              throw new Error(detailFallback);
            }
            var answerFallback = String(data.respuesta || data.answer || data.message || 'La IA respondio sin contenido.');
            var extractedFallback = extractPCSActionBlock(answerFallback);
            extractedFallback.meta = normalizeResponseModelMeta(data);
            return extractedFallback;
          });
      });
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
        var extracted = extractPCSActionBlock(answer);
        extracted.meta = normalizeResponseModelMeta(data);
        return extracted;
      });
  }

  function handleRobotInlineSubmit(event) {
    if (event && typeof event.preventDefault === 'function') {
      event.preventDefault();
    }
    if (!state.chatEnabled || !state.robotEnabled) return;
    if (state.loading) return;
    if (state.listening) {
      stopActiveSpeechRecognition(true);
    }
    var input = document.getElementById(ROBOT_INLINE_INPUT_ID);
    if (!input) return;
    var query = String(input.value || '').trim();
    if (!query) return;
    if (state.setupWizard && state.setupWizard.active) {
      input.value = '';
      input.style.height = 'auto';
      return handleGuidedSetupAnswer(query);
    }

    input.value = '';
    input.style.height = 'auto';
    setRobotInlineVisible(true);
    setRobotUserText(query);
    var wantsSecretary = shouldAutoEnableSecretary(query);
    var wantsRobot = !wantsSecretary && shouldAutoEnableRobot(query);
    var wantsVoice = shouldAutoEnableVoice(query);
    if (wantsSecretary) {
      activateSecretaryFromCommand();
    } else if (wantsRobot) {
      activateRobotFromCommand();
    }
    if (wantsVoice) {
      activateVoiceFromCommand(document.getElementById(MIC_ID), document.getElementById(VOICE_ID), document.getElementById(CONV_ID));
    }
    if ((wantsSecretary || wantsRobot || wantsVoice) && (isOnlyLocalPreferenceCommand(query) || isOnlyVoiceEnableCommand(query))) {
      var preferenceReadyMessage = wantsSecretary ? 'Listo. Active la secretaria IA 3D con voz femenina y guarde esta preferencia para los proximos reinicios.' : buildPreferenceCommandMessage(wantsRobot, wantsVoice);
      setRobotAssistantText(preferenceReadyMessage);
      speakAssistantText(preferenceReadyMessage);
      focusRobotInput();
      return;
    }
    setRobotLoading(true);
    state.loading = true;
    updateVoiceButtons(document.getElementById(MIC_ID), document.getElementById(VOICE_ID), document.getElementById(CONV_ID));

    sendQuery(query, null, {
      onStreamStart: function () {
        setLastResponseModelMeta(null);
        setRobotAssistantText('Respondiendo en tiempo real...');
        setNotice('Respondiendo en tiempo real...');
      },
      onStreamMeta: function (meta) {
        setLastResponseModelMeta(meta);
      },
      onStreamDelta: function (text) {
        setRobotAssistantText(text || 'Respondiendo en tiempo real...');
      },
      onStreamFallback: function () {
        setNotice('El modo en tiempo real no estuvo disponible. Continuo con respuesta normal.');
      }
    }).then(function (result) {
      setLastResponseModelMeta(result && result.meta ? result.meta : null);
      var answer = result && result.clean ? result.clean : 'Respuesta lista.';
      var hasActions = !!(result && result.proposal && Array.isArray(result.proposal.actions) && result.proposal.actions.length);
      if (hasActions) {
        answer += '\n\nPrepare acciones sugeridas. Puedes confirmarlas desde estos botones del robot.';
      }
      setRobotAssistantText(answer);
      if (hasActions) {
        setRobotMood('action', 3200);
        renderRobotProposalActions(result.proposal);
      } else if (result && result.document) {
        setRobotMood('action', 3200);
        renderRobotGeneratedDocumentActions(result.document);
      } else {
        renderRobotDocumentExportActions(answer);
      }
      if (!(result && result.streamed)) {
        speakAssistantText(answer);
      }
      setNotice('Respuesta lista desde el robot.');
    }).catch(function (err) {
      var message = err && err.message ? err.message : 'Error al procesar la consulta.';
      setRobotMood('error', 2600);
      setRobotAssistantText(message, true);
      setNotice('No se pudo completar la solicitud. ' + String(message), true);
    }).finally(function () {
      state.loading = false;
      setRobotLoading(false);
      updateVoiceButtons(document.getElementById(MIC_ID), document.getElementById(VOICE_ID), document.getElementById(CONV_ID));
      focusRobotInput();
    });
  }

  function sendRobotPrompt(prompt) {
    if (!state.chatEnabled || !state.robotEnabled) return false;
    var input = document.getElementById(ROBOT_INLINE_INPUT_ID);
    var value = normalize(prompt);
    if (!input || !value || state.loading) return;
    if (value.indexOf('__PCS_ROBOT_CONFIRM_ACTIONS__') === 0) {
      clearRobotActionChips();
      executeActionProposal(parseInt(value.replace('__PCS_ROBOT_CONFIRM_ACTIONS__', ''), 10));
      return;
    }
    if (value.indexOf('__PCS_ROBOT_CANCEL_ACTIONS__') === 0) {
      clearRobotActionChips();
      cancelActionProposal(parseInt(value.replace('__PCS_ROBOT_CANCEL_ACTIONS__', ''), 10));
      setRobotAssistantText('Acciones canceladas. Puedes pedirme otra configuración o ajustar la propuesta.');
      return;
    }
    if (value === '__PCS_START_GUIDED_SETUP__') {
      loadAndStartGuidedSetup();
      return true;
    }
    if (value.indexOf('__PCS_GUIDED_SETUP_OPTION__') === 0) {
      handleGuidedSetupAnswer(value.replace('__PCS_GUIDED_SETUP_OPTION__', ''));
      return true;
    }
    if (value === '__PCS_GUIDED_SETUP_CANCEL__') {
      cancelGuidedSetup();
      return true;
    }
    input.value = value;
    input.style.height = 'auto';
    input.style.height = Math.min(input.scrollHeight, 96) + 'px';
    clearRobotActionChips();
    handleRobotInlineSubmit({ preventDefault: function () {} });
    return true;
  }

  function getGuidedSetupEndpoint() {
    var empresaId = getCurrentEmpresaId();
    if (!empresaId) {
      throw new Error('No se encontro una empresa activa para iniciar la configuracion guiada.');
    }
    return '/api/empresa/configuracion_guiada?empresa_id=' + encodeURIComponent(String(empresaId));
  }

  function normalizeGuidedSetupOptionLabel(question, raw) {
    var value = normalize(raw).toLowerCase();
    if (question && question.type === 'boolean') {
      return value === 'si' ? 'Si' : 'No';
    }
    if (value === 'comprobante_pago') return 'Comprobante de pago';
    if (value === 'factura_electronica') return 'Factura electronica';
    return raw;
  }

  function buildGuidedSetupQuestionActions(question) {
    if (!question) return [];
    var actions = [];
    if (question.type === 'boolean') {
      actions.push({ label: 'Si', prompt: '__PCS_GUIDED_SETUP_OPTION__si' });
      actions.push({ label: 'No', prompt: '__PCS_GUIDED_SETUP_OPTION__no' });
    } else if (Array.isArray(question.options) && question.options.length) {
      question.options.forEach(function (option) {
        actions.push({
          label: normalizeGuidedSetupOptionLabel(question, option),
          prompt: '__PCS_GUIDED_SETUP_OPTION__' + String(option)
        });
      });
    }
    actions.push({ label: 'Cancelar guía', prompt: '__PCS_GUIDED_SETUP_CANCEL__' });
    return actions;
  }

  function renderGuidedSetupQuestion() {
    var wizard = state.setupWizard;
    if (!wizard || !wizard.active) return false;
    var question = wizard.questions[wizard.index];
    if (!question) return false;
    var prefix = wizard.index === 0
      ? 'Vamos a terminar la configuracion base de esta empresa con preguntas cortas.'
      : 'Perfecto. Continuemos con la siguiente decision.';
    var text = prefix + '\n\n' + (question.prompt || question.label || 'Responde este dato.');
    if (question.help) {
      text += '\n\n' + question.help;
    }
    if (question.default_value) {
      text += '\n\nValor sugerido: ' + question.default_value;
    }
    setRobotAssistantText(text);
    renderRobotActionChips(buildGuidedSetupQuestionActions(question));
    setNotice('Configuracion guiada en curso: ' + (question.label || 'pregunta'));
    focusRobotInput();
    return true;
  }

  function startGuidedSetupWizard(payload) {
    if (!state.chatEnabled || !state.robotEnabled) return false;
    var wizardPayload = payload && payload.wizard ? payload.wizard : payload;
    var questions = wizardPayload && Array.isArray(wizardPayload.questions) ? wizardPayload.questions.slice() : [];
    if (!questions.length) {
      setRobotAssistantText('No encontre preguntas disponibles para esta configuracion guiada. Puedes abrir la pagina de configuracion y revisar el contexto de la empresa.');
      renderRobotActionChips([]);
      return false;
    }
    state.setupWizard = {
      active: true,
      context: payload || {},
      questions: questions,
      answers: {},
      index: 0
    };
    setRobotMood('thinking', 2200);
    return renderGuidedSetupQuestion();
  }

  function cancelGuidedSetup() {
    state.setupWizard = null;
    clearRobotActionChips();
    setRobotAssistantText('Configuracion guiada cancelada. Puedes retomarla cuando quieras desde Configuracion > Configuracion guiada.');
    setNotice('Configuracion guiada cancelada.');
    focusRobotInput();
    return true;
  }

  function completeGuidedSetupWizard() {
    var wizard = state.setupWizard;
    if (!wizard || !wizard.active) return false;
    var answers = Object.assign({}, wizard.answers || {});
    state.setupWizard = null;
    setRobotLoading(true);
    state.loading = true;
    clearRobotActionChips();
    setRobotAssistantText('Estoy aplicando la configuracion guiada en la empresa...');
    setNotice('Aplicando configuracion guiada...');
    return fetch(getGuidedSetupEndpoint() + '&action=aplicar', {
      method: 'POST',
      credentials: 'same-origin',
      headers: {
        'Content-Type': 'application/json',
        'X-PCS-Source': 'ai_drawer'
      },
      body: JSON.stringify({ answers: answers })
    }).then(function (resp) {
      if (!resp.ok) {
        return parseErrorResponse(resp);
      }
      return resp.json();
    }).then(function (data) {
      var result = data && data.resultado ? data.resultado : {};
      var summary = result && result.resumen ? result.resumen : {};
      var message = normalize(result.mensaje || 'Configuracion guiada aplicada.');
      if (summary && summary.nombre_comercial) {
        message += '\n\nNombre comercial: ' + summary.nombre_comercial;
      }
      if (summary && summary.cantidad_estaciones) {
        message += '\nEstaciones: ' + summary.cantidad_estaciones;
      }
      if (summary && summary.modo_documento_venta) {
        message += '\nDocumento: ' + (summary.modo_documento_venta === 'factura_electronica' ? 'Factura electronica' : 'Comprobante de pago');
      }
      setRobotMood('success', 3200);
      setRobotAssistantText(message);
      var actions = [];
      if (Array.isArray(result.acciones)) {
        result.acciones.forEach(function (item) {
          if (!item || !item.url) return;
          actions.push({
            label: normalize(item.label || 'Abrir modulo'),
            prompt: 'Abre la pagina ' + item.url + ' y ayudame a revisar la configuracion aplicada.'
          });
        });
      }
      renderRobotActionChips(actions);
      setNotice('Configuracion guiada aplicada correctamente.');
      speakRobotAnnouncement(message);
      return true;
    }).catch(function (err) {
      var message = err && err.message ? err.message : 'No se pudo aplicar la configuracion guiada.';
      setRobotMood('error', 2600);
      setRobotAssistantText(message, true);
      setNotice(message, true);
      return false;
    }).finally(function () {
      state.loading = false;
      setRobotLoading(false);
      updateVoiceButtons(document.getElementById(MIC_ID), document.getElementById(VOICE_ID), document.getElementById(CONV_ID));
      focusRobotInput();
    });
  }

  function handleGuidedSetupAnswer(rawValue) {
    var wizard = state.setupWizard;
    if (!wizard || !wizard.active || state.loading) return false;
    var question = wizard.questions[wizard.index];
    if (!question) return false;
    var answer = normalize(rawValue);
    if (!answer) {
      return false;
    }
    wizard.answers[question.id] = answer;
    setRobotUserText(answer);
    wizard.index += 1;
    if (wizard.index >= wizard.questions.length) {
      return completeGuidedSetupWizard();
    }
    return renderGuidedSetupQuestion();
  }

  function loadAndStartGuidedSetup() {
    if (!state.chatEnabled || !state.robotEnabled || state.loading) return false;
    setRobotLoading(true);
    state.loading = true;
    setRobotAssistantText('Estoy preparando la configuracion guiada de esta empresa...');
    setNotice('Preparando configuracion guiada...');
    fetch(getGuidedSetupEndpoint(), {
      credentials: 'same-origin',
      headers: { 'X-PCS-Source': 'ai_drawer' }
    }).then(function (resp) {
      if (!resp.ok) {
        return parseErrorResponse(resp);
      }
      return resp.json();
    }).then(function (data) {
      startGuidedSetupWizard(data || {});
    }).catch(function (err) {
      var message = err && err.message ? err.message : 'No se pudo iniciar la configuracion guiada.';
      setRobotMood('error', 2600);
      setRobotAssistantText(message, true);
      setNotice(message, true);
    }).finally(function () {
      state.loading = false;
      setRobotLoading(false);
      updateVoiceButtons(document.getElementById(MIC_ID), document.getElementById(VOICE_ID), document.getElementById(CONV_ID));
      focusRobotInput();
    });
    return true;
  }

  function buildConfigurationAssistantActions(summary) {
    var tipo = normalize(summary && summary.tipo_empresa_nombre);
    var tipoText = tipo ? (' para una empresa tipo ' + tipo) : '';
    return [
      {
        label: 'Configuracion guiada',
        prompt: '__PCS_START_GUIDED_SETUP__'
      },
      {
        label: 'Agregar productos',
        prompt: 'Actúa como asistente de configuración inicial' + tipoText + '. Revisa el contexto de preconfiguración y guíame para agregar o ajustar productos, categorías, precios, costos, impuestos y stock mínimo. Si puedes proponer acciones seguras, proponlas para confirmarlas.'
      },
      {
        label: 'Configurar tarifas',
        prompt: 'Actúa como asistente de configuración inicial' + tipoText + '. Ayúdame a configurar tarifas por minutos, por día o por servicio según el tipo de empresa. Pregúntame los datos faltantes y propón una configuración profesional.'
      },
      {
        label: 'Estaciones y caja',
        prompt: 'Actúa como asistente de configuración inicial' + tipoText + '. Revisa estaciones, nombres, caja, carrito, notas y vista móvil. Guíame para dejar operativa la empresa sin romper la configuración actual.'
      },
      {
        label: 'Usuarios y roles',
        prompt: 'Actúa como asistente de configuración inicial' + tipoText + '. Guíame para convertir usuarios guía en usuarios reales, asignar roles, permisos y tareas iniciales.'
      },
      {
        label: 'Facturación',
        prompt: 'Actúa como asistente de configuración inicial' + tipoText + '. Guíame por la configuración de facturación, resoluciones, impuestos, DIAN si aplica y pruebas necesarias antes de vender.'
      },
      {
        label: 'Plan completo',
        prompt: 'Actúa como asistente de configuración inicial' + tipoText + '. Dame un plan paso a paso para terminar productos, tarifas, estaciones, usuarios, facturación, reportes y auditoría usando la preconfiguración de esta empresa.'
      }
    ];
  }

  function startConfigurationAssistant(summary) {
    if (!state.chatEnabled || !state.robotEnabled) return false;
    var toggleBtn = document.getElementById(TOGGLE_ID);
    setChatPersonalityMode('robot');
    closeChatDrawerFully();
    ensureRobotInlineUI(toggleBtn);
    showRobotAssistant(toggleBtn);
    var tipo = normalize(summary && summary.tipo_empresa_nombre);
    var estaciones = parsePositiveInt(summary && summary.estaciones_creadas);
    var productos = parsePositiveInt(summary && summary.productos_creados);
    var usuarios = parsePositiveInt(summary && summary.usuarios_creados);
    var intro = 'Hola. Soy tu robot asistente de configuración. ';
    if (tipo || estaciones || productos || usuarios) {
      intro += 'Detecté una preconfiguración inicial';
      if (tipo) intro += ' para ' + tipo;
      intro += '. ';
      intro += 'Ya puedo ayudarte a revisar';
      if (estaciones) intro += ' ' + estaciones + ' estaciones,';
      if (productos) intro += ' ' + productos + ' productos guía,';
      if (usuarios) intro += ' ' + usuarios + ' usuarios guía,';
      intro = intro.replace(/,\s*$/, '') + '. ';
    }
    intro += 'Elige una opción o escríbeme qué quieres configurar: productos, tarifas, estaciones, usuarios, facturación o reportes.';
    setRobotAssistantText(intro);
    renderRobotActionChips(buildConfigurationAssistantActions(summary || {}));
    setNotice('Asistente de configuración inicial activo.');
    speakRobotAnnouncement(intro);
    focusRobotInput();
    if (summary && summary.auto_start_guided_setup) {
      window.setTimeout(function () {
        loadAndStartGuidedSetup();
      }, 320);
    }
    return true;
  }

  function notifyRobotReminder(payload) {
    if (!state.chatEnabled || !state.robotEnabled) return false;
    var toggleBtn = document.getElementById(TOGGLE_ID);
    setChatPersonalityMode('robot');
    closeChatDrawerFully();
    ensureRobotInlineUI(toggleBtn);
    showRobotAssistant(toggleBtn);
    var title = normalize(payload && payload.title) || 'Recordatorio de notas';
    var detail = normalize(payload && payload.detail);
    var message = 'Tiempo cumplido: ' + title + '.';
    if (detail) {
      message += ' ' + detail;
    }
    setRobotAssistantText(message);
    renderRobotActionChips([
      { label: 'Qué hago ahora', prompt: 'Se cumplió un recordatorio de nota: "' + title + '". Ayúdame a decidir el siguiente paso operativo con una respuesta corta y accionable.' },
      { label: 'Crear tarea', prompt: 'Se cumplió un recordatorio de nota: "' + title + '". Guíame para crear una tarea o seguimiento relacionado y dejar evidencia en el sistema.' }
    ]);
    setNotice('Recordatorio de notas cumplido.');
    speakRobotAnnouncement(message);
    return true;
  }

  function handleSubmit(event) {
    event.preventDefault();
    if (!state.chatEnabled) return;
    if (state.loading) return;
    if (state.listening) {
      stopActiveSpeechRecognition(true);
    }
    var input = document.getElementById(INPUT_ID);
    if (!input) return;

    var query = String(input.value || '').trim();
    var attachment = getCurrentAttachment();
    if (!query) return;

    input.value = '';
    appendMessage('user', attachment ? (query + '\n\n[Adjunto: ' + describeAttachment(attachment) + ']') : query);
    var wantsSecretary = shouldAutoEnableSecretary(query);
    var wantsRobot = !wantsSecretary && shouldAutoEnableRobot(query);
    var wantsVoice = shouldAutoEnableVoice(query);
    if (wantsSecretary) {
      activateSecretaryFromCommand();
    } else if (wantsRobot) {
      activateRobotFromCommand();
    }
    if (wantsVoice) {
      activateVoiceFromCommand(document.getElementById(MIC_ID), document.getElementById(VOICE_ID), document.getElementById(CONV_ID));
    }
    if ((wantsSecretary || wantsRobot || wantsVoice) && !attachment && (isOnlyLocalPreferenceCommand(query) || isOnlyVoiceEnableCommand(query))) {
      var preferenceReadyMessage = wantsSecretary ? 'Listo. Active la secretaria IA 3D con voz femenina y guarde esta preferencia para los proximos reinicios.' : buildPreferenceCommandMessage(wantsRobot, wantsVoice);
      appendMessage('assistant', preferenceReadyMessage);
      setRobotAssistantText(preferenceReadyMessage);
      speakAssistantText(preferenceReadyMessage);
      return;
    }
    setNotice(attachment ? 'Procesando consulta con adjunto...' : 'Procesando tu consulta...');
    if (isAvatarPersonalityMode(getChatPersonalityMode())) setRobotMood('thinking');
    state.loading = true;
    updateVoiceButtons(document.getElementById(MIC_ID), document.getElementById(VOICE_ID), document.getElementById(CONV_ID));

    var liveAssistantMessage = null;
    var liveAssistantMeta = null;
    sendQuery(query, attachment, {
      onStreamStart: function () {
        setLastResponseModelMeta(null);
        liveAssistantMessage = appendStreamingAssistantMessage('Respondiendo en tiempo real...', liveAssistantMeta);
        setNotice('Respondiendo en tiempo real...');
      },
      onStreamMeta: function (meta) {
        liveAssistantMeta = meta;
        setLastResponseModelMeta(meta);
        if (!liveAssistantMessage) {
          liveAssistantMessage = appendStreamingAssistantMessage('Respondiendo en tiempo real...', liveAssistantMeta);
        } else if (liveAssistantMessage.item) {
          ensureMessageModelBadge(liveAssistantMessage.item, liveAssistantMeta);
        }
      },
      onStreamDelta: function (text) {
        if (!liveAssistantMessage) {
          liveAssistantMessage = appendStreamingAssistantMessage('Respondiendo en tiempo real...', liveAssistantMeta);
        }
        updateStreamingAssistantMessage(liveAssistantMessage, text);
      },
      onStreamFallback: function () {
        if (liveAssistantMessage && liveAssistantMessage.item && liveAssistantMessage.item.parentNode) {
          liveAssistantMessage.item.parentNode.removeChild(liveAssistantMessage.item);
        }
        liveAssistantMessage = null;
        setNotice('El modo en tiempo real no estuvo disponible. Continuo con respuesta normal.');
      }
    }).then(function (result) {
      setLastResponseModelMeta(result && result.meta ? result.meta : null);
      var hasActions = !!(result && result.proposal && Array.isArray(result.proposal.actions) && result.proposal.actions.length);
      if (liveAssistantMessage && result && result.streamed) {
        finalizeStreamingAssistantMessage(liveAssistantMessage, result.clean, result.proposal);
      } else {
        appendMessage('assistant', result.clean, result && result.document ? 'document' : null, result.proposal, result && result.meta ? result.meta : null);
      }
      if (hasActions) setRobotMood('action', 3200);
      if (!(result && result.streamed)) {
        speakAssistantText(result.clean);
      }
      setNotice(result && result.document ? 'Documento listo. Elige formato y presiona Descargar.' : 'Respuesta lista. Puedes seguir escribiendo otra consulta.');
      clearAttachmentSelection();
    }).catch(function (err) {
      if (isAvatarPersonalityMode(getChatPersonalityMode())) setRobotMood('error', 2600);
      if (liveAssistantMessage && liveAssistantMessage.item && liveAssistantMessage.item.parentNode) {
        liveAssistantMessage.item.parentNode.removeChild(liveAssistantMessage.item);
      }
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
    if (!state.chatEnabled) return;
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
    ensureDrawerShell();
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
    var clearBtn = document.getElementById(CLEAR_CHAT_ID);
    var modeEl = document.getElementById(MODE_ID);
    var input = document.getElementById(INPUT_ID);

    if (!toggle || !drawer || !closeBtn || !form || !messagesEl) return;
    var submitBtn = form.querySelector('button[type="submit"]');
    ensureDocumentModeUI();

    toggle.addEventListener('click', function () {
      if (!state.chatEnabled) return;
      if (isAvatarPersonalityMode(getChatPersonalityMode())) {
        closeChatDrawerFully();
        showRobotAssistant(toggle);
        return;
      }
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
    if (submitBtn) {
      submitBtn.addEventListener('click', function (event) {
        event.preventDefault();
        event.stopPropagation();
        submitFormSafely(form, handleSubmit);
      });
    }
    if (modeEl) {
      modeEl.addEventListener('change', function () {
        syncModeUI();
        setNotice(isDocumentMode()
          ? 'Modo Documentos IA activo. El sistema elegira automaticamente el mejor modelo disponible para generar documentos descargables.'
          : (isReportMode()
            ? 'Modo reportes activo. Este chat central usará el flujo de reportes y exportaciones de la empresa.'
            : 'Modo actualizado. Puedes seguir consultando normalmente.'));
      });
    }
    if (hintToggle && hints) {
      hintToggle.addEventListener('click', function () {
        hints.classList.toggle('is-hidden');
        hintToggle.textContent = hints.classList.contains('is-hidden') ? 'Ver ejemplos' : 'Ocultar ejemplos';
      });
    }

    var configBtn = document.getElementById('aiChatConfigBtn');
    if (configBtn) {
      configBtn.addEventListener('click', function () {
        openChatConfigPage();
      });
    }

    if (attachBtn && attachInput) {
      attachBtn.addEventListener('click', function () {
        if (state.loading) return;
        if (isReportMode()) {
          setNotice('El modo reportes no admite adjuntos en este flujo.', true);
          return;
        }
        if (isDocumentMode()) {
          setNotice('El modo Documentos IA no usa adjuntos. Para fotos cambia a modo operativo.', true);
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
        if (!isImageFileForAI(file)) {
          clearAttachmentSelection();
          setNotice('GPT-5.5 solo se usa para subir y analizar fotos o imagenes. Para documentos usa el modo Documentos IA.', true);
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

    if (clearBtn) {
      clearBtn.addEventListener('click', function () {
        if (state.loading) return;
        resetChatConversation();
      });
    }

    setupSpeechControls(input, document.getElementById(MIC_ID), document.getElementById(VOICE_ID), document.getElementById(CONV_ID));

    messagesEl.addEventListener('click', function (event) {
      var target = event.target;
      if (!target) return;
      var confirmButton = target.closest('button[data-action-confirm]');
      var cancelButton = target.closest('button[data-action-cancel]');
      var exportButton = target.closest('button[data-document-export]');
      if (confirmButton) {
        executeActionProposal(parseInt(confirmButton.dataset.actionConfirm, 10));
      } else if (cancelButton) {
        cancelActionProposal(parseInt(cancelButton.dataset.actionCancel, 10));
      } else if (exportButton) {
        exportChatDocumentByIndex(exportButton.dataset.documentExport, exportButton.dataset.exportFormat, exportButton);
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
          submitFormSafely(form, handleSubmit);
        }
      });
    }

    window.addEventListener('message', function (event) {
      var data = event && event.data;
      if (!data) return;
      if (data.type === 'pcs-ai-chat-personality-updated') {
        setChatPersonalityMode(data.mode);
        return;
      }
      if (data.type === 'pcs-ai-chat-enabled-updated') {
        setChatEnabledPreference(!!data.enabled);
        return;
      }
      if (data.type === 'pcs-ai-robot-enabled-updated') {
        setRobotEnabledPreference(!!data.enabled);
        return;
      }
      if (data.type === 'pcs-ai-chat-voice-updated') {
        applyVoicePreference(!!data.enabled);
        return;
      }
      if (data.type === 'pcs-ai-chat-robot-voice-updated') {
        setRobotVoicePreference(data.voice);
        return;
      }
      if (data.type === 'pcs-ai-config-assistant-start') {
        if (!state.chatEnabled || !state.robotEnabled) return;
        startConfigurationAssistant(data.summary || data.preconfiguracion || {});
        return;
      }
      if (data.type === 'pcs-ai-config-wizard-start') {
        if (!state.chatEnabled || !state.robotEnabled) return;
        var toggleBtnWizard = document.getElementById(TOGGLE_ID);
        setChatPersonalityMode('robot');
        closeChatDrawerFully();
        ensureRobotInlineUI(toggleBtnWizard);
        showRobotAssistant(toggleBtnWizard);
        startGuidedSetupWizard(data.payload || {});
        return;
      }
      if (data.type === 'pcs-ai-robot-reminder') {
        if (!state.chatEnabled || !state.robotEnabled) return;
        notifyRobotReminder(data.payload || data);
        return;
      }
      if (data.type !== 'pcs-ai-drawer-open') return;
      if (!state.chatEnabled) return;
      if (isAvatarPersonalityMode(getChatPersonalityMode())) {
        closeChatDrawerFully();
        showRobotAssistant(document.getElementById(TOGGLE_ID));
      } else {
        openChatDrawerFromUser();
      }
      if (modeEl && normalize(data.mode)) {
        modeEl.value = normalize(data.mode);
        syncModeUI();
      }
      if (input && normalize(data.prompt)) {
        input.value = normalize(data.prompt);
      }
      if (isAvatarPersonalityMode(getChatPersonalityMode())) {
        var robotInput = document.getElementById(ROBOT_INLINE_INPUT_ID);
        if (robotInput && normalize(data.prompt)) {
          robotInput.value = normalize(data.prompt);
        }
      }
      window.setTimeout(function () {
        if (isAvatarPersonalityMode(getChatPersonalityMode())) {
          focusRobotInput();
        } else if (input) {
          input.focus();
        }
      }, 50);
    });

    if (!messagesEl.querySelector('.ai-chat-message')) {
      appendMessage('assistant', getDefaultAssistantGreeting());
    }

    renderAttachmentState();
    syncModeUI();
    applyChatPersonalityMode();

    window.PCSAIChatRobot = {
      startConfigurationAssistant: startConfigurationAssistant,
      startGuidedSetupWizard: startGuidedSetupWizard,
      notifyReminder: notifyRobotReminder,
      showMessage: function (text, options) {
        if (!state.chatEnabled || !state.robotEnabled) return false;
        var toggleBtn = document.getElementById(TOGGLE_ID);
        setChatPersonalityMode('robot');
        closeChatDrawerFully();
        ensureRobotInlineUI(toggleBtn);
        showRobotAssistant(toggleBtn);
        setRobotAssistantText(text);
        renderRobotActionChips(options && options.actions);
        if (options && options.speak) {
          speakRobotAnnouncement(text);
        }
        return true;
      },
      sendPrompt: sendRobotPrompt
    };

    window.dispatchEvent(new CustomEvent('pcs-ai-chat-robot-ready'));

    window.addEventListener('storage', function (event) {
      if (!event || event.key !== CHAT_PERSONALITY_STORAGE_KEY) return;
      applyChatPersonalityMode();
    });
  }

  document.addEventListener('DOMContentLoaded', initDrawer);
})();




