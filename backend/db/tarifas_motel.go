package db

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"sync"
)

// EmpresaTarifaMotel define planes tarifarios especializados para moteles.
type EmpresaTarifaMotel struct {
	ID                  int64   `json:"id"`
	EmpresaID           int64   `json:"empresa_id"`
	EstacionID          int64   `json:"estacion_id"`
	EstacionCodigo      string  `json:"estacion_codigo,omitempty"`
	EstacionNombre      string  `json:"estacion_nombre,omitempty"`
	NombrePlan          string  `json:"nombre_plan"`
	TipoPlan            string  `json:"tipo_plan"`
	CategoriaHabitacion string  `json:"categoria_habitacion,omitempty"`
	DiaSemanaDesde      int     `json:"dia_semana_desde"`
	DiaSemanaHasta      int     `json:"dia_semana_hasta"`
	HoraInicio          string  `json:"hora_inicio"`
	HoraFin             string  `json:"hora_fin"`
	MinutosIncluidos    int     `json:"minutos_incluidos"`
	ValorBase           float64 `json:"valor_base"`
	MinutosExtra        int     `json:"minutos_extra"`
	ValorExtra          float64 `json:"valor_extra"`
	CobrarPorFraccion   bool    `json:"cobrar_por_fraccion"`
	ToleranciaMinutos   int     `json:"tolerancia_minutos"`
	Moneda              string  `json:"moneda,omitempty"`
	Prioridad           int     `json:"prioridad"`
	AplicarAutomatico   bool    `json:"aplicar_automaticamente"`
	FechaCreacion       string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion  string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador      string  `json:"usuario_creador,omitempty"`
	Estado              string  `json:"estado,omitempty"`
	Observaciones       string  `json:"observaciones,omitempty"`
}

// EmpresaTarifaMotelFilter define filtros para consulta de planes motel.
type EmpresaTarifaMotelFilter struct {
	EstacionID      int64
	DiaSemana       int
	TipoPlan        string
	IncludeInactive bool
	Limit           int
}

// EmpresaTarifaMotelCalculo representa una simulacion de cobro motel.
type EmpresaTarifaMotelCalculo struct {
	TarifaID             int64   `json:"tarifa_id"`
	EstacionID           int64   `json:"estacion_id"`
	NombrePlan           string  `json:"nombre_plan"`
	TipoPlan             string  `json:"tipo_plan"`
	MinutosConsumidos    float64 `json:"minutos_consumidos"`
	MinutosIncluidos     int     `json:"minutos_incluidos"`
	MinutosTolerancia    int     `json:"minutos_tolerancia"`
	MinutosFacturables   float64 `json:"minutos_facturables"`
	MinutosExtraCobrados float64 `json:"minutos_extra_cobrados"`
	BloquesExtra         int     `json:"bloques_extra"`
	MontoBase            float64 `json:"monto_base"`
	MontoExtra           float64 `json:"monto_extra"`
	MontoTotal           float64 `json:"monto_total"`
	Moneda               string  `json:"moneda"`
}

var (
	empresaTarifasMotelSchemaEnsured sync.Map
	empresaTarifasMotelSchemaMu      sync.Mutex
)

// EnsureEmpresaTarifasMotelSchema crea/migra la tabla especializada de planes motel.
func EnsureEmpresaTarifasMotelSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	if dbConn == nil {
		return fmt.Errorf("dbConn es obligatorio")
	}
	cacheKey := fmt.Sprintf("%p", dbConn)
	if _, ok := empresaTarifasMotelSchemaEnsured.Load(cacheKey); ok {
		return nil
	}
	empresaTarifasMotelSchemaMu.Lock()
	defer empresaTarifasMotelSchemaMu.Unlock()
	if _, ok := empresaTarifasMotelSchemaEnsured.Load(cacheKey); ok {
		return nil
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_tarifas_motel (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			estacion_id INTEGER NOT NULL,
			estacion_codigo TEXT,
			estacion_nombre TEXT,
			nombre_plan TEXT NOT NULL DEFAULT 'Plan express',
			tipo_plan TEXT NOT NULL DEFAULT 'express',
			categoria_habitacion TEXT,
			dia_semana_desde INTEGER NOT NULL DEFAULT 1,
			dia_semana_hasta INTEGER NOT NULL DEFAULT 7,
			hora_inicio TEXT NOT NULL DEFAULT '00:00',
			hora_fin TEXT NOT NULL DEFAULT '23:59',
			minutos_incluidos INTEGER NOT NULL DEFAULT 180,
			valor_base REAL NOT NULL DEFAULT 0,
			minutos_extra INTEGER NOT NULL DEFAULT 60,
			valor_extra REAL NOT NULL DEFAULT 0,
			cobrar_por_fraccion INTEGER NOT NULL DEFAULT 1,
			tolerancia_minutos INTEGER NOT NULL DEFAULT 0,
			moneda TEXT DEFAULT 'COP',
			prioridad INTEGER DEFAULT 1,
			aplicar_automaticamente INTEGER NOT NULL DEFAULT 1,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_tarifas_motel_empresa_estacion ON empresa_tarifas_motel(empresa_id, estacion_id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_tarifas_motel_empresa_estado ON empresa_tarifas_motel(empresa_id, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_tarifas_motel_tipo ON empresa_tarifas_motel(empresa_id, tipo_plan, estado);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	columns := []struct{ name, ddl string }{
		{"estacion_codigo", "TEXT"},
		{"estacion_nombre", "TEXT"},
		{"nombre_plan", "TEXT NOT NULL DEFAULT 'Plan express'"},
		{"tipo_plan", "TEXT NOT NULL DEFAULT 'express'"},
		{"categoria_habitacion", "TEXT"},
		{"dia_semana_desde", "INTEGER NOT NULL DEFAULT 1"},
		{"dia_semana_hasta", "INTEGER NOT NULL DEFAULT 7"},
		{"hora_inicio", "TEXT NOT NULL DEFAULT '00:00'"},
		{"hora_fin", "TEXT NOT NULL DEFAULT '23:59'"},
		{"minutos_incluidos", "INTEGER NOT NULL DEFAULT 180"},
		{"valor_base", "REAL NOT NULL DEFAULT 0"},
		{"minutos_extra", "INTEGER NOT NULL DEFAULT 60"},
		{"valor_extra", "REAL NOT NULL DEFAULT 0"},
		{"cobrar_por_fraccion", "INTEGER NOT NULL DEFAULT 1"},
		{"tolerancia_minutos", "INTEGER NOT NULL DEFAULT 0"},
		{"moneda", "TEXT DEFAULT 'COP'"},
		{"prioridad", "INTEGER DEFAULT 1"},
		{"aplicar_automaticamente", "INTEGER NOT NULL DEFAULT 1"},
		{"fecha_actualizacion", "TEXT"},
		{"usuario_creador", "TEXT"},
		{"estado", "TEXT DEFAULT 'activo'"},
		{"observaciones", "TEXT"},
	}
	for _, col := range columns {
		if err := ensureColumnIfMissing(dbConn, "empresa_tarifas_motel", col.name, col.ddl); err != nil {
			return err
		}
	}
	empresaTarifasMotelSchemaEnsured.Store(cacheKey, true)
	return nil
}

func normalizeEmpresaTarifaMotelPayload(payload *EmpresaTarifaMotel) error {
	if payload == nil {
		return fmt.Errorf("payload es obligatorio")
	}
	if payload.EmpresaID <= 0 {
		return fmt.Errorf("empresa_id es obligatorio")
	}
	if payload.EstacionID <= 0 {
		return fmt.Errorf("estacion_id es obligatorio")
	}
	payload.NombrePlan = strings.TrimSpace(payload.NombrePlan)
	if payload.NombrePlan == "" {
		payload.NombrePlan = "Plan express"
	}
	payload.TipoPlan = normalizeEmpresaTarifaMotelTipo(payload.TipoPlan)
	payload.CategoriaHabitacion = strings.TrimSpace(payload.CategoriaHabitacion)
	payload.EstacionCodigo = strings.TrimSpace(payload.EstacionCodigo)
	payload.EstacionNombre = strings.TrimSpace(payload.EstacionNombre)
	payload.Moneda = normalizeTarifaPorDiaMoneda(payload.Moneda)
	payload.Estado = normalizeTarifaPorDiaEstado(payload.Estado)
	payload.Prioridad = normalizeTarifaPorDiaPrioridad(payload.Prioridad)
	payload.Observaciones = strings.TrimSpace(payload.Observaciones)
	payload.UsuarioCreador = strings.TrimSpace(payload.UsuarioCreador)
	if payload.DiaSemanaDesde <= 0 {
		payload.DiaSemanaDesde = 1
	}
	if payload.DiaSemanaHasta <= 0 {
		payload.DiaSemanaHasta = 7
	}
	if _, err := normalizeTarifaDiaSemana(payload.DiaSemanaDesde); err != nil {
		return fmt.Errorf("dia_semana_desde invalido")
	}
	if _, err := normalizeTarifaDiaSemana(payload.DiaSemanaHasta); err != nil {
		return fmt.Errorf("dia_semana_hasta invalido")
	}
	if payload.MinutosIncluidos <= 0 {
		payload.MinutosIncluidos = 180
	}
	if payload.MinutosExtra <= 0 {
		payload.MinutosExtra = 60
	}
	if payload.ValorBase < 0 || payload.ValorExtra < 0 {
		return fmt.Errorf("valores de tarifa no pueden ser negativos")
	}
	if payload.ToleranciaMinutos < 0 {
		payload.ToleranciaMinutos = 0
	}
	if payload.ToleranciaMinutos > 1440 {
		payload.ToleranciaMinutos = 1440
	}
	horaInicio, err := normalizeTarifaPorDiaHora(payload.HoraInicio, "00:00")
	if err != nil {
		return fmt.Errorf("hora_inicio invalida")
	}
	horaFin, err := normalizeTarifaPorDiaHora(payload.HoraFin, "23:59")
	if err != nil {
		return fmt.Errorf("hora_fin invalida")
	}
	payload.HoraInicio = horaInicio
	payload.HoraFin = horaFin
	return nil
}

func normalizeEmpresaTarifaMotelTipo(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	switch s {
	case "express", "day_use", "nocturno", "amanecida", "suite", "vip", "promocion":
		return s
	case "noche":
		return "nocturno"
	default:
		return "express"
	}
}

// CreateEmpresaTarifaMotel crea un plan motel.
func CreateEmpresaTarifaMotel(dbConn *sql.DB, payload EmpresaTarifaMotel) (int64, error) {
	if err := normalizeEmpresaTarifaMotelPayload(&payload); err != nil {
		return 0, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_tarifas_motel (
		empresa_id, estacion_id, estacion_codigo, estacion_nombre, nombre_plan, tipo_plan,
		categoria_habitacion, dia_semana_desde, dia_semana_hasta, hora_inicio, hora_fin,
		minutos_incluidos, valor_base, minutos_extra, valor_extra, cobrar_por_fraccion,
		tolerancia_minutos, moneda, prioridad, aplicar_automaticamente, usuario_creador,
		estado, observaciones, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		payload.EmpresaID, payload.EstacionID, payload.EstacionCodigo, payload.EstacionNombre,
		payload.NombrePlan, payload.TipoPlan, payload.CategoriaHabitacion, payload.DiaSemanaDesde,
		payload.DiaSemanaHasta, payload.HoraInicio, payload.HoraFin, payload.MinutosIncluidos,
		payload.ValorBase, payload.MinutosExtra, payload.ValorExtra, boolToInt(payload.CobrarPorFraccion),
		payload.ToleranciaMinutos, payload.Moneda, payload.Prioridad, boolToInt(payload.AplicarAutomatico),
		payload.UsuarioCreador, payload.Estado, payload.Observaciones,
	)
}

// UpdateEmpresaTarifaMotel actualiza un plan motel existente.
func UpdateEmpresaTarifaMotel(dbConn *sql.DB, payload EmpresaTarifaMotel) error {
	if payload.ID <= 0 {
		return fmt.Errorf("id es obligatorio")
	}
	if err := normalizeEmpresaTarifaMotelPayload(&payload); err != nil {
		return err
	}
	res, err := dbConn.Exec(`UPDATE empresa_tarifas_motel
	SET estacion_id = ?, estacion_codigo = ?, estacion_nombre = ?, nombre_plan = ?, tipo_plan = ?,
		categoria_habitacion = ?, dia_semana_desde = ?, dia_semana_hasta = ?, hora_inicio = ?, hora_fin = ?,
		minutos_incluidos = ?, valor_base = ?, minutos_extra = ?, valor_extra = ?, cobrar_por_fraccion = ?,
		tolerancia_minutos = ?, moneda = ?, prioridad = ?, aplicar_automaticamente = ?, usuario_creador = ?,
		estado = ?, observaciones = ?, fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ? AND id = ?`,
		payload.EstacionID, payload.EstacionCodigo, payload.EstacionNombre, payload.NombrePlan, payload.TipoPlan,
		payload.CategoriaHabitacion, payload.DiaSemanaDesde, payload.DiaSemanaHasta, payload.HoraInicio, payload.HoraFin,
		payload.MinutosIncluidos, payload.ValorBase, payload.MinutosExtra, payload.ValorExtra, boolToInt(payload.CobrarPorFraccion),
		payload.ToleranciaMinutos, payload.Moneda, payload.Prioridad, boolToInt(payload.AplicarAutomatico), payload.UsuarioCreador,
		payload.Estado, payload.Observaciones, payload.EmpresaID, payload.ID,
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

// DeleteEmpresaTarifaMotel elimina un plan motel.
func DeleteEmpresaTarifaMotel(dbConn *sql.DB, empresaID, id int64) error {
	if empresaID <= 0 || id <= 0 {
		return fmt.Errorf("empresa_id e id son obligatorios")
	}
	res, err := dbConn.Exec(`DELETE FROM empresa_tarifas_motel WHERE empresa_id = ? AND id = ?`, empresaID, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SetEmpresaTarifaMotelEstado cambia el estado de un plan motel.
func SetEmpresaTarifaMotelEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	if empresaID <= 0 || id <= 0 {
		return fmt.Errorf("empresa_id e id son obligatorios")
	}
	estado = normalizeTarifaPorDiaEstado(estado)
	res, err := dbConn.Exec(`UPDATE empresa_tarifas_motel SET estado = ?, fecha_actualizacion = CURRENT_TIMESTAMP WHERE empresa_id = ? AND id = ?`, estado, empresaID, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetEmpresaTarifaMotelByID obtiene un plan motel por id.
func GetEmpresaTarifaMotelByID(dbConn *sql.DB, empresaID, id int64) (*EmpresaTarifaMotel, error) {
	if empresaID <= 0 || id <= 0 {
		return nil, fmt.Errorf("empresa_id e id son obligatorios")
	}
	row := dbConn.QueryRow(`SELECT
		id, empresa_id, estacion_id, COALESCE(estacion_codigo, ''), COALESCE(estacion_nombre, ''),
		COALESCE(nombre_plan, 'Plan express'), COALESCE(tipo_plan, 'express'), COALESCE(categoria_habitacion, ''),
		COALESCE(dia_semana_desde, 1), COALESCE(dia_semana_hasta, 7), COALESCE(hora_inicio, '00:00'), COALESCE(hora_fin, '23:59'),
		COALESCE(minutos_incluidos, 180), COALESCE(valor_base, 0), COALESCE(minutos_extra, 60), COALESCE(valor_extra, 0),
		COALESCE(cobrar_por_fraccion, 1), COALESCE(tolerancia_minutos, 0), COALESCE(moneda, 'COP'), COALESCE(prioridad, 1),
		COALESCE(aplicar_automaticamente, 1), COALESCE(fecha_creacion, ''), COALESCE(fecha_actualizacion, ''), COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'), COALESCE(observaciones, '')
	FROM empresa_tarifas_motel WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, id)
	return scanEmpresaTarifaMotel(row)
}

// ListEmpresaTarifasMotel lista planes motel.
func ListEmpresaTarifasMotel(dbConn *sql.DB, empresaID int64, filter EmpresaTarifaMotelFilter) ([]EmpresaTarifaMotel, error) {
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
		id, empresa_id, estacion_id, COALESCE(estacion_codigo, ''), COALESCE(estacion_nombre, ''),
		COALESCE(nombre_plan, 'Plan express'), COALESCE(tipo_plan, 'express'), COALESCE(categoria_habitacion, ''),
		COALESCE(dia_semana_desde, 1), COALESCE(dia_semana_hasta, 7), COALESCE(hora_inicio, '00:00'), COALESCE(hora_fin, '23:59'),
		COALESCE(minutos_incluidos, 180), COALESCE(valor_base, 0), COALESCE(minutos_extra, 60), COALESCE(valor_extra, 0),
		COALESCE(cobrar_por_fraccion, 1), COALESCE(tolerancia_minutos, 0), COALESCE(moneda, 'COP'), COALESCE(prioridad, 1),
		COALESCE(aplicar_automaticamente, 1), COALESCE(fecha_creacion, ''), COALESCE(fecha_actualizacion, ''), COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'), COALESCE(observaciones, '')
	FROM empresa_tarifas_motel WHERE empresa_id = ?`
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
		query += ` AND dia_semana_desde <= ? AND dia_semana_hasta >= ?`
		args = append(args, filter.DiaSemana, filter.DiaSemana)
	}
	if strings.TrimSpace(filter.TipoPlan) != "" {
		query += ` AND tipo_plan = ?`
		args = append(args, normalizeEmpresaTarifaMotelTipo(filter.TipoPlan))
	}
	query += ` ORDER BY prioridad ASC, estacion_id ASC, tipo_plan ASC, valor_base ASC, id ASC LIMIT ?`
	args = append(args, filter.Limit)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []EmpresaTarifaMotel{}
	for rows.Next() {
		item, err := scanEmpresaTarifaMotel(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

type empresaTarifaMotelScanner interface {
	Scan(dest ...interface{}) error
}

func scanEmpresaTarifaMotel(scanner empresaTarifaMotelScanner) (*EmpresaTarifaMotel, error) {
	var item EmpresaTarifaMotel
	var cobrar, automatico int
	if err := scanner.Scan(
		&item.ID, &item.EmpresaID, &item.EstacionID, &item.EstacionCodigo, &item.EstacionNombre,
		&item.NombrePlan, &item.TipoPlan, &item.CategoriaHabitacion, &item.DiaSemanaDesde, &item.DiaSemanaHasta,
		&item.HoraInicio, &item.HoraFin, &item.MinutosIncluidos, &item.ValorBase, &item.MinutosExtra, &item.ValorExtra,
		&cobrar, &item.ToleranciaMinutos, &item.Moneda, &item.Prioridad, &automatico, &item.FechaCreacion,
		&item.FechaActualizacion, &item.UsuarioCreador, &item.Estado, &item.Observaciones,
	); err != nil {
		return nil, err
	}
	item.CobrarPorFraccion = cobrar != 0
	item.AplicarAutomatico = automatico != 0
	return &item, nil
}

// CalcularDetalleTarifaMotel calcula el cobro de un plan motel.
func CalcularDetalleTarifaMotel(tarifa EmpresaTarifaMotel, minutosConsumidos float64) EmpresaTarifaMotelCalculo {
	if minutosConsumidos < 0 {
		minutosConsumidos = 0
	}
	facturables := minutosConsumidos - float64(tarifa.ToleranciaMinutos)
	if facturables < 0 {
		facturables = 0
	}
	extraMinutes := facturables - float64(tarifa.MinutosIncluidos)
	if extraMinutes < 0 {
		extraMinutes = 0
	}
	blocks := 0
	if extraMinutes > 0 && tarifa.MinutosExtra > 0 {
		if tarifa.CobrarPorFraccion {
			blocks = int(math.Ceil(extraMinutes / float64(tarifa.MinutosExtra)))
		} else {
			blocks = int(math.Floor(extraMinutes / float64(tarifa.MinutosExtra)))
		}
	}
	montoExtra := float64(blocks) * tarifa.ValorExtra
	return EmpresaTarifaMotelCalculo{
		TarifaID:             tarifa.ID,
		EstacionID:           tarifa.EstacionID,
		NombrePlan:           tarifa.NombrePlan,
		TipoPlan:             tarifa.TipoPlan,
		MinutosConsumidos:    minutosConsumidos,
		MinutosIncluidos:     tarifa.MinutosIncluidos,
		MinutosTolerancia:    tarifa.ToleranciaMinutos,
		MinutosFacturables:   facturables,
		MinutosExtraCobrados: extraMinutes,
		BloquesExtra:         blocks,
		MontoBase:            tarifa.ValorBase,
		MontoExtra:           montoExtra,
		MontoTotal:           tarifa.ValorBase + montoExtra,
		Moneda:               tarifa.Moneda,
	}
}
