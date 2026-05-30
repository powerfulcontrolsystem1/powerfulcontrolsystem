(function () {
  "use strict";

  var CORE_MODULES = ["clientes", "inventario", "ventas", "pagos", "finanzas", "facturacion", "reportes", "seguridad"];
  var FINANCIAL_CORE_MODULES = ["ventas", "pagos", "finanzas", "bancos_pagos", "tesoreria_presupuesto", "reportes"];
  var DEFAULT_INCOME_FLOW = ["servicio/producto vendible de la plantilla", "carrito o venta central", "pago central", "movimiento ingreso en empresa_finanzas_movimientos", "reporte financiero consolidado"];
  var DEFAULT_EXPENSE_FLOW = ["compra/gasto operativo de la plantilla", "soporte o documento central", "movimiento egreso en empresa_finanzas_movimientos", "conciliacion bancaria/tesoreria", "reporte financiero consolidado"];
  var DEFAULT_FINANCIAL_TABLES = ["carritos_compras", "carrito_compra_items", "empresa_finanzas_movimientos", "empresa_finanzas_configuracion", "empresa_finanzas_periodos"];
  var DEFAULT_FINANCIAL_REPORTS = ["ingresos por plantilla", "egresos por plantilla", "margen operativo", "flujo de caja", "estado de resultados por empresa"];

  function applyFinancialDefaults(meta) {
    meta.coreModules = Array.isArray(meta.coreModules) && meta.coreModules.length ? meta.coreModules : CORE_MODULES.slice();
    meta.financialCoreModules = Array.isArray(meta.financialCoreModules) && meta.financialCoreModules.length ? meta.financialCoreModules : FINANCIAL_CORE_MODULES.slice();
    meta.incomeFlow = Array.isArray(meta.incomeFlow) && meta.incomeFlow.length ? meta.incomeFlow : DEFAULT_INCOME_FLOW.slice();
    meta.expenseFlow = Array.isArray(meta.expenseFlow) && meta.expenseFlow.length ? meta.expenseFlow : DEFAULT_EXPENSE_FLOW.slice();
    meta.financialTables = Array.isArray(meta.financialTables) && meta.financialTables.length ? meta.financialTables : DEFAULT_FINANCIAL_TABLES.slice();
    meta.financialReports = Array.isArray(meta.financialReports) && meta.financialReports.length ? meta.financialReports : DEFAULT_FINANCIAL_REPORTS.slice();
    return meta;
  }

  var catalog = {
    gimnasio: {
      estado: "plantilla_integrada_nucleo",
      visibleOperativo: true,
      motivo: "Plantilla fitness conectada al nucleo comun: socios, planes y pagos operan desde clientes, servicios, ventas y pagos centrales.",
      duplicados: [],
      supportModules: ["estaciones", "turnos_atencion"],
      similarTemplates: ["club_deportivo"]
    },
    odontologia: {
      estado: "plantilla_integrada_nucleo",
      visibleOperativo: true,
      motivo: "Plantilla clinica conectada al nucleo comun: pacientes, tratamientos y recaudos usan clientes, servicios, ventas y pagos centrales.",
      duplicados: [],
      fusedModules: ["consultorio_odontologico"],
      supportModules: ["turnos_atencion", "estaciones"],
      similarTemplates: ["clinica_consultorios"]
    },
    parqueadero: {
      estado: "plantilla_integrada_nucleo",
      visibleOperativo: true,
      motivo: "Plantilla de parqueadero conectada al nucleo comun: tickets y cobros crean servicio, venta y pago central sin modulo comercial paralelo.",
      duplicados: [],
      supportModules: ["estaciones", "turnos_atencion"],
      similarTemplates: ["parque_recreativo"]
    },
    taxi_system: {
      estado: "plantilla_integrada_nucleo",
      visibleOperativo: true,
      motivo: "Plantilla de transporte conectada al nucleo comun: clientes, servicios de viaje, ventas y pagos se gobiernan desde el nucleo.",
      duplicados: [],
      fusedModules: ["taxi"],
      supportModules: ["estaciones"],
      similarTemplates: ["transporte_carga_tms"]
    },
    domicilios: {
      estado: "plantilla_integrada_nucleo",
      visibleOperativo: true,
      motivo: "Plantilla logistica conectada al nucleo comun: pedidos, clientes, menu, ventas y pagos se resuelven en los modulos centrales.",
      duplicados: []
    },
    apartamentos_turisticos: {
      estado: "plantilla_integrada_nucleo",
      visibleOperativo: true,
      motivo: "Plantilla de alojamiento conectada al nucleo comun: huespedes, unidades vendibles, reservas, ventas y pagos comparten el motor central.",
      duplicados: []
    },
    propiedad_horizontal: {
      estado: "plantilla_integrada_nucleo",
      visibleOperativo: true,
      motivo: "Plantilla de copropiedad conectada al nucleo comun: terceros, unidades, cargos, recaudos, cartera y reportes no duplican clientes ni pagos.",
      duplicados: []
    },
    alquileres: {
      estado: "plantilla_integrada_nucleo",
      visibleOperativo: true,
      motivo: "Plantilla de alquiler conectada al nucleo comun: clientes, activos vendibles, contratos, ventas y pagos usan la fuente unica.",
      duplicados: []
    },
    drogueria_farmacia: {
      estado: "plantilla_integrada_nucleo",
      visibleOperativo: true,
      motivo: "Plantilla sanitaria conectada al nucleo comun: productos, inventario, compras, clientes, ventas y facturacion siguen en modulos centrales.",
      duplicados: []
    },
    aiu_construccion: {
      estado: "plantilla_integrada_nucleo",
      visibleOperativo: true,
      motivo: "Plantilla de construccion conectada al nucleo comun: clientes, contratos, conceptos, ventas, impuestos y reportes se enlazan sin duplicar documentos comerciales.",
      duplicados: []
    }
  };

  Object.keys(catalog).forEach(function (key) {
    catalog[key] = applyFinancialDefaults(catalog[key] || {});
  });

  function normalizeModule(module) {
    return String(module || "").trim().toLowerCase();
  }

  function get(module) {
    ensureNewVerticalFallback();
    var key = normalizeModule(module);
    return catalog[key] || null;
  }

  function isOperationalVisible(module) {
    var meta = get(module);
    return !meta || meta.visibleOperativo !== false;
  }

  function normalizeServerItem(item) {
    if (!item || typeof item !== "object") return null;
    var module = normalizeModule(item.module || item.modulo || item.id);
    if (!module) return null;
    var normalized = {
      module: module,
      meta: {
        estado: String(item.integration_status || item.estado || "").trim() || "pendiente_integracion_nucleo",
        visibleOperativo: item.operational_visible !== false && item.visible_operativo !== false,
        motivo: String(item.reason || item.motivo || "").trim(),
        duplicados: Array.isArray(item.duplicates_core) ? item.duplicates_core.slice() : (Array.isArray(item.duplicados) ? item.duplicados.slice() : []),
        aliasDe: String(item.alias_of || item.aliasDe || "").trim(),
        coreModules: Array.isArray(item.core_modules) ? item.core_modules.slice() : CORE_MODULES.slice(),
        templateActivates: Array.isArray(item.template_activates) ? item.template_activates.slice() : [],
        tablesTouched: Array.isArray(item.tables_touched) ? item.tables_touched.slice() : [],
        requiredPermissions: Array.isArray(item.required_permissions) ? item.required_permissions.slice() : [],
        saleFlow: Array.isArray(item.sale_flow) ? item.sale_flow.slice() : [],
        reportsProduced: Array.isArray(item.reports_produced) ? item.reports_produced.slice() : [],
        flujoPropioPermitido: Array.isArray(item.own_flow_allowed) ? item.own_flow_allowed.slice() : [],
        decision: String(item.decision || "").trim(),
        page: String(item.page || "").trim(),
        title: String(item.title || item.titulo || "").trim(),
        professionalReady: item.professional_ready === true,
        readinessScore: Number(item.readiness_score || 0),
        readinessChecks: Array.isArray(item.readiness_checks) ? item.readiness_checks.slice() : [],
        configurationScope: Array.isArray(item.configuration_scope) ? item.configuration_scope.slice() : [],
        fusedModules: Array.isArray(item.fused_modules) ? item.fused_modules.slice() : [],
        supportModules: Array.isArray(item.support_modules) ? item.support_modules.slice() : [],
        similarTemplates: Array.isArray(item.similar_templates) ? item.similar_templates.slice() : [],
        financialCoreModules: Array.isArray(item.financial_core_modules) ? item.financial_core_modules.slice() : [],
        incomeFlow: Array.isArray(item.income_flow) ? item.income_flow.slice() : [],
        expenseFlow: Array.isArray(item.expense_flow) ? item.expense_flow.slice() : [],
        financialTables: Array.isArray(item.financial_tables) ? item.financial_tables.slice() : [],
        financialReports: Array.isArray(item.financial_reports) ? item.financial_reports.slice() : []
      }
    };
    normalized.meta = applyFinancialDefaults(normalized.meta);
    return normalized;
  }

  function applyCatalogItems(items) {
    if (!Array.isArray(items)) return 0;
    var count = 0;
    items.forEach(function (item) {
      var normalized = normalizeServerItem(item);
      if (!normalized) return;
      catalog[normalized.module] = normalized.meta;
      count += 1;
    });
    return count;
  }

  function summary() {
    ensureNewVerticalFallback();
    var keys = Object.keys(catalog);
    var visible = 0;
    var hidden = 0;
    var pending = 0;
    var duplicates = 0;
    keys.forEach(function (key) {
      var meta = catalog[key] || {};
      if (meta.visibleOperativo === false) hidden += 1;
      else visible += 1;
      if (meta.estado === "pendiente_integracion_nucleo") pending += 1;
      if (Array.isArray(meta.duplicados) && meta.duplicados.length) duplicates += 1;
    });
    return {
      total: keys.length,
      visible: visible,
      hidden: hidden,
      pending: pending,
      duplicates: duplicates
    };
  }

  function ensureNewVerticalFallback() {
    if (ensureNewVerticalFallback.done || !Array.isArray(window.PCS_NUEVAS_PLANTILLAS)) return;
    window.PCS_NUEVAS_PLANTILLAS.forEach(function (item) {
      if (!item || !item.module) return;
      var module = normalizeModule(item.module);
      if (!module || catalog[module]) return;
      catalog[module] = applyFinancialDefaults({
        estado: item.integrationStatus || "plantilla_integrada_nucleo",
        visibleOperativo: item.operationalVisible !== false,
        motivo: item.decisionReason || item.summary || "",
        duplicados: [],
        coreModules: Array.isArray(item.coreModules) ? item.coreModules.slice() : CORE_MODULES.slice(),
        templateActivates: Array.isArray(item.templateActivates) ? item.templateActivates.slice() : [],
        tablesTouched: Array.isArray(item.tablesTouched) ? item.tablesTouched.slice() : [],
        requiredPermissions: Array.isArray(item.requiredPermissions) ? item.requiredPermissions.slice() : [],
        saleFlow: Array.isArray(item.saleFlow) ? item.saleFlow.slice() : [String(item.saleFlow || "venta central")],
        reportsProduced: Array.isArray(item.reportsProduced) ? item.reportsProduced.slice() : [],
        flujoPropioPermitido: Array.isArray(item.sections) ? item.sections.slice() : [],
        decision: item.decisionPreconfig || "integrar_v1_produccion_masiva",
        page: item.id || "",
        title: item.fullTitle || item.title || module,
        professionalReady: true,
        readinessScore: 100,
        configurationScope: ["tipo_empresa_preconfiguracion", "licencia", "roles", "menu", "datos_guia", "reportes"],
        fusedModules: [],
        supportModules: [],
        similarTemplates: []
      });
    });
    ensureNewVerticalFallback.done = true;
  }

  window.PCS_VERTICAL_CORE_MODULES = CORE_MODULES.slice();
  window.PCS_VERTICAL_INTEGRATION = {
    catalog: catalog,
    get: get,
    isOperationalVisible: isOperationalVisible,
    applyCatalogItems: applyCatalogItems,
    summary: summary
  };
})();
