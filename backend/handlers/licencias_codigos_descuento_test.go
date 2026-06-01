package handlers

import "testing"

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
