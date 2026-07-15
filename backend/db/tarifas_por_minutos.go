package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// EmpresaTarifaPorMinutos define la regla de cobro por permanencia para una estacion.
type EmpresaTarifaPorMinutos struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	EstacionID         int64   `json:"estacion_id"`
	EstacionCodigo     string  `json:"estacion_codigo,omitempty"`
	EstacionNombre     string  `json:"estacion_nombre,omitempty"`
	DiaSemanaDesde     int     `json:"dia_semana_desde"`
	DiaSemanaHasta     int     `json:"dia_semana_hasta"`
	MinutosBase        int     `json:"minutos_base"`
	ValorBase          float64 `json:"valor_base"`
	MinutosExtra       int     `json:"minutos_extra"`
	ValorExtra         float64 `json:"valor_extra"`
	CobrarPorFraccion  bool    `json:"cobrar_por_fraccion"`
	Moneda             string  `json:"moneda,omitempty"`
	Prioridad          int     `json:"prioridad"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// EmpresaTarifaPorMinutosFilter define filtros de consulta de tarifas por minutos.
type EmpresaTarifaPorMinutosFilter struct {
	EstacionID      int64
	DiaSemana       int
	IncludeInactive bool
	Limit           int
}

const (
	tarifaPorMinutosRedondeoNinguno    = "ninguno"
	tarifaPorMinutosRedondeoArriba     = "arriba"
	tarifaPorMinutosRedondeoAbajo      = "abajo"
	tarifaPorMinutosRedondeoMatematico = "matematico"
)

// EmpresaTarifaPorMinutosConfiguracion define reglas globales de calculo por empresa.
type EmpresaTarifaPorMinutosConfiguracion struct {
	ID                             int64   `json:"id"`
	EmpresaID                      int64   `json:"empresa_id"`
	RedondeoModo                   string  `json:"redondeo_modo"`
	RedondeoUnidad                 float64 `json:"redondeo_unidad"`
	MontoMinimoDiario              float64 `json:"monto_minimo_diario"`
	MontoMaximoDiario              float64 `json:"monto_maximo_diario"`
	MargenToleranciaEntradaMinutos int     `json:"margen_tolerancia_entrada_minutos"`
	SensorAutoActivarEstacion      bool    `json:"sensor_auto_activar_estacion"`
	MargenDesactivacionHabilitado  bool    `json:"margen_desactivacion_habilitado"`
	MargenDesactivacionMinutos     int     `json:"margen_desactivacion_minutos"`
	FechaCreacion                  string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion             string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador                 string  `json:"usuario_creador,omitempty"`
	Estado                         string  `json:"estado,omitempty"`
	Observaciones                  string  `json:"observaciones,omitempty"`
}

// EmpresaTarifaPorMinutosCalculo representa el detalle de calculo por consumo.
type EmpresaTarifaPorMinutosCalculo struct {
	TarifaID            int64   `json:"tarifa_id"`
	EstacionID          int64   `json:"estacion_id"`
	DiaSemana           int     `json:"dia_semana"`
	MinutosConsumidos   float64 `json:"minutos_consumidos"`
	MinutosFacturables  float64 `json:"minutos_facturables"`
	MinutosTolerancia   int     `json:"minutos_tolerancia"`
	BloquesExtra        int     `json:"bloques_extra"`
	MontoBase           float64 `json:"monto_base"`
	MontoExtra          float64 `json:"monto_extra"`
	MontoSubtotal       float64 `json:"monto_subtotal"`
	MontoRedondeado     float64 `json:"monto_redondeado"`
	AjusteRedondeo      float64 `json:"ajuste_redondeo"`
	MontoMinimoAplicado bool    `json:"monto_minimo_aplicado"`
	MontoMaximoAplicado bool    `json:"monto_maximo_aplicado"`
	MontoTotal          float64 `json:"monto_total"`
	Moneda              string  `json:"moneda"`
}

// EmpresaTarifaPorMinutosAplicacionMasivaResultado resume la aplicacion masiva por estaciones.
type EmpresaTarifaPorMinutosAplicacionMasivaResultado struct {
	EmpresaID           int64   `json:"empresa_id"`
	DiaSemanaDesde      int     `json:"dia_semana_desde"`
	DiaSemanaHasta      int     `json:"dia_semana_hasta"`
	EstacionesObjetivo  int     `json:"estaciones_objetivo"`
	TarifasCreadas      int     `json:"tarifas_creadas"`
	TarifasActualizadas int     `json:"tarifas_actualizadas"`
	TarifaIDs           []int64 `json:"tarifa_ids"`
}

type empresaTarifaPorMinutosEstacionRef struct {
	ID     int64
	Codigo string
	Nombre string
}

// EnsureEmpresaTarifasPorMinutosSchema crea/migra tablas de tarifas por minutos por estacion.
func EnsureEmpresaTarifasPorMinutosSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_tarifas_por_minutos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			estacion_id INTEGER NOT NULL,
			estacion_codigo TEXT,
			estacion_nombre TEXT,
			dia_semana_desde INTEGER NOT NULL DEFAULT 1,
			dia_semana_hasta INTEGER NOT NULL DEFAULT 7,
			minutos_base INTEGER NOT NULL DEFAULT 120,
			valor_base REAL NOT NULL DEFAULT 0,
			minutos_extra INTEGER NOT NULL DEFAULT 60,
			valor_extra REAL NOT NULL DEFAULT 0,
			cobrar_por_fraccion INTEGER NOT NULL DEFAULT 0,
			moneda TEXT DEFAULT 'COP',
			prioridad INTEGER DEFAULT 1,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_tarifas_por_minutos_estacion_rango ON empresa_tarifas_por_minutos(empresa_id, estacion_id, dia_semana_desde, dia_semana_hasta);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_tarifas_por_minutos_empresa_estacion_estado ON empresa_tarifas_por_minutos(empresa_id, estacion_id, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_tarifas_por_minutos_empresa_dias ON empresa_tarifas_por_minutos(empresa_id, dia_semana_desde, dia_semana_hasta);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos", "estacion_codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos", "estacion_nombre", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos", "dia_semana_desde", "INTEGER NOT NULL DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos", "dia_semana_hasta", "INTEGER NOT NULL DEFAULT 7"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos", "minutos_base", "INTEGER NOT NULL DEFAULT 120"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos", "valor_base", "REAL NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos", "minutos_extra", "INTEGER NOT NULL DEFAULT 60"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos", "valor_extra", "REAL NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos", "cobrar_por_fraccion", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos", "moneda", "TEXT DEFAULT 'COP'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos", "prioridad", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos", "observaciones", "TEXT"); err != nil {
		return err
	}
	if err := EnsureEmpresaTarifasPorMinutosConfiguracionSchema(dbConn); err != nil {
		return err
	}

	return nil
}

// EnsureEmpresaTarifasPorMinutosConfiguracionSchema crea/migra configuracion de calculo por empresa.
func EnsureEmpresaTarifasPorMinutosConfiguracionSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_tarifas_por_minutos_configuracion (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			redondeo_modo TEXT NOT NULL DEFAULT 'ninguno',
			redondeo_unidad REAL NOT NULL DEFAULT 100,
			monto_minimo_diario REAL NOT NULL DEFAULT 0,
			monto_maximo_diario REAL NOT NULL DEFAULT 0,
			margen_tolerancia_entrada_minutos INTEGER NOT NULL DEFAULT 0,
			sensor_auto_activar_estacion INTEGER NOT NULL DEFAULT 0,
			margen_desactivacion_habilitado INTEGER NOT NULL DEFAULT 0,
			margen_desactivacion_minutos INTEGER NOT NULL DEFAULT 0,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_tarifas_por_minutos_cfg_empresa_estado ON empresa_tarifas_por_minutos_configuracion(empresa_id, estado);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos_configuracion", "redondeo_modo", "TEXT NOT NULL DEFAULT 'ninguno'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos_configuracion", "redondeo_unidad", "REAL NOT NULL DEFAULT 100"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos_configuracion", "monto_minimo_diario", "REAL NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos_configuracion", "monto_maximo_diario", "REAL NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos_configuracion", "margen_tolerancia_entrada_minutos", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos_configuracion", "sensor_auto_activar_estacion", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos_configuracion", "margen_desactivacion_habilitado", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos_configuracion", "margen_desactivacion_minutos", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos_configuracion", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos_configuracion", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos_configuracion", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_minutos_configuracion", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

func defaultEmpresaTarifaPorMinutosConfiguracion(empresaID int64) EmpresaTarifaPorMinutosConfiguracion {
	return EmpresaTarifaPorMinutosConfiguracion{
		EmpresaID:                      empresaID,
		RedondeoModo:                   tarifaPorMinutosRedondeoNinguno,
		RedondeoUnidad:                 100,
		MontoMinimoDiario:              0,
		MontoMaximoDiario:              0,
		MargenToleranciaEntradaMinutos: 0,
		SensorAutoActivarEstacion:      false,
		MargenDesactivacionHabilitado:  false,
		MargenDesactivacionMinutos:     0,
		Estado:                         "activo",
	}
}

func normalizeTarifaDiaSemana(v int) (int, error) {
	if v < 1 || v > 7 {
		return 0, fmt.Errorf("dia_semana debe estar entre 1 y 7")
	}
	return v, nil
}

func normalizeTarifaDiasRange(desde, hasta int) (int, int, error) {
	if desde == 0 {
		desde = 1
	}
	if hasta == 0 {
		hasta = 7
	}
	var err error
	desde, err = normalizeTarifaDiaSemana(desde)
	if err != nil {
		return 0, 0, fmt.Errorf("dia_semana_desde invalido")
	}
	hasta, err = normalizeTarifaDiaSemana(hasta)
	if err != nil {
		return 0, 0, fmt.Errorf("dia_semana_hasta invalido")
	}
	return desde, hasta, nil
}

func normalizeTarifaEstado(estado string) string {
	if strings.EqualFold(strings.TrimSpace(estado), "inactivo") {
		return "inactivo"
	}
	return "activo"
}

func normalizeTarifaMoneda(moneda string) string {
	m := strings.ToUpper(strings.TrimSpace(moneda))
	if m == "" {
		return "COP"
	}
	return m
}

func normalizeTarifaPrioridad(v int) int {
	if v <= 0 {
		return 1
	}
	if v > 999 {
		return 999
	}
	return v
}

func normalizeTarifaPorMinutosRedondeoModo(v string) string {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case tarifaPorMinutosRedondeoArriba:
		return tarifaPorMinutosRedondeoArriba
	case tarifaPorMinutosRedondeoAbajo:
		return tarifaPorMinutosRedondeoAbajo
	case tarifaPorMinutosRedondeoMatematico:
		return tarifaPorMinutosRedondeoMatematico
	default:
		return tarifaPorMinutosRedondeoNinguno
	}
}

func normalizeTarifaPorMinutosRedondeoUnidad(v float64) float64 {
	if v <= 0 {
		return 100
	}
	if v > 1000000 {
		return 1000000
	}
	return round2(v)
}

func normalizeTarifaPorMinutosMargin(v int) int {
	if v < 0 {
		return 0
	}
	if v > 1440 {
		return 1440
	}
	return v
}

func normalizeEmpresaTarifaPorMinutosConfiguracionPayload(payload *EmpresaTarifaPorMinutosConfiguracion) error {
	if payload == nil {
		return fmt.Errorf("payload invalido")
	}
	if payload.EmpresaID <= 0 {
		return fmt.Errorf("empresa_id es obligatorio")
	}

	payload.RedondeoModo = normalizeTarifaPorMinutosRedondeoModo(payload.RedondeoModo)
	payload.RedondeoUnidad = normalizeTarifaPorMinutosRedondeoUnidad(payload.RedondeoUnidad)
	payload.MontoMinimoDiario = round2(payload.MontoMinimoDiario)
	payload.MontoMaximoDiario = round2(payload.MontoMaximoDiario)
	if payload.MontoMinimoDiario < 0 {
		return fmt.Errorf("monto_minimo_diario no puede ser negativo")
	}
	if payload.MontoMaximoDiario < 0 {
		return fmt.Errorf("monto_maximo_diario no puede ser negativo")
	}
	if payload.MontoMaximoDiario > 0 && payload.MontoMinimoDiario > payload.MontoMaximoDiario {
		return fmt.Errorf("monto_minimo_diario no puede ser mayor que monto_maximo_diario")
	}
	payload.MargenToleranciaEntradaMinutos = normalizeTarifaPorMinutosMargin(payload.MargenToleranciaEntradaMinutos)
	payload.MargenDesactivacionMinutos = normalizeTarifaPorMinutosMargin(payload.MargenDesactivacionMinutos)
	if payload.MargenDesactivacionMinutos == 0 {
		payload.MargenDesactivacionHabilitado = false
	}
	payload.UsuarioCreador = strings.TrimSpace(payload.UsuarioCreador)
	payload.Estado = normalizeTarifaEstado(payload.Estado)
	payload.Observaciones = strings.TrimSpace(payload.Observaciones)

	return nil
}

func normalizeEmpresaTarifaPayload(payload *EmpresaTarifaPorMinutos) error {
	if payload == nil {
		return fmt.Errorf("payload invalido")
	}
	if payload.EmpresaID <= 0 {
		return fmt.Errorf("empresa_id es obligatorio")
	}
	if payload.EstacionID <= 0 {
		return fmt.Errorf("estacion_id es obligatorio")
	}

	var err error
	payload.DiaSemanaDesde, payload.DiaSemanaHasta, err = normalizeTarifaDiasRange(payload.DiaSemanaDesde, payload.DiaSemanaHasta)
	if err != nil {
		return err
	}

	if payload.MinutosBase <= 0 {
		return fmt.Errorf("minutos_base debe ser mayor a cero")
	}
	if payload.MinutosExtra <= 0 {
		return fmt.Errorf("minutos_extra debe ser mayor a cero")
	}
	if payload.ValorBase < 0 {
		return fmt.Errorf("valor_base no puede ser negativo")
	}
	if payload.ValorExtra < 0 {
		return fmt.Errorf("valor_extra no puede ser negativo")
	}

	payload.EstacionCodigo = strings.TrimSpace(payload.EstacionCodigo)
	payload.EstacionNombre = strings.TrimSpace(payload.EstacionNombre)
	payload.Moneda = normalizeTarifaMoneda(payload.Moneda)
	payload.Prioridad = normalizeTarifaPrioridad(payload.Prioridad)
	payload.Estado = normalizeTarifaEstado(payload.Estado)
	payload.ValorBase = round2(payload.ValorBase)
	payload.ValorExtra = round2(payload.ValorExtra)
	payload.UsuarioCreador = strings.TrimSpace(payload.UsuarioCreador)
	payload.Observaciones = strings.TrimSpace(payload.Observaciones)
	return nil
}

func applyTarifaPorMinutosRounding(value float64, cfg EmpresaTarifaPorMinutosConfiguracion) float64 {
	value = round2(value)
	unidad := normalizeTarifaPorMinutosRedondeoUnidad(cfg.RedondeoUnidad)
	if unidad <= 0 {
		return value
	}

	modo := normalizeTarifaPorMinutosRedondeoModo(cfg.RedondeoModo)
	if modo == tarifaPorMinutosRedondeoNinguno {
		return value
	}

	ratio := value / unidad
	switch modo {
	case tarifaPorMinutosRedondeoArriba:
		return round2(math.Ceil(ratio) * unidad)
	case tarifaPorMinutosRedondeoAbajo:
		return round2(math.Floor(ratio) * unidad)
	case tarifaPorMinutosRedondeoMatematico:
		return round2(math.Round(ratio) * unidad)
	default:
		return value
	}
}

func diaSemanaInRange(dia, desde, hasta int) bool {
	if dia < 1 || dia > 7 {
		return false
	}
	if desde <= hasta {
		return dia >= desde && dia <= hasta
	}
	return dia >= desde || dia <= hasta
}

// DayOfWeekISO devuelve dia de la semana en formato ISO: lunes=1 ... domingo=7.
func DayOfWeekISO(t time.Time) int {
	wd := int(t.Weekday())
	if wd == 0 {
		return 7
	}
	return wd
}

// CreateEmpresaTarifaPorMinutos crea una tarifa por minutos para una estacion.
func CreateEmpresaTarifaPorMinutos(dbConn *sql.DB, payload EmpresaTarifaPorMinutos) (int64, error) {
	if err := normalizeEmpresaTarifaPayload(&payload); err != nil {
		return 0, err
	}

	return insertSQLCompat(dbConn, `INSERT INTO empresa_tarifas_por_minutos (
		empresa_id,
		estacion_id,
		estacion_codigo,
		estacion_nombre,
		dia_semana_desde,
		dia_semana_hasta,
		minutos_base,
		valor_base,
		minutos_extra,
		valor_extra,
		cobrar_por_fraccion,
		moneda,
		prioridad,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		payload.EmpresaID,
		payload.EstacionID,
		payload.EstacionCodigo,
		payload.EstacionNombre,
		payload.DiaSemanaDesde,
		payload.DiaSemanaHasta,
		payload.MinutosBase,
		payload.ValorBase,
		payload.MinutosExtra,
		payload.ValorExtra,
		boolToInt(payload.CobrarPorFraccion),
		payload.Moneda,
		payload.Prioridad,
		payload.UsuarioCreador,
		payload.Estado,
		payload.Observaciones,
	)
}

// UpdateEmpresaTarifaPorMinutos actualiza una tarifa por minutos existente.
func UpdateEmpresaTarifaPorMinutos(dbConn *sql.DB, payload EmpresaTarifaPorMinutos) error {
	if payload.ID <= 0 {
		return fmt.Errorf("id es obligatorio")
	}
	if err := normalizeEmpresaTarifaPayload(&payload); err != nil {
		return err
	}

	res, err := dbConn.Exec(`UPDATE empresa_tarifas_por_minutos
	SET
		estacion_id = ?,
		estacion_codigo = ?,
		estacion_nombre = ?,
		dia_semana_desde = ?,
		dia_semana_hasta = ?,
		minutos_base = ?,
		valor_base = ?,
		minutos_extra = ?,
		valor_extra = ?,
		cobrar_por_fraccion = ?,
		moneda = ?,
		prioridad = ?,
		usuario_creador = ?,
		estado = ?,
		observaciones = ?,
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ? AND id = ?`,
		payload.EstacionID,
		payload.EstacionCodigo,
		payload.EstacionNombre,
		payload.DiaSemanaDesde,
		payload.DiaSemanaHasta,
		payload.MinutosBase,
		payload.ValorBase,
		payload.MinutosExtra,
		payload.ValorExtra,
		boolToInt(payload.CobrarPorFraccion),
		payload.Moneda,
		payload.Prioridad,
		payload.UsuarioCreador,
		payload.Estado,
		payload.Observaciones,
		payload.EmpresaID,
		payload.ID,
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

// SetEmpresaTarifaPorMinutosEstado activa o desactiva una tarifa por minutos.
func SetEmpresaTarifaPorMinutosEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	if empresaID <= 0 || id <= 0 {
		return fmt.Errorf("empresa_id e id son obligatorios")
	}
	nextEstado := normalizeTarifaEstado(estado)
	res, err := dbConn.Exec(`UPDATE empresa_tarifas_por_minutos
	SET estado = ?, fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ? AND id = ?`, nextEstado, empresaID, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteEmpresaTarifaPorMinutos elimina una tarifa por minutos.
func DeleteEmpresaTarifaPorMinutos(dbConn *sql.DB, empresaID, id int64) error {
	if empresaID <= 0 || id <= 0 {
		return fmt.Errorf("empresa_id e id son obligatorios")
	}
	res, err := dbConn.Exec(`DELETE FROM empresa_tarifas_por_minutos WHERE empresa_id = ? AND id = ?`, empresaID, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetEmpresaTarifaPorMinutosByID obtiene una tarifa puntual por id y empresa.
func GetEmpresaTarifaPorMinutosByID(dbConn *sql.DB, empresaID, id int64) (*EmpresaTarifaPorMinutos, error) {
	if empresaID <= 0 || id <= 0 {
		return nil, fmt.Errorf("empresa_id e id son obligatorios")
	}
	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		estacion_id,
		COALESCE(estacion_codigo, ''),
		COALESCE(estacion_nombre, ''),
		COALESCE(dia_semana_desde, 1),
		COALESCE(dia_semana_hasta, 7),
		COALESCE(minutos_base, 120),
		COALESCE(valor_base, 0),
		COALESCE(minutos_extra, 60),
		COALESCE(valor_extra, 0),
		COALESCE(cobrar_por_fraccion, 0),
		COALESCE(moneda, 'COP'),
		COALESCE(prioridad, 1),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_tarifas_por_minutos
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, id)

	var item EmpresaTarifaPorMinutos
	var cobrarPorFraccion int
	if err := row.Scan(
		&item.ID,
		&item.EmpresaID,
		&item.EstacionID,
		&item.EstacionCodigo,
		&item.EstacionNombre,
		&item.DiaSemanaDesde,
		&item.DiaSemanaHasta,
		&item.MinutosBase,
		&item.ValorBase,
		&item.MinutosExtra,
		&item.ValorExtra,
		&cobrarPorFraccion,
		&item.Moneda,
		&item.Prioridad,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	); err != nil {
		return nil, err
	}
	item.CobrarPorFraccion = cobrarPorFraccion > 0
	return &item, nil
}

// ListEmpresaTarifasPorMinutos lista tarifas por empresa con filtros operativos.
func ListEmpresaTarifasPorMinutos(dbConn *sql.DB, empresaID int64, filter EmpresaTarifaPorMinutosFilter) ([]EmpresaTarifaPorMinutos, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if filter.Limit <= 0 {
		filter.Limit = 300
	}
	if filter.Limit > 2000 {
		filter.Limit = 2000
	}

	query := `SELECT
		id,
		empresa_id,
		estacion_id,
		COALESCE(estacion_codigo, ''),
		COALESCE(estacion_nombre, ''),
		COALESCE(dia_semana_desde, 1),
		COALESCE(dia_semana_hasta, 7),
		COALESCE(minutos_base, 120),
		COALESCE(valor_base, 0),
		COALESCE(minutos_extra, 60),
		COALESCE(valor_extra, 0),
		COALESCE(cobrar_por_fraccion, 0),
		COALESCE(moneda, 'COP'),
		COALESCE(prioridad, 1),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_tarifas_por_minutos
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	if !filter.IncludeInactive {
		query += ` AND COALESCE(estado, 'activo') = 'activo'`
	}
	if filter.EstacionID > 0 {
		query += ` AND estacion_id = ?`
		args = append(args, filter.EstacionID)
	}
	if filter.DiaSemana > 0 {
		if _, err := normalizeTarifaDiaSemana(filter.DiaSemana); err != nil {
			return nil, err
		}
		query += ` AND ((? BETWEEN dia_semana_desde AND dia_semana_hasta) OR (dia_semana_desde > dia_semana_hasta AND (? >= dia_semana_desde OR ? <= dia_semana_hasta)))`
		args = append(args, filter.DiaSemana, filter.DiaSemana, filter.DiaSemana)
	}

	query += ` ORDER BY estacion_id ASC, prioridad ASC, dia_semana_desde ASC, id ASC LIMIT ?`
	args = append(args, filter.Limit)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaTarifaPorMinutos, 0)
	for rows.Next() {
		var item EmpresaTarifaPorMinutos
		var cobrarPorFraccion int
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.EstacionID,
			&item.EstacionCodigo,
			&item.EstacionNombre,
			&item.DiaSemanaDesde,
			&item.DiaSemanaHasta,
			&item.MinutosBase,
			&item.ValorBase,
			&item.MinutosExtra,
			&item.ValorExtra,
			&cobrarPorFraccion,
			&item.Moneda,
			&item.Prioridad,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		item.CobrarPorFraccion = cobrarPorFraccion > 0
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// GetEmpresaTarifaPorMinutosAplicable resuelve la tarifa activa para estacion y dia de semana.
func GetEmpresaTarifaPorMinutosAplicable(dbConn *sql.DB, empresaID, estacionID int64, diaSemana int) (*EmpresaTarifaPorMinutos, error) {
	if empresaID <= 0 || estacionID <= 0 {
		return nil, fmt.Errorf("empresa_id y estacion_id son obligatorios")
	}
	if diaSemana == 0 {
		diaSemana = DayOfWeekISO(time.Now())
	}
	if _, err := normalizeTarifaDiaSemana(diaSemana); err != nil {
		return nil, err
	}

	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		estacion_id,
		COALESCE(estacion_codigo, ''),
		COALESCE(estacion_nombre, ''),
		COALESCE(dia_semana_desde, 1),
		COALESCE(dia_semana_hasta, 7),
		COALESCE(minutos_base, 120),
		COALESCE(valor_base, 0),
		COALESCE(minutos_extra, 60),
		COALESCE(valor_extra, 0),
		COALESCE(cobrar_por_fraccion, 0),
		COALESCE(moneda, 'COP'),
		COALESCE(prioridad, 1),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_tarifas_por_minutos
	WHERE empresa_id = ?
		AND estacion_id = ?
		AND COALESCE(estado, 'activo') = 'activo'
		AND ((? BETWEEN dia_semana_desde AND dia_semana_hasta) OR (dia_semana_desde > dia_semana_hasta AND (? >= dia_semana_desde OR ? <= dia_semana_hasta)))
	ORDER BY prioridad ASC, id ASC
	LIMIT 1`, empresaID, estacionID, diaSemana, diaSemana, diaSemana)

	var item EmpresaTarifaPorMinutos
	var cobrarPorFraccion int
	if err := row.Scan(
		&item.ID,
		&item.EmpresaID,
		&item.EstacionID,
		&item.EstacionCodigo,
		&item.EstacionNombre,
		&item.DiaSemanaDesde,
		&item.DiaSemanaHasta,
		&item.MinutosBase,
		&item.ValorBase,
		&item.MinutosExtra,
		&item.ValorExtra,
		&cobrarPorFraccion,
		&item.Moneda,
		&item.Prioridad,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	item.CobrarPorFraccion = cobrarPorFraccion > 0
	return &item, nil
}

// GetEmpresaTarifaPorMinutosConfiguracion obtiene la configuracion de calculo para una empresa.
func GetEmpresaTarifaPorMinutosConfiguracion(dbConn *sql.DB, empresaID int64) (*EmpresaTarifaPorMinutosConfiguracion, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaTarifasPorMinutosConfiguracionSchema(dbConn); err != nil {
		return nil, err
	}

	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(redondeo_modo, 'ninguno'),
		COALESCE(redondeo_unidad, 100),
		COALESCE(monto_minimo_diario, 0),
		COALESCE(monto_maximo_diario, 0),
		COALESCE(margen_tolerancia_entrada_minutos, 0),
		COALESCE(sensor_auto_activar_estacion, 0),
		COALESCE(margen_desactivacion_habilitado, 0),
		COALESCE(margen_desactivacion_minutos, 0),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_tarifas_por_minutos_configuracion
	WHERE empresa_id = ?
	LIMIT 1`, empresaID)

	var item EmpresaTarifaPorMinutosConfiguracion
	var sensorAutoActivarEstacion int
	var margenDesactivacionHabilitado int
	if err := row.Scan(
		&item.ID,
		&item.EmpresaID,
		&item.RedondeoModo,
		&item.RedondeoUnidad,
		&item.MontoMinimoDiario,
		&item.MontoMaximoDiario,
		&item.MargenToleranciaEntradaMinutos,
		&sensorAutoActivarEstacion,
		&margenDesactivacionHabilitado,
		&item.MargenDesactivacionMinutos,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	); err != nil {
		if err == sql.ErrNoRows {
			def := defaultEmpresaTarifaPorMinutosConfiguracion(empresaID)
			return &def, nil
		}
		return nil, err
	}
	item.SensorAutoActivarEstacion = sensorAutoActivarEstacion > 0
	item.MargenDesactivacionHabilitado = margenDesactivacionHabilitado > 0
	if err := normalizeEmpresaTarifaPorMinutosConfiguracionPayload(&item); err != nil {
		return nil, err
	}
	return &item, nil
}

// UpsertEmpresaTarifaPorMinutosConfiguracion crea/actualiza configuracion de calculo por empresa.
func UpsertEmpresaTarifaPorMinutosConfiguracion(dbConn *sql.DB, payload EmpresaTarifaPorMinutosConfiguracion) (*EmpresaTarifaPorMinutosConfiguracion, error) {
	if err := normalizeEmpresaTarifaPorMinutosConfiguracionPayload(&payload); err != nil {
		return nil, err
	}
	if err := EnsureEmpresaTarifasPorMinutosConfiguracionSchema(dbConn); err != nil {
		return nil, err
	}

	_, err := dbConn.Exec(`INSERT INTO empresa_tarifas_por_minutos_configuracion (
		empresa_id,
		redondeo_modo,
		redondeo_unidad,
		monto_minimo_diario,
		monto_maximo_diario,
		margen_tolerancia_entrada_minutos,
		sensor_auto_activar_estacion,
		margen_desactivacion_habilitado,
		margen_desactivacion_minutos,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT(empresa_id) DO UPDATE SET
		redondeo_modo = excluded.redondeo_modo,
		redondeo_unidad = excluded.redondeo_unidad,
		monto_minimo_diario = excluded.monto_minimo_diario,
		monto_maximo_diario = excluded.monto_maximo_diario,
		margen_tolerancia_entrada_minutos = excluded.margen_tolerancia_entrada_minutos,
		sensor_auto_activar_estacion = excluded.sensor_auto_activar_estacion,
		margen_desactivacion_habilitado = excluded.margen_desactivacion_habilitado,
		margen_desactivacion_minutos = excluded.margen_desactivacion_minutos,
		usuario_creador = excluded.usuario_creador,
		estado = excluded.estado,
		observaciones = excluded.observaciones,
		fecha_actualizacion = CURRENT_TIMESTAMP`,
		payload.EmpresaID,
		payload.RedondeoModo,
		payload.RedondeoUnidad,
		payload.MontoMinimoDiario,
		payload.MontoMaximoDiario,
		payload.MargenToleranciaEntradaMinutos,
		boolToInt(payload.SensorAutoActivarEstacion),
		boolToInt(payload.MargenDesactivacionHabilitado),
		payload.MargenDesactivacionMinutos,
		payload.UsuarioCreador,
		payload.Estado,
		payload.Observaciones,
	)
	if err != nil {
		return nil, err
	}
	return GetEmpresaTarifaPorMinutosConfiguracion(dbConn, payload.EmpresaID)
}

// CalcularDetalleTarifaPorMinutos calcula montos base/extra/redondeo/limites para una tarifa.
func CalcularDetalleTarifaPorMinutos(tarifa EmpresaTarifaPorMinutos, minutosConsumidos float64, cfg EmpresaTarifaPorMinutosConfiguracion) EmpresaTarifaPorMinutosCalculo {
	if minutosConsumidos <= 0 {
		minutosConsumidos = float64(tarifa.MinutosBase)
	}
	if minutosConsumidos < 0 {
		minutosConsumidos = 0
	}

	montoBase := round2(tarifa.ValorBase)
	minutosFacturables := minutosConsumidos
	margenTolerancia := normalizeTarifaPorMinutosMargin(cfg.MargenToleranciaEntradaMinutos)
	if margenTolerancia > 0 && minutosFacturables > float64(tarifa.MinutosBase) {
		minutosFacturables -= float64(margenTolerancia)
		if minutosFacturables < float64(tarifa.MinutosBase) {
			minutosFacturables = float64(tarifa.MinutosBase)
		}
	}
	bloquesExtra := 0
	if minutosFacturables > float64(tarifa.MinutosBase) && tarifa.MinutosExtra > 0 {
		extraMinutos := minutosFacturables - float64(tarifa.MinutosBase)
		bloquesExtra = int(math.Ceil(extraMinutos / float64(tarifa.MinutosExtra)))
		if bloquesExtra < 0 {
			bloquesExtra = 0
		}
	}

	montoExtra := round2(float64(bloquesExtra) * round2(tarifa.ValorExtra))
	subtotal := round2(montoBase + montoExtra)
	redondeado := applyTarifaPorMinutosRounding(subtotal, cfg)
	ajuste := round2(redondeado - subtotal)

	total := redondeado
	minApplied := false
	maxApplied := false
	minimo := round2(cfg.MontoMinimoDiario)
	maximo := round2(cfg.MontoMaximoDiario)
	if minimo > 0 && total < minimo {
		total = minimo
		minApplied = true
	}
	if maximo > 0 && total > maximo {
		total = maximo
		maxApplied = true
	}

	return EmpresaTarifaPorMinutosCalculo{
		TarifaID:            tarifa.ID,
		EstacionID:          tarifa.EstacionID,
		MinutosConsumidos:   round2(minutosConsumidos),
		MinutosFacturables:  round2(minutosFacturables),
		MinutosTolerancia:   margenTolerancia,
		BloquesExtra:        bloquesExtra,
		MontoBase:           montoBase,
		MontoExtra:          montoExtra,
		MontoSubtotal:       subtotal,
		MontoRedondeado:     redondeado,
		AjusteRedondeo:      ajuste,
		MontoMinimoAplicado: minApplied,
		MontoMaximoAplicado: maxApplied,
		MontoTotal:          round2(total),
		Moneda:              normalizeTarifaMoneda(tarifa.Moneda),
	}
}

// CalcularMontoTarifaPorMinutos calcula el valor total segun minutos consumidos.
func CalcularMontoTarifaPorMinutos(tarifa EmpresaTarifaPorMinutos, minutosConsumidos int) (float64, int) {
	detalle := CalcularDetalleTarifaPorMinutos(tarifa, float64(minutosConsumidos), defaultEmpresaTarifaPorMinutosConfiguracion(tarifa.EmpresaID))
	return detalle.MontoTotal, detalle.BloquesExtra
}

func findEmpresaTarifaPorMinutosByStationRange(dbConn *sql.DB, empresaID, estacionID int64, diaDesde, diaHasta int) (*EmpresaTarifaPorMinutos, error) {
	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		estacion_id,
		COALESCE(estacion_codigo, ''),
		COALESCE(estacion_nombre, ''),
		COALESCE(dia_semana_desde, 1),
		COALESCE(dia_semana_hasta, 7),
		COALESCE(minutos_base, 120),
		COALESCE(valor_base, 0),
		COALESCE(minutos_extra, 60),
		COALESCE(valor_extra, 0),
		COALESCE(cobrar_por_fraccion, 0),
		COALESCE(moneda, 'COP'),
		COALESCE(prioridad, 1),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_tarifas_por_minutos
	WHERE empresa_id = ?
		AND estacion_id = ?
		AND dia_semana_desde = ?
		AND dia_semana_hasta = ?
	ORDER BY id DESC
	LIMIT 1`, empresaID, estacionID, diaDesde, diaHasta)

	var item EmpresaTarifaPorMinutos
	var cobrarPorFraccion int
	if err := row.Scan(
		&item.ID,
		&item.EmpresaID,
		&item.EstacionID,
		&item.EstacionCodigo,
		&item.EstacionNombre,
		&item.DiaSemanaDesde,
		&item.DiaSemanaHasta,
		&item.MinutosBase,
		&item.ValorBase,
		&item.MinutosExtra,
		&item.ValorExtra,
		&cobrarPorFraccion,
		&item.Moneda,
		&item.Prioridad,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	item.CobrarPorFraccion = cobrarPorFraccion > 0
	return &item, nil
}

func mergeTarifaPorMinutosEstacionRef(store map[int64]empresaTarifaPorMinutosEstacionRef, ref empresaTarifaPorMinutosEstacionRef, empresaID int64) {
	if ref.ID <= 0 {
		return
	}
	if strings.TrimSpace(ref.Codigo) == "" {
		ref.Codigo = fmt.Sprintf("EST-%d-%d", empresaID, ref.ID)
	}
	if strings.TrimSpace(ref.Nombre) == "" {
		ref.Nombre = fmt.Sprintf("Estacion %d", ref.ID)
	}
	current, ok := store[ref.ID]
	if !ok {
		store[ref.ID] = ref
		return
	}
	if strings.TrimSpace(current.Codigo) == "" {
		current.Codigo = ref.Codigo
	}
	if strings.TrimSpace(current.Nombre) == "" {
		current.Nombre = ref.Nombre
	}
	store[ref.ID] = current
}

func listEmpresaTarifaPorMinutosStationRefs(dbConn *sql.DB, empresaID int64) ([]empresaTarifaPorMinutosEstacionRef, error) {
	refs := make(map[int64]empresaTarifaPorMinutosEstacionRef)

	if pref, err := GetEmpresaEstacionPref(dbConn, empresaID, 0, "estaciones_config"); err != nil {
		return nil, err
	} else if pref != nil && strings.TrimSpace(pref.Valor) != "" {
		cfg, err := parseEmpresaEstacionesConfig(pref.Valor)
		if err != nil {
			return nil, err
		}
		if cfg != nil {
			for _, station := range cfg.Estaciones {
				mergeTarifaPorMinutosEstacionRef(refs, empresaTarifaPorMinutosEstacionRef{
					ID:     station.ID,
					Codigo: fmt.Sprintf("EST-%d-%d", empresaID, station.ID),
					Nombre: station.Nombre,
				}, empresaID)
			}
		}
	}

	rowsTarifas, err := dbConn.Query(`SELECT DISTINCT
		COALESCE(estacion_id, 0),
		COALESCE(estacion_codigo, ''),
		COALESCE(estacion_nombre, '')
	FROM empresa_tarifas_por_minutos
	WHERE empresa_id = ?`, empresaID)
	if err != nil {
		return nil, err
	}
	for rowsTarifas.Next() {
		var ref empresaTarifaPorMinutosEstacionRef
		if err := rowsTarifas.Scan(&ref.ID, &ref.Codigo, &ref.Nombre); err != nil {
			_ = rowsTarifas.Close()
			return nil, err
		}
		mergeTarifaPorMinutosEstacionRef(refs, ref, empresaID)
	}
	if err := rowsTarifas.Err(); err != nil {
		_ = rowsTarifas.Close()
		return nil, err
	}
	_ = rowsTarifas.Close()

	hasCarritos, err := tableExists(dbConn, "carritos_compras")
	if err != nil {
		return nil, err
	}
	if hasCarritos {
		rowsCarritos, err := dbConn.Query(`SELECT
			COALESCE(referencia_externa, ''),
			COALESCE(codigo, ''),
			COALESCE(nombre, '')
		FROM carritos_compras
		WHERE empresa_id = ?
			AND COALESCE(estado, 'activo') = 'activo'`, empresaID)
		if err != nil {
			return nil, err
		}
		for rowsCarritos.Next() {
			var referenciaExterna, codigo, nombre string
			if err := rowsCarritos.Scan(&referenciaExterna, &codigo, &nombre); err != nil {
				_ = rowsCarritos.Close()
				return nil, err
			}
			estacionID := parseReservaHotelEstacionID(referenciaExterna, codigo, empresaID)
			if estacionID <= 0 {
				continue
			}
			mergeTarifaPorMinutosEstacionRef(refs, empresaTarifaPorMinutosEstacionRef{
				ID:     estacionID,
				Codigo: strings.TrimSpace(codigo),
				Nombre: strings.TrimSpace(nombre),
			}, empresaID)
		}
		if err := rowsCarritos.Err(); err != nil {
			_ = rowsCarritos.Close()
			return nil, err
		}
		_ = rowsCarritos.Close()
	}

	hasReservas, err := tableExists(dbConn, "reservas_hotel")
	if err != nil {
		return nil, err
	}
	if hasReservas {
		rowsReservas, err := dbConn.Query(`SELECT DISTINCT COALESCE(estacion_id, 0)
		FROM reservas_hotel
		WHERE empresa_id = ?
			AND COALESCE(estacion_id, 0) > 0`, empresaID)
		if err != nil {
			return nil, err
		}
		for rowsReservas.Next() {
			var estacionID int64
			if err := rowsReservas.Scan(&estacionID); err != nil {
				_ = rowsReservas.Close()
				return nil, err
			}
			mergeTarifaPorMinutosEstacionRef(refs, empresaTarifaPorMinutosEstacionRef{ID: estacionID}, empresaID)
		}
		if err := rowsReservas.Err(); err != nil {
			_ = rowsReservas.Close()
			return nil, err
		}
		_ = rowsReservas.Close()
	}

	out := make([]empresaTarifaPorMinutosEstacionRef, 0, len(refs))
	for _, ref := range refs {
		out = append(out, ref)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out, nil
}

// ApplyEmpresaTarifaPorMinutosToAllStations aplica una misma regla de tarifa a todas las estaciones detectadas de la empresa.
func ApplyEmpresaTarifaPorMinutosToAllStations(dbConn *sql.DB, template EmpresaTarifaPorMinutos) (*EmpresaTarifaPorMinutosAplicacionMasivaResultado, error) {
	if template.EmpresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaTarifasPorMinutosSchema(dbConn); err != nil {
		return nil, err
	}

	normalized := template
	if normalized.EstacionID <= 0 {
		normalized.EstacionID = 1
	}
	if err := normalizeEmpresaTarifaPayload(&normalized); err != nil {
		return nil, err
	}

	refs, err := listEmpresaTarifaPorMinutosStationRefs(dbConn, template.EmpresaID)
	if err != nil {
		return nil, err
	}
	if template.EstacionID > 0 {
		templateRef := empresaTarifaPorMinutosEstacionRef{ID: template.EstacionID, Codigo: template.EstacionCodigo, Nombre: template.EstacionNombre}
		m := make(map[int64]empresaTarifaPorMinutosEstacionRef, len(refs)+1)
		for _, ref := range refs {
			m[ref.ID] = ref
		}
		mergeTarifaPorMinutosEstacionRef(m, templateRef, template.EmpresaID)
		refs = refs[:0]
		for _, ref := range m {
			refs = append(refs, ref)
		}
		sort.Slice(refs, func(i, j int) bool {
			return refs[i].ID < refs[j].ID
		})
	}
	if len(refs) == 0 {
		return nil, fmt.Errorf("no se encontraron estaciones para aplicar la tarifa")
	}

	result := &EmpresaTarifaPorMinutosAplicacionMasivaResultado{
		EmpresaID:          template.EmpresaID,
		DiaSemanaDesde:     normalized.DiaSemanaDesde,
		DiaSemanaHasta:     normalized.DiaSemanaHasta,
		EstacionesObjetivo: len(refs),
		TarifaIDs:          make([]int64, 0, len(refs)),
	}

	for _, ref := range refs {
		payload := normalized
		payload.ID = 0
		payload.EstacionID = ref.ID
		payload.EmpresaID = template.EmpresaID
		if strings.TrimSpace(ref.Codigo) != "" {
			payload.EstacionCodigo = strings.TrimSpace(ref.Codigo)
		} else {
			payload.EstacionCodigo = fmt.Sprintf("EST-%d-%d", template.EmpresaID, ref.ID)
		}
		if strings.TrimSpace(ref.Nombre) != "" {
			payload.EstacionNombre = strings.TrimSpace(ref.Nombre)
		} else {
			payload.EstacionNombre = fmt.Sprintf("Estacion %d", ref.ID)
		}

		existing, err := findEmpresaTarifaPorMinutosByStationRange(dbConn, template.EmpresaID, ref.ID, normalized.DiaSemanaDesde, normalized.DiaSemanaHasta)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			payload.ID = existing.ID
			if err := UpdateEmpresaTarifaPorMinutos(dbConn, payload); err != nil {
				return nil, err
			}
			result.TarifasActualizadas++
			result.TarifaIDs = append(result.TarifaIDs, existing.ID)
			continue
		}

		id, err := CreateEmpresaTarifaPorMinutos(dbConn, payload)
		if err != nil {
			return nil, err
		}
		result.TarifasCreadas++
		result.TarifaIDs = append(result.TarifaIDs, id)
	}

	return result, nil
}

// RegisterTarifaPorMinutosCalculoContable registra trazabilidad contable del calculo de tarifa.
func RegisterTarifaPorMinutosCalculoContable(
	dbConn *sql.DB,
	empresaID int64,
	tarifa EmpresaTarifaPorMinutos,
	cfg EmpresaTarifaPorMinutosConfiguracion,
	diaSemana int,
	minutosConsumidos float64,
	detalle EmpresaTarifaPorMinutosCalculo,
	usuarioCreador string,
	requestID string,
) (int64, string, string, error) {
	if empresaID <= 0 {
		return 0, "", "", fmt.Errorf("empresa_id es obligatorio")
	}
	if tarifa.ID <= 0 {
		return 0, "", "", fmt.Errorf("tarifa_id es obligatorio")
	}
	if err := EnsureEmpresaEventosContablesSchema(dbConn); err != nil {
		return 0, "", "", err
	}

	if diaSemana <= 0 {
		diaSemana = DayOfWeekISO(time.Now())
	}
	periodoContable := time.Now().Format("2006-01")
	documentoCodigo := fmt.Sprintf("TPM-%d-%d", empresaID, time.Now().UnixNano())

	payloadMap := map[string]interface{}{
		"modulo":                "tarifas_por_minutos",
		"tipo_calculo":          "simulacion_operativa",
		"tarifa_id":             tarifa.ID,
		"estacion_id":           tarifa.EstacionID,
		"dia_semana":            diaSemana,
		"minutos_consumidos":    round2(minutosConsumidos),
		"bloques_extra":         detalle.BloquesExtra,
		"monto_base":            detalle.MontoBase,
		"monto_extra":           detalle.MontoExtra,
		"monto_subtotal":        detalle.MontoSubtotal,
		"monto_redondeado":      detalle.MontoRedondeado,
		"ajuste_redondeo":       detalle.AjusteRedondeo,
		"monto_minimo_aplicado": detalle.MontoMinimoAplicado,
		"monto_maximo_aplicado": detalle.MontoMaximoAplicado,
		"monto_total":           detalle.MontoTotal,
		"redondeo_modo":         normalizeTarifaPorMinutosRedondeoModo(cfg.RedondeoModo),
		"redondeo_unidad":       normalizeTarifaPorMinutosRedondeoUnidad(cfg.RedondeoUnidad),
		"monto_minimo_diario":   round2(cfg.MontoMinimoDiario),
		"monto_maximo_diario":   round2(cfg.MontoMaximoDiario),
		"documento_codigo":      documentoCodigo,
		"periodo_contable":      periodoContable,
		"request_id":            strings.TrimSpace(requestID),
	}
	payloadJSON, err := json.Marshal(payloadMap)
	if err != nil {
		return 0, "", "", err
	}

	usuario := strings.TrimSpace(usuarioCreador)
	if usuario == "" {
		usuario = "sistema"
	}
	eventoID, err := CreateEmpresaEventoContable(dbConn, EmpresaEventoContable{
		EmpresaID:       empresaID,
		Modulo:          "finanzas",
		Evento:          "tarifa_por_minutos_calculada",
		Entidad:         "tarifa_por_minutos",
		EntidadID:       tarifa.ID,
		DocumentoTipo:   "tarifa_por_minutos_calculo",
		DocumentoCodigo: documentoCodigo,
		PeriodoContable: periodoContable,
		MontoTotal:      detalle.MontoTotal,
		Moneda:          normalizeTarifaMoneda(tarifa.Moneda),
		PayloadJSON:     string(payloadJSON),
		Origen:          "tarifas_por_minutos",
		UsuarioCreador:  usuario,
		Observaciones:   "trazabilidad de calculo de tarifa por minutos",
	})
	if err != nil {
		return 0, "", "", err
	}

	if _, err := dbConn.Exec(`UPDATE empresa_eventos_contables
	SET
		procesado = 1,
		fecha_procesado = CURRENT_TIMESTAMP,
		fecha_ultimo_intento = CURRENT_TIMESTAMP,
		intentos_procesamiento = CASE WHEN COALESCE(intentos_procesamiento, 0) <= 0 THEN 1 ELSE intentos_procesamiento END,
		error_procesamiento = '',
		asiento_contable_id = COALESCE(asiento_contable_id, 0),
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE id = ?`, eventoID); err != nil {
		return 0, "", "", err
	}

	return eventoID, documentoCodigo, periodoContable, nil
}
