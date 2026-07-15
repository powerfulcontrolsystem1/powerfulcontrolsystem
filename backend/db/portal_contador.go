package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type EmpresaPortalContadorCliente struct {
	ID                   int64  `json:"id"`
	EmpresaID            int64  `json:"empresa_id"`
	ClienteEmpresaID     int64  `json:"cliente_empresa_id"`
	Codigo               string `json:"codigo"`
	RazonSocial          string `json:"razon_social"`
	NIT                  string `json:"nit"`
	Regimen              string `json:"regimen"`
	Responsable          string `json:"responsable"`
	ContactoNombre       string `json:"contacto_nombre"`
	ContactoEmail        string `json:"contacto_email"`
	ContactoTelefono     string `json:"contacto_telefono"`
	Periodicidad         string `json:"periodicidad"`
	FechaInicio          string `json:"fecha_inicio"`
	CierreMensualDia     int    `json:"cierre_mensual_dia"`
	EstadoCliente        string `json:"estado_cliente"`
	RiesgoNivel          string `json:"riesgo_nivel"`
	Usuario              string `json:"usuario_creador"`
	Estado               string `json:"estado"`
	Observaciones        string `json:"observaciones"`
	FechaCreacion        string `json:"fecha_creacion"`
	ObligacionesAbiertas int    `json:"obligaciones_abiertas,omitempty"`
	SolicitudesAbiertas  int    `json:"solicitudes_abiertas,omitempty"`
}

type EmpresaPortalContadorObligacion struct {
	ID                int64   `json:"id"`
	EmpresaID         int64   `json:"empresa_id"`
	ClienteID         int64   `json:"cliente_id"`
	Codigo            string  `json:"codigo"`
	Tipo              string  `json:"tipo"`
	Periodo           string  `json:"periodo"`
	FechaVencimiento  string  `json:"fecha_vencimiento"`
	EstadoObligacion  string  `json:"estado_obligacion"`
	Prioridad         string  `json:"prioridad"`
	ValorEstimado     float64 `json:"valor_estimado"`
	Responsable       string  `json:"responsable"`
	FechaPresentacion string  `json:"fecha_presentacion"`
	SoporteURL        string  `json:"soporte_url"`
	Usuario           string  `json:"usuario_creador"`
	Observaciones     string  `json:"observaciones"`
	FechaCreacion     string  `json:"fecha_creacion"`
	ClienteNombre     string  `json:"cliente_nombre,omitempty"`
	DiasParaVencer    int     `json:"dias_para_vencer,omitempty"`
}

type EmpresaPortalContadorSolicitud struct {
	ID              int64  `json:"id"`
	EmpresaID       int64  `json:"empresa_id"`
	ClienteID       int64  `json:"cliente_id"`
	Codigo          string `json:"codigo"`
	Titulo          string `json:"titulo"`
	Categoria       string `json:"categoria"`
	FechaSolicitud  string `json:"fecha_solicitud"`
	FechaLimite     string `json:"fecha_limite"`
	EstadoSolicitud string `json:"estado_solicitud"`
	Prioridad       string `json:"prioridad"`
	Responsable     string `json:"responsable"`
	Respuesta       string `json:"respuesta"`
	SoporteURL      string `json:"soporte_url"`
	Usuario         string `json:"usuario_creador"`
	Observaciones   string `json:"observaciones"`
	FechaCreacion   string `json:"fecha_creacion"`
	ClienteNombre   string `json:"cliente_nombre,omitempty"`
	DiasParaVencer  int    `json:"dias_para_vencer,omitempty"`
}

type EmpresaPortalContadorComunicacion struct {
	ID            int64  `json:"id"`
	EmpresaID     int64  `json:"empresa_id"`
	ClienteID     int64  `json:"cliente_id"`
	Canal         string `json:"canal"`
	Asunto        string `json:"asunto"`
	Mensaje       string `json:"mensaje"`
	FechaMensaje  string `json:"fecha_mensaje"`
	LeidoCliente  bool   `json:"leido_cliente"`
	Usuario       string `json:"usuario_creador"`
	Observaciones string `json:"observaciones"`
	FechaCreacion string `json:"fecha_creacion"`
	ClienteNombre string `json:"cliente_nombre,omitempty"`
}

type EmpresaPortalContadorDashboard struct {
	EmpresaID                int64                               `json:"empresa_id"`
	ClientesActivos          int                                 `json:"clientes_activos"`
	ClientesRiesgoAlto       int                                 `json:"clientes_riesgo_alto"`
	ObligacionesPendientes   int                                 `json:"obligaciones_pendientes"`
	ObligacionesVencen7Dias  int                                 `json:"obligaciones_vencen_7_dias"`
	SolicitudesAbiertas      int                                 `json:"solicitudes_abiertas"`
	SolicitudesVencidas      int                                 `json:"solicitudes_vencidas"`
	ComunicacionesMes        int                                 `json:"comunicaciones_mes"`
	Clientes                 []EmpresaPortalContadorCliente      `json:"clientes"`
	ObligacionesPrioritarias []EmpresaPortalContadorObligacion   `json:"obligaciones_prioritarias"`
	SolicitudesPrioritarias  []EmpresaPortalContadorSolicitud    `json:"solicitudes_prioritarias"`
	UltimasComunicaciones    []EmpresaPortalContadorComunicacion `json:"ultimas_comunicaciones"`
	Alertas                  []string                            `json:"alertas"`
}

func EnsureEmpresaPortalContadorSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_portal_contador_clientes (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			cliente_empresa_id INTEGER DEFAULT 0,
			codigo TEXT NOT NULL,
			razon_social TEXT NOT NULL,
			nit TEXT,
			regimen TEXT DEFAULT 'responsable_iva',
			responsable TEXT,
			contacto_nombre TEXT,
			contacto_email TEXT,
			contacto_telefono TEXT,
			periodicidad TEXT DEFAULT 'mensual',
			fecha_inicio TEXT,
			cierre_mensual_dia INTEGER DEFAULT 10,
			estado_cliente TEXT DEFAULT 'activo',
			riesgo_nivel TEXT DEFAULT 'medio',
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_portal_contador_clientes_empresa ON empresa_portal_contador_clientes(empresa_id, estado_cliente, riesgo_nivel);`,
		`CREATE TABLE IF NOT EXISTS empresa_portal_contador_obligaciones (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			cliente_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			tipo TEXT DEFAULT 'iva',
			periodo TEXT,
			fecha_vencimiento TEXT,
			estado_obligacion TEXT DEFAULT 'pendiente',
			prioridad TEXT DEFAULT 'media',
			valor_estimado REAL DEFAULT 0,
			responsable TEXT,
			fecha_presentacion TEXT,
			soporte_url TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_portal_contador_obligaciones_empresa ON empresa_portal_contador_obligaciones(empresa_id, estado_obligacion, fecha_vencimiento);`,
		`CREATE TABLE IF NOT EXISTS empresa_portal_contador_solicitudes (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			cliente_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			titulo TEXT NOT NULL,
			categoria TEXT DEFAULT 'soportes',
			fecha_solicitud TEXT DEFAULT (CURRENT_DATE),
			fecha_limite TEXT,
			estado_solicitud TEXT DEFAULT 'abierta',
			prioridad TEXT DEFAULT 'media',
			responsable TEXT,
			respuesta TEXT,
			soporte_url TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_portal_contador_solicitudes_empresa ON empresa_portal_contador_solicitudes(empresa_id, estado_solicitud, fecha_limite);`,
		`CREATE TABLE IF NOT EXISTS empresa_portal_contador_comunicaciones (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			cliente_id INTEGER NOT NULL,
			canal TEXT DEFAULT 'interno',
			asunto TEXT,
			mensaje TEXT,
			fecha_mensaje TEXT DEFAULT (CURRENT_TIMESTAMP),
			leido_cliente INTEGER DEFAULT 0,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_portal_contador_comunicaciones_empresa ON empresa_portal_contador_comunicaciones(empresa_id, fecha_mensaje DESC);`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func BuildEmpresaPortalContadorDashboard(dbConn *sql.DB, empresaID int64) (EmpresaPortalContadorDashboard, error) {
	if err := EnsureEmpresaPortalContadorSchema(dbConn); err != nil {
		return EmpresaPortalContadorDashboard{}, err
	}
	clientes, err := listEmpresaPortalContadorClientes(dbConn, empresaID, "", 100)
	if err != nil {
		return EmpresaPortalContadorDashboard{}, err
	}
	obligaciones, _ := listEmpresaPortalContadorObligaciones(dbConn, empresaID, "", 200)
	solicitudes, _ := listEmpresaPortalContadorSolicitudes(dbConn, empresaID, "", 200)
	comunicaciones, _ := listEmpresaPortalContadorComunicaciones(dbConn, empresaID, 10)
	now := time.Now()
	dash := EmpresaPortalContadorDashboard{EmpresaID: empresaID, Clientes: clientes, UltimasComunicaciones: comunicaciones}
	for _, c := range clientes {
		if c.EstadoCliente == "activo" {
			dash.ClientesActivos++
		}
		if c.RiesgoNivel == "alta" || c.RiesgoNivel == "critica" {
			dash.ClientesRiesgoAlto++
		}
	}
	for _, o := range obligaciones {
		if o.EstadoObligacion == "pendiente" || o.EstadoObligacion == "en_revision" {
			dash.ObligacionesPendientes++
			if o.DiasParaVencer <= 7 {
				dash.ObligacionesVencen7Dias++
			}
			if len(dash.ObligacionesPrioritarias) < 12 && (o.DiasParaVencer <= 7 || o.Prioridad == "alta" || o.Prioridad == "critica") {
				dash.ObligacionesPrioritarias = append(dash.ObligacionesPrioritarias, o)
			}
		}
	}
	for _, s := range solicitudes {
		if s.EstadoSolicitud == "abierta" || s.EstadoSolicitud == "enviada" || s.EstadoSolicitud == "en_revision" {
			dash.SolicitudesAbiertas++
			if s.DiasParaVencer < 0 {
				dash.SolicitudesVencidas++
			}
			if len(dash.SolicitudesPrioritarias) < 12 && (s.DiasParaVencer <= 5 || s.Prioridad == "alta" || s.Prioridad == "critica") {
				dash.SolicitudesPrioritarias = append(dash.SolicitudesPrioritarias, s)
			}
		}
	}
	dash.ComunicacionesMes, _ = countPortalContador(dbConn, `SELECT COUNT(1) FROM empresa_portal_contador_comunicaciones WHERE empresa_id=? AND substr(COALESCE(fecha_mensaje,''),1,7)=?`, empresaID, now.Format("2006-01"))
	if dash.ObligacionesVencen7Dias > 0 {
		dash.Alertas = append(dash.Alertas, "Hay obligaciones tributarias o contables que vencen en los proximos 7 dias.")
	}
	if dash.SolicitudesVencidas > 0 {
		dash.Alertas = append(dash.Alertas, "Existen solicitudes de documentos vencidas; requiere seguimiento con el cliente.")
	}
	if dash.ClientesRiesgoAlto > 0 {
		dash.Alertas = append(dash.Alertas, "Hay clientes marcados con riesgo alto o critico.")
	}
	if len(clientes) == 0 {
		dash.Alertas = append(dash.Alertas, "No hay clientes contables registrados en el portal.")
	}
	return dash, nil
}

func UpsertEmpresaPortalContadorCliente(dbConn *sql.DB, row EmpresaPortalContadorCliente) (int64, error) {
	if err := EnsureEmpresaPortalContadorSchema(dbConn); err != nil {
		return 0, err
	}
	row.RazonSocial = strings.TrimSpace(row.RazonSocial)
	if row.EmpresaID <= 0 || row.RazonSocial == "" {
		return 0, errors.New("empresa_id y razon_social son obligatorios")
	}
	if row.Codigo == "" {
		row.Codigo = nextPortalContadorCode(dbConn, row.EmpresaID, "empresa_portal_contador_clientes", "PC-CLI")
	}
	row.Regimen = normalizePortalContadorRegimen(row.Regimen)
	row.Periodicidad = normalizePortalContadorPeriodicidad(row.Periodicidad)
	row.EstadoCliente = normalizePortalContadorEstadoCliente(row.EstadoCliente)
	row.RiesgoNivel = normalizePortalContadorPrioridad(row.RiesgoNivel)
	row.Estado = normalizePortalContadorEstado(row.Estado)
	if row.CierreMensualDia <= 0 || row.CierreMensualDia > 31 {
		row.CierreMensualDia = 10
	}
	if row.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_portal_contador_clientes SET cliente_empresa_id=?,razon_social=?,nit=?,regimen=?,responsable=?,contacto_nombre=?,contacto_email=?,contacto_telefono=?,periodicidad=?,fecha_inicio=?,cierre_mensual_dia=?,estado_cliente=?,riesgo_nivel=?,fecha_actualizacion=CURRENT_TIMESTAMP,usuario_creador=?,estado=?,observaciones=? WHERE id=? AND empresa_id=?`, row.ClienteEmpresaID, row.RazonSocial, row.NIT, row.Regimen, row.Responsable, row.ContactoNombre, row.ContactoEmail, row.ContactoTelefono, row.Periodicidad, row.FechaInicio, row.CierreMensualDia, row.EstadoCliente, row.RiesgoNivel, row.Usuario, row.Estado, row.Observaciones, row.ID, row.EmpresaID)
		return row.ID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_portal_contador_clientes (empresa_id,cliente_empresa_id,codigo,razon_social,nit,regimen,responsable,contacto_nombre,contacto_email,contacto_telefono,periodicidad,fecha_inicio,cierre_mensual_dia,estado_cliente,riesgo_nivel,usuario_creador,estado,observaciones) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, row.EmpresaID, row.ClienteEmpresaID, row.Codigo, row.RazonSocial, row.NIT, row.Regimen, row.Responsable, row.ContactoNombre, row.ContactoEmail, row.ContactoTelefono, row.Periodicidad, row.FechaInicio, row.CierreMensualDia, row.EstadoCliente, row.RiesgoNivel, row.Usuario, row.Estado, row.Observaciones)
}

func UpsertEmpresaPortalContadorObligacion(dbConn *sql.DB, row EmpresaPortalContadorObligacion) (int64, error) {
	if err := EnsureEmpresaPortalContadorSchema(dbConn); err != nil {
		return 0, err
	}
	if row.EmpresaID <= 0 || row.ClienteID <= 0 {
		return 0, errors.New("empresa_id y cliente_id son obligatorios")
	}
	if row.Codigo == "" {
		row.Codigo = nextPortalContadorCode(dbConn, row.EmpresaID, "empresa_portal_contador_obligaciones", "PC-OBL")
	}
	row.Tipo = normalizePortalContadorTipoObligacion(row.Tipo)
	row.EstadoObligacion = normalizePortalContadorEstadoObligacion(row.EstadoObligacion)
	row.Prioridad = normalizePortalContadorPrioridad(row.Prioridad)
	if row.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_portal_contador_obligaciones SET cliente_id=?,tipo=?,periodo=?,fecha_vencimiento=?,estado_obligacion=?,prioridad=?,valor_estimado=?,responsable=?,fecha_presentacion=?,soporte_url=?,fecha_actualizacion=CURRENT_TIMESTAMP,usuario_creador=?,observaciones=? WHERE id=? AND empresa_id=?`, row.ClienteID, row.Tipo, row.Periodo, row.FechaVencimiento, row.EstadoObligacion, row.Prioridad, row.ValorEstimado, row.Responsable, row.FechaPresentacion, row.SoporteURL, row.Usuario, row.Observaciones, row.ID, row.EmpresaID)
		return row.ID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_portal_contador_obligaciones (empresa_id,cliente_id,codigo,tipo,periodo,fecha_vencimiento,estado_obligacion,prioridad,valor_estimado,responsable,fecha_presentacion,soporte_url,usuario_creador,observaciones) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, row.EmpresaID, row.ClienteID, row.Codigo, row.Tipo, row.Periodo, row.FechaVencimiento, row.EstadoObligacion, row.Prioridad, row.ValorEstimado, row.Responsable, row.FechaPresentacion, row.SoporteURL, row.Usuario, row.Observaciones)
}

func UpsertEmpresaPortalContadorSolicitud(dbConn *sql.DB, row EmpresaPortalContadorSolicitud) (int64, error) {
	if err := EnsureEmpresaPortalContadorSchema(dbConn); err != nil {
		return 0, err
	}
	row.Titulo = strings.TrimSpace(row.Titulo)
	if row.EmpresaID <= 0 || row.ClienteID <= 0 || row.Titulo == "" {
		return 0, errors.New("empresa_id, cliente_id y titulo son obligatorios")
	}
	if row.Codigo == "" {
		row.Codigo = nextPortalContadorCode(dbConn, row.EmpresaID, "empresa_portal_contador_solicitudes", "PC-SOL")
	}
	if row.FechaSolicitud == "" {
		row.FechaSolicitud = time.Now().Format("2006-01-02")
	}
	row.Categoria = normalizePortalContadorCategoriaSolicitud(row.Categoria)
	row.EstadoSolicitud = normalizePortalContadorEstadoSolicitud(row.EstadoSolicitud)
	row.Prioridad = normalizePortalContadorPrioridad(row.Prioridad)
	if row.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_portal_contador_solicitudes SET cliente_id=?,titulo=?,categoria=?,fecha_solicitud=?,fecha_limite=?,estado_solicitud=?,prioridad=?,responsable=?,respuesta=?,soporte_url=?,fecha_actualizacion=CURRENT_TIMESTAMP,usuario_creador=?,observaciones=? WHERE id=? AND empresa_id=?`, row.ClienteID, row.Titulo, row.Categoria, row.FechaSolicitud, row.FechaLimite, row.EstadoSolicitud, row.Prioridad, row.Responsable, row.Respuesta, row.SoporteURL, row.Usuario, row.Observaciones, row.ID, row.EmpresaID)
		return row.ID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_portal_contador_solicitudes (empresa_id,cliente_id,codigo,titulo,categoria,fecha_solicitud,fecha_limite,estado_solicitud,prioridad,responsable,respuesta,soporte_url,usuario_creador,observaciones) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, row.EmpresaID, row.ClienteID, row.Codigo, row.Titulo, row.Categoria, row.FechaSolicitud, row.FechaLimite, row.EstadoSolicitud, row.Prioridad, row.Responsable, row.Respuesta, row.SoporteURL, row.Usuario, row.Observaciones)
}

func CreateEmpresaPortalContadorComunicacion(dbConn *sql.DB, row EmpresaPortalContadorComunicacion) (int64, error) {
	if err := EnsureEmpresaPortalContadorSchema(dbConn); err != nil {
		return 0, err
	}
	if row.EmpresaID <= 0 || row.ClienteID <= 0 || strings.TrimSpace(row.Mensaje) == "" {
		return 0, errors.New("empresa_id, cliente_id y mensaje son obligatorios")
	}
	row.Canal = normalizePortalContadorCanal(row.Canal)
	if row.FechaMensaje == "" {
		row.FechaMensaje = time.Now().Format("2006-01-02 15:04:05")
	}
	leido := 0
	if row.LeidoCliente {
		leido = 1
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_portal_contador_comunicaciones (empresa_id,cliente_id,canal,asunto,mensaje,fecha_mensaje,leido_cliente,usuario_creador,observaciones) VALUES (?,?,?,?,?,?,?,?,?)`, row.EmpresaID, row.ClienteID, row.Canal, row.Asunto, row.Mensaje, row.FechaMensaje, leido, row.Usuario, row.Observaciones)
}

func ListEmpresaPortalContadorClientes(dbConn *sql.DB, empresaID int64, q string, limit int) ([]EmpresaPortalContadorCliente, error) {
	if err := EnsureEmpresaPortalContadorSchema(dbConn); err != nil {
		return nil, err
	}
	return listEmpresaPortalContadorClientes(dbConn, empresaID, q, limit)
}

func listEmpresaPortalContadorClientes(dbConn *sql.DB, empresaID int64, q string, limit int) ([]EmpresaPortalContadorCliente, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	where := "empresa_id=? AND COALESCE(estado,'activo')='activo'"
	args := []interface{}{empresaID}
	if strings.TrimSpace(q) != "" {
		where += " AND (LOWER(COALESCE(razon_social,'')) LIKE ? OR LOWER(COALESCE(nit,'')) LIKE ? OR LOWER(COALESCE(responsable,'')) LIKE ?)"
		like := "%" + strings.ToLower(strings.TrimSpace(q)) + "%"
		args = append(args, like, like, like)
	}
	args = append(args, limit)
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,COALESCE(cliente_empresa_id,0),COALESCE(codigo,''),COALESCE(razon_social,''),COALESCE(nit,''),COALESCE(regimen,'responsable_iva'),COALESCE(responsable,''),COALESCE(contacto_nombre,''),COALESCE(contacto_email,''),COALESCE(contacto_telefono,''),COALESCE(periodicidad,'mensual'),COALESCE(fecha_inicio,''),COALESCE(cierre_mensual_dia,10),COALESCE(estado_cliente,'activo'),COALESCE(riesgo_nivel,'medio'),COALESCE(usuario_creador,''),COALESCE(estado,'activo'),COALESCE(observaciones,''),COALESCE(fecha_creacion,'') FROM empresa_portal_contador_clientes WHERE `+where+` ORDER BY razon_social LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaPortalContadorCliente{}
	for rows.Next() {
		var row EmpresaPortalContadorCliente
		if err := rows.Scan(&row.ID, &row.EmpresaID, &row.ClienteEmpresaID, &row.Codigo, &row.RazonSocial, &row.NIT, &row.Regimen, &row.Responsable, &row.ContactoNombre, &row.ContactoEmail, &row.ContactoTelefono, &row.Periodicidad, &row.FechaInicio, &row.CierreMensualDia, &row.EstadoCliente, &row.RiesgoNivel, &row.Usuario, &row.Estado, &row.Observaciones, &row.FechaCreacion); err != nil {
			return nil, err
		}
		row.ObligacionesAbiertas, _ = countPortalContador(dbConn, `SELECT COUNT(1) FROM empresa_portal_contador_obligaciones WHERE empresa_id=? AND cliente_id=? AND estado_obligacion IN ('pendiente','en_revision')`, empresaID, row.ID)
		row.SolicitudesAbiertas, _ = countPortalContador(dbConn, `SELECT COUNT(1) FROM empresa_portal_contador_solicitudes WHERE empresa_id=? AND cliente_id=? AND estado_solicitud IN ('abierta','enviada','en_revision')`, empresaID, row.ID)
		out = append(out, row)
	}
	return out, rows.Err()
}

func ListEmpresaPortalContadorObligaciones(dbConn *sql.DB, empresaID int64, estado string, limit int) ([]EmpresaPortalContadorObligacion, error) {
	if err := EnsureEmpresaPortalContadorSchema(dbConn); err != nil {
		return nil, err
	}
	return listEmpresaPortalContadorObligaciones(dbConn, empresaID, estado, limit)
}

func listEmpresaPortalContadorObligaciones(dbConn *sql.DB, empresaID int64, estado string, limit int) ([]EmpresaPortalContadorObligacion, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	where := "o.empresa_id=?"
	args := []interface{}{empresaID}
	if e := normalizePortalContadorEstadoObligacion(estado); e != "" && e != "todas" {
		where += " AND o.estado_obligacion=?"
		args = append(args, e)
	}
	args = append(args, limit)
	rows, err := ExecQueryCompat(dbConn, `SELECT o.id,o.empresa_id,o.cliente_id,COALESCE(o.codigo,''),COALESCE(o.tipo,'iva'),COALESCE(o.periodo,''),COALESCE(o.fecha_vencimiento,''),COALESCE(o.estado_obligacion,'pendiente'),COALESCE(o.prioridad,'media'),COALESCE(o.valor_estimado,0),COALESCE(o.responsable,''),COALESCE(o.fecha_presentacion,''),COALESCE(o.soporte_url,''),COALESCE(o.usuario_creador,''),COALESCE(o.observaciones,''),COALESCE(o.fecha_creacion,''),COALESCE(c.razon_social,'') FROM empresa_portal_contador_obligaciones o LEFT JOIN empresa_portal_contador_clientes c ON c.id=o.cliente_id AND c.empresa_id=o.empresa_id WHERE `+where+` ORDER BY o.fecha_vencimiento ASC, o.prioridad DESC LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaPortalContadorObligacion{}
	for rows.Next() {
		var row EmpresaPortalContadorObligacion
		if err := rows.Scan(&row.ID, &row.EmpresaID, &row.ClienteID, &row.Codigo, &row.Tipo, &row.Periodo, &row.FechaVencimiento, &row.EstadoObligacion, &row.Prioridad, &row.ValorEstimado, &row.Responsable, &row.FechaPresentacion, &row.SoporteURL, &row.Usuario, &row.Observaciones, &row.FechaCreacion, &row.ClienteNombre); err != nil {
			return nil, err
		}
		row.DiasParaVencer = portalContadorDiasParaVencer(row.FechaVencimiento, time.Now())
		out = append(out, row)
	}
	return out, rows.Err()
}

func ListEmpresaPortalContadorSolicitudes(dbConn *sql.DB, empresaID int64, estado string, limit int) ([]EmpresaPortalContadorSolicitud, error) {
	if err := EnsureEmpresaPortalContadorSchema(dbConn); err != nil {
		return nil, err
	}
	return listEmpresaPortalContadorSolicitudes(dbConn, empresaID, estado, limit)
}

func listEmpresaPortalContadorSolicitudes(dbConn *sql.DB, empresaID int64, estado string, limit int) ([]EmpresaPortalContadorSolicitud, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	where := "s.empresa_id=?"
	args := []interface{}{empresaID}
	if e := normalizePortalContadorEstadoSolicitud(estado); e != "" && e != "todas" {
		where += " AND s.estado_solicitud=?"
		args = append(args, e)
	}
	args = append(args, limit)
	rows, err := ExecQueryCompat(dbConn, `SELECT s.id,s.empresa_id,s.cliente_id,COALESCE(s.codigo,''),COALESCE(s.titulo,''),COALESCE(s.categoria,'soportes'),COALESCE(s.fecha_solicitud,''),COALESCE(s.fecha_limite,''),COALESCE(s.estado_solicitud,'abierta'),COALESCE(s.prioridad,'media'),COALESCE(s.responsable,''),COALESCE(s.respuesta,''),COALESCE(s.soporte_url,''),COALESCE(s.usuario_creador,''),COALESCE(s.observaciones,''),COALESCE(s.fecha_creacion,''),COALESCE(c.razon_social,'') FROM empresa_portal_contador_solicitudes s LEFT JOIN empresa_portal_contador_clientes c ON c.id=s.cliente_id AND c.empresa_id=s.empresa_id WHERE `+where+` ORDER BY s.fecha_limite ASC, s.prioridad DESC LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaPortalContadorSolicitud{}
	for rows.Next() {
		var row EmpresaPortalContadorSolicitud
		if err := rows.Scan(&row.ID, &row.EmpresaID, &row.ClienteID, &row.Codigo, &row.Titulo, &row.Categoria, &row.FechaSolicitud, &row.FechaLimite, &row.EstadoSolicitud, &row.Prioridad, &row.Responsable, &row.Respuesta, &row.SoporteURL, &row.Usuario, &row.Observaciones, &row.FechaCreacion, &row.ClienteNombre); err != nil {
			return nil, err
		}
		row.DiasParaVencer = portalContadorDiasParaVencer(row.FechaLimite, time.Now())
		out = append(out, row)
	}
	return out, rows.Err()
}

func ListEmpresaPortalContadorComunicaciones(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaPortalContadorComunicacion, error) {
	if err := EnsureEmpresaPortalContadorSchema(dbConn); err != nil {
		return nil, err
	}
	return listEmpresaPortalContadorComunicaciones(dbConn, empresaID, limit)
}

func listEmpresaPortalContadorComunicaciones(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaPortalContadorComunicacion, error) {
	if limit <= 0 || limit > 300 {
		limit = 100
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT m.id,m.empresa_id,m.cliente_id,COALESCE(m.canal,'interno'),COALESCE(m.asunto,''),COALESCE(m.mensaje,''),COALESCE(m.fecha_mensaje,''),COALESCE(m.leido_cliente,0),COALESCE(m.usuario_creador,''),COALESCE(m.observaciones,''),COALESCE(m.fecha_creacion,''),COALESCE(c.razon_social,'') FROM empresa_portal_contador_comunicaciones m LEFT JOIN empresa_portal_contador_clientes c ON c.id=m.cliente_id AND c.empresa_id=m.empresa_id WHERE m.empresa_id=? ORDER BY m.fecha_mensaje DESC, m.id DESC LIMIT ?`, empresaID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaPortalContadorComunicacion{}
	for rows.Next() {
		var row EmpresaPortalContadorComunicacion
		var leido int
		if err := rows.Scan(&row.ID, &row.EmpresaID, &row.ClienteID, &row.Canal, &row.Asunto, &row.Mensaje, &row.FechaMensaje, &leido, &row.Usuario, &row.Observaciones, &row.FechaCreacion, &row.ClienteNombre); err != nil {
			return nil, err
		}
		row.LeidoCliente = leido != 0
		out = append(out, row)
	}
	return out, rows.Err()
}

func SeedEmpresaPortalContadorDemo(dbConn *sql.DB, empresaID int64, usuario string) error {
	if err := EnsureEmpresaPortalContadorSchema(dbConn); err != nil {
		return err
	}
	if usuario == "" {
		usuario = "sistema"
	}
	count, _ := countPortalContador(dbConn, `SELECT COUNT(1) FROM empresa_portal_contador_clientes WHERE empresa_id=?`, empresaID)
	if count > 0 {
		return nil
	}
	clienteID, err := UpsertEmpresaPortalContadorCliente(dbConn, EmpresaPortalContadorCliente{
		EmpresaID: empresaID, RazonSocial: "Motel Calipso SAS", NIT: "900123456-7", Regimen: "responsable_iva", Responsable: "Contador principal", ContactoNombre: "Gerencia", ContactoEmail: "gerencia@calipso.local", Periodicidad: "mensual", FechaInicio: time.Now().AddDate(0, -6, 0).Format("2006-01-02"), CierreMensualDia: 10, EstadoCliente: "activo", RiesgoNivel: "medio", Usuario: usuario, Observaciones: "Cliente demo del portal contador",
	})
	if err != nil {
		return err
	}
	_, _ = UpsertEmpresaPortalContadorCliente(dbConn, EmpresaPortalContadorCliente{
		EmpresaID: empresaID, RazonSocial: "Restaurante La Avenida", NIT: "901777333-1", Regimen: "simple", Responsable: "Auxiliar contable", ContactoNombre: "Administrador", ContactoEmail: "admin@avenida.local", Periodicidad: "bimestral", FechaInicio: time.Now().AddDate(-1, 0, 0).Format("2006-01-02"), CierreMensualDia: 12, EstadoCliente: "activo", RiesgoNivel: "alto", Usuario: usuario,
	})
	_, _ = UpsertEmpresaPortalContadorObligacion(dbConn, EmpresaPortalContadorObligacion{EmpresaID: empresaID, ClienteID: clienteID, Tipo: "iva", Periodo: time.Now().Format("2006-01"), FechaVencimiento: time.Now().AddDate(0, 0, 6).Format("2006-01-02"), EstadoObligacion: "pendiente", Prioridad: "alta", ValorEstimado: 1250000, Responsable: "Contador principal", Usuario: usuario, Observaciones: "Preparar IVA del periodo"})
	_, _ = UpsertEmpresaPortalContadorObligacion(dbConn, EmpresaPortalContadorObligacion{EmpresaID: empresaID, ClienteID: clienteID, Tipo: "retencion_fuente", Periodo: time.Now().Format("2006-01"), FechaVencimiento: time.Now().AddDate(0, 0, 12).Format("2006-01-02"), EstadoObligacion: "en_revision", Prioridad: "media", ValorEstimado: 320000, Responsable: "Auxiliar contable", Usuario: usuario})
	_, _ = UpsertEmpresaPortalContadorSolicitud(dbConn, EmpresaPortalContadorSolicitud{EmpresaID: empresaID, ClienteID: clienteID, Titulo: "Enviar extractos bancarios del mes", Categoria: "extractos", FechaLimite: time.Now().AddDate(0, 0, 3).Format("2006-01-02"), EstadoSolicitud: "abierta", Prioridad: "alta", Responsable: "Cliente", Usuario: usuario})
	_, _ = CreateEmpresaPortalContadorComunicacion(dbConn, EmpresaPortalContadorComunicacion{EmpresaID: empresaID, ClienteID: clienteID, Canal: "interno", Asunto: "Cierre mensual", Mensaje: "Se solicita cargar soportes de bancos, compras y nomina para preparar el cierre mensual.", Usuario: usuario})
	return nil
}

func countPortalContador(dbConn *sql.DB, query string, args ...interface{}) (int, error) {
	var n int
	err := QueryRowCompat(dbConn, query, args...).Scan(&n)
	return n, err
}

func nextPortalContadorCode(dbConn *sql.DB, empresaID int64, table, prefix string) string {
	var count int
	_ = QueryRowCompat(dbConn, fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE empresa_id=?", table), empresaID).Scan(&count)
	return fmt.Sprintf("%s-%04d", prefix, count+1)
}

func portalContadorDiasParaVencer(fecha string, today time.Time) int {
	fecha = strings.TrimSpace(fecha)
	if fecha == "" {
		return 9999
	}
	for _, layout := range []string{"2006-01-02", "2006-01-02 15:04:05", time.RFC3339} {
		if parsed, err := time.ParseInLocation(layout, fecha, today.Location()); err == nil {
			return int(parsed.Truncate(24*time.Hour).Sub(today.Truncate(24*time.Hour)).Hours() / 24)
		}
	}
	return 9999
}

func normalizePortalContadorRegimen(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "responsable_iva", "no_responsable_iva", "simple", "gran_contribuyente", "persona_natural":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "responsable_iva"
	}
}

func normalizePortalContadorPeriodicidad(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "mensual", "bimestral", "trimestral", "semestral", "anual":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "mensual"
	}
}

func normalizePortalContadorEstadoCliente(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "activo", "pausado", "en_revision", "retirado":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "activo"
	}
}

func normalizePortalContadorEstadoObligacion(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "pendiente", "en_revision", "presentada", "pagada", "vencida", "cancelada", "todas":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "pendiente"
	}
}

func normalizePortalContadorEstadoSolicitud(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "abierta", "enviada", "en_revision", "resuelta", "vencida", "cancelada", "todas":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "abierta"
	}
}

func normalizePortalContadorPrioridad(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "baja", "media", "alta", "critica":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "media"
	}
}

func normalizePortalContadorTipoObligacion(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "iva", "retencion_fuente", "ica", "renta", "exogena", "nomina_electronica", "medios_magneticos", "cierre_contable", "otro":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "iva"
	}
}

func normalizePortalContadorCategoriaSolicitud(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "soportes", "extractos", "facturas", "nomina", "inventario", "impuestos", "contratos", "otro":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "soportes"
	}
}

func normalizePortalContadorCanal(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "interno", "email", "whatsapp", "llamada", "reunion":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "interno"
	}
}

func normalizePortalContadorEstado(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "inactivo" || v == "archivado" {
		return v
	}
	return "activo"
}
