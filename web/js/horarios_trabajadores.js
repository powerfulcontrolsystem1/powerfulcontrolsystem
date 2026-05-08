(function () {
  "use strict";

  function getQueryParam(name) {
    var params = new URLSearchParams(window.location.search || "");
    return (params.get(name) || "").trim();
  }

  function parsePositiveInt(raw) {
    var n = Number(String(raw || "").trim());
    if (!Number.isFinite(n)) return 0;
    n = Math.trunc(n);
    return n > 0 ? n : 0;
  }

  function getEmpresaId() {
    var direct = parsePositiveInt(getQueryParam("empresa_id") || getQueryParam("id"));
    if (direct > 0) return String(direct);
    var keys = ["active_empresa_id", "empresa_id", "admin_empresa_id"];
    var stores = [window.sessionStorage, window.localStorage];
    for (var s = 0; s < stores.length; s += 1) {
      var store = stores[s];
      if (!store) continue;
      for (var i = 0; i < keys.length; i += 1) {
        var val = parsePositiveInt(store.getItem(keys[i]) || "");
        if (val > 0) return String(val);
      }
    }
    return "";
  }

  function isoDate(date) {
    var y = date.getFullYear();
    var m = String(date.getMonth() + 1).padStart(2, "0");
    var d = String(date.getDate()).padStart(2, "0");
    return y + "-" + m + "-" + d;
  }

  function addDays(date, days) {
    var copy = new Date(date.getTime());
    copy.setDate(copy.getDate() + days);
    return copy;
  }

  function escapeHtml(value) {
    return String(value || "")
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;")
      .replace(/'/g, "&#39;");
  }

  var empresaId = getEmpresaId();
  var endpoint = "/api/empresa/horarios_trabajadores?empresa_id=" + encodeURIComponent(empresaId);
  var editingId = 0;
  var currentRows = [];
  var currentDashboard = null;

  var els = {
    filtroDesde: document.getElementById("filtro_desde"),
    filtroHasta: document.getElementById("filtro_hasta"),
    filtroQ: document.getElementById("filtro_q"),
    filtroArea: document.getElementById("filtro_area"),
    filtroSede: document.getElementById("filtro_sede"),
    filtroEstado: document.getElementById("filtro_estado"),
    filtroPublicados: document.getElementById("filtro_publicados"),
    btnRefrescar: document.getElementById("btnRefrescar"),
    btnPublicarRango: document.getElementById("btnPublicarRango"),
    tablaBody: document.getElementById("tablaTurnosBody"),
    semaforosBox: document.getElementById("semaforosBox"),
    alertasBox: document.getElementById("alertasBox"),
    oportunidadesBox: document.getElementById("oportunidadesBox"),
    areasBox: document.getElementById("areasBox"),
    sedesBox: document.getElementById("sedesBox"),
    turnoMsg: document.getElementById("turnoMsg"),
    configMsg: document.getElementById("configMsg"),
    formTurno: document.getElementById("formTurno"),
    turnoId: document.getElementById("turno_id"),
    usuarioId: document.getElementById("usuario_id"),
    nombreEmpleado: document.getElementById("nombre_empleado"),
    cargo: document.getElementById("cargo"),
    area: document.getElementById("area"),
    sede: document.getElementById("sede"),
    fechaInicio: document.getElementById("fecha_inicio"),
    fechaFin: document.getElementById("fecha_fin"),
    horaInicio: document.getElementById("hora_inicio"),
    horaFin: document.getElementById("hora_fin"),
    descansoMinutos: document.getElementById("descanso_minutos"),
    turnoNombre: document.getElementById("turno_nombre"),
    tipoTurno: document.getElementById("tipo_turno"),
    canal: document.getElementById("canal"),
    estadoTurno: document.getElementById("estado_turno"),
    colorTurno: document.getElementById("color_turno"),
    observaciones: document.getElementById("observaciones"),
    publicado: document.getElementById("publicado"),
    requiereCobertura: document.getElementById("requiere_cobertura"),
    btnLimpiarTurno: document.getElementById("btnLimpiarTurno"),
    cfgHorasDia: document.getElementById("cfg_horas_dia"),
    cfgHorasSemana: document.getElementById("cfg_horas_semana"),
    cfgDescanso: document.getElementById("cfg_descanso"),
    cfgAnticipacion: document.getElementById("cfg_anticipacion"),
    cfgPermitirSolapados: document.getElementById("cfg_permitir_solapados"),
    btnGuardarConfig: document.getElementById("btnGuardarConfig"),
    kpiTurnos: document.getElementById("kpiTurnos"),
    kpiEmpleados: document.getElementById("kpiEmpleados"),
    kpiHoras: document.getElementById("kpiHoras"),
    kpiPublicados: document.getElementById("kpiPublicados"),
    kpiPendientes: document.getElementById("kpiPendientes"),
    kpiConflictos: document.getElementById("kpiConflictos"),
    kpiCobertura: document.getElementById("kpiCobertura"),
    kpiPromedio: document.getElementById("kpiPromedio")
  };

  function weekdayValues() {
    return Array.prototype.slice.call(document.querySelectorAll(".weekday-check:checked"))
      .map(function (el) { return Number(el.value); })
      .filter(function (value) { return Number.isFinite(value); });
  }

  function setMessage(el, text, isError) {
    if (!el) return;
    el.textContent = text || "";
    el.style.color = isError ? "#ffb4b4" : "#b9d8ff";
  }

  function renderMetric(el, value) {
    if (el) el.textContent = String(value == null ? 0 : value);
  }

  function renderSimpleList(container, items, formatter, emptyText) {
    if (!container) return;
    if (!items || !items.length) {
      container.innerHTML = '<div class="horarios-list-item">' + escapeHtml(emptyText) + "</div>";
      return;
    }
    container.innerHTML = items.map(function (item) {
      return '<div class="horarios-list-item">' + formatter(item) + "</div>";
    }).join("");
  }

  async function fetchJSON(url, options) {
    var res = await fetch(url, Object.assign({ credentials: "same-origin" }, options || {}));
    var text = await res.text();
    var data = {};
    if (text) {
      try { data = JSON.parse(text); } catch (e) { data = { raw: text }; }
    }
    if (!res.ok) {
      var message = (data && (data.error || data.message || data.raw)) || ("HTTP " + res.status);
      throw new Error(message);
    }
    return data;
  }

  function applyConfigToForm(cfg) {
    els.cfgHorasDia.value = cfg.horas_objetivo_dia || 8;
    els.cfgHorasSemana.value = cfg.horas_objetivo_semana || 48;
    els.cfgDescanso.value = cfg.descanso_minimo_minutos || 30;
    els.cfgAnticipacion.value = cfg.anticipacion_publicacion_horas || 24;
    els.cfgPermitirSolapados.checked = !!cfg.permitir_solapados;
  }

  async function loadConfig() {
    var cfg = await fetchJSON(endpoint + "&action=config");
    applyConfigToForm(cfg || {});
    return cfg;
  }

  function renderDashboard(data) {
    currentDashboard = data || {};
    renderMetric(els.kpiTurnos, currentDashboard.total_turnos || 0);
    renderMetric(els.kpiEmpleados, currentDashboard.empleados_programados || 0);
    renderMetric(els.kpiHoras, currentDashboard.horas_programadas || 0);
    renderMetric(els.kpiPublicados, currentDashboard.turnos_publicados || 0);
    renderMetric(els.kpiPendientes, currentDashboard.turnos_pendientes || 0);
    renderMetric(els.kpiConflictos, currentDashboard.conflictos || 0);
    renderMetric(els.kpiCobertura, currentDashboard.coberturas_pendientes || 0);
    renderMetric(els.kpiPromedio, currentDashboard.promedio_horas_por_empleado || 0);

    renderSimpleList(els.alertasBox, currentDashboard.alertas || [], function (text) {
      return "<strong>Atención</strong><span>" + escapeHtml(text) + "</span>";
    }, "Sin alertas críticas.");

    renderSimpleList(els.oportunidadesBox, currentDashboard.oportunidades || [], function (text) {
      return "<strong>Oportunidad</strong><span>" + escapeHtml(text) + "</span>";
    }, "Sin oportunidades destacadas todavía.");

    renderSimpleList(els.areasBox, currentDashboard.areas || [], function (item) {
      return "<strong>" + escapeHtml(item.etiqueta) + "</strong><span>" + escapeHtml(item.cantidad) + " turnos · " + escapeHtml(item.horas) + " h</span>";
    }, "Todavía no hay carga por área.");

    renderSimpleList(els.sedesBox, currentDashboard.sedes || [], function (item) {
      return "<strong>" + escapeHtml(item.etiqueta) + "</strong><span>" + escapeHtml(item.cantidad) + " turnos · " + escapeHtml(item.horas) + " h</span>";
    }, "Todavía no hay carga por sede.");

    if (!els.semaforosBox) return;
    var semaforos = currentDashboard.semaforos || [];
    els.semaforosBox.innerHTML = semaforos.map(function (item) {
      return '<div class="horarios-semaforo" data-state="' + escapeHtml(item.estado) + '">' +
        "<strong>" + escapeHtml(item.titulo) + "</strong>" +
        "<span>" + escapeHtml(item.detalle) + "</span>" +
        "</div>";
    }).join("");
  }

  function statusBadge(item) {
    var label = item.estado || "programado";
    if (item.publicado && label === "programado") {
      label = "publicado";
    }
    var badgeClass = "badge low";
    if (item.conflicto) badgeClass = "badge urgent";
    else if (item.requiere_cobertura) badgeClass = "badge high";
    else if (label === "publicado") badgeClass = "badge mid";
    return '<span class="' + badgeClass + '">' + escapeHtml(label) + "</span>";
  }

  function renderRows(rows) {
    currentRows = rows || [];
    if (!currentRows.length) {
      els.tablaBody.innerHTML = '<tr><td colspan="7" class="horarios-empty">No hay turnos en el rango actual.</td></tr>';
      return;
    }
    els.tablaBody.innerHTML = currentRows.map(function (item) {
      var warnings = (item.conflictos_detectados || []).length
        ? '<div class="form-help" style="color:#ffb4b4;margin-top:6px;">' + escapeHtml(item.conflictos_detectados.join(" | ")) + "</div>"
        : "";
      var scope = escapeHtml((item.area || "Sin área") + " / " + (item.sede || "Sin sede"));
      var actions = [
        '<button class="btn secondary" type="button" data-action="edit" data-id="' + item.id + '">Editar</button>',
        '<button class="btn secondary" type="button" data-action="publish" data-id="' + item.id + '">Publicar</button>',
        '<button class="btn danger" type="button" data-action="delete" data-id="' + item.id + '">Eliminar</button>'
      ].join(" ");
      return "<tr>" +
        "<td><span class=\"horarios-color-dot\" style=\"background:" + escapeHtml(item.color || "#2563eb") + "\"></span><strong>" + escapeHtml(item.nombre_empleado || "") + "</strong><div class=\"form-help\">" + escapeHtml(item.cargo || "Sin cargo") + "</div>" + warnings + "</td>" +
        "<td>" + escapeHtml(item.fecha || "") + "<div class=\"form-help\">" + escapeHtml(item.turno_nombre || item.tipo_turno || "Turno") + "</div></td>" +
        "<td>" + escapeHtml(item.hora_inicio || "") + " - " + escapeHtml(item.hora_fin || "") + "<div class=\"form-help\">Descanso " + escapeHtml(item.descanso_minutos || 0) + " min</div></td>" +
        "<td>" + scope + "<div class=\"form-help\">" + escapeHtml(item.canal || "presencial") + "</div></td>" +
        "<td>" + statusBadge(item) + "</td>" +
        "<td>" + escapeHtml(item.horas_programadas || 0) + " h</td>" +
        "<td>" + actions + "</td>" +
        "</tr>";
    }).join("");
  }

  async function loadDashboard() {
    var url = endpoint + "&action=dashboard&desde=" + encodeURIComponent(els.filtroDesde.value) + "&hasta=" + encodeURIComponent(els.filtroHasta.value);
    var data = await fetchJSON(url);
    renderDashboard(data || {});
  }

  async function loadRows() {
    var url = endpoint +
      "&desde=" + encodeURIComponent(els.filtroDesde.value) +
      "&hasta=" + encodeURIComponent(els.filtroHasta.value) +
      "&q=" + encodeURIComponent(els.filtroQ.value || "") +
      "&area=" + encodeURIComponent(els.filtroArea.value || "") +
      "&sede=" + encodeURIComponent(els.filtroSede.value || "") +
      "&estado=" + encodeURIComponent(els.filtroEstado.value || "") +
      "&published_only=" + encodeURIComponent(els.filtroPublicados.checked ? "1" : "0");
    var data = await fetchJSON(url);
    renderRows((data && data.items) || []);
  }

  async function loadUsers() {
    try {
      var data = await fetchJSON("/api/empresa/usuarios?empresa_id=" + encodeURIComponent(empresaId));
      var items = Array.isArray(data) ? data : (data.items || []);
      els.usuarioId.innerHTML = '<option value="">Registro manual</option>' + items.map(function (item) {
        var name = item.nombre_completo || item.nombre || item.email || ("Usuario " + item.id);
        var role = item.rol || item.cargo || "";
        return '<option value="' + escapeHtml(item.id) + '" data-name="' + escapeHtml(name) + '" data-role="' + escapeHtml(role) + '">' + escapeHtml(name) + (role ? " · " + escapeHtml(role) : "") + "</option>";
      }).join("");
    } catch (error) {
      els.usuarioId.innerHTML = '<option value="">Registro manual</option>';
    }
  }

  function resetForm() {
    editingId = 0;
    els.formTurno.reset();
    els.turnoId.value = "";
    els.descansoMinutos.value = "30";
    els.colorTurno.value = "#2563eb";
    els.fechaInicio.value = els.filtroDesde.value;
    els.fechaFin.value = els.filtroDesde.value;
    setMessage(els.turnoMsg, "");
    Array.prototype.forEach.call(document.querySelectorAll(".weekday-check"), function (el) {
      el.checked = false;
    });
  }

  function pickRowById(id) {
    for (var i = 0; i < currentRows.length; i += 1) {
      if (String(currentRows[i].id) === String(id)) return currentRows[i];
    }
    return null;
  }

  function hydrateForm(item) {
    editingId = Number(item.id || 0);
    els.turnoId.value = String(item.id || "");
    els.usuarioId.value = item.usuario_id || "";
    els.nombreEmpleado.value = item.nombre_empleado || "";
    els.cargo.value = item.cargo || "";
    els.area.value = item.area || "";
    els.sede.value = item.sede || "";
    els.fechaInicio.value = item.fecha || "";
    els.fechaFin.value = item.fecha || "";
    els.horaInicio.value = item.hora_inicio || "";
    els.horaFin.value = item.hora_fin || "";
    els.descansoMinutos.value = item.descanso_minutos || 0;
    els.turnoNombre.value = item.turno_nombre || "";
    els.tipoTurno.value = item.tipo_turno || "operativo";
    els.canal.value = item.canal || "presencial";
    els.estadoTurno.value = item.estado || "programado";
    els.colorTurno.value = item.color || "#2563eb";
    els.observaciones.value = item.observaciones || "";
    els.publicado.checked = !!item.publicado;
    els.requiereCobertura.checked = !!item.requiere_cobertura;
    Array.prototype.forEach.call(document.querySelectorAll(".weekday-check"), function (el) { el.checked = false; });
    setMessage(els.turnoMsg, "Editando turno de " + (item.nombre_empleado || "empleado") + ".");
  }

  function buildPayload() {
    return {
      usuario_id: parsePositiveInt(els.usuarioId.value) || null,
      nombre_empleado: els.nombreEmpleado.value.trim(),
      cargo: els.cargo.value.trim(),
      area: els.area.value.trim(),
      sede: els.sede.value.trim(),
      fecha_inicio: els.fechaInicio.value,
      fecha_fin: els.fechaFin.value,
      hora_inicio: els.horaInicio.value,
      hora_fin: els.horaFin.value,
      descanso_minutos: Number(els.descansoMinutos.value || 0),
      turno_nombre: els.turnoNombre.value.trim(),
      tipo_turno: els.tipoTurno.value,
      canal: els.canal.value,
      color: els.colorTurno.value,
      estado: els.estadoTurno.value,
      publicado: !!els.publicado.checked,
      requiere_cobertura: !!els.requiereCobertura.checked,
      observaciones: els.observaciones.value.trim(),
      dias_semana: weekdayValues()
    };
  }

  async function saveTurno(ev) {
    ev.preventDefault();
    setMessage(els.turnoMsg, "");
    var payload = buildPayload();
    if (!payload.nombre_empleado) {
      setMessage(els.turnoMsg, "Debes indicar el nombre del trabajador.", true);
      return;
    }
    try {
      if (editingId > 0) {
        payload.id = editingId;
        payload.fecha = payload.fecha_inicio;
        delete payload.fecha_inicio;
        delete payload.fecha_fin;
        delete payload.dias_semana;
        var updateRes = await fetchJSON(endpoint, {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload)
        });
        setMessage(els.turnoMsg, updateRes.conflictos_detectados && updateRes.conflictos_detectados.length
          ? "Turno actualizado con conflictos controlados."
          : "Turno actualizado correctamente.");
      } else if (payload.fecha_inicio === payload.fecha_fin && payload.dias_semana.length <= 1) {
        var singlePayload = {
          usuario_id: payload.usuario_id,
          nombre_empleado: payload.nombre_empleado,
          cargo: payload.cargo,
          area: payload.area,
          sede: payload.sede,
          fecha: payload.fecha_inicio,
          hora_inicio: payload.hora_inicio,
          hora_fin: payload.hora_fin,
          descanso_minutos: payload.descanso_minutos,
          turno_nombre: payload.turno_nombre,
          tipo_turno: payload.tipo_turno,
          canal: payload.canal,
          color: payload.color,
          estado: payload.estado,
          publicado: payload.publicado,
          requiere_cobertura: payload.requiere_cobertura,
          observaciones: payload.observaciones
        };
        var createRes = await fetchJSON(endpoint, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(singlePayload)
        });
        setMessage(els.turnoMsg, createRes.conflictos_detectados && createRes.conflictos_detectados.length
          ? "Turno guardado con conflictos controlados."
          : "Turno creado correctamente.");
      } else {
        var bulkRes = await fetchJSON(endpoint + "&action=bulk_create", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload)
        });
        var msg = "Se programaron " + (bulkRes.creados || 0) + " turnos.";
        if (bulkRes.warnings && bulkRes.warnings.length) {
          msg += " Hubo " + bulkRes.warnings.length + " fechas omitidas por conflicto.";
        }
        setMessage(els.turnoMsg, msg);
      }
      resetForm();
      await refreshAll();
    } catch (error) {
      setMessage(els.turnoMsg, error.message || "No se pudo guardar la programación.", true);
    }
  }

  async function saveConfig() {
    setMessage(els.configMsg, "");
    try {
      await fetchJSON(endpoint + "&action=config", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          horas_objetivo_dia: Number(els.cfgHorasDia.value || 8),
          horas_objetivo_semana: Number(els.cfgHorasSemana.value || 48),
          descanso_minimo_minutos: Number(els.cfgDescanso.value || 30),
          permitir_solapados: !!els.cfgPermitirSolapados.checked,
          anticipacion_publicacion_horas: Number(els.cfgAnticipacion.value || 24)
        })
      });
      setMessage(els.configMsg, "Reglas guardadas correctamente.");
      await refreshAll();
    } catch (error) {
      setMessage(els.configMsg, error.message || "No se pudo guardar la configuración.", true);
    }
  }

  async function publishRange() {
    try {
      await fetchJSON(endpoint + "&action=publish", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          desde: els.filtroDesde.value,
          hasta: els.filtroHasta.value
        })
      });
      setMessage(els.turnoMsg, "Rango publicado correctamente.");
      await refreshAll();
    } catch (error) {
      setMessage(els.turnoMsg, error.message || "No se pudo publicar el rango.", true);
    }
  }

  async function handleTableClick(ev) {
    var target = ev.target;
    if (!target || !target.dataset || !target.dataset.action) return;
    var action = target.dataset.action;
    var id = target.dataset.id;
    var row = pickRowById(id);
    if (!row) return;
    if (action === "edit") {
      hydrateForm(row);
      return;
    }
    if (action === "publish") {
      try {
        await fetchJSON(endpoint + "&action=publish", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ ids: [Number(id)] })
        });
        await refreshAll();
      } catch (error) {
        setMessage(els.turnoMsg, error.message || "No se pudo publicar el turno.", true);
      }
      return;
    }
    if (action === "delete") {
      if (!window.confirm("¿Deseas eliminar este turno?")) return;
      try {
        await fetchJSON(endpoint + "&id=" + encodeURIComponent(id), { method: "DELETE" });
        if (editingId === Number(id)) resetForm();
        await refreshAll();
      } catch (error) {
        setMessage(els.turnoMsg, error.message || "No se pudo eliminar el turno.", true);
      }
    }
  }

  async function refreshAll() {
    await Promise.all([loadDashboard(), loadRows()]);
  }

  function bindEvents() {
    els.formTurno.addEventListener("submit", saveTurno);
    els.btnGuardarConfig.addEventListener("click", saveConfig);
    els.btnRefrescar.addEventListener("click", function () { refreshAll().catch(console.error); });
    els.btnPublicarRango.addEventListener("click", function () { publishRange().catch(console.error); });
    els.btnLimpiarTurno.addEventListener("click", resetForm);
    els.usuarioId.addEventListener("change", function () {
      var selected = els.usuarioId.options[els.usuarioId.selectedIndex];
      if (!selected) return;
      if (!els.nombreEmpleado.value.trim() && selected.dataset.name) {
        els.nombreEmpleado.value = selected.dataset.name;
      }
      if (!els.cargo.value.trim() && selected.dataset.role) {
        els.cargo.value = selected.dataset.role;
      }
    });
    els.tablaBody.addEventListener("click", handleTableClick);
  }

  async function init() {
    if (!empresaId) {
      setMessage(els.turnoMsg, "No se pudo resolver la empresa activa.", true);
      return;
    }
    var today = new Date();
    els.filtroDesde.value = isoDate(today);
    els.filtroHasta.value = isoDate(addDays(today, 6));
    els.fechaInicio.value = els.filtroDesde.value;
    els.fechaFin.value = els.filtroDesde.value;
    els.horaInicio.value = "08:00";
    els.horaFin.value = "17:00";
    bindEvents();
    await Promise.all([loadUsers(), loadConfig()]);
    await refreshAll();
  }

  init().catch(function (error) {
    setMessage(els.turnoMsg, error && error.message ? error.message : "No se pudo inicializar el módulo.", true);
  });
})();
