(function () {
  "use strict";

  var defaultSections = ["Dashboard", "Configuración", "Registros", "Seguimiento", "Aprobaciones", "Evidencias"];

  var modules = [
    {
      id: "linkAgenciaViajes",
      module: "agencia_viajes",
      title: "Agencia de viajes",
      fullTitle: "Agencia de viajes y planes turísticos",
      lead: "Paquetes, reservas, itinerarios, vouchers, pagos por cuotas, proveedores y comisiones.",
      description: "Gestiona paquetes turísticos, cotizaciones, itinerarios, reservas, vouchers, pagos por cuotas, proveedores, comisiones y seguimiento comercial desde una sola operación. La agencia puede controlar clientes, fechas de viaje, saldos pendientes, confirmaciones y evidencias para vender planes completos con trazabilidad y mejor servicio.",
      summary: "Paquetes, reservas, itinerarios, vouchers y comisiones.",
      icon: "/img/hotel-logo.svg",
      secondaryIcon: "/img/portal-systems/realistic/agencia-viajes.jpg",
      sections: ["Dashboard comercial", "Paquetes y cotizaciones", "Reservas y vouchers", "Pagos y comisiones", "Aprobaciones", "Evidencias"]
    },
    {
      id: "linkOperadorTuristico",
      module: "operador_turistico",
      title: "Operador turístico",
      fullTitle: "Operador turístico local",
      lead: "Tours, guías, cupos, check-in, rutas, transporte y novedades.",
      description: "Organiza tours locales con rutas, guías, cupos, transporte, check-in, novedades y cierre operativo. El módulo ayuda a coordinar grupos, validar asistencia, registrar evidencias del recorrido, controlar disponibilidad por fecha y entregar reportes claros sobre ocupación, cumplimiento y rentabilidad de cada experiencia.",
      summary: "Tours locales, guías, cupos, check-in y rutas.",
      icon: "/img/gps.svg",
      secondaryIcon: "/img/portal-systems/realistic/operador-turistico.jpg",
      sections: ["Dashboard operativo", "Tours y rutas", "Guias y cupos", "Check-in", "Evidencias", "Cierre"]
    },
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
      id: "linkClinicaConsultorios",
      module: "clinica_consultorios",
      title: "Clínica y consultorios",
      fullTitle: "Clínica médica y consultorios múltiples",
      lead: "Pacientes, citas, profesionales, órdenes, historia clínica básica y remisiones.",
      description: "Gestiona pacientes, citas, profesionales, órdenes, remisiones, historia clínica básica y seguimiento administrativo para clínicas o consultorios múltiples. El flujo permite coordinar recepción, atención, autorizaciones, resultados, pagos y reportes manteniendo separación de permisos y trazabilidad por sede o especialidad.",
      summary: "Pacientes, citas, órdenes e historia clínica básica.",
      icon: "/img/report.svg",
      secondaryIcon: "/img/portal-systems/realistic/clinica-consultorios.jpg",
      sections: ["Pacientes", "Citas", "Historia", "Órdenes", "Aprobaciones", "Reportes"]
    },
    {
      id: "linkLaboratorioClinico",
      module: "laboratorio_clinico",
      title: "Laboratorio clínico",
      fullTitle: "Laboratorio clínico",
      lead: "Órdenes de examen, muestras, trazabilidad, resultados, validación y entrega digital.",
      description: "Administra órdenes de examen, toma de muestras, trazabilidad, control de calidad, resultados, validación profesional y entrega digital. El laboratorio puede seguir cada muestra por estado, priorizar tiempos de respuesta, documentar incidencias y cerrar servicios con reportes para usuarios, convenios y auditoría.",
      summary: "Órdenes, muestras, resultados y entrega digital.",
      icon: "/img/analytics-color.svg",
      secondaryIcon: "/img/portal-systems/realistic/laboratorio-clinico.jpg",
      sections: ["Órdenes", "Muestras", "Resultados", "Calidad", "Entrega", "SLA"]
    },
    {
      id: "linkColegioAcademia",
      module: "colegio_academia",
      title: "Colegio o academia",
      fullTitle: "Colegio, academia o instituto",
      lead: "Estudiantes, matrículas, cursos, mensualidades, asistencia, boletines y certificados.",
      description: "Organiza estudiantes, acudientes, matrículas, cursos, mensualidades, asistencia, boletines, certificados y cartera educativa. La institución puede controlar grupos, pagos, novedades académicas, documentos emitidos y comunicación administrativa, conectando la operación diaria con reportes para dirección y control financiero.",
      summary: "Matrículas, cursos, asistencia, mensualidades y boletines.",
      icon: "/img/company-briefcase-color.svg",
      secondaryIcon: "/img/portal-systems/realistic/colegio-academia.jpg",
      sections: ["Matriculas", "Cursos", "Asistencia", "Mensualidades", "Boletines", "Certificados"]
    },
    {
      id: "linkGuarderiaInfantil",
      module: "guarderia_infantil",
      title: "Guardería infantil",
      fullTitle: "Guardería y jardín infantil",
      lead: "Niños, acudientes, autorizaciones, alimentación, novedades, entregas y pagos.",
      description: "Controla niños, acudientes, autorizaciones, alimentación, novedades, entregas, pagos y seguimiento diario para guarderías o jardines infantiles. El módulo ayuda a registrar entrada y salida, responsables autorizados, observaciones de cuidado, cobros recurrentes y reportes para una operación segura y ordenada.",
      summary: "Niños, acudientes, autorizaciones, novedades y pagos.",
      icon: "/img/user-avatar-color.svg",
      secondaryIcon: "/img/portal-systems/realistic/guarderia-infantil.jpg",
      sections: ["Jornada", "Ninos", "Acudientes", "Autorizaciones", "Novedades", "Pagos"]
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
      id: "linkInmobiliariaComercial",
      module: "inmobiliaria_comercial",
      title: "Inmobiliaria",
      fullTitle: "Inmobiliaria comercial",
      lead: "Propiedades, arriendos, ventas, visitas, leads, propietarios, contratos y comisiones.",
      description: "Administra propiedades, arriendos, ventas, visitas, leads, propietarios, contratos, comisiones y seguimiento comercial inmobiliario. La agencia puede publicar inventario, programar recorridos, registrar interesados, controlar documentación, medir conversiones y cerrar negocios con trazabilidad financiera y contractual.",
      summary: "Propiedades, visitas, leads, propietarios y comisiones.",
      icon: "/img/company-briefcase-color.svg",
      secondaryIcon: "/img/portal-systems/realistic/inmobiliaria-comercial.jpg",
      sections: ["Propiedades", "Leads", "Visitas", "Contratos", "Comisiones", "Cierre"]
    },
    {
      id: "linkSeguridadPrivada",
      module: "seguridad_privada",
      title: "Seguridad privada",
      fullTitle: "Seguridad privada y vigilancia",
      lead: "Guardas, puestos, turnos, rondas QR, novedades, incidentes y escalamiento.",
      description: "Controla guardas, puestos, turnos, rondas QR, novedades, incidentes, escalamiento y reportes para empresas de seguridad privada. La operación puede verificar presencia, documentar eventos, asignar supervisores, medir cumplimiento de rondas y conservar evidencia para clientes, auditoría y control interno.",
      summary: "Guardas, puestos, turnos, rondas QR e incidentes.",
      icon: "/img/shield-license-color.svg",
      secondaryIcon: "/img/portal-systems/realistic/seguridad-privada.jpg",
      sections: ["Puestos", "Turnos", "Guardas", "Rondas QR", "Incidentes", "Reportes"]
    },
    {
      id: "linkClubDeportivo",
      module: "club_deportivo",
      title: "Club deportivo",
      fullTitle: "Club deportivo y escuela deportiva",
      lead: "Alumnos, entrenadores, disciplinas, clases, torneos, pagos y asistencia.",
      description: "Gestiona alumnos, entrenadores, disciplinas, clases, torneos, pagos, asistencia y comunicaciones para clubes o escuelas deportivas. El módulo ayuda a organizar grupos, programar entrenamientos, controlar mensualidades, registrar participación en eventos y medir avance operativo por disciplina o categoría.",
      summary: "Alumnos, entrenadores, disciplinas, torneos y asistencia.",
      icon: "/img/analytics-color.svg",
      secondaryIcon: "/img/portal-systems/realistic/club-deportivo.jpg",
      sections: ["Alumnos", "Disciplinas", "Clases", "Torneos", "Pagos", "Asistencia"]
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
    },
    {
      id: "linkCooperativaFondo",
      module: "cooperativa_fondo",
      title: "Cooperativa",
      fullTitle: "Cooperativa y fondo de empleados",
      lead: "Asociados, aportes, créditos internos, cartera, beneficios, convenios y asambleas.",
      description: "Gestiona asociados, aportes, créditos internos, cartera, beneficios, convenios, asambleas y trazabilidad administrativa para cooperativas o fondos de empleados. El módulo permite controlar solicitudes, aprobaciones, obligaciones, recaudos y comunicaciones manteniendo gobierno interno, reportes y permisos por responsabilidad.",
      summary: "Asociados, aportes, créditos internos, cartera y beneficios.",
      icon: "/img/money.svg",
      secondaryIcon: "/img/portal-systems/realistic/cooperativa-fondo.jpg",
      sections: ["Asociados", "Aportes", "Creditos", "Cartera", "Beneficios", "Asambleas"]
    },
    {
      id: "linkCapacitacionEmpresarial",
      module: "capacitacion_empresarial",
      title: "Capacitación empresarial",
      fullTitle: "Centro de capacitación empresarial",
      lead: "Cursos, instructores, cohortes, asistencia, evaluaciones, certificados y empresas cliente.",
      description: "Administra cursos, instructores, cohortes, empresas cliente, asistencia, evaluaciones, certificados y seguimiento comercial para centros de capacitación. La operación puede vender programas, controlar grupos, medir cumplimiento académico, emitir constancias y conservar historial por participante, empresa y periodo.",
      summary: "Cursos, instructores, certificados, asistencia y empresas.",
      icon: "/img/excel.svg",
      secondaryIcon: "/img/portal-systems/realistic/capacitacion-empresarial.jpg",
      sections: ["Cursos", "Cohortes", "Instructores", "Asistencia", "Evaluaciones", "Certificados"]
    }
  ];

  modules = modules.filter(function (item) {
    return String((item && item.module) || "").trim().toLowerCase() !== "colegio_academia";
  });

  var productionMassRank = {
    salon_spa: 1,
    veterinaria_petshop: 2,
    clinica_consultorios: 3,
    laboratorio_clinico: 4,
    taller_mecanico: 5,
    servicios_tecnicos: 6,
    lavanderia_tintoreria: 7,
    agencia_viajes: 8,
    eventos_boleteria: 9,
    transporte_carga_tms: 10,
    operador_turistico: 11,
    guarderia_infantil: 12,
    inmobiliaria_comercial: 13,
    seguridad_privada: 14,
    club_deportivo: 15,
    funeraria_exequial: 16,
    parque_recreativo: 17,
    cooperativa_fondo: 18,
    capacitacion_empresarial: 19
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
