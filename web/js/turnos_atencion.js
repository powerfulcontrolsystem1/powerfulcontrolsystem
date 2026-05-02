(function () {
  "use strict";

  function q(name) { return (new URLSearchParams(window.location.search)).get(name) || ""; }
  function empresaId() {
    var direct = q("empresa_id") || q("id");
    if (direct) return direct;
    return sessionStorage.getItem("active_empresa_id") || localStorage.getItem("active_empresa_id") || "";
  }
  function esc(v) { return String(v || "").replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;"); }
  async function j(url, opts) {
    var res = await fetch(url, Object.assign({ credentials: "same-origin" }, opts || {}));
    var text = await res.text();
    var data = {};
    try { data = text ? JSON.parse(text) : {}; } catch (e) { data = { raw: text }; }
    if (!res.ok) throw new Error(data.error || data.message || data.raw || ("HTTP " + res.status));
    return data;
  }

  var eid = empresaId();
  var base = "/api/empresa/turnos_atencion?empresa_id=" + encodeURIComponent(eid);
  var els = {
    configForm: document.getElementById("configForm"),
    cfgNombreSistema: document.getElementById("cfgNombreSistema"),
    cfgNombrePantalla: document.getElementById("cfgNombrePantalla"),
    cfgPrefijo: document.getElementById("cfgPrefijo"),
    cfgTiempo: document.getElementById("cfgTiempo"),
    cfgEmisionPublica: document.getElementById("cfgEmisionPublica"),
    cfgMostrarCompletados: document.getElementById("cfgMostrarCompletados"),
    configMsg: document.getElementById("configMsg"),
    serviceForm: document.getElementById("serviceForm"),
    svcCodigo: document.getElementById("svcCodigo"),
    svcNombre: document.getElementById("svcNombre"),
    svcDescripcion: document.getElementById("svcDescripcion"),
    svcPrefijo: document.getElementById("svcPrefijo"),
    svcPrioridad: document.getElementById("svcPrioridad"),
    svcColor: document.getElementById("svcColor"),
    serviceMsg: document.getElementById("serviceMsg"),
    puestoForm: document.getElementById("puestoForm"),
    pstCodigo: document.getElementById("pstCodigo"),
    pstNombre: document.getElementById("pstNombre"),
    pstArea: document.getElementById("pstArea"),
    pstUbicacion: document.getElementById("pstUbicacion"),
    pstServicios: document.getElementById("pstServicios"),
    puestoMsg: document.getElementById("puestoMsg"),
    emitServicio: document.getElementById("emitServicio"),
    emitPuesto: document.getElementById("emitPuesto"),
    emitNombre: document.getElementById("emitNombre"),
    emitDocumento: document.getElementById("emitDocumento"),
    emitMsg: document.getElementById("emitMsg"),
    lastTicketBox: document.getElementById("lastTicketBox"),
    servicesList: document.getElementById("servicesList"),
    puestosList: document.getElementById("puestosList"),
    ticketsList: document.getElementById("ticketsList"),
    recentCallsList: document.getElementById("recentCallsList"),
    openPublicKiosk: document.getElementById("openPublicKiosk"),
    openDisplayScreen: document.getElementById("openDisplayScreen")
  };

  function setMsg(el, text, bad) { if (!el) return; el.textContent = text || ""; el.style.color = bad ? "#ffb4b4" : "#b8d8ff"; }

  async function loadConfig() {
    var cfg = await j(base + "&action=config");
    els.cfgNombreSistema.value = cfg.nombre_sistema || "";
    els.cfgNombrePantalla.value = cfg.nombre_pantalla || "";
    els.cfgPrefijo.value = cfg.prefijo_general || "T";
    els.cfgTiempo.value = cfg.tiempo_llamado_segundos || 20;
    els.cfgEmisionPublica.checked = !!cfg.permitir_emision_publica;
    els.cfgMostrarCompletados.checked = !!cfg.mostrar_tickets_completados;
  }

  function serviceOptionHtml(items) {
    return '<option value="">Seleccione...</option>' + items.map(function (x) {
      return '<option value="' + esc(x.id) + '">' + esc(x.nombre) + " · " + esc(x.prefijo) + "</option>";
    }).join("");
  }
  function puestoOptionHtml(items) {
    return '<option value="">Seleccione...</option>' + items.map(function (x) {
      return '<option value="' + esc(x.id) + '">' + esc(x.nombre) + (x.area ? " · " + esc(x.area) : "") + "</option>";
    }).join("");
  }

  async function loadLists() {
    var services = await j(base + "&action=servicios");
    var puestos = await j(base + "&action=puestos");
    var tickets = await j(base + "&action=tickets");
    var dashboard = await j(base + "&action=dashboard");

    els.emitServicio.innerHTML = serviceOptionHtml(services);
    els.emitPuesto.innerHTML = puestoOptionHtml(puestos);
    els.servicesList.innerHTML = services.length ? services.map(function (x) {
      return '<div class="turnos-list-item"><strong>' + esc(x.nombre) + '</strong><span class="form-help">Código ' + esc(x.codigo) + ' · Prefijo ' + esc(x.prefijo) + ' · Prioridad ' + esc(x.prioridad) + '</span></div>';
    }).join("") : '<div class="turnos-list-item">Sin servicios.</div>';
    els.puestosList.innerHTML = puestos.length ? puestos.map(function (x) {
      return '<div class="turnos-list-item"><strong>' + esc(x.nombre) + '</strong><span class="form-help">' + esc(x.area || "Sin área") + ' · ' + esc(x.ubicacion || "Sin ubicación") + '</span></div>';
    }).join("") : '<div class="turnos-list-item">Sin puestos.</div>';
    els.ticketsList.innerHTML = tickets.length ? tickets.map(function (x) {
      return '<div class="turnos-list-item"><strong><span class="turno-code" style="font-size:1.3rem;">' + esc(x.codigo_turno) + '</span></strong><span>' + esc(x.servicio_nombre) + ' · ' + esc(x.estado) + (x.puesto_nombre ? ' · ' + esc(x.puesto_nombre) : '') + '</span><div class="turnos-actions" style="margin-top:10px;"><button class="btn secondary" data-action="llamar" data-id="' + x.id + '">Re-llamar</button><button class="btn secondary" data-action="atender" data-id="' + x.id + '">Atender</button><button class="btn secondary" data-action="completar" data-id="' + x.id + '">Completar</button><button class="btn danger" data-action="cancelar" data-id="' + x.id + '">Cancelar</button></div></div>';
    }).join("") : '<div class="turnos-list-item">Sin tickets activos.</div>';
    els.recentCallsList.innerHTML = (dashboard.llamados_recientes || []).length ? dashboard.llamados_recientes.map(function (x) {
      return '<div class="turnos-list-item"><strong>' + esc(x.codigo_turno) + '</strong><span>' + esc(x.servicio_nombre) + (x.puesto_nombre ? ' · ' + esc(x.puesto_nombre) : '') + ' · ' + esc(x.estado) + '</span></div>';
    }).join("") : '<div class="turnos-list-item">Todavía no hay llamados.</div>';

    document.getElementById("kpiEsperando").textContent = dashboard.esperando || 0;
    document.getElementById("kpiLlamando").textContent = dashboard.llamando || 0;
    document.getElementById("kpiAtendiendo").textContent = dashboard.en_atencion || 0;
    document.getElementById("kpiCompletados").textContent = dashboard.completados || 0;
    document.getElementById("kpiCancelados").textContent = dashboard.cancelados || 0;
    document.getElementById("kpiEspera").textContent = dashboard.tiempo_espera_prom_min || 0;
    document.getElementById("kpiAtencion").textContent = dashboard.tiempo_atencion_prom_min || 0;
  }

  async function refreshAll() {
    await Promise.all([loadConfig(), loadLists()]);
  }

  async function saveConfig(ev) {
    ev.preventDefault();
    try {
      await j(base + "&action=config", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          nombre_sistema: els.cfgNombreSistema.value.trim(),
          nombre_pantalla: els.cfgNombrePantalla.value.trim(),
          prefijo_general: els.cfgPrefijo.value.trim(),
          tiempo_llamado_segundos: Number(els.cfgTiempo.value || 20),
          permitir_emision_publica: !!els.cfgEmisionPublica.checked,
          mostrar_tickets_completados: !!els.cfgMostrarCompletados.checked
        })
      });
      setMsg(els.configMsg, "Configuración guardada.");
      await refreshAll();
    } catch (e) { setMsg(els.configMsg, e.message, true); }
  }

  async function createServicio(ev) {
    ev.preventDefault();
    try {
      await j(base + "&action=servicios", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          codigo: els.svcCodigo.value.trim(),
          nombre: els.svcNombre.value.trim(),
          descripcion: els.svcDescripcion.value.trim(),
          prefijo: els.svcPrefijo.value.trim(),
          prioridad: Number(els.svcPrioridad.value || 100),
          color: els.svcColor.value
        })
      });
      els.serviceForm.reset();
      els.svcPrioridad.value = 100;
      els.svcColor.value = "#2563eb";
      setMsg(els.serviceMsg, "Servicio creado.");
      await refreshAll();
    } catch (e) { setMsg(els.serviceMsg, e.message, true); }
  }

  async function createPuesto(ev) {
    ev.preventDefault();
    try {
      await j(base + "&action=puestos", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          codigo: els.pstCodigo.value.trim(),
          nombre: els.pstNombre.value.trim(),
          area: els.pstArea.value.trim(),
          ubicacion: els.pstUbicacion.value.trim(),
          servicios_permitidos: els.pstServicios.value.trim()
        })
      });
      els.puestoForm.reset();
      setMsg(els.puestoMsg, "Puesto creado.");
      await refreshAll();
    } catch (e) { setMsg(els.puestoMsg, e.message, true); }
  }

  async function emitirTicket() {
    try {
      var item = await j(base + "&action=emitir_ticket", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          servicio_id: Number(els.emitServicio.value || 0),
          nombre_cliente: els.emitNombre.value.trim(),
          documento_cliente: els.emitDocumento.value.trim(),
          canal_emision: "modulo"
        })
      });
      els.lastTicketBox.innerHTML = '<div class="turnos-list-item"><strong>Ticket emitido</strong><div class="turno-code">' + esc(item.codigo_turno) + '</div><span>' + esc(item.servicio_nombre) + '</span></div>';
      setMsg(els.emitMsg, "Ticket emitido correctamente.");
      await refreshAll();
    } catch (e) { setMsg(els.emitMsg, e.message, true); }
  }

  async function llamarSiguiente() {
    try {
      if (!els.emitPuesto.value) throw new Error("Selecciona un puesto para llamar el siguiente turno.");
      var item = await j(base + "&action=llamar_siguiente", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ puesto_id: Number(els.emitPuesto.value || 0) })
      });
      els.lastTicketBox.innerHTML = '<div class="turnos-list-item"><strong>Llamando ahora</strong><div class="turno-code">' + esc(item.codigo_turno) + '</div><span>' + esc(item.servicio_nombre) + ' · ' + esc(item.puesto_nombre) + '</span></div>';
      setMsg(els.emitMsg, "Se llamó el siguiente turno.");
      await refreshAll();
    } catch (e) { setMsg(els.emitMsg, e.message, true); }
  }

  async function changeTicketState(id, estado) {
    try {
      await j(base + "&action=cambiar_estado", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          ticket_id: Number(id),
          puesto_id: Number(els.emitPuesto.value || 0),
          estado: estado
        })
      });
      await refreshAll();
    } catch (e) { setMsg(els.emitMsg, e.message, true); }
  }

  function wireActions() {
    els.configForm.addEventListener("submit", saveConfig);
    els.serviceForm.addEventListener("submit", createServicio);
    els.puestoForm.addEventListener("submit", createPuesto);
    document.getElementById("btnEmitirTicket").addEventListener("click", emitirTicket);
    document.getElementById("btnLlamarSiguiente").addEventListener("click", llamarSiguiente);
    els.ticketsList.addEventListener("click", function (ev) {
      var target = ev.target;
      if (!target.dataset || !target.dataset.action) return;
      var action = target.dataset.action;
      var id = target.dataset.id;
      if (action === "llamar") changeTicketState(id, "llamando");
      if (action === "atender") changeTicketState(id, "atendiendo");
      if (action === "completar") changeTicketState(id, "completado");
      if (action === "cancelar") changeTicketState(id, "cancelado");
    });
    els.openPublicKiosk.href = "/turnos_publicos.html?empresa_id=" + encodeURIComponent(eid);
    els.openDisplayScreen.href = "/pantalla_turnos.html?empresa_id=" + encodeURIComponent(eid);
  }

  wireActions();
  refreshAll().catch(function (e) { setMsg(els.emitMsg, e.message, true); });
  setInterval(function () { refreshAll().catch(function () {}); }, 10000);
})();
