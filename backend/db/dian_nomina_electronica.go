package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/you/pos-backend/internal/platform/valueutil"
)

const (
	empresaNominaDIANFuenteEsquema = "pcs.nomina.fuente_fiscal"
	empresaNominaDIANFuenteVersion = 2
)

// EmpresaNominaDIANPerfil contains only explicit payroll-annex data. It is
// separate from the HR master so missing fiscal data remains visible instead
// of being inferred from a display name or an internal contract label.
type EmpresaNominaDIANPerfil struct {
	ID                                 int64  `json:"id"`
	EmpresaID                          int64  `json:"empresa_id"`
	EmpleadoNominaID                   int64  `json:"empleado_nomina_id"`
	TipoDocumentoDIAN                  string `json:"tipo_documento_dian"`
	PrimerApellido                     string `json:"primer_apellido"`
	SegundoApellido                    string `json:"segundo_apellido"`
	PrimerNombre                       string `json:"primer_nombre"`
	OtrosNombres                       string `json:"otros_nombres,omitempty"`
	TipoTrabajadorDIAN                 string `json:"tipo_trabajador_dian"`
	SubTipoTrabajadorDIAN              string `json:"subtipo_trabajador_dian"`
	AltoRiesgoPension                  bool   `json:"alto_riesgo_pension"`
	SalarioIntegral                    bool   `json:"salario_integral"`
	FechaRetiro                        string `json:"fecha_retiro,omitempty"`
	LugarTrabajoPais                   string `json:"lugar_trabajo_pais"`
	LugarTrabajoDepartamentoCodigoDANE string `json:"lugar_trabajo_departamento_codigo_dane"`
	LugarTrabajoMunicipioCodigoDANE    string `json:"lugar_trabajo_municipio_codigo_dane"`
	LugarTrabajoDireccion              string `json:"lugar_trabajo_direccion"`
	TipoContratoDIAN                   string `json:"tipo_contrato_dian"`
	MetodoPagoDIAN                     string `json:"metodo_pago_dian"`
	Banco                              string `json:"banco,omitempty"`
	TipoCuenta                         string `json:"tipo_cuenta,omitempty"`
	NumeroCuenta                       string `json:"numero_cuenta,omitempty"`
	UsuarioCreador                     string `json:"usuario_creador,omitempty"`
	Estado                             string `json:"estado"`
	Observaciones                      string `json:"observaciones,omitempty"`
	FechaCreacion                      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion                 string `json:"fecha_actualizacion,omitempty"`
}

type EmpresaNominaDIANParte struct {
	RazonSocial     string `json:"razon_social,omitempty"`
	PrimerApellido  string `json:"primer_apellido,omitempty"`
	SegundoApellido string `json:"segundo_apellido,omitempty"`
	PrimerNombre    string `json:"primer_nombre,omitempty"`
	OtrosNombres    string `json:"otros_nombres,omitempty"`
	NIT             string `json:"nit"`
	DV              string `json:"dv"`
	Pais            string `json:"pais,omitempty"`
	Departamento    string `json:"departamento_codigo_dane,omitempty"`
	Municipio       string `json:"municipio_codigo_dane,omitempty"`
	Direccion       string `json:"direccion,omitempty"`
}

type EmpresaNominaDIANTrabajador struct {
	TipoTrabajador           string  `json:"tipo_trabajador"`
	SubTipoTrabajador        string  `json:"subtipo_trabajador"`
	AltoRiesgoPension        bool    `json:"alto_riesgo_pension"`
	TipoDocumento            string  `json:"tipo_documento"`
	NumeroDocumento          string  `json:"numero_documento"`
	PrimerApellido           string  `json:"primer_apellido"`
	SegundoApellido          string  `json:"segundo_apellido"`
	PrimerNombre             string  `json:"primer_nombre"`
	OtrosNombres             string  `json:"otros_nombres,omitempty"`
	LugarTrabajoPais         string  `json:"lugar_trabajo_pais"`
	LugarTrabajoDepartamento string  `json:"lugar_trabajo_departamento_codigo_dane"`
	LugarTrabajoMunicipio    string  `json:"lugar_trabajo_municipio_codigo_dane"`
	LugarTrabajoDireccion    string  `json:"lugar_trabajo_direccion"`
	SalarioIntegral          bool    `json:"salario_integral"`
	TipoContrato             string  `json:"tipo_contrato"`
	Sueldo                   float64 `json:"sueldo"`
	CodigoTrabajador         string  `json:"codigo_trabajador,omitempty"`
}

type EmpresaNominaDIANPagoFuente struct {
	FechaPago    string  `json:"fecha_pago"`
	Forma        string  `json:"forma"`
	Metodo       string  `json:"metodo"`
	Banco        string  `json:"banco,omitempty"`
	TipoCuenta   string  `json:"tipo_cuenta,omitempty"`
	NumeroCuenta string  `json:"numero_cuenta,omitempty"`
	NetoPagado   float64 `json:"neto_pagado"`
}

type EmpresaNominaDIANDevengados struct {
	DiasTrabajados            int     `json:"dias_trabajados"`
	SueldoTrabajado           float64 `json:"sueldo_trabajado"`
	AuxilioTransporte         float64 `json:"auxilio_transporte"`
	BonificacionSalarial      float64 `json:"bonificacion_salarial"`
	Comisiones                float64 `json:"comisiones"`
	Total                     float64 `json:"total"`
	TieneHorasSinTrazabilidad bool    `json:"tiene_horas_sin_trazabilidad"`
}

type EmpresaNominaDIANDeducciones struct {
	PorcentajeSalud          float64 `json:"porcentaje_salud"`
	Salud                    float64 `json:"salud"`
	PorcentajePension        float64 `json:"porcentaje_pension"`
	Pension                  float64 `json:"pension"`
	PorcentajeFondoSolidario float64 `json:"porcentaje_fondo_solidario"`
	FondoSolidario           float64 `json:"fondo_solidario"`
	DeduccionFija            float64 `json:"deduccion_fija"`
	OtrasDeducciones         float64 `json:"otras_deducciones"`
	Total                    float64 `json:"total"`
}

// EmpresaNominaDIANFuente is the immutable, secret-free source for one
// NominaIndividual. Software PIN, certificate material and tokens never enter
// this JSON; the signed XML is persisted separately before transmission.
type EmpresaNominaDIANFuente struct {
	Esquema           string                       `json:"esquema"`
	Version           int                          `json:"version"`
	EmpresaID         int64                        `json:"empresa_id"`
	NominaID          int64                        `json:"nomina_id,omitempty"`
	LiquidacionID     int64                        `json:"liquidacion_id"`
	LiquidacionIDs    []int64                      `json:"liquidacion_ids"`
	EmpleadoID        int64                        `json:"empleado_id,omitempty"`
	EmpleadoNominaID  int64                        `json:"empleado_nomina_id"`
	PagoID            int64                        `json:"pago_id"`
	PagoIDs           []int64                      `json:"pago_ids"`
	FechasPago        []string                     `json:"fechas_pago"`
	PeriodoReporte    string                       `json:"periodo_reporte"`
	PeriodoDesde      string                       `json:"periodo_desde"`
	PeriodoHasta      string                       `json:"periodo_hasta"`
	FechaIngreso      string                       `json:"fecha_ingreso"`
	FechaRetiro       string                       `json:"fecha_retiro,omitempty"`
	TiempoLaborado    int                          `json:"tiempo_laborado"`
	PeriodoNomina     int                          `json:"periodo_nomina"`
	Prefijo           string                       `json:"prefijo,omitempty"`
	Consecutivo       int64                        `json:"consecutivo,omitempty"`
	NumeroLegal       string                       `json:"numero_legal,omitempty"`
	FechaEmisionLegal string                       `json:"fecha_emision_legal,omitempty"`
	TipoAmbiente      string                       `json:"tipo_ambiente,omitempty"`
	SoftwareID        string                       `json:"software_id"`
	Empleador         EmpresaNominaDIANParte       `json:"empleador"`
	ProveedorXML      EmpresaNominaDIANParte       `json:"proveedor_xml"`
	Trabajador        EmpresaNominaDIANTrabajador  `json:"trabajador"`
	Pago              EmpresaNominaDIANPagoFuente  `json:"pago"`
	Devengados        EmpresaNominaDIANDevengados  `json:"devengados"`
	Deducciones       EmpresaNominaDIANDeducciones `json:"deducciones"`
	ComprobanteTotal  float64                      `json:"comprobante_total"`
}

type EmpresaNominaDIANConfiguracionSnapshot struct {
	TipoDocumento       string `json:"tipo_documento"`
	TipoAmbiente        string `json:"tipo_ambiente"`
	ModoOperacionCodigo string `json:"modo_operacion_codigo,omitempty"`
	TestSetID           string `json:"test_set_id,omitempty"`
	Prefijo             string `json:"prefijo"`
	ConsecutivoAsignado int64  `json:"consecutivo_asignado"`
	URLDIANOverride     string `json:"url_dian_override,omitempty"`
}

type EmpresaNominaDIANReserva struct {
	NominaID            int64  `json:"nomina_id"`
	EmpresaID           int64  `json:"empresa_id"`
	LiquidacionID       int64  `json:"liquidacion_id"`
	EmpleadoNominaID    int64  `json:"empleado_nomina_id"`
	PeriodoReporte      string `json:"periodo_reporte"`
	NumeroLegal         string `json:"numero_legal"`
	FechaEmisionLegal   string `json:"fecha_emision_legal"`
	CUNE                string `json:"cune,omitempty"`
	EstadoDIAN          string `json:"estado_dian"`
	FuenteFiscalSellada bool   `json:"fuente_fiscal_sellada"`
}

// DecodeEmpresaNominaDIANSnapshots decodes the immutable source and its
// document-family configuration without accepting fields from a newer or
// unrelated schema. A stored fiscal source must fail closed when its contract
// cannot be understood exactly by the running release.
func DecodeEmpresaNominaDIANSnapshots(sourceJSON, configJSON string) (*EmpresaNominaDIANFuente, *EmpresaNominaDIANConfiguracionSnapshot, error) {
	decode := func(raw, label string, target interface{}) error {
		raw = strings.TrimSpace(raw)
		if raw == "" || raw == "{}" || raw == "null" {
			return fmt.Errorf("%s de nómina electrónica no disponible", label)
		}
		decoder := json.NewDecoder(strings.NewReader(raw))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(target); err != nil {
			return fmt.Errorf("%s de nómina electrónica inválida: %w", label, err)
		}
		if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
			return fmt.Errorf("%s de nómina electrónica contiene datos adicionales", label)
		}
		return nil
	}
	var source EmpresaNominaDIANFuente
	if err := decode(sourceJSON, "fuente fiscal", &source); err != nil {
		return nil, nil, err
	}
	var snapshot EmpresaNominaDIANConfiguracionSnapshot
	if err := decode(configJSON, "instantánea DIAN", &snapshot); err != nil {
		return nil, nil, err
	}
	if blockers := ValidateEmpresaNominaDIANFuente(&source); len(blockers) > 0 {
		return nil, nil, fmt.Errorf("fuente fiscal de nómina electrónica inválida: %s", strings.Join(blockers, " | "))
	}
	if snapshot.TipoDocumento != "nomina_electronica" || snapshot.ConsecutivoAsignado <= 0 || strings.TrimSpace(snapshot.Prefijo) == "" {
		return nil, nil, errors.New("instantánea DIAN de nómina electrónica incompleta")
	}
	return &source, &snapshot, nil
}

func normalizeEmpresaNominaDIANPerfil(item *EmpresaNominaDIANPerfil) error {
	if item == nil || item.EmpresaID <= 0 || item.EmpleadoNominaID <= 0 {
		return errors.New("empresa_id y empleado_nomina_id son obligatorios")
	}
	item.TipoDocumentoDIAN = strings.TrimSpace(item.TipoDocumentoDIAN)
	item.PrimerApellido = strings.TrimSpace(item.PrimerApellido)
	item.SegundoApellido = strings.TrimSpace(item.SegundoApellido)
	item.PrimerNombre = strings.TrimSpace(item.PrimerNombre)
	item.OtrosNombres = strings.TrimSpace(item.OtrosNombres)
	item.TipoTrabajadorDIAN = strings.TrimSpace(item.TipoTrabajadorDIAN)
	item.SubTipoTrabajadorDIAN = strings.TrimSpace(item.SubTipoTrabajadorDIAN)
	item.FechaRetiro = strings.TrimSpace(item.FechaRetiro)
	item.LugarTrabajoPais = strings.ToUpper(strings.TrimSpace(item.LugarTrabajoPais))
	item.LugarTrabajoDepartamentoCodigoDANE = strings.TrimSpace(item.LugarTrabajoDepartamentoCodigoDANE)
	item.LugarTrabajoMunicipioCodigoDANE = strings.TrimSpace(item.LugarTrabajoMunicipioCodigoDANE)
	item.LugarTrabajoDireccion = strings.TrimSpace(item.LugarTrabajoDireccion)
	item.TipoContratoDIAN = strings.TrimSpace(item.TipoContratoDIAN)
	item.MetodoPagoDIAN = strings.ToUpper(strings.TrimSpace(item.MetodoPagoDIAN))
	item.Banco = strings.TrimSpace(item.Banco)
	item.TipoCuenta = strings.TrimSpace(item.TipoCuenta)
	item.NumeroCuenta = strings.TrimSpace(item.NumeroCuenta)
	item.UsuarioCreador = strings.TrimSpace(item.UsuarioCreador)
	item.Estado = strings.ToLower(strings.TrimSpace(item.Estado))
	item.Observaciones = strings.TrimSpace(item.Observaciones)
	if item.LugarTrabajoPais == "" {
		item.LugarTrabajoPais = "CO"
	}
	if item.Estado == "" {
		item.Estado = "activo"
	}
	if item.Estado != "activo" && item.Estado != "inactivo" {
		return errors.New("estado del perfil DIAN invalido")
	}
	if item.FechaRetiro != "" {
		if _, err := time.Parse("2006-01-02", item.FechaRetiro); err != nil {
			return errors.New("fecha_retiro invalida; use YYYY-MM-DD")
		}
	}
	if item.LugarTrabajoPais == "CO" && (item.LugarTrabajoDepartamentoCodigoDANE != "" || item.LugarTrabajoMunicipioCodigoDANE != "") {
		departamento, municipio, err := normalizeFacturacionDANECodes(item.LugarTrabajoDepartamentoCodigoDANE, item.LugarTrabajoMunicipioCodigoDANE)
		if err != nil {
			return fmt.Errorf("codigos DANE del lugar de trabajo invalidos: %w", err)
		}
		item.LugarTrabajoDepartamentoCodigoDANE = departamento
		item.LugarTrabajoMunicipioCodigoDANE = municipio
	}
	for field, value := range map[string]string{
		"tipo_documento_dian":     item.TipoDocumentoDIAN,
		"tipo_trabajador_dian":    item.TipoTrabajadorDIAN,
		"subtipo_trabajador_dian": item.SubTipoTrabajadorDIAN,
		"tipo_contrato_dian":      item.TipoContratoDIAN,
		"metodo_pago_dian":        item.MetodoPagoDIAN,
	} {
		if len(value) > 3 {
			return fmt.Errorf("%s invalido", field)
		}
	}
	return nil
}

func UpsertEmpresaNominaDIANPerfilContext(ctx context.Context, dbConn *sql.DB, item EmpresaNominaDIANPerfil) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("conexion empresarial no disponible")
	}
	if err := normalizeEmpresaNominaDIANPerfil(&item); err != nil {
		return 0, err
	}
	tx, err := dbConn.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var marker int
	if err := queryRowTxSQLCompatContext(ctx, tx, `SELECT 1 FROM empresa_nomina_empleados WHERE empresa_id = ? AND id = ? FOR UPDATE`, item.EmpresaID, item.EmpleadoNominaID).Scan(&marker); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("empleado de nomina no pertenece a la empresa")
		}
		return 0, err
	}
	var id int64
	err = queryRowTxSQLCompatContext(ctx, tx, `INSERT INTO empresa_nomina_dian_perfiles (
		empresa_id, empleado_nomina_id, tipo_documento_dian, primer_apellido, segundo_apellido,
		primer_nombre, otros_nombres, tipo_trabajador_dian, subtipo_trabajador_dian,
		alto_riesgo_pension, salario_integral, fecha_retiro, lugar_trabajo_pais,
		lugar_trabajo_departamento_codigo_dane, lugar_trabajo_municipio_codigo_dane,
		lugar_trabajo_direccion, tipo_contrato_dian, metodo_pago_dian, banco, tipo_cuenta,
		numero_cuenta, usuario_creador, estado, observaciones, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?, '')::DATE, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT (empresa_id, empleado_nomina_id) DO UPDATE SET
		tipo_documento_dian = EXCLUDED.tipo_documento_dian,
		primer_apellido = EXCLUDED.primer_apellido,
		segundo_apellido = EXCLUDED.segundo_apellido,
		primer_nombre = EXCLUDED.primer_nombre,
		otros_nombres = EXCLUDED.otros_nombres,
		tipo_trabajador_dian = EXCLUDED.tipo_trabajador_dian,
		subtipo_trabajador_dian = EXCLUDED.subtipo_trabajador_dian,
		alto_riesgo_pension = EXCLUDED.alto_riesgo_pension,
		salario_integral = EXCLUDED.salario_integral,
		fecha_retiro = EXCLUDED.fecha_retiro,
		lugar_trabajo_pais = EXCLUDED.lugar_trabajo_pais,
		lugar_trabajo_departamento_codigo_dane = EXCLUDED.lugar_trabajo_departamento_codigo_dane,
		lugar_trabajo_municipio_codigo_dane = EXCLUDED.lugar_trabajo_municipio_codigo_dane,
		lugar_trabajo_direccion = EXCLUDED.lugar_trabajo_direccion,
		tipo_contrato_dian = EXCLUDED.tipo_contrato_dian,
		metodo_pago_dian = EXCLUDED.metodo_pago_dian,
		banco = EXCLUDED.banco,
		tipo_cuenta = EXCLUDED.tipo_cuenta,
		numero_cuenta = EXCLUDED.numero_cuenta,
		usuario_creador = EXCLUDED.usuario_creador,
		estado = EXCLUDED.estado,
		observaciones = EXCLUDED.observaciones,
		fecha_actualizacion = CURRENT_TIMESTAMP
	RETURNING id`,
		item.EmpresaID, item.EmpleadoNominaID, item.TipoDocumentoDIAN, item.PrimerApellido, item.SegundoApellido,
		item.PrimerNombre, item.OtrosNombres, item.TipoTrabajadorDIAN, item.SubTipoTrabajadorDIAN,
		item.AltoRiesgoPension, item.SalarioIntegral, item.FechaRetiro, item.LugarTrabajoPais,
		item.LugarTrabajoDepartamentoCodigoDANE, item.LugarTrabajoMunicipioCodigoDANE,
		item.LugarTrabajoDireccion, item.TipoContratoDIAN, item.MetodoPagoDIAN, item.Banco,
		item.TipoCuenta, item.NumeroCuenta, item.UsuarioCreador, item.Estado, item.Observaciones,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func GetEmpresaNominaDIANPerfilContext(ctx context.Context, dbConn *sql.DB, empresaID, empleadoNominaID int64) (*EmpresaNominaDIANPerfil, error) {
	if dbConn == nil || empresaID <= 0 || empleadoNominaID <= 0 {
		return nil, errors.New("empresa_id y empleado_nomina_id son obligatorios")
	}
	var item EmpresaNominaDIANPerfil
	err := QueryRowCompatContext(ctx, dbConn, empresaNominaDIANPerfilSelect+` WHERE empresa_id = ? AND empleado_nomina_id = ? LIMIT 1`, empresaID, empleadoNominaID).Scan(empresaNominaDIANPerfilScanDest(&item)...)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

const empresaNominaDIANPerfilSelect = `SELECT id, empresa_id, empleado_nomina_id,
	COALESCE(tipo_documento_dian, ''), COALESCE(primer_apellido, ''), COALESCE(segundo_apellido, ''),
	COALESCE(primer_nombre, ''), COALESCE(otros_nombres, ''), COALESCE(tipo_trabajador_dian, ''),
	COALESCE(subtipo_trabajador_dian, ''), COALESCE(alto_riesgo_pension, FALSE),
	COALESCE(salario_integral, FALSE), COALESCE(fecha_retiro::TEXT, ''),
	COALESCE(lugar_trabajo_pais, 'CO'), COALESCE(lugar_trabajo_departamento_codigo_dane, ''),
	COALESCE(lugar_trabajo_municipio_codigo_dane, ''), COALESCE(lugar_trabajo_direccion, ''),
	COALESCE(tipo_contrato_dian, ''), COALESCE(metodo_pago_dian, ''), COALESCE(banco, ''),
	COALESCE(tipo_cuenta, ''), COALESCE(numero_cuenta, ''), COALESCE(usuario_creador, ''),
	COALESCE(estado, 'activo'), COALESCE(observaciones, ''), COALESCE(fecha_creacion::TEXT, ''),
	COALESCE(fecha_actualizacion::TEXT, '') FROM empresa_nomina_dian_perfiles`

func empresaNominaDIANPerfilScanDest(item *EmpresaNominaDIANPerfil) []interface{} {
	return []interface{}{
		&item.ID, &item.EmpresaID, &item.EmpleadoNominaID, &item.TipoDocumentoDIAN,
		&item.PrimerApellido, &item.SegundoApellido, &item.PrimerNombre, &item.OtrosNombres,
		&item.TipoTrabajadorDIAN, &item.SubTipoTrabajadorDIAN, &item.AltoRiesgoPension,
		&item.SalarioIntegral, &item.FechaRetiro, &item.LugarTrabajoPais,
		&item.LugarTrabajoDepartamentoCodigoDANE, &item.LugarTrabajoMunicipioCodigoDANE,
		&item.LugarTrabajoDireccion, &item.TipoContratoDIAN, &item.MetodoPagoDIAN,
		&item.Banco, &item.TipoCuenta, &item.NumeroCuenta, &item.UsuarioCreador, &item.Estado,
		&item.Observaciones, &item.FechaCreacion, &item.FechaActualizacion,
	}
}

// CalcularTiempoLaboradoDIAN follows section 8.3.1 of the payroll annex:
// 1 year = 360 days and 1 month = 30 days. A same-day employment relation is
// represented as one day because validation rule NIE006 requires at least 1.
func CalcularTiempoLaboradoDIAN(fechaIngreso, fechaCorte time.Time) (int, error) {
	if fechaIngreso.IsZero() || fechaCorte.IsZero() || fechaCorte.Before(fechaIngreso) {
		return 0, errors.New("fechas invalidas para calcular tiempo laborado DIAN")
	}
	days := (fechaCorte.Year()-fechaIngreso.Year())*360 + (int(fechaCorte.Month())-int(fechaIngreso.Month()))*30 + (fechaCorte.Day() - fechaIngreso.Day())
	if days < 1 {
		days = 1
	}
	return days, nil
}

type empresaNominaDIANQueryRow func(query string, args ...interface{}) *sql.Row
type empresaNominaDIANQueryRows func(query string, args ...interface{}) (*sql.Rows, error)

type empresaNominaDIANLiquidacionFuenteRow struct {
	ID                                                              int64
	EmpleadoNominaID, EmpleadoID                                    int64
	EmpleadoCodigo, EmpleadoNombre, EmpleadoDocumento               string
	PeriodoDesde, PeriodoHasta                                      string
	DiasLiquidados                                                  float64
	HorasRecargoNocturno, HorasExtraDiurnas, HorasExtraNocturnas    float64
	HorasDominicalesDiurnas, HorasDominicalesNocturnas              float64
	HorasExtraDominicalesDiurnas, HorasExtraDominicalesNocturnas    float64
	BaseSalario, Auxilio, Bonificacion, Comisiones, DevengadoTotal  float64
	DeduccionSalud, DeduccionPension, DeduccionFondo, DeduccionFija float64
	OtrasDeducciones, DeduccionTotal, NetoPagar                     float64
}

func empresaNominaDIANLiquidacionScanDest(item *empresaNominaDIANLiquidacionFuenteRow) []interface{} {
	return []interface{}{
		&item.ID, &item.EmpleadoNominaID, &item.EmpleadoID, &item.EmpleadoCodigo,
		&item.EmpleadoNombre, &item.EmpleadoDocumento, &item.PeriodoDesde, &item.PeriodoHasta,
		&item.DiasLiquidados, &item.HorasRecargoNocturno, &item.HorasExtraDiurnas,
		&item.HorasExtraNocturnas, &item.HorasDominicalesDiurnas, &item.HorasDominicalesNocturnas,
		&item.HorasExtraDominicalesDiurnas, &item.HorasExtraDominicalesNocturnas,
		&item.BaseSalario, &item.Auxilio, &item.Bonificacion, &item.Comisiones,
		&item.DevengadoTotal, &item.DeduccionSalud, &item.DeduccionPension,
		&item.DeduccionFondo, &item.DeduccionFija, &item.OtrasDeducciones,
		&item.DeduccionTotal, &item.NetoPagar,
	}
}

type empresaNominaDIANPagoFuenteRow struct {
	ID, LiquidacionID, EmpleadoNominaID           int64
	EmpleadoNombre, EmpleadoDocumento             string
	PeriodoDesde, PeriodoHasta, FechaPago         string
	MetodoPagoInterno, CuentaBancaria, EstadoPago string
	DevengadoTotal, DeduccionTotal, NetoPagado    float64
}

type empresaNominaDIANLiquidacionTotals struct {
	DiasLiquidados                                                                    float64
	HorasRecargoNocturno, HorasExtraDiurnas, HorasExtraNocturnas                      float64
	HorasDominicalesDiurnas, HorasDominicalesNocturnas                                float64
	HorasExtraDominicalesDiurnas, HorasExtraDominicalesNocturnas                      float64
	BaseSalario, Auxilio, Bonificacion, Comisiones, DevengadoTotal                    float64
	DeduccionSalud, DeduccionPension, DeduccionFondo, DeduccionFija, OtrasDeducciones float64
	DeduccionTotal, NetoPagar                                                         float64
}

type empresaNominaDIANMonthSource struct {
	SelectedEmpleadoNominaID             int64
	PeriodoReporte, MonthStart, MonthEnd string
	Liquidaciones                        []empresaNominaDIANLiquidacionFuenteRow
	LiquidacionesPorID                   map[int64]empresaNominaDIANLiquidacionFuenteRow
	LiquidacionIDs                       []int64
	CanonicalLiquidacionID, EmpleadoID   int64
	PeriodoDesde, PeriodoHasta           string
	Totals                               empresaNominaDIANLiquidacionTotals
	Blockers                             []string
}

type empresaNominaDIANEmployeeSource struct {
	Codigo, Nombre, Documento, FechaIngreso, Estado string
	SalarioMensual                                  float64
	Profile                                         EmpresaNominaDIANPerfil
	ProfileMissing                                  bool
}

type empresaNominaDIANPaymentsSource struct {
	PagosPorLiquidacion map[int64]empresaNominaDIANPagoFuenteRow
	PagoIDs             []int64
	FechasPagoISO       []string
	PagoID              int64
	FechaPago           string
	CuentaBancaria      string
	NetoPagado          float64
	Blockers            []string
}

type empresaNominaDIANFiscalSource struct {
	PeriodoNomina                                       int
	PorcentajeSalud, PorcentajePension, PorcentajeFondo float64
	Employer, Provider                                  EmpresaNominaDIANParte
}

func empresaNominaDIANPeriodoReporte(periodoDesde, periodoHasta string) (string, string, string, error) {
	desde, err := time.Parse("2006-01-02", strings.TrimSpace(periodoDesde))
	if err != nil {
		return "", "", "", errors.New("periodo_desde de la liquidación no es una fecha ISO válida")
	}
	hasta, err := time.Parse("2006-01-02", strings.TrimSpace(periodoHasta))
	if err != nil || hasta.Before(desde) {
		return "", "", "", errors.New("periodo_hasta de la liquidación no es una fecha ISO válida")
	}
	if desde.Year() != hasta.Year() || desde.Month() != hasta.Month() {
		return "", "", "", errors.New("una liquidación que cruza meses no puede convertirse automáticamente en NominaIndividual")
	}
	monthStart := time.Date(desde.Year(), desde.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, -1)
	return monthStart.Format("2006-01"), monthStart.Format("2006-01-02"), monthEnd.Format("2006-01-02"), nil
}

// EmpresaNominaDIANPeriodoCerrado prevents reserving the one monthly fiscal
// document while its reporting month can still receive more payroll payments.
// DIAN requires the employer to accumulate the month independently of its
// weekly, biweekly or other internal payment frequency.
func EmpresaNominaDIANPeriodoCerrado(periodoReporte string, emissionTime time.Time) bool {
	location := facturacionColombiaLocation()
	reportMonth, err := time.ParseInLocation("2006-01", strings.TrimSpace(periodoReporte), location)
	if err != nil {
		return false
	}
	if emissionTime.IsZero() {
		emissionTime = time.Now()
	}
	current := emissionTime.In(location)
	currentMonth := time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, location)
	return reportMonth.Before(currentMonth)
}

func LoadEmpresaNominaDIANFuenteContext(ctx context.Context, dbConn *sql.DB, empresaID, liquidacionID int64, softwareID string) (*EmpresaNominaDIANFuente, []string, error) {
	if dbConn == nil || empresaID <= 0 || liquidacionID <= 0 {
		return nil, nil, errors.New("empresa_id y liquidacion_id son obligatorios")
	}
	query := func(statement string, args ...interface{}) *sql.Row {
		return QueryRowCompatContext(ctx, dbConn, statement, args...)
	}
	queryRows := func(statement string, args ...interface{}) (*sql.Rows, error) {
		return querySQLCompatContext(ctx, dbConn, statement, args...)
	}
	return loadEmpresaNominaDIANFuente(query, queryRows, empresaID, liquidacionID, softwareID, false)
}

func loadEmpresaNominaDIANMonth(query empresaNominaDIANQueryRow, queryRows empresaNominaDIANQueryRows, empresaID, liquidacionID int64, lockClause string) (*empresaNominaDIANMonthSource, error) {
	month := &empresaNominaDIANMonthSource{}
	var selectedFrom, selectedTo string
	err := query(`SELECT empleado_nomina_id, periodo_desde, periodo_hasta
		FROM empresa_nomina_liquidaciones
		WHERE empresa_id = ? AND id = ? AND estado = 'activo'`+lockClause, empresaID, liquidacionID).Scan(
		&month.SelectedEmpleadoNominaID, &selectedFrom, &selectedTo)
	if err != nil {
		return nil, err
	}
	month.PeriodoReporte, month.MonthStart, month.MonthEnd, err = empresaNominaDIANPeriodoReporte(selectedFrom, selectedTo)
	if err != nil {
		return nil, err
	}
	rows, err := queryRows(`SELECT id, empleado_nomina_id, COALESCE(empleado_id, 0), COALESCE(empleado_codigo, ''),
		COALESCE(empleado_nombre, ''), COALESCE(empleado_documento, ''), periodo_desde, periodo_hasta,
		COALESCE(dias_liquidados, 0), COALESCE(horas_recargo_nocturno, 0), COALESCE(horas_extra_diurnas, 0),
		COALESCE(horas_extra_nocturnas, 0), COALESCE(horas_dominicales_diurnas, 0), COALESCE(horas_dominicales_nocturnas, 0),
		COALESCE(horas_extra_dominicales_diurnas, 0), COALESCE(horas_extra_dominicales_nocturnas, 0),
		COALESCE(base_salario_proporcional, 0), COALESCE(auxilio_transporte, 0), COALESCE(bonificacion, 0),
		COALESCE(comisiones_servicio_total, 0), COALESCE(devengado_total, 0), COALESCE(deduccion_salud, 0),
		COALESCE(deduccion_pension, 0), COALESCE(deduccion_fondo_solidaridad, 0), COALESCE(deduccion_fija, 0),
		COALESCE(otras_deducciones, 0), COALESCE(deduccion_total, 0), COALESCE(neto_pagar, 0)
	FROM empresa_nomina_liquidaciones
	WHERE empresa_id = ? AND empleado_nomina_id = ? AND estado = 'activo' AND periodo_desde <= ? AND periodo_hasta >= ?
	ORDER BY periodo_desde, periodo_hasta, id`+lockClause, empresaID, month.SelectedEmpleadoNominaID, month.MonthEnd, month.MonthStart)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item empresaNominaDIANLiquidacionFuenteRow
		if err := rows.Scan(empresaNominaDIANLiquidacionScanDest(&item)...); err != nil {
			return nil, err
		}
		report, _, _, periodErr := empresaNominaDIANPeriodoReporte(item.PeriodoDesde, item.PeriodoHasta)
		if periodErr != nil || report != month.PeriodoReporte {
			return nil, fmt.Errorf("la liquidación %d cruza el límite del mes %s; requiere conciliación manual", item.ID, month.PeriodoReporte)
		}
		month.Liquidaciones = append(month.Liquidaciones, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(month.Liquidaciones) == 0 {
		return nil, errors.New("no hay liquidaciones activas para el empleado y mes seleccionados")
	}
	return month, nil
}

func accumulateEmpresaNominaDIANLiquidation(totals *empresaNominaDIANLiquidacionTotals, item empresaNominaDIANLiquidacionFuenteRow) {
	totals.DiasLiquidados += item.DiasLiquidados
	totals.HorasRecargoNocturno += item.HorasRecargoNocturno
	totals.HorasExtraDiurnas += item.HorasExtraDiurnas
	totals.HorasExtraNocturnas += item.HorasExtraNocturnas
	totals.HorasDominicalesDiurnas += item.HorasDominicalesDiurnas
	totals.HorasDominicalesNocturnas += item.HorasDominicalesNocturnas
	totals.HorasExtraDominicalesDiurnas += item.HorasExtraDominicalesDiurnas
	totals.HorasExtraDominicalesNocturnas += item.HorasExtraDominicalesNocturnas
	totals.BaseSalario += item.BaseSalario
	totals.Auxilio += item.Auxilio
	totals.Bonificacion += item.Bonificacion
	totals.Comisiones += item.Comisiones
	totals.DevengadoTotal += item.DevengadoTotal
	totals.DeduccionSalud += item.DeduccionSalud
	totals.DeduccionPension += item.DeduccionPension
	totals.DeduccionFondo += item.DeduccionFondo
	totals.DeduccionFija += item.DeduccionFija
	totals.OtrasDeducciones += item.OtrasDeducciones
	totals.DeduccionTotal += item.DeduccionTotal
	totals.NetoPagar += item.NetoPagar
}

func aggregateEmpresaNominaDIANMonth(month *empresaNominaDIANMonthSource, selectedLiquidacionID int64) error {
	month.CanonicalLiquidacionID = month.Liquidaciones[0].ID
	month.PeriodoDesde, month.PeriodoHasta = month.Liquidaciones[0].PeriodoDesde, month.Liquidaciones[0].PeriodoHasta
	month.LiquidacionesPorID = make(map[int64]empresaNominaDIANLiquidacionFuenteRow, len(month.Liquidaciones))
	selectedFound := false
	var previousEnd time.Time
	for _, item := range month.Liquidaciones {
		selectedFound = selectedFound || item.ID == selectedLiquidacionID
		if item.ID < month.CanonicalLiquidacionID {
			month.CanonicalLiquidacionID = item.ID
		}
		month.LiquidacionIDs = append(month.LiquidacionIDs, item.ID)
		month.LiquidacionesPorID[item.ID] = item
		if item.EmpleadoNominaID != month.SelectedEmpleadoNominaID {
			month.Blockers = append(month.Blockers, fmt.Sprintf("La liquidación %d no pertenece al empleado mensual seleccionado.", item.ID))
		}
		if month.EmpleadoID == 0 && item.EmpleadoID > 0 {
			month.EmpleadoID = item.EmpleadoID
		} else if month.EmpleadoID > 0 && item.EmpleadoID > 0 && item.EmpleadoID != month.EmpleadoID {
			month.Blockers = append(month.Blockers, "Las liquidaciones mensuales apuntan a identificadores de empleado distintos.")
		}
		start, startErr := time.Parse("2006-01-02", strings.TrimSpace(item.PeriodoDesde))
		end, endErr := time.Parse("2006-01-02", strings.TrimSpace(item.PeriodoHasta))
		if startErr != nil || endErr != nil || end.Before(start) {
			month.Blockers = append(month.Blockers, fmt.Sprintf("La liquidación %d tiene un período inválido.", item.ID))
		} else {
			if !previousEnd.IsZero() && !start.After(previousEnd) {
				month.Blockers = append(month.Blockers, "Las liquidaciones activas del empleado se solapan dentro del mes.")
			}
			previousEnd = end
		}
		if item.PeriodoDesde < month.PeriodoDesde {
			month.PeriodoDesde = item.PeriodoDesde
		}
		if item.PeriodoHasta > month.PeriodoHasta {
			month.PeriodoHasta = item.PeriodoHasta
		}
		accumulateEmpresaNominaDIANLiquidation(&month.Totals, item)
	}
	if !selectedFound {
		return errors.New("la liquidación seleccionada dejó de pertenecer al mes fiscal consultado")
	}
	sort.Slice(month.LiquidacionIDs, func(i, j int) bool { return month.LiquidacionIDs[i] < month.LiquidacionIDs[j] })
	return nil
}

func loadEmpresaNominaDIANEmployee(query empresaNominaDIANQueryRow, empresaID, empleadoNominaID int64, lockClause string) (*empresaNominaDIANEmployeeSource, error) {
	employee := &empresaNominaDIANEmployeeSource{}
	err := query(`SELECT COALESCE(empleado_codigo, ''), COALESCE(empleado_nombre, ''),
		COALESCE(empleado_documento, ''), COALESCE(fecha_ingreso, ''), COALESCE(salario_basico_mensual, 0), COALESCE(estado, 'activo')
	FROM empresa_nomina_empleados WHERE empresa_id = ? AND id = ?`+lockClause, empresaID, empleadoNominaID).Scan(
		&employee.Codigo, &employee.Nombre, &employee.Documento, &employee.FechaIngreso, &employee.SalarioMensual, &employee.Estado)
	if err != nil {
		return nil, err
	}
	err = query(empresaNominaDIANPerfilSelect+` WHERE empresa_id = ? AND empleado_nomina_id = ? AND estado = 'activo'`+lockClause,
		empresaID, empleadoNominaID).Scan(empresaNominaDIANPerfilScanDest(&employee.Profile)...)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	employee.ProfileMissing = errors.Is(err, sql.ErrNoRows)
	return employee, nil
}

func readEmpresaNominaDIANPayments(queryRows empresaNominaDIANQueryRows, empresaID int64, month *empresaNominaDIANMonthSource, profile EmpresaNominaDIANPerfil, lock bool) (*empresaNominaDIANPaymentsSource, error) {
	lockClause := ""
	if lock {
		lockClause = " FOR UPDATE OF p"
	}
	rows, err := queryRows(`SELECT p.id, p.liquidacion_id, p.empleado_nomina_id,
		COALESCE(p.empleado_nombre, ''), COALESCE(p.empleado_documento, ''), COALESCE(p.periodo_desde, ''),
		COALESCE(p.periodo_hasta, ''), COALESCE(p.fecha_pago, ''), COALESCE(p.metodo_pago, ''),
		COALESCE(p.cuenta_bancaria, ''), COALESCE(p.estado_pago, ''), COALESCE(p.devengado_total, 0),
		COALESCE(p.deduccion_total, 0), COALESCE(p.neto_pagado, 0)
	FROM empresa_nomina_pagos p
	JOIN empresa_nomina_liquidaciones l ON l.empresa_id = p.empresa_id AND l.id = p.liquidacion_id
	WHERE p.empresa_id = ? AND p.estado = 'activo' AND l.estado = 'activo'
	  AND l.empleado_nomina_id = ? AND l.periodo_desde <= ? AND l.periodo_hasta >= ?
	ORDER BY p.fecha_pago, p.id`+lockClause, empresaID, month.SelectedEmpleadoNominaID, month.MonthEnd, month.MonthStart)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := &empresaNominaDIANPaymentsSource{PagosPorLiquidacion: make(map[int64]empresaNominaDIANPagoFuenteRow, len(month.Liquidaciones))}
	dateSet := make(map[string]struct{}, len(month.Liquidaciones))
	for rows.Next() {
		var payment empresaNominaDIANPagoFuenteRow
		if err := rows.Scan(&payment.ID, &payment.LiquidacionID, &payment.EmpleadoNominaID, &payment.EmpleadoNombre,
			&payment.EmpleadoDocumento, &payment.PeriodoDesde, &payment.PeriodoHasta, &payment.FechaPago,
			&payment.MetodoPagoInterno, &payment.CuentaBancaria, &payment.EstadoPago, &payment.DevengadoTotal,
			&payment.DeduccionTotal, &payment.NetoPagado); err != nil {
			return nil, err
		}
		if _, included := month.LiquidacionesPorID[payment.LiquidacionID]; !included {
			continue
		}
		if _, duplicate := out.PagosPorLiquidacion[payment.LiquidacionID]; duplicate {
			out.Blockers = append(out.Blockers, fmt.Sprintf("La liquidación %d tiene más de un pago activo.", payment.LiquidacionID))
			continue
		}
		out.PagosPorLiquidacion[payment.LiquidacionID] = payment
		out.PagoIDs = append(out.PagoIDs, payment.ID)
		date := valueutil.TrimmedPrefix(payment.FechaPago, 10)
		if _, err := time.Parse("2006-01-02", date); err != nil {
			out.Blockers = append(out.Blockers, fmt.Sprintf("El pago %d no tiene una fecha ISO válida.", payment.ID))
		} else {
			dateSet[date] = struct{}{}
		}
		account := strings.TrimSpace(payment.CuentaBancaria)
		if account != "" && out.CuentaBancaria == "" {
			out.CuentaBancaria = account
		} else if account != "" && out.CuentaBancaria != account && strings.TrimSpace(profile.NumeroCuenta) == "" {
			out.Blockers = append(out.Blockers, "Los pagos mensuales usan cuentas bancarias distintas y el perfil DIAN no define una cuenta canónica.")
		}
		out.NetoPagado += payment.NetoPagado
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for date := range dateSet {
		out.FechasPagoISO = append(out.FechasPagoISO, date)
	}
	return out, nil
}

func validateEmpresaNominaDIANPayments(month *empresaNominaDIANMonthSource, payments *empresaNominaDIANPaymentsSource) {
	for _, item := range month.Liquidaciones {
		payment, ok := payments.PagosPorLiquidacion[item.ID]
		if !ok {
			payments.Blockers = append(payments.Blockers, fmt.Sprintf("No existe un pago real activo para la liquidación %d.", item.ID))
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(payment.EstadoPago), "pagado") {
			payments.Blockers = append(payments.Blockers, fmt.Sprintf("El pago de la liquidación %d no está confirmado como pagado.", item.ID))
		}
		if payment.EmpleadoNominaID != month.SelectedEmpleadoNominaID || payment.PeriodoDesde != item.PeriodoDesde || payment.PeriodoHasta != item.PeriodoHasta {
			payments.Blockers = append(payments.Blockers, fmt.Sprintf("El pago %d no coincide con la fuente de su liquidación.", payment.ID))
		}
		if math.Abs(payment.DevengadoTotal-item.DevengadoTotal) > 0.01 || math.Abs(payment.DeduccionTotal-item.DeduccionTotal) > 0.01 || math.Abs(payment.NetoPagado-item.NetoPagar) > 0.01 {
			payments.Blockers = append(payments.Blockers, fmt.Sprintf("Los importes del pago %d no concilian con la liquidación %d.", payment.ID, item.ID))
		}
		if strings.TrimSpace(payment.EmpleadoNombre) != "" && !strings.EqualFold(strings.TrimSpace(payment.EmpleadoNombre), strings.TrimSpace(item.EmpleadoNombre)) {
			payments.Blockers = append(payments.Blockers, fmt.Sprintf("El nombre del pago %d no coincide con la liquidación.", payment.ID))
		}
		if strings.TrimSpace(payment.EmpleadoDocumento) != "" && strings.TrimSpace(payment.EmpleadoDocumento) != strings.TrimSpace(item.EmpleadoDocumento) {
			payments.Blockers = append(payments.Blockers, fmt.Sprintf("El documento del pago %d no coincide con la liquidación.", payment.ID))
		}
	}
	sort.Slice(payments.PagoIDs, func(i, j int) bool { return payments.PagoIDs[i] < payments.PagoIDs[j] })
	sort.Strings(payments.FechasPagoISO)
	if len(payments.PagoIDs) > 0 {
		payments.PagoID = payments.PagoIDs[0]
	}
	if len(payments.FechasPagoISO) > 0 {
		payments.FechaPago = payments.FechasPagoISO[0]
	}
}

func loadEmpresaNominaDIANFiscalSource(query empresaNominaDIANQueryRow, empresaID int64, lockClause string) (*empresaNominaDIANFiscalSource, error) {
	fiscal := &empresaNominaDIANFiscalSource{}
	err := query(`SELECT COALESCE(periodo_nomina_dian, 0), COALESCE(deduccion_salud_porcentaje, 0),
		COALESCE(deduccion_pension_porcentaje, 0), COALESCE(deduccion_fondo_solidaridad_porcentaje, 0)
	FROM empresa_nomina_configuracion WHERE empresa_id = ?`+lockClause, empresaID).Scan(
		&fiscal.PeriodoNomina, &fiscal.PorcentajeSalud, &fiscal.PorcentajePension, &fiscal.PorcentajeFondo)
	if err != nil {
		return nil, err
	}
	err = query(`SELECT COALESCE(razon_social, ''), COALESCE(nit, ''), COALESCE(digito_verificacion, ''),
		COALESCE(pais_codigo, ''), COALESCE(departamento_codigo_dane, ''), COALESCE(municipio_codigo_dane, ''),
		COALESCE(direccion_fiscal, '') FROM empresa_configuracion_avanzada WHERE empresa_id = ?`+lockClause, empresaID).Scan(
		&fiscal.Employer.RazonSocial, &fiscal.Employer.NIT, &fiscal.Employer.DV, &fiscal.Employer.Pais,
		&fiscal.Employer.Departamento, &fiscal.Employer.Municipio, &fiscal.Employer.Direccion)
	if err != nil {
		return nil, err
	}
	var useShared int
	err = query(`SELECT COALESCE(usar_software_compartido, 0), COALESCE(software_proveedor_nit, ''),
		COALESCE(software_proveedor_dv, ''), COALESCE(software_proveedor_razon_social, ''),
		COALESCE(software_proveedor_primer_apellido, ''), COALESCE(software_proveedor_segundo_apellido, ''),
		COALESCE(software_proveedor_primer_nombre, ''), COALESCE(software_proveedor_otros_nombres, '')
	FROM empresa_dian_configuracion WHERE empresa_id = ? ORDER BY id DESC LIMIT 1`+lockClause, empresaID).Scan(
		&useShared, &fiscal.Provider.NIT, &fiscal.Provider.DV, &fiscal.Provider.RazonSocial, &fiscal.Provider.PrimerApellido,
		&fiscal.Provider.SegundoApellido, &fiscal.Provider.PrimerNombre, &fiscal.Provider.OtrosNombres)
	if err != nil {
		return nil, err
	}
	if useShared == 0 {
		fiscal.Provider = EmpresaNominaDIANParte{RazonSocial: fiscal.Employer.RazonSocial, NIT: fiscal.Employer.NIT, DV: fiscal.Employer.DV}
	}
	return fiscal, nil
}

type empresaNominaDIANEmploymentSource struct {
	FechaRetiro    string
	TiempoLaborado int
	StartErr       error
	PeriodStartErr error
	EndErr         error
}

func buildEmpresaNominaDIANEmployment(month *empresaNominaDIANMonthSource, employee *empresaNominaDIANEmployeeSource) (*empresaNominaDIANEmploymentSource, []string) {
	out := &empresaNominaDIANEmploymentSource{}
	blockers := make([]string, 0, 3)
	employmentStart, startErr := time.Parse("2006-01-02", strings.TrimSpace(employee.FechaIngreso))
	liquidationStart, periodStartErr := time.Parse("2006-01-02", strings.TrimSpace(month.PeriodoDesde))
	liquidationEnd, endErr := time.Parse("2006-01-02", strings.TrimSpace(month.PeriodoHasta))
	out.StartErr, out.PeriodStartErr, out.EndErr = startErr, periodStartErr, endErr
	employmentCutoff := liquidationEnd
	if strings.TrimSpace(employee.Profile.FechaRetiro) != "" {
		retirement, err := time.Parse("2006-01-02", strings.TrimSpace(employee.Profile.FechaRetiro))
		if err != nil {
			blockers = append(blockers, "La fecha de retiro del perfil DIAN no es válida.")
		} else if startErr == nil && retirement.Before(employmentStart) {
			blockers = append(blockers, "La fecha de retiro es anterior a la fecha de ingreso del empleado.")
		} else if periodStartErr == nil && retirement.Before(liquidationStart) {
			blockers = append(blockers, "La fecha de retiro es anterior al período de nómina reportado.")
		} else if endErr == nil && !retirement.After(liquidationEnd) {
			out.FechaRetiro, employmentCutoff = retirement.Format("2006-01-02"), retirement
		}
	}
	if startErr == nil && endErr == nil {
		out.TiempoLaborado, _ = CalcularTiempoLaboradoDIAN(employmentStart, employmentCutoff)
	}
	return out, blockers
}

func buildEmpresaNominaDIANSource(empresaID int64, softwareID string, month *empresaNominaDIANMonthSource, employee *empresaNominaDIANEmployeeSource, payments *empresaNominaDIANPaymentsSource, fiscal *empresaNominaDIANFiscalSource, employment *empresaNominaDIANEmploymentSource) *EmpresaNominaDIANFuente {
	totals, profile := month.Totals, employee.Profile
	days := int(math.Round(totals.DiasLiquidados))
	untraceableHours := math.Abs(totals.HorasRecargoNocturno)+math.Abs(totals.HorasExtraDiurnas)+math.Abs(totals.HorasExtraNocturnas)+
		math.Abs(totals.HorasDominicalesDiurnas)+math.Abs(totals.HorasDominicalesNocturnas)+
		math.Abs(totals.HorasExtraDominicalesDiurnas)+math.Abs(totals.HorasExtraDominicalesNocturnas) > 0.005
	return &EmpresaNominaDIANFuente{
		Esquema: empresaNominaDIANFuenteEsquema, Version: empresaNominaDIANFuenteVersion, EmpresaID: empresaID,
		LiquidacionID: month.CanonicalLiquidacionID, LiquidacionIDs: month.LiquidacionIDs, EmpleadoID: month.EmpleadoID,
		EmpleadoNominaID: month.SelectedEmpleadoNominaID, PagoID: payments.PagoID, PagoIDs: payments.PagoIDs,
		FechasPago: payments.FechasPagoISO, PeriodoReporte: month.PeriodoReporte, PeriodoDesde: strings.TrimSpace(month.PeriodoDesde),
		PeriodoHasta: strings.TrimSpace(month.PeriodoHasta), FechaIngreso: strings.TrimSpace(employee.FechaIngreso),
		FechaRetiro: employment.FechaRetiro, TiempoLaborado: employment.TiempoLaborado, PeriodoNomina: fiscal.PeriodoNomina,
		SoftwareID: strings.TrimSpace(softwareID), Empleador: fiscal.Employer, ProveedorXML: fiscal.Provider,
		Trabajador: EmpresaNominaDIANTrabajador{
			TipoTrabajador: profile.TipoTrabajadorDIAN, SubTipoTrabajador: profile.SubTipoTrabajadorDIAN,
			AltoRiesgoPension: profile.AltoRiesgoPension, TipoDocumento: profile.TipoDocumentoDIAN,
			NumeroDocumento: strings.TrimSpace(employee.Documento), PrimerApellido: profile.PrimerApellido,
			SegundoApellido: profile.SegundoApellido, PrimerNombre: profile.PrimerNombre, OtrosNombres: profile.OtrosNombres,
			LugarTrabajoPais: profile.LugarTrabajoPais, LugarTrabajoDepartamento: profile.LugarTrabajoDepartamentoCodigoDANE,
			LugarTrabajoMunicipio: profile.LugarTrabajoMunicipioCodigoDANE, LugarTrabajoDireccion: profile.LugarTrabajoDireccion,
			SalarioIntegral: profile.SalarioIntegral, TipoContrato: profile.TipoContratoDIAN,
			Sueldo: round2(employee.SalarioMensual), CodigoTrabajador: strings.TrimSpace(employee.Codigo),
		},
		Pago: EmpresaNominaDIANPagoFuente{
			FechaPago: strings.TrimSpace(payments.FechaPago), Forma: "1", Metodo: profile.MetodoPagoDIAN, Banco: profile.Banco,
			TipoCuenta: profile.TipoCuenta, NumeroCuenta: firstNonEmptyNomina(profile.NumeroCuenta, payments.CuentaBancaria),
			NetoPagado: round2(payments.NetoPagado),
		},
		Devengados: EmpresaNominaDIANDevengados{
			DiasTrabajados: days, SueldoTrabajado: round2(totals.BaseSalario), AuxilioTransporte: round2(totals.Auxilio),
			BonificacionSalarial: round2(totals.Bonificacion), Comisiones: round2(totals.Comisiones),
			Total: round2(totals.DevengadoTotal), TieneHorasSinTrazabilidad: untraceableHours,
		},
		Deducciones: EmpresaNominaDIANDeducciones{
			PorcentajeSalud: round2(fiscal.PorcentajeSalud), Salud: round2(totals.DeduccionSalud),
			PorcentajePension: round2(fiscal.PorcentajePension), Pension: round2(totals.DeduccionPension),
			PorcentajeFondoSolidario: round2(fiscal.PorcentajeFondo), FondoSolidario: round2(totals.DeduccionFondo),
			DeduccionFija: round2(totals.DeduccionFija), OtrasDeducciones: round2(totals.OtrasDeducciones), Total: round2(totals.DeduccionTotal),
		},
		ComprobanteTotal: round2(totals.NetoPagar),
	}
}

func validateEmpresaNominaDIANLoadedSource(source *EmpresaNominaDIANFuente, month *empresaNominaDIANMonthSource, employee *empresaNominaDIANEmployeeSource, employment *empresaNominaDIANEmploymentSource) []string {
	blockers := ValidateEmpresaNominaDIANFuente(source)
	if employee.ProfileMissing {
		blockers = append(blockers, "No existe perfil fiscal DIAN activo para el empleado.")
	}
	if state := strings.ToLower(strings.TrimSpace(employee.Estado)); state != "activo" && state != "inactivo" {
		blockers = append(blockers, "El estado de la ficha del empleado no es reconocible para la fuente fiscal.")
	}
	for _, item := range month.Liquidaciones {
		if item.EmpleadoID > 0 && strings.TrimSpace(item.EmpleadoNombre) != "" && !strings.EqualFold(strings.TrimSpace(item.EmpleadoNombre), strings.TrimSpace(employee.Nombre)) {
			blockers = append(blockers, fmt.Sprintf("El nombre del empleado cambió después de la liquidación %d; requiere conciliación.", item.ID))
		}
		if strings.TrimSpace(item.EmpleadoDocumento) != "" && strings.TrimSpace(item.EmpleadoDocumento) != strings.TrimSpace(employee.Documento) {
			blockers = append(blockers, fmt.Sprintf("El documento del empleado cambió después de la liquidación %d; requiere conciliación.", item.ID))
		}
		if strings.TrimSpace(item.EmpleadoCodigo) != "" && strings.TrimSpace(item.EmpleadoCodigo) != strings.TrimSpace(employee.Codigo) {
			blockers = append(blockers, fmt.Sprintf("El código del empleado cambió después de la liquidación %d; requiere conciliación.", item.ID))
		}
	}
	if employment.StartErr != nil {
		blockers = append(blockers, "La fecha de ingreso del empleado no es válida.")
	} else if end, err := time.Parse("2006-01-02", source.PeriodoHasta); err == nil {
		if start, _ := time.Parse("2006-01-02", source.FechaIngreso); start.After(end) {
			blockers = append(blockers, "La fecha de ingreso es posterior al período de nómina reportado.")
		}
	}
	if employment.PeriodStartErr != nil {
		blockers = append(blockers, "La fecha inicial de la liquidación no es válida.")
	}
	if employment.EndErr != nil {
		blockers = append(blockers, "La fecha final de la liquidación no es válida.")
	}
	days := int(math.Round(month.Totals.DiasLiquidados))
	if math.Abs(month.Totals.DiasLiquidados-float64(days)) > 0.001 {
		blockers = append(blockers, "Los días liquidados deben ser enteros para NominaIndividual.")
	}
	if math.Abs(source.Pago.NetoPagado-source.ComprobanteTotal) > 0.01 {
		blockers = append(blockers, "El pago real no coincide con el neto de la liquidación.")
	}
	return blockers
}

func loadEmpresaNominaDIANFuente(query empresaNominaDIANQueryRow, queryRows empresaNominaDIANQueryRows, empresaID, liquidacionID int64, softwareID string, lock bool) (*EmpresaNominaDIANFuente, []string, error) {
	lockClause := ""
	if lock {
		lockClause = " FOR UPDATE"
	}
	month, err := loadEmpresaNominaDIANMonth(query, queryRows, empresaID, liquidacionID, lockClause)
	if err != nil {
		return nil, nil, err
	}
	if err := aggregateEmpresaNominaDIANMonth(month, liquidacionID); err != nil {
		return nil, nil, err
	}
	employee, err := loadEmpresaNominaDIANEmployee(query, empresaID, month.SelectedEmpleadoNominaID, lockClause)
	if err != nil {
		return nil, nil, err
	}
	payments, err := readEmpresaNominaDIANPayments(queryRows, empresaID, month, employee.Profile, lock)
	if err != nil {
		return nil, nil, err
	}
	validateEmpresaNominaDIANPayments(month, payments)
	fiscal, err := loadEmpresaNominaDIANFiscalSource(query, empresaID, lockClause)
	if err != nil {
		return nil, nil, err
	}
	employment, employmentBlockers := buildEmpresaNominaDIANEmployment(month, employee)
	source := buildEmpresaNominaDIANSource(empresaID, softwareID, month, employee, payments, fiscal, employment)
	blockers := append([]string(nil), month.Blockers...)
	blockers = append(blockers, payments.Blockers...)
	blockers = append(blockers, employmentBlockers...)
	blockers = append(blockers, validateEmpresaNominaDIANLoadedSource(source, month, employee, employment)...)
	return source, valueutil.UniqueSortedNonBlank(blockers), nil
}

// EmpresaNominaDIANFuenteOperacionalCoincide compares the current monthly
// source with a reserved snapshot without treating numbering/emission metadata
// as mutable business data. Any changed liquidation, payment or fiscal profile
// makes a reserved document fail closed.
func EmpresaNominaDIANFuenteOperacionalCoincide(stored, current *EmpresaNominaDIANFuente) bool {
	if stored == nil || current == nil {
		return false
	}
	left, right := *stored, *current
	for _, item := range []*EmpresaNominaDIANFuente{&left, &right} {
		item.NominaID = 0
		item.Prefijo = ""
		item.Consecutivo = 0
		item.NumeroLegal = ""
		item.FechaEmisionLegal = ""
		item.TipoAmbiente = ""
	}
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && string(leftJSON) == string(rightJSON)
}

func firstNonEmptyNomina(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func validateEmpresaNominaDIANSourceIDs(source *EmpresaNominaDIANFuente) []string {
	blockers := make([]string, 0, 4)
	if len(source.LiquidacionIDs) == 0 || source.LiquidacionIDs[0] != source.LiquidacionID {
		blockers = append(blockers, "La fuente mensual no identifica su liquidación canónica y todas las liquidaciones incluidas.")
	} else {
		for i, id := range source.LiquidacionIDs {
			if id <= 0 || i > 0 && id <= source.LiquidacionIDs[i-1] {
				blockers = append(blockers, "Los identificadores de liquidación mensual deben ser positivos, únicos y ordenados.")
				break
			}
		}
	}
	if len(source.PagoIDs) != len(source.LiquidacionIDs) || len(source.PagoIDs) == 0 || source.PagoIDs[0] != source.PagoID {
		blockers = append(blockers, "Cada liquidación mensual debe tener un pago canónico incluido en la fuente fiscal.")
	} else {
		for i, id := range source.PagoIDs {
			if id <= 0 || i > 0 && id <= source.PagoIDs[i-1] {
				blockers = append(blockers, "Los identificadores de pago mensual deben ser positivos, únicos y ordenados.")
				break
			}
		}
	}
	return blockers
}

func validateEmpresaNominaDIANSourcePeriod(source *EmpresaNominaDIANFuente) []string {
	blockers := make([]string, 0, 5)
	if source.PeriodoNomina < 1 || source.PeriodoNomina > 6 {
		blockers = append(blockers, "Configure la periodicidad DIAN de nómina con un código entre 1 y 6.")
	}
	periodStart, startErr := time.Parse("2006-01-02", source.PeriodoDesde)
	periodEnd, endErr := time.Parse("2006-01-02", source.PeriodoHasta)
	if startErr != nil || endErr != nil || periodEnd.Before(periodStart) {
		blockers = append(blockers, "El período de liquidación es inválido.")
	} else if periodStart.Year() != periodEnd.Year() || periodStart.Month() != periodEnd.Month() || source.PeriodoReporte != periodStart.Format("2006-01") {
		blockers = append(blockers, "La fuente debe consolidar un único mes calendario identificado por periodo_reporte.")
	}
	employmentStart, employmentErr := time.Parse("2006-01-02", source.FechaIngreso)
	employmentCutoff := periodEnd
	if employmentErr != nil || endErr != nil || employmentStart.After(periodEnd) {
		blockers = append(blockers, "La fecha de ingreso no es válida para el período de nómina.")
	}
	if strings.TrimSpace(source.FechaRetiro) != "" {
		retirement, err := time.Parse("2006-01-02", source.FechaRetiro)
		if err != nil || employmentErr == nil && retirement.Before(employmentStart) || startErr == nil && retirement.Before(periodStart) || endErr == nil && retirement.After(periodEnd) {
			blockers = append(blockers, "La fecha de retiro debe pertenecer al período reportado y no ser anterior al ingreso.")
		} else {
			employmentCutoff = retirement
		}
	}
	if expected, err := CalcularTiempoLaboradoDIAN(employmentStart, employmentCutoff); err != nil || source.TiempoLaborado != expected {
		blockers = append(blockers, "El tiempo laborado DIAN no coincide con las fechas laborales de la fuente.")
	}
	return blockers
}

func validateEmpresaNominaDIANSourceParties(source *EmpresaNominaDIANFuente) []string {
	blockers := make([]string, 0, 10)
	if !nominaDIANDigits(source.Empleador.NIT) || !nominaDIANDigitoVerificacionValido(source.Empleador.DV) || source.Empleador.RazonSocial == "" {
		blockers = append(blockers, "La identidad fiscal del empleador está incompleta.")
	}
	if source.Empleador.Pais != "CO" || source.Empleador.Direccion == "" {
		blockers = append(blockers, "La ubicación fiscal colombiana del empleador está incompleta.")
	}
	if _, _, err := normalizeFacturacionDANECodes(source.Empleador.Departamento, source.Empleador.Municipio); err != nil {
		blockers = append(blockers, "Los códigos DANE del empleador son inválidos.")
	}
	providerNatural := source.ProveedorXML.PrimerApellido != "" && source.ProveedorXML.SegundoApellido != "" && source.ProveedorXML.PrimerNombre != ""
	if !nominaDIANDigits(source.ProveedorXML.NIT) || !nominaDIANDigitoVerificacionValido(source.ProveedorXML.DV) || source.ProveedorXML.RazonSocial == "" && !providerNatural {
		blockers = append(blockers, "La identidad del proveedor del software DIAN está incompleta.")
	}
	if strings.TrimSpace(source.SoftwareID) == "" {
		blockers = append(blockers, "No está disponible el identificador del software DIAN de nómina.")
	}
	workerTypes := map[string]bool{"01": true, "02": true, "04": true, "12": true, "18": true, "19": true, "21": true, "22": true, "23": true, "30": true, "31": true, "47": true, "51": true, "54": true, "56": true, "58": true}
	if !workerTypes[source.Trabajador.TipoTrabajador] {
		blockers = append(blockers, "El tipo de trabajador DIAN no pertenece al catálogo admitido.")
	}
	if source.Trabajador.SubTipoTrabajador != "00" && source.Trabajador.SubTipoTrabajador != "01" {
		blockers = append(blockers, "El subtipo de trabajador DIAN debe ser 00 o 01.")
	}
	documentTypes := map[string]bool{"11": true, "12": true, "13": true, "21": true, "22": true, "31": true, "41": true, "42": true, "47": true, "50": true, "91": true}
	if !documentTypes[source.Trabajador.TipoDocumento] || !nominaDIANDigits(source.Trabajador.NumeroDocumento) {
		blockers = append(blockers, "La identificación del trabajador no cumple el catálogo/formato DIAN.")
	}
	if source.Trabajador.PrimerApellido == "" || source.Trabajador.SegundoApellido == "" || source.Trabajador.PrimerNombre == "" {
		blockers = append(blockers, "Los nombres y dos apellidos obligatorios del trabajador están incompletos.")
	}
	if source.Trabajador.LugarTrabajoPais != "CO" || source.Trabajador.LugarTrabajoDireccion == "" {
		blockers = append(blockers, "La ubicación colombiana del trabajador está incompleta.")
	}
	if _, _, err := normalizeFacturacionDANECodes(source.Trabajador.LugarTrabajoDepartamento, source.Trabajador.LugarTrabajoMunicipio); err != nil {
		blockers = append(blockers, "Los códigos DANE del lugar de trabajo son inválidos.")
	}
	if source.Trabajador.TipoContrato < "1" || source.Trabajador.TipoContrato > "5" || len(source.Trabajador.TipoContrato) != 1 {
		blockers = append(blockers, "El tipo de contrato DIAN debe estar entre 1 y 5.")
	}
	if source.Trabajador.Sueldo <= 0 {
		blockers = append(blockers, "El sueldo base mensual del trabajador debe ser positivo.")
	}
	return blockers
}

func validateEmpresaNominaDIANSourcePayment(source *EmpresaNominaDIANFuente) []string {
	blockers := make([]string, 0, 7)
	if source.Pago.Forma != "1" || !nominaDIANPaymentMethodValid(source.Pago.Metodo) {
		blockers = append(blockers, "La forma o método de pago no pertenece al catálogo DIAN de nómina.")
	}
	if len(source.FechasPago) == 0 || source.Pago.FechaPago != source.FechasPago[0] {
		blockers = append(blockers, "La fuente mensual debe conservar todas las fechas de pago reales y su fecha canónica.")
	} else {
		for i, fecha := range source.FechasPago {
			if _, err := time.Parse("2006-01-02", fecha); err != nil || i > 0 && fecha <= source.FechasPago[i-1] {
				blockers = append(blockers, "Las fechas de pago reales deben ser ISO, únicas y ordenadas.")
				break
			}
		}
	}
	if source.Devengados.DiasTrabajados < 1 || source.Devengados.DiasTrabajados > 30 || source.Devengados.SueldoTrabajado <= 0 {
		blockers = append(blockers, "El básico debe tener entre 1 y 30 días y sueldo trabajado positivo.")
	}
	if source.Devengados.TieneHorasSinTrazabilidad {
		blockers = append(blockers, "La liquidación contiene horas extra o recargos agregados sin intervalos de inicio/fin; no se pueden inventar para DIAN.")
	}
	devengados := round2(source.Devengados.SueldoTrabajado + source.Devengados.AuxilioTransporte + source.Devengados.BonificacionSalarial + source.Devengados.Comisiones)
	if math.Abs(devengados-source.Devengados.Total) > 0.01 || source.Devengados.Total <= 0 {
		blockers = append(blockers, "Los devengados no concilian con los conceptos estructurados disponibles.")
	}
	deducciones := round2(source.Deducciones.Salud + source.Deducciones.Pension + source.Deducciones.FondoSolidario + source.Deducciones.DeduccionFija + source.Deducciones.OtrasDeducciones)
	if math.Abs(deducciones-source.Deducciones.Total) > 0.01 || source.Deducciones.Total < 0 {
		blockers = append(blockers, "Las deducciones no concilian con los conceptos estructurados disponibles.")
	}
	if math.Abs(round2(source.Devengados.Total-source.Deducciones.Total)-source.ComprobanteTotal) > 0.01 || source.ComprobanteTotal < 0 {
		blockers = append(blockers, "El comprobante total no coincide con devengados menos deducciones.")
	}
	return blockers
}

func ValidateEmpresaNominaDIANFuente(source *EmpresaNominaDIANFuente) []string {
	if source == nil || source.Esquema != empresaNominaDIANFuenteEsquema || source.Version != empresaNominaDIANFuenteVersion || source.EmpresaID <= 0 || source.LiquidacionID <= 0 || source.EmpleadoNominaID <= 0 {
		return []string{"La fuente fiscal de nómina no pertenece al esquema o empresa esperados."}
	}
	blockers := validateEmpresaNominaDIANSourceIDs(source)
	blockers = append(blockers, validateEmpresaNominaDIANSourcePeriod(source)...)
	blockers = append(blockers, validateEmpresaNominaDIANSourceParties(source)...)
	blockers = append(blockers, validateEmpresaNominaDIANSourcePayment(source)...)
	return valueutil.UniqueSortedNonBlank(blockers)
}

func nominaDIANPaymentMethodValid(value string) bool {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "ZZZ" {
		return true
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return false
	}
	return n >= 1 && n <= 53 || n >= 60 && n <= 72 || n >= 74 && n <= 78 || n >= 91 && n <= 98
}

func nominaDIANDigits(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func nominaDIANDigitoVerificacionValido(value string) bool {
	value = strings.TrimSpace(value)
	return len(value) == 1 && value[0] >= '0' && value[0] <= '9'
}

func ValidateEmpresaNominaElectronicaConfigForEmission(config EmpresaDIANDocumentoConfiguracion) error {
	environment := strings.ToLower(strings.TrimSpace(config.TipoAmbiente))
	state := strings.ToLower(strings.TrimSpace(config.Estado))
	switch environment {
	case "habilitacion":
		if state != "habilitacion" {
			return errors.New("la configuración de nómina electrónica debe estar en estado habilitación")
		}
		if !nominaDIANTestSetIDValid(config.TestSetID) {
			return errors.New("TestSetId DIAN de nómina electrónica inválido para habilitación")
		}
	case "produccion":
		if state != "activo" {
			return errors.New("la configuración de nómina electrónica debe estar activa para producción")
		}
	default:
		return errors.New("ambiente DIAN de nómina electrónica inválido")
	}
	prefix := strings.ToUpper(strings.TrimSpace(config.Prefijo))
	if prefix == "" || len(prefix) > 20 || !documentoSoportePrefixValid(prefix) {
		return errors.New("el prefijo interno de nómina electrónica debe ser alfanumérico y tener entre 1 y 20 caracteres")
	}
	if config.ConsecutivoActual <= 0 || config.ConsecutivoActual > 999999999999 {
		return errors.New("el consecutivo interno de nómina electrónica debe estar entre 1 y 999999999999")
	}
	return nil
}

func nominaDIANTestSetIDValid(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return false
	}
	for index, char := range value {
		if index == 8 || index == 13 || index == 18 || index == 23 {
			continue
		}
		if !(char >= '0' && char <= '9') && !(char >= 'a' && char <= 'f') {
			return false
		}
	}
	return true
}

func loadEmpresaNominaDIANExistingReservation(ctx context.Context, tx *sql.Tx, empresaID int64, source *EmpresaNominaDIANFuente) (*EmpresaNominaDIANReserva, *EmpresaNominaDIANFuente, *EmpresaNominaDIANConfiguracionSnapshot, bool, error) {
	existing := &EmpresaNominaDIANReserva{}
	var sourceJSON, configJSON string
	err := queryRowTxSQLCompatContext(ctx, tx, `SELECT id, empresa_id, liquidacion_id, empleado_nomina_id,
		COALESCE(periodo_reporte, ''), COALESCE(numero_legal, ''), COALESCE(fecha_emision_legal::TEXT, ''), COALESCE(cune, ''),
		COALESCE(estado_dian, 'borrador'), COALESCE(fuente_fiscal_sellada, FALSE),
		COALESCE(fuente_fiscal_json::TEXT, '{}'), COALESCE(configuracion_dian_json::TEXT, '{}')
	FROM empresa_contabilidad_nomina_electronica
	WHERE empresa_id = ? AND empleado_nomina_id = ? AND periodo_reporte = ? FOR UPDATE`, empresaID, source.EmpleadoNominaID, source.PeriodoReporte).Scan(
		&existing.NominaID, &existing.EmpresaID, &existing.LiquidacionID, &existing.EmpleadoNominaID, &existing.PeriodoReporte,
		&existing.NumeroLegal, &existing.FechaEmisionLegal, &existing.CUNE, &existing.EstadoDIAN,
		&existing.FuenteFiscalSellada, &sourceJSON, &configJSON)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, nil, false, err
	}
	if err != nil || strings.TrimSpace(existing.NumeroLegal) == "" {
		return existing, nil, nil, false, nil
	}
	stored, snapshot, err := DecodeEmpresaNominaDIANSnapshots(sourceJSON, configJSON)
	if err != nil || stored.EmpresaID != empresaID || stored.LiquidacionID != existing.LiquidacionID ||
		stored.EmpleadoNominaID != source.EmpleadoNominaID || stored.PeriodoReporte != source.PeriodoReporte ||
		stored.NominaID > 0 && stored.NominaID != existing.NominaID || !strings.EqualFold(stored.NumeroLegal, existing.NumeroLegal) ||
		!strings.EqualFold(stored.Prefijo, snapshot.Prefijo) || stored.Consecutivo != snapshot.ConsecutivoAsignado ||
		!EmpresaNominaDIANFuenteOperacionalCoincide(stored, source) {
		return nil, nil, nil, false, errors.New("nómina reservada sin fuente/configuración DIAN válida; requiere conciliación manual")
	}
	return existing, stored, snapshot, true, nil
}

func loadEmpresaNominaDIANEmissionConfig(ctx context.Context, tx *sql.Tx, empresaID int64) (*EmpresaDIANDocumentoConfiguracion, error) {
	config := &EmpresaDIANDocumentoConfiguracion{}
	err := queryRowTxSQLCompatContext(ctx, tx, `SELECT id, empresa_id, tipo_documento, estado, tipo_ambiente,
		COALESCE(modo_operacion_codigo, ''), COALESCE(test_set_id, ''), COALESCE(prefijo, ''),
		COALESCE(resolucion_numero, ''), COALESCE(resolucion_fecha_desde::TEXT, ''), COALESCE(resolucion_fecha_hasta::TEXT, ''),
		COALESCE(rango_desde, 0), COALESCE(rango_hasta, 0), COALESCE(consecutivo_actual, 0), COALESCE(url_dian_override, ''),
		COALESCE(observaciones, ''), COALESCE(usuario_creador, ''), COALESCE(fecha_creacion::TEXT, ''), COALESCE(fecha_actualizacion::TEXT, '')
	FROM empresa_dian_documentos_configuracion
	WHERE empresa_id = ? AND tipo_documento = 'nomina_electronica' FOR UPDATE`, empresaID).Scan(
		&config.ID, &config.EmpresaID, &config.TipoDocumento, &config.Estado, &config.TipoAmbiente,
		&config.ModoOperacionCodigo, &config.TestSetID, &config.Prefijo, &config.ResolucionNumero,
		&config.ResolucionFechaDesde, &config.ResolucionFechaHasta, &config.RangoDesde, &config.RangoHasta,
		&config.ConsecutivoActual, &config.URLDIANOverride, &config.Observaciones, &config.UsuarioCreador,
		&config.FechaCreacion, &config.FechaActualizacion)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("no existe configuración DIAN separada para nómina electrónica")
	}
	if err != nil {
		return nil, err
	}
	if err := ValidateEmpresaNominaElectronicaConfigForEmission(*config); err != nil {
		return nil, err
	}
	return config, nil
}

func prepareEmpresaNominaDIANReservation(source *EmpresaNominaDIANFuente, config *EmpresaDIANDocumentoConfiguracion, emissionTime time.Time) (*EmpresaNominaDIANConfiguracionSnapshot, []byte, []byte, error) {
	assigned := config.ConsecutivoActual
	snapshot := &EmpresaNominaDIANConfiguracionSnapshot{
		TipoDocumento: "nomina_electronica", TipoAmbiente: config.TipoAmbiente, ModoOperacionCodigo: config.ModoOperacionCodigo,
		TestSetID: config.TestSetID, Prefijo: strings.ToUpper(strings.TrimSpace(config.Prefijo)),
		ConsecutivoAsignado: assigned, URLDIANOverride: strings.TrimSpace(config.URLDIANOverride),
	}
	source.Prefijo, source.Consecutivo = snapshot.Prefijo, assigned
	source.NumeroLegal = snapshot.Prefijo + strconv.FormatInt(assigned, 10)
	source.FechaEmisionLegal, source.TipoAmbiente = emissionTime.Format(time.RFC3339), snapshot.TipoAmbiente
	sourceRaw, err := json.Marshal(source)
	if err != nil {
		return nil, nil, nil, err
	}
	snapshotRaw, err := json.Marshal(snapshot)
	if err != nil {
		return nil, nil, nil, err
	}
	return snapshot, sourceRaw, snapshotRaw, nil
}

func persistEmpresaNominaDIANReservation(ctx context.Context, tx *sql.Tx, existing *EmpresaNominaDIANReserva, source *EmpresaNominaDIANFuente, snapshot *EmpresaNominaDIANConfiguracionSnapshot, sourceRaw, snapshotRaw []byte, usuario string) error {
	period := source.PeriodoDesde + "/" + source.PeriodoHasta
	workerName := strings.TrimSpace(source.Trabajador.PrimerNombre + " " + source.Trabajador.OtrosNombres + " " + source.Trabajador.PrimerApellido + " " + source.Trabajador.SegundoApellido)
	if existing.NominaID <= 0 {
		err := queryRowTxSQLCompatContext(ctx, tx, `INSERT INTO empresa_contabilidad_nomina_electronica (
			empresa_id, empleado_id, tipo_documento, documento, nombre, periodo, fecha_pago,
			salario_base, devengados, deducciones, total, cune, estado_dian, respuesta_dian,
			json_payload, fecha_actualizacion, usuario_creador, liquidacion_id, empleado_nomina_id,
			periodo_reporte, numero_legal, fecha_emision_legal, configuracion_dian_json, fuente_fiscal_json, fuente_fiscal_sellada
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '', 'preparado', '', ?, CURRENT_TIMESTAMP, ?, ?, ?, ?, ?, ?, ?::jsonb, ?::jsonb, FALSE)
		RETURNING id`, source.EmpresaID, source.EmpleadoID, source.Trabajador.TipoDocumento, source.Trabajador.NumeroDocumento,
			workerName, period, valueutil.TrimmedPrefix(source.Pago.FechaPago, 10), source.Trabajador.Sueldo, source.Devengados.Total,
			source.Deducciones.Total, source.ComprobanteTotal, string(sourceRaw), strings.TrimSpace(usuario), source.LiquidacionID,
			source.EmpleadoNominaID, source.PeriodoReporte, source.NumeroLegal, source.FechaEmisionLegal, string(snapshotRaw), string(sourceRaw)).Scan(&existing.NominaID)
		if err != nil {
			return err
		}
	} else {
		result, err := execTxSQLCompat(tx, `UPDATE empresa_contabilidad_nomina_electronica SET
			empleado_id = ?, tipo_documento = ?, documento = ?, nombre = ?, periodo = ?, fecha_pago = ?,
			salario_base = ?, devengados = ?, deducciones = ?, total = ?, estado_dian = 'preparado', json_payload = ?,
			usuario_creador = ?, liquidacion_id = ?, empleado_nomina_id = ?, periodo_reporte = ?, numero_legal = ?,
			fecha_emision_legal = ?, configuracion_dian_json = ?::jsonb, fuente_fiscal_json = ?::jsonb,
			fuente_fiscal_sellada = FALSE, fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE empresa_id = ? AND id = ? AND numero_legal = ''`, source.EmpleadoID, source.Trabajador.TipoDocumento,
			source.Trabajador.NumeroDocumento, workerName, period, valueutil.TrimmedPrefix(source.Pago.FechaPago, 10),
			source.Trabajador.Sueldo, source.Devengados.Total, source.Deducciones.Total, source.ComprobanteTotal,
			string(sourceRaw), strings.TrimSpace(usuario), source.LiquidacionID, source.EmpleadoNominaID, source.PeriodoReporte,
			source.NumeroLegal, source.FechaEmisionLegal, string(snapshotRaw), string(sourceRaw), source.EmpresaID, existing.NominaID)
		if err != nil {
			return err
		}
		if rows, _ := result.RowsAffected(); rows != 1 {
			return errors.New("la nómina cambió mientras se reservaba su número")
		}
	}
	source.NominaID = existing.NominaID
	sealedRaw, err := json.Marshal(source)
	if err != nil {
		return err
	}
	if _, err := execTxSQLCompat(tx, `UPDATE empresa_contabilidad_nomina_electronica
		SET fuente_fiscal_json = ?::jsonb, json_payload = ?, fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE empresa_id = ? AND id = ? AND numero_legal = ?`, string(sealedRaw), string(sealedRaw), source.EmpresaID, existing.NominaID, source.NumeroLegal); err != nil {
		return err
	}
	_, err = execTxSQLCompat(tx, `UPDATE empresa_dian_documentos_configuracion SET consecutivo_actual = ?, fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE empresa_id = ? AND tipo_documento = 'nomina_electronica'`, snapshot.ConsecutivoAsignado+1, source.EmpresaID)
	return err
}

func ReserveEmpresaNominaElectronicaNumeroContext(ctx context.Context, dbConn *sql.DB, empresaID, liquidacionID int64, softwareID string, emissionTime time.Time, usuario string) (*EmpresaNominaDIANReserva, *EmpresaNominaDIANFuente, *EmpresaNominaDIANConfiguracionSnapshot, error) {
	if dbConn == nil || empresaID <= 0 || liquidacionID <= 0 {
		return nil, nil, nil, errors.New("empresa_id y liquidacion_id son obligatorios")
	}
	if emissionTime.IsZero() {
		emissionTime = time.Now()
	}
	emissionTime = emissionTime.In(facturacionColombiaLocation())
	tx, err := dbConn.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, nil, err
	}
	defer tx.Rollback()

	query := func(statement string, args ...interface{}) *sql.Row {
		return queryRowTxSQLCompatContext(ctx, tx, statement, args...)
	}
	queryRows := func(statement string, args ...interface{}) (*sql.Rows, error) {
		return queryTxSQLCompatContext(ctx, tx, statement, args...)
	}
	source, blockers, err := loadEmpresaNominaDIANFuente(query, queryRows, empresaID, liquidacionID, softwareID, true)
	if err != nil {
		return nil, nil, nil, err
	}
	if !EmpresaNominaDIANPeriodoCerrado(source.PeriodoReporte, emissionTime) {
		return nil, source, nil, errors.New("el mes de nómina aún no está cerrado; DIAN exige consolidar todos los pagos mensuales antes de reservar NominaIndividual")
	}
	if len(blockers) > 0 {
		return nil, source, nil, fmt.Errorf("preflight de nómina electrónica bloqueado: %s", strings.Join(blockers, " | "))
	}

	existing, storedSource, storedSnapshot, alreadyReserved, err := loadEmpresaNominaDIANExistingReservation(ctx, tx, empresaID, source)
	if err != nil {
		return nil, nil, nil, err
	}
	if alreadyReserved {
		if err := tx.Commit(); err != nil {
			return nil, nil, nil, err
		}
		return existing, storedSource, storedSnapshot, nil
	}
	config, err := loadEmpresaNominaDIANEmissionConfig(ctx, tx, empresaID)
	if err != nil {
		return nil, source, nil, err
	}
	snapshot, sourceRaw, snapshotRaw, err := prepareEmpresaNominaDIANReservation(source, config, emissionTime)
	if err != nil {
		return nil, nil, nil, err
	}
	if err := persistEmpresaNominaDIANReservation(ctx, tx, existing, source, snapshot, sourceRaw, snapshotRaw, usuario); err != nil {
		return nil, nil, nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, nil, err
	}
	existing.EmpresaID = source.EmpresaID
	existing.LiquidacionID = source.LiquidacionID
	existing.EmpleadoNominaID = source.EmpleadoNominaID
	existing.PeriodoReporte = source.PeriodoReporte
	existing.NumeroLegal = source.NumeroLegal
	existing.FechaEmisionLegal = source.FechaEmisionLegal
	existing.EstadoDIAN = "preparado"
	return existing, source, snapshot, nil
}

func UpdateEmpresaNominaDIANResultContext(ctx context.Context, dbConn *sql.DB, empresaID, nominaID int64, estado, cune, respuesta string, fuenteSellada bool, registrarIntento bool) error {
	if dbConn == nil || empresaID <= 0 || nominaID <= 0 {
		return errors.New("empresa_id y nomina_id son obligatorios")
	}
	estado = strings.ToLower(strings.TrimSpace(estado))
	switch estado {
	case "preparado", "pendiente", "enviado", "aceptado", "rechazado", "fallido", "contingencia":
	default:
		return errors.New("estado DIAN de nómina electrónica inválido")
	}
	cune = strings.ToLower(strings.TrimSpace(cune))
	if cune != "" && !valueutil.IsHexLength(cune, 96) {
		return errors.New("CUNE de nómina electrónica inválido")
	}
	intento := 0
	if registrarIntento {
		intento = 1
	}
	result, err := execSQLCompatContext(ctx, dbConn, `UPDATE empresa_contabilidad_nomina_electronica SET
		estado_dian = ?, cune = COALESCE(NULLIF(?, ''), cune),
		respuesta_dian = COALESCE(NULLIF(?, ''), respuesta_dian), fuente_fiscal_sellada = ?,
		intentos = intentos + ?, fecha_ultimo_intento = CASE WHEN ? = 1 THEN CURRENT_TIMESTAMP ELSE fecha_ultimo_intento END,
		fecha_actualizacion = CURRENT_TIMESTAMP WHERE empresa_id = ? AND id = ?`, estado, cune,
		strings.TrimSpace(respuesta), fuenteSellada, intento, intento, empresaID, nominaID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return errors.New("nómina electrónica no encontrada para la empresa")
	}
	return nil
}
