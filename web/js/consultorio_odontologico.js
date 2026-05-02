(function () {
  function parsePositiveInt(raw) {
    var n = Number(String(raw || "").trim());
    if (!Number.isFinite(n)) return 0;
    n = Math.trunc(n);
    return n > 0 ? n : 0;
  }

  function getEmpresaId() {
    try {
      if (typeof window.__resolveEmpresaIdContext === "function") {
        var id = parsePositiveInt(window.__resolveEmpresaIdContext());
        if (id > 0) return String(id);
      }
    } catch (e) {}
    var params = new URLSearchParams(window.location.search || "");
    return String(parsePositiveInt(params.get("empresa_id") || params.get("id") || ""));
  }

  var empresaId = getEmpresaId();
  var apiBase = "/api/empresa/odontologia?empresa_id=" + encodeURIComponent(empresaId);
  var state = {
    pacientes: [],
    profesionales: [],
    consultorios: [],
    tratamientos: [],
    presupuestos: []
  };

  function escapeHtml(value) {
    return String(value == null ? "" : value).replace(/[&<>\"']/g, function (m) {
      return { "&": "&amp;", "<": "&lt;", ">": "&gt;", "\"": "&quot;", "'": "&#39;" }[m];
    });
  }

  function setNotice(text, isError) {
    var el = document.getElementById("odontoNotice");
    if (!el) return;
    el.textContent = text || "";
    el.style.color = isError ? "#b91c1c" : "";
  }

  async function fetchJSON(url, options) {
    var resp = await fetch(url, Object.assign({ credentials: "same-origin" }, options || {}));
    var raw = await resp.text();
    var data = raw ? JSON.parse(raw) : null;
    if (!resp.ok) {
      throw new Error((data && (data.error || data.message)) || raw || "Solicitud fallida");
    }
    return data;
  }

  function toPayload(form) {
    var data = {};
    new FormData(form).forEach(function (value, key) {
      var text = String(value || "").trim();
      if (text === "") return;
      if (["paciente_id", "profesional_id", "consultorio_id", "tratamiento_id", "sesiones_total", "sesiones_realizadas", "duracion_minutos"].indexOf(key) >= 0) {
        data[key] = parsePositiveInt(text);
        return;
      }
      if (["valor_total", "cuota_inicial", "monto", "costo_estimado", "costo_real"].indexOf(key) >= 0) {
        data[key] = Number(text) || 0;
        return;
      }
      data[key] = text;
    });
    return data;
  }

  function metricCard(label, value) {
    return '<article class="erp-metric-card"><span class="erp-metric-label">' + escapeHtml(label) + '</span><strong class="erp-metric-value">' + escapeHtml(value) + "</strong></article>";
  }

  function renderDashboard(row) {
    var host = document.getElementById("odontoDashboard");
    if (!host || !row) return;
    host.innerHTML = [
      metricCard("Pacientes activos", row.pacientes_activos || 0),
      metricCard("Profesionales", row.profesionales_activos || 0),
      metricCard("Citas hoy", row.citas_hoy || 0),
      metricCard("Pendientes", row.citas_pendientes || 0),
      metricCard("Tratamientos", row.tratamientos_activos || 0),
      metricCard("Presupuestos", row.presupuestos_vigentes || 0),
      metricCard("Recaudo mes", row.recaudo_mes || 0),
      metricCard("Saldo pendiente", row.saldo_pendiente || 0)
    ].join("");
  }

  function renderSimpleList(hostId, items, formatter) {
    var host = document.getElementById(hostId);
    if (!host) return;
    if (!Array.isArray(items) || !items.length) {
      host.innerHTML = '<div class="muted">Sin registros todavia.</div>';
      return;
    }
    host.innerHTML = items.map(formatter).join("");
  }

  function itemCard(title, lines, actionsHtml) {
    return '<article class="erp-list-card"><strong>' + escapeHtml(title) + '</strong><div class="muted" style="margin-top:6px;">' + lines.map(escapeHtml).join(" · ") + '</div>' + (actionsHtml || "") + "</article>";
  }

  function renderSelectOptions(selector, items, valueKey, labelFn, includeEmptyLabel) {
    document.querySelectorAll(selector).forEach(function (select) {
      var current = String(select.value || "");
      var html = includeEmptyLabel ? '<option value="">' + includeEmptyLabel + "</option>" : "";
      html += items.map(function (item) {
        var value = String(item[valueKey] || "");
        var label = labelFn(item);
        return '<option value="' + escapeHtml(value) + '">' + escapeHtml(label) + "</option>";
      }).join("");
      select.innerHTML = html;
      if (current) select.value = current;
    });
  }

  async function loadDashboard() {
    renderDashboard(await fetchJSON(apiBase + "&action=dashboard"));
  }

  async function loadPacientes() {
    state.pacientes = await fetchJSON(apiBase + "&action=pacientes");
    renderSimpleList("odontoPacientesList", state.pacientes, function (item) {
      return itemCard(item.nombre_completo || "Paciente", [
        item.documento || "Sin documento",
        item.telefono || "Sin telefono",
        item.aseguradora || "Particular",
        "Saldo " + (item.saldo || 0),
        item.estado || "activo"
      ], '<div style="margin-top:8px;"><button type="button" class="btn secondary small" data-entity="paciente" data-id="' + escapeHtml(item.id) + '" data-next="' + escapeHtml(item.estado === "inactivo" ? "activo" : "inactivo") + '">Cambiar estado</button></div>');
    });
    renderSelectOptions('select[name="paciente_id"]', state.pacientes, "id", function (item) { return (item.nombre_completo || "Paciente") + " · " + (item.documento || item.codigo || ""); }, "Seleccione paciente");
  }

  async function loadProfesionales() {
    state.profesionales = await fetchJSON(apiBase + "&action=profesionales");
    renderSimpleList("odontoProfesionalesList", state.profesionales, function (item) {
      return itemCard(item.nombre_completo || "Profesional", [
        item.especialidad || "General",
        item.registro_profesional || "Sin registro",
        item.estado || "activo"
      ], '<div style="margin-top:8px;"><button type="button" class="btn secondary small" data-entity="profesional" data-id="' + escapeHtml(item.id) + '" data-next="' + escapeHtml(item.estado === "inactivo" ? "activo" : "inactivo") + '">Cambiar estado</button></div>');
    });
    renderSelectOptions('select[name="profesional_id"]', state.profesionales, "id", function (item) { return (item.nombre_completo || "Profesional") + " · " + (item.especialidad || "General"); }, "Seleccione profesional");
  }

  async function loadConsultorios() {
    state.consultorios = await fetchJSON(apiBase + "&action=consultorios");
    renderSimpleList("odontoConsultoriosList", state.consultorios, function (item) {
      return itemCard(item.nombre || "Consultorio", [item.sede || "Sin sede", item.sillon || "Sin sillon", item.estado || "activo"], '<div style="margin-top:8px;"><button type="button" class="btn secondary small" data-entity="consultorio" data-id="' + escapeHtml(item.id) + '" data-next="' + escapeHtml(item.estado === "inactivo" ? "activo" : "inactivo") + '">Cambiar estado</button></div>');
    });
    renderSelectOptions('select[name="consultorio_id"]', state.consultorios, "id", function (item) { return (item.nombre || "Consultorio") + " · " + (item.sede || ""); }, "Sin consultorio");
  }

  async function loadCitas() {
    var rows = await fetchJSON(apiBase + "&action=citas");
    renderSimpleList("odontoCitasList", rows, function (item) {
      return itemCard((item.paciente_nombre || "Paciente") + " con " + (item.profesional_nombre || "Profesional"), [
        item.fecha_hora || "Sin fecha",
        item.consultorio_nombre || "Sin consultorio",
        item.estado || "programada",
        item.motivo || "Sin motivo"
      ], '<div style="margin-top:8px;"><button type="button" class="btn secondary small" data-entity="cita" data-id="' + escapeHtml(item.id) + '" data-next="atendida">Marcar atendida</button></div>');
    });
  }

  async function loadHistorias() {
    var rows = await fetchJSON(apiBase + "&action=historias");
    renderSimpleList("odontoHistoriasList", rows, function (item) {
      return itemCard(item.paciente_nombre || "Historia", [item.fecha_atencion || "Sin fecha", item.diagnostico || "Sin diagnostico", item.profesional_nombre || "Sin profesional"]);
    });
  }

  async function loadOdontogramas() {
    var rows = await fetchJSON(apiBase + "&action=odontogramas");
    renderSimpleList("odontoOdontogramasList", rows, function (item) {
      return itemCard(item.paciente_nombre || "Odontograma", [item.fecha_registro || "Sin fecha", item.estado || "activo"]);
    });
  }

  async function loadTratamientos() {
    state.tratamientos = await fetchJSON(apiBase + "&action=tratamientos");
    renderSimpleList("odontoTratamientosList", state.tratamientos, function (item) {
      return itemCard(item.nombre || "Tratamiento", [
        item.paciente_nombre || "Sin paciente",
        "Sesiones " + (item.sesiones_realizadas || 0) + "/" + (item.sesiones_total || 0),
        "Estimado " + (item.costo_estimado || 0),
        item.estado || "planificado"
      ], '<div style="margin-top:8px;"><button type="button" class="btn secondary small" data-entity="tratamiento" data-id="' + escapeHtml(item.id) + '" data-next="en_proceso">Pasar a en proceso</button></div>');
    });
    renderSelectOptions('select[name="tratamiento_id"]', state.tratamientos, "id", function (item) { return (item.nombre || "Tratamiento") + " · " + (item.paciente_nombre || ""); }, "Sin tratamiento");
  }

  async function loadPresupuestos() {
    state.presupuestos = await fetchJSON(apiBase + "&action=presupuestos");
    renderSimpleList("odontoPresupuestosList", state.presupuestos, function (item) {
      return itemCard(item.nombre || "Presupuesto", [
        item.paciente_nombre || "Sin paciente",
        "Total " + (item.valor_total || 0),
        "Saldo " + (item.saldo || 0),
        item.estado || "vigente"
      ], '<div style="margin-top:8px;"><button type="button" class="btn secondary small" data-entity="presupuesto" data-id="' + escapeHtml(item.id) + '" data-next="aprobado">Aprobar</button></div>');
    });
    renderSelectOptions('select[name="presupuesto_id"]', state.presupuestos, "id", function (item) { return (item.nombre || "Presupuesto") + " · saldo " + (item.saldo || 0); }, "Sin presupuesto");
  }

  async function loadPagos() {
    var rows = await fetchJSON(apiBase + "&action=pagos");
    renderSimpleList("odontoPagosList", rows, function (item) {
      return itemCard(item.concepto || "Pago", [
        item.paciente_nombre || "Sin paciente",
        "Monto " + (item.monto || 0),
        item.metodo_pago || "Sin metodo",
        item.fecha_pago || "Sin fecha"
      ]);
    });
  }

  async function refreshAll() {
    await loadDashboard();
    await loadPacientes();
    await loadProfesionales();
    await loadConsultorios();
    await loadCitas();
    await loadHistorias();
    await loadOdontogramas();
    await loadTratamientos();
    await loadPresupuestos();
    await loadPagos();
  }

  function showTab(name) {
    document.querySelectorAll(".odonto-tab").forEach(function (panel) {
      panel.hidden = panel.getAttribute("data-panel") !== name;
    });
  }

  async function submitCreate(formId, action, successMessage) {
    var form = document.getElementById(formId);
    if (!form) return;
    form.addEventListener("submit", async function (event) {
      event.preventDefault();
      try {
        await fetchJSON(apiBase + "&action=" + encodeURIComponent(action), {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(toPayload(form))
        });
        form.reset();
        setNotice(successMessage, false);
        await refreshAll();
      } catch (err) {
        setNotice(err.message || "No se pudo guardar", true);
      }
    });
  }

  async function changeEstado(entity, id, nextState) {
    var map = {
      paciente: "estado_paciente",
      profesional: "estado_profesional",
      consultorio: "estado_consultorio",
      cita: "estado_cita",
      tratamiento: "estado_tratamiento",
      presupuesto: "estado_presupuesto"
    };
    var action = map[entity];
    if (!action) return;
    await fetchJSON(apiBase + "&action=" + encodeURIComponent(action) + "&id=" + encodeURIComponent(id) + "&estado=" + encodeURIComponent(nextState), {
      method: "PUT"
    });
    await refreshAll();
  }

  document.addEventListener("click", function (event) {
    var tabBtn = event.target.closest("[data-tab]");
    if (tabBtn) {
      showTab(tabBtn.getAttribute("data-tab"));
      return;
    }
    var stateBtn = event.target.closest("[data-entity]");
    if (stateBtn) {
      changeEstado(stateBtn.getAttribute("data-entity"), stateBtn.getAttribute("data-id"), stateBtn.getAttribute("data-next")).catch(function (err) {
        setNotice(err.message || "No se pudo cambiar el estado", true);
      });
    }
  });

  submitCreate("odontoPacienteForm", "pacientes", "Paciente guardado.");
  submitCreate("odontoProfesionalForm", "profesionales", "Profesional guardado.");
  submitCreate("odontoConsultorioForm", "consultorios", "Consultorio guardado.");
  submitCreate("odontoCitaForm", "citas", "Cita guardada.");
  submitCreate("odontoHistoriaForm", "historias", "Historia clinica guardada.");
  submitCreate("odontoOdontogramaForm", "odontogramas", "Odontograma guardado.");
  submitCreate("odontoTratamientoForm", "tratamientos", "Tratamiento guardado.");
  submitCreate("odontoPresupuestoForm", "presupuestos", "Presupuesto guardado.");
  submitCreate("odontoPagoForm", "pagos", "Pago registrado.");

  showTab("pacientes");
  refreshAll().catch(function (err) {
    setNotice(err.message || "No se pudo cargar el modulo", true);
  });
})();
