package db

import (
	"database/sql"
	"fmt"
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

// EnsureEmpresaTarifasPorMinutosSchema crea/migra tablas de tarifas por minutos por estacion.
func EnsureEmpresaTarifasPorMinutosSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_tarifas_por_minutos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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
			moneda TEXT DEFAULT 'COP',
			prioridad INTEGER DEFAULT 1,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
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

	return nil
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

	res, err := dbConn.Exec(`INSERT INTO empresa_tarifas_por_minutos (
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
		moneda,
		prioridad,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
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
		payload.Moneda,
		payload.Prioridad,
		payload.UsuarioCreador,
		payload.Estado,
		payload.Observaciones,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
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
		moneda = ?,
		prioridad = ?,
		usuario_creador = ?,
		estado = ?,
		observaciones = ?,
		fecha_actualizacion = datetime('now','localtime')
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
	SET estado = ?, fecha_actualizacion = datetime('now','localtime')
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
	return &item, nil
}

// CalcularMontoTarifaPorMinutos calcula el valor total segun minutos consumidos.
func CalcularMontoTarifaPorMinutos(tarifa EmpresaTarifaPorMinutos, minutosConsumidos int) (float64, int) {
	if minutosConsumidos <= 0 {
		minutosConsumidos = tarifa.MinutosBase
	}
	total := round2(tarifa.ValorBase)
	if minutosConsumidos <= tarifa.MinutosBase {
		return total, 0
	}
	extraMinutos := minutosConsumidos - tarifa.MinutosBase
	bloques := extraMinutos / tarifa.MinutosExtra
	if extraMinutos%tarifa.MinutosExtra != 0 {
		bloques += 1
	}
	total += round2(float64(bloques) * tarifa.ValorExtra)
	return round2(total), bloques
}
