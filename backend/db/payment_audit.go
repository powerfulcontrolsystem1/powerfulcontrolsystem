package db

import (
	"database/sql"
	"fmt"
	"strings"
)

// SuperPaymentAuditFilters limits the global payment ledger exposed to the
// Super Administrator. Provider and status are allow-listed by the handler.
type SuperPaymentAuditFilters struct {
	Provider  string
	Status    string
	Search    string
	EmpresaID int64
	Limit     int
	Offset    int
}

type SuperPaymentTransaction struct {
	Provider           string `json:"provider"`
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	LicenciaID         int64  `json:"licencia_id"`
	TransactionID      string `json:"transaction_id"`
	Reference          string `json:"reference"`
	Status             string `json:"status"`
	ActivationStatus   string `json:"activation_status"`
	ActivationAttempts int    `json:"activation_attempts"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

type SuperPaymentCheckoutAttempt struct {
	Provider     string `json:"provider"`
	EmpresaID    int64  `json:"empresa_id"`
	Reference    string `json:"reference"`
	Status       string `json:"status"`
	ResponseCode int    `json:"response_code"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type SuperPaymentPostEffect struct {
	Provider        string `json:"provider"`
	PaymentRecordID int64  `json:"payment_record_id"`
	EmpresaID       int64  `json:"empresa_id"`
	LicenciaID      int64  `json:"licencia_id"`
	Effect          string `json:"effect"`
	Status          string `json:"status"`
	Attempts        int    `json:"attempts"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type SuperPaymentAudit struct {
	Transactions     []SuperPaymentTransaction     `json:"transactions"`
	CheckoutAttempts []SuperPaymentCheckoutAttempt `json:"checkout_attempts"`
	PostEffects      []SuperPaymentPostEffect      `json:"post_effects"`
}

func normalizeSuperPaymentAuditFilters(filters SuperPaymentAuditFilters) SuperPaymentAuditFilters {
	filters.Provider = strings.ToLower(strings.TrimSpace(filters.Provider))
	filters.Status = strings.ToUpper(strings.TrimSpace(filters.Status))
	filters.Search = strings.TrimSpace(filters.Search)
	if filters.Limit <= 0 {
		filters.Limit = 50
	}
	if filters.Limit > 200 {
		filters.Limit = 200
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}
	return filters
}

func paymentAuditWhere(filters SuperPaymentAuditFilters, statusColumn string) (string, []interface{}) {
	clauses := []string{"1 = 1"}
	args := make([]interface{}, 0, 8)
	if filters.Provider != "" {
		clauses = append(clauses, "provider = ?")
		args = append(args, filters.Provider)
	}
	if filters.Status != "" {
		clauses = append(clauses, "UPPER(COALESCE("+statusColumn+", '')) = ?")
		args = append(args, filters.Status)
	}
	if filters.EmpresaID > 0 {
		clauses = append(clauses, "empresa_id = ?")
		args = append(args, filters.EmpresaID)
	}
	if filters.Search != "" {
		like := "%" + filters.Search + "%"
		clauses = append(clauses, "(transaction_id ILIKE ? OR reference ILIKE ? OR CAST(empresa_id AS TEXT) = ? OR CAST(licencia_id AS TEXT) = ?)")
		args = append(args, like, like, filters.Search, filters.Search)
	}
	return strings.Join(clauses, " AND "), args
}

// ListSuperPaymentAudit returns sanitized operational metadata only. It never
// selects raw provider payloads, signatures, customer data, idempotency hashes,
// cached responses, or remote error bodies.
func ListSuperPaymentAudit(dbConn *sql.DB, filters SuperPaymentAuditFilters) (*SuperPaymentAudit, error) {
	if dbConn == nil {
		return nil, fmt.Errorf("db connection is nil")
	}
	filters = normalizeSuperPaymentAuditFilters(filters)
	out := &SuperPaymentAudit{
		Transactions: make([]SuperPaymentTransaction, 0), CheckoutAttempts: make([]SuperPaymentCheckoutAttempt, 0), PostEffects: make([]SuperPaymentPostEffect, 0),
	}

	where, args := paymentAuditWhere(filters, "status")
	transactionQuery := `SELECT provider, id, empresa_id, licencia_id, transaction_id, reference, status,
		activation_status, activation_attempts, created_at, updated_at
	FROM (
		SELECT 'wompi' AS provider, id, COALESCE(empresa_id, 0) AS empresa_id,
			COALESCE(licencia_id, 0) AS licencia_id, COALESCE(transaction_id, '') AS transaction_id,
			COALESCE(reference, '') AS reference, COALESCE(status, '') AS status,
			COALESCE(licencia_activation_status, '') AS activation_status,
			COALESCE(licencia_activation_attempts, 0) AS activation_attempts,
			COALESCE(fecha_creacion, '') AS created_at, COALESCE(fecha_actualizacion, '') AS updated_at
		FROM pagos_wompi
		UNION ALL
		SELECT 'epayco', id, COALESCE(empresa_id, 0), COALESCE(licencia_id, 0),
			COALESCE(transaction_id, ''), COALESCE(reference, ''), COALESCE(status, ''),
			COALESCE(licencia_activation_status, ''), COALESCE(licencia_activation_attempts, 0),
			COALESCE(fecha_creacion, ''), COALESCE(fecha_actualizacion, '')
		FROM pagos_epayco
	) payment_rows WHERE ` + where + ` ORDER BY COALESCE(NULLIF(updated_at, ''), created_at) DESC, id DESC LIMIT ? OFFSET ?`
	args = append(args, filters.Limit, filters.Offset)
	rows, err := querySQLCompat(dbConn, transactionQuery, args...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var item SuperPaymentTransaction
		if err := rows.Scan(&item.Provider, &item.ID, &item.EmpresaID, &item.LicenciaID, &item.TransactionID, &item.Reference, &item.Status, &item.ActivationStatus, &item.ActivationAttempts, &item.CreatedAt, &item.UpdatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		out.Transactions = append(out.Transactions, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	checkoutWhere, checkoutArgs := paymentAuditWhere(filters, "status")
	checkoutQuery := `SELECT provider, empresa_id, reference, status, response_code, created_at, updated_at FROM (
		SELECT proveedor AS provider, empresa_id, 0::BIGINT AS licencia_id, ''::TEXT AS transaction_id,
			referencia AS reference, COALESCE(estado, '') AS status, COALESCE(codigo_respuesta, 0) AS response_code,
			fecha_creacion::TEXT AS created_at, fecha_actualizacion::TEXT AS updated_at
		FROM payment_checkout_idempotencia
	) checkout_rows WHERE ` + checkoutWhere + ` ORDER BY updated_at DESC LIMIT ? OFFSET ?`
	checkoutArgs = append(checkoutArgs, filters.Limit, filters.Offset)
	rows, err = querySQLCompat(dbConn, checkoutQuery, checkoutArgs...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var item SuperPaymentCheckoutAttempt
		if err := rows.Scan(&item.Provider, &item.EmpresaID, &item.Reference, &item.Status, &item.ResponseCode, &item.CreatedAt, &item.UpdatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		out.CheckoutAttempts = append(out.CheckoutAttempts, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	effectWhere, effectArgs := paymentAuditWhere(filters, "status")
	effectQuery := `SELECT provider, payment_record_id, empresa_id, licencia_id, effect, status, attempts, created_at, updated_at FROM (
		SELECT e.proveedor AS provider, e.pago_registro_id AS payment_record_id,
			COALESCE(CASE WHEN e.proveedor = 'wompi' THEN w.empresa_id ELSE p.empresa_id END, 0) AS empresa_id,
			COALESCE(CASE WHEN e.proveedor = 'wompi' THEN w.licencia_id ELSE p.licencia_id END, 0) AS licencia_id,
			COALESCE(CASE WHEN e.proveedor = 'wompi' THEN w.transaction_id ELSE p.transaction_id END, '') AS transaction_id,
			COALESCE(CASE WHEN e.proveedor = 'wompi' THEN w.reference ELSE p.reference END, '') AS reference,
			e.efecto AS effect, COALESCE(e.estado, '') AS status, COALESCE(e.intentos, 0) AS attempts,
			e.fecha_creacion::TEXT AS created_at, e.fecha_actualizacion::TEXT AS updated_at
		FROM payment_post_effect_idempotencia e
		LEFT JOIN pagos_wompi w ON e.proveedor = 'wompi' AND w.id = e.pago_registro_id
		LEFT JOIN pagos_epayco p ON e.proveedor = 'epayco' AND p.id = e.pago_registro_id
	) effect_rows WHERE ` + effectWhere + ` ORDER BY updated_at DESC LIMIT ? OFFSET ?`
	effectArgs = append(effectArgs, filters.Limit, filters.Offset)
	rows, err = querySQLCompat(dbConn, effectQuery, effectArgs...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var item SuperPaymentPostEffect
		if err := rows.Scan(&item.Provider, &item.PaymentRecordID, &item.EmpresaID, &item.LicenciaID, &item.Effect, &item.Status, &item.Attempts, &item.CreatedAt, &item.UpdatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		out.PostEffects = append(out.PostEffects, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	return out, nil
}
