package db

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
)

type EmpresaTesoreriaConfig struct {
	EmpresaID              int64  `json:"empresa_id"`
	NombreSistema          string `json:"nombre_sistema"`
	Moneda                 string `json:"moneda"`
	PeriodoTrabajo         string `json:"periodo_trabajo"`
	MetodoProyeccion       string `json:"metodo_proyeccion"`
	AlertaSaldoMinimo      bool   `json:"alerta_saldo_minimo"`
	RequiereAprobacionPago bool   `json:"requiere_aprobacion_pago"`
	FechaActualizacion     string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador         string `json:"usuario_creador,omitempty"`
}

type EmpresaTesoreriaCuenta struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	Codigo             string  `json:"codigo"`
	Nombre             string  `json:"nombre"`
	Tipo               string  `json:"tipo"`
	Entidad            string  `json:"entidad,omitempty"`
	Numero             string  `json:"numero,omitempty"`
	Moneda             string  `json:"moneda"`
	SaldoInicial       float64 `json:"saldo_inicial"`
	SaldoActual        float64 `json:"saldo_actual"`
	SaldoMinimo        float64 `json:"saldo_minimo"`
	Responsable        string  `json:"responsable,omitempty"`
	Estado             string  `json:"estado"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
}

type EmpresaTesoreriaPresupuesto struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	Codigo             string  `json:"codigo"`
	Nombre             string  `json:"nombre"`
	Periodo            string  `json:"periodo"`
	Escenario          string  `json:"escenario"`
	IngresosMeta       float64 `json:"ingresos_meta"`
	EgresosMeta        float64 `json:"egresos_meta"`
	SaldoInicial       float64 `json:"saldo_inicial"`
	Estado             string  `json:"estado"`
	Responsable        string  `json:"responsable,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
}

type EmpresaTesoreriaPartida struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	PresupuestoID      int64   `json:"presupuesto_id"`
	Categoria          string  `json:"categoria"`
	Tipo               string  `json:"tipo"`
	Concepto           string  `json:"concepto"`
	ValorPresupuestado float64 `json:"valor_presupuestado"`
	ValorEjecutado     float64 `json:"valor_ejecutado"`
	Periodicidad       string  `json:"periodicidad"`
	CentroCosto        string  `json:"centro_costo,omitempty"`
	Estado             string  `json:"estado"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
}

type EmpresaTesoreriaFlujo struct {
	ID             int64   `json:"id"`
	EmpresaID      int64   `json:"empresa_id"`
	CuentaID       int64   `json:"cuenta_id,omitempty"`
	CuentaNombre   string  `json:"cuenta_nombre,omitempty"`
	PresupuestoID  int64   `json:"presupuesto_id,omitempty"`
	FechaFlujo     string  `json:"fecha_flujo"`
	Periodo        string  `json:"periodo"`
	Tipo           string  `json:"tipo"`
	Categoria      string  `json:"categoria"`
	Concepto       string  `json:"concepto"`
	Valor          float64 `json:"valor"`
	ValorEjecutado float64 `json:"valor_ejecutado"`
	OrigenModulo   string  `json:"origen_modulo"`
	Referencia     string  `json:"referencia,omitempty"`
	Estado         string  `json:"estado"`
	UsuarioCreador string  `json:"usuario_creador,omitempty"`
	FechaCreacion  string  `json:"fecha_creacion,omitempty"`
}

type EmpresaTesoreriaDashboard struct {
	EmpresaID           int64                         `json:"empresa_id"`
	CuentasActivas      int                           `json:"cuentas_activas"`
	PresupuestosActivos int                           `json:"presupuestos_activos"`
	SaldoDisponible     float64                       `json:"saldo_disponible"`
	IngresosProyectados float64                       `json:"ingresos_proyectados"`
	EgresosProyectados  float64                       `json:"egresos_proyectados"`
	FlujoNeto           float64                       `json:"flujo_neto"`
	EjecucionIngresos   float64                       `json:"ejecucion_ingresos"`
	EjecucionEgresos    float64                       `json:"ejecucion_egresos"`
	Config              EmpresaTesoreriaConfig        `json:"config"`
	Cuentas             []EmpresaTesoreriaCuenta      `json:"cuentas"`
	Presupuestos        []EmpresaTesoreriaPresupuesto `json:"presupuestos"`
	Partidas            []EmpresaTesoreriaPartida     `json:"partidas"`
	Flujo               []EmpresaTesoreriaFlujo       `json:"flujo"`
}

func EnsureEmpresaTesoreriaPresupuestoSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_tesoreria_config (
			empresa_id BIGINT PRIMARY KEY,
			nombre_sistema TEXT DEFAULT 'Tesoreria y presupuesto',
			moneda TEXT DEFAULT 'COP',
			periodo_trabajo TEXT,
			metodo_proyeccion TEXT DEFAULT 'mensual',
			alerta_saldo_minimo INTEGER DEFAULT 1,
			requiere_aprobacion_pago INTEGER DEFAULT 1,
			fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_tesoreria_cuentas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			tipo TEXT DEFAULT 'banco',
			entidad TEXT,
			numero TEXT,
			moneda TEXT DEFAULT 'COP',
			saldo_inicial NUMERIC(14,2) DEFAULT 0,
			saldo_actual NUMERIC(14,2) DEFAULT 0,
			saldo_minimo NUMERIC(14,2) DEFAULT 0,
			responsable TEXT,
			estado TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_tes_cuenta_empresa_codigo ON empresa_tesoreria_cuentas(empresa_id,codigo)`,
		`CREATE INDEX IF NOT EXISTS ix_tes_cuenta_empresa_estado ON empresa_tesoreria_cuentas(empresa_id,estado)`,
		`CREATE TABLE IF NOT EXISTS empresa_tesoreria_presupuestos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			periodo TEXT NOT NULL,
			escenario TEXT DEFAULT 'base',
			ingresos_meta NUMERIC(14,2) DEFAULT 0,
			egresos_meta NUMERIC(14,2) DEFAULT 0,
			saldo_inicial NUMERIC(14,2) DEFAULT 0,
			estado TEXT DEFAULT 'borrador',
			responsable TEXT,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_tes_pres_empresa_codigo ON empresa_tesoreria_presupuestos(empresa_id,codigo)`,
		`CREATE INDEX IF NOT EXISTS ix_tes_pres_empresa_periodo ON empresa_tesoreria_presupuestos(empresa_id,periodo,estado)`,
		`CREATE TABLE IF NOT EXISTS empresa_tesoreria_partidas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			presupuesto_id BIGINT NOT NULL,
			categoria TEXT DEFAULT 'general',
			tipo TEXT DEFAULT 'egreso',
			concepto TEXT NOT NULL,
			valor_presupuestado NUMERIC(14,2) DEFAULT 0,
			valor_ejecutado NUMERIC(14,2) DEFAULT 0,
			periodicidad TEXT DEFAULT 'mensual',
			centro_costo TEXT,
			estado TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_tes_partida_empresa_pres ON empresa_tesoreria_partidas(empresa_id,presupuesto_id,tipo,estado)`,
		`CREATE TABLE IF NOT EXISTS empresa_tesoreria_flujo_caja (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			cuenta_id BIGINT DEFAULT 0,
			presupuesto_id BIGINT DEFAULT 0,
			fecha_flujo TEXT NOT NULL,
			periodo TEXT NOT NULL,
			tipo TEXT DEFAULT 'egreso',
			categoria TEXT DEFAULT 'general',
			concepto TEXT NOT NULL,
			valor NUMERIC(14,2) DEFAULT 0,
			valor_ejecutado NUMERIC(14,2) DEFAULT 0,
			origen_modulo TEXT DEFAULT 'tesoreria',
			referencia TEXT,
			estado TEXT DEFAULT 'proyectado',
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_tes_flujo_empresa_periodo ON empresa_tesoreria_flujo_caja(empresa_id,periodo,tipo,estado)`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func defaultTesoreriaConfig(empresaID int64) EmpresaTesoreriaConfig {
	return EmpresaTesoreriaConfig{EmpresaID: empresaID, NombreSistema: "Tesoreria y presupuesto", Moneda: "COP", PeriodoTrabajo: currentPeriodTesoreria(), MetodoProyeccion: "mensual", AlertaSaldoMinimo: true, RequiereAprobacionPago: true}
}

func GetEmpresaTesoreriaConfig(dbConn *sql.DB, empresaID int64) (EmpresaTesoreriaConfig, error) {
	if err := EnsureEmpresaTesoreriaPresupuestoSchema(dbConn); err != nil {
		return EmpresaTesoreriaConfig{}, err
	}
	cfg := defaultTesoreriaConfig(empresaID)
	var alerta, aprobacion int
	err := QueryRowCompat(dbConn, `SELECT empresa_id,COALESCE(nombre_sistema,''),COALESCE(moneda,'COP'),COALESCE(periodo_trabajo,''),COALESCE(metodo_proyeccion,'mensual'),COALESCE(alerta_saldo_minimo,1),COALESCE(requiere_aprobacion_pago,1),COALESCE(fecha_actualizacion,''),COALESCE(usuario_creador,'') FROM empresa_tesoreria_config WHERE empresa_id=?`, empresaID).Scan(&cfg.EmpresaID, &cfg.NombreSistema, &cfg.Moneda, &cfg.PeriodoTrabajo, &cfg.MetodoProyeccion, &alerta, &aprobacion, &cfg.FechaActualizacion, &cfg.UsuarioCreador)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return cfg, nil
		}
		return EmpresaTesoreriaConfig{}, err
	}
	cfg.AlertaSaldoMinimo = alerta > 0
	cfg.RequiereAprobacionPago = aprobacion > 0
	return normalizeTesoreriaConfig(cfg), nil
}

func UpsertEmpresaTesoreriaConfig(dbConn *sql.DB, cfg EmpresaTesoreriaConfig) error {
	if err := EnsureEmpresaTesoreriaPresupuestoSchema(dbConn); err != nil {
		return err
	}
	cfg = normalizeTesoreriaConfig(cfg)
	_, err := ExecCompat(dbConn, `INSERT INTO empresa_tesoreria_config (empresa_id,nombre_sistema,moneda,periodo_trabajo,metodo_proyeccion,alerta_saldo_minimo,requiere_aprobacion_pago,fecha_actualizacion,usuario_creador) VALUES (?,?,?,?,?,?,?,CAST(CURRENT_TIMESTAMP AS TEXT),?)
	ON CONFLICT (empresa_id) DO UPDATE SET nombre_sistema=EXCLUDED.nombre_sistema, moneda=EXCLUDED.moneda, periodo_trabajo=EXCLUDED.periodo_trabajo, metodo_proyeccion=EXCLUDED.metodo_proyeccion, alerta_saldo_minimo=EXCLUDED.alerta_saldo_minimo, requiere_aprobacion_pago=EXCLUDED.requiere_aprobacion_pago, fecha_actualizacion=CAST(CURRENT_TIMESTAMP AS TEXT), usuario_creador=EXCLUDED.usuario_creador`,
		cfg.EmpresaID, cfg.NombreSistema, cfg.Moneda, cfg.PeriodoTrabajo, cfg.MetodoProyeccion, boolIntTesoreria(cfg.AlertaSaldoMinimo), boolIntTesoreria(cfg.RequiereAprobacionPago), cfg.UsuarioCreador)
	return err
}

func UpsertEmpresaTesoreriaCuenta(dbConn *sql.DB, item EmpresaTesoreriaCuenta) (int64, error) {
	if err := EnsureEmpresaTesoreriaPresupuestoSchema(dbConn); err != nil {
		return 0, err
	}
	item = normalizeTesoreriaCuenta(item)
	if item.EmpresaID <= 0 || item.Codigo == "" || item.Nombre == "" {
		return 0, errors.New("empresa_id, codigo y nombre son obligatorios")
	}
	if item.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_tesoreria_cuentas SET codigo=?,nombre=?,tipo=?,entidad=?,numero=?,moneda=?,saldo_inicial=?,saldo_actual=?,saldo_minimo=?,responsable=?,estado=?,fecha_actualizacion=CAST(CURRENT_TIMESTAMP AS TEXT) WHERE empresa_id=? AND id=?`,
			item.Codigo, item.Nombre, item.Tipo, item.Entidad, item.Numero, item.Moneda, item.SaldoInicial, item.SaldoActual, item.SaldoMinimo, item.Responsable, item.Estado, item.EmpresaID, item.ID)
		return item.ID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_tesoreria_cuentas (empresa_id,codigo,nombre,tipo,entidad,numero,moneda,saldo_inicial,saldo_actual,saldo_minimo,responsable,estado,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		item.EmpresaID, item.Codigo, item.Nombre, item.Tipo, item.Entidad, item.Numero, item.Moneda, item.SaldoInicial, item.SaldoActual, item.SaldoMinimo, item.Responsable, item.Estado, item.UsuarioCreador)
}

func ListEmpresaTesoreriaCuentas(dbConn *sql.DB, empresaID int64, estado string, limit int) ([]EmpresaTesoreriaCuenta, error) {
	if err := EnsureEmpresaTesoreriaPresupuestoSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 150
	}
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if strings.TrimSpace(estado) != "" {
		where += " AND LOWER(COALESCE(estado,''))=?"
		args = append(args, normalizeEstadoTesoreria(estado))
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,COALESCE(codigo,''),COALESCE(nombre,''),COALESCE(tipo,'banco'),COALESCE(entidad,''),COALESCE(numero,''),COALESCE(moneda,'COP'),COALESCE(saldo_inicial,0),COALESCE(saldo_actual,0),COALESCE(saldo_minimo,0),COALESCE(responsable,''),COALESCE(estado,'activo'),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,''),COALESCE(usuario_creador,'') FROM empresa_tesoreria_cuentas WHERE %s ORDER BY id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaTesoreriaCuenta{}
	for rows.Next() {
		var x EmpresaTesoreriaCuenta
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Codigo, &x.Nombre, &x.Tipo, &x.Entidad, &x.Numero, &x.Moneda, &x.SaldoInicial, &x.SaldoActual, &x.SaldoMinimo, &x.Responsable, &x.Estado, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func UpsertEmpresaTesoreriaPresupuesto(dbConn *sql.DB, item EmpresaTesoreriaPresupuesto) (int64, error) {
	if err := EnsureEmpresaTesoreriaPresupuestoSchema(dbConn); err != nil {
		return 0, err
	}
	item = normalizeTesoreriaPresupuesto(item)
	if item.EmpresaID <= 0 || item.Codigo == "" || item.Nombre == "" || item.Periodo == "" {
		return 0, errors.New("empresa_id, codigo, nombre y periodo son obligatorios")
	}
	if item.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_tesoreria_presupuestos SET codigo=?,nombre=?,periodo=?,escenario=?,ingresos_meta=?,egresos_meta=?,saldo_inicial=?,estado=?,responsable=?,observaciones=?,fecha_actualizacion=CAST(CURRENT_TIMESTAMP AS TEXT) WHERE empresa_id=? AND id=?`,
			item.Codigo, item.Nombre, item.Periodo, item.Escenario, item.IngresosMeta, item.EgresosMeta, item.SaldoInicial, item.Estado, item.Responsable, item.Observaciones, item.EmpresaID, item.ID)
		return item.ID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_tesoreria_presupuestos (empresa_id,codigo,nombre,periodo,escenario,ingresos_meta,egresos_meta,saldo_inicial,estado,responsable,observaciones,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
		item.EmpresaID, item.Codigo, item.Nombre, item.Periodo, item.Escenario, item.IngresosMeta, item.EgresosMeta, item.SaldoInicial, item.Estado, item.Responsable, item.Observaciones, item.UsuarioCreador)
}

func ListEmpresaTesoreriaPresupuestos(dbConn *sql.DB, empresaID int64, periodo string, limit int) ([]EmpresaTesoreriaPresupuesto, error) {
	if err := EnsureEmpresaTesoreriaPresupuestoSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 120
	}
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if strings.TrimSpace(periodo) != "" {
		where += " AND periodo=?"
		args = append(args, strings.TrimSpace(periodo))
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,COALESCE(codigo,''),COALESCE(nombre,''),COALESCE(periodo,''),COALESCE(escenario,'base'),COALESCE(ingresos_meta,0),COALESCE(egresos_meta,0),COALESCE(saldo_inicial,0),COALESCE(estado,'borrador'),COALESCE(responsable,''),COALESCE(observaciones,''),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,''),COALESCE(usuario_creador,'') FROM empresa_tesoreria_presupuestos WHERE %s ORDER BY periodo DESC,id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaTesoreriaPresupuesto{}
	for rows.Next() {
		var x EmpresaTesoreriaPresupuesto
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Codigo, &x.Nombre, &x.Periodo, &x.Escenario, &x.IngresosMeta, &x.EgresosMeta, &x.SaldoInicial, &x.Estado, &x.Responsable, &x.Observaciones, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func UpsertEmpresaTesoreriaPartida(dbConn *sql.DB, item EmpresaTesoreriaPartida) (int64, error) {
	if err := EnsureEmpresaTesoreriaPresupuestoSchema(dbConn); err != nil {
		return 0, err
	}
	item = normalizeTesoreriaPartida(item)
	if item.EmpresaID <= 0 || item.PresupuestoID <= 0 || item.Concepto == "" {
		return 0, errors.New("presupuesto y concepto son obligatorios")
	}
	if item.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_tesoreria_partidas SET categoria=?,tipo=?,concepto=?,valor_presupuestado=?,valor_ejecutado=?,periodicidad=?,centro_costo=?,estado=?,fecha_actualizacion=CAST(CURRENT_TIMESTAMP AS TEXT) WHERE empresa_id=? AND id=?`,
			item.Categoria, item.Tipo, item.Concepto, item.ValorPresupuestado, item.ValorEjecutado, item.Periodicidad, item.CentroCosto, item.Estado, item.EmpresaID, item.ID)
		return item.ID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_tesoreria_partidas (empresa_id,presupuesto_id,categoria,tipo,concepto,valor_presupuestado,valor_ejecutado,periodicidad,centro_costo,estado,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		item.EmpresaID, item.PresupuestoID, item.Categoria, item.Tipo, item.Concepto, item.ValorPresupuestado, item.ValorEjecutado, item.Periodicidad, item.CentroCosto, item.Estado, item.UsuarioCreador)
}

func ListEmpresaTesoreriaPartidas(dbConn *sql.DB, empresaID, presupuestoID int64, limit int) ([]EmpresaTesoreriaPartida, error) {
	if limit <= 0 || limit > 500 {
		limit = 250
	}
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if presupuestoID > 0 {
		where += " AND presupuesto_id=?"
		args = append(args, presupuestoID)
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,presupuesto_id,COALESCE(categoria,'general'),COALESCE(tipo,'egreso'),COALESCE(concepto,''),COALESCE(valor_presupuestado,0),COALESCE(valor_ejecutado,0),COALESCE(periodicidad,'mensual'),COALESCE(centro_costo,''),COALESCE(estado,'activo'),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,''),COALESCE(usuario_creador,'') FROM empresa_tesoreria_partidas WHERE %s ORDER BY id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaTesoreriaPartida{}
	for rows.Next() {
		var x EmpresaTesoreriaPartida
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.PresupuestoID, &x.Categoria, &x.Tipo, &x.Concepto, &x.ValorPresupuestado, &x.ValorEjecutado, &x.Periodicidad, &x.CentroCosto, &x.Estado, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func CreateEmpresaTesoreriaFlujo(dbConn *sql.DB, item EmpresaTesoreriaFlujo) (int64, error) {
	if err := EnsureEmpresaTesoreriaPresupuestoSchema(dbConn); err != nil {
		return 0, err
	}
	item = normalizeTesoreriaFlujo(item)
	if item.EmpresaID <= 0 || item.Concepto == "" || item.FechaFlujo == "" {
		return 0, errors.New("fecha y concepto son obligatorios")
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_tesoreria_flujo_caja (empresa_id,cuenta_id,presupuesto_id,fecha_flujo,periodo,tipo,categoria,concepto,valor,valor_ejecutado,origen_modulo,referencia,estado,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		item.EmpresaID, item.CuentaID, item.PresupuestoID, item.FechaFlujo, item.Periodo, item.Tipo, item.Categoria, item.Concepto, item.Valor, item.ValorEjecutado, item.OrigenModulo, item.Referencia, item.Estado, item.UsuarioCreador)
}

func ListEmpresaTesoreriaFlujo(dbConn *sql.DB, empresaID int64, periodo string, limit int) ([]EmpresaTesoreriaFlujo, error) {
	if err := EnsureEmpresaTesoreriaPresupuestoSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 250
	}
	args := []interface{}{empresaID}
	where := "f.empresa_id=?"
	if strings.TrimSpace(periodo) != "" {
		where += " AND f.periodo=?"
		args = append(args, strings.TrimSpace(periodo))
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT f.id,f.empresa_id,COALESCE(f.cuenta_id,0),COALESCE(c.nombre,''),COALESCE(f.presupuesto_id,0),COALESCE(f.fecha_flujo,''),COALESCE(f.periodo,''),COALESCE(f.tipo,'egreso'),COALESCE(f.categoria,'general'),COALESCE(f.concepto,''),COALESCE(f.valor,0),COALESCE(f.valor_ejecutado,0),COALESCE(f.origen_modulo,'tesoreria'),COALESCE(f.referencia,''),COALESCE(f.estado,'proyectado'),COALESCE(f.usuario_creador,''),COALESCE(f.fecha_creacion,'') FROM empresa_tesoreria_flujo_caja f LEFT JOIN empresa_tesoreria_cuentas c ON c.id=f.cuenta_id AND c.empresa_id=f.empresa_id WHERE %s ORDER BY f.fecha_flujo ASC,f.id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaTesoreriaFlujo{}
	for rows.Next() {
		var x EmpresaTesoreriaFlujo
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.CuentaID, &x.CuentaNombre, &x.PresupuestoID, &x.FechaFlujo, &x.Periodo, &x.Tipo, &x.Categoria, &x.Concepto, &x.Valor, &x.ValorEjecutado, &x.OrigenModulo, &x.Referencia, &x.Estado, &x.UsuarioCreador, &x.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func GenerarEmpresaTesoreriaFlujoDesdePresupuesto(dbConn *sql.DB, empresaID, presupuestoID int64, usuario string) ([]EmpresaTesoreriaFlujo, error) {
	partidas, err := ListEmpresaTesoreriaPartidas(dbConn, empresaID, presupuestoID, 500)
	if err != nil {
		return nil, err
	}
	var periodo string
	if err := QueryRowCompat(dbConn, `SELECT COALESCE(periodo,'') FROM empresa_tesoreria_presupuestos WHERE empresa_id=? AND id=?`, empresaID, presupuestoID).Scan(&periodo); err != nil {
		return nil, err
	}
	if periodo == "" {
		periodo = currentPeriodTesoreria()
	}
	_, _ = ExecCompat(dbConn, `UPDATE empresa_tesoreria_flujo_caja SET estado='reemplazado' WHERE empresa_id=? AND presupuesto_id=? AND estado='proyectado'`, empresaID, presupuestoID)
	for _, p := range partidas {
		_, err := CreateEmpresaTesoreriaFlujo(dbConn, EmpresaTesoreriaFlujo{EmpresaID: empresaID, PresupuestoID: presupuestoID, FechaFlujo: periodo + "-01", Periodo: periodo, Tipo: p.Tipo, Categoria: p.Categoria, Concepto: p.Concepto, Valor: p.ValorPresupuestado, ValorEjecutado: p.ValorEjecutado, OrigenModulo: "presupuesto", Estado: "proyectado", UsuarioCreador: usuario})
		if err != nil {
			return nil, err
		}
	}
	return ListEmpresaTesoreriaFlujo(dbConn, empresaID, periodo, 500)
}

func BuildEmpresaTesoreriaDashboard(dbConn *sql.DB, empresaID int64) (EmpresaTesoreriaDashboard, error) {
	if err := EnsureEmpresaTesoreriaPresupuestoSchema(dbConn); err != nil {
		return EmpresaTesoreriaDashboard{}, err
	}
	cfg, err := GetEmpresaTesoreriaConfig(dbConn, empresaID)
	if err != nil {
		return EmpresaTesoreriaDashboard{}, err
	}
	cuentas, _ := ListEmpresaTesoreriaCuentas(dbConn, empresaID, "", 80)
	pres, _ := ListEmpresaTesoreriaPresupuestos(dbConn, empresaID, cfg.PeriodoTrabajo, 80)
	partidas, _ := ListEmpresaTesoreriaPartidas(dbConn, empresaID, 0, 120)
	flujo, _ := ListEmpresaTesoreriaFlujo(dbConn, empresaID, cfg.PeriodoTrabajo, 160)
	ds := EmpresaTesoreriaDashboard{EmpresaID: empresaID, Config: cfg, Cuentas: cuentas, Presupuestos: pres, Partidas: partidas, Flujo: flujo}
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*),COALESCE(SUM(saldo_actual),0) FROM empresa_tesoreria_cuentas WHERE empresa_id=? AND estado='activo'`, empresaID).Scan(&ds.CuentasActivas, &ds.SaldoDisponible)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_tesoreria_presupuestos WHERE empresa_id=? AND estado IN ('aprobado','activo')`, empresaID).Scan(&ds.PresupuestosActivos)
	for _, f := range flujo {
		if f.Tipo == "ingreso" {
			ds.IngresosProyectados += f.Valor
			ds.EjecucionIngresos += f.ValorEjecutado
		} else {
			ds.EgresosProyectados += f.Valor
			ds.EjecucionEgresos += f.ValorEjecutado
		}
	}
	ds.FlujoNeto = roundMoneyTesoreria(ds.SaldoDisponible + ds.IngresosProyectados - ds.EgresosProyectados)
	return ds, nil
}

func SeedEmpresaTesoreriaDemo(dbConn *sql.DB, empresaID int64, user string) error {
	periodo := currentPeriodTesoreria()
	if err := UpsertEmpresaTesoreriaConfig(dbConn, EmpresaTesoreriaConfig{EmpresaID: empresaID, NombreSistema: "Tesoreria y presupuesto", Moneda: "COP", PeriodoTrabajo: periodo, MetodoProyeccion: "mensual", AlertaSaldoMinimo: true, RequiereAprobacionPago: true, UsuarioCreador: user}); err != nil {
		return err
	}
	cuentaID, err := UpsertEmpresaTesoreriaCuenta(dbConn, EmpresaTesoreriaCuenta{EmpresaID: empresaID, Codigo: "BANCO-PPAL", Nombre: "Banco principal", Tipo: "banco", Entidad: "Banco empresarial", Numero: "000-000", Moneda: "COP", SaldoInicial: 25000000, SaldoActual: 25000000, SaldoMinimo: 5000000, Responsable: "Tesoreria", Estado: "activo", UsuarioCreador: user})
	if err != nil {
		return err
	}
	presID, err := UpsertEmpresaTesoreriaPresupuesto(dbConn, EmpresaTesoreriaPresupuesto{EmpresaID: empresaID, Codigo: "PRES-" + periodo, Nombre: "Presupuesto operativo " + periodo, Periodo: periodo, Escenario: "base", IngresosMeta: 42000000, EgresosMeta: 28000000, SaldoInicial: 25000000, Estado: "aprobado", Responsable: "Gerencia", UsuarioCreador: user})
	if err != nil {
		return err
	}
	_, _ = UpsertEmpresaTesoreriaPartida(dbConn, EmpresaTesoreriaPartida{EmpresaID: empresaID, PresupuestoID: presID, Categoria: "ventas", Tipo: "ingreso", Concepto: "Ingresos por ventas proyectadas", ValorPresupuestado: 42000000, Periodicidad: "mensual", Estado: "activo", UsuarioCreador: user})
	_, _ = UpsertEmpresaTesoreriaPartida(dbConn, EmpresaTesoreriaPartida{EmpresaID: empresaID, PresupuestoID: presID, Categoria: "nomina", Tipo: "egreso", Concepto: "Nomina y prestaciones", ValorPresupuestado: 14000000, Periodicidad: "mensual", Estado: "activo", UsuarioCreador: user})
	_, _ = UpsertEmpresaTesoreriaPartida(dbConn, EmpresaTesoreriaPartida{EmpresaID: empresaID, PresupuestoID: presID, Categoria: "proveedores", Tipo: "egreso", Concepto: "Pagos a proveedores", ValorPresupuestado: 9000000, Periodicidad: "mensual", Estado: "activo", UsuarioCreador: user})
	_, err = GenerarEmpresaTesoreriaFlujoDesdePresupuesto(dbConn, empresaID, presID, user)
	if err != nil {
		return err
	}
	_, err = CreateEmpresaTesoreriaFlujo(dbConn, EmpresaTesoreriaFlujo{EmpresaID: empresaID, CuentaID: cuentaID, PresupuestoID: presID, FechaFlujo: periodo + "-15", Periodo: periodo, Tipo: "egreso", Categoria: "impuestos", Concepto: "Reserva impuestos", Valor: 3500000, OrigenModulo: "tesoreria", Estado: "proyectado", UsuarioCreador: user})
	return err
}

func normalizeTesoreriaConfig(x EmpresaTesoreriaConfig) EmpresaTesoreriaConfig {
	x.NombreSistema = strings.TrimSpace(x.NombreSistema)
	if x.NombreSistema == "" {
		x.NombreSistema = "Tesoreria y presupuesto"
	}
	x.Moneda = strings.ToUpper(strings.TrimSpace(x.Moneda))
	if x.Moneda == "" {
		x.Moneda = "COP"
	}
	x.PeriodoTrabajo = strings.TrimSpace(x.PeriodoTrabajo)
	if x.PeriodoTrabajo == "" {
		x.PeriodoTrabajo = currentPeriodTesoreria()
	}
	x.MetodoProyeccion = normalizeOneOfTesoreria(x.MetodoProyeccion, "mensual", "semanal", "mensual", "trimestral", "manual")
	return x
}

func normalizeTesoreriaCuenta(x EmpresaTesoreriaCuenta) EmpresaTesoreriaCuenta {
	x.Codigo = strings.ToUpper(strings.TrimSpace(x.Codigo))
	x.Nombre = strings.TrimSpace(x.Nombre)
	x.Tipo = normalizeOneOfTesoreria(x.Tipo, "banco", "banco", "caja", "pasarela", "fiducia", "otro")
	x.Moneda = strings.ToUpper(strings.TrimSpace(x.Moneda))
	if x.Moneda == "" {
		x.Moneda = "COP"
	}
	if x.SaldoActual == 0 && x.SaldoInicial != 0 {
		x.SaldoActual = x.SaldoInicial
	}
	x.Estado = normalizeEstadoTesoreria(x.Estado)
	return x
}

func normalizeTesoreriaPresupuesto(x EmpresaTesoreriaPresupuesto) EmpresaTesoreriaPresupuesto {
	x.Codigo = strings.ToUpper(strings.TrimSpace(x.Codigo))
	x.Nombre = strings.TrimSpace(x.Nombre)
	x.Periodo = strings.TrimSpace(x.Periodo)
	if x.Periodo == "" {
		x.Periodo = currentPeriodTesoreria()
	}
	x.Escenario = normalizeOneOfTesoreria(x.Escenario, "base", "base", "optimista", "conservador", "estres")
	x.Estado = normalizeOneOfTesoreria(x.Estado, "borrador", "borrador", "aprobado", "activo", "cerrado", "cancelado")
	x.IngresosMeta = math.Max(0, x.IngresosMeta)
	x.EgresosMeta = math.Max(0, x.EgresosMeta)
	return x
}

func normalizeTesoreriaPartida(x EmpresaTesoreriaPartida) EmpresaTesoreriaPartida {
	x.Categoria = normalizeSlugTesoreria(x.Categoria, "general")
	x.Tipo = normalizeTipoTesoreria(x.Tipo)
	x.Concepto = strings.TrimSpace(x.Concepto)
	x.ValorPresupuestado = math.Max(0, x.ValorPresupuestado)
	x.ValorEjecutado = math.Max(0, x.ValorEjecutado)
	x.Periodicidad = normalizeOneOfTesoreria(x.Periodicidad, "mensual", "unica", "semanal", "mensual", "trimestral", "anual")
	x.Estado = normalizeEstadoTesoreria(x.Estado)
	return x
}

func normalizeTesoreriaFlujo(x EmpresaTesoreriaFlujo) EmpresaTesoreriaFlujo {
	x.Periodo = strings.TrimSpace(x.Periodo)
	if x.Periodo == "" && len(x.FechaFlujo) >= 7 {
		x.Periodo = x.FechaFlujo[:7]
	}
	if x.Periodo == "" {
		x.Periodo = currentPeriodTesoreria()
	}
	x.Tipo = normalizeTipoTesoreria(x.Tipo)
	x.Categoria = normalizeSlugTesoreria(x.Categoria, "general")
	x.Concepto = strings.TrimSpace(x.Concepto)
	x.Valor = math.Max(0, x.Valor)
	x.ValorEjecutado = math.Max(0, x.ValorEjecutado)
	x.OrigenModulo = normalizeSlugTesoreria(x.OrigenModulo, "tesoreria")
	x.Estado = normalizeOneOfTesoreria(x.Estado, "proyectado", "proyectado", "programado", "ejecutado", "cancelado", "reemplazado")
	return x
}

func normalizeTipoTesoreria(v string) string {
	return normalizeOneOfTesoreria(v, "egreso", "ingreso", "egreso")
}

func normalizeEstadoTesoreria(v string) string {
	return normalizeOneOfTesoreria(v, "activo", "activo", "inactivo", "bloqueado", "cerrado")
}

func normalizeOneOfTesoreria(v, fallback string, allowed ...string) string {
	s := normalizeSlugTesoreria(v, fallback)
	for _, a := range allowed {
		if s == a {
			return s
		}
	}
	return fallback
}

func normalizeSlugTesoreria(v, fallback string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	if s == "" {
		return fallback
	}
	return s
}

func currentPeriodTesoreria() string {
	return strings.TrimSpace(fmt.Sprintf("%s", sqlNowPeriodFallback()))
}

func sqlNowPeriodFallback() string {
	return strings.TrimSpace(timeNowCompatString()[:7])
}

func timeNowCompatString() string {
	return strings.ReplaceAll(fmt.Sprintf("%s", nowDomicilio()), "T", " ")
}

func roundMoneyTesoreria(v float64) float64 {
	return math.Round(v*100) / 100
}

func boolIntTesoreria(v bool) int {
	if v {
		return 1
	}
	return 0
}
