package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

type paidSaleDocumentRecoveryPayload struct {
	EmpresaID      int64                            `json:"empresa_id"`
	CarritoID      int64                            `json:"carrito_id"`
	TotalPagado    float64                          `json:"total_pagado"`
	Usuario        string                           `json:"usuario"`
	DocumentIntent *dbpkg.CarritoPaidDocumentIntent `json:"document_intent,omitempty"`
}

// RecoverPaidSaleDocuments repairs the document side of a committed cart
// payment from its durable outbox event. It never recalculates automatic
// frequency and never trusts a tenant different from the job envelope.
func RecoverPaidSaleDocuments(ctx context.Context, dbEmp, dbSuper *sql.DB, job dbpkg.AsyncJob) error {
	if dbEmp == nil || job.EmpresaID <= 0 {
		return fmt.Errorf("evento de venta pagada sin empresa valida")
	}
	var payload paidSaleDocumentRecoveryPayload
	if err := json.Unmarshal([]byte(job.PayloadJSON), &payload); err != nil {
		return fmt.Errorf("decodificar recuperacion documental de venta: %w", err)
	}
	if payload.EmpresaID != job.EmpresaID || payload.CarritoID <= 0 {
		return fmt.Errorf("evento de venta pagada fuera del tenant o sin carrito")
	}

	resolvedMode := "comprobante_pago"
	if payload.DocumentIntent != nil {
		switch strings.ToLower(strings.TrimSpace(payload.DocumentIntent.ResolvedMode)) {
		case "factura_electronica":
			resolvedMode = "factura_electronica"
		case "comprobante_pago":
			resolvedMode = "comprobante_pago"
		default:
			return fmt.Errorf("evento de venta pagada con intencion documental invalida")
		}
	}

	carrito, err := dbpkg.GetCarritoCompraByID(dbEmp, job.EmpresaID, payload.CarritoID)
	if err != nil {
		return fmt.Errorf("cargar carrito pagado para recuperar documento: %w", err)
	}
	if !isCarritoVentaPagada(carrito) {
		return fmt.Errorf("el carrito de la recuperacion documental no esta pagado")
	}
	total := payload.TotalPagado
	if total <= 0 {
		total = carrito.TotalPagado
	}
	if total <= 0 {
		total = carrito.Total
	}
	usuario := strings.TrimSpace(payload.Usuario)
	if usuario == "" {
		usuario = "sistema.pcs-worker"
	}
	_, err = registrarDocumentoVentaDesdeCarritoPagadoContext(ctx, dbEmp, dbSuper, carrito, total, usuario, resolvedMode)
	if err != nil {
		return fmt.Errorf("recuperar documento de venta pagada: %w", err)
	}
	return nil
}
