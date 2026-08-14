package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFinanzasDoesNotExposeRawAPIErrorBodies(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "finanzas.html"))
	if err != nil {
		t.Fatalf("read finanzas UI: %v", err)
	}
	page := string(raw)
	if strings.Contains(page, "const txt = await res.text();") {
		t.Fatal("finanzas must not render raw API response bodies")
	}
	for _, expected := range []string{
		"res.status === 401",
		"Tu sesión venció. Inicia sesión nuevamente para continuar.",
		"res.status === 403",
		"No tienes permiso para realizar esta acción.",
	} {
		if !strings.Contains(page, expected) {
			t.Fatalf("missing safe finance error contract %q", expected)
		}
	}
}

func TestFinanzasOpensPrintPreviewBeforeAsyncPrinterResolution(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "finanzas.html"))
	if err != nil {
		t.Fatalf("read finanzas UI: %v", err)
	}
	page := string(raw)
	start := strings.Index(page, "async function openPrint(item)")
	if start < 0 {
		t.Fatal("missing finance print preview function")
	}
	endOffset := strings.Index(page[start:], "\n    function renderMovimientos()")
	if endOffset < 0 {
		t.Fatal("could not isolate finance print preview function")
	}
	openPrint := page[start : start+endOffset]
	windowOpen := strings.Index(openPrint, "window.open('', '_blank'")
	printerResolution := strings.Index(openPrint, "await resolvePrinterForFinanzas()")
	if windowOpen < 0 || printerResolution < 0 {
		t.Fatal("print preview must open a window and resolve its printer")
	}
	if windowOpen > printerResolution {
		t.Fatal("print preview window must open synchronously before awaiting printer resolution")
	}
	if !strings.Contains(openPrint, "Preparando comprobante...") {
		t.Fatal("print preview must render a visible loading state while resolving its printer")
	}
}

func TestFinanzasHTTPPropagaContextoAlRepositorio(t *testing.T) {
	raw, err := os.ReadFile("finanzas.go")
	if err != nil {
		t.Fatalf("read finance handler: %v", err)
	}
	source := string(raw)
	for _, forbidden := range []string{
		"dbpkg.ListEmpresaFinanzasMovimientos(dbEmp",
		"dbpkg.CreateEmpresaFinanzasMovimiento(dbEmp",
		"dbpkg.UpdateEmpresaFinanzasMovimiento(dbEmp",
		"dbpkg.DeleteEmpresaFinanzasMovimiento(dbEmp",
		"dbpkg.GetEmpresaFinanzasConfiguracion(dbEmp",
		"dbpkg.UpsertEmpresaFinanzasConfiguracion(dbEmp",
		"dbpkg.ListEmpresaFinanzasPeriodos(dbEmp",
		"dbpkg.UpsertEmpresaFinanzasPeriodo(dbEmp",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("finance handler bypasses request context with %q", forbidden)
		}
	}
	for _, expected := range []string{
		"ListEmpresaFinanzasMovimientosContext(r.Context()",
		"CreateEmpresaFinanzasMovimientoContext(r.Context()",
		"UpdateEmpresaFinanzasMovimientoContext(r.Context()",
		"UpsertEmpresaFinanzasPeriodoContext(r.Context()",
	} {
		if !strings.Contains(source, expected) {
			t.Fatalf("missing cancellable finance operation %q", expected)
		}
	}
}
