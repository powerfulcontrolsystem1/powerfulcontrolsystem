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

func TestCountAndListEmpresaAuditoriaEventosWithPaginationAndSearch(t *testing.T) {
	dbConn := openFinanzasTestDB(t)
	if err := EnsureEmpresaAuditoriaSchema(dbConn); err != nil {
		t.Fatalf("ensure auditoria schema: %v", err)
	}

	fixtures := []EmpresaAuditoriaEvento{
		{
			EmpresaID:      123,
			Modulo:         "clientes",
			Accion:         "crear",
			Recurso:        "clientes",
			RecursoID:      1001,
			MetodoHTTP:     "POST",
			Endpoint:       "/api/empresa/clientes",
			Resultado:      "ok",
			CodigoHTTP:     201,
			UsuarioCreador: "cajero@empresa.com",
		},
		{
			EmpresaID:      123,
			Modulo:         "clientes",
			Accion:         "actualizar",
			Recurso:        "clientes",
			RecursoID:      1002,
			MetodoHTTP:     "PUT",
			Endpoint:       "/api/empresa/clientes",
			Resultado:      "ok",
			CodigoHTTP:     200,
			UsuarioCreador: "supervisor@empresa.com",
		},
		{
			EmpresaID:      123,
			Modulo:         "inventario",
			Accion:         "eliminar",
			Recurso:        "productos",
			RecursoID:      2001,
			MetodoHTTP:     "DELETE",
			Endpoint:       "/api/empresa/productos",
			Resultado:      "error",
			CodigoHTTP:     409,
			UsuarioCreador: "inventario@empresa.com",
		},
	}

	for i, fixture := range fixtures {
		if _, err := CreateEmpresaAuditoriaEvento(dbConn, fixture); err != nil {
			t.Fatalf("create fixture auditoria %d: %v", i+1, err)
		}
	}

	baseFilter := EmpresaAuditoriaEventoFilter{Modulo: "clientes", Limit: 1}
	total, err := CountEmpresaAuditoriaEventos(dbConn, 123, baseFilter)
	if err != nil {
		t.Fatalf("count auditoria eventos: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total=2 for modulo clientes, got %d", total)
	}

	firstPage, err := ListEmpresaAuditoriaEventos(dbConn, 123, baseFilter)
	if err != nil {
		t.Fatalf("list auditoria first page: %v", err)
	}
	if len(firstPage) != 1 {
		t.Fatalf("expected 1 row on first page, got %d", len(firstPage))
	}
	if firstPage[0].Accion != "actualizar" {
		t.Fatalf("expected first row accion=actualizar, got %q", firstPage[0].Accion)
	}

	secondPage, err := ListEmpresaAuditoriaEventos(dbConn, 123, EmpresaAuditoriaEventoFilter{Modulo: "clientes", Limit: 1, Offset: 1})
	if err != nil {
		t.Fatalf("list auditoria second page: %v", err)
	}
	if len(secondPage) != 1 {
		t.Fatalf("expected 1 row on second page, got %d", len(secondPage))
	}
	if secondPage[0].Accion != "crear" {
		t.Fatalf("expected second row accion=crear, got %q", secondPage[0].Accion)
	}

	byMethod, err := ListEmpresaAuditoriaEventos(dbConn, 123, EmpresaAuditoriaEventoFilter{Modulo: "clientes", MetodoHTTP: "POST", Limit: 10})
	if err != nil {
		t.Fatalf("list auditoria by method: %v", err)
	}
	if len(byMethod) != 1 || byMethod[0].MetodoHTTP != "POST" {
		t.Fatalf("expected one POST row for clientes, got len=%d metodo=%q", len(byMethod), func() string {
			if len(byMethod) == 0 {
				return ""
			}
			return byMethod[0].MetodoHTTP
		}())
	}

	bySearch, err := ListEmpresaAuditoriaEventos(dbConn, 123, EmpresaAuditoriaEventoFilter{Search: "supervisor", Limit: 10})
	if err != nil {
		t.Fatalf("list auditoria by search: %v", err)
	}
	if len(bySearch) != 1 {
		t.Fatalf("expected one row for search supervisor, got %d", len(bySearch))
	}
	if bySearch[0].UsuarioCreador != "supervisor@empresa.com" {
		t.Fatalf("expected supervisor row, got %q", bySearch[0].UsuarioCreador)
	}
}
