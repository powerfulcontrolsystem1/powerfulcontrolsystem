package db

import "testing"

func TestCreateAndListEmpresaAuditoriaEventos(t *testing.T) {
	dbConn := openFinanzasTestDB(t)
	if err := EnsureEmpresaAuditoriaSchema(dbConn); err != nil {
		t.Fatalf("ensure auditoria schema: %v", err)
	}

	_, err := CreateEmpresaAuditoriaEvento(dbConn, EmpresaAuditoriaEvento{
		EmpresaID:      55,
		Modulo:         "finanzas",
		Accion:         "procesar_asientos",
		Recurso:        "finanzas/asientos_contables",
		RecursoID:      901,
		MetodoHTTP:     "PUT",
		Endpoint:       "/api/empresa/finanzas/asientos_contables",
		Resultado:      "ok",
		CodigoHTTP:     200,
		RequestID:      "req-audit-001",
		IPOrigen:       "127.0.0.1",
		UserAgent:      "go-test",
		MetadataJSON:   `{"lote":100,"fallidos":0}`,
		RetencionDias:  120,
		UsuarioCreador: "conta@test.com",
	})
	if err != nil {
		t.Fatalf("create auditoria evento: %v", err)
	}

	_, err = CreateEmpresaAuditoriaEvento(dbConn, EmpresaAuditoriaEvento{
		EmpresaID:      55,
		Modulo:         "finanzas",
		Accion:         "aprobar",
		Recurso:        "finanzas/cierres_caja",
		MetodoHTTP:     "PUT",
		Endpoint:       "/api/empresa/finanzas/cierres_caja",
		Resultado:      "error",
		CodigoHTTP:     409,
		MetadataJSON:   `{"estado":"abierto"}`,
		UsuarioCreador: "conta@test.com",
	})
	if err != nil {
		t.Fatalf("create auditoria evento 2: %v", err)
	}

	rows, err := ListEmpresaAuditoriaEventos(dbConn, 55, EmpresaAuditoriaEventoFilter{
		Modulo:     "finanzas",
		Accion:     "procesar_asientos",
		RecursoID:  901,
		CodigoHTTP: 200,
		Resultado:  "ok",
		Limit:      20,
	})
	if err != nil {
		t.Fatalf("list auditoria eventos: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 auditoria row, got %d", len(rows))
	}
	if rows[0].Accion != "procesar_asientos" {
		t.Fatalf("expected accion procesar_asientos, got %q", rows[0].Accion)
	}
	if rows[0].CodigoHTTP != 200 {
		t.Fatalf("expected codigo_http=200, got %d", rows[0].CodigoHTTP)
	}
	if rows[0].RetencionDias != 120 {
		t.Fatalf("expected retencion_dias=120, got %d", rows[0].RetencionDias)
	}
}

func TestPurgeEmpresaAuditoriaEventos(t *testing.T) {
	dbConn := openFinanzasTestDB(t)
	if err := EnsureEmpresaAuditoriaSchema(dbConn); err != nil {
		t.Fatalf("ensure auditoria schema: %v", err)
	}

	_, err := CreateEmpresaAuditoriaEvento(dbConn, EmpresaAuditoriaEvento{
		EmpresaID:      88,
		Modulo:         "seguridad",
		Accion:         "actualizar",
		Recurso:        "usuarios",
		MetodoHTTP:     "PUT",
		Endpoint:       "/api/empresa/usuarios",
		Resultado:      "ok",
		CodigoHTTP:     200,
		FechaEvento:    "2024-01-10 10:00:00",
		UsuarioCreador: "admin@test.com",
	})
	if err != nil {
		t.Fatalf("create old auditoria row: %v", err)
	}

	_, err = CreateEmpresaAuditoriaEvento(dbConn, EmpresaAuditoriaEvento{
		EmpresaID:      88,
		Modulo:         "seguridad",
		Accion:         "crear",
		Recurso:        "usuarios",
		MetodoHTTP:     "POST",
		Endpoint:       "/api/empresa/usuarios",
		Resultado:      "ok",
		CodigoHTTP:     201,
		UsuarioCreador: "admin@test.com",
	})
	if err != nil {
		t.Fatalf("create fresh auditoria row: %v", err)
	}

	deleted, err := PurgeEmpresaAuditoriaEventos(dbConn, 88, 30)
	if err != nil {
		t.Fatalf("purge auditoria rows: %v", err)
	}
	if deleted < 1 {
		t.Fatalf("expected at least 1 row deleted, got %d", deleted)
	}

	rows, err := ListEmpresaAuditoriaEventos(dbConn, 88, EmpresaAuditoriaEventoFilter{Limit: 20})
	if err != nil {
		t.Fatalf("list auditoria rows after purge: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row after purge, got %d", len(rows))
	}
	if rows[0].CodigoHTTP != 201 {
		t.Fatalf("expected remaining row with codigo_http=201, got %d", rows[0].CodigoHTTP)
	}
}

func TestPurgeExpiredEmpresaAuditoriaEventos(t *testing.T) {
	dbConn := openFinanzasTestDB(t)
	if err := EnsureEmpresaAuditoriaSchema(dbConn); err != nil {
		t.Fatalf("ensure auditoria schema: %v", err)
	}

	_, err := CreateEmpresaAuditoriaEvento(dbConn, EmpresaAuditoriaEvento{
		EmpresaID:      91,
		Modulo:         "ventas",
		Accion:         "cerrar",
		Recurso:        "carritos_compra",
		MetodoHTTP:     "PUT",
		Endpoint:       "/api/empresa/carritos_compra",
		Resultado:      "ok",
		CodigoHTTP:     200,
		FechaEvento:    "2024-01-01 00:00:00",
		RetencionDias:  7,
		UsuarioCreador: "tester",
	})
	if err != nil {
		t.Fatalf("create expired auditoria row: %v", err)
	}

	_, err = CreateEmpresaAuditoriaEvento(dbConn, EmpresaAuditoriaEvento{
		EmpresaID:      91,
		Modulo:         "ventas",
		Accion:         "crear",
		Recurso:        "carritos_compra",
		MetodoHTTP:     "POST",
		Endpoint:       "/api/empresa/carritos_compra",
		Resultado:      "ok",
		CodigoHTTP:     201,
		RetencionDias:  365,
		UsuarioCreador: "tester",
	})
	if err != nil {
		t.Fatalf("create active auditoria row: %v", err)
	}

	deleted, err := PurgeExpiredEmpresaAuditoriaEventos(dbConn)
	if err != nil {
		t.Fatalf("purge expired auditoria rows: %v", err)
	}
	if deleted < 1 {
		t.Fatalf("expected at least 1 row deleted by expiration, got %d", deleted)
	}

	rows, err := ListEmpresaAuditoriaEventos(dbConn, 91, EmpresaAuditoriaEventoFilter{Limit: 20})
	if err != nil {
		t.Fatalf("list auditoria rows after purge expired: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row after purge expired, got %d", len(rows))
	}
	if rows[0].CodigoHTTP != 201 {
		t.Fatalf("expected remaining row codigo_http=201, got %d", rows[0].CodigoHTTP)
	}
}
