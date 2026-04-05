package db

import (
	"math"
	"testing"
)

func TestEmpresaPropinasConfigUpsertAndGet(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaPropinasSchema(dbConn); err != nil {
		t.Fatalf("ensure propinas schema: %v", err)
	}

	id, err := UpsertEmpresaPropinasConfiguracion(dbConn, EmpresaPropinasConfiguracion{
		EmpresaID:              1,
		HabilitarPropina:       true,
		PorcentajePropina:      12.5,
		ModoDistribucion:       EmpresaPropinaModoPorUsuario,
		AplicarAutomaticamente: true,
		UsuarioCreador:         "qa@empresa.com",
		Estado:                 "activo",
	})
	if err != nil {
		t.Fatalf("upsert config propinas: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected id > 0, got %d", id)
	}

	cfg, err := GetEmpresaPropinasConfiguracion(dbConn, 1)
	if err != nil {
		t.Fatalf("get config propinas: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected cfg not nil")
	}
	if !cfg.HabilitarPropina {
		t.Fatal("expected habilitar_propina=true")
	}
	if math.Abs(cfg.PorcentajePropina-12.5) > 0.001 {
		t.Fatalf("expected porcentaje_propina 12.5, got %.4f", cfg.PorcentajePropina)
	}
	if cfg.ModoDistribucion != EmpresaPropinaModoPorUsuario {
		t.Fatalf("expected modo por_usuario, got %q", cfg.ModoDistribucion)
	}
}

func TestEmpresaPropinasReporteDistribucionUniversal(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaPropinasSchema(dbConn); err != nil {
		t.Fatalf("ensure propinas schema: %v", err)
	}

	if _, err := dbConn.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT,
		name TEXT,
		empresa_id INTEGER,
		estado TEXT DEFAULT 'activo'
	);`); err != nil {
		t.Fatalf("create users table: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO users (email, name, empresa_id, estado) VALUES
		('mesero1@empresa.com', 'Mesero 1', 1, 'activo'),
		('mesero2@empresa.com', 'Mesero 2', 1, 'activo'),
		('inactivo@empresa.com', 'Inactivo', 1, 'inactivo')`); err != nil {
		t.Fatalf("seed users: %v", err)
	}

	if _, err := UpsertEmpresaPropinasConfiguracion(dbConn, EmpresaPropinasConfiguracion{
		EmpresaID:              1,
		HabilitarPropina:       true,
		PorcentajePropina:      10,
		ModoDistribucion:       EmpresaPropinaModoUniversal,
		AplicarAutomaticamente: true,
		UsuarioCreador:         "contabilidad@empresa.com",
		Estado:                 "activo",
	}); err != nil {
		t.Fatalf("upsert config universal: %v", err)
	}

	if _, err := CreateEmpresaPropinaMovimiento(dbConn, EmpresaPropinaMovimiento{
		EmpresaID:         1,
		CarritoID:         10,
		VentaReferencia:   "CAJ-10",
		UsuarioOrigen:     "mesero1@empresa.com",
		ModoDistribucion:  EmpresaPropinaModoUniversal,
		Moneda:            "COP",
		BaseCobro:         20000,
		PorcentajePropina: 10,
		MontoPropina:      2000,
		UsuarioCreador:    "mesero1@empresa.com",
	}); err != nil {
		t.Fatalf("create universal movement: %v", err)
	}

	if _, err := CreateEmpresaPropinaMovimiento(dbConn, EmpresaPropinaMovimiento{
		EmpresaID:         1,
		CarritoID:         11,
		VentaReferencia:   "CAJ-11",
		UsuarioOrigen:     "mesero1@empresa.com",
		UsuarioAsignado:   "mesero1@empresa.com",
		ModoDistribucion:  EmpresaPropinaModoPorUsuario,
		Moneda:            "COP",
		BaseCobro:         5000,
		PorcentajePropina: 10,
		MontoPropina:      500,
		UsuarioCreador:    "mesero1@empresa.com",
	}); err != nil {
		t.Fatalf("create per-user movement: %v", err)
	}

	report, err := GetEmpresaPropinasReporte(dbConn, 1, EmpresaPropinaMovimientoFilter{Limit: 100})
	if err != nil {
		t.Fatalf("get report propinas: %v", err)
	}
	if report == nil {
		t.Fatal("expected report not nil")
	}

	if math.Abs(report.Resumen.TotalPropinas-2500) > 0.001 {
		t.Fatalf("expected total propinas 2500, got %.4f", report.Resumen.TotalPropinas)
	}
	if report.Resumen.UsuariosActivos != 2 {
		t.Fatalf("expected usuarios activos=2, got %d", report.Resumen.UsuariosActivos)
	}
	if math.Abs(report.Resumen.CuotaUniversalPorUsuario-1000) > 0.001 {
		t.Fatalf("expected cuota universal 1000, got %.4f", report.Resumen.CuotaUniversalPorUsuario)
	}

	user1Found := false
	user2Found := false
	for _, row := range report.Usuarios {
		if row.UsuarioClave == "mesero1@empresa.com" {
			user1Found = true
			if math.Abs(row.PropinaPorUsuario-500) > 0.001 {
				t.Fatalf("expected user1 propina por usuario 500, got %.4f", row.PropinaPorUsuario)
			}
			if math.Abs(row.PropinaUniversal-1000) > 0.001 {
				t.Fatalf("expected user1 propina universal 1000, got %.4f", row.PropinaUniversal)
			}
			if math.Abs(row.PropinaTotal-1500) > 0.001 {
				t.Fatalf("expected user1 total 1500, got %.4f", row.PropinaTotal)
			}
		}
		if row.UsuarioClave == "mesero2@empresa.com" {
			user2Found = true
			if math.Abs(row.PropinaPorUsuario-0) > 0.001 {
				t.Fatalf("expected user2 propina por usuario 0, got %.4f", row.PropinaPorUsuario)
			}
			if math.Abs(row.PropinaUniversal-1000) > 0.001 {
				t.Fatalf("expected user2 propina universal 1000, got %.4f", row.PropinaUniversal)
			}
			if math.Abs(row.PropinaTotal-1000) > 0.001 {
				t.Fatalf("expected user2 total 1000, got %.4f", row.PropinaTotal)
			}
		}
	}
	if !user1Found {
		t.Fatal("expected mesero1 row in report")
	}
	if !user2Found {
		t.Fatal("expected mesero2 row in report")
	}
}
