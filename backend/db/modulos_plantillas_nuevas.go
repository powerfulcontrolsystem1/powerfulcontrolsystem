package db

func init() {
	for modulo, plantilla := range empresaModuloColombiaPlantillasPlantillas {
		empresaModuloColombiaTitulos[modulo] = plantilla.Titulo
	}
}

func buildVerticalPlantilla(titulo string, tipos, categorias, estados, acciones []string, etiquetaTercero, etiquetaReferencia, metadata string) EmpresaModuloColombiaPlantilla {
	return EmpresaModuloColombiaPlantilla{
		Titulo:             titulo,
		Tipos:              tipos,
		Categorias:         categorias,
		EstadosFlujo:       estados,
		AccionesSugeridas:  acciones,
		EtiquetaTercero:    etiquetaTercero,
		EtiquetaReferencia: etiquetaReferencia,
		MetadataEjemplo:    metadata,
	}
}

var empresaModuloColombiaPlantillasPlantillas = map[string]EmpresaModuloColombiaPlantilla{
	"eventos_boleteria": buildVerticalPlantilla(
		"Eventos y boleteria",
		[]string{"evento", "boleta", "preventa", "aforo", "validacion_qr", "patrocinio"},
		[]string{"concierto", "conferencia", "deportivo", "feria", "privado", "corporativo"},
		[]string{"planeado", "en_venta", "agotado", "en_ingreso", "realizado", "cancelado"},
		[]string{"publicacion", "venta", "validacion", "control_aforo", "cierre"},
		"Cliente / promotor",
		"Evento / boleta QR",
		`{"aforo":500,"sector":"VIP","codigo_qr":"EVT-001","canal":"web"}`,
	),
	"salon_spa": buildVerticalPlantilla(
		"Salon de belleza, barberia y spa",
		[]string{"cita", "servicio", "paquete", "cabina", "insumo", "comision"},
		[]string{"cabello", "barberia", "unas", "spa", "estetica", "venta_producto"},
		[]string{"agendado", "confirmado", "en_servicio", "finalizado", "pagado", "cancelado"},
		[]string{"agendar", "confirmar", "consumo_insumo", "comision", "cierre"},
		"Cliente / profesional",
		"Cita / cabina",
		`{"profesional":"Estilista demo","duracion_min":60,"cabina":"Cabina 1","comision":30}`,
	),
	"veterinaria_petshop": buildVerticalPlantilla(
		"Veterinaria y pet shop",
		[]string{"mascota", "consulta", "vacuna", "peluqueria", "producto", "hospitalizacion"},
		[]string{"canino", "felino", "exotico", "alimento", "medicamento", "accesorio"},
		[]string{"agendado", "en_atencion", "observacion", "entregado", "facturado", "cancelado"},
		[]string{"historia", "vacunacion", "formula", "seguimiento", "cierre"},
		"Propietario / mascota",
		"Historia / vacuna",
		`{"mascota":"Luna","especie":"Canino","peso_kg":12.5,"proxima_vacuna":"2026-06-10"}`,
	),
	"lavanderia_tintoreria": buildVerticalPlantilla(
		"Lavanderia y tintoreria",
		[]string{"orden", "prenda", "etiqueta", "lavado", "planchado", "entrega"},
		[]string{"lavado", "tintoreria", "planchado", "domicilio", "industrial", "delicado"},
		[]string{"recibido", "clasificado", "en_proceso", "listo", "entregado", "reclamacion"},
		[]string{"recepcion", "etiquetado", "control_calidad", "entrega", "reclamacion"},
		"Cliente / prenda",
		"Orden / etiqueta",
		`{"prendas":5,"peso_kg":3.2,"servicio":"Lavado y planchado","domicilio":false}`,
	),
	"taller_mecanico": buildVerticalPlantilla(
		"Taller mecanico, motos y autos",
		[]string{"orden_trabajo", "diagnostico", "repuesto", "mano_obra", "garantia", "entrega"},
		[]string{"moto", "automovil", "preventivo", "correctivo", "latoneria", "electrico"},
		[]string{"recibido", "diagnosticado", "aprobado", "en_taller", "listo", "entregado"},
		[]string{"diagnostico", "cotizacion", "aprobacion", "reparacion", "garantia", "cierre"},
		"Cliente / vehiculo",
		"Placa / orden",
		`{"placa":"ABC123","km":45200,"tecnico":"Mecanico demo","garantia_dias":30}`,
	),
	"transporte_carga_tms": buildVerticalPlantilla(
		"Transporte de carga / TMS",
		[]string{"flete", "manifiesto", "conductor", "vehiculo", "entrega", "novedad"},
		[]string{"urbano", "nacional", "refrigerado", "paqueteo", "contenedor", "ultima_milla"},
		[]string{"programado", "cargado", "en_ruta", "entregado", "liquidado", "cancelado"},
		[]string{"programacion", "cargue", "tracking", "entrega", "liquidacion"},
		"Cliente / conductor",
		"Manifiesto / placa",
		`{"origen":"Barranquilla","destino":"Cartagena","placa":"TRK001","peso_kg":1200}`,
	),
	"servicios_tecnicos": buildVerticalPlantilla(
		"Servicios tecnicos a domicilio",
		[]string{"orden_servicio", "visita", "diagnostico", "repuesto", "firma", "garantia"},
		[]string{"electrodomestico", "computo", "aire", "plomeria", "electricidad", "instalacion"},
		[]string{"solicitado", "programado", "en_visita", "cotizado", "finalizado", "garantia"},
		[]string{"programacion", "visita", "evidencia", "firma_cliente", "cierre"},
		"Cliente / tecnico",
		"Orden / equipo",
		`{"equipo":"Aire acondicionado","serial":"SER-001","tecnico":"Tecnico demo","firma":false}`,
	),
	"funeraria_exequial": buildVerticalPlantilla(
		"Funeraria y servicios exequiales",
		[]string{"plan", "afiliado", "servicio", "sala", "contrato", "documento"},
		[]string{"prevision", "servicio_inmediato", "sala_velacion", "traslado", "documentos", "floristeria"},
		[]string{"afiliado", "solicitado", "en_servicio", "documentando", "cerrado", "anulado"},
		[]string{"afiliacion", "servicio", "documentacion", "autorizacion", "cierre"},
		"Afiliado / responsable",
		"Contrato / servicio",
		`{"plan":"Familiar","beneficiarios":4,"sala":"Sala 1","documentos":"pendiente"}`,
	),
	"parque_recreativo": buildVerticalPlantilla(
		"Parque recreativo y atracciones",
		[]string{"entrada", "manilla_qr", "atraccion", "aforo", "consumo", "incidente"},
		[]string{"general", "vip", "infantil", "atraccion", "alimentos", "evento"},
		[]string{"emitido", "ingresado", "en_parque", "consumido", "cerrado", "bloqueado"},
		[]string{"venta_entrada", "ingreso_qr", "control_aforo", "consumo", "cierre"},
		"Visitante / operador",
		"Manilla / entrada",
		`{"manilla":"PK-001","aforo_zona":120,"atracciones":5,"saldo_consumo":50000}`,
	),
}

var empresaModuloColombiaSeccionesPlantillas = map[string][]string{
	"eventos_boleteria":     {"Dashboard", "Eventos", "Boletas QR", "Aforo", "Validacion", "Reportes"},
	"salon_spa":             {"Agenda", "Servicios", "Profesionales", "Insumos", "Comisiones", "Cierre"},
	"veterinaria_petshop":   {"Pacientes", "Historia", "Vacunas", "Peluqueria", "Productos", "Seguimiento"},
	"lavanderia_tintoreria": {"Recepcion", "Prendas", "Etiquetas", "Proceso", "Entrega", "Reclamos"},
	"taller_mecanico":       {"Ordenes", "Diagnostico", "Repuestos", "Mano de obra", "Garantias", "Entrega"},
	"transporte_carga_tms":  {"Fletes", "Manifiestos", "Conductores", "Tracking", "Entregas", "Liquidacion"},
	"servicios_tecnicos":    {"Ordenes", "Agenda", "Tecnicos", "Repuestos", "Firmas", "Garantias"},
	"funeraria_exequial":    {"Planes", "Afiliados", "Servicios", "Salas", "Documentos", "Cierre"},
	"parque_recreativo":     {"Entradas", "Manillas QR", "Aforo", "Atracciones", "Consumo", "Incidentes"},
}

func GetEmpresaModuloColombiaSeccionesFlujo(modulo string) []string {
	modulo = NormalizeEmpresaModuloColombia(modulo)
	if rows, ok := empresaModuloColombiaSeccionesPlantillas[modulo]; ok {
		return append([]string{}, rows...)
	}
	return []string{"Dashboard", "Configuracion", "Registros", "Seguimiento", "Aprobaciones", "Evidencias"}
}
