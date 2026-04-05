package db

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
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

// EmpresaComisionServicioMovimiento representa una comision calculada sobre un item de servicio.
type EmpresaComisionServicioMovimiento struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	CarritoID          int64   `json:"carrito_id,omitempty"`
	CarritoItemID      int64   `json:"carrito_item_id,omitempty"`
	ServicioID         int64   `json:"servicio_id,omitempty"`
	ServicioCodigo     string  `json:"servicio_codigo,omitempty"`
	ServicioNombre     string  `json:"servicio_nombre,omitempty"`
	ServicioCategoria  string  `json:"servicio_categoria,omitempty"`
	UsuarioOrigen      string  `json:"usuario_origen,omitempty"`
	UsuarioLavador     string  `json:"usuario_lavador,omitempty"`
	VentaReferencia    string  `json:"venta_referencia,omitempty"`
	Moneda             string  `json:"moneda,omitempty"`
	BaseServicio       float64 `json:"base_servicio"`
	PorcentajeComision float64 `json:"porcentaje_comision"`
	MontoComision      float64 `json:"monto_comision"`
	FechaMovimiento    string  `json:"fecha_movimiento,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// EmpresaComisionServicioMovimientoFilter filtra movimientos de comisiones por servicio.
type EmpresaComisionServicioMovimientoFilter struct {
	Desde           string
	Hasta           string
	UsuarioLavador  string
	ServicioFiltro  string
	IncludeInactive bool
	Limit           int
}

// EmpresaComisionesServicioResumen consolida metricas de comisiones por servicio.
type EmpresaComisionesServicioResumen struct {
	TotalBaseServicios   float64 `json:"total_base_servicios"`
	TotalComisiones      float64 `json:"total_comisiones"`
	CantidadMovimientos  int64   `json:"cantidad_movimientos"`
	LavadoresConComision int64   `json:"lavadores_con_comision"`
}

// EmpresaComisionServicioLavadorResumen presenta acumulado por lavador.
type EmpresaComisionServicioLavadorResumen struct {
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
	BaseServicios          float64 `json:"base_servicios"`
	MontoComision          float64 `json:"monto_comision"`
	MovimientosRegistrados int     `json:"movimientos_registrados"`
	RegistroIDs            []int64 `json:"registro_ids,omitempty"`
	Warning                string  `json:"warning,omitempty"`
}

// EnsureEmpresaComisionesServicioSchema crea/migra tablas de comisiones por servicio.
func EnsureEmpresaComisionesServicioSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_comisiones_servicio_configuracion (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL UNIQUE,
			habilitar_comisiones INTEGER DEFAULT 0,
			porcentaje_comision REAL DEFAULT 10,
			filtro_servicio TEXT DEFAULT 'lavado',
			aplicar_automaticamente INTEGER DEFAULT 1,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_comisiones_servicio_cfg_empresa ON empresa_comisiones_servicio_configuracion(empresa_id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_comisiones_servicio_cfg_estado ON empresa_comisiones_servicio_configuracion(empresa_id, estado);`,
		`CREATE TABLE IF NOT EXISTS empresa_comisiones_servicio_movimientos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			carrito_id INTEGER DEFAULT 0,
			carrito_item_id INTEGER DEFAULT 0,
			servicio_id INTEGER DEFAULT 0,
			servicio_codigo TEXT,
			servicio_nombre TEXT,
			servicio_categoria TEXT,
			usuario_origen TEXT,
			usuario_lavador TEXT,
			venta_referencia TEXT,
			moneda TEXT DEFAULT 'COP',
			base_servicio REAL DEFAULT 0,
			porcentaje_comision REAL DEFAULT 0,
			monto_comision REAL DEFAULT 0,
			fecha_movimiento TEXT DEFAULT (datetime('now','localtime')),
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_comisiones_servicio_mov_empresa_fecha ON empresa_comisiones_servicio_movimientos(empresa_id, fecha_movimiento DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_comisiones_servicio_mov_empresa_lavador ON empresa_comisiones_servicio_movimientos(empresa_id, usuario_lavador);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_comisiones_servicio_mov_item_lavador ON empresa_comisiones_servicio_movimientos(empresa_id, carrito_item_id, usuario_lavador);`,
	}
	for _, stmt := range stmts {
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
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "usuario_lavador", "TEXT"); err != nil {
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
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "monto_comision", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_comisiones_servicio_movimientos", "fecha_movimiento", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
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
			fecha_actualizacion = datetime('now','localtime')
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

	res, err := dbConn.Exec(`INSERT INTO empresa_comisiones_servicio_configuracion (
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
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
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
	return res.LastInsertId()
}

// CreateEmpresaComisionServicioMovimiento registra un movimiento de comision por servicio.
func CreateEmpresaComisionServicioMovimiento(dbConn *sql.DB, payload EmpresaComisionServicioMovimiento) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	payload.UsuarioOrigen = strings.TrimSpace(payload.UsuarioOrigen)
	payload.UsuarioLavador = defaultComisionLavador(payload.UsuarioLavador, payload.UsuarioOrigen)
	payload.Moneda = normalizeComisionMoneda(payload.Moneda)
	payload.BaseServicio = round2(payload.BaseServicio)
	payload.PorcentajeComision = normalizeComisionPorcentaje(payload.PorcentajeComision)
	payload.MontoComision = round2(payload.MontoComision)
	payload.ServicioCodigo = strings.TrimSpace(payload.ServicioCodigo)
	payload.ServicioNombre = strings.TrimSpace(payload.ServicioNombre)
	payload.ServicioCategoria = strings.TrimSpace(payload.ServicioCategoria)
	if payload.MontoComision <= 0 {
		return 0, fmt.Errorf("monto_comision debe ser mayor a cero")
	}
	if strings.TrimSpace(payload.Estado) == "" {
		payload.Estado = "activo"
	}

	res, err := dbConn.Exec(`INSERT INTO empresa_comisiones_servicio_movimientos (
		empresa_id,
		carrito_id,
		carrito_item_id,
		servicio_id,
		servicio_codigo,
		servicio_nombre,
		servicio_categoria,
		usuario_origen,
		usuario_lavador,
		venta_referencia,
		moneda,
		base_servicio,
		porcentaje_comision,
		monto_comision,
		fecha_movimiento,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
		payload.EmpresaID,
		payload.CarritoID,
		payload.CarritoItemID,
		payload.ServicioID,
		payload.ServicioCodigo,
		payload.ServicioNombre,
		payload.ServicioCategoria,
		payload.UsuarioOrigen,
		payload.UsuarioLavador,
		strings.TrimSpace(payload.VentaReferencia),
		payload.Moneda,
		payload.BaseServicio,
		payload.PorcentajeComision,
		payload.MontoComision,
		strings.TrimSpace(payload.UsuarioCreador),
		strings.TrimSpace(payload.Estado),
		strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func buildEmpresaComisionServicioMovWhere(empresaID int64, filter EmpresaComisionServicioMovimientoFilter) (string, []interface{}) {
	clauses := []string{"empresa_id = ?"}
	args := []interface{}{empresaID}

	if !filter.IncludeInactive {
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
		clauses = append(clauses, "(lower(COALESCE(usuario_lavador, '')) LIKE ? OR lower(COALESCE(usuario_origen, '')) LIKE ?)")
		args = append(args, like, like)
	}
	if servicioFiltro := strings.TrimSpace(filter.ServicioFiltro); servicioFiltro != "" {
		like := "%" + strings.ToLower(servicioFiltro) + "%"
		clauses = append(clauses, "(lower(COALESCE(servicio_nombre, '')) LIKE ? OR lower(COALESCE(servicio_categoria, '')) LIKE ? OR lower(COALESCE(servicio_codigo, '')) LIKE ?)")
		args = append(args, like, like, like)
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
		COALESCE(usuario_lavador, ''),
		COALESCE(venta_referencia, ''),
		COALESCE(moneda, 'COP'),
		COALESCE(base_servicio, 0),
		COALESCE(porcentaje_comision, 0),
		COALESCE(monto_comision, 0),
		COALESCE(fecha_movimiento, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_comisiones_servicio_movimientos
	WHERE %s
	ORDER BY datetime(COALESCE(fecha_movimiento, fecha_creacion)) DESC, id DESC
	LIMIT %d`, whereSQL, limit)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]EmpresaComisionServicioMovimiento, 0)
	for rows.Next() {
		var row EmpresaComisionServicioMovimiento
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
			&row.UsuarioLavador,
			&row.VentaReferencia,
			&row.Moneda,
			&row.BaseServicio,
			&row.PorcentajeComision,
			&row.MontoComision,
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
		row.MontoComision = round2(row.MontoComision)
		row.UsuarioLavador = defaultComisionLavador(row.UsuarioLavador, row.UsuarioOrigen)
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
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

		key := defaultComisionLavador(mov.UsuarioLavador, mov.UsuarioOrigen)
		entry := byLavador[key]
		if entry == nil {
			entry = &EmpresaComisionServicioLavadorResumen{UsuarioLavador: key}
			byLavador[key] = entry
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
		Resumen:       resumen,
		Lavadores:     lavadores,
		Movimientos:   movs,
	}, nil
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

// RegisterEmpresaComisionesServicioDesdeCarrito calcula y registra comisiones por servicios al cerrar carrito.
func RegisterEmpresaComisionesServicioDesdeCarrito(dbConn *sql.DB, empresaID, carritoID int64, usuarioLavador, usuarioOrigen string) (*EmpresaComisionServicioRegistroResultado, error) {
	if empresaID <= 0 || carritoID <= 0 {
		return nil, fmt.Errorf("empresa_id y carrito_id son obligatorios")
	}

	cfg, err := GetEmpresaComisionesServicioConfiguracion(dbConn, empresaID)
	if err != nil {
		return nil, err
	}

	result := &EmpresaComisionServicioRegistroResultado{
		Habilitada:           cfg.HabilitarComisiones,
		AplicacionAutomatica: cfg.AplicarAutomaticamente,
		PorcentajeComision:   cfg.PorcentajeComision,
		FiltroServicio:       cfg.FiltroServicio,
		UsuarioLavador:       defaultComisionLavador(usuarioLavador, usuarioOrigen),
	}

	if !cfg.HabilitarComisiones {
		result.Warning = "configuracion de comisiones deshabilitada"
		return result, nil
	}
	if !cfg.AplicarAutomaticamente {
		result.Warning = "comisiones configuradas en modo manual"
		return result, nil
	}
	if cfg.PorcentajeComision <= 0 {
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

	for _, item := range items {
		if !servicioCumpleFiltroComision(item, cfg.FiltroServicio) {
			continue
		}
		baseServicio := round2(item.TotalLinea)
		if baseServicio <= 0 {
			continue
		}
		montoComision := round2(baseServicio * (cfg.PorcentajeComision / 100))
		if montoComision <= 0 {
			continue
		}

		id, err := CreateEmpresaComisionServicioMovimiento(dbConn, EmpresaComisionServicioMovimiento{
			EmpresaID:          empresaID,
			CarritoID:          carritoID,
			CarritoItemID:      item.CarritoItemID,
			ServicioID:         item.ServicioID,
			ServicioCodigo:     firstNonEmpty(item.ServicioCodigo, item.CodigoItem),
			ServicioNombre:     firstNonEmpty(item.ServicioNombre, item.Descripcion),
			ServicioCategoria:  item.ServicioCategoria,
			UsuarioOrigen:      usuarioOrigen,
			UsuarioLavador:     result.UsuarioLavador,
			VentaReferencia:    ventaReferencia,
			Moneda:             moneda,
			BaseServicio:       baseServicio,
			PorcentajeComision: cfg.PorcentajeComision,
			MontoComision:      montoComision,
			UsuarioCreador:     usuarioOrigen,
			Estado:             "activo",
			Observaciones:      "comision registrada al cerrar carrito en estacion",
		})
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "unique") {
				continue
			}
			return nil, err
		}

		result.RegistroIDs = append(result.RegistroIDs, id)
		result.MovimientosRegistrados++
		result.BaseServicios = round2(result.BaseServicios + baseServicio)
		result.MontoComision = round2(result.MontoComision + montoComision)
	}

	if result.MovimientosRegistrados == 0 {
		result.Warning = "sin servicios coincidentes con el filtro de comision"
		return result, nil
	}
	result.Aplicada = true
	return result, nil
}
