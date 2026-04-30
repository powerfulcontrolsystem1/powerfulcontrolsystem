package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// CarritoTarifaPorDiaCalculo representa el ajuste diario aplicado sobre un carrito de estacion.
type CarritoTarifaPorDiaCalculo struct {
	EmpresaID      int64   `json:"empresa_id"`
	CarritoID      int64   `json:"carrito_id"`
	EstacionID     int64   `json:"estacion_id"`
	TarifaID       int64   `json:"tarifa_id"`
	Aplicada       bool    `json:"aplicada"`
	DiasCobrados   int     `json:"dias_cobrados"`
	ValorDia       float64 `json:"valor_dia"`
	MontoTarifa    float64 `json:"monto_tarifa"`
	Moneda         string  `json:"moneda"`
	HoraCheckIn    string  `json:"hora_check_in"`
	HoraCheckOut   string  `json:"hora_check_out"`
	ActivadoEn     string  `json:"activado_en"`
	FechaCorte     string  `json:"fecha_corte"`
	BaseSubtotal   float64 `json:"base_subtotal"`
	BaseTotal      float64 `json:"base_total"`
	SubtotalFinal  float64 `json:"subtotal_final"`
	TotalFinal     float64 `json:"total_final"`
	ServicioNombre string  `json:"servicio_nombre"`
}

type carritoTarifaPorDiaSnapshot struct {
	Subtotal          float64
	Total             float64
	Estado            string
	EstadoCarrito     string
	ActivadoEn        string
	PagadoEn          string
	ReferenciaExterna string
	Codigo            string
	Moneda            string
}

func empresaTarifaPorDiaTableExists(dbConn *sql.DB) (bool, error) {
	return tableExists(dbConn, "empresa_tarifas_por_dia")
}

func empresaTarifaPorDiaTableExistsTx(tx *sql.Tx) (bool, error) {
	var exists bool
	err := queryRowTxSQLCompat(tx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = ANY (current_schemas(false))
			  AND table_name = ?
		)
	`, "empresa_tarifas_por_dia").Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func getEmpresaTarifaPorDiaAplicableTx(tx *sql.Tx, empresaID, estacionID int64) (*EmpresaTarifaPorDia, error) {
	row := tx.QueryRow(`SELECT
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
		AND COALESCE(estado, 'activo') = 'activo'
		AND COALESCE(aplicar_automaticamente, 1) = 1
	ORDER BY prioridad ASC, id ASC
	LIMIT 1`, empresaID, estacionID)

	item, err := scanEmpresaTarifaPorDia(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func refreshCarritoTotalConTarifaPorDiaTx(tx *sql.Tx, empresaID, carritoID int64, fechaCorte time.Time) (*CarritoTarifaPorDiaCalculo, error) {
	if err := recalculateCarritoTotalsTx(tx, empresaID, carritoID); err != nil {
		return nil, err
	}

	snapshot := &carritoTarifaPorDiaSnapshot{}
	err := tx.QueryRow(`SELECT
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

	calc := &CarritoTarifaPorDiaCalculo{
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

	exists, err := empresaTarifaPorDiaTableExistsTx(tx)
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

	if strings.TrimSpace(snapshot.ActivadoEn) == "" {
		return calc, nil
	}
	activadoAt, err := parseTarifaPorDiaDateTime(snapshot.ActivadoEn)
	if err != nil {
		return nil, fmt.Errorf("activado_en invalido para carrito %d", carritoID)
	}
	if fechaCorte.Before(activadoAt) {
		fechaCorte = activadoAt
		calc.FechaCorte = fechaCorte.Format("2006-01-02 15:04:05")
	}

	tarifa, err := getEmpresaTarifaPorDiaAplicableTx(tx, empresaID, estacionID)
	if err != nil {
		return nil, err
	}
	if tarifa == nil {
		return calc, nil
	}

	detalle := CalcularDetalleTarifaPorDia(*tarifa, activadoAt, fechaCorte)
	calc.TarifaID = tarifa.ID
	calc.DiasCobrados = detalle.DiasCobrados
	calc.ValorDia = detalle.ValorDia
	calc.MontoTarifa = detalle.MontoTotal
	calc.HoraCheckIn = tarifa.HoraCheckIn
	calc.HoraCheckOut = tarifa.HoraCheckOut
	calc.ServicioNombre = tarifa.ServicioNombre
	calc.Aplicada = detalle.MontoTotal > 0 || detalle.DiasCobrados > 0

	calc.SubtotalFinal = round2(calc.BaseSubtotal + detalle.MontoTotal)
	calc.TotalFinal = round2(calc.BaseTotal + detalle.MontoTotal)

	if _, err := tx.Exec(`UPDATE carritos_compras SET
		subtotal = ?,
		total = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`, calc.SubtotalFinal, calc.TotalFinal, empresaID, carritoID); err != nil {
		return nil, err
	}

	return calc, nil
}

// RefreshCarritoTotalConTarifaPorDia recalcula totales base e integra cobro diario automatico si aplica.
func RefreshCarritoTotalConTarifaPorDia(dbConn *sql.DB, empresaID, carritoID int64, fechaCorte time.Time) (*CarritoTarifaPorDiaCalculo, error) {
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

	calc, err := refreshCarritoTotalConTarifaPorDiaTx(tx, empresaID, carritoID, fechaCorte)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return calc, nil
}

// RefreshCarritosActivosConTarifaPorDia recalcula masivamente cobros diarios para carritos activos de estaciones.
func RefreshCarritosActivosConTarifaPorDia(dbConn *sql.DB, empresaID int64, fechaCorte time.Time) error {
	if empresaID <= 0 {
		return fmt.Errorf("empresa_id es obligatorio")
	}
	if fechaCorte.IsZero() {
		fechaCorte = time.Now()
	}

	exists, err := empresaTarifaPorDiaTableExists(dbConn)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	rows, err := dbConn.Query(`SELECT id
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
		if _, err := RefreshCarritoTotalConTarifaPorDia(dbConn, empresaID, carritoID, fechaCorte); err != nil {
			return err
		}
	}
	return nil
}

// ResolveCarritoTarifaPorDiaResumen obtiene metadata operativa de la tarifa diaria vigente
// para pintar tarjetas de estacion sin depender de estado temporal del navegador.
func ResolveCarritoTarifaPorDiaResumen(dbConn *sql.DB, item CarritoCompra, fechaCorte time.Time) (*CarritoTarifaPorDiaCalculo, error) {
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

	tarifa, err := GetEmpresaTarifaPorDiaAplicable(dbConn, item.EmpresaID, estacionID)
	if err != nil || tarifa == nil {
		return nil, err
	}
	detalle := CalcularDetalleTarifaPorDia(*tarifa, activadoAt, fechaCorte)
	return &CarritoTarifaPorDiaCalculo{
		EmpresaID:      item.EmpresaID,
		CarritoID:      item.ID,
		EstacionID:     estacionID,
		TarifaID:       tarifa.ID,
		Aplicada:       detalle.MontoTotal > 0 || detalle.DiasCobrados > 0,
		DiasCobrados:   detalle.DiasCobrados,
		ValorDia:       detalle.ValorDia,
		MontoTarifa:    detalle.MontoTotal,
		Moneda:         normalizeTarifaPorDiaMoneda(tarifa.Moneda),
		HoraCheckIn:    tarifa.HoraCheckIn,
		HoraCheckOut:   tarifa.HoraCheckOut,
		ActivadoEn:     activadoAt.Format("2006-01-02 15:04:05"),
		FechaCorte:     fechaCorte.Format("2006-01-02 15:04:05"),
		BaseSubtotal:   round2(item.Subtotal),
		BaseTotal:      round2(item.Total),
		SubtotalFinal:  round2(item.Subtotal),
		TotalFinal:     round2(item.Total),
		ServicioNombre: strings.TrimSpace(tarifa.ServicioNombre),
	}, nil
}
