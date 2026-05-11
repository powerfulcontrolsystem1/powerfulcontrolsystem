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
      summary: "Cursos, instructores, certificados, asistencia y empresas.",
      icon: "/img/excel.svg",
      secondaryIcon: "/img/company-briefcase-color.svg",
      sections: ["Cursos", "Cohortes", "Instructores", "Asistencia", "Evaluaciones", "Certificados"]
    }
  ];

  modules.forEach(function (item) {
    if (!Array.isArray(item.sections) || item.sections.length === 0) {
      item.sections = defaultSections.slice();
    }
  });

  window.PCS_NUEVOS_VERTICALES = modules;
  window.PCS_NUEVOS_VERTICALES_MODULES = modules.map(function (item) {
    return [item.id, item.module];
  });
  window.PCS_NUEVOS_VERTICALES_KEYS = modules.map(function (item) {
    return item.module;
  });
  window.PCS_NUEVOS_VERTICALES_CSV = window.PCS_NUEVOS_VERTICALES_KEYS.join(",");
})();
