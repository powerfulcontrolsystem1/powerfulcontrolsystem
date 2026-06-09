package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// EmpresaVentaOfflineSync registra ventas capturadas sin internet y sincronizadas luego.
type EmpresaVentaOfflineSync struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	SyncKey            string `json:"sync_key"`
	CarritoID          int64  `json:"carrito_id,omitempty"`
	DocumentoCodigo    string `json:"documento_codigo,omitempty"`
	PayloadJSON        string `json:"payload_json,omitempty"`
	ResultadoJSON      string `json:"resultado_json,omitempty"`
	EstadoSync         string `json:"estado_sync"`
	ErrorMensaje       string `json:"error_mensaje,omitempty"`
	FechaOffline       string `json:"fecha_offline,omitempty"`
	FechaSync          string `json:"fecha_sync,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
}

// EnsureEmpresaVentasOfflineSchema crea la bitacora idempotente de ventas offline.
func EnsureEmpresaVentasOfflineSchema(dbConn *sql.DB) error {
	startedAt := time.Now()
	defer func() {
		PerfLogf("[perf][schema] EnsureEmpresaVentasOfflineSchema dur=%s", time.Since(startedAt))
	}()
	if dbConn == nil {
		return fmt.Errorf("db connection is nil")
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_ventas_offline_sync (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			sync_key TEXT NOT NULL,
			carrito_id INTEGER DEFAULT 0,
			documento_codigo TEXT,
			payload_json TEXT,
			resultado_json TEXT,
			estado_sync TEXT DEFAULT 'pendiente',
			error_mensaje TEXT,
			fecha_offline TEXT,
			fecha_sync TEXT,
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP)
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_ventas_offline_sync_key ON empresa_ventas_offline_sync(empresa_id, sync_key);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_ventas_offline_estado ON empresa_ventas_offline_sync(empresa_id, estado_sync, fecha_creacion DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_ventas_offline_carrito ON empresa_ventas_offline_sync(empresa_id, carrito_id);`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

// GetEmpresaVentaOfflineSyncByKey busca una sincronizacion por llave idempotente.
func GetEmpresaVentaOfflineSyncByKey(dbConn *sql.DB, empresaID int64, syncKey string) (*EmpresaVentaOfflineSync, error) {
	if err := EnsureEmpresaVentasOfflineSchema(dbConn); err != nil {
		return nil, err
	}
	syncKey = strings.TrimSpace(syncKey)
	if empresaID <= 0 || syncKey == "" {
		return nil, sql.ErrNoRows
	}
	row := QueryRowCompat(dbConn, `SELECT
		id, empresa_id, sync_key, COALESCE(carrito_id,0), COALESCE(documento_codigo,''),
		COALESCE(payload_json,''), COALESCE(resultado_json,''), COALESCE(estado_sync,'pendiente'),
		COALESCE(error_mensaje,''), COALESCE(fecha_offline,''), COALESCE(fecha_sync,''),
		COALESCE(usuario_creador,''), COALESCE(estado,'activo'), COALESCE(observaciones,''),
		COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,'')
	FROM empresa_ventas_offline_sync
	WHERE empresa_id=? AND sync_key=?
	LIMIT 1`, empresaID, syncKey)
	var out EmpresaVentaOfflineSync
	if err := row.Scan(
		&out.ID,
		&out.EmpresaID,
		&out.SyncKey,
		&out.CarritoID,
		&out.DocumentoCodigo,
		&out.PayloadJSON,
		&out.ResultadoJSON,
		&out.EstadoSync,
		&out.ErrorMensaje,
		&out.FechaOffline,
		&out.FechaSync,
		&out.UsuarioCreador,
		&out.Estado,
		&out.Observaciones,
		&out.FechaCreacion,
		&out.FechaActualizacion,
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpsertEmpresaVentaOfflineSyncPending crea o refresca el registro pendiente.
func UpsertEmpresaVentaOfflineSyncPending(dbConn *sql.DB, empresaID int64, syncKey, payloadJSON, fechaOffline, usuario, observaciones string) error {
	if err := EnsureEmpresaVentasOfflineSchema(dbConn); err != nil {
		return err
	}
	if empresaID <= 0 || strings.TrimSpace(syncKey) == "" {
		return fmt.Errorf("empresa_id y sync_key son obligatorios")
	}
	_, err := ExecCompat(dbConn, `INSERT INTO empresa_ventas_offline_sync (
		empresa_id, sync_key, payload_json, estado_sync, fecha_offline, usuario_creador, estado, observaciones,
		fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, 'pendiente', ?, ?, 'activo', ?, `+sqlNowExpr()+`, `+sqlNowExpr()+`)
	ON CONFLICT(empresa_id, sync_key) DO UPDATE SET
		payload_json=excluded.payload_json,
		error_mensaje='',
		fecha_offline=excluded.fecha_offline,
		usuario_creador=excluded.usuario_creador,
		observaciones=excluded.observaciones,
		fecha_actualizacion=`+sqlNowExpr()+`
	WHERE empresa_ventas_offline_sync.estado_sync <> 'sincronizado'`,
		empresaID,
		strings.TrimSpace(syncKey),
		strings.TrimSpace(payloadJSON),
		strings.TrimSpace(fechaOffline),
		strings.TrimSpace(usuario),
		strings.TrimSpace(observaciones),
	)
	return err
}

// MarkEmpresaVentaOfflineSyncResult guarda el resultado final o error de sincronizacion.
func MarkEmpresaVentaOfflineSyncResult(dbConn *sql.DB, empresaID int64, syncKey, estadoSync string, carritoID int64, documentoCodigo, resultadoJSON, errorMensaje string) error {
	if err := EnsureEmpresaVentasOfflineSchema(dbConn); err != nil {
		return err
	}
	estadoSync = strings.TrimSpace(strings.ToLower(estadoSync))
	if estadoSync == "" {
		estadoSync = "pendiente"
	}
	if carritoID < 0 {
		carritoID = 0
	}
	_, err := ExecCompat(dbConn, `UPDATE empresa_ventas_offline_sync SET
		estado_sync=?,
		carrito_id=?,
		documento_codigo=?,
		resultado_json=?,
		error_mensaje=?,
		fecha_sync=CASE WHEN ?='sincronizado' THEN `+sqlNowExpr()+` ELSE fecha_sync END,
		fecha_actualizacion=`+sqlNowExpr()+`
	WHERE empresa_id=? AND sync_key=?`,
		estadoSync,
		carritoID,
		strings.TrimSpace(documentoCodigo),
		strings.TrimSpace(resultadoJSON),
		strings.TrimSpace(errorMensaje),
		estadoSync,
		empresaID,
		strings.TrimSpace(syncKey),
	)
	return err
}

// ListEmpresaVentasOfflineSync lista ventas offline recientes para auditoria operativa.
func ListEmpresaVentasOfflineSync(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaVentaOfflineSync, error) {
	if err := EnsureEmpresaVentasOfflineSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT
		id, empresa_id, sync_key, COALESCE(carrito_id,0), COALESCE(documento_codigo,''),
		COALESCE(payload_json,''), COALESCE(resultado_json,''), COALESCE(estado_sync,'pendiente'),
		COALESCE(error_mensaje,''), COALESCE(fecha_offline,''), COALESCE(fecha_sync,''),
		COALESCE(usuario_creador,''), COALESCE(estado,'activo'), COALESCE(observaciones,''),
		COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,'')
	FROM empresa_ventas_offline_sync
	WHERE empresa_id=?
	ORDER BY id DESC
	LIMIT %d`, limit), empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaVentaOfflineSync{}
	for rows.Next() {
		var item EmpresaVentaOfflineSync
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.SyncKey,
			&item.CarritoID,
			&item.DocumentoCodigo,
			&item.PayloadJSON,
			&item.ResultadoJSON,
			&item.EstadoSync,
			&item.ErrorMensaje,
			&item.FechaOffline,
			&item.FechaSync,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
			&item.FechaCreacion,
			&item.FechaActualizacion,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
