package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
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
	"empresa_cotizaciones_venta":        {},
	"empresa_pedidos_venta":             {},
	"empresa_devoluciones_venta":        {},
	"empresa_plan_cuentas":              {},
	"empresa_cuentas_por_cobrar":        {},
	"empresa_cuentas_por_pagar":         {},
	"inventario_lotes_series":           {},
	"empresa_devoluciones_proveedor":    {},
	"empresa_rrhh_vacaciones_licencias": {},
	"crm_leads":                         {},
	"crm_interacciones":                 {},
	"crm_campanas":                      {},
	"produccion_bom":                    {},
	"produccion_bom_detalle":            {},
	"produccion_ordenes":                {},
	"logistica_transportistas":          {},
	"logistica_rutas":                   {},
	"logistica_envios":                  {},
	"empresa_documentos_gestion":        {},
	"empresa_documentos_firmas":         {},
	"empresa_integraciones_apis":        {},
	"empresa_integraciones_bancos":      {},
	"empresa_dian_configuracion":        {},
}

// EnsureEmpresaModulosFaltantesSchema crea tablas base para los modulos ERP faltantes.
func EnsureEmpresaModulosFaltantesSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_cotizaciones_venta (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			cliente_id INTEGER DEFAULT 0,
			cliente_nombre TEXT,
			fecha_documento TEXT DEFAULT (date('now','localtime')),
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
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_cotizaciones_empresa_estado ON empresa_cotizaciones_venta(empresa_id, estado_documento, fecha_documento DESC);`,

		`CREATE TABLE IF NOT EXISTS empresa_pedidos_venta (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			cliente_id INTEGER DEFAULT 0,
			cliente_nombre TEXT,
			cotizacion_id INTEGER DEFAULT 0,
			fecha_pedido TEXT DEFAULT (date('now','localtime')),
			fecha_entrega_estimada TEXT,
			estado_pedido TEXT DEFAULT 'borrador',
			subtotal REAL DEFAULT 0,
			descuento_total REAL DEFAULT 0,
			impuesto_total REAL DEFAULT 0,
			total REAL DEFAULT 0,
			moneda TEXT DEFAULT 'COP',
			notas TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_pedidos_empresa_estado ON empresa_pedidos_venta(empresa_id, estado_pedido, fecha_pedido DESC);`,

		`CREATE TABLE IF NOT EXISTS empresa_devoluciones_venta (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			carrito_id INTEGER DEFAULT 0,
			documento_referencia TEXT,
			motivo TEXT,
			fecha_devolucion TEXT DEFAULT (date('now','localtime')),
			estado_devolucion TEXT DEFAULT 'borrador',
			subtotal REAL DEFAULT 0,
			impuesto_total REAL DEFAULT 0,
			total REAL DEFAULT 0,
			moneda TEXT DEFAULT 'COP',
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_devoluciones_venta_empresa_fecha ON empresa_devoluciones_venta(empresa_id, fecha_devolucion DESC);`,

		`CREATE TABLE IF NOT EXISTS empresa_plan_cuentas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			tipo_cuenta TEXT DEFAULT 'activo',
			naturaleza TEXT DEFAULT 'debito',
			nivel INTEGER DEFAULT 1,
			cuenta_padre_codigo TEXT,
			admite_movimiento INTEGER DEFAULT 1,
			aplica_impuesto INTEGER DEFAULT 0,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_plan_cuentas_empresa_tipo ON empresa_plan_cuentas(empresa_id, tipo_cuenta, codigo);`,

		`CREATE TABLE IF NOT EXISTS empresa_cuentas_por_cobrar (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_cxc_empresa_estado ON empresa_cuentas_por_cobrar(empresa_id, estado_cartera, fecha_vencimiento);`,

		`CREATE TABLE IF NOT EXISTS empresa_cuentas_por_pagar (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_cxp_empresa_estado ON empresa_cuentas_por_pagar(empresa_id, estado_cartera, fecha_vencimiento);`,

		`CREATE TABLE IF NOT EXISTS inventario_lotes_series (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			producto_id INTEGER NOT NULL,
			bodega_id INTEGER DEFAULT 0,
			tipo_control TEXT DEFAULT 'lote',
			codigo_lote_serie TEXT NOT NULL,
			fecha_fabricacion TEXT,
			fecha_vencimiento TEXT,
			cantidad_inicial REAL DEFAULT 0,
			cantidad_disponible REAL DEFAULT 0,
			costo_unitario REAL DEFAULT 0,
			estado_lote TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, producto_id, bodega_id, codigo_lote_serie)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_lotes_empresa_vencimiento ON inventario_lotes_series(empresa_id, fecha_vencimiento, producto_id);`,

		`CREATE TABLE IF NOT EXISTS empresa_devoluciones_proveedor (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			proveedor_id INTEGER DEFAULT 0,
			proveedor_nombre TEXT,
			documento_compra_codigo TEXT,
			fecha_devolucion TEXT DEFAULT (date('now','localtime')),
			motivo TEXT,
			estado_devolucion TEXT DEFAULT 'borrador',
			subtotal REAL DEFAULT 0,
			impuesto_total REAL DEFAULT 0,
			total REAL DEFAULT 0,
			moneda TEXT DEFAULT 'COP',
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_dev_prov_empresa_fecha ON empresa_devoluciones_proveedor(empresa_id, fecha_devolucion DESC);`,

		`CREATE TABLE IF NOT EXISTS empresa_rrhh_vacaciones_licencias (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			empleado_id INTEGER DEFAULT 0,
			empleado_nombre TEXT,
			tipo_novedad TEXT DEFAULT 'vacacion',
			fecha_inicio TEXT,
			fecha_fin TEXT,
			dias REAL DEFAULT 0,
			remunerada INTEGER DEFAULT 1,
			estado_novedad TEXT DEFAULT 'solicitada',
			soporte_url TEXT,
			aprobado_por TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_rrhh_vac_lic_empresa_fechas ON empresa_rrhh_vacaciones_licencias(empresa_id, fecha_inicio, fecha_fin);`,

		`CREATE TABLE IF NOT EXISTS crm_leads (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_crm_leads_empresa_estado ON crm_leads(empresa_id, estado_lead, proximo_contacto);`,

		`CREATE TABLE IF NOT EXISTS crm_interacciones (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			lead_id INTEGER DEFAULT 0,
			cliente_id INTEGER DEFAULT 0,
			tipo_interaccion TEXT DEFAULT 'llamada',
			fecha_interaccion TEXT DEFAULT (datetime('now','localtime')),
			resumen TEXT,
			resultado TEXT,
			usuario_responsable TEXT,
			proxima_accion TEXT,
			estado_interaccion TEXT DEFAULT 'abierta',
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_crm_interacciones_empresa_fecha ON crm_interacciones(empresa_id, fecha_interaccion DESC);`,

		`CREATE TABLE IF NOT EXISTS crm_campanas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_crm_campanas_empresa_estado ON crm_campanas(empresa_id, estado_campana, fecha_inicio);`,

		`CREATE TABLE IF NOT EXISTS produccion_bom (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			producto_id INTEGER DEFAULT 0,
			producto_nombre TEXT,
			version TEXT DEFAULT '1.0',
			rendimiento REAL DEFAULT 1,
			unidad_medida TEXT,
			costo_estimado_total REAL DEFAULT 0,
			estado_bom TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo, version)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_bom_empresa_producto ON produccion_bom(empresa_id, producto_id, version);`,

		`CREATE TABLE IF NOT EXISTS produccion_bom_detalle (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			bom_id INTEGER NOT NULL,
			insumo_producto_id INTEGER DEFAULT 0,
			insumo_nombre TEXT,
			cantidad REAL DEFAULT 0,
			unidad_medida TEXT,
			costo_unitario REAL DEFAULT 0,
			costo_total REAL DEFAULT 0,
			merma_porcentaje REAL DEFAULT 0,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_bom_detalle_empresa_bom ON produccion_bom_detalle(empresa_id, bom_id);`,

		`CREATE TABLE IF NOT EXISTS produccion_ordenes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_produccion_ordenes_empresa_estado ON produccion_ordenes(empresa_id, estado_orden, fecha_programada);`,

		`CREATE TABLE IF NOT EXISTS logistica_transportistas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_log_transportistas_empresa_nombre ON logistica_transportistas(empresa_id, nombre);`,

		`CREATE TABLE IF NOT EXISTS logistica_rutas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT,
			origen TEXT,
			destino TEXT,
			distancia_km REAL DEFAULT 0,
			tiempo_estimado_min REAL DEFAULT 0,
			estado_ruta TEXT DEFAULT 'activa',
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_log_rutas_empresa_origen_destino ON logistica_rutas(empresa_id, origen, destino);`,

		`CREATE TABLE IF NOT EXISTS logistica_envios (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_log_envios_empresa_estado ON logistica_envios(empresa_id, estado_envio, fecha_programada);`,

		`CREATE TABLE IF NOT EXISTS empresa_documentos_gestion (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_doc_gestion_empresa_modulo ON empresa_documentos_gestion(empresa_id, modulo, entidad_id);`,

		`CREATE TABLE IF NOT EXISTS empresa_documentos_firmas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_doc_firmas_empresa_doc ON empresa_documentos_firmas(empresa_id, documento_gestion_id, fecha_firma);`,

		`CREATE TABLE IF NOT EXISTS empresa_integraciones_apis (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_integraciones_api_empresa_estado ON empresa_integraciones_apis(empresa_id, estado_integracion);`,

		`CREATE TABLE IF NOT EXISTS empresa_integraciones_bancos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_integraciones_bancos_empresa_estado ON empresa_integraciones_bancos(empresa_id, estado_integracion);`,

		`CREATE TABLE IF NOT EXISTS empresa_dian_configuracion (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			nit TEXT,
			digito_verificacion TEXT,
			razon_social TEXT,
			tipo_ambiente TEXT DEFAULT 'habilitacion',
			software_id TEXT,
			software_pin TEXT,
			test_set_id TEXT,
			certificado_url TEXT,
			certificado_clave_ref TEXT,
			prefijo TEXT,
			resolucion_numero TEXT,
			resolucion_fecha_desde TEXT,
			resolucion_fecha_hasta TEXT,
			rango_desde INTEGER DEFAULT 0,
			rango_hasta INTEGER DEFAULT 0,
			consecutivo_actual INTEGER DEFAULT 0,
			url_dian TEXT,
			token_emisor_ref TEXT,
			ultimo_envio TEXT,
			estado_dian TEXT DEFAULT 'pendiente',
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id),
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_dian_empresa_ambiente ON empresa_dian_configuracion(empresa_id, tipo_ambiente, estado_dian);`,
	}

	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
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
	res, err := dbConn.Exec(query, values...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
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

	setParts = append(setParts, "fecha_actualizacion = datetime('now','localtime')")
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

	estado = strings.ToLower(strings.TrimSpace(estado))
	if estado == "" {
		estado = "activo"
	}

	_, err := dbConn.Exec("UPDATE "+table+" SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = ? AND id = ?", estado, empresaID, id)
	return err
}

// DeleteEmpresaGenericRow aplica eliminacion logica por empresa e id.
func DeleteEmpresaGenericRow(dbConn *sql.DB, table string, empresaID, id int64) error {
	return SetEmpresaGenericRowEstado(dbConn, table, empresaID, id, "inactivo")
}
