function getQueryParam(name) {
  var params = new URLSearchParams(window.location.search);
  var value = params.get(name);
  if (value) {
    return value;
  }
  try {
    if (window.parent && window.parent !== window && window.parent.location) {
      var parentParams = new URLSearchParams(window.parent.location.search || "");
      var parentValue = parentParams.get(name);
      if (parentValue) {
        return parentValue;
      }
    }
  } catch (e) {
    // no-op: acceso a parent puede fallar en algunos contextos
  }
  return "";
}

function parsePositiveInt(raw) {
  var n = Number(String(raw || "").trim());
  if (!Number.isFinite(n)) return 0;
  n = Math.trunc(n);
  return n > 0 ? n : 0;
}

function readEmpresaIdFromStorage() {
  var keys = ["active_empresa_id", "empresa_id", "admin_empresa_id"];
  var stores = [];
  try { stores.push(window.sessionStorage); } catch (e) {}
  try { stores.push(window.localStorage); } catch (e) {}

  for (var s = 0; s < stores.length; s += 1) {
    var store = stores[s];
    if (!store) continue;
    for (var i = 0; i < keys.length; i += 1) {
      var key = keys[i];
      var raw = "";
      try {
        raw = store.getItem(key) || "";
      } catch (e) {
        raw = "";
      }
      var id = parsePositiveInt(raw);
      if (id > 0) {
        return String(id);
      }
    }
  }
  return "";
}

function persistEmpresaIdInStorage(rawEmpresaId) {
  var id = parsePositiveInt(rawEmpresaId);
  if (!id) return "";
  var value = String(id);
  try {
    window.sessionStorage.setItem("active_empresa_id", value);
    window.sessionStorage.setItem("empresa_id", value);
    window.sessionStorage.setItem("admin_empresa_id", value);
  } catch (e) {}
  try {
    window.localStorage.setItem("active_empresa_id", value);
    window.localStorage.setItem("empresa_id", value);
    window.localStorage.setItem("admin_empresa_id", value);
  } catch (e) {}
  return value;
}

function resolveEmpresaIdContext() {
  var fromUrl = parsePositiveInt(getQueryParam("empresa_id") || getQueryParam("id"));
  if (fromUrl > 0) {
    return persistEmpresaIdInStorage(fromUrl);
  }
  var fromStorage = readEmpresaIdFromStorage();
  if (fromStorage) {
    return persistEmpresaIdInStorage(fromStorage);
  }
  return "";
}

try {
  window.__resolveEmpresaIdContext = function () {
    return resolveEmpresaIdContext();
  };
} catch (e) {
  // no-op
}

(function () {
  var id = persistEmpresaIdInStorage(getQueryParam("id") || getQueryParam("empresa_id"));
  if (!id) {
    id = resolveEmpresaIdContext();
  }
  var titleMenu = document.getElementById("empresaTitleMenu");
  var empresaNameMenu = document.getElementById("empresaNameMenu");
  var title = titleMenu || document.getElementById("empresaTitle");
  var frame = document.getElementById("contentFrame") || document.querySelector("iframe.admin-empresa-frame");
  var frameResizeObserver = null;
  var favoriteBtn = document.getElementById("adminFavoriteBtn");
  var frameTargetName = frame ? String(frame.getAttribute("name") || frame.name || frame.id || "").trim() : "";
  var initialFrameSrc = frame ? normalizeHref(frame.getAttribute("src") || frame.src || "") : "";
  var portalUsuariosLink = document.getElementById("linkPortalUsuarios");
  var companySelectorLink = document.querySelector("a.select-company");
  var permsEvidence = document.getElementById("menuPermsEvidence");
  var verticalIntegrationEvidence = document.getElementById("verticalIntegrationEvidence");
  var nuevasPlantillasCatalog = Array.isArray(window.PCS_NUEVAS_PLANTILLAS)
    ? window.PCS_NUEVAS_PLANTILLAS.slice()
    : [];
  var verticalIntegration = window.PCS_VERTICAL_INTEGRATION || null;
  var nuevasPlantillasModules = Array.isArray(window.PCS_NUEVAS_PLANTILLAS_MODULES)
    ? window.PCS_NUEVAS_PLANTILLAS_MODULES.slice()
    : nuevasPlantillasCatalog.map(function (item) { return [item.id, item.module]; });
  var nuevasPlantillasMenuLinks = renderNuevasPlantillasMenuLinks();
  var storage = null;
  try {
    storage = window.sessionStorage;
  } catch (e) {
    storage = null;
  }
  var links = [
    document.getElementById("linkVentaDirecta"),
    document.getElementById("linkEstaciones"),
    document.getElementById("linkPanelEmpresa"),
    document.getElementById("linkProductos"),
    document.getElementById("linkInventarioAvanzado"),
    document.getElementById("linkCompras"),
    document.getElementById("linkComprasAvanzadas"),
    document.getElementById("linkSoportesComprasIA"),
    document.getElementById("linkImportacionesCosteo"),
    document.getElementById("linkProduccionMRP"),
    document.getElementById("linkLogisticaWMS"),
    document.getElementById("linkPlantillasIntegracion"),
    document.getElementById("linkGimnasio"),
    document.getElementById("linkTaxiSystem"),
    document.getElementById("linkParqueadero"),
    document.getElementById("linkApartamentosTuristicos"),
    document.getElementById("linkPropiedadHorizontal"),
    document.getElementById("linkAlquileres"),
    document.getElementById("linkTurnosAtencion"),
    document.getElementById("linkConfiguracion"),
    document.getElementById("linkEmpresasCompartidas"),
    document.getElementById("linkLicenciaSistema"),
    document.getElementById("linkConfiguracionMain"),
    document.getElementById("linkConfiguracionIdentidadVisual"),
    document.getElementById("linkConfiguracionCobroOperativo"),
    document.getElementById("linkConfiguracionReporteCorte"),
    document.getElementById("linkConfiguracionBackupsPasarelas"),
    document.getElementById("linkConfiguracionPasarelasPago"),
    document.getElementById("linkConfiguracionImpresora"),
    document.getElementById("linkConfiguracionPermisos"),
    document.getElementById("linkConfiguracionGuiada"),
    document.getElementById("linkConfiguracionChatFlotante"),
    document.getElementById("linkConfiguracionAvanzada"),
    document.getElementById("linkConfiguracionCarritoEmpresa"),
    document.getElementById("linkCarritoCompras"),
    document.getElementById("linkFacturacionElectronica"),
    document.getElementById("linkFacturacionMain"),
    document.getElementById("linkFacturacionEcuador"),
    document.getElementById("linkFacturacionPanama"),
    document.getElementById("linkPruebasDian"),
    document.getElementById("linkProveedoresFirmaDigital"),
    document.getElementById("linkFacturasElectronicas"),
    document.getElementById("linkAIUConstruccion"),
    document.getElementById("linkChatIA"),
    document.getElementById("linkFinanzas"),
    document.getElementById("linkFinanzasMain"),
    document.getElementById("linkContabilidadColombia"),
    document.getElementById("linkContabilidadColombiaAvanzada"),
    document.getElementById("linkCentrosCosto"),
    document.getElementById("linkBancosPagos"),
    document.getElementById("linkCierreFiscal"),
    document.getElementById("linkActivosFijosNIIF"),
    document.getElementById("linkDeclaracionesTributarias"),
    document.getElementById("linkPortalTercerosCertificados"),
    document.getElementById("linkCumplimientoKYC"),
    document.getElementById("linkTesoreriaPresupuesto"),
    document.getElementById("linkBackups"),
    document.getElementById("linkSoporteRemoto"),
    document.getElementById("linkUbicacionGPS"),
    document.getElementById("linkReservasHotel"),
    document.getElementById("linkTarifasHotel"),
    document.getElementById("linkTarifasMotel"),
    document.getElementById("linkHotelTarjetasAcceso"),
    document.getElementById("linkControlElectrico"),
    document.getElementById("linkConsultorioOdontologico"),
    document.getElementById("linkDrogueriaFarmacia"),
    document.getElementById("linkDomicilios"),
    document.getElementById("linkReportes"),
    document.getElementById("linkReportesEjecutivos"),
    document.getElementById("linkReportesTurnos"),
    document.getElementById("linkUsuarios"),
    document.getElementById("linkPortalUsuarios"),
    document.getElementById("linkMiHorario"),
    document.getElementById("linkHorariosTrabajadores"),
    document.getElementById("linkCodigosDescuento"),
    document.getElementById("linkCorteCaja"),
    document.getElementById("linkGeneradorCodigosBarras"),
    document.getElementById("linkAsistenciaEmpleados"),
    document.getElementById("linkCarnets"),
    document.getElementById("linkVehiculosRegistro"),
    document.getElementById("linkHojaVidaOperativa"),
    document.getElementById("linkAuditoria"),
    document.getElementById("linkCalidadProcesos"),
    document.getElementById("linkEnergiaSolar"),
    document.getElementById("linkChatTareas"),
    document.getElementById("linkClientes"),
    document.getElementById("linkCRMComercial"),
    document.getElementById("linkVentaPublica"),
    document.getElementById("linkRedSocialComercial"),
    document.getElementById("linkDocumentosOnlyOffice"),
    document.getElementById("linkGestionDocumental"),
    document.getElementById("linkContratosObligaciones"),
    document.getElementById("linkRadioOnline"),
    document.getElementById("linkImpuestos"),
    document.getElementById("linkEgresosIngresos"),
    document.getElementById("linkCreditos"),
    document.getElementById("linkCreditosMenu"),
    document.getElementById("linkCreditosPanelMenu"),
    document.getElementById("linkCreditosCrearMenu"),
    document.getElementById("linkCreditosCarteraMenu"),
    document.getElementById("linkCreditosMorosidadMenu"),
    document.getElementById("linkCreditosLimitesMenu"),
    document.getElementById("linkCreditosOperacionesMenu"),
    document.getElementById("linkCreditosAprobacionesMenu"),
    document.getElementById("linkCreditosEstadoMenu"),
    document.getElementById("linkCobranza"),
    document.getElementById("linkCentrosCostoMenu"),
    document.getElementById("linkCierreFiscalMenu"),
    document.getElementById("linkActivosFijosNIIFMenu"),
    document.getElementById("linkDeclaracionesTributariasMenu"),
    document.getElementById("linkCobranzaMenu"),
    document.getElementById("linkPortalContadorMenu"),
    document.getElementById("linkPortalTercerosCertificadosMenu"),
    document.getElementById("linkNominaMenu"),
    document.getElementById("linkERPExtendido"),
    document.getElementById("linkERPExtendidoMenu"),
    document.getElementById("linkPropinas"),
    document.getElementById("linkComisiones"),
    document.getElementById("linkConfigEstaciones"),
    document.getElementById("linkConfiguracionSensoresRaspberry"),
    document.getElementById("linkTarifasPorMinutos"),
    document.getElementById("linkTarifasPorDia"),
    document.getElementById("linkFrecuenciaFE"),
  ].concat(nuevasPlantillasMenuLinks);
  var frameLinks = [];

  var permActionRead = "R";
  var permActionCreate = "C";
  var permActionUpdate = "U";
  var permActionApprove = "A";

  var permModuleVentas = "ventas";
  var permModuleInventario = "inventario";
  var permModuleCompras = "compras";
  var permModuleFinanzas = "finanzas";
  var permModuleContabilidadCO = "contabilidad_colombia";
  var permModuleContabilidadCOAv = "contabilidad_colombia_avanzada";
  var permModuleCentrosCosto = "centros_costo";
  var permModuleCierreFiscal = "cierre_fiscal";
  var permModuleActivosFijosNIIF = "activos_fijos_niif_fiscal";
  var permModuleDeclaracionesTrib = "declaraciones_tributarias";
  var permModuleTesoreria = "tesoreria_presupuesto";
  var permModuleImportaciones = "importaciones_costeo";
  var permModuleLogisticaWMS = "logistica_wms";
  var permModuleCobranza = "cobranza";
  var permModulePortalContador = "portal_contador";
  var permModulePortalTerceros = "portal_terceros_certificados";
  var permModuleSoportesComprasIA = "soportes_compras_ia";
  var permModuleBancosPagos = "bancos_pagos";
  var permModuleGestionDocumental = "gestion_documental";
  var permModuleCumplimientoKYC = "cumplimiento_kyc";
  var permModuleContratosObligaciones = "contratos_obligaciones";
  var permModuleCalidadProcesos = "calidad_procesos";
  var permModuleDrogueriaFarmacia = "drogueria_farmacia";
  var permModuleAIUConstruccion = "aiu_construccion";
  var permModuleClientes = "clientes";
  var permModuleCRMUnificado = "crm_unificado";
  var permModuleFacturacion = "facturacion";
  var permModuleFacturacionEcuador = "facturacion_ecuador";
  var permModuleFacturacionPanama = "facturacion_panama";
  var permModuleSeguridad = "seguridad";
  var permModuleVentaPublica = "venta_publica";
  var permModuleReservasHotel = "reservas_hotel";
  var permModuleChatTareas = "chat_tareas";
  var permModuleGimnasio = "gimnasio";
  var permModuleTaxiSystem = "taxi_system";
  var permModuleDomicilios = "domicilios";
  var permModuleParqueadero = "parqueadero";
  var permModuleApartTuristicos = "apartamentos_turisticos";
  var permModulePropiedadHorizontal = "propiedad_horizontal";
  var permModuleAlquileres = "alquileres";
  var permModuleOdontologia = "odontologia";
  var permModuleTurnos = "turnos_atencion";
  var permModuleControlElectrico = "control_electrico";
  var permModuleCarnets = "carnets";
  var permModuleProduccionMRP = "produccion_mrp";
  var permModuleHorariosTrab = "horarios_trabajadores";
  var permModuleAsistenciaEmpleados = "asistencia_empleados";
  var permModuleVehiculosRegistro = "vehiculos_registro";
  var permModuleHojaVidaOperativa = "hoja_vida_operativa";
  var permModuleUbicacionGPS = "ubicacion_gps";
  var permModuleNominaSueldos = "nomina_sueldos";
  var permModuleReportes = "reportes";
  var permModuleAuditoria = "auditoria";
  var permModuleEnergiaSolar = "energia_solar";
  var permModuleBackups = "backups";
  var permModuleDocumentosOnlyOffice = "documentos_onlyoffice";
  var menuPermissionCatalog = {
    linkCarritoCompras: { module: permModuleVentas, action: permActionCreate },
    linkVentaDirecta: { module: permModuleVentas, action: permActionCreate },
    linkCodigosDescuento: { module: permModuleVentas, action: permActionCreate },
    linkRedSocialComercial: { module: permModuleVentas, action: permActionCreate },
    linkChatIA: { module: permModuleVentas, action: permActionRead },
    linkConfigEstaciones: { module: permModuleVentas, action: permActionApprove },
    linkTarifasPorMinutos: { module: permModuleVentas, action: permActionCreate },
    linkTarifasPorDia: { module: permModuleVentas, action: permActionCreate },
    linkTarifasHotel: { module: permModuleVentas, action: permActionCreate },
    linkTarifasMotel: { module: permModuleVentas, action: permActionCreate },
    linkReservasHotel: { module: permModuleReservasHotel, action: permActionCreate },
    linkHotelTarjetasAcceso: { module: permModuleReservasHotel, action: permActionCreate },
    linkChatTareas: { module: permModuleChatTareas, action: permActionCreate },
    linkConfiguracionChatFlotante: { module: permModuleChatTareas, action: permActionUpdate },
    linkTurnosAtencion: { module: permModuleTurnos, action: permActionCreate },
    linkPlantillasIntegracion: { module: permModuleSeguridad, action: permActionRead },

    linkProductos: { module: permModuleInventario, action: permActionCreate },
    linkProductosMain: { module: permModuleInventario, action: permActionCreate },
    linkInventarioAvanzado: { module: permModuleInventario, action: permActionCreate },
    linkRecetasProductos: { module: permModuleInventario, action: permActionCreate },
    linkPreciosHistorial: { module: permModuleInventario, action: permActionRead },
    linkBodegas: { module: permModuleInventario, action: permActionUpdate },
    linkCategorias: { module: permModuleInventario, action: permActionUpdate },
    linkGeneradorCodigosBarras: { module: permModuleInventario, action: permActionUpdate },
    linkCompras: { module: permModuleCompras, action: permActionCreate },
    linkComprasDoc: { module: permModuleCompras, action: permActionCreate },
    linkProveedores: { module: permModuleCompras, action: permActionCreate },
    linkComprasAvanzadas: { module: permModuleCompras, action: permActionCreate },
    linkSoportesComprasIA: { module: permModuleSoportesComprasIA, action: permActionCreate },
    linkSoportesComprasIAMenu: { module: permModuleSoportesComprasIA, action: permActionCreate },
    linkImportacionesCosteo: { module: permModuleImportaciones, action: permActionCreate },
    linkProduccionMRP: { module: permModuleProduccionMRP, action: permActionCreate },
    linkLogisticaWMS: { module: permModuleLogisticaWMS, action: permActionCreate },
    linkCartaProductosPublica: { module: permModuleVentaPublica, action: permActionCreate },
    linkVentaPublica: { module: permModuleVentaPublica, action: permActionCreate },
    linkConfiguracionCarritoEmpresa: { module: permModuleVentaPublica, action: permActionApprove },

    linkGimnasio: { module: permModuleGimnasio, action: permActionCreate },
    linkGimnasioDashboard: { module: permModuleGimnasio, action: permActionRead },
    linkGimnasioSocios: { module: permModuleGimnasio, action: permActionCreate },
    linkGimnasioPlanes: { module: permModuleGimnasio, action: permActionUpdate },
    linkGimnasioEntrenadores: { module: permModuleGimnasio, action: permActionUpdate },
    linkGimnasioClases: { module: permModuleGimnasio, action: permActionCreate },
    linkGimnasioInscripciones: { module: permModuleGimnasio, action: permActionCreate },
    linkGimnasioAsistencias: { module: permModuleGimnasio, action: permActionCreate },
    linkGimnasioPagos: { module: permModuleGimnasio, action: permActionCreate },
    linkGimnasioAcceso: { module: permModuleGimnasio, action: permActionApprove },
    linkTaxiSystem: { module: permModuleTaxiSystem, action: permActionCreate },
    linkDomicilios: { module: permModuleDomicilios, action: permActionCreate },
    linkParqueadero: { module: permModuleParqueadero, action: permActionCreate },
    linkApartamentosTuristicos: { module: permModuleApartTuristicos, action: permActionCreate },
    linkPropiedadHorizontal: { module: permModulePropiedadHorizontal, action: permActionCreate },
    linkAlquileres: { module: permModuleAlquileres, action: permActionCreate },
    linkConsultorioOdontologico: { module: permModuleOdontologia, action: permActionCreate },
    linkDrogueriaFarmacia: { module: permModuleDrogueriaFarmacia, action: permActionCreate },
    linkAIUConstruccion: { module: permModuleAIUConstruccion, action: permActionCreate },

    linkClientes: { module: permModuleClientes, action: permActionCreate },
    linkCRMComercial: { module: permModuleCRMUnificado, action: permActionCreate },

    linkFinanzas: { module: permModuleFinanzas, action: permActionCreate },
    linkFinanzasMain: { module: permModuleFinanzas, action: permActionCreate },
    linkEgresosIngresos: { module: permModuleFinanzas, action: permActionCreate },
    linkEgresos: { module: permModuleFinanzas, action: permActionCreate },
    linkIngresos: { module: permModuleFinanzas, action: permActionCreate },
    linkCorteCaja: { alwaysVisible: true },
    linkCreditos: { module: permModuleFinanzas, action: permActionCreate },
    linkCreditosMenu: { module: permModuleFinanzas, action: permActionCreate },
    linkCreditosPanelMenu: { module: permModuleFinanzas, action: permActionCreate },
    linkCreditosCrearMenu: { module: permModuleFinanzas, action: permActionCreate },
    linkCreditosCarteraMenu: { module: permModuleFinanzas, action: permActionCreate },
    linkCreditosMorosidadMenu: { module: permModuleFinanzas, action: permActionCreate },
    linkCreditosLimitesMenu: { module: permModuleFinanzas, action: permActionCreate },
    linkCreditosOperacionesMenu: { module: permModuleFinanzas, action: permActionCreate },
    linkCreditosAprobacionesMenu: { module: permModuleFinanzas, action: permActionCreate },
    linkCreditosEstadoMenu: { module: permModuleFinanzas, action: permActionCreate },
    linkPropinas: { module: permModuleFinanzas, action: permActionCreate },
    linkComisiones: { module: permModuleFinanzas, action: permActionCreate },
    linkContabilidadColombia: { module: permModuleContabilidadCO, action: permActionCreate },
    linkContabilidadColombiaAvanzada: { module: permModuleContabilidadCOAv, action: permActionCreate },
    linkCentrosCosto: { module: permModuleCentrosCosto, action: permActionCreate },
    linkCentrosCostoMenu: { module: permModuleCentrosCosto, action: permActionCreate },
    linkBancosPagos: { module: permModuleBancosPagos, action: permActionCreate },
    linkCierreFiscal: { module: permModuleCierreFiscal, action: permActionApprove },
    linkCierreFiscalMenu: { module: permModuleCierreFiscal, action: permActionApprove },
    linkActivosFijosNIIF: { module: permModuleActivosFijosNIIF, action: permActionCreate },
    linkActivosFijosNIIFMenu: { module: permModuleActivosFijosNIIF, action: permActionCreate },
    linkDeclaracionesTributarias: { module: permModuleDeclaracionesTrib, action: permActionCreate },
    linkDeclaracionesTributariasMenu: { module: permModuleDeclaracionesTrib, action: permActionCreate },
    linkTesoreriaPresupuesto: { module: permModuleTesoreria, action: permActionCreate },
    linkCobranza: { module: permModuleCobranza, action: permActionCreate },
    linkCobranzaMenu: { module: permModuleCobranza, action: permActionCreate },
    linkPortalContador: { module: permModulePortalContador, action: permActionCreate },
    linkPortalContadorMenu: { module: permModulePortalContador, action: permActionCreate },
    linkPortalTercerosCertificados: { module: permModulePortalTerceros, action: permActionCreate },
    linkPortalTercerosCertificadosMenu: { module: permModulePortalTerceros, action: permActionCreate },
    linkCumplimientoKYC: { module: permModuleCumplimientoKYC, action: permActionApprove },
    linkNominaSueldos: { module: permModuleNominaSueldos, action: permActionCreate },
    linkNominaMenu: { module: permModuleNominaSueldos, action: permActionCreate },

    linkFacturacionElectronica: { anyModules: [permModuleFacturacion, permModuleFacturacionEcuador, permModuleFacturacionPanama], action: permActionCreate },
    linkFacturacionMain: { module: permModuleFacturacion, action: permActionCreate },
    linkFacturacionEcuador: { module: permModuleFacturacionEcuador, action: permActionCreate },
    linkFacturacionPanama: { module: permModuleFacturacionPanama, action: permActionCreate },
    linkPruebasDian: { module: permModuleFacturacion, action: permActionApprove },
    linkProveedoresFirmaDigital: { module: permModuleFacturacion, action: permActionRead },
    linkFacturasElectronicas: { module: permModuleFacturacion, action: permActionRead },
    linkImpuestos: { module: permModuleFacturacion, action: permActionUpdate },
    linkFrecuenciaFE: { module: permModuleFacturacion, action: permActionApprove },

    linkReportes: { module: permModuleReportes, action: permActionRead },
    linkReportesEjecutivos: { module: permModuleReportes, action: permActionRead },
    linkReportesTurnos: { module: permModuleReportes, action: permActionRead },
    linkCalculadora: { module: permModuleFinanzas, action: permActionRead },

    linkUsuarios: { module: permModuleSeguridad, action: permActionUpdate },
    linkPortalUsuarios: { module: permModuleSeguridad, action: permActionRead },
    linkMiHorario: { module: permModuleHorariosTrab, action: permActionRead },
    linkHorariosTrabajadores: { module: permModuleHorariosTrab, action: permActionUpdate },
    linkAsistenciaEmpleados: { module: permModuleAsistenciaEmpleados, action: permActionUpdate },
    linkCarnets: { module: permModuleCarnets, action: permActionCreate },
    linkVehiculosRegistro: { module: permModuleVehiculosRegistro, action: permActionCreate },
    linkHojaVidaOperativa: { module: permModuleHojaVidaOperativa, action: permActionUpdate },
    linkUbicacionGPS: { module: permModuleUbicacionGPS, action: permActionCreate },

    linkAuditoria: { module: permModuleAuditoria, action: permActionRead },
    linkCalidadProcesos: { module: permModuleCalidadProcesos, action: permActionCreate },
    linkEnergiaSolar: { module: permModuleEnergiaSolar, action: permActionCreate },
    linkBackups: { module: permModuleBackups, action: permActionApprove },

    linkDocumentosOnlyOffice: { module: permModuleDocumentosOnlyOffice, action: permActionRead },
    linkGestionDocumental: { module: permModuleGestionDocumental, action: permActionCreate },
    linkContratosObligaciones: { module: permModuleContratosObligaciones, action: permActionCreate },
    linkSoporteRemoto: { module: permModuleSeguridad, action: permActionApprove },

    linkConfiguracion: { module: permModuleSeguridad, action: permActionUpdate },
    linkEmpresasCompartidas: { module: permModuleSeguridad, action: permActionUpdate },
    linkLicenciaSistema: { module: permModuleSeguridad, action: permActionRead },
    linkConfiguracionMain: { module: permModuleSeguridad, action: permActionUpdate },
    linkConfiguracionIdentidadVisual: { module: permModuleSeguridad, action: permActionUpdate },
    linkConfiguracionCobroOperativo: { module: permModuleSeguridad, action: permActionUpdate },
    linkConfiguracionReporteCorte: { module: permModuleSeguridad, action: permActionUpdate },
    linkConfiguracionBackupsPasarelas: { module: permModuleSeguridad, action: permActionUpdate },
    linkConfiguracionPasarelasPago: { module: permModuleSeguridad, action: permActionUpdate },
    linkConfiguracionImpresora: { module: permModuleSeguridad, action: permActionUpdate },
    linkConfiguracionPermisos: { module: permModuleSeguridad, action: permActionApprove },
    linkConfiguracionAvanzada: { module: permModuleSeguridad, action: permActionUpdate },
    linkConfiguracionGuiada: { module: permModuleSeguridad, action: permActionUpdate },
    linkConfiguracionSensoresRaspberry: { module: permModuleControlElectrico, action: permActionUpdate },
    linkControlElectrico: { module: permModuleControlElectrico, action: permActionUpdate },
    linkRadioOnline: { module: permModuleSeguridad, action: permActionRead },
    linkERPExtendido: { module: permModuleSeguridad, action: permActionUpdate },
    linkERPExtendidoMenu: { module: permModuleSeguridad, action: permActionUpdate },
    linkChatIAGlobal: { module: permModuleSeguridad, action: permActionRead },
    linkEstaciones: { alwaysVisible: true },
    linkPanelEmpresa: { alwaysVisible: true }
  };
  nuevasPlantillasModules.forEach(function (item) {
    menuPermissionCatalog[item[0]] = { module: item[1], action: permActionCreate };
  });

  function isNuevoVerticalModule(module) {
    var normalized = String(module || "").trim().toLowerCase();
    return nuevasPlantillasModules.some(function (item) { return item[1] === normalized; });
  }

  function verticalIsOperationalVisible(module) {
    var normalized = String(module || "").trim().toLowerCase();
    if (!normalized) return true;
    if (verticalIntegration && typeof verticalIntegration.isOperationalVisible === "function") {
      return verticalIntegration.isOperationalVisible(normalized);
    }
    return true;
  }

  function menuLinkPassesVerticalIntegration(link) {
    if (!link) return true;
    var module = String(link.getAttribute("data-vertical-module") || "").trim().toLowerCase();
    var rule = menuPermissionCatalog[link.id || ""];
    if (!module && rule && rule.module) {
      module = String(rule.module || "").trim().toLowerCase();
    }
    return verticalIsOperationalVisible(module);
  }

  function escHTML(value) {
    return String(value == null ? "" : value).replace(/[&<>"']/g, function (ch) {
      return {"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#39;"}[ch];
    });
  }

  function renderNuevasPlantillasMenuLinks() {
    var mount = document.getElementById("adminBusinessVerticalsMount");
    if (!mount || !nuevasPlantillasCatalog.length) {
      return [];
    }
    Array.prototype.slice.call(document.querySelectorAll(".admin-business-vertical-item")).forEach(function (item) {
      if (item && item.parentElement) item.parentElement.removeChild(item);
    });
    var html = nuevasPlantillasCatalog.filter(function (item) {
      if (item && item.operationalVisible === false) return false;
      return verticalIsOperationalVisible(item && item.module);
    }).map(function (item) {
      var id = String(item.id || "").trim();
      var module = String(item.module || "").trim();
      if (!id || !module) return "";
      var title = String(item.title || item.fullTitle || module).trim();
      var icon = String(item.icon || "/img/company-briefcase-color.svg").trim();
      var href = "/administrar_empresa/modulo_menu.html?module=" + encodeURIComponent(module);
      return '<li class="admin-business-vertical-item"><a id="' + escHTML(id) + '" href="' + escHTML(href) + '" target="contentFrame" data-vertical-module="' + escHTML(module) + '">' +
        '<img class="icon" src="' + escHTML(icon) + '" alt="">' + escHTML(title) +
        '</a></li>';
    }).join("");
    mount.insertAdjacentHTML("beforebegin", html);
    return Array.prototype.slice.call(document.querySelectorAll(".admin-business-vertical-item a[data-vertical-module]"));
  }

  function storageKey(empresaId) {
    return "admin_empresa:last_page:" + String(empresaId || "global");
  }

  function getFrameLinks() {
    if (!frame) return [];
    var navLinks = Array.prototype.slice.call(document.querySelectorAll(".admin-sidebar .nav a[target]"));
    var filtered = navLinks.filter(function (link) {
      if (!link) return false;
      var target = String(link.getAttribute("target") || "").trim();
      if (!target) return false;
      if (!frameTargetName) return true;
      return target === frameTargetName;
    });
    if (filtered.length > 0) {
      return filtered;
    }
    return links.filter(function (link) {
      return !!link;
    });
  }

  function normalizeHref(href) {
    var raw = String(href || "").trim();
    if (!raw) return "";
    try {
      var url = new URL(raw, window.location.origin);
      return url.pathname + url.search;
    } catch (e) {
      return "";
    }
  }

  function isAllowedFrameHref(href) {
    var normalized = normalizeHref(href);
    return normalized.indexOf("/administrar_empresa/") === 0;
  }

  function defaultFrameSrc(empresaId) {
    if (initialFrameSrc && isAllowedFrameHref(initialFrameSrc)) {
      return withEmpresaParam(initialFrameSrc, empresaId) || initialFrameSrc;
    }
    var activeLinks = frameLinks.length > 0 ? frameLinks : getFrameLinks();
    for (var i = 0; i < activeLinks.length; i += 1) {
      var link = activeLinks[i];
      if (!link) continue;
      var href = withEmpresaParam(link.getAttribute("href"), empresaId);
      if (isAllowedFrameHref(href)) {
        return href;
      }
    }
    var base = new URL("/administrar_empresa/administrar_productos_menu.html", window.location.origin);
    if (empresaId) {
      base.searchParams.set("empresa_id", empresaId);
    }
    return base.pathname + base.search;
  }

  function withEmpresaParam(href, empresaId) {
    var normalized = normalizeHref(href);
    if (!normalized) return "";
    try {
      var url = new URL(normalized, window.location.origin);
      if (empresaId) {
        url.searchParams.set("empresa_id", empresaId);
      }
      return url.pathname + url.search;
    } catch (e) {
      return "";
    }
  }

  function favoritesStorageKey(empresaId) {
    return "admin_empresa:favorites:" + String(empresaId || "global");
  }

  function readFavorites(empresaId) {
    try {
      var raw = window.localStorage.getItem(favoritesStorageKey(empresaId)) || "[]";
      var parsed = JSON.parse(raw);
      return Array.isArray(parsed) ? parsed.filter(function (item) {
        return item && isAllowedFrameHref(item.href);
      }) : [];
    } catch (e) {
      return [];
    }
  }

  function writeFavorites(empresaId, favorites) {
    try {
      window.localStorage.setItem(favoritesStorageKey(empresaId), JSON.stringify(favorites.slice(0, 24)));
    } catch (e) {}
  }

  function stripEmpresaParam(href) {
    var normalized = normalizeHref(href);
    if (!normalized) return "";
    try {
      var url = new URL(normalized, window.location.origin);
      url.searchParams.delete("empresa_id");
      url.searchParams.delete("id");
      return url.pathname + url.search;
    } catch (e) {
      return normalized.split("?")[0];
    }
  }

  function menuMatchHref(href) {
    var normalized = normalizeHref(href);
    if (!normalized) return "";
    try {
      var url = new URL(normalized, window.location.origin);
      url.searchParams.delete("empresa_id");
      url.searchParams.delete("id");
      var params = [];
      url.searchParams.forEach(function (value, key) {
        params.push([key, value]);
      });
      params.sort(function (a, b) {
        if (a[0] === b[0]) return a[1] < b[1] ? -1 : (a[1] > b[1] ? 1 : 0);
        return a[0] < b[0] ? -1 : 1;
      });
      var query = params.map(function (pair) {
        return encodeURIComponent(pair[0]) + "=" + encodeURIComponent(pair[1]);
      }).join("&");
      return url.pathname + (query ? "?" + query : "");
    } catch (e) {
      return normalized;
    }
  }

  function getCurrentFrameHref() {
    if (!frame) return "";
    try {
      return normalizeHref(frame.contentWindow.location.pathname + frame.contentWindow.location.search);
    } catch (e) {
      return normalizeHref(frame.getAttribute("src") || frame.src || "");
    }
  }

  function findMenuLinkByHref(href) {
    var current = stripEmpresaParam(href);
    var activeLinks = frameLinks.length > 0 ? frameLinks : getFrameLinks();
    for (var i = 0; i < activeLinks.length; i += 1) {
      var link = activeLinks[i];
      if (!link) continue;
      if (stripEmpresaParam(link.getAttribute("href")) === current) {
        return link;
      }
    }
    return null;
  }

  function favoriteTitleFromFrame(href) {
    var link = findMenuLinkByHref(href);
    if (link) {
      return String(link.textContent || "").replace(/\s+/g, " ").trim();
    }
    try {
      var doc = frame && frame.contentDocument ? frame.contentDocument : null;
      var heading = doc ? doc.querySelector("h1,h2,.page-title") : null;
      var titleText = heading ? String(heading.textContent || "").trim() : "";
      if (titleText) return titleText;
      if (doc && doc.title) return String(doc.title).trim();
    } catch (e) {}
    try {
      var url = new URL(href, window.location.origin);
      var name = url.pathname.split("/").pop().replace(/\.html?$/i, "").replace(/[_-]+/g, " ");
      return name ? name.charAt(0).toUpperCase() + name.slice(1) : "Pagina";
    } catch (e) {
      return "Pagina";
    }
  }

  function favoriteIconFromMenu(href) {
    var link = findMenuLinkByHref(href);
    if (link) {
      var img = link.querySelector("img.icon");
      if (img && img.getAttribute("src")) {
        return { type: "img", src: img.getAttribute("src") };
      }
      var symbol = link.querySelector(".menu-symbol-icon,.icon");
      if (symbol && String(symbol.textContent || "").trim()) {
        return { type: "symbol", symbol: String(symbol.textContent || "").trim() };
      }
    }
    return { type: "img", src: "/img/analytics-color.svg" };
  }

  function isFavoriteHref(href, empresaId) {
    var current = stripEmpresaParam(href);
    if (!current) return false;
    return readFavorites(empresaId).some(function (item) {
      return stripEmpresaParam(item.href) === current;
    });
  }

  function notifyFavoritesChanged(empresaId) {
    try {
      window.dispatchEvent(new CustomEvent("pcs-admin-favorites-changed", { detail: { empresa_id: empresaId } }));
    } catch (e) {}
    try {
      if (frame && frame.contentWindow) {
        frame.contentWindow.postMessage({ type: "pcs-admin-favorites-changed", empresa_id: empresaId }, window.location.origin);
      }
    } catch (e) {}
  }

  function updateFavoriteButton(href) {
    if (!favoriteBtn) return;
    var currentHref = normalizeHref(href || getCurrentFrameHref());
    var allowed = isAllowedFrameHref(currentHref);
    favoriteBtn.hidden = !allowed;
    if (!allowed) return;
    var active = isFavoriteHref(currentHref, id);
    favoriteBtn.setAttribute("aria-pressed", active ? "true" : "false");
    favoriteBtn.setAttribute("title", active ? "Quitar de favoritos" : "Agregar a favoritos");
    favoriteBtn.setAttribute("aria-label", active ? "Quitar pagina de favoritos" : "Agregar pagina a favoritos");
  }

  function toggleCurrentFavorite() {
    if (!favoriteBtn) return;
    var currentHref = getCurrentFrameHref();
    if (!isAllowedFrameHref(currentHref)) return;
    var href = withEmpresaParam(currentHref, id) || currentHref;
    var currentKey = stripEmpresaParam(href);
    var favorites = readFavorites(id);
    var existingIndex = -1;
    for (var i = 0; i < favorites.length; i += 1) {
      if (stripEmpresaParam(favorites[i].href) === currentKey) {
        existingIndex = i;
        break;
      }
    }
    if (existingIndex >= 0) {
      favorites.splice(existingIndex, 1);
    } else {
      favorites.unshift({
        href: href,
        title: favoriteTitleFromFrame(href),
        icon: favoriteIconFromMenu(href),
        updatedAt: new Date().toISOString()
      });
    }
    writeFavorites(id, favorites);
    updateFavoriteButton(href);
    notifyFavoritesChanged(id);
  }

  function buildPortalUsuariosURL(empresaId, config) {
  var fallback = "/login_usuario.html";
  if (empresaId) {
    fallback += "?empresa_id=" + encodeURIComponent(String(empresaId));
  }
  var cfg = config || {};
  var targetEmpresaId = Number(empresaId || 0);
  var customDomain = String(cfg.dominio_publico || "").trim();
  if (customDomain) {
    try {
    if (customDomain.indexOf("://") === -1) {
      customDomain = "https://" + customDomain;
    }
    var customURL = new URL(customDomain);
    customURL.pathname = "/login_usuario.html";
    customURL.search = "";
    if (targetEmpresaId > 0) {
      customURL.searchParams.set("empresa_id", String(targetEmpresaId));
    }
    return customURL.toString();
    } catch (e) {
    return fallback;
    }
  }
  var slug = String(cfg.empresa_slug || "").trim().toLowerCase();
  if (!slug) return fallback;
  try {
    var url = new URL(window.location.origin);
    var host = String(url.hostname || "").toLowerCase();
    if (host === "powerfulcontrolsystem.com" || host === "www.powerfulcontrolsystem.com" || host.endsWith(".powerfulcontrolsystem.com")) {
    url.protocol = "https:";
    url.hostname = slug + ".powerfulcontrolsystem.com";
    url.pathname = "/login_usuario.html";
    url.search = "";
    if (targetEmpresaId > 0) {
      url.searchParams.set("empresa_id", String(targetEmpresaId));
    }
    return url.toString();
    }
  } catch (e) {
    return fallback;
  }
  return fallback;
  }

  async function resolvePortalUsuariosURL(empresaId) {
  var fallback = buildPortalUsuariosURL(empresaId, null);
  if (!empresaId) return fallback;
  try {
    var res = await fetch("/api/empresa/venta_publica?empresa_id=" + encodeURIComponent(String(empresaId)) + "&action=config", { credentials: "same-origin" });
    if (!res.ok) return fallback;
    var body = await res.json();
    return buildPortalUsuariosURL(empresaId, body && body.config ? body.config : null);
  } catch (e) {
    return fallback;
  }
  }

  function persistFrameSrc(href, empresaId) {
    if (!storage) return;
    var normalized = withEmpresaParam(href, empresaId);
    if (!isAllowedFrameHref(normalized)) return;
    try {
      storage.setItem(storageKey(empresaId), normalized);
    } catch (e) {}
  }

  function getStoredFrameSrc(empresaId) {
    if (!storage) return "";
    try {
      var raw = storage.getItem(storageKey(empresaId)) || "";
      var normalized = withEmpresaParam(raw, empresaId);
      if (!isAllowedFrameHref(normalized)) return "";
      return normalized;
    } catch (e) {
      return "";
    }
  }

  function clearActive() {
    frameLinks.forEach(function (link) {
      if (!link) return;
      link.classList.remove("active");
    });
  }

  function setActiveByHref(href) {
    var current = menuMatchHref(href);
    clearActive();
    frameLinks.forEach(function (link) {
      if (!link) return;
      var linkHref = menuMatchHref(link.getAttribute("href"));
      if (linkHref && linkHref === current) {
        link.classList.add("active");
        openMenuGroupForLink(link);
      }
    });
  }

  function setAdminNavGroupOpen(group, open) {
    if (!group) return;
    if (open && group.parentElement) {
      var siblings = Array.prototype.slice.call(group.parentElement.querySelectorAll(".admin-nav-group"));
      siblings.forEach(function (other) {
        if (other !== group) {
          other.classList.remove("is-open");
          var otherTitle = other.querySelector(".admin-nav-group-title");
          if (otherTitle) otherTitle.setAttribute("aria-expanded", "false");
        }
      });
    }
    group.classList.toggle("is-open", !!open);
    var title = group.querySelector(".admin-nav-group-title");
    if (title) {
      title.setAttribute("aria-expanded", open ? "true" : "false");
    }
  }

  function openMenuGroupForLink(link) {
    if (!link || typeof link.closest !== "function") return;
    var group = link.closest(".admin-nav-group");
    if (!group) return;
    setAdminNavGroupOpen(group, true);
  }

  function setupAdminNavGroups() {
    var groups = Array.prototype.slice.call(document.querySelectorAll(".admin-sidebar .admin-nav-group"));
    groups.forEach(function (group, index) {
      var title = group.querySelector(".admin-nav-group-title");
      if (!title) return;
      if (title.tagName && title.tagName.toLowerCase() !== "button") {
        title.setAttribute("role", "button");
        title.setAttribute("tabindex", "0");
      }
      var defaultOpen = group.classList.contains("is-open") || index === 0;
      setAdminNavGroupOpen(group, defaultOpen);
      var toggle = function () {
        setAdminNavGroupOpen(group, !group.classList.contains("is-open"));
      };
      title.addEventListener("click", toggle);
      title.addEventListener("keydown", function (event) {
        if (event.key === "Enter" || event.key === " ") {
          event.preventDefault();
          toggle();
        }
      });
    });
  }

  if (portalUsuariosLink) {
  resolvePortalUsuariosURL(id).then(function (url) {
    portalUsuariosLink.href = url;
  }).catch(function () {
    portalUsuariosLink.href = buildPortalUsuariosURL(id, null);
  });
  portalUsuariosLink.addEventListener("click", function (event) {
    event.preventDefault();
    resolvePortalUsuariosURL(id).then(function (url) {
    portalUsuariosLink.href = url;
    window.location.href = url;
    }).catch(function () {
    window.location.href = buildPortalUsuariosURL(id, null);
    });
  });
  }

  function normalizePermissionRole(raw) {
    var value = String(raw || "").trim().toLowerCase();
    switch (value) {
      case "super_administrador":
      case "superadmin":
      case "super":
        return "super_administrador";
      case "administrador":
      case "admin":
      case "admin_empresa":
        return "admin_empresa";
      case "supervisor":
      case "supervisor_sucursal":
        return "supervisor_sucursal";
      case "cajero":
        return "cajero";
      case "vendedor":
      case "ventas":
        return "vendedor";
      case "recepcion":
      case "recepcionista":
        return "recepcion";
      case "portero":
      case "porter":
      case "guardia":
      case "porteria":
      case "vigilante":
        return "portero";
      case "servicio_limpieza":
      case "servicio de limpieza":
      case "limpieza":
      case "aseadora":
      case "aseo":
      case "housekeeping":
        return "servicio_limpieza";
      case "inventario":
        return "inventario";
      case "jefe_bodega":
      case "jefe de bodega":
      case "bodega":
      case "bodeguero":
      case "almacenista":
        return "jefe_bodega";
      case "compras":
        return "compras";
      case "recursos_humanos":
      case "rrhh":
      case "talento_humano":
      case "talento humano":
        return "recursos_humanos";
      case "tecnico_solar":
      case "tecnico solar":
      case "técnico solar":
      case "solar":
        return "tecnico_solar";
      case "contabilidad":
        return "contabilidad";
      case "contador":
        return "contador";
      case "empresario":
      case "dueno":
      case "dueño":
      case "propietario":
      case "gerente_propietario":
        return "empresario";
      case "auditor":
        return "auditor";
      default:
        return value;
    }
  }

  function normalizePermissionAction(raw) {
    var value = String(raw || "").trim().toUpperCase();
    if (!value) return permActionRead;
    return value;
  }

  function roleIn(role, allowedRoles) {
    var normalized = String(role || "").trim().toLowerCase();
    if (!normalized) return false;
    for (var i = 0; i < allowedRoles.length; i += 1) {
      if (normalized === String(allowedRoles[i] || "").trim().toLowerCase()) {
        return true;
      }
    }
    return false;
  }

  function roleAllowsModuleAction(role, module, action) {
    var normalizedRole = normalizePermissionRole(role);
    var normalizedModule = String(module || "").trim().toLowerCase();
    var normalizedAction = normalizePermissionAction(action);
    var allReadRoles = ["admin_empresa", "supervisor_sucursal", "cajero", "inventario", "compras", "contabilidad", "auditor"];

    if (normalizedRole === "super_administrador" || normalizedRole === "administrador_total") {
      return true;
    }

    if (isNuevoVerticalModule(normalizedModule)) {
      if (normalizedAction === permActionRead) return roleIn(normalizedRole, allReadRoles);
      if (normalizedAction === permActionCreate || normalizedAction === permActionUpdate || normalizedAction === permActionApprove) {
        return roleIn(normalizedRole, ["admin_empresa", "supervisor_sucursal", "cajero"]);
      }
      if (normalizedAction === "D") {
        return roleIn(normalizedRole, ["admin_empresa", "supervisor_sucursal"]);
      }
    }

    switch (normalizedModule) {
      case permModuleVentas:
        if (normalizedAction === permActionRead) return roleIn(normalizedRole, allReadRoles.concat(["vendedor", "recepcion", "portero", "servicio_limpieza"]));
        if (normalizedAction === permActionApprove) {
          return roleIn(normalizedRole, ["admin_empresa", "supervisor_sucursal", "cajero", "portero"]);
        }
        if (normalizedAction === permActionCreate || normalizedAction === permActionUpdate || normalizedAction === "D") {
          return roleIn(normalizedRole, ["admin_empresa", "supervisor_sucursal", "cajero", "vendedor", "recepcion"]);
        }
        break;

      case permModuleInventario:
        if (normalizedAction === permActionRead) return roleIn(normalizedRole, allReadRoles.concat(["vendedor", "recepcion", "jefe_bodega"]));
        if (normalizedAction === permActionCreate || normalizedAction === permActionUpdate || normalizedAction === permActionApprove) {
          return roleIn(normalizedRole, ["admin_empresa", "supervisor_sucursal", "inventario", "jefe_bodega"]);
        }
        if (normalizedAction === "D") {
          return roleIn(normalizedRole, ["admin_empresa", "supervisor_sucursal", "inventario"]);
        }
        break;

      case permModuleCompras:
        if (normalizedAction === permActionRead) return roleIn(normalizedRole, allReadRoles.concat(["jefe_bodega"]));
        if (normalizedAction === permActionCreate || normalizedAction === permActionUpdate || normalizedAction === "D" || normalizedAction === permActionApprove) {
          return roleIn(normalizedRole, ["admin_empresa", "supervisor_sucursal", "compras"]);
        }
        break;

      case permModuleSoportesComprasIA:
        if (normalizedAction === permActionRead) return roleIn(normalizedRole, allReadRoles);
        if (normalizedAction === permActionCreate || normalizedAction === permActionUpdate || normalizedAction === permActionApprove) {
          return roleIn(normalizedRole, ["admin_empresa", "supervisor_sucursal", "compras", "contabilidad"]);
        }
        if (normalizedAction === "D") {
          return false;
        }
        break;

      case permModuleFinanzas:
      case permModuleContabilidadCO:
      case permModuleContabilidadCOAv:
      case permModuleCentrosCosto:
      case permModuleCierreFiscal:
      case permModuleActivosFijosNIIF:
      case permModuleDeclaracionesTrib:
      case permModuleTesoreria:
      case permModuleNominaSueldos:
      case permModuleCobranza:
      case permModulePortalContador:
      case permModulePortalTerceros:
        if (normalizedAction === permActionRead) {
          if (normalizedModule === permModuleFinanzas) return roleIn(normalizedRole, allReadRoles.concat(["contador"]));
          if (normalizedModule === permModuleNominaSueldos) return roleIn(normalizedRole, allReadRoles.concat(["recursos_humanos"]));
          return roleIn(normalizedRole, allReadRoles);
        }
        if (normalizedAction === permActionCreate || normalizedAction === permActionUpdate || normalizedAction === permActionApprove) {
          if (normalizedModule === permModuleNominaSueldos) {
            if (normalizedAction === permActionApprove) return roleIn(normalizedRole, ["admin_empresa", "contabilidad"]);
            return roleIn(normalizedRole, ["admin_empresa", "contabilidad", "recursos_humanos"]);
          }
          return roleIn(normalizedRole, ["admin_empresa", "contabilidad"]);
        }
        if (normalizedAction === "D") {
          return roleIn(normalizedRole, ["contabilidad"]);
        }
        break;

      case permModuleBancosPagos:
      case permModuleGestionDocumental:
      case permModuleCumplimientoKYC:
      case permModuleContratosObligaciones:
      case permModuleCalidadProcesos:
      case permModuleEnergiaSolar:
      case permModuleAuditoria:
      case permModuleBackups:
      case permModuleDocumentosOnlyOffice:
        if (normalizedAction === permActionRead) {
          if (normalizedModule === permModuleEnergiaSolar) return roleIn(normalizedRole, allReadRoles.concat(["tecnico_solar"]));
          return roleIn(normalizedRole, allReadRoles);
        }
        if (normalizedAction === permActionCreate || normalizedAction === permActionUpdate || normalizedAction === permActionApprove) {
          return roleIn(normalizedRole, ["admin_empresa", "supervisor_sucursal", "contabilidad", "auditor"]);
        }
        if (normalizedAction === "D") {
          return roleIn(normalizedRole, ["admin_empresa"]);
        }
        break;

      case permModuleClientes:
      case permModuleCRMUnificado:
        if (normalizedAction === permActionRead) return roleIn(normalizedRole, allReadRoles.concat(["vendedor", "recepcion"]));
        if (normalizedAction === permActionCreate || normalizedAction === permActionUpdate || normalizedAction === permActionApprove) {
          return roleIn(normalizedRole, ["admin_empresa", "supervisor_sucursal", "cajero", "vendedor", "recepcion"]);
        }
        if (normalizedAction === "D") {
          return false;
        }
        break;

      case permModuleFacturacion:
      case permModuleFacturacionEcuador:
      case permModuleFacturacionPanama:
        if (normalizedAction === permActionRead) {
          if (normalizedModule === permModuleFacturacion) return roleIn(normalizedRole, allReadRoles.concat(["contador"]));
          return roleIn(normalizedRole, allReadRoles);
        }
        if (normalizedAction === permActionCreate || normalizedAction === permActionUpdate || normalizedAction === permActionApprove) {
          return roleIn(normalizedRole, ["admin_empresa", "cajero"]);
        }
        if (normalizedAction === "D") {
          return false;
        }
        break;

      case permModuleAIUConstruccion:
        if (normalizedAction === permActionRead) return roleIn(normalizedRole, allReadRoles);
        if (normalizedAction === permActionCreate || normalizedAction === permActionUpdate || normalizedAction === permActionApprove) {
          return roleIn(normalizedRole, ["admin_empresa", "contabilidad", "supervisor_sucursal"]);
        }
        if (normalizedAction === "D") {
          return roleIn(normalizedRole, ["admin_empresa", "contabilidad"]);
        }
        break;

      case permModuleSeguridad:
        if (normalizedAction === permActionRead) return roleIn(normalizedRole, allReadRoles);
        if (normalizedAction === permActionCreate || normalizedAction === permActionUpdate || normalizedAction === "D" || normalizedAction === permActionApprove) {
          return roleIn(normalizedRole, ["admin_empresa"]);
        }
        break;

      case permModuleVentaPublica:
      case permModuleReservasHotel:
      case permModuleChatTareas:
      case permModuleGimnasio:
      case permModuleTaxiSystem:
      case permModuleDomicilios:
      case permModuleParqueadero:
      case permModuleApartTuristicos:
      case permModulePropiedadHorizontal:
      case permModuleAlquileres:
      case permModuleOdontologia:
      case permModuleDrogueriaFarmacia:
      case permModuleTurnos:
      case permModuleCarnets:
        if (normalizedAction === permActionRead) return roleIn(normalizedRole, allReadRoles);
        if (normalizedAction === permActionCreate || normalizedAction === permActionUpdate || normalizedAction === permActionApprove) {
          return roleIn(normalizedRole, ["admin_empresa", "supervisor_sucursal", "cajero"]);
        }
        if (normalizedAction === "D") {
          return roleIn(normalizedRole, ["admin_empresa", "supervisor_sucursal"]);
        }
        break;

      case permModuleProduccionMRP:
      case permModuleLogisticaWMS:
      case permModuleImportaciones:
        if (normalizedAction === permActionRead) return roleIn(normalizedRole, allReadRoles);
        if (normalizedAction === permActionCreate || normalizedAction === permActionUpdate || normalizedAction === "D" || normalizedAction === permActionApprove) {
          return roleIn(normalizedRole, ["admin_empresa", "supervisor_sucursal", "inventario", "compras"]);
        }
        break;

      case permModuleHorariosTrab:
      case permModuleAsistenciaEmpleados:
      case permModuleVehiculosRegistro:
      case permModuleHojaVidaOperativa:
      case permModuleUbicacionGPS:
        if (normalizedAction === permActionRead) {
          if (normalizedModule === permModuleHorariosTrab || normalizedModule === permModuleAsistenciaEmpleados) return roleIn(normalizedRole, allReadRoles.concat(["recursos_humanos"]));
          return roleIn(normalizedRole, allReadRoles);
        }
        if (normalizedAction === permActionCreate || normalizedAction === permActionUpdate) {
          if (normalizedModule === permModuleHorariosTrab || normalizedModule === permModuleAsistenciaEmpleados) return roleIn(normalizedRole, ["admin_empresa", "supervisor_sucursal", "recursos_humanos"]);
          return roleIn(normalizedRole, ["admin_empresa", "supervisor_sucursal"]);
        }
        if (normalizedAction === "D" || normalizedAction === permActionApprove) {
          return roleIn(normalizedRole, ["admin_empresa"]);
        }
        break;

      case permModuleReportes:
        if (normalizedAction === permActionRead) return roleIn(normalizedRole, allReadRoles.concat(["empresario"]));
        if (normalizedAction === permActionCreate || normalizedAction === permActionUpdate || normalizedAction === permActionApprove) {
          return roleIn(normalizedRole, ["admin_empresa", "supervisor_sucursal", "contabilidad", "auditor"]);
        }
        if (normalizedAction === "D") {
          return roleIn(normalizedRole, ["admin_empresa"]);
        }
        break;

      case permModuleControlElectrico:
        if (normalizedAction === permActionRead) return roleIn(normalizedRole, allReadRoles);
        if (normalizedAction === permActionCreate || normalizedAction === permActionUpdate) {
          return roleIn(normalizedRole, ["admin_empresa", "supervisor_sucursal"]);
        }
        if (normalizedAction === "D" || normalizedAction === permActionApprove) {
          return roleIn(normalizedRole, ["admin_empresa"]);
        }
        break;
    }

    return false;
  }

  function setMenuPermissionsEvidence(text, isFallback) {
    if (!permsEvidence) return;
    permsEvidence.textContent = text || "";
    if (isFallback) {
      permsEvidence.style.opacity = "0.85";
      return;
    }
    permsEvidence.style.opacity = "1";
  }

  function getPermissionContextModuleRow(permissionContext, moduleName) {
    if (!permissionContext || !Array.isArray(permissionContext.modulos)) {
      return null;
    }
    var target = String(moduleName || "").trim().toLowerCase();
    if (!target) {
      return null;
    }
    for (var i = 0; i < permissionContext.modulos.length; i += 1) {
      var row = permissionContext.modulos[i];
      var rowModule = String(row && row.modulo || "").trim().toLowerCase();
      if (rowModule && rowModule === target) {
        return row;
      }
    }
    return null;
  }

  function boolFromActionMap(actionMap, actionKey) {
    if (!actionMap || typeof actionMap !== "object") {
      return false;
    }
    if (Object.prototype.hasOwnProperty.call(actionMap, actionKey)) {
      return !!actionMap[actionKey];
    }
    var lowerKey = String(actionKey || "").toLowerCase();
    if (Object.prototype.hasOwnProperty.call(actionMap, lowerKey)) {
      return !!actionMap[lowerKey];
    }
    return false;
  }

  function isContextModuleActionAllowed(moduleRow, action) {
    if (!moduleRow || typeof moduleRow !== "object") {
      return false;
    }
    var actionKey = normalizePermissionAction(action);
    if (actionKey === permActionRead && typeof moduleRow.read !== "undefined") {
      return !!moduleRow.read;
    }
    if (actionKey === permActionCreate && typeof moduleRow.create !== "undefined") {
      return !!moduleRow.create;
    }
    if (actionKey === permActionUpdate && typeof moduleRow.update !== "undefined") {
      return !!moduleRow.update;
    }
    if (actionKey === "D" && typeof moduleRow.delete !== "undefined") {
      return !!moduleRow.delete;
    }
    if (actionKey === permActionApprove && typeof moduleRow.approve !== "undefined") {
      return !!moduleRow.approve;
    }
    return boolFromActionMap(moduleRow.acciones, actionKey);
  }

  function canPermissionContextAccessLink(permissionContext, link) {
    if (!link) return false;
    var rule = menuPermissionCatalog[link.id || ""];
    if (!rule || rule.alwaysVisible) {
      return true;
    }
    var pageKey = link.id || "";
    var pages = permissionContext && permissionContext.paginas;
    if (pageKey && pages && typeof pages === "object" && Object.prototype.hasOwnProperty.call(pages, pageKey)) {
      if (pages[pageKey]) return true;
      if (!Array.isArray(rule.anyModules) || !rule.anyModules.length) return false;
    }
    if (Array.isArray(rule.anyModules) && rule.anyModules.length) {
      return rule.anyModules.some(function (module) {
        var moduleRow = getPermissionContextModuleRow(permissionContext, module);
        return isContextModuleActionAllowed(moduleRow, rule.action);
      });
    }
    var moduleRow = getPermissionContextModuleRow(permissionContext, rule.module);
    return isContextModuleActionAllowed(moduleRow, rule.action);
  }

  function applyMenuPermissionsByContext(permissionContext) {
    links.forEach(function (link) {
      setMenuLinkVisible(link, true);
    });
    if (!permissionContext) {
      setSecondaryMenuVisibility(true);
      refreshMenuGroups();
      return;
    }
    if (normalizePermissionRole(permissionContext.rol || permissionContext.role || "") === "portero") {
      links.forEach(function (link) {
        setMenuLinkVisible(link, !!link && link.id === "linkEstaciones");
      });
      setSecondaryMenuVisibility(false);
      refreshMenuGroups();
      return;
    }
    if (normalizePermissionRole(permissionContext.rol || permissionContext.role || "") === "servicio_limpieza") {
      links.forEach(function (link) {
        setMenuLinkVisible(link, !!link && link.id === "linkEstaciones");
      });
      setSecondaryMenuVisibility(false);
      refreshMenuGroups();
      return;
    }
    if (normalizePermissionRole(permissionContext.rol || permissionContext.role || "") === "contador") {
      links.forEach(function (link) {
        setMenuLinkVisible(link, !!link && ["linkFinanzas", "linkFinanzasMain", "linkImpuestos"].indexOf(link.id) !== -1);
      });
      setSecondaryMenuVisibility(false);
      refreshMenuGroups();
      return;
    }
    if (normalizePermissionRole(permissionContext.rol || permissionContext.role || "") === "empresario") {
      links.forEach(function (link) {
        setMenuLinkVisible(link, !!link && ["linkReportes", "linkReportesEjecutivos"].indexOf(link.id) !== -1);
      });
      setSecondaryMenuVisibility(false);
      refreshMenuGroups();
      return;
    }
    if (normalizePermissionRole(permissionContext.rol || permissionContext.role || "") === "tecnico_solar") {
      links.forEach(function (link) {
        setMenuLinkVisible(link, !!link && link.id === "linkEnergiaSolar");
      });
      setSecondaryMenuVisibility(false);
      refreshMenuGroups();
      return;
    }
    if (normalizePermissionRole(permissionContext.rol || permissionContext.role || "") === "jefe_bodega") {
      var bodegaLinks = ["linkProductos", "linkProductosMain", "linkInventarioAvanzado", "linkRecetasProductos", "linkPreciosHistorial", "linkBodegas", "linkCategorias", "linkGeneradorCodigosBarras"];
      links.forEach(function (link) {
        setMenuLinkVisible(link, !!link && bodegaLinks.indexOf(link.id) !== -1);
      });
      setSecondaryMenuVisibility(false);
      refreshMenuGroups();
      return;
    }
    if (normalizePermissionRole(permissionContext.rol || permissionContext.role || "") === "recursos_humanos") {
      var rrhhLinks = ["linkHorariosTrabajadores", "linkAsistenciaEmpleados", "linkNominaSueldos", "linkNominaMenu", "linkMiHorario"];
      links.forEach(function (link) {
        setMenuLinkVisible(link, !!link && rrhhLinks.indexOf(link.id) !== -1);
      });
      setSecondaryMenuVisibility(false);
      refreshMenuGroups();
      return;
    }
    links.forEach(function (link) {
      setMenuLinkVisible(link, canPermissionContextAccessLink(permissionContext, link));
    });
    setSecondaryMenuVisibility(shouldShowSecondaryMenuLinks(permissionContext));
    refreshMenuGroups();
  }

  if (favoriteBtn) {
    favoriteBtn.addEventListener("click", function (ev) {
      ev.preventDefault();
      toggleCurrentFavorite();
    });
    updateFavoriteButton("");
  }

  function describePermissionContext(permissionContext) {
    if (!permissionContext || typeof permissionContext !== "object") {
      return "Permisos de menú: sin contexto disponible.";
    }
    var role = normalizePermissionRole(permissionContext.rol || "sin_rol") || "sin_rol";
    var summary = permissionContext.resumen || {};
    var modulesTotal = Number(summary.modulos_total || 0);
    var modulesRead = Number(summary.modulos_lectura || 0);
    var modulesApprove = Number(summary.modulos_aprobacion || 0);
    var enabledActions = Number(summary.acciones_habilitadas || 0);
    return "Permisos de menú: rol " + role +
      " | lectura " + modulesRead + "/" + modulesTotal +
      " | aprobación " + modulesApprove +
      " | acciones habilitadas " + enabledActions +
      " | fuente: /api/empresa/permisos_contexto";
  }

  function fetchEmpresaPermisosContexto(empresaId) {
    if (!empresaId) {
      return Promise.resolve(null);
    }
    var url = "/api/empresa/permisos_contexto?empresa_id=" + encodeURIComponent(empresaId);
    return fetch(url, { credentials: "same-origin" })
      .then(function (resp) {
        if (!resp.ok) return null;
        return resp.json();
      })
      .then(function (data) {
        if (!data || typeof data !== "object") return null;
        if (!Array.isArray(data.modulos)) return null;
        return data;
      })
      .catch(function () {
        return null;
      });
  }

  function fetchVerticalIntegrationCatalog(empresaId) {
    if (!verticalIntegration || typeof verticalIntegration.applyCatalogItems !== "function") {
      return Promise.resolve(null);
    }
    var url = empresaId
      ? "/api/empresa/plantillas_integracion/catalogo?empresa_id=" + encodeURIComponent(empresaId)
      : "/api/public/plantillas_integracion/catalogo";
    return fetch(url, { credentials: "same-origin" })
      .then(function (resp) {
        if (!resp.ok) return null;
        return resp.json();
      })
      .then(function (payload) {
        var items = payload && Array.isArray(payload.items) ? payload.items : [];
        if (!items.length) return null;
        return { total: verticalIntegration.applyCatalogItems(items), source: url };
      })
      .catch(function () {
        return null;
      });
  }

  function setVerticalIntegrationEvidence(result) {
    if (!verticalIntegrationEvidence || !verticalIntegration || typeof verticalIntegration.summary !== "function") {
      return;
    }
    var s = verticalIntegration.summary();
    if (!s || !s.total) {
      verticalIntegrationEvidence.hidden = true;
      return;
    }
    var source = result && result.source ? "API" : "local";
    var parts = ["Plantillas", String(s.visible) + "/" + String(s.total), source];
    if (s.hidden) parts.push(String(s.hidden) + " ocultos");
    if (s.pending) parts.push(String(s.pending) + " pendientes");
    verticalIntegrationEvidence.textContent = parts.join(" · ");
    verticalIntegrationEvidence.hidden = false;
    verticalIntegrationEvidence.setAttribute("data-source", source.toLowerCase());
  }

  function setMenuLinkVisible(link, visible) {
    if (!link) return;
    visible = !!visible && menuLinkPassesVerticalIntegration(link);
    var item = null;
    if (typeof link.closest === "function") {
      item = link.closest("li");
    }
    if (!item) {
      item = link.parentElement;
    }
    if (item) {
      item.style.display = visible ? "" : "none";
    }
    link.setAttribute("data-menu-visible", visible ? "1" : "0");
    if (!visible) {
      link.classList.remove("active");
    }
  }

  function setSecondaryMenuVisibility(visible) {
    if (portalUsuariosLink) {
      var portalItem = typeof portalUsuariosLink.closest === "function"
        ? portalUsuariosLink.closest("li")
        : portalUsuariosLink.parentElement;
      if (portalItem) {
        portalItem.style.display = visible ? "" : "none";
      }
    }
    if (companySelectorLink) {
      var companyItem = typeof companySelectorLink.closest === "function"
        ? companySelectorLink.closest("li")
        : companySelectorLink.parentElement;
      if (companyItem) {
        companyItem.style.display = visible ? "" : "none";
      }
    }
  }

  function shouldShowSecondaryMenuLinks(permissionContext) {
    var pages = permissionContext && permissionContext.paginas;
    if (!pages || typeof pages !== "object") {
      return true;
    }
    var allowedCount = 0;
    for (var key in pages) {
      if (!Object.prototype.hasOwnProperty.call(pages, key)) continue;
      if (pages[key]) {
        allowedCount += 1;
      }
      if (allowedCount > 1) {
        return true;
      }
    }
    return allowedCount !== 1;
  }

  function refreshMenuGroups() {
    refreshNuevasPlantillasMenuVisibility();
    var groups = Array.prototype.slice.call(document.querySelectorAll(".admin-sidebar .admin-nav-group"));
    groups.forEach(function (group) {
      var items = Array.prototype.slice.call(group.querySelectorAll(".admin-nav-sublist > li"));
      if (items.length === 0) {
        group.style.display = "";
        return;
      }
      var hasVisibleItem = items.some(function (item) {
        return item.style.display !== "none" && item.hidden !== true;
      });
      group.style.display = hasVisibleItem ? "" : "none";
    });
  }

  function refreshNuevasPlantillasMenuVisibility() {
    var mount = document.getElementById("adminBusinessVerticalsMount");
    if (mount) mount.style.display = "none";
  }

  function isMenuLinkVisible(link) {
    if (!link) return false;
    return link.getAttribute("data-menu-visible") !== "0";
  }

  function canRoleAccessLink(role, link) {
    if (!link) return false;
    var rule = menuPermissionCatalog[link.id || ""];
    if (!rule || rule.alwaysVisible) {
      return true;
    }
    if (Array.isArray(rule.anyModules) && rule.anyModules.length) {
      return rule.anyModules.some(function (module) {
        return roleAllowsModuleAction(role, module, rule.action);
      });
    }
    return roleAllowsModuleAction(role, rule.module, rule.action);
  }

  function applyMenuPermissionsByRole(rawRole) {
    var normalizedRole = normalizePermissionRole(rawRole);
    links.forEach(function (link) {
      setMenuLinkVisible(link, true);
    });
    if (!normalizedRole) {
      setSecondaryMenuVisibility(true);
      refreshMenuGroups();
      return;
    }
    if (normalizedRole === "portero") {
      links.forEach(function (link) {
        setMenuLinkVisible(link, !!link && link.id === "linkEstaciones");
      });
      setSecondaryMenuVisibility(false);
      refreshMenuGroups();
      return;
    }
    if (normalizedRole === "servicio_limpieza") {
      links.forEach(function (link) {
        setMenuLinkVisible(link, !!link && link.id === "linkEstaciones");
      });
      setSecondaryMenuVisibility(false);
      refreshMenuGroups();
      return;
    }
    if (normalizedRole === "contador") {
      links.forEach(function (link) {
        setMenuLinkVisible(link, !!link && ["linkFinanzas", "linkFinanzasMain", "linkImpuestos"].indexOf(link.id) !== -1);
      });
      setSecondaryMenuVisibility(false);
      refreshMenuGroups();
      return;
    }
    if (normalizedRole === "empresario") {
      links.forEach(function (link) {
        setMenuLinkVisible(link, !!link && ["linkReportes", "linkReportesEjecutivos"].indexOf(link.id) !== -1);
      });
      setSecondaryMenuVisibility(false);
      refreshMenuGroups();
      return;
    }
    if (normalizedRole === "tecnico_solar") {
      links.forEach(function (link) {
        setMenuLinkVisible(link, !!link && link.id === "linkEnergiaSolar");
      });
      setSecondaryMenuVisibility(false);
      refreshMenuGroups();
      return;
    }
    if (normalizedRole === "jefe_bodega") {
      var bodegaLinks = ["linkProductos", "linkProductosMain", "linkInventarioAvanzado", "linkRecetasProductos", "linkPreciosHistorial", "linkBodegas", "linkCategorias", "linkGeneradorCodigosBarras"];
      links.forEach(function (link) {
        setMenuLinkVisible(link, !!link && bodegaLinks.indexOf(link.id) !== -1);
      });
      setSecondaryMenuVisibility(false);
      refreshMenuGroups();
      return;
    }
    if (normalizedRole === "recursos_humanos") {
      var rrhhLinks = ["linkHorariosTrabajadores", "linkAsistenciaEmpleados", "linkNominaSueldos", "linkNominaMenu", "linkMiHorario"];
      links.forEach(function (link) {
        setMenuLinkVisible(link, !!link && rrhhLinks.indexOf(link.id) !== -1);
      });
      setSecondaryMenuVisibility(false);
      refreshMenuGroups();
      return;
    }
    links.forEach(function (link) {
      setMenuLinkVisible(link, canRoleAccessLink(normalizedRole, link));
    });
    setSecondaryMenuVisibility(true);
    refreshMenuGroups();
  }

  function isVisibleMenuHref(href) {
    var current = normalizeHref(href).split("?")[0];
    if (!current) return false;
    for (var i = 0; i < frameLinks.length; i += 1) {
      var link = frameLinks[i];
      if (!isMenuLinkVisible(link)) continue;
      var linkHref = normalizeHref(link.getAttribute("href")).split("?")[0];
      if (linkHref && linkHref === current) {
        return true;
      }
    }
    return false;
  }

  function firstVisibleFrameSrc(empresaId) {
    for (var i = 0; i < frameLinks.length; i += 1) {
      var link = frameLinks[i];
      if (!isMenuLinkVisible(link)) continue;
      var href = withEmpresaParam(link.getAttribute("href"), empresaId);
      if (isAllowedFrameHref(href)) {
        return href;
      }
    }
    return defaultFrameSrc(empresaId);
  }

  function preferredStartupFrameSrc(empresaId) {
    var panelLink = document.getElementById("linkPanelEmpresa");
    var href = panelLink
      ? withEmpresaParam(panelLink.getAttribute("href"), empresaId)
      : "";
    if (href && isAllowedFrameHref(href) && isVisibleMenuHref(href)) {
      return href;
    }
    return "";
  }

  function resolveInitialFrameSrc(empresaId) {
    var preferred = preferredStartupFrameSrc(empresaId);
    if (preferred) {
      return preferred;
    }
    var restored = getStoredFrameSrc(empresaId);
    if (restored && isVisibleMenuHref(restored)) {
      return restored;
    }
    return firstVisibleFrameSrc(empresaId);
  }

  function applyMenuPermissionsWithSource(empresaId, role) {
    var normalizedRole = normalizePermissionRole(role);
    return fetchEmpresaPermisosContexto(empresaId)
      .then(function (permissionContext) {
        if (permissionContext) {
          applyMenuPermissionsByContext(permissionContext);
          setMenuPermissionsEvidence(describePermissionContext(permissionContext), false);
          return;
        }
        applyMenuPermissionsByRole(normalizedRole);
        if (normalizedRole) {
          setMenuPermissionsEvidence("Permisos de menú: rol " + normalizedRole + " | fuente local de respaldo.", true);
        } else {
          setMenuPermissionsEvidence("Permisos de menú: sin rol detectado | fuente local de respaldo.", true);
        }
      });
  }

  function setLinksWithEmpresa(empresaId) {
    frameLinks.forEach(function (link) {
      if (!link) return;
      var href = link.getAttribute("href");
      if (!href) return;
      var target = new URL(href, window.location.origin);
      if (empresaId) {
        target.searchParams.set("empresa_id", empresaId);
      }
      link.setAttribute("href", target.pathname + target.search);

      link.addEventListener("click", function (ev) {
        ev.preventDefault();
        var linkHref = link.getAttribute("href");
        if (!frame || !linkHref) {
          window.location.href = linkHref;
          return;
        }
        frame.setAttribute("src", linkHref);
        persistFrameSrc(linkHref, empresaId);
        setActiveByHref(linkHref);
        updateFavoriteButton(linkHref);
      });
    });
  }

  function redirectToAdminLogin() {
    try {
      if (window.top && window.top !== window) {
        window.top.location.href = "/login.html";
        return;
      }
    } catch (error) {
      // fallback local
    }
    window.location.href = "/login.html";
  }

  function fetchCurrentAdminSession() {
    return fetch("/me", { credentials: "same-origin" })
      .then(function (resp) {
        if (resp.status === 401 || resp.status === 403) {
          return { authenticated: false, role: "" };
        }
        if (!resp.ok) {
          return { authenticated: null, role: "" };
        }
        return resp.json()
          .then(function (data) {
            return {
              authenticated: true,
              role: (!data || typeof data !== "object")
                ? ""
                : String(data.role || data.Role || "").trim()
            };
          })
          .catch(function () {
            return { authenticated: true, role: "" };
          });
      })
      .catch(function () {
        return { authenticated: null, role: "" };
      });
  }

  function loadEmpresaTitle(empresaId) {
    return fetch("/api/empresa/configuracion_guiada?empresa_id=" + encodeURIComponent(empresaId), { credentials: "same-origin" })
      .then(function (resp) {
        if (!resp.ok) {
          if (titleMenu) titleMenu.textContent = "Administrar Empresa";
          else if (title) title.textContent = "Administrar Empresa";
          throw new Error("empresa no encontrada");
        }
        return resp.json();
      })
      .then(function (data) {
          var estado = data && data.estado && typeof data.estado === "object" ? data.estado : data;
          var nombre = estado && (estado.empresa_nombre || estado.nombre || estado.Nombre);
          if (nombre) {
            if (titleMenu) titleMenu.textContent = "Administrar Empresa";
            if (empresaNameMenu) empresaNameMenu.textContent = String(nombre);
            // Keep the browser title including the company name for clarity
            document.title = "Administrar Empresa - " + nombre;
          } else {
            if (titleMenu) titleMenu.textContent = "Administrar Empresa";
            if (empresaNameMenu) empresaNameMenu.textContent = "";
          }
      })
      .catch(function (err) {
        console.warn("No se pudo cargar empresa:", err);
        if (titleMenu) titleMenu.textContent = "Administrar Empresa";
        else if (title) {
          var cur3 = String(title.textContent || "").trim();
          if (!cur3 || cur3.indexOf("Administrar Empresa -") === 0 || cur3 === "Administrar Empresa") {
            title.textContent = "Administrar Empresa";
          }
        }
      });
  }

  function initializeMenuAndFrame(empresaId) {
    frameLinks = getFrameLinks();
    setLinksWithEmpresa(empresaId);
    if (!frame) return;
    var initialSrc = resolveInitialFrameSrc(empresaId);
    frame.src = initialSrc;
    setActiveByHref(initialSrc);
  }

  function clearPendingConfigurationAssistant(empresaId) {
    if (!empresaId) return;
    var key = "pcs_config_assistant_pending_" + String(empresaId);
    try { window.sessionStorage.removeItem(key); } catch (e) {}
    try { window.localStorage.removeItem(key); } catch (e) {}
  }

  function isMobileAdminViewport() {
    try {
      return window.matchMedia && window.matchMedia("(max-width: 900px)").matches;
    } catch (e) {
      return window.innerWidth <= 900;
    }
  }

  function disconnectFrameResizeObserver() {
    if (frameResizeObserver && typeof frameResizeObserver.disconnect === "function") {
      try { frameResizeObserver.disconnect(); } catch (e) {}
    }
    frameResizeObserver = null;
  }

  function readFrameContentHeight() {
    if (!frame || !frame.contentDocument) return 0;
    var doc = frame.contentDocument;
    var body = doc.body;
    var root = doc.documentElement;
    if (!body || !root) return 0;
    return Math.max(
      body.scrollHeight || 0,
      body.offsetHeight || 0,
      root.scrollHeight || 0,
      root.offsetHeight || 0
    );
  }

  function resizeAdminFrameForMobile() {
    if (!frame) return;
    if (!isMobileAdminViewport()) {
      frame.style.height = "";
      frame.removeAttribute("data-mobile-auto-height");
      disconnectFrameResizeObserver();
      return;
    }
    try {
      var height = readFrameContentHeight();
      var minimum = Math.max(Math.round(window.innerHeight * 0.72), 520);
      if (height > 0) {
        frame.style.height = String(Math.max(height + 18, minimum)) + "px";
        frame.setAttribute("data-mobile-auto-height", "1");
      }
    } catch (e) {
      frame.style.height = "72vh";
      frame.setAttribute("data-mobile-auto-height", "1");
    }
  }

  function watchFrameContentForMobileResize() {
    disconnectFrameResizeObserver();
    if (!frame || !isMobileAdminViewport()) return;
    try {
      var doc = frame.contentDocument;
      if (!doc || !doc.body || typeof MutationObserver === "undefined") return;
      frameResizeObserver = new MutationObserver(function () {
        window.requestAnimationFrame(resizeAdminFrameForMobile);
      });
      frameResizeObserver.observe(doc.body, { childList: true, subtree: true, attributes: true });
    } catch (e) {
      frameResizeObserver = null;
    }
  }

  function scheduleMobileFrameResize() {
    resizeAdminFrameForMobile();
    watchFrameContentForMobileResize();
    [80, 260, 700, 1400].forEach(function (delay) {
      window.setTimeout(function () {
        resizeAdminFrameForMobile();
        watchFrameContentForMobileResize();
      }, delay);
    });
  }

  if (frame) {
    frame.addEventListener("load", function () {
      var currentHref = "";
      try {
        currentHref = frame.contentWindow.location.pathname + frame.contentWindow.location.search;
      } catch (e) {
        currentHref = frame.getAttribute("src") || "";
      }
      if (!currentHref) return;

      // Si una navegación interna del iframe pierde empresa_id,
      // se corrige automáticamente usando el contexto activo.
      if (id) {
        try {
          var normalizedCurrent = normalizeHref(currentHref);
          var currentURL = new URL(normalizedCurrent || currentHref, window.location.origin);
          var hasEmpresaID = parsePositiveInt(currentURL.searchParams.get("empresa_id")) > 0;
          if (!hasEmpresaID) {
            var correctedHref = withEmpresaParam(currentURL.pathname + currentURL.search, id);
            if (correctedHref && correctedHref !== normalizedCurrent) {
              frame.setAttribute("src", correctedHref);
              return;
            }
          }
        } catch (e) {
          // no-op
        }
      }

      persistFrameSrc(currentHref, id);
      setActiveByHref(currentHref);
      updateFavoriteButton(currentHref);
      scheduleMobileFrameResize();
    });
    // Interceptar F5 / Ctrl+R para recargar solo el iframe y mantener la subpágina activa.
    // Si el foco está en un campo editable (input/textarea/contentEditable) se respeta el comportamiento por defecto.
    document.addEventListener('keydown', function (ev) {
      try {
        var isF5 = ev.key === 'F5' || ev.keyCode === 116;
        var isCtrlR = (ev.ctrlKey || ev.metaKey) && (ev.key === 'r' || ev.keyCode === 82);
        if (!isF5 && !isCtrlR) return;

        var active = document.activeElement;
        var tag = (active && active.tagName) ? active.tagName.toLowerCase() : '';
        var isEditable = tag === 'input' || tag === 'textarea' || (active && active.isContentEditable);
        if (isEditable && !active.readOnly) {
          // permitir refresco normal cuando el usuario está editando
          return;
        }

        ev.preventDefault();
        if (frame && frame.contentWindow) {
          try {
            frame.contentWindow.location.reload();
            return;
          } catch (e) {
            // si por alguna razón no es posible acceder al contentWindow, forzamos reload asignando src
            try {
              var src = frame.getAttribute('src') || frame.src;
              frame.setAttribute('src', src);
              return;
            } catch (e2) {
              // fallback al reload global
            }
          }
        }
        window.location.reload();
      } catch (e) {
        // no-op
      }
    });
  }

  window.addEventListener("resize", function () {
    scheduleMobileFrameResize();
  });

  setupAdminNavGroups();

  fetchVerticalIntegrationCatalog(id)
    .then(function (integrationResult) {
      setVerticalIntegrationEvidence(integrationResult);
      return fetchCurrentAdminSession();
    })
    .then(function (session) {
      if (session && session.authenticated === false) {
        redirectToAdminLogin();
        return null;
      }
      var role = session && session.role ? session.role : "";
      if (id) {
        return applyMenuPermissionsWithSource(id, role)
          .then(function () {
            initializeMenuAndFrame(id);
            loadEmpresaTitle(id);
            clearPendingConfigurationAssistant(id);
          });
      }
      applyMenuPermissionsByRole(role);
      initializeMenuAndFrame("");
      if (title) {
        title.textContent = "Administrar Empresa";
      }
      return null;
    })
    .catch(function () {
      setVerticalIntegrationEvidence(null);
      if (id) {
        applyMenuPermissionsByRole("");
        setMenuPermissionsEvidence("Permisos de menú: no se pudo resolver contexto, se mantiene visibilidad base.", true);
        initializeMenuAndFrame(id);
        loadEmpresaTitle(id);
        clearPendingConfigurationAssistant(id);
        return;
      }
      applyMenuPermissionsByRole("");
      initializeMenuAndFrame("");
      if (title) {
        title.textContent = "Administrar Empresa";
      }
    });
})();
