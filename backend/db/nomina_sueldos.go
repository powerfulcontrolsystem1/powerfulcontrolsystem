package db

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	nominaEstadoActivo   = "activo"
	nominaEstadoInactivo = "inactivo"
)

// EmpresaNominaConfiguracion representa la parametrizacion legal y operativa de nomina por empresa.
type EmpresaNominaConfiguracion struct {
	ID                                    int64   `json:"id"`
	EmpresaID                             int64   `json:"empresa_id"`
	PaisCodigo                            string  `json:"pais_codigo"`
	Moneda                                string  `json:"moneda"`
	HorasOrdinariasSemana                 float64 `json:"horas_ordinarias_semana"`
	HorasOrdinariasDia                    float64 `json:"horas_ordinarias_dia"`
	DiasNominaMes                         int     `json:"dias_nomina_mes"`
	DivisorHoraOrdinaria                  float64 `json:"divisor_hora_ordinaria"`
	HoraNocturnaDesde                     string  `json:"hora_nocturna_desde"`
	HoraNocturnaHasta                     string  `json:"hora_nocturna_hasta"`
	RecargoNocturnoPorcentaje             float64 `json:"recargo_nocturno_porcentaje"`
	HoraExtraDiurnaPorcentaje             float64 `json:"hora_extra_diurna_porcentaje"`
	HoraExtraNocturnaPorcentaje           float64 `json:"hora_extra_nocturna_porcentaje"`
	RecargoDominicalDiurnoPorcentaje      float64 `json:"recargo_dominical_diurno_porcentaje"`
	RecargoDominicalNocturnoPorcentaje    float64 `json:"recargo_dominical_nocturno_porcentaje"`
	HoraExtraDominicalDiurnaPorcentaje    float64 `json:"hora_extra_dominical_diurna_porcentaje"`
	HoraExtraDominicalNocturnaPorcentaje  float64 `json:"hora_extra_dominical_nocturna_porcentaje"`
	DeduccionSaludPorcentaje              float64 `json:"deduccion_salud_porcentaje"`
	DeduccionPensionPorcentaje            float64 `json:"deduccion_pension_porcentaje"`
	DeduccionFondoSolidaridadPorcentaje   float64 `json:"deduccion_fondo_solidaridad_porcentaje"`
	AporteSaludEmpleadorPorcentaje        float64 `json:"aporte_salud_empleador_porcentaje"`
	AportePensionEmpleadorPorcentaje      float64 `json:"aporte_pension_empleador_porcentaje"`
	AporteARLPorcentaje                   float64 `json:"aporte_arl_porcentaje"`
	AporteCajaCompensacionPorcentaje      float64 `json:"aporte_caja_compensacion_porcentaje"`
	AporteICBFPorcentaje                  float64 `json:"aporte_icbf_porcentaje"`
	AporteSENAPorcentaje                  float64 `json:"aporte_sena_porcentaje"`
	ProvisionCesantiasPorcentaje          float64 `json:"provision_cesantias_porcentaje"`
	ProvisionInteresesCesantiasPorcentaje float64 `json:"provision_intereses_cesantias_porcentaje"`
	ProvisionPrimaPorcentaje              float64 `json:"provision_prima_porcentaje"`
	ProvisionVacacionesPorcentaje         float64 `json:"provision_vacaciones_porcentaje"`
	FechaCreacion                         string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion                    string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador                        string  `json:"usuario_creador,omitempty"`
	Estado                                string  `json:"estado,omitempty"`
	Observaciones                         string  `json:"observaciones,omitempty"`
}

// EmpresaNominaEmpleado representa la ficha salarial de un empleado para nomina.
type EmpresaNominaEmpleado struct {
	ID                       int64   `json:"id"`
	EmpresaID                int64   `json:"empresa_id"`
	EmpleadoID               int64   `json:"empleado_id"`
	EmpleadoCodigo           string  `json:"empleado_codigo,omitempty"`
	EmpleadoNombre           string  `json:"empleado_nombre"`
	EmpleadoDocumento        string  `json:"empleado_documento,omitempty"`
	Cargo                    string  `json:"cargo,omitempty"`
	SedeCodigo               string  `json:"sede_codigo,omitempty"`
	SedeNombre               string  `json:"sede_nombre,omitempty"`
	CentroCosto              string  `json:"centro_costo,omitempty"`
	TipoContrato             string  `json:"tipo_contrato,omitempty"`
	FechaIngreso             string  `json:"fecha_ingreso,omitempty"`
	SalarioBasicoMensual     float64 `json:"salario_basico_mensual"`
	AuxilioTransporteMensual float64 `json:"auxilio_transporte_mensual"`
	BonificacionFijaMensual  float64 `json:"bonificacion_fija_mensual"`
	DeduccionFijaMensual     float64 `json:"deduccion_fija_mensual"`
	JornadaHorasDia          float64 `json:"jornada_horas_dia"`
	IncluirAuxilioTransporte bool    `json:"incluir_auxilio_transporte"`
	FechaCreacion            string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion       string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador           string  `json:"usuario_creador,omitempty"`
	Estado                   string  `json:"estado,omitempty"`
	Observaciones            string  `json:"observaciones,omitempty"`
}

// EmpresaNominaFestivo representa un dia festivo configurado para calculo de recargos.
type EmpresaNominaFestivo struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	FechaFestivo       string `json:"fecha_festivo"`
	Descripcion        string `json:"descripcion,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// EmpresaNominaLiquidacion representa una liquidacion salarial por empleado y periodo.
type EmpresaNominaLiquidacion struct {
	ID                             int64   `json:"id"`
	EmpresaID                      int64   `json:"empresa_id"`
	EmpleadoNominaID               int64   `json:"empleado_nomina_id"`
	EmpleadoID                     int64   `json:"empleado_id"`
	EmpleadoCodigo                 string  `json:"empleado_codigo,omitempty"`
	EmpleadoNombre                 string  `json:"empleado_nombre"`
	EmpleadoDocumento              string  `json:"empleado_documento,omitempty"`
	Cargo                          string  `json:"cargo,omitempty"`
	SedeCodigo                     string  `json:"sede_codigo,omitempty"`
	SedeNombre                     string  `json:"sede_nombre,omitempty"`
	CentroCosto                    string  `json:"centro_costo,omitempty"`
	PeriodoDesde                   string  `json:"periodo_desde"`
	PeriodoHasta                   string  `json:"periodo_hasta"`
	DiasLiquidados                 float64 `json:"dias_liquidados"`
	HorasAsistenciaTotal           float64 `json:"horas_asistencia_total"`
	RegistrosAsistencia            int64   `json:"registros_asistencia"`
	HorasOrdinarias                float64 `json:"horas_ordinarias"`
	HorasRecargoNocturno           float64 `json:"horas_recargo_nocturno"`
	HorasExtraDiurnas              float64 `json:"horas_extra_diurnas"`
	HorasExtraNocturnas            float64 `json:"horas_extra_nocturnas"`
	HorasDominicalesDiurnas        float64 `json:"horas_dominicales_diurnas"`
	HorasDominicalesNocturnas      float64 `json:"horas_dominicales_nocturnas"`
	HorasExtraDominicalesDiurnas   float64 `json:"horas_extra_dominicales_diurnas"`
	HorasExtraDominicalesNocturnas float64 `json:"horas_extra_dominicales_nocturnas"`
	ValorHoraOrdinaria             float64 `json:"valor_hora_ordinaria"`
	BaseSalarioProporcional        float64 `json:"base_salario_proporcional"`
	ValorRecargoNocturno           float64 `json:"valor_recargo_nocturno"`
	ValorDominicalDiurno           float64 `json:"valor_dominical_diurno"`
	ValorDominicalNocturno         float64 `json:"valor_dominical_nocturno"`
	ValorExtraDiurna               float64 `json:"valor_extra_diurna"`
	ValorExtraNocturna             float64 `json:"valor_extra_nocturna"`
	ValorExtraDominicalDiurna      float64 `json:"valor_extra_dominical_diurna"`
	ValorExtraDominicalNocturna    float64 `json:"valor_extra_dominical_nocturna"`
	TotalRecargosHorasExtras       float64 `json:"total_recargos_horas_extras"`
	AuxilioTransporte              float64 `json:"auxilio_transporte"`
	Bonificacion                   float64 `json:"bonificacion"`
	ComisionesServicioTotal        float64 `json:"comisiones_servicio_total"`
	ComisionesServicioMovimientos  int64   `json:"comisiones_servicio_movimientos"`
	ComisionesServicioAjustes      float64 `json:"comisiones_servicio_ajustes"`
	DevengadoTotal                 float64 `json:"devengado_total"`
	IngresoBaseCotizacion          float64 `json:"ingreso_base_cotizacion"`
	DeduccionSalud                 float64 `json:"deduccion_salud"`
	DeduccionPension               float64 `json:"deduccion_pension"`
	DeduccionFondoSolidaridad      float64 `json:"deduccion_fondo_solidaridad"`
	DeduccionFija                  float64 `json:"deduccion_fija"`
	OtrasDeducciones               float64 `json:"otras_deducciones"`
	DeduccionTotal                 float64 `json:"deduccion_total"`
	NetoPagar                      float64 `json:"neto_pagar"`
	OrigenCalculo                  string  `json:"origen_calculo,omitempty"`
	ResumenJSON                    string  `json:"resumen_json,omitempty"`
	FechaGeneracion                string  `json:"fecha_generacion,omitempty"`
	FechaCreacion                  string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion             string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador                 string  `json:"usuario_creador,omitempty"`
	Estado                         string  `json:"estado,omitempty"`
	Observaciones                  string  `json:"observaciones,omitempty"`
}

// EmpresaNominaLiquidacionFilter aplica filtros de consulta para liquidaciones de nomina.
type EmpresaNominaLiquidacionFilter struct {
	PeriodoDesde     string
	PeriodoHasta     string
	EmpleadoNominaID int64
	IncludeInactive  bool
	Limit            int
}

type EmpresaNominaPago struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	LiquidacionID      int64   `json:"liquidacion_id"`
	EmpleadoNominaID   int64   `json:"empleado_nomina_id"`
	EmpleadoNombre     string  `json:"empleado_nombre"`
	EmpleadoDocumento  string  `json:"empleado_documento,omitempty"`
	PeriodoDesde       string  `json:"periodo_desde"`
	PeriodoHasta       string  `json:"periodo_hasta"`
	FechaPago          string  `json:"fecha_pago"`
	MetodoPago         string  `json:"metodo_pago"`
	CuentaBancaria     string  `json:"cuenta_bancaria,omitempty"`
	ReferenciaPago     string  `json:"referencia_pago,omitempty"`
	DevengadoTotal     float64 `json:"devengado_total"`
	DeduccionTotal     float64 `json:"deduccion_total"`
	NetoPagado         float64 `json:"neto_pagado"`
	EstadoPago         string  `json:"estado_pago"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

type EmpresaNominaPagoFilter struct {
	PeriodoDesde     string
	PeriodoHasta     string
	EmpleadoNominaID int64
	IncludeInactive  bool
	Limit            int
}

type EmpresaNominaPagosResult struct {
	EmpresaID       int64               `json:"empresa_id"`
	PeriodoDesde    string              `json:"periodo_desde"`
	PeriodoHasta    string              `json:"periodo_hasta"`
	Generados       int                 `json:"generados"`
	Omitidos        int                 `json:"omitidos"`
	TotalNetoPagado float64             `json:"total_neto_pagado"`
	Pagos           []EmpresaNominaPago `json:"pagos"`
}

type EmpresaNominaProvisionesResumen struct {
	EmpresaID                   int64   `json:"empresa_id"`
	PeriodoDesde                string  `json:"periodo_desde"`
	PeriodoHasta                string  `json:"periodo_hasta"`
	Liquidaciones               int     `json:"liquidaciones"`
	TotalDevengado              float64 `json:"total_devengado"`
	TotalIBC                    float64 `json:"total_ibc"`
	TotalNetoPagar              float64 `json:"total_neto_pagar"`
	AporteSaludEmpleador        float64 `json:"aporte_salud_empleador"`
	AportePensionEmpleador      float64 `json:"aporte_pension_empleador"`
	AporteARL                   float64 `json:"aporte_arl"`
	AporteCajaCompensacion      float64 `json:"aporte_caja_compensacion"`
	AporteICBF                  float64 `json:"aporte_icbf"`
	AporteSENA                  float64 `json:"aporte_sena"`
	ProvisionCesantias          float64 `json:"provision_cesantias"`
	ProvisionInteresesCesantias float64 `json:"provision_intereses_cesantias"`
	ProvisionPrima              float64 `json:"provision_prima"`
	ProvisionVacaciones         float64 `json:"provision_vacaciones"`
	CostoEmpresaEstimado        float64 `json:"costo_empresa_estimado"`
}

// EmpresaNominaCalculoRequest representa una solicitud de calculo de nomina por periodo.
type EmpresaNominaCalculoRequest struct {
	EmpresaID        int64   `json:"empresa_id"`
	PeriodoDesde     string  `json:"periodo_desde"`
	PeriodoHasta     string  `json:"periodo_hasta"`
	EmpleadoNominaID int64   `json:"empleado_nomina_id,omitempty"`
	Overwrite        bool    `json:"overwrite"`
	OtrasDeducciones float64 `json:"otras_deducciones,omitempty"`
	UsuarioCreador   string  `json:"usuario_creador,omitempty"`
	Observaciones    string  `json:"observaciones,omitempty"`
}

// EmpresaNominaCalculoResult resume el resultado de un proceso de calculo de nomina.
type EmpresaNominaCalculoResult struct {
	EmpresaID      int64                      `json:"empresa_id"`
	PeriodoDesde   string                     `json:"periodo_desde"`
	PeriodoHasta   string                     `json:"periodo_hasta"`
	Calculados     int                        `json:"calculados"`
	Liquidaciones  []EmpresaNominaLiquidacion `json:"liquidaciones"`
	TotalDevengado float64                    `json:"total_devengado"`
	TotalDeduccion float64                    `json:"total_deduccion"`
	TotalNeto      float64                    `json:"total_neto"`
	Mensajes       []string                   `json:"mensajes,omitempty"`
}

// EmpresaNominaDesprendible representa un desprendible estandar por empleado y periodo.
type EmpresaNominaDesprendible struct {
	EmpresaID                      int64   `json:"empresa_id"`
	PaisCodigo                     string  `json:"pais_codigo"`
	Moneda                         string  `json:"moneda"`
	PeriodoDesde                   string  `json:"periodo_desde"`
	PeriodoHasta                   string  `json:"periodo_hasta"`
	FechaGeneracion                string  `json:"fecha_generacion,omitempty"`
	EmpleadoNominaID               int64   `json:"empleado_nomina_id"`
	EmpleadoID                     int64   `json:"empleado_id"`
	EmpleadoCodigo                 string  `json:"empleado_codigo,omitempty"`
	EmpleadoNombre                 string  `json:"empleado_nombre"`
	EmpleadoDocumento              string  `json:"empleado_documento,omitempty"`
	Cargo                          string  `json:"cargo,omitempty"`
	SedeCodigo                     string  `json:"sede_codigo,omitempty"`
	SedeNombre                     string  `json:"sede_nombre,omitempty"`
	CentroCosto                    string  `json:"centro_costo,omitempty"`
	TipoContrato                   string  `json:"tipo_contrato,omitempty"`
	FechaIngreso                   string  `json:"fecha_ingreso,omitempty"`
	DiasLiquidados                 float64 `json:"dias_liquidados"`
	HorasAsistencia                float64 `json:"horas_asistencia_total"`
	RegistrosAsistencia            int64   `json:"registros_asistencia"`
	HorasOrdinarias                float64 `json:"horas_ordinarias"`
	HorasRecargoNocturno           float64 `json:"horas_recargo_nocturno"`
	HorasExtraDiurnas              float64 `json:"horas_extra_diurnas"`
	HorasExtraNocturnas            float64 `json:"horas_extra_nocturnas"`
	HorasDominicalesDiurnas        float64 `json:"horas_dominicales_diurnas"`
	HorasDominicalesNocturnas      float64 `json:"horas_dominicales_nocturnas"`
	HorasExtraDominicalesDiurnas   float64 `json:"horas_extra_dominicales_diurnas"`
	HorasExtraDominicalesNocturnas float64 `json:"horas_extra_dominicales_nocturnas"`
	ValorHoraOrdinaria             float64 `json:"valor_hora_ordinaria"`
	BaseSalarioProporcional        float64 `json:"base_salario_proporcional"`
	ValorRecargoNocturno           float64 `json:"valor_recargo_nocturno"`
	ValorDominicalDiurno           float64 `json:"valor_dominical_diurno"`
	ValorDominicalNocturno         float64 `json:"valor_dominical_nocturno"`
	ValorExtraDiurna               float64 `json:"valor_extra_diurna"`
	ValorExtraNocturna             float64 `json:"valor_extra_nocturna"`
	ValorExtraDominicalDiurna      float64 `json:"valor_extra_dominical_diurna"`
	ValorExtraDominicalNocturna    float64 `json:"valor_extra_dominical_nocturna"`
	TotalRecargosHorasExtras       float64 `json:"total_recargos_horas_extras"`
	AuxilioTransporte              float64 `json:"auxilio_transporte"`
	Bonificacion                   float64 `json:"bonificacion"`
	ComisionesServicioTotal        float64 `json:"comisiones_servicio_total"`
	ComisionesServicioMovimientos  int64   `json:"comisiones_servicio_movimientos"`
	ComisionesServicioAjustes      float64 `json:"comisiones_servicio_ajustes"`
	DevengadoTotal                 float64 `json:"devengado_total"`
	IngresoBaseCotizacion          float64 `json:"ingreso_base_cotizacion"`
	DeduccionSalud                 float64 `json:"deduccion_salud"`
	DeduccionPension               float64 `json:"deduccion_pension"`
	DeduccionFondoSolidaridad      float64 `json:"deduccion_fondo_solidaridad"`
	DeduccionFija                  float64 `json:"deduccion_fija"`
	OtrasDeducciones               float64 `json:"otras_deducciones"`
	DeduccionTotal                 float64 `json:"deduccion_total"`
	NetoPagar                      float64 `json:"neto_pagar"`
	ResumenJSON                    string  `json:"resumen_json,omitempty"`
}

// EmpresaNominaConciliacionRequest representa una solicitud de conciliacion entre asistencia y nomina.
type EmpresaNominaConciliacionRequest struct {
	EmpresaID        int64  `json:"empresa_id"`
	PeriodoDesde     string `json:"periodo_desde"`
	PeriodoHasta     string `json:"periodo_hasta"`
	EmpleadoNominaID int64  `json:"empleado_nomina_id,omitempty"`
	AutoRecalcular   bool   `json:"auto_recalcular"`
	UsuarioCreador   string `json:"usuario_creador,omitempty"`
	Observaciones    string `json:"observaciones,omitempty"`
}

// EmpresaNominaConciliacionItem representa el resultado de conciliacion por empleado.
type EmpresaNominaConciliacionItem struct {
	EmpleadoNominaID               int64   `json:"empleado_nomina_id"`
	EmpleadoNombre                 string  `json:"empleado_nombre"`
	Estado                         string  `json:"estado"`
	RegistrosAsistenciaLiquidacion int64   `json:"registros_asistencia_liquidacion"`
	RegistrosAsistenciaCalculado   int64   `json:"registros_asistencia_calculado"`
	HorasAsistenciaLiquidacion     float64 `json:"horas_asistencia_liquidacion"`
	HorasAsistenciaCalculado       float64 `json:"horas_asistencia_calculado"`
	Mensaje                        string  `json:"mensaje,omitempty"`
}

// EmpresaNominaConciliacionResult resume la conciliacion de nomina vs asistencia por periodo.
type EmpresaNominaConciliacionResult struct {
	EmpresaID            int64                           `json:"empresa_id"`
	PeriodoDesde         string                          `json:"periodo_desde"`
	PeriodoHasta         string                          `json:"periodo_hasta"`
	AutoRecalcular       bool                            `json:"auto_recalcular"`
	TotalEvaluados       int                             `json:"total_evaluados"`
	TotalConciliados     int                             `json:"total_conciliados"`
	TotalInconsistencias int                             `json:"total_inconsistencias"`
	TotalRecalculados    int                             `json:"total_recalculados"`
	Items                []EmpresaNominaConciliacionItem `json:"items"`
	Mensajes             []string                        `json:"mensajes,omitempty"`
}

type nominaAsistenciaRow struct {
	FechaAsistencia  string
	HoraEntrada      string
	HoraSalida       string
	HorasTrabajadas  float64
	EstadoAsistencia string
}

type nominaHorasDetalle struct {
	DiasLiquidados                 float64
	HorasAsistenciaTotal           float64
	RegistrosAsistencia            int64
	HorasOrdinarias                float64
	HorasRecargoNocturno           float64
	HorasExtraDiurnas              float64
	HorasExtraNocturnas            float64
	HorasDominicalesDiurnas        float64
	HorasDominicalesNocturnas      float64
	HorasExtraDominicalesDiurnas   float64
	HorasExtraDominicalesNocturnas float64
}

type nominaHorasMinutos struct {
	registros                 int64
	totalMinutos              int
	ordinariasMin             int
	recargoNocturnoMin        int
	extraDiurnaMin            int
	extraNocturnaMin          int
	dominicalDiurnaMin        int
	dominicalNocturnaMin      int
	extraDominicalDiurnaMin   int
	extraDominicalNocturnaMin int
	dias                      map[string]struct{}
}

// EnsureEmpresaNominaSchema crea y migra las tablas del modulo de nomina por empresa.
func EnsureEmpresaNominaSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_nomina_configuracion (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL UNIQUE,
			pais_codigo TEXT DEFAULT 'CO',
			moneda TEXT DEFAULT 'COP',
			horas_ordinarias_semana REAL DEFAULT 44,
			horas_ordinarias_dia REAL DEFAULT 8,
			dias_nomina_mes INTEGER DEFAULT 30,
			divisor_hora_ordinaria REAL DEFAULT 220,
			hora_nocturna_desde TEXT DEFAULT '21:00:00',
			hora_nocturna_hasta TEXT DEFAULT '06:00:00',
			recargo_nocturno_porcentaje REAL DEFAULT 35,
			hora_extra_diurna_porcentaje REAL DEFAULT 25,
			hora_extra_nocturna_porcentaje REAL DEFAULT 75,
			recargo_dominical_diurno_porcentaje REAL DEFAULT 75,
			recargo_dominical_nocturno_porcentaje REAL DEFAULT 110,
			hora_extra_dominical_diurna_porcentaje REAL DEFAULT 100,
			hora_extra_dominical_nocturna_porcentaje REAL DEFAULT 150,
			deduccion_salud_porcentaje REAL DEFAULT 4,
			deduccion_pension_porcentaje REAL DEFAULT 4,
			deduccion_fondo_solidaridad_porcentaje REAL DEFAULT 0,
			aporte_salud_empleador_porcentaje REAL DEFAULT 8.5,
			aporte_pension_empleador_porcentaje REAL DEFAULT 12,
			aporte_arl_porcentaje REAL DEFAULT 0.522,
			aporte_caja_compensacion_porcentaje REAL DEFAULT 4,
			aporte_icbf_porcentaje REAL DEFAULT 3,
			aporte_sena_porcentaje REAL DEFAULT 2,
			provision_cesantias_porcentaje REAL DEFAULT 8.33,
			provision_intereses_cesantias_porcentaje REAL DEFAULT 1,
			provision_prima_porcentaje REAL DEFAULT 8.33,
			provision_vacaciones_porcentaje REAL DEFAULT 4.17,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_nomina_configuracion_empresa ON empresa_nomina_configuracion(empresa_id);`,
		`CREATE TABLE IF NOT EXISTS empresa_nomina_empleados (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			empleado_id INTEGER DEFAULT 0,
			empleado_codigo TEXT,
			empleado_nombre TEXT NOT NULL,
			empleado_documento TEXT,
			cargo TEXT,
			sede_codigo TEXT,
			sede_nombre TEXT,
			centro_costo TEXT,
			tipo_contrato TEXT DEFAULT 'indefinido',
			fecha_ingreso TEXT,
			salario_basico_mensual REAL DEFAULT 0,
			auxilio_transporte_mensual REAL DEFAULT 0,
			bonificacion_fija_mensual REAL DEFAULT 0,
			deduccion_fija_mensual REAL DEFAULT 0,
			jornada_horas_dia REAL DEFAULT 8,
			incluir_auxilio_transporte INTEGER DEFAULT 1,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_nomina_empleados_empresa_estado ON empresa_nomina_empleados(empresa_id, estado, empleado_nombre);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_nomina_empleados_empresa_documento ON empresa_nomina_empleados(empresa_id, empleado_documento);`,
		`CREATE TABLE IF NOT EXISTS empresa_nomina_festivos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			fecha_festivo TEXT NOT NULL,
			descripcion TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, fecha_festivo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_nomina_festivos_empresa_fecha ON empresa_nomina_festivos(empresa_id, fecha_festivo DESC);`,
		`CREATE TABLE IF NOT EXISTS empresa_nomina_liquidaciones (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			empleado_nomina_id INTEGER NOT NULL,
			empleado_id INTEGER DEFAULT 0,
			empleado_codigo TEXT,
			empleado_nombre TEXT NOT NULL,
			empleado_documento TEXT,
			cargo TEXT,
			sede_codigo TEXT,
			sede_nombre TEXT,
			centro_costo TEXT,
			periodo_desde TEXT NOT NULL,
			periodo_hasta TEXT NOT NULL,
			dias_liquidados REAL DEFAULT 0,
			horas_asistencia_total REAL DEFAULT 0,
			registros_asistencia INTEGER DEFAULT 0,
			horas_ordinarias REAL DEFAULT 0,
			horas_recargo_nocturno REAL DEFAULT 0,
			horas_extra_diurnas REAL DEFAULT 0,
			horas_extra_nocturnas REAL DEFAULT 0,
			horas_dominicales_diurnas REAL DEFAULT 0,
			horas_dominicales_nocturnas REAL DEFAULT 0,
			horas_extra_dominicales_diurnas REAL DEFAULT 0,
			horas_extra_dominicales_nocturnas REAL DEFAULT 0,
			valor_hora_ordinaria REAL DEFAULT 0,
			base_salario_proporcional REAL DEFAULT 0,
			valor_recargo_nocturno REAL DEFAULT 0,
			valor_dominical_diurno REAL DEFAULT 0,
			valor_dominical_nocturno REAL DEFAULT 0,
			valor_extra_diurna REAL DEFAULT 0,
			valor_extra_nocturna REAL DEFAULT 0,
			valor_extra_dominical_diurna REAL DEFAULT 0,
			valor_extra_dominical_nocturna REAL DEFAULT 0,
			total_recargos_horas_extras REAL DEFAULT 0,
			auxilio_transporte REAL DEFAULT 0,
			bonificacion REAL DEFAULT 0,
			comisiones_servicio_total REAL DEFAULT 0,
			comisiones_servicio_movimientos INTEGER DEFAULT 0,
			comisiones_servicio_ajustes REAL DEFAULT 0,
			devengado_total REAL DEFAULT 0,
			ingreso_base_cotizacion REAL DEFAULT 0,
			deduccion_salud REAL DEFAULT 0,
			deduccion_pension REAL DEFAULT 0,
			deduccion_fondo_solidaridad REAL DEFAULT 0,
			deduccion_fija REAL DEFAULT 0,
			otras_deducciones REAL DEFAULT 0,
			deduccion_total REAL DEFAULT 0,
			neto_pagar REAL DEFAULT 0,
			origen_calculo TEXT DEFAULT 'asistencia',
			resumen_json TEXT,
			fecha_generacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, empleado_nomina_id, periodo_desde, periodo_hasta)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_nomina_liquidaciones_empresa_periodo ON empresa_nomina_liquidaciones(empresa_id, periodo_desde DESC, periodo_hasta DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_nomina_liquidaciones_empresa_empleado ON empresa_nomina_liquidaciones(empresa_id, empleado_nomina_id, empleado_documento);`,
		`CREATE TABLE IF NOT EXISTS empresa_nomina_pagos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			liquidacion_id INTEGER NOT NULL,
			empleado_nomina_id INTEGER NOT NULL,
			empleado_nombre TEXT NOT NULL,
			empleado_documento TEXT,
			periodo_desde TEXT NOT NULL,
			periodo_hasta TEXT NOT NULL,
			fecha_pago TEXT DEFAULT (datetime('now','localtime')),
			metodo_pago TEXT DEFAULT 'transferencia_bancaria',
			cuenta_bancaria TEXT,
			referencia_pago TEXT,
			devengado_total REAL DEFAULT 0,
			deduccion_total REAL DEFAULT 0,
			neto_pagado REAL DEFAULT 0,
			estado_pago TEXT DEFAULT 'pagado',
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, liquidacion_id)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_nomina_pagos_empresa_periodo ON empresa_nomina_pagos(empresa_id, periodo_desde DESC, periodo_hasta DESC, id DESC);`,
	}

	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_nomina_configuracion", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_nomina_configuracion", "aporte_salud_empleador_porcentaje", "REAL DEFAULT 8.5"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_nomina_configuracion", "aporte_pension_empleador_porcentaje", "REAL DEFAULT 12"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_nomina_configuracion", "aporte_arl_porcentaje", "REAL DEFAULT 0.522"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_nomina_configuracion", "aporte_caja_compensacion_porcentaje", "REAL DEFAULT 4"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_nomina_configuracion", "aporte_icbf_porcentaje", "REAL DEFAULT 3"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_nomina_configuracion", "aporte_sena_porcentaje", "REAL DEFAULT 2"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_nomina_configuracion", "provision_cesantias_porcentaje", "REAL DEFAULT 8.33"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_nomina_configuracion", "provision_intereses_cesantias_porcentaje", "REAL DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_nomina_configuracion", "provision_prima_porcentaje", "REAL DEFAULT 8.33"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_nomina_configuracion", "provision_vacaciones_porcentaje", "REAL DEFAULT 4.17"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_nomina_empleados", "incluir_auxilio_transporte", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	for _, col := range []string{"sede_codigo", "sede_nombre", "centro_costo"} {
		if err := ensureColumnIfMissing(dbConn, "empresa_nomina_empleados", col, "TEXT"); err != nil {
			return err
		}
		if err := ensureColumnIfMissing(dbConn, "empresa_nomina_liquidaciones", col, "TEXT"); err != nil {
			return err
		}
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_nomina_liquidaciones", "valor_extra_dominical_nocturna", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_nomina_liquidaciones", "comisiones_servicio_total", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_nomina_liquidaciones", "comisiones_servicio_movimientos", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_nomina_liquidaciones", "comisiones_servicio_ajustes", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_nomina_liquidaciones", "otras_deducciones", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := EnsureEmpresaNominaColombiaAvanzadaSchema(dbConn); err != nil {
		return err
	}
	return nil
}

func defaultEmpresaNominaConfiguracion(empresaID int64) EmpresaNominaConfiguracion {
	cfg := EmpresaNominaConfiguracion{
		EmpresaID:                             empresaID,
		PaisCodigo:                            "CO",
		Moneda:                                "COP",
		HorasOrdinariasSemana:                 44,
		HorasOrdinariasDia:                    8,
		DiasNominaMes:                         30,
		DivisorHoraOrdinaria:                  220,
		HoraNocturnaDesde:                     "21:00:00",
		HoraNocturnaHasta:                     "06:00:00",
		RecargoNocturnoPorcentaje:             35,
		HoraExtraDiurnaPorcentaje:             25,
		HoraExtraNocturnaPorcentaje:           75,
		RecargoDominicalDiurnoPorcentaje:      75,
		RecargoDominicalNocturnoPorcentaje:    110,
		HoraExtraDominicalDiurnaPorcentaje:    100,
		HoraExtraDominicalNocturnaPorcentaje:  150,
		DeduccionSaludPorcentaje:              4,
		DeduccionPensionPorcentaje:            4,
		DeduccionFondoSolidaridadPorcentaje:   0,
		AporteSaludEmpleadorPorcentaje:        8.5,
		AportePensionEmpleadorPorcentaje:      12,
		AporteARLPorcentaje:                   0.522,
		AporteCajaCompensacionPorcentaje:      4,
		AporteICBFPorcentaje:                  3,
		AporteSENAPorcentaje:                  2,
		ProvisionCesantiasPorcentaje:          8.33,
		ProvisionInteresesCesantiasPorcentaje: 1,
		ProvisionPrimaPorcentaje:              8.33,
		ProvisionVacacionesPorcentaje:         4.17,
		Estado:                                nominaEstadoActivo,
	}
	cfg.DivisorHoraOrdinaria = recommendedNominaHourDivisor(cfg.HorasOrdinariasSemana)
	return cfg
}

func normalizeNominaEstado(v string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	if s == nominaEstadoInactivo {
		return nominaEstadoInactivo
	}
	return nominaEstadoActivo
}

func normalizeNominaMoneda(v string) string {
	s := strings.ToUpper(strings.TrimSpace(v))
	if s == "" {
		return "COP"
	}
	return s
}

func normalizeNominaPais(v string) string {
	s := strings.ToUpper(strings.TrimSpace(v))
	if s == "" {
		return "CO"
	}
	return s
}

func normalizeNominaPorcentaje(v float64) float64 {
	if v < 0 {
		v = 0
	}
	if v > 1000 {
		v = 1000
	}
	return round2(v)
}

func normalizeNominaHoras(v, fallback float64) float64 {
	if v <= 0 {
		v = fallback
	}
	if v < 0 {
		v = 0
	}
	if v > 24 {
		v = 24
	}
	return round2(v)
}

func recommendedNominaHourDivisor(horasSemana float64) float64 {
	h := horasSemana
	if h <= 0 {
		h = 44
	}
	div := h * 30.0 / 6.0
	if div <= 0 {
		div = 220
	}
	return round2(div)
}

func normalizeNominaTimeWindow(raw, fallback string) string {
	v := strings.TrimSpace(raw)
	if v == "" {
		v = fallback
	}
	parsed, err := normalizeAsistenciaTime(v)
	if err != nil {
		parsed, _ = normalizeAsistenciaTime(fallback)
	}
	if parsed == "" {
		parsed = fallback
	}
	return parsed
}

func normalizeNominaDate(raw string) (string, error) {
	return normalizeAsistenciaDate(raw)
}

func normalizeNominaTipoContrato(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "indefinido", "fijo", "obra_labor", "servicios", "aprendizaje", "temporal":
		return v
	default:
		if v == "" {
			return "indefinido"
		}
		return v
	}
}

// GetEmpresaNominaConfiguracion obtiene configuracion legal de nomina por empresa.
func GetEmpresaNominaConfiguracion(dbConn *sql.DB, empresaID int64) (*EmpresaNominaConfiguracion, error) {
	cfg := defaultEmpresaNominaConfiguracion(empresaID)
	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(pais_codigo, 'CO'),
		COALESCE(moneda, 'COP'),
		COALESCE(horas_ordinarias_semana, 44),
		COALESCE(horas_ordinarias_dia, 8),
		COALESCE(dias_nomina_mes, 30),
		COALESCE(divisor_hora_ordinaria, 220),
		COALESCE(hora_nocturna_desde, '21:00:00'),
		COALESCE(hora_nocturna_hasta, '06:00:00'),
		COALESCE(recargo_nocturno_porcentaje, 35),
		COALESCE(hora_extra_diurna_porcentaje, 25),
		COALESCE(hora_extra_nocturna_porcentaje, 75),
		COALESCE(recargo_dominical_diurno_porcentaje, 75),
		COALESCE(recargo_dominical_nocturno_porcentaje, 110),
		COALESCE(hora_extra_dominical_diurna_porcentaje, 100),
		COALESCE(hora_extra_dominical_nocturna_porcentaje, 150),
		COALESCE(deduccion_salud_porcentaje, 4),
		COALESCE(deduccion_pension_porcentaje, 4),
		COALESCE(deduccion_fondo_solidaridad_porcentaje, 0),
		COALESCE(aporte_salud_empleador_porcentaje, 8.5),
		COALESCE(aporte_pension_empleador_porcentaje, 12),
		COALESCE(aporte_arl_porcentaje, 0.522),
		COALESCE(aporte_caja_compensacion_porcentaje, 4),
		COALESCE(aporte_icbf_porcentaje, 3),
		COALESCE(aporte_sena_porcentaje, 2),
		COALESCE(provision_cesantias_porcentaje, 8.33),
		COALESCE(provision_intereses_cesantias_porcentaje, 1),
		COALESCE(provision_prima_porcentaje, 8.33),
		COALESCE(provision_vacaciones_porcentaje, 4.17),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_nomina_configuracion
	WHERE empresa_id = ?
	LIMIT 1`, empresaID)

	err := row.Scan(
		&cfg.ID,
		&cfg.EmpresaID,
		&cfg.PaisCodigo,
		&cfg.Moneda,
		&cfg.HorasOrdinariasSemana,
		&cfg.HorasOrdinariasDia,
		&cfg.DiasNominaMes,
		&cfg.DivisorHoraOrdinaria,
		&cfg.HoraNocturnaDesde,
		&cfg.HoraNocturnaHasta,
		&cfg.RecargoNocturnoPorcentaje,
		&cfg.HoraExtraDiurnaPorcentaje,
		&cfg.HoraExtraNocturnaPorcentaje,
		&cfg.RecargoDominicalDiurnoPorcentaje,
		&cfg.RecargoDominicalNocturnoPorcentaje,
		&cfg.HoraExtraDominicalDiurnaPorcentaje,
		&cfg.HoraExtraDominicalNocturnaPorcentaje,
		&cfg.DeduccionSaludPorcentaje,
		&cfg.DeduccionPensionPorcentaje,
		&cfg.DeduccionFondoSolidaridadPorcentaje,
		&cfg.AporteSaludEmpleadorPorcentaje,
		&cfg.AportePensionEmpleadorPorcentaje,
		&cfg.AporteARLPorcentaje,
		&cfg.AporteCajaCompensacionPorcentaje,
		&cfg.AporteICBFPorcentaje,
		&cfg.AporteSENAPorcentaje,
		&cfg.ProvisionCesantiasPorcentaje,
		&cfg.ProvisionInteresesCesantiasPorcentaje,
		&cfg.ProvisionPrimaPorcentaje,
		&cfg.ProvisionVacacionesPorcentaje,
		&cfg.FechaCreacion,
		&cfg.FechaActualizacion,
		&cfg.UsuarioCreador,
		&cfg.Estado,
		&cfg.Observaciones,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return &cfg, nil
		}
		return nil, err
	}
	return normalizeEmpresaNominaConfiguracion(cfg), nil
}

func normalizeEmpresaNominaConfiguracion(cfg EmpresaNominaConfiguracion) *EmpresaNominaConfiguracion {
	cfg.PaisCodigo = normalizeNominaPais(cfg.PaisCodigo)
	cfg.Moneda = normalizeNominaMoneda(cfg.Moneda)
	cfg.HorasOrdinariasSemana = normalizeNominaHoras(cfg.HorasOrdinariasSemana, 44)
	cfg.HorasOrdinariasDia = normalizeNominaHoras(cfg.HorasOrdinariasDia, 8)
	if cfg.DiasNominaMes <= 0 {
		cfg.DiasNominaMes = 30
	}
	if cfg.DiasNominaMes > 31 {
		cfg.DiasNominaMes = 31
	}
	cfg.DivisorHoraOrdinaria = round2(cfg.DivisorHoraOrdinaria)
	if cfg.DivisorHoraOrdinaria <= 0 {
		cfg.DivisorHoraOrdinaria = recommendedNominaHourDivisor(cfg.HorasOrdinariasSemana)
	}
	cfg.HoraNocturnaDesde = normalizeNominaTimeWindow(cfg.HoraNocturnaDesde, "21:00:00")
	cfg.HoraNocturnaHasta = normalizeNominaTimeWindow(cfg.HoraNocturnaHasta, "06:00:00")
	cfg.RecargoNocturnoPorcentaje = normalizeNominaPorcentaje(cfg.RecargoNocturnoPorcentaje)
	cfg.HoraExtraDiurnaPorcentaje = normalizeNominaPorcentaje(cfg.HoraExtraDiurnaPorcentaje)
	cfg.HoraExtraNocturnaPorcentaje = normalizeNominaPorcentaje(cfg.HoraExtraNocturnaPorcentaje)
	cfg.RecargoDominicalDiurnoPorcentaje = normalizeNominaPorcentaje(cfg.RecargoDominicalDiurnoPorcentaje)
	cfg.RecargoDominicalNocturnoPorcentaje = normalizeNominaPorcentaje(cfg.RecargoDominicalNocturnoPorcentaje)
	cfg.HoraExtraDominicalDiurnaPorcentaje = normalizeNominaPorcentaje(cfg.HoraExtraDominicalDiurnaPorcentaje)
	cfg.HoraExtraDominicalNocturnaPorcentaje = normalizeNominaPorcentaje(cfg.HoraExtraDominicalNocturnaPorcentaje)
	cfg.DeduccionSaludPorcentaje = normalizeNominaPorcentaje(cfg.DeduccionSaludPorcentaje)
	cfg.DeduccionPensionPorcentaje = normalizeNominaPorcentaje(cfg.DeduccionPensionPorcentaje)
	cfg.DeduccionFondoSolidaridadPorcentaje = normalizeNominaPorcentaje(cfg.DeduccionFondoSolidaridadPorcentaje)
	cfg.AporteSaludEmpleadorPorcentaje = normalizeNominaPorcentaje(cfg.AporteSaludEmpleadorPorcentaje)
	cfg.AportePensionEmpleadorPorcentaje = normalizeNominaPorcentaje(cfg.AportePensionEmpleadorPorcentaje)
	cfg.AporteARLPorcentaje = normalizeNominaPorcentaje(cfg.AporteARLPorcentaje)
	cfg.AporteCajaCompensacionPorcentaje = normalizeNominaPorcentaje(cfg.AporteCajaCompensacionPorcentaje)
	cfg.AporteICBFPorcentaje = normalizeNominaPorcentaje(cfg.AporteICBFPorcentaje)
	cfg.AporteSENAPorcentaje = normalizeNominaPorcentaje(cfg.AporteSENAPorcentaje)
	cfg.ProvisionCesantiasPorcentaje = normalizeNominaPorcentaje(cfg.ProvisionCesantiasPorcentaje)
	cfg.ProvisionInteresesCesantiasPorcentaje = normalizeNominaPorcentaje(cfg.ProvisionInteresesCesantiasPorcentaje)
	cfg.ProvisionPrimaPorcentaje = normalizeNominaPorcentaje(cfg.ProvisionPrimaPorcentaje)
	cfg.ProvisionVacacionesPorcentaje = normalizeNominaPorcentaje(cfg.ProvisionVacacionesPorcentaje)
	cfg.Estado = normalizeNominaEstado(cfg.Estado)
	return &cfg
}

// UpsertEmpresaNominaConfiguracion crea o actualiza la configuracion de nomina por empresa.
func UpsertEmpresaNominaConfiguracion(dbConn *sql.DB, payload EmpresaNominaConfiguracion) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	cfg := normalizeEmpresaNominaConfiguracion(payload)

	var existingID int64
	err := dbConn.QueryRow(`SELECT id FROM empresa_nomina_configuracion WHERE empresa_id = ? LIMIT 1`, cfg.EmpresaID).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	if existingID > 0 {
		_, err = dbConn.Exec(`UPDATE empresa_nomina_configuracion
		SET
			pais_codigo = ?,
			moneda = ?,
			horas_ordinarias_semana = ?,
			horas_ordinarias_dia = ?,
			dias_nomina_mes = ?,
			divisor_hora_ordinaria = ?,
			hora_nocturna_desde = ?,
			hora_nocturna_hasta = ?,
			recargo_nocturno_porcentaje = ?,
			hora_extra_diurna_porcentaje = ?,
			hora_extra_nocturna_porcentaje = ?,
			recargo_dominical_diurno_porcentaje = ?,
			recargo_dominical_nocturno_porcentaje = ?,
			hora_extra_dominical_diurna_porcentaje = ?,
			hora_extra_dominical_nocturna_porcentaje = ?,
			deduccion_salud_porcentaje = ?,
			deduccion_pension_porcentaje = ?,
			deduccion_fondo_solidaridad_porcentaje = ?,
			aporte_salud_empleador_porcentaje = ?,
			aporte_pension_empleador_porcentaje = ?,
			aporte_arl_porcentaje = ?,
			aporte_caja_compensacion_porcentaje = ?,
			aporte_icbf_porcentaje = ?,
			aporte_sena_porcentaje = ?,
			provision_cesantias_porcentaje = ?,
			provision_intereses_cesantias_porcentaje = ?,
			provision_prima_porcentaje = ?,
			provision_vacaciones_porcentaje = ?,
			usuario_creador = ?,
			estado = ?,
			observaciones = ?,
			fecha_actualizacion = datetime('now','localtime')
		WHERE empresa_id = ?`,
			cfg.PaisCodigo,
			cfg.Moneda,
			cfg.HorasOrdinariasSemana,
			cfg.HorasOrdinariasDia,
			cfg.DiasNominaMes,
			cfg.DivisorHoraOrdinaria,
			cfg.HoraNocturnaDesde,
			cfg.HoraNocturnaHasta,
			cfg.RecargoNocturnoPorcentaje,
			cfg.HoraExtraDiurnaPorcentaje,
			cfg.HoraExtraNocturnaPorcentaje,
			cfg.RecargoDominicalDiurnoPorcentaje,
			cfg.RecargoDominicalNocturnoPorcentaje,
			cfg.HoraExtraDominicalDiurnaPorcentaje,
			cfg.HoraExtraDominicalNocturnaPorcentaje,
			cfg.DeduccionSaludPorcentaje,
			cfg.DeduccionPensionPorcentaje,
			cfg.DeduccionFondoSolidaridadPorcentaje,
			cfg.AporteSaludEmpleadorPorcentaje,
			cfg.AportePensionEmpleadorPorcentaje,
			cfg.AporteARLPorcentaje,
			cfg.AporteCajaCompensacionPorcentaje,
			cfg.AporteICBFPorcentaje,
			cfg.AporteSENAPorcentaje,
			cfg.ProvisionCesantiasPorcentaje,
			cfg.ProvisionInteresesCesantiasPorcentaje,
			cfg.ProvisionPrimaPorcentaje,
			cfg.ProvisionVacacionesPorcentaje,
			strings.TrimSpace(cfg.UsuarioCreador),
			cfg.Estado,
			strings.TrimSpace(cfg.Observaciones),
			cfg.EmpresaID,
		)
		if err != nil {
			return 0, err
		}
		return existingID, nil
	}

	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_nomina_configuracion (
		empresa_id,
		pais_codigo,
		moneda,
		horas_ordinarias_semana,
		horas_ordinarias_dia,
		dias_nomina_mes,
		divisor_hora_ordinaria,
		hora_nocturna_desde,
		hora_nocturna_hasta,
		recargo_nocturno_porcentaje,
		hora_extra_diurna_porcentaje,
		hora_extra_nocturna_porcentaje,
		recargo_dominical_diurno_porcentaje,
		recargo_dominical_nocturno_porcentaje,
		hora_extra_dominical_diurna_porcentaje,
		hora_extra_dominical_nocturna_porcentaje,
		deduccion_salud_porcentaje,
		deduccion_pension_porcentaje,
		deduccion_fondo_solidaridad_porcentaje,
		aporte_salud_empleador_porcentaje,
		aporte_pension_empleador_porcentaje,
		aporte_arl_porcentaje,
		aporte_caja_compensacion_porcentaje,
		aporte_icbf_porcentaje,
		aporte_sena_porcentaje,
		provision_cesantias_porcentaje,
		provision_intereses_cesantias_porcentaje,
		provision_prima_porcentaje,
		provision_vacaciones_porcentaje,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
		cfg.EmpresaID,
		cfg.PaisCodigo,
		cfg.Moneda,
		cfg.HorasOrdinariasSemana,
		cfg.HorasOrdinariasDia,
		cfg.DiasNominaMes,
		cfg.DivisorHoraOrdinaria,
		cfg.HoraNocturnaDesde,
		cfg.HoraNocturnaHasta,
		cfg.RecargoNocturnoPorcentaje,
		cfg.HoraExtraDiurnaPorcentaje,
		cfg.HoraExtraNocturnaPorcentaje,
		cfg.RecargoDominicalDiurnoPorcentaje,
		cfg.RecargoDominicalNocturnoPorcentaje,
		cfg.HoraExtraDominicalDiurnaPorcentaje,
		cfg.HoraExtraDominicalNocturnaPorcentaje,
		cfg.DeduccionSaludPorcentaje,
		cfg.DeduccionPensionPorcentaje,
		cfg.DeduccionFondoSolidaridadPorcentaje,
		cfg.AporteSaludEmpleadorPorcentaje,
		cfg.AportePensionEmpleadorPorcentaje,
		cfg.AporteARLPorcentaje,
		cfg.AporteCajaCompensacionPorcentaje,
		cfg.AporteICBFPorcentaje,
		cfg.AporteSENAPorcentaje,
		cfg.ProvisionCesantiasPorcentaje,
		cfg.ProvisionInteresesCesantiasPorcentaje,
		cfg.ProvisionPrimaPorcentaje,
		cfg.ProvisionVacacionesPorcentaje,
		strings.TrimSpace(cfg.UsuarioCreador),
		cfg.Estado,
		strings.TrimSpace(cfg.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// CreateEmpresaNominaEmpleado crea una ficha de empleado para nomina.
func CreateEmpresaNominaEmpleado(dbConn *sql.DB, payload EmpresaNominaEmpleado) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	payload.EmpleadoNombre = strings.TrimSpace(payload.EmpleadoNombre)
	if payload.EmpleadoNombre == "" {
		return 0, fmt.Errorf("empleado_nombre es obligatorio")
	}
	if payload.SalarioBasicoMensual < 0 {
		payload.SalarioBasicoMensual = 0
	}
	if payload.AuxilioTransporteMensual < 0 {
		payload.AuxilioTransporteMensual = 0
	}
	if payload.BonificacionFijaMensual < 0 {
		payload.BonificacionFijaMensual = 0
	}
	if payload.DeduccionFijaMensual < 0 {
		payload.DeduccionFijaMensual = 0
	}
	payload.JornadaHorasDia = normalizeNominaHoras(payload.JornadaHorasDia, 8)
	payload.TipoContrato = normalizeNominaTipoContrato(payload.TipoContrato)
	payload.Estado = normalizeNominaEstado(payload.Estado)
	if payload.FechaIngreso != "" {
		date, err := normalizeNominaDate(payload.FechaIngreso)
		if err != nil {
			return 0, fmt.Errorf("fecha_ingreso invalida (use YYYY-MM-DD)")
		}
		payload.FechaIngreso = date
	}

	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_nomina_empleados (
		empresa_id,
		empleado_id,
		empleado_codigo,
		empleado_nombre,
		empleado_documento,
		cargo,
		sede_codigo,
		sede_nombre,
		centro_costo,
		tipo_contrato,
		fecha_ingreso,
		salario_basico_mensual,
		auxilio_transporte_mensual,
		bonificacion_fija_mensual,
		deduccion_fija_mensual,
		jornada_horas_dia,
		incluir_auxilio_transporte,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
		payload.EmpresaID,
		payload.EmpleadoID,
		strings.TrimSpace(payload.EmpleadoCodigo),
		payload.EmpleadoNombre,
		strings.TrimSpace(payload.EmpleadoDocumento),
		strings.TrimSpace(payload.Cargo),
		strings.TrimSpace(payload.SedeCodigo),
		strings.TrimSpace(payload.SedeNombre),
		strings.TrimSpace(payload.CentroCosto),
		payload.TipoContrato,
		strings.TrimSpace(payload.FechaIngreso),
		round2(payload.SalarioBasicoMensual),
		round2(payload.AuxilioTransporteMensual),
		round2(payload.BonificacionFijaMensual),
		round2(payload.DeduccionFijaMensual),
		payload.JornadaHorasDia,
		boolToInt(payload.IncluirAuxilioTransporte),
		strings.TrimSpace(payload.UsuarioCreador),
		payload.Estado,
		strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateEmpresaNominaEmpleado actualiza una ficha de empleado de nomina.
func UpdateEmpresaNominaEmpleado(dbConn *sql.DB, payload EmpresaNominaEmpleado) error {
	if payload.EmpresaID <= 0 || payload.ID <= 0 {
		return fmt.Errorf("empresa_id e id son obligatorios")
	}
	payload.EmpleadoNombre = strings.TrimSpace(payload.EmpleadoNombre)
	if payload.EmpleadoNombre == "" {
		return fmt.Errorf("empleado_nombre es obligatorio")
	}
	if payload.SalarioBasicoMensual < 0 {
		payload.SalarioBasicoMensual = 0
	}
	if payload.AuxilioTransporteMensual < 0 {
		payload.AuxilioTransporteMensual = 0
	}
	if payload.BonificacionFijaMensual < 0 {
		payload.BonificacionFijaMensual = 0
	}
	if payload.DeduccionFijaMensual < 0 {
		payload.DeduccionFijaMensual = 0
	}
	payload.JornadaHorasDia = normalizeNominaHoras(payload.JornadaHorasDia, 8)
	payload.TipoContrato = normalizeNominaTipoContrato(payload.TipoContrato)
	if payload.FechaIngreso != "" {
		date, err := normalizeNominaDate(payload.FechaIngreso)
		if err != nil {
			return fmt.Errorf("fecha_ingreso invalida (use YYYY-MM-DD)")
		}
		payload.FechaIngreso = date
	}

	res, err := dbConn.Exec(`UPDATE empresa_nomina_empleados
	SET
		empleado_id = ?,
		empleado_codigo = ?,
		empleado_nombre = ?,
		empleado_documento = ?,
		cargo = ?,
		sede_codigo = ?,
		sede_nombre = ?,
		centro_costo = ?,
		tipo_contrato = ?,
		fecha_ingreso = ?,
		salario_basico_mensual = ?,
		auxilio_transporte_mensual = ?,
		bonificacion_fija_mensual = ?,
		deduccion_fija_mensual = ?,
		jornada_horas_dia = ?,
		incluir_auxilio_transporte = ?,
		observaciones = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`,
		payload.EmpleadoID,
		strings.TrimSpace(payload.EmpleadoCodigo),
		payload.EmpleadoNombre,
		strings.TrimSpace(payload.EmpleadoDocumento),
		strings.TrimSpace(payload.Cargo),
		strings.TrimSpace(payload.SedeCodigo),
		strings.TrimSpace(payload.SedeNombre),
		strings.TrimSpace(payload.CentroCosto),
		payload.TipoContrato,
		strings.TrimSpace(payload.FechaIngreso),
		round2(payload.SalarioBasicoMensual),
		round2(payload.AuxilioTransporteMensual),
		round2(payload.BonificacionFijaMensual),
		round2(payload.DeduccionFijaMensual),
		payload.JornadaHorasDia,
		boolToInt(payload.IncluirAuxilioTransporte),
		strings.TrimSpace(payload.Observaciones),
		payload.EmpresaID,
		payload.ID,
	)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SetEmpresaNominaEmpleadoEstado activa o desactiva un empleado de nomina.
func SetEmpresaNominaEmpleadoEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	next := normalizeNominaEstado(estado)
	res, err := dbConn.Exec(`UPDATE empresa_nomina_empleados
	SET estado = ?, fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`, next, empresaID, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteEmpresaNominaEmpleado elimina una ficha de empleado de nomina.
func DeleteEmpresaNominaEmpleado(dbConn *sql.DB, empresaID, id int64) error {
	res, err := dbConn.Exec(`DELETE FROM empresa_nomina_empleados WHERE empresa_id = ? AND id = ?`, empresaID, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ListEmpresaNominaEmpleados lista empleados de nomina por empresa con filtros.
func ListEmpresaNominaEmpleados(dbConn *sql.DB, empresaID int64, includeInactive bool, q string, limit int) ([]EmpresaNominaEmpleado, error) {
	query := `SELECT
		id,
		empresa_id,
		COALESCE(empleado_id, 0),
		COALESCE(empleado_codigo, ''),
		COALESCE(empleado_nombre, ''),
		COALESCE(empleado_documento, ''),
		COALESCE(cargo, ''),
		COALESCE(sede_codigo, ''),
		COALESCE(sede_nombre, ''),
		COALESCE(centro_costo, ''),
		COALESCE(tipo_contrato, 'indefinido'),
		COALESCE(fecha_ingreso, ''),
		COALESCE(salario_basico_mensual, 0),
		COALESCE(auxilio_transporte_mensual, 0),
		COALESCE(bonificacion_fija_mensual, 0),
		COALESCE(deduccion_fija_mensual, 0),
		COALESCE(jornada_horas_dia, 8),
		COALESCE(incluir_auxilio_transporte, 1),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_nomina_empleados
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	if !includeInactive {
		query += ` AND estado = 'activo'`
	}

	q = strings.ToLower(strings.TrimSpace(q))
	if q != "" {
		query += ` AND (
			LOWER(COALESCE(empleado_codigo, '')) LIKE ?
			OR LOWER(COALESCE(empleado_nombre, '')) LIKE ?
			OR LOWER(COALESCE(empleado_documento, '')) LIKE ?
			OR LOWER(COALESCE(cargo, '')) LIKE ?
			OR LOWER(COALESCE(sede_codigo, '')) LIKE ?
			OR LOWER(COALESCE(sede_nombre, '')) LIKE ?
			OR LOWER(COALESCE(centro_costo, '')) LIKE ?
		)`
		like := "%" + q + "%"
		args = append(args, like, like, like, like, like, like, like)
	}

	if limit <= 0 {
		limit = 300
	}
	if limit > 2000 {
		limit = 2000
	}
	query += ` ORDER BY COALESCE(empleado_nombre, '') ASC, id ASC LIMIT ?`
	args = append(args, limit)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaNominaEmpleado, 0)
	for rows.Next() {
		var item EmpresaNominaEmpleado
		var incluirAuxInt int
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.EmpleadoID,
			&item.EmpleadoCodigo,
			&item.EmpleadoNombre,
			&item.EmpleadoDocumento,
			&item.Cargo,
			&item.SedeCodigo,
			&item.SedeNombre,
			&item.CentroCosto,
			&item.TipoContrato,
			&item.FechaIngreso,
			&item.SalarioBasicoMensual,
			&item.AuxilioTransporteMensual,
			&item.BonificacionFijaMensual,
			&item.DeduccionFijaMensual,
			&item.JornadaHorasDia,
			&incluirAuxInt,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		item.IncluirAuxilioTransporte = incluirAuxInt != 0
		item.TipoContrato = normalizeNominaTipoContrato(item.TipoContrato)
		item.Estado = normalizeNominaEstado(item.Estado)
		item.SalarioBasicoMensual = round2(item.SalarioBasicoMensual)
		item.AuxilioTransporteMensual = round2(item.AuxilioTransporteMensual)
		item.BonificacionFijaMensual = round2(item.BonificacionFijaMensual)
		item.DeduccionFijaMensual = round2(item.DeduccionFijaMensual)
		item.JornadaHorasDia = normalizeNominaHoras(item.JornadaHorasDia, 8)
		out = append(out, item)
	}
	return out, rows.Err()
}

func getEmpresaNominaEmpleadoByID(dbConn *sql.DB, empresaID, id int64) (*EmpresaNominaEmpleado, error) {
	rows, err := ListEmpresaNominaEmpleados(dbConn, empresaID, true, "", 2000)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row.ID == id {
			copy := row
			return &copy, nil
		}
	}
	return nil, sql.ErrNoRows
}

// CreateEmpresaNominaFestivo registra un dia festivo por empresa.
func CreateEmpresaNominaFestivo(dbConn *sql.DB, payload EmpresaNominaFestivo) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	date, err := normalizeNominaDate(payload.FechaFestivo)
	if err != nil {
		return 0, fmt.Errorf("fecha_festivo invalida (use YYYY-MM-DD)")
	}
	payload.FechaFestivo = date
	payload.Estado = normalizeNominaEstado(payload.Estado)

	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_nomina_festivos (
		empresa_id,
		fecha_festivo,
		descripcion,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
		payload.EmpresaID,
		payload.FechaFestivo,
		strings.TrimSpace(payload.Descripcion),
		strings.TrimSpace(payload.UsuarioCreador),
		payload.Estado,
		strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return 0, fmt.Errorf("ya existe un festivo registrado para esa fecha")
		}
		return 0, err
	}
	return id, nil
}

// DeleteEmpresaNominaFestivo elimina un festivo por empresa.
func DeleteEmpresaNominaFestivo(dbConn *sql.DB, empresaID, id int64) error {
	res, err := dbConn.Exec(`DELETE FROM empresa_nomina_festivos WHERE empresa_id = ? AND id = ?`, empresaID, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ListEmpresaNominaFestivos lista festivos configurados por empresa.
func ListEmpresaNominaFestivos(dbConn *sql.DB, empresaID int64, includeInactive bool, desde, hasta string, limit int) ([]EmpresaNominaFestivo, error) {
	query := `SELECT
		id,
		empresa_id,
		COALESCE(fecha_festivo, ''),
		COALESCE(descripcion, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_nomina_festivos
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if !includeInactive {
		query += ` AND estado = 'activo'`
	}
	if strings.TrimSpace(desde) != "" {
		query += ` AND fecha_festivo >= ?`
		args = append(args, strings.TrimSpace(desde))
	}
	if strings.TrimSpace(hasta) != "" {
		query += ` AND fecha_festivo <= ?`
		args = append(args, strings.TrimSpace(hasta))
	}
	if limit <= 0 {
		limit = 500
	}
	if limit > 3000 {
		limit = 3000
	}
	query += ` ORDER BY fecha_festivo DESC, id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaNominaFestivo, 0)
	for rows.Next() {
		var item EmpresaNominaFestivo
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.FechaFestivo,
			&item.Descripcion,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		item.Estado = normalizeNominaEstado(item.Estado)
		out = append(out, item)
	}
	return out, rows.Err()
}

func buildEmpresaNominaFestivoSet(dbConn *sql.DB, empresaID int64, desde, hasta string) (map[string]bool, error) {
	rows, err := ListEmpresaNominaFestivos(dbConn, empresaID, false, desde, hasta, 5000)
	if err != nil {
		return nil, err
	}
	set := make(map[string]bool)
	for _, row := range rows {
		if row.FechaFestivo != "" {
			set[row.FechaFestivo] = true
		}
	}
	return set, nil
}

func listAsistenciaRowsForNomina(dbConn *sql.DB, empresaID int64, empleado EmpresaNominaEmpleado, desde, hasta string) ([]nominaAsistenciaRow, error) {
	query := `SELECT
		COALESCE(fecha_asistencia, ''),
		COALESCE(hora_entrada, ''),
		COALESCE(hora_salida, ''),
		COALESCE(horas_trabajadas, 0),
		COALESCE(estado_asistencia, 'pendiente')
	FROM empresa_asistencia_empleados
	WHERE empresa_id = ?
		AND estado = 'activo'
		AND fecha_asistencia >= ?
		AND fecha_asistencia <= ?`
	args := []interface{}{empresaID, desde, hasta}

	if empleado.EmpleadoID > 0 {
		query += ` AND COALESCE(empleado_id, 0) = ?`
		args = append(args, empleado.EmpleadoID)
	} else if strings.TrimSpace(empleado.EmpleadoDocumento) != "" {
		query += ` AND LOWER(COALESCE(empleado_documento, '')) = ?`
		args = append(args, strings.ToLower(strings.TrimSpace(empleado.EmpleadoDocumento)))
	} else if strings.TrimSpace(empleado.EmpleadoCodigo) != "" {
		query += ` AND LOWER(COALESCE(empleado_codigo, '')) = ?`
		args = append(args, strings.ToLower(strings.TrimSpace(empleado.EmpleadoCodigo)))
	} else {
		query += ` AND LOWER(COALESCE(empleado_nombre, '')) = ?`
		args = append(args, strings.ToLower(strings.TrimSpace(empleado.EmpleadoNombre)))
	}

	query += ` ORDER BY fecha_asistencia ASC, id ASC`
	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]nominaAsistenciaRow, 0)
	for rows.Next() {
		var item nominaAsistenciaRow
		if err := rows.Scan(&item.FechaAsistencia, &item.HoraEntrada, &item.HoraSalida, &item.HorasTrabajadas, &item.EstadoAsistencia); err != nil {
			return nil, err
		}
		item.EstadoAsistencia = strings.ToLower(strings.TrimSpace(item.EstadoAsistencia))
		if item.EstadoAsistencia == "" {
			item.EstadoAsistencia = "pendiente"
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func isNominaEstadoLaborado(estado string) bool {
	switch strings.ToLower(strings.TrimSpace(estado)) {
	case "ausente":
		return false
	default:
		return true
	}
}

func parseNominaHourMinute(raw string, fallback int) int {
	t, err := normalizeAsistenciaTime(raw)
	if err != nil || t == "" {
		return fallback
	}
	parts := strings.Split(t, ":")
	if len(parts) < 2 {
		return fallback
	}
	hh := 0
	mm := 0
	fmt.Sscanf(parts[0], "%d", &hh)
	fmt.Sscanf(parts[1], "%d", &mm)
	if hh < 0 {
		hh = 0
	}
	if hh > 23 {
		hh = 23
	}
	if mm < 0 {
		mm = 0
	}
	if mm > 59 {
		mm = 59
	}
	return hh*60 + mm
}

func isNominaNightMinute(at time.Time, cfg *EmpresaNominaConfiguracion) bool {
	startMinute := parseNominaHourMinute(cfg.HoraNocturnaDesde, 21*60)
	endMinute := parseNominaHourMinute(cfg.HoraNocturnaHasta, 6*60)
	minute := at.Hour()*60 + at.Minute()
	if startMinute == endMinute {
		return false
	}
	if startMinute < endMinute {
		return minute >= startMinute && minute < endMinute
	}
	return minute >= startMinute || minute < endMinute
}

func isNominaSpecialDay(dateKey string, date time.Time, festivos map[string]bool) bool {
	if date.Weekday() == time.Sunday {
		return true
	}
	return festivos[dateKey]
}

func accrueNominaMinuto(stats *nominaHorasMinutos, dateKey string, ordinaryUsedByDate map[string]int, cfg *EmpresaNominaConfiguracion, isSpecialDay bool, isNight bool) {
	if stats.dias == nil {
		stats.dias = make(map[string]struct{})
	}
	stats.dias[dateKey] = struct{}{}
	stats.totalMinutos += 1

	limit := int(cfg.HorasOrdinariasDia*60 + 0.5)
	if limit <= 0 {
		limit = 8 * 60
	}
	used := ordinaryUsedByDate[dateKey]
	isExtra := used >= limit
	if !isExtra {
		ordinaryUsedByDate[dateKey] = used + 1
	}

	if isExtra {
		if isSpecialDay {
			if isNight {
				stats.extraDominicalNocturnaMin += 1
			} else {
				stats.extraDominicalDiurnaMin += 1
			}
			return
		}
		if isNight {
			stats.extraNocturnaMin += 1
		} else {
			stats.extraDiurnaMin += 1
		}
		return
	}

	if isSpecialDay {
		if isNight {
			stats.dominicalNocturnaMin += 1
		} else {
			stats.dominicalDiurnaMin += 1
		}
		return
	}

	if isNight {
		stats.recargoNocturnoMin += 1
	} else {
		stats.ordinariasMin += 1
	}
}

func minutesToHours(min int) float64 {
	if min <= 0 {
		return 0
	}
	return round2(float64(min) / 60.0)
}

func nominaFloatEqual(a, b float64) bool {
	if a == b {
		return true
	}
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= 0.01
}

func nominaLiquidacionAlineadaConDetalle(liq EmpresaNominaLiquidacion, detail nominaHorasDetalle) bool {
	if liq.RegistrosAsistencia != detail.RegistrosAsistencia {
		return false
	}
	if !nominaFloatEqual(liq.HorasAsistenciaTotal, detail.HorasAsistenciaTotal) {
		return false
	}
	if !nominaFloatEqual(liq.HorasOrdinarias, detail.HorasOrdinarias) {
		return false
	}
	if !nominaFloatEqual(liq.HorasRecargoNocturno, detail.HorasRecargoNocturno) {
		return false
	}
	if !nominaFloatEqual(liq.HorasExtraDiurnas, detail.HorasExtraDiurnas) {
		return false
	}
	if !nominaFloatEqual(liq.HorasExtraNocturnas, detail.HorasExtraNocturnas) {
		return false
	}
	if !nominaFloatEqual(liq.HorasDominicalesDiurnas, detail.HorasDominicalesDiurnas) {
		return false
	}
	if !nominaFloatEqual(liq.HorasDominicalesNocturnas, detail.HorasDominicalesNocturnas) {
		return false
	}
	if !nominaFloatEqual(liq.HorasExtraDominicalesDiurnas, detail.HorasExtraDominicalesDiurnas) {
		return false
	}
	if !nominaFloatEqual(liq.HorasExtraDominicalesNocturnas, detail.HorasExtraDominicalesNocturnas) {
		return false
	}
	return true
}

func buildNominaHorasDetalle(rows []nominaAsistenciaRow, cfg *EmpresaNominaConfiguracion, festivos map[string]bool) nominaHorasDetalle {
	stats := nominaHorasMinutos{
		dias: make(map[string]struct{}),
	}
	ordinaryUsedByDate := make(map[string]int)

	for _, row := range rows {
		if !isNominaEstadoLaborado(row.EstadoAsistencia) {
			continue
		}
		stats.registros += 1

		fecha, err := normalizeNominaDate(row.FechaAsistencia)
		if err != nil {
			continue
		}
		dateBase, err := time.Parse("2006-01-02", fecha)
		if err != nil {
			continue
		}

		entrada, entradaErr := normalizeAsistenciaTime(row.HoraEntrada)
		salida, salidaErr := normalizeAsistenciaTime(row.HoraSalida)
		if entradaErr == nil && salidaErr == nil && entrada != "" && salida != "" {
			start, startErr := time.Parse("2006-01-02 15:04:05", fecha+" "+entrada)
			end, endErr := time.Parse("2006-01-02 15:04:05", fecha+" "+salida)
			if startErr == nil && endErr == nil {
				if end.Before(start) {
					end = end.Add(24 * time.Hour)
				}
				if end.After(start.Add(48 * time.Hour)) {
					end = start.Add(48 * time.Hour)
				}
				for cursor := start; cursor.Before(end); cursor = cursor.Add(time.Minute) {
					dateKey := cursor.Format("2006-01-02")
					isSpecial := isNominaSpecialDay(dateKey, cursor, festivos)
					isNight := isNominaNightMinute(cursor, cfg)
					accrueNominaMinuto(&stats, dateKey, ordinaryUsedByDate, cfg, isSpecial, isNight)
				}
				continue
			}
		}

		minutes := int(row.HorasTrabajadas*60 + 0.5)
		if minutes <= 0 {
			continue
		}
		dateKey := dateBase.Format("2006-01-02")
		isSpecial := isNominaSpecialDay(dateKey, dateBase, festivos)
		for i := 0; i < minutes; i++ {
			accrueNominaMinuto(&stats, dateKey, ordinaryUsedByDate, cfg, isSpecial, false)
		}
	}

	detail := nominaHorasDetalle{
		DiasLiquidados:                 float64(len(stats.dias)),
		HorasAsistenciaTotal:           minutesToHours(stats.totalMinutos),
		RegistrosAsistencia:            stats.registros,
		HorasOrdinarias:                minutesToHours(stats.ordinariasMin),
		HorasRecargoNocturno:           minutesToHours(stats.recargoNocturnoMin),
		HorasExtraDiurnas:              minutesToHours(stats.extraDiurnaMin),
		HorasExtraNocturnas:            minutesToHours(stats.extraNocturnaMin),
		HorasDominicalesDiurnas:        minutesToHours(stats.dominicalDiurnaMin),
		HorasDominicalesNocturnas:      minutesToHours(stats.dominicalNocturnaMin),
		HorasExtraDominicalesDiurnas:   minutesToHours(stats.extraDominicalDiurnaMin),
		HorasExtraDominicalesNocturnas: minutesToHours(stats.extraDominicalNocturnaMin),
	}
	return detail
}

func buildNominaLiquidacion(
	empleado EmpresaNominaEmpleado,
	cfg *EmpresaNominaConfiguracion,
	req EmpresaNominaCalculoRequest,
	detail nominaHorasDetalle,
) EmpresaNominaLiquidacion {
	divisor := cfg.DivisorHoraOrdinaria
	if divisor <= 0 {
		divisor = recommendedNominaHourDivisor(cfg.HorasOrdinariasSemana)
	}
	valorHora := 0.0
	if divisor > 0 {
		valorHora = round2(empleado.SalarioBasicoMensual / divisor)
	}

	diasNomina := cfg.DiasNominaMes
	if diasNomina <= 0 {
		diasNomina = 30
	}
	factorDias := 0.0
	if diasNomina > 0 {
		factorDias = detail.DiasLiquidados / float64(diasNomina)
	}
	if factorDias < 0 {
		factorDias = 0
	}
	if factorDias > 1 {
		factorDias = 1
	}

	baseSalario := round2(empleado.SalarioBasicoMensual * factorDias)
	auxilio := 0.0
	if empleado.IncluirAuxilioTransporte {
		auxilio = round2(empleado.AuxilioTransporteMensual * factorDias)
	}
	bonificacion := round2(empleado.BonificacionFijaMensual * factorDias)
	deduccionFija := round2(empleado.DeduccionFijaMensual * factorDias)
	otrasDeducciones := round2(req.OtrasDeducciones)

	valorRecargoNocturno := round2(detail.HorasRecargoNocturno * valorHora * (cfg.RecargoNocturnoPorcentaje / 100.0))
	valorDominicalDiurno := round2(detail.HorasDominicalesDiurnas * valorHora * (cfg.RecargoDominicalDiurnoPorcentaje / 100.0))
	valorDominicalNocturno := round2(detail.HorasDominicalesNocturnas * valorHora * (cfg.RecargoDominicalNocturnoPorcentaje / 100.0))
	valorExtraDiurna := round2(detail.HorasExtraDiurnas * valorHora * (1 + cfg.HoraExtraDiurnaPorcentaje/100.0))
	valorExtraNocturna := round2(detail.HorasExtraNocturnas * valorHora * (1 + cfg.HoraExtraNocturnaPorcentaje/100.0))
	valorExtraDominicalDiurna := round2(detail.HorasExtraDominicalesDiurnas * valorHora * (1 + cfg.HoraExtraDominicalDiurnaPorcentaje/100.0))
	valorExtraDominicalNocturna := round2(detail.HorasExtraDominicalesNocturnas * valorHora * (1 + cfg.HoraExtraDominicalNocturnaPorcentaje/100.0))
	totalRecargos := round2(valorRecargoNocturno + valorDominicalDiurno + valorDominicalNocturno + valorExtraDiurna + valorExtraNocturna + valorExtraDominicalDiurna + valorExtraDominicalNocturna)

	devengado := round2(baseSalario + totalRecargos + auxilio + bonificacion)
	ibc := round2(baseSalario + totalRecargos + bonificacion)
	dedSalud := round2(ibc * (cfg.DeduccionSaludPorcentaje / 100.0))
	dedPension := round2(ibc * (cfg.DeduccionPensionPorcentaje / 100.0))
	dedFondo := round2(ibc * (cfg.DeduccionFondoSolidaridadPorcentaje / 100.0))
	dedTotal := round2(dedSalud + dedPension + dedFondo + deduccionFija + otrasDeducciones)
	neto := round2(devengado - dedTotal)

	resumenJSON := fmt.Sprintf(`{"asistencia_registros":%d,"dias_liquidados":%.2f,"horas_totales":%.2f}`, detail.RegistrosAsistencia, detail.DiasLiquidados, detail.HorasAsistenciaTotal)

	return EmpresaNominaLiquidacion{
		EmpresaID:                      empleado.EmpresaID,
		EmpleadoNominaID:               empleado.ID,
		EmpleadoID:                     empleado.EmpleadoID,
		EmpleadoCodigo:                 strings.TrimSpace(empleado.EmpleadoCodigo),
		EmpleadoNombre:                 empleado.EmpleadoNombre,
		EmpleadoDocumento:              strings.TrimSpace(empleado.EmpleadoDocumento),
		Cargo:                          strings.TrimSpace(empleado.Cargo),
		SedeCodigo:                     strings.TrimSpace(empleado.SedeCodigo),
		SedeNombre:                     strings.TrimSpace(empleado.SedeNombre),
		CentroCosto:                    strings.TrimSpace(empleado.CentroCosto),
		PeriodoDesde:                   req.PeriodoDesde,
		PeriodoHasta:                   req.PeriodoHasta,
		DiasLiquidados:                 round2(detail.DiasLiquidados),
		HorasAsistenciaTotal:           round2(detail.HorasAsistenciaTotal),
		RegistrosAsistencia:            detail.RegistrosAsistencia,
		HorasOrdinarias:                round2(detail.HorasOrdinarias),
		HorasRecargoNocturno:           round2(detail.HorasRecargoNocturno),
		HorasExtraDiurnas:              round2(detail.HorasExtraDiurnas),
		HorasExtraNocturnas:            round2(detail.HorasExtraNocturnas),
		HorasDominicalesDiurnas:        round2(detail.HorasDominicalesDiurnas),
		HorasDominicalesNocturnas:      round2(detail.HorasDominicalesNocturnas),
		HorasExtraDominicalesDiurnas:   round2(detail.HorasExtraDominicalesDiurnas),
		HorasExtraDominicalesNocturnas: round2(detail.HorasExtraDominicalesNocturnas),
		ValorHoraOrdinaria:             valorHora,
		BaseSalarioProporcional:        baseSalario,
		ValorRecargoNocturno:           valorRecargoNocturno,
		ValorDominicalDiurno:           valorDominicalDiurno,
		ValorDominicalNocturno:         valorDominicalNocturno,
		ValorExtraDiurna:               valorExtraDiurna,
		ValorExtraNocturna:             valorExtraNocturna,
		ValorExtraDominicalDiurna:      valorExtraDominicalDiurna,
		ValorExtraDominicalNocturna:    valorExtraDominicalNocturna,
		TotalRecargosHorasExtras:       totalRecargos,
		AuxilioTransporte:              auxilio,
		Bonificacion:                   bonificacion,
		ComisionesServicioTotal:        0,
		ComisionesServicioMovimientos:  0,
		ComisionesServicioAjustes:      0,
		DevengadoTotal:                 devengado,
		IngresoBaseCotizacion:          ibc,
		DeduccionSalud:                 dedSalud,
		DeduccionPension:               dedPension,
		DeduccionFondoSolidaridad:      dedFondo,
		DeduccionFija:                  deduccionFija,
		OtrasDeducciones:               otrasDeducciones,
		DeduccionTotal:                 dedTotal,
		NetoPagar:                      neto,
		OrigenCalculo:                  "asistencia",
		ResumenJSON:                    resumenJSON,
		UsuarioCreador:                 strings.TrimSpace(req.UsuarioCreador),
		Estado:                         nominaEstadoActivo,
		Observaciones:                  strings.TrimSpace(req.Observaciones),
	}
}

func incorporarComisionesServicioEnLiquidacion(liq *EmpresaNominaLiquidacion, cfg *EmpresaNominaConfiguracion, resumen *EmpresaComisionServicioLiquidacionResumen) {
	if liq == nil || cfg == nil || resumen == nil {
		return
	}
	totalComisiones := round2(resumen.TotalComisiones)
	if totalComisiones == 0 {
		liq.ComisionesServicioTotal = 0
		liq.ComisionesServicioMovimientos = 0
		liq.ComisionesServicioAjustes = 0
		return
	}

	liq.ComisionesServicioTotal = totalComisiones
	liq.ComisionesServicioMovimientos = resumen.CantidadMovimientos
	liq.ComisionesServicioAjustes = round2(resumen.TotalAjustesManuales)

	devengado := round2(liq.BaseSalarioProporcional + liq.TotalRecargosHorasExtras + liq.AuxilioTransporte + liq.Bonificacion + totalComisiones)
	ibc := round2(liq.BaseSalarioProporcional + liq.TotalRecargosHorasExtras + liq.Bonificacion + totalComisiones)

	liq.DevengadoTotal = devengado
	liq.IngresoBaseCotizacion = ibc
	liq.DeduccionSalud = round2(ibc * (cfg.DeduccionSaludPorcentaje / 100.0))
	liq.DeduccionPension = round2(ibc * (cfg.DeduccionPensionPorcentaje / 100.0))
	liq.DeduccionFondoSolidaridad = round2(ibc * (cfg.DeduccionFondoSolidaridadPorcentaje / 100.0))
	liq.DeduccionTotal = round2(liq.DeduccionSalud + liq.DeduccionPension + liq.DeduccionFondoSolidaridad + liq.DeduccionFija + liq.OtrasDeducciones)
	liq.NetoPagar = round2(liq.DevengadoTotal - liq.DeduccionTotal)
}

func upsertEmpresaNominaLiquidacion(dbConn *sql.DB, payload EmpresaNominaLiquidacion, overwrite bool) (int64, error) {
	if payload.EmpresaID <= 0 || payload.EmpleadoNominaID <= 0 {
		return 0, fmt.Errorf("empresa_id y empleado_nomina_id son obligatorios")
	}
	if strings.TrimSpace(payload.PeriodoDesde) == "" || strings.TrimSpace(payload.PeriodoHasta) == "" {
		return 0, fmt.Errorf("periodo_desde y periodo_hasta son obligatorios")
	}

	var existingID int64
	err := dbConn.QueryRow(`SELECT id FROM empresa_nomina_liquidaciones WHERE empresa_id = ? AND empleado_nomina_id = ? AND periodo_desde = ? AND periodo_hasta = ? LIMIT 1`,
		payload.EmpresaID, payload.EmpleadoNominaID, payload.PeriodoDesde, payload.PeriodoHasta,
	).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	if existingID > 0 && !overwrite {
		return existingID, nil
	}

	if existingID > 0 {
		_, err = dbConn.Exec(`UPDATE empresa_nomina_liquidaciones
		SET
			empleado_id = ?,
			empleado_codigo = ?,
			empleado_nombre = ?,
			empleado_documento = ?,
			cargo = ?,
			sede_codigo = ?,
			sede_nombre = ?,
			centro_costo = ?,
			dias_liquidados = ?,
			horas_asistencia_total = ?,
			registros_asistencia = ?,
			horas_ordinarias = ?,
			horas_recargo_nocturno = ?,
			horas_extra_diurnas = ?,
			horas_extra_nocturnas = ?,
			horas_dominicales_diurnas = ?,
			horas_dominicales_nocturnas = ?,
			horas_extra_dominicales_diurnas = ?,
			horas_extra_dominicales_nocturnas = ?,
			valor_hora_ordinaria = ?,
			base_salario_proporcional = ?,
			valor_recargo_nocturno = ?,
			valor_dominical_diurno = ?,
			valor_dominical_nocturno = ?,
			valor_extra_diurna = ?,
			valor_extra_nocturna = ?,
			valor_extra_dominical_diurna = ?,
			valor_extra_dominical_nocturna = ?,
			total_recargos_horas_extras = ?,
			auxilio_transporte = ?,
			bonificacion = ?,
			comisiones_servicio_total = ?,
			comisiones_servicio_movimientos = ?,
			comisiones_servicio_ajustes = ?,
			devengado_total = ?,
			ingreso_base_cotizacion = ?,
			deduccion_salud = ?,
			deduccion_pension = ?,
			deduccion_fondo_solidaridad = ?,
			deduccion_fija = ?,
			otras_deducciones = ?,
			deduccion_total = ?,
			neto_pagar = ?,
			origen_calculo = ?,
			resumen_json = ?,
			fecha_generacion = datetime('now','localtime'),
			usuario_creador = ?,
			estado = ?,
			observaciones = ?,
			fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`,
			payload.EmpleadoID,
			strings.TrimSpace(payload.EmpleadoCodigo),
			payload.EmpleadoNombre,
			strings.TrimSpace(payload.EmpleadoDocumento),
			strings.TrimSpace(payload.Cargo),
			strings.TrimSpace(payload.SedeCodigo),
			strings.TrimSpace(payload.SedeNombre),
			strings.TrimSpace(payload.CentroCosto),
			payload.DiasLiquidados,
			payload.HorasAsistenciaTotal,
			payload.RegistrosAsistencia,
			payload.HorasOrdinarias,
			payload.HorasRecargoNocturno,
			payload.HorasExtraDiurnas,
			payload.HorasExtraNocturnas,
			payload.HorasDominicalesDiurnas,
			payload.HorasDominicalesNocturnas,
			payload.HorasExtraDominicalesDiurnas,
			payload.HorasExtraDominicalesNocturnas,
			payload.ValorHoraOrdinaria,
			payload.BaseSalarioProporcional,
			payload.ValorRecargoNocturno,
			payload.ValorDominicalDiurno,
			payload.ValorDominicalNocturno,
			payload.ValorExtraDiurna,
			payload.ValorExtraNocturna,
			payload.ValorExtraDominicalDiurna,
			payload.ValorExtraDominicalNocturna,
			payload.TotalRecargosHorasExtras,
			payload.AuxilioTransporte,
			payload.Bonificacion,
			payload.ComisionesServicioTotal,
			payload.ComisionesServicioMovimientos,
			payload.ComisionesServicioAjustes,
			payload.DevengadoTotal,
			payload.IngresoBaseCotizacion,
			payload.DeduccionSalud,
			payload.DeduccionPension,
			payload.DeduccionFondoSolidaridad,
			payload.DeduccionFija,
			payload.OtrasDeducciones,
			payload.DeduccionTotal,
			payload.NetoPagar,
			strings.TrimSpace(payload.OrigenCalculo),
			strings.TrimSpace(payload.ResumenJSON),
			strings.TrimSpace(payload.UsuarioCreador),
			normalizeNominaEstado(payload.Estado),
			strings.TrimSpace(payload.Observaciones),
			existingID,
			payload.EmpresaID,
		)
		if err != nil {
			return 0, err
		}
		return existingID, nil
	}

	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_nomina_liquidaciones (
		empresa_id,
		empleado_nomina_id,
		empleado_id,
		empleado_codigo,
		empleado_nombre,
		empleado_documento,
		cargo,
		sede_codigo,
		sede_nombre,
		centro_costo,
		periodo_desde,
		periodo_hasta,
		dias_liquidados,
		horas_asistencia_total,
		registros_asistencia,
		horas_ordinarias,
		horas_recargo_nocturno,
		horas_extra_diurnas,
		horas_extra_nocturnas,
		horas_dominicales_diurnas,
		horas_dominicales_nocturnas,
		horas_extra_dominicales_diurnas,
		horas_extra_dominicales_nocturnas,
		valor_hora_ordinaria,
		base_salario_proporcional,
		valor_recargo_nocturno,
		valor_dominical_diurno,
		valor_dominical_nocturno,
		valor_extra_diurna,
		valor_extra_nocturna,
		valor_extra_dominical_diurna,
		valor_extra_dominical_nocturna,
		total_recargos_horas_extras,
		auxilio_transporte,
		bonificacion,
		comisiones_servicio_total,
		comisiones_servicio_movimientos,
		comisiones_servicio_ajustes,
		devengado_total,
		ingreso_base_cotizacion,
		deduccion_salud,
		deduccion_pension,
		deduccion_fondo_solidaridad,
		deduccion_fija,
		otras_deducciones,
		deduccion_total,
		neto_pagar,
		origen_calculo,
		resumen_json,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		payload.EmpresaID,
		payload.EmpleadoNominaID,
		payload.EmpleadoID,
		strings.TrimSpace(payload.EmpleadoCodigo),
		payload.EmpleadoNombre,
		strings.TrimSpace(payload.EmpleadoDocumento),
		strings.TrimSpace(payload.Cargo),
		strings.TrimSpace(payload.SedeCodigo),
		strings.TrimSpace(payload.SedeNombre),
		strings.TrimSpace(payload.CentroCosto),
		payload.PeriodoDesde,
		payload.PeriodoHasta,
		payload.DiasLiquidados,
		payload.HorasAsistenciaTotal,
		payload.RegistrosAsistencia,
		payload.HorasOrdinarias,
		payload.HorasRecargoNocturno,
		payload.HorasExtraDiurnas,
		payload.HorasExtraNocturnas,
		payload.HorasDominicalesDiurnas,
		payload.HorasDominicalesNocturnas,
		payload.HorasExtraDominicalesDiurnas,
		payload.HorasExtraDominicalesNocturnas,
		payload.ValorHoraOrdinaria,
		payload.BaseSalarioProporcional,
		payload.ValorRecargoNocturno,
		payload.ValorDominicalDiurno,
		payload.ValorDominicalNocturno,
		payload.ValorExtraDiurna,
		payload.ValorExtraNocturna,
		payload.ValorExtraDominicalDiurna,
		payload.ValorExtraDominicalNocturna,
		payload.TotalRecargosHorasExtras,
		payload.AuxilioTransporte,
		payload.Bonificacion,
		payload.ComisionesServicioTotal,
		payload.ComisionesServicioMovimientos,
		payload.ComisionesServicioAjustes,
		payload.DevengadoTotal,
		payload.IngresoBaseCotizacion,
		payload.DeduccionSalud,
		payload.DeduccionPension,
		payload.DeduccionFondoSolidaridad,
		payload.DeduccionFija,
		payload.OtrasDeducciones,
		payload.DeduccionTotal,
		payload.NetoPagar,
		strings.TrimSpace(payload.OrigenCalculo),
		strings.TrimSpace(payload.ResumenJSON),
		strings.TrimSpace(payload.UsuarioCreador),
		normalizeNominaEstado(payload.Estado),
		strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ListEmpresaNominaLiquidaciones lista liquidaciones por empresa y filtros.
func ListEmpresaNominaLiquidaciones(dbConn *sql.DB, empresaID int64, filter EmpresaNominaLiquidacionFilter) ([]EmpresaNominaLiquidacion, error) {
	query := `SELECT
		id,
		empresa_id,
		COALESCE(empleado_nomina_id, 0),
		COALESCE(empleado_id, 0),
		COALESCE(empleado_codigo, ''),
		COALESCE(empleado_nombre, ''),
		COALESCE(empleado_documento, ''),
		COALESCE(cargo, ''),
		COALESCE(sede_codigo, ''),
		COALESCE(sede_nombre, ''),
		COALESCE(centro_costo, ''),
		COALESCE(periodo_desde, ''),
		COALESCE(periodo_hasta, ''),
		COALESCE(dias_liquidados, 0),
		COALESCE(horas_asistencia_total, 0),
		COALESCE(registros_asistencia, 0),
		COALESCE(horas_ordinarias, 0),
		COALESCE(horas_recargo_nocturno, 0),
		COALESCE(horas_extra_diurnas, 0),
		COALESCE(horas_extra_nocturnas, 0),
		COALESCE(horas_dominicales_diurnas, 0),
		COALESCE(horas_dominicales_nocturnas, 0),
		COALESCE(horas_extra_dominicales_diurnas, 0),
		COALESCE(horas_extra_dominicales_nocturnas, 0),
		COALESCE(valor_hora_ordinaria, 0),
		COALESCE(base_salario_proporcional, 0),
		COALESCE(valor_recargo_nocturno, 0),
		COALESCE(valor_dominical_diurno, 0),
		COALESCE(valor_dominical_nocturno, 0),
		COALESCE(valor_extra_diurna, 0),
		COALESCE(valor_extra_nocturna, 0),
		COALESCE(valor_extra_dominical_diurna, 0),
		COALESCE(valor_extra_dominical_nocturna, 0),
		COALESCE(total_recargos_horas_extras, 0),
		COALESCE(auxilio_transporte, 0),
		COALESCE(bonificacion, 0),
		COALESCE(comisiones_servicio_total, 0),
		COALESCE(comisiones_servicio_movimientos, 0),
		COALESCE(comisiones_servicio_ajustes, 0),
		COALESCE(devengado_total, 0),
		COALESCE(ingreso_base_cotizacion, 0),
		COALESCE(deduccion_salud, 0),
		COALESCE(deduccion_pension, 0),
		COALESCE(deduccion_fondo_solidaridad, 0),
		COALESCE(deduccion_fija, 0),
		COALESCE(otras_deducciones, 0),
		COALESCE(deduccion_total, 0),
		COALESCE(neto_pagar, 0),
		COALESCE(origen_calculo, 'asistencia'),
		COALESCE(resumen_json, ''),
		COALESCE(fecha_generacion, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_nomina_liquidaciones
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	if !filter.IncludeInactive {
		query += ` AND estado = 'activo'`
	}

	if strings.TrimSpace(filter.PeriodoDesde) != "" {
		query += ` AND periodo_desde >= ?`
		args = append(args, strings.TrimSpace(filter.PeriodoDesde))
	}
	if strings.TrimSpace(filter.PeriodoHasta) != "" {
		query += ` AND periodo_hasta <= ?`
		args = append(args, strings.TrimSpace(filter.PeriodoHasta))
	}
	if filter.EmpleadoNominaID > 0 {
		query += ` AND empleado_nomina_id = ?`
		args = append(args, filter.EmpleadoNominaID)
	}

	if filter.Limit <= 0 {
		filter.Limit = 300
	}
	if filter.Limit > 3000 {
		filter.Limit = 3000
	}
	query += ` ORDER BY periodo_desde DESC, empleado_nombre ASC, id DESC LIMIT ?`
	args = append(args, filter.Limit)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaNominaLiquidacion, 0)
	for rows.Next() {
		var item EmpresaNominaLiquidacion
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.EmpleadoNominaID,
			&item.EmpleadoID,
			&item.EmpleadoCodigo,
			&item.EmpleadoNombre,
			&item.EmpleadoDocumento,
			&item.Cargo,
			&item.SedeCodigo,
			&item.SedeNombre,
			&item.CentroCosto,
			&item.PeriodoDesde,
			&item.PeriodoHasta,
			&item.DiasLiquidados,
			&item.HorasAsistenciaTotal,
			&item.RegistrosAsistencia,
			&item.HorasOrdinarias,
			&item.HorasRecargoNocturno,
			&item.HorasExtraDiurnas,
			&item.HorasExtraNocturnas,
			&item.HorasDominicalesDiurnas,
			&item.HorasDominicalesNocturnas,
			&item.HorasExtraDominicalesDiurnas,
			&item.HorasExtraDominicalesNocturnas,
			&item.ValorHoraOrdinaria,
			&item.BaseSalarioProporcional,
			&item.ValorRecargoNocturno,
			&item.ValorDominicalDiurno,
			&item.ValorDominicalNocturno,
			&item.ValorExtraDiurna,
			&item.ValorExtraNocturna,
			&item.ValorExtraDominicalDiurna,
			&item.ValorExtraDominicalNocturna,
			&item.TotalRecargosHorasExtras,
			&item.AuxilioTransporte,
			&item.Bonificacion,
			&item.ComisionesServicioTotal,
			&item.ComisionesServicioMovimientos,
			&item.ComisionesServicioAjustes,
			&item.DevengadoTotal,
			&item.IngresoBaseCotizacion,
			&item.DeduccionSalud,
			&item.DeduccionPension,
			&item.DeduccionFondoSolidaridad,
			&item.DeduccionFija,
			&item.OtrasDeducciones,
			&item.DeduccionTotal,
			&item.NetoPagar,
			&item.OrigenCalculo,
			&item.ResumenJSON,
			&item.FechaGeneracion,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		item.Estado = normalizeNominaEstado(item.Estado)
		out = append(out, item)
	}
	return out, rows.Err()
}

func CreateEmpresaNominaPago(dbConn *sql.DB, payload EmpresaNominaPago) (int64, error) {
	if payload.EmpresaID <= 0 || payload.LiquidacionID <= 0 {
		return 0, fmt.Errorf("empresa_id y liquidacion_id son obligatorios")
	}
	if strings.TrimSpace(payload.FechaPago) == "" {
		payload.FechaPago = time.Now().Format("2006-01-02 15:04:05")
	}
	payload.MetodoPago = strings.ToLower(strings.TrimSpace(payload.MetodoPago))
	if payload.MetodoPago == "" {
		payload.MetodoPago = "transferencia_bancaria"
	}
	payload.EstadoPago = strings.ToLower(strings.TrimSpace(payload.EstadoPago))
	if payload.EstadoPago == "" {
		payload.EstadoPago = "pagado"
	}
	payload.Estado = normalizeNominaEstado(payload.Estado)
	if payload.UsuarioCreador == "" {
		payload.UsuarioCreador = "sistema"
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_nomina_pagos (
		empresa_id, liquidacion_id, empleado_nomina_id, empleado_nombre, empleado_documento,
		periodo_desde, periodo_hasta, fecha_pago, metodo_pago, cuenta_bancaria, referencia_pago,
		devengado_total, deduccion_total, neto_pagado, estado_pago,
		usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))`,
		payload.EmpresaID,
		payload.LiquidacionID,
		payload.EmpleadoNominaID,
		strings.TrimSpace(payload.EmpleadoNombre),
		strings.TrimSpace(payload.EmpleadoDocumento),
		strings.TrimSpace(payload.PeriodoDesde),
		strings.TrimSpace(payload.PeriodoHasta),
		strings.TrimSpace(payload.FechaPago),
		payload.MetodoPago,
		strings.TrimSpace(payload.CuentaBancaria),
		strings.TrimSpace(payload.ReferenciaPago),
		payload.DevengadoTotal,
		payload.DeduccionTotal,
		payload.NetoPagado,
		payload.EstadoPago,
		strings.TrimSpace(payload.UsuarioCreador),
		payload.Estado,
		strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func ListEmpresaNominaPagos(dbConn *sql.DB, empresaID int64, filter EmpresaNominaPagoFilter) ([]EmpresaNominaPago, error) {
	query := `SELECT
		id, empresa_id, COALESCE(liquidacion_id,0), COALESCE(empleado_nomina_id,0),
		COALESCE(empleado_nombre,''), COALESCE(empleado_documento,''),
		COALESCE(periodo_desde,''), COALESCE(periodo_hasta,''), COALESCE(fecha_pago,''),
		COALESCE(metodo_pago,''), COALESCE(cuenta_bancaria,''), COALESCE(referencia_pago,''),
		COALESCE(devengado_total,0), COALESCE(deduccion_total,0), COALESCE(neto_pagado,0),
		COALESCE(estado_pago,'pagado'), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''),
		COALESCE(usuario_creador,''), COALESCE(estado,'activo'), COALESCE(observaciones,'')
	FROM empresa_nomina_pagos WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if !filter.IncludeInactive {
		query += ` AND estado = 'activo'`
	}
	if strings.TrimSpace(filter.PeriodoDesde) != "" {
		query += ` AND periodo_desde >= ?`
		args = append(args, strings.TrimSpace(filter.PeriodoDesde))
	}
	if strings.TrimSpace(filter.PeriodoHasta) != "" {
		query += ` AND periodo_hasta <= ?`
		args = append(args, strings.TrimSpace(filter.PeriodoHasta))
	}
	if filter.EmpleadoNominaID > 0 {
		query += ` AND empleado_nomina_id = ?`
		args = append(args, filter.EmpleadoNominaID)
	}
	if filter.Limit <= 0 {
		filter.Limit = 300
	}
	if filter.Limit > 3000 {
		filter.Limit = 3000
	}
	query += ` ORDER BY fecha_pago DESC, id DESC LIMIT ?`
	args = append(args, filter.Limit)
	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmpresaNominaPago, 0)
	for rows.Next() {
		var p EmpresaNominaPago
		if err := rows.Scan(
			&p.ID, &p.EmpresaID, &p.LiquidacionID, &p.EmpleadoNominaID,
			&p.EmpleadoNombre, &p.EmpleadoDocumento, &p.PeriodoDesde, &p.PeriodoHasta,
			&p.FechaPago, &p.MetodoPago, &p.CuentaBancaria, &p.ReferenciaPago,
			&p.DevengadoTotal, &p.DeduccionTotal, &p.NetoPagado, &p.EstadoPago,
			&p.FechaCreacion, &p.FechaActualizacion, &p.UsuarioCreador, &p.Estado, &p.Observaciones,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func GenerateEmpresaNominaPagos(dbConn *sql.DB, empresaID int64, periodoDesde, periodoHasta string, empleadoNominaID int64, metodoPago, cuentaBancaria, usuario string) (*EmpresaNominaPagosResult, error) {
	liqs, err := ListEmpresaNominaLiquidaciones(dbConn, empresaID, EmpresaNominaLiquidacionFilter{
		PeriodoDesde:     periodoDesde,
		PeriodoHasta:     periodoHasta,
		EmpleadoNominaID: empleadoNominaID,
		IncludeInactive:  false,
		Limit:            3000,
	})
	if err != nil {
		return nil, err
	}
	res := &EmpresaNominaPagosResult{EmpresaID: empresaID, PeriodoDesde: periodoDesde, PeriodoHasta: periodoHasta, Pagos: make([]EmpresaNominaPago, 0)}
	for _, liq := range liqs {
		var exists int64
		err := dbConn.QueryRow(`SELECT id FROM empresa_nomina_pagos WHERE empresa_id = ? AND liquidacion_id = ? AND estado = 'activo' LIMIT 1`, empresaID, liq.ID).Scan(&exists)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		if exists > 0 {
			res.Omitidos++
			continue
		}
		ref := fmt.Sprintf("NOM-%d-%s-%s", liq.ID, strings.ReplaceAll(liq.PeriodoDesde, "-", ""), strings.ReplaceAll(liq.PeriodoHasta, "-", ""))
		pago := EmpresaNominaPago{
			EmpresaID:         empresaID,
			LiquidacionID:     liq.ID,
			EmpleadoNominaID:  liq.EmpleadoNominaID,
			EmpleadoNombre:    liq.EmpleadoNombre,
			EmpleadoDocumento: liq.EmpleadoDocumento,
			PeriodoDesde:      liq.PeriodoDesde,
			PeriodoHasta:      liq.PeriodoHasta,
			MetodoPago:        metodoPago,
			CuentaBancaria:    cuentaBancaria,
			ReferenciaPago:    ref,
			DevengadoTotal:    liq.DevengadoTotal,
			DeduccionTotal:    liq.DeduccionTotal,
			NetoPagado:        liq.NetoPagar,
			EstadoPago:        "pagado",
			UsuarioCreador:    usuario,
			Estado:            "activo",
			Observaciones:     "pago de nomina generado desde liquidacion",
		}
		id, err := CreateEmpresaNominaPago(dbConn, pago)
		if err != nil {
			return nil, err
		}
		pago.ID = id
		res.Generados++
		res.TotalNetoPagado += pago.NetoPagado
		res.Pagos = append(res.Pagos, pago)
	}
	return res, nil
}

func GetEmpresaNominaProvisionesResumen(dbConn *sql.DB, empresaID int64, periodoDesde, periodoHasta string, empleadoNominaID int64) (*EmpresaNominaProvisionesResumen, error) {
	cfg, err := GetEmpresaNominaConfiguracion(dbConn, empresaID)
	if err != nil {
		return nil, err
	}
	liqs, err := ListEmpresaNominaLiquidaciones(dbConn, empresaID, EmpresaNominaLiquidacionFilter{
		PeriodoDesde:     periodoDesde,
		PeriodoHasta:     periodoHasta,
		EmpleadoNominaID: empleadoNominaID,
		IncludeInactive:  false,
		Limit:            3000,
	})
	if err != nil {
		return nil, err
	}
	res := &EmpresaNominaProvisionesResumen{EmpresaID: empresaID, PeriodoDesde: periodoDesde, PeriodoHasta: periodoHasta, Liquidaciones: len(liqs)}
	for _, liq := range liqs {
		res.TotalDevengado += liq.DevengadoTotal
		res.TotalIBC += liq.IngresoBaseCotizacion
		res.TotalNetoPagar += liq.NetoPagar
		base := liq.IngresoBaseCotizacion
		res.AporteSaludEmpleador += base * cfg.AporteSaludEmpleadorPorcentaje / 100
		res.AportePensionEmpleador += base * cfg.AportePensionEmpleadorPorcentaje / 100
		res.AporteARL += base * cfg.AporteARLPorcentaje / 100
		res.AporteCajaCompensacion += base * cfg.AporteCajaCompensacionPorcentaje / 100
		res.AporteICBF += base * cfg.AporteICBFPorcentaje / 100
		res.AporteSENA += base * cfg.AporteSENAPorcentaje / 100
		res.ProvisionCesantias += liq.DevengadoTotal * cfg.ProvisionCesantiasPorcentaje / 100
		res.ProvisionInteresesCesantias += liq.DevengadoTotal * cfg.ProvisionInteresesCesantiasPorcentaje / 100
		res.ProvisionPrima += liq.DevengadoTotal * cfg.ProvisionPrimaPorcentaje / 100
		res.ProvisionVacaciones += liq.BaseSalarioProporcional * cfg.ProvisionVacacionesPorcentaje / 100
	}
	res.CostoEmpresaEstimado = res.TotalDevengado + res.AporteSaludEmpleador + res.AportePensionEmpleador + res.AporteARL + res.AporteCajaCompensacion + res.AporteICBF + res.AporteSENA + res.ProvisionCesantias + res.ProvisionInteresesCesantias + res.ProvisionPrima + res.ProvisionVacaciones
	return res, nil
}

// GenerateEmpresaNominaLiquidaciones calcula y guarda liquidaciones de nomina integradas con asistencia.
func GenerateEmpresaNominaLiquidaciones(dbConn *sql.DB, req EmpresaNominaCalculoRequest) (*EmpresaNominaCalculoResult, error) {
	if req.EmpresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	desde, err := normalizeNominaDate(req.PeriodoDesde)
	if err != nil {
		return nil, fmt.Errorf("periodo_desde invalido (use YYYY-MM-DD)")
	}
	hasta, err := normalizeNominaDate(req.PeriodoHasta)
	if err != nil {
		return nil, fmt.Errorf("periodo_hasta invalido (use YYYY-MM-DD)")
	}
	if hasta < desde {
		return nil, fmt.Errorf("periodo_hasta no puede ser menor a periodo_desde")
	}
	req.PeriodoDesde = desde
	req.PeriodoHasta = hasta

	cfg, err := GetEmpresaNominaConfiguracion(dbConn, req.EmpresaID)
	if err != nil {
		return nil, err
	}
	messages := make([]string, 0)
	comisionesDisponibles := true
	if err := EnsureEmpresaComisionesServicioSchema(dbConn); err != nil {
		comisionesDisponibles = false
		messages = append(messages, fmt.Sprintf("No se pudo habilitar esquema de comisiones para liquidacion: %v", err))
	}

	empleados, err := ListEmpresaNominaEmpleados(dbConn, req.EmpresaID, false, "", 5000)
	if err != nil {
		return nil, err
	}
	if req.EmpleadoNominaID > 0 {
		filtered := make([]EmpresaNominaEmpleado, 0, 1)
		for _, row := range empleados {
			if row.ID == req.EmpleadoNominaID {
				filtered = append(filtered, row)
				break
			}
		}
		empleados = filtered
	}
	if len(empleados) == 0 {
		return &EmpresaNominaCalculoResult{
			EmpresaID:    req.EmpresaID,
			PeriodoDesde: req.PeriodoDesde,
			PeriodoHasta: req.PeriodoHasta,
			Mensajes:     []string{"No hay empleados de nomina activos para calcular."},
		}, nil
	}

	festivos, err := buildEmpresaNominaFestivoSet(dbConn, req.EmpresaID, req.PeriodoDesde, req.PeriodoHasta)
	if err != nil {
		return nil, err
	}

	liquidaciones := make([]EmpresaNominaLiquidacion, 0, len(empleados))
	usuarioVinculoComision := strings.TrimSpace(req.UsuarioCreador)
	if usuarioVinculoComision == "" {
		usuarioVinculoComision = "sistema"
	}
	for _, empleado := range empleados {
		existingLiquidacionID := int64(0)
		if req.Overwrite {
			errExisting := dbConn.QueryRow(`SELECT id
				FROM empresa_nomina_liquidaciones
				WHERE empresa_id = ? AND empleado_nomina_id = ? AND periodo_desde = ? AND periodo_hasta = ?
				LIMIT 1`, req.EmpresaID, empleado.ID, req.PeriodoDesde, req.PeriodoHasta).Scan(&existingLiquidacionID)
			if errExisting != nil && errExisting != sql.ErrNoRows {
				messages = append(messages, fmt.Sprintf("No se pudo consultar liquidacion previa para %s: %v", empleado.EmpleadoNombre, errExisting))
				existingLiquidacionID = 0
			}
			if comisionesDisponibles && existingLiquidacionID > 0 {
				if errClear := LimpiarVinculoEmpresaComisionesServicioLiquidacion(dbConn, req.EmpresaID, existingLiquidacionID); errClear != nil {
					messages = append(messages, fmt.Sprintf("No se pudo limpiar vinculo de comisiones para %s: %v", empleado.EmpleadoNombre, errClear))
				}
			}
		}

		rowsAsistencia, err := listAsistenciaRowsForNomina(dbConn, req.EmpresaID, empleado, req.PeriodoDesde, req.PeriodoHasta)
		if err != nil {
			messages = append(messages, fmt.Sprintf("No se pudo consultar asistencia para %s: %v", empleado.EmpleadoNombre, err))
			continue
		}

		detail := buildNominaHorasDetalle(rowsAsistencia, cfg, festivos)
		liq := buildNominaLiquidacion(empleado, cfg, req, detail)
		if err := aplicarNovedadesColombiaEnLiquidacion(dbConn, &liq, cfg); err != nil {
			messages = append(messages, fmt.Sprintf("No se pudieron aplicar novedades Colombia para %s: %v", empleado.EmpleadoNombre, err))
		}
		var resumenComision *EmpresaComisionServicioLiquidacionResumen
		if comisionesDisponibles {
			aliases := BuildEmpresaComisionServicioAliases(
				strings.TrimSpace(empleado.EmpleadoCodigo),
				strings.TrimSpace(empleado.EmpleadoDocumento),
				strings.TrimSpace(empleado.EmpleadoNombre),
			)
			if len(aliases) > 0 {
				resumenComision, err = GetEmpresaComisionServicioLiquidacionResumen(dbConn, req.EmpresaID, aliases, req.PeriodoDesde, req.PeriodoHasta)
				if err != nil {
					messages = append(messages, fmt.Sprintf("No se pudo calcular comisiones para %s: %v", empleado.EmpleadoNombre, err))
				} else if resumenComision != nil {
					incorporarComisionesServicioEnLiquidacion(&liq, cfg, resumenComision)
				}
			}
		}

		id, err := upsertEmpresaNominaLiquidacion(dbConn, liq, req.Overwrite)
		if err != nil {
			messages = append(messages, fmt.Sprintf("No se pudo guardar liquidacion para %s: %v", empleado.EmpleadoNombre, err))
			continue
		}
		if comisionesDisponibles && resumenComision != nil && len(resumenComision.MovimientoIDs) > 0 {
			if errVinculo := VincularEmpresaComisionesServicioALiquidacion(
				dbConn,
				req.EmpresaID,
				id,
				req.PeriodoDesde,
				req.PeriodoHasta,
				usuarioVinculoComision,
				resumenComision.MovimientoIDs,
			); errVinculo != nil {
				messages = append(messages, fmt.Sprintf("No se pudo vincular comisiones en liquidacion para %s: %v", empleado.EmpleadoNombre, errVinculo))
			}
		}
		liq.ID = id
		liquidaciones = append(liquidaciones, liq)
	}

	sort.Slice(liquidaciones, func(i, j int) bool {
		return strings.ToLower(liquidaciones[i].EmpleadoNombre) < strings.ToLower(liquidaciones[j].EmpleadoNombre)
	})

	result := &EmpresaNominaCalculoResult{
		EmpresaID:     req.EmpresaID,
		PeriodoDesde:  req.PeriodoDesde,
		PeriodoHasta:  req.PeriodoHasta,
		Calculados:    len(liquidaciones),
		Liquidaciones: liquidaciones,
		Mensajes:      messages,
	}
	for _, row := range liquidaciones {
		result.TotalDevengado = round2(result.TotalDevengado + row.DevengadoTotal)
		result.TotalDeduccion = round2(result.TotalDeduccion + row.DeduccionTotal)
		result.TotalNeto = round2(result.TotalNeto + row.NetoPagar)
	}
	return result, nil
}

// GetEmpresaNominaDesprendible obtiene un desprendible estandar por empleado y periodo.
func GetEmpresaNominaDesprendible(dbConn *sql.DB, empresaID, empleadoNominaID int64, periodoDesde, periodoHasta string) (*EmpresaNominaDesprendible, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if empleadoNominaID <= 0 {
		return nil, fmt.Errorf("empleado_nomina_id es obligatorio")
	}
	desde, err := normalizeNominaDate(periodoDesde)
	if err != nil {
		return nil, fmt.Errorf("periodo_desde invalido (use YYYY-MM-DD)")
	}
	hasta, err := normalizeNominaDate(periodoHasta)
	if err != nil {
		return nil, fmt.Errorf("periodo_hasta invalido (use YYYY-MM-DD)")
	}
	if hasta < desde {
		return nil, fmt.Errorf("periodo_hasta no puede ser menor a periodo_desde")
	}

	rows, err := ListEmpresaNominaLiquidaciones(dbConn, empresaID, EmpresaNominaLiquidacionFilter{
		PeriodoDesde:     desde,
		PeriodoHasta:     hasta,
		EmpleadoNominaID: empleadoNominaID,
		IncludeInactive:  true,
		Limit:            200,
	})
	if err != nil {
		return nil, err
	}

	var liq *EmpresaNominaLiquidacion
	for i := range rows {
		if rows[i].EmpleadoNominaID == empleadoNominaID && rows[i].PeriodoDesde == desde && rows[i].PeriodoHasta == hasta {
			copy := rows[i]
			liq = &copy
			break
		}
	}
	if liq == nil {
		return nil, sql.ErrNoRows
	}

	cfg, err := GetEmpresaNominaConfiguracion(dbConn, empresaID)
	if err != nil {
		return nil, err
	}

	emp, err := getEmpresaNominaEmpleadoByID(dbConn, empresaID, empleadoNominaID)
	if err != nil {
		emp = &EmpresaNominaEmpleado{}
	}

	out := &EmpresaNominaDesprendible{
		EmpresaID:                      empresaID,
		PaisCodigo:                     normalizeNominaPais(cfg.PaisCodigo),
		Moneda:                         normalizeNominaMoneda(cfg.Moneda),
		PeriodoDesde:                   liq.PeriodoDesde,
		PeriodoHasta:                   liq.PeriodoHasta,
		FechaGeneracion:                strings.TrimSpace(liq.FechaGeneracion),
		EmpleadoNominaID:               liq.EmpleadoNominaID,
		EmpleadoID:                     liq.EmpleadoID,
		EmpleadoCodigo:                 strings.TrimSpace(liq.EmpleadoCodigo),
		EmpleadoNombre:                 liq.EmpleadoNombre,
		EmpleadoDocumento:              strings.TrimSpace(liq.EmpleadoDocumento),
		Cargo:                          strings.TrimSpace(liq.Cargo),
		SedeCodigo:                     strings.TrimSpace(liq.SedeCodigo),
		SedeNombre:                     strings.TrimSpace(liq.SedeNombre),
		CentroCosto:                    strings.TrimSpace(liq.CentroCosto),
		TipoContrato:                   normalizeNominaTipoContrato(strings.TrimSpace(emp.TipoContrato)),
		FechaIngreso:                   strings.TrimSpace(emp.FechaIngreso),
		DiasLiquidados:                 round2(liq.DiasLiquidados),
		HorasAsistencia:                round2(liq.HorasAsistenciaTotal),
		RegistrosAsistencia:            liq.RegistrosAsistencia,
		HorasOrdinarias:                round2(liq.HorasOrdinarias),
		HorasRecargoNocturno:           round2(liq.HorasRecargoNocturno),
		HorasExtraDiurnas:              round2(liq.HorasExtraDiurnas),
		HorasExtraNocturnas:            round2(liq.HorasExtraNocturnas),
		HorasDominicalesDiurnas:        round2(liq.HorasDominicalesDiurnas),
		HorasDominicalesNocturnas:      round2(liq.HorasDominicalesNocturnas),
		HorasExtraDominicalesDiurnas:   round2(liq.HorasExtraDominicalesDiurnas),
		HorasExtraDominicalesNocturnas: round2(liq.HorasExtraDominicalesNocturnas),
		ValorHoraOrdinaria:             round2(liq.ValorHoraOrdinaria),
		BaseSalarioProporcional:        round2(liq.BaseSalarioProporcional),
		ValorRecargoNocturno:           round2(liq.ValorRecargoNocturno),
		ValorDominicalDiurno:           round2(liq.ValorDominicalDiurno),
		ValorDominicalNocturno:         round2(liq.ValorDominicalNocturno),
		ValorExtraDiurna:               round2(liq.ValorExtraDiurna),
		ValorExtraNocturna:             round2(liq.ValorExtraNocturna),
		ValorExtraDominicalDiurna:      round2(liq.ValorExtraDominicalDiurna),
		ValorExtraDominicalNocturna:    round2(liq.ValorExtraDominicalNocturna),
		TotalRecargosHorasExtras:       round2(liq.TotalRecargosHorasExtras),
		AuxilioTransporte:              round2(liq.AuxilioTransporte),
		Bonificacion:                   round2(liq.Bonificacion),
		ComisionesServicioTotal:        round2(liq.ComisionesServicioTotal),
		ComisionesServicioMovimientos:  liq.ComisionesServicioMovimientos,
		ComisionesServicioAjustes:      round2(liq.ComisionesServicioAjustes),
		DevengadoTotal:                 round2(liq.DevengadoTotal),
		IngresoBaseCotizacion:          round2(liq.IngresoBaseCotizacion),
		DeduccionSalud:                 round2(liq.DeduccionSalud),
		DeduccionPension:               round2(liq.DeduccionPension),
		DeduccionFondoSolidaridad:      round2(liq.DeduccionFondoSolidaridad),
		DeduccionFija:                  round2(liq.DeduccionFija),
		OtrasDeducciones:               round2(liq.OtrasDeducciones),
		DeduccionTotal:                 round2(liq.DeduccionTotal),
		NetoPagar:                      round2(liq.NetoPagar),
		ResumenJSON:                    strings.TrimSpace(liq.ResumenJSON),
	}

	if out.TipoContrato == "" {
		out.TipoContrato = "indefinido"
	}
	return out, nil
}

// ConciliarEmpresaNominaAsistencia verifica y opcionalmente corrige diferencias entre asistencia y liquidaciones.
func ConciliarEmpresaNominaAsistencia(dbConn *sql.DB, req EmpresaNominaConciliacionRequest) (*EmpresaNominaConciliacionResult, error) {
	if req.EmpresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	desde, err := normalizeNominaDate(req.PeriodoDesde)
	if err != nil {
		return nil, fmt.Errorf("periodo_desde invalido (use YYYY-MM-DD)")
	}
	hasta, err := normalizeNominaDate(req.PeriodoHasta)
	if err != nil {
		return nil, fmt.Errorf("periodo_hasta invalido (use YYYY-MM-DD)")
	}
	if hasta < desde {
		return nil, fmt.Errorf("periodo_hasta no puede ser menor a periodo_desde")
	}

	cfg, err := GetEmpresaNominaConfiguracion(dbConn, req.EmpresaID)
	if err != nil {
		return nil, err
	}
	festivos, err := buildEmpresaNominaFestivoSet(dbConn, req.EmpresaID, desde, hasta)
	if err != nil {
		return nil, err
	}

	liquidaciones, err := ListEmpresaNominaLiquidaciones(dbConn, req.EmpresaID, EmpresaNominaLiquidacionFilter{
		PeriodoDesde:     desde,
		PeriodoHasta:     hasta,
		EmpleadoNominaID: req.EmpleadoNominaID,
		IncludeInactive:  true,
		Limit:            5000,
	})
	if err != nil {
		return nil, err
	}

	result := &EmpresaNominaConciliacionResult{
		EmpresaID:      req.EmpresaID,
		PeriodoDesde:   desde,
		PeriodoHasta:   hasta,
		AutoRecalcular: req.AutoRecalcular,
		Items:          make([]EmpresaNominaConciliacionItem, 0),
		Mensajes:       make([]string, 0),
	}

	procesados := make(map[int64]bool)
	for _, liq := range liquidaciones {
		if liq.PeriodoDesde != desde || liq.PeriodoHasta != hasta {
			continue
		}
		procesados[liq.EmpleadoNominaID] = true

		item := EmpresaNominaConciliacionItem{
			EmpleadoNominaID:               liq.EmpleadoNominaID,
			EmpleadoNombre:                 liq.EmpleadoNombre,
			RegistrosAsistenciaLiquidacion: liq.RegistrosAsistencia,
			HorasAsistenciaLiquidacion:     round2(liq.HorasAsistenciaTotal),
		}

		emp, err := getEmpresaNominaEmpleadoByID(dbConn, req.EmpresaID, liq.EmpleadoNominaID)
		if err != nil {
			item.Estado = "omitido"
			item.Mensaje = "No se encontro ficha de empleado para esta liquidacion"
			result.Items = append(result.Items, item)
			result.TotalEvaluados++
			continue
		}

		rowsAsistencia, err := listAsistenciaRowsForNomina(dbConn, req.EmpresaID, *emp, desde, hasta)
		if err != nil {
			item.Estado = "omitido"
			item.Mensaje = fmt.Sprintf("No se pudo consultar asistencia: %v", err)
			result.Items = append(result.Items, item)
			result.TotalEvaluados++
			continue
		}

		detail := buildNominaHorasDetalle(rowsAsistencia, cfg, festivos)
		item.RegistrosAsistenciaCalculado = detail.RegistrosAsistencia
		item.HorasAsistenciaCalculado = round2(detail.HorasAsistenciaTotal)
		result.TotalEvaluados++

		if nominaLiquidacionAlineadaConDetalle(liq, detail) {
			item.Estado = "conciliado"
			result.TotalConciliados++
			result.Items = append(result.Items, item)
			continue
		}

		result.TotalInconsistencias++
		item.Estado = "inconsistente"
		item.Mensaje = "Asistencia y liquidacion difieren en horas o registros"

		if req.AutoRecalcular {
			calcReq := EmpresaNominaCalculoRequest{
				EmpresaID:        req.EmpresaID,
				PeriodoDesde:     desde,
				PeriodoHasta:     hasta,
				EmpleadoNominaID: emp.ID,
				Overwrite:        true,
				OtrasDeducciones: liq.OtrasDeducciones,
				UsuarioCreador:   strings.TrimSpace(req.UsuarioCreador),
				Observaciones:    strings.TrimSpace(req.Observaciones),
			}
			if calcReq.Observaciones == "" {
				calcReq.Observaciones = "conciliacion_asistencia_nomina"
			}
			newLiq := buildNominaLiquidacion(*emp, cfg, calcReq, detail)
			newLiq.Estado = liq.Estado
			if _, err := upsertEmpresaNominaLiquidacion(dbConn, newLiq, true); err != nil {
				item.Mensaje = "Inconsistente y sin recalculo: " + err.Error()
			} else {
				item.Estado = "recalculado"
				item.Mensaje = "Liquidacion ajustada segun asistencia del periodo"
				result.TotalRecalculados++
			}
		}

		result.Items = append(result.Items, item)
	}

	if req.EmpleadoNominaID <= 0 {
		empleados, err := ListEmpresaNominaEmpleados(dbConn, req.EmpresaID, false, "", 5000)
		if err == nil {
			for _, emp := range empleados {
				if procesados[emp.ID] {
					continue
				}
				rowsAsistencia, err := listAsistenciaRowsForNomina(dbConn, req.EmpresaID, emp, desde, hasta)
				if err != nil {
					continue
				}
				detail := buildNominaHorasDetalle(rowsAsistencia, cfg, festivos)
				if detail.RegistrosAsistencia == 0 && detail.HorasAsistenciaTotal <= 0 {
					continue
				}

				result.TotalEvaluados++
				result.TotalInconsistencias++
				item := EmpresaNominaConciliacionItem{
					EmpleadoNominaID:               emp.ID,
					EmpleadoNombre:                 emp.EmpleadoNombre,
					Estado:                         "inconsistente",
					RegistrosAsistenciaLiquidacion: 0,
					RegistrosAsistenciaCalculado:   detail.RegistrosAsistencia,
					HorasAsistenciaLiquidacion:     0,
					HorasAsistenciaCalculado:       round2(detail.HorasAsistenciaTotal),
					Mensaje:                        "Existe asistencia en el periodo sin liquidacion registrada",
				}

				if req.AutoRecalcular {
					calcReq := EmpresaNominaCalculoRequest{
						EmpresaID:        req.EmpresaID,
						PeriodoDesde:     desde,
						PeriodoHasta:     hasta,
						EmpleadoNominaID: emp.ID,
						Overwrite:        true,
						UsuarioCreador:   strings.TrimSpace(req.UsuarioCreador),
						Observaciones:    strings.TrimSpace(req.Observaciones),
					}
					if calcReq.Observaciones == "" {
						calcReq.Observaciones = "conciliacion_asistencia_nomina"
					}
					newLiq := buildNominaLiquidacion(emp, cfg, calcReq, detail)
					if _, err := upsertEmpresaNominaLiquidacion(dbConn, newLiq, true); err != nil {
						item.Mensaje = "Inconsistencia detectada y sin recalculo: " + err.Error()
					} else {
						item.Estado = "recalculado"
						item.Mensaje = "Se genero liquidacion faltante segun asistencia"
						result.TotalRecalculados++
					}
				}
				result.Items = append(result.Items, item)
			}
		} else {
			result.Mensajes = append(result.Mensajes, fmt.Sprintf("No se pudo validar empleados sin liquidacion: %v", err))
		}
	}

	sort.Slice(result.Items, func(i, j int) bool {
		return strings.ToLower(strings.TrimSpace(result.Items[i].EmpleadoNombre)) < strings.ToLower(strings.TrimSpace(result.Items[j].EmpleadoNombre))
	})

	if len(result.Items) == 0 {
		result.Mensajes = append(result.Mensajes, "No se encontraron liquidaciones o asistencias para conciliar en el periodo.")
	}

	return result, nil
}
