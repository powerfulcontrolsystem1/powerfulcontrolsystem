package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type EmpresaPropiedadHorizontalConfig struct {
	EmpresaID               int64   `json:"empresa_id"`
	NombreCopropiedad       string  `json:"nombre_copropiedad"`
	NIT                     string  `json:"nit"`
	TipoCopropiedad         string  `json:"tipo_copropiedad"`
	Direccion               string  `json:"direccion"`
	Ciudad                  string  `json:"ciudad"`
	Administrador           string  `json:"administrador"`
	Telefono                string  `json:"telefono"`
	Email                   string  `json:"email"`
	InteresMoraMensual      float64 `json:"interes_mora_mensual"`
	DiasGracia              int     `json:"dias_gracia"`
	FacturacionElectronica  bool    `json:"facturacion_electronica"`
	PermitirPortalResidente bool    `json:"permitir_portal_residente"`
	FechaActualizacion      string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador          string  `json:"usuario_creador,omitempty"`
}

type EmpresaPropiedadHorizontalUnidad struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	ServicioID         int64   `json:"servicio_id,omitempty"`
	Codigo             string  `json:"codigo"`
	Torre              string  `json:"torre"`
	Piso               string  `json:"piso"`
	TipoUnidad         string  `json:"tipo_unidad"`
	AreaM2             float64 `json:"area_m2"`
	Coeficiente        float64 `json:"coeficiente"`
	CuotaBase          float64 `json:"cuota_base"`
	Parqueadero        string  `json:"parqueadero"`
	Deposito           string  `json:"deposito"`
	Estado             string  `json:"estado"`
	Observaciones      string  `json:"observaciones,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
}

type EmpresaPropiedadHorizontalPersona struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	UnidadID           int64  `json:"unidad_id"`
	ClienteID          int64  `json:"cliente_id,omitempty"`
	UnidadCodigo       string `json:"unidad_codigo,omitempty"`
	TipoRelacion       string `json:"tipo_relacion"`
	Nombre             string `json:"nombre"`
	Documento          string `json:"documento"`
	Telefono           string `json:"telefono"`
	Email              string `json:"email"`
	ContactoEmergencia string `json:"contacto_emergencia"`
	Estado             string `json:"estado"`
	Observaciones      string `json:"observaciones,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
}

type EmpresaPropiedadHorizontalCargo struct {
	ID               int64   `json:"id"`
	EmpresaID        int64   `json:"empresa_id"`
	UnidadID         int64   `json:"unidad_id"`
	ServicioID       int64   `json:"servicio_id,omitempty"`
	UnidadCodigo     string  `json:"unidad_codigo,omitempty"`
	Periodo          string  `json:"periodo"`
	Concepto         string  `json:"concepto"`
	TipoCargo        string  `json:"tipo_cargo"`
	ValorBase        float64 `json:"valor_base"`
	InteresMora      float64 `json:"interes_mora"`
	Descuento        float64 `json:"descuento"`
	Total            float64 `json:"total"`
	SaldoPendiente   float64 `json:"saldo_pendiente"`
	FechaVencimiento string  `json:"fecha_vencimiento"`
	Estado           string  `json:"estado"`
	Observaciones    string  `json:"observaciones,omitempty"`
	FechaCreacion    string  `json:"fecha_creacion,omitempty"`
	UsuarioCreador   string  `json:"usuario_creador,omitempty"`
}

type EmpresaPropiedadHorizontalRecaudo struct {
	ID             int64   `json:"id"`
	EmpresaID      int64   `json:"empresa_id"`
	CargoID        int64   `json:"cargo_id"`
	UnidadID       int64   `json:"unidad_id"`
	ClienteID      int64   `json:"cliente_id,omitempty"`
	ServicioID     int64   `json:"servicio_id,omitempty"`
	CarritoID      int64   `json:"carrito_id,omitempty"`
	CarritoItemID  int64   `json:"carrito_item_id,omitempty"`
	UnidadCodigo   string  `json:"unidad_codigo,omitempty"`
	FechaPago      string  `json:"fecha_pago"`
	MetodoPago     string  `json:"metodo_pago"`
	Referencia     string  `json:"referencia"`
	ValorPagado    float64 `json:"valor_pagado"`
	Observaciones  string  `json:"observaciones,omitempty"`
	FechaCreacion  string  `json:"fecha_creacion,omitempty"`
	UsuarioCreador string  `json:"usuario_creador,omitempty"`
}

type EmpresaPropiedadHorizontalPQR struct {
	ID             int64  `json:"id"`
	EmpresaID      int64  `json:"empresa_id"`
	UnidadID       int64  `json:"unidad_id"`
	UnidadCodigo   string `json:"unidad_codigo,omitempty"`
	Tipo           string `json:"tipo"`
	Prioridad      string `json:"prioridad"`
	Estado         string `json:"estado"`
	Asunto         string `json:"asunto"`
	Descripcion    string `json:"descripcion"`
	Responsable    string `json:"responsable"`
	FechaLimite    string `json:"fecha_limite"`
	FechaCreacion  string `json:"fecha_creacion,omitempty"`
	UsuarioCreador string `json:"usuario_creador,omitempty"`
}

type EmpresaPropiedadHorizontalAsamblea struct {
	ID             int64   `json:"id"`
	EmpresaID      int64   `json:"empresa_id"`
	Titulo         string  `json:"titulo"`
	TipoAsamblea   string  `json:"tipo_asamblea"`
	Fecha          string  `json:"fecha"`
	Estado         string  `json:"estado"`
	QuorumObjetivo float64 `json:"quorum_objetivo"`
	QuorumActual   float64 `json:"quorum_actual"`
	ActaURL        string  `json:"acta_url"`
	Observaciones  string  `json:"observaciones,omitempty"`
	FechaCreacion  string  `json:"fecha_creacion,omitempty"`
	UsuarioCreador string  `json:"usuario_creador,omitempty"`
}

type EmpresaPropiedadHorizontalDashboard struct {
	EmpresaID         int64                                `json:"empresa_id"`
	Config            EmpresaPropiedadHorizontalConfig     `json:"config"`
	TotalUnidades     int                                  `json:"total_unidades"`
	UnidadesOcupadas  int                                  `json:"unidades_ocupadas"`
	ResidentesActivos int                                  `json:"residentes_activos"`
	CarteraPendiente  float64                              `json:"cartera_pendiente"`
	RecaudoMes        float64                              `json:"recaudo_mes"`
	PQRPendientes     int                                  `json:"pqr_pendientes"`
	AsambleasAbiertas int                                  `json:"asambleas_abiertas"`
	Unidades          []EmpresaPropiedadHorizontalUnidad   `json:"unidades"`
	Personas          []EmpresaPropiedadHorizontalPersona  `json:"personas"`
	Cargos            []EmpresaPropiedadHorizontalCargo    `json:"cargos"`
	Recaudos          []EmpresaPropiedadHorizontalRecaudo  `json:"recaudos"`
	PQRs              []EmpresaPropiedadHorizontalPQR      `json:"pqrs"`
	Asambleas         []EmpresaPropiedadHorizontalAsamblea `json:"asambleas"`
}

func EnsureEmpresaPropiedadHorizontalSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_propiedad_horizontal_config (
			empresa_id BIGINT PRIMARY KEY,
			nombre_copropiedad TEXT,
			nit TEXT,
			tipo_copropiedad TEXT DEFAULT 'conjunto_residencial',
			direccion TEXT,
			ciudad TEXT,
			administrador TEXT,
			telefono TEXT,
			email TEXT,
			interes_mora_mensual NUMERIC(8,4) DEFAULT 2,
			dias_gracia INTEGER DEFAULT 5,
			facturacion_electronica INTEGER DEFAULT 1,
			permitir_portal_residente INTEGER DEFAULT 1,
			fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_propiedad_horizontal_unidades (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			servicio_id BIGINT,
			codigo TEXT NOT NULL,
			torre TEXT,
			piso TEXT,
			tipo_unidad TEXT DEFAULT 'apartamento',
			area_m2 NUMERIC(14,2) DEFAULT 0,
			coeficiente NUMERIC(12,6) DEFAULT 0,
			cuota_base NUMERIC(14,2) DEFAULT 0,
			parqueadero TEXT,
			deposito TEXT,
			estado TEXT DEFAULT 'ocupada',
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_prop_h_unidad_empresa_codigo ON empresa_propiedad_horizontal_unidades(empresa_id,codigo)`,
		`CREATE TABLE IF NOT EXISTS empresa_propiedad_horizontal_personas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			unidad_id BIGINT DEFAULT 0,
			cliente_id BIGINT,
			tipo_relacion TEXT DEFAULT 'propietario',
			nombre TEXT NOT NULL,
			documento TEXT,
			telefono TEXT,
			email TEXT,
			contacto_emergencia TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_prop_h_personas_empresa ON empresa_propiedad_horizontal_personas(empresa_id,unidad_id,estado)`,
		`CREATE TABLE IF NOT EXISTS empresa_propiedad_horizontal_cargos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			unidad_id BIGINT NOT NULL,
			servicio_id BIGINT,
			periodo TEXT NOT NULL,
			concepto TEXT NOT NULL,
			tipo_cargo TEXT DEFAULT 'cuota_administracion',
			valor_base NUMERIC(14,2) DEFAULT 0,
			interes_mora NUMERIC(14,2) DEFAULT 0,
			descuento NUMERIC(14,2) DEFAULT 0,
			total NUMERIC(14,2) DEFAULT 0,
			saldo_pendiente NUMERIC(14,2) DEFAULT 0,
			fecha_vencimiento TEXT,
			estado TEXT DEFAULT 'pendiente',
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_prop_h_cargos_empresa ON empresa_propiedad_horizontal_cargos(empresa_id,periodo,estado)`,
		`CREATE TABLE IF NOT EXISTS empresa_propiedad_horizontal_recaudos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			cargo_id BIGINT DEFAULT 0,
			unidad_id BIGINT NOT NULL,
			cliente_id BIGINT,
			servicio_id BIGINT,
			carrito_id BIGINT,
			carrito_item_id BIGINT,
			fecha_pago TEXT,
			metodo_pago TEXT DEFAULT 'transferencia',
			referencia TEXT,
			valor_pagado NUMERIC(14,2) DEFAULT 0,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_prop_h_recaudos_empresa ON empresa_propiedad_horizontal_recaudos(empresa_id,fecha_pago)`,
		`CREATE TABLE IF NOT EXISTS empresa_propiedad_horizontal_pqrs (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			unidad_id BIGINT DEFAULT 0,
			tipo TEXT DEFAULT 'peticion',
			prioridad TEXT DEFAULT 'media',
			estado TEXT DEFAULT 'abierta',
			asunto TEXT NOT NULL,
			descripcion TEXT,
			responsable TEXT,
			fecha_limite TEXT,
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_prop_h_pqrs_empresa ON empresa_propiedad_horizontal_pqrs(empresa_id,estado,prioridad)`,
		`CREATE TABLE IF NOT EXISTS empresa_propiedad_horizontal_asambleas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			titulo TEXT NOT NULL,
			tipo_asamblea TEXT DEFAULT 'ordinaria',
			fecha TEXT,
			estado TEXT DEFAULT 'programada',
			quorum_objetivo NUMERIC(8,4) DEFAULT 51,
			quorum_actual NUMERIC(8,4) DEFAULT 0,
			acta_url TEXT,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_prop_h_asambleas_empresa ON empresa_propiedad_horizontal_asambleas(empresa_id,estado,fecha)`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	extraColumns := []struct {
		table  string
		column string
		def    string
	}{
		{"empresa_propiedad_horizontal_unidades", "servicio_id", "BIGINT"},
		{"empresa_propiedad_horizontal_personas", "cliente_id", "BIGINT"},
		{"empresa_propiedad_horizontal_cargos", "servicio_id", "BIGINT"},
		{"empresa_propiedad_horizontal_recaudos", "cliente_id", "BIGINT"},
		{"empresa_propiedad_horizontal_recaudos", "servicio_id", "BIGINT"},
		{"empresa_propiedad_horizontal_recaudos", "carrito_id", "BIGINT"},
		{"empresa_propiedad_horizontal_recaudos", "carrito_item_id", "BIGINT"},
	}
	for _, col := range extraColumns {
		if err := ensureColumnIfMissing(dbConn, col.table, col.column, col.def); err != nil {
			return err
		}
	}
	for _, stmt := range []string{
		`CREATE INDEX IF NOT EXISTS ix_prop_h_personas_cliente ON empresa_propiedad_horizontal_personas(empresa_id, cliente_id)`,
		`CREATE INDEX IF NOT EXISTS ix_prop_h_cargos_servicio ON empresa_propiedad_horizontal_cargos(empresa_id, servicio_id)`,
		`CREATE INDEX IF NOT EXISTS ix_prop_h_recaudos_carrito ON empresa_propiedad_horizontal_recaudos(empresa_id, carrito_id)`,
	} {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func GetEmpresaPropiedadHorizontalConfig(dbConn *sql.DB, empresaID int64) (EmpresaPropiedadHorizontalConfig, error) {
	if err := EnsureEmpresaPropiedadHorizontalSchema(dbConn); err != nil {
		return EmpresaPropiedadHorizontalConfig{}, err
	}
	var cfg EmpresaPropiedadHorizontalConfig
	var fe, portal int
	err := QueryRowCompat(dbConn, `SELECT empresa_id,COALESCE(nombre_copropiedad,''),COALESCE(nit,''),COALESCE(tipo_copropiedad,'conjunto_residencial'),COALESCE(direccion,''),COALESCE(ciudad,''),COALESCE(administrador,''),COALESCE(telefono,''),COALESCE(email,''),COALESCE(interes_mora_mensual,2),COALESCE(dias_gracia,5),COALESCE(facturacion_electronica,1),COALESCE(permitir_portal_residente,1),COALESCE(fecha_actualizacion,''),COALESCE(usuario_creador,'') FROM empresa_propiedad_horizontal_config WHERE empresa_id=?`, empresaID).Scan(&cfg.EmpresaID, &cfg.NombreCopropiedad, &cfg.NIT, &cfg.TipoCopropiedad, &cfg.Direccion, &cfg.Ciudad, &cfg.Administrador, &cfg.Telefono, &cfg.Email, &cfg.InteresMoraMensual, &cfg.DiasGracia, &fe, &portal, &cfg.FechaActualizacion, &cfg.UsuarioCreador)
	if errors.Is(err, sql.ErrNoRows) {
		cfg = EmpresaPropiedadHorizontalConfig{EmpresaID: empresaID, NombreCopropiedad: "Copropiedad empresarial", TipoCopropiedad: "conjunto_residencial", InteresMoraMensual: 2, DiasGracia: 5, FacturacionElectronica: true, PermitirPortalResidente: true}
		return cfg, nil
	}
	cfg.FacturacionElectronica = fe > 0
	cfg.PermitirPortalResidente = portal > 0
	return cfg, err
}

func UpsertEmpresaPropiedadHorizontalConfig(dbConn *sql.DB, cfg EmpresaPropiedadHorizontalConfig) error {
	if err := EnsureEmpresaPropiedadHorizontalSchema(dbConn); err != nil {
		return err
	}
	cfg = normalizePropiedadHorizontalConfig(cfg)
	if cfg.EmpresaID <= 0 {
		return errors.New("empresa_id es obligatorio")
	}
	_, err := ExecCompat(dbConn, `INSERT INTO empresa_propiedad_horizontal_config (empresa_id,nombre_copropiedad,nit,tipo_copropiedad,direccion,ciudad,administrador,telefono,email,interes_mora_mensual,dias_gracia,facturacion_electronica,permitir_portal_residente,fecha_actualizacion,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,CAST(CURRENT_TIMESTAMP AS TEXT),?)
		ON CONFLICT (empresa_id) DO UPDATE SET nombre_copropiedad=EXCLUDED.nombre_copropiedad,nit=EXCLUDED.nit,tipo_copropiedad=EXCLUDED.tipo_copropiedad,direccion=EXCLUDED.direccion,ciudad=EXCLUDED.ciudad,administrador=EXCLUDED.administrador,telefono=EXCLUDED.telefono,email=EXCLUDED.email,interes_mora_mensual=EXCLUDED.interes_mora_mensual,dias_gracia=EXCLUDED.dias_gracia,facturacion_electronica=EXCLUDED.facturacion_electronica,permitir_portal_residente=EXCLUDED.permitir_portal_residente,fecha_actualizacion=CAST(CURRENT_TIMESTAMP AS TEXT),usuario_creador=EXCLUDED.usuario_creador`,
		cfg.EmpresaID, cfg.NombreCopropiedad, cfg.NIT, cfg.TipoCopropiedad, cfg.Direccion, cfg.Ciudad, cfg.Administrador, cfg.Telefono, cfg.Email, cfg.InteresMoraMensual, cfg.DiasGracia, boolIntPropH(cfg.FacturacionElectronica), boolIntPropH(cfg.PermitirPortalResidente), cfg.UsuarioCreador)
	return err
}

func UpsertEmpresaPropiedadHorizontalUnidad(dbConn *sql.DB, x EmpresaPropiedadHorizontalUnidad) (int64, error) {
	if err := EnsureEmpresaPropiedadHorizontalSchema(dbConn); err != nil {
		return 0, err
	}
	x = normalizePropiedadHorizontalUnidad(x)
	if x.EmpresaID <= 0 || x.Codigo == "" {
		return 0, errors.New("empresa_id y codigo son obligatorios")
	}
	var id int64
	err := QueryRowCompat(dbConn, `INSERT INTO empresa_propiedad_horizontal_unidades (empresa_id,servicio_id,codigo,torre,piso,tipo_unidad,area_m2,coeficiente,cuota_base,parqueadero,deposito,estado,observaciones,fecha_creacion,fecha_actualizacion,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,CAST(CURRENT_TIMESTAMP AS TEXT),CAST(CURRENT_TIMESTAMP AS TEXT),?)
		ON CONFLICT (empresa_id,codigo) DO UPDATE SET torre=EXCLUDED.torre,piso=EXCLUDED.piso,tipo_unidad=EXCLUDED.tipo_unidad,area_m2=EXCLUDED.area_m2,coeficiente=EXCLUDED.coeficiente,cuota_base=EXCLUDED.cuota_base,parqueadero=EXCLUDED.parqueadero,deposito=EXCLUDED.deposito,estado=EXCLUDED.estado,observaciones=EXCLUDED.observaciones,fecha_actualizacion=CAST(CURRENT_TIMESTAMP AS TEXT),usuario_creador=EXCLUDED.usuario_creador RETURNING id`,
		x.EmpresaID, nullableID(x.ServicioID), x.Codigo, x.Torre, x.Piso, x.TipoUnidad, x.AreaM2, x.Coeficiente, x.CuotaBase, x.Parqueadero, x.Deposito, x.Estado, x.Observaciones, x.UsuarioCreador).Scan(&id)
	if err == nil {
		x.ID = id
		if servicioID, srvErr := ensurePropHUnidadServicio(dbConn, x, x.UsuarioCreador); srvErr == nil && servicioID > 0 {
			_, _ = ExecCompat(dbConn, `UPDATE empresa_propiedad_horizontal_unidades SET servicio_id=?, fecha_actualizacion=CAST(CURRENT_TIMESTAMP AS TEXT) WHERE empresa_id=? AND id=?`, nullableID(servicioID), x.EmpresaID, id)
		}
	}
	return id, err
}

func ListEmpresaPropiedadHorizontalUnidades(dbConn *sql.DB, empresaID int64) ([]EmpresaPropiedadHorizontalUnidad, error) {
	if err := EnsureEmpresaPropiedadHorizontalSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,COALESCE(servicio_id,0),COALESCE(codigo,''),COALESCE(torre,''),COALESCE(piso,''),COALESCE(tipo_unidad,''),COALESCE(area_m2,0),COALESCE(coeficiente,0),COALESCE(cuota_base,0),COALESCE(parqueadero,''),COALESCE(deposito,''),COALESCE(estado,''),COALESCE(observaciones,''),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,''),COALESCE(usuario_creador,'') FROM empresa_propiedad_horizontal_unidades WHERE empresa_id=? ORDER BY torre,codigo`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaPropiedadHorizontalUnidad{}
	for rows.Next() {
		var x EmpresaPropiedadHorizontalUnidad
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.ServicioID, &x.Codigo, &x.Torre, &x.Piso, &x.TipoUnidad, &x.AreaM2, &x.Coeficiente, &x.CuotaBase, &x.Parqueadero, &x.Deposito, &x.Estado, &x.Observaciones, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func UpsertEmpresaPropiedadHorizontalPersona(dbConn *sql.DB, x EmpresaPropiedadHorizontalPersona) (int64, error) {
	if err := EnsureEmpresaPropiedadHorizontalSchema(dbConn); err != nil {
		return 0, err
	}
	x = normalizePropiedadHorizontalPersona(x)
	if x.EmpresaID <= 0 || x.Nombre == "" {
		return 0, errors.New("empresa_id y nombre son obligatorios")
	}
	clienteID, err := ensurePropHPersonaClienteCore(dbConn, x, x.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	x.ClienteID = clienteID
	if x.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_propiedad_horizontal_personas SET unidad_id=?,cliente_id=?,tipo_relacion=?,nombre=?,documento=?,telefono=?,email=?,contacto_emergencia=?,estado=?,observaciones=?,usuario_creador=? WHERE empresa_id=? AND id=?`, x.UnidadID, nullableID(x.ClienteID), x.TipoRelacion, x.Nombre, x.Documento, x.Telefono, x.Email, x.ContactoEmergencia, x.Estado, x.Observaciones, x.UsuarioCreador, x.EmpresaID, x.ID)
		return x.ID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_propiedad_horizontal_personas (empresa_id,unidad_id,cliente_id,tipo_relacion,nombre,documento,telefono,email,contacto_emergencia,estado,observaciones,fecha_creacion,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,CAST(CURRENT_TIMESTAMP AS TEXT),?)`, x.EmpresaID, x.UnidadID, nullableID(x.ClienteID), x.TipoRelacion, x.Nombre, x.Documento, x.Telefono, x.Email, x.ContactoEmergencia, x.Estado, x.Observaciones, x.UsuarioCreador)
}

func ListEmpresaPropiedadHorizontalPersonas(dbConn *sql.DB, empresaID int64) ([]EmpresaPropiedadHorizontalPersona, error) {
	if err := EnsureEmpresaPropiedadHorizontalSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT p.id,p.empresa_id,COALESCE(p.unidad_id,0),COALESCE(p.cliente_id,0),COALESCE(u.codigo,''),COALESCE(p.tipo_relacion,''),COALESCE(p.nombre,''),COALESCE(p.documento,''),COALESCE(p.telefono,''),COALESCE(p.email,''),COALESCE(p.contacto_emergencia,''),COALESCE(p.estado,''),COALESCE(p.observaciones,''),COALESCE(p.fecha_creacion,''),COALESCE(p.usuario_creador,'') FROM empresa_propiedad_horizontal_personas p LEFT JOIN empresa_propiedad_horizontal_unidades u ON u.id=p.unidad_id AND u.empresa_id=p.empresa_id WHERE p.empresa_id=? ORDER BY p.id DESC LIMIT 300`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaPropiedadHorizontalPersona{}
	for rows.Next() {
		var x EmpresaPropiedadHorizontalPersona
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.UnidadID, &x.ClienteID, &x.UnidadCodigo, &x.TipoRelacion, &x.Nombre, &x.Documento, &x.Telefono, &x.Email, &x.ContactoEmergencia, &x.Estado, &x.Observaciones, &x.FechaCreacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func CreateEmpresaPropiedadHorizontalCargo(dbConn *sql.DB, x EmpresaPropiedadHorizontalCargo) (int64, error) {
	if err := EnsureEmpresaPropiedadHorizontalSchema(dbConn); err != nil {
		return 0, err
	}
	x = normalizePropiedadHorizontalCargo(x)
	if x.EmpresaID <= 0 || x.UnidadID <= 0 || x.Concepto == "" {
		return 0, errors.New("empresa_id, unidad_id y concepto son obligatorios")
	}
	servicioID, err := ensurePropHCargoServicio(dbConn, x, x.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	x.ServicioID = servicioID
	return insertSQLCompat(dbConn, `INSERT INTO empresa_propiedad_horizontal_cargos (empresa_id,unidad_id,servicio_id,periodo,concepto,tipo_cargo,valor_base,interes_mora,descuento,total,saldo_pendiente,fecha_vencimiento,estado,observaciones,fecha_creacion,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,CAST(CURRENT_TIMESTAMP AS TEXT),?)`, x.EmpresaID, x.UnidadID, nullableID(x.ServicioID), x.Periodo, x.Concepto, x.TipoCargo, x.ValorBase, x.InteresMora, x.Descuento, x.Total, x.SaldoPendiente, x.FechaVencimiento, x.Estado, x.Observaciones, x.UsuarioCreador)
}

func ListEmpresaPropiedadHorizontalCargos(dbConn *sql.DB, empresaID int64) ([]EmpresaPropiedadHorizontalCargo, error) {
	if err := EnsureEmpresaPropiedadHorizontalSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT c.id,c.empresa_id,c.unidad_id,COALESCE(c.servicio_id,0),COALESCE(u.codigo,''),COALESCE(c.periodo,''),COALESCE(c.concepto,''),COALESCE(c.tipo_cargo,''),COALESCE(c.valor_base,0),COALESCE(c.interes_mora,0),COALESCE(c.descuento,0),COALESCE(c.total,0),COALESCE(c.saldo_pendiente,0),COALESCE(c.fecha_vencimiento,''),COALESCE(c.estado,''),COALESCE(c.observaciones,''),COALESCE(c.fecha_creacion,''),COALESCE(c.usuario_creador,'') FROM empresa_propiedad_horizontal_cargos c LEFT JOIN empresa_propiedad_horizontal_unidades u ON u.id=c.unidad_id AND u.empresa_id=c.empresa_id WHERE c.empresa_id=? ORDER BY c.id DESC LIMIT 300`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaPropiedadHorizontalCargo{}
	for rows.Next() {
		var x EmpresaPropiedadHorizontalCargo
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.UnidadID, &x.ServicioID, &x.UnidadCodigo, &x.Periodo, &x.Concepto, &x.TipoCargo, &x.ValorBase, &x.InteresMora, &x.Descuento, &x.Total, &x.SaldoPendiente, &x.FechaVencimiento, &x.Estado, &x.Observaciones, &x.FechaCreacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func CreateEmpresaPropiedadHorizontalRecaudo(dbConn *sql.DB, x EmpresaPropiedadHorizontalRecaudo) (int64, error) {
	if err := EnsureEmpresaPropiedadHorizontalSchema(dbConn); err != nil {
		return 0, err
	}
	x.FechaPago = strings.TrimSpace(x.FechaPago)
	if x.FechaPago == "" {
		x.FechaPago = time.Now().Format("2006-01-02")
	}
	x.MetodoPago = normalizeOneOfPropH(x.MetodoPago, "transferencia", "efectivo", "transferencia", "pse", "wompi", "epayco", "consignacion", "otro")
	x.Referencia, x.Observaciones, x.UsuarioCreador = strings.TrimSpace(x.Referencia), strings.TrimSpace(x.Observaciones), strings.TrimSpace(x.UsuarioCreador)
	if x.EmpresaID <= 0 || x.UnidadID <= 0 || x.ValorPagado <= 0 {
		return 0, errors.New("empresa_id, unidad_id y valor_pagado son obligatorios")
	}
	clienteID, servicioID, err := preparePropHRecaudoCoreRefs(dbConn, x, x.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	x.ClienteID, x.ServicioID = clienteID, servicioID
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_propiedad_horizontal_recaudos (empresa_id,cargo_id,unidad_id,cliente_id,servicio_id,fecha_pago,metodo_pago,referencia,valor_pagado,observaciones,fecha_creacion,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,CAST(CURRENT_TIMESTAMP AS TEXT),?)`, x.EmpresaID, x.CargoID, x.UnidadID, nullableID(x.ClienteID), nullableID(x.ServicioID), x.FechaPago, x.MetodoPago, x.Referencia, x.ValorPagado, x.Observaciones, x.UsuarioCreador)
	if err == nil && x.CargoID > 0 {
		_, _ = ExecCompat(dbConn, `UPDATE empresa_propiedad_horizontal_cargos SET saldo_pendiente=CASE WHEN saldo_pendiente-? < 0 THEN 0 ELSE saldo_pendiente-? END, estado=CASE WHEN saldo_pendiente-? <= 0 THEN 'pagado' ELSE estado END WHERE empresa_id=? AND id=?`, x.ValorPagado, x.ValorPagado, x.ValorPagado, x.EmpresaID, x.CargoID)
	}
	if err == nil {
		x.ID = id
		if carritoID, itemID, clienteID, servicioID, cartErr := createPropHRecaudoCarrito(dbConn, x, x.UsuarioCreador); cartErr == nil && carritoID > 0 {
			_, _ = ExecCompat(dbConn, `UPDATE empresa_propiedad_horizontal_recaudos SET cliente_id=?, servicio_id=?, carrito_id=?, carrito_item_id=? WHERE empresa_id=? AND id=?`, nullableID(clienteID), nullableID(servicioID), nullableID(carritoID), nullableID(itemID), x.EmpresaID, id)
		}
	}
	return id, err
}

func ListEmpresaPropiedadHorizontalRecaudos(dbConn *sql.DB, empresaID int64) ([]EmpresaPropiedadHorizontalRecaudo, error) {
	if err := EnsureEmpresaPropiedadHorizontalSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT r.id,r.empresa_id,COALESCE(r.cargo_id,0),r.unidad_id,COALESCE(r.cliente_id,0),COALESCE(r.servicio_id,0),COALESCE(r.carrito_id,0),COALESCE(r.carrito_item_id,0),COALESCE(u.codigo,''),COALESCE(r.fecha_pago,''),COALESCE(r.metodo_pago,''),COALESCE(r.referencia,''),COALESCE(r.valor_pagado,0),COALESCE(r.observaciones,''),COALESCE(r.fecha_creacion,''),COALESCE(r.usuario_creador,'') FROM empresa_propiedad_horizontal_recaudos r LEFT JOIN empresa_propiedad_horizontal_unidades u ON u.id=r.unidad_id AND u.empresa_id=r.empresa_id WHERE r.empresa_id=? ORDER BY r.id DESC LIMIT 200`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaPropiedadHorizontalRecaudo{}
	for rows.Next() {
		var x EmpresaPropiedadHorizontalRecaudo
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.CargoID, &x.UnidadID, &x.ClienteID, &x.ServicioID, &x.CarritoID, &x.CarritoItemID, &x.UnidadCodigo, &x.FechaPago, &x.MetodoPago, &x.Referencia, &x.ValorPagado, &x.Observaciones, &x.FechaCreacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func UpsertEmpresaPropiedadHorizontalPQR(dbConn *sql.DB, x EmpresaPropiedadHorizontalPQR) (int64, error) {
	if err := EnsureEmpresaPropiedadHorizontalSchema(dbConn); err != nil {
		return 0, err
	}
	x.Tipo = normalizeOneOfPropH(x.Tipo, "peticion", "peticion", "queja", "reclamo", "solicitud", "mantenimiento", "seguridad")
	x.Prioridad = normalizeOneOfPropH(x.Prioridad, "media", "baja", "media", "alta", "critica")
	x.Estado = normalizeOneOfPropH(x.Estado, "abierta", "abierta", "en_proceso", "resuelta", "cerrada")
	x.Asunto, x.Descripcion, x.Responsable, x.FechaLimite, x.UsuarioCreador = strings.TrimSpace(x.Asunto), strings.TrimSpace(x.Descripcion), strings.TrimSpace(x.Responsable), strings.TrimSpace(x.FechaLimite), strings.TrimSpace(x.UsuarioCreador)
	if x.EmpresaID <= 0 || x.Asunto == "" {
		return 0, errors.New("empresa_id y asunto son obligatorios")
	}
	if x.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_propiedad_horizontal_pqrs SET unidad_id=?,tipo=?,prioridad=?,estado=?,asunto=?,descripcion=?,responsable=?,fecha_limite=?,usuario_creador=? WHERE empresa_id=? AND id=?`, x.UnidadID, x.Tipo, x.Prioridad, x.Estado, x.Asunto, x.Descripcion, x.Responsable, x.FechaLimite, x.UsuarioCreador, x.EmpresaID, x.ID)
		return x.ID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_propiedad_horizontal_pqrs (empresa_id,unidad_id,tipo,prioridad,estado,asunto,descripcion,responsable,fecha_limite,fecha_creacion,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,CAST(CURRENT_TIMESTAMP AS TEXT),?)`, x.EmpresaID, x.UnidadID, x.Tipo, x.Prioridad, x.Estado, x.Asunto, x.Descripcion, x.Responsable, x.FechaLimite, x.UsuarioCreador)
}

func ListEmpresaPropiedadHorizontalPQRs(dbConn *sql.DB, empresaID int64) ([]EmpresaPropiedadHorizontalPQR, error) {
	if err := EnsureEmpresaPropiedadHorizontalSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT p.id,p.empresa_id,COALESCE(p.unidad_id,0),COALESCE(u.codigo,''),COALESCE(p.tipo,''),COALESCE(p.prioridad,''),COALESCE(p.estado,''),COALESCE(p.asunto,''),COALESCE(p.descripcion,''),COALESCE(p.responsable,''),COALESCE(p.fecha_limite,''),COALESCE(p.fecha_creacion,''),COALESCE(p.usuario_creador,'') FROM empresa_propiedad_horizontal_pqrs p LEFT JOIN empresa_propiedad_horizontal_unidades u ON u.id=p.unidad_id AND u.empresa_id=p.empresa_id WHERE p.empresa_id=? ORDER BY p.id DESC LIMIT 200`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaPropiedadHorizontalPQR{}
	for rows.Next() {
		var x EmpresaPropiedadHorizontalPQR
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.UnidadID, &x.UnidadCodigo, &x.Tipo, &x.Prioridad, &x.Estado, &x.Asunto, &x.Descripcion, &x.Responsable, &x.FechaLimite, &x.FechaCreacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func UpsertEmpresaPropiedadHorizontalAsamblea(dbConn *sql.DB, x EmpresaPropiedadHorizontalAsamblea) (int64, error) {
	if err := EnsureEmpresaPropiedadHorizontalSchema(dbConn); err != nil {
		return 0, err
	}
	x.Titulo = strings.TrimSpace(x.Titulo)
	x.TipoAsamblea = normalizeOneOfPropH(x.TipoAsamblea, "ordinaria", "ordinaria", "extraordinaria", "consejo", "comite")
	x.Estado = normalizeOneOfPropH(x.Estado, "programada", "programada", "abierta", "cerrada", "cancelada")
	x.Fecha, x.ActaURL, x.Observaciones, x.UsuarioCreador = strings.TrimSpace(x.Fecha), strings.TrimSpace(x.ActaURL), strings.TrimSpace(x.Observaciones), strings.TrimSpace(x.UsuarioCreador)
	if x.EmpresaID <= 0 || x.Titulo == "" {
		return 0, errors.New("empresa_id y titulo son obligatorios")
	}
	if x.QuorumObjetivo <= 0 {
		x.QuorumObjetivo = 51
	}
	if x.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_propiedad_horizontal_asambleas SET titulo=?,tipo_asamblea=?,fecha=?,estado=?,quorum_objetivo=?,quorum_actual=?,acta_url=?,observaciones=?,usuario_creador=? WHERE empresa_id=? AND id=?`, x.Titulo, x.TipoAsamblea, x.Fecha, x.Estado, x.QuorumObjetivo, x.QuorumActual, x.ActaURL, x.Observaciones, x.UsuarioCreador, x.EmpresaID, x.ID)
		return x.ID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_propiedad_horizontal_asambleas (empresa_id,titulo,tipo_asamblea,fecha,estado,quorum_objetivo,quorum_actual,acta_url,observaciones,fecha_creacion,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,CAST(CURRENT_TIMESTAMP AS TEXT),?)`, x.EmpresaID, x.Titulo, x.TipoAsamblea, x.Fecha, x.Estado, x.QuorumObjetivo, x.QuorumActual, x.ActaURL, x.Observaciones, x.UsuarioCreador)
}

func ListEmpresaPropiedadHorizontalAsambleas(dbConn *sql.DB, empresaID int64) ([]EmpresaPropiedadHorizontalAsamblea, error) {
	if err := EnsureEmpresaPropiedadHorizontalSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,COALESCE(titulo,''),COALESCE(tipo_asamblea,''),COALESCE(fecha,''),COALESCE(estado,''),COALESCE(quorum_objetivo,51),COALESCE(quorum_actual,0),COALESCE(acta_url,''),COALESCE(observaciones,''),COALESCE(fecha_creacion,''),COALESCE(usuario_creador,'') FROM empresa_propiedad_horizontal_asambleas WHERE empresa_id=? ORDER BY id DESC LIMIT 100`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaPropiedadHorizontalAsamblea{}
	for rows.Next() {
		var x EmpresaPropiedadHorizontalAsamblea
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Titulo, &x.TipoAsamblea, &x.Fecha, &x.Estado, &x.QuorumObjetivo, &x.QuorumActual, &x.ActaURL, &x.Observaciones, &x.FechaCreacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func BuildEmpresaPropiedadHorizontalDashboard(dbConn *sql.DB, empresaID int64) (EmpresaPropiedadHorizontalDashboard, error) {
	cfg, err := GetEmpresaPropiedadHorizontalConfig(dbConn, empresaID)
	if err != nil {
		return EmpresaPropiedadHorizontalDashboard{}, err
	}
	unidades, err := ListEmpresaPropiedadHorizontalUnidades(dbConn, empresaID)
	if err != nil {
		return EmpresaPropiedadHorizontalDashboard{}, err
	}
	personas, err := ListEmpresaPropiedadHorizontalPersonas(dbConn, empresaID)
	if err != nil {
		return EmpresaPropiedadHorizontalDashboard{}, err
	}
	cargos, err := ListEmpresaPropiedadHorizontalCargos(dbConn, empresaID)
	if err != nil {
		return EmpresaPropiedadHorizontalDashboard{}, err
	}
	recaudos, err := ListEmpresaPropiedadHorizontalRecaudos(dbConn, empresaID)
	if err != nil {
		return EmpresaPropiedadHorizontalDashboard{}, err
	}
	pqrs, err := ListEmpresaPropiedadHorizontalPQRs(dbConn, empresaID)
	if err != nil {
		return EmpresaPropiedadHorizontalDashboard{}, err
	}
	asambleas, err := ListEmpresaPropiedadHorizontalAsambleas(dbConn, empresaID)
	if err != nil {
		return EmpresaPropiedadHorizontalDashboard{}, err
	}
	out := EmpresaPropiedadHorizontalDashboard{EmpresaID: empresaID, Config: cfg, Unidades: unidades, Personas: personas, Cargos: cargos, Recaudos: recaudos, PQRs: pqrs, Asambleas: asambleas}
	out.TotalUnidades = len(unidades)
	for _, u := range unidades {
		if u.Estado == "ocupada" {
			out.UnidadesOcupadas++
		}
	}
	for _, p := range personas {
		if p.Estado == "activo" {
			out.ResidentesActivos++
		}
	}
	for _, c := range cargos {
		if c.Estado != "pagado" && c.Estado != "anulado" {
			out.CarteraPendiente += c.SaldoPendiente
		}
	}
	currentMonth := time.Now().Format("2006-01")
	for _, r := range recaudos {
		if strings.HasPrefix(r.FechaPago, currentMonth) {
			out.RecaudoMes += r.ValorPagado
		}
	}
	for _, p := range pqrs {
		if p.Estado == "abierta" || p.Estado == "en_proceso" {
			out.PQRPendientes++
		}
	}
	for _, a := range asambleas {
		if a.Estado == "programada" || a.Estado == "abierta" {
			out.AsambleasAbiertas++
		}
	}
	return out, nil
}

func SeedEmpresaPropiedadHorizontalDemo(dbConn *sql.DB, empresaID int64, usuario string) error {
	if err := UpsertEmpresaPropiedadHorizontalConfig(dbConn, EmpresaPropiedadHorizontalConfig{EmpresaID: empresaID, NombreCopropiedad: "Conjunto Residencial Camino Real", NIT: "900123456-7", TipoCopropiedad: "conjunto_residencial", Direccion: "Calle 10 # 20-30", Ciudad: "Bogota", Administrador: "Administracion Principal", Telefono: "3001234567", Email: "admin@copropiedad.com", InteresMoraMensual: 2, DiasGracia: 5, FacturacionElectronica: true, PermitirPortalResidente: true, UsuarioCreador: usuario}); err != nil {
		return err
	}
	u1, err := UpsertEmpresaPropiedadHorizontalUnidad(dbConn, EmpresaPropiedadHorizontalUnidad{EmpresaID: empresaID, Codigo: "T1-301", Torre: "Torre 1", Piso: "3", TipoUnidad: "apartamento", AreaM2: 72, Coeficiente: 1.25, CuotaBase: 385000, Parqueadero: "P-45", Deposito: "D-12", Estado: "ocupada", UsuarioCreador: usuario})
	if err != nil {
		return err
	}
	u2, err := UpsertEmpresaPropiedadHorizontalUnidad(dbConn, EmpresaPropiedadHorizontalUnidad{EmpresaID: empresaID, Codigo: "T2-502", Torre: "Torre 2", Piso: "5", TipoUnidad: "apartamento", AreaM2: 88, Coeficiente: 1.48, CuotaBase: 455000, Parqueadero: "P-62", Estado: "ocupada", UsuarioCreador: usuario})
	if err != nil {
		return err
	}
	_, _ = UpsertEmpresaPropiedadHorizontalPersona(dbConn, EmpresaPropiedadHorizontalPersona{EmpresaID: empresaID, UnidadID: u1, TipoRelacion: "propietario", Nombre: "Laura Mejia", Documento: "52123456", Telefono: "3105551111", Email: "laura@example.com", Estado: "activo", UsuarioCreador: usuario})
	_, _ = UpsertEmpresaPropiedadHorizontalPersona(dbConn, EmpresaPropiedadHorizontalPersona{EmpresaID: empresaID, UnidadID: u2, TipoRelacion: "arrendatario", Nombre: "Carlos Gomez", Documento: "80111222", Telefono: "3205552222", Email: "carlos@example.com", Estado: "activo", UsuarioCreador: usuario})
	period := time.Now().Format("2006-01")
	c1, _ := CreateEmpresaPropiedadHorizontalCargo(dbConn, EmpresaPropiedadHorizontalCargo{EmpresaID: empresaID, UnidadID: u1, Periodo: period, Concepto: "Cuota administracion", TipoCargo: "cuota_administracion", ValorBase: 385000, FechaVencimiento: period + "-10", UsuarioCreador: usuario})
	_, _ = CreateEmpresaPropiedadHorizontalCargo(dbConn, EmpresaPropiedadHorizontalCargo{EmpresaID: empresaID, UnidadID: u2, Periodo: period, Concepto: "Cuota administracion", TipoCargo: "cuota_administracion", ValorBase: 455000, InteresMora: 25000, FechaVencimiento: period + "-10", UsuarioCreador: usuario})
	_, _ = CreateEmpresaPropiedadHorizontalRecaudo(dbConn, EmpresaPropiedadHorizontalRecaudo{EmpresaID: empresaID, CargoID: c1, UnidadID: u1, MetodoPago: "transferencia", Referencia: "TRX-DEMO-001", ValorPagado: 385000, UsuarioCreador: usuario})
	_, _ = UpsertEmpresaPropiedadHorizontalPQR(dbConn, EmpresaPropiedadHorizontalPQR{EmpresaID: empresaID, UnidadID: u2, Tipo: "mantenimiento", Prioridad: "alta", Asunto: "Humedad en zona comun", Descripcion: "Reporte de humedad cerca al shut de basuras.", Responsable: "Mantenimiento", FechaLimite: time.Now().AddDate(0, 0, 5).Format("2006-01-02"), UsuarioCreador: usuario})
	_, _ = UpsertEmpresaPropiedadHorizontalAsamblea(dbConn, EmpresaPropiedadHorizontalAsamblea{EmpresaID: empresaID, Titulo: "Asamblea ordinaria anual", TipoAsamblea: "ordinaria", Fecha: time.Now().AddDate(0, 1, 0).Format("2006-01-02"), Estado: "programada", QuorumObjetivo: 51, QuorumActual: 0, UsuarioCreador: usuario})
	return nil
}

func propHCoreCode(prefix string, parts ...string) string {
	var b strings.Builder
	for _, part := range parts {
		for _, r := range strings.ToUpper(strings.TrimSpace(part)) {
			switch r {
			case '\u00c1', '\u00c0', '\u00c4', '\u00c2':
				r = 'A'
			case '\u00c9', '\u00c8', '\u00cb', '\u00ca':
				r = 'E'
			case '\u00cd', '\u00cc', '\u00cf', '\u00ce':
				r = 'I'
			case '\u00d3', '\u00d2', '\u00d6', '\u00d4':
				r = 'O'
			case '\u00da', '\u00d9', '\u00dc', '\u00db':
				r = 'U'
			case '\u00d1':
				r = 'N'
			}
			if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				b.WriteRune(r)
				continue
			}
			if b.Len() > 0 && b.String()[b.Len()-1] != '-' {
				b.WriteRune('-')
			}
		}
		if b.Len() > 0 && b.String()[b.Len()-1] != '-' {
			b.WriteRune('-')
		}
	}
	code := strings.Trim(b.String(), "-")
	if code == "" {
		code = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if len(code) > 42 {
		code = code[:42]
	}
	var prefixBuilder strings.Builder
	for _, r := range strings.ToUpper(strings.TrimSpace(prefix)) {
		switch r {
		case '\u00c1', '\u00c0', '\u00c4', '\u00c2':
			r = 'A'
		case '\u00c9', '\u00c8', '\u00cb', '\u00ca':
			r = 'E'
		case '\u00cd', '\u00cc', '\u00cf', '\u00ce':
			r = 'I'
		case '\u00d3', '\u00d2', '\u00d6', '\u00d4':
			r = 'O'
		case '\u00da', '\u00d9', '\u00dc', '\u00db':
			r = 'U'
		case '\u00d1':
			r = 'N'
		}
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			prefixBuilder.WriteRune(r)
			continue
		}
		if prefixBuilder.Len() > 0 && prefixBuilder.String()[prefixBuilder.Len()-1] != '-' {
			prefixBuilder.WriteRune('-')
		}
	}
	prefixCode := strings.Trim(prefixBuilder.String(), "-")
	if prefixCode == "" {
		prefixCode = "PH"
	}
	return prefixCode + "-" + strings.Trim(code, "-")
}

func ensurePropHUnidadServicio(dbConn *sql.DB, unidad EmpresaPropiedadHorizontalUnidad, usuario string) (int64, error) {
	if unidad.ServicioID > 0 {
		return unidad.ServicioID, nil
	}
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		return 0, err
	}
	code := propHCoreCode("PH-UNIDAD", unidad.Codigo)
	var servicioID int64
	err := QueryRowCompat(dbConn, `SELECT id FROM servicios WHERE empresa_id=? AND codigo=? LIMIT 1`, unidad.EmpresaID, code).Scan(&servicioID)
	if err == nil {
		_, _ = ExecCompat(dbConn, `UPDATE servicios SET nombre=?, descripcion=?, categoria='propiedad_horizontal', precio=?, estado='activo', fecha_actualizacion=? WHERE empresa_id=? AND id=?`, "Cuota base "+unidad.Codigo, strings.TrimSpace(unidad.TipoUnidad+" "+unidad.Torre+" "+unidad.Piso), unidad.CuotaBase, time.Now().Format("2006-01-02 15:04:05"), unidad.EmpresaID, servicioID)
		return servicioID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	return CreateServicio(dbConn, Servicio{
		EmpresaID:      unidad.EmpresaID,
		Codigo:         code,
		Nombre:         "Cuota base " + unidad.Codigo,
		Descripcion:    strings.TrimSpace("Servicio base de propiedad horizontal para " + unidad.TipoUnidad + " " + unidad.Torre + " " + unidad.Piso),
		Categoria:      "propiedad_horizontal",
		Precio:         unidad.CuotaBase,
		Estado:         "activo",
		UsuarioCreador: strings.TrimSpace(usuario),
		Observaciones:  "Servicio sincronizado desde propiedad horizontal.",
	})
}

func ensurePropHCargoServicio(dbConn *sql.DB, cargo EmpresaPropiedadHorizontalCargo, usuario string) (int64, error) {
	if cargo.ServicioID > 0 {
		return cargo.ServicioID, nil
	}
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		return 0, err
	}
	code := propHCoreCode("PH-CARGO", cargo.TipoCargo, cargo.Concepto)
	var servicioID int64
	err := QueryRowCompat(dbConn, `SELECT id FROM servicios WHERE empresa_id=? AND codigo=? LIMIT 1`, cargo.EmpresaID, code).Scan(&servicioID)
	if err == nil {
		_, _ = ExecCompat(dbConn, `UPDATE servicios SET nombre=?, descripcion=?, categoria='propiedad_horizontal', precio=?, estado='activo', fecha_actualizacion=? WHERE empresa_id=? AND id=?`, cargo.Concepto, cargo.TipoCargo, cargo.ValorBase, time.Now().Format("2006-01-02 15:04:05"), cargo.EmpresaID, servicioID)
		return servicioID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	return CreateServicio(dbConn, Servicio{
		EmpresaID:      cargo.EmpresaID,
		Codigo:         code,
		Nombre:         cargo.Concepto,
		Descripcion:    "Cargo de propiedad horizontal: " + cargo.TipoCargo,
		Categoria:      "propiedad_horizontal",
		Precio:         cargo.ValorBase,
		Estado:         "activo",
		UsuarioCreador: strings.TrimSpace(usuario),
		Observaciones:  "Servicio sincronizado desde cargos de propiedad horizontal.",
	})
}

func ensurePropHPersonaClienteCore(dbConn *sql.DB, persona EmpresaPropiedadHorizontalPersona, usuario string) (int64, error) {
	if persona.ClienteID > 0 {
		return persona.ClienteID, nil
	}
	if strings.TrimSpace(persona.Nombre) == "" && strings.TrimSpace(persona.Documento) == "" && strings.TrimSpace(persona.Telefono) == "" && strings.TrimSpace(persona.Email) == "" {
		return 0, nil
	}
	if err := EnsureEmpresaClientesSchema(dbConn); err != nil {
		return 0, err
	}
	if documentoNorm := normalizeClienteDocumentoValue(persona.Documento); documentoNorm != "" {
		query := fmt.Sprintf(`SELECT id FROM clientes WHERE empresa_id = ? AND %s = ? LIMIT 1`, clienteDocumentoSQLExpr("numero_documento"))
		if id, err := findClienteDuplicateID(dbConn, query, persona.EmpresaID, documentoNorm); err != nil {
			return 0, err
		} else if id > 0 {
			return id, nil
		}
	}
	if telefonoNorm := normalizeClienteTelefonoValue(persona.Telefono); telefonoNorm != "" {
		query := fmt.Sprintf(`SELECT id FROM clientes WHERE empresa_id = ? AND %s = ? LIMIT 1`, clienteTelefonoSQLExpr("telefono"))
		if id, err := findClienteDuplicateID(dbConn, query, persona.EmpresaID, telefonoNorm); err != nil {
			return 0, err
		} else if id > 0 {
			return id, nil
		}
	}
	if emailNorm := normalizeClienteEmailValue(persona.Email); emailNorm != "" {
		if id, err := findClienteDuplicateID(dbConn, `SELECT id FROM clientes WHERE empresa_id = ? AND lower(trim(COALESCE(email, ''))) = ? LIMIT 1`, persona.EmpresaID, emailNorm); err != nil {
			return 0, err
		} else if id > 0 {
			return id, nil
		}
	}
	tipoDocumento := "CC"
	numeroDocumento := strings.TrimSpace(persona.Documento)
	if numeroDocumento == "" {
		tipoDocumento = "OTRO"
		numeroDocumento = propHCoreCode("PH-CLI", persona.UnidadCodigo, persona.Telefono, persona.Email, persona.Nombre)
	}
	nombre := strings.TrimSpace(persona.Nombre)
	if nombre == "" {
		nombre = "Residente propiedad horizontal"
	}
	id, err := CreateCliente(dbConn, Cliente{
		EmpresaID:         persona.EmpresaID,
		TipoDocumento:     tipoDocumento,
		NumeroDocumento:   numeroDocumento,
		TipoPersona:       "natural",
		NombreRazonSocial: nombre,
		NombreComercial:   nombre,
		Email:             strings.TrimSpace(persona.Email),
		Telefono:          strings.TrimSpace(persona.Telefono),
		Pais:              "CO",
		UsuarioCreador:    strings.TrimSpace(usuario),
		Estado:            "activo",
		Observaciones:     "Cliente creado/sincronizado desde propiedad horizontal.",
	})
	if err != nil {
		var dup *ClienteDuplicadoError
		if errors.As(err, &dup) && dup.ClienteID > 0 {
			return dup.ClienteID, nil
		}
		return 0, err
	}
	return id, nil
}

func findPropHClienteByUnidad(dbConn *sql.DB, empresaID, unidadID int64, usuario string) (int64, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,COALESCE(unidad_id,0),COALESCE(cliente_id,0),COALESCE(tipo_relacion,''),COALESCE(nombre,''),COALESCE(documento,''),COALESCE(telefono,''),COALESCE(email,''),COALESCE(estado,'') FROM empresa_propiedad_horizontal_personas WHERE empresa_id=? AND unidad_id=? AND estado='activo' ORDER BY CASE tipo_relacion WHEN 'propietario' THEN 1 WHEN 'arrendatario' THEN 2 WHEN 'residente' THEN 3 ELSE 4 END, id DESC LIMIT 1`, empresaID, unidadID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	if rows.Next() {
		var p EmpresaPropiedadHorizontalPersona
		if err := rows.Scan(&p.ID, &p.EmpresaID, &p.UnidadID, &p.ClienteID, &p.TipoRelacion, &p.Nombre, &p.Documento, &p.Telefono, &p.Email, &p.Estado); err != nil {
			return 0, err
		}
		clienteID, err := ensurePropHPersonaClienteCore(dbConn, p, usuario)
		if err != nil {
			return 0, err
		}
		if clienteID > 0 {
			_, _ = ExecCompat(dbConn, `UPDATE empresa_propiedad_horizontal_personas SET cliente_id=? WHERE empresa_id=? AND id=?`, nullableID(clienteID), empresaID, p.ID)
			return clienteID, nil
		}
	}
	var unidad EmpresaPropiedadHorizontalUnidad
	err = QueryRowCompat(dbConn, `SELECT id,empresa_id,COALESCE(codigo,''),COALESCE(torre,''),COALESCE(piso,'') FROM empresa_propiedad_horizontal_unidades WHERE empresa_id=? AND id=? LIMIT 1`, empresaID, unidadID).
		Scan(&unidad.ID, &unidad.EmpresaID, &unidad.Codigo, &unidad.Torre, &unidad.Piso)
	if err != nil {
		return 0, err
	}
	return ensurePropHPersonaClienteCore(dbConn, EmpresaPropiedadHorizontalPersona{EmpresaID: empresaID, UnidadID: unidadID, UnidadCodigo: unidad.Codigo, Nombre: "Unidad " + unidad.Codigo, Documento: propHCoreCode("PH-UNIDAD", unidad.Codigo)}, usuario)
}

func getPropHCargoByID(dbConn *sql.DB, empresaID, cargoID int64) (EmpresaPropiedadHorizontalCargo, error) {
	var x EmpresaPropiedadHorizontalCargo
	err := QueryRowCompat(dbConn, `SELECT c.id,c.empresa_id,c.unidad_id,COALESCE(c.servicio_id,0),COALESCE(u.codigo,''),COALESCE(c.periodo,''),COALESCE(c.concepto,''),COALESCE(c.tipo_cargo,''),COALESCE(c.valor_base,0),COALESCE(c.interes_mora,0),COALESCE(c.descuento,0),COALESCE(c.total,0),COALESCE(c.saldo_pendiente,0),COALESCE(c.fecha_vencimiento,''),COALESCE(c.estado,''),COALESCE(c.observaciones,''),COALESCE(c.fecha_creacion,''),COALESCE(c.usuario_creador,'') FROM empresa_propiedad_horizontal_cargos c LEFT JOIN empresa_propiedad_horizontal_unidades u ON u.id=c.unidad_id AND u.empresa_id=c.empresa_id WHERE c.empresa_id=? AND c.id=?`, empresaID, cargoID).
		Scan(&x.ID, &x.EmpresaID, &x.UnidadID, &x.ServicioID, &x.UnidadCodigo, &x.Periodo, &x.Concepto, &x.TipoCargo, &x.ValorBase, &x.InteresMora, &x.Descuento, &x.Total, &x.SaldoPendiente, &x.FechaVencimiento, &x.Estado, &x.Observaciones, &x.FechaCreacion, &x.UsuarioCreador)
	return x, err
}

func preparePropHRecaudoCoreRefs(dbConn *sql.DB, recaudo EmpresaPropiedadHorizontalRecaudo, usuario string) (int64, int64, error) {
	clienteID, err := findPropHClienteByUnidad(dbConn, recaudo.EmpresaID, recaudo.UnidadID, usuario)
	if err != nil {
		return 0, 0, err
	}
	var cargo EmpresaPropiedadHorizontalCargo
	if recaudo.CargoID > 0 {
		cargo, err = getPropHCargoByID(dbConn, recaudo.EmpresaID, recaudo.CargoID)
		if err != nil {
			return 0, 0, err
		}
	} else {
		cargo = EmpresaPropiedadHorizontalCargo{EmpresaID: recaudo.EmpresaID, UnidadID: recaudo.UnidadID, Concepto: "Recaudo propiedad horizontal", TipoCargo: "otro", ValorBase: recaudo.ValorPagado, Total: recaudo.ValorPagado}
	}
	servicioID, err := ensurePropHCargoServicio(dbConn, cargo, usuario)
	if err != nil {
		return 0, 0, err
	}
	if recaudo.CargoID > 0 && servicioID > 0 {
		_, _ = ExecCompat(dbConn, `UPDATE empresa_propiedad_horizontal_cargos SET servicio_id=? WHERE empresa_id=? AND id=?`, nullableID(servicioID), recaudo.EmpresaID, recaudo.CargoID)
	}
	return clienteID, servicioID, nil
}

func propHMetodoPagoCarrito(metodo string) string {
	normalized := NormalizeMetodoPagoCarrito(metodo)
	if normalized != "" {
		return normalized
	}
	switch strings.ToLower(strings.TrimSpace(metodo)) {
	case "pse", "wompi", "epayco", "consignacion", "otro":
		return "transferencia_bancaria"
	default:
		return "efectivo"
	}
}

func createPropHRecaudoCarrito(dbConn *sql.DB, recaudo EmpresaPropiedadHorizontalRecaudo, usuario string) (int64, int64, int64, int64, error) {
	if recaudo.ValorPagado <= 0 {
		return recaudo.CarritoID, recaudo.CarritoItemID, recaudo.ClienteID, recaudo.ServicioID, nil
	}
	if err := EnsureEmpresaCarritosSchema(dbConn); err != nil {
		return 0, 0, 0, 0, err
	}
	clienteID, servicioID, err := preparePropHRecaudoCoreRefs(dbConn, recaudo, usuario)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	referencia := strings.TrimSpace(recaudo.Referencia)
	if referencia == "" {
		referencia = fmt.Sprintf("RECAUDO-%d", recaudo.ID)
	}
	referenciaExterna := fmt.Sprintf("propiedad_horizontal:recaudo:%d:%s", recaudo.ID, referencia)
	var carritoExistente int64
	err = QueryRowCompat(dbConn, `SELECT id FROM carritos_compras WHERE empresa_id=? AND referencia_externa=? LIMIT 1`, recaudo.EmpresaID, referenciaExterna).Scan(&carritoExistente)
	if err == nil && carritoExistente > 0 {
		return carritoExistente, recaudo.CarritoItemID, clienteID, servicioID, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, 0, 0, 0, err
	}
	metodo := propHMetodoPagoCarrito(recaudo.MetodoPago)
	carritoID, err := CreateCarritoCompra(dbConn, CarritoCompra{
		EmpresaID:         recaudo.EmpresaID,
		Codigo:            propHCoreCode("PH-REC", fmt.Sprintf("%d", recaudo.ID), referencia),
		Nombre:            "Recaudo propiedad horizontal " + recaudo.UnidadCodigo,
		CanalVenta:        "propiedad_horizontal",
		ClienteID:         clienteID,
		EstadoCarrito:     "abierto",
		Moneda:            "COP",
		ReferenciaExterna: referenciaExterna,
		MetodoPago:        metodo,
		ReferenciaPago:    referencia,
		UsuarioCreador:    strings.TrimSpace(usuario),
		Observaciones:     "Venta central generada desde recaudo de propiedad horizontal.",
	})
	if err != nil {
		return 0, 0, 0, 0, err
	}
	descripcion := "Recaudo propiedad horizontal"
	if recaudo.CargoID > 0 {
		if cargo, err := getPropHCargoByID(dbConn, recaudo.EmpresaID, recaudo.CargoID); err == nil && strings.TrimSpace(cargo.Concepto) != "" {
			descripcion = cargo.Concepto
		}
	}
	itemID, err := CreateCarritoCompraItem(dbConn, CarritoCompraItem{
		EmpresaID:          recaudo.EmpresaID,
		CarritoID:          carritoID,
		TipoItem:           "servicio",
		ReferenciaID:       servicioID,
		CodigoItem:         propHCoreCode("PH-ITEM", referencia),
		Descripcion:        descripcion,
		UnidadMedida:       "servicio",
		Cantidad:           1,
		PrecioUnitario:     recaudo.ValorPagado,
		ImpuestoPorcentaje: 0,
		UsuarioCreador:     strings.TrimSpace(usuario),
		Estado:             "activo",
		Observaciones:      recaudo.Observaciones,
	})
	if err != nil {
		return 0, 0, 0, 0, err
	}
	if err := PayCarritoStationSession(dbConn, recaudo.EmpresaID, carritoID, metodo, referencia, "", "", 0, 0, recaudo.ValorPagado, 0, strings.TrimSpace(usuario)); err != nil {
		return 0, 0, 0, 0, err
	}
	return carritoID, itemID, clienteID, servicioID, nil
}

func normalizePropiedadHorizontalConfig(x EmpresaPropiedadHorizontalConfig) EmpresaPropiedadHorizontalConfig {
	x.NombreCopropiedad = strings.TrimSpace(x.NombreCopropiedad)
	if x.NombreCopropiedad == "" {
		x.NombreCopropiedad = "Copropiedad empresarial"
	}
	x.NIT, x.Direccion, x.Ciudad, x.Administrador, x.Telefono, x.Email, x.UsuarioCreador = strings.TrimSpace(x.NIT), strings.TrimSpace(x.Direccion), strings.TrimSpace(x.Ciudad), strings.TrimSpace(x.Administrador), strings.TrimSpace(x.Telefono), strings.TrimSpace(x.Email), strings.TrimSpace(x.UsuarioCreador)
	x.TipoCopropiedad = normalizeOneOfPropH(x.TipoCopropiedad, "conjunto_residencial", "conjunto_residencial", "edificio", "condominio", "centro_comercial", "mixto", "colegio")
	if x.InteresMoraMensual < 0 {
		x.InteresMoraMensual = 0
	}
	if x.InteresMoraMensual > 100 {
		x.InteresMoraMensual = 100
	}
	if x.DiasGracia < 0 {
		x.DiasGracia = 0
	}
	return x
}

func normalizePropiedadHorizontalUnidad(x EmpresaPropiedadHorizontalUnidad) EmpresaPropiedadHorizontalUnidad {
	x.Codigo = strings.ToUpper(strings.TrimSpace(x.Codigo))
	x.Torre, x.Piso, x.Parqueadero, x.Deposito, x.Observaciones, x.UsuarioCreador = strings.TrimSpace(x.Torre), strings.TrimSpace(x.Piso), strings.TrimSpace(x.Parqueadero), strings.TrimSpace(x.Deposito), strings.TrimSpace(x.Observaciones), strings.TrimSpace(x.UsuarioCreador)
	x.TipoUnidad = normalizeOneOfPropH(x.TipoUnidad, "apartamento", "apartamento", "casa", "local", "oficina", "parqueadero", "deposito", "zona_comun")
	x.Estado = normalizeOneOfPropH(x.Estado, "ocupada", "ocupada", "desocupada", "en_mora", "inactiva")
	if x.AreaM2 < 0 {
		x.AreaM2 = 0
	}
	if x.Coeficiente < 0 {
		x.Coeficiente = 0
	}
	if x.CuotaBase < 0 {
		x.CuotaBase = 0
	}
	return x
}

func normalizePropiedadHorizontalPersona(x EmpresaPropiedadHorizontalPersona) EmpresaPropiedadHorizontalPersona {
	x.Nombre, x.Documento, x.Telefono, x.Email, x.ContactoEmergencia, x.Observaciones, x.UsuarioCreador = strings.TrimSpace(x.Nombre), strings.TrimSpace(x.Documento), strings.TrimSpace(x.Telefono), strings.TrimSpace(x.Email), strings.TrimSpace(x.ContactoEmergencia), strings.TrimSpace(x.Observaciones), strings.TrimSpace(x.UsuarioCreador)
	x.TipoRelacion = normalizeOneOfPropH(x.TipoRelacion, "propietario", "propietario", "residente", "arrendatario", "apoderado", "visitante")
	x.Estado = normalizeOneOfPropH(x.Estado, "activo", "activo", "inactivo", "moroso")
	return x
}

func normalizePropiedadHorizontalCargo(x EmpresaPropiedadHorizontalCargo) EmpresaPropiedadHorizontalCargo {
	x.Periodo = normalizePropHPeriodo(x.Periodo)
	if x.Periodo == "" {
		x.Periodo = time.Now().Format("2006-01")
	}
	x.Concepto, x.FechaVencimiento, x.Observaciones, x.UsuarioCreador = strings.TrimSpace(x.Concepto), strings.TrimSpace(x.FechaVencimiento), strings.TrimSpace(x.Observaciones), strings.TrimSpace(x.UsuarioCreador)
	x.TipoCargo = normalizeOneOfPropH(x.TipoCargo, "cuota_administracion", "cuota_administracion", "cuota_extraordinaria", "multa", "interes_mora", "reserva_zona", "otro")
	x.Estado = normalizeOneOfPropH(x.Estado, "pendiente", "pendiente", "pagado", "parcial", "vencido", "anulado")
	x.Total = x.ValorBase + x.InteresMora - x.Descuento
	if x.Total < 0 {
		x.Total = 0
	}
	if x.SaldoPendiente <= 0 && x.Estado != "pagado" {
		x.SaldoPendiente = x.Total
	}
	return x
}

func normalizePropHPeriodo(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 7 {
		return v[:7]
	}
	return v
}

func normalizeOneOfPropH(v, fallback string, allowed ...string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	v = strings.ReplaceAll(v, " ", "_")
	for _, a := range allowed {
		if v == a {
			return v
		}
	}
	return fallback
}

func boolIntPropH(v bool) int {
	if v {
		return 1
	}
	return 0
}
