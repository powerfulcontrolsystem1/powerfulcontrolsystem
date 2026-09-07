package db

import "testing"

func TestEmpresaConfigBackupDefaultTablesCoverModernCompanyConfiguration(t *testing.T) {
	tables := EmpresaConfigBackupDefaultTables()
	got := make(map[string]bool, len(tables))
	for _, table := range tables {
		if got[table] {
			t.Fatalf("tabla duplicada en catalogo de configuracion: %s", table)
		}
		got[table] = true
	}

	required := []string{
		"empresa_configuracion_general",
		"empresa_configuracion_avanzada",
		"empresa_permisos_modulos",
		"empresa_permisos_paginas",
		"empresa_ai_modelo_preferido",
		"empresa_finanzas_configuracion",
		"empresa_corte_caja_configuracion",
		"empresa_inventario_configuracion",
		"empresa_tarifas_por_minutos_configuracion",
		"empresa_tarifas_por_minutos",
		"empresa_tarifas_por_dia",
		"empresa_integraciones_apis",
		"empresa_integraciones_bancos",
		"empresa_payment_settings",
		"empresa_sensor_puertas_devices",
		"empresa_reportes_plantillas",
		"empresa_reportes_programaciones",
		"empresa_venta_publica_configuracion",
		"empresa_venta_publica_paginas",
		"empresa_venta_publica_items",
		"admin_empresa_compartida",
	}
	for _, table := range required {
		if !got[table] {
			t.Fatalf("falta %s en el catalogo de configuracion por empresa", table)
		}
	}
}

func TestEmpresaConfigBackupDefaultTablesReturnsCopy(t *testing.T) {
	tables := EmpresaConfigBackupDefaultTables()
	if len(tables) == 0 {
		t.Fatal("catalogo de configuracion vacio")
	}
	tables[0] = "tabla_mutada_desde_test"

	fresh := EmpresaConfigBackupDefaultTables()
	if fresh[0] == "tabla_mutada_desde_test" {
		t.Fatal("EmpresaConfigBackupDefaultTables debe devolver una copia defensiva")
	}
}

func TestNormalizeEmpresaConfigBackupTablesAllowsOnlyConfigCatalog(t *testing.T) {
	normalized := normalizeEmpresaConfigBackupTables([]string{
		"EMPRESA_PERMISOS_MODULOS",
		"empresa_payment_settings",
		"empresa_backups",
		"ventas;drop",
		"empresa_no_catalogada",
	})

	want := map[string]bool{
		"empresa_payment_settings": true,
		"empresa_permisos_modulos": true,
	}
	if len(normalized) != len(want) {
		t.Fatalf("tablas normalizadas inesperadas: got %#v", normalized)
	}
	for _, table := range normalized {
		if !want[table] {
			t.Fatalf("tabla no permitida en normalizacion: %s", table)
		}
	}
}

func TestNormalizeEmpresaConfigBackupTablesFallsBackToFullCatalog(t *testing.T) {
	normalized := normalizeEmpresaConfigBackupTables([]string{"empresa_backups", "tabla_no_catalogada"})
	defaults := EmpresaConfigBackupDefaultTables()
	if len(normalized) != len(defaults) {
		t.Fatalf("debe volver al catalogo completo cuando no hay tablas validas: got %d want %d", len(normalized), len(defaults))
	}
}

func TestEmpresaBackupProtectedOperationalResetTable(t *testing.T) {
	protected := []string{
		"empresa_configuracion_general",
		"empresa_permisos_modulos",
		"empresa_impresoras",
		"empresa_integraciones_apis",
		"empresa_estacion_prefs",
		"empresa_backups",
		"users",
	}
	for _, table := range protected {
		if !empresaBackupProtectedOperationalResetTable(table) {
			t.Fatalf("tabla de configuracion/sistema debe estar protegida: %s", table)
		}
	}

	operational := []string{
		"carritos_compras",
		"carrito_compra_items",
		"codigos_de_descuento",
		"empresa_finanzas_movimientos",
		"facturas_electronicas",
		"clientes",
	}
	for _, table := range operational {
		if empresaBackupProtectedOperationalResetTable(table) {
			t.Fatalf("tabla operativa no debe estar protegida por defecto: %s", table)
		}
	}
}
