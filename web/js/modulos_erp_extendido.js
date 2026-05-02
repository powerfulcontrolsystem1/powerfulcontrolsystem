(function () {
  var MODULES = [
    { domain: "ventas", key: "ventas_cotizaciones", label: "Ventas - Cotizaciones", endpoint: "/api/empresa/ventas/cotizaciones" },
    { domain: "ventas", key: "ventas_pedidos", label: "Ventas - Pedidos", endpoint: "/api/empresa/ventas/pedidos" },
    { domain: "ventas", key: "ventas_devoluciones", label: "Ventas - Devoluciones", endpoint: "/api/empresa/ventas/devoluciones" },

    { domain: "finanzas", key: "finanzas_plan_cuentas", label: "Finanzas - Plan de cuentas", endpoint: "/api/empresa/finanzas/plan_cuentas" },
    { domain: "finanzas", key: "finanzas_cxc", label: "Finanzas - Cuentas por cobrar", endpoint: "/api/empresa/finanzas/cuentas_cobrar" },
    { domain: "finanzas", key: "finanzas_cxp", label: "Finanzas - Cuentas por pagar", endpoint: "/api/empresa/finanzas/cuentas_pagar" },

    { domain: "inventario_compras_rrhh", key: "inventario_lotes_series", label: "Inventario - Lotes y series", endpoint: "/api/empresa/inventario/lotes_series" },
    { domain: "inventario_compras_rrhh", key: "compras_devoluciones_proveedor", label: "Compras - Devoluciones proveedor", endpoint: "/api/empresa/compras/devoluciones_proveedor" },
    { domain: "inventario_compras_rrhh", key: "rrhh_vacaciones_licencias", label: "RRHH - Vacaciones y licencias", endpoint: "/api/empresa/rrhh/vacaciones_licencias" },

    { domain: "crm", key: "crm_leads", label: "CRM - Leads", endpoint: "/api/empresa/crm/leads" },
    { domain: "crm", key: "crm_interacciones", label: "CRM - Interacciones", endpoint: "/api/empresa/crm/interacciones" },
    { domain: "crm", key: "crm_campanas", label: "CRM - Campanas", endpoint: "/api/empresa/crm/campanas" },

    { domain: "produccion", key: "produccion_bom", label: "Produccion - BOM", endpoint: "/api/empresa/produccion/bom" },
    { domain: "produccion", key: "produccion_bom_detalle", label: "Produccion - BOM detalle", endpoint: "/api/empresa/produccion/bom_detalle" },
    { domain: "produccion", key: "produccion_ordenes", label: "Produccion - Ordenes", endpoint: "/api/empresa/produccion/ordenes" },

    { domain: "logistica", key: "logistica_transportistas", label: "Logistica - Transportistas", endpoint: "/api/empresa/logistica/transportistas" },
    { domain: "logistica", key: "logistica_rutas", label: "Logistica - Rutas", endpoint: "/api/empresa/logistica/rutas" },
    { domain: "logistica", key: "logistica_envios", label: "Logistica - Envios", endpoint: "/api/empresa/logistica/envios" },

    { domain: "documental_integraciones_dian", key: "documentos_gestion", label: "Documentos - Gestion", endpoint: "/api/empresa/documentos/gestion" },
    { domain: "documental_integraciones_dian", key: "documentos_firmas", label: "Documentos - Firmas", endpoint: "/api/empresa/documentos/firmas" },
    { domain: "documental_integraciones_dian", key: "integraciones_apis", label: "Integraciones - APIs", endpoint: "/api/empresa/integraciones/apis" },
    { domain: "documental_integraciones_dian", key: "integraciones_bancos", label: "Integraciones - Bancos", endpoint: "/api/empresa/integraciones/bancos" },
    { domain: "documental_integraciones_dian", key: "facturacion_dian", label: "Facturacion - DIAN Colombia", endpoint: "/api/empresa/facturacion_electronica/dian", dian: true }
  ];

  var DOMAIN_LABELS = {
    ventas: "Ventas",
    finanzas: "Finanzas",
    inventario_compras_rrhh: "Inventario / Compras / RRHH",
    crm: "CRM",
    produccion: "Produccion",
    logistica: "Logistica",
    documental_integraciones_dian: "Documental / Integraciones / DIAN"
  };

  var state = {
    empresaID: 0,
    domain: "ventas",
    rows: [],
    selectedModuleKey: "",
    guidedFields: [],
    quickActions: [],
    quickActionKey: ""
  };

  var STATE_MACHINE_MODULE_KEYS = {
    ventas_cotizaciones: true,
    ventas_pedidos: true,
    ventas_devoluciones: true,
    crm_leads: true,
    crm_interacciones: true,
    crm_campanas: true
  };

  var INTEGRATION_MODULE_KEYS = {
    integraciones_apis: true,
    integraciones_bancos: true
  };

  var DOMAIN_GUIDES = {
    ventas: {
      objetivo: "Convertir oportunidades en documentos comerciales validos y trazables.",
      flujo: [
        "Registra cotizacion con cliente y montos base.",
        "Promueve a pedido cuando haya aceptacion.",
        "Usa devolucion cuando debas corregir o revertir una entrega."
      ],
      controles: [
        "Verifica que total no sea menor a subtotal.",
        "Consulta transiciones para mover estado sin romper el flujo.",
        "Exporta el embudo para seguimiento comercial."
      ]
    },
    finanzas: {
      objetivo: "Mantener cuentas contables y cartera con saldos consistentes.",
      flujo: [
        "Define plan de cuentas por tipo de empresa.",
        "Registra CxC y CxP con documento de referencia.",
        "Concilia pagos y valida cierres de periodo antes de editar."
      ],
      controles: [
        "Saldo no debe superar valor original.",
        "Usa estados activo/inactivo para control operativo.",
        "Bloquea cambios en periodos cerrados."
      ]
    },
    inventario_compras_rrhh: {
      objetivo: "Operar lotes, devoluciones y novedades RRHH sin perder trazabilidad.",
      flujo: [
        "Configura lote o serie con cantidades iniciales y disponibles.",
        "Registra devolucion al proveedor con motivo y montos.",
        "Captura vacaciones/licencias con rango de fechas valido."
      ],
      controles: [
        "No permitas cantidades negativas.",
        "Valida fecha fin mayor o igual a fecha inicio en RRHH.",
        "Consulta detalle antes de desactivar o eliminar."
      ]
    },
    crm: {
      objetivo: "Gestionar embudo CRM por etapas y acciones comerciales.",
      flujo: [
        "Crea lead con canal de origen.",
        "Registra interacciones con resumen operativo.",
        "Activa campanas y monitorea conversion."
      ],
      controles: [
        "Usa transiciones para mover estados de leads/campanas.",
        "Registra observaciones de contacto para auditoria.",
        "Mantiene filtros por texto para seguimiento rapido."
      ]
    },
    produccion: {
      objetivo: "Planificar BOM y ordenes con capacidad operativa real.",
      flujo: [
        "Crea BOM maestro y su detalle de insumos.",
        "Programa ordenes con cantidad objetivo.",
        "Consulta plan de capacidad para alertas de sobrecarga."
      ],
      controles: [
        "Evita cantidades en cero para ordenes clave.",
        "Usa detalle por ID para validar insumos.",
        "Documenta incidencias en observaciones."
      ]
    },
    logistica: {
      objetivo: "Coordinar despacho, rutas y entregas con SLA.",
      flujo: [
        "Registra transportistas y rutas operativas.",
        "Crea envios con direccion y cliente.",
        "Consulta seguimiento de hitos para SLA."
      ],
      controles: [
        "Mantiene datos de origen/destino normalizados.",
        "Verifica estado antes de transicionar.",
        "Aplica desactivacion logica para historico."
      ]
    },
    documental_integraciones_dian: {
      objetivo: "Gestionar evidencia documental e integraciones seguras por empresa.",
      flujo: [
        "Registra documento/firma con entidad y modulo.",
        "Ejecuta health y sync en conectores API/Bancos.",
        "Usa herramientas DIAN para checklist, validar y demos de CUFE/XML."
      ],
      controles: [
        "Nunca persistas secretos en texto plano.",
        "Rota credenciales con referencias seguras.",
        "Monitorea conectores y corrige alertas de latencia."
      ]
    }
  };

  var MODULE_REQUIRED_FIELDS = {
    ventas_cotizaciones: ["cliente_nombre", "subtotal", "total"],
    ventas_pedidos: ["cliente_nombre", "subtotal", "total"],
    ventas_devoluciones: ["motivo", "subtotal", "total"],
    finanzas_plan_cuentas: ["codigo", "nombre", "tipo_cuenta"],
    finanzas_cxc: ["cliente_nombre", "documento_codigo", "valor_original", "saldo"],
    finanzas_cxp: ["proveedor_nombre", "documento_codigo", "valor_original", "saldo"],
    inventario_lotes_series: ["producto_id", "codigo_lote_serie", "cantidad_inicial", "cantidad_disponible"],
    compras_devoluciones_proveedor: ["proveedor_nombre", "motivo", "total"],
    rrhh_vacaciones_licencias: ["empleado_nombre", "tipo_novedad", "fecha_inicio", "fecha_fin"],
    crm_leads: ["nombre", "canal_origen"],
    crm_interacciones: ["tipo_interaccion", "resumen"],
    crm_campanas: ["nombre", "canal"],
    produccion_bom: ["codigo", "producto_nombre"],
    produccion_bom_detalle: ["bom_id", "insumo_nombre", "cantidad"],
    produccion_ordenes: ["producto_nombre", "cantidad_programada"],
    logistica_transportistas: ["nombre"],
    logistica_rutas: ["nombre", "origen", "destino"],
    logistica_envios: ["cliente_nombre", "direccion_entrega"],
    documentos_gestion: ["modulo", "entidad", "nombre_documento"],
    documentos_firmas: ["documento_gestion_id", "tipo_firma", "firmante_nombre"],
    integraciones_apis: ["nombre_integracion", "tipo_integracion"],
    integraciones_bancos: ["banco_nombre", "numero_cuenta"],
    facturacion_dian: ["nit", "razon_social", "tipo_ambiente"]
  };

  var NON_NEGATIVE_FIELDS = {
    subtotal: true,
    total: true,
    saldo: true,
    valor_original: true,
    cantidad: true,
    cantidad_inicial: true,
    cantidad_disponible: true,
    cantidad_programada: true,
    bom_id: true,
    producto_id: true,
    documento_gestion_id: true
  };

  var FIELD_PRESETS = {
    tipo_cuenta: { type: "select", options: ["activo", "pasivo", "patrimonio", "ingreso", "egreso", "costo"] },
    tipo_novedad: { type: "select", options: ["vacacion", "licencia", "permiso", "incapacidad"] },
    tipo_ambiente: { type: "select", options: ["habilitacion", "produccion"] },
    tipo_integracion: { type: "select", options: ["rest", "soap", "sftp", "manual"] },
    tipo_interaccion: { type: "select", options: ["llamada", "correo", "visita", "chat", "otro"] },
    canal: { type: "select", options: ["email", "sms", "whatsapp", "web", "otro"] },
    canal_origen: { type: "select", options: ["web", "telefono", "referido", "redes", "otro"] },
    tipo_firma: { type: "select", options: ["digital", "biometrica", "manual"] },
    estado: { type: "select", options: ["activo", "inactivo"] }
  };

  function getQueryParam(name) {
    var params = new URLSearchParams(window.location.search);
    return params.get(name);
  }

  function normalize(value) {
    return String(value == null ? "" : value).trim();
  }

  function toInt(value, fallback) {
    var parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : fallback;
  }

  function isDomainValid(domainKey) {
    return Object.prototype.hasOwnProperty.call(DOMAIN_LABELS, domainKey);
  }

  function ensureEmpresaID() {
    var raw = getQueryParam("empresa_id") || getQueryParam("id");
    var parsed = Number(raw);
    if (!Number.isFinite(parsed) || parsed <= 0) {
      throw new Error("empresa_id es obligatorio en la URL");
    }
    state.empresaID = Math.trunc(parsed);
  }

  function ensureDomain() {
    var raw = normalize(getQueryParam("domain")).toLowerCase();
    if (isDomainValid(raw)) {
      state.domain = raw;
      return;
    }
    state.domain = "ventas";
  }

  function modulesForCurrentDomain() {
    return MODULES.filter(function (moduleConfig) {
      return moduleConfig.domain === state.domain;
    });
  }

  function currentModule() {
    var domainModules = modulesForCurrentDomain();
    if (domainModules.length === 0) {
      return null;
    }
    for (var i = 0; i < domainModules.length; i += 1) {
      if (domainModules[i].key === state.selectedModuleKey) {
        return domainModules[i];
      }
    }
    return domainModules[0];
  }

  function setOpsMessage(text, isError) {
    var box = document.getElementById("opsMsg");
    if (!box) {
      return;
    }
    box.textContent = normalize(text);
    box.className = isError ? "form-help erp-text-error" : "form-help erp-text-success";
  }

  function writeOutput(value) {
    var box = document.getElementById("outputBox");
    if (!box) {
      return;
    }
    if (typeof value === "string") {
      box.textContent = value;
      return;
    }
    box.textContent = JSON.stringify(value, null, 2);
  }

  function setDomainHeader() {
    var title = document.getElementById("erpDomainTitle");
    if (title) {
      title.textContent = DOMAIN_LABELS[state.domain] || "Dominio";
    }
  }

  function setDomainLinks() {
    var domainLinks = document.querySelectorAll("[data-domain-link]");
    for (var i = 0; i < domainLinks.length; i += 1) {
      var link = domainLinks[i];
      var domainKey = normalize(link.getAttribute("data-domain-link")).toLowerCase();
      if (!isDomainValid(domainKey)) {
        continue;
      }
      var target = new URL("/administrar_empresa/modulos_erp_dominio.html", window.location.origin);
      target.searchParams.set("empresa_id", String(state.empresaID));
      target.searchParams.set("domain", domainKey);
      link.setAttribute("href", target.pathname + target.search);
      if (domainKey === state.domain) {
        link.classList.add("active");
      } else {
        link.classList.remove("active");
      }
    }

    var backLink = document.getElementById("backToHub");
    if (backLink) {
      var hub = new URL("/administrar_empresa/modulos_erp_extendido.html", window.location.origin);
      hub.searchParams.set("empresa_id", String(state.empresaID));
      backLink.setAttribute("href", hub.pathname + hub.search);
    }
  }

  async function requestJSON(url, options) {
    var response = await fetch(url, options || {});
    var text = await response.text();
    if (!response.ok) {
      throw new Error(text || ("HTTP " + response.status));
    }
    if (!normalize(text)) {
      return {};
    }
    try {
      return JSON.parse(text);
    } catch (error) {
      return { raw: text };
    }
  }

  function buildURL(endpoint, params) {
    var url = new URL(endpoint, window.location.origin);
    url.searchParams.set("empresa_id", String(state.empresaID));
    var keys = Object.keys(params || {});
    for (var i = 0; i < keys.length; i += 1) {
      var key = keys[i];
      var value = params[key];
      if (value == null) {
        continue;
      }
      if (normalize(value) === "") {
        continue;
      }
      url.searchParams.set(key, String(value));
    }
    return url.pathname + url.search;
  }

  function moduleTemplate(moduleConfig) {
    var base = {
      empresa_id: state.empresaID,
      observaciones: "registro inicial"
    };

    if (!moduleConfig) {
      return base;
    }

    switch (moduleConfig.key) {
      case "ventas_cotizaciones":
      case "ventas_pedidos":
        return Object.assign({}, base, { cliente_nombre: "Cliente demo", subtotal: 0, total: 0 });
      case "ventas_devoluciones":
        return Object.assign({}, base, { motivo: "Ajuste operativo", subtotal: 0, total: 0 });

      case "finanzas_plan_cuentas":
        return Object.assign({}, base, { codigo: "110505", nombre: "Caja general", tipo_cuenta: "activo" });
      case "finanzas_cxc":
        return Object.assign({}, base, { cliente_nombre: "Cliente demo", documento_codigo: "FV-1001", valor_original: 0, saldo: 0 });
      case "finanzas_cxp":
        return Object.assign({}, base, { proveedor_nombre: "Proveedor demo", documento_codigo: "CP-1001", valor_original: 0, saldo: 0 });

      case "inventario_lotes_series":
        return Object.assign({}, base, { producto_id: 1, codigo_lote_serie: "L-001", cantidad_inicial: 0, cantidad_disponible: 0 });
      case "compras_devoluciones_proveedor":
        return Object.assign({}, base, { proveedor_nombre: "Proveedor demo", motivo: "Devolucion de prueba", subtotal: 0, total: 0 });
      case "rrhh_vacaciones_licencias":
        return Object.assign({}, base, { empleado_nombre: "Empleado demo", tipo_novedad: "vacacion", fecha_inicio: "2026-04-06", fecha_fin: "2026-04-07" });

      case "crm_leads":
        return Object.assign({}, base, { nombre: "Lead demo", canal_origen: "web" });
      case "crm_interacciones":
        return Object.assign({}, base, { tipo_interaccion: "llamada", resumen: "Contacto inicial" });
      case "crm_campanas":
        return Object.assign({}, base, { nombre: "Campana demo", canal: "email" });

      case "produccion_bom":
        return Object.assign({}, base, { codigo: "BOM-001", producto_nombre: "Producto demo" });
      case "produccion_bom_detalle":
        return Object.assign({}, base, { bom_id: 1, insumo_nombre: "Insumo demo", cantidad: 1 });
      case "produccion_ordenes":
        return Object.assign({}, base, { producto_nombre: "Producto demo", cantidad_programada: 1 });

      case "logistica_transportistas":
        return Object.assign({}, base, { nombre: "Transportista demo" });
      case "logistica_rutas":
        return Object.assign({}, base, { nombre: "Ruta demo", origen: "Origen", destino: "Destino" });
      case "logistica_envios":
        return Object.assign({}, base, { cliente_nombre: "Cliente demo", direccion_entrega: "Direccion demo" });

      case "documentos_gestion":
        return Object.assign({}, base, { modulo: "ventas", entidad: "factura", nombre_documento: "Documento demo" });
      case "documentos_firmas":
        return Object.assign({}, base, { documento_gestion_id: 1, tipo_firma: "digital", firmante_nombre: "Firmante demo" });
      case "integraciones_apis":
        return Object.assign({}, base, { nombre_integracion: "API demo", tipo_integracion: "rest" });
      case "integraciones_bancos":
        return Object.assign({}, base, { banco_nombre: "Banco demo", numero_cuenta: "000123456" });
      case "facturacion_dian":
        return Object.assign({}, base, { nit: "900123456", razon_social: "Empresa Demo SAS", tipo_ambiente: "habilitacion" });

      default:
        return base;
    }
  }

  function formatFieldLabel(fieldName) {
    var source = normalize(fieldName).replace(/_/g, " ");
    if (!source) {
      return "Campo";
    }
    return source.charAt(0).toUpperCase() + source.slice(1);
  }

  function isFieldRequired(moduleConfig, fieldName) {
    if (!moduleConfig || !fieldName) {
      return false;
    }
    var required = MODULE_REQUIRED_FIELDS[moduleConfig.key] || [];
    for (var i = 0; i < required.length; i += 1) {
      if (required[i] === fieldName) {
        return true;
      }
    }
    return false;
  }

  function isDateFieldName(fieldName) {
    var key = normalize(fieldName).toLowerCase();
    return key.indexOf("fecha_") === 0 || key.indexOf("_fecha") > 0;
  }

  function isLongTextFieldName(fieldName) {
    var key = normalize(fieldName).toLowerCase();
    return (
      key.indexOf("observ") >= 0 ||
      key.indexOf("resumen") >= 0 ||
      key.indexOf("motivo") >= 0 ||
      key.indexOf("descripcion") >= 0 ||
      key.indexOf("direccion") >= 0 ||
      key.indexOf("json") >= 0
    );
  }

  function inferFieldType(fieldName, value) {
    var preset = FIELD_PRESETS[fieldName] || null;
    if (preset && preset.type) {
      return preset.type;
    }
    if (typeof value === "number") {
      return "number";
    }
    if (typeof value === "boolean") {
      return "checkbox";
    }
    if (isDateFieldName(fieldName)) {
      return "date";
    }
    var key = normalize(fieldName).toLowerCase();
    if (key.indexOf("email") >= 0) {
      return "email";
    }
    if (isLongTextFieldName(fieldName)) {
      return "textarea";
    }
    return "text";
  }

  function guidedFieldID(fieldName) {
    return "guidedField_" + fieldName;
  }

  function quickFieldID(fieldName) {
    return "quickParam_" + fieldName;
  }

  function buildGuidedFields(moduleConfig) {
    var template = moduleTemplate(moduleConfig);
    var keys = Object.keys(template || {}).filter(function (fieldName) {
      return fieldName !== "empresa_id" && fieldName !== "id";
    });
    var fields = [];
    for (var i = 0; i < keys.length; i += 1) {
      var fieldName = keys[i];
      var value = template[fieldName];
      var preset = FIELD_PRESETS[fieldName] || null;
      var type = inferFieldType(fieldName, value);
      fields.push({
        name: fieldName,
        label: formatFieldLabel(fieldName),
        type: type,
        required: isFieldRequired(moduleConfig, fieldName),
        options: preset && Array.isArray(preset.options) ? preset.options : null,
        defaultValue: value,
        min: NON_NEGATIVE_FIELDS[fieldName] ? 0 : null
      });
    }
    return fields;
  }

  function setControlValue(control, fieldDef, value) {
    if (!control) {
      return;
    }
    if (fieldDef.type === "checkbox") {
      control.checked = !!value;
      return;
    }
    if (value == null) {
      control.value = "";
      return;
    }
    control.value = String(value);
  }

  function parseControlValue(control, fieldDef) {
    if (!control || !fieldDef) {
      return "";
    }
    if (fieldDef.type === "checkbox") {
      return !!control.checked;
    }
    var raw = normalize(control.value);
    if (fieldDef.type === "number") {
      if (!raw) {
        return "";
      }
      return Number(raw);
    }
    return raw;
  }

  function buildFieldWrapper(fieldDef, controlID, initialValue, asGuidedField) {
    var wrapper = document.createElement("div");
    wrapper.className = "form-col erp-guided-field";
    if (asGuidedField !== false) {
      wrapper.setAttribute("data-guided-wrapper", fieldDef.name);
    }

    var label = document.createElement("label");
    label.className = "form-label";
    label.setAttribute("for", controlID);
    label.textContent = fieldDef.required ? (fieldDef.label + " *") : fieldDef.label;
    wrapper.appendChild(label);

    var control;
    if (fieldDef.type === "textarea") {
      control = document.createElement("textarea");
      control.className = "form-textarea erp-guided-textarea";
      control.rows = 3;
    } else if (fieldDef.type === "select") {
      control = document.createElement("select");
      control.className = "form-input";
      var options = fieldDef.options || [];
      for (var i = 0; i < options.length; i += 1) {
        var option = document.createElement("option");
        option.value = options[i];
        option.textContent = options[i];
        control.appendChild(option);
      }
    } else {
      control = document.createElement("input");
      control.className = "form-input";
      if (fieldDef.type === "number") {
        control.type = "number";
        control.step = "any";
      } else if (fieldDef.type === "date") {
        control.type = "date";
      } else if (fieldDef.type === "email") {
        control.type = "email";
      } else {
        control.type = "text";
      }
    }

    if (fieldDef.min != null) {
      control.setAttribute("min", String(fieldDef.min));
    }
    control.id = controlID;
    control.setAttribute("data-field-name", fieldDef.name);
    control.setAttribute("data-field-type", fieldDef.type);
    if (fieldDef.required) {
      control.setAttribute("data-required", "1");
    }
    setControlValue(control, fieldDef, initialValue);
    wrapper.appendChild(control);

    var help = document.createElement("div");
    help.className = "form-help erp-guided-help";
    if (fieldDef.type === "date") {
      help.textContent = "Formato esperado: AAAA-MM-DD";
    } else if (fieldDef.type === "number") {
      help.textContent = "Solo valores numericos.";
    } else {
      help.textContent = "Campo operativo del modulo.";
    }
    wrapper.appendChild(help);

    return wrapper;
  }

  function clearGuidedValidationUI() {
    var validationBox = document.getElementById("guidedValidation");
    if (validationBox) {
      validationBox.classList.add("erp-hidden");
      validationBox.innerHTML = "";
    }
    var wrappers = document.querySelectorAll("[data-guided-wrapper]");
    for (var i = 0; i < wrappers.length; i += 1) {
      wrappers[i].classList.remove("has-error");
    }
  }

  function showGuidedValidation(errors) {
    var validationBox = document.getElementById("guidedValidation");
    if (!validationBox) {
      return;
    }
    if (!Array.isArray(errors) || errors.length === 0) {
      clearGuidedValidationUI();
      return;
    }

    clearGuidedValidationUI();
    var html = ["<strong>Validaciones pendientes:</strong>", "<ul>"];
    for (var i = 0; i < errors.length; i += 1) {
      html.push("<li>" + errors[i].message + "</li>");
      if (errors[i].field) {
        var wrapper = document.querySelector('[data-guided-wrapper="' + errors[i].field + '"]');
        if (wrapper) {
          wrapper.classList.add("has-error");
        }
      }
    }
    html.push("</ul>");
    validationBox.innerHTML = html.join("");
    validationBox.classList.remove("erp-hidden");
  }

  function renderGuidedFields() {
    var box = document.getElementById("guidedFields");
    if (!box) {
      return;
    }
    var moduleConfig = currentModule();
    if (!moduleConfig) {
      box.innerHTML = '<div class="erp-placeholder">No hay modulo seleccionado.</div>';
      state.guidedFields = [];
      return;
    }

    state.guidedFields = buildGuidedFields(moduleConfig);
    box.innerHTML = "";
    for (var i = 0; i < state.guidedFields.length; i += 1) {
      var fieldDef = state.guidedFields[i];
      var controlID = guidedFieldID(fieldDef.name);
      box.appendChild(buildFieldWrapper(fieldDef, controlID, fieldDef.defaultValue, true));
    }
    clearGuidedValidationUI();
  }

  function readGuidedPayload(isUpdate) {
    var payload = { empresa_id: state.empresaID };
    for (var i = 0; i < state.guidedFields.length; i += 1) {
      var fieldDef = state.guidedFields[i];
      var control = document.getElementById(guidedFieldID(fieldDef.name));
      payload[fieldDef.name] = parseControlValue(control, fieldDef);
    }
    if (isUpdate) {
      var recordID = toInt((document.getElementById("guidedRecordID") || {}).value, 0);
      payload.id = recordID;
    }
    return payload;
  }

  function validateGuidedPayload(moduleConfig, payload, isUpdate) {
    var errors = [];
    if (!moduleConfig) {
      errors.push({ field: "", message: "Selecciona un modulo antes de guardar." });
      return errors;
    }

    for (var i = 0; i < state.guidedFields.length; i += 1) {
      var fieldDef = state.guidedFields[i];
      var value = payload[fieldDef.name];
      if (fieldDef.required && (value === "" || value == null)) {
        errors.push({ field: fieldDef.name, message: fieldDef.label + " es obligatorio." });
        continue;
      }

      if (fieldDef.type === "number" && value !== "") {
        if (!Number.isFinite(value)) {
          errors.push({ field: fieldDef.name, message: fieldDef.label + " debe ser numerico." });
          continue;
        }
        if (fieldDef.min != null && value < fieldDef.min) {
          errors.push({ field: fieldDef.name, message: fieldDef.label + " no puede ser negativo." });
        }
      }

      if (fieldDef.type === "email" && value) {
        var emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailPattern.test(String(value))) {
          errors.push({ field: fieldDef.name, message: fieldDef.label + " no tiene formato de correo valido." });
        }
      }

      if (fieldDef.type === "date" && value) {
        var datePattern = /^\d{4}-\d{2}-\d{2}$/;
        if (!datePattern.test(String(value))) {
          errors.push({ field: fieldDef.name, message: fieldDef.label + " debe usar formato AAAA-MM-DD." });
        }
      }
    }

    if (isUpdate && (!Number.isFinite(payload.id) || payload.id <= 0)) {
      errors.push({ field: "", message: "Debes ingresar ID valido para actualizar." });
    }

    if (payload.total !== "" && payload.subtotal !== "" && Number.isFinite(payload.total) && Number.isFinite(payload.subtotal) && payload.total < payload.subtotal) {
      errors.push({ field: "total", message: "Total no puede ser menor que subtotal." });
    }

    if (payload.saldo !== "" && payload.valor_original !== "" && Number.isFinite(payload.saldo) && Number.isFinite(payload.valor_original) && payload.saldo > payload.valor_original) {
      errors.push({ field: "saldo", message: "Saldo no puede superar el valor original." });
    }

    if (moduleConfig.key === "rrhh_vacaciones_licencias") {
      var inicio = normalize(payload.fecha_inicio);
      var fin = normalize(payload.fecha_fin);
      if (inicio && fin && fin < inicio) {
        errors.push({ field: "fecha_fin", message: "Fecha fin no puede ser menor a fecha inicio." });
      }
    }

    if (moduleConfig.key === "facturacion_dian") {
      var nit = normalize(payload.nit);
      if (nit && !/^\d{6,15}$/.test(nit)) {
        errors.push({ field: "nit", message: "NIT debe contener solo digitos (6 a 15)." });
      }
    }

    return errors;
  }

  function copyGuidedToJSON(jsonAction) {
    var payloadArea = document.getElementById("payloadArea");
    if (!payloadArea) {
      return;
    }
    var payload = readGuidedPayload(jsonAction === "update");
    payloadArea.value = JSON.stringify(payload, null, 2);

    var crudAction = document.getElementById("crudAction");
    if (crudAction) {
      crudAction.value = jsonAction || "create";
    }
    var recordID = document.getElementById("recordID");
    if (recordID) {
      recordID.value = Number.isFinite(payload.id) && payload.id > 0 ? String(payload.id) : "";
    }
  }

  function fillGuidedFromRow(rowSnapshot, keepID) {
    if (!rowSnapshot || typeof rowSnapshot !== "object") {
      return;
    }
    for (var i = 0; i < state.guidedFields.length; i += 1) {
      var fieldDef = state.guidedFields[i];
      var control = document.getElementById(guidedFieldID(fieldDef.name));
      if (!control) {
        continue;
      }
      if (Object.prototype.hasOwnProperty.call(rowSnapshot, fieldDef.name)) {
        setControlValue(control, fieldDef, rowSnapshot[fieldDef.name]);
      }
    }
    var guidedID = document.getElementById("guidedRecordID");
    if (guidedID) {
      guidedID.value = keepID && Number((rowSnapshot && rowSnapshot.id) || 0) > 0 ? String(rowSnapshot.id) : "";
    }
    clearGuidedValidationUI();
  }

  function resetGuidedForm() {
    renderGuidedFields();
    copyGuidedToJSON("create");
    setOpsMessage("Formulario guiado reiniciado.", false);
  }

  async function executeGuidedCrud(isUpdate) {
    var moduleConfig = currentModule();
    if (!moduleConfig) {
      setOpsMessage("No hay modulo seleccionado.", true);
      return;
    }

    var payload = readGuidedPayload(isUpdate);
    var errors = validateGuidedPayload(moduleConfig, payload, isUpdate);
    showGuidedValidation(errors);
    if (errors.length > 0) {
      setOpsMessage("Corrige los errores del formulario guiado.", true);
      writeOutput({ errores: errors });
      return;
    }

    try {
      var data = await requestJSON(moduleConfig.endpoint, {
        method: isUpdate ? "PUT" : "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });
      writeOutput(data);
      setOpsMessage(isUpdate ? "Registro actualizado con formulario guiado." : "Registro creado con formulario guiado.", false);
      copyGuidedToJSON(isUpdate ? "update" : "create");
      clearGuidedValidationUI();
      await listRows();
    } catch (error) {
      setOpsMessage(error.message || "No se pudo guardar con formulario guiado", true);
      writeOutput({ error: error.message || "No se pudo guardar con formulario guiado" });
    }
  }

  function quickActionsForModule(moduleConfig) {
    var actions = [];
    if (!moduleConfig) {
      return actions;
    }

    actions.push({
      key: "detalle",
      label: "Consultar detalle por ID",
      params: [{ name: "id", label: "ID", type: "number", required: true, min: 1 }],
      execute: function (values) { return detailRecord(values.id); }
    });

    actions.push({
      key: "activar",
      label: "Activar registro",
      params: [{ name: "id", label: "ID", type: "number", required: true, min: 1 }],
      execute: function (values) { return toggleEstado(values.id, "activar"); }
    });

    actions.push({
      key: "desactivar",
      label: "Desactivar registro",
      params: [{ name: "id", label: "ID", type: "number", required: true, min: 1 }],
      execute: function (values) { return toggleEstado(values.id, "desactivar"); }
    });

    if (supportsStateMachineActions(moduleConfig)) {
      actions.push({
        key: "transiciones",
        label: "Consultar transiciones por ID",
        params: [{ name: "id", label: "ID", type: "number", required: true, min: 1 }],
        execute: function (values) { return runModuleAction("transiciones", "GET", values.id, null); }
      });

      actions.push({
        key: "transicionar",
        label: "Transicionar estado",
        params: [
          { name: "id", label: "ID", type: "number", required: true, min: 1 },
          { name: "nuevo_estado", label: "Nuevo estado", type: "text", required: true },
          { name: "motivo", label: "Motivo", type: "text", required: false }
        ],
        execute: function (values) {
          var payload = { nuevo_estado: values.nuevo_estado };
          if (values.motivo) {
            payload.motivo = values.motivo;
          }
          return runModuleAction("transicionar", "PUT", values.id, payload);
        }
      });
    }

    if (supportsIntegrationActions(moduleConfig)) {
      actions.push({
        key: "health",
        label: "Health check por ID",
        params: [{ name: "id", label: "ID", type: "number", required: true, min: 1 }],
        execute: function (values) { return runModuleAction("health_check", "GET", values.id, null); }
      });

      actions.push({
        key: "sync",
        label: "Sincronizacion manual por ID",
        params: [{ name: "id", label: "ID", type: "number", required: true, min: 1 }],
        execute: function (values) { return runModuleAction("sync_manual", "POST", values.id, null); }
      });

      actions.push({
        key: "estado",
        label: "Consultar estado por ID",
        params: [{ name: "id", label: "ID", type: "number", required: true, min: 1 }],
        execute: function (values) { return runModuleAction("estado", "GET", values.id, null); }
      });

      actions.push({
        key: "monitoreo",
        label: "Monitoreo de conector por ID",
        params: [{ name: "id", label: "ID", type: "number", required: true, min: 1 }],
        execute: function (values) { return runModuleAction("monitoreo", "GET", values.id, null); }
      });
    }

    if (moduleConfig.key === "ventas_cotizaciones") {
      actions.push({
        key: "embudo",
        label: "Ver embudo comercial",
        params: [],
        execute: function () { return runModuleAction("embudo", "GET", null, null); }
      });
    }

    if (moduleConfig.key === "finanzas_plan_cuentas") {
      actions.push({
        key: "plantillas",
        label: "Consultar plantillas",
        params: [],
        execute: function () { return runModuleAction("plantillas", "GET", null, null); }
      });

      actions.push({
        key: "aplicar_plantilla",
        label: "Aplicar plantilla contable",
        params: [{ name: "tipo_empresa", label: "Tipo empresa", type: "text", required: true }],
        execute: function (values) { return runModuleAction("aplicar_plantilla", "POST", null, { tipo_empresa: values.tipo_empresa }); }
      });
    }

    if (moduleConfig.key === "finanzas_cxc" || moduleConfig.key === "finanzas_cxp") {
      actions.push({
        key: "conciliar_pagos",
        label: "Conciliar pagos por ID",
        params: [{ name: "id", label: "ID", type: "number", required: true, min: 1 }],
        execute: function (values) { return runModuleAction("conciliar_pagos", "POST", values.id, null); }
      });

      actions.push({
        key: "validar_cierre_periodo",
        label: "Validar cierre de periodo",
        params: [],
        execute: function () { return runModuleAction("validar_cierre_periodo", "GET", null, null); }
      });
    }

    if (moduleConfig.key === "rrhh_vacaciones_licencias") {
      actions.push({
        key: "resumen_saldo",
        label: "Resumen de saldo RRHH",
        params: [{ name: "empleado_id", label: "Empleado ID", type: "number", required: false, min: 1 }],
        execute: function (values) {
          var payload = values.empleado_id ? { empleado_id: values.empleado_id } : null;
          return runModuleAction("resumen_saldo", payload ? "POST" : "GET", null, payload);
        }
      });
    }

    if (moduleConfig.key === "produccion_ordenes") {
      actions.push({
        key: "plan_capacidad",
        label: "Consultar plan de capacidad",
        params: [],
        execute: function () { return runModuleAction("plan_capacidad", "GET", null, null); }
      });
    }

    if (moduleConfig.key === "logistica_envios") {
      actions.push({
        key: "seguimiento_hitos",
        label: "Seguimiento de hitos",
        params: [],
        execute: function () { return runModuleAction("seguimiento_hitos", "GET", null, null); }
      });
    }

    if (moduleConfig.dian) {
      actions.push({
        key: "dian_guia_onboarding",
        label: "DIAN guía onboarding",
        params: [],
        execute: function () { return runDianAction("guia_onboarding", "GET"); }
      });
      actions.push({
        key: "dian_checklist",
        label: "DIAN checklist",
        params: [],
        execute: function () { return runDianAction("checklist", "GET"); }
      });
      actions.push({
        key: "dian_validar",
        label: "DIAN validar",
        params: [],
        execute: function () { return runDianAction("validar", "GET"); }
      });
      actions.push({
        key: "dian_validar_credenciales",
        label: "DIAN validar credenciales",
        params: [],
        execute: function () { return runDianAction("validar_credenciales", "POST", {}); }
      });
      actions.push({
        key: "dian_pruebas",
        label: "Pruebas Dian",
        params: [{ name: "simular", label: "Simular sin enviar", type: "checkbox", required: false }],
        execute: function (values) {
          values = values || {};
          values.detener_en_error = true;
          return runDianAction("pruebas_dian", "POST", values);
        }
      });
      actions.push({
        key: "dian_cufe",
        label: "DIAN CUFE demo",
        params: [
          { name: "documento_codigo", label: "Documento codigo", type: "text", required: false },
          { name: "total", label: "Total", type: "number", required: false, min: 0 }
        ],
        execute: function (values) { return runDianAction("generar_cufe_demo", "POST", values); }
      });
      actions.push({
        key: "dian_xml",
        label: "DIAN XML demo",
        params: [{ name: "documento_codigo", label: "Documento codigo", type: "text", required: false }],
        execute: function (values) { return runDianAction("generar_xml_demo", "POST", values); }
      });
    }

    return actions;
  }

  function currentQuickAction() {
    for (var i = 0; i < state.quickActions.length; i += 1) {
      if (state.quickActions[i].key === state.quickActionKey) {
        return state.quickActions[i];
      }
    }
    return state.quickActions.length > 0 ? state.quickActions[0] : null;
  }

  function renderQuickActionParams() {
    var paramsBox = document.getElementById("quickActionParams");
    if (!paramsBox) {
      return;
    }
    paramsBox.innerHTML = "";

    var actionConfig = currentQuickAction();
    if (!actionConfig) {
      paramsBox.innerHTML = '<div class="erp-placeholder">No hay acciones rapidas para este modulo.</div>';
      return;
    }

    var params = actionConfig.params || [];
    if (params.length === 0) {
      paramsBox.innerHTML = '<div class="erp-placeholder">La accion seleccionada no requiere parametros.</div>';
      return;
    }

    for (var i = 0; i < params.length; i += 1) {
      var fieldDef = {
        name: params[i].name,
        label: params[i].label || formatFieldLabel(params[i].name),
        type: params[i].type || "text",
        required: !!params[i].required,
        options: params[i].options || null,
        defaultValue: params[i].defaultValue || "",
        min: params[i].min == null ? null : params[i].min
      };
      paramsBox.appendChild(buildFieldWrapper(fieldDef, quickFieldID(fieldDef.name), fieldDef.defaultValue, false));
    }
  }

  function renderQuickActions() {
    var selector = document.getElementById("quickActionSelector");
    if (!selector) {
      return;
    }

    var moduleConfig = currentModule();
    state.quickActions = quickActionsForModule(moduleConfig);
    selector.innerHTML = "";

    for (var i = 0; i < state.quickActions.length; i += 1) {
      var actionConfig = state.quickActions[i];
      var option = document.createElement("option");
      option.value = actionConfig.key;
      option.textContent = actionConfig.label;
      selector.appendChild(option);
    }

    state.quickActionKey = state.quickActions.length > 0 ? state.quickActions[0].key : "";
    selector.value = state.quickActionKey;
    selector.onchange = function () {
      state.quickActionKey = normalize(selector.value);
      renderQuickActionParams();
    };

    renderQuickActionParams();
  }

  function readQuickActionValues(actionConfig) {
    var params = (actionConfig && actionConfig.params) || [];
    var values = {};
    var errors = [];

    for (var i = 0; i < params.length; i += 1) {
      var param = params[i];
      var fieldDef = {
        name: param.name,
        label: param.label || formatFieldLabel(param.name),
        type: param.type || "text",
        required: !!param.required,
        min: param.min == null ? null : param.min
      };
      var control = document.getElementById(quickFieldID(param.name));
      var value = parseControlValue(control, fieldDef);
      values[param.name] = value;

      if (fieldDef.required && (value === "" || value == null)) {
        errors.push(fieldDef.label + " es obligatorio.");
      }
      if (fieldDef.type === "number" && value !== "") {
        if (!Number.isFinite(value)) {
          errors.push(fieldDef.label + " debe ser numerico.");
          continue;
        }
        if (fieldDef.min != null && value < fieldDef.min) {
          errors.push(fieldDef.label + " no puede ser menor a " + fieldDef.min + ".");
        }
      }
    }

    return { values: values, errors: errors };
  }

  async function executeQuickAction() {
    var actionConfig = currentQuickAction();
    if (!actionConfig) {
      setOpsMessage("No hay accion rapida disponible para este modulo.", true);
      return;
    }

    var parsed = readQuickActionValues(actionConfig);
    if (parsed.errors.length > 0) {
      setOpsMessage(parsed.errors.join(" "), true);
      writeOutput({ errores: parsed.errors });
      return;
    }

    try {
      await actionConfig.execute(parsed.values);
    } catch (error) {
      setOpsMessage(error.message || "No se pudo ejecutar accion rapida", true);
      writeOutput({ error: error.message || "No se pudo ejecutar accion rapida" });
    }
  }

  function renderGuideList(items) {
    if (!Array.isArray(items) || items.length === 0) {
      return "";
    }
    var html = ["<ul>"];
    for (var i = 0; i < items.length; i += 1) {
      html.push("<li>" + items[i] + "</li>");
    }
    html.push("</ul>");
    return html.join("");
  }

  function renderDomainGuide() {
    var box = document.getElementById("domainGuide");
    if (!box) {
      return;
    }
    var guide = DOMAIN_GUIDES[state.domain];
    if (!guide) {
      box.innerHTML = '<div class="erp-placeholder">No hay guia operativa para este dominio.</div>';
      return;
    }

    var parts = [];
    parts.push('<div class="erp-guide-block">');
    parts.push('<h3 class="erp-guide-title">Objetivo operativo</h3>');
    parts.push('<p class="erp-guide-text">' + guide.objetivo + "</p>");
    parts.push("</div>");

    parts.push('<div class="erp-guide-block">');
    parts.push('<h3 class="erp-guide-title">Flujo recomendado</h3>');
    parts.push(renderGuideList(guide.flujo));
    parts.push("</div>");

    parts.push('<div class="erp-guide-block">');
    parts.push('<h3 class="erp-guide-title">Controles clave</h3>');
    parts.push(renderGuideList(guide.controles));
    parts.push("</div>");

    box.innerHTML = parts.join("");
  }

  function loadTemplateToPayload() {
    var moduleConfig = currentModule();
    var payloadArea = document.getElementById("payloadArea");
    if (!payloadArea) {
      return;
    }
    payloadArea.value = JSON.stringify(moduleTemplate(moduleConfig), null, 2);
    var recordID = document.getElementById("recordID");
    var crudAction = document.getElementById("crudAction");
    if (recordID) {
      recordID.value = "";
    }
    if (crudAction) {
      crudAction.value = "create";
    }
  }

  function supportsStateMachineActions(moduleConfig) {
    if (!moduleConfig) {
      return false;
    }
    return !!STATE_MACHINE_MODULE_KEYS[moduleConfig.key];
  }

  function supportsIntegrationActions(moduleConfig) {
    if (!moduleConfig) {
      return false;
    }
    return !!INTEGRATION_MODULE_KEYS[moduleConfig.key];
  }

  function resolveStateColumnForModule(moduleConfig) {
    if (!moduleConfig) {
      return "";
    }
    switch (moduleConfig.key) {
      case "ventas_cotizaciones":
        return "estado_documento";
      case "ventas_pedidos":
        return "estado_pedido";
      case "ventas_devoluciones":
        return "estado_devolucion";
      case "crm_leads":
        return "estado_lead";
      case "crm_interacciones":
        return "estado_interaccion";
      case "crm_campanas":
        return "estado_campana";
      default:
        return "";
    }
  }

  async function runModuleAction(action, method, id, payload) {
    var moduleConfig = currentModule();
    if (!moduleConfig) {
      return;
    }

    var requestMethod = normalize(method || "GET").toUpperCase();
    var query = { action: action };
    if (Number.isFinite(id) && id > 0) {
      query.id = id;
    }

    var options = { method: requestMethod };
    if (payload && (requestMethod === "POST" || requestMethod === "PUT" || requestMethod === "PATCH")) {
      var bodyPayload = Object.assign({}, payload, { empresa_id: state.empresaID });
      if (Number.isFinite(id) && id > 0) {
        bodyPayload.id = id;
      }
      options.headers = { "Content-Type": "application/json" };
      options.body = JSON.stringify(bodyPayload);
    }

    try {
      var url = buildURL(moduleConfig.endpoint, query);
      var data = await requestJSON(url, options);
      writeOutput(data);
      setOpsMessage("Accion " + action + " ejecutada.", false);
      if (action === "health_check" || action === "sync_manual" || action === "transicionar") {
        await listRows();
      }
    } catch (error) {
      setOpsMessage(error.message || ("No se pudo ejecutar accion " + action), true);
      writeOutput({ error: error.message || ("No se pudo ejecutar accion " + action) });
    }
  }

  function promptStateTransition(id, rowSnapshot) {
    var moduleConfig = currentModule();
    if (!supportsStateMachineActions(moduleConfig)) {
      return;
    }

    var stateColumn = resolveStateColumnForModule(moduleConfig);
    var currentState = normalize((rowSnapshot || {})[stateColumn]).toLowerCase();
    var message = "Nuevo estado para ID " + id;
    if (currentState) {
      message += " (actual: " + currentState + ")";
    }
    var nuevoEstado = normalize(window.prompt(message + ":", ""));
    if (!nuevoEstado) {
      return;
    }
    var motivo = normalize(window.prompt("Motivo de transicion (opcional):", ""));
    var payload = { nuevo_estado: nuevoEstado };
    if (motivo) {
      payload.motivo = motivo;
    }
    runModuleAction("transicionar", "PUT", id, payload);
  }

  function renderRows() {
    var box = document.getElementById("rowsList");
    if (!box) {
      return;
    }

    if (!Array.isArray(state.rows) || state.rows.length === 0) {
      box.innerHTML = '<div class="erp-placeholder">No hay registros para los filtros actuales.</div>';
      return;
    }

    var allKeysMap = {};
    for (var i = 0; i < state.rows.length; i += 1) {
      var row = state.rows[i] || {};
      var rowKeys = Object.keys(row);
      for (var k = 0; k < rowKeys.length; k += 1) {
        allKeysMap[rowKeys[k]] = true;
      }
    }

    var keys = Object.keys(allKeysMap).sort(function (a, b) {
      if (a === "id") {
        return -1;
      }
      if (b === "id") {
        return 1;
      }
      return a.localeCompare(b);
    });

    var moduleConfig = currentModule();

    var table = document.createElement("table");
    table.className = "table";

    var thead = document.createElement("thead");
    var headRow = document.createElement("tr");
    for (var h = 0; h < keys.length; h += 1) {
      var th = document.createElement("th");
      th.textContent = keys[h];
      headRow.appendChild(th);
    }
    var actionsTH = document.createElement("th");
    actionsTH.textContent = "Acciones";
    headRow.appendChild(actionsTH);
    thead.appendChild(headRow);
    table.appendChild(thead);

    var tbody = document.createElement("tbody");
    for (var r = 0; r < state.rows.length; r += 1) {
      var tableRow = state.rows[r] || {};
      var tr = document.createElement("tr");
      for (var c = 0; c < keys.length; c += 1) {
        var key = keys[c];
        var td = document.createElement("td");
        var value = tableRow[key];
        if (typeof value === "object" && value !== null) {
          td.textContent = JSON.stringify(value);
        } else {
          td.textContent = value == null ? "" : String(value);
        }
        tr.appendChild(td);
      }

      var actionTD = document.createElement("td");
      actionTD.className = "actions";
      var rowID = Number((tableRow && tableRow.id) || 0);

      if (Number.isFinite(rowID) && rowID > 0) {
        (function (safeID, rowSnapshot) {
          var detailBtn = document.createElement("button");
          detailBtn.type = "button";
          detailBtn.className = "btn secondary";
          detailBtn.textContent = "Detalle";
          detailBtn.onclick = function () {
            detailRecord(safeID);
          };
          actionTD.appendChild(detailBtn);

          var editBtn = document.createElement("button");
          editBtn.type = "button";
          editBtn.className = "btn secondary";
          editBtn.textContent = "Editar";
          editBtn.onclick = function () {
            var actionSelect = document.getElementById("crudAction");
            var recordID = document.getElementById("recordID");
            var payloadArea = document.getElementById("payloadArea");
            if (actionSelect) {
              actionSelect.value = "update";
            }
            if (recordID) {
              recordID.value = String(safeID);
            }
            if (payloadArea) {
              payloadArea.value = JSON.stringify(rowSnapshot, null, 2);
            }
            fillGuidedFromRow(rowSnapshot, true);
            writeOutput({ modo: "edicion", id: safeID, row: rowSnapshot });
          };
          actionTD.appendChild(editBtn);

          if (supportsIntegrationActions(moduleConfig)) {
            var healthBtn = document.createElement("button");
            healthBtn.type = "button";
            healthBtn.className = "btn secondary";
            healthBtn.textContent = "Health";
            healthBtn.onclick = function () {
              runModuleAction("health_check", "GET", safeID, null);
            };
            actionTD.appendChild(healthBtn);

            var syncBtn = document.createElement("button");
            syncBtn.type = "button";
            syncBtn.className = "btn secondary";
            syncBtn.textContent = "Sync";
            syncBtn.onclick = function () {
              runModuleAction("sync_manual", "POST", safeID, null);
            };
            actionTD.appendChild(syncBtn);

            var statusBtn = document.createElement("button");
            statusBtn.type = "button";
            statusBtn.className = "btn secondary";
            statusBtn.textContent = "Estado";
            statusBtn.onclick = function () {
              runModuleAction("estado", "GET", safeID, null);
            };
            actionTD.appendChild(statusBtn);
          }

          if (supportsStateMachineActions(moduleConfig)) {
            var flowBtn = document.createElement("button");
            flowBtn.type = "button";
            flowBtn.className = "btn secondary";
            flowBtn.textContent = "Transiciones";
            flowBtn.onclick = function () {
              runModuleAction("transiciones", "GET", safeID, null);
            };
            actionTD.appendChild(flowBtn);

            var moveBtn = document.createElement("button");
            moveBtn.type = "button";
            moveBtn.className = "btn secondary";
            moveBtn.textContent = "Transicionar";
            moveBtn.onclick = function () {
              promptStateTransition(safeID, rowSnapshot);
            };
            actionTD.appendChild(moveBtn);
          }

          var currentEstado = normalize(rowSnapshot.estado).toLowerCase();
          var action = currentEstado === "inactivo" ? "activar" : "desactivar";
          var toggleBtn = document.createElement("button");
          toggleBtn.type = "button";
          toggleBtn.className = "btn secondary";
          toggleBtn.textContent = action === "activar" ? "Activar" : "Desactivar";
          toggleBtn.onclick = function () {
            toggleEstado(safeID, action);
          };
          actionTD.appendChild(toggleBtn);

          var deleteBtn = document.createElement("button");
          deleteBtn.type = "button";
          deleteBtn.className = "btn danger";
          deleteBtn.textContent = "Eliminar";
          deleteBtn.onclick = function () {
            deleteRecord(safeID);
          };
          actionTD.appendChild(deleteBtn);
        })(rowID, tableRow);
      } else {
        actionTD.textContent = "-";
      }

      tr.appendChild(actionTD);
      tbody.appendChild(tr);
    }

    table.appendChild(tbody);
    box.innerHTML = "";
    box.appendChild(table);
  }

  async function listRows() {
    var moduleConfig = currentModule();
    if (!moduleConfig) {
      setOpsMessage("No hay modulos disponibles para este dominio.", true);
      writeOutput({ error: "No hay modulos disponibles para este dominio." });
      return;
    }

    var q = normalize((document.getElementById("busquedaQ") || {}).value);
    var limit = toInt((document.getElementById("limitRows") || {}).value, 100);
    var offset = toInt((document.getElementById("offsetRows") || {}).value, 0);
    var includeInactiveField = document.getElementById("includeInactive");
    var includeInactive = includeInactiveField && includeInactiveField.checked ? "1" : "";

    try {
      setOpsMessage("Consultando registros...", false);
      var url = buildURL(moduleConfig.endpoint, {
        q: q,
        limit: limit,
        offset: offset,
        include_inactive: includeInactive
      });
      var data = await requestJSON(url);
      state.rows = Array.isArray(data) ? data : [];
      renderRows();
      writeOutput({ dominio: DOMAIN_LABELS[state.domain], modulo: moduleConfig.label, total_registros: state.rows.length });
      setOpsMessage("Consulta exitosa.", false);
    } catch (error) {
      setOpsMessage(error.message || "No se pudo listar", true);
      writeOutput({ error: error.message || "No se pudo listar" });
    }
  }

  async function detailRecord(id) {
    var moduleConfig = currentModule();
    if (!moduleConfig) {
      return;
    }
    try {
      var url = buildURL(moduleConfig.endpoint, { action: "detalle", id: id });
      var data = await requestJSON(url);
      writeOutput(data);
      setOpsMessage("Detalle consultado.", false);
    } catch (error) {
      setOpsMessage(error.message || "No se pudo consultar detalle", true);
      writeOutput({ error: error.message || "No se pudo consultar detalle" });
    }
  }

  async function toggleEstado(id, action) {
    var moduleConfig = currentModule();
    if (!moduleConfig) {
      return;
    }
    try {
      var url = buildURL(moduleConfig.endpoint, { id: id, action: action });
      var data = await requestJSON(url, { method: "PUT" });
      writeOutput(data);
      setOpsMessage("Estado actualizado.", false);
      await listRows();
    } catch (error) {
      setOpsMessage(error.message || "No se pudo cambiar estado", true);
      writeOutput({ error: error.message || "No se pudo cambiar estado" });
    }
  }

  async function deleteRecord(id) {
    var moduleConfig = currentModule();
    if (!moduleConfig) {
      return;
    }
    if (!window.confirm("\u00BFEliminar registro ID " + id + "?")) {
      return;
    }
    try {
      var url = buildURL(moduleConfig.endpoint, { id: id });
      var data = await requestJSON(url, { method: "DELETE" });
      writeOutput(data);
      setOpsMessage("Registro eliminado logicamente.", false);
      await listRows();
    } catch (error) {
      setOpsMessage(error.message || "No se pudo eliminar", true);
      writeOutput({ error: error.message || "No se pudo eliminar" });
    }
  }

  function parsePayloadArea(idRequired) {
    var raw = normalize((document.getElementById("payloadArea") || {}).value);
    if (!raw) {
      throw new Error("Debes enviar payload JSON");
    }
    var payload;
    try {
      payload = JSON.parse(raw);
    } catch (error) {
      throw new Error("Payload JSON invalido");
    }
    if (!payload || typeof payload !== "object" || Array.isArray(payload)) {
      throw new Error("Payload debe ser un objeto JSON");
    }
    payload.empresa_id = state.empresaID;
    if (idRequired) {
      var id = toInt((document.getElementById("recordID") || {}).value, 0);
      if (id <= 0) {
        throw new Error("ID obligatorio para actualizar");
      }
      payload.id = id;
    }
    return payload;
  }

  async function executeCrud(event) {
    event.preventDefault();
    var moduleConfig = currentModule();
    if (!moduleConfig) {
      return;
    }

    var action = normalize((document.getElementById("crudAction") || {}).value).toLowerCase();
    var isUpdate = action === "update";
    try {
      var payload = parsePayloadArea(isUpdate);
      var data = await requestJSON(moduleConfig.endpoint, {
        method: isUpdate ? "PUT" : "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });
      writeOutput(data);
      setOpsMessage(isUpdate ? "Registro actualizado." : "Registro creado.", false);
      await listRows();
    } catch (error) {
      setOpsMessage(error.message || "No se pudo ejecutar operacion", true);
      writeOutput({ error: error.message || "No se pudo ejecutar operacion" });
    }
  }

  function parseDIANPayload(customPayload) {
    var payload = customPayload;
    if (!payload || typeof payload !== "object" || Array.isArray(payload)) {
      var raw = normalize((document.getElementById("dianPayload") || {}).value);
      if (!raw) {
        payload = {};
      } else {
        try {
          payload = JSON.parse(raw);
        } catch (error) {
          throw new Error("Payload DIAN invalido");
        }
      }
    }
    if (!payload || typeof payload !== "object" || Array.isArray(payload)) {
      throw new Error("Payload DIAN debe ser objeto JSON");
    }
    payload.empresa_id = state.empresaID;
    return payload;
  }

  async function runDianAction(action, method, payloadOverride) {
    var moduleConfig = currentModule();
    if (!moduleConfig || !moduleConfig.dian) {
      setOpsMessage("Selecciona el modulo de DIAN para usar estas herramientas.", true);
      return;
    }

    try {
      var url = buildURL(moduleConfig.endpoint, { action: action });
      var data;
      if (method === "GET") {
        data = await requestJSON(url);
      } else {
        var payload = parseDIANPayload(payloadOverride);
        data = await requestJSON(url, {
          method: method,
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload)
        });
      }
      writeOutput(data);
      setOpsMessage("Accion DIAN ejecutada.", false);
    } catch (error) {
      setOpsMessage(error.message || "No se pudo ejecutar accion DIAN", true);
      writeOutput({ error: error.message || "No se pudo ejecutar accion DIAN" });
    }
  }

  function refreshDianToolsVisibility() {
    var section = document.getElementById("dianToolsSection");
    if (!section) {
      return;
    }
    var moduleConfig = currentModule();
    if (moduleConfig && moduleConfig.dian) {
      section.classList.remove("erp-hidden");
      return;
    }
    section.classList.add("erp-hidden");
  }

  function initModuleSelector() {
    var selector = document.getElementById("moduloSelector");
    if (!selector) {
      return;
    }

    var domainModules = modulesForCurrentDomain();
    selector.innerHTML = "";

    for (var i = 0; i < domainModules.length; i += 1) {
      var moduleConfig = domainModules[i];
      var option = document.createElement("option");
      option.value = moduleConfig.key;
      option.textContent = moduleConfig.label;
      selector.appendChild(option);
    }

    if (domainModules.length === 0) {
      state.selectedModuleKey = "";
      setOpsMessage("Este dominio no tiene modulos configurados.", true);
      return;
    }

    var requestedModule = normalize(getQueryParam("modulo"));
    var selectedExists = false;
    for (var m = 0; m < domainModules.length; m += 1) {
      if (domainModules[m].key === requestedModule) {
        selectedExists = true;
        break;
      }
    }

    if (selectedExists) {
      state.selectedModuleKey = requestedModule;
    } else {
      state.selectedModuleKey = domainModules[0].key;
    }

    selector.value = state.selectedModuleKey;

    selector.onchange = function () {
      state.selectedModuleKey = normalize(selector.value);
      loadTemplateToPayload();
      renderGuidedFields();
      renderQuickActions();
      renderDomainGuide();
      refreshDianToolsVisibility();
      setOpsMessage("Modulo cambiado a " + (currentModule() ? currentModule().label : "N/A") + ".", false);
      listRows();
    };
  }

  function bindEvents() {
    var btnListar = document.getElementById("btnListar");
    if (btnListar) {
      btnListar.onclick = listRows;
    }

    var btnLimpiar = document.getElementById("btnLimpiar");
    if (btnLimpiar) {
      btnLimpiar.onclick = function () {
        writeOutput("Sin resultados todavia.");
        setOpsMessage("", false);
      };
    }

    var btnTemplate = document.getElementById("btnTemplate");
    if (btnTemplate) {
      btnTemplate.onclick = function () {
        loadTemplateToPayload();
        renderGuidedFields();
        copyGuidedToJSON("create");
      };
    }

    var btnResetGuided = document.getElementById("btnResetGuided");
    if (btnResetGuided) {
      btnResetGuided.onclick = resetGuidedForm;
    }

    var btnGuidedToJSON = document.getElementById("btnGuiadoToJSON");
    if (btnGuidedToJSON) {
      btnGuidedToJSON.onclick = function () {
        var guidedID = toInt((document.getElementById("guidedRecordID") || {}).value, 0);
        copyGuidedToJSON(guidedID > 0 ? "update" : "create");
        setOpsMessage("Payload avanzado actualizado desde formulario guiado.", false);
      };
    }

    var btnCrearGuiado = document.getElementById("btnCrearGuiado");
    if (btnCrearGuiado) {
      btnCrearGuiado.onclick = function () {
        executeGuidedCrud(false);
      };
    }

    var btnActualizarGuiado = document.getElementById("btnActualizarGuiado");
    if (btnActualizarGuiado) {
      btnActualizarGuiado.onclick = function () {
        executeGuidedCrud(true);
      };
    }

    var btnRunQuickAction = document.getElementById("btnRunQuickAction");
    if (btnRunQuickAction) {
      btnRunQuickAction.onclick = executeQuickAction;
    }

    var form = document.getElementById("jsonCrudForm");
    if (form) {
      form.addEventListener("submit", executeCrud);
    }

    var btnChecklist = document.getElementById("btnDianChecklist");
    if (btnChecklist) {
      btnChecklist.onclick = function () { runDianAction("checklist", "GET"); };
    }
    var btnValidar = document.getElementById("btnDianValidar");
    if (btnValidar) {
      btnValidar.onclick = function () { runDianAction("validar", "GET"); };
    }
    var btnCUFE = document.getElementById("btnDianCUFE");
    if (btnCUFE) {
      btnCUFE.onclick = function () { runDianAction("generar_cufe_demo", "POST"); };
    }
    var btnXML = document.getElementById("btnDianXML");
    if (btnXML) {
      btnXML.onclick = function () { runDianAction("generar_xml_demo", "POST"); };
    }
    var btnPruebas = document.getElementById("btnDianPruebas");
    if (btnPruebas) {
      btnPruebas.onclick = function () { runDianAction("pruebas_dian", "POST", { simular: false, detener_en_error: true }); };
    }
  }

  async function init() {
    try {
      ensureEmpresaID();
      ensureDomain();
    } catch (error) {
      setOpsMessage(error.message || "No se pudo inicializar", true);
      writeOutput({ error: error.message || "No se pudo inicializar" });
      return;
    }

    setDomainHeader();
    setDomainLinks();
    initModuleSelector();
    bindEvents();
    loadTemplateToPayload();
    renderGuidedFields();
    renderQuickActions();
    renderDomainGuide();
    copyGuidedToJSON("create");
    refreshDianToolsVisibility();
    await listRows();
  }

  init();
})();
