package handlers

import (
	"strings"
	"testing"
	"time"
)

func TestParseLicenciaDiscountCodeAdminLines(t *testing.T) {
	items := parseLicenciaDiscountCodeAdminLines(" LANZA20 = 20%\n# CORTESIA=gratis\nMOTO50=50000\nINVALIDO\n")
	if len(items) != 3 {
		t.Fatalf("esperaba 3 codigos, obtuvo %d", len(items))
	}
	if items[0].Codigo != "LANZA20" || !items[0].Activo || items[0].Tipo != "porcentaje" || items[0].Valor != 20 {
		t.Fatalf("codigo porcentaje inesperado: %+v", items[0])
	}
	if items[1].Codigo != "CORTESIA" || items[1].Activo || items[1].Tipo != "gratis" {
		t.Fatalf("codigo gratis inactivo inesperado: %+v", items[1])
	}
	if items[2].Codigo != "MOTO50" || !items[2].Activo || items[2].Tipo != "valor" || items[2].Valor != 50000 {
		t.Fatalf("codigo valor inesperado: %+v", items[2])
	}
}

func TestParseLicenciaDiscountCodeAdminLinesMetadata(t *testing.T) {
	items := parseLicenciaDiscountCodeAdminLines(`PCS-1234=15% | {"nombre":"Lanzamiento","descripcion":"Promo inicial","email":"cliente@example.com","vence":"2026-07-31","enviado_email":"true","ultimo_envio":"2026-06-19T10:00:00Z"}`)
	if len(items) != 1 {
		t.Fatalf("esperaba 1 codigo, obtuvo %d", len(items))
	}
	item := items[0]
	if item.Spec != "15%" || item.Nombre != "Lanzamiento" || item.Descripcion != "Promo inicial" || item.Email != "cliente@example.com" || item.Vence != "2026-07-31" || !item.EnviadoEmail {
		t.Fatalf("metadata inesperada: %+v", item)
	}
	line := formatLicenciaDiscountCodeLineSpec(item)
	if !strings.Contains(line, "15% |") || !strings.Contains(line, `"email":"cliente@example.com"`) {
		t.Fatalf("linea serializada inesperada: %s", line)
	}
}

func TestLicenciaDiscountCodeExpired(t *testing.T) {
	if !licenciaDiscountCodeExpired(`10% | {"vence":"2026-06-01"}`, time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("esperaba codigo vencido")
	}
	if licenciaDiscountCodeExpired(`10% | {"vence":"2026-06-30"}`, time.Date(2026, 6, 19, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("no esperaba codigo vencido")
	}
}

func TestBuildLicenciaDiscountCodeAdminItemValidaCodigoYSpec(t *testing.T) {
	item, err := buildLicenciaDiscountCodeAdminItem(licenciaDiscountCodeAdminPayload{
		Codigo: " promo mayo ",
		Tipo:   "porcentaje",
		Valor:  15,
		Activo: true,
	})
	if err != nil {
		t.Fatalf("no esperaba error: %v", err)
	}
	if item.Codigo != "PROMO-MAYO" || item.Spec != "15.00%" || item.Tipo != "porcentaje" {
		t.Fatalf("item inesperado: %+v", item)
	}

	if _, err := buildLicenciaDiscountCodeAdminItem(licenciaDiscountCodeAdminPayload{
		Codigo: "CODIGO CON ESPACIO!",
		Tipo:   "valor",
		Valor:  10000,
		Activo: true,
	}); err == nil {
		t.Fatalf("esperaba error por caracter invalido")
	}
}
