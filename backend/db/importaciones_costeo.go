package db

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

type EmpresaImportacionCosteo struct {
	ID                       int64                     `json:"id"`
	EmpresaID                int64                     `json:"empresa_id"`
	Codigo                   string                    `json:"codigo"`
	Proveedor                string                    `json:"proveedor"`
	PaisOrigen               string                    `json:"pais_origen"`
	Incoterm                 string                    `json:"incoterm"`
	MonedaOrigen             string                    `json:"moneda_origen"`
	TRM                      float64                   `json:"trm"`
	FechaDocumento           string                    `json:"fecha_documento"`
	FechaEstimadaLlegada     string                    `json:"fecha_estimada_llegada,omitempty"`
	DocumentoReferencia      string                    `json:"documento_referencia,omitempty"`
	Estado                   string                    `json:"estado"`
	SubtotalOrigen           float64                   `json:"subtotal_origen"`
	SubtotalCOP              float64                   `json:"subtotal_cop"`
	CostosNacionalizacionCOP float64                   `json:"costos_nacionalizacion_cop"`
	CostoTotalCOP            float64                   `json:"costo_total_cop"`
	UsuarioCreador           string                    `json:"usuario_creador,omitempty"`
	FechaCreacion            string                    `json:"fecha_creacion,omitempty"`
	FechaActualizacion       string                    `json:"fecha_actualizacion,omitempty"`
	Items                    []EmpresaImportacionItem  `json:"items,omitempty"`
	Costos                   []EmpresaImportacionCosto `json:"costos,omitempty"`
}

type EmpresaImportacionItem struct {
	ID                    int64   `json:"id"`
	EmpresaID             int64   `json:"empresa_id"`
	ImportacionID         int64   `json:"importacion_id"`
	ProductoID            int64   `json:"producto_id,omitempty"`
	ProductoNombre        string  `json:"producto_nombre"`
	SKU                   string  `json:"sku,omitempty"`
	Cantidad              float64 `json:"cantidad"`
	Unidad                string  `json:"unidad"`
	PesoKG                float64 `json:"peso_kg"`
	VolumenM3             float64 `json:"volumen_m3"`
	CostoUnitarioOrigen   float64 `json:"costo_unitario_origen"`
	CostoOrigen           float64 `json:"costo_origen"`
	CostoBaseCOP          float64 `json:"costo_base_cop"`
	CostoDistribuidoCOP   float64 `json:"costo_distribuido_cop"`
	CostoUnitarioFinalCOP float64 `json:"costo_unitario_final_cop"`
	Estado                string  `json:"estado"`
}

type EmpresaImportacionCosto struct {
	ID               int64   `json:"id"`
	EmpresaID        int64   `json:"empresa_id"`
	ImportacionID    int64   `json:"importacion_id"`
	Tipo             string  `json:"tipo"`
	Concepto         string  `json:"concepto"`
	BaseDistribucion string  `json:"base_distribucion"`
	ValorCOP         float64 `json:"valor_cop"`
	Tercero          string  `json:"tercero,omitempty"`
	Documento        string  `json:"documento,omitempty"`
	CuentaContable   string  `json:"cuenta_contable,omitempty"`
	Estado           string  `json:"estado"`
	UsuarioCreador   string  `json:"usuario_creador,omitempty"`
	FechaCreacion    string  `json:"fecha_creacion,omitempty"`
}

type EmpresaImportacionesCosteoDashboard struct {
	EmpresaID             int64                      `json:"empresa_id"`
	ImportacionesAbiertas int                        `json:"importaciones_abiertas"`
	ImportacionesCerradas int                        `json:"importaciones_cerradas"`
	CostosPendientesCOP   float64                    `json:"costos_pendientes_cop"`
	CostoTotalCOP         float64                    `json:"costo_total_cop"`
	UltimasImportaciones  []EmpresaImportacionCosteo `json:"ultimas_importaciones"`
}

func EnsureEmpresaImportacionesCosteoSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_importaciones_costeo (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			proveedor TEXT DEFAULT '',
			pais_origen TEXT DEFAULT '',
			incoterm TEXT DEFAULT 'FOB',
			moneda_origen TEXT DEFAULT 'USD',
			trm REAL DEFAULT 1,
			fecha_documento TEXT NOT NULL,
			fecha_estimada_llegada TEXT,
			documento_referencia TEXT,
			estado TEXT DEFAULT 'borrador',
			subtotal_origen REAL DEFAULT 0,
			subtotal_cop REAL DEFAULT 0,
			costos_nacionalizacion_cop REAL DEFAULT 0,
			costo_total_cop REAL DEFAULT 0,
			usuario_creador TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(empresa_id, codigo)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_importaciones_costeo_empresa_estado ON empresa_importaciones_costeo(empresa_id, estado)`,
		`CREATE TABLE IF NOT EXISTS empresa_importaciones_costeo_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			importacion_id INTEGER NOT NULL,
			producto_id INTEGER DEFAULT 0,
			producto_nombre TEXT NOT NULL,
			sku TEXT,
			cantidad REAL DEFAULT 0,
			unidad TEXT DEFAULT 'und',
			peso_kg REAL DEFAULT 0,
			volumen_m3 REAL DEFAULT 0,
			costo_unitario_origen REAL DEFAULT 0,
			costo_origen REAL DEFAULT 0,
			costo_base_cop REAL DEFAULT 0,
			costo_distribuido_cop REAL DEFAULT 0,
			costo_unitario_final_cop REAL DEFAULT 0,
			estado TEXT DEFAULT 'activo'
		)`,
		`CREATE INDEX IF NOT EXISTS ix_importaciones_items_importacion ON empresa_importaciones_costeo_items(empresa_id, importacion_id)`,
		`CREATE TABLE IF NOT EXISTS empresa_importaciones_costeo_costos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			importacion_id INTEGER NOT NULL,
			tipo TEXT DEFAULT 'nacionalizacion',
			concepto TEXT NOT NULL,
			base_distribucion TEXT DEFAULT 'valor',
			valor_cop REAL DEFAULT 0,
			tercero TEXT,
			documento TEXT,
			cuenta_contable TEXT,
			estado TEXT DEFAULT 'activo',
			usuario_creador TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS ix_importaciones_costos_importacion ON empresa_importaciones_costeo_costos(empresa_id, importacion_id)`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func CreateEmpresaImportacionCosteo(dbConn *sql.DB, x EmpresaImportacionCosteo) (int64, error) {
	x = normalizeImportacionCosteo(x)
	if x.EmpresaID <= 0 || x.Codigo == "" {
		return 0, errors.New("empresa_id y codigo son requeridos")
	}
	if x.FechaDocumento == "" {
		x.FechaDocumento = time.Now().Format("2006-01-02")
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_importaciones_costeo
		(empresa_id,codigo,proveedor,pais_origen,incoterm,moneda_origen,trm,fecha_documento,fecha_estimada_llegada,documento_referencia,estado,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (empresa_id,codigo) DO UPDATE SET
			proveedor=EXCLUDED.proveedor,
			pais_origen=EXCLUDED.pais_origen,
			incoterm=EXCLUDED.incoterm,
			moneda_origen=EXCLUDED.moneda_origen,
			trm=EXCLUDED.trm,
			fecha_documento=EXCLUDED.fecha_documento,
			fecha_estimada_llegada=EXCLUDED.fecha_estimada_llegada,
			documento_referencia=EXCLUDED.documento_referencia,
			estado=EXCLUDED.estado,
			fecha_actualizacion=CURRENT_TIMESTAMP,
			usuario_creador=EXCLUDED.usuario_creador`,
		x.EmpresaID, x.Codigo, x.Proveedor, x.PaisOrigen, x.Incoterm, x.MonedaOrigen, x.TRM, x.FechaDocumento, x.FechaEstimadaLlegada, x.DocumentoReferencia, x.Estado, x.UsuarioCreador)
}

func CreateEmpresaImportacionItem(dbConn *sql.DB, item EmpresaImportacionItem, trm float64) (int64, error) {
	item = normalizeImportacionItem(item, trm)
	if item.EmpresaID <= 0 || item.ImportacionID <= 0 || item.ProductoNombre == "" {
		return 0, errors.New("importacion y producto son requeridos")
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_importaciones_costeo_items
		(empresa_id,importacion_id,producto_id,producto_nombre,sku,cantidad,unidad,peso_kg,volumen_m3,costo_unitario_origen,costo_origen,costo_base_cop,costo_distribuido_cop,costo_unitario_final_cop,estado)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		item.EmpresaID, item.ImportacionID, item.ProductoID, item.ProductoNombre, item.SKU, item.Cantidad, item.Unidad, item.PesoKG, item.VolumenM3, item.CostoUnitarioOrigen, item.CostoOrigen, item.CostoBaseCOP, item.CostoDistribuidoCOP, item.CostoUnitarioFinalCOP, item.Estado)
	if err != nil {
		return 0, err
	}
	return id, recalcularEmpresaImportacionTotales(dbConn, item.EmpresaID, item.ImportacionID)
}

func CreateEmpresaImportacionCosto(dbConn *sql.DB, costo EmpresaImportacionCosto) (int64, error) {
	costo = normalizeImportacionCosto(costo)
	if costo.EmpresaID <= 0 || costo.ImportacionID <= 0 || costo.Concepto == "" {
		return 0, errors.New("importacion y concepto son requeridos")
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_importaciones_costeo_costos
		(empresa_id,importacion_id,tipo,concepto,base_distribucion,valor_cop,tercero,documento,cuenta_contable,estado,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		costo.EmpresaID, costo.ImportacionID, costo.Tipo, costo.Concepto, costo.BaseDistribucion, costo.ValorCOP, costo.Tercero, costo.Documento, costo.CuentaContable, costo.Estado, costo.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	return id, recalcularEmpresaImportacionTotales(dbConn, costo.EmpresaID, costo.ImportacionID)
}

func ListEmpresaImportacionesCosteo(dbConn *sql.DB, empresaID int64, estado string, limit int) ([]EmpresaImportacionCosteo, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if strings.TrimSpace(estado) != "" {
		where += " AND estado=?"
		args = append(args, normalizeImportacionEstado(estado))
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,codigo,COALESCE(proveedor,''),COALESCE(pais_origen,''),COALESCE(incoterm,'FOB'),COALESCE(moneda_origen,'USD'),COALESCE(trm,1),COALESCE(fecha_documento,''),COALESCE(fecha_estimada_llegada,''),COALESCE(documento_referencia,''),COALESCE(estado,'borrador'),COALESCE(subtotal_origen,0),COALESCE(subtotal_cop,0),COALESCE(costos_nacionalizacion_cop,0),COALESCE(costo_total_cop,0),COALESCE(usuario_creador,''),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,'') FROM empresa_importaciones_costeo WHERE %s ORDER BY id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaImportacionCosteo{}
	for rows.Next() {
		var x EmpresaImportacionCosteo
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Codigo, &x.Proveedor, &x.PaisOrigen, &x.Incoterm, &x.MonedaOrigen, &x.TRM, &x.FechaDocumento, &x.FechaEstimadaLlegada, &x.DocumentoReferencia, &x.Estado, &x.SubtotalOrigen, &x.SubtotalCOP, &x.CostosNacionalizacionCOP, &x.CostoTotalCOP, &x.UsuarioCreador, &x.FechaCreacion, &x.FechaActualizacion); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func GetEmpresaImportacionCosteo(dbConn *sql.DB, empresaID, id int64) (EmpresaImportacionCosteo, error) {
	rows, err := ListEmpresaImportacionesCosteo(dbConn, empresaID, "", 500)
	if err != nil {
		return EmpresaImportacionCosteo{}, err
	}
	for _, row := range rows {
		if row.ID == id {
			items, _ := ListEmpresaImportacionItems(dbConn, empresaID, id)
			costos, _ := ListEmpresaImportacionCostos(dbConn, empresaID, id)
			row.Items = items
			row.Costos = costos
			return row, nil
		}
	}
	return EmpresaImportacionCosteo{}, sql.ErrNoRows
}

func ListEmpresaImportacionItems(dbConn *sql.DB, empresaID, importacionID int64) ([]EmpresaImportacionItem, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,importacion_id,COALESCE(producto_id,0),COALESCE(producto_nombre,''),COALESCE(sku,''),COALESCE(cantidad,0),COALESCE(unidad,'und'),COALESCE(peso_kg,0),COALESCE(volumen_m3,0),COALESCE(costo_unitario_origen,0),COALESCE(costo_origen,0),COALESCE(costo_base_cop,0),COALESCE(costo_distribuido_cop,0),COALESCE(costo_unitario_final_cop,0),COALESCE(estado,'activo') FROM empresa_importaciones_costeo_items WHERE empresa_id=? AND importacion_id=? ORDER BY id`, empresaID, importacionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaImportacionItem{}
	for rows.Next() {
		var x EmpresaImportacionItem
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.ImportacionID, &x.ProductoID, &x.ProductoNombre, &x.SKU, &x.Cantidad, &x.Unidad, &x.PesoKG, &x.VolumenM3, &x.CostoUnitarioOrigen, &x.CostoOrigen, &x.CostoBaseCOP, &x.CostoDistribuidoCOP, &x.CostoUnitarioFinalCOP, &x.Estado); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func ListEmpresaImportacionCostos(dbConn *sql.DB, empresaID, importacionID int64) ([]EmpresaImportacionCosto, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,importacion_id,COALESCE(tipo,'nacionalizacion'),COALESCE(concepto,''),COALESCE(base_distribucion,'valor'),COALESCE(valor_cop,0),COALESCE(tercero,''),COALESCE(documento,''),COALESCE(cuenta_contable,''),COALESCE(estado,'activo'),COALESCE(usuario_creador,''),COALESCE(fecha_creacion,'') FROM empresa_importaciones_costeo_costos WHERE empresa_id=? AND importacion_id=? ORDER BY id`, empresaID, importacionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaImportacionCosto{}
	for rows.Next() {
		var x EmpresaImportacionCosto
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.ImportacionID, &x.Tipo, &x.Concepto, &x.BaseDistribucion, &x.ValorCOP, &x.Tercero, &x.Documento, &x.CuentaContable, &x.Estado, &x.UsuarioCreador, &x.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func DistribuirEmpresaImportacionCostos(dbConn *sql.DB, empresaID, importacionID int64, usuario string) (EmpresaImportacionCosteo, error) {
	imp, err := GetEmpresaImportacionCosteo(dbConn, empresaID, importacionID)
	if err != nil {
		return EmpresaImportacionCosteo{}, err
	}
	items := imp.Items
	costos := imp.Costos
	if len(items) == 0 {
		return EmpresaImportacionCosteo{}, errors.New("la importacion no tiene items")
	}
	distribuido := map[int64]float64{}
	for _, c := range costos {
		baseTotal := 0.0
		for _, it := range items {
			baseTotal += importacionBaseDistribucionItem(it, c.BaseDistribucion)
		}
		if baseTotal <= 0 {
			baseTotal = float64(len(items))
		}
		for _, it := range items {
			base := importacionBaseDistribucionItem(it, c.BaseDistribucion)
			if base <= 0 {
				base = 1
			}
			distribuido[it.ID] += roundImportacion(c.ValorCOP * base / baseTotal)
		}
	}
	for _, it := range items {
		dist := roundImportacion(distribuido[it.ID])
		finalTotal := roundImportacion(it.CostoBaseCOP + dist)
		unitFinal := 0.0
		if it.Cantidad > 0 {
			unitFinal = roundImportacion(finalTotal / it.Cantidad)
		}
		if _, err := ExecCompat(dbConn, `UPDATE empresa_importaciones_costeo_items SET costo_distribuido_cop=?, costo_unitario_final_cop=? WHERE empresa_id=? AND id=?`, dist, unitFinal, empresaID, it.ID); err != nil {
			return EmpresaImportacionCosteo{}, err
		}
	}
	if _, err := ExecCompat(dbConn, `UPDATE empresa_importaciones_costeo SET estado='costeado', fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=COALESCE(NULLIF(?,''),usuario_creador) WHERE empresa_id=? AND id=?`, usuario, empresaID, importacionID); err != nil {
		return EmpresaImportacionCosteo{}, err
	}
	if err := recalcularEmpresaImportacionTotales(dbConn, empresaID, importacionID); err != nil {
		return EmpresaImportacionCosteo{}, err
	}
	return GetEmpresaImportacionCosteo(dbConn, empresaID, importacionID)
}

func BuildEmpresaImportacionesCosteoDashboard(dbConn *sql.DB, empresaID int64) (EmpresaImportacionesCosteoDashboard, error) {
	ds := EmpresaImportacionesCosteoDashboard{EmpresaID: empresaID}
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_importaciones_costeo WHERE empresa_id=? AND estado IN ('borrador','en_transito','costeado')`, empresaID).Scan(&ds.ImportacionesAbiertas)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_importaciones_costeo WHERE empresa_id=? AND estado IN ('cerrado','contabilizado')`, empresaID).Scan(&ds.ImportacionesCerradas)
	_ = QueryRowCompat(dbConn, `SELECT COALESCE(SUM(costos_nacionalizacion_cop),0),COALESCE(SUM(costo_total_cop),0) FROM empresa_importaciones_costeo WHERE empresa_id=?`, empresaID).Scan(&ds.CostosPendientesCOP, &ds.CostoTotalCOP)
	rows, err := ListEmpresaImportacionesCosteo(dbConn, empresaID, "", 20)
	if err != nil {
		return ds, err
	}
	ds.UltimasImportaciones = rows
	return ds, nil
}

func SeedEmpresaImportacionesCosteoDemo(dbConn *sql.DB, empresaID int64, usuario string) error {
	if err := EnsureEmpresaImportacionesCosteoSchema(dbConn); err != nil {
		return err
	}
	run := time.Now().Format("20060102150405")
	id, err := CreateEmpresaImportacionCosteo(dbConn, EmpresaImportacionCosteo{EmpresaID: empresaID, Codigo: "IMP-DEMO-" + run, Proveedor: "Proveedor internacional demo", PaisOrigen: "China", Incoterm: "FOB", MonedaOrigen: "USD", TRM: 3900, FechaDocumento: time.Now().Format("2006-01-02"), Estado: "en_transito", UsuarioCreador: usuario})
	if err != nil {
		return err
	}
	if _, err := CreateEmpresaImportacionItem(dbConn, EmpresaImportacionItem{EmpresaID: empresaID, ImportacionID: id, ProductoNombre: "Sensor importado demo", SKU: "IMP-SEN", Cantidad: 50, Unidad: "und", PesoKG: 12, VolumenM3: 0.25, CostoUnitarioOrigen: 9}, 3900); err != nil {
		return err
	}
	if _, err := CreateEmpresaImportacionItem(dbConn, EmpresaImportacionItem{EmpresaID: empresaID, ImportacionID: id, ProductoNombre: "Controlador importado demo", SKU: "IMP-CTRL", Cantidad: 20, Unidad: "und", PesoKG: 18, VolumenM3: 0.4, CostoUnitarioOrigen: 34}, 3900); err != nil {
		return err
	}
	for _, c := range []EmpresaImportacionCosto{
		{EmpresaID: empresaID, ImportacionID: id, Tipo: "flete", Concepto: "Flete internacional demo", BaseDistribucion: "peso", ValorCOP: 1450000, UsuarioCreador: usuario},
		{EmpresaID: empresaID, ImportacionID: id, Tipo: "arancel", Concepto: "Arancel e IVA demo", BaseDistribucion: "valor", ValorCOP: 980000, UsuarioCreador: usuario},
	} {
		if _, err := CreateEmpresaImportacionCosto(dbConn, c); err != nil {
			return err
		}
	}
	_, err = DistribuirEmpresaImportacionCostos(dbConn, empresaID, id, usuario)
	return err
}

func recalcularEmpresaImportacionTotales(dbConn *sql.DB, empresaID, importacionID int64) error {
	var subtotalOrigen, subtotalCOP, costosCOP float64
	_ = QueryRowCompat(dbConn, `SELECT COALESCE(SUM(costo_origen),0),COALESCE(SUM(costo_base_cop),0) FROM empresa_importaciones_costeo_items WHERE empresa_id=? AND importacion_id=?`, empresaID, importacionID).Scan(&subtotalOrigen, &subtotalCOP)
	_ = QueryRowCompat(dbConn, `SELECT COALESCE(SUM(valor_cop),0) FROM empresa_importaciones_costeo_costos WHERE empresa_id=? AND importacion_id=? AND estado='activo'`, empresaID, importacionID).Scan(&costosCOP)
	_, err := ExecCompat(dbConn, `UPDATE empresa_importaciones_costeo SET subtotal_origen=?, subtotal_cop=?, costos_nacionalizacion_cop=?, costo_total_cop=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, roundImportacion(subtotalOrigen), roundImportacion(subtotalCOP), roundImportacion(costosCOP), roundImportacion(subtotalCOP+costosCOP), empresaID, importacionID)
	return err
}

func normalizeImportacionCosteo(x EmpresaImportacionCosteo) EmpresaImportacionCosteo {
	x.Codigo = strings.ToUpper(strings.TrimSpace(x.Codigo))
	x.Proveedor = strings.TrimSpace(x.Proveedor)
	x.PaisOrigen = strings.TrimSpace(x.PaisOrigen)
	x.Incoterm = strings.ToUpper(firstImportacionValue(x.Incoterm, "FOB"))
	x.MonedaOrigen = strings.ToUpper(firstImportacionValue(x.MonedaOrigen, "USD"))
	if x.TRM <= 0 {
		x.TRM = 1
	}
	x.Estado = normalizeImportacionEstado(x.Estado)
	return x
}

func normalizeImportacionItem(x EmpresaImportacionItem, trm float64) EmpresaImportacionItem {
	x.ProductoNombre = strings.TrimSpace(x.ProductoNombre)
	x.SKU = strings.ToUpper(strings.TrimSpace(x.SKU))
	x.Unidad = strings.ToLower(firstImportacionValue(x.Unidad, "und"))
	if x.Cantidad < 0 {
		x.Cantidad = 0
	}
	if x.CostoOrigen <= 0 {
		x.CostoOrigen = roundImportacion(x.Cantidad * x.CostoUnitarioOrigen)
	}
	if trm <= 0 {
		trm = 1
	}
	x.CostoBaseCOP = roundImportacion(x.CostoOrigen * trm)
	if x.Cantidad > 0 {
		x.CostoUnitarioFinalCOP = roundImportacion((x.CostoBaseCOP + x.CostoDistribuidoCOP) / x.Cantidad)
	}
	x.Estado = firstImportacionValue(x.Estado, "activo")
	return x
}

func normalizeImportacionCosto(x EmpresaImportacionCosto) EmpresaImportacionCosto {
	x.Tipo = strings.ToLower(firstImportacionValue(x.Tipo, "nacionalizacion"))
	x.Concepto = strings.TrimSpace(x.Concepto)
	x.BaseDistribucion = normalizeImportacionBase(x.BaseDistribucion)
	if x.ValorCOP < 0 {
		x.ValorCOP = 0
	}
	x.Estado = firstImportacionValue(x.Estado, "activo")
	return x
}

func normalizeImportacionEstado(v string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case "borrador", "en_transito", "costeado", "cerrado", "contabilizado", "anulado":
		return s
	default:
		return "borrador"
	}
}

func normalizeImportacionBase(v string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case "peso", "volumen", "cantidad":
		return s
	default:
		return "valor"
	}
}

func importacionBaseDistribucionItem(it EmpresaImportacionItem, base string) float64 {
	switch normalizeImportacionBase(base) {
	case "peso":
		return it.PesoKG
	case "volumen":
		return it.VolumenM3
	case "cantidad":
		return it.Cantidad
	default:
		return it.CostoBaseCOP
	}
}

func firstImportacionValue(v, fallback string) string {
	s := strings.TrimSpace(v)
	if s == "" {
		return fallback
	}
	return s
}

func roundImportacion(v float64) float64 {
	if v < 0 {
		v = 0
	}
	return math.Round(v*100) / 100
}
