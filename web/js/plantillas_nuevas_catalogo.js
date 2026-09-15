(function () {
  "use strict";

  // Compatibilidad interna: este catálogo ya no representa una colección de
  // plantillas. Solo conserva el sistema adicional autorizado para talleres.
  var modules = [{
    id: "linkTallerMecanico",
    module: "taller_mecanico",
    title: "Taller de motos",
    fullTitle: "Taller de motos",
    lead: "Órdenes de trabajo, diagnóstico, repuestos, mano de obra, garantía y entrega.",
    description: "Administra órdenes de trabajo para motos, diagnósticos, repuestos, mano de obra, aprobaciones, garantías y entrega. El taller registra entradas, cotiza reparaciones, controla inventario, adjunta evidencias e informa avances al cliente.",
    summary: "Órdenes, diagnóstico, repuestos, mano de obra y garantías.",
    icon: "/img/settings-color.svg",
    secondaryIcon: "/img/portal-systems/realistic/taller-mecanico.jpg",
    sections: ["Órdenes", "Diagnóstico", "Repuestos", "Mano de obra", "Garantías", "Entrega"],
    productionMass: true,
    productionRank: 1,
    decisionPreconfig: "sistema_destacado",
    decisionLabel: "Sistema destacado",
    decisionReason: "Sistema operativo conservado fuera del catálogo de siete preconfiguraciones básicas.",
    integrationStatus: "sistema_integrado_nucleo",
    operationalVisible: true,
    coreModules: ["clientes", "inventario", "ventas", "pagos", "facturacion", "reportes", "seguridad"],
    templateActivates: ["taller_mecanico", "clientes", "inventario", "ventas", "pagos", "facturacion", "reportes", "seguridad", "permisos", "licencias"],
    tablesTouched: ["empresa_modulos_colombia_registros", "clientes", "servicios", "carritos_compras", "carrito_compra_items", "empresa_finanzas_movimientos"],
    requiredPermissions: ["ver", "crear", "editar", "reportar", "cobrar"],
    saleFlow: "Orden del taller, aprobación, repuestos y mano de obra, venta o pago central y cierre con evidencia.",
    reportsProduced: ["Órdenes del taller", "Ventas por empresa", "Caja y pagos", "Inventario de repuestos"],
    portalStatus: "Sistema destacado",
    portalDescription: "Sistema de Taller de motos integrado al núcleo empresarial; no es una octava preconfiguración básica."
  }];

  window.PCS_NUEVAS_PLANTILLAS = modules;
  window.PCS_NUEVAS_PLANTILLAS_PRODUCCION_MASIVA = modules.slice();
  window.PCS_NUEVAS_PLANTILLAS_DIFERIDAS = [];
  window.PCS_NUEVAS_PLANTILLAS_MODULES = [["linkTallerMecanico", "taller_mecanico"]];
  window.PCS_NUEVAS_PLANTILLAS_KEYS = ["taller_mecanico"];
  window.PCS_NUEVAS_PLANTILLAS_CSV = "taller_mecanico";
})();
