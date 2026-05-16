package db

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

type EmpresaCompraRequisicion struct {
	ID             int64                          `json:"id"`
	EmpresaID      int64                          `json:"empresa_id"`
	Codigo         string                         `json:"codigo"`
	Solicitante    string                         `json:"solicitante"`
	Area           string                         `json:"area"`
	CentroCosto    string                         `json:"centro_costo"`
	Prioridad      string                         `json:"prioridad"`
	FechaSolicitud string                         `json:"fecha_solicitud"`
	FechaNecesidad string                         `json:"fecha_necesidad,omitempty"`
	EstadoFlujo    string                         `json:"estado_flujo"`
	TotalEstimado  float64                        `json:"total_estimado"`
	Justificacion  string                         `json:"justificacion,omitempty"`
	UsuarioCreador string                         `json:"usuario_creador,omitempty"`
	FechaCreacion  string                         `json:"fecha_creacion,omitempty"`
	Items          []EmpresaCompraRequisicionItem `json:"items,omitempty"`
	Cotizaciones   []EmpresaCompraCotizacion      `json:"cotizaciones,omitempty"`
	Aprobaciones   []EmpresaCompraAprobacion      `json:"aprobaciones,omitempty"`
	Recepciones    []EmpresaCompraRecepcion       `json:"recepciones,omitempty"`
}

type EmpresaCompraRequisicionItem struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	RequisicionID      int64   `json:"requisicion_id"`
	ProductoID         int64   `json:"producto_id,omitempty"`
	ProductoNombre     string  `json:"producto_nombre"`
	CantidadSolicitada float64 `json:"cantidad_solicitada"`
	CantidadRecibida   float64 `json:"cantidad_recibida"`
	Unidad             string  `json:"unidad"`
	CostoEstimado      float64 `json:"costo_estimado"`
	ProveedorSugerido  string  `json:"proveedor_sugerido,omitempty"`
	Especificacion     string  `json:"especificacion,omitempty"`
	Estado             string  `json:"estado"`
}

type EmpresaCompraCotizacion struct {
	ID                int64   `json:"id"`
	EmpresaID         int64   `json:"empresa_id"`
	RequisicionID     int64   `json:"requisicion_id"`
	ProveedorID       int64   `json:"proveedor_id,omitempty"`
	ProveedorNombre   string  `json:"proveedor_nombre"`
	Numero            string  `json:"numero"`
	FechaCotizacion   string  `json:"fecha_cotizacion"`
	ValidezHasta      string  `json:"validez_hasta,omitempty"`
	TiempoEntregaDias int     `json:"tiempo_entrega_dias"`
	Subtotal          float64 `json:"subtotal"`
	Impuestos         float64 `json:"impuestos"`
	Total             float64 `json:"total"`
	CondicionesPago   string  `json:"condiciones_pago,omitempty"`
	Observaciones     string  `json:"observaciones,omitempty"`
	Estado            string  `json:"estado"`
	UsuarioCreador    string  `json:"usuario_creador,omitempty"`
	FechaCreacion     string  `json:"fecha_creacion,omitempty"`
}

type EmpresaCompraAprobacion struct {
	ID              int64   `json:"id"`
	EmpresaID       int64   `json:"empresa_id"`
	RequisicionID   int64   `json:"requisicion_id"`
	CotizacionID    int64   `json:"cotizacion_id,omitempty"`
	Nivel           int     `json:"nivel"`
	Aprobador       string  `json:"aprobador"`
	Decision        string  `json:"decision"`
	Comentario      string  `json:"comentario,omitempty"`
	MontoAutorizado float64 `json:"monto_autorizado"`
	FechaDecision   string  `json:"fecha_decision,omitempty"`
}

type EmpresaCompraRecepcion struct {
	ID              int64                        `json:"id"`
	EmpresaID       int64                        `json:"empresa_id"`
	RequisicionID   int64                        `json:"requisicion_id"`
	CotizacionID    int64                        `json:"cotizacion_id,omitempty"`
	ProveedorID     int64                        `json:"proveedor_id,omitempty"`
	ProveedorNombre string                       `json:"proveedor_nombre"`
	Documento       string                       `json:"documento"`
	FechaRecepcion  string                       `json:"fecha_recepcion"`
	EstadoRecepcion string                       `json:"estado_recepcion"`
	Responsable     string                       `json:"responsable,omitempty"`
	Observaciones   string                       `json:"observaciones,omitempty"`
	UsuarioCreador  string                       `json:"usuario_creador,omitempty"`
	FechaCreacion   string                       `json:"fecha_creacion,omitempty"`
	Items           []EmpresaCompraRecepcionItem `json:"items,omitempty"`
}

type EmpresaCompraRecepcionItem struct {
	ID                int64   `json:"id"`
	EmpresaID         int64   `json:"empresa_id"`
	RecepcionID       int64   `json:"recepcion_id"`
	RequisicionItemID int64   `json:"requisicion_item_id,omitempty"`
	ProductoNombre    string  `json:"producto_nombre"`
	CantidadOrdenada  float64 `json:"cantidad_ordenada"`
	CantidadRecibida  float64 `json:"cantidad_recibida"`
	CantidadPendiente float64 `json:"cantidad_pendiente"`
	CostoUnitario     float64 `json:"costo_unitario"`
	Lote              string  `json:"lote,omitempty"`
	EstadoCalidad     string  `json:"estado_calidad"`
}

type EmpresaComprasAvanzadasDashboard struct {
	EmpresaID                   int64                      `json:"empresa_id"`
	RequisicionesAbiertas       int                        `json:"requisiciones_abiertas"`
	RequisicionesPendAprobacion int                        `json:"requisiciones_pendientes_aprobacion"`
	CotizacionesEnEvaluacion    int                        `json:"cotizaciones_en_evaluacion"`
	RecepcionesPendientes       int                        `json:"recepciones_pendientes"`
	ValorPendienteAprobacion    float64                    `json:"valor_pendiente_aprobacion"`
	UltimasRequisiciones        []EmpresaCompraRequisicion `json:"ultimas_requisiciones"`
}

func EnsureEmpresaComprasAvanzadasSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_compras_requisiciones (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			solicitante TEXT DEFAULT '',
			area TEXT DEFAULT '',
			centro_costo TEXT DEFAULT '',
			prioridad TEXT DEFAULT 'media',
			fecha_solicitud TEXT NOT NULL,
			fecha_necesidad TEXT,
			estado_flujo TEXT DEFAULT 'borrador',
			total_estimado REAL DEFAULT 0,
			justificacion TEXT,
			usuario_creador TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(empresa_id, codigo)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_compras_req_empresa_estado ON empresa_compras_requisiciones(empresa_id, estado_flujo)`,
		`CREATE TABLE IF NOT EXISTS empresa_compras_requisicion_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			requisicion_id INTEGER NOT NULL,
			producto_id INTEGER DEFAULT 0,
			producto_nombre TEXT NOT NULL,
			cantidad_solicitada REAL DEFAULT 0,
			cantidad_recibida REAL DEFAULT 0,
			unidad TEXT DEFAULT 'und',
			costo_estimado REAL DEFAULT 0,
			proveedor_sugerido TEXT,
			especificacion TEXT,
			estado TEXT DEFAULT 'pendiente'
		)`,
		`CREATE INDEX IF NOT EXISTS ix_compras_req_items_req ON empresa_compras_requisicion_items(empresa_id, requisicion_id)`,
		`CREATE TABLE IF NOT EXISTS empresa_compras_cotizaciones (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			requisicion_id INTEGER NOT NULL,
			proveedor_id INTEGER DEFAULT 0,
			proveedor_nombre TEXT NOT NULL,
			numero TEXT NOT NULL,
			fecha_cotizacion TEXT NOT NULL,
			validez_hasta TEXT,
			tiempo_entrega_dias INTEGER DEFAULT 0,
			subtotal REAL DEFAULT 0,
			impuestos REAL DEFAULT 0,
			total REAL DEFAULT 0,
			condiciones_pago TEXT,
			observaciones TEXT,
			estado TEXT DEFAULT 'recibida',
			usuario_creador TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(empresa_id, requisicion_id, numero)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_compras_cot_req ON empresa_compras_cotizaciones(empresa_id, requisicion_id)`,
		`CREATE TABLE IF NOT EXISTS empresa_compras_aprobaciones (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			requisicion_id INTEGER NOT NULL,
			cotizacion_id INTEGER DEFAULT 0,
			nivel INTEGER DEFAULT 1,
			aprobador TEXT DEFAULT '',
			decision TEXT DEFAULT 'pendiente',
			comentario TEXT,
			monto_autorizado REAL DEFAULT 0,
			fecha_decision TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS ix_compras_aprob_req ON empresa_compras_aprobaciones(empresa_id, requisicion_id)`,
		`CREATE TABLE IF NOT EXISTS empresa_compras_recepciones_avanzadas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			requisicion_id INTEGER NOT NULL,
			cotizacion_id INTEGER DEFAULT 0,
			proveedor_id INTEGER DEFAULT 0,
			proveedor_nombre TEXT DEFAULT '',
			documento TEXT NOT NULL,
			fecha_recepcion TEXT NOT NULL,
			estado_recepcion TEXT DEFAULT 'parcial',
			responsable TEXT,
			observaciones TEXT,
			usuario_creador TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS ix_compras_recep_req ON empresa_compras_recepciones_avanzadas(empresa_id, requisicion_id)`,
		`CREATE TABLE IF NOT EXISTS empresa_compras_recepcion_items_avanzadas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			recepcion_id INTEGER NOT NULL,
			requisicion_item_id INTEGER DEFAULT 0,
			producto_nombre TEXT NOT NULL,
			cantidad_ordenada REAL DEFAULT 0,
			cantidad_recibida REAL DEFAULT 0,
			cantidad_pendiente REAL DEFAULT 0,
			costo_unitario REAL DEFAULT 0,
			lote TEXT,
			estado_calidad TEXT DEFAULT 'aprobado'
		)`,
		`CREATE INDEX IF NOT EXISTS ix_compras_recep_items ON empresa_compras_recepcion_items_avanzadas(empresa_id, recepcion_id)`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_compras_recepciones_avanzadas", "proveedor_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	return nil
}

func CreateEmpresaCompraRequisicion(dbConn *sql.DB, req EmpresaCompraRequisicion) (int64, error) {
	req = normalizeCompraRequisicion(req)
	if req.EmpresaID <= 0 || req.Codigo == "" {
		return 0, errors.New("empresa_id y codigo son requeridos")
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	id, err := insertTxSQLCompat(tx, `INSERT INTO empresa_compras_requisiciones
		(empresa_id,codigo,solicitante,area,centro_costo,prioridad,fecha_solicitud,fecha_necesidad,estado_flujo,total_estimado,justificacion,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (empresa_id,codigo) DO UPDATE SET
			solicitante=EXCLUDED.solicitante,
			area=EXCLUDED.area,
			centro_costo=EXCLUDED.centro_costo,
			prioridad=EXCLUDED.prioridad,
			fecha_solicitud=EXCLUDED.fecha_solicitud,
			fecha_necesidad=EXCLUDED.fecha_necesidad,
			estado_flujo=EXCLUDED.estado_flujo,
			total_estimado=EXCLUDED.total_estimado,
			justificacion=EXCLUDED.justificacion,
			usuario_creador=EXCLUDED.usuario_creador,
			fecha_actualizacion=CURRENT_TIMESTAMP`,
		req.EmpresaID, req.Codigo, req.Solicitante, req.Area, req.CentroCosto, req.Prioridad, req.FechaSolicitud, req.FechaNecesidad, req.EstadoFlujo, req.TotalEstimado, req.Justificacion, req.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	if len(req.Items) > 0 {
		if _, err := execTxSQLCompat(tx, `DELETE FROM empresa_compras_requisicion_items WHERE empresa_id=? AND requisicion_id=?`, req.EmpresaID, id); err != nil {
			return 0, err
		}
		total := 0.0
		for _, item := range req.Items {
			item.EmpresaID = req.EmpresaID
			item.RequisicionID = id
			item = normalizeCompraRequisicionItem(item)
			total += item.CantidadSolicitada * item.CostoEstimado
			if _, err := insertTxSQLCompat(tx, `INSERT INTO empresa_compras_requisicion_items
				(empresa_id,requisicion_id,producto_id,producto_nombre,cantidad_solicitada,cantidad_recibida,unidad,costo_estimado,proveedor_sugerido,especificacion,estado)
				VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
				item.EmpresaID, item.RequisicionID, item.ProductoID, item.ProductoNombre, item.CantidadSolicitada, item.CantidadRecibida, item.Unidad, item.CostoEstimado, item.ProveedorSugerido, item.Especificacion, item.Estado); err != nil {
				return 0, err
			}
		}
		if total > 0 {
			if _, err := execTxSQLCompat(tx, `UPDATE empresa_compras_requisiciones SET total_estimado=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, roundMoney(total), req.EmpresaID, id); err != nil {
				return 0, err
			}
		}
	}
	return id, tx.Commit()
}

func CreateEmpresaCompraCotizacion(dbConn *sql.DB, cot EmpresaCompraCotizacion) (int64, error) {
	cot = normalizeCompraCotizacion(cot)
	if cot.EmpresaID <= 0 || cot.RequisicionID <= 0 || cot.ProveedorNombre == "" || cot.Numero == "" {
		return 0, errors.New("requisicion, proveedor y numero son requeridos")
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_compras_cotizaciones
		(empresa_id,requisicion_id,proveedor_id,proveedor_nombre,numero,fecha_cotizacion,validez_hasta,tiempo_entrega_dias,subtotal,impuestos,total,condiciones_pago,observaciones,estado,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (empresa_id,requisicion_id,numero) DO UPDATE SET
			proveedor_id=EXCLUDED.proveedor_id,
			proveedor_nombre=EXCLUDED.proveedor_nombre,
			fecha_cotizacion=EXCLUDED.fecha_cotizacion,
			validez_hasta=EXCLUDED.validez_hasta,
			tiempo_entrega_dias=EXCLUDED.tiempo_entrega_dias,
			subtotal=EXCLUDED.subtotal,
			impuestos=EXCLUDED.impuestos,
			total=EXCLUDED.total,
			condiciones_pago=EXCLUDED.condiciones_pago,
			observaciones=EXCLUDED.observaciones,
			estado=EXCLUDED.estado,
			usuario_creador=EXCLUDED.usuario_creador`,
		cot.EmpresaID, cot.RequisicionID, cot.ProveedorID, cot.ProveedorNombre, cot.Numero, cot.FechaCotizacion, cot.ValidezHasta, cot.TiempoEntregaDias, cot.Subtotal, cot.Impuestos, cot.Total, cot.CondicionesPago, cot.Observaciones, cot.Estado, cot.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	_, _ = ExecCompat(dbConn, `UPDATE empresa_compras_requisiciones SET estado_flujo='cotizando', fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=? AND estado_flujo IN ('borrador','solicitada','cotizando')`, cot.EmpresaID, cot.RequisicionID)
	return id, nil
}

func ResolverEmpresaCompraAprobacion(dbConn *sql.DB, apr EmpresaCompraAprobacion) (int64, error) {
	apr = normalizeCompraAprobacion(apr)
	if apr.EmpresaID <= 0 || apr.RequisicionID <= 0 || apr.Aprobador == "" {
		return 0, errors.New("requisicion y aprobador son requeridos")
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_compras_aprobaciones
		(empresa_id,requisicion_id,cotizacion_id,nivel,aprobador,decision,comentario,monto_autorizado)
		VALUES (?,?,?,?,?,?,?,?)`,
		apr.EmpresaID, apr.RequisicionID, apr.CotizacionID, apr.Nivel, apr.Aprobador, apr.Decision, apr.Comentario, apr.MontoAutorizado)
	if err != nil {
		return 0, err
	}
	if apr.CotizacionID > 0 && apr.Decision == "aprobada" {
		_, _ = ExecCompat(dbConn, `UPDATE empresa_compras_cotizaciones SET estado='seleccionada' WHERE empresa_id=? AND id=?`, apr.EmpresaID, apr.CotizacionID)
		_, _ = ExecCompat(dbConn, `UPDATE empresa_compras_cotizaciones SET estado='no_seleccionada' WHERE empresa_id=? AND requisicion_id=? AND id<>? AND estado IN ('recibida','evaluacion')`, apr.EmpresaID, apr.RequisicionID, apr.CotizacionID)
	}
	nextState := "pendiente_aprobacion"
	if apr.Decision == "aprobada" {
		nextState = "aprobada"
	} else if apr.Decision == "rechazada" {
		nextState = "rechazada"
	}
	_, _ = ExecCompat(dbConn, `UPDATE empresa_compras_requisiciones SET estado_flujo=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, nextState, apr.EmpresaID, apr.RequisicionID)
	return id, nil
}

func CreateEmpresaCompraRecepcion(dbConn *sql.DB, rec EmpresaCompraRecepcion) (int64, error) {
	rec = normalizeCompraRecepcion(rec)
	if rec.EmpresaID <= 0 || rec.RequisicionID <= 0 || rec.Documento == "" {
		return 0, errors.New("requisicion y documento son requeridos")
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	id, err := insertTxSQLCompat(tx, `INSERT INTO empresa_compras_recepciones_avanzadas
		(empresa_id,requisicion_id,cotizacion_id,proveedor_id,proveedor_nombre,documento,fecha_recepcion,estado_recepcion,responsable,observaciones,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		rec.EmpresaID, rec.RequisicionID, rec.CotizacionID, rec.ProveedorID, rec.ProveedorNombre, rec.Documento, rec.FechaRecepcion, rec.EstadoRecepcion, rec.Responsable, rec.Observaciones, rec.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	for _, item := range rec.Items {
		item.EmpresaID = rec.EmpresaID
		item.RecepcionID = id
		item = normalizeCompraRecepcionItem(item)
		if _, err := insertTxSQLCompat(tx, `INSERT INTO empresa_compras_recepcion_items_avanzadas
			(empresa_id,recepcion_id,requisicion_item_id,producto_nombre,cantidad_ordenada,cantidad_recibida,cantidad_pendiente,costo_unitario,lote,estado_calidad)
			VALUES (?,?,?,?,?,?,?,?,?,?)`,
			item.EmpresaID, item.RecepcionID, item.RequisicionItemID, item.ProductoNombre, item.CantidadOrdenada, item.CantidadRecibida, item.CantidadPendiente, item.CostoUnitario, item.Lote, item.EstadoCalidad); err != nil {
			return 0, err
		}
		if item.RequisicionItemID > 0 {
			if _, err := execTxSQLCompat(tx, `UPDATE empresa_compras_requisicion_items SET cantidad_recibida=COALESCE(cantidad_recibida,0)+?, estado=CASE WHEN COALESCE(cantidad_recibida,0)+? >= cantidad_solicitada THEN 'recibido' ELSE 'parcial' END WHERE empresa_id=? AND id=?`, item.CantidadRecibida, item.CantidadRecibida, item.EmpresaID, item.RequisicionItemID); err != nil {
				return 0, err
			}
		}
	}
	state := "recibida_parcial"
	if rec.EstadoRecepcion == "total" || compraRecepcionCompletaTx(tx, rec.EmpresaID, rec.RequisicionID) {
		state = "recibida_total"
	}
	if _, err := execTxSQLCompat(tx, `UPDATE empresa_compras_requisiciones SET estado_flujo=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, state, rec.EmpresaID, rec.RequisicionID); err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

func ListEmpresaCompraRequisiciones(dbConn *sql.DB, empresaID int64, estado string, limit int) ([]EmpresaCompraRequisicion, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if strings.TrimSpace(estado) != "" {
		where += " AND estado_flujo=?"
		args = append(args, normalizeCompraEstadoFlujo(estado))
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,codigo,COALESCE(solicitante,''),COALESCE(area,''),COALESCE(centro_costo,''),COALESCE(prioridad,'media'),COALESCE(fecha_solicitud,''),COALESCE(fecha_necesidad,''),COALESCE(estado_flujo,'borrador'),COALESCE(total_estimado,0),COALESCE(justificacion,''),COALESCE(usuario_creador,''),COALESCE(fecha_creacion,'') FROM empresa_compras_requisiciones WHERE %s ORDER BY id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCompraRequisicion{}
	for rows.Next() {
		var r EmpresaCompraRequisicion
		if err := rows.Scan(&r.ID, &r.EmpresaID, &r.Codigo, &r.Solicitante, &r.Area, &r.CentroCosto, &r.Prioridad, &r.FechaSolicitud, &r.FechaNecesidad, &r.EstadoFlujo, &r.TotalEstimado, &r.Justificacion, &r.UsuarioCreador, &r.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func GetEmpresaCompraRequisicion(dbConn *sql.DB, empresaID, id int64) (EmpresaCompraRequisicion, error) {
	rows, err := ListEmpresaCompraRequisiciones(dbConn, empresaID, "", 500)
	if err != nil {
		return EmpresaCompraRequisicion{}, err
	}
	for _, row := range rows {
		if row.ID == id {
			row.Items, _ = ListEmpresaCompraRequisicionItems(dbConn, empresaID, id)
			row.Cotizaciones, _ = ListEmpresaCompraCotizaciones(dbConn, empresaID, id)
			row.Aprobaciones, _ = ListEmpresaCompraAprobaciones(dbConn, empresaID, id)
			row.Recepciones, _ = ListEmpresaCompraRecepciones(dbConn, empresaID, id)
			return row, nil
		}
	}
	return EmpresaCompraRequisicion{}, sql.ErrNoRows
}

func ListEmpresaCompraRequisicionItems(dbConn *sql.DB, empresaID, requisicionID int64) ([]EmpresaCompraRequisicionItem, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,requisicion_id,COALESCE(producto_id,0),producto_nombre,COALESCE(cantidad_solicitada,0),COALESCE(cantidad_recibida,0),COALESCE(unidad,'und'),COALESCE(costo_estimado,0),COALESCE(proveedor_sugerido,''),COALESCE(especificacion,''),COALESCE(estado,'pendiente') FROM empresa_compras_requisicion_items WHERE empresa_id=? AND requisicion_id=? ORDER BY id ASC`, empresaID, requisicionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCompraRequisicionItem{}
	for rows.Next() {
		var x EmpresaCompraRequisicionItem
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.RequisicionID, &x.ProductoID, &x.ProductoNombre, &x.CantidadSolicitada, &x.CantidadRecibida, &x.Unidad, &x.CostoEstimado, &x.ProveedorSugerido, &x.Especificacion, &x.Estado); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func ListEmpresaCompraCotizaciones(dbConn *sql.DB, empresaID, requisicionID int64) ([]EmpresaCompraCotizacion, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,requisicion_id,COALESCE(proveedor_id,0),proveedor_nombre,numero,COALESCE(fecha_cotizacion,''),COALESCE(validez_hasta,''),COALESCE(tiempo_entrega_dias,0),COALESCE(subtotal,0),COALESCE(impuestos,0),COALESCE(total,0),COALESCE(condiciones_pago,''),COALESCE(observaciones,''),COALESCE(estado,'recibida'),COALESCE(usuario_creador,''),COALESCE(fecha_creacion,'') FROM empresa_compras_cotizaciones WHERE empresa_id=? AND requisicion_id=? ORDER BY total ASC, id DESC`, empresaID, requisicionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCompraCotizacion{}
	for rows.Next() {
		var x EmpresaCompraCotizacion
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.RequisicionID, &x.ProveedorID, &x.ProveedorNombre, &x.Numero, &x.FechaCotizacion, &x.ValidezHasta, &x.TiempoEntregaDias, &x.Subtotal, &x.Impuestos, &x.Total, &x.CondicionesPago, &x.Observaciones, &x.Estado, &x.UsuarioCreador, &x.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func ListEmpresaCompraAprobaciones(dbConn *sql.DB, empresaID, requisicionID int64) ([]EmpresaCompraAprobacion, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,requisicion_id,COALESCE(cotizacion_id,0),COALESCE(nivel,1),COALESCE(aprobador,''),COALESCE(decision,'pendiente'),COALESCE(comentario,''),COALESCE(monto_autorizado,0),COALESCE(fecha_decision,'') FROM empresa_compras_aprobaciones WHERE empresa_id=? AND requisicion_id=? ORDER BY id DESC`, empresaID, requisicionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCompraAprobacion{}
	for rows.Next() {
		var x EmpresaCompraAprobacion
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.RequisicionID, &x.CotizacionID, &x.Nivel, &x.Aprobador, &x.Decision, &x.Comentario, &x.MontoAutorizado, &x.FechaDecision); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func ListEmpresaCompraRecepciones(dbConn *sql.DB, empresaID, requisicionID int64) ([]EmpresaCompraRecepcion, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,requisicion_id,COALESCE(cotizacion_id,0),COALESCE(proveedor_id,0),COALESCE(proveedor_nombre,''),documento,COALESCE(fecha_recepcion,''),COALESCE(estado_recepcion,'parcial'),COALESCE(responsable,''),COALESCE(observaciones,''),COALESCE(usuario_creador,''),COALESCE(fecha_creacion,'') FROM empresa_compras_recepciones_avanzadas WHERE empresa_id=? AND requisicion_id=? ORDER BY id DESC`, empresaID, requisicionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCompraRecepcion{}
	for rows.Next() {
		var x EmpresaCompraRecepcion
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.RequisicionID, &x.CotizacionID, &x.ProveedorID, &x.ProveedorNombre, &x.Documento, &x.FechaRecepcion, &x.EstadoRecepcion, &x.Responsable, &x.Observaciones, &x.UsuarioCreador, &x.FechaCreacion); err != nil {
			return nil, err
		}
		x.Items, _ = ListEmpresaCompraRecepcionItems(dbConn, empresaID, x.ID)
		out = append(out, x)
	}
	return out, rows.Err()
}

func ListEmpresaCompraRecepcionItems(dbConn *sql.DB, empresaID, recepcionID int64) ([]EmpresaCompraRecepcionItem, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,recepcion_id,COALESCE(requisicion_item_id,0),producto_nombre,COALESCE(cantidad_ordenada,0),COALESCE(cantidad_recibida,0),COALESCE(cantidad_pendiente,0),COALESCE(costo_unitario,0),COALESCE(lote,''),COALESCE(estado_calidad,'aprobado') FROM empresa_compras_recepcion_items_avanzadas WHERE empresa_id=? AND recepcion_id=? ORDER BY id ASC`, empresaID, recepcionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCompraRecepcionItem{}
	for rows.Next() {
		var x EmpresaCompraRecepcionItem
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.RecepcionID, &x.RequisicionItemID, &x.ProductoNombre, &x.CantidadOrdenada, &x.CantidadRecibida, &x.CantidadPendiente, &x.CostoUnitario, &x.Lote, &x.EstadoCalidad); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func BuildEmpresaComprasAvanzadasDashboard(dbConn *sql.DB, empresaID int64) (EmpresaComprasAvanzadasDashboard, error) {
	var d EmpresaComprasAvanzadasDashboard
	d.EmpresaID = empresaID
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_compras_requisiciones WHERE empresa_id=? AND estado_flujo NOT IN ('rechazada','recibida_total','cancelada')`, empresaID).Scan(&d.RequisicionesAbiertas)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1), COALESCE(SUM(total_estimado),0) FROM empresa_compras_requisiciones WHERE empresa_id=? AND estado_flujo IN ('cotizando','pendiente_aprobacion')`, empresaID).Scan(&d.RequisicionesPendAprobacion, &d.ValorPendienteAprobacion)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_compras_cotizaciones WHERE empresa_id=? AND estado IN ('recibida','evaluacion')`, empresaID).Scan(&d.CotizacionesEnEvaluacion)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_compras_requisiciones WHERE empresa_id=? AND estado_flujo IN ('aprobada','recibida_parcial')`, empresaID).Scan(&d.RecepcionesPendientes)
	items, err := ListEmpresaCompraRequisiciones(dbConn, empresaID, "", 8)
	if err != nil {
		return d, err
	}
	d.UltimasRequisiciones = items
	return d, nil
}

func SeedEmpresaComprasAvanzadasDemo(dbConn *sql.DB, empresaID int64, usuario string) (int64, error) {
	code := "REQ-DEMO-" + time.Now().Format("20060102150405")
	reqID, err := CreateEmpresaCompraRequisicion(dbConn, EmpresaCompraRequisicion{
		EmpresaID: empresaID, Codigo: code, Solicitante: "Coordinacion de compras", Area: "Operaciones", CentroCosto: "Motel Calipso",
		Prioridad: "alta", FechaSolicitud: time.Now().Format("2006-01-02"), FechaNecesidad: time.Now().AddDate(0, 0, 5).Format("2006-01-02"),
		EstadoFlujo: "solicitada", Justificacion: "Reposicion operativa para habitaciones y mantenimiento", UsuarioCreador: usuario,
		Items: []EmpresaCompraRequisicionItem{
			{ProductoNombre: "Kit amenidades habitacion", CantidadSolicitada: 80, Unidad: "und", CostoEstimado: 3200, ProveedorSugerido: "Proveedor hotelero"},
			{ProductoNombre: "Control remoto universal", CantidadSolicitada: 12, Unidad: "und", CostoEstimado: 42000, ProveedorSugerido: "Proveedor electronico"},
		},
	})
	if err != nil {
		return 0, err
	}
	cotID, err := CreateEmpresaCompraCotizacion(dbConn, EmpresaCompraCotizacion{
		EmpresaID: empresaID, RequisicionID: reqID, ProveedorNombre: "Proveedor Integral Calipso", Numero: "COT-" + code,
		FechaCotizacion: time.Now().Format("2006-01-02"), ValidezHasta: time.Now().AddDate(0, 0, 15).Format("2006-01-02"),
		TiempoEntregaDias: 3, Subtotal: 760000, Impuestos: 144400, CondicionesPago: "Credito 15 dias", Estado: "evaluacion", UsuarioCreador: usuario,
	})
	if err != nil {
		return 0, err
	}
	if _, err := ResolverEmpresaCompraAprobacion(dbConn, EmpresaCompraAprobacion{EmpresaID: empresaID, RequisicionID: reqID, CotizacionID: cotID, Nivel: 1, Aprobador: usuario, Decision: "aprobada", Comentario: "Aprobado por QA compras avanzadas", MontoAutorizado: 904400}); err != nil {
		return 0, err
	}
	items, err := ListEmpresaCompraRequisicionItems(dbConn, empresaID, reqID)
	if err != nil {
		return 0, err
	}
	recItems := make([]EmpresaCompraRecepcionItem, 0, len(items))
	for _, item := range items {
		recItems = append(recItems, EmpresaCompraRecepcionItem{RequisicionItemID: item.ID, ProductoNombre: item.ProductoNombre, CantidadOrdenada: item.CantidadSolicitada, CantidadRecibida: item.CantidadSolicitada, CostoUnitario: item.CostoEstimado, EstadoCalidad: "aprobado"})
	}
	_, err = CreateEmpresaCompraRecepcion(dbConn, EmpresaCompraRecepcion{EmpresaID: empresaID, RequisicionID: reqID, CotizacionID: cotID, ProveedorNombre: "Proveedor Integral Calipso", Documento: "REM-" + code, FechaRecepcion: time.Now().Format("2006-01-02"), EstadoRecepcion: "total", Responsable: usuario, UsuarioCreador: usuario, Items: recItems})
	return reqID, err
}

func compraRecepcionCompletaTx(tx *sql.Tx, empresaID, requisicionID int64) bool {
	var pendientes int
	_ = queryRowTxSQLCompat(tx, `SELECT COUNT(1) FROM empresa_compras_requisicion_items WHERE empresa_id=? AND requisicion_id=? AND COALESCE(cantidad_recibida,0) < COALESCE(cantidad_solicitada,0)`, empresaID, requisicionID).Scan(&pendientes)
	return pendientes == 0
}

func normalizeCompraRequisicion(req EmpresaCompraRequisicion) EmpresaCompraRequisicion {
	req.Codigo = strings.ToUpper(strings.TrimSpace(req.Codigo))
	if req.Codigo == "" {
		req.Codigo = "REQ-" + time.Now().Format("20060102150405")
	}
	req.Solicitante = strings.TrimSpace(req.Solicitante)
	req.Area = strings.TrimSpace(req.Area)
	req.CentroCosto = strings.TrimSpace(req.CentroCosto)
	req.Prioridad = normalizeCompraPrioridad(req.Prioridad)
	req.EstadoFlujo = normalizeCompraEstadoFlujo(req.EstadoFlujo)
	if strings.TrimSpace(req.FechaSolicitud) == "" {
		req.FechaSolicitud = time.Now().Format("2006-01-02")
	}
	req.FechaNecesidad = strings.TrimSpace(req.FechaNecesidad)
	req.Justificacion = strings.TrimSpace(req.Justificacion)
	req.UsuarioCreador = strings.TrimSpace(req.UsuarioCreador)
	req.TotalEstimado = roundMoney(req.TotalEstimado)
	return req
}

func normalizeCompraRequisicionItem(item EmpresaCompraRequisicionItem) EmpresaCompraRequisicionItem {
	item.ProductoNombre = strings.TrimSpace(item.ProductoNombre)
	item.Unidad = strings.TrimSpace(item.Unidad)
	if item.Unidad == "" {
		item.Unidad = "und"
	}
	item.CantidadSolicitada = math.Max(0, item.CantidadSolicitada)
	item.CantidadRecibida = math.Max(0, item.CantidadRecibida)
	item.CostoEstimado = roundMoney(math.Max(0, item.CostoEstimado))
	item.ProveedorSugerido = strings.TrimSpace(item.ProveedorSugerido)
	item.Especificacion = strings.TrimSpace(item.Especificacion)
	item.Estado = normalizeCompraItemEstado(item.Estado)
	return item
}

func normalizeCompraCotizacion(cot EmpresaCompraCotizacion) EmpresaCompraCotizacion {
	cot.ProveedorNombre = strings.TrimSpace(cot.ProveedorNombre)
	cot.Numero = strings.ToUpper(strings.TrimSpace(cot.Numero))
	if strings.TrimSpace(cot.FechaCotizacion) == "" {
		cot.FechaCotizacion = time.Now().Format("2006-01-02")
	}
	cot.ValidezHasta = strings.TrimSpace(cot.ValidezHasta)
	cot.TiempoEntregaDias = int(math.Max(0, float64(cot.TiempoEntregaDias)))
	cot.Subtotal = roundMoney(math.Max(0, cot.Subtotal))
	cot.Impuestos = roundMoney(math.Max(0, cot.Impuestos))
	if cot.Total <= 0 {
		cot.Total = cot.Subtotal + cot.Impuestos
	}
	cot.Total = roundMoney(math.Max(0, cot.Total))
	cot.CondicionesPago = strings.TrimSpace(cot.CondicionesPago)
	cot.Observaciones = strings.TrimSpace(cot.Observaciones)
	cot.Estado = normalizeCompraCotizacionEstado(cot.Estado)
	cot.UsuarioCreador = strings.TrimSpace(cot.UsuarioCreador)
	return cot
}

func normalizeCompraAprobacion(apr EmpresaCompraAprobacion) EmpresaCompraAprobacion {
	if apr.Nivel <= 0 {
		apr.Nivel = 1
	}
	apr.Aprobador = strings.TrimSpace(apr.Aprobador)
	apr.Decision = strings.ToLower(strings.TrimSpace(apr.Decision))
	if apr.Decision != "rechazada" && apr.Decision != "aprobada" {
		apr.Decision = "pendiente"
	}
	apr.Comentario = strings.TrimSpace(apr.Comentario)
	apr.MontoAutorizado = roundMoney(math.Max(0, apr.MontoAutorizado))
	return apr
}

func normalizeCompraRecepcion(rec EmpresaCompraRecepcion) EmpresaCompraRecepcion {
	rec.ProveedorNombre = strings.TrimSpace(rec.ProveedorNombre)
	rec.Documento = strings.ToUpper(strings.TrimSpace(rec.Documento))
	if strings.TrimSpace(rec.FechaRecepcion) == "" {
		rec.FechaRecepcion = time.Now().Format("2006-01-02")
	}
	rec.EstadoRecepcion = strings.ToLower(strings.TrimSpace(rec.EstadoRecepcion))
	if rec.EstadoRecepcion != "total" {
		rec.EstadoRecepcion = "parcial"
	}
	rec.Responsable = strings.TrimSpace(rec.Responsable)
	rec.Observaciones = strings.TrimSpace(rec.Observaciones)
	rec.UsuarioCreador = strings.TrimSpace(rec.UsuarioCreador)
	return rec
}

func normalizeCompraRecepcionItem(item EmpresaCompraRecepcionItem) EmpresaCompraRecepcionItem {
	item.ProductoNombre = strings.TrimSpace(item.ProductoNombre)
	item.CantidadOrdenada = math.Max(0, item.CantidadOrdenada)
	item.CantidadRecibida = math.Max(0, item.CantidadRecibida)
	item.CantidadPendiente = roundMoney(math.Max(0, item.CantidadOrdenada-item.CantidadRecibida))
	item.CostoUnitario = roundMoney(math.Max(0, item.CostoUnitario))
	item.Lote = strings.TrimSpace(item.Lote)
	item.EstadoCalidad = strings.ToLower(strings.TrimSpace(item.EstadoCalidad))
	if item.EstadoCalidad == "" {
		item.EstadoCalidad = "aprobado"
	}
	return item
}

func normalizeCompraPrioridad(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "baja", "media", "alta", "urgente":
		return v
	default:
		return "media"
	}
}

func normalizeCompraEstadoFlujo(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "borrador", "solicitada", "cotizando", "pendiente_aprobacion", "aprobada", "rechazada", "recibida_parcial", "recibida_total", "cancelada":
		return v
	default:
		return "borrador"
	}
}

func normalizeCompraItemEstado(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "pendiente", "parcial", "recibido", "cancelado":
		return v
	default:
		return "pendiente"
	}
}

func normalizeCompraCotizacionEstado(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "recibida", "evaluacion", "seleccionada", "no_seleccionada", "vencida", "rechazada":
		return v
	default:
		return "recibida"
	}
}
