package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// CarritoTarifaPorMinutosCalculo representa el ajuste automático por minutos aplicado sobre un carrito de estación.
type CarritoTarifaPorMinutosCalculo struct {
	EmpresaID            int64   `json:"empresa_id"`
	CarritoID            int64   `json:"carrito_id"`
	EstacionID           int64   `json:"estacion_id"`
	TarifaID             int64   `json:"tarifa_id"`
	Aplicada             bool    `json:"aplicada"`
	DiaSemana            int     `json:"dia_semana"`
	MinutosConsumidos    float64 `json:"minutos_consumidos"`
	MinutosFacturables   float64 `json:"minutos_facturables"`
	MinutosTolerancia    int     `json:"minutos_tolerancia"`
	MinutosBase          int     `json:"minutos_base"`
	MinutosExtra         int     `json:"minutos_extra"`
	BloquesExtra         int     `json:"bloques_extra"`
	ValorBase            float64 `json:"valor_base"`
	ValorExtra           float64 `json:"valor_extra"`
	MontoTarifa          float64 `json:"monto_tarifa"`
	Moneda               string  `json:"moneda"`
	ActivadoEn           string  `json:"activado_en"`
	FechaCorte           string  `json:"fecha_corte"`
	FechaFinTarifaActual string  `json:"fecha_fin_tarifa_actual"`
	BaseSubtotal         float64 `json:"base_subtotal"`
	BaseTotal            float64 `json:"base_total"`
	SubtotalFinal        float64 `json:"subtotal_final"`
	TotalFinal           float64 `json:"total_final"`
	MontoMinimoAplicado  bool    `json:"monto_minimo_aplicado"`
	MontoMaximoAplicado  bool    `json:"monto_maximo_aplicado"`
	RedondeoAplicado     float64 `json:"redondeo_aplicado"`
}

// CarritoTarifaPorMinutosResumen resume la tarifa vigente para pintar la tarjeta de estación.
type CarritoTarifaPorMinutosResumen struct {
	TarifaID             int64   `json:"tarifa_id"`
	EstacionID           int64   `json:"estacion_id"`
	Aplicada             bool    `json:"aplicada"`
	DiaSemana            int     `json:"dia_semana"`
	MinutosConsumidos    float64 `json:"minutos_consumidos"`
	MinutosFacturables   float64 `json:"minutos_facturables"`
	MinutosTolerancia    int     `json:"minutos_tolerancia"`
	MinutosBase          int     `json:"minutos_base"`
	MinutosExtra         int     `json:"minutos_extra"`
	BloquesExtra         int     `json:"bloques_extra"`
	ValorBase            float64 `json:"valor_base"`
	ValorExtra           float64 `json:"valor_extra"`
	MontoTarifa          float64 `json:"monto_tarifa"`
	Moneda               string  `json:"moneda"`
	FechaInicioTarifa    string  `json:"fecha_inicio_tarifa"`
	FechaFinTarifaActual string  `json:"fecha_fin_tarifa_actual"`
	TotalActual          float64 `json:"total_actual"`
}

// CarritoTarifasTiempoCalculo resume qué tarifa temporal se aplicó a un carrito.
type CarritoTarifasTiempoCalculo struct {
	TarifaPorMinutos *CarritoTarifaPorMinutosCalculo `json:"tarifa_por_minutos,omitempty"`
	TarifaPorDia     *CarritoTarifaPorDiaCalculo     `json:"tarifa_por_dia,omitempty"`
}

func empresaTarifaPorMinutosTableExistsTx(tx *sql.Tx) (bool, error) {
	var exists bool
	err := queryRowTxSQLCompat(tx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = ANY (current_schemas(false))
			  AND table_name = ?
		)
	`, "empresa_tarifas_por_minutos").Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func getEmpresaTarifaPorMinutosAplicableTx(tx *sql.Tx, empresaID, estacionID int64, diaSemana int) (*EmpresaTarifaPorMinutos, error) {
	row := queryRowTxSQLCompat(tx, `SELECT
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

func getEmpresaTarifaPorMinutosConfiguracionTx(tx *sql.Tx, empresaID int64) (*EmpresaTarifaPorMinutosConfiguracion, error) {
	row := queryRowTxSQLCompat(tx, `SELECT
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

func resolveCarritoTarifaPorMinutosCurrentEnd(activadoAt time.Time, tarifa EmpresaTarifaPorMinutos, detalle EmpresaTarifaPorMinutosCalculo) time.Time {
	minutosTramoActual := tarifa.MinutosBase + detalle.MinutosTolerancia
	if detalle.BloquesExtra > 0 {
		minutosTramoActual = tarifa.MinutosBase + detalle.MinutosTolerancia + detalle.BloquesExtra*tarifa.MinutosExtra
	}
	if minutosTramoActual < tarifa.MinutosBase {
		minutosTramoActual = tarifa.MinutosBase
	}
	return activadoAt.Add(time.Duration(minutosTramoActual) * time.Minute)
}

func refreshCarritoTotalConTarifaPorMinutosTx(tx *sql.Tx, empresaID, carritoID int64, fechaCorte time.Time) (*CarritoTarifaPorMinutosCalculo, error) {
	if err := recalculateCarritoTotalsTx(tx, empresaID, carritoID); err != nil {
		return nil, err
	}

	snapshot := &carritoTarifaPorDiaSnapshot{}
	err := queryRowTxSQLCompat(tx, `SELECT
		COALESCE(subtotal, 0),
		COALESCE(total, 0),
		COALESCE(estado, 'activo'),
		COALESCE(estado_carrito, 'abierto'),
		COALESCE(activado_en, ''),
		COALESCE(pagado_en, ''),
		COALESCE(referencia_externa, ''),
		COALESCE(codigo, ''),
		COALESCE(moneda, 'COP')
	FROM carritos_compras
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, carritoID).Scan(
		&snapshot.Subtotal,
		&snapshot.Total,
		&snapshot.Estado,
		&snapshot.EstadoCarrito,
		&snapshot.ActivadoEn,
		&snapshot.PagadoEn,
		&snapshot.ReferenciaExterna,
		&snapshot.Codigo,
		&snapshot.Moneda,
	)
	if err != nil {
		return nil, err
	}

	calc := &CarritoTarifaPorMinutosCalculo{
		EmpresaID:     empresaID,
		CarritoID:     carritoID,
		Aplicada:      false,
		BaseSubtotal:  round2(snapshot.Subtotal),
		BaseTotal:     round2(snapshot.Total),
		SubtotalFinal: round2(snapshot.Subtotal),
		TotalFinal:    round2(snapshot.Total),
		Moneda:        strings.TrimSpace(strings.ToUpper(snapshot.Moneda)),
		FechaCorte:    fechaCorte.Format("2006-01-02 15:04:05"),
	}
	if calc.Moneda == "" {
		calc.Moneda = "COP"
	}

	exists, err := empresaTarifaPorMinutosTableExistsTx(tx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return calc, nil
	}

	estadoRegistro := strings.TrimSpace(strings.ToLower(snapshot.Estado))
	estadoCarrito := strings.TrimSpace(strings.ToLower(snapshot.EstadoCarrito))
	if estadoRegistro == "" {
		estadoRegistro = "activo"
	}
	if estadoCarrito == "" {
		estadoCarrito = "abierto"
	}
	if estadoRegistro != "activo" || estadoCarrito != "abierto" || strings.TrimSpace(snapshot.PagadoEn) != "" {
		return calc, nil
	}

	estacionID := parseReservaHotelEstacionID(snapshot.ReferenciaExterna, snapshot.Codigo, empresaID)
	if estacionID <= 0 {
		return calc, nil
	}
	calc.EstacionID = estacionID
	calc.ActivadoEn = strings.TrimSpace(snapshot.ActivadoEn)
	if calc.ActivadoEn == "" {
		return calc, nil
	}

	activadoAt, err := parseTarifaPorDiaDateTime(calc.ActivadoEn)
	if err != nil {
		return nil, fmt.Errorf("activado_en invalido para carrito %d", carritoID)
	}
	if fechaCorte.Before(activadoAt) {
		fechaCorte = activadoAt
		calc.FechaCorte = fechaCorte.Format("2006-01-02 15:04:05")
	}

	diaSemana := DayOfWeekISO(fechaCorte)
	tarifa, err := getEmpresaTarifaPorMinutosAplicableTx(tx, empresaID, estacionID, diaSemana)
	if err != nil {
		return nil, err
	}
	if tarifa == nil {
		return calc, nil
	}
	cfg, err := getEmpresaTarifaPorMinutosConfiguracionTx(tx, empresaID)
	if err != nil {
		return nil, err
	}

	minutosConsumidos := fechaCorte.Sub(activadoAt).Minutes()
	if minutosConsumidos < 0 {
		minutosConsumidos = 0
	}
	detalle := CalcularDetalleTarifaPorMinutos(*tarifa, minutosConsumidos, *cfg)

	calc.TarifaID = tarifa.ID
	calc.Aplicada = true
	calc.DiaSemana = diaSemana
	calc.MinutosConsumidos = detalle.MinutosConsumidos
	calc.MinutosFacturables = detalle.MinutosFacturables
	calc.MinutosTolerancia = detalle.MinutosTolerancia
	calc.MinutosBase = tarifa.MinutosBase
	calc.MinutosExtra = tarifa.MinutosExtra
	calc.BloquesExtra = detalle.BloquesExtra
	calc.ValorBase = round2(tarifa.ValorBase)
	calc.ValorExtra = round2(tarifa.ValorExtra)
	calc.MontoTarifa = detalle.MontoTotal
	calc.Moneda = detalle.Moneda
	calc.FechaFinTarifaActual = resolveCarritoTarifaPorMinutosCurrentEnd(activadoAt, *tarifa, detalle).Format("2006-01-02 15:04:05")
	calc.MontoMinimoAplicado = detalle.MontoMinimoAplicado
	calc.MontoMaximoAplicado = detalle.MontoMaximoAplicado
	calc.RedondeoAplicado = detalle.AjusteRedondeo
	calc.SubtotalFinal = round2(calc.BaseSubtotal + detalle.MontoTotal)
	calc.TotalFinal = round2(calc.BaseTotal + detalle.MontoTotal)

	if _, err := execTxSQLCompat(tx, `UPDATE carritos_compras SET
		subtotal = ?,
		total = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`, calc.SubtotalFinal, calc.TotalFinal, empresaID, carritoID); err != nil {
		return nil, err
	}

	return calc, nil
}

// RefreshCarritoTotalConTarifaPorMinutos recalcula totales base e integra cobro automático por minutos si aplica.
func RefreshCarritoTotalConTarifaPorMinutos(dbConn *sql.DB, empresaID, carritoID int64, fechaCorte time.Time) (*CarritoTarifaPorMinutosCalculo, error) {
	if empresaID <= 0 || carritoID <= 0 {
		return nil, fmt.Errorf("empresa_id y carrito_id son obligatorios")
	}
	if fechaCorte.IsZero() {
		fechaCorte = time.Now()
	}

	tx, err := dbConn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	calc, err := refreshCarritoTotalConTarifaPorMinutosTx(tx, empresaID, carritoID, fechaCorte)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return calc, nil
}

// RefreshCarritoTotalConTarifasTiempo aplica tarifa por minutos si existe; si no, mantiene el flujo automático por día.
func RefreshCarritoTotalConTarifasTiempo(dbConn *sql.DB, empresaID, carritoID int64, fechaCorte time.Time) (*CarritoTarifasTiempoCalculo, error) {
	result := &CarritoTarifasTiempoCalculo{}
	minuteCalc, err := RefreshCarritoTotalConTarifaPorMinutos(dbConn, empresaID, carritoID, fechaCorte)
	if err != nil {
		return nil, err
	}
	if minuteCalc != nil && minuteCalc.TarifaID > 0 {
		result.TarifaPorMinutos = minuteCalc
		return result, nil
	}
	dayCalc, err := RefreshCarritoTotalConTarifaPorDia(dbConn, empresaID, carritoID, fechaCorte)
	if err != nil {
		return nil, err
	}
	if dayCalc != nil && dayCalc.TarifaID > 0 {
		result.TarifaPorDia = dayCalc
	}
	return result, nil
}

// RefreshCarritosActivosConTarifasTiempo recalcula masivamente cobros temporales para carritos activos de estaciones.
func RefreshCarritosActivosConTarifasTiempo(dbConn *sql.DB, empresaID int64, fechaCorte time.Time) error {
	if empresaID <= 0 {
		return fmt.Errorf("empresa_id es obligatorio")
	}
	if fechaCorte.IsZero() {
		fechaCorte = time.Now()
	}

	rows, err := querySQLCompat(dbConn, `SELECT id
	FROM carritos_compras
	WHERE empresa_id = ?
		AND COALESCE(estado, 'activo') = 'activo'
		AND lower(COALESCE(estado_carrito, 'abierto')) = 'abierto'
		AND COALESCE(pagado_en, '') = ''
		AND (
			upper(COALESCE(referencia_externa, '')) LIKE 'ESTACION_%'
			OR upper(COALESCE(codigo, '')) LIKE upper(?)
		)
	ORDER BY id ASC`, empresaID, fmt.Sprintf("EST-%d-%%", empresaID))
	if err != nil {
		return err
	}
	defer rows.Close()

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, carritoID := range ids {
		if _, err := RefreshCarritoTotalConTarifasTiempo(dbConn, empresaID, carritoID, fechaCorte); err != nil {
			return err
		}
	}
	return nil
}

// ResolveCarritoTarifaPorMinutosResumen obtiene metadata operativa de la tarifa vigente para pintar estaciones.
func ResolveCarritoTarifaPorMinutosResumen(dbConn *sql.DB, item CarritoCompra, fechaCorte time.Time) (*CarritoTarifaPorMinutosResumen, error) {
	if item.EmpresaID <= 0 || item.ID <= 0 {
		return nil, nil
	}
	estadoRegistro := strings.TrimSpace(strings.ToLower(item.Estado))
	estadoCarrito := strings.TrimSpace(strings.ToLower(item.EstadoCarrito))
	if estadoRegistro == "" {
		estadoRegistro = "activo"
	}
	if estadoCarrito == "" {
		estadoCarrito = "abierto"
	}
	if estadoRegistro != "activo" || estadoCarrito != "abierto" || strings.TrimSpace(item.PagadoEn) != "" {
		return nil, nil
	}
	if fechaCorte.IsZero() {
		fechaCorte = time.Now()
	}
	estacionID := parseReservaHotelEstacionID(item.ReferenciaExterna, item.Codigo, item.EmpresaID)
	if estacionID <= 0 {
		return nil, nil
	}
	activadoAt, err := parseTarifaPorDiaDateTime(item.ActivadoEn)
	if err != nil {
		return nil, nil
	}
	diaSemana := DayOfWeekISO(fechaCorte)
	tarifa, err := GetEmpresaTarifaPorMinutosAplicable(dbConn, item.EmpresaID, estacionID, diaSemana)
	if err != nil || tarifa == nil {
		return nil, err
	}
	cfg, err := GetEmpresaTarifaPorMinutosConfiguracion(dbConn, item.EmpresaID)
	if err != nil {
		return nil, err
	}
	minutosConsumidos := fechaCorte.Sub(activadoAt).Minutes()
	if minutosConsumidos < 0 {
		minutosConsumidos = 0
	}
	detalle := CalcularDetalleTarifaPorMinutos(*tarifa, minutosConsumidos, *cfg)
	return &CarritoTarifaPorMinutosResumen{
		TarifaID:             tarifa.ID,
		EstacionID:           estacionID,
		Aplicada:             true,
		DiaSemana:            diaSemana,
		MinutosConsumidos:    detalle.MinutosConsumidos,
		MinutosFacturables:   detalle.MinutosFacturables,
		MinutosTolerancia:    detalle.MinutosTolerancia,
		MinutosBase:          tarifa.MinutosBase,
		MinutosExtra:         tarifa.MinutosExtra,
		BloquesExtra:         detalle.BloquesExtra,
		ValorBase:            round2(tarifa.ValorBase),
		ValorExtra:           round2(tarifa.ValorExtra),
		MontoTarifa:          detalle.MontoTotal,
		Moneda:               detalle.Moneda,
		FechaInicioTarifa:    activadoAt.Format("2006-01-02 15:04:05"),
		FechaFinTarifaActual: resolveCarritoTarifaPorMinutosCurrentEnd(activadoAt, *tarifa, detalle).Format("2006-01-02 15:04:05"),
		TotalActual:          round2(item.Total),
	}, nil
}
