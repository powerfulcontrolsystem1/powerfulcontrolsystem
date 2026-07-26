package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

const empresaCxPAtomicSchemaFingerprint = "empresa_cxp_pagos:v1:tenant-lock-idempotency-finance-outbox"

var (
	ErrEmpresaCxPIdempotencyKeyRequired = errors.New("la clave de idempotencia es obligatoria")
	ErrEmpresaCxPAmountExceedsBalance   = errors.New("el monto supera el saldo pendiente de la cuenta por pagar")
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

func applyEmpresaCxPAtomicSchemaTx(tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	for _, statement := range empresaCxPAtomicSchemaStatements() {
		if _, err := execTxSQLCompat(tx, statement); err != nil {
			return err
		}
	}
	return nil
}

func empresaCxPAtomicSchemaStatements() []string {
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

// RegistrarEmpresaCxPAbono rejects overpayments rather than silently reducing
// them. A retry with the same key returns the original committed result.
func RegistrarEmpresaCxPAbono(dbConn *sql.DB, input EmpresaCxPAbonoInput) (EmpresaCxPAbonoResult, error) {
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
	tx, err := dbConn.Begin()
	if err != nil {
		return result, err
	}
	defer func() { _ = tx.Rollback() }()

	var existing EmpresaCxPAbonoResult
	err = queryRowTxSQLCompat(tx, `SELECT cuenta_por_pagar_id, id, movimiento_finanzas_id, monto
		FROM empresa_cxp_pagos WHERE empresa_id = ? AND idempotency_key_hash = ?`, input.EmpresaID, keyHash).
		Scan(&existing.CuentaPorPagarID, &existing.PagoID, &existing.MovimientoFinanzasID, &existing.MontoAplicado)
	if err == nil {
		if existing.CuentaPorPagarID != input.CuentaPorPagarID {
			return result, fmt.Errorf("la clave de idempotencia ya fue usada para otra cuenta por pagar")
		}
		var saldo float64
		var estado string
		if err := queryRowTxSQLCompat(tx, `SELECT COALESCE(saldo, 0), COALESCE(estado_cartera, 'pendiente')
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
	err = queryRowTxSQLCompat(tx, `SELECT COALESCE(codigo, ''), COALESCE(proveedor_nombre, ''),
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
	err = queryRowTxSQLCompat(tx, `SELECT cuenta_por_pagar_id, id, movimiento_finanzas_id, monto
		FROM empresa_cxp_pagos WHERE empresa_id = ? AND idempotency_key_hash = ?`, input.EmpresaID, keyHash).
		Scan(&concurrentReplay.CuentaPorPagarID, &concurrentReplay.PagoID, &concurrentReplay.MovimientoFinanzasID, &concurrentReplay.MontoAplicado)
	if err == nil {
		if concurrentReplay.CuentaPorPagarID != input.CuentaPorPagarID {
			return result, fmt.Errorf("la clave de idempotencia ya fue usada para otra cuenta por pagar")
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
		return result, fmt.Errorf("la cuenta por pagar no tiene saldo pendiente")
	}
	if input.Monto > saldoAnterior {
		return result, ErrEmpresaCxPAmountExceedsBalance
	}

	var periodoEstado string
	err = queryRowTxSQLCompat(tx, `SELECT COALESCE(estado, 'abierto') FROM empresa_finanzas_periodos
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
	movementID, err := insertEmpresaFinanzasMovimientoTx(tx, movement)
	if err != nil {
		return result, err
	}

	saldoNuevo := roundReportesMoney(saldoAnterior - input.Monto)
	valorPagadoNuevo := roundReportesMoney(valorPagado + input.Monto)
	estadoNuevo := empresaCxPEstado(saldoNuevo, fechaVencimiento)
	if _, err := execTxSQLCompat(tx, `UPDATE empresa_cuentas_por_pagar SET valor_pagado = ?, saldo = ?,
		estado_cartera = ?, dias_mora = ?, periodo_contable = ?, fecha_ultimo_pago = CURRENT_TIMESTAMP,
		conciliado_en = CURRENT_TIMESTAMP, conciliado_por = ?, fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE empresa_id = ? AND id = ?`, valorPagadoNuevo, saldoNuevo, estadoNuevo,
		empresaCxPDiasMora(fechaVencimiento, saldoNuevo), input.PeriodoContable, input.Usuario, input.EmpresaID, input.CuentaPorPagarID); err != nil {
		return result, err
	}
	pagoID, err := insertTxSQLCompat(tx, `INSERT INTO empresa_cxp_pagos
		(empresa_id, cuenta_por_pagar_id, movimiento_finanzas_id, monto, metodo_pago, referencia_externa, concepto, observaciones, usuario_creador, idempotency_key_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, input.EmpresaID, input.CuentaPorPagarID, movementID, input.Monto,
		input.MetodoPago, input.ReferenciaExterna, input.Concepto, input.Observaciones, input.Usuario, keyHash)
	if err != nil {
		return result, err
	}
	payload, _ := json.Marshal(map[string]interface{}{"cuenta_por_pagar_id": input.CuentaPorPagarID, "pago_id": pagoID, "movimiento_finanzas_id": movementID, "monto": input.Monto})
	if err := InsertOutboxEvent(tx, OutboxEvent{EmpresaID: input.EmpresaID, Topic: "cuentas_por_pagar.pago_registrado", PayloadJSON: string(payload), IdempotencyKey: "cxp-outbox:" + keyHash}); err != nil {
		return result, err
	}
	if err := tx.Commit(); err != nil {
		return result, err
	}
	return EmpresaCxPAbonoResult{EmpresaID: input.EmpresaID, CuentaPorPagarID: input.CuentaPorPagarID, PagoID: pagoID, MovimientoFinanzasID: movementID, MontoAplicado: input.Monto, SaldoAnterior: saldoAnterior, SaldoNuevo: saldoNuevo, EstadoCartera: estadoNuevo}, nil
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

func insertEmpresaFinanzasMovimientoTx(tx *sql.Tx, m EmpresaFinanzasMovimiento) (int64, error) {
	return insertTxSQLCompat(tx, `INSERT INTO empresa_finanzas_movimientos (
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
