(function () {
  "use strict";

  var state = {
    empresaID: 0,
    tab: "dashboard",
    socios: [],
    planes: [],
    entrenadores: [],
    clases: [],
    inscripciones: [],
    asistencias: [],
    pagos: []
  };

  function resolveEmpresaID() {
    if (window.__resolveEmpresaIdContext) {
      return Number(window.__resolveEmpresaIdContext() || 0);
    }
    try {
      return Number(new URLSearchParams(window.location.search).get("empresa_id") || 0);
    } catch (error) {
      return 0;
    }
  }

  function byId(id) {
    return document.getElementById(id);
  }

  function setMessage(text, isError) {
    var node = byId("gymMsg");
    if (!node) return;
    node.textContent = text || "";
    node.classList.toggle("is-hidden", !text);
    node.classList.toggle("error", !!text && !!isError);
    node.classList.toggle("success", !!text && !isError);
  }

  function apiURL(action) {
    return "/api/empresa/gimnasio?empresa_id=" + encodeURIComponent(state.empresaID) + "&action=" + encodeURIComponent(action);
  }

  function readJSON(response) {
    return response.text().then(function (text) {
      var payload = {};
      try {
        payload = text ? JSON.parse(text) : {};
      } catch (error) {
        payload = { message: text || ("HTTP " + response.status) };
      }
      if (!response.ok) {
        throw new Error(payload.message || payload.error || ("HTTP " + response.status));
      }
      return payload;
    });
  }

  function fetchJSON(action) {
    return fetch(apiURL(action), { credentials: "same-origin" }).then(readJSON);
  }

  function sendJSON(method, action, payload, extraQuery) {
    var url = apiURL(action);
    if (extraQuery) {
      url += "&" + extraQuery;
    }
    return fetch(url, {
      method: method,
      credentials: "same-origin",
      headers: { "Content-Type": "application/json" },
      body: payload ? JSON.stringify(payload) : null
    }).then(readJSON);
  }

  function formatMoney(value) {
    return new Intl.NumberFormat("es-CO", { style: "currency", currency: "COP", maximumFractionDigits: 0 }).format(Number(value || 0));
  }

  function escapeHTML(value) {
    return String(value || "").replace(/[&<>\"']/g, function (char) {
      return ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "\"": "&quot;", "'": "&#39;" })[char];
    });
  }

  function renderList(items, fields) {
    if (!items || !items.length) {
      return '<p class="form-help">Sin registros todavía.</p>';
    }
    var rows = items.map(function (item) {
      return "<tr>" + fields.map(function (field) {
        return "<td>" + escapeHTML(field.format ? field.format(item[field.key], item) : item[field.key]) + "</td>";
      }).join("") + (fields.actions ? "<td>" + fields.actions(item) + "</td>" : "") + "</tr>";
    }).join("");
    var head = fields.map(function (field) { return "<th>" + escapeHTML(field.label) + "</th>"; }).join("") + (fields.actions ? "<th>Acciones</th>" : "");
    return '<div class="table-responsive"><table class="data-table"><thead><tr>' + head + '</tr></thead><tbody>' + rows + '</tbody></table></div>';
  }

  function setTab(tab) {
    state.tab = tab;
    Array.prototype.slice.call(document.querySelectorAll(".gym-tab")).forEach(function (section) {
      section.hidden = section.id !== ("gymTab-" + tab);
    });
  }

  function populateSelect(selectId, items, valueKey, labelBuilder, includeBlank) {
    var select = byId(selectId);
    if (!select) return;
    var currentValue = String(select.value || "");
    var options = [];
    if (includeBlank) {
      options.push('<option value="">Sin seleccionar</option>');
    }
    items.forEach(function (item) {
      var value = String(item[valueKey] || "");
      options.push('<option value="' + escapeHTML(value) + '">' + escapeHTML(labelBuilder(item)) + "</option>");
    });
    select.innerHTML = options.join("");
    if (currentValue) {
      select.value = currentValue;
    }
  }

  function syncSelects() {
    populateSelect("gymSocioPlan", state.planes, "id", function (item) { return item.nombre; }, true);
    populateSelect("gymClassTrainer", state.entrenadores, "id", function (item) { return item.nombre_completo; }, true);
    populateSelect("gymEnrollmentSocio", state.socios, "id", function (item) { return item.nombre_completo; }, true);
    populateSelect("gymAttendanceSocio", state.socios, "id", function (item) { return item.nombre_completo; }, true);
    populateSelect("gymPaymentSocio", state.socios, "id", function (item) { return item.nombre_completo; }, true);
    populateSelect("gymEnrollmentClase", state.clases, "id", function (item) { return item.nombre + " · " + (item.fecha_programada || ""); }, true);
    populateSelect("gymAttendanceClase", state.clases, "id", function (item) { return item.nombre + " · " + (item.fecha_programada || ""); }, true);
    populateSelect("gymPaymentPlan", state.planes, "id", function (item) { return item.nombre; }, true);
  }

  function renderDashboard(data) {
    var cards = [
      ["Socios activos", data.socios_activos || 0],
      ["Planes activos", data.planes_activos || 0],
      ["Clases hoy", data.clases_hoy || 0],
      ["Accesos hoy", data.accesos_hoy || 0],
      ["Ingresos del mes", formatMoney(data.ingresos_mes || 0)],
      ["Renovaciones próximas", data.renovaciones_proximas || 0],
      ["Inscripciones activas", data.inscripciones_activas || 0]
    ];
    byId("gymDashboardCards").innerHTML = cards.map(function (card) {
      return '<article class="card"><p class="form-help" style="margin-bottom:6px;">' + escapeHTML(card[0]) + '</p><strong style="font-size:1.4rem;">' + escapeHTML(card[1]) + "</strong></article>";
    }).join("");

    byId("gymRenewals").innerHTML = renderList(data.vencimientos_proximos || [], [
      { key: "nombre_completo", label: "Socio" },
      { key: "plan_nombre", label: "Plan" },
      { key: "fecha_fin_plan", label: "Vence" }
    ]);
    byId("gymTodayClasses").innerHTML = renderList(data.clases_programadas_hoy || [], [
      { key: "nombre", label: "Clase" },
      { key: "entrenador_nombre", label: "Entrenador" },
      { key: "fecha_programada", label: "Hora" }
    ]);
    byId("gymByChannel").innerHTML = renderList(data.ingresos_por_canal || [], [
      { key: "etiqueta", label: "Canal" },
      { key: "cantidad", label: "Movimientos" },
      { key: "monto", label: "Ingresos", format: function (value) { return formatMoney(value); } }
    ]);
    byId("gymByLine").innerHTML = renderList(data.rentabilidad_por_linea || [], [
      { key: "etiqueta", label: "Línea" },
      { key: "monto", label: "Ingresos", format: function (value) { return formatMoney(value); } },
      { key: "margen", label: "Margen", format: function (value) { return formatMoney(value); } }
    ]);
    byId("gymBySede").innerHTML = renderList(data.rentabilidad_por_sede || [], [
      { key: "etiqueta", label: "Sede" },
      { key: "monto", label: "Ingresos", format: function (value) { return formatMoney(value); } },
      { key: "margen", label: "Margen", format: function (value) { return formatMoney(value); } }
    ]);
  }

  function renderTables() {
    byId("gymSociosTable").innerHTML = renderList(state.socios, [
      { key: "nombre_completo", label: "Socio" },
      { key: "plan_nombre", label: "Plan" },
      { key: "telefono", label: "Teléfono" },
      { key: "fecha_fin_plan", label: "Fin del plan" },
      { key: "saldo", label: "Saldo", format: function (value) { return formatMoney(value); } },
      { actions: function (item) { return '<button type="button" class="btn secondary small" data-edit="socio" data-id="' + item.id + '">Editar</button> <button type="button" class="btn secondary small" data-delete="socios" data-id="' + item.id + '">Eliminar</button>'; } }
    ]);

    byId("gymPlanesTable").innerHTML = renderList(state.planes, [
      { key: "nombre", label: "Plan" },
      { key: "precio", label: "Precio", format: function (value) { return formatMoney(value); } },
      { key: "duracion_dias", label: "Días" },
      { key: "clases_incluidas", label: "Clases" },
      { key: "estado", label: "Estado" },
      { actions: function (item) { return '<button type="button" class="btn secondary small" data-edit="plan" data-id="' + item.id + '">Editar</button> <button type="button" class="btn secondary small" data-delete="planes" data-id="' + item.id + '">Eliminar</button>'; } }
    ]);

    byId("gymTrainersTable").innerHTML = renderList(state.entrenadores, [
      { key: "nombre_completo", label: "Entrenador" },
      { key: "especialidad", label: "Especialidad" },
      { key: "telefono", label: "Teléfono" },
      { key: "estado", label: "Estado" },
      { actions: function (item) { return '<button type="button" class="btn secondary small" data-edit="trainer" data-id="' + item.id + '">Editar</button> <button type="button" class="btn secondary small" data-delete="entrenadores" data-id="' + item.id + '">Eliminar</button>'; } }
    ]);

    byId("gymClassesTable").innerHTML = renderList(state.clases, [
      { key: "nombre", label: "Clase" },
      { key: "categoria", label: "Categoría" },
      { key: "entrenador_nombre", label: "Entrenador" },
      { key: "fecha_programada", label: "Programación" },
      { key: "cupos", label: "Cupos" },
      { actions: function (item) { return '<button type="button" class="btn secondary small" data-edit="class" data-id="' + item.id + '">Editar</button> <button type="button" class="btn secondary small" data-delete="clases" data-id="' + item.id + '">Eliminar</button>'; } }
    ]);

    byId("gymEnrollmentsTable").innerHTML = renderList(state.inscripciones, [
      { key: "socio_nombre", label: "Socio" },
      { key: "clase_nombre", label: "Clase" },
      { key: "estado", label: "Estado" },
      { key: "asistencia_marcada", label: "Asistencia", format: function (value) { return value ? "Sí" : "No"; } },
      { actions: function (item) { return '<button type="button" class="btn secondary small" data-cancel-enrollment="' + item.id + '">Cancelar</button>'; } }
    ]);

    byId("gymAttendanceTable").innerHTML = renderList(state.asistencias, [
      { key: "socio_nombre", label: "Socio" },
      { key: "clase_nombre", label: "Clase" },
      { key: "fecha_hora", label: "Fecha y hora" },
      { key: "sede", label: "Sede" },
      { key: "canal", label: "Canal" }
    ]);

    byId("gymPaymentsTable").innerHTML = renderList(state.pagos, [
      { key: "socio_nombre", label: "Socio" },
      { key: "concepto", label: "Concepto" },
      { key: "monto", label: "Monto", format: function (value) { return formatMoney(value); } },
      { key: "metodo_pago", label: "Método" },
      { key: "fecha_pago", label: "Fecha" }
    ]);
  }

  function toDateInput(value) {
    return String(value || "").slice(0, 10);
  }

  function toDateTimeInput(value) {
    return String(value || "").replace(" ", "T").slice(0, 16);
  }

  function resetForm(formId) {
    var form = byId(formId);
    if (!form) return;
    form.reset();
    Array.prototype.slice.call(form.querySelectorAll('input[type="hidden"]')).forEach(function (input) { input.value = ""; });
  }

  function loadForEdit(kind, id) {
    var source = [];
    if (kind === "socio") source = state.socios;
    if (kind === "plan") source = state.planes;
    if (kind === "trainer") source = state.entrenadores;
    if (kind === "class") source = state.clases;
    var item = source.find(function (row) { return Number(row.id) === Number(id); });
    if (!item) return;
    if (kind === "socio") {
      byId("gymSocioId").value = item.id;
      byId("gymSocioNombre").value = item.nombre_completo || "";
      byId("gymSocioCodigo").value = item.codigo || "";
      byId("gymSocioTelefono").value = item.telefono || "";
      byId("gymSocioEmail").value = item.email || "";
      byId("gymSocioPlan").value = item.plan_id || "";
      byId("gymSocioFechaFin").value = toDateInput(item.fecha_fin_plan);
      byId("gymSocioObjetivo").value = item.objetivo || "";
      byId("gymSocioSaldo").value = item.saldo || 0;
      byId("gymSocioEstado").value = item.estado || "activo";
      setTab("socios");
    }
    if (kind === "plan") {
      byId("gymPlanId").value = item.id;
      byId("gymPlanNombre").value = item.nombre || "";
      byId("gymPlanPrecio").value = item.precio || 0;
      byId("gymPlanDias").value = item.duracion_dias || 30;
      byId("gymPlanClases").value = item.clases_incluidas || 0;
      byId("gymPlanSesiones").value = item.sesiones_personalizadas || 0;
      byId("gymPlanEstado").value = item.estado || "activo";
      byId("gymPlanIlimitado").checked = !!item.acceso_ilimitado;
      setTab("planes");
    }
    if (kind === "trainer") {
      byId("gymTrainerId").value = item.id;
      byId("gymTrainerNombre").value = item.nombre_completo || "";
      byId("gymTrainerEspecialidad").value = item.especialidad || "";
      byId("gymTrainerTelefono").value = item.telefono || "";
      byId("gymTrainerEmail").value = item.email || "";
      byId("gymTrainerDisponibilidad").value = item.disponibilidad || "";
      byId("gymTrainerEstado").value = item.estado || "activo";
      setTab("entrenadores");
    }
    if (kind === "class") {
      byId("gymClassId").value = item.id;
      byId("gymClassNombre").value = item.nombre || "";
      byId("gymClassCategoria").value = item.categoria || "";
      byId("gymClassTrainer").value = item.entrenador_id || "";
      byId("gymClassSede").value = item.sede || "principal";
      byId("gymClassCanal").value = item.canal || "presencial";
      byId("gymClassFecha").value = toDateTimeInput(item.fecha_programada);
      byId("gymClassCupos").value = item.cupos || 20;
      byId("gymClassDuracion").value = item.duracion_minutos || 60;
      byId("gymClassPrecio").value = item.precio || 0;
      setTab("clases");
    }
  }

  function collectValue(id) {
    return byId(id) ? byId(id).value : "";
  }

  function parseNumber(id) {
    return Number(collectValue(id) || 0);
  }

  function payloadSocio() {
    return {
      id: parseNumber("gymSocioId"),
      empresa_id: state.empresaID,
      nombre_completo: collectValue("gymSocioNombre"),
      codigo: collectValue("gymSocioCodigo"),
      telefono: collectValue("gymSocioTelefono"),
      email: collectValue("gymSocioEmail"),
      plan_id: parseNumber("gymSocioPlan"),
      fecha_fin_plan: collectValue("gymSocioFechaFin"),
      objetivo: collectValue("gymSocioObjetivo"),
      saldo: parseNumber("gymSocioSaldo"),
      estado: collectValue("gymSocioEstado")
    };
  }

  function payloadPlan() {
    return {
      id: parseNumber("gymPlanId"),
      empresa_id: state.empresaID,
      nombre: collectValue("gymPlanNombre"),
      precio: parseNumber("gymPlanPrecio"),
      duracion_dias: parseNumber("gymPlanDias"),
      clases_incluidas: parseNumber("gymPlanClases"),
      sesiones_personalizadas: parseNumber("gymPlanSesiones"),
      acceso_ilimitado: byId("gymPlanIlimitado").checked,
      estado: collectValue("gymPlanEstado")
    };
  }

  function payloadTrainer() {
    return {
      id: parseNumber("gymTrainerId"),
      empresa_id: state.empresaID,
      nombre_completo: collectValue("gymTrainerNombre"),
      especialidad: collectValue("gymTrainerEspecialidad"),
      telefono: collectValue("gymTrainerTelefono"),
      email: collectValue("gymTrainerEmail"),
      disponibilidad: collectValue("gymTrainerDisponibilidad"),
      estado: collectValue("gymTrainerEstado")
    };
  }

  function payloadClass() {
    var fecha = collectValue("gymClassFecha");
    return {
      id: parseNumber("gymClassId"),
      empresa_id: state.empresaID,
      nombre: collectValue("gymClassNombre"),
      categoria: collectValue("gymClassCategoria"),
      entrenador_id: parseNumber("gymClassTrainer"),
      sede: collectValue("gymClassSede"),
      canal: collectValue("gymClassCanal"),
      fecha_programada: fecha ? fecha.replace("T", " ") + ":00" : "",
      cupos: parseNumber("gymClassCupos"),
      duracion_minutos: parseNumber("gymClassDuracion"),
      precio: parseNumber("gymClassPrecio")
    };
  }

  function bindForms() {
    byId("gymSocioForm").addEventListener("submit", function (event) {
      event.preventDefault();
      var payload = payloadSocio();
      var method = payload.id ? "PUT" : "POST";
      sendJSON(method, "socios", payload).then(refreshAllWithMessage.bind(null, "Socio guardado correctamente.")).catch(showError);
    });
    byId("gymPlanForm").addEventListener("submit", function (event) {
      event.preventDefault();
      var payload = payloadPlan();
      var method = payload.id ? "PUT" : "POST";
      sendJSON(method, "planes", payload).then(refreshAllWithMessage.bind(null, "Plan guardado correctamente.")).catch(showError);
    });
    byId("gymTrainerForm").addEventListener("submit", function (event) {
      event.preventDefault();
      var payload = payloadTrainer();
      var method = payload.id ? "PUT" : "POST";
      sendJSON(method, "entrenadores", payload).then(refreshAllWithMessage.bind(null, "Entrenador guardado correctamente.")).catch(showError);
    });
    byId("gymClassForm").addEventListener("submit", function (event) {
      event.preventDefault();
      var payload = payloadClass();
      var method = payload.id ? "PUT" : "POST";
      sendJSON(method, "clases", payload).then(refreshAllWithMessage.bind(null, "Clase guardada correctamente.")).catch(showError);
    });
    byId("gymEnrollmentForm").addEventListener("submit", function (event) {
      event.preventDefault();
      sendJSON("POST", "inscripciones", {
        empresa_id: state.empresaID,
        socio_id: parseNumber("gymEnrollmentSocio"),
        clase_id: parseNumber("gymEnrollmentClase"),
        estado: collectValue("gymEnrollmentEstado")
      }).then(refreshAllWithMessage.bind(null, "Inscripción registrada.")).catch(showError);
    });
    byId("gymAttendanceForm").addEventListener("submit", function (event) {
      event.preventDefault();
      var fecha = collectValue("gymAttendanceFecha");
      sendJSON("POST", "asistencias", {
        empresa_id: state.empresaID,
        socio_id: parseNumber("gymAttendanceSocio"),
        clase_id: parseNumber("gymAttendanceClase"),
        sede: collectValue("gymAttendanceSede"),
        canal: collectValue("gymAttendanceCanal"),
        tipo_acceso: collectValue("gymAttendanceTipo"),
        fecha_hora: fecha ? fecha.replace("T", " ") + ":00" : ""
      }).then(refreshAllWithMessage.bind(null, "Asistencia registrada.")).catch(showError);
    });
    byId("gymPaymentForm").addEventListener("submit", function (event) {
      event.preventDefault();
      sendJSON("POST", "pagos", {
        empresa_id: state.empresaID,
        socio_id: parseNumber("gymPaymentSocio"),
        plan_id: parseNumber("gymPaymentPlan"),
        concepto: collectValue("gymPaymentConcepto"),
        monto: parseNumber("gymPaymentMonto"),
        metodo_pago: collectValue("gymPaymentMetodo"),
        canal: collectValue("gymPaymentCanal")
      }).then(refreshAllWithMessage.bind(null, "Pago registrado.")).catch(showError);
    });
  }

  function showError(error) {
    setMessage(error && error.message ? error.message : "No se pudo completar la acción.", true);
  }

  function refreshAllWithMessage(message) {
    refreshAll().then(function () {
      setMessage(message, false);
      ["gymSocioForm", "gymPlanForm", "gymTrainerForm", "gymClassForm"].forEach(resetForm);
    }).catch(showError);
  }

  function refreshAll() {
    return Promise.all([
      fetchJSON("dashboard").then(renderDashboard),
      fetchJSON("socios").then(function (rows) { state.socios = rows || []; }),
      fetchJSON("planes").then(function (rows) { state.planes = rows || []; }),
      fetchJSON("entrenadores").then(function (rows) { state.entrenadores = rows || []; }),
      fetchJSON("clases").then(function (rows) { state.clases = rows || []; }),
      fetchJSON("inscripciones").then(function (rows) { state.inscripciones = rows || []; }),
      fetchJSON("asistencias").then(function (rows) { state.asistencias = rows || []; }),
      fetchJSON("pagos").then(function (rows) { state.pagos = rows || []; })
    ]).then(function () {
      syncSelects();
      renderTables();
    });
  }

  function bindStaticEvents() {
    Array.prototype.slice.call(document.querySelectorAll("[data-gym-tab]")).forEach(function (button) {
      button.addEventListener("click", function () {
        setTab(button.getAttribute("data-gym-tab"));
      });
    });
    Array.prototype.slice.call(document.querySelectorAll("[data-gym-reset]")).forEach(function (button) {
      button.addEventListener("click", function () {
        resetForm(button.getAttribute("data-gym-reset"));
      });
    });
    document.body.addEventListener("click", function (event) {
      var editKind = event.target.getAttribute("data-edit");
      var deleteAction = event.target.getAttribute("data-delete");
      var cancelEnrollment = event.target.getAttribute("data-cancel-enrollment");
      if (editKind) {
        loadForEdit(editKind, event.target.getAttribute("data-id"));
        return;
      }
      if (deleteAction) {
        sendJSON("DELETE", deleteAction, null, "id=" + encodeURIComponent(event.target.getAttribute("data-id"))).then(refreshAllWithMessage.bind(null, "Registro eliminado.")).catch(showError);
        return;
      }
      if (cancelEnrollment) {
        sendJSON("PUT", "cancelar_inscripcion", {}, "id=" + encodeURIComponent(cancelEnrollment)).then(refreshAllWithMessage.bind(null, "Inscripción cancelada.")).catch(showError);
      }
    });
  }

  function init() {
    state.empresaID = resolveEmpresaID();
    if (!state.empresaID) {
      setMessage("No se encontró la empresa activa para el módulo de gimnasio.", true);
      return;
    }
    bindForms();
    bindStaticEvents();
    refreshAll().catch(showError);
  }

  init();
})();
