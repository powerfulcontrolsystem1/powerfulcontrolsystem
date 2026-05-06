package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type EmpresaWMSUbicacion struct {
	ID             int64   `json:"id"`
	EmpresaID      int64   `json:"empresa_id"`
	Codigo         string  `json:"codigo"`
	Bodega         string  `json:"bodega"`
	Zona           string  `json:"zona,omitempty"`
	Pasillo        string  `json:"pasillo,omitempty"`
	Rack           string  `json:"rack,omitempty"`
	Nivel          string  `json:"nivel,omitempty"`
	Posicion       string  `json:"posicion,omitempty"`
	Tipo           string  `json:"tipo"`
	Capacidad      float64 `json:"capacidad"`
	Ocupacion      float64 `json:"ocupacion"`
	Estado         string  `json:"estado"`
	Observaciones  string  `json:"observaciones,omitempty"`
	UsuarioCreador string  `json:"usuario_creador,omitempty"`
	FechaCreacion  string  `json:"fecha_creacion,omitempty"`
}

type EmpresaWMSOrden struct {
	ID                 int64                `json:"id"`
	EmpresaID          int64                `json:"empresa_id"`
	Codigo             string               `json:"codigo"`
	Tipo               string               `json:"tipo"`
	OrigenDocumento    string               `json:"origen_documento,omitempty"`
	Tercero            string               `json:"tercero,omitempty"`
	Cliente            string               `json:"cliente,omitempty"`
	FechaCompromiso    string               `json:"fecha_compromiso,omitempty"`
	Prioridad          string               `json:"prioridad"`
	Responsable        string               `json:"responsable,omitempty"`
	Estado             string               `json:"estado"`
	TotalItems         int                  `json:"total_items"`
	TotalUnidades      float64              `json:"total_unidades"`
	UnidadesPickeadas  float64              `json:"unidades_pickeadas"`
	UnidadesEmpacadas  float64              `json:"unidades_empacadas"`
	ProgresoPicking    float64              `json:"progreso_picking"`
	ProgresoPacking    float64              `json:"progreso_packing"`
	Observaciones      string               `json:"observaciones,omitempty"`
	UsuarioCreador     string               `json:"usuario_creador,omitempty"`
	FechaCreacion      string               `json:"fecha_creacion,omitempty"`
	FechaActualizacion string               `json:"fecha_actualizacion,omitempty"`
	Items              []EmpresaWMSItem     `json:"items,omitempty"`
	Despachos          []EmpresaWMSDespacho `json:"despachos,omitempty"`
}

type EmpresaWMSItem struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	OrdenID            int64   `json:"orden_id"`
	ProductoID         int64   `json:"producto_id,omitempty"`
	ProductoNombre     string  `json:"producto_nombre"`
	SKU                string  `json:"sku,omitempty"`
	UbicacionOrigen    string  `json:"ubicacion_origen,omitempty"`
	UbicacionDestino   string  `json:"ubicacion_destino,omitempty"`
	Lote               string  `json:"lote,omitempty"`
	Serial             string  `json:"serial,omitempty"`
	CantidadSolicitada float64 `json:"cantidad_solicitada"`
	CantidadPickeada   float64 `json:"cantidad_pickeada"`
	CantidadEmpacada   float64 `json:"cantidad_empacada"`
	Estado             string  `json:"estado"`
	Observaciones      string  `json:"observaciones,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
}

type EmpresaWMSDespacho struct {
	ID             int64   `json:"id"`
	EmpresaID      int64   `json:"empresa_id"`
	OrdenID        int64   `json:"orden_id"`
	Codigo         string  `json:"codigo"`
	Transportadora string  `json:"transportadora,omitempty"`
	Guia           string  `json:"guia,omitempty"`
	Conductor      string  `json:"conductor,omitempty"`
	Vehiculo       string  `json:"vehiculo,omitempty"`
	Ruta           string  `json:"ruta,omitempty"`
	Estado         string  `json:"estado"`
	FechaSalida    string  `json:"fecha_salida,omitempty"`
	FechaEntrega   string  `json:"fecha_entrega,omitempty"`
	CostoFlete     float64 `json:"costo_flete"`
	Observaciones  string  `json:"observaciones,omitempty"`
	UsuarioCreador string  `json:"usuario_creador,omitempty"`
	FechaCreacion  string  `json:"fecha_creacion,omitempty"`
}

type EmpresaWMSEvento struct {
	ID             int64  `json:"id"`
	EmpresaID      int64  `json:"empresa_id"`
	ReferenciaTipo string `json:"referencia_tipo"`
	ReferenciaID   int64  `json:"referencia_id"`
	Evento         string `json:"evento"`
	EstadoAnterior string `json:"estado_anterior,omitempty"`
	EstadoNuevo    string `json:"estado_nuevo,omitempty"`
	Detalle        string `json:"detalle,omitempty"`
	Usuario        string `json:"usuario,omitempty"`
	FechaCreacion  string `json:"fecha_creacion,omitempty"`
}

type EmpresaWMSDashboard struct {
	EmpresaID          int64                 `json:"empresa_id"`
	UbicacionesActivas int                   `json:"ubicaciones_activas"`
	OrdenesAbiertas    int                   `json:"ordenes_abiertas"`
	OrdenesPicking     int                   `json:"ordenes_picking"`
	OrdenesPacking     int                   `json:"ordenes_packing"`
	DespachosEnRuta    int                   `json:"despachos_en_ruta"`
	UnidadesPendientes float64               `json:"unidades_pendientes"`
	OrdenesRecientes   []EmpresaWMSOrden     `json:"ordenes_recientes"`
	Ubicaciones        []EmpresaWMSUbicacion `json:"ubicaciones"`
	DespachosRecientes []EmpresaWMSDespacho  `json:"despachos_recientes"`
	EventosRecientes   []EmpresaWMSEvento    `json:"eventos_recientes"`
	Alertas            []string              `json:"alertas"`
}

func EnsureEmpresaWMSSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_wms_ubicaciones (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			bodega TEXT DEFAULT 'Principal',
			zona TEXT,
			pasillo TEXT,
			rack TEXT,
			nivel TEXT,
			posicion TEXT,
			tipo TEXT DEFAULT 'almacenamiento',
			capacidad REAL DEFAULT 0,
			ocupacion REAL DEFAULT 0,
			estado TEXT DEFAULT 'activa',
			observaciones TEXT,
			usuario_creador TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(empresa_id,codigo)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_wms_ubicaciones_empresa ON empresa_wms_ubicaciones(empresa_id,bodega,estado)`,
		`CREATE TABLE IF NOT EXISTS empresa_wms_ordenes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			tipo TEXT DEFAULT 'picking',
			origen_documento TEXT,
			tercero TEXT,
			cliente TEXT,
			fecha_compromiso TEXT,
			prioridad TEXT DEFAULT 'normal',
			responsable TEXT,
			estado TEXT DEFAULT 'borrador',
			observaciones TEXT,
			usuario_creador TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(empresa_id,codigo)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_wms_ordenes_empresa ON empresa_wms_ordenes(empresa_id,tipo,estado,fecha_compromiso)`,
		`CREATE TABLE IF NOT EXISTS empresa_wms_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			orden_id INTEGER NOT NULL,
			producto_id INTEGER DEFAULT 0,
			producto_nombre TEXT NOT NULL,
			sku TEXT,
			ubicacion_origen TEXT,
			ubicacion_destino TEXT,
			lote TEXT,
			serial TEXT,
			cantidad_solicitada REAL DEFAULT 0,
			cantidad_pickeada REAL DEFAULT 0,
			cantidad_empacada REAL DEFAULT 0,
			estado TEXT DEFAULT 'pendiente',
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS ix_wms_items_orden ON empresa_wms_items(empresa_id,orden_id,estado)`,
		`CREATE TABLE IF NOT EXISTS empresa_wms_despachos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			orden_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			transportadora TEXT,
			guia TEXT,
			conductor TEXT,
			vehiculo TEXT,
			ruta TEXT,
			estado TEXT DEFAULT 'programado',
			fecha_salida TEXT,
			fecha_entrega TEXT,
			costo_flete REAL DEFAULT 0,
			observaciones TEXT,
			usuario_creador TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(empresa_id,codigo)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_wms_despachos_empresa ON empresa_wms_despachos(empresa_id,orden_id,estado)`,
		`CREATE TABLE IF NOT EXISTS empresa_wms_eventos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			referencia_tipo TEXT NOT NULL,
			referencia_id INTEGER DEFAULT 0,
			evento TEXT NOT NULL,
			estado_anterior TEXT,
			estado_nuevo TEXT,
			detalle TEXT,
			usuario TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS ix_wms_eventos_empresa ON empresa_wms_eventos(empresa_id,referencia_tipo,referencia_id)`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func BuildEmpresaWMSDashboard(dbConn *sql.DB, empresaID int64) (EmpresaWMSDashboard, error) {
	if err := EnsureEmpresaWMSSchema(dbConn); err != nil {
		return EmpresaWMSDashboard{}, err
	}
	d := EmpresaWMSDashboard{EmpresaID: empresaID}
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_wms_ubicaciones WHERE empresa_id=? AND estado='activa'`, empresaID).Scan(&d.UbicacionesActivas)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_wms_ordenes WHERE empresa_id=? AND estado NOT IN ('cerrada','cancelada')`, empresaID).Scan(&d.OrdenesAbiertas)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_wms_ordenes WHERE empresa_id=? AND estado='en_picking'`, empresaID).Scan(&d.OrdenesPicking)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_wms_ordenes WHERE empresa_id=? AND estado='en_packing'`, empresaID).Scan(&d.OrdenesPacking)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_wms_despachos WHERE empresa_id=? AND estado='en_ruta'`, empresaID).Scan(&d.DespachosEnRuta)
	_ = QueryRowCompat(dbConn, `SELECT COALESCE(SUM(CASE WHEN cantidad_solicitada>cantidad_pickeada THEN cantidad_solicitada-cantidad_pickeada ELSE 0 END),0) FROM empresa_wms_items WHERE empresa_id=? AND estado NOT IN ('completado','cancelado')`, empresaID).Scan(&d.UnidadesPendientes)
	d.OrdenesRecientes, _ = ListEmpresaWMSOrdenes(dbConn, empresaID, "", "", 50)
	d.Ubicaciones, _ = ListEmpresaWMSUbicaciones(dbConn, empresaID, "", 200)
	d.DespachosRecientes, _ = ListEmpresaWMSDespachos(dbConn, empresaID, 0, 50)
	d.EventosRecientes, _ = ListEmpresaWMSEventos(dbConn, empresaID, 50)
	if d.UbicacionesActivas == 0 {
		d.Alertas = append(d.Alertas, "No hay ubicaciones WMS activas para operar picking, packing o conteos.")
	}
	if d.UnidadesPendientes > 0 {
		d.Alertas = append(d.Alertas, "Hay unidades pendientes por pickear en ordenes abiertas.")
	}
	if d.DespachosEnRuta > 0 {
		d.Alertas = append(d.Alertas, "Hay despachos en ruta pendientes de confirmar entrega.")
	}
	if len(d.Alertas) == 0 {
		d.Alertas = append(d.Alertas, "Operacion WMS sin alertas criticas.")
	}
	return d, nil
}

func UpsertEmpresaWMSUbicacion(dbConn *sql.DB, x EmpresaWMSUbicacion) (int64, error) {
	if err := EnsureEmpresaWMSSchema(dbConn); err != nil {
		return 0, err
	}
	x = normalizeWMSUbicacion(x)
	if x.EmpresaID <= 0 || x.Codigo == "" {
		return 0, errors.New("empresa_id y codigo son requeridos")
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_wms_ubicaciones (empresa_id,codigo,bodega,zona,pasillo,rack,nivel,posicion,tipo,capacidad,ocupacion,estado,observaciones,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (empresa_id,codigo) DO UPDATE SET bodega=EXCLUDED.bodega,zona=EXCLUDED.zona,pasillo=EXCLUDED.pasillo,rack=EXCLUDED.rack,nivel=EXCLUDED.nivel,posicion=EXCLUDED.posicion,tipo=EXCLUDED.tipo,capacidad=EXCLUDED.capacidad,ocupacion=EXCLUDED.ocupacion,estado=EXCLUDED.estado,observaciones=EXCLUDED.observaciones,usuario_creador=EXCLUDED.usuario_creador`,
		x.EmpresaID, x.Codigo, x.Bodega, x.Zona, x.Pasillo, x.Rack, x.Nivel, x.Posicion, x.Tipo, x.Capacidad, x.Ocupacion, x.Estado, x.Observaciones, x.UsuarioCreador)
}

func UpsertEmpresaWMSOrden(dbConn *sql.DB, x EmpresaWMSOrden) (int64, error) {
	if err := EnsureEmpresaWMSSchema(dbConn); err != nil {
		return 0, err
	}
	x = normalizeWMSOrden(x)
	if x.EmpresaID <= 0 || x.Codigo == "" {
		return 0, errors.New("empresa_id y codigo son requeridos")
	}
	if x.ID > 0 {
		old := ""
		_ = QueryRowCompat(dbConn, `SELECT COALESCE(estado,'') FROM empresa_wms_ordenes WHERE empresa_id=? AND id=?`, x.EmpresaID, x.ID).Scan(&old)
		_, err := ExecCompat(dbConn, `UPDATE empresa_wms_ordenes SET codigo=?,tipo=?,origen_documento=?,tercero=?,cliente=?,fecha_compromiso=?,prioridad=?,responsable=?,estado=?,observaciones=?,usuario_creador=?,fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`,
			x.Codigo, x.Tipo, x.OrigenDocumento, x.Tercero, x.Cliente, x.FechaCompromiso, x.Prioridad, x.Responsable, x.Estado, x.Observaciones, x.UsuarioCreador, x.EmpresaID, x.ID)
		if err == nil && old != x.Estado {
			_ = registrarWMSEvento(dbConn, x.EmpresaID, "orden", x.ID, "cambio_estado", old, x.Estado, "Cambio manual de estado", x.UsuarioCreador)
		}
		return x.ID, err
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_wms_ordenes (empresa_id,codigo,tipo,origen_documento,tercero,cliente,fecha_compromiso,prioridad,responsable,estado,observaciones,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (empresa_id,codigo) DO UPDATE SET tipo=EXCLUDED.tipo,origen_documento=EXCLUDED.origen_documento,tercero=EXCLUDED.tercero,cliente=EXCLUDED.cliente,fecha_compromiso=EXCLUDED.fecha_compromiso,prioridad=EXCLUDED.prioridad,responsable=EXCLUDED.responsable,estado=EXCLUDED.estado,observaciones=EXCLUDED.observaciones,usuario_creador=EXCLUDED.usuario_creador,fecha_actualizacion=CURRENT_TIMESTAMP`,
		x.EmpresaID, x.Codigo, x.Tipo, x.OrigenDocumento, x.Tercero, x.Cliente, x.FechaCompromiso, x.Prioridad, x.Responsable, x.Estado, x.Observaciones, x.UsuarioCreador)
	if err == nil {
		_ = registrarWMSEvento(dbConn, x.EmpresaID, "orden", id, "orden_guardada", "", x.Estado, x.Codigo, x.UsuarioCreador)
	}
	return id, err
}

func CreateEmpresaWMSItem(dbConn *sql.DB, x EmpresaWMSItem, usuario string) (int64, error) {
	if err := EnsureEmpresaWMSSchema(dbConn); err != nil {
		return 0, err
	}
	x = normalizeWMSItem(x)
	if x.EmpresaID <= 0 || x.OrdenID <= 0 || x.ProductoNombre == "" || x.CantidadSolicitada <= 0 {
		return 0, errors.New("orden, producto y cantidad son requeridos")
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_wms_items (empresa_id,orden_id,producto_id,producto_nombre,sku,ubicacion_origen,ubicacion_destino,lote,serial,cantidad_solicitada,cantidad_pickeada,cantidad_empacada,estado,observaciones)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		x.EmpresaID, x.OrdenID, x.ProductoID, x.ProductoNombre, x.SKU, x.UbicacionOrigen, x.UbicacionDestino, x.Lote, x.Serial, x.CantidadSolicitada, x.CantidadPickeada, x.CantidadEmpacada, x.Estado, x.Observaciones)
	if err == nil {
		_ = registrarWMSEvento(dbConn, x.EmpresaID, "item", id, "item_agregado", "", x.Estado, x.ProductoNombre, usuario)
	}
	return id, err
}

func ActualizarEmpresaWMSItemAvance(dbConn *sql.DB, empresaID, itemID int64, pickeada, empacada float64, estado, usuario string) error {
	if err := EnsureEmpresaWMSSchema(dbConn); err != nil {
		return err
	}
	var old string
	var solicitada float64
	if err := QueryRowCompat(dbConn, `SELECT COALESCE(estado,'pendiente'),COALESCE(cantidad_solicitada,0) FROM empresa_wms_items WHERE empresa_id=? AND id=?`, empresaID, itemID).Scan(&old, &solicitada); err != nil {
		return err
	}
	if pickeada < 0 {
		pickeada = 0
	}
	if empacada < 0 {
		empacada = 0
	}
	if solicitada > 0 {
		if pickeada > solicitada {
			pickeada = solicitada
		}
		if empacada > pickeada {
			empacada = pickeada
		}
	}
	estado = normalizeWMSItemEstado(estado)
	if estado == "pendiente" {
		estado = inferWMSItemEstado(solicitada, pickeada, empacada)
	}
	_, err := ExecCompat(dbConn, `UPDATE empresa_wms_items SET cantidad_pickeada=?,cantidad_empacada=?,estado=? WHERE empresa_id=? AND id=?`, pickeada, empacada, estado, empresaID, itemID)
	if err == nil {
		_ = registrarWMSEvento(dbConn, empresaID, "item", itemID, "avance_item", old, estado, fmt.Sprintf("pick %.2f pack %.2f", pickeada, empacada), usuario)
	}
	return err
}

func UpsertEmpresaWMSDespacho(dbConn *sql.DB, x EmpresaWMSDespacho) (int64, error) {
	if err := EnsureEmpresaWMSSchema(dbConn); err != nil {
		return 0, err
	}
	x = normalizeWMSDespacho(x)
	if x.EmpresaID <= 0 || x.OrdenID <= 0 || x.Codigo == "" {
		return 0, errors.New("orden y codigo de despacho son requeridos")
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_wms_despachos (empresa_id,orden_id,codigo,transportadora,guia,conductor,vehiculo,ruta,estado,fecha_salida,fecha_entrega,costo_flete,observaciones,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (empresa_id,codigo) DO UPDATE SET orden_id=EXCLUDED.orden_id,transportadora=EXCLUDED.transportadora,guia=EXCLUDED.guia,conductor=EXCLUDED.conductor,vehiculo=EXCLUDED.vehiculo,ruta=EXCLUDED.ruta,estado=EXCLUDED.estado,fecha_salida=EXCLUDED.fecha_salida,fecha_entrega=EXCLUDED.fecha_entrega,costo_flete=EXCLUDED.costo_flete,observaciones=EXCLUDED.observaciones,usuario_creador=EXCLUDED.usuario_creador`,
		x.EmpresaID, x.OrdenID, x.Codigo, x.Transportadora, x.Guia, x.Conductor, x.Vehiculo, x.Ruta, x.Estado, x.FechaSalida, x.FechaEntrega, x.CostoFlete, x.Observaciones, x.UsuarioCreador)
	if err == nil {
		_ = registrarWMSEvento(dbConn, x.EmpresaID, "despacho", id, "despacho_guardado", "", x.Estado, x.Codigo, x.UsuarioCreador)
	}
	return id, err
}

func ListEmpresaWMSUbicaciones(dbConn *sql.DB, empresaID int64, estado string, limit int) ([]EmpresaWMSUbicacion, error) {
	if err := EnsureEmpresaWMSSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 1000 {
		limit = 300
	}
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if strings.TrimSpace(estado) != "" && strings.ToLower(strings.TrimSpace(estado)) != "todos" {
		where += " AND estado=?"
		args = append(args, normalizeWMSActivoEstado(estado))
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,codigo,COALESCE(bodega,''),COALESCE(zona,''),COALESCE(pasillo,''),COALESCE(rack,''),COALESCE(nivel,''),COALESCE(posicion,''),COALESCE(tipo,''),COALESCE(capacidad,0),COALESCE(ocupacion,0),COALESCE(estado,''),COALESCE(observaciones,''),COALESCE(usuario_creador,''),COALESCE(fecha_creacion,'') FROM empresa_wms_ubicaciones WHERE %s ORDER BY bodega,zona,pasillo,rack,nivel,posicion LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaWMSUbicacion{}
	for rows.Next() {
		var x EmpresaWMSUbicacion
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Codigo, &x.Bodega, &x.Zona, &x.Pasillo, &x.Rack, &x.Nivel, &x.Posicion, &x.Tipo, &x.Capacidad, &x.Ocupacion, &x.Estado, &x.Observaciones, &x.UsuarioCreador, &x.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func ListEmpresaWMSOrdenes(dbConn *sql.DB, empresaID int64, tipo, estado string, limit int) ([]EmpresaWMSOrden, error) {
	if err := EnsureEmpresaWMSSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 1000 {
		limit = 300
	}
	args := []interface{}{empresaID}
	where := "o.empresa_id=?"
	if tipo = normalizeWMSTipoOrden(tipo); tipo != "" && tipo != "todos" {
		where += " AND o.tipo=?"
		args = append(args, tipo)
	}
	if estado = normalizeWMSOrdenEstado(estado); estado != "" && estado != "todos" {
		where += " AND o.estado=?"
		args = append(args, estado)
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT o.id,o.empresa_id,o.codigo,COALESCE(o.tipo,''),COALESCE(o.origen_documento,''),COALESCE(o.tercero,''),COALESCE(o.cliente,''),COALESCE(o.fecha_compromiso,''),COALESCE(o.prioridad,''),COALESCE(o.responsable,''),COALESCE(o.estado,''),COALESCE(o.observaciones,''),COALESCE(o.usuario_creador,''),COALESCE(o.fecha_creacion,''),COALESCE(o.fecha_actualizacion,''),COUNT(i.id),COALESCE(SUM(i.cantidad_solicitada),0),COALESCE(SUM(i.cantidad_pickeada),0),COALESCE(SUM(i.cantidad_empacada),0) FROM empresa_wms_ordenes o LEFT JOIN empresa_wms_items i ON i.empresa_id=o.empresa_id AND i.orden_id=o.id WHERE %s GROUP BY o.id ORDER BY o.id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaWMSOrden{}
	for rows.Next() {
		var x EmpresaWMSOrden
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Codigo, &x.Tipo, &x.OrigenDocumento, &x.Tercero, &x.Cliente, &x.FechaCompromiso, &x.Prioridad, &x.Responsable, &x.Estado, &x.Observaciones, &x.UsuarioCreador, &x.FechaCreacion, &x.FechaActualizacion, &x.TotalItems, &x.TotalUnidades, &x.UnidadesPickeadas, &x.UnidadesEmpacadas); err != nil {
			return nil, err
		}
		x.ProgresoPicking, x.ProgresoPacking = calcularProgresoWMS(x.TotalUnidades, x.UnidadesPickeadas, x.UnidadesEmpacadas)
		out = append(out, x)
	}
	return out, rows.Err()
}

func GetEmpresaWMSOrden(dbConn *sql.DB, empresaID, id int64) (EmpresaWMSOrden, error) {
	rows, err := ListEmpresaWMSOrdenes(dbConn, empresaID, "", "", 1000)
	if err != nil {
		return EmpresaWMSOrden{}, err
	}
	for _, x := range rows {
		if x.ID == id {
			x.Items, _ = ListEmpresaWMSItems(dbConn, empresaID, id)
			x.Despachos, _ = ListEmpresaWMSDespachos(dbConn, empresaID, id, 100)
			return x, nil
		}
	}
	return EmpresaWMSOrden{}, sql.ErrNoRows
}

func ListEmpresaWMSItems(dbConn *sql.DB, empresaID, ordenID int64) ([]EmpresaWMSItem, error) {
	if err := EnsureEmpresaWMSSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,orden_id,COALESCE(producto_id,0),producto_nombre,COALESCE(sku,''),COALESCE(ubicacion_origen,''),COALESCE(ubicacion_destino,''),COALESCE(lote,''),COALESCE(serial,''),COALESCE(cantidad_solicitada,0),COALESCE(cantidad_pickeada,0),COALESCE(cantidad_empacada,0),COALESCE(estado,''),COALESCE(observaciones,''),COALESCE(fecha_creacion,'') FROM empresa_wms_items WHERE empresa_id=? AND orden_id=? ORDER BY id`, empresaID, ordenID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaWMSItem{}
	for rows.Next() {
		var x EmpresaWMSItem
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.OrdenID, &x.ProductoID, &x.ProductoNombre, &x.SKU, &x.UbicacionOrigen, &x.UbicacionDestino, &x.Lote, &x.Serial, &x.CantidadSolicitada, &x.CantidadPickeada, &x.CantidadEmpacada, &x.Estado, &x.Observaciones, &x.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func ListEmpresaWMSDespachos(dbConn *sql.DB, empresaID int64, ordenID int64, limit int) ([]EmpresaWMSDespacho, error) {
	if err := EnsureEmpresaWMSSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 1000 {
		limit = 300
	}
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if ordenID > 0 {
		where += " AND orden_id=?"
		args = append(args, ordenID)
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,orden_id,codigo,COALESCE(transportadora,''),COALESCE(guia,''),COALESCE(conductor,''),COALESCE(vehiculo,''),COALESCE(ruta,''),COALESCE(estado,''),COALESCE(fecha_salida,''),COALESCE(fecha_entrega,''),COALESCE(costo_flete,0),COALESCE(observaciones,''),COALESCE(usuario_creador,''),COALESCE(fecha_creacion,'') FROM empresa_wms_despachos WHERE %s ORDER BY id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaWMSDespacho{}
	for rows.Next() {
		var x EmpresaWMSDespacho
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.OrdenID, &x.Codigo, &x.Transportadora, &x.Guia, &x.Conductor, &x.Vehiculo, &x.Ruta, &x.Estado, &x.FechaSalida, &x.FechaEntrega, &x.CostoFlete, &x.Observaciones, &x.UsuarioCreador, &x.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func ListEmpresaWMSEventos(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaWMSEvento, error) {
	if err := EnsureEmpresaWMSSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,referencia_tipo,COALESCE(referencia_id,0),evento,COALESCE(estado_anterior,''),COALESCE(estado_nuevo,''),COALESCE(detalle,''),COALESCE(usuario,''),COALESCE(fecha_creacion,'') FROM empresa_wms_eventos WHERE empresa_id=? ORDER BY id DESC LIMIT %d`, limit), empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaWMSEvento{}
	for rows.Next() {
		var x EmpresaWMSEvento
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.ReferenciaTipo, &x.ReferenciaID, &x.Evento, &x.EstadoAnterior, &x.EstadoNuevo, &x.Detalle, &x.Usuario, &x.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func SeedEmpresaWMSDemo(dbConn *sql.DB, empresaID int64, usuario string) error {
	if err := EnsureEmpresaWMSSchema(dbConn); err != nil {
		return err
	}
	for _, u := range []EmpresaWMSUbicacion{
		{EmpresaID: empresaID, Codigo: "BOD-A-P01-R01-N01", Bodega: "Principal", Zona: "A", Pasillo: "P01", Rack: "R01", Nivel: "N01", Posicion: "01", Tipo: "picking", Capacidad: 120, Ocupacion: 62, Estado: "activa", UsuarioCreador: usuario},
		{EmpresaID: empresaID, Codigo: "BOD-A-P02-R03-N02", Bodega: "Principal", Zona: "A", Pasillo: "P02", Rack: "R03", Nivel: "N02", Posicion: "04", Tipo: "almacenamiento", Capacidad: 260, Ocupacion: 140, Estado: "activa", UsuarioCreador: usuario},
		{EmpresaID: empresaID, Codigo: "PACK-01", Bodega: "Principal", Zona: "Packing", Tipo: "packing", Capacidad: 80, Ocupacion: 25, Estado: "activa", UsuarioCreador: usuario},
	} {
		_, _ = UpsertEmpresaWMSUbicacion(dbConn, u)
	}
	ordenID, err := UpsertEmpresaWMSOrden(dbConn, EmpresaWMSOrden{EmpresaID: empresaID, Codigo: fmt.Sprintf("WMS-%s-001", time.Now().Format("20060102")), Tipo: "despacho", OrigenDocumento: "pedido_demo", Cliente: "Cliente mostrador", FechaCompromiso: time.Now().AddDate(0, 0, 1).Format("2006-01-02"), Prioridad: "alta", Responsable: "Bodega", Estado: "en_picking", Observaciones: "Orden demo WMS", UsuarioCreador: usuario})
	if err != nil {
		return err
	}
	_, _ = CreateEmpresaWMSItem(dbConn, EmpresaWMSItem{EmpresaID: empresaID, OrdenID: ordenID, ProductoNombre: "Producto demo WMS", SKU: "WMS-DEMO-001", UbicacionOrigen: "BOD-A-P01-R01-N01", UbicacionDestino: "PACK-01", CantidadSolicitada: 12, CantidadPickeada: 6, CantidadEmpacada: 0, Estado: "en_picking"}, usuario)
	_, _ = CreateEmpresaWMSItem(dbConn, EmpresaWMSItem{EmpresaID: empresaID, OrdenID: ordenID, ProductoNombre: "Combo embalado demo", SKU: "WMS-DEMO-002", UbicacionOrigen: "BOD-A-P02-R03-N02", UbicacionDestino: "PACK-01", CantidadSolicitada: 4, Estado: "pendiente"}, usuario)
	_, _ = UpsertEmpresaWMSDespacho(dbConn, EmpresaWMSDespacho{EmpresaID: empresaID, OrdenID: ordenID, Codigo: fmt.Sprintf("DSP-%s-001", time.Now().Format("20060102")), Transportadora: "Mensajeria interna", Conductor: "Domiciliario demo", Vehiculo: "Moto 01", Ruta: "Ruta centro", Estado: "programado", FechaSalida: time.Now().AddDate(0, 0, 1).Format("2006-01-02"), CostoFlete: 8500, UsuarioCreador: usuario})
	return nil
}

func normalizeWMSUbicacion(x EmpresaWMSUbicacion) EmpresaWMSUbicacion {
	x.Codigo = normalizeWMSCodigo(x.Codigo)
	if x.Bodega = strings.TrimSpace(x.Bodega); x.Bodega == "" {
		x.Bodega = "Principal"
	}
	x.Tipo = normalizeWMSTipoUbicacion(x.Tipo)
	x.Estado = normalizeWMSActivoEstado(x.Estado)
	if x.Capacidad < 0 {
		x.Capacidad = 0
	}
	if x.Ocupacion < 0 {
		x.Ocupacion = 0
	}
	return x
}

func normalizeWMSOrden(x EmpresaWMSOrden) EmpresaWMSOrden {
	x.Codigo = normalizeWMSCodigo(x.Codigo)
	if x.Codigo == "" {
		x.Codigo = "WMS-" + time.Now().Format("20060102150405")
	}
	x.Tipo = normalizeWMSTipoOrden(x.Tipo)
	x.Estado = normalizeWMSOrdenEstado(x.Estado)
	x.Prioridad = normalizeWMSPrioridad(x.Prioridad)
	return x
}

func normalizeWMSItem(x EmpresaWMSItem) EmpresaWMSItem {
	x.SKU = strings.ToUpper(strings.TrimSpace(x.SKU))
	x.UbicacionOrigen = normalizeWMSCodigo(x.UbicacionOrigen)
	x.UbicacionDestino = normalizeWMSCodigo(x.UbicacionDestino)
	x.Estado = normalizeWMSItemEstado(x.Estado)
	if x.CantidadSolicitada < 0 {
		x.CantidadSolicitada = 0
	}
	if x.CantidadPickeada < 0 {
		x.CantidadPickeada = 0
	}
	if x.CantidadEmpacada < 0 {
		x.CantidadEmpacada = 0
	}
	return x
}

func normalizeWMSDespacho(x EmpresaWMSDespacho) EmpresaWMSDespacho {
	x.Codigo = normalizeWMSCodigo(x.Codigo)
	if x.Codigo == "" {
		x.Codigo = "DSP-" + time.Now().Format("20060102150405")
	}
	x.Estado = normalizeWMSDespachoEstado(x.Estado)
	if x.CostoFlete < 0 {
		x.CostoFlete = 0
	}
	return x
}

func normalizeWMSCodigo(v string) string {
	v = strings.ToUpper(strings.TrimSpace(v))
	repl := strings.NewReplacer(" ", "-", "_", "-", "/", "-", "\\", "-", ".", "-", "--", "-")
	v = repl.Replace(v)
	for strings.Contains(v, "--") {
		v = strings.ReplaceAll(v, "--", "-")
	}
	return strings.Trim(v, "-")
}

func normalizeWMSTipoUbicacion(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "picking", "packing", "almacenamiento", "recepcion", "despacho", "cuarentena", "devolucion", "conteo":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "almacenamiento"
	}
}

func normalizeWMSTipoOrden(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "picking", "packing", "despacho", "conteo", "reabastecimiento", "traslado", "devolucion", "todos":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "picking"
	}
}

func normalizeWMSOrdenEstado(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "borrador", "liberada", "en_picking", "en_packing", "lista_despacho", "despachada", "cerrada", "cancelada", "todos":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "borrador"
	}
}

func normalizeWMSItemEstado(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "pendiente", "en_picking", "pickeado", "en_packing", "empacado", "completado", "cancelado":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "pendiente"
	}
}

func normalizeWMSDespachoEstado(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "programado", "en_ruta", "entregado", "devuelto", "cancelado":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "programado"
	}
}

func normalizeWMSActivoEstado(v string) string {
	if strings.EqualFold(strings.TrimSpace(v), "inactiva") || strings.EqualFold(strings.TrimSpace(v), "inactivo") {
		return "inactiva"
	}
	return "activa"
}

func normalizeWMSPrioridad(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "baja", "normal", "alta", "urgente":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "normal"
	}
}

func calcularProgresoWMS(total, pickeadas, empacadas float64) (float64, float64) {
	if total <= 0 {
		return 0, 0
	}
	if pickeadas < 0 {
		pickeadas = 0
	}
	if empacadas < 0 {
		empacadas = 0
	}
	if pickeadas > total {
		pickeadas = total
	}
	if empacadas > total {
		empacadas = total
	}
	return float64(int64((pickeadas/total)*10000+0.5)) / 100, float64(int64((empacadas/total)*10000+0.5)) / 100
}

func inferWMSItemEstado(total, pickeada, empacada float64) string {
	if total <= 0 {
		return "pendiente"
	}
	if empacada >= total {
		return "completado"
	}
	if empacada > 0 {
		return "en_packing"
	}
	if pickeada >= total {
		return "pickeado"
	}
	if pickeada > 0 {
		return "en_picking"
	}
	return "pendiente"
}

func registrarWMSEvento(dbConn *sql.DB, empresaID int64, refTipo string, refID int64, evento, old, nuevo, detalle, usuario string) error {
	_, err := insertSQLCompat(dbConn, `INSERT INTO empresa_wms_eventos (empresa_id,referencia_tipo,referencia_id,evento,estado_anterior,estado_nuevo,detalle,usuario) VALUES (?,?,?,?,?,?,?,?)`, empresaID, refTipo, refID, evento, old, nuevo, detalle, usuario)
	return err
}
