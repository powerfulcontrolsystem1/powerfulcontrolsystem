package db

import (
	"database/sql"
	"encoding/json"
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
            id BIGSERIAL PRIMARY KEY,
            empresa_id BIGINT NOT NULL,
            estacion_id BIGINT NOT NULL,
            clave TEXT NOT NULL,
            valor TEXT,
            fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            fecha_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            usuario_creador TEXT,
            estado TEXT DEFAULT 'activo',
            observaciones TEXT
        );`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_estacion_prefs_empresa_estacion_clave ON empresa_estacion_prefs(empresa_id, estacion_id, clave);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_estacion_prefs_empresa_estacion ON empresa_estacion_prefs(empresa_id, estacion_id);`,
	}

	for _, s := range stmts {
		if _, err := execSQLCompat(dbConn, s); err != nil {
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

// DisableLegacyFloatingRobotAndRadioPrefs mantiene el chat IA activo por empresa,
// pero retira los avatares legados y deja la emisora como opcion explicita.
func DisableLegacyFloatingRobotAndRadioPrefs(dbConn *sql.DB) error {
	if dbConn == nil {
		return nil
	}
	if err := EnsureEmpresaEstacionPrefsSchema(dbConn); err != nil {
		return err
	}
	return ApplySchemaMigration(dbConn, "empresas", "20260610_chat_ia_activo_empresas_reaplicar", "Reactiva chat IA flotante y retira robot/secretaria en empresas existentes", func(tx *sql.DB) error {
		_, err := execSQLCompat(tx, `
			INSERT INTO empresa_estacion_prefs (
				empresa_id, estacion_id, clave, valor, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
			)
			SELECT COALESCE(NULLIF(e.empresa_id, 0), e.id),
			       0,
			       pref.clave,
			       pref.valor,
			       CURRENT_TIMESTAMP,
			       CURRENT_TIMESTAMP,
			       'sistema.preproduccion',
			       'activo',
			       '[preproduccion_2026-06-10] chat IA activo para todas las empresas; robot/secretaria retirados; emisora opt-in'
			FROM empresas e
			CROSS JOIN (
				VALUES
					('chat_flotante.chat_enabled', '1'),
					('chat_flotante.robot_enabled', '0'),
					('chat_flotante.personality_mode', 'normal'),
					('chat_flotante.radio_online_enabled', '0')
			) AS pref(clave, valor)
			WHERE COALESCE(NULLIF(e.empresa_id, 0), e.id) > 0
			  AND LOWER(COALESCE(NULLIF(TRIM(e.estado), ''), 'activo')) NOT IN ('inactivo', 'eliminado')
			ON CONFLICT(empresa_id, estacion_id, clave) DO UPDATE SET
				valor = EXCLUDED.valor,
				fecha_actualizacion = CURRENT_TIMESTAMP,
				usuario_creador = 'sistema.preproduccion',
				estado = 'activo',
				observaciones = '[preproduccion_2026-06-10] chat IA activo para todas las empresas; robot/secretaria retirados; emisora opt-in'
		`)
		return err
	})
}

// DisableFloatingChatVoicePrefs deja el chat IA en modo texto por defecto para
// empresas existentes; la voz queda como opcion manual del usuario.
func DisableFloatingChatVoicePrefs(dbConn *sql.DB) error {
	if dbConn == nil {
		return nil
	}
	if err := EnsureEmpresaEstacionPrefsSchema(dbConn); err != nil {
		return err
	}
	return ApplySchemaMigration(dbConn, "empresas", "20260620_chat_ia_voz_apagada_empresas", "Desactiva la voz del chat IA por defecto en empresas existentes", func(tx *sql.DB) error {
		_, err := execSQLCompat(tx, `
			INSERT INTO empresa_estacion_prefs (
				empresa_id, estacion_id, clave, valor, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
			)
			SELECT COALESCE(NULLIF(e.empresa_id, 0), e.id),
			       0,
			       'chat_flotante.voice_enabled',
			       '0',
			       CURRENT_TIMESTAMP,
			       CURRENT_TIMESTAMP,
			       'sistema.preproduccion',
			       'activo',
			       '[preproduccion_2026-06-20] voz del chat IA apagada por defecto; el usuario puede activarla manualmente'
			FROM empresas e
			WHERE COALESCE(NULLIF(e.empresa_id, 0), e.id) > 0
			  AND LOWER(COALESCE(NULLIF(TRIM(e.estado), ''), 'activo')) NOT IN ('inactivo', 'eliminado')
			ON CONFLICT(empresa_id, estacion_id, clave) DO UPDATE SET
				valor = EXCLUDED.valor,
				fecha_actualizacion = CURRENT_TIMESTAMP,
				usuario_creador = 'sistema.preproduccion',
				estado = 'activo',
				observaciones = '[preproduccion_2026-06-20] voz del chat IA apagada por defecto; el usuario puede activarla manualmente'
		`)
		return err
	})
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

	rows, err := querySQLCompat(dbConn, q, args...)
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

	row := queryRowSQLCompat(dbConn, `SELECT id, empresa_id, estacion_id, COALESCE(clave,''), COALESCE(valor,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), COALESCE(NULLIF(TRIM(estado), ''),'activo'), COALESCE(observaciones,'') FROM empresa_estacion_prefs WHERE empresa_id = ? AND estacion_id = ? AND clave = ? LIMIT 1`, empresaID, estacionID, clave)
	var p EmpresaEstacionPref
	if err := row.Scan(&p.ID, &p.EmpresaID, &p.EstacionID, &p.Clave, &p.Valor, &p.FechaCreacion, &p.FechaActualizacion, &p.UsuarioCreador, &p.Estado, &p.Observaciones); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// UpsertEmpresaEstacionPref crea o actualiza una preferencia por clave (usa ON CONFLICT).
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

	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_estacion_prefs (
        empresa_id, estacion_id, clave, valor, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
    ) VALUES (
        ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?, ?
    ) ON CONFLICT(empresa_id, estacion_id, clave) DO UPDATE SET
        valor = excluded.valor,
        fecha_actualizacion = CURRENT_TIMESTAMP,
        usuario_creador = CASE WHEN trim(excluded.usuario_creador) <> '' THEN excluded.usuario_creador ELSE empresa_estacion_prefs.usuario_creador END,
		estado = COALESCE(NULLIF(TRIM(excluded.estado), ''), 'activo'),
        observaciones = excluded.observaciones
	RETURNING id`,
		p.EmpresaID, p.EstacionID, p.Clave, strings.TrimSpace(p.Valor), strings.TrimSpace(p.UsuarioCreador), p.Estado, strings.TrimSpace(p.Observaciones))
	if err != nil {
		return 0, fmt.Errorf("upsert error: %w", err)
	}
	return id, nil
}

// DeleteEmpresaEstacionPrefsByKeys elimina preferencias puntuales de una empresa.
func DeleteEmpresaEstacionPrefsByKeys(dbConn *sql.DB, empresaID int64, estacionID int64, claves []string) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return 0, errors.New("empresa_id invalido")
	}
	if estacionID < 0 {
		estacionID = 0
	}
	clean := make([]string, 0, len(claves))
	for _, clave := range claves {
		clave = strings.TrimSpace(clave)
		if clave != "" {
			clean = append(clean, clave)
		}
	}
	if len(clean) == 0 {
		return 0, nil
	}
	placeholders := make([]string, 0, len(clean))
	args := []interface{}{empresaID, estacionID}
	for _, clave := range clean {
		placeholders = append(placeholders, "?")
		args = append(args, clave)
	}
	res, err := execSQLCompat(dbConn, "DELETE FROM empresa_estacion_prefs WHERE empresa_id = ? AND estacion_id = ? AND clave IN ("+strings.Join(placeholders, ",")+")", args...)
	if err != nil {
		return 0, err
	}
	affected, _ := res.RowsAffected()
	return affected, nil
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
