(function () {
  "use strict";

  var root = null;
  var state = { empresaId: "", modulo: "", titulo: "", lead: "", dashboard: null, reporte: null, agenda: null, sla: null, riesgo: null, responsables: [], expediente: null, evidencias: [], aprobaciones: [], tareas: [], plantilla: null, busqueda: null, selectedRegistroID: 0 };
  var money = new Intl.NumberFormat("es-CO", { style: "currency", currency: "COP", maximumFractionDigits: 0 });

  function esc(value) {
    return String(value == null ? "" : value).replace(/[&<>"']/g, function (ch) {
      return { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[ch];
    });
  }

  function resolveEmpresaId() {
    var sources = [window.location.search || ""];
    try {
      if (window.parent && window.parent !== window) sources.push(window.parent.location.search || "");
    } catch (_) {}
    for (var i = 0; i < sources.length; i += 1) {
      var id = new URLSearchParams(sources[i]).get("empresa_id");
      if (id) return id;
    }
    try {
      return sessionStorage.getItem("empresa_id") || localStorage.getItem("empresa_id") || "";
    } catch (_) {
      return "";
    }
  }

  function api(action, options) {
    if (!state.empresaId) return Promise.reject(new Error("empresa_id no disponible"));
    var actionName = action || "dashboard";
    var extra = "";
    if (String(actionName).indexOf("&") >= 0) {
      var parts = String(actionName).split("&");
      actionName = parts.shift();
      extra = "&" + parts.join("&");
    }
    var url = "/api/empresa/" + encodeURIComponent(state.modulo) + "?empresa_id=" + encodeURIComponent(state.empresaId) + "&action=" + encodeURIComponent(actionName) + extra;
    return fetch(url, Object.assign({ credentials: "same-origin", headers: { "Content-Type": "application/json" } }, options || {})).then(function (res) {
      if (!res.ok) {
        return res.text().then(function (text) {
          throw new Error(text || res.statusText);
        });
      }
      return res.json();
    });
  }

  function chip(value, tone) {
    return '<span class="mc-chip ' + esc(tone || "") + '">' + esc(String(value || "-").replace(/_/g, " ")) + "</span>";
  }

  function setMsg(text, isError) {
    var node = document.getElementById("mcMsg");
    if (!node) return;
    node.textContent = text || "";
    node.classList.toggle("is-error", !!isError);
  }

  function renderShell() {
    root.innerHTML =
      '<main class="mc-shell">' +
      '<header class="mc-head"><div><h1>' + esc(state.titulo) + '</h1><p>' + esc(state.lead) + '</p></div>' +
      '<div class="mc-actions"><button class="mc-btn" id="mcRefresh" type="button">Actualizar</button><button class="mc-btn" id="mcSeed" type="button">Cargar demo</button><button class="mc-btn" id="mcImportBtn" type="button">Importar CSV</button><button class="mc-btn primary" id="mcExport" type="button">Exportar auditoria CSV</button><input id="mcImportFile" type="file" accept=".csv,text/csv" hidden></div></header>' +
      '<section class="mc-kpis"><article><span>Total</span><strong id="kpiTotal">0</strong></article><article><span>Abiertos</span><strong id="kpiAbiertos">0</strong></article><article><span>En proceso</span><strong id="kpiProceso">0</strong></article><article><span>Aprobados / cerrados</span><strong id="kpiAprobados">0</strong></article><article><span>Vencidos</span><strong id="kpiVencidos">0</strong></article><article><span>Valor</span><strong id="kpiValor">$0</strong></article></section>' +
      '<div class="mc-toolbar mc-searchbar"><label>Buscar<input id="mcTextoFiltro" placeholder="Codigo, nombre, tercero, referencia"></label><label>Estado<select id="mcEstadoFiltro"><option value="">Todos</option></select></label><label>Tipo<select id="mcTipoFiltro"><option value="">Todos</option></select></label><label>Categoria<select id="mcCategoriaFiltro"><option value="">Todas</option></select></label><label>Prioridad<select id="mcPrioridadFiltro"><option value="">Todas</option><option value="baja">Baja</option><option value="normal">Normal</option><option value="alta">Alta</option><option value="critica">Critica</option><option value="urgente">Urgente</option></select></label><label>Responsable<input id="mcResponsableFiltro" placeholder="Usuario, rol o area"></label><label>Vencimiento<select id="mcVenceFiltro"><option value="">Todos</option><option value="vencidos">Vencidos</option><option value="7">Proximos 7 dias</option><option value="30">Proximos 30 dias</option></select></label><button class="mc-btn primary" id="mcSearch" type="button">Buscar</button><button class="mc-btn" id="mcClearSearch" type="button">Limpiar</button><button class="mc-btn" id="mcClearForm" type="button">Nuevo registro</button></div>' +
      '<div id="mcMsg" class="mc-msg"></div>' +
      '<section class="mc-card"><h2>Agenda y alertas</h2><div id="mcAgenda" class="mc-agenda"></div></section>' +
      '<section class="mc-card"><h2>SLA y cumplimiento</h2><div id="mcSLA" class="mc-sla"></div></section>' +
      '<section class="mc-card"><h2>Matriz de riesgo operativo</h2><div id="mcRiesgo" class="mc-riesgo"></div></section>' +
      '<section class="mc-card"><h2>Responsables y carga</h2><div id="mcResponsables" class="mc-responsables"></div></section>' +
      '<section class="mc-card"><h2>Reporte ejecutivo</h2><div id="mcReporte" class="mc-report"></div></section>' +
      '<section class="mc-card"><h2>Expediente 360</h2><div id="mcExpediente" class="mc-expediente"><div class="mc-empty">Selecciona Expediente en un registro para ver su trazabilidad completa.</div></div></section>' +
      '<section class="mc-grid"><form id="mcForm" class="mc-card mc-form">' +
      '<h2>Registro operativo</h2><input type="hidden" id="mcId">' +
      '<div class="mc-row"><label>Codigo<input id="mcCodigo" required placeholder="AUTO-001"></label><label>Tipo<select id="mcTipo"></select></label></div>' +
      '<label>Nombre<input id="mcNombre" required></label>' +
      '<div class="mc-row"><label><span id="mcTerceroLabel">Tercero / area</span><input id="mcTercero"></label><label>Responsable<input id="mcResponsable"></label></div>' +
      '<div class="mc-row"><label>Categoria<select id="mcCategoria"></select></label><label><span id="mcReferenciaLabel">Referencia</span><input id="mcReferencia"></label></div>' +
      '<div class="mc-row"><label>Prioridad<select id="mcPrioridad"><option value="normal">Normal</option><option value="baja">Baja</option><option value="alta">Alta</option><option value="critica">Critica</option><option value="urgente">Urgente</option></select></label><label>Estado<select id="mcEstado"></select></label></div>' +
      '<div class="mc-row"><label>Fecha<input id="mcFecha" type="date"></label><label>Vencimiento<input id="mcVence" type="date"></label></div>' +
      '<label>Valor<input id="mcValor" type="number" step="0.01" value="0"></label>' +
      '<label>Metadata JSON<textarea id="mcMetadata" placeholder=\'{"nota":"detalle"}\'></textarea></label>' +
      '<button class="mc-btn primary" type="submit">Guardar</button></form>' +
      '<div class="mc-card"><h2>Registros</h2><div class="mc-bulk"><strong id="mcBulkCount">0 seleccionados</strong><label>Estado<select id="mcBulkEstado"><option value="">Sin cambio</option></select></label><label>Prioridad<select id="mcBulkPrioridad"><option value="">Sin cambio</option><option value="baja">Baja</option><option value="normal">Normal</option><option value="alta">Alta</option><option value="critica">Critica</option><option value="urgente">Urgente</option></select></label><label>Responsable<input id="mcBulkResponsable" placeholder="Asignar responsable"></label><label>Detalle<input id="mcBulkDetalle" placeholder="Motivo de accion masiva"></label><button class="mc-btn primary" id="mcBulkApply" type="button">Aplicar</button></div><div id="mcTable" class="mc-table-wrap"></div></div></section>' +
      '<section class="mc-grid mc-grid-secondary"><div class="mc-card"><h2>Seguimiento profesional</h2><form id="mcFollowForm" class="mc-form"><input type="hidden" id="mcFollowId"><div class="mc-row"><label>Registro<select id="mcFollowRegistro"></select></label><label>Accion<select id="mcFollowEvento"></select></label></div><div class="mc-row"><label>Cambiar estado<select id="mcFollowEstado"></select></label><label>Detalle<input id="mcFollowDetalle" placeholder="Gestion realizada, evidencia o decision"></label></div><button class="mc-btn primary" type="submit">Registrar seguimiento</button></form><div id="mcAlerts"></div></div><div class="mc-card"><h2>Evidencias y soportes</h2><form id="mcEvidenceForm" class="mc-form"><div class="mc-row"><label>Registro<select id="mcEvidenceRegistro"></select></label><label>Tipo<select id="mcEvidenceTipo"><option value="soporte">Soporte</option><option value="contrato">Contrato</option><option value="foto">Foto</option><option value="acta">Acta</option><option value="documento">Documento</option><option value="enlace">Enlace</option></select></label></div><label>Nombre<input id="mcEvidenceNombre" required placeholder="Nombre del soporte"></label><label>URL o ruta<input id="mcEvidenceUrl" placeholder="https://... o ruta interna"></label><label>Descripcion<input id="mcEvidenceDesc" placeholder="Detalle breve"></label><button class="mc-btn primary" type="submit">Agregar evidencia</button></form><div id="mcEvidenceList"></div></div><div class="mc-card"><h2>Aprobaciones</h2><form id="mcApprovalForm" class="mc-form"><div class="mc-row"><label>Registro<select id="mcApprovalRegistro"></select></label><label>Nivel<select id="mcApprovalNivel"><option value="operativo">Operativo</option><option value="supervisor">Supervisor</option><option value="contable">Contable</option><option value="gerencia">Gerencia</option><option value="cumplimiento">Cumplimiento</option><option value="juridico">Juridico</option></select></label></div><div class="mc-row"><label>Solicitado a<input id="mcApprovalTo" required placeholder="usuario, rol o area"></label><label>Vence<input id="mcApprovalDue" type="date"></label></div><label>Comentario<input id="mcApprovalComment" placeholder="Motivo o instruccion"></label><button class="mc-btn primary" type="submit">Solicitar aprobacion</button></form><div id="mcApprovalList"></div></div><div class="mc-card"><h2>Tareas y compromisos</h2><form id="mcTaskForm" class="mc-form"><div class="mc-row"><label>Registro<select id="mcTaskRegistro"></select></label><label>Prioridad<select id="mcTaskPrioridad"><option value="normal">Normal</option><option value="baja">Baja</option><option value="alta">Alta</option><option value="critica">Critica</option><option value="urgente">Urgente</option></select></label></div><div class="mc-row"><label>Responsable<input id="mcTaskResponsable" placeholder="Usuario, rol o area"></label><label>Vence<input id="mcTaskDue" type="date"></label></div><label>Tarea<input id="mcTaskTitulo" required placeholder="Compromiso operativo"></label><label>Comentario<input id="mcTaskComment" placeholder="Detalle breve"></label><button class="mc-btn primary" type="submit">Crear tarea</button></form><div id="mcTaskList"></div></div><div class="mc-card"><h2>Bitacora</h2><div id="mcEvents"></div></div></section></main>';
  }

  function render() {
    var d = state.dashboard || {};
    document.getElementById("kpiTotal").textContent = d.total_registros || 0;
    document.getElementById("kpiAbiertos").textContent = d.abiertos || 0;
    document.getElementById("kpiProceso").textContent = d.en_proceso || 0;
    document.getElementById("kpiAprobados").textContent = d.aprobados || 0;
    document.getElementById("kpiVencidos").textContent = d.vencidos || 0;
    document.getElementById("kpiValor").textContent = money.format(d.valor_total || 0);
    renderTable(d.registros || []);
    renderFollowOptions(d.registros || []);
    renderAlerts(d.alertas || []);
    renderEvents(d.eventos_recientes || []);
    renderAgenda(state.agenda || {});
    renderSLA(state.sla || {});
    renderRiesgo(state.riesgo || {});
    renderResponsables(state.responsables || []);
    renderReporte(state.reporte || {});
    renderExpediente(state.expediente);
    renderEvidencias(state.evidencias || []);
    renderAprobaciones(state.aprobaciones || []);
    renderTareas(state.tareas || []);
    scrollToRequestedSection();
  }

  function scrollToRequestedSection() {
    var hash = "";
    try { hash = window.location.hash || ""; } catch (_) { hash = ""; }
    if (!hash || hash.length < 2) return;
    var id = "";
    try { id = decodeURIComponent(hash.slice(1)); } catch (_) { id = hash.slice(1); }
    var target = document.getElementById(id);
    if (!target) return;
    window.setTimeout(function () {
      try { target.scrollIntoView({ behavior: "smooth", block: "start" }); } catch (_) { target.scrollIntoView(); }
    }, 80);
  }

  function optionList(values, selected) {
    return (values || []).map(function (value) {
      var text = String(value || "");
      return '<option value="' + esc(text) + '"' + (text === selected ? " selected" : "") + ">" + esc(text.replace(/_/g, " ")) + "</option>";
    }).join("");
  }

  function applyPlantilla() {
    var p = state.plantilla || {};
    var tipos = p.tipos && p.tipos.length ? p.tipos : ["general", "seguimiento", "aprobacion"];
    var categorias = p.categorias && p.categorias.length ? p.categorias : ["general", "operacion", "finanzas"];
    var estados = p.estados_flujo && p.estados_flujo.length ? p.estados_flujo : ["pendiente", "en_proceso", "aprobado", "cerrado"];
    var acciones = p.acciones_sugeridas && p.acciones_sugeridas.length ? p.acciones_sugeridas : ["seguimiento", "comentario", "aprobacion", "cierre"];
    document.getElementById("mcTipo").innerHTML = optionList(tipos);
    document.getElementById("mcCategoria").innerHTML = optionList(categorias);
    document.getElementById("mcEstado").innerHTML = optionList(estados);
    document.getElementById("mcFollowEstado").innerHTML = '<option value="">No cambiar</option>' + optionList(estados);
    document.getElementById("mcBulkEstado").innerHTML = '<option value="">Sin cambio</option>' + optionList(estados);
    document.getElementById("mcFollowEvento").innerHTML = optionList(acciones);
    document.getElementById("mcEstadoFiltro").innerHTML = '<option value="">Todos</option>' + optionList(estados);
    document.getElementById("mcTipoFiltro").innerHTML = '<option value="">Todos</option>' + optionList(tipos);
    document.getElementById("mcCategoriaFiltro").innerHTML = '<option value="">Todas</option>' + optionList(categorias);
    document.getElementById("mcTerceroLabel").textContent = p.etiqueta_tercero || "Tercero / area";
    document.getElementById("mcReferenciaLabel").textContent = p.etiqueta_referencia || "Referencia";
    document.getElementById("mcMetadata").placeholder = p.metadata_ejemplo || '{"nota":"detalle"}';
  }

  function renderTable(rows) {
    var host = document.getElementById("mcTable");
    if (!rows.length) {
      host.innerHTML = '<div class="mc-empty">Sin registros. Puedes cargar datos demo o crear el primer registro.</div>';
      updateBulkCount();
      return;
    }
    host.innerHTML = '<table class="mc-table"><thead><tr><th><input id="mcSelectAll" type="checkbox" aria-label="Seleccionar todos"></th><th>Codigo</th><th>Nombre</th><th>Tipo</th><th>Tercero</th><th>Responsable</th><th>Estado</th><th>Vence</th><th>Valor</th><th></th></tr></thead><tbody>' +
      rows.map(function (row) {
        var vencido = row.fecha_vencimiento && row.fecha_vencimiento < new Date().toISOString().slice(0, 10) && ["cerrado", "cancelado", "cumplido", "pagado", "resuelto"].indexOf(row.estado) < 0;
        return '<tr><td><input class="mc-row-select" type="checkbox" value="' + esc(row.id) + '" aria-label="Seleccionar registro"></td><td><strong>' + esc(row.codigo) + '</strong><br><span>' + esc(row.categoria || row.referencia || "") + '</span></td><td>' + esc(row.nombre) + '</td><td>' + chip(row.tipo) + '</td><td>' + esc(row.tercero || "-") + '</td><td>' + esc(row.responsable || "-") + '</td><td>' + chip(row.estado, vencido ? "bad" : "") + '</td><td>' + esc(row.fecha_vencimiento || "-") + '</td><td class="num">' + money.format(row.valor || 0) + '</td><td><button class="mc-icon-btn" data-edit="' + esc(row.id) + '" type="button">Editar</button> <button class="mc-icon-btn" data-follow="' + esc(row.id) + '" type="button">Seguimiento</button> <button class="mc-icon-btn" data-expediente="' + esc(row.id) + '" type="button">Expediente</button></td></tr>';
      }).join("") + "</tbody></table>";
    updateBulkCount();
  }

  function renderFollowOptions(rows) {
    var html = rows.map(function (row) {
      return '<option value="' + esc(row.id) + '">' + esc(row.codigo + " - " + row.nombre) + "</option>";
    }).join("");
    document.getElementById("mcFollowRegistro").innerHTML = html || '<option value="">Sin registros</option>';
    document.getElementById("mcEvidenceRegistro").innerHTML = html || '<option value="">Sin registros</option>';
    document.getElementById("mcApprovalRegistro").innerHTML = html || '<option value="">Sin registros</option>';
    document.getElementById("mcTaskRegistro").innerHTML = html || '<option value="">Sin registros</option>';
    if (state.selectedRegistroID) document.getElementById("mcFollowRegistro").value = String(state.selectedRegistroID);
    if (state.selectedRegistroID) document.getElementById("mcEvidenceRegistro").value = String(state.selectedRegistroID);
    if (state.selectedRegistroID) document.getElementById("mcApprovalRegistro").value = String(state.selectedRegistroID);
    if (state.selectedRegistroID) document.getElementById("mcTaskRegistro").value = String(state.selectedRegistroID);
  }

  function renderAlerts(rows) {
    document.getElementById("mcAlerts").innerHTML = rows.map(function (row) {
      return '<div class="mc-alert">' + esc(row) + "</div>";
    }).join("") || '<div class="mc-empty">Sin alertas.</div>';
  }

  function renderEvents(rows) {
    document.getElementById("mcEvents").innerHTML = rows.map(function (row) {
      return '<div class="mc-event"><strong>' + esc(row.evento) + '</strong><span>' + esc([row.detalle, row.usuario, row.fecha_creacion].filter(Boolean).join(" - ")) + "</span></div>";
    }).join("") || '<div class="mc-empty">Sin bitacora.</div>';
  }

  function renderAgenda(agenda) {
    var host = document.getElementById("mcAgenda");
    if (!host) return;
    var items = agenda.items || [];
    var summary =
      '<div class="mc-agenda-summary">' +
      '<article><span>Alertas</span><strong>' + esc(agenda.total_alertas || 0) + '</strong></article>' +
      '<article><span>Registros vencidos</span><strong>' + esc(agenda.registros_vencidos || 0) + '</strong></article>' +
      '<article><span>Tareas vencidas</span><strong>' + esc(agenda.tareas_vencidas || 0) + '</strong></article>' +
      '<article><span>Aprob. pendientes</span><strong>' + esc(agenda.aprobaciones_pendientes || 0) + '</strong></article>' +
      '</div>';
    var recs = (agenda.recomendaciones || []).map(function (row) {
      return '<div class="mc-alert">' + esc(row) + "</div>";
    }).join("");
    var list = items.slice(0, 12).map(function (row) {
      return '<div class="mc-agenda-item ' + esc(row.severidad || "") + '"><strong>' + esc(row.titulo) + " " + chip(row.tipo) + '</strong><span>' + esc([row.detalle, row.responsable, row.estado, row.fecha_vencimiento].filter(Boolean).join(" - ")) + '</span><button class="mc-icon-btn" data-expediente="' + esc(row.registro_id) + '" type="button">Expediente</button></div>';
    }).join("") || '<div class="mc-empty">Agenda sin alertas.</div>';
    host.innerHTML = summary + '<div class="mc-actions"><button class="mc-btn primary" id="mcPlanAccion" type="button">Generar plan de accion</button></div><div class="mc-agenda-grid"><div>' + (recs || '<div class="mc-empty">Sin recomendaciones.</div>') + '</div><div>' + list + '</div></div>';
  }

  function renderResponsables(rows) {
    var host = document.getElementById("mcResponsables");
    if (!host) return;
    host.innerHTML = (rows || []).map(function (row) {
      return '<div class="mc-responsable"><strong>' + esc(row.responsable) + '</strong><div class="mc-responsable-kpis"><span>Registros ' + esc(row.registros_abiertos || 0) + '</span><span>Tareas ' + esc(row.tareas_abiertas || 0) + '</span><span>Aprob. ' + esc(row.aprobaciones_pendientes || 0) + '</span><span>Total ' + esc(row.total_pendiente || 0) + '</span></div><em>' + esc(row.recomendacion || "") + '</em></div>';
    }).join("") || '<div class="mc-empty">Sin responsables con carga pendiente.</div>';
  }

  function renderSLA(sla) {
    var host = document.getElementById("mcSLA");
    if (!host) return;
    var buckets = sla.buckets || {};
    var pct = Number(sla.cumplimiento_pct == null ? 100 : sla.cumplimiento_pct);
    var recs = (sla.recomendaciones || []).map(function (row) {
      return '<div class="mc-alert">' + esc(row) + "</div>";
    }).join("") || '<div class="mc-empty">Sin recomendaciones.</div>';
    host.innerHTML =
      '<div class="mc-sla-head ' + esc(sla.semaforo || "verde") + '"><strong>' + esc(pct.toFixed(1)) + '%</strong><span>' + esc(sla.semaforo || "verde") + '</span></div>' +
      '<div class="mc-sla-grid">' +
      '<article><span>Abiertos</span><strong>' + esc(sla.total_abiertos || 0) + '</strong></article>' +
      '<article><span>Vencidos</span><strong>' + esc(sla.vencidos || 0) + '</strong></article>' +
      '<article><span>Próx. 7 días</span><strong>' + esc(sla.proximos_7_dias || 0) + '</strong></article>' +
      '<article><span>Sin SLA</span><strong>' + esc(sla.sin_vencimiento || 0) + '</strong></article>' +
      '<article><span>Tareas abiertas</span><strong>' + esc(sla.tareas_abiertas || 0) + '</strong></article>' +
      '<article><span>Tareas vencidas</span><strong>' + esc(sla.tareas_vencidas || 0) + '</strong></article>' +
      '</div><div class="mc-sla-buckets"><span>Vencido ' + esc(buckets.vencido || 0) + '</span><span>0-7 ' + esc(buckets["0_7"] || 0) + '</span><span>8-30 ' + esc(buckets["8_30"] || 0) + '</span><span>+30 ' + esc(buckets.mas_30 || 0) + '</span></div>' + recs;
  }

  function renderRiesgo(riesgo) {
    var host = document.getElementById("mcRiesgo");
    if (!host) return;
    var recs = (riesgo.recomendaciones || []).map(function (row) {
      return '<div class="mc-alert">' + esc(row) + "</div>";
    }).join("") || '<div class="mc-empty">Sin recomendaciones.</div>';
    var factores = (riesgo.factores || []).map(function (row) {
      return '<span>' + esc(row) + "</span>";
    }).join("") || '<span>Sin factores criticos detectados</span>';
    host.innerHTML =
      '<div class="mc-risk-head ' + esc(riesgo.nivel || "bajo") + '"><strong>' + esc(riesgo.score || 0) + '/100</strong><span>' + esc(riesgo.nivel || "bajo") + '</span></div>' +
      '<div class="mc-risk-grid"><article><span>Vencidos</span><strong>' + esc(riesgo.registros_vencidos || 0) + '</strong></article><article><span>Críticos</span><strong>' + esc(riesgo.criticos_abiertos || 0) + '</strong></article><article><span>Sin responsable</span><strong>' + esc(riesgo.sin_responsable || 0) + '</strong></article><article><span>Sin evidencia</span><strong>' + esc(riesgo.sin_evidencia || 0) + '</strong></article><article><span>Aprobaciones</span><strong>' + esc(riesgo.aprobaciones_pendientes || 0) + '</strong></article><article><span>Tareas abiertas</span><strong>' + esc(riesgo.tareas_abiertas || 0) + '</strong></article></div>' +
      '<div class="mc-risk-factors">' + factores + '</div>' + recs;
  }

  function renderEvidencias(rows) {
    document.getElementById("mcEvidenceList").innerHTML = (rows || []).map(function (row) {
      var link = row.url ? '<a href="' + esc(row.url) + '" target="_blank" rel="noopener">Abrir</a>' : "";
      return '<div class="mc-evidence"><strong>' + esc(row.nombre) + '</strong><span>' + esc([row.tipo, row.descripcion, row.usuario, row.fecha_creacion].filter(Boolean).join(" - ")) + '</span>' + link + "</div>";
    }).join("") || '<div class="mc-empty">Sin evidencias registradas.</div>';
  }

  function renderAprobaciones(rows) {
    document.getElementById("mcApprovalList").innerHTML = (rows || []).map(function (row) {
      var actions = row.estado === "pendiente" ? '<div class="mc-approval-actions"><button class="mc-icon-btn" data-approve="' + esc(row.id) + '" type="button">Aprobar</button><button class="mc-icon-btn" data-reject="' + esc(row.id) + '" type="button">Rechazar</button></div>' : "";
      return '<div class="mc-approval"><strong>' + esc(row.solicitado_a) + " " + chip(row.estado) + '</strong><span>' + esc([row.nivel, row.comentario, row.solicitado_por, row.fecha_vencimiento].filter(Boolean).join(" - ")) + "</span>" + actions + "</div>";
    }).join("") || '<div class="mc-empty">Sin aprobaciones registradas.</div>';
  }

  function renderTareas(rows) {
    document.getElementById("mcTaskList").innerHTML = (rows || []).map(function (row) {
      var pending = ["pendiente", "en_proceso"].indexOf(row.estado) >= 0;
      var actions = pending ? '<div class="mc-task-actions"><button class="mc-icon-btn" data-task-progress="' + esc(row.id) + '" type="button">En proceso</button><button class="mc-icon-btn" data-task-done="' + esc(row.id) + '" type="button">Cumplida</button><button class="mc-icon-btn" data-task-cancel="' + esc(row.id) + '" type="button">Cancelar</button></div>' : "";
      return '<div class="mc-task"><strong>' + esc(row.titulo) + " " + chip(row.estado) + '</strong><span>' + esc([row.responsable, row.prioridad, row.fecha_vencimiento, row.comentario].filter(Boolean).join(" - ")) + "</span>" + actions + "</div>";
    }).join("") || '<div class="mc-empty">Sin tareas registradas.</div>';
  }

  function metricList(title, rows) {
    rows = rows || [];
    return '<div class="mc-metric-list"><strong>' + esc(title) + '</strong>' + (rows.length ? rows.map(function (row) {
      return '<span><b>' + esc(String(row.clave || "sin_dato").replace(/_/g, " ")) + '</b><em>' + esc(row.total || 0) + ' / ' + money.format(row.valor || 0) + '</em></span>';
    }).join("") : '<span class="mc-muted">Sin datos</span>') + "</div>";
  }

  function renderReporte(rep) {
    var host = document.getElementById("mcReporte");
    if (!host) return;
    var resumen =
      '<div class="mc-report-summary">' +
      '<article><span>Vencidos</span><strong>' + esc(rep.vencidos || 0) + '</strong></article>' +
      '<article><span>Vencen 7 dias</span><strong>' + esc(rep.vencen_7_dias || 0) + '</strong></article>' +
      '<article><span>Vencen 30 dias</span><strong>' + esc(rep.vencen_30_dias || 0) + '</strong></article>' +
      '<article><span>Criticos abiertos</span><strong>' + esc(rep.criticos_abiertos || 0) + '</strong></article>' +
      '<article><span>Sin responsable</span><strong>' + esc(rep.sin_responsable || 0) + '</strong></article>' +
      '<article><span>Valor pendiente</span><strong>' + esc(money.format(rep.valor_pendiente || 0)) + '</strong></article>' +
      '</div>';
    var recomendaciones = (rep.recomendaciones || []).map(function (item) {
      return '<div class="mc-alert">' + esc(item) + "</div>";
    }).join("") || '<div class="mc-empty">Sin recomendaciones.</div>';
    host.innerHTML = resumen + '<div class="mc-report-grid">' +
      metricList("Por estado", rep.por_estado) +
      metricList("Por tipo", rep.por_tipo) +
      metricList("Por categoria", rep.por_categoria) +
      metricList("Por prioridad", rep.por_prioridad) +
      '<div class="mc-report-recommend"><strong>Recomendaciones</strong>' + recomendaciones + '</div>' +
      '</div>';
  }

  function miniList(title, rows, mapper) {
    rows = rows || [];
    return '<div class="mc-exp-block"><strong>' + esc(title) + '</strong>' + (rows.length ? rows.slice(0, 8).map(mapper).join("") : '<span class="mc-muted">Sin datos</span>') + "</div>";
  }

  function renderExpediente(exp) {
    var host = document.getElementById("mcExpediente");
    if (!host) return;
    if (!exp || !exp.registro) {
      host.innerHTML = '<div class="mc-empty">Selecciona Expediente en un registro para ver su trazabilidad completa.</div>';
      return;
    }
    var r = exp.registro;
    var resumen = exp.resumen || {};
    host.innerHTML =
      '<div class="mc-exp-head"><div><strong>' + esc(r.codigo + " - " + r.nombre) + '</strong><span>' + esc([r.tipo, r.categoria, r.tercero, r.responsable].filter(Boolean).join(" - ")) + '</span></div><div class="mc-exp-actions">' + chip(r.estado) + '<button class="mc-icon-btn" data-close-controlled="' + esc(r.id) + '" type="button">Cerrar validado</button></div></div>' +
      '<div class="mc-exp-summary"><article><span>Eventos</span><strong>' + esc(resumen.eventos || 0) + '</strong></article><article><span>Evidencias</span><strong>' + esc(resumen.evidencias || 0) + '</strong></article><article><span>Aprob. pendientes</span><strong>' + esc(resumen.aprobaciones_pendientes || 0) + '</strong></article><article><span>Tareas abiertas</span><strong>' + esc(resumen.tareas_abiertas || 0) + '</strong></article></div>' +
      '<div class="mc-alert">' + esc(exp.recomendacion || "Expediente listo para seguimiento.") + '</div>' +
      '<div class="mc-exp-grid">' +
      miniList("Eventos", exp.eventos, function (row) { return '<span><b>' + esc(row.evento) + '</b><em>' + esc([row.detalle, row.fecha_creacion].filter(Boolean).join(" - ")) + '</em></span>'; }) +
      miniList("Evidencias", exp.evidencias, function (row) { return '<span><b>' + esc(row.nombre) + '</b><em>' + esc([row.tipo, row.fecha_creacion].filter(Boolean).join(" - ")) + '</em></span>'; }) +
      miniList("Aprobaciones", exp.aprobaciones, function (row) { return '<span><b>' + esc(row.solicitado_a) + '</b><em>' + esc([row.estado, row.nivel].filter(Boolean).join(" - ")) + '</em></span>'; }) +
      miniList("Tareas", exp.tareas, function (row) { return '<span><b>' + esc(row.titulo) + '</b><em>' + esc([row.estado, row.responsable].filter(Boolean).join(" - ")) + '</em></span>'; }) +
      '</div>';
  }

  function formPayload() {
    var metadata = {};
    var raw = document.getElementById("mcMetadata").value.trim();
    if (raw) metadata = JSON.parse(raw);
    return {
      id: Number(document.getElementById("mcId").value || 0),
      codigo: document.getElementById("mcCodigo").value,
      tipo: document.getElementById("mcTipo").value,
      nombre: document.getElementById("mcNombre").value,
      tercero: document.getElementById("mcTercero").value,
      responsable: document.getElementById("mcResponsable").value,
      categoria: document.getElementById("mcCategoria").value,
      referencia: document.getElementById("mcReferencia").value,
      prioridad: document.getElementById("mcPrioridad").value,
      estado: document.getElementById("mcEstado").value,
      fecha: document.getElementById("mcFecha").value,
      fecha_vencimiento: document.getElementById("mcVence").value,
      valor: Number(document.getElementById("mcValor").value || 0),
      metadata: metadata
    };
  }

  function editRow(id) {
    var rows = (state.dashboard && state.dashboard.registros) || [];
    var row = rows.find(function (item) { return String(item.id) === String(id); });
    if (!row) return;
    document.getElementById("mcId").value = row.id || 0;
    document.getElementById("mcCodigo").value = row.codigo || "";
    document.getElementById("mcTipo").value = row.tipo || "general";
    document.getElementById("mcNombre").value = row.nombre || "";
    document.getElementById("mcTercero").value = row.tercero || "";
    document.getElementById("mcResponsable").value = row.responsable || "";
    document.getElementById("mcCategoria").value = row.categoria || "";
    document.getElementById("mcReferencia").value = row.referencia || "";
    document.getElementById("mcPrioridad").value = row.prioridad || "normal";
    document.getElementById("mcEstado").value = row.estado || "pendiente";
    document.getElementById("mcFecha").value = row.fecha || "";
    document.getElementById("mcVence").value = row.fecha_vencimiento || "";
    document.getElementById("mcValor").value = row.valor || 0;
    document.getElementById("mcMetadata").value = JSON.stringify(row.metadata || {}, null, 2);
    state.selectedRegistroID = row.id || 0;
    var follow = document.getElementById("mcFollowRegistro");
    if (follow) follow.value = String(state.selectedRegistroID);
    var evidence = document.getElementById("mcEvidenceRegistro");
    if (evidence) evidence.value = String(state.selectedRegistroID);
    var approval = document.getElementById("mcApprovalRegistro");
    if (approval) approval.value = String(state.selectedRegistroID);
    var task = document.getElementById("mcTaskRegistro");
    if (task) task.value = String(state.selectedRegistroID);
  }

  function buildSearchQuery() {
    var params = new URLSearchParams();
    var texto = document.getElementById("mcTextoFiltro").value.trim();
    var estado = document.getElementById("mcEstadoFiltro").value;
    var tipo = document.getElementById("mcTipoFiltro").value;
    var categoria = document.getElementById("mcCategoriaFiltro").value;
    var prioridad = document.getElementById("mcPrioridadFiltro").value;
    var responsable = document.getElementById("mcResponsableFiltro").value.trim();
    var vence = document.getElementById("mcVenceFiltro").value;
    if (texto) params.set("texto", texto);
    if (estado) params.set("estado", estado);
    if (tipo) params.set("tipo", tipo);
    if (categoria) params.set("categoria", categoria);
    if (prioridad) params.set("prioridad", prioridad);
    if (responsable) params.set("responsable", responsable);
    if (vence === "vencidos") params.set("vencidos", "true");
    if (vence === "7" || vence === "30") params.set("proximos_dias", vence);
    return params.toString();
  }

  function hasSearchFilters() {
    return !!buildSearchQuery();
  }

  function clearSearchFilters() {
    ["mcTextoFiltro", "mcResponsableFiltro"].forEach(function (id) {
      document.getElementById(id).value = "";
    });
    ["mcEstadoFiltro", "mcTipoFiltro", "mcCategoriaFiltro", "mcPrioridadFiltro", "mcVenceFiltro"].forEach(function (id) {
      document.getElementById(id).value = "";
    });
    state.busqueda = null;
    load();
  }

  function load() {
    setMsg("Cargando...");
    var searchQuery = buildSearchQuery();
    var searchRequest = searchQuery ? api("buscar&" + searchQuery) : Promise.resolve(null);
    Promise.all([api("dashboard"), api("reporte"), api("agenda"), api("sla"), api("riesgo"), api("responsables"), api("evidencias"), api("aprobaciones"), api("tareas"), searchRequest]).then(function (results) {
      state.dashboard = results[0] || {};
      state.reporte = results[1] || {};
      state.agenda = results[2] || {};
      state.sla = results[3] || {};
      state.riesgo = results[4] || {};
      state.responsables = results[5] || [];
      state.evidencias = results[6] || [];
      state.aprobaciones = results[7] || [];
      state.tareas = results[8] || [];
      state.busqueda = results[9] || null;
      if (state.busqueda) {
        state.dashboard.registros = state.busqueda.registros || [];
      }
      render();
      setMsg(state.busqueda ? "Busqueda aplicada: " + (state.busqueda.total || 0) + " registro(s)." : "Actualizado");
    }).catch(function (err) {
      setMsg(err.message, true);
    });
  }

  function loadExpediente(registroID) {
    if (!registroID) return;
    state.selectedRegistroID = Number(registroID || 0);
    api("expediente&registro_id=" + encodeURIComponent(state.selectedRegistroID)).then(function (data) {
      state.expediente = data || null;
      renderExpediente(state.expediente);
      setMsg("Expediente actualizado");
    }).catch(function (err) { setMsg(err.message, true); });
  }

  function saveFollow(ev) {
    ev.preventDefault();
    var registroID = Number(document.getElementById("mcFollowRegistro").value || 0);
    var estado = document.getElementById("mcFollowEstado").value;
    var detalle = document.getElementById("mcFollowDetalle").value.trim();
    var evento = document.getElementById("mcFollowEvento").value;
    if (!registroID) {
      setMsg("Selecciona un registro para seguimiento.", true);
      return;
    }
    var request = estado
      ? api("estado", { method: "POST", body: JSON.stringify({ registro_id: registroID, estado: estado, detalle: detalle || "Cambio de estado" }) })
      : api("evento", { method: "POST", body: JSON.stringify({ registro_id: registroID, evento: evento, detalle: detalle || "Seguimiento registrado" }) });
    request.then(function () {
      document.getElementById("mcFollowDetalle").value = "";
      document.getElementById("mcFollowEstado").value = "";
      state.selectedRegistroID = registroID;
      load();
    }).catch(function (err) { setMsg(err.message, true); });
  }

  function saveEvidence(ev) {
    ev.preventDefault();
    var payload = {
      registro_id: Number(document.getElementById("mcEvidenceRegistro").value || 0),
      tipo: document.getElementById("mcEvidenceTipo").value,
      nombre: document.getElementById("mcEvidenceNombre").value.trim(),
      url: document.getElementById("mcEvidenceUrl").value.trim(),
      descripcion: document.getElementById("mcEvidenceDesc").value.trim()
    };
    if (!payload.registro_id || !payload.nombre) {
      setMsg("Selecciona un registro y escribe el nombre de la evidencia.", true);
      return;
    }
    api("evidencia", { method: "POST", body: JSON.stringify(payload) }).then(function () {
      state.selectedRegistroID = payload.registro_id;
      ev.target.reset();
      load();
    }).catch(function (err) { setMsg(err.message, true); });
  }

  function saveApproval(ev) {
    ev.preventDefault();
    var payload = {
      registro_id: Number(document.getElementById("mcApprovalRegistro").value || 0),
      nivel: document.getElementById("mcApprovalNivel").value,
      solicitado_a: document.getElementById("mcApprovalTo").value.trim(),
      comentario: document.getElementById("mcApprovalComment").value.trim(),
      fecha_vencimiento: document.getElementById("mcApprovalDue").value
    };
    if (!payload.registro_id || !payload.solicitado_a) {
      setMsg("Selecciona un registro e indica quien debe aprobar.", true);
      return;
    }
    api("aprobacion_solicitar", { method: "POST", body: JSON.stringify(payload) }).then(function () {
      state.selectedRegistroID = payload.registro_id;
      ev.target.reset();
      load();
    }).catch(function (err) { setMsg(err.message, true); });
  }

  function decideApproval(id, decision) {
    var comentario = window.prompt(decision === "aprobado" ? "Comentario de aprobacion" : "Motivo de rechazo") || "";
    api("aprobacion_decidir", { method: "POST", body: JSON.stringify({ aprobacion_id: Number(id || 0), decision: decision, comentario: comentario }) }).then(load).catch(function (err) { setMsg(err.message, true); });
  }

  function saveTask(ev) {
    ev.preventDefault();
    var payload = {
      registro_id: Number(document.getElementById("mcTaskRegistro").value || 0),
      titulo: document.getElementById("mcTaskTitulo").value.trim(),
      responsable: document.getElementById("mcTaskResponsable").value.trim(),
      prioridad: document.getElementById("mcTaskPrioridad").value,
      fecha_vencimiento: document.getElementById("mcTaskDue").value,
      comentario: document.getElementById("mcTaskComment").value.trim()
    };
    if (!payload.registro_id || !payload.titulo) {
      setMsg("Selecciona un registro y escribe la tarea.", true);
      return;
    }
    api("tarea", { method: "POST", body: JSON.stringify(payload) }).then(function () {
      state.selectedRegistroID = payload.registro_id;
      ev.target.reset();
      load();
    }).catch(function (err) { setMsg(err.message, true); });
  }

  function updateTask(id, estado) {
    var comentario = window.prompt("Comentario de la tarea") || "";
    api("tarea_estado", { method: "POST", body: JSON.stringify({ tarea_id: Number(id || 0), estado: estado, comentario: comentario }) }).then(load).catch(function (err) { setMsg(err.message, true); });
  }

  function closeControlled(id) {
    var detalle = window.prompt("Comentario de cierre controlado") || "";
    api("cierre_controlado", { method: "POST", body: JSON.stringify({ registro_id: Number(id || 0), detalle: detalle }) }).then(function () {
      load();
      loadExpediente(id);
    }).catch(function (err) { setMsg(err.message, true); });
  }

  function generateActionPlan() {
    api("generar_plan_accion", { method: "POST", body: "{}" }).then(function (result) {
      setMsg("Plan de accion generado: " + (result.tareas_creadas || 0) + " tareas creadas, " + (result.omitidas || 0) + " omitidas.");
      load();
    }).catch(function (err) { setMsg(err.message, true); });
  }

  function selectedRegistroIDs() {
    return Array.prototype.slice.call(document.querySelectorAll(".mc-row-select:checked")).map(function (node) {
      return Number(node.value || 0);
    }).filter(Boolean);
  }

  function updateBulkCount() {
    var node = document.getElementById("mcBulkCount");
    if (!node) return;
    var total = selectedRegistroIDs().length;
    node.textContent = total + (total === 1 ? " seleccionado" : " seleccionados");
  }

  function applyBulkAction() {
    var ids = selectedRegistroIDs();
    var payload = {
      registro_ids: ids,
      estado: document.getElementById("mcBulkEstado").value,
      prioridad: document.getElementById("mcBulkPrioridad").value,
      responsable: document.getElementById("mcBulkResponsable").value.trim(),
      detalle: document.getElementById("mcBulkDetalle").value.trim()
    };
    if (!ids.length) {
      setMsg("Selecciona al menos un registro para accion masiva.", true);
      return;
    }
    if (!payload.estado && !payload.prioridad && !payload.responsable) {
      setMsg("Indica estado, prioridad o responsable para actualizar.", true);
      return;
    }
    api("accion_masiva", { method: "POST", body: JSON.stringify(payload) }).then(function (result) {
      setMsg("Accion masiva aplicada: " + (result.actualizados || 0) + "/" + (result.total || ids.length) + " registros actualizados.");
      document.getElementById("mcBulkEstado").value = "";
      document.getElementById("mcBulkPrioridad").value = "";
      document.getElementById("mcBulkResponsable").value = "";
      document.getElementById("mcBulkDetalle").value = "";
      load();
    }).catch(function (err) { setMsg(err.message, true); });
  }

  function exportCSV() {
    setMsg("Generando exportacion...");
    api("exportacion").then(function (data) {
      var lines = [["modulo", data.modulo || state.modulo], ["titulo", data.titulo || state.titulo], ["empresa_id", data.empresa_id || state.empresaId], ["fecha_corte", data.fecha_corte || ""]];
      var csv = rowsToCSV(lines) + "\n\n" + (data.secciones || []).map(function (section) {
        var body = [["seccion", section.nombre || ""]].concat([section.headers || []], section.rows || []);
        return rowsToCSV(body);
      }).join("\n\n");
      downloadCSV(csv, state.modulo + "_auditoria_empresa_" + state.empresaId + ".csv");
      setMsg("Exportacion generada");
    }).catch(function (err) { setMsg(err.message, true); });
  }

  function rowsToCSV(rows) {
    return (rows || []).map(function (row) {
      return (row || []).map(function (value) { return '"' + String(value == null ? "" : value).replace(/"/g, '""') + '"'; }).join(",");
    }).join("\n");
  }

  function downloadCSV(csv, filename) {
    var blob = new Blob([csv], { type: "text/csv;charset=utf-8" });
    var url = URL.createObjectURL(blob);
    var a = document.createElement("a");
    a.href = url;
    a.download = filename;
    a.click();
    setTimeout(function () { URL.revokeObjectURL(url); }, 1000);
  }

  function parseCSV(text) {
    var rows = [];
    var row = [];
    var value = "";
    var quoted = false;
    for (var i = 0; i < text.length; i += 1) {
      var ch = text[i];
      var next = text[i + 1];
      if (ch === '"' && quoted && next === '"') {
        value += '"';
        i += 1;
      } else if (ch === '"') {
        quoted = !quoted;
      } else if (ch === "," && !quoted) {
        row.push(value);
        value = "";
      } else if ((ch === "\n" || ch === "\r") && !quoted) {
        if (ch === "\r" && next === "\n") i += 1;
        row.push(value);
        if (row.some(function (cell) { return cell.trim() !== ""; })) rows.push(row);
        row = [];
        value = "";
      } else {
        value += ch;
      }
    }
    row.push(value);
    if (row.some(function (cell) { return cell.trim() !== ""; })) rows.push(row);
    return rows;
  }

  function importCSV(file) {
    if (!file) return;
    var reader = new FileReader();
    reader.onload = function () {
      try {
        var rows = parseCSV(String(reader.result || ""));
        if (rows.length < 2) throw new Error("El archivo no tiene filas para importar.");
        var headers = rows[0].map(function (h) { return String(h || "").trim().toLowerCase(); });
        var records = rows.slice(1).map(function (cells) {
          var item = {};
          headers.forEach(function (h, idx) { item[h] = cells[idx] || ""; });
          return {
            codigo: item.codigo,
            tipo: item.tipo,
            nombre: item.nombre,
            tercero: item.tercero,
            responsable: item.responsable,
            categoria: item.categoria,
            referencia: item.referencia,
            prioridad: item.prioridad,
            estado: item.estado,
            fecha: item.fecha,
            fecha_vencimiento: item.vencimiento || item.fecha_vencimiento,
            valor: Number(item.valor || 0)
          };
        }).filter(function (item) { return item.codigo || item.nombre; });
        api("importar_registros", { method: "POST", body: JSON.stringify({ registros: records }) }).then(function (result) {
          var errores = (result.errores || []).length ? " Errores: " + result.errores.join(" | ") : "";
          setMsg("Importacion finalizada: " + (result.guardados || 0) + "/" + (result.total || 0) + " registros." + errores, !!errores);
          load();
        }).catch(function (err) { setMsg(err.message, true); });
      } catch (err) {
        setMsg("No se pudo importar CSV: " + err.message, true);
      }
    };
    reader.readAsText(file, "utf-8");
  }

  document.addEventListener("DOMContentLoaded", function () {
    root = document.getElementById("moduloColombiaApp");
    if (!root) return;
    state.empresaId = resolveEmpresaId();
    state.modulo = document.body.getAttribute("data-module") || "";
    state.titulo = document.body.getAttribute("data-title") || "Modulo empresarial";
    state.lead = document.body.getAttribute("data-lead") || "";
    renderShell();
    scrollToRequestedSection();
    api("plantilla").then(function (plantilla) {
      state.plantilla = plantilla || {};
      applyPlantilla();
    }).catch(function () {
      state.plantilla = {};
      applyPlantilla();
    });
    document.getElementById("mcRefresh").addEventListener("click", load);
    document.getElementById("mcSearch").addEventListener("click", load);
    document.getElementById("mcClearSearch").addEventListener("click", clearSearchFilters);
    ["mcEstadoFiltro", "mcTipoFiltro", "mcCategoriaFiltro", "mcPrioridadFiltro", "mcVenceFiltro"].forEach(function (id) {
      document.getElementById(id).addEventListener("change", load);
    });
    ["mcTextoFiltro", "mcResponsableFiltro"].forEach(function (id) {
      document.getElementById(id).addEventListener("keydown", function (ev) {
        if (ev.key === "Enter") {
          ev.preventDefault();
          load();
        }
      });
    });
    document.getElementById("mcClearForm").addEventListener("click", function () {
      document.getElementById("mcForm").reset();
      document.getElementById("mcId").value = "";
      state.selectedRegistroID = 0;
      applyPlantilla();
    });
    document.getElementById("mcBulkApply").addEventListener("click", applyBulkAction);
    document.getElementById("mcSeed").addEventListener("click", function () {
      api("seed_demo", { method: "POST", body: "{}" }).then(load).catch(function (err) { setMsg(err.message, true); });
    });
    document.getElementById("mcExport").addEventListener("click", exportCSV);
    document.getElementById("mcImportBtn").addEventListener("click", function () {
      document.getElementById("mcImportFile").click();
    });
    document.getElementById("mcImportFile").addEventListener("change", function (ev) {
      importCSV(ev.target.files && ev.target.files[0]);
      ev.target.value = "";
    });
    document.getElementById("mcFollowForm").addEventListener("submit", saveFollow);
    document.getElementById("mcEvidenceForm").addEventListener("submit", saveEvidence);
    document.getElementById("mcApprovalForm").addEventListener("submit", saveApproval);
    document.getElementById("mcTaskForm").addEventListener("submit", saveTask);
    document.getElementById("mcForm").addEventListener("submit", function (ev) {
      ev.preventDefault();
      var payload;
      try {
        payload = formPayload();
      } catch (err) {
        setMsg("Metadata JSON invalido: " + err.message, true);
        return;
      }
      api("registro", { method: payload.id ? "PUT" : "POST", body: JSON.stringify(payload) }).then(function () {
        ev.target.reset();
        document.getElementById("mcId").value = "";
        load();
      }).catch(function (err) { setMsg(err.message, true); });
    });
    document.addEventListener("click", function (ev) {
      var btn = ev.target.closest("[data-edit]");
      if (btn) editRow(btn.getAttribute("data-edit"));
      var expBtn = ev.target.closest("[data-expediente]");
      if (expBtn) loadExpediente(expBtn.getAttribute("data-expediente"));
      var closeBtn = ev.target.closest("[data-close-controlled]");
      if (closeBtn) closeControlled(closeBtn.getAttribute("data-close-controlled"));
      if (ev.target.closest("#mcPlanAccion")) generateActionPlan();
      var followBtn = ev.target.closest("[data-follow]");
      if (followBtn) {
        state.selectedRegistroID = Number(followBtn.getAttribute("data-follow") || 0);
        document.getElementById("mcFollowRegistro").value = String(state.selectedRegistroID);
        document.getElementById("mcEvidenceRegistro").value = String(state.selectedRegistroID);
        document.getElementById("mcApprovalRegistro").value = String(state.selectedRegistroID);
        document.getElementById("mcTaskRegistro").value = String(state.selectedRegistroID);
        document.getElementById("mcFollowDetalle").focus();
      }
      var approveBtn = ev.target.closest("[data-approve]");
      if (approveBtn) decideApproval(approveBtn.getAttribute("data-approve"), "aprobado");
      var rejectBtn = ev.target.closest("[data-reject]");
      if (rejectBtn) decideApproval(rejectBtn.getAttribute("data-reject"), "rechazado");
      var progressBtn = ev.target.closest("[data-task-progress]");
      if (progressBtn) updateTask(progressBtn.getAttribute("data-task-progress"), "en_proceso");
      var doneBtn = ev.target.closest("[data-task-done]");
      if (doneBtn) updateTask(doneBtn.getAttribute("data-task-done"), "cumplida");
      var cancelBtn = ev.target.closest("[data-task-cancel]");
      if (cancelBtn) updateTask(cancelBtn.getAttribute("data-task-cancel"), "cancelada");
    });
    document.addEventListener("change", function (ev) {
      if (ev.target && ev.target.id === "mcSelectAll") {
        Array.prototype.slice.call(document.querySelectorAll(".mc-row-select")).forEach(function (node) {
          node.checked = ev.target.checked;
        });
        updateBulkCount();
      }
      if (ev.target && ev.target.classList && ev.target.classList.contains("mc-row-select")) {
        updateBulkCount();
      }
    });
    load();
  });
})();
