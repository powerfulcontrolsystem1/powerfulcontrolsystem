(function setupPCSButtonIcons() {
  "use strict";

  var rules = [
    { re: /pagar|cobrar|facturar|checkout|licencia|comprar|adquirir|renovar|pago|payment|total/i, symbol: "$", tone: "success" },
    { re: /agregar|anadir|añadir|nuevo|nueva|crear|registrar|activar|encender|online|iniciar|aceptar|confirmar|guardar|save/i, symbol: "+", tone: "success" },
    { re: /guardar|aplicar|confirmar|listo|aprobar|validar|finalizar/i, symbol: "✓", tone: "success" },
    { re: /buscar|catalogo|catálogo|consultar|filtrar|verificar|explorar|scanner|codigo|código|lupa/i, symbol: "⌕", tone: "info" },
    { re: /actualizar|recargar|reintentar|reiniciar|sincronizar|sync|refresh|reload|evaluar ahora/i, symbol: "↺", tone: "info" },
    { re: /editar|modificar|ajustar|configurar|configuracion|configuración|preferencias|settings/i, symbol: "⚙", tone: "slate" },
    { re: /eliminar|borrar|quitar|cancelar|rechazar|anular|detener|apagar|cerrar|delete|remove|stop|danger/i, symbol: "X", tone: "danger" },
    { re: /descargar|exportar|csv|excel|pdf|json|respaldo|backup|download/i, symbol: "↓", tone: "teal" },
    { re: /subir|cargar|importar|upload|restaurar|adjuntar|archivo|firma/i, symbol: "↑", tone: "purple" },
    { re: /imprimir|print|ticket|pos/i, symbol: "P", tone: "slate" },
    { re: /correo|email|mail|gmail|smtp|enviar|invitacion|invitación/i, symbol: "@", tone: "purple" },
    { re: /whatsapp|compartir|share|publicar|red social|canales|digital|colaboracion|colaboración/i, symbol: "↗", tone: "teal" },
    { re: /cliente|usuario|empleado|asesor|administrador|\brol\b|permiso|vip|perfil|login|entrar|sesion|sesión/i, symbol: "ID", tone: "purple" },
    { re: /producto|item|servicio|receta|inventario|bodega|stock|pedido|orden|soluciones|negocio|plantillas/i, symbol: "■", tone: "teal" },
    { re: /reporte|informe|historial|movimiento|auditoria|auditoría|dashboard|panel|vista previa|analisis|análisis|control/i, symbol: "R", tone: "info" },
    { re: /caja|turno|corte|egreso|ingreso|finanza|banco|impuesto|saldo|operacion|operación|venta|ventas/i, symbol: "$", tone: "success" },
    { re: /personas|activos|nomina|nómina|empleados/i, symbol: "ID", tone: "purple" },
    { re: /administracion|administración|sistema|super/i, symbol: "⚙", tone: "slate" },
    { re: /qr|nequi|bre|wompi|epayco|datafono|datáfono|tarjeta|transferencia/i, symbol: "QR", tone: "teal" },
    { re: /ayuda|soporte|ticket|chat|ia|pregunta|test|probar/i, symbol: "?", tone: "warning" },
    { re: /ubicacion|ubicación|gps|mapa|ruta|domicilio|taxi|vehiculo|vehículo/i, symbol: "⌖", tone: "warning" },
    { re: /calendario|agenda|cita|fecha|programar|turnos/i, symbol: "D", tone: "info" },
    { re: /radio|play|pausar|sonido|audio|voz|microfono|mic/i, symbol: "▶", tone: "purple" },
    { re: /volver|regresar|atras|atrás|back/i, symbol: "←", tone: "slate" },
    { re: /siguiente|continuar|abrir|ir a|ver /i, symbol: "→", tone: "info" }
  ];

  function normalize(value) {
    return String(value || "").replace(/\s+/g, " ").trim();
  }

  function buttonKey(button) {
    var dataset = "";
    if (button.dataset) {
      dataset = Object.keys(button.dataset).map(function(key) {
        return key + " " + button.dataset[key];
      }).join(" ");
    }
    return [
      button.id,
      button.className,
      dataset,
      button.getAttribute("name"),
      button.getAttribute("type"),
      button.getAttribute("aria-label"),
      button.getAttribute("title"),
      button.textContent
    ].map(normalize).filter(Boolean).join(" ");
  }

  function resolveIcon(button) {
    var key = buttonKey(button);
    for (var i = 0; i < rules.length; i += 1) {
      if (rules[i].re.test(key)) return rules[i];
    }
    return { symbol: "•", tone: "default" };
  }

  function isNativeIconButton(button) {
    return Boolean(button.querySelector("img.icon, img[class*='icon'], svg, .pcs-btn-icon, .cart-btn-icon"));
  }

  function isControlButton(button, label) {
    if (button.matches(".calc-btn, .arcade-key, .game-btn, .mc-icon-btn, .ai-chat-icon-btn, .ai-chat-minibar-btn, .ai-chat-header-icon-btn, .ai-chat-close, .radio-mini-close, .pcs-field-help-button")) {
      return true;
    }
    if (/^[0-9.]$/.test(label)) return true;
    if (/^[+\-*/%=]$/.test(label)) return true;
    if (/^(x|×|✕|▲|▼|◀|▶|←|→|↑|↓|★|☆)$/i.test(label)) return true;
    if (label.indexOf("☰") !== -1) return true;
    return false;
  }

  function shouldDecorate(button) {
    if (!button || button.nodeType !== 1) return false;
    if (button.closest("[data-button-icons='off'], [data-pcs-button-icons='off']")) return false;
    if (button.querySelector(".pcs-btn-icon, .cart-btn-icon")) return false;
    var label = normalize(button.textContent || button.getAttribute("aria-label") || button.getAttribute("title"));
    if (!label) return false;
    if (isControlButton(button, label)) return false;
    if (isNativeIconButton(button)) {
      button.classList.add("pcs-btn-with-native-icon");
      return false;
    }
    return true;
  }

  function decorate(button) {
    if (!shouldDecorate(button)) return;
    var icon = resolveIcon(button);
    var badge = document.createElement("span");
    badge.className = "pcs-btn-icon pcs-btn-icon-" + icon.tone;
    badge.setAttribute("aria-hidden", "true");
    badge.textContent = icon.symbol;
    button.classList.add("pcs-btn-with-icon");
    button.insertBefore(badge, button.firstChild);
  }

  function decorateAll(root) {
    var scope = root && root.querySelectorAll ? root : document;
    if (scope.matches && (scope.matches("button") || scope.matches("a.btn") || scope.matches("[role='button']"))) {
      decorate(scope);
    }
    scope.querySelectorAll("button, a.btn, [role='button']").forEach(decorate);
  }

  var pending = false;
  function schedule(root) {
    if (pending) return;
    pending = true;
    window.requestAnimationFrame(function() {
      pending = false;
      decorateAll(root || document);
    });
  }

  window.PCSDecorateButtonIcons = function(root) {
    schedule(root || document);
  };

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", function() { schedule(document); });
  } else {
    schedule(document);
  }

  new MutationObserver(function(records) {
    for (var i = 0; i < records.length; i += 1) {
      if (records[i].type === "childList" || records[i].type === "characterData") {
        schedule(document);
        break;
      }
    }
  }).observe(document.documentElement, { childList: true, characterData: true, subtree: true });
})();
