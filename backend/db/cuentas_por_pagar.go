package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	empresaCxPAtomicSchemaFingerprint = "empresa_cxp_pagos:v1:tenant-lock-idempotency-finance-outbox"
	EmpresaCxPPaymentOutboxTopic      = "cuentas_por_pagar.pago_registrado"
)

var (
	ErrEmpresaCxPIdempotencyKeyRequired = errors.New("la clave de idempotencia es obligatoria")
	ErrEmpresaCxPAmountExceedsBalance   = errors.New("el monto supera el saldo pendiente de la cuenta por pagar")
	ErrEmpresaCxPNoPendingBalance       = errors.New("la cuenta por pagar no tiene saldo pendiente")
	ErrEmpresaCxPIdempotencyConflict    = errors.New("la clave de idempotencia ya fue usada con otro abono")
)

// EmpresaCxPAbonoInput is the explicit, tenant-scoped application of one
// supplier payment to one canonical account payable. It records a payment
// allocation, the finance movement and the outbox event in one transaction.
// The raw idempotency key is never persisted.
type EmpresaCxPAbonoInput struct {
	EmpresaID         int64
	CuentaPorPagarID  int64
	Monto             float64
	PeriodoContable   string
	MetodoPago        string
	ReferenciaExterna string
	Concepto          string
	Observaciones     string
	Usuario           string
	IdempotencyKey    string
}

type EmpresaCxPAbonoResult struct {
	EmpresaID            int64   `json:"empresa_id"`
	CuentaPorPagarID     int64   `json:"cartera_id"`
	PagoID               int64   `json:"pago_id"`
	MovimientoFinanzasID int64   `json:"movimiento_finanzas_id"`
	MontoAplicado        float64 `json:"monto_aplicado"`
	SaldoAnterior        float64 `json:"saldo_anterior"`
	SaldoNuevo           float64 `json:"saldo"`
	EstadoCartera        string  `json:"estado_cartera"`
	IdempotentReplay     bool    `json:"idempotent_replay"`
}

type EmpresaCxPPaymentAccountingResult struct {
	EmpresaID        int64 `json:"empresa_id"`
	PagoID           int64 `json:"pago_id"`
	EventoContableID int64 `json:"evento_contable_id"`
	IdempotentReplay bool  `json:"idempotent_replay"`
}

func applyEmpresaCxPAtomicSchemaTx(tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	for _, statement := range empresaCxPAtomicSchemaV1Statements() {
		if _, err := execTxSQLCompat(tx, statement); err != nil {
			return err
		}
	}
	return nil
}

func empresaCxPAtomicSchemaV1Statements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS empresa_cxp_pagos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			cuenta_por_pagar_id BIGINT NOT NULL,
			movimiento_finanzas_id BIGINT NOT NULL,
			monto NUMERIC(18,2) NOT NULL,
			fecha_aplicacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			metodo_pago TEXT NOT NULL DEFAULT 'efectivo',
			referencia_externa TEXT NOT NULL DEFAULT '',
			concepto TEXT NOT NULL DEFAULT '',
			observaciones TEXT NOT NULL DEFAULT '',
			usuario_creador TEXT NOT NULL DEFAULT 'sistema',
			idempotency_key_hash TEXT NOT NULL,
			fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (empresa_id, idempotency_key_hash),
			CHECK (monto > 0)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_cxp_pagos_cuenta_fecha
				ON empresa_cxp_pagos (empresa_id, cuenta_por_pagar_id, fecha_aplicacion DESC, id DESC)`,
	}
}

func empresaCxPAtomicSchemaStatements() []string {
	statements := append([]string{}, empresaCxPAtomicSchemaV1Statements()...)
	return append(statements, `ALTER TABLE empresa_cxp_pagos ADD COLUMN IF NOT EXISTS request_hash TEXT NOT NULL DEFAULT ''`)
}

// RegistrarEmpresaCxPAbono rejects overpayments rather than silently reducing
// them. A retry with the same key returns the original committed result.
func RegistrarEmpresaCxPAbono(dbConn *sql.DB, input EmpresaCxPAbonoInput) (EmpresaCxPAbonoResult, error) {
	return RegistrarEmpresaCxPAbonoContext(context.Background(), dbConn, input)
}

// RegistrarEmpresaCxPAbonoContext aplica el pago canónico conservando la
// cancelación desde el handler hasta cada operación de la transacción.
func RegistrarEmpresaCxPAbonoContext(ctx context.Context, dbConn *sql.DB, input EmpresaCxPAbonoInput) (EmpresaCxPAbonoResult, error) {
	var result EmpresaCxPAbonoResult
	if dbConn == nil || input.EmpresaID <= 0 || input.CuentaPorPagarID <= 0 {
		return result, fmt.Errorf("empresa_id e id de cuenta por pagar son obligatorios")
	}
	if len(strings.TrimSpace(input.IdempotencyKey)) == 0 {
		return result, ErrEmpresaCxPIdempotencyKeyRequired
	}
	if len(strings.TrimSpace(input.IdempotencyKey)) > 512 {
		return result, fmt.Errorf("la clave de idempotencia supera el limite permitido")
	}
	input.Monto = roundReportesMoney(input.Monto)
	if math.IsNaN(input.Monto) || math.IsInf(input.Monto, 0) || input.Monto <= 0 {
		return result, fmt.Errorf("monto del abono debe ser mayor que cero")
	}
	input.PeriodoContable = normalizePeriodoContable(input.PeriodoContable)
	if input.PeriodoContable == "" {
		input.PeriodoContable = time.Now().In(time.Local).Format("2006-01")
	}
	var metodoErr error
	input.MetodoPago, metodoErr = normalizeEmpresaCxPMetodoPago(input.MetodoPago)
	if metodoErr != nil {
		return result, metodoErr
	}
	input.Usuario = strings.TrimSpace(input.Usuario)
	if input.Usuario == "" {
		input.Usuario = "sistema"
	}

	keyHash := empresaCxPIdempotencyHash(input.IdempotencyKey)
	requestPayload, _ := json.Marshal(map[string]interface{}{
		"empresa_id":          input.EmpresaID,
		"cuenta_por_pagar_id": input.CuentaPorPagarID,
		"monto":               input.Monto,
		"periodo_contable":    input.PeriodoContable,
		"metodo_pago":         input.MetodoPago,
		"referencia_externa":  strings.TrimSpace(input.ReferenciaExterna),
		"concepto":            strings.TrimSpace(input.Concepto),
		"observaciones":       strings.TrimSpace(input.Observaciones),
		"usuario":             input.Usuario,
	})
	requestHash := empresaCxPIdempotencyHash(string(requestPayload))
	tx, err := dbConn.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}
	defer func() { _ = tx.Rollback() }()

	var existing EmpresaCxPAbonoResult
	var existingRequestHash string
	err = queryRowTxSQLCompatContext(ctx, tx, `SELECT cuenta_por_pagar_id, id, movimiento_finanzas_id, monto, COALESCE(request_hash, '')
		FROM empresa_cxp_pagos WHERE empresa_id = ? AND idempotency_key_hash = ?`, input.EmpresaID, keyHash).
		Scan(&existing.CuentaPorPagarID, &existing.PagoID, &existing.MovimientoFinanzasID, &existing.MontoAplicado, &existingRequestHash)
	if err == nil {
		if existing.CuentaPorPagarID != input.CuentaPorPagarID || roundReportesMoney(existing.MontoAplicado) != input.Monto || (existingRequestHash != "" && existingRequestHash != requestHash) {
			return result, ErrEmpresaCxPIdempotencyConflict
		}
		var saldo float64
		var estado string
		if err := queryRowTxSQLCompatContext(ctx, tx, `SELECT COALESCE(saldo, 0), COALESCE(estado_cartera, 'pendiente')
			FROM empresa_cuentas_por_pagar WHERE empresa_id = ? AND id = ?`, input.EmpresaID, input.CuentaPorPagarID).
			Scan(&saldo, &estado); err != nil {
			return result, err
		}
		existing.EmpresaID = input.EmpresaID
		existing.SaldoNuevo = roundReportesMoney(saldo)
		existing.EstadoCartera = estado
		existing.IdempotentReplay = true
		if err := tx.Commit(); err != nil {
			return result, err
		}
		return existing, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return result, err
	}

	var codigo, proveedorNombre, documentoCodigo, moneda, fechaVencimiento string
	var valorOriginal, valorPagado, saldoAnterior float64
	err = queryRowTxSQLCompatContext(ctx, tx, `SELECT COALESCE(codigo, ''), COALESCE(proveedor_nombre, ''),
		COALESCE(documento_codigo, ''), COALESCE(moneda, 'COP'), COALESCE(fecha_vencimiento, ''),
		COALESCE(valor_original, 0), COALESCE(valor_pagado, 0), COALESCE(saldo, 0)
		FROM empresa_cuentas_por_pagar WHERE empresa_id = ? AND id = ? FOR UPDATE`, input.EmpresaID, input.CuentaPorPagarID).
		Scan(&codigo, &proveedorNombre, &documentoCodigo, &moneda, &fechaVencimiento, &valorOriginal, &valorPagado, &saldoAnterior)
	if err != nil {
		return result, err
	}
	// A concurrent request with the same idempotency key can have passed the
	// optimistic lookup above while waiting for this account lock. Recheck after
	// FOR UPDATE so it returns the committed application instead of reaching the
	// unique constraint after creating a duplicate financial movement attempt.
	var concurrentReplay EmpresaCxPAbonoResult
	var concurrentRequestHash string
	err = queryRowTxSQLCompatContext(ctx, tx, `SELECT cuenta_por_pagar_id, id, movimiento_finanzas_id, monto, COALESCE(request_hash, '')
		FROM empresa_cxp_pagos WHERE empresa_id = ? AND idempotency_key_hash = ?`, input.EmpresaID, keyHash).
		Scan(&concurrentReplay.CuentaPorPagarID, &concurrentReplay.PagoID, &concurrentReplay.MovimientoFinanzasID, &concurrentReplay.MontoAplicado, &concurrentRequestHash)
	if err == nil {
		if concurrentReplay.CuentaPorPagarID != input.CuentaPorPagarID || roundReportesMoney(concurrentReplay.MontoAplicado) != input.Monto || (concurrentRequestHash != "" && concurrentRequestHash != requestHash) {
			return result, ErrEmpresaCxPIdempotencyConflict
		}
		concurrentReplay.EmpresaID = input.EmpresaID
		concurrentReplay.SaldoNuevo = roundReportesMoney(saldoAnterior)
		concurrentReplay.EstadoCartera = empresaCxPEstado(saldoAnterior, fechaVencimiento)
		concurrentReplay.IdempotentReplay = true
		if err := tx.Commit(); err != nil {
			return result, err
		}
		return concurrentReplay, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return result, err
	}
	saldoAnterior = roundReportesMoney(saldoAnterior)
	if saldoAnterior <= 0 {
		return result, ErrEmpresaCxPNoPendingBalance
	}
	if input.Monto > saldoAnterior {
		return result, ErrEmpresaCxPAmountExceedsBalance
	}

	var periodoEstado string
	err = queryRowTxSQLCompatContext(ctx, tx, `SELECT COALESCE(estado, 'abierto') FROM empresa_finanzas_periodos
		WHERE empresa_id = ? AND periodo = ? FOR UPDATE`, input.EmpresaID, input.PeriodoContable).Scan(&periodoEstado)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return result, err
	}
	if normalizeEstadoPeriodo(periodoEstado) == "cerrado" {
		return result, ErrPeriodoFinancieroCerrado
	}

	if strings.TrimSpace(input.ReferenciaExterna) == "" {
		input.ReferenciaExterna = firstNonEmpty(documentoCodigo, codigo)
	}
	if strings.TrimSpace(input.Concepto) == "" {
		input.Concepto = "Abono cuenta por pagar " + codigo
	}
	movement := EmpresaFinanzasMovimiento{
		EmpresaID:         input.EmpresaID,
		TipoMovimiento:    "egreso",
		Codigo:            fmt.Sprintf("CXP-%d-%s", input.CuentaPorPagarID, keyHash[:12]),
		PeriodoContable:   input.PeriodoContable,
		FechaMovimiento:   time.Now().In(time.Local).Format("2006-01-02 15:04:05"),
		Categoria:         "cuentas_pagar",
		Subcategoria:      "abono_cartera",
		Concepto:          input.Concepto,
		Descripcion:       "Abono aplicado a cuenta por pagar " + codigo,
		MetodoPago:        input.MetodoPago,
		Moneda:            moneda,
		Monto:             input.Monto,
		Total:             input.Monto,
		TotalNeto:         input.Monto,
		TerceroNombre:     proveedorNombre,
		TipoComprobante:   "recibo_interno",
		NumeroComprobante: input.ReferenciaExterna,
		ReferenciaExterna: input.ReferenciaExterna,
		UsuarioCreador:    input.Usuario,
		Estado:            "activo",
		Observaciones:     input.Observaciones,
	}
	movement, err = normalizeEmpresaFinanzasMovimiento(dbConn, movement, true)
	if err != nil {
		return result, err
	}
	movementID, err := insertEmpresaFinanzasMovimientoTx(ctx, tx, movement)
	if err != nil {
		return result, err
	}

	saldoNuevo := roundReportesMoney(saldoAnterior - input.Monto)
	valorPagadoNuevo := roundReportesMoney(valorPagado + input.Monto)
	estadoNuevo := empresaCxPEstado(saldoNuevo, fechaVencimiento)
	if _, err := execTxSQLCompatContext(ctx, tx, `UPDATE empresa_cuentas_por_pagar SET valor_pagado = ?, saldo = ?,
		estado_cartera = ?, dias_mora = ?, periodo_contable = ?, fecha_ultimo_pago = CURRENT_TIMESTAMP,
		conciliado_en = CURRENT_TIMESTAMP, conciliado_por = ?, fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE empresa_id = ? AND id = ?`, valorPagadoNuevo, saldoNuevo, estadoNuevo,
		empresaCxPDiasMora(fechaVencimiento, saldoNuevo), input.PeriodoContable, input.Usuario, input.EmpresaID, input.CuentaPorPagarID); err != nil {
		return result, err
	}
	pagoID, err := insertTxSQLCompatContext(ctx, tx, `INSERT INTO empresa_cxp_pagos
		(empresa_id, cuenta_por_pagar_id, movimiento_finanzas_id, monto, metodo_pago, referencia_externa, concepto, observaciones, usuario_creador, idempotency_key_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, input.EmpresaID, input.CuentaPorPagarID, movementID, input.Monto,
		input.MetodoPago, input.ReferenciaExterna, input.Concepto, input.Observaciones, input.Usuario, keyHash)
	if err != nil {
		return result, err
	}
	payload, err := json.Marshal(map[string]interface{}{"cuenta_por_pagar_id": input.CuentaPorPagarID, "pago_id": pagoID, "movimiento_finanzas_id": movementID, "monto": input.Monto})
	if err != nil {
		return result, fmt.Errorf("no se pudo serializar el evento outbox CxP: %w", err)
	}
	if err := InsertOutboxEvent(tx, OutboxEvent{EmpresaID: input.EmpresaID, Topic: EmpresaCxPPaymentOutboxTopic, PayloadJSON: string(payload), IdempotencyKey: "cxp-outbox:" + keyHash}); err != nil {
		return result, err
	}
	if err := tx.Commit(); err != nil {
		return result, err
	}
	return EmpresaCxPAbonoResult{EmpresaID: input.EmpresaID, CuentaPorPagarID: input.CuentaPorPagarID, PagoID: pagoID, MovimientoFinanzasID: movementID, MontoAplicado: input.Monto, SaldoAnterior: saldoAnterior, SaldoNuevo: saldoNuevo, EstadoCartera: estadoNuevo}, nil
}

// ProcessEmpresaCxPPaymentAccounting converts the committed CxP outbox event
// into one accounting event. The payment row is locked before checking the
// natural event key so a worker retry cannot create a duplicate journal entry.
func ProcessEmpresaCxPPaymentAccounting(ctx context.Context, dbConn *sql.DB, empresaID int64, payloadJSON string) (EmpresaCxPPaymentAccountingResult, error) {
	var result EmpresaCxPPaymentAccountingResult
	if dbConn == nil {
		return result, fmt.Errorf("database not available")
	}
	if empresaID <= 0 {
		return result, fmt.Errorf("empresa_id es obligatorio")
	}
	var payload struct {
		CuentaPorPagarID     int64   `json:"cuenta_por_pagar_id"`
		PagoID               int64   `json:"pago_id"`
		MovimientoFinanzasID int64   `json:"movimiento_finanzas_id"`
		Monto                float64 `json:"monto"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(payloadJSON)), &payload); err != nil {
		return result, fmt.Errorf("payload CxP invalido")
	}
	if payload.CuentaPorPagarID <= 0 || payload.PagoID <= 0 || payload.MovimientoFinanzasID <= 0 ||
		payload.Monto <= 0 || math.IsNaN(payload.Monto) || math.IsInf(payload.Monto, 0) {
		return result, fmt.Errorf("payload CxP incompleto")
	}

	tx, err := dbConn.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}
	defer func() { _ = tx.Rollback() }()

	var (
		paymentAmount, movementTotal float64
		metodoPago, referencia       string
		codigoCxP, documentoCodigo   string
		periodoContable, moneda      string
		fechaMovimiento              string
	)
	err = queryRowTxSQLCompatContext(ctx, tx, `SELECT
		COALESCE(p.monto, 0), COALESCE(p.metodo_pago, ''), COALESCE(p.referencia_externa, ''),
		COALESCE(c.codigo, ''), COALESCE(c.documento_codigo, ''),
		COALESCE(m.periodo_contable, ''), COALESCE(m.moneda, 'COP'),
		COALESCE(m.fecha_movimiento, ''), COALESCE(NULLIF(m.total_neto, 0), NULLIF(m.total, 0), m.monto, 0)
		FROM empresa_cxp_pagos p
		JOIN empresa_cuentas_por_pagar c
		  ON c.empresa_id = p.empresa_id AND c.id = p.cuenta_por_pagar_id
		JOIN empresa_finanzas_movimientos m
		  ON m.empresa_id = p.empresa_id AND m.id = p.movimiento_finanzas_id
		WHERE p.empresa_id = ? AND p.id = ? AND p.cuenta_por_pagar_id = ?
		  AND p.movimiento_finanzas_id = ? AND COALESCE(m.estado, 'activo') = 'activo'
		FOR UPDATE OF p`,
		empresaID, payload.PagoID, payload.CuentaPorPagarID, payload.MovimientoFinanzasID).
		Scan(&paymentAmount, &metodoPago, &referencia, &codigoCxP, &documentoCodigo,
			&periodoContable, &moneda, &fechaMovimiento, &movementTotal)
	if err != nil {
		return result, err
	}
	if roundReportesMoney(paymentAmount) != roundReportesMoney(payload.Monto) ||
		roundReportesMoney(movementTotal) != roundReportesMoney(payload.Monto) {
		return result, fmt.Errorf("pago CxP y movimiento financiero no concilian")
	}
	periodoContable = normalizePeriodoEventoContable(periodoContable)
	if periodoContable == "" {
		periodoContable = normalizePeriodoEventoContable(fechaMovimiento)
	}
	if periodoContable == "" {
		periodoContable = time.Now().In(time.Local).Format("2006-01")
	}
	moneda = strings.ToUpper(strings.TrimSpace(moneda))
	if moneda == "" {
		moneda = "COP"
	}
	if strings.TrimSpace(fechaMovimiento) == "" {
		fechaMovimiento = time.Now().In(time.Local).Format("2006-01-02 15:04:05")
	}

	result = EmpresaCxPPaymentAccountingResult{EmpresaID: empresaID, PagoID: payload.PagoID}
	err = queryRowTxSQLCompatContext(ctx, tx, `SELECT id
		FROM empresa_eventos_contables
		WHERE empresa_id = ? AND modulo = 'finanzas' AND evento = 'abono_proveedor_registrado'
		  AND entidad = 'empresa_cxp_pagos' AND entidad_id = ?
		  AND COALESCE(estado, 'activo') <> 'anulado'
		ORDER BY id ASC LIMIT 1`, empresaID, payload.PagoID).Scan(&result.EventoContableID)
	if err == nil {
		result.IdempotentReplay = true
		if err := tx.Commit(); err != nil {
			return EmpresaCxPPaymentAccountingResult{}, err
		}
		return result, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return EmpresaCxPPaymentAccountingResult{}, err
	}

	accountingPayload, err := json.Marshal(map[string]interface{}{
		"cuenta_por_pagar_id":    payload.CuentaPorPagarID,
		"pago_id":                payload.PagoID,
		"movimiento_finanzas_id": payload.MovimientoFinanzasID,
		"monto":                  roundReportesMoney(paymentAmount),
		"total_neto":             roundReportesMoney(paymentAmount),
		"tipo_movimiento":        "egreso",
		"metodo_pago":            metodoPago,
		"referencia_externa":     referencia,
		"cuenta_cxp":             "220505",
		"cuenta_caja":            "110505",
	})
	if err != nil {
		return EmpresaCxPPaymentAccountingResult{}, err
	}
	result.EventoContableID, err = insertTxSQLCompatContext(ctx, tx, `INSERT INTO empresa_eventos_contables (
		empresa_id, modulo, evento, entidad, entidad_id, documento_tipo, documento_codigo,
		periodo_contable, monto_total, moneda, payload_json, origen, fecha_evento,
		procesado, fecha_procesado, fecha_creacion, fecha_actualizacion,
		usuario_creador, estado, observaciones
	) VALUES (?, 'finanzas', 'abono_proveedor_registrado', 'empresa_cxp_pagos', ?, 'pago_cxp', ?,
		?, ?, ?, ?, 'pcs-worker.cuentas_por_pagar.pago_registrado', ?,
		0, NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP,
		'pcs-worker', 'activo', ?)`,
		empresaID, payload.PagoID, firstNonEmpty(documentoCodigo, codigoCxP),
		periodoContable, roundReportesMoney(paymentAmount), moneda, string(accountingPayload),
		fechaMovimiento, "Evento contable idempotente del pago CxP "+codigoCxP)
	if err != nil {
		return EmpresaCxPPaymentAccountingResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return EmpresaCxPPaymentAccountingResult{}, err
	}
	return result, nil
}

func empresaCxPIdempotencyHash(key string) string {
	return hashOutboxKey("cxp-abono:" + strings.TrimSpace(key))
}

func normalizeEmpresaCxPMetodoPago(raw string) (string, error) {
	metodo := strings.ToLower(strings.TrimSpace(raw))
	if metodo == "" {
		return "efectivo", nil
	}
	if len(metodo) > 80 || strings.ContainsAny(metodo, "\r\n\t") {
		return "", fmt.Errorf("metodo de pago invalido")
	}
	return metodo, nil
}

func insertEmpresaFinanzasMovimientoTx(ctx context.Context, tx *sql.Tx, m EmpresaFinanzasMovimiento) (int64, error) {
	return insertTxSQLCompatContext(ctx, tx, `INSERT INTO empresa_finanzas_movimientos (
		empresa_id, tipo_movimiento, codigo, fecha_movimiento, periodo_contable, categoria, subcategoria, concepto, descripcion,
		metodo_pago, moneda, monto, impuesto, retencion_fuente, retencion_ica, retencion_iva, total_retenciones, total, total_neto,
		tercero_nombre, tercero_documento, tipo_comprobante, numero_comprobante, comprobante_url, referencia_externa,
		cierre_caja_id, caja_codigo, caja_turno, caja_sucursal_id, aprobado_por, usuario_creador, estado, observaciones,
		fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		m.EmpresaID, m.TipoMovimiento, m.Codigo, m.FechaMovimiento, m.PeriodoContable, m.Categoria, m.Subcategoria, m.Concepto, m.Descripcion,
		m.MetodoPago, m.Moneda, m.Monto, m.Impuesto, m.RetencionFuente, m.RetencionICA, m.RetencionIVA, m.TotalRetenciones, m.Total, m.TotalNeto,
		m.TerceroNombre, m.TerceroDocumento, m.TipoComprobante, m.NumeroComprobante, m.ComprobanteURL, m.ReferenciaExterna,
		m.CierreCajaID, m.CajaCodigo, m.CajaTurno, m.CajaSucursalID, m.AprobadoPor, m.UsuarioCreador, m.Estado, m.Observaciones)
}

func empresaCxPEstado(saldo float64, fechaVencimiento string) string {
	if saldo <= 0 {
		return "pagada"
	}
	if fecha, err := time.Parse("2006-01-02", strings.TrimSpace(fechaVencimiento)); err == nil && fecha.Before(time.Now().In(time.Local).Truncate(24*time.Hour)) {
		return "vencida"
	}
	return "parcial"
}

func empresaCxPDiasMora(fechaVencimiento string, saldo float64) int {
	if saldo <= 0 {
		return 0
	}
	fecha, err := time.Parse("2006-01-02", strings.TrimSpace(fechaVencimiento))
	if err != nil {
		return 0
	}
	days := int(time.Since(fecha).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}
