package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

var (
	empresaFinanzasSchemaMu    sync.Mutex
	empresaFinanzasSchemaReady bool
)

var ErrPeriodoFinancieroCerrado = errors.New("el periodo contable esta cerrado")
var ErrCierreCajaTransicionInvalida = errors.New("la transicion de estado del cierre de caja no es valida")
var ErrCierreCajaAprobadoBloqueado = errors.New("el cierre de caja aprobado no permite modificaciones")

// EmpresaFinanzasMovimiento representa un registro de ingreso/egreso por empresa.
type EmpresaFinanzasMovimiento struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	TipoMovimiento     string  `json:"tipo_movimiento"`
	Codigo             string  `json:"codigo"`
	FechaMovimiento    string  `json:"fecha_movimiento"`
	PeriodoContable    string  `json:"periodo_contable"`
	Categoria          string  `json:"categoria"`
	Subcategoria       string  `json:"subcategoria"`
	Concepto           string  `json:"concepto"`
	Descripcion        string  `json:"descripcion"`
	MetodoPago         string  `json:"metodo_pago"`
	Moneda             string  `json:"moneda"`
	Monto              float64 `json:"monto"`
	Impuesto           float64 `json:"impuesto"`
	RetencionFuente    float64 `json:"retencion_fuente"`
	RetencionICA       float64 `json:"retencion_ica"`
	RetencionIVA       float64 `json:"retencion_iva"`
	TotalRetenciones   float64 `json:"total_retenciones"`
	Total              float64 `json:"total"`
	TotalNeto          float64 `json:"total_neto"`
	TerceroNombre      string  `json:"tercero_nombre"`
	TerceroDocumento   string  `json:"tercero_documento"`
	TipoComprobante    string  `json:"tipo_comprobante"`
	NumeroComprobante  string  `json:"numero_comprobante"`
	ComprobanteURL     string  `json:"comprobante_url"`
	ReferenciaExterna  string  `json:"referencia_externa"`
	AprobadoPor        string  `json:"aprobado_por"`
	FechaCreacion      string  `json:"fecha_creacion"`
	FechaActualizacion string  `json:"fecha_actualizacion"`
	UsuarioCreador     string  `json:"usuario_creador"`
	Estado             string  `json:"estado"`
	Observaciones      string  `json:"observaciones"`
}

// EmpresaFinanzasMovimientoFilter permite filtrar listados financieros por empresa.
type EmpresaFinanzasMovimientoFilter struct {
	Tipo            string
	Desde           string
	Hasta           string
	Periodo         string
	Q               string
	IncludeInactive bool
	Limit           int
}

type EmpresaFinanzasPeriodo struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Periodo            string `json:"periodo"`
	FechaInicio        string `json:"fecha_inicio"`
	FechaFin           string `json:"fecha_fin"`
	Estado             string `json:"estado"`
	FechaCierre        string `json:"fecha_cierre"`
	CerradoPor         string `json:"cerrado_por"`
	FechaCreacion      string `json:"fecha_creacion"`
	FechaActualizacion string `json:"fecha_actualizacion"`
	UsuarioCreador     string `json:"usuario_creador"`
	Observaciones      string `json:"observaciones"`
}

type EmpresaCierreCaja struct {
	ID                    int64   `json:"id"`
	EmpresaID             int64   `json:"empresa_id"`
	SucursalID            int64   `json:"sucursal_id"`
	CajaCodigo            string  `json:"caja_codigo"`
	Turno                 string  `json:"turno"`
	FechaOperacion        string  `json:"fecha_operacion"`
	FechaApertura         string  `json:"fecha_apertura"`
	FechaCierre           string  `json:"fecha_cierre"`
	EstadoCierre          string  `json:"estado_cierre"`
	AperturaMonto         float64 `json:"apertura_monto"`
	IngresosEfectivo      float64 `json:"ingresos_efectivo"`
	EgresosEfectivo       float64 `json:"egresos_efectivo"`
	RetirosEfectivo       float64 `json:"retiros_efectivo"`
	CajaTeorica           float64 `json:"caja_teorica"`
	CajaFisica            float64 `json:"caja_fisica"`
	DiferenciaCaja        float64 `json:"diferencia_caja"`
	Moneda                string  `json:"moneda"`
	CerradoPor            string  `json:"cerrado_por"`
	AprobadoPor           string  `json:"aprobado_por"`
	AprobadoEn            string  `json:"aprobado_en"`
	TieneIncidencia       bool    `json:"tiene_incidencia"`
	UmbralIncidencia      float64 `json:"umbral_incidencia"`
	PropinasMovimientos   int64   `json:"propinas_movimientos"`
	PropinasTotal         float64 `json:"propinas_total"`
	PropinasAjustes       float64 `json:"propinas_ajustes"`
	PropinasImpuesto      float64 `json:"propinas_impuesto"`
	PropinasNeto          float64 `json:"propinas_neto"`
	PropinasConciliadoEn  string  `json:"propinas_conciliado_en"`
	PropinasConciliadoPor string  `json:"propinas_conciliado_por"`
	FechaCreacion         string  `json:"fecha_creacion"`
	FechaActualizacion    string  `json:"fecha_actualizacion"`
	UsuarioCreador        string  `json:"usuario_creador"`
	Estado                string  `json:"estado"`
	Observaciones         string  `json:"observaciones"`
}

type EmpresaCierreCajaFilter struct {
	SucursalID      int64
	CajaCodigo      string
	EstadoCierre    string
	Desde           string
	Hasta           string
	IncludeInactive bool
	Limit           int
}

// EmpresaFinanzasConfiguracion define la parametrizacion por empresa del modulo financiero.
type EmpresaFinanzasConfiguracion struct {
	ID                         int64  `json:"id"`
	EmpresaID                  int64  `json:"empresa_id"`
	HabilitarIngresos          bool   `json:"habilitar_ingresos"`
	HabilitarEgresos           bool   `json:"habilitar_egresos"`
	Moneda                     string `json:"moneda"`
	CategoriasIngreso          string `json:"categorias_ingreso"`
	CategoriasEgreso           string `json:"categorias_egreso"`
	PrefijoIngreso             string `json:"prefijo_ingreso"`
	PrefijoEgreso              string `json:"prefijo_egreso"`
	FormatoImpresion           string `json:"formato_impresion"`
	RequiereAprobacion         bool   `json:"requiere_aprobacion"`
	IntegracionContableDestino string `json:"integracion_contable_destino"`
	CuentaCajaBancos           string `json:"cuenta_caja_bancos"`
	CuentaIngresos             string `json:"cuenta_ingresos"`
	CuentaIVAGenerado          string `json:"cuenta_iva_generado"`
	CuentaGastos               string `json:"cuenta_gastos"`
	CuentaIVADescontable       string `json:"cuenta_iva_descontable"`
	CuentaRetencionesCobrar    string `json:"cuenta_retenciones_cobrar"`
	CuentaRetencionesPagar     string `json:"cuenta_retenciones_pagar"`
	CuentasIngresoCategoria    string `json:"cuentas_ingreso_categoria"`
	CuentasEgresoCategoria     string `json:"cuentas_egreso_categoria"`
	FechaCreacion              string `json:"fecha_creacion"`
	FechaActualizacion         string `json:"fecha_actualizacion"`
	UsuarioCreador             string `json:"usuario_creador"`
	Estado                     string `json:"estado"`
	Observaciones              string `json:"observaciones"`
}

// EnsureEmpresaFinanzasSchema crea y migra las tablas del modulo financiero en empresas.db.
func EnsureEmpresaFinanzasSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("db connection is nil")
	}
	empresaFinanzasSchemaMu.Lock()
	defer empresaFinanzasSchemaMu.Unlock()

	if empresaFinanzasSchemaReady {
		return nil
	}
	ready, err := empresaFinanzasSchemaLooksReady(dbConn)
	if err == nil && ready {
		empresaFinanzasSchemaReady = true
		return nil
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_finanzas_movimientos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			tipo_movimiento TEXT NOT NULL,
			codigo TEXT NOT NULL,
			fecha_movimiento TEXT DEFAULT (datetime('now','localtime')),
			periodo_contable TEXT,
			categoria TEXT,
			subcategoria TEXT,
			concepto TEXT,
			descripcion TEXT,
			metodo_pago TEXT,
			moneda TEXT DEFAULT 'COP',
			monto REAL DEFAULT 0,
			impuesto REAL DEFAULT 0,
			retencion_fuente REAL DEFAULT 0,
			retencion_ica REAL DEFAULT 0,
			retencion_iva REAL DEFAULT 0,
			total_retenciones REAL DEFAULT 0,
			total REAL DEFAULT 0,
			total_neto REAL DEFAULT 0,
			tercero_nombre TEXT,
			tercero_documento TEXT,
			tipo_comprobante TEXT DEFAULT 'recibo_interno',
			numero_comprobante TEXT,
			comprobante_url TEXT,
			referencia_externa TEXT,
			aprobado_por TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE TABLE IF NOT EXISTS empresa_finanzas_configuracion (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL UNIQUE,
			habilitar_ingresos INTEGER DEFAULT 1,
			habilitar_egresos INTEGER DEFAULT 1,
			moneda TEXT DEFAULT 'COP',
			categorias_ingreso TEXT,
			categorias_egreso TEXT,
			prefijo_ingreso TEXT DEFAULT 'ING',
			prefijo_egreso TEXT DEFAULT 'EGR',
			formato_impresion TEXT DEFAULT 'carta',
			requiere_aprobacion INTEGER DEFAULT 0,
			integracion_contable_destino TEXT DEFAULT 'generico',
			cuenta_caja_bancos TEXT DEFAULT '110505',
			cuenta_ingresos TEXT DEFAULT '413595',
			cuenta_iva_generado TEXT DEFAULT '240805',
			cuenta_gastos TEXT DEFAULT '519595',
			cuenta_iva_descontable TEXT DEFAULT '240810',
			cuenta_retenciones_cobrar TEXT DEFAULT '135595',
			cuenta_retenciones_pagar TEXT DEFAULT '236595',
			cuentas_ingreso_categoria TEXT,
			cuentas_egreso_categoria TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS empresa_finanzas_periodos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			periodo TEXT NOT NULL,
			fecha_inicio TEXT,
			fecha_fin TEXT,
			estado TEXT DEFAULT 'abierto',
			fecha_cierre TEXT,
			cerrado_por TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			observaciones TEXT,
			UNIQUE(empresa_id, periodo)
		);`,
		`CREATE TABLE IF NOT EXISTS empresa_cierres_caja (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			sucursal_id INTEGER DEFAULT 0,
			caja_codigo TEXT NOT NULL,
			turno TEXT DEFAULT 'general',
			fecha_operacion TEXT DEFAULT (date('now','localtime')),
			fecha_apertura TEXT DEFAULT (datetime('now','localtime')),
			fecha_cierre TEXT,
			estado_cierre TEXT DEFAULT 'abierto',
			apertura_monto REAL DEFAULT 0,
			ingresos_efectivo REAL DEFAULT 0,
			egresos_efectivo REAL DEFAULT 0,
			retiros_efectivo REAL DEFAULT 0,
			caja_teorica REAL DEFAULT 0,
			caja_fisica REAL DEFAULT 0,
			diferencia_caja REAL DEFAULT 0,
			moneda TEXT DEFAULT 'COP',
			cerrado_por TEXT,
			aprobado_por TEXT,
			aprobado_en TEXT,
			tiene_incidencia INTEGER DEFAULT 0,
			umbral_incidencia REAL DEFAULT 0,
			propinas_movimientos INTEGER DEFAULT 0,
			propinas_total REAL DEFAULT 0,
			propinas_ajustes REAL DEFAULT 0,
			propinas_impuesto REAL DEFAULT 0,
			propinas_neto REAL DEFAULT 0,
			propinas_conciliado_en TEXT,
			propinas_conciliado_por TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, sucursal_id, caja_codigo, fecha_operacion, turno)
		);`,
		`CREATE TABLE IF NOT EXISTS empresa_finanzas_bancos_movimientos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			periodo_contable TEXT,
			fecha_movimiento TEXT DEFAULT (datetime('now','localtime')),
			fecha_valor TEXT,
			cuenta_bancaria TEXT,
			banco_nombre TEXT,
			tipo_movimiento TEXT NOT NULL,
			descripcion TEXT,
			referencia_bancaria TEXT,
			documento_codigo TEXT,
			moneda TEXT DEFAULT 'COP',
			monto REAL DEFAULT 0,
			total REAL DEFAULT 0,
			movimiento_finanzas_id INTEGER DEFAULT 0,
			estado_conciliacion TEXT DEFAULT 'pendiente',
			conciliado_en TEXT,
			conciliado_por TEXT,
			origen TEXT DEFAULT 'manual',
			hash_movimiento TEXT NOT NULL,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, hash_movimiento)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_finanzas_movimientos_empresa_fecha ON empresa_finanzas_movimientos(empresa_id, fecha_movimiento DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_finanzas_movimientos_empresa_tipo_estado ON empresa_finanzas_movimientos(empresa_id, tipo_movimiento, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_finanzas_movimientos_empresa_estado_fecha_usuario ON empresa_finanzas_movimientos(empresa_id, estado, fecha_movimiento DESC, usuario_creador);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_finanzas_movimientos_empresa_comprobante ON empresa_finanzas_movimientos(empresa_id, numero_comprobante);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_cierres_caja_empresa_fecha ON empresa_cierres_caja(empresa_id, fecha_operacion DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_cierres_caja_empresa_estado ON empresa_cierres_caja(empresa_id, estado_cierre, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_finanzas_bancos_movimientos_empresa_fecha ON empresa_finanzas_bancos_movimientos(empresa_id, fecha_movimiento DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_finanzas_bancos_movimientos_empresa_periodo_estado ON empresa_finanzas_bancos_movimientos(empresa_id, periodo_contable, estado_conciliacion, estado);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_movimientos", "fecha_movimiento", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_movimientos", "periodo_contable", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_movimientos", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_movimientos", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_movimientos", "observaciones", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_movimientos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_movimientos", "tipo_comprobante", "TEXT DEFAULT 'recibo_interno'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_movimientos", "numero_comprobante", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_movimientos", "comprobante_url", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_movimientos", "retencion_fuente", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_movimientos", "retencion_ica", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_movimientos", "retencion_iva", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_movimientos", "total_retenciones", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_movimientos", "total_neto", "REAL DEFAULT 0"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "habilitar_ingresos", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "habilitar_egresos", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "prefijo_ingreso", "TEXT DEFAULT 'ING'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "prefijo_egreso", "TEXT DEFAULT 'EGR'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "formato_impresion", "TEXT DEFAULT 'carta'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "requiere_aprobacion", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "integracion_contable_destino", "TEXT DEFAULT 'generico'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "cuenta_caja_bancos", "TEXT DEFAULT '110505'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "cuenta_ingresos", "TEXT DEFAULT '413595'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "cuenta_iva_generado", "TEXT DEFAULT '240805'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "cuenta_gastos", "TEXT DEFAULT '519595'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "cuenta_iva_descontable", "TEXT DEFAULT '240810'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "cuenta_retenciones_cobrar", "TEXT DEFAULT '135595'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "cuenta_retenciones_pagar", "TEXT DEFAULT '236595'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "cuentas_ingreso_categoria", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "cuentas_egreso_categoria", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "observaciones", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_periodos", "fecha_inicio", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_periodos", "fecha_fin", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_periodos", "estado", "TEXT DEFAULT 'abierto'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_periodos", "fecha_cierre", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_periodos", "cerrado_por", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_periodos", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_periodos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_periodos", "observaciones", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "sucursal_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "caja_codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "turno", "TEXT DEFAULT 'general'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "fecha_operacion", "TEXT DEFAULT (date('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "fecha_apertura", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "fecha_cierre", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "estado_cierre", "TEXT DEFAULT 'abierto'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "apertura_monto", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "ingresos_efectivo", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "egresos_efectivo", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "retiros_efectivo", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "caja_teorica", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "caja_fisica", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "diferencia_caja", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "moneda", "TEXT DEFAULT 'COP'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "cerrado_por", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "aprobado_por", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "aprobado_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "tiene_incidencia", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "umbral_incidencia", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "propinas_movimientos", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "propinas_total", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "propinas_ajustes", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "propinas_impuesto", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "propinas_neto", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "propinas_conciliado_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "propinas_conciliado_por", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_cierres_caja", "observaciones", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_bancos_movimientos", "periodo_contable", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_bancos_movimientos", "fecha_movimiento", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_bancos_movimientos", "fecha_valor", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_bancos_movimientos", "cuenta_bancaria", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_bancos_movimientos", "banco_nombre", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_bancos_movimientos", "descripcion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_bancos_movimientos", "referencia_bancaria", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_bancos_movimientos", "documento_codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_bancos_movimientos", "moneda", "TEXT DEFAULT 'COP'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_bancos_movimientos", "monto", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_bancos_movimientos", "total", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_bancos_movimientos", "movimiento_finanzas_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_bancos_movimientos", "estado_conciliacion", "TEXT DEFAULT 'pendiente'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_bancos_movimientos", "conciliado_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_bancos_movimientos", "conciliado_por", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_bancos_movimientos", "origen", "TEXT DEFAULT 'manual'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_bancos_movimientos", "hash_movimiento", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_bancos_movimientos", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_bancos_movimientos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_bancos_movimientos", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_bancos_movimientos", "observaciones", "TEXT"); err != nil {
		return err
	}

	// Los indices que dependen de columnas agregadas en migracion se crean al final
	// para mantener compatibilidad con bases existentes.
	postMigrationIndexes := []string{
		`CREATE INDEX IF NOT EXISTS ix_empresa_finanzas_movimientos_empresa_periodo ON empresa_finanzas_movimientos(empresa_id, periodo_contable, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_finanzas_periodos_empresa_estado ON empresa_finanzas_periodos(empresa_id, estado, periodo DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_cierres_caja_empresa_sucursal ON empresa_cierres_caja(empresa_id, sucursal_id, caja_codigo, fecha_operacion DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_finanzas_bancos_movimientos_empresa_hash ON empresa_finanzas_bancos_movimientos(empresa_id, hash_movimiento);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_finanzas_bancos_movimientos_empresa_movimiento ON empresa_finanzas_bancos_movimientos(empresa_id, movimiento_finanzas_id, estado_conciliacion);`,
	}
	for _, stmt := range postMigrationIndexes {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	empresaFinanzasSchemaReady = true
	return nil
}

func empresaFinanzasSchemaLooksReady(dbConn *sql.DB) (bool, error) {
	ok, err := tableExists(dbConn, "empresa_finanzas_movimientos")
	if err != nil || !ok {
		return false, err
	}
	ok, err = tableExists(dbConn, "empresa_finanzas_configuracion")
	if err != nil || !ok {
		return false, err
	}
	ok, err = tableExists(dbConn, "empresa_finanzas_periodos")
	if err != nil || !ok {
		return false, err
	}
	ok, err = tableExists(dbConn, "empresa_cierres_caja")
	if err != nil || !ok {
		return false, err
	}
	ok, err = tableExists(dbConn, "empresa_finanzas_bancos_movimientos")
	if err != nil || !ok {
		return false, err
	}

	requiredIndexes := []string{
		"ix_empresa_finanzas_movimientos_empresa_fecha",
		"ix_empresa_cierres_caja_empresa_estado",
		"ix_empresa_finanzas_bancos_movimientos_empresa_periodo_estado",
		"ix_empresa_finanzas_periodos_empresa_estado",
	}
	for _, indexName := range requiredIndexes {
		indexOK, idxErr := empresaFinanzasIndexExists(dbConn, indexName)
		if idxErr != nil || !indexOK {
			return false, idxErr
		}
	}
	return true, nil
}

func empresaFinanzasIndexExists(dbConn *sql.DB, indexName string) (bool, error) {
	var exists bool
	err := queryRowSQLCompat(dbConn, `
		SELECT EXISTS (
			SELECT 1
			FROM sqlite_master
			WHERE type = 'index' AND name = ?
		)
	`, indexName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// GetEmpresaFinanzasConfiguracion obtiene configuracion financiera por empresa con defaults seguros.
func GetEmpresaFinanzasConfiguracion(dbConn *sql.DB, empresaID int64) (*EmpresaFinanzasConfiguracion, error) {
	cfg := defaultEmpresaFinanzasConfiguracion(empresaID)
	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(habilitar_ingresos, 1),
		COALESCE(habilitar_egresos, 1),
		COALESCE(moneda, 'COP'),
		COALESCE(categorias_ingreso, ''),
		COALESCE(categorias_egreso, ''),
		COALESCE(prefijo_ingreso, 'ING'),
		COALESCE(prefijo_egreso, 'EGR'),
		COALESCE(formato_impresion, 'carta'),
		COALESCE(requiere_aprobacion, 0),
		COALESCE(integracion_contable_destino, 'generico'),
		COALESCE(cuenta_caja_bancos, '110505'),
		COALESCE(cuenta_ingresos, '413595'),
		COALESCE(cuenta_iva_generado, '240805'),
		COALESCE(cuenta_gastos, '519595'),
		COALESCE(cuenta_iva_descontable, '240810'),
		COALESCE(cuenta_retenciones_cobrar, '135595'),
		COALESCE(cuenta_retenciones_pagar, '236595'),
		COALESCE(cuentas_ingreso_categoria, ''),
		COALESCE(cuentas_egreso_categoria, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_finanzas_configuracion
	WHERE empresa_id = ?
	LIMIT 1`, empresaID)

	var habilitarIngresos int
	var habilitarEgresos int
	var requiereAprobacion int
	err := row.Scan(
		&cfg.ID,
		&cfg.EmpresaID,
		&habilitarIngresos,
		&habilitarEgresos,
		&cfg.Moneda,
		&cfg.CategoriasIngreso,
		&cfg.CategoriasEgreso,
		&cfg.PrefijoIngreso,
		&cfg.PrefijoEgreso,
		&cfg.FormatoImpresion,
		&requiereAprobacion,
		&cfg.IntegracionContableDestino,
		&cfg.CuentaCajaBancos,
		&cfg.CuentaIngresos,
		&cfg.CuentaIVAGenerado,
		&cfg.CuentaGastos,
		&cfg.CuentaIVADescontable,
		&cfg.CuentaRetencionesCobrar,
		&cfg.CuentaRetencionesPagar,
		&cfg.CuentasIngresoCategoria,
		&cfg.CuentasEgresoCategoria,
		&cfg.FechaCreacion,
		&cfg.FechaActualizacion,
		&cfg.UsuarioCreador,
		&cfg.Estado,
		&cfg.Observaciones,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return cfg, nil
		}
		return nil, err
	}
	cfg.HabilitarIngresos = habilitarIngresos == 1
	cfg.HabilitarEgresos = habilitarEgresos == 1
	cfg.RequiereAprobacion = requiereAprobacion == 1
	return cfg, nil
}

// UpsertEmpresaFinanzasConfiguracion crea o actualiza la configuracion financiera por empresa.
func UpsertEmpresaFinanzasConfiguracion(dbConn *sql.DB, cfg EmpresaFinanzasConfiguracion) (int64, error) {
	cfg = normalizeEmpresaFinanzasConfiguracion(cfg)

	res, err := dbConn.Exec(`INSERT INTO empresa_finanzas_configuracion (
		empresa_id, habilitar_ingresos, habilitar_egresos, moneda,
		categorias_ingreso, categorias_egreso,
		prefijo_ingreso, prefijo_egreso,
		formato_impresion, requiere_aprobacion,
		integracion_contable_destino,
		cuenta_caja_bancos, cuenta_ingresos, cuenta_iva_generado,
		cuenta_gastos, cuenta_iva_descontable,
		cuenta_retenciones_cobrar, cuenta_retenciones_pagar,
		cuentas_ingreso_categoria, cuentas_egreso_categoria,
		usuario_creador, estado, observaciones,
		fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))
	ON CONFLICT(empresa_id) DO UPDATE SET
		habilitar_ingresos = excluded.habilitar_ingresos,
		habilitar_egresos = excluded.habilitar_egresos,
		moneda = excluded.moneda,
		categorias_ingreso = excluded.categorias_ingreso,
		categorias_egreso = excluded.categorias_egreso,
		prefijo_ingreso = excluded.prefijo_ingreso,
		prefijo_egreso = excluded.prefijo_egreso,
		formato_impresion = excluded.formato_impresion,
		requiere_aprobacion = excluded.requiere_aprobacion,
		integracion_contable_destino = excluded.integracion_contable_destino,
		cuenta_caja_bancos = excluded.cuenta_caja_bancos,
		cuenta_ingresos = excluded.cuenta_ingresos,
		cuenta_iva_generado = excluded.cuenta_iva_generado,
		cuenta_gastos = excluded.cuenta_gastos,
		cuenta_iva_descontable = excluded.cuenta_iva_descontable,
		cuenta_retenciones_cobrar = excluded.cuenta_retenciones_cobrar,
		cuenta_retenciones_pagar = excluded.cuenta_retenciones_pagar,
		cuentas_ingreso_categoria = excluded.cuentas_ingreso_categoria,
		cuentas_egreso_categoria = excluded.cuentas_egreso_categoria,
		usuario_creador = excluded.usuario_creador,
		estado = excluded.estado,
		observaciones = excluded.observaciones,
		fecha_actualizacion = datetime('now','localtime')`,
		cfg.EmpresaID,
		boolToInt(cfg.HabilitarIngresos),
		boolToInt(cfg.HabilitarEgresos),
		cfg.Moneda,
		cfg.CategoriasIngreso,
		cfg.CategoriasEgreso,
		cfg.PrefijoIngreso,
		cfg.PrefijoEgreso,
		cfg.FormatoImpresion,
		boolToInt(cfg.RequiereAprobacion),
		cfg.IntegracionContableDestino,
		cfg.CuentaCajaBancos,
		cfg.CuentaIngresos,
		cfg.CuentaIVAGenerado,
		cfg.CuentaGastos,
		cfg.CuentaIVADescontable,
		cfg.CuentaRetencionesCobrar,
		cfg.CuentaRetencionesPagar,
		cfg.CuentasIngresoCategoria,
		cfg.CuentasEgresoCategoria,
		cfg.UsuarioCreador,
		cfg.Estado,
		cfg.Observaciones,
	)
	if err != nil {
		return 0, err
	}
	if id, errID := res.LastInsertId(); errID == nil && id > 0 {
		return id, nil
	}
	current, err := GetEmpresaFinanzasConfiguracion(dbConn, cfg.EmpresaID)
	if err != nil {
		return 0, err
	}
	return current.ID, nil
}

// CreateEmpresaFinanzasMovimiento crea un movimiento financiero por empresa.
func CreateEmpresaFinanzasMovimiento(dbConn *sql.DB, m EmpresaFinanzasMovimiento) (int64, error) {
	m, err := normalizeEmpresaFinanzasMovimiento(dbConn, m, true)
	if err != nil {
		return 0, err
	}
	cerrado, err := IsEmpresaFinanzasPeriodoCerrado(dbConn, m.EmpresaID, m.PeriodoContable)
	if err != nil {
		return 0, err
	}
	if cerrado {
		return 0, ErrPeriodoFinancieroCerrado
	}

	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_finanzas_movimientos (
		empresa_id, tipo_movimiento, codigo, fecha_movimiento,
		periodo_contable,
		categoria, subcategoria, concepto, descripcion,
		metodo_pago, moneda, monto, impuesto,
		retencion_fuente, retencion_ica, retencion_iva, total_retenciones,
		total, total_neto,
		tercero_nombre, tercero_documento,
		tipo_comprobante, numero_comprobante, comprobante_url,
		referencia_externa, aprobado_por,
		usuario_creador, estado, observaciones,
		fecha_creacion, fecha_actualizacion
	) VALUES (
		?, ?, ?, ?,
		?,
		?, ?, ?, ?,
		?, ?, ?, ?,
		?, ?, ?, ?,
		?, ?,
		?, ?,
		?, ?, ?,
		?, ?,
		?, ?, ?,
		datetime('now','localtime'), datetime('now','localtime')
	)`,
		m.EmpresaID, m.TipoMovimiento, m.Codigo, m.FechaMovimiento,
		m.PeriodoContable,
		m.Categoria, m.Subcategoria, m.Concepto, m.Descripcion,
		m.MetodoPago, m.Moneda, m.Monto, m.Impuesto,
		m.RetencionFuente, m.RetencionICA, m.RetencionIVA, m.TotalRetenciones,
		m.Total, m.TotalNeto,
		m.TerceroNombre, m.TerceroDocumento,
		m.TipoComprobante, m.NumeroComprobante, m.ComprobanteURL,
		m.ReferenciaExterna, m.AprobadoPor,
		m.UsuarioCreador, m.Estado, m.Observaciones,
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ListEmpresaFinanzasMovimientos lista movimientos financieros por empresa con filtros.
func ListEmpresaFinanzasMovimientos(dbConn *sql.DB, empresaID int64, f EmpresaFinanzasMovimientoFilter) ([]EmpresaFinanzasMovimiento, error) {
	query := `SELECT
		id, empresa_id, COALESCE(tipo_movimiento, 'egreso'), COALESCE(codigo, ''),
		COALESCE(fecha_movimiento, ''), COALESCE(periodo_contable, ''), COALESCE(categoria, ''), COALESCE(subcategoria, ''),
		COALESCE(concepto, ''), COALESCE(descripcion, ''), COALESCE(metodo_pago, ''),
		COALESCE(moneda, 'COP'), COALESCE(monto, 0), COALESCE(impuesto, 0),
		COALESCE(retencion_fuente, 0), COALESCE(retencion_ica, 0), COALESCE(retencion_iva, 0), COALESCE(total_retenciones, 0),
		COALESCE(total, 0), COALESCE(total_neto, 0),
		COALESCE(tercero_nombre, ''), COALESCE(tercero_documento, ''),
		COALESCE(tipo_comprobante, 'recibo_interno'), COALESCE(numero_comprobante, ''), COALESCE(comprobante_url, ''),
		COALESCE(referencia_externa, ''), COALESCE(aprobado_por, ''),
		COALESCE(fecha_creacion, ''), COALESCE(fecha_actualizacion, ''), COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'), COALESCE(observaciones, '')
	FROM empresa_finanzas_movimientos
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	f.Tipo = normalizeTipoMovimiento(f.Tipo)
	if f.Tipo != "" {
		query += ` AND tipo_movimiento = ?`
		args = append(args, f.Tipo)
	}
	if !f.IncludeInactive {
		query += ` AND COALESCE(estado, 'activo') = 'activo'`
	}
	if strings.TrimSpace(f.Desde) != "" {
		query += ` AND date(fecha_movimiento) >= date(?)`
		args = append(args, strings.TrimSpace(f.Desde))
	}
	if strings.TrimSpace(f.Hasta) != "" {
		query += ` AND date(fecha_movimiento) <= date(?)`
		args = append(args, strings.TrimSpace(f.Hasta))
	}
	if p := normalizePeriodoContable(f.Periodo); p != "" {
		query += ` AND COALESCE(periodo_contable, '') = ?`
		args = append(args, p)
	}
	if q := strings.TrimSpace(strings.ToLower(f.Q)); q != "" {
		like := "%" + q + "%"
		query += ` AND (
			LOWER(COALESCE(codigo, '')) LIKE ? OR
			LOWER(COALESCE(concepto, '')) LIKE ? OR
			LOWER(COALESCE(descripcion, '')) LIKE ? OR
			LOWER(COALESCE(numero_comprobante, '')) LIKE ? OR
			LOWER(COALESCE(tercero_nombre, '')) LIKE ?
		)`
		args = append(args, like, like, like, like, like)
	}

	query += ` ORDER BY datetime(fecha_movimiento) DESC, id DESC`
	limit := f.Limit
	if limit <= 0 {
		limit = 200
	}
	query += ` LIMIT ?`
	args = append(args, limit)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaFinanzasMovimiento, 0)
	for rows.Next() {
		var m EmpresaFinanzasMovimiento
		if err := rows.Scan(
			&m.ID,
			&m.EmpresaID,
			&m.TipoMovimiento,
			&m.Codigo,
			&m.FechaMovimiento,
			&m.PeriodoContable,
			&m.Categoria,
			&m.Subcategoria,
			&m.Concepto,
			&m.Descripcion,
			&m.MetodoPago,
			&m.Moneda,
			&m.Monto,
			&m.Impuesto,
			&m.RetencionFuente,
			&m.RetencionICA,
			&m.RetencionIVA,
			&m.TotalRetenciones,
			&m.Total,
			&m.TotalNeto,
			&m.TerceroNombre,
			&m.TerceroDocumento,
			&m.TipoComprobante,
			&m.NumeroComprobante,
			&m.ComprobanteURL,
			&m.ReferenciaExterna,
			&m.AprobadoPor,
			&m.FechaCreacion,
			&m.FechaActualizacion,
			&m.UsuarioCreador,
			&m.Estado,
			&m.Observaciones,
		); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// UpdateEmpresaFinanzasMovimiento actualiza un movimiento financiero por empresa.
func UpdateEmpresaFinanzasMovimiento(dbConn *sql.DB, m EmpresaFinanzasMovimiento) error {
	m, err := normalizeEmpresaFinanzasMovimiento(dbConn, m, false)
	if err != nil {
		return err
	}
	if err := ensurePeriodoAbiertoParaActualizarMovimiento(dbConn, m.EmpresaID, m.ID, m.PeriodoContable); err != nil {
		return err
	}
	res, err := dbConn.Exec(`UPDATE empresa_finanzas_movimientos SET
		tipo_movimiento = ?,
		codigo = ?,
		fecha_movimiento = ?,
		periodo_contable = ?,
		categoria = ?,
		subcategoria = ?,
		concepto = ?,
		descripcion = ?,
		metodo_pago = ?,
		moneda = ?,
		monto = ?,
		impuesto = ?,
		retencion_fuente = ?,
		retencion_ica = ?,
		retencion_iva = ?,
		total_retenciones = ?,
		total = ?,
		total_neto = ?,
		tercero_nombre = ?,
		tercero_documento = ?,
		tipo_comprobante = ?,
		numero_comprobante = ?,
		comprobante_url = ?,
		referencia_externa = ?,
		aprobado_por = ?,
		observaciones = ?,
		estado = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`,
		m.TipoMovimiento,
		m.Codigo,
		m.FechaMovimiento,
		m.PeriodoContable,
		m.Categoria,
		m.Subcategoria,
		m.Concepto,
		m.Descripcion,
		m.MetodoPago,
		m.Moneda,
		m.Monto,
		m.Impuesto,
		m.RetencionFuente,
		m.RetencionICA,
		m.RetencionIVA,
		m.TotalRetenciones,
		m.Total,
		m.TotalNeto,
		m.TerceroNombre,
		m.TerceroDocumento,
		m.TipoComprobante,
		m.NumeroComprobante,
		m.ComprobanteURL,
		m.ReferenciaExterna,
		m.AprobadoPor,
		m.Observaciones,
		m.Estado,
		m.EmpresaID,
		m.ID,
	)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// UpdateEmpresaFinanzasMovimientoComprobante actualiza la URL del comprobante adjunto de un movimiento.
func UpdateEmpresaFinanzasMovimientoComprobante(dbConn *sql.DB, empresaID, id int64, comprobanteURL string) error {
	if err := EnsureEmpresaFinanzasSchema(dbConn); err != nil {
		return err
	}
	if empresaID <= 0 {
		return fmt.Errorf("empresa_id es obligatorio")
	}
	if id <= 0 {
		return fmt.Errorf("id es obligatorio")
	}
	comprobanteURL = strings.TrimSpace(comprobanteURL)
	if comprobanteURL == "" {
		return fmt.Errorf("comprobante_url es obligatorio")
	}

	res, err := dbConn.Exec(`UPDATE empresa_finanzas_movimientos
		SET comprobante_url = ?,
			fecha_actualizacion = datetime('now','localtime')
		WHERE empresa_id = ? AND id = ?`, comprobanteURL, empresaID, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SetEmpresaFinanzasMovimientoEstado activa/desactiva/anula un movimiento financiero.
func SetEmpresaFinanzasMovimientoEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	estado = normalizeEstadoMovimiento(estado)
	periodo, err := getPeriodoContableMovimiento(dbConn, empresaID, id)
	if err != nil {
		return err
	}
	cerrado, err := IsEmpresaFinanzasPeriodoCerrado(dbConn, empresaID, periodo)
	if err != nil {
		return err
	}
	if cerrado {
		return ErrPeriodoFinancieroCerrado
	}
	res, err := dbConn.Exec(`UPDATE empresa_finanzas_movimientos
	SET estado = ?, fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`, estado, empresaID, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteEmpresaFinanzasMovimiento elimina un movimiento financiero por empresa.
func DeleteEmpresaFinanzasMovimiento(dbConn *sql.DB, empresaID, id int64) error {
	periodo, err := getPeriodoContableMovimiento(dbConn, empresaID, id)
	if err != nil {
		return err
	}
	cerrado, err := IsEmpresaFinanzasPeriodoCerrado(dbConn, empresaID, periodo)
	if err != nil {
		return err
	}
	if cerrado {
		return ErrPeriodoFinancieroCerrado
	}
	res, err := dbConn.Exec(`DELETE FROM empresa_finanzas_movimientos WHERE empresa_id = ? AND id = ?`, empresaID, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func ListEmpresaFinanzasPeriodos(dbConn *sql.DB, empresaID int64, includeInactive bool) ([]EmpresaFinanzasPeriodo, error) {
	query := `SELECT
		id,
		empresa_id,
		COALESCE(periodo, ''),
		COALESCE(fecha_inicio, ''),
		COALESCE(fecha_fin, ''),
		COALESCE(estado, 'abierto'),
		COALESCE(fecha_cierre, ''),
		COALESCE(cerrado_por, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(observaciones, '')
	FROM empresa_finanzas_periodos
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if !includeInactive {
		query += ` AND COALESCE(estado, 'abierto') <> 'inactivo'`
	}
	query += ` ORDER BY periodo DESC, id DESC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaFinanzasPeriodo, 0)
	for rows.Next() {
		var p EmpresaFinanzasPeriodo
		if err := rows.Scan(
			&p.ID,
			&p.EmpresaID,
			&p.Periodo,
			&p.FechaInicio,
			&p.FechaFin,
			&p.Estado,
			&p.FechaCierre,
			&p.CerradoPor,
			&p.FechaCreacion,
			&p.FechaActualizacion,
			&p.UsuarioCreador,
			&p.Observaciones,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func UpsertEmpresaFinanzasPeriodo(dbConn *sql.DB, p EmpresaFinanzasPeriodo) (int64, error) {
	if p.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	p.Periodo = normalizePeriodoContable(p.Periodo)
	if p.Periodo == "" {
		return 0, fmt.Errorf("periodo es obligatorio (YYYY-MM)")
	}
	p.Estado = normalizeEstadoPeriodo(p.Estado)
	if p.UsuarioCreador == "" {
		p.UsuarioCreador = "sistema"
	}
	if strings.TrimSpace(p.FechaInicio) == "" || strings.TrimSpace(p.FechaFin) == "" {
		ini, fin := periodRangeFromPeriodo(p.Periodo)
		if strings.TrimSpace(p.FechaInicio) == "" {
			p.FechaInicio = ini
		}
		if strings.TrimSpace(p.FechaFin) == "" {
			p.FechaFin = fin
		}
	}
	if p.Estado == "cerrado" {
		if strings.TrimSpace(p.FechaCierre) == "" {
			p.FechaCierre = time.Now().Format("2006-01-02 15:04:05")
		}
		if strings.TrimSpace(p.CerradoPor) == "" {
			p.CerradoPor = p.UsuarioCreador
		}
	} else if p.Estado == "abierto" {
		p.FechaCierre = ""
		p.CerradoPor = ""
	}

	res, err := dbConn.Exec(`INSERT INTO empresa_finanzas_periodos (
		empresa_id,
		periodo,
		fecha_inicio,
		fecha_fin,
		estado,
		fecha_cierre,
		cerrado_por,
		usuario_creador,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))
	ON CONFLICT(empresa_id, periodo) DO UPDATE SET
		fecha_inicio = excluded.fecha_inicio,
		fecha_fin = excluded.fecha_fin,
		estado = excluded.estado,
		fecha_cierre = excluded.fecha_cierre,
		cerrado_por = excluded.cerrado_por,
		usuario_creador = excluded.usuario_creador,
		observaciones = excluded.observaciones,
		fecha_actualizacion = datetime('now','localtime')`,
		p.EmpresaID,
		p.Periodo,
		p.FechaInicio,
		p.FechaFin,
		p.Estado,
		p.FechaCierre,
		p.CerradoPor,
		p.UsuarioCreador,
		strings.TrimSpace(p.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	if id, errID := res.LastInsertId(); errID == nil && id > 0 {
		return id, nil
	}
	var currentID int64
	if err := dbConn.QueryRow(`SELECT id FROM empresa_finanzas_periodos WHERE empresa_id = ? AND periodo = ? LIMIT 1`, p.EmpresaID, p.Periodo).Scan(&currentID); err != nil {
		return 0, err
	}
	return currentID, nil
}

func SetEmpresaFinanzasPeriodoEstado(dbConn *sql.DB, empresaID int64, periodo, estado, usuario, observaciones string) error {
	periodo = normalizePeriodoContable(periodo)
	if periodo == "" {
		return fmt.Errorf("periodo es obligatorio")
	}
	estado = normalizeEstadoPeriodo(estado)
	if usuario == "" {
		usuario = "sistema"
	}
	_, err := UpsertEmpresaFinanzasPeriodo(dbConn, EmpresaFinanzasPeriodo{
		EmpresaID:      empresaID,
		Periodo:        periodo,
		Estado:         estado,
		UsuarioCreador: usuario,
		CerradoPor:     usuario,
		Observaciones:  strings.TrimSpace(observaciones),
	})
	return err
}

func IsEmpresaFinanzasPeriodoCerrado(dbConn *sql.DB, empresaID int64, periodo string) (bool, error) {
	periodo = normalizePeriodoContable(periodo)
	if periodo == "" {
		return false, nil
	}
	var estado string
	err := dbConn.QueryRow(`SELECT COALESCE(estado, 'abierto') FROM empresa_finanzas_periodos WHERE empresa_id = ? AND periodo = ? LIMIT 1`, empresaID, periodo).Scan(&estado)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return normalizeEstadoPeriodo(estado) == "cerrado", nil
}

func CreateEmpresaCierreCaja(dbConn *sql.DB, cierre EmpresaCierreCaja) (int64, error) {
	cierre, err := normalizeEmpresaCierreCaja(cierre, true)
	if err != nil {
		return 0, err
	}

	if cierre.EstadoCierre == "cerrado" && strings.TrimSpace(cierre.FechaCierre) == "" {
		cierre.FechaCierre = time.Now().Format("2006-01-02 15:04:05")
	}
	if cierre.EstadoCierre == "cerrado" && strings.TrimSpace(cierre.CerradoPor) == "" {
		cierre.CerradoPor = cierre.UsuarioCreador
	}
	if cierre.EstadoCierre == "aprobado" {
		if strings.TrimSpace(cierre.FechaCierre) == "" {
			cierre.FechaCierre = time.Now().Format("2006-01-02 15:04:05")
		}
		if strings.TrimSpace(cierre.CerradoPor) == "" {
			cierre.CerradoPor = cierre.UsuarioCreador
		}
		if strings.TrimSpace(cierre.AprobadoPor) == "" {
			cierre.AprobadoPor = cierre.UsuarioCreador
		}
		if strings.TrimSpace(cierre.AprobadoEn) == "" {
			cierre.AprobadoEn = time.Now().Format("2006-01-02 15:04:05")
		}
	}

	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_cierres_caja (
		empresa_id, sucursal_id, caja_codigo, turno,
		fecha_operacion, fecha_apertura, fecha_cierre, estado_cierre,
		apertura_monto, ingresos_efectivo, egresos_efectivo, retiros_efectivo,
		caja_teorica, caja_fisica, diferencia_caja,
		moneda, cerrado_por, aprobado_por, aprobado_en,
		tiene_incidencia, umbral_incidencia,
		propinas_movimientos, propinas_total, propinas_ajustes, propinas_impuesto, propinas_neto,
		propinas_conciliado_en, propinas_conciliado_por,
		usuario_creador, estado, observaciones,
		fecha_creacion, fecha_actualizacion
	) VALUES (
		?, ?, ?, ?,
		?, ?, ?, ?,
		?, ?, ?, ?,
		?, ?, ?,
		?, ?, ?, ?,
		?, ?,
		?, ?, ?, ?, ?,
		?, ?,
		?, ?, ?,
		datetime('now','localtime'), datetime('now','localtime')
	)`,
		cierre.EmpresaID, cierre.SucursalID, cierre.CajaCodigo, cierre.Turno,
		cierre.FechaOperacion, cierre.FechaApertura, cierre.FechaCierre, cierre.EstadoCierre,
		cierre.AperturaMonto, cierre.IngresosEfectivo, cierre.EgresosEfectivo, cierre.RetirosEfectivo,
		cierre.CajaTeorica, cierre.CajaFisica, cierre.DiferenciaCaja,
		cierre.Moneda, cierre.CerradoPor, cierre.AprobadoPor, cierre.AprobadoEn,
		boolToInt(cierre.TieneIncidencia), cierre.UmbralIncidencia,
		cierre.PropinasMovimientos, cierre.PropinasTotal, cierre.PropinasAjustes, cierre.PropinasImpuesto, cierre.PropinasNeto,
		cierre.PropinasConciliadoEn, cierre.PropinasConciliadoPor,
		cierre.UsuarioCreador, cierre.Estado, cierre.Observaciones,
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func ListEmpresaCierresCaja(dbConn *sql.DB, empresaID int64, f EmpresaCierreCajaFilter) ([]EmpresaCierreCaja, error) {
	query := `SELECT
		id,
		empresa_id,
		COALESCE(sucursal_id, 0),
		COALESCE(caja_codigo, ''),
		COALESCE(turno, 'general'),
		COALESCE(fecha_operacion, ''),
		COALESCE(fecha_apertura, ''),
		COALESCE(fecha_cierre, ''),
		COALESCE(estado_cierre, 'abierto'),
		COALESCE(apertura_monto, 0),
		COALESCE(ingresos_efectivo, 0),
		COALESCE(egresos_efectivo, 0),
		COALESCE(retiros_efectivo, 0),
		COALESCE(caja_teorica, 0),
		COALESCE(caja_fisica, 0),
		COALESCE(diferencia_caja, 0),
		COALESCE(moneda, 'COP'),
		COALESCE(cerrado_por, ''),
		COALESCE(aprobado_por, ''),
		COALESCE(aprobado_en, ''),
		COALESCE(tiene_incidencia, 0),
		COALESCE(umbral_incidencia, 0),
		COALESCE(propinas_movimientos, 0),
		COALESCE(propinas_total, 0),
		COALESCE(propinas_ajustes, 0),
		COALESCE(propinas_impuesto, 0),
		COALESCE(propinas_neto, 0),
		COALESCE(propinas_conciliado_en, ''),
		COALESCE(propinas_conciliado_por, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_cierres_caja
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	if f.SucursalID > 0 {
		query += ` AND COALESCE(sucursal_id, 0) = ?`
		args = append(args, f.SucursalID)
	}
	if caja := sanitizeCajaCodigo(f.CajaCodigo); caja != "" {
		query += ` AND UPPER(COALESCE(caja_codigo, '')) = ?`
		args = append(args, caja)
	}
	if estadoCierre := normalizeEstadoCierre(f.EstadoCierre); estadoCierre != "" {
		query += ` AND LOWER(COALESCE(estado_cierre, 'abierto')) = ?`
		args = append(args, estadoCierre)
	}
	if strings.TrimSpace(f.Desde) != "" {
		query += ` AND date(fecha_operacion) >= date(?)`
		args = append(args, strings.TrimSpace(f.Desde))
	}
	if strings.TrimSpace(f.Hasta) != "" {
		query += ` AND date(fecha_operacion) <= date(?)`
		args = append(args, strings.TrimSpace(f.Hasta))
	}
	if !f.IncludeInactive {
		query += ` AND LOWER(COALESCE(estado, 'activo')) = 'activo'`
	}

	query += ` ORDER BY date(fecha_operacion) DESC, id DESC`
	limit := f.Limit
	if limit <= 0 {
		limit = 200
	}
	query += ` LIMIT ?`
	args = append(args, limit)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]EmpresaCierreCaja, 0)
	for rows.Next() {
		var item EmpresaCierreCaja
		var incidencia int
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.SucursalID,
			&item.CajaCodigo,
			&item.Turno,
			&item.FechaOperacion,
			&item.FechaApertura,
			&item.FechaCierre,
			&item.EstadoCierre,
			&item.AperturaMonto,
			&item.IngresosEfectivo,
			&item.EgresosEfectivo,
			&item.RetirosEfectivo,
			&item.CajaTeorica,
			&item.CajaFisica,
			&item.DiferenciaCaja,
			&item.Moneda,
			&item.CerradoPor,
			&item.AprobadoPor,
			&item.AprobadoEn,
			&incidencia,
			&item.UmbralIncidencia,
			&item.PropinasMovimientos,
			&item.PropinasTotal,
			&item.PropinasAjustes,
			&item.PropinasImpuesto,
			&item.PropinasNeto,
			&item.PropinasConciliadoEn,
			&item.PropinasConciliadoPor,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		item.TieneIncidencia = incidencia == 1
		items = append(items, item)
	}
	return items, rows.Err()
}

func UpdateEmpresaCierreCaja(dbConn *sql.DB, cierre EmpresaCierreCaja) error {
	var estadoActual string
	err := dbConn.QueryRow(`SELECT COALESCE(estado_cierre, 'abierto')
	FROM empresa_cierres_caja
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, cierre.EmpresaID, cierre.ID).Scan(&estadoActual)
	if err != nil {
		return err
	}
	if normalizeEstadoCierre(estadoActual) == "aprobado" {
		return ErrCierreCajaAprobadoBloqueado
	}

	cierre, err = normalizeEmpresaCierreCaja(cierre, false)
	if err != nil {
		return err
	}
	if cierre.EstadoCierre == "cerrado" && strings.TrimSpace(cierre.FechaCierre) == "" {
		cierre.FechaCierre = time.Now().Format("2006-01-02 15:04:05")
	}
	if cierre.EstadoCierre == "cerrado" && strings.TrimSpace(cierre.CerradoPor) == "" {
		cierre.CerradoPor = cierre.UsuarioCreador
	}
	if cierre.EstadoCierre == "aprobado" {
		if strings.TrimSpace(cierre.FechaCierre) == "" {
			cierre.FechaCierre = time.Now().Format("2006-01-02 15:04:05")
		}
		if strings.TrimSpace(cierre.CerradoPor) == "" {
			cierre.CerradoPor = cierre.UsuarioCreador
		}
		if strings.TrimSpace(cierre.AprobadoPor) == "" {
			cierre.AprobadoPor = cierre.UsuarioCreador
		}
		if strings.TrimSpace(cierre.AprobadoEn) == "" {
			cierre.AprobadoEn = time.Now().Format("2006-01-02 15:04:05")
		}
	}

	res, err := dbConn.Exec(`UPDATE empresa_cierres_caja SET
		sucursal_id = ?,
		caja_codigo = ?,
		turno = ?,
		fecha_operacion = ?,
		fecha_apertura = ?,
		fecha_cierre = ?,
		estado_cierre = ?,
		apertura_monto = ?,
		ingresos_efectivo = ?,
		egresos_efectivo = ?,
		retiros_efectivo = ?,
		caja_teorica = ?,
		caja_fisica = ?,
		diferencia_caja = ?,
		moneda = ?,
		cerrado_por = ?,
		aprobado_por = ?,
		aprobado_en = ?,
		tiene_incidencia = ?,
		umbral_incidencia = ?,
		propinas_movimientos = ?,
		propinas_total = ?,
		propinas_ajustes = ?,
		propinas_impuesto = ?,
		propinas_neto = ?,
		propinas_conciliado_en = ?,
		propinas_conciliado_por = ?,
		estado = ?,
		observaciones = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`,
		cierre.SucursalID,
		cierre.CajaCodigo,
		cierre.Turno,
		cierre.FechaOperacion,
		cierre.FechaApertura,
		cierre.FechaCierre,
		cierre.EstadoCierre,
		cierre.AperturaMonto,
		cierre.IngresosEfectivo,
		cierre.EgresosEfectivo,
		cierre.RetirosEfectivo,
		cierre.CajaTeorica,
		cierre.CajaFisica,
		cierre.DiferenciaCaja,
		cierre.Moneda,
		cierre.CerradoPor,
		cierre.AprobadoPor,
		cierre.AprobadoEn,
		boolToInt(cierre.TieneIncidencia),
		cierre.UmbralIncidencia,
		cierre.PropinasMovimientos,
		cierre.PropinasTotal,
		cierre.PropinasAjustes,
		cierre.PropinasImpuesto,
		cierre.PropinasNeto,
		cierre.PropinasConciliadoEn,
		cierre.PropinasConciliadoPor,
		cierre.Estado,
		cierre.Observaciones,
		cierre.EmpresaID,
		cierre.ID,
	)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func SetEmpresaCierreCajaEstado(dbConn *sql.DB, empresaID, id int64, estado string, cajaFisica *float64, usuario, observaciones string) error {
	estado = normalizeEstadoCierre(estado)
	if estado == "" {
		return fmt.Errorf("estado_cierre invalido")
	}
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = "sistema"
	}

	var actualEstado string
	var fechaCierreActual string
	var apertura float64
	var ingresos float64
	var egresos float64
	var retiros float64
	var cajaFisicaActual float64
	var umbral float64
	var cerradoPorActual string
	var aprobadoPorActual string
	var aprobadoEnActual string
	var observacionesActual string
	err := dbConn.QueryRow(`SELECT
		COALESCE(estado_cierre, 'abierto'),
		COALESCE(fecha_cierre, ''),
		COALESCE(apertura_monto, 0),
		COALESCE(ingresos_efectivo, 0),
		COALESCE(egresos_efectivo, 0),
		COALESCE(retiros_efectivo, 0),
		COALESCE(caja_fisica, 0),
		COALESCE(umbral_incidencia, 0),
		COALESCE(cerrado_por, ''),
		COALESCE(aprobado_por, ''),
		COALESCE(aprobado_en, ''),
		COALESCE(observaciones, '')
	FROM empresa_cierres_caja
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, id).Scan(
		&actualEstado,
		&fechaCierreActual,
		&apertura,
		&ingresos,
		&egresos,
		&retiros,
		&cajaFisicaActual,
		&umbral,
		&cerradoPorActual,
		&aprobadoPorActual,
		&aprobadoEnActual,
		&observacionesActual,
	)
	if err != nil {
		return err
	}

	actualEstado = normalizeEstadoCierre(actualEstado)
	if !isValidCierreCajaTransition(actualEstado, estado) {
		return ErrCierreCajaTransicionInvalida
	}

	if estado == "cerrado" && cajaFisica == nil {
		return fmt.Errorf("caja_fisica es obligatoria para cerrar caja")
	}
	if cajaFisica != nil {
		cajaFisicaActual = maxFloat64(*cajaFisica, 0)
	}

	cajaTeorica := calculateCajaTeorica(apertura, ingresos, egresos, retiros)
	diferencia := cajaTeorica - cajaFisicaActual
	tieneIncidencia := hasCajaIncidencia(diferencia, umbral)
	fechaCierre := fechaCierreActual
	cerradoPor := cerradoPorActual
	aprobadoPor := aprobadoPorActual
	aprobadoEn := aprobadoEnActual

	switch estado {
	case "abierto":
		fechaCierre = ""
		cerradoPor = ""
		aprobadoPor = ""
		aprobadoEn = ""
		cajaFisicaActual = 0
		diferencia = 0
		tieneIncidencia = false
	case "cerrado":
		fechaCierre = time.Now().Format("2006-01-02 15:04:05")
		cerradoPor = usuario
		aprobadoPor = ""
		aprobadoEn = ""
	case "aprobado":
		if strings.TrimSpace(fechaCierre) == "" {
			fechaCierre = time.Now().Format("2006-01-02 15:04:05")
		}
		if strings.TrimSpace(cerradoPor) == "" {
			cerradoPor = usuario
		}
		aprobadoPor = usuario
		aprobadoEn = time.Now().Format("2006-01-02 15:04:05")
	case "anulado":
		if strings.TrimSpace(fechaCierre) == "" {
			fechaCierre = time.Now().Format("2006-01-02 15:04:05")
		}
		if strings.TrimSpace(cerradoPor) == "" {
			cerradoPor = usuario
		}
	}

	obs := strings.TrimSpace(observaciones)
	if obs == "" {
		obs = observacionesActual
	}

	res, err := dbConn.Exec(`UPDATE empresa_cierres_caja SET
		estado_cierre = ?,
		fecha_cierre = ?,
		cerrado_por = ?,
		aprobado_por = ?,
		aprobado_en = ?,
		caja_teorica = ?,
		caja_fisica = ?,
		diferencia_caja = ?,
		tiene_incidencia = ?,
		observaciones = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`,
		estado,
		fechaCierre,
		cerradoPor,
		aprobadoPor,
		aprobadoEn,
		cajaTeorica,
		cajaFisicaActual,
		diferencia,
		boolToInt(tieneIncidencia),
		obs,
		empresaID,
		id,
	)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func SetEmpresaCierreCajaRegistroEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	estado = normalizeEstadoMovimiento(estado)
	res, err := dbConn.Exec(`UPDATE empresa_cierres_caja
	SET estado = ?, fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`, estado, empresaID, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func DeleteEmpresaCierreCaja(dbConn *sql.DB, empresaID, id int64) error {
	var estadoCierre string
	err := dbConn.QueryRow(`SELECT COALESCE(estado_cierre, 'abierto')
	FROM empresa_cierres_caja
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, id).Scan(&estadoCierre)
	if err != nil {
		return err
	}
	if normalizeEstadoCierre(estadoCierre) == "aprobado" {
		return ErrCierreCajaAprobadoBloqueado
	}

	res, err := dbConn.Exec(`DELETE FROM empresa_cierres_caja WHERE empresa_id = ? AND id = ?`, empresaID, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func isValidCierreCajaTransition(fromEstado, toEstado string) bool {
	fromEstado = normalizeEstadoCierre(fromEstado)
	toEstado = normalizeEstadoCierre(toEstado)
	if fromEstado == "" || toEstado == "" {
		return false
	}
	if fromEstado == toEstado {
		return true
	}
	switch fromEstado {
	case "abierto":
		return toEstado == "cerrado" || toEstado == "anulado"
	case "cerrado":
		return toEstado == "abierto" || toEstado == "aprobado" || toEstado == "anulado"
	case "aprobado":
		return toEstado == "abierto"
	case "anulado":
		return toEstado == "abierto"
	default:
		return false
	}
}

func getPeriodoContableMovimiento(dbConn *sql.DB, empresaID, id int64) (string, error) {
	var periodo string
	var fechaMovimiento string
	err := dbConn.QueryRow(`SELECT COALESCE(periodo_contable, ''), COALESCE(fecha_movimiento, '') FROM empresa_finanzas_movimientos WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, id).Scan(&periodo, &fechaMovimiento)
	if err != nil {
		return "", err
	}
	periodo = normalizePeriodoContable(periodo)
	if periodo == "" {
		periodo = normalizePeriodoContable(fechaMovimiento)
	}
	return periodo, nil
}

func ensurePeriodoAbiertoParaActualizarMovimiento(dbConn *sql.DB, empresaID, id int64, nuevoPeriodo string) error {
	periodoActual, err := getPeriodoContableMovimiento(dbConn, empresaID, id)
	if err != nil {
		return err
	}
	cerradoActual, err := IsEmpresaFinanzasPeriodoCerrado(dbConn, empresaID, periodoActual)
	if err != nil {
		return err
	}
	if cerradoActual {
		return ErrPeriodoFinancieroCerrado
	}
	nuevoPeriodo = normalizePeriodoContable(nuevoPeriodo)
	if nuevoPeriodo == "" || nuevoPeriodo == periodoActual {
		return nil
	}
	cerradoNuevo, err := IsEmpresaFinanzasPeriodoCerrado(dbConn, empresaID, nuevoPeriodo)
	if err != nil {
		return err
	}
	if cerradoNuevo {
		return ErrPeriodoFinancieroCerrado
	}
	return nil
}

func periodRangeFromPeriodo(periodo string) (string, string) {
	t, err := time.Parse("2006-01", periodo)
	if err != nil {
		return "", ""
	}
	inicio := t.Format("2006-01-02")
	fin := t.AddDate(0, 1, 0).Add(-time.Second).Format("2006-01-02")
	return inicio, fin
}

func normalizeEmpresaFinanzasMovimiento(dbConn *sql.DB, m EmpresaFinanzasMovimiento, isCreate bool) (EmpresaFinanzasMovimiento, error) {
	if m.EmpresaID <= 0 {
		return m, fmt.Errorf("empresa_id es obligatorio")
	}
	m.TipoMovimiento = normalizeTipoMovimiento(m.TipoMovimiento)
	if m.TipoMovimiento == "" {
		return m, fmt.Errorf("tipo_movimiento debe ser ingreso o egreso")
	}
	if m.Monto <= 0 {
		return m, fmt.Errorf("monto debe ser mayor que cero")
	}
	m.Impuesto = maxFloat64(m.Impuesto, 0)
	m.RetencionFuente = maxFloat64(m.RetencionFuente, 0)
	m.RetencionICA = maxFloat64(m.RetencionICA, 0)
	m.RetencionIVA = maxFloat64(m.RetencionIVA, 0)
	m.TotalRetenciones = m.RetencionFuente + m.RetencionICA + m.RetencionIVA
	m.Total = maxFloat64(m.Total, 0)
	if m.Total <= 0 {
		m.Total = m.Monto + m.Impuesto
	}
	if m.TotalRetenciones > m.Total {
		return m, fmt.Errorf("total_retenciones no puede superar el total")
	}
	m.TotalNeto = m.Total - m.TotalRetenciones
	m.Moneda = strings.ToUpper(strings.TrimSpace(m.Moneda))
	if m.Moneda == "" {
		m.Moneda = "COP"
	}
	m.Codigo = sanitizeFinancialCode(m.Codigo)
	if m.Codigo == "" {
		cfg, err := GetEmpresaFinanzasConfiguracion(dbConn, m.EmpresaID)
		if err != nil {
			return m, err
		}
		m.Codigo = buildDefaultFinancialCode(m.TipoMovimiento, cfg.PrefijoIngreso, cfg.PrefijoEgreso)
	}
	m.FechaMovimiento = strings.TrimSpace(m.FechaMovimiento)
	if m.FechaMovimiento == "" {
		m.FechaMovimiento = time.Now().Format("2006-01-02 15:04:05")
	}
	m.PeriodoContable = normalizePeriodoContable(m.PeriodoContable)
	if m.PeriodoContable == "" {
		m.PeriodoContable = normalizePeriodoContable(m.FechaMovimiento)
	}
	if m.PeriodoContable == "" {
		m.PeriodoContable = time.Now().Format("2006-01")
	}
	m.Categoria = strings.TrimSpace(m.Categoria)
	m.Subcategoria = strings.TrimSpace(m.Subcategoria)
	m.Concepto = strings.TrimSpace(m.Concepto)
	if m.Concepto == "" {
		return m, fmt.Errorf("concepto es obligatorio")
	}
	m.Descripcion = strings.TrimSpace(m.Descripcion)
	m.MetodoPago = strings.TrimSpace(m.MetodoPago)
	if m.MetodoPago == "" {
		m.MetodoPago = "efectivo"
	}
	m.TerceroNombre = strings.TrimSpace(m.TerceroNombre)
	m.TerceroDocumento = strings.TrimSpace(m.TerceroDocumento)
	m.TipoComprobante = strings.TrimSpace(m.TipoComprobante)
	if m.TipoComprobante == "" {
		m.TipoComprobante = "recibo_interno"
	}
	m.NumeroComprobante = strings.TrimSpace(m.NumeroComprobante)
	if m.NumeroComprobante == "" {
		m.NumeroComprobante = m.Codigo
	}
	m.ComprobanteURL = strings.TrimSpace(m.ComprobanteURL)
	m.ReferenciaExterna = strings.TrimSpace(m.ReferenciaExterna)
	m.AprobadoPor = strings.TrimSpace(m.AprobadoPor)
	m.UsuarioCreador = strings.TrimSpace(m.UsuarioCreador)
	if m.UsuarioCreador == "" {
		m.UsuarioCreador = "sistema"
	}
	m.Estado = normalizeEstadoMovimiento(m.Estado)
	m.Observaciones = strings.TrimSpace(m.Observaciones)
	if !isCreate && m.ID <= 0 {
		return m, fmt.Errorf("id es obligatorio")
	}
	return m, nil
}

func normalizeEmpresaCierreCaja(cierre EmpresaCierreCaja, isCreate bool) (EmpresaCierreCaja, error) {
	if cierre.EmpresaID <= 0 {
		return cierre, fmt.Errorf("empresa_id es obligatorio")
	}
	if !isCreate && cierre.ID <= 0 {
		return cierre, fmt.Errorf("id es obligatorio")
	}

	if cierre.SucursalID < 0 {
		cierre.SucursalID = 0
	}
	cierre.CajaCodigo = sanitizeCajaCodigo(cierre.CajaCodigo)
	if cierre.CajaCodigo == "" {
		return cierre, fmt.Errorf("caja_codigo es obligatorio")
	}
	cierre.Turno = strings.ToLower(strings.TrimSpace(cierre.Turno))
	if cierre.Turno == "" {
		cierre.Turno = "general"
	}

	cierre.FechaOperacion = normalizeDateOnly(cierre.FechaOperacion)
	if cierre.FechaOperacion == "" {
		cierre.FechaOperacion = time.Now().Format("2006-01-02")
	}
	cierre.FechaApertura = strings.TrimSpace(cierre.FechaApertura)
	if cierre.FechaApertura == "" {
		cierre.FechaApertura = time.Now().Format("2006-01-02 15:04:05")
	}
	cierre.FechaCierre = strings.TrimSpace(cierre.FechaCierre)

	cierre.EstadoCierre = normalizeEstadoCierre(cierre.EstadoCierre)
	if cierre.EstadoCierre == "" {
		cierre.EstadoCierre = "abierto"
	}

	cierre.AperturaMonto = maxFloat64(cierre.AperturaMonto, 0)
	cierre.IngresosEfectivo = maxFloat64(cierre.IngresosEfectivo, 0)
	cierre.EgresosEfectivo = maxFloat64(cierre.EgresosEfectivo, 0)
	cierre.RetirosEfectivo = maxFloat64(cierre.RetirosEfectivo, 0)
	cierre.CajaTeorica = calculateCajaTeorica(cierre.AperturaMonto, cierre.IngresosEfectivo, cierre.EgresosEfectivo, cierre.RetirosEfectivo)
	cierre.CajaFisica = maxFloat64(cierre.CajaFisica, 0)
	cierre.DiferenciaCaja = cierre.CajaTeorica - cierre.CajaFisica
	cierre.UmbralIncidencia = maxFloat64(cierre.UmbralIncidencia, 0)
	cierre.TieneIncidencia = hasCajaIncidencia(cierre.DiferenciaCaja, cierre.UmbralIncidencia)

	cierre.Moneda = strings.ToUpper(strings.TrimSpace(cierre.Moneda))
	if cierre.Moneda == "" {
		cierre.Moneda = "COP"
	}
	cierre.CerradoPor = strings.TrimSpace(cierre.CerradoPor)
	cierre.AprobadoPor = strings.TrimSpace(cierre.AprobadoPor)
	cierre.AprobadoEn = strings.TrimSpace(cierre.AprobadoEn)
	if cierre.PropinasMovimientos < 0 {
		cierre.PropinasMovimientos = 0
	}
	cierre.PropinasTotal = roundReportesMoney(cierre.PropinasTotal)
	cierre.PropinasAjustes = roundReportesMoney(cierre.PropinasAjustes)
	cierre.PropinasImpuesto = roundReportesMoney(cierre.PropinasImpuesto)
	cierre.PropinasNeto = roundReportesMoney(cierre.PropinasNeto)
	cierre.PropinasConciliadoEn = strings.TrimSpace(cierre.PropinasConciliadoEn)
	cierre.PropinasConciliadoPor = strings.TrimSpace(cierre.PropinasConciliadoPor)

	if cierre.EstadoCierre == "abierto" {
		cierre.FechaCierre = ""
		cierre.CerradoPor = ""
		cierre.AprobadoPor = ""
		cierre.AprobadoEn = ""
		cierre.CajaFisica = 0
		cierre.DiferenciaCaja = 0
		cierre.TieneIncidencia = false
		cierre.PropinasMovimientos = 0
		cierre.PropinasTotal = 0
		cierre.PropinasAjustes = 0
		cierre.PropinasImpuesto = 0
		cierre.PropinasNeto = 0
		cierre.PropinasConciliadoEn = ""
		cierre.PropinasConciliadoPor = ""
	}

	cierre.UsuarioCreador = strings.TrimSpace(cierre.UsuarioCreador)
	if cierre.UsuarioCreador == "" {
		cierre.UsuarioCreador = "sistema"
	}
	cierre.Estado = normalizeEstadoMovimiento(cierre.Estado)
	cierre.Observaciones = strings.TrimSpace(cierre.Observaciones)

	return cierre, nil
}

func normalizeEmpresaFinanzasConfiguracion(cfg EmpresaFinanzasConfiguracion) EmpresaFinanzasConfiguracion {
	cfg.HabilitarIngresos = cfg.HabilitarIngresos || !cfg.HabilitarEgresos
	cfg.HabilitarEgresos = cfg.HabilitarEgresos || !cfg.HabilitarIngresos
	cfg.Moneda = strings.ToUpper(strings.TrimSpace(cfg.Moneda))
	if cfg.Moneda == "" {
		cfg.Moneda = "COP"
	}
	cfg.CategoriasIngreso = strings.TrimSpace(cfg.CategoriasIngreso)
	if cfg.CategoriasIngreso == "" {
		cfg.CategoriasIngreso = "ventas\nservicios\nhabitaciones\nrestaurante\nbar\nlavanderia\npropinas\notros ingresos"
	}
	cfg.CategoriasEgreso = strings.TrimSpace(cfg.CategoriasEgreso)
	if cfg.CategoriasEgreso == "" {
		cfg.CategoriasEgreso = "compras\nnomina\nservicios publicos\narriendo\nmantenimiento\naseo y lavanderia\ncomisiones\nimpuestos\nbancos\notros gastos"
	}
	cfg.PrefijoIngreso = sanitizeFinancialCode(cfg.PrefijoIngreso)
	if cfg.PrefijoIngreso == "" {
		cfg.PrefijoIngreso = "ING"
	}
	cfg.PrefijoEgreso = sanitizeFinancialCode(cfg.PrefijoEgreso)
	if cfg.PrefijoEgreso == "" {
		cfg.PrefijoEgreso = "EGR"
	}
	cfg.FormatoImpresion = strings.ToLower(strings.TrimSpace(cfg.FormatoImpresion))
	if cfg.FormatoImpresion != "pos" {
		cfg.FormatoImpresion = "carta"
	}
	cfg.IntegracionContableDestino = normalizeIntegracionContableDestino(cfg.IntegracionContableDestino)
	cfg.CuentaCajaBancos = sanitizeContableAccount(cfg.CuentaCajaBancos)
	if cfg.CuentaCajaBancos == "" {
		cfg.CuentaCajaBancos = "110505"
	}
	cfg.CuentaIngresos = sanitizeContableAccount(cfg.CuentaIngresos)
	if cfg.CuentaIngresos == "" {
		cfg.CuentaIngresos = "413595"
	}
	cfg.CuentaIVAGenerado = sanitizeContableAccount(cfg.CuentaIVAGenerado)
	if cfg.CuentaIVAGenerado == "" {
		cfg.CuentaIVAGenerado = "240805"
	}
	cfg.CuentaGastos = sanitizeContableAccount(cfg.CuentaGastos)
	if cfg.CuentaGastos == "" {
		cfg.CuentaGastos = "519595"
	}
	cfg.CuentaIVADescontable = sanitizeContableAccount(cfg.CuentaIVADescontable)
	if cfg.CuentaIVADescontable == "" {
		cfg.CuentaIVADescontable = "240810"
	}
	cfg.CuentaRetencionesCobrar = sanitizeContableAccount(cfg.CuentaRetencionesCobrar)
	if cfg.CuentaRetencionesCobrar == "" {
		cfg.CuentaRetencionesCobrar = "135595"
	}
	cfg.CuentaRetencionesPagar = sanitizeContableAccount(cfg.CuentaRetencionesPagar)
	if cfg.CuentaRetencionesPagar == "" {
		cfg.CuentaRetencionesPagar = "236595"
	}
	cfg.CuentasIngresoCategoria = normalizeCuentaCategoriasMapping(cfg.CuentasIngresoCategoria)
	cfg.CuentasEgresoCategoria = normalizeCuentaCategoriasMapping(cfg.CuentasEgresoCategoria)
	cfg.UsuarioCreador = strings.TrimSpace(cfg.UsuarioCreador)
	if cfg.UsuarioCreador == "" {
		cfg.UsuarioCreador = "sistema"
	}
	cfg.Estado = normalizeEstadoMovimiento(cfg.Estado)
	cfg.Observaciones = strings.TrimSpace(cfg.Observaciones)
	return cfg
}

func defaultEmpresaFinanzasConfiguracion(empresaID int64) *EmpresaFinanzasConfiguracion {
	return &EmpresaFinanzasConfiguracion{
		EmpresaID:                  empresaID,
		HabilitarIngresos:          true,
		HabilitarEgresos:           true,
		Moneda:                     "COP",
		CategoriasIngreso:          "ventas\nservicios\nhabitaciones\nrestaurante\nbar\nlavanderia\npropinas\notros ingresos",
		CategoriasEgreso:           "compras\nnomina\nservicios publicos\narriendo\nmantenimiento\naseo y lavanderia\ncomisiones\nimpuestos\nbancos\notros gastos",
		PrefijoIngreso:             "ING",
		PrefijoEgreso:              "EGR",
		FormatoImpresion:           "carta",
		RequiereAprobacion:         false,
		IntegracionContableDestino: "generico",
		CuentaCajaBancos:           "110505",
		CuentaIngresos:             "413595",
		CuentaIVAGenerado:          "240805",
		CuentaGastos:               "519595",
		CuentaIVADescontable:       "240810",
		CuentaRetencionesCobrar:    "135595",
		CuentaRetencionesPagar:     "236595",
		CuentasIngresoCategoria:    "ventas=413595\nservicios=417595\nhabitaciones=414095\nrestaurante=413595\nbar=413595\nlavanderia=417595\npropinas=429595\notros ingresos=429595",
		CuentasEgresoCategoria:     "compras=613595\nnomina=510506\nservicios publicos=513595\narriendo=512001\nmantenimiento=514525\naseo y lavanderia=519595\ncomisiones=519520\nimpuestos=511505\nbancos=530505\notros gastos=519595",
		Estado:                     "activo",
	}
}

func normalizeIntegracionContableDestino(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "siigo", "world_office", "alegra", "helisa", "loggro", "contapyme":
		return v
	default:
		return "generico"
	}
}

func normalizeTipoMovimiento(tipo string) string {
	t := strings.ToLower(strings.TrimSpace(tipo))
	if t == "ingreso" || t == "egreso" {
		return t
	}
	return ""
}

func normalizeEstadoMovimiento(estado string) string {
	e := strings.ToLower(strings.TrimSpace(estado))
	if e == "inactivo" || e == "anulado" {
		return e
	}
	return "activo"
}

func normalizeEstadoPeriodo(estado string) string {
	e := strings.ToLower(strings.TrimSpace(estado))
	if e == "cerrado" || e == "inactivo" {
		return e
	}
	return "abierto"
}

func normalizeEstadoCierre(estado string) string {
	e := strings.ToLower(strings.TrimSpace(estado))
	switch e {
	case "abierto", "cerrado", "aprobado", "anulado":
		return e
	default:
		return ""
	}
}

func normalizePeriodoContable(v string) string {
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
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, v); err == nil {
			return t.Format("2006-01")
		}
	}
	return ""
}

func normalizeDateOnly(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if len(v) >= 10 {
		candidate := v[:10]
		if _, err := time.Parse("2006-01-02", candidate); err == nil {
			return candidate
		}
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, v); err == nil {
			return t.Format("2006-01-02")
		}
	}
	return ""
}

func calculateCajaTeorica(apertura, ingresos, egresos, retiros float64) float64 {
	return maxFloat64(apertura, 0) + maxFloat64(ingresos, 0) - maxFloat64(egresos, 0) - maxFloat64(retiros, 0)
}

func hasCajaIncidencia(diferencia, umbral float64) bool {
	umbral = maxFloat64(umbral, 0)
	if umbral == 0 {
		return math.Abs(diferencia) > 0
	}
	return math.Abs(diferencia) > umbral
}

func sanitizeFinancialCode(v string) string {
	v = strings.ToUpper(strings.TrimSpace(v))
	if v == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range v {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_' || r == '/':
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

func sanitizeCajaCodigo(v string) string {
	v = strings.ToUpper(strings.TrimSpace(v))
	if v == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range v {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_' || r == '/':
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

func sanitizeContableAccount(v string) string {
	v = strings.ToUpper(strings.TrimSpace(v))
	if v == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range v {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_' || r == '/' || r == '.':
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

func normalizeCuentaCategoriasMapping(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	lines := strings.Split(raw, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		l := strings.TrimSpace(line)
		if l == "" {
			continue
		}
		idx := strings.IndexAny(l, "=:")
		if idx <= 0 {
			continue
		}
		categoria := strings.ToLower(strings.TrimSpace(l[:idx]))
		cuenta := sanitizeContableAccount(strings.TrimSpace(l[idx+1:]))
		if categoria == "" || cuenta == "" {
			continue
		}
		out = append(out, categoria+"="+cuenta)
	}
	return strings.Join(out, "\n")
}

func buildDefaultFinancialCode(tipo, prefIngreso, prefEgreso string) string {
	prefix := sanitizeFinancialCode(prefEgreso)
	if tipo == "ingreso" {
		prefix = sanitizeFinancialCode(prefIngreso)
	}
	if prefix == "" {
		if tipo == "ingreso" {
			prefix = "ING"
		} else {
			prefix = "EGR"
		}
	}
	now := time.Now()
	suffix := fmt.Sprintf("%s%03d", now.Format("20060102150405"), now.UnixNano()%1000)
	return prefix + "-" + suffix
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func maxFloat64(v, min float64) float64 {
	if v < min {
		return min
	}
	return v
}

// EmpresaReportesTableroOperativo consolida KPI operativos por empresa.
type EmpresaReportesTableroOperativo struct {
	VentasCerradas      int64   `json:"ventas_cerradas"`
	VentasHoy           int64   `json:"ventas_hoy"`
	IngresosVentas      float64 `json:"ingresos_ventas"`
	TicketPromedio      float64 `json:"ticket_promedio"`
	ClientesActivos     int64   `json:"clientes_activos"`
	ProductosActivos    int64   `json:"productos_activos"`
	ProductosBajoMinimo int64   `json:"productos_bajo_minimo"`
	ComprasMovimientos  int64   `json:"compras_movimientos"`
	ComprasCosto        float64 `json:"compras_costo"`
}

// EmpresaReportesTableroFinanciero consolida KPI financieros por empresa.
type EmpresaReportesTableroFinanciero struct {
	MovimientosIngresos int64   `json:"movimientos_ingresos"`
	MovimientosEgresos  int64   `json:"movimientos_egresos"`
	Ingresos            float64 `json:"ingresos"`
	Egresos             float64 `json:"egresos"`
	Balance             float64 `json:"balance"`
	PeriodosAbiertos    int64   `json:"periodos_abiertos"`
	PeriodosCerrados    int64   `json:"periodos_cerrados"`
}

// EmpresaReportesTableroContable consolida KPI contables por empresa.
type EmpresaReportesTableroContable struct {
	EventosPendientes            int64   `json:"eventos_pendientes"`
	EventosProcesados            int64   `json:"eventos_procesados"`
	EventosTotal                 int64   `json:"eventos_total"`
	EventosMontoTotal            float64 `json:"eventos_monto_total"`
	AsientosGenerados            int64   `json:"asientos_generados"`
	AsientosMontoTotal           float64 `json:"asientos_monto_total"`
	DocumentosFacturacionActivos int64   `json:"documentos_facturacion_activos"`
	DocumentosComprasActivos     int64   `json:"documentos_compras_activos"`
}

// EmpresaReportesEstadoResultados consolida ingresos, gastos y utilidad operacional.
type EmpresaReportesEstadoResultados struct {
	Ingresos            float64 `json:"ingresos"`
	Gastos              float64 `json:"gastos"`
	UtilidadOperacional float64 `json:"utilidad_operacional"`
}

// EmpresaReportesBalanceGeneral consolida saldos principales de balance.
type EmpresaReportesBalanceGeneral struct {
	Activos            float64 `json:"activos"`
	Pasivos            float64 `json:"pasivos"`
	Patrimonio         float64 `json:"patrimonio"`
	ResultadoEjercicio float64 `json:"resultado_ejercicio"`
	Cuadre             float64 `json:"cuadre"`
}

// EmpresaReportesTableroResumen agrupa el tablero minimo financiero-operativo.
type EmpresaReportesTableroResumen struct {
	EmpresaID        int64                            `json:"empresa_id"`
	Desde            string                           `json:"desde"`
	Hasta            string                           `json:"hasta"`
	GeneradoEn       string                           `json:"generado_en"`
	Operativo        EmpresaReportesTableroOperativo  `json:"operativo"`
	Financiero       EmpresaReportesTableroFinanciero `json:"financiero"`
	Contable         EmpresaReportesTableroContable   `json:"contable"`
	EstadoResultados EmpresaReportesEstadoResultados  `json:"estado_resultados"`
	BalanceGeneral   EmpresaReportesBalanceGeneral    `json:"balance_general"`
}

// GetEmpresaReportesTableroResumen devuelve el tablero minimo financiero-operativo para una empresa.
func GetEmpresaReportesTableroResumen(dbConn *sql.DB, empresaID int64, desde, hasta string) (*EmpresaReportesTableroResumen, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}

	resumen := &EmpresaReportesTableroResumen{
		EmpresaID:  empresaID,
		Desde:      strings.TrimSpace(desde),
		Hasta:      strings.TrimSpace(hasta),
		GeneradoEn: time.Now().Format("2006-01-02 15:04:05"),
	}

	ventasCond, ventasArgs := buildDateRangeCondition("COALESCE(c.pagado_en, c.fecha_actualizacion, c.fecha_creacion)", resumen.Desde, resumen.Hasta)
	ventasQuery := `SELECT
		COALESCE(COUNT(1), 0),
		COALESCE(SUM(CASE WHEN COALESCE(c.total_pagado, 0) > 0 THEN COALESCE(c.total_pagado, 0) ELSE COALESCE(c.total, 0) END), 0),
		COALESCE(SUM(CASE WHEN date(COALESCE(c.pagado_en, c.fecha_actualizacion, c.fecha_creacion)) = CURRENT_DATE THEN 1 ELSE 0 END), 0)
	FROM carritos_compras c
	WHERE c.empresa_id = ?
		AND LOWER(COALESCE(c.estado_carrito, '')) = 'cerrado'
		AND LOWER(COALESCE(c.estado, 'activo')) = 'activo'` + ventasCond
	ventasParams := append([]interface{}{empresaID}, ventasArgs...)
	if err := queryRowSQLCompat(dbConn, ventasQuery, ventasParams...).Scan(
		&resumen.Operativo.VentasCerradas,
		&resumen.Operativo.IngresosVentas,
		&resumen.Operativo.VentasHoy,
	); err != nil {
		return nil, err
	}
	if resumen.Operativo.VentasCerradas > 0 {
		resumen.Operativo.TicketPromedio = resumen.Operativo.IngresosVentas / float64(resumen.Operativo.VentasCerradas)
	}

	if err := queryRowSQLCompat(dbConn, `SELECT COALESCE(COUNT(1), 0)
		FROM clientes
		WHERE empresa_id = ? AND LOWER(COALESCE(estado, 'activo')) = 'activo'`, empresaID).Scan(&resumen.Operativo.ClientesActivos); err != nil {
		return nil, err
	}

	if err := queryRowSQLCompat(dbConn, `SELECT
		COALESCE(COUNT(1), 0),
		COALESCE(SUM(CASE WHEN COALESCE(p.stock_minimo, 0) > 0 AND COALESCE(inv.stock_total, 0) <= COALESCE(p.stock_minimo, 0) THEN 1 ELSE 0 END), 0)
	FROM productos p
	LEFT JOIN (
		SELECT producto_id, COALESCE(SUM(COALESCE(cantidad, 0)), 0) AS stock_total
		FROM inventario_existencias
		WHERE empresa_id = ? AND LOWER(COALESCE(estado, 'activo')) = 'activo'
		GROUP BY producto_id
	) inv ON inv.producto_id = p.id
	WHERE p.empresa_id = ? AND LOWER(COALESCE(p.estado, 'activo')) = 'activo'`, empresaID, empresaID).Scan(&resumen.Operativo.ProductosActivos, &resumen.Operativo.ProductosBajoMinimo); err != nil {
		return nil, err
	}

	comprasCond, comprasArgs := buildDateRangeCondition("m.fecha_movimiento", resumen.Desde, resumen.Hasta)
	comprasQuery := `SELECT
		COALESCE(COUNT(1), 0),
		COALESCE(SUM(COALESCE(m.cantidad, 0) * COALESCE(m.costo_unitario, 0)), 0)
	FROM inventario_movimientos m
	WHERE m.empresa_id = ?
		AND LOWER(COALESCE(m.estado, 'activo')) = 'activo'
		AND LOWER(COALESCE(m.tipo, '')) IN ('entrada', 'ajuste_entrada', 'ajuste_positivo', 'compra')` + comprasCond
	comprasParams := append([]interface{}{empresaID}, comprasArgs...)
	if err := dbConn.QueryRow(comprasQuery, comprasParams...).Scan(&resumen.Operativo.ComprasMovimientos, &resumen.Operativo.ComprasCosto); err != nil {
		return nil, err
	}

	finanzasCond, finanzasArgs := buildDateRangeCondition("m.fecha_movimiento", resumen.Desde, resumen.Hasta)
	finanzasQuery := `SELECT
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(m.tipo_movimiento, '')) = 'ingreso' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(m.tipo_movimiento, '')) = 'egreso' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(m.tipo_movimiento, '')) = 'ingreso' THEN COALESCE(NULLIF(m.total_neto, 0), NULLIF(m.total, 0), m.monto, 0) ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(m.tipo_movimiento, '')) = 'egreso' THEN COALESCE(NULLIF(m.total_neto, 0), NULLIF(m.total, 0), m.monto, 0) ELSE 0 END), 0)
	FROM empresa_finanzas_movimientos m
	WHERE m.empresa_id = ?
		AND LOWER(COALESCE(m.estado, 'activo')) = 'activo'
		AND LOWER(COALESCE(m.tipo_movimiento, '')) IN ('ingreso', 'egreso')` + finanzasCond
	finanzasParams := append([]interface{}{empresaID}, finanzasArgs...)
	if err := queryRowSQLCompat(dbConn, finanzasQuery, finanzasParams...).Scan(
		&resumen.Financiero.MovimientosIngresos,
		&resumen.Financiero.MovimientosEgresos,
		&resumen.Financiero.Ingresos,
		&resumen.Financiero.Egresos,
	); err != nil {
		return nil, err
	}
	resumen.Financiero.Balance = resumen.Financiero.Ingresos - resumen.Financiero.Egresos

	if err := queryRowSQLCompat(dbConn, `SELECT
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(estado, 'abierto')) = 'abierto' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(estado, 'abierto')) = 'cerrado' THEN 1 ELSE 0 END), 0)
	FROM empresa_finanzas_periodos
	WHERE empresa_id = ? AND LOWER(COALESCE(estado, 'abierto')) <> 'inactivo'`, empresaID).Scan(
		&resumen.Financiero.PeriodosAbiertos,
		&resumen.Financiero.PeriodosCerrados,
	); err != nil {
		return nil, err
	}

	eventosCond, eventosArgs := buildDateRangeCondition("e.fecha_evento", resumen.Desde, resumen.Hasta)
	eventosQuery := `SELECT
		COALESCE(SUM(CASE WHEN COALESCE(e.procesado, 0) = 0 THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN COALESCE(e.procesado, 0) = 1 THEN 1 ELSE 0 END), 0),
		COALESCE(COUNT(1), 0),
		COALESCE(SUM(COALESCE(e.monto_total, 0)), 0)
	FROM empresa_eventos_contables e
	WHERE e.empresa_id = ?
		AND LOWER(COALESCE(e.estado, 'activo')) = 'activo'` + eventosCond
	eventosParams := append([]interface{}{empresaID}, eventosArgs...)
	if err := queryRowSQLCompat(dbConn, eventosQuery, eventosParams...).Scan(
		&resumen.Contable.EventosPendientes,
		&resumen.Contable.EventosProcesados,
		&resumen.Contable.EventosTotal,
		&resumen.Contable.EventosMontoTotal,
	); err != nil {
		return nil, err
	}

	if err := queryRowSQLCompat(dbConn, `SELECT COALESCE(COUNT(1), 0)
		FROM empresa_facturacion_documentos
		WHERE empresa_id = ? AND LOWER(COALESCE(estado, 'activo')) = 'activo'`, empresaID).Scan(&resumen.Contable.DocumentosFacturacionActivos); err != nil {
		return nil, err
	}
	if err := queryRowSQLCompat(dbConn, `SELECT COALESCE(COUNT(1), 0)
		FROM empresa_compras_documentos
		WHERE empresa_id = ? AND LOWER(COALESCE(estado, 'activo')) = 'activo'`, empresaID).Scan(&resumen.Contable.DocumentosComprasActivos); err != nil {
		return nil, err
	}

	estadoResultados, balanceGeneral, asientosGenerados, asientosMontoTotal, err := getEmpresaEstadosFinancierosDesdeAsientos(dbConn, empresaID, resumen.Desde, resumen.Hasta)
	if err != nil {
		return nil, err
	}
	if asientosGenerados == 0 {
		estadoResultados.Ingresos = roundReportesMoney(resumen.Financiero.Ingresos)
		estadoResultados.Gastos = roundReportesMoney(resumen.Financiero.Egresos)
		estadoResultados.UtilidadOperacional = roundReportesMoney(resumen.Financiero.Ingresos - resumen.Financiero.Egresos)

		if estadoResultados.UtilidadOperacional >= 0 {
			balanceGeneral.Activos = estadoResultados.UtilidadOperacional
			balanceGeneral.Pasivos = 0
			balanceGeneral.Patrimonio = estadoResultados.UtilidadOperacional
		} else {
			balanceGeneral.Activos = 0
			balanceGeneral.Pasivos = -estadoResultados.UtilidadOperacional
			balanceGeneral.Patrimonio = 0
		}
		balanceGeneral.ResultadoEjercicio = estadoResultados.UtilidadOperacional
		balanceGeneral.Cuadre = roundReportesMoney(balanceGeneral.Activos - (balanceGeneral.Pasivos + balanceGeneral.Patrimonio))
	}
	resumen.EstadoResultados = estadoResultados
	resumen.BalanceGeneral = balanceGeneral
	resumen.Contable.AsientosGenerados = asientosGenerados
	resumen.Contable.AsientosMontoTotal = asientosMontoTotal

	return resumen, nil
}

func buildDateRangeCondition(dateExpr, desde, hasta string) (string, []interface{}) {
	cond := ""
	args := make([]interface{}, 0, 2)
	if strings.TrimSpace(desde) != "" {
		cond += " AND date(" + dateExpr + ") >= date(?)"
		args = append(args, strings.TrimSpace(desde))
	}
	if strings.TrimSpace(hasta) != "" {
		cond += " AND date(" + dateExpr + ") <= date(?)"
		args = append(args, strings.TrimSpace(hasta))
	}
	return cond, args
}

func getEmpresaEstadosFinancierosDesdeAsientos(dbConn *sql.DB, empresaID int64, desde, hasta string) (EmpresaReportesEstadoResultados, EmpresaReportesBalanceGeneral, int64, float64, error) {
	estadoResultados := EmpresaReportesEstadoResultados{}
	balanceGeneral := EmpresaReportesBalanceGeneral{}

	cond, args := buildDateRangeCondition("a.fecha_asiento", desde, hasta)
	query := `SELECT
		COALESCE(lineas_json, '[]'),
		COALESCE(total_debito, 0),
		COALESCE(total_credito, 0)
	FROM empresa_asientos_contables a
	WHERE a.empresa_id = ?
		AND LOWER(COALESCE(a.estado, 'activo')) = 'activo'` + cond
	query += ` ORDER BY COALESCE(a.fecha_asiento, '') ASC, a.id ASC`
	params := append([]interface{}{empresaID}, args...)

	rows, err := querySQLCompat(dbConn, query, params...)
	if err != nil {
		return estadoResultados, balanceGeneral, 0, 0, err
	}
	defer rows.Close()

	var asientosGenerados int64
	asientosMontoTotal := 0.0
	activos := 0.0
	pasivos := 0.0
	patrimonioBase := 0.0
	ingresos := 0.0
	gastos := 0.0

	for rows.Next() {
		var lineasJSON string
		var totalDebito float64
		var totalCredito float64
		if err := rows.Scan(&lineasJSON, &totalDebito, &totalCredito); err != nil {
			return estadoResultados, balanceGeneral, 0, 0, err
		}
		asientosGenerados++
		if totalDebito > totalCredito {
			asientosMontoTotal += totalDebito
		} else {
			asientosMontoTotal += totalCredito
		}

		var lineas []EmpresaAsientoContableLinea
		if err := json.Unmarshal([]byte(strings.TrimSpace(lineasJSON)), &lineas); err != nil {
			continue
		}
		for _, ln := range lineas {
			cuenta := strings.TrimSpace(ln.Cuenta)
			if cuenta == "" {
				continue
			}
			delta := maxFloat64(ln.Debito, 0) - maxFloat64(ln.Credito, 0)
			switch cuenta[0] {
			case '1':
				activos += delta
			case '2':
				pasivos += -delta
			case '3':
				patrimonioBase += -delta
			case '4':
				ingresos += -delta
			case '5', '6', '7':
				gastos += delta
			}
		}
	}
	if err := rows.Err(); err != nil {
		return estadoResultados, balanceGeneral, 0, 0, err
	}

	utilidad := ingresos - gastos
	estadoResultados.Ingresos = roundReportesMoney(ingresos)
	estadoResultados.Gastos = roundReportesMoney(gastos)
	estadoResultados.UtilidadOperacional = roundReportesMoney(utilidad)

	balanceGeneral.Activos = roundReportesMoney(activos)
	balanceGeneral.Pasivos = roundReportesMoney(pasivos)
	balanceGeneral.ResultadoEjercicio = roundReportesMoney(utilidad)
	balanceGeneral.Patrimonio = roundReportesMoney(patrimonioBase + utilidad)
	balanceGeneral.Cuadre = roundReportesMoney(balanceGeneral.Activos - (balanceGeneral.Pasivos + balanceGeneral.Patrimonio))

	// Fallback mínimo cuando aún no hay asientos generados en el rango.
	if asientosGenerados == 0 {
		balance := roundReportesMoney(estadoResultados.UtilidadOperacional)
		if balance > 0 {
			balanceGeneral.Activos = balance
			balanceGeneral.Pasivos = 0
			balanceGeneral.Patrimonio = balance
			balanceGeneral.Cuadre = 0
		} else if balance < 0 {
			balanceGeneral.Activos = 0
			balanceGeneral.Pasivos = -balance
			balanceGeneral.Patrimonio = 0
			balanceGeneral.Cuadre = roundReportesMoney(balanceGeneral.Activos - (balanceGeneral.Pasivos + balanceGeneral.Patrimonio))
		}
	}

	return estadoResultados, balanceGeneral, asientosGenerados, roundReportesMoney(asientosMontoTotal), nil
}

func roundReportesMoney(v float64) float64 {
	return math.Round(v*100) / 100
}
