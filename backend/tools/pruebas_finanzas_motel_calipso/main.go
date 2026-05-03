package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func must(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}

func main() {
	var (
		empresaID = flag.Int64("empresa_id", 7, "empresa_id (Motel Calipso=7)")
		periodo   = flag.String("periodo", "", "periodo contable YYYY-MM (por defecto mes actual)")
		usuario   = flag.String("usuario", "qa_tool", "usuario_creador/audit")
	)
	flag.Parse()

	dsn := strings.TrimSpace(os.Getenv("DB_EMPRESAS_DSN"))
	if dsn == "" {
		log.Fatal("DB_EMPRESAS_DSN no está definido (debe apuntar a pcs_empresas en PostgreSQL)")
	}
	if *empresaID <= 0 {
		log.Fatal("empresa_id inválido")
	}
	if strings.TrimSpace(*periodo) == "" {
		*periodo = time.Now().Format("2006-01")
	}

	driver := dbpkg.PostgresCompatDriverName()
	db, err := sql.Open(driver, dsn)
	must(err, "sql.Open")
	defer db.Close()
	must(db.Ping(), "db.Ping")

	_ = dbpkg.EnsurePostgresRuntimeCompat(db)
	must(dbpkg.EnsureEmpresaFinanzasSchema(db), "EnsureEmpresaFinanzasSchema")
	must(dbpkg.EnsureEmpresaModulosFaltantesSchema(db), "EnsureEmpresaModulosFaltantesSchema")

	log.Printf("OK conectó a DB_EMPRESAS_DSN | empresa_id=%d periodo=%s", *empresaID, *periodo)

	// 1) Configuración: Get y Upsert (idempotente)
	cfg, err := dbpkg.GetEmpresaFinanzasConfiguracion(db, *empresaID)
	must(err, "GetEmpresaFinanzasConfiguracion")
	cfg.UsuarioCreador = strings.TrimSpace(*usuario)
	cfg.Observaciones = "qa: verificación configuración finanzas"
	cfg.RequiereAprobacion = true
	cfg.IntegracionContableDestino = "siigo"
	idCfg, err := dbpkg.UpsertEmpresaFinanzasConfiguracion(db, *cfg)
	must(err, "UpsertEmpresaFinanzasConfiguracion")
	log.Printf("OK configuración finanzas upsert id=%d destino=%s requiere_aprobacion=%t", idCfg, cfg.IntegracionContableDestino, cfg.RequiereAprobacion)

	// 2) Periodo: abrir (upsert) y asegurar que no está cerrado
	_, err = dbpkg.UpsertEmpresaFinanzasPeriodo(db, dbpkg.EmpresaFinanzasPeriodo{
		EmpresaID:      *empresaID,
		Periodo:        *periodo,
		Estado:         "abierto",
		UsuarioCreador: strings.TrimSpace(*usuario),
		Observaciones:  "qa: apertura periodo",
	})
	must(err, "UpsertEmpresaFinanzasPeriodo(abierto)")
	cerrado, err := dbpkg.IsEmpresaFinanzasPeriodoCerrado(db, *empresaID, *periodo)
	must(err, "IsEmpresaFinanzasPeriodoCerrado")
	if cerrado {
		log.Fatal("periodo quedó cerrado inesperadamente")
	}
	log.Printf("OK periodo %s está abierto", *periodo)

	// 3) Movimientos: crear ingreso y egreso con retenciones
	ingID, err := dbpkg.CreateEmpresaFinanzasMovimiento(db, dbpkg.EmpresaFinanzasMovimiento{
		EmpresaID:         *empresaID,
		TipoMovimiento:    "ingreso",
		PeriodoContable:   *periodo,
		Categoria:         "ventas",
		Concepto:          "Ingreso QA (venta)",
		Descripcion:       "prueba operativa automatizada (ingreso)",
		MetodoPago:        "efectivo",
		Moneda:            "COP",
		Monto:             100000,
		Impuesto:          0,
		RetencionFuente:   0,
		RetencionICA:      0,
		RetencionIVA:      0,
		UsuarioCreador:    strings.TrimSpace(*usuario),
		TipoComprobante:   "recibo_interno",
		NumeroComprobante: "", // autogenera
		Estado:            "activo",
		Observaciones:     "qa ingreso",
	})
	must(err, "CreateEmpresaFinanzasMovimiento(ingreso)")
	log.Printf("OK creó ingreso movimiento_id=%d", ingID)

	egrID, err := dbpkg.CreateEmpresaFinanzasMovimiento(db, dbpkg.EmpresaFinanzasMovimiento{
		EmpresaID:         *empresaID,
		TipoMovimiento:    "egreso",
		PeriodoContable:   *periodo,
		Categoria:         "compras",
		Concepto:          "Egreso QA (compra)",
		Descripcion:       "prueba operativa automatizada (egreso) con retenciones",
		MetodoPago:        "transferencia_bancaria",
		Moneda:            "COP",
		Monto:             50000,
		Impuesto:          9500,
		RetencionFuente:   2000,
		RetencionICA:      500,
		RetencionIVA:      0,
		UsuarioCreador:    strings.TrimSpace(*usuario),
		TipoComprobante:   "soporte_externo",
		NumeroComprobante: "",
		Estado:            "activo",
		Observaciones:     "qa egreso",
	})
	must(err, "CreateEmpresaFinanzasMovimiento(egreso)")
	log.Printf("OK creó egreso movimiento_id=%d", egrID)

	// 4) Listado / filtros básicos
	rows, err := dbpkg.ListEmpresaFinanzasMovimientos(db, *empresaID, dbpkg.EmpresaFinanzasMovimientoFilter{
		Periodo: *periodo,
		Limit:   20,
	})
	must(err, "ListEmpresaFinanzasMovimientos")
	log.Printf("OK list movimientos periodo=%s count=%d (limit=20)", *periodo, len(rows))

	// 5) Cierre de periodo: al cerrarlo, bloquear creación (error esperado)
	must(dbpkg.SetEmpresaFinanzasPeriodoEstado(db, *empresaID, *periodo, "cerrado", strings.TrimSpace(*usuario), "qa: cierre periodo"), "SetEmpresaFinanzasPeriodoEstado(cerrado)")
	cerrado, err = dbpkg.IsEmpresaFinanzasPeriodoCerrado(db, *empresaID, *periodo)
	must(err, "IsEmpresaFinanzasPeriodoCerrado(post cierre)")
	if !cerrado {
		log.Fatal("periodo no quedó cerrado")
	}
	log.Printf("OK periodo %s quedó cerrado", *periodo)

	_, err = dbpkg.CreateEmpresaFinanzasMovimiento(db, dbpkg.EmpresaFinanzasMovimiento{
		EmpresaID:       *empresaID,
		TipoMovimiento:  "egreso",
		PeriodoContable: *periodo,
		Categoria:       "impuestos",
		Concepto:        "Egreso QA bloqueado por periodo cerrado",
		MetodoPago:      "efectivo",
		Moneda:          "COP",
		Monto:           1000,
		UsuarioCreador:  strings.TrimSpace(*usuario),
	})
	if err == nil {
		log.Fatal("ERROR: se permitió crear movimiento en periodo cerrado (esperado: ErrPeriodoFinancieroCerrado)")
	}
	if !errors.Is(err, dbpkg.ErrPeriodoFinancieroCerrado) {
		log.Fatalf("ERROR: movimiento en periodo cerrado devolvió error inesperado: %v", err)
	}
	log.Printf("OK bloqueo por periodo cerrado (ErrPeriodoFinancieroCerrado) validado")

	// 6) Reabrir periodo
	must(dbpkg.SetEmpresaFinanzasPeriodoEstado(db, *empresaID, *periodo, "abierto", strings.TrimSpace(*usuario), "qa: reapertura periodo"), "SetEmpresaFinanzasPeriodoEstado(abierto)")
	cerrado, err = dbpkg.IsEmpresaFinanzasPeriodoCerrado(db, *empresaID, *periodo)
	must(err, "IsEmpresaFinanzasPeriodoCerrado(post reapertura)")
	if cerrado {
		log.Fatal("periodo quedó cerrado luego de reapertura")
	}
	log.Printf("OK periodo %s reabierto", *periodo)

	// 7) Plan de cuentas, CxC/CxP y conciliacion bancaria con extracto QA.
	planID, err := dbpkg.CreateEmpresaGenericRow(db, "empresa_plan_cuentas", *empresaID, map[string]interface{}{
		"codigo":              "QA110505",
		"nombre":              "QA Caja general pruebas",
		"tipo_cuenta":         "activo",
		"naturaleza":          "debito",
		"nivel":               2,
		"cuenta_padre_codigo": "1105",
		"admite_movimiento":   1,
		"aplica_impuesto":     0,
		"cuenta_clave":        "qa_caja",
		"usuario_creador":     strings.TrimSpace(*usuario),
		"estado":              "activo",
		"observaciones":       "QA plan de cuentas",
	}, []string{"codigo", "nombre", "tipo_cuenta", "naturaleza", "nivel", "cuenta_padre_codigo", "admite_movimiento", "aplica_impuesto", "cuenta_clave", "usuario_creador", "estado", "observaciones"})
	if err != nil && !strings.Contains(strings.ToLower(err.Error()), "unique") {
		must(err, "CreateEmpresaGenericRow(plan_cuentas)")
	}
	log.Printf("OK plan de cuentas QA validado id=%d (0 si ya existia)", planID)

	qaSuffix := time.Now().Format("150405")
	cxcDoc := "QA-CXC-DOC-" + qaSuffix
	cxcID, err := dbpkg.CreateEmpresaGenericRow(db, "empresa_cuentas_por_cobrar", *empresaID, map[string]interface{}{
		"codigo":            "QA-CXC-" + qaSuffix,
		"cliente_nombre":    "Cliente QA Finanzas",
		"documento_tipo":    "factura",
		"documento_codigo":  cxcDoc,
		"fecha_emision":     time.Now().Format("2006-01-02"),
		"fecha_vencimiento": time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
		"valor_original":    100000,
		"valor_pagado":      0,
		"saldo":             100000,
		"estado_cartera":    "pendiente",
		"moneda":            "COP",
		"periodo_contable":  *periodo,
		"usuario_creador":   strings.TrimSpace(*usuario),
		"estado":            "activo",
		"observaciones":     "QA CxC",
	}, []string{"codigo", "cliente_nombre", "documento_tipo", "documento_codigo", "fecha_emision", "fecha_vencimiento", "valor_original", "valor_pagado", "saldo", "estado_cartera", "moneda", "periodo_contable", "usuario_creador", "estado", "observaciones"})
	must(err, "CreateEmpresaGenericRow(cxc)")
	log.Printf("OK CxC QA creada id=%d", cxcID)

	pagoCxCID, err := dbpkg.CreateEmpresaFinanzasMovimiento(db, dbpkg.EmpresaFinanzasMovimiento{
		EmpresaID:         *empresaID,
		TipoMovimiento:    "ingreso",
		PeriodoContable:   *periodo,
		Categoria:         "cuentas_cobrar",
		Subcategoria:      "abono_cartera",
		Concepto:          "Abono QA CxC",
		Descripcion:       "pago parcial QA cartera",
		MetodoPago:        "transferencia_bancaria",
		Moneda:            "COP",
		Monto:             60000,
		Total:             60000,
		TotalNeto:         60000,
		TerceroNombre:     "Cliente QA Finanzas",
		TipoComprobante:   "recibo_interno",
		NumeroComprobante: cxcDoc,
		ReferenciaExterna: cxcDoc,
		UsuarioCreador:    strings.TrimSpace(*usuario),
		Estado:            "activo",
		Observaciones:     "QA abono CxC",
	})
	must(err, "CreateEmpresaFinanzasMovimiento(pago cxc)")
	must(dbpkg.UpdateEmpresaGenericRow(db, "empresa_cuentas_por_cobrar", *empresaID, cxcID, map[string]interface{}{
		"valor_pagado":      60000,
		"saldo":             40000,
		"estado_cartera":    "parcial",
		"dias_mora":         1,
		"fecha_ultimo_pago": time.Now().Format("2006-01-02 15:04:05"),
		"observaciones":     "QA CxC con abono movimiento_id=" + fmt.Sprint(pagoCxCID),
	}, []string{"valor_pagado", "saldo", "estado_cartera", "dias_mora", "fecha_ultimo_pago", "observaciones"}), "UpdateEmpresaGenericRow(cxc abono)")
	log.Printf("OK abono CxC QA movimiento_id=%d", pagoCxCID)

	cxpSuffix := time.Now().Format("150405")
	cxpID, err := dbpkg.CreateEmpresaGenericRow(db, "empresa_cuentas_por_pagar", *empresaID, map[string]interface{}{
		"codigo":            "QA-CXP-" + cxpSuffix,
		"proveedor_nombre":  "Proveedor QA Finanzas",
		"documento_tipo":    "factura_proveedor",
		"documento_codigo":  "QA-CXP-DOC-" + cxpSuffix,
		"fecha_emision":     time.Now().Format("2006-01-02"),
		"fecha_vencimiento": time.Now().AddDate(0, 0, 7).Format("2006-01-02"),
		"valor_original":    80000,
		"valor_pagado":      0,
		"saldo":             80000,
		"estado_cartera":    "pendiente",
		"moneda":            "COP",
		"periodo_contable":  *periodo,
		"usuario_creador":   strings.TrimSpace(*usuario),
		"estado":            "activo",
		"observaciones":     "QA CxP",
	}, []string{"codigo", "proveedor_nombre", "documento_tipo", "documento_codigo", "fecha_emision", "fecha_vencimiento", "valor_original", "valor_pagado", "saldo", "estado_cartera", "moneda", "periodo_contable", "usuario_creador", "estado", "observaciones"})
	must(err, "CreateEmpresaGenericRow(cxp)")
	log.Printf("OK CxP QA creada id=%d", cxpID)

	_, err = dbpkg.UpsertEmpresaFinanzasMovimientosBancarios(db, *empresaID, []dbpkg.EmpresaFinanzasMovimientoBancario{
		{
			EmpresaID:          *empresaID,
			PeriodoContable:    *periodo,
			FechaMovimiento:    time.Now().Format("2006-01-02"),
			FechaValor:         time.Now().Format("2006-01-02"),
			CuentaBancaria:     "QA-0001",
			BancoNombre:        "Banco QA",
			TipoMovimiento:     "ingreso",
			Descripcion:        "Extracto QA CxC",
			ReferenciaBancaria: cxcDoc,
			DocumentoCodigo:    cxcDoc,
			Moneda:             "COP",
			Monto:              60000,
			Total:              60000,
			Origen:             "qa_runner",
			UsuarioCreador:     strings.TrimSpace(*usuario),
			Estado:             "activo",
			Observaciones:      "QA extracto bancario",
		},
	})
	must(err, "UpsertEmpresaFinanzasMovimientosBancarios")
	conciliacion, err := dbpkg.ConciliarEmpresaMovimientosBancariosAutomatico(db, *empresaID, dbpkg.EmpresaConciliacionBancariaAutoConfig{
		PeriodoContable: *periodo,
		ToleranciaDias:  3,
		ToleranciaMonto: 1,
		Limit:           100,
		Usuario:         strings.TrimSpace(*usuario),
	})
	must(err, "ConciliarEmpresaMovimientosBancariosAutomatico")
	log.Printf("OK conciliación bancaria QA revisados=%d conciliados=%d pendientes=%d", conciliacion.Revisados, conciliacion.Conciliados, conciliacion.Pendientes)

	// 8) Cierres de caja: validar transiciones y bloqueo cuando aprobado
	cajaCodigo := "CAJA_QA_" + time.Now().Format("150405")
	cierreID, err := dbpkg.CreateEmpresaCierreCaja(db, dbpkg.EmpresaCierreCaja{
		EmpresaID:        *empresaID,
		SucursalID:       0,
		CajaCodigo:       cajaCodigo,
		Turno:            "general",
		FechaOperacion:   time.Now().Format("2006-01-02"),
		Moneda:           "COP",
		AperturaMonto:    10000,
		IngresosEfectivo: 50000,
		EgresosEfectivo:  5000,
		RetirosEfectivo:  0,
		UmbralIncidencia: 0,
		UsuarioCreador:   strings.TrimSpace(*usuario),
		Estado:           "activo",
		Observaciones:    "qa: apertura caja",
	})
	must(err, "CreateEmpresaCierreCaja")
	log.Printf("OK creó cierre_caja id=%d estado=abierto", cierreID)

	cajaFisica := 55000.0
	must(dbpkg.SetEmpresaCierreCajaEstado(db, *empresaID, cierreID, "cerrado", &cajaFisica, strings.TrimSpace(*usuario), "qa: arqueo"), "SetEmpresaCierreCajaEstado(cerrado)")
	log.Printf("OK caja id=%d pasó a cerrado con arqueo", cierreID)

	must(dbpkg.SetEmpresaCierreCajaEstado(db, *empresaID, cierreID, "aprobado", nil, strings.TrimSpace(*usuario), "qa: aprobación"), "SetEmpresaCierreCajaEstado(aprobado)")
	log.Printf("OK caja id=%d pasó a aprobado", cierreID)

	// Intento de update tras aprobado debe bloquearse en UpdateEmpresaCierreCaja.
	err = dbpkg.UpdateEmpresaCierreCaja(db, dbpkg.EmpresaCierreCaja{
		ID:             cierreID,
		EmpresaID:      *empresaID,
		SucursalID:     0,
		CajaCodigo:     cajaCodigo,
		Turno:          "general",
		FechaOperacion: time.Now().Format("2006-01-02"),
		Moneda:         "COP",
		UsuarioCreador: strings.TrimSpace(*usuario),
		Estado:         "activo",
		Observaciones:  "qa: intento update tras aprobado",
	})
	if err == nil {
		log.Fatal("ERROR: se permitió UpdateEmpresaCierreCaja en estado aprobado (esperado: ErrCierreCajaAprobadoBloqueado)")
	}
	if !errors.Is(err, dbpkg.ErrCierreCajaAprobadoBloqueado) {
		log.Fatalf("ERROR: UpdateEmpresaCierreCaja devolvió error inesperado: %v", err)
	}
	log.Printf("OK bloqueo de UpdateEmpresaCierreCaja en aprobado validado")

	fmt.Println("RESULTADO_FINAL=OK pruebas_finanzas_motel_calipso")
}
