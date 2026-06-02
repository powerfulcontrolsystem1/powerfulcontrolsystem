package db

import "testing"

func TestNormalizeRolCatalogKeyCajaDeduplicaComoCajero(t *testing.T) {
	for _, raw := range []string{"Caja", "Caja principal", "caja_turno", "Cajero"} {
		if got := normalizeRolCatalogKey(raw); got != "cajero" {
			t.Fatalf("normalizeRolCatalogKey(%q)=%q, want cajero", raw, got)
		}
	}
}

func TestPreferredRolCatalogDisplayRankPrefiereCajero(t *testing.T) {
	if preferredRolCatalogDisplayRank("cajero", "Cajero") >= preferredRolCatalogDisplayRank("cajero", "Caja principal") {
		t.Fatal("el catalogo global debe preferir mostrar Cajero antes que variantes como Caja principal")
	}
}
