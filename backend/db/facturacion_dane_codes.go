package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

const empresaFacturacionDANECodesFingerprint = "empresa-facturacion-dane-codes:v1:emisor-cliente-departamento-municipio"

// applyEmpresaFacturacionDANECodesTx adds the authoritative geographic codes
// required by Colombian electronic-document party addresses. Existing rows are
// intentionally left blank: a migration must never infer a fiscal location.
func applyEmpresaFacturacionDANECodesTx(ctx context.Context, tx *sql.Tx) error {
	statements := []string{
		`ALTER TABLE empresa_configuracion_avanzada ADD COLUMN IF NOT EXISTS departamento_codigo_dane TEXT`,
		`ALTER TABLE empresa_configuracion_avanzada ADD COLUMN IF NOT EXISTS municipio_codigo_dane TEXT`,
		`ALTER TABLE clientes ADD COLUMN IF NOT EXISTS departamento_codigo_dane TEXT`,
		`ALTER TABLE clientes ADD COLUMN IF NOT EXISTS municipio_codigo_dane TEXT`,
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func normalizeFacturacionDANECodes(departamento, municipio string) (string, string, error) {
	departamento = strings.TrimSpace(departamento)
	municipio = strings.TrimSpace(municipio)
	onlyDigits := func(value string) bool {
		if value == "" {
			return true
		}
		for _, char := range value {
			if char < '0' || char > '9' {
				return false
			}
		}
		return true
	}
	if departamento != "" && (len(departamento) != 2 || !onlyDigits(departamento)) {
		return "", "", fmt.Errorf("departamento_codigo_dane debe contener 2 digitos")
	}
	if municipio != "" && (len(municipio) != 5 || !onlyDigits(municipio)) {
		return "", "", fmt.Errorf("municipio_codigo_dane debe contener 5 digitos")
	}
	if municipio != "" && departamento == "" {
		return "", "", fmt.Errorf("departamento_codigo_dane es obligatorio cuando se registra municipio_codigo_dane")
	}
	if municipio != "" && !strings.HasPrefix(municipio, departamento) {
		return "", "", fmt.Errorf("municipio_codigo_dane no pertenece al departamento indicado")
	}
	return departamento, municipio, nil
}
