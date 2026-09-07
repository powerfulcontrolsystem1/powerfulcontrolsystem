// Package ventas owns durable sale use cases that do not belong to HTTP or to
// the worker transport. This keeps business recovery independently testable.
package ventas

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

type SalePaidEvent struct {
	EmpresaID   int64   `json:"empresa_id"`
	CarritoID   int64   `json:"carrito_id"`
	TotalPagado float64 `json:"total_pagado"`
	MetodoPago  string  `json:"metodo_pago"`
	Usuario     string  `json:"usuario"`
}

type Service struct {
	DB *sql.DB
}

// RecoverPaidSaleAccounting guarantees that a paid cart eventually owns one
// accounting event even if the non-blocking HTTP projection was interrupted.
func (s Service) RecoverPaidSaleAccounting(ctx context.Context, job dbpkg.AsyncJob) error {
	if s.DB == nil {
		return fmt.Errorf("ventas database is required")
	}
	var payload SalePaidEvent
	if err := json.Unmarshal([]byte(job.PayloadJSON), &payload); err != nil {
		return fmt.Errorf("invalid sale-paid payload")
	}
	if payload.EmpresaID <= 0 || payload.EmpresaID != job.EmpresaID || payload.CarritoID <= 0 {
		return fmt.Errorf("invalid sale-paid tenant reference")
	}
	var existing int
	err := dbpkg.QueryRowCompat(s.DB, `SELECT COUNT(*) FROM empresa_eventos_contables
		WHERE empresa_id=? AND modulo='ventas' AND evento='venta_pagada'
		AND entidad='carrito_compra' AND entidad_id=?`, payload.EmpresaID, payload.CarritoID).Scan(&existing)
	if err != nil || existing > 0 {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	carrito, err := dbpkg.GetCarritoCompraByID(s.DB, payload.EmpresaID, payload.CarritoID)
	if err != nil {
		return err
	}
	if carrito == nil || strings.TrimSpace(carrito.PagadoEn) == "" {
		return fmt.Errorf("sale-paid cart is not paid")
	}
	accountingPayload, _ := json.Marshal(map[string]interface{}{
		"origen": "transactional_outbox", "metodo_pago": payload.MetodoPago, "total_neto": payload.TotalPagado,
	})
	_, err = dbpkg.CreateEmpresaEventoContable(s.DB, dbpkg.EmpresaEventoContable{
		EmpresaID: payload.EmpresaID, Modulo: "ventas", Evento: "venta_pagada",
		Entidad: "carrito_compra", EntidadID: payload.CarritoID, DocumentoTipo: "carrito",
		DocumentoCodigo: carrito.Codigo, MontoTotal: payload.TotalPagado, Moneda: carrito.Moneda,
		PayloadJSON: string(accountingPayload), Origen: "pcs-worker", UsuarioCreador: payload.Usuario,
		Estado: "activo", Observaciones: "evento recuperado desde outbox transaccional",
	})
	return err
}
