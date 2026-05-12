(function () {
  "use strict";

  var CORE_MODULES = ["clientes", "inventario", "ventas", "pagos", "facturacion", "reportes", "seguridad"];

  var catalog = {
    gimnasio: {
      estado: "plantilla_integrada_nucleo",
      visibleOperativo: true,
      motivo: "Plantilla fitness conectada al nucleo comun: socios, planes y pagos operan desde clientes, servicios, ventas y pagos centrales.",
      duplicados: []
    },
    odontologia: {
      estado: "plantilla_integrada_nucleo",
      visibleOperativo: true,
      motivo: "Plantilla clinica conectada al nucleo comun: pacientes, tratamientos y recaudos usan clientes, servicios, ventas y pagos centrales.",
      duplicados: []
    },
    consultorio_odontologico: {
      estado: "plantilla_integrada_nucleo",
      visibleOperativo: true,
      aliasDe: "odontologia",
      motivo: "Vista especializada de odontologia integrada al nucleo operativo.",
      duplicados: []
    },
    parqueadero: {
      estado: "plantilla_integrada_nucleo",
      visibleOperativo: true,
      motivo: "Plantilla de parqueadero conectada al nucleo comun: tickets y cobros crean servicio, venta y pago central sin modulo comercial paralelo.",
      duplicados: []
    },
    taxi_system: {
      estado: "plantilla_integrada_nucleo",
      visibleOperativo: true,
      motivo: "Plantilla de transporte conectada al nucleo comun: clientes, servicios de viaje, ventas y pagos se gobiernan desde el nucleo.",
      duplicados: []
    },
    taxi: {
      estado: "plantilla_integrada_nucleo",
      visibleOperativo: true,
      aliasDe: "taxi_system",
      motivo: "Alias visual de taxi_system integrado al nucleo.",
      duplicados: []
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
    },
    turnos_atencion: {
      estado: "integrado_soporte",
      visibleOperativo: true,
      motivo: "Funciona como capacidad operativa transversal y no reemplaza clientes, productos, ventas ni pagos.",
      duplicados: []
    },
    turnos: {
      estado: "integrado_soporte",
      visibleOperativo: true,
      aliasDe: "turnos_atencion",
      motivo: "Alias visual de turnos_atencion.",
      duplicados: []
    }
  };

  function normalizeModule(module) {
    return String(module || "").trim().toLowerCase();
  }

  function get(module) {
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
    return {
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
        title: String(item.title || item.titulo || "").trim()
      }
    };
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

  window.PCS_VERTICAL_CORE_MODULES = CORE_MODULES.slice();
  window.PCS_VERTICAL_INTEGRATION = {
    catalog: catalog,
    get: get,
    isOperationalVisible: isOperationalVisible,
    applyCatalogItems: applyCatalogItems,
    summary: summary
  };
})();
