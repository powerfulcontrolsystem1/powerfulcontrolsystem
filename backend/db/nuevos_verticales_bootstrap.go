package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type NuevoVerticalTipoEmpresa struct {
	Modulo        string
	Nombre        string
	Observaciones string
	StationPrefix string
	StationCount  int
	Roles         []string
}

type NuevoVerticalLicenciaPlan struct {
	Nombre                 string
	Descripcion            string
	Valor                  float64
	DuracionDias           int
	MaxDocumentosMensuales int
	ModulosHabilitados     string
}

type nuevoVerticalBootstrapMeta struct {
	Modulo        string
	StationPrefix string
	StationCount  int
	Roles         []string
}

var nuevosVerticalesBootstrapMeta = []nuevoVerticalBootstrapMeta{
	{"agencia_viajes", "Asesor", 3, []string{"asesor_viajes", "operacion", "caja"}},
	{"operador_turistico", "Ruta", 4, []string{"coordinador_tours", "guia", "caja"}},
	{"eventos_boleteria", "Taquilla", 4, []string{"productor_eventos", "taquilla", "control_acceso", "caja"}},
	{"salon_spa", "Puesto", 4, []string{"recepcion", "profesional_belleza", "caja"}},
	{"veterinaria_petshop", "Consultorio", 3, []string{"veterinario", "auxiliar_veterinaria", "caja"}},
	{"clinica_consultorios", "Consultorio", 5, []string{"recepcion", "profesional_salud", "caja"}},
	{"laboratorio_clinico", "Toma", 4, []string{"recepcion", "bacteriologo", "auxiliar_laboratorio", "caja"}},
	{"colegio_academia", "Aula", 6, []string{"coordinador_academico", "docente", "caja"}},
	{"guarderia_infantil", "Salon", 4, []string{"coordinador_guarderia", "docente", "auxiliar", "caja"}},
	{"lavanderia_tintoreria", "Punto", 3, []string{"recepcion", "operario_lavanderia", "caja"}},
	{"taller_mecanico", "Bahia", 5, []string{"recepcion", "tecnico", "compras", "caja"}},
	{"transporte_carga_tms", "Ruta", 5, []string{"coordinador_logistico", "conductor", "caja"}},
	{"servicios_tecnicos", "Tecnico", 4, []string{"coordinador_servicio", "tecnico", "caja"}},
	{"inmobiliaria_comercial", "Asesor", 4, []string{"asesor_inmobiliario", "coordinador", "caja"}},
	{"seguridad_privada", "Puesto", 6, []string{"coordinador_seguridad", "guarda", "supervisor", "caja"}},
	{"club_deportivo", "Cancha", 5, []string{"coordinador_deportivo", "entrenador", "caja"}},
	{"funeraria_exequial", "Sala", 4, []string{"coordinador_servicio", "asesor_exequial", "caja"}},
	{"parque_recreativo", "Atraccion", 6, []string{"operador_parque", "taquilla", "supervisor", "caja"}},
	{"cooperativa_fondo", "Oficina", 3, []string{"analista_credito", "cartera", "caja"}},
	{"capacitacion_empresarial", "Aula", 5, []string{"coordinador_capacitacion", "instructor", "caja"}},
}

var nuevosVerticalesTipoEmpresaCatalog = buildNuevosVerticalesTipoEmpresaCatalog()

var nuevosVerticalesProduccionMasiva = map[string]int{
	"salon_spa":                1,
	"veterinaria_petshop":      2,
	"clinica_consultorios":     3,
	"laboratorio_clinico":      4,
	"taller_mecanico":          5,
	"servicios_tecnicos":       6,
	"lavanderia_tintoreria":    7,
	"agencia_viajes":           8,
	"eventos_boleteria":        9,
	"transporte_carga_tms":     10,
	"operador_turistico":       11,
	"colegio_academia":         12,
	"guarderia_infantil":       13,
	"inmobiliaria_comercial":   14,
	"seguridad_privada":        15,
	"club_deportivo":           16,
	"funeraria_exequial":       17,
	"parque_recreativo":        18,
	"cooperativa_fondo":        19,
	"capacitacion_empresarial": 20,
}

func buildNuevosVerticalesTipoEmpresaCatalog() []NuevoVerticalTipoEmpresa {
	out := make([]NuevoVerticalTipoEmpresa, 0, len(nuevosVerticalesBootstrapMeta))
	for _, meta := range nuevosVerticalesBootstrapMeta {
		modulo := NormalizeEmpresaModuloColombia(meta.Modulo)
		plantilla := GetEmpresaModuloColombiaPlantilla(modulo)
		if modulo == "" || strings.TrimSpace(plantilla.Titulo) == "" {
			continue
		}
		out = append(out, NuevoVerticalTipoEmpresa{
			Modulo:        modulo,
			Nombre:        plantilla.Titulo,
			Observaciones: nuevoVerticalObservaciones(plantilla),
			StationPrefix: meta.StationPrefix,
			StationCount:  meta.StationCount,
			Roles:         append([]string{}, meta.Roles...),
		})
	}
	return out
}

func nuevoVerticalObservaciones(plantilla EmpresaModuloColombiaPlantilla) string {
	tipos := readableList(plantilla.Tipos, 6)
	categorias := readableList(plantilla.Categorias, 4)
	if tipos == "" {
		tipos = "registros operativos, agenda, evidencias y reportes"
	}
	if categorias == "" {
		return "Gestion profesional de " + tipos + "."
	}
	return "Gestion profesional de " + tipos + " con categorias de " + categorias + "."
}

func NuevosVerticalesTipoEmpresaCatalog() []NuevoVerticalTipoEmpresa {
	out := make([]NuevoVerticalTipoEmpresa, len(nuevosVerticalesTipoEmpresaCatalog))
	for i, item := range nuevosVerticalesTipoEmpresaCatalog {
		out[i] = item
		out[i].Roles = append([]string{}, item.Roles...)
	}
	return out
}

func NuevosVerticalesProduccionMasivaSeleccionados() []string {
	maxRank := 0
	for _, rank := range nuevosVerticalesProduccionMasiva {
		if rank > maxRank {
			maxRank = rank
		}
	}
	out := make([]string, maxRank)
	for modulo, rank := range nuevosVerticalesProduccionMasiva {
		if rank >= 1 && rank <= len(out) {
			out[rank-1] = modulo
		}
	}
	return uniqueTrimmedStrings(out, false)
}

func NuevoVerticalProduccionMasivaRank(modulo string) int {
	modulo = NormalizeEmpresaModuloColombia(modulo)
	return nuevosVerticalesProduccionMasiva[modulo]
}

func BuildTipoEmpresaPreconfigIntegracionVertical(modulo string) *TipoEmpresaPreconfigIntegracionVertical {
	clean := NormalizeEmpresaModuloColombia(modulo)
	if clean == "" {
		clean = strings.ToLower(strings.TrimSpace(modulo))
	}
	if clean == "" {
		return nil
	}
	if !isNuevoVerticalTipoEmpresaModulo(clean) {
		return buildClassicTipoEmpresaPreconfigIntegracionVertical(clean)
	}
	plantilla := GetEmpresaModuloColombiaPlantilla(clean)
	if strings.TrimSpace(plantilla.Titulo) == "" {
		return buildClassicTipoEmpresaPreconfigIntegracionVertical(clean)
	}
	rank := nuevosVerticalesProduccionMasiva[clean]
	produccion := rank > 0
	decision := "integrar_v1_produccion_masiva"
	motivo := "Plantilla real de produccion masiva: opera sobre clientes, servicios, ventas, pagos, facturacion, reportes y seguridad centrales, conservando solo la especialidad del vertical en empresa_modulos_colombia."
	if produccion {
		decision = "integrar_v1_produccion_masiva"
	}
	if motivo == "" {
		motivo = "Mantener como plantilla real disponible sobre el nucleo unico."
	}
	return &TipoEmpresaPreconfigIntegracionVertical{
		Modulo:               clean,
		EstadoIntegracion:    "plantilla_integrada_nucleo",
		Decision:             decision,
		ProduccionMasiva:     produccion,
		PrioridadProduccion:  rank,
		MotivoDecision:       motivo,
		TemplateActivates:    []string{clean, "clientes", "inventario/servicios", "ventas", "pagos", "facturacion", "reportes", "seguridad", "empresa_modulos_colombia"},
		TablesTouched:        []string{"empresa_modulos_colombia_registros", "empresa_modulos_colombia_eventos", "empresa_modulos_colombia_evidencias", "empresa_modulos_colombia_aprobaciones", "empresa_modulos_colombia_tareas", "clientes", "servicios", "carritos_compras", "carrito_compra_items", "empresa_finanzas_movimientos"},
		RequiredPermissions:  []string{"seguridad:R", clean + ":R", clean + ":C", "clientes:R/C", "inventario:R/C servicios", "ventas:C", "pagos:C", "finanzas:R/C", "reportes:R"},
		SaleFlow:             []string{"registro del vertical", "cliente y servicio central", "cotizacion o venta central", "pago/facturacion central", "ingreso conciliable en finanzas", "seguimiento y cierre por empresa_modulos_colombia"},
		ReportsProduced:      []string{"dashboard del vertical", "agenda y SLA", "responsables", "riesgo operativo", "ventas centrales", "ingresos y egresos del nucleo", "auditoria por empresa"},
		FinancialCoreModules: []string{"ventas", "pagos", "finanzas", "bancos_pagos", "tesoreria_presupuesto", "reportes"},
		IncomeFlow:           []string{"servicio/producto vendible del vertical", "carrito o venta central", "pago central", "movimiento ingreso en empresa_finanzas_movimientos", "reporte financiero consolidado"},
		ExpenseFlow:          []string{"compra/gasto operativo del vertical", "soporte o documento central", "movimiento egreso en empresa_finanzas_movimientos", "conciliacion bancaria/tesoreria", "reporte financiero consolidado"},
		FinancialTables:      []string{"carritos_compras", "carrito_compra_items", "empresa_finanzas_movimientos", "empresa_finanzas_configuracion", "empresa_finanzas_periodos"},
		FinancialReports:     []string{"ingresos por vertical", "egresos por vertical", "margen operativo", "flujo de caja", "estado de resultados por empresa"},
	}
}

func isNuevoVerticalTipoEmpresaModulo(modulo string) bool {
	modulo = NormalizeEmpresaModuloColombia(modulo)
	for _, item := range nuevosVerticalesTipoEmpresaCatalog {
		if item.Modulo == modulo {
			return true
		}
	}
	return false
}

func buildClassicTipoEmpresaPreconfigIntegracionVertical(modulo string) *TipoEmpresaPreconfigIntegracionVertical {
	switch strings.ToLower(strings.TrimSpace(modulo)) {
	case "gimnasio", "odontologia", "turnos_atencion", "vehiculos", "taller", "lavadero_autos", "hotel", "motel", "drogueria_farmacia", "alquileres", "constructora":
		return &TipoEmpresaPreconfigIntegracionVertical{
			Modulo:               strings.ToLower(strings.TrimSpace(modulo)),
			EstadoIntegracion:    "plantilla_integrada_nucleo",
			Decision:             "mantener_como_plantilla",
			ProduccionMasiva:     false,
			MotivoDecision:       "Vertical clasico conectado a preconfiguracion; la produccion masiva de nuevos verticales se gobierna en el catalogo de 20 plantillas nuevas.",
			TemplateActivates:    []string{modulo, "clientes", "inventario/servicios", "ventas", "pagos", "reportes", "seguridad"},
			TablesTouched:        []string{"clientes", "servicios", "carritos_compras", "carrito_compra_items", "empresa_finanzas_movimientos"},
			RequiredPermissions:  []string{"seguridad:R", modulo + ":R", modulo + ":C", "ventas:C", "pagos:C", "finanzas:R/C", "reportes:R"},
			SaleFlow:             []string{"registro especializado", "cliente/servicio central", "carrito central", "pago o factura central", "ingreso conciliable en finanzas", "reporte consolidado"},
			ReportsProduced:      []string{"reporte operativo", "ventas centrales", "ingresos y egresos del nucleo", "auditoria por empresa"},
			FinancialCoreModules: []string{"ventas", "pagos", "finanzas", "bancos_pagos", "tesoreria_presupuesto", "reportes"},
			IncomeFlow:           []string{"servicio/producto vendible del vertical", "carrito o venta central", "pago central", "movimiento ingreso en empresa_finanzas_movimientos", "reporte financiero consolidado"},
			ExpenseFlow:          []string{"compra/gasto operativo del vertical", "soporte o documento central", "movimiento egreso en empresa_finanzas_movimientos", "conciliacion bancaria/tesoreria", "reporte financiero consolidado"},
			FinancialTables:      []string{"carritos_compras", "carrito_compra_items", "empresa_finanzas_movimientos", "empresa_finanzas_configuracion", "empresa_finanzas_periodos"},
			FinancialReports:     []string{"ingresos por vertical", "egresos por vertical", "margen operativo", "flujo de caja", "estado de resultados por empresa"},
		}
	default:
		return nil
	}
}

func DefaultNuevoVerticalLicenciaModules(modulo string) string {
	modulo = NormalizeEmpresaModuloColombia(modulo)
	modules := []string{
		"ventas",
		"inventario",
		"compras",
		"clientes",
		"crm_unificado",
		"finanzas",
		"bancos_pagos",
		"tesoreria_presupuesto",
		"cobranza",
		"gestion_documental",
		"contratos_obligaciones",
		"calidad_procesos",
		"facturacion",
		"seguridad",
	}
	if modulo != "" {
		modules = append(modules, modulo)
	}
	return strings.Join(uniqueNonEmptyStrings(modules), ",")
}

func DefaultNuevoVerticalLicenciaPlans(item NuevoVerticalTipoEmpresa) []NuevoVerticalLicenciaPlan {
	modules := DefaultNuevoVerticalLicenciaModules(item.Modulo)
	nombre := strings.TrimSpace(item.Nombre)
	descripcion := "Licencia para " + strings.ToLower(nombre) + ": " + strings.TrimSpace(item.Observaciones)
	return []NuevoVerticalLicenciaPlan{
		{Nombre: nombre + " prueba 15 dias", Descripcion: descripcion, Valor: 0, DuracionDias: 15, MaxDocumentosMensuales: 250, ModulosHabilitados: modules},
		{Nombre: nombre + " 30 dias - 1000 documentos", Descripcion: descripcion, Valor: 60000, DuracionDias: 30, MaxDocumentosMensuales: 1000, ModulosHabilitados: modules},
		{Nombre: nombre + " 30 dias - 2000 documentos", Descripcion: descripcion, Valor: 100000, DuracionDias: 30, MaxDocumentosMensuales: 2000, ModulosHabilitados: modules},
		{Nombre: nombre + " 30 dias - 4000 documentos", Descripcion: descripcion, Valor: 150000, DuracionDias: 30, MaxDocumentosMensuales: 4000, ModulosHabilitados: modules},
	}
}

func EnsureNuevosVerticalesTipoEmpresaYLicencias(dbConn *sql.DB, usuario string) (tiposAsegurados, licenciasAseguradas int, err error) {
	if dbConn == nil {
		return 0, 0, errors.New("db connection is nil")
	}
	if err := EnsureLicenciasSchema(dbConn); err != nil {
		return 0, 0, err
	}
	if err := EnsureCanonicalTiposEmpresaPreconfigurables(dbConn); err != nil {
		return 0, 0, err
	}
	if err := EnsureTipoEmpresaPreconfiguracionSchema(dbConn); err != nil {
		return 0, 0, err
	}
	for _, item := range nuevosVerticalesTipoEmpresaCatalog {
		tipoID, err := ensureNuevoVerticalTipoEmpresa(dbConn, item)
		if err != nil {
			return tiposAsegurados, licenciasAseguradas, err
		}
		tiposAsegurados++
		for _, plan := range DefaultNuevoVerticalLicenciaPlans(item) {
			if err := ensureNuevoVerticalLicenciaPlan(dbConn, tipoID, usuario, plan); err != nil {
				return tiposAsegurados, licenciasAseguradas, err
			}
			licenciasAseguradas++
		}
		if _, err := UpsertTipoEmpresaPreconfiguracion(dbConn, defaultNuevoVerticalTipoEmpresaPreconfiguracion(tipoID, item, usuario)); err != nil {
			return tiposAsegurados, licenciasAseguradas, err
		}
	}
	return tiposAsegurados, licenciasAseguradas, nil
}

func EnsureNuevosVerticalesProduccionMasivaLicencias(dbConn *sql.DB, usuario string) (tiposAsegurados, licenciasAseguradas int, err error) {
	if dbConn == nil {
		return 0, 0, errors.New("db connection is nil")
	}
	if err := EnsureLicenciasSchema(dbConn); err != nil {
		return 0, 0, err
	}
	if err := EnsureCanonicalTiposEmpresaPreconfigurables(dbConn); err != nil {
		return 0, 0, err
	}
	if err := EnsureTipoEmpresaPreconfiguracionSchema(dbConn); err != nil {
		return 0, 0, err
	}
	for _, modulo := range NuevosVerticalesProduccionMasivaSeleccionados() {
		item, ok := getNuevoVerticalTipoEmpresaByModulo(modulo)
		if !ok {
			return tiposAsegurados, licenciasAseguradas, fmt.Errorf("vertical de produccion no encontrado: %s", modulo)
		}
		tipoID, err := ensureNuevoVerticalTipoEmpresa(dbConn, item)
		if err != nil {
			return tiposAsegurados, licenciasAseguradas, err
		}
		tiposAsegurados++
		for _, plan := range DefaultNuevoVerticalLicenciaPlans(item) {
			if err := ensureNuevoVerticalLicenciaPlan(dbConn, tipoID, usuario, plan); err != nil {
				return tiposAsegurados, licenciasAseguradas, err
			}
			licenciasAseguradas++
		}
		if _, err := UpsertTipoEmpresaPreconfiguracion(dbConn, defaultNuevoVerticalTipoEmpresaPreconfiguracion(tipoID, item, usuario)); err != nil {
			return tiposAsegurados, licenciasAseguradas, err
		}
	}
	return tiposAsegurados, licenciasAseguradas, nil
}

func getNuevoVerticalTipoEmpresaByModulo(modulo string) (NuevoVerticalTipoEmpresa, bool) {
	modulo = NormalizeEmpresaModuloColombia(modulo)
	for _, item := range nuevosVerticalesTipoEmpresaCatalog {
		if item.Modulo == modulo {
			return item, true
		}
	}
	return NuevoVerticalTipoEmpresa{}, false
}

func ensureNuevoVerticalTipoEmpresa(dbConn *sql.DB, item NuevoVerticalTipoEmpresa) (int64, error) {
	tipos, err := GetTiposEmpresas(dbConn)
	if err != nil {
		return 0, err
	}
	for _, tipo := range tipos {
		if nuevoVerticalTipoEmpresaMatches(tipo.Nombre, item) {
			if strings.EqualFold(strings.TrimSpace(tipo.Estado), "inactivo") {
				if err := SetTipoEmpresaActivo(dbConn, tipo.ID, "activo"); err != nil {
					return 0, err
				}
			}
			if strings.TrimSpace(tipo.Observaciones) == "" {
				_ = UpdateTipoEmpresa(dbConn, tipo.ID, item.Nombre, item.Observaciones)
			}
			return tipo.ID, nil
		}
	}
	return CreateTipoEmpresa(dbConn, item.Nombre, item.Observaciones)
}

func ensureNuevoVerticalLicenciaPlan(dbConn *sql.DB, tipoID int64, usuario string, plan NuevoVerticalLicenciaPlan) error {
	if tipoID <= 0 {
		return fmt.Errorf("tipo_id nuevo vertical invalido")
	}
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = "sistema.verticales"
	}
	nowExpr := sqlNowExpr()
	var existingID int64
	err := queryRowSQLCompat(dbConn, `SELECT id
		FROM licencias
		WHERE tipo_id = ?
			AND COALESCE(empresa_id, 0) = 0
			AND LOWER(TRIM(COALESCE(nombre, ''))) = LOWER(TRIM(?))
		LIMIT 1`, tipoID, plan.Nombre).Scan(&existingID)
	if err == nil && existingID > 0 {
		_, err = execSQLCompat(dbConn, `UPDATE licencias
			SET pais_codigo = 'CO',
				descripcion = ?,
				valor = ?,
				duracion_dias = ?,
				max_documentos_mensuales = ?,
				modulos_habilitados = ?,
				es_adicional = 0,
				codigo_funcion = '',
				super_rol_habilitado = 0,
				activo = 1,
				estado = 'activo',
				usuario_creador = COALESCE(NULLIF(TRIM(usuario_creador), ''), ?),
				fecha_actualizacion = `+nowExpr+`
			WHERE id = ?`, plan.Descripcion, plan.Valor, plan.DuracionDias, plan.MaxDocumentosMensuales, plan.ModulosHabilitados, usuario, existingID)
		return err
	}
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	id, err := CreateLicenciaAdvancedWithLimits(dbConn, tipoID, "CO", plan.Nombre, plan.Descripcion, plan.Valor, plan.DuracionDias, plan.ModulosHabilitados, 0, "", 0, plan.MaxDocumentosMensuales)
	if err != nil {
		return err
	}
	_, _ = execSQLCompat(dbConn, "UPDATE licencias SET usuario_creador = COALESCE(NULLIF(TRIM(usuario_creador), ''), ?), fecha_actualizacion = "+nowExpr+" WHERE id = ?", usuario, id)
	return nil
}

func defaultNuevoVerticalTipoEmpresaPreconfiguracion(tipoID int64, item NuevoVerticalTipoEmpresa, usuario string) TipoEmpresaPreconfiguracion {
	template, _ := defaultNuevoVerticalTipoEmpresaPreconfigTemplate(item.Nombre)
	raw, _ := MarshalTipoEmpresaPreconfigTemplate(template)
	return TipoEmpresaPreconfiguracion{
		TipoEmpresaID:  tipoID,
		Enabled:        true,
		Nombre:         "Preconfiguracion " + strings.TrimSpace(item.Nombre),
		Descripcion:    "Plantilla inicial profesional para " + strings.ToLower(strings.TrimSpace(item.Nombre)) + ".",
		ConfigJSON:     raw,
		UsuarioCreador: strings.TrimSpace(usuario),
		Estado:         "activo",
	}
}

func defaultNuevoVerticalTipoEmpresaPreconfigTemplate(tipoNombre string) (TipoEmpresaPreconfigTemplate, bool) {
	item, ok := matchNuevoVerticalTipoEmpresa(tipoNombre)
	if !ok {
		return TipoEmpresaPreconfigTemplate{}, false
	}
	plantilla := GetEmpresaModuloColombiaPlantilla(item.Modulo)
	tipoPrincipal := firstString(plantilla.Tipos, "servicio")
	tipoSecundario := nthString(plantilla.Tipos, 1, "seguimiento")
	categoriaPrincipal := firstString(plantilla.Categorias, "operacion")
	categoriaSecundaria := nthString(plantilla.Categorias, 1, "administracion")
	prefix := strings.ToUpper(strings.ReplaceAll(item.Modulo, "_", "-"))
	template := withPreconfigOperacion(newDefaultTipoEmpresaPreconfigTemplate(prefix, item.StationPrefix, item.StationCount, []TipoEmpresaPreconfigProducto{
		productoPreconfig("DEMO-"+prefix+"-001", plantilla.Titulo+" - "+strings.ReplaceAll(tipoPrincipal, "_", " "), categoriaPrincipal, "Producto o servicio guia para iniciar operacion del vertical.", 0, 85000, 0),
		productoPreconfig("DEMO-"+prefix+"-002", plantilla.Titulo+" - "+strings.ReplaceAll(tipoSecundario, "_", " "), categoriaSecundaria, "Concepto guia para seguimiento, aprobacion o control documental.", 0, 120000, 0),
		productoPreconfig("DEMO-"+prefix+"-003", "Paquete operativo "+plantilla.Titulo, "Paquetes", "Paquete demostrativo para venta, agenda, seguimiento y reporte.", 0, 250000, 0),
	}, defaultNuevoVerticalUsuarios(item), "Asistente para "+strings.ToLower(plantilla.Titulo)+": configuracion, registros, agenda, SLA, evidencias, aprobaciones, reportes y seguimiento comercial."), operacionPreconfig(item.Modulo, item.StationPrefix, pluralizeTipoEmpresaStationName(item.StationPrefix), item.StationCount > 0, true, true, firstString(item.Roles, "operacion"), "servicio", 15, item.Roles))
	template.TareasGuia = append(template.TareasGuia,
		TipoEmpresaPreconfigTareaGuia{Modulo: plantilla.Titulo, Titulo: "Configurar flujo principal", Descripcion: "Revisar tipos, categorias, estados, responsables, SLA, evidencias y aprobaciones del modulo."},
		TipoEmpresaPreconfigTareaGuia{Modulo: "Licencias", Titulo: "Validar cobertura modular", Descripcion: "Confirmar que la licencia activa incluya " + item.Modulo + " y los modulos base de ventas, clientes, finanzas, documentos y seguridad."},
		TipoEmpresaPreconfigTareaGuia{Modulo: "Reportes", Titulo: "Definir indicadores iniciales", Descripcion: "Crear metas de registros abiertos, vencimientos, tiempos de atencion, cartera y cierre operativo."},
	)
	return enrichTipoEmpresaPreconfigTemplate(template), true
}

func defaultNuevoVerticalUsuarios(item NuevoVerticalTipoEmpresa) []TipoEmpresaPreconfigUsuario {
	roles := item.Roles
	if len(roles) == 0 {
		roles = []string{"administrador", "operacion", "caja"}
	}
	out := make([]TipoEmpresaPreconfigUsuario, 0, len(roles))
	for _, rol := range roles {
		label := strings.ReplaceAll(rol, "_", " ")
		out = append(out, usuarioPreconfig(strings.Title(label), rol, "Rol guia para "+strings.ToLower(item.Nombre)+".")) //nolint:staticcheck
	}
	return out
}

func matchNuevoVerticalTipoEmpresa(tipoNombre string) (NuevoVerticalTipoEmpresa, bool) {
	for _, item := range nuevosVerticalesTipoEmpresaCatalog {
		if nuevoVerticalTipoEmpresaMatches(tipoNombre, item) {
			return item, true
		}
	}
	return NuevoVerticalTipoEmpresa{}, false
}

func matchNuevoVerticalTipoEmpresaTitulo(tipoNombre string) (NuevoVerticalTipoEmpresa, bool) {
	clean := normalizeTipoEmpresaName(tipoNombre)
	for _, item := range nuevosVerticalesTipoEmpresaCatalog {
		if clean == normalizeTipoEmpresaName(item.Nombre) {
			return item, true
		}
	}
	return NuevoVerticalTipoEmpresa{}, false
}

func nuevoVerticalTipoEmpresaMatches(tipoNombre string, item NuevoVerticalTipoEmpresa) bool {
	clean := normalizeTipoEmpresaName(tipoNombre)
	return clean == normalizeTipoEmpresaName(item.Nombre) || clean == normalizeTipoEmpresaName(strings.ReplaceAll(item.Modulo, "_", " "))
}

func uniqueNonEmptyStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
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

func firstString(values []string, fallback string) string {
	return nthString(values, 0, fallback)
}

func nthString(values []string, idx int, fallback string) string {
	if idx >= 0 && idx < len(values) && strings.TrimSpace(values[idx]) != "" {
		return strings.TrimSpace(values[idx])
	}
	return fallback
}

func readableList(values []string, maxItems int) string {
	if maxItems <= 0 {
		maxItems = len(values)
	}
	out := make([]string, 0, maxItems)
	for _, value := range values {
		clean := strings.ReplaceAll(strings.TrimSpace(value), "_", " ")
		if clean == "" {
			continue
		}
		out = append(out, clean)
		if len(out) >= maxItems {
			break
		}
	}
	return strings.Join(out, ", ")
}
