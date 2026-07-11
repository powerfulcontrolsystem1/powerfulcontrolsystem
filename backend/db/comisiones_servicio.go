package db

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	EmpresaComisionServicioOrigenVenta        = "venta"
	EmpresaComisionServicioOrigenAjusteManual = "ajuste_manual"

	EmpresaComisionServicioAjustePendiente = "pendiente"
	EmpresaComisionServicioAjusteAprobado  = "aprobado"
	EmpresaComisionServicioAjusteRechazado = "rechazado"
)

// EmpresaComisionesServicioConfiguracion define el comportamiento de comisiones por servicio.
type EmpresaComisionesServicioConfiguracion struct {
	ID                     int64   `json:"id"`
	EmpresaID              int64   `json:"empresa_id"`
	HabilitarComisiones    bool    `json:"habilitar_comisiones"`
	PorcentajeComision     float64 `json:"porcentaje_comision"`
	FiltroServicio         string  `json:"filtro_servicio,omitempty"`
	AplicarAutomaticamente bool    `json:"aplicar_automaticamente"`
	FechaCreacion          string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion     string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador         string  `json:"usuario_creador,omitempty"`
	Estado                 string  `json:"estado,omitempty"`
	Observaciones          string  `json:"observaciones,omitempty"`
}

// EmpresaComisionServicioEscala define reglas de porcentaje/tope por rol y servicio.
type EmpresaComisionServicioEscala struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	RolOperacion       string  `json:"rol_operacion,omitempty"`
	ServicioFiltro     string  `json:"servicio_filtro,omitempty"`
	PorcentajeComision float64 `json:"porcentaje_comision"`
	TopeComision       float64 `json:"tope_comision"`
	Prioridad          int     `json:"prioridad"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// EmpresaComisionServicioMovimiento representa una comision calculada sobre un item de servicio.
type EmpresaComisionServicioMovimiento struct {
	ID                      int64   `json:"id"`
	EmpresaID               int64   `json:"empresa_id"`
	CarritoID               int64   `json:"carrito_id,omitempty"`
	CarritoItemID           int64   `json:"carrito_item_id,omitempty"`
	ServicioID              int64   `json:"servicio_id,omitempty"`
	ServicioCodigo          string  `json:"servicio_codigo,omitempty"`
	ServicioNombre          string  `json:"servicio_nombre,omitempty"`
	ServicioCategoria       string  `json:"servicio_categoria,omitempty"`
	UsuarioOrigen           string  `json:"usuario_origen,omitempty"`
	UsuarioOrigenID         int64   `json:"usuario_origen_id,omitempty"`
	UsuarioLavador          string  `json:"usuario_lavador,omitempty"`
	UsuarioLavadorID        int64   `json:"usuario_lavador_id,omitempty"`
	RolOperacion            string  `json:"rol_operacion,omitempty"`
	EscalaID                int64   `json:"escala_id,omitempty"`
	VentaReferencia         string  `json:"venta_referencia,omitempty"`
	Moneda                  string  `json:"moneda,omitempty"`
	BaseServicio            float64 `json:"base_servicio"`
	PorcentajeComision      float64 `json:"porcentaje_comision"`
	MontoComisionBruto      float64 `json:"monto_comision_bruto"`
	TopeComisionAplicado    float64 `json:"tope_comision_aplicado"`
	MontoComision           float64 `json:"monto_comision"`
	OrigenMovimiento        string  `json:"origen_movimiento"`
	EsAjusteManual          bool    `json:"es_ajuste_manual"`
	ReferenciaAjuste        string  `json:"referencia_ajuste,omitempty"`
	AjusteEstado            string  `json:"ajuste_estado"`
	AprobadoPor             string  `json:"aprobado_por,omitempty"`
	AprobadoEn              string  `json:"aprobado_en,omitempty"`
	LiquidacionNominaID     int64   `json:"liquidacion_nomina_id,omitempty"`
	PeriodoLiquidacionDesde string  `json:"periodo_liquidacion_desde,omitempty"`
	PeriodoLiquidacionHasta string  `json:"periodo_liquidacion_hasta,omitempty"`
	LiquidadoEn             string  `json:"liquidado_en,omitempty"`
	LiquidadoPor            string  `json:"liquidado_por,omitempty"`
	FechaMovimiento         string  `json:"fecha_movimiento,omitempty"`
	FechaCreacion           string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion      string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador          string  `json:"usuario_creador,omitempty"`
	Estado                  string  `json:"estado,omitempty"`
	Observaciones           string  `json:"observaciones,omitempty"`
}

// EmpresaComisionServicioMovimientoFilter filtra movimientos de comisiones por servicio.
type EmpresaComisionServicioMovimientoFilter struct {
	MovimientoID        int64
	Desde               string
	Hasta               string
	UsuarioLavador      string
	RolOperacion        string
	ServicioFiltro      string
	OrigenMovimiento    string
	AjusteEstado        string
	LiquidacionNominaID int64
	SoloAjustes         bool
	SoloPendientes      bool
	NoLiquidado         bool
	IncludeInactive     bool
	Limit               int
}

// EmpresaComisionesServicioResumen consolida metricas de comisiones por servicio.
type EmpresaComisionesServicioResumen struct {
	TotalBaseServicios      float64 `json:"total_base_servicios"`
	TotalComisiones         float64 `json:"total_comisiones"`
	TotalAjustesManuales    float64 `json:"total_ajustes_manuales"`
	TotalLiquidadas         float64 `json:"total_liquidadas"`
	TotalPendientesLiquidar float64 `json:"total_pendientes_liquidar"`
	CantidadMovimientos     int64   `json:"cantidad_movimientos"`
	LavadoresConComision    int64   `json:"lavadores_con_comision"`
	PendientesAprobacion    int64   `json:"pendientes_aprobacion"`
}

// EmpresaComisionServicioLavadorResumen presenta acumulado por lavador.
type EmpresaComisionServicioLavadorResumen struct {
	UsuarioID           int64   `json:"usuario_id,omitempty"`
	UsuarioLavador      string  `json:"usuario_lavador"`
	TotalBaseServicios  float64 `json:"total_base_servicios"`
	TotalComision       float64 `json:"total_comision"`
	CantidadMovimientos int64   `json:"cantidad_movimientos"`
}

// EmpresaComisionesServicioReporte devuelve configuracion, resumen y detalle de comisiones.
type EmpresaComisionesServicioReporte struct {
	EmpresaID     int64                                   `json:"empresa_id"`
	Desde         string                                  `json:"desde,omitempty"`
	Hasta         string                                  `json:"hasta,omitempty"`
	Configuracion *EmpresaComisionesServicioConfiguracion `json:"configuracion"`
	Escalas       []EmpresaComisionServicioEscala         `json:"escalas"`
	Resumen       EmpresaComisionesServicioResumen        `json:"resumen"`
	Lavadores     []EmpresaComisionServicioLavadorResumen `json:"lavadores"`
	Movimientos   []EmpresaComisionServicioMovimiento     `json:"movimientos"`
}

// EmpresaComisionServicioRegistroResultado resume el registro automatico al cerrar venta.
type EmpresaComisionServicioRegistroResultado struct {
	Aplicada               bool    `json:"aplicada"`
	Habilitada             bool    `json:"habilitada"`
	AplicacionAutomatica   bool    `json:"aplicacion_automatica"`
	PorcentajeComision     float64 `json:"porcentaje_comision"`
	FiltroServicio         string  `json:"filtro_servicio,omitempty"`
	UsuarioLavador         string  `json:"usuario_lavador,omitempty"`
	UsuarioLavadorID       int64   `json:"usuario_lavador_id,omitempty"`
	RolOperacion           string  `json:"rol_operacion,omitempty"`
	BaseServicios          float64 `json:"base_servicios"`
	MontoComision          float64 `json:"monto_comision"`
	TotalTopesAplicados    float64 `json:"total_topes_aplicados"`
	MovimientosRegistrados int     `json:"movimientos_registrados"`
	RegistroIDs            []int64 `json:"registro_ids,omitempty"`
	EscalasAplicadas       []int64 `json:"escalas_aplicadas,omitempty"`
	Warning                string  `json:"warning,omitempty"`
}

// EmpresaComisionServicioLiquidacionResumen resume comisiones por periodo para vincular a nomina.
type EmpresaComisionServicioLiquidacionResumen struct {
	EmpresaID            int64    `json:"empresa_id"`
	PeriodoDesde         string   `json:"periodo_desde"`
	PeriodoHasta         string   `json:"periodo_hasta"`
	Identificadores      []string `json:"identificadores,omitempty"`
	CantidadMovimientos  int64    `json:"cantidad_movimientos"`
	TotalBaseServicios   float64  `json:"total_base_servicios"`
	TotalComisiones      float64  `json:"total_comisiones"`
	TotalAjustesManuales float64  `json:"total_ajustes_manuales"`
	MovimientoIDs        []int64  `json:"movimiento_ids,omitempty"`
}

type comisionServicioItemSnapshot struct {
	CarritoItemID     int64
	ServicioID        int64
	ServicioCodigo    string
	ServicioNombre    string
	ServicioCategoria string
	CodigoItem        string
	Descripcion       string
	TotalLinea        float64
}

// EnsureEmpresaComisionesServicioSchema crea/migra tablas de comisiones por servicio.
func EnsureEmpresaComisionesServicioSchema(dbConn *sql.DB) error {
	bootstrapStmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_comisiones_servicio_configuracion (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL UNIQUE,
			habilitar_comisiones INTEGER DEFAULT 0,
			porcentaje_comision REAL DEFAULT 10,
			filtro_servicio TEXT DEFAULT 'lavado',
			aplicar_automaticamente INTEGER DEFAULT 1,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS empresa_comisiones_servicio_escalas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			rol_operacion TEXT,
			servicio_filtro TEXT,
			porcentaje_comision REAL DEFAULT 0,
			tope_comision REAL DEFAULT 0,
			prioridad INTEGER DEFAULT 100,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS empresa_comisiones_servicio_movimientos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			carrito_id INTEGER DEFAULT 0,
			carrito_item_id INTEGER DEFAULT 0,
			servicio_id INTEGER DEFAULT 0,
			servicio_codigo TEXT,
			servicio_nombre TEXT,
			servicio_categoria TEXT,
			usuario_origen TEXT,
			usuario_origen_id INTEGER DEFAULT 0,
			usuario_lavador TEXT,
			usuario_lavador_id INTEGER DEFAULT 0,
			rol_operacion TEXT,
			escala_id INTEGER DEFAULT 0,
			venta_referencia TEXT,
			moneda TEXT DEFAULT 'COP',
			base_servicio REAL DEFAULT 0,
			porcentaje_comision REAL DEFAULT 0,
			monto_comision_bruto REAL DEFAULT 0,
			tope_comision_aplicado REAL DEFAULT 0,
			monto_comision REAL DEFAULT 0,
			origen_movimiento TEXT DEFAULT 'venta',
			ajuste_manual INTEGER DEFAULT 0,
			referencia_ajuste TEXT,
			ajuste_estado TEXT DEFAULT 'aprobado',
			aprobado_por TEXT,
			aprobado_en TEXT,
			liquidacion_nomina_id INTEGER DEFAULT 0,
			periodo_liquidacion_desde TEXT,
			periodo_liquidacion_hasta TEXT,
			liquidado_en TEXT,
			liquidado_por TEXT,
			fecha_movimiento TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
	}
	for _, stmt := range bootstrapStmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_configuracion", "habilitar_comisiones", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_configuracion", "porcentaje_comision", "REAL DEFAULT 10"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_configuracion", "filtro_servicio", "TEXT DEFAULT 'lavado'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_configuracion", "aplicar_automaticamente", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_configuracion", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_configuracion", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_configuracion", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_configuracion", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_escalas", "empresa_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_escalas", "rol_operacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_escalas", "servicio_filtro", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_escalas", "porcentaje_comision", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_escalas", "tope_comision", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_escalas", "prioridad", "INTEGER DEFAULT 100"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_escalas", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_escalas", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_escalas", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_escalas", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "carrito_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "carrito_item_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "servicio_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "servicio_codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "servicio_nombre", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "servicio_categoria", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "usuario_origen", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "usuario_origen_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "usuario_lavador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "usuario_lavador_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "rol_operacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "escala_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "venta_referencia", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "moneda", "TEXT DEFAULT 'COP'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "base_servicio", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "porcentaje_comision", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "monto_comision_bruto", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "tope_comision_aplicado", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "monto_comision", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "origen_movimiento", "TEXT DEFAULT 'venta'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "ajuste_manual", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "referencia_ajuste", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "ajuste_estado", "TEXT DEFAULT 'aprobado'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "aprobado_por", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "aprobado_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "liquidacion_nomina_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "periodo_liquidacion_desde", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "periodo_liquidacion_hasta", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "liquidado_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "liquidado_por", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "fecha_movimiento", "TEXT DEFAULT (CURRENT_TIMESTAMP)"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "observaciones", "TEXT"); err != nil {
		return err
	}

	indexStmts := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_comisiones_servicio_cfg_empresa ON empresa_comisiones_servicio_configuracion(empresa_id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_comisiones_servicio_cfg_estado ON empresa_comisiones_servicio_configuracion(empresa_id, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_comisiones_servicio_escala_empresa_prioridad ON empresa_comisiones_servicio_escalas(empresa_id, prioridad ASC, id ASC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_comisiones_servicio_escala_empresa_rol ON empresa_comisiones_servicio_escalas(empresa_id, rol_operacion, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_comisiones_servicio_mov_empresa_fecha ON empresa_comisiones_servicio_movimientos(empresa_id, fecha_movimiento DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_comisiones_servicio_mov_empresa_lavador ON empresa_comisiones_servicio_movimientos(empresa_id, usuario_lavador);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_comisiones_servicio_mov_empresa_lavador_id ON empresa_comisiones_servicio_movimientos(empresa_id, usuario_lavador_id, usuario_origen_id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_comisiones_servicio_mov_empresa_ajuste ON empresa_comisiones_servicio_movimientos(empresa_id, ajuste_manual, ajuste_estado, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_comisiones_servicio_mov_empresa_liquidacion ON empresa_comisiones_servicio_movimientos(empresa_id, liquidacion_nomina_id, periodo_liquidacion_desde, periodo_liquidacion_hasta);`,
		`DROP INDEX IF EXISTS ux_empresa_comisiones_servicio_mov_item_lavador;`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_comisiones_servicio_mov_item_lavador ON empresa_comisiones_servicio_movimientos(empresa_id, carrito_item_id, usuario_lavador)
			WHERE COALESCE(origen_movimiento, 'venta') = 'venta' AND COALESCE(carrito_item_id, 0) > 0;`,
	}
	for _, stmt := range indexStmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}

func defaultEmpresaComisionesServicioConfiguracion(empresaID int64) EmpresaComisionesServicioConfiguracion {
	return EmpresaComisionesServicioConfiguracion{
		EmpresaID:              empresaID,
		HabilitarComisiones:    false,
		PorcentajeComision:     10,
		FiltroServicio:         "lavado",
		AplicarAutomaticamente: true,
		Estado:                 "activo",
	}
}

func normalizeComisionPorcentaje(v float64) float64 {
	if v < 0 {
		v = 0
	}
	if v > 100 {
		v = 100
	}
	return round2(v)
}

func normalizeComisionMoneda(v string) string {
	m := strings.ToUpper(strings.TrimSpace(v))
	if m == "" {
		return "COP"
	}
	return m
}

func normalizeComisionFiltro(v string) string {
	return strings.TrimSpace(strings.ToLower(v))
}

func normalizeComisionRol(v string) string {
	return strings.TrimSpace(strings.ToLower(v))
}

func normalizeComisionTope(v float64) float64 {
	if v < 0 {
		v = 0
	}
	return round2(v)
}

func normalizeComisionPrioridad(v int) int {
	if v <= 0 {
		return 100
	}
	if v > 9999 {
		return 9999
	}
	return v
}

func normalizeComisionEscalaEstado(v string) string {
	if strings.ToLower(strings.TrimSpace(v)) == "inactivo" {
		return "inactivo"
	}
	return "activo"
}

func normalizeComisionEstado(v string) string {
	state := strings.ToLower(strings.TrimSpace(v))
	switch state {
	case "inactivo":
		return "inactivo"
	case "pendiente":
		return "pendiente"
	default:
		return "activo"
	}
}

func normalizeComisionOrigen(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case EmpresaComisionServicioOrigenAjusteManual, "ajuste":
		return EmpresaComisionServicioOrigenAjusteManual
	default:
		return EmpresaComisionServicioOrigenVenta
	}
}

func normalizeComisionAjusteEstado(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case EmpresaComisionServicioAjustePendiente:
		return EmpresaComisionServicioAjustePendiente
	case EmpresaComisionServicioAjusteRechazado:
		return EmpresaComisionServicioAjusteRechazado
	default:
		return EmpresaComisionServicioAjusteAprobado
	}
}

func defaultComisionLavador(usuarioLavador, usuarioOrigen string) string {
	user := strings.TrimSpace(usuarioLavador)
	if user != "" {
		return user
	}
	user = strings.TrimSpace(usuarioOrigen)
	if user != "" {
		return user
	}
	return "sistema"
}

// BuildEmpresaComisionServicioAliases normaliza identificadores para vincular comisiones con liquidaciones.
func BuildEmpresaComisionServicioAliases(values ...string) []string {
	seen := make(map[string]struct{})
	for _, raw := range values {
		v := strings.ToLower(strings.TrimSpace(raw))
		if v == "" {
			continue
		}
		seen[v] = struct{}{}
		collapsed := strings.Join(strings.Fields(v), " ")
		if collapsed != "" {
			seen[collapsed] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// GetEmpresaComisionServicioEscalaByID obtiene una escala por id y empresa.
func GetEmpresaComisionServicioEscalaByID(dbConn *sql.DB, empresaID, escalaID int64) (*EmpresaComisionServicioEscala, error) {
	if empresaID <= 0 || escalaID <= 0 {
		return nil, fmt.Errorf("empresa_id y escala_id son obligatorios")
	}

	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(rol_operacion, ''),
		COALESCE(servicio_filtro, ''),
		COALESCE(porcentaje_comision, 0),
		COALESCE(tope_comision, 0),
		COALESCE(prioridad, 100),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_comisiones_servicio_escalas
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, escalaID)

	item := EmpresaComisionServicioEscala{}
	if err := row.Scan(
		&item.ID,
		&item.EmpresaID,
		&item.RolOperacion,
		&item.ServicioFiltro,
		&item.PorcentajeComision,
		&item.TopeComision,
		&item.Prioridad,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	); err != nil {
		return nil, err
	}

	item.RolOperacion = normalizeComisionRol(item.RolOperacion)
	item.ServicioFiltro = normalizeComisionFiltro(item.ServicioFiltro)
	item.PorcentajeComision = normalizeComisionPorcentaje(item.PorcentajeComision)
	item.TopeComision = normalizeComisionTope(item.TopeComision)
	item.Prioridad = normalizeComisionPrioridad(item.Prioridad)
	item.Estado = normalizeComisionEscalaEstado(item.Estado)

	return &item, nil
}

// ListEmpresaComisionServicioEscalas lista escalas por empresa y filtro opcional de rol.
func ListEmpresaComisionServicioEscalas(dbConn *sql.DB, empresaID int64, includeInactive bool, rolOperacion string, limit int) ([]EmpresaComisionServicioEscala, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if limit <= 0 {
		limit = 500
	}
	if limit > 3000 {
		limit = 3000
	}

	clauses := []string{"empresa_id = ?"}
	args := []interface{}{empresaID}

	if !includeInactive {
		clauses = append(clauses, "COALESCE(estado, 'activo') = 'activo'")
	}
	if rol := normalizeComisionRol(rolOperacion); rol != "" {
		clauses = append(clauses, "(lower(COALESCE(rol_operacion, '')) = ? OR trim(COALESCE(rol_operacion, '')) = '')")
		args = append(args, rol)
	}

	query := `SELECT
		id,
		empresa_id,
		COALESCE(rol_operacion, ''),
		COALESCE(servicio_filtro, ''),
		COALESCE(porcentaje_comision, 0),
		COALESCE(tope_comision, 0),
		COALESCE(prioridad, 100),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_comisiones_servicio_escalas
	WHERE ` + strings.Join(clauses, " AND ") + `
	ORDER BY prioridad ASC, id ASC
	LIMIT ?`
	args = append(args, limit)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaComisionServicioEscala, 0)
	for rows.Next() {
		var row EmpresaComisionServicioEscala
		if err := rows.Scan(
			&row.ID,
			&row.EmpresaID,
			&row.RolOperacion,
			&row.ServicioFiltro,
			&row.PorcentajeComision,
			&row.TopeComision,
			&row.Prioridad,
			&row.FechaCreacion,
			&row.FechaActualizacion,
			&row.UsuarioCreador,
			&row.Estado,
			&row.Observaciones,
		); err != nil {
			return nil, err
		}
		row.RolOperacion = normalizeComisionRol(row.RolOperacion)
		row.ServicioFiltro = normalizeComisionFiltro(row.ServicioFiltro)
		row.PorcentajeComision = normalizeComisionPorcentaje(row.PorcentajeComision)
		row.TopeComision = normalizeComisionTope(row.TopeComision)
		row.Prioridad = normalizeComisionPrioridad(row.Prioridad)
		row.Estado = normalizeComisionEscalaEstado(row.Estado)
		out = append(out, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

// CreateEmpresaComisionServicioEscala registra una nueva escala de comision por servicio/rol.
func CreateEmpresaComisionServicioEscala(dbConn *sql.DB, payload EmpresaComisionServicioEscala) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	payload.RolOperacion = normalizeComisionRol(payload.RolOperacion)
	payload.ServicioFiltro = normalizeComisionFiltro(payload.ServicioFiltro)
	payload.PorcentajeComision = normalizeComisionPorcentaje(payload.PorcentajeComision)
	payload.TopeComision = normalizeComisionTope(payload.TopeComision)
	payload.Prioridad = normalizeComisionPrioridad(payload.Prioridad)
	payload.Estado = normalizeComisionEscalaEstado(payload.Estado)
	payload.UsuarioCreador = strings.TrimSpace(payload.UsuarioCreador)
	payload.Observaciones = strings.TrimSpace(payload.Observaciones)

	if payload.PorcentajeComision <= 0 {
		return 0, fmt.Errorf("porcentaje_comision debe ser mayor a cero")
	}

	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_comisiones_servicio_escalas (
		empresa_id,
		rol_operacion,
		servicio_filtro,
		porcentaje_comision,
		tope_comision,
		prioridad,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		payload.EmpresaID,
		payload.RolOperacion,
		payload.ServicioFiltro,
		payload.PorcentajeComision,
		payload.TopeComision,
		payload.Prioridad,
		payload.UsuarioCreador,
		payload.Estado,
		payload.Observaciones,
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateEmpresaComisionServicioEscala actualiza una escala existente.
func UpdateEmpresaComisionServicioEscala(dbConn *sql.DB, payload EmpresaComisionServicioEscala) (int64, error) {
	if payload.EmpresaID <= 0 || payload.ID <= 0 {
		return 0, fmt.Errorf("empresa_id e id son obligatorios")
	}
	payload.RolOperacion = normalizeComisionRol(payload.RolOperacion)
	payload.ServicioFiltro = normalizeComisionFiltro(payload.ServicioFiltro)
	payload.PorcentajeComision = normalizeComisionPorcentaje(payload.PorcentajeComision)
	payload.TopeComision = normalizeComisionTope(payload.TopeComision)
	payload.Prioridad = normalizeComisionPrioridad(payload.Prioridad)
	payload.Estado = normalizeComisionEscalaEstado(payload.Estado)
	payload.UsuarioCreador = strings.TrimSpace(payload.UsuarioCreador)
	payload.Observaciones = strings.TrimSpace(payload.Observaciones)

	if payload.PorcentajeComision <= 0 {
		return 0, fmt.Errorf("porcentaje_comision debe ser mayor a cero")
	}

	res, err := dbConn.Exec(`UPDATE empresa_comisiones_servicio_escalas
	SET
		rol_operacion = ?,
		servicio_filtro = ?,
		porcentaje_comision = ?,
		tope_comision = ?,
		prioridad = ?,
		usuario_creador = ?,
		estado = ?,
		observaciones = ?,
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ? AND id = ?`,
		payload.RolOperacion,
		payload.ServicioFiltro,
		payload.PorcentajeComision,
		payload.TopeComision,
		payload.Prioridad,
		payload.UsuarioCreador,
		payload.Estado,
		payload.Observaciones,
		payload.EmpresaID,
		payload.ID,
	)
	if err != nil {
		return 0, err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return 0, sql.ErrNoRows
	}
	return payload.ID, nil
}

// SetEmpresaComisionServicioEscalaEstado activa o desactiva una escala.
func SetEmpresaComisionServicioEscalaEstado(dbConn *sql.DB, empresaID, escalaID int64, activo bool, usuario, observaciones string) (int64, error) {
	if empresaID <= 0 || escalaID <= 0 {
		return 0, fmt.Errorf("empresa_id y escala_id son obligatorios")
	}
	estado := "inactivo"
	if activo {
		estado = "activo"
	}
	res, err := dbConn.Exec(`UPDATE empresa_comisiones_servicio_escalas
	SET
		estado = ?,
		usuario_creador = ?,
		observaciones = ?,
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ? AND id = ?`,
		estado,
		strings.TrimSpace(usuario),
		strings.TrimSpace(observaciones),
		empresaID,
		escalaID,
	)
	if err != nil {
		return 0, err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return 0, sql.ErrNoRows
	}
	return escalaID, nil
}

// GetEmpresaComisionesServicioConfiguracion obtiene la configuracion de comisiones por empresa.
func GetEmpresaComisionesServicioConfiguracion(dbConn *sql.DB, empresaID int64) (*EmpresaComisionesServicioConfiguracion, error) {
	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(habilitar_comisiones, 0),
		COALESCE(porcentaje_comision, 10),
		COALESCE(filtro_servicio, 'lavado'),
		COALESCE(aplicar_automaticamente, 1),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_comisiones_servicio_configuracion
	WHERE empresa_id = ?
	LIMIT 1`, empresaID)

	cfg := defaultEmpresaComisionesServicioConfiguracion(empresaID)
	var habilitarInt int
	var aplicarAutoInt int
	if err := row.Scan(
		&cfg.ID,
		&cfg.EmpresaID,
		&habilitarInt,
		&cfg.PorcentajeComision,
		&cfg.FiltroServicio,
		&aplicarAutoInt,
		&cfg.FechaCreacion,
		&cfg.FechaActualizacion,
		&cfg.UsuarioCreador,
		&cfg.Estado,
		&cfg.Observaciones,
	); err != nil {
		if err == sql.ErrNoRows {
			return &cfg, nil
		}
		return nil, err
	}

	cfg.HabilitarComisiones = habilitarInt == 1
	cfg.AplicarAutomaticamente = aplicarAutoInt != 0
	cfg.PorcentajeComision = normalizeComisionPorcentaje(cfg.PorcentajeComision)
	cfg.FiltroServicio = normalizeComisionFiltro(cfg.FiltroServicio)
	if strings.TrimSpace(cfg.Estado) == "" {
		cfg.Estado = "activo"
	}

	return &cfg, nil
}

// UpsertEmpresaComisionesServicioConfiguracion inserta o actualiza la configuracion por empresa.
func UpsertEmpresaComisionesServicioConfiguracion(dbConn *sql.DB, payload EmpresaComisionesServicioConfiguracion) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	payload.PorcentajeComision = normalizeComisionPorcentaje(payload.PorcentajeComision)
	payload.FiltroServicio = normalizeComisionFiltro(payload.FiltroServicio)
	if payload.FiltroServicio == "" {
		payload.FiltroServicio = "lavado"
	}
	if strings.TrimSpace(payload.Estado) == "" {
		payload.Estado = "activo"
	}

	var existingID int64
	err := dbConn.QueryRow(`SELECT id FROM empresa_comisiones_servicio_configuracion WHERE empresa_id = ? LIMIT 1`, payload.EmpresaID).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	if existingID > 0 {
		_, err = dbConn.Exec(`UPDATE empresa_comisiones_servicio_configuracion
		SET
			habilitar_comisiones = ?,
			porcentaje_comision = ?,
			filtro_servicio = ?,
			aplicar_automaticamente = ?,
			usuario_creador = ?,
			estado = ?,
			observaciones = ?,
			fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE empresa_id = ?`,
			boolToInt(payload.HabilitarComisiones),
			payload.PorcentajeComision,
			payload.FiltroServicio,
			boolToInt(payload.AplicarAutomaticamente),
			strings.TrimSpace(payload.UsuarioCreador),
			strings.TrimSpace(payload.Estado),
			strings.TrimSpace(payload.Observaciones),
			payload.EmpresaID,
		)
		if err != nil {
			return 0, err
		}
		return existingID, nil
	}

	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_comisiones_servicio_configuracion (
		empresa_id,
		habilitar_comisiones,
		porcentaje_comision,
		filtro_servicio,
		aplicar_automaticamente,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		payload.EmpresaID,
		boolToInt(payload.HabilitarComisiones),
		payload.PorcentajeComision,
		payload.FiltroServicio,
		boolToInt(payload.AplicarAutomaticamente),
		strings.TrimSpace(payload.UsuarioCreador),
		strings.TrimSpace(payload.Estado),
		strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// CreateEmpresaComisionServicioMovimiento registra un movimiento de comision por servicio.
func CreateEmpresaComisionServicioMovimiento(dbConn *sql.DB, payload EmpresaComisionServicioMovimiento) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	payload.UsuarioOrigen = strings.TrimSpace(payload.UsuarioOrigen)
	payload.UsuarioLavador = defaultComisionLavador(payload.UsuarioLavador, payload.UsuarioOrigen)
	payload.UsuarioOrigenID = resolveEmpresaUsuarioIDByReferenceSilent(dbConn, payload.EmpresaID, payload.UsuarioOrigenID, payload.UsuarioOrigen)
	payload.UsuarioLavadorID = resolveEmpresaUsuarioIDByReferenceSilent(dbConn, payload.EmpresaID, payload.UsuarioLavadorID, payload.UsuarioLavador)
	payload.RolOperacion = normalizeComisionRol(payload.RolOperacion)
	payload.Moneda = normalizeComisionMoneda(payload.Moneda)
	payload.BaseServicio = round2(payload.BaseServicio)
	payload.PorcentajeComision = normalizeComisionPorcentaje(payload.PorcentajeComision)
	payload.MontoComision = round2(payload.MontoComision)
	payload.MontoComisionBruto = round2(payload.MontoComisionBruto)
	payload.TopeComisionAplicado = normalizeComisionTope(payload.TopeComisionAplicado)
	payload.ServicioCodigo = strings.TrimSpace(payload.ServicioCodigo)
	payload.ServicioNombre = strings.TrimSpace(payload.ServicioNombre)
	payload.ServicioCategoria = strings.TrimSpace(payload.ServicioCategoria)
	payload.OrigenMovimiento = normalizeComisionOrigen(payload.OrigenMovimiento)
	payload.EsAjusteManual = payload.EsAjusteManual || payload.OrigenMovimiento == EmpresaComisionServicioOrigenAjusteManual
	if payload.EsAjusteManual {
		payload.OrigenMovimiento = EmpresaComisionServicioOrigenAjusteManual
	}
	payload.ReferenciaAjuste = strings.TrimSpace(payload.ReferenciaAjuste)
	payload.AjusteEstado = normalizeComisionAjusteEstado(payload.AjusteEstado)
	payload.AprobadoPor = strings.TrimSpace(payload.AprobadoPor)
	payload.AprobadoEn = strings.TrimSpace(payload.AprobadoEn)
	payload.PeriodoLiquidacionDesde = strings.TrimSpace(payload.PeriodoLiquidacionDesde)
	payload.PeriodoLiquidacionHasta = strings.TrimSpace(payload.PeriodoLiquidacionHasta)
	payload.LiquidadoEn = strings.TrimSpace(payload.LiquidadoEn)
	payload.LiquidadoPor = strings.TrimSpace(payload.LiquidadoPor)
	payload.FechaMovimiento = strings.TrimSpace(payload.FechaMovimiento)

	if payload.MontoComisionBruto == 0 {
		payload.MontoComisionBruto = payload.MontoComision
	}
	if payload.TopeComisionAplicado > 0 && payload.MontoComision > payload.TopeComisionAplicado {
		payload.MontoComision = payload.TopeComisionAplicado
	}

	if payload.EsAjusteManual {
		if payload.MontoComision > -0.0001 && payload.MontoComision < 0.0001 {
			return 0, fmt.Errorf("monto_comision debe ser diferente de cero para ajuste manual")
		}
		if payload.ReferenciaAjuste == "" {
			payload.ReferenciaAjuste = "AJ-COM-" + time.Now().Format("20060102150405")
		}
		if strings.TrimSpace(payload.AjusteEstado) == "" {
			payload.AjusteEstado = EmpresaComisionServicioAjustePendiente
		}
		switch payload.AjusteEstado {
		case EmpresaComisionServicioAjustePendiente:
			if strings.TrimSpace(payload.Estado) == "" {
				payload.Estado = "pendiente"
			}
			payload.AprobadoPor = ""
			payload.AprobadoEn = ""
		case EmpresaComisionServicioAjusteRechazado:
			if strings.TrimSpace(payload.Estado) == "" {
				payload.Estado = "inactivo"
			}
			if payload.AprobadoEn == "" {
				payload.AprobadoEn = time.Now().Format("2006-01-02 15:04:05")
			}
		default:
			payload.AjusteEstado = EmpresaComisionServicioAjusteAprobado
			if strings.TrimSpace(payload.Estado) == "" {
				payload.Estado = "activo"
			}
			if payload.AprobadoEn == "" {
				payload.AprobadoEn = time.Now().Format("2006-01-02 15:04:05")
			}
			if payload.AprobadoPor == "" {
				payload.AprobadoPor = strings.TrimSpace(payload.UsuarioCreador)
			}
		}
	} else {
		if payload.MontoComision <= 0 {
			return 0, fmt.Errorf("monto_comision debe ser mayor a cero")
		}
		payload.AjusteEstado = EmpresaComisionServicioAjusteAprobado
		if strings.TrimSpace(payload.Estado) == "" {
			payload.Estado = "activo"
		}
		if payload.AprobadoEn == "" {
			payload.AprobadoEn = time.Now().Format("2006-01-02 15:04:05")
		}
		if payload.AprobadoPor == "" {
			payload.AprobadoPor = strings.TrimSpace(payload.UsuarioCreador)
		}
	}

	if strings.TrimSpace(payload.Estado) == "" {
		payload.Estado = "activo"
	}
	payload.Estado = normalizeComisionEstado(payload.Estado)
	payload.UsuarioCreador = strings.TrimSpace(payload.UsuarioCreador)
	if payload.UsuarioCreador == "" {
		payload.UsuarioCreador = payload.UsuarioOrigen
	}
	if payload.UsuarioCreador == "" {
		payload.UsuarioCreador = "sistema"
	}
	payload.Observaciones = strings.TrimSpace(payload.Observaciones)
	if payload.LiquidacionNominaID < 0 {
		payload.LiquidacionNominaID = 0
	}

	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_comisiones_servicio_movimientos (
		empresa_id,
		carrito_id,
		carrito_item_id,
		servicio_id,
		servicio_codigo,
		servicio_nombre,
		servicio_categoria,
		usuario_origen,
		usuario_origen_id,
		usuario_lavador,
		usuario_lavador_id,
		rol_operacion,
		escala_id,
		venta_referencia,
		moneda,
		base_servicio,
		porcentaje_comision,
		monto_comision_bruto,
		tope_comision_aplicado,
		monto_comision,
		origen_movimiento,
		ajuste_manual,
		referencia_ajuste,
		ajuste_estado,
		aprobado_por,
		aprobado_en,
		liquidacion_nomina_id,
		periodo_liquidacion_desde,
		periodo_liquidacion_hasta,
		liquidado_en,
		liquidado_por,
		fecha_movimiento,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, COALESCE(NULLIF(?, ''), CAST(CURRENT_TIMESTAMP AS TEXT)), ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		payload.EmpresaID,
		payload.CarritoID,
		payload.CarritoItemID,
		payload.ServicioID,
		payload.ServicioCodigo,
		payload.ServicioNombre,
		payload.ServicioCategoria,
		payload.UsuarioOrigen,
		payload.UsuarioOrigenID,
		payload.UsuarioLavador,
		payload.UsuarioLavadorID,
		payload.RolOperacion,
		payload.EscalaID,
		strings.TrimSpace(payload.VentaReferencia),
		payload.Moneda,
		payload.BaseServicio,
		payload.PorcentajeComision,
		payload.MontoComisionBruto,
		payload.TopeComisionAplicado,
		payload.MontoComision,
		payload.OrigenMovimiento,
		boolToInt(payload.EsAjusteManual),
		payload.ReferenciaAjuste,
		payload.AjusteEstado,
		payload.AprobadoPor,
		payload.AprobadoEn,
		payload.LiquidacionNominaID,
		payload.PeriodoLiquidacionDesde,
		payload.PeriodoLiquidacionHasta,
		payload.LiquidadoEn,
		payload.LiquidadoPor,
		payload.FechaMovimiento,
		payload.UsuarioCreador,
		payload.Estado,
		payload.Observaciones,
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// CreateEmpresaComisionServicioAjusteManual crea una comision de ajuste en estado pendiente.
func CreateEmpresaComisionServicioAjusteManual(dbConn *sql.DB, payload EmpresaComisionServicioMovimiento) (int64, error) {
	payload.EsAjusteManual = true
	payload.OrigenMovimiento = EmpresaComisionServicioOrigenAjusteManual
	payload.AjusteEstado = EmpresaComisionServicioAjustePendiente
	if strings.TrimSpace(payload.Estado) == "" {
		payload.Estado = "pendiente"
	}
	return CreateEmpresaComisionServicioMovimiento(dbConn, payload)
}

// ResolverEmpresaComisionServicioAjusteManual aprueba o rechaza un ajuste manual pendiente.
func ResolverEmpresaComisionServicioAjusteManual(dbConn *sql.DB, empresaID, movimientoID int64, aprobar bool, usuario, observaciones string) (*EmpresaComisionServicioMovimiento, error) {
	if empresaID <= 0 || movimientoID <= 0 {
		return nil, fmt.Errorf("empresa_id y movimiento_id son obligatorios")
	}
	mov, err := GetEmpresaComisionServicioMovimientoByID(dbConn, empresaID, movimientoID)
	if err != nil {
		return nil, err
	}
	if !mov.EsAjusteManual {
		return nil, fmt.Errorf("el movimiento no corresponde a un ajuste manual")
	}
	if mov.AjusteEstado != EmpresaComisionServicioAjustePendiente {
		return nil, fmt.Errorf("el ajuste ya fue procesado")
	}

	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = "sistema"
	}
	nuevoAjusteEstado := EmpresaComisionServicioAjusteRechazado
	nuevoEstado := "inactivo"
	if aprobar {
		nuevoAjusteEstado = EmpresaComisionServicioAjusteAprobado
		nuevoEstado = "activo"
	}

	obs := strings.TrimSpace(observaciones)
	if obs == "" {
		if aprobar {
			obs = firstNonEmpty(strings.TrimSpace(mov.Observaciones), "ajuste manual aprobado")
		} else {
			obs = firstNonEmpty(strings.TrimSpace(mov.Observaciones), "ajuste manual rechazado")
		}
	}

	res, err := dbConn.Exec(`UPDATE empresa_comisiones_servicio_movimientos
	SET
		ajuste_estado = ?,
		aprobado_por = ?,
		aprobado_en = CURRENT_TIMESTAMP,
		estado = ?,
		observaciones = ?,
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ? AND id = ?`,
		nuevoAjusteEstado,
		usuario,
		nuevoEstado,
		obs,
		empresaID,
		movimientoID,
	)
	if err != nil {
		return nil, err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return nil, sql.ErrNoRows
	}

	return GetEmpresaComisionServicioMovimientoByID(dbConn, empresaID, movimientoID)
}

func buildEmpresaComisionServicioMovWhere(empresaID int64, filter EmpresaComisionServicioMovimientoFilter) (string, []interface{}) {
	clauses := []string{"empresa_id = ?"}
	args := []interface{}{empresaID}

	if filter.MovimientoID > 0 {
		clauses = append(clauses, "id = ?")
		args = append(args, filter.MovimientoID)
	}
	if !filter.IncludeInactive && !filter.SoloPendientes {
		clauses = append(clauses, "COALESCE(estado, 'activo') = 'activo'")
	}
	if strings.TrimSpace(filter.Desde) != "" {
		clauses = append(clauses, "date(COALESCE(fecha_movimiento, fecha_creacion)) >= date(?)")
		args = append(args, strings.TrimSpace(filter.Desde))
	}
	if strings.TrimSpace(filter.Hasta) != "" {
		clauses = append(clauses, "date(COALESCE(fecha_movimiento, fecha_creacion)) <= date(?)")
		args = append(args, strings.TrimSpace(filter.Hasta))
	}
	if usuario := strings.TrimSpace(filter.UsuarioLavador); usuario != "" {
		like := "%" + strings.ToLower(usuario) + "%"
		clauses = append(clauses, `(lower(COALESCE(usuario_lavador, '')) LIKE ?
			OR lower(COALESCE(usuario_origen, '')) LIKE ?
			OR CAST(COALESCE(usuario_lavador_id, 0) AS TEXT) = ?
			OR CAST(COALESCE(usuario_origen_id, 0) AS TEXT) = ?)`)
		args = append(args, like, like, usuario, usuario)
	}
	if rol := normalizeComisionRol(filter.RolOperacion); rol != "" {
		clauses = append(clauses, "lower(COALESCE(rol_operacion, '')) = ?")
		args = append(args, rol)
	}
	if servicioFiltro := strings.TrimSpace(filter.ServicioFiltro); servicioFiltro != "" {
		like := "%" + strings.ToLower(servicioFiltro) + "%"
		clauses = append(clauses, "(lower(COALESCE(servicio_nombre, '')) LIKE ? OR lower(COALESCE(servicio_categoria, '')) LIKE ? OR lower(COALESCE(servicio_codigo, '')) LIKE ?)")
		args = append(args, like, like, like)
	}
	if filter.SoloAjustes {
		clauses = append(clauses, "COALESCE(ajuste_manual, 0) = 1")
	}
	if origin := strings.TrimSpace(filter.OrigenMovimiento); origin != "" {
		clauses = append(clauses, "lower(COALESCE(origen_movimiento, 'venta')) = ?")
		args = append(args, normalizeComisionOrigen(origin))
	}
	if filter.SoloPendientes {
		clauses = append(clauses, "lower(COALESCE(ajuste_estado, 'aprobado')) = ?")
		args = append(args, EmpresaComisionServicioAjustePendiente)
	}
	if estadoAjuste := strings.TrimSpace(filter.AjusteEstado); estadoAjuste != "" {
		clauses = append(clauses, "lower(COALESCE(ajuste_estado, 'aprobado')) = ?")
		args = append(args, normalizeComisionAjusteEstado(estadoAjuste))
	}
	if filter.NoLiquidado {
		clauses = append(clauses, "COALESCE(liquidacion_nomina_id, 0) = 0")
	}
	if filter.LiquidacionNominaID > 0 {
		clauses = append(clauses, "COALESCE(liquidacion_nomina_id, 0) = ?")
		args = append(args, filter.LiquidacionNominaID)
	}

	return strings.Join(clauses, " AND "), args
}

// ListEmpresaComisionServicioMovimientos lista movimientos segun filtros.
func ListEmpresaComisionServicioMovimientos(dbConn *sql.DB, empresaID int64, filter EmpresaComisionServicioMovimientoFilter) ([]EmpresaComisionServicioMovimiento, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 200
	}
	if limit > 2000 {
		limit = 2000
	}

	whereSQL, args := buildEmpresaComisionServicioMovWhere(empresaID, filter)
	query := fmt.Sprintf(`SELECT
		id,
		empresa_id,
		COALESCE(carrito_id, 0),
		COALESCE(carrito_item_id, 0),
		COALESCE(servicio_id, 0),
		COALESCE(servicio_codigo, ''),
		COALESCE(servicio_nombre, ''),
		COALESCE(servicio_categoria, ''),
		COALESCE(usuario_origen, ''),
		COALESCE(usuario_origen_id, 0),
		COALESCE(usuario_lavador, ''),
		COALESCE(usuario_lavador_id, 0),
		COALESCE(rol_operacion, ''),
		COALESCE(escala_id, 0),
		COALESCE(venta_referencia, ''),
		COALESCE(moneda, 'COP'),
		COALESCE(base_servicio, 0),
		COALESCE(porcentaje_comision, 0),
		COALESCE(monto_comision_bruto, 0),
		COALESCE(tope_comision_aplicado, 0),
		COALESCE(monto_comision, 0),
		COALESCE(origen_movimiento, 'venta'),
		COALESCE(ajuste_manual, 0),
		COALESCE(referencia_ajuste, ''),
		COALESCE(ajuste_estado, 'aprobado'),
		COALESCE(aprobado_por, ''),
		COALESCE(aprobado_en, ''),
		COALESCE(liquidacion_nomina_id, 0),
		COALESCE(periodo_liquidacion_desde, ''),
		COALESCE(periodo_liquidacion_hasta, ''),
		COALESCE(liquidado_en, ''),
		COALESCE(liquidado_por, ''),
		COALESCE(fecha_movimiento, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_comisiones_servicio_movimientos
	WHERE %s
	ORDER BY pcs_ts(COALESCE(fecha_movimiento, fecha_creacion)) DESC, id DESC
	LIMIT %d`, whereSQL, limit)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]EmpresaComisionServicioMovimiento, 0)
	for rows.Next() {
		var row EmpresaComisionServicioMovimiento
		var ajusteManualInt int
		if err := rows.Scan(
			&row.ID,
			&row.EmpresaID,
			&row.CarritoID,
			&row.CarritoItemID,
			&row.ServicioID,
			&row.ServicioCodigo,
			&row.ServicioNombre,
			&row.ServicioCategoria,
			&row.UsuarioOrigen,
			&row.UsuarioOrigenID,
			&row.UsuarioLavador,
			&row.UsuarioLavadorID,
			&row.RolOperacion,
			&row.EscalaID,
			&row.VentaReferencia,
			&row.Moneda,
			&row.BaseServicio,
			&row.PorcentajeComision,
			&row.MontoComisionBruto,
			&row.TopeComisionAplicado,
			&row.MontoComision,
			&row.OrigenMovimiento,
			&ajusteManualInt,
			&row.ReferenciaAjuste,
			&row.AjusteEstado,
			&row.AprobadoPor,
			&row.AprobadoEn,
			&row.LiquidacionNominaID,
			&row.PeriodoLiquidacionDesde,
			&row.PeriodoLiquidacionHasta,
			&row.LiquidadoEn,
			&row.LiquidadoPor,
			&row.FechaMovimiento,
			&row.FechaCreacion,
			&row.FechaActualizacion,
			&row.UsuarioCreador,
			&row.Estado,
			&row.Observaciones,
		); err != nil {
			return nil, err
		}
		row.Moneda = normalizeComisionMoneda(row.Moneda)
		row.BaseServicio = round2(row.BaseServicio)
		row.PorcentajeComision = normalizeComisionPorcentaje(row.PorcentajeComision)
		row.MontoComisionBruto = round2(row.MontoComisionBruto)
		row.TopeComisionAplicado = normalizeComisionTope(row.TopeComisionAplicado)
		row.MontoComision = round2(row.MontoComision)
		row.OrigenMovimiento = normalizeComisionOrigen(row.OrigenMovimiento)
		row.EsAjusteManual = ajusteManualInt == 1 || row.OrigenMovimiento == EmpresaComisionServicioOrigenAjusteManual
		row.AjusteEstado = normalizeComisionAjusteEstado(row.AjusteEstado)
		if row.UsuarioOrigenID == 0 {
			row.UsuarioOrigenID = resolveEmpresaUsuarioIDByReferenceSilent(dbConn, row.EmpresaID, 0, row.UsuarioOrigen)
		}
		row.UsuarioLavador = defaultComisionLavador(row.UsuarioLavador, row.UsuarioOrigen)
		if row.UsuarioLavadorID == 0 {
			row.UsuarioLavadorID = resolveEmpresaUsuarioIDByReferenceSilent(dbConn, row.EmpresaID, 0, row.UsuarioLavador)
		}
		row.RolOperacion = normalizeComisionRol(row.RolOperacion)
		row.Estado = normalizeComisionEstado(row.Estado)
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// GetEmpresaComisionServicioMovimientoByID obtiene un movimiento de comisiones por id.
func GetEmpresaComisionServicioMovimientoByID(dbConn *sql.DB, empresaID, movimientoID int64) (*EmpresaComisionServicioMovimiento, error) {
	rows, err := ListEmpresaComisionServicioMovimientos(dbConn, empresaID, EmpresaComisionServicioMovimientoFilter{
		MovimientoID:    movimientoID,
		IncludeInactive: true,
		Limit:           1,
	})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, sql.ErrNoRows
	}
	copy := rows[0]
	return &copy, nil
}

// GetEmpresaComisionesServicioReporte construye reporte agregado por lavador.
func GetEmpresaComisionesServicioReporte(dbConn *sql.DB, empresaID int64, filter EmpresaComisionServicioMovimientoFilter) (*EmpresaComisionesServicioReporte, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}

	cfg, err := GetEmpresaComisionesServicioConfiguracion(dbConn, empresaID)
	if err != nil {
		return nil, err
	}

	escalas, err := ListEmpresaComisionServicioEscalas(dbConn, empresaID, false, filter.RolOperacion, 1000)
	if err != nil {
		return nil, err
	}

	movs, err := ListEmpresaComisionServicioMovimientos(dbConn, empresaID, filter)
	if err != nil {
		return nil, err
	}

	resumen := EmpresaComisionesServicioResumen{}
	byLavador := map[string]*EmpresaComisionServicioLavadorResumen{}

	for _, mov := range movs {
		resumen.TotalBaseServicios = round2(resumen.TotalBaseServicios + mov.BaseServicio)
		resumen.TotalComisiones = round2(resumen.TotalComisiones + mov.MontoComision)
		resumen.CantidadMovimientos++
		if mov.EsAjusteManual {
			resumen.TotalAjustesManuales = round2(resumen.TotalAjustesManuales + mov.MontoComision)
		}
		if mov.AjusteEstado == EmpresaComisionServicioAjustePendiente {
			resumen.PendientesAprobacion++
		}
		if mov.LiquidacionNominaID > 0 {
			resumen.TotalLiquidadas = round2(resumen.TotalLiquidadas + mov.MontoComision)
		} else {
			resumen.TotalPendientesLiquidar = round2(resumen.TotalPendientesLiquidar + mov.MontoComision)
		}

		key := defaultComisionLavador(mov.UsuarioLavador, mov.UsuarioOrigen)
		if mov.UsuarioLavadorID > 0 {
			key = fmt.Sprintf("usuario:%d", mov.UsuarioLavadorID)
		}
		entry := byLavador[key]
		if entry == nil {
			entry = &EmpresaComisionServicioLavadorResumen{
				UsuarioID:      mov.UsuarioLavadorID,
				UsuarioLavador: defaultComisionLavador(mov.UsuarioLavador, mov.UsuarioOrigen),
			}
			byLavador[key] = entry
		}
		if entry.UsuarioID == 0 {
			entry.UsuarioID = mov.UsuarioLavadorID
		}
		entry.TotalBaseServicios = round2(entry.TotalBaseServicios + mov.BaseServicio)
		entry.TotalComision = round2(entry.TotalComision + mov.MontoComision)
		entry.CantidadMovimientos++
	}

	lavadores := make([]EmpresaComisionServicioLavadorResumen, 0, len(byLavador))
	for _, row := range byLavador {
		lavadores = append(lavadores, *row)
	}
	sort.Slice(lavadores, func(i, j int) bool {
		if lavadores[i].TotalComision == lavadores[j].TotalComision {
			return lavadores[i].UsuarioLavador < lavadores[j].UsuarioLavador
		}
		return lavadores[i].TotalComision > lavadores[j].TotalComision
	})

	resumen.LavadoresConComision = int64(len(lavadores))

	return &EmpresaComisionesServicioReporte{
		EmpresaID:     empresaID,
		Desde:         strings.TrimSpace(filter.Desde),
		Hasta:         strings.TrimSpace(filter.Hasta),
		Configuracion: cfg,
		Escalas:       escalas,
		Resumen:       resumen,
		Lavadores:     lavadores,
		Movimientos:   movs,
	}, nil
}

func listComisionServicioItemsFromCarrito(dbConn *sql.DB, empresaID, carritoID int64) ([]comisionServicioItemSnapshot, error) {
	if empresaID <= 0 || carritoID <= 0 {
		return nil, fmt.Errorf("empresa_id y carrito_id son obligatorios")
	}

	hasServicios, err := tableExists(dbConn, "servicios")
	if err != nil {
		hasServicios = false
	}

	query := `SELECT
		i.id,
		COALESCE(i.referencia_id, 0),
		COALESCE(i.codigo_item, ''),
		COALESCE(i.descripcion, ''),
		COALESCE(i.total_linea, 0),
		'',
		'',
		''
	FROM carrito_compra_items i
	WHERE i.empresa_id = ?
		AND i.carrito_id = ?
		AND COALESCE(i.estado, 'activo') = 'activo'
		AND lower(COALESCE(i.tipo_item, 'producto')) = 'servicio'
	ORDER BY i.id ASC`
	if hasServicios {
		query = `SELECT
			i.id,
			COALESCE(i.referencia_id, 0),
			COALESCE(i.codigo_item, ''),
			COALESCE(i.descripcion, ''),
			COALESCE(i.total_linea, 0),
			COALESCE(s.codigo, ''),
			COALESCE(s.nombre, ''),
			COALESCE(s.categoria, '')
		FROM carrito_compra_items i
		LEFT JOIN servicios s ON s.empresa_id = i.empresa_id AND s.id = i.referencia_id
		WHERE i.empresa_id = ?
			AND i.carrito_id = ?
			AND COALESCE(i.estado, 'activo') = 'activo'
			AND lower(COALESCE(i.tipo_item, 'producto')) = 'servicio'
		ORDER BY i.id ASC`
	}

	rows, err := dbConn.Query(query, empresaID, carritoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]comisionServicioItemSnapshot, 0)
	for rows.Next() {
		var row comisionServicioItemSnapshot
		if err := rows.Scan(
			&row.CarritoItemID,
			&row.ServicioID,
			&row.CodigoItem,
			&row.Descripcion,
			&row.TotalLinea,
			&row.ServicioCodigo,
			&row.ServicioNombre,
			&row.ServicioCategoria,
		); err != nil {
			return nil, err
		}
		items = append(items, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func servicioCumpleFiltroComision(item comisionServicioItemSnapshot, filtro string) bool {
	filtro = normalizeComisionFiltro(filtro)
	if filtro == "" {
		return true
	}
	haystack := strings.ToLower(strings.Join([]string{
		item.ServicioCodigo,
		item.ServicioNombre,
		item.ServicioCategoria,
		item.CodigoItem,
		item.Descripcion,
	}, " "))
	return strings.Contains(haystack, filtro)
}

func findMatchingComisionEscala(escalas []EmpresaComisionServicioEscala, item comisionServicioItemSnapshot, rolOperacion string) *EmpresaComisionServicioEscala {
	rol := normalizeComisionRol(rolOperacion)
	for i := range escalas {
		escala := escalas[i]
		if escala.Estado != "activo" {
			continue
		}
		if escala.RolOperacion != "" && escala.RolOperacion != rol {
			continue
		}
		if !servicioCumpleFiltroComision(item, escala.ServicioFiltro) {
			continue
		}
		copy := escala
		return &copy
	}
	return nil
}

// RegisterEmpresaComisionesServicioDesdeCarrito calcula y registra comisiones por servicios al cerrar carrito.
func RegisterEmpresaComisionesServicioDesdeCarrito(dbConn *sql.DB, empresaID, carritoID int64, usuarioLavador, usuarioOrigen, rolOperacion string) (*EmpresaComisionServicioRegistroResultado, error) {
	if empresaID <= 0 || carritoID <= 0 {
		return nil, fmt.Errorf("empresa_id y carrito_id son obligatorios")
	}

	cfg, err := GetEmpresaComisionesServicioConfiguracion(dbConn, empresaID)
	if err != nil {
		return nil, err
	}

	escalas, err := ListEmpresaComisionServicioEscalas(dbConn, empresaID, false, rolOperacion, 1000)
	if err != nil {
		return nil, err
	}

	result := &EmpresaComisionServicioRegistroResultado{
		Habilitada:           cfg.HabilitarComisiones,
		AplicacionAutomatica: cfg.AplicarAutomaticamente,
		PorcentajeComision:   cfg.PorcentajeComision,
		FiltroServicio:       cfg.FiltroServicio,
		UsuarioLavador:       defaultComisionLavador(usuarioLavador, usuarioOrigen),
		UsuarioLavadorID:     resolveEmpresaUsuarioIDByReferenceSilent(dbConn, empresaID, 0, defaultComisionLavador(usuarioLavador, usuarioOrigen)),
		RolOperacion:         normalizeComisionRol(rolOperacion),
	}
	usuarioOrigenID := resolveEmpresaUsuarioIDByReferenceSilent(dbConn, empresaID, 0, usuarioOrigen)

	if !cfg.HabilitarComisiones {
		result.Warning = "configuracion de comisiones deshabilitada"
		return result, nil
	}
	if !cfg.AplicarAutomaticamente {
		result.Warning = "comisiones configuradas en modo manual"
		return result, nil
	}
	if cfg.PorcentajeComision <= 0 && len(escalas) == 0 {
		result.Warning = "porcentaje de comision no configurado"
		return result, nil
	}

	var ventaReferencia string
	var moneda string
	err = dbConn.QueryRow(`SELECT
		COALESCE(codigo, ''),
		COALESCE(moneda, 'COP')
	FROM carritos_compras
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, carritoID).Scan(&ventaReferencia, &moneda)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("carrito no encontrado")
		}
		return nil, err
	}

	items, err := listComisionServicioItemsFromCarrito(dbConn, empresaID, carritoID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		result.Warning = "sin items de tipo servicio para comision"
		return result, nil
	}

	usuarioOrigen = strings.TrimSpace(usuarioOrigen)
	if usuarioOrigen == "" {
		usuarioOrigen = "sistema"
	}

	escalasAplicadas := make(map[int64]struct{})
	for _, item := range items {
		escala := findMatchingComisionEscala(escalas, item, rolOperacion)
		coincideFiltroBase := servicioCumpleFiltroComision(item, cfg.FiltroServicio)
		if escala == nil && !coincideFiltroBase {
			continue
		}

		porcentaje := cfg.PorcentajeComision
		filtroAplicado := cfg.FiltroServicio
		escalaID := int64(0)
		tope := 0.0
		if escala != nil {
			porcentaje = escala.PorcentajeComision
			filtroAplicado = escala.ServicioFiltro
			escalaID = escala.ID
			tope = escala.TopeComision
			escalasAplicadas[escala.ID] = struct{}{}
		}

		if porcentaje <= 0 {
			continue
		}

		baseServicio := round2(item.TotalLinea)
		if baseServicio <= 0 {
			continue
		}
		montoBruto := round2(baseServicio * (porcentaje / 100))
		if montoBruto <= 0 {
			continue
		}

		montoComision := montoBruto
		if tope > 0 && montoComision > tope {
			montoComision = tope
			result.TotalTopesAplicados = round2(result.TotalTopesAplicados + (montoBruto - montoComision))
		}

		id, err := CreateEmpresaComisionServicioMovimiento(dbConn, EmpresaComisionServicioMovimiento{
			EmpresaID:            empresaID,
			CarritoID:            carritoID,
			CarritoItemID:        item.CarritoItemID,
			ServicioID:           item.ServicioID,
			ServicioCodigo:       firstNonEmpty(item.ServicioCodigo, item.CodigoItem),
			ServicioNombre:       firstNonEmpty(item.ServicioNombre, item.Descripcion),
			ServicioCategoria:    item.ServicioCategoria,
			UsuarioOrigen:        usuarioOrigen,
			UsuarioOrigenID:      usuarioOrigenID,
			UsuarioLavador:       result.UsuarioLavador,
			UsuarioLavadorID:     result.UsuarioLavadorID,
			RolOperacion:         result.RolOperacion,
			EscalaID:             escalaID,
			VentaReferencia:      ventaReferencia,
			Moneda:               moneda,
			BaseServicio:         baseServicio,
			PorcentajeComision:   porcentaje,
			MontoComisionBruto:   montoBruto,
			TopeComisionAplicado: tope,
			MontoComision:        montoComision,
			OrigenMovimiento:     EmpresaComisionServicioOrigenVenta,
			AjusteEstado:         EmpresaComisionServicioAjusteAprobado,
			UsuarioCreador:       usuarioOrigen,
			Estado:               "activo",
			Observaciones:        "comision registrada al cerrar carrito en estacion",
		})
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "unique") {
				continue
			}
			return nil, err
		}

		result.PorcentajeComision = porcentaje
		result.FiltroServicio = filtroAplicado
		result.RegistroIDs = append(result.RegistroIDs, id)
		result.MovimientosRegistrados++
		result.BaseServicios = round2(result.BaseServicios + baseServicio)
		result.MontoComision = round2(result.MontoComision + montoComision)
	}

	if len(escalasAplicadas) > 0 {
		ids := make([]int64, 0, len(escalasAplicadas))
		for id := range escalasAplicadas {
			ids = append(ids, id)
		}
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
		result.EscalasAplicadas = ids
	}

	if result.MovimientosRegistrados == 0 {
		if len(escalas) > 0 {
			result.Warning = "sin servicios coincidentes con escalas configuradas"
		} else {
			result.Warning = "sin servicios coincidentes con el filtro de comision"
		}
		return result, nil
	}
	result.Aplicada = true
	return result, nil
}

// GetEmpresaComisionServicioLiquidacionResumen calcula comisiones aplicables a una liquidacion de nomina.
func GetEmpresaComisionServicioLiquidacionResumen(dbConn *sql.DB, empresaID int64, aliases []string, periodoDesde, periodoHasta string) (*EmpresaComisionServicioLiquidacionResumen, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	desde := strings.TrimSpace(periodoDesde)
	hasta := strings.TrimSpace(periodoHasta)
	if desde == "" || hasta == "" {
		return nil, fmt.Errorf("periodo_desde y periodo_hasta son obligatorios")
	}

	ids := BuildEmpresaComisionServicioAliases(aliases...)
	resumen := &EmpresaComisionServicioLiquidacionResumen{
		EmpresaID:       empresaID,
		PeriodoDesde:    desde,
		PeriodoHasta:    hasta,
		Identificadores: ids,
	}
	if len(ids) == 0 {
		return resumen, nil
	}

	aliasPH := strings.TrimRight(strings.Repeat("?,", len(ids)), ",")
	args := make([]interface{}, 0, 5+len(ids)*4)
	args = append(args, empresaID, desde, hasta, desde, hasta)
	for _, alias := range ids {
		args = append(args, alias)
	}
	for _, alias := range ids {
		args = append(args, alias)
	}
	for _, alias := range ids {
		args = append(args, alias)
	}
	for _, alias := range ids {
		args = append(args, alias)
	}

	query := `SELECT
		id,
		COALESCE(base_servicio, 0),
		COALESCE(monto_comision, 0),
		COALESCE(ajuste_manual, 0)
	FROM empresa_comisiones_servicio_movimientos
	WHERE empresa_id = ?
		AND COALESCE(estado, 'activo') = 'activo'
		AND lower(COALESCE(ajuste_estado, 'aprobado')) = 'aprobado'
		AND date(COALESCE(fecha_movimiento, fecha_creacion)) >= date(?)
		AND date(COALESCE(fecha_movimiento, fecha_creacion)) <= date(?)
		AND (
			COALESCE(liquidacion_nomina_id, 0) = 0
			OR (
				COALESCE(periodo_liquidacion_desde, '') = ?
				AND COALESCE(periodo_liquidacion_hasta, '') = ?
			)
		)
		AND (
			lower(COALESCE(usuario_lavador, '')) IN (` + aliasPH + `)
			OR lower(COALESCE(usuario_origen, '')) IN (` + aliasPH + `)
			OR CAST(COALESCE(usuario_lavador_id, 0) AS TEXT) IN (` + aliasPH + `)
			OR CAST(COALESCE(usuario_origen_id, 0) AS TEXT) IN (` + aliasPH + `)
		)
	ORDER BY id ASC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var base float64
		var monto float64
		var ajusteInt int
		if err := rows.Scan(&id, &base, &monto, &ajusteInt); err != nil {
			return nil, err
		}
		resumen.CantidadMovimientos++
		resumen.TotalBaseServicios = round2(resumen.TotalBaseServicios + base)
		resumen.TotalComisiones = round2(resumen.TotalComisiones + monto)
		if ajusteInt == 1 {
			resumen.TotalAjustesManuales = round2(resumen.TotalAjustesManuales + monto)
		}
		resumen.MovimientoIDs = append(resumen.MovimientoIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return resumen, nil
}

// VincularEmpresaComisionesServicioALiquidacion marca movimientos de comision como liquidados.
func VincularEmpresaComisionesServicioALiquidacion(dbConn *sql.DB, empresaID, liquidacionNominaID int64, periodoDesde, periodoHasta, usuario string, movimientoIDs []int64) error {
	if empresaID <= 0 || liquidacionNominaID <= 0 {
		return fmt.Errorf("empresa_id y liquidacion_nomina_id son obligatorios")
	}
	desde := strings.TrimSpace(periodoDesde)
	hasta := strings.TrimSpace(periodoHasta)
	if desde == "" || hasta == "" {
		return fmt.Errorf("periodo_desde y periodo_hasta son obligatorios")
	}
	ids := uniquePositiveInt64(movimientoIDs)
	if len(ids) == 0 {
		return nil
	}

	if strings.TrimSpace(usuario) == "" {
		usuario = "sistema"
	}

	ph := strings.TrimRight(strings.Repeat("?,", len(ids)), ",")
	args := make([]interface{}, 0, 5+len(ids))
	args = append(args,
		liquidacionNominaID,
		desde,
		hasta,
		strings.TrimSpace(usuario),
		empresaID,
	)
	for _, id := range ids {
		args = append(args, id)
	}

	query := `UPDATE empresa_comisiones_servicio_movimientos
	SET
		liquidacion_nomina_id = ?,
		periodo_liquidacion_desde = ?,
		periodo_liquidacion_hasta = ?,
		liquidado_en = CURRENT_TIMESTAMP,
		liquidado_por = ?,
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ?
		AND id IN (` + ph + `)`

	_, err := dbConn.Exec(query, args...)
	return err
}

// LimpiarVinculoEmpresaComisionesServicioLiquidacion desasocia comisiones previamente vinculadas a una liquidacion.
func LimpiarVinculoEmpresaComisionesServicioLiquidacion(dbConn *sql.DB, empresaID, liquidacionNominaID int64) error {
	if empresaID <= 0 || liquidacionNominaID <= 0 {
		return nil
	}
	_, err := dbConn.Exec(`UPDATE empresa_comisiones_servicio_movimientos
	SET
		liquidacion_nomina_id = 0,
		periodo_liquidacion_desde = '',
		periodo_liquidacion_hasta = '',
		liquidado_en = '',
		liquidado_por = '',
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ? AND liquidacion_nomina_id = ?`, empresaID, liquidacionNominaID)
	return err
}

func uniquePositiveInt64(values []int64) []int64 {
	seen := make(map[int64]struct{})
	out := make([]int64, 0, len(values))
	for _, v := range values {
		if v <= 0 {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
