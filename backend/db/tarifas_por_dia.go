package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// EmpresaTarifaPorDia define la regla de cobro diario por estacion.
type EmpresaTarifaPorDia struct {
	ID                     int64   `json:"id"`
	EmpresaID              int64   `json:"empresa_id"`
	EstacionID             int64   `json:"estacion_id"`
	EstacionCodigo         string  `json:"estacion_codigo,omitempty"`
	EstacionNombre         string  `json:"estacion_nombre,omitempty"`
	ServicioNombre         string  `json:"servicio_nombre,omitempty"`
	ValorDia               float64 `json:"valor_dia"`
	HoraCheckIn            string  `json:"hora_check_in"`
	HoraCheckOut           string  `json:"hora_check_out"`
	Moneda                 string  `json:"moneda,omitempty"`
	Prioridad              int     `json:"prioridad"`
	AplicarAutomaticamente bool    `json:"aplicar_automaticamente"`
	FechaCreacion          string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion     string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador         string  `json:"usuario_creador,omitempty"`
	Estado                 string  `json:"estado,omitempty"`
	Observaciones          string  `json:"observaciones,omitempty"`
}

// EmpresaTarifaPorDiaFilter define filtros para listar tarifas diarias.
type EmpresaTarifaPorDiaFilter struct {
	EstacionID      int64
	IncludeInactive bool
	Limit           int
}

// EmpresaTarifaPorDiaCalculo representa un calculo puntual de tarifa diaria.
type EmpresaTarifaPorDiaCalculo struct {
	TarifaID     int64   `json:"tarifa_id"`
	EstacionID   int64   `json:"estacion_id"`
	DiasCobrados int     `json:"dias_cobrados"`
	ValorDia     float64 `json:"valor_dia"`
	MontoTotal   float64 `json:"monto_total"`
	Moneda       string  `json:"moneda"`
	HoraCheckIn  string  `json:"hora_check_in"`
	HoraCheckOut string  `json:"hora_check_out"`
	FechaInicio  string  `json:"fecha_inicio"`
	FechaCorte   string  `json:"fecha_corte"`
}

// EnsureEmpresaTarifasPorDiaSchema crea/migra tabla de tarifas diarias por estacion.
func EnsureEmpresaTarifasPorDiaSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_tarifas_por_dia (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			estacion_id INTEGER NOT NULL,
			estacion_codigo TEXT,
			estacion_nombre TEXT,
			servicio_nombre TEXT DEFAULT 'hospedaje',
			valor_dia REAL NOT NULL DEFAULT 0,
			hora_check_in TEXT DEFAULT '15:00',
			hora_check_out TEXT DEFAULT '12:00',
			moneda TEXT DEFAULT 'COP',
			prioridad INTEGER DEFAULT 1,
			aplicar_automaticamente INTEGER NOT NULL DEFAULT 1,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_tarifas_por_dia_estacion ON empresa_tarifas_por_dia(empresa_id, estacion_id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_tarifas_por_dia_empresa_estado ON empresa_tarifas_por_dia(empresa_id, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_tarifas_por_dia_empresa_estacion ON empresa_tarifas_por_dia(empresa_id, estacion_id);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_dia", "estacion_codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_dia", "estacion_nombre", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_dia", "servicio_nombre", "TEXT DEFAULT 'hospedaje'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_dia", "valor_dia", "REAL NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_dia", "hora_check_in", "TEXT DEFAULT '15:00'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_dia", "hora_check_out", "TEXT DEFAULT '12:00'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_dia", "moneda", "TEXT DEFAULT 'COP'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_dia", "prioridad", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_dia", "aplicar_automaticamente", "INTEGER NOT NULL DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_dia", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_dia", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_dia", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_dia", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

func normalizeTarifaPorDiaEstado(estado string) string {
	if strings.EqualFold(strings.TrimSpace(estado), "inactivo") {
		return "inactivo"
	}
	return "activo"
}

func normalizeTarifaPorDiaMoneda(moneda string) string {
	trim := strings.ToUpper(strings.TrimSpace(moneda))
	if trim == "" {
		return "COP"
	}
	return trim
}

func normalizeTarifaPorDiaServicio(servicio string) string {
	trim := strings.TrimSpace(servicio)
	if trim == "" {
		return "hospedaje"
	}
	return trim
}

func normalizeTarifaPorDiaPrioridad(v int) int {
	if v <= 0 {
		return 1
	}
	if v > 999 {
		return 999
	}
	return v
}

func parseTarifaPorDiaHora(raw string) (int, int, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, 0, fmt.Errorf("hora vacia")
	}
	layouts := []string{"15:04", "15:04:05", "15"}
	for _, layout := range layouts {
		ts, err := time.Parse(layout, value)
		if err == nil {
			return ts.Hour(), ts.Minute(), nil
		}
	}
	return 0, 0, fmt.Errorf("hora invalida")
}

func normalizeTarifaPorDiaHora(raw, fallback string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = strings.TrimSpace(fallback)
	}
	hour, minute, err := parseTarifaPorDiaHora(value)
	if err != nil {
		return "", fmt.Errorf("hora invalida")
	}
	return fmt.Sprintf("%02d:%02d", hour, minute), nil
}

func parseTarifaPorDiaDateTime(raw string) (time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, fmt.Errorf("fecha vacia")
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2006-01-02",
	}
	for _, layout := range layouts {
		ts, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			return ts, nil
		}
	}
	return time.Time{}, fmt.Errorf("fecha invalida")
}

func normalizeEmpresaTarifaPorDiaPayload(payload *EmpresaTarifaPorDia) error {
	if payload == nil {
		return fmt.Errorf("payload invalido")
	}
	if payload.EmpresaID <= 0 {
		return fmt.Errorf("empresa_id es obligatorio")
	}
	if payload.EstacionID <= 0 {
		return fmt.Errorf("estacion_id es obligatorio")
	}
	if payload.ValorDia <= 0 {
		return fmt.Errorf("valor_dia debe ser mayor a cero")
	}

	horaCheckIn, err := normalizeTarifaPorDiaHora(payload.HoraCheckIn, "15:00")
	if err != nil {
		return fmt.Errorf("hora_check_in invalida")
	}
	horaCheckOut, err := normalizeTarifaPorDiaHora(payload.HoraCheckOut, "12:00")
	if err != nil {
		return fmt.Errorf("hora_check_out invalida")
	}

	payload.EstacionCodigo = strings.TrimSpace(payload.EstacionCodigo)
	payload.EstacionNombre = strings.TrimSpace(payload.EstacionNombre)
	payload.ServicioNombre = normalizeTarifaPorDiaServicio(payload.ServicioNombre)
	payload.ValorDia = round2(payload.ValorDia)
	payload.HoraCheckIn = horaCheckIn
	payload.HoraCheckOut = horaCheckOut
	payload.Moneda = normalizeTarifaPorDiaMoneda(payload.Moneda)
	payload.Prioridad = normalizeTarifaPorDiaPrioridad(payload.Prioridad)
	payload.Estado = normalizeTarifaPorDiaEstado(payload.Estado)
	payload.UsuarioCreador = strings.TrimSpace(payload.UsuarioCreador)
	payload.Observaciones = strings.TrimSpace(payload.Observaciones)
	return nil
}

func scanEmpresaTarifaPorDia(scanner interface {
	Scan(dest ...interface{}) error
}) (*EmpresaTarifaPorDia, error) {
	item := &EmpresaTarifaPorDia{}
	var aplicarAuto int64
	if err := scanner.Scan(
		&item.ID,
		&item.EmpresaID,
		&item.EstacionID,
		&item.EstacionCodigo,
		&item.EstacionNombre,
		&item.ServicioNombre,
		&item.ValorDia,
		&item.HoraCheckIn,
		&item.HoraCheckOut,
		&item.Moneda,
		&item.Prioridad,
		&aplicarAuto,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	); err != nil {
		return nil, err
	}
	item.AplicarAutomaticamente = aplicarAuto != 0
	return item, nil
}

// CreateEmpresaTarifaPorDia crea una tarifa diaria por estacion.
func CreateEmpresaTarifaPorDia(dbConn *sql.DB, payload EmpresaTarifaPorDia) (int64, error) {
	if err := normalizeEmpresaTarifaPorDiaPayload(&payload); err != nil {
		return 0, err
	}

	aplicarAuto := 0
	if payload.AplicarAutomaticamente {
		aplicarAuto = 1
	}

	res, err := dbConn.Exec(`INSERT INTO empresa_tarifas_por_dia (
		empresa_id,
		estacion_id,
		estacion_codigo,
		estacion_nombre,
		servicio_nombre,
		valor_dia,
		hora_check_in,
		hora_check_out,
		moneda,
		prioridad,
		aplicar_automaticamente,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
		payload.EmpresaID,
		payload.EstacionID,
		payload.EstacionCodigo,
		payload.EstacionNombre,
		payload.ServicioNombre,
		payload.ValorDia,
		payload.HoraCheckIn,
		payload.HoraCheckOut,
		payload.Moneda,
		payload.Prioridad,
		aplicarAuto,
		payload.UsuarioCreador,
		payload.Estado,
		payload.Observaciones,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateEmpresaTarifaPorDia actualiza una tarifa diaria existente.
func UpdateEmpresaTarifaPorDia(dbConn *sql.DB, payload EmpresaTarifaPorDia) error {
	if payload.ID <= 0 {
		return fmt.Errorf("id es obligatorio")
	}
	if err := normalizeEmpresaTarifaPorDiaPayload(&payload); err != nil {
		return err
	}

	aplicarAuto := 0
	if payload.AplicarAutomaticamente {
		aplicarAuto = 1
	}

	res, err := dbConn.Exec(`UPDATE empresa_tarifas_por_dia
	SET
		estacion_id = ?,
		estacion_codigo = ?,
		estacion_nombre = ?,
		servicio_nombre = ?,
		valor_dia = ?,
		hora_check_in = ?,
		hora_check_out = ?,
		moneda = ?,
		prioridad = ?,
		aplicar_automaticamente = ?,
		usuario_creador = ?,
		estado = ?,
		observaciones = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`,
		payload.EstacionID,
		payload.EstacionCodigo,
		payload.EstacionNombre,
		payload.ServicioNombre,
		payload.ValorDia,
		payload.HoraCheckIn,
		payload.HoraCheckOut,
		payload.Moneda,
		payload.Prioridad,
		aplicarAuto,
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

// SetEmpresaTarifaPorDiaEstado activa o desactiva una tarifa diaria.
func SetEmpresaTarifaPorDiaEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	if empresaID <= 0 || id <= 0 {
		return fmt.Errorf("empresa_id e id son obligatorios")
	}
	nextEstado := normalizeTarifaPorDiaEstado(estado)
	res, err := dbConn.Exec(`UPDATE empresa_tarifas_por_dia
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

// DeleteEmpresaTarifaPorDia elimina una tarifa diaria.
func DeleteEmpresaTarifaPorDia(dbConn *sql.DB, empresaID, id int64) error {
	if empresaID <= 0 || id <= 0 {
		return fmt.Errorf("empresa_id e id son obligatorios")
	}
	res, err := dbConn.Exec(`DELETE FROM empresa_tarifas_por_dia WHERE empresa_id = ? AND id = ?`, empresaID, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetEmpresaTarifaPorDiaByID obtiene una tarifa diaria por id y empresa.
func GetEmpresaTarifaPorDiaByID(dbConn *sql.DB, empresaID, id int64) (*EmpresaTarifaPorDia, error) {
	if empresaID <= 0 || id <= 0 {
		return nil, fmt.Errorf("empresa_id e id son obligatorios")
	}
	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		estacion_id,
		COALESCE(estacion_codigo, ''),
		COALESCE(estacion_nombre, ''),
		COALESCE(servicio_nombre, 'hospedaje'),
		COALESCE(valor_dia, 0),
		COALESCE(hora_check_in, '15:00'),
		COALESCE(hora_check_out, '12:00'),
		COALESCE(moneda, 'COP'),
		COALESCE(prioridad, 1),
		COALESCE(aplicar_automaticamente, 1),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_tarifas_por_dia
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, id)

	return scanEmpresaTarifaPorDia(row)
}

// ListEmpresaTarifasPorDia lista tarifas diarias por empresa.
func ListEmpresaTarifasPorDia(dbConn *sql.DB, empresaID int64, filter EmpresaTarifaPorDiaFilter) ([]EmpresaTarifaPorDia, error) {
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
		COALESCE(servicio_nombre, 'hospedaje'),
		COALESCE(valor_dia, 0),
		COALESCE(hora_check_in, '15:00'),
		COALESCE(hora_check_out, '12:00'),
		COALESCE(moneda, 'COP'),
		COALESCE(prioridad, 1),
		COALESCE(aplicar_automaticamente, 1),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_tarifas_por_dia
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	if !filter.IncludeInactive {
		query += ` AND COALESCE(estado, 'activo') = 'activo'`
	}
	if filter.EstacionID > 0 {
		query += ` AND estacion_id = ?`
		args = append(args, filter.EstacionID)
	}

	query += ` ORDER BY estacion_id ASC, prioridad ASC, id ASC LIMIT ?`
	args = append(args, filter.Limit)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaTarifaPorDia, 0)
	for rows.Next() {
		item, err := scanEmpresaTarifaPorDia(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func queryEmpresaTarifaPorDiaEstacion(dbConn *sql.DB, empresaID, estacionID int64, requireAutomatic bool) (*EmpresaTarifaPorDia, error) {
	if empresaID <= 0 || estacionID <= 0 {
		return nil, fmt.Errorf("empresa_id y estacion_id son obligatorios")
	}

	query := `SELECT
		id,
		empresa_id,
		estacion_id,
		COALESCE(estacion_codigo, ''),
		COALESCE(estacion_nombre, ''),
		COALESCE(servicio_nombre, 'hospedaje'),
		COALESCE(valor_dia, 0),
		COALESCE(hora_check_in, '15:00'),
		COALESCE(hora_check_out, '12:00'),
		COALESCE(moneda, 'COP'),
		COALESCE(prioridad, 1),
		COALESCE(aplicar_automaticamente, 1),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_tarifas_por_dia
	WHERE empresa_id = ?
		AND estacion_id = ?
		AND COALESCE(estado, 'activo') = 'activo'`
	args := []interface{}{empresaID, estacionID}
	if requireAutomatic {
		query += ` AND COALESCE(aplicar_automaticamente, 1) = 1`
	}
	query += ` ORDER BY prioridad ASC, id ASC LIMIT 1`

	row := dbConn.QueryRow(query, args...)
	item, err := scanEmpresaTarifaPorDia(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

// GetEmpresaTarifaPorDiaActiva devuelve la tarifa activa de una estacion.
func GetEmpresaTarifaPorDiaActiva(dbConn *sql.DB, empresaID, estacionID int64) (*EmpresaTarifaPorDia, error) {
	return queryEmpresaTarifaPorDiaEstacion(dbConn, empresaID, estacionID, false)
}

// GetEmpresaTarifaPorDiaAplicable devuelve la tarifa activa y automatica de una estacion.
func GetEmpresaTarifaPorDiaAplicable(dbConn *sql.DB, empresaID, estacionID int64) (*EmpresaTarifaPorDia, error) {
	return queryEmpresaTarifaPorDiaEstacion(dbConn, empresaID, estacionID, true)
}

func resolveTarifaPorDiaNextCheckoutBoundary(fechaInicio time.Time, horaCheckIn, horaCheckOut string) time.Time {
	location := fechaInicio.Location()
	if location == nil {
		location = time.Local
	}

	checkInHour, checkInMinute, errIn := parseTarifaPorDiaHora(horaCheckIn)
	if errIn != nil {
		checkInHour, checkInMinute = 15, 0
	}
	checkOutHour, checkOutMinute, errOut := parseTarifaPorDiaHora(horaCheckOut)
	if errOut != nil {
		checkOutHour, checkOutMinute = 12, 0
	}

	checkInMinutes := (checkInHour * 60) + checkInMinute
	checkOutMinutes := (checkOutHour * 60) + checkOutMinute
	startMinutes := (fechaInicio.Hour() * 60) + fechaInicio.Minute()

	baseDate := time.Date(fechaInicio.Year(), fechaInicio.Month(), fechaInicio.Day(), 0, 0, 0, 0, location)
	checkoutToday := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), checkOutHour, checkOutMinute, 0, 0, location)

	if checkInMinutes > checkOutMinutes {
		if startMinutes < checkOutMinutes && !fechaInicio.After(checkoutToday) {
			return checkoutToday
		}
		return checkoutToday.Add(24 * time.Hour)
	}

	if startMinutes < checkInMinutes {
		if !fechaInicio.After(checkoutToday) {
			return checkoutToday
		}
		return checkoutToday.Add(24 * time.Hour)
	}

	if startMinutes < checkOutMinutes && !fechaInicio.After(checkoutToday) {
		return checkoutToday
	}

	return checkoutToday.Add(24 * time.Hour)
}

// CalcularMontoTarifaPorDia calcula dias cobrados y monto total de la tarifa diaria.
func CalcularMontoTarifaPorDia(tarifa EmpresaTarifaPorDia, fechaInicio, fechaCorte time.Time) (int, float64) {
	if fechaInicio.IsZero() {
		return 0, 0
	}
	if fechaCorte.IsZero() || fechaCorte.Before(fechaInicio) {
		fechaCorte = fechaInicio
	}

	valorDia := round2(tarifa.ValorDia)
	if valorDia < 0 {
		valorDia = 0
	}

	dias := 1
	boundary := resolveTarifaPorDiaNextCheckoutBoundary(fechaInicio, tarifa.HoraCheckIn, tarifa.HoraCheckOut)
	for fechaCorte.After(boundary) {
		dias++
		if dias > 100000 {
			break
		}
		boundary = boundary.Add(24 * time.Hour)
	}

	monto := round2(float64(dias) * valorDia)
	return dias, monto
}

// CalcularDetalleTarifaPorDia construye el detalle completo del calculo diario.
func CalcularDetalleTarifaPorDia(tarifa EmpresaTarifaPorDia, fechaInicio, fechaCorte time.Time) EmpresaTarifaPorDiaCalculo {
	dias, monto := CalcularMontoTarifaPorDia(tarifa, fechaInicio, fechaCorte)
	if fechaCorte.IsZero() {
		fechaCorte = fechaInicio
	}
	return EmpresaTarifaPorDiaCalculo{
		TarifaID:     tarifa.ID,
		EstacionID:   tarifa.EstacionID,
		DiasCobrados: dias,
		ValorDia:     round2(tarifa.ValorDia),
		MontoTotal:   monto,
		Moneda:       normalizeTarifaPorDiaMoneda(tarifa.Moneda),
		HoraCheckIn:  tarifa.HoraCheckIn,
		HoraCheckOut: tarifa.HoraCheckOut,
		FechaInicio:  fechaInicio.Format("2006-01-02 15:04:05"),
		FechaCorte:   fechaCorte.Format("2006-01-02 15:04:05"),
	}
}
