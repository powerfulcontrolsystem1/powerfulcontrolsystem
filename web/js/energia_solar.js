(function () {
  "use strict";

  var state = {
    empresaId: "",
    catalogo: { proveedores: [], baterias: [], alertas: [] },
    sistemas: [],
    selectedSistemaId: 0
  };

  function byId(id) { return document.getElementById(id); }
  function esc(value) {
    return String(value == null ? "" : value).replace(/[&<>"']/g, function (ch) {
      return {"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#39;"}[ch];
    });
  }
  function moneyNum(value) {
    var n = Number(value || 0);
    return Number.isFinite(n) ? n : 0;
  }
  function fmtW(value) {
    var n = moneyNum(value);
    if (Math.abs(n) >= 1000) return (n / 1000).toFixed(2) + " kW";
    return n.toFixed(0) + " W";
  }
  function fmtPct(value) {
    return moneyNum(value).toFixed(1) + "%";
  }
  function resolveEmpresaId() {
    var params = new URLSearchParams(window.location.search || "");
    var id = params.get("empresa_id") || params.get("id") || "";
    if (!id && window.__resolveEmpresaIdContext) {
      try { id = window.__resolveEmpresaIdContext() || ""; } catch (e) { id = ""; }
    }
    if (!id) {
      ["active_empresa_id", "empresa_id", "admin_empresa_id"].some(function (key) {
        try { id = sessionStorage.getItem(key) || localStorage.getItem(key) || ""; } catch (e) { id = ""; }
        return !!id;
      });
    }
    state.empresaId = String(id || "").replace(/\D+/g, "");
    return state.empresaId;
  }
  function api(action, options) {
    var url = "/api/empresa/energia_solar?empresa_id=" + encodeURIComponent(state.empresaId);
    if (action) {
      var actionParts = String(action).split("&");
      url += "&action=" + encodeURIComponent(actionParts.shift() || "");
      if (actionParts.length) url += "&" + actionParts.join("&");
    }
    return fetch(url, options || {}).then(async function (res) {
      var text = await res.text();
      var data = {};
      try { data = text ? JSON.parse(text) : {}; } catch (e) { data = { ok: false, error: text }; }
      if (!res.ok) throw new Error(data.error || text || "Error " + res.status);
      return data;
    });
  }
  function post(action, payload) {
    return api(action, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload || {})
    });
  }
  function setStatus(message, isError) {
    var heading = document.querySelector(".page-heading");
    var old = byId("solarStatus");
    if (!old && heading) {
      old = document.createElement("p");
      old.id = "solarStatus";
      old.className = "solar-status";
      heading.appendChild(old);
    }
    if (!old) return;
    old.textContent = message || "";
    old.classList.toggle("is-error", !!isError);
  }
  function fillSelect(el, options, getValue, getLabel) {
    if (!el) return;
    el.innerHTML = options.map(function (item) {
      return '<option value="' + esc(getValue(item)) + '">' + esc(getLabel(item)) + '</option>';
    }).join("");
  }
  function renderCatalog() {
    fillSelect(byId("solarProveedor"), state.catalogo.proveedores, function (p) { return p.proveedor; }, function (p) { return p.nombre + " - " + p.plataforma; });
    fillSelect(byId("solarAlertaTipo"), state.catalogo.alertas, function (a) { return a.tipo; }, function (a) { return a.nombre; });
    var batteryOptions = [{ marca: "", modelo: "Seleccionar" }].concat(state.catalogo.baterias || []);
    fillSelect(byId("solarBateriaMarca"), batteryOptions, function (b) { return b.marca || ""; }, function (b) { return b.marca ? b.marca + " - " + b.modelo : "Seleccionar batería"; });
    var grid = byId("solarProviderGrid");
    if (grid) {
      grid.innerHTML = (state.catalogo.proveedores || []).map(function (p) {
        return '<article class="solar-provider-card">' +
          '<strong>' + esc(p.nombre) + '</strong>' +
          '<span>' + esc(p.plataforma) + '</span>' +
          '<p>' + esc(p.nota || "") + '</p>' +
          '<small>Modelos: ' + esc((p.modelos || []).slice(0, 4).join(", ")) + '</small>' +
          '</article>';
      }).join("");
    }
  }
  function updateSystemSelects() {
    var options = state.sistemas.map(function (s) { return { id: s.id, label: s.nombre + " · " + s.proveedor }; });
    ["solarAlertaSistema", "solarLecturaSistema"].forEach(function (id) {
      fillSelect(byId(id), options, function (s) { return s.id; }, function (s) { return s.label; });
    });
    if (state.selectedSistemaId) {
      ["solarAlertaSistema", "solarLecturaSistema"].forEach(function (id) {
        var el = byId(id);
        if (el) el.value = String(state.selectedSistemaId);
      });
    }
  }
  function renderSystems() {
    var list = byId("solarSystemsList");
    if (!list) return;
    if (!state.sistemas.length) {
      list.innerHTML = '<p class="empty-state">Aún no hay sistemas solares configurados.</p>';
      return;
    }
    list.innerHTML = state.sistemas.map(function (s) {
      return '<article class="solar-system-row" data-id="' + esc(s.id) + '">' +
        '<div><strong>' + esc(s.nombre) + '</strong><span>' + esc(s.proveedor + " · " + (s.modelo || "sin modelo")) + '</span>' +
        '<small>' + esc((s.bateria_marca || "Batería") + " " + (s.bateria_modelo || "") + " · " + (s.capacidad_bateria_kwh || 0) + " kWh") + '</small></div>' +
        '<button type="button" class="btn small secondary" data-edit="' + esc(s.id) + '">Editar</button>' +
        '</article>';
    }).join("");
    Array.prototype.forEach.call(list.querySelectorAll("[data-edit]"), function (btn) {
      btn.addEventListener("click", function () { editSystem(Number(btn.getAttribute("data-edit") || 0)); });
    });
  }
  function editSystem(id) {
    var s = state.sistemas.find(function (item) { return Number(item.id) === Number(id); });
    if (!s) return;
    state.selectedSistemaId = Number(s.id);
    byId("solarSistemaId").value = s.id || "";
    byId("solarNombre").value = s.nombre || "";
    byId("solarProveedor").value = s.proveedor || "";
    byId("solarModelo").value = s.modelo || "";
    byId("solarUbicacion").value = s.ubicacion || "";
    byId("solarCapacidad").value = s.capacidad_kwp || 0;
    byId("solarBateriaMarca").value = s.bateria_marca || "";
    byId("solarBateriaModelo").value = s.bateria_modelo || "";
    byId("solarBateriaSerial").value = s.bateria_serial || "";
    byId("solarBms").value = s.bms_protocolo || "";
    byId("solarCapacidadBateria").value = s.capacidad_bateria_kwh || 0;
    byId("solarInstalacion").value = s.instalacion_ref || "";
    byId("solarApiBase").value = s.api_base_url || "";
    byId("solarApiKeyRef").value = s.api_key_ref || "";
    byId("solarGateway").value = s.local_gateway_url || "";
    byId("solarIntervalo").value = s.intervalo_segundos || 300;
    byId("solarEmails").value = s.email_alertas || "";
    byId("solarEmailActivo").checked = !!s.alertas_email_activas;
    byId("solarActivo").checked = !!s.activo;
    updateSystemSelects();
    loadAlerts();
  }
  function renderDashboard(payload) {
    var k = payload.kpis || {};
    byId("solarKpiSistemas").textContent = String(k.sistemas_activos || 0);
    byId("solarKpiProduccion").textContent = fmtW(k.potencia_solar_w || 0);
    byId("solarKpiBateria").textContent = fmtPct(k.bateria_soc_promedio || 0);
    byId("solarKpiAlertas").textContent = String(k.alertas_activas_reciente || 0);
    state.sistemas = payload.sistemas || [];
    if (!state.selectedSistemaId && state.sistemas.length) state.selectedSistemaId = Number(state.sistemas[0].id || 0);
    renderSystems();
    updateSystemSelects();
    renderEventos(payload.eventos || []);
    renderLecturas(payload.lecturas || []);
  }
  function renderEventos(items) {
    var wrap = byId("solarEventos");
    if (!wrap) return;
    wrap.innerHTML = '<h3>Eventos</h3>' + table(["Fecha", "Tipo", "Severidad", "Mensaje", "Email"], items.map(function (e) {
      return [e.fecha_creacion || "", e.tipo || "", e.severidad || "", e.mensaje || "", e.email_enviado ? "Enviado" : (e.email_error || "No")];
    }));
  }
  function renderLecturas(items) {
    var wrap = byId("solarLecturas");
    if (!wrap) return;
    wrap.innerHTML = '<h3>Lecturas</h3>' + table(["Fecha", "Solar", "SOC", "SOH", "Temp.", "BMS"], items.map(function (l) {
      return [l.fecha_lectura || "", fmtW(l.potencia_solar_w), fmtPct(l.bateria_soc_pct), fmtPct(l.bateria_soh_pct), moneyNum(l.temperatura_c).toFixed(1) + " C", l.estado_bateria || "normal"];
    }));
  }
  function table(headers, rows) {
    if (!rows.length) return '<p class="empty-state">Sin datos registrados.</p>';
    return '<table class="solar-table"><thead><tr>' + headers.map(function (h) { return '<th>' + esc(h) + '</th>'; }).join("") +
      '</tr></thead><tbody>' + rows.map(function (row) {
        return '<tr>' + row.map(function (cell) { return '<td>' + esc(cell) + '</td>'; }).join("") + '</tr>';
      }).join("") + '</tbody></table>';
  }
  async function loadDashboard() {
    if (!state.empresaId) {
      setStatus("No se pudo detectar la empresa activa.", true);
      return;
    }
    var data = await api("dashboard");
    renderDashboard(data);
  }
  async function loadAlerts() {
    var systemId = Number((byId("solarAlertaSistema") || {}).value || state.selectedSistemaId || 0);
    if (!systemId) {
      byId("solarAlertsList").innerHTML = '<p class="empty-state">Guarda un sistema para configurar alertas.</p>';
      return;
    }
    var data = await api("alertas&sistema_id=" + encodeURIComponent(systemId));
    var items = data.items || [];
    byId("solarAlertsList").innerHTML = table(["Tipo", "Nombre", "Condición", "Severidad", "Email"], items.map(function (a) {
      return [a.tipo || "", a.nombre || "", (a.operador || "") + " " + (a.umbral || 0), a.severidad || "", a.enviar_email ? "Sí" : "No"];
    }));
  }
  function collectSystem() {
    return {
      id: Number(byId("solarSistemaId").value || 0),
      nombre: byId("solarNombre").value,
      proveedor: byId("solarProveedor").value,
      modelo: byId("solarModelo").value,
      ubicacion: byId("solarUbicacion").value,
      capacidad_kwp: moneyNum(byId("solarCapacidad").value),
      bateria_marca: byId("solarBateriaMarca").value,
      bateria_modelo: byId("solarBateriaModelo").value,
      bateria_serial: byId("solarBateriaSerial").value,
      bms_protocolo: byId("solarBms").value,
      capacidad_bateria_kwh: moneyNum(byId("solarCapacidadBateria").value),
      instalacion_ref: byId("solarInstalacion").value,
      api_base_url: byId("solarApiBase").value,
      api_key_ref: byId("solarApiKeyRef").value,
      local_gateway_url: byId("solarGateway").value,
      intervalo_segundos: Number(byId("solarIntervalo").value || 300),
      email_alertas: byId("solarEmails").value,
      alertas_email_activas: byId("solarEmailActivo").checked,
      activo: byId("solarActivo").checked,
      estado: byId("solarActivo").checked ? "activo" : "inactivo"
    };
  }
  function bindEvents() {
    byId("btnSolarRefresh").addEventListener("click", function () { init().catch(showError); });
    byId("btnSolarClear").addEventListener("click", function () { byId("solarSystemForm").reset(); byId("solarSistemaId").value = ""; byId("solarEmailActivo").checked = true; byId("solarActivo").checked = true; });
    byId("solarAlertaSistema").addEventListener("change", function () { state.selectedSistemaId = Number(this.value || 0); loadAlerts().catch(showError); });
    byId("solarSystemForm").addEventListener("submit", async function (ev) {
      ev.preventDefault();
      var saved = await post("sistema", { sistema: collectSystem() });
      state.selectedSistemaId = Number(saved.id || state.selectedSistemaId || 0);
      setStatus("Sistema solar guardado.", false);
      await loadDashboard();
      await loadAlerts();
    });
    byId("solarAlertForm").addEventListener("submit", async function (ev) {
      ev.preventDefault();
      await post("alerta", { alerta: {
        id: Number(byId("solarAlertaId").value || 0),
        sistema_id: Number(byId("solarAlertaSistema").value || state.selectedSistemaId || 0),
        tipo: byId("solarAlertaTipo").value,
        nombre: byId("solarAlertaNombre").value || byId("solarAlertaTipo").value,
        operador: byId("solarAlertaOperador").value,
        umbral: moneyNum(byId("solarAlertaUmbral").value),
        severidad: byId("solarAlertaSeveridad").value,
        enviar_email: byId("solarAlertaEmail").checked,
        activo: byId("solarAlertaActiva").checked,
        estado: byId("solarAlertaActiva").checked ? "activo" : "inactivo"
      }});
      setStatus("Alerta solar guardada.", false);
      await loadAlerts();
    });
    byId("solarReadingForm").addEventListener("submit", async function (ev) {
      ev.preventDefault();
      var sistemaId = Number(byId("solarLecturaSistema").value || state.selectedSistemaId || 0);
      var result = await post("lectura", { lectura: {
        sistema_id: sistemaId,
        potencia_solar_w: moneyNum(byId("solarLecturaProduccion").value),
        bateria_soc_pct: moneyNum(byId("solarLecturaSoc").value),
        bateria_soh_pct: moneyNum(byId("solarLecturaSoh").value),
        bateria_carga_w: moneyNum(byId("solarLecturaCarga").value),
        bateria_descarga_w: moneyNum(byId("solarLecturaDescarga").value),
        temperatura_c: moneyNum(byId("solarLecturaTemp").value),
        celda_voltaje_min_v: moneyNum(byId("solarLecturaCeldaMin").value),
        celda_voltaje_max_v: moneyNum(byId("solarLecturaCeldaMax").value),
        estado_bateria: byId("solarLecturaEstadoBateria").value || "normal",
        estado_inversor: byId("solarLecturaEstadoInversor").value || "normal",
        raw: { origen: "prueba_visual", modulo: "energia_solar" }
      }});
      setStatus("Lectura registrada. Alertas generadas: " + ((result.eventos || []).length), false);
      await loadDashboard();
      await loadAlerts();
    });
    byId("btnSolarTestEmail").addEventListener("click", async function () {
      var sistemaId = Number(state.selectedSistemaId || (state.sistemas[0] && state.sistemas[0].id) || 0);
      if (!sistemaId) throw new Error("Guarda un sistema antes de probar alertas.");
      var result = await api("probar_alerta&sistema_id=" + encodeURIComponent(sistemaId), { method: "POST" });
      setStatus(result.email_enviado ? "Correo de prueba registrado/enviado." : "Prueba registrada sin correo: " + (result.email_error || ""), !result.email_enviado);
      await loadDashboard();
    });
    byId("solarProveedor").addEventListener("change", function () {
      var item = (state.catalogo.proveedores || []).find(function (p) { return p.proveedor === byId("solarProveedor").value; });
      if (item && item.api_base_sugerida && !byId("solarApiBase").value) byId("solarApiBase").value = item.api_base_sugerida;
    });
    byId("solarAlertaTipo").addEventListener("change", function () {
      var item = (state.catalogo.alertas || []).find(function (a) { return a.tipo === byId("solarAlertaTipo").value; });
      if (!item) return;
      byId("solarAlertaNombre").value = item.nombre || item.tipo;
      byId("solarAlertaOperador").value = item.operador || "<";
      byId("solarAlertaUmbral").value = item.umbral || 0;
    });
  }
  function showError(err) {
    console.error(err);
    setStatus(err && err.message ? err.message : "Ocurrió un problema en energía solar.", true);
  }
  async function init() {
    resolveEmpresaId();
    var catalog = await api("catalogo");
    state.catalogo = catalog || state.catalogo;
    renderCatalog();
    bindOnce();
    await loadDashboard();
    await loadAlerts();
    setStatus("Módulo listo.", false);
  }
  var eventsBound = false;
  function bindOnce() {
    if (eventsBound) return;
    eventsBound = true;
    bindEvents();
  }

  document.addEventListener("DOMContentLoaded", function () {
    init().catch(showError);
  });
})();
