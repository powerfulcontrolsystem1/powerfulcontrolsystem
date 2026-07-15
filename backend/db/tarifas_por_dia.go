package db

import (
	"database/sql"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// EmpresaTarifaPorDia define la regla de cobro diario por estacion.
type EmpresaTarifaPorDia struct {
	ID                     int64   `json:"id"`
	EmpresaID              int64   `json:"empresa_id"`
	NombreTarifa           string  `json:"nombre_tarifa,omitempty"`
	EstacionID             int64   `json:"estacion_id"`
	EstacionCodigo         string  `json:"estacion_codigo,omitempty"`
	EstacionNombre         string  `json:"estacion_nombre,omitempty"`
	ServicioNombre         string  `json:"servicio_nombre,omitempty"`
	ValorDia               float64 `json:"valor_dia"`
	PersonasDesde          int     `json:"personas_desde"`
	PersonasHasta          int     `json:"personas_hasta"`
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
	Personas        int
	IncludeInactive bool
	Limit           int
}

// EmpresaTarifaPorDiaCalculo representa un calculo puntual de tarifa diaria.
type EmpresaTarifaPorDiaCalculo struct {
	TarifaID                    int64   `json:"tarifa_id"`
	EstacionID                  int64   `json:"estacion_id"`
	DiasCobrados                int     `json:"dias_cobrados"`
	DiasCompletos               int     `json:"dias_completos"`
	DiasEquivalentes            float64 `json:"dias_equivalentes"`
	ValorDia                    float64 `json:"valor_dia"`
	Personas                    int     `json:"personas"`
	PersonasDesde               int     `json:"personas_desde"`
	PersonasHasta               int     `json:"personas_hasta"`
	MontoDiasCompletos          float64 `json:"monto_dias_completos"`
	MontoProrrateoEntrada       float64 `json:"monto_prorrateo_entrada"`
	MontoProrrateoIntermedio    float64 `json:"monto_prorrateo_intermedio"`
	MontoProrrateoSalida        float64 `json:"monto_prorrateo_salida"`
	MontoTotal                  float64 `json:"monto_total"`
	Moneda                      string  `json:"moneda"`
	HoraCheckIn                 string  `json:"hora_check_in"`
	HoraCheckOut                string  `json:"hora_check_out"`
	MinutosVentanaDia           int64   `json:"minutos_ventana_dia"`
	MinutosProrrateoEntrada     int64   `json:"minutos_prorrateo_entrada"`
	MinutosProrrateoIntermedio  int64   `json:"minutos_prorrateo_intermedio"`
	MinutosProrrateoSalida      int64   `json:"minutos_prorrateo_salida"`
	MinutosProrrateoFueraWindow int64   `json:"minutos_prorrateo_fuera_ventana"`
	ReglaProrrateo              string  `json:"regla_prorrateo"`
	FechaInicio                 string  `json:"fecha_inicio"`
	FechaCorte                  string  `json:"fecha_corte"`
}

// EmpresaTarifaPorDiaAplicacionMasivaResultado resume la aplicacion masiva por estaciones.
type EmpresaTarifaPorDiaAplicacionMasivaResultado struct {
	EmpresaID           int64   `json:"empresa_id"`
	ValorDia            float64 `json:"valor_dia"`
	PersonasDesde       int     `json:"personas_desde"`
	PersonasHasta       int     `json:"personas_hasta"`
	HoraCheckIn         string  `json:"hora_check_in"`
	HoraCheckOut        string  `json:"hora_check_out"`
	EstacionesObjetivo  int     `json:"estaciones_objetivo"`
	TarifasCreadas      int     `json:"tarifas_creadas"`
	TarifasActualizadas int     `json:"tarifas_actualizadas"`
	TarifaIDs           []int64 `json:"tarifa_ids,omitempty"`
}

type empresaTarifaPorDiaEstacionRef struct {
	ID     int64
	Codigo string
	Nombre string
}

type tarifaPorDiaCalculoInterno struct {
	diasCobrados             int
	diasCompletos            int
	diasEquivalentes         float64
	windowMinutes            int64
	outsideEntryMinutes      int64
	outsideIntermedioMinutes int64
	outsideSalidaMinutes     int64
	outsideTotalMinutes      int64
	montoDiasCompletos       float64
	montoProrrateoEntrada    float64
	montoProrrateoIntermedio float64
	montoProrrateoSalida     float64
	montoTotal               float64
}

// EnsureEmpresaTarifasPorDiaSchema crea/migra tabla de tarifas diarias por estacion.
func EnsureEmpresaTarifasPorDiaSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_tarifas_por_dia (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			nombre_tarifa TEXT,
			estacion_id INTEGER NOT NULL,
			estacion_codigo TEXT,
			estacion_nombre TEXT,
			servicio_nombre TEXT DEFAULT 'hospedaje',
			valor_dia REAL NOT NULL DEFAULT 0,
			personas_desde INTEGER NOT NULL DEFAULT 1,
			personas_hasta INTEGER NOT NULL DEFAULT 0,
			hora_check_in TEXT DEFAULT '15:00',
			hora_check_out TEXT DEFAULT '12:00',
			moneda TEXT DEFAULT 'COP',
			prioridad INTEGER DEFAULT 1,
			aplicar_automaticamente INTEGER NOT NULL DEFAULT 1,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_tarifas_por_dia_empresa_estado ON empresa_tarifas_por_dia(empresa_id, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_tarifas_por_dia_empresa_estacion ON empresa_tarifas_por_dia(empresa_id, estacion_id);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}
	if _, err := dbConn.Exec(`DROP INDEX IF EXISTS ux_empresa_tarifas_por_dia_estacion`); err != nil {
		return err
	}
	if err := dropLegacyEmpresaTarifasPorDiaUniqueIndexes(dbConn); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_dia", "estacion_codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_dia", "nombre_tarifa", "TEXT"); err != nil {
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
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_dia", "personas_desde", "INTEGER NOT NULL DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_por_dia", "personas_hasta", "INTEGER NOT NULL DEFAULT 0"); err != nil {
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
	if _, err := dbConn.Exec(`CREATE INDEX IF NOT EXISTS ix_empresa_tarifas_por_dia_personas ON empresa_tarifas_por_dia(empresa_id, estacion_id, personas_desde, personas_hasta, estado);`); err != nil {
		return err
	}

	return nil
}

func dropLegacyEmpresaTarifasPorDiaUniqueIndexes(dbConn *sql.DB) error {
	rows, err := dbConn.Query(`SELECT indexname, indexdef
		FROM pg_indexes
		WHERE schemaname = current_schema()
			AND tablename = 'empresa_tarifas_por_dia'`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var indexName string
		var indexDef string
		if err := rows.Scan(&indexName, &indexDef); err != nil {
			return err
		}
		lowerDef := strings.ToLower(indexDef)
		if !strings.Contains(lowerDef, "unique index") {
			continue
		}
		compactDef := strings.ReplaceAll(lowerDef, " ", "")
		if !strings.Contains(compactDef, "(empresa_id,estacion_id)") {
			continue
		}
		if strings.Contains(compactDef, "personas_desde") || strings.Contains(compactDef, "personas_hasta") {
			continue
		}
		if _, err := dbConn.Exec(fmt.Sprintf(`DROP INDEX IF EXISTS %s`, quotePostgresIdentifier(indexName))); err != nil {
			return err
		}
	}
	return rows.Err()
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

func normalizeTarifaPorDiaPersonasDesde(v int) int {
	if v <= 0 {
		return 1
	}
	if v > 999 {
		return 999
	}
	return v
}

func normalizeTarifaPorDiaPersonasHasta(desde, hasta int) int {
	if hasta <= 0 {
		return 0
	}
	if hasta < desde {
		return desde
	}
	if hasta > 999 {
		return 999
	}
	return hasta
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
	payload.NombreTarifa = strings.TrimSpace(payload.NombreTarifa)
	payload.ServicioNombre = normalizeTarifaPorDiaServicio(payload.ServicioNombre)
	if payload.NombreTarifa == "" {
		payload.NombreTarifa = payload.ServicioNombre
	}
	payload.ValorDia = round2(payload.ValorDia)
	payload.HoraCheckIn = horaCheckIn
	payload.HoraCheckOut = horaCheckOut
	payload.PersonasDesde = normalizeTarifaPorDiaPersonasDesde(payload.PersonasDesde)
	payload.PersonasHasta = normalizeTarifaPorDiaPersonasHasta(payload.PersonasDesde, payload.PersonasHasta)
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
		&item.NombreTarifa,
		&item.EstacionID,
		&item.EstacionCodigo,
		&item.EstacionNombre,
		&item.ServicioNombre,
		&item.ValorDia,
		&item.PersonasDesde,
		&item.PersonasHasta,
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
	if err := EnsureEmpresaTarifasPorDiaSchema(dbConn); err != nil {
		return 0, err
	}
	if err := normalizeEmpresaTarifaPorDiaPayload(&payload); err != nil {
		return 0, err
	}

	aplicarAuto := 0
	if payload.AplicarAutomaticamente {
		aplicarAuto = 1
	}

	return insertSQLCompat(dbConn, `INSERT INTO empresa_tarifas_por_dia (
		empresa_id,
		nombre_tarifa,
		estacion_id,
		estacion_codigo,
		estacion_nombre,
		servicio_nombre,
		valor_dia,
		personas_desde,
		personas_hasta,
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
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		payload.EmpresaID,
		payload.NombreTarifa,
		payload.EstacionID,
		payload.EstacionCodigo,
		payload.EstacionNombre,
		payload.ServicioNombre,
		payload.ValorDia,
		payload.PersonasDesde,
		payload.PersonasHasta,
		payload.HoraCheckIn,
		payload.HoraCheckOut,
		payload.Moneda,
		payload.Prioridad,
		aplicarAuto,
		payload.UsuarioCreador,
		payload.Estado,
		payload.Observaciones,
	)
}

// UpdateEmpresaTarifaPorDia actualiza una tarifa diaria existente.
func UpdateEmpresaTarifaPorDia(dbConn *sql.DB, payload EmpresaTarifaPorDia) error {
	if err := EnsureEmpresaTarifasPorDiaSchema(dbConn); err != nil {
		return err
	}
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
		nombre_tarifa = ?,
		estacion_codigo = ?,
		estacion_nombre = ?,
		servicio_nombre = ?,
		valor_dia = ?,
		personas_desde = ?,
		personas_hasta = ?,
		hora_check_in = ?,
		hora_check_out = ?,
		moneda = ?,
		prioridad = ?,
		aplicar_automaticamente = ?,
		usuario_creador = ?,
		estado = ?,
		observaciones = ?,
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ? AND id = ?`,
		payload.EstacionID,
		payload.NombreTarifa,
		payload.EstacionCodigo,
		payload.EstacionNombre,
		payload.ServicioNombre,
		payload.ValorDia,
		payload.PersonasDesde,
		payload.PersonasHasta,
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
	if err := EnsureEmpresaTarifasPorDiaSchema(dbConn); err != nil {
		return err
	}
	if empresaID <= 0 || id <= 0 {
		return fmt.Errorf("empresa_id e id son obligatorios")
	}
	nextEstado := normalizeTarifaPorDiaEstado(estado)
	res, err := dbConn.Exec(`UPDATE empresa_tarifas_por_dia
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

// DeleteEmpresaTarifaPorDia elimina una tarifa diaria.
func DeleteEmpresaTarifaPorDia(dbConn *sql.DB, empresaID, id int64) error {
	if err := EnsureEmpresaTarifasPorDiaSchema(dbConn); err != nil {
		return err
	}
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
	if err := EnsureEmpresaTarifasPorDiaSchema(dbConn); err != nil {
		return nil, err
	}
	if empresaID <= 0 || id <= 0 {
		return nil, fmt.Errorf("empresa_id e id son obligatorios")
	}
	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(nombre_tarifa, ''),
		estacion_id,
		COALESCE(estacion_codigo, ''),
		COALESCE(estacion_nombre, ''),
		COALESCE(servicio_nombre, 'hospedaje'),
		COALESCE(valor_dia, 0),
		COALESCE(personas_desde, 1),
		COALESCE(personas_hasta, 0),
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
	if err := EnsureEmpresaTarifasPorDiaSchema(dbConn); err != nil {
		return nil, err
	}
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
		COALESCE(nombre_tarifa, ''),
		estacion_id,
		COALESCE(estacion_codigo, ''),
		COALESCE(estacion_nombre, ''),
		COALESCE(servicio_nombre, 'hospedaje'),
		COALESCE(valor_dia, 0),
		COALESCE(personas_desde, 1),
		COALESCE(personas_hasta, 0),
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
	if filter.Personas > 0 {
		personas := normalizeTarifaPorDiaPersonasDesde(filter.Personas)
		query += ` AND COALESCE(personas_desde, 1) <= ? AND (COALESCE(personas_hasta, 0) = 0 OR COALESCE(personas_hasta, 0) >= ?)`
		args = append(args, personas, personas)
	}

	query += ` ORDER BY estacion_id ASC, COALESCE(personas_desde, 1) ASC, COALESCE(personas_hasta, 0) ASC, prioridad ASC, id ASC LIMIT ?`
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

func queryEmpresaTarifaPorDiaEstacion(dbConn *sql.DB, empresaID, estacionID int64, personas int, requireAutomatic bool) (*EmpresaTarifaPorDia, error) {
	if err := EnsureEmpresaTarifasPorDiaSchema(dbConn); err != nil {
		return nil, err
	}
	if empresaID <= 0 || estacionID <= 0 {
		return nil, fmt.Errorf("empresa_id y estacion_id son obligatorios")
	}
	personas = normalizeTarifaPorDiaPersonasDesde(personas)

	query := `SELECT
		id,
		empresa_id,
		COALESCE(nombre_tarifa, ''),
		estacion_id,
		COALESCE(estacion_codigo, ''),
		COALESCE(estacion_nombre, ''),
		COALESCE(servicio_nombre, 'hospedaje'),
		COALESCE(valor_dia, 0),
		COALESCE(personas_desde, 1),
		COALESCE(personas_hasta, 0),
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
		AND COALESCE(personas_desde, 1) <= ?
		AND (COALESCE(personas_hasta, 0) = 0 OR COALESCE(personas_hasta, 0) >= ?)
		AND COALESCE(estado, 'activo') = 'activo'`
	args := []interface{}{empresaID, estacionID, personas, personas}
	if requireAutomatic {
		query += ` AND COALESCE(aplicar_automaticamente, 1) = 1`
	}
	query += ` ORDER BY prioridad ASC, COALESCE(personas_desde, 1) DESC, COALESCE(personas_hasta, 0) ASC, id ASC LIMIT 1`

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
	return GetEmpresaTarifaPorDiaActivaPorPersonas(dbConn, empresaID, estacionID, 1)
}

// GetEmpresaTarifaPorDiaActivaPorPersonas devuelve la tarifa activa de una estacion para la ocupacion indicada.
func GetEmpresaTarifaPorDiaActivaPorPersonas(dbConn *sql.DB, empresaID, estacionID int64, personas int) (*EmpresaTarifaPorDia, error) {
	return queryEmpresaTarifaPorDiaEstacion(dbConn, empresaID, estacionID, personas, false)
}

// GetEmpresaTarifaPorDiaAplicable devuelve la tarifa activa y automatica de una estacion.
func GetEmpresaTarifaPorDiaAplicable(dbConn *sql.DB, empresaID, estacionID int64) (*EmpresaTarifaPorDia, error) {
	return GetEmpresaTarifaPorDiaAplicablePorPersonas(dbConn, empresaID, estacionID, 1)
}

// GetEmpresaTarifaPorDiaAplicablePorPersonas devuelve la tarifa automatica para la ocupacion indicada.
func GetEmpresaTarifaPorDiaAplicablePorPersonas(dbConn *sql.DB, empresaID, estacionID int64, personas int) (*EmpresaTarifaPorDia, error) {
	return queryEmpresaTarifaPorDiaEstacion(dbConn, empresaID, estacionID, personas, true)
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

func tarifaPorDiaWindowMinutes(horaCheckIn, horaCheckOut string) int64 {
	checkInHour, checkInMinute, errIn := parseTarifaPorDiaHora(horaCheckIn)
	if errIn != nil {
		checkInHour, checkInMinute = 15, 0
	}
	checkOutHour, checkOutMinute, errOut := parseTarifaPorDiaHora(horaCheckOut)
	if errOut != nil {
		checkOutHour, checkOutMinute = 12, 0
	}

	checkInTotal := (checkInHour * 60) + checkInMinute
	checkOutTotal := (checkOutHour * 60) + checkOutMinute

	if checkInTotal == checkOutTotal {
		return 24 * 60
	}
	if checkInTotal < checkOutTotal {
		return int64(checkOutTotal - checkInTotal)
	}
	return int64((24*60 - checkInTotal) + checkOutTotal)
}

func tarifaPorDiaIsInsideWindow(t time.Time, horaCheckIn, horaCheckOut string) bool {
	checkInHour, checkInMinute, errIn := parseTarifaPorDiaHora(horaCheckIn)
	if errIn != nil {
		checkInHour, checkInMinute = 15, 0
	}
	checkOutHour, checkOutMinute, errOut := parseTarifaPorDiaHora(horaCheckOut)
	if errOut != nil {
		checkOutHour, checkOutMinute = 12, 0
	}

	checkInTotal := (checkInHour * 60) + checkInMinute
	checkOutTotal := (checkOutHour * 60) + checkOutMinute
	total := (t.Hour() * 60) + t.Minute()

	if checkInTotal == checkOutTotal {
		return true
	}
	if checkInTotal < checkOutTotal {
		return total >= checkInTotal && total < checkOutTotal
	}
	return total >= checkInTotal || total < checkOutTotal
}

func tarifaPorDiaWindowStartForInside(t time.Time, horaCheckIn, horaCheckOut string) time.Time {
	location := t.Location()
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

	baseDate := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, location)
	checkInToday := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), checkInHour, checkInMinute, 0, 0, location)

	checkInTotal := (checkInHour * 60) + checkInMinute
	checkOutTotal := (checkOutHour * 60) + checkOutMinute
	total := (t.Hour() * 60) + t.Minute()

	if checkInTotal == checkOutTotal {
		if total >= checkInTotal {
			return checkInToday
		}
		return checkInToday.Add(-24 * time.Hour)
	}

	if checkInTotal > checkOutTotal {
		if total >= checkInTotal {
			return checkInToday
		}
		return checkInToday.Add(-24 * time.Hour)
	}

	return checkInToday
}

func tarifaPorDiaNextTransition(cursor time.Time, horaCheckIn, horaCheckOut string) time.Time {
	location := cursor.Location()
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

	baseDate := time.Date(cursor.Year(), cursor.Month(), cursor.Day(), 0, 0, 0, 0, location)
	var next time.Time

	for i := -1; i <= 2; i++ {
		day := baseDate.AddDate(0, 0, i)
		candidates := []time.Time{
			time.Date(day.Year(), day.Month(), day.Day(), checkInHour, checkInMinute, 0, 0, location),
			time.Date(day.Year(), day.Month(), day.Day(), checkOutHour, checkOutMinute, 0, 0, location),
		}
		for _, candidate := range candidates {
			if !candidate.After(cursor) {
				continue
			}
			if next.IsZero() || candidate.Before(next) {
				next = candidate
			}
		}
	}

	if next.IsZero() {
		return cursor.Add(24 * time.Hour)
	}
	return next
}

func calcularInternoTarifaPorDia(tarifa EmpresaTarifaPorDia, fechaInicio, fechaCorte time.Time) tarifaPorDiaCalculoInterno {
	valorDia := round2(tarifa.ValorDia)
	if valorDia < 0 {
		valorDia = 0
	}

	if fechaInicio.IsZero() {
		return tarifaPorDiaCalculoInterno{}
	}
	if fechaCorte.IsZero() || fechaCorte.Before(fechaInicio) {
		fechaCorte = fechaInicio
	}

	windowMinutes := tarifaPorDiaWindowMinutes(tarifa.HoraCheckIn, tarifa.HoraCheckOut)
	if windowMinutes <= 0 {
		windowMinutes = 24 * 60
	}
	windowSeconds := float64(windowMinutes * 60)

	outsideEntrySeconds := float64(0)
	outsideInterSeconds := float64(0)
	outsideSalidaSeconds := float64(0)
	pendingOutsideSeconds := float64(0)
	windowTouched := make(map[int64]struct{})
	seenInside := false

	if !fechaCorte.After(fechaInicio) {
		if tarifaPorDiaIsInsideWindow(fechaInicio, tarifa.HoraCheckIn, tarifa.HoraCheckOut) {
			windowStart := tarifaPorDiaWindowStartForInside(fechaInicio, tarifa.HoraCheckIn, tarifa.HoraCheckOut)
			windowTouched[windowStart.Unix()] = struct{}{}
			seenInside = true
		}
	} else {
		cursor := fechaInicio
		for cursor.Before(fechaCorte) {
			next := tarifaPorDiaNextTransition(cursor, tarifa.HoraCheckIn, tarifa.HoraCheckOut)
			if !next.After(cursor) {
				next = cursor.Add(time.Minute)
			}
			if next.After(fechaCorte) {
				next = fechaCorte
			}

			segmentSeconds := next.Sub(cursor).Seconds()
			if segmentSeconds < 0 {
				segmentSeconds = 0
			}

			if tarifaPorDiaIsInsideWindow(cursor, tarifa.HoraCheckIn, tarifa.HoraCheckOut) {
				if pendingOutsideSeconds > 0 {
					outsideInterSeconds += pendingOutsideSeconds
					pendingOutsideSeconds = 0
				}
				windowStart := tarifaPorDiaWindowStartForInside(cursor, tarifa.HoraCheckIn, tarifa.HoraCheckOut)
				windowTouched[windowStart.Unix()] = struct{}{}
				seenInside = true
			} else {
				if !seenInside {
					outsideEntrySeconds += segmentSeconds
				} else {
					pendingOutsideSeconds += segmentSeconds
				}
			}

			cursor = next
		}
	}

	if pendingOutsideSeconds > 0 {
		outsideSalidaSeconds += pendingOutsideSeconds
	}

	diasCompletos := len(windowTouched)
	outsideTotalSeconds := outsideEntrySeconds + outsideInterSeconds + outsideSalidaSeconds

	diasEquivalentes := float64(diasCompletos)
	if outsideTotalSeconds > 0 {
		diasEquivalentes += outsideTotalSeconds / windowSeconds
	}

	diasCobrados := diasCompletos
	if diasEquivalentes > 0 {
		if diasCobrados == 0 || diasEquivalentes > float64(diasCompletos) {
			diasCobrados = int(math.Ceil(diasEquivalentes))
		}
	}

	montoDiasCompletos := round2(float64(diasCompletos) * valorDia)
	montoProrrateoEntrada := round2((outsideEntrySeconds / windowSeconds) * valorDia)
	montoProrrateoIntermedio := round2((outsideInterSeconds / windowSeconds) * valorDia)
	montoProrrateoSalida := round2((outsideSalidaSeconds / windowSeconds) * valorDia)
	montoTotal := round2(montoDiasCompletos + montoProrrateoEntrada + montoProrrateoIntermedio + montoProrrateoSalida)

	return tarifaPorDiaCalculoInterno{
		diasCobrados:             diasCobrados,
		diasCompletos:            diasCompletos,
		diasEquivalentes:         round2(diasEquivalentes),
		windowMinutes:            windowMinutes,
		outsideEntryMinutes:      int64(math.Round(outsideEntrySeconds / 60.0)),
		outsideIntermedioMinutes: int64(math.Round(outsideInterSeconds / 60.0)),
		outsideSalidaMinutes:     int64(math.Round(outsideSalidaSeconds / 60.0)),
		outsideTotalMinutes:      int64(math.Round(outsideTotalSeconds / 60.0)),
		montoDiasCompletos:       montoDiasCompletos,
		montoProrrateoEntrada:    montoProrrateoEntrada,
		montoProrrateoIntermedio: montoProrrateoIntermedio,
		montoProrrateoSalida:     montoProrrateoSalida,
		montoTotal:               montoTotal,
	}
}

// CalcularMontoTarifaPorDia calcula dias cobrados y monto total de la tarifa diaria.
func CalcularMontoTarifaPorDia(tarifa EmpresaTarifaPorDia, fechaInicio, fechaCorte time.Time) (int, float64) {
	calculo := calcularInternoTarifaPorDia(tarifa, fechaInicio, fechaCorte)
	return calculo.diasCobrados, calculo.montoTotal
}

// CalcularDetalleTarifaPorDia construye el detalle completo del calculo diario.
func CalcularDetalleTarifaPorDia(tarifa EmpresaTarifaPorDia, fechaInicio, fechaCorte time.Time) EmpresaTarifaPorDiaCalculo {
	return CalcularDetalleTarifaPorDiaConPersonas(tarifa, fechaInicio, fechaCorte, 1)
}

// CalcularDetalleTarifaPorDiaConPersonas construye el detalle incluyendo la ocupacion usada para seleccionar la tarifa.
func CalcularDetalleTarifaPorDiaConPersonas(tarifa EmpresaTarifaPorDia, fechaInicio, fechaCorte time.Time, personas int) EmpresaTarifaPorDiaCalculo {
	calculo := calcularInternoTarifaPorDia(tarifa, fechaInicio, fechaCorte)
	if fechaCorte.IsZero() {
		fechaCorte = fechaInicio
	}
	personas = normalizeTarifaPorDiaPersonasDesde(personas)
	return EmpresaTarifaPorDiaCalculo{
		TarifaID:                    tarifa.ID,
		EstacionID:                  tarifa.EstacionID,
		DiasCobrados:                calculo.diasCobrados,
		DiasCompletos:               calculo.diasCompletos,
		DiasEquivalentes:            calculo.diasEquivalentes,
		ValorDia:                    round2(tarifa.ValorDia),
		Personas:                    personas,
		PersonasDesde:               normalizeTarifaPorDiaPersonasDesde(tarifa.PersonasDesde),
		PersonasHasta:               normalizeTarifaPorDiaPersonasHasta(normalizeTarifaPorDiaPersonasDesde(tarifa.PersonasDesde), tarifa.PersonasHasta),
		MontoDiasCompletos:          calculo.montoDiasCompletos,
		MontoProrrateoEntrada:       calculo.montoProrrateoEntrada,
		MontoProrrateoIntermedio:    calculo.montoProrrateoIntermedio,
		MontoProrrateoSalida:        calculo.montoProrrateoSalida,
		MontoTotal:                  calculo.montoTotal,
		Moneda:                      normalizeTarifaPorDiaMoneda(tarifa.Moneda),
		HoraCheckIn:                 tarifa.HoraCheckIn,
		HoraCheckOut:                tarifa.HoraCheckOut,
		MinutosVentanaDia:           calculo.windowMinutes,
		MinutosProrrateoEntrada:     calculo.outsideEntryMinutes,
		MinutosProrrateoIntermedio:  calculo.outsideIntermedioMinutes,
		MinutosProrrateoSalida:      calculo.outsideSalidaMinutes,
		MinutosProrrateoFueraWindow: calculo.outsideTotalMinutes,
		ReglaProrrateo:              "fuera_ventana_checkin_checkout",
		FechaInicio:                 fechaInicio.Format("2006-01-02 15:04:05"),
		FechaCorte:                  fechaCorte.Format("2006-01-02 15:04:05"),
	}
}

func mergeTarifaPorDiaEstacionRef(dest map[int64]empresaTarifaPorDiaEstacionRef, ref empresaTarifaPorDiaEstacionRef, empresaID int64) {
	if ref.ID <= 0 {
		return
	}
	current, exists := dest[ref.ID]
	if !exists {
		current = empresaTarifaPorDiaEstacionRef{ID: ref.ID}
	}
	if strings.TrimSpace(current.Codigo) == "" {
		if strings.TrimSpace(ref.Codigo) != "" {
			current.Codigo = strings.TrimSpace(ref.Codigo)
		} else {
			current.Codigo = fmt.Sprintf("EST-%d-%d", empresaID, ref.ID)
		}
	}
	if strings.TrimSpace(ref.Nombre) != "" {
		current.Nombre = strings.TrimSpace(ref.Nombre)
	}
	if strings.TrimSpace(current.Nombre) == "" {
		current.Nombre = fmt.Sprintf("Estacion %d", ref.ID)
	}
	dest[ref.ID] = current
}

func listEmpresaTarifaPorDiaStationRefs(dbConn *sql.DB, empresaID int64) ([]empresaTarifaPorDiaEstacionRef, error) {
	refs := make(map[int64]empresaTarifaPorDiaEstacionRef)

	tarifas, err := ListEmpresaTarifasPorDia(dbConn, empresaID, EmpresaTarifaPorDiaFilter{IncludeInactive: true, Limit: 2000})
	if err != nil {
		return nil, err
	}
	for _, tarifa := range tarifas {
		mergeTarifaPorDiaEstacionRef(refs, empresaTarifaPorDiaEstacionRef{ID: tarifa.EstacionID, Codigo: tarifa.EstacionCodigo, Nombre: tarifa.EstacionNombre}, empresaID)
	}

	hasCarritos, err := tableExists(dbConn, "carritos_compras")
	if err != nil {
		return nil, err
	}
	if hasCarritos {
		rowsCarritos, err := dbConn.Query(`SELECT DISTINCT
			COALESCE(referencia_externa, ''),
			COALESCE(codigo, ''),
			COALESCE(nombre, '')
		FROM carritos_compras
		WHERE empresa_id = ?`, empresaID)
		if err != nil {
			return nil, err
		}
		for rowsCarritos.Next() {
			var referenciaExterna string
			var codigo string
			var nombre string
			if err := rowsCarritos.Scan(&referenciaExterna, &codigo, &nombre); err != nil {
				_ = rowsCarritos.Close()
				return nil, err
			}
			estacionID := parseReservaHotelEstacionID(referenciaExterna, codigo, empresaID)
			if estacionID <= 0 {
				continue
			}
			mergeTarifaPorDiaEstacionRef(refs, empresaTarifaPorDiaEstacionRef{ID: estacionID, Codigo: codigo, Nombre: nombre}, empresaID)
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
		rowsReservas, err := dbConn.Query(`SELECT DISTINCT
			COALESCE(estacion_id, 0)
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
			mergeTarifaPorDiaEstacionRef(refs, empresaTarifaPorDiaEstacionRef{ID: estacionID}, empresaID)
		}
		if err := rowsReservas.Err(); err != nil {
			_ = rowsReservas.Close()
			return nil, err
		}
		_ = rowsReservas.Close()
	}

	out := make([]empresaTarifaPorDiaEstacionRef, 0, len(refs))
	for _, ref := range refs {
		out = append(out, ref)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func findEmpresaTarifaPorDiaByStationPersonas(dbConn *sql.DB, empresaID, estacionID int64, personasDesde, personasHasta int) (*EmpresaTarifaPorDia, error) {
	personasDesde = normalizeTarifaPorDiaPersonasDesde(personasDesde)
	personasHasta = normalizeTarifaPorDiaPersonasHasta(personasDesde, personasHasta)
	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(nombre_tarifa, ''),
		estacion_id,
		COALESCE(estacion_codigo, ''),
		COALESCE(estacion_nombre, ''),
		COALESCE(servicio_nombre, 'hospedaje'),
		COALESCE(valor_dia, 0),
		COALESCE(personas_desde, 1),
		COALESCE(personas_hasta, 0),
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
		AND COALESCE(personas_desde, 1) = ?
		AND COALESCE(personas_hasta, 0) = ?
	ORDER BY prioridad ASC, id ASC
	LIMIT 1`, empresaID, estacionID, personasDesde, personasHasta)

	item, err := scanEmpresaTarifaPorDia(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

// ApplyEmpresaTarifaPorDiaToAllStations aplica una misma regla de tarifa diaria a todas las estaciones detectadas de la empresa.
func ApplyEmpresaTarifaPorDiaToAllStations(dbConn *sql.DB, template EmpresaTarifaPorDia) (*EmpresaTarifaPorDiaAplicacionMasivaResultado, error) {
	if template.EmpresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaTarifasPorDiaSchema(dbConn); err != nil {
		return nil, err
	}

	normalized := template
	if normalized.EstacionID <= 0 {
		normalized.EstacionID = 1
	}
	if err := normalizeEmpresaTarifaPorDiaPayload(&normalized); err != nil {
		return nil, err
	}

	refs, err := listEmpresaTarifaPorDiaStationRefs(dbConn, template.EmpresaID)
	if err != nil {
		return nil, err
	}
	if template.EstacionID > 0 {
		base := make(map[int64]empresaTarifaPorDiaEstacionRef, len(refs)+1)
		for _, ref := range refs {
			base[ref.ID] = ref
		}
		mergeTarifaPorDiaEstacionRef(base, empresaTarifaPorDiaEstacionRef{ID: template.EstacionID, Codigo: template.EstacionCodigo, Nombre: template.EstacionNombre}, template.EmpresaID)
		refs = refs[:0]
		for _, ref := range base {
			refs = append(refs, ref)
		}
		sort.Slice(refs, func(i, j int) bool {
			return refs[i].ID < refs[j].ID
		})
	}
	if len(refs) == 0 {
		return nil, fmt.Errorf("no se encontraron estaciones para aplicar la tarifa")
	}

	result := &EmpresaTarifaPorDiaAplicacionMasivaResultado{
		EmpresaID:          template.EmpresaID,
		ValorDia:           normalized.ValorDia,
		PersonasDesde:      normalized.PersonasDesde,
		PersonasHasta:      normalized.PersonasHasta,
		HoraCheckIn:        normalized.HoraCheckIn,
		HoraCheckOut:       normalized.HoraCheckOut,
		EstacionesObjetivo: len(refs),
		TarifaIDs:          make([]int64, 0, len(refs)),
	}

	for _, ref := range refs {
		payload := normalized
		payload.ID = 0
		payload.EmpresaID = template.EmpresaID
		payload.EstacionID = ref.ID
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

		existing, err := findEmpresaTarifaPorDiaByStationPersonas(dbConn, template.EmpresaID, ref.ID, payload.PersonasDesde, payload.PersonasHasta)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			payload.ID = existing.ID
			if err := UpdateEmpresaTarifaPorDia(dbConn, payload); err != nil {
				return nil, err
			}
			result.TarifasActualizadas++
			result.TarifaIDs = append(result.TarifaIDs, existing.ID)
			continue
		}

		id, err := CreateEmpresaTarifaPorDia(dbConn, payload)
		if err != nil {
			return nil, err
		}
		result.TarifasCreadas++
		result.TarifaIDs = append(result.TarifaIDs, id)
	}

	return result, nil
}
