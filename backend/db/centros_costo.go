package db

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

type EmpresaCentroCosto struct {
	ID                    int64   `json:"id"`
	EmpresaID             int64   `json:"empresa_id"`
	Codigo                string  `json:"codigo"`
	Nombre                string  `json:"nombre"`
	Tipo                  string  `json:"tipo"`
	Nivel                 int     `json:"nivel"`
	PadreID               int64   `json:"padre_id,omitempty"`
	PadreCodigo           string  `json:"padre_codigo,omitempty"`
	Responsable           string  `json:"responsable,omitempty"`
	Sucursal              string  `json:"sucursal,omitempty"`
	Area                  string  `json:"area,omitempty"`
	UnidadNegocio         string  `json:"unidad_negocio,omitempty"`
	MetaMargenPct         float64 `json:"meta_margen_pct"`
	Estado                string  `json:"estado"`
	FechaInicio           string  `json:"fecha_inicio,omitempty"`
	FechaFin              string  `json:"fecha_fin,omitempty"`
	Observaciones         string  `json:"observaciones,omitempty"`
	FechaCreacion         string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion    string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador        string  `json:"usuario_creador,omitempty"`
	InferidoDeMovimientos bool    `json:"inferido_de_movimientos,omitempty"`
}

type EmpresaCentroCostoRegla struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	CentroCostoID      int64   `json:"centro_costo_id,omitempty"`
	CentroCostoCodigo  string  `json:"centro_costo_codigo"`
	Nombre             string  `json:"nombre"`
	OrigenModulo       string  `json:"origen_modulo"`
	Categoria          string  `json:"categoria,omitempty"`
	TerceroPatron      string  `json:"tercero_patron,omitempty"`
	CuentaPatron       string  `json:"cuenta_patron,omitempty"`
	Porcentaje         float64 `json:"porcentaje"`
	Prioridad          int     `json:"prioridad"`
	Activa             bool    `json:"activa"`
	Estado             string  `json:"estado"`
	Observaciones      string  `json:"observaciones,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
}

type EmpresaCentroCostoPresupuesto struct {
	ID                  int64   `json:"id"`
	EmpresaID           int64   `json:"empresa_id"`
	CentroCostoID       int64   `json:"centro_costo_id,omitempty"`
	CentroCostoCodigo   string  `json:"centro_costo_codigo"`
	CentroCostoNombre   string  `json:"centro_costo_nombre,omitempty"`
	Periodo             string  `json:"periodo"`
	Escenario           string  `json:"escenario"`
	IngresosPresupuesto float64 `json:"ingresos_presupuesto"`
	EgresosPresupuesto  float64 `json:"egresos_presupuesto"`
	MetaMargenPct       float64 `json:"meta_margen_pct"`
	Responsable         string  `json:"responsable,omitempty"`
	Estado              string  `json:"estado"`
	Observaciones       string  `json:"observaciones,omitempty"`
	FechaCreacion       string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion  string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador      string  `json:"usuario_creador,omitempty"`
}

type EmpresaCentroCostoMovimiento struct {
	CentroCostoCodigo string  `json:"centro_costo_codigo"`
	CentroCostoNombre string  `json:"centro_costo_nombre,omitempty"`
	OrigenModulo      string  `json:"origen_modulo"`
	Referencia        string  `json:"referencia,omitempty"`
	Fecha             string  `json:"fecha,omitempty"`
	Periodo           string  `json:"periodo,omitempty"`
	Tipo              string  `json:"tipo"`
	Categoria         string  `json:"categoria,omitempty"`
	Concepto          string  `json:"concepto,omitempty"`
	Ingresos          float64 `json:"ingresos"`
	Egresos           float64 `json:"egresos"`
	Valor             float64 `json:"valor"`
	Estado            string  `json:"estado,omitempty"`
}

type EmpresaCentroCostoRentabilidad struct {
	CentroCostoCodigo     string  `json:"centro_costo_codigo"`
	CentroCostoNombre     string  `json:"centro_costo_nombre"`
	Tipo                  string  `json:"tipo,omitempty"`
	Sucursal              string  `json:"sucursal,omitempty"`
	Area                  string  `json:"area,omitempty"`
	UnidadNegocio         string  `json:"unidad_negocio,omitempty"`
	Ingresos              float64 `json:"ingresos"`
	Egresos               float64 `json:"egresos"`
	PresupuestoIngresos   float64 `json:"presupuesto_ingresos"`
	PresupuestoEgresos    float64 `json:"presupuesto_egresos"`
	Margen                float64 `json:"margen"`
	MargenPct             float64 `json:"margen_pct"`
	EjecucionIngresosPct  float64 `json:"ejecucion_ingresos_pct"`
	EjecucionEgresosPct   float64 `json:"ejecucion_egresos_pct"`
	MetaMargenPct         float64 `json:"meta_margen_pct"`
	Movimientos           int     `json:"movimientos"`
	InferidoDeMovimientos bool    `json:"inferido_de_movimientos,omitempty"`
	Alerta                string  `json:"alerta,omitempty"`
}

type EmpresaCentrosCostoDashboard struct {
	EmpresaID            int64                            `json:"empresa_id"`
	Periodo              string                           `json:"periodo"`
	CentrosActivos       int                              `json:"centros_activos"`
	ReglasActivas        int                              `json:"reglas_activas"`
	MovimientosTotal     int                              `json:"movimientos_total"`
	IngresosTotal        float64                          `json:"ingresos_total"`
	EgresosTotal         float64                          `json:"egresos_total"`
	MargenTotal          float64                          `json:"margen_total"`
	MargenPct            float64                          `json:"margen_pct"`
	PresupuestoIngresos  float64                          `json:"presupuesto_ingresos"`
	PresupuestoEgresos   float64                          `json:"presupuesto_egresos"`
	EjecucionIngresosPct float64                          `json:"ejecucion_ingresos_pct"`
	EjecucionEgresosPct  float64                          `json:"ejecucion_egresos_pct"`
	Rentabilidad         []EmpresaCentroCostoRentabilidad `json:"rentabilidad"`
	MovimientosRecientes []EmpresaCentroCostoMovimiento   `json:"movimientos_recientes"`
	Centros              []EmpresaCentroCosto             `json:"centros"`
	Reglas               []EmpresaCentroCostoRegla        `json:"reglas"`
	Presupuestos         []EmpresaCentroCostoPresupuesto  `json:"presupuestos"`
	Alertas              []string                         `json:"alertas"`
}

func EnsureEmpresaCentrosCostoSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_centros_costo (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			tipo TEXT DEFAULT 'operativo',
			nivel INTEGER DEFAULT 1,
			padre_id BIGINT DEFAULT 0,
			padre_codigo TEXT,
			responsable TEXT,
			sucursal TEXT,
			area TEXT,
			unidad_negocio TEXT,
			meta_margen_pct NUMERIC(8,2) DEFAULT 0,
			estado TEXT DEFAULT 'activo',
			fecha_inicio TEXT,
			fecha_fin TEXT,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_centros_costo_empresa_codigo ON empresa_centros_costo(empresa_id,codigo)`,
		`CREATE INDEX IF NOT EXISTS ix_centros_costo_empresa_estado ON empresa_centros_costo(empresa_id,estado,tipo)`,
		`CREATE TABLE IF NOT EXISTS empresa_centros_costo_reglas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			centro_costo_id BIGINT DEFAULT 0,
			centro_costo_codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			origen_modulo TEXT DEFAULT 'general',
			categoria TEXT,
			tercero_patron TEXT,
			cuenta_patron TEXT,
			porcentaje NUMERIC(8,2) DEFAULT 100,
			prioridad INTEGER DEFAULT 100,
			activa INTEGER DEFAULT 1,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_centros_costo_reglas_empresa ON empresa_centros_costo_reglas(empresa_id,origen_modulo,activa,prioridad)`,
		`CREATE TABLE IF NOT EXISTS empresa_centros_costo_presupuestos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			centro_costo_id BIGINT DEFAULT 0,
			centro_costo_codigo TEXT NOT NULL,
			periodo TEXT NOT NULL,
			escenario TEXT DEFAULT 'base',
			ingresos_presupuesto NUMERIC(14,2) DEFAULT 0,
			egresos_presupuesto NUMERIC(14,2) DEFAULT 0,
			meta_margen_pct NUMERIC(8,2) DEFAULT 0,
			responsable TEXT,
			estado TEXT DEFAULT 'aprobado',
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_centros_costo_pres_empresa ON empresa_centros_costo_presupuestos(empresa_id,centro_costo_codigo,periodo,escenario)`,
		`CREATE INDEX IF NOT EXISTS ix_centros_costo_pres_periodo ON empresa_centros_costo_presupuestos(empresa_id,periodo,estado)`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func UpsertEmpresaCentroCosto(dbConn *sql.DB, item EmpresaCentroCosto) (int64, error) {
	if err := EnsureEmpresaCentrosCostoSchema(dbConn); err != nil {
		return 0, err
	}
	item = normalizeEmpresaCentroCosto(item)
	if item.EmpresaID <= 0 || item.Codigo == "" || item.Nombre == "" {
		return 0, errors.New("empresa_id, codigo y nombre son obligatorios")
	}
	if item.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_centros_costo SET codigo=?,nombre=?,tipo=?,nivel=?,padre_id=?,padre_codigo=?,responsable=?,sucursal=?,area=?,unidad_negocio=?,meta_margen_pct=?,estado=?,fecha_inicio=?,fecha_fin=?,observaciones=?,fecha_actualizacion=CAST(CURRENT_TIMESTAMP AS TEXT) WHERE empresa_id=? AND id=?`,
			item.Codigo, item.Nombre, item.Tipo, item.Nivel, item.PadreID, item.PadreCodigo, item.Responsable, item.Sucursal, item.Area, item.UnidadNegocio, item.MetaMargenPct, item.Estado, item.FechaInicio, item.FechaFin, item.Observaciones, item.EmpresaID, item.ID)
		return item.ID, err
	}
	var id int64
	err := QueryRowCompat(dbConn, `INSERT INTO empresa_centros_costo (empresa_id,codigo,nombre,tipo,nivel,padre_id,padre_codigo,responsable,sucursal,area,unidad_negocio,meta_margen_pct,estado,fecha_inicio,fecha_fin,observaciones,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (empresa_id,codigo) DO UPDATE SET nombre=EXCLUDED.nombre,tipo=EXCLUDED.tipo,nivel=EXCLUDED.nivel,padre_id=EXCLUDED.padre_id,padre_codigo=EXCLUDED.padre_codigo,responsable=EXCLUDED.responsable,sucursal=EXCLUDED.sucursal,area=EXCLUDED.area,unidad_negocio=EXCLUDED.unidad_negocio,meta_margen_pct=EXCLUDED.meta_margen_pct,estado=EXCLUDED.estado,fecha_inicio=EXCLUDED.fecha_inicio,fecha_fin=EXCLUDED.fecha_fin,observaciones=EXCLUDED.observaciones,fecha_actualizacion=CAST(CURRENT_TIMESTAMP AS TEXT),usuario_creador=EXCLUDED.usuario_creador
		RETURNING id`,
		item.EmpresaID, item.Codigo, item.Nombre, item.Tipo, item.Nivel, item.PadreID, item.PadreCodigo, item.Responsable, item.Sucursal, item.Area, item.UnidadNegocio, item.MetaMargenPct, item.Estado, item.FechaInicio, item.FechaFin, item.Observaciones, item.UsuarioCreador).Scan(&id)
	return id, err
}

func ListEmpresaCentrosCosto(dbConn *sql.DB, empresaID int64, estado string, limit int) ([]EmpresaCentroCosto, error) {
	if err := EnsureEmpresaCentrosCostoSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 1000 {
		limit = 300
	}
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if strings.TrimSpace(estado) != "" {
		where += " AND estado=?"
		args = append(args, normalizeCentroCostoEstado(estado))
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,COALESCE(codigo,''),COALESCE(nombre,''),COALESCE(tipo,'operativo'),COALESCE(nivel,1),COALESCE(padre_id,0),COALESCE(padre_codigo,''),COALESCE(responsable,''),COALESCE(sucursal,''),COALESCE(area,''),COALESCE(unidad_negocio,''),COALESCE(meta_margen_pct,0),COALESCE(estado,'activo'),COALESCE(fecha_inicio,''),COALESCE(fecha_fin,''),COALESCE(observaciones,''),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,''),COALESCE(usuario_creador,'') FROM empresa_centros_costo WHERE %s ORDER BY tipo,nivel,codigo LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCentroCosto{}
	for rows.Next() {
		var x EmpresaCentroCosto
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Codigo, &x.Nombre, &x.Tipo, &x.Nivel, &x.PadreID, &x.PadreCodigo, &x.Responsable, &x.Sucursal, &x.Area, &x.UnidadNegocio, &x.MetaMargenPct, &x.Estado, &x.FechaInicio, &x.FechaFin, &x.Observaciones, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func UpsertEmpresaCentroCostoRegla(dbConn *sql.DB, item EmpresaCentroCostoRegla) (int64, error) {
	if err := EnsureEmpresaCentrosCostoSchema(dbConn); err != nil {
		return 0, err
	}
	item = normalizeEmpresaCentroCostoRegla(item)
	if item.EmpresaID <= 0 || item.CentroCostoCodigo == "" || item.Nombre == "" {
		return 0, errors.New("empresa_id, centro_costo_codigo y nombre son obligatorios")
	}
	if item.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_centros_costo_reglas SET centro_costo_id=?,centro_costo_codigo=?,nombre=?,origen_modulo=?,categoria=?,tercero_patron=?,cuenta_patron=?,porcentaje=?,prioridad=?,activa=?,estado=?,observaciones=?,fecha_actualizacion=CAST(CURRENT_TIMESTAMP AS TEXT) WHERE empresa_id=? AND id=?`,
			item.CentroCostoID, item.CentroCostoCodigo, item.Nombre, item.OrigenModulo, item.Categoria, item.TerceroPatron, item.CuentaPatron, item.Porcentaje, item.Prioridad, boolIntCentroCosto(item.Activa), item.Estado, item.Observaciones, item.EmpresaID, item.ID)
		return item.ID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_centros_costo_reglas (empresa_id,centro_costo_id,centro_costo_codigo,nombre,origen_modulo,categoria,tercero_patron,cuenta_patron,porcentaje,prioridad,activa,estado,observaciones,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		item.EmpresaID, item.CentroCostoID, item.CentroCostoCodigo, item.Nombre, item.OrigenModulo, item.Categoria, item.TerceroPatron, item.CuentaPatron, item.Porcentaje, item.Prioridad, boolIntCentroCosto(item.Activa), item.Estado, item.Observaciones, item.UsuarioCreador)
}

func ListEmpresaCentroCostoReglas(dbConn *sql.DB, empresaID int64, origen string, limit int) ([]EmpresaCentroCostoRegla, error) {
	if err := EnsureEmpresaCentrosCostoSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 1000 {
		limit = 300
	}
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if strings.TrimSpace(origen) != "" {
		where += " AND origen_modulo=?"
		args = append(args, normalizeCentroCostoText(origen, "general"))
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,COALESCE(centro_costo_id,0),COALESCE(centro_costo_codigo,''),COALESCE(nombre,''),COALESCE(origen_modulo,'general'),COALESCE(categoria,''),COALESCE(tercero_patron,''),COALESCE(cuenta_patron,''),COALESCE(porcentaje,100),COALESCE(prioridad,100),COALESCE(activa,1),COALESCE(estado,'activo'),COALESCE(observaciones,''),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,''),COALESCE(usuario_creador,'') FROM empresa_centros_costo_reglas WHERE %s ORDER BY activa DESC,prioridad ASC,id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCentroCostoRegla{}
	for rows.Next() {
		var x EmpresaCentroCostoRegla
		var activa int
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.CentroCostoID, &x.CentroCostoCodigo, &x.Nombre, &x.OrigenModulo, &x.Categoria, &x.TerceroPatron, &x.CuentaPatron, &x.Porcentaje, &x.Prioridad, &activa, &x.Estado, &x.Observaciones, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		x.Activa = activa > 0
		out = append(out, x)
	}
	return out, rows.Err()
}

func UpsertEmpresaCentroCostoPresupuesto(dbConn *sql.DB, item EmpresaCentroCostoPresupuesto) (int64, error) {
	if err := EnsureEmpresaCentrosCostoSchema(dbConn); err != nil {
		return 0, err
	}
	item = normalizeEmpresaCentroCostoPresupuesto(item)
	if item.EmpresaID <= 0 || item.CentroCostoCodigo == "" || item.Periodo == "" {
		return 0, errors.New("empresa_id, centro_costo_codigo y periodo son obligatorios")
	}
	if item.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_centros_costo_presupuestos SET centro_costo_id=?,centro_costo_codigo=?,periodo=?,escenario=?,ingresos_presupuesto=?,egresos_presupuesto=?,meta_margen_pct=?,responsable=?,estado=?,observaciones=?,fecha_actualizacion=CAST(CURRENT_TIMESTAMP AS TEXT) WHERE empresa_id=? AND id=?`,
			item.CentroCostoID, item.CentroCostoCodigo, item.Periodo, item.Escenario, item.IngresosPresupuesto, item.EgresosPresupuesto, item.MetaMargenPct, item.Responsable, item.Estado, item.Observaciones, item.EmpresaID, item.ID)
		return item.ID, err
	}
	var id int64
	err := QueryRowCompat(dbConn, `INSERT INTO empresa_centros_costo_presupuestos (empresa_id,centro_costo_id,centro_costo_codigo,periodo,escenario,ingresos_presupuesto,egresos_presupuesto,meta_margen_pct,responsable,estado,observaciones,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (empresa_id,centro_costo_codigo,periodo,escenario) DO UPDATE SET centro_costo_id=EXCLUDED.centro_costo_id,ingresos_presupuesto=EXCLUDED.ingresos_presupuesto,egresos_presupuesto=EXCLUDED.egresos_presupuesto,meta_margen_pct=EXCLUDED.meta_margen_pct,responsable=EXCLUDED.responsable,estado=EXCLUDED.estado,observaciones=EXCLUDED.observaciones,fecha_actualizacion=CAST(CURRENT_TIMESTAMP AS TEXT),usuario_creador=EXCLUDED.usuario_creador
		RETURNING id`,
		item.EmpresaID, item.CentroCostoID, item.CentroCostoCodigo, item.Periodo, item.Escenario, item.IngresosPresupuesto, item.EgresosPresupuesto, item.MetaMargenPct, item.Responsable, item.Estado, item.Observaciones, item.UsuarioCreador).Scan(&id)
	return id, err
}

func ListEmpresaCentroCostoPresupuestos(dbConn *sql.DB, empresaID int64, periodo, escenario string, limit int) ([]EmpresaCentroCostoPresupuesto, error) {
	if err := EnsureEmpresaCentrosCostoSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 1000 {
		limit = 300
	}
	args := []interface{}{empresaID}
	where := "p.empresa_id=?"
	if strings.TrimSpace(periodo) != "" {
		where += " AND p.periodo=?"
		args = append(args, strings.TrimSpace(periodo))
	}
	if strings.TrimSpace(escenario) != "" {
		where += " AND p.escenario=?"
		args = append(args, normalizeCentroCostoText(escenario, "base"))
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT p.id,p.empresa_id,COALESCE(p.centro_costo_id,0),COALESCE(p.centro_costo_codigo,''),COALESCE(c.nombre,''),COALESCE(p.periodo,''),COALESCE(p.escenario,'base'),COALESCE(p.ingresos_presupuesto,0),COALESCE(p.egresos_presupuesto,0),COALESCE(p.meta_margen_pct,0),COALESCE(p.responsable,''),COALESCE(p.estado,'aprobado'),COALESCE(p.observaciones,''),COALESCE(p.fecha_creacion,''),COALESCE(p.fecha_actualizacion,''),COALESCE(p.usuario_creador,'') FROM empresa_centros_costo_presupuestos p LEFT JOIN empresa_centros_costo c ON c.empresa_id=p.empresa_id AND c.codigo=p.centro_costo_codigo WHERE %s ORDER BY p.periodo DESC,p.centro_costo_codigo LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCentroCostoPresupuesto{}
	for rows.Next() {
		var x EmpresaCentroCostoPresupuesto
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.CentroCostoID, &x.CentroCostoCodigo, &x.CentroCostoNombre, &x.Periodo, &x.Escenario, &x.IngresosPresupuesto, &x.EgresosPresupuesto, &x.MetaMargenPct, &x.Responsable, &x.Estado, &x.Observaciones, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func BuildEmpresaCentrosCostoDashboard(dbConn *sql.DB, empresaID int64, periodo string) (EmpresaCentrosCostoDashboard, error) {
	if err := EnsureEmpresaCentrosCostoSchema(dbConn); err != nil {
		return EmpresaCentrosCostoDashboard{}, err
	}
	periodo = normalizeCentroCostoPeriodo(periodo)
	centros, err := ListEmpresaCentrosCosto(dbConn, empresaID, "", 500)
	if err != nil {
		return EmpresaCentrosCostoDashboard{}, err
	}
	reglas, err := ListEmpresaCentroCostoReglas(dbConn, empresaID, "", 500)
	if err != nil {
		return EmpresaCentrosCostoDashboard{}, err
	}
	pres, err := ListEmpresaCentroCostoPresupuestos(dbConn, empresaID, periodo, "base", 500)
	if err != nil {
		return EmpresaCentrosCostoDashboard{}, err
	}
	movs, err := ListEmpresaCentroCostoMovimientos(dbConn, empresaID, periodo, 500)
	if err != nil {
		return EmpresaCentrosCostoDashboard{}, err
	}
	return buildEmpresaCentrosCostoDashboardFromRows(empresaID, periodo, centros, reglas, pres, movs), nil
}

func ListEmpresaCentroCostoMovimientos(dbConn *sql.DB, empresaID int64, periodo string, limit int) ([]EmpresaCentroCostoMovimiento, error) {
	if limit <= 0 || limit > 1000 {
		limit = 500
	}
	periodo = strings.TrimSpace(periodo)
	out := []EmpresaCentroCostoMovimiento{}
	appenders := []func(*sql.DB, int64, string, int) ([]EmpresaCentroCostoMovimiento, error){
		listCentroCostoMovimientosTesoreria,
		listCentroCostoMovimientosContabilidad,
		listCentroCostoMovimientosSoportes,
		listCentroCostoMovimientosCompras,
		listCentroCostoMovimientosAIU,
	}
	perSourceLimit := limit
	if perSourceLimit > 160 {
		perSourceLimit = 160
	}
	for _, fn := range appenders {
		rows, err := fn(dbConn, empresaID, periodo, perSourceLimit)
		if err != nil {
			if isOptionalCentroCostoSourceError(err) {
				continue
			}
			return out, err
		}
		out = append(out, rows...)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Fecha > out[j].Fecha
	})
	if len(out) > limit {
		out = out[:limit]
	}
	for i := range out {
		out[i] = normalizeEmpresaCentroCostoMovimiento(out[i])
	}
	return out, nil
}

func SeedEmpresaCentrosCostoDemo(dbConn *sql.DB, empresaID int64, user string) error {
	periodo := currentPeriodCentrosCosto()
	centros := []EmpresaCentroCosto{
		{EmpresaID: empresaID, Codigo: "ADMIN", Nombre: "Administracion general", Tipo: "area", Nivel: 1, Responsable: "Gerencia", Sucursal: "Principal", Area: "Administracion", UnidadNegocio: "Corporativo", MetaMargenPct: 18, Estado: "activo", UsuarioCreador: user},
		{EmpresaID: empresaID, Codigo: "OPERACIONES", Nombre: "Operacion principal", Tipo: "area", Nivel: 1, Responsable: "Jefe operativo", Sucursal: "Principal", Area: "Operacion", UnidadNegocio: "Servicios", MetaMargenPct: 28, Estado: "activo", UsuarioCreador: user},
		{EmpresaID: empresaID, Codigo: "VENTAS", Nombre: "Ventas y canales digitales", Tipo: "area", Nivel: 1, Responsable: "Coordinacion comercial", Sucursal: "Principal", Area: "Ventas", UnidadNegocio: "Comercial", MetaMargenPct: 32, Estado: "activo", UsuarioCreador: user},
		{EmpresaID: empresaID, Codigo: "MOTEL-CALIPSO", Nombre: "Motel Calipso", Tipo: "sucursal", Nivel: 1, Responsable: "Administrador Calipso", Sucursal: "Motel Calipso", Area: "Operacion hotelera", UnidadNegocio: "Alojamiento", MetaMargenPct: 35, Estado: "activo", UsuarioCreador: user},
		{EmpresaID: empresaID, Codigo: "OBRAS-DEMO", Nombre: "Contratos AIU y obras", Tipo: "proyecto", Nivel: 1, Responsable: "Residente de obra", Sucursal: "Principal", Area: "Construccion", UnidadNegocio: "AIU", MetaMargenPct: 22, Estado: "activo", UsuarioCreador: user},
	}
	ids := map[string]int64{}
	for _, c := range centros {
		id, err := UpsertEmpresaCentroCosto(dbConn, c)
		if err != nil {
			return err
		}
		ids[c.Codigo] = id
	}
	pres := []EmpresaCentroCostoPresupuesto{
		{EmpresaID: empresaID, CentroCostoID: ids["ADMIN"], CentroCostoCodigo: "ADMIN", Periodo: periodo, Escenario: "base", IngresosPresupuesto: 0, EgresosPresupuesto: 9500000, MetaMargenPct: 0, Responsable: "Gerencia", Estado: "aprobado", UsuarioCreador: user},
		{EmpresaID: empresaID, CentroCostoID: ids["OPERACIONES"], CentroCostoCodigo: "OPERACIONES", Periodo: periodo, Escenario: "base", IngresosPresupuesto: 28000000, EgresosPresupuesto: 17500000, MetaMargenPct: 28, Responsable: "Operacion", Estado: "aprobado", UsuarioCreador: user},
		{EmpresaID: empresaID, CentroCostoID: ids["VENTAS"], CentroCostoCodigo: "VENTAS", Periodo: periodo, Escenario: "base", IngresosPresupuesto: 18000000, EgresosPresupuesto: 5200000, MetaMargenPct: 32, Responsable: "Comercial", Estado: "aprobado", UsuarioCreador: user},
		{EmpresaID: empresaID, CentroCostoID: ids["MOTEL-CALIPSO"], CentroCostoCodigo: "MOTEL-CALIPSO", Periodo: periodo, Escenario: "base", IngresosPresupuesto: 42000000, EgresosPresupuesto: 24000000, MetaMargenPct: 35, Responsable: "Administrador Calipso", Estado: "aprobado", UsuarioCreador: user},
		{EmpresaID: empresaID, CentroCostoID: ids["OBRAS-DEMO"], CentroCostoCodigo: "OBRAS-DEMO", Periodo: periodo, Escenario: "base", IngresosPresupuesto: 65000000, EgresosPresupuesto: 50000000, MetaMargenPct: 22, Responsable: "Residente de obra", Estado: "aprobado", UsuarioCreador: user},
	}
	for _, p := range pres {
		if _, err := UpsertEmpresaCentroCostoPresupuesto(dbConn, p); err != nil {
			return err
		}
	}
	reglas := []EmpresaCentroCostoRegla{
		{EmpresaID: empresaID, CentroCostoID: ids["MOTEL-CALIPSO"], CentroCostoCodigo: "MOTEL-CALIPSO", Nombre: "Compras hoteleras Calipso", OrigenModulo: "compras", Categoria: "operacion_hotelera", TerceroPatron: "proveedor hotelero", Porcentaje: 100, Prioridad: 10, Activa: true, Estado: "activo", UsuarioCreador: user},
		{EmpresaID: empresaID, CentroCostoID: ids["OBRAS-DEMO"], CentroCostoCodigo: "OBRAS-DEMO", Nombre: "Contratos AIU de obra", OrigenModulo: "aiu_construccion", Categoria: "contratos_obra", Porcentaje: 100, Prioridad: 20, Activa: true, Estado: "activo", UsuarioCreador: user},
		{EmpresaID: empresaID, CentroCostoID: ids["VENTAS"], CentroCostoCodigo: "VENTAS", Nombre: "Ingresos comerciales", OrigenModulo: "contabilidad", CuentaPatron: "4135", Porcentaje: 100, Prioridad: 30, Activa: true, Estado: "activo", UsuarioCreador: user},
	}
	for _, r := range reglas {
		if _, err := UpsertEmpresaCentroCostoRegla(dbConn, r); err != nil {
			return err
		}
	}
	return nil
}

func buildEmpresaCentrosCostoDashboardFromRows(empresaID int64, periodo string, centros []EmpresaCentroCosto, reglas []EmpresaCentroCostoRegla, presupuestos []EmpresaCentroCostoPresupuesto, movimientos []EmpresaCentroCostoMovimiento) EmpresaCentrosCostoDashboard {
	periodo = normalizeCentroCostoPeriodo(periodo)
	ds := EmpresaCentrosCostoDashboard{
		EmpresaID:            empresaID,
		Periodo:              periodo,
		Centros:              centros,
		Reglas:               reglas,
		Presupuestos:         presupuestos,
		MovimientosRecientes: movimientos,
	}
	rows := map[string]*EmpresaCentroCostoRentabilidad{}
	for _, c := range centros {
		c = normalizeEmpresaCentroCosto(c)
		if c.Estado == "activo" {
			ds.CentrosActivos++
		}
		rows[c.Codigo] = &EmpresaCentroCostoRentabilidad{CentroCostoCodigo: c.Codigo, CentroCostoNombre: c.Nombre, Tipo: c.Tipo, Sucursal: c.Sucursal, Area: c.Area, UnidadNegocio: c.UnidadNegocio, MetaMargenPct: c.MetaMargenPct, InferidoDeMovimientos: c.InferidoDeMovimientos}
	}
	for _, r := range reglas {
		if r.Activa && r.Estado == "activo" {
			ds.ReglasActivas++
		}
	}
	for _, p := range presupuestos {
		p = normalizeEmpresaCentroCostoPresupuesto(p)
		row := ensureCentroCostoRentabilidadRow(rows, p.CentroCostoCodigo, p.CentroCostoNombre)
		row.PresupuestoIngresos += p.IngresosPresupuesto
		row.PresupuestoEgresos += p.EgresosPresupuesto
		if p.MetaMargenPct > 0 {
			row.MetaMargenPct = p.MetaMargenPct
		}
		ds.PresupuestoIngresos += p.IngresosPresupuesto
		ds.PresupuestoEgresos += p.EgresosPresupuesto
	}
	for _, mov := range movimientos {
		mov = normalizeEmpresaCentroCostoMovimiento(mov)
		if mov.CentroCostoCodigo == "" {
			continue
		}
		row := ensureCentroCostoRentabilidadRow(rows, mov.CentroCostoCodigo, mov.CentroCostoNombre)
		row.Ingresos += mov.Ingresos
		row.Egresos += mov.Egresos
		row.Movimientos++
		ds.IngresosTotal += mov.Ingresos
		ds.EgresosTotal += mov.Egresos
		ds.MovimientosTotal++
	}
	ds.IngresosTotal = roundMoneyCentroCosto(ds.IngresosTotal)
	ds.EgresosTotal = roundMoneyCentroCosto(ds.EgresosTotal)
	ds.MargenTotal = roundMoneyCentroCosto(ds.IngresosTotal - ds.EgresosTotal)
	ds.MargenPct = percentCentroCosto(ds.MargenTotal, ds.IngresosTotal)
	ds.EjecucionIngresosPct = percentCentroCosto(ds.IngresosTotal, ds.PresupuestoIngresos)
	ds.EjecucionEgresosPct = percentCentroCosto(ds.EgresosTotal, ds.PresupuestoEgresos)
	for _, row := range rows {
		row.Ingresos = roundMoneyCentroCosto(row.Ingresos)
		row.Egresos = roundMoneyCentroCosto(row.Egresos)
		row.PresupuestoIngresos = roundMoneyCentroCosto(row.PresupuestoIngresos)
		row.PresupuestoEgresos = roundMoneyCentroCosto(row.PresupuestoEgresos)
		row.Margen = roundMoneyCentroCosto(row.Ingresos - row.Egresos)
		row.MargenPct = percentCentroCosto(row.Margen, row.Ingresos)
		row.EjecucionIngresosPct = percentCentroCosto(row.Ingresos, row.PresupuestoIngresos)
		row.EjecucionEgresosPct = percentCentroCosto(row.Egresos, row.PresupuestoEgresos)
		if row.MetaMargenPct > 0 && row.Ingresos > 0 && row.MargenPct < row.MetaMargenPct {
			row.Alerta = fmt.Sprintf("Margen %.1f%% por debajo de meta %.1f%%", row.MargenPct, row.MetaMargenPct)
			ds.Alertas = append(ds.Alertas, row.CentroCostoCodigo+": "+row.Alerta)
		}
		if row.PresupuestoEgresos > 0 && row.EjecucionEgresosPct > 105 {
			ds.Alertas = append(ds.Alertas, row.CentroCostoCodigo+": egresos superan 105% del presupuesto")
		}
		ds.Rentabilidad = append(ds.Rentabilidad, *row)
	}
	sort.SliceStable(ds.Rentabilidad, func(i, j int) bool {
		if ds.Rentabilidad[i].Margen == ds.Rentabilidad[j].Margen {
			return ds.Rentabilidad[i].CentroCostoCodigo < ds.Rentabilidad[j].CentroCostoCodigo
		}
		return ds.Rentabilidad[i].Margen > ds.Rentabilidad[j].Margen
	})
	if len(ds.MovimientosRecientes) > 120 {
		ds.MovimientosRecientes = ds.MovimientosRecientes[:120]
	}
	if len(ds.Alertas) == 0 {
		ds.Alertas = append(ds.Alertas, "Centros de costo sin alertas criticas para el periodo "+periodo+".")
	}
	return ds
}

func ensureCentroCostoRentabilidadRow(rows map[string]*EmpresaCentroCostoRentabilidad, codigo, nombre string) *EmpresaCentroCostoRentabilidad {
	codigo = normalizeCentroCostoCodigo(codigo)
	if codigo == "" {
		codigo = "SIN-CENTRO"
	}
	if row, ok := rows[codigo]; ok {
		if strings.TrimSpace(row.CentroCostoNombre) == "" && strings.TrimSpace(nombre) != "" {
			row.CentroCostoNombre = strings.TrimSpace(nombre)
		}
		return row
	}
	nombre = strings.TrimSpace(nombre)
	if nombre == "" {
		nombre = strings.ReplaceAll(strings.Title(strings.ToLower(strings.ReplaceAll(codigo, "-", " "))), "  ", " ")
	}
	rows[codigo] = &EmpresaCentroCostoRentabilidad{CentroCostoCodigo: codigo, CentroCostoNombre: nombre, InferidoDeMovimientos: true}
	return rows[codigo]
}

func listCentroCostoMovimientosTesoreria(dbConn *sql.DB, empresaID int64, periodo string, limit int) ([]EmpresaCentroCostoMovimiento, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT COALESCE(p.centro_costo,''),COALESCE(pr.nombre,''),COALESCE(pr.periodo,''),COALESCE(pr.codigo,''),COALESCE(p.tipo,'egreso'),COALESCE(p.categoria,''),COALESCE(p.concepto,''),COALESCE(p.valor_ejecutado,0),COALESCE(p.valor_presupuestado,0),COALESCE(p.estado,'activo'),COALESCE(p.fecha_creacion,'') FROM empresa_tesoreria_partidas p JOIN empresa_tesoreria_presupuestos pr ON pr.empresa_id=p.empresa_id AND pr.id=p.presupuesto_id WHERE p.empresa_id=? AND TRIM(COALESCE(p.centro_costo,''))<>'' AND (?='' OR pr.periodo=?) ORDER BY p.id DESC LIMIT ?`, empresaID, periodo, periodo, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCentroCostoMovimiento{}
	for rows.Next() {
		var codigo, nombre, per, ref, tipo, cat, concepto, estado, fecha string
		var ejecutado, presupuestado float64
		if err := rows.Scan(&codigo, &nombre, &per, &ref, &tipo, &cat, &concepto, &ejecutado, &presupuestado, &estado, &fecha); err != nil {
			return nil, err
		}
		valor := ejecutado
		if valor == 0 {
			valor = presupuestado
		}
		out = append(out, movimientoCentroCostoDesdeValor(codigo, nombre, "tesoreria_presupuesto", ref, fecha, per, tipo, cat, concepto, valor, estado))
	}
	return out, rows.Err()
}

func listCentroCostoMovimientosContabilidad(dbConn *sql.DB, empresaID int64, periodo string, limit int) ([]EmpresaCentroCostoMovimiento, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT COALESCE(l.centro_costo,''),COALESCE(c.fecha_comprobante,''),COALESCE(c.periodo_contable,''),COALESCE(c.codigo,''),COALESCE(c.concepto,''),COALESCE(l.detalle,''),COALESCE(l.debito,0),COALESCE(l.credito,0),COALESCE(c.estado,'') FROM empresa_contabilidad_colombia_lineas l JOIN empresa_contabilidad_colombia_comprobantes c ON c.empresa_id=l.empresa_id AND c.id=l.comprobante_id WHERE l.empresa_id=? AND TRIM(COALESCE(l.centro_costo,''))<>'' AND (?='' OR c.periodo_contable=?) ORDER BY c.fecha_comprobante DESC,l.id DESC LIMIT ?`, empresaID, periodo, periodo, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCentroCostoMovimiento{}
	for rows.Next() {
		var codigo, fecha, per, ref, concepto, detalle, estado string
		var debito, credito float64
		if err := rows.Scan(&codigo, &fecha, &per, &ref, &concepto, &detalle, &debito, &credito, &estado); err != nil {
			return nil, err
		}
		if strings.TrimSpace(concepto) == "" {
			concepto = detalle
		}
		out = append(out, EmpresaCentroCostoMovimiento{CentroCostoCodigo: codigo, OrigenModulo: "contabilidad_colombia", Referencia: ref, Fecha: fecha, Periodo: per, Tipo: "contable", Categoria: "asiento", Concepto: concepto, Ingresos: roundMoneyCentroCosto(credito), Egresos: roundMoneyCentroCosto(debito), Valor: roundMoneyCentroCosto(credito - debito), Estado: estado})
	}
	return out, rows.Err()
}

func listCentroCostoMovimientosSoportes(dbConn *sql.DB, empresaID int64, periodo string, limit int) ([]EmpresaCentroCostoMovimiento, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT COALESCE(centro_costo,''),COALESCE(codigo,''),COALESCE(fecha_documento,''),COALESCE(categoria_contable,''),COALESCE(proveedor_nombre,''),COALESCE(total,0),COALESCE(estado_soporte,'') FROM empresa_soportes_compras_ia WHERE empresa_id=? AND TRIM(COALESCE(centro_costo,''))<>'' AND (?='' OR SUBSTRING(COALESCE(fecha_documento,CAST(fecha_creacion AS TEXT),'') FROM 1 FOR 7)=?) ORDER BY fecha_creacion DESC,id DESC LIMIT ?`, empresaID, periodo, periodo, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCentroCostoMovimiento{}
	for rows.Next() {
		var codigo, ref, fecha, cat, proveedor, estado string
		var total float64
		if err := rows.Scan(&codigo, &ref, &fecha, &cat, &proveedor, &total, &estado); err != nil {
			return nil, err
		}
		out = append(out, movimientoCentroCostoDesdeValor(codigo, "", "soportes_compras_ia", ref, fecha, normalizeCentroCostoPeriodoFromFecha(fecha), "egreso", cat, "Compra/gasto "+proveedor, total, estado))
	}
	return out, rows.Err()
}

func listCentroCostoMovimientosCompras(dbConn *sql.DB, empresaID int64, periodo string, limit int) ([]EmpresaCentroCostoMovimiento, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT COALESCE(centro_costo,''),COALESCE(area,''),COALESCE(codigo,''),COALESCE(fecha_solicitud,''),COALESCE(total_estimado,0),COALESCE(estado_flujo,''),COALESCE(justificacion,'') FROM empresa_compras_requisiciones WHERE empresa_id=? AND TRIM(COALESCE(centro_costo,''))<>'' AND estado_flujo NOT IN ('rechazada','cancelada') AND (?='' OR SUBSTRING(COALESCE(fecha_solicitud,CAST(fecha_creacion AS TEXT),'') FROM 1 FOR 7)=?) ORDER BY id DESC LIMIT ?`, empresaID, periodo, periodo, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCentroCostoMovimiento{}
	for rows.Next() {
		var codigo, area, ref, fecha, estado, concepto string
		var total float64
		if err := rows.Scan(&codigo, &area, &ref, &fecha, &total, &estado, &concepto); err != nil {
			return nil, err
		}
		out = append(out, movimientoCentroCostoDesdeValor(codigo, "", "compras_avanzadas", ref, fecha, normalizeCentroCostoPeriodoFromFecha(fecha), "egreso", area, concepto, total, estado))
	}
	return out, rows.Err()
}

func listCentroCostoMovimientosAIU(dbConn *sql.DB, empresaID int64, periodo string, limit int) ([]EmpresaCentroCostoMovimiento, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT COALESCE(c.centro_costo,''),COALESCE(c.nombre,''),COALESCE(f.documento_codigo,''),COALESCE(f.fecha_documento,''),COALESCE(f.periodo_contable,''),COALESCE(f.costo_directo,0),COALESCE(f.neto_cobrar,0),COALESCE(f.total_factura,0),COALESCE(f.estado,'') FROM empresa_aiu_facturas f JOIN empresa_aiu_contratos c ON c.empresa_id=f.empresa_id AND c.id=f.contrato_id WHERE f.empresa_id=? AND TRIM(COALESCE(c.centro_costo,''))<>'' AND (?='' OR f.periodo_contable=?) ORDER BY f.id DESC LIMIT ?`, empresaID, periodo, periodo, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCentroCostoMovimiento{}
	for rows.Next() {
		var codigo, nombre, ref, fecha, per, estado string
		var costo, neto, total float64
		if err := rows.Scan(&codigo, &nombre, &ref, &fecha, &per, &costo, &neto, &total, &estado); err != nil {
			return nil, err
		}
		ingreso := neto
		if ingreso == 0 {
			ingreso = total
		}
		out = append(out, EmpresaCentroCostoMovimiento{CentroCostoCodigo: codigo, CentroCostoNombre: nombre, OrigenModulo: "aiu_construccion", Referencia: ref, Fecha: fecha, Periodo: per, Tipo: "ingreso", Categoria: "contrato_obra", Concepto: "Factura AIU " + nombre, Ingresos: roundMoneyCentroCosto(ingreso), Egresos: roundMoneyCentroCosto(costo), Valor: roundMoneyCentroCosto(ingreso - costo), Estado: estado})
	}
	return out, rows.Err()
}

func movimientoCentroCostoDesdeValor(codigo, nombre, origen, ref, fecha, periodo, tipo, categoria, concepto string, valor float64, estado string) EmpresaCentroCostoMovimiento {
	tipo = normalizeCentroCostoText(tipo, "egreso")
	valor = math.Max(0, valor)
	mov := EmpresaCentroCostoMovimiento{CentroCostoCodigo: codigo, CentroCostoNombre: nombre, OrigenModulo: origen, Referencia: ref, Fecha: fecha, Periodo: periodo, Tipo: tipo, Categoria: categoria, Concepto: concepto, Valor: roundMoneyCentroCosto(valor), Estado: estado}
	if tipo == "ingreso" {
		mov.Ingresos = mov.ValueAbs()
	} else {
		mov.Egresos = mov.ValueAbs()
		mov.Valor = -mov.ValueAbs()
	}
	return mov
}

func (m EmpresaCentroCostoMovimiento) ValueAbs() float64 {
	return roundMoneyCentroCosto(math.Abs(m.Valor))
}

func normalizeEmpresaCentroCosto(x EmpresaCentroCosto) EmpresaCentroCosto {
	x.Codigo = normalizeCentroCostoCodigo(x.Codigo)
	x.Nombre = strings.TrimSpace(x.Nombre)
	x.Tipo = normalizeOneOfCentroCosto(x.Tipo, "operativo", "operativo", "sucursal", "area", "unidad_negocio", "proyecto", "cliente", "canal", "administrativo")
	if x.Nivel <= 0 {
		x.Nivel = 1
	}
	x.PadreCodigo = normalizeCentroCostoCodigo(x.PadreCodigo)
	x.Responsable = strings.TrimSpace(x.Responsable)
	x.Sucursal = strings.TrimSpace(x.Sucursal)
	x.Area = strings.TrimSpace(x.Area)
	x.UnidadNegocio = strings.TrimSpace(x.UnidadNegocio)
	x.MetaMargenPct = clampCentroCostoPct(x.MetaMargenPct)
	x.Estado = normalizeCentroCostoEstado(x.Estado)
	x.FechaInicio = strings.TrimSpace(x.FechaInicio)
	x.FechaFin = strings.TrimSpace(x.FechaFin)
	x.Observaciones = strings.TrimSpace(x.Observaciones)
	x.UsuarioCreador = strings.TrimSpace(x.UsuarioCreador)
	return x
}

func normalizeEmpresaCentroCostoRegla(x EmpresaCentroCostoRegla) EmpresaCentroCostoRegla {
	x.CentroCostoCodigo = normalizeCentroCostoCodigo(x.CentroCostoCodigo)
	x.Nombre = strings.TrimSpace(x.Nombre)
	x.OrigenModulo = normalizeCentroCostoText(x.OrigenModulo, "general")
	x.Categoria = normalizeCentroCostoText(x.Categoria, "")
	x.TerceroPatron = strings.TrimSpace(x.TerceroPatron)
	x.CuentaPatron = strings.TrimSpace(x.CuentaPatron)
	x.Porcentaje = clampCentroCostoPct(x.Porcentaje)
	if x.Porcentaje == 0 {
		x.Porcentaje = 100
	}
	if x.Prioridad <= 0 {
		x.Prioridad = 100
	}
	x.Estado = normalizeCentroCostoEstado(x.Estado)
	x.Observaciones = strings.TrimSpace(x.Observaciones)
	x.UsuarioCreador = strings.TrimSpace(x.UsuarioCreador)
	return x
}

func normalizeEmpresaCentroCostoPresupuesto(x EmpresaCentroCostoPresupuesto) EmpresaCentroCostoPresupuesto {
	x.CentroCostoCodigo = normalizeCentroCostoCodigo(x.CentroCostoCodigo)
	x.Periodo = normalizeCentroCostoPeriodo(x.Periodo)
	x.Escenario = normalizeCentroCostoText(x.Escenario, "base")
	x.IngresosPresupuesto = roundMoneyCentroCosto(math.Max(0, x.IngresosPresupuesto))
	x.EgresosPresupuesto = roundMoneyCentroCosto(math.Max(0, x.EgresosPresupuesto))
	x.MetaMargenPct = clampCentroCostoPct(x.MetaMargenPct)
	x.Responsable = strings.TrimSpace(x.Responsable)
	x.Estado = normalizeOneOfCentroCosto(x.Estado, "aprobado", "borrador", "aprobado", "activo", "cerrado", "anulado")
	x.Observaciones = strings.TrimSpace(x.Observaciones)
	x.UsuarioCreador = strings.TrimSpace(x.UsuarioCreador)
	return x
}

func normalizeEmpresaCentroCostoMovimiento(x EmpresaCentroCostoMovimiento) EmpresaCentroCostoMovimiento {
	x.CentroCostoCodigo = normalizeCentroCostoCodigo(x.CentroCostoCodigo)
	x.CentroCostoNombre = strings.TrimSpace(x.CentroCostoNombre)
	x.OrigenModulo = normalizeCentroCostoText(x.OrigenModulo, "general")
	x.Referencia = strings.TrimSpace(x.Referencia)
	x.Fecha = strings.TrimSpace(x.Fecha)
	x.Periodo = normalizeCentroCostoPeriodo(x.Periodo)
	if x.Periodo == "" {
		x.Periodo = normalizeCentroCostoPeriodoFromFecha(x.Fecha)
	}
	x.Tipo = normalizeCentroCostoText(x.Tipo, "egreso")
	x.Categoria = normalizeCentroCostoText(x.Categoria, "")
	x.Concepto = strings.TrimSpace(x.Concepto)
	x.Ingresos = roundMoneyCentroCosto(math.Max(0, x.Ingresos))
	x.Egresos = roundMoneyCentroCosto(math.Max(0, x.Egresos))
	if x.Valor == 0 {
		x.Valor = x.Ingresos - x.Egresos
	}
	x.Valor = roundMoneyCentroCosto(x.Valor)
	x.Estado = normalizeCentroCostoText(x.Estado, "")
	return x
}

func normalizeCentroCostoCodigo(v string) string {
	v = strings.ToUpper(strings.TrimSpace(v))
	replacer := strings.NewReplacer("_", "-", " ", "-", "/", "-", "\\", "-", ".", "-", ":", "-", "--", "-")
	v = replacer.Replace(v)
	for strings.Contains(v, "--") {
		v = strings.ReplaceAll(v, "--", "-")
	}
	return strings.Trim(v, "-")
}

func normalizeCentroCostoEstado(v string) string {
	return normalizeOneOfCentroCosto(v, "activo", "activo", "inactivo", "cerrado", "suspendido")
}

func normalizeCentroCostoText(v, fallback string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	v = strings.ReplaceAll(v, " ", "_")
	if v == "" {
		return fallback
	}
	return v
}

func normalizeOneOfCentroCosto(v, fallback string, allowed ...string) string {
	v = normalizeCentroCostoText(v, fallback)
	for _, item := range allowed {
		if v == item {
			return v
		}
	}
	return fallback
}

func normalizeCentroCostoPeriodo(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 7 {
		return v[:7]
	}
	return v
}

func normalizeCentroCostoPeriodoFromFecha(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 7 {
		return v[:7]
	}
	return ""
}

func currentPeriodCentrosCosto() string {
	return time.Now().Format("2006-01")
}

func clampCentroCostoPct(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return math.Round(v*100) / 100
}

func roundMoneyCentroCosto(v float64) float64 {
	return math.Round(v*100) / 100
}

func percentCentroCosto(value, base float64) float64 {
	if math.Abs(base) < 0.0001 {
		return 0
	}
	return math.Round((value/base)*10000) / 100
}

func boolIntCentroCosto(v bool) int {
	if v {
		return 1
	}
	return 0
}

func isOptionalCentroCostoSourceError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "does not exist") || strings.Contains(msg, "no such table") || strings.Contains(msg, "undefined_table") || strings.Contains(msg, "undefined column")
}
