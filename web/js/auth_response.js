(function (global) {
  "use strict";

  var SERVICE_UNAVAILABLE = "El servicio no está disponible en este momento. Intenta de nuevo en unos segundos.";

  function safeText(value) {
    var text = String(value || "").trim();
    if (!text || text.length > 500 || /<\/?(?:html|head|title|body|center|hr|script|style)\b/i.test(text)) {
      return "";
    }
    return text;
  }

  function getMessage(response, fallback) {
    if (response && Number(response.status) >= 500) {
      return SERVICE_UNAVAILABLE;
    }
    if (response && response.json) {
      if (response.json.message) {
        return String(response.json.message);
      }
      if (response.json.error) {
        return String(response.json.error);
      }
    }
    var text = response ? safeText(response.text) : "";
    return text || fallback;
  }

  async function read(response) {
    var text = await response.text();
    try {
      return { ok: response.ok, status: response.status, json: JSON.parse(text) };
    } catch (error) {
      return { ok: response.ok, status: response.status, text: text };
    }
  }

  global.PCSAuthResponse = {
    getMessage: getMessage,
    read: read,
    serviceUnavailableMessage: SERVICE_UNAVAILABLE
  };
})(window);
