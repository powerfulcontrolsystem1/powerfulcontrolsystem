package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	permActionRead    = "R"
	permActionCreate  = "C"
	permActionUpdate  = "U"
	permActionDelete  = "D"
	permActionApprove = "A"

	permModuleVentas               = "ventas"
	permModuleInventario           = "inventario"
	permModuleFinanzas             = "finanzas"
	permModuleContabilidadCO       = "contabilidad_colombia"
	permModuleContabilidadCOAv     = "contabilidad_colombia_avanzada"
	permModuleCentrosCosto         = "centros_costo"
	permModuleCierreFiscal         = "cierre_fiscal"
	permModuleActivosFijosNIIF     = "activos_fijos_niif_fiscal"
	permModuleDeclaracionesTrib    = "declaraciones_tributarias"
	permModuleClientes             = "clientes"
	permModuleCRMUnificado         = "crm_unificado"
	permModuleCompras              = "compras"
	permModuleFacturacion          = "facturacion"
	permModuleSeguridad            = "seguridad"
	permModuleVentaPublica         = "venta_publica"
	permModuleReservasHotel        = "reservas_hotel"
	permModuleChatTareas           = "chat_tareas"
	permModuleGimnasio             = "gimnasio"
	permModuleTaxiSystem           = "taxi_system"
	permModuleDomicilios           = "domicilios"
	permModuleParqueadero          = "parqueadero"
	permModuleApartTuristicos      = "apartamentos_turisticos"
	permModulePropiedadHorizontal  = "propiedad_horizontal"
	permModuleAlquileres           = "alquileres"
	permModuleOdontologia          = "odontologia"
	permModuleTurnos               = "turnos_atencion"
	permModuleControlElectrico     = "control_electrico"
	permModuleCarnets              = "carnets"
	permModuleHorariosTrab         = "horarios_trabajadores"
	permModuleAsistenciaEmpleados  = "asistencia_empleados"
	permModuleVehiculosRegistro    = "vehiculos_registro"
	permModuleHojaVidaOperativa    = "hoja_vida_operativa"
	permModuleUbicacionGPS         = "ubicacion_gps"
	permModuleProduccionMRP        = "produccion_mrp"
	permModuleLogisticaWMS         = "logistica_wms"
	permModuleTesoreria            = "tesoreria_presupuesto"
	permModuleNominaSueldos        = "nomina_sueldos"
	permModuleImportaciones        = "importaciones_costeo"
	permModuleAIUConstruccion      = "aiu_construccion"
	permModuleCobranza             = "cobranza"
	permModuleReportes             = "reportes"
	permModulePortalContador       = "portal_contador"
	permModulePortalTerceros       = "portal_terceros_certificados"
	permModuleSoportesComprasIA    = "soportes_compras_ia"
	permModuleBancosPagos          = "bancos_pagos"
	permModuleGestionDocumental    = "gestion_documental"
	permModuleCumplimientoKYC      = "cumplimiento_kyc"
	permModuleContratosOblig       = "contratos_obligaciones"
	permModuleCalidadProcesos      = "calidad_procesos"
	permModuleDrogueriaFarmacia    = "drogueria_farmacia"
	permModuleAuditoria            = "auditoria"
	permModuleBackups              = "backups"
	permModuleDocumentosOnlyOffice = "documentos_onlyoffice"

	permissionApprovalHeaderBy       = "X-Permission-Approved-By"
	permissionApprovalHeaderCode     = "X-Permission-Approval-Code"
	permissionApprovalHeaderReason   = "X-Permission-Approval-Reason"
	permissionApprovalHeaderRequired = "X-Permission-Approval-Required"
)

type permissionApprovalEvidence struct {
	ApprovedBy   string
	ApprovalCode string
	Reason       string
}

type empresaRateLimitBucket struct {
	WindowStart time.Time
	Count       int64
}

type empresaPermissionSnapshot struct {
	AdminRole              string
	EffectiveRole          string
	CanAccess              bool
	AllowedModules         map[string]bool
	AllowedVerticalModules map[string]bool
	RoleModuleActions      map[string]bool
	AllowedPages           map[string]bool
	ShareAccess            *empresaCompartidaScopeCtx
	LoadedAt               time.Time
}

type empresaPermissionSnapshotInflight struct {
	done     chan struct{}
	snapshot empresaPermissionSnapshot
	err      error
}

type permissionRoleOverrideCacheEntry struct {
	ModuleOverrides map[string]bool
	PageOverrides   map[string]bool
	LoadedAt        time.Time
}

type empresaPermissionOverrideCacheEntry struct {
	ModuleOverrides map[string]bool
	PageOverrides   map[string]bool
	Ctx             *empresaPermisosFinosCtx
	LoadedAt        time.Time
}

type permissionRoleModuleMatrixCacheEntry struct {
	Rows     []permissionModuleMatrixRow
	LoadedAt time.Time
}

var (
	empresaRateLimitMu              sync.Mutex
	empresaRateLimitBuckets         = map[string]empresaRateLimitBucket{}
	empresaPermissionCacheMu        sync.Mutex
	empresaPermissionCache          = map[string]empresaPermissionSnapshot{}
	empresaPermissionInflight       = map[string]*empresaPermissionSnapshotInflight{}
	rolePermissionModuleMatrixCache = map[string]permissionRoleModuleMatrixCacheEntry{}
	rolePermissionOverrideCache     = map[string]permissionRoleOverrideCacheEntry{}
	empresaPermissionOverrideCache  = map[int64]empresaPermissionOverrideCacheEntry{}
)

const empresaPermissionCacheTTL = 60 * time.Second
const permissionOverrideCacheTTL = 60 * time.Second

func invalidateEmpresaPermissionCacheForEmpresa(empresaID int64) {
	if empresaID <= 0 {
		return
	}
	suffix := fmt.Sprintf("|%d", empresaID)
	empresaPermissionCacheMu.Lock()
	for key := range empresaPermissionCache {
		if strings.HasSuffix(key, suffix) {
			delete(empresaPermissionCache, key)
		}
	}
	empresaPermissionCacheMu.Unlock()
}

var legacyPermissionVisibleTextReplacer = strings.NewReplacer(
	"Operaci\u00c3\u00b3n", "Operaci\u00f3n",
	"Configuraci\u00c3\u00b3n", "Configuraci\u00f3n",
	"Facturaci\u00c3\u00b3n", "Facturaci\u00f3n",
	"electr\u00c3\u00b3nica", "electr\u00f3nica",
	"cat\u00c3\u00a1logo", "cat\u00e1logo",
	"c\u00c3\u00b3digos", "c\u00f3digos",
	"c\u00c3\u00b3digo", "c\u00f3digo",
	"\u00c3\u00b3rdenes", "\u00f3rdenes",
	"N\u00c3\u00b3mina", "N\u00f3mina",
	"veh\u00c3\u00adculos", "veh\u00edculos",
	"veh\u00c3\u00adculo", "veh\u00edculo",
	"Auditor\u00c3\u00ada", "Auditor\u00eda",
	"Cr\u00c3\u00a9ditos", "Cr\u00e9ditos",
	"cr\u00c3\u00a9ditos", "cr\u00e9ditos",
	"Ubicaci\u00c3\u00b3n", "Ubicaci\u00f3n",
	"Aprobaci\u00c3\u00b3n", "Aprobaci\u00f3n",
	"d\u00c3\u00ada", "d\u00eda",
	"Gr\u00c3\u00a1ficos", "Gr\u00e1ficos",
	"estad\u00c3\u00adsticas", "estad\u00edsticas",
	"m\u00c3\u00b3dulo", "m\u00f3dulo",
	"acci\u00c3\u00b3n", "acci\u00f3n",
	"integraci\u00c3\u00b3n", "integraci\u00f3n",
)

func sanitizeLegacyPermissionVisibleText(value string) string {
	clean := strings.TrimSpace(value)
	if clean == "" {
		return ""
	}
	return strings.TrimSpace(legacyPermissionVisibleTextReplacer.Replace(clean))
}

var permissionUniversalGroupLabels = map[string]string{
	"Operaci\u00f3n diaria y ventas":                "Operaci\u00f3n universal y ventas",
	"Operaci\u00f3n y venta":                        "Operaci\u00f3n universal y ventas",
	"Operacion diaria y ventas":                     "Operaci\u00f3n universal y ventas",
	"Operacion y venta":                             "Operaci\u00f3n universal y ventas",
	"Verticales de negocio":                         "Soluciones universales por negocio",
	"Inventario y compras":                          "Inventario y compras universales",
	"Inventario y cat\u00e1logo":                    "Inventario y compras universales",
	"Inventario y catalogo":                         "Inventario y compras universales",
	"Compras":                                       "Inventario y compras universales",
	"Centro financiero y contable":                  "Centro financiero universal y contable",
	"Finanzas y reportes":                           "Centro financiero universal y contable",
	"Finanzas y n\u00f3mina":                        "Centro financiero universal y contable",
	"Administraci\u00f3n y configuraci\u00f3n":      "Administraci\u00f3n universal y configuraci\u00f3n",
	"Seguridad e integraci\u00f3n":                  "Administraci\u00f3n universal y configuraci\u00f3n",
	"Configuraci\u00f3n":                            "Administraci\u00f3n universal y configuraci\u00f3n",
	"Administracion y configuracion":                "Administraci\u00f3n universal y configuraci\u00f3n",
	"Seguridad e integracion":                       "Administraci\u00f3n universal y configuraci\u00f3n",
	"Configuracion":                                 "Administraci\u00f3n universal y configuraci\u00f3n",
	"Facturaci\u00f3n electr\u00f3nica":             "Facturaci\u00f3n electr\u00f3nica universal",
	"Facturaci\u00f3n DIAN":                         "Facturaci\u00f3n electr\u00f3nica universal",
	"Facturacion electronica":                       "Facturaci\u00f3n electr\u00f3nica universal",
	"Facturacion DIAN":                              "Facturaci\u00f3n electr\u00f3nica universal",
	"Gesti\u00f3n de Relaciones con Clientes (CRM)": "CRM universal y clientes",
	"Gestion de Relaciones con Clientes (CRM)":      "CRM universal y clientes",
	"Clientes":                   "CRM universal y clientes",
	"Personas y activos":         "Personas y activos universales",
	"An\u00e1lisis y control":    "An\u00e1lisis universal y control",
	"Analisis y control":         "An\u00e1lisis universal y control",
	"Documentos, nube y soporte": "Documentos universales, nube y soporte",
}

var permissionUniversalModuleLabels = map[string]string{
	"Ventas y servicio al cliente":             "Ventas universales y servicio al cliente",
	"Inventario y almac\u00e9n":                "Inventario universal y almac\u00e9n",
	"Finanzas, caja y reportes":                "Finanzas universales, caja y reportes",
	"Clientes y cartera comercial":             "CRM universal, clientes y cartera comercial",
	"Compras y proveedores":                    "Compras universales y proveedores",
	"Facturaci\u00f3n electr\u00f3nica (DIAN)": "Facturaci\u00f3n electr\u00f3nica universal (DIAN)",
	"Seguridad, usuarios e integraci\u00f3n":   "Administraci\u00f3n universal, usuarios e integraci\u00f3n",
}

func universalPermissionGroupLabel(value string) string {
	clean := sanitizeLegacyPermissionVisibleText(value)
	if clean == "" {
		return ""
	}
	if replacement, ok := permissionUniversalGroupLabels[clean]; ok {
		return replacement
	}
	return clean
}

func universalPermissionModuleLabel(value string) string {
	clean := sanitizeLegacyPermissionVisibleText(value)
	if clean == "" {
		return ""
	}
	if replacement, ok := permissionUniversalModuleLabels[clean]; ok {
		return replacement
	}
	return clean
}

var permissionModulesCatalogOrdered = []string{
	permModuleVentas,
	permModuleInventario,
	permModuleFinanzas,
	permModuleContabilidadCO,
	permModuleContabilidadCOAv,
	permModuleCentrosCosto,
	permModuleCierreFiscal,
	permModuleActivosFijosNIIF,
	permModuleDeclaracionesTrib,
	permModuleClientes,
	permModuleCRMUnificado,
	permModuleCompras,
	permModuleFacturacion,
	permModuleSeguridad,
	permModuleVentaPublica,
	permModuleReservasHotel,
	permModuleChatTareas,
	permModuleGimnasio,
	permModuleTaxiSystem,
	permModuleDomicilios,
	permModuleParqueadero,
	permModuleApartTuristicos,
	permModulePropiedadHorizontal,
	permModuleAlquileres,
	permModuleOdontologia,
	permModuleTurnos,
	permModuleControlElectrico,
	permModuleCarnets,
	permModuleHorariosTrab,
	permModuleAsistenciaEmpleados,
	permModuleVehiculosRegistro,
	permModuleHojaVidaOperativa,
	permModuleUbicacionGPS,
	permModuleProduccionMRP,
	permModuleLogisticaWMS,
	permModuleTesoreria,
	permModuleNominaSueldos,
	permModuleImportaciones,
	permModuleAIUConstruccion,
	permModuleCobranza,
	permModuleReportes,
	permModulePortalContador,
	permModulePortalTerceros,
	permModuleSoportesComprasIA,
	permModuleBancosPagos,
	permModuleGestionDocumental,
	permModuleCumplimientoKYC,
	permModuleContratosOblig,
	permModuleCalidadProcesos,
	permModuleDrogueriaFarmacia,
	permModuleAuditoria,
	permModuleBackups,
	permModuleDocumentosOnlyOffice,
}

var permissionActionsCatalogOrdered = []string{
	permActionRead,
	permActionCreate,
	permActionUpdate,
	permActionDelete,
	permActionApprove,
}

// Etiquetas cortas para UI (super: permisos por rol) y documentación.
var permissionActionDisplayNames = map[string]string{
	permActionRead:    "Leer / consultar",
	permActionCreate:  "Crear / registrar",
	permActionUpdate:  "Actualizar / modificar",
	permActionDelete:  "Eliminar / anular",
	permActionApprove: "Aprobar / auditar",
}

// permissionModuleDisplayNames nombres de negocio por clave de módulo.
var permissionModuleDisplayNames = map[string]string{
	permModuleVentas:               "Ventas y servicio al cliente",
	permModuleInventario:           "Inventario y almacén",
	permModuleFinanzas:             "Finanzas, caja y reportes",
	permModuleContabilidadCO:       "Contabilidad Colombia NIIF/DIAN",
	permModuleContabilidadCOAv:     "Suite contable Colombia avanzada",
	permModuleCentrosCosto:         "Centros de costo y rentabilidad",
	permModuleCierreFiscal:         "Cierre y bloqueo fiscal",
	permModuleActivosFijosNIIF:     "Activos fijos e intangibles NIIF/Fiscal",
	permModuleDeclaracionesTrib:    "Declaraciones tributarias Colombia",
	permModuleClientes:             "Clientes y cartera comercial",
	permModuleCRMUnificado:         "CRM unificado comercial",
	permModuleCompras:              "Compras y proveedores",
	permModuleFacturacion:          "Facturación electrónica (DIAN)",
	permModuleSeguridad:            "Seguridad, usuarios e integración",
	permModuleVentaPublica:         "Venta publica y carta de productos",
	permModuleReservasHotel:        "Reservas hoteleras",
	permModuleChatTareas:           "Chat, tareas y agenda compartida",
	permModuleGimnasio:             "Gimnasio y membresias",
	permModuleTaxiSystem:           "Taxi system y despacho GPS",
	permModuleDomicilios:           "Domicilios y delivery",
	permModuleParqueadero:          "Parqueadero y tickets QR",
	permModuleApartTuristicos:      "Apartamentos turisticos",
	permModulePropiedadHorizontal:  "Propiedad horizontal",
	permModuleAlquileres:           "Alquiler universal de activos",
	permModuleOdontologia:          "Odontologia y agenda clinica",
	permModuleTurnos:               "Turnos de atencion",
	permModuleControlElectrico:     "Control electrico e IoT",
	permModuleCarnets:              "Carnets empresariales",
	permModuleHorariosTrab:         "Horarios laborales",
	permModuleAsistenciaEmpleados:  "Asistencia de empleados",
	permModuleVehiculosRegistro:    "Registro de vehiculos",
	permModuleHojaVidaOperativa:    "Historial de activos",
	permModuleUbicacionGPS:         "Ubicacion GPS de activos",
	permModuleProduccionMRP:        "Produccion / MRP",
	permModuleLogisticaWMS:         "Logistica avanzada / WMS",
	permModuleTesoreria:            "Tesoreria y presupuesto",
	permModuleNominaSueldos:        "Nomina y sueldos",
	permModuleImportaciones:        "Importaciones y costeo",
	permModuleAIUConstruccion:      "AIU construccion y contratos de obra",
	permModuleCobranza:             "Gestion de cobranza",
	permModuleReportes:             "Reportes ejecutivos y analitica",
	permModulePortalContador:       "Portal contador",
	permModulePortalTerceros:       "Portal de terceros y certificados tributarios",
	permModuleSoportesComprasIA:    "Captura inteligente de compras y gastos",
	permModuleBancosPagos:          "Bancos y pagos masivos Colombia",
	permModuleGestionDocumental:    "Gestion documental y aprobaciones",
	permModuleCumplimientoKYC:      "Cumplimiento KYC/KYB y riesgo LAFT",
	permModuleContratosOblig:       "Contratos, obligaciones y firma electronica",
	permModuleCalidadProcesos:      "Calidad, procesos y no conformidades",
	permModuleDrogueriaFarmacia:    "Drogueria y farmacia",
	permModuleAuditoria:            "Auditoria empresarial",
	permModuleBackups:              "Backups empresariales",
	permModuleDocumentosOnlyOffice: "Documentos OnlyOffice",
}

var permissionRolesCatalogOrdered = []string{
	"super_administrador",
	"administrador_total",
	"admin_empresa",
	"supervisor_sucursal",
	"cajero",
	"inventario",
	"compras",
	"contabilidad",
	"auditor",
}

type permissionPageRule struct {
	PaginaClave   string   `json:"pagina_clave"`
	Modulo        string   `json:"modulo,omitempty"`
	Accion        string   `json:"accion,omitempty"`
	AnyModules    []string `json:"any_modules,omitempty"`
	AlwaysVisible bool     `json:"always_visible,omitempty"`
	Titulo        string   `json:"titulo,omitempty"`
	Grupo         string   `json:"grupo,omitempty"`
}

var permissionPagesCatalogOrdered = []permissionPageRule{
	{PaginaClave: "linkInicio", AlwaysVisible: true, Titulo: "Inicio (tablero)", Grupo: "Acceso general"},
	{PaginaClave: "linkPanelEmpresa", AlwaysVisible: true, Titulo: "Panel de empresa", Grupo: "Acceso general"},
	{PaginaClave: "linkVentas", Modulo: permModuleVentas, Accion: permActionRead, Titulo: "Punto de venta / TPV", Grupo: "Operacion diaria y ventas"},
	{PaginaClave: "linkVentaDirecta", Modulo: permModuleVentas, Accion: permActionCreate, Titulo: "Venta directa sin estacion", Grupo: "Operacion diaria y ventas"},
	{PaginaClave: "linkEstaciones", Modulo: permModuleVentas, Accion: permActionRead, Titulo: "Estaciones y terminales", Grupo: "Operacion diaria y ventas"},
	{PaginaClave: "linkCarritoCompras", Modulo: permModuleVentas, Accion: permActionCreate, Titulo: "Carritos", Grupo: "Configuracion - Ventas y cobro"},
	{PaginaClave: "linkCodigosDescuento", Modulo: permModuleVentas, Accion: permActionCreate, Titulo: "Codigos de descuento", Grupo: "Operacion diaria y ventas"},
	{PaginaClave: "linkRedSocialComercial", Modulo: permModuleVentas, Accion: permActionCreate, Titulo: "Red social empresarial", Grupo: "Operacion diaria y ventas"},
	{PaginaClave: "linkChatIA", Modulo: permModuleVentas, Accion: permActionRead, Titulo: "Asistente IA (chat empresarial)", Grupo: "Operacion diaria y ventas"},
	{PaginaClave: "linkReservasHotel", Modulo: permModuleReservasHotel, Accion: permActionCreate, Titulo: "Reservas (hotel / habitaciones)", Grupo: "Operacion diaria y ventas"},
	{PaginaClave: "linkChatTareas", Modulo: permModuleChatTareas, Accion: permActionCreate, Titulo: "Chat y tareas", Grupo: "Operacion diaria y ventas"},
	{PaginaClave: "linkTurnosAtencion", Modulo: permModuleTurnos, Accion: permActionCreate, Titulo: "Turnos de atencion y fila", Grupo: "Operacion diaria y ventas"},

	{PaginaClave: "linkVerticalesIntegracion", Modulo: permModuleSeguridad, Accion: permActionRead, Titulo: "Matriz de integracion de verticales", Grupo: "Administracion y configuracion"},
	{PaginaClave: "linkGimnasio", Modulo: permModuleGimnasio, Accion: permActionCreate, Titulo: "Gestion de gimnasio", Grupo: "Verticales de negocio"},
	{PaginaClave: "linkGimnasioDashboard", Modulo: permModuleGimnasio, Accion: permActionRead, Titulo: "Gimnasio - dashboard", Grupo: "Verticales de negocio"},
	{PaginaClave: "linkGimnasioSocios", Modulo: permModuleGimnasio, Accion: permActionCreate, Titulo: "Gimnasio - socios", Grupo: "Verticales de negocio"},
	{PaginaClave: "linkGimnasioPlanes", Modulo: permModuleGimnasio, Accion: permActionUpdate, Titulo: "Gimnasio - planes", Grupo: "Verticales de negocio"},
	{PaginaClave: "linkGimnasioEntrenadores", Modulo: permModuleGimnasio, Accion: permActionUpdate, Titulo: "Gimnasio - entrenadores", Grupo: "Verticales de negocio"},
	{PaginaClave: "linkGimnasioClases", Modulo: permModuleGimnasio, Accion: permActionCreate, Titulo: "Gimnasio - clases", Grupo: "Verticales de negocio"},
	{PaginaClave: "linkGimnasioInscripciones", Modulo: permModuleGimnasio, Accion: permActionCreate, Titulo: "Gimnasio - inscripciones", Grupo: "Verticales de negocio"},
	{PaginaClave: "linkGimnasioAsistencias", Modulo: permModuleGimnasio, Accion: permActionCreate, Titulo: "Gimnasio - asistencias", Grupo: "Verticales de negocio"},
	{PaginaClave: "linkGimnasioPagos", Modulo: permModuleGimnasio, Accion: permActionCreate, Titulo: "Gimnasio - pagos", Grupo: "Verticales de negocio"},
	{PaginaClave: "linkGimnasioAcceso", Modulo: permModuleGimnasio, Accion: permActionApprove, Titulo: "Gimnasio - control de acceso", Grupo: "Verticales de negocio"},
	{PaginaClave: "linkTaxiSystem", Modulo: permModuleTaxiSystem, Accion: permActionCreate, Titulo: "Taxi system y despacho GPS", Grupo: "Verticales de negocio"},
	{PaginaClave: "linkDomicilios", Modulo: permModuleDomicilios, Accion: permActionCreate, Titulo: "Domicilios y delivery", Grupo: "Verticales de negocio"},
	{PaginaClave: "linkParqueadero", Modulo: permModuleParqueadero, Accion: permActionCreate, Titulo: "Parqueadero y tickets QR", Grupo: "Verticales de negocio"},
	{PaginaClave: "linkApartamentosTuristicos", Modulo: permModuleApartTuristicos, Accion: permActionCreate, Titulo: "Apartamentos turisticos", Grupo: "Verticales de negocio"},
	{PaginaClave: "linkPropiedadHorizontal", Modulo: permModulePropiedadHorizontal, Accion: permActionCreate, Titulo: "Propiedad horizontal", Grupo: "Verticales de negocio"},
	{PaginaClave: "linkAlquileres", Modulo: permModuleAlquileres, Accion: permActionCreate, Titulo: "Alquiler universal de activos", Grupo: "Verticales de negocio"},
	{PaginaClave: "linkConsultorioOdontologico", Modulo: permModuleOdontologia, Accion: permActionCreate, Titulo: "Consultorio odontologico", Grupo: "Verticales de negocio"},
	{PaginaClave: "linkDrogueriaFarmacia", Modulo: permModuleDrogueriaFarmacia, Accion: permActionCreate, Titulo: "Drogueria / Farmacia", Grupo: "Verticales de negocio"},
	{PaginaClave: "linkAIUConstruccion", Modulo: permModuleAIUConstruccion, Accion: permActionCreate, Titulo: "AIU construccion y contratos de obra", Grupo: "Verticales de negocio"},

	{PaginaClave: "linkProductos", Modulo: permModuleInventario, Accion: permActionCreate, Titulo: "Productos y servicios", Grupo: "Inventario y compras"},
	{PaginaClave: "linkProductosMain", Modulo: permModuleInventario, Accion: permActionCreate, Titulo: "Inventario (Productos)", Grupo: "Inventario y compras"},
	{PaginaClave: "linkInventarioAvanzado", Modulo: permModuleInventario, Accion: permActionCreate, Titulo: "Inventario avanzado", Grupo: "Inventario y compras"},
	{PaginaClave: "linkCombosProductos", Modulo: permModuleInventario, Accion: permActionCreate, Titulo: "Combos y paquetes", Grupo: "Inventario y compras"},
	{PaginaClave: "linkPreciosHistorial", Modulo: permModuleInventario, Accion: permActionRead, Titulo: "Historial de productos", Grupo: "Inventario y compras"},
	{PaginaClave: "linkBodegas", Modulo: permModuleInventario, Accion: permActionUpdate, Titulo: "Bodegas", Grupo: "Inventario y compras"},
	{PaginaClave: "linkCategorias", Modulo: permModuleInventario, Accion: permActionUpdate, Titulo: "Categorias", Grupo: "Inventario y compras"},
	{PaginaClave: "linkGeneradorCodigosBarras", Modulo: permModuleInventario, Accion: permActionUpdate, Titulo: "Generador de codigos de barras", Grupo: "Inventario y compras"},
	{PaginaClave: "linkCompras", Modulo: permModuleCompras, Accion: permActionCreate, Titulo: "Compras y ordenes", Grupo: "Inventario y compras"},
	{PaginaClave: "linkComprasDoc", Modulo: permModuleCompras, Accion: permActionCreate, Titulo: "Gestion de compras", Grupo: "Inventario y compras"},
	{PaginaClave: "linkProveedores", Modulo: permModuleCompras, Accion: permActionCreate, Titulo: "Proveedores", Grupo: "Inventario y compras"},
	{PaginaClave: "linkComprasAvanzadas", Modulo: permModuleCompras, Accion: permActionCreate, Titulo: "Compras avanzadas", Grupo: "Inventario y compras"},
	{PaginaClave: "linkSoportesComprasIA", Modulo: permModuleSoportesComprasIA, Accion: permActionCreate, Titulo: "Captura inteligente de compras y gastos", Grupo: "Inventario y compras"},
	{PaginaClave: "linkSoportesComprasIAMenu", Modulo: permModuleSoportesComprasIA, Accion: permActionCreate, Titulo: "Captura inteligente de compras y gastos", Grupo: "Inventario y compras"},
	{PaginaClave: "linkImportacionesCosteo", Modulo: permModuleImportaciones, Accion: permActionCreate, Titulo: "Importaciones y costeo", Grupo: "Inventario y compras"},
	{PaginaClave: "linkProduccionMRP", Modulo: permModuleProduccionMRP, Accion: permActionCreate, Titulo: "Produccion / MRP", Grupo: "Inventario y compras"},
	{PaginaClave: "linkLogisticaWMS", Modulo: permModuleLogisticaWMS, Accion: permActionCreate, Titulo: "Logistica avanzada / WMS", Grupo: "Inventario y compras"},
	{PaginaClave: "linkCartaProductosPublica", Modulo: permModuleVentaPublica, Accion: permActionCreate, Titulo: "Carta publica de productos", Grupo: "Inventario y compras"},
	{PaginaClave: "linkVentaPublica", Modulo: permModuleVentaPublica, Accion: permActionCreate, Titulo: "Venta publica (e-commerce)", Grupo: "Operacion diaria y ventas"},

	{PaginaClave: "linkClientes", Modulo: permModuleClientes, Accion: permActionCreate, Titulo: "Clientes y CRM basico", Grupo: "Gestion de Relaciones con Clientes (CRM)"},
	{PaginaClave: "linkCRMComercial", Modulo: permModuleCRMUnificado, Accion: permActionCreate, Titulo: "CRM unificado", Grupo: "Gestion de Relaciones con Clientes (CRM)"},

	{PaginaClave: "linkFinanzas", Modulo: permModuleFinanzas, Accion: permActionCreate, Titulo: "Centro financiero y contable", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkFinanzasMain", Modulo: permModuleFinanzas, Accion: permActionCreate, Titulo: "Finanzas operativas", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkEgresosIngresos", Modulo: permModuleFinanzas, Accion: permActionCreate, Titulo: "Egresos e ingresos", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkEgresos", Modulo: permModuleFinanzas, Accion: permActionCreate, Titulo: "Egresos", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkIngresos", Modulo: permModuleFinanzas, Accion: permActionCreate, Titulo: "Ingresos", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkCorteCaja", Modulo: permModuleFinanzas, Accion: permActionCreate, Titulo: "Corte de caja", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkCreditos", Modulo: permModuleFinanzas, Accion: permActionCreate, Titulo: "Creditos y cartera", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkCreditosMenu", Modulo: permModuleFinanzas, Accion: permActionCreate, Titulo: "Creditos y cartera", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkPropinas", Modulo: permModuleFinanzas, Accion: permActionCreate, Titulo: "Propinas", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkComisiones", Modulo: permModuleFinanzas, Accion: permActionCreate, Titulo: "Comisiones de personal", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkContabilidadColombia", Modulo: permModuleContabilidadCO, Accion: permActionCreate, Titulo: "Contabilidad Colombia NIIF/DIAN", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkContabilidadColombiaAvanzada", Modulo: permModuleContabilidadCOAv, Accion: permActionCreate, Titulo: "Suite contable Colombia avanzada", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkCentrosCosto", Modulo: permModuleCentrosCosto, Accion: permActionCreate, Titulo: "Centros de costo y rentabilidad", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkCentrosCostoMenu", Modulo: permModuleCentrosCosto, Accion: permActionCreate, Titulo: "Centros de costo y rentabilidad", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkCierreFiscal", Modulo: permModuleCierreFiscal, Accion: permActionApprove, Titulo: "Cierre y bloqueo fiscal", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkCierreFiscalMenu", Modulo: permModuleCierreFiscal, Accion: permActionApprove, Titulo: "Cierre y bloqueo fiscal", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkActivosFijosNIIF", Modulo: permModuleActivosFijosNIIF, Accion: permActionCreate, Titulo: "Activos fijos e intangibles NIIF/Fiscal", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkActivosFijosNIIFMenu", Modulo: permModuleActivosFijosNIIF, Accion: permActionCreate, Titulo: "Activos fijos e intangibles NIIF/Fiscal", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkDeclaracionesTributarias", Modulo: permModuleDeclaracionesTrib, Accion: permActionCreate, Titulo: "Declaraciones tributarias Colombia", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkDeclaracionesTributariasMenu", Modulo: permModuleDeclaracionesTrib, Accion: permActionCreate, Titulo: "Declaraciones tributarias Colombia", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkTesoreriaPresupuesto", Modulo: permModuleTesoreria, Accion: permActionCreate, Titulo: "Tesoreria y presupuesto", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkBancosPagos", Modulo: permModuleBancosPagos, Accion: permActionCreate, Titulo: "Bancos y pagos masivos", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkCobranza", Modulo: permModuleCobranza, Accion: permActionCreate, Titulo: "Gestion de cobranza", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkCobranzaMenu", Modulo: permModuleCobranza, Accion: permActionCreate, Titulo: "Gestion de cobranza", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkPortalContador", Modulo: permModulePortalContador, Accion: permActionCreate, Titulo: "Portal contador", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkPortalContadorMenu", Modulo: permModulePortalContador, Accion: permActionCreate, Titulo: "Portal contador", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkPortalTercerosCertificados", Modulo: permModulePortalTerceros, Accion: permActionCreate, Titulo: "Portal de terceros y certificados tributarios", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkPortalTercerosCertificadosMenu", Modulo: permModulePortalTerceros, Accion: permActionCreate, Titulo: "Portal de terceros y certificados tributarios", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkCumplimientoKYC", Modulo: permModuleCumplimientoKYC, Accion: permActionApprove, Titulo: "Cumplimiento KYC/KYB y LAFT", Grupo: "Centro financiero y contable"},
	{PaginaClave: "linkNominaSueldos", Modulo: permModuleNominaSueldos, Accion: permActionCreate, Titulo: "Nomina y sueldos", Grupo: "Personas y activos"},
	{PaginaClave: "linkNominaMenu", Modulo: permModuleNominaSueldos, Accion: permActionCreate, Titulo: "Nomina de sueldos", Grupo: "Centro financiero y contable"},

	{PaginaClave: "linkFacturacionElectronica", Modulo: permModuleFacturacion, Accion: permActionCreate, Titulo: "Facturacion electronica (emitir)", Grupo: "Facturacion electronica"},
	{PaginaClave: "linkFacturacionMain", Modulo: permModuleFacturacion, Accion: permActionCreate, Titulo: "Facturacion electronica", Grupo: "Facturacion electronica"},
	{PaginaClave: "linkFacturasElectronicas", Modulo: permModuleFacturacion, Accion: permActionRead, Titulo: "Documentos y consultas FE", Grupo: "Facturacion electronica"},
	{PaginaClave: "linkImpuestos", Modulo: permModuleFacturacion, Accion: permActionUpdate, Titulo: "Impuestos", Grupo: "Administracion y configuracion"},
	{PaginaClave: "linkFrecuenciaFE", Modulo: permModuleFacturacion, Accion: permActionApprove, Titulo: "Frecuencia FE", Grupo: "Administracion y configuracion"},

	{PaginaClave: "linkReportes", Modulo: permModuleReportes, Accion: permActionRead, Titulo: "Centro de reportes", Grupo: "Analisis y control"},
	{PaginaClave: "linkCalculadora", Modulo: permModuleFinanzas, Accion: permActionRead, Titulo: "Calculadora financiera", Grupo: "Centro financiero y contable"},

	{PaginaClave: "linkUsuarios", Modulo: permModuleSeguridad, Accion: permActionUpdate, Titulo: "Usuarios y accesos", Grupo: "Personas y activos"},
	{PaginaClave: "linkPortalUsuarios", Modulo: permModuleSeguridad, Accion: permActionRead, Titulo: "Portal de usuarios", Grupo: "Personas y activos"},
	{PaginaClave: "linkMiHorario", Modulo: permModuleHorariosTrab, Accion: permActionRead, Titulo: "Mi horario", Grupo: "Personas y activos"},
	{PaginaClave: "linkHorariosTrabajadores", Modulo: permModuleHorariosTrab, Accion: permActionUpdate, Titulo: "Horarios laborales", Grupo: "Personas y activos"},
	{PaginaClave: "linkAsistenciaEmpleados", Modulo: permModuleAsistenciaEmpleados, Accion: permActionUpdate, Titulo: "Asistencia de empleados", Grupo: "Personas y activos"},
	{PaginaClave: "linkCarnets", Modulo: permModuleCarnets, Accion: permActionCreate, Titulo: "Carnets de empleados y usuarios", Grupo: "Personas y activos"},
	{PaginaClave: "linkVehiculosRegistro", Modulo: permModuleVehiculosRegistro, Accion: permActionCreate, Titulo: "Registro de vehiculos", Grupo: "Personas y activos"},
	{PaginaClave: "linkHojaVidaOperativa", Modulo: permModuleHojaVidaOperativa, Accion: permActionUpdate, Titulo: "Hoja de vida operativa", Grupo: "Personas y activos"},
	{PaginaClave: "linkUbicacionGPS", Modulo: permModuleUbicacionGPS, Accion: permActionCreate, Titulo: "Ubicacion / GPS (activos)", Grupo: "Personas y activos"},

	{PaginaClave: "linkAuditoria", Modulo: permModuleAuditoria, Accion: permActionRead, Titulo: "Auditoria de acciones", Grupo: "Analisis y control"},
	{PaginaClave: "linkCalidadProcesos", Modulo: permModuleCalidadProcesos, Accion: permActionCreate, Titulo: "Calidad, procesos y no conformidades", Grupo: "Analisis y control"},
	{PaginaClave: "linkBackups", Modulo: permModuleBackups, Accion: permActionApprove, Titulo: "Backups empresariales", Grupo: "Analisis y control"},

	{PaginaClave: "linkDocumentosOnlyOffice", Modulo: permModuleDocumentosOnlyOffice, Accion: permActionRead, Titulo: "Documentos OnlyOffice", Grupo: "Documentos, nube y soporte"},
	{PaginaClave: "linkGestionDocumental", Modulo: permModuleGestionDocumental, Accion: permActionCreate, Titulo: "Gestion documental y aprobaciones", Grupo: "Documentos, nube y soporte"},
	{PaginaClave: "linkContratosObligaciones", Modulo: permModuleContratosOblig, Accion: permActionCreate, Titulo: "Contratos y obligaciones", Grupo: "Documentos, nube y soporte"},
	{PaginaClave: "linkSoporteRemoto", Modulo: permModuleSeguridad, Accion: permActionApprove, Titulo: "Soporte remoto", Grupo: "Documentos, nube y soporte"},

	{PaginaClave: "linkConfiguracion", Modulo: permModuleSeguridad, Accion: permActionUpdate, Titulo: "Configuracion de empresa", Grupo: "Administracion y configuracion"},
	{PaginaClave: "linkConfiguracionMain", Modulo: permModuleSeguridad, Accion: permActionUpdate, Titulo: "Configuracion general", Grupo: "Administracion y configuracion"},
	{PaginaClave: "linkConfiguracionImpresora", Modulo: permModuleSeguridad, Accion: permActionUpdate, Titulo: "Configuracion de impresora", Grupo: "Administracion y configuracion"},
	{PaginaClave: "linkConfiguracionPermisos", Modulo: permModuleSeguridad, Accion: permActionApprove, Titulo: "Permisos y roles", Grupo: "Administracion y configuracion"},
	{PaginaClave: "linkConfiguracionGuiada", Modulo: permModuleSeguridad, Accion: permActionUpdate, Titulo: "Configuracion guiada con IA", Grupo: "Administracion y configuracion"},
	{PaginaClave: "linkConfiguracionChatFlotante", Modulo: permModuleChatTareas, Accion: permActionUpdate, Titulo: "Configurar chat y robot", Grupo: "Administracion y configuracion"},
	{PaginaClave: "linkConfiguracionCarritoEmpresa", Modulo: permModuleVentaPublica, Accion: permActionApprove, Titulo: "Carrito unificado", Grupo: "Administracion y configuracion"},
	{PaginaClave: "linkConfiguracionAvanzada", Modulo: permModuleSeguridad, Accion: permActionUpdate, Titulo: "Documento y formato", Grupo: "Administracion y configuracion"},
	{PaginaClave: "linkConfigEstaciones", Modulo: permModuleVentas, Accion: permActionApprove, Titulo: "Aprobacion: configuracion de estaciones", Grupo: "Administracion y configuracion"},
	{PaginaClave: "linkConfiguracionSensoresRaspberry", Modulo: permModuleControlElectrico, Accion: permActionUpdate, Titulo: "Raspberry Pi y sensores", Grupo: "Administracion y configuracion"},
	{PaginaClave: "linkControlElectrico", Modulo: permModuleControlElectrico, Accion: permActionUpdate, Titulo: "Control electrico por habitacion", Grupo: "Administracion y configuracion"},
	{PaginaClave: "linkRadioOnline", Modulo: permModuleSeguridad, Accion: permActionRead, Titulo: "Radio online", Grupo: "Administracion y configuracion"},
	{PaginaClave: "linkERPExtendido", Modulo: permModuleSeguridad, Accion: permActionUpdate, Titulo: "Integraciones / ERP extendido", Grupo: "Administracion y configuracion"},
	{PaginaClave: "linkERPExtendidoMenu", Modulo: permModuleSeguridad, Accion: permActionUpdate, Titulo: "ERP extendido", Grupo: "Administracion y configuracion"},
	{PaginaClave: "linkChatIAGlobal", Modulo: permModuleSeguridad, Accion: permActionRead, Titulo: "Chat IA global (super)", Grupo: "Administracion y configuracion"},
	{PaginaClave: "linkTarifasPorMinutos", Modulo: permModuleVentas, Accion: permActionCreate, Titulo: "Tarifas por minutos", Grupo: "Administracion y configuracion"},
	{PaginaClave: "linkTarifasPorDia", Modulo: permModuleVentas, Accion: permActionCreate, Titulo: "Tarifas por dia", Grupo: "Administracion y configuracion"},
	{PaginaClave: "linkTarifasHotel", Modulo: permModuleVentas, Accion: permActionCreate, Titulo: "Tarifas de hotel", Grupo: "Administracion y configuracion"},
	{PaginaClave: "linkTarifasMotel", Modulo: permModuleVentas, Accion: permActionCreate, Titulo: "Tarifas de motel", Grupo: "Administracion y configuracion"},
	{PaginaClave: "linkHotelTarjetasAcceso", Modulo: permModuleReservasHotel, Accion: permActionCreate, Titulo: "Tarjetas habitacion", Grupo: "Administracion y configuracion"},
}

type permissionModuleMatrixRow struct {
	Modulo   string          `json:"modulo"`
	Read     bool            `json:"read"`
	Create   bool            `json:"create"`
	Update   bool            `json:"update"`
	Delete   bool            `json:"delete"`
	Approve  bool            `json:"approve"`
	Acciones map[string]bool `json:"acciones"`
}

type permissionPageAccessRow struct {
	PaginaClave   string   `json:"pagina_clave"`
	Modulo        string   `json:"modulo,omitempty"`
	Accion        string   `json:"accion,omitempty"`
	AnyModules    []string `json:"any_modules,omitempty"`
	Permitido     bool     `json:"permitido"`
	AlwaysVisible bool     `json:"always_visible,omitempty"`
	Titulo        string   `json:"titulo,omitempty"`
	Grupo         string   `json:"grupo,omitempty"`
}

type permissionSummary struct {
	ModulosTotal        int `json:"modulos_total"`
	ModulosLectura      int `json:"modulos_lectura"`
	ModulosAprobacion   int `json:"modulos_aprobacion"`
	AccionesHabilitadas int `json:"acciones_habilitadas"`
}

type empresaPermisosRolMatriz struct {
	Rol     string                      `json:"rol"`
	Modulos []permissionModuleMatrixRow `json:"modulos"`
	Resumen permissionSummary           `json:"resumen"`
}

type empresaPermisosContextResponse struct {
	EmpresaID        int64                       `json:"empresa_id"`
	AdminEmail       string                      `json:"admin_email"`
	Rol              string                      `json:"rol"`
	RolEfectivo      string                      `json:"rol_efectivo,omitempty"`
	AccionesCatalogo []string                    `json:"acciones_catalogo"`
	Modulos          []permissionModuleMatrixRow `json:"modulos"`
	Paginas          map[string]bool             `json:"paginas,omitempty"`
	Resumen          permissionSummary           `json:"resumen"`
	Licencia         *empresaPermisosLicenciaCtx `json:"licencia,omitempty"`
	VerticalScope    *empresaVerticalScopeCtx    `json:"vertical_scope,omitempty"`
	EmpresaPolicy    *empresaPermisosFinosCtx    `json:"empresa_policy,omitempty"`
	ShareAccess      *empresaCompartidaScopeCtx  `json:"share_access,omitempty"`
	IncluyeMatriz    bool                        `json:"incluye_matriz"`
	MatrizRoles      []empresaPermisosRolMatriz  `json:"matriz_roles,omitempty"`
}

type empresaCompartidaScopeCtx struct {
	Compartida        bool     `json:"compartida"`
	NivelAcceso       string   `json:"nivel_acceso,omitempty"`
	ModulosPermitidos []string `json:"modulos_permitidos,omitempty"`
	CompartidoPor     string   `json:"compartido_por,omitempty"`
	Etiqueta          string   `json:"etiqueta,omitempty"`
}

type empresaPermisosLicenciaCtx struct {
	LicenciaID         int64    `json:"licencia_id,omitempty"`
	Nombre             string   `json:"nombre,omitempty"`
	ModulosHabilitados []string `json:"modulos_habilitados,omitempty"`
	SuperRolHabilitado bool     `json:"super_rol_habilitado"`
	RestringeModulos   bool     `json:"restringe_modulos"`
}

type empresaVerticalScopeCtx struct {
	Restringe         bool     `json:"restringe"`
	TipoEmpresaID     int64    `json:"tipo_empresa_id,omitempty"`
	TipoEmpresaNombre string   `json:"tipo_empresa_nombre,omitempty"`
	ModulosPermitidos []string `json:"modulos_permitidos,omitempty"`
	Fuente            string   `json:"fuente,omitempty"`
}

type empresaPermisosFinosCtx struct {
	ReglasModulo int  `json:"reglas_modulo"`
	ReglasPagina int  `json:"reglas_pagina"`
	Activo       bool `json:"activo"`
}

type empresaPermisoModuloPayload struct {
	Modulo    string `json:"modulo"`
	Accion    string `json:"accion"`
	Permitido bool   `json:"permitido"`
}

type empresaPermisoPaginaPayload struct {
	PaginaClave string `json:"pagina_clave"`
	Permitido   bool   `json:"permitido"`
}

type empresaPermisosFinosPayload struct {
	EmpresaID      int64                         `json:"empresa_id"`
	PermisosModulo []empresaPermisoModuloPayload `json:"permisos_modulo"`
	PermisosPagina []empresaPermisoPaginaPayload `json:"permisos_pagina"`
}

// EmpresaPermisosContextoHandler expone el contexto de permisos efectivo por rol/modulo.
// Endpoint recomendado: GET /api/empresa/permisos_contexto?empresa_id={id}[&include_matrix=1]
func EmpresaPermisosContextoHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}

		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		adminEmail := strings.ToLower(strings.TrimSpace(adminEmailFromRequest(r)))
		role := normalizePermissionRole(adminRoleFromRequest(r))
		if role == "" && dbSuper != nil && adminEmail != "" && adminEmail != "sistema" {
			admin, err := dbpkg.GetAdminByEmail(dbSuper, adminEmail)
			if err == nil && admin != nil {
				role = normalizePermissionRole(admin.Role)
			} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
				log.Printf("[authz] permisos_contexto get admin email=%s error: %v", adminEmail, err)
			}
		}
		if role == "" {
			role = "sin_rol"
		}

		licenciaPolicy, err := dbpkg.GetLicenciaPermisoPolicyByEmpresa(dbSuper, empresaID)
		if err != nil {
			log.Printf("[authz] permisos_contexto licencia empresa=%d error: %v", empresaID, err)
		}

		allowedModules, allowedModulesList := parseLicenciaModulosCSV("")
		if licenciaPolicy != nil {
			allowedModules, allowedModulesList = parseLicenciaModulosCSV(licenciaPolicy.ModulosHabilitados)
		}
		effectiveRole := resolveEffectiveRoleByLicencia(role, licenciaPolicy)

		modulos := buildPermissionModuleMatrixForRoleDynamic(dbSuper, effectiveRole)
		modulos = applyLicenciaRestriccionesToModuleRows(modulos, allowedModules)
		verticalScope := resolveEmpresaVerticalScope(dbSuper, empresaID, licenciaPolicy)
		modulos = applyEmpresaVerticalScopeToModuleRows(modulos, verticalScope)
		empresaModuleOverrides, empresaPageOverrides, empresaPolicyCtx := loadEmpresaPermissionOverrides(dbSuper, empresaID)
		modulos = applyEmpresaRestriccionesToModuleRows(modulos, empresaModuleOverrides)
		sharedAccess, sharedAccessErr := dbpkg.GetActiveAdminEmpresaCompartidaAcceso(dbSuper, empresaID, adminEmail)
		if sharedAccessErr != nil {
			log.Printf("[authz] permisos_contexto acceso_compartido empresa=%d email=%s error: %v", empresaID, adminEmail, sharedAccessErr)
		}
		if sharedAccessErr == nil {
			modulos = applyAdminEmpresaCompartidaScopeToModuleRows(modulos, sharedAccess)
		}
		shareCtx := adminEmpresaCompartidaScopeContext(sharedAccess)
		paginas := buildPermissionPagesMapForRoleDynamic(dbSuper, effectiveRole, modulos)
		paginas = applyEmpresaPageRestrictionsToMap(paginas, empresaPageOverrides)

		var licenciaCtx *empresaPermisosLicenciaCtx
		if licenciaPolicy != nil {
			licenciaCtx = &empresaPermisosLicenciaCtx{
				LicenciaID:         licenciaPolicy.LicenciaID,
				Nombre:             strings.TrimSpace(licenciaPolicy.Nombre),
				ModulosHabilitados: append([]string{}, allowedModulesList...),
				SuperRolHabilitado: licenciaPolicy.SuperRolHabilitado,
				RestringeModulos:   len(allowedModules) > 0,
			}
		}

		resp := empresaPermisosContextResponse{
			EmpresaID:        empresaID,
			AdminEmail:       adminEmail,
			Rol:              role,
			RolEfectivo:      effectiveRole,
			AccionesCatalogo: append([]string{}, permissionActionsCatalogOrdered...),
			Modulos:          modulos,
			Paginas:          paginas,
			Resumen:          summarizePermissionModules(modulos),
			Licencia:         licenciaCtx,
			VerticalScope:    verticalScope.toContext(),
			EmpresaPolicy:    empresaPolicyCtx,
			ShareAccess:      shareCtx,
			IncluyeMatriz:    false,
		}

		if queryBool(r, "include_matrix") {
			resp.IncluyeMatriz = true
			resp.MatrizRoles = make([]empresaPermisosRolMatriz, 0, len(permissionRolesCatalogOrdered))
			for _, catalogRole := range permissionRolesCatalogOrdered {
				rows := buildPermissionModuleMatrixForRoleDynamic(dbSuper, catalogRole)
				rows = applyLicenciaRestriccionesToModuleRows(rows, allowedModules)
				rows = applyEmpresaVerticalScopeToModuleRows(rows, verticalScope)
				rows = applyEmpresaRestriccionesToModuleRows(rows, empresaModuleOverrides)
				resp.MatrizRoles = append(resp.MatrizRoles, empresaPermisosRolMatriz{
					Rol:     catalogRole,
					Modulos: rows,
					Resumen: summarizePermissionModules(rows),
				})
			}
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

// EmpresaPermisosFinosHandler administra el techo fino de modulos, acciones y paginas para una empresa.
func EmpresaPermisosFinosHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := dbpkg.EnsureEmpresaPermisosFinosSchema(dbSuper); err != nil {
			http.Error(w, "failed to ensure empresa permisos finos schema: "+err.Error(), http.StatusInternalServerError)
			return
		}

		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			modulos := buildEmpresaPermisosDefaultModuleRows()
			moduleItems, err := dbpkg.ListEmpresaPermisosModuloByEmpresaID(dbSuper, empresaID)
			if err != nil {
				http.Error(w, "failed to load empresa modulo permisos: "+err.Error(), http.StatusInternalServerError)
				return
			}
			moduleOverrides := make(map[string]bool, len(moduleItems))
			for _, item := range moduleItems {
				moduleOverrides[permissionModuleActionKey(item.Modulo, item.Accion)] = item.Permitido
			}
			modulos = applyEmpresaRestriccionesToModuleRows(modulos, moduleOverrides)

			pageItems, err := dbpkg.ListEmpresaPermisosPaginaByEmpresaID(dbSuper, empresaID)
			if err != nil {
				http.Error(w, "failed to load empresa pagina permisos: "+err.Error(), http.StatusInternalServerError)
				return
			}
			pageOverrides := make(map[string]bool, len(pageItems))
			for _, item := range pageItems {
				pageOverrides[strings.TrimSpace(item.PaginaClave)] = item.Permitido
			}
			paginas := buildPermissionPagesCatalogFromModuleRows(modulos, pageOverrides)

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"empresa_id":          empresaID,
				"acciones_catalogo":   append([]string{}, permissionActionsCatalogOrdered...),
				"acciones_etiqueta":   PermissionActionDisplayNameMap(),
				"modulos_catalogo":    append([]string{}, permissionModulesCatalogOrdered...),
				"modulos_etiqueta":    PermissionModuleDisplayNameMap(),
				"modulos":             modulos,
				"paginas":             paginas,
				"reglas_modulo":       len(moduleItems),
				"reglas_pagina":       len(pageItems),
				"comportamiento_base": "sin reglas guardadas, la empresa no restringe el catalogo; licencia y rol siguen aplicando",
			})
			return

		case http.MethodPut:
			var payload empresaPermisosFinosPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			empresaID := payload.EmpresaID
			if empresaID <= 0 {
				if qID, err := parseOptionalInt64Query(r, "empresa_id"); err == nil && qID > 0 {
					empresaID = qID
				}
			}
			if empresaID <= 0 {
				http.Error(w, "empresa_id required", http.StatusBadRequest)
				return
			}

			moduleRows := make([]dbpkg.EmpresaPermisoModulo, 0, len(payload.PermisosModulo))
			for _, item := range payload.PermisosModulo {
				moduleRows = append(moduleRows, dbpkg.EmpresaPermisoModulo{
					EmpresaID: empresaID,
					Modulo:    strings.ToLower(strings.TrimSpace(item.Modulo)),
					Accion:    strings.ToUpper(strings.TrimSpace(item.Accion)),
					Permitido: item.Permitido,
				})
			}

			pageRows := make([]dbpkg.EmpresaPermisoPagina, 0, len(payload.PermisosPagina))
			for _, item := range payload.PermisosPagina {
				pageRows = append(pageRows, dbpkg.EmpresaPermisoPagina{
					EmpresaID:   empresaID,
					PaginaClave: strings.TrimSpace(item.PaginaClave),
					Permitido:   item.Permitido,
				})
			}

			if err := dbpkg.ReplaceEmpresaPermisosFinos(dbSuper, empresaID, moduleRows, pageRows, adminEmailFromRequest(r)); err != nil {
				http.Error(w, "failed to save empresa permisos finos: "+err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusNoContent)
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

// WithEmpresaVentasPermissions aplica control de alcance por empresa y permisos por rol para ventas.
func WithEmpresaVentasPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleVentas, resolveVentasPermissionAction, next)
}

// WithEmpresaInventarioPermissions aplica control de alcance por empresa y permisos por rol para inventario.
func WithEmpresaInventarioPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleInventario, resolveInventarioPermissionAction, next)
}

// WithEmpresaFinanzasPermissions aplica control de alcance por empresa y permisos por rol para finanzas.
func WithEmpresaFinanzasPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleFinanzas, resolveFinanzasPermissionAction, next)
}

// WithEmpresaContabilidadColombiaPermissions aplica permisos independientes para contabilidad legal colombiana.
func WithEmpresaContabilidadColombiaPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleContabilidadCO, resolveContabilidadColombiaPermissionAction, next)
}

// WithEmpresaContabilidadColombiaAvanzadaPermissions aplica permisos para exogena, nomina DIAN, activos y libros oficiales.
func WithEmpresaContabilidadColombiaAvanzadaPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleContabilidadCOAv, resolveContabilidadColombiaAvanzadaPermissionAction, next)
}

// WithEmpresaCentrosCostoPermissions aplica permisos independientes para rentabilidad por sucursal, area o unidad.
func WithEmpresaCentrosCostoPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleCentrosCosto, resolveCentrosCostoPermissionAction, next)
}

// WithEmpresaCierreFiscalPermissions aplica permisos para bloqueo fiscal y reapertura de periodos.
func WithEmpresaCierreFiscalPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleCierreFiscal, resolveCierreFiscalPermissionAction, next)
}

// WithEmpresaActivosFijosNIIFPermissions aplica permisos para activos fijos e intangibles NIIF/Fiscal.
func WithEmpresaActivosFijosNIIFPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleActivosFijosNIIF, resolveActivosFijosNIIFPermissionAction, next)
}

// WithEmpresaDeclaracionesTributariasPermissions aplica permisos para liquidacion y presentacion de impuestos.
func WithEmpresaDeclaracionesTributariasPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleDeclaracionesTrib, resolveDeclaracionesTributariasPermissionAction, next)
}

// WithEmpresaClientesPermissions aplica control de alcance por empresa y permisos por rol para clientes.
func WithEmpresaClientesPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleClientes, resolveClientesPermissionAction, next)
}

// WithEmpresaComprasPermissions aplica control de alcance por empresa y permisos por rol para compras/proveedores.
func WithEmpresaComprasPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleCompras, resolveComprasPermissionAction, next)
}

// WithEmpresaSoportesComprasIAPermissions aplica permisos independientes para captura OCR/IA de compras y gastos.
func WithEmpresaSoportesComprasIAPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleSoportesComprasIA, resolveSoportesComprasIAPermissionAction, next)
}

// WithEmpresaFacturacionPermissions aplica control de alcance por empresa y permisos por rol para facturacion.
func WithEmpresaFacturacionPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleFacturacion, resolveFacturacionPermissionAction, next)
}

// WithEmpresaAIUConstruccionPermissions aplica permisos independientes para contratos de obra y AIU.
func WithEmpresaAIUConstruccionPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleAIUConstruccion, resolveAIUConstruccionPermissionAction, next)
}

// WithEmpresaCobranzaPermissions aplica permisos independientes para gestion de cobranza.
func WithEmpresaCobranzaPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleCobranza, resolveCobranzaPermissionAction, next)
}

// WithEmpresaPortalContadorPermissions aplica permisos independientes para oficina virtual de contadores.
func WithEmpresaPortalContadorPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModulePortalContador, resolvePortalContadorPermissionAction, next)
}

// WithEmpresaPortalTercerosPermissions aplica permisos para certificados tributarios y autoservicio de terceros.
func WithEmpresaPortalTercerosPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModulePortalTerceros, resolvePortalTercerosPermissionAction, next)
}

// WithEmpresaSeguridadPermissions aplica control de alcance por empresa y permisos por rol para seguridad/usuarios.
func WithEmpresaSeguridadPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleSeguridad, resolveSeguridadPermissionAction, next)
}

// WithEmpresaChatTareasPermissions aplica permisos independientes para chat, tareas y agenda.
func WithEmpresaChatTareasPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleChatTareas, resolveVerticalPermissionAction, next)
}

// WithEmpresaReservasHotelPermissions aplica permisos para reservas hoteleras y tarjetas de acceso.
func WithEmpresaReservasHotelPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleReservasHotel, resolveVerticalPermissionAction, next)
}

// WithEmpresaHorariosTrabajadoresPermissions aplica permisos para programacion de turnos laborales.
func WithEmpresaHorariosTrabajadoresPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleHorariosTrab, resolveVerticalPermissionAction, next)
}

// WithEmpresaAsistenciaEmpleadosPermissions aplica permisos para control de asistencia.
func WithEmpresaAsistenciaEmpleadosPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleAsistenciaEmpleados, resolveVerticalPermissionAction, next)
}

// WithEmpresaVehiculosRegistroPermissions aplica permisos para registro y salida de vehiculos.
func WithEmpresaVehiculosRegistroPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleVehiculosRegistro, resolveVerticalPermissionAction, next)
}

// WithEmpresaHojaVidaOperativaPermissions aplica permisos para historial operativo de activos.
func WithEmpresaHojaVidaOperativaPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleHojaVidaOperativa, resolveVerticalPermissionAction, next)
}

// WithEmpresaUbicacionGPSPermissions aplica permisos para dispositivos, recorridos y mapa GPS.
func WithEmpresaUbicacionGPSPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleUbicacionGPS, resolveVerticalPermissionAction, next)
}

// WithEmpresaAuditoriaPermissions aplica permisos para auditoria empresarial.
func WithEmpresaAuditoriaPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleAuditoria, resolveVerticalPermissionAction, next)
}

// WithEmpresaBackupsPermissions aplica permisos para snapshots, restauracion y descargas de backups.
func WithEmpresaBackupsPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleBackups, resolveVerticalPermissionAction, next)
}

// WithEmpresaDocumentosOnlyOfficePermissions aplica permisos para documentos OnlyOffice.
func WithEmpresaDocumentosOnlyOfficePermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleDocumentosOnlyOffice, resolveVerticalPermissionAction, next)
}

// WithEmpresaReportesPermissions aplica permisos para analitica y reportes ejecutivos.
func WithEmpresaReportesPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleReportes, resolveFinanzasPermissionAction, next)
}

// WithEmpresaNominaSueldosPermissions aplica permisos independientes para nomina y sueldos.
func WithEmpresaNominaSueldosPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleNominaSueldos, resolveFinanzasPermissionAction, next)
}

// WithEmpresaCRMUnificadoPermissions aplica permisos para embudo comercial, leads y seguimiento.
func WithEmpresaCRMUnificadoPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleCRMUnificado, resolveClientesPermissionAction, next)
}

// WithEmpresaVentaPublicaPermissions aplica permisos independientes para venta publica y carta.
func WithEmpresaVentaPublicaPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleVentaPublica, resolveVerticalPermissionAction, next)
}

// WithEmpresaGimnasioPermissions aplica permisos independientes para gimnasio.
func WithEmpresaGimnasioPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleGimnasio, resolveVerticalPermissionAction, next)
}

// WithEmpresaTaxiSystemPermissions aplica permisos independientes para taxi system.
func WithEmpresaTaxiSystemPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleTaxiSystem, resolveVerticalPermissionAction, next)
}

// WithEmpresaDomiciliosPermissions aplica permisos independientes para domicilios.
func WithEmpresaDomiciliosPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleDomicilios, resolveVerticalPermissionAction, next)
}

// WithEmpresaParqueaderoPermissions aplica permisos independientes para parqueadero.
func WithEmpresaParqueaderoPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleParqueadero, resolveVerticalPermissionAction, next)
}

// WithEmpresaApartamentosTuristicosPermissions aplica permisos independientes para apartamentos turisticos.
func WithEmpresaApartamentosTuristicosPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleApartTuristicos, resolveVerticalPermissionAction, next)
}

// WithEmpresaPropiedadHorizontalPermissions aplica permisos independientes para copropiedades.
func WithEmpresaPropiedadHorizontalPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModulePropiedadHorizontal, resolveVerticalPermissionAction, next)
}

// WithEmpresaAlquileresPermissions aplica permisos independientes para alquileres.
func WithEmpresaAlquileresPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleAlquileres, resolveVerticalPermissionAction, next)
}

// WithEmpresaOdontologiaPermissions aplica permisos independientes para odontologia.
func WithEmpresaOdontologiaPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleOdontologia, resolveVerticalPermissionAction, next)
}

// WithEmpresaTurnosAtencionPermissions aplica permisos independientes para turnos.
func WithEmpresaTurnosAtencionPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleTurnos, resolveVerticalPermissionAction, next)
}

// WithEmpresaControlElectricoPermissions aplica permisos independientes para control electrico.
func WithEmpresaControlElectricoPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleControlElectrico, resolveControlElectricoPermissionAction, next)
}

// WithEmpresaCarnetsPermissions aplica permisos independientes para carnets empresariales.
func WithEmpresaCarnetsPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleCarnets, resolveVerticalPermissionAction, next)
}

// WithEmpresaProduccionMRPPermissions aplica permisos independientes para produccion y planeacion MRP.
func WithEmpresaProduccionMRPPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleProduccionMRP, resolveVerticalPermissionAction, next)
}

// WithEmpresaWMSPermissions aplica permisos para logistica avanzada, picking, packing y despachos.
func WithEmpresaWMSPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleLogisticaWMS, resolveVerticalPermissionAction, next)
}

// WithEmpresaTesoreriaPresupuestoPermissions aplica permisos para tesoreria, bancos y presupuesto.
func WithEmpresaTesoreriaPresupuestoPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleTesoreria, resolveFinanzasPermissionAction, next)
}

// WithEmpresaImportacionesCosteoPermissions aplica permisos para importaciones, nacionalizacion y costo aterrizado.
func WithEmpresaImportacionesCosteoPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleImportaciones, resolveComprasPermissionAction, next)
}

// WithEmpresaBancosPagosPermissions aplica permisos para conciliacion bancaria y pagos masivos.
func WithEmpresaBancosPagosPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleBancosPagos, resolveFinanzasPermissionAction, next)
}

// WithEmpresaGestionDocumentalPermissions aplica permisos para expedientes y aprobaciones documentales.
func WithEmpresaGestionDocumentalPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleGestionDocumental, resolveVerticalPermissionAction, next)
}

// WithEmpresaCumplimientoKYCPermissions aplica permisos para debida diligencia y riesgo LAFT.
func WithEmpresaCumplimientoKYCPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleCumplimientoKYC, resolveVerticalPermissionAction, next)
}

// WithEmpresaContratosObligacionesPermissions aplica permisos para contratos, polizas y vencimientos.
func WithEmpresaContratosObligacionesPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleContratosOblig, resolveVerticalPermissionAction, next)
}

// WithEmpresaCalidadProcesosPermissions aplica permisos para auditorias, procesos y no conformidades.
func WithEmpresaCalidadProcesosPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleCalidadProcesos, resolveVerticalPermissionAction, next)
}

// WithEmpresaDrogueriaFarmaciaPermissions aplica permisos para operacion sanitaria, lotes y dispensacion.
func WithEmpresaDrogueriaFarmaciaPermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return withEmpresaRolePermissions(dbEmp, dbSuper, permModuleDrogueriaFarmacia, resolveVerticalPermissionAction, next)
}

// WithEmpresaPublicScope aplica validacion minima de alcance por empresa para endpoints publicos
// que no pueden exigir autenticacion previa (por ejemplo login y primer establecimiento de password).
func WithEmpresaPublicScope(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID := extractEmpresaIDForPermissions(r)
		if empresaID <= 0 {
			next.ServeHTTP(w, r)
			return
		}
		if err := validateEmpresaIDConsistency(r, empresaID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), "empresaID", empresaID)
		r = r.WithContext(ctx)
		w.Header().Set("X-Empresa-ID", strconv.FormatInt(empresaID, 10))

		next.ServeHTTP(w, r)
	}
}

// WithEmpresaSelfServicePermissions protege endpoints de autoservicio del usuario
// autenticado. Valida empresa y alcance, pero no exige permisos administrativos
// de creacion/edicion sobre el modulo.
func WithEmpresaSelfServicePermissions(dbEmp, dbSuper *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID := extractEmpresaIDForPermissions(r)
		if empresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		if err := validateEmpresaIDConsistency(r, empresaID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		adminEmail := strings.ToLower(strings.TrimSpace(adminEmailFromRequest(r)))
		if adminEmail == "" || adminEmail == "sistema" {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		canAccess, err := dbpkg.CanAdminAccessEmpresaIA(dbEmp, dbSuper, adminEmail, empresaID)
		if err != nil {
			log.Printf("[authz] self-service empresa=%d email=%s error: %v", empresaID, adminEmail, err)
			http.Error(w, "No se pudo validar el alcance del usuario", http.StatusInternalServerError)
			return
		}
		if !canAccess {
			http.Error(w, "forbidden: empresa_id fuera del alcance del usuario autenticado", http.StatusForbidden)
			return
		}
		ctx := context.WithValue(r.Context(), "empresaID", empresaID)
		r = r.WithContext(ctx)
		w.Header().Set("X-Empresa-ID", strconv.FormatInt(empresaID, 10))
		next.ServeHTTP(w, r)
	}
}

func empresaRateLimitScopeForRequest(r *http.Request) string {
	path := ""
	if r != nil && r.URL != nil {
		path = strings.TrimSpace(r.URL.Path)
	}
	if strings.HasPrefix(path, "/api/empresa/db_admin") {
		return "db_admin"
	}
	return "api"
}

func empresaRateLimitMaxForRequest(dbSuper *sql.DB, r *http.Request) int64 {
	if dbSuper == nil {
		if empresaRateLimitScopeForRequest(r) == "db_admin" {
			return defaultEmpresaDBQueriesPerMinute
		}
		return defaultEmpresaAPIRequestsPerMinute
	}
	switch empresaRateLimitScopeForRequest(r) {
	case "db_admin":
		value, _, _, err := getLimitacionInt64(dbSuper, superEmpresaLimitDBQueriesPerMinuteKey, defaultEmpresaDBQueriesPerMinute)
		if err != nil {
			log.Printf("[rate_limit] no se pudo leer limite db_admin: %v", err)
			return defaultEmpresaDBQueriesPerMinute
		}
		return value
	default:
		value, _, _, err := getLimitacionInt64(dbSuper, superEmpresaLimitAPIRequestsPerMinuteKey, defaultEmpresaAPIRequestsPerMinute)
		if err != nil {
			log.Printf("[rate_limit] no se pudo leer limite api: %v", err)
			return defaultEmpresaAPIRequestsPerMinute
		}
		return value
	}
}

func checkEmpresaRateLimitAt(now time.Time, empresaID int64, scope string, maxPerMinute int64) (allowed bool, remaining int64, retryAfterSeconds int64, current int64) {
	if empresaID <= 0 || maxPerMinute <= 0 {
		return true, 0, 0, 0
	}
	scope = strings.TrimSpace(strings.ToLower(scope))
	if scope == "" {
		scope = "api"
	}
	if now.IsZero() {
		now = time.Now()
	}
	windowStart := now.Truncate(time.Minute)
	key := scope + ":" + strconv.FormatInt(empresaID, 10)

	empresaRateLimitMu.Lock()
	defer empresaRateLimitMu.Unlock()

	bucket := empresaRateLimitBuckets[key]
	if bucket.WindowStart.IsZero() || !bucket.WindowStart.Equal(windowStart) {
		bucket = empresaRateLimitBucket{WindowStart: windowStart}
	}
	if bucket.Count >= maxPerMinute {
		retryAfter := int64(bucket.WindowStart.Add(time.Minute).Sub(now).Seconds())
		if retryAfter < 1 {
			retryAfter = 1
		}
		empresaRateLimitBuckets[key] = bucket
		return false, 0, retryAfter, bucket.Count
	}

	bucket.Count++
	empresaRateLimitBuckets[key] = bucket
	remaining = maxPerMinute - bucket.Count
	if remaining < 0 {
		remaining = 0
	}
	return true, remaining, 0, bucket.Count
}

func applyEmpresaRateLimitHeaders(w http.ResponseWriter, limit, remaining, retryAfterSeconds int64) {
	if w == nil {
		return
	}
	if limit > 0 {
		w.Header().Set("X-Empresa-RateLimit-Limit", strconv.FormatInt(limit, 10))
		w.Header().Set("X-Empresa-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
	}
	if retryAfterSeconds > 0 {
		w.Header().Set("Retry-After", strconv.FormatInt(retryAfterSeconds, 10))
	}
}

func withEmpresaRolePermissions(dbEmp, dbSuper *sql.DB, module string, resolveAction func(*http.Request) string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		defer func() {
			dbpkg.PerfLogf("[perf][authz] module=%s method=%s path=%s dur=%s", module, r.Method, r.URL.Path, time.Since(startedAt))
		}()
		empresaID := extractEmpresaIDForPermissions(r)
		if empresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		if err := validateEmpresaIDConsistency(r, empresaID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		action := defaultPermissionActionFromMethod(r.Method)
		if resolveAction != nil {
			action = normalizePermissionAction(resolveAction(r), action)
		}

		rateLimit := empresaRateLimitMaxForRequest(dbSuper, r)
		rateScope := empresaRateLimitScopeForRequest(r)
		allowedByRate, remaining, retryAfter, current := checkEmpresaRateLimitAt(time.Now(), empresaID, rateScope, rateLimit)
		applyEmpresaRateLimitHeaders(w, rateLimit, remaining, retryAfter)
		if !allowedByRate {
			path := ""
			if r.URL != nil {
				path = strings.TrimSpace(r.URL.Path)
			}
			log.Printf("[rate_limit] empresa_id=%d scope=%s limite=%d actual=%d path=%s", empresaID, rateScope, rateLimit, current, path)
			http.Error(w, "limite de consumo por empresa excedido; intenta de nuevo en unos segundos", http.StatusTooManyRequests)
			registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusTooManyRequests, 0)
			return
		}

		adminEmail := strings.ToLower(strings.TrimSpace(adminEmailFromRequest(r)))
		if adminEmail == "" || adminEmail == "sistema" {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusUnauthorized, 0)
			return
		}

		snapshotStartedAt := time.Now()
		snapshot, err := getEmpresaPermissionSnapshot(dbEmp, dbSuper, adminEmail, empresaID)
		dbpkg.PerfLogf("[perf][authz] module=%s snapshot empresa=%d email=%s dur=%s", module, empresaID, adminEmail, time.Since(snapshotStartedAt))
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "unauthenticated", http.StatusUnauthorized)
				registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusUnauthorized, 0)
				return
			}
			log.Printf("[authz] snapshot module=%s email=%s empresa_id=%d error: %v", module, adminEmail, empresaID, err)
			http.Error(w, "No se pudo validar permisos del usuario", http.StatusInternalServerError)
			registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusInternalServerError, 0)
			return
		}

		if !snapshot.CanAccess {
			http.Error(w, "forbidden: empresa_id fuera del alcance del usuario autenticado", http.StatusForbidden)
			registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusForbidden, 0)
			return
		}

		role := snapshot.AdminRole
		skipLicenciaModuloCheck := module == permModuleSeguridad && strings.HasPrefix(strings.TrimSpace(r.URL.Path), "/api/empresa/permisos_contexto")
		if !skipLicenciaModuloCheck && !isModuloPermitidoByLicencia(module, snapshot.AllowedModules) {
			http.Error(w, "forbidden: modulo no habilitado por licencia activa", http.StatusForbidden)
			registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusForbidden, 0)
			return
		}
		if !skipLicenciaModuloCheck && len(snapshot.AllowedVerticalModules) > 0 && isEmpresaBusinessVerticalModule(module) && !snapshot.AllowedVerticalModules[normalizeVerticalScopeModule(module)] {
			http.Error(w, "forbidden: vertical no corresponde al tipo de empresa/licencia activa", http.StatusForbidden)
			registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusForbidden, 0)
			return
		}

		requestPath := strings.TrimSpace(r.URL.Path)
		effectiveRole := snapshot.EffectiveRole
		skipRoleModuloCheck := module == permModuleSeguridad && strings.HasPrefix(requestPath, "/api/empresa/permisos_contexto")
		if !skipRoleModuloCheck && !snapshot.RoleModuleActions[permissionModuleActionKey(module, action)] {
			http.Error(w, "forbidden: rol sin permiso para la accion solicitada", http.StatusForbidden)
			registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusForbidden, 0)
			return
		}
		pageKey := resolvePermissionPageKeyForRequest(r)
		// El flujo operativo de estaciones usa el mismo endpoint de carritos para
		// listar/recuperar sesiones por estacion. En ese caso el control real ya
		// queda cubierto por el modulo de ventas + el contexto estacion_id, asi
		// que no debemos bloquearlo por la pagina generica de carritos.
		if strings.EqualFold(strings.TrimSpace(requestPath), "/api/empresa/carritos_compra") &&
			strings.TrimSpace(r.URL.Query().Get("estacion_id")) != "" {
			pageKey = ""
		}
		if pageKey != "" {
			if !snapshot.AllowedPages[pageKey] {
				http.Error(w, "forbidden: rol sin acceso a la funcionalidad solicitada", http.StatusForbidden)
				registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusForbidden, 0)
				return
			}
		}

		if permissionChangeRequiresApproval(module, r, action) {
			evidence, err := extractPermissionApprovalEvidence(r)
			if err != nil {
				http.Error(w, "no se pudo validar evidencia de aprobacion para el cambio de permisos", http.StatusBadRequest)
				registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusBadRequest, 0)
				return
			}
			if evidence.ApprovedBy == "" || evidence.ApprovalCode == "" {
				http.Error(w, "se requiere aprobacion trazable (aprobado_por y codigo_aprobacion) para cambios de permisos", http.StatusBadRequest)
				registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, http.StatusBadRequest, 0)
				return
			}

			r.Header.Set(permissionApprovalHeaderRequired, "1")
			r.Header.Set(permissionApprovalHeaderBy, evidence.ApprovedBy)
			r.Header.Set(permissionApprovalHeaderCode, evidence.ApprovalCode)
			if evidence.Reason != "" {
				r.Header.Set(permissionApprovalHeaderReason, evidence.Reason)
			}
		}

		ctx := context.WithValue(r.Context(), "adminRole", role)
		ctx = context.WithValue(ctx, "adminRoleEfectivo", effectiveRole)
		ctx = context.WithValue(ctx, "empresaID", empresaID)
		r = r.WithContext(ctx)

		w.Header().Set("X-Empresa-ID", strconv.FormatInt(empresaID, 10))
		r.Header.Set("X-Admin-Role", role)
		r.Header.Set("X-Admin-Role-Efectivo", effectiveRole)

		auditStart := time.Now()
		auditRW := &auditCaptureResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(auditRW, r)
		dbpkg.PerfLogf("[perf][authz] module=%s next empresa=%d path=%s dur=%s", module, empresaID, requestPath, time.Since(auditStart))
		registrarAuditoriaOperacionNoBloqueante(dbEmp, r, empresaID, module, action, auditRW.status, time.Since(auditStart))
	}
}

func extractEmpresaIDForPermissions(r *http.Request) int64 {
	if id, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && id > 0 {
		return id
	}
	if id := parsePositiveInt64(strings.TrimSpace(r.Header.Get("X-Empresa-ID"))); id > 0 {
		return id
	}

	method := strings.ToUpper(strings.TrimSpace(r.Method))
	if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch && method != http.MethodDelete {
		return 0
	}

	contentType := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
	if strings.Contains(contentType, "application/json") {
		return extractEmpresaIDFromJSONBody(r)
	}
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		if err := r.ParseForm(); err == nil {
			if id := parsePositiveInt64(strings.TrimSpace(r.FormValue("empresa_id"))); id > 0 {
				return id
			}
		}
	}
	if strings.Contains(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(12 << 20); err == nil {
			if id := parsePositiveInt64(strings.TrimSpace(r.FormValue("empresa_id"))); id > 0 {
				return id
			}
		}
	}

	return 0
}

func extractEmpresaIDFromJSONBody(r *http.Request) int64 {
	if r.Body == nil {
		return 0
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		r.Body = io.NopCloser(bytes.NewReader(raw))
		return 0
	}
	r.Body = io.NopCloser(bytes.NewReader(raw))
	if len(raw) == 0 {
		return 0
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return 0
	}

	if v, ok := payload["empresa_id"]; ok {
		if id := toPositiveInt64(v); id > 0 {
			return id
		}
	}
	if v, ok := payload["empresaId"]; ok {
		if id := toPositiveInt64(v); id > 0 {
			return id
		}
	}
	if empresaObj, ok := payload["empresa"].(map[string]interface{}); ok {
		if v, exists := empresaObj["id"]; exists {
			if id := toPositiveInt64(v); id > 0 {
				return id
			}
		}
	}
	return 0
}

func validateEmpresaIDConsistency(r *http.Request, empresaID int64) error {
	if r == nil || empresaID <= 0 {
		return nil
	}
	if id, err := parseEmpresaIDQueryValue(r); err != nil {
		return fmt.Errorf("empresa_id invalido en query")
	} else if id > 0 && id != empresaID {
		return fmt.Errorf("empresa_id no coincide con el contexto de empresa")
	}
	if id := parsePositiveInt64(strings.TrimSpace(r.Header.Get("X-Empresa-ID"))); id > 0 && id != empresaID {
		return fmt.Errorf("empresa_id no coincide con el contexto de empresa")
	}

	method := strings.ToUpper(strings.TrimSpace(r.Method))
	if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch && method != http.MethodDelete {
		return nil
	}

	contentType := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
	if strings.Contains(contentType, "application/json") {
		if id, err := extractEmpresaIDFromJSONBodyStrict(r); err != nil {
			return err
		} else if id > 0 && id != empresaID {
			return fmt.Errorf("empresa_id no coincide con el contexto de empresa")
		}
		return nil
	}
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		if err := r.ParseForm(); err != nil {
			return fmt.Errorf("empresa_id invalido en formulario")
		}
		if id := parsePositiveInt64(strings.TrimSpace(r.FormValue("empresa_id"))); id > 0 && id != empresaID {
			return fmt.Errorf("empresa_id no coincide con el contexto de empresa")
		}
		return nil
	}
	if strings.Contains(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(12 << 20); err != nil {
			return fmt.Errorf("empresa_id invalido en multipart")
		}
		if id := parsePositiveInt64(strings.TrimSpace(r.FormValue("empresa_id"))); id > 0 && id != empresaID {
			return fmt.Errorf("empresa_id no coincide con el contexto de empresa")
		}
	}
	return nil
}

func parseEmpresaIDQueryValue(r *http.Request) (int64, error) {
	if r == nil || r.URL == nil {
		return 0, nil
	}
	raw := strings.TrimSpace(r.URL.Query().Get("empresa_id"))
	if raw == "" {
		return 0, nil
	}
	return strconv.ParseInt(raw, 10, 64)
}

func extractEmpresaIDFromJSONBodyStrict(r *http.Request) (int64, error) {
	if r.Body == nil {
		return 0, nil
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		r.Body = io.NopCloser(bytes.NewReader(raw))
		return 0, fmt.Errorf("empresa_id invalido en JSON")
	}
	r.Body = io.NopCloser(bytes.NewReader(raw))
	if len(raw) == 0 {
		return 0, nil
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return 0, fmt.Errorf("empresa_id invalido en JSON")
	}
	if v, ok := payload["empresa_id"]; ok {
		return toPositiveInt64(v), nil
	}
	if v, ok := payload["empresaId"]; ok {
		return toPositiveInt64(v), nil
	}
	if empresaObj, ok := payload["empresa"].(map[string]interface{}); ok {
		if v, exists := empresaObj["id"]; exists {
			return toPositiveInt64(v), nil
		}
	}
	return 0, nil
}

func toPositiveInt64(v interface{}) int64 {
	switch n := v.(type) {
	case float64:
		if n > 0 {
			return int64(n)
		}
	case int64:
		if n > 0 {
			return n
		}
	case int:
		if n > 0 {
			return int64(n)
		}
	case string:
		return parsePositiveInt64(n)
	}
	return 0
}

func parsePositiveInt64(raw string) int64 {
	v := strings.TrimSpace(raw)
	if v == "" {
		return 0
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil || n <= 0 {
		return 0
	}
	return n
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func trimWithLimit(raw string, maxLen int) string {
	v := strings.TrimSpace(raw)
	if maxLen > 0 && len(v) > maxLen {
		return v[:maxLen]
	}
	return v
}

func extractJSONBodyMap(r *http.Request) (map[string]interface{}, error) {
	if r == nil || r.Body == nil {
		return nil, nil
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		r.Body = io.NopCloser(bytes.NewReader(raw))
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewReader(raw))
	if len(raw) == 0 {
		return nil, nil
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, nil
	}
	return payload, nil
}

func extractStringField(payload map[string]interface{}, keys ...string) string {
	if payload == nil {
		return ""
	}
	for _, key := range keys {
		if value, ok := payload[key]; ok {
			switch typed := value.(type) {
			case string:
				if trimmed := strings.TrimSpace(typed); trimmed != "" {
					return trimmed
				}
			case float64:
				if typed > 0 {
					return strings.TrimSpace(strconv.FormatFloat(typed, 'f', -1, 64))
				}
			case int64:
				if typed > 0 {
					return strings.TrimSpace(strconv.FormatInt(typed, 10))
				}
			case int:
				if typed > 0 {
					return strings.TrimSpace(strconv.Itoa(typed))
				}
			}
		}
	}
	return ""
}

func extractPermissionApprovalEvidence(r *http.Request) (permissionApprovalEvidence, error) {
	evidence := permissionApprovalEvidence{
		ApprovedBy: trimWithLimit(firstNonEmpty(
			r.URL.Query().Get("aprobado_por"),
			r.URL.Query().Get("approved_by"),
			r.Header.Get(permissionApprovalHeaderBy),
		), 160),
		ApprovalCode: trimWithLimit(firstNonEmpty(
			r.URL.Query().Get("codigo_aprobacion"),
			r.URL.Query().Get("approval_code"),
			r.Header.Get(permissionApprovalHeaderCode),
		), 160),
		Reason: trimWithLimit(firstNonEmpty(
			r.URL.Query().Get("motivo_aprobacion"),
			r.URL.Query().Get("approval_reason"),
			r.Header.Get(permissionApprovalHeaderReason),
		), 320),
	}

	payload, err := extractJSONBodyMap(r)
	if err != nil {
		return permissionApprovalEvidence{}, err
	}
	if payload == nil {
		return evidence, nil
	}

	evidence.ApprovedBy = trimWithLimit(firstNonEmpty(
		evidence.ApprovedBy,
		extractStringField(payload, "aprobado_por", "approved_by"),
	), 160)
	evidence.ApprovalCode = trimWithLimit(firstNonEmpty(
		evidence.ApprovalCode,
		extractStringField(payload, "codigo_aprobacion", "approval_code"),
	), 160)
	evidence.Reason = trimWithLimit(firstNonEmpty(
		evidence.Reason,
		extractStringField(payload, "motivo_aprobacion", "approval_reason"),
	), 320)

	aprobacionPayload, _ := payload["aprobacion"].(map[string]interface{})
	evidence.ApprovedBy = trimWithLimit(firstNonEmpty(
		evidence.ApprovedBy,
		extractStringField(aprobacionPayload, "aprobado_por", "approved_by"),
	), 160)
	evidence.ApprovalCode = trimWithLimit(firstNonEmpty(
		evidence.ApprovalCode,
		extractStringField(aprobacionPayload, "codigo_aprobacion", "approval_code"),
	), 160)
	evidence.Reason = trimWithLimit(firstNonEmpty(
		evidence.Reason,
		extractStringField(aprobacionPayload, "motivo_aprobacion", "approval_reason"),
	), 320)

	return evidence, nil
}

func permissionChangeRequiresApproval(module string, r *http.Request, action string) bool {
	if module != permModuleSeguridad {
		return false
	}

	switch strings.ToUpper(strings.TrimSpace(action)) {
	case permActionCreate, permActionUpdate, permActionDelete, permActionApprove:
	default:
		return false
	}

	path := strings.ToLower(strings.TrimSpace(r.URL.Path))
	if path == "/api/empresa/roles_de_usuario" {
		return !strings.EqualFold(strings.TrimSpace(r.Method), http.MethodGet)
	}
	if path == "/api/empresa/permisos_empresa" {
		return !strings.EqualFold(strings.TrimSpace(r.Method), http.MethodGet)
	}
	if path != "/api/empresa/usuarios" {
		return false
	}

	queryAction := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	if queryAction == "reenviar_confirmacion" || queryAction == "activar" {
		return false
	}

	return true
}

func defaultPermissionActionFromMethod(method string) string {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return permActionRead
	case http.MethodPost:
		return permActionCreate
	case http.MethodPut, http.MethodPatch:
		return permActionUpdate
	case http.MethodDelete:
		return permActionDelete
	default:
		return permActionRead
	}
}

func normalizePermissionAction(candidate, fallback string) string {
	v := strings.ToUpper(strings.TrimSpace(candidate))
	if v == "" {
		return fallback
	}
	switch v {
	case permActionRead, permActionCreate, permActionUpdate, permActionDelete, permActionApprove:
		return v
	default:
		return fallback
	}
}

func resolveVentasPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	switch action {
	case "pagar_estacion", "pagar":
		return permActionCreate
	case "cerrar", "reabrir", "activar_estacion", "suspender", "suspender_venta", "reactivar", "reabrir_venta", "convertir_pedido", "convertir_documento_final":
		return permActionApprove
	case "activar", "desactivar":
		return permActionUpdate
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolveInventarioPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	if action == "activar" || action == "desactivar" {
		return permActionUpdate
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolveFinanzasPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	switch action {
	case "cerrar", "reabrir", "aprobar", "procesar_asientos", "procesar", "conciliar_bancaria_auto", "conciliar_bancos", "conciliar_bancaria_automatica", "aprobar_workflow", "aprobar_reverso", "aprobar_refinanciacion", "rechazar_workflow", "rechazar_reverso", "rechazar_refinanciacion":
		return permActionApprove
	case "anular":
		return permActionDelete
	case "activar", "desactivar":
		return permActionUpdate
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolveCobranzaPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	switch action {
	case "dashboard", "cuentas", "plantillas", "campanas", "gestiones", "promesas":
		return permActionRead
	case "marcar_promesa":
		return permActionApprove
	case "plantilla", "campana", "gestion", "promesa", "simular_envio", "seed_demo":
		return permActionCreate
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolvePortalContadorPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	switch action {
	case "dashboard", "clientes", "obligaciones", "solicitudes", "comunicaciones":
		return permActionRead
	case "obligacion", "solicitud":
		if r.Method == http.MethodPut {
			return permActionUpdate
		}
		return permActionCreate
	case "cliente", "comunicacion", "seed_demo":
		return permActionCreate
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolveCentrosCostoPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	switch action {
	case "", "dashboard", "centros", "reglas", "presupuestos", "movimientos":
		return permActionRead
	case "seed_demo":
		return permActionCreate
	case "centro", "regla", "presupuesto":
		if strings.EqualFold(strings.TrimSpace(r.Method), http.MethodPut) {
			return permActionUpdate
		}
		return permActionCreate
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolveCierreFiscalPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	switch action {
	case "", "dashboard", "politicas", "periodos", "excepciones", "eventos", "validar":
		return permActionRead
	case "estado_periodo", "excepcion":
		return permActionApprove
	case "seed_demo":
		return permActionCreate
	case "politica", "periodo":
		if strings.EqualFold(strings.TrimSpace(r.Method), http.MethodPut) {
			return permActionUpdate
		}
		return permActionCreate
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolvePortalTercerosPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	switch action {
	case "dashboard", "terceros", "certificados", "descargas":
		return permActionRead
	case "certificado":
		if r.Method == http.MethodPut {
			return permActionUpdate
		}
		return permActionApprove
	case "tercero", "seed_demo":
		return permActionCreate
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolveSoportesComprasIAPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	switch action {
	case "dashboard", "soportes", "eventos":
		return permActionRead
	case "aprobar", "rechazar", "contabilizar":
		return permActionApprove
	case "extraer_ia":
		return permActionUpdate
	case "radicar", "seed_demo":
		return permActionCreate
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolveContabilidadColombiaPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	switch action {
	case "dashboard", "comprobante", "periodos", "eventos":
		return permActionRead
	case "cerrar_periodo", "reabrir_periodo", "seed":
		return permActionApprove
	case "anular_comprobante":
		return permActionDelete
	case "config", "cuentas", "terceros", "impuestos", "comprobantes":
		return defaultPermissionActionFromMethod(r.Method)
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolveContabilidadColombiaAvanzadaPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	switch action {
	case "seed", "generar_exogena", "libros", "libros_resumen":
		return permActionApprove
	case "exogena_formatos", "exogena_registros", "nomina_electronica", "documentos_soporte", "activos_fijos", "cartera_cxp":
		return defaultPermissionActionFromMethod(r.Method)
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolveActivosFijosNIIFPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	switch action {
	case "dashboard", "activos", "libro", "depreciaciones", "eventos":
		return permActionRead
	case "depreciacion", "seed_demo":
		return permActionApprove
	case "evento":
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			return permActionUpdate
		}
		return permActionRead
	case "activo":
		return defaultPermissionActionFromMethod(r.Method)
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolveDeclaracionesTributariasPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	switch action {
	case "dashboard", "declaraciones", "movimientos":
		return permActionRead
	case "calendario":
		if strings.EqualFold(r.Method, http.MethodGet) {
			return permActionRead
		}
		return defaultPermissionActionFromMethod(r.Method)
	case "preliquidar", "seed_demo":
		return permActionApprove
	case "declaracion":
		if r.Method == http.MethodPut {
			return permActionUpdate
		}
		return permActionCreate
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolveClientesPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	if action == "activar" || action == "desactivar" {
		return permActionUpdate
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolveComprasPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	if action == "activar" || action == "desactivar" {
		return permActionUpdate
	}
	if action == "anular" || action == "cancelar" {
		return permActionDelete
	}
	if action == "aprobar" || action == "cerrar" || action == "emitir" || action == "emitir_orden" || action == "recepcionar" || action == "recepcionar_compra" || action == "recepcionar_parcial_compra" || action == "contabilizar" || action == "contabilizar_compra" || action == "solicitar_aprobacion" || action == "aprobar_compra" || action == "rechazar_compra" || action == "validar_documentos" {
		return permActionApprove
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolveFacturacionPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	if action == "activar" || action == "desactivar" {
		return permActionUpdate
	}
	if (action == "procesar_reintentos" || action == "reconciliar_estados" || action == "firmar_xml_real" || action == "enviar_documento_real" || action == "reconexion_dian" || action == "consultar_acuse_real") && (r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch) {
		return permActionApprove
	}
	if action == "aprobar" || action == "emitir" || action == "emitir_factura" || action == "emitir_documento" || action == "nota_credito" || action == "emitir_nota_credito" {
		return permActionApprove
	}
	if action == "anular" {
		return permActionDelete
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolveAIUConstruccionPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	switch action {
	case "dashboard", "contratos", "detalle", "facturas", "eventos", "reporte":
		return permActionRead
	case "generar_factura", "aprobar", "cerrar", "facturar", "estado":
		return permActionApprove
	case "anular", "eliminar":
		return permActionDelete
	default:
		return defaultPermissionActionFromMethod(r.Method)
	}
}

func resolveSeguridadPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	switch action {
	case "activar", "desactivar":
		return permActionUpdate
	case "solicitar_aprobacion", "iniciar_aprobacion":
		return permActionUpdate
	case "versionar":
		return permActionApprove
	case "restaurar", "restore", "rollback_backup":
		return permActionApprove
	case "depurar_fecha", "purgar_fecha", "eliminar_hasta_fecha", "depurar_hasta_fecha":
		return permActionApprove
	case "sync_manual", "rotar_credencial", "rotar_credenciales":
		return permActionApprove
	case "aprobar", "rechazar", "vincular_nomina", "enlazar_nomina":
		return permActionApprove
	case "reenviar_confirmacion":
		return permActionApprove
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolveVerticalPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	switch action {
	case "activar", "desactivar":
		return permActionUpdate
	case "anular", "cancelar", "eliminar":
		return permActionDelete
	case "aprobar", "cerrar", "despachar", "dispatch":
		return permActionApprove
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func resolveControlElectricoPermissionAction(r *http.Request) string {
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	switch action {
	case "config", "raspberry_pi", "rele", "rele_foto":
		return defaultPermissionActionFromMethod(r.Method)
	case "probar_rele", "sincronizar", "ejecutar_programacion":
		return permActionApprove
	case "activar", "desactivar":
		return permActionUpdate
	}
	return defaultPermissionActionFromMethod(r.Method)
}

func normalizePermissionRole(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "super_administrador", "superadmin", "super":
		return "super_administrador"
	case "administrador_total", "admin_total", "admin_full", "full_admin":
		return "administrador_total"
	case "administrador", "admin", "admin_empresa":
		return "admin_empresa"
	case "supervisor", "supervisor_sucursal":
		return "supervisor_sucursal"
	case "cajero":
		return "cajero"
	case "inventario":
		return "inventario"
	case "compras":
		return "compras"
	case "contabilidad", "contador":
		return "contabilidad"
	case "auditor":
		return "auditor"
	default:
		return strings.ToLower(strings.TrimSpace(raw))
	}
}

func roleAllowsModuleAction(role, module, action string) bool {
	if role == "super_administrador" {
		return true
	}
	if role == "administrador_total" {
		return true
	}

	allReadRoles := []string{"admin_empresa", "supervisor_sucursal", "cajero", "inventario", "compras", "contabilidad", "auditor"}
	if isPermModuleNuevoVertical(module) {
		switch action {
		case permActionRead:
			return roleIn(role, allReadRoles...)
		case permActionCreate, permActionUpdate, permActionApprove:
			return roleIn(role, "admin_empresa", "supervisor_sucursal", "cajero")
		case permActionDelete:
			return roleIn(role, "admin_empresa", "supervisor_sucursal")
		}
	}

	switch module {
	case permModuleVentas:
		switch action {
		case permActionRead:
			return roleIn(role, allReadRoles...)
		case permActionCreate, permActionUpdate, permActionDelete, permActionApprove:
			return roleIn(role, "admin_empresa", "supervisor_sucursal", "cajero")
		}

	case permModuleVentaPublica, permModuleReservasHotel, permModuleChatTareas, permModuleGimnasio, permModuleTaxiSystem, permModuleDomicilios, permModuleParqueadero, permModuleApartTuristicos, permModulePropiedadHorizontal, permModuleAlquileres, permModuleOdontologia, permModuleDrogueriaFarmacia, permModuleTurnos, permModuleCarnets:
		switch action {
		case permActionRead:
			return roleIn(role, allReadRoles...)
		case permActionCreate, permActionUpdate, permActionApprove:
			return roleIn(role, "admin_empresa", "supervisor_sucursal", "cajero")
		case permActionDelete:
			return roleIn(role, "admin_empresa", "supervisor_sucursal")
		}

	case permModuleInventario:
		switch action {
		case permActionRead:
			return roleIn(role, allReadRoles...)
		case permActionCreate, permActionUpdate, permActionDelete, permActionApprove:
			return roleIn(role, "admin_empresa", "supervisor_sucursal", "inventario")
		}

	case permModuleProduccionMRP, permModuleLogisticaWMS, permModuleImportaciones:
		switch action {
		case permActionRead:
			return roleIn(role, allReadRoles...)
		case permActionCreate, permActionUpdate, permActionDelete, permActionApprove:
			return roleIn(role, "admin_empresa", "supervisor_sucursal", "inventario", "compras")
		}

	case permModuleFinanzas, permModuleContabilidadCO, permModuleContabilidadCOAv, permModuleCentrosCosto, permModuleCierreFiscal, permModuleActivosFijosNIIF, permModuleDeclaracionesTrib, permModuleTesoreria, permModuleNominaSueldos, permModuleCobranza, permModulePortalContador, permModulePortalTerceros:
		switch action {
		case permActionRead:
			return roleIn(role, allReadRoles...)
		case permActionCreate, permActionUpdate, permActionApprove:
			return roleIn(role, "admin_empresa", "contabilidad")
		case permActionDelete:
			return roleIn(role, "contabilidad")
		}

	case permModuleBancosPagos, permModuleGestionDocumental, permModuleCumplimientoKYC, permModuleContratosOblig, permModuleCalidadProcesos, permModuleAuditoria, permModuleBackups, permModuleDocumentosOnlyOffice:
		switch action {
		case permActionRead:
			return roleIn(role, allReadRoles...)
		case permActionCreate, permActionUpdate, permActionApprove:
			return roleIn(role, "admin_empresa", "supervisor_sucursal", "contabilidad", "auditor")
		case permActionDelete:
			return roleIn(role, "admin_empresa")
		}

	case permModuleClientes, permModuleCRMUnificado:
		switch action {
		case permActionRead:
			return roleIn(role, allReadRoles...)
		case permActionCreate, permActionUpdate, permActionApprove:
			return roleIn(role, "admin_empresa", "supervisor_sucursal", "cajero")
		case permActionDelete:
			return false
		}

	case permModuleCompras:
		switch action {
		case permActionRead:
			return roleIn(role, allReadRoles...)
		case permActionCreate, permActionUpdate, permActionApprove:
			return roleIn(role, "admin_empresa", "supervisor_sucursal", "compras")
		case permActionDelete:
			return false
		}

	case permModuleSoportesComprasIA:
		switch action {
		case permActionRead:
			return roleIn(role, allReadRoles...)
		case permActionCreate, permActionUpdate, permActionApprove:
			return roleIn(role, "admin_empresa", "supervisor_sucursal", "compras", "contabilidad")
		case permActionDelete:
			return false
		}

	case permModuleFacturacion:
		switch action {
		case permActionRead:
			return roleIn(role, allReadRoles...)
		case permActionCreate, permActionUpdate, permActionApprove:
			return roleIn(role, "admin_empresa", "cajero")
		case permActionDelete:
			return false
		}

	case permModuleAIUConstruccion:
		switch action {
		case permActionRead:
			return roleIn(role, allReadRoles...)
		case permActionCreate, permActionUpdate, permActionApprove:
			return roleIn(role, "admin_empresa", "contabilidad", "supervisor_sucursal")
		case permActionDelete:
			return roleIn(role, "admin_empresa", "contabilidad")
		}

	case permModuleSeguridad:
		switch action {
		case permActionRead:
			return roleIn(role, allReadRoles...)
		case permActionCreate, permActionUpdate, permActionDelete, permActionApprove:
			return roleIn(role, "admin_empresa")
		}

	case permModuleControlElectrico:
		switch action {
		case permActionRead:
			return roleIn(role, allReadRoles...)
		case permActionCreate, permActionUpdate:
			return roleIn(role, "admin_empresa", "supervisor_sucursal")
		case permActionDelete, permActionApprove:
			return roleIn(role, "admin_empresa")
		}

	case permModuleHorariosTrab, permModuleAsistenciaEmpleados, permModuleVehiculosRegistro, permModuleHojaVidaOperativa, permModuleUbicacionGPS:
		switch action {
		case permActionRead:
			return roleIn(role, allReadRoles...)
		case permActionCreate, permActionUpdate:
			return roleIn(role, "admin_empresa", "supervisor_sucursal")
		case permActionDelete, permActionApprove:
			return roleIn(role, "admin_empresa")
		}

	case permModuleReportes:
		switch action {
		case permActionRead:
			return roleIn(role, allReadRoles...)
		case permActionCreate, permActionUpdate, permActionApprove:
			return roleIn(role, "admin_empresa", "supervisor_sucursal", "contabilidad", "auditor")
		case permActionDelete:
			return roleIn(role, "admin_empresa")
		}
	}

	return false
}

func roleAllowsModuleActionWithOverrides(dbSuper *sql.DB, role, module, action string) bool {
	normalizedRole := normalizePermissionRole(role)
	normalizedModule := strings.ToLower(strings.TrimSpace(module))
	normalizedAction := strings.ToUpper(strings.TrimSpace(action))

	allowed := roleAllowsModuleAction(normalizedRole, normalizedModule, normalizedAction)
	if dbSuper == nil || normalizedRole == "" || normalizedRole == "sin_rol" {
		return allowed
	}

	found, permitido, err := dbpkg.LookupRolPermisoModuloByRoleName(dbSuper, normalizedRole, normalizedModule, normalizedAction)
	if err != nil {
		if isPermissionMissingTableError(err) {
			return allowed
		}
		log.Printf("[authz] modulo override lookup role=%s modulo=%s accion=%s error: %v", normalizedRole, normalizedModule, normalizedAction, err)
		return allowed
	}
	if found {
		return permitido
	}

	return allowed
}

func empresaAllowsModuleActionWithOverrides(dbSuper *sql.DB, empresaID int64, module, action string) bool {
	if dbSuper == nil || empresaID <= 0 {
		return true
	}
	normalizedModule := strings.ToLower(strings.TrimSpace(module))
	normalizedAction := strings.ToUpper(strings.TrimSpace(action))
	if normalizedModule == "" || normalizedAction == "" {
		return true
	}
	found, permitido, err := dbpkg.LookupEmpresaPermisoModulo(dbSuper, empresaID, normalizedModule, normalizedAction)
	if err != nil {
		if isPermissionMissingTableError(err) {
			return true
		}
		log.Printf("[authz] empresa permiso lookup empresa_id=%d modulo=%s accion=%s error: %v", empresaID, normalizedModule, normalizedAction, err)
		return true
	}
	if found {
		return permitido
	}
	return true
}

func loadPermissionOverridesByRoleName(dbSuper *sql.DB, role string) (map[string]bool, map[string]bool, error) {
	moduleOverrides := map[string]bool{}
	pageOverrides := map[string]bool{}
	if dbSuper == nil {
		return moduleOverrides, pageOverrides, nil
	}

	normalizedRole := normalizePermissionRole(role)
	if normalizedRole == "" || normalizedRole == "sin_rol" {
		return moduleOverrides, pageOverrides, nil
	}

	empresaPermissionCacheMu.Lock()
	if cached, ok := rolePermissionOverrideCache[normalizedRole]; ok && time.Since(cached.LoadedAt) < permissionOverrideCacheTTL {
		empresaPermissionCacheMu.Unlock()
		return clonePermissionBoolMap(cached.ModuleOverrides), clonePermissionBoolMap(cached.PageOverrides), nil
	}
	empresaPermissionCacheMu.Unlock()

	rolID, err := dbpkg.ResolveRolDeUsuarioIDByNombre(dbSuper, normalizedRole)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || isPermissionMissingTableError(err) {
			return moduleOverrides, pageOverrides, nil
		}
		return nil, nil, err
	}

	var (
		modulos    []dbpkg.RolPermisoModulo
		paginas    []dbpkg.RolPermisoPagina
		modulosErr error
		paginasErr error
	)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		modulos, modulosErr = dbpkg.ListRolPermisosModuloByRolID(dbSuper, rolID)
	}()
	go func() {
		defer wg.Done()
		paginas, paginasErr = dbpkg.ListRolPermisosPaginaByRolID(dbSuper, rolID)
	}()
	wg.Wait()
	if modulosErr != nil {
		if isPermissionMissingTableError(modulosErr) {
			return moduleOverrides, pageOverrides, nil
		}
		return nil, nil, modulosErr
	}
	for _, item := range modulos {
		moduleOverrides[permissionModuleActionKey(item.Modulo, item.Accion)] = item.Permitido
	}

	if paginasErr != nil {
		if isPermissionMissingTableError(paginasErr) {
			return moduleOverrides, pageOverrides, nil
		}
		return nil, nil, paginasErr
	}
	for _, item := range paginas {
		pageOverrides[strings.TrimSpace(item.PaginaClave)] = item.Permitido
	}

	empresaPermissionCacheMu.Lock()
	rolePermissionOverrideCache[normalizedRole] = permissionRoleOverrideCacheEntry{
		ModuleOverrides: clonePermissionBoolMap(moduleOverrides),
		PageOverrides:   clonePermissionBoolMap(pageOverrides),
		LoadedAt:        time.Now(),
	}
	empresaPermissionCacheMu.Unlock()

	return moduleOverrides, pageOverrides, nil
}

func loadEmpresaPermissionOverrides(dbSuper *sql.DB, empresaID int64) (map[string]bool, map[string]bool, *empresaPermisosFinosCtx) {
	moduleOverrides := map[string]bool{}
	pageOverrides := map[string]bool{}
	ctx := &empresaPermisosFinosCtx{}
	if dbSuper == nil || empresaID <= 0 {
		return moduleOverrides, pageOverrides, ctx
	}

	empresaPermissionCacheMu.Lock()
	if cached, ok := empresaPermissionOverrideCache[empresaID]; ok && time.Since(cached.LoadedAt) < permissionOverrideCacheTTL {
		empresaPermissionCacheMu.Unlock()
		return clonePermissionBoolMap(cached.ModuleOverrides), clonePermissionBoolMap(cached.PageOverrides), cloneEmpresaPermisosFinosCtx(cached.Ctx)
	}
	empresaPermissionCacheMu.Unlock()

	var (
		modulos    []dbpkg.EmpresaPermisoModulo
		paginas    []dbpkg.EmpresaPermisoPagina
		modulosErr error
		paginasErr error
	)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		modulos, modulosErr = dbpkg.ListEmpresaPermisosModuloByEmpresaID(dbSuper, empresaID)
	}()
	go func() {
		defer wg.Done()
		paginas, paginasErr = dbpkg.ListEmpresaPermisosPaginaByEmpresaID(dbSuper, empresaID)
	}()
	wg.Wait()
	if modulosErr != nil {
		if !isPermissionMissingTableError(modulosErr) {
			log.Printf("[authz] load empresa modulo overrides empresa_id=%d error: %v", empresaID, modulosErr)
		}
		modulos = []dbpkg.EmpresaPermisoModulo{}
	}
	for _, item := range modulos {
		moduleOverrides[permissionModuleActionKey(item.Modulo, item.Accion)] = item.Permitido
	}

	if paginasErr != nil {
		if !isPermissionMissingTableError(paginasErr) {
			log.Printf("[authz] load empresa page overrides empresa_id=%d error: %v", empresaID, paginasErr)
		}
		paginas = []dbpkg.EmpresaPermisoPagina{}
	}
	for _, item := range paginas {
		pageOverrides[strings.TrimSpace(item.PaginaClave)] = item.Permitido
	}

	ctx.ReglasModulo = len(moduleOverrides)
	ctx.ReglasPagina = len(pageOverrides)
	ctx.Activo = ctx.ReglasModulo > 0 || ctx.ReglasPagina > 0

	empresaPermissionCacheMu.Lock()
	empresaPermissionOverrideCache[empresaID] = empresaPermissionOverrideCacheEntry{
		ModuleOverrides: clonePermissionBoolMap(moduleOverrides),
		PageOverrides:   clonePermissionBoolMap(pageOverrides),
		Ctx:             cloneEmpresaPermisosFinosCtx(ctx),
		LoadedAt:        time.Now(),
	}
	empresaPermissionCacheMu.Unlock()
	return moduleOverrides, pageOverrides, ctx
}

func clonePermissionBoolMap(input map[string]bool) map[string]bool {
	if len(input) == 0 {
		return map[string]bool{}
	}
	out := make(map[string]bool, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func cloneEmpresaPermisosFinosCtx(input *empresaPermisosFinosCtx) *empresaPermisosFinosCtx {
	if input == nil {
		return &empresaPermisosFinosCtx{}
	}
	out := *input
	return &out
}

func clonePermissionModuleRows(input []permissionModuleMatrixRow) []permissionModuleMatrixRow {
	if len(input) == 0 {
		return []permissionModuleMatrixRow{}
	}
	out := make([]permissionModuleMatrixRow, 0, len(input))
	for _, row := range input {
		copied := row
		copied.Acciones = clonePermissionBoolMap(row.Acciones)
		out = append(out, copied)
	}
	return out
}

func permissionModuleActionKey(modulo, accion string) string {
	return strings.ToLower(strings.TrimSpace(modulo)) + "|" + strings.ToUpper(strings.TrimSpace(accion))
}

func setPermissionActionOnModuleRow(row *permissionModuleMatrixRow, action string, permitido bool) {
	normalizedAction := strings.ToUpper(strings.TrimSpace(action))
	switch normalizedAction {
	case permActionRead:
		row.Read = permitido
	case permActionCreate:
		row.Create = permitido
	case permActionUpdate:
		row.Update = permitido
	case permActionDelete:
		row.Delete = permitido
	case permActionApprove:
		row.Approve = permitido
	}
	if row.Acciones == nil {
		row.Acciones = map[string]bool{}
	}
	row.Acciones[normalizedAction] = permitido
}

func isPermissionMissingTableError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(msg, "no such table") || strings.Contains(msg, "does not exist")
}

func roleIn(role string, allowed ...string) bool {
	role = strings.TrimSpace(strings.ToLower(role))
	if role == "" {
		return false
	}
	for _, it := range allowed {
		if role == strings.TrimSpace(strings.ToLower(it)) {
			return true
		}
	}
	return false
}

func buildPermissionModuleMatrixForRole(role string) []permissionModuleMatrixRow {
	normalizedRole := normalizePermissionRole(role)
	out := make([]permissionModuleMatrixRow, 0, len(permissionModulesCatalogOrdered))
	for _, modulo := range permissionModulesCatalogOrdered {
		readAllowed := roleAllowsModuleAction(normalizedRole, modulo, permActionRead)
		createAllowed := roleAllowsModuleAction(normalizedRole, modulo, permActionCreate)
		updateAllowed := roleAllowsModuleAction(normalizedRole, modulo, permActionUpdate)
		deleteAllowed := roleAllowsModuleAction(normalizedRole, modulo, permActionDelete)
		approveAllowed := roleAllowsModuleAction(normalizedRole, modulo, permActionApprove)

		out = append(out, permissionModuleMatrixRow{
			Modulo:  modulo,
			Read:    readAllowed,
			Create:  createAllowed,
			Update:  updateAllowed,
			Delete:  deleteAllowed,
			Approve: approveAllowed,
			Acciones: map[string]bool{
				permActionRead:    readAllowed,
				permActionCreate:  createAllowed,
				permActionUpdate:  updateAllowed,
				permActionDelete:  deleteAllowed,
				permActionApprove: approveAllowed,
			},
		})
	}
	return out
}

func buildPermissionModuleMatrixForRoleDynamic(dbSuper *sql.DB, role string) []permissionModuleMatrixRow {
	normalizedRole := normalizePermissionRole(role)
	empresaPermissionCacheMu.Lock()
	if cached, ok := rolePermissionModuleMatrixCache[normalizedRole]; ok && time.Since(cached.LoadedAt) < permissionOverrideCacheTTL {
		empresaPermissionCacheMu.Unlock()
		return clonePermissionModuleRows(cached.Rows)
	}
	empresaPermissionCacheMu.Unlock()

	rows := buildPermissionModuleMatrixForRole(role)
	moduleOverrides, _, err := loadPermissionOverridesByRoleName(dbSuper, role)
	if err != nil {
		log.Printf("[authz] load permission overrides role=%s error: %v", role, err)
		return rows
	}
	if len(moduleOverrides) == 0 {
		return rows
	}

	for idx := range rows {
		row := &rows[idx]
		for _, action := range permissionActionsCatalogOrdered {
			if permitido, ok := moduleOverrides[permissionModuleActionKey(row.Modulo, action)]; ok {
				setPermissionActionOnModuleRow(row, action, permitido)
			}
		}
	}

	empresaPermissionCacheMu.Lock()
	rolePermissionModuleMatrixCache[normalizedRole] = permissionRoleModuleMatrixCacheEntry{
		Rows:     clonePermissionModuleRows(rows),
		LoadedAt: time.Now(),
	}
	empresaPermissionCacheMu.Unlock()
	return rows
}

func buildEmpresaPermisosDefaultModuleRows() []permissionModuleMatrixRow {
	out := make([]permissionModuleMatrixRow, 0, len(permissionModulesCatalogOrdered))
	for _, modulo := range permissionModulesCatalogOrdered {
		out = append(out, permissionModuleMatrixRow{
			Modulo:  modulo,
			Read:    true,
			Create:  true,
			Update:  true,
			Delete:  true,
			Approve: true,
			Acciones: map[string]bool{
				permActionRead:    true,
				permActionCreate:  true,
				permActionUpdate:  true,
				permActionDelete:  true,
				permActionApprove: true,
			},
		})
	}
	return out
}

func buildPermissionPagesMapForRoleDynamic(dbSuper *sql.DB, role string, modulos []permissionModuleMatrixRow) map[string]bool {
	_, pageOverrides, err := loadPermissionOverridesByRoleName(dbSuper, role)
	if err != nil {
		log.Printf("[authz] load page overrides role=%s error: %v", role, err)
		pageOverrides = map[string]bool{}
	}
	return buildPermissionPagesMapFromModuleRows(modulos, pageOverrides)
}

func buildPermissionPagesCatalogForRoleDynamic(dbSuper *sql.DB, role string, modulos []permissionModuleMatrixRow) []permissionPageAccessRow {
	_, pageOverrides, err := loadPermissionOverridesByRoleName(dbSuper, role)
	if err != nil {
		log.Printf("[authz] load page catalog overrides role=%s error: %v", role, err)
		pageOverrides = map[string]bool{}
	}
	return buildPermissionPagesCatalogFromModuleRows(modulos, pageOverrides)
}

func buildPermissionPagesMapFromModuleRows(modulos []permissionModuleMatrixRow, pageOverrides map[string]bool) map[string]bool {
	rows := buildPermissionPagesCatalogFromModuleRows(modulos, pageOverrides)
	out := make(map[string]bool, len(rows))
	for _, row := range rows {
		out[row.PaginaClave] = row.Permitido
	}
	return out
}

func buildPermissionPagesCatalogFromModuleRows(modulos []permissionModuleMatrixRow, pageOverrides map[string]bool) []permissionPageAccessRow {
	moduleRows := make(map[string]permissionModuleMatrixRow, len(modulos))
	for _, row := range modulos {
		moduleRows[strings.ToLower(strings.TrimSpace(row.Modulo))] = row
	}

	out := make([]permissionPageAccessRow, 0, len(permissionPagesCatalogOrdered))
	for _, rule := range permissionPagesCatalogOrdered {
		permitido := true
		if !rule.AlwaysVisible {
			permitido = permissionPageRulePermittedByModuleRows(rule, moduleRows)
		}
		if override, ok := pageOverrides[rule.PaginaClave]; ok {
			permitido = override
		}

		titulo := sanitizeLegacyPermissionVisibleText(rule.Titulo)
		if titulo == "" {
			titulo = sanitizeLegacyPermissionVisibleText(rule.PaginaClave)
		}
		grupo := universalPermissionGroupLabel(rule.Grupo)
		if grupo == "" {
			grupo = "Otras"
		}
		out = append(out, permissionPageAccessRow{
			PaginaClave:   rule.PaginaClave,
			Modulo:        sanitizeLegacyPermissionVisibleText(rule.Modulo),
			Accion:        sanitizeLegacyPermissionVisibleText(rule.Accion),
			AnyModules:    append([]string{}, rule.AnyModules...),
			Permitido:     permitido,
			AlwaysVisible: rule.AlwaysVisible,
			Titulo:        titulo,
			Grupo:         grupo,
		})
	}

	return out
}

func permissionPageRulePermittedByModuleRows(rule permissionPageRule, moduleRows map[string]permissionModuleMatrixRow) bool {
	action := strings.ToUpper(strings.TrimSpace(rule.Accion))
	if action == "" {
		action = permActionRead
	}
	if len(rule.AnyModules) > 0 {
		for _, module := range rule.AnyModules {
			if moduleRow, ok := moduleRows[strings.ToLower(strings.TrimSpace(module))]; ok && moduleRow.Acciones[action] {
				return true
			}
		}
		return false
	}
	if moduleRow, ok := moduleRows[strings.ToLower(strings.TrimSpace(rule.Modulo))]; ok {
		return moduleRow.Acciones[action]
	}
	return false
}

func parseLicenciaModulosCSV(raw string) (map[string]bool, []string) {
	allowed := map[string]bool{}
	ordered := make([]string, 0)
	for _, chunk := range strings.Split(raw, ",") {
		modulo := strings.ToLower(strings.TrimSpace(chunk))
		if modulo == "" || !isPermissionModuleKnown(modulo) {
			continue
		}
		if allowed[modulo] {
			continue
		}
		allowed[modulo] = true
		ordered = append(ordered, modulo)
	}
	return allowed, ordered
}

type empresaVerticalScope struct {
	Enabled           bool
	TipoEmpresaID     int64
	TipoEmpresaNombre string
	Allowed           map[string]bool
	AllowedList       []string
	Source            string
}

func (scope empresaVerticalScope) toContext() *empresaVerticalScopeCtx {
	if !scope.Enabled {
		return nil
	}
	return &empresaVerticalScopeCtx{
		Restringe:         true,
		TipoEmpresaID:     scope.TipoEmpresaID,
		TipoEmpresaNombre: strings.TrimSpace(scope.TipoEmpresaNombre),
		ModulosPermitidos: append([]string{}, scope.AllowedList...),
		Fuente:            strings.TrimSpace(scope.Source),
	}
}

func isEmpresaBusinessVerticalModule(module string) bool {
	clean := normalizeVerticalScopeModule(module)
	if clean == "" {
		return false
	}
	switch clean {
	case permModuleGimnasio,
		permModuleTaxiSystem,
		permModuleDomicilios,
		permModuleParqueadero,
		permModuleApartTuristicos,
		permModulePropiedadHorizontal,
		permModuleAlquileres,
		permModuleOdontologia,
		permModuleDrogueriaFarmacia,
		permModuleAIUConstruccion,
		permModuleReservasHotel:
		return true
	default:
		return isPermModuleNuevoVertical(clean)
	}
}

func normalizeVerticalScopeModule(module string) string {
	clean := strings.ToLower(strings.TrimSpace(module))
	clean = strings.ReplaceAll(clean, "-", "_")
	switch clean {
	case "consultorio", "consultorio_odontologico", "odontologico", "pacientes":
		return permModuleOdontologia
	case "taxi":
		return permModuleTaxiSystem
	case "apartamentos", "apartamento_turistico", "apartamentos_turisticos":
		return permModuleApartTuristicos
	case "propiedad", "copropiedad", "ph":
		return permModulePropiedadHorizontal
	case "alquiler", "rentas":
		return permModuleAlquileres
	case "drogueria", "farmacia":
		return permModuleDrogueriaFarmacia
	case "aiu", "constructora", "construccion":
		return permModuleAIUConstruccion
	case "hotel", "motel", "reservas":
		return permModuleReservasHotel
	default:
		if normalized := strings.TrimSpace(dbpkg.NormalizeEmpresaModuloColombia(clean)); normalized != "" {
			return normalized
		}
		return clean
	}
}

func addVerticalScopeModule(scope *empresaVerticalScope, module string) {
	if scope == nil {
		return
	}
	clean := normalizeVerticalScopeModule(module)
	if clean == "" || !isEmpresaBusinessVerticalModule(clean) {
		return
	}
	if scope.Allowed == nil {
		scope.Allowed = map[string]bool{}
	}
	if scope.Allowed[clean] {
		return
	}
	scope.Allowed[clean] = true
	scope.AllowedList = append(scope.AllowedList, clean)
}

func verticalModulesFromLicenciaPolicy(policy *dbpkg.LicenciaPermisoPolicy) []string {
	if policy == nil {
		return nil
	}
	out := make([]string, 0)
	seen := map[string]bool{}
	for _, chunk := range strings.Split(policy.ModulosHabilitados, ",") {
		clean := normalizeVerticalScopeModule(chunk)
		if clean == "" || !isEmpresaBusinessVerticalModule(clean) || seen[clean] {
			continue
		}
		seen[clean] = true
		out = append(out, clean)
	}
	return out
}

func resolveEmpresaVerticalScope(dbSuper *sql.DB, empresaID int64, policy *dbpkg.LicenciaPermisoPolicy) empresaVerticalScope {
	scope := empresaVerticalScope{Allowed: map[string]bool{}}
	if empresaID <= 0 || dbSuper == nil {
		return scope
	}

	var empresa *dbpkg.Empresa
	if dbEmp := dbpkg.GetDB(); dbEmp != nil {
		if item, err := dbpkg.GetEmpresaByScopeID(dbEmp, empresaID); err == nil && item != nil {
			empresa = item
		}
	}
	if empresa == nil {
		if item, err := dbpkg.GetEmpresaByID(dbSuper, empresaID); err == nil && item != nil {
			empresa = item
		}
	}

	activeLic, _ := dbpkg.GetActiveLicenciaByEmpresa(dbSuper, empresaID)
	tipoID := int64(0)
	tipoNombre := ""
	if empresa != nil {
		tipoID = empresa.TipoID
		tipoNombre = strings.TrimSpace(empresa.TipoNombre)
	}
	if activeLic != nil && activeLic.TipoID > 0 {
		tipoID = activeLic.TipoID
	}

	if tipoID > 0 || tipoNombre != "" {
		if preconfig, err := dbpkg.ResolveTipoEmpresaPreconfiguracion(dbSuper, tipoID, tipoNombre); err == nil && preconfig != nil && preconfig.Enabled {
			if template, parseErr := dbpkg.ParseTipoEmpresaPreconfigTemplate(preconfig.ConfigJSON); parseErr == nil {
				scope.TipoEmpresaID = tipoID
				scope.TipoEmpresaNombre = strings.TrimSpace(firstNonEmptyString(tipoNombre, preconfig.TipoEmpresaNombre))
				if template.IntegracionVertical != nil {
					addVerticalScopeModule(&scope, template.IntegracionVertical.Modulo)
					if len(scope.AllowedList) > 0 {
						scope.Source = "preconfiguracion_tipo_empresa"
					}
				}
				if template.Modulos.Gimnasio != nil {
					addVerticalScopeModule(&scope, permModuleGimnasio)
				}
				if template.Modulos.Odontologia != nil {
					addVerticalScopeModule(&scope, permModuleOdontologia)
				}
			}
		}
	}

	if len(scope.AllowedList) == 0 && empresa != nil {
		addVerticalScopeModule(&scope, empresa.TipoNombre)
		if len(scope.AllowedList) > 0 {
			scope.TipoEmpresaID = tipoID
			scope.TipoEmpresaNombre = strings.TrimSpace(empresa.TipoNombre)
			scope.Source = "tipo_empresa"
		}
	}

	if len(scope.AllowedList) == 0 {
		verticals := verticalModulesFromLicenciaPolicy(policy)
		if len(verticals) == 1 {
			addVerticalScopeModule(&scope, verticals[0])
			scope.TipoEmpresaID = tipoID
			scope.Source = "licencia_modulos"
		}
	}

	scope.Enabled = len(scope.AllowedList) > 0
	if scope.Source == "" && scope.Enabled {
		scope.Source = "alcance_vertical"
	}
	return scope
}

func isModuloPermitidoByVerticalScope(modulo string, scope empresaVerticalScope) bool {
	clean := normalizeVerticalScopeModule(modulo)
	if clean == "" || !isEmpresaBusinessVerticalModule(clean) {
		return true
	}
	if !scope.Enabled {
		return true
	}
	return scope.Allowed[clean]
}

func isPermissionModuleKnown(modulo string) bool {
	target := strings.ToLower(strings.TrimSpace(modulo))
	if target == "" {
		return false
	}
	for _, known := range permissionModulesCatalogOrdered {
		if target == strings.ToLower(strings.TrimSpace(known)) {
			return true
		}
	}
	return false
}

var permissionModuleLicenseFallbacks = map[string][]string{
	permModuleCRMUnificado:         {permModuleClientes},
	permModuleReservasHotel:        {permModuleVentas},
	permModuleChatTareas:           {permModuleVentas},
	permModuleHorariosTrab:         {permModuleSeguridad},
	permModuleAsistenciaEmpleados:  {permModuleSeguridad},
	permModuleVehiculosRegistro:    {permModuleSeguridad},
	permModuleHojaVidaOperativa:    {permModuleSeguridad},
	permModuleUbicacionGPS:         {permModuleInventario, permModuleSeguridad},
	permModuleNominaSueldos:        {permModuleFinanzas},
	permModuleReportes:             {permModuleFinanzas},
	permModuleAuditoria:            {permModuleSeguridad},
	permModuleBackups:              {permModuleSeguridad},
	permModuleDocumentosOnlyOffice: {permModuleSeguridad},
}

func isModuloPermitidoByLicencia(modulo string, allowed map[string]bool) bool {
	if len(allowed) == 0 {
		return true
	}
	key := strings.ToLower(strings.TrimSpace(modulo))
	if key == "" {
		return false
	}
	if allowed[key] {
		return true
	}
	for _, fallback := range permissionModuleLicenseFallbacks[key] {
		if allowed[strings.ToLower(strings.TrimSpace(fallback))] {
			return true
		}
	}
	return false
}

func applyLicenciaRestriccionesToModuleRows(rows []permissionModuleMatrixRow, allowed map[string]bool) []permissionModuleMatrixRow {
	if len(allowed) == 0 {
		return rows
	}
	out := make([]permissionModuleMatrixRow, 0, len(rows))
	for _, row := range rows {
		next := row
		next.Acciones = map[string]bool{}
		for _, action := range permissionActionsCatalogOrdered {
			next.Acciones[action] = row.Acciones[action]
		}
		if !isModuloPermitidoByLicencia(next.Modulo, allowed) {
			setPermissionActionOnModuleRow(&next, permActionRead, false)
			setPermissionActionOnModuleRow(&next, permActionCreate, false)
			setPermissionActionOnModuleRow(&next, permActionUpdate, false)
			setPermissionActionOnModuleRow(&next, permActionDelete, false)
			setPermissionActionOnModuleRow(&next, permActionApprove, false)
		}
		out = append(out, next)
	}
	return out
}

func applyEmpresaVerticalScopeToModuleRows(rows []permissionModuleMatrixRow, scope empresaVerticalScope) []permissionModuleMatrixRow {
	if !scope.Enabled {
		return rows
	}
	out := make([]permissionModuleMatrixRow, 0, len(rows))
	for _, row := range rows {
		next := row
		next.Acciones = map[string]bool{}
		for _, action := range permissionActionsCatalogOrdered {
			next.Acciones[action] = row.Acciones[action]
		}
		if !isModuloPermitidoByVerticalScope(next.Modulo, scope) {
			setPermissionActionOnModuleRow(&next, permActionRead, false)
			setPermissionActionOnModuleRow(&next, permActionCreate, false)
			setPermissionActionOnModuleRow(&next, permActionUpdate, false)
			setPermissionActionOnModuleRow(&next, permActionDelete, false)
			setPermissionActionOnModuleRow(&next, permActionApprove, false)
		}
		out = append(out, next)
	}
	return out
}

func applyEmpresaRestriccionesToModuleRows(rows []permissionModuleMatrixRow, overrides map[string]bool) []permissionModuleMatrixRow {
	if len(overrides) == 0 {
		return rows
	}
	out := make([]permissionModuleMatrixRow, 0, len(rows))
	for _, row := range rows {
		next := row
		next.Acciones = map[string]bool{}
		for _, action := range permissionActionsCatalogOrdered {
			next.Acciones[action] = row.Acciones[action]
		}
		for _, action := range permissionActionsCatalogOrdered {
			if permitido, ok := overrides[permissionModuleActionKey(next.Modulo, action)]; ok {
				setPermissionActionOnModuleRow(&next, action, permitido && row.Acciones[action])
			}
		}
		out = append(out, next)
	}
	return out
}

func parseAdminEmpresaCompartidaModulosPermitidosCSV(value string) (map[string]bool, []string) {
	allowed := map[string]bool{}
	list := []string{}
	for _, part := range strings.Split(value, ",") {
		modulo := strings.ToLower(strings.TrimSpace(part))
		if modulo == "" || allowed[modulo] || !isPermissionModuleKnown(modulo) {
			continue
		}
		allowed[modulo] = true
		list = append(list, modulo)
	}
	return allowed, list
}

func copyPermissionModuleRow(row permissionModuleMatrixRow) permissionModuleMatrixRow {
	next := row
	next.Acciones = map[string]bool{}
	for _, action := range permissionActionsCatalogOrdered {
		next.Acciones[action] = row.Acciones[action]
	}
	return next
}

func disablePermissionModuleRow(row *permissionModuleMatrixRow) {
	for _, action := range permissionActionsCatalogOrdered {
		setPermissionActionOnModuleRow(row, action, false)
	}
}

func applyAdminEmpresaCompartidaScopeToModuleRows(rows []permissionModuleMatrixRow, access *dbpkg.AdminEmpresaCompartidaAcceso) []permissionModuleMatrixRow {
	if access == nil {
		return rows
	}
	nivel := normalizeAdminEmpresaCompartidaNivel(access.NivelAcceso)
	if nivel == "acceso_total" {
		return rows
	}
	allowedModules, _ := parseAdminEmpresaCompartidaModulosPermitidosCSV(access.ModulosPermitidos)
	out := make([]permissionModuleMatrixRow, 0, len(rows))
	for _, row := range rows {
		next := copyPermissionModuleRow(row)
		switch nivel {
		case "solo_ver":
			for _, action := range []string{permActionCreate, permActionUpdate, permActionDelete, permActionApprove} {
				setPermissionActionOnModuleRow(&next, action, false)
			}
		case "modulos":
			if !allowedModules[strings.ToLower(strings.TrimSpace(row.Modulo))] {
				disablePermissionModuleRow(&next)
			}
		}
		out = append(out, next)
	}
	return out
}

func adminEmpresaCompartidaScopeContext(access *dbpkg.AdminEmpresaCompartidaAcceso) *empresaCompartidaScopeCtx {
	if access == nil {
		return nil
	}
	_, list := parseAdminEmpresaCompartidaModulosPermitidosCSV(access.ModulosPermitidos)
	nivel := normalizeAdminEmpresaCompartidaNivel(access.NivelAcceso)
	return &empresaCompartidaScopeCtx{
		Compartida:        true,
		NivelAcceso:       nivel,
		ModulosPermitidos: list,
		CompartidoPor:     strings.TrimSpace(access.CompartidoPorEmail),
		Etiqueta:          adminEmpresaCompartidaScopeLabel(nivel, access.ModulosPermitidos),
	}
}

func applyEmpresaPageRestrictionsToMap(paginas map[string]bool, overrides map[string]bool) map[string]bool {
	if len(overrides) == 0 {
		return paginas
	}
	out := make(map[string]bool, len(paginas))
	for k, v := range paginas {
		out[k] = v
	}
	for key, permitido := range overrides {
		clean := strings.TrimSpace(key)
		if clean == "" {
			continue
		}
		out[clean] = permitido && out[clean]
	}
	return out
}

func resolvePermissionPageKeyForRequest(r *http.Request) string {
	if r == nil || r.URL == nil {
		return ""
	}
	path := strings.ToLower(strings.TrimSpace(r.URL.Path))
	action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
	if page, ok := permissionPageForNuevoVerticalAPIPath(path); ok {
		return page
	}
	switch {
	case strings.HasPrefix(path, "/api/empresa/crm/"):
		return "linkCRMComercial"
	case path == "/api/empresa/crm_avanzado":
		return "linkCRMComercial"
	case path == "/api/empresa/clientes":
		return "linkClientes"
	case strings.HasPrefix(path, "/api/empresa/chat_tareas"):
		return "linkChatTareas"
	case strings.HasPrefix(path, "/api/empresa/chat_con_inteligencia_artificial"):
		return "linkChatIA"
	case path == "/api/empresa/impuestos":
		return "linkImpuestos"
	case strings.HasPrefix(path, "/api/empresa/facturacion_electronica"):
		if action == "emitir" || !strings.EqualFold(strings.TrimSpace(r.Method), http.MethodGet) {
			return "linkFacturacionElectronica"
		}
		return "linkFacturasElectronicas"
	case path == "/api/empresa/aiu_construccion":
		return "linkAIUConstruccion"
	case path == "/api/empresa/corte_caja":
		return "linkCorteCaja"
	case strings.HasPrefix(path, "/api/empresa/finanzas/"):
		return "linkFinanzas"
	case path == "/api/empresa/contabilidad_colombia":
		return "linkContabilidadColombia"
	case path == "/api/empresa/contabilidad_colombia_avanzada":
		return "linkContabilidadColombiaAvanzada"
	case path == "/api/empresa/centros_costo":
		return "linkCentrosCosto"
	case path == "/api/empresa/cierre_fiscal":
		return "linkCierreFiscal"
	case path == "/api/empresa/activos_fijos_niif_fiscal":
		return "linkActivosFijosNIIF"
	case path == "/api/empresa/declaraciones_tributarias":
		return "linkDeclaracionesTributarias"
	case path == "/api/empresa/cobranza":
		return "linkCobranza"
	case path == "/api/empresa/portal_contador":
		return "linkPortalContador"
	case path == "/api/empresa/portal_terceros_certificados":
		return "linkPortalTercerosCertificados"
	case path == "/api/empresa/bancos_pagos":
		return "linkBancosPagos"
	case path == "/api/empresa/gestion_documental":
		return "linkGestionDocumental"
	case path == "/api/empresa/cumplimiento_kyc":
		return "linkCumplimientoKYC"
	case path == "/api/empresa/contratos_obligaciones":
		return "linkContratosObligaciones"
	case path == "/api/empresa/calidad_procesos":
		return "linkCalidadProcesos"
	case strings.HasPrefix(path, "/api/empresa/creditos") ||
		strings.HasPrefix(path, "/api/empresa/cuentas_por_cobrar") ||
		strings.HasPrefix(path, "/api/empresa/cuentas_por_pagar"):
		return "linkCreditos"
	case path == "/api/empresa/propinas":
		return "linkPropinas"
	case path == "/api/empresa/comisiones":
		return "linkComisiones"
	case path == "/api/empresa/codigos_de_descuento":
		return "linkCodigosDescuento"
	case strings.HasPrefix(path, "/api/empresa/publicaciones"):
		return "linkRedSocialComercial"
	case path == "/api/empresa/venta_publica":
		if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("perm_page")), "linkCartaProductosPublica") {
			return "linkCartaProductosPublica"
		}
		return "linkVentaPublica"
	case strings.HasPrefix(path, "/api/empresa/carritos_compra"):
		modo := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("modo")))
		carritoCodigo := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("carrito_codigo")))
		permPage := strings.TrimSpace(r.URL.Query().Get("perm_page"))
		if strings.EqualFold(permPage, "linkVentaDirecta") || modo == "venta_directa" || strings.HasPrefix(carritoCodigo, "VENTA-DIRECTA") || strings.HasPrefix(carritoCodigo, "VENTA_DIRECTA") {
			return "linkVentaDirecta"
		}
		if strings.Contains(action, "estacion") || strings.TrimSpace(r.URL.Query().Get("estacion_id")) != "" {
			return "linkEstaciones"
		}
		return "linkCarritoCompras"
	case strings.HasPrefix(path, "/api/empresa/estaciones") ||
		strings.HasPrefix(path, "/api/empresa/estacion_") ||
		strings.HasPrefix(path, "/api/empresa/ventas_estacion"):
		return "linkEstaciones"
	case path == "/api/empresa/reservas_hotel":
		return "linkReservasHotel"
	case path == "/api/empresa/tarifas_por_minutos":
		return "linkTarifasPorMinutos"
	case path == "/api/empresa/tarifas_por_dia":
		return "linkTarifasPorDia"
	case path == "/api/empresa/tarifas_motel":
		return "linkTarifasMotel"
	case path == "/api/empresa/hotel_tarjetas_acceso":
		return "linkHotelTarjetasAcceso"
	case path == "/api/empresa/nomina_sueldos":
		return "linkNominaSueldos"
	case path == "/api/empresa/horarios_trabajadores":
		return "linkHorariosTrabajadores"
	case path == "/api/empresa/configuracion_guiada":
		return "linkConfiguracionGuiada"
	case path == "/api/empresa/asistencia_empleados":
		return "linkAsistenciaEmpleados"
	case path == "/api/empresa/vehiculos_registro":
		return "linkVehiculosRegistro"
	case path == "/api/empresa/carnets":
		return "linkCarnets"
	case path == "/api/empresa/gimnasio":
		return "linkGimnasio"
	case path == "/api/empresa/taxi_system":
		return "linkTaxiSystem"
	case path == "/api/empresa/domicilios":
		return "linkDomicilios"
	case path == "/api/empresa/parqueadero":
		return "linkParqueadero"
	case path == "/api/empresa/apartamentos_turisticos":
		return "linkApartamentosTuristicos"
	case path == "/api/empresa/propiedad_horizontal":
		return "linkPropiedadHorizontal"
	case path == "/api/empresa/alquileres":
		return "linkAlquileres"
	case path == "/api/empresa/odontologia":
		return "linkConsultorioOdontologico"
	case path == "/api/empresa/drogueria_farmacia":
		return "linkDrogueriaFarmacia"
	case path == "/api/empresa/turnos_atencion":
		return "linkTurnosAtencion"
	case path == "/api/empresa/control_electrico":
		return "linkControlElectrico"
	case path == "/api/empresa/produccion_mrp":
		return "linkProduccionMRP"
	case path == "/api/empresa/logistica_wms":
		return "linkLogisticaWMS"
	case path == "/api/empresa/inventario_avanzado":
		return "linkInventarioAvanzado"
	case path == "/api/empresa/importaciones_costeo":
		return "linkImportacionesCosteo"
	case path == "/api/empresa/soportes_compras_ia":
		return "linkSoportesComprasIA"
	case path == "/api/empresa/compras_avanzadas":
		return "linkComprasAvanzadas"
	case path == "/api/empresa/tesoreria_presupuesto":
		return "linkTesoreriaPresupuesto"
	case path == "/api/empresa/documentos":
		return "linkDocumentosOnlyOffice"
	case strings.HasPrefix(path, "/api/empresa/reportes"):
		return "linkReportes"
	}
	return ""
}

func getEmpresaPermissionSnapshot(dbEmp, dbSuper *sql.DB, adminEmail string, empresaID int64) (empresaPermissionSnapshot, error) {
	startedAt := time.Now()
	defer func() {
		dbpkg.PerfLogf("[perf][authz] getEmpresaPermissionSnapshot empresa=%d email=%s dur=%s", empresaID, strings.ToLower(strings.TrimSpace(adminEmail)), time.Since(startedAt))
	}()
	cacheKey := strings.ToLower(strings.TrimSpace(adminEmail)) + "|" + strconv.FormatInt(empresaID, 10)
	if strings.TrimSpace(adminEmail) == "" || empresaID <= 0 {
		return empresaPermissionSnapshot{}, sql.ErrNoRows
	}
	var snapshotResult empresaPermissionSnapshot
	var snapshotErr error

	empresaPermissionCacheMu.Lock()
	if cached, ok := empresaPermissionCache[cacheKey]; ok && time.Since(cached.LoadedAt) < empresaPermissionCacheTTL {
		empresaPermissionCacheMu.Unlock()
		return cached, nil
	}
	if inflight, ok := empresaPermissionInflight[cacheKey]; ok {
		empresaPermissionCacheMu.Unlock()
		<-inflight.done
		return inflight.snapshot, inflight.err
	}
	inflight := &empresaPermissionSnapshotInflight{done: make(chan struct{})}
	empresaPermissionInflight[cacheKey] = inflight
	empresaPermissionCacheMu.Unlock()
	defer func() {
		empresaPermissionCacheMu.Lock()
		delete(empresaPermissionInflight, cacheKey)
		inflight.snapshot = snapshotResult
		inflight.err = snapshotErr
		close(inflight.done)
		empresaPermissionCacheMu.Unlock()
	}()
	stepStarted := time.Now()
	admin, err := dbpkg.GetAdminByEmail(dbSuper, adminEmail)
	if err != nil {
		snapshotErr = err
		return empresaPermissionSnapshot{}, err
	}
	dbpkg.PerfLogf("[perf][authz] snapshot empresa=%d email=%s step=admin dur=%s", empresaID, strings.ToLower(strings.TrimSpace(adminEmail)), time.Since(stepStarted))
	role := normalizePermissionRole(admin.Role)

	var (
		canAccess              bool
		canAccessErr           error
		licenciaPolicy         *dbpkg.LicenciaPermisoPolicy
		licenciaErr            error
		moduleRows             []permissionModuleMatrixRow
		empresaModuleOverrides map[string]bool
		empresaPageOverrides   map[string]bool
		sharedAccess           *dbpkg.AdminEmpresaCompartidaAcceso
		sharedAccessErr        error
	)

	var snapshotWG sync.WaitGroup
	snapshotWG.Add(5)

	go func() {
		defer snapshotWG.Done()
		step := time.Now()
		canAccess, canAccessErr = dbpkg.CanAdminAccessEmpresaIA(dbEmp, dbSuper, adminEmail, empresaID)
		dbpkg.PerfLogf("[perf][authz] snapshot empresa=%d email=%s step=access dur=%s", empresaID, strings.ToLower(strings.TrimSpace(adminEmail)), time.Since(step))
	}()

	go func() {
		defer snapshotWG.Done()
		step := time.Now()
		licenciaPolicy, licenciaErr = dbpkg.GetLicenciaPermisoPolicyByEmpresa(dbSuper, empresaID)
		dbpkg.PerfLogf("[perf][authz] snapshot empresa=%d email=%s step=licencia dur=%s", empresaID, strings.ToLower(strings.TrimSpace(adminEmail)), time.Since(step))
	}()

	go func() {
		defer snapshotWG.Done()
		step := time.Now()
		moduleRows = buildPermissionModuleMatrixForRoleDynamic(dbSuper, role)
		dbpkg.PerfLogf("[perf][authz] snapshot empresa=%d email=%s step=module_rows dur=%s", empresaID, strings.ToLower(strings.TrimSpace(adminEmail)), time.Since(step))
	}()

	go func() {
		defer snapshotWG.Done()
		step := time.Now()
		empresaModuleOverrides, empresaPageOverrides, _ = loadEmpresaPermissionOverrides(dbSuper, empresaID)
		dbpkg.PerfLogf("[perf][authz] snapshot empresa=%d email=%s step=empresa_overrides dur=%s", empresaID, strings.ToLower(strings.TrimSpace(adminEmail)), time.Since(step))
	}()

	go func() {
		defer snapshotWG.Done()
		step := time.Now()
		sharedAccess, sharedAccessErr = dbpkg.GetActiveAdminEmpresaCompartidaAcceso(dbSuper, empresaID, adminEmail)
		dbpkg.PerfLogf("[perf][authz] snapshot empresa=%d email=%s step=shared_scope dur=%s", empresaID, strings.ToLower(strings.TrimSpace(adminEmail)), time.Since(step))
	}()

	snapshotWG.Wait()
	if canAccessErr != nil {
		snapshotErr = canAccessErr
		return empresaPermissionSnapshot{}, canAccessErr
	}
	if licenciaErr != nil {
		snapshotErr = licenciaErr
		return empresaPermissionSnapshot{}, licenciaErr
	}
	if sharedAccessErr != nil {
		snapshotErr = sharedAccessErr
		return empresaPermissionSnapshot{}, sharedAccessErr
	}

	allowedModules, _ := parseLicenciaModulosCSV("")
	if licenciaPolicy != nil {
		allowedModules, _ = parseLicenciaModulosCSV(licenciaPolicy.ModulosHabilitados)
	}
	verticalScope := resolveEmpresaVerticalScope(dbSuper, empresaID, licenciaPolicy)
	effectiveRole := resolveEffectiveRoleByLicencia(role, licenciaPolicy)
	if effectiveRole != role {
		stepStarted = time.Now()
		moduleRows = buildPermissionModuleMatrixForRoleDynamic(dbSuper, effectiveRole)
		dbpkg.PerfLogf("[perf][authz] snapshot empresa=%d email=%s step=module_rows_effective dur=%s", empresaID, strings.ToLower(strings.TrimSpace(adminEmail)), time.Since(stepStarted))
	}
	moduleRows = applyLicenciaRestriccionesToModuleRows(moduleRows, allowedModules)
	moduleRows = applyEmpresaVerticalScopeToModuleRows(moduleRows, verticalScope)
	moduleRows = applyEmpresaRestriccionesToModuleRows(moduleRows, empresaModuleOverrides)
	moduleRows = applyAdminEmpresaCompartidaScopeToModuleRows(moduleRows, sharedAccess)
	stepStarted = time.Now()
	allowedPages := buildPermissionPagesMapForRoleDynamic(dbSuper, effectiveRole, moduleRows)
	dbpkg.PerfLogf("[perf][authz] snapshot empresa=%d email=%s step=allowed_pages dur=%s", empresaID, strings.ToLower(strings.TrimSpace(adminEmail)), time.Since(stepStarted))
	allowedPages = applyEmpresaPageRestrictionsToMap(allowedPages, empresaPageOverrides)

	roleModuleActions := map[string]bool{}
	for _, row := range moduleRows {
		for _, permissionAction := range permissionActionsCatalogOrdered {
			roleModuleActions[permissionModuleActionKey(row.Modulo, permissionAction)] = row.Acciones[permissionAction]
		}
	}

	snapshot := empresaPermissionSnapshot{
		AdminRole:              role,
		EffectiveRole:          effectiveRole,
		CanAccess:              canAccess,
		AllowedModules:         allowedModules,
		AllowedVerticalModules: verticalScope.Allowed,
		RoleModuleActions:      roleModuleActions,
		AllowedPages:           allowedPages,
		ShareAccess:            adminEmpresaCompartidaScopeContext(sharedAccess),
		LoadedAt:               time.Now(),
	}

	empresaPermissionCacheMu.Lock()
	empresaPermissionCache[cacheKey] = snapshot
	empresaPermissionCacheMu.Unlock()
	snapshotResult = snapshot
	return snapshot, nil
}

func resolveEffectiveRoleByLicencia(role string, licenciaPolicy *dbpkg.LicenciaPermisoPolicy) string {
	resolved := normalizePermissionRole(role)
	if licenciaPolicy == nil || !licenciaPolicy.SuperRolHabilitado {
		return resolved
	}
	if resolved == "supervisor_sucursal" {
		return "admin_empresa"
	}
	return resolved
}

func summarizePermissionModules(rows []permissionModuleMatrixRow) permissionSummary {
	summary := permissionSummary{ModulosTotal: len(rows)}
	for _, row := range rows {
		if row.Read {
			summary.ModulosLectura++
			summary.AccionesHabilitadas++
		}
		if row.Create {
			summary.AccionesHabilitadas++
		}
		if row.Update {
			summary.AccionesHabilitadas++
		}
		if row.Delete {
			summary.AccionesHabilitadas++
		}
		if row.Approve {
			summary.ModulosAprobacion++
			summary.AccionesHabilitadas++
		}
	}
	return summary
}

// PermissionModuleDisplayNameMap devuelve etiquetas de negocio por clave de módulo (API super: permisos por rol).
func PermissionModuleDisplayNameMap() map[string]string {
	out := make(map[string]string, len(permissionModulesCatalogOrdered))
	for _, m := range permissionModulesCatalogOrdered {
		if lab, ok := permissionModuleDisplayNames[m]; ok && strings.TrimSpace(lab) != "" {
			out[m] = universalPermissionModuleLabel(lab)
		} else {
			out[m] = universalPermissionModuleLabel(m)
		}
	}
	return out
}

// PermissionActionDisplayNameMap devuelve etiquetas por letra de acción (R/C/U/D/A).
func PermissionActionDisplayNameMap() map[string]string {
	out := make(map[string]string, len(permissionActionsCatalogOrdered))
	for _, a := range permissionActionsCatalogOrdered {
		if lab, ok := permissionActionDisplayNames[a]; ok && strings.TrimSpace(lab) != "" {
			out[a] = sanitizeLegacyPermissionVisibleText(lab)
		} else {
			out[a] = sanitizeLegacyPermissionVisibleText(a)
		}
	}
	return out
}
