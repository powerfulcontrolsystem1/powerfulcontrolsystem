package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type empresaModuloGenericConfig struct {
	Table            string
	SearchColumns    []string
	AllowedColumns   []string
	RequiredOnCreate []string
	CodeColumn       string
	CodePrefix       string
	DefaultValues    map[string]interface{}
}

type empresaModuloStateMachineConfig struct {
	ModuleName   string
	StateColumn  string
	InitialState string
	Transitions  map[string][]string
}

type empresaModuloIntegracionesOpsConfig struct {
	ModuleName    string
	EndpointField string
	LastSyncField string
	ResponseField string
	NameField     string
}

type empresaIntegracionProbeResult struct {
	ID                int64  `json:"id"`
	Codigo            string `json:"codigo,omitempty"`
	Nombre            string `json:"nombre,omitempty"`
	Endpoint          string `json:"endpoint,omitempty"`
	HTTPStatus        int    `json:"http_status,omitempty"`
	Reachable         bool   `json:"reachable"`
	LatencyMS         int64  `json:"latency_ms,omitempty"`
	EstadoIntegracion string `json:"estado_integracion"`
	Message           string `json:"message,omitempty"`
	Updated           bool   `json:"updated"`
}

var (
	cfgCotizacionesVenta = empresaModuloGenericConfig{
		Table:         "empresa_cotizaciones_venta",
		SearchColumns: []string{"codigo", "cliente_nombre", "estado_documento", "notas"},
		AllowedColumns: []string{
			"codigo", "cliente_id", "cliente_nombre", "fecha_documento", "vigencia_hasta", "estado_documento",
			"subtotal", "descuento_total", "impuesto_total", "total", "moneda", "notas", "origen",
			"convertido_pedido_id", "usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"cliente_nombre"},
		CodeColumn:       "codigo",
		CodePrefix:       "COT",
		DefaultValues: map[string]interface{}{
			"estado_documento": "borrador",
			"moneda":           "COP",
		},
	}

	cfgPedidosVenta = empresaModuloGenericConfig{
		Table:         "empresa_pedidos_venta",
		SearchColumns: []string{"codigo", "cliente_nombre", "estado_pedido", "notas"},
		AllowedColumns: []string{
			"codigo", "cliente_id", "cliente_nombre", "cotizacion_id", "fecha_pedido", "fecha_entrega_estimada",
			"estado_pedido", "subtotal", "descuento_total", "impuesto_total", "total", "moneda", "notas",
			"usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"cliente_nombre"},
		CodeColumn:       "codigo",
		CodePrefix:       "PED",
		DefaultValues: map[string]interface{}{
			"estado_pedido": "borrador",
			"moneda":        "COP",
		},
	}

	cfgDevolucionesVenta = empresaModuloGenericConfig{
		Table:         "empresa_devoluciones_venta",
		SearchColumns: []string{"codigo", "documento_referencia", "motivo", "estado_devolucion"},
		AllowedColumns: []string{
			"codigo", "carrito_id", "documento_referencia", "motivo", "fecha_devolucion", "estado_devolucion",
			"subtotal", "impuesto_total", "total", "moneda", "usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"motivo"},
		CodeColumn:       "codigo",
		CodePrefix:       "DEVV",
		DefaultValues: map[string]interface{}{
			"estado_devolucion": "borrador",
			"moneda":            "COP",
		},
	}

	cfgPlanCuentas = empresaModuloGenericConfig{
		Table:         "empresa_plan_cuentas",
		SearchColumns: []string{"codigo", "nombre", "tipo_cuenta"},
		AllowedColumns: []string{
			"codigo", "nombre", "tipo_cuenta", "naturaleza", "nivel", "cuenta_padre_codigo", "admite_movimiento",
			"aplica_impuesto", "usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"codigo", "nombre"},
		DefaultValues: map[string]interface{}{
			"tipo_cuenta":       "activo",
			"naturaleza":        "debito",
			"nivel":             1,
			"admite_movimiento": 1,
		},
	}

	cfgCxC = empresaModuloGenericConfig{
		Table:         "empresa_cuentas_por_cobrar",
		SearchColumns: []string{"codigo", "cliente_nombre", "documento_codigo", "estado_cartera"},
		AllowedColumns: []string{
			"codigo", "cliente_id", "cliente_nombre", "documento_tipo", "documento_codigo", "fecha_emision",
			"fecha_vencimiento", "dias_mora", "valor_original", "valor_pagado", "saldo", "estado_cartera",
			"moneda", "usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"cliente_nombre", "documento_codigo"},
		CodeColumn:       "codigo",
		CodePrefix:       "CXC",
		DefaultValues: map[string]interface{}{
			"estado_cartera": "pendiente",
			"moneda":         "COP",
		},
	}

	cfgCxP = empresaModuloGenericConfig{
		Table:         "empresa_cuentas_por_pagar",
		SearchColumns: []string{"codigo", "proveedor_nombre", "documento_codigo", "estado_cartera"},
		AllowedColumns: []string{
			"codigo", "proveedor_id", "proveedor_nombre", "documento_tipo", "documento_codigo", "fecha_emision",
			"fecha_vencimiento", "dias_mora", "valor_original", "valor_pagado", "saldo", "estado_cartera",
			"moneda", "usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"proveedor_nombre", "documento_codigo"},
		CodeColumn:       "codigo",
		CodePrefix:       "CXP",
		DefaultValues: map[string]interface{}{
			"estado_cartera": "pendiente",
			"moneda":         "COP",
		},
	}

	cfgLotesSeries = empresaModuloGenericConfig{
		Table:         "inventario_lotes_series",
		SearchColumns: []string{"codigo_lote_serie", "tipo_control", "estado_lote"},
		AllowedColumns: []string{
			"producto_id", "bodega_id", "tipo_control", "codigo_lote_serie", "fecha_fabricacion", "fecha_vencimiento",
			"cantidad_inicial", "cantidad_disponible", "costo_unitario", "estado_lote", "usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"producto_id", "codigo_lote_serie"},
		DefaultValues: map[string]interface{}{
			"tipo_control": "lote",
			"estado_lote":  "activo",
		},
	}

	cfgDevProveedor = empresaModuloGenericConfig{
		Table:         "empresa_devoluciones_proveedor",
		SearchColumns: []string{"codigo", "proveedor_nombre", "documento_compra_codigo", "estado_devolucion"},
		AllowedColumns: []string{
			"codigo", "proveedor_id", "proveedor_nombre", "documento_compra_codigo", "fecha_devolucion",
			"motivo", "estado_devolucion", "subtotal", "impuesto_total", "total", "moneda", "usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"proveedor_nombre", "motivo"},
		CodeColumn:       "codigo",
		CodePrefix:       "DPROV",
		DefaultValues: map[string]interface{}{
			"estado_devolucion": "borrador",
			"moneda":            "COP",
		},
	}

	cfgRRHHVacLic = empresaModuloGenericConfig{
		Table:         "empresa_rrhh_vacaciones_licencias",
		SearchColumns: []string{"codigo", "empleado_nombre", "tipo_novedad", "estado_novedad"},
		AllowedColumns: []string{
			"codigo", "empleado_id", "empleado_nombre", "tipo_novedad", "fecha_inicio", "fecha_fin", "dias",
			"remunerada", "estado_novedad", "soporte_url", "aprobado_por", "usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"empleado_nombre", "tipo_novedad", "fecha_inicio", "fecha_fin"},
		CodeColumn:       "codigo",
		CodePrefix:       "RRHH",
		DefaultValues: map[string]interface{}{
			"tipo_novedad":   "vacacion",
			"estado_novedad": "solicitada",
			"remunerada":     1,
		},
	}

	cfgCRMLeads = empresaModuloGenericConfig{
		Table:         "crm_leads",
		SearchColumns: []string{"codigo", "nombre", "empresa_origen", "email", "estado_lead"},
		AllowedColumns: []string{
			"codigo", "nombre", "empresa_origen", "email", "telefono", "canal_origen", "estado_lead",
			"valor_potencial", "probabilidad", "propietario", "proximo_contacto", "notas", "usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"nombre"},
		CodeColumn:       "codigo",
		CodePrefix:       "LEAD",
		DefaultValues: map[string]interface{}{
			"estado_lead": "nuevo",
		},
	}

	cfgCRMInteracciones = empresaModuloGenericConfig{
		Table:         "crm_interacciones",
		SearchColumns: []string{"codigo", "tipo_interaccion", "resultado", "usuario_responsable"},
		AllowedColumns: []string{
			"codigo", "lead_id", "cliente_id", "tipo_interaccion", "fecha_interaccion", "resumen", "resultado",
			"usuario_responsable", "proxima_accion", "estado_interaccion", "usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"tipo_interaccion", "resumen"},
		CodeColumn:       "codigo",
		CodePrefix:       "INT",
		DefaultValues: map[string]interface{}{
			"estado_interaccion": "abierta",
		},
	}

	cfgCRMCampanas = empresaModuloGenericConfig{
		Table:         "crm_campanas",
		SearchColumns: []string{"codigo", "nombre", "canal", "estado_campana"},
		AllowedColumns: []string{
			"codigo", "nombre", "canal", "objetivo", "presupuesto", "fecha_inicio", "fecha_fin",
			"estado_campana", "audiencia", "kpi_objetivo", "resultado_json", "usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"nombre", "canal"},
		CodeColumn:       "codigo",
		CodePrefix:       "CAMP",
		DefaultValues: map[string]interface{}{
			"estado_campana": "planificada",
		},
	}

	cfgProduccionBOM = empresaModuloGenericConfig{
		Table:         "produccion_bom",
		SearchColumns: []string{"codigo", "producto_nombre", "version", "estado_bom"},
		AllowedColumns: []string{
			"codigo", "producto_id", "producto_nombre", "version", "rendimiento", "unidad_medida",
			"costo_estimado_total", "estado_bom", "usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"codigo", "producto_nombre"},
		DefaultValues: map[string]interface{}{
			"version":    "1.0",
			"estado_bom": "activo",
		},
	}

	cfgProduccionBOMDetalle = empresaModuloGenericConfig{
		Table:         "produccion_bom_detalle",
		SearchColumns: []string{"insumo_nombre", "unidad_medida"},
		AllowedColumns: []string{
			"bom_id", "insumo_producto_id", "insumo_nombre", "cantidad", "unidad_medida",
			"costo_unitario", "costo_total", "merma_porcentaje", "usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"bom_id", "insumo_nombre", "cantidad"},
	}

	cfgProduccionOrdenes = empresaModuloGenericConfig{
		Table:         "produccion_ordenes",
		SearchColumns: []string{"codigo", "producto_nombre", "estado_orden", "responsable"},
		AllowedColumns: []string{
			"codigo", "bom_id", "producto_id", "producto_nombre", "cantidad_programada", "cantidad_producida",
			"fecha_programada", "fecha_inicio", "fecha_fin", "estado_orden", "costo_estimado", "costo_real",
			"responsable", "notas", "usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"producto_nombre", "cantidad_programada"},
		CodeColumn:       "codigo",
		CodePrefix:       "OP",
		DefaultValues: map[string]interface{}{
			"estado_orden": "planificada",
		},
	}

	cfgLogisticaTransportistas = empresaModuloGenericConfig{
		Table:         "logistica_transportistas",
		SearchColumns: []string{"codigo", "nombre", "documento", "placa"},
		AllowedColumns: []string{
			"codigo", "nombre", "documento", "telefono", "email", "placa", "vehiculo_tipo",
			"capacidad_carga", "estado_transportista", "usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"nombre"},
		CodeColumn:       "codigo",
		CodePrefix:       "TRN",
		DefaultValues: map[string]interface{}{
			"estado_transportista": "activo",
		},
	}

	cfgLogisticaRutas = empresaModuloGenericConfig{
		Table:         "logistica_rutas",
		SearchColumns: []string{"codigo", "nombre", "origen", "destino"},
		AllowedColumns: []string{
			"codigo", "nombre", "origen", "destino", "distancia_km", "tiempo_estimado_min",
			"estado_ruta", "usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"nombre", "origen", "destino"},
		CodeColumn:       "codigo",
		CodePrefix:       "RUT",
		DefaultValues: map[string]interface{}{
			"estado_ruta": "activa",
		},
	}

	cfgLogisticaEnvios = empresaModuloGenericConfig{
		Table:         "logistica_envios",
		SearchColumns: []string{"codigo", "cliente_nombre", "documento_referencia", "estado_envio"},
		AllowedColumns: []string{
			"codigo", "cliente_id", "cliente_nombre", "documento_referencia", "direccion_entrega", "ruta_id",
			"transportista_id", "fecha_programada", "fecha_salida", "fecha_entrega", "estado_envio", "costo_envio",
			"latitud", "longitud", "observaciones_seguimiento", "usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"cliente_nombre", "direccion_entrega"},
		CodeColumn:       "codigo",
		CodePrefix:       "ENV",
		DefaultValues: map[string]interface{}{
			"estado_envio": "programado",
		},
	}

	cfgDocumentosGestion = empresaModuloGenericConfig{
		Table:         "empresa_documentos_gestion",
		SearchColumns: []string{"codigo", "modulo", "entidad", "nombre_documento", "documento_codigo"},
		AllowedColumns: []string{
			"codigo", "modulo", "entidad", "entidad_id", "documento_codigo", "nombre_documento", "tipo_documento",
			"mime_type", "url_archivo", "hash_archivo", "tamano_bytes", "version", "estado_documento",
			"usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"modulo", "entidad", "nombre_documento"},
		CodeColumn:       "codigo",
		CodePrefix:       "DOC",
		DefaultValues: map[string]interface{}{
			"version":          "1",
			"estado_documento": "vigente",
		},
	}

	cfgDocumentosFirmas = empresaModuloGenericConfig{
		Table:         "empresa_documentos_firmas",
		SearchColumns: []string{"codigo", "tipo_firma", "firmante_nombre", "estado_firma"},
		AllowedColumns: []string{
			"codigo", "documento_gestion_id", "tipo_firma", "firmante_nombre", "firmante_documento", "firmante_email",
			"certificado_serial", "algoritmo_firma", "hash_firma", "fecha_firma", "validez_hasta", "estado_firma",
			"usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"documento_gestion_id", "tipo_firma", "firmante_nombre"},
		CodeColumn:       "codigo",
		CodePrefix:       "FIR",
		DefaultValues: map[string]interface{}{
			"tipo_firma":      "digital",
			"algoritmo_firma": "SHA256",
			"estado_firma":    "pendiente",
		},
	}

	cfgIntegracionesAPIs = empresaModuloGenericConfig{
		Table:         "empresa_integraciones_apis",
		SearchColumns: []string{"codigo", "nombre_integracion", "tipo_integracion", "base_url", "estado_integracion"},
		AllowedColumns: []string{
			"codigo", "nombre_integracion", "tipo_integracion", "base_url", "auth_tipo", "api_key_ref",
			"estado_integracion", "ultima_sincronizacion", "respuesta_ultimo_sync", "usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"nombre_integracion", "tipo_integracion"},
		CodeColumn:       "codigo",
		CodePrefix:       "API",
		DefaultValues: map[string]interface{}{
			"estado_integracion": "inactiva",
		},
	}

	cfgIntegracionesBancos = empresaModuloGenericConfig{
		Table:         "empresa_integraciones_bancos",
		SearchColumns: []string{"codigo", "banco_nombre", "numero_cuenta", "estado_integracion"},
		AllowedColumns: []string{
			"codigo", "banco_nombre", "tipo_conexion", "numero_cuenta", "titular", "moneda", "api_endpoint",
			"credencial_ref", "estado_integracion", "ultima_conciliacion", "usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"banco_nombre", "numero_cuenta"},
		CodeColumn:       "codigo",
		CodePrefix:       "BANK",
		DefaultValues: map[string]interface{}{
			"estado_integracion": "inactiva",
			"moneda":             "COP",
		},
	}

	cfgDIAN = empresaModuloGenericConfig{
		Table:         "empresa_dian_configuracion",
		SearchColumns: []string{"nit", "razon_social", "tipo_ambiente", "estado_dian", "prefijo", "resolucion_numero"},
		AllowedColumns: []string{
			"codigo", "nit", "digito_verificacion", "razon_social", "tipo_ambiente", "software_id", "software_pin",
			"test_set_id", "certificado_url", "certificado_clave_ref", "prefijo", "resolucion_numero",
			"resolucion_fecha_desde", "resolucion_fecha_hasta", "rango_desde", "rango_hasta", "consecutivo_actual",
			"url_dian", "token_emisor_ref", "ultimo_envio", "estado_dian", "usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"nit", "razon_social", "tipo_ambiente"},
		CodeColumn:       "codigo",
		CodePrefix:       "DIAN",
		DefaultValues: map[string]interface{}{
			"tipo_ambiente": "habilitacion",
			"estado_dian":   "pendiente",
			"url_dian":      "https://vpfe-hab.dian.gov.co",
		},
	}

	stateMachineCotizaciones = empresaModuloStateMachineConfig{
		ModuleName:   "ventas_cotizaciones",
		StateColumn:  "estado_documento",
		InitialState: "borrador",
		Transitions: map[string][]string{
			"borrador":   {"emitida", "anulada"},
			"emitida":    {"aprobada", "rechazada", "vencida", "anulada"},
			"aprobada":   {"convertida", "anulada"},
			"rechazada":  {"borrador", "anulada"},
			"vencida":    {"borrador", "anulada"},
			"convertida": []string{},
			"anulada":    []string{},
		},
	}

	stateMachinePedidos = empresaModuloStateMachineConfig{
		ModuleName:   "ventas_pedidos",
		StateColumn:  "estado_pedido",
		InitialState: "borrador",
		Transitions: map[string][]string{
			"borrador":       {"confirmado", "cancelado"},
			"confirmado":     {"en_preparacion", "cancelado"},
			"en_preparacion": {"despachado", "cancelado"},
			"despachado":     {"entregado", "devuelto", "cancelado"},
			"entregado":      {"cerrado"},
			"devuelto":       {"cerrado"},
			"cancelado":      []string{},
			"cerrado":        []string{},
		},
	}

	stateMachineDevoluciones = empresaModuloStateMachineConfig{
		ModuleName:   "ventas_devoluciones",
		StateColumn:  "estado_devolucion",
		InitialState: "borrador",
		Transitions: map[string][]string{
			"borrador":   {"solicitada", "anulada"},
			"solicitada": {"aprobada", "rechazada", "anulada"},
			"aprobada":   {"aplicada", "anulada"},
			"rechazada":  {"borrador", "anulada"},
			"aplicada":   {"cerrada"},
			"cerrada":    []string{},
			"anulada":    []string{},
		},
	}

	stateMachineCRMLeads = empresaModuloStateMachineConfig{
		ModuleName:   "crm_leads",
		StateColumn:  "estado_lead",
		InitialState: "nuevo",
		Transitions: map[string][]string{
			"nuevo":         {"contactado", "descalificado"},
			"contactado":    {"calificado", "descalificado"},
			"calificado":    {"propuesta", "descalificado"},
			"propuesta":     {"negociacion", "ganado", "perdido", "descalificado"},
			"negociacion":   {"ganado", "perdido", "descalificado"},
			"perdido":       {"reactivado"},
			"reactivado":    {"contactado", "calificado", "descalificado"},
			"ganado":        {"postventa"},
			"postventa":     {"cerrado"},
			"descalificado": []string{},
			"cerrado":       []string{},
		},
	}

	stateMachineCRMInteracciones = empresaModuloStateMachineConfig{
		ModuleName:   "crm_interacciones",
		StateColumn:  "estado_interaccion",
		InitialState: "abierta",
		Transitions: map[string][]string{
			"abierta":     {"en_progreso", "cerrada", "cancelada"},
			"en_progreso": {"cerrada", "cancelada"},
			"cerrada":     {"reabierta"},
			"reabierta":   {"en_progreso", "cerrada", "cancelada"},
			"cancelada":   {"reabierta"},
		},
	}

	stateMachineCRMCampanas = empresaModuloStateMachineConfig{
		ModuleName:   "crm_campanas",
		StateColumn:  "estado_campana",
		InitialState: "planificada",
		Transitions: map[string][]string{
			"planificada": {"activa", "cancelada"},
			"activa":      {"pausada", "finalizada", "cancelada"},
			"pausada":     {"activa", "finalizada", "cancelada"},
			"finalizada":  {"archivada"},
			"archivada":   []string{},
			"cancelada":   []string{},
		},
	}

	integrationOpsAPIs = empresaModuloIntegracionesOpsConfig{
		ModuleName:    "integraciones_apis",
		EndpointField: "base_url",
		LastSyncField: "ultima_sincronizacion",
		ResponseField: "respuesta_ultimo_sync",
		NameField:     "nombre_integracion",
	}

	integrationOpsBancos = empresaModuloIntegracionesOpsConfig{
		ModuleName:    "integraciones_bancos",
		EndpointField: "api_endpoint",
		LastSyncField: "ultima_conciliacion",
		ResponseField: "",
		NameField:     "banco_nombre",
	}
)

// RegisterEmpresaModulosFaltantesRoutes registra endpoints para modulos faltantes ERP/POS.
func RegisterEmpresaModulosFaltantesRoutes(dbEmp, dbSuper *sql.DB) {
	http.HandleFunc("/api/empresa/ventas/cotizaciones", WithEmpresaVentasPermissions(dbEmp, dbSuper, EmpresaVentasCotizacionesHandler(dbEmp)))
	http.HandleFunc("/api/empresa/ventas/pedidos", WithEmpresaVentasPermissions(dbEmp, dbSuper, EmpresaVentasPedidosHandler(dbEmp)))
	http.HandleFunc("/api/empresa/ventas/devoluciones", WithEmpresaVentasPermissions(dbEmp, dbSuper, EmpresaVentasDevolucionesHandler(dbEmp)))

	http.HandleFunc("/api/empresa/finanzas/plan_cuentas", WithEmpresaFinanzasPermissions(dbEmp, dbSuper, empresaModuloGenericCRUDHandler(dbEmp, cfgPlanCuentas)))
	http.HandleFunc("/api/empresa/finanzas/cuentas_cobrar", WithEmpresaFinanzasPermissions(dbEmp, dbSuper, empresaModuloGenericCRUDHandler(dbEmp, cfgCxC)))
	http.HandleFunc("/api/empresa/finanzas/cuentas_pagar", WithEmpresaFinanzasPermissions(dbEmp, dbSuper, empresaModuloGenericCRUDHandler(dbEmp, cfgCxP)))

	http.HandleFunc("/api/empresa/inventario/lotes_series", WithEmpresaInventarioPermissions(dbEmp, dbSuper, empresaModuloGenericCRUDHandler(dbEmp, cfgLotesSeries)))
	http.HandleFunc("/api/empresa/compras/devoluciones_proveedor", WithEmpresaComprasPermissions(dbEmp, dbSuper, empresaModuloGenericCRUDHandler(dbEmp, cfgDevProveedor)))
	http.HandleFunc("/api/empresa/rrhh/vacaciones_licencias", WithEmpresaSeguridadPermissions(dbEmp, dbSuper, empresaModuloGenericCRUDHandler(dbEmp, cfgRRHHVacLic)))

	http.HandleFunc("/api/empresa/crm/leads", WithEmpresaClientesPermissions(dbEmp, dbSuper, EmpresaCRMLeadsHandler(dbEmp)))
	http.HandleFunc("/api/empresa/crm/interacciones", WithEmpresaClientesPermissions(dbEmp, dbSuper, EmpresaCRMInteraccionesHandler(dbEmp)))
	http.HandleFunc("/api/empresa/crm/campanas", WithEmpresaClientesPermissions(dbEmp, dbSuper, EmpresaCRMCampanasHandler(dbEmp)))

	http.HandleFunc("/api/empresa/produccion/bom", WithEmpresaInventarioPermissions(dbEmp, dbSuper, empresaModuloGenericCRUDHandler(dbEmp, cfgProduccionBOM)))
	http.HandleFunc("/api/empresa/produccion/bom_detalle", WithEmpresaInventarioPermissions(dbEmp, dbSuper, empresaModuloGenericCRUDHandler(dbEmp, cfgProduccionBOMDetalle)))
	http.HandleFunc("/api/empresa/produccion/ordenes", WithEmpresaInventarioPermissions(dbEmp, dbSuper, empresaModuloGenericCRUDHandler(dbEmp, cfgProduccionOrdenes)))

	http.HandleFunc("/api/empresa/logistica/transportistas", WithEmpresaInventarioPermissions(dbEmp, dbSuper, empresaModuloGenericCRUDHandler(dbEmp, cfgLogisticaTransportistas)))
	http.HandleFunc("/api/empresa/logistica/rutas", WithEmpresaInventarioPermissions(dbEmp, dbSuper, empresaModuloGenericCRUDHandler(dbEmp, cfgLogisticaRutas)))
	http.HandleFunc("/api/empresa/logistica/envios", WithEmpresaInventarioPermissions(dbEmp, dbSuper, empresaModuloGenericCRUDHandler(dbEmp, cfgLogisticaEnvios)))

	http.HandleFunc("/api/empresa/documentos/gestion", WithEmpresaSeguridadPermissions(dbEmp, dbSuper, empresaModuloGenericCRUDHandler(dbEmp, cfgDocumentosGestion)))
	http.HandleFunc("/api/empresa/documentos/firmas", WithEmpresaSeguridadPermissions(dbEmp, dbSuper, empresaModuloGenericCRUDHandler(dbEmp, cfgDocumentosFirmas)))

	http.HandleFunc("/api/empresa/integraciones/apis", WithEmpresaSeguridadPermissions(dbEmp, dbSuper, EmpresaIntegracionesAPIsHandler(dbEmp)))
	http.HandleFunc("/api/empresa/integraciones/bancos", WithEmpresaSeguridadPermissions(dbEmp, dbSuper, EmpresaIntegracionesBancosHandler(dbEmp)))

	http.HandleFunc("/api/empresa/facturacion_electronica/dian", WithEmpresaFacturacionPermissions(dbEmp, dbSuper, EmpresaDIANColombiaHandler(dbEmp)))
}

func EmpresaVentasCotizacionesHandler(dbEmp *sql.DB) http.HandlerFunc {
	return empresaModuloStateMachineCRUDHandler(dbEmp, cfgCotizacionesVenta, stateMachineCotizaciones)
}

func EmpresaVentasPedidosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return empresaModuloStateMachineCRUDHandler(dbEmp, cfgPedidosVenta, stateMachinePedidos)
}

func EmpresaVentasDevolucionesHandler(dbEmp *sql.DB) http.HandlerFunc {
	return empresaModuloStateMachineCRUDHandler(dbEmp, cfgDevolucionesVenta, stateMachineDevoluciones)
}

func EmpresaCRMLeadsHandler(dbEmp *sql.DB) http.HandlerFunc {
	return empresaModuloStateMachineCRUDHandler(dbEmp, cfgCRMLeads, stateMachineCRMLeads)
}

func EmpresaCRMInteraccionesHandler(dbEmp *sql.DB) http.HandlerFunc {
	return empresaModuloStateMachineCRUDHandler(dbEmp, cfgCRMInteracciones, stateMachineCRMInteracciones)
}

func EmpresaCRMCampanasHandler(dbEmp *sql.DB) http.HandlerFunc {
	return empresaModuloStateMachineCRUDHandler(dbEmp, cfgCRMCampanas, stateMachineCRMCampanas)
}

func EmpresaIntegracionesAPIsHandler(dbEmp *sql.DB) http.HandlerFunc {
	return empresaModuloIntegracionesCRUDHandler(dbEmp, cfgIntegracionesAPIs, integrationOpsAPIs)
}

func EmpresaIntegracionesBancosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return empresaModuloIntegracionesCRUDHandler(dbEmp, cfgIntegracionesBancos, integrationOpsBancos)
}

func empresaModuloGenericCRUDHandler(dbEmp *sql.DB, cfg empresaModuloGenericConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "detalle" {
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				id, err := parseInt64Query(r, "id")
				if err != nil || id <= 0 {
					http.Error(w, "id required", http.StatusBadRequest)
					return
				}
				item, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfg.Table, empresaID, id)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "registro no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, "No se pudo consultar registro", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, item)
				return
			}

			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			limit, err := parseIntQueryOptional(r, "limit")
			if err != nil {
				http.Error(w, "limit invalido", http.StatusBadRequest)
				return
			}
			offset, err := parseIntQueryOptional(r, "offset")
			if err != nil {
				http.Error(w, "offset invalido", http.StatusBadRequest)
				return
			}
			items, err := dbpkg.ListEmpresaGenericRows(dbEmp, cfg.Table, empresaID, dbpkg.EmpresaGenericListFilter{
				IncludeInactive: parseBoolQuery(r, "include_inactive"),
				Q:               strings.TrimSpace(r.URL.Query().Get("q")),
				Limit:           limit,
				Offset:          offset,
				SearchColumns:   cfg.SearchColumns,
			})
			if err != nil {
				http.Error(w, "No se pudo listar registros", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, items)
			return

		case http.MethodPost:
			payload, err := decodeGenericBodyMap(r)
			if err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			empresaID, err := resolveEmpresaIDFromPayloadOrRequest(r, payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			applyGenericDefaultValues(payload, cfg.DefaultValues)
			ensureGenericCode(payload, cfg.CodeColumn, cfg.CodePrefix)
			if hasAllowedColumn(cfg.AllowedColumns, "usuario_creador") {
				if isEmptyGenericValue(payload["usuario_creador"]) {
					payload["usuario_creador"] = adminEmailFromRequest(r)
				}
			}
			if hasAllowedColumn(cfg.AllowedColumns, "estado") && isEmptyGenericValue(payload["estado"]) {
				payload["estado"] = "activo"
			}

			if err := validateGenericRequiredCreate(payload, cfg.RequiredOnCreate); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			id, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfg.Table, empresaID, payload, cfg.AllowedColumns)
			if err != nil {
				http.Error(w, "No se pudo crear registro", http.StatusBadRequest)
				return
			}
			item, _ := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfg.Table, empresaID, id)
			writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id, "item": item})
			return

		case http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "activar" || action == "desactivar" {
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				id, err := parseInt64Query(r, "id")
				if err != nil || id <= 0 {
					http.Error(w, "id required", http.StatusBadRequest)
					return
				}
				estado := "activo"
				if action == "desactivar" {
					estado = "inactivo"
				}
				if err := dbpkg.SetEmpresaGenericRowEstado(dbEmp, cfg.Table, empresaID, id, estado); err != nil {
					http.Error(w, "No se pudo actualizar estado", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": estado})
				return
			}

			payload, err := decodeGenericBodyMap(r)
			if err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			empresaID, err := resolveEmpresaIDFromPayloadOrRequest(r, payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			id := resolveIDFromPayloadOrQuery(payload, r)
			if id <= 0 {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			if err := dbpkg.UpdateEmpresaGenericRow(dbEmp, cfg.Table, empresaID, id, payload, cfg.AllowedColumns); err != nil {
				http.Error(w, "No se pudo actualizar registro", http.StatusBadRequest)
				return
			}
			item, _ := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfg.Table, empresaID, id)
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "item": item})
			return

		case http.MethodDelete:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			id, err := parseInt64Query(r, "id")
			if err != nil || id <= 0 {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			if err := dbpkg.DeleteEmpresaGenericRow(dbEmp, cfg.Table, empresaID, id); err != nil {
				http.Error(w, "No se pudo eliminar registro", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": "inactivo"})
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

func empresaModuloStateMachineCRUDHandler(dbEmp *sql.DB, cfg empresaModuloGenericConfig, sm empresaModuloStateMachineConfig) http.HandlerFunc {
	base := empresaModuloGenericCRUDHandler(dbEmp, cfg)
	return func(w http.ResponseWriter, r *http.Request) {
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch action {
		case "estado":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleStateMachineEstadoAction(dbEmp, cfg, sm, w, r)
			return

		case "transiciones":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleStateMachineTransicionesAction(dbEmp, cfg, sm, w, r)
			return

		case "transicionar":
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleStateMachineTransicionarAction(dbEmp, cfg, sm, w, r)
			return
		}

		base.ServeHTTP(w, r)
	}
}

func empresaModuloIntegracionesCRUDHandler(dbEmp *sql.DB, cfg empresaModuloGenericConfig, ops empresaModuloIntegracionesOpsConfig) http.HandlerFunc {
	base := empresaModuloGenericCRUDHandler(dbEmp, cfg)
	return func(w http.ResponseWriter, r *http.Request) {
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch action {
		case "estado":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleIntegracionesEstadoAction(dbEmp, cfg, ops, w, r)
			return

		case "health_check":
			if r.Method != http.MethodGet && r.Method != http.MethodPost && r.Method != http.MethodPut {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleIntegracionesProbeAction(dbEmp, cfg, ops, w, r, false)
			return

		case "sync_manual":
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleIntegracionesProbeAction(dbEmp, cfg, ops, w, r, true)
			return
		}

		base.ServeHTTP(w, r)
	}
}

func handleStateMachineEstadoAction(dbEmp *sql.DB, cfg empresaModuloGenericConfig, sm empresaModuloStateMachineConfig, w http.ResponseWriter, r *http.Request) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := parseInt64QueryOptional(r, "id")
	if err != nil {
		http.Error(w, "id invalido", http.StatusBadRequest)
		return
	}
	limit, err := parseIntQueryOptional(r, "limit")
	if err != nil {
		http.Error(w, "limit invalido", http.StatusBadRequest)
		return
	}
	offset, err := parseIntQueryOptional(r, "offset")
	if err != nil {
		http.Error(w, "offset invalido", http.StatusBadRequest)
		return
	}
	rows, err := loadEmpresaRowsForAction(dbEmp, cfg, empresaID, id, parseBoolQuery(r, "include_inactive"), strings.TrimSpace(r.URL.Query().Get("q")), limit, offset)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "registro no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo consultar estado", http.StatusInternalServerError)
		return
	}

	items := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		items = append(items, buildStateMachineSummary(row, sm))
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"modulo":     sm.ModuleName,
		"items":      items,
	})
}

func handleStateMachineTransicionesAction(dbEmp *sql.DB, cfg empresaModuloGenericConfig, sm empresaModuloStateMachineConfig, w http.ResponseWriter, r *http.Request) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := parseInt64QueryOptional(r, "id")
	if err != nil {
		http.Error(w, "id invalido", http.StatusBadRequest)
		return
	}

	if id > 0 {
		row, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfg.Table, empresaID, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "registro no encontrado", http.StatusNotFound)
				return
			}
			http.Error(w, "No se pudo consultar registro", http.StatusInternalServerError)
			return
		}
		current := resolveStateMachineRowState(row, sm)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":                       true,
			"empresa_id":               empresaID,
			"modulo":                   sm.ModuleName,
			"id":                       id,
			"estado_actual":            current,
			"transiciones_disponibles": allowedStateMachineTransitions(sm, current),
		})
		return
	}

	state := normalizeStateMachineValue(r.URL.Query().Get("estado"))
	if state == "" {
		http.Error(w, "id o estado required", http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                       true,
		"empresa_id":               empresaID,
		"modulo":                   sm.ModuleName,
		"estado_actual":            state,
		"transiciones_disponibles": allowedStateMachineTransitions(sm, state),
	})
}

func handleStateMachineTransicionarAction(dbEmp *sql.DB, cfg empresaModuloGenericConfig, sm empresaModuloStateMachineConfig, w http.ResponseWriter, r *http.Request) {
	payload, err := decodeGenericBodyMapOptional(r)
	if err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	empresaID, err := resolveEmpresaIDFromPayloadOrRequest(r, payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := resolveIDFromPayloadOrQuery(payload, r)
	if id <= 0 {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	targetState := resolveStateMachineTarget(payload, r, sm.StateColumn)
	if targetState == "" {
		http.Error(w, "nuevo_estado required", http.StatusBadRequest)
		return
	}

	row, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfg.Table, empresaID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "registro no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo consultar registro", http.StatusInternalServerError)
		return
	}

	currentState := resolveStateMachineRowState(row, sm)
	if currentState == targetState {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":                       true,
			"changed":                  false,
			"empresa_id":               empresaID,
			"id":                       id,
			"estado_actual":            currentState,
			"transiciones_disponibles": allowedStateMachineTransitions(sm, currentState),
		})
		return
	}

	allowed := allowedStateMachineTransitions(sm, currentState)
	if !containsStateMachineValue(allowed, targetState) {
		writeJSON(w, http.StatusConflict, map[string]interface{}{
			"ok":                       false,
			"empresa_id":               empresaID,
			"id":                       id,
			"modulo":                   sm.ModuleName,
			"estado_actual":            currentState,
			"estado_solicitado":        targetState,
			"transiciones_disponibles": allowed,
			"error":                    "transicion no permitida",
		})
		return
	}

	updatePayload := map[string]interface{}{sm.StateColumn: targetState}
	if hasAllowedColumn(cfg.AllowedColumns, "observaciones") {
		motivo := genericStringValue(payload["motivo"])
		updatePayload["observaciones"] = appendStateMachineObservation(genericStringValue(row["observaciones"]), currentState, targetState, motivo, adminEmailFromRequest(r))
	}

	if err := dbpkg.UpdateEmpresaGenericRow(dbEmp, cfg.Table, empresaID, id, updatePayload, cfg.AllowedColumns); err != nil {
		http.Error(w, "No se pudo transicionar estado", http.StatusBadRequest)
		return
	}

	updated, _ := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfg.Table, empresaID, id)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                       true,
		"changed":                  true,
		"empresa_id":               empresaID,
		"id":                       id,
		"modulo":                   sm.ModuleName,
		"estado_anterior":          currentState,
		"estado_nuevo":             targetState,
		"transiciones_disponibles": allowedStateMachineTransitions(sm, targetState),
		"item":                     updated,
	})
}

func handleIntegracionesEstadoAction(dbEmp *sql.DB, cfg empresaModuloGenericConfig, ops empresaModuloIntegracionesOpsConfig, w http.ResponseWriter, r *http.Request) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := parseInt64QueryOptional(r, "id")
	if err != nil {
		http.Error(w, "id invalido", http.StatusBadRequest)
		return
	}
	limit, err := parseIntQueryOptional(r, "limit")
	if err != nil {
		http.Error(w, "limit invalido", http.StatusBadRequest)
		return
	}
	offset, err := parseIntQueryOptional(r, "offset")
	if err != nil {
		http.Error(w, "offset invalido", http.StatusBadRequest)
		return
	}
	rows, err := loadEmpresaRowsForAction(dbEmp, cfg, empresaID, id, parseBoolQuery(r, "include_inactive"), strings.TrimSpace(r.URL.Query().Get("q")), limit, offset)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "registro no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo consultar estado de integracion", http.StatusInternalServerError)
		return
	}

	items := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		endpoint := normalizeIntegracionEndpoint(genericStringValue(row[ops.EndpointField]))
		estado := normalizeStateMachineValue(genericStringDefault(row["estado_integracion"], "inactiva"))
		if endpoint == "" {
			estado = "inactiva"
		}
		item := map[string]interface{}{
			"id":                 anyToInt64(row["id"]),
			"codigo":             genericStringValue(row["codigo"]),
			"nombre":             genericStringValue(row[ops.NameField]),
			"endpoint":           endpoint,
			"estado_integracion": estado,
			"estado_registro":    genericStringDefault(row["estado"], "activo"),
			"ultima_ejecucion":   row[ops.LastSyncField],
		}
		if strings.TrimSpace(ops.ResponseField) != "" {
			item["respuesta_ultimo_sync"] = row[ops.ResponseField]
		}
		items = append(items, item)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"modulo":     ops.ModuleName,
		"items":      items,
	})
}

func handleIntegracionesProbeAction(dbEmp *sql.DB, cfg empresaModuloGenericConfig, ops empresaModuloIntegracionesOpsConfig, w http.ResponseWriter, r *http.Request, syncManual bool) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := parseInt64QueryOptional(r, "id")
	if err != nil {
		http.Error(w, "id invalido", http.StatusBadRequest)
		return
	}
	limit, err := parseIntQueryOptional(r, "limit")
	if err != nil {
		http.Error(w, "limit invalido", http.StatusBadRequest)
		return
	}
	offset, err := parseIntQueryOptional(r, "offset")
	if err != nil {
		http.Error(w, "offset invalido", http.StatusBadRequest)
		return
	}
	rows, err := loadEmpresaRowsForAction(dbEmp, cfg, empresaID, id, parseBoolQuery(r, "include_inactive"), strings.TrimSpace(r.URL.Query().Get("q")), limit, offset)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "registro no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo preparar verificacion", http.StatusInternalServerError)
		return
	}

	actionName := "health_check"
	if syncManual {
		actionName = "sync_manual"
	}

	executedAt := time.Now().Format("2006-01-02 15:04:05")
	results := make([]empresaIntegracionProbeResult, 0, len(rows))
	errorsList := make([]map[string]interface{}, 0)

	for _, row := range rows {
		probe := buildIntegracionProbeResult(row, ops)

		updatePayload := map[string]interface{}{
			"estado_integracion": probe.EstadoIntegracion,
		}
		if syncManual && strings.TrimSpace(ops.LastSyncField) != "" {
			updatePayload[ops.LastSyncField] = executedAt
		}
		if strings.TrimSpace(ops.ResponseField) != "" {
			snapshot, _ := json.Marshal(map[string]interface{}{
				"checked_at":         executedAt,
				"endpoint":           probe.Endpoint,
				"http_status":        probe.HTTPStatus,
				"reachable":          probe.Reachable,
				"latency_ms":         probe.LatencyMS,
				"estado_integracion": probe.EstadoIntegracion,
				"message":            probe.Message,
				"action":             actionName,
			})
			updatePayload[ops.ResponseField] = string(snapshot)
		}

		if probe.ID > 0 {
			if err := dbpkg.UpdateEmpresaGenericRow(dbEmp, cfg.Table, empresaID, probe.ID, updatePayload, cfg.AllowedColumns); err != nil {
				errorsList = append(errorsList, map[string]interface{}{
					"id":    probe.ID,
					"error": err.Error(),
				})
				probe.Updated = false
			} else {
				probe.Updated = true
			}
		}

		results = append(results, probe)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":           len(errorsList) == 0,
		"empresa_id":   empresaID,
		"modulo":       ops.ModuleName,
		"accion":       actionName,
		"ejecutado_en": executedAt,
		"resultados":   results,
		"errores":      errorsList,
	})
}

func loadEmpresaRowsForAction(dbEmp *sql.DB, cfg empresaModuloGenericConfig, empresaID, id int64, includeInactive bool, q string, limit, offset int) ([]map[string]interface{}, error) {
	if id > 0 {
		item, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfg.Table, empresaID, id)
		if err != nil {
			return nil, err
		}
		return []map[string]interface{}{item}, nil
	}

	return dbpkg.ListEmpresaGenericRows(dbEmp, cfg.Table, empresaID, dbpkg.EmpresaGenericListFilter{
		IncludeInactive: includeInactive,
		Q:               q,
		Limit:           limit,
		Offset:          offset,
		SearchColumns:   cfg.SearchColumns,
	})
}

func normalizeStateMachineValue(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func resolveStateMachineRowState(row map[string]interface{}, sm empresaModuloStateMachineConfig) string {
	current := normalizeStateMachineValue(genericStringValue(row[sm.StateColumn]))
	if current == "" {
		current = normalizeStateMachineValue(sm.InitialState)
	}
	return current
}

func allowedStateMachineTransitions(sm empresaModuloStateMachineConfig, current string) []string {
	key := normalizeStateMachineValue(current)
	values, ok := sm.Transitions[key]
	if !ok || len(values) == 0 {
		return []string{}
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}

func containsStateMachineValue(values []string, target string) bool {
	target = normalizeStateMachineValue(target)
	for _, value := range values {
		if normalizeStateMachineValue(value) == target {
			return true
		}
	}
	return false
}

func buildStateMachineSummary(row map[string]interface{}, sm empresaModuloStateMachineConfig) map[string]interface{} {
	current := resolveStateMachineRowState(row, sm)
	return map[string]interface{}{
		"id":                       anyToInt64(row["id"]),
		"codigo":                   genericStringValue(row["codigo"]),
		"estado_actual":            current,
		"transiciones_disponibles": allowedStateMachineTransitions(sm, current),
	}
}

func resolveStateMachineTarget(payload map[string]interface{}, r *http.Request, stateColumn string) string {
	candidates := []string{}
	if payload != nil {
		candidates = append(candidates,
			genericStringValue(payload["nuevo_estado"]),
			genericStringValue(payload["estado_destino"]),
			genericStringValue(payload[stateColumn]),
		)
	}
	candidates = append(candidates,
		strings.TrimSpace(r.URL.Query().Get("nuevo_estado")),
		strings.TrimSpace(r.URL.Query().Get("estado_destino")),
		strings.TrimSpace(r.URL.Query().Get("estado")),
	)
	for _, candidate := range candidates {
		normalized := normalizeStateMachineValue(candidate)
		if normalized != "" {
			return normalized
		}
	}
	return ""
}

func appendStateMachineObservation(previous, current, target, motivo, actor string) string {
	parts := []string{fmt.Sprintf("[%s] transicion %s -> %s", time.Now().Format("2006-01-02 15:04:05"), current, target)}
	if actor != "" {
		parts = append(parts, "por "+actor)
	}
	if motivo != "" {
		parts = append(parts, "motivo: "+motivo)
	}
	line := strings.Join(parts, " | ")
	if strings.TrimSpace(previous) == "" {
		return line
	}
	return strings.TrimSpace(previous) + "\n" + line
}

func decodeGenericBodyMapOptional(r *http.Request) (map[string]interface{}, error) {
	payload := map[string]interface{}{}
	if r.Body == nil {
		return payload, nil
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		if errors.Is(err, io.EOF) {
			return payload, nil
		}
		return nil, err
	}
	return payload, nil
}

func normalizeIntegracionEndpoint(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "//") {
		raw = "https:" + raw
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	parsed, err := url.Parse(raw)
	if err != nil || strings.TrimSpace(parsed.Host) == "" {
		return ""
	}
	if strings.TrimSpace(parsed.Scheme) == "" {
		parsed.Scheme = "https"
	}
	return parsed.String()
}

func runIntegracionProbe(endpoint string) (int, bool, int64, string) {
	endpoint = normalizeIntegracionEndpoint(endpoint)
	if endpoint == "" {
		return 0, false, 0, "endpoint no configurado o invalido"
	}

	client := &http.Client{Timeout: 8 * time.Second}
	methods := []string{http.MethodHead, http.MethodGet}
	startedAt := time.Now()
	lastErr := ""
	statusCode := 0

	for _, method := range methods {
		req, err := http.NewRequest(method, endpoint, nil)
		if err != nil {
			return 0, false, time.Since(startedAt).Milliseconds(), err.Error()
		}
		req.Header.Set("User-Agent", "powerfulcontrolsystem-integration-check/1.0")
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err.Error()
			continue
		}
		statusCode = resp.StatusCode
		_ = resp.Body.Close()
		return statusCode, true, time.Since(startedAt).Milliseconds(), method + " " + resp.Status
	}

	if lastErr == "" {
		lastErr = "sin respuesta del endpoint"
	}
	return statusCode, false, time.Since(startedAt).Milliseconds(), lastErr
}

func buildIntegracionProbeResult(row map[string]interface{}, ops empresaModuloIntegracionesOpsConfig) empresaIntegracionProbeResult {
	result := empresaIntegracionProbeResult{
		ID:     anyToInt64(row["id"]),
		Codigo: genericStringValue(row["codigo"]),
		Nombre: genericStringValue(row[ops.NameField]),
	}

	endpointRaw := genericStringValue(row[ops.EndpointField])
	normalizedEndpoint := normalizeIntegracionEndpoint(endpointRaw)
	result.Endpoint = normalizedEndpoint

	if normalizedEndpoint == "" {
		result.EstadoIntegracion = "inactiva"
		result.Reachable = false
		result.Message = "endpoint no configurado o invalido"
		return result
	}

	statusCode, reachable, latencyMS, message := runIntegracionProbe(normalizedEndpoint)
	result.HTTPStatus = statusCode
	result.Reachable = reachable
	result.LatencyMS = latencyMS
	result.Message = message
	if reachable {
		result.EstadoIntegracion = "activa"
	} else {
		result.EstadoIntegracion = "error"
	}
	return result
}

// EmpresaDIANColombiaHandler expone configuracion DIAN y utilidades operativas de validacion.
func EmpresaDIANColombiaHandler(dbEmp *sql.DB) http.HandlerFunc {
	base := empresaModuloGenericCRUDHandler(dbEmp, cfgDIAN)
	return func(w http.ResponseWriter, r *http.Request) {
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch action {
		case "checklist", "validar":
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			cfg, _ := getEmpresaDIANConfig(dbEmp, empresaID)
			missing := missingDIANFields(cfg)
			response := map[string]interface{}{
				"ok":                len(missing) == 0,
				"empresa_id":        empresaID,
				"faltantes":         missing,
				"pasos_minimos":     dianChecklistSteps(),
				"ambiente_sugerido": chooseDIANAmbiente(cfg),
			}
			if action == "validar" {
				response["recomendaciones"] = []string{
					"Validar que software_id/software_pin coincidan con la plataforma DIAN.",
					"Confirmar rango de numeracion vigente y consecutivo dentro del rango.",
					"Mantener certificado y claves fuera del codigo fuente (referencias seguras).",
				}
			}
			writeJSON(w, http.StatusOK, response)
			return

		case "generar_cufe_demo":
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			payload, err := decodeGenericBodyMap(r)
			if err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			empresaID, err := resolveEmpresaIDFromPayloadOrRequest(r, payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			cfg, _ := getEmpresaDIANConfig(dbEmp, empresaID)
			documento := genericStringValue(payload["documento_codigo"])
			if documento == "" {
				documento = "FV-" + time.Now().Format("20060102150405")
			}
			total := genericStringValue(payload["total"])
			if total == "" {
				total = "0"
			}
			fecha := genericStringValue(payload["fecha_emision"])
			if fecha == "" {
				fecha = time.Now().Format("2006-01-02T15:04:05-07:00")
			}
			nit := genericStringValue(cfg["nit"])
			softwareID := genericStringValue(cfg["software_id"])
			softwarePIN := genericStringValue(cfg["software_pin"])
			seed := nit + "|" + documento + "|" + fecha + "|" + total + "|" + softwareID + "|" + softwarePIN
			sum := sha256.Sum256([]byte(seed))
			cufe := strings.ToUpper(hex.EncodeToString(sum[:]))

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":               true,
				"empresa_id":       empresaID,
				"documento_codigo": documento,
				"fecha_emision":    fecha,
				"total":            total,
				"cufe_demo":        cufe,
				"algoritmo":        "SHA-256",
			})
			return

		case "generar_xml_demo":
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			payload, err := decodeGenericBodyMap(r)
			if err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			empresaID, err := resolveEmpresaIDFromPayloadOrRequest(r, payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			cfg, _ := getEmpresaDIANConfig(dbEmp, empresaID)
			doc := genericStringValue(payload["documento_codigo"])
			if doc == "" {
				doc = "FV-" + time.Now().Format("20060102150405")
			}
			nit := genericStringValue(cfg["nit"])
			if nit == "" {
				nit = "000000000"
			}
			razon := genericStringValue(cfg["razon_social"])
			if razon == "" {
				razon = "EMPRESA DEMO"
			}
			fecha := time.Now().Format("2006-01-02")
			xml := fmt.Sprintf("<?xml version=\"1.0\" encoding=\"UTF-8\"?><Invoice><ProfileExecutionID>%s</ProfileExecutionID><ID>%s</ID><IssueDate>%s</IssueDate><AccountingSupplierParty><PartyTaxScheme><CompanyID>%s</CompanyID><RegistrationName>%s</RegistrationName></PartyTaxScheme></AccountingSupplierParty></Invoice>",
				genericStringDefault(cfg["tipo_ambiente"], "habilitacion"), doc, fecha, nit, escapeXML(razon))

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":               true,
				"empresa_id":       empresaID,
				"documento_codigo": doc,
				"xml_demo":         xml,
			})
			return
		}

		base.ServeHTTP(w, r)
	}
}

func decodeGenericBodyMap(r *http.Request) (map[string]interface{}, error) {
	payload := map[string]interface{}{}
	if r.Body == nil {
		return payload, nil
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func resolveEmpresaIDFromPayloadOrRequest(r *http.Request, payload map[string]interface{}) (int64, error) {
	if payload != nil {
		if v, ok := payload["empresa_id"]; ok {
			id := anyToInt64(v)
			if id > 0 {
				return id, nil
			}
		}
	}
	id, err := parseEmpresaIDQuery(r)
	if err != nil {
		return 0, fmt.Errorf("empresa_id required")
	}
	if id <= 0 {
		return 0, fmt.Errorf("empresa_id required")
	}
	return id, nil
}

func resolveIDFromPayloadOrQuery(payload map[string]interface{}, r *http.Request) int64 {
	if payload != nil {
		if v, ok := payload["id"]; ok {
			if id := anyToInt64(v); id > 0 {
				return id
			}
		}
	}
	id, _ := parseInt64QueryOptional(r, "id")
	return id
}

func anyToInt64(v interface{}) int64 {
	switch value := v.(type) {
	case int:
		return int64(value)
	case int32:
		return int64(value)
	case int64:
		return value
	case float64:
		return int64(value)
	case json.Number:
		i, _ := value.Int64()
		return i
	case string:
		i, _ := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		return i
	default:
		return 0
	}
}

func genericStringValue(v interface{}) string {
	switch value := v.(type) {
	case string:
		return strings.TrimSpace(value)
	case []byte:
		return strings.TrimSpace(string(value))
	case nil:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", value))
	}
}

func genericStringDefault(v interface{}, fallback string) string {
	value := genericStringValue(v)
	if value == "" {
		return fallback
	}
	return value
}

func isEmptyGenericValue(v interface{}) bool {
	if v == nil {
		return true
	}
	s := genericStringValue(v)
	return s == ""
}

func applyGenericDefaultValues(payload map[string]interface{}, defaults map[string]interface{}) {
	for key, value := range defaults {
		if isEmptyGenericValue(payload[key]) {
			payload[key] = value
		}
	}
}

func ensureGenericCode(payload map[string]interface{}, codeColumn, codePrefix string) {
	if strings.TrimSpace(codeColumn) == "" {
		return
	}
	if !isEmptyGenericValue(payload[codeColumn]) {
		return
	}
	prefix := strings.ToUpper(strings.TrimSpace(codePrefix))
	if prefix == "" {
		prefix = "DOC"
	}
	payload[codeColumn] = prefix + "-" + time.Now().Format("20060102150405")
}

func validateGenericRequiredCreate(payload map[string]interface{}, required []string) error {
	for _, field := range required {
		if isEmptyGenericValue(payload[field]) {
			return fmt.Errorf("%s es obligatorio", field)
		}
	}
	return nil
}

func hasAllowedColumn(allowed []string, name string) bool {
	target := strings.ToLower(strings.TrimSpace(name))
	for _, col := range allowed {
		if strings.ToLower(strings.TrimSpace(col)) == target {
			return true
		}
	}
	return false
}

func getEmpresaDIANConfig(dbEmp *sql.DB, empresaID int64) (map[string]interface{}, error) {
	items, err := dbpkg.ListEmpresaGenericRows(dbEmp, cfgDIAN.Table, empresaID, dbpkg.EmpresaGenericListFilter{IncludeInactive: true, Limit: 1, Offset: 0, SearchColumns: cfgDIAN.SearchColumns})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return map[string]interface{}{}, nil
	}
	return items[0], nil
}

func missingDIANFields(cfg map[string]interface{}) []string {
	required := []string{
		"nit",
		"razon_social",
		"tipo_ambiente",
		"software_id",
		"software_pin",
		"prefijo",
		"resolucion_numero",
		"rango_desde",
		"rango_hasta",
	}
	missing := make([]string, 0)
	for _, field := range required {
		if isEmptyGenericValue(cfg[field]) {
			missing = append(missing, field)
		}
	}
	return missing
}

func chooseDIANAmbiente(cfg map[string]interface{}) string {
	ambiente := strings.ToLower(genericStringValue(cfg["tipo_ambiente"]))
	if ambiente == "produccion" {
		return "produccion"
	}
	return "habilitacion"
}

func dianChecklistSteps() []map[string]interface{} {
	return []map[string]interface{}{
		{"paso": 1, "titulo": "Habilitar facturador en portal DIAN", "detalle": "Registrar empresa como facturador electronico y elegir tipo de software."},
		{"paso": 2, "titulo": "Obtener Software ID y Software PIN", "detalle": "Desde DIAN tomar credenciales del software de facturacion."},
		{"paso": 3, "titulo": "Solicitar numeracion", "detalle": "Solicitar prefijo, resolucion y rango autorizado en la DIAN."},
		{"paso": 4, "titulo": "Cargar configuracion en el sistema", "detalle": "Configurar NIT, razon social, ambiente, software, resolucion y rangos."},
		{"paso": 5, "titulo": "Ejecutar set de pruebas", "detalle": "Enviar casos de habilitacion hasta obtener aprobacion DIAN."},
		{"paso": 6, "titulo": "Pasar a produccion", "detalle": "Activar ambiente produccion, validar consecutivos y monitorear respuestas."},
	}
}

func escapeXML(v string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&apos;",
	)
	return replacer.Replace(v)
}
