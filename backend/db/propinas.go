package db

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

const (
	EmpresaPropinaModoPorUsuario = "por_usuario"
	EmpresaPropinaModoUniversal  = "universal"
)

// EmpresaPropinasConfiguracion define el comportamiento de propinas por empresa.
type EmpresaPropinasConfiguracion struct {
	ID                     int64   `json:"id"`
	EmpresaID              int64   `json:"empresa_id"`
	HabilitarPropina       bool    `json:"habilitar_propina"`
	PorcentajePropina      float64 `json:"porcentaje_propina"`
	ModoDistribucion       string  `json:"modo_distribucion"`
	AplicarAutomaticamente bool    `json:"aplicar_automaticamente"`
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
	VentaReferencia    string  `json:"venta_referencia,omitempty"`
	UsuarioOrigen      string  `json:"usuario_origen,omitempty"`
	UsuarioAsignado    string  `json:"usuario_asignado,omitempty"`
	ModoDistribucion   string  `json:"modo_distribucion"`
	Moneda             string  `json:"moneda,omitempty"`
	BaseCobro          float64 `json:"base_cobro"`
	PorcentajePropina  float64 `json:"porcentaje_propina"`
	MontoPropina       float64 `json:"monto_propina"`
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
	IncludeInactive  bool
	Limit            int
}

// EmpresaPropinasResumen consolida metricas de propinas en un periodo.
type EmpresaPropinasResumen struct {
	TotalBaseCobro           float64 `json:"total_base_cobro"`
	TotalPropinas            float64 `json:"total_propinas"`
	TotalPropinasPorUsuario  float64 `json:"total_propinas_por_usuario"`
	TotalPropinasUniversal   float64 `json:"total_propinas_universal"`
	CantidadMovimientos      int64   `json:"cantidad_movimientos"`
	UsuariosActivos          int     `json:"usuarios_activos"`
	CuotaUniversalPorUsuario float64 `json:"cuota_universal_por_usuario"`
}

// EmpresaPropinaUsuarioResumen presenta acumulados por usuario.
type EmpresaPropinaUsuarioResumen struct {
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
	Clave    string
	Etiqueta string
}

// EnsureEmpresaPropinasSchema crea/migra las tablas de configuracion y movimientos de propinas.
func EnsureEmpresaPropinasSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_propinas_configuracion (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL UNIQUE,
			habilitar_propina INTEGER DEFAULT 0,
			porcentaje_propina REAL DEFAULT 10,
			modo_distribucion TEXT DEFAULT 'por_usuario',
			aplicar_automaticamente INTEGER DEFAULT 1,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_propinas_configuracion_empresa ON empresa_propinas_configuracion(empresa_id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_propinas_configuracion_estado ON empresa_propinas_configuracion(empresa_id, estado);`,
		`CREATE TABLE IF NOT EXISTS empresa_propinas_movimientos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			carrito_id INTEGER DEFAULT 0,
			venta_referencia TEXT,
			usuario_origen TEXT,
			usuario_asignado TEXT,
			modo_distribucion TEXT DEFAULT 'por_usuario',
			moneda TEXT DEFAULT 'COP',
			base_cobro REAL DEFAULT 0,
			porcentaje_propina REAL DEFAULT 0,
			monto_propina REAL DEFAULT 0,
			fecha_movimiento TEXT DEFAULT (datetime('now','localtime')),
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_propinas_movimientos_empresa_fecha ON empresa_propinas_movimientos(empresa_id, fecha_movimiento DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_propinas_movimientos_empresa_usuario ON empresa_propinas_movimientos(empresa_id, usuario_asignado, usuario_origen);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_propinas_movimientos_empresa_modo ON empresa_propinas_movimientos(empresa_id, modo_distribucion);`,
	}
	for _, stmt := range stmts {
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
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "venta_referencia", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "usuario_origen", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "usuario_asignado", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_propinas_movimientos", "modo_distribucion", "TEXT DEFAULT 'por_usuario'"); err != nil {
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

	return nil
}

func defaultEmpresaPropinasConfiguracion(empresaID int64) EmpresaPropinasConfiguracion {
	return EmpresaPropinasConfiguracion{
		EmpresaID:              empresaID,
		HabilitarPropina:       false,
		PorcentajePropina:      10,
		ModoDistribucion:       EmpresaPropinaModoPorUsuario,
		AplicarAutomaticamente: true,
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

// GetEmpresaPropinasConfiguracion obtiene la configuracion activa de propinas por empresa.
func GetEmpresaPropinasConfiguracion(dbConn *sql.DB, empresaID int64) (*EmpresaPropinasConfiguracion, error) {
	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(habilitar_propina, 0),
		COALESCE(porcentaje_propina, 10),
		COALESCE(modo_distribucion, 'por_usuario'),
		COALESCE(aplicar_automaticamente, 1),
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
			usuario_creador = ?,
			estado = ?,
			observaciones = ?,
			fecha_actualizacion = datetime('now','localtime')
		WHERE empresa_id = ?`,
			boolToInt(payload.HabilitarPropina),
			payload.PorcentajePropina,
			payload.ModoDistribucion,
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

	res, err := dbConn.Exec(`INSERT INTO empresa_propinas_configuracion (
		empresa_id,
		habilitar_propina,
		porcentaje_propina,
		modo_distribucion,
		aplicar_automaticamente,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
		payload.EmpresaID,
		boolToInt(payload.HabilitarPropina),
		payload.PorcentajePropina,
		payload.ModoDistribucion,
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

// CreateEmpresaPropinaMovimiento registra una propina asociada al cierre de una venta.
func CreateEmpresaPropinaMovimiento(dbConn *sql.DB, payload EmpresaPropinaMovimiento) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	payload.ModoDistribucion = normalizePropinaModo(payload.ModoDistribucion)
	payload.Moneda = normalizePropinaMoneda(payload.Moneda)
	payload.BaseCobro = round2(payload.BaseCobro)
	payload.PorcentajePropina = normalizePropinaPorcentaje(payload.PorcentajePropina)
	payload.MontoPropina = round2(payload.MontoPropina)
	if payload.MontoPropina <= 0 {
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
	if strings.TrimSpace(payload.Estado) == "" {
		payload.Estado = "activo"
	}

	res, err := dbConn.Exec(`INSERT INTO empresa_propinas_movimientos (
		empresa_id,
		carrito_id,
		venta_referencia,
		usuario_origen,
		usuario_asignado,
		modo_distribucion,
		moneda,
		base_cobro,
		porcentaje_propina,
		monto_propina,
		fecha_movimiento,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
		payload.EmpresaID,
		payload.CarritoID,
		strings.TrimSpace(payload.VentaReferencia),
		payload.UsuarioOrigen,
		payload.UsuarioAsignado,
		payload.ModoDistribucion,
		payload.Moneda,
		payload.BaseCobro,
		payload.PorcentajePropina,
		payload.MontoPropina,
		strings.TrimSpace(payload.UsuarioCreador),
		strings.TrimSpace(payload.Estado),
		strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListEmpresaPropinaMovimientos lista movimientos de propinas por empresa.
func ListEmpresaPropinaMovimientos(dbConn *sql.DB, empresaID int64, filter EmpresaPropinaMovimientoFilter) ([]EmpresaPropinaMovimiento, error) {
	query := `SELECT
		id,
		empresa_id,
		COALESCE(carrito_id, 0),
		COALESCE(venta_referencia, ''),
		COALESCE(usuario_origen, ''),
		COALESCE(usuario_asignado, ''),
		COALESCE(modo_distribucion, 'por_usuario'),
		COALESCE(moneda, 'COP'),
		COALESCE(base_cobro, 0),
		COALESCE(porcentaje_propina, 0),
		COALESCE(monto_propina, 0),
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
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.CarritoID,
			&item.VentaReferencia,
			&item.UsuarioOrigen,
			&item.UsuarioAsignado,
			&item.ModoDistribucion,
			&item.Moneda,
			&item.BaseCobro,
			&item.PorcentajePropina,
			&item.MontoPropina,
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
		item.Moneda = normalizePropinaMoneda(item.Moneda)
		item.BaseCobro = round2(item.BaseCobro)
		item.PorcentajePropina = normalizePropinaPorcentaje(item.PorcentajePropina)
		item.MontoPropina = round2(item.MontoPropina)
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
		&resumen.CantidadMovimientos,
		&resumen.TotalPropinasPorUsuario,
		&resumen.TotalPropinasUniversal,
	); err != nil {
		return nil, err
	}
	resumen.TotalBaseCobro = round2(resumen.TotalBaseCobro)
	resumen.TotalPropinas = round2(resumen.TotalPropinas)
	resumen.TotalPropinasPorUsuario = round2(resumen.TotalPropinasPorUsuario)
	resumen.TotalPropinasUniversal = round2(resumen.TotalPropinasUniversal)

	directUsersQuery := `SELECT
		COALESCE(NULLIF(TRIM(usuario_asignado), ''), NULLIF(TRIM(usuario_origen), ''), 'sistema') AS usuario,
		COALESCE(SUM(monto_propina), 0)
	FROM empresa_propinas_movimientos
	WHERE empresa_id = ? AND COALESCE(modo_distribucion, 'por_usuario') = 'por_usuario'`
	directArgs := []interface{}{empresaID}
	directUsersQuery, directArgs = appendPropinaCommonFilters(directUsersQuery, directArgs, filter)
	directUsersQuery += ` GROUP BY usuario ORDER BY COALESCE(SUM(monto_propina), 0) DESC, usuario ASC`

	perUser := map[string]*EmpresaPropinaUsuarioResumen{}
	rows, err := dbConn.Query(directUsersQuery, directArgs...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var usuario string
		var total float64
		if err := rows.Scan(&usuario, &total); err != nil {
			rows.Close()
			return nil, err
		}
		clave := normalizePropinaUsuarioClave(usuario, "")
		item := &EmpresaPropinaUsuarioResumen{
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
				UsuarioClave:    user.Clave,
				UsuarioEtiqueta: user.Etiqueta,
			}
			perUser[user.Clave] = entry
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
	if usuario := strings.TrimSpace(strings.ToLower(filter.Usuario)); usuario != "" {
		like := "%" + usuario + "%"
		query += ` AND (
			LOWER(COALESCE(usuario_asignado, '')) LIKE ?
			OR LOWER(COALESCE(usuario_origen, '')) LIKE ?
		)`
		args = append(args, like, like)
	}
	return query, args
}

func listActiveUsersForPropinas(dbConn *sql.DB, empresaID int64) ([]propinaActiveUser, error) {
	rows, err := dbConn.Query(`SELECT
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
		var email string
		var name string
		if err := rows.Scan(&email, &name); err != nil {
			return nil, err
		}
		key := normalizePropinaUsuarioClave(email, name)
		if key == "" {
			continue
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, propinaActiveUser{
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
