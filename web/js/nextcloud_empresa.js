(function () {
  "use strict";

  var state = { empresaID: 0, data: null };

  function byID(id) { return document.getElementById(id); }
  function setStatus(message) {
    var node = byID("nextcloudStatus");
    if (node) node.textContent = String(message || "");
  }
  function cookie(name) {
    var match = String(document.cookie || "").match(new RegExp("(?:^|;\\s*)" + name + "=([^;]*)"));
    return match ? decodeURIComponent(match[1]) : "";
  }
  function resolveEmpresaID() {
    var value = new URLSearchParams(window.location.search).get("empresa_id");
    if (!value && window.parent && typeof window.parent.__resolveEmpresaIdContext === "function") {
      value = window.parent.__resolveEmpresaIdContext();
    }
    if (!value && window.parent && window.parent.__empresaModuleGuard && typeof window.parent.__empresaModuleGuard.resolveEmpresaId === "function") {
      value = window.parent.__empresaModuleGuard.resolveEmpresaId();
    }
    var parsed = Number(value || 0);
    return Number.isInteger(parsed) && parsed > 0 ? parsed : 0;
  }
  async function request(action, method) {
    var endpoint = "/api/empresa/nextcloud?empresa_id=" + encodeURIComponent(state.empresaID);
    if (action) endpoint += "&action=" + encodeURIComponent(action);
    var headers = {};
    if (method !== "GET") headers["X-CSRF-Token"] = cookie("pcs_csrf");
    var response = await fetch(endpoint, { method: method, credentials: "same-origin", headers: headers });
    var text = await response.text();
    var data = {};
    try { data = text ? JSON.parse(text) : {}; } catch (error) { data = {}; }
    if (!response.ok) {
      if (response.status === 401 || response.status === 403) throw new Error("No tienes permiso para administrar el espacio documental de esta empresa.");
      throw new Error("No se pudo completar la operacion con Nextcloud.");
    }
    return data;
  }
  function render(data) {
    state.data = data || {};
  }
  function embedNextcloud(url) {
    if (!url) return;
    var frame = byID("nextcloudFrame");
    if (!frame) return;
    setStatus("Iniciando sesion segura.");
    if (frame.getAttribute("src") !== url) frame.setAttribute("src", url);
  }
  function openCompanyNextcloudWhenReady(data) {
    if (!data || !data.provisioned || !data.active || !data.web_url) return;
    if (data.autologin_url) {
      embedNextcloud(data.autologin_url);
      return;
    }
    setStatus(data.autologin_error || "Inicio automatico no disponible.");
  }
  async function run(action) {
    setStatus("Preparando el espacio de Nextcloud.");
    try {
      var data = await request(action, "POST");
      render(data);
      if (action === "provision") openCompanyNextcloudWhenReady(data);
    }
    catch (error) { setStatus(error.message); }
  }
  async function load() {
    state.empresaID = resolveEmpresaID();
    if (!state.empresaID) { setStatus("No se pudo resolver la empresa activa."); return; }
    try {
      var data = await request("", "GET");
      render(data);
      if (data.configured && data.active && !data.provisioned) run("provision");
      else openCompanyNextcloudWhenReady(data);
    }
    catch (error) { setStatus(error.message); }
  }

  byID("nextcloudFrame").addEventListener("load", function () {
    setStatus("Nextcloud abierto para esta empresa.");
  });
  load();
}());
