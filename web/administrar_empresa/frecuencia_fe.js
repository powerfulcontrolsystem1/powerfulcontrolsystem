const state = { empresaID: 0 };

    function txt(v) { return (v || "").toString().trim(); }

    function showMessage(text, isError) {
      const el = document.getElementById("msg");
      if (!el) return;
      el.textContent = text || "";
      el.style.color = isError ? "#ef5350" : "";
    }

    function parsePositiveInt(raw) {
      const n = Number(String(raw || "").trim());
      if (!Number.isFinite(n)) return 0;
      const t = Math.trunc(n);
      return t > 0 ? t : 0;
    }

    function getEmpresaID() {
      const params = new URLSearchParams(window.location.search || "");
      let id = parsePositiveInt(params.get("empresa_id"));
      if (id > 0) return id;
      try {
        id = parsePositiveInt(window.sessionStorage.getItem("active_empresa_id"));
        if (id > 0) return id;
      } catch (_) {}
      try {
        id = parsePositiveInt(window.localStorage.getItem("active_empresa_id"));
        if (id > 0) return id;
      } catch (_) {}
      return 0;
    }

    function ensureOk(res) {
      if (res.ok) return Promise.resolve(res);
      return res.text().then(function(t) { throw new Error(t || ("HTTP " + res.status)); });
    }

    async function loadConfig() {
      const res = await fetch("/api/empresa/configuracion_avanzada?empresa_id=" + encodeURIComponent(String(state.empresaID)));
      await ensureOk(res);
      const cfg = await res.json();
      document.getElementById("freqEnabled").checked = cfg && cfg.facturacion_frecuencia_automatica_activa === true;
      document.getElementById("freqCadaNNo").value = Number(cfg && cfg.facturacion_frecuencia_cada_n_no || 0);
      document.getElementById("freqContador").value = Number(cfg && cfg.facturacion_frecuencia_contador || 0);
    }

    async function saveConfig() {
      const enabled = !!document.getElementById("freqEnabled").checked;
      const n = Math.max(0, Math.min(1000, Math.trunc(Number(document.getElementById("freqCadaNNo").value || 0))));
      const payload = {
        empresa_id: state.empresaID,
        facturacion_frecuencia_automatica_activa: enabled,
        facturacion_frecuencia_cada_n_no: n
      };
      const res = await fetch("/api/empresa/configuracion_avanzada?empresa_id=" + encodeURIComponent(String(state.empresaID)), {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });
      await ensureOk(res);
      const out = await res.json();
      if (out && out.configuracion) {
        document.getElementById("freqContador").value = Number(out.configuracion.facturacion_frecuencia_contador || 0);
      }
    }

    document.getElementById("freqForm").addEventListener("submit", async function(ev) {
      ev.preventDefault();
      try {
        await saveConfig();
        showMessage("Frecuencia guardada correctamente.", false);
      } catch (err) {
        showMessage(err && err.message ? err.message : "No se pudo guardar la frecuencia.", true);
      }
    });

    document.getElementById("btnReload").addEventListener("click", async function() {
      try {
        await loadConfig();
        showMessage("Configuración recargada.", false);
      } catch (err) {
        showMessage(err && err.message ? err.message : "No se pudo recargar.", true);
      }
    });

    (async function init() {
      state.empresaID = getEmpresaID();
      if (!state.empresaID) {
        showMessage("empresa_id requerido para cargar la configuración.", true);
        document.getElementById("freqForm").style.display = "none";
        return;
      }
      try {
        await loadConfig();
      } catch (err) {
        showMessage(err && err.message ? err.message : "No se pudo cargar la configuración.", true);
      }
    })();
