package handlers

import (
	"os"
	"strings"
	"testing"
)

func TestConfiguracionImpresoraDiferenciaReclamarDeImprimir(t *testing.T) {
	raw, err := os.ReadFile("../../web/administrar_empresa/configuracion_impresora.html")
	if err != nil {
		t.Fatalf("read configuracion de impresora: %v", err)
	}
	src := string(raw)
	for _, required := range []string{
		"Reclamar pendientes (sin imprimir)",
		"Confirmar papel impreso",
		"Reintentar si no salio",
		"Esta accion no imprime",
		"Confirma que el papel de este trabajo salio correctamente",
		"Reintentar puede producir una segunda copia",
	} {
		if !strings.Contains(src, required) {
			t.Fatalf("la UI de cola debe informar el limite fisico e idempotente: falta %q", required)
		}
	}

	start := strings.Index(src, "async function tomarTrabajosImpresion()")
	endOffset := -1
	if start >= 0 {
		endOffset = strings.Index(src[start:], "async function actualizarTrabajoImpresion")
	}
	if start < 0 || endOffset < 0 {
		t.Fatal("no se encontro el flujo de reclamo de cola")
	}
	end := start + endOffset
	if strings.Contains(src[start:end], "window.print") {
		t.Fatal("reclamar trabajos no debe presentarse como impresion fisica")
	}
}
