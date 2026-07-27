package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSoportesComprasIARevisionExposesEditableHumanReview(t *testing.T) {
	page, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "soportes_compras_ia.html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"formEditarSoporte", "Revision humana editable", "btnGuardarRevision", "editProveedor"} {
		if !strings.Contains(string(page), want) {
			t.Fatalf("review UI missing %q", want)
		}
	}
	script, err := os.ReadFile(filepath.Join("..", "..", "web", "js", "soportes_compras_ia.js"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"editar_revision", "saveRevision", "/api/empresa/proveedores", "Aprueba nuevamente antes de contabilizar"} {
		if !strings.Contains(string(script), want) {
			t.Fatalf("review client contract missing %q", want)
		}
	}
}

func TestSoportesComprasIARevisionKeepsApprovalHumanAndCanonicalSupplier(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "db", "soportes_compras_ia.go"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, want := range []string{
		"func UpdateEmpresaSoporteComprasIARevision",
		"aprobado_por='', fecha_aprobacion=''",
		"GetEmpresaProveedorCxP(dbConn, empresaID, revision.ProveedorID)",
		"FROM proveedores WHERE empresa_id=? AND id=?",
		"func findEmpresaSoporteComprasIADuplicadoExcept",
		"(archivo_hash=? OR documento_numero=?)",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("review persistence contract missing %q", want)
		}
	}
}
