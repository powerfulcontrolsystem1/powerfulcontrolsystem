package db

import (
	"database/sql"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	empresaPropinasSchemaMu    sync.Mutex
	empresaPropinasSchemaReady bool
)

const (
	EmpresaPropinaModoPorUsuario = "por_usuario"
	EmpresaPropinaModoUniversal  = "universal"

	EmpresaPropinaOrigenVenta        = "venta"
	EmpresaPropinaOrigenAjusteManual = "ajuste_manual"

	EmpresaPropinaTratamientoNoGravada = "no_gravada"
	EmpresaPropinaTratamientoGravada   = "gravada"
)

// EmpresaPropinasConfiguracion define el comportamiento de propinas por empresa.
type EmpresaPropinasConfiguracion struct {
	ID                     int64   `json:"id"`
	EmpresaID              int64   `json:"empresa_id"`
	HabilitarPropina       bool    `json:"habilitar_propina"`
	PorcentajePropina      float64 `json:"porcentaje_propina"`
	ModoDistribucion       string  `json:"modo_distribucion"`
	AplicarAutomaticamente bool    `json:"aplicar_automaticamente"`
	PaisFiscal             string  `json:"pais_fiscal"`
	RegimenFiscal          string  `json:"regimen_fiscal"`
	TratamientoFiscal      string  `json:"tratamiento_fiscal"`
	PorcentajeImpuesto     float64 `json:"porcentaje_impuesto_propina"`
	FechaCreacion          string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion     string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador         string  `json:"usuario_creador,omitempty"`
	Estado                 string  `json:"estado,omitempty"`
	Observaciones          string  `json:"observaciones,omitempty"`
}

// EmpresaPropinaMovimiento representa una propina registrada al cerrar una venta.
type EmpresaPropinaMovimiento struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	CarritoID          int64   `json:"carrito_id,omitempty"`
	CierreCajaID       int64   `json:"cierre_caja_id,omitempty"`
	VentaReferencia    string  `json:"venta_referencia,omitempty"`
	UsuarioOrigen      string  `json:"usuario_origen,omitempty"`
	UsuarioOrigenID    int64   `json:"usuario_origen_id,omitempty"`
	UsuarioAsignado    string  `json:"usuario_asignado,omitempty"`
	UsuarioAsignadoID  int64   `json:"usuario_asignado_id,omitempty"`
	ModoDistribucion   string  `json:"modo_distribucion"`
	OrigenMovimiento   string  `json:"origen_movimiento"`
	EsAjusteManual     bool    `json:"es_ajuste_manual"`
	ReferenciaAjuste   string  `json:"referencia_ajuste,omitempty"`
	Moneda             string  `json:"moneda,omitempty"`
	BaseCobro          float64 `json:"base_cobro"`
	PorcentajePropina  float64 `json:"porcentaje_propina"`
	MontoPropina       float64 `json:"monto_propina"`
	FiscalPais         string  `json:"fiscal_pais,omitempty"`
	FiscalRegimen      string  `json:"fiscal_regimen,omitempty"`
	FiscalTratamiento  string  `json:"fiscal_tratamiento,omitempty"`
	FiscalPorcentaje   float64 `json:"fiscal_porcentaje_impuesto"`
	FiscalImpuesto     float64 `json:"fiscal_impuesto_monto"`
	FiscalTotal        float64 `json:"fiscal_total"`
	ConciliadoEn       string  `json:"conciliado_en,omitempty"`
	FechaMovimiento    string  `json:"fecha_movimiento,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// EmpresaPropinaMovimientoFilter permite filtrar movimientos de propinas.
type EmpresaPropinaMovimientoFilter struct {
	Desde            string
	Hasta            string
	ModoDistribucion string
	Usuario          string
	OrigenMovimiento string
	CierreCajaID     int64
	SoloAjustes      bool
	IncludeInactive  bool
	Limit            int
}

// EmpresaPropinasResumen consolida metricas de propinas en un periodo.
type EmpresaPropinasResumen struct {
	TotalBaseCobro           float64 `json:"total_base_cobro"`
	TotalPropinas            float64 `json:"total_propinas"`
	TotalAjustesManuales     float64 `json:"total_ajustes_manuales"`
	TotalImpuestoPropina     float64 `json:"total_impuesto_propina"`
	TotalPropinasConImpuesto float64 `json:"total_propinas_con_impuesto"`
	TotalPropinasPorUsuario  float64 `json:"total_propinas_por_usuario"`
	TotalPropinasUniversal   float64 `json:"total_propinas_universal"`
	CantidadMovimientos      int64   `json:"cantidad_movimientos"`
	UsuariosActivos          int     `json:"usuarios_activos"`
	CuotaUniversalPorUsuario float64 `json:"cuota_universal_por_usuario"`
}

// EmpresaPropinaConciliacionCierre resume conciliacion de propinas para un cierre de caja/turno.
type EmpresaPropinaConciliacionCierre struct {
	EmpresaID           int64   `json:"empresa_id"`
	CierreCajaID        int64   `json:"cierre_caja_id"`
	FechaOperacion      string  `json:"fecha_operacion"`
	CantidadMovimientos int64   `json:"cantidad_movimientos"`
	TotalPropinas       float64 `json:"total_propinas"`
	TotalAjustes        float64 `json:"total_ajustes"`
	TotalImpuesto       float64 `json:"total_impuesto"`
	TotalNeto           float64 `json:"total_neto"`
	ConciliadoEn        string  `json:"conciliado_en"`
	ConciliadoPor       string  `json:"conciliado_por"`
}

// EmpresaPropinaUsuarioResumen presenta acumulados por usuario.
type EmpresaPropinaUsuarioResumen struct {
	UsuarioID         int64   `json:"usuario_id,omitempty"`
	UsuarioClave      string  `json:"usuario_clave"`
	UsuarioEtiqueta   string  `json:"usuario_etiqueta"`
	EsUsuarioActivo   bool    `json:"es_usuario_activo"`
	PropinaPorUsuario float64 `json:"propina_por_usuario"`
	PropinaUniversal  float64 `json:"propina_universal"`
	PropinaTotal      float64 `json:"propina_total"`
}

// EmpresaPropinasReporte devuelve configuracion, resumen y detalle para reportes.
type EmpresaPropinasReporte struct {
	EmpresaID     int64                          `json:"empresa_id"`
	Desde         string                         `json:"desde,omitempty"`
	Hasta         string                         `json:"hasta,omitempty"`
	Configuracion *EmpresaPropinasConfiguracion  `json:"configuracion"`
	Resumen       EmpresaPropinasResumen         `json:"resumen"`
	Usuarios      []EmpresaPropinaUsuarioResumen `json:"usuarios"`
	Movimientos   []EmpresaPropinaMovimiento     `json:"movimientos"`
}

type propinaActiveUser struct {
	ID       int64
	Clave    string
	Etiqueta string
}

// EnsureEmpresaPropinasSchema crea/migra las tablas de configuracion y movimientos de propinas.
func EnsureEmpresaPropinasSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("db connection is nil")
	}
	empresaPropinasSchemaMu.Lock()
	defer empresaPropinasSchemaMu.Unlock()

	if empresaPropinasSchemaReady {
		return nil
	}
	ready, err := empresaPropinasSchemaLooksReady(dbConn)
	if err == nil && ready {
		empresaPropinasSchemaReady = true
		return nil
	}

	bootstrapStmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_propinas_configuracion (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL UNIQUE,
			habilitar_propina INTEGER DEFAULT 0,
			porcentaje_propina REAL DEFAULT 10,
			modo_distribucion TEXT DEFAULT 'por_usuario',
			aplicar_automaticamente INTEGER DEFAULT 1,
			pais_fiscal TEXT DEFAULT 'generico',
			regimen_fiscal TEXT DEFAULT 'general',
			tratamiento_fiscal TEXT DEFAULT 'no_gravada',
			porcentaje_impuesto_propina REAL DEFAULT 0,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS empresa_propinas_movimientos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			carrito_id INTEGER DEFAULT 0,
			cierre_caja_id INTEGER DEFAULT 0,
			venta_referencia TEXT,
			usuario_origen TEXT,
			usuario_origen_id INTEGER DEFAULT 0,
			usuario_asignado TEXT,
			usuario_asignado_id INTEGER DEFAULT 0,
			modo_distribucion TEXT DEFAULT 'por_usuario',
			origen_movimiento TEXT DEFAULT 'venta',
			ajuste_manual INTEGER DEFAULT 0,
			referencia_ajuste TEXT,
			moneda TEXT DEFAULT 'COP',
			base_cobro REAL DEFAULT 0,
			porcentaje_propina REAL DEFAULT 0,
			monto_propina REAL DEFAULT 0,
			fiscal_pais TEXT DEFAULT 'generico',
			fiscal_regimen TEXT DEFAULT 'general',
			fiscal_tratamiento TEXT DEFAULT 'no_gravada',
			fiscal_porcentaje_impuesto REAL DEFAULT 0,
			fiscal_impuesto_monto REAL DEFAULT 0,
			fiscal_total REAL DEFAULT 0,
			conciliado_en TEXT,
			fecha_movimiento TEXT DEFAULT (datetime('now','localtime')),
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
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

	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_configuracion", "habilitar_propina", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_configuracion", "porcentaje_propina", "REAL DEFAULT 10"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_configuracion", "modo_distribucion", "TEXT DEFAULT 'por_usuario'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_configuracion", "aplicar_automaticamente", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_configuracion", "pais_fiscal", "TEXT DEFAULT 'generico'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_configuracion", "regimen_fiscal", "TEXT DEFAULT 'general'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_configuracion", "tratamiento_fiscal", "TEXT DEFAULT 'no_gravada'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_configuracion", "porcentaje_impuesto_propina", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_configuracion", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_configuracion", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_configuracion", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_configuracion", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "carrito_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "cierre_caja_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "venta_referencia", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "usuario_origen", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "usuario_origen_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "usuario_asignado", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "usuario_asignado_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "modo_distribucion", "TEXT DEFAULT 'por_usuario'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "origen_movimiento", "TEXT DEFAULT 'venta'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "ajuste_manual", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "referencia_ajuste", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "moneda", "TEXT DEFAULT 'COP'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "base_cobro", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "porcentaje_propina", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "monto_propina", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "fiscal_pais", "TEXT DEFAULT 'generico'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "fiscal_regimen", "TEXT DEFAULT 'general'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "fiscal_tratamiento", "TEXT DEFAULT 'no_gravada'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "fiscal_porcentaje_impuesto", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "fiscal_impuesto_monto", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "fiscal_total", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "conciliado_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "fecha_movimiento", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "observaciones", "TEXT"); err != nil {
		return err
	}

	indexStmts := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_propinas_configuracion_empresa ON empresa_propinas_configuracion(empresa_id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_propinas_configuracion_estado ON empresa_propinas_configuracion(empresa_id, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_propinas_movimientos_empresa_fecha ON empresa_propinas_movimientos(empresa_id, fecha_movimiento DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_propinas_movimientos_empresa_usuario ON empresa_propinas_movimientos(empresa_id, usuario_asignado, usuario_origen);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_propinas_movimientos_empresa_usuario_id ON empresa_propinas_movimientos(empresa_id, usuario_asignado_id, usuario_origen_id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_propinas_movimientos_empresa_modo ON empresa_propinas_movimientos(empresa_id, modo_distribucion);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_propinas_movimientos_empresa_cierre ON empresa_propinas_movimientos(empresa_id, cierre_caja_id, fecha_movimiento DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_propinas_movimientos_empresa_origen ON empresa_propinas_movimientos(empresa_id, origen_movimiento, ajuste_manual);`,
	}
	for _, stmt := range indexStmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	empresaPropinasSchemaReady = true
	return nil
}

func empresaPropinasSchemaLooksReady(dbConn *sql.DB) (bool, error) {
	ok, err := tableExists(dbConn, "empresa_propinas_configuracion")
	if err != nil || !ok {
		return false, err
	}
	ok, err = tableExists(dbConn, "empresa_propinas_movimientos")
	if err != nil || !ok {
		return false, err
	}

	requiredIndexes := []string{
		"ux_empresa_propinas_configuracion_empresa",
		"ix_empresa_propinas_movimientos_empresa_fecha",
		"ix_empresa_propinas_movimientos_empresa_cierre",
		"ix_empresa_propinas_movimientos_empresa_usuario_id",
	}
	for _, indexName := range requiredIndexes {
		indexOK, idxErr := empresaPropinasIndexExists(dbConn, indexName)
		if idxErr != nil || !indexOK {
			return false, idxErr
		}
	}
	return true, nil
}

func empresaPropinasIndexExists(dbConn *sql.DB, indexName string) (bool, error) {
	var exists bool
	err := queryRowSQLCompat(dbConn, `
		SELECT EXISTS (
			SELECT 1
			FROM pg_indexes
			WHERE schemaname = ANY (current_schemas(false))
			  AND indexname = ?
		)
	`, indexName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func defaultEmpresaPropinasConfiguracion(empresaID int64) EmpresaPropinasConfiguracion {
	return EmpresaPropinasConfiguracion{
		EmpresaID:              empresaID,
		HabilitarPropina:       false,
		PorcentajePropina:      10,
		ModoDistribucion:       EmpresaPropinaModoPorUsuario,
		AplicarAutomaticamente: true,
		PaisFiscal:             "generico",
		RegimenFiscal:          "general",
		TratamientoFiscal:      EmpresaPropinaTratamientoNoGravada,
		PorcentajeImpuesto:     0,
		Estado:                 "activo",
	}
}

func normalizePropinaModo(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "por_usuario", "usuario", "individual":
		return EmpresaPropinaModoPorUsuario
	case "universal", "global", "todos":
		return EmpresaPropinaModoUniversal
	default:
		return EmpresaPropinaModoPorUsuario
	}
}

func normalizePropinaPorcentaje(v float64) float64 {
	if v < 0 {
		v = 0
	}
	if v > 100 {
		v = 100
	}
	return round2(v)
}

func normalizePropinaMoneda(v string) string {
	m := strings.ToUpper(strings.TrimSpace(v))
	if m == "" {
		return "COP"
	}
	return m
}

func normalizePropinaFiscalPais(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return "generico"
	}
	if len(v) > 30 {
		v = v[:30]
	}
	return v
}

func normalizePropinaFiscalRegimen(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return "general"
	}
	if len(v) > 50 {
		v = v[:50]
	}
	return v
}

func normalizePropinaFiscalTratamiento(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case EmpresaPropinaTratamientoGravada:
		return EmpresaPropinaTratamientoGravada
	case EmpresaPropinaTratamientoNoGravada:
		return EmpresaPropinaTratamientoNoGravada
	default:
		return EmpresaPropinaTratamientoNoGravada
	}
}

func normalizePropinaOrigen(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case EmpresaPropinaOrigenAjusteManual, "ajuste":
		return EmpresaPropinaOrigenAjusteManual
	case EmpresaPropinaOrigenVenta, "":
		return EmpresaPropinaOrigenVenta
	default:
		return EmpresaPropinaOrigenVenta
	}
}

// GetEmpresaPropinasConfiguracion obtiene la configuracion activa de propinas por empresa.
func GetEmpresaPropinasConfiguracion(dbConn *sql.DB, empresaID int64) (*EmpresaPropinasConfiguracion, error) {
	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(habilitar_propina, 0),
		COALESCE(porcentaje_propina, 10),
		COALESCE(modo_distribucion, 'por_usuario'),
		COALESCE(aplicar_automaticamente, 1),
		COALESCE(pais_fiscal, 'generico'),
		COALESCE(regimen_fiscal, 'general'),
		COALESCE(tratamiento_fiscal, 'no_gravada'),
		COALESCE(porcentaje_impuesto_propina, 0),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_propinas_configuracion
	WHERE empresa_id = ?
	LIMIT 1`, empresaID)

	cfg := defaultEmpresaPropinasConfiguracion(empresaID)
	var habilitarInt int
	var aplicarAutoInt int
	if err := row.Scan(
		&cfg.ID,
		&cfg.EmpresaID,
		&habilitarInt,
		&cfg.PorcentajePropina,
		&cfg.ModoDistribucion,
		&aplicarAutoInt,
		&cfg.PaisFiscal,
		&cfg.RegimenFiscal,
		&cfg.TratamientoFiscal,
		&cfg.PorcentajeImpuesto,
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
	cfg.HabilitarPropina = habilitarInt == 1
	cfg.AplicarAutomaticamente = aplicarAutoInt != 0
	cfg.PorcentajePropina = normalizePropinaPorcentaje(cfg.PorcentajePropina)
	cfg.ModoDistribucion = normalizePropinaModo(cfg.ModoDistribucion)
	cfg.PaisFiscal = normalizePropinaFiscalPais(cfg.PaisFiscal)
	cfg.RegimenFiscal = normalizePropinaFiscalRegimen(cfg.RegimenFiscal)
	cfg.TratamientoFiscal = normalizePropinaFiscalTratamiento(cfg.TratamientoFiscal)
	cfg.PorcentajeImpuesto = normalizePropinaPorcentaje(cfg.PorcentajeImpuesto)
	if cfg.TratamientoFiscal != EmpresaPropinaTratamientoGravada {
		cfg.PorcentajeImpuesto = 0
	}
	if strings.TrimSpace(cfg.Estado) == "" {
		cfg.Estado = "activo"
	}

	return &cfg, nil
}

// UpsertEmpresaPropinasConfiguracion inserta o actualiza configuracion de propinas por empresa.
func UpsertEmpresaPropinasConfiguracion(dbConn *sql.DB, payload EmpresaPropinasConfiguracion) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	payload.PorcentajePropina = normalizePropinaPorcentaje(payload.PorcentajePropina)
	payload.ModoDistribucion = normalizePropinaModo(payload.ModoDistribucion)
	payload.PaisFiscal = normalizePropinaFiscalPais(payload.PaisFiscal)
	payload.RegimenFiscal = normalizePropinaFiscalRegimen(payload.RegimenFiscal)
	payload.TratamientoFiscal = normalizePropinaFiscalTratamiento(payload.TratamientoFiscal)
	payload.PorcentajeImpuesto = normalizePropinaPorcentaje(payload.PorcentajeImpuesto)
	if payload.TratamientoFiscal != EmpresaPropinaTratamientoGravada {
		payload.PorcentajeImpuesto = 0
	}
	if strings.TrimSpace(payload.Estado) == "" {
		payload.Estado = "activo"
	}

	var existingID int64
	err := dbConn.QueryRow(`SELECT id FROM empresa_propinas_configuracion WHERE empresa_id = ? LIMIT 1`, payload.EmpresaID).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	if existingID > 0 {
		_, err = dbConn.Exec(`UPDATE empresa_propinas_configuracion
		SET
			habilitar_propina = ?,
			porcentaje_propina = ?,
			modo_distribucion = ?,
			aplicar_automaticamente = ?,
			pais_fiscal = ?,
			regimen_fiscal = ?,
			tratamiento_fiscal = ?,
			porcentaje_impuesto_propina = ?,
			usuario_creador = ?,
			estado = ?,
			observaciones = ?,
			fecha_actualizacion = datetime('now','localtime')
		WHERE empresa_id = ?`,
			boolToInt(payload.HabilitarPropina),
			payload.PorcentajePropina,
			payload.ModoDistribucion,
			boolToInt(payload.AplicarAutomaticamente),
			payload.PaisFiscal,
			payload.RegimenFiscal,
			payload.TratamientoFiscal,
			payload.PorcentajeImpuesto,
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

	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_propinas_configuracion (
		empresa_id,
		habilitar_propina,
		porcentaje_propina,
		modo_distribucion,
		aplicar_automaticamente,
		pais_fiscal,
		regimen_fiscal,
		tratamiento_fiscal,
		porcentaje_impuesto_propina,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
		payload.EmpresaID,
		boolToInt(payload.HabilitarPropina),
		payload.PorcentajePropina,
		payload.ModoDistribucion,
		boolToInt(payload.AplicarAutomaticamente),
		payload.PaisFiscal,
		payload.RegimenFiscal,
		payload.TratamientoFiscal,
		payload.PorcentajeImpuesto,
		strings.TrimSpace(payload.UsuarioCreador),
		strings.TrimSpace(payload.Estado),
		strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// CreateEmpresaPropinaMovimiento registra una propina asociada al cierre de una venta.
func CreateEmpresaPropinaMovimiento(dbConn *sql.DB, payload EmpresaPropinaMovimiento) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}

	cfg, err := GetEmpresaPropinasConfiguracion(dbConn, payload.EmpresaID)
	if err != nil {
		return 0, err
	}

	payload.ModoDistribucion = normalizePropinaModo(payload.ModoDistribucion)
	payload.OrigenMovimiento = normalizePropinaOrigen(payload.OrigenMovimiento)
	payload.EsAjusteManual = payload.EsAjusteManual || payload.OrigenMovimiento == EmpresaPropinaOrigenAjusteManual
	if payload.EsAjusteManual {
		payload.OrigenMovimiento = EmpresaPropinaOrigenAjusteManual
	}
	if payload.CierreCajaID < 0 {
		payload.CierreCajaID = 0
	}
	payload.ReferenciaAjuste = strings.TrimSpace(payload.ReferenciaAjuste)
	payload.FechaMovimiento = strings.TrimSpace(payload.FechaMovimiento)
	payload.Moneda = normalizePropinaMoneda(payload.Moneda)
	payload.BaseCobro = round2(payload.BaseCobro)
	payload.PorcentajePropina = normalizePropinaPorcentaje(payload.PorcentajePropina)
	payload.MontoPropina = round2(payload.MontoPropina)
	if payload.EsAjusteManual {
		if math.Abs(payload.MontoPropina) < 0.0001 {
			return 0, fmt.Errorf("monto_propina debe ser diferente de cero para ajuste manual")
		}
		if payload.ReferenciaAjuste == "" {
			payload.ReferenciaAjuste = "AJ-PROP-" + time.Now().Format("20060102150405")
		}
	} else if payload.MontoPropina <= 0 {
		return 0, fmt.Errorf("monto_propina debe ser mayor a cero")
	}
	payload.UsuarioOrigen = strings.TrimSpace(payload.UsuarioOrigen)
	if payload.UsuarioOrigen == "" {
		payload.UsuarioOrigen = "sistema"
	}
	payload.UsuarioAsignado = strings.TrimSpace(payload.UsuarioAsignado)
	if payload.ModoDistribucion == EmpresaPropinaModoPorUsuario && payload.UsuarioAsignado == "" {
		payload.UsuarioAsignado = payload.UsuarioOrigen
	}
	payload.UsuarioOrigenID = resolveEmpresaUsuarioIDByReferenceSilent(dbConn, payload.EmpresaID, payload.UsuarioOrigenID, payload.UsuarioOrigen)
	if payload.ModoDistribucion == EmpresaPropinaModoPorUsuario {
		payload.UsuarioAsignadoID = resolveEmpresaUsuarioIDByReferenceSilent(dbConn, payload.EmpresaID, payload.UsuarioAsignadoID, payload.UsuarioAsignado)
	} else {
		payload.UsuarioAsignadoID = 0
	}
	payload.FiscalPais = strings.TrimSpace(payload.FiscalPais)
	payload.FiscalRegimen = strings.TrimSpace(payload.FiscalRegimen)
	payload.FiscalTratamiento = strings.TrimSpace(payload.FiscalTratamiento)
	if cfg != nil {
		if payload.FiscalPais == "" {
			payload.FiscalPais = cfg.PaisFiscal
		}
		if payload.FiscalRegimen == "" {
			payload.FiscalRegimen = cfg.RegimenFiscal
		}
		if payload.FiscalTratamiento == "" {
			payload.FiscalTratamiento = cfg.TratamientoFiscal
		}
		if payload.FiscalPorcentaje <= 0 {
			payload.FiscalPorcentaje = cfg.PorcentajeImpuesto
		}
	}
	payload.FiscalPais = normalizePropinaFiscalPais(payload.FiscalPais)
	payload.FiscalRegimen = normalizePropinaFiscalRegimen(payload.FiscalRegimen)
	payload.FiscalTratamiento = normalizePropinaFiscalTratamiento(payload.FiscalTratamiento)
	payload.FiscalPorcentaje = normalizePropinaPorcentaje(payload.FiscalPorcentaje)
	if payload.FiscalTratamiento != EmpresaPropinaTratamientoGravada {
		payload.FiscalPorcentaje = 0
	}
	payload.FiscalImpuesto = 0
	if payload.FiscalTratamiento == EmpresaPropinaTratamientoGravada && payload.FiscalPorcentaje > 0 {
		payload.FiscalImpuesto = round2(payload.MontoPropina * (payload.FiscalPorcentaje / 100.0))
	}
	payload.FiscalTotal = round2(payload.MontoPropina + payload.FiscalImpuesto)
	payload.ConciliadoEn = strings.TrimSpace(payload.ConciliadoEn)
	if strings.TrimSpace(payload.Estado) == "" {
		payload.Estado = "activo"
	}
	if payload.EsAjusteManual && strings.TrimSpace(payload.Observaciones) == "" {
		payload.Observaciones = "ajuste manual de propina"
	}

	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_propinas_movimientos (
		empresa_id,
		carrito_id,
		cierre_caja_id,
		venta_referencia,
		usuario_origen,
		usuario_origen_id,
		usuario_asignado,
		usuario_asignado_id,
		modo_distribucion,
		origen_movimiento,
		ajuste_manual,
		referencia_ajuste,
		moneda,
		base_cobro,
		porcentaje_propina,
		monto_propina,
		fiscal_pais,
		fiscal_regimen,
		fiscal_tratamiento,
		fiscal_porcentaje_impuesto,
		fiscal_impuesto_monto,
		fiscal_total,
		conciliado_en,
		fecha_movimiento,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, COALESCE(NULLIF(?, ''), datetime('now','localtime')), ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
		payload.EmpresaID,
		payload.CarritoID,
		payload.CierreCajaID,
		strings.TrimSpace(payload.VentaReferencia),
		payload.UsuarioOrigen,
		payload.UsuarioOrigenID,
		payload.UsuarioAsignado,
		payload.UsuarioAsignadoID,
		payload.ModoDistribucion,
		payload.OrigenMovimiento,
		boolToInt(payload.EsAjusteManual),
		payload.ReferenciaAjuste,
		payload.Moneda,
		payload.BaseCobro,
		payload.PorcentajePropina,
		payload.MontoPropina,
		payload.FiscalPais,
		payload.FiscalRegimen,
		payload.FiscalTratamiento,
		payload.FiscalPorcentaje,
		payload.FiscalImpuesto,
		payload.FiscalTotal,
		payload.ConciliadoEn,
		payload.FechaMovimiento,
		strings.TrimSpace(payload.UsuarioCreador),
		strings.TrimSpace(payload.Estado),
		strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// CreateEmpresaPropinaAjusteManual registra un ajuste manual de propina (positivo o negativo).
func CreateEmpresaPropinaAjusteManual(dbConn *sql.DB, payload EmpresaPropinaMovimiento) (int64, error) {
	payload.EsAjusteManual = true
	payload.OrigenMovimiento = EmpresaPropinaOrigenAjusteManual
	return CreateEmpresaPropinaMovimiento(dbConn, payload)
}

// ListEmpresaPropinaMovimientos lista movimientos de propinas por empresa.
func ListEmpresaPropinaMovimientos(dbConn *sql.DB, empresaID int64, filter EmpresaPropinaMovimientoFilter) ([]EmpresaPropinaMovimiento, error) {
	query := `SELECT
		id,
		empresa_id,
		COALESCE(carrito_id, 0),
		COALESCE(cierre_caja_id, 0),
		COALESCE(venta_referencia, ''),
		COALESCE(usuario_origen, ''),
		COALESCE(usuario_origen_id, 0),
		COALESCE(usuario_asignado, ''),
		COALESCE(usuario_asignado_id, 0),
		COALESCE(modo_distribucion, 'por_usuario'),
		COALESCE(origen_movimiento, 'venta'),
		COALESCE(ajuste_manual, 0),
		COALESCE(referencia_ajuste, ''),
		COALESCE(moneda, 'COP'),
		COALESCE(base_cobro, 0),
		COALESCE(porcentaje_propina, 0),
		COALESCE(monto_propina, 0),
		COALESCE(fiscal_pais, 'generico'),
		COALESCE(fiscal_regimen, 'general'),
		COALESCE(fiscal_tratamiento, 'no_gravada'),
		COALESCE(fiscal_porcentaje_impuesto, 0),
		COALESCE(fiscal_impuesto_monto, 0),
		COALESCE(fiscal_total, 0),
		COALESCE(conciliado_en, ''),
		COALESCE(fecha_movimiento, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_propinas_movimientos
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	query, args = appendPropinaCommonFilters(query, args, filter)

	if filter.Limit <= 0 {
		filter.Limit = 200
	}
	if filter.Limit > 2000 {
		filter.Limit = 2000
	}
	query += ` ORDER BY COALESCE(fecha_movimiento, fecha_creacion) DESC, id DESC LIMIT ?`
	args = append(args, filter.Limit)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaPropinaMovimiento, 0)
	for rows.Next() {
		var item EmpresaPropinaMovimiento
		var ajusteManualInt int
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.CarritoID,
			&item.CierreCajaID,
			&item.VentaReferencia,
			&item.UsuarioOrigen,
			&item.UsuarioOrigenID,
			&item.UsuarioAsignado,
			&item.UsuarioAsignadoID,
			&item.ModoDistribucion,
			&item.OrigenMovimiento,
			&ajusteManualInt,
			&item.ReferenciaAjuste,
			&item.Moneda,
			&item.BaseCobro,
			&item.PorcentajePropina,
			&item.MontoPropina,
			&item.FiscalPais,
			&item.FiscalRegimen,
			&item.FiscalTratamiento,
			&item.FiscalPorcentaje,
			&item.FiscalImpuesto,
			&item.FiscalTotal,
			&item.ConciliadoEn,
			&item.FechaMovimiento,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		item.ModoDistribucion = normalizePropinaModo(item.ModoDistribucion)
		item.OrigenMovimiento = normalizePropinaOrigen(item.OrigenMovimiento)
		if item.UsuarioOrigenID == 0 {
			item.UsuarioOrigenID = resolveEmpresaUsuarioIDByReferenceSilent(dbConn, item.EmpresaID, 0, item.UsuarioOrigen)
		}
		if item.UsuarioAsignadoID == 0 && item.ModoDistribucion == EmpresaPropinaModoPorUsuario {
			item.UsuarioAsignadoID = resolveEmpresaUsuarioIDByReferenceSilent(dbConn, item.EmpresaID, 0, item.UsuarioAsignado)
		}
		item.EsAjusteManual = ajusteManualInt == 1 || item.OrigenMovimiento == EmpresaPropinaOrigenAjusteManual
		item.Moneda = normalizePropinaMoneda(item.Moneda)
		item.BaseCobro = round2(item.BaseCobro)
		item.PorcentajePropina = normalizePropinaPorcentaje(item.PorcentajePropina)
		item.MontoPropina = round2(item.MontoPropina)
		item.FiscalPais = normalizePropinaFiscalPais(item.FiscalPais)
		item.FiscalRegimen = normalizePropinaFiscalRegimen(item.FiscalRegimen)
		item.FiscalTratamiento = normalizePropinaFiscalTratamiento(item.FiscalTratamiento)
		item.FiscalPorcentaje = normalizePropinaPorcentaje(item.FiscalPorcentaje)
		if item.FiscalTratamiento != EmpresaPropinaTratamientoGravada {
			item.FiscalPorcentaje = 0
		}
		item.FiscalImpuesto = round2(item.FiscalImpuesto)
		item.FiscalTotal = round2(item.FiscalTotal)
		if math.Abs(item.FiscalTotal) < 0.0001 {
			item.FiscalTotal = round2(item.MontoPropina + item.FiscalImpuesto)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// GetEmpresaPropinasReporte construye un reporte de propinas por empresa y periodo.
func GetEmpresaPropinasReporte(dbConn *sql.DB, empresaID int64, filter EmpresaPropinaMovimientoFilter) (*EmpresaPropinasReporte, error) {
	cfg, err := GetEmpresaPropinasConfiguracion(dbConn, empresaID)
	if err != nil {
		return nil, err
	}

	baseQuery := `SELECT
		COALESCE(SUM(base_cobro), 0),
		COALESCE(SUM(monto_propina), 0),
		COALESCE(SUM(CASE WHEN COALESCE(ajuste_manual, 0) = 1 THEN monto_propina ELSE 0 END), 0),
		COALESCE(SUM(COALESCE(fiscal_impuesto_monto, 0)), 0),
		COALESCE(SUM(CASE
			WHEN COALESCE(fiscal_total, 0) != 0 THEN COALESCE(fiscal_total, 0)
			ELSE COALESCE(monto_propina, 0) + COALESCE(fiscal_impuesto_monto, 0)
		END), 0),
		COALESCE(COUNT(1), 0),
		COALESCE(SUM(CASE WHEN COALESCE(modo_distribucion, 'por_usuario') = 'por_usuario' THEN monto_propina ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN COALESCE(modo_distribucion, 'por_usuario') = 'universal' THEN monto_propina ELSE 0 END), 0)
	FROM empresa_propinas_movimientos
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	baseQuery, args = appendPropinaCommonFilters(baseQuery, args, filter)

	var resumen EmpresaPropinasResumen
	if err := dbConn.QueryRow(baseQuery, args...).Scan(
		&resumen.TotalBaseCobro,
		&resumen.TotalPropinas,
		&resumen.TotalAjustesManuales,
		&resumen.TotalImpuestoPropina,
		&resumen.TotalPropinasConImpuesto,
		&resumen.CantidadMovimientos,
		&resumen.TotalPropinasPorUsuario,
		&resumen.TotalPropinasUniversal,
	); err != nil {
		return nil, err
	}
	resumen.TotalBaseCobro = round2(resumen.TotalBaseCobro)
	resumen.TotalPropinas = round2(resumen.TotalPropinas)
	resumen.TotalAjustesManuales = round2(resumen.TotalAjustesManuales)
	resumen.TotalImpuestoPropina = round2(resumen.TotalImpuestoPropina)
	resumen.TotalPropinasConImpuesto = round2(resumen.TotalPropinasConImpuesto)
	if math.Abs(resumen.TotalPropinasConImpuesto) < 0.0001 && math.Abs(resumen.TotalPropinas) > 0.0001 {
		resumen.TotalPropinasConImpuesto = round2(resumen.TotalPropinas + resumen.TotalImpuestoPropina)
	}
	resumen.TotalPropinasPorUsuario = round2(resumen.TotalPropinasPorUsuario)
	resumen.TotalPropinasUniversal = round2(resumen.TotalPropinasUniversal)

	directUsersQuery := `SELECT
		COALESCE(usuario_asignado_id, 0),
		COALESCE(NULLIF(TRIM(usuario_asignado), ''), NULLIF(TRIM(usuario_origen), ''), 'sistema') AS usuario,
		COALESCE(SUM(monto_propina), 0)
	FROM empresa_propinas_movimientos
	WHERE empresa_id = ? AND COALESCE(modo_distribucion, 'por_usuario') = 'por_usuario'`
	directArgs := []interface{}{empresaID}
	directUsersQuery, directArgs = appendPropinaCommonFilters(directUsersQuery, directArgs, filter)
	directUsersQuery += ` GROUP BY COALESCE(usuario_asignado_id, 0), usuario ORDER BY COALESCE(SUM(monto_propina), 0) DESC, usuario ASC`

	perUser := map[string]*EmpresaPropinaUsuarioResumen{}
	rows, err := dbConn.Query(directUsersQuery, directArgs...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var usuarioID int64
		var usuario string
		var total float64
		if err := rows.Scan(&usuarioID, &usuario, &total); err != nil {
			rows.Close()
			return nil, err
		}
		if usuarioID == 0 {
			usuarioID = resolveEmpresaUsuarioIDByReferenceSilent(dbConn, empresaID, 0, usuario)
		}
		clave := normalizePropinaUsuarioClave(usuario, "")
		if usuarioID > 0 {
			clave = fmt.Sprintf("usuario:%d", usuarioID)
		}
		item := &EmpresaPropinaUsuarioResumen{
			UsuarioID:         usuarioID,
			UsuarioClave:      clave,
			UsuarioEtiqueta:   strings.TrimSpace(usuario),
			PropinaPorUsuario: round2(total),
		}
		if item.UsuarioEtiqueta == "" {
			item.UsuarioEtiqueta = "sistema"
		}
		perUser[clave] = item
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	activeUsers, err := listActiveUsersForPropinas(dbConn, empresaID)
	if err != nil {
		return nil, err
	}
	resumen.UsuariosActivos = len(activeUsers)
	if resumen.TotalPropinasUniversal > 0 && len(activeUsers) > 0 {
		resumen.CuotaUniversalPorUsuario = round2(resumen.TotalPropinasUniversal / float64(len(activeUsers)))
	}

	for _, user := range activeUsers {
		entry, ok := perUser[user.Clave]
		if !ok {
			entry = &EmpresaPropinaUsuarioResumen{
				UsuarioID:       user.ID,
				UsuarioClave:    user.Clave,
				UsuarioEtiqueta: user.Etiqueta,
			}
			perUser[user.Clave] = entry
		}
		if entry.UsuarioID == 0 {
			entry.UsuarioID = user.ID
		}
		entry.EsUsuarioActivo = true
		if strings.TrimSpace(entry.UsuarioEtiqueta) == "" {
			entry.UsuarioEtiqueta = user.Etiqueta
		}
		entry.PropinaUniversal = round2(entry.PropinaUniversal + resumen.CuotaUniversalPorUsuario)
	}

	usuarios := make([]EmpresaPropinaUsuarioResumen, 0, len(perUser))
	for _, it := range perUser {
		it.PropinaPorUsuario = round2(it.PropinaPorUsuario)
		it.PropinaUniversal = round2(it.PropinaUniversal)
		it.PropinaTotal = round2(it.PropinaPorUsuario + it.PropinaUniversal)
		if strings.TrimSpace(it.UsuarioEtiqueta) == "" {
			it.UsuarioEtiqueta = "sistema"
		}
		usuarios = append(usuarios, *it)
	}
	sort.SliceStable(usuarios, func(i, j int) bool {
		if usuarios[i].PropinaTotal == usuarios[j].PropinaTotal {
			return usuarios[i].UsuarioEtiqueta < usuarios[j].UsuarioEtiqueta
		}
		return usuarios[i].PropinaTotal > usuarios[j].PropinaTotal
	})

	movementsFilter := filter
	if movementsFilter.Limit <= 0 {
		movementsFilter.Limit = 300
	}
	movs, err := ListEmpresaPropinaMovimientos(dbConn, empresaID, movementsFilter)
	if err != nil {
		return nil, err
	}

	report := &EmpresaPropinasReporte{
		EmpresaID:     empresaID,
		Desde:         strings.TrimSpace(filter.Desde),
		Hasta:         strings.TrimSpace(filter.Hasta),
		Configuracion: cfg,
		Resumen:       resumen,
		Usuarios:      usuarios,
		Movimientos:   movs,
	}
	return report, nil
}

// ConciliarEmpresaPropinasConCierreCaja consolida propinas del dia y las asocia a un cierre de caja.
func ConciliarEmpresaPropinasConCierreCaja(dbConn *sql.DB, empresaID, cierreCajaID int64, usuario string) (*EmpresaPropinaConciliacionCierre, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if cierreCajaID <= 0 {
		return nil, fmt.Errorf("cierre_caja_id es obligatorio")
	}
	if err := EnsureEmpresaPropinasSchema(dbConn); err != nil {
		return nil, err
	}
	if err := EnsureEmpresaFinanzasSchema(dbConn); err != nil {
		return nil, err
	}

	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = "sistema"
	}

	var fechaOperacion string
	var estadoCierre string
	err := dbConn.QueryRow(`SELECT
		COALESCE(fecha_operacion, ''),
		COALESCE(estado_cierre, 'abierto')
	FROM empresa_cierres_caja
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, cierreCajaID).Scan(&fechaOperacion, &estadoCierre)
	if err != nil {
		return nil, err
	}

	fechaOperacion = normalizeDateOnly(fechaOperacion)
	if fechaOperacion == "" {
		fechaOperacion = time.Now().Format("2006-01-02")
	}
	estadoCierre = normalizeEstadoCierre(estadoCierre)
	if estadoCierre == "anulado" {
		return nil, fmt.Errorf("no se puede conciliar propinas sobre un cierre anulado")
	}

	result := &EmpresaPropinaConciliacionCierre{
		EmpresaID:      empresaID,
		CierreCajaID:   cierreCajaID,
		FechaOperacion: fechaOperacion,
		ConciliadoPor:  usuario,
	}

	err = dbConn.QueryRow(`SELECT
		COALESCE(COUNT(1), 0),
		COALESCE(SUM(COALESCE(monto_propina, 0)), 0),
		COALESCE(SUM(CASE WHEN COALESCE(ajuste_manual, 0) = 1 THEN COALESCE(monto_propina, 0) ELSE 0 END), 0),
		COALESCE(SUM(COALESCE(fiscal_impuesto_monto, 0)), 0),
		COALESCE(SUM(CASE
			WHEN COALESCE(fiscal_total, 0) != 0 THEN COALESCE(fiscal_total, 0)
			ELSE COALESCE(monto_propina, 0) + COALESCE(fiscal_impuesto_monto, 0)
		END), 0)
	FROM empresa_propinas_movimientos
	WHERE empresa_id = ?
		AND LOWER(COALESCE(estado, 'activo')) = 'activo'
		AND COALESCE(cierre_caja_id, 0) IN (0, ?)
		AND date(COALESCE(fecha_movimiento, fecha_creacion)) = date(?)`,
		empresaID,
		cierreCajaID,
		fechaOperacion,
	).Scan(
		&result.CantidadMovimientos,
		&result.TotalPropinas,
		&result.TotalAjustes,
		&result.TotalImpuesto,
		&result.TotalNeto,
	)
	if err != nil {
		return nil, err
	}

	result.TotalPropinas = round2(result.TotalPropinas)
	result.TotalAjustes = round2(result.TotalAjustes)
	result.TotalImpuesto = round2(result.TotalImpuesto)
	result.TotalNeto = round2(result.TotalNeto)

	if _, err := dbConn.Exec(`UPDATE empresa_propinas_movimientos
	SET
		cierre_caja_id = ?,
		conciliado_en = datetime('now','localtime'),
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ?
		AND LOWER(COALESCE(estado, 'activo')) = 'activo'
		AND COALESCE(cierre_caja_id, 0) IN (0, ?)
		AND date(COALESCE(fecha_movimiento, fecha_creacion)) = date(?)`,
		cierreCajaID,
		empresaID,
		cierreCajaID,
		fechaOperacion,
	); err != nil {
		return nil, err
	}

	if _, err := dbConn.Exec(`UPDATE empresa_cierres_caja
	SET
		propinas_movimientos = ?,
		propinas_total = ?,
		propinas_ajustes = ?,
		propinas_impuesto = ?,
		propinas_neto = ?,
		propinas_conciliado_en = datetime('now','localtime'),
		propinas_conciliado_por = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`,
		result.CantidadMovimientos,
		result.TotalPropinas,
		result.TotalAjustes,
		result.TotalImpuesto,
		result.TotalNeto,
		usuario,
		empresaID,
		cierreCajaID,
	); err != nil {
		return nil, err
	}

	if err := dbConn.QueryRow(`SELECT COALESCE(propinas_conciliado_en, '')
	FROM empresa_cierres_caja
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, cierreCajaID).Scan(&result.ConciliadoEn); err != nil {
		return nil, err
	}

	return result, nil
}

func appendPropinaCommonFilters(query string, args []interface{}, filter EmpresaPropinaMovimientoFilter) (string, []interface{}) {
	if !filter.IncludeInactive {
		query += ` AND COALESCE(estado, 'activo') = 'activo'`
	}
	if desde := strings.TrimSpace(filter.Desde); desde != "" {
		query += ` AND date(COALESCE(fecha_movimiento, fecha_creacion)) >= date(?)`
		args = append(args, desde)
	}
	if hasta := strings.TrimSpace(filter.Hasta); hasta != "" {
		query += ` AND date(COALESCE(fecha_movimiento, fecha_creacion)) <= date(?)`
		args = append(args, hasta)
	}
	if modo := strings.TrimSpace(filter.ModoDistribucion); modo != "" {
		query += ` AND COALESCE(modo_distribucion, 'por_usuario') = ?`
		args = append(args, normalizePropinaModo(modo))
	}
	if filter.SoloAjustes {
		query += ` AND COALESCE(ajuste_manual, 0) = 1`
	}
	if filter.CierreCajaID > 0 {
		query += ` AND COALESCE(cierre_caja_id, 0) = ?`
		args = append(args, filter.CierreCajaID)
	}
	if origen := strings.TrimSpace(filter.OrigenMovimiento); origen != "" {
		query += ` AND COALESCE(origen_movimiento, 'venta') = ?`
		args = append(args, normalizePropinaOrigen(origen))
	}
	if usuario := strings.TrimSpace(strings.ToLower(filter.Usuario)); usuario != "" {
		like := "%" + usuario + "%"
		query += ` AND (
			LOWER(COALESCE(usuario_asignado, '')) LIKE ?
			OR LOWER(COALESCE(usuario_origen, '')) LIKE ?
			OR CAST(COALESCE(usuario_asignado_id, 0) AS TEXT) = ?
			OR CAST(COALESCE(usuario_origen_id, 0) AS TEXT) = ?
		)`
		args = append(args, like, like, usuario, usuario)
	}
	return query, args
}

func listActiveUsersForPropinas(dbConn *sql.DB, empresaID int64) ([]propinaActiveUser, error) {
	rows, err := dbConn.Query(`SELECT
		id,
		COALESCE(email, ''),
		COALESCE(name, '')
	FROM users
	WHERE empresa_id = ? AND COALESCE(estado, 'activo') = 'activo'
	ORDER BY COALESCE(name, ''), COALESCE(email, '')`, empresaID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such table") {
			return []propinaActiveUser{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	out := make([]propinaActiveUser, 0)
	seen := map[string]bool{}
	for rows.Next() {
		var id int64
		var email string
		var name string
		if err := rows.Scan(&id, &email, &name); err != nil {
			return nil, err
		}
		key := ""
		if id > 0 {
			key = fmt.Sprintf("usuario:%d", id)
		}
		if key == "" {
			key = normalizePropinaUsuarioClave(email, name)
		}
		if key == "" {
			continue
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, propinaActiveUser{
			ID:       id,
			Clave:    key,
			Etiqueta: normalizePropinaUsuarioEtiqueta(email, name),
		})
	}
	return out, rows.Err()
}

func normalizePropinaUsuarioClave(email, name string) string {
	email = strings.ToLower(strings.TrimSpace(email))
	if email != "" {
		return email
	}
	name = strings.ToLower(strings.TrimSpace(name))
	if name != "" {
		return name
	}
	return "sistema"
}

func normalizePropinaUsuarioEtiqueta(email, name string) string {
	email = strings.TrimSpace(email)
	name = strings.TrimSpace(name)
	if name != "" && email != "" {
		return name + " (" + email + ")"
	}
	if name != "" {
		return name
	}
	if email != "" {
		return email
	}
	return "sistema"
}
