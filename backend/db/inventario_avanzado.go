package db

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

type EmpresaInventarioLoteAvanzado struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	ProductoID         int64   `json:"producto_id"`
	ProductoNombre     string  `json:"producto_nombre,omitempty"`
	BodegaID           int64   `json:"bodega_id"`
	BodegaNombre       string  `json:"bodega_nombre,omitempty"`
	LoteCodigo         string  `json:"lote_codigo"`
	FechaFabricacion   string  `json:"fecha_fabricacion,omitempty"`
	FechaVencimiento   string  `json:"fecha_vencimiento,omitempty"`
	CantidadInicial    float64 `json:"cantidad_inicial"`
	CantidadDisponible float64 `json:"cantidad_disponible"`
	CantidadReservada  float64 `json:"cantidad_reservada"`
	CantidadLibre      float64 `json:"cantidad_libre"`
	CostoUnitario      float64 `json:"costo_unitario"`
	ValorDisponible    float64 `json:"valor_disponible"`
	EstadoCalidad      string  `json:"estado_calidad"`
	Proveedor          string  `json:"proveedor,omitempty"`
	DocumentoRef       string  `json:"documento_ref,omitempty"`
	UbicacionInterna   string  `json:"ubicacion_interna,omitempty"`
	EstadoVencimiento  string  `json:"estado_vencimiento"`
	DiasParaVencer     int     `json:"dias_para_vencer"`
	Estado             string  `json:"estado"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
}

type EmpresaInventarioSerialAvanzado struct {
	ID               int64  `json:"id"`
	EmpresaID        int64  `json:"empresa_id"`
	LoteID           int64  `json:"lote_id,omitempty"`
	ProductoID       int64  `json:"producto_id"`
	ProductoNombre   string `json:"producto_nombre,omitempty"`
	BodegaID         int64  `json:"bodega_id"`
	BodegaNombre     string `json:"bodega_nombre,omitempty"`
	Serial           string `json:"serial"`
	EstadoOperativo  string `json:"estado_operativo"`
	EstadoInventario string `json:"estado_inventario"`
	FechaIngreso     string `json:"fecha_ingreso,omitempty"`
	GarantiaHasta    string `json:"garantia_hasta,omitempty"`
	ClienteReserva   string `json:"cliente_reserva,omitempty"`
	Observaciones    string `json:"observaciones,omitempty"`
	UsuarioCreador   string `json:"usuario_creador,omitempty"`
	FechaCreacion    string `json:"fecha_creacion,omitempty"`
}

type EmpresaInventarioReservaAvanzada struct {
	ID             int64   `json:"id"`
	EmpresaID      int64   `json:"empresa_id"`
	ProductoID     int64   `json:"producto_id"`
	ProductoNombre string  `json:"producto_nombre,omitempty"`
	BodegaID       int64   `json:"bodega_id"`
	BodegaNombre   string  `json:"bodega_nombre,omitempty"`
	LoteID         int64   `json:"lote_id,omitempty"`
	LoteCodigo     string  `json:"lote_codigo,omitempty"`
	SerialID       int64   `json:"serial_id,omitempty"`
	Serial         string  `json:"serial,omitempty"`
	Cantidad       float64 `json:"cantidad"`
	OrigenModulo   string  `json:"origen_modulo"`
	OrigenRef      string  `json:"origen_ref"`
	ClienteNombre  string  `json:"cliente_nombre,omitempty"`
	FechaReserva   string  `json:"fecha_reserva"`
	FechaExpira    string  `json:"fecha_expira,omitempty"`
	Estado         string  `json:"estado"`
	UsuarioCreador string  `json:"usuario_creador,omitempty"`
	FechaCreacion  string  `json:"fecha_creacion,omitempty"`
}

type EmpresaInventarioValorizacionAvanzada struct {
	ProductoID         int64   `json:"producto_id"`
	ProductoNombre     string  `json:"producto_nombre"`
	BodegaID           int64   `json:"bodega_id"`
	BodegaNombre       string  `json:"bodega_nombre"`
	CantidadDisponible float64 `json:"cantidad_disponible"`
	CantidadReservada  float64 `json:"cantidad_reservada"`
	CantidadLibre      float64 `json:"cantidad_libre"`
	CostoPromedio      float64 `json:"costo_promedio"`
	ValorDisponible    float64 `json:"valor_disponible"`
	LotesActivos       int     `json:"lotes_activos"`
}

type EmpresaInventarioAvanzadoDashboard struct {
	EmpresaID          int64                                   `json:"empresa_id"`
	LotesActivos       int                                     `json:"lotes_activos"`
	SerialesActivos    int                                     `json:"seriales_activos"`
	ReservasActivas    int                                     `json:"reservas_activas"`
	UnidadesReservadas float64                                 `json:"unidades_reservadas"`
	ValorDisponible    float64                                 `json:"valor_disponible"`
	LotesPorVencer     int                                     `json:"lotes_por_vencer"`
	LotesVencidos      int                                     `json:"lotes_vencidos"`
	UltimosLotes       []EmpresaInventarioLoteAvanzado         `json:"ultimos_lotes"`
	Valorizacion       []EmpresaInventarioValorizacionAvanzada `json:"valorizacion"`
}

func EnsureEmpresaInventarioAvanzadoSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_inventario_lotes_avanzados (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			producto_id INTEGER NOT NULL,
			bodega_id INTEGER NOT NULL,
			lote_codigo TEXT NOT NULL,
			fecha_fabricacion TEXT,
			fecha_vencimiento TEXT,
			cantidad_inicial REAL DEFAULT 0,
			cantidad_disponible REAL DEFAULT 0,
			costo_unitario REAL DEFAULT 0,
			estado_calidad TEXT DEFAULT 'liberado',
			proveedor TEXT,
			documento_ref TEXT,
			ubicacion_interna TEXT,
			estado TEXT DEFAULT 'activo',
			usuario_creador TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(empresa_id, producto_id, bodega_id, lote_codigo)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_inv_lotes_av_empresa_venc ON empresa_inventario_lotes_avanzados(empresa_id, fecha_vencimiento, estado)`,
		`CREATE TABLE IF NOT EXISTS empresa_inventario_seriales_avanzados (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			lote_id INTEGER DEFAULT 0,
			producto_id INTEGER NOT NULL,
			bodega_id INTEGER NOT NULL,
			serial TEXT NOT NULL,
			estado_operativo TEXT DEFAULT 'operativo',
			estado_inventario TEXT DEFAULT 'disponible',
			fecha_ingreso TEXT,
			garantia_hasta TEXT,
			cliente_reserva TEXT,
			observaciones TEXT,
			usuario_creador TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(empresa_id, serial)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_inv_seriales_av_producto ON empresa_inventario_seriales_avanzados(empresa_id, producto_id, bodega_id)`,
		`CREATE TABLE IF NOT EXISTS empresa_inventario_reservas_avanzadas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			producto_id INTEGER NOT NULL,
			bodega_id INTEGER NOT NULL,
			lote_id INTEGER DEFAULT 0,
			serial_id INTEGER DEFAULT 0,
			cantidad REAL DEFAULT 0,
			origen_modulo TEXT DEFAULT 'manual',
			origen_ref TEXT DEFAULT '',
			cliente_nombre TEXT,
			fecha_reserva TEXT NOT NULL,
			fecha_expira TEXT,
			estado TEXT DEFAULT 'activa',
			usuario_creador TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS ix_inv_reservas_av_estado ON empresa_inventario_reservas_avanzadas(empresa_id, estado, producto_id, bodega_id)`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func CreateEmpresaInventarioLoteAvanzado(dbConn *sql.DB, lote EmpresaInventarioLoteAvanzado) (int64, error) {
	lote = normalizeInventarioLoteAvanzado(lote)
	if lote.EmpresaID <= 0 || lote.ProductoID <= 0 || lote.BodegaID <= 0 || lote.LoteCodigo == "" {
		return 0, errors.New("empresa, producto, bodega y lote son requeridos")
	}
	if lote.CantidadInicial <= 0 {
		return 0, errors.New("cantidad inicial debe ser mayor a cero")
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if err := validateProductoEmpresaTx(tx, lote.EmpresaID, lote.ProductoID); err != nil {
		return 0, err
	}
	if err := validateBodegaEmpresaTx(tx, lote.EmpresaID, lote.BodegaID); err != nil {
		return 0, err
	}
	id, err := insertTxSQLCompat(tx, `INSERT INTO empresa_inventario_lotes_avanzados
		(empresa_id,producto_id,bodega_id,lote_codigo,fecha_fabricacion,fecha_vencimiento,cantidad_inicial,cantidad_disponible,costo_unitario,estado_calidad,proveedor,documento_ref,ubicacion_interna,estado,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (empresa_id,producto_id,bodega_id,lote_codigo) DO UPDATE SET
			fecha_fabricacion=EXCLUDED.fecha_fabricacion,
			fecha_vencimiento=EXCLUDED.fecha_vencimiento,
			cantidad_inicial=empresa_inventario_lotes_avanzados.cantidad_inicial+EXCLUDED.cantidad_inicial,
			cantidad_disponible=empresa_inventario_lotes_avanzados.cantidad_disponible+EXCLUDED.cantidad_disponible,
			costo_unitario=EXCLUDED.costo_unitario,
			estado_calidad=EXCLUDED.estado_calidad,
			proveedor=EXCLUDED.proveedor,
			documento_ref=EXCLUDED.documento_ref,
			ubicacion_interna=EXCLUDED.ubicacion_interna,
			estado=EXCLUDED.estado,
			usuario_creador=EXCLUDED.usuario_creador,
			fecha_actualizacion=CURRENT_TIMESTAMP`,
		lote.EmpresaID, lote.ProductoID, lote.BodegaID, lote.LoteCodigo, lote.FechaFabricacion, lote.FechaVencimiento, lote.CantidadInicial, lote.CantidadDisponible, lote.CostoUnitario, lote.EstadoCalidad, lote.Proveedor, lote.DocumentoRef, lote.UbicacionInterna, lote.Estado, lote.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	if err := upsertExistenciaTx(tx, lote.EmpresaID, lote.ProductoID, lote.BodegaID, lote.CantidadInicial, lote.UsuarioCreador, "entrada lote avanzado "+lote.LoteCodigo); err != nil {
		return 0, err
	}
	if err := insertMovimientoTx(tx, InventarioMovimiento{EmpresaID: lote.EmpresaID, ProductoID: lote.ProductoID, BodegaDestinoID: lote.BodegaID, Tipo: "entrada", Cantidad: lote.CantidadInicial, CostoUnitario: lote.CostoUnitario, Referencia: lote.DocumentoRef, UsuarioCreador: lote.UsuarioCreador, Estado: "activo", Observaciones: "entrada lote " + lote.LoteCodigo}); err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

func CreateEmpresaInventarioSerialAvanzado(dbConn *sql.DB, serial EmpresaInventarioSerialAvanzado) (int64, error) {
	serial = normalizeInventarioSerialAvanzado(serial)
	if serial.EmpresaID <= 0 || serial.ProductoID <= 0 || serial.BodegaID <= 0 || serial.Serial == "" {
		return 0, errors.New("empresa, producto, bodega y serial son requeridos")
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_inventario_seriales_avanzados
		(empresa_id,lote_id,producto_id,bodega_id,serial,estado_operativo,estado_inventario,fecha_ingreso,garantia_hasta,cliente_reserva,observaciones,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (empresa_id,serial) DO UPDATE SET
			lote_id=EXCLUDED.lote_id,
			producto_id=EXCLUDED.producto_id,
			bodega_id=EXCLUDED.bodega_id,
			estado_operativo=EXCLUDED.estado_operativo,
			estado_inventario=EXCLUDED.estado_inventario,
			fecha_ingreso=EXCLUDED.fecha_ingreso,
			garantia_hasta=EXCLUDED.garantia_hasta,
			cliente_reserva=EXCLUDED.cliente_reserva,
			observaciones=EXCLUDED.observaciones,
			usuario_creador=EXCLUDED.usuario_creador,
			fecha_actualizacion=CURRENT_TIMESTAMP`,
		serial.EmpresaID, serial.LoteID, serial.ProductoID, serial.BodegaID, serial.Serial, serial.EstadoOperativo, serial.EstadoInventario, serial.FechaIngreso, serial.GarantiaHasta, serial.ClienteReserva, serial.Observaciones, serial.UsuarioCreador)
}

func CreateEmpresaInventarioReservaAvanzada(dbConn *sql.DB, reserva EmpresaInventarioReservaAvanzada) (int64, error) {
	reserva = normalizeInventarioReservaAvanzada(reserva)
	if reserva.EmpresaID <= 0 || reserva.ProductoID <= 0 || reserva.BodegaID <= 0 || reserva.Cantidad <= 0 {
		return 0, errors.New("empresa, producto, bodega y cantidad son requeridos")
	}
	disponible, err := getInventarioDisponibleAvanzado(dbConn, reserva.EmpresaID, reserva.ProductoID, reserva.BodegaID, reserva.LoteID)
	if err != nil {
		return 0, err
	}
	if disponible < reserva.Cantidad {
		return 0, ErrStockInsuficiente
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_inventario_reservas_avanzadas
		(empresa_id,producto_id,bodega_id,lote_id,serial_id,cantidad,origen_modulo,origen_ref,cliente_nombre,fecha_reserva,fecha_expira,estado,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		reserva.EmpresaID, reserva.ProductoID, reserva.BodegaID, reserva.LoteID, reserva.SerialID, reserva.Cantidad, reserva.OrigenModulo, reserva.OrigenRef, reserva.ClienteNombre, reserva.FechaReserva, reserva.FechaExpira, reserva.Estado, reserva.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	if reserva.SerialID > 0 {
		_, _ = ExecCompat(dbConn, `UPDATE empresa_inventario_seriales_avanzados SET estado_inventario='reservado', cliente_reserva=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, reserva.ClienteNombre, reserva.EmpresaID, reserva.SerialID)
	}
	return id, nil
}

func ConfirmarEmpresaInventarioReservaAvanzada(dbConn *sql.DB, empresaID, reservaID int64, usuario string) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var r EmpresaInventarioReservaAvanzada
	if err := queryRowTxSQLCompat(tx, `SELECT id,empresa_id,producto_id,bodega_id,COALESCE(lote_id,0),COALESCE(serial_id,0),COALESCE(cantidad,0),COALESCE(origen_modulo,''),COALESCE(origen_ref,''),COALESCE(cliente_nombre,''),COALESCE(estado,'activa') FROM empresa_inventario_reservas_avanzadas WHERE empresa_id=? AND id=?`, empresaID, reservaID).Scan(&r.ID, &r.EmpresaID, &r.ProductoID, &r.BodegaID, &r.LoteID, &r.SerialID, &r.Cantidad, &r.OrigenModulo, &r.OrigenRef, &r.ClienteNombre, &r.Estado); err != nil {
		return err
	}
	if r.Estado != "activa" {
		return fmt.Errorf("reserva no activa")
	}
	if r.LoteID > 0 {
		res, err := execTxSQLCompat(tx, `UPDATE empresa_inventario_lotes_avanzados SET cantidad_disponible=cantidad_disponible-?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=? AND cantidad_disponible>=?`, r.Cantidad, empresaID, r.LoteID, r.Cantidad)
		if err != nil {
			return err
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			return ErrStockInsuficiente
		}
	}
	if _, err := execTxSQLCompat(tx, `UPDATE inventario_existencias SET cantidad=cantidad-?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND producto_id=? AND bodega_id=?`, r.Cantidad, empresaID, r.ProductoID, r.BodegaID); err != nil {
		return err
	}
	if _, err := execTxSQLCompat(tx, `UPDATE empresa_inventario_reservas_avanzadas SET estado='confirmada', fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, empresaID, reservaID); err != nil {
		return err
	}
	if r.SerialID > 0 {
		if _, err := execTxSQLCompat(tx, `UPDATE empresa_inventario_seriales_avanzados SET estado_inventario='salido', fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, empresaID, r.SerialID); err != nil {
			return err
		}
	}
	if err := insertMovimientoTx(tx, InventarioMovimiento{EmpresaID: empresaID, ProductoID: r.ProductoID, BodegaOrigenID: r.BodegaID, Tipo: "salida", Cantidad: r.Cantidad, Referencia: r.OrigenRef, UsuarioCreador: strings.TrimSpace(usuario), Estado: "activo", Observaciones: "confirmacion reserva avanzada " + r.OrigenModulo}); err != nil {
		return err
	}
	return tx.Commit()
}

func ListEmpresaInventarioLotesAvanzados(dbConn *sql.DB, empresaID, productoID, bodegaID int64, estado string, limit int) ([]EmpresaInventarioLoteAvanzado, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	args := []interface{}{empresaID}
	where := "l.empresa_id=?"
	if productoID > 0 {
		where += " AND l.producto_id=?"
		args = append(args, productoID)
	}
	if bodegaID > 0 {
		where += " AND l.bodega_id=?"
		args = append(args, bodegaID)
	}
	if strings.TrimSpace(estado) != "" {
		where += " AND l.estado=?"
		args = append(args, normalizeInventarioAvanzadoEstado(estado))
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT l.id,l.empresa_id,l.producto_id,COALESCE(p.nombre,''),l.bodega_id,COALESCE(b.nombre,''),l.lote_codigo,COALESCE(l.fecha_fabricacion,''),COALESCE(l.fecha_vencimiento,''),COALESCE(l.cantidad_inicial,0),COALESCE(l.cantidad_disponible,0),COALESCE(l.costo_unitario,0),COALESCE(l.estado_calidad,'liberado'),COALESCE(l.proveedor,''),COALESCE(l.documento_ref,''),COALESCE(l.ubicacion_interna,''),COALESCE(l.estado,'activo'),COALESCE(l.usuario_creador,''),COALESCE(l.fecha_creacion,''),
		COALESCE((SELECT SUM(r.cantidad) FROM empresa_inventario_reservas_avanzadas r WHERE r.empresa_id=l.empresa_id AND r.lote_id=l.id AND r.estado='activa'),0)
		FROM empresa_inventario_lotes_avanzados l
		JOIN productos p ON p.empresa_id=l.empresa_id AND p.id=l.producto_id
		JOIN bodegas b ON b.empresa_id=l.empresa_id AND b.id=l.bodega_id
		WHERE %s ORDER BY l.id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaInventarioLoteAvanzado{}
	for rows.Next() {
		var x EmpresaInventarioLoteAvanzado
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.ProductoID, &x.ProductoNombre, &x.BodegaID, &x.BodegaNombre, &x.LoteCodigo, &x.FechaFabricacion, &x.FechaVencimiento, &x.CantidadInicial, &x.CantidadDisponible, &x.CostoUnitario, &x.EstadoCalidad, &x.Proveedor, &x.DocumentoRef, &x.UbicacionInterna, &x.Estado, &x.UsuarioCreador, &x.FechaCreacion, &x.CantidadReservada); err != nil {
			return nil, err
		}
		applyInventarioLoteRuntime(&x)
		out = append(out, x)
	}
	return out, rows.Err()
}

func ListEmpresaInventarioSerialesAvanzados(dbConn *sql.DB, empresaID, productoID, bodegaID int64, estado string, limit int) ([]EmpresaInventarioSerialAvanzado, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	args := []interface{}{empresaID}
	where := "s.empresa_id=?"
	if productoID > 0 {
		where += " AND s.producto_id=?"
		args = append(args, productoID)
	}
	if bodegaID > 0 {
		where += " AND s.bodega_id=?"
		args = append(args, bodegaID)
	}
	if strings.TrimSpace(estado) != "" {
		where += " AND s.estado_inventario=?"
		args = append(args, normalizeInventarioSerialEstado(estado))
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT s.id,s.empresa_id,COALESCE(s.lote_id,0),s.producto_id,COALESCE(p.nombre,''),s.bodega_id,COALESCE(b.nombre,''),s.serial,COALESCE(s.estado_operativo,'operativo'),COALESCE(s.estado_inventario,'disponible'),COALESCE(s.fecha_ingreso,''),COALESCE(s.garantia_hasta,''),COALESCE(s.cliente_reserva,''),COALESCE(s.observaciones,''),COALESCE(s.usuario_creador,''),COALESCE(s.fecha_creacion,'')
		FROM empresa_inventario_seriales_avanzados s
		JOIN productos p ON p.empresa_id=s.empresa_id AND p.id=s.producto_id
		JOIN bodegas b ON b.empresa_id=s.empresa_id AND b.id=s.bodega_id
		WHERE %s ORDER BY s.id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaInventarioSerialAvanzado{}
	for rows.Next() {
		var x EmpresaInventarioSerialAvanzado
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.LoteID, &x.ProductoID, &x.ProductoNombre, &x.BodegaID, &x.BodegaNombre, &x.Serial, &x.EstadoOperativo, &x.EstadoInventario, &x.FechaIngreso, &x.GarantiaHasta, &x.ClienteReserva, &x.Observaciones, &x.UsuarioCreador, &x.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func ListEmpresaInventarioReservasAvanzadas(dbConn *sql.DB, empresaID int64, estado string, limit int) ([]EmpresaInventarioReservaAvanzada, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	args := []interface{}{empresaID}
	where := "r.empresa_id=?"
	if strings.TrimSpace(estado) != "" {
		where += " AND r.estado=?"
		args = append(args, normalizeInventarioReservaEstado(estado))
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT r.id,r.empresa_id,r.producto_id,COALESCE(p.nombre,''),r.bodega_id,COALESCE(b.nombre,''),COALESCE(r.lote_id,0),COALESCE(l.lote_codigo,''),COALESCE(r.serial_id,0),COALESCE(s.serial,''),COALESCE(r.cantidad,0),COALESCE(r.origen_modulo,''),COALESCE(r.origen_ref,''),COALESCE(r.cliente_nombre,''),COALESCE(r.fecha_reserva,''),COALESCE(r.fecha_expira,''),COALESCE(r.estado,'activa'),COALESCE(r.usuario_creador,''),COALESCE(r.fecha_creacion,'')
		FROM empresa_inventario_reservas_avanzadas r
		JOIN productos p ON p.empresa_id=r.empresa_id AND p.id=r.producto_id
		JOIN bodegas b ON b.empresa_id=r.empresa_id AND b.id=r.bodega_id
		LEFT JOIN empresa_inventario_lotes_avanzados l ON l.empresa_id=r.empresa_id AND l.id=r.lote_id
		LEFT JOIN empresa_inventario_seriales_avanzados s ON s.empresa_id=r.empresa_id AND s.id=r.serial_id
		WHERE %s ORDER BY r.id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaInventarioReservaAvanzada{}
	for rows.Next() {
		var x EmpresaInventarioReservaAvanzada
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.ProductoID, &x.ProductoNombre, &x.BodegaID, &x.BodegaNombre, &x.LoteID, &x.LoteCodigo, &x.SerialID, &x.Serial, &x.Cantidad, &x.OrigenModulo, &x.OrigenRef, &x.ClienteNombre, &x.FechaReserva, &x.FechaExpira, &x.Estado, &x.UsuarioCreador, &x.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func BuildEmpresaInventarioAvanzadoDashboard(dbConn *sql.DB, empresaID int64) (EmpresaInventarioAvanzadoDashboard, error) {
	var d EmpresaInventarioAvanzadoDashboard
	d.EmpresaID = empresaID
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1), COALESCE(SUM(cantidad_disponible*costo_unitario),0) FROM empresa_inventario_lotes_avanzados WHERE empresa_id=? AND estado='activo'`, empresaID).Scan(&d.LotesActivos, &d.ValorDisponible)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_inventario_seriales_avanzados WHERE empresa_id=? AND estado_inventario IN ('disponible','reservado')`, empresaID).Scan(&d.SerialesActivos)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1), COALESCE(SUM(cantidad),0) FROM empresa_inventario_reservas_avanzadas WHERE empresa_id=? AND estado='activa'`, empresaID).Scan(&d.ReservasActivas, &d.UnidadesReservadas)
	lotes, err := ListEmpresaInventarioLotesAvanzados(dbConn, empresaID, 0, 0, "activo", 8)
	if err != nil {
		return d, err
	}
	for _, lote := range lotes {
		switch lote.EstadoVencimiento {
		case "vencido":
			d.LotesVencidos++
		case "vence_hoy", "proximo_vencer":
			d.LotesPorVencer++
		}
	}
	d.UltimosLotes = lotes
	val, err := ListEmpresaInventarioValorizacionAvanzada(dbConn, empresaID, 20)
	if err != nil {
		return d, err
	}
	d.Valorizacion = val
	return d, nil
}

func ListEmpresaInventarioValorizacionAvanzada(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaInventarioValorizacionAvanzada, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT l.producto_id,COALESCE(p.nombre,''),l.bodega_id,COALESCE(b.nombre,''),COALESCE(SUM(l.cantidad_disponible),0),COALESCE(SUM(l.cantidad_disponible*l.costo_unitario),0),COUNT(l.id),
		COALESCE((SELECT SUM(r.cantidad) FROM empresa_inventario_reservas_avanzadas r WHERE r.empresa_id=l.empresa_id AND r.producto_id=l.producto_id AND r.bodega_id=l.bodega_id AND r.estado='activa'),0)
		FROM empresa_inventario_lotes_avanzados l
		JOIN productos p ON p.empresa_id=l.empresa_id AND p.id=l.producto_id
		JOIN bodegas b ON b.empresa_id=l.empresa_id AND b.id=l.bodega_id
		WHERE l.empresa_id=? AND l.estado='activo'
		GROUP BY l.producto_id,p.nombre,l.bodega_id,b.nombre,l.empresa_id
		ORDER BY COALESCE(SUM(l.cantidad_disponible*l.costo_unitario),0) DESC
		LIMIT %d`, limit), empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaInventarioValorizacionAvanzada{}
	for rows.Next() {
		var x EmpresaInventarioValorizacionAvanzada
		if err := rows.Scan(&x.ProductoID, &x.ProductoNombre, &x.BodegaID, &x.BodegaNombre, &x.CantidadDisponible, &x.ValorDisponible, &x.LotesActivos, &x.CantidadReservada); err != nil {
			return nil, err
		}
		if x.CantidadDisponible > 0 {
			x.CostoPromedio = roundMoney(x.ValorDisponible / x.CantidadDisponible)
		}
		x.CantidadLibre = roundMoney(math.Max(0, x.CantidadDisponible-x.CantidadReservada))
		out = append(out, x)
	}
	return out, rows.Err()
}

func SeedEmpresaInventarioAvanzadoDemo(dbConn *sql.DB, empresaID int64, usuario string) (int64, error) {
	bodegas, err := GetBodegasByEmpresa(dbConn, empresaID, false)
	if err != nil {
		return 0, err
	}
	var bodegaID int64
	if len(bodegas) > 0 {
		bodegaID = bodegas[0].ID
	} else {
		bodegaID, err = CreateBodega(dbConn, Bodega{EmpresaID: empresaID, Codigo: "BOD-INV-AV", Nombre: "Bodega inventario avanzado", UsuarioCreador: usuario, Estado: "activo"})
		if err != nil {
			return 0, err
		}
	}
	productos, err := GetProductosByEmpresa(dbConn, empresaID, "", "activo", 0, 0, 1, 0)
	if err != nil {
		return 0, err
	}
	var productoID int64
	if len(productos) > 0 {
		productoID = productos[0].ID
	} else {
		productoID, err = CreateProducto(dbConn, Producto{EmpresaID: empresaID, BodegaPrincipalID: bodegaID, SKU: "INV-AV-DEMO", Nombre: "Producto inventario avanzado demo", UnidadMedida: "und", Costo: 12000, Precio: 18000, StockMinimo: 5, StockMaximo: 100, UsuarioCreador: usuario, Estado: "activo"}, 0, "demo inventario avanzado")
		if err != nil {
			return 0, err
		}
	}
	code := "LOT-AV-" + time.Now().Format("20060102150405")
	loteID, err := CreateEmpresaInventarioLoteAvanzado(dbConn, EmpresaInventarioLoteAvanzado{EmpresaID: empresaID, ProductoID: productoID, BodegaID: bodegaID, LoteCodigo: code, FechaFabricacion: time.Now().AddDate(0, -1, 0).Format("2006-01-02"), FechaVencimiento: time.Now().AddDate(0, 2, 0).Format("2006-01-02"), CantidadInicial: 25, CostoUnitario: 12500, EstadoCalidad: "liberado", Proveedor: "Proveedor demo inventario", DocumentoRef: "QA-" + code, UbicacionInterna: "Rack A1", UsuarioCreador: usuario})
	if err != nil {
		return 0, err
	}
	serialID, err := CreateEmpresaInventarioSerialAvanzado(dbConn, EmpresaInventarioSerialAvanzado{EmpresaID: empresaID, LoteID: loteID, ProductoID: productoID, BodegaID: bodegaID, Serial: "SER-" + code, EstadoOperativo: "operativo", EstadoInventario: "disponible", FechaIngreso: time.Now().Format("2006-01-02"), GarantiaHasta: time.Now().AddDate(1, 0, 0).Format("2006-01-02"), UsuarioCreador: usuario})
	if err != nil {
		return 0, err
	}
	if _, err := CreateEmpresaInventarioReservaAvanzada(dbConn, EmpresaInventarioReservaAvanzada{EmpresaID: empresaID, ProductoID: productoID, BodegaID: bodegaID, LoteID: loteID, SerialID: serialID, Cantidad: 1, OrigenModulo: "qa_demo", OrigenRef: "RSV-" + code, ClienteNombre: "Cliente inventario avanzado", FechaReserva: time.Now().Format("2006-01-02"), FechaExpira: time.Now().AddDate(0, 0, 2).Format("2006-01-02"), UsuarioCreador: usuario}); err != nil {
		return 0, err
	}
	return loteID, nil
}

func getInventarioDisponibleAvanzado(dbConn *sql.DB, empresaID, productoID, bodegaID, loteID int64) (float64, error) {
	args := []interface{}{empresaID, productoID, bodegaID}
	where := "empresa_id=? AND producto_id=? AND bodega_id=? AND estado='activo'"
	if loteID > 0 {
		where += " AND id=?"
		args = append(args, loteID)
	}
	var disponible, reservado float64
	if err := QueryRowCompat(dbConn, "SELECT COALESCE(SUM(cantidad_disponible),0) FROM empresa_inventario_lotes_avanzados WHERE "+where, args...).Scan(&disponible); err != nil {
		return 0, err
	}
	resArgs := []interface{}{empresaID, productoID, bodegaID}
	resWhere := "empresa_id=? AND producto_id=? AND bodega_id=? AND estado='activa'"
	if loteID > 0 {
		resWhere += " AND lote_id=?"
		resArgs = append(resArgs, loteID)
	}
	if err := QueryRowCompat(dbConn, "SELECT COALESCE(SUM(cantidad),0) FROM empresa_inventario_reservas_avanzadas WHERE "+resWhere, resArgs...).Scan(&reservado); err != nil {
		return 0, err
	}
	return math.Max(0, disponible-reservado), nil
}

func normalizeInventarioLoteAvanzado(x EmpresaInventarioLoteAvanzado) EmpresaInventarioLoteAvanzado {
	x.LoteCodigo = strings.ToUpper(strings.TrimSpace(x.LoteCodigo))
	x.FechaFabricacion = strings.TrimSpace(x.FechaFabricacion)
	x.FechaVencimiento = strings.TrimSpace(x.FechaVencimiento)
	x.CantidadInicial = roundMoney(math.Max(0, x.CantidadInicial))
	if x.CantidadDisponible <= 0 {
		x.CantidadDisponible = x.CantidadInicial
	}
	x.CantidadDisponible = roundMoney(math.Max(0, x.CantidadDisponible))
	x.CostoUnitario = roundMoney(math.Max(0, x.CostoUnitario))
	x.EstadoCalidad = normalizeInventarioCalidad(x.EstadoCalidad)
	x.Proveedor = strings.TrimSpace(x.Proveedor)
	x.DocumentoRef = strings.TrimSpace(x.DocumentoRef)
	x.UbicacionInterna = strings.TrimSpace(x.UbicacionInterna)
	x.Estado = normalizeInventarioAvanzadoEstado(x.Estado)
	x.UsuarioCreador = strings.TrimSpace(x.UsuarioCreador)
	return x
}

func normalizeInventarioSerialAvanzado(x EmpresaInventarioSerialAvanzado) EmpresaInventarioSerialAvanzado {
	x.Serial = strings.ToUpper(strings.TrimSpace(x.Serial))
	x.EstadoOperativo = strings.ToLower(strings.TrimSpace(x.EstadoOperativo))
	if x.EstadoOperativo == "" {
		x.EstadoOperativo = "operativo"
	}
	x.EstadoInventario = normalizeInventarioSerialEstado(x.EstadoInventario)
	x.FechaIngreso = strings.TrimSpace(x.FechaIngreso)
	if x.FechaIngreso == "" {
		x.FechaIngreso = time.Now().Format("2006-01-02")
	}
	x.GarantiaHasta = strings.TrimSpace(x.GarantiaHasta)
	x.ClienteReserva = strings.TrimSpace(x.ClienteReserva)
	x.Observaciones = strings.TrimSpace(x.Observaciones)
	x.UsuarioCreador = strings.TrimSpace(x.UsuarioCreador)
	return x
}

func normalizeInventarioReservaAvanzada(x EmpresaInventarioReservaAvanzada) EmpresaInventarioReservaAvanzada {
	x.Cantidad = roundMoney(math.Max(0, x.Cantidad))
	x.OrigenModulo = strings.ToLower(strings.TrimSpace(x.OrigenModulo))
	if x.OrigenModulo == "" {
		x.OrigenModulo = "manual"
	}
	x.OrigenRef = strings.TrimSpace(x.OrigenRef)
	x.ClienteNombre = strings.TrimSpace(x.ClienteNombre)
	x.FechaReserva = strings.TrimSpace(x.FechaReserva)
	if x.FechaReserva == "" {
		x.FechaReserva = time.Now().Format("2006-01-02")
	}
	x.FechaExpira = strings.TrimSpace(x.FechaExpira)
	x.Estado = normalizeInventarioReservaEstado(x.Estado)
	x.UsuarioCreador = strings.TrimSpace(x.UsuarioCreador)
	return x
}

func applyInventarioLoteRuntime(x *EmpresaInventarioLoteAvanzado) {
	x.CantidadLibre = roundMoney(math.Max(0, x.CantidadDisponible-x.CantidadReservada))
	x.ValorDisponible = roundMoney(x.CantidadDisponible * x.CostoUnitario)
	x.EstadoVencimiento, x.DiasParaVencer = productoVencimientoStatus(x.FechaVencimiento, 30, time.Now())
	if strings.TrimSpace(x.FechaVencimiento) == "" {
		x.EstadoVencimiento = "no_aplica"
		x.DiasParaVencer = 0
	}
}

func normalizeInventarioCalidad(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "cuarentena", "liberado", "rechazado", "bloqueado":
		return v
	default:
		return "liberado"
	}
}

func normalizeInventarioAvanzadoEstado(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "activo", "inactivo", "agotado", "bloqueado":
		return v
	default:
		return "activo"
	}
}

func normalizeInventarioSerialEstado(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "disponible", "reservado", "salido", "mantenimiento", "bloqueado":
		return v
	default:
		return "disponible"
	}
}

func normalizeInventarioReservaEstado(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "activa", "confirmada", "liberada", "vencida", "cancelada":
		return v
	default:
		return "activa"
	}
}
