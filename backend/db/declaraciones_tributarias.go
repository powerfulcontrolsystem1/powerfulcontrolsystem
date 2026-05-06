package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type EmpresaDeclaracionTributaria struct {
	ID                      int64   `json:"id"`
	EmpresaID               int64   `json:"empresa_id"`
	TipoDeclaracion         string  `json:"tipo_declaracion"`
	Periodo                 string  `json:"periodo"`
	Anio                    int     `json:"anio"`
	Periodicidad            string  `json:"periodicidad"`
	FechaDesde              string  `json:"fecha_desde"`
	FechaHasta              string  `json:"fecha_hasta"`
	FechaVencimiento        string  `json:"fecha_vencimiento"`
	NIT                     string  `json:"nit,omitempty"`
	Municipio               string  `json:"municipio,omitempty"`
	Formulario              string  `json:"formulario,omitempty"`
	IngresosGravados        float64 `json:"ingresos_gravados"`
	IngresosExcluidos       float64 `json:"ingresos_excluidos"`
	ComprasGravadas         float64 `json:"compras_gravadas"`
	IVAGenerado             float64 `json:"iva_generado"`
	IVADescontable          float64 `json:"iva_descontable"`
	ImpuestoConsumo         float64 `json:"impuesto_consumo"`
	RetencionFuente         float64 `json:"retencion_fuente"`
	RetencionIVA            float64 `json:"retencion_iva"`
	RetencionICA            float64 `json:"retencion_ica"`
	Autorretencion          float64 `json:"autorretencion"`
	Anticipo                float64 `json:"anticipo"`
	Sanciones               float64 `json:"sanciones"`
	Intereses               float64 `json:"intereses"`
	SaldoFavorAnterior      float64 `json:"saldo_favor_anterior"`
	SaldoPagar              float64 `json:"saldo_pagar"`
	SaldoFavor              float64 `json:"saldo_favor"`
	Estado                  string  `json:"estado"`
	FechaPresentacion       string  `json:"fecha_presentacion,omitempty"`
	FechaPago               string  `json:"fecha_pago,omitempty"`
	SoporteURL              string  `json:"soporte_url,omitempty"`
	ReciboPagoURL           string  `json:"recibo_pago_url,omitempty"`
	ConciliacionObservacion string  `json:"conciliacion_observacion,omitempty"`
	Observaciones           string  `json:"observaciones,omitempty"`
	FechaCreacion           string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion      string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador          string  `json:"usuario_creador,omitempty"`
}

type EmpresaDeclaracionTributariaMovimiento struct {
	ID              int64   `json:"id"`
	EmpresaID       int64   `json:"empresa_id"`
	DeclaracionID   int64   `json:"declaracion_id"`
	TipoDeclaracion string  `json:"tipo_declaracion"`
	Periodo         string  `json:"periodo"`
	OrigenModulo    string  `json:"origen_modulo"`
	Referencia      string  `json:"referencia"`
	FechaMovimiento string  `json:"fecha_movimiento"`
	TerceroNombre   string  `json:"tercero_nombre,omitempty"`
	Concepto        string  `json:"concepto"`
	BaseValor       float64 `json:"base_valor"`
	ImpuestoValor   float64 `json:"impuesto_valor"`
	RetencionValor  float64 `json:"retencion_valor"`
	Naturaleza      string  `json:"naturaleza"`
	Estado          string  `json:"estado"`
	FechaCreacion   string  `json:"fecha_creacion,omitempty"`
}

type EmpresaCalendarioTributario struct {
	ID               int64  `json:"id"`
	EmpresaID        int64  `json:"empresa_id"`
	TipoDeclaracion  string `json:"tipo_declaracion"`
	Anio             int    `json:"anio"`
	Periodo          string `json:"periodo"`
	Periodicidad     string `json:"periodicidad"`
	FechaDesde       string `json:"fecha_desde"`
	FechaHasta       string `json:"fecha_hasta"`
	FechaVencimiento string `json:"fecha_vencimiento"`
	DigitoNITDesde   int    `json:"digito_nit_desde"`
	DigitoNITHasta   int    `json:"digito_nit_hasta"`
	Estado           string `json:"estado"`
	Observaciones    string `json:"observaciones,omitempty"`
	UsuarioCreador   string `json:"usuario_creador,omitempty"`
	FechaCreacion    string `json:"fecha_creacion,omitempty"`
}

type EmpresaDeclaracionesTributariasDashboard struct {
	EmpresaID              int64                                    `json:"empresa_id"`
	PeriodoActual          string                                   `json:"periodo_actual"`
	DeclaracionesBorrador  int                                      `json:"declaraciones_borrador"`
	DeclaracionesRevisadas int                                      `json:"declaraciones_revisadas"`
	DeclaracionesVencidas  int                                      `json:"declaraciones_vencidas"`
	DeclaracionesPagadas   int                                      `json:"declaraciones_pagadas"`
	SaldoPagarTotal        float64                                  `json:"saldo_pagar_total"`
	RetencionesTotal       float64                                  `json:"retenciones_total"`
	ProximosVencimientos   []EmpresaCalendarioTributario            `json:"proximos_vencimientos"`
	Declaraciones          []EmpresaDeclaracionTributaria           `json:"declaraciones"`
	MovimientosRecientes   []EmpresaDeclaracionTributariaMovimiento `json:"movimientos_recientes"`
	Alertas                []string                                 `json:"alertas"`
}

func EnsureEmpresaDeclaracionesTributariasSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_declaraciones_tributarias (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			tipo_declaracion TEXT NOT NULL,
			periodo TEXT NOT NULL,
			anio INTEGER DEFAULT 0,
			periodicidad TEXT DEFAULT 'mensual',
			fecha_desde TEXT NOT NULL,
			fecha_hasta TEXT NOT NULL,
			fecha_vencimiento TEXT,
			nit TEXT,
			municipio TEXT,
			formulario TEXT,
			ingresos_gravados REAL DEFAULT 0,
			ingresos_excluidos REAL DEFAULT 0,
			compras_gravadas REAL DEFAULT 0,
			iva_generado REAL DEFAULT 0,
			iva_descontable REAL DEFAULT 0,
			impuesto_consumo REAL DEFAULT 0,
			retencion_fuente REAL DEFAULT 0,
			retencion_iva REAL DEFAULT 0,
			retencion_ica REAL DEFAULT 0,
			autorretencion REAL DEFAULT 0,
			anticipo REAL DEFAULT 0,
			sanciones REAL DEFAULT 0,
			intereses REAL DEFAULT 0,
			saldo_favor_anterior REAL DEFAULT 0,
			saldo_pagar REAL DEFAULT 0,
			saldo_favor REAL DEFAULT 0,
			estado TEXT DEFAULT 'borrador',
			fecha_presentacion TEXT,
			fecha_pago TEXT,
			soporte_url TEXT,
			recibo_pago_url TEXT,
			conciliacion_observacion TEXT,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			UNIQUE(empresa_id,tipo_declaracion,periodo,municipio)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_declaraciones_tributarias_empresa ON empresa_declaraciones_tributarias(empresa_id,tipo_declaracion,estado,fecha_vencimiento)`,
		`CREATE TABLE IF NOT EXISTS empresa_declaraciones_tributarias_movimientos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			declaracion_id INTEGER DEFAULT 0,
			tipo_declaracion TEXT NOT NULL,
			periodo TEXT NOT NULL,
			origen_modulo TEXT DEFAULT 'manual',
			referencia TEXT,
			fecha_movimiento TEXT,
			tercero_nombre TEXT,
			concepto TEXT,
			base_valor REAL DEFAULT 0,
			impuesto_valor REAL DEFAULT 0,
			retencion_valor REAL DEFAULT 0,
			naturaleza TEXT DEFAULT 'debito',
			estado TEXT DEFAULT 'incluido',
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS ix_declaraciones_movimientos_empresa ON empresa_declaraciones_tributarias_movimientos(empresa_id,tipo_declaracion,periodo,declaracion_id)`,
		`CREATE TABLE IF NOT EXISTS empresa_calendario_tributario (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			tipo_declaracion TEXT NOT NULL,
			anio INTEGER NOT NULL,
			periodo TEXT NOT NULL,
			periodicidad TEXT DEFAULT 'mensual',
			fecha_desde TEXT NOT NULL,
			fecha_hasta TEXT NOT NULL,
			fecha_vencimiento TEXT NOT NULL,
			digito_nit_desde INTEGER DEFAULT 0,
			digito_nit_hasta INTEGER DEFAULT 9,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			usuario_creador TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(empresa_id,tipo_declaracion,periodo,digito_nit_desde,digito_nit_hasta)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_calendario_tributario_empresa ON empresa_calendario_tributario(empresa_id,anio,fecha_vencimiento,tipo_declaracion)`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func BuildEmpresaDeclaracionesTributariasDashboard(dbConn *sql.DB, empresaID int64) (EmpresaDeclaracionesTributariasDashboard, error) {
	if err := EnsureEmpresaDeclaracionesTributariasSchema(dbConn); err != nil {
		return EmpresaDeclaracionesTributariasDashboard{}, err
	}
	periodo := time.Now().Format("2006-01")
	d := EmpresaDeclaracionesTributariasDashboard{EmpresaID: empresaID, PeriodoActual: periodo}
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_declaraciones_tributarias WHERE empresa_id=? AND estado='borrador'`, empresaID).Scan(&d.DeclaracionesBorrador)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_declaraciones_tributarias WHERE empresa_id=? AND estado='revisada'`, empresaID).Scan(&d.DeclaracionesRevisadas)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_declaraciones_tributarias WHERE empresa_id=? AND estado='pagada'`, empresaID).Scan(&d.DeclaracionesPagadas)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_declaraciones_tributarias WHERE empresa_id=? AND estado NOT IN ('pagada','anulada') AND COALESCE(fecha_vencimiento,'')<>'' AND fecha_vencimiento<?`, empresaID, time.Now().Format("2006-01-02")).Scan(&d.DeclaracionesVencidas)
	_ = QueryRowCompat(dbConn, `SELECT COALESCE(SUM(saldo_pagar),0),COALESCE(SUM(retencion_fuente+retencion_iva+retencion_ica),0) FROM empresa_declaraciones_tributarias WHERE empresa_id=? AND estado NOT IN ('anulada')`, empresaID).Scan(&d.SaldoPagarTotal, &d.RetencionesTotal)
	d.ProximosVencimientos, _ = ListEmpresaCalendarioTributario(dbConn, empresaID, time.Now().Year(), 12)
	d.Declaraciones, _ = ListEmpresaDeclaracionesTributarias(dbConn, empresaID, "", "", 100)
	d.MovimientosRecientes, _ = ListEmpresaDeclaracionesTributariasMovimientos(dbConn, empresaID, "", "", 50)
	if d.DeclaracionesVencidas > 0 {
		d.Alertas = append(d.Alertas, "Hay declaraciones vencidas pendientes de revisar o pagar.")
	}
	if d.DeclaracionesBorrador > 0 {
		d.Alertas = append(d.Alertas, "Hay declaraciones en borrador pendientes de aprobacion.")
	}
	if len(d.ProximosVencimientos) == 0 {
		d.Alertas = append(d.Alertas, "No hay calendario tributario cargado para el ano actual.")
	}
	if len(d.Alertas) == 0 {
		d.Alertas = append(d.Alertas, "Declaraciones tributarias sin alertas criticas.")
	}
	return d, nil
}

func UpsertEmpresaDeclaracionTributaria(dbConn *sql.DB, row EmpresaDeclaracionTributaria) (int64, error) {
	if err := EnsureEmpresaDeclaracionesTributariasSchema(dbConn); err != nil {
		return 0, err
	}
	row = normalizeDeclaracionTributaria(row)
	if row.EmpresaID <= 0 || row.TipoDeclaracion == "" || row.Periodo == "" || row.FechaDesde == "" || row.FechaHasta == "" {
		return 0, errors.New("empresa_id, tipo_declaracion, periodo, fecha_desde y fecha_hasta son requeridos")
	}
	if row.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_declaraciones_tributarias SET tipo_declaracion=?,periodo=?,anio=?,periodicidad=?,fecha_desde=?,fecha_hasta=?,fecha_vencimiento=?,nit=?,municipio=?,formulario=?,ingresos_gravados=?,ingresos_excluidos=?,compras_gravadas=?,iva_generado=?,iva_descontable=?,impuesto_consumo=?,retencion_fuente=?,retencion_iva=?,retencion_ica=?,autorretencion=?,anticipo=?,sanciones=?,intereses=?,saldo_favor_anterior=?,saldo_pagar=?,saldo_favor=?,estado=?,fecha_presentacion=?,fecha_pago=?,soporte_url=?,recibo_pago_url=?,conciliacion_observacion=?,observaciones=?,fecha_actualizacion=CURRENT_TIMESTAMP,usuario_creador=? WHERE empresa_id=? AND id=?`,
			row.TipoDeclaracion, row.Periodo, row.Anio, row.Periodicidad, row.FechaDesde, row.FechaHasta, row.FechaVencimiento, row.NIT, row.Municipio, row.Formulario, row.IngresosGravados, row.IngresosExcluidos, row.ComprasGravadas, row.IVAGenerado, row.IVADescontable, row.ImpuestoConsumo, row.RetencionFuente, row.RetencionIVA, row.RetencionICA, row.Autorretencion, row.Anticipo, row.Sanciones, row.Intereses, row.SaldoFavorAnterior, row.SaldoPagar, row.SaldoFavor, row.Estado, row.FechaPresentacion, row.FechaPago, row.SoporteURL, row.ReciboPagoURL, row.ConciliacionObservacion, row.Observaciones, row.UsuarioCreador, row.EmpresaID, row.ID)
		return row.ID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_declaraciones_tributarias (empresa_id,tipo_declaracion,periodo,anio,periodicidad,fecha_desde,fecha_hasta,fecha_vencimiento,nit,municipio,formulario,ingresos_gravados,ingresos_excluidos,compras_gravadas,iva_generado,iva_descontable,impuesto_consumo,retencion_fuente,retencion_iva,retencion_ica,autorretencion,anticipo,sanciones,intereses,saldo_favor_anterior,saldo_pagar,saldo_favor,estado,fecha_presentacion,fecha_pago,soporte_url,recibo_pago_url,conciliacion_observacion,observaciones,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (empresa_id,tipo_declaracion,periodo,municipio) DO UPDATE SET anio=EXCLUDED.anio,periodicidad=EXCLUDED.periodicidad,fecha_desde=EXCLUDED.fecha_desde,fecha_hasta=EXCLUDED.fecha_hasta,fecha_vencimiento=EXCLUDED.fecha_vencimiento,nit=EXCLUDED.nit,formulario=EXCLUDED.formulario,ingresos_gravados=EXCLUDED.ingresos_gravados,ingresos_excluidos=EXCLUDED.ingresos_excluidos,compras_gravadas=EXCLUDED.compras_gravadas,iva_generado=EXCLUDED.iva_generado,iva_descontable=EXCLUDED.iva_descontable,impuesto_consumo=EXCLUDED.impuesto_consumo,retencion_fuente=EXCLUDED.retencion_fuente,retencion_iva=EXCLUDED.retencion_iva,retencion_ica=EXCLUDED.retencion_ica,autorretencion=EXCLUDED.autorretencion,anticipo=EXCLUDED.anticipo,sanciones=EXCLUDED.sanciones,intereses=EXCLUDED.intereses,saldo_favor_anterior=EXCLUDED.saldo_favor_anterior,saldo_pagar=EXCLUDED.saldo_pagar,saldo_favor=EXCLUDED.saldo_favor,estado=EXCLUDED.estado,fecha_presentacion=EXCLUDED.fecha_presentacion,fecha_pago=EXCLUDED.fecha_pago,soporte_url=EXCLUDED.soporte_url,recibo_pago_url=EXCLUDED.recibo_pago_url,conciliacion_observacion=EXCLUDED.conciliacion_observacion,observaciones=EXCLUDED.observaciones,fecha_actualizacion=CURRENT_TIMESTAMP,usuario_creador=EXCLUDED.usuario_creador`,
		row.EmpresaID, row.TipoDeclaracion, row.Periodo, row.Anio, row.Periodicidad, row.FechaDesde, row.FechaHasta, row.FechaVencimiento, row.NIT, row.Municipio, row.Formulario, row.IngresosGravados, row.IngresosExcluidos, row.ComprasGravadas, row.IVAGenerado, row.IVADescontable, row.ImpuestoConsumo, row.RetencionFuente, row.RetencionIVA, row.RetencionICA, row.Autorretencion, row.Anticipo, row.Sanciones, row.Intereses, row.SaldoFavorAnterior, row.SaldoPagar, row.SaldoFavor, row.Estado, row.FechaPresentacion, row.FechaPago, row.SoporteURL, row.ReciboPagoURL, row.ConciliacionObservacion, row.Observaciones, row.UsuarioCreador)
}

func PreliquidarEmpresaDeclaracionTributaria(dbConn *sql.DB, empresaID int64, tipo, periodo, usuario string) (EmpresaDeclaracionTributaria, error) {
	if err := EnsureEmpresaDeclaracionesTributariasSchema(dbConn); err != nil {
		return EmpresaDeclaracionTributaria{}, err
	}
	tipo = normalizeDeclaracionTipo(tipo)
	periodo = normalizeDeclaracionPeriodo(periodo)
	if empresaID <= 0 || periodo == "" {
		return EmpresaDeclaracionTributaria{}, errors.New("empresa_id y periodo son requeridos")
	}
	desde, hasta := periodoRangeDeclaracion(periodo)
	row := EmpresaDeclaracionTributaria{
		EmpresaID:        empresaID,
		TipoDeclaracion:  tipo,
		Periodo:          periodo,
		Anio:             yearFromPeriodo(periodo),
		Periodicidad:     periodicidadDeclaracion(tipo),
		FechaDesde:       desde,
		FechaHasta:       hasta,
		FechaVencimiento: defaultVencimientoDeclaracion(periodo, tipo),
		Formulario:       formularioDeclaracion(tipo),
		Estado:           "borrador",
		UsuarioCreador:   usuario,
	}
	fillDeclaracionFromSources(dbConn, &row)
	row = calcularSaldosDeclaracion(row)
	id, err := UpsertEmpresaDeclaracionTributaria(dbConn, row)
	if err != nil {
		return EmpresaDeclaracionTributaria{}, err
	}
	row.ID = id
	_ = regenerarMovimientosDeclaracion(dbConn, row)
	return row, nil
}

func ListEmpresaDeclaracionesTributarias(dbConn *sql.DB, empresaID int64, tipo, estado string, limit int) ([]EmpresaDeclaracionTributaria, error) {
	if err := EnsureEmpresaDeclaracionesTributariasSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 1000 {
		limit = 300
	}
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if t := normalizeDeclaracionTipo(tipo); strings.TrimSpace(tipo) != "" && t != "todos" {
		where += " AND tipo_declaracion=?"
		args = append(args, t)
	}
	if e := normalizeDeclaracionEstado(estado); strings.TrimSpace(estado) != "" && e != "todos" {
		where += " AND estado=?"
		args = append(args, e)
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,tipo_declaracion,periodo,COALESCE(anio,0),COALESCE(periodicidad,'mensual'),fecha_desde,fecha_hasta,COALESCE(fecha_vencimiento,''),COALESCE(nit,''),COALESCE(municipio,''),COALESCE(formulario,''),COALESCE(ingresos_gravados,0),COALESCE(ingresos_excluidos,0),COALESCE(compras_gravadas,0),COALESCE(iva_generado,0),COALESCE(iva_descontable,0),COALESCE(impuesto_consumo,0),COALESCE(retencion_fuente,0),COALESCE(retencion_iva,0),COALESCE(retencion_ica,0),COALESCE(autorretencion,0),COALESCE(anticipo,0),COALESCE(sanciones,0),COALESCE(intereses,0),COALESCE(saldo_favor_anterior,0),COALESCE(saldo_pagar,0),COALESCE(saldo_favor,0),COALESCE(estado,'borrador'),COALESCE(fecha_presentacion,''),COALESCE(fecha_pago,''),COALESCE(soporte_url,''),COALESCE(recibo_pago_url,''),COALESCE(conciliacion_observacion,''),COALESCE(observaciones,''),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,''),COALESCE(usuario_creador,'') FROM empresa_declaraciones_tributarias WHERE %s ORDER BY periodo DESC,id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaDeclaracionTributaria{}
	for rows.Next() {
		var x EmpresaDeclaracionTributaria
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.TipoDeclaracion, &x.Periodo, &x.Anio, &x.Periodicidad, &x.FechaDesde, &x.FechaHasta, &x.FechaVencimiento, &x.NIT, &x.Municipio, &x.Formulario, &x.IngresosGravados, &x.IngresosExcluidos, &x.ComprasGravadas, &x.IVAGenerado, &x.IVADescontable, &x.ImpuestoConsumo, &x.RetencionFuente, &x.RetencionIVA, &x.RetencionICA, &x.Autorretencion, &x.Anticipo, &x.Sanciones, &x.Intereses, &x.SaldoFavorAnterior, &x.SaldoPagar, &x.SaldoFavor, &x.Estado, &x.FechaPresentacion, &x.FechaPago, &x.SoporteURL, &x.ReciboPagoURL, &x.ConciliacionObservacion, &x.Observaciones, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func UpsertEmpresaCalendarioTributario(dbConn *sql.DB, row EmpresaCalendarioTributario) (int64, error) {
	if err := EnsureEmpresaDeclaracionesTributariasSchema(dbConn); err != nil {
		return 0, err
	}
	row.TipoDeclaracion = normalizeDeclaracionTipo(row.TipoDeclaracion)
	row.Periodicidad = normalizeDeclaracionPeriodicidad(row.Periodicidad)
	row.Estado = normalizeDeclaracionActivoEstado(row.Estado)
	if row.EmpresaID <= 0 || row.Anio <= 0 || row.Periodo == "" || row.FechaDesde == "" || row.FechaHasta == "" || row.FechaVencimiento == "" {
		return 0, errors.New("empresa_id, anio, periodo y fechas son requeridos")
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_calendario_tributario (empresa_id,tipo_declaracion,anio,periodo,periodicidad,fecha_desde,fecha_hasta,fecha_vencimiento,digito_nit_desde,digito_nit_hasta,estado,observaciones,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (empresa_id,tipo_declaracion,periodo,digito_nit_desde,digito_nit_hasta) DO UPDATE SET periodicidad=EXCLUDED.periodicidad,fecha_desde=EXCLUDED.fecha_desde,fecha_hasta=EXCLUDED.fecha_hasta,fecha_vencimiento=EXCLUDED.fecha_vencimiento,estado=EXCLUDED.estado,observaciones=EXCLUDED.observaciones,usuario_creador=EXCLUDED.usuario_creador`,
		row.EmpresaID, row.TipoDeclaracion, row.Anio, row.Periodo, row.Periodicidad, row.FechaDesde, row.FechaHasta, row.FechaVencimiento, row.DigitoNITDesde, row.DigitoNITHasta, row.Estado, row.Observaciones, row.UsuarioCreador)
}

func ListEmpresaCalendarioTributario(dbConn *sql.DB, empresaID int64, anio int, limit int) ([]EmpresaCalendarioTributario, error) {
	if err := EnsureEmpresaDeclaracionesTributariasSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if anio <= 0 {
		anio = time.Now().Year()
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,tipo_declaracion,anio,periodo,COALESCE(periodicidad,'mensual'),fecha_desde,fecha_hasta,fecha_vencimiento,COALESCE(digito_nit_desde,0),COALESCE(digito_nit_hasta,9),COALESCE(estado,'activo'),COALESCE(observaciones,''),COALESCE(usuario_creador,''),COALESCE(fecha_creacion,'') FROM empresa_calendario_tributario WHERE empresa_id=? AND anio=? ORDER BY fecha_vencimiento LIMIT %d`, limit), empresaID, anio)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCalendarioTributario{}
	for rows.Next() {
		var x EmpresaCalendarioTributario
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.TipoDeclaracion, &x.Anio, &x.Periodo, &x.Periodicidad, &x.FechaDesde, &x.FechaHasta, &x.FechaVencimiento, &x.DigitoNITDesde, &x.DigitoNITHasta, &x.Estado, &x.Observaciones, &x.UsuarioCreador, &x.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func ListEmpresaDeclaracionesTributariasMovimientos(dbConn *sql.DB, empresaID int64, tipo, periodo string, limit int) ([]EmpresaDeclaracionTributariaMovimiento, error) {
	if err := EnsureEmpresaDeclaracionesTributariasSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 1000 {
		limit = 300
	}
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if tipo = normalizeDeclaracionTipo(tipo); strings.TrimSpace(tipo) != "" && tipo != "todos" {
		where += " AND tipo_declaracion=?"
		args = append(args, tipo)
	}
	if periodo = normalizeDeclaracionPeriodo(periodo); periodo != "" {
		where += " AND periodo=?"
		args = append(args, periodo)
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,COALESCE(declaracion_id,0),tipo_declaracion,periodo,COALESCE(origen_modulo,''),COALESCE(referencia,''),COALESCE(fecha_movimiento,''),COALESCE(tercero_nombre,''),COALESCE(concepto,''),COALESCE(base_valor,0),COALESCE(impuesto_valor,0),COALESCE(retencion_valor,0),COALESCE(naturaleza,'debito'),COALESCE(estado,'incluido'),COALESCE(fecha_creacion,'') FROM empresa_declaraciones_tributarias_movimientos WHERE %s ORDER BY fecha_movimiento DESC,id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaDeclaracionTributariaMovimiento{}
	for rows.Next() {
		var x EmpresaDeclaracionTributariaMovimiento
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.DeclaracionID, &x.TipoDeclaracion, &x.Periodo, &x.OrigenModulo, &x.Referencia, &x.FechaMovimiento, &x.TerceroNombre, &x.Concepto, &x.BaseValor, &x.ImpuestoValor, &x.RetencionValor, &x.Naturaleza, &x.Estado, &x.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func SeedEmpresaDeclaracionesTributariasDemo(dbConn *sql.DB, empresaID int64, usuario string) error {
	if err := EnsureEmpresaDeclaracionesTributariasSchema(dbConn); err != nil {
		return err
	}
	anio := time.Now().Year()
	for _, tipo := range []string{"iva", "retencion_fuente", "ica"} {
		for m := 1; m <= 3; m++ {
			periodo := fmt.Sprintf("%d-%02d", anio, m)
			desde, hasta := periodoRangeDeclaracion(periodo)
			_, _ = UpsertEmpresaCalendarioTributario(dbConn, EmpresaCalendarioTributario{EmpresaID: empresaID, TipoDeclaracion: tipo, Anio: anio, Periodo: periodo, Periodicidad: periodicidadDeclaracion(tipo), FechaDesde: desde, FechaHasta: hasta, FechaVencimiento: defaultVencimientoDeclaracion(periodo, tipo), DigitoNITDesde: 0, DigitoNITHasta: 9, Estado: "activo", Observaciones: "Calendario demo", UsuarioCreador: usuario})
		}
	}
	_, err := PreliquidarEmpresaDeclaracionTributaria(dbConn, empresaID, "iva", time.Now().Format("2006-01"), usuario)
	return err
}

func fillDeclaracionFromSources(dbConn *sql.DB, row *EmpresaDeclaracionTributaria) {
	if row == nil {
		return
	}
	dash, err := EmpresaImpuestosDashboardData(dbConn, row.EmpresaID, row.FechaDesde, row.FechaHasta)
	if err == nil {
		row.IngresosGravados = dash.Resumen.BaseGravable
		row.IVAGenerado = dash.Resumen.ImpuestoGenerado
	}
	_ = QueryRowCompat(dbConn, `SELECT COALESCE(SUM(subtotal),0),COALESCE(SUM(iva),0),COALESCE(SUM(retenciones),0) FROM empresa_contabilidad_documentos_soporte WHERE empresa_id=? AND fecha_documento>=? AND fecha_documento<=?`, row.EmpresaID, row.FechaDesde, row.FechaHasta).Scan(&row.ComprasGravadas, &row.IVADescontable, &row.RetencionFuente)
	var reteIVA, reteICA float64
	_ = QueryRowCompat(dbConn, `SELECT COALESCE(SUM(CASE WHEN impuesto_codigo IN ('RETEIVA','RET_IVA') THEN credito+debito ELSE 0 END),0), COALESCE(SUM(CASE WHEN impuesto_codigo IN ('RETEICA','ICA') THEN credito+debito ELSE 0 END),0) FROM empresa_contabilidad_colombia_lineas WHERE empresa_id=?`, row.EmpresaID).Scan(&reteIVA, &reteICA)
	row.RetencionIVA += reteIVA
	row.RetencionICA += reteICA
}

func calcularSaldosDeclaracion(row EmpresaDeclaracionTributaria) EmpresaDeclaracionTributaria {
	base := 0.0
	switch row.TipoDeclaracion {
	case "iva":
		base = row.IVAGenerado - row.IVADescontable - row.SaldoFavorAnterior
	case "retencion_fuente":
		base = row.RetencionFuente + row.Autorretencion
	case "reteiva":
		base = row.RetencionIVA
	case "reteica", "ica":
		base = row.RetencionICA
	case "consumo":
		base = row.ImpuestoConsumo
	case "regimen_simple", "renta":
		base = (row.IngresosGravados * 0.03) + row.Anticipo - row.SaldoFavorAnterior
	default:
		base = row.IVAGenerado + row.RetencionFuente + row.RetencionIVA + row.RetencionICA - row.IVADescontable - row.SaldoFavorAnterior
	}
	base += row.Sanciones + row.Intereses
	if base >= 0 {
		row.SaldoPagar = roundDeclaracion(base)
		row.SaldoFavor = 0
	} else {
		row.SaldoPagar = 0
		row.SaldoFavor = roundDeclaracion(-base)
	}
	return row
}

func regenerarMovimientosDeclaracion(dbConn *sql.DB, row EmpresaDeclaracionTributaria) error {
	_, _ = ExecCompat(dbConn, `DELETE FROM empresa_declaraciones_tributarias_movimientos WHERE empresa_id=? AND declaracion_id=?`, row.EmpresaID, row.ID)
	movs := []EmpresaDeclaracionTributariaMovimiento{
		{EmpresaID: row.EmpresaID, DeclaracionID: row.ID, TipoDeclaracion: row.TipoDeclaracion, Periodo: row.Periodo, OrigenModulo: "ventas", Referencia: "ventas_periodo", FechaMovimiento: row.FechaHasta, Concepto: "Ingresos gravados e IVA generado", BaseValor: row.IngresosGravados, ImpuestoValor: row.IVAGenerado, Naturaleza: "credito", Estado: "incluido"},
		{EmpresaID: row.EmpresaID, DeclaracionID: row.ID, TipoDeclaracion: row.TipoDeclaracion, Periodo: row.Periodo, OrigenModulo: "compras", Referencia: "documentos_soporte", FechaMovimiento: row.FechaHasta, Concepto: "Compras gravadas e IVA descontable", BaseValor: row.ComprasGravadas, ImpuestoValor: row.IVADescontable, Naturaleza: "debito", Estado: "incluido"},
		{EmpresaID: row.EmpresaID, DeclaracionID: row.ID, TipoDeclaracion: row.TipoDeclaracion, Periodo: row.Periodo, OrigenModulo: "retenciones", Referencia: "contabilidad", FechaMovimiento: row.FechaHasta, Concepto: "Retenciones del periodo", RetencionValor: row.RetencionFuente + row.RetencionIVA + row.RetencionICA, Naturaleza: "credito", Estado: "incluido"},
	}
	for _, m := range movs {
		if _, err := insertSQLCompat(dbConn, `INSERT INTO empresa_declaraciones_tributarias_movimientos (empresa_id,declaracion_id,tipo_declaracion,periodo,origen_modulo,referencia,fecha_movimiento,tercero_nombre,concepto,base_valor,impuesto_valor,retencion_valor,naturaleza,estado) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, m.EmpresaID, m.DeclaracionID, m.TipoDeclaracion, m.Periodo, m.OrigenModulo, m.Referencia, m.FechaMovimiento, m.TerceroNombre, m.Concepto, m.BaseValor, m.ImpuestoValor, m.RetencionValor, m.Naturaleza, m.Estado); err != nil {
			return err
		}
	}
	return nil
}

func normalizeDeclaracionTributaria(row EmpresaDeclaracionTributaria) EmpresaDeclaracionTributaria {
	row.TipoDeclaracion = normalizeDeclaracionTipo(row.TipoDeclaracion)
	row.Periodo = normalizeDeclaracionPeriodo(row.Periodo)
	row.Periodicidad = normalizeDeclaracionPeriodicidad(row.Periodicidad)
	row.Estado = normalizeDeclaracionEstado(row.Estado)
	if row.Anio <= 0 {
		row.Anio = yearFromPeriodo(row.Periodo)
	}
	if row.FechaDesde == "" || row.FechaHasta == "" {
		row.FechaDesde, row.FechaHasta = periodoRangeDeclaracion(row.Periodo)
	}
	if row.FechaVencimiento == "" {
		row.FechaVencimiento = defaultVencimientoDeclaracion(row.Periodo, row.TipoDeclaracion)
	}
	if row.Formulario == "" {
		row.Formulario = formularioDeclaracion(row.TipoDeclaracion)
	}
	return calcularSaldosDeclaracion(row)
}

func normalizeDeclaracionTipo(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "iva", "retencion_fuente", "reteiva", "reteica", "ica", "consumo", "regimen_simple", "renta", "todos":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "iva"
	}
}

func normalizeDeclaracionEstado(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "borrador", "revisada", "presentada", "pagada", "vencida", "anulada", "todos":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "borrador"
	}
}

func normalizeDeclaracionActivoEstado(v string) string {
	if strings.EqualFold(strings.TrimSpace(v), "inactivo") {
		return "inactivo"
	}
	return "activo"
}

func normalizeDeclaracionPeriodicidad(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "mensual", "bimestral", "cuatrimestral", "trimestral", "semestral", "anual":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "mensual"
	}
}

func normalizeDeclaracionPeriodo(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 7 {
		return v[:7]
	}
	return time.Now().Format("2006-01")
}

func periodoRangeDeclaracion(periodo string) (string, string) {
	periodo = normalizeDeclaracionPeriodo(periodo)
	t, err := time.Parse("2006-01-02", periodo+"-01")
	if err != nil {
		t = time.Now()
	}
	start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, -1)
	return start.Format("2006-01-02"), end.Format("2006-01-02")
}

func yearFromPeriodo(periodo string) int {
	t, err := time.Parse("2006-01", normalizeDeclaracionPeriodo(periodo))
	if err != nil {
		return time.Now().Year()
	}
	return t.Year()
}

func periodicidadDeclaracion(tipo string) string {
	switch normalizeDeclaracionTipo(tipo) {
	case "iva", "ica":
		return "bimestral"
	case "renta", "regimen_simple":
		return "anual"
	default:
		return "mensual"
	}
}

func formularioDeclaracion(tipo string) string {
	switch normalizeDeclaracionTipo(tipo) {
	case "iva":
		return "DIAN 300"
	case "retencion_fuente", "reteiva":
		return "DIAN 350"
	case "renta":
		return "DIAN 110/210"
	case "regimen_simple":
		return "DIAN 260"
	case "ica", "reteica":
		return "Municipal ICA"
	case "consumo":
		return "Impuesto al consumo"
	default:
		return "Tributario"
	}
}

func defaultVencimientoDeclaracion(periodo, tipo string) string {
	_, hasta := periodoRangeDeclaracion(periodo)
	t, err := time.Parse("2006-01-02", hasta)
	if err != nil {
		t = time.Now()
	}
	days := 12
	if normalizeDeclaracionTipo(tipo) == "renta" || normalizeDeclaracionTipo(tipo) == "regimen_simple" {
		days = 90
	}
	return t.AddDate(0, 1, days).Format("2006-01-02")
}

func roundDeclaracion(v float64) float64 {
	if v == 0 {
		return 0
	}
	return float64(int64(v*100+0.5)) / 100
}
