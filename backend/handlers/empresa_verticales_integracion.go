package handlers

import (
	"net/http"
	"strings"
)

type empresaVerticalIntegracionItem struct {
	ID                  string   `json:"id"`
	Modulo              string   `json:"module"`
	Page                string   `json:"page"`
	Titulo              string   `json:"title"`
	IntegrationStatus   string   `json:"integration_status"`
	OperationalVisible  bool     `json:"operational_visible"`
	CoreModules         []string `json:"core_modules"`
	TemplateActivates   []string `json:"template_activates,omitempty"`
	TablesTouched       []string `json:"tables_touched,omitempty"`
	RequiredPermissions []string `json:"required_permissions,omitempty"`
	SaleFlow            []string `json:"sale_flow,omitempty"`
	ReportsProduced     []string `json:"reports_produced,omitempty"`
	DuplicatesCore      []string `json:"duplicates_core"`
	OwnFlowAllowed      []string `json:"own_flow_allowed"`
	Decision            string   `json:"decision"`
	SyncAction          string   `json:"sync_action,omitempty"`
	SyncPath            string   `json:"sync_path,omitempty"`
	SyncActionName      string   `json:"sync_action_name,omitempty"`
	AliasDe             string   `json:"alias_of,omitempty"`
	Motivo              string   `json:"reason"`
}

var empresaVerticalesCoreModules = []string{"clientes", "inventario", "ventas", "pagos", "facturacion", "reportes", "seguridad"}

type empresaVerticalIntegracionDetalle struct {
	TemplateActivates   []string
	TablesTouched       []string
	RequiredPermissions []string
	SaleFlow            []string
	ReportsProduced     []string
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
	return []empresaVerticalIntegracionItem{
		classicVertical("gimnasio", "linkGimnasio", "Gimnasio", "Socios, planes, acceso, clases, asistencias y dispositivos", "Sincroniza socios con clientes, planes con servicios vendibles y pagos con ventas/pagos centrales.", "/api/empresa/gimnasio", "sincronizar_nucleo"),
		classicVertical("odontologia", "linkConsultorioOdontologico", "Odontologia", "Historia clinica, odontograma, profesionales, consultorios y citas clinicas", "Sincroniza pacientes con clientes, tratamientos con servicios vendibles y pagos con ventas/pagos centrales.", "/api/empresa/odontologia", "sincronizar_nucleo"),
		classicVerticalAlias("consultorio_odontologico", "linkConsultorioOdontologico", "Consultorio odontologico", "odontologia", "Vista especializada de odontologia integrada al nucleo operativo."),
		classicVertical("parqueadero", "linkParqueadero", "Parqueadero", "Ticket QR, placa, entrada/salida, tiempos y reglas tarifarias", "Cierra tickets creando venta, item de servicio y pago central en carritos.", "/api/empresa/parqueadero", "sincronizar_nucleo"),
		classicVertical("taxi_system", "linkTaxiSystem", "Taxi system", "Conductores, despacho, GPS, tracking, ofertas y rutas", "Sincroniza clientes de viaje y servicios completados con clientes, servicios, ventas y pagos centrales.", "/api/empresa/taxi_system", "sincronizar_nucleo"),
		classicVerticalAlias("taxi", "linkTaxiSystem", "Taxi", "taxi_system", "Alias visual de taxi_system integrado al nucleo."),
		classicVertical("domicilios", "linkDomicilios", "Domicilios", "Tracking, domiciliarios, restaurantes aliados, menu, ofertas y estados logisticos", "Sincroniza pedidos entregados con clientes, servicios de menu, ventas y pagos centrales.", "/api/empresa/domicilios", "sincronizar_nucleo"),
		classicVertical("apartamentos_turisticos", "linkApartamentosTuristicos", "Apartamentos turisticos", "Unidades, disponibilidad, tareas, tarifas, check-in y checkout", "Sincroniza huespedes con clientes, apartamentos con servicios y reservas cerradas con ventas/pagos centrales.", "/api/empresa/apartamentos_turisticos", "sincronizar_nucleo"),
		classicVertical("propiedad_horizontal", "linkPropiedadHorizontal", "Propiedad horizontal", "Unidades, asambleas, PQR, residentes, cartera y recaudos", "Sincroniza propietarios/residentes con clientes, cargos con servicios y recaudos con ventas/pagos centrales.", "/api/empresa/propiedad_horizontal", "sincronizar_nucleo"),
		classicVertical("alquileres", "linkAlquileres", "Alquileres", "Contratos, activos, garantias, mantenimientos, kilometraje y mapa GPS", "Sincroniza clientes de contratos con clientes centrales, activos/tarifas con servicios y contratos con ventas/pagos centrales.", "/api/empresa/alquileres", "sincronizar_nucleo"),
		classicVertical("drogueria_farmacia", "linkDrogueriaFarmacia", "Drogueria / farmacia", "Expediente sanitario, lotes, INVIMA, formulas, controlados y farmacovigilancia", "Opera como expediente sanitario sobre el nucleo: productos, inventario, ventas, clientes y facturacion siguen en modulos centrales.", "", ""),
		classicVertical("aiu_construccion", "linkAIUConstruccion", "Construccion / AIU", "Capitulos, AIU, presupuestos de obra, retenciones, anticipo, garantia y auditoria tecnica", "Sincroniza clientes de obra, contratos y conceptos como servicios; facturas AIU quedan enlazadas a ventas centrales sin duplicar impuestos.", "/api/empresa/aiu_construccion", "sincronizar_nucleo"),
		supportVertical("turnos_atencion", "linkTurnosAtencion", "Turnos de atencion", "Turnos, puestos, pantalla publica y seguimiento de fila", "Funciona como capacidad operativa transversal y no reemplaza clientes, productos, ventas ni pagos."),
		supportVerticalAlias("turnos", "linkTurnosAtencion", "Turnos", "turnos_atencion", "Alias visual de turnos_atencion."),
	}
}

func classicVertical(module, page, title, ownFlow, reason, syncPath, syncActionName string) empresaVerticalIntegracionItem {
	syncPath = strings.TrimSpace(syncPath)
	syncActionName = strings.TrimSpace(syncActionName)
	detail := classicVerticalIntegrationDetail(module, page)
	syncAction := ""
	if syncPath != "" && syncActionName != "" {
		syncAction = "POST " + syncPath + "?action=" + syncActionName
	}
	return empresaVerticalIntegracionItem{
		ID:                  strings.TrimSpace(page),
		Modulo:              strings.ToLower(strings.TrimSpace(module)),
		Page:                strings.TrimSpace(page),
		Titulo:              strings.TrimSpace(title),
		IntegrationStatus:   "plantilla_integrada_nucleo",
		OperationalVisible:  true,
		CoreModules:         append([]string{}, empresaVerticalesCoreModules...),
		TemplateActivates:   copyStringSlice(detail.TemplateActivates),
		TablesTouched:       copyStringSlice(detail.TablesTouched),
		RequiredPermissions: copyStringSlice(detail.RequiredPermissions),
		SaleFlow:            copyStringSlice(detail.SaleFlow),
		ReportsProduced:     copyStringSlice(detail.ReportsProduced),
		DuplicatesCore:      []string{},
		OwnFlowAllowed:      []string{strings.TrimSpace(ownFlow)},
		Decision:            "mantener_visible",
		SyncAction:          syncAction,
		SyncPath:            syncPath,
		SyncActionName:      syncActionName,
		Motivo:              strings.TrimSpace(reason),
	}
}

func classicVerticalAlias(module, page, title, aliasOf, reason string) empresaVerticalIntegracionItem {
	item := classicVertical(module, page, title, "Alias operativo; usa el flujo permitido del modulo principal", reason, "", "")
	detail := classicVerticalIntegrationDetail(aliasOf, page)
	item.TemplateActivates = copyStringSlice(detail.TemplateActivates)
	item.TablesTouched = copyStringSlice(detail.TablesTouched)
	item.RequiredPermissions = copyStringSlice(detail.RequiredPermissions)
	item.SaleFlow = copyStringSlice(detail.SaleFlow)
	item.ReportsProduced = copyStringSlice(detail.ReportsProduced)
	item.AliasDe = strings.ToLower(strings.TrimSpace(aliasOf))
	return item
}

func supportVertical(module, page, title, ownFlow, reason string) empresaVerticalIntegracionItem {
	detail := supportVerticalIntegrationDetail(module, page)
	return empresaVerticalIntegracionItem{
		ID:                  strings.TrimSpace(page),
		Modulo:              strings.ToLower(strings.TrimSpace(module)),
		Page:                strings.TrimSpace(page),
		Titulo:              strings.TrimSpace(title),
		IntegrationStatus:   "integrado_soporte",
		OperationalVisible:  true,
		CoreModules:         []string{"seguridad", "reportes", "operacion"},
		TemplateActivates:   copyStringSlice(detail.TemplateActivates),
		TablesTouched:       copyStringSlice(detail.TablesTouched),
		RequiredPermissions: copyStringSlice(detail.RequiredPermissions),
		SaleFlow:            copyStringSlice(detail.SaleFlow),
		ReportsProduced:     copyStringSlice(detail.ReportsProduced),
		DuplicatesCore:      []string{},
		OwnFlowAllowed:      []string{strings.TrimSpace(ownFlow)},
		Decision:            "mantener_visible",
		Motivo:              strings.TrimSpace(reason),
	}
}

func supportVerticalAlias(module, page, title, aliasOf, reason string) empresaVerticalIntegracionItem {
	item := supportVertical(module, page, title, "Alias operativo; usa el flujo permitido del modulo principal", reason)
	detail := supportVerticalIntegrationDetail(aliasOf, page)
	item.TemplateActivates = copyStringSlice(detail.TemplateActivates)
	item.TablesTouched = copyStringSlice(detail.TablesTouched)
	item.RequiredPermissions = copyStringSlice(detail.RequiredPermissions)
	item.SaleFlow = copyStringSlice(detail.SaleFlow)
	item.ReportsProduced = copyStringSlice(detail.ReportsProduced)
	item.AliasDe = strings.ToLower(strings.TrimSpace(aliasOf))
	return item
}

func classicVerticalIntegrationDetail(module, page string) empresaVerticalIntegracionDetalle {
	module = strings.ToLower(strings.TrimSpace(module))
	page = strings.TrimSpace(page)
	baseTables := []string{"clientes", "servicios", "carritos_compras", "carrito_compra_items"}
	baseReports := []string{"reporte operativo del vertical", "ventas por servicio", "ingresos por periodo", "auditoria por empresa"}
	basePermissions := []string{
		"seguridad:R",
		module + ":R",
		module + ":C",
		"clientes:R/C",
		"inventario:R/C servicios",
		"ventas:C",
		"pagos:C",
		"reportes:R",
	}
	d := empresaVerticalIntegracionDetalle{
		TemplateActivates:   []string{module, page, "clientes", "inventario/servicios", "ventas", "pagos", "reportes"},
		TablesTouched:       append([]string{}, baseTables...),
		RequiredPermissions: basePermissions,
		ReportsProduced:     baseReports,
		SaleFlow:            []string{"registro especializado", "cliente/servicio central", "carrito central", "pago o factura central", "reporte consolidado"},
	}
	switch module {
	case "gimnasio":
		d.TablesTouched = append(d.TablesTouched, "empresa_gimnasio_socios", "empresa_gimnasio_planes", "empresa_gimnasio_pagos", "empresa_gimnasio_clases", "empresa_gimnasio_asistencias", "empresa_gimnasio_eventos_acceso")
		d.SaleFlow = []string{"socio/plan", "cliente y servicio central", "pago de membresia", "carrito con item de servicio", "pago central conciliable"}
		d.ReportsProduced = []string{"socios activos", "asistencias", "ingresos por plan", "vencimientos de membresia", "ventas centrales por servicio"}
	case "odontologia", "consultorio_odontologico":
		d.TablesTouched = append(d.TablesTouched, "empresa_odontologia_pacientes", "empresa_odontologia_tratamientos", "empresa_odontologia_pagos", "empresa_odontologia_historias", "empresa_odontologia_odontogramas", "empresa_odontologia_citas")
		d.SaleFlow = []string{"paciente/tratamiento", "cliente y servicio central", "pago aplicado", "carrito con item clinico", "venta/pago central"}
		d.ReportsProduced = []string{"agenda clinica", "tratamientos facturables", "pagos por paciente", "ventas centrales por servicio", "historia clinica no comercial"}
	case "parqueadero":
		d.TablesTouched = append(d.TablesTouched, "empresa_parqueadero_config", "empresa_parqueadero_tickets")
		d.SaleFlow = []string{"ticket de entrada", "tarifa por tipo de vehiculo", "cobro de salida", "carrito con servicio de parqueo", "pago central con referencia de ticket"}
		d.ReportsProduced = []string{"tickets cerrados", "ocupacion", "ingresos por tipo de vehiculo", "anulaciones", "ventas centrales por servicio"}
	case "taxi_system", "taxi":
		d.TablesTouched = append(d.TablesTouched, "empresa_taxi_customers", "empresa_taxi_drivers", "empresa_taxi_requests", "empresa_taxi_offers", "empresa_taxi_route_points")
		d.SaleFlow = []string{"solicitud de viaje", "cliente y servicio central", "viaje completado", "carrito con servicio de transporte", "pago central por metodo"}
		d.ReportsProduced = []string{"viajes completados", "conductores online", "ofertas y despacho", "rutas GPS", "ventas centrales de transporte"}
	case "domicilios":
		d.TablesTouched = append(d.TablesTouched, "empresa_domicilios_restaurantes", "empresa_domicilios_menu_items", "empresa_domicilios_orders", "empresa_domicilios_order_items", "empresa_domicilios_tracking", "empresa_domicilios_couriers")
		d.SaleFlow = []string{"pedido entregado", "cliente y menu como servicios", "items, domicilio y propina", "carrito central", "pago central normalizado"}
		d.ReportsProduced = []string{"pedidos entregados", "ventas por restaurante/menu", "tracking y couriers", "tarifas de entrega", "ventas centrales por servicio"}
	case "apartamentos_turisticos":
		d.TablesTouched = append(d.TablesTouched, "empresa_apartamentos_turisticos_unidades", "empresa_apartamentos_turisticos_tarifas", "empresa_apartamentos_turisticos_reservas", "empresa_apartamentos_turisticos_tareas")
		d.SaleFlow = []string{"reserva check-in/checkout", "huesped como cliente", "unidad como servicio", "carrito con alojamiento/limpieza/impuestos", "pago central"}
		d.ReportsProduced = []string{"ocupacion", "reservas cerradas", "ingresos por unidad", "tareas operativas", "ventas centrales de alojamiento"}
	case "propiedad_horizontal":
		d.TablesTouched = append(d.TablesTouched, "empresa_propiedad_horizontal_personas", "empresa_propiedad_horizontal_unidades", "empresa_propiedad_horizontal_cargos", "empresa_propiedad_horizontal_recaudos", "empresa_propiedad_horizontal_pqrs", "empresa_propiedad_horizontal_asambleas")
		d.SaleFlow = []string{"cargo/recaudo", "propietario o residente como cliente", "unidad/cargo como servicio", "carrito central", "pago central de recaudo"}
		d.ReportsProduced = []string{"cartera por unidad", "recaudos", "PQR", "asambleas", "ventas centrales por cargo"}
	case "alquileres":
		d.TablesTouched = append(d.TablesTouched, "empresa_alquileres_activos", "empresa_alquileres_tarifas", "empresa_alquileres_contratos", "empresa_alquileres_mantenimientos", "empresa_alquileres_ubicaciones")
		d.SaleFlow = []string{"contrato de alquiler", "cliente central", "activo/tarifa como servicio", "carrito central del contrato", "pago central al cerrar saldo"}
		d.ReportsProduced = []string{"activos disponibles", "contratos", "garantias", "mantenimientos", "ventas centrales por activo"}
	case "drogueria_farmacia":
		d.TemplateActivates = []string{module, page, "inventario", "compras", "clientes", "ventas", "facturacion", "reportes"}
		d.TablesTouched = []string{"empresa_modulos_colombia_registros", "empresa_modulos_colombia_eventos", "productos", "inventario", "clientes", "carritos_compras", "carrito_compra_items"}
		d.SaleFlow = []string{"producto central con expediente sanitario", "inventario central", "venta central", "facturacion central", "trazabilidad sanitaria por registro"}
		d.ReportsProduced = []string{"lotes/INVIMA/controlados", "dispensacion", "devoluciones", "farmacovigilancia", "ventas e inventario central"}
	case "aiu_construccion":
		d.TablesTouched = append(d.TablesTouched, "empresa_aiu_contratos", "empresa_aiu_items", "empresa_aiu_facturas", "empresa_aiu_eventos")
		d.SaleFlow = []string{"contrato AIU", "cliente y contrato como servicio", "conceptos como servicios", "factura AIU enlazada a carrito", "facturacion central sin recalcular impuestos"}
		d.ReportsProduced = []string{"contratos por estado", "capitulos/conceptos", "avance y riesgo", "facturas AIU", "ventas centrales enlazadas"}
	}
	return d
}

func supportVerticalIntegrationDetail(module, page string) empresaVerticalIntegracionDetalle {
	module = strings.ToLower(strings.TrimSpace(module))
	return empresaVerticalIntegracionDetalle{
		TemplateActivates:   []string{module, strings.TrimSpace(page), "seguridad", "reportes", "operacion"},
		TablesTouched:       []string{"empresa_turnos_atencion_config", "empresa_turnos_atencion_servicios", "empresa_turnos_atencion_puestos", "empresa_turnos_atencion_tickets"},
		RequiredPermissions: []string{"seguridad:R", module + ":R", module + ":C", "reportes:R"},
		SaleFlow:            []string{"turno o fila operativa", "atencion por puesto", "evento de seguimiento", "reporte operativo", "sin venta ni pago duplicado"},
		ReportsProduced:     []string{"turnos atendidos", "tiempos de espera", "puestos activos", "eventos de fila"},
	}
}

func copyStringSlice(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}
