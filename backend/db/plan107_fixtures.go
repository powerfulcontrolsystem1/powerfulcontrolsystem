package db

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Plan107FixtureManifest is the immutable contract consumed by a staging-only
// fixture runner. Keeping it independent from a database connection makes the
// scenarios reviewable before a single operational record is created.
type Plan107FixtureManifest struct {
	Version        string                   `json:"version"`
	EmpresaID      int64                    `json:"empresa_id"`
	RunID          string                   `json:"run_id"`
	Prefix         string                   `json:"prefix"`
	StagingOnly    bool                     `json:"staging_only"`
	Scenarios      []Plan107FixtureScenario `json:"scenarios"`
	CleanupSteps   []string                 `json:"cleanup_steps"`
	ManifestSHA256 string                   `json:"manifest_sha256"`
}

// Plan107FixtureScenario is a business case with a stable idempotency key and
// its minimum accounting evidence. It deliberately contains no credentials,
// SQL, bank account or DIAN information.
type Plan107FixtureScenario struct {
	ID               string   `json:"id"`
	IdempotencyKey   string   `json:"idempotency_key"`
	Operation        string   `json:"operation"`
	ExpectedEvidence []string `json:"expected_evidence"`
}

// BuildPlan107FixtureManifest builds a deterministic, company-scoped fixture
// contract. The manifest is intentionally not an authorization to execute it.
func BuildPlan107FixtureManifest(empresaID int64, runID string) (Plan107FixtureManifest, error) {
	if empresaID <= 0 {
		return Plan107FixtureManifest{}, errors.New("empresa_id debe ser positivo")
	}
	runID = strings.ToUpper(strings.TrimSpace(runID))
	if runID == "" {
		runID = "P107-QA"
	}
	if !strings.HasPrefix(runID, "P107-QA") || len(runID) > 80 {
		return Plan107FixtureManifest{}, errors.New("run_id debe iniciar con P107-QA")
	}
	for _, r := range runID {
		if (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '-' && r != '_' {
			return Plan107FixtureManifest{}, errors.New("run_id contiene caracteres no permitidos")
		}
	}
	manifest := Plan107FixtureManifest{
		Version:     "p107-fixtures-v1",
		EmpresaID:   empresaID,
		RunID:       runID,
		Prefix:      runID,
		StagingOnly: true,
		Scenarios: []Plan107FixtureScenario{
			plan107FixtureScenario(runID, "OPENING", "saldos_apertura", "comprobante balanceado", "PUC", "asiento", "balance_prueba"),
			plan107FixtureScenario(runID, "SALE-CASH", "venta_menta_efectivo", "venta", "inventario", "caja", "impuesto", "evento", "asiento"),
			plan107FixtureScenario(runID, "SALE-TRANSFER", "venta_menta_transferencia", "venta", "pago_transferencia", "banco", "inventario", "asiento"),
			plan107FixtureScenario(runID, "SALE-CREDIT", "venta_menta_credito", "venta", "cuenta_por_cobrar", "inventario", "impuesto", "asiento"),
			plan107FixtureScenario(runID, "CXC-PARTIAL", "abono_parcial_cartera", "abono", "saldo_cxc", "caja_o_banco", "asiento"),
			plan107FixtureScenario(runID, "PURCHASE-CREDIT", "compra_credito_proveedor", "proveedor", "cuenta_por_pagar", "inventario", "iva_descontable", "asiento"),
			plan107FixtureScenario(runID, "CXP-PAYMENT", "abono_cuenta_por_pagar", "abono", "saldo_cxp", "banco", "asiento"),
			plan107FixtureScenario(runID, "TAX", "impuestos_y_retenciones", "iva_generado", "iva_descontable", "retenciones", "PUC", "reconciliacion"),
			plan107FixtureScenario(runID, "REVERSAL", "reverso_auditable", "reverso", "auditoria", "saldos_reconciliados"),
		},
		CleanupSteps: []string{
			"Verificar que todos los registros con el prefijo pertenezcan a la empresa objetivo.",
			"Revertir mediante documentos y movimientos auditables; no ejecutar borrados directos.",
			"Recalcular inventario, caja, bancos, CxC, CxP, impuestos y asientos antes de cerrar el run.",
			"Guardar el manifiesto, resultados y diferencias explicadas en la evidencia de staging.",
		},
	}
	encoded, err := json.Marshal(manifest)
	if err != nil {
		return Plan107FixtureManifest{}, fmt.Errorf("serializar manifiesto P107: %w", err)
	}
	digest := sha256.Sum256(encoded)
	manifest.ManifestSHA256 = hex.EncodeToString(digest[:])
	return manifest, nil
}

func plan107FixtureScenario(runID, id, operation string, evidence ...string) Plan107FixtureScenario {
	return Plan107FixtureScenario{
		ID:               id,
		IdempotencyKey:   strings.ToLower(runID + ":" + id),
		Operation:        operation,
		ExpectedEvidence: evidence,
	}
}
