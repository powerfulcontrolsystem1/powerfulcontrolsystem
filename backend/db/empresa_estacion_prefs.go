package db

import (
	"encoding/json"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// EmpresaEstacionPref representa una preferencia por estación y empresa.
type EmpresaEstacionPref struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	EstacionID         int64  `json:"estacion_id"`
	Clave              string `json:"clave"`
	Valor              string `json:"valor"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// EmpresaEstacionCarritosSyncResult resume la sincronizacion de carritos por estacion.
type EmpresaEstacionCarritosSyncResult struct {
	Processed int `json:"processed"`
	Created   int `json:"created"`
	Updated   int `json:"updated"`
	Ignored   int `json:"ignored"`
}

type empresaEstacionesConfig struct {
	Cantidad   int                           `json:"cantidad"`
	Estaciones []empresaEstacionConfigRecord `json:"estaciones"`
}

type empresaEstacionConfigRecord struct {
	ID     int64  `json:"id"`
	Nombre string `json:"nombre"`
}

// EnsureEmpresaEstacionPrefsSchema crea/migra la tabla de preferencias por estacion.
func EnsureEmpresaEstacionPrefsSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_estacion_prefs (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            empresa_id INTEGER NOT NULL,
            estacion_id INTEGER NOT NULL,
            clave TEXT NOT NULL,
            valor TEXT,
            fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
            fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
            usuario_creador TEXT,
            estado TEXT DEFAULT 'activo',
            observaciones TEXT
        );`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_estacion_prefs_empresa_estacion_clave ON empresa_estacion_prefs(empresa_id, estacion_id, clave);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_estacion_prefs_empresa_estacion ON empresa_estacion_prefs(empresa_id, estacion_id);`,
	}

	for _, s := range stmts {
		if _, err := dbConn.Exec(s); err != nil {
			return err
		}
	}

	// ensure common columns in case of incremental migrations
	if err := ensureColumnIfMissing(dbConn, "empresa_estacion_prefs", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_estacion_prefs", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_estacion_prefs", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_estacion_prefs", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

// ListEmpresaEstacionPrefs lista preferencias por empresa y opcionalmente por estacion.
func ListEmpresaEstacionPrefs(dbConn *sql.DB, empresaID int64, estacionID int64, includeInactive bool) ([]EmpresaEstacionPref, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, errors.New("empresa_id invalido")
	}

	where := "WHERE empresa_id = ?"
	args := []interface{}{empresaID}
	if estacionID > 0 {
		where += " AND estacion_id = ?"
		args = append(args, estacionID)
	}
	if !includeInactive {
		where += " AND LOWER(COALESCE(NULLIF(TRIM(estado), ''), 'activo')) = 'activo'"
	}

	q := "SELECT id, empresa_id, estacion_id, COALESCE(clave, ''), COALESCE(valor, ''), COALESCE(fecha_creacion, ''), COALESCE(fecha_actualizacion, ''), COALESCE(usuario_creador, ''), COALESCE(NULLIF(TRIM(estado), ''), 'activo'), COALESCE(observaciones, '') FROM empresa_estacion_prefs " + where + " ORDER BY estacion_id, clave"

	rows, err := dbConn.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaEstacionPref, 0)
	for rows.Next() {
		var p EmpresaEstacionPref
		if err := rows.Scan(&p.ID, &p.EmpresaID, &p.EstacionID, &p.Clave, &p.Valor, &p.FechaCreacion, &p.FechaActualizacion, &p.UsuarioCreador, &p.Estado, &p.Observaciones); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

// GetEmpresaEstacionPref obtiene una preferencia por clave (si existe).
func GetEmpresaEstacionPref(dbConn *sql.DB, empresaID int64, estacionID int64, clave string) (*EmpresaEstacionPref, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, errors.New("empresa_id invalido")
	}
	clave = strings.TrimSpace(clave)
	if clave == "" {
		return nil, errors.New("clave es obligatoria")
	}

	row := dbConn.QueryRow(`SELECT id, empresa_id, estacion_id, COALESCE(clave,''), COALESCE(valor,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), COALESCE(NULLIF(TRIM(estado), ''),'activo'), COALESCE(observaciones,'') FROM empresa_estacion_prefs WHERE empresa_id = ? AND estacion_id = ? AND clave = ? LIMIT 1`, empresaID, estacionID, clave)
	var p EmpresaEstacionPref
	if err := row.Scan(&p.ID, &p.EmpresaID, &p.EstacionID, &p.Clave, &p.Valor, &p.FechaCreacion, &p.FechaActualizacion, &p.UsuarioCreador, &p.Estado, &p.Observaciones); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// UpsertEmpresaEstacionPref crea o actualiza una preferencia por clave (usa ON CONFLICT para SQLite).
func UpsertEmpresaEstacionPref(dbConn *sql.DB, p EmpresaEstacionPref) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	if p.EmpresaID <= 0 {
		return 0, errors.New("empresa_id invalido")
	}
	if p.EstacionID < 0 {
		p.EstacionID = 0
	}
	p.Clave = strings.TrimSpace(p.Clave)
	if p.Clave == "" {
		return 0, errors.New("clave es obligatoria")
	}
	p.Estado = strings.ToLower(strings.TrimSpace(p.Estado))
	if p.Estado == "" {
		p.Estado = "activo"
	}

	// Use upsert via ON CONFLICT on unique index
	res, err := dbConn.Exec(`INSERT INTO empresa_estacion_prefs (
        empresa_id, estacion_id, clave, valor, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
    ) VALUES (
        ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'), ?, ?, ?
    ) ON CONFLICT(empresa_id, estacion_id, clave) DO UPDATE SET
        valor = excluded.valor,
        fecha_actualizacion = datetime('now','localtime'),
        usuario_creador = CASE WHEN trim(excluded.usuario_creador) <> '' THEN excluded.usuario_creador ELSE empresa_estacion_prefs.usuario_creador END,
		estado = COALESCE(NULLIF(TRIM(excluded.estado), ''), 'activo'),
        observaciones = excluded.observaciones`,
		p.EmpresaID, p.EstacionID, p.Clave, strings.TrimSpace(p.Valor), strings.TrimSpace(p.UsuarioCreador), p.Estado, strings.TrimSpace(p.Observaciones))
	if err != nil {
		return 0, fmt.Errorf("upsert error: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil && id > 0 {
		return id, nil
	}
	// If no new insert id, try to fetch existing id
	var existing int64
	if err := dbConn.QueryRow(`SELECT id FROM empresa_estacion_prefs WHERE empresa_id = ? AND estacion_id = ? AND clave = ? LIMIT 1`, p.EmpresaID, p.EstacionID, p.Clave).Scan(&existing); err != nil {
		return 0, err
	}
	return existing, nil
}

func parseEmpresaEstacionesConfig(raw string) (*empresaEstacionesConfig, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	var current interface{} = trimmed
	for i := 0; i < 3; i++ {
		asString, ok := current.(string)
		if !ok {
			break
		}
		asString = strings.TrimSpace(asString)
		if asString == "" {
			return nil, nil
		}
		var decoded interface{}
		if err := json.Unmarshal([]byte(asString), &decoded); err != nil {
			return nil, fmt.Errorf("configuracion de estaciones invalida: %w", err)
		}
		current = decoded
	}

	blob, err := json.Marshal(current)
	if err != nil {
		return nil, fmt.Errorf("no se pudo normalizar configuracion de estaciones: %w", err)
	}

	var cfg empresaEstacionesConfig
	if err := json.Unmarshal(blob, &cfg); err != nil {
		return nil, fmt.Errorf("no se pudo interpretar configuracion de estaciones: %w", err)
	}
	return &cfg, nil
}

func normalizeEmpresaEstacionCarritoLookup(v string) string {
	return strings.TrimSpace(strings.ToLower(v))
}

func normalizeEmpresaEstacionNombre(estacionID int64, nombre string) string {
	nombre = strings.TrimSpace(nombre)
	if nombre != "" {
		return nombre
	}
	if estacionID > 0 {
		return fmt.Sprintf("Estacion %d", estacionID)
	}
	return "Estacion"
}

func ensureEmpresaEstacionCarritoDefaultState(dbConn *sql.DB, empresaID, carritoID int64) (int, error) {
	carrito, err := GetCarritoCompraByID(dbConn, empresaID, carritoID)
	if err != nil {
		return 0, err
	}

	updated := 0
	if strings.TrimSpace(strings.ToLower(carrito.EstadoCarrito)) != "cerrado" {
		if err := SetCarritoOperacionEstado(dbConn, empresaID, carritoID, "cerrado"); err != nil {
			return updated, err
		}
		updated += 1
	}
	if strings.TrimSpace(strings.ToLower(carrito.Estado)) != "inactivo" {
		if err := SetCarritoCompraEstado(dbConn, empresaID, carritoID, "inactivo"); err != nil {
			return updated, err
		}
		updated += 1
	}

	return updated, nil
}

func listEmpresaEstacionCarritosForSync(dbConn *sql.DB, empresaID int64) ([]CarritoCompra, error) {
	rows, err := dbConn.Query(`SELECT
		id,
		empresa_id,
		COALESCE(codigo, ''),
		COALESCE(nombre, ''),
		COALESCE(canal_venta, 'mostrador'),
		COALESCE(cliente_id, 0),
		COALESCE(estado_carrito, 'abierto'),
		COALESCE(moneda, 'COP'),
		COALESCE(referencia_externa, ''),
		COALESCE(subtotal, 0),
		COALESCE(descuento_total, 0),
		COALESCE(impuesto_total, 0),
		COALESCE(total, 0),
		COALESCE(activado_en, ''),
		COALESCE(pagado_en, ''),
		COALESCE(descuento_tipo, ''),
		COALESCE(descuento_codigo, ''),
		COALESCE(descuento_valor, 0),
		COALESCE(devolucion_total, 0),
		COALESCE(total_pagado, 0),
		COALESCE(metodo_pago, 'efectivo'),
		COALESCE(referencia_pago, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM carritos_compras
	WHERE empresa_id = ?
	ORDER BY id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]CarritoCompra, 0)
	for rows.Next() {
		var item CarritoCompra
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.Codigo,
			&item.Nombre,
			&item.CanalVenta,
			&item.ClienteID,
			&item.EstadoCarrito,
			&item.Moneda,
			&item.ReferenciaExterna,
			&item.Subtotal,
			&item.DescuentoTotal,
			&item.ImpuestoTotal,
			&item.Total,
			&item.ActivadoEn,
			&item.PagadoEn,
			&item.DescuentoTipo,
			&item.DescuentoCodigo,
			&item.DescuentoValor,
			&item.DevolucionTotal,
			&item.TotalPagado,
			&item.MetodoPago,
			&item.ReferenciaPago,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// SyncEmpresaEstacionCarritos asegura que cada estacion configurada tenga un carrito enlazado por defecto.
func SyncEmpresaEstacionCarritos(dbConn *sql.DB, empresaID int64, rawConfig string, usuario string) (*EmpresaEstacionCarritosSyncResult, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, errors.New("empresa_id invalido")
	}

	cfg, err := parseEmpresaEstacionesConfig(rawConfig)
	if err != nil {
		return nil, err
	}
	result := &EmpresaEstacionCarritosSyncResult{}
	if cfg == nil || len(cfg.Estaciones) == 0 {
		return result, nil
	}

	if err := EnsureEmpresaCarritosSchema(dbConn); err != nil {
		return nil, err
	}

	existing, err := listEmpresaEstacionCarritosForSync(dbConn, empresaID)
	if err != nil {
		return nil, err
	}

	byCode := make(map[string]CarritoCompra)
	byReference := make(map[string]CarritoCompra)
	byName := make(map[string]CarritoCompra)
	indexCarrito := func(item CarritoCompra) {
		if key := normalizeEmpresaEstacionCarritoLookup(item.Codigo); key != "" {
			byCode[key] = item
		}
		if key := normalizeEmpresaEstacionCarritoLookup(item.ReferenciaExterna); key != "" {
			byReference[key] = item
		}
		if key := normalizeEmpresaEstacionCarritoLookup(item.Nombre); key != "" {
			byName[key] = item
		}
	}
	for _, item := range existing {
		indexCarrito(item)
	}

	resolveCarrito := func(code, reference, name string) *CarritoCompra {
		if item, ok := byCode[normalizeEmpresaEstacionCarritoLookup(code)]; ok {
			copied := item
			return &copied
		}
		if item, ok := byReference[normalizeEmpresaEstacionCarritoLookup(reference)]; ok {
			copied := item
			return &copied
		}
		if item, ok := byName[normalizeEmpresaEstacionCarritoLookup(name)]; ok {
			copied := item
			return &copied
		}
		return nil
	}

	refreshIndexes := func() error {
		rows, err := listEmpresaEstacionCarritosForSync(dbConn, empresaID)
		if err != nil {
			return err
		}
		byCode = make(map[string]CarritoCompra)
		byReference = make(map[string]CarritoCompra)
		byName = make(map[string]CarritoCompra)
		for _, item := range rows {
			indexCarrito(item)
		}
		return nil
	}

	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = "sistema"
	}

	for _, station := range cfg.Estaciones {
		if station.ID <= 0 {
			result.Ignored += 1
			continue
		}
		result.Processed += 1

		stationName := normalizeEmpresaEstacionNombre(station.ID, station.Nombre)
		code := fmt.Sprintf("EST-%d-%d", empresaID, station.ID)
		reference := fmt.Sprintf("ESTACION_%d", station.ID)

		current := resolveCarrito(code, reference, stationName)
		if current == nil {
			createdID, createErr := CreateCarritoCompra(dbConn, CarritoCompra{
				EmpresaID:         empresaID,
				Codigo:            code,
				Nombre:            stationName,
				CanalVenta:        "mostrador",
				EstadoCarrito:     "cerrado",
				ReferenciaExterna: reference,
				UsuarioCreador:    usuario,
				Observaciones:     fmt.Sprintf("Carrito generado automaticamente para estacion #%d", station.ID),
			})
			if createErr != nil {
				if !isUniqueConstraintError(createErr) {
					return nil, createErr
				}
				if err := refreshIndexes(); err != nil {
					return nil, err
				}
				current = resolveCarrito(code, reference, stationName)
				if current == nil {
					return nil, createErr
				}
			} else {
				created, err := GetCarritoCompraByID(dbConn, empresaID, createdID)
				if err != nil {
					return nil, err
				}
				current = created
				result.Created += 1
			}
		}

		if current == nil {
			result.Ignored += 1
			continue
		}

		updated := 0
		if strings.TrimSpace(current.Nombre) != stationName ||
			normalizeEmpresaEstacionCarritoLookup(current.Codigo) != normalizeEmpresaEstacionCarritoLookup(code) ||
			normalizeEmpresaEstacionCarritoLookup(current.ReferenciaExterna) != normalizeEmpresaEstacionCarritoLookup(reference) {
			payload := *current
			payload.Codigo = code
			payload.Nombre = stationName
			payload.ReferenciaExterna = reference
			if err := UpdateCarritoCompra(dbConn, payload); err != nil {
				return nil, err
			}
			updated += 1
		}

		stateUpdates, err := ensureEmpresaEstacionCarritoDefaultState(dbConn, empresaID, current.ID)
		if err != nil {
			return nil, err
		}
		updated += stateUpdates
		if updated == 0 {
			result.Ignored += 1
		} else {
			result.Updated += updated
		}

		if latest, err := GetCarritoCompraByID(dbConn, empresaID, current.ID); err == nil {
			indexCarrito(*latest)
		}
	}

	return result, nil
}
