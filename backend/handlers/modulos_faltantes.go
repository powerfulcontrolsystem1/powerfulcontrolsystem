package handlers

import (
	"archive/zip"
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"golang.org/x/crypto/pkcs12"
)

var empresaModulosFaltantesSchemaState = struct {
	sync.Mutex
	ready map[*sql.DB]bool
}{
	ready: make(map[*sql.DB]bool),
}

func ensureEmpresaModulosFaltantesSchemaReady(dbEmp *sql.DB) error {
	if dbEmp == nil {
		return errors.New("db connection is nil")
	}

	empresaModulosFaltantesSchemaState.Lock()
	defer empresaModulosFaltantesSchemaState.Unlock()

	if empresaModulosFaltantesSchemaState.ready[dbEmp] {
		return nil
	}
	if err := dbpkg.EnsureEmpresaModulosFaltantesSchema(dbEmp); err != nil {
		return err
	}
	empresaModulosFaltantesSchemaState.ready[dbEmp] = true
	return nil
}

func withEmpresaModulosFaltantesSchema(dbEmp *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := ensureEmpresaModulosFaltantesSchemaReady(dbEmp); err != nil {
			http.Error(w, "No se pudo preparar el esquema de modulos ERP faltantes", http.StatusInternalServerError)
			return
		}
		next(w, r)
	}
}

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
	ModuleName      string
	EndpointField   string
	LastSyncField   string
	ResponseField   string
	NameField       string
	CredentialField string
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
			"aplica_impuesto", "plantilla_tipo_empresa", "plantilla_codigo", "plantilla_version", "cuenta_clave",
			"requerida", "orden_plantilla", "usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"codigo", "nombre"},
		DefaultValues: map[string]interface{}{
			"tipo_cuenta":            "activo",
			"naturaleza":             "debito",
			"nivel":                  1,
			"admite_movimiento":      1,
			"plantilla_tipo_empresa": "general",
			"plantilla_version":      "1",
			"requerida":              0,
			"orden_plantilla":        0,
		},
	}

	cfgCxC = empresaModuloGenericConfig{
		Table:         "empresa_cuentas_por_cobrar",
		SearchColumns: []string{"codigo", "cliente_nombre", "documento_codigo", "estado_cartera"},
		AllowedColumns: []string{
			"codigo", "cliente_id", "cliente_nombre", "documento_tipo", "documento_codigo", "fecha_emision",
			"fecha_vencimiento", "dias_mora", "valor_original", "valor_pagado", "saldo", "estado_cartera",
			"moneda", "periodo_contable", "referencia_pagos_json", "fecha_ultimo_pago", "conciliado_en", "conciliado_por",
			"usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"cliente_nombre", "documento_codigo"},
		CodeColumn:       "codigo",
		CodePrefix:       "CXC",
		DefaultValues: map[string]interface{}{
			"estado_cartera": "pendiente",
			"moneda":         "COP",
			"valor_pagado":   0,
		},
	}

	cfgCxP = empresaModuloGenericConfig{
		Table:         "empresa_cuentas_por_pagar",
		SearchColumns: []string{"codigo", "proveedor_nombre", "documento_codigo", "estado_cartera"},
		AllowedColumns: []string{
			"codigo", "proveedor_id", "proveedor_nombre", "documento_tipo", "documento_codigo", "fecha_emision",
			"fecha_vencimiento", "dias_mora", "valor_original", "valor_pagado", "saldo", "estado_cartera",
			"moneda", "periodo_contable", "referencia_pagos_json", "fecha_ultimo_pago", "conciliado_en", "conciliado_por",
			"usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"proveedor_nombre", "documento_codigo"},
		CodeColumn:       "codigo",
		CodePrefix:       "CXP",
		DefaultValues: map[string]interface{}{
			"estado_cartera": "pendiente",
			"moneda":         "COP",
			"valor_pagado":   0,
		},
	}

	cfgLotesSeries = empresaModuloGenericConfig{
		Table:         "inventario_lotes_series",
		SearchColumns: []string{"codigo_lote_serie", "tipo_control", "estado_lote"},
		AllowedColumns: []string{
			"producto_id", "bodega_id", "tipo_control", "codigo_lote_serie", "fecha_fabricacion", "fecha_vencimiento",
			"cantidad_inicial", "cantidad_disponible", "reservado_cantidad", "vendido_cantidad", "costo_unitario",
			"estado_lote", "bloqueado_venta", "bloqueo_motivo", "ultima_operacion_tipo", "ultima_operacion_ref", "ultima_operacion_en",
			"usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"producto_id", "codigo_lote_serie"},
		DefaultValues: map[string]interface{}{
			"tipo_control":       "lote",
			"estado_lote":        "activo",
			"bloqueado_venta":    0,
			"reservado_cantidad": 0,
			"vendido_cantidad":   0,
		},
	}

	cfgDevProveedor = empresaModuloGenericConfig{
		Table:         "empresa_devoluciones_proveedor",
		SearchColumns: []string{"codigo", "proveedor_nombre", "documento_compra_codigo", "estado_devolucion"},
		AllowedColumns: []string{
			"codigo", "proveedor_id", "proveedor_nombre", "documento_compra_codigo", "fecha_devolucion",
			"motivo", "estado_devolucion", "subtotal", "impuesto_total", "total", "moneda", "periodo_contable",
			"impacto_contable_movimiento_id", "impacto_contable_evento_id", "fecha_contabilizacion", "contabilizado_por", "total_reintegrado",
			"usuario_creador", "estado", "observaciones",
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
			"codigo", "empleado_id", "empleado_nomina_id", "empleado_nombre", "tipo_novedad", "fecha_inicio", "fecha_fin", "dias",
			"remunerada", "estado_novedad", "soporte_url", "aprobado_por", "nivel_aprobacion_actual", "nivel_aprobacion_requerido",
			"aprobadores_json", "historial_aprobaciones_json", "fecha_aprobacion_final", "periodo_acumulado_desde", "periodo_acumulado_hasta",
			"saldo_dias_antes", "saldo_dias_despues", "saldo_snapshot_json", "nomina_liquidacion_id", "nomina_periodo_desde",
			"nomina_periodo_hasta", "nomina_vinculada_en", "nomina_vinculada_por", "usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"empleado_nombre", "tipo_novedad", "fecha_inicio", "fecha_fin"},
		CodeColumn:       "codigo",
		CodePrefix:       "RRHH",
		DefaultValues: map[string]interface{}{
			"tipo_novedad":               "vacacion",
			"estado_novedad":             "solicitada",
			"remunerada":                 1,
			"nivel_aprobacion_actual":    0,
			"nivel_aprobacion_requerido": 1,
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
			"usar_software_compartido", "software_id_compartido_ref", "software_pin_compartido_ref",
			"modo_operacion_descripcion", "modo_operacion_fecha_inicio", "modo_operacion_fecha_termino",
			"test_set_id", "certificado_url", "certificado_clave_ref", "prefijo", "resolucion_numero",
			"certificado_vencimiento", "certificado_vencimiento_en", "certificado_alerta_dias",
			"certificado_alerta_ultimo_envio", "certificado_alerta_email",
			"certificado_ultima_carga_en", "certificado_archivo_original", "certificado_formato",
			"certificado_subject", "certificado_issuer", "certificado_serial", "certificado_clave_estado",
			"resolucion_fecha_desde", "resolucion_fecha_hasta", "rango_desde", "rango_hasta", "consecutivo_actual",
			"llave_tecnica", "set_documentos_requeridos", "set_facturas_requeridas", "set_notas_debito_requeridas",
			"set_notas_credito_requeridas", "set_documentos_aceptados_requeridos", "set_facturas_aceptadas_requeridas",
			"set_notas_debito_aceptadas_requeridas", "set_notas_credito_aceptadas_requeridas",
			"url_dian", "token_emisor_ref", "ultimo_envio", "estado_dian", "usuario_creador", "estado", "observaciones",
		},
		RequiredOnCreate: []string{"nit", "razon_social", "tipo_ambiente"},
		CodeColumn:       "codigo",
		CodePrefix:       "DIAN",
		DefaultValues: map[string]interface{}{
			"tipo_ambiente":            "habilitacion",
			"estado_dian":              "pendiente",
			"url_dian":                 "https://vpfe-hab.dian.gov.co/WcfDianCustomerServices.svc?wsdl",
			"usar_software_compartido": 0,
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
		ModuleName:      "integraciones_apis",
		EndpointField:   "base_url",
		LastSyncField:   "ultima_sincronizacion",
		ResponseField:   "respuesta_ultimo_sync",
		NameField:       "nombre_integracion",
		CredentialField: "api_key_ref",
	}

	integrationOpsBancos = empresaModuloIntegracionesOpsConfig{
		ModuleName:      "integraciones_bancos",
		EndpointField:   "api_endpoint",
		LastSyncField:   "ultima_conciliacion",
		ResponseField:   "",
		NameField:       "banco_nombre",
		CredentialField: "credencial_ref",
	}
)

// RegisterEmpresaModulosFaltantesRoutes registra endpoints para modulos faltantes ERP/POS.
func RegisterEmpresaModulosFaltantesRoutes(dbEmp, dbSuper *sql.DB) {
	http.HandleFunc("/api/empresa/ventas/cotizaciones", WithEmpresaVentasPermissions(dbEmp, dbSuper, withEmpresaModulosFaltantesSchema(dbEmp, EmpresaVentasCotizacionesHandler(dbEmp))))
	http.HandleFunc("/api/empresa/ventas/pedidos", WithEmpresaVentasPermissions(dbEmp, dbSuper, withEmpresaModulosFaltantesSchema(dbEmp, EmpresaVentasPedidosHandler(dbEmp))))
	http.HandleFunc("/api/empresa/ventas/devoluciones", WithEmpresaVentasPermissions(dbEmp, dbSuper, withEmpresaModulosFaltantesSchema(dbEmp, EmpresaVentasDevolucionesHandler(dbEmp))))

	http.HandleFunc("/api/empresa/finanzas/plan_cuentas", WithEmpresaFinanzasPermissions(dbEmp, dbSuper, withEmpresaModulosFaltantesSchema(dbEmp, EmpresaFinanzasPlanCuentasHandler(dbEmp))))
	http.HandleFunc("/api/empresa/finanzas/cuentas_cobrar", WithEmpresaFinanzasPermissions(dbEmp, dbSuper, withEmpresaModulosFaltantesSchema(dbEmp, EmpresaFinanzasCuentasCobrarHandler(dbEmp))))
	http.HandleFunc("/api/empresa/finanzas/cuentas_pagar", WithEmpresaFinanzasPermissions(dbEmp, dbSuper, withEmpresaModulosFaltantesSchema(dbEmp, EmpresaFinanzasCuentasPagarHandler(dbEmp))))

	http.HandleFunc("/api/empresa/inventario/lotes_series", WithEmpresaInventarioPermissions(dbEmp, dbSuper, withEmpresaModulosFaltantesSchema(dbEmp, EmpresaInventarioLotesSeriesHandler(dbEmp))))
	http.HandleFunc("/api/empresa/compras/devoluciones_proveedor", WithEmpresaComprasPermissions(dbEmp, dbSuper, withEmpresaModulosFaltantesSchema(dbEmp, EmpresaComprasDevolucionesProveedorHandler(dbEmp))))
	http.HandleFunc("/api/empresa/rrhh/vacaciones_licencias", WithEmpresaSeguridadPermissions(dbEmp, dbSuper, withEmpresaModulosFaltantesSchema(dbEmp, EmpresaRRHHVacacionesLicenciasHandler(dbEmp))))

	http.HandleFunc("/api/empresa/crm/leads", WithEmpresaCRMUnificadoPermissions(dbEmp, dbSuper, withEmpresaModulosFaltantesSchema(dbEmp, EmpresaCRMLeadsHandler(dbEmp))))
	http.HandleFunc("/api/empresa/crm/interacciones", WithEmpresaCRMUnificadoPermissions(dbEmp, dbSuper, withEmpresaModulosFaltantesSchema(dbEmp, EmpresaCRMInteraccionesHandler(dbEmp))))
	http.HandleFunc("/api/empresa/crm/campanas", WithEmpresaCRMUnificadoPermissions(dbEmp, dbSuper, withEmpresaModulosFaltantesSchema(dbEmp, EmpresaCRMCampanasHandler(dbEmp))))

	http.HandleFunc("/api/empresa/produccion/bom", WithEmpresaInventarioPermissions(dbEmp, dbSuper, withEmpresaModulosFaltantesSchema(dbEmp, empresaModuloGenericCRUDHandler(dbEmp, cfgProduccionBOM))))
	http.HandleFunc("/api/empresa/produccion/bom_detalle", WithEmpresaInventarioPermissions(dbEmp, dbSuper, withEmpresaModulosFaltantesSchema(dbEmp, empresaModuloGenericCRUDHandler(dbEmp, cfgProduccionBOMDetalle))))
	http.HandleFunc("/api/empresa/produccion/ordenes", WithEmpresaInventarioPermissions(dbEmp, dbSuper, withEmpresaModulosFaltantesSchema(dbEmp, EmpresaProduccionOrdenesHandler(dbEmp))))

	http.HandleFunc("/api/empresa/logistica/transportistas", WithEmpresaInventarioPermissions(dbEmp, dbSuper, withEmpresaModulosFaltantesSchema(dbEmp, empresaModuloGenericCRUDHandler(dbEmp, cfgLogisticaTransportistas))))
	http.HandleFunc("/api/empresa/logistica/rutas", WithEmpresaInventarioPermissions(dbEmp, dbSuper, withEmpresaModulosFaltantesSchema(dbEmp, empresaModuloGenericCRUDHandler(dbEmp, cfgLogisticaRutas))))
	http.HandleFunc("/api/empresa/logistica/envios", WithEmpresaInventarioPermissions(dbEmp, dbSuper, withEmpresaModulosFaltantesSchema(dbEmp, EmpresaLogisticaEnviosHandler(dbEmp))))

	http.HandleFunc("/api/empresa/documentos/gestion", WithEmpresaSeguridadPermissions(dbEmp, dbSuper, withEmpresaModulosFaltantesSchema(dbEmp, EmpresaDocumentosGestionHandler(dbEmp))))
	http.HandleFunc("/api/empresa/documentos/firmas", WithEmpresaSeguridadPermissions(dbEmp, dbSuper, withEmpresaModulosFaltantesSchema(dbEmp, EmpresaDocumentosFirmasHandler(dbEmp))))

	http.HandleFunc("/api/empresa/integraciones/apis", WithEmpresaSeguridadPermissions(dbEmp, dbSuper, withEmpresaModulosFaltantesSchema(dbEmp, EmpresaIntegracionesAPIsHandler(dbEmp))))
	http.HandleFunc("/api/empresa/integraciones/bancos", WithEmpresaSeguridadPermissions(dbEmp, dbSuper, withEmpresaModulosFaltantesSchema(dbEmp, EmpresaIntegracionesBancosHandler(dbEmp))))

	http.HandleFunc("/api/empresa/facturacion_electronica/dian", WithEmpresaFacturacionPermissions(dbEmp, dbSuper, withEmpresaModulosFaltantesSchema(dbEmp, EmpresaDIANColombiaHandler(dbEmp, dbSuper))))
}

func EmpresaVentasCotizacionesHandler(dbEmp *sql.DB) http.HandlerFunc {
	base := empresaModuloStateMachineCRUDHandler(dbEmp, cfgCotizacionesVenta, stateMachineCotizaciones)
	return func(w http.ResponseWriter, r *http.Request) {
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch action {
		case "convertir_pedido":
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleVentasCotizacionConvertirPedidoAction(dbEmp, w, r)
			return

		case "convertir_documento_final":
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleVentasCotizacionConvertirDocumentoFinalAction(dbEmp, w, r)
			return

		case "embudo", "reporte_conversion":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleVentasEmbudoConversionAction(dbEmp, w, r)
			return
		}

		base.ServeHTTP(w, r)
	}
}

func EmpresaVentasPedidosHandler(dbEmp *sql.DB) http.HandlerFunc {
	base := empresaModuloStateMachineCRUDHandler(dbEmp, cfgPedidosVenta, stateMachinePedidos)
	return func(w http.ResponseWriter, r *http.Request) {
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "convertir_documento_final" {
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleVentasPedidoConvertirDocumentoFinalAction(dbEmp, w, r)
			return
		}

		base.ServeHTTP(w, r)
	}
}

func EmpresaVentasDevolucionesHandler(dbEmp *sql.DB) http.HandlerFunc {
	return empresaModuloStateMachineCRUDHandler(dbEmp, cfgDevolucionesVenta, stateMachineDevoluciones)
}

type planCuentaTemplateItem struct {
	Codigo           string
	Nombre           string
	TipoCuenta       string
	Naturaleza       string
	Nivel            int64
	CuentaPadre      string
	AdmiteMovimiento bool
	AplicaImpuesto   bool
	CuentaClave      string
	Requerida        bool
	Orden            int64
	Observaciones    string
}

type carteraPagoRelacionado struct {
	MovimientoID      int64   `json:"movimiento_id"`
	Codigo            string  `json:"codigo,omitempty"`
	FechaMovimiento   string  `json:"fecha_movimiento,omitempty"`
	MontoAplicado     float64 `json:"monto_aplicado"`
	ReferenciaExterna string  `json:"referencia_externa,omitempty"`
	NumeroComprobante string  `json:"numero_comprobante,omitempty"`
}

func dedupePagosCarteraRelacionados(items []carteraPagoRelacionado) []carteraPagoRelacionado {
	if len(items) <= 1 {
		return items
	}
	seen := make(map[string]struct{}, len(items))
	out := make([]carteraPagoRelacionado, 0, len(items))
	for _, item := range items {
		key := ""
		if item.MovimientoID > 0 {
			key = "mov:" + strconv.FormatInt(item.MovimientoID, 10)
		} else {
			key = strings.TrimSpace(item.Codigo) + "|" +
				strings.TrimSpace(item.FechaMovimiento) + "|" +
				strconv.FormatFloat(reportesRound(item.MontoAplicado), 'f', 2, 64) + "|" +
				strings.TrimSpace(item.ReferenciaExterna) + "|" +
				strings.TrimSpace(item.NumeroComprobante)
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func EmpresaFinanzasPlanCuentasHandler(dbEmp *sql.DB) http.HandlerFunc {
	base := empresaModuloGenericCRUDHandler(dbEmp, cfgPlanCuentas)
	return func(w http.ResponseWriter, r *http.Request) {
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch action {
		case "plantillas":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handlePlanCuentasPlantillasAction(w, r)
			return

		case "aplicar_plantilla":
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handlePlanCuentasAplicarPlantillaAction(dbEmp, w, r)
			return

		case "validar_cierre_periodo":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleFinanzasValidarCierrePeriodoAction(dbEmp, w, r)
			return
		}

		base.ServeHTTP(w, r)
	}
}

func EmpresaFinanzasCuentasCobrarHandler(dbEmp *sql.DB) http.HandlerFunc {
	return empresaFinanzasCarteraHandler(dbEmp, cfgCxC, "ingreso", "cliente_nombre", "cuentas_cobrar")
}

func EmpresaFinanzasCuentasPagarHandler(dbEmp *sql.DB) http.HandlerFunc {
	return empresaFinanzasCarteraHandler(dbEmp, cfgCxP, "egreso", "proveedor_nombre", "cuentas_pagar")
}

func EmpresaInventarioLotesSeriesHandler(dbEmp *sql.DB) http.HandlerFunc {
	base := empresaModuloGenericCRUDHandler(dbEmp, cfgLotesSeries)
	return func(w http.ResponseWriter, r *http.Request) {
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch action {
		case "trazabilidad":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleInventarioLotesSeriesTrazabilidadAction(dbEmp, w, r)
			return

		case "validar_disponibilidad":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleInventarioLotesSeriesValidarDisponibilidadAction(dbEmp, w, r)
			return

		case "operar", "reservar", "vender", "liberar_reserva", "ajuste_entrada", "ajuste_salida", "devolucion_proveedor":
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleInventarioLotesSeriesOperacionAction(dbEmp, action, w, r)
			return
		}

		base.ServeHTTP(w, r)
	}
}

func EmpresaComprasDevolucionesProveedorHandler(dbEmp *sql.DB) http.HandlerFunc {
	base := empresaModuloGenericCRUDHandler(dbEmp, cfgDevProveedor)
	return func(w http.ResponseWriter, r *http.Request) {
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch action {
		case "contabilizar", "impacto_contable":
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleComprasDevolucionProveedorContabilizarAction(dbEmp, w, r)
			return
		}

		base.ServeHTTP(w, r)
	}
}

type rrhhVacacionesSaldoFilter struct {
	EmpleadoID       int64
	EmpleadoNominaID int64
	EmpleadoNombre   string
	ExcluirNovedadID int64
	FechaCorte       time.Time
}

type rrhhNominaEmpleadoRef struct {
	NominaID     int64
	EmpleadoID   int64
	Nombre       string
	FechaIngreso string
}

func EmpresaRRHHVacacionesLicenciasHandler(dbEmp *sql.DB) http.HandlerFunc {
	base := empresaModuloGenericCRUDHandler(dbEmp, cfgRRHHVacLic)
	return func(w http.ResponseWriter, r *http.Request) {
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch action {
		case "saldo", "saldo_vacaciones", "resumen_saldo":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleRRHHVacacionesSaldoAction(dbEmp, w, r)
			return

		case "solicitar_aprobacion", "iniciar_aprobacion", "aprobar":
			if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleRRHHVacacionesAprobacionAction(dbEmp, action, w, r)
			return

		case "rechazar":
			if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleRRHHVacacionesRechazoAction(dbEmp, w, r)
			return

		case "vincular_nomina", "enlazar_nomina":
			if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleRRHHVacacionesVincularNominaAction(dbEmp, w, r)
			return
		}

		base.ServeHTTP(w, r)
	}
}

func handleRRHHVacacionesSaldoAction(dbEmp *sql.DB, w http.ResponseWriter, r *http.Request) {
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

	fechaCorte := time.Now().In(time.Local)
	if raw := strings.TrimSpace(r.URL.Query().Get("fecha_corte")); raw != "" {
		parsed, ok := ventasParseDateTime(raw)
		if !ok {
			http.Error(w, "fecha_corte invalida (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		fechaCorte = parsed.In(time.Local)
	}

	filtro := rrhhVacacionesSaldoFilter{
		EmpleadoNominaID: anyToInt64(r.URL.Query().Get("empleado_nomina_id")),
		EmpleadoID:       anyToInt64(r.URL.Query().Get("empleado_id")),
		EmpleadoNombre:   strings.TrimSpace(r.URL.Query().Get("empleado_nombre")),
		FechaCorte:       fechaCorte,
	}

	if id > 0 {
		item, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgRRHHVacLic.Table, empresaID, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "novedad RRHH no encontrada", http.StatusNotFound)
				return
			}
			http.Error(w, "no se pudo consultar novedad RRHH", http.StatusInternalServerError)
			return
		}
		if filtro.EmpleadoNominaID <= 0 {
			filtro.EmpleadoNominaID = anyToInt64(item["empleado_nomina_id"])
		}
		if filtro.EmpleadoID <= 0 {
			filtro.EmpleadoID = anyToInt64(item["empleado_id"])
		}
		if strings.TrimSpace(filtro.EmpleadoNombre) == "" {
			filtro.EmpleadoNombre = genericStringValue(item["empleado_nombre"])
		}
	}

	if filtro.EmpleadoNominaID <= 0 && filtro.EmpleadoID <= 0 && strings.TrimSpace(filtro.EmpleadoNombre) == "" {
		http.Error(w, "id, empleado_nomina_id, empleado_id o empleado_nombre es obligatorio", http.StatusBadRequest)
		return
	}

	saldo, err := buildRRHHVacacionesSaldoByFilter(dbEmp, empresaID, filtro)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "empleado de nomina no encontrado para calcular saldo", http.StatusNotFound)
			return
		}
		http.Error(w, "no se pudo calcular saldo de vacaciones", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"saldo":      saldo,
	})
}

func handleRRHHVacacionesAprobacionAction(dbEmp *sql.DB, action string, w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "id es obligatorio", http.StatusBadRequest)
		return
	}

	item, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgRRHHVacLic.Table, empresaID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "novedad RRHH no encontrada", http.StatusNotFound)
			return
		}
		http.Error(w, "no se pudo consultar novedad RRHH", http.StatusInternalServerError)
		return
	}

	actor := strings.TrimSpace(adminEmailFromRequest(r))
	if actor == "" {
		actor = "sistema"
	}
	now := time.Now().In(time.Local)
	nowText := now.Format("2006-01-02 15:04:05")
	estadoActual := strings.ToLower(strings.TrimSpace(genericStringValue(item["estado_novedad"])))
	nivelActual := anyToInt64(item["nivel_aprobacion_actual"])
	nivelRequerido := anyToInt64(payload["nivel_aprobacion_requerido"])
	if nivelRequerido <= 0 {
		nivelRequerido = anyToInt64(item["nivel_aprobacion_requerido"])
	}
	if nivelRequerido <= 0 {
		nivelRequerido = 1
	}

	if nivelActual < 0 {
		nivelActual = 0
	}

	update := map[string]interface{}{
		"nivel_aprobacion_requerido": nivelRequerido,
	}
	comentario := strings.TrimSpace(finanzasFirstNonBlank(genericStringValue(payload["comentario"]), genericStringValue(payload["motivo"])))
	historial := parseJSONArrayObjects(genericStringValue(item["historial_aprobaciones_json"]))
	aprobadores := parseJSONArrayStrings(genericStringValue(item["aprobadores_json"]))

	action = strings.ToLower(strings.TrimSpace(action))
	switch action {
	case "solicitar_aprobacion", "iniciar_aprobacion":
		if estadoActual == "aprobada" || estadoActual == "contabilizada" || estadoActual == "cerrada" {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":                  true,
				"empresa_id":          empresaID,
				"id":                  id,
				"estado_novedad":      estadoActual,
				"mensaje":             "novedad ya aprobada o cerrada",
				"nivel_actual":        nivelActual,
				"nivel_requerido":     nivelRequerido,
				"aprobadores_totales": len(aprobadores),
			})
			return
		}
		if estadoActual == "rechazada" {
			http.Error(w, "la novedad se encuentra rechazada", http.StatusConflict)
			return
		}
		if nivelActual <= 0 {
			nivelActual = 0
		}
		update["estado_novedad"] = "en_aprobacion"
		update["nivel_aprobacion_actual"] = nivelActual
		historial = append(historial, map[string]interface{}{
			"accion":         "solicitar_aprobacion",
			"actor":          actor,
			"fecha":          nowText,
			"nivel_anterior": nivelActual,
			"nivel_nuevo":    nivelActual,
			"comentario":     comentario,
		})

	case "aprobar":
		if estadoActual == "rechazada" {
			http.Error(w, "la novedad se encuentra rechazada", http.StatusConflict)
			return
		}
		if estadoActual == "anulada" || estadoActual == "cancelada" {
			http.Error(w, "la novedad no permite aprobacion", http.StatusConflict)
			return
		}
		if estadoActual == "aprobada" || estadoActual == "contabilizada" || estadoActual == "cerrada" {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":                  true,
				"empresa_id":          empresaID,
				"id":                  id,
				"estado_novedad":      estadoActual,
				"mensaje":             "novedad ya aprobada o cerrada",
				"nivel_actual":        nivelActual,
				"nivel_requerido":     nivelRequerido,
				"aprobadores_totales": len(aprobadores),
			})
			return
		}

		nivelAnterior := nivelActual
		nivelActual++
		if nivelActual > nivelRequerido {
			nivelActual = nivelRequerido
		}

		estadoNuevo := "en_aprobacion"
		if nivelActual >= nivelRequerido {
			estadoNuevo = "aprobada"
			update["fecha_aprobacion_final"] = nowText
			update["aprobado_por"] = actor
		}

		if !containsStringFold(aprobadores, actor) {
			aprobadores = append(aprobadores, actor)
		}
		historial = append(historial, map[string]interface{}{
			"accion":         "aprobar",
			"actor":          actor,
			"fecha":          nowText,
			"nivel_anterior": nivelAnterior,
			"nivel_nuevo":    nivelActual,
			"comentario":     comentario,
		})

		update["estado_novedad"] = estadoNuevo
		update["nivel_aprobacion_actual"] = nivelActual
		update["aprobadores_json"] = marshalToJSONString(aprobadores)

		if estadoNuevo == "aprobada" {
			diasNovedad := rrhhComputeDiasSolicitud(item)
			if diasNovedad > 0 {
				update["dias"] = reportesRound(diasNovedad)
			}

			fechaCorte := now
			if parsed, ok := ventasParseDateTime(finanzasFirstNonBlank(genericStringValue(item["fecha_fin"]), genericStringValue(item["fecha_inicio"]))); ok {
				fechaCorte = parsed.In(time.Local)
			}

			saldoAntes, saldoErr := buildRRHHVacacionesSaldoByFilter(dbEmp, empresaID, rrhhVacacionesSaldoFilter{
				EmpleadoNominaID: anyToInt64(item["empleado_nomina_id"]),
				EmpleadoID:       anyToInt64(item["empleado_id"]),
				EmpleadoNombre:   genericStringValue(item["empleado_nombre"]),
				ExcluirNovedadID: id,
				FechaCorte:       fechaCorte,
			})
			if saldoErr == nil {
				diasAntes := ventasAnyToFloat64(saldoAntes["dias_pendientes"])
				diasDespues := diasAntes - diasNovedad
				if diasDespues < 0 {
					diasDespues = 0
				}

				snapshot := map[string]interface{}{
					"fecha_corte":             fechaCorte.Format("2006-01-02"),
					"dias_solicitud":          reportesRound(diasNovedad),
					"dias_pendientes_antes":   reportesRound(diasAntes),
					"dias_pendientes_despues": reportesRound(diasDespues),
					"referencia_novedad_id":   id,
				}
				update["periodo_acumulado_desde"] = genericStringValue(saldoAntes["fecha_ingreso"])
				update["periodo_acumulado_hasta"] = genericStringValue(snapshot["fecha_corte"])
				update["saldo_dias_antes"] = reportesRound(diasAntes)
				update["saldo_dias_despues"] = reportesRound(diasDespues)
				update["saldo_snapshot_json"] = marshalToJSONString(snapshot)
			}
		}

	default:
		http.Error(w, "accion de aprobacion RRHH no soportada", http.StatusBadRequest)
		return
	}

	update["historial_aprobaciones_json"] = marshalToJSONString(historial)
	update["observaciones"] = appendGenericObservation(genericStringValue(item["observaciones"]), fmt.Sprintf("rrhh_%s actor=%s nivel=%d/%d", action, actor, anyToInt64(update["nivel_aprobacion_actual"]), nivelRequerido))

	if err := dbpkg.UpdateEmpresaGenericRow(dbEmp, cfgRRHHVacLic.Table, empresaID, id, update, cfgRRHHVacLic.AllowedColumns); err != nil {
		http.Error(w, "no se pudo actualizar aprobacion RRHH", http.StatusInternalServerError)
		return
	}

	updated, _ := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgRRHHVacLic.Table, empresaID, id)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                  true,
		"empresa_id":          empresaID,
		"id":                  id,
		"estado_novedad":      genericStringValue(updated["estado_novedad"]),
		"nivel_actual":        anyToInt64(updated["nivel_aprobacion_actual"]),
		"nivel_requerido":     anyToInt64(updated["nivel_aprobacion_requerido"]),
		"aprobadores_totales": len(parseJSONArrayStrings(genericStringValue(updated["aprobadores_json"]))),
		"item":                updated,
	})
}

func handleRRHHVacacionesRechazoAction(dbEmp *sql.DB, w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "id es obligatorio", http.StatusBadRequest)
		return
	}

	item, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgRRHHVacLic.Table, empresaID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "novedad RRHH no encontrada", http.StatusNotFound)
			return
		}
		http.Error(w, "no se pudo consultar novedad RRHH", http.StatusInternalServerError)
		return
	}

	actor := strings.TrimSpace(adminEmailFromRequest(r))
	if actor == "" {
		actor = "sistema"
	}
	nowText := time.Now().In(time.Local).Format("2006-01-02 15:04:05")
	comentario := strings.TrimSpace(finanzasFirstNonBlank(genericStringValue(payload["comentario"]), genericStringValue(payload["motivo"]), "rechazo de solicitud RRHH"))

	historial := parseJSONArrayObjects(genericStringValue(item["historial_aprobaciones_json"]))
	historial = append(historial, map[string]interface{}{
		"accion":         "rechazar",
		"actor":          actor,
		"fecha":          nowText,
		"nivel_anterior": anyToInt64(item["nivel_aprobacion_actual"]),
		"nivel_nuevo":    anyToInt64(item["nivel_aprobacion_actual"]),
		"comentario":     comentario,
	})

	update := map[string]interface{}{
		"estado_novedad":              "rechazada",
		"historial_aprobaciones_json": marshalToJSONString(historial),
		"aprobado_por":                actor,
		"observaciones": appendGenericObservation(
			genericStringValue(item["observaciones"]),
			fmt.Sprintf("rrhh_rechazo actor=%s motivo=%s", actor, comentario),
		),
	}

	if err := dbpkg.UpdateEmpresaGenericRow(dbEmp, cfgRRHHVacLic.Table, empresaID, id, update, cfgRRHHVacLic.AllowedColumns); err != nil {
		http.Error(w, "no se pudo rechazar novedad RRHH", http.StatusInternalServerError)
		return
	}

	updated, _ := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgRRHHVacLic.Table, empresaID, id)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"id":         id,
		"estado":     genericStringValue(updated["estado_novedad"]),
		"item":       updated,
	})
}

func handleRRHHVacacionesVincularNominaAction(dbEmp *sql.DB, w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "id es obligatorio", http.StatusBadRequest)
		return
	}

	item, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgRRHHVacLic.Table, empresaID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "novedad RRHH no encontrada", http.StatusNotFound)
			return
		}
		http.Error(w, "no se pudo consultar novedad RRHH", http.StatusInternalServerError)
		return
	}

	estadoActual := strings.ToLower(strings.TrimSpace(genericStringValue(item["estado_novedad"])))
	if estadoActual != "aprobada" && estadoActual != "contabilizada" && estadoActual != "cerrada" {
		http.Error(w, "solo se pueden vincular novedades aprobadas o contabilizadas", http.StatusConflict)
		return
	}

	empleadoNominaID := anyToInt64(payload["empleado_nomina_id"])
	if empleadoNominaID <= 0 {
		empleadoNominaID = anyToInt64(item["empleado_nomina_id"])
	}
	empleadoID := anyToInt64(item["empleado_id"])
	empleadoNombre := genericStringValue(item["empleado_nombre"])

	ref, err := loadRRHHNominaEmpleadoRef(dbEmp, empresaID, empleadoNominaID, empleadoID, empleadoNombre)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "empleado de nomina no encontrado para vincular", http.StatusNotFound)
			return
		}
		http.Error(w, "no se pudo consultar empleado de nomina", http.StatusInternalServerError)
		return
	}
	empleadoNominaID = ref.NominaID

	nominaLiquidacionID := anyToInt64(payload["nomina_liquidacion_id"])
	if nominaLiquidacionID <= 0 {
		nominaLiquidacionID = anyToInt64(r.URL.Query().Get("nomina_liquidacion_id"))
	}

	nominaPeriodoDesde := strings.TrimSpace(finanzasFirstNonBlank(genericStringValue(payload["nomina_periodo_desde"]), strings.TrimSpace(r.URL.Query().Get("nomina_periodo_desde"))))
	nominaPeriodoHasta := strings.TrimSpace(finanzasFirstNonBlank(genericStringValue(payload["nomina_periodo_hasta"]), strings.TrimSpace(r.URL.Query().Get("nomina_periodo_hasta"))))

	if nominaLiquidacionID > 0 {
		var liqEmpleadoNominaID int64
		var liqDesde string
		var liqHasta string
		err := dbEmp.QueryRow(`SELECT
			COALESCE(empleado_nomina_id, 0),
			COALESCE(periodo_desde, ''),
			COALESCE(periodo_hasta, '')
		FROM empresa_nomina_liquidaciones
		WHERE empresa_id = ? AND id = ?
		LIMIT 1`, empresaID, nominaLiquidacionID).Scan(&liqEmpleadoNominaID, &liqDesde, &liqHasta)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "liquidacion de nomina no encontrada", http.StatusNotFound)
				return
			}
			http.Error(w, "no se pudo consultar liquidacion de nomina", http.StatusInternalServerError)
			return
		}
		if liqEmpleadoNominaID > 0 && liqEmpleadoNominaID != empleadoNominaID {
			http.Error(w, "la liquidacion no corresponde al empleado de la novedad", http.StatusConflict)
			return
		}
		if nominaPeriodoDesde == "" {
			nominaPeriodoDesde = strings.TrimSpace(liqDesde)
		}
		if nominaPeriodoHasta == "" {
			nominaPeriodoHasta = strings.TrimSpace(liqHasta)
		}
	} else {
		fechaInicio := rrhhNormalizeDateOnly(finanzasFirstNonBlank(genericStringValue(item["fecha_inicio"]), nominaPeriodoDesde))
		fechaFin := rrhhNormalizeDateOnly(finanzasFirstNonBlank(genericStringValue(item["fecha_fin"]), nominaPeriodoHasta, fechaInicio))
		if fechaInicio != "" && fechaFin != "" {
			var foundID int64
			var foundDesde string
			var foundHasta string
			err := dbEmp.QueryRow(`SELECT
				id,
				COALESCE(periodo_desde, ''),
				COALESCE(periodo_hasta, '')
			FROM empresa_nomina_liquidaciones
			WHERE empresa_id = ?
			  AND COALESCE(empleado_nomina_id, 0) = ?
			  AND date(COALESCE(periodo_desde, '')) <= date(?)
			  AND date(COALESCE(periodo_hasta, '')) >= date(?)
			ORDER BY date(COALESCE(periodo_hasta, periodo_desde)) DESC, id DESC
			LIMIT 1`, empresaID, empleadoNominaID, fechaFin, fechaInicio).Scan(&foundID, &foundDesde, &foundHasta)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "no se pudo resolver liquidacion de nomina", http.StatusInternalServerError)
				return
			}
			if err == nil {
				nominaLiquidacionID = foundID
				if nominaPeriodoDesde == "" {
					nominaPeriodoDesde = strings.TrimSpace(foundDesde)
				}
				if nominaPeriodoHasta == "" {
					nominaPeriodoHasta = strings.TrimSpace(foundHasta)
				}
			}
		}
	}

	if nominaPeriodoDesde == "" && nominaPeriodoHasta != "" {
		nominaPeriodoDesde = nominaPeriodoHasta
	}
	if nominaPeriodoHasta == "" && nominaPeriodoDesde != "" {
		nominaPeriodoHasta = nominaPeriodoDesde
	}
	if nominaLiquidacionID <= 0 && (nominaPeriodoDesde == "" || nominaPeriodoHasta == "") {
		http.Error(w, "nomina_liquidacion_id o periodo de nomina es obligatorio", http.StatusBadRequest)
		return
	}

	actor := strings.TrimSpace(adminEmailFromRequest(r))
	if actor == "" {
		actor = "sistema"
	}
	nowText := time.Now().In(time.Local).Format("2006-01-02 15:04:05")

	historial := parseJSONArrayObjects(genericStringValue(item["historial_aprobaciones_json"]))
	historial = append(historial, map[string]interface{}{
		"accion":                "vincular_nomina",
		"actor":                 actor,
		"fecha":                 nowText,
		"nomina_liquidacion_id": nominaLiquidacionID,
		"nomina_periodo_desde":  nominaPeriodoDesde,
		"nomina_periodo_hasta":  nominaPeriodoHasta,
	})

	update := map[string]interface{}{
		"empleado_nomina_id":          empleadoNominaID,
		"nomina_liquidacion_id":       nominaLiquidacionID,
		"nomina_periodo_desde":        nominaPeriodoDesde,
		"nomina_periodo_hasta":        nominaPeriodoHasta,
		"nomina_vinculada_en":         nowText,
		"nomina_vinculada_por":        actor,
		"historial_aprobaciones_json": marshalToJSONString(historial),
		"observaciones": appendGenericObservation(
			genericStringValue(item["observaciones"]),
			fmt.Sprintf("rrhh_vincular_nomina actor=%s liquidacion_id=%d periodo=%s/%s", actor, nominaLiquidacionID, nominaPeriodoDesde, nominaPeriodoHasta),
		),
	}
	if estadoActual == "aprobada" {
		update["estado_novedad"] = "contabilizada"
	}

	if err := dbpkg.UpdateEmpresaGenericRow(dbEmp, cfgRRHHVacLic.Table, empresaID, id, update, cfgRRHHVacLic.AllowedColumns); err != nil {
		http.Error(w, "no se pudo vincular novedad RRHH a nomina", http.StatusInternalServerError)
		return
	}

	updated, _ := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgRRHHVacLic.Table, empresaID, id)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                    true,
		"empresa_id":            empresaID,
		"id":                    id,
		"nomina_liquidacion_id": nominaLiquidacionID,
		"nomina_periodo_desde":  nominaPeriodoDesde,
		"nomina_periodo_hasta":  nominaPeriodoHasta,
		"item":                  updated,
	})
}

func buildRRHHVacacionesSaldoByFilter(dbEmp *sql.DB, empresaID int64, filter rrhhVacacionesSaldoFilter) (map[string]interface{}, error) {
	if filter.FechaCorte.IsZero() {
		filter.FechaCorte = time.Now().In(time.Local)
	}

	ref, err := loadRRHHNominaEmpleadoRef(dbEmp, empresaID, filter.EmpleadoNominaID, filter.EmpleadoID, filter.EmpleadoNombre)
	if err != nil {
		return nil, err
	}

	fechaIngresoRaw := strings.TrimSpace(ref.FechaIngreso)
	if fechaIngresoRaw == "" {
		return nil, fmt.Errorf("empleado sin fecha_ingreso en nomina")
	}

	fechaIngreso, ok := ventasParseDateTime(fechaIngresoRaw)
	if !ok {
		return nil, fmt.Errorf("fecha_ingreso invalida")
	}

	fechaIngresoDate, _ := time.Parse("2006-01-02", fechaIngreso.In(time.Local).Format("2006-01-02"))
	fechaCorteDate, _ := time.Parse("2006-01-02", filter.FechaCorte.In(time.Local).Format("2006-01-02"))
	diasServicio := int64(fechaCorteDate.Sub(fechaIngresoDate).Hours()/24) + 1
	if diasServicio < 0 {
		diasServicio = 0
	}
	diasAcumulados := (float64(diasServicio) * 15.0) / 360.0

	queryTomados := `SELECT
		COALESCE(SUM(COALESCE(dias, 0)), 0),
		COALESCE(COUNT(1), 0)
	FROM empresa_rrhh_vacaciones_licencias
	WHERE empresa_id = ?
	  AND LOWER(COALESCE(estado, 'activo')) = 'activo'
	  AND LOWER(COALESCE(estado_novedad, '')) IN ('aprobada', 'contabilizada', 'cerrada')
	  AND LOWER(COALESCE(tipo_novedad, '')) LIKE 'vacacion%'
	  AND date(COALESCE(NULLIF(fecha_inicio, ''), NULLIF(fecha_fin, ''), fecha_creacion)) <= date(?)`
	argsTomados := []interface{}{empresaID, fechaCorteDate.Format("2006-01-02")}

	if ref.NominaID > 0 {
		queryTomados += ` AND COALESCE(empleado_nomina_id, 0) = ?`
		argsTomados = append(argsTomados, ref.NominaID)
	} else if ref.EmpleadoID > 0 {
		queryTomados += ` AND COALESCE(empleado_id, 0) = ?`
		argsTomados = append(argsTomados, ref.EmpleadoID)
	} else {
		queryTomados += ` AND UPPER(COALESCE(empleado_nombre, '')) = UPPER(?)`
		argsTomados = append(argsTomados, ref.Nombre)
	}
	if filter.ExcluirNovedadID > 0 {
		queryTomados += ` AND id <> ?`
		argsTomados = append(argsTomados, filter.ExcluirNovedadID)
	}

	var diasTomados float64
	var solicitudesAprobadas int64
	if err := dbEmp.QueryRow(queryTomados, argsTomados...).Scan(&diasTomados, &solicitudesAprobadas); err != nil {
		return nil, err
	}

	queryPendientes := `SELECT COALESCE(COUNT(1), 0)
	FROM empresa_rrhh_vacaciones_licencias
	WHERE empresa_id = ?
	  AND LOWER(COALESCE(estado, 'activo')) = 'activo'
	  AND LOWER(COALESCE(estado_novedad, '')) IN ('solicitada', 'en_aprobacion', 'aprobada_parcial')`
	argsPendientes := []interface{}{empresaID}
	if ref.NominaID > 0 {
		queryPendientes += ` AND COALESCE(empleado_nomina_id, 0) = ?`
		argsPendientes = append(argsPendientes, ref.NominaID)
	} else if ref.EmpleadoID > 0 {
		queryPendientes += ` AND COALESCE(empleado_id, 0) = ?`
		argsPendientes = append(argsPendientes, ref.EmpleadoID)
	} else {
		queryPendientes += ` AND UPPER(COALESCE(empleado_nombre, '')) = UPPER(?)`
		argsPendientes = append(argsPendientes, ref.Nombre)
	}
	if filter.ExcluirNovedadID > 0 {
		queryPendientes += ` AND id <> ?`
		argsPendientes = append(argsPendientes, filter.ExcluirNovedadID)
	}

	solicitudesPendientes := int64(0)
	if err := dbEmp.QueryRow(queryPendientes, argsPendientes...).Scan(&solicitudesPendientes); err != nil {
		return nil, err
	}

	diasPendientes := diasAcumulados - diasTomados
	if diasPendientes < 0 {
		diasPendientes = 0
	}

	return map[string]interface{}{
		"empleado_nomina_id":     ref.NominaID,
		"empleado_id":            ref.EmpleadoID,
		"empleado_nombre":        ref.Nombre,
		"fecha_ingreso":          fechaIngresoDate.Format("2006-01-02"),
		"fecha_corte":            fechaCorteDate.Format("2006-01-02"),
		"dias_servicio":          diasServicio,
		"dias_acumulados":        reportesRound(diasAcumulados),
		"dias_tomados_aprobados": reportesRound(diasTomados),
		"dias_pendientes":        reportesRound(diasPendientes),
		"solicitudes_aprobadas":  solicitudesAprobadas,
		"solicitudes_pendientes": solicitudesPendientes,
	}, nil
}

func loadRRHHNominaEmpleadoRef(dbEmp *sql.DB, empresaID, empleadoNominaID, empleadoID int64, empleadoNombre string) (*rrhhNominaEmpleadoRef, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id invalido")
	}

	query := `SELECT
		id,
		COALESCE(empleado_id, 0),
		COALESCE(empleado_nombre, ''),
		COALESCE(fecha_ingreso, '')
	FROM empresa_nomina_empleados
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	if empleadoNominaID > 0 {
		query += ` AND id = ?`
		args = append(args, empleadoNominaID)
	} else if empleadoID > 0 {
		query += ` AND COALESCE(empleado_id, 0) = ?`
		args = append(args, empleadoID)
	} else {
		nombre := strings.TrimSpace(empleadoNombre)
		if nombre == "" {
			return nil, fmt.Errorf("empleado de nomina no identificado")
		}
		query += ` AND UPPER(COALESCE(empleado_nombre, '')) LIKE UPPER(?)`
		args = append(args, "%"+nombre+"%")
	}

	query += ` ORDER BY CASE WHEN LOWER(COALESCE(estado, 'activo')) = 'activo' THEN 0 ELSE 1 END, id DESC LIMIT 1`

	out := &rrhhNominaEmpleadoRef{}
	err := dbEmp.QueryRow(query, args...).Scan(&out.NominaID, &out.EmpleadoID, &out.Nombre, &out.FechaIngreso)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func parseJSONArrayStrings(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{}
	}
	direct := make([]string, 0)
	if err := json.Unmarshal([]byte(raw), &direct); err == nil {
		out := make([]string, 0, len(direct))
		for _, item := range direct {
			trimmed := strings.TrimSpace(item)
			if trimmed == "" || containsStringFold(out, trimmed) {
				continue
			}
			out = append(out, trimmed)
		}
		return out
	}

	generic := make([]interface{}, 0)
	if err := json.Unmarshal([]byte(raw), &generic); err != nil {
		return []string{}
	}
	out := make([]string, 0, len(generic))
	for _, item := range generic {
		str := strings.TrimSpace(genericStringValue(item))
		if str == "" || containsStringFold(out, str) {
			continue
		}
		out = append(out, str)
	}
	return out
}

func parseJSONArrayObjects(raw string) []map[string]interface{} {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []map[string]interface{}{}
	}
	out := make([]map[string]interface{}, 0)
	if err := json.Unmarshal([]byte(raw), &out); err == nil {
		return out
	}
	return []map[string]interface{}{}
}

func marshalToJSONString(v interface{}) string {
	encoded, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func containsStringFold(values []string, candidate string) bool {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return false
	}
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), candidate) {
			return true
		}
	}
	return false
}

func rrhhComputeDiasSolicitud(item map[string]interface{}) float64 {
	dias := ventasAnyToFloat64(item["dias"])
	if dias > 0 {
		return dias
	}

	fechaInicioRaw := strings.TrimSpace(genericStringValue(item["fecha_inicio"]))
	fechaFinRaw := strings.TrimSpace(genericStringValue(item["fecha_fin"]))
	if fechaInicioRaw == "" || fechaFinRaw == "" {
		return 0
	}

	fechaInicio, okInicio := ventasParseDateTime(fechaInicioRaw)
	fechaFin, okFin := ventasParseDateTime(fechaFinRaw)
	if !okInicio || !okFin {
		return 0
	}

	startDate, _ := time.Parse("2006-01-02", fechaInicio.In(time.Local).Format("2006-01-02"))
	endDate, _ := time.Parse("2006-01-02", fechaFin.In(time.Local).Format("2006-01-02"))
	if endDate.Before(startDate) {
		startDate, endDate = endDate, startDate
	}
	diasCalc := int64(endDate.Sub(startDate).Hours()/24) + 1
	if diasCalc < 0 {
		diasCalc = 0
	}
	return float64(diasCalc)
}

func rrhhNormalizeDateOnly(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if parsed, ok := ventasParseDateTime(raw); ok {
		return parsed.In(time.Local).Format("2006-01-02")
	}
	if len(raw) >= 10 {
		candidate := raw[:10]
		if _, err := time.Parse("2006-01-02", candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func handleInventarioLotesSeriesValidarDisponibilidadAction(dbEmp *sql.DB, w http.ResponseWriter, r *http.Request) {
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
	codigo := strings.TrimSpace(r.URL.Query().Get("codigo_lote_serie"))
	if id <= 0 && codigo == "" {
		http.Error(w, "id o codigo_lote_serie es obligatorio", http.StatusBadRequest)
		return
	}

	row, err := loadInventarioLoteSerieByIDOrCode(dbEmp, empresaID, id, codigo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "lote/serie no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "no se pudo consultar lote/serie", http.StatusInternalServerError)
		return
	}

	now := time.Now().In(time.Local)
	if updateErr := ensureLoteSerieVencimientoBloqueo(dbEmp, empresaID, row, now, strings.TrimSpace(adminEmailFromRequest(r))); updateErr == nil {
		row, _ = loadInventarioLoteSerieByIDOrCode(dbEmp, empresaID, anyToInt64(row["id"]), "")
	}

	cantidadSolicitada := 0.0
	if rawCantidad := strings.TrimSpace(r.URL.Query().Get("cantidad")); rawCantidad != "" {
		parsed, parseErr := strconv.ParseFloat(rawCantidad, 64)
		if parseErr != nil || parsed < 0 {
			http.Error(w, "cantidad invalida", http.StatusBadRequest)
			return
		}
		cantidadSolicitada = parsed
	}

	estadoLote := strings.ToLower(strings.TrimSpace(genericStringValue(row["estado_lote"])))
	estadoRegistro := strings.ToLower(strings.TrimSpace(genericStringValue(row["estado"])))
	bloqueadoVenta := anyToInt64(row["bloqueado_venta"]) > 0
	vencido := loteSerieEstaVencido(genericStringValue(row["fecha_vencimiento"]), now)
	if estadoLote == "vencido" {
		vencido = true
	}

	disponible := ventasAnyToFloat64(row["cantidad_disponible"])
	activo := estadoRegistro != "inactivo"
	noBloqueadoOperativamente := estadoLote != "inactivo" && estadoLote != "bloqueado" && !bloqueadoVenta && !vencido
	disponibleOperacion := activo && noBloqueadoOperativamente
	disponibleReserva := disponibleOperacion && (cantidadSolicitada <= 0 || disponible+1e-9 >= cantidadSolicitada)
	disponibleVenta := disponibleReserva

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                      true,
		"empresa_id":              empresaID,
		"item":                    row,
		"cantidad_solicitada":     reportesRound(cantidadSolicitada),
		"cantidad_disponible":     reportesRound(disponible),
		"vencido":                 vencido,
		"bloqueado_venta":         bloqueadoVenta,
		"bloqueo_motivo":          genericStringValue(row["bloqueo_motivo"]),
		"disponible_para_reserva": disponibleReserva,
		"disponible_para_venta":   disponibleVenta,
	})
}

func handleInventarioLotesSeriesOperacionAction(dbEmp *sql.DB, defaultAction string, w http.ResponseWriter, r *http.Request) {
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
	codigo := finanzasFirstNonBlank(genericStringValue(payload["codigo_lote_serie"]), strings.TrimSpace(r.URL.Query().Get("codigo_lote_serie")))
	if id <= 0 && codigo == "" {
		http.Error(w, "id o codigo_lote_serie es obligatorio", http.StatusBadRequest)
		return
	}

	operacion := normalizeLoteSerieOperacion(finanzasFirstNonBlank(genericStringValue(payload["tipo_operacion"]), strings.TrimSpace(r.URL.Query().Get("tipo_operacion")), defaultAction))
	if operacion == "" {
		http.Error(w, "tipo_operacion no soportada", http.StatusBadRequest)
		return
	}

	cantidad := ventasAnyToFloat64(payload["cantidad"])
	if cantidad <= 0 {
		http.Error(w, "cantidad es obligatoria y debe ser mayor a cero", http.StatusBadRequest)
		return
	}

	actor := strings.TrimSpace(adminEmailFromRequest(r))
	if actor == "" {
		actor = "sistema"
	}

	row, err := loadInventarioLoteSerieByIDOrCode(dbEmp, empresaID, id, codigo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "lote/serie no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "no se pudo consultar lote/serie", http.StatusInternalServerError)
		return
	}

	now := time.Now().In(time.Local)
	nowText := now.Format("2006-01-02 15:04:05")

	if err := ensureLoteSerieVencimientoBloqueo(dbEmp, empresaID, row, now, actor); err == nil {
		row, _ = loadInventarioLoteSerieByIDOrCode(dbEmp, empresaID, anyToInt64(row["id"]), "")
	}

	estadoLote := strings.ToLower(strings.TrimSpace(genericStringValue(row["estado_lote"])))
	bloqueadoVenta := anyToInt64(row["bloqueado_venta"]) > 0
	vencido := loteSerieEstaVencido(genericStringValue(row["fecha_vencimiento"]), now) || estadoLote == "vencido"

	if (operacion == "reserva" || operacion == "venta") && (vencido || bloqueadoVenta) {
		writeJSON(w, http.StatusConflict, map[string]interface{}{
			"ok":             false,
			"empresa_id":     empresaID,
			"id":             anyToInt64(row["id"]),
			"operacion":      operacion,
			"motivo_bloqueo": finanzasFirstNonBlank(genericStringValue(row["bloqueo_motivo"]), "lote/serie vencido o bloqueado para venta y reserva"),
		})
		return
	}

	disponibleAntes := ventasAnyToFloat64(row["cantidad_disponible"])
	reservadoAntes := ventasAnyToFloat64(row["reservado_cantidad"])
	vendidoAntes := ventasAnyToFloat64(row["vendido_cantidad"])

	disponibleDespues := disponibleAntes
	reservadoDespues := reservadoAntes
	vendidoDespues := vendidoAntes

	switch operacion {
	case "reserva":
		if disponibleAntes+1e-9 < cantidad {
			writeJSON(w, http.StatusConflict, map[string]interface{}{"ok": false, "detalle": "cantidad_disponible insuficiente para reserva"})
			return
		}
		disponibleDespues -= cantidad
		reservadoDespues += cantidad

	case "venta":
		pendiente := cantidad
		if reservadoDespues > 0 {
			consumoReserva := pendiente
			if reservadoDespues < consumoReserva {
				consumoReserva = reservadoDespues
			}
			reservadoDespues -= consumoReserva
			pendiente -= consumoReserva
		}
		if pendiente > 0 {
			if disponibleDespues+1e-9 < pendiente {
				writeJSON(w, http.StatusConflict, map[string]interface{}{"ok": false, "detalle": "cantidad_disponible insuficiente para venta"})
				return
			}
			disponibleDespues -= pendiente
		}
		vendidoDespues += cantidad

	case "liberar_reserva":
		if reservadoAntes+1e-9 < cantidad {
			writeJSON(w, http.StatusConflict, map[string]interface{}{"ok": false, "detalle": "cantidad reservada insuficiente para liberar"})
			return
		}
		reservadoDespues -= cantidad
		disponibleDespues += cantidad

	case "ajuste_entrada":
		disponibleDespues += cantidad

	case "ajuste_salida", "devolucion_proveedor":
		if disponibleAntes+1e-9 < cantidad {
			writeJSON(w, http.StatusConflict, map[string]interface{}{"ok": false, "detalle": "cantidad_disponible insuficiente para salida"})
			return
		}
		disponibleDespues -= cantidad
	}

	if disponibleDespues < 0 {
		disponibleDespues = 0
	}
	if reservadoDespues < 0 {
		reservadoDespues = 0
	}
	if vendidoDespues < 0 {
		vendidoDespues = 0
	}

	estadoLoteNuevo := estadoLote
	if estadoLoteNuevo == "" || estadoLoteNuevo == "vencido" || estadoLoteNuevo == "agotado" {
		estadoLoteNuevo = "activo"
	}
	if disponibleDespues <= 0.000001 {
		estadoLoteNuevo = "agotado"
	}
	if vencido {
		estadoLoteNuevo = "vencido"
	}

	referenciaTipo := finanzasFirstNonBlank(genericStringValue(payload["referencia_tipo"]), genericStringValue(payload["origen"]), "operacion_lote")
	referenciaCodigo := finanzasFirstNonBlank(
		genericStringValue(payload["referencia_codigo"]),
		genericStringValue(payload["carrito_codigo"]),
		genericStringValue(payload["reserva_codigo"]),
		fmt.Sprintf("%s-%d", strings.ToUpper(operacion), anyToInt64(row["id"])),
	)
	clienteID := anyToInt64(payload["cliente_id"])
	clienteNombre := genericStringValue(payload["cliente_nombre"])

	detalleJSON, _ := json.Marshal(map[string]interface{}{
		"antes": map[string]interface{}{
			"cantidad_disponible": reportesRound(disponibleAntes),
			"reservado_cantidad":  reportesRound(reservadoAntes),
			"vendido_cantidad":    reportesRound(vendidoAntes),
		},
		"despues": map[string]interface{}{
			"cantidad_disponible": reportesRound(disponibleDespues),
			"reservado_cantidad":  reportesRound(reservadoDespues),
			"vendido_cantidad":    reportesRound(vendidoDespues),
		},
		"motivo": genericStringValue(payload["motivo"]),
	})

	observaciones := appendGenericObservation(genericStringValue(row["observaciones"]), fmt.Sprintf("operacion_lote tipo=%s cantidad=%.6f referencia=%s", operacion, cantidad, referenciaCodigo))

	tx, err := dbEmp.Begin()
	if err != nil {
		http.Error(w, "no se pudo iniciar transaccion de lote/serie", http.StatusInternalServerError)
		return
	}

	movimientoID := int64(0)
	defer func() {
		if movimientoID == 0 {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.Exec(`UPDATE inventario_lotes_series
	SET cantidad_disponible = ?,
		reservado_cantidad = ?,
		vendido_cantidad = ?,
		estado_lote = ?,
		bloqueado_venta = ?,
		bloqueo_motivo = ?,
		ultima_operacion_tipo = ?,
		ultima_operacion_ref = ?,
		ultima_operacion_en = ?,
		observaciones = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`,
		disponibleDespues,
		reservadoDespues,
		vendidoDespues,
		estadoLoteNuevo,
		0,
		"",
		operacion,
		referenciaCodigo,
		nowText,
		observaciones,
		empresaID,
		anyToInt64(row["id"]),
	)
	if err != nil {
		http.Error(w, "no se pudo actualizar lote/serie", http.StatusInternalServerError)
		return
	}

	resMov, err := tx.Exec(`INSERT INTO inventario_lotes_series_movimientos (
		empresa_id,
		lote_serie_id,
		producto_id,
		bodega_id,
		codigo_lote_serie,
		tipo_operacion,
		cantidad,
		saldo_lote,
		referencia_tipo,
		referencia_codigo,
		cliente_id,
		cliente_nombre,
		detalle_json,
		fecha_operacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'activo', ?)`,
		empresaID,
		anyToInt64(row["id"]),
		anyToInt64(row["producto_id"]),
		anyToInt64(row["bodega_id"]),
		genericStringValue(row["codigo_lote_serie"]),
		operacion,
		cantidad,
		disponibleDespues,
		referenciaTipo,
		referenciaCodigo,
		clienteID,
		clienteNombre,
		string(detalleJSON),
		nowText,
		actor,
		fmt.Sprintf("operacion=%s", operacion),
	)
	if err != nil {
		http.Error(w, "no se pudo registrar trazabilidad de lote/serie", http.StatusInternalServerError)
		return
	}
	movimientoID, _ = resMov.LastInsertId()

	if err := tx.Commit(); err != nil {
		movimientoID = 0
		http.Error(w, "no se pudo confirmar transaccion de lote/serie", http.StatusInternalServerError)
		return
	}

	item, _ := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgLotesSeries.Table, empresaID, anyToInt64(row["id"]))
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                         true,
		"empresa_id":                 empresaID,
		"id":                         anyToInt64(row["id"]),
		"tipo_operacion":             operacion,
		"cantidad":                   reportesRound(cantidad),
		"movimiento_trazabilidad_id": movimientoID,
		"item":                       item,
	})
}

func handleInventarioLotesSeriesTrazabilidadAction(dbEmp *sql.DB, w http.ResponseWriter, r *http.Request) {
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
	codigo := strings.TrimSpace(r.URL.Query().Get("codigo_lote_serie"))

	limit, err := parseIntQueryOptional(r, "limit")
	if err != nil {
		http.Error(w, "limit invalido", http.StatusBadRequest)
		return
	}
	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}

	query := `SELECT
		id,
		COALESCE(lote_serie_id, 0),
		COALESCE(producto_id, 0),
		COALESCE(bodega_id, 0),
		COALESCE(codigo_lote_serie, ''),
		COALESCE(tipo_operacion, ''),
		COALESCE(cantidad, 0),
		COALESCE(saldo_lote, 0),
		COALESCE(referencia_tipo, ''),
		COALESCE(referencia_codigo, ''),
		COALESCE(cliente_id, 0),
		COALESCE(cliente_nombre, ''),
		COALESCE(fecha_operacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(observaciones, '')
	FROM inventario_lotes_series_movimientos
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	if id > 0 {
		query += ` AND lote_serie_id = ?`
		args = append(args, id)
	} else if codigo != "" {
		query += ` AND UPPER(COALESCE(codigo_lote_serie, '')) = UPPER(?)`
		args = append(args, codigo)
	}

	if !parseBoolQuery(r, "include_inactive") {
		query += ` AND LOWER(COALESCE(estado, 'activo')) = 'activo'`
	}

	query += ` ORDER BY datetime(COALESCE(NULLIF(fecha_operacion, ''), fecha_creacion)) DESC, id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := dbEmp.Query(query, args...)
	if err != nil {
		http.Error(w, "no se pudo consultar trazabilidad de lotes/series", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	totalCantidad := 0.0
	totalPorOperacion := map[string]float64{}
	conteoPorOperacion := map[string]int64{}

	for rows.Next() {
		item := map[string]interface{}{}
		var idMov int64
		var loteID int64
		var productoID int64
		var bodegaID int64
		var codigoLote string
		var tipoOperacion string
		var cantidad float64
		var saldoLote float64
		var referenciaTipo string
		var referenciaCodigo string
		var clienteID int64
		var clienteNombre string
		var fechaOperacion string
		var usuarioCreador string
		var observaciones string
		if err := rows.Scan(
			&idMov,
			&loteID,
			&productoID,
			&bodegaID,
			&codigoLote,
			&tipoOperacion,
			&cantidad,
			&saldoLote,
			&referenciaTipo,
			&referenciaCodigo,
			&clienteID,
			&clienteNombre,
			&fechaOperacion,
			&usuarioCreador,
			&observaciones,
		); err != nil {
			http.Error(w, "no se pudo leer trazabilidad de lotes/series", http.StatusInternalServerError)
			return
		}
		tipoOperacion = normalizeLoteSerieOperacion(tipoOperacion)
		item["id"] = idMov
		item["lote_serie_id"] = loteID
		item["producto_id"] = productoID
		item["bodega_id"] = bodegaID
		item["codigo_lote_serie"] = codigoLote
		item["tipo_operacion"] = tipoOperacion
		item["cantidad"] = reportesRound(cantidad)
		item["saldo_lote"] = reportesRound(saldoLote)
		item["referencia_tipo"] = referenciaTipo
		item["referencia_codigo"] = referenciaCodigo
		item["cliente_id"] = clienteID
		item["cliente_nombre"] = clienteNombre
		item["fecha_operacion"] = fechaOperacion
		item["usuario_creador"] = usuarioCreador
		item["observaciones"] = observaciones
		items = append(items, item)

		totalCantidad += cantidad
		totalPorOperacion[tipoOperacion] += cantidad
		conteoPorOperacion[tipoOperacion]++
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "no se pudo consultar trazabilidad de lotes/series", http.StatusInternalServerError)
		return
	}

	resumenOps := make([]map[string]interface{}, 0, len(totalPorOperacion))
	for op, total := range totalPorOperacion {
		resumenOps = append(resumenOps, map[string]interface{}{
			"tipo_operacion": op,
			"movimientos":    conteoPorOperacion[op],
			"cantidad_total": reportesRound(total),
		})
	}

	var lote map[string]interface{}
	if id > 0 || codigo != "" {
		lote, _ = loadInventarioLoteSerieByIDOrCode(dbEmp, empresaID, id, codigo)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                true,
		"empresa_id":        empresaID,
		"lote":              lote,
		"total_movimientos": int64(len(items)),
		"cantidad_total":    reportesRound(totalCantidad),
		"resumen":           resumenOps,
		"items":             items,
	})
}

func handleComprasDevolucionProveedorContabilizarAction(dbEmp *sql.DB, w http.ResponseWriter, r *http.Request) {
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

	devolucionID := resolveIDFromPayloadOrQuery(payload, r)
	if devolucionID <= 0 {
		http.Error(w, "id es obligatorio", http.StatusBadRequest)
		return
	}

	item, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgDevProveedor.Table, empresaID, devolucionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "devolucion a proveedor no encontrada", http.StatusNotFound)
			return
		}
		http.Error(w, "no se pudo consultar devolucion a proveedor", http.StatusInternalServerError)
		return
	}

	if existingMov := anyToInt64(item["impacto_contable_movimiento_id"]); existingMov > 0 {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":                             true,
			"empresa_id":                     empresaID,
			"id":                             devolucionID,
			"ya_contabilizada":               true,
			"impacto_contable_movimiento_id": existingMov,
			"impacto_contable_evento_id":     anyToInt64(item["impacto_contable_evento_id"]),
			"item":                           item,
		})
		return
	}

	total := ventasAnyToFloat64(item["total"])
	impuesto := ventasAnyToFloat64(item["impuesto_total"])
	subtotal := ventasAnyToFloat64(item["subtotal"])
	if total <= 0 {
		total = subtotal + impuesto
	}
	if total <= 0 {
		http.Error(w, "la devolucion debe tener un total mayor a cero para contabilizar", http.StatusBadRequest)
		return
	}

	periodo := normalizePeriodoContableInput(finanzasFirstNonBlank(
		genericStringValue(payload["periodo_contable"]),
		genericStringValue(item["periodo_contable"]),
		genericStringValue(item["fecha_devolucion"]),
		time.Now().In(time.Local).Format("2006-01"),
	))
	if periodo == "" {
		periodo = time.Now().In(time.Local).Format("2006-01")
	}

	actor := strings.TrimSpace(adminEmailFromRequest(r))
	if actor == "" {
		actor = "sistema"
	}
	nowText := time.Now().In(time.Local).Format("2006-01-02 15:04:05")

	montoBase := total - impuesto
	if montoBase <= 0 {
		montoBase = subtotal
	}
	if montoBase <= 0 {
		montoBase = total
	}

	codigoDevolucion := genericStringValue(item["codigo"])
	movimiento, createErr := dbpkg.CreateEmpresaFinanzasMovimiento(dbEmp, dbpkg.EmpresaFinanzasMovimiento{
		EmpresaID:         empresaID,
		TipoMovimiento:    "ingreso",
		Codigo:            fmt.Sprintf("DPROV-MOV-%d", devolucionID),
		FechaMovimiento:   nowText,
		PeriodoContable:   periodo,
		Categoria:         "compras",
		Subcategoria:      "devolucion_proveedor",
		Concepto:          "Devolucion a proveedor " + codigoDevolucion,
		Descripcion:       finanzasFirstNonBlank(genericStringValue(item["motivo"]), "devolucion a proveedor contabilizada"),
		MetodoPago:        "ajuste_contable",
		Moneda:            finanzasFirstNonBlank(genericStringValue(item["moneda"]), "COP"),
		Monto:             montoBase,
		Impuesto:          impuesto,
		Total:             total,
		TotalNeto:         total,
		TerceroNombre:     genericStringValue(item["proveedor_nombre"]),
		TipoComprobante:   "nota_credito_proveedor",
		NumeroComprobante: finanzasFirstNonBlank(genericStringValue(item["documento_compra_codigo"]), codigoDevolucion),
		ReferenciaExterna: codigoDevolucion,
		AprobadoPor:       actor,
		UsuarioCreador:    actor,
		Estado:            "activo",
		Observaciones:     "impacto contable generado desde devoluciones proveedor",
	})
	if createErr != nil {
		if errors.Is(createErr, dbpkg.ErrPeriodoFinancieroCerrado) {
			http.Error(w, "el periodo contable de la devolucion esta cerrado", http.StatusConflict)
			return
		}
		http.Error(w, "no se pudo crear movimiento financiero de devolucion", http.StatusBadRequest)
		return
	}

	if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmp); err != nil {
		http.Error(w, "no se pudo preparar esquema de eventos contables", http.StatusInternalServerError)
		return
	}

	eventoPayload, _ := json.Marshal(map[string]interface{}{
		"devolucion_proveedor_id": devolucionID,
		"movimiento_finanzas_id":  movimiento,
		"codigo_devolucion":       codigoDevolucion,
		"periodo_contable":        periodo,
		"total":                   total,
	})

	eventoID, err := dbpkg.CreateEmpresaEventoContable(dbEmp, dbpkg.EmpresaEventoContable{
		EmpresaID:       empresaID,
		Modulo:          "compras",
		Evento:          "devolucion_proveedor_contabilizada",
		Entidad:         "devolucion_proveedor",
		EntidadID:       devolucionID,
		DocumentoTipo:   "devolucion_proveedor",
		DocumentoCodigo: codigoDevolucion,
		PeriodoContable: periodo,
		MontoTotal:      total,
		Moneda:          finanzasFirstNonBlank(genericStringValue(item["moneda"]), "COP"),
		PayloadJSON:     string(eventoPayload),
		Origen:          "api_devoluciones_proveedor",
		UsuarioCreador:  actor,
		Estado:          "activo",
		Observaciones:   "evento contable de devolucion a proveedor",
	})
	if err != nil {
		http.Error(w, "no se pudo registrar evento contable de devolucion", http.StatusInternalServerError)
		return
	}

	procesarAsientos := !strings.EqualFold(genericStringValue(payload["procesar_asientos"]), "false")
	resumenAsientos := map[string]interface{}{}
	if procesarAsientos {
		resultado, procErr := dbpkg.ProcessEmpresaEventosContablesPendientesConPolitica(dbEmp, empresaID, actor, 20, 0)
		if procErr == nil {
			resumenAsientos = map[string]interface{}{
				"eventos_revisados":   resultado.EventosRevisados,
				"eventos_procesados":  resultado.EventosProcesados,
				"asientos_creados":    resultado.AsientosCreados,
				"asientos_existentes": resultado.AsientosExistentes,
				"fallidos":            resultado.Fallidos,
				"errores":             resultado.Errores,
			}
		}
	}

	updatePayload := map[string]interface{}{
		"periodo_contable":               periodo,
		"impacto_contable_movimiento_id": movimiento,
		"impacto_contable_evento_id":     eventoID,
		"fecha_contabilizacion":          nowText,
		"contabilizado_por":              actor,
		"total_reintegrado":              total,
		"estado_devolucion":              "contabilizada",
		"observaciones": appendGenericObservation(
			genericStringValue(item["observaciones"]),
			fmt.Sprintf("impacto contable generado movimiento=%d evento=%d periodo=%s", movimiento, eventoID, periodo),
		),
	}
	if err := dbpkg.UpdateEmpresaGenericRow(dbEmp, cfgDevProveedor.Table, empresaID, devolucionID, updatePayload, cfgDevProveedor.AllowedColumns); err != nil {
		http.Error(w, "no se pudo actualizar devolucion con impacto contable", http.StatusInternalServerError)
		return
	}

	updated, _ := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgDevProveedor.Table, empresaID, devolucionID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                             true,
		"empresa_id":                     empresaID,
		"id":                             devolucionID,
		"impacto_contable_movimiento_id": movimiento,
		"impacto_contable_evento_id":     eventoID,
		"periodo_contable":               periodo,
		"procesamiento_asientos":         resumenAsientos,
		"item":                           updated,
	})
}

func loadInventarioLoteSerieByIDOrCode(dbEmp *sql.DB, empresaID, id int64, codigo string) (map[string]interface{}, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id invalido")
	}
	if id <= 0 && strings.TrimSpace(codigo) == "" {
		return nil, fmt.Errorf("id o codigo_lote_serie requerido")
	}

	query := `SELECT
		id,
		COALESCE(producto_id, 0),
		COALESCE(bodega_id, 0),
		COALESCE(tipo_control, ''),
		COALESCE(codigo_lote_serie, ''),
		COALESCE(fecha_fabricacion, ''),
		COALESCE(fecha_vencimiento, ''),
		COALESCE(cantidad_inicial, 0),
		COALESCE(cantidad_disponible, 0),
		COALESCE(reservado_cantidad, 0),
		COALESCE(vendido_cantidad, 0),
		COALESCE(costo_unitario, 0),
		COALESCE(estado_lote, ''),
		COALESCE(bloqueado_venta, 0),
		COALESCE(bloqueo_motivo, ''),
		COALESCE(ultima_operacion_tipo, ''),
		COALESCE(ultima_operacion_ref, ''),
		COALESCE(ultima_operacion_en, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM inventario_lotes_series
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if id > 0 {
		query += ` AND id = ?`
		args = append(args, id)
	} else {
		query += ` AND UPPER(COALESCE(codigo_lote_serie, '')) = UPPER(?)`
		args = append(args, strings.TrimSpace(codigo))
	}
	query += ` LIMIT 1`

	out := map[string]interface{}{}
	var outID int64
	var productoID int64
	var bodegaID int64
	var tipoControl string
	var codigoLote string
	var fechaFabricacion string
	var fechaVencimiento string
	var cantidadInicial float64
	var cantidadDisponible float64
	var reservadoCantidad float64
	var vendidoCantidad float64
	var costoUnitario float64
	var estadoLote string
	var bloqueadoVenta int64
	var bloqueoMotivo string
	var ultimaOperacionTipo string
	var ultimaOperacionRef string
	var ultimaOperacionEn string
	var estado string
	var observaciones string
	err := dbEmp.QueryRow(query, args...).Scan(
		&outID,
		&productoID,
		&bodegaID,
		&tipoControl,
		&codigoLote,
		&fechaFabricacion,
		&fechaVencimiento,
		&cantidadInicial,
		&cantidadDisponible,
		&reservadoCantidad,
		&vendidoCantidad,
		&costoUnitario,
		&estadoLote,
		&bloqueadoVenta,
		&bloqueoMotivo,
		&ultimaOperacionTipo,
		&ultimaOperacionRef,
		&ultimaOperacionEn,
		&estado,
		&observaciones,
	)
	if err != nil {
		return nil, err
	}

	out["id"] = outID
	out["producto_id"] = productoID
	out["bodega_id"] = bodegaID
	out["tipo_control"] = tipoControl
	out["codigo_lote_serie"] = codigoLote
	out["fecha_fabricacion"] = fechaFabricacion
	out["fecha_vencimiento"] = fechaVencimiento
	out["cantidad_inicial"] = cantidadInicial
	out["cantidad_disponible"] = cantidadDisponible
	out["reservado_cantidad"] = reservadoCantidad
	out["vendido_cantidad"] = vendidoCantidad
	out["costo_unitario"] = costoUnitario
	out["estado_lote"] = estadoLote
	out["bloqueado_venta"] = bloqueadoVenta
	out["bloqueo_motivo"] = bloqueoMotivo
	out["ultima_operacion_tipo"] = ultimaOperacionTipo
	out["ultima_operacion_ref"] = ultimaOperacionRef
	out["ultima_operacion_en"] = ultimaOperacionEn
	out["estado"] = estado
	out["observaciones"] = observaciones
	return out, nil
}

func ensureLoteSerieVencimientoBloqueo(dbEmp *sql.DB, empresaID int64, row map[string]interface{}, now time.Time, actor string) error {
	if row == nil {
		return nil
	}
	if !loteSerieEstaVencido(genericStringValue(row["fecha_vencimiento"]), now) {
		return nil
	}
	if strings.ToLower(strings.TrimSpace(genericStringValue(row["estado_lote"]))) == "vencido" && anyToInt64(row["bloqueado_venta"]) > 0 {
		return nil
	}
	if actor == "" {
		actor = "sistema"
	}
	nowText := now.Format("2006-01-02 15:04:05")
	update := map[string]interface{}{
		"estado_lote":           "vencido",
		"bloqueado_venta":       1,
		"bloqueo_motivo":        "lote/serie vencido: bloqueo automatico en venta y reserva",
		"ultima_operacion_tipo": "bloqueo_vencimiento",
		"ultima_operacion_ref":  "AUTO-VENCIMIENTO",
		"ultima_operacion_en":   nowText,
		"observaciones": appendGenericObservation(
			genericStringValue(row["observaciones"]),
			"bloqueo automatico por vencimiento para venta/reserva",
		),
	}
	return dbpkg.UpdateEmpresaGenericRow(dbEmp, cfgLotesSeries.Table, empresaID, anyToInt64(row["id"]), update, cfgLotesSeries.AllowedColumns)
}

func loteSerieEstaVencido(fechaVencimiento string, now time.Time) bool {
	fechaVencimiento = strings.TrimSpace(fechaVencimiento)
	if fechaVencimiento == "" {
		return false
	}
	parsed, ok := ventasParseDateTime(fechaVencimiento)
	if !ok {
		return false
	}
	vencimientoDate, _ := time.Parse("2006-01-02", parsed.In(time.Local).Format("2006-01-02"))
	nowDate, _ := time.Parse("2006-01-02", now.In(time.Local).Format("2006-01-02"))
	return vencimientoDate.Before(nowDate)
}

func normalizeLoteSerieOperacion(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	switch raw {
	case "operar":
		return ""
	case "reservar", "reserva":
		return "reserva"
	case "vender", "venta":
		return "venta"
	case "liberar_reserva", "liberar", "anular_reserva":
		return "liberar_reserva"
	case "ajuste_entrada", "entrada":
		return "ajuste_entrada"
	case "ajuste_salida", "salida":
		return "ajuste_salida"
	case "devolucion_proveedor", "devolucion":
		return "devolucion_proveedor"
	default:
		return ""
	}
}

func empresaFinanzasCarteraHandler(dbEmp *sql.DB, cfg empresaModuloGenericConfig, tipoMovimiento, terceroField, modulo string) http.HandlerFunc {
	base := empresaModuloGenericCRUDHandler(dbEmp, cfg)
	return func(w http.ResponseWriter, r *http.Request) {
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch action {
		case "conciliar_pagos", "conciliar":
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleConciliarCarteraPagosAction(dbEmp, cfg, tipoMovimiento, terceroField, modulo, w, r)
			return

		case "registrar_pago", "abonar", "registrar_abono":
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleRegistrarPagoCarteraAction(dbEmp, cfg, tipoMovimiento, terceroField, modulo, w, r)
			return

		case "validar_cierre_periodo":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleFinanzasValidarCierrePeriodoAction(dbEmp, w, r)
			return
		}

		base.ServeHTTP(w, r)
	}
}

func handleFinanzasValidarCierrePeriodoAction(dbEmp *sql.DB, w http.ResponseWriter, r *http.Request) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	periodo := normalizePeriodoContableInput(strings.TrimSpace(r.URL.Query().Get("periodo")))
	if periodo == "" {
		http.Error(w, "periodo es obligatorio (YYYY-MM)", http.StatusBadRequest)
		return
	}

	cerrado, err := dbpkg.IsEmpresaFinanzasPeriodoCerrado(dbEmp, empresaID, periodo)
	if err != nil {
		http.Error(w, "No se pudo validar el estado del periodo", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                         true,
		"empresa_id":                 empresaID,
		"periodo":                    periodo,
		"cerrado":                    cerrado,
		"bloqueo_retroactivo_activo": cerrado,
	})
}

func handlePlanCuentasPlantillasAction(w http.ResponseWriter, r *http.Request) {
	tipo := normalizePlanCuentaTipoEmpresa(strings.TrimSpace(r.URL.Query().Get("tipo_empresa")))
	tipos := []string{"general", "comercio", "restaurante", "hotel", "motel", "servicios"}

	if tipo == "" {
		resumen := make([]map[string]interface{}, 0, len(tipos))
		for _, key := range tipos {
			items := mergePlanCuentasTemplate(key)
			resumen = append(resumen, map[string]interface{}{
				"tipo_empresa": key,
				"cuentas":      len(items),
			})
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":                true,
			"tipos_disponibles": tipos,
			"resumen":           resumen,
		})
		return
	}

	items := mergePlanCuentasTemplate(tipo)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":           true,
		"tipo_empresa": tipo,
		"items":        items,
		"total":        len(items),
	})
}

func handlePlanCuentasAplicarPlantillaAction(dbEmp *sql.DB, w http.ResponseWriter, r *http.Request) {
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
	tipoEmpresa := normalizePlanCuentaTipoEmpresa(finanzasFirstNonBlank(genericStringValue(payload["tipo_empresa"]), strings.TrimSpace(r.URL.Query().Get("tipo_empresa"))))
	if tipoEmpresa == "" {
		tipoEmpresa = "general"
	}

	sobrescribir := parseTruthy(genericStringValue(payload["sobrescribir"])) || parseBoolQuery(r, "sobrescribir")
	actor := strings.TrimSpace(adminEmailFromRequest(r))
	if actor == "" {
		actor = "sistema"
	}

	result, err := applyPlanCuentasTemplate(dbEmp, empresaID, tipoEmpresa, sobrescribir, actor)
	if err != nil {
		http.Error(w, "No se pudo aplicar plantilla de plan de cuentas", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func applyPlanCuentasTemplate(dbEmp *sql.DB, empresaID int64, tipoEmpresa string, sobrescribir bool, actor string) (map[string]interface{}, error) {
	if err := dbpkg.EnsureEmpresaModulosFaltantesSchema(dbEmp); err != nil {
		return nil, err
	}

	items := mergePlanCuentasTemplate(tipoEmpresa)
	if len(items) == 0 {
		return nil, fmt.Errorf("plantilla sin cuentas")
	}

	created := int64(0)
	updated := int64(0)
	skipped := int64(0)

	for _, item := range items {
		if strings.TrimSpace(item.Codigo) == "" || strings.TrimSpace(item.Nombre) == "" {
			continue
		}

		payload := map[string]interface{}{
			"codigo":                 strings.TrimSpace(item.Codigo),
			"nombre":                 strings.TrimSpace(item.Nombre),
			"tipo_cuenta":            strings.TrimSpace(item.TipoCuenta),
			"naturaleza":             strings.TrimSpace(item.Naturaleza),
			"nivel":                  item.Nivel,
			"cuenta_padre_codigo":    strings.TrimSpace(item.CuentaPadre),
			"admite_movimiento":      boolToInt(item.AdmiteMovimiento),
			"aplica_impuesto":        boolToInt(item.AplicaImpuesto),
			"plantilla_tipo_empresa": tipoEmpresa,
			"plantilla_codigo":       strings.TrimSpace(item.Codigo),
			"plantilla_version":      "1",
			"cuenta_clave":           strings.TrimSpace(item.CuentaClave),
			"requerida":              boolToInt(item.Requerida),
			"orden_plantilla":        item.Orden,
			"usuario_creador":        actor,
			"estado":                 "activo",
			"observaciones":          strings.TrimSpace(item.Observaciones),
		}

		var existingID int64
		err := dbEmp.QueryRow(`SELECT id FROM empresa_plan_cuentas WHERE empresa_id = ? AND codigo = ? LIMIT 1`, empresaID, item.Codigo).Scan(&existingID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		if existingID > 0 {
			if !sobrescribir {
				skipped++
				continue
			}
			if err := dbpkg.UpdateEmpresaGenericRow(dbEmp, cfgPlanCuentas.Table, empresaID, existingID, payload, cfgPlanCuentas.AllowedColumns); err != nil {
				return nil, err
			}
			updated++
			continue
		}

		if _, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgPlanCuentas.Table, empresaID, payload, cfgPlanCuentas.AllowedColumns); err != nil {
			return nil, err
		}
		created++
	}

	return map[string]interface{}{
		"ok":                true,
		"empresa_id":        empresaID,
		"tipo_empresa":      tipoEmpresa,
		"sobrescribir":      sobrescribir,
		"cuentas_plantilla": len(items),
		"creadas":           created,
		"actualizadas":      updated,
		"omitidas":          skipped,
	}, nil
}

func planCuentasTemplatesByTipo() map[string][]planCuentaTemplateItem {
	return map[string][]planCuentaTemplateItem{
		"general": {
			{Codigo: "1105", Nombre: "Caja general", TipoCuenta: "activo", Naturaleza: "debito", Nivel: 1, AdmiteMovimiento: true, CuentaClave: "caja", Requerida: true, Orden: 10},
			{Codigo: "1110", Nombre: "Bancos", TipoCuenta: "activo", Naturaleza: "debito", Nivel: 1, AdmiteMovimiento: true, CuentaClave: "bancos", Requerida: true, Orden: 20},
			{Codigo: "1305", Nombre: "Clientes nacionales", TipoCuenta: "activo", Naturaleza: "debito", Nivel: 1, AdmiteMovimiento: true, CuentaClave: "clientes", Requerida: true, Orden: 30},
			{Codigo: "2205", Nombre: "Proveedores nacionales", TipoCuenta: "pasivo", Naturaleza: "credito", Nivel: 1, AdmiteMovimiento: true, CuentaClave: "proveedores", Requerida: true, Orden: 40},
			{Codigo: "2408", Nombre: "IVA por pagar", TipoCuenta: "pasivo", Naturaleza: "credito", Nivel: 1, AdmiteMovimiento: true, AplicaImpuesto: true, CuentaClave: "iva_por_pagar", Requerida: true, Orden: 50},
			{Codigo: "3105", Nombre: "Capital social", TipoCuenta: "patrimonio", Naturaleza: "credito", Nivel: 1, AdmiteMovimiento: true, CuentaClave: "capital", Requerida: true, Orden: 60},
			{Codigo: "4135", Nombre: "Ingresos operacionales", TipoCuenta: "ingreso", Naturaleza: "credito", Nivel: 1, AdmiteMovimiento: true, AplicaImpuesto: true, CuentaClave: "ingresos_operacionales", Requerida: true, Orden: 70},
			{Codigo: "5105", Nombre: "Costo de ventas", TipoCuenta: "gasto", Naturaleza: "debito", Nivel: 1, AdmiteMovimiento: true, CuentaClave: "costos", Requerida: true, Orden: 80},
			{Codigo: "5195", Nombre: "Gastos operacionales", TipoCuenta: "gasto", Naturaleza: "debito", Nivel: 1, AdmiteMovimiento: true, CuentaClave: "gastos", Requerida: true, Orden: 90},
		},
		"comercio": {
			{Codigo: "413510", Nombre: "Ventas mostrador comercio", TipoCuenta: "ingreso", Naturaleza: "credito", Nivel: 2, CuentaPadre: "4135", AdmiteMovimiento: true, AplicaImpuesto: true, CuentaClave: "ingresos_comercio", Requerida: true, Orden: 110},
			{Codigo: "510510", Nombre: "Costo mercancias comercio", TipoCuenta: "gasto", Naturaleza: "debito", Nivel: 2, CuentaPadre: "5105", AdmiteMovimiento: true, CuentaClave: "costos_comercio", Requerida: true, Orden: 120},
		},
		"restaurante": {
			{Codigo: "413520", Nombre: "Ventas restaurante", TipoCuenta: "ingreso", Naturaleza: "credito", Nivel: 2, CuentaPadre: "4135", AdmiteMovimiento: true, AplicaImpuesto: true, CuentaClave: "ingresos_restaurante", Requerida: true, Orden: 130},
			{Codigo: "413530", Nombre: "Propinas cobradas", TipoCuenta: "ingreso", Naturaleza: "credito", Nivel: 2, CuentaPadre: "4135", AdmiteMovimiento: true, CuentaClave: "propinas", Requerida: false, Orden: 140},
			{Codigo: "510520", Nombre: "Insumos de cocina", TipoCuenta: "gasto", Naturaleza: "debito", Nivel: 2, CuentaPadre: "5105", AdmiteMovimiento: true, CuentaClave: "insumos_cocina", Requerida: true, Orden: 150},
		},
		"hotel": {
			{Codigo: "413540", Nombre: "Ingresos hospedaje", TipoCuenta: "ingreso", Naturaleza: "credito", Nivel: 2, CuentaPadre: "4135", AdmiteMovimiento: true, AplicaImpuesto: true, CuentaClave: "ingresos_hospedaje", Requerida: true, Orden: 160},
			{Codigo: "413550", Nombre: "Ingresos minibar y extras", TipoCuenta: "ingreso", Naturaleza: "credito", Nivel: 2, CuentaPadre: "4135", AdmiteMovimiento: true, AplicaImpuesto: true, CuentaClave: "ingresos_extras", Requerida: false, Orden: 170},
			{Codigo: "510540", Nombre: "Amenidades y lenceria", TipoCuenta: "gasto", Naturaleza: "debito", Nivel: 2, CuentaPadre: "5105", AdmiteMovimiento: true, CuentaClave: "amenidades", Requerida: false, Orden: 180},
		},
		"motel": {
			{Codigo: "413545", Nombre: "Ingresos hospedaje corta estancia", TipoCuenta: "ingreso", Naturaleza: "credito", Nivel: 2, CuentaPadre: "4135", AdmiteMovimiento: true, AplicaImpuesto: true, CuentaClave: "ingresos_motel", Requerida: true, Orden: 190},
			{Codigo: "510545", Nombre: "Kits y aseo estacion", TipoCuenta: "gasto", Naturaleza: "debito", Nivel: 2, CuentaPadre: "5105", AdmiteMovimiento: true, CuentaClave: "kits_aseo", Requerida: false, Orden: 200},
		},
		"servicios": {
			{Codigo: "413560", Nombre: "Ingresos por servicios", TipoCuenta: "ingreso", Naturaleza: "credito", Nivel: 2, CuentaPadre: "4135", AdmiteMovimiento: true, AplicaImpuesto: true, CuentaClave: "ingresos_servicios", Requerida: true, Orden: 210},
			{Codigo: "510560", Nombre: "Costos directos de servicios", TipoCuenta: "gasto", Naturaleza: "debito", Nivel: 2, CuentaPadre: "5105", AdmiteMovimiento: true, CuentaClave: "costos_servicios", Requerida: true, Orden: 220},
		},
	}
}

func normalizePlanCuentaTipoEmpresa(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "comercio", "retail", "tienda":
		return "comercio"
	case "restaurante", "restaurant":
		return "restaurante"
	case "hotel", "hospedaje":
		return "hotel"
	case "motel":
		return "motel"
	case "servicios", "service":
		return "servicios"
	case "general", "":
		return "general"
	default:
		return "general"
	}
}

func mergePlanCuentasTemplate(tipoEmpresa string) []planCuentaTemplateItem {
	tipoEmpresa = normalizePlanCuentaTipoEmpresa(tipoEmpresa)
	templates := planCuentasTemplatesByTipo()

	base := templates["general"]
	out := make([]planCuentaTemplateItem, 0, len(base)+8)
	seen := make(map[string]bool)

	appendItems := func(items []planCuentaTemplateItem) {
		for _, item := range items {
			codigo := strings.TrimSpace(item.Codigo)
			if codigo == "" || seen[codigo] {
				continue
			}
			seen[codigo] = true
			out = append(out, item)
		}
	}

	appendItems(base)
	if tipoEmpresa != "general" {
		appendItems(templates[tipoEmpresa])
	}

	return out
}

func handleConciliarCarteraPagosAction(dbEmp *sql.DB, cfg empresaModuloGenericConfig, tipoMovimiento, terceroField, modulo string, w http.ResponseWriter, r *http.Request) {
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

	limit, err := parseIntQueryOptional(r, "limit")
	if err != nil {
		http.Error(w, "limit invalido", http.StatusBadRequest)
		return
	}
	if limit <= 0 {
		limit = 500
	}
	if limit > 2000 {
		limit = 2000
	}

	periodoFiltro := normalizePeriodoContableInput(finanzasFirstNonBlank(genericStringValue(payload["periodo_contable"]), strings.TrimSpace(r.URL.Query().Get("periodo"))))
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	includeInactive := parseBoolQuery(r, "include_inactive")
	actor := strings.TrimSpace(adminEmailFromRequest(r))
	if actor == "" {
		actor = "sistema"
	}
	now := time.Now().In(time.Local).Format("2006-01-02 15:04:05")

	rows, err := dbpkg.ListEmpresaGenericRows(dbEmp, cfg.Table, empresaID, dbpkg.EmpresaGenericListFilter{
		IncludeInactive: includeInactive,
		Q:               q,
		Limit:           limit,
		Offset:          0,
		SearchColumns:   cfg.SearchColumns,
	})
	if err != nil {
		http.Error(w, "No se pudo consultar cartera para conciliacion", http.StatusInternalServerError)
		return
	}

	conciliados := int64(0)
	bloqueadosPeriodo := int64(0)
	errores := int64(0)
	considerados := int64(0)
	totalOriginal := 0.0
	totalPagado := 0.0
	totalSaldo := 0.0
	detalles := make([]map[string]interface{}, 0)

	for _, row := range rows {
		id := anyToInt64(row["id"])
		if id <= 0 {
			continue
		}

		periodoRow := normalizePeriodoContableInput(finanzasFirstNonBlank(genericStringValue(row["periodo_contable"]), genericStringValue(row["fecha_emision"])))
		if periodoRow == "" {
			periodoRow = finanzasFirstNonBlank(periodoFiltro, time.Now().In(time.Local).Format("2006-01"))
		}
		if periodoFiltro != "" && periodoRow != periodoFiltro {
			continue
		}

		considerados++

		cerrado, err := dbpkg.IsEmpresaFinanzasPeriodoCerrado(dbEmp, empresaID, periodoRow)
		if err != nil {
			errores++
			detalles = append(detalles, map[string]interface{}{
				"id":               id,
				"codigo":           genericStringValue(row["codigo"]),
				"periodo_contable": periodoRow,
				"resultado":        "error",
				"detalle":          "no se pudo validar estado del periodo",
			})
			continue
		}
		if cerrado {
			bloqueadosPeriodo++
			detalles = append(detalles, map[string]interface{}{
				"id":               id,
				"codigo":           genericStringValue(row["codigo"]),
				"periodo_contable": periodoRow,
				"resultado":        "bloqueado",
				"detalle":          "periodo contable cerrado",
			})
			continue
		}

		documentoCodigo := genericStringValue(row["documento_codigo"])
		terceroNombre := genericStringValue(row[terceroField])
		pagosRelacionados, montoPagado, fechaUltimoPago, err := loadPagosCarteraRelacionados(dbEmp, empresaID, tipoMovimiento, periodoRow, documentoCodigo, terceroNombre)
		if err != nil {
			errores++
			detalles = append(detalles, map[string]interface{}{
				"id":               id,
				"codigo":           genericStringValue(row["codigo"]),
				"periodo_contable": periodoRow,
				"resultado":        "error",
				"detalle":          "no se pudieron consultar pagos reales",
			})
			continue
		}

		valorOriginal := ventasAnyToFloat64(row["valor_original"])
		if valorOriginal < 0 {
			valorOriginal = 0
		}
		if valorOriginal <= 0 {
			valorOriginal = ventasAnyToFloat64(row["saldo"]) + montoPagado
		}
		if montoPagado > valorOriginal && valorOriginal > 0 {
			montoPagado = valorOriginal
		}
		if montoPagado < 0 {
			montoPagado = 0
		}

		saldo := valorOriginal - montoPagado
		if saldo < 0 {
			saldo = 0
		}
		estadoNuevo := finanzasCarteraEstado(saldo, montoPagado, genericStringValue(row["fecha_vencimiento"]))
		diasMora := finanzasCarteraDiasMora(genericStringValue(row["fecha_vencimiento"]), saldo)

		referenciaPagosJSON := "[]"
		if len(pagosRelacionados) > 0 {
			encoded, _ := json.Marshal(pagosRelacionados)
			referenciaPagosJSON = string(encoded)
		}

		update := map[string]interface{}{
			"valor_original":        valorOriginal,
			"valor_pagado":          montoPagado,
			"saldo":                 saldo,
			"estado_cartera":        estadoNuevo,
			"dias_mora":             diasMora,
			"periodo_contable":      periodoRow,
			"referencia_pagos_json": referenciaPagosJSON,
			"fecha_ultimo_pago":     fechaUltimoPago,
			"conciliado_en":         now,
			"conciliado_por":        actor,
			"observaciones":         appendGenericObservation(genericStringValue(row["observaciones"]), "conciliacion automatica contra pagos reales"),
		}

		if err := dbpkg.UpdateEmpresaGenericRow(dbEmp, cfg.Table, empresaID, id, update, cfg.AllowedColumns); err != nil {
			if errors.Is(err, dbpkg.ErrPeriodoFinancieroCerrado) {
				bloqueadosPeriodo++
				detalles = append(detalles, map[string]interface{}{
					"id":               id,
					"codigo":           genericStringValue(row["codigo"]),
					"periodo_contable": periodoRow,
					"resultado":        "bloqueado",
					"detalle":          "periodo contable cerrado",
				})
				continue
			}
			errores++
			detalles = append(detalles, map[string]interface{}{
				"id":               id,
				"codigo":           genericStringValue(row["codigo"]),
				"periodo_contable": periodoRow,
				"resultado":        "error",
				"detalle":          "no se pudo actualizar cartera conciliada",
			})
			continue
		}

		conciliados++
		totalOriginal += valorOriginal
		totalPagado += montoPagado
		totalSaldo += saldo
		detalles = append(detalles, map[string]interface{}{
			"id":                 id,
			"codigo":             genericStringValue(row["codigo"]),
			"documento_codigo":   documentoCodigo,
			"periodo_contable":   periodoRow,
			"valor_original":     reportesRound(valorOriginal),
			"valor_pagado":       reportesRound(montoPagado),
			"saldo":              reportesRound(saldo),
			"estado_cartera":     estadoNuevo,
			"pagos_relacionados": len(pagosRelacionados),
			"resultado":          "conciliado",
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                     true,
		"empresa_id":             empresaID,
		"modulo":                 modulo,
		"tipo_movimiento":        tipoMovimiento,
		"periodo_contable":       periodoFiltro,
		"registros_consultados":  int64(len(rows)),
		"registros_considerados": considerados,
		"conciliados":            conciliados,
		"bloqueados_periodo":     bloqueadosPeriodo,
		"errores":                errores,
		"total_original":         reportesRound(totalOriginal),
		"total_pagado":           reportesRound(totalPagado),
		"total_saldo":            reportesRound(totalSaldo),
		"items":                  detalles,
	})
}

func handleRegistrarPagoCarteraAction(dbEmp *sql.DB, cfg empresaModuloGenericConfig, tipoMovimiento, terceroField, modulo string, w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "id es obligatorio", http.StatusBadRequest)
		return
	}

	item, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfg.Table, empresaID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "registro de cartera no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo consultar cartera", http.StatusInternalServerError)
		return
	}

	periodo := normalizePeriodoContableInput(finanzasFirstNonBlank(genericStringValue(item["periodo_contable"]), genericStringValue(item["fecha_emision"])))
	if raw := normalizePeriodoContableInput(finanzasFirstNonBlank(genericStringValue(payload["periodo_contable"]), strings.TrimSpace(r.URL.Query().Get("periodo")))); raw != "" {
		periodo = raw
	}
	if periodo == "" {
		periodo = time.Now().In(time.Local).Format("2006-01")
	}
	cerrado, err := dbpkg.IsEmpresaFinanzasPeriodoCerrado(dbEmp, empresaID, periodo)
	if err != nil {
		http.Error(w, "No se pudo validar el periodo contable", http.StatusInternalServerError)
		return
	}
	if cerrado {
		http.Error(w, "el periodo contable del registro esta cerrado", http.StatusConflict)
		return
	}

	abono := ventasAnyToFloat64(payload["monto"])
	if abono <= 0 {
		abono = ventasAnyToFloat64(payload["valor_pagado"])
	}
	if abono <= 0 {
		http.Error(w, "monto del abono debe ser mayor que cero", http.StatusBadRequest)
		return
	}

	valorOriginal := ventasAnyToFloat64(item["valor_original"])
	if valorOriginal <= 0 {
		valorOriginal = ventasAnyToFloat64(item["saldo"]) + ventasAnyToFloat64(item["valor_pagado"])
	}
	valorPagadoActual := ventasAnyToFloat64(item["valor_pagado"])
	saldoActual := ventasAnyToFloat64(item["saldo"])
	if saldoActual <= 0 && valorOriginal > 0 {
		saldoActual = valorOriginal - valorPagadoActual
	}
	if saldoActual < 0 {
		saldoActual = 0
	}
	if abono > saldoActual && saldoActual > 0 {
		abono = saldoActual
	}
	if abono <= 0 {
		http.Error(w, "la cartera ya no tiene saldo por aplicar", http.StatusBadRequest)
		return
	}

	actor := strings.TrimSpace(adminEmailFromRequest(r))
	if actor == "" {
		actor = "sistema"
	}
	now := time.Now().In(time.Local).Format("2006-01-02 15:04:05")
	documentoCodigo := genericStringValue(item["documento_codigo"])
	terceroNombre := genericStringValue(item[terceroField])
	metodoPago := finanzasFirstNonBlank(genericStringValue(payload["metodo_pago"]), strings.TrimSpace(r.URL.Query().Get("metodo_pago")), "efectivo")
	referenciaExterna := finanzasFirstNonBlank(genericStringValue(payload["referencia_externa"]), documentoCodigo, genericStringValue(item["codigo"]))
	concepto := finanzasFirstNonBlank(genericStringValue(payload["concepto"]), "Abono cartera "+genericStringValue(item["codigo"]))
	moneda := finanzasFirstNonBlank(genericStringValue(payload["moneda"]), genericStringValue(item["moneda"]), "COP")

	movID, err := dbpkg.CreateEmpresaFinanzasMovimiento(dbEmp, dbpkg.EmpresaFinanzasMovimiento{
		EmpresaID:         empresaID,
		TipoMovimiento:    tipoMovimiento,
		PeriodoContable:   periodo,
		FechaMovimiento:   now,
		Categoria:         modulo,
		Subcategoria:      "abono_cartera",
		Concepto:          concepto,
		Descripcion:       "Abono aplicado a " + modulo + " " + genericStringValue(item["codigo"]),
		MetodoPago:        metodoPago,
		Moneda:            moneda,
		Monto:             abono,
		Total:             abono,
		TotalNeto:         abono,
		TerceroNombre:     terceroNombre,
		TerceroDocumento:  "",
		TipoComprobante:   "recibo_interno",
		NumeroComprobante: finanzasFirstNonBlank(genericStringValue(payload["numero_comprobante"]), referenciaExterna),
		ReferenciaExterna: referenciaExterna,
		UsuarioCreador:    actor,
		Estado:            "activo",
		Observaciones:     appendGenericObservation(genericStringValue(payload["observaciones"]), "pago aplicado desde cartera "+modulo),
	})
	if err != nil {
		if errors.Is(err, dbpkg.ErrPeriodoFinancieroCerrado) {
			http.Error(w, "el periodo contable del movimiento esta cerrado", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	valorPagadoNuevo := valorPagadoActual + abono
	if valorOriginal > 0 && valorPagadoNuevo > valorOriginal {
		valorPagadoNuevo = valorOriginal
	}
	saldoNuevo := valorOriginal - valorPagadoNuevo
	if saldoNuevo < 0 {
		saldoNuevo = 0
	}
	pagosRelacionados, _, _, _ := loadPagosCarteraRelacionados(dbEmp, empresaID, tipoMovimiento, periodo, documentoCodigo, terceroNombre)
	pagosRelacionados = dedupePagosCarteraRelacionados(pagosRelacionados)
	referenciaPagosJSON := "[]"
	if len(pagosRelacionados) > 0 {
		encoded, _ := json.Marshal(pagosRelacionados)
		referenciaPagosJSON = string(encoded)
	}

	update := map[string]interface{}{
		"valor_original":        valorOriginal,
		"valor_pagado":          valorPagadoNuevo,
		"saldo":                 saldoNuevo,
		"estado_cartera":        finanzasCarteraEstado(saldoNuevo, valorPagadoNuevo, genericStringValue(item["fecha_vencimiento"])),
		"dias_mora":             finanzasCarteraDiasMora(genericStringValue(item["fecha_vencimiento"]), saldoNuevo),
		"periodo_contable":      periodo,
		"referencia_pagos_json": referenciaPagosJSON,
		"fecha_ultimo_pago":     now,
		"conciliado_en":         now,
		"conciliado_por":        actor,
		"observaciones":         appendGenericObservation(genericStringValue(item["observaciones"]), "abono registrado: "+strconv.FormatFloat(abono, 'f', 2, 64)),
	}
	if err := dbpkg.UpdateEmpresaGenericRow(dbEmp, cfg.Table, empresaID, id, update, cfg.AllowedColumns); err != nil {
		if errors.Is(err, dbpkg.ErrPeriodoFinancieroCerrado) {
			http.Error(w, "el periodo contable del registro esta cerrado", http.StatusConflict)
			return
		}
		http.Error(w, "No se pudo actualizar cartera con el abono", http.StatusInternalServerError)
		return
	}
	itemActualizado, _ := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfg.Table, empresaID, id)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                     true,
		"empresa_id":             empresaID,
		"modulo":                 modulo,
		"cartera_id":             id,
		"movimiento_finanzas_id": movID,
		"monto_aplicado":         reportesRound(abono),
		"saldo":                  reportesRound(saldoNuevo),
		"estado_cartera":         update["estado_cartera"],
		"item":                   itemActualizado,
	})
}

func loadPagosCarteraRelacionados(dbEmp *sql.DB, empresaID int64, tipoMovimiento, periodo, documentoCodigo, terceroNombre string) ([]carteraPagoRelacionado, float64, string, error) {
	documentoCodigo = strings.ToUpper(strings.TrimSpace(documentoCodigo))
	terceroNombre = strings.ToUpper(strings.TrimSpace(terceroNombre))
	if documentoCodigo == "" && terceroNombre == "" {
		return []carteraPagoRelacionado{}, 0, "", nil
	}

	query := `SELECT
		id,
		COALESCE(codigo, ''),
		COALESCE(total_neto, 0),
		COALESCE(total, 0),
		COALESCE(monto, 0),
		COALESCE(fecha_movimiento, ''),
		COALESCE(referencia_externa, ''),
		COALESCE(numero_comprobante, '')
	FROM empresa_finanzas_movimientos
	WHERE empresa_id = ?
	  AND LOWER(COALESCE(estado, 'activo')) = 'activo'
	  AND LOWER(COALESCE(tipo_movimiento, '')) = ?`
	args := []interface{}{empresaID, strings.ToLower(strings.TrimSpace(tipoMovimiento))}

	if periodo != "" {
		query += ` AND COALESCE(periodo_contable, '') = ?`
		args = append(args, periodo)
	}

	if documentoCodigo != "" {
		pattern := finanzasLikePattern(documentoCodigo)
		query += ` AND (
			UPPER(COALESCE(referencia_externa, '')) = ?
			OR UPPER(COALESCE(numero_comprobante, '')) = ?
			OR UPPER(COALESCE(concepto, '')) LIKE ? ESCAPE '!'
			OR UPPER(COALESCE(descripcion, '')) LIKE ? ESCAPE '!'
		)`
		args = append(args, documentoCodigo, documentoCodigo, pattern, pattern)
	} else {
		pattern := finanzasLikePattern(terceroNombre)
		query += ` AND (
			UPPER(COALESCE(tercero_nombre, '')) = ?
			OR UPPER(COALESCE(tercero_nombre, '')) LIKE ? ESCAPE '!'
		)`
		args = append(args, terceroNombre, pattern)
	}

	query += ` ORDER BY datetime(COALESCE(NULLIF(fecha_movimiento, ''), fecha_creacion)) DESC, id DESC LIMIT 500`

	rows, err := dbEmp.Query(query, args...)
	if err != nil {
		return nil, 0, "", err
	}
	defer rows.Close()

	out := make([]carteraPagoRelacionado, 0)
	total := 0.0
	ultimoPago := ""
	ultimoPagoAt := time.Time{}

	for rows.Next() {
		var item carteraPagoRelacionado
		var totalNeto float64
		var totalBruto float64
		var montoBase float64
		if err := rows.Scan(
			&item.MovimientoID,
			&item.Codigo,
			&totalNeto,
			&totalBruto,
			&montoBase,
			&item.FechaMovimiento,
			&item.ReferenciaExterna,
			&item.NumeroComprobante,
		); err != nil {
			return nil, 0, "", err
		}

		monto := totalNeto
		if monto <= 0 {
			monto = totalBruto
		}
		if monto <= 0 {
			monto = montoBase
		}
		if monto <= 0 {
			continue
		}
		item.MontoAplicado = monto
		total += monto
		out = append(out, item)

		if parsed, ok := ventasParseDateTime(item.FechaMovimiento); ok {
			if ultimoPago == "" || parsed.After(ultimoPagoAt) {
				ultimoPagoAt = parsed
				ultimoPago = parsed.Format("2006-01-02 15:04:05")
			}
		} else if ultimoPago == "" {
			ultimoPago = strings.TrimSpace(item.FechaMovimiento)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, 0, "", err
	}

	out = dedupePagosCarteraRelacionados(out)
	total = 0
	for _, item := range out {
		total += item.MontoAplicado
	}
	return out, total, ultimoPago, nil
}

func finanzasLikePattern(raw string) string {
	raw = strings.TrimSpace(strings.ToUpper(raw))
	raw = strings.ReplaceAll(raw, "!", "!!")
	raw = strings.ReplaceAll(raw, "%", "!%")
	raw = strings.ReplaceAll(raw, "_", "!_")
	if raw == "" {
		return "%"
	}
	return "%" + raw + "%"
}

func normalizePeriodoContableInput(v string) string {
	v = strings.TrimSpace(strings.ReplaceAll(v, "/", "-"))
	if v == "" {
		return ""
	}
	if len(v) >= 7 {
		candidate := v[:7]
		if _, err := time.Parse("2006-01", candidate); err == nil {
			return candidate
		}
	}
	if parsed, ok := ventasParseDateTime(v); ok {
		return parsed.Format("2006-01")
	}
	return ""
}

func finanzasCarteraEstado(saldo, valorPagado float64, fechaVencimiento string) string {
	if saldo <= 0.009 {
		return "pagada"
	}
	if valorPagado > 0 {
		return "parcial"
	}
	if dueDate := strings.TrimSpace(fechaVencimiento); dueDate != "" {
		if finanzasCarteraDiasMora(dueDate, saldo) > 0 {
			return "vencida"
		}
	}
	return "pendiente"
}

func finanzasCarteraDiasMora(fechaVencimiento string, saldo float64) int64 {
	if saldo <= 0.009 {
		return 0
	}
	dueDate := strings.TrimSpace(fechaVencimiento)
	if dueDate == "" {
		return 0
	}
	parsedDue, ok := ventasParseDateTime(dueDate)
	if !ok {
		return 0
	}
	now := time.Now().In(time.Local)
	if !now.After(parsedDue) {
		return 0
	}
	days := int64(now.Sub(parsedDue).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

func finanzasFirstNonBlank(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func boolToInt(v bool) int64 {
	if v {
		return 1
	}
	return 0
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

func EmpresaProduccionOrdenesHandler(dbEmp *sql.DB) http.HandlerFunc {
	base := empresaModuloGenericCRUDHandler(dbEmp, cfgProduccionOrdenes)
	return func(w http.ResponseWriter, r *http.Request) {
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch action {
		case "plan_capacidad", "capacidad":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleProduccionOrdenesPlanCapacidadAction(dbEmp, w, r)
			return
		}

		base.ServeHTTP(w, r)
	}
}

func EmpresaLogisticaEnviosHandler(dbEmp *sql.DB) http.HandlerFunc {
	base := empresaModuloGenericCRUDHandler(dbEmp, cfgLogisticaEnvios)
	return func(w http.ResponseWriter, r *http.Request) {
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch action {
		case "seguimiento_hitos", "hitos", "alertas_incumplimiento":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleLogisticaEnviosSeguimientoHitosAction(dbEmp, w, r)
			return
		}

		base.ServeHTTP(w, r)
	}
}

func EmpresaDocumentosGestionHandler(dbEmp *sql.DB) http.HandlerFunc {
	base := empresaModuloGenericCRUDHandler(dbEmp, cfgDocumentosGestion)
	return func(w http.ResponseWriter, r *http.Request) {
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch action {
		case "acceso":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleDocumentosGestionAccesoAction(dbEmp, w, r)
			return

		case "repositorio", "repository":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleDocumentosGestionRepositorioAction(dbEmp, w, r)
			return

		case "versiones", "historial_versiones":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleDocumentosGestionVersionesAction(dbEmp, w, r)
			return

		case "versionar":
			if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleDocumentosGestionVersionarAction(dbEmp, w, r)
			return
		}

		base.ServeHTTP(w, r)
	}
}

func EmpresaDocumentosFirmasHandler(dbEmp *sql.DB) http.HandlerFunc {
	base := empresaModuloGenericCRUDHandler(dbEmp, cfgDocumentosFirmas)
	return func(w http.ResponseWriter, r *http.Request) {
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "acceso" {
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleDocumentosFirmasAccesoAction(dbEmp, w, r)
			return
		}

		base.ServeHTTP(w, r)
	}
}

func handleDocumentosGestionAccesoAction(dbEmp *sql.DB, w http.ResponseWriter, r *http.Request) {
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

	requestedAction := resolveDocumentoPermissionActionFromRequest(r)
	role := normalizePermissionRole(adminRoleFromRequest(r))

	if id > 0 {
		row, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgDocumentosGestion.Table, empresaID, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "registro no encontrado", http.StatusNotFound)
				return
			}
			http.Error(w, "No se pudo validar acceso del documento", http.StatusInternalServerError)
			return
		}

		allowed, normalizedRole, permissionModule := evaluateDocumentoGestionAccess(r, row, requestedAction)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":               true,
			"empresa_id":       empresaID,
			"id":               id,
			"modulo_documento": genericStringValue(row["modulo"]),
			"modulo_permiso":   permissionModule,
			"accion_requerida": requestedAction,
			"rol":              normalizedRole,
			"acceso_permitido": allowed,
			"documento_codigo": genericStringValue(row["documento_codigo"]),
			"estado_documento": genericStringValue(row["estado_documento"]),
			"estado_registro":  genericStringDefault(row["estado"], "activo"),
		})
		return
	}

	moduloDocumento := strings.TrimSpace(r.URL.Query().Get("modulo"))
	if moduloDocumento == "" {
		http.Error(w, "id o modulo required", http.StatusBadRequest)
		return
	}

	permissionModule := mapDocumentoModuloToPermissionModule(moduloDocumento)
	allowed := true
	if role != "" {
		allowed = roleAllowsModuleAction(role, permissionModule, requestedAction)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":               true,
		"empresa_id":       empresaID,
		"modulo_documento": moduloDocumento,
		"modulo_permiso":   permissionModule,
		"accion_requerida": requestedAction,
		"rol":              role,
		"acceso_permitido": allowed,
	})
}

func handleDocumentosFirmasAccesoAction(dbEmp *sql.DB, w http.ResponseWriter, r *http.Request) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := parseInt64QueryOptional(r, "id")
	if err != nil || id <= 0 {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	firmaRow, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgDocumentosFirmas.Table, empresaID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "registro no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo consultar firma documental", http.StatusInternalServerError)
		return
	}

	requestedAction := resolveDocumentoPermissionActionFromRequest(r)
	role := normalizePermissionRole(adminRoleFromRequest(r))
	documentoID := anyToInt64(firmaRow["documento_gestion_id"])
	permissionModule := permModuleSeguridad
	moduloDocumento := "documentos"
	allowed := true

	if documentoID > 0 {
		docRow, docErr := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgDocumentosGestion.Table, empresaID, documentoID)
		if docErr == nil {
			moduloDocumento = genericStringValue(docRow["modulo"])
			allowed, role, permissionModule = evaluateDocumentoGestionAccess(r, docRow, requestedAction)
		} else if role != "" {
			allowed = roleAllowsModuleAction(role, permissionModule, requestedAction)
		}
	} else if role != "" {
		allowed = roleAllowsModuleAction(role, permissionModule, requestedAction)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":               true,
		"empresa_id":       empresaID,
		"id":               id,
		"documento_id":     documentoID,
		"modulo_documento": moduloDocumento,
		"modulo_permiso":   permissionModule,
		"accion_requerida": requestedAction,
		"rol":              role,
		"acceso_permitido": allowed,
	})
}

func handleDocumentosGestionRepositorioAction(dbEmp *sql.DB, w http.ResponseWriter, r *http.Request) {
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

	includeInactive := parseBoolQuery(r, "include_inactive")
	includeDenied := parseBoolQuery(r, "include_denegados")
	requestedAction := resolveDocumentoPermissionActionFromRequest(r)

	rows, err := loadEmpresaRowsForAction(dbEmp, cfgDocumentosGestion, empresaID, id, includeInactive, strings.TrimSpace(r.URL.Query().Get("q")), limit, offset)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "registro no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo consultar repositorio documental", http.StatusInternalServerError)
		return
	}

	items := make([]map[string]interface{}, 0, len(rows))
	denegados := int64(0)

	for _, row := range rows {
		allowed, role, permissionModule := evaluateDocumentoGestionAccess(r, row, requestedAction)
		if !allowed {
			denegados++
			if !includeDenied {
				continue
			}
		}

		item := map[string]interface{}{}
		for k, v := range row {
			item[k] = v
		}
		item["modulo_permiso"] = permissionModule
		item["rol"] = role
		item["accion_requerida"] = requestedAction
		item["acceso_permitido"] = allowed
		items = append(items, item)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                true,
		"empresa_id":        empresaID,
		"modulo":            "documentos_gestion",
		"accion_requerida":  requestedAction,
		"total_consultados": int64(len(rows)),
		"visibles":          int64(len(items)),
		"denegados":         denegados,
		"items":             items,
	})
}

func handleDocumentosGestionVersionesAction(dbEmp *sql.DB, w http.ResponseWriter, r *http.Request) {
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

	documentoCodigo := strings.TrimSpace(r.URL.Query().Get("documento_codigo"))
	if id > 0 {
		row, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgDocumentosGestion.Table, empresaID, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "registro no encontrado", http.StatusNotFound)
				return
			}
			http.Error(w, "No se pudo consultar documento base", http.StatusInternalServerError)
			return
		}
		documentoCodigo = finanzasFirstNonBlank(documentoCodigo, genericStringValue(row["documento_codigo"]), genericStringValue(row["codigo"]))
	}

	if strings.TrimSpace(documentoCodigo) == "" {
		http.Error(w, "documento_codigo required", http.StatusBadRequest)
		return
	}

	limit, err := parseIntQueryOptional(r, "limit")
	if err != nil {
		http.Error(w, "limit invalido", http.StatusBadRequest)
		return
	}
	includeInactive := parseBoolQuery(r, "include_inactive")
	includeDenied := parseBoolQuery(r, "include_denegados")
	requestedAction := resolveDocumentoPermissionActionFromRequest(r)

	rows, err := loadDocumentoGestionVersionRows(dbEmp, empresaID, documentoCodigo, includeInactive, limit)
	if err != nil {
		http.Error(w, "No se pudo consultar historial de versiones", http.StatusInternalServerError)
		return
	}

	items := make([]map[string]interface{}, 0, len(rows))
	denegados := int64(0)

	for _, row := range rows {
		allowed, role, permissionModule := evaluateDocumentoGestionAccess(r, row, requestedAction)
		if !allowed {
			denegados++
			if !includeDenied {
				continue
			}
		}

		item := map[string]interface{}{}
		for k, v := range row {
			item[k] = v
		}
		item["modulo_permiso"] = permissionModule
		item["rol"] = role
		item["accion_requerida"] = requestedAction
		item["acceso_permitido"] = allowed
		items = append(items, item)
	}

	versionActual := int64(0)
	if len(rows) > 0 {
		versionActual = parseDocumentoVersionInt(rows[0]["version"])
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":               true,
		"empresa_id":       empresaID,
		"documento_codigo": documentoCodigo,
		"accion_requerida": requestedAction,
		"version_actual":   versionActual,
		"total_versiones":  int64(len(rows)),
		"visibles":         int64(len(items)),
		"denegados":        denegados,
		"items":            items,
	})
}

func handleDocumentosGestionVersionarAction(dbEmp *sql.DB, w http.ResponseWriter, r *http.Request) {
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

	baseRow, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgDocumentosGestion.Table, empresaID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "registro no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo consultar documento a versionar", http.StatusInternalServerError)
		return
	}

	allowed, role, permissionModule := evaluateDocumentoGestionAccess(r, baseRow, permActionUpdate)
	if !allowed {
		http.Error(w, "forbidden: el rol no tiene permisos para versionar este documento", http.StatusForbidden)
		return
	}

	documentoCodigo := finanzasFirstNonBlank(genericStringValue(payload["documento_codigo"]), genericStringValue(baseRow["documento_codigo"]), genericStringValue(baseRow["codigo"]))
	if documentoCodigo == "" {
		http.Error(w, "documento_codigo no disponible para versionado", http.StatusBadRequest)
		return
	}

	maxVersion, err := queryDocumentoGestionMaxVersion(dbEmp, empresaID, documentoCodigo)
	if err != nil {
		http.Error(w, "No se pudo calcular version documental", http.StatusInternalServerError)
		return
	}
	currentVersion := parseDocumentoVersionInt(baseRow["version"])
	if currentVersion <= 0 {
		currentVersion = 1
	}
	nextVersion := maxInt64(maxVersion, currentVersion) + 1

	entidadID := anyToInt64(baseRow["entidad_id"])
	if raw, ok := payload["entidad_id"]; ok {
		entidadID = anyToInt64(raw)
	}

	tamanoBytes := anyToInt64(baseRow["tamano_bytes"])
	if raw, ok := payload["tamano_bytes"]; ok {
		tamanoBytes = anyToInt64(raw)
	}

	actor := strings.TrimSpace(adminEmailFromRequest(r))
	observacionesBase := strings.TrimSpace(genericStringValue(baseRow["observaciones"]))
	if extraObs := strings.TrimSpace(genericStringValue(payload["observaciones"])); extraObs != "" {
		if observacionesBase == "" {
			observacionesBase = extraObs
		} else {
			observacionesBase += " | " + extraObs
		}
	}
	auditVersion := fmt.Sprintf("[%s] versionado desde id=%d version=%d -> version=%d por %s", time.Now().Format("2006-01-02 15:04:05"), id, currentVersion, nextVersion, finanzasFirstNonBlank(actor, "sistema"))
	if observacionesBase == "" {
		observacionesBase = auditVersion
	} else {
		observacionesBase += " | " + auditVersion
	}

	newPayload := map[string]interface{}{
		"modulo":           finanzasFirstNonBlank(genericStringValue(payload["modulo"]), genericStringValue(baseRow["modulo"])),
		"entidad":          finanzasFirstNonBlank(genericStringValue(payload["entidad"]), genericStringValue(baseRow["entidad"])),
		"entidad_id":       entidadID,
		"documento_codigo": documentoCodigo,
		"nombre_documento": finanzasFirstNonBlank(genericStringValue(payload["nombre_documento"]), genericStringValue(baseRow["nombre_documento"])),
		"tipo_documento":   finanzasFirstNonBlank(genericStringValue(payload["tipo_documento"]), genericStringValue(baseRow["tipo_documento"])),
		"mime_type":        finanzasFirstNonBlank(genericStringValue(payload["mime_type"]), genericStringValue(baseRow["mime_type"])),
		"url_archivo":      finanzasFirstNonBlank(genericStringValue(payload["url_archivo"]), genericStringValue(baseRow["url_archivo"])),
		"hash_archivo":     finanzasFirstNonBlank(genericStringValue(payload["hash_archivo"]), genericStringValue(baseRow["hash_archivo"])),
		"tamano_bytes":     tamanoBytes,
		"version":          strconv.FormatInt(nextVersion, 10),
		"estado_documento": "vigente",
		"usuario_creador":  finanzasFirstNonBlank(genericStringValue(payload["usuario_creador"]), actor, genericStringValue(baseRow["usuario_creador"])),
		"estado":           finanzasFirstNonBlank(genericStringValue(payload["estado"]), "activo"),
		"observaciones":    observacionesBase,
	}
	ensureGenericCode(newPayload, cfgDocumentosGestion.CodeColumn, cfgDocumentosGestion.CodePrefix)

	nuevoID, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgDocumentosGestion.Table, empresaID, newPayload, cfgDocumentosGestion.AllowedColumns)
	if err != nil {
		http.Error(w, "No se pudo crear nueva version del documento", http.StatusBadRequest)
		return
	}

	warningMsg := ""
	updateAnterior := map[string]interface{}{"estado_documento": "historico"}
	auditAnterior := fmt.Sprintf("[%s] reemplazado por version=%d (id=%d)", time.Now().Format("2006-01-02 15:04:05"), nextVersion, nuevoID)
	if previousObs := strings.TrimSpace(genericStringValue(baseRow["observaciones"])); previousObs == "" {
		updateAnterior["observaciones"] = auditAnterior
	} else {
		updateAnterior["observaciones"] = previousObs + "\n" + auditAnterior
	}
	if err := dbpkg.UpdateEmpresaGenericRow(dbEmp, cfgDocumentosGestion.Table, empresaID, id, updateAnterior, cfgDocumentosGestion.AllowedColumns); err != nil {
		warningMsg = "se creo la nueva version, pero no se pudo marcar la version anterior como historica"
	}

	itemNuevo, _ := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgDocumentosGestion.Table, empresaID, nuevoID)
	itemAnterior, _ := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgDocumentosGestion.Table, empresaID, id)

	response := map[string]interface{}{
		"ok":               true,
		"empresa_id":       empresaID,
		"id_anterior":      id,
		"id_nuevo":         nuevoID,
		"documento_codigo": documentoCodigo,
		"version_anterior": currentVersion,
		"version_nueva":    nextVersion,
		"rol":              role,
		"modulo_permiso":   permissionModule,
		"item_anterior":    itemAnterior,
		"item_nuevo":       itemNuevo,
		"acceso_permitido": true,
	}
	if warningMsg != "" {
		response["warning"] = warningMsg
	}

	writeJSON(w, http.StatusCreated, response)
}

func loadDocumentoGestionVersionRows(dbEmp *sql.DB, empresaID int64, documentoCodigo string, includeInactive bool, limit int) ([]map[string]interface{}, error) {
	documentoCodigo = strings.TrimSpace(documentoCodigo)
	if documentoCodigo == "" {
		return []map[string]interface{}{}, nil
	}
	if limit <= 0 {
		limit = 200
	}
	if limit > 2000 {
		limit = 2000
	}

	query := `SELECT id
	FROM empresa_documentos_gestion
	WHERE empresa_id = ?
	  AND UPPER(COALESCE(documento_codigo, '')) = UPPER(?)`
	args := []interface{}{empresaID, documentoCodigo}

	if !includeInactive {
		query += ` AND LOWER(COALESCE(estado, 'activo')) = 'activo'`
	}

	query += ` ORDER BY CAST(COALESCE(NULLIF(version, ''), '0') AS INTEGER) DESC, id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := dbEmp.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	items := make([]map[string]interface{}, 0, len(ids))
	for _, id := range ids {
		item, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgDocumentosGestion.Table, empresaID, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func queryDocumentoGestionMaxVersion(dbEmp *sql.DB, empresaID int64, documentoCodigo string) (int64, error) {
	var maxVersion sql.NullInt64
	err := dbEmp.QueryRow(`SELECT MAX(CAST(COALESCE(NULLIF(version, ''), '0') AS INTEGER))
	FROM empresa_documentos_gestion
	WHERE empresa_id = ?
	  AND UPPER(COALESCE(documento_codigo, '')) = UPPER(?)`, empresaID, strings.TrimSpace(documentoCodigo)).Scan(&maxVersion)
	if err != nil {
		return 0, err
	}
	if !maxVersion.Valid {
		return 0, nil
	}
	if maxVersion.Int64 < 0 {
		return 0, nil
	}
	return maxVersion.Int64, nil
}

func parseDocumentoVersionInt(v interface{}) int64 {
	switch n := v.(type) {
	case int64:
		if n > 0 {
			return n
		}
	case int:
		if n > 0 {
			return int64(n)
		}
	case float64:
		if n > 0 {
			return int64(n)
		}
	case string:
		trimmed := strings.TrimSpace(n)
		if trimmed == "" {
			return 0
		}
		if parsed, err := strconv.ParseInt(trimmed, 10, 64); err == nil && parsed > 0 {
			return parsed
		}
		if parsed, err := strconv.ParseFloat(trimmed, 64); err == nil && parsed > 0 {
			return int64(parsed)
		}
	}
	return 0
}

func resolveDocumentoPermissionActionFromRequest(r *http.Request) string {
	return parseDocumentoPermissionAction(finanzasFirstNonBlank(
		strings.TrimSpace(r.URL.Query().Get("permiso")),
		strings.TrimSpace(r.URL.Query().Get("accion_permiso")),
		strings.TrimSpace(r.URL.Query().Get("permission_action")),
		strings.TrimSpace(r.URL.Query().Get("action_permiso")),
	))
}

func parseDocumentoPermissionAction(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "r", "read", "leer", "lectura":
		return permActionRead
	case "c", "create", "crear", "creacion":
		return permActionCreate
	case "u", "update", "editar", "actualizar", "modificar":
		return permActionUpdate
	case "d", "delete", "eliminar", "borrar":
		return permActionDelete
	case "a", "approve", "aprobar", "aprobacion":
		return permActionApprove
	default:
		return normalizePermissionAction(strings.ToUpper(strings.TrimSpace(raw)), permActionRead)
	}
}

func evaluateDocumentoGestionAccess(r *http.Request, row map[string]interface{}, requestedAction string) (bool, string, string) {
	role := normalizePermissionRole(adminRoleFromRequest(r))
	permissionModule := mapDocumentoModuloToPermissionModule(genericStringValue(row["modulo"]))
	if role == "" {
		return true, role, permissionModule
	}
	allowed := roleAllowsModuleAction(role, permissionModule, requestedAction)
	return allowed, role, permissionModule
}

func mapDocumentoModuloToPermissionModule(moduloRaw string) string {
	modulo := strings.ToLower(strings.TrimSpace(moduloRaw))
	if modulo == "" {
		return permModuleSeguridad
	}

	switch modulo {
	case permModuleVentas, permModuleInventario, permModuleFinanzas, permModuleClientes, permModuleCompras, permModuleFacturacion, permModuleSeguridad:
		return modulo
	}

	if strings.Contains(modulo, "factur") || strings.Contains(modulo, "dian") {
		return permModuleFacturacion
	}
	if strings.Contains(modulo, "compra") || strings.Contains(modulo, "proveedor") {
		return permModuleCompras
	}
	if strings.Contains(modulo, "invent") || strings.Contains(modulo, "bodega") || strings.Contains(modulo, "produccion") || strings.Contains(modulo, "logistica") {
		return permModuleInventario
	}
	if strings.Contains(modulo, "cliente") || strings.Contains(modulo, "crm") || strings.Contains(modulo, "reserva") || strings.Contains(modulo, "vehiculo") {
		return permModuleClientes
	}
	if strings.Contains(modulo, "finanza") || strings.Contains(modulo, "conta") || strings.Contains(modulo, "nomina") || strings.Contains(modulo, "rrhh") || strings.Contains(modulo, "cartera") {
		return permModuleFinanzas
	}
	if strings.Contains(modulo, "venta") || strings.Contains(modulo, "pedido") || strings.Contains(modulo, "cotizacion") || strings.Contains(modulo, "devolucion") {
		return permModuleVentas
	}

	return permModuleSeguridad
}

type produccionPlanDiaAgg struct {
	Fecha      string
	Ordenes    int64
	Programada float64
	Producida  float64
}

func handleProduccionOrdenesPlanCapacidadAction(dbEmp *sql.DB, w http.ResponseWriter, r *http.Request) {
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
	if limit <= 0 {
		limit = 500
	}
	if limit > 2000 {
		limit = 2000
	}

	metaDiaria := 100.0
	if rawMeta := strings.TrimSpace(r.URL.Query().Get("meta_diaria")); rawMeta != "" {
		parsed, parseErr := strconv.ParseFloat(rawMeta, 64)
		if parseErr != nil || parsed <= 0 {
			http.Error(w, "meta_diaria invalida", http.StatusBadRequest)
			return
		}
		metaDiaria = parsed
	}

	desde := rrhhNormalizeDateOnly(strings.TrimSpace(r.URL.Query().Get("desde")))
	hasta := rrhhNormalizeDateOnly(strings.TrimSpace(r.URL.Query().Get("hasta")))
	if desde != "" && hasta != "" {
		if d, okD := ventasParseDateTime(desde); okD {
			if h, okH := ventasParseDateTime(hasta); okH && d.After(h) {
				desde, hasta = hasta, desde
			}
		}
	}

	query := `SELECT
		id,
		COALESCE(codigo, ''),
		COALESCE(producto_nombre, ''),
		COALESCE(fecha_programada, ''),
		COALESCE(fecha_inicio, ''),
		COALESCE(fecha_fin, ''),
		COALESCE(estado_orden, ''),
		COALESCE(cantidad_programada, 0),
		COALESCE(cantidad_producida, 0),
		COALESCE(costo_estimado, 0),
		COALESCE(costo_real, 0),
		COALESCE(responsable, ''),
		COALESCE(estado, 'activo')
	FROM produccion_ordenes
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	if !parseBoolQuery(r, "include_inactive") {
		query += ` AND LOWER(COALESCE(estado, 'activo')) = 'activo'`
	}

	dateExpr := `date(COALESCE(NULLIF(fecha_programada, ''), NULLIF(fecha_inicio, ''), fecha_creacion))`
	if desde != "" {
		query += ` AND ` + dateExpr + ` >= date(?)`
		args = append(args, desde)
	}
	if hasta != "" {
		query += ` AND ` + dateExpr + ` <= date(?)`
		args = append(args, hasta)
	}

	query += ` ORDER BY ` + dateExpr + ` ASC, id ASC LIMIT ?`
	args = append(args, limit)

	rows, err := dbEmp.Query(query, args...)
	if err != nil {
		http.Error(w, "no se pudo consultar plan de capacidad", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	now := time.Now().In(time.Local)
	nowDate, _ := time.Parse("2006-01-02", now.Format("2006-01-02"))

	items := make([]map[string]interface{}, 0)
	planByDate := make(map[string]*produccionPlanDiaAgg)
	totalProgramada := 0.0
	totalProducida := 0.0
	totalPendiente := 0.0
	ordenesAtrasadas := int64(0)
	ordenesFinalizadas := int64(0)
	ordenesEnProceso := int64(0)
	alertasTotales := int64(0)

	for rows.Next() {
		var id int64
		var codigo string
		var productoNombre string
		var fechaProgramada string
		var fechaInicio string
		var fechaFin string
		var estadoOrden string
		var cantidadProgramada float64
		var cantidadProducida float64
		var costoEstimado float64
		var costoReal float64
		var responsable string
		var estadoRegistro string

		if scanErr := rows.Scan(
			&id,
			&codigo,
			&productoNombre,
			&fechaProgramada,
			&fechaInicio,
			&fechaFin,
			&estadoOrden,
			&cantidadProgramada,
			&cantidadProducida,
			&costoEstimado,
			&costoReal,
			&responsable,
			&estadoRegistro,
		); scanErr != nil {
			http.Error(w, "no se pudo leer orden de produccion", http.StatusInternalServerError)
			return
		}

		if cantidadProgramada < 0 {
			cantidadProgramada = 0
		}
		if cantidadProducida < 0 {
			cantidadProducida = 0
		}

		pendiente := cantidadProgramada - cantidadProducida
		if pendiente < 0 {
			pendiente = 0
		}

		cumplimiento := 0.0
		if cantidadProgramada > 0 {
			cumplimiento = reportesRound((cantidadProducida * 100.0) / cantidadProgramada)
		}

		fechaBase := rrhhNormalizeDateOnly(finanzasFirstNonBlank(fechaProgramada, fechaInicio, fechaFin))
		isFinalizada := produccionEstadoEsFinal(estadoOrden)
		if isFinalizada {
			ordenesFinalizadas++
		} else {
			ordenesEnProceso++
		}

		atrasada := false
		if fechaBase != "" && !isFinalizada {
			if parsedFecha, ok := ventasParseDateTime(fechaBase); ok {
				baseDate, _ := time.Parse("2006-01-02", parsedFecha.Format("2006-01-02"))
				if baseDate.Before(nowDate) {
					atrasada = true
				}
			}
		}

		alertaTipo := ""
		alerta := ""
		if atrasada && pendiente > 0.009 {
			alertaTipo = "atrasada_con_pendiente"
			alerta = "Orden programada vencida con cantidad pendiente de produccion."
		} else if cantidadProgramada > metaDiaria {
			alertaTipo = "sobrecapacidad_programada"
			alerta = "Orden supera la meta diaria de capacidad planificada."
		} else if !isFinalizada && cumplimiento > 0 && cumplimiento < 80 {
			alertaTipo = "cumplimiento_bajo"
			alerta = "Orden en proceso con cumplimiento acumulado por debajo del umbral operativo."
		}
		if alertaTipo != "" {
			alertasTotales++
		}

		desviacionMetaCantidad := reportesRound(cantidadProgramada - metaDiaria)
		desviacionMetaPct := 0.0
		if metaDiaria > 0 {
			desviacionMetaPct = reportesRound((desviacionMetaCantidad * 100.0) / metaDiaria)
		}

		items = append(items, map[string]interface{}{
			"id":                     id,
			"codigo":                 codigo,
			"producto_nombre":        productoNombre,
			"fecha_programada":       fechaProgramada,
			"fecha_inicio":           fechaInicio,
			"fecha_fin":              fechaFin,
			"estado_orden":           estadoOrden,
			"cantidad_programada":    reportesRound(cantidadProgramada),
			"cantidad_producida":     reportesRound(cantidadProducida),
			"cantidad_pendiente":     reportesRound(pendiente),
			"cumplimiento_pct":       cumplimiento,
			"meta_diaria":            reportesRound(metaDiaria),
			"desviacion_vs_meta":     desviacionMetaCantidad,
			"desviacion_vs_meta_pct": desviacionMetaPct,
			"atrasada":               atrasada,
			"costo_estimado":         reportesRound(costoEstimado),
			"costo_real":             reportesRound(costoReal),
			"responsable":            responsable,
			"estado":                 estadoRegistro,
			"alerta_tipo":            alertaTipo,
			"alerta":                 alerta,
		})

		totalProgramada += cantidadProgramada
		totalProducida += cantidadProducida
		totalPendiente += pendiente
		if atrasada {
			ordenesAtrasadas++
		}

		if fechaBase != "" {
			agg := planByDate[fechaBase]
			if agg == nil {
				agg = &produccionPlanDiaAgg{Fecha: fechaBase}
				planByDate[fechaBase] = agg
			}
			agg.Ordenes++
			agg.Programada = reportesRound(agg.Programada + cantidadProgramada)
			agg.Producida = reportesRound(agg.Producida + cantidadProducida)
		}
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "no se pudo consultar plan de capacidad", http.StatusInternalServerError)
		return
	}

	planFechas := make([]string, 0, len(planByDate))
	for fecha := range planByDate {
		planFechas = append(planFechas, fecha)
	}
	sort.Strings(planFechas)

	planDiario := make([]map[string]interface{}, 0, len(planFechas))
	for _, fecha := range planFechas {
		agg := planByDate[fecha]
		desvCantidad := reportesRound(agg.Programada - metaDiaria)
		desvPct := 0.0
		if metaDiaria > 0 {
			desvPct = reportesRound((desvCantidad * 100.0) / metaDiaria)
		}
		planDiario = append(planDiario, map[string]interface{}{
			"fecha":               fecha,
			"ordenes":             agg.Ordenes,
			"cantidad_programada": reportesRound(agg.Programada),
			"cantidad_producida":  reportesRound(agg.Producida),
			"meta_diaria":         reportesRound(metaDiaria),
			"desviacion_cantidad": desvCantidad,
			"desviacion_pct":      desvPct,
		})
	}

	diasPlan := produccionDiasPlanificacion(desde, hasta, int64(len(planDiario)))
	objetivoTotal := reportesRound(float64(diasPlan) * metaDiaria)
	desviacionTotal := reportesRound(totalProgramada - objetivoTotal)
	desviacionTotalPct := 0.0
	if objetivoTotal > 0 {
		desviacionTotalPct = reportesRound((desviacionTotal * 100.0) / objetivoTotal)
	}

	cumplimientoGlobal := 0.0
	if totalProgramada > 0 {
		cumplimientoGlobal = reportesRound((totalProducida * 100.0) / totalProgramada)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":          true,
		"empresa_id":  empresaID,
		"meta_diaria": reportesRound(metaDiaria),
		"resumen": map[string]interface{}{
			"ordenes_total":             int64(len(items)),
			"ordenes_en_proceso":        ordenesEnProceso,
			"ordenes_finalizadas":       ordenesFinalizadas,
			"ordenes_atrasadas":         ordenesAtrasadas,
			"alertas_totales":           alertasTotales,
			"dias_planificados":         diasPlan,
			"capacidad_objetivo_total":  objetivoTotal,
			"cantidad_programada_total": reportesRound(totalProgramada),
			"cantidad_producida_total":  reportesRound(totalProducida),
			"cantidad_pendiente_total":  reportesRound(totalPendiente),
			"cumplimiento_global_pct":   cumplimientoGlobal,
			"desviacion_objetivo":       desviacionTotal,
			"desviacion_objetivo_pct":   desviacionTotalPct,
		},
		"plan_diario": planDiario,
		"items":       items,
	})
}

func handleLogisticaEnviosSeguimientoHitosAction(dbEmp *sql.DB, w http.ResponseWriter, r *http.Request) {
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
	if limit <= 0 {
		limit = 500
	}
	if limit > 2000 {
		limit = 2000
	}

	slaHoras := int64(24)
	if rawSLA := strings.TrimSpace(r.URL.Query().Get("sla_horas")); rawSLA != "" {
		parsed, parseErr := strconv.ParseInt(rawSLA, 10, 64)
		if parseErr != nil || parsed <= 0 {
			http.Error(w, "sla_horas invalido", http.StatusBadRequest)
			return
		}
		slaHoras = parsed
	}

	desde := rrhhNormalizeDateOnly(strings.TrimSpace(r.URL.Query().Get("desde")))
	hasta := rrhhNormalizeDateOnly(strings.TrimSpace(r.URL.Query().Get("hasta")))
	if desde != "" && hasta != "" {
		if d, okD := ventasParseDateTime(desde); okD {
			if h, okH := ventasParseDateTime(hasta); okH && d.After(h) {
				desde, hasta = hasta, desde
			}
		}
	}

	query := `SELECT
		id,
		COALESCE(codigo, ''),
		COALESCE(cliente_nombre, ''),
		COALESCE(documento_referencia, ''),
		COALESCE(direccion_entrega, ''),
		COALESCE(fecha_programada, ''),
		COALESCE(fecha_salida, ''),
		COALESCE(fecha_entrega, ''),
		COALESCE(estado_envio, ''),
		COALESCE(observaciones_seguimiento, ''),
		COALESCE(estado, 'activo')
	FROM logistica_envios
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	if !parseBoolQuery(r, "include_inactive") {
		query += ` AND LOWER(COALESCE(estado, 'activo')) = 'activo'`
	}

	dateExpr := `date(COALESCE(NULLIF(fecha_programada, ''), NULLIF(fecha_salida, ''), fecha_creacion))`
	if desde != "" {
		query += ` AND ` + dateExpr + ` >= date(?)`
		args = append(args, desde)
	}
	if hasta != "" {
		query += ` AND ` + dateExpr + ` <= date(?)`
		args = append(args, hasta)
	}

	query += ` ORDER BY ` + dateExpr + ` ASC, id ASC LIMIT ?`
	args = append(args, limit)

	rows, err := dbEmp.Query(query, args...)
	if err != nil {
		http.Error(w, "no se pudo consultar seguimiento logistico", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	now := time.Now().In(time.Local)
	items := make([]map[string]interface{}, 0)
	alertas := make([]map[string]interface{}, 0)
	totalEnvios := int64(0)
	conHitoSalida := int64(0)
	conHitoEntrega := int64(0)
	incumplidos := int64(0)
	cumplen := int64(0)

	for rows.Next() {
		var id int64
		var codigo string
		var clienteNombre string
		var documentoReferencia string
		var direccionEntrega string
		var fechaProgramada string
		var fechaSalida string
		var fechaEntrega string
		var estadoEnvio string
		var observacionesSeguimiento string
		var estadoRegistro string

		if scanErr := rows.Scan(
			&id,
			&codigo,
			&clienteNombre,
			&documentoReferencia,
			&direccionEntrega,
			&fechaProgramada,
			&fechaSalida,
			&fechaEntrega,
			&estadoEnvio,
			&observacionesSeguimiento,
			&estadoRegistro,
		); scanErr != nil {
			http.Error(w, "no se pudo leer seguimiento logistico", http.StatusInternalServerError)
			return
		}

		totalEnvios++

		programadaAt, hasProgramada := ventasParseDateTime(fechaProgramada)
		salidaAt, hasSalida := ventasParseDateTime(fechaSalida)
		entregaAt, hasEntrega := ventasParseDateTime(fechaEntrega)

		if hasSalida {
			conHitoSalida++
		}
		if hasEntrega {
			conHitoEntrega++
		}

		horasDesdeProgramacion := int64(0)
		if hasProgramada {
			horasDesdeProgramacion = int64(now.Sub(programadaAt).Hours())
			if horasDesdeProgramacion < 0 {
				horasDesdeProgramacion = 0
			}
		}

		horasSalidaDesdeProgramacion := int64(0)
		if hasProgramada && hasSalida {
			horasSalidaDesdeProgramacion = int64(salidaAt.Sub(programadaAt).Hours())
		}

		horasEntregaDesdeProgramacion := int64(0)
		if hasProgramada && hasEntrega {
			horasEntregaDesdeProgramacion = int64(entregaAt.Sub(programadaAt).Hours())
		}

		alertaTipo := ""
		alerta := ""
		incumplido := false
		if !hasProgramada {
			alertaTipo = "sin_programacion"
			alerta = "Envio sin fecha programada para controlar hitos de SLA."
			incumplido = true
		} else {
			limiteSalida := programadaAt.Add(time.Duration(maxInt64(1, slaHoras/2)) * time.Hour)
			limiteEntrega := programadaAt.Add(time.Duration(slaHoras) * time.Hour)
			estadoFinal := logisticaEstadoEsFinal(estadoEnvio)

			switch {
			case !hasSalida && now.After(limiteSalida):
				alertaTipo = "sin_salida"
				alerta = "Envio programado sin hito de salida dentro del SLA de despacho."
				incumplido = true
			case hasEntrega && entregaAt.After(limiteEntrega):
				alertaTipo = "entrega_tardia"
				alerta = "Entrega registrada fuera del SLA objetivo."
				incumplido = true
			case !hasEntrega && now.After(limiteEntrega) && !estadoFinal:
				alertaTipo = "entrega_pendiente_vencida"
				alerta = "Envio sin entrega confirmada y SLA vencido."
				incumplido = true
			case estadoFinal && !hasEntrega:
				alertaTipo = "sin_hito_entrega"
				alerta = "Envio en estado final sin fecha de entrega registrada."
				incumplido = true
			}
		}

		if incumplido {
			incumplidos++
			alertas = append(alertas, map[string]interface{}{
				"id":          id,
				"codigo":      codigo,
				"alerta_tipo": alertaTipo,
				"alerta":      alerta,
			})
		} else {
			cumplen++
		}

		items = append(items, map[string]interface{}{
			"id":                               id,
			"codigo":                           codigo,
			"cliente_nombre":                   clienteNombre,
			"documento_referencia":             documentoReferencia,
			"direccion_entrega":                direccionEntrega,
			"fecha_programada":                 fechaProgramada,
			"fecha_salida":                     fechaSalida,
			"fecha_entrega":                    fechaEntrega,
			"estado_envio":                     estadoEnvio,
			"horas_desde_programacion":         horasDesdeProgramacion,
			"horas_salida_desde_programacion":  horasSalidaDesdeProgramacion,
			"horas_entrega_desde_programacion": horasEntregaDesdeProgramacion,
			"hito_programacion":                hasProgramada,
			"hito_salida":                      hasSalida,
			"hito_entrega":                     hasEntrega,
			"incumplido":                       incumplido,
			"alerta_tipo":                      alertaTipo,
			"alerta":                           alerta,
			"observaciones_seguimiento":        observacionesSeguimiento,
			"estado":                           estadoRegistro,
		})
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "no se pudo consultar seguimiento logistico", http.StatusInternalServerError)
		return
	}

	cumplimientoPct := 0.0
	if totalEnvios > 0 {
		cumplimientoPct = reportesRound((float64(cumplen) * 100.0) / float64(totalEnvios))
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"sla_horas":  slaHoras,
		"resumen": map[string]interface{}{
			"envios_total":         totalEnvios,
			"con_hito_salida":      conHitoSalida,
			"con_hito_entrega":     conHitoEntrega,
			"incumplidos":          incumplidos,
			"cumplen_sla":          cumplen,
			"cumplimiento_sla_pct": cumplimientoPct,
		},
		"alertas": alertas,
		"items":   items,
	})
}

func produccionEstadoEsFinal(estado string) bool {
	switch normalizeStateMachineValue(estado) {
	case "entregado", "cerrado", "finalizada", "aplicada", "cancelada", "anulada":
		return true
	default:
		return false
	}
}

func produccionDiasPlanificacion(desde, hasta string, diasConDatos int64) int64 {
	if desde != "" && hasta != "" {
		if parsedDesde, okDesde := ventasParseDateTime(desde); okDesde {
			if parsedHasta, okHasta := ventasParseDateTime(hasta); okHasta {
				dInicio, _ := time.Parse("2006-01-02", parsedDesde.Format("2006-01-02"))
				dFin, _ := time.Parse("2006-01-02", parsedHasta.Format("2006-01-02"))
				if dFin.Before(dInicio) {
					dInicio, dFin = dFin, dInicio
				}
				dias := int64(dFin.Sub(dInicio).Hours()/24) + 1
				if dias > 0 {
					return dias
				}
			}
		}
	}
	if diasConDatos > 0 {
		return diasConDatos
	}
	return 1
}

func logisticaEstadoEsFinal(estado string) bool {
	switch normalizeStateMachineValue(estado) {
	case "entregado", "cerrado", "aplicado", "cancelado", "anulado":
		return true
	default:
		return false
	}
}

func maxInt64(a, b int64) int64 {
	if a >= b {
		return a
	}
	return b
}

func EmpresaIntegracionesAPIsHandler(dbEmp *sql.DB) http.HandlerFunc {
	return empresaModuloIntegracionesCRUDHandler(dbEmp, cfgIntegracionesAPIs, integrationOpsAPIs)
}

func EmpresaIntegracionesBancosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return empresaModuloIntegracionesCRUDHandler(dbEmp, cfgIntegracionesBancos, integrationOpsBancos)
}

type ventasEmbudoSnapshot struct {
	Rows    []map[string]interface{}
	Summary map[string]interface{}
	Alertas []map[string]interface{}
}

type ventasConversionError struct {
	status  int
	message string
}

func (e *ventasConversionError) Error() string {
	return e.message
}

func newVentasConversionError(status int, message string) error {
	return &ventasConversionError{status: status, message: strings.TrimSpace(message)}
}

func ventasErrorStatus(err error, fallback int) int {
	if err == nil {
		return fallback
	}
	var typed *ventasConversionError
	if errors.As(err, &typed) {
		if typed.status > 0 {
			return typed.status
		}
	}
	return fallback
}

func handleVentasCotizacionConvertirPedidoAction(dbEmp *sql.DB, w http.ResponseWriter, r *http.Request) {
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
	cotizacionID := resolveIDFromPayloadOrQuery(payload, r)
	if cotizacionID <= 0 {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	actor := strings.TrimSpace(adminEmailFromRequest(r))
	pedidoCodigo := genericStringValue(payload["pedido_codigo"])
	cotizacion, pedido, pedidoCreado, autoAprobada, err := convertCotizacionToPedido(dbEmp, empresaID, cotizacionID, actor, pedidoCodigo)
	if err != nil {
		http.Error(w, err.Error(), ventasErrorStatus(err, http.StatusBadRequest))
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                       true,
		"empresa_id":               empresaID,
		"cotizacion_id":            cotizacionID,
		"pedido_id":                anyToInt64(pedido["id"]),
		"pedido_creado":            pedidoCreado,
		"cotizacion_auto_aprobada": autoAprobada,
		"cotizacion":               cotizacion,
		"pedido":                   pedido,
	})
}

func handleVentasCotizacionConvertirDocumentoFinalAction(dbEmp *sql.DB, w http.ResponseWriter, r *http.Request) {
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
	cotizacionID := resolveIDFromPayloadOrQuery(payload, r)
	if cotizacionID <= 0 {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	actor := strings.TrimSpace(adminEmailFromRequest(r))
	pedidoCodigo := genericStringValue(payload["pedido_codigo"])
	cotizacion, pedido, pedidoCreado, autoAprobada, err := convertCotizacionToPedido(dbEmp, empresaID, cotizacionID, actor, pedidoCodigo)
	if err != nil {
		http.Error(w, err.Error(), ventasErrorStatus(err, http.StatusBadRequest))
		return
	}

	pedidoID := anyToInt64(pedido["id"])
	if pedidoID <= 0 {
		http.Error(w, "pedido no disponible para conversion documental", http.StatusConflict)
		return
	}

	pedidoUpdated, documentoFinal, documentoCreado, err := convertPedidoToDocumentoFinal(dbEmp, empresaID, pedidoID, payload, actor)
	if err != nil {
		http.Error(w, err.Error(), ventasErrorStatus(err, http.StatusBadRequest))
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                       true,
		"empresa_id":               empresaID,
		"cotizacion_id":            cotizacionID,
		"pedido_id":                pedidoID,
		"pedido_creado":            pedidoCreado,
		"cotizacion_auto_aprobada": autoAprobada,
		"documento_final_creado":   documentoCreado,
		"cotizacion":               cotizacion,
		"pedido":                   pedidoUpdated,
		"documento_final":          documentoFinal,
	})
}

func handleVentasPedidoConvertirDocumentoFinalAction(dbEmp *sql.DB, w http.ResponseWriter, r *http.Request) {
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
	pedidoID := resolveIDFromPayloadOrQuery(payload, r)
	if pedidoID <= 0 {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}

	actor := strings.TrimSpace(adminEmailFromRequest(r))
	pedidoUpdated, documentoFinal, documentoCreado, err := convertPedidoToDocumentoFinal(dbEmp, empresaID, pedidoID, payload, actor)
	if err != nil {
		http.Error(w, err.Error(), ventasErrorStatus(err, http.StatusBadRequest))
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                     true,
		"empresa_id":             empresaID,
		"pedido_id":              pedidoID,
		"documento_final_creado": documentoCreado,
		"pedido":                 pedidoUpdated,
		"documento_final":        documentoFinal,
	})
}

func handleVentasEmbudoConversionAction(dbEmp *sql.DB, w http.ResponseWriter, r *http.Request) {
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
	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}

	slaCotizacionHoras, err := parseIntQueryOptional(r, "sla_cotizacion_horas")
	if err != nil {
		http.Error(w, "sla_cotizacion_horas invalido", http.StatusBadRequest)
		return
	}
	slaPedidoHoras, err := parseIntQueryOptional(r, "sla_pedido_horas")
	if err != nil {
		http.Error(w, "sla_pedido_horas invalido", http.StatusBadRequest)
		return
	}
	if slaCotizacionHoras <= 0 {
		slaCotizacionHoras = 48
	}
	if slaPedidoHoras <= 0 {
		slaPedidoHoras = 72
	}

	desde := strings.TrimSpace(r.URL.Query().Get("desde"))
	hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))

	snapshot, err := buildVentasEmbudoConversionSnapshot(
		dbEmp,
		empresaID,
		desde,
		hasta,
		slaCotizacionHoras,
		slaPedidoHoras,
		parseBoolQuery(r, "include_inactive"),
		limit,
	)
	if err != nil {
		http.Error(w, "No se pudo construir embudo comercial", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                   true,
		"empresa_id":           empresaID,
		"desde":                desde,
		"hasta":                hasta,
		"sla_cotizacion_horas": slaCotizacionHoras,
		"sla_pedido_horas":     slaPedidoHoras,
		"summary":              snapshot.Summary,
		"items":                snapshot.Rows,
		"alertas":              snapshot.Alertas,
	})
}

func convertCotizacionToPedido(dbEmp *sql.DB, empresaID, cotizacionID int64, actor, pedidoCodigo string) (map[string]interface{}, map[string]interface{}, bool, bool, error) {
	if empresaID <= 0 {
		return nil, nil, false, false, newVentasConversionError(http.StatusBadRequest, "empresa_id required")
	}
	if cotizacionID <= 0 {
		return nil, nil, false, false, newVentasConversionError(http.StatusBadRequest, "id required")
	}
	if err := dbpkg.EnsureEmpresaModulosFaltantesSchema(dbEmp); err != nil {
		return nil, nil, false, false, err
	}

	cotizacion, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgCotizacionesVenta.Table, empresaID, cotizacionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, false, false, newVentasConversionError(http.StatusNotFound, "cotizacion no encontrada")
		}
		return nil, nil, false, false, err
	}

	estadoActual := normalizeStateMachineValue(genericStringValue(cotizacion["estado_documento"]))
	autoAprobada := false

	switch estadoActual {
	case "aprobada", "convertida":
		// Estado habilitado para conversión o ya convertido.
	case "borrador", "emitida":
		update := map[string]interface{}{"estado_documento": "aprobada"}
		if hasAllowedColumn(cfgCotizacionesVenta.AllowedColumns, "observaciones") {
			update["observaciones"] = appendStateMachineObservation(genericStringValue(cotizacion["observaciones"]), estadoActual, "aprobada", "aprobacion automatica para conversion a pedido", actor)
		}
		if err := dbpkg.UpdateEmpresaGenericRow(dbEmp, cfgCotizacionesVenta.Table, empresaID, cotizacionID, update, cfgCotizacionesVenta.AllowedColumns); err != nil {
			return nil, nil, false, false, err
		}
		autoAprobada = true
		cotizacion, _ = dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgCotizacionesVenta.Table, empresaID, cotizacionID)
		estadoActual = "aprobada"
	default:
		return nil, nil, false, autoAprobada, newVentasConversionError(http.StatusConflict, "la cotizacion debe estar aprobada o emitida para convertir")
	}

	pedido, err := resolvePedidoForCotizacion(dbEmp, empresaID, cotizacionID, anyToInt64(cotizacion["convertido_pedido_id"]))
	if err != nil {
		return nil, nil, false, autoAprobada, err
	}

	pedidoCreado := false
	if pedido == nil {
		pedidoPayload := map[string]interface{}{
			"cliente_id":             anyToInt64(cotizacion["cliente_id"]),
			"cliente_nombre":         genericStringValue(cotizacion["cliente_nombre"]),
			"cotizacion_id":          cotizacionID,
			"fecha_pedido":           ventasFirstNonBlank(genericStringValue(cotizacion["fecha_documento"]), time.Now().In(time.Local).Format("2006-01-02 15:04:05")),
			"fecha_entrega_estimada": genericStringValue(cotizacion["vigencia_hasta"]),
			"estado_pedido":          "confirmado",
			"subtotal":               cotizacion["subtotal"],
			"descuento_total":        cotizacion["descuento_total"],
			"impuesto_total":         cotizacion["impuesto_total"],
			"total":                  cotizacion["total"],
			"moneda":                 genericStringDefault(cotizacion["moneda"], "COP"),
			"usuario_creador":        actor,
			"estado":                 "activo",
		}

		notas := strings.TrimSpace(genericStringValue(cotizacion["notas"]))
		refCotizacion := genericStringValue(cotizacion["codigo"])
		if refCotizacion != "" {
			if notas == "" {
				notas = "convertida desde cotizacion " + refCotizacion
			} else {
				notas = notas + " | convertida desde cotizacion " + refCotizacion
			}
		}
		if notas != "" {
			pedidoPayload["notas"] = notas
		}

		pedidoCodigo = strings.TrimSpace(pedidoCodigo)
		if pedidoCodigo != "" {
			pedidoPayload["codigo"] = pedidoCodigo
		}
		if genericStringValue(pedidoPayload["codigo"]) == "" {
			token := ventasSanitizeCodeToken(refCotizacion)
			if token != "" {
				pedidoPayload["codigo"] = "PED-" + token
			}
		}

		applyGenericDefaultValues(pedidoPayload, cfgPedidosVenta.DefaultValues)
		ensureGenericCode(pedidoPayload, cfgPedidosVenta.CodeColumn, cfgPedidosVenta.CodePrefix)

		pedidoID, createErr := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgPedidosVenta.Table, empresaID, pedidoPayload, cfgPedidosVenta.AllowedColumns)
		if createErr != nil && strings.Contains(strings.ToLower(createErr.Error()), "unique") {
			pedidoPayload["codigo"] = cfgPedidosVenta.CodePrefix + "-" + time.Now().Format("20060102150405") + "-" + strconv.FormatInt(time.Now().UnixNano()%1000000, 10)
			pedidoID, createErr = dbpkg.CreateEmpresaGenericRow(dbEmp, cfgPedidosVenta.Table, empresaID, pedidoPayload, cfgPedidosVenta.AllowedColumns)
		}
		if createErr != nil {
			return nil, nil, false, autoAprobada, createErr
		}

		pedido, err = dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgPedidosVenta.Table, empresaID, pedidoID)
		if err != nil {
			return nil, nil, false, autoAprobada, err
		}
		pedidoCreado = true
	}

	pedidoID := anyToInt64(pedido["id"])
	if pedidoID <= 0 {
		return nil, nil, false, autoAprobada, newVentasConversionError(http.StatusConflict, "pedido no disponible para cotizacion")
	}

	updateCotizacion := map[string]interface{}{
		"estado_documento":     "convertida",
		"convertido_pedido_id": pedidoID,
	}
	if hasAllowedColumn(cfgCotizacionesVenta.AllowedColumns, "observaciones") {
		updateCotizacion["observaciones"] = appendStateMachineObservation(
			genericStringValue(cotizacion["observaciones"]),
			estadoActual,
			"convertida",
			"conversion automatica a pedido "+genericStringValue(pedido["codigo"]),
			actor,
		)
	}
	if err := dbpkg.UpdateEmpresaGenericRow(dbEmp, cfgCotizacionesVenta.Table, empresaID, cotizacionID, updateCotizacion, cfgCotizacionesVenta.AllowedColumns); err != nil {
		return nil, nil, false, autoAprobada, err
	}

	cotizacion, _ = dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgCotizacionesVenta.Table, empresaID, cotizacionID)
	return cotizacion, pedido, pedidoCreado, autoAprobada, nil
}

func convertPedidoToDocumentoFinal(dbEmp *sql.DB, empresaID, pedidoID int64, payload map[string]interface{}, actor string) (map[string]interface{}, *dbpkg.EmpresaDocumentoFacturacion, bool, error) {
	if empresaID <= 0 {
		return nil, nil, false, newVentasConversionError(http.StatusBadRequest, "empresa_id required")
	}
	if pedidoID <= 0 {
		return nil, nil, false, newVentasConversionError(http.StatusBadRequest, "id required")
	}
	if err := dbpkg.EnsureEmpresaModulosFaltantesSchema(dbEmp); err != nil {
		return nil, nil, false, err
	}
	if err := dbpkg.EnsureEmpresaDocumentosTransaccionalesSchema(dbEmp); err != nil {
		return nil, nil, false, err
	}

	pedido, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgPedidosVenta.Table, empresaID, pedidoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, false, newVentasConversionError(http.StatusNotFound, "pedido no encontrado")
		}
		return nil, nil, false, err
	}

	estadoPedido := normalizeStateMachineValue(genericStringValue(pedido["estado_pedido"]))
	if estadoPedido == "borrador" {
		update := map[string]interface{}{"estado_pedido": "confirmado"}
		if hasAllowedColumn(cfgPedidosVenta.AllowedColumns, "observaciones") {
			update["observaciones"] = appendStateMachineObservation(genericStringValue(pedido["observaciones"]), estadoPedido, "confirmado", "aprobacion automatica para documento final", actor)
		}
		if err := dbpkg.UpdateEmpresaGenericRow(dbEmp, cfgPedidosVenta.Table, empresaID, pedidoID, update, cfgPedidosVenta.AllowedColumns); err != nil {
			return nil, nil, false, err
		}
		pedido, _ = dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgPedidosVenta.Table, empresaID, pedidoID)
		estadoPedido = "confirmado"
	}
	if estadoPedido == "cancelado" {
		return nil, nil, false, newVentasConversionError(http.StatusConflict, "no se puede generar documento final desde un pedido cancelado")
	}

	docExistente, err := findDocumentoFacturacionByPedidoID(dbEmp, empresaID, pedidoID)
	if err != nil {
		return nil, nil, false, err
	}

	tipoDocumento := strings.ToLower(strings.TrimSpace(genericStringDefault(payload["tipo_documento"], "factura_electronica")))
	if tipoDocumento == "" {
		tipoDocumento = "factura_electronica"
	}

	documentoCodigo := strings.TrimSpace(genericStringValue(payload["documento_codigo"]))
	if docExistente != nil {
		if documentoCodigo == "" {
			documentoCodigo = strings.TrimSpace(docExistente.DocumentoCodigo)
		}
		if tipoDocumento == "" {
			tipoDocumento = strings.TrimSpace(docExistente.TipoDocumento)
		}
	}
	if documentoCodigo == "" {
		token := ventasSanitizeCodeToken(genericStringValue(pedido["codigo"]))
		if token == "" {
			token = strconv.FormatInt(pedidoID, 10)
		}
		documentoCodigo = "FV-" + token
	}

	estadoDocumento := strings.ToLower(strings.TrimSpace(genericStringDefault(payload["estado_documento"], "emitida")))
	if estadoDocumento == "" {
		estadoDocumento = "emitida"
	}

	fechaDocumento := strings.TrimSpace(genericStringValue(payload["fecha_documento"]))
	if fechaDocumento == "" {
		fechaDocumento = strings.TrimSpace(genericStringValue(pedido["fecha_pedido"]))
	}
	if fechaDocumento == "" {
		fechaDocumento = time.Now().In(time.Local).Format("2006-01-02")
	}

	observaciones := strings.TrimSpace(genericStringValue(payload["observaciones"]))
	if observaciones == "" {
		observaciones = "Documento final generado desde pedido " + genericStringValue(pedido["codigo"])
	}

	docPayload := dbpkg.EmpresaDocumentoFacturacion{
		EmpresaID:            empresaID,
		TipoDocumento:        tipoDocumento,
		DocumentoCodigo:      documentoCodigo,
		EstadoDocumento:      estadoDocumento,
		EstadoAnterior:       "",
		EventoUltimo:         "convertido_desde_pedido",
		PeriodoContable:      strings.TrimSpace(genericStringValue(payload["periodo_contable"])),
		MontoTotal:           ventasAnyToFloat64(pedido["total"]),
		Moneda:               strings.TrimSpace(genericStringDefault(pedido["moneda"], "COP")),
		FechaDocumento:       fechaDocumento,
		EntidadRelacionadaID: pedidoID,
		UsuarioCreador:       actor,
		Estado:               "activo",
		Observaciones:        observaciones,
	}

	documentoCreado := docExistente == nil
	if docExistente != nil {
		docPayload.EstadoAnterior = strings.TrimSpace(docExistente.EstadoDocumento)
	}

	documentoFinal, err := dbpkg.UpsertEmpresaDocumentoFacturacion(dbEmp, docPayload)
	if err != nil {
		return nil, nil, false, err
	}

	if hasAllowedColumn(cfgPedidosVenta.AllowedColumns, "observaciones") {
		updatePedido := map[string]interface{}{
			"observaciones": appendGenericObservation(
				genericStringValue(pedido["observaciones"]),
				"documento final "+documentoFinal.DocumentoCodigo+" ("+documentoFinal.TipoDocumento+") generado",
			),
		}
		_ = dbpkg.UpdateEmpresaGenericRow(dbEmp, cfgPedidosVenta.Table, empresaID, pedidoID, updatePedido, cfgPedidosVenta.AllowedColumns)
	}

	pedido, _ = dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgPedidosVenta.Table, empresaID, pedidoID)
	return pedido, documentoFinal, documentoCreado, nil
}

func buildVentasEmbudoConversionSnapshot(dbEmp *sql.DB, empresaID int64, desde, hasta string, slaCotizacionHoras, slaPedidoHoras int, includeInactive bool, maxRows int) (ventasEmbudoSnapshot, error) {
	if empresaID <= 0 {
		return ventasEmbudoSnapshot{}, fmt.Errorf("empresa_id required")
	}
	if maxRows <= 0 {
		maxRows = 200
	}
	if maxRows > 1000 {
		maxRows = 1000
	}
	if slaCotizacionHoras <= 0 {
		slaCotizacionHoras = 48
	}
	if slaPedidoHoras <= 0 {
		slaPedidoHoras = 72
	}

	if err := dbpkg.EnsureEmpresaModulosFaltantesSchema(dbEmp); err != nil {
		return ventasEmbudoSnapshot{}, err
	}
	if err := dbpkg.EnsureEmpresaDocumentosTransaccionalesSchema(dbEmp); err != nil {
		return ventasEmbudoSnapshot{}, err
	}

	query := `SELECT
		id,
		COALESCE(codigo, ''),
		COALESCE(fecha_documento, ''),
		COALESCE(vigencia_hasta, ''),
		COALESCE(estado_documento, 'borrador'),
		COALESCE(total, 0),
		COALESCE(convertido_pedido_id, 0)
	FROM empresa_cotizaciones_venta
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	if !includeInactive {
		query += ` AND LOWER(COALESCE(estado, 'activo')) = 'activo'`
	}
	if strings.TrimSpace(desde) != "" {
		query += ` AND substr(COALESCE(fecha_documento, ''), 1, 10) >= ?`
		args = append(args, strings.TrimSpace(desde))
	}
	if strings.TrimSpace(hasta) != "" {
		query += ` AND substr(COALESCE(fecha_documento, ''), 1, 10) <= ?`
		args = append(args, strings.TrimSpace(hasta))
	}
	query += ` ORDER BY id DESC LIMIT ?`
	args = append(args, maxRows)

	rows, err := dbEmp.Query(query, args...)
	if err != nil {
		return ventasEmbudoSnapshot{}, err
	}

	type cotizacionEmbudoRow struct {
		ID          int64
		Codigo      string
		Fecha       string
		Vigencia    string
		Estado      string
		Total       float64
		PedidoRefID int64
	}

	cotizaciones := make([]cotizacionEmbudoRow, 0, maxRows)
	for rows.Next() {
		var item cotizacionEmbudoRow
		if err := rows.Scan(
			&item.ID,
			&item.Codigo,
			&item.Fecha,
			&item.Vigencia,
			&item.Estado,
			&item.Total,
			&item.PedidoRefID,
		); err != nil {
			_ = rows.Close()
			return ventasEmbudoSnapshot{}, err
		}
		cotizaciones = append(cotizaciones, item)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return ventasEmbudoSnapshot{}, err
	}
	_ = rows.Close()

	now := time.Now().In(time.Local)
	snapshot := ventasEmbudoSnapshot{
		Rows:    make([]map[string]interface{}, 0),
		Summary: make(map[string]interface{}),
		Alertas: make([]map[string]interface{}, 0),
	}

	pedidosSet := make(map[int64]bool)
	pedidosConDocumentoSet := make(map[int64]bool)

	var cotizacionesTotal int64
	var cotizacionesConvertidasPedido int64
	var cotizacionesDocumentoFinal int64
	var alertasCotizacionSLA int64
	var alertasCotizacionVigencia int64
	var alertasPedidoSLA int64

	for _, cotizacion := range cotizaciones {
		cotizacionID := cotizacion.ID
		cotizacionCodigo := cotizacion.Codigo
		fechaCotizacion := cotizacion.Fecha
		vigenciaHasta := cotizacion.Vigencia
		estadoCotizacion := cotizacion.Estado
		totalCotizacion := cotizacion.Total
		pedidoRefID := cotizacion.PedidoRefID

		cotizacionesTotal++
		estadoCotizacion = normalizeStateMachineValue(estadoCotizacion)

		horasCotizacion := ventasElapsedHoursSince(fechaCotizacion, now)
		pedidoRow, err := resolvePedidoForCotizacion(dbEmp, empresaID, cotizacionID, pedidoRefID)
		if err != nil {
			return ventasEmbudoSnapshot{}, err
		}

		pedidoID := int64(0)
		pedidoCodigo := ""
		pedidoEstado := ""
		fechaPedido := ""
		totalPedido := float64(0)
		horasPedido := int64(0)
		if pedidoRow != nil {
			pedidoID = anyToInt64(pedidoRow["id"])
			pedidoCodigo = genericStringValue(pedidoRow["codigo"])
			pedidoEstado = normalizeStateMachineValue(genericStringValue(pedidoRow["estado_pedido"]))
			fechaPedido = genericStringValue(pedidoRow["fecha_pedido"])
			totalPedido = ventasAnyToFloat64(pedidoRow["total"])
			horasPedido = ventasElapsedHoursSince(fechaPedido, now)
			if pedidoID > 0 {
				pedidosSet[pedidoID] = true
				cotizacionesConvertidasPedido++
			}
		}

		docFinal, err := findDocumentoFacturacionByPedidoID(dbEmp, empresaID, pedidoID)
		if err != nil {
			return ventasEmbudoSnapshot{}, err
		}
		documentoFinalID := int64(0)
		documentoFinalCodigo := ""
		documentoFinalTipo := ""
		estadoDocumentoFinal := ""
		fechaDocumentoFinal := ""
		if docFinal != nil {
			documentoFinalID = docFinal.ID
			documentoFinalCodigo = docFinal.DocumentoCodigo
			documentoFinalTipo = docFinal.TipoDocumento
			estadoDocumentoFinal = normalizeStateMachineValue(docFinal.EstadoDocumento)
			fechaDocumentoFinal = docFinal.FechaDocumento
			if pedidoID > 0 {
				pedidosConDocumentoSet[pedidoID] = true
			}
			cotizacionesDocumentoFinal++
		}

		conversionEtapa := "cotizacion"
		if pedidoID > 0 {
			conversionEtapa = "pedido"
		}
		if documentoFinalID > 0 {
			conversionEtapa = "documento_final"
		}

		alertaTipos := make([]string, 0, 2)
		alertasMensajes := make([]string, 0, 2)

		if pedidoID == 0 {
			if ventasCotizacionSLAAplica(estadoCotizacion) && horasCotizacion >= int64(slaCotizacionHoras) {
				alertaTipos = append(alertaTipos, "cotizacion_sla_vencida")
				alertasMensajes = append(alertasMensajes, "Cotizacion supera SLA de conversion a pedido")
				alertasCotizacionSLA++
			}
			if ventasIsPastDueDate(vigenciaHasta, now) && estadoCotizacion != "anulada" {
				alertaTipos = append(alertaTipos, "cotizacion_vigencia_vencida")
				alertasMensajes = append(alertasMensajes, "Cotizacion vencida por vigencia")
				alertasCotizacionVigencia++
			}
		}

		if pedidoID > 0 && documentoFinalID == 0 && ventasPedidoSLAAplica(pedidoEstado) && horasPedido >= int64(slaPedidoHoras) {
			alertaTipos = append(alertaTipos, "pedido_sla_vencido")
			alertasMensajes = append(alertasMensajes, "Pedido supera SLA de conversion a documento final")
			alertasPedidoSLA++
		}

		alertaTipo := strings.Join(alertaTipos, ",")
		alerta := strings.Join(alertasMensajes, "; ")
		if alertaTipo != "" {
			snapshot.Alertas = append(snapshot.Alertas, map[string]interface{}{
				"alerta_tipo":       alertaTipo,
				"alerta":            alerta,
				"cotizacion_id":     cotizacionID,
				"cotizacion_codigo": cotizacionCodigo,
				"pedido_id":         pedidoID,
				"pedido_codigo":     pedidoCodigo,
				"estado_cotizacion": estadoCotizacion,
				"estado_pedido":     pedidoEstado,
				"fecha_cotizacion":  fechaCotizacion,
				"fecha_pedido":      fechaPedido,
				"vigencia_hasta":    vigenciaHasta,
				"horas_cotizacion":  horasCotizacion,
				"horas_pedido":      horasPedido,
				"conversion_etapa":  conversionEtapa,
			})
		}

		snapshot.Rows = append(snapshot.Rows, map[string]interface{}{
			"cotizacion_id":          cotizacionID,
			"cotizacion_codigo":      cotizacionCodigo,
			"fecha_cotizacion":       fechaCotizacion,
			"vigencia_hasta":         vigenciaHasta,
			"estado_cotizacion":      estadoCotizacion,
			"total_cotizacion":       reportesRound(totalCotizacion),
			"pedido_id":              pedidoID,
			"pedido_codigo":          pedidoCodigo,
			"fecha_pedido":           fechaPedido,
			"estado_pedido":          pedidoEstado,
			"total_pedido":           reportesRound(totalPedido),
			"documento_final_id":     documentoFinalID,
			"documento_final_codigo": documentoFinalCodigo,
			"documento_final_tipo":   documentoFinalTipo,
			"estado_documento_final": estadoDocumentoFinal,
			"fecha_documento_final":  fechaDocumentoFinal,
			"horas_desde_cotizacion": horasCotizacion,
			"horas_desde_pedido":     horasPedido,
			"conversion_etapa":       conversionEtapa,
			"alerta_tipo":            alertaTipo,
			"alerta":                 alerta,
		})
	}

	pedidosTotal := int64(len(pedidosSet))
	pedidosConDocumento := int64(len(pedidosConDocumentoSet))

	conversionCotizacionPedidoPct := 0.0
	if cotizacionesTotal > 0 {
		conversionCotizacionPedidoPct = reportesRound((float64(cotizacionesConvertidasPedido) * 100.0) / float64(cotizacionesTotal))
	}

	conversionPedidoDocumentoPct := 0.0
	if pedidosTotal > 0 {
		conversionPedidoDocumentoPct = reportesRound((float64(pedidosConDocumento) * 100.0) / float64(pedidosTotal))
	}

	conversionTotalPct := 0.0
	if cotizacionesTotal > 0 {
		conversionTotalPct = reportesRound((float64(cotizacionesDocumentoFinal) * 100.0) / float64(cotizacionesTotal))
	}

	snapshot.Summary["cotizaciones_total"] = cotizacionesTotal
	snapshot.Summary["cotizaciones_convertidas_pedido"] = cotizacionesConvertidasPedido
	snapshot.Summary["cotizaciones_documento_final"] = cotizacionesDocumentoFinal
	snapshot.Summary["pedidos_total"] = pedidosTotal
	snapshot.Summary["pedidos_con_documento_final"] = pedidosConDocumento
	snapshot.Summary["conversion_cotizacion_pedido_pct"] = conversionCotizacionPedidoPct
	snapshot.Summary["conversion_pedido_documento_pct"] = conversionPedidoDocumentoPct
	snapshot.Summary["conversion_total_pct"] = conversionTotalPct
	snapshot.Summary["alertas_total"] = int64(len(snapshot.Alertas))
	snapshot.Summary["alertas_cotizacion_sla"] = alertasCotizacionSLA
	snapshot.Summary["alertas_cotizacion_vigencia"] = alertasCotizacionVigencia
	snapshot.Summary["alertas_pedido_sla"] = alertasPedidoSLA
	snapshot.Summary["sla_cotizacion_horas"] = slaCotizacionHoras
	snapshot.Summary["sla_pedido_horas"] = slaPedidoHoras
	snapshot.Summary["rango_desde"] = reportesFirstNonBlank(strings.TrimSpace(desde), "sin_desde")
	snapshot.Summary["rango_hasta"] = reportesFirstNonBlank(strings.TrimSpace(hasta), "sin_hasta")

	return snapshot, nil
}

func resolvePedidoForCotizacion(dbEmp *sql.DB, empresaID, cotizacionID, pedidoRefID int64) (map[string]interface{}, error) {
	if pedidoRefID > 0 {
		item, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgPedidosVenta.Table, empresaID, pedidoRefID)
		if err == nil {
			return item, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	return findPedidoByCotizacionID(dbEmp, empresaID, cotizacionID)
}

func findPedidoByCotizacionID(dbEmp *sql.DB, empresaID, cotizacionID int64) (map[string]interface{}, error) {
	if empresaID <= 0 || cotizacionID <= 0 {
		return nil, nil
	}

	var pedidoID int64
	err := dbEmp.QueryRow(`SELECT id
	FROM empresa_pedidos_venta
	WHERE empresa_id = ?
	  AND COALESCE(cotizacion_id, 0) = ?
	  AND LOWER(COALESCE(estado, 'activo')) = 'activo'
	ORDER BY id DESC
	LIMIT 1`, empresaID, cotizacionID).Scan(&pedidoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if pedidoID <= 0 {
		return nil, nil
	}

	item, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgPedidosVenta.Table, empresaID, pedidoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func findDocumentoFacturacionByPedidoID(dbEmp *sql.DB, empresaID, pedidoID int64) (*dbpkg.EmpresaDocumentoFacturacion, error) {
	if empresaID <= 0 || pedidoID <= 0 {
		return nil, nil
	}

	var tipoDocumento string
	var documentoCodigo string
	err := dbEmp.QueryRow(`SELECT
		COALESCE(tipo_documento, 'factura_electronica'),
		COALESCE(documento_codigo, '')
	FROM empresa_facturacion_documentos
	WHERE empresa_id = ?
	  AND COALESCE(entidad_relacionada_id, 0) = ?
	  AND LOWER(COALESCE(estado, 'activo')) = 'activo'
	ORDER BY id DESC
	LIMIT 1`, empresaID, pedidoID).Scan(&tipoDocumento, &documentoCodigo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if strings.TrimSpace(documentoCodigo) == "" {
		return nil, nil
	}

	item, err := dbpkg.GetEmpresaDocumentoFacturacionByCodigo(dbEmp, empresaID, tipoDocumento, documentoCodigo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func ventasAnyToFloat64(v interface{}) float64 {
	switch value := v.(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int32:
		return float64(value)
	case int64:
		return float64(value)
	case json.Number:
		f, _ := value.Float64()
		return f
	case string:
		f, _ := strconv.ParseFloat(strings.TrimSpace(value), 64)
		return f
	default:
		return 0
	}
}

func ventasParseDateTime(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}

	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02",
	}

	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return parsed.In(time.Local), true
		}
	}
	return time.Time{}, false
}

func ventasElapsedHoursSince(raw string, now time.Time) int64 {
	parsed, ok := ventasParseDateTime(raw)
	if !ok {
		return 0
	}
	hours := int64(now.Sub(parsed).Hours())
	if hours < 0 {
		return 0
	}
	return hours
}

func ventasIsPastDueDate(raw string, now time.Time) bool {
	parsed, ok := ventasParseDateTime(raw)
	if !ok {
		return false
	}
	if len(strings.TrimSpace(raw)) <= 10 {
		parsed = parsed.Add(24*time.Hour - time.Second)
	}
	return now.After(parsed)
}

func ventasCotizacionSLAAplica(estado string) bool {
	estado = normalizeStateMachineValue(estado)
	switch estado {
	case "emitida", "aprobada":
		return true
	default:
		return false
	}
}

func ventasPedidoSLAAplica(estado string) bool {
	estado = normalizeStateMachineValue(estado)
	switch estado {
	case "", "cerrado", "cancelado", "devuelto":
		return false
	default:
		return true
	}
}

func ventasSanitizeCodeToken(raw string) string {
	raw = strings.ToUpper(strings.TrimSpace(raw))
	if raw == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range raw {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

func appendGenericObservation(previous, note string) string {
	previous = strings.TrimSpace(previous)
	note = strings.TrimSpace(note)
	if note == "" {
		return previous
	}
	entry := time.Now().In(time.Local).Format("2006-01-02 15:04:05") + " " + note
	if previous == "" {
		return entry
	}
	return previous + " | " + entry
}

func ventasFirstNonBlank(values ...string) string {
	for _, v := range values {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
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
				if errors.Is(err, dbpkg.ErrPeriodoFinancieroCerrado) {
					http.Error(w, "el periodo contable del registro esta cerrado", http.StatusConflict)
					return
				}
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
					if errors.Is(err, dbpkg.ErrPeriodoFinancieroCerrado) {
						http.Error(w, "el periodo contable del registro esta cerrado", http.StatusConflict)
						return
					}
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
				if errors.Is(err, dbpkg.ErrPeriodoFinancieroCerrado) {
					http.Error(w, "el periodo contable del registro esta cerrado", http.StatusConflict)
					return
				}
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
				if errors.Is(err, dbpkg.ErrPeriodoFinancieroCerrado) {
					http.Error(w, "el periodo contable del registro esta cerrado", http.StatusConflict)
					return
				}
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

		case "rotar_credencial", "rotar_credenciales":
			if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleIntegracionesRotarCredencialAction(dbEmp, cfg, ops, w, r)
			return

		case "monitoreo", "alertas":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handleIntegracionesMonitoreoAction(dbEmp, cfg, ops, w, r)
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

func handleIntegracionesRotarCredencialAction(dbEmp *sql.DB, cfg empresaModuloGenericConfig, ops empresaModuloIntegracionesOpsConfig, w http.ResponseWriter, r *http.Request) {
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

	credentialField := resolveIntegracionCredentialField(cfg, ops)
	if credentialField == "" {
		http.Error(w, "el modulo no soporta rotacion de credenciales", http.StatusBadRequest)
		return
	}

	credentialRefRaw := resolveIntegracionCredentialReferenceFromPayload(payload, r, credentialField)
	credentialRef, err := validateIntegracionCredentialReference(credentialRefRaw)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	currentRow, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfg.Table, empresaID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "registro no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo consultar integracion", http.StatusInternalServerError)
		return
	}

	actualRef := strings.TrimSpace(genericStringValue(currentRow[credentialField]))
	if strings.EqualFold(actualRef, credentialRef) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":                true,
			"empresa_id":        empresaID,
			"id":                id,
			"accion":            "rotar_credencial",
			"campo_credencial":  credentialField,
			"rotada":            false,
			"referencia_actual": actualRef,
			"item":              currentRow,
		})
		return
	}

	executedAt := time.Now().Format("2006-01-02 15:04:05")
	updatePayload := map[string]interface{}{
		credentialField:      credentialRef,
		"estado_integracion": "inactiva",
	}
	if strings.TrimSpace(ops.LastSyncField) != "" {
		updatePayload[ops.LastSyncField] = ""
	}
	if strings.TrimSpace(ops.ResponseField) != "" {
		snapshot, _ := json.Marshal(map[string]interface{}{
			"checked_at":         executedAt,
			"endpoint":           normalizeIntegracionEndpoint(genericStringValue(currentRow[ops.EndpointField])),
			"estado_integracion": "inactiva",
			"message":            "referencia de credencial rotada; ejecutar health_check o sync_manual",
			"action":             "rotar_credencial",
		})
		updatePayload[ops.ResponseField] = string(snapshot)
	}
	if hasAllowedColumn(cfg.AllowedColumns, "observaciones") {
		obs := strings.TrimSpace(genericStringValue(currentRow["observaciones"]))
		audit := fmt.Sprintf("[%s] referencia de credencial rotada en %s por %s", executedAt, credentialField, finanzasFirstNonBlank(strings.TrimSpace(adminEmailFromRequest(r)), "sistema"))
		if obs == "" {
			updatePayload["observaciones"] = audit
		} else {
			updatePayload["observaciones"] = obs + "\n" + audit
		}
	}

	if err := dbpkg.UpdateEmpresaGenericRow(dbEmp, cfg.Table, empresaID, id, updatePayload, cfg.AllowedColumns); err != nil {
		http.Error(w, "No se pudo rotar la referencia de credencial", http.StatusBadRequest)
		return
	}

	updated, _ := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfg.Table, empresaID, id)
	validation := map[string]interface{}{}

	if parseBoolQuery(r, "validar") {
		probe := buildIntegracionProbeResult(updated, ops)
		validation = map[string]interface{}{
			"endpoint":           probe.Endpoint,
			"http_status":        probe.HTTPStatus,
			"reachable":          probe.Reachable,
			"latency_ms":         probe.LatencyMS,
			"estado_integracion": probe.EstadoIntegracion,
			"message":            probe.Message,
		}

		validatePayload := map[string]interface{}{
			"estado_integracion": probe.EstadoIntegracion,
		}
		if strings.TrimSpace(ops.LastSyncField) != "" {
			validatePayload[ops.LastSyncField] = executedAt
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
				"action":             "rotar_credencial_validar",
			})
			validatePayload[ops.ResponseField] = string(snapshot)
		}
		_ = dbpkg.UpdateEmpresaGenericRow(dbEmp, cfg.Table, empresaID, id, validatePayload, cfg.AllowedColumns)
		updated, _ = dbpkg.GetEmpresaGenericRowByID(dbEmp, cfg.Table, empresaID, id)
	}

	response := map[string]interface{}{
		"ok":               true,
		"empresa_id":       empresaID,
		"id":               id,
		"modulo":           ops.ModuleName,
		"accion":           "rotar_credencial",
		"campo_credencial": credentialField,
		"rotada":           true,
		"item":             updated,
	}
	if len(validation) > 0 {
		response["validacion"] = validation
	}

	writeJSON(w, http.StatusOK, response)
}

func handleIntegracionesMonitoreoAction(dbEmp *sql.DB, cfg empresaModuloGenericConfig, ops empresaModuloIntegracionesOpsConfig, w http.ResponseWriter, r *http.Request) {
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

	latencyAlertMS, err := parseIntQueryOptional(r, "latencia_alerta_ms")
	if err != nil {
		http.Error(w, "latencia_alerta_ms invalido", http.StatusBadRequest)
		return
	}
	if latencyAlertMS <= 0 {
		latencyAlertMS = 2500
	}

	staleHours, err := parseIntQueryOptional(r, "stale_hours")
	if err != nil {
		http.Error(w, "stale_hours invalido", http.StatusBadRequest)
		return
	}
	if staleHours <= 0 {
		staleHours = 24
	}

	includeInactive := parseBoolQuery(r, "include_inactive")
	persistir := parseBoolQuery(r, "persistir")

	rows, err := loadEmpresaRowsForAction(dbEmp, cfg, empresaID, id, includeInactive, strings.TrimSpace(r.URL.Query().Get("q")), limit, offset)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "registro no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo preparar monitoreo de integraciones", http.StatusInternalServerError)
		return
	}

	executedAt := time.Now().In(time.Local)
	executedAtStr := executedAt.Format("2006-01-02 15:04:05")
	items := make([]map[string]interface{}, 0, len(rows))
	alertas := make([]map[string]interface{}, 0)
	erroresPersistencia := make([]map[string]interface{}, 0)
	erroresConectividad := int64(0)
	saludables := int64(0)

	for _, row := range rows {
		probe := buildIntegracionProbeResult(row, ops)
		estadoActual := normalizeStateMachineValue(genericStringDefault(row["estado_integracion"], "inactiva"))
		ultimaSync := ""
		if strings.TrimSpace(ops.LastSyncField) != "" {
			ultimaSync = strings.TrimSpace(genericStringValue(row[ops.LastSyncField]))
		}

		horasDesdeSync := -1.0
		if ultimaSync != "" {
			if parsed, ok := ventasParseDateTime(ultimaSync); ok {
				horasDesdeSync = reportesRound(executedAt.Sub(parsed).Hours())
			}
		}

		itemAlertas := make([]map[string]interface{}, 0)
		appendAlert := func(tipo, severidad, mensaje string) {
			alerta := map[string]interface{}{
				"id":        probe.ID,
				"codigo":    probe.Codigo,
				"nombre":    probe.Nombre,
				"tipo":      tipo,
				"severidad": severidad,
				"mensaje":   mensaje,
			}
			itemAlertas = append(itemAlertas, alerta)
			alertas = append(alertas, alerta)
		}

		if probe.Endpoint == "" {
			appendAlert("endpoint_invalido", "alta", "Endpoint vacio o invalido para el conector")
		}
		if !probe.Reachable {
			erroresConectividad++
			appendAlert("sin_conectividad", "alta", "No se logro conectar con el endpoint configurado")
		}
		if probe.Reachable && probe.LatencyMS > int64(latencyAlertMS) {
			appendAlert("latencia_alta", "media", "La latencia del conector supera el umbral permitido")
		}
		if ultimaSync == "" {
			appendAlert("sin_sync_reciente", "media", "No hay registro de sincronizacion reciente")
		} else if horasDesdeSync >= 0 && horasDesdeSync > float64(staleHours) {
			appendAlert("sync_atrasada", "media", "La ultima sincronizacion excede el umbral de antiguedad")
		}
		if estadoActual == "error" {
			appendAlert("estado_error", "media", "El conector figura en estado de error")
		}

		if persistir && probe.ID > 0 {
			updatePayload := map[string]interface{}{
				"estado_integracion": probe.EstadoIntegracion,
			}
			if strings.TrimSpace(ops.LastSyncField) != "" {
				updatePayload[ops.LastSyncField] = executedAtStr
			}
			if strings.TrimSpace(ops.ResponseField) != "" {
				snapshot, _ := json.Marshal(map[string]interface{}{
					"checked_at":         executedAtStr,
					"endpoint":           probe.Endpoint,
					"http_status":        probe.HTTPStatus,
					"reachable":          probe.Reachable,
					"latency_ms":         probe.LatencyMS,
					"estado_integracion": probe.EstadoIntegracion,
					"message":            probe.Message,
					"action":             "monitoreo",
				})
				updatePayload[ops.ResponseField] = string(snapshot)
			}
			if err := dbpkg.UpdateEmpresaGenericRow(dbEmp, cfg.Table, empresaID, probe.ID, updatePayload, cfg.AllowedColumns); err != nil {
				erroresPersistencia = append(erroresPersistencia, map[string]interface{}{
					"id":    probe.ID,
					"error": err.Error(),
				})
			}
		}

		if len(itemAlertas) == 0 && probe.Reachable {
			saludables++
		}

		item := map[string]interface{}{
			"id":                    probe.ID,
			"codigo":                probe.Codigo,
			"nombre":                probe.Nombre,
			"endpoint":              probe.Endpoint,
			"http_status":           probe.HTTPStatus,
			"reachable":             probe.Reachable,
			"latency_ms":            probe.LatencyMS,
			"estado_actual":         estadoActual,
			"estado_integracion":    probe.EstadoIntegracion,
			"message":               probe.Message,
			"ultima_sincronizacion": ultimaSync,
			"alertas":               itemAlertas,
		}
		if horasDesdeSync >= 0 {
			item["horas_desde_sync"] = horasDesdeSync
		}
		items = append(items, item)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                 len(erroresPersistencia) == 0,
		"empresa_id":         empresaID,
		"modulo":             ops.ModuleName,
		"accion":             "monitoreo",
		"ejecutado_en":       executedAtStr,
		"latencia_alerta_ms": latencyAlertMS,
		"stale_hours":        staleHours,
		"persistir":          persistir,
		"resumen": map[string]interface{}{
			"conectores_total":      int64(len(rows)),
			"conectores_saludables": saludables,
			"errores_conectividad":  erroresConectividad,
			"alertas_total":         int64(len(alertas)),
		},
		"items":                items,
		"alertas":              alertas,
		"errores_persistencia": erroresPersistencia,
	})
}

func resolveIntegracionCredentialField(cfg empresaModuloGenericConfig, ops empresaModuloIntegracionesOpsConfig) string {
	if strings.TrimSpace(ops.CredentialField) != "" {
		return strings.TrimSpace(ops.CredentialField)
	}
	if hasAllowedColumn(cfg.AllowedColumns, "api_key_ref") {
		return "api_key_ref"
	}
	if hasAllowedColumn(cfg.AllowedColumns, "credencial_ref") {
		return "credencial_ref"
	}
	return ""
}

func resolveIntegracionCredentialReferenceFromPayload(payload map[string]interface{}, r *http.Request, credentialField string) string {
	if payload != nil {
		candidates := []string{}
		if credentialField != "" {
			candidates = append(candidates, genericStringValue(payload[credentialField]))
		}
		candidates = append(candidates,
			genericStringValue(payload["nueva_credencial_ref"]),
			genericStringValue(payload["credencial_ref"]),
			genericStringValue(payload["api_key_ref"]),
		)
		for _, candidate := range candidates {
			if trimmed := strings.TrimSpace(candidate); trimmed != "" {
				return trimmed
			}
		}
	}

	queryCandidates := []string{
		strings.TrimSpace(r.URL.Query().Get("nueva_credencial_ref")),
		strings.TrimSpace(r.URL.Query().Get("credencial_ref")),
	}
	if credentialField != "" {
		queryCandidates = append([]string{strings.TrimSpace(r.URL.Query().Get(credentialField))}, queryCandidates...)
	}
	for _, candidate := range queryCandidates {
		if candidate != "" {
			return candidate
		}
	}

	return ""
}

func validateIntegracionCredentialReference(raw string) (string, error) {
	ref := strings.TrimSpace(raw)
	if ref == "" {
		return "", fmt.Errorf("nueva_credencial_ref requerida")
	}
	if len(ref) > 320 {
		return "", fmt.Errorf("nueva_credencial_ref excede longitud permitida")
	}
	if strings.ContainsAny(ref, " \t\r\n") {
		return "", fmt.Errorf("nueva_credencial_ref no puede contener espacios")
	}

	lower := strings.ToLower(ref)
	allowedPrefixes := []string{
		"env:",
		"file:",
		"vault:",
		"secret:",
		"kms:",
		"keyring:",
		"ref:",
		"base64:",
		"azurekeyvault:",
		"awssecrets:",
		"gcpsecret:",
	}

	allowed := false
	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(lower, prefix) {
			allowed = true
			break
		}
	}
	if !allowed {
		return "", fmt.Errorf("referencia insegura: use prefijos seguros como env:, file:, vault:, secret:, kms:, keyring: o ref:")
	}

	parts := strings.SplitN(ref, ":", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
		return "", fmt.Errorf("nueva_credencial_ref invalida: debe incluir identificador despues del prefijo")
	}

	return ref, nil
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
func EmpresaDIANColombiaHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	base := empresaModuloGenericCRUDHandler(dbEmp, cfgDIAN)
	return func(w http.ResponseWriter, r *http.Request) {
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch action {
		case "guia_onboarding", "ayuda_configuracion", "onboarding_empresa":
			empresaID, _ := parseInt64QueryOptional(r, "empresa_id")
			if empresaID <= 0 {
				empresaID = parseEmpresaIDFromContext(r)
			}
			cfg := map[string]interface{}{}
			if empresaID > 0 {
				cfg, _ = getEmpresaDIANConfig(dbEmp, empresaID)
			}
			writeJSON(w, http.StatusOK, buildDIANOnboardingGuide(cfg, empresaID))
			return

		case "validar_credenciales", "validar_secretos":
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
			cfg, _ := getEmpresaDIANConfig(dbEmp, empresaID)
			response, status, err := validateDIANCredentialRefs(cfg, empresaID, payload)
			if err != nil {
				http.Error(w, err.Error(), status)
				return
			}
			writeJSON(w, status, response)
			return

		case "subir_firma", "upload_firma":
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			response, status, err := uploadDIANCompanySignature(dbEmp, dbSuper, r)
			if err != nil {
				http.Error(w, err.Error(), status)
				return
			}
			writeJSON(w, status, response)
			return

		case "vencimiento_certificado", "verificar_vencimiento_certificado", "alerta_certificado":
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			cfg, _ := getEmpresaDIANConfig(dbEmp, empresaID)
			if len(cfg) == 0 {
				http.Error(w, "configuracion DIAN no existe para la empresa; registre base DIAN primero", http.StatusBadRequest)
				return
			}
			notificar := true
			if raw := strings.TrimSpace(r.URL.Query().Get("notificar")); raw != "" {
				notificar = parseTruthy(raw)
			}
			response := checkDIANCertificateExpiry(dbEmp, dbSuper, empresaID, cfg, notificar)
			writeJSON(w, http.StatusOK, response)
			return

		case "checklist", "validar":
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			cfg, _ := getEmpresaDIANConfig(dbEmp, empresaID)
			missing := missingDIANFields(cfg)
			effectiveSoftwareID, _, useSharedSoftware, softwareErr := resolveDIANSoftwareCredentials(cfg, nil)
			response := map[string]interface{}{
				"ok":                len(missing) == 0,
				"empresa_id":        empresaID,
				"faltantes":         missing,
				"pasos_minimos":     dianChecklistSteps(),
				"ambiente_sugerido": chooseDIANAmbiente(cfg),
				"software_modo":     map[bool]string{true: "compartido", false: "empresa"}[useSharedSoftware],
				"software_id":       effectiveSoftwareID,
			}
			if softwareErr != nil {
				response["software_error"] = softwareErr.Error()
			}
			if action == "validar" {
				response["recomendaciones"] = []string{
					"Definir modo DIAN por empresa: software compartido (SaaS) o software propio por empresa.",
					"Si usa software compartido, configurar software_id_compartido_ref/software_pin_compartido_ref o variables DIAN_SHARED_SOFTWARE_ID/DIAN_SHARED_SOFTWARE_PIN.",
					"Si no usa software compartido, validar que software_id/software_pin de la empresa coincidan con la plataforma DIAN.",
					"Confirmar rango de numeracion vigente y consecutivo dentro del rango.",
					"Mantener certificado, token y PIN fuera del codigo fuente (referencias seguras por empresa).",
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
			softwareID, softwarePIN, useSharedSoftware, err := resolveDIANSoftwareCredentials(cfg, payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			seed := nit + "|" + documento + "|" + fecha + "|" + total + "|" + softwareID + "|" + softwarePIN
			sum := sha256.Sum256([]byte(seed))
			cufe := strings.ToUpper(hex.EncodeToString(sum[:]))

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":               true,
				"empresa_id":       empresaID,
				"documento_codigo": documento,
				"fecha_emision":    fecha,
				"total":            total,
				"software_modo":    map[bool]string{true: "compartido", false: "empresa"}[useSharedSoftware],
				"software_id":      softwareID,
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

		case "generar_xml_ubl_base":
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
			response, status, err := generateDIANUBLBase(cfg, empresaID, payload)
			if err != nil {
				http.Error(w, err.Error(), status)
				return
			}
			writeJSON(w, status, response)
			return

		case "firmar_xml_real":
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
			response, status, err := signDIANXMLReal(cfg, empresaID, payload)
			if err != nil {
				http.Error(w, err.Error(), status)
				return
			}
			writeJSON(w, status, response)
			return

		case "firmar_xml_xades_base":
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
			response, status, err := signDIANXMLXAdESBase(cfg, empresaID, payload)
			if err != nil {
				http.Error(w, err.Error(), status)
				return
			}
			writeJSON(w, status, response)
			return

		case "diagnostico_oficial", "diagnosticar_oficial":
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
			cfg, _ := getEmpresaDIANConfig(dbEmp, empresaID)
			response, status, err := buildDIANOfficialReadinessReport(cfg, empresaID)
			if err != nil {
				http.Error(w, err.Error(), status)
				return
			}
			writeJSON(w, status, response)
			return

		case "validar_documento_dian", "preflight_documento", "validar_previo_envio":
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
			cfg, _ := getEmpresaDIANConfig(dbEmp, empresaID)
			xmlPayload := dianFirstNonBlank(
				genericStringValue(payload["xml_firmado"]),
				genericStringValue(payload["xml_ubl_base"]),
				genericStringValue(payload["xml"]),
			)
			response := validateDIANDocumentPreflight(cfg, empresaID, payload, xmlPayload, genericStringDefault(payload["etapa"], "validacion_manual"))
			status := http.StatusOK
			if parseTruthy(genericStringValue(response["bloqueado"])) {
				status = http.StatusConflict
			}
			writeJSON(w, status, response)
			return

		case "enviar_documento_real":
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
			response, status, err := sendDIANDocumentoReal(dbEmp, cfg, empresaID, payload)
			if err != nil {
				http.Error(w, err.Error(), status)
				return
			}
			writeJSON(w, status, response)
			return

		case "consultar_acuse_real":
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
			cfg, _ := getEmpresaDIANConfig(dbEmp, empresaID)
			response, status, err := consultarDIANAcuseReal(dbEmp, cfg, empresaID, payload, r)
			if err != nil {
				http.Error(w, err.Error(), status)
				return
			}
			writeJSON(w, status, response)
			return

		case "historial_tracks", "historial_trackid", "track_history":
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			limit := 80
			if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
				if parsed, parseErr := strconv.Atoi(rawLimit); parseErr == nil {
					limit = parsed
				}
			}
			items, err := listDIANTrackHistory(dbEmp, empresaID, limit)
			if err != nil {
				http.Error(w, "no se pudo consultar historial TrackId DIAN", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":         true,
				"empresa_id": empresaID,
				"items":      items,
				"total":      len(items),
			})
			return

		case "reconexion_dian":
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
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
			cfg, _ := getEmpresaDIANConfig(dbEmp, empresaID)
			response, status, err := runDIANReconexion(dbEmp, cfg, empresaID, payload)
			if err != nil {
				http.Error(w, err.Error(), status)
				return
			}
			writeJSON(w, status, response)
			return

		case "pruebas_dian", "pruebas_habilitacion", "test_habilitacion":
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
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
			cfg, _ := getEmpresaDIANConfig(dbEmp, empresaID)
			response, status, err := runDIANPruebasHabilitacion(dbEmp, cfg, empresaID, payload)
			if err != nil {
				http.Error(w, err.Error(), status)
				return
			}
			writeJSON(w, status, response)
			return

		case "activar_produccion_local", "pasar_a_produccion_local":
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
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
			cfg, _ := getEmpresaDIANConfig(dbEmp, empresaID)
			response, status, err := activateDIANProductionLocal(dbEmp, cfg, empresaID, payload)
			if err != nil {
				http.Error(w, err.Error(), status)
				return
			}
			writeJSON(w, status, response)
			return

		case "enviar_set_pruebas", "enviar_set_habilitacion":
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
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
			cfg, _ := getEmpresaDIANConfig(dbEmp, empresaID)
			response, status, err := runDIANSetPruebasEnvio(dbEmp, cfg, empresaID, payload)
			if err != nil {
				http.Error(w, err.Error(), status)
				return
			}
			writeJSON(w, status, response)
			return
		}

		base.ServeHTTP(w, r)
	}
}

func dianNowLocal() string {
	return time.Now().In(time.Local).Format("2006-01-02 15:04:05")
}

const (
	dianOfficialSetFacturas     = 30
	dianOfficialSetNotasDebito  = 10
	dianOfficialSetNotasCredito = 10
	dianOfficialSetTotal        = dianOfficialSetFacturas + dianOfficialSetNotasDebito + dianOfficialSetNotasCredito
	dianCertificateAlertDays    = 30
)

func parseDIANStoredTime(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func dianCertificateAlertThresholdDays(cfg map[string]interface{}) int {
	days := int(anyToInt64(cfg["certificado_alerta_dias"]))
	if days <= 0 {
		return dianCertificateAlertDays
	}
	if days > 365 {
		return 365
	}
	return days
}

func dianCertificateDaysRemaining(now, notAfter time.Time) int {
	hours := notAfter.Sub(now).Hours()
	if hours >= 0 {
		return int(math.Ceil(hours / 24))
	}
	return -int(math.Ceil(math.Abs(hours) / 24))
}

func dianCertificateExpiryStatus(now, notBefore, notAfter time.Time, alertDays int) map[string]interface{} {
	if alertDays <= 0 {
		alertDays = dianCertificateAlertDays
	}
	if notAfter.IsZero() {
		return map[string]interface{}{
			"ok":              false,
			"configurado":     false,
			"mensaje":         "No hay certificado X.509 registrado para calcular vencimiento.",
			"alerta_dias":     alertDays,
			"dias_restantes":  nil,
			"proximo_vencer":  false,
			"vencido":         false,
			"requiere_alerta": false,
		}
	}
	days := dianCertificateDaysRemaining(now, notAfter)
	expired := now.After(notAfter)
	closeToExpiry := !expired && days <= alertDays
	status := "vigente"
	startDate := ""
	if !notBefore.IsZero() {
		startDate = notBefore.Format("2006-01-02")
	}
	message := fmt.Sprintf("Certificado vigente; vence el %s.", notAfter.Format("2006-01-02"))
	if expired {
		status = "vencido"
		message = fmt.Sprintf("Certificado vencido el %s.", notAfter.Format("2006-01-02"))
	} else if closeToExpiry {
		status = "proximo_a_vencer"
		message = fmt.Sprintf("Certificado proximo a vencer en %d dia(s).", days)
	}
	return map[string]interface{}{
		"ok":                   !expired,
		"configurado":          true,
		"estado":               status,
		"mensaje":              message,
		"fecha_inicio":         startDate,
		"fecha_vencimiento":    notAfter.Format("2006-01-02"),
		"vence_en":             notAfter.Format(time.RFC3339),
		"alerta_dias":          alertDays,
		"dias_restantes":       days,
		"proximo_a_vencer":     closeToExpiry,
		"proximo_vencer":       closeToExpiry,
		"vencido":              expired,
		"requiere_alerta":      expired || closeToExpiry,
		"validacion_x509":      true,
		"zona_horaria_sistema": time.Local.String(),
	}
}

func resolveDIANCertificateExpiryFromConfig(cfg map[string]interface{}) (time.Time, time.Time, string, error) {
	if cfg == nil {
		return time.Time{}, time.Time{}, "", fmt.Errorf("configuracion DIAN no disponible")
	}
	if notAfter, ok := parseDIANStoredTime(genericStringValue(cfg["certificado_vencimiento_en"])); ok {
		notBefore, _ := parseDIANStoredTime(genericStringValue(cfg["certificado_inicio_en"]))
		return notBefore, notAfter, "configuracion", nil
	}
	if notAfter, ok := parseDIANStoredTime(genericStringValue(cfg["certificado_vencimiento"])); ok {
		notBefore, _ := parseDIANStoredTime(genericStringValue(cfg["certificado_inicio"]))
		return notBefore, notAfter, "configuracion", nil
	}
	certRef := genericStringValue(cfg["certificado_url"])
	if certRef == "" {
		return time.Time{}, time.Time{}, "", nil
	}
	cert, err := parseDIANCertificate(certRef)
	if err != nil {
		return time.Time{}, time.Time{}, "certificado_url", err
	}
	return cert.NotBefore, cert.NotAfter, "certificado_x509", nil
}

func dianCertificateAlertRecentlySent(cfg map[string]interface{}, now time.Time) bool {
	lastRaw := genericStringValue(cfg["certificado_alerta_ultimo_envio"])
	last, ok := parseDIANStoredTime(lastRaw)
	if !ok {
		return false
	}
	return now.Sub(last) < 24*time.Hour
}

func sendDIANCertificateExpiryEmail(dbEmp, dbSuper *sql.DB, empresaID int64, cfg map[string]interface{}, status map[string]interface{}) (string, error) {
	if dbSuper == nil {
		return "", fmt.Errorf("configuracion de correo no disponible")
	}
	var empresa *dbpkg.Empresa
	var err error
	if dbEmp != nil {
		empresa, err = dbpkg.GetEmpresaByScopeID(dbEmp, empresaID)
	}
	if empresa == nil && dbSuper != nil {
		empresa, err = dbpkg.GetEmpresaByScopeID(dbSuper, empresaID)
	}
	if err != nil {
		return "", err
	}
	if empresa == nil {
		return "", fmt.Errorf("empresa no encontrada")
	}
	toEmail := strings.TrimSpace(empresa.UsuarioCreador)
	if toEmail == "" {
		return "", fmt.Errorf("la empresa no tiene correo administrador registrado")
	}
	if _, err := mail.ParseAddress(toEmail); err != nil {
		return toEmail, fmt.Errorf("correo administrador invalido")
	}
	empresaNombre := strings.TrimSpace(empresa.Nombre)
	if empresaNombre == "" {
		empresaNombre = fmt.Sprintf("Empresa %d", empresaID)
	}
	subject := "Certificado digital DIAN proximo a vencer - " + empresaNombre
	if parseTruthy(genericStringValue(status["vencido"])) {
		subject = "Certificado digital DIAN vencido - " + empresaNombre
	}
	body := fmt.Sprintf(
		"Hola,\n\nEl certificado digital usado para facturacion electronica DIAN de %s requiere revision.\n\nEstado: %s\nVence: %s\nDias restantes: %s\n\nIngresa a Administrar empresa > Facturacion electronica, carga una nueva firma/certificado y ejecuta la validacion DIAN antes de emitir documentos.\n\nPowerful Control System",
		empresaNombre,
		genericStringValue(status["estado"]),
		genericStringValue(status["fecha_vencimiento"]),
		genericStringValue(status["dias_restantes"]),
	)
	if err := sendEmpresaUsuarioGmailPlain(dbSuper, toEmail, subject, body); err != nil {
		gmailErr := err
		if fallbackErr := sendEmpresaUsuarioMailuPlain(dbSuper, toEmail, subject, body); fallbackErr == nil {
			return toEmail, nil
		}
		return toEmail, gmailErr
	}
	return toEmail, nil
}

func checkDIANCertificateExpiry(dbEmp, dbSuper *sql.DB, empresaID int64, cfg map[string]interface{}, notify bool) map[string]interface{} {
	now := time.Now()
	alertDays := dianCertificateAlertThresholdDays(cfg)
	notBefore, notAfter, source, err := resolveDIANCertificateExpiryFromConfig(cfg)
	status := dianCertificateExpiryStatus(now, notBefore, notAfter, alertDays)
	status["empresa_id"] = empresaID
	status["fuente"] = source
	status["correo_enviado"] = false
	status["correo_destino"] = genericStringValue(cfg["certificado_alerta_email"])
	if err != nil {
		status["ok"] = false
		status["certificado_error"] = dianTruncate(err.Error(), 180)
		return status
	}
	if !notify || !parseTruthy(genericStringValue(status["requiere_alerta"])) {
		return status
	}
	if dianCertificateAlertRecentlySent(cfg, now) {
		status["correo_omitido"] = "alerta enviada en las ultimas 24 horas"
		return status
	}
	toEmail, sendErr := sendDIANCertificateExpiryEmail(dbEmp, dbSuper, empresaID, cfg, status)
	if strings.TrimSpace(toEmail) != "" {
		status["correo_destino"] = toEmail
	}
	if sendErr != nil {
		status["email_error"] = dianTruncate(sendErr.Error(), 180)
		return status
	}
	status["correo_enviado"] = true
	status["certificado_alerta_ultimo_envio"] = dianNowLocal()
	_ = updateDIANConfigFields(dbEmp, empresaID, cfg, map[string]interface{}{
		"certificado_alerta_ultimo_envio": status["certificado_alerta_ultimo_envio"],
		"certificado_alerta_email":        toEmail,
	})
	return status
}

func dianDefaultSetRequirement() map[string]interface{} {
	return map[string]interface{}{
		"ambiente":               "habilitacion",
		"facturas_electronicas":  dianOfficialSetFacturas,
		"notas_debito":           dianOfficialSetNotasDebito,
		"notas_credito":          dianOfficialSetNotasCredito,
		"total_documentos":       dianOfficialSetTotal,
		"estado_requerido_final": "Aceptado",
		"nota":                   "Base operativa para software propio/proveedor segun el objetivo cargado en portal DIAN; verifique siempre el set exacto asignado a esta empresa.",
	}
}

func dianConfiguredSetCounts(cfg map[string]interface{}) (int, int, int, int) {
	facturas := int(anyToInt64(cfg["set_facturas_requeridas"]))
	notasDebito := int(anyToInt64(cfg["set_notas_debito_requeridas"]))
	notasCredito := int(anyToInt64(cfg["set_notas_credito_requeridas"]))
	total := int(anyToInt64(cfg["set_documentos_requeridos"]))
	if facturas < 0 {
		facturas = 0
	}
	if notasDebito < 0 {
		notasDebito = 0
	}
	if notasCredito < 0 {
		notasCredito = 0
	}
	suma := facturas + notasDebito + notasCredito
	if suma <= 0 {
		return dianOfficialSetFacturas, dianOfficialSetNotasDebito, dianOfficialSetNotasCredito, dianOfficialSetTotal
	}
	if total < suma {
		total = suma
	}
	if total == 0 {
		total = suma
	}
	return facturas, notasDebito, notasCredito, total
}

func dianConfiguredAcceptedCounts(cfg map[string]interface{}) (int, int, int, int) {
	facturas := int(anyToInt64(cfg["set_facturas_aceptadas_requeridas"]))
	notasDebito := int(anyToInt64(cfg["set_notas_debito_aceptadas_requeridas"]))
	notasCredito := int(anyToInt64(cfg["set_notas_credito_aceptadas_requeridas"]))
	total := int(anyToInt64(cfg["set_documentos_aceptados_requeridos"]))
	if facturas < 0 {
		facturas = 0
	}
	if notasDebito < 0 {
		notasDebito = 0
	}
	if notasCredito < 0 {
		notasCredito = 0
	}
	suma := facturas + notasDebito + notasCredito
	if total < suma {
		total = suma
	}
	return facturas, notasDebito, notasCredito, total
}

func dianEffectiveAcceptedCounts(cfg map[string]interface{}, payload map[string]interface{}) (int, int, int, int) {
	defaultFacturas, defaultDebito, defaultCredito, defaultTotal := dianConfiguredAcceptedCounts(cfg)
	facturas := dianPayloadNonNegativeInt(payload, defaultFacturas, "set_facturas_aceptadas_requeridas", "facturas_aceptadas_requeridas", "accepted_invoices_required")
	notasDebito := dianPayloadNonNegativeInt(payload, defaultDebito, "set_notas_debito_aceptadas_requeridas", "notas_debito_aceptadas_requeridas", "accepted_debit_notes_required")
	notasCredito := dianPayloadNonNegativeInt(payload, defaultCredito, "set_notas_credito_aceptadas_requeridas", "notas_credito_aceptadas_requeridas", "accepted_credit_notes_required")
	total := dianPayloadNonNegativeInt(payload, defaultTotal, "set_documentos_aceptados_requeridos", "documentos_aceptados_requeridos", "accepted_documents_required")
	suma := facturas + notasDebito + notasCredito
	if total < suma {
		total = suma
	}
	return facturas, notasDebito, notasCredito, total
}

func dianConfiguredSetRequirement(cfg map[string]interface{}) map[string]interface{} {
	facturas, notasDebito, notasCredito, totalDocumentos := dianConfiguredSetCounts(cfg)
	accFacturas, accDebito, accCredito, accTotal := dianConfiguredAcceptedCounts(cfg)
	return map[string]interface{}{
		"ambiente":                               "habilitacion",
		"modo_operacion":                         genericStringValue(cfg["modo_operacion_descripcion"]),
		"facturas_electronicas":                  facturas,
		"notas_debito":                           notasDebito,
		"notas_credito":                          notasCredito,
		"total_documentos":                       totalDocumentos,
		"facturas_electronicas_aceptadas_minimo": accFacturas,
		"notas_debito_aceptadas_minimo":          accDebito,
		"notas_credito_aceptadas_minimo":         accCredito,
		"total_documentos_aceptados_minimo":      accTotal,
		"estado_requerido_final":                 "Aceptado",
		"nota":                                   "Objetivo cargado desde el modo de operacion del portal DIAN para esta empresa.",
	}
}

func dianEffectiveSetRequirement(payload map[string]interface{}) map[string]interface{} {
	facturas := dianPayloadNonNegativeInt(payload, dianOfficialSetFacturas, "facturas_electronicas", "facturas", "invoices_total_required")
	notasDebito := dianPayloadNonNegativeInt(payload, dianOfficialSetNotasDebito, "notas_debito", "debit_notes", "total_debit_notes_required")
	notasCredito := dianPayloadNonNegativeInt(payload, dianOfficialSetNotasCredito, "notas_credito", "credit_notes", "total_credit_notes_required")
	totalDocumentos := facturas + notasDebito + notasCredito
	return map[string]interface{}{
		"ambiente":               "habilitacion",
		"facturas_electronicas":  facturas,
		"notas_debito":           notasDebito,
		"notas_credito":          notasCredito,
		"total_documentos":       totalDocumentos,
		"estado_requerido_final": "Aceptado",
		"nota":                   "Valores configurables segun el set que DIAN asigne a la empresa.",
	}
}

func dianEffectiveSetRequirementForConfig(cfg map[string]interface{}, payload map[string]interface{}) map[string]interface{} {
	defaultFacturas, defaultDebito, defaultCredito, _ := dianConfiguredSetCounts(cfg)
	facturas := dianPayloadNonNegativeInt(payload, defaultFacturas, "facturas_electronicas", "facturas", "invoices_total_required")
	notasDebito := dianPayloadNonNegativeInt(payload, defaultDebito, "notas_debito", "debit_notes", "total_debit_notes_required")
	notasCredito := dianPayloadNonNegativeInt(payload, defaultCredito, "notas_credito", "credit_notes", "total_credit_notes_required")
	totalDocumentos := facturas + notasDebito + notasCredito
	accFacturas, accDebito, accCredito, accTotal := dianEffectiveAcceptedCounts(cfg, payload)
	return map[string]interface{}{
		"ambiente":                               "habilitacion",
		"modo_operacion":                         genericStringValue(cfg["modo_operacion_descripcion"]),
		"facturas_electronicas":                  facturas,
		"notas_debito":                           notasDebito,
		"notas_credito":                          notasCredito,
		"total_documentos":                       totalDocumentos,
		"facturas_electronicas_aceptadas_minimo": accFacturas,
		"notas_debito_aceptadas_minimo":          accDebito,
		"notas_credito_aceptadas_minimo":         accCredito,
		"total_documentos_aceptados_minimo":      accTotal,
		"estado_requerido_final":                 "Aceptado",
		"nota":                                   "Valores configurables segun el set que DIAN asigne a la empresa.",
	}
}

func dianFirstNonBlank(values ...string) string {
	for _, v := range values {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

func dianTruncate(raw string, max int) string {
	raw = strings.TrimSpace(raw)
	if max <= 0 || len(raw) <= max {
		return raw
	}
	return strings.TrimSpace(raw[:max])
}

func resolveDIANSecretValue(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("referencia secreta DIAN no configurada")
	}

	lower := strings.ToLower(raw)
	switch {
	case strings.HasPrefix(lower, "env:"):
		key := strings.TrimSpace(raw[4:])
		if key == "" {
			return "", fmt.Errorf("referencia env invalida")
		}
		val := strings.TrimSpace(os.Getenv(key))
		if val == "" {
			return "", fmt.Errorf("variable de entorno %s vacia", key)
		}
		return val, nil

	case strings.HasPrefix(lower, "file:"):
		path := strings.TrimSpace(raw[5:])
		if path == "" {
			return "", fmt.Errorf("referencia file invalida")
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		val := strings.TrimSpace(string(content))
		if val == "" {
			return "", fmt.Errorf("archivo secreto vacio")
		}
		return val, nil

	case strings.HasPrefix(lower, "base64:"):
		encoded := strings.TrimSpace(raw[7:])
		if encoded == "" {
			return "", fmt.Errorf("referencia base64 invalida")
		}
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return "", err
		}
		val := strings.TrimSpace(string(decoded))
		if val == "" {
			return "", fmt.Errorf("valor base64 vacio")
		}
		return val, nil
	}

	return raw, nil
}

func parseDIANRSAPrivateKey(raw string) (*rsa.PrivateKey, error) {
	resolved, err := resolveDIANSecretValue(raw)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode([]byte(resolved))
	if block == nil {
		return nil, fmt.Errorf("llave privada DIAN no esta en formato PEM")
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("no se pudo parsear llave privada RSA")
	}

	rsaKey, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("la llave privada no es RSA")
	}
	return rsaKey, nil
}

func parseDIANRSAPrivateKeyBlock(block *pem.Block, password string) (*rsa.PrivateKey, error) {
	if block == nil {
		return nil, fmt.Errorf("bloque PEM vacio")
	}
	keyBytes := block.Bytes
	if x509.IsEncryptedPEMBlock(block) {
		if strings.TrimSpace(password) == "" {
			return nil, fmt.Errorf("la llave privada esta cifrada; ingrese la clave de la firma")
		}
		decrypted, err := x509.DecryptPEMBlock(block, []byte(password))
		if err != nil {
			return nil, fmt.Errorf("no se pudo descifrar la llave privada")
		}
		keyBytes = decrypted
	}
	if key, err := x509.ParsePKCS1PrivateKey(keyBytes); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("no se pudo parsear llave privada RSA")
	}
	rsaKey, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("la llave privada no es RSA")
	}
	return rsaKey, nil
}

func encodeDIANRSAPrivateKeyPEM(key *rsa.PrivateKey) string {
	if key == nil {
		return ""
	}
	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
	return string(pem.EncodeToMemory(block))
}

func encodeDIANCertificatePEM(cert *x509.Certificate) string {
	if cert == nil {
		return ""
	}
	block := &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}
	return string(pem.EncodeToMemory(block))
}

type dianSignatureUploadMaterial struct {
	PrivateKeyPEM  string
	CertificatePEM string
	Format         string
	Subject        string
	Issuer         string
	Serial         string
	NotBefore      time.Time
	NotAfter       time.Time
}

func setDIANSignatureCertificateMetadata(material *dianSignatureUploadMaterial, cert *x509.Certificate) {
	if material == nil || cert == nil {
		return
	}
	material.Subject = cert.Subject.String()
	material.Issuer = cert.Issuer.String()
	material.Serial = cert.SerialNumber.String()
	material.NotBefore = cert.NotBefore
	material.NotAfter = cert.NotAfter
}

func decodeDIANP12WithOpenSSL(contentBytes []byte, password, format string) (dianSignatureUploadMaterial, error) {
	var material dianSignatureUploadMaterial
	opensslPath, err := exec.LookPath("openssl")
	if err != nil {
		return material, err
	}
	tmp, err := os.CreateTemp("", "pcs-dian-firma-*.p12")
	if err != nil {
		return material, err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(contentBytes); err != nil {
		tmp.Close()
		return material, err
	}
	if err := tmp.Close(); err != nil {
		return material, err
	}
	_ = os.Chmod(tmpName, 0o600)

	var stdout bytes.Buffer
	var lastErr error
	for _, args := range [][]string{
		{"pkcs12", "-in", tmpName, "-nodes", "-passin", "env:PCS_DIAN_P12_PASSWORD"},
		{"pkcs12", "-legacy", "-in", tmpName, "-nodes", "-passin", "env:PCS_DIAN_P12_PASSWORD"},
	} {
		stdout.Reset()
		var stderr bytes.Buffer
		cmd := exec.Command(opensslPath, args...)
		cmd.Env = append(os.Environ(), "PCS_DIAN_P12_PASSWORD="+password)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			lastErr = err
			continue
		}
		lastErr = nil
		break
	}
	if lastErr != nil {
		return material, fmt.Errorf("no se pudo decodificar P12/PFX con OpenSSL")
	}
	material, err = decodeDIANSignatureUpload(stdout.Bytes(), "firma_convertida.pem", "")
	if err != nil {
		return material, err
	}
	if strings.TrimSpace(format) != "" {
		material.Format = strings.TrimSpace(format)
	}
	return material, nil
}

func decodeDIANSignatureUpload(contentBytes []byte, filename, password string) (dianSignatureUploadMaterial, error) {
	var material dianSignatureUploadMaterial
	name := strings.ToLower(strings.TrimSpace(filename))
	ext := strings.ToLower(filepath.Ext(name))
	if ext == ".p12" || ext == ".pfx" {
		privateKey, cert, err := pkcs12.Decode(contentBytes, password)
		if err == nil {
			rsaKey, ok := privateKey.(*rsa.PrivateKey)
			if !ok || rsaKey == nil {
				return material, fmt.Errorf("el P12/PFX no contiene llave privada RSA")
			}
			material.PrivateKeyPEM = encodeDIANRSAPrivateKeyPEM(rsaKey)
			material.CertificatePEM = encodeDIANCertificatePEM(cert)
			material.Format = strings.TrimPrefix(ext, ".")
			if cert != nil {
				setDIANSignatureCertificateMetadata(&material, cert)
			}
			return material, nil
		}

		blocks, pemErr := pkcs12.ToPEM(contentBytes, password)
		if pemErr != nil {
			opensslMaterial, opensslErr := decodeDIANP12WithOpenSSL(contentBytes, password, strings.TrimPrefix(ext, "."))
			if opensslErr == nil {
				return opensslMaterial, nil
			}
			return material, fmt.Errorf("no se pudo decodificar P12/PFX; verifique la clave o convierta el certificado a PEM compatible")
		}
		for _, block := range blocks {
			if block == nil {
				continue
			}
			blockType := strings.ToUpper(strings.TrimSpace(block.Type))
			if material.PrivateKeyPEM == "" && strings.Contains(blockType, "PRIVATE KEY") {
				rsaKey, keyErr := parseDIANRSAPrivateKeyBlock(block, "")
				if keyErr == nil {
					material.PrivateKeyPEM = encodeDIANRSAPrivateKeyPEM(rsaKey)
				}
			}
			if material.CertificatePEM == "" && strings.Contains(blockType, "CERTIFICATE") {
				parsedCert, certErr := x509.ParseCertificate(block.Bytes)
				if certErr == nil {
					material.CertificatePEM = encodeDIANCertificatePEM(parsedCert)
					setDIANSignatureCertificateMetadata(&material, parsedCert)
				}
			}
		}
		if material.PrivateKeyPEM == "" {
			return material, fmt.Errorf("el P12/PFX no contiene llave privada RSA")
		}
		if material.CertificatePEM == "" {
			return material, fmt.Errorf("el P12/PFX no contiene certificado X.509")
		}
		material.Format = strings.TrimPrefix(ext, ".")
		return material, nil
	}

	remaining := contentBytes
	for len(remaining) > 0 {
		block, rest := pem.Decode(remaining)
		if block == nil {
			break
		}
		blockType := strings.ToUpper(strings.TrimSpace(block.Type))
		switch {
		case strings.Contains(blockType, "PRIVATE KEY"):
			rsaKey, err := parseDIANRSAPrivateKeyBlock(block, password)
			if err != nil {
				return material, err
			}
			material.PrivateKeyPEM = encodeDIANRSAPrivateKeyPEM(rsaKey)
		case strings.Contains(blockType, "CERTIFICATE"):
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return material, fmt.Errorf("certificado X.509 invalido")
			}
			if material.CertificatePEM == "" {
				material.CertificatePEM = encodeDIANCertificatePEM(cert)
				setDIANSignatureCertificateMetadata(&material, cert)
			}
		}
		remaining = rest
	}

	if material.PrivateKeyPEM == "" && material.CertificatePEM == "" {
		if cert, err := x509.ParseCertificate(contentBytes); err == nil {
			material.CertificatePEM = encodeDIANCertificatePEM(cert)
			setDIANSignatureCertificateMetadata(&material, cert)
		}
	}
	if material.PrivateKeyPEM == "" {
		if material.CertificatePEM != "" {
			return material, fmt.Errorf("el archivo contiene certificado, pero no contiene llave privada RSA para firmar")
		}
		return material, fmt.Errorf("firma invalida: no se encontro llave privada RSA ni certificado X.509")
	}
	if material.Format == "" {
		if ext != "" {
			material.Format = strings.TrimPrefix(ext, ".")
		} else {
			material.Format = "pem"
		}
	}
	return material, nil
}

func normalizeDIANAcuseEstado(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	switch {
	case raw == "":
		return ""
	case strings.Contains(raw, "acept") || strings.Contains(raw, "validad") || strings.Contains(raw, "approved"):
		return "aceptado"
	case strings.Contains(raw, "rechaz") || strings.Contains(raw, "error") || strings.Contains(raw, "fail"):
		return "rechazado"
	case strings.Contains(raw, "pend") || strings.Contains(raw, "proces") || strings.Contains(raw, "queue"):
		return "pendiente"
	case strings.Contains(raw, "conting"):
		return "contingencia"
	default:
		return raw
	}
}

func extractDIANResponseMap(raw string) map[string]interface{} {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]interface{}{}
	}
	out := map[string]interface{}{}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return map[string]interface{}{}
	}
	return out
}

func resolveDIANAcuseFromResponse(statusCode int, response map[string]interface{}) (string, string) {
	message := dianFirstNonBlank(
		genericStringValue(response["error_message"]),
		genericStringValue(response["mensaje"]),
		genericStringValue(response["message"]),
		genericStringValue(response["detalle"]),
		genericStringValue(response["status_description"]),
		genericStringValue(response["status_message"]),
		genericStringValue(response["description"]),
		genericStringValue(response["error"]),
	)
	statusCodeDIAN := strings.TrimSpace(genericStringValue(response["status_code"]))
	isValidRaw := strings.ToLower(strings.TrimSpace(genericStringValue(response["is_valid"])))
	if isValidRaw == "true" && (statusCodeDIAN == "" || statusCodeDIAN == "00") {
		return "aceptado", dianFirstNonBlank(message, "documento aceptado por DIAN")
	}
	if isValidRaw == "false" {
		normalizedMessage := normalizeDIANAcuseEstado(message)
		if normalizedMessage == "pendiente" {
			return "pendiente", dianFirstNonBlank(message, "Batch en proceso de validacion.")
		}
		if statusCodeDIAN != "" && statusCodeDIAN != "00" {
			return "rechazado", dianFirstNonBlank(message, "documento rechazado por DIAN")
		}
		if genericStringValue(response["error_message"]) != "" {
			return "rechazado", message
		}
		if message != "" {
			return "rechazado", message
		}
	}
	keys := []string{"acuse", "estado", "status", "estado_dian", "resultado", "status_description", "status_message", "error_message"}
	for _, key := range keys {
		if v := normalizeDIANAcuseEstado(genericStringValue(response[key])); v != "" {
			return v, message
		}
	}

	accepted := false
	if parseTruthy(genericStringValue(response["accepted"])) || parseTruthy(genericStringValue(response["ok"])) || parseTruthy(genericStringValue(response["success"])) {
		accepted = true
	}

	if statusCode >= 200 && statusCode < 300 {
		if accepted {
			return "aceptado", "acuse positivo del proveedor DIAN"
		}
		return "enviado", dianFirstNonBlank(message, "documento enviado a DIAN")
	}
	if statusCode >= 500 {
		return "contingencia", "error de transporte DIAN"
	}
	if statusCode >= 400 {
		return "rechazado", "DIAN rechazo la solicitud"
	}
	return "pendiente", dianFirstNonBlank(message, "sin acuse concluyente")
}

func buildDIANSHA384Hex(parts ...string) string {
	seed := strings.Join([]string{
		strings.TrimSpace(strings.Join(parts, "")),
	}, "")
	sum := sha512.Sum384([]byte(seed))
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}

const dianAgencyName = "CO, DIAN (Dirección de Impuestos y Aduanas Nacionales)"

func buildDIANCUFE(nit, documentoCodigo, fechaEmision, total, softwareID, softwarePIN string) string {
	return buildDIANSHA384Hex(nit, documentoCodigo, fechaEmision, total, softwareID, softwarePIN)
}

func dianFormatDecimal(raw string, fallback float64) string {
	value := ventasAnyToFloat64(raw)
	if value == 0 && strings.TrimSpace(raw) == "" {
		value = fallback
	}
	if value < 0 {
		value = 0
	}
	return fmt.Sprintf("%.2f", value)
}

func dianFormatQuantity(raw string, fallback float64) string {
	value := ventasAnyToFloat64(raw)
	if value == 0 && strings.TrimSpace(raw) == "" {
		value = fallback
	}
	if value <= 0 {
		value = fallback
	}
	return fmt.Sprintf("%.6f", value)
}

func dianIssueDateTime(raw string) (string, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		now := time.Now()
		return now.Format("2006-01-02"), now.Format("15:04:05-07:00")
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05-07:00", "2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02"} {
		if parsed, err := time.Parse(layout, raw); err == nil {
			if layout == "2006-01-02" {
				return parsed.Format("2006-01-02"), time.Now().Format("15:04:05-07:00")
			}
			return parsed.Format("2006-01-02"), parsed.Format("15:04:05-07:00")
		}
	}
	if len(raw) >= 10 {
		return raw[:10], time.Now().Format("15:04:05-07:00")
	}
	now := time.Now()
	return now.Format("2006-01-02"), now.Format("15:04:05-07:00")
}

func dianDocumentKind(raw string) (rootName, lineName, customizationID, uuidSchemeName, typeCode, totalTag, quantityTag, correctionCode string) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "nota_credito", "credit_note", "creditnote", "credito", "credit":
		return "CreditNote", "CreditNoteLine", "11", "CUDE-SHA384", "91", "LegalMonetaryTotal", "CreditedQuantity", "1"
	case "nota_debito", "debit_note", "debitnote", "debito", "debit":
		return "DebitNote", "DebitNoteLine", "11", "CUDE-SHA384", "", "RequestedMonetaryTotal", "DebitedQuantity", "1"
	default:
		return "Invoice", "InvoiceLine", "01", "CUFE-SHA384", "01", "LegalMonetaryTotal", "InvoicedQuantity", ""
	}
}

func dianDocumentProfileID(rootName string) string {
	switch rootName {
	case "CreditNote":
		return "DIAN 2.1: Nota Crédito de Factura Electrónica de Venta"
	case "DebitNote":
		return "DIAN 2.1: Nota Débito de Factura Electrónica de Venta"
	default:
		return "DIAN 2.1: Factura Electrónica de Venta"
	}
}

func dianCompanyIDSchemeID(nit, dv string) string {
	if strings.TrimSpace(dv) != "" {
		return dianOnlyDigits(dv)
	}
	if expected, ok := calculateColombianNITDV(nit); ok {
		return strconv.Itoa(expected)
	}
	return "0"
}

func dianSupplierPartyXML(nit, dv, registrationName, prefijo, taxLevel string) string {
	nit = escapeXML(dianOnlyDigits(nit))
	dv = escapeXML(dianCompanyIDSchemeID(nit, dv))
	registrationName = escapeXML(dianFirstNonBlank(registrationName, "EMPRESA SIN RAZON SOCIAL"))
	prefijo = escapeXML(dianFirstNonBlank(prefijo, "SETP"))
	taxLevel = escapeXML(dianFirstNonBlank(taxLevel, "R-99-PN"))
	return fmt.Sprintf(
		`<cac:AccountingSupplierParty>`+
			`<cbc:AdditionalAccountID schemeAgencyID="195">1</cbc:AdditionalAccountID>`+
			`<cac:Party>`+
			`<cac:PartyName><cbc:Name>%s</cbc:Name></cac:PartyName>`+
			`<cac:PhysicalLocation><cac:Address><cbc:ID>11001</cbc:ID><cbc:CityName>Bogota, D.C.</cbc:CityName><cbc:CountrySubentity>Bogota</cbc:CountrySubentity><cbc:CountrySubentityCode>11</cbc:CountrySubentityCode><cac:AddressLine><cbc:Line>Direccion registrada en la empresa</cbc:Line></cac:AddressLine><cac:Country><cbc:IdentificationCode>CO</cbc:IdentificationCode><cbc:Name languageID="es">Colombia</cbc:Name></cac:Country></cac:Address></cac:PhysicalLocation>`+
			`<cac:PartyTaxScheme><cbc:RegistrationName>%s</cbc:RegistrationName><cbc:CompanyID schemeAgencyID="195" schemeAgencyName="%s" schemeID="%s" schemeName="31">%s</cbc:CompanyID><cbc:TaxLevelCode listName="05">%s</cbc:TaxLevelCode><cac:RegistrationAddress><cbc:ID>11001</cbc:ID><cbc:CityName>Bogota, D.C.</cbc:CityName><cbc:CountrySubentity>Bogota</cbc:CountrySubentity><cbc:CountrySubentityCode>11</cbc:CountrySubentityCode><cac:AddressLine><cbc:Line>Direccion registrada en la empresa</cbc:Line></cac:AddressLine><cac:Country><cbc:IdentificationCode>CO</cbc:IdentificationCode><cbc:Name languageID="es">Colombia</cbc:Name></cac:Country></cac:RegistrationAddress><cac:TaxScheme><cbc:ID>01</cbc:ID><cbc:Name>IVA</cbc:Name></cac:TaxScheme></cac:PartyTaxScheme>`+
			`<cac:PartyLegalEntity><cbc:RegistrationName>%s</cbc:RegistrationName><cbc:CompanyID schemeAgencyID="195" schemeAgencyName="%s" schemeID="%s" schemeName="31">%s</cbc:CompanyID><cac:CorporateRegistrationScheme><cbc:ID>%s</cbc:ID></cac:CorporateRegistrationScheme></cac:PartyLegalEntity>`+
			`<cac:Contact><cbc:Name>%s</cbc:Name><cbc:ElectronicMail>facturacion@powerfulcontrolsystem.com</cbc:ElectronicMail></cac:Contact>`+
			`</cac:Party>`+
			`</cac:AccountingSupplierParty>`,
		registrationName,
		registrationName, escapeXML(dianAgencyName), dv, nit, taxLevel,
		registrationName, escapeXML(dianAgencyName), dv, nit, prefijo,
		registrationName,
	)
}

func dianIsConsumidorFinalNIT(customerNIT string) bool {
	digits := dianOnlyDigits(customerNIT)
	return len(digits) >= 6 && strings.Trim(digits, "2") == ""
}

func dianCustomerPartyXML(customerName, customerNIT string) string {
	customerName = escapeXML(dianFirstNonBlank(customerName, "CONSUMIDOR FINAL"))
	customerNITDigits := dianOnlyDigits(dianFirstNonBlank(customerNIT, "2222222222"))
	additionalAccountID := "1"
	schemeID := dianCompanyIDSchemeID(customerNITDigits, "")
	schemeName := "31"
	if dianIsConsumidorFinalNIT(customerNITDigits) {
		customerName = "consumidor o usuario final"
		return fmt.Sprintf(
			`<cac:AccountingCustomerParty>`+
				`<cbc:AdditionalAccountID>2</cbc:AdditionalAccountID>`+
				`<cac:Party>`+
				`<cac:PartyName><cbc:Name>%s</cbc:Name></cac:PartyName>`+
				`<cac:PartyTaxScheme><cbc:RegistrationName>%s</cbc:RegistrationName><cbc:CompanyID schemeAgencyID="195" schemeAgencyName="%s" schemeName="13">%s</cbc:CompanyID><cbc:TaxLevelCode listName="49">R-99-PN</cbc:TaxLevelCode><cac:TaxScheme><cbc:ID>ZZ</cbc:ID><cbc:Name>No aplica</cbc:Name></cac:TaxScheme></cac:PartyTaxScheme>`+
				`</cac:Party>`+
				`</cac:AccountingCustomerParty>`,
			customerName,
			customerName,
			escapeXML(dianAgencyName),
			escapeXML(customerNITDigits),
		)
	}
	if customerNITDigits == "2222222222" || len(customerNITDigits) >= 10 && strings.Trim(customerNITDigits, "2") == "" {
		additionalAccountID = "2"
		schemeID = ""
		schemeName = "13"
	}
	taxLevelListName := "04"
	if additionalAccountID == "2" {
		taxLevelListName = "49"
	}
	companyIDAttrs := fmt.Sprintf(`schemeAgencyID="195" schemeAgencyName="%s" schemeName="%s"`, escapeXML(dianAgencyName), escapeXML(schemeName))
	if schemeID != "" {
		companyIDAttrs += fmt.Sprintf(` schemeID="%s"`, escapeXML(schemeID))
	}
	return fmt.Sprintf(
		`<cac:AccountingCustomerParty>`+
			`<cbc:AdditionalAccountID>%s</cbc:AdditionalAccountID>`+
			`<cac:Party>`+
			`<cac:PartyName><cbc:Name>%s</cbc:Name></cac:PartyName>`+
			`<cac:PhysicalLocation><cac:Address><cbc:ID>11001</cbc:ID><cbc:CityName>Bogota, D.C.</cbc:CityName><cbc:CountrySubentity>Bogota</cbc:CountrySubentity><cbc:CountrySubentityCode>11</cbc:CountrySubentityCode><cac:AddressLine><cbc:Line>Direccion del adquiriente</cbc:Line></cac:AddressLine><cac:Country><cbc:IdentificationCode>CO</cbc:IdentificationCode><cbc:Name languageID="es">Colombia</cbc:Name></cac:Country></cac:Address></cac:PhysicalLocation>`+
			`<cac:PartyTaxScheme><cbc:RegistrationName>%s</cbc:RegistrationName><cbc:CompanyID %s>%s</cbc:CompanyID><cbc:TaxLevelCode listName="%s">R-99-PN</cbc:TaxLevelCode><cac:RegistrationAddress><cbc:ID>11001</cbc:ID><cbc:CityName>Bogota, D.C.</cbc:CityName><cbc:CountrySubentity>Bogota</cbc:CountrySubentity><cbc:CountrySubentityCode>11</cbc:CountrySubentityCode><cac:AddressLine><cbc:Line>Direccion del adquiriente</cbc:Line></cac:AddressLine><cac:Country><cbc:IdentificationCode>CO</cbc:IdentificationCode><cbc:Name languageID="es">Colombia</cbc:Name></cac:Country></cac:RegistrationAddress><cac:TaxScheme><cbc:ID>01</cbc:ID><cbc:Name>IVA</cbc:Name></cac:TaxScheme></cac:PartyTaxScheme>`+
			`<cac:PartyLegalEntity><cbc:RegistrationName>%s</cbc:RegistrationName><cbc:CompanyID %s>%s</cbc:CompanyID></cac:PartyLegalEntity>`+
			`<cac:Contact><cbc:Name>%s</cbc:Name><cbc:ElectronicMail>cliente@example.com</cbc:ElectronicMail></cac:Contact>`+
			`</cac:Party>`+
			`</cac:AccountingCustomerParty>`,
		additionalAccountID,
		customerName,
		customerName, companyIDAttrs, escapeXML(customerNITDigits), escapeXML(taxLevelListName),
		customerName, companyIDAttrs, escapeXML(customerNITDigits),
		customerName,
	)
}

func dianTaxTotalXML(currency, taxable, tax, percent string) string {
	return fmt.Sprintf(
		`<cac:TaxTotal><cbc:TaxAmount currencyID="%s">%s</cbc:TaxAmount><cac:TaxSubtotal><cbc:TaxableAmount currencyID="%s">%s</cbc:TaxableAmount><cbc:TaxAmount currencyID="%s">%s</cbc:TaxAmount><cac:TaxCategory><cbc:Percent>%s</cbc:Percent><cac:TaxScheme><cbc:ID>01</cbc:ID><cbc:Name>IVA</cbc:Name></cac:TaxScheme></cac:TaxCategory></cac:TaxSubtotal></cac:TaxTotal>`,
		escapeXML(currency), escapeXML(tax), escapeXML(currency), escapeXML(taxable), escapeXML(currency), escapeXML(tax), escapeXML(percent),
	)
}

func dianMonetaryTotalXML(tagName, currency, taxable, tax, total string) string {
	return fmt.Sprintf(
		`<cac:%s><cbc:LineExtensionAmount currencyID="%s">%s</cbc:LineExtensionAmount><cbc:TaxExclusiveAmount currencyID="%s">%s</cbc:TaxExclusiveAmount><cbc:TaxInclusiveAmount currencyID="%s">%s</cbc:TaxInclusiveAmount><cbc:AllowanceTotalAmount currencyID="%s">0.00</cbc:AllowanceTotalAmount><cbc:ChargeTotalAmount currencyID="%s">0.00</cbc:ChargeTotalAmount><cbc:PrepaidAmount currencyID="%s">0.00</cbc:PrepaidAmount><cbc:PayableAmount currencyID="%s">%s</cbc:PayableAmount></cac:%s>`,
		tagName, escapeXML(currency), escapeXML(taxable), escapeXML(currency), escapeXML(taxable), escapeXML(currency), escapeXML(total), escapeXML(currency), escapeXML(currency), escapeXML(currency), escapeXML(currency), escapeXML(total), tagName,
	)
}

func dianPaymentMeansXML(issueDate string) string {
	return fmt.Sprintf(`<cac:PaymentMeans><cbc:ID>1</cbc:ID><cbc:PaymentMeansCode>10</cbc:PaymentMeansCode><cbc:PaymentDueDate>%s</cbc:PaymentDueDate><cbc:PaymentID>1</cbc:PaymentID></cac:PaymentMeans>`, escapeXML(issueDate))
}

func dianLineXML(lineName, quantityTag, currency, taxable, tax, percent, quantity string) string {
	return fmt.Sprintf(
		`<cac:%s><cbc:ID>1</cbc:ID><cbc:%s unitCode="EA">%s</cbc:%s><cbc:LineExtensionAmount currencyID="%s">%s</cbc:LineExtensionAmount>%s<cac:Item><cbc:Description>Servicio de habilitacion DIAN</cbc:Description><cac:SellersItemIdentification><cbc:ID>PCS-DIAN-001</cbc:ID></cac:SellersItemIdentification><cac:StandardItemIdentification><cbc:ID schemeID="999" schemeName="EAN13">7700000000019</cbc:ID></cac:StandardItemIdentification></cac:Item><cac:Price><cbc:PriceAmount currencyID="%s">%s</cbc:PriceAmount><cbc:BaseQuantity unitCode="EA">%s</cbc:BaseQuantity></cac:Price></cac:%s>`,
		lineName, quantityTag, escapeXML(quantity), quantityTag, escapeXML(currency), escapeXML(taxable), dianTaxTotalXML(currency, taxable, tax, percent), escapeXML(currency), escapeXML(taxable), escapeXML(quantity), lineName,
	)
}

func dianProfileExecutionSchemeID(cfg map[string]interface{}) string {
	return dianExpectedProfileExecutionID(cfg)
}

func dianValidationIssue(code, severity, field, message, source string) map[string]interface{} {
	return map[string]interface{}{
		"code":     code,
		"severity": severity,
		"field":    field,
		"message":  message,
		"source":   source,
	}
}

func dianAppendValidationIssue(issues *[]map[string]interface{}, warnings *[]map[string]interface{}, code, severity, field, message, source string) {
	item := dianValidationIssue(code, severity, field, message, source)
	if severity == "error" {
		*issues = append(*issues, item)
		return
	}
	*warnings = append(*warnings, item)
}

func dianOnlyDigits(raw string) string {
	var b strings.Builder
	for _, ch := range raw {
		if ch >= '0' && ch <= '9' {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func calculateColombianNITDV(nit string) (int, bool) {
	digits := dianOnlyDigits(nit)
	if digits == "" {
		return 0, false
	}
	factors := []int{3, 7, 13, 17, 19, 23, 29, 37, 41, 43, 47, 53, 59, 67, 71}
	sum := 0
	factorIndex := 0
	for i := len(digits) - 1; i >= 0; i-- {
		if factorIndex >= len(factors) {
			return 0, false
		}
		sum += int(digits[i]-'0') * factors[factorIndex]
		factorIndex++
	}
	remainder := sum % 11
	if remainder > 1 {
		return 11 - remainder, true
	}
	return remainder, true
}

func parseDIANDate(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func parseDIANXMLTextValues(xmlPayload string) (map[string][]string, string, error) {
	values := map[string][]string{}
	decoder := xml.NewDecoder(strings.NewReader(xmlPayload))
	root := ""
	var current string
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return values, root, err
		}
		switch node := token.(type) {
		case xml.StartElement:
			if root == "" {
				root = node.Name.Local
			}
			current = node.Name.Local
		case xml.CharData:
			if current != "" {
				text := strings.TrimSpace(string(node))
				if text != "" {
					values[current] = append(values[current], text)
				}
			}
		case xml.EndElement:
			if current == node.Name.Local {
				current = ""
			}
		}
	}
	return values, root, nil
}

func dianXMLFirst(values map[string][]string, names ...string) string {
	for _, name := range names {
		for key, list := range values {
			if strings.EqualFold(key, name) && len(list) > 0 {
				return strings.TrimSpace(list[0])
			}
		}
	}
	return ""
}

func dianExpectedProfileExecutionID(cfg map[string]interface{}) string {
	if chooseDIANAmbiente(cfg) == "produccion" {
		return "1"
	}
	return "2"
}

func validateDIANDocumentPreflight(cfg map[string]interface{}, empresaID int64, payload map[string]interface{}, xmlPayload, stage string) map[string]interface{} {
	if payload == nil {
		payload = map[string]interface{}{}
	}
	stage = strings.ToLower(strings.TrimSpace(stage))
	if stage == "" {
		stage = "validacion_manual"
	}

	issues := make([]map[string]interface{}, 0)
	warnings := make([]map[string]interface{}, 0)
	checks := map[string]interface{}{}
	source := "sistema_preflight_dian"

	if empresaID <= 0 {
		dianAppendValidationIssue(&issues, &warnings, "DIAN-PRE-EMPRESA", "error", "empresa_id", "empresa_id es obligatorio", source)
	}
	if len(cfg) == 0 {
		dianAppendValidationIssue(&issues, &warnings, "DIAN-PRE-CONFIG", "error", "configuracion", "no existe configuracion DIAN para la empresa", source)
	} else {
		missing := missingDIANFields(cfg)
		checks["configuracion_minima"] = map[string]interface{}{"ok": len(missing) == 0, "faltantes": missing}
		for _, field := range missing {
			dianAppendValidationIssue(&issues, &warnings, "DIAN-CFG-FALTANTE", "error", field, "campo DIAN obligatorio no configurado: "+field, "configuracion_dian")
		}
	}

	nit := dianFirstNonBlank(genericStringValue(payload["nit"]), genericStringValue(cfg["nit"]))
	dvRaw := dianFirstNonBlank(genericStringValue(payload["digito_verificacion"]), genericStringValue(payload["dv"]), genericStringValue(cfg["digito_verificacion"]))
	if nit == "" {
		dianAppendValidationIssue(&issues, &warnings, "DIAN-NIT-001", "error", "nit", "NIT del emisor es obligatorio", "anexo_tecnico_dian")
	} else if dianOnlyDigits(nit) != nit && strings.ContainsAny(nit, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		dianAppendValidationIssue(&issues, &warnings, "DIAN-NIT-002", "error", "nit", "NIT del emisor debe contener solo digitos o separadores numericos", "anexo_tecnico_dian")
	} else if expectedDV, ok := calculateColombianNITDV(nit); ok {
		checks["nit_dv"] = map[string]interface{}{"ok": dvRaw == "" || dvRaw == strconv.Itoa(expectedDV), "nit": dianOnlyDigits(nit), "dv_esperado": expectedDV, "dv_configurado": dvRaw}
		if dvRaw != "" && dvRaw != strconv.Itoa(expectedDV) {
			dianAppendValidationIssue(&issues, &warnings, "DIAN-NIT-DV", "error", "digito_verificacion", fmt.Sprintf("digito de verificacion no coincide; esperado %d", expectedDV), "regla_nit_colombia")
		}
	}

	ambiente := chooseDIANAmbiente(cfg)
	if ambiente != "habilitacion" && ambiente != "produccion" {
		dianAppendValidationIssue(&issues, &warnings, "DIAN-AMB-001", "error", "tipo_ambiente", "ambiente DIAN debe ser habilitacion o produccion", source)
	}
	if ambiente == "habilitacion" && strings.TrimSpace(genericStringValue(cfg["test_set_id"])) == "" && !parseTruthy(dianPayloadString(payload, "simular")) {
		dianAppendValidationIssue(&issues, &warnings, "DIAN-SET-001", "error", "test_set_id", "en habilitacion real debe existir TestSetId asignado por DIAN", "habilitacion_dian")
	}

	prefijo := dianFirstNonBlank(genericStringValue(payload["prefijo"]), genericStringValue(cfg["prefijo"]))
	documentoCodigo := dianFirstNonBlank(genericStringValue(payload["documento_codigo"]), genericStringValue(payload["numero"]), genericStringValue(payload["id_documento"]))
	if documentoCodigo == "" {
		dianAppendValidationIssue(&issues, &warnings, "DIAN-DOC-001", "error", "documento_codigo", "codigo/numero del documento es obligatorio", "anexo_tecnico_dian")
	} else if prefijo != "" && !strings.HasPrefix(strings.ToUpper(documentoCodigo), strings.ToUpper(prefijo)) {
		dianAppendValidationIssue(&issues, &warnings, "DIAN-NUM-001", "warning", "documento_codigo", "el documento no inicia con el prefijo configurado", "numeracion_dian")
	}

	rangoDesde := anyToInt64(cfg["rango_desde"])
	rangoHasta := anyToInt64(cfg["rango_hasta"])
	consecutivoActual := anyToInt64(cfg["consecutivo_actual"])
	if rangoDesde > 0 && rangoHasta > 0 && rangoDesde > rangoHasta {
		dianAppendValidationIssue(&issues, &warnings, "DIAN-RNG-001", "error", "rango_desde", "rango de numeracion DIAN invalido: desde mayor que hasta", "numeracion_dian")
	}
	if consecutivoActual > 0 && rangoDesde > 0 && consecutivoActual < rangoDesde {
		dianAppendValidationIssue(&issues, &warnings, "DIAN-RNG-002", "error", "consecutivo_actual", "consecutivo actual esta antes del rango autorizado", "numeracion_dian")
	}
	if consecutivoActual > 0 && rangoHasta > 0 && consecutivoActual > rangoHasta {
		dianAppendValidationIssue(&issues, &warnings, "DIAN-RNG-003", "error", "consecutivo_actual", "consecutivo actual supera el rango autorizado", "numeracion_dian")
	}
	checks["numeracion"] = map[string]interface{}{"prefijo": prefijo, "rango_desde": rangoDesde, "rango_hasta": rangoHasta, "consecutivo_actual": consecutivoActual}

	resDesde := genericStringValue(cfg["resolucion_fecha_desde"])
	resHasta := genericStringValue(cfg["resolucion_fecha_hasta"])
	now := time.Now()
	if resDesde != "" {
		if parsed, ok := parseDIANDate(resDesde); !ok {
			dianAppendValidationIssue(&issues, &warnings, "DIAN-RES-001", "error", "resolucion_fecha_desde", "fecha inicial de resolucion invalida", "resolucion_dian")
		} else if now.Before(parsed) {
			dianAppendValidationIssue(&issues, &warnings, "DIAN-RES-002", "error", "resolucion_fecha_desde", "la resolucion aun no esta vigente", "resolucion_dian")
		}
	}
	if resHasta != "" {
		if parsed, ok := parseDIANDate(resHasta); !ok {
			dianAppendValidationIssue(&issues, &warnings, "DIAN-RES-003", "error", "resolucion_fecha_hasta", "fecha final de resolucion invalida", "resolucion_dian")
		} else if now.After(parsed) {
			dianAppendValidationIssue(&issues, &warnings, "DIAN-RES-004", "error", "resolucion_fecha_hasta", "la resolucion de facturacion esta vencida", "resolucion_dian")
		}
	}

	fechaEmision := dianFirstNonBlank(genericStringValue(payload["fecha_emision"]), genericStringValue(payload["fecha"]))
	if fechaEmision != "" {
		if parsed, ok := parseDIANDate(fechaEmision); !ok {
			dianAppendValidationIssue(&issues, &warnings, "DIAN-FEC-001", "error", "fecha_emision", "fecha de emision invalida", "anexo_tecnico_dian")
		} else {
			if parsed.After(now.Add(10 * time.Minute)) {
				dianAppendValidationIssue(&issues, &warnings, "DIAN-FEC-002", "error", "fecha_emision", "fecha de emision no puede estar en el futuro", "anexo_tecnico_dian")
			}
			if parsed.Before(now.AddDate(0, 0, -10)) {
				dianAppendValidationIssue(&issues, &warnings, "DIAN-FEC-003", "warning", "fecha_emision", "fecha de emision es antigua; revise reglas de expedicion/contingencia", "operacion_dian")
			}
		}
	}

	total := ventasAnyToFloat64(dianFirstNonBlank(genericStringValue(payload["total"]), genericStringValue(payload["payable_amount"])))
	impuesto := ventasAnyToFloat64(dianFirstNonBlank(genericStringValue(payload["impuesto_total"]), genericStringValue(payload["tax_amount"])))
	if total <= 0 {
		dianAppendValidationIssue(&issues, &warnings, "DIAN-TOT-001", "error", "total", "total del documento debe ser mayor que cero", "anexo_tecnico_dian")
	}
	if impuesto < 0 {
		dianAppendValidationIssue(&issues, &warnings, "DIAN-TAX-001", "error", "impuesto_total", "impuesto total no puede ser negativo", "anexo_tecnico_dian")
	}
	if total > 0 && impuesto > total {
		dianAppendValidationIssue(&issues, &warnings, "DIAN-TAX-002", "error", "impuesto_total", "impuesto total no puede superar el total del documento", "anexo_tecnico_dian")
	}
	moneda := strings.ToUpper(dianFirstNonBlank(genericStringValue(payload["moneda"]), "COP"))
	if len(moneda) != 3 {
		dianAppendValidationIssue(&issues, &warnings, "DIAN-MON-001", "error", "moneda", "moneda debe usar codigo ISO de 3 letras", "anexo_tecnico_dian")
	}

	clienteNombre := dianFirstNonBlank(genericStringValue(payload["cliente_nombre"]), genericStringValue(payload["adquiriente_nombre"]))
	clienteNIT := dianFirstNonBlank(genericStringValue(payload["cliente_nit"]), genericStringValue(payload["adquiriente_nit"]))
	if strings.TrimSpace(clienteNombre) == "" {
		severity := "error"
		if strings.TrimSpace(xmlPayload) != "" {
			severity = "warning"
		}
		dianAppendValidationIssue(&issues, &warnings, "DIAN-CLI-001", severity, "cliente_nombre", "nombre/razon social del adquiriente es obligatorio; si viene en XML, verifique AccountingCustomerParty", "anexo_tecnico_dian")
	}
	if strings.TrimSpace(clienteNIT) == "" {
		severity := "error"
		if strings.TrimSpace(xmlPayload) != "" {
			severity = "warning"
		}
		dianAppendValidationIssue(&issues, &warnings, "DIAN-CLI-002", severity, "cliente_nit", "identificacion del adquiriente es obligatoria; si viene en XML, verifique CompanyID del adquiriente", "anexo_tecnico_dian")
	}

	if stage == "envio_real" || stage == "set_habilitacion" || stage == "pre_envio" {
		if keyRef := dianFirstNonBlank(genericStringValue(payload["certificado_clave_ref"]), genericStringValue(cfg["certificado_clave_ref"])); keyRef == "" {
			dianAppendValidationIssue(&issues, &warnings, "DIAN-FIR-001", "error", "certificado_clave_ref", "llave privada de firma es obligatoria antes del envio", "firma_digital")
		} else if _, err := parseDIANRSAPrivateKey(keyRef); err != nil {
			dianAppendValidationIssue(&issues, &warnings, "DIAN-FIR-002", "error", "certificado_clave_ref", "llave privada de firma invalida: "+dianTruncate(err.Error(), 120), "firma_digital")
		}
		if certRef := dianFirstNonBlank(genericStringValue(payload["certificado_ref"]), genericStringValue(payload["certificado_pem"]), genericStringValue(cfg["certificado_url"])); certRef == "" {
			dianAppendValidationIssue(&issues, &warnings, "DIAN-CER-001", "error", "certificado_url", "certificado X.509 es obligatorio antes del envio", "firma_digital")
		} else if cert, err := parseDIANCertificate(certRef); err != nil {
			dianAppendValidationIssue(&issues, &warnings, "DIAN-CER-002", "error", "certificado_url", "certificado X.509 invalido: "+dianTruncate(err.Error(), 120), "firma_digital")
		} else {
			expiry := dianCertificateExpiryStatus(time.Now(), cert.NotBefore, cert.NotAfter, dianCertificateAlertThresholdDays(cfg))
			if parseTruthy(genericStringValue(expiry["vencido"])) {
				dianAppendValidationIssue(&issues, &warnings, "DIAN-CER-003", "error", "certificado_url", "certificado X.509 vencido", "firma_digital")
			} else if parseTruthy(genericStringValue(expiry["proximo_a_vencer"])) {
				dianAppendValidationIssue(&issues, &warnings, "DIAN-CER-004", "warning", "certificado_url", genericStringValue(expiry["mensaje"]), "firma_digital")
			}
		}
	}

	xmlChecks := map[string]interface{}{"presente": strings.TrimSpace(xmlPayload) != ""}
	if strings.TrimSpace(xmlPayload) == "" {
		if stage == "envio_real" || stage == "set_habilitacion" || stage == "pre_envio" {
			dianAppendValidationIssue(&issues, &warnings, "DIAN-XML-001", "error", "xml", "XML UBL firmado es obligatorio antes del envio", "ubl_2_1")
		}
	} else {
		values, root, err := parseDIANXMLTextValues(xmlPayload)
		xmlChecks["root"] = root
		if err != nil {
			dianAppendValidationIssue(&issues, &warnings, "DIAN-XML-002", "error", "xml", "XML no esta bien formado: "+dianTruncate(err.Error(), 160), "ubl_2_1")
		} else {
			if root != "Invoice" && root != "CreditNote" && root != "DebitNote" {
				dianAppendValidationIssue(&issues, &warnings, "DIAN-UBL-ROOT", "error", "xml", "raiz UBL debe ser Invoice, CreditNote o DebitNote", "ubl_2_1")
			}
			ublVersion := dianXMLFirst(values, "UBLVersionID")
			profileID := dianXMLFirst(values, "ProfileID")
			profileExecutionID := dianXMLFirst(values, "ProfileExecutionID")
			xmlID := dianXMLFirst(values, "ID")
			xmlUUID := dianXMLFirst(values, "UUID")
			xmlCustomizationID := dianXMLFirst(values, "CustomizationID")
			xmlPayable := ventasAnyToFloat64(dianXMLFirst(values, "PayableAmount"))
			xmlTax := ventasAnyToFloat64(dianXMLFirst(values, "TaxAmount"))
			xmlInvoiceType := dianXMLFirst(values, "InvoiceTypeCode")
			xmlCreditType := dianXMLFirst(values, "CreditNoteTypeCode")
			expectedRootName, expectedLineName, expectedCustomizationID, expectedUUIDScheme, expectedTypeCode, expectedTotalTag, _, _ := dianDocumentKind(dianFirstNonBlank(genericStringValue(payload["documento_tipo"]), root))
			expectedProfileID := dianDocumentProfileID(expectedRootName)
			xmlChecks["ubl_version"] = ublVersion
			xmlChecks["profile_id"] = profileID
			xmlChecks["profile_execution_id"] = profileExecutionID
			xmlChecks["customization_id"] = xmlCustomizationID
			xmlChecks["id"] = xmlID
			xmlChecks["uuid_presente"] = xmlUUID != ""
			xmlChecks["signature_presente"] = strings.Contains(xmlPayload, "<ds:Signature") || strings.Contains(xmlPayload, ":Signature")
			xmlChecks["x509_presente"] = strings.Contains(xmlPayload, "X509Certificate")
			xmlChecks["dian_extensions_presente"] = strings.Contains(xmlPayload, "DianExtensions")
			xmlChecks["software_security_code_presente"] = strings.Contains(xmlPayload, "SoftwareSecurityCode")
			xmlChecks["linea_esperada"] = expectedLineName
			xmlChecks["linea_presente"] = strings.Contains(xmlPayload, "<cac:"+expectedLineName)
			xmlChecks["total_esperado"] = expectedTotalTag
			xmlChecks["total_presente"] = strings.Contains(xmlPayload, "<cac:"+expectedTotalTag)
			if ublVersion != "2.1" && ublVersion != "UBL 2.1" {
				dianAppendValidationIssue(&issues, &warnings, "DIAN-UBL-001", "error", "UBLVersionID", "UBLVersionID debe ser UBL 2.1", "ubl_2_1")
			}
			if profileID != "" && profileID != expectedProfileID && !strings.EqualFold(profileID, "DIAN 2.1") {
				dianAppendValidationIssue(&issues, &warnings, "DIAN-UBL-PROFILE", "error", "ProfileID", "ProfileID debe ser "+expectedProfileID, "ubl_2_1")
			}
			if profileExecutionID != "" && profileExecutionID != dianExpectedProfileExecutionID(cfg) {
				dianAppendValidationIssue(&issues, &warnings, "DIAN-UBL-002", "error", "ProfileExecutionID", "ProfileExecutionID no coincide con el ambiente DIAN configurado", "ubl_2_1")
			}
			if xmlCustomizationID != "" && expectedCustomizationID != "" && xmlCustomizationID != expectedCustomizationID {
				dianAppendValidationIssue(&issues, &warnings, "DIAN-UBL-CUSTOM", "error", "CustomizationID", "CustomizationID no corresponde al tipo de documento DIAN", "ubl_2_1")
			}
			if documentoCodigo != "" && xmlID != "" && xmlID != documentoCodigo {
				dianAppendValidationIssue(&issues, &warnings, "DIAN-UBL-003", "error", "ID", "ID del XML no coincide con documento_codigo", "ubl_2_1")
			}
			if xmlUUID == "" {
				dianAppendValidationIssue(&issues, &warnings, "DIAN-UBL-004", "error", "UUID", "CUFE/CUDE UUID es obligatorio en el XML", "ubl_2_1")
			}
			if xmlUUID != "" && expectedUUIDScheme != "" && !strings.Contains(xmlPayload, `schemeName="`+expectedUUIDScheme+`"`) {
				dianAppendValidationIssue(&issues, &warnings, "DIAN-UBL-UUID-SCHEME", "error", "UUID", "UUID debe usar "+expectedUUIDScheme+" segun el tipo de documento", "ubl_2_1")
			}
			if root == "Invoice" && expectedTypeCode != "" && xmlInvoiceType != expectedTypeCode {
				dianAppendValidationIssue(&issues, &warnings, "DIAN-UBL-TIPO-FE", "error", "InvoiceTypeCode", "factura electronica de venta debe usar InvoiceTypeCode "+expectedTypeCode, "anexo_tecnico_dian")
			}
			if root == "CreditNote" && expectedTypeCode != "" && xmlCreditType != expectedTypeCode {
				dianAppendValidationIssue(&issues, &warnings, "DIAN-UBL-TIPO-NC", "error", "CreditNoteTypeCode", "nota credito debe usar CreditNoteTypeCode "+expectedTypeCode, "anexo_tecnico_dian")
			}
			if (root == "CreditNote" || root == "DebitNote") && (!strings.Contains(xmlPayload, "<cac:DiscrepancyResponse") || !strings.Contains(xmlPayload, "<cac:BillingReference")) {
				dianAppendValidationIssue(&issues, &warnings, "DIAN-UBL-NOTA-REF", "error", "BillingReference", "notas credito/debito deben referenciar la factura afectada con DiscrepancyResponse y BillingReference", "anexo_tecnico_dian")
			}
			if expectedLineName != "" && !strings.Contains(xmlPayload, "<cac:"+expectedLineName) {
				dianAppendValidationIssue(&issues, &warnings, "DIAN-UBL-LINEA", "error", expectedLineName, "el tipo de documento debe usar "+expectedLineName, "xsd_dian_ubl_2_1")
			}
			if expectedTotalTag != "" && !strings.Contains(xmlPayload, "<cac:"+expectedTotalTag) {
				dianAppendValidationIssue(&issues, &warnings, "DIAN-UBL-TOTAL", "error", expectedTotalTag, "el tipo de documento debe usar "+expectedTotalTag, "xsd_dian_ubl_2_1")
			}
			if !strings.Contains(xmlPayload, "DianExtensions") {
				dianAppendValidationIssue(&issues, &warnings, "DIAN-UBL-EXT", "error", "DianExtensions", "XML debe incluir ext:UBLExtensions/sts:DianExtensions con datos DIAN", "anexo_tecnico_dian")
			}
			if !strings.Contains(xmlPayload, "SoftwareSecurityCode") {
				dianAppendValidationIssue(&issues, &warnings, "DIAN-UBL-SOFTWARE-CODE", "error", "SoftwareSecurityCode", "XML debe incluir SoftwareSecurityCode calculado con software, PIN y numero de documento", "anexo_tecnico_dian")
			}
			if strings.Contains(strings.ToUpper(xmlPayload), "PENDIENTE") || strings.Contains(strings.ToUpper(xmlPayload), "DEMO") {
				severity := "warning"
				if stage == "envio_real" || stage == "set_habilitacion" || stage == "pre_envio" {
					severity = "error"
				}
				dianAppendValidationIssue(&issues, &warnings, "DIAN-UBL-DEMO", severity, "xml", "XML contiene marcadores DEMO/PENDIENTE; reemplace esos valores antes de enviarlo a DIAN", "ubl_2_1")
			}
			if total > 0 && xmlPayable > 0 && math.Abs(xmlPayable-total) > 0.01 {
				dianAppendValidationIssue(&issues, &warnings, "DIAN-TOT-XML", "error", "PayableAmount", "PayableAmount del XML no coincide con el total enviado", "ubl_2_1")
			}
			if impuesto >= 0 && xmlTax >= 0 && math.Abs(xmlTax-impuesto) > 0.01 {
				dianAppendValidationIssue(&issues, &warnings, "DIAN-TAX-XML", "warning", "TaxAmount", "TaxAmount del XML no coincide con impuesto_total del payload", "ubl_2_1")
			}
			if (stage == "envio_real" || stage == "set_habilitacion" || stage == "pre_envio") && !parseTruthy(genericStringValue(xmlChecks["signature_presente"])) {
				dianAppendValidationIssue(&issues, &warnings, "DIAN-FIR-XML", "error", "xml_firmado", "XML de envio debe incluir ds:Signature", "firma_digital")
			}
			if (stage == "envio_real" || stage == "set_habilitacion" || stage == "pre_envio") && !parseTruthy(genericStringValue(xmlChecks["x509_presente"])) {
				dianAppendValidationIssue(&issues, &warnings, "DIAN-FIR-X509", "warning", "xml_firmado", "XML firmado no incluye X509Certificate; DIAN/proveedor puede rechazarlo", "firma_digital")
			}
		}
	}
	checks["xml_ubl"] = xmlChecks

	ok := len(issues) == 0
	return map[string]interface{}{
		"ok":                    ok,
		"bloqueado":             !ok,
		"empresa_id":            empresaID,
		"etapa":                 stage,
		"ambiente":              ambiente,
		"issues":                issues,
		"warnings":              warnings,
		"checks":                checks,
		"fuente_normativa":      []string{"DIAN - Anexo Tecnico Factura Electronica de Venta UBL 2.1", "DIAN - validacion previa en tiempo real por plataforma DIAN"},
		"nota":                  "Estas validaciones preventivas reducen rechazos antes del envio; la aceptacion fiscal final la confirma DIAN con su acuse.",
		"total_errores":         len(issues),
		"total_advertencias":    len(warnings),
		"validacion_en_sistema": true,
	}
}

func updateDIANConfigFields(dbEmp *sql.DB, empresaID int64, cfg map[string]interface{}, updates map[string]interface{}) error {
	if dbEmp == nil || len(cfg) == 0 || len(updates) == 0 {
		return nil
	}
	id := anyToInt64(cfg["id"])
	if id <= 0 {
		return nil
	}
	if err := dbpkg.UpdateEmpresaGenericRow(dbEmp, cfgDIAN.Table, empresaID, id, updates, cfgDIAN.AllowedColumns); err != nil {
		return err
	}
	for key, value := range updates {
		cfg[key] = value
	}
	return nil
}

func signDIANXMLReal(cfg map[string]interface{}, empresaID int64, payload map[string]interface{}) (map[string]interface{}, int, error) {
	xmlPayload := dianFirstNonBlank(
		genericStringValue(payload["xml"]),
		genericStringValue(payload["xml_demo"]),
		genericStringValue(payload["xml_firmado"]),
	)
	if xmlPayload == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("xml es obligatorio para firmar")
	}

	keyRef := dianFirstNonBlank(
		genericStringValue(payload["private_key_pem"]),
		genericStringValue(payload["certificado_clave_ref"]),
		genericStringValue(cfg["certificado_clave_ref"]),
	)
	if keyRef == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("certificado_clave_ref es obligatorio para firma real")
	}

	privateKey, err := parseDIANRSAPrivateKey(keyRef)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	digest := sha256.Sum256([]byte(xmlPayload))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("no se pudo firmar XML con RSA-SHA256")
	}

	doc := dianFirstNonBlank(genericStringValue(payload["documento_codigo"]), "FV-"+time.Now().Format("20060102150405"))
	return map[string]interface{}{
		"ok":                true,
		"empresa_id":        empresaID,
		"documento_codigo":  doc,
		"algoritmo":         "RSA-SHA256",
		"digest_sha256_hex": strings.ToUpper(hex.EncodeToString(digest[:])),
		"firma_base64":      base64.StdEncoding.EncodeToString(signature),
		"xml_firmado":       xmlPayload,
		"timestamp":         dianNowLocal(),
		"gestion_secreto":   "certificado_clave_ref via env:/file:/base64: o valor inline controlado",
	}, http.StatusOK, nil
}

func generateDIANUBLBase(cfg map[string]interface{}, empresaID int64, payload map[string]interface{}) (map[string]interface{}, int, error) {
	if empresaID <= 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("empresa_id es obligatorio")
	}

	documentoCodigo := dianFirstNonBlank(genericStringValue(payload["documento_codigo"]), "FV-"+time.Now().Format("20060102150405"))
	documentoTipo := strings.ToLower(dianFirstNonBlank(genericStringValue(payload["documento_tipo"]), "factura"))
	issueDateTime := dianFirstNonBlank(genericStringValue(payload["fecha_emision"]), time.Now().Format(time.RFC3339))
	total := dianFormatDecimal(genericStringValue(payload["total"]), 1190.00)
	impuestoTotal := dianFormatDecimal(genericStringValue(payload["impuesto_total"]), 190.00)
	moneda := strings.ToUpper(dianFirstNonBlank(genericStringValue(payload["moneda"]), "COP"))
	clienteNombre := dianFirstNonBlank(genericStringValue(payload["cliente_nombre"]), "CONSUMIDOR FINAL")
	clienteNIT := dianOnlyDigits(dianFirstNonBlank(genericStringValue(payload["cliente_nit"]), "2222222222"))
	emisorNIT := dianOnlyDigits(dianFirstNonBlank(genericStringValue(cfg["nit"]), "000000000"))
	emisorDV := dianFirstNonBlank(genericStringValue(cfg["digito_verificacion"]), genericStringValue(payload["digito_verificacion"]))
	emisorRazon := dianFirstNonBlank(genericStringValue(cfg["razon_social"]), "EMPRESA SIN RAZON SOCIAL CONFIGURADA")
	prefijo := dianFirstNonBlank(genericStringValue(cfg["prefijo"]), "SETP")
	llaveTecnica := dianFirstNonBlank(genericStringValue(cfg["llave_tecnica"]), genericStringValue(cfg["clave_tecnica"]), genericStringValue(cfg["technical_key"]))
	resolucionNumero := dianFirstNonBlank(genericStringValue(cfg["resolucion_numero"]), "18760000001")
	resolucionDesde := dianFirstNonBlank(genericStringValue(cfg["resolucion_fecha_desde"]), "2019-01-19")
	resolucionHasta := dianFirstNonBlank(genericStringValue(cfg["resolucion_fecha_hasta"]), "2030-01-19")
	rangoDesde := anyToInt64(cfg["rango_desde"])
	if rangoDesde <= 0 {
		rangoDesde = 990000000
	}
	rangoHasta := anyToInt64(cfg["rango_hasta"])
	if rangoHasta <= 0 {
		rangoHasta = 995000000
	}
	profileExecutionID := "2"
	if chooseDIANAmbiente(cfg) == "produccion" {
		profileExecutionID = "1"
	}

	issueDateOnly, issueTime := dianIssueDateTime(issueDateTime)
	rootName, lineName, customizationID, uuidSchemeName, typeCode, monetaryTotalTag, quantityTag, correctionCode := dianDocumentKind(documentoTipo)
	totalFloat := ventasAnyToFloat64(total)
	taxFloat := ventasAnyToFloat64(impuestoTotal)
	taxableFloat := totalFloat - taxFloat
	if taxableFloat < 0 {
		taxableFloat = totalFloat
	}
	taxable := fmt.Sprintf("%.2f", taxableFloat)
	percent := "0.00"
	if taxableFloat > 0 && taxFloat > 0 {
		percent = fmt.Sprintf("%.2f", (taxFloat/taxableFloat)*100)
	}
	quantity := dianFormatQuantity(genericStringValue(payload["cantidad"]), 1)
	softwareID, softwarePIN, _, credErr := resolveDIANSoftwareCredentials(cfg, payload)
	if credErr != nil {
		softwareID = dianFirstNonBlank(genericStringValue(cfg["software_id"]), genericStringValue(payload["software_id"]))
		softwarePIN = dianFirstNonBlank(genericStringValue(cfg["software_pin"]), genericStringValue(payload["software_pin"]))
	}
	softwareSecurityCode := buildDIANSHA384Hex(softwareID, softwarePIN, documentoCodigo)
	uuidSecret := llaveTecnica
	if rootName == "CreditNote" || rootName == "DebitNote" {
		uuidSecret = softwarePIN
	}
	uuidValue := buildDIANSHA384Hex(documentoCodigo, issueDateOnly, issueTime, taxable, "01", impuestoTotal, "04", "0.00", "03", "0.00", total, emisorNIT, clienteNIT, uuidSecret, profileExecutionID)
	referenceID := dianFirstNonBlank(genericStringValue(payload["referencia_documento_codigo"]), genericStringValue(payload["documento_referencia"]), "SETP990000001")
	referenceUUID := dianFirstNonBlank(genericStringValue(payload["referencia_cufe"]), genericStringValue(payload["cufe_referencia"]), buildDIANSHA384Hex(referenceID, issueDateOnly, emisorNIT))
	referenceIssueDate := dianFirstNonBlank(genericStringValue(payload["referencia_fecha_emision"]), issueDateOnly)
	qrURL := "https://catalogo-vpfe-hab.dian.gov.co/Document/FindDocument?documentKey=" + strings.ToLower(uuidValue)
	if chooseDIANAmbiente(cfg) == "produccion" {
		qrURL = "https://catalogo-vpfe.dian.gov.co/Document/FindDocument?documentKey=" + strings.ToLower(uuidValue)
	}
	xmlnsRoot := rootName
	schemaFile := "UBL-" + rootName + "-2.1.xsd"
	invoiceControl := ""
	if rootName == "Invoice" {
		invoiceControl = fmt.Sprintf(`<sts:InvoiceControl><sts:InvoiceAuthorization>%s</sts:InvoiceAuthorization><sts:AuthorizationPeriod><cbc:StartDate>%s</cbc:StartDate><cbc:EndDate>%s</cbc:EndDate></sts:AuthorizationPeriod><sts:AuthorizedInvoices><sts:Prefix>%s</sts:Prefix><sts:From>%d</sts:From><sts:To>%d</sts:To></sts:AuthorizedInvoices></sts:InvoiceControl>`,
			escapeXML(resolucionNumero), escapeXML(resolucionDesde), escapeXML(resolucionHasta), escapeXML(prefijo), rangoDesde, rangoHasta)
	}
	dianExtensions := fmt.Sprintf(`<ext:UBLExtensions><ext:UBLExtension><ext:ExtensionContent><sts:DianExtensions>%s<sts:InvoiceSource><cbc:IdentificationCode listAgencyID="6" listAgencyName="United Nations Economic Commission for Europe" listSchemeURI="urn:oasis:names:specification:ubl:codelist:gc:CountryIdentificationCode-2.1">CO</cbc:IdentificationCode></sts:InvoiceSource><sts:SoftwareProvider><sts:ProviderID schemeAgencyID="195" schemeAgencyName="%s" schemeID="%s" schemeName="31">%s</sts:ProviderID><sts:SoftwareID schemeAgencyID="195" schemeAgencyName="%s">%s</sts:SoftwareID></sts:SoftwareProvider><sts:SoftwareSecurityCode schemeAgencyID="195" schemeAgencyName="%s">%s</sts:SoftwareSecurityCode><sts:AuthorizationProvider><sts:AuthorizationProviderID schemeAgencyID="195" schemeAgencyName="%s" schemeID="4" schemeName="31">800197268</sts:AuthorizationProviderID></sts:AuthorizationProvider><sts:QRCode>NroFactura=%s&#10;NitFacturador=%s&#10;NitAdquiriente=%s&#10;FechaFactura=%s&#10;ValorTotalFactura=%s&#10;CUFE=%s&#10;URL=%s</sts:QRCode></sts:DianExtensions></ext:ExtensionContent></ext:UBLExtension><ext:UBLExtension><ext:ExtensionContent></ext:ExtensionContent></ext:UBLExtension></ext:UBLExtensions>`,
		invoiceControl,
		escapeXML(dianAgencyName),
		escapeXML(dianCompanyIDSchemeID(emisorNIT, emisorDV)),
		escapeXML(emisorNIT),
		escapeXML(dianAgencyName),
		escapeXML(softwareID),
		escapeXML(dianAgencyName),
		escapeXML(softwareSecurityCode),
		escapeXML(dianAgencyName),
		escapeXML(documentoCodigo),
		escapeXML(emisorNIT),
		escapeXML(clienteNIT),
		escapeXML(issueDateOnly),
		escapeXML(total),
		escapeXML(strings.ToLower(uuidValue)),
		escapeXML(qrURL),
	)
	header := fmt.Sprintf(`<cbc:UBLVersionID>UBL 2.1</cbc:UBLVersionID><cbc:CustomizationID>%s</cbc:CustomizationID><cbc:ProfileID>%s</cbc:ProfileID><cbc:ProfileExecutionID>%s</cbc:ProfileExecutionID><cbc:ID>%s</cbc:ID><cbc:UUID schemeID="%s" schemeName="%s">%s</cbc:UUID><cbc:IssueDate>%s</cbc:IssueDate><cbc:IssueTime>%s</cbc:IssueTime>`,
		escapeXML(customizationID), escapeXML(dianDocumentProfileID(rootName)), escapeXML(profileExecutionID), escapeXML(documentoCodigo), escapeXML(profileExecutionID), escapeXML(uuidSchemeName), escapeXML(strings.ToLower(uuidValue)), escapeXML(issueDateOnly), escapeXML(issueTime))
	switch rootName {
	case "Invoice":
		header += fmt.Sprintf(`<cbc:DueDate>%s</cbc:DueDate><cbc:InvoiceTypeCode>%s</cbc:InvoiceTypeCode>`, escapeXML(issueDateOnly), escapeXML(typeCode))
	case "CreditNote":
		header += fmt.Sprintf(`<cbc:CreditNoteTypeCode>%s</cbc:CreditNoteTypeCode>`, escapeXML(typeCode))
	}
	header += fmt.Sprintf(`<cbc:Note>%s</cbc:Note><cbc:DocumentCurrencyCode listAgencyID="6" listAgencyName="United Nations Economic Commission for Europe" listID="ISO 4217 Alpha">%s</cbc:DocumentCurrencyCode><cbc:LineCountNumeric>1</cbc:LineCountNumeric>`,
		escapeXML("Documento electronico generado por Powerful Control System para validacion previa DIAN."), escapeXML(moneda))
	references := ""
	if rootName == "CreditNote" || rootName == "DebitNote" {
		references = fmt.Sprintf(`<cac:DiscrepancyResponse><cbc:ReferenceID>%s</cbc:ReferenceID><cbc:ResponseCode>%s</cbc:ResponseCode><cbc:Description>Ajuste de habilitacion DIAN</cbc:Description></cac:DiscrepancyResponse><cac:BillingReference><cac:InvoiceDocumentReference><cbc:ID>%s</cbc:ID><cbc:UUID schemeName="CUFE-SHA384">%s</cbc:UUID><cbc:IssueDate>%s</cbc:IssueDate></cac:InvoiceDocumentReference></cac:BillingReference>`,
			escapeXML(referenceID), escapeXML(correctionCode), escapeXML(referenceID), escapeXML(strings.ToLower(referenceUUID)), escapeXML(referenceIssueDate))
	}
	supplierParty := dianSupplierPartyXML(emisorNIT, emisorDV, emisorRazon, prefijo, dianFirstNonBlank(genericStringValue(cfg["responsabilidad_fiscal"]), "R-99-PN"))
	customerParty := dianCustomerPartyXML(clienteNombre, clienteNIT)
	xmlPayload := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>`+
		`<%s xmlns="urn:oasis:names:specification:ubl:schema:xsd:%s-2" xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2" xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2" xmlns:ds="http://www.w3.org/2000/09/xmldsig#" xmlns:ext="urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2" xmlns:sts="dian:gov:co:facturaelectronica:Structures-2-1" xmlns:xades="http://uri.etsi.org/01903/v1.3.2#" xmlns:xades141="http://uri.etsi.org/01903/v1.4.1#" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="urn:oasis:names:specification:ubl:schema:xsd:%s-2 http://docs.oasis-open.org/ubl/os-UBL-2.1/xsd/maindoc/%s">`+
		`%s%s%s%s%s%s%s%s</%s>`,
		rootName, xmlnsRoot, xmlnsRoot, schemaFile,
		dianExtensions,
		header,
		references,
		supplierParty,
		customerParty,
		dianPaymentMeansXML(issueDateOnly),
		dianTaxTotalXML(moneda, taxable, impuestoTotal, percent),
		dianMonetaryTotalXML(monetaryTotalTag, moneda, taxable, impuestoTotal, total)+dianLineXML(lineName, quantityTag, moneda, taxable, impuestoTotal, percent, quantity),
		rootName,
	)

	return map[string]interface{}{
		"ok":                      true,
		"empresa_id":              empresaID,
		"documento_codigo":        documentoCodigo,
		"documento_tipo":          documentoTipo,
		"ubl_version":             "UBL 2.1",
		"profile_execution_id":    profileExecutionID,
		"customization_id":        customizationID,
		"uuid_scheme":             uuidSchemeName,
		"uuid":                    strings.ToLower(uuidValue),
		"software_security_code":  "[calculado]",
		"xml_ubl_base":            xmlPayload,
		"estado_preparacion":      "pre_envio_validable",
		"advertencia_oficialidad": "XML UBL 2.1 generado con estructura DIAN, CUFE/CUDE SHA384 y extensiones DIAN; la aceptacion fiscal final la confirma DIAN con su acuse.",
	}, http.StatusOK, nil
}

func parseDIANCertificate(raw string) (*x509.Certificate, error) {
	resolved, err := resolveDIANSecretValue(raw)
	if err != nil {
		return nil, err
	}
	remaining := []byte(resolved)
	for len(remaining) > 0 {
		block, rest := pem.Decode(remaining)
		if block == nil {
			break
		}
		if strings.Contains(strings.ToUpper(strings.TrimSpace(block.Type)), "CERTIFICATE") {
			cert, certErr := x509.ParseCertificate(block.Bytes)
			if certErr == nil {
				return cert, nil
			}
		}
		remaining = rest
	}
	return nil, fmt.Errorf("certificado DIAN no esta en formato PEM X.509")
}

func dianBuildSigningCertificateBlock(cert *x509.Certificate) string {
	if cert == nil {
		return ""
	}
	digest := sha256.Sum256(cert.Raw)
	issuer := escapeXML(cert.Issuer.String())
	serial := escapeXML(cert.SerialNumber.String())
	return fmt.Sprintf(`<xades:SigningCertificate><xades:Cert><xades:CertDigest><ds:DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256"></ds:DigestMethod><ds:DigestValue>%s</ds:DigestValue></xades:CertDigest><xades:IssuerSerial><ds:X509IssuerName>%s</ds:X509IssuerName><ds:X509SerialNumber>%s</ds:X509SerialNumber></xades:IssuerSerial></xades:Cert></xades:SigningCertificate>`, base64.StdEncoding.EncodeToString(digest[:]), issuer, serial)
}

func dianCanonicalXMLDigestInput(xmlPayload string) string {
	xmlPayload = strings.TrimSpace(xmlPayload)
	if strings.HasPrefix(xmlPayload, "<?xml") {
		if idx := strings.Index(xmlPayload, "?>"); idx >= 0 {
			xmlPayload = strings.TrimSpace(xmlPayload[idx+2:])
		}
	}
	return xmlPayload
}

func dianCanonicalPrefixForNamespace(space string) string {
	switch space {
	case "urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2":
		return "cac"
	case "urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2":
		return "cbc"
	case "http://www.w3.org/2000/09/xmldsig#":
		return "ds"
	case "urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2":
		return "ext"
	case "dian:gov:co:facturaelectronica:Structures-2-1":
		return "sts"
	case "http://uri.etsi.org/01903/v1.3.2#":
		return "xades"
	case "http://uri.etsi.org/01903/v1.4.1#":
		return "xades141"
	case "http://www.w3.org/2001/XMLSchema-instance":
		return "xsi"
	case "http://www.w3.org/XML/1998/namespace":
		return "xml"
	default:
		return ""
	}
}

func dianIsDocumentRootNamespace(space string) bool {
	switch space {
	case "urn:oasis:names:specification:ubl:schema:xsd:Invoice-2",
		"urn:oasis:names:specification:ubl:schema:xsd:CreditNote-2",
		"urn:oasis:names:specification:ubl:schema:xsd:DebitNote-2":
		return true
	default:
		return false
	}
}

func dianCanonicalRootNamespaceAttrs(root xml.Name) []string {
	if dianIsDocumentRootNamespace(root.Space) {
		return []string{
			fmt.Sprintf(`xmlns="%s"`, dianEscapeCanonicalAttr(root.Space)),
			`xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"`,
			`xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2"`,
			`xmlns:ds="http://www.w3.org/2000/09/xmldsig#"`,
			`xmlns:ext="urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2"`,
			`xmlns:sts="dian:gov:co:facturaelectronica:Structures-2-1"`,
			`xmlns:xades="http://uri.etsi.org/01903/v1.3.2#"`,
			`xmlns:xades141="http://uri.etsi.org/01903/v1.4.1#"`,
			`xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"`,
		}
	}
	switch root.Space {
	case "http://www.w3.org/2000/09/xmldsig#":
		return []string{`xmlns:ds="http://www.w3.org/2000/09/xmldsig#"`}
	case "http://uri.etsi.org/01903/v1.3.2#":
		return []string{
			`xmlns:ds="http://www.w3.org/2000/09/xmldsig#"`,
			`xmlns:xades="http://uri.etsi.org/01903/v1.3.2#"`,
		}
	default:
		return nil
	}
}

func dianCanonicalElementName(name xml.Name, rootSpace string) string {
	if dianIsDocumentRootNamespace(rootSpace) && name.Space == rootSpace {
		return name.Local
	}
	if prefix := dianCanonicalPrefixForNamespace(name.Space); prefix != "" {
		return prefix + ":" + name.Local
	}
	return name.Local
}

func dianEscapeCanonicalText(v string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\r", "&#xD;",
	)
	return replacer.Replace(v)
}

func dianEscapeCanonicalAttr(v string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		"\"", "&quot;",
		"\t", "&#x9;",
		"\n", "&#xA;",
		"\r", "&#xD;",
	)
	return replacer.Replace(v)
}

func dianCanonicalAttrName(attr xml.Attr) (name, sortKey string, ok bool) {
	if attr.Name.Space == "xmlns" || attr.Name.Space == "http://www.w3.org/2000/xmlns/" || attr.Name.Local == "xmlns" {
		return "", "", false
	}
	if attr.Name.Space == "" {
		return attr.Name.Local, "\x00" + attr.Name.Local, true
	}
	prefix := dianCanonicalPrefixForNamespace(attr.Name.Space)
	if prefix == "" {
		return attr.Name.Local, attr.Name.Space + "\x00" + attr.Name.Local, true
	}
	return prefix + ":" + attr.Name.Local, attr.Name.Space + "\x00" + attr.Name.Local, true
}

func dianCanonicalizeXML(raw string) (string, error) {
	raw = dianCanonicalXMLDigestInput(raw)
	decoder := xml.NewDecoder(strings.NewReader(raw))
	var builder strings.Builder
	var rootSpace string
	depth := 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		switch node := token.(type) {
		case xml.StartElement:
			if depth == 0 {
				rootSpace = node.Name.Space
			}
			builder.WriteString("<")
			builder.WriteString(dianCanonicalElementName(node.Name, rootSpace))
			if depth == 0 {
				for _, nsAttr := range dianCanonicalRootNamespaceAttrs(node.Name) {
					builder.WriteByte(' ')
					builder.WriteString(nsAttr)
				}
			}
			attrs := make([]struct {
				name    string
				value   string
				sortKey string
			}, 0, len(node.Attr))
			for _, attr := range node.Attr {
				attrName, sortKey, ok := dianCanonicalAttrName(attr)
				if !ok {
					continue
				}
				attrs = append(attrs, struct {
					name    string
					value   string
					sortKey string
				}{name: attrName, value: attr.Value, sortKey: sortKey})
			}
			sort.Slice(attrs, func(i, j int) bool { return attrs[i].sortKey < attrs[j].sortKey })
			for _, attr := range attrs {
				builder.WriteByte(' ')
				builder.WriteString(attr.name)
				builder.WriteString(`="`)
				builder.WriteString(dianEscapeCanonicalAttr(attr.value))
				builder.WriteByte('"')
			}
			builder.WriteString(">")
			depth++
		case xml.EndElement:
			depth--
			builder.WriteString("</")
			builder.WriteString(dianCanonicalElementName(node.Name, rootSpace))
			builder.WriteString(">")
		case xml.CharData:
			builder.WriteString(dianEscapeCanonicalText(string(node)))
		}
	}
	return builder.String(), nil
}

func dianCanonicalSHA256Base64(xmlPayload string) (string, error) {
	canonical, err := dianCanonicalizeXML(xmlPayload)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256([]byte(canonical))
	return base64.StdEncoding.EncodeToString(digest[:]), nil
}

func dianBuildXAdESBaseSignature(xmlPayload string, privateKey *rsa.PrivateKey, cert *x509.Certificate) (map[string]string, error) {
	documentDigestBase64, err := dianCanonicalSHA256Base64(xmlPayload)
	if err != nil {
		return nil, fmt.Errorf("no se pudo canonicalizar XML DIAN para digest: %w", err)
	}
	signingTime := time.Now().Format(time.RFC3339)
	signedPropertiesID := "SignedPropertiesPCS"
	signatureID := "SignaturePCS"
	keyInfoID := "KeyInfoPCS"

	signaturePolicy := `<xades:SignaturePolicyIdentifier><xades:SignaturePolicyId><xades:SigPolicyId><xades:Identifier>https://facturaelectronica.dian.gov.co/politicadefirma/v1/politicadefirmav2.pdf</xades:Identifier></xades:SigPolicyId><xades:SigPolicyHash><ds:DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256"></ds:DigestMethod><ds:DigestValue>dMoMvtcG5aIzgYo0tIsSQeVJBDnUnfSOfBpxXrmor0Y=</ds:DigestValue></xades:SigPolicyHash></xades:SignaturePolicyId></xades:SignaturePolicyIdentifier>`
	signerRole := `<xades:SignerRole><xades:ClaimedRoles><xades:ClaimedRole>supplier</xades:ClaimedRole></xades:ClaimedRoles></xades:SignerRole>`
	signedProperties := fmt.Sprintf(`<xades:SignedProperties xmlns:ds="http://www.w3.org/2000/09/xmldsig#" xmlns:xades="http://uri.etsi.org/01903/v1.3.2#" Id="%s"><xades:SignedSignatureProperties><xades:SigningTime>%s</xades:SigningTime>%s%s%s</xades:SignedSignatureProperties></xades:SignedProperties>`, signedPropertiesID, escapeXML(signingTime), dianBuildSigningCertificateBlock(cert), signaturePolicy, signerRole)
	propsDigestBase64, err := dianCanonicalSHA256Base64(signedProperties)
	if err != nil {
		return nil, fmt.Errorf("no se pudo canonicalizar SignedProperties DIAN: %w", err)
	}

	keyInfo := ""
	keyInfoReference := ""
	if cert != nil {
		keyInfo = fmt.Sprintf(`<ds:KeyInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#" Id="%s"><ds:X509Data><ds:X509Certificate>%s</ds:X509Certificate></ds:X509Data></ds:KeyInfo>`, keyInfoID, base64.StdEncoding.EncodeToString(cert.Raw))
		keyInfoDigestBase64, err := dianCanonicalSHA256Base64(keyInfo)
		if err != nil {
			return nil, fmt.Errorf("no se pudo canonicalizar KeyInfo DIAN: %w", err)
		}
		keyInfoReference = fmt.Sprintf(`<ds:Reference URI="#%s"><ds:DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256"></ds:DigestMethod><ds:DigestValue>%s</ds:DigestValue></ds:Reference>`, keyInfoID, keyInfoDigestBase64)
	}

	signedInfo := fmt.Sprintf(`<ds:SignedInfo xmlns:ds="http://www.w3.org/2000/09/xmldsig#"><ds:CanonicalizationMethod Algorithm="http://www.w3.org/TR/2001/REC-xml-c14n-20010315"></ds:CanonicalizationMethod><ds:SignatureMethod Algorithm="http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"></ds:SignatureMethod><ds:Reference Id="%s-ref0" URI=""><ds:Transforms><ds:Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"></ds:Transform></ds:Transforms><ds:DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256"></ds:DigestMethod><ds:DigestValue>%s</ds:DigestValue></ds:Reference>%s<ds:Reference Type="http://uri.etsi.org/01903#SignedProperties" URI="#%s"><ds:DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256"></ds:DigestMethod><ds:DigestValue>%s</ds:DigestValue></ds:Reference></ds:SignedInfo>`, signatureID, documentDigestBase64, keyInfoReference, signedPropertiesID, propsDigestBase64)
	canonicalSignedInfo, err := dianCanonicalizeXML(signedInfo)
	if err != nil {
		return nil, fmt.Errorf("no se pudo canonicalizar SignedInfo DIAN: %w", err)
	}

	signedInfoDigest := sha256.Sum256([]byte(canonicalSignedInfo))
	signatureValue, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, signedInfoDigest[:])
	if err != nil {
		return nil, fmt.Errorf("no se pudo firmar SignedInfo con RSA-SHA256")
	}

	signatureXML := fmt.Sprintf(`<ds:Signature Id="%s" xmlns:ds="http://www.w3.org/2000/09/xmldsig#">%s<ds:SignatureValue Id="%s-sigvalue">%s</ds:SignatureValue>%s<ds:Object><xades:QualifyingProperties xmlns:xades="http://uri.etsi.org/01903/v1.3.2#" Target="#%s">%s</xades:QualifyingProperties></ds:Object></ds:Signature>`,
		signatureID,
		canonicalSignedInfo,
		signatureID,
		base64.StdEncoding.EncodeToString(signatureValue),
		keyInfo,
		signatureID,
		signedProperties,
	)

	return map[string]string{
		"document_digest_base64":          documentDigestBase64,
		"signed_properties_digest_base64": propsDigestBase64,
		"signature_value_base64":          base64.StdEncoding.EncodeToString(signatureValue),
		"signature_xml":                   signatureXML,
		"signing_time":                    signingTime,
	}, nil
}

func dianInjectSignatureIntoXML(xmlPayload, signatureXML string) string {
	if strings.Contains(xmlPayload, "<ext:ExtensionContent></ext:ExtensionContent>") {
		return strings.Replace(xmlPayload, "<ext:ExtensionContent></ext:ExtensionContent>", "<ext:ExtensionContent>"+signatureXML+"</ext:ExtensionContent>", 1)
	}
	if strings.Contains(xmlPayload, "<ext:ExtensionContent/>") {
		return strings.Replace(xmlPayload, "<ext:ExtensionContent/>", "<ext:ExtensionContent>"+signatureXML+"</ext:ExtensionContent>", 1)
	}
	rootName := dianDetectXMLRootName(xmlPayload)
	if rootName == "" {
		return xmlPayload + signatureXML
	}
	closingTag := "</" + rootName + ">"
	idx := strings.LastIndex(xmlPayload, closingTag)
	if idx < 0 {
		return xmlPayload + signatureXML
	}
	return xmlPayload[:idx] + signatureXML + xmlPayload[idx:]
}

func dianDetectXMLRootName(xmlPayload string) string {
	trimmed := strings.TrimSpace(xmlPayload)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "<?xml") {
		if idx := strings.Index(trimmed, "?>"); idx >= 0 {
			trimmed = strings.TrimSpace(trimmed[idx+2:])
		}
	}
	if !strings.HasPrefix(trimmed, "<") {
		return ""
	}
	trimmed = trimmed[1:]
	end := len(trimmed)
	for i, ch := range trimmed {
		if ch == ' ' || ch == '>' || ch == '\t' || ch == '\n' || ch == '\r' {
			end = i
			break
		}
	}
	return strings.TrimSpace(trimmed[:end])
}

func signDIANXMLXAdESBase(cfg map[string]interface{}, empresaID int64, payload map[string]interface{}) (map[string]interface{}, int, error) {
	xmlPayload := dianFirstNonBlank(
		genericStringValue(payload["xml_ubl_base"]),
		genericStringValue(payload["xml"]),
		genericStringValue(payload["xml_demo"]),
		genericStringValue(payload["xml_firmado"]),
	)
	if xmlPayload == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("xml_ubl_base/xml es obligatorio para firma XAdES base")
	}

	keyRef := dianFirstNonBlank(
		genericStringValue(payload["private_key_pem"]),
		genericStringValue(payload["certificado_clave_ref"]),
		genericStringValue(cfg["certificado_clave_ref"]),
	)
	if keyRef == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("certificado_clave_ref es obligatorio para firma XAdES base")
	}

	privateKey, err := parseDIANRSAPrivateKey(keyRef)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	certificateRef := dianFirstNonBlank(
		genericStringValue(payload["certificado_pem"]),
		genericStringValue(payload["certificado_ref"]),
		genericStringValue(payload["certificado_x509_ref"]),
		genericStringValue(cfg["certificado_url"]),
	)
	var certificate *x509.Certificate
	certificateIncluded := false
	warnings := make([]string, 0)
	if strings.TrimSpace(certificateRef) != "" {
		certificate, err = parseDIANCertificate(certificateRef)
		if err != nil {
			warnings = append(warnings, err.Error())
		} else {
			certificateIncluded = true
		}
	} else {
		warnings = append(warnings, "certificado X.509 no suministrado; la firma base se genera sin KeyInfo/X509Certificate")
	}

	signatureData, err := dianBuildXAdESBaseSignature(xmlPayload, privateKey, certificate)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	xmlFirmado := dianInjectSignatureIntoXML(xmlPayload, signatureData["signature_xml"])

	return map[string]interface{}{
		"ok":                              true,
		"empresa_id":                      empresaID,
		"documento_codigo":                dianFirstNonBlank(genericStringValue(payload["documento_codigo"]), "FV-"+time.Now().Format("20060102150405")),
		"nivel_firma":                     "xades_base_validacion_preventiva",
		"algoritmo":                       "RSA-SHA256",
		"certificate_included":            certificateIncluded,
		"warnings":                        warnings,
		"digest_documento_base64":         signatureData["document_digest_base64"],
		"digest_signed_properties_base64": signatureData["signed_properties_digest_base64"],
		"signature_value_base64":          signatureData["signature_value_base64"],
		"xml_signature":                   signatureData["signature_xml"],
		"xml_firmado":                     xmlFirmado,
		"signing_time":                    signatureData["signing_time"],
		"advertencia_oficialidad":         "Firma XAdES generada para validacion preventiva; el acuse DIAN/proveedor compatible confirma la aceptacion fiscal.",
		"gestion_secreto":                 "certificado_clave_ref y certificado_pem/certificado_ref admiten env:/file:/base64: o valor inline controlado",
	}, http.StatusOK, nil
}

func isDIANOfficialEndpoint(endpoint string) bool {
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Host)
	return strings.Contains(host, "dian.gov.co")
}

func dianConfiguredEndpoint(cfg map[string]interface{}, payload map[string]interface{}) string {
	if payload == nil {
		payload = map[string]interface{}{}
	}
	return normalizeIntegracionEndpoint(dianFirstNonBlank(
		genericStringValue(payload["url_dian"]),
		genericStringValue(payload["endpoint"]),
		genericStringValue(cfg["url_dian"]),
	))
}

func dianTokenRequiredForEndpoint(cfg map[string]interface{}, payload map[string]interface{}) bool {
	endpoint := dianConfiguredEndpoint(cfg, payload)
	if endpoint == "" {
		return false
	}
	return !isDIANOfficialEndpoint(endpoint)
}

func normalizeDIANSOAPEndpoint(endpoint string) string {
	endpoint = normalizeIntegracionEndpoint(endpoint)
	if endpoint == "" {
		return ""
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return endpoint
	}
	query := parsed.Query()
	query.Del("wsdl")
	query.Del("singleWsdl")
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func buildDIANZipContent(fileName, xmlContent string) ([]byte, error) {
	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		fileName = "documento.xml"
	}
	if !strings.HasSuffix(strings.ToLower(fileName), ".xml") {
		fileName += ".xml"
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create(fileName)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write([]byte(xmlContent)); err != nil {
		_ = zw.Close()
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

const (
	dianSOAPNamespace       = "http://www.w3.org/2003/05/soap-envelope"
	dianAddressingNamespace = "http://www.w3.org/2005/08/addressing"
	dianWCFNamespace        = "http://wcf.dian.colombia"
	dianWSUSecurityNS       = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd"
	dianWSSESecurityNS      = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd"
	dianWSSX509ValueType    = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-x509-token-profile-1.0#X509v3"
	dianWSSBase64Encoding   = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-soap-message-security-1.0#Base64Binary"
	dianDSigNamespace       = "http://www.w3.org/2000/09/xmldsig#"
	dianExcC14NAlgorithm    = "http://www.w3.org/2001/10/xml-exc-c14n#"
	dianRSASHA256Algorithm  = "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"
	dianSHA256DigestAlg     = "http://www.w3.org/2001/04/xmlenc#sha256"
)

func dianSOAPSafeID(prefix string) string {
	raw := make([]byte, 12)
	if _, err := rand.Read(raw); err != nil {
		return prefix + "-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return prefix + "-" + hex.EncodeToString(raw)
}

func dianSOAPSHA256DigestBase64(canonicalXML string) string {
	digest := sha256.Sum256([]byte(canonicalXML))
	return base64.StdEncoding.EncodeToString(digest[:])
}

func dianSOAPSignedToReference(id, canonicalXML string) string {
	return fmt.Sprintf(`<ds:Reference URI="#%s"><ds:Transforms><ds:Transform Algorithm="%s"><ec:InclusiveNamespaces xmlns:ec="%s" PrefixList="soap wcf"></ec:InclusiveNamespaces></ds:Transform></ds:Transforms><ds:DigestMethod Algorithm="%s"></ds:DigestMethod><ds:DigestValue>%s</ds:DigestValue></ds:Reference>`,
		escapeXML(id),
		dianExcC14NAlgorithm,
		dianExcC14NAlgorithm,
		dianSHA256DigestAlg,
		dianSOAPSHA256DigestBase64(canonicalXML),
	)
}

func dianSOAPCanonicalToHeader(id, endpoint string) string {
	return fmt.Sprintf(`<wsa:To xmlns:soap="%s" xmlns:wcf="%s" xmlns:wsa="%s" xmlns:wsu="%s" wsu:Id="%s">%s</wsa:To>`,
		dianSOAPNamespace,
		dianWCFNamespace,
		dianAddressingNamespace,
		dianWSUSecurityNS,
		escapeXML(id),
		escapeXML(endpoint),
	)
}

func dianSOAPCanonicalSignedInfoForTo(toID, toCanonicalXML string) string {
	return fmt.Sprintf(`<ds:SignedInfo xmlns:ds="%s" xmlns:soap="%s" xmlns:wcf="%s" xmlns:wsa="%s"><ds:CanonicalizationMethod Algorithm="%s"><ec:InclusiveNamespaces xmlns:ec="%s" PrefixList="wsa soap wcf"></ec:InclusiveNamespaces></ds:CanonicalizationMethod><ds:SignatureMethod Algorithm="%s"></ds:SignatureMethod>%s</ds:SignedInfo>`,
		dianDSigNamespace,
		dianSOAPNamespace,
		dianWCFNamespace,
		dianAddressingNamespace,
		dianExcC14NAlgorithm,
		dianExcC14NAlgorithm,
		dianRSASHA256Algorithm,
		dianSOAPSignedToReference(toID, toCanonicalXML),
	)
}

func dianSOAPCanonicalTimestamp(id, created, expires string) string {
	return fmt.Sprintf(`<wsu:Timestamp xmlns:wsu="%s" wsu:Id="%s"><wsu:Created>%s</wsu:Created><wsu:Expires>%s</wsu:Expires></wsu:Timestamp>`,
		dianWSUSecurityNS,
		escapeXML(id),
		escapeXML(created),
		escapeXML(expires),
	)
}

func dianBuildSOAPBody(operation, fileName string, zipBytes []byte, testSetID string) string {
	contentBase64 := base64.StdEncoding.EncodeToString(zipBytes)
	switch operation {
	case "SendTestSetAsync":
		return fmt.Sprintf(`<wcf:SendTestSetAsync><wcf:fileName>%s</wcf:fileName><wcf:contentFile>%s</wcf:contentFile><wcf:testSetId>%s</wcf:testSetId></wcf:SendTestSetAsync>`,
			escapeXML(fileName), contentBase64, escapeXML(testSetID))
	case "SendBillAsync":
		return fmt.Sprintf(`<wcf:SendBillAsync><wcf:fileName>%s</wcf:fileName><wcf:contentFile>%s</wcf:contentFile></wcf:SendBillAsync>`,
			escapeXML(fileName), contentBase64)
	default:
		return fmt.Sprintf(`<wcf:SendBillSync><wcf:fileName>%s</wcf:fileName><wcf:contentFile>%s</wcf:contentFile></wcf:SendBillSync>`,
			escapeXML(fileName), contentBase64)
	}
}

func buildDIANSOAPEnvelopeWithWSSecurity(operation, endpoint, fileName string, zipBytes []byte, testSetID string, privateKey *rsa.PrivateKey, cert *x509.Certificate, now time.Time) (string, map[string]interface{}, error) {
	if privateKey == nil {
		return "", nil, fmt.Errorf("llave privada DIAN requerida para WS-Security")
	}
	if cert == nil {
		return "", nil, fmt.Errorf("certificado X.509 requerido para WS-Security")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	action := dianSOAPAction(operation)
	bodyContent := dianBuildSOAPBody(operation, fileName, zipBytes, testSetID)
	toID := dianSOAPSafeID("ID")
	timestampID := dianSOAPSafeID("PCSTS")
	tokenID := dianSOAPSafeID("X509")
	signatureID := dianSOAPSafeID("SIG")
	keyInfoID := dianSOAPSafeID("KI")
	strID := dianSOAPSafeID("STR")
	messageID := "urn:uuid:" + strings.TrimPrefix(dianSOAPSafeID(""), "-")

	created := now.Format("2006-01-02T15:04:05.000Z")
	expires := now.Add(60 * time.Second).Format("2006-01-02T15:04:05.000Z")
	actionHeader := fmt.Sprintf(`<wsa:Action xmlns:wsa="%s">%s</wsa:Action>`,
		dianAddressingNamespace, escapeXML(action))
	messageIDHeader := fmt.Sprintf(`<wsa:MessageID xmlns:wsa="%s">%s</wsa:MessageID>`,
		dianAddressingNamespace, escapeXML(messageID))
	replyToHeader := fmt.Sprintf(`<wsa:ReplyTo xmlns:wsa="%s"><wsa:Address>http://www.w3.org/2005/08/addressing/anonymous</wsa:Address></wsa:ReplyTo>`,
		dianAddressingNamespace)
	toHeader := fmt.Sprintf(`<wsa:To xmlns:wsu="%s" wsu:Id="%s">%s</wsa:To>`,
		dianWSUSecurityNS, toID, escapeXML(endpoint))
	body := fmt.Sprintf(`<soap:Body xmlns:soap="%s" xmlns:wcf="%s">%s</soap:Body>`,
		dianSOAPNamespace, dianWCFNamespace, bodyContent)
	timestamp := fmt.Sprintf(`<wsu:Timestamp wsu:Id="%s" xmlns:wsu="%s"><wsu:Created>%s</wsu:Created><wsu:Expires>%s</wsu:Expires></wsu:Timestamp>`,
		timestampID, dianWSUSecurityNS, created, expires)
	binaryToken := fmt.Sprintf(`<wsse:BinarySecurityToken wsu:Id="%s" EncodingType="%s" ValueType="%s" xmlns:wsse="%s" xmlns:wsu="%s">%s</wsse:BinarySecurityToken>`,
		tokenID, dianWSSBase64Encoding, dianWSSX509ValueType, dianWSSESecurityNS, dianWSUSecurityNS, base64.StdEncoding.EncodeToString(cert.Raw))

	signedInfo := dianSOAPCanonicalSignedInfoForTo(toID, dianSOAPCanonicalToHeader(toID, endpoint))
	signedInfoDigest := sha256.Sum256([]byte(signedInfo))
	signatureValue, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, signedInfoDigest[:])
	if err != nil {
		return "", nil, fmt.Errorf("no se pudo firmar WS-Security DIAN")
	}
	signature := fmt.Sprintf(`<ds:Signature Id="%s" xmlns:ds="%s">%s<ds:SignatureValue>%s</ds:SignatureValue><ds:KeyInfo Id="%s"><wsse:SecurityTokenReference wsu:Id="%s" xmlns:wsse="%s" xmlns:wsu="%s"><wsse:Reference URI="#%s" ValueType="%s"></wsse:Reference></wsse:SecurityTokenReference></ds:KeyInfo></ds:Signature>`,
		signatureID,
		dianDSigNamespace,
		signedInfo,
		base64.StdEncoding.EncodeToString(signatureValue),
		keyInfoID,
		strID,
		dianWSSESecurityNS,
		dianWSUSecurityNS,
		tokenID,
		dianWSSX509ValueType,
	)
	security := fmt.Sprintf(`<wsse:Security xmlns:wsse="%s" xmlns:wsu="%s">%s%s%s</wsse:Security>`,
		dianWSSESecurityNS, dianWSUSecurityNS, timestamp, binaryToken, signature)
	envelope := fmt.Sprintf(`<soap:Envelope xmlns:soap="%s" xmlns:wcf="%s"><soap:Header xmlns:wsa="%s">%s%s%s%s%s</soap:Header>%s</soap:Envelope>`,
		dianSOAPNamespace, dianWCFNamespace, dianAddressingNamespace, security, actionHeader, messageIDHeader, replyToHeader, toHeader, body)
	meta := map[string]interface{}{
		"ws_security":              true,
		"signed_parts":             []string{"To"},
		"key_reference":            "BinarySecurityTokenReference",
		"signature_algorithm":      "RSA-SHA256",
		"digest_algorithm":         "SHA-256",
		"canonicalization":         "exclusive_c14n",
		"timestamp_created":        created,
		"timestamp_expires":        expires,
		"binary_security_token_id": tokenID,
		"security_layout":          "strict_timestamp_token_signature",
		"message_id":               messageID,
	}
	return envelope, meta, nil
}

func dianSOAPOperationFromPayload(payload map[string]interface{}, ambiente, testSetID string) string {
	operation := strings.TrimSpace(dianPayloadString(payload, "soap_operacion", "operacion_soap", "operation"))
	if operation != "" {
		return operation
	}
	if parseTruthy(dianPayloadString(payload, "set_habilitacion", "test_set", "testset")) || strings.TrimSpace(testSetID) != "" && chooseDIANAmbiente(map[string]interface{}{"tipo_ambiente": ambiente}) == "habilitacion" {
		return "SendTestSetAsync"
	}
	if parseTruthy(dianPayloadString(payload, "asincrono", "async")) {
		return "SendBillAsync"
	}
	return "SendBillSync"
}

func dianSOAPAction(operation string) string {
	operation = strings.TrimSpace(operation)
	if operation == "" {
		operation = "SendBillSync"
	}
	return "http://wcf.dian.colombia/IWcfDianCustomerServices/" + operation
}

func buildDIANSOAPEnvelope(operation, endpoint, fileName string, zipBytes []byte, testSetID string) string {
	action := dianSOAPAction(operation)
	body := dianBuildSOAPBody(operation, fileName, zipBytes, testSetID)
	return fmt.Sprintf(`<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://www.w3.org/2005/08/addressing" xmlns:wcf="http://wcf.dian.colombia"><s:Header><a:Action s:mustUnderstand="1">%s</a:Action><a:To s:mustUnderstand="1">%s</a:To></s:Header><s:Body>%s</s:Body></s:Envelope>`,
		escapeXML(action), escapeXML(endpoint), body)
}

func buildDIANGetStatusZipEnvelope(endpoint, trackID string) string {
	operation := "GetStatusZip"
	action := dianSOAPAction(operation)
	body := fmt.Sprintf(`<wcf:GetStatusZip><wcf:trackId>%s</wcf:trackId></wcf:GetStatusZip>`, escapeXML(trackID))
	return fmt.Sprintf(`<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://www.w3.org/2005/08/addressing" xmlns:wcf="http://wcf.dian.colombia"><s:Header><a:Action s:mustUnderstand="1">%s</a:Action><a:To s:mustUnderstand="1">%s</a:To></s:Header><s:Body>%s</s:Body></s:Envelope>`,
		escapeXML(action), escapeXML(endpoint), body)
}

func buildDIANGetStatusZipEnvelopeWithWSSecurity(endpoint, trackID string, privateKey *rsa.PrivateKey, cert *x509.Certificate, now time.Time) (string, map[string]interface{}, error) {
	if privateKey == nil {
		return "", nil, fmt.Errorf("llave privada DIAN requerida para WS-Security")
	}
	if cert == nil {
		return "", nil, fmt.Errorf("certificado X.509 requerido para WS-Security")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	action := dianSOAPAction("GetStatusZip")
	toID := dianSOAPSafeID("ID")
	timestampID := dianSOAPSafeID("PCSTS")
	tokenID := dianSOAPSafeID("X509")
	signatureID := dianSOAPSafeID("SIG")
	keyInfoID := dianSOAPSafeID("KI")
	strID := dianSOAPSafeID("STR")
	messageID := "urn:uuid:" + strings.TrimPrefix(dianSOAPSafeID(""), "-")

	created := now.Format("2006-01-02T15:04:05.000Z")
	expires := now.Add(60 * time.Second).Format("2006-01-02T15:04:05.000Z")
	actionHeader := fmt.Sprintf(`<wsa:Action xmlns:wsa="%s">%s</wsa:Action>`,
		dianAddressingNamespace, escapeXML(action))
	messageIDHeader := fmt.Sprintf(`<wsa:MessageID xmlns:wsa="%s">%s</wsa:MessageID>`,
		dianAddressingNamespace, escapeXML(messageID))
	replyToHeader := fmt.Sprintf(`<wsa:ReplyTo xmlns:wsa="%s"><wsa:Address>http://www.w3.org/2005/08/addressing/anonymous</wsa:Address></wsa:ReplyTo>`,
		dianAddressingNamespace)
	toHeader := fmt.Sprintf(`<wsa:To xmlns:wsu="%s" wsu:Id="%s">%s</wsa:To>`,
		dianWSUSecurityNS, toID, escapeXML(endpoint))
	body := fmt.Sprintf(`<soap:Body xmlns:soap="%s" xmlns:wcf="%s"><wcf:GetStatusZip><wcf:trackId>%s</wcf:trackId></wcf:GetStatusZip></soap:Body>`,
		dianSOAPNamespace, dianWCFNamespace, escapeXML(trackID))
	timestamp := fmt.Sprintf(`<wsu:Timestamp wsu:Id="%s" xmlns:wsu="%s"><wsu:Created>%s</wsu:Created><wsu:Expires>%s</wsu:Expires></wsu:Timestamp>`,
		timestampID, dianWSUSecurityNS, created, expires)
	binaryToken := fmt.Sprintf(`<wsse:BinarySecurityToken wsu:Id="%s" EncodingType="%s" ValueType="%s" xmlns:wsse="%s" xmlns:wsu="%s">%s</wsse:BinarySecurityToken>`,
		tokenID, dianWSSBase64Encoding, dianWSSX509ValueType, dianWSSESecurityNS, dianWSUSecurityNS, base64.StdEncoding.EncodeToString(cert.Raw))

	signedInfo := dianSOAPCanonicalSignedInfoForTo(toID, dianSOAPCanonicalToHeader(toID, endpoint))
	signedInfoDigest := sha256.Sum256([]byte(signedInfo))
	signatureValue, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, signedInfoDigest[:])
	if err != nil {
		return "", nil, fmt.Errorf("no se pudo firmar WS-Security DIAN")
	}
	signature := fmt.Sprintf(`<ds:Signature Id="%s" xmlns:ds="%s">%s<ds:SignatureValue>%s</ds:SignatureValue><ds:KeyInfo Id="%s"><wsse:SecurityTokenReference wsu:Id="%s" xmlns:wsse="%s" xmlns:wsu="%s"><wsse:Reference URI="#%s" ValueType="%s"></wsse:Reference></wsse:SecurityTokenReference></ds:KeyInfo></ds:Signature>`,
		signatureID,
		dianDSigNamespace,
		signedInfo,
		base64.StdEncoding.EncodeToString(signatureValue),
		keyInfoID,
		strID,
		dianWSSESecurityNS,
		dianWSUSecurityNS,
		tokenID,
		dianWSSX509ValueType,
	)
	security := fmt.Sprintf(`<wsse:Security xmlns:wsse="%s" xmlns:wsu="%s">%s%s%s</wsse:Security>`,
		dianWSSESecurityNS, dianWSUSecurityNS, timestamp, binaryToken, signature)
	envelope := fmt.Sprintf(`<soap:Envelope xmlns:soap="%s" xmlns:wcf="%s"><soap:Header xmlns:wsa="%s">%s%s%s%s%s</soap:Header>%s</soap:Envelope>`,
		dianSOAPNamespace, dianWCFNamespace, dianAddressingNamespace, security, actionHeader, messageIDHeader, replyToHeader, toHeader, body)
	meta := map[string]interface{}{
		"ws_security":         true,
		"signed_parts":        []string{"To"},
		"key_reference":       "BinarySecurityTokenReference",
		"signature_algorithm": "RSA-SHA256",
		"digest_algorithm":    "SHA-256",
		"canonicalization":    "exclusive_c14n",
		"timestamp_created":   created,
		"timestamp_expires":   expires,
		"security_layout":     "strict_timestamp_token_signature",
		"message_id":          messageID,
	}
	return envelope, meta, nil
}

func shouldUseDIANSOAPTransport(endpoint string, cfg map[string]interface{}, payload map[string]interface{}) bool {
	if isDIANOfficialEndpoint(endpoint) {
		return true
	}
	if strings.TrimSpace(dianPayloadString(payload, "soap_operacion", "operacion_soap", "operation")) != "" {
		return true
	}
	if parseTruthy(dianPayloadString(payload, "usar_soap_dian", "soap_dian", "transporte_soap")) {
		return true
	}
	return false
}

func extractDIANSOAPTag(raw string, names ...string) string {
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		patterns := []string{
			"<" + name + ">",
			":" + name + ">",
		}
		for _, openMarker := range patterns {
			idx := strings.Index(raw, openMarker)
			if idx < 0 {
				continue
			}
			start := idx + len(openMarker)
			end := strings.Index(raw[start:], "</")
			if end < 0 {
				continue
			}
			return strings.TrimSpace(raw[start : start+end])
		}
	}
	return ""
}

var (
	dianSOAPErrorListRE    = regexp.MustCompile(`(?is)<(?:[A-Za-z0-9_.-]+:)?ErrorMessageList[^>]*>(.*?)</(?:[A-Za-z0-9_.-]+:)?ErrorMessageList>`)
	dianSOAPErrorMessageRE = regexp.MustCompile(`(?is)<(?:[A-Za-z0-9_.-]+:)?(?:string|ErrorMessage|Message|Description)[^>]*>(.*?)</(?:[A-Za-z0-9_.-]+:)?(?:string|ErrorMessage|Message|Description)>`)
	dianSOAPAnyTagRE       = regexp.MustCompile(`(?is)<[^>]+>`)
)

func dianSOAPCleanText(raw string) string {
	raw = dianSOAPAnyTagRE.ReplaceAllString(raw, " ")
	replacements := map[string]string{
		"&lt;":   "<",
		"&gt;":   ">",
		"&amp;":  "&",
		"&quot;": `"`,
		"&#39;":  "'",
	}
	for old, repl := range replacements {
		raw = strings.ReplaceAll(raw, old, repl)
	}
	return strings.Join(strings.Fields(strings.TrimSpace(raw)), " ")
}

func extractDIANSOAPErrorMessages(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0)
	add := func(value string) {
		value = dianSOAPCleanText(value)
		if value == "" || seen[value] {
			return
		}
		seen[value] = true
		out = append(out, value)
	}
	scopes := []string{raw}
	for _, match := range dianSOAPErrorListRE.FindAllStringSubmatch(raw, -1) {
		if len(match) > 1 {
			scopes = append(scopes, match[1])
		}
	}
	for _, scope := range scopes {
		for _, match := range dianSOAPErrorMessageRE.FindAllStringSubmatch(scope, -1) {
			if len(match) > 1 {
				add(match[1])
			}
		}
	}
	for _, match := range dianSOAPErrorListRE.FindAllStringSubmatch(raw, -1) {
		if len(match) > 1 {
			children := dianSOAPErrorMessageRE.FindAllStringSubmatch(match[1], -1)
			for _, child := range children {
				if len(child) > 1 {
					add(child[1])
				}
			}
			if len(children) == 0 {
				add(match[1])
			}
		}
	}
	return out
}

func extractDIANSOAPResponseMap(raw string) map[string]interface{} {
	out := map[string]interface{}{}
	if raw = strings.TrimSpace(raw); raw == "" {
		return out
	}
	for key, aliases := range map[string][]string{
		"track_id":           {"ZipKey", "TrackId", "trackId"},
		"status_code":        {"StatusCode", "statusCode"},
		"status_message":     {"StatusMessage", "statusMessage"},
		"status_description": {"StatusDescription", "statusDescription"},
		"error_message":      {"ErrorMessage", "errorMessage"},
		"is_valid":           {"IsValid", "isValid"},
		"xml_document_key":   {"XmlDocumentKey", "xmlDocumentKey"},
	} {
		if value := extractDIANSOAPTag(raw, aliases...); value != "" {
			out[key] = value
		}
	}
	if messages := extractDIANSOAPErrorMessages(raw); len(messages) > 0 {
		out["error_messages"] = messages
		out["error_message"] = strings.Join(messages, " | ")
	}
	out["raw_xml"] = raw
	return out
}

func dianSafeTrackHistoryJSON(raw map[string]interface{}) string {
	if len(raw) == 0 {
		return ""
	}
	out := make(map[string]interface{}, len(raw))
	for key, value := range raw {
		lowerKey := strings.ToLower(strings.TrimSpace(key))
		if lowerKey == "" || lowerKey == "raw_xml" || lowerKey == "raw_response" {
			continue
		}
		if strings.Contains(lowerKey, "certificado") || strings.Contains(lowerKey, "private_key") || strings.Contains(lowerKey, "token") || strings.Contains(lowerKey, "password") || strings.Contains(lowerKey, "clave") || strings.Contains(lowerKey, "pin") {
			out[key] = "[oculto]"
			continue
		}
		if str, ok := value.(string); ok {
			out[key] = dianTruncate(str, 2000)
			continue
		}
		out[key] = value
	}
	bytes, err := json.Marshal(out)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func upsertDIANTrackHistory(dbEmp *sql.DB, empresaID int64, data map[string]interface{}) {
	if dbEmp == nil || empresaID <= 0 || data == nil {
		return
	}
	trackID := strings.TrimSpace(genericStringValue(data["track_id"]))
	if trackID == "" {
		return
	}
	zipKey := dianFirstNonBlank(genericStringValue(data["zip_key"]), trackID)
	now := dianNowLocal()
	_, _ = dbEmp.Exec(`INSERT INTO empresa_dian_track_historial (
		empresa_id, documento_codigo, tipo_documento, track_id, zip_key, test_set_id, ambiente, endpoint,
		operacion_envio, operacion_acuse, http_status_envio, http_status_acuse, estado_dian, acuse_estado,
		acuse_mensaje, status_code, status_description, is_valid, intento_consulta, respuesta_envio_json,
		respuesta_acuse_json, fecha_envio, fecha_ultimo_acuse, fecha_actualizacion, usuario_creador, observaciones
	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
	ON CONFLICT(empresa_id, track_id) DO UPDATE SET
		documento_codigo=COALESCE(NULLIF(EXCLUDED.documento_codigo,''), empresa_dian_track_historial.documento_codigo),
		tipo_documento=COALESCE(NULLIF(EXCLUDED.tipo_documento,''), empresa_dian_track_historial.tipo_documento),
		zip_key=COALESCE(NULLIF(EXCLUDED.zip_key,''), empresa_dian_track_historial.zip_key),
		test_set_id=COALESCE(NULLIF(EXCLUDED.test_set_id,''), empresa_dian_track_historial.test_set_id),
		ambiente=COALESCE(NULLIF(EXCLUDED.ambiente,''), empresa_dian_track_historial.ambiente),
		endpoint=COALESCE(NULLIF(EXCLUDED.endpoint,''), empresa_dian_track_historial.endpoint),
		operacion_envio=COALESCE(NULLIF(EXCLUDED.operacion_envio,''), empresa_dian_track_historial.operacion_envio),
		operacion_acuse=COALESCE(NULLIF(EXCLUDED.operacion_acuse,''), empresa_dian_track_historial.operacion_acuse),
		http_status_envio=CASE WHEN EXCLUDED.http_status_envio > 0 THEN EXCLUDED.http_status_envio ELSE empresa_dian_track_historial.http_status_envio END,
		http_status_acuse=CASE WHEN EXCLUDED.http_status_acuse > 0 THEN EXCLUDED.http_status_acuse ELSE empresa_dian_track_historial.http_status_acuse END,
		estado_dian=COALESCE(NULLIF(EXCLUDED.estado_dian,''), empresa_dian_track_historial.estado_dian),
		acuse_estado=COALESCE(NULLIF(EXCLUDED.acuse_estado,''), empresa_dian_track_historial.acuse_estado),
		acuse_mensaje=COALESCE(NULLIF(EXCLUDED.acuse_mensaje,''), empresa_dian_track_historial.acuse_mensaje),
		status_code=COALESCE(NULLIF(EXCLUDED.status_code,''), empresa_dian_track_historial.status_code),
		status_description=COALESCE(NULLIF(EXCLUDED.status_description,''), empresa_dian_track_historial.status_description),
		is_valid=COALESCE(NULLIF(EXCLUDED.is_valid,''), empresa_dian_track_historial.is_valid),
		intento_consulta=empresa_dian_track_historial.intento_consulta + EXCLUDED.intento_consulta,
		respuesta_envio_json=COALESCE(NULLIF(EXCLUDED.respuesta_envio_json,''), empresa_dian_track_historial.respuesta_envio_json),
		respuesta_acuse_json=COALESCE(NULLIF(EXCLUDED.respuesta_acuse_json,''), empresa_dian_track_historial.respuesta_acuse_json),
		fecha_envio=COALESCE(NULLIF(EXCLUDED.fecha_envio,''), empresa_dian_track_historial.fecha_envio),
		fecha_ultimo_acuse=COALESCE(NULLIF(EXCLUDED.fecha_ultimo_acuse,''), empresa_dian_track_historial.fecha_ultimo_acuse),
		fecha_actualizacion=EXCLUDED.fecha_actualizacion,
		usuario_creador=COALESCE(NULLIF(EXCLUDED.usuario_creador,''), empresa_dian_track_historial.usuario_creador),
		observaciones=COALESCE(NULLIF(EXCLUDED.observaciones,''), empresa_dian_track_historial.observaciones)`,
		empresaID,
		dianTruncate(genericStringValue(data["documento_codigo"]), 120),
		dianTruncate(genericStringValue(data["tipo_documento"]), 80),
		trackID,
		dianTruncate(zipKey, 160),
		dianTruncate(genericStringValue(data["test_set_id"]), 120),
		dianTruncate(genericStringDefault(data["ambiente"], "habilitacion"), 40),
		dianTruncate(genericStringValue(data["endpoint"]), 300),
		dianTruncate(genericStringValue(data["operacion_envio"]), 80),
		dianTruncate(genericStringValue(data["operacion_acuse"]), 80),
		anyToInt64(data["http_status_envio"]),
		anyToInt64(data["http_status_acuse"]),
		dianTruncate(genericStringValue(data["estado_dian"]), 60),
		dianTruncate(genericStringValue(data["acuse_estado"]), 60),
		dianTruncate(genericStringValue(data["acuse_mensaje"]), 500),
		dianTruncate(genericStringValue(data["status_code"]), 80),
		dianTruncate(genericStringValue(data["status_description"]), 500),
		dianTruncate(genericStringValue(data["is_valid"]), 20),
		anyToInt64(data["intento_consulta"]),
		genericStringValue(data["respuesta_envio_json"]),
		genericStringValue(data["respuesta_acuse_json"]),
		genericStringValue(data["fecha_envio"]),
		genericStringValue(data["fecha_ultimo_acuse"]),
		now,
		dianTruncate(genericStringValue(data["usuario_creador"]), 160),
		dianTruncate(genericStringValue(data["observaciones"]), 500),
	)
}

func listDIANTrackHistory(dbEmp *sql.DB, empresaID int64, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 80
	}
	if limit > 300 {
		limit = 300
	}
	rows, err := dbEmp.Query(`SELECT id, empresa_id, COALESCE(documento_codigo,''), COALESCE(tipo_documento,''),
		track_id, COALESCE(zip_key,''), COALESCE(test_set_id,''), COALESCE(ambiente,''), COALESCE(endpoint,''),
		COALESCE(operacion_envio,''), COALESCE(operacion_acuse,''), COALESCE(http_status_envio,0), COALESCE(http_status_acuse,0),
		COALESCE(estado_dian,''), COALESCE(acuse_estado,''), COALESCE(acuse_mensaje,''), COALESCE(status_code,''),
		COALESCE(status_description,''), COALESCE(is_valid,''), COALESCE(intento_consulta,0),
		COALESCE(fecha_envio,''), COALESCE(fecha_ultimo_acuse,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,'')
		FROM empresa_dian_track_historial
		WHERE empresa_id=? AND COALESCE(estado,'activo') <> 'eliminado'
		ORDER BY id DESC LIMIT ?`, empresaID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, rowEmpresaID, httpEnvio, httpAcuse, intento int64
		var documentoCodigo, tipoDocumento, trackID, zipKey, testSetID, ambiente, endpoint string
		var operacionEnvio, operacionAcuse, estadoDIAN, acuseEstado, acuseMensaje, statusCode, statusDescription, isValid string
		var fechaEnvio, fechaUltimoAcuse, fechaCreacion, fechaActualizacion string
		if err := rows.Scan(&id, &rowEmpresaID, &documentoCodigo, &tipoDocumento, &trackID, &zipKey, &testSetID, &ambiente, &endpoint, &operacionEnvio, &operacionAcuse, &httpEnvio, &httpAcuse, &estadoDIAN, &acuseEstado, &acuseMensaje, &statusCode, &statusDescription, &isValid, &intento, &fechaEnvio, &fechaUltimoAcuse, &fechaCreacion, &fechaActualizacion); err != nil {
			return nil, err
		}
		items = append(items, map[string]interface{}{
			"id":                  id,
			"empresa_id":          rowEmpresaID,
			"documento_codigo":    documentoCodigo,
			"tipo_documento":      tipoDocumento,
			"track_id":            trackID,
			"zip_key":             zipKey,
			"test_set_id":         testSetID,
			"ambiente":            ambiente,
			"endpoint":            endpoint,
			"operacion_envio":     operacionEnvio,
			"operacion_acuse":     operacionAcuse,
			"http_status_envio":   httpEnvio,
			"http_status_acuse":   httpAcuse,
			"estado_dian":         estadoDIAN,
			"acuse_estado":        acuseEstado,
			"acuse_mensaje":       acuseMensaje,
			"status_code":         statusCode,
			"status_description":  statusDescription,
			"is_valid":            isValid,
			"intento_consulta":    intento,
			"fecha_envio":         fechaEnvio,
			"fecha_ultimo_acuse":  fechaUltimoAcuse,
			"fecha_creacion":      fechaCreacion,
			"fecha_actualizacion": fechaActualizacion,
		})
	}
	return items, rows.Err()
}

func sendDIANDocumentoRealSOAP(dbEmp *sql.DB, cfg map[string]interface{}, empresaID int64, payload map[string]interface{}, requestBody map[string]interface{}, endpoint, documentoCodigo, documentoTipo, xmlFirmado, cufe, token, softwareID string, useSharedSoftware bool) (map[string]interface{}, int, error) {
	testSetID := dianFirstNonBlank(genericStringValue(payload["test_set_id"]), genericStringValue(cfg["test_set_id"]))
	ambiente := genericStringDefault(cfg["tipo_ambiente"], "habilitacion")
	operation := dianSOAPOperationFromPayload(payload, ambiente, testSetID)
	if operation == "SendTestSetAsync" && strings.TrimSpace(testSetID) == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("test_set_id es obligatorio para SendTestSetAsync")
	}

	xmlFileName := documentoCodigo + ".xml"
	zipFileName := documentoCodigo + ".zip"
	zipBytes, err := buildDIANZipContent(xmlFileName, xmlFirmado)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("no se pudo crear ZIP DIAN")
	}
	if len(zipBytes) > 50<<20 {
		return nil, http.StatusBadRequest, fmt.Errorf("ZIP DIAN supera el limite de 50 MB")
	}

	soapEndpoint := normalizeDIANSOAPEndpoint(endpoint)
	envelope := buildDIANSOAPEnvelope(operation, soapEndpoint, zipFileName, zipBytes, testSetID)
	wsSecurityMeta := map[string]interface{}{"ws_security": false}
	if isDIANOfficialEndpoint(soapEndpoint) {
		keyRef := dianFirstNonBlank(
			genericStringValue(payload["private_key_pem"]),
			genericStringValue(payload["certificado_clave_ref"]),
			genericStringValue(cfg["certificado_clave_ref"]),
		)
		certRef := dianFirstNonBlank(
			genericStringValue(payload["certificado_pem"]),
			genericStringValue(payload["certificado_ref"]),
			genericStringValue(payload["certificado_x509_ref"]),
			genericStringValue(cfg["certificado_url"]),
		)
		privateKey, keyErr := parseDIANRSAPrivateKey(keyRef)
		certificate, certErr := parseDIANCertificate(certRef)
		if keyErr != nil || certErr != nil {
			missingParts := make([]string, 0)
			if keyErr != nil {
				missingParts = append(missingParts, "llave privada")
			}
			if certErr != nil {
				missingParts = append(missingParts, "certificado X.509")
			}
			return map[string]interface{}{
				"ok":                  false,
				"empresa_id":          empresaID,
				"documento_codigo":    documentoCodigo,
				"cufe":                cufe,
				"transporte":          "soap_dian",
				"operacion_soap":      operation,
				"contingencia_activa": true,
				"estado_dian":         "contingencia",
				"error":               "no se pudo preparar WS-Security DIAN: " + strings.Join(missingParts, ", "),
			}, http.StatusOK, nil
		}
		secureEnvelope, meta, secureErr := buildDIANSOAPEnvelopeWithWSSecurity(operation, soapEndpoint, zipFileName, zipBytes, testSetID, privateKey, certificate, time.Now())
		if secureErr != nil {
			return map[string]interface{}{
				"ok":                  false,
				"empresa_id":          empresaID,
				"documento_codigo":    documentoCodigo,
				"cufe":                cufe,
				"transporte":          "soap_dian",
				"operacion_soap":      operation,
				"contingencia_activa": true,
				"estado_dian":         "contingencia",
				"error":               dianTruncate(secureErr.Error(), 240),
			}, http.StatusOK, nil
		}
		envelope = secureEnvelope
		wsSecurityMeta = meta
	}
	req, err := http.NewRequest(http.MethodPost, soapEndpoint, strings.NewReader(envelope))
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("no se pudo construir request SOAP DIAN")
	}
	action := dianSOAPAction(operation)
	req.Header.Set("Content-Type", `application/soap+xml; charset=utf-8; action="`+action+`"`)
	req.Header.Set("SOAPAction", action)
	req.Header.Set("Accept", "application/soap+xml, text/xml")
	req.Header.Set("User-Agent", "powerfulcontrolsystem-dian-soap/1.0")
	if strings.TrimSpace(token) != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	startedAt := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		_ = updateDIANConfigFields(dbEmp, empresaID, cfg, map[string]interface{}{
			"ultimo_envio": dianNowLocal(),
			"estado_dian":  "contingencia",
			"observaciones": appendStateMachineObservation(
				genericStringValue(cfg["observaciones"]),
				genericStringValue(cfg["estado_dian"]),
				"contingencia",
				dianTruncate(err.Error(), 240),
				"dian_envio_soap",
			),
		})
		return map[string]interface{}{
			"ok":                  false,
			"empresa_id":          empresaID,
			"documento_codigo":    documentoCodigo,
			"cufe":                cufe,
			"transporte":          "soap_dian",
			"operacion_soap":      operation,
			"contingencia_activa": true,
			"estado_dian":         "contingencia",
			"error":               dianTruncate(err.Error(), 240),
			"latency_ms":          time.Since(startedAt).Milliseconds(),
		}, http.StatusOK, nil
	}
	defer resp.Body.Close()

	responseBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	rawResponse := strings.TrimSpace(string(responseBytes))
	responseMap := extractDIANSOAPResponseMap(rawResponse)
	acuseEstado, acuseMensaje := resolveDIANAcuseFromResponse(resp.StatusCode, responseMap)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if parseTruthy(genericStringValue(responseMap["is_valid"])) {
			acuseEstado = "aceptado"
		} else if genericStringValue(responseMap["track_id"]) != "" && acuseEstado == "enviado" {
			acuseEstado = "pendiente"
			acuseMensaje = "TrackId recibido; consultar GetStatusZip para acuse final"
		}
	}

	estadoDIAN := "enviado"
	contingenciaActiva := false
	switch acuseEstado {
	case "aceptado":
		estadoDIAN = "aceptado"
	case "rechazado":
		estadoDIAN = "rechazado"
	case "contingencia":
		estadoDIAN = "contingencia"
		contingenciaActiva = true
	case "pendiente":
		estadoDIAN = "enviado"
	}
	if resp.StatusCode >= 500 {
		estadoDIAN = "contingencia"
		contingenciaActiva = true
	}

	updates := map[string]interface{}{
		"ultimo_envio": dianNowLocal(),
		"estado_dian":  estadoDIAN,
		"observaciones": appendStateMachineObservation(
			genericStringValue(cfg["observaciones"]),
			genericStringValue(cfg["estado_dian"]),
			estadoDIAN,
			dianFirstNonBlank(acuseMensaje, genericStringValue(responseMap["track_id"]), fmt.Sprintf("HTTP %d", resp.StatusCode)),
			"dian_envio_soap",
		),
	}
	if estadoDIAN == "aceptado" {
		updates["consecutivo_actual"] = anyToInt64(cfg["consecutivo_actual"]) + 1
	}
	_ = updateDIANConfigFields(dbEmp, empresaID, cfg, updates)
	upsertDIANTrackHistory(dbEmp, empresaID, map[string]interface{}{
		"documento_codigo":     documentoCodigo,
		"tipo_documento":       documentoTipo,
		"track_id":             genericStringValue(responseMap["track_id"]),
		"zip_key":              genericStringValue(responseMap["track_id"]),
		"test_set_id":          testSetID,
		"ambiente":             ambiente,
		"endpoint":             soapEndpoint,
		"operacion_envio":      operation,
		"http_status_envio":    resp.StatusCode,
		"estado_dian":          estadoDIAN,
		"acuse_estado":         acuseEstado,
		"acuse_mensaje":        acuseMensaje,
		"status_code":          genericStringValue(responseMap["status_code"]),
		"status_description":   genericStringValue(responseMap["status_description"]),
		"is_valid":             genericStringValue(responseMap["is_valid"]),
		"respuesta_envio_json": dianSafeTrackHistoryJSON(responseMap),
		"fecha_envio":          dianNowLocal(),
		"observaciones":        "envio SOAP DIAN registrado por PCS",
	})

	return map[string]interface{}{
		"ok":                  estadoDIAN == "aceptado" || estadoDIAN == "enviado" || acuseEstado == "pendiente",
		"empresa_id":          empresaID,
		"documento_codigo":    documentoCodigo,
		"documento_tipo":      documentoTipo,
		"software_modo":       map[bool]string{true: "compartido", false: "empresa"}[useSharedSoftware],
		"software_id":         softwareID,
		"cufe":                cufe,
		"endpoint":            soapEndpoint,
		"transporte":          "soap_dian",
		"operacion_soap":      operation,
		"zip_file_name":       zipFileName,
		"zip_size_bytes":      len(zipBytes),
		"test_set_id":         testSetID,
		"seguridad_soap":      wsSecurityMeta,
		"http_status":         resp.StatusCode,
		"acuse_estado":        acuseEstado,
		"acuse_mensaje":       acuseMensaje,
		"estado_dian":         estadoDIAN,
		"track_id":            genericStringValue(responseMap["track_id"]),
		"contingencia_activa": contingenciaActiva,
		"latency_ms":          time.Since(startedAt).Milliseconds(),
		"request_resumen":     requestBody,
		"respuesta_dian":      responseMap,
		"raw_response":        rawResponse,
	}, http.StatusOK, nil
}

func consultarDIANStatusZipSOAP(dbEmp *sql.DB, cfg map[string]interface{}, empresaID int64, endpoint, trackID, token string) (map[string]interface{}, int, error) {
	trackID = strings.TrimSpace(trackID)
	if trackID == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("track_id es obligatorio para GetStatusZip")
	}
	soapEndpoint := normalizeDIANSOAPEndpoint(endpoint)
	envelope := buildDIANGetStatusZipEnvelope(soapEndpoint, trackID)
	wsSecurityMeta := map[string]interface{}{"ws_security": false}
	if isDIANOfficialEndpoint(soapEndpoint) {
		keyRef := genericStringValue(cfg["certificado_clave_ref"])
		certRef := genericStringValue(cfg["certificado_url"])
		privateKey, keyErr := parseDIANRSAPrivateKey(keyRef)
		certificate, certErr := parseDIANCertificate(certRef)
		if keyErr != nil || certErr != nil {
			return map[string]interface{}{
				"ok":                  false,
				"empresa_id":          empresaID,
				"track_id":            trackID,
				"endpoint":            soapEndpoint,
				"transporte":          "soap_dian",
				"operacion_soap":      "GetStatusZip",
				"acuse_estado":        "contingencia",
				"estado_dian":         "contingencia",
				"error":               "no se pudo preparar WS-Security DIAN para GetStatusZip",
				"contingencia_activa": true,
			}, http.StatusOK, nil
		}
		secureEnvelope, meta, secureErr := buildDIANGetStatusZipEnvelopeWithWSSecurity(soapEndpoint, trackID, privateKey, certificate, time.Now())
		if secureErr != nil {
			return map[string]interface{}{
				"ok":                  false,
				"empresa_id":          empresaID,
				"track_id":            trackID,
				"endpoint":            soapEndpoint,
				"transporte":          "soap_dian",
				"operacion_soap":      "GetStatusZip",
				"acuse_estado":        "contingencia",
				"estado_dian":         "contingencia",
				"error":               dianTruncate(secureErr.Error(), 240),
				"contingencia_activa": true,
			}, http.StatusOK, nil
		}
		envelope = secureEnvelope
		wsSecurityMeta = meta
	}
	action := dianSOAPAction("GetStatusZip")
	req, err := http.NewRequest(http.MethodPost, soapEndpoint, strings.NewReader(envelope))
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("no se pudo construir GetStatusZip DIAN")
	}
	req.Header.Set("Content-Type", `application/soap+xml; charset=utf-8; action="`+action+`"`)
	req.Header.Set("SOAPAction", action)
	req.Header.Set("Accept", "application/soap+xml, text/xml")
	req.Header.Set("User-Agent", "powerfulcontrolsystem-dian-soap/1.0")
	if strings.TrimSpace(token) != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	startedAt := time.Now()
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return map[string]interface{}{
			"ok":                  false,
			"empresa_id":          empresaID,
			"track_id":            trackID,
			"endpoint":            soapEndpoint,
			"transporte":          "soap_dian",
			"operacion_soap":      "GetStatusZip",
			"acuse_estado":        "contingencia",
			"estado_dian":         "contingencia",
			"error":               dianTruncate(err.Error(), 240),
			"contingencia_activa": true,
			"latency_ms":          time.Since(startedAt).Milliseconds(),
		}, http.StatusOK, nil
	}
	defer resp.Body.Close()

	responseBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	rawResponse := strings.TrimSpace(string(responseBytes))
	responseMap := extractDIANSOAPResponseMap(rawResponse)
	acuseEstado, acuseMensaje := resolveDIANAcuseFromResponse(resp.StatusCode, responseMap)
	if parseTruthy(genericStringValue(responseMap["is_valid"])) {
		acuseEstado = "aceptado"
	}
	estadoDIAN := genericStringDefault(cfg["estado_dian"], "pendiente")
	switch acuseEstado {
	case "aceptado":
		estadoDIAN = "aceptado"
	case "rechazado":
		estadoDIAN = "rechazado"
	case "contingencia":
		estadoDIAN = "contingencia"
	case "enviado", "pendiente":
		estadoDIAN = "enviado"
	}
	_ = updateDIANConfigFields(dbEmp, empresaID, cfg, map[string]interface{}{
		"estado_dian": estadoDIAN,
		"observaciones": appendStateMachineObservation(
			genericStringValue(cfg["observaciones"]),
			genericStringValue(cfg["estado_dian"]),
			estadoDIAN,
			dianFirstNonBlank(acuseMensaje, genericStringValue(responseMap["status_description"]), genericStringValue(responseMap["status_message"]), fmt.Sprintf("HTTP %d", resp.StatusCode)),
			"dian_get_status_zip",
		),
	})
	upsertDIANTrackHistory(dbEmp, empresaID, map[string]interface{}{
		"track_id":             trackID,
		"zip_key":              trackID,
		"test_set_id":          genericStringValue(cfg["test_set_id"]),
		"ambiente":             genericStringDefault(cfg["tipo_ambiente"], "habilitacion"),
		"endpoint":             soapEndpoint,
		"operacion_acuse":      "GetStatusZip",
		"http_status_acuse":    resp.StatusCode,
		"estado_dian":          estadoDIAN,
		"acuse_estado":         acuseEstado,
		"acuse_mensaje":        acuseMensaje,
		"status_code":          genericStringValue(responseMap["status_code"]),
		"status_description":   genericStringValue(responseMap["status_description"]),
		"is_valid":             genericStringValue(responseMap["is_valid"]),
		"intento_consulta":     1,
		"respuesta_acuse_json": dianSafeTrackHistoryJSON(responseMap),
		"fecha_ultimo_acuse":   dianNowLocal(),
		"observaciones":        "acuse GetStatusZip registrado por PCS",
	})

	return map[string]interface{}{
		"ok":             acuseEstado == "aceptado" || acuseEstado == "enviado" || acuseEstado == "pendiente",
		"empresa_id":     empresaID,
		"track_id":       trackID,
		"endpoint":       soapEndpoint,
		"transporte":     "soap_dian",
		"operacion_soap": "GetStatusZip",
		"seguridad_soap": wsSecurityMeta,
		"http_status":    resp.StatusCode,
		"acuse_estado":   acuseEstado,
		"acuse_mensaje":  acuseMensaje,
		"estado_dian":    estadoDIAN,
		"latency_ms":     time.Since(startedAt).Milliseconds(),
		"respuesta_dian": responseMap,
		"raw_response":   rawResponse,
	}, http.StatusOK, nil
}

func consultarDIANSetTrackFinal(dbEmp *sql.DB, cfg map[string]interface{}, empresaID int64, endpoint, trackID, token string, payload map[string]interface{}) map[string]interface{} {
	trackID = strings.TrimSpace(trackID)
	if trackID == "" {
		return nil
	}
	intentos := dianPayloadPositiveInt(payload, 3, "acuse_intentos", "get_status_zip_intentos", "consultas_acuse")
	if intentos <= 0 {
		intentos = 1
	}
	esperaMS := dianPayloadNonNegativeInt(payload, 1200, "acuse_espera_ms", "get_status_zip_espera_ms", "espera_acuse_ms")
	var ultimo map[string]interface{}
	for intento := 0; intento < intentos; intento++ {
		if intento > 0 && esperaMS > 0 {
			time.Sleep(time.Duration(esperaMS) * time.Millisecond)
		}
		resp, _, err := consultarDIANStatusZipSOAP(dbEmp, cfg, empresaID, endpoint, trackID, token)
		if err != nil {
			ultimo = map[string]interface{}{
				"ok":             false,
				"track_id":       trackID,
				"operacion_soap": "GetStatusZip",
				"acuse_estado":   "contingencia",
				"estado_dian":    "contingencia",
				"error":          dianTruncate(err.Error(), 240),
				"intento":        intento + 1,
			}
			continue
		}
		ultimo = resp
		ultimo["intento"] = intento + 1
		estado := strings.ToLower(strings.TrimSpace(genericStringValue(resp["estado_dian"])))
		acuse := strings.ToLower(strings.TrimSpace(genericStringValue(resp["acuse_estado"])))
		if estado == "aceptado" || estado == "rechazado" || estado == "contingencia" || acuse == "aceptado" || acuse == "rechazado" || acuse == "contingencia" {
			break
		}
	}
	return ultimo
}

func buildDIANOfficialReadinessReport(cfg map[string]interface{}, empresaID int64) (map[string]interface{}, int, error) {
	if empresaID <= 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("empresa_id es obligatorio")
	}

	configured := len(cfg) > 0
	missingConfig := make([]string, 0)
	addMissing := func(field string) {
		field = strings.TrimSpace(field)
		if field == "" {
			return
		}
		for _, existing := range missingConfig {
			if existing == field {
				return
			}
		}
		missingConfig = append(missingConfig, field)
	}
	if configured {
		for _, field := range missingDIANFields(cfg) {
			addMissing(field)
		}
		if strings.TrimSpace(genericStringValue(cfg["certificado_clave_ref"])) == "" {
			addMissing("certificado_clave_ref")
		}
		if strings.TrimSpace(genericStringValue(cfg["url_dian"])) == "" {
			addMissing("url_dian")
		}
	}

	technicalGaps := []string{
		"la aceptacion fiscal final depende del acuse DIAN del set real de la empresa",
		"si DIAN/proveedor exige politica XAdES o WS-Security adicional, el acuse lo reportara como rechazo y el sistema lo bloquea/observa",
		"GetStatusZip queda disponible para cerrar acuses asincronos",
	}

	readiness := "sin_configuracion"
	if configured && len(missingConfig) == 0 {
		readiness = "pre_envio_validable"
	}

	return map[string]interface{}{
		"ok":                          true,
		"empresa_id":                  empresaID,
		"configurada":                 configured,
		"estado_preparacion":          readiness,
		"ambiente":                    chooseDIANAmbiente(cfg),
		"faltantes_configuracion":     missingConfig,
		"brechas_tecnicas":            technicalGaps,
		"wsdl_operaciones_objetivo":   []string{"SendBillAsync", "SendBillSync", "SendTestSetAsync", "GetStatusZip", "GetNumberingRange"},
		"wsdl_operaciones_conectadas": []string{"SendBillAsync", "SendBillSync", "SendTestSetAsync", "GetStatusZip"},
		"acciones_pre_envio":          []string{"generar_xml_ubl_base", "firmar_xml_xades_base", "validar_documento_dian", "diagnostico_oficial", "pruebas_dian", "consultar_acuse_real"},
		"siguiente_fase":              "ejecutar set real con certificado y TestSetId de la empresa hasta recibir acuses DIAN aceptados",
	}, http.StatusOK, nil
}

func sendDIANDocumentoReal(dbEmp *sql.DB, cfg map[string]interface{}, empresaID int64, payload map[string]interface{}) (map[string]interface{}, int, error) {
	if empresaID <= 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("empresa_id es obligatorio")
	}
	if len(cfg) == 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("no existe configuracion DIAN para la empresa")
	}

	documentoCodigo := dianFirstNonBlank(genericStringValue(payload["documento_codigo"]), "FV-"+time.Now().Format("20060102150405"))
	xmlFirmado := dianFirstNonBlank(genericStringValue(payload["xml_firmado"]), genericStringValue(payload["xml"]))
	if xmlFirmado == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("xml_firmado o xml es obligatorio para envio real")
	}

	endpoint := normalizeIntegracionEndpoint(dianFirstNonBlank(
		genericStringValue(payload["url_dian"]),
		genericStringValue(payload["endpoint"]),
		genericStringValue(cfg["url_dian"]),
	))
	if endpoint == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("url_dian no configurada o invalida")
	}

	token := strings.TrimSpace(genericStringValue(payload["token"]))
	if token == "" {
		resolved, err := resolveDIANSecretValue(genericStringValue(cfg["token_emisor_ref"]))
		if err == nil {
			token = resolved
		}
	}

	softwareID, softwarePIN, useSharedSoftware, err := resolveDIANSoftwareCredentials(cfg, payload)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	if softwareID == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("software_id no configurado para el modo DIAN actual")
	}
	if softwarePIN == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("software_pin no configurado para el modo DIAN actual")
	}

	fechaEmision := dianFirstNonBlank(genericStringValue(payload["fecha_emision"]), time.Now().Format("2006-01-02T15:04:05-07:00"))
	total := dianFirstNonBlank(genericStringValue(payload["total"]), "0")
	cufe := dianFirstNonBlank(genericStringValue(payload["cufe"]), buildDIANCUFE(genericStringValue(cfg["nit"]), documentoCodigo, fechaEmision, total, softwareID, softwarePIN))
	documentoTipo := genericStringDefault(payload["documento_tipo"], "factura")

	requestBody := map[string]interface{}{
		"empresa_id":       empresaID,
		"documento_codigo": documentoCodigo,
		"documento_tipo":   documentoTipo,
		"fecha_emision":    fechaEmision,
		"total":            total,
		"cufe":             cufe,
		"ambiente":         genericStringDefault(cfg["tipo_ambiente"], "habilitacion"),
		"nit":              genericStringValue(cfg["nit"]),
		"software_id":      softwareID,
		"test_set_id":      dianFirstNonBlank(genericStringValue(payload["test_set_id"]), genericStringValue(cfg["test_set_id"])),
		"xml_firmado":      xmlFirmado,
	}
	for _, key := range []string{"cliente_nombre", "cliente_nit", "impuesto_total", "moneda", "certificado_ref", "certificado_pem", "certificado_clave_ref"} {
		if _, exists := requestBody[key]; !exists && !isEmptyGenericValue(payload[key]) {
			requestBody[key] = payload[key]
		}
	}

	preflight := validateDIANDocumentPreflight(cfg, empresaID, requestBody, xmlFirmado, "envio_real")
	if parseTruthy(genericStringValue(preflight["bloqueado"])) {
		return map[string]interface{}{
			"ok":                   false,
			"empresa_id":           empresaID,
			"documento_codigo":     documentoCodigo,
			"estado_dian":          "bloqueado_validacion",
			"bloqueado":            true,
			"motivo":               "El sistema bloqueo el envio porque fallaron validaciones preventivas DIAN.",
			"validacion_pre_envio": preflight,
		}, http.StatusConflict, nil
	}

	if shouldUseDIANSOAPTransport(endpoint, cfg, payload) {
		return sendDIANDocumentoRealSOAP(dbEmp, cfg, empresaID, payload, requestBody, endpoint, documentoCodigo, documentoTipo, xmlFirmado, cufe, token, softwareID, useSharedSoftware)
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("no se pudo construir request DIAN")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "powerfulcontrolsystem-dian/1.0")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 12 * time.Second}
	startedAt := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		_ = updateDIANConfigFields(dbEmp, empresaID, cfg, map[string]interface{}{
			"ultimo_envio": dianNowLocal(),
			"estado_dian":  "contingencia",
			"observaciones": appendStateMachineObservation(
				genericStringValue(cfg["observaciones"]),
				genericStringValue(cfg["estado_dian"]),
				"contingencia",
				dianTruncate(err.Error(), 240),
				"dian_envio_real",
			),
		})
		return map[string]interface{}{
			"ok":                  false,
			"empresa_id":          empresaID,
			"documento_codigo":    documentoCodigo,
			"cufe":                cufe,
			"contingencia_activa": true,
			"estado_dian":         "contingencia",
			"error":               dianTruncate(err.Error(), 240),
			"latency_ms":          time.Since(startedAt).Milliseconds(),
		}, http.StatusOK, nil
	}
	defer resp.Body.Close()

	responseBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	rawResponse := strings.TrimSpace(string(responseBytes))
	responseMap := extractDIANResponseMap(rawResponse)

	acuseEstado, acuseMensaje := resolveDIANAcuseFromResponse(resp.StatusCode, responseMap)
	if acuseEstado == "" {
		acuseEstado = "pendiente"
	}

	estadoDIAN := "enviado"
	contingenciaActiva := false
	switch acuseEstado {
	case "aceptado":
		estadoDIAN = "aceptado"
	case "rechazado":
		estadoDIAN = "rechazado"
	case "contingencia":
		estadoDIAN = "contingencia"
		contingenciaActiva = true
	default:
		if resp.StatusCode >= 500 {
			estadoDIAN = "contingencia"
			contingenciaActiva = true
		}
	}

	consecutivoActual := anyToInt64(cfg["consecutivo_actual"])
	updates := map[string]interface{}{
		"ultimo_envio": dianNowLocal(),
		"estado_dian":  estadoDIAN,
		"observaciones": appendStateMachineObservation(
			genericStringValue(cfg["observaciones"]),
			genericStringValue(cfg["estado_dian"]),
			estadoDIAN,
			dianFirstNonBlank(acuseMensaje, dianTruncate(rawResponse, 180), fmt.Sprintf("HTTP %d", resp.StatusCode)),
			"dian_envio_real",
		),
	}
	if estadoDIAN == "aceptado" {
		updates["consecutivo_actual"] = consecutivoActual + 1
	}
	_ = updateDIANConfigFields(dbEmp, empresaID, cfg, updates)

	return map[string]interface{}{
		"ok":                  estadoDIAN == "aceptado" || estadoDIAN == "enviado" || acuseEstado == "pendiente",
		"empresa_id":          empresaID,
		"documento_codigo":    documentoCodigo,
		"documento_tipo":      documentoTipo,
		"software_modo":       map[bool]string{true: "compartido", false: "empresa"}[useSharedSoftware],
		"software_id":         softwareID,
		"cufe":                cufe,
		"endpoint":            endpoint,
		"http_status":         resp.StatusCode,
		"acuse_estado":        acuseEstado,
		"acuse_mensaje":       acuseMensaje,
		"estado_dian":         estadoDIAN,
		"contingencia_activa": contingenciaActiva,
		"latency_ms":          time.Since(startedAt).Milliseconds(),
		"respuesta_dian":      responseMap,
		"raw_response":        rawResponse,
	}, http.StatusOK, nil
}

func consultarDIANAcuseReal(dbEmp *sql.DB, cfg map[string]interface{}, empresaID int64, payload map[string]interface{}, r *http.Request) (map[string]interface{}, int, error) {
	if empresaID <= 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("empresa_id es obligatorio")
	}
	if len(cfg) == 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("no existe configuracion DIAN para la empresa")
	}

	documentoCodigo := dianFirstNonBlank(
		genericStringValue(payload["documento_codigo"]),
		strings.TrimSpace(r.URL.Query().Get("documento_codigo")),
		strings.TrimSpace(r.URL.Query().Get("documento")),
	)
	cufe := dianFirstNonBlank(
		genericStringValue(payload["cufe"]),
		strings.TrimSpace(r.URL.Query().Get("cufe")),
	)
	trackID := dianFirstNonBlank(
		genericStringValue(payload["track_id"]),
		genericStringValue(payload["trackId"]),
		strings.TrimSpace(r.URL.Query().Get("track_id")),
		strings.TrimSpace(r.URL.Query().Get("trackId")),
	)
	if documentoCodigo == "" && cufe == "" && trackID == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("documento_codigo, cufe o track_id es obligatorio para consultar acuse")
	}

	endpoint := normalizeIntegracionEndpoint(dianFirstNonBlank(
		genericStringValue(payload["url_acuse"]),
		strings.TrimSpace(r.URL.Query().Get("url_acuse")),
		genericStringValue(payload["url_dian"]),
		genericStringValue(cfg["url_dian"]),
	))
	if endpoint == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("url_dian/url_acuse no configurada o invalida")
	}

	token := strings.TrimSpace(genericStringValue(payload["token"]))
	if token == "" {
		resolved, err := resolveDIANSecretValue(genericStringValue(cfg["token_emisor_ref"]))
		if err == nil {
			token = resolved
		}
	}

	if isDIANOfficialEndpoint(endpoint) && trackID != "" {
		return consultarDIANStatusZipSOAP(dbEmp, cfg, empresaID, endpoint, trackID, token)
	}

	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("endpoint de acuse invalido")
	}
	query := parsedURL.Query()
	query.Set("action", "acuse")
	if documentoCodigo != "" {
		query.Set("documento_codigo", documentoCodigo)
	}
	if cufe != "" {
		query.Set("cufe", cufe)
	}
	parsedURL.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("no se pudo construir consulta de acuse")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "powerfulcontrolsystem-dian/1.0")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	startedAt := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{
			"ok":                  false,
			"empresa_id":          empresaID,
			"documento_codigo":    documentoCodigo,
			"cufe":                cufe,
			"acuse_estado":        "contingencia",
			"estado_dian":         "contingencia",
			"error":               dianTruncate(err.Error(), 240),
			"contingencia_activa": true,
			"latency_ms":          time.Since(startedAt).Milliseconds(),
		}, http.StatusOK, nil
	}
	defer resp.Body.Close()

	responseBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	rawResponse := strings.TrimSpace(string(responseBytes))
	responseMap := extractDIANResponseMap(rawResponse)
	acuseEstado, acuseMensaje := resolveDIANAcuseFromResponse(resp.StatusCode, responseMap)
	if acuseEstado == "" {
		acuseEstado = "pendiente"
	}

	estadoDIAN := genericStringDefault(cfg["estado_dian"], "pendiente")
	switch acuseEstado {
	case "aceptado":
		estadoDIAN = "aceptado"
	case "rechazado":
		estadoDIAN = "rechazado"
	case "contingencia":
		estadoDIAN = "contingencia"
	case "enviado", "pendiente":
		estadoDIAN = "enviado"
	}

	_ = updateDIANConfigFields(dbEmp, empresaID, cfg, map[string]interface{}{
		"estado_dian": estadoDIAN,
		"observaciones": appendStateMachineObservation(
			genericStringValue(cfg["observaciones"]),
			genericStringValue(cfg["estado_dian"]),
			estadoDIAN,
			dianFirstNonBlank(acuseMensaje, dianTruncate(rawResponse, 180), fmt.Sprintf("HTTP %d", resp.StatusCode)),
			"dian_consultar_acuse",
		),
	})

	return map[string]interface{}{
		"ok":                  acuseEstado == "aceptado" || acuseEstado == "enviado" || acuseEstado == "pendiente",
		"empresa_id":          empresaID,
		"documento_codigo":    documentoCodigo,
		"cufe":                cufe,
		"endpoint":            parsedURL.String(),
		"http_status":         resp.StatusCode,
		"acuse_estado":        acuseEstado,
		"acuse_mensaje":       acuseMensaje,
		"estado_dian":         estadoDIAN,
		"contingencia_activa": estadoDIAN == "contingencia",
		"latency_ms":          time.Since(startedAt).Milliseconds(),
		"respuesta_dian":      responseMap,
		"raw_response":        rawResponse,
	}, http.StatusOK, nil
}

func runDIANReconexion(dbEmp *sql.DB, cfg map[string]interface{}, empresaID int64, payload map[string]interface{}) (map[string]interface{}, int, error) {
	if empresaID <= 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("empresa_id es obligatorio")
	}
	if len(cfg) == 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("no existe configuracion DIAN para la empresa")
	}

	endpoint := normalizeIntegracionEndpoint(dianFirstNonBlank(genericStringValue(payload["url_dian"]), genericStringValue(cfg["url_dian"])))
	if endpoint == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("url_dian no configurada o invalida")
	}

	httpStatus, reachable, latencyMS, message := runIntegracionProbe(endpoint)
	estadoAnterior := genericStringDefault(cfg["estado_dian"], "pendiente")
	estadoNuevo := "contingencia"
	if reachable {
		estadoNuevo = "reconectado"
	}

	_ = updateDIANConfigFields(dbEmp, empresaID, cfg, map[string]interface{}{
		"estado_dian": estadoNuevo,
		"observaciones": appendStateMachineObservation(
			genericStringValue(cfg["observaciones"]),
			estadoAnterior,
			estadoNuevo,
			dianFirstNonBlank(message, fmt.Sprintf("HTTP %d", httpStatus)),
			"dian_reconexion",
		),
	})

	response := map[string]interface{}{
		"ok":                  reachable,
		"empresa_id":          empresaID,
		"endpoint":            endpoint,
		"http_status":         httpStatus,
		"reachable":           reachable,
		"latency_ms":          latencyMS,
		"message":             message,
		"estado_anterior":     estadoAnterior,
		"estado_dian":         estadoNuevo,
		"contingencia_activa": !reachable,
	}

	if reachable && parseTruthy(genericStringValue(payload["reenviar"])) {
		envioResp, _, err := sendDIANDocumentoReal(dbEmp, cfg, empresaID, payload)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		response["reenvio"] = envioResp
	}

	return response, http.StatusOK, nil
}

func dianHasMissingField(missing []string, field string) bool {
	field = strings.TrimSpace(field)
	for _, item := range missing {
		if strings.TrimSpace(item) == field {
			return true
		}
	}
	return false
}

func dianPayloadString(payload map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if payload == nil {
			continue
		}
		if value, ok := payload[strings.TrimSpace(key)]; ok {
			if s := strings.TrimSpace(genericStringValue(value)); s != "" {
				return s
			}
		}
	}
	return ""
}

func dianPayloadPositiveInt(payload map[string]interface{}, fallback int, keys ...string) int {
	if fallback < 0 {
		fallback = 0
	}
	for _, key := range keys {
		if payload == nil {
			continue
		}
		value, ok := payload[strings.TrimSpace(key)]
		if !ok {
			continue
		}
		if n := int(anyToInt64(value)); n > 0 {
			return n
		}
		raw := strings.TrimSpace(genericStringValue(value))
		if raw == "" {
			continue
		}
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}

func dianPayloadNonNegativeInt(payload map[string]interface{}, fallback int, keys ...string) int {
	if fallback < 0 {
		fallback = 0
	}
	for _, key := range keys {
		if payload == nil {
			continue
		}
		value, ok := payload[strings.TrimSpace(key)]
		if !ok {
			continue
		}
		if raw := strings.TrimSpace(genericStringValue(value)); raw == "" {
			continue
		}
		if _, isString := value.(string); !isString {
			n := anyToInt64(value)
			if n >= 0 {
				return int(n)
			}
			continue
		}
		raw := strings.TrimSpace(genericStringValue(value))
		if n, err := strconv.Atoi(raw); err == nil && n >= 0 {
			return int(n)
		}
	}
	return fallback
}

func dianBuildDocumentoCodigo(prefijo string, consecutivo int64) string {
	prefijo = strings.TrimSpace(prefijo)
	if prefijo == "" {
		prefijo = "SETP"
	}
	if consecutivo <= 0 {
		consecutivo = 1
	}
	return prefijo + strconv.FormatInt(consecutivo, 10)
}

func dianReceptionMessage(estadoDIAN, acuseEstado, providerMessage string) string {
	estado := strings.ToLower(strings.TrimSpace(estadoDIAN))
	acuse := strings.ToLower(strings.TrimSpace(acuseEstado))
	if msg := strings.TrimSpace(providerMessage); msg != "" {
		return msg
	}
	switch {
	case estado == "aceptado" || acuse == "aceptado":
		return "Documento recibido y aceptado por DIAN."
	case estado == "enviado" || estado == "pendiente" || acuse == "recibido" || acuse == "procesando" || acuse == "pendiente":
		return "Documento recibido por DIAN/proveedor; consulta el acuse final."
	case estado == "rechazado" || acuse == "rechazado":
		return "Documento recibido pero rechazado; revise el detalle del acuse."
	case estado == "contingencia":
		return "Documento en contingencia; debe reprocesarse cuando DIAN/proveedor responda."
	case estado == "simulado" || acuse == "simulado":
		return "Simulacion local correcta; no se envio a DIAN."
	default:
		return "Resultado recibido; revise el estado DIAN y el acuse."
	}
}

func dianAcceptanceRequirementMet(cfg map[string]interface{}, payload map[string]interface{}, resumen map[string]int, aceptadosPorTipo map[string]int, totalDocumentos int) bool {
	accFacturas, accDebito, accCredito, accTotal := dianEffectiveAcceptedCounts(cfg, payload)
	hasConfiguredRequirement := accFacturas > 0 || accDebito > 0 || accCredito > 0 || accTotal > 0
	if !hasConfiguredRequirement {
		return resumen["aceptado"] >= totalDocumentos
	}
	if accTotal > 0 && resumen["aceptado"] < accTotal {
		return false
	}
	if accFacturas > 0 && aceptadosPorTipo["factura"] < accFacturas {
		return false
	}
	if accDebito > 0 && aceptadosPorTipo["nota_debito"] < accDebito {
		return false
	}
	if accCredito > 0 && aceptadosPorTipo["nota_credito"] < accCredito {
		return false
	}
	return true
}

func dianBuildDocumentoXML(cfg map[string]interface{}, documentoCodigo, documentoTipo, issueDate, total string) string {
	root := "Invoice"
	switch strings.ToLower(strings.TrimSpace(documentoTipo)) {
	case "nota_debito", "debit_note", "debitnote", "debito":
		root = "DebitNote"
	case "nota_credito", "credit_note", "creditnote", "credito":
		root = "CreditNote"
	}

	nit := genericStringDefault(cfg["nit"], "000000000")
	razonSocial := escapeXML(genericStringDefault(cfg["razon_social"], "EMPRESA SIN RAZON SOCIAL CONFIGURADA"))
	ambiente := escapeXML(genericStringDefault(cfg["tipo_ambiente"], "habilitacion"))

	issueDateOnly := strings.TrimSpace(issueDate)
	if issueDateOnly == "" {
		issueDateOnly = time.Now().Format("2006-01-02")
	} else if t, err := time.Parse(time.RFC3339, issueDateOnly); err == nil {
		issueDateOnly = t.Format("2006-01-02")
	} else if len(issueDateOnly) >= 10 {
		issueDateOnly = issueDateOnly[:10]
	}

	total = strings.TrimSpace(total)
	if total == "" {
		total = "0"
	}

	return fmt.Sprintf("<?xml version=\"1.0\" encoding=\"UTF-8\"?><%s><ProfileExecutionID>%s</ProfileExecutionID><ID>%s</ID><IssueDate>%s</IssueDate><DocumentCurrencyCode>COP</DocumentCurrencyCode><LegalMonetaryTotal><PayableAmount>%s</PayableAmount></LegalMonetaryTotal><AccountingSupplierParty><PartyTaxScheme><CompanyID>%s</CompanyID><RegistrationName>%s</RegistrationName></PartyTaxScheme></AccountingSupplierParty></%s>",
		root,
		ambiente,
		escapeXML(documentoCodigo),
		escapeXML(issueDateOnly),
		escapeXML(total),
		escapeXML(nit),
		razonSocial,
		root,
	)
}

func runDIANPruebasHabilitacion(dbEmp *sql.DB, cfg map[string]interface{}, empresaID int64, payload map[string]interface{}) (map[string]interface{}, int, error) {
	if empresaID <= 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("empresa_id es obligatorio")
	}
	if len(cfg) == 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("no existe configuracion DIAN para la empresa")
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}

	ambiente := chooseDIANAmbiente(cfg)
	if ambiente == "produccion" {
		return map[string]interface{}{
			"ok":         false,
			"empresa_id": empresaID,
			"bloqueado":  true,
			"motivo":     "Las Pruebas Dian deben ejecutarse en ambiente de habilitacion, no en produccion.",
			"ambiente":   ambiente,
		}, http.StatusConflict, nil
	}

	defaultFacturas, defaultDebito, defaultCredito, defaultTotal := dianConfiguredSetCounts(cfg)
	for key, value := range map[string]interface{}{
		"facturas_electronicas": defaultFacturas,
		"notas_debito":          defaultDebito,
		"notas_credito":         defaultCredito,
		"total_documentos":      defaultTotal,
	} {
		if _, exists := payload[key]; !exists {
			payload[key] = value
		}
	}
	if _, exists := payload["detener_en_error"]; !exists {
		payload["detener_en_error"] = true
	}
	if _, exists := payload["simular"]; !exists {
		payload["simular"] = false
	}

	credenciales, _, credErr := validateDIANCredentialRefs(cfg, empresaID, payload)
	simular := parseTruthy(genericStringValue(payload["simular"]))
	if simular {
		return map[string]interface{}{
			"ok":         false,
			"empresa_id": empresaID,
			"bloqueado":  true,
			"motivo":     "Las pruebas DIAN automaticas deben ejecutarse con envio real; simular=true no esta permitido en este flujo.",
		}, http.StatusBadRequest, nil
	}
	if credErr != nil {
		return nil, http.StatusBadRequest, credErr
	}
	if !simular && !parseTruthy(genericStringValue(credenciales["ok"])) {
		faltantes := credentialMissingList(credenciales)
		motivo := "Faltan datos DIAN antes de ejecutar el set real."
		if len(faltantes) > 0 {
			motivo = "Faltan datos DIAN antes de ejecutar el set real: " + strings.Join(faltantes, ", ") + "."
		}
		return map[string]interface{}{
			"ok":                      false,
			"empresa_id":              empresaID,
			"bloqueado":               true,
			"paso":                    "validar_credenciales",
			"motivo":                  motivo,
			"faltantes":               faltantes,
			"issues":                  credenciales["issues"],
			"validacion_credenciales": credenciales,
			"requisito_set_dian":      dianEffectiveSetRequirementForConfig(cfg, payload),
		}, http.StatusConflict, nil
	}

	result, status, err := runDIANSetPruebasEnvio(dbEmp, cfg, empresaID, payload)
	if result != nil {
		result["accion"] = "pruebas_dian"
		result["validacion_credenciales"] = credenciales
		result["fuente_requisito"] = "DIAN - validar en portal de habilitacion el objetivo exacto del set asignado a la empresa."
	}
	return result, status, err
}

func credentialMissingList(response map[string]interface{}) []string {
	out := make([]string, 0)
	add := func(raw string) {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return
		}
		for _, existing := range out {
			if existing == raw {
				return
			}
		}
		out = append(out, raw)
	}
	if arr, ok := response["faltantes"].([]string); ok {
		for _, item := range arr {
			add(item)
		}
	} else if arr, ok := response["faltantes"].([]interface{}); ok {
		for _, item := range arr {
			add(genericStringValue(item))
		}
	}
	if arr, ok := response["issues"].([]string); ok {
		for _, item := range arr {
			if strings.Contains(strings.ToLower(item), "test_set_id") {
				add("test_set_id")
			}
		}
	} else if arr, ok := response["issues"].([]interface{}); ok {
		for _, item := range arr {
			if strings.Contains(strings.ToLower(genericStringValue(item)), "test_set_id") {
				add("test_set_id")
			}
		}
	}
	return out
}

func runDIANSetPruebasEnvio(dbEmp *sql.DB, cfg map[string]interface{}, empresaID int64, payload map[string]interface{}) (map[string]interface{}, int, error) {
	if empresaID <= 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("empresa_id es obligatorio")
	}
	if len(cfg) == 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("no existe configuracion DIAN para la empresa")
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}

	defaultFacturas, defaultDebito, defaultCredito, _ := dianConfiguredSetCounts(cfg)
	facturas := dianPayloadNonNegativeInt(payload, defaultFacturas, "facturas_electronicas", "facturas", "invoices_total_required")
	notasDebito := dianPayloadNonNegativeInt(payload, defaultDebito, "notas_debito", "debit_notes", "total_debit_notes_required")
	notasCredito := dianPayloadNonNegativeInt(payload, defaultCredito, "notas_credito", "credit_notes", "total_credit_notes_required")

	sumaBase := facturas + notasDebito + notasCredito
	if sumaBase <= 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("define al menos un documento para el set DIAN")
	}
	totalDocumentos := dianPayloadPositiveInt(payload, sumaBase, "total_documentos", "documentos", "total_document_required")
	if totalDocumentos < sumaBase {
		totalDocumentos = sumaBase
	}
	if extra := totalDocumentos - sumaBase; extra > 0 {
		facturas += extra
	}

	maxEnvios := dianPayloadPositiveInt(payload, totalDocumentos, "max_envios", "limit")
	if maxEnvios <= 0 || maxEnvios > totalDocumentos {
		maxEnvios = totalDocumentos
	}

	simular := parseTruthy(dianPayloadString(payload, "simular", "dry_run", "solo_plan"))
	if simular {
		return map[string]interface{}{
			"ok":         false,
			"empresa_id": empresaID,
			"bloqueado":  true,
			"motivo":     "El set DIAN debe enviarse realmente; simular/dry_run/solo_plan no esta permitido.",
		}, http.StatusBadRequest, nil
	}
	detenerEnError := parseTruthy(dianPayloadString(payload, "detener_en_error", "stop_on_error"))
	totalPorDocumento := dianFirstNonBlank(dianPayloadString(payload, "total_por_documento", "total"), "1190.00")
	prefijo := dianFirstNonBlank(dianPayloadString(payload, "prefijo"), genericStringValue(cfg["prefijo"]), "SETP")
	softwareID, _, useSharedSoftware, err := resolveDIANSoftwareCredentials(cfg, payload)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	if strings.TrimSpace(softwareID) == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("software_id no configurado para ejecutar el set DIAN")
	}

	rangoDesde := anyToInt64(cfg["rango_desde"])
	if rangoDesde <= 0 {
		rangoDesde = 1
	}
	rangoHasta := anyToInt64(cfg["rango_hasta"])

	consecutivoInicial := anyToInt64(cfg["consecutivo_actual"])
	if customInicio := anyToInt64(payload["consecutivo_inicial"]); customInicio > 0 {
		consecutivoInicial = customInicio
	}
	if consecutivoInicial <= 0 || consecutivoInicial < rangoDesde {
		consecutivoInicial = rangoDesde
	}

	if rangoHasta > 0 {
		capacidad := int(rangoHasta - consecutivoInicial + 1)
		if capacidad < maxEnvios {
			return nil, http.StatusConflict, fmt.Errorf("rango DIAN insuficiente para el set solicitado: capacidad=%d requerido=%d", capacidad, maxEnvios)
		}
	}

	targets := []struct {
		Tipo     string
		Cantidad int
	}{
		{Tipo: "factura", Cantidad: facturas},
		{Tipo: "nota_debito", Cantidad: notasDebito},
		{Tipo: "nota_credito", Cantidad: notasCredito},
	}

	resumen := map[string]int{
		"aceptado":     0,
		"rechazado":    0,
		"enviado":      0,
		"pendiente":    0,
		"contingencia": 0,
		"error":        0,
		"simulado":     0,
	}
	aceptadosPorTipo := map[string]int{
		"factura":      0,
		"nota_debito":  0,
		"nota_credito": 0,
	}

	procesados := 0
	detenidoPorError := false
	detalles := make([]map[string]interface{}, 0, maxEnvios)
	siguienteConsecutivo := consecutivoInicial
	lastAcceptedInvoiceReference := map[string]string{}

	for _, target := range targets {
		for i := 0; i < target.Cantidad; i++ {
			if procesados >= maxEnvios {
				break
			}

			documentoCodigo := dianBuildDocumentoCodigo(prefijo, siguienteConsecutivo)
			siguienteConsecutivo++
			fechaEmision := time.Now().Format("2006-01-02T15:04:05-07:00")

			detalle := map[string]interface{}{
				"indice":           procesados + 1,
				"tipo_documento":   target.Tipo,
				"documento_codigo": documentoCodigo,
				"fecha_emision":    fechaEmision,
				"total":            totalPorDocumento,
			}

			if simular {
				detalle["ok"] = true
				detalle["estado_dian"] = "simulado"
				detalle["acuse_estado"] = "simulado"
				detalle["software_modo"] = map[bool]string{true: "compartido", false: "empresa"}[useSharedSoftware]
				detalle["software_id"] = softwareID
				resumen["simulado"]++
				detalles = append(detalles, detalle)
				procesados++
				continue
			}

			docPayload := map[string]interface{}{
				"empresa_id":         empresaID,
				"documento_codigo":   documentoCodigo,
				"documento_tipo":     target.Tipo,
				"fecha_emision":      fechaEmision,
				"total":              totalPorDocumento,
				"impuesto_total":     dianFirstNonBlank(dianPayloadString(payload, "impuesto_total"), "190.00"),
				"moneda":             dianFirstNonBlank(dianPayloadString(payload, "moneda"), "COP"),
				"cliente_nombre":     dianFirstNonBlank(dianPayloadString(payload, "cliente_nombre"), "Cliente habilitacion DIAN"),
				"cliente_nit":        dianFirstNonBlank(dianPayloadString(payload, "cliente_nit"), "2222222222"),
				"certificado_ref":    dianPayloadString(payload, "certificado_ref", "certificado_pem", "certificado_x509_ref"),
				"private_key_pem":    dianPayloadString(payload, "private_key_pem"),
				"software_id":        dianPayloadString(payload, "software_id"),
				"software_pin":       dianPayloadString(payload, "software_pin"),
				"test_set_id":        genericStringValue(cfg["test_set_id"]),
				"total_documentos":   totalDocumentos,
				"set_habilitacion":   true,
				"requisito_set_dian": fmt.Sprintf("%d facturas electronicas, %d notas debito y %d notas credito", facturas, notasDebito, notasCredito),
			}
			if target.Tipo != "factura" {
				refCode := dianFirstNonBlank(
					dianPayloadString(payload, "referencia_documento_codigo", "documento_referencia"),
					lastAcceptedInvoiceReference["documento_codigo"],
				)
				refCUFE := dianFirstNonBlank(
					dianPayloadString(payload, "referencia_cufe", "cufe_referencia"),
					lastAcceptedInvoiceReference["cufe"],
				)
				refDate := dianFirstNonBlank(
					dianPayloadString(payload, "referencia_fecha_emision"),
					lastAcceptedInvoiceReference["fecha_emision"],
				)
				if refCode == "" || refCUFE == "" || refDate == "" {
					detalle["ok"] = false
					detalle["estado_dian"] = "bloqueado_referencia"
					detalle["acuse_estado"] = "no_enviado"
					detalle["error"] = "nota requiere factura aceptada por DIAN como referencia; envie primero una factura y espere statusCode 00"
					resumen["error"]++
					detenidoPorError = true
					detalles = append(detalles, detalle)
					procesados++
					break
				}
				if refCode != "" {
					docPayload["referencia_documento_codigo"] = refCode
				}
				if refCUFE != "" {
					docPayload["referencia_cufe"] = refCUFE
				}
				if refDate != "" {
					docPayload["referencia_fecha_emision"] = refDate
				}
			}
			ublResp, _, err := generateDIANUBLBase(cfg, empresaID, docPayload)
			if err != nil {
				detalle["ok"] = false
				detalle["estado_dian"] = "error"
				detalle["error"] = dianTruncate("generar_xml_ubl_base: "+err.Error(), 240)
				resumen["error"]++
				detenidoPorError = detenerEnError
				detalles = append(detalles, detalle)
				procesados++
				if detenidoPorError {
					break
				}
				continue
			}
			docPayload["xml_ubl_base"] = genericStringValue(ublResp["xml_ubl_base"])
			signResp, _, err := signDIANXMLXAdESBase(cfg, empresaID, docPayload)
			if err != nil {
				detalle["ok"] = false
				detalle["estado_dian"] = "error"
				detalle["error"] = dianTruncate("firmar_xml_xades_base: "+err.Error(), 240)
				resumen["error"]++
				detenidoPorError = detenerEnError
				detalles = append(detalles, detalle)
				procesados++
				if detenidoPorError {
					break
				}
				continue
			}
			detalle["digest_documento_base64"] = genericStringValue(signResp["digest_documento_base64"])
			if warnings, ok := signResp["warnings"]; ok {
				detalle["firma_warnings"] = warnings
			}
			preflightPayload := map[string]interface{}{}
			for key, value := range docPayload {
				preflightPayload[key] = value
			}
			preflightPayload["xml_firmado"] = genericStringValue(signResp["xml_firmado"])
			preflight := validateDIANDocumentPreflight(cfg, empresaID, preflightPayload, genericStringValue(signResp["xml_firmado"]), "set_habilitacion")
			detalle["validacion_pre_envio"] = preflight
			if parseTruthy(genericStringValue(preflight["bloqueado"])) {
				detalle["ok"] = false
				detalle["estado_dian"] = "bloqueado_validacion"
				detalle["error"] = "validacion preventiva DIAN no superada"
				resumen["error"]++
				detenidoPorError = detenerEnError
				detalles = append(detalles, detalle)
				procesados++
				if detenidoPorError {
					break
				}
				continue
			}

			envioPayload := map[string]interface{}{
				"empresa_id":       empresaID,
				"documento_codigo": documentoCodigo,
				"documento_tipo":   target.Tipo,
				"xml_firmado":      genericStringValue(signResp["xml_firmado"]),
				"total":            totalPorDocumento,
				"fecha_emision":    fechaEmision,
				"test_set_id":      genericStringValue(cfg["test_set_id"]),
			}
			if parseTruthy(dianPayloadString(payload, "usar_soap_dian", "soap_dian", "transporte_soap")) || isDIANOfficialEndpoint(dianFirstNonBlank(dianPayloadString(payload, "url_dian", "endpoint"), genericStringValue(cfg["url_dian"]))) {
				envioPayload["usar_soap_dian"] = true
			}
			if overrideURL := dianPayloadString(payload, "url_dian", "endpoint"); overrideURL != "" {
				envioPayload["url_dian"] = overrideURL
			}
			if overrideToken := dianPayloadString(payload, "token"); overrideToken != "" {
				envioPayload["token"] = overrideToken
			}
			if overridePIN := dianPayloadString(payload, "software_pin"); overridePIN != "" {
				envioPayload["software_pin"] = overridePIN
			}
			if parseTruthy(genericStringValue(payload["set_habilitacion"])) || genericStringValue(cfg["test_set_id"]) != "" {
				envioPayload["set_habilitacion"] = true
			}

			envioResp, _, err := sendDIANDocumentoReal(dbEmp, cfg, empresaID, envioPayload)
			if err != nil {
				detalle["ok"] = false
				detalle["estado_dian"] = "error"
				detalle["error"] = dianTruncate(err.Error(), 240)
				resumen["error"]++
				detenidoPorError = detenerEnError
				detalles = append(detalles, detalle)
				procesados++
				if detenidoPorError {
					break
				}
				continue
			}

			if trackID := genericStringValue(envioResp["track_id"]); trackID != "" && !parseTruthy(dianPayloadString(payload, "omitir_consulta_acuse", "skip_get_status_zip")) {
				finalResp := consultarDIANSetTrackFinal(
					dbEmp,
					cfg,
					empresaID,
					dianFirstNonBlank(dianPayloadString(payload, "url_dian", "endpoint"), genericStringValue(cfg["url_dian"])),
					trackID,
					dianPayloadString(payload, "token"),
					payload,
				)
				if finalResp != nil {
					envioResp["consulta_acuse"] = finalResp
					for _, key := range []string{"ok", "http_status", "acuse_estado", "acuse_mensaje", "estado_dian", "latency_ms", "respuesta_dian", "raw_response", "seguridad_soap"} {
						if value, exists := finalResp[key]; exists {
							envioResp[key] = value
						}
					}
					envioResp["track_id"] = trackID
				}
			}

			detalle["ok"] = parseTruthy(genericStringValue(envioResp["ok"]))
			detalle["estado_dian"] = genericStringDefault(envioResp["estado_dian"], "pendiente")
			detalle["acuse_estado"] = genericStringDefault(envioResp["acuse_estado"], "pendiente")
			detalle["mensaje_recepcion"] = dianReceptionMessage(
				genericStringValue(detalle["estado_dian"]),
				genericStringValue(detalle["acuse_estado"]),
				genericStringValue(envioResp["acuse_mensaje"]),
			)
			detalle["http_status"] = anyToInt64(envioResp["http_status"])
			detalle["latency_ms"] = anyToInt64(envioResp["latency_ms"])
			for _, debugKey := range []string{
				"endpoint",
				"transporte",
				"operacion_soap",
				"zip_file_name",
				"zip_size_bytes",
				"test_set_id",
				"seguridad_soap",
				"track_id",
				"consulta_acuse",
				"respuesta_dian",
				"raw_response",
			} {
				if value, exists := envioResp[debugKey]; exists {
					detalle[debugKey] = value
				}
			}
			detalle["contingencia_activa"] = parseTruthy(genericStringValue(envioResp["contingencia_activa"]))
			detalle["software_modo"] = genericStringDefault(envioResp["software_modo"], map[bool]string{true: "compartido", false: "empresa"}[useSharedSoftware])
			detalle["software_id"] = dianFirstNonBlank(genericStringValue(envioResp["software_id"]), softwareID)

			estado := strings.ToLower(strings.TrimSpace(genericStringValue(envioResp["estado_dian"])))
			switch estado {
			case "aceptado":
				resumen["aceptado"]++
				aceptadosPorTipo[target.Tipo]++
				if target.Tipo == "factura" {
					lastAcceptedInvoiceReference = map[string]string{
						"documento_codigo": documentoCodigo,
						"cufe":             genericStringValue(ublResp["uuid"]),
						"fecha_emision":    strings.TrimSpace(fechaEmision[:10]),
					}
				}
			case "rechazado":
				resumen["rechazado"]++
			case "contingencia":
				resumen["contingencia"]++
			case "enviado":
				resumen["enviado"]++
			default:
				resumen["pendiente"]++
			}

			detalles = append(detalles, detalle)
			procesados++

			if detenerEnError && (estado == "rechazado" || estado == "contingencia") {
				detenidoPorError = true
				break
			}
		}

		if procesados >= maxEnvios || detenidoPorError {
			break
		}
	}

	if !simular && procesados > 0 {
		okSet := resumen["error"] == 0 && resumen["rechazado"] == 0 && resumen["contingencia"] == 0
		estadoSet := "pruebas_habilitacion_enviadas"
		habilitacionAprobada := dianAcceptanceRequirementMet(cfg, payload, resumen, aceptadosPorTipo, totalDocumentos)
		if habilitacionAprobada {
			estadoSet = "habilitacion_aprobada"
		} else if !okSet {
			estadoSet = "habilitacion_observada"
		}
		_ = updateDIANConfigFields(dbEmp, empresaID, cfg, map[string]interface{}{
			"consecutivo_actual": siguienteConsecutivo,
			"estado_dian":        estadoSet,
			"observaciones": appendStateMachineObservation(
				genericStringValue(cfg["observaciones"]),
				genericStringValue(cfg["estado_dian"]),
				estadoSet,
				fmt.Sprintf("set_pruebas procesado=%d aceptado=%d rechazado=%d contingencia=%d", procesados, resumen["aceptado"], resumen["rechazado"], resumen["contingencia"]),
				"dian_set_pruebas",
			),
		})
	}

	ok := resumen["error"] == 0 && resumen["rechazado"] == 0 && resumen["contingencia"] == 0
	if simular {
		ok = true
	}

	return map[string]interface{}{
		"ok":                     ok,
		"empresa_id":             empresaID,
		"simulado":               simular,
		"requisito_set_dian":     dianEffectiveSetRequirementForConfig(cfg, payload),
		"requisito_configurado":  dianConfiguredSetRequirement(cfg),
		"requisito_base_sistema": dianDefaultSetRequirement(),
		"habilitacion_aprobada":  !simular && dianAcceptanceRequirementMet(cfg, payload, resumen, aceptadosPorTipo, totalDocumentos),
		"software_modo":          map[bool]string{true: "compartido", false: "empresa"}[useSharedSoftware],
		"software_id":            softwareID,
		"test_set_id":            genericStringValue(cfg["test_set_id"]),
		"endpoint":               dianFirstNonBlank(dianPayloadString(payload, "url_dian", "endpoint"), genericStringValue(cfg["url_dian"])),
		"objetivo":               map[string]interface{}{"total_documentos": totalDocumentos, "facturas_electronicas": facturas, "notas_debito": notasDebito, "notas_credito": notasCredito},
		"procesados":             procesados,
		"resumen":                resumen,
		"aceptados_por_tipo":     aceptadosPorTipo,
		"detenido_por_error":     detenidoPorError,
		"consecutivo_inicial":    consecutivoInicial,
		"consecutivo_siguiente":  siguienteConsecutivo,
		"detalles":               detalles,
	}, http.StatusOK, nil
}

func activateDIANProductionLocal(dbEmp *sql.DB, cfg map[string]interface{}, empresaID int64, payload map[string]interface{}) (map[string]interface{}, int, error) {
	if empresaID <= 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("empresa_id es obligatorio")
	}
	if len(cfg) == 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("no existe configuracion DIAN para la empresa")
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}

	estadoActual := strings.ToLower(strings.TrimSpace(genericStringValue(cfg["estado_dian"])))
	confirmadoDIAN := parseTruthy(dianPayloadString(payload, "confirmado_dian", "confirmar_habilitacion_dian", "habilitado_en_dian"))
	if estadoActual != "habilitacion_aprobada" && !confirmadoDIAN {
		return map[string]interface{}{
			"ok":            false,
			"empresa_id":    empresaID,
			"bloqueado":     true,
			"estado_dian":   estadoActual,
			"motivo":        "Antes de activar produccion local debe existir evidencia de set aceptado en DIAN o confirmar manualmente que la empresa quedo habilitada en el portal DIAN.",
			"accion_segura": "Ejecute Pruebas Dian hasta estado aceptado o marque confirmar_habilitacion_dian=true despues de verificarlo en el portal DIAN.",
		}, http.StatusConflict, nil
	}

	urlProduccion := dianFirstNonBlank(
		dianPayloadString(payload, "url_dian_produccion", "url_produccion", "endpoint_produccion"),
		genericStringValue(cfg["url_dian"]),
	)
	if strings.Contains(strings.ToLower(urlProduccion), "hab") && !parseTruthy(dianPayloadString(payload, "permitir_url_habilitacion")) {
		return map[string]interface{}{
			"ok":         false,
			"empresa_id": empresaID,
			"bloqueado":  true,
			"motivo":     "La URL configurada parece de habilitacion. Configure la URL de produccion DIAN/proveedor antes de activar produccion local.",
			"url_dian":   urlProduccion,
		}, http.StatusConflict, nil
	}

	missing := missingDIANFields(cfg)
	if len(missing) > 0 {
		filtered := make([]string, 0, len(missing))
		for _, item := range missing {
			if item == "test_set_id" || item == "url_dian" {
				continue
			}
			filtered = append(filtered, item)
		}
		if strings.TrimSpace(urlProduccion) == "" && dianHasMissingField(missing, "url_dian") {
			filtered = append(filtered, "url_dian_produccion")
		}
		missing = filtered
	}
	if len(missing) > 0 {
		return map[string]interface{}{
			"ok":         false,
			"empresa_id": empresaID,
			"bloqueado":  true,
			"motivo":     "Faltan campos DIAN antes de activar produccion local.",
			"faltantes":  missing,
		}, http.StatusConflict, nil
	}

	updates := map[string]interface{}{
		"tipo_ambiente": "produccion",
		"estado_dian":   "produccion_local_activa",
		"observaciones": appendStateMachineObservation(
			genericStringValue(cfg["observaciones"]),
			genericStringValue(cfg["estado_dian"]),
			"produccion_local_activa",
			"produccion local activada despues de validar habilitacion DIAN",
			"dian_activar_produccion_local",
		),
	}
	if strings.TrimSpace(urlProduccion) != "" {
		updates["url_dian"] = urlProduccion
	}
	if err := updateDIANConfigFields(dbEmp, empresaID, cfg, updates); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("no se pudo activar produccion local")
	}

	return map[string]interface{}{
		"ok":              true,
		"empresa_id":      empresaID,
		"estado_dian":     "produccion_local_activa",
		"tipo_ambiente":   "produccion",
		"url_dian":        urlProduccion,
		"advertencia":     "Este boton cambia la configuracion local del sistema. La habilitacion oficial y el paso a produccion se verifican en la plataforma DIAN.",
		"siguiente_paso":  "Emitir documentos reales solo con resolucion/rangos de produccion asociados y monitorear acuses DIAN.",
		"confirmado_dian": confirmadoDIAN || estadoActual == "habilitacion_aprobada",
	}, http.StatusOK, nil
}

func buildDIANOnboardingGuide(cfg map[string]interface{}, empresaID int64) map[string]interface{} {
	configured := len(cfg) > 0
	missing := []string{}
	if configured {
		missing = missingDIANFields(cfg)
	}

	softwareID, _, useSharedSoftware, softwareErr := resolveDIANSoftwareCredentials(cfg, nil)
	softwareMode := map[bool]string{true: "compartido", false: "empresa"}[useSharedSoftware]
	if !configured {
		softwareMode = "sin_configuracion"
	}

	pasos := []map[string]interface{}{
		{"paso": 1, "titulo": "Registrar empresa DIAN", "detalle": "Guardar NIT, razon social, ambiente, prefijo, resolucion y rango en /api/empresa/facturacion_electronica/dian (CRUD por empresa_id)."},
		{"paso": 2, "titulo": "Definir modelo de software", "detalle": "Activar usar_software_compartido=1 para SaaS o mantener 0 para software propio por empresa."},
		{"paso": 3, "titulo": "Configurar credenciales por empresa", "detalle": "Registrar certificado_clave_ref, certificado X.509, Software ID/PIN y TestSetId por empresa; token_emisor_ref solo aplica si se usa proveedor/API con bearer token."},
		{"paso": 4, "titulo": "Subir firma digital", "detalle": "Usar action=subir_firma (multipart) para adjuntar PEM y guardar referencia segura automaticamente."},
		{"paso": 5, "titulo": "Generar XML base y firma base", "detalle": "Usar action=generar_xml_ubl_base y action=firmar_xml_xades_base para preparar la estructura UBL/firma antes del transporte oficial."},
		{"paso": 6, "titulo": "Validar antes de emitir", "detalle": "Ejecutar action=checklist, action=validar, action=validar_credenciales, action=validar_documento_dian y action=diagnostico_oficial para bloquear errores preventivos antes del envio."},
		{"paso": 7, "titulo": "Probar set de habilitacion", "detalle": "Ejecutar action=pruebas_dian con el objetivo que DIAN muestre para el modo de operacion de la empresa. Todos los documentos deben quedar en estado Aceptado."},
	}

	plantillas := map[string]interface{}{
		"config_base_empresa": map[string]interface{}{
			"empresa_id":               empresaID,
			"nit":                      "900123456",
			"razon_social":             "Empresa Demo SAS",
			"tipo_ambiente":            "habilitacion",
			"prefijo":                  "SETP",
			"resolucion_numero":        "18760000000001",
			"rango_desde":              1,
			"rango_hasta":              999999,
			"consecutivo_actual":       1,
			"url_dian":                 "https://vpfe-hab.dian.gov.co/WcfDianCustomerServices.svc?wsdl",
			"token_emisor_ref":         "opcional si usa proveedor/API bearer: env:DIAN_TOKEN_EMPRESA_XXX",
			"certificado_clave_ref":    "file:/ruta/segura/empresa_xxx_key.pem",
			"usar_software_compartido": 1,
		},
		"activar_software_compartido": map[string]interface{}{
			"empresa_id":                  empresaID,
			"usar_software_compartido":    1,
			"software_id_compartido_ref":  "env:DIAN_SHARED_SOFTWARE_ID",
			"software_pin_compartido_ref": "env:DIAN_SHARED_SOFTWARE_PIN",
		},
		"validar_credenciales": map[string]interface{}{
			"endpoint": "POST /api/empresa/facturacion_electronica/dian?action=validar_credenciales",
			"body":     map[string]interface{}{"empresa_id": empresaID},
		},
		"generar_xml_ubl_base": map[string]interface{}{
			"endpoint": "POST /api/empresa/facturacion_electronica/dian?action=generar_xml_ubl_base",
			"body": map[string]interface{}{
				"empresa_id":       empresaID,
				"documento_codigo": "SETP990000001",
				"documento_tipo":   "factura",
				"cliente_nombre":   "Cliente de habilitacion",
				"cliente_nit":      "222222222222",
				"total":            "1000.00",
			},
		},
		"firmar_xml_xades_base": map[string]interface{}{
			"endpoint": "POST /api/empresa/facturacion_electronica/dian?action=firmar_xml_xades_base",
			"body": map[string]interface{}{
				"empresa_id":       empresaID,
				"documento_codigo": "SETP990000001",
				"xml_ubl_base":     "<Invoice>...</Invoice>",
				"certificado_ref":  "file:/ruta/segura/certificado_empresa.pem",
			},
		},
		"validar_documento_dian": map[string]interface{}{
			"endpoint": "POST /api/empresa/facturacion_electronica/dian?action=validar_documento_dian",
			"body": map[string]interface{}{
				"empresa_id":       empresaID,
				"documento_codigo": "SETP990000001",
				"fecha_emision":    time.Now().Format(time.RFC3339),
				"cliente_nombre":   "Cliente de habilitacion",
				"cliente_nit":      "222222222222",
				"total":            "1000.00",
				"impuesto_total":   "0.00",
				"moneda":           "COP",
				"xml_firmado":      "<Invoice>...</Invoice>",
				"etapa":            "pre_envio",
			},
		},
		"diagnostico_oficial": map[string]interface{}{
			"endpoint": "GET /api/empresa/facturacion_electronica/dian?action=diagnostico_oficial&empresa_id=" + strconv.FormatInt(empresaID, 10),
		},
		"subir_firma": map[string]interface{}{
			"endpoint": "POST /api/empresa/facturacion_electronica/dian?action=subir_firma",
			"multipart_fields": []string{
				"empresa_id",
				"archivo_firma (PEM/P12/PFX con llave privada RSA)",
				"password_firma (opcional)",
			},
		},
		"pruebas_dian": map[string]interface{}{
			"endpoint": "POST /api/empresa/facturacion_electronica/dian?action=pruebas_dian",
			"body": map[string]interface{}{
				"empresa_id":            empresaID,
				"facturas_electronicas": "definir segun set asignado por DIAN",
				"notas_debito":          "definir segun set asignado por DIAN",
				"notas_credito":         "definir segun set asignado por DIAN",
				"total_documentos":      "suma del set asignado",
				"simular":               false,
				"detener_en_error":      true,
			},
		},
	}

	response := map[string]interface{}{
		"ok":                      true,
		"empresa_id":              empresaID,
		"configurada":             configured,
		"faltantes":               missing,
		"software_modo":           softwareMode,
		"software_id_efectivo":    softwareID,
		"pasos":                   pasos,
		"plantillas":              plantillas,
		"recomendacion_operativa": "Modelo SaaS: software compartido + NIT/firma/TestSetId por empresa. El token de emisor solo se exige para proveedor/API bearer; el transporte SOAP/WSDL base esta conectado y el acuse real DIAN confirma la aceptacion fiscal.",
	}
	if softwareErr != nil {
		response["software_error"] = softwareErr.Error()
	}
	return response
}

func dianReferenceSource(raw string) string {
	lower := strings.ToLower(strings.TrimSpace(raw))
	switch {
	case strings.HasPrefix(lower, "env:"):
		return "env"
	case strings.HasPrefix(lower, "file:"):
		return "file"
	case strings.HasPrefix(lower, "base64:"):
		return "base64"
	case lower == "":
		return "vacio"
	default:
		return "inline"
	}
}

func validateDIANCredentialRefs(cfg map[string]interface{}, empresaID int64, payload map[string]interface{}) (map[string]interface{}, int, error) {
	if empresaID <= 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("empresa_id es obligatorio")
	}
	if len(cfg) == 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("no existe configuracion DIAN para la empresa")
	}

	issues := make([]string, 0)
	checks := map[string]interface{}{}

	softwareID, softwarePIN, useSharedSoftware, softwareErr := resolveDIANSoftwareCredentials(cfg, payload)
	if softwareErr != nil {
		issues = append(issues, softwareErr.Error())
	}
	checks["software"] = map[string]interface{}{
		"ok":          softwareErr == nil && strings.TrimSpace(softwareID) != "" && strings.TrimSpace(softwarePIN) != "",
		"modo":        map[bool]string{true: "compartido", false: "empresa"}[useSharedSoftware],
		"software_id": softwareID,
	}

	tokenRequired := dianTokenRequiredForEndpoint(cfg, payload)
	tokenRef := dianFirstNonBlank(genericStringValue(payload["token_emisor_ref"]), genericStringValue(cfg["token_emisor_ref"]))
	tokenPayload := strings.TrimSpace(genericStringValue(payload["token"]))
	tokenOK := false
	tokenMessage := ""
	tokenSource := ""
	if tokenPayload != "" {
		tokenOK = true
		tokenSource = "payload.token"
		tokenMessage = "token entregado en payload"
	} else if tokenRef == "" {
		tokenSource = "vacio"
		if tokenRequired {
			issues = append(issues, "token_emisor_ref no configurado")
			tokenMessage = "faltante"
		} else {
			tokenOK = true
			tokenSource = "no_requerido"
			tokenMessage = "no requerido para endpoint oficial SOAP DIAN; se usa certificado, Software ID/PIN y TestSetId"
		}
	} else {
		tokenSource = dianReferenceSource(tokenRef)
		if !tokenRequired {
			tokenOK = true
			tokenMessage = "configurado, pero no requerido para endpoint oficial SOAP DIAN"
		} else if _, err := resolveDIANSecretValue(tokenRef); err != nil {
			issues = append(issues, "token_emisor_ref invalido")
			tokenMessage = err.Error()
		} else {
			tokenOK = true
			tokenMessage = "resuelto correctamente"
		}
	}
	checks["token_emisor"] = map[string]interface{}{
		"ok":       tokenOK,
		"required": tokenRequired,
		"source":   tokenSource,
		"message":  dianTruncate(tokenMessage, 180),
	}

	ambiente := chooseDIANAmbiente(cfg)
	simular := parseTruthy(genericStringValue(payload["simular"]))
	testSetID := dianFirstNonBlank(genericStringValue(payload["test_set_id"]), genericStringValue(cfg["test_set_id"]))
	testSetRequired := ambiente == "habilitacion"
	testSetOK := strings.TrimSpace(testSetID) != ""
	testSetMessage := "configurado"
	testSetSource := "configuracion"
	if strings.TrimSpace(genericStringValue(payload["test_set_id"])) != "" {
		testSetSource = "payload"
	}
	if !testSetOK {
		testSetSource = "vacio"
		testSetMessage = "faltante para envio real de habilitacion"
		if testSetRequired && !simular {
			issues = append(issues, "test_set_id no configurado")
		}
	}
	checks["test_set_id"] = map[string]interface{}{
		"ok":       testSetOK,
		"required": testSetRequired,
		"source":   testSetSource,
		"message":  dianTruncate(testSetMessage, 180),
	}

	keyRef := dianFirstNonBlank(genericStringValue(payload["certificado_clave_ref"]), genericStringValue(cfg["certificado_clave_ref"]))
	keyOK := false
	keyMessage := ""
	keySource := dianReferenceSource(keyRef)
	if keyRef == "" {
		issues = append(issues, "certificado_clave_ref no configurado")
		keyMessage = "faltante"
	} else if _, err := parseDIANRSAPrivateKey(keyRef); err != nil {
		issues = append(issues, "certificado_clave_ref invalido")
		keyMessage = err.Error()
	} else {
		keyOK = true
		keyMessage = "llave RSA valida"
	}
	checks["firma_digital"] = map[string]interface{}{
		"ok":      keyOK,
		"source":  keySource,
		"message": dianTruncate(keyMessage, 180),
	}

	certRef := dianFirstNonBlank(
		genericStringValue(payload["certificado_url"]),
		genericStringValue(payload["certificado_ref"]),
		genericStringValue(payload["certificado_pem"]),
		genericStringValue(cfg["certificado_url"]),
	)
	certOK := false
	certMessage := ""
	certSource := dianReferenceSource(certRef)
	if certRef == "" {
		issues = append(issues, "certificado_url no configurado")
		certMessage = "faltante"
	} else if cert, err := parseDIANCertificate(certRef); err != nil {
		issues = append(issues, "certificado_url invalido")
		certMessage = err.Error()
	} else {
		certOK = true
		expiry := dianCertificateExpiryStatus(time.Now(), cert.NotBefore, cert.NotAfter, dianCertificateAlertThresholdDays(cfg))
		certMessage = genericStringValue(expiry["mensaje"])
		if parseTruthy(genericStringValue(expiry["vencido"])) {
			certOK = false
			issues = append(issues, "certificado_url vencido")
		} else if parseTruthy(genericStringValue(expiry["proximo_a_vencer"])) {
			checks["certificado_vencimiento"] = expiry
		}
	}
	checks["certificado_x509"] = map[string]interface{}{
		"ok":      certOK,
		"source":  certSource,
		"message": dianTruncate(certMessage, 180),
	}

	ok := len(issues) == 0
	return map[string]interface{}{
		"ok":            ok,
		"empresa_id":    empresaID,
		"software_modo": map[bool]string{true: "compartido", false: "empresa"}[useSharedSoftware],
		"checks":        checks,
		"issues":        issues,
		"faltantes":     missingDIANFields(cfg),
		"recomendaciones": []string{
			"Mantener certificado_clave_ref y certificado X.509 por empresa.",
			"Configurar token_emisor_ref solo cuando el endpoint sea de proveedor/API que use bearer token.",
			"Para endpoint oficial DIAN SOAP/WCF, mantener certificado, Software ID/PIN y TestSetId de habilitacion.",
			"Usar referencias seguras env:/file:/base64: en lugar de secretos inline.",
			"Ejecutar pruebas_dian contra el ambiente de habilitacion DIAN con TestSetId, firma y acuse GetStatusZip final.",
		},
	}, http.StatusOK, nil
}

func resolveEmpresaIDFromMultipartRequest(r *http.Request) (int64, error) {
	if empresaID := parseEmpresaIDFromContext(r); empresaID > 0 {
		return empresaID, nil
	}
	raw := strings.TrimSpace(r.FormValue("empresa_id"))
	if raw == "" {
		raw = strings.TrimSpace(r.URL.Query().Get("empresa_id"))
	}
	if raw == "" {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("empresa_id invalido")
	}
	return id, nil
}

func uploadDIANCompanySignature(dbEmp, dbSuper *sql.DB, r *http.Request) (map[string]interface{}, int, error) {
	if err := r.ParseMultipartForm(12 << 20); err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("payload multipart invalido")
	}

	empresaID, err := resolveEmpresaIDFromMultipartRequest(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	cfg, _ := getEmpresaDIANConfig(dbEmp, empresaID)
	if len(cfg) == 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("configuracion DIAN no existe para la empresa; registre base DIAN primero")
	}

	file, header, err := r.FormFile("archivo_firma")
	if err != nil {
		file, header, err = r.FormFile("firma")
	}
	if err != nil {
		file, header, err = r.FormFile("archivo")
	}
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("archivo_firma es obligatorio")
	}
	defer file.Close()

	contentBytes, err := io.ReadAll(io.LimitReader(file, 10<<20))
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("no se pudo leer la firma")
	}
	if len(contentBytes) == 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("archivo de firma vacio")
	}
	if len(contentBytes) >= 10<<20 {
		return nil, http.StatusBadRequest, fmt.Errorf("archivo de firma demasiado grande; maximo 10 MB")
	}

	password := dianFirstNonBlank(
		r.FormValue("password_firma"),
		r.FormValue("clave_firma"),
		r.FormValue("certificado_password"),
	)
	material, err := decodeDIANSignatureUpload(contentBytes, header.Filename, password)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	dir, empresaFolder := empresaFacturacionFirmaElectronicaDir(dbEmp, empresaID)
	if _, err := ensureEmpresaUploadFolders(dbEmp, empresaID); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("no se pudo preparar directorio de firma")
	}
	_ = os.Chmod(dir, empresaFirmaElectronicaPrivateDirPerms)

	suffix := time.Now().UnixNano()
	keyFileName := fmt.Sprintf("firma_privada_%d.pem", suffix)
	keyAbsPath := filepath.Join(dir, keyFileName)
	if err := os.WriteFile(keyAbsPath, []byte(strings.TrimSpace(material.PrivateKeyPEM)+"\n"), 0o600); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("no se pudo guardar firma en servidor")
	}

	keyRef := "file:" + keyAbsPath
	certRef := ""
	certFileName := ""
	if strings.TrimSpace(material.CertificatePEM) != "" {
		certFileName = fmt.Sprintf("certificado_publico_%d.pem", suffix)
		certAbsPath := filepath.Join(dir, certFileName)
		if err := os.WriteFile(certAbsPath, []byte(strings.TrimSpace(material.CertificatePEM)+"\n"), 0o600); err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("no se pudo guardar certificado en servidor")
		}
		certRef = "file:" + certAbsPath
	}

	estadoActual := genericStringDefault(cfg["estado_dian"], "pendiente")
	uploadTime := time.Now().Format(time.RFC3339)
	keyStatus := "clave no requerida"
	if strings.TrimSpace(password) != "" {
		keyStatus = "clave recibida y usada; no se muestra ni se guarda en claro"
	}
	updates := map[string]interface{}{
		"certificado_clave_ref":        keyRef,
		"certificado_ultima_carga_en":  uploadTime,
		"certificado_archivo_original": strings.TrimSpace(header.Filename),
		"certificado_formato":          material.Format,
		"certificado_subject":          material.Subject,
		"certificado_issuer":           material.Issuer,
		"certificado_serial":           material.Serial,
		"certificado_clave_estado":     keyStatus,
		"observaciones": appendStateMachineObservation(
			genericStringValue(cfg["observaciones"]),
			estadoActual,
			estadoActual,
			"firma digital DIAN actualizada desde carga segura",
			"dian_subir_firma",
		),
	}
	if certRef != "" {
		updates["certificado_url"] = certRef
	}
	if !material.NotAfter.IsZero() {
		updates["certificado_vencimiento"] = material.NotAfter.Format("2006-01-02")
		updates["certificado_vencimiento_en"] = material.NotAfter.Format(time.RFC3339)
		if anyToInt64(cfg["certificado_alerta_dias"]) <= 0 {
			updates["certificado_alerta_dias"] = dianCertificateAlertDays
		}
	}
	if err := updateDIANConfigFields(dbEmp, empresaID, cfg, updates); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("no se pudo actualizar certificado_clave_ref")
	}
	vencimiento := checkDIANCertificateExpiry(dbEmp, dbSuper, empresaID, cfg, true)

	return map[string]interface{}{
		"ok":                          true,
		"empresa_id":                  empresaID,
		"archivo_original":            strings.TrimSpace(header.Filename),
		"archivo_guardado":            keyFileName,
		"certificado_guardado":        certFileName,
		"carpeta_empresa":             empresaFolder,
		"carpeta_firma":               filepath.ToSlash(filepath.Join("uploads", "empresas", empresaFolder, empresaFacturacionElectronicaDirName, empresaFirmaElectronicaDirName)),
		"certificado_clave_ref":       keyRef,
		"certificado_url":             certRef,
		"formato_detectado":           material.Format,
		"certificado_subject":         material.Subject,
		"certificado_issuer":          material.Issuer,
		"certificado_serial":          material.Serial,
		"certificado_ultima_carga_en": uploadTime,
		"certificado_clave_estado":    keyStatus,
		"certificado_vencimiento":     genericStringValue(vencimiento["fecha_vencimiento"]),
		"vencimiento_certificado":     vencimiento,
		"tamano_bytes":                len(contentBytes),
		"siguiente_paso":              "ejecutar action=validar_credenciales y luego action=pruebas_dian",
	}, http.StatusOK, nil
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

func dianResolveOptionalReference(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	resolved, err := resolveDIANSecretValue(raw)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(resolved), nil
}

func resolveDIANSoftwareCredentials(cfg map[string]interface{}, payload map[string]interface{}) (string, string, bool, error) {
	useSharedSoftware := false
	if payload != nil {
		if _, ok := payload["usar_software_compartido"]; ok {
			useSharedSoftware = parseTruthy(genericStringValue(payload["usar_software_compartido"]))
		} else {
			useSharedSoftware = parseTruthy(genericStringValue(cfg["usar_software_compartido"]))
		}
	} else {
		useSharedSoftware = parseTruthy(genericStringValue(cfg["usar_software_compartido"]))
	}

	softwareID, err := dianResolveOptionalReference(genericStringValue(payload["software_id"]))
	if err != nil {
		return "", "", useSharedSoftware, fmt.Errorf("software_id invalido: %w", err)
	}
	softwarePIN, err := dianResolveOptionalReference(genericStringValue(payload["software_pin"]))
	if err != nil {
		return "", "", useSharedSoftware, fmt.Errorf("software_pin invalido: %w", err)
	}

	if useSharedSoftware {
		if softwareID == "" {
			softwareID, err = dianResolveOptionalReference(genericStringValue(cfg["software_id_compartido_ref"]))
			if err != nil {
				return "", "", useSharedSoftware, fmt.Errorf("software_id_compartido_ref invalido: %w", err)
			}
		}
		if softwarePIN == "" {
			softwarePIN, err = dianResolveOptionalReference(genericStringValue(cfg["software_pin_compartido_ref"]))
			if err != nil {
				return "", "", useSharedSoftware, fmt.Errorf("software_pin_compartido_ref invalido: %w", err)
			}
		}
		if softwareID == "" {
			softwareID = strings.TrimSpace(os.Getenv("DIAN_SHARED_SOFTWARE_ID"))
		}
		if softwarePIN == "" {
			softwarePIN = strings.TrimSpace(os.Getenv("DIAN_SHARED_SOFTWARE_PIN"))
		}
	}

	if softwareID == "" {
		softwareID, err = dianResolveOptionalReference(genericStringValue(cfg["software_id"]))
		if err != nil {
			return "", "", useSharedSoftware, fmt.Errorf("software_id de empresa invalido: %w", err)
		}
	}
	if softwarePIN == "" {
		softwarePIN, err = dianResolveOptionalReference(genericStringValue(cfg["software_pin"]))
		if err != nil {
			return "", "", useSharedSoftware, fmt.Errorf("software_pin de empresa invalido: %w", err)
		}
	}

	return softwareID, softwarePIN, useSharedSoftware, nil
}

func missingDIANFields(cfg map[string]interface{}) []string {
	required := []string{
		"nit",
		"digito_verificacion",
		"razon_social",
		"tipo_ambiente",
		"prefijo",
		"resolucion_numero",
		"resolucion_fecha_desde",
		"resolucion_fecha_hasta",
		"rango_desde",
		"rango_hasta",
		"consecutivo_actual",
		"llave_tecnica",
		"url_dian",
		"certificado_clave_ref",
		"certificado_url",
	}
	missing := make([]string, 0)
	addMissing := func(field string) {
		field = strings.TrimSpace(field)
		if field == "" {
			return
		}
		for _, existing := range missing {
			if existing == field {
				return
			}
		}
		missing = append(missing, field)
	}
	for _, field := range required {
		if isEmptyGenericValue(cfg[field]) {
			addMissing(field)
		}
	}
	if dianTokenRequiredForEndpoint(cfg, nil) && isEmptyGenericValue(cfg["token_emisor_ref"]) {
		addMissing("token_emisor_ref")
	}
	ambiente := chooseDIANAmbiente(cfg)
	if ambiente == "habilitacion" && isEmptyGenericValue(cfg["test_set_id"]) {
		addMissing("test_set_id")
	}
	rangoDesde := anyToInt64(cfg["rango_desde"])
	rangoHasta := anyToInt64(cfg["rango_hasta"])
	consecutivo := anyToInt64(cfg["consecutivo_actual"])
	if rangoDesde > 0 && rangoHasta > 0 && rangoDesde > rangoHasta {
		addMissing("rango_numeracion_invalido")
	}
	if consecutivo > 0 && rangoDesde > 0 && consecutivo < rangoDesde {
		addMissing("consecutivo_actual_fuera_de_rango")
	}
	if consecutivo > 0 && rangoHasta > 0 && consecutivo > rangoHasta {
		addMissing("consecutivo_actual_fuera_de_rango")
	}

	softwareID, softwarePIN, useSharedSoftware, err := resolveDIANSoftwareCredentials(cfg, nil)
	if err != nil {
		addMissing("software_configuracion_invalida")
	}
	if strings.TrimSpace(softwareID) == "" {
		if useSharedSoftware {
			addMissing("software_id_compartido_ref|DIAN_SHARED_SOFTWARE_ID|software_id")
		} else {
			addMissing("software_id")
		}
	}
	if strings.TrimSpace(softwarePIN) == "" {
		if useSharedSoftware {
			addMissing("software_pin_compartido_ref|DIAN_SHARED_SOFTWARE_PIN|software_pin")
		} else {
			addMissing("software_pin")
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
		{"paso": 2, "titulo": "Definir modelo de software", "detalle": "Elegir software compartido (SaaS) o software por empresa; configurar Software ID/PIN segun el modelo."},
		{"paso": 3, "titulo": "Solicitar numeracion", "detalle": "Solicitar prefijo, resolucion y rango autorizado en la DIAN."},
		{"paso": 4, "titulo": "Cargar configuracion por empresa", "detalle": "Configurar NIT, razon social, ambiente, certificado/token por empresa y parametros de software (compartido o propio)."},
		{"paso": 5, "titulo": "Ejecutar Pruebas Dian", "detalle": "Enviar el set configurado en la plataforma DIAN para el modo de operacion de la empresa hasta que todos los documentos queden en estado Aceptado."},
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
