package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"
	"time"
)

type EstacionVIPCodigo struct {
	ID             int64  `json:"id"`
	EmpresaID      int64  `json:"empresa_id"`
	EstacionID     int64  `json:"estacion_id"`
	CarritoID      int64  `json:"carrito_id"`
	Codigo         string `json:"codigo"`
	ExpiraEn       string `json:"expira_en"`
	Estado         string `json:"estado"`
	FechaCreacion  string `json:"fecha_creacion,omitempty"`
	UsuarioCreador string `json:"usuario_creador,omitempty"`
	Observaciones  string `json:"observaciones,omitempty"`
}

func EnsureEstacionVIPCodigosSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	if dbConn == nil {
		return nil
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS estacion_vip_codigos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			estacion_id INTEGER NOT NULL,
			carrito_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			expira_en TEXT,
			estado TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_estacion_vip_codigo ON estacion_vip_codigos(codigo);`,
		`CREATE INDEX IF NOT EXISTS ix_estacion_vip_empresa_estacion ON estacion_vip_codigos(empresa_id, estacion_id, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_estacion_vip_empresa_carrito ON estacion_vip_codigos(empresa_id, carrito_id, estado);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}
	if err := ensureColumnIfMissing(dbConn, "estacion_vip_codigos", "expira_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "estacion_vip_codigos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "estacion_vip_codigos", "observaciones", "TEXT"); err != nil {
		return err
	}
	return nil
}

func generateVIPCode() (string, error) {
	var b [10]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	// base32 sin padding, fácil de dictar, sólo A-Z2-7
	out := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b[:])
	out = strings.TrimSpace(out)
	if len(out) > 14 {
		out = out[:14]
	}
	return out, nil
}

func CreateEstacionVIPCodigo(dbConn *sql.DB, empresaID, estacionID, carritoID int64, ttlMinutes int, usuarioCreador, observaciones string) (*EstacionVIPCodigo, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 || estacionID <= 0 || carritoID <= 0 {
		return nil, errors.New("empresa_id, estacion_id y carrito_id son obligatorios")
	}
	if ttlMinutes <= 0 {
		ttlMinutes = 720 // 12h
	}
	if err := EnsureEstacionVIPCodigosSchema(dbConn); err != nil {
		return nil, err
	}

	expira := time.Now().Add(time.Duration(ttlMinutes) * time.Minute).Format("2006-01-02 15:04:05")
	var codigo string
	var lastErr error
	for i := 0; i < 6; i++ {
		c, err := generateVIPCode()
		if err != nil {
			return nil, err
		}
		codigo = c
		_, err = dbConn.Exec(`INSERT INTO estacion_vip_codigos (empresa_id, estacion_id, carrito_id, codigo, expira_en, estado, usuario_creador, observaciones)
			VALUES (?, ?, ?, ?, ?, 'activo', ?, ?)`,
			empresaID, estacionID, carritoID, codigo, expira, strings.TrimSpace(usuarioCreador), strings.TrimSpace(observaciones),
		)
		if err == nil {
			return &EstacionVIPCodigo{
				EmpresaID:      empresaID,
				EstacionID:     estacionID,
				CarritoID:      carritoID,
				Codigo:         codigo,
				ExpiraEn:       expira,
				Estado:         "activo",
				UsuarioCreador: strings.TrimSpace(usuarioCreador),
				Observaciones:  strings.TrimSpace(observaciones),
			}, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("no se pudo generar codigo vip: %w", lastErr)
}

func GetVIPByCodigo(dbConn *sql.DB, codigo string) (*EstacionVIPCodigo, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	codigo = strings.TrimSpace(strings.ToUpper(codigo))
	if codigo == "" {
		return nil, errors.New("codigo es obligatorio")
	}
	if err := EnsureEstacionVIPCodigosSchema(dbConn); err != nil {
		return nil, err
	}
	row := dbConn.QueryRow(`SELECT id, empresa_id, estacion_id, carrito_id, codigo, COALESCE(expira_en,''), COALESCE(NULLIF(TRIM(estado),''),'activo'),
		COALESCE(fecha_creacion,''), COALESCE(usuario_creador,''), COALESCE(observaciones,'')
		FROM estacion_vip_codigos WHERE UPPER(TRIM(codigo)) = ? LIMIT 1`, codigo)
	var out EstacionVIPCodigo
	if err := row.Scan(&out.ID, &out.EmpresaID, &out.EstacionID, &out.CarritoID, &out.Codigo, &out.ExpiraEn, &out.Estado, &out.FechaCreacion, &out.UsuarioCreador, &out.Observaciones); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func InvalidateVIPCodesForCarrito(dbConn *sql.DB, empresaID, carritoID int64, motivo string) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if empresaID <= 0 || carritoID <= 0 {
		return nil
	}
	if err := EnsureEstacionVIPCodigosSchema(dbConn); err != nil {
		return err
	}
	obs := strings.TrimSpace(motivo)
	if obs == "" {
		obs = "carrito_pagado"
	}
	_, err := dbConn.Exec(`UPDATE estacion_vip_codigos
		SET estado='inactivo', observaciones=CASE WHEN TRIM(observaciones)<>'' THEN observaciones || ' | ' || ? ELSE ? END
		WHERE empresa_id = ? AND carrito_id = ? AND LOWER(COALESCE(NULLIF(TRIM(estado),''),'activo')) = 'activo'`, obs, obs, empresaID, carritoID)
	return err
}
