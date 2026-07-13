(function () {
  "use strict";
  function byID(id) { return document.getElementById(id); }
  function cookie(name) { var m = String(document.cookie || "").match(new RegExp("(?:^|;\\s*)" + name + "=([^;]*)")); return m ? decodeURIComponent(m[1]) : ""; }
  async function request(path, options) {
    options = options || {};
    options.credentials = "same-origin";
    options.headers = new Headers(options.headers || {});
    if (String(options.method || "GET").toUpperCase() !== "GET") options.headers.set("X-CSRF-Token", cookie("pcs_csrf"));
    var response = await fetch(path, options);
    var text = await response.text();
    var data = {};
    try { data = text ? JSON.parse(text) : {}; } catch (error) { data = {}; }
    if (!response.ok) throw new Error((data && data.error) || text || "No se pudo completar la operacion.");
    return data;
  }
  function busy(value) { byID("nextcloudSave").disabled = value; byID("nextcloudTest").disabled = value; }
  function message(value) { byID("nextcloudMessage").textContent = value || ""; }
  async function load() {
    try {
      var data = await request("/super/api/config/nextcloud", { method: "GET" });
      byID("nextcloudEnabled").checked = !!data.enabled;
      byID("nextcloudBaseURL").value = data.base_url || "";
      byID("nextcloudAdminUser").value = data.admin_user || "";
      byID("nextcloudQuota").value = data.default_quota_mb || 1024;
      message(data.admin_secret_set ? "Credencial administrativa cifrada y configurada." : "Falta la credencial administrativa.");
    } catch (error) { message(error.message); }
  }
  byID("nextcloudSave").addEventListener("click", async function () {
    busy(true); message("Guardando...");
    try {
      await request("/super/api/config/nextcloud", {
        method: "PUT", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          enabled: byID("nextcloudEnabled").checked,
          base_url: byID("nextcloudBaseURL").value,
          admin_user: byID("nextcloudAdminUser").value,
          admin_secret: byID("nextcloudAdminSecret").value,
          default_quota_mb: Number(byID("nextcloudQuota").value || 0)
        })
      });
      byID("nextcloudAdminSecret").value = ""; message("Configuracion guardada.");
    } catch (error) { message(error.message); } finally { busy(false); }
  });
  byID("nextcloudTest").addEventListener("click", async function () {
    busy(true); message("Probando conexion...");
    try { await request("/super/api/config/nextcloud?action=test", { method: "POST" }); message("Conexion OCS verificada."); }
    catch (error) { message(error.message); } finally { busy(false); }
  });
  load();
}());
