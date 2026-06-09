package db

import (
	"os"
	"strings"
	"testing"
)

func TestEmpresaImpresorasColaUsaPostgreSQLYEmpresaID(t *testing.T) {
	raw, err := os.ReadFile("empresa_impresoras.go")
	if err != nil {
		t.Fatalf("read empresa_impresoras.go: %v", err)
	}
	src := string(raw)
	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS empresa_impresoras_cola",
		"empresa_id INTEGER NOT NULL",
		"ix_empresa_impresoras_cola_estado",
		"CrearEmpresaImpresoraTrabajo",
		"TomarEmpresaImpresoraTrabajos",
		"ActualizarEmpresaImpresoraTrabajoEstado",
		"ReintentarEmpresaImpresoraTrabajo",
	} {
		if !strings.Contains(src, required) {
			t.Fatalf("cola de impresoras debe conservar %s", required)
		}
	}
	for _, forbidden := range []string{"INSERT OR ", "AUTOINCREMENT", "sqlite"} {
		if strings.Contains(strings.ToLower(src), strings.ToLower(forbidden)) {
			t.Fatalf("empresa_impresoras.go no debe reintroducir SQLite; encontro %s", forbidden)
		}
	}
	start := strings.Index(src, "func TomarEmpresaImpresoraTrabajos(")
	if start < 0 {
		t.Fatal("no se encontro TomarEmpresaImpresoraTrabajos")
	}
	end := strings.Index(src[start:], "// ActualizarEmpresaImpresoraTrabajoEstado")
	if end < 0 {
		t.Fatal("no se encontro limite de TomarEmpresaImpresoraTrabajos")
	}
	body := src[start : start+end]
	for _, required := range []string{
		"WHERE c.empresa_id = ?",
		"COALESCE(c.estado, 'pendiente') = 'pendiente'",
		"COALESCE(c.agente_id, '') = '' OR COALESCE(c.agente_id, '') = ?",
		"COALESCE(c.estacion_id, 0) = 0 OR COALESCE(c.estacion_id, 0) = ?",
		"RowsAffected",
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("TomarEmpresaImpresoraTrabajos debe conservar aislamiento/claim %s en: %s", required, body)
		}
	}
}
