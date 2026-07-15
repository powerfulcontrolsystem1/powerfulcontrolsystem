package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// EmpresaGenericListFilter define filtros de listado para tablas genericas por empresa.
type EmpresaGenericListFilter struct {
	IncludeInactive bool
	Q               string
	Limit           int
	Offset          int
	SearchColumns   []string
}

var empresaGenericAllowedTables = map[string]struct{}{
	"empresa_cotizaciones_venta":          {},
	"empresa_pedidos_venta":               {},
	"empresa_devoluciones_venta":          {},
	"empresa_compras_documentos":          {},
	"empresa_facturacion_documentos":      {},
	"empresa_plan_cuentas":                {},
	"empresa_cuentas_por_cobrar":          {},
	"empresa_cuentas_por_pagar":           {},
	"inventario_lotes_series":             {},
	"inventario_lotes_series_movimientos": {},
	"empresa_devoluciones_proveedor":      {},
	"empresa_rrhh_vacaciones_licencias":   {},
	"crm_leads":                           {},
	"crm_interacciones":                   {},
	"crm_campanas":                        {},
	"produccion_bom":                      {},
	"produccion_bom_detalle":              {},
	"produccion_ordenes":                  {},
	"logistica_transportistas":            {},
	"logistica_rutas":                     {},
	"logistica_envios":                    {},
	"empresa_documentos_gestion":          {},
	"empresa_documentos_firmas":           {},
	"empresa_integraciones_apis":          {},
	"empresa_integraciones_bancos":        {},
	"empresa_dian_configuracion":          {},
}

// EnsureEmpresaModulosFaltantesSchema crea tablas base para los modulos ERP faltantes.
func EnsureEmpresaModulosFaltantesSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	if dbConn == nil {
		return errors.New("db connection is nil")
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_cotizaciones_venta (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			cliente_id INTEGER DEFAULT 0,
			cliente_nombre TEXT,
			fecha_documento TEXT DEFAULT (CURRENT_DATE),
			vigencia_hasta TEXT,
			estado_documento TEXT DEFAULT 'borrador',
			subtotal REAL DEFAULT 0,
			descuento_total REAL DEFAULT 0,
			impuesto_total REAL DEFAULT 0,
			total REAL DEFAULT 0,
			moneda TEXT DEFAULT 'COP',
			notas TEXT,
			origen TEXT,
			convertido_pedido_id INTEGER DEFAULT 0,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_cotizaciones_empresa_estado ON empresa_cotizaciones_venta(empresa_id, estado_documento, fecha_documento DESC);`,

		`CREATE TABLE IF NOT EXISTS empresa_pedidos_venta (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			cliente_id INTEGER DEFAULT 0,
			cliente_nombre TEXT,
			cotizacion_id INTEGER DEFAULT 0,
			fecha_pedido TEXT DEFAULT (CURRENT_DATE),
			fecha_entrega_estimada TEXT,
			estado_pedido TEXT DEFAULT 'borrador',
			subtotal REAL DEFAULT 0,
			descuento_total REAL DEFAULT 0,
			impuesto_total REAL DEFAULT 0,
			total REAL DEFAULT 0,
			moneda TEXT DEFAULT 'COP',
			notas TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_pedidos_empresa_estado ON empresa_pedidos_venta(empresa_id, estado_pedido, fecha_pedido DESC);`,

		`CREATE TABLE IF NOT EXISTS empresa_devoluciones_venta (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			carrito_id INTEGER DEFAULT 0,
			documento_referencia TEXT,
			motivo TEXT,
			fecha_devolucion TEXT DEFAULT (CURRENT_DATE),
			estado_devolucion TEXT DEFAULT 'borrador',
			subtotal REAL DEFAULT 0,
			impuesto_total REAL DEFAULT 0,
			total REAL DEFAULT 0,
			moneda TEXT DEFAULT 'COP',
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_devoluciones_venta_empresa_fecha ON empresa_devoluciones_venta(empresa_id, fecha_devolucion DESC);`,

		`CREATE TABLE IF NOT EXISTS empresa_plan_cuentas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			tipo_cuenta TEXT DEFAULT 'activo',
			naturaleza TEXT DEFAULT 'debito',
			nivel INTEGER DEFAULT 1,
			cuenta_padre_codigo TEXT,
			admite_movimiento INTEGER DEFAULT 1,
			aplica_impuesto INTEGER DEFAULT 0,
			plantilla_tipo_empresa TEXT DEFAULT 'general',
			plantilla_codigo TEXT,
			plantilla_version TEXT DEFAULT '1',
			cuenta_clave TEXT,
			requerida INTEGER DEFAULT 0,
			orden_plantilla INTEGER DEFAULT 0,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_plan_cuentas_empresa_tipo ON empresa_plan_cuentas(empresa_id, tipo_cuenta, codigo);`,

		`CREATE TABLE IF NOT EXISTS empresa_cuentas_por_cobrar (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			cliente_id INTEGER DEFAULT 0,
			cliente_nombre TEXT,
			documento_tipo TEXT,
			documento_codigo TEXT,
			fecha_emision TEXT,
			fecha_vencimiento TEXT,
			dias_mora INTEGER DEFAULT 0,
			valor_original REAL DEFAULT 0,
			valor_pagado REAL DEFAULT 0,
			saldo REAL DEFAULT 0,
			estado_cartera TEXT DEFAULT 'pendiente',
			moneda TEXT DEFAULT 'COP',
			periodo_contable TEXT,
			referencia_pagos_json TEXT,
			fecha_ultimo_pago TEXT,
			conciliado_en TEXT,
			conciliado_por TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_cxc_empresa_estado ON empresa_cuentas_por_cobrar(empresa_id, estado_cartera, fecha_vencimiento);`,

		`CREATE TABLE IF NOT EXISTS empresa_cuentas_por_pagar (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			proveedor_id INTEGER DEFAULT 0,
			proveedor_nombre TEXT,
			documento_tipo TEXT,
			documento_codigo TEXT,
			fecha_emision TEXT,
			fecha_vencimiento TEXT,
			dias_mora INTEGER DEFAULT 0,
			valor_original REAL DEFAULT 0,
			valor_pagado REAL DEFAULT 0,
			saldo REAL DEFAULT 0,
			estado_cartera TEXT DEFAULT 'pendiente',
			moneda TEXT DEFAULT 'COP',
			periodo_contable TEXT,
			referencia_pagos_json TEXT,
			fecha_ultimo_pago TEXT,
			conciliado_en TEXT,
			conciliado_por TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_cxp_empresa_estado ON empresa_cuentas_por_pagar(empresa_id, estado_cartera, fecha_vencimiento);`,

		`CREATE TABLE IF NOT EXISTS inventario_lotes_series (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			producto_id INTEGER NOT NULL,
			bodega_id INTEGER DEFAULT 0,
			tipo_control TEXT DEFAULT 'lote',
			codigo_lote_serie TEXT NOT NULL,
			fecha_fabricacion TEXT,
			fecha_vencimiento TEXT,
			cantidad_inicial REAL DEFAULT 0,
			cantidad_disponible REAL DEFAULT 0,
			reservado_cantidad REAL DEFAULT 0,
			vendido_cantidad REAL DEFAULT 0,
			costo_unitario REAL DEFAULT 0,
			estado_lote TEXT DEFAULT 'activo',
			bloqueado_venta INTEGER DEFAULT 0,
			bloqueo_motivo TEXT,
			ultima_operacion_tipo TEXT,
			ultima_operacion_ref TEXT,
			ultima_operacion_en TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, producto_id, bodega_id, codigo_lote_serie)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_lotes_empresa_vencimiento ON inventario_lotes_series(empresa_id, fecha_vencimiento, producto_id);`,

		`CREATE TABLE IF NOT EXISTS inventario_lotes_series_movimientos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			lote_serie_id INTEGER NOT NULL,
			producto_id INTEGER NOT NULL,
			bodega_id INTEGER DEFAULT 0,
			codigo_lote_serie TEXT NOT NULL,
			tipo_operacion TEXT NOT NULL,
			cantidad REAL DEFAULT 0,
			saldo_lote REAL DEFAULT 0,
			referencia_tipo TEXT,
			referencia_codigo TEXT,
			cliente_id INTEGER DEFAULT 0,
			cliente_nombre TEXT,
			detalle_json TEXT,
			fecha_operacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_lotes_mov_empresa_lote_fecha ON inventario_lotes_series_movimientos(empresa_id, lote_serie_id, fecha_operacion DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_lotes_mov_empresa_referencia ON inventario_lotes_series_movimientos(empresa_id, referencia_tipo, referencia_codigo);`,

		`CREATE TABLE IF NOT EXISTS empresa_devoluciones_proveedor (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			proveedor_id INTEGER DEFAULT 0,
			proveedor_nombre TEXT,
			documento_compra_codigo TEXT,
			fecha_devolucion TEXT DEFAULT (CURRENT_DATE),
			motivo TEXT,
			estado_devolucion TEXT DEFAULT 'borrador',
			subtotal REAL DEFAULT 0,
			impuesto_total REAL DEFAULT 0,
			total REAL DEFAULT 0,
			moneda TEXT DEFAULT 'COP',
			periodo_contable TEXT,
			impacto_contable_movimiento_id INTEGER DEFAULT 0,
			impacto_contable_evento_id INTEGER DEFAULT 0,
			fecha_contabilizacion TEXT,
			contabilizado_por TEXT,
			total_reintegrado REAL DEFAULT 0,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_dev_prov_empresa_fecha ON empresa_devoluciones_proveedor(empresa_id, fecha_devolucion DESC);`,

		`CREATE TABLE IF NOT EXISTS empresa_rrhh_vacaciones_licencias (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			empleado_id INTEGER DEFAULT 0,
			empleado_nomina_id INTEGER DEFAULT 0,
			empleado_nombre TEXT,
			tipo_novedad TEXT DEFAULT 'vacacion',
			fecha_inicio TEXT,
			fecha_fin TEXT,
			dias REAL DEFAULT 0,
			remunerada INTEGER DEFAULT 1,
			estado_novedad TEXT DEFAULT 'solicitada',
			soporte_url TEXT,
			aprobado_por TEXT,
			nivel_aprobacion_actual INTEGER DEFAULT 0,
			nivel_aprobacion_requerido INTEGER DEFAULT 1,
			aprobadores_json TEXT,
			historial_aprobaciones_json TEXT,
			fecha_aprobacion_final TEXT,
			periodo_acumulado_desde TEXT,
			periodo_acumulado_hasta TEXT,
			saldo_dias_antes REAL DEFAULT 0,
			saldo_dias_despues REAL DEFAULT 0,
			saldo_snapshot_json TEXT,
			nomina_liquidacion_id INTEGER DEFAULT 0,
			nomina_periodo_desde TEXT,
			nomina_periodo_hasta TEXT,
			nomina_vinculada_en TEXT,
			nomina_vinculada_por TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_rrhh_vac_lic_empresa_fechas ON empresa_rrhh_vacaciones_licencias(empresa_id, fecha_inicio, fecha_fin);`,

		`CREATE TABLE IF NOT EXISTS crm_leads (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT,
			empresa_origen TEXT,
			email TEXT,
			telefono TEXT,
			canal_origen TEXT,
			estado_lead TEXT DEFAULT 'nuevo',
			valor_potencial REAL DEFAULT 0,
			probabilidad REAL DEFAULT 0,
			propietario TEXT,
			proximo_contacto TEXT,
			notas TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_crm_leads_empresa_estado ON crm_leads(empresa_id, estado_lead, proximo_contacto);`,

		`CREATE TABLE IF NOT EXISTS crm_interacciones (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			lead_id INTEGER DEFAULT 0,
			cliente_id INTEGER DEFAULT 0,
			tipo_interaccion TEXT DEFAULT 'llamada',
			fecha_interaccion TEXT DEFAULT (CURRENT_TIMESTAMP),
			resumen TEXT,
			resultado TEXT,
			usuario_responsable TEXT,
			proxima_accion TEXT,
			estado_interaccion TEXT DEFAULT 'abierta',
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_crm_interacciones_empresa_fecha ON crm_interacciones(empresa_id, fecha_interaccion DESC);`,

		`CREATE TABLE IF NOT EXISTS crm_campanas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT,
			canal TEXT,
			objetivo TEXT,
			presupuesto REAL DEFAULT 0,
			fecha_inicio TEXT,
			fecha_fin TEXT,
			estado_campana TEXT DEFAULT 'planificada',
			audiencia TEXT,
			kpi_objetivo TEXT,
			resultado_json TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_crm_campanas_empresa_estado ON crm_campanas(empresa_id, estado_campana, fecha_inicio);`,

		`CREATE TABLE IF NOT EXISTS produccion_bom (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			producto_id INTEGER DEFAULT 0,
			producto_nombre TEXT,
			version TEXT DEFAULT '1.0',
			rendimiento REAL DEFAULT 1,
			unidad_medida TEXT,
			costo_estimado_total REAL DEFAULT 0,
			estado_bom TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo, version)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_bom_empresa_producto ON produccion_bom(empresa_id, producto_id, version);`,

		`CREATE TABLE IF NOT EXISTS produccion_bom_detalle (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			bom_id INTEGER NOT NULL,
			insumo_producto_id INTEGER DEFAULT 0,
			insumo_nombre TEXT,
			cantidad REAL DEFAULT 0,
			unidad_medida TEXT,
			costo_unitario REAL DEFAULT 0,
			costo_total REAL DEFAULT 0,
			merma_porcentaje REAL DEFAULT 0,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_bom_detalle_empresa_bom ON produccion_bom_detalle(empresa_id, bom_id);`,

		`CREATE TABLE IF NOT EXISTS produccion_ordenes (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			bom_id INTEGER DEFAULT 0,
			producto_id INTEGER DEFAULT 0,
			producto_nombre TEXT,
			cantidad_programada REAL DEFAULT 0,
			cantidad_producida REAL DEFAULT 0,
			fecha_programada TEXT,
			fecha_inicio TEXT,
			fecha_fin TEXT,
			estado_orden TEXT DEFAULT 'planificada',
			costo_estimado REAL DEFAULT 0,
			costo_real REAL DEFAULT 0,
			responsable TEXT,
			notas TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_produccion_ordenes_empresa_estado ON produccion_ordenes(empresa_id, estado_orden, fecha_programada);`,

		`CREATE TABLE IF NOT EXISTS logistica_transportistas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT,
			documento TEXT,
			telefono TEXT,
			email TEXT,
			placa TEXT,
			vehiculo_tipo TEXT,
			capacidad_carga REAL DEFAULT 0,
			estado_transportista TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_log_transportistas_empresa_nombre ON logistica_transportistas(empresa_id, nombre);`,

		`CREATE TABLE IF NOT EXISTS logistica_rutas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT,
			origen TEXT,
			destino TEXT,
			distancia_km REAL DEFAULT 0,
			tiempo_estimado_min REAL DEFAULT 0,
			estado_ruta TEXT DEFAULT 'activa',
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_log_rutas_empresa_origen_destino ON logistica_rutas(empresa_id, origen, destino);`,

		`CREATE TABLE IF NOT EXISTS logistica_envios (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			cliente_id INTEGER DEFAULT 0,
			cliente_nombre TEXT,
			documento_referencia TEXT,
			direccion_entrega TEXT,
			ruta_id INTEGER DEFAULT 0,
			transportista_id INTEGER DEFAULT 0,
			fecha_programada TEXT,
			fecha_salida TEXT,
			fecha_entrega TEXT,
			estado_envio TEXT DEFAULT 'programado',
			costo_envio REAL DEFAULT 0,
			latitud REAL DEFAULT 0,
			longitud REAL DEFAULT 0,
			observaciones_seguimiento TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_log_envios_empresa_estado ON logistica_envios(empresa_id, estado_envio, fecha_programada);`,

		`CREATE TABLE IF NOT EXISTS empresa_documentos_gestion (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			modulo TEXT,
			entidad TEXT,
			entidad_id INTEGER DEFAULT 0,
			documento_codigo TEXT,
			nombre_documento TEXT,
			tipo_documento TEXT,
			mime_type TEXT,
			url_archivo TEXT,
			hash_archivo TEXT,
			tamano_bytes INTEGER DEFAULT 0,
			version TEXT DEFAULT '1',
			estado_documento TEXT DEFAULT 'vigente',
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_doc_gestion_empresa_modulo ON empresa_documentos_gestion(empresa_id, modulo, entidad_id);`,

		`CREATE TABLE IF NOT EXISTS empresa_documentos_firmas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			documento_gestion_id INTEGER DEFAULT 0,
			tipo_firma TEXT DEFAULT 'digital',
			firmante_nombre TEXT,
			firmante_documento TEXT,
			firmante_email TEXT,
			certificado_serial TEXT,
			algoritmo_firma TEXT DEFAULT 'SHA256',
			hash_firma TEXT,
			fecha_firma TEXT,
			validez_hasta TEXT,
			estado_firma TEXT DEFAULT 'pendiente',
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_doc_firmas_empresa_doc ON empresa_documentos_firmas(empresa_id, documento_gestion_id, fecha_firma);`,

		`CREATE TABLE IF NOT EXISTS empresa_integraciones_apis (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			nombre_integracion TEXT,
			tipo_integracion TEXT,
			base_url TEXT,
			auth_tipo TEXT,
			api_key_ref TEXT,
			estado_integracion TEXT DEFAULT 'inactiva',
			ultima_sincronizacion TEXT,
			respuesta_ultimo_sync TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_integraciones_api_empresa_estado ON empresa_integraciones_apis(empresa_id, estado_integracion);`,

		`CREATE TABLE IF NOT EXISTS empresa_integraciones_bancos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			banco_nombre TEXT,
			tipo_conexion TEXT,
			numero_cuenta TEXT,
			titular TEXT,
			moneda TEXT DEFAULT 'COP',
			api_endpoint TEXT,
			credencial_ref TEXT,
			estado_integracion TEXT DEFAULT 'inactiva',
			ultima_conciliacion TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_integraciones_bancos_empresa_estado ON empresa_integraciones_bancos(empresa_id, estado_integracion);`,

		`CREATE TABLE IF NOT EXISTS empresa_dian_configuracion (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			nit TEXT,
			digito_verificacion TEXT,
			razon_social TEXT,
			tipo_ambiente TEXT DEFAULT 'habilitacion',
			software_id TEXT,
			software_pin TEXT,
			usar_software_compartido INTEGER DEFAULT 0,
			software_id_compartido_ref TEXT,
			software_pin_compartido_ref TEXT,
			modo_operacion_descripcion TEXT,
			modo_operacion_fecha_inicio TEXT,
			modo_operacion_fecha_termino TEXT,
			test_set_id TEXT,
			certificado_url TEXT,
			certificado_clave_ref TEXT,
			certificado_vencimiento TEXT,
			certificado_vencimiento_en TEXT,
			certificado_alerta_dias INTEGER DEFAULT 30,
			certificado_alerta_ultimo_envio TEXT,
			certificado_alerta_email TEXT,
			certificado_ultima_carga_en TEXT,
			certificado_archivo_original TEXT,
			certificado_formato TEXT,
			certificado_subject TEXT,
			certificado_issuer TEXT,
			certificado_serial TEXT,
			certificado_clave_estado TEXT,
			resolucion_alerta_dias INTEGER DEFAULT 30,
			resolucion_alerta_ultimo_envio TEXT,
			prefijo TEXT,
			resolucion_numero TEXT,
			resolucion_fecha_desde TEXT,
			resolucion_fecha_hasta TEXT,
			rango_desde INTEGER DEFAULT 0,
			rango_hasta INTEGER DEFAULT 0,
			consecutivo_actual INTEGER DEFAULT 0,
			llave_tecnica TEXT,
			set_documentos_requeridos INTEGER DEFAULT 0,
			set_facturas_requeridas INTEGER DEFAULT 0,
			set_notas_debito_requeridas INTEGER DEFAULT 0,
			set_notas_credito_requeridas INTEGER DEFAULT 0,
			set_documentos_aceptados_requeridos INTEGER DEFAULT 0,
			set_facturas_aceptadas_requeridas INTEGER DEFAULT 0,
			set_notas_debito_aceptadas_requeridas INTEGER DEFAULT 0,
			set_notas_credito_aceptadas_requeridas INTEGER DEFAULT 0,
			url_dian TEXT,
			token_emisor_ref TEXT,
			ultimo_envio TEXT,
			estado_dian TEXT DEFAULT 'pendiente',
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id),
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_dian_empresa_ambiente ON empresa_dian_configuracion(empresa_id, tipo_ambiente, estado_dian);`,
		`CREATE TABLE IF NOT EXISTS empresa_dian_track_historial (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			documento_codigo TEXT,
			tipo_documento TEXT,
			track_id TEXT NOT NULL,
			zip_key TEXT,
			test_set_id TEXT,
			ambiente TEXT DEFAULT 'habilitacion',
			endpoint TEXT,
			operacion_envio TEXT,
			operacion_acuse TEXT,
			http_status_envio INTEGER DEFAULT 0,
			http_status_acuse INTEGER DEFAULT 0,
			estado_dian TEXT DEFAULT 'pendiente',
			acuse_estado TEXT DEFAULT 'pendiente',
			acuse_mensaje TEXT,
			status_code TEXT,
			status_description TEXT,
			is_valid TEXT,
			intento_consulta INTEGER DEFAULT 0,
			respuesta_envio_json TEXT,
			respuesta_acuse_json TEXT,
			fecha_envio TEXT,
			fecha_ultimo_acuse TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, track_id)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_dian_track_empresa_estado ON empresa_dian_track_historial(empresa_id, acuse_estado, estado_dian, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_dian_track_empresa_documento ON empresa_dian_track_historial(empresa_id, documento_codigo, tipo_documento);`,
	}

	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_plan_cuentas", "plantilla_tipo_empresa", "TEXT DEFAULT 'general'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_plan_cuentas", "plantilla_codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_plan_cuentas", "plantilla_version", "TEXT DEFAULT '1'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_plan_cuentas", "cuenta_clave", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_plan_cuentas", "requerida", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_plan_cuentas", "orden_plantilla", "INTEGER DEFAULT 0"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_cuentas_por_cobrar", "periodo_contable", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cuentas_por_cobrar", "referencia_pagos_json", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cuentas_por_cobrar", "fecha_ultimo_pago", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cuentas_por_cobrar", "conciliado_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cuentas_por_cobrar", "conciliado_por", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_cuentas_por_pagar", "periodo_contable", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cuentas_por_pagar", "referencia_pagos_json", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cuentas_por_pagar", "fecha_ultimo_pago", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cuentas_por_pagar", "conciliado_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cuentas_por_pagar", "conciliado_por", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "inventario_lotes_series", "reservado_cantidad", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "inventario_lotes_series", "vendido_cantidad", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "inventario_lotes_series", "bloqueado_venta", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "inventario_lotes_series", "bloqueo_motivo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "inventario_lotes_series", "ultima_operacion_tipo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "inventario_lotes_series", "ultima_operacion_ref", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "inventario_lotes_series", "ultima_operacion_en", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_devoluciones_proveedor", "periodo_contable", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_devoluciones_proveedor", "impacto_contable_movimiento_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_devoluciones_proveedor", "impacto_contable_evento_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_devoluciones_proveedor", "fecha_contabilizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_devoluciones_proveedor", "contabilizado_por", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_devoluciones_proveedor", "total_reintegrado", "REAL DEFAULT 0"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_rrhh_vacaciones_licencias", "empleado_nomina_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_rrhh_vacaciones_licencias", "nivel_aprobacion_actual", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_rrhh_vacaciones_licencias", "nivel_aprobacion_requerido", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_rrhh_vacaciones_licencias", "aprobadores_json", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_rrhh_vacaciones_licencias", "historial_aprobaciones_json", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_rrhh_vacaciones_licencias", "fecha_aprobacion_final", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_rrhh_vacaciones_licencias", "periodo_acumulado_desde", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_rrhh_vacaciones_licencias", "periodo_acumulado_hasta", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_rrhh_vacaciones_licencias", "saldo_dias_antes", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_rrhh_vacaciones_licencias", "saldo_dias_despues", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_rrhh_vacaciones_licencias", "saldo_snapshot_json", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_rrhh_vacaciones_licencias", "nomina_liquidacion_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_rrhh_vacaciones_licencias", "nomina_periodo_desde", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_rrhh_vacaciones_licencias", "nomina_periodo_hasta", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_rrhh_vacaciones_licencias", "nomina_vinculada_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_rrhh_vacaciones_licencias", "nomina_vinculada_por", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "usar_software_compartido", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "software_id_compartido_ref", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "software_pin_compartido_ref", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "certificado_vencimiento", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "certificado_vencimiento_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "certificado_alerta_dias", "INTEGER DEFAULT 30"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "certificado_alerta_ultimo_envio", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "certificado_alerta_email", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "certificado_ultima_carga_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "certificado_archivo_original", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "certificado_formato", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "certificado_subject", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "certificado_issuer", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "certificado_serial", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "certificado_clave_estado", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "resolucion_alerta_dias", "INTEGER DEFAULT 30"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "resolucion_alerta_ultimo_envio", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "llave_tecnica", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "modo_operacion_descripcion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "modo_operacion_fecha_inicio", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "modo_operacion_fecha_termino", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "set_documentos_requeridos", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "set_facturas_requeridas", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "set_notas_debito_requeridas", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "set_notas_credito_requeridas", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "set_documentos_aceptados_requeridos", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "set_facturas_aceptadas_requeridas", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "set_notas_debito_aceptadas_requeridas", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_dian_configuracion", "set_notas_credito_aceptadas_requeridas", "INTEGER DEFAULT 0"); err != nil {
		return err
	}

	if _, err := dbConn.Exec(`CREATE INDEX IF NOT EXISTS ix_cxc_empresa_periodo ON empresa_cuentas_por_cobrar(empresa_id, periodo_contable, estado_cartera);`); err != nil {
		return err
	}
	if _, err := dbConn.Exec(`CREATE INDEX IF NOT EXISTS ix_cxp_empresa_periodo ON empresa_cuentas_por_pagar(empresa_id, periodo_contable, estado_cartera);`); err != nil {
		return err
	}
	if _, err := dbConn.Exec(`CREATE INDEX IF NOT EXISTS ix_lotes_empresa_estado_bloqueo ON inventario_lotes_series(empresa_id, estado_lote, bloqueado_venta);`); err != nil {
		return err
	}
	if _, err := dbConn.Exec(`CREATE INDEX IF NOT EXISTS ix_dev_prov_empresa_estado_contable ON empresa_devoluciones_proveedor(empresa_id, estado_devolucion, periodo_contable);`); err != nil {
		return err
	}
	if _, err := dbConn.Exec(`CREATE INDEX IF NOT EXISTS ix_rrhh_vac_lic_empresa_estado_nivel ON empresa_rrhh_vacaciones_licencias(empresa_id, estado_novedad, nivel_aprobacion_actual, nivel_aprobacion_requerido, id DESC);`); err != nil {
		return err
	}
	if _, err := dbConn.Exec(`CREATE INDEX IF NOT EXISTS ix_rrhh_vac_lic_empresa_nomina_periodo ON empresa_rrhh_vacaciones_licencias(empresa_id, empleado_nomina_id, nomina_liquidacion_id, nomina_periodo_desde, nomina_periodo_hasta);`); err != nil {
		return err
	}
	if _, err := dbConn.Exec(`CREATE INDEX IF NOT EXISTS ix_dian_empresa_shared_mode ON empresa_dian_configuracion(empresa_id, usar_software_compartido);`); err != nil {
		return err
	}

	return nil
}

// VerifyEmpresaModulosFaltantesSchema confirms that the migration role has
// provisioned the ERP extension tables. HTTP handlers must verify readiness,
// never create or alter tables while serving a tenant request.
func VerifyEmpresaModulosFaltantesSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	var exists int
	err := queryRowSQLCompat(dbConn, `SELECT 1 FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = ? LIMIT 1`, "empresa_cotizaciones_venta").Scan(&exists)
	if err == sql.ErrNoRows {
		return errors.New("esquema ERP no migrado; ejecute pcs-migrate")
	}
	return err
}

func isAllowedGenericTable(table string) bool {
	table = strings.TrimSpace(table)
	if table == "" || !isSafeSQLIdentifier(table) {
		return false
	}
	_, ok := empresaGenericAllowedTables[table]
	return ok
}

func isSafeSQLIdentifier(v string) bool {
	v = strings.TrimSpace(v)
	if v == "" {
		return false
	}
	for _, ch := range v {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' {
			continue
		}
		return false
	}
	return true
}

func normalizeGenericLimitOffset(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func sanitizeGenericColumns(columns []string) []string {
	out := make([]string, 0, len(columns))
	for _, column := range columns {
		trimmed := strings.TrimSpace(column)
		if trimmed == "" || !isSafeSQLIdentifier(trimmed) {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func genericLikePattern(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, "!", "!!")
	raw = strings.ReplaceAll(raw, "%", "!%")
	raw = strings.ReplaceAll(raw, "_", "!_")
	return "%" + raw + "%"
}

func rowsToMapSlice(rows *sql.Rows) ([]map[string]interface{}, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		item := make(map[string]interface{}, len(cols))
		for i, col := range cols {
			v := values[i]
			switch vv := v.(type) {
			case []byte:
				item[col] = string(vv)
			default:
				item[col] = vv
			}
		}
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func normalizeGenericValue(v interface{}) interface{} {
	switch vv := v.(type) {
	case bool:
		if vv {
			return 1
		}
		return 0
	default:
		return vv
	}
}

func isEmpresaCarteraTable(table string) bool {
	switch strings.TrimSpace(table) {
	case "empresa_cuentas_por_cobrar", "empresa_cuentas_por_pagar":
		return true
	default:
		return false
	}
}

func floatFromGenericValue(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int32:
		return float64(t)
	case int64:
		return float64(t)
	case string:
		f, _ := strconv.ParseFloat(strings.TrimSpace(t), 64)
		return f
	default:
		return 0
	}
}

func firstNonBlankString(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func carteraPeriodoFromPayload(payload map[string]interface{}) string {
	if payload == nil {
		return ""
	}
	periodo := normalizePeriodoContable(fmt.Sprintf("%v", payload["periodo_contable"]))
	if periodo != "" {
		return periodo
	}
	periodo = normalizePeriodoContable(fmt.Sprintf("%v", payload["fecha_emision"]))
	if periodo != "" {
		return periodo
	}
	return ""
}

func carteraEstadoFromValues(saldo, valorPagado float64, fechaVencimiento string) string {
	if saldo <= 0.009 {
		return "pagada"
	}
	if valorPagado > 0 {
		return "parcial"
	}
	if dueDate := normalizeDateOnly(fechaVencimiento); dueDate != "" {
		if parsed, err := time.Parse("2006-01-02", dueDate); err == nil {
			nowDate := time.Now().In(time.Local).Format("2006-01-02")
			nowParsed, _ := time.Parse("2006-01-02", nowDate)
			if parsed.Before(nowParsed) {
				return "vencida"
			}
		}
	}
	return "pendiente"
}

func carteraDiasMora(fechaVencimiento string, saldo float64) int64 {
	if saldo <= 0.009 {
		return 0
	}
	dueDate := normalizeDateOnly(fechaVencimiento)
	if dueDate == "" {
		return 0
	}
	parsed, err := time.Parse("2006-01-02", dueDate)
	if err != nil {
		return 0
	}
	nowDate := time.Now().In(time.Local).Format("2006-01-02")
	nowParsed, err := time.Parse("2006-01-02", nowDate)
	if err != nil || !nowParsed.After(parsed) {
		return 0
	}
	return int64(nowParsed.Sub(parsed).Hours() / 24)
}

func normalizeCarteraPayloadValues(payload map[string]interface{}, current map[string]interface{}) {
	if payload == nil {
		return
	}

	valorOriginal := floatFromGenericValue(payload["valor_original"])
	valorPagado := floatFromGenericValue(payload["valor_pagado"])
	saldo := floatFromGenericValue(payload["saldo"])
	fechaVencimiento := strings.TrimSpace(fmt.Sprintf("%v", payload["fecha_vencimiento"]))

	if current != nil {
		if _, ok := payload["valor_original"]; !ok {
			valorOriginal = floatFromGenericValue(current["valor_original"])
		}
		if _, ok := payload["valor_pagado"]; !ok {
			valorPagado = floatFromGenericValue(current["valor_pagado"])
		}
		if _, ok := payload["saldo"]; !ok {
			saldo = floatFromGenericValue(current["saldo"])
		}
		if strings.TrimSpace(fechaVencimiento) == "" {
			fechaVencimiento = strings.TrimSpace(fmt.Sprintf("%v", current["fecha_vencimiento"]))
		}
	}

	if valorOriginal < 0 {
		valorOriginal = 0
	}
	if valorPagado < 0 {
		valorPagado = 0
	}

	if valorOriginal <= 0 && (saldo > 0 || valorPagado > 0) {
		valorOriginal = saldo + valorPagado
	}
	if valorOriginal > 0 && valorPagado > valorOriginal {
		valorPagado = valorOriginal
	}

	if saldo < 0 || (current == nil && saldo == 0 && valorOriginal > 0 && valorPagado >= 0) {
		saldo = valorOriginal - valorPagado
	}
	if saldo < 0 {
		saldo = 0
	}

	periodo := carteraPeriodoFromPayload(payload)
	if periodo == "" && current != nil {
		periodo = normalizePeriodoContable(firstNonBlankString(fmt.Sprintf("%v", current["periodo_contable"]), fmt.Sprintf("%v", current["fecha_emision"])))
	}
	if periodo == "" {
		periodo = time.Now().In(time.Local).Format("2006-01")
	}

	if strings.TrimSpace(fmt.Sprintf("%v", payload["moneda"])) == "" {
		if current == nil || strings.TrimSpace(fmt.Sprintf("%v", current["moneda"])) == "" {
			payload["moneda"] = "COP"
		}
	}

	payload["valor_original"] = valorOriginal
	payload["valor_pagado"] = valorPagado
	payload["saldo"] = saldo
	payload["periodo_contable"] = periodo
	payload["dias_mora"] = carteraDiasMora(fechaVencimiento, saldo)

	estadoPayload := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", payload["estado_cartera"])))
	if estadoPayload == "" {
		estadoPayload = carteraEstadoFromValues(saldo, valorPagado, fechaVencimiento)
	}
	payload["estado_cartera"] = estadoPayload
}

func carteraPeriodoByID(dbConn *sql.DB, table string, empresaID, id int64) (string, error) {
	var periodo string
	var fechaEmision string
	err := dbConn.QueryRow(`SELECT COALESCE(periodo_contable, ''), COALESCE(fecha_emision, '') FROM `+table+` WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, id).Scan(&periodo, &fechaEmision)
	if err != nil {
		return "", err
	}
	periodo = normalizePeriodoContable(periodo)
	if periodo == "" {
		periodo = normalizePeriodoContable(fechaEmision)
	}
	return periodo, nil
}

func ensureCarteraPeriodoEditable(dbConn *sql.DB, empresaID int64, periodo string) error {
	periodo = normalizePeriodoContable(periodo)
	if periodo == "" {
		return nil
	}
	cerrado, err := IsEmpresaFinanzasPeriodoCerrado(dbConn, empresaID, periodo)
	if err != nil {
		return err
	}
	if cerrado {
		return ErrPeriodoFinancieroCerrado
	}
	return nil
}

// ListEmpresaGenericRows lista registros de una tabla generica por empresa.
func ListEmpresaGenericRows(dbConn *sql.DB, table string, empresaID int64, filter EmpresaGenericListFilter) ([]map[string]interface{}, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, errors.New("empresa_id invalido")
	}
	if !isAllowedGenericTable(table) {
		return nil, fmt.Errorf("tabla no permitida: %s", table)
	}

	limit, offset := normalizeGenericLimitOffset(filter.Limit, filter.Offset)
	searchCols := sanitizeGenericColumns(filter.SearchColumns)

	// #nosec G202 -- SQL structure is assembled only from server-side allowlists; all external values remain bound parameters.
	query := "SELECT * FROM " + table + " WHERE empresa_id = ?"
	args := []interface{}{empresaID}

	if !filter.IncludeInactive {
		query += " AND LOWER(COALESCE(estado, 'activo')) = 'activo'"
	}

	q := strings.TrimSpace(filter.Q)
	if q != "" && len(searchCols) > 0 {
		pattern := genericLikePattern(q)
		clauses := make([]string, 0, len(searchCols))
		for _, col := range searchCols {
			clauses = append(clauses, "LOWER(COALESCE("+col+", '')) LIKE LOWER(?) ESCAPE '!'")
			args = append(args, pattern)
		}
		query += " AND (" + strings.Join(clauses, " OR ") + ")"
	}

	query += " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rowsToMapSlice(rows)
}

// GetEmpresaGenericRowByID obtiene un registro de tabla generica por empresa e id.
func GetEmpresaGenericRowByID(dbConn *sql.DB, table string, empresaID, id int64) (map[string]interface{}, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 || id <= 0 {
		return nil, errors.New("empresa_id o id invalido")
	}
	if !isAllowedGenericTable(table) {
		return nil, fmt.Errorf("tabla no permitida: %s", table)
	}

	// #nosec G202 -- SQL structure is assembled only from server-side allowlists; all external values remain bound parameters.
	rows, err := dbConn.Query("SELECT * FROM "+table+" WHERE empresa_id = ? AND id = ? LIMIT 1", empresaID, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items, err := rowsToMapSlice(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, sql.ErrNoRows
	}

	return items[0], nil
}

// CreateEmpresaGenericRow inserta un registro en una tabla generica por empresa.
func CreateEmpresaGenericRow(dbConn *sql.DB, table string, empresaID int64, payload map[string]interface{}, allowedColumns []string) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return 0, errors.New("empresa_id invalido")
	}
	if !isAllowedGenericTable(table) {
		return 0, fmt.Errorf("tabla no permitida: %s", table)
	}
	if isEmpresaCarteraTable(table) {
		normalizeCarteraPayloadValues(payload, nil)
		if err := ensureCarteraPeriodoEditable(dbConn, empresaID, carteraPeriodoFromPayload(payload)); err != nil {
			return 0, err
		}
	}

	allowed := sanitizeGenericColumns(allowedColumns)
	columns := []string{"empresa_id"}
	values := []interface{}{empresaID}

	for _, col := range allowed {
		if col == "id" || col == "empresa_id" {
			continue
		}
		if v, ok := payload[col]; ok {
			columns = append(columns, col)
			values = append(values, normalizeGenericValue(v))
		}
	}

	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	query := "INSERT INTO " + table + " (" + strings.Join(columns, ", ") + ") VALUES (" + strings.Join(placeholders, ", ") + ")"
	return insertSQLCompat(dbConn, query, values...)
}

// UpdateEmpresaGenericRow actualiza columnas permitidas de un registro por empresa e id.
func UpdateEmpresaGenericRow(dbConn *sql.DB, table string, empresaID, id int64, payload map[string]interface{}, allowedColumns []string) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if empresaID <= 0 || id <= 0 {
		return errors.New("empresa_id o id invalido")
	}
	if !isAllowedGenericTable(table) {
		return fmt.Errorf("tabla no permitida: %s", table)
	}

	if isEmpresaCarteraTable(table) {
		current, err := GetEmpresaGenericRowByID(dbConn, table, empresaID, id)
		if err != nil {
			return err
		}
		periodoActual := normalizePeriodoContable(firstNonBlankString(fmt.Sprintf("%v", current["periodo_contable"]), fmt.Sprintf("%v", current["fecha_emision"])))
		if err := ensureCarteraPeriodoEditable(dbConn, empresaID, periodoActual); err != nil {
			return err
		}
		normalizeCarteraPayloadValues(payload, current)
		periodoNuevo := carteraPeriodoFromPayload(payload)
		if periodoNuevo != "" && periodoNuevo != periodoActual {
			if err := ensureCarteraPeriodoEditable(dbConn, empresaID, periodoNuevo); err != nil {
				return err
			}
		}
	}

	allowed := sanitizeGenericColumns(allowedColumns)
	setParts := make([]string, 0, len(allowed)+1)
	args := make([]interface{}, 0, len(allowed)+2)

	for _, col := range allowed {
		if col == "id" || col == "empresa_id" {
			continue
		}
		if v, ok := payload[col]; ok {
			setParts = append(setParts, col+" = ?")
			args = append(args, normalizeGenericValue(v))
		}
	}

	if len(setParts) == 0 {
		return errors.New("no hay campos para actualizar")
	}

	setParts = append(setParts, "fecha_actualizacion = CURRENT_TIMESTAMP")
	// #nosec G202 -- SQL structure is assembled only from server-side allowlists; all external values remain bound parameters.
	query := "UPDATE " + table + " SET " + strings.Join(setParts, ", ") + " WHERE empresa_id = ? AND id = ?"
	args = append(args, empresaID, id)

	_, err := dbConn.Exec(query, args...)
	return err
}

// SetEmpresaGenericRowEstado ajusta el estado activo/inactivo de un registro.
func SetEmpresaGenericRowEstado(dbConn *sql.DB, table string, empresaID, id int64, estado string) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if empresaID <= 0 || id <= 0 {
		return errors.New("empresa_id o id invalido")
	}
	if !isAllowedGenericTable(table) {
		return fmt.Errorf("tabla no permitida: %s", table)
	}
	if isEmpresaCarteraTable(table) {
		periodo, err := carteraPeriodoByID(dbConn, table, empresaID, id)
		if err != nil {
			return err
		}
		if err := ensureCarteraPeriodoEditable(dbConn, empresaID, periodo); err != nil {
			return err
		}
	}

	estado = strings.ToLower(strings.TrimSpace(estado))
	if estado == "" {
		estado = "activo"
	}

	// #nosec G202 -- SQL structure is assembled only from server-side allowlists; all external values remain bound parameters.
	_, err := dbConn.Exec("UPDATE "+table+" SET estado = ?, fecha_actualizacion = CURRENT_TIMESTAMP WHERE empresa_id = ? AND id = ?", estado, empresaID, id)
	return err
}

// DeleteEmpresaGenericRow aplica eliminacion logica por empresa e id.
func DeleteEmpresaGenericRow(dbConn *sql.DB, table string, empresaID, id int64) error {
	return SetEmpresaGenericRowEstado(dbConn, table, empresaID, id, "inactivo")
}
