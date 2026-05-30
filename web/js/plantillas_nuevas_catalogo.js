(function () {
  "use strict";

  var defaultSections = ["Dashboard", "Configuracion", "Registros", "Seguimiento", "Aprobaciones", "Evidencias"];

  var modules = [
    {
      id: "linkAgenciaViajes",
      module: "agencia_viajes",
      title: "Agencia de viajes",
      fullTitle: "Agencia de viajes y planes turisticos",
      lead: "Paquetes, reservas, itinerarios, vouchers, pagos por cuotas, proveedores y comisiones.",
      description: "Gestiona paquetes turisticos, cotizaciones, itinerarios, reservas, vouchers, pagos por cuotas, proveedores, comisiones y seguimiento comercial desde una sola operacion. La agencia puede controlar clientes, fechas de viaje, saldos pendientes, confirmaciones y evidencias para vender planes completos con trazabilidad y mejor servicio.",
      summary: "Paquetes, reservas, itinerarios, vouchers y comisiones.",
      icon: "/img/hotel-logo.svg",
      secondaryIcon: "/img/report.svg",
      sections: ["Dashboard comercial", "Paquetes y cotizaciones", "Reservas y vouchers", "Pagos y comisiones", "Aprobaciones", "Evidencias"]
    },
    {
      id: "linkOperadorTuristico",
      module: "operador_turistico",
      title: "Operador turistico",
      fullTitle: "Operador turistico local",
      lead: "Tours, guias, cupos, check-in, rutas, transporte y novedades.",
      description: "Organiza tours locales con rutas, guias, cupos, transporte, check-in, novedades y cierre operativo. El modulo ayuda a coordinar grupos, validar asistencia, registrar evidencias del recorrido, controlar disponibilidad por fecha y entregar reportes claros sobre ocupacion, cumplimiento y rentabilidad de cada experiencia.",
      summary: "Tours locales, guias, cupos, check-in y rutas.",
      icon: "/img/gps.svg",
      secondaryIcon: "/img/analytics-color.svg",
      sections: ["Dashboard operativo", "Tours y rutas", "Guias y cupos", "Check-in", "Evidencias", "Cierre"]
    },
    {
      id: "linkEventosBoleteria",
      module: "eventos_boleteria",
      title: "Eventos y boleteria",
      fullTitle: "Eventos y boleteria",
      lead: "Eventos, boletas QR, preventa, aforo, validacion en puerta y patrocinadores.",
      description: "Administra eventos, preventas, boletas QR, aforo, validacion en puerta, patrocinadores y reportes de ingreso. La operacion queda preparada para vender entradas, controlar accesos, monitorear ocupacion, registrar cortes por jornada y mantener trazabilidad de asistentes, canales comerciales y autorizaciones.",
      summary: "Eventos, boletas QR, aforo, preventa y puerta.",
      icon: "/img/tags-color.svg",
      secondaryIcon: "/img/shield-license-color.svg",
      sections: ["Dashboard", "Eventos", "Boletas QR", "Aforo", "Validacion", "Reportes"]
    },
    {
      id: "linkSalonSpa",
      module: "salon_spa",
      title: "Salon, barberia y spa",
      fullTitle: "Salon de belleza, barberia y spa",
      lead: "Agenda por profesional, cabinas, paquetes, servicios, insumos y comisiones.",
      description: "Controla agenda por profesional, cabinas, servicios, paquetes, insumos, comisiones y cierre de caja para salon, barberia o spa. El negocio puede reservar citas, asignar recursos, medir productividad, descontar productos usados y dar seguimiento a clientes frecuentes con historial comercial y operativo.",
      summary: "Agenda por profesional, cabinas, paquetes e insumos.",
      icon: "/img/customer.svg",
      secondaryIcon: "/img/settings-color.svg",
      sections: ["Agenda", "Servicios", "Profesionales", "Insumos", "Comisiones", "Cierre"]
    },
    {
      id: "linkVeterinariaPetshop",
      module: "veterinaria_petshop",
      title: "Veterinaria y pet shop",
      fullTitle: "Veterinaria y pet shop",
      lead: "Mascotas, vacunas, historia veterinaria, peluqueria, productos y hospitalizacion.",
      description: "Centraliza mascotas, propietarios, vacunas, historia veterinaria, peluqueria, productos, hospitalizacion y recordatorios. La veterinaria puede atender consultas, registrar tratamientos, vender articulos de pet shop, programar seguimientos y conservar evidencia clinica y comercial de cada paciente con permisos por rol.",
      summary: "Mascotas, vacunas, historia, peluqueria y productos.",
      icon: "/img/shield-license-color.svg",
      secondaryIcon: "/img/customer.svg",
      sections: ["Pacientes", "Historia", "Vacunas", "Peluqueria", "Productos", "Seguimiento"]
    },
    {
      id: "linkClinicaConsultorios",
      module: "clinica_consultorios",
      title: "Clinica y consultorios",
      fullTitle: "Clinica medica y consultorios multiples",
      lead: "Pacientes, citas, profesionales, ordenes, historia clinica basica y remisiones.",
      description: "Gestiona pacientes, citas, profesionales, ordenes, remisiones, historia clinica basica y seguimiento administrativo para clinicas o consultorios multiples. El flujo permite coordinar recepcion, atencion, autorizaciones, resultados, pagos y reportes manteniendo separacion de permisos y trazabilidad por sede o especialidad.",
      summary: "Pacientes, citas, ordenes e historia clinica basica.",
      icon: "/img/report.svg",
      secondaryIcon: "/img/user-avatar-color.svg",
      sections: ["Pacientes", "Citas", "Historia", "Ordenes", "Aprobaciones", "Reportes"]
    },
    {
      id: "linkLaboratorioClinico",
      module: "laboratorio_clinico",
      title: "Laboratorio clinico",
      fullTitle: "Laboratorio clinico",
      lead: "Ordenes de examen, muestras, trazabilidad, resultados, validacion y entrega digital.",
      description: "Administra ordenes de examen, toma de muestras, trazabilidad, control de calidad, resultados, validacion profesional y entrega digital. El laboratorio puede seguir cada muestra por estado, priorizar tiempos de respuesta, documentar incidencias y cerrar servicios con reportes para usuarios, convenios y auditoria.",
      summary: "Ordenes, muestras, resultados y entrega digital.",
      icon: "/img/analytics-color.svg",
      secondaryIcon: "/img/report.svg",
      sections: ["Ordenes", "Muestras", "Resultados", "Calidad", "Entrega", "SLA"]
    },
    {
      id: "linkColegioAcademia",
      module: "colegio_academia",
      title: "Colegio o academia",
      fullTitle: "Colegio, academia o instituto",
      lead: "Estudiantes, matriculas, cursos, mensualidades, asistencia, boletines y certificados.",
      description: "Organiza estudiantes, acudientes, matriculas, cursos, mensualidades, asistencia, boletines, certificados y cartera educativa. La institucion puede controlar grupos, pagos, novedades academicas, documentos emitidos y comunicacion administrativa, conectando la operacion diaria con reportes para direccion y control financiero.",
      summary: "Matriculas, cursos, asistencia, mensualidades y boletines.",
      icon: "/img/company-briefcase-color.svg",
      secondaryIcon: "/img/excel.svg",
      sections: ["Matriculas", "Cursos", "Asistencia", "Mensualidades", "Boletines", "Certificados"]
    },
    {
      id: "linkGuarderiaInfantil",
      module: "guarderia_infantil",
      title: "Guarderia infantil",
      fullTitle: "Guarderia y jardin infantil",
      lead: "Ninos, acudientes, autorizaciones, alimentacion, novedades, entregas y pagos.",
      description: "Controla ninos, acudientes, autorizaciones, alimentacion, novedades, entregas, pagos y seguimiento diario para guarderias o jardines infantiles. El modulo ayuda a registrar entrada y salida, responsables autorizados, observaciones de cuidado, cobros recurrentes y reportes para una operacion segura y ordenada.",
      summary: "Ninos, acudientes, autorizaciones, novedades y pagos.",
      icon: "/img/user-avatar-color.svg",
      secondaryIcon: "/img/report.svg",
      sections: ["Jornada", "Ninos", "Acudientes", "Autorizaciones", "Novedades", "Pagos"]
    },
    {
      id: "linkLavanderiaTintoreria",
      module: "lavanderia_tintoreria",
      title: "Lavanderia",
      fullTitle: "Lavanderia y tintoreria",
      lead: "Ordenes por prenda, etiquetas, lavado, planchado, calidad, entregas y reclamos.",
      description: "Gestiona recepcion de prendas, etiquetas, estados de lavado, planchado, control de calidad, entregas, pagos y reclamos. La lavanderia puede saber donde esta cada orden, evitar perdidas, documentar novedades, programar rutas o domicilios y cerrar servicios con evidencia para el cliente.",
      summary: "Ordenes por prenda, etiquetas, estados y entregas.",
      icon: "/img/report.svg",
      secondaryIcon: "/img/tags-color.svg",
      sections: ["Recepcion", "Prendas", "Etiquetas", "Proceso", "Entrega", "Reclamos"]
    },
    {
      id: "linkTallerMecanico",
      module: "taller_mecanico",
      title: "Taller mecanico",
      fullTitle: "Taller mecanico, motos y autos",
      lead: "Ordenes de trabajo, diagnostico, repuestos, mano de obra, garantia y entrega.",
      description: "Administra ordenes de trabajo, diagnosticos, repuestos, mano de obra, aprobaciones, garantias y entrega de vehiculos. El taller puede registrar entrada, cotizar reparaciones, controlar inventario usado, tomar evidencias, informar avances al cliente y medir tiempos, costos y rentabilidad por servicio.",
      summary: "Ordenes, diagnostico, repuestos, mano de obra y garantias.",
      icon: "/img/settings-color.svg",
      secondaryIcon: "/img/vehiculos-flotas-logo.svg",
      sections: ["Ordenes", "Diagnostico", "Repuestos", "Mano de obra", "Garantias", "Entrega"]
    },
    {
      id: "linkTransporteCargaTMS",
      module: "transporte_carga_tms",
      title: "Transporte TMS",
      fullTitle: "Transporte de carga / TMS",
      lead: "Fletes, manifiestos, conductores, vehiculos, rutas, tracking, entregas y liquidacion.",
      description: "Controla fletes, manifiestos, conductores, vehiculos, rutas, tracking, entregas, novedades y liquidacion para transporte de carga. La empresa puede coordinar operaciones, registrar evidencia de entrega, monitorear cumplimiento, asignar costos y mantener trazabilidad documental desde la solicitud hasta el cierre.",
      summary: "Fletes, manifiestos, conductores, rutas y entregas.",
      icon: "/img/vehiculos-flotas-logo.svg",
      secondaryIcon: "/img/gps.svg",
      sections: ["Fletes", "Manifiestos", "Conductores", "Tracking", "Entregas", "Liquidacion"]
    },
    {
      id: "linkServiciosTecnicos",
      module: "servicios_tecnicos",
      title: "Servicios tecnicos",
      fullTitle: "Servicios tecnicos a domicilio",
      lead: "Ordenes de servicio, tecnicos, visitas, repuestos, firmas, evidencias y garantias.",
      description: "Gestiona ordenes de servicio, agenda de tecnicos, visitas a domicilio, repuestos, firmas, evidencias, garantias y seguimiento posventa. El modulo permite asignar responsables, controlar tiempos de atencion, documentar trabajos realizados, cobrar servicios y mantener historial por cliente, equipo o ubicacion.",
      summary: "Ordenes a domicilio, tecnicos, repuestos, firmas y evidencias.",
      icon: "/img/network-color.svg",
      secondaryIcon: "/img/settings-color.svg",
      sections: ["Ordenes", "Agenda", "Tecnicos", "Repuestos", "Firmas", "Garantias"]
    },
    {
      id: "linkInmobiliariaComercial",
      module: "inmobiliaria_comercial",
      title: "Inmobiliaria",
      fullTitle: "Inmobiliaria comercial",
      lead: "Propiedades, arriendos, ventas, visitas, leads, propietarios, contratos y comisiones.",
      description: "Administra propiedades, arriendos, ventas, visitas, leads, propietarios, contratos, comisiones y seguimiento comercial inmobiliario. La agencia puede publicar inventario, programar recorridos, registrar interesados, controlar documentacion, medir conversiones y cerrar negocios con trazabilidad financiera y contractual.",
      summary: "Propiedades, visitas, leads, propietarios y comisiones.",
      icon: "/img/company-briefcase-color.svg",
      secondaryIcon: "/img/money.svg",
      sections: ["Propiedades", "Leads", "Visitas", "Contratos", "Comisiones", "Cierre"]
    },
    {
      id: "linkSeguridadPrivada",
      module: "seguridad_privada",
      title: "Seguridad privada",
      fullTitle: "Seguridad privada y vigilancia",
      lead: "Guardas, puestos, turnos, rondas QR, novedades, incidentes y escalamiento.",
      description: "Controla guardas, puestos, turnos, rondas QR, novedades, incidentes, escalamiento y reportes para empresas de seguridad privada. La operacion puede verificar presencia, documentar eventos, asignar supervisores, medir cumplimiento de rondas y conservar evidencia para clientes, auditoria y control interno.",
      summary: "Guardas, puestos, turnos, rondas QR e incidentes.",
      icon: "/img/shield-license-color.svg",
      secondaryIcon: "/img/audit.svg",
      sections: ["Puestos", "Turnos", "Guardas", "Rondas QR", "Incidentes", "Reportes"]
    },
    {
      id: "linkClubDeportivo",
      module: "club_deportivo",
      title: "Club deportivo",
      fullTitle: "Club deportivo y escuela deportiva",
      lead: "Alumnos, entrenadores, disciplinas, clases, torneos, pagos y asistencia.",
      description: "Gestiona alumnos, entrenadores, disciplinas, clases, torneos, pagos, asistencia y comunicaciones para clubes o escuelas deportivas. El modulo ayuda a organizar grupos, programar entrenamientos, controlar mensualidades, registrar participacion en eventos y medir avance operativo por disciplina o categoria.",
      summary: "Alumnos, entrenadores, disciplinas, torneos y asistencia.",
      icon: "/img/analytics-color.svg",
      secondaryIcon: "/img/customer.svg",
      sections: ["Alumnos", "Disciplinas", "Clases", "Torneos", "Pagos", "Asistencia"]
    },
    {
      id: "linkFunerariaExequial",
      module: "funeraria_exequial",
      title: "Funeraria exequial",
      fullTitle: "Funeraria y servicios exequiales",
      lead: "Planes, afiliados, salas, servicios, traslados, contratos y documentacion.",
      description: "Administra planes exequiales, afiliados, contratos, salas, servicios, traslados, documentacion y cierre operativo para funerarias. La empresa puede coordinar atenciones sensibles, controlar obligaciones, registrar pagos, organizar recursos y mantener expedientes claros con permisos, evidencias y reportes.",
      summary: "Planes, afiliados, salas, servicios y documentos.",
      icon: "/img/report.svg",
      secondaryIcon: "/img/company-briefcase-color.svg",
      sections: ["Planes", "Afiliados", "Servicios", "Salas", "Documentos", "Cierre"]
    },
    {
      id: "linkParqueRecreativo",
      module: "parque_recreativo",
      title: "Parque recreativo",
      fullTitle: "Parque recreativo y atracciones",
      lead: "Entradas, manillas QR, aforo, atracciones, consumo interno, incidentes y cierre.",
      description: "Controla entradas, manillas QR, aforo, atracciones, consumo interno, incidentes y cierre de jornada para parques recreativos. La operacion puede vender accesos, validar ingreso, monitorear ocupacion, registrar novedades, administrar puntos de consumo y evaluar ventas, seguridad y experiencia del visitante.",
      summary: "Entradas, manillas QR, aforo, atracciones y consumo.",
      icon: "/img/tags-color.svg",
      secondaryIcon: "/img/analytics-color.svg",
      sections: ["Entradas", "Manillas QR", "Aforo", "Atracciones", "Consumo", "Incidentes"]
    },
    {
      id: "linkCooperativaFondo",
      module: "cooperativa_fondo",
      title: "Cooperativa",
      fullTitle: "Cooperativa y fondo de empleados",
      lead: "Asociados, aportes, creditos internos, cartera, beneficios, convenios y asambleas.",
      description: "Gestiona asociados, aportes, creditos internos, cartera, beneficios, convenios, asambleas y trazabilidad administrativa para cooperativas o fondos de empleados. El modulo permite controlar solicitudes, aprobaciones, obligaciones, recaudos y comunicaciones manteniendo gobierno interno, reportes y permisos por responsabilidad.",
      summary: "Asociados, aportes, creditos internos, cartera y beneficios.",
      icon: "/img/money.svg",
      secondaryIcon: "/img/report.svg",
      sections: ["Asociados", "Aportes", "Creditos", "Cartera", "Beneficios", "Asambleas"]
    },
    {
      id: "linkCapacitacionEmpresarial",
      module: "capacitacion_empresarial",
      title: "Capacitacion empresarial",
      fullTitle: "Centro de capacitacion empresarial",
      lead: "Cursos, instructores, cohortes, asistencia, evaluaciones, certificados y empresas cliente.",
      description: "Administra cursos, instructores, cohortes, empresas cliente, asistencia, evaluaciones, certificados y seguimiento comercial para centros de capacitacion. La operacion puede vender programas, controlar grupos, medir cumplimiento academico, emitir constancias y conservar historial por participante, empresa y periodo.",
      summary: "Cursos, instructores, certificados, asistencia y empresas.",
      icon: "/img/excel.svg",
      secondaryIcon: "/img/company-briefcase-color.svg",
      sections: ["Cursos", "Cohortes", "Instructores", "Asistencia", "Evaluaciones", "Certificados"]
    }
  ];

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
    colegio_academia: 12,
    guarderia_infantil: 13,
    inmobiliaria_comercial: 14,
    seguridad_privada: 15,
    club_deportivo: 16,
    funeraria_exequial: 17,
    parque_recreativo: 18,
    cooperativa_fondo: 19,
    capacitacion_empresarial: 20
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
