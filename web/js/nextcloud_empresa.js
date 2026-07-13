(function () {
  "use strict";

  var state = { empresaID: 0, data: null };

  function byID(id) { return document.getElementById(id); }
  function cookie(name) {
    var match = String(document.cookie || "").match(new RegExp("(?:^|;\\s*)" + name + "=([^;]*)"));
    return match ? decodeURIComponent(match[1]) : "";
  }
  function resolveEmpresaID() {
    var value = new URLSearchParams(window.location.search).get("empresa_id");
    if (!value && window.parent && typeof window.parent.__resolveEmpresaIdContext === "function") {
      value = window.parent.__resolveEmpresaIdContext();
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
    if (!response.ok) throw new Error((data && data.error) || text || "No se pudo completar la operacion.");
    return data;
  }
  function setBusy(value) {
    ["nextcloudProvision", "nextcloudReset", "nextcloudOpen"].forEach(function (id) {
      var node = byID(id); if (node) node.disabled = !!value;
    });
  }
  function render(data) {
    state.data = data || {};
    byID("nextcloudUser").textContent = state.data.nextcloud_user || "-";
    byID("nextcloudQuota").textContent = state.data.quota_mb ? state.data.quota_mb + " MB" : "-";
    byID("nextcloudProvisioned").textContent = state.data.provisioned ? "Aprovisionado" : "Pendiente";
    var status = !state.data.enabled ? "Servicio desactivado por el super administrador."
      : !state.data.configured ? "Falta completar la configuracion global de Nextcloud."
      : state.data.provisioned ? "Cuenta lista para usar." : "Cuenta asignada; prepara el espacio para crearla en Nextcloud.";
    byID("nextcloudStatus").textContent = status;
    byID("nextcloudProvision").disabled = !state.data.configured || !!state.data.provisioned;
    byID("nextcloudReset").disabled = !state.data.configured || !state.data.provisioned;
    byID("nextcloudOpen").disabled = !state.data.web_url || !state.data.provisioned;
    if (state.data.temporary_password) {
      byID("nextcloudTemporaryPassword").textContent = state.data.temporary_password;
      byID("nextcloudCredential").classList.add("visible");
    }
  }
  async function run(action) {
    setBusy(true);
    byID("nextcloudStatus").textContent = "Procesando...";
    try { render(await request(action, "POST")); }
    catch (error) { byID("nextcloudStatus").textContent = error.message; }
    finally { if (state.data) render(state.data); }
  }
  async function load() {
    state.empresaID = resolveEmpresaID();
    if (!state.empresaID) { byID("nextcloudStatus").textContent = "No se pudo resolver la empresa activa."; setBusy(true); return; }
    try { render(await request("", "GET")); }
    catch (error) { byID("nextcloudStatus").textContent = error.message; setBusy(true); }
  }

  byID("nextcloudProvision").addEventListener("click", function () { run("provision"); });
  byID("nextcloudReset").addEventListener("click", function () {
    if (window.confirm("Se invalidara la contraseña actual de Nextcloud. ¿Continuar?")) run("reset_password");
  });
  byID("nextcloudOpen").addEventListener("click", function () {
    if (state.data && state.data.web_url) window.open(state.data.web_url, "_blank", "noopener,noreferrer");
  });
  byID("nextcloudCopy").addEventListener("click", function () {
    var value = byID("nextcloudTemporaryPassword").textContent;
    if (value && navigator.clipboard) navigator.clipboard.writeText(value);
  });
  load();
}());
