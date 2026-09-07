(function () {
  "use strict";

  var defaultSections = ["Dashboard", "Configuración", "Registros", "Seguimiento", "Aprobaciones", "Evidencias"];

  var modules = [
    {
      id: "linkEventosBoleteria",
      module: "eventos_boleteria",
      title: "Eventos y boletería",
      fullTitle: "Eventos y boletería",
      lead: "Eventos, boletas QR, preventa, aforo, validación en puerta y patrocinadores.",
      description: "Administra eventos, preventas, boletas QR, aforo, validación en puerta, patrocinadores y reportes de ingreso. La operación queda preparada para vender entradas, controlar accesos, monitorear ocupación, registrar cortes por jornada y mantener trazabilidad de asistentes, canales comerciales y autorizaciones.",
      summary: "Eventos, boletas QR, aforo, preventa y puerta.",
      icon: "/img/tags-color.svg",
      secondaryIcon: "/img/portal-systems/realistic/eventos-boleteria.jpg",
      sections: ["Dashboard", "Eventos", "Boletas QR", "Aforo", "Validación", "Reportes"]
    },
    {
      id: "linkSalonSpa",
      module: "salon_spa",
      title: "Salón, barbería y spa",
      fullTitle: "Salón de belleza, barbería y spa",
      lead: "Agenda por profesional, cabinas, paquetes, servicios, insumos y comisiones.",
      description: "Controla agenda por profesional, cabinas, servicios, paquetes, insumos, comisiones y cierre de caja para salón, barbería o spa. El negocio puede reservar citas, asignar recursos, medir productividad, descontar productos usados y dar seguimiento a clientes frecuentes con historial comercial y operativo.",
      summary: "Agenda por profesional, cabinas, paquetes e insumos.",
      icon: "/img/customer.svg",
      secondaryIcon: "/img/portal-systems/realistic/salon-spa.jpg",
      sections: ["Agenda", "Servicios", "Profesionales", "Insumos", "Comisiones", "Cierre"]
    },
    {
      id: "linkVeterinariaPetshop",
      module: "veterinaria_petshop",
      title: "Veterinaria y pet shop",
      fullTitle: "Veterinaria y pet shop",
      lead: "Mascotas, vacunas, historia veterinaria, peluquería, productos y hospitalización.",
      description: "Centraliza mascotas, propietarios, vacunas, historia veterinaria, peluquería, productos, hospitalización y recordatorios. La veterinaria puede atender consultas, registrar tratamientos, vender artículos de pet shop, programar seguimientos y conservar evidencia clínica y comercial de cada paciente con permisos por rol.",
      summary: "Mascotas, vacunas, historia, peluquería y productos.",
      icon: "/img/shield-license-color.svg",
      secondaryIcon: "/img/portal-systems/realistic/veterinaria-petshop.jpg",
      sections: ["Pacientes", "Historia", "Vacunas", "Peluqueria", "Productos", "Seguimiento"]
    },
    {
      id: "linkLavanderiaTintoreria",
      module: "lavanderia_tintoreria",
      title: "Lavandería",
      fullTitle: "Lavandería y tintorería",
      lead: "Órdenes por prenda, etiquetas, lavado, planchado, calidad, entregas y reclamos.",
      description: "Gestiona recepción de prendas, etiquetas, estados de lavado, planchado, control de calidad, entregas, pagos y reclamos. La lavandería puede saber dónde está cada orden, evitar pérdidas, documentar novedades, programar rutas o domicilios y cerrar servicios con evidencia para el cliente.",
      summary: "Órdenes por prenda, etiquetas, estados y entregas.",
      icon: "/img/report.svg",
      secondaryIcon: "/img/portal-systems/realistic/lavanderia-tintoreria.jpg",
      sections: ["Recepcion", "Prendas", "Etiquetas", "Proceso", "Entrega", "Reclamos"]
    },
    {
      id: "linkTallerMecanico",
      module: "taller_mecanico",
      title: "Taller mecánico",
      fullTitle: "Taller mecánico, motos y autos",
      lead: "Órdenes de trabajo, diagnóstico, repuestos, mano de obra, garantía y entrega.",
      description: "Administra órdenes de trabajo, diagnósticos, repuestos, mano de obra, aprobaciones, garantías y entrega de vehículos. El taller puede registrar entrada, cotizar reparaciones, controlar inventario usado, tomar evidencias, informar avances al cliente y medir tiempos, costos y rentabilidad por servicio.",
      summary: "Órdenes, diagnóstico, repuestos, mano de obra y garantías.",
      icon: "/img/settings-color.svg",
      secondaryIcon: "/img/portal-systems/realistic/taller-mecanico.jpg",
      sections: ["Órdenes", "Diagnóstico", "Repuestos", "Mano de obra", "Garantías", "Entrega"]
    },
    {
      id: "linkTransporteCargaTMS",
      module: "transporte_carga_tms",
      title: "Transporte TMS",
      fullTitle: "Transporte de carga / TMS",
      lead: "Fletes, manifiestos, conductores, vehículos, rutas, tracking, entregas y liquidación.",
      description: "Controla fletes, manifiestos, conductores, vehículos, rutas, tracking, entregas, novedades y liquidación para transporte de carga. La empresa puede coordinar operaciones, registrar evidencia de entrega, monitorear cumplimiento, asignar costos y mantener trazabilidad documental desde la solicitud hasta el cierre.",
      summary: "Fletes, manifiestos, conductores, rutas y entregas.",
      icon: "/img/vehiculos-flotas-logo.svg",
      secondaryIcon: "/img/portal-systems/realistic/transporte-carga-tms.jpg",
      sections: ["Fletes", "Manifiestos", "Conductores", "Tracking", "Entregas", "Liquidacion"]
    },
    {
      id: "linkServiciosTecnicos",
      module: "servicios_tecnicos",
      title: "Servicios técnicos",
      fullTitle: "Servicios técnicos a domicilio",
      lead: "Órdenes de servicio, técnicos, visitas, repuestos, firmas, evidencias y garantías.",
      description: "Gestiona órdenes de servicio, agenda de técnicos, visitas a domicilio, repuestos, firmas, evidencias, garantías y seguimiento posventa. El módulo permite asignar responsables, controlar tiempos de atención, documentar trabajos realizados, cobrar servicios y mantener historial por cliente, equipo o ubicación.",
      summary: "Órdenes a domicilio, técnicos, repuestos, firmas y evidencias.",
      icon: "/img/network-color.svg",
      secondaryIcon: "/img/portal-systems/realistic/servicios-tecnicos.jpg",
      sections: ["Órdenes", "Agenda", "Técnicos", "Repuestos", "Firmas", "Garantías"]
    },
    {
      id: "linkFunerariaExequial",
      module: "funeraria_exequial",
      title: "Funeraria exequial",
      fullTitle: "Funeraria y servicios exequiales",
      lead: "Planes, afiliados, salas, servicios, traslados, contratos y documentación.",
      description: "Administra planes exequiales, afiliados, contratos, salas, servicios, traslados, documentación y cierre operativo para funerarias. La empresa puede coordinar atenciones sensibles, controlar obligaciones, registrar pagos, organizar recursos y mantener expedientes claros con permisos, evidencias y reportes.",
      summary: "Planes, afiliados, salas, servicios y documentos.",
      icon: "/img/report.svg",
      secondaryIcon: "/img/portal-systems/realistic/funeraria-exequial.jpg",
      sections: ["Planes", "Afiliados", "Servicios", "Salas", "Documentos", "Cierre"]
    },
    {
      id: "linkParqueRecreativo",
      module: "parque_recreativo",
      title: "Parque recreativo",
      fullTitle: "Parque recreativo y atracciones",
      lead: "Entradas, manillas QR, aforo, atracciones, consumo interno, incidentes y cierre.",
      description: "Controla entradas, manillas QR, aforo, atracciones, consumo interno, incidentes y cierre de jornada para parques recreativos. La operación puede vender accesos, validar ingreso, monitorear ocupación, registrar novedades, administrar puntos de consumo y evaluar ventas, seguridad y experiencia del visitante.",
      summary: "Entradas, manillas QR, aforo, atracciones y consumo.",
      icon: "/img/tags-color.svg",
      secondaryIcon: "/img/portal-systems/realistic/parque-recreativo.jpg",
      sections: ["Entradas", "Manillas QR", "Aforo", "Atracciones", "Consumo", "Incidentes"]
    }
  ];

  var productionMassRank = {
    salon_spa: 1,
    veterinaria_petshop: 2,
    taller_mecanico: 5,
    servicios_tecnicos: 6,
    lavanderia_tintoreria: 7,
    eventos_boleteria: 9,
    transporte_carga_tms: 10,
    funeraria_exequial: 16,
    parque_recreativo: 17,
  };

  function uniqueList(values) {
    var seen = {};
    var out = [];
    (values || []).forEach(function (value) {
      var clean = String(value || "").trim();
      if (!clean || seen[clean]) return;
      seen[clean] = true;
      out.push(clean);
    });
    return out;
  }

  modules.forEach(function (item) {
    if (!Array.isArray(item.sections) || item.sections.length === 0) {
      item.sections = defaultSections.slice();
    }
    var module = String(item.module || "").trim();
    var rank = productionMassRank[module] || 0;
    var isProductionMass = rank > 0;
    item.productionMass = isProductionMass;
    item.productionRank = rank;
    item.decisionPreconfig = isProductionMass ? "integrar_v1_produccion_masiva" : "no_productivo";
    item.decisionLabel = isProductionMass ? "Produccion" : "No productivo";
    item.decisionReason = isProductionMass
      ? "Plantilla real de produccion masiva sobre el nucleo unico, sin duplicar clientes, productos, ventas ni pagos."
      : "Plantilla conservada en catalogo, pero no priorizada para la primera venta masiva.";
    item.integrationStatus = item.integrationStatus || "plantilla_integrada_nucleo";
    item.operationalVisible = isProductionMass;
    item.coreModules = uniqueList(item.coreModules || ["clientes", "inventario", "ventas", "pagos", "facturacion", "reportes", "seguridad"]);
    item.templateActivates = uniqueList((item.templateActivates || []).concat(item.coreModules, [module, "permisos", "licencias"]));
    item.tablesTouched = uniqueList((item.tablesTouched || []).concat([
      "tipo_empresa",
      "licencias",
      "paginas",
      "roles",
      "permisos",
      "productos/servicios",
      "clientes",
      "ventas",
      "pagos",
      "reportes"
    ]));
    item.requiredPermissions = uniqueList(item.requiredPermissions || ["ver", "crear", "editar", "reportar", "cobrar"]);
    item.saleFlow = item.saleFlow || "Cotizacion o venta directa usando clientes, productos/servicios, carritos, pagos y facturacion centrales.";
    item.reportsProduced = uniqueList(item.reportsProduced || ["Ventas por empresa", "Caja y pagos", "Clientes", "Servicios/productos", "Reporte operativo de la plantilla"]);
    item.portalStatus = isProductionMass ? ("Produccion #" + rank) : "No productivo";
    item.portalDescription = (isProductionMass
      ? ("Plantilla real de produccion masiva. Activa " + item.templateActivates.slice(0, 6).join(", ") + " sobre el nucleo unico. ")
      : ("Plantilla tecnica no publicada como operacion. No duplica clientes, productos, ventas ni pagos. "))
      + item.description;
  });

  window.PCS_NUEVAS_PLANTILLAS = modules;
  window.PCS_NUEVAS_PLANTILLAS_PRODUCCION_MASIVA = modules
    .filter(function (item) { return item.productionMass; })
    .sort(function (a, b) { return a.productionRank - b.productionRank; });
  window.PCS_NUEVAS_PLANTILLAS_DIFERIDAS = modules.filter(function (item) { return !item.productionMass; });
  window.PCS_NUEVAS_PLANTILLAS_MODULES = modules.map(function (item) {
    return [item.id, item.module];
  });
  window.PCS_NUEVAS_PLANTILLAS_KEYS = modules.map(function (item) {
    return item.module;
  });
  window.PCS_NUEVAS_PLANTILLAS_CSV = window.PCS_NUEVAS_PLANTILLAS_KEYS.join(",");
})();
