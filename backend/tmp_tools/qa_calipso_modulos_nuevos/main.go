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
		FechaCompra: today, Costo: 2400000, ValorResidual: 200000, VidaUtilMeses: 48, CuentaActivo: "1528", CuentaDepreciacion: "1592", CuentaGasto: "5160",
		Ubicacion: "Recepcion Motel Calipso", Responsable: "QA Administracion", Estado: "activo", UsuarioCreador: *usuario,
	})
	mustStep("contabilidad_avanzada", "crear_activo_fijo", activoID, nil, err)
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

	add(step{Module: "resumen", Action: "conteos_empresa_7", OK: true, Meta: map[string]interface{}{
		"contabilidad_comprobantes": countRows("empresa_contabilidad_colombia_comprobantes"),
		"domicilios_orders":         countRows("empresa_domicilios_orders"),
		"parqueadero_tickets":       countRows("empresa_parqueadero_tickets"),
		"apartamentos_reservas":     countRows("empresa_apartamentos_turisticos_reservas"),
		"carnets":                   countRows("empresa_carnets"),
		"taxi_requests":             countRows("empresa_taxi_requests"),
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
