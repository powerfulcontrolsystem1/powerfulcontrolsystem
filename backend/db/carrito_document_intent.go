package db

import (
	"database/sql"
	"fmt"
	"strings"
)

// CarritoPaidDocumentIntent freezes the fiscal document decision inside the
// same transaction that closes the cart. The worker can therefore recover the
// document without recalculating a mutable per-company frequency counter.
type CarritoPaidDocumentIntent struct {
	RequestedMode           string `json:"requested_mode,omitempty"`
	ResolvedMode            string `json:"resolved_mode"`
	AutomaticInvoiceEnabled bool   `json:"automatic_invoice_enabled"`
	FrequencyApplied        bool   `json:"frequency_applied"`
	FrequencyEveryNReceipts int64  `json:"frequency_every_n_receipts,omitempty"`
	FrequencyCounterBefore  int64  `json:"frequency_counter_before,omitempty"`
	FrequencyCounterAfter   int64  `json:"frequency_counter_after,omitempty"`
}

func normalizeCarritoPaidRequestedMode(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "factura", "factura_electronica", "venta_con_factura", "venta_factura_electronica":
		return "factura_electronica"
	case "venta_sola", "comprobante", "comprobante_pago", "sin_factura", "venta_simple":
		return "comprobante_pago"
	default:
		return ""
	}
}

func resolveCarritoPaidDocumentIntentTx(tx *sql.Tx, empresaID int64, requestedMode string) (*CarritoPaidDocumentIntent, error) {
	if tx == nil || empresaID <= 0 {
		return nil, fmt.Errorf("empresa o transaccion invalida para resolver documento de venta")
	}
	requested := normalizeCarritoPaidRequestedMode(requestedMode)
	intent := &CarritoPaidDocumentIntent{
		RequestedMode: requested,
		ResolvedMode:  "comprobante_pago",
	}
	if requested != "" {
		intent.ResolvedMode = requested
		return intent, nil
	}

	var configuredMode string
	var frequencyEnabled int
	var everyN int64
	var counter int64
	err := queryRowTxSQLCompat(tx, `SELECT
		COALESCE(modo_documento_venta, 'comprobante_pago'),
		COALESCE(facturacion_frecuencia_automatica_activa, 0),
		COALESCE(facturacion_frecuencia_cada_n_no, 0),
		COALESCE(facturacion_frecuencia_contador, 0)
	FROM empresa_configuracion_avanzada
	WHERE empresa_id = ?
	FOR UPDATE`, empresaID).Scan(&configuredMode, &frequencyEnabled, &everyN, &counter)
	if err != nil {
		if err == sql.ErrNoRows {
			return intent, nil
		}
		return nil, fmt.Errorf("bloquear configuracion de documento de venta: %w", err)
	}

	intent.AutomaticInvoiceEnabled = normalizeCarritoPaidRequestedMode(configuredMode) == "factura_electronica"
	if !intent.AutomaticInvoiceEnabled {
		return intent, nil
	}
	intent.ResolvedMode = "factura_electronica"
	if frequencyEnabled != 1 || everyN <= 0 {
		return intent, nil
	}

	cycle := everyN + 1
	if cycle <= 1 {
		return nil, fmt.Errorf("frecuencia automatica de facturacion invalida")
	}
	counter %= cycle
	if counter < 0 {
		counter = 0
	}
	next := (counter + 1) % cycle
	intent.FrequencyApplied = true
	intent.FrequencyEveryNReceipts = everyN
	intent.FrequencyCounterBefore = counter
	intent.FrequencyCounterAfter = next
	if counter != 0 {
		intent.ResolvedMode = "comprobante_pago"
	}
	result, err := execTxSQLCompat(tx, `UPDATE empresa_configuracion_avanzada
		SET facturacion_frecuencia_contador = ?,
			fecha_actualizacion = `+sqlNowExpr()+`
		WHERE empresa_id = ?`, next, empresaID)
	if err != nil {
		return nil, fmt.Errorf("avanzar frecuencia automatica de facturacion: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return nil, fmt.Errorf("avanzar frecuencia automatica de facturacion: %w", sql.ErrNoRows)
	}
	return intent, nil
}
