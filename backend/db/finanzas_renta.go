package db

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"
)

const EmpresaRentaTarifaGeneralColombia = 35.0

type EmpresaFinanzasRentaInputs struct {
	EmpresaID                  int64   `json:"empresa_id"`
	Desde                      string  `json:"desde"`
	Hasta                      string  `json:"hasta"`
	TarifaRenta                float64 `json:"tarifa_renta"`
	IngresosNoConstitutivos    float64 `json:"ingresos_no_constitutivos"`
	RentasExentas              float64 `json:"rentas_exentas"`
	DeduccionesAdicionales     float64 `json:"deducciones_adicionales"`
	DescuentosTributarios      float64 `json:"descuentos_tributarios"`
	AnticipoRenta              float64 `json:"anticipo_renta"`
	RetencionesAdicionales     float64 `json:"retenciones_adicionales"`
	SobretasaPuntos            float64 `json:"sobretasa_puntos"`
	UsarVentasPOSComoIngreso   bool    `json:"usar_ventas_pos_como_ingreso"`
	UsarMovimientosComoIngreso bool    `json:"usar_movimientos_como_ingreso"`
	UsarComprasYNominaEgreso   bool    `json:"usar_compras_y_nomina_egreso"`
	UsarMovimientosComoEgreso  bool    `json:"usar_movimientos_como_egreso"`
}

type EmpresaFinanzasRentaFuente struct {
	Nombre      string  `json:"nombre"`
	Valor       float64 `json:"valor"`
	Registros   int64   `json:"registros"`
	Usado       bool    `json:"usado"`
	Descripcion string  `json:"descripcion"`
}

type EmpresaFinanzasRentaResultado struct {
	EmpresaID               int64                        `json:"empresa_id"`
	Desde                   string                       `json:"desde"`
	Hasta                   string                       `json:"hasta"`
	GeneradoEn              string                       `json:"generado_en"`
	Moneda                  string                       `json:"moneda"`
	TarifaRenta             float64                      `json:"tarifa_renta"`
	TarifaTotal             float64                      `json:"tarifa_total"`
	IngresosBase            float64                      `json:"ingresos_base"`
	DeduccionesBase         float64                      `json:"deducciones_base"`
	IngresosNoConstitutivos float64                      `json:"ingresos_no_constitutivos"`
	RentasExentas           float64                      `json:"rentas_exentas"`
	DeduccionesAdicionales  float64                      `json:"deducciones_adicionales"`
	UtilidadAntesAjustes    float64                      `json:"utilidad_antes_ajustes"`
	RentaLiquidaEstimacion  float64                      `json:"renta_liquida_estimacion"`
	RentaLiquidaGravable    float64                      `json:"renta_liquida_gravable"`
	ImpuestoRentaEstimado   float64                      `json:"impuesto_renta_estimado"`
	DescuentosTributarios   float64                      `json:"descuentos_tributarios"`
	RetencionesDescontables float64                      `json:"retenciones_descontables"`
	AnticipoRenta           float64                      `json:"anticipo_renta"`
	SaldoEstimado           float64                      `json:"saldo_estimado"`
	MargenAntesImpuesto     float64                      `json:"margen_antes_impuesto"`
	MargenDespuesImpuesto   float64                      `json:"margen_despues_impuesto"`
	Fuentes                 []EmpresaFinanzasRentaFuente `json:"fuentes"`
	Alertas                 []string                     `json:"alertas"`
	Supuestos               []string                     `json:"supuestos"`
	Resumen                 map[string]float64           `json:"resumen"`
}

type empresaFinanzasRentaBaseDatos struct {
	EmpresaID               int64
	Desde                   string
	Hasta                   string
	Moneda                  string
	VentasPOS               float64
	VentasPOSRegistros      int64
	IngresosMovimientos     float64
	IngresosMovimientosRows int64
	EgresosMovimientos      float64
	EgresosMovimientosRows  int64
	ComprasInventario       float64
	ComprasInventarioRows   int64
	NominaDevengada         float64
	NominaLiquidacionesRows int64
	RetencionesIngresos     float64
}

func NormalizeEmpresaFinanzasRentaInputs(in EmpresaFinanzasRentaInputs) EmpresaFinanzasRentaInputs {
	in.Desde = strings.TrimSpace(in.Desde)
	in.Hasta = strings.TrimSpace(in.Hasta)
	if in.TarifaRenta <= 0 {
		in.TarifaRenta = EmpresaRentaTarifaGeneralColombia
	}
	if in.TarifaRenta > 100 {
		in.TarifaRenta = 100
	}
	if in.SobretasaPuntos < 0 {
		in.SobretasaPuntos = 0
	}
	if in.SobretasaPuntos > 50 {
		in.SobretasaPuntos = 50
	}
	in.IngresosNoConstitutivos = clampMoneyNonNegative(in.IngresosNoConstitutivos)
	in.RentasExentas = clampMoneyNonNegative(in.RentasExentas)
	in.DeduccionesAdicionales = clampMoneyNonNegative(in.DeduccionesAdicionales)
	in.DescuentosTributarios = clampMoneyNonNegative(in.DescuentosTributarios)
	in.AnticipoRenta = clampMoneyNonNegative(in.AnticipoRenta)
	in.RetencionesAdicionales = clampMoneyNonNegative(in.RetencionesAdicionales)
	if !in.UsarVentasPOSComoIngreso && !in.UsarMovimientosComoIngreso {
		in.UsarMovimientosComoIngreso = true
	}
	if !in.UsarComprasYNominaEgreso && !in.UsarMovimientosComoEgreso {
		in.UsarMovimientosComoEgreso = true
	}
	return in
}

func CalculateEmpresaFinanzasRenta(in EmpresaFinanzasRentaInputs, base empresaFinanzasRentaBaseDatos) EmpresaFinanzasRentaResultado {
	in = NormalizeEmpresaFinanzasRentaInputs(in)
	if in.EmpresaID <= 0 {
		in.EmpresaID = base.EmpresaID
	}
	if in.Desde == "" {
		in.Desde = base.Desde
	}
	if in.Hasta == "" {
		in.Hasta = base.Hasta
	}
	moneda := strings.TrimSpace(base.Moneda)
	if moneda == "" {
		moneda = "COP"
	}

	fuentes := []EmpresaFinanzasRentaFuente{
		{Nombre: "ventas_pos", Valor: roundEmpresaRentaMoney(base.VentasPOS), Registros: base.VentasPOSRegistros, Usado: in.UsarVentasPOSComoIngreso, Descripcion: "Ventas cerradas del POS/carritos"},
		{Nombre: "ingresos_financieros", Valor: roundEmpresaRentaMoney(base.IngresosMovimientos), Registros: base.IngresosMovimientosRows, Usado: in.UsarMovimientosComoIngreso, Descripcion: "Movimientos tipo ingreso en finanzas"},
		{Nombre: "egresos_financieros", Valor: roundEmpresaRentaMoney(base.EgresosMovimientos), Registros: base.EgresosMovimientosRows, Usado: in.UsarMovimientosComoEgreso, Descripcion: "Movimientos tipo egreso en finanzas"},
		{Nombre: "compras_inventario", Valor: roundEmpresaRentaMoney(base.ComprasInventario), Registros: base.ComprasInventarioRows, Usado: in.UsarComprasYNominaEgreso, Descripcion: "Entradas/compra de inventario valoradas al costo"},
		{Nombre: "nomina_devengada", Valor: roundEmpresaRentaMoney(base.NominaDevengada), Registros: base.NominaLiquidacionesRows, Usado: in.UsarComprasYNominaEgreso, Descripcion: "Nomina liquidada por devengado total"},
		{Nombre: "retenciones_ingresos", Valor: roundEmpresaRentaMoney(base.RetencionesIngresos), Registros: base.IngresosMovimientosRows, Usado: true, Descripcion: "Retenciones registradas en ingresos financieros"},
	}

	ingresosBase := 0.0
	switch {
	case in.UsarVentasPOSComoIngreso && in.UsarMovimientosComoIngreso:
		ingresosBase = math.Max(base.VentasPOS, base.IngresosMovimientos)
	case in.UsarVentasPOSComoIngreso:
		ingresosBase = base.VentasPOS
	case in.UsarMovimientosComoIngreso:
		ingresosBase = base.IngresosMovimientos
	}

	costosYNomina := base.ComprasInventario + base.NominaDevengada
	deduccionesBase := 0.0
	switch {
	case in.UsarComprasYNominaEgreso && in.UsarMovimientosComoEgreso:
		deduccionesBase = math.Max(base.EgresosMovimientos, costosYNomina)
	case in.UsarComprasYNominaEgreso:
		deduccionesBase = costosYNomina
	case in.UsarMovimientosComoEgreso:
		deduccionesBase = base.EgresosMovimientos
	}

	alertas := make([]string, 0, 4)
	if in.UsarVentasPOSComoIngreso && in.UsarMovimientosComoIngreso && base.VentasPOS > 0 && base.IngresosMovimientos > 0 {
		alertas = append(alertas, "Se detectaron ventas POS e ingresos financieros; se usa el mayor para reducir riesgo de doble conteo.")
	}
	if in.UsarComprasYNominaEgreso && in.UsarMovimientosComoEgreso && base.EgresosMovimientos > 0 && costosYNomina > 0 {
		alertas = append(alertas, "Se detectaron egresos financieros y costos/nomina; se usa el mayor para reducir riesgo de doble conteo.")
	}
	if ingresosBase <= 0 {
		alertas = append(alertas, "No hay ingresos suficientes en el rango seleccionado.")
	}
	if in.SobretasaPuntos > 0 {
		alertas = append(alertas, "Se aplico sobretasa manual; confirme que la actividad economica y periodo realmente la exigen.")
	}

	utilidad := ingresosBase - deduccionesBase
	rentaLiquida := utilidad - in.IngresosNoConstitutivos - in.RentasExentas - in.DeduccionesAdicionales
	rentaGravable := math.Max(0, rentaLiquida)
	tarifaTotal := math.Min(100, in.TarifaRenta+in.SobretasaPuntos)
	impuesto := rentaGravable * tarifaTotal / 100
	retenciones := base.RetencionesIngresos + in.RetencionesAdicionales
	saldo := impuesto - in.DescuentosTributarios - retenciones - in.AnticipoRenta

	margenAntes := 0.0
	margenDespues := 0.0
	if ingresosBase > 0 {
		margenAntes = utilidad / ingresosBase * 100
		margenDespues = (utilidad - impuesto) / ingresosBase * 100
	}

	supuestos := []string{
		"Estimacion gerencial basada en datos registrados en el sistema; no reemplaza la declaracion oficial ni revision del contador.",
		"Tarifa general editable. Para Colombia se propone 35% como tarifa general DIAN para personas juridicas en el sistema ordinario.",
		"Los ajustes manuales permiten depurar ingresos no constitutivos, rentas exentas, deducciones, descuentos, retenciones y anticipos.",
	}

	res := EmpresaFinanzasRentaResultado{
		EmpresaID:               in.EmpresaID,
		Desde:                   in.Desde,
		Hasta:                   in.Hasta,
		GeneradoEn:              time.Now().Format("2006-01-02 15:04:05"),
		Moneda:                  moneda,
		TarifaRenta:             roundEmpresaRentaMoney(in.TarifaRenta),
		TarifaTotal:             roundEmpresaRentaMoney(tarifaTotal),
		IngresosBase:            roundEmpresaRentaMoney(ingresosBase),
		DeduccionesBase:         roundEmpresaRentaMoney(deduccionesBase),
		IngresosNoConstitutivos: roundEmpresaRentaMoney(in.IngresosNoConstitutivos),
		RentasExentas:           roundEmpresaRentaMoney(in.RentasExentas),
		DeduccionesAdicionales:  roundEmpresaRentaMoney(in.DeduccionesAdicionales),
		UtilidadAntesAjustes:    roundEmpresaRentaMoney(utilidad),
		RentaLiquidaEstimacion:  roundEmpresaRentaMoney(rentaLiquida),
		RentaLiquidaGravable:    roundEmpresaRentaMoney(rentaGravable),
		ImpuestoRentaEstimado:   roundEmpresaRentaMoney(impuesto),
		DescuentosTributarios:   roundEmpresaRentaMoney(in.DescuentosTributarios),
		RetencionesDescontables: roundEmpresaRentaMoney(retenciones),
		AnticipoRenta:           roundEmpresaRentaMoney(in.AnticipoRenta),
		SaldoEstimado:           roundEmpresaRentaMoney(saldo),
		MargenAntesImpuesto:     roundEmpresaRentaMoney(margenAntes),
		MargenDespuesImpuesto:   roundEmpresaRentaMoney(margenDespues),
		Fuentes:                 fuentes,
		Alertas:                 alertas,
		Supuestos:               supuestos,
		Resumen: map[string]float64{
			"ventas_pos":              roundEmpresaRentaMoney(base.VentasPOS),
			"ingresos_financieros":    roundEmpresaRentaMoney(base.IngresosMovimientos),
			"egresos_financieros":     roundEmpresaRentaMoney(base.EgresosMovimientos),
			"compras_inventario":      roundEmpresaRentaMoney(base.ComprasInventario),
			"nomina_devengada":        roundEmpresaRentaMoney(base.NominaDevengada),
			"retenciones_registradas": roundEmpresaRentaMoney(base.RetencionesIngresos),
			"costos_y_nomina":         roundEmpresaRentaMoney(costosYNomina),
			"saldo_a_pagar_o_favor":   roundEmpresaRentaMoney(saldo),
		},
	}
	return res
}

func GetEmpresaFinanzasRentaEstimacion(dbConn *sql.DB, in EmpresaFinanzasRentaInputs) (EmpresaFinanzasRentaResultado, error) {
	if dbConn == nil {
		return EmpresaFinanzasRentaResultado{}, fmt.Errorf("db connection is nil")
	}
	in = NormalizeEmpresaFinanzasRentaInputs(in)
	if in.EmpresaID <= 0 {
		return EmpresaFinanzasRentaResultado{}, fmt.Errorf("empresa_id es obligatorio")
	}
	if in.Desde == "" || in.Hasta == "" {
		now := time.Now()
		if in.Desde == "" {
			in.Desde = time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
		}
		if in.Hasta == "" {
			in.Hasta = now.Format("2006-01-02")
		}
	}
	if err := EnsureEmpresaCarritosSchema(dbConn); err != nil {
		return EmpresaFinanzasRentaResultado{}, err
	}
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		return EmpresaFinanzasRentaResultado{}, err
	}
	if err := EnsureEmpresaFinanzasSchema(dbConn); err != nil {
		return EmpresaFinanzasRentaResultado{}, err
	}
	if err := EnsureEmpresaNominaSchema(dbConn); err != nil {
		return EmpresaFinanzasRentaResultado{}, err
	}

	base := empresaFinanzasRentaBaseDatos{EmpresaID: in.EmpresaID, Desde: in.Desde, Hasta: in.Hasta, Moneda: "COP"}
	if cfg, err := GetEmpresaFinanzasConfiguracion(dbConn, in.EmpresaID); err == nil && cfg != nil && strings.TrimSpace(cfg.Moneda) != "" {
		base.Moneda = strings.ToUpper(strings.TrimSpace(cfg.Moneda))
	}

	ventasCond, ventasArgs := buildDateRangeCondition("COALESCE(c.pagado_en, c.fecha_actualizacion, c.fecha_creacion)", in.Desde, in.Hasta)
	ventasParams := append([]interface{}{in.EmpresaID}, ventasArgs...)
	if err := queryRowSQLCompat(dbConn, `SELECT
		COALESCE(COUNT(1), 0),
		COALESCE(SUM(CASE WHEN COALESCE(c.total_pagado, 0) > 0 THEN COALESCE(c.total_pagado, 0) ELSE COALESCE(c.total, 0) END), 0)
	FROM carritos_compras c
	WHERE c.empresa_id = ?
		AND LOWER(COALESCE(c.estado_carrito, '')) = 'cerrado'
		AND LOWER(COALESCE(c.estado, 'activo')) = 'activo'`+ventasCond, ventasParams...).Scan(&base.VentasPOSRegistros, &base.VentasPOS); err != nil {
		return EmpresaFinanzasRentaResultado{}, err
	}

	finanzasCond, finanzasArgs := buildDateRangeCondition("m.fecha_movimiento", in.Desde, in.Hasta)
	finanzasParams := append([]interface{}{in.EmpresaID}, finanzasArgs...)
	if err := queryRowSQLCompat(dbConn, `SELECT
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(m.tipo_movimiento, '')) = 'ingreso' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(m.tipo_movimiento, '')) = 'egreso' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(m.tipo_movimiento, '')) = 'ingreso' THEN COALESCE(NULLIF(m.total_neto, 0), NULLIF(m.total, 0), m.monto, 0) ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(m.tipo_movimiento, '')) = 'egreso' THEN COALESCE(NULLIF(m.total_neto, 0), NULLIF(m.total, 0), m.monto, 0) ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(m.tipo_movimiento, '')) = 'ingreso' THEN COALESCE(m.retencion_fuente,0)+COALESCE(m.retencion_ica,0)+COALESCE(m.retencion_iva,0)+COALESCE(m.total_retenciones,0) ELSE 0 END), 0)
	FROM empresa_finanzas_movimientos m
	WHERE m.empresa_id = ?
		AND LOWER(COALESCE(m.estado, 'activo')) = 'activo'
		AND LOWER(COALESCE(m.tipo_movimiento, '')) IN ('ingreso', 'egreso')`+finanzasCond, finanzasParams...).Scan(
		&base.IngresosMovimientosRows,
		&base.EgresosMovimientosRows,
		&base.IngresosMovimientos,
		&base.EgresosMovimientos,
		&base.RetencionesIngresos,
	); err != nil {
		return EmpresaFinanzasRentaResultado{}, err
	}

	comprasCond, comprasArgs := buildDateRangeCondition("m.fecha_movimiento", in.Desde, in.Hasta)
	comprasParams := append([]interface{}{in.EmpresaID}, comprasArgs...)
	if err := queryRowSQLCompat(dbConn, `SELECT
		COALESCE(COUNT(1), 0),
		COALESCE(SUM(COALESCE(m.cantidad, 0) * COALESCE(m.costo_unitario, 0)), 0)
	FROM inventario_movimientos m
	WHERE m.empresa_id = ?
		AND LOWER(COALESCE(m.estado, 'activo')) = 'activo'
		AND LOWER(COALESCE(m.tipo, '')) IN ('entrada', 'ajuste_entrada', 'ajuste_positivo', 'compra')`+comprasCond, comprasParams...).Scan(&base.ComprasInventarioRows, &base.ComprasInventario); err != nil {
		return EmpresaFinanzasRentaResultado{}, err
	}

	nominaCond, nominaArgs := buildDateRangeCondition("l.periodo_hasta", in.Desde, in.Hasta)
	nominaParams := append([]interface{}{in.EmpresaID}, nominaArgs...)
	if err := queryRowSQLCompat(dbConn, `SELECT
		COALESCE(COUNT(1), 0),
		COALESCE(SUM(COALESCE(l.devengado_total, 0)), 0)
	FROM empresa_nomina_liquidaciones l
	WHERE l.empresa_id = ?
		AND LOWER(COALESCE(l.estado, 'activo')) = 'activo'`+nominaCond, nominaParams...).Scan(&base.NominaLiquidacionesRows, &base.NominaDevengada); err != nil {
		return EmpresaFinanzasRentaResultado{}, err
	}

	return CalculateEmpresaFinanzasRenta(in, base), nil
}

func clampMoneyNonNegative(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
		return 0
	}
	return v
}

func roundEmpresaRentaMoney(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return math.Round(v*100) / 100
}
