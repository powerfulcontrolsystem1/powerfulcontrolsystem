package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestWriteInventarioEntidadNoDisponibleDevuelve404Seguro(t *testing.T) {
	recorder := httptest.NewRecorder()
	err := fmt.Errorf("%w: producto", dbpkg.ErrInventarioEntidadNoDisponible)

	if !writeInventarioEntidadNoDisponible(recorder, err) {
		t.Fatal("el error tipado de ownership no fue reconocido")
	}
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status=%d, want %d", recorder.Code, http.StatusNotFound)
	}
	body := strings.ToLower(recorder.Body.String())
	if strings.Contains(body, "id=") || strings.Contains(body, "empresa 7") {
		t.Fatalf("la respuesta filtro datos de otra empresa: %q", recorder.Body.String())
	}
}

func TestWriteInventarioEntidadNoDisponibleIgnoraErroresInternos(t *testing.T) {
	recorder := httptest.NewRecorder()
	if writeInventarioEntidadNoDisponible(recorder, fmt.Errorf("fallo interno")) {
		t.Fatal("un error interno no debe convertirse en un falso 404")
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("el helper escribio una respuesta para un error no reconocido: %q", recorder.Body.String())
	}
}
