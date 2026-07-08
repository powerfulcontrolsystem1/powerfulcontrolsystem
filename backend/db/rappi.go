package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type EmpresaRappiConfig struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Activo             bool   `json:"activo"`
	Ambiente           string `json:"ambiente"`
	CountryDomain      string `json:"country_domain"`
	NewDomain          string `json:"new_domain"`
	ClientID           string `json:"client_id"`
	ClientSecretRef    string `json:"client_secret_ref,omitempty"`
	WebhookSecretRef   string `json:"webhook_secret_ref,omitempty"`
	StoreIntegrationID string `json:"store_integration_id,omitempty"`
	RappiStoreID       string `json:"rappi_store_id,omitempty"`
	AutoTomarOrdenes   bool   `json:"auto_tomar_ordenes"`
	CookingTimeMinutes int    `json:"cooking_time_minutes"`
	CrearVentaInterna  bool   `json:"crear_venta_interna"`
	Observaciones      string `json:"observaciones,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
}

type EmpresaRappiOrderLog struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	RappiOrderID       string  `json:"rappi_order_id"`
	RappiStoreID       string  `json:"rappi_store_id,omitempty"`
	StoreIntegrationID string  `json:"store_integration_id,omitempty"`
	EstadoRappi        string  `json:"estado_rappi,omitempty"`
	EstadoLocal        string  `json:"estado_local,omitempty"`
	Total              float64 `json:"total,omitempty"`
	Moneda             string  `json:"moneda,omitempty"`
	ItemsJSON          string  `json:"items_json,omitempty"`
	RawPayloadJSON     string  `json:"raw_payload_json,omitempty"`
	Origen             string  `json:"origen,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

func EnsureEmpresaRappiSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_rappi_configuracion (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			activo INTEGER NOT NULL DEFAULT 0,
			ambiente TEXT NOT NULL DEFAULT 'development',
			country_domain TEXT NOT NULL DEFAULT 'https://api.rappi.com.co',
			new_domain TEXT NOT NULL DEFAULT 'https://api.rappi.com.co',
			client_id TEXT,
			client_secret_ref TEXT,
			webhook_secret_ref TEXT,
			store_integration_id TEXT,
			rappi_store_id TEXT,
			auto_tomar_ordenes INTEGER NOT NULL DEFAULT 0,
			cooking_time_minutes INTEGER NOT NULL DEFAULT 15,
			crear_venta_interna INTEGER NOT NULL DEFAULT 0,
			observaciones TEXT,
			fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			estado TEXT NOT NULL DEFAULT 'activo'
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_empresa_rappi_config_empresa ON empresa_rappi_configuracion(empresa_id);`,
		`CREATE TABLE IF NOT EXISTS empresa_rappi_ordenes (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			rappi_order_id TEXT NOT NULL,
			rappi_store_id TEXT,
			store_integration_id TEXT,
			estado_rappi TEXT,
			estado_local TEXT,
			total NUMERIC DEFAULT 0,
			moneda TEXT DEFAULT 'COP',
			items_json TEXT,
			raw_payload_json TEXT,
			origen TEXT,
			fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			estado TEXT NOT NULL DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_empresa_rappi_orden ON empresa_rappi_ordenes(empresa_id, rappi_order_id);`,
		`CREATE INDEX IF NOT EXISTS idx_empresa_rappi_ordenes_empresa_fecha ON empresa_rappi_ordenes(empresa_id, fecha_actualizacion);`,
	}
	for _, stmt := range stmts {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func NormalizeRappiAmbiente(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "production", "prod", "real":
		return "production"
	default:
		return "development"
	}
}

func NormalizeRappiDomain(raw, fallback string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = fallback
	}
	value = strings.TrimRight(value, "/")
	if value == "" {
		return fallback
	}
	return value
}

func GetEmpresaRappiConfig(dbConn *sql.DB, empresaID int64) (EmpresaRappiConfig, error) {
	if dbConn == nil {
		return EmpresaRappiConfig{}, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return EmpresaRappiConfig{}, fmt.Errorf("empresa_id invalido")
	}
	if err := EnsureEmpresaRappiSchema(dbConn); err != nil {
		return EmpresaRappiConfig{}, err
	}
	var out EmpresaRappiConfig
	var activo, autoTomar, crearVenta int64
	err := queryRowSQLCompat(dbConn, `SELECT
		id, empresa_id, COALESCE(activo,0), COALESCE(ambiente,'development'),
		COALESCE(country_domain,''), COALESCE(new_domain,''), COALESCE(client_id,''),
		COALESCE(client_secret_ref,''), COALESCE(webhook_secret_ref,''), COALESCE(store_integration_id,''),
		COALESCE(rappi_store_id,''), COALESCE(auto_tomar_ordenes,0), COALESCE(cooking_time_minutes,15),
		COALESCE(crear_venta_interna,0), COALESCE(observaciones,''), COALESCE(CAST(fecha_creacion AS TEXT),''),
		COALESCE(CAST(fecha_actualizacion AS TEXT),''), COALESCE(usuario_creador,'')
	FROM empresa_rappi_configuracion
	WHERE empresa_id = ? AND COALESCE(estado,'activo') <> 'inactivo'
	LIMIT 1`, empresaID).Scan(
		&out.ID, &out.EmpresaID, &activo, &out.Ambiente, &out.CountryDomain, &out.NewDomain, &out.ClientID,
		&out.ClientSecretRef, &out.WebhookSecretRef, &out.StoreIntegrationID, &out.RappiStoreID, &autoTomar,
		&out.CookingTimeMinutes, &crearVenta, &out.Observaciones, &out.FechaCreacion, &out.FechaActualizacion,
		&out.UsuarioCreador,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return DefaultEmpresaRappiConfig(empresaID), nil
		}
		return EmpresaRappiConfig{}, err
	}
	out.Activo = activo > 0
	out.AutoTomarOrdenes = autoTomar > 0
	out.CrearVentaInterna = crearVenta > 0
	out.Ambiente = NormalizeRappiAmbiente(out.Ambiente)
	out.CountryDomain = NormalizeRappiDomain(out.CountryDomain, "https://api.rappi.com.co")
	out.NewDomain = NormalizeRappiDomain(out.NewDomain, out.CountryDomain)
	if out.CookingTimeMinutes <= 0 {
		out.CookingTimeMinutes = 15
	}
	return out, nil
}

func DefaultEmpresaRappiConfig(empresaID int64) EmpresaRappiConfig {
	return EmpresaRappiConfig{
		EmpresaID:          empresaID,
		Ambiente:           "development",
		CountryDomain:      "https://api.rappi.com.co",
		NewDomain:          "https://api.rappi.com.co",
		CookingTimeMinutes: 15,
		Estado:             "activo",
	}
}

func SaveEmpresaRappiConfig(dbConn *sql.DB, cfg EmpresaRappiConfig) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if cfg.EmpresaID <= 0 {
		return fmt.Errorf("empresa_id invalido")
	}
	if err := EnsureEmpresaRappiSchema(dbConn); err != nil {
		return err
	}
	cfg.Ambiente = NormalizeRappiAmbiente(cfg.Ambiente)
	cfg.CountryDomain = NormalizeRappiDomain(cfg.CountryDomain, "https://api.rappi.com.co")
	cfg.NewDomain = NormalizeRappiDomain(cfg.NewDomain, cfg.CountryDomain)
	if cfg.CookingTimeMinutes <= 0 {
		cfg.CookingTimeMinutes = 15
	}
	_, err := execSQLCompat(dbConn, `INSERT INTO empresa_rappi_configuracion (
		empresa_id, activo, ambiente, country_domain, new_domain, client_id, client_secret_ref,
		webhook_secret_ref, store_integration_id, rappi_store_id, auto_tomar_ordenes,
		cooking_time_minutes, crear_venta_interna, observaciones, fecha_creacion,
		fecha_actualizacion, usuario_creador, estado
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, 'activo')
	ON CONFLICT(empresa_id) DO UPDATE SET
		activo = excluded.activo,
		ambiente = excluded.ambiente,
		country_domain = excluded.country_domain,
		new_domain = excluded.new_domain,
		client_id = excluded.client_id,
		client_secret_ref = excluded.client_secret_ref,
		webhook_secret_ref = excluded.webhook_secret_ref,
		store_integration_id = excluded.store_integration_id,
		rappi_store_id = excluded.rappi_store_id,
		auto_tomar_ordenes = excluded.auto_tomar_ordenes,
		cooking_time_minutes = excluded.cooking_time_minutes,
		crear_venta_interna = excluded.crear_venta_interna,
		observaciones = excluded.observaciones,
		fecha_actualizacion = CURRENT_TIMESTAMP,
		usuario_creador = excluded.usuario_creador,
		estado = 'activo'`,
		cfg.EmpresaID, boolToIntRappi(cfg.Activo), cfg.Ambiente, cfg.CountryDomain, cfg.NewDomain,
		strings.TrimSpace(cfg.ClientID), strings.TrimSpace(cfg.ClientSecretRef), strings.TrimSpace(cfg.WebhookSecretRef),
		strings.TrimSpace(cfg.StoreIntegrationID), strings.TrimSpace(cfg.RappiStoreID), boolToIntRappi(cfg.AutoTomarOrdenes),
		cfg.CookingTimeMinutes, boolToIntRappi(cfg.CrearVentaInterna), strings.TrimSpace(cfg.Observaciones),
		strings.TrimSpace(cfg.UsuarioCreador),
	)
	return err
}

func UpsertEmpresaRappiOrderLog(dbConn *sql.DB, row EmpresaRappiOrderLog) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	if row.EmpresaID <= 0 || strings.TrimSpace(row.RappiOrderID) == "" {
		return 0, fmt.Errorf("empresa_id y rappi_order_id son obligatorios")
	}
	if err := EnsureEmpresaRappiSchema(dbConn); err != nil {
		return 0, err
	}
	if strings.TrimSpace(row.Moneda) == "" {
		row.Moneda = "COP"
	}
	if strings.TrimSpace(row.EstadoLocal) == "" {
		row.EstadoLocal = "recibida"
	}
	var id int64
	err := queryRowSQLCompat(dbConn, `INSERT INTO empresa_rappi_ordenes (
		empresa_id, rappi_order_id, rappi_store_id, store_integration_id, estado_rappi,
		estado_local, total, moneda, items_json, raw_payload_json, origen, fecha_creacion,
		fecha_actualizacion, usuario_creador, estado, observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, 'activo', ?)
	ON CONFLICT(empresa_id, rappi_order_id) DO UPDATE SET
		rappi_store_id = excluded.rappi_store_id,
		store_integration_id = excluded.store_integration_id,
		estado_rappi = excluded.estado_rappi,
		estado_local = excluded.estado_local,
		total = excluded.total,
		moneda = excluded.moneda,
		items_json = CASE WHEN excluded.items_json = '' THEN empresa_rappi_ordenes.items_json ELSE excluded.items_json END,
		raw_payload_json = CASE WHEN excluded.raw_payload_json = '' THEN empresa_rappi_ordenes.raw_payload_json ELSE excluded.raw_payload_json END,
		origen = excluded.origen,
		fecha_actualizacion = CURRENT_TIMESTAMP,
		usuario_creador = excluded.usuario_creador,
		observaciones = excluded.observaciones
	RETURNING id`,
		row.EmpresaID, strings.TrimSpace(row.RappiOrderID), strings.TrimSpace(row.RappiStoreID),
		strings.TrimSpace(row.StoreIntegrationID), strings.TrimSpace(row.EstadoRappi), strings.TrimSpace(row.EstadoLocal),
		row.Total, strings.TrimSpace(row.Moneda), strings.TrimSpace(row.ItemsJSON), strings.TrimSpace(row.RawPayloadJSON),
		strings.TrimSpace(row.Origen), strings.TrimSpace(row.UsuarioCreador), strings.TrimSpace(row.Observaciones),
	).Scan(&id)
	return id, err
}

func ListEmpresaRappiOrderLogs(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaRappiOrderLog, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id invalido")
	}
	if err := EnsureEmpresaRappiSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 80
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT
		id, empresa_id, COALESCE(rappi_order_id,''), COALESCE(rappi_store_id,''), COALESCE(store_integration_id,''),
		COALESCE(estado_rappi,''), COALESCE(estado_local,''), COALESCE(total,0), COALESCE(moneda,'COP'),
		COALESCE(items_json,''), COALESCE(raw_payload_json,''), COALESCE(origen,''), COALESCE(CAST(fecha_creacion AS TEXT),''),
		COALESCE(CAST(fecha_actualizacion AS TEXT),''), COALESCE(usuario_creador,''), COALESCE(estado,'activo'), COALESCE(observaciones,'')
	FROM empresa_rappi_ordenes
	WHERE empresa_id = ? AND COALESCE(estado,'activo') <> 'inactivo'
	ORDER BY fecha_actualizacion DESC, id DESC
	LIMIT ?`, empresaID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmpresaRappiOrderLog, 0)
	for rows.Next() {
		var item EmpresaRappiOrderLog
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.RappiOrderID, &item.RappiStoreID, &item.StoreIntegrationID,
			&item.EstadoRappi, &item.EstadoLocal, &item.Total, &item.Moneda, &item.ItemsJSON, &item.RawPayloadJSON,
			&item.Origen, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador, &item.Estado,
			&item.Observaciones); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func boolToIntRappi(v bool) int {
	if v {
		return 1
	}
	return 0
}
