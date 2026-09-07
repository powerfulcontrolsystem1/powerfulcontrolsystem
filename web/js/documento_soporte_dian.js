(function () {
  "use strict";

  var CONFIRMACION_DIAN = "EMITIR DOCUMENTO SOPORTE DIAN";
  var panel = document.getElementById("panel-soporte");
  if (!panel) return;

  var ids = [
    "msg", "soporteBody", "soportePreflight", "dsCfgEstado", "dsCfgAmbiente", "dsCfgModo",
    "dsCfgPrefijo", "dsCfgResolucion", "dsCfgDesde", "dsCfgHasta", "dsCfgURL",
    "dsCfgRangoDesde", "dsCfgRangoHasta", "dsCfgConsecutivo", "dsCfgObservaciones",
    "dsCfgResumen", "btnSaveSoporteConfig", "dsPeriodo", "dsFecha", "dsResidencia",
    "dsTipoDocumento", "dsDoc", "dsDV", "dsProveedor", "dsTipoPersona", "dsResponsabilidad",
    "dsPais", "dsDireccion", "dsPostal", "dsDepartamento", "dsDepartamentoDANE",
    "dsMunicipio", "dsMunicipioDANE", "dsEmail", "dsTelefono", "dsConcepto", "dsMoneda",
    "dsFormaPago", "dsMedioPago", "dsVencimiento", "dsLineasBody", "btnAddSoporteLinea",
    "dsTotalSubtotal", "dsTotalIVA", "dsTotalRetenciones", "dsTotalDIAN", "dsTotalNeto",
    "btnSaveSoporte", "dsEmitDialog", "dsEmitConfirmacion", "btnCancelarSoporteEmision",
    "btnConfirmarSoporteEmision", "btnRefresh"
  ];
  var els = {};
  ids.forEach(function (id) { els[id] = document.getElementById(id); });

  var state = {
    empresaID: getEmpresaID(),
    documentos: [],
    configuracion: null,
    lineas: [newLinea()],
    documentoSeleccionadoID: 0,
    cargando: false
  };

  function getEmpresaID() {
    try {
      var params = new URLSearchParams(location.search || "");
      var id = params.get("empresa_id") || params.get("id");
      if (id) return String(id);
    } catch (error) {}
    try {
      if (parent && parent.__resolveEmpresaIdContext) {
        return String(parent.__resolveEmpresaIdContext() || "");
      }
    } catch (error) {}
    return "";
  }

  function contabilidadAPI(action, extra) {
    return "/api/empresa/contabilidad_colombia_avanzada?empresa_id=" + encodeURIComponent(state.empresaID) +
      "&action=" + encodeURIComponent(action) + (extra || "");
  }

  function facturacionAPI(action, extra) {
    return "/api/empresa/facturacion_electronica?empresa_id=" + encodeURIComponent(state.empresaID) +
      "&action=" + encodeURIComponent(action) + (extra || "");
  }

  async function request(url, options) {
    var response = await fetch(url, Object.assign({ credentials: "same-origin" }, options || {}));
    var raw = await response.text();
    var data = {};
    try { data = raw ? JSON.parse(raw) : {}; } catch (error) { data = { error: raw }; }
    if (!response.ok) {
      var message = data.error || data.mensaje || raw || ("HTTP " + response.status);
      if (data.preflight && Array.isArray(data.preflight.bloqueos)) {
        message += ": " + data.preflight.bloqueos.join(" · ");
      }
      throw new Error(message);
    }
    return data;
  }

  function escapeHTML(value) {
    return String(value == null ? "" : value).replace(/[&<>"']/g, function (character) {
      return { "&": "&amp;", "<": "&lt;", ">": "&gt;", "\"": "&quot;", "'": "&#39;" }[character];
    });
  }

  function money(value) {
    try {
      return new Intl.NumberFormat("es-CO", { style: "currency", currency: "COP", minimumFractionDigits: 2 }).format(Number(value || 0));
    } catch (error) {
      return "$" + Number(value || 0).toFixed(2);
    }
  }

  function notify(message, type) {
    if (!els.msg) return;
    els.msg.textContent = message || "";
    els.msg.className = "coadv-msg " + (type || "");
  }

  function round2(value) {
    return Math.round((Number(value || 0) + Number.EPSILON) * 100) / 100;
  }

  function localDateParts() {
    var now = new Date();
    return [now.getFullYear(), String(now.getMonth() + 1).padStart(2, "0"), String(now.getDate()).padStart(2, "0")];
  }

  function today() { return localDateParts().join("-"); }
  function period() { return localDateParts().slice(0, 2).join("-"); }

  function newLinea() {
    return {
      codigo: "", descripcion: "", unidad_medida: "94", cantidad: 1, precio_unitario: 0,
      descuento_porcentaje: 0, iva_porcentaje: 0, reteiva_porcentaje: 0, reterenta_porcentaje: 0
    };
  }

  function option(value, label, selected) {
    return '<option value="' + escapeHTML(value) + '"' + (String(value) === String(selected) ? " selected" : "") + ">" + escapeHTML(label) + "</option>";
  }

  function renderTipoDocumento() {
    var current = els.dsTipoDocumento.value;
    var options = els.dsResidencia.value === "residente" ? [
      ["NIT", "31 · NIT"]
    ] : [
      ["TE", "21 · Tarjeta de extranjería"], ["CE", "22 · Cédula de extranjería"],
      ["NIT", "31 · NIT"], ["PAS", "41 · Pasaporte"],
      ["DIE", "42 · Documento de identificación extranjero"], ["PEP", "47 · PEP"],
      ["NIT otro pais", "50 · NIT de otro país"]
    ];
    if (!options.some(function (item) { return item[0] === current; })) current = options[0][0];
    els.dsTipoDocumento.innerHTML = options.map(function (item) { return option(item[0], item[1], current); }).join("");
    var resident = els.dsResidencia.value === "residente";
    els.dsDV.disabled = !resident && els.dsTipoDocumento.value !== "NIT";
    els.dsDepartamentoDANE.disabled = !resident;
    els.dsMunicipioDANE.disabled = !resident;
    if (resident && !els.dsPais.value) els.dsPais.value = "CO";
    if (!resident && els.dsPais.value.trim().toUpperCase() === "CO") els.dsPais.value = "";
  }

  function renderLineas() {
    var iva = [0, 5, 19];
    var reteIVA = [0, 15, 100];
    var reteRenta = [0, 0.10, 0.50, 1, 1.50, 2, 2.50, 3, 3.50, 4, 6, 7, 10, 11, 20];
    els.dsLineasBody.innerHTML = state.lineas.map(function (linea, index) {
      return '<tr data-ds-line-index="' + index + '">' +
        '<td><input class="form-input" data-ds-line="codigo" value="' + escapeHTML(linea.codigo) + '" aria-label="Código línea ' + (index + 1) + '"></td>' +
        '<td><input class="form-input" data-ds-line="descripcion" value="' + escapeHTML(linea.descripcion) + '" aria-label="Descripción línea ' + (index + 1) + '"></td>' +
        '<td><input class="form-input" data-ds-line="unidad_medida" list="dsUnidadMedidaOpciones" maxlength="3" value="' + escapeHTML(linea.unidad_medida) + '" aria-label="Unidad DIAN línea ' + (index + 1) + '"></td>' +
        '<td><input class="form-input" data-ds-line="cantidad" type="number" min="0.000001" step="0.000001" value="' + escapeHTML(linea.cantidad) + '"></td>' +
        '<td><input class="form-input" data-ds-line="precio_unitario" type="number" min="0" step="0.01" value="' + escapeHTML(linea.precio_unitario) + '"></td>' +
        '<td><input class="form-input" data-ds-line="descuento_porcentaje" type="number" min="0" max="100" step="0.01" value="' + escapeHTML(linea.descuento_porcentaje) + '"></td>' +
        '<td><select class="form-input" data-ds-line="iva_porcentaje">' + iva.map(function (rate) { return option(rate, rate + "%", linea.iva_porcentaje); }).join("") + '</select></td>' +
        '<td><select class="form-input" data-ds-line="reteiva_porcentaje">' + reteIVA.map(function (rate) { return option(rate, rate + "%", linea.reteiva_porcentaje); }).join("") + '</select></td>' +
        '<td><select class="form-input" data-ds-line="reterenta_porcentaje">' + reteRenta.map(function (rate) { return option(rate, rate + "%", linea.reterenta_porcentaje); }).join("") + '</select></td>' +
        '<td><strong data-ds-line-neto>' + money(0) + '</strong></td>' +
        '<td><button class="btn secondary" type="button" data-ds-remove-line="' + index + '"' + (state.lineas.length === 1 ? " disabled" : "") + '>Quitar</button></td></tr>';
    }).join("");
    calculateTotals();
  }

  function syncLineInput(target) {
    var row = target.closest("[data-ds-line-index]");
    if (!row) return;
    var index = Number(row.getAttribute("data-ds-line-index"));
    var field = target.getAttribute("data-ds-line");
    if (!state.lineas[index] || !field) return;
    if (["cantidad", "precio_unitario", "descuento_porcentaje", "iva_porcentaje", "reteiva_porcentaje", "reterenta_porcentaje"].indexOf(field) >= 0) {
      state.lineas[index][field] = Number(target.value || 0);
    } else {
      state.lineas[index][field] = field === "unidad_medida" ? target.value.trim().toUpperCase() : target.value;
    }
    calculateTotals();
  }

  function calculatedLine(linea) {
    var gross = round2(Number(linea.cantidad || 0) * Number(linea.precio_unitario || 0));
    var discount = round2(gross * Number(linea.descuento_porcentaje || 0) / 100);
    var base = round2(gross - discount);
    var iva = round2(base * Number(linea.iva_porcentaje || 0) / 100);
    var reteIVA = round2(iva * Number(linea.reteiva_porcentaje || 0) / 100);
    var reteRenta = round2(base * Number(linea.reterenta_porcentaje || 0) / 100);
    var total = round2(base + iva);
    return { base: base, iva: iva, retenciones: round2(reteIVA + reteRenta), total: total, neto: round2(total - reteIVA - reteRenta) };
  }

  function calculateTotals() {
    var totals = { base: 0, iva: 0, retenciones: 0, total: 0, neto: 0 };
    state.lineas.forEach(function (linea, index) {
      var values = calculatedLine(linea);
      Object.keys(totals).forEach(function (key) { totals[key] += values[key]; });
      var row = els.dsLineasBody.querySelector('[data-ds-line-index="' + index + '"] [data-ds-line-neto]');
      if (row) row.textContent = money(values.neto);
    });
    Object.keys(totals).forEach(function (key) { totals[key] = round2(totals[key]); });
    els.dsTotalSubtotal.textContent = money(totals.base);
    els.dsTotalIVA.textContent = money(totals.iva);
    els.dsTotalRetenciones.textContent = money(totals.retenciones);
    els.dsTotalDIAN.textContent = money(totals.total);
    els.dsTotalNeto.textContent = money(totals.neto);
    return totals;
  }

  function configPayload() {
    return {
      tipo_documento: "documento_soporte",
      estado: els.dsCfgEstado.value,
      tipo_ambiente: els.dsCfgAmbiente.value,
      modo_operacion_codigo: els.dsCfgModo.value.trim(),
      prefijo: els.dsCfgPrefijo.value.trim(),
      resolucion_numero: els.dsCfgResolucion.value.trim(),
      resolucion_fecha_desde: els.dsCfgDesde.value,
      resolucion_fecha_hasta: els.dsCfgHasta.value,
      rango_desde: Number(els.dsCfgRangoDesde.value || 0),
      rango_hasta: Number(els.dsCfgRangoHasta.value || 0),
      consecutivo_actual: Number(els.dsCfgConsecutivo.value || 0),
      url_dian_override: els.dsCfgURL.value.trim(),
      observaciones: els.dsCfgObservaciones.value.trim()
    };
  }

  function fillConfig(item) {
    state.configuracion = item || null;
    item = item || { estado: "configurando", tipo_ambiente: "habilitacion", rango_desde: 1, rango_hasta: 1, consecutivo_actual: 1 };
    els.dsCfgEstado.value = item.estado || "configurando";
    els.dsCfgAmbiente.value = item.tipo_ambiente || "habilitacion";
    els.dsCfgModo.value = item.modo_operacion_codigo || "";
    els.dsCfgPrefijo.value = item.prefijo || "";
    els.dsCfgResolucion.value = item.resolucion_numero || "";
    els.dsCfgDesde.value = item.resolucion_fecha_desde || "";
    els.dsCfgHasta.value = item.resolucion_fecha_hasta || "";
    els.dsCfgRangoDesde.value = Number(item.rango_desde || 0);
    els.dsCfgRangoHasta.value = Number(item.rango_hasta || 0);
    els.dsCfgConsecutivo.value = Number(item.consecutivo_actual || 0);
    els.dsCfgURL.value = item.url_dian_override || "";
    els.dsCfgObservaciones.value = item.observaciones || "";
    els.dsCfgResumen.textContent = state.configuracion ?
      ("Configuración " + (item.estado || "-") + " · " + (item.tipo_ambiente || "-") + " · próximo " + (item.prefijo || "") + (item.consecutivo_actual || 0)) :
      "Aún no existe configuración separada para documento soporte.";
  }

  async function loadConfig() {
    var response = await request(facturacionAPI("configuracion_documentos_dian"));
    var items = Array.isArray(response.items) ? response.items : [];
    fillConfig(items.find(function (item) { return item.tipo_documento === "documento_soporte"; }) || null);
  }

  async function saveConfig() {
    var payload = configPayload();
    if (payload.tipo_ambiente === "habilitacion" && payload.estado === "activo") {
      throw new Error("En habilitación usa el estado Habilitación; Activo corresponde a producción.");
    }
    if (payload.tipo_ambiente === "produccion" && payload.estado === "habilitacion") {
      throw new Error("En producción el estado operativo debe ser Activo.");
    }
    if ((payload.estado === "habilitacion" || payload.estado === "activo") && !/^[0-9]{14}$/.test(payload.resolucion_numero)) {
      throw new Error("La autorización DIAN debe contener exactamente 14 dígitos.");
    }
    if (!/^[A-Za-z0-9]{0,4}$/.test(payload.prefijo)) {
      throw new Error("El prefijo es opcional; si se usa, debe ser alfanumérico y tener máximo 4 caracteres.");
    }
    if ((payload.estado === "habilitacion" || payload.estado === "activo") && (!payload.resolucion_fecha_desde || !payload.resolucion_fecha_hasta)) {
      throw new Error("La vigencia inicial y final de la autorización DIAN es obligatoria.");
    }
    if (payload.rango_desde < 1 || payload.rango_hasta < payload.rango_desde || payload.rango_hasta > 999999999) {
      throw new Error("El rango autorizado debe estar entre 1 y 999999999.");
    }
    if (payload.consecutivo_actual < payload.rango_desde || payload.consecutivo_actual > payload.rango_hasta) {
      throw new Error("El próximo consecutivo debe estar dentro del rango autorizado.");
    }
    els.btnSaveSoporteConfig.disabled = true;
    try {
      var response = await request(facturacionAPI("configuracion_documentos_dian"), {
        method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(payload)
      });
      fillConfig(response.item || payload);
      notify("Configuración DIAN de documento soporte guardada.", "success");
    } finally {
      els.btnSaveSoporteConfig.disabled = false;
    }
  }

  function draftPayload() {
    return {
      tipo_documento: els.dsTipoDocumento.value,
      documento: els.dsDoc.value.trim(),
      vendedor_digito_verificacion: els.dsDV.value.trim(),
      vendedor_tipo_persona: els.dsTipoPersona.value,
      vendedor_residencia: els.dsResidencia.value,
      nombre_proveedor: els.dsProveedor.value.trim(),
      vendedor_direccion: els.dsDireccion.value.trim(),
      vendedor_pais_codigo: els.dsPais.value.trim().toUpperCase(),
      vendedor_departamento: els.dsDepartamento.value.trim(),
      vendedor_departamento_codigo_dane: els.dsDepartamentoDANE.value.trim(),
      vendedor_municipio: els.dsMunicipio.value.trim(),
      vendedor_municipio_codigo_dane: els.dsMunicipioDANE.value.trim(),
      vendedor_codigo_postal: els.dsPostal.value.trim(),
      vendedor_responsabilidad_tributaria: els.dsResponsabilidad.value.trim(),
      vendedor_email: els.dsEmail.value.trim(),
      vendedor_telefono: els.dsTelefono.value.trim(),
      fecha_documento: els.dsFecha.value,
      periodo: els.dsPeriodo.value.trim(),
      concepto: els.dsConcepto.value.trim(),
      moneda: "COP",
      forma_pago_codigo: els.dsFormaPago.value,
      medio_pago_codigo: els.dsMedioPago.value,
      fecha_vencimiento: els.dsFormaPago.value === "2" ? els.dsVencimiento.value : "",
      lineas_json: JSON.stringify(state.lineas.map(function (linea, index) {
        return {
          numero: index + 1, codigo: linea.codigo.trim(), descripcion: linea.descripcion.trim(),
          unidad_medida: linea.unidad_medida.trim().toUpperCase(), cantidad: Number(linea.cantidad || 0),
          precio_unitario: Number(linea.precio_unitario || 0),
          descuento_porcentaje: Number(linea.descuento_porcentaje || 0),
          iva_porcentaje: Number(linea.iva_porcentaje || 0),
          reteiva_porcentaje: Number(linea.reteiva_porcentaje || 0),
          reterenta_porcentaje: Number(linea.reterenta_porcentaje || 0)
        };
      }))
    };
  }

  function validateDraft(payload) {
    if (!payload.fecha_documento || !/^\d{4}-\d{2}$/.test(payload.periodo)) return "Fecha y periodo AAAA-MM son obligatorios.";
    if (!payload.documento || !payload.nombre_proveedor || !payload.concepto) return "Identificación, vendedor y concepto son obligatorios.";
    if (!payload.vendedor_direccion || !payload.vendedor_departamento || !payload.vendedor_municipio || !payload.vendedor_responsabilidad_tributaria) return "Completa ubicación y responsabilidad tributaria del vendedor.";
    if (!/^[A-Z]{2}$/.test(payload.vendedor_pais_codigo)) return "El país debe usar dos letras ISO.";
    if (payload.vendedor_residencia === "residente") {
      if (payload.tipo_documento !== "NIT" || payload.vendedor_pais_codigo !== "CO" || !/^\d$/.test(payload.vendedor_digito_verificacion)) return "Un vendedor residente requiere NIT, DV y país CO.";
      if (!/^\d{2}$/.test(payload.vendedor_departamento_codigo_dane) || !/^\d{5}$/.test(payload.vendedor_municipio_codigo_dane)) return "El vendedor residente requiere códigos DANE válidos.";
    } else if (payload.vendedor_pais_codigo === "CO") {
      return "Un vendedor no residente requiere un país distinto de CO.";
    }
    if (payload.forma_pago_codigo === "2" && !payload.fecha_vencimiento) return "El pago a crédito requiere fecha de vencimiento.";
    if (!state.lineas.length) return "Agrega al menos una línea.";
    for (var index = 0; index < state.lineas.length; index += 1) {
      var line = state.lineas[index];
      if (!line.codigo.trim() || !line.descripcion.trim() || Number(line.cantidad) <= 0 || Number(line.precio_unitario) < 0) return "Completa código, descripción, cantidad y precio de la línea " + (index + 1) + ".";
      if (!/^[A-Z0-9]{1,3}$/.test(line.unidad_medida.trim().toUpperCase())) return "La unidad DIAN de la línea " + (index + 1) + " debe ser un código alfanumérico de hasta 3 caracteres.";
    }
    if (calculateTotals().total <= 0) return "El total DIAN calculado debe ser positivo.";
    return "";
  }

  async function saveDraft() {
    var payload = draftPayload();
    var validation = validateDraft(payload);
    if (validation) throw new Error(validation);
    els.btnSaveSoporte.disabled = true;
    try {
      await request(contabilidadAPI("documentos_soporte"), {
        method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(payload)
      });
      notify("Borrador guardado. No se consumió consecutivo ni se transmitió a DIAN.", "success");
      state.lineas = [newLinea()];
      renderLineas();
      await loadDocumentos();
    } finally {
      els.btnSaveSoporte.disabled = false;
    }
  }

  function statusBadge(value) {
    var stateValue = String(value || "borrador");
    var kind = /aceptado|enviado/i.test(stateValue) ? "ok" : (/rechazado|fallido/i.test(stateValue) ? "danger" : "warn");
    return '<span class="coadv-badge ' + kind + '">' + escapeHTML(stateValue) + "</span>";
  }

  function shortCUDS(value) {
    value = String(value || "");
    if (!value) return "—";
    return '<span title="' + escapeHTML(value) + '">' + escapeHTML(value.slice(0, 12)) + "…</span>";
  }

  function renderDocumentos() {
    els.soporteBody.innerHTML = state.documentos.map(function (item) {
      var number = item.numero_legal || ("Borrador #" + item.id);
      var artifacts = item.numero_legal || item.fuente_fiscal_sellada ? '<button class="btn secondary" type="button" data-ds-artifacts="' + escapeHTML(item.id) + '">Artefactos</button>' : "";
      return "<tr><td><strong>" + escapeHTML(number) + "</strong></td><td>" + escapeHTML(item.fecha_documento) + "</td><td>" +
        escapeHTML(item.nombre_proveedor) + '<div class="form-help">' + escapeHTML(item.tipo_documento + " " + item.documento) + "</div></td><td>" +
        money(item.total) + "</td><td>" + money(item.total_neto_contable) + "</td><td>" + statusBadge(item.estado_dian) + "</td><td>" +
        shortCUDS(item.cuds) + '</td><td><div class="coadv-actions"><button class="btn secondary" type="button" data-ds-preflight="' + escapeHTML(item.id) +
        '">Revisar DIAN</button>' + artifacts + "</div></td></tr>";
    }).join("") || '<tr><td colspan="8"><div class="coadv-empty">Sin documentos soporte. No existe una compra real registrada para emitir.</div></td></tr>';
  }

  async function loadDocumentos() {
    var rows = await request(contabilidadAPI("documentos_soporte"));
    state.documentos = Array.isArray(rows) ? rows : [];
    renderDocumentos();
  }

  function renderPreflight(out, id) {
    var blocks = Array.isArray(out.bloqueos) ? out.bloqueos : [];
    var warnings = Array.isArray(out.advertencias) ? out.advertencias : [];
    var emission = out.puede_emitir ? '<div class="coadv-actions"><button class="btn" type="button" data-ds-emit="' + escapeHTML(id) + '">Emitir a DIAN</button><span class="form-help">Requiere escribir la frase de confirmación antes de reservar el consecutivo.</span></div>' : "";
    els.soportePreflight.innerHTML = '<strong>Revisión DIAN: ' + escapeHTML(out.estado || "pendiente") + "</strong> " + statusBadge(out.puede_emitir ? "listo" : "bloqueado") +
      '<div class="form-help">Ambiente ' + escapeHTML(out.ambiente || "sin configurar") + " · configuración " + escapeHTML(out.estado_configuracion || "sin configurar") + ". No se generó XML, no se consumió consecutivo y no se transmitió información.</div>" +
      (blocks.length ? '<div class="coadv-badge coadv-badge-spaced danger">Bloqueos ' + blocks.length + "</div><ul>" + blocks.map(function (item) { return "<li>" + escapeHTML(item) + "</li>"; }).join("") + "</ul>" : "") +
      (warnings.length ? '<div class="coadv-badge warn">Advertencias ' + warnings.length + "</div><ul>" + warnings.map(function (item) { return "<li>" + escapeHTML(item) + "</li>"; }).join("") + "</ul>" : "") + emission;
  }

  async function preflight(id) {
    state.documentoSeleccionadoID = Number(id || 0);
    var out = await request(contabilidadAPI("documento_soporte_preflight", "&documento_soporte_id=" + encodeURIComponent(id)));
    renderPreflight(out, id);
    notify(out.puede_emitir ? "Preflight completo: el documento está listo para confirmación." : "Preflight completado con bloqueos; no hubo emisión.", out.puede_emitir ? "success" : "error");
  }

  async function loadArtifacts(id) {
    var code = "DS-SOPORTE-" + Number(id);
    var response = await request(facturacionAPI("artefactos", "&tipo_documento=documento_soporte&documento_codigo=" + encodeURIComponent(code)));
    var items = Array.isArray(response.items) ? response.items : [];
    els.soportePreflight.innerHTML = '<strong>Artefactos fiscales de ' + escapeHTML(code) + "</strong>" +
      (items.length ? '<div class="coadv-artifacts">' + items.map(function (item) {
        return '<a class="btn secondary" href="' + escapeHTML(item.download_url) + '">' + escapeHTML(item.tipo_artefacto) + " · " + escapeHTML(item.sha256.slice(0, 12)) + "…</a>";
      }).join("") + "</div>" : '<div class="form-help">Aún no hay XML firmado, respuesta o representación fiscal conservada.</div>');
  }

  function openEmitDialog(id) {
    state.documentoSeleccionadoID = Number(id || 0);
    els.dsEmitConfirmacion.value = "";
    els.btnConfirmarSoporteEmision.disabled = true;
    if (typeof els.dsEmitDialog.showModal === "function") els.dsEmitDialog.showModal();
    else els.dsEmitDialog.setAttribute("open", "open");
    els.dsEmitConfirmacion.focus();
  }

  function closeEmitDialog() {
    if (typeof els.dsEmitDialog.close === "function") els.dsEmitDialog.close();
    else els.dsEmitDialog.removeAttribute("open");
    els.dsEmitConfirmacion.value = "";
    els.btnConfirmarSoporteEmision.disabled = true;
  }

  async function emitSelected() {
    if (els.dsEmitConfirmacion.value !== CONFIRMACION_DIAN || state.documentoSeleccionadoID <= 0) return;
    els.btnConfirmarSoporteEmision.disabled = true;
    try {
      var response = await request(facturacionAPI("emitir_documento_soporte"), {
        method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({
          empresa_id: Number(state.empresaID), documento_soporte_id: state.documentoSeleccionadoID,
          confirmar_emision: true, mensaje_confirmacion_dian: CONFIRMACION_DIAN
        })
      });
      closeEmitDialog();
      await loadDocumentos();
      var integration = response.integracion_fiscal || {};
      notify("Emisión procesada. Estado de integración: " + (integration.estado_envio || "pendiente") + ".", response.ok ? "success" : "error");
    } finally {
      els.btnConfirmarSoporteEmision.disabled = els.dsEmitConfirmacion.value !== CONFIRMACION_DIAN;
    }
  }

  async function loadAll() {
    if (!state.empresaID || state.cargando) return;
    state.cargando = true;
    try {
      var results = await Promise.allSettled([loadConfig(), loadDocumentos()]);
      var errors = results.filter(function (result) { return result.status === "rejected"; });
      if (errors.length) throw errors[0].reason;
    } finally {
      state.cargando = false;
    }
  }

  els.dsLineasBody.addEventListener("input", function (event) { if (event.target.matches("[data-ds-line]")) syncLineInput(event.target); });
  els.dsLineasBody.addEventListener("change", function (event) { if (event.target.matches("[data-ds-line]")) syncLineInput(event.target); });
  els.dsLineasBody.addEventListener("click", function (event) {
    var button = event.target.closest("[data-ds-remove-line]");
    if (!button || state.lineas.length <= 1) return;
    state.lineas.splice(Number(button.getAttribute("data-ds-remove-line")), 1);
    renderLineas();
  });
  els.soporteBody.addEventListener("click", function (event) {
    var review = event.target.closest("[data-ds-preflight]");
    var artifacts = event.target.closest("[data-ds-artifacts]");
    if (review) preflight(review.getAttribute("data-ds-preflight")).catch(function (error) { notify(error.message, "error"); });
    if (artifacts) loadArtifacts(artifacts.getAttribute("data-ds-artifacts")).catch(function (error) { notify(error.message, "error"); });
  });
  els.soportePreflight.addEventListener("click", function (event) {
    var button = event.target.closest("[data-ds-emit]");
    if (button) openEmitDialog(button.getAttribute("data-ds-emit"));
  });
  els.btnAddSoporteLinea.addEventListener("click", function () { state.lineas.push(newLinea()); renderLineas(); });
  els.dsResidencia.addEventListener("change", renderTipoDocumento);
  els.dsTipoDocumento.addEventListener("change", renderTipoDocumento);
  els.dsFormaPago.addEventListener("change", function () {
    var credit = els.dsFormaPago.value === "2";
    els.dsVencimiento.disabled = !credit;
    if (!credit) els.dsVencimiento.value = "";
  });
  els.btnSaveSoporteConfig.addEventListener("click", function () { saveConfig().catch(function (error) { notify(error.message, "error"); }); });
  els.btnSaveSoporte.addEventListener("click", function () { saveDraft().catch(function (error) { notify(error.message, "error"); }); });
  els.dsEmitConfirmacion.addEventListener("input", function () { els.btnConfirmarSoporteEmision.disabled = els.dsEmitConfirmacion.value !== CONFIRMACION_DIAN; });
  els.btnCancelarSoporteEmision.addEventListener("click", closeEmitDialog);
  els.btnConfirmarSoporteEmision.addEventListener("click", function () { emitSelected().catch(function (error) { notify(error.message, "error"); }); });
  if (els.btnRefresh) els.btnRefresh.addEventListener("click", function () { loadAll().catch(function (error) { notify(error.message, "error"); }); });

  els.dsPeriodo.value = period();
  els.dsFecha.value = today();
  renderTipoDocumento();
  renderLineas();
  loadAll().catch(function (error) { notify(error.message, "error"); });
})();
