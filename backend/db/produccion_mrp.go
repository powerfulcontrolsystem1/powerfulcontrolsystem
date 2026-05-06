package db

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
)

type EmpresaProduccionMRPConfig struct {
	EmpresaID                   int64  `json:"empresa_id"`
	NombreSistema               string `json:"nombre_sistema"`
	Moneda                      string `json:"moneda"`
	CosteoModo                  string `json:"costeo_modo"`
	AprobarOrdenes              bool   `json:"aprobar_ordenes"`
	ConsumirInventarioAlIniciar bool   `json:"consumir_inventario_al_iniciar"`
	CerrarConCalidad            bool   `json:"cerrar_con_calidad"`
	FechaActualizacion          string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador              string `json:"usuario_creador,omitempty"`
}

type EmpresaProduccionReceta struct {
	ID                      int64                         `json:"id"`
	EmpresaID               int64                         `json:"empresa_id"`
	Codigo                  string                        `json:"codigo"`
	Nombre                  string                        `json:"nombre"`
	ProductoTerminadoID     int64                         `json:"producto_terminado_id,omitempty"`
	ProductoTerminadoNombre string                        `json:"producto_terminado_nombre"`
	Version                 string                        `json:"version"`
	Unidad                  string                        `json:"unidad"`
	CantidadBase            float64                       `json:"cantidad_base"`
	CostoEstandar           float64                       `json:"costo_estandar"`
	MermaPorcentaje         float64                       `json:"merma_porcentaje"`
	TiempoEstimadoMin       int                           `json:"tiempo_estimado_min"`
	Estado                  string                        `json:"estado"`
	FechaCreacion           string                        `json:"fecha_creacion,omitempty"`
	FechaActualizacion      string                        `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador          string                        `json:"usuario_creador,omitempty"`
	Componentes             []EmpresaProduccionComponente `json:"componentes,omitempty"`
}

type EmpresaProduccionComponente struct {
	ID              int64   `json:"id"`
	EmpresaID       int64   `json:"empresa_id"`
	RecetaID        int64   `json:"receta_id"`
	ProductoID      int64   `json:"producto_id,omitempty"`
	ProductoNombre  string  `json:"producto_nombre"`
	Unidad          string  `json:"unidad"`
	Cantidad        float64 `json:"cantidad"`
	CostoUnitario   float64 `json:"costo_unitario"`
	MermaPorcentaje float64 `json:"merma_porcentaje"`
	Obligatoria     bool    `json:"obligatoria"`
	Etapa           string  `json:"etapa"`
	Orden           int     `json:"orden"`
}

type EmpresaProduccionOrden struct {
	ID                      int64   `json:"id"`
	EmpresaID               int64   `json:"empresa_id"`
	Codigo                  string  `json:"codigo"`
	RecetaID                int64   `json:"receta_id"`
	RecetaNombre            string  `json:"receta_nombre,omitempty"`
	ProductoTerminadoID     int64   `json:"producto_terminado_id,omitempty"`
	ProductoTerminadoNombre string  `json:"producto_terminado_nombre"`
	CantidadPlanificada     float64 `json:"cantidad_planificada"`
	CantidadProducida       float64 `json:"cantidad_producida"`
	Estado                  string  `json:"estado"`
	Prioridad               string  `json:"prioridad"`
	FechaProgramada         string  `json:"fecha_programada,omitempty"`
	FechaInicio             string  `json:"fecha_inicio,omitempty"`
	FechaCierre             string  `json:"fecha_cierre,omitempty"`
	CostoEstimado           float64 `json:"costo_estimado"`
	CostoReal               float64 `json:"costo_real"`
	Responsable             string  `json:"responsable,omitempty"`
	Observaciones           string  `json:"observaciones,omitempty"`
	FechaCreacion           string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion      string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador          string  `json:"usuario_creador,omitempty"`
}

type EmpresaProduccionConsumo struct {
	ID                  int64   `json:"id"`
	EmpresaID           int64   `json:"empresa_id"`
	OrdenID             int64   `json:"orden_id"`
	ProductoID          int64   `json:"producto_id,omitempty"`
	ProductoNombre      string  `json:"producto_nombre"`
	CantidadPlanificada float64 `json:"cantidad_planificada"`
	CantidadConsumida   float64 `json:"cantidad_consumida"`
	CostoUnitario       float64 `json:"costo_unitario"`
	CostoTotal          float64 `json:"costo_total"`
	LoteCodigo          string  `json:"lote_codigo,omitempty"`
	Merma               float64 `json:"merma"`
	FechaConsumo        string  `json:"fecha_consumo,omitempty"`
	UsuarioCreador      string  `json:"usuario_creador,omitempty"`
}

type EmpresaProduccionCalidad struct {
	ID                int64   `json:"id"`
	EmpresaID         int64   `json:"empresa_id"`
	OrdenID           int64   `json:"orden_id"`
	Resultado         string  `json:"resultado"`
	ChecklistJSON     string  `json:"checklist_json,omitempty"`
	CantidadAprobada  float64 `json:"cantidad_aprobada"`
	CantidadRechazada float64 `json:"cantidad_rechazada"`
	Responsable       string  `json:"responsable,omitempty"`
	Observaciones     string  `json:"observaciones,omitempty"`
	FechaRevision     string  `json:"fecha_revision,omitempty"`
}

type EmpresaProduccionMRPPlan struct {
	ID                       int64   `json:"id"`
	EmpresaID                int64   `json:"empresa_id"`
	ProductoID               int64   `json:"producto_id,omitempty"`
	ProductoNombre           string  `json:"producto_nombre"`
	Periodo                  string  `json:"periodo"`
	DemandaEstimada          float64 `json:"demanda_estimada"`
	StockActual              float64 `json:"stock_actual"`
	StockSeguridad           float64 `json:"stock_seguridad"`
	RequeridoBruto           float64 `json:"requerido_bruto"`
	DisponibleProyectado     float64 `json:"disponible_proyectado"`
	CantidadSugeridaCompra   float64 `json:"cantidad_sugerida_compra"`
	CantidadSugeridaProducir float64 `json:"cantidad_sugerida_producir"`
	Origen                   string  `json:"origen"`
	Estado                   string  `json:"estado"`
	FechaCreacion            string  `json:"fecha_creacion,omitempty"`
	UsuarioCreador           string  `json:"usuario_creador,omitempty"`
}

type EmpresaProduccionMRPDashboard struct {
	EmpresaID            int64                      `json:"empresa_id"`
	RecetasActivas       int                        `json:"recetas_activas"`
	OrdenesAbiertas      int                        `json:"ordenes_abiertas"`
	OrdenesCalidad       int                        `json:"ordenes_calidad"`
	OrdenesCerradas      int                        `json:"ordenes_cerradas"`
	CostoEstimadoAbierto float64                    `json:"costo_estimado_abierto"`
	CostoRealMes         float64                    `json:"costo_real_mes"`
	Config               EmpresaProduccionMRPConfig `json:"config"`
	Recetas              []EmpresaProduccionReceta  `json:"recetas"`
	Ordenes              []EmpresaProduccionOrden   `json:"ordenes"`
	Plan                 []EmpresaProduccionMRPPlan `json:"plan"`
	ConsumosRecientes    []EmpresaProduccionConsumo `json:"consumos_recientes"`
	RevisionesCalidad    []EmpresaProduccionCalidad `json:"revisiones_calidad"`
}

func EnsureEmpresaProduccionMRPSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_produccion_mrp_config (
			empresa_id BIGINT PRIMARY KEY,
			nombre_sistema TEXT DEFAULT 'Produccion / MRP',
			moneda TEXT DEFAULT 'COP',
			costeo_modo TEXT DEFAULT 'estandar',
			aprobar_ordenes INTEGER DEFAULT 1,
			consumir_inventario_al_iniciar INTEGER DEFAULT 0,
			cerrar_con_calidad INTEGER DEFAULT 1,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_produccion_recetas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			producto_terminado_id BIGINT DEFAULT 0,
			producto_terminado_nombre TEXT,
			version TEXT DEFAULT '1.0',
			unidad TEXT DEFAULT 'und',
			cantidad_base NUMERIC(14,4) DEFAULT 1,
			costo_estandar NUMERIC(14,2) DEFAULT 0,
			merma_porcentaje NUMERIC(7,2) DEFAULT 0,
			tiempo_estimado_min INTEGER DEFAULT 0,
			estado TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_prod_receta_empresa_codigo ON empresa_produccion_recetas(empresa_id, codigo)`,
		`CREATE INDEX IF NOT EXISTS ix_prod_receta_empresa_estado ON empresa_produccion_recetas(empresa_id, estado)`,
		`CREATE TABLE IF NOT EXISTS empresa_produccion_receta_componentes (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			receta_id BIGINT NOT NULL,
			producto_id BIGINT DEFAULT 0,
			producto_nombre TEXT NOT NULL,
			unidad TEXT DEFAULT 'und',
			cantidad NUMERIC(14,4) DEFAULT 0,
			costo_unitario NUMERIC(14,2) DEFAULT 0,
			merma_porcentaje NUMERIC(7,2) DEFAULT 0,
			obligatoria INTEGER DEFAULT 1,
			etapa TEXT DEFAULT 'preparacion',
			orden INTEGER DEFAULT 0
		)`,
		`CREATE INDEX IF NOT EXISTS ix_prod_comp_empresa_receta ON empresa_produccion_receta_componentes(empresa_id, receta_id, orden, id)`,
		`CREATE TABLE IF NOT EXISTS empresa_produccion_ordenes (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			codigo TEXT NOT NULL,
			receta_id BIGINT NOT NULL,
			producto_terminado_id BIGINT DEFAULT 0,
			producto_terminado_nombre TEXT,
			cantidad_planificada NUMERIC(14,4) DEFAULT 0,
			cantidad_producida NUMERIC(14,4) DEFAULT 0,
			estado TEXT DEFAULT 'borrador',
			prioridad TEXT DEFAULT 'normal',
			fecha_programada TEXT,
			fecha_inicio TEXT,
			fecha_cierre TEXT,
			costo_estimado NUMERIC(14,2) DEFAULT 0,
			costo_real NUMERIC(14,2) DEFAULT 0,
			responsable TEXT,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_prod_orden_empresa_codigo ON empresa_produccion_ordenes(empresa_id, codigo)`,
		`CREATE INDEX IF NOT EXISTS ix_prod_orden_empresa_estado ON empresa_produccion_ordenes(empresa_id, estado, id DESC)`,
		`CREATE TABLE IF NOT EXISTS empresa_produccion_consumos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			orden_id BIGINT NOT NULL,
			producto_id BIGINT DEFAULT 0,
			producto_nombre TEXT NOT NULL,
			cantidad_planificada NUMERIC(14,4) DEFAULT 0,
			cantidad_consumida NUMERIC(14,4) DEFAULT 0,
			costo_unitario NUMERIC(14,2) DEFAULT 0,
			costo_total NUMERIC(14,2) DEFAULT 0,
			lote_codigo TEXT,
			merma NUMERIC(14,4) DEFAULT 0,
			fecha_consumo TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_prod_consumo_empresa_orden ON empresa_produccion_consumos(empresa_id, orden_id, id DESC)`,
		`CREATE TABLE IF NOT EXISTS empresa_produccion_calidad (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			orden_id BIGINT NOT NULL,
			resultado TEXT DEFAULT 'pendiente',
			checklist_json TEXT,
			cantidad_aprobada NUMERIC(14,4) DEFAULT 0,
			cantidad_rechazada NUMERIC(14,4) DEFAULT 0,
			responsable TEXT,
			observaciones TEXT,
			fecha_revision TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS ix_prod_calidad_empresa_orden ON empresa_produccion_calidad(empresa_id, orden_id, id DESC)`,
		`CREATE TABLE IF NOT EXISTS empresa_produccion_mrp_plan (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			producto_id BIGINT DEFAULT 0,
			producto_nombre TEXT NOT NULL,
			periodo TEXT NOT NULL,
			demanda_estimada NUMERIC(14,4) DEFAULT 0,
			stock_actual NUMERIC(14,4) DEFAULT 0,
			stock_seguridad NUMERIC(14,4) DEFAULT 0,
			requerido_bruto NUMERIC(14,4) DEFAULT 0,
			disponible_proyectado NUMERIC(14,4) DEFAULT 0,
			cantidad_sugerida_compra NUMERIC(14,4) DEFAULT 0,
			cantidad_sugerida_producir NUMERIC(14,4) DEFAULT 0,
			origen TEXT DEFAULT 'manual',
			estado TEXT DEFAULT 'borrador',
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_prod_mrp_plan_empresa_periodo ON empresa_produccion_mrp_plan(empresa_id, periodo, estado)`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func defaultProduccionMRPConfig(empresaID int64) EmpresaProduccionMRPConfig {
	return EmpresaProduccionMRPConfig{
		EmpresaID:                   empresaID,
		NombreSistema:               "Produccion / MRP",
		Moneda:                      "COP",
		CosteoModo:                  "estandar",
		AprobarOrdenes:              true,
		ConsumirInventarioAlIniciar: false,
		CerrarConCalidad:            true,
	}
}

func GetEmpresaProduccionMRPConfig(dbConn *sql.DB, empresaID int64) (EmpresaProduccionMRPConfig, error) {
	if err := EnsureEmpresaProduccionMRPSchema(dbConn); err != nil {
		return EmpresaProduccionMRPConfig{}, err
	}
	cfg := defaultProduccionMRPConfig(empresaID)
	var aprobar, consumir, calidad int
	err := QueryRowCompat(dbConn, `SELECT empresa_id, COALESCE(nombre_sistema,''), COALESCE(moneda,'COP'), COALESCE(costeo_modo,'estandar'), COALESCE(aprobar_ordenes,1), COALESCE(consumir_inventario_al_iniciar,0), COALESCE(cerrar_con_calidad,1), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'') FROM empresa_produccion_mrp_config WHERE empresa_id=?`, empresaID).Scan(&cfg.EmpresaID, &cfg.NombreSistema, &cfg.Moneda, &cfg.CosteoModo, &aprobar, &consumir, &calidad, &cfg.FechaActualizacion, &cfg.UsuarioCreador)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return cfg, nil
		}
		return EmpresaProduccionMRPConfig{}, err
	}
	cfg.AprobarOrdenes = aprobar > 0
	cfg.ConsumirInventarioAlIniciar = consumir > 0
	cfg.CerrarConCalidad = calidad > 0
	return normalizeProduccionMRPConfig(cfg), nil
}

func UpsertEmpresaProduccionMRPConfig(dbConn *sql.DB, cfg EmpresaProduccionMRPConfig) error {
	if err := EnsureEmpresaProduccionMRPSchema(dbConn); err != nil {
		return err
	}
	cfg = normalizeProduccionMRPConfig(cfg)
	_, err := ExecCompat(dbConn, `INSERT INTO empresa_produccion_mrp_config (
		empresa_id,nombre_sistema,moneda,costeo_modo,aprobar_ordenes,consumir_inventario_al_iniciar,cerrar_con_calidad,fecha_actualizacion,usuario_creador
	) VALUES (?,?,?,?,?,?,?,CURRENT_TIMESTAMP,?)
	ON CONFLICT (empresa_id) DO UPDATE SET
		nombre_sistema=EXCLUDED.nombre_sistema,
		moneda=EXCLUDED.moneda,
		costeo_modo=EXCLUDED.costeo_modo,
		aprobar_ordenes=EXCLUDED.aprobar_ordenes,
		consumir_inventario_al_iniciar=EXCLUDED.consumir_inventario_al_iniciar,
		cerrar_con_calidad=EXCLUDED.cerrar_con_calidad,
		fecha_actualizacion=CURRENT_TIMESTAMP,
		usuario_creador=EXCLUDED.usuario_creador`,
		cfg.EmpresaID, cfg.NombreSistema, cfg.Moneda, cfg.CosteoModo, boolIntProduccion(cfg.AprobarOrdenes), boolIntProduccion(cfg.ConsumirInventarioAlIniciar), boolIntProduccion(cfg.CerrarConCalidad), cfg.UsuarioCreador)
	return err
}

func UpsertEmpresaProduccionReceta(dbConn *sql.DB, item EmpresaProduccionReceta) (int64, error) {
	if err := EnsureEmpresaProduccionMRPSchema(dbConn); err != nil {
		return 0, err
	}
	item = normalizeProduccionReceta(item)
	if item.EmpresaID <= 0 {
		return 0, errors.New("empresa_id es obligatorio")
	}
	if item.Codigo == "" || item.Nombre == "" {
		return 0, errors.New("codigo y nombre de receta son obligatorios")
	}
	if item.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_produccion_recetas SET codigo=?, nombre=?, producto_terminado_id=?, producto_terminado_nombre=?, version=?, unidad=?, cantidad_base=?, costo_estandar=?, merma_porcentaje=?, tiempo_estimado_min=?, estado=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`,
			item.Codigo, item.Nombre, item.ProductoTerminadoID, item.ProductoTerminadoNombre, item.Version, item.Unidad, item.CantidadBase, item.CostoEstandar, item.MermaPorcentaje, item.TiempoEstimadoMin, item.Estado, item.EmpresaID, item.ID)
		if err != nil {
			return 0, err
		}
		if item.Componentes != nil {
			if err := ReplaceEmpresaProduccionComponentes(dbConn, item.EmpresaID, item.ID, item.Componentes); err != nil {
				return 0, err
			}
		}
		return item.ID, nil
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_produccion_recetas (empresa_id,codigo,nombre,producto_terminado_id,producto_terminado_nombre,version,unidad,cantidad_base,costo_estandar,merma_porcentaje,tiempo_estimado_min,estado,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		item.EmpresaID, item.Codigo, item.Nombre, item.ProductoTerminadoID, item.ProductoTerminadoNombre, item.Version, item.Unidad, item.CantidadBase, item.CostoEstandar, item.MermaPorcentaje, item.TiempoEstimadoMin, item.Estado, item.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	if item.Componentes != nil {
		if err := ReplaceEmpresaProduccionComponentes(dbConn, item.EmpresaID, id, item.Componentes); err != nil {
			return 0, err
		}
	}
	return id, nil
}

func ReplaceEmpresaProduccionComponentes(dbConn *sql.DB, empresaID, recetaID int64, rows []EmpresaProduccionComponente) error {
	if empresaID <= 0 || recetaID <= 0 {
		return errors.New("empresa_id y receta_id son obligatorios")
	}
	if _, err := ExecCompat(dbConn, `DELETE FROM empresa_produccion_receta_componentes WHERE empresa_id=? AND receta_id=?`, empresaID, recetaID); err != nil {
		return err
	}
	for i, row := range rows {
		row = normalizeProduccionComponente(row)
		if row.ProductoNombre == "" || row.Cantidad <= 0 {
			continue
		}
		if row.Orden <= 0 {
			row.Orden = i + 1
		}
		_, err := ExecCompat(dbConn, `INSERT INTO empresa_produccion_receta_componentes (empresa_id,receta_id,producto_id,producto_nombre,unidad,cantidad,costo_unitario,merma_porcentaje,obligatoria,etapa,orden) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
			empresaID, recetaID, row.ProductoID, row.ProductoNombre, row.Unidad, row.Cantidad, row.CostoUnitario, row.MermaPorcentaje, boolIntProduccion(row.Obligatoria), row.Etapa, row.Orden)
		if err != nil {
			return err
		}
	}
	return nil
}

func ListEmpresaProduccionRecetas(dbConn *sql.DB, empresaID int64, estado string, limit int) ([]EmpresaProduccionReceta, error) {
	if err := EnsureEmpresaProduccionMRPSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 150
	}
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if strings.TrimSpace(estado) != "" {
		where += " AND LOWER(COALESCE(estado,''))=?"
		args = append(args, normalizeProduccionEstadoReceta(estado))
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,COALESCE(codigo,''),COALESCE(nombre,''),COALESCE(producto_terminado_id,0),COALESCE(producto_terminado_nombre,''),COALESCE(version,'1.0'),COALESCE(unidad,'und'),COALESCE(cantidad_base,1),COALESCE(costo_estandar,0),COALESCE(merma_porcentaje,0),COALESCE(tiempo_estimado_min,0),COALESCE(estado,'activo'),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,''),COALESCE(usuario_creador,'') FROM empresa_produccion_recetas WHERE %s ORDER BY id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaProduccionReceta{}
	for rows.Next() {
		var x EmpresaProduccionReceta
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Codigo, &x.Nombre, &x.ProductoTerminadoID, &x.ProductoTerminadoNombre, &x.Version, &x.Unidad, &x.CantidadBase, &x.CostoEstandar, &x.MermaPorcentaje, &x.TiempoEstimadoMin, &x.Estado, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		x.Componentes, _ = ListEmpresaProduccionComponentes(dbConn, empresaID, x.ID)
		out = append(out, x)
	}
	return out, rows.Err()
}

func ListEmpresaProduccionComponentes(dbConn *sql.DB, empresaID, recetaID int64) ([]EmpresaProduccionComponente, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,receta_id,COALESCE(producto_id,0),COALESCE(producto_nombre,''),COALESCE(unidad,'und'),COALESCE(cantidad,0),COALESCE(costo_unitario,0),COALESCE(merma_porcentaje,0),COALESCE(obligatoria,1),COALESCE(etapa,'preparacion'),COALESCE(orden,0) FROM empresa_produccion_receta_componentes WHERE empresa_id=? AND receta_id=? ORDER BY orden,id`, empresaID, recetaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaProduccionComponente{}
	for rows.Next() {
		var x EmpresaProduccionComponente
		var obligatoria int
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.RecetaID, &x.ProductoID, &x.ProductoNombre, &x.Unidad, &x.Cantidad, &x.CostoUnitario, &x.MermaPorcentaje, &obligatoria, &x.Etapa, &x.Orden); err != nil {
			return nil, err
		}
		x.Obligatoria = obligatoria > 0
		out = append(out, x)
	}
	return out, rows.Err()
}

func CreateEmpresaProduccionOrden(dbConn *sql.DB, item EmpresaProduccionOrden) (EmpresaProduccionOrden, error) {
	if err := EnsureEmpresaProduccionMRPSchema(dbConn); err != nil {
		return EmpresaProduccionOrden{}, err
	}
	item = normalizeProduccionOrden(item)
	if item.EmpresaID <= 0 || item.RecetaID <= 0 {
		return EmpresaProduccionOrden{}, errors.New("empresa_id y receta_id son obligatorios")
	}
	if item.Codigo == "" {
		code, err := nextProduccionOrdenCode(dbConn, item.EmpresaID)
		if err != nil {
			return EmpresaProduccionOrden{}, err
		}
		item.Codigo = code
	}
	if item.ProductoTerminadoNombre == "" || item.CostoEstimado <= 0 {
		rec, err := getEmpresaProduccionRecetaByID(dbConn, item.EmpresaID, item.RecetaID)
		if err == nil {
			if item.ProductoTerminadoNombre == "" {
				item.ProductoTerminadoNombre = rec.ProductoTerminadoNombre
			}
			item.ProductoTerminadoID = rec.ProductoTerminadoID
			if item.CostoEstimado <= 0 {
				item.CostoEstimado = roundMoneyProduccion(rec.CostoEstandar * item.CantidadPlanificada / rec.CantidadBase)
			}
		}
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_produccion_ordenes (empresa_id,codigo,receta_id,producto_terminado_id,producto_terminado_nombre,cantidad_planificada,cantidad_producida,estado,prioridad,fecha_programada,costo_estimado,costo_real,responsable,observaciones,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		item.EmpresaID, item.Codigo, item.RecetaID, item.ProductoTerminadoID, item.ProductoTerminadoNombre, item.CantidadPlanificada, item.CantidadProducida, item.Estado, item.Prioridad, item.FechaProgramada, item.CostoEstimado, item.CostoReal, item.Responsable, item.Observaciones, item.UsuarioCreador)
	if err != nil {
		return EmpresaProduccionOrden{}, err
	}
	if err := materializeProduccionConsumosFromReceta(dbConn, item.EmpresaID, id, item.RecetaID, item.CantidadPlanificada, item.UsuarioCreador); err != nil {
		return EmpresaProduccionOrden{}, err
	}
	return GetEmpresaProduccionOrdenByID(dbConn, item.EmpresaID, id)
}

func UpdateEmpresaProduccionOrden(dbConn *sql.DB, item EmpresaProduccionOrden) error {
	if err := EnsureEmpresaProduccionMRPSchema(dbConn); err != nil {
		return err
	}
	item = normalizeProduccionOrden(item)
	if item.ID <= 0 || item.EmpresaID <= 0 {
		return errors.New("id y empresa_id son obligatorios")
	}
	_, err := ExecCompat(dbConn, `UPDATE empresa_produccion_ordenes SET producto_terminado_nombre=?, cantidad_planificada=?, cantidad_producida=?, prioridad=?, fecha_programada=?, costo_estimado=?, costo_real=?, responsable=?, observaciones=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`,
		item.ProductoTerminadoNombre, item.CantidadPlanificada, item.CantidadProducida, item.Prioridad, item.FechaProgramada, item.CostoEstimado, item.CostoReal, item.Responsable, item.Observaciones, item.EmpresaID, item.ID)
	return err
}

func GetEmpresaProduccionOrdenByID(dbConn *sql.DB, empresaID, id int64) (EmpresaProduccionOrden, error) {
	var x EmpresaProduccionOrden
	err := QueryRowCompat(dbConn, produccionOrdenSelect()+` WHERE o.empresa_id=? AND o.id=?`, empresaID, id).Scan(&x.ID, &x.EmpresaID, &x.Codigo, &x.RecetaID, &x.RecetaNombre, &x.ProductoTerminadoID, &x.ProductoTerminadoNombre, &x.CantidadPlanificada, &x.CantidadProducida, &x.Estado, &x.Prioridad, &x.FechaProgramada, &x.FechaInicio, &x.FechaCierre, &x.CostoEstimado, &x.CostoReal, &x.Responsable, &x.Observaciones, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador)
	return x, err
}

func ListEmpresaProduccionOrdenes(dbConn *sql.DB, empresaID int64, estado string, limit int) ([]EmpresaProduccionOrden, error) {
	if err := EnsureEmpresaProduccionMRPSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 150
	}
	args := []interface{}{empresaID}
	where := "o.empresa_id=?"
	if strings.TrimSpace(estado) != "" {
		where += " AND LOWER(COALESCE(o.estado,''))=?"
		args = append(args, normalizeProduccionEstadoOrden(estado))
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`%s WHERE %s ORDER BY o.id DESC LIMIT %d`, produccionOrdenSelect(), where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaProduccionOrden{}
	for rows.Next() {
		var x EmpresaProduccionOrden
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Codigo, &x.RecetaID, &x.RecetaNombre, &x.ProductoTerminadoID, &x.ProductoTerminadoNombre, &x.CantidadPlanificada, &x.CantidadProducida, &x.Estado, &x.Prioridad, &x.FechaProgramada, &x.FechaInicio, &x.FechaCierre, &x.CostoEstimado, &x.CostoReal, &x.Responsable, &x.Observaciones, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func CambiarEstadoEmpresaProduccionOrden(dbConn *sql.DB, empresaID, ordenID int64, estado, usuario string) (EmpresaProduccionOrden, error) {
	estado = normalizeProduccionEstadoOrden(estado)
	if ordenID <= 0 || estado == "" {
		return EmpresaProduccionOrden{}, errors.New("orden_id y estado son obligatorios")
	}
	extra := ""
	args := []interface{}{estado}
	switch estado {
	case "en_proceso":
		extra = ", fecha_inicio=COALESCE(fecha_inicio,CAST(CURRENT_TIMESTAMP AS TEXT))"
	case "cerrada", "cancelada":
		extra = ", fecha_cierre=CAST(CURRENT_TIMESTAMP AS TEXT)"
	case "calidad":
		extra = ", fecha_inicio=COALESCE(fecha_inicio,CAST(CURRENT_TIMESTAMP AS TEXT))"
	}
	args = append(args, ordenID, empresaID)
	_, err := ExecCompat(dbConn, `UPDATE empresa_produccion_ordenes SET estado=?`+extra+`, fecha_actualizacion=CURRENT_TIMESTAMP WHERE id=? AND empresa_id=?`, args...)
	if err != nil {
		return EmpresaProduccionOrden{}, err
	}
	return GetEmpresaProduccionOrdenByID(dbConn, empresaID, ordenID)
}

func RegistrarEmpresaProduccionConsumo(dbConn *sql.DB, item EmpresaProduccionConsumo) (int64, error) {
	if err := EnsureEmpresaProduccionMRPSchema(dbConn); err != nil {
		return 0, err
	}
	item = normalizeProduccionConsumo(item)
	if item.EmpresaID <= 0 || item.OrdenID <= 0 || item.ProductoNombre == "" {
		return 0, errors.New("orden y producto son obligatorios")
	}
	item.CostoTotal = roundMoneyProduccion(item.CantidadConsumida * item.CostoUnitario)
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_produccion_consumos (empresa_id,orden_id,producto_id,producto_nombre,cantidad_planificada,cantidad_consumida,costo_unitario,costo_total,lote_codigo,merma,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		item.EmpresaID, item.OrdenID, item.ProductoID, item.ProductoNombre, item.CantidadPlanificada, item.CantidadConsumida, item.CostoUnitario, item.CostoTotal, item.LoteCodigo, item.Merma, item.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	_, _ = ExecCompat(dbConn, `UPDATE empresa_produccion_ordenes SET costo_real=(SELECT COALESCE(SUM(costo_total),0) FROM empresa_produccion_consumos WHERE empresa_id=? AND orden_id=?), fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, item.EmpresaID, item.OrdenID, item.EmpresaID, item.OrdenID)
	return id, nil
}

func ListEmpresaProduccionConsumos(dbConn *sql.DB, empresaID, ordenID int64, limit int) ([]EmpresaProduccionConsumo, error) {
	if limit <= 0 || limit > 500 {
		limit = 150
	}
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if ordenID > 0 {
		where += " AND orden_id=?"
		args = append(args, ordenID)
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,orden_id,COALESCE(producto_id,0),COALESCE(producto_nombre,''),COALESCE(cantidad_planificada,0),COALESCE(cantidad_consumida,0),COALESCE(costo_unitario,0),COALESCE(costo_total,0),COALESCE(lote_codigo,''),COALESCE(merma,0),COALESCE(fecha_consumo,''),COALESCE(usuario_creador,'') FROM empresa_produccion_consumos WHERE %s ORDER BY id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaProduccionConsumo{}
	for rows.Next() {
		var x EmpresaProduccionConsumo
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.OrdenID, &x.ProductoID, &x.ProductoNombre, &x.CantidadPlanificada, &x.CantidadConsumida, &x.CostoUnitario, &x.CostoTotal, &x.LoteCodigo, &x.Merma, &x.FechaConsumo, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func RegistrarEmpresaProduccionCalidad(dbConn *sql.DB, item EmpresaProduccionCalidad) (int64, error) {
	if err := EnsureEmpresaProduccionMRPSchema(dbConn); err != nil {
		return 0, err
	}
	item = normalizeProduccionCalidad(item)
	if item.EmpresaID <= 0 || item.OrdenID <= 0 {
		return 0, errors.New("orden_id es obligatorio")
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_produccion_calidad (empresa_id,orden_id,resultado,checklist_json,cantidad_aprobada,cantidad_rechazada,responsable,observaciones) VALUES (?,?,?,?,?,?,?,?)`,
		item.EmpresaID, item.OrdenID, item.Resultado, item.ChecklistJSON, item.CantidadAprobada, item.CantidadRechazada, item.Responsable, item.Observaciones)
	if err != nil {
		return 0, err
	}
	if item.Resultado == "aprobado" {
		_, _ = ExecCompat(dbConn, `UPDATE empresa_produccion_ordenes SET cantidad_producida=?, estado='cerrada', fecha_cierre=CAST(CURRENT_TIMESTAMP AS TEXT), fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, item.CantidadAprobada, item.EmpresaID, item.OrdenID)
	}
	return id, nil
}

func ListEmpresaProduccionCalidad(dbConn *sql.DB, empresaID, ordenID int64, limit int) ([]EmpresaProduccionCalidad, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if ordenID > 0 {
		where += " AND orden_id=?"
		args = append(args, ordenID)
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,orden_id,COALESCE(resultado,'pendiente'),COALESCE(checklist_json,''),COALESCE(cantidad_aprobada,0),COALESCE(cantidad_rechazada,0),COALESCE(responsable,''),COALESCE(observaciones,''),COALESCE(fecha_revision,'') FROM empresa_produccion_calidad WHERE %s ORDER BY id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaProduccionCalidad{}
	for rows.Next() {
		var x EmpresaProduccionCalidad
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.OrdenID, &x.Resultado, &x.ChecklistJSON, &x.CantidadAprobada, &x.CantidadRechazada, &x.Responsable, &x.Observaciones, &x.FechaRevision); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func GenerarEmpresaProduccionMRPPlan(dbConn *sql.DB, empresaID int64, periodo, usuario string) ([]EmpresaProduccionMRPPlan, error) {
	if err := EnsureEmpresaProduccionMRPSchema(dbConn); err != nil {
		return nil, err
	}
	periodo = strings.TrimSpace(periodo)
	if periodo == "" {
		periodo = "actual"
	}
	recetas, err := ListEmpresaProduccionRecetas(dbConn, empresaID, "activo", 500)
	if err != nil {
		return nil, err
	}
	_, _ = ExecCompat(dbConn, `UPDATE empresa_produccion_mrp_plan SET estado='reemplazado' WHERE empresa_id=? AND periodo=? AND estado='borrador'`, empresaID, periodo)
	for _, rec := range recetas {
		demanda := 10.0
		requerido := roundQtyProduccion(demanda * rec.CantidadBase)
		stockSeguridad := roundQtyProduccion(demanda * 0.15)
		sugerida := math.Max(0, requerido+stockSeguridad)
		_, err := ExecCompat(dbConn, `INSERT INTO empresa_produccion_mrp_plan (empresa_id,producto_id,producto_nombre,periodo,demanda_estimada,stock_actual,stock_seguridad,requerido_bruto,disponible_proyectado,cantidad_sugerida_compra,cantidad_sugerida_producir,origen,estado,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			empresaID, rec.ProductoTerminadoID, rec.ProductoTerminadoNombre, periodo, demanda, 0, stockSeguridad, requerido, -sugerida, 0, sugerida, "recetas", "borrador", usuario)
		if err != nil {
			return nil, err
		}
	}
	return ListEmpresaProduccionMRPPlan(dbConn, empresaID, periodo, 300)
}

func ListEmpresaProduccionMRPPlan(dbConn *sql.DB, empresaID int64, periodo string, limit int) ([]EmpresaProduccionMRPPlan, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if strings.TrimSpace(periodo) != "" {
		where += " AND periodo=?"
		args = append(args, strings.TrimSpace(periodo))
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,COALESCE(producto_id,0),COALESCE(producto_nombre,''),COALESCE(periodo,''),COALESCE(demanda_estimada,0),COALESCE(stock_actual,0),COALESCE(stock_seguridad,0),COALESCE(requerido_bruto,0),COALESCE(disponible_proyectado,0),COALESCE(cantidad_sugerida_compra,0),COALESCE(cantidad_sugerida_producir,0),COALESCE(origen,''),COALESCE(estado,''),COALESCE(fecha_creacion,''),COALESCE(usuario_creador,'') FROM empresa_produccion_mrp_plan WHERE %s ORDER BY id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaProduccionMRPPlan{}
	for rows.Next() {
		var x EmpresaProduccionMRPPlan
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.ProductoID, &x.ProductoNombre, &x.Periodo, &x.DemandaEstimada, &x.StockActual, &x.StockSeguridad, &x.RequeridoBruto, &x.DisponibleProyectado, &x.CantidadSugeridaCompra, &x.CantidadSugeridaProducir, &x.Origen, &x.Estado, &x.FechaCreacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func BuildEmpresaProduccionMRPDashboard(dbConn *sql.DB, empresaID int64) (EmpresaProduccionMRPDashboard, error) {
	if err := EnsureEmpresaProduccionMRPSchema(dbConn); err != nil {
		return EmpresaProduccionMRPDashboard{}, err
	}
	cfg, err := GetEmpresaProduccionMRPConfig(dbConn, empresaID)
	if err != nil {
		return EmpresaProduccionMRPDashboard{}, err
	}
	recetas, _ := ListEmpresaProduccionRecetas(dbConn, empresaID, "", 40)
	ordenes, _ := ListEmpresaProduccionOrdenes(dbConn, empresaID, "", 80)
	plan, _ := ListEmpresaProduccionMRPPlan(dbConn, empresaID, "", 80)
	consumos, _ := ListEmpresaProduccionConsumos(dbConn, empresaID, 0, 30)
	calidad, _ := ListEmpresaProduccionCalidad(dbConn, empresaID, 0, 30)
	ds := EmpresaProduccionMRPDashboard{EmpresaID: empresaID, Config: cfg, Recetas: recetas, Ordenes: ordenes, Plan: plan, ConsumosRecientes: consumos, RevisionesCalidad: calidad}
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_produccion_recetas WHERE empresa_id=? AND estado='activo'`, empresaID).Scan(&ds.RecetasActivas)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*), COALESCE(SUM(costo_estimado),0) FROM empresa_produccion_ordenes WHERE empresa_id=? AND estado IN ('borrador','programada','en_proceso')`, empresaID).Scan(&ds.OrdenesAbiertas, &ds.CostoEstimadoAbierto)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_produccion_ordenes WHERE empresa_id=? AND estado='calidad'`, empresaID).Scan(&ds.OrdenesCalidad)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_produccion_ordenes WHERE empresa_id=? AND estado='cerrada'`, empresaID).Scan(&ds.OrdenesCerradas)
	_ = QueryRowCompat(dbConn, `SELECT COALESCE(SUM(costo_real),0) FROM empresa_produccion_ordenes WHERE empresa_id=? AND estado='cerrada' AND substr(COALESCE(fecha_cierre,''),1,7)=substr(CAST(CURRENT_TIMESTAMP AS TEXT),1,7)`, empresaID).Scan(&ds.CostoRealMes)
	return ds, nil
}

func SeedEmpresaProduccionMRPDemo(dbConn *sql.DB, empresaID int64, user string) error {
	if err := UpsertEmpresaProduccionMRPConfig(dbConn, EmpresaProduccionMRPConfig{EmpresaID: empresaID, NombreSistema: "Produccion / MRP", Moneda: "COP", CosteoModo: "estandar", AprobarOrdenes: true, CerrarConCalidad: true, UsuarioCreador: user}); err != nil {
		return err
	}
	recID, err := UpsertEmpresaProduccionReceta(dbConn, EmpresaProduccionReceta{
		EmpresaID: empresaID, Codigo: "BOM-KIT-ASEO", Nombre: "Kit de aseo hotelero", ProductoTerminadoNombre: "Kit de aseo habitacion", Version: "1.0", Unidad: "kit", CantidadBase: 1, CostoEstandar: 4200, MermaPorcentaje: 2, TiempoEstimadoMin: 12, Estado: "activo", UsuarioCreador: user,
		Componentes: []EmpresaProduccionComponente{
			{ProductoNombre: "Shampoo sachet", Unidad: "und", Cantidad: 1, CostoUnitario: 900, Etapa: "alistamiento", Obligatoria: true},
			{ProductoNombre: "Jabon hotelero", Unidad: "und", Cantidad: 1, CostoUnitario: 1100, Etapa: "alistamiento", Obligatoria: true},
			{ProductoNombre: "Empaque sellado", Unidad: "und", Cantidad: 1, CostoUnitario: 450, Etapa: "empaque", Obligatoria: true},
		},
	})
	if err != nil {
		return err
	}
	orden, err := CreateEmpresaProduccionOrden(dbConn, EmpresaProduccionOrden{EmpresaID: empresaID, RecetaID: recID, ProductoTerminadoNombre: "Kit de aseo habitacion", CantidadPlanificada: 50, Estado: "programada", Prioridad: "normal", Responsable: "Operaciones", Observaciones: "Demo inicial para produccion hotelera/motelera", UsuarioCreador: user})
	if err != nil {
		return err
	}
	_, _ = RegistrarEmpresaProduccionCalidad(dbConn, EmpresaProduccionCalidad{EmpresaID: empresaID, OrdenID: orden.ID, Resultado: "pendiente", Responsable: "Calidad", ChecklistJSON: `{"empaque":"pendiente","contenido":"pendiente"}`})
	_, err = GenerarEmpresaProduccionMRPPlan(dbConn, empresaID, "demo", user)
	return err
}

func getEmpresaProduccionRecetaByID(dbConn *sql.DB, empresaID, id int64) (EmpresaProduccionReceta, error) {
	var x EmpresaProduccionReceta
	err := QueryRowCompat(dbConn, `SELECT id,empresa_id,COALESCE(codigo,''),COALESCE(nombre,''),COALESCE(producto_terminado_id,0),COALESCE(producto_terminado_nombre,''),COALESCE(version,'1.0'),COALESCE(unidad,'und'),COALESCE(cantidad_base,1),COALESCE(costo_estandar,0),COALESCE(merma_porcentaje,0),COALESCE(tiempo_estimado_min,0),COALESCE(estado,'activo'),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,''),COALESCE(usuario_creador,'') FROM empresa_produccion_recetas WHERE empresa_id=? AND id=?`, empresaID, id).Scan(&x.ID, &x.EmpresaID, &x.Codigo, &x.Nombre, &x.ProductoTerminadoID, &x.ProductoTerminadoNombre, &x.Version, &x.Unidad, &x.CantidadBase, &x.CostoEstandar, &x.MermaPorcentaje, &x.TiempoEstimadoMin, &x.Estado, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador)
	return x, err
}

func materializeProduccionConsumosFromReceta(dbConn *sql.DB, empresaID, ordenID, recetaID int64, cantidadOrden float64, user string) error {
	rec, err := getEmpresaProduccionRecetaByID(dbConn, empresaID, recetaID)
	if err != nil {
		return nil
	}
	base := rec.CantidadBase
	if base <= 0 {
		base = 1
	}
	componentes, err := ListEmpresaProduccionComponentes(dbConn, empresaID, recetaID)
	if err != nil {
		return err
	}
	for _, c := range componentes {
		cantidad := roundQtyProduccion((c.Cantidad * cantidadOrden / base) * (1 + c.MermaPorcentaje/100))
		_, err := RegistrarEmpresaProduccionConsumo(dbConn, EmpresaProduccionConsumo{EmpresaID: empresaID, OrdenID: ordenID, ProductoID: c.ProductoID, ProductoNombre: c.ProductoNombre, CantidadPlanificada: cantidad, CantidadConsumida: 0, CostoUnitario: c.CostoUnitario, Merma: roundQtyProduccion(cantidad - (c.Cantidad * cantidadOrden / base)), UsuarioCreador: user})
		if err != nil {
			return err
		}
	}
	return nil
}

func nextProduccionOrdenCode(dbConn *sql.DB, empresaID int64) (string, error) {
	var count int64
	if err := QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_produccion_ordenes WHERE empresa_id=?`, empresaID).Scan(&count); err != nil {
		return "", err
	}
	return fmt.Sprintf("OP-%06d", count+1), nil
}

func produccionOrdenSelect() string {
	return `SELECT o.id,o.empresa_id,COALESCE(o.codigo,''),COALESCE(o.receta_id,0),COALESCE(r.nombre,''),COALESCE(o.producto_terminado_id,0),COALESCE(o.producto_terminado_nombre,''),COALESCE(o.cantidad_planificada,0),COALESCE(o.cantidad_producida,0),COALESCE(o.estado,'borrador'),COALESCE(o.prioridad,'normal'),COALESCE(o.fecha_programada,''),COALESCE(o.fecha_inicio,''),COALESCE(o.fecha_cierre,''),COALESCE(o.costo_estimado,0),COALESCE(o.costo_real,0),COALESCE(o.responsable,''),COALESCE(o.observaciones,''),COALESCE(o.fecha_creacion,''),COALESCE(o.fecha_actualizacion,''),COALESCE(o.usuario_creador,'') FROM empresa_produccion_ordenes o LEFT JOIN empresa_produccion_recetas r ON r.id=o.receta_id AND r.empresa_id=o.empresa_id`
}

func normalizeProduccionMRPConfig(cfg EmpresaProduccionMRPConfig) EmpresaProduccionMRPConfig {
	if cfg.NombreSistema = strings.TrimSpace(cfg.NombreSistema); cfg.NombreSistema == "" {
		cfg.NombreSistema = "Produccion / MRP"
	}
	cfg.Moneda = strings.ToUpper(strings.TrimSpace(cfg.Moneda))
	if cfg.Moneda == "" {
		cfg.Moneda = "COP"
	}
	cfg.CosteoModo = normalizeOneOfProduccion(cfg.CosteoModo, "estandar", "promedio", "estandar", "real")
	return cfg
}

func normalizeProduccionReceta(x EmpresaProduccionReceta) EmpresaProduccionReceta {
	x.Codigo = strings.ToUpper(strings.TrimSpace(x.Codigo))
	x.Nombre = strings.TrimSpace(x.Nombre)
	x.ProductoTerminadoNombre = strings.TrimSpace(x.ProductoTerminadoNombre)
	if x.ProductoTerminadoNombre == "" {
		x.ProductoTerminadoNombre = x.Nombre
	}
	x.Version = strings.TrimSpace(x.Version)
	if x.Version == "" {
		x.Version = "1.0"
	}
	x.Unidad = strings.ToLower(strings.TrimSpace(x.Unidad))
	if x.Unidad == "" {
		x.Unidad = "und"
	}
	if x.CantidadBase <= 0 {
		x.CantidadBase = 1
	}
	if x.TiempoEstimadoMin < 0 {
		x.TiempoEstimadoMin = 0
	}
	x.MermaPorcentaje = clampProduccion(x.MermaPorcentaje, 0, 100)
	x.Estado = normalizeProduccionEstadoReceta(x.Estado)
	return x
}

func normalizeProduccionComponente(x EmpresaProduccionComponente) EmpresaProduccionComponente {
	x.ProductoNombre = strings.TrimSpace(x.ProductoNombre)
	x.Unidad = strings.ToLower(strings.TrimSpace(x.Unidad))
	if x.Unidad == "" {
		x.Unidad = "und"
	}
	if x.Cantidad < 0 {
		x.Cantidad = 0
	}
	if x.CostoUnitario < 0 {
		x.CostoUnitario = 0
	}
	x.MermaPorcentaje = clampProduccion(x.MermaPorcentaje, 0, 100)
	x.Etapa = normalizeSlugProduccion(x.Etapa, "preparacion")
	if !x.Obligatoria {
		x.Obligatoria = true
	}
	return x
}

func normalizeProduccionOrden(x EmpresaProduccionOrden) EmpresaProduccionOrden {
	x.Codigo = strings.ToUpper(strings.TrimSpace(x.Codigo))
	x.ProductoTerminadoNombre = strings.TrimSpace(x.ProductoTerminadoNombre)
	if x.CantidadPlanificada <= 0 {
		x.CantidadPlanificada = 1
	}
	if x.CantidadProducida < 0 {
		x.CantidadProducida = 0
	}
	x.Estado = normalizeProduccionEstadoOrden(x.Estado)
	x.Prioridad = normalizeOneOfProduccion(x.Prioridad, "normal", "baja", "normal", "alta", "urgente")
	x.Responsable = strings.TrimSpace(x.Responsable)
	x.Observaciones = strings.TrimSpace(x.Observaciones)
	return x
}

func normalizeProduccionConsumo(x EmpresaProduccionConsumo) EmpresaProduccionConsumo {
	x.ProductoNombre = strings.TrimSpace(x.ProductoNombre)
	if x.CantidadPlanificada < 0 {
		x.CantidadPlanificada = 0
	}
	if x.CantidadConsumida < 0 {
		x.CantidadConsumida = 0
	}
	if x.CostoUnitario < 0 {
		x.CostoUnitario = 0
	}
	x.LoteCodigo = strings.TrimSpace(x.LoteCodigo)
	return x
}

func normalizeProduccionCalidad(x EmpresaProduccionCalidad) EmpresaProduccionCalidad {
	x.Resultado = normalizeOneOfProduccion(x.Resultado, "pendiente", "pendiente", "aprobado", "rechazado", "reproceso")
	if x.CantidadAprobada < 0 {
		x.CantidadAprobada = 0
	}
	if x.CantidadRechazada < 0 {
		x.CantidadRechazada = 0
	}
	x.Responsable = strings.TrimSpace(x.Responsable)
	x.Observaciones = strings.TrimSpace(x.Observaciones)
	return x
}

func normalizeProduccionEstadoReceta(v string) string {
	return normalizeOneOfProduccion(v, "activo", "activo", "inactivo", "borrador")
}

func normalizeProduccionEstadoOrden(v string) string {
	return normalizeOneOfProduccion(v, "borrador", "borrador", "programada", "en_proceso", "calidad", "cerrada", "cancelada")
}

func normalizeOneOfProduccion(v, fallback string, allowed ...string) string {
	s := normalizeSlugProduccion(v, fallback)
	for _, a := range allowed {
		if s == a {
			return s
		}
	}
	return fallback
}

func normalizeSlugProduccion(v, fallback string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	if s == "" {
		return fallback
	}
	return s
}

func clampProduccion(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func roundMoneyProduccion(v float64) float64 {
	return math.Round(v*100) / 100
}

func roundQtyProduccion(v float64) float64 {
	return math.Round(v*10000) / 10000
}

func boolIntProduccion(v bool) int {
	if v {
		return 1
	}
	return 0
}
