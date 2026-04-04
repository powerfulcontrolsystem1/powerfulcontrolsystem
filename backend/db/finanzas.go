package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// EmpresaFinanzasMovimiento representa un registro de ingreso/egreso por empresa.
type EmpresaFinanzasMovimiento struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	TipoMovimiento     string  `json:"tipo_movimiento"`
	Codigo             string  `json:"codigo"`
	FechaMovimiento    string  `json:"fecha_movimiento"`
	Categoria          string  `json:"categoria"`
	Subcategoria       string  `json:"subcategoria"`
	Concepto           string  `json:"concepto"`
	Descripcion        string  `json:"descripcion"`
	MetodoPago         string  `json:"metodo_pago"`
	Moneda             string  `json:"moneda"`
	Monto              float64 `json:"monto"`
	Impuesto           float64 `json:"impuesto"`
	Total              float64 `json:"total"`
	TerceroNombre      string  `json:"tercero_nombre"`
	TerceroDocumento   string  `json:"tercero_documento"`
	TipoComprobante    string  `json:"tipo_comprobante"`
	NumeroComprobante  string  `json:"numero_comprobante"`
	ComprobanteURL     string  `json:"comprobante_url"`
	ReferenciaExterna  string  `json:"referencia_externa"`
	AprobadoPor        string  `json:"aprobado_por"`
	FechaCreacion      string  `json:"fecha_creacion"`
	FechaActualizacion string  `json:"fecha_actualizacion"`
	UsuarioCreador     string  `json:"usuario_creador"`
	Estado             string  `json:"estado"`
	Observaciones      string  `json:"observaciones"`
}

// EmpresaFinanzasMovimientoFilter permite filtrar listados financieros por empresa.
type EmpresaFinanzasMovimientoFilter struct {
	Tipo            string
	Desde           string
	Hasta           string
	Q               string
	IncludeInactive bool
	Limit           int
}

// EmpresaFinanzasConfiguracion define la parametrizacion por empresa del modulo financiero.
type EmpresaFinanzasConfiguracion struct {
	ID                         int64  `json:"id"`
	EmpresaID                  int64  `json:"empresa_id"`
	HabilitarIngresos          bool   `json:"habilitar_ingresos"`
	HabilitarEgresos           bool   `json:"habilitar_egresos"`
	Moneda                     string `json:"moneda"`
	CategoriasIngreso          string `json:"categorias_ingreso"`
	CategoriasEgreso           string `json:"categorias_egreso"`
	PrefijoIngreso             string `json:"prefijo_ingreso"`
	PrefijoEgreso              string `json:"prefijo_egreso"`
	FormatoImpresion           string `json:"formato_impresion"`
	RequiereAprobacion         bool   `json:"requiere_aprobacion"`
	IntegracionContableDestino string `json:"integracion_contable_destino"`
	CuentaCajaBancos           string `json:"cuenta_caja_bancos"`
	CuentaIngresos             string `json:"cuenta_ingresos"`
	CuentaIVAGenerado          string `json:"cuenta_iva_generado"`
	CuentaGastos               string `json:"cuenta_gastos"`
	CuentaIVADescontable       string `json:"cuenta_iva_descontable"`
	CuentasIngresoCategoria    string `json:"cuentas_ingreso_categoria"`
	CuentasEgresoCategoria     string `json:"cuentas_egreso_categoria"`
	FechaCreacion              string `json:"fecha_creacion"`
	FechaActualizacion         string `json:"fecha_actualizacion"`
	UsuarioCreador             string `json:"usuario_creador"`
	Estado                     string `json:"estado"`
	Observaciones              string `json:"observaciones"`
}

// EnsureEmpresaFinanzasSchema crea y migra las tablas del modulo financiero en empresas.db.
func EnsureEmpresaFinanzasSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_finanzas_movimientos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			tipo_movimiento TEXT NOT NULL,
			codigo TEXT NOT NULL,
			fecha_movimiento TEXT DEFAULT (datetime('now','localtime')),
			categoria TEXT,
			subcategoria TEXT,
			concepto TEXT,
			descripcion TEXT,
			metodo_pago TEXT,
			moneda TEXT DEFAULT 'COP',
			monto REAL DEFAULT 0,
			impuesto REAL DEFAULT 0,
			total REAL DEFAULT 0,
			tercero_nombre TEXT,
			tercero_documento TEXT,
			tipo_comprobante TEXT DEFAULT 'recibo_interno',
			numero_comprobante TEXT,
			comprobante_url TEXT,
			referencia_externa TEXT,
			aprobado_por TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE TABLE IF NOT EXISTS empresa_finanzas_configuracion (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL UNIQUE,
			habilitar_ingresos INTEGER DEFAULT 1,
			habilitar_egresos INTEGER DEFAULT 1,
			moneda TEXT DEFAULT 'COP',
			categorias_ingreso TEXT,
			categorias_egreso TEXT,
			prefijo_ingreso TEXT DEFAULT 'ING',
			prefijo_egreso TEXT DEFAULT 'EGR',
			formato_impresion TEXT DEFAULT 'carta',
			requiere_aprobacion INTEGER DEFAULT 0,
			integracion_contable_destino TEXT DEFAULT 'generico',
			cuenta_caja_bancos TEXT DEFAULT '110505',
			cuenta_ingresos TEXT DEFAULT '413595',
			cuenta_iva_generado TEXT DEFAULT '240805',
			cuenta_gastos TEXT DEFAULT '519595',
			cuenta_iva_descontable TEXT DEFAULT '240810',
			cuentas_ingreso_categoria TEXT,
			cuentas_egreso_categoria TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_finanzas_movimientos_empresa_fecha ON empresa_finanzas_movimientos(empresa_id, fecha_movimiento DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_finanzas_movimientos_empresa_tipo_estado ON empresa_finanzas_movimientos(empresa_id, tipo_movimiento, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_finanzas_movimientos_empresa_comprobante ON empresa_finanzas_movimientos(empresa_id, numero_comprobante);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_movimientos", "fecha_movimiento", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_movimientos", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_movimientos", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_movimientos", "observaciones", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_movimientos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_movimientos", "tipo_comprobante", "TEXT DEFAULT 'recibo_interno'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_movimientos", "numero_comprobante", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_movimientos", "comprobante_url", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "habilitar_ingresos", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "habilitar_egresos", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "prefijo_ingreso", "TEXT DEFAULT 'ING'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "prefijo_egreso", "TEXT DEFAULT 'EGR'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "formato_impresion", "TEXT DEFAULT 'carta'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "requiere_aprobacion", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "integracion_contable_destino", "TEXT DEFAULT 'generico'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "cuenta_caja_bancos", "TEXT DEFAULT '110505'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "cuenta_ingresos", "TEXT DEFAULT '413595'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "cuenta_iva_generado", "TEXT DEFAULT '240805'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "cuenta_gastos", "TEXT DEFAULT '519595'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "cuenta_iva_descontable", "TEXT DEFAULT '240810'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "cuentas_ingreso_categoria", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "cuentas_egreso_categoria", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_finanzas_configuracion", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

// GetEmpresaFinanzasConfiguracion obtiene configuracion financiera por empresa con defaults seguros.
func GetEmpresaFinanzasConfiguracion(dbConn *sql.DB, empresaID int64) (*EmpresaFinanzasConfiguracion, error) {
	cfg := defaultEmpresaFinanzasConfiguracion(empresaID)
	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(habilitar_ingresos, 1),
		COALESCE(habilitar_egresos, 1),
		COALESCE(moneda, 'COP'),
		COALESCE(categorias_ingreso, ''),
		COALESCE(categorias_egreso, ''),
		COALESCE(prefijo_ingreso, 'ING'),
		COALESCE(prefijo_egreso, 'EGR'),
		COALESCE(formato_impresion, 'carta'),
		COALESCE(requiere_aprobacion, 0),
		COALESCE(integracion_contable_destino, 'generico'),
		COALESCE(cuenta_caja_bancos, '110505'),
		COALESCE(cuenta_ingresos, '413595'),
		COALESCE(cuenta_iva_generado, '240805'),
		COALESCE(cuenta_gastos, '519595'),
		COALESCE(cuenta_iva_descontable, '240810'),
		COALESCE(cuentas_ingreso_categoria, ''),
		COALESCE(cuentas_egreso_categoria, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_finanzas_configuracion
	WHERE empresa_id = ?
	LIMIT 1`, empresaID)

	var habilitarIngresos int
	var habilitarEgresos int
	var requiereAprobacion int
	err := row.Scan(
		&cfg.ID,
		&cfg.EmpresaID,
		&habilitarIngresos,
		&habilitarEgresos,
		&cfg.Moneda,
		&cfg.CategoriasIngreso,
		&cfg.CategoriasEgreso,
		&cfg.PrefijoIngreso,
		&cfg.PrefijoEgreso,
		&cfg.FormatoImpresion,
		&requiereAprobacion,
		&cfg.IntegracionContableDestino,
		&cfg.CuentaCajaBancos,
		&cfg.CuentaIngresos,
		&cfg.CuentaIVAGenerado,
		&cfg.CuentaGastos,
		&cfg.CuentaIVADescontable,
		&cfg.CuentasIngresoCategoria,
		&cfg.CuentasEgresoCategoria,
		&cfg.FechaCreacion,
		&cfg.FechaActualizacion,
		&cfg.UsuarioCreador,
		&cfg.Estado,
		&cfg.Observaciones,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return cfg, nil
		}
		return nil, err
	}
	cfg.HabilitarIngresos = habilitarIngresos == 1
	cfg.HabilitarEgresos = habilitarEgresos == 1
	cfg.RequiereAprobacion = requiereAprobacion == 1
	return cfg, nil
}

// UpsertEmpresaFinanzasConfiguracion crea o actualiza la configuracion financiera por empresa.
func UpsertEmpresaFinanzasConfiguracion(dbConn *sql.DB, cfg EmpresaFinanzasConfiguracion) (int64, error) {
	cfg = normalizeEmpresaFinanzasConfiguracion(cfg)

	res, err := dbConn.Exec(`INSERT INTO empresa_finanzas_configuracion (
		empresa_id, habilitar_ingresos, habilitar_egresos, moneda,
		categorias_ingreso, categorias_egreso,
		prefijo_ingreso, prefijo_egreso,
		formato_impresion, requiere_aprobacion,
		integracion_contable_destino,
		cuenta_caja_bancos, cuenta_ingresos, cuenta_iva_generado,
		cuenta_gastos, cuenta_iva_descontable,
		cuentas_ingreso_categoria, cuentas_egreso_categoria,
		usuario_creador, estado, observaciones,
		fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))
	ON CONFLICT(empresa_id) DO UPDATE SET
		habilitar_ingresos = excluded.habilitar_ingresos,
		habilitar_egresos = excluded.habilitar_egresos,
		moneda = excluded.moneda,
		categorias_ingreso = excluded.categorias_ingreso,
		categorias_egreso = excluded.categorias_egreso,
		prefijo_ingreso = excluded.prefijo_ingreso,
		prefijo_egreso = excluded.prefijo_egreso,
		formato_impresion = excluded.formato_impresion,
		requiere_aprobacion = excluded.requiere_aprobacion,
		integracion_contable_destino = excluded.integracion_contable_destino,
		cuenta_caja_bancos = excluded.cuenta_caja_bancos,
		cuenta_ingresos = excluded.cuenta_ingresos,
		cuenta_iva_generado = excluded.cuenta_iva_generado,
		cuenta_gastos = excluded.cuenta_gastos,
		cuenta_iva_descontable = excluded.cuenta_iva_descontable,
		cuentas_ingreso_categoria = excluded.cuentas_ingreso_categoria,
		cuentas_egreso_categoria = excluded.cuentas_egreso_categoria,
		usuario_creador = excluded.usuario_creador,
		estado = excluded.estado,
		observaciones = excluded.observaciones,
		fecha_actualizacion = datetime('now','localtime')`,
		cfg.EmpresaID,
		boolToInt(cfg.HabilitarIngresos),
		boolToInt(cfg.HabilitarEgresos),
		cfg.Moneda,
		cfg.CategoriasIngreso,
		cfg.CategoriasEgreso,
		cfg.PrefijoIngreso,
		cfg.PrefijoEgreso,
		cfg.FormatoImpresion,
		boolToInt(cfg.RequiereAprobacion),
		cfg.IntegracionContableDestino,
		cfg.CuentaCajaBancos,
		cfg.CuentaIngresos,
		cfg.CuentaIVAGenerado,
		cfg.CuentaGastos,
		cfg.CuentaIVADescontable,
		cfg.CuentasIngresoCategoria,
		cfg.CuentasEgresoCategoria,
		cfg.UsuarioCreador,
		cfg.Estado,
		cfg.Observaciones,
	)
	if err != nil {
		return 0, err
	}
	if id, errID := res.LastInsertId(); errID == nil && id > 0 {
		return id, nil
	}
	current, err := GetEmpresaFinanzasConfiguracion(dbConn, cfg.EmpresaID)
	if err != nil {
		return 0, err
	}
	return current.ID, nil
}

// CreateEmpresaFinanzasMovimiento crea un movimiento financiero por empresa.
func CreateEmpresaFinanzasMovimiento(dbConn *sql.DB, m EmpresaFinanzasMovimiento) (int64, error) {
	m, err := normalizeEmpresaFinanzasMovimiento(dbConn, m, true)
	if err != nil {
		return 0, err
	}

	res, err := dbConn.Exec(`INSERT INTO empresa_finanzas_movimientos (
		empresa_id, tipo_movimiento, codigo, fecha_movimiento,
		categoria, subcategoria, concepto, descripcion,
		metodo_pago, moneda, monto, impuesto, total,
		tercero_nombre, tercero_documento,
		tipo_comprobante, numero_comprobante, comprobante_url,
		referencia_externa, aprobado_por,
		usuario_creador, estado, observaciones,
		fecha_creacion, fecha_actualizacion
	) VALUES (
		?, ?, ?, ?,
		?, ?, ?, ?,
		?, ?, ?, ?, ?,
		?, ?,
		?, ?, ?,
		?, ?,
		?, ?, ?,
		datetime('now','localtime'), datetime('now','localtime')
	)`,
		m.EmpresaID, m.TipoMovimiento, m.Codigo, m.FechaMovimiento,
		m.Categoria, m.Subcategoria, m.Concepto, m.Descripcion,
		m.MetodoPago, m.Moneda, m.Monto, m.Impuesto, m.Total,
		m.TerceroNombre, m.TerceroDocumento,
		m.TipoComprobante, m.NumeroComprobante, m.ComprobanteURL,
		m.ReferenciaExterna, m.AprobadoPor,
		m.UsuarioCreador, m.Estado, m.Observaciones,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListEmpresaFinanzasMovimientos lista movimientos financieros por empresa con filtros.
func ListEmpresaFinanzasMovimientos(dbConn *sql.DB, empresaID int64, f EmpresaFinanzasMovimientoFilter) ([]EmpresaFinanzasMovimiento, error) {
	query := `SELECT
		id, empresa_id, COALESCE(tipo_movimiento, 'egreso'), COALESCE(codigo, ''),
		COALESCE(fecha_movimiento, ''), COALESCE(categoria, ''), COALESCE(subcategoria, ''),
		COALESCE(concepto, ''), COALESCE(descripcion, ''), COALESCE(metodo_pago, ''),
		COALESCE(moneda, 'COP'), COALESCE(monto, 0), COALESCE(impuesto, 0), COALESCE(total, 0),
		COALESCE(tercero_nombre, ''), COALESCE(tercero_documento, ''),
		COALESCE(tipo_comprobante, 'recibo_interno'), COALESCE(numero_comprobante, ''), COALESCE(comprobante_url, ''),
		COALESCE(referencia_externa, ''), COALESCE(aprobado_por, ''),
		COALESCE(fecha_creacion, ''), COALESCE(fecha_actualizacion, ''), COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'), COALESCE(observaciones, '')
	FROM empresa_finanzas_movimientos
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	f.Tipo = normalizeTipoMovimiento(f.Tipo)
	if f.Tipo != "" {
		query += ` AND tipo_movimiento = ?`
		args = append(args, f.Tipo)
	}
	if !f.IncludeInactive {
		query += ` AND COALESCE(estado, 'activo') = 'activo'`
	}
	if strings.TrimSpace(f.Desde) != "" {
		query += ` AND date(fecha_movimiento) >= date(?)`
		args = append(args, strings.TrimSpace(f.Desde))
	}
	if strings.TrimSpace(f.Hasta) != "" {
		query += ` AND date(fecha_movimiento) <= date(?)`
		args = append(args, strings.TrimSpace(f.Hasta))
	}
	if q := strings.TrimSpace(strings.ToLower(f.Q)); q != "" {
		like := "%" + q + "%"
		query += ` AND (
			LOWER(COALESCE(codigo, '')) LIKE ? OR
			LOWER(COALESCE(concepto, '')) LIKE ? OR
			LOWER(COALESCE(descripcion, '')) LIKE ? OR
			LOWER(COALESCE(numero_comprobante, '')) LIKE ? OR
			LOWER(COALESCE(tercero_nombre, '')) LIKE ?
		)`
		args = append(args, like, like, like, like, like)
	}

	query += ` ORDER BY datetime(fecha_movimiento) DESC, id DESC`
	limit := f.Limit
	if limit <= 0 {
		limit = 200
	}
	query += ` LIMIT ?`
	args = append(args, limit)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaFinanzasMovimiento, 0)
	for rows.Next() {
		var m EmpresaFinanzasMovimiento
		if err := rows.Scan(
			&m.ID,
			&m.EmpresaID,
			&m.TipoMovimiento,
			&m.Codigo,
			&m.FechaMovimiento,
			&m.Categoria,
			&m.Subcategoria,
			&m.Concepto,
			&m.Descripcion,
			&m.MetodoPago,
			&m.Moneda,
			&m.Monto,
			&m.Impuesto,
			&m.Total,
			&m.TerceroNombre,
			&m.TerceroDocumento,
			&m.TipoComprobante,
			&m.NumeroComprobante,
			&m.ComprobanteURL,
			&m.ReferenciaExterna,
			&m.AprobadoPor,
			&m.FechaCreacion,
			&m.FechaActualizacion,
			&m.UsuarioCreador,
			&m.Estado,
			&m.Observaciones,
		); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// UpdateEmpresaFinanzasMovimiento actualiza un movimiento financiero por empresa.
func UpdateEmpresaFinanzasMovimiento(dbConn *sql.DB, m EmpresaFinanzasMovimiento) error {
	m, err := normalizeEmpresaFinanzasMovimiento(dbConn, m, false)
	if err != nil {
		return err
	}
	res, err := dbConn.Exec(`UPDATE empresa_finanzas_movimientos SET
		tipo_movimiento = ?,
		codigo = ?,
		fecha_movimiento = ?,
		categoria = ?,
		subcategoria = ?,
		concepto = ?,
		descripcion = ?,
		metodo_pago = ?,
		moneda = ?,
		monto = ?,
		impuesto = ?,
		total = ?,
		tercero_nombre = ?,
		tercero_documento = ?,
		tipo_comprobante = ?,
		numero_comprobante = ?,
		comprobante_url = ?,
		referencia_externa = ?,
		aprobado_por = ?,
		observaciones = ?,
		estado = ?,
		fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`,
		m.TipoMovimiento,
		m.Codigo,
		m.FechaMovimiento,
		m.Categoria,
		m.Subcategoria,
		m.Concepto,
		m.Descripcion,
		m.MetodoPago,
		m.Moneda,
		m.Monto,
		m.Impuesto,
		m.Total,
		m.TerceroNombre,
		m.TerceroDocumento,
		m.TipoComprobante,
		m.NumeroComprobante,
		m.ComprobanteURL,
		m.ReferenciaExterna,
		m.AprobadoPor,
		m.Observaciones,
		m.Estado,
		m.EmpresaID,
		m.ID,
	)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SetEmpresaFinanzasMovimientoEstado activa/desactiva/anula un movimiento financiero.
func SetEmpresaFinanzasMovimientoEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	estado = normalizeEstadoMovimiento(estado)
	res, err := dbConn.Exec(`UPDATE empresa_finanzas_movimientos
	SET estado = ?, fecha_actualizacion = datetime('now','localtime')
	WHERE empresa_id = ? AND id = ?`, estado, empresaID, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteEmpresaFinanzasMovimiento elimina un movimiento financiero por empresa.
func DeleteEmpresaFinanzasMovimiento(dbConn *sql.DB, empresaID, id int64) error {
	res, err := dbConn.Exec(`DELETE FROM empresa_finanzas_movimientos WHERE empresa_id = ? AND id = ?`, empresaID, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func normalizeEmpresaFinanzasMovimiento(dbConn *sql.DB, m EmpresaFinanzasMovimiento, isCreate bool) (EmpresaFinanzasMovimiento, error) {
	if m.EmpresaID <= 0 {
		return m, fmt.Errorf("empresa_id es obligatorio")
	}
	m.TipoMovimiento = normalizeTipoMovimiento(m.TipoMovimiento)
	if m.TipoMovimiento == "" {
		return m, fmt.Errorf("tipo_movimiento debe ser ingreso o egreso")
	}
	if m.Monto <= 0 {
		return m, fmt.Errorf("monto debe ser mayor que cero")
	}
	m.Impuesto = maxFloat64(m.Impuesto, 0)
	m.Total = maxFloat64(m.Total, 0)
	if m.Total <= 0 {
		m.Total = m.Monto + m.Impuesto
	}
	m.Moneda = strings.ToUpper(strings.TrimSpace(m.Moneda))
	if m.Moneda == "" {
		m.Moneda = "COP"
	}
	m.Codigo = sanitizeFinancialCode(m.Codigo)
	if m.Codigo == "" {
		cfg, err := GetEmpresaFinanzasConfiguracion(dbConn, m.EmpresaID)
		if err != nil {
			return m, err
		}
		m.Codigo = buildDefaultFinancialCode(m.TipoMovimiento, cfg.PrefijoIngreso, cfg.PrefijoEgreso)
	}
	m.FechaMovimiento = strings.TrimSpace(m.FechaMovimiento)
	if m.FechaMovimiento == "" {
		m.FechaMovimiento = time.Now().Format("2006-01-02 15:04:05")
	}
	m.Categoria = strings.TrimSpace(m.Categoria)
	m.Subcategoria = strings.TrimSpace(m.Subcategoria)
	m.Concepto = strings.TrimSpace(m.Concepto)
	if m.Concepto == "" {
		return m, fmt.Errorf("concepto es obligatorio")
	}
	m.Descripcion = strings.TrimSpace(m.Descripcion)
	m.MetodoPago = strings.TrimSpace(m.MetodoPago)
	if m.MetodoPago == "" {
		m.MetodoPago = "efectivo"
	}
	m.TerceroNombre = strings.TrimSpace(m.TerceroNombre)
	m.TerceroDocumento = strings.TrimSpace(m.TerceroDocumento)
	m.TipoComprobante = strings.TrimSpace(m.TipoComprobante)
	if m.TipoComprobante == "" {
		m.TipoComprobante = "recibo_interno"
	}
	m.NumeroComprobante = strings.TrimSpace(m.NumeroComprobante)
	if m.NumeroComprobante == "" {
		m.NumeroComprobante = m.Codigo
	}
	m.ComprobanteURL = strings.TrimSpace(m.ComprobanteURL)
	m.ReferenciaExterna = strings.TrimSpace(m.ReferenciaExterna)
	m.AprobadoPor = strings.TrimSpace(m.AprobadoPor)
	m.UsuarioCreador = strings.TrimSpace(m.UsuarioCreador)
	if m.UsuarioCreador == "" {
		m.UsuarioCreador = "sistema"
	}
	m.Estado = normalizeEstadoMovimiento(m.Estado)
	m.Observaciones = strings.TrimSpace(m.Observaciones)
	if !isCreate && m.ID <= 0 {
		return m, fmt.Errorf("id es obligatorio")
	}
	return m, nil
}

func normalizeEmpresaFinanzasConfiguracion(cfg EmpresaFinanzasConfiguracion) EmpresaFinanzasConfiguracion {
	cfg.HabilitarIngresos = cfg.HabilitarIngresos || !cfg.HabilitarEgresos
	cfg.HabilitarEgresos = cfg.HabilitarEgresos || !cfg.HabilitarIngresos
	cfg.Moneda = strings.ToUpper(strings.TrimSpace(cfg.Moneda))
	if cfg.Moneda == "" {
		cfg.Moneda = "COP"
	}
	cfg.CategoriasIngreso = strings.TrimSpace(cfg.CategoriasIngreso)
	if cfg.CategoriasIngreso == "" {
		cfg.CategoriasIngreso = "ventas\nservicios\notros ingresos"
	}
	cfg.CategoriasEgreso = strings.TrimSpace(cfg.CategoriasEgreso)
	if cfg.CategoriasEgreso == "" {
		cfg.CategoriasEgreso = "compras\nnomina\nservicios\narriendo\notros gastos"
	}
	cfg.PrefijoIngreso = sanitizeFinancialCode(cfg.PrefijoIngreso)
	if cfg.PrefijoIngreso == "" {
		cfg.PrefijoIngreso = "ING"
	}
	cfg.PrefijoEgreso = sanitizeFinancialCode(cfg.PrefijoEgreso)
	if cfg.PrefijoEgreso == "" {
		cfg.PrefijoEgreso = "EGR"
	}
	cfg.FormatoImpresion = strings.ToLower(strings.TrimSpace(cfg.FormatoImpresion))
	if cfg.FormatoImpresion != "pos" {
		cfg.FormatoImpresion = "carta"
	}
	cfg.IntegracionContableDestino = normalizeIntegracionContableDestino(cfg.IntegracionContableDestino)
	cfg.CuentaCajaBancos = sanitizeContableAccount(cfg.CuentaCajaBancos)
	if cfg.CuentaCajaBancos == "" {
		cfg.CuentaCajaBancos = "110505"
	}
	cfg.CuentaIngresos = sanitizeContableAccount(cfg.CuentaIngresos)
	if cfg.CuentaIngresos == "" {
		cfg.CuentaIngresos = "413595"
	}
	cfg.CuentaIVAGenerado = sanitizeContableAccount(cfg.CuentaIVAGenerado)
	if cfg.CuentaIVAGenerado == "" {
		cfg.CuentaIVAGenerado = "240805"
	}
	cfg.CuentaGastos = sanitizeContableAccount(cfg.CuentaGastos)
	if cfg.CuentaGastos == "" {
		cfg.CuentaGastos = "519595"
	}
	cfg.CuentaIVADescontable = sanitizeContableAccount(cfg.CuentaIVADescontable)
	if cfg.CuentaIVADescontable == "" {
		cfg.CuentaIVADescontable = "240810"
	}
	cfg.CuentasIngresoCategoria = normalizeCuentaCategoriasMapping(cfg.CuentasIngresoCategoria)
	cfg.CuentasEgresoCategoria = normalizeCuentaCategoriasMapping(cfg.CuentasEgresoCategoria)
	cfg.UsuarioCreador = strings.TrimSpace(cfg.UsuarioCreador)
	if cfg.UsuarioCreador == "" {
		cfg.UsuarioCreador = "sistema"
	}
	cfg.Estado = normalizeEstadoMovimiento(cfg.Estado)
	cfg.Observaciones = strings.TrimSpace(cfg.Observaciones)
	return cfg
}

func defaultEmpresaFinanzasConfiguracion(empresaID int64) *EmpresaFinanzasConfiguracion {
	return &EmpresaFinanzasConfiguracion{
		EmpresaID:                  empresaID,
		HabilitarIngresos:          true,
		HabilitarEgresos:           true,
		Moneda:                     "COP",
		CategoriasIngreso:          "ventas\nservicios\notros ingresos",
		CategoriasEgreso:           "compras\nnomina\nservicios\narriendo\notros gastos",
		PrefijoIngreso:             "ING",
		PrefijoEgreso:              "EGR",
		FormatoImpresion:           "carta",
		RequiereAprobacion:         false,
		IntegracionContableDestino: "generico",
		CuentaCajaBancos:           "110505",
		CuentaIngresos:             "413595",
		CuentaIVAGenerado:          "240805",
		CuentaGastos:               "519595",
		CuentaIVADescontable:       "240810",
		CuentasIngresoCategoria:    "ventas=413595\nservicios=417595\notros ingresos=429595",
		CuentasEgresoCategoria:     "compras=613595\nnomina=510506\nservicios=513595\narriendo=512001\notros gastos=519595",
		Estado:                     "activo",
	}
}

func normalizeIntegracionContableDestino(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "siigo", "world_office", "alegra":
		return v
	default:
		return "generico"
	}
}

func normalizeTipoMovimiento(tipo string) string {
	t := strings.ToLower(strings.TrimSpace(tipo))
	if t == "ingreso" || t == "egreso" {
		return t
	}
	return ""
}

func normalizeEstadoMovimiento(estado string) string {
	e := strings.ToLower(strings.TrimSpace(estado))
	if e == "inactivo" || e == "anulado" {
		return e
	}
	return "activo"
}

func sanitizeFinancialCode(v string) string {
	v = strings.ToUpper(strings.TrimSpace(v))
	if v == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range v {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_' || r == '/':
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

func sanitizeContableAccount(v string) string {
	v = strings.ToUpper(strings.TrimSpace(v))
	if v == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range v {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_' || r == '/' || r == '.':
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

func normalizeCuentaCategoriasMapping(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	lines := strings.Split(raw, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		l := strings.TrimSpace(line)
		if l == "" {
			continue
		}
		idx := strings.IndexAny(l, "=:")
		if idx <= 0 {
			continue
		}
		categoria := strings.ToLower(strings.TrimSpace(l[:idx]))
		cuenta := sanitizeContableAccount(strings.TrimSpace(l[idx+1:]))
		if categoria == "" || cuenta == "" {
			continue
		}
		out = append(out, categoria+"="+cuenta)
	}
	return strings.Join(out, "\n")
}

func buildDefaultFinancialCode(tipo, prefIngreso, prefEgreso string) string {
	prefix := sanitizeFinancialCode(prefEgreso)
	if tipo == "ingreso" {
		prefix = sanitizeFinancialCode(prefIngreso)
	}
	if prefix == "" {
		if tipo == "ingreso" {
			prefix = "ING"
		} else {
			prefix = "EGR"
		}
	}
	now := time.Now()
	suffix := fmt.Sprintf("%s%03d", now.Format("20060102150405"), now.UnixNano()%1000)
	return prefix + "-" + suffix
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func maxFloat64(v, min float64) float64 {
	if v < min {
		return min
	}
	return v
}
