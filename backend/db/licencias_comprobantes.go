package db

import (
	"database/sql"
	"fmt"
	"strings"
)

// EmpresaLicenciaPagoResumen contiene solo los metadatos necesarios para que una
// empresa consulte sus compras de licencia. Deliberadamente no expone payloads
// de pasarela, descuentos ni datos del medio de pago.
type EmpresaLicenciaPagoResumen struct {
	ID             int64  `json:"id"`
	Proveedor      string `json:"proveedor"`
	LicenciaID     int64  `json:"licencia_id"`
	LicenciaNombre string `json:"licencia_nombre"`
	Referencia     string `json:"referencia"`
	TransaccionID  string `json:"transaccion_id"`
	Estado         string `json:"estado"`
	FechaCreacion  string `json:"fecha_creacion"`
}

// ListEmpresaLicenciaPagos devuelve el historial de pasarelas asociado a una
// empresa. Cada rama del UNION filtra empresa_id antes de combinar resultados.
func ListEmpresaLicenciaPagos(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaLicenciaPagoResumen, error) {
	if dbConn == nil || empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id valido requerido")
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if err := EnsurePaymentGatewaySchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT id, proveedor, licencia_id, licencia_nombre, referencia, transaccion_id, estado, fecha_creacion
		FROM (
			SELECT p.id, 'wompi' AS proveedor, COALESCE(p.licencia_id, 0) AS licencia_id,
				COALESCE(NULLIF(l.nombre, ''), 'Licencia del sistema') AS licencia_nombre,
				COALESCE(p.reference, '') AS referencia, COALESCE(p.transaction_id, '') AS transaccion_id,
				COALESCE(p.status, '') AS estado, COALESCE(p.fecha_creacion, '') AS fecha_creacion
			FROM pagos_wompi p
			LEFT JOIN licencias l ON l.id = p.licencia_id
			WHERE p.empresa_id = ?
			UNION ALL
			SELECT p.id, 'epayco' AS proveedor, COALESCE(p.licencia_id, 0) AS licencia_id,
				COALESCE(NULLIF(l.nombre, ''), 'Licencia del sistema') AS licencia_nombre,
				COALESCE(p.reference, '') AS referencia, COALESCE(p.transaction_id, '') AS transaccion_id,
				COALESCE(p.status, '') AS estado, COALESCE(p.fecha_creacion, '') AS fecha_creacion
			FROM pagos_epayco p
			LEFT JOIN licencias l ON l.id = p.licencia_id
			WHERE p.empresa_id = ?
		) pagos
		ORDER BY fecha_creacion DESC, id DESC
		LIMIT ?`, empresaID, empresaID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]EmpresaLicenciaPagoResumen, 0)
	for rows.Next() {
		var item EmpresaLicenciaPagoResumen
		if err := rows.Scan(&item.ID, &item.Proveedor, &item.LicenciaID, &item.LicenciaNombre, &item.Referencia, &item.TransaccionID, &item.Estado, &item.FechaCreacion); err != nil {
			return nil, err
		}
		item.Proveedor = strings.ToLower(strings.TrimSpace(item.Proveedor))
		items = append(items, item)
	}
	return items, rows.Err()
}

// GetEmpresaLicenciaPago obtiene un pago por su identificador interno dentro
// del alcance de la empresa. El proveedor se convierte primero a una tabla
// cerrada; nunca se usa texto del cliente como identificador SQL.
func GetEmpresaLicenciaPago(dbConn *sql.DB, empresaID int64, proveedor string, paymentID int64) (*EmpresaLicenciaPagoResumen, error) {
	if dbConn == nil || empresaID <= 0 || paymentID <= 0 {
		return nil, fmt.Errorf("pago o empresa invalido")
	}
	table, ok := licenciaPaymentTable(proveedor)
	if !ok {
		return nil, fmt.Errorf("proveedor no permitido")
	}
	if err := EnsurePaymentGatewaySchema(dbConn); err != nil {
		return nil, err
	}
	row := queryRowSQLCompat(dbConn, fmt.Sprintf(`SELECT p.id, ?, COALESCE(p.licencia_id, 0),
		COALESCE(NULLIF(l.nombre, ''), 'Licencia del sistema'), COALESCE(p.reference, ''),
		COALESCE(p.transaction_id, ''), COALESCE(p.status, ''), COALESCE(p.fecha_creacion, '')
		FROM %s p
		LEFT JOIN licencias l ON l.id = p.licencia_id
		WHERE p.id = ? AND p.empresa_id = ?
		LIMIT 1`, table), strings.ToLower(strings.TrimSpace(proveedor)), paymentID, empresaID)
	var item EmpresaLicenciaPagoResumen
	if err := row.Scan(&item.ID, &item.Proveedor, &item.LicenciaID, &item.LicenciaNombre, &item.Referencia, &item.TransaccionID, &item.Estado, &item.FechaCreacion); err != nil {
		return nil, err
	}
	return &item, nil
}

// GetPowerfulSystemEmpresa consulta la empresa emisora configurada sin crearla
// ni modificar licencias/configuracion durante la lectura de comprobantes.
func GetPowerfulSystemEmpresa(dbEmp, dbSuper *sql.DB) (*Empresa, error) {
	if empresa, err := getPowerfulSystemEmpresaFromConfig(dbEmp, dbSuper); err != nil {
		return nil, err
	} else if empresa != nil {
		return empresa, nil
	}
	return findPowerfulSystemEmpresaByName(dbEmp)
}
