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

type EmpresaNominaProfesionalDemoResult struct {
	EmpresaID          int64                                  `json:"empresa_id"`
	PeriodoDesde       string                                 `json:"periodo_desde"`
	PeriodoHasta       string                                 `json:"periodo_hasta"`
	Empleados          []EmpresaNominaEmpleado                `json:"empleados"`
	AsistenciasCreadas int                                    `json:"asistencias_creadas"`
	NovedadesCreadas   int                                    `json:"novedades_creadas"`
	Liquidacion        *EmpresaNominaCalculoResult            `json:"liquidacion,omitempty"`
	PILA               []EmpresaNominaPILAResumenColombia     `json:"pila"`
	Pagos              *EmpresaNominaPagosResult              `json:"pagos,omitempty"`
	Dashboard          EmpresaNominaColombiaAvanzadaDashboard `json:"dashboard"`
	Mensajes           []string                               `json:"mensajes,omitempty"`
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

func SetEmpresaNominaNovedadColombiaEstadoAprobacion(dbConn *sql.DB, empresaID, id int64, estadoAprobacion, usuario string) error {
	if empresaID <= 0 || id <= 0 {
		return errors.New("empresa_id e id son obligatorios")
	}
	estado := normalizeNominaAprobacionColombia(estadoAprobacion)
	res, err := ExecCompat(dbConn, `UPDATE empresa_nomina_colombia_novedades
		SET estado_aprobacion=?, fecha_actualizacion=CAST(CURRENT_TIMESTAMP AS TEXT), usuario_creador=COALESCE(NULLIF(?,''), usuario_creador)
		WHERE empresa_id=? AND id=? AND estado='activo'`, estado, strings.TrimSpace(usuario), empresaID, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
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
	for _, row := range nominaColombiaConceptosProfesionales(empresaID, usuario) {
		if _, err := UpsertEmpresaNominaConceptoColombia(dbConn, row); err != nil {
			return err
		}
	}
	return nil
}

func SeedEmpresaNominaProfesionalDemo(dbConn *sql.DB, empresaID int64, usuario string) (*EmpresaNominaProfesionalDemoResult, error) {
	if empresaID <= 0 {
		return nil, errors.New("empresa_id es obligatorio")
	}
	if err := EnsureEmpresaNominaSchema(dbConn); err != nil {
		return nil, err
	}
	if err := EnsureEmpresaAsistenciaSchema(dbConn); err != nil {
		return nil, err
	}
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = "sistema"
	}
	if err := SeedEmpresaNominaColombiaAvanzadaDemo(dbConn, empresaID, usuario); err != nil {
		return nil, err
	}

	desde, hasta := nominaDemoPeriodoActual()
	res := &EmpresaNominaProfesionalDemoResult{EmpresaID: empresaID, PeriodoDesde: desde, PeriodoHasta: hasta}

	cfg := defaultEmpresaNominaConfiguracion(empresaID)
	cfg.UsuarioCreador = usuario
	cfg.Observaciones = "Configuracion demo profesional de nomina para Motel Calipso"
	if _, err := UpsertEmpresaNominaConfiguracion(dbConn, cfg); err != nil {
		res.Mensajes = append(res.Mensajes, fmt.Sprintf("No se pudo guardar configuracion demo: %v", err))
	}

	empleados, err := upsertNominaDemoEmpleados(dbConn, empresaID, usuario)
	if err != nil {
		return nil, err
	}
	res.Empleados = empleados

	asistencias, err := seedNominaDemoAsistencias(dbConn, empresaID, desde, hasta, empleados, usuario)
	if err != nil {
		return nil, err
	}
	res.AsistenciasCreadas = asistencias

	novedades, err := seedNominaDemoNovedades(dbConn, empresaID, desde, hasta, empleados, usuario)
	if err != nil {
		return nil, err
	}
	res.NovedadesCreadas = novedades

	liq, err := GenerateEmpresaNominaLiquidaciones(dbConn, EmpresaNominaCalculoRequest{
		EmpresaID:      empresaID,
		PeriodoDesde:   desde,
		PeriodoHasta:   hasta,
		Overwrite:      true,
		UsuarioCreador: usuario,
		Observaciones:  "demo profesional motel calipso",
	})
	if err != nil {
		return nil, err
	}
	res.Liquidacion = liq

	pila, err := GenerarEmpresaNominaPILAResumenColombia(dbConn, empresaID, desde, hasta, usuario)
	if err != nil {
		return nil, err
	}
	res.PILA = pila

	pagos, err := GenerateEmpresaNominaPagos(dbConn, empresaID, desde, hasta, 0, "transferencia_bancaria", "demo-motel-calipso", usuario)
	if err != nil {
		res.Mensajes = append(res.Mensajes, fmt.Sprintf("Liquidacion y PILA creadas; pagos no generados: %v", err))
	} else {
		res.Pagos = pagos
	}

	dashboard, err := BuildEmpresaNominaColombiaAvanzadaDashboard(dbConn, empresaID, nominaColombiaPeriodo(desde, hasta))
	if err != nil {
		return nil, err
	}
	res.Dashboard = dashboard
	return res, nil
}

func nominaColombiaConceptosProfesionales(empresaID int64, usuario string) []EmpresaNominaConceptoColombia {
	base := func(codigo, nombre, tipo, cuenta string, ibc, pila, electronica bool, porcentaje float64) EmpresaNominaConceptoColombia {
		return EmpresaNominaConceptoColombia{
			EmpresaID: empresaID, Codigo: codigo, Nombre: nombre, Tipo: tipo,
			BaseCotizacion: ibc, AfectaPILA: pila, AfectaNominaElectronica: electronica,
			Porcentaje: porcentaje, CuentaContable: cuenta, Estado: "activo", UsuarioCreador: usuario,
		}
	}
	return []EmpresaNominaConceptoColombia{
		base("BASICO", "Salario basico", "devengado", "510506", true, true, true, 0),
		base("AUXTRANS", "Auxilio de transporte", "devengado", "510527", false, false, true, 0),
		base("HED", "Hora extra diurna", "devengado", "510515", true, true, true, 125),
		base("HEN", "Hora extra nocturna", "devengado", "510515", true, true, true, 175),
		base("RECNOCT", "Recargo nocturno", "devengado", "510515", true, true, true, 35),
		base("DOMFEST", "Dominical o festivo", "devengado", "510515", true, true, true, 75),
		base("BONO", "Bonificacion no salarial", "devengado", "510548", false, false, true, 0),
		base("COMISION", "Comisiones", "devengado", "510548", true, true, true, 0),
		base("VACACIONES", "Vacaciones disfrutadas", "devengado", "510530", true, true, true, 0),
		base("INCAP", "Incapacidad reconocida", "devengado", "510536", true, true, true, 0),
		base("SALUD", "Deduccion salud empleado", "deduccion", "237005", false, true, true, 4),
		base("PENSION", "Deduccion pension empleado", "deduccion", "238030", false, true, true, 4),
		base("SOLIDARIDAD", "Fondo de solidaridad pensional", "deduccion", "238095", false, true, true, 1),
		base("PRESTAMO", "Prestamo o anticipo", "deduccion", "136595", false, false, true, 0),
		base("EMBARGO", "Embargo judicial", "deduccion", "237090", false, false, true, 0),
		base("RETEFTE", "Retencion en la fuente", "deduccion", "236505", false, false, true, 0),
		base("APOSALUD", "Aporte salud empleador", "aporte", "510568", false, true, false, 8.5),
		base("APOPENSION", "Aporte pension empleador", "aporte", "510570", false, true, false, 12),
		base("ARL", "Riesgos laborales ARL", "aporte", "510572", false, true, false, 0.522),
		base("CAJA", "Caja de compensacion", "aporte", "510578", false, true, false, 4),
		base("ICBF", "Aporte ICBF", "aporte", "510575", false, true, false, 3),
		base("SENA", "Aporte SENA", "aporte", "510578", false, true, false, 2),
		base("CESANTIAS", "Provision cesantias", "provision", "510530", false, false, false, 8.33),
		base("INTCES", "Provision intereses cesantias", "provision", "510533", false, false, false, 1),
		base("PRIMA", "Provision prima de servicios", "provision", "510536", false, false, false, 8.33),
		base("PROVVAC", "Provision vacaciones", "provision", "510539", false, false, false, 4.17),
	}
}

func nominaDemoPeriodoActual() (string, string) {
	now := time.Now()
	desde := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	hasta := now
	if hasta.Day() < 15 {
		hasta = time.Date(now.Year(), now.Month(), 15, 0, 0, 0, 0, now.Location())
	}
	return desde.Format("2006-01-02"), hasta.Format("2006-01-02")
}

func upsertNominaDemoEmpleados(dbConn *sql.DB, empresaID int64, usuario string) ([]EmpresaNominaEmpleado, error) {
	out := make([]EmpresaNominaEmpleado, 0, len(nominaDemoEmpleados(empresaID, usuario)))
	for _, empleado := range nominaDemoEmpleados(empresaID, usuario) {
		existing, err := findNominaEmpleadoByCodigo(dbConn, empresaID, empleado.EmpleadoCodigo)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		if existing != nil && existing.ID > 0 {
			empleado.ID = existing.ID
			if err := UpdateEmpresaNominaEmpleado(dbConn, empleado); err != nil {
				return nil, err
			}
			_ = SetEmpresaNominaEmpleadoEstado(dbConn, empresaID, existing.ID, "activo")
			empleado.ID = existing.ID
		} else {
			id, err := CreateEmpresaNominaEmpleado(dbConn, empleado)
			if err != nil {
				return nil, err
			}
			empleado.ID = id
		}
		out = append(out, empleado)
	}
	return out, nil
}

func nominaDemoEmpleados(empresaID int64, usuario string) []EmpresaNominaEmpleado {
	return []EmpresaNominaEmpleado{
		{EmpresaID: empresaID, EmpleadoCodigo: "CAL-NOM-001", EmpleadoNombre: "Ana Maria Rojas", EmpleadoDocumento: "1002003001", Cargo: "Recepcionista", SedeCodigo: "CAL-SM", SedeNombre: "Motel Calipso - Sede Principal", CentroCosto: "OPERACION", TipoContrato: "indefinido", FechaIngreso: "2025-11-03", SalarioBasicoMensual: 1800000, AuxilioTransporteMensual: 200000, BonificacionFijaMensual: 120000, JornadaHorasDia: 8, IncluirAuxilioTransporte: true, Estado: "activo", UsuarioCreador: usuario},
		{EmpresaID: empresaID, EmpleadoCodigo: "CAL-NOM-002", EmpleadoNombre: "Jose David Perez", EmpleadoDocumento: "1002003002", Cargo: "Cajero nocturno", SedeCodigo: "CAL-SM", SedeNombre: "Motel Calipso - Sede Principal", CentroCosto: "CAJA", TipoContrato: "indefinido", FechaIngreso: "2025-09-15", SalarioBasicoMensual: 2100000, AuxilioTransporteMensual: 200000, BonificacionFijaMensual: 180000, JornadaHorasDia: 8, IncluirAuxilioTransporte: true, Estado: "activo", UsuarioCreador: usuario},
		{EmpresaID: empresaID, EmpleadoCodigo: "CAL-NOM-003", EmpleadoNombre: "Luz Marina Gomez", EmpleadoDocumento: "1002003003", Cargo: "Servicio de limpieza", SedeCodigo: "CAL-ROD", SedeNombre: "Motel Calipso - Rodadero", CentroCosto: "ASEO", TipoContrato: "fijo", FechaIngreso: "2026-01-10", SalarioBasicoMensual: 1600000, AuxilioTransporteMensual: 200000, JornadaHorasDia: 8, IncluirAuxilioTransporte: true, Estado: "activo", UsuarioCreador: usuario},
		{EmpresaID: empresaID, EmpleadoCodigo: "CAL-NOM-004", EmpleadoNombre: "Mateo Sierra", EmpleadoDocumento: "1002003004", Cargo: "Administrador de turno", SedeCodigo: "CAL-ADM", SedeNombre: "Motel Calipso - Administracion", CentroCosto: "ADMIN", TipoContrato: "indefinido", FechaIngreso: "2024-08-01", SalarioBasicoMensual: 2800000, BonificacionFijaMensual: 250000, JornadaHorasDia: 8, IncluirAuxilioTransporte: false, Estado: "activo", UsuarioCreador: usuario},
		{EmpresaID: empresaID, EmpleadoCodigo: "CAL-NOM-005", EmpleadoNombre: "Karen Paola Ruiz", EmpleadoDocumento: "1002003005", Cargo: "Auxiliar de servicios", SedeCodigo: "CAL-ROD", SedeNombre: "Motel Calipso - Rodadero", CentroCosto: "OPERACION", TipoContrato: "obra_labor", FechaIngreso: "2026-02-04", SalarioBasicoMensual: 1700000, AuxilioTransporteMensual: 200000, JornadaHorasDia: 8, IncluirAuxilioTransporte: true, Estado: "activo", UsuarioCreador: usuario},
	}
}

func findNominaEmpleadoByCodigo(dbConn *sql.DB, empresaID int64, codigo string) (*EmpresaNominaEmpleado, error) {
	rows, err := ListEmpresaNominaEmpleados(dbConn, empresaID, true, strings.TrimSpace(codigo), 20)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if strings.EqualFold(strings.TrimSpace(row.EmpleadoCodigo), strings.TrimSpace(codigo)) {
			item := row
			return &item, nil
		}
	}
	return nil, sql.ErrNoRows
}

func seedNominaDemoAsistencias(dbConn *sql.DB, empresaID int64, desde, hasta string, empleados []EmpresaNominaEmpleado, usuario string) (int, error) {
	if _, err := ExecCompat(dbConn, `DELETE FROM empresa_asistencia_empleados
		WHERE empresa_id=? AND empleado_codigo LIKE 'CAL-NOM-%' AND fecha_asistencia>=? AND fecha_asistencia<=?`, empresaID, desde, hasta); err != nil {
		return 0, err
	}
	dias, err := nominaDemoDiasLaborales(desde, hasta)
	if err != nil {
		return 0, err
	}
	total := 0
	for _, empleado := range empleados {
		for i, dia := range dias {
			turno, entrada, salida := "manana", "08:00:00", "16:00:00"
			if strings.Contains(strings.ToLower(empleado.Cargo), "nocturno") {
				turno, entrada, salida = "noche", "22:00:00", "06:00:00"
			} else if i%5 == 4 {
				turno, entrada, salida = "tarde", "14:00:00", "22:00:00"
			}
			if _, err := CreateEmpresaAsistenciaEmpleado(dbConn, EmpresaAsistenciaEmpleado{
				EmpresaID: empleado.EmpresaID, EmpleadoID: empleado.ID, EmpleadoCodigo: empleado.EmpleadoCodigo,
				EmpleadoNombre: empleado.EmpleadoNombre, EmpleadoDocumento: empleado.EmpleadoDocumento, Cargo: empleado.Cargo,
				Turno: turno, FechaAsistencia: dia, HoraEntrada: entrada, HoraSalida: salida,
				EstadoAsistencia: "presente", Novedad: "demo nomina motel calipso",
				UsuarioCreador: usuario, Estado: "activo",
			}); err != nil {
				return total, err
			}
			total++
		}
	}
	return total, nil
}

func nominaDemoDiasLaborales(desde, hasta string) ([]string, error) {
	start, err := time.Parse("2006-01-02", desde)
	if err != nil {
		return nil, err
	}
	end, err := time.Parse("2006-01-02", hasta)
	if err != nil {
		return nil, err
	}
	out := []string{}
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		if day.Weekday() == time.Sunday {
			continue
		}
		out = append(out, day.Format("2006-01-02"))
		if len(out) >= 24 {
			break
		}
	}
	return out, nil
}

func seedNominaDemoNovedades(dbConn *sql.DB, empresaID int64, desde, hasta string, empleados []EmpresaNominaEmpleado, usuario string) (int, error) {
	if _, err := ExecCompat(dbConn, `DELETE FROM empresa_nomina_colombia_novedades
		WHERE empresa_id=? AND periodo_desde=? AND periodo_hasta=? AND descripcion LIKE 'Demo Motel Calipso:%'`, empresaID, desde, hasta); err != nil {
		return 0, err
	}
	byCode := map[string]EmpresaNominaEmpleado{}
	for _, empleado := range empleados {
		byCode[empleado.EmpleadoCodigo] = empleado
	}
	defs := []struct {
		empleado string
		tipo     string
		concepto string
		desc     string
		cantidad float64
		unitario float64
		ibc      bool
	}{
		{"CAL-NOM-002", "devengado", "RECNOCT", "Demo Motel Calipso: recargo nocturno controlado", 6, 18000, true},
		{"CAL-NOM-003", "devengado", "HED", "Demo Motel Calipso: horas extra aseo y cierre", 4, 22000, true},
		{"CAL-NOM-001", "deduccion", "PRESTAMO", "Demo Motel Calipso: descuento prestamo interno", 1, 50000, false},
		{"CAL-NOM-004", "devengado", "BONO", "Demo Motel Calipso: bono de responsabilidad", 1, 120000, false},
	}
	total := 0
	for _, def := range defs {
		empleado, ok := byCode[def.empleado]
		if !ok || empleado.ID <= 0 {
			continue
		}
		if _, err := CreateEmpresaNominaNovedadColombia(dbConn, EmpresaNominaNovedadColombia{
			EmpresaID: empresaID, EmpleadoNominaID: empleado.ID, PeriodoDesde: desde, PeriodoHasta: hasta, FechaNovedad: desde,
			Tipo: def.tipo, CodigoConcepto: def.concepto, Descripcion: def.desc, Cantidad: def.cantidad, ValorUnitario: def.unitario,
			AfectaIBC: def.ibc, EstadoAprobacion: "aprobado", Estado: "activo", UsuarioCreador: usuario,
		}); err != nil {
			return total, err
		}
		total++
	}
	return total, nil
}

func aplicarNovedadesColombiaEnLiquidacion(dbConn *sql.DB, liq *EmpresaNominaLiquidacion, cfg *EmpresaNominaConfiguracion) error {
	if dbConn == nil || liq == nil || cfg == nil || liq.EmpresaID <= 0 || liq.EmpleadoNominaID <= 0 {
		return nil
	}
	if err := EnsureEmpresaNominaColombiaAvanzadaSchema(dbConn); err != nil {
		return err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,empleado_nomina_id,periodo_desde,periodo_hasta,fecha_novedad,tipo,concepto_id,codigo_concepto,descripcion,cantidad,valor_unitario,valor_total,afecta_ibc,estado_aprobacion,estado,usuario_creador
		FROM empresa_nomina_colombia_novedades
		WHERE empresa_id=? AND empleado_nomina_id=? AND estado='activo' AND estado_aprobacion='aprobado'
		  AND periodo_desde<=? AND periodo_hasta>=?
		ORDER BY fecha_novedad ASC, id ASC`, liq.EmpresaID, liq.EmpleadoNominaID, liq.PeriodoHasta, liq.PeriodoDesde)
	if err != nil {
		return err
	}
	defer rows.Close()
	novedades := []EmpresaNominaNovedadColombia{}
	for rows.Next() {
		var item EmpresaNominaNovedadColombia
		var afectaIBC int
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.EmpleadoNominaID, &item.PeriodoDesde, &item.PeriodoHasta, &item.FechaNovedad, &item.Tipo, &item.ConceptoID, &item.CodigoConcepto, &item.Descripcion, &item.Cantidad, &item.ValorUnitario, &item.ValorTotal, &afectaIBC, &item.EstadoAprobacion, &item.Estado, &item.UsuarioCreador); err != nil {
			return err
		}
		item.AfectaIBC = afectaIBC > 0
		novedades = append(novedades, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	aplicarNovedadesAprobadasEnLiquidacion(liq, cfg, novedades)
	return nil
}

func aplicarNovedadesAprobadasEnLiquidacion(liq *EmpresaNominaLiquidacion, cfg *EmpresaNominaConfiguracion, novedades []EmpresaNominaNovedadColombia) (int, float64, float64) {
	if liq == nil || cfg == nil || len(novedades) == 0 {
		return 0, 0, 0
	}
	devengadoNovedades := 0.0
	deduccionNovedades := 0.0
	ibcExtra := 0.0
	aplicadas := 0
	for _, item := range novedades {
		if normalizeNominaAprobacionColombia(item.EstadoAprobacion) != "aprobado" || normalizeNominaEstado(item.Estado) != "activo" {
			continue
		}
		valor := round2(item.ValorTotal)
		if valor <= 0 {
			valor = round2(item.Cantidad * item.ValorUnitario)
		}
		switch normalizeNominaColombiaTipo(item.Tipo) {
		case "deduccion":
			deduccionNovedades = round2(deduccionNovedades + valor)
		case "devengado":
			devengadoNovedades = round2(devengadoNovedades + valor)
			if item.AfectaIBC {
				ibcExtra = round2(ibcExtra + valor)
			}
		}
		aplicadas++
	}
	if aplicadas == 0 {
		return 0, 0, 0
	}
	liq.Bonificacion = round2(liq.Bonificacion + devengadoNovedades)
	liq.OtrasDeducciones = round2(liq.OtrasDeducciones + deduccionNovedades)
	liq.DevengadoTotal = round2(liq.DevengadoTotal + devengadoNovedades)
	liq.IngresoBaseCotizacion = round2(liq.IngresoBaseCotizacion + ibcExtra)
	liq.DeduccionSalud = round2(liq.IngresoBaseCotizacion * (cfg.DeduccionSaludPorcentaje / 100.0))
	liq.DeduccionPension = round2(liq.IngresoBaseCotizacion * (cfg.DeduccionPensionPorcentaje / 100.0))
	liq.DeduccionFondoSolidaridad = round2(liq.IngresoBaseCotizacion * (cfg.DeduccionFondoSolidaridadPorcentaje / 100.0))
	liq.DeduccionTotal = round2(liq.DeduccionSalud + liq.DeduccionPension + liq.DeduccionFondoSolidaridad + liq.DeduccionFija + liq.OtrasDeducciones)
	liq.NetoPagar = round2(liq.DevengadoTotal - liq.DeduccionTotal)
	liq.ResumenJSON = appendNominaResumenJSON(liq.ResumenJSON, fmt.Sprintf(`"novedades_colombia":%d,"devengado_novedades":%.2f,"deduccion_novedades":%.2f`, aplicadas, devengadoNovedades, deduccionNovedades))
	return aplicadas, devengadoNovedades, deduccionNovedades
}

func appendNominaResumenJSON(base, fragment string) string {
	base = strings.TrimSpace(base)
	fragment = strings.TrimSpace(fragment)
	if fragment == "" {
		return base
	}
	if base == "" || base == "{}" {
		return "{" + fragment + "}"
	}
	if strings.HasSuffix(base, "}") {
		return strings.TrimSuffix(base, "}") + "," + fragment + "}"
	}
	return base
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
