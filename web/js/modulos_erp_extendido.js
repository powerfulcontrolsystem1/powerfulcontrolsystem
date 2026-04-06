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
    selectedModuleKey: ""
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

  function parseDIANPayload() {
    var raw = normalize((document.getElementById("dianPayload") || {}).value);
    if (!raw) {
      return { empresa_id: state.empresaID };
    }
    var payload;
    try {
      payload = JSON.parse(raw);
    } catch (error) {
      throw new Error("Payload DIAN invalido");
    }
    if (!payload || typeof payload !== "object" || Array.isArray(payload)) {
      throw new Error("Payload DIAN debe ser objeto JSON");
    }
    payload.empresa_id = state.empresaID;
    return payload;
  }

  async function runDianAction(action, method) {
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
        var payload = parseDIANPayload();
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
      refreshDianToolsVisibility();
      setOpsMessage("Modulo cambiado a " + (currentModule() ? currentModule().label : "N/A") + ".", false);
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
      btnTemplate.onclick = loadTemplateToPayload;
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
    refreshDianToolsVisibility();
    await listRows();
  }

  init();
})();
