package db

import (
	"strings"
	"testing"
)

func TestEmpresaAIUsoDiarioUniqueMigrationIsRegistered(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatalf("PlatformMigrations(empresas): %v", err)
	}
	for _, migration := range migrations {
		if migration.Version != "20260731-001-ai-usage-unique-v1" {
			continue
		}
		if migration.Body != empresaAIUsoDiarioUniqueSchemaFingerprint || migration.Apply == nil {
			t.Fatal("AI daily usage uniqueness migration must be executable and checksummed")
		}
		return
	}
	t.Fatal("AI daily usage uniqueness migration is missing from enterprise catalog")
}

func TestEmpresaAIConsultaInsertMatchesBusinessColumns(t *testing.T) {
	query := empresaAIConsultaInsertStatement("CURRENT_TIMESTAMP")
	if got := strings.Count(query, "?"); got != 14 {
		t.Fatalf("empresa_ai_consultas placeholders = %d, want 14", got)
	}
	if got := strings.Count(query, "CURRENT_TIMESTAMP"); got != 2 {
		t.Fatalf("empresa_ai_consultas timestamp expressions = %d, want 2", got)
	}
}
