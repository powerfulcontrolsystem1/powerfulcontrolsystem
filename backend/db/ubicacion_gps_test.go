package db

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func openGPSDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "gps_test.db")
	dbConn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	dbConn.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = dbConn.Close()
	})
	return dbConn
}

func TestEmpresaGPSDispositivosYRecorridosCRUD(t *testing.T) {
	dbConn := openGPSDB(t)
	if err := EnsureEmpresaUbicacionGPSSchema(dbConn); err != nil {
		t.Fatalf("ensure gps schema: %v", err)
	}

	dispositivoID, err := CreateEmpresaGPSDispositivo(dbConn, EmpresaGPSDispositivo{
		EmpresaID:   22,
		Codigo:      "CAM-01",
		Nombre:      "Camion norte",
		Descripcion: "Ruta principal",
	})
	if err != nil {
		t.Fatalf("create dispositivo: %v", err)
	}

	dispositivos, err := GetEmpresaGPSDispositivos(dbConn, 22, true, "")
	if err != nil {
		t.Fatalf("list dispositivos: %v", err)
	}
	if len(dispositivos) != 1 {
		t.Fatalf("expected 1 dispositivo, got %d", len(dispositivos))
	}

	puntoAID, err := CreateEmpresaGPSRecorrido(dbConn, EmpresaGPSRecorrido{
		EmpresaID:       22,
		DispositivoID:   dispositivoID,
		Latitud:         4.650123,
		Longitud:        -74.120456,
		PrecisionMetros: 7,
		VelocidadKMH:    22,
		Fuente:          "simulado_10s",
	})
	if err != nil {
		t.Fatalf("create punto A: %v", err)
	}

	_, err = CreateEmpresaGPSRecorrido(dbConn, EmpresaGPSRecorrido{
		EmpresaID:       22,
		DispositivoID:   dispositivoID,
		Latitud:         4.651111,
		Longitud:        -74.119999,
		PrecisionMetros: 6,
		VelocidadKMH:    25,
		Fuente:          "simulado_10s",
	})
	if err != nil {
		t.Fatalf("create punto B: %v", err)
	}

	recorridos, err := ListEmpresaGPSRecorridos(dbConn, 22, dispositivoID, true, 0, 100)
	if err != nil {
		t.Fatalf("list recorridos: %v", err)
	}
	if len(recorridos) != 2 {
		t.Fatalf("expected 2 recorridos, got %d", len(recorridos))
	}

	dispositivos, err = GetEmpresaGPSDispositivos(dbConn, 22, true, "cam")
	if err != nil {
		t.Fatalf("list dispositivos with filter: %v", err)
	}
	if len(dispositivos) != 1 {
		t.Fatalf("expected 1 dispositivo filtered, got %d", len(dispositivos))
	}
	if dispositivos[0].UltimaLatitud == 0 || dispositivos[0].UltimaLongitud == 0 {
		t.Fatalf("expected ultima posicion actualizada, got lat=%f lng=%f", dispositivos[0].UltimaLatitud, dispositivos[0].UltimaLongitud)
	}

	if err := UpdateEmpresaGPSRecorrido(dbConn, EmpresaGPSRecorrido{
		ID:              puntoAID,
		EmpresaID:       22,
		DispositivoID:   dispositivoID,
		Latitud:         4.652222,
		Longitud:        -74.118888,
		PrecisionMetros: 5,
		VelocidadKMH:    18,
		Fuente:          "manual",
	}); err != nil {
		t.Fatalf("update recorrido: %v", err)
	}

	if err := SetEmpresaGPSRecorridoEstado(dbConn, 22, puntoAID, "inactivo"); err != nil {
		t.Fatalf("set recorrido estado: %v", err)
	}

	activos, err := ListEmpresaGPSRecorridos(dbConn, 22, dispositivoID, false, 0, 100)
	if err != nil {
		t.Fatalf("list recorridos activos: %v", err)
	}
	if len(activos) != 1 {
		t.Fatalf("expected 1 recorrido activo after toggle, got %d", len(activos))
	}

	if err := SetEmpresaGPSDispositivoEstado(dbConn, 22, dispositivoID, "inactivo"); err != nil {
		t.Fatalf("set dispositivo estado: %v", err)
	}

	dispositivosActivos, err := GetEmpresaGPSDispositivos(dbConn, 22, false, "")
	if err != nil {
		t.Fatalf("list dispositivos activos: %v", err)
	}
	if len(dispositivosActivos) != 0 {
		t.Fatalf("expected 0 dispositivos activos, got %d", len(dispositivosActivos))
	}

	if err := DeleteEmpresaGPSDispositivo(dbConn, 22, dispositivoID); err != nil {
		t.Fatalf("delete dispositivo: %v", err)
	}

	dispositivos, err = GetEmpresaGPSDispositivos(dbConn, 22, true, "")
	if err != nil {
		t.Fatalf("list dispositivos after delete: %v", err)
	}
	if len(dispositivos) != 0 {
		t.Fatalf("expected 0 dispositivos after delete, got %d", len(dispositivos))
	}

	recorridos, err = ListEmpresaGPSRecorridos(dbConn, 22, 0, true, 0, 100)
	if err != nil {
		t.Fatalf("list recorridos after delete: %v", err)
	}
	if len(recorridos) != 0 {
		t.Fatalf("expected 0 recorridos after delete dispositivo, got %d", len(recorridos))
	}
}
