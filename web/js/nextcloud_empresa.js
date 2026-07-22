(function () {
  "use strict";

  var state = { empresaID: 0, data: null, redirected: false };

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
  function setBusy(value) {
    ["nextcloudProvision", "nextcloudReset", "nextcloudToggle", "nextcloudOpen"].forEach(function (id) {
      var node = byID(id); if (node) node.disabled = !!value;
    });
  }
  function render(data) {
    state.data = data || {};
    byID("nextcloudUser").textContent = state.data.nextcloud_user || "-";
    byID("nextcloudQuota").textContent = state.data.quota_mb ? state.data.quota_mb + " MB" : "-";
    byID("nextcloudProvisioned").textContent = state.data.provisioned ? "Aprovisionado" : "Pendiente";
    byID("nextcloudToggle").textContent = state.data.active ? "Desactivar espacio" : "Activar espacio";
    var status = !state.data.enabled ? "Servicio desactivado por el super administrador."
      : !state.data.configured ? "Falta completar la configuracion global de Nextcloud."
      : !state.data.active ? "Espacio documental desactivado para esta empresa."
      : state.data.provisioned ? "Cuenta lista para usar." : "Cuenta asignada; prepara el espacio para crearla en Nextcloud.";
    byID("nextcloudStatus").textContent = status;
    byID("nextcloudProvision").disabled = !state.data.configured || !state.data.active || !!state.data.provisioned;
    byID("nextcloudReset").disabled = !state.data.configured || !state.data.active || !state.data.provisioned;
    byID("nextcloudOpen").disabled = !state.data.web_url || !state.data.active || !state.data.provisioned;
    byID("nextcloudToggle").disabled = !state.data.configured && !state.data.active;
    if (state.data.temporary_password) {
      byID("nextcloudTemporaryPassword").textContent = state.data.temporary_password;
      byID("nextcloudCredential").classList.add("visible");
    }
  }
  function openCompanyNextcloudWhenReady(data) {
    if (state.redirected || !data || !data.provisioned || !data.active || !data.web_url || data.temporary_password) return;
    state.redirected = true;
    // La pagina se carga dentro del iframe de Administrar empresa. Salir al nivel
    // superior evita que las politicas anti-iframe de Nextcloud bloqueen el acceso.
    if (window.top && window.top !== window) window.top.location.assign(data.web_url);
    else window.location.assign(data.web_url);
  }
  async function run(action) {
    setBusy(true);
    byID("nextcloudStatus").textContent = "Procesando...";
    try {
      var data = await request(action, "POST");
      render(data);
    }
    catch (error) { byID("nextcloudStatus").textContent = error.message; }
    finally { if (state.data) render(state.data); }
  }
  async function load() {
    state.empresaID = resolveEmpresaID();
    if (!state.empresaID) { byID("nextcloudStatus").textContent = "No se pudo resolver la empresa activa. Vuelve a abrir esta opcion desde Administrar empresa."; setBusy(true); return; }
    try {
      var data = await request("", "GET");
      render(data);
      if (data.configured && data.active && !data.provisioned) run("provision");
      else openCompanyNextcloudWhenReady(data);
    }
    catch (error) { byID("nextcloudStatus").textContent = error.message; setBusy(true); }
  }

  byID("nextcloudProvision").addEventListener("click", function () { run("provision"); });
  byID("nextcloudReset").addEventListener("click", function () {
    if (window.confirm("Se invalidara la contraseña actual de Nextcloud. ¿Continuar?")) run("reset_password");
  });
  byID("nextcloudToggle").addEventListener("click", function () {
    var action = state.data && state.data.active ? "deactivate" : "activate";
    run(action);
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
