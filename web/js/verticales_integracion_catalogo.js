(function () {
  "use strict";

  var CORE_MODULES = ["clientes", "inventario", "ventas", "pagos", "facturacion", "reportes", "seguridad"];

  var catalog = {
    gimnasio: {
      estado: "plantilla_integrada_nucleo",
      visibleOperativo: true,
      motivo: "Sincroniza socios con clientes, planes con servicios vendibles y pagos con ventas/pagos centrales.",
      duplicados: []
    },
    odontologia: {
      estado: "plantilla_integrada_nucleo",
      visibleOperativo: true,
      motivo: "Sincroniza pacientes con clientes, tratamientos con servicios vendibles y pagos con ventas/pagos centrales.",
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
      motivo: "Cierra tickets creando venta, item de servicio y pago central en carritos.",
      duplicados: []
    },
    taxi_system: {
      estado: "plantilla_integrada_nucleo",
      visibleOperativo: true,
      motivo: "Sincroniza clientes de viaje y servicios completados con clientes, servicios, ventas y pagos centrales.",
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
      motivo: "Sincroniza pedidos entregados con clientes, servicios de menu, ventas y pagos centrales.",
      duplicados: []
    },
    apartamentos_turisticos: {
      estado: "plantilla_integrada_nucleo",
      visibleOperativo: true,
      motivo: "Sincroniza huespedes con clientes, apartamentos con servicios y reservas cerradas con ventas/pagos centrales.",
      duplicados: []
    },
    propiedad_horizontal: {
      estado: "plantilla_integrada_nucleo",
      visibleOperativo: true,
      motivo: "Sincroniza propietarios/residentes con clientes, cargos con servicios y recaudos con ventas/pagos centrales.",
      duplicados: []
    },
    alquileres: {
      estado: "plantilla_integrada_nucleo",
      visibleOperativo: true,
      motivo: "Sincroniza clientes de contratos con clientes centrales, activos/tarifas con servicios y contratos con ventas/pagos centrales.",
      duplicados: []
    },
    drogueria_farmacia: {
      estado: "plantilla_integrada_nucleo",
      visibleOperativo: true,
      motivo: "Opera como expediente sanitario sobre el nucleo: productos, inventario, ventas, clientes y facturacion siguen en modulos centrales.",
      duplicados: []
    },
    aiu_construccion: {
      estado: "plantilla_integrada_nucleo",
      visibleOperativo: true,
      motivo: "Sincroniza clientes de obra, contratos y conceptos como servicios; facturas AIU quedan enlazadas a ventas centrales sin duplicar impuestos.",
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
        syncAction: String(item.sync_action || "").trim(),
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
