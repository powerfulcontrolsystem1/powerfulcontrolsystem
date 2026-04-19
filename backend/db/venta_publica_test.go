package db

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func openVentaPublicaSQLiteDB(t *testing.T, name string) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), name)
	dbConn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = dbConn.Close()
	})
	return dbConn
}

func TestEmpresaVentaPublicaConfigPersistsTemaVisual(t *testing.T) {
	dbConn := openVentaPublicaSQLiteDB(t, "venta_publica_tema_visual.db")
	if err := EnsureEmpresaVentaPublicaSchema(dbConn); err != nil {
		t.Fatalf("EnsureEmpresaVentaPublicaSchema: %v", err)
	}

	cfgID, err := UpsertEmpresaVentaPublicaConfig(dbConn, EmpresaVentaPublicaConfig{
		EmpresaID:     77,
		EmpresaSlug:   "cafeteria-norte",
		NombreTienda:  "Cafeteria Norte",
		TemaVisual:    "moderno",
		Moneda:        "cop",
		MostrarStock:  true,
		WompiActivo:   false,
		EpaycoActivo:  false,
		Estado:        "activo",
	})
	if err != nil {
		t.Fatalf("UpsertEmpresaVentaPublicaConfig: %v", err)
	}
	if cfgID <= 0 {
		t.Fatalf("expected cfgID > 0, got %d", cfgID)
	}

	stored, err := GetEmpresaVentaPublicaConfig(dbConn, 77)
	if err != nil {
		t.Fatalf("GetEmpresaVentaPublicaConfig: %v", err)
	}
	if stored.TemaVisual != "moderno" {
		t.Fatalf("expected tema_visual moderno, got %q", stored.TemaVisual)
	}
	if stored.Moneda != "COP" {
		t.Fatalf("expected moneda normalized COP, got %q", stored.Moneda)
	}
}
