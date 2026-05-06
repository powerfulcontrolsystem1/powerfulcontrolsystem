package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type step struct {
	Module string                 `json:"module"`
	Action string                 `json:"action"`
	OK     bool                   `json:"ok"`
	ID     int64                  `json:"id,omitempty"`
	Meta   map[string]interface{} `json:"meta,omitempty"`
	Error  string                 `json:"error,omitempty"`
}

type report struct {
	EmpresaID int64  `json:"empresa_id"`
	RunID     string `json:"run_id"`
	StartedAt string `json:"started_at"`
	Steps     []step `json:"steps"`
	Failures  []step `json:"failures"`
}

func main() {
	empresaID := flag.Int64("empresa_id", 7, "empresa_id de Motel Calipso")
	usuario := flag.String("usuario", "qa.calipso.modulos@powerfulcontrolsystem.local", "usuario auditor")
	publicBase := flag.String("public_base", "https://powerfulcontrolsystem.com", "base publica para QR de parqueadero")
	flag.Parse()

	dsn := strings.TrimSpace(os.Getenv("DB_EMPRESAS_DSN"))
	if dsn == "" {
		log.Fatal("DB_EMPRESAS_DSN no esta definido")
	}
	db, err := sql.Open(dbpkg.PostgresCompatDriverName(), dsn)
	must(err, "sql.Open")
	defer db.Close()
	must(db.Ping(), "db.Ping")
	_ = dbpkg.EnsurePostgresRuntimeCompat(db)

	runID := time.Now().Format("20060102-150405")
	rep := report{EmpresaID: *empresaID, RunID: runID, StartedAt: time.Now().Format(time.RFC3339)}
	add := func(s step) {
		rep.Steps = append(rep.Steps, s)
		if !s.OK {
			rep.Failures = append(rep.Failures, s)
		}
	}
	fail := func(module, action string, err error) {
		add(step{Module: module, Action: action, OK: false, Error: err.Error()})
		writeReport(rep)
		log.Fatalf("%s/%s: %v", module, action, err)
	}
	mustStep := func(module, action string, id int64, meta map[string]interface{}, err error) int64 {
		if err != nil {
			fail(module, action, err)
		}
		add(step{Module: module, Action: action, OK: true, ID: id, Meta: meta})
		return id
	}
	countRows := func(table string) int64 {
		var n int64
		_ = dbpkg.QueryRowCompat(db, "SELECT COUNT(1) FROM "+table+" WHERE empresa_id=?", *empresaID).Scan(&n)
		return n
	}

	mustStep("schema", "ensure_modulos_nuevos", 0, nil, ensureSchemas(db))

	periodo := time.Now().Format("2006-01")
	today := time.Now().Format("2006-01-02")

	terceroID, err := createTercero(db, *empresaID, runID, *usuario)
	mustStep("contabilidad_colombia", "crear_tercero_qa", terceroID, nil, err)
	compID, err := dbpkg.CreateEmpresaContabilidadComprobante(db, dbpkg.EmpresaContabilidadComprobante{
		EmpresaID: *empresaID, TipoComprobante: "nota_contable", FechaComprobante: today, PeriodoContable: periodo,
		TerceroID: terceroID, Concepto: "QA Calipso venta contable con IVA " + runID, OrigenModulo: "qa_calipso",
		ReferenciaExterna: "QA-CALIPSO-CONT-" + runID, Estado: "contabilizado", UsuarioCreador: *usuario,
		Lineas: []dbpkg.EmpresaContabilidadAsientoLinea{
			{CuentaCodigo: "110505", TerceroID: terceroID, Detalle: "Caja general QA", Debito: 119000},
			{CuentaCodigo: "4135", TerceroID: terceroID, Detalle: "Ingreso operativo QA", Credito: 100000, BaseGravable: 100000},
			{CuentaCodigo: "240805", TerceroID: terceroID, Detalle: "IVA generado QA", Credito: 19000, BaseGravable: 100000, ImpuestoCodigo: "IVA19"},
		},
	})
	mustStep("contabilidad_colombia", "crear_comprobante_balanceado", compID, map[string]interface{}{"periodo": periodo}, err)

	mustStep("contabilidad_avanzada", "seed_formatos_dian", 0, nil, dbpkg.SeedEmpresaContabilidadAvanzadaBase(db, *empresaID, *usuario, time.Now().Year()))
	formats, err := dbpkg.ListEmpresaExogenaFormatos(db, *empresaID, time.Now().Year())
	if err != nil {
		fail("contabilidad_avanzada", "listar_formatos", err)
	}
	if len(formats) == 0 {
		fail("contabilidad_avanzada", "listar_formatos", fmt.Errorf("no se encontraron formatos exogena"))
	}
	regID, err := dbpkg.CreateEmpresaExogenaRegistro(db, dbpkg.EmpresaExogenaRegistro{
		EmpresaID: *empresaID, FormatoID: formats[0].ID, TipoDocumento: "NIT", Documento: "900" + digits(runID),
		DigitoVerificacion: "1", RazonSocial: "Proveedor QA Calipso " + runID, Concepto: "Compra y retencion QA",
		CuentaCodigo: "2365", BaseValor: 450000, IVA: 85500, Retencion: 11250, Estado: "validado", UsuarioCreador: *usuario,
	})
	mustStep("contabilidad_avanzada", "crear_registro_exogena", regID, map[string]interface{}{"formato": formats[0].Formato}, err)
	gen, err := dbpkg.GenerateEmpresaExogenaFromAccounting(db, *empresaID, formats[0].ID, *usuario)
	mustStep("contabilidad_avanzada", "generar_exogena_desde_asientos", 0, map[string]interface{}{"registros_generados": gen}, err)
	nominaID, err := dbpkg.CreateEmpresaNominaElectronica(db, dbpkg.EmpresaNominaElectronica{
		EmpresaID: *empresaID, TipoDocumento: "CC", Documento: "1010" + digits(runID), Nombre: "Empleado QA Calipso " + runID,
		Periodo: periodo, FechaPago: today, SalarioBase: 2600000, Devengados: 2800000, Deducciones: 310000, EstadoDIAN: "borrador", UsuarioCreador: *usuario,
	})
	mustStep("contabilidad_avanzada", "crear_nomina_electronica", nominaID, nil, err)
	docID, err := dbpkg.CreateEmpresaDocumentoSoporte(db, dbpkg.EmpresaDocumentoSoporteElectronico{
		EmpresaID: *empresaID, TipoDocumento: "NIT", Documento: "830" + digits(runID), NombreProveedor: "Proveedor Servicios QA Calipso",
		FechaDocumento: today, Periodo: periodo, Concepto: "Mantenimiento modulo QA", Subtotal: 380000, IVA: 72200, Retenciones: 9500, EstadoDIAN: "borrador", UsuarioCreador: *usuario,
	})
	mustStep("contabilidad_avanzada", "crear_documento_soporte", docID, nil, err)
	activoID, err := dbpkg.CreateEmpresaActivoFijo(db, dbpkg.EmpresaActivoFijo{
		EmpresaID: *empresaID, Codigo: "QA-AF-" + runID, Nombre: "Controlador electrico QA Calipso", Categoria: "equipo",
		Serial: "SER-" + short(runID), Placa: "PL-AF-" + short(runID), FechaCompra: today, Costo: 2400000, ValorResidual: 200000, VidaUtilMeses: 48, MetodoDepreciacion: "linea_recta", CuentaActivo: "1528", CuentaDepreciacion: "1592", CuentaGasto: "5160",
		Ubicacion: "Recepcion Motel Calipso", Responsable: "QA Administracion", CentroCosto: "Operaciones", Proveedor: "Proveedor activos QA", MantenimientoCadaDias: 90, Estado: "activo", UsuarioCreador: *usuario,
	})
	mustStep("contabilidad_avanzada", "crear_activo_fijo", activoID, nil, err)
	depsAct, err := dbpkg.GenerarEmpresaActivosDepreciacion(db, *empresaID, periodo, *usuario)
	mustStep("contabilidad_avanzada", "generar_depreciacion_activos", 0, map[string]interface{}{"depreciaciones": len(depsAct)}, err)
	eventoActID, err := dbpkg.RegistrarEmpresaActivoEvento(db, dbpkg.EmpresaActivoEvento{EmpresaID: *empresaID, ActivoID: activoID, Tipo: "mantenimiento", FechaEvento: today, Valor: 185000, Detalle: "Mantenimiento QA activo fijo Calipso " + runID, UsuarioCreador: *usuario})
	mustStep("contabilidad_avanzada", "registrar_evento_activo", eventoActID, nil, err)
	actResumen, err := dbpkg.BuildEmpresaActivosFijosAvanzadoResumen(db, *empresaID, periodo)
	mustStep("contabilidad_avanzada", "dashboard_activos_avanzado", 0, map[string]interface{}{"activos": actResumen.ActivosActivos, "depreciacion_periodo": actResumen.DepreciacionPeriodoTotal, "eventos": len(actResumen.UltimosEventos)}, err)
	carteraID, err := dbpkg.CreateEmpresaCarteraCXP(db, dbpkg.EmpresaCarteraCXP{
		EmpresaID: *empresaID, Tipo: "cxc", TerceroID: terceroID, TerceroNombre: "Cliente QA Calipso " + runID,
		Documento: "QA-CXC-AV-" + runID, FechaEmision: today, FechaVencimiento: time.Now().AddDate(0, 0, 15).Format("2006-01-02"),
		CuentaCodigo: "130505", Concepto: "Cuenta por cobrar QA modulo avanzado", ValorOriginal: 540000, ValorPagado: 140000, Estado: "parcial",
		OrigenModulo: "qa_calipso", ReferenciaExterna: "QA-CXC-AV-" + runID, UsuarioCreador: *usuario,
	})
	mustStep("contabilidad_avanzada", "crear_cartera_cxc", carteraID, nil, err)
	libros, err := dbpkg.ListEmpresaLibroOficial(db, *empresaID, "mayor", periodo)
	mustStep("contabilidad_avanzada", "validar_libro_mayor", 0, map[string]interface{}{"lineas": len(libros)}, err)
	dashCont, err := dbpkg.BuildEmpresaContabilidadAvanzadaDashboard(db, *empresaID)
	mustStep("contabilidad_avanzada", "dashboard", 0, map[string]interface{}{"formatos": dashCont.FormatosExogena, "exogena": dashCont.RegistrosExogena, "activos": dashCont.ActivosFijos}, err)

	restID, err := createDomicilios(db, *empresaID, runID, *usuario)
	mustStep("domicilios", "flujo_restaurante_domiciliario_pedido", restID, nil, err)
	domDash, err := dbpkg.BuildEmpresaDomiciliosDashboard(db, *empresaID)
	mustStep("domicilios", "dashboard", 0, map[string]interface{}{"pedidos_ruta": domDash.PedidosRuta, "restaurantes": domDash.RestaurantesActivos, "domiciliarios_online": domDash.DomiciliariosOnline}, err)

	parkID, err := runParqueadero(db, *empresaID, runID, *usuario, *publicBase)
	mustStep("parqueadero", "ticket_qr_y_cobro", parkID, nil, err)
	parkDash, err := dbpkg.BuildEmpresaParqueaderoDashboard(db, *empresaID)
	mustStep("parqueadero", "dashboard", 0, map[string]interface{}{"abiertos": parkDash.Abiertos, "salidos_hoy": parkDash.SalidosHoy, "ingresos_hoy": parkDash.IngresosHoy}, err)

	aptID, err := runApartamentos(db, *empresaID, runID, *usuario)
	mustStep("apartamentos_turisticos", "unidad_reserva_checkout_limpieza", aptID, nil, err)
	aptDash, err := dbpkg.BuildEmpresaApartamentoTuristicoDashboard(db, *empresaID)
	mustStep("apartamentos_turisticos", "dashboard", 0, map[string]interface{}{"unidades": aptDash.Apartamentos, "reservas_activas": aptDash.ReservasActivas, "limpieza": aptDash.Limpieza}, err)

	carnetID, err := runCarnets(db, *empresaID, runID, *usuario)
	mustStep("carnets", "plantilla_carnet_impresion", carnetID, nil, err)
	carnetDash, err := dbpkg.BuildEmpresaCarnetsDashboard(db, *empresaID)
	if err != nil {
		fail("carnets", "dashboard", err)
	}
	mustStep("carnets", "dashboard", 0, map[string]interface{}{"vigentes": carnetDash.Vigentes, "plantillas": carnetDash.PlantillasActivas}, nil)

	taxiID, err := runTaxi(db, *empresaID, runID, *usuario)
	mustStep("taxi_system", "servicio_con_gps_y_mapa", taxiID, nil, err)
	taxiDash, err := dbpkg.BuildEmpresaTaxiDashboard(db, *empresaID)
	mustStep("taxi_system", "dashboard", 0, map[string]interface{}{"conductores_online": taxiDash.ConductoresOnline, "servicios_activos": taxiDash.ServiciosActivos, "solicitudes": len(taxiDash.Requests)}, err)

	prodID, err := runProduccionMRP(db, *empresaID, runID, *usuario)
	mustStep("produccion_mrp", "receta_orden_consumo_calidad_mrp", prodID, nil, err)
	prodDash, err := dbpkg.BuildEmpresaProduccionMRPDashboard(db, *empresaID)
	mustStep("produccion_mrp", "dashboard", 0, map[string]interface{}{"recetas_activas": prodDash.RecetasActivas, "ordenes_abiertas": prodDash.OrdenesAbiertas, "ordenes_cerradas": prodDash.OrdenesCerradas, "plan": len(prodDash.Plan)}, err)

	invID, err := runInventarioAvanzado(db, *empresaID, runID, *usuario)
	mustStep("inventario_avanzado", "lote_serial_reserva_valorizacion", invID, nil, err)
	invDash, err := dbpkg.BuildEmpresaInventarioAvanzadoDashboard(db, *empresaID)
	mustStep("inventario_avanzado", "dashboard", 0, map[string]interface{}{"lotes": invDash.LotesActivos, "reservas": invDash.ReservasActivas, "valor": invDash.ValorDisponible}, err)

	crmID, err := runCRMVentasAvanzadas(db, *empresaID, runID, *usuario)
	mustStep("crm_ventas_avanzadas", "lead_meta_scoring_cotizacion_forecast", crmID, nil, err)
	crmDash, err := dbpkg.BuildEmpresaCRMVentasAvanzadasDashboard(db, *empresaID, periodo)
	mustStep("crm_ventas_avanzadas", "dashboard", 0, map[string]interface{}{"leads_activos": crmDash.LeadsActivos, "forecast": crmDash.ForecastPonderado, "cotizaciones": crmDash.CotizacionesAbiertas}, err)

	tesID, err := runTesoreriaPresupuesto(db, *empresaID, runID, *usuario)
	mustStep("tesoreria_presupuesto", "cuentas_presupuesto_flujo", tesID, nil, err)
	tesDash, err := dbpkg.BuildEmpresaTesoreriaDashboard(db, *empresaID)
	mustStep("tesoreria_presupuesto", "dashboard", 0, map[string]interface{}{"cuentas": tesDash.CuentasActivas, "presupuestos": tesDash.PresupuestosActivos, "flujo_neto": tesDash.FlujoNeto}, err)

	impID, err := runImportacionesCosteo(db, *empresaID, runID, *usuario)
	mustStep("importaciones_costeo", "embarque_items_costos_distribucion", impID, nil, err)
	impDash, err := dbpkg.BuildEmpresaImportacionesCosteoDashboard(db, *empresaID)
	mustStep("importaciones_costeo", "dashboard", 0, map[string]interface{}{"abiertas": impDash.ImportacionesAbiertas, "cerradas": impDash.ImportacionesCerradas, "costo_total": impDash.CostoTotalCOP}, err)

	compraID, err := runComprasAvanzadas(db, *empresaID, runID, *usuario)
	mustStep("compras_avanzadas", "requisicion_cotizacion_aprobacion_recepcion", compraID, nil, err)
	compraDash, err := dbpkg.BuildEmpresaComprasAvanzadasDashboard(db, *empresaID)
	mustStep("compras_avanzadas", "dashboard", 0, map[string]interface{}{"abiertas": compraDash.RequisicionesAbiertas, "cotizaciones": compraDash.CotizacionesEnEvaluacion, "pendientes": compraDash.RecepcionesPendientes}, err)

	nomID, err := runNominaColombiaAvanzada(db, *empresaID, runID, *usuario)
	mustStep("nomina_colombia_avanzada", "empleado_conceptos_novedades_liquidacion_pila", nomID, nil, err)
	nomDash, err := dbpkg.BuildEmpresaNominaColombiaAvanzadaDashboard(db, *empresaID, periodo)
	mustStep("nomina_colombia_avanzada", "dashboard", 0, map[string]interface{}{"conceptos_activos": nomDash.ConceptosActivos, "novedades_pendientes": nomDash.NovedadesPendientes, "total_pila": nomDash.TotalPILA}, err)

	add(step{Module: "resumen", Action: "conteos_empresa_7", OK: true, Meta: map[string]interface{}{
		"contabilidad_comprobantes": countRows("empresa_contabilidad_colombia_comprobantes"),
		"domicilios_orders":         countRows("empresa_domicilios_orders"),
		"parqueadero_tickets":       countRows("empresa_parqueadero_tickets"),
		"apartamentos_reservas":     countRows("empresa_apartamentos_turisticos_reservas"),
		"carnets":                   countRows("empresa_carnets"),
		"taxi_requests":             countRows("empresa_taxi_requests"),
		"produccion_ordenes":        countRows("empresa_produccion_ordenes"),
		"tesoreria_flujo":           countRows("empresa_tesoreria_flujo_caja"),
		"importaciones_costeo":      countRows("empresa_importaciones_costeo"),
		"inventario_lotes_av":       countRows("empresa_inventario_lotes_avanzados"),
		"crm_metas_comerciales":     countRows("empresa_crm_metas_comerciales"),
		"compras_requisiciones":     countRows("empresa_compras_requisiciones"),
		"nomina_colombia_pila":      countRows("empresa_nomina_colombia_pila_resumen"),
		"activos_depreciacion":      countRows("empresa_contabilidad_activos_depreciacion"),
		"activos_eventos":           countRows("empresa_contabilidad_activos_eventos"),
	}})
	writeReport(rep)
	fmt.Println("RESULTADO_FINAL=OK qa_calipso_modulos_nuevos")
	fmt.Println("REPORTE=" + reportPath())
}

func ensureSchemas(db *sql.DB) error {
	for _, fn := range []func(*sql.DB) error{
		dbpkg.EnsureEmpresaContabilidadColombiaSchema,
		dbpkg.EnsureEmpresaContabilidadColombiaAvanzadaSchema,
		dbpkg.EnsureEmpresaDomiciliosSchema,
		dbpkg.EnsureEmpresaParqueaderoSchema,
		dbpkg.EnsureEmpresaApartamentosTuristicosSchema,
		dbpkg.EnsureEmpresaCarnetsSchema,
		dbpkg.EnsureEmpresaTaxiSystemSchema,
		dbpkg.EnsureEmpresaProduccionMRPSchema,
		dbpkg.EnsureEmpresaProductosSchema,
		dbpkg.EnsureEmpresaInventarioAvanzadoSchema,
		dbpkg.EnsureEmpresaCRMVentasAvanzadasSchema,
		dbpkg.EnsureEmpresaTesoreriaPresupuestoSchema,
		dbpkg.EnsureEmpresaImportacionesCosteoSchema,
		dbpkg.EnsureEmpresaComprasAvanzadasSchema,
		dbpkg.EnsureEmpresaNominaSchema,
	} {
		if err := fn(db); err != nil {
			return err
		}
	}
	return nil
}

func createTercero(db *sql.DB, empresaID int64, runID, usuario string) (int64, error) {
	return dbpkg.CreateEmpresaContabilidadTercero(db, dbpkg.EmpresaContabilidadTercero{
		EmpresaID: empresaID, TipoDocumento: "NIT", Documento: "901" + digits(runID), DigitoVerificacion: "4",
		Nombre: "Cliente Empresarial QA Calipso " + runID, TipoTercero: "cliente_proveedor", RegimenFiscal: "responsable_iva",
		Email: "qa.calipso." + strings.ReplaceAll(runID, "-", "") + "@empresa.local", Municipio: "Cali", Estado: "activo", UsuarioCreador: usuario,
	})
}

func createDomicilios(db *sql.DB, empresaID int64, runID, usuario string) (int64, error) {
	if err := dbpkg.UpsertEmpresaDomiciliosConfig(db, dbpkg.EmpresaDomiciliosConfig{EmpresaID: empresaID, NombreSistema: "Domicilios Motel Calipso QA", NombrePortal: "Calipso a domicilio", Moneda: "COP", RadioCoberturaKM: 12, RadioAsignacionKM: 8, TarifaBase: 4500, TarifaKM: 1100, ComisionPorcentaje: 12, TiempoPreparacionDefaultMin: 18, DomiciliariosPorRonda: 5, AutoAsignar: true, PermitirPedidosPublicos: true, ExigirCodigoEntrega: true, LatitudBase: 3.4516, LongitudBase: -76.5320, UsuarioCreador: usuario}); err != nil {
		return 0, err
	}
	restID, err := dbpkg.CreateDomicilioRestaurant(db, dbpkg.EmpresaDomicilioRestaurant{EmpresaID: empresaID, Codigo: "REST-QA-" + short(runID), Nombre: "Restaurante QA Calipso " + runID, Categoria: "Comida rapida", Responsable: "QA Central", Telefono: "3000000000", Email: "rest.qa@calipso.local", Direccion: "Motel Calipso", Latitud: 3.4516, Longitud: -76.5320, TiempoPreparacionMin: 16, ComisionPorcentaje: 12, AceptaPedidos: true, Pin: "1234", Estado: "activo", UsuarioCreador: usuario})
	if err != nil {
		return 0, err
	}
	itemID, err := dbpkg.UpsertDomicilioMenuItem(db, dbpkg.EmpresaDomicilioMenuItem{EmpresaID: empresaID, RestaurantID: restID, Codigo: "COMBO-" + short(runID), Nombre: "Combo QA Calipso", Descripcion: "Producto de prueba para domicilio", Categoria: "Combos", Precio: 28000, Disponible: true, TiempoPreparacionMin: 12, Orden: 1})
	if err != nil {
		return 0, err
	}
	courierID, err := dbpkg.CreateDomicilioCourier(db, dbpkg.EmpresaDomicilioCourier{EmpresaID: empresaID, Codigo: "DOM-QA-" + short(runID), Nombre: "Domiciliario QA Calipso " + runID, Documento: "DOM" + digits(runID), Telefono: "3110000000", VehiculoTipo: "moto", VehiculoPlaca: "D" + short(runID), ZonaBase: "Calipso", Pin: "1234", Estado: "activo", UsuarioCreador: usuario})
	if err != nil {
		return 0, err
	}
	if err := dbpkg.UpdateDomicilioCourierPresence(db, empresaID, courierID, true, true); err != nil {
		return 0, err
	}
	if err := dbpkg.UpdateDomicilioCourierLocation(db, empresaID, courierID, 0, dbpkg.EmpresaDomicilioTrackPoint{Latitud: 3.4520, Longitud: -76.5318, PrecisionMetros: 7, VelocidadKMH: 16}); err != nil {
		return 0, err
	}
	order, err := dbpkg.CreateDomicilioOrder(db, dbpkg.EmpresaDomicilioOrder{EmpresaID: empresaID, RestaurantID: restID, ClienteNombre: "Cliente Domicilio QA " + runID, ClienteTelefono: "3020000000", ClienteDireccion: "Direccion QA Cali", ClienteLatitud: 3.4550, ClienteLongitud: -76.5290, MetodoPago: "efectivo", Canal: "web", Propina: 2000, Items: []dbpkg.EmpresaDomicilioOrderItem{{MenuItemID: itemID, Cantidad: 2}}})
	if err != nil {
		return 0, err
	}
	offers, err := dbpkg.ListDomicilioOffersForCourier(db, empresaID, courierID)
	if err != nil {
		return 0, err
	}
	if len(offers) == 0 {
		offers, err = dbpkg.DispatchDomicilioOrder(db, empresaID, order.ID, 5)
		if err != nil {
			return 0, err
		}
	}
	if len(offers) > 0 {
		order, err = dbpkg.RespondDomicilioOffer(db, empresaID, offers[0].ID, courierID, true, "QA acepta domicilio")
		if err != nil {
			return 0, err
		}
	}
	if _, err = dbpkg.UpdateDomicilioOrderState(db, empresaID, order.ID, courierID, "restaurante", "listo", "QA listo", ""); err != nil {
		return 0, err
	}
	if _, err = dbpkg.UpdateDomicilioOrderState(db, empresaID, order.ID, courierID, "domiciliario", "recogido", "QA recogido", ""); err != nil {
		return 0, err
	}
	if err := dbpkg.UpdateDomicilioCourierLocation(db, empresaID, courierID, order.ID, dbpkg.EmpresaDomicilioTrackPoint{Latitud: 3.4540, Longitud: -76.5300, PrecisionMetros: 6, VelocidadKMH: 22}); err != nil {
		return 0, err
	}
	_, err = dbpkg.UpdateDomicilioOrderState(db, empresaID, order.ID, courierID, "domiciliario", "entregado", "QA entregado con codigo", order.CodigoEntrega)
	return restID, err
}

func runParqueadero(db *sql.DB, empresaID int64, runID, usuario, publicBase string) (int64, error) {
	if err := dbpkg.UpsertEmpresaParqueaderoConfig(db, dbpkg.EmpresaParqueaderoConfig{EmpresaID: empresaID, Nombre: "Parqueadero Motel Calipso QA", PrefijoTicket: "CAL", Moneda: "COP", MinutosTolerancia: 5, MinutosBase: 60, TarifaBase: 5000, MinutosFraccion: 15, TarifaFraccion: 1500, TarifaDiaMax: 28000, CobrarFraccionCompleta: true, IVAIncluido: true, SalidaRequiereQR: true, ImprimirTicketEntrada: true, ImprimirReciboSalida: true, UsuarioCreador: usuario}); err != nil {
		return 0, err
	}
	t, err := dbpkg.CreateEmpresaParqueaderoTicket(db, dbpkg.EmpresaParqueaderoTicket{EmpresaID: empresaID, Placa: "QA" + short(runID), TipoVehiculo: "carro", ClienteNombre: "Cliente Parqueadero QA", Observaciones: "QA Calipso", UsuarioCreador: usuario}, publicBase)
	if err != nil {
		return 0, err
	}
	_, _ = dbpkg.ExecCompat(db, `UPDATE empresa_parqueadero_tickets SET fecha_entrada=? WHERE empresa_id=? AND id=?`, time.Now().Add(-145*time.Minute).Format("2006-01-02 15:04:05"), empresaID, t.ID)
	closed, cobro, err := dbpkg.CerrarEmpresaParqueaderoTicket(db, empresaID, t.ID, "efectivo", usuario)
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(closed.QRPayload) == "" || cobro.Total <= 0 {
		return 0, fmt.Errorf("ticket parqueadero sin QR o sin cobro automatico")
	}
	return closed.ID, nil
}

func runApartamentos(db *sql.DB, empresaID int64, runID, usuario string) (int64, error) {
	if err := dbpkg.UpsertEmpresaApartamentoTuristicoConfig(db, dbpkg.EmpresaApartamentoTuristicoConfig{EmpresaID: empresaID, NombreSistema: "Apartamentos turisticos Calipso QA", Moneda: "COP", HoraCheckIn: "15:00", HoraCheckOut: "11:00", DepositoPorcentaje: 30, ImpuestoPorcentaje: 19, AutoProgramarLimpieza: true, PermitirReservasPublicas: true, RequerirDocumentoHuesped: true, UsuarioCreador: usuario}); err != nil {
		return 0, err
	}
	unitID, err := dbpkg.CreateEmpresaApartamentoTuristicoUnidad(db, dbpkg.EmpresaApartamentoTuristicoUnidad{EmpresaID: empresaID, Codigo: "APT-QA-" + short(runID), Nombre: "Apartamento QA Calipso " + runID, Tipo: "apartamento", Ubicacion: "Cali", Capacidad: 4, Habitaciones: 2, Camas: 3, Banos: 2, PrecioBaseNoche: 220000, TarifaLimpieza: 45000, DepositoSugerido: 120000, EstadoOperativo: "activo", EstadoOcupacion: "disponible", Amenidades: "wifi,cocina,aire", ReglasCasa: "sin fiestas", UsuarioCreador: usuario})
	if err != nil {
		return 0, err
	}
	_, err = dbpkg.CreateEmpresaApartamentoTuristicoTarifa(db, dbpkg.EmpresaApartamentoTuristicoTarifa{EmpresaID: empresaID, ApartamentoID: unitID, Nombre: "Tarifa directa QA", Canal: "directo", PrecioNoche: 220000, MinimoNoches: 1, Estado: "activo", UsuarioCreador: usuario})
	if err != nil {
		return 0, err
	}
	in := time.Now().AddDate(0, 0, 30).Format("2006-01-02")
	out := time.Now().AddDate(0, 0, 32).Format("2006-01-02")
	resID, err := dbpkg.CreateEmpresaApartamentoTuristicoReserva(db, dbpkg.EmpresaApartamentoTuristicoReserva{EmpresaID: empresaID, ApartamentoID: unitID, HuespedNombre: "Huesped QA Calipso " + runID, HuespedDocumento: "APT" + digits(runID), HuespedTelefono: "3040000000", HuespedEmail: "apt.qa@calipso.local", CantidadHuespedes: 2, FechaEntrada: in, FechaSalida: out, Canal: "directo", EstadoReserva: "confirmada", EstadoPago: "pagado", UsuarioCreador: usuario})
	if err != nil {
		return 0, err
	}
	if err := dbpkg.CambiarEstadoApartamentoTuristicoReserva(db, empresaID, resID, "checkin", usuario); err != nil {
		return 0, err
	}
	if err := dbpkg.CambiarEstadoApartamentoTuristicoReserva(db, empresaID, resID, "checkout", usuario); err != nil {
		return 0, err
	}
	_, err = dbpkg.CreateEmpresaApartamentoTuristicoTarea(db, dbpkg.EmpresaApartamentoTuristicoTarea{EmpresaID: empresaID, ApartamentoID: unitID, ReservaID: resID, Tipo: "mantenimiento", Prioridad: "media", Estado: "pendiente", Responsable: "QA Operaciones", FechaProgramada: time.Now().AddDate(0, 0, 33).Format("2006-01-02 10:00:00"), CostoEstimado: 35000, Descripcion: "Revision QA posterior a reserva", UsuarioCreador: usuario})
	return resID, err
}

func runCarnets(db *sql.DB, empresaID int64, runID, usuario string) (int64, error) {
	templateID, err := dbpkg.EnsureEmpresaCarnetDefaultTemplate(db, empresaID, usuario)
	if err != nil {
		return 0, err
	}
	id, err := dbpkg.CreateEmpresaCarnet(db, dbpkg.EmpresaCarnet{EmpresaID: empresaID, PlantillaID: templateID, Codigo: "QA-CARNET-" + short(runID), TipoPersona: "empleado", NombreCompleto: "Empleado QA Calipso " + runID, Documento: "CC " + digits(runID), Cargo: "Administrador QA", Area: "Operacion", Email: "carnet.qa@calipso.local", Telefono: "3050000000", NivelAcceso: "administracion", GrupoSanguineo: "O+", ContactoEmergencia: "Contacto QA", TelefonoEmergencia: "3001112233", FechaVencimiento: time.Now().AddDate(1, 0, 0).Format("2006-01-02"), UsuarioCreador: usuario})
	if err != nil {
		return 0, err
	}
	if err := dbpkg.MarkEmpresaCarnetImpreso(db, empresaID, id, usuario); err != nil {
		return 0, err
	}
	return id, nil
}

func runTaxi(db *sql.DB, empresaID int64, runID, usuario string) (int64, error) {
	if err := dbpkg.UpsertEmpresaTaxiConfig(db, dbpkg.EmpresaTaxiConfig{EmpresaID: empresaID, NombreSistema: "Taxi System Calipso QA", NombrePortal: "Pide tu taxi Calipso", RadioBusquedaKM: 9, ConductoresPorRonda: 5, TimeoutOfertaSegundos: 25, PermitirRegistroCliente: true, PermitirUbicacionCliente: true, PermitirDespachoAutomatico: true, LatitudBase: 3.4516, LongitudBase: -76.5320, UsuarioCreador: usuario}); err != nil {
		return 0, err
	}
	driverID, err := dbpkg.CreateEmpresaTaxiDriver(db, dbpkg.EmpresaTaxiDriver{EmpresaID: empresaID, Codigo: "TX-QA-" + short(runID), Nombre: "Conductor QA Calipso " + runID, Documento: "TX" + digits(runID), Telefono: "3120000000", Email: "taxi.qa@calipso.local", VehiculoPlaca: "TX" + short(runID), VehiculoModelo: "Spark", VehiculoTipo: "taxi", VehiculoColor: "amarillo", LicenciaConduccion: "LC" + digits(runID), GPSCodigo: "GPS-QA-" + short(runID), GPSTipo: "teltonika", GPSProveedor: "QA GPS", GPSProtocolo: "traccar", Pin: "1234", Estado: "activo", UsuarioCreador: usuario})
	if err != nil {
		return 0, err
	}
	if err := dbpkg.UpdateTaxiDriverPresence(db, empresaID, driverID, true, true); err != nil {
		return 0, err
	}
	if err := dbpkg.UpdateTaxiDriverLocation(db, empresaID, driverID, 0, dbpkg.EmpresaTaxiRoutePoint{Latitud: 3.4521, Longitud: -76.5319, PrecisionMetros: 5, VelocidadKMH: 18}); err != nil {
		return 0, err
	}
	customer, err := dbpkg.RegisterTaxiCustomer(db, dbpkg.EmpresaTaxiCustomer{EmpresaID: empresaID, Nombre: "Cliente Taxi QA " + runID, Documento: "CTX" + digits(runID), Telefono: "315" + digits(runID), Email: "cliente.taxi.qa@calipso.local", Pin: "1234"})
	if err != nil {
		return 0, err
	}
	req, err := dbpkg.CreateTaxiRequest(db, dbpkg.EmpresaTaxiRequest{EmpresaID: empresaID, CustomerID: customer.ID, ClienteNombre: customer.Nombre, ClienteTelefono: customer.Telefono, ClienteDocumento: customer.Documento, RecogerTexto: "Motel Calipso", RecogerLatitud: 3.4516, RecogerLongitud: -76.5320, DestinoTexto: "Centro Cali", DestinoLatitud: 3.4372, DestinoLongitud: -76.5225, ComparteUbicacionCliente: true, MetodoSolicitud: "app", Canal: "web", Notas: "QA servicio taxi"})
	if err != nil {
		return 0, err
	}
	offers, err := dbpkg.ListTaxiOffersForDriver(db, empresaID, driverID)
	if err != nil {
		return 0, err
	}
	if len(offers) == 0 {
		offers, err = dbpkg.DispatchTaxiRequestToNearbyDrivers(db, empresaID, req.ID, 0)
		if err != nil {
			return 0, err
		}
	}
	if len(offers) == 0 {
		return 0, fmt.Errorf("no se generaron ofertas de taxi para conductor online")
	}
	req, err = dbpkg.RespondTaxiOffer(db, empresaID, offers[0].ID, driverID, true, "QA acepta servicio")
	if err != nil {
		return 0, err
	}
	if err := dbpkg.AddTaxiCustomerRoutePoint(db, empresaID, req.ID, customer.ID, dbpkg.EmpresaTaxiRoutePoint{Latitud: 3.4517, Longitud: -76.5321, PrecisionMetros: 8}); err != nil {
		return 0, err
	}
	if err := dbpkg.UpdateTaxiDriverLocation(db, empresaID, driverID, req.ID, dbpkg.EmpresaTaxiRoutePoint{Latitud: 3.4450, Longitud: -76.5270, PrecisionMetros: 5, VelocidadKMH: 28}); err != nil {
		return 0, err
	}
	if _, err = dbpkg.UpdateTaxiRequestState(db, empresaID, req.ID, driverID, "en_camino", "QA conductor en camino"); err != nil {
		return 0, err
	}
	if _, err = dbpkg.UpdateTaxiRequestState(db, empresaID, req.ID, driverID, "abordo", "QA cliente abordo"); err != nil {
		return 0, err
	}
	if _, err = dbpkg.UpdateTaxiRequestState(db, empresaID, req.ID, driverID, "completado", "QA servicio completado"); err != nil {
		return 0, err
	}
	return req.ID, nil
}

func runProduccionMRP(db *sql.DB, empresaID int64, runID, usuario string) (int64, error) {
	if err := dbpkg.UpsertEmpresaProduccionMRPConfig(db, dbpkg.EmpresaProduccionMRPConfig{EmpresaID: empresaID, NombreSistema: "Produccion Calipso QA", Moneda: "COP", CosteoModo: "estandar", AprobarOrdenes: true, CerrarConCalidad: true, UsuarioCreador: usuario}); err != nil {
		return 0, err
	}
	recID, err := dbpkg.UpsertEmpresaProduccionReceta(db, dbpkg.EmpresaProduccionReceta{
		EmpresaID: empresaID, Codigo: "BOM-QA-" + short(runID), Nombre: "Kit amenidades QA Calipso " + runID, ProductoTerminadoNombre: "Kit amenidades QA", Version: "1.0", Unidad: "kit", CantidadBase: 1, CostoEstandar: 5200, MermaPorcentaje: 2, TiempoEstimadoMin: 14, Estado: "activo", UsuarioCreador: usuario,
		Componentes: []dbpkg.EmpresaProduccionComponente{
			{ProductoNombre: "Shampoo QA", Unidad: "und", Cantidad: 1, CostoUnitario: 900, Etapa: "alistamiento", Obligatoria: true},
			{ProductoNombre: "Jabon QA", Unidad: "und", Cantidad: 1, CostoUnitario: 1200, Etapa: "alistamiento", Obligatoria: true},
			{ProductoNombre: "Empaque QA", Unidad: "und", Cantidad: 1, CostoUnitario: 500, Etapa: "empaque", Obligatoria: true},
		},
	})
	if err != nil {
		return 0, err
	}
	orden, err := dbpkg.CreateEmpresaProduccionOrden(db, dbpkg.EmpresaProduccionOrden{EmpresaID: empresaID, RecetaID: recID, ProductoTerminadoNombre: "Kit amenidades QA", CantidadPlanificada: 25, Estado: "programada", Prioridad: "alta", Responsable: "QA Operaciones", Observaciones: "Orden QA produccion/MRP " + runID, UsuarioCreador: usuario})
	if err != nil {
		return 0, err
	}
	if _, err := dbpkg.CambiarEstadoEmpresaProduccionOrden(db, empresaID, orden.ID, "en_proceso", usuario); err != nil {
		return 0, err
	}
	if _, err := dbpkg.RegistrarEmpresaProduccionConsumo(db, dbpkg.EmpresaProduccionConsumo{EmpresaID: empresaID, OrdenID: orden.ID, ProductoNombre: "Shampoo QA ajuste", CantidadPlanificada: 25, CantidadConsumida: 25, CostoUnitario: 900, LoteCodigo: "LQA-" + short(runID), UsuarioCreador: usuario}); err != nil {
		return 0, err
	}
	if _, err := dbpkg.RegistrarEmpresaProduccionCalidad(db, dbpkg.EmpresaProduccionCalidad{EmpresaID: empresaID, OrdenID: orden.ID, Resultado: "aprobado", CantidadAprobada: 25, Responsable: "QA Calidad", ChecklistJSON: `{"contenido":"ok","empaque":"ok"}`}); err != nil {
		return 0, err
	}
	if _, err := dbpkg.GenerarEmpresaProduccionMRPPlan(db, empresaID, time.Now().Format("2006-01"), usuario); err != nil {
		return 0, err
	}
	return orden.ID, nil
}

func runTesoreriaPresupuesto(db *sql.DB, empresaID int64, runID, usuario string) (int64, error) {
	periodo := time.Now().Format("2006-01")
	if err := dbpkg.UpsertEmpresaTesoreriaConfig(db, dbpkg.EmpresaTesoreriaConfig{EmpresaID: empresaID, NombreSistema: "Tesoreria Calipso QA", Moneda: "COP", PeriodoTrabajo: periodo, MetodoProyeccion: "mensual", AlertaSaldoMinimo: true, RequiereAprobacionPago: true, UsuarioCreador: usuario}); err != nil {
		return 0, err
	}
	cuentaID, err := dbpkg.UpsertEmpresaTesoreriaCuenta(db, dbpkg.EmpresaTesoreriaCuenta{EmpresaID: empresaID, Codigo: "QA-BANCO-" + short(runID), Nombre: "Banco QA Calipso " + runID, Tipo: "banco", Entidad: "Banco QA", Numero: "QA" + digits(runID), Moneda: "COP", SaldoInicial: 18000000, SaldoActual: 18000000, SaldoMinimo: 2500000, Responsable: "QA Tesoreria", Estado: "activo", UsuarioCreador: usuario})
	if err != nil {
		return 0, err
	}
	presID, err := dbpkg.UpsertEmpresaTesoreriaPresupuesto(db, dbpkg.EmpresaTesoreriaPresupuesto{EmpresaID: empresaID, Codigo: "QA-PRES-" + short(runID), Nombre: "Presupuesto QA Calipso " + runID, Periodo: periodo, Escenario: "base", IngresosMeta: 52000000, EgresosMeta: 31000000, SaldoInicial: 18000000, Estado: "aprobado", Responsable: "QA Gerencia", UsuarioCreador: usuario})
	if err != nil {
		return 0, err
	}
	if _, err := dbpkg.UpsertEmpresaTesoreriaPartida(db, dbpkg.EmpresaTesoreriaPartida{EmpresaID: empresaID, PresupuestoID: presID, Categoria: "ventas", Tipo: "ingreso", Concepto: "Ingresos QA hospedaje y servicios", ValorPresupuestado: 52000000, ValorEjecutado: 12000000, Periodicidad: "mensual", Estado: "activo", UsuarioCreador: usuario}); err != nil {
		return 0, err
	}
	if _, err := dbpkg.UpsertEmpresaTesoreriaPartida(db, dbpkg.EmpresaTesoreriaPartida{EmpresaID: empresaID, PresupuestoID: presID, Categoria: "proveedores", Tipo: "egreso", Concepto: "Pagos QA proveedores", ValorPresupuestado: 18000000, ValorEjecutado: 4500000, Periodicidad: "mensual", Estado: "activo", UsuarioCreador: usuario}); err != nil {
		return 0, err
	}
	if _, err := dbpkg.GenerarEmpresaTesoreriaFlujoDesdePresupuesto(db, empresaID, presID, usuario); err != nil {
		return 0, err
	}
	if _, err := dbpkg.CreateEmpresaTesoreriaFlujo(db, dbpkg.EmpresaTesoreriaFlujo{EmpresaID: empresaID, CuentaID: cuentaID, PresupuestoID: presID, FechaFlujo: periodo + "-20", Periodo: periodo, Tipo: "egreso", Categoria: "impuestos", Concepto: "Reserva impuestos QA", Valor: 3900000, OrigenModulo: "qa_calipso", Estado: "programado", UsuarioCreador: usuario}); err != nil {
		return 0, err
	}
	return presID, nil
}

func runImportacionesCosteo(db *sql.DB, empresaID int64, runID, usuario string) (int64, error) {
	impID, err := dbpkg.CreateEmpresaImportacionCosteo(db, dbpkg.EmpresaImportacionCosteo{
		EmpresaID: empresaID, Codigo: "IMP-QA-" + short(runID), Proveedor: "Proveedor internacional QA Calipso " + runID,
		PaisOrigen: "China", Incoterm: "FOB", MonedaOrigen: "USD", TRM: 3925, FechaDocumento: time.Now().Format("2006-01-02"),
		DocumentoReferencia: "BL-QA-" + short(runID), Estado: "en_transito", UsuarioCreador: usuario,
	})
	if err != nil {
		return 0, err
	}
	if _, err := dbpkg.CreateEmpresaImportacionItem(db, dbpkg.EmpresaImportacionItem{EmpresaID: empresaID, ImportacionID: impID, ProductoNombre: "Sensor puerta importado QA", SKU: "SEN-QA-" + short(runID), Cantidad: 40, Unidad: "und", PesoKG: 16, VolumenM3: 0.3, CostoUnitarioOrigen: 11.5}, 3925); err != nil {
		return 0, err
	}
	if _, err := dbpkg.CreateEmpresaImportacionItem(db, dbpkg.EmpresaImportacionItem{EmpresaID: empresaID, ImportacionID: impID, ProductoNombre: "Controlador relay importado QA", SKU: "REL-QA-" + short(runID), Cantidad: 18, Unidad: "und", PesoKG: 22, VolumenM3: 0.45, CostoUnitarioOrigen: 37.2}, 3925); err != nil {
		return 0, err
	}
	for _, c := range []dbpkg.EmpresaImportacionCosto{
		{EmpresaID: empresaID, ImportacionID: impID, Tipo: "flete", Concepto: "Flete internacional QA", BaseDistribucion: "peso", ValorCOP: 1720000, CuentaContable: "1435", UsuarioCreador: usuario},
		{EmpresaID: empresaID, ImportacionID: impID, Tipo: "arancel", Concepto: "Arancel nacionalizacion QA", BaseDistribucion: "valor", ValorCOP: 1360000, CuentaContable: "1435", UsuarioCreador: usuario},
		{EmpresaID: empresaID, ImportacionID: impID, Tipo: "agencia_aduanas", Concepto: "Agencia de aduanas QA", BaseDistribucion: "cantidad", ValorCOP: 480000, CuentaContable: "1435", UsuarioCreador: usuario},
	} {
		if _, err := dbpkg.CreateEmpresaImportacionCosto(db, c); err != nil {
			return 0, err
		}
	}
	row, err := dbpkg.DistribuirEmpresaImportacionCostos(db, empresaID, impID, usuario)
	if err != nil {
		return 0, err
	}
	if len(row.Items) == 0 || row.CostoTotalCOP <= 0 {
		return 0, fmt.Errorf("importacion sin costeo distribuido")
	}
	return impID, nil
}

func runInventarioAvanzado(db *sql.DB, empresaID int64, runID, usuario string) (int64, error) {
	bodegas, err := dbpkg.GetBodegasByEmpresa(db, empresaID, false)
	if err != nil {
		return 0, err
	}
	var bodegaID int64
	if len(bodegas) > 0 {
		bodegaID = bodegas[0].ID
	} else {
		bodegaID, err = dbpkg.CreateBodega(db, dbpkg.Bodega{EmpresaID: empresaID, Codigo: "BOD-QA-" + short(runID), Nombre: "Bodega QA inventario " + runID, UsuarioCreador: usuario, Estado: "activo"})
		if err != nil {
			return 0, err
		}
	}
	productoID, err := dbpkg.CreateProducto(db, dbpkg.Producto{EmpresaID: empresaID, BodegaPrincipalID: bodegaID, SKU: "INVAV-QA-" + short(runID), Nombre: "Producto QA inventario avanzado " + runID, UnidadMedida: "und", Costo: 18500, Precio: 29000, StockMinimo: 3, StockMaximo: 80, ManejaVencimiento: true, FechaVencimiento: time.Now().AddDate(0, 6, 0).Format("2006-01-02"), DiasAlertaVencimiento: 45, UsuarioCreador: usuario, Estado: "activo"}, 0, "QA inventario avanzado")
	if err != nil {
		return 0, err
	}
	loteID, err := dbpkg.CreateEmpresaInventarioLoteAvanzado(db, dbpkg.EmpresaInventarioLoteAvanzado{EmpresaID: empresaID, ProductoID: productoID, BodegaID: bodegaID, LoteCodigo: "LOT-QA-" + short(runID), FechaFabricacion: time.Now().AddDate(0, -1, 0).Format("2006-01-02"), FechaVencimiento: time.Now().AddDate(0, 3, 0).Format("2006-01-02"), CantidadInicial: 18, CostoUnitario: 18500, EstadoCalidad: "liberado", Proveedor: "Proveedor QA inventario", DocumentoRef: "DOC-INV-" + short(runID), UbicacionInterna: "Rack QA", UsuarioCreador: usuario})
	if err != nil {
		return 0, err
	}
	serialID, err := dbpkg.CreateEmpresaInventarioSerialAvanzado(db, dbpkg.EmpresaInventarioSerialAvanzado{EmpresaID: empresaID, LoteID: loteID, ProductoID: productoID, BodegaID: bodegaID, Serial: "SER-INV-" + short(runID), EstadoOperativo: "operativo", EstadoInventario: "disponible", FechaIngreso: time.Now().Format("2006-01-02"), GarantiaHasta: time.Now().AddDate(1, 0, 0).Format("2006-01-02"), UsuarioCreador: usuario})
	if err != nil {
		return 0, err
	}
	resID, err := dbpkg.CreateEmpresaInventarioReservaAvanzada(db, dbpkg.EmpresaInventarioReservaAvanzada{EmpresaID: empresaID, ProductoID: productoID, BodegaID: bodegaID, LoteID: loteID, SerialID: serialID, Cantidad: 1, OrigenModulo: "qa_calipso", OrigenRef: "RSV-INV-" + short(runID), ClienteNombre: "Cliente QA inventario", FechaReserva: time.Now().Format("2006-01-02"), FechaExpira: time.Now().AddDate(0, 0, 2).Format("2006-01-02"), UsuarioCreador: usuario})
	if err != nil {
		return 0, err
	}
	if err := dbpkg.ConfirmarEmpresaInventarioReservaAvanzada(db, empresaID, resID, usuario); err != nil {
		return 0, err
	}
	val, err := dbpkg.ListEmpresaInventarioValorizacionAvanzada(db, empresaID, 50)
	if err != nil {
		return 0, err
	}
	if len(val) == 0 {
		return 0, fmt.Errorf("inventario avanzado sin valorizacion")
	}
	return loteID, nil
}

func runCRMVentasAvanzadas(db *sql.DB, empresaID int64, runID, usuario string) (int64, error) {
	leadID, err := dbpkg.SeedEmpresaCRMVentasAvanzadasDemo(db, empresaID, usuario)
	if err != nil {
		return 0, err
	}
	scores, err := dbpkg.ListEmpresaCRMLeadScores(db, empresaID, 10)
	if err != nil {
		return 0, err
	}
	if len(scores) == 0 {
		return 0, fmt.Errorf("crm avanzado sin scoring")
	}
	dash, err := dbpkg.BuildEmpresaCRMVentasAvanzadasDashboard(db, empresaID, time.Now().Format("2006-01"))
	if err != nil {
		return 0, err
	}
	if dash.ForecastPonderado <= 0 || dash.CotizacionesAbiertas <= 0 || len(dash.Agenda) == 0 {
		return 0, fmt.Errorf("crm avanzado sin forecast/cotizaciones/agenda")
	}
	return leadID, nil
}

func runComprasAvanzadas(db *sql.DB, empresaID int64, runID, usuario string) (int64, error) {
	reqID, err := dbpkg.CreateEmpresaCompraRequisicion(db, dbpkg.EmpresaCompraRequisicion{
		EmpresaID: empresaID, Codigo: "REQ-QA-" + short(runID), Solicitante: "QA Compras Calipso", Area: "Operaciones",
		CentroCosto: "Motel Calipso", Prioridad: "alta", FechaSolicitud: time.Now().Format("2006-01-02"), FechaNecesidad: time.Now().AddDate(0, 0, 4).Format("2006-01-02"),
		EstadoFlujo: "solicitada", Justificacion: "QA ciclo profesional de compras avanzadas", UsuarioCreador: usuario,
		Items: []dbpkg.EmpresaCompraRequisicionItem{
			{ProductoNombre: "Amenidad premium QA", CantidadSolicitada: 60, Unidad: "und", CostoEstimado: 3900, ProveedorSugerido: "Proveedor hotelero QA"},
			{ProductoNombre: "Relay domotico QA", CantidadSolicitada: 10, Unidad: "und", CostoEstimado: 85000, ProveedorSugerido: "Proveedor electrico QA"},
		},
	})
	if err != nil {
		return 0, err
	}
	cotID, err := dbpkg.CreateEmpresaCompraCotizacion(db, dbpkg.EmpresaCompraCotizacion{
		EmpresaID: empresaID, RequisicionID: reqID, ProveedorNombre: "Proveedor Compras QA Calipso", Numero: "COT-QA-" + short(runID),
		FechaCotizacion: time.Now().Format("2006-01-02"), ValidezHasta: time.Now().AddDate(0, 0, 12).Format("2006-01-02"),
		TiempoEntregaDias: 2, Subtotal: 1084000, Impuestos: 205960, CondicionesPago: "Credito 15 dias", Estado: "evaluacion", UsuarioCreador: usuario,
	})
	if err != nil {
		return 0, err
	}
	if _, err := dbpkg.ResolverEmpresaCompraAprobacion(db, dbpkg.EmpresaCompraAprobacion{EmpresaID: empresaID, RequisicionID: reqID, CotizacionID: cotID, Nivel: 1, Aprobador: usuario, Decision: "aprobada", Comentario: "QA aprobacion compra avanzada", MontoAutorizado: 1289960}); err != nil {
		return 0, err
	}
	items, err := dbpkg.ListEmpresaCompraRequisicionItems(db, empresaID, reqID)
	if err != nil {
		return 0, err
	}
	recItems := make([]dbpkg.EmpresaCompraRecepcionItem, 0, len(items))
	for _, item := range items {
		recItems = append(recItems, dbpkg.EmpresaCompraRecepcionItem{
			RequisicionItemID: item.ID, ProductoNombre: item.ProductoNombre, CantidadOrdenada: item.CantidadSolicitada,
			CantidadRecibida: item.CantidadSolicitada, CostoUnitario: item.CostoEstimado, Lote: "QA-" + short(runID), EstadoCalidad: "aprobado",
		})
	}
	if _, err := dbpkg.CreateEmpresaCompraRecepcion(db, dbpkg.EmpresaCompraRecepcion{EmpresaID: empresaID, RequisicionID: reqID, CotizacionID: cotID, ProveedorNombre: "Proveedor Compras QA Calipso", Documento: "REM-QA-" + short(runID), FechaRecepcion: time.Now().Format("2006-01-02"), EstadoRecepcion: "total", Responsable: usuario, UsuarioCreador: usuario, Items: recItems}); err != nil {
		return 0, err
	}
	row, err := dbpkg.GetEmpresaCompraRequisicion(db, empresaID, reqID)
	if err != nil {
		return 0, err
	}
	if row.EstadoFlujo != "recibida_total" || len(row.Cotizaciones) == 0 || len(row.Recepciones) == 0 {
		return 0, fmt.Errorf("flujo compras avanzadas incompleto")
	}
	return reqID, nil
}

func runNominaColombiaAvanzada(db *sql.DB, empresaID int64, runID, usuario string) (int64, error) {
	periodo := time.Now().Format("2006-01")
	desde := periodo + "-01"
	hasta := periodo + "-28"
	if err := dbpkg.SeedEmpresaNominaColombiaAvanzadaDemo(db, empresaID, usuario); err != nil {
		return 0, err
	}
	empleadoID, err := dbpkg.CreateEmpresaNominaEmpleado(db, dbpkg.EmpresaNominaEmpleado{
		EmpresaID:                empresaID,
		EmpleadoCodigo:           "NOM-QA-" + short(runID),
		EmpleadoNombre:           "Empleado Nomina QA Calipso " + runID,
		EmpleadoDocumento:        "QA" + digits(runID),
		Cargo:                    "Auxiliar operativo QA",
		TipoContrato:             "indefinido",
		FechaIngreso:             desde,
		SalarioBasicoMensual:     2500000,
		AuxilioTransporteMensual: 162000,
		BonificacionFijaMensual:  180000,
		DeduccionFijaMensual:     40000,
		JornadaHorasDia:          8,
		IncluirAuxilioTransporte: true,
		UsuarioCreador:           usuario,
		Estado:                   "activo",
		Observaciones:            "QA nomina Colombia avanzada " + runID,
	})
	if err != nil {
		return 0, err
	}
	conceptoID, err := dbpkg.UpsertEmpresaNominaConceptoColombia(db, dbpkg.EmpresaNominaConceptoColombia{
		EmpresaID:               empresaID,
		Codigo:                  "BONOQA" + short(runID),
		Nombre:                  "Bono QA no salarial " + runID,
		Tipo:                    "devengado",
		BaseCotizacion:          false,
		AfectaPILA:              false,
		AfectaNominaElectronica: true,
		ValorFijo:               85000,
		CuentaContable:          "510548",
		Estado:                  "activo",
		UsuarioCreador:          usuario,
	})
	if err != nil {
		return 0, err
	}
	if _, err := dbpkg.CreateEmpresaNominaNovedadColombia(db, dbpkg.EmpresaNominaNovedadColombia{
		EmpresaID:        empresaID,
		EmpleadoNominaID: empleadoID,
		PeriodoDesde:     desde,
		PeriodoHasta:     hasta,
		FechaNovedad:     desde,
		Tipo:             "devengado",
		ConceptoID:       conceptoID,
		Descripcion:      "Novedad QA bono no salarial " + runID,
		Cantidad:         1,
		ValorUnitario:    85000,
		AfectaIBC:        false,
		EstadoAprobacion: "aprobado",
		Estado:           "activo",
		UsuarioCreador:   usuario,
	}); err != nil {
		return 0, err
	}
	if _, err := dbpkg.GenerateEmpresaNominaLiquidaciones(db, dbpkg.EmpresaNominaCalculoRequest{
		EmpresaID:        empresaID,
		PeriodoDesde:     desde,
		PeriodoHasta:     hasta,
		EmpleadoNominaID: empleadoID,
		Overwrite:        true,
		UsuarioCreador:   usuario,
		Observaciones:    "QA liquidacion nomina Colombia avanzada " + runID,
	}); err != nil {
		return 0, err
	}
	rows, err := dbpkg.GenerarEmpresaNominaPILAResumenColombia(db, empresaID, desde, hasta, usuario)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, fmt.Errorf("no se genero resumen PILA")
	}
	return empleadoID, nil
}

func writeReport(rep report) {
	path := reportPath()
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	raw, _ := json.MarshalIndent(rep, "", "  ")
	_ = os.WriteFile(path, raw, 0644)
}

func reportPath() string {
	return filepath.Join("tmp_tools", "qa_calipso_modulos_nuevos", "reporte_calipso_modulos_nuevos.json")
}

func must(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}

func digits(v string) string {
	out := strings.NewReplacer("-", "", ":", "", " ", "").Replace(v)
	if len(out) > 10 {
		return out[len(out)-10:]
	}
	return out
}

func short(v string) string {
	d := digits(v)
	if len(d) > 6 {
		return d[len(d)-6:]
	}
	return d
}
