package db

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

type EmpresaNominaConceptoColombia struct {
	ID                      int64   `json:"id"`
	EmpresaID               int64   `json:"empresa_id"`
	Codigo                  string  `json:"codigo"`
	Nombre                  string  `json:"nombre"`
	Tipo                    string  `json:"tipo"`
	BaseCotizacion          bool    `json:"base_cotizacion"`
	AfectaPILA              bool    `json:"afecta_pila"`
	AfectaNominaElectronica bool    `json:"afecta_nomina_electronica"`
	Porcentaje              float64 `json:"porcentaje"`
	ValorFijo               float64 `json:"valor_fijo"`
	CuentaContable          string  `json:"cuenta_contable,omitempty"`
	Estado                  string  `json:"estado"`
	FechaCreacion           string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion      string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador          string  `json:"usuario_creador,omitempty"`
}

type EmpresaNominaNovedadColombia struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	EmpleadoNominaID   int64   `json:"empleado_nomina_id"`
	EmpleadoNombre     string  `json:"empleado_nombre,omitempty"`
	PeriodoDesde       string  `json:"periodo_desde"`
	PeriodoHasta       string  `json:"periodo_hasta"`
	FechaNovedad       string  `json:"fecha_novedad"`
	Tipo               string  `json:"tipo"`
	ConceptoID         int64   `json:"concepto_id,omitempty"`
	CodigoConcepto     string  `json:"codigo_concepto,omitempty"`
	Descripcion        string  `json:"descripcion"`
	Cantidad           float64 `json:"cantidad"`
	ValorUnitario      float64 `json:"valor_unitario"`
	ValorTotal         float64 `json:"valor_total"`
	AfectaIBC          bool    `json:"afecta_ibc"`
	EstadoAprobacion   string  `json:"estado_aprobacion"`
	Estado             string  `json:"estado"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
}

type EmpresaNominaPILAResumenColombia struct {
	ID                int64   `json:"id"`
	EmpresaID         int64   `json:"empresa_id"`
	Periodo           string  `json:"periodo"`
	EmpleadoNominaID  int64   `json:"empleado_nomina_id"`
	EmpleadoNombre    string  `json:"empleado_nombre"`
	EmpleadoDocumento string  `json:"empleado_documento,omitempty"`
	IBC               float64 `json:"ibc"`
	SaludEmpleado     float64 `json:"salud_empleado"`
	PensionEmpleado   float64 `json:"pension_empleado"`
	SaludEmpleador    float64 `json:"salud_empleador"`
	PensionEmpleador  float64 `json:"pension_empleador"`
	ARL               float64 `json:"arl"`
	CajaCompensacion  float64 `json:"caja_compensacion"`
	ICBF              float64 `json:"icbf"`
	SENA              float64 `json:"sena"`
	TotalAportes      float64 `json:"total_aportes"`
	Estado            string  `json:"estado"`
	FechaGeneracion   string  `json:"fecha_generacion,omitempty"`
	UsuarioCreador    string  `json:"usuario_creador,omitempty"`
}

type EmpresaNominaColombiaAvanzadaDashboard struct {
	EmpresaID           int64                              `json:"empresa_id"`
	ConceptosActivos    int                                `json:"conceptos_activos"`
	NovedadesPendientes int                                `json:"novedades_pendientes"`
	NovedadesAprobadas  int                                `json:"novedades_aprobadas"`
	TotalPILA           float64                            `json:"total_pila"`
	Conceptos           []EmpresaNominaConceptoColombia    `json:"conceptos"`
	Novedades           []EmpresaNominaNovedadColombia     `json:"novedades"`
	PILA                []EmpresaNominaPILAResumenColombia `json:"pila"`
}

func EnsureEmpresaNominaColombiaAvanzadaSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_nomina_colombia_conceptos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			tipo TEXT DEFAULT 'devengado',
			base_cotizacion INTEGER DEFAULT 1,
			afecta_pila INTEGER DEFAULT 1,
			afecta_nomina_electronica INTEGER DEFAULT 1,
			porcentaje NUMERIC(9,4) DEFAULT 0,
			valor_fijo NUMERIC(14,2) DEFAULT 0,
			cuenta_contable TEXT,
			estado TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_nomina_co_concepto_codigo ON empresa_nomina_colombia_conceptos(empresa_id,codigo)`,
		`CREATE INDEX IF NOT EXISTS ix_nomina_co_concepto_estado ON empresa_nomina_colombia_conceptos(empresa_id,tipo,estado)`,
		`CREATE TABLE IF NOT EXISTS empresa_nomina_colombia_novedades (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			empleado_nomina_id BIGINT NOT NULL,
			periodo_desde TEXT NOT NULL,
			periodo_hasta TEXT NOT NULL,
			fecha_novedad TEXT NOT NULL,
			tipo TEXT DEFAULT 'devengado',
			concepto_id BIGINT DEFAULT 0,
			codigo_concepto TEXT,
			descripcion TEXT NOT NULL,
			cantidad NUMERIC(14,4) DEFAULT 1,
			valor_unitario NUMERIC(14,2) DEFAULT 0,
			valor_total NUMERIC(14,2) DEFAULT 0,
			afecta_ibc INTEGER DEFAULT 1,
			estado_aprobacion TEXT DEFAULT 'pendiente',
			estado TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_nomina_co_novedad_empresa_periodo ON empresa_nomina_colombia_novedades(empresa_id,periodo_desde,periodo_hasta,estado_aprobacion)`,
		`CREATE TABLE IF NOT EXISTS empresa_nomina_colombia_pila_resumen (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			periodo TEXT NOT NULL,
			empleado_nomina_id BIGINT NOT NULL,
			empleado_nombre TEXT NOT NULL,
			empleado_documento TEXT,
			ibc NUMERIC(14,2) DEFAULT 0,
			salud_empleado NUMERIC(14,2) DEFAULT 0,
			pension_empleado NUMERIC(14,2) DEFAULT 0,
			salud_empleador NUMERIC(14,2) DEFAULT 0,
			pension_empleador NUMERIC(14,2) DEFAULT 0,
			arl NUMERIC(14,2) DEFAULT 0,
			caja_compensacion NUMERIC(14,2) DEFAULT 0,
			icbf NUMERIC(14,2) DEFAULT 0,
			sena NUMERIC(14,2) DEFAULT 0,
			total_aportes NUMERIC(14,2) DEFAULT 0,
			estado TEXT DEFAULT 'generado',
			fecha_generacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_nomina_co_pila_empleado_periodo ON empresa_nomina_colombia_pila_resumen(empresa_id,periodo,empleado_nomina_id)`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func UpsertEmpresaNominaConceptoColombia(dbConn *sql.DB, item EmpresaNominaConceptoColombia) (int64, error) {
	if err := EnsureEmpresaNominaColombiaAvanzadaSchema(dbConn); err != nil {
		return 0, err
	}
	item = normalizeNominaConceptoColombia(item)
	if item.EmpresaID <= 0 || item.Codigo == "" || item.Nombre == "" {
		return 0, errors.New("empresa_id, codigo y nombre son obligatorios")
	}
	if item.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_nomina_colombia_conceptos SET codigo=?,nombre=?,tipo=?,base_cotizacion=?,afecta_pila=?,afecta_nomina_electronica=?,porcentaje=?,valor_fijo=?,cuenta_contable=?,estado=?,fecha_actualizacion=CAST(CURRENT_TIMESTAMP AS TEXT) WHERE empresa_id=? AND id=?`,
			item.Codigo, item.Nombre, item.Tipo, boolIntNominaColombia(item.BaseCotizacion), boolIntNominaColombia(item.AfectaPILA), boolIntNominaColombia(item.AfectaNominaElectronica), item.Porcentaje, item.ValorFijo, item.CuentaContable, item.Estado, item.EmpresaID, item.ID)
		return item.ID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_nomina_colombia_conceptos (empresa_id,codigo,nombre,tipo,base_cotizacion,afecta_pila,afecta_nomina_electronica,porcentaje,valor_fijo,cuenta_contable,estado,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)
	ON CONFLICT (empresa_id,codigo) DO UPDATE SET nombre=EXCLUDED.nombre,tipo=EXCLUDED.tipo,base_cotizacion=EXCLUDED.base_cotizacion,afecta_pila=EXCLUDED.afecta_pila,afecta_nomina_electronica=EXCLUDED.afecta_nomina_electronica,porcentaje=EXCLUDED.porcentaje,valor_fijo=EXCLUDED.valor_fijo,cuenta_contable=EXCLUDED.cuenta_contable,estado=EXCLUDED.estado,fecha_actualizacion=CAST(CURRENT_TIMESTAMP AS TEXT)`,
		item.EmpresaID, item.Codigo, item.Nombre, item.Tipo, boolIntNominaColombia(item.BaseCotizacion), boolIntNominaColombia(item.AfectaPILA), boolIntNominaColombia(item.AfectaNominaElectronica), item.Porcentaje, item.ValorFijo, item.CuentaContable, item.Estado, item.UsuarioCreador)
}

func ListEmpresaNominaConceptosColombia(dbConn *sql.DB, empresaID int64, tipo string, limit int) ([]EmpresaNominaConceptoColombia, error) {
	if err := EnsureEmpresaNominaColombiaAvanzadaSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if strings.TrimSpace(tipo) != "" {
		where += " AND tipo=?"
		args = append(args, normalizeNominaColombiaTipo(tipo))
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,COALESCE(codigo,''),COALESCE(nombre,''),COALESCE(tipo,'devengado'),COALESCE(base_cotizacion,1),COALESCE(afecta_pila,1),COALESCE(afecta_nomina_electronica,1),COALESCE(porcentaje,0),COALESCE(valor_fijo,0),COALESCE(cuenta_contable,''),COALESCE(estado,'activo'),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,''),COALESCE(usuario_creador,'') FROM empresa_nomina_colombia_conceptos WHERE %s ORDER BY tipo,codigo LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaNominaConceptoColombia{}
	for rows.Next() {
		var x EmpresaNominaConceptoColombia
		var base, pila, electronica int
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Codigo, &x.Nombre, &x.Tipo, &base, &pila, &electronica, &x.Porcentaje, &x.ValorFijo, &x.CuentaContable, &x.Estado, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		x.BaseCotizacion = base > 0
		x.AfectaPILA = pila > 0
		x.AfectaNominaElectronica = electronica > 0
		out = append(out, x)
	}
	return out, rows.Err()
}

func CreateEmpresaNominaNovedadColombia(dbConn *sql.DB, item EmpresaNominaNovedadColombia) (int64, error) {
	if err := EnsureEmpresaNominaColombiaAvanzadaSchema(dbConn); err != nil {
		return 0, err
	}
	item = normalizeNominaNovedadColombia(item)
	if item.EmpresaID <= 0 || item.EmpleadoNominaID <= 0 || item.Descripcion == "" {
		return 0, errors.New("empleado y descripcion son obligatorios")
	}
	if item.CodigoConcepto == "" && item.ConceptoID > 0 {
		_ = QueryRowCompat(dbConn, `SELECT COALESCE(codigo,'') FROM empresa_nomina_colombia_conceptos WHERE empresa_id=? AND id=?`, item.EmpresaID, item.ConceptoID).Scan(&item.CodigoConcepto)
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_nomina_colombia_novedades (empresa_id,empleado_nomina_id,periodo_desde,periodo_hasta,fecha_novedad,tipo,concepto_id,codigo_concepto,descripcion,cantidad,valor_unitario,valor_total,afecta_ibc,estado_aprobacion,estado,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		item.EmpresaID, item.EmpleadoNominaID, item.PeriodoDesde, item.PeriodoHasta, item.FechaNovedad, item.Tipo, item.ConceptoID, item.CodigoConcepto, item.Descripcion, item.Cantidad, item.ValorUnitario, item.ValorTotal, boolIntNominaColombia(item.AfectaIBC), item.EstadoAprobacion, item.Estado, item.UsuarioCreador)
}

func ListEmpresaNominaNovedadesColombia(dbConn *sql.DB, empresaID int64, periodoDesde, periodoHasta, estadoAprobacion string, limit int) ([]EmpresaNominaNovedadColombia, error) {
	if err := EnsureEmpresaNominaColombiaAvanzadaSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	args := []interface{}{empresaID}
	where := "n.empresa_id=?"
	if strings.TrimSpace(periodoDesde) != "" {
		where += " AND n.periodo_desde>=?"
		args = append(args, strings.TrimSpace(periodoDesde))
	}
	if strings.TrimSpace(periodoHasta) != "" {
		where += " AND n.periodo_hasta<=?"
		args = append(args, strings.TrimSpace(periodoHasta))
	}
	if strings.TrimSpace(estadoAprobacion) != "" {
		where += " AND n.estado_aprobacion=?"
		args = append(args, normalizeNominaAprobacionColombia(estadoAprobacion))
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT n.id,n.empresa_id,n.empleado_nomina_id,COALESCE(e.empleado_nombre,''),COALESCE(n.periodo_desde,''),COALESCE(n.periodo_hasta,''),COALESCE(n.fecha_novedad,''),COALESCE(n.tipo,'devengado'),COALESCE(n.concepto_id,0),COALESCE(n.codigo_concepto,''),COALESCE(n.descripcion,''),COALESCE(n.cantidad,0),COALESCE(n.valor_unitario,0),COALESCE(n.valor_total,0),COALESCE(n.afecta_ibc,1),COALESCE(n.estado_aprobacion,'pendiente'),COALESCE(n.estado,'activo'),COALESCE(n.fecha_creacion,''),COALESCE(n.fecha_actualizacion,''),COALESCE(n.usuario_creador,'') FROM empresa_nomina_colombia_novedades n LEFT JOIN empresa_nomina_empleados e ON e.id=n.empleado_nomina_id AND e.empresa_id=n.empresa_id WHERE %s ORDER BY n.id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaNominaNovedadColombia{}
	for rows.Next() {
		var x EmpresaNominaNovedadColombia
		var afecta int
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.EmpleadoNominaID, &x.EmpleadoNombre, &x.PeriodoDesde, &x.PeriodoHasta, &x.FechaNovedad, &x.Tipo, &x.ConceptoID, &x.CodigoConcepto, &x.Descripcion, &x.Cantidad, &x.ValorUnitario, &x.ValorTotal, &afecta, &x.EstadoAprobacion, &x.Estado, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		x.AfectaIBC = afecta > 0
		out = append(out, x)
	}
	return out, rows.Err()
}

func GenerarEmpresaNominaPILAResumenColombia(dbConn *sql.DB, empresaID int64, periodoDesde, periodoHasta, usuario string) ([]EmpresaNominaPILAResumenColombia, error) {
	if err := EnsureEmpresaNominaColombiaAvanzadaSchema(dbConn); err != nil {
		return nil, err
	}
	cfg, err := GetEmpresaNominaConfiguracion(dbConn, empresaID)
	if err != nil {
		return nil, err
	}
	liqs, err := ListEmpresaNominaLiquidaciones(dbConn, empresaID, EmpresaNominaLiquidacionFilter{PeriodoDesde: periodoDesde, PeriodoHasta: periodoHasta, IncludeInactive: true, Limit: 1000})
	if err != nil {
		return nil, err
	}
	periodo := nominaColombiaPeriodo(periodoDesde, periodoHasta)
	for _, liq := range liqs {
		row := buildNominaPILARowColombia(empresaID, periodo, liq, cfg, usuario)
		_, err := ExecCompat(dbConn, `INSERT INTO empresa_nomina_colombia_pila_resumen (empresa_id,periodo,empleado_nomina_id,empleado_nombre,empleado_documento,ibc,salud_empleado,pension_empleado,salud_empleador,pension_empleador,arl,caja_compensacion,icbf,sena,total_aportes,estado,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (empresa_id,periodo,empleado_nomina_id) DO UPDATE SET empleado_nombre=EXCLUDED.empleado_nombre,empleado_documento=EXCLUDED.empleado_documento,ibc=EXCLUDED.ibc,salud_empleado=EXCLUDED.salud_empleado,pension_empleado=EXCLUDED.pension_empleado,salud_empleador=EXCLUDED.salud_empleador,pension_empleador=EXCLUDED.pension_empleador,arl=EXCLUDED.arl,caja_compensacion=EXCLUDED.caja_compensacion,icbf=EXCLUDED.icbf,sena=EXCLUDED.sena,total_aportes=EXCLUDED.total_aportes,estado=EXCLUDED.estado,fecha_generacion=CAST(CURRENT_TIMESTAMP AS TEXT),usuario_creador=EXCLUDED.usuario_creador`,
			row.EmpresaID, row.Periodo, row.EmpleadoNominaID, row.EmpleadoNombre, row.EmpleadoDocumento, row.IBC, row.SaludEmpleado, row.PensionEmpleado, row.SaludEmpleador, row.PensionEmpleador, row.ARL, row.CajaCompensacion, row.ICBF, row.SENA, row.TotalAportes, row.Estado, row.UsuarioCreador)
		if err != nil {
			return nil, err
		}
	}
	return ListEmpresaNominaPILAResumenColombia(dbConn, empresaID, periodo, 1000)
}

func ListEmpresaNominaPILAResumenColombia(dbConn *sql.DB, empresaID int64, periodo string, limit int) ([]EmpresaNominaPILAResumenColombia, error) {
	if limit <= 0 || limit > 2000 {
		limit = 1000
	}
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if strings.TrimSpace(periodo) != "" {
		where += " AND periodo=?"
		args = append(args, strings.TrimSpace(periodo))
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,COALESCE(periodo,''),empleado_nomina_id,COALESCE(empleado_nombre,''),COALESCE(empleado_documento,''),COALESCE(ibc,0),COALESCE(salud_empleado,0),COALESCE(pension_empleado,0),COALESCE(salud_empleador,0),COALESCE(pension_empleador,0),COALESCE(arl,0),COALESCE(caja_compensacion,0),COALESCE(icbf,0),COALESCE(sena,0),COALESCE(total_aportes,0),COALESCE(estado,'generado'),COALESCE(fecha_generacion,''),COALESCE(usuario_creador,'') FROM empresa_nomina_colombia_pila_resumen WHERE %s ORDER BY empleado_nombre LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaNominaPILAResumenColombia{}
	for rows.Next() {
		var x EmpresaNominaPILAResumenColombia
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Periodo, &x.EmpleadoNominaID, &x.EmpleadoNombre, &x.EmpleadoDocumento, &x.IBC, &x.SaludEmpleado, &x.PensionEmpleado, &x.SaludEmpleador, &x.PensionEmpleador, &x.ARL, &x.CajaCompensacion, &x.ICBF, &x.SENA, &x.TotalAportes, &x.Estado, &x.FechaGeneracion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func BuildEmpresaNominaColombiaAvanzadaDashboard(dbConn *sql.DB, empresaID int64, periodo string) (EmpresaNominaColombiaAvanzadaDashboard, error) {
	if err := EnsureEmpresaNominaColombiaAvanzadaSchema(dbConn); err != nil {
		return EmpresaNominaColombiaAvanzadaDashboard{}, err
	}
	conceptos, _ := ListEmpresaNominaConceptosColombia(dbConn, empresaID, "", 120)
	novedades, _ := ListEmpresaNominaNovedadesColombia(dbConn, empresaID, "", "", "", 120)
	pila, _ := ListEmpresaNominaPILAResumenColombia(dbConn, empresaID, periodo, 120)
	ds := EmpresaNominaColombiaAvanzadaDashboard{EmpresaID: empresaID, Conceptos: conceptos, Novedades: novedades, PILA: pila}
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_nomina_colombia_conceptos WHERE empresa_id=? AND estado='activo'`, empresaID).Scan(&ds.ConceptosActivos)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_nomina_colombia_novedades WHERE empresa_id=? AND estado_aprobacion='pendiente' AND estado='activo'`, empresaID).Scan(&ds.NovedadesPendientes)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_nomina_colombia_novedades WHERE empresa_id=? AND estado_aprobacion='aprobado' AND estado='activo'`, empresaID).Scan(&ds.NovedadesAprobadas)
	for _, row := range pila {
		ds.TotalPILA += row.TotalAportes
	}
	return ds, nil
}

func SeedEmpresaNominaColombiaAvanzadaDemo(dbConn *sql.DB, empresaID int64, usuario string) error {
	if err := EnsureEmpresaNominaSchema(dbConn); err != nil {
		return err
	}
	defaults := []EmpresaNominaConceptoColombia{
		{EmpresaID: empresaID, Codigo: "BASICO", Nombre: "Salario basico", Tipo: "devengado", BaseCotizacion: true, AfectaPILA: true, AfectaNominaElectronica: true, CuentaContable: "510506", Estado: "activo", UsuarioCreador: usuario},
		{EmpresaID: empresaID, Codigo: "AUXTRANS", Nombre: "Auxilio de transporte", Tipo: "devengado", BaseCotizacion: false, AfectaPILA: false, AfectaNominaElectronica: true, CuentaContable: "510527", Estado: "activo", UsuarioCreador: usuario},
		{EmpresaID: empresaID, Codigo: "SALUD", Nombre: "Deduccion salud empleado", Tipo: "deduccion", BaseCotizacion: false, AfectaPILA: true, AfectaNominaElectronica: true, Porcentaje: 4, CuentaContable: "237005", Estado: "activo", UsuarioCreador: usuario},
		{EmpresaID: empresaID, Codigo: "PENSION", Nombre: "Deduccion pension empleado", Tipo: "deduccion", BaseCotizacion: false, AfectaPILA: true, AfectaNominaElectronica: true, Porcentaje: 4, CuentaContable: "238030", Estado: "activo", UsuarioCreador: usuario},
		{EmpresaID: empresaID, Codigo: "BONO", Nombre: "Bonificacion no salarial", Tipo: "devengado", BaseCotizacion: false, AfectaPILA: false, AfectaNominaElectronica: true, CuentaContable: "510548", Estado: "activo", UsuarioCreador: usuario},
	}
	for _, row := range defaults {
		if _, err := UpsertEmpresaNominaConceptoColombia(dbConn, row); err != nil {
			return err
		}
	}
	return nil
}

func buildNominaPILARowColombia(empresaID int64, periodo string, liq EmpresaNominaLiquidacion, cfg *EmpresaNominaConfiguracion, usuario string) EmpresaNominaPILAResumenColombia {
	ibc := liq.IngresoBaseCotizacion
	row := EmpresaNominaPILAResumenColombia{EmpresaID: empresaID, Periodo: periodo, EmpleadoNominaID: liq.EmpleadoNominaID, EmpleadoNombre: liq.EmpleadoNombre, EmpleadoDocumento: liq.EmpleadoDocumento, IBC: ibc, Estado: "generado", UsuarioCreador: usuario}
	row.SaludEmpleado = roundNominaColombia(ibc * cfg.DeduccionSaludPorcentaje / 100)
	row.PensionEmpleado = roundNominaColombia(ibc * cfg.DeduccionPensionPorcentaje / 100)
	row.SaludEmpleador = roundNominaColombia(ibc * cfg.AporteSaludEmpleadorPorcentaje / 100)
	row.PensionEmpleador = roundNominaColombia(ibc * cfg.AportePensionEmpleadorPorcentaje / 100)
	row.ARL = roundNominaColombia(ibc * cfg.AporteARLPorcentaje / 100)
	row.CajaCompensacion = roundNominaColombia(ibc * cfg.AporteCajaCompensacionPorcentaje / 100)
	row.ICBF = roundNominaColombia(ibc * cfg.AporteICBFPorcentaje / 100)
	row.SENA = roundNominaColombia(ibc * cfg.AporteSENAPorcentaje / 100)
	row.TotalAportes = roundNominaColombia(row.SaludEmpleado + row.PensionEmpleado + row.SaludEmpleador + row.PensionEmpleador + row.ARL + row.CajaCompensacion + row.ICBF + row.SENA)
	return row
}

func normalizeNominaConceptoColombia(x EmpresaNominaConceptoColombia) EmpresaNominaConceptoColombia {
	x.Codigo = strings.ToUpper(strings.TrimSpace(x.Codigo))
	x.Nombre = strings.TrimSpace(x.Nombre)
	x.Tipo = normalizeNominaColombiaTipo(x.Tipo)
	x.Porcentaje = normalizeNominaPorcentaje(x.Porcentaje)
	if x.ValorFijo < 0 {
		x.ValorFijo = 0
	}
	x.CuentaContable = strings.TrimSpace(x.CuentaContable)
	x.Estado = normalizeNominaEstado(x.Estado)
	return x
}

func normalizeNominaNovedadColombia(x EmpresaNominaNovedadColombia) EmpresaNominaNovedadColombia {
	x.Tipo = normalizeNominaColombiaTipo(x.Tipo)
	x.CodigoConcepto = strings.ToUpper(strings.TrimSpace(x.CodigoConcepto))
	x.Descripcion = strings.TrimSpace(x.Descripcion)
	if x.Cantidad <= 0 {
		x.Cantidad = 1
	}
	if x.ValorUnitario < 0 {
		x.ValorUnitario = 0
	}
	if x.ValorTotal <= 0 {
		x.ValorTotal = roundNominaColombia(x.Cantidad * x.ValorUnitario)
	}
	if x.FechaNovedad == "" {
		x.FechaNovedad = x.PeriodoDesde
	}
	x.EstadoAprobacion = normalizeNominaAprobacionColombia(x.EstadoAprobacion)
	x.Estado = normalizeNominaEstado(x.Estado)
	return x
}

func normalizeNominaColombiaTipo(v string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case "deduccion", "aporte", "provision":
		return s
	default:
		return "devengado"
	}
}

func normalizeNominaAprobacionColombia(v string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case "aprobado", "rechazado", "anulado":
		return s
	default:
		return "pendiente"
	}
}

func nominaColombiaPeriodo(desde, hasta string) string {
	if len(strings.TrimSpace(desde)) >= 7 {
		return strings.TrimSpace(desde)[:7]
	}
	if len(strings.TrimSpace(hasta)) >= 7 {
		return strings.TrimSpace(hasta)[:7]
	}
	return time.Now().Format("2006-01")
}

func roundNominaColombia(v float64) float64 {
	if v < 0 {
		v = 0
	}
	return math.Round(v*100) / 100
}

func boolIntNominaColombia(v bool) int {
	if v {
		return 1
	}
	return 0
}
