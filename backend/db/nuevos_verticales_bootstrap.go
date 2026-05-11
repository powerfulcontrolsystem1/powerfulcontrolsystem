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
		"helpdesk",
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
