package handlers

import (
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

type empresaVerticalIntegracionItem struct {
	ID                   string   `json:"id"`
	Modulo               string   `json:"module"`
	Page                 string   `json:"page"`
	Titulo               string   `json:"title"`
	IntegrationStatus    string   `json:"integration_status"`
	OperationalVisible   bool     `json:"operational_visible"`
	CoreModules          []string `json:"core_modules"`
	TemplateActivates    []string `json:"template_activates,omitempty"`
	TablesTouched        []string `json:"tables_touched,omitempty"`
	RequiredPermissions  []string `json:"required_permissions,omitempty"`
	SaleFlow             []string `json:"sale_flow,omitempty"`
	ReportsProduced      []string `json:"reports_produced,omitempty"`
	FinancialCoreModules []string `json:"financial_core_modules,omitempty"`
	IncomeFlow           []string `json:"income_flow,omitempty"`
	ExpenseFlow          []string `json:"expense_flow,omitempty"`
	FinancialTables      []string `json:"financial_tables,omitempty"`
	FinancialReports     []string `json:"financial_reports,omitempty"`
	DuplicatesCore       []string `json:"duplicates_core"`
	OwnFlowAllowed       []string `json:"own_flow_allowed"`
	Decision             string   `json:"decision"`
	AliasDe              string   `json:"alias_of,omitempty"`
	FusedModules         []string `json:"fused_modules,omitempty"`
	SupportModules       []string `json:"support_modules,omitempty"`
	SimilarTemplates     []string `json:"similar_templates,omitempty"`
	Motivo               string   `json:"reason"`
	ProfessionalReady    bool     `json:"professional_ready"`
	ReadinessScore       int      `json:"readiness_score"`
	ReadinessChecks      []string `json:"readiness_checks,omitempty"`
	ConfigurationScope   []string `json:"configuration_scope,omitempty"`
}

var empresaVerticalesCoreModules = []string{"clientes", "inventario", "ventas", "pagos", "finanzas", "facturacion", "reportes", "seguridad"}

type empresaVerticalIntegracionDetalle struct {
	TemplateActivates    []string
	TablesTouched        []string
	RequiredPermissions  []string
	SaleFlow             []string
	ReportsProduced      []string
	FinancialCoreModules []string
	IncomeFlow           []string
	ExpenseFlow          []string
	FinancialTables      []string
	FinancialReports     []string
}

func EmpresaVerticalesIntegracionCatalogoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		items := buildEmpresaVerticalesIntegracionCatalogo()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"total": len(items),
			"items": items,
		})
	}
}

func SuperVerticalesIntegracionCatalogoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		items := buildEmpresaVerticalesIntegracionCatalogo()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"total": len(items),
			"items": items,
		})
	}
}

func PublicVerticalesIntegracionCatalogoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		items := buildEmpresaVerticalesIntegracionCatalogo()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":    true,
			"total": len(items),
			"items": items,
		})
	}
}

func buildEmpresaVerticalesIntegracionCatalogo() []empresaVerticalIntegracionItem {
	items := []empresaVerticalIntegracionItem{
		withVerticalFusion(classicVertical("gimnasio", "linkGimnasio", "Gimnasio", "Socios, planes, acceso, clases, asistencias y dispositivos", "Plantilla fitness conectada al nucleo comun: socios, planes y pagos operan desde clientes, servicios, ventas y pagos centrales."), nil, []string{"estaciones", "turnos_atencion"}, []string{"club_deportivo"}),
		withVerticalFusion(classicVertical("odontologia", "linkConsultorioOdontologico", "Odontologia", "Historia clinica, odontograma, profesionales, consultorios y citas clinicas", "Plantilla clinica conectada al nucleo comun: pacientes, tratamientos y recaudos usan clientes, servicios, ventas y pagos centrales."), []string{"consultorio_odontologico"}, []string{"turnos_atencion", "estaciones"}, []string{"clinica_consultorios"}),
		withVerticalFusion(classicVertical("parqueadero", "linkParqueadero", "Parqueadero", "Ticket QR, placa, entrada/salida, tiempos y reglas tarifarias", "Plantilla de parqueadero conectada al nucleo comun: tickets y cobros crean servicio, venta y pago central sin modulo comercial paralelo."), nil, []string{"estaciones", "turnos_atencion"}, []string{"parque_recreativo"}),
		withVerticalFusion(classicVertical("taxi_system", "linkTaxiSystem", "Taxi system", "Conductores, despacho, GPS, tracking, ofertas y rutas", "Plantilla de transporte conectada al nucleo comun: clientes, servicios de viaje, ventas y pagos se gobiernan desde el nucleo."), []string{"taxi"}, []string{"estaciones"}, []string{"transporte_carga_tms"}),
		classicVertical("domicilios", "linkDomicilios", "Domicilios", "Tracking, domiciliarios, restaurantes aliados, menu, ofertas y estados logisticos", "Plantilla logistica conectada al nucleo comun: pedidos, clientes, menu, ventas y pagos se resuelven en los modulos centrales."),
		classicVertical("apartamentos_turisticos", "linkApartamentosTuristicos", "Apartamentos turisticos", "Unidades, disponibilidad, tareas, tarifas, check-in y checkout", "Plantilla de alojamiento conectada al nucleo comun: huespedes, unidades vendibles, reservas, ventas y pagos comparten el motor central."),
		classicVertical("propiedad_horizontal", "linkPropiedadHorizontal", "Propiedad horizontal", "Unidades, asambleas, PQR, residentes, cartera y recaudos", "Plantilla de copropiedad conectada al nucleo comun: terceros, unidades, cargos, recaudos, cartera y reportes no duplican clientes ni pagos."),
		classicVertical("alquileres", "linkAlquileres", "Alquileres", "Contratos, activos, garantias, mantenimientos, kilometraje y mapa GPS", "Plantilla de alquiler conectada al nucleo comun: clientes, activos vendibles, contratos, ventas y pagos usan la fuente unica."),
		classicVertical("drogueria_farmacia", "linkDrogueriaFarmacia", "Drogueria / farmacia", "Expediente sanitario, lotes, INVIMA, formulas, controlados y farmacovigilancia", "Plantilla sanitaria conectada al nucleo comun: productos, inventario, compras, clientes, ventas y facturacion siguen en modulos centrales."),
		classicVertical("aiu_construccion", "linkAIUConstruccion", "Construccion / AIU", "Capitulos, AIU, presupuestos de obra, retenciones, anticipo, garantia y auditoria tecnica", "Plantilla de construccion conectada al nucleo comun: clientes, contratos, conceptos, ventas, impuestos y reportes se enlazan sin duplicar documentos comerciales."),
	}
	items = append(items, nuevosVerticalesIntegracionItems()...)
	for idx := range items {
		items[idx] = enrichEmpresaVerticalReadiness(items[idx])
	}
	return items
}

func withVerticalFusion(item empresaVerticalIntegracionItem, fused, support, similar []string) empresaVerticalIntegracionItem {
	item.FusedModules = normalizedStringSlice(fused)
	item.SupportModules = normalizedStringSlice(support)
	item.SimilarTemplates = normalizedStringSlice(similar)
	return item
}

func normalizedStringSlice(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		clean := strings.ToLower(strings.TrimSpace(value))
		if clean == "" || seen[clean] {
			continue
		}
		seen[clean] = true
		out = append(out, clean)
	}
	return out
}

func classicVertical(module, page, title, ownFlow, reason string) empresaVerticalIntegracionItem {
	detail := classicVerticalIntegrationDetail(module, page)
	return empresaVerticalIntegracionItem{
		ID:                   strings.TrimSpace(page),
		Modulo:               strings.ToLower(strings.TrimSpace(module)),
		Page:                 strings.TrimSpace(page),
		Titulo:               strings.TrimSpace(title),
		IntegrationStatus:    "plantilla_integrada_nucleo",
		OperationalVisible:   true,
		CoreModules:          append([]string{}, empresaVerticalesCoreModules...),
		TemplateActivates:    copyStringSlice(detail.TemplateActivates),
		TablesTouched:        copyStringSlice(detail.TablesTouched),
		RequiredPermissions:  copyStringSlice(detail.RequiredPermissions),
		SaleFlow:             copyStringSlice(detail.SaleFlow),
		ReportsProduced:      copyStringSlice(detail.ReportsProduced),
		FinancialCoreModules: copyStringSlice(detail.FinancialCoreModules),
		IncomeFlow:           copyStringSlice(detail.IncomeFlow),
		ExpenseFlow:          copyStringSlice(detail.ExpenseFlow),
		FinancialTables:      copyStringSlice(detail.FinancialTables),
		FinancialReports:     copyStringSlice(detail.FinancialReports),
		DuplicatesCore:       []string{},
		OwnFlowAllowed:       []string{strings.TrimSpace(ownFlow)},
		Decision:             "plantilla_universal_nucleo",
		Motivo:               strings.TrimSpace(reason),
	}
}

func nuevosVerticalesIntegracionItems() []empresaVerticalIntegracionItem {
	catalog := dbpkg.NuevosVerticalesTipoEmpresaCatalog()
	out := make([]empresaVerticalIntegracionItem, 0, len(catalog))
	for _, item := range catalog {
		modulo := strings.ToLower(strings.TrimSpace(item.Modulo))
		if modulo == "" {
			continue
		}
		plantilla := dbpkg.GetEmpresaModuloColombiaPlantilla(modulo)
		integracion := dbpkg.BuildTipoEmpresaPreconfigIntegracionVertical(modulo)
		if integracion == nil {
			continue
		}
		page := nuevoVerticalPageKey(modulo)
		title := strings.TrimSpace(firstNonEmptyString(item.Nombre, plantilla.Titulo, modulo))
		reason := strings.TrimSpace(integracion.MotivoDecision)
		if reason == "" {
			reason = "Plantilla vertical real conectada al nucleo comun sin duplicar clientes, productos, ventas, pagos ni reportes."
		}
		out = append(out, empresaVerticalIntegracionItem{
			ID:                   page,
			Modulo:               modulo,
			Page:                 page,
			Titulo:               title,
			IntegrationStatus:    strings.TrimSpace(integracion.EstadoIntegracion),
			OperationalVisible:   true,
			CoreModules:          append([]string{}, empresaVerticalesCoreModules...),
			TemplateActivates:    copyStringSlice(integracion.TemplateActivates),
			TablesTouched:        copyStringSlice(integracion.TablesTouched),
			RequiredPermissions:  copyStringSlice(integracion.RequiredPermissions),
			SaleFlow:             copyStringSlice(integracion.SaleFlow),
			ReportsProduced:      copyStringSlice(integracion.ReportsProduced),
			FinancialCoreModules: copyStringSlice(integracion.FinancialCoreModules),
			IncomeFlow:           copyStringSlice(integracion.IncomeFlow),
			ExpenseFlow:          copyStringSlice(integracion.ExpenseFlow),
			FinancialTables:      copyStringSlice(integracion.FinancialTables),
			FinancialReports:     copyStringSlice(integracion.FinancialReports),
			DuplicatesCore:       []string{},
			OwnFlowAllowed:       copyStringSlice(plantilla.SeccionesFlujo),
			Decision:             strings.TrimSpace(integracion.Decision),
			Motivo:               reason,
			ConfigurationScope:   []string{"tipo_empresa_preconfiguracion", "licencia", "roles", "menu", "datos_guia", "reportes"},
		})
	}
	return out
}

func enrichEmpresaVerticalReadiness(item empresaVerticalIntegracionItem) empresaVerticalIntegracionItem {
	checks := make([]string, 0, 8)
	total := 0
	ok := 0
	add := func(name string, passed bool) {
		total++
		if passed {
			ok++
			checks = append(checks, name)
		}
	}

	status := strings.ToLower(strings.TrimSpace(item.IntegrationStatus))
	add("visible_operativo", item.OperationalVisible)
	add("sin_duplicados_del_nucleo", len(item.DuplicatesCore) == 0)
	add("plantilla_de_configuracion", len(item.TemplateActivates) > 0)
	add("tablas_y_datos_declarados", len(item.TablesTouched) > 0)
	add("permisos_declarados", len(item.RequiredPermissions) > 0)
	add("flujo_de_venta_o_operacion", len(item.SaleFlow) > 0)
	add("reportes_declarados", len(item.ReportsProduced) > 0)
	add("nucleo_financiero_declarado", hasAllStringValues(item.FinancialCoreModules, []string{"finanzas", "ventas", "pagos", "reportes"}))
	add("ingresos_del_nucleo_declarados", len(item.IncomeFlow) > 0)
	add("egresos_del_nucleo_declarados", len(item.ExpenseFlow) > 0)
	add("tablas_financieras_declaradas", hasAllStringValues(item.FinancialTables, []string{"empresa_finanzas_movimientos"}))
	add("reportes_financieros_declarados", len(item.FinancialReports) > 0)
	if status == "integrado_soporte" {
		add("nucleo_soporte_reportes_seguridad", hasAllStringValues(item.CoreModules, []string{"seguridad", "reportes"}))
	} else {
		add("nucleo_comercial_completo", hasAllStringValues(item.CoreModules, []string{"clientes", "inventario", "ventas", "pagos", "finanzas", "reportes"}))
	}

	score := 0
	if total > 0 {
		score = (ok * 100) / total
	}
	item.ReadinessScore = score
	item.ProfessionalReady = score == 100
	item.ReadinessChecks = checks
	if len(item.ConfigurationScope) == 0 {
		item.ConfigurationScope = []string{"tipo_empresa", "licencia", "roles", "menu", "datos_guia", "reportes"}
	}
	return item
}

func hasAllStringValues(values []string, required []string) bool {
	seen := map[string]bool{}
	for _, value := range values {
		clean := strings.ToLower(strings.TrimSpace(value))
		if clean != "" {
			seen[clean] = true
		}
	}
	for _, value := range required {
		if !seen[strings.ToLower(strings.TrimSpace(value))] {
			return false
		}
	}
	return true
}

func classicVerticalIntegrationDetail(module, page string) empresaVerticalIntegracionDetalle {
	module = strings.ToLower(strings.TrimSpace(module))
	page = strings.TrimSpace(page)
	baseTables := []string{"clientes", "servicios", "carritos_compras", "carrito_compra_items", "empresa_finanzas_movimientos"}
	baseReports := []string{"reporte operativo del vertical", "ventas por servicio", "ingresos por periodo", "egresos por periodo", "auditoria por empresa"}
	basePermissions := []string{
		"seguridad:R",
		module + ":R",
		module + ":C",
		"clientes:R/C",
		"inventario:R/C servicios",
		"ventas:C",
		"pagos:C",
		"finanzas:R/C",
		"reportes:R",
	}
	d := empresaVerticalIntegracionDetalle{
		TemplateActivates:    []string{module, page, "clientes", "inventario/servicios", "ventas", "pagos", "finanzas", "reportes"},
		TablesTouched:        append([]string{}, baseTables...),
		RequiredPermissions:  basePermissions,
		ReportsProduced:      baseReports,
		SaleFlow:             []string{"registro especializado", "cliente/servicio central", "carrito central", "pago o factura central", "ingreso conciliable en finanzas", "reporte consolidado"},
		FinancialCoreModules: []string{"ventas", "pagos", "finanzas", "bancos_pagos", "tesoreria_presupuesto", "reportes"},
		IncomeFlow:           []string{"servicio/producto vendible del vertical", "carrito o venta central", "pago central", "movimiento ingreso en empresa_finanzas_movimientos", "reporte financiero consolidado"},
		ExpenseFlow:          []string{"compra/gasto operativo del vertical", "soporte o documento central", "movimiento egreso en empresa_finanzas_movimientos", "conciliacion bancaria/tesoreria", "reporte financiero consolidado"},
		FinancialTables:      []string{"carritos_compras", "carrito_compra_items", "empresa_finanzas_movimientos", "empresa_finanzas_configuracion", "empresa_finanzas_periodos"},
		FinancialReports:     []string{"ingresos por vertical", "egresos por vertical", "margen operativo", "flujo de caja", "estado de resultados por empresa"},
	}
	switch module {
	case "gimnasio":
		d.TablesTouched = append(d.TablesTouched, "empresa_gimnasio_socios", "empresa_gimnasio_planes", "empresa_gimnasio_pagos", "empresa_gimnasio_clases", "empresa_gimnasio_asistencias", "empresa_gimnasio_eventos_acceso")
		d.SaleFlow = []string{"socio/plan", "cliente y servicio central", "pago de membresia", "carrito con item de servicio", "pago central conciliable"}
		d.ReportsProduced = []string{"socios activos", "asistencias", "ingresos por plan", "egresos por entrenador/insumo", "vencimientos de membresia", "ventas centrales por servicio"}
	case "odontologia", "consultorio_odontologico":
		d.TablesTouched = append(d.TablesTouched, "empresa_odontologia_pacientes", "empresa_odontologia_tratamientos", "empresa_odontologia_pagos", "empresa_odontologia_historias", "empresa_odontologia_odontogramas", "empresa_odontologia_citas")
		d.SaleFlow = []string{"paciente/tratamiento", "cliente y servicio central", "pago aplicado", "carrito con item clinico", "venta/pago central"}
		d.ReportsProduced = []string{"agenda clinica", "tratamientos facturables", "pagos por paciente", "egresos clinicos", "ventas centrales por servicio", "historia clinica no comercial"}
	case "parqueadero":
		d.TablesTouched = append(d.TablesTouched, "empresa_parqueadero_config", "empresa_parqueadero_tickets")
		d.SaleFlow = []string{"ticket de entrada", "tarifa por tipo de vehiculo", "cobro de salida", "carrito con servicio de parqueo", "pago central con referencia de ticket"}
		d.ReportsProduced = []string{"tickets cerrados", "ocupacion", "ingresos por tipo de vehiculo", "egresos operativos", "anulaciones", "ventas centrales por servicio"}
	case "taxi_system", "taxi":
		d.TablesTouched = append(d.TablesTouched, "empresa_taxi_customers", "empresa_taxi_drivers", "empresa_taxi_requests", "empresa_taxi_offers", "empresa_taxi_route_points")
		d.SaleFlow = []string{"solicitud de viaje", "cliente y servicio central", "viaje completado", "carrito con servicio de transporte", "pago central por metodo"}
		d.ReportsProduced = []string{"viajes completados", "conductores online", "ofertas y despacho", "rutas GPS", "ingresos y egresos de transporte", "ventas centrales de transporte"}
	case "domicilios":
		d.TablesTouched = append(d.TablesTouched, "empresa_domicilios_restaurantes", "empresa_domicilios_menu_items", "empresa_domicilios_orders", "empresa_domicilios_order_items", "empresa_domicilios_tracking", "empresa_domicilios_couriers")
		d.SaleFlow = []string{"pedido entregado", "cliente y menu como servicios", "items, domicilio y propina", "carrito central", "pago central normalizado"}
		d.ReportsProduced = []string{"pedidos entregados", "ventas por restaurante/menu", "tracking y couriers", "tarifas de entrega", "egresos logisticos", "ventas centrales por servicio"}
	case "apartamentos_turisticos":
		d.TablesTouched = append(d.TablesTouched, "empresa_apartamentos_turisticos_unidades", "empresa_apartamentos_turisticos_tarifas", "empresa_apartamentos_turisticos_reservas", "empresa_apartamentos_turisticos_tareas")
		d.SaleFlow = []string{"reserva check-in/checkout", "huesped como cliente", "unidad como servicio", "carrito con alojamiento/limpieza/impuestos", "pago central"}
		d.ReportsProduced = []string{"ocupacion", "reservas cerradas", "ingresos por unidad", "egresos por limpieza/mantenimiento", "tareas operativas", "ventas centrales de alojamiento"}
	case "propiedad_horizontal":
		d.TablesTouched = append(d.TablesTouched, "empresa_propiedad_horizontal_personas", "empresa_propiedad_horizontal_unidades", "empresa_propiedad_horizontal_cargos", "empresa_propiedad_horizontal_recaudos", "empresa_propiedad_horizontal_pqrs", "empresa_propiedad_horizontal_asambleas")
		d.SaleFlow = []string{"cargo/recaudo", "propietario o residente como cliente", "unidad/cargo como servicio", "carrito central", "pago central de recaudo"}
		d.ReportsProduced = []string{"cartera por unidad", "recaudos", "egresos de administracion", "PQR", "asambleas", "ventas centrales por cargo"}
	case "alquileres":
		d.TablesTouched = append(d.TablesTouched, "empresa_alquileres_activos", "empresa_alquileres_tarifas", "empresa_alquileres_contratos", "empresa_alquileres_mantenimientos", "empresa_alquileres_ubicaciones")
		d.SaleFlow = []string{"contrato de alquiler", "cliente central", "activo/tarifa como servicio", "carrito central del contrato", "pago central al cerrar saldo"}
		d.ReportsProduced = []string{"activos disponibles", "contratos", "garantias", "mantenimientos", "ingresos y egresos por activo", "ventas centrales por activo"}
	case "drogueria_farmacia":
		d.TemplateActivates = []string{module, page, "inventario", "compras", "clientes", "ventas", "pagos", "finanzas", "facturacion", "reportes"}
		d.TablesTouched = []string{"empresa_modulos_colombia_registros", "empresa_modulos_colombia_eventos", "productos", "inventario", "clientes", "carritos_compras", "carrito_compra_items", "empresa_finanzas_movimientos"}
		d.SaleFlow = []string{"producto central con expediente sanitario", "inventario central", "venta central", "facturacion central", "trazabilidad sanitaria por registro"}
		d.ReportsProduced = []string{"lotes/INVIMA/controlados", "dispensacion", "devoluciones", "farmacovigilancia", "ingresos y egresos de farmacia", "ventas e inventario central"}
	case "aiu_construccion":
		d.TablesTouched = append(d.TablesTouched, "empresa_aiu_contratos", "empresa_aiu_items", "empresa_aiu_facturas", "empresa_aiu_eventos")
		d.SaleFlow = []string{"contrato AIU", "cliente y contrato como servicio", "conceptos como servicios", "factura AIU enlazada a carrito", "facturacion central sin recalcular impuestos"}
		d.ReportsProduced = []string{"contratos por estado", "capitulos/conceptos", "avance y riesgo", "facturas AIU", "ingresos y egresos por obra", "ventas centrales enlazadas"}
	}
	return d
}

func copyStringSlice(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}
