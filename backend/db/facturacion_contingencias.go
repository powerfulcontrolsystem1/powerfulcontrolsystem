package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const empresaFacturacionContingenciasFingerprint = "empresa-facturacion-contingencias:v1:tenant-incidents-paper-authorization-history-document-deadline"

const (
	FacturacionContingenciaFallaDIAN       = "falla_dian"
	FacturacionContingenciaFallaFacturador = "falla_facturador"
)

type EmpresaFacturacionContingenciaConfiguracion struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	PaisCodigo         string `json:"pais_codigo"`
	Prefijo            string `json:"prefijo"`
	ResolucionNumero   string `json:"resolucion_numero"`
	FechaDesde         string `json:"fecha_desde"`
	FechaHasta         string `json:"fecha_hasta"`
	RangoDesde         int64  `json:"rango_desde"`
	RangoHasta         int64  `json:"rango_hasta"`
	ProximoNumero      int64  `json:"proximo_numero"`
	Estado             string `json:"estado"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
}

type EmpresaFacturacionContingencia struct {
	ID                       int64  `json:"id"`
	EmpresaID                int64  `json:"empresa_id"`
	ConfiguracionTalonarioID int64  `json:"configuracion_talonario_id,omitempty"`
	PaisCodigo               string `json:"pais_codigo"`
	Tipo                     string `json:"tipo"`
	Estado                   string `json:"estado"`
	Motivo                   string `json:"motivo"`
	EvidenciaReferencia      string `json:"evidencia_referencia"`
	FechaInicio              string `json:"fecha_inicio"`
	FechaRecuperacion        string `json:"fecha_recuperacion,omitempty"`
	FechaLimiteTransmision   string `json:"fecha_limite_transmision,omitempty"`
	FechaCierre              string `json:"fecha_cierre,omitempty"`
	UsuarioCreador           string `json:"usuario_creador"`
	UsuarioRecuperacion      string `json:"usuario_recuperacion,omitempty"`
	UsuarioCierre            string `json:"usuario_cierre,omitempty"`
	Observaciones            string `json:"observaciones,omitempty"`
	DocumentosPendientes     int64  `json:"documentos_pendientes"`
}

type EmpresaFacturacionContingenciaDocumento struct {
	ID                   int64   `json:"id"`
	EmpresaID            int64   `json:"empresa_id"`
	ContingenciaID       int64   `json:"contingencia_id"`
	ConfiguracionID      int64   `json:"configuracion_talonario_id,omitempty"`
	TipoDocumento        string  `json:"tipo_documento"`
	DocumentoCodigo      string  `json:"documento_codigo"`
	NumeroPapel          string  `json:"numero_papel,omitempty"`
	FechaExpedicionPapel string  `json:"fecha_expedicion_papel,omitempty"`
	EstadoTransmision    string  `json:"estado_transmision"`
	RetryID              int64   `json:"retry_id,omitempty"`
	CarritoID            int64   `json:"carrito_id,omitempty"`
	ComprobanteCodigo    string  `json:"comprobante_codigo,omitempty"`
	MontoTotal           float64 `json:"monto_total,omitempty"`
	FechaActualizacion   string  `json:"fecha_actualizacion,omitempty"`
}

func applyEmpresaFacturacionContingenciasTx(_ context.Context, tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("facturacion contingencias migration transaction is nil")
	}
	statements := []string{
		`CREATE TABLE IF NOT EXISTS empresa_facturacion_contingencia_configuracion (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			pais_codigo TEXT NOT NULL DEFAULT 'CO',
			prefijo TEXT NOT NULL,
			resolucion_numero TEXT NOT NULL,
			fecha_desde DATE NOT NULL,
			fecha_hasta DATE NOT NULL,
			rango_desde BIGINT NOT NULL,
			rango_hasta BIGINT NOT NULL,
			proximo_numero BIGINT NOT NULL,
			estado TEXT NOT NULL DEFAULT 'configurando',
			usuario_creador TEXT NOT NULL DEFAULT '',
			observaciones TEXT NOT NULL DEFAULT '',
			fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CHECK (pais_codigo = 'CO'),
			CHECK (estado IN ('configurando','activo','suspendido')),
			CHECK (rango_desde > 0 AND rango_hasta >= rango_desde),
			CHECK (proximo_numero BETWEEN rango_desde AND rango_hasta + 1),
			CHECK (fecha_hasta >= fecha_desde),
			UNIQUE (id, empresa_id),
			UNIQUE (empresa_id, prefijo, resolucion_numero)
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_facturacion_contingencia_config_activa
			ON empresa_facturacion_contingencia_configuracion (empresa_id) WHERE estado = 'activo'`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_facturacion_contingencia_config_historial
			ON empresa_facturacion_contingencia_configuracion (empresa_id, fecha_actualizacion DESC, id DESC)`,
		`CREATE TABLE IF NOT EXISTS empresa_facturacion_contingencias (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			configuracion_talonario_id BIGINT,
			pais_codigo TEXT NOT NULL DEFAULT 'CO',
			tipo TEXT NOT NULL,
			estado TEXT NOT NULL DEFAULT 'activa',
			motivo TEXT NOT NULL,
			evidencia_referencia TEXT NOT NULL,
			fecha_inicio TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			fecha_recuperacion TIMESTAMPTZ,
			fecha_limite_transmision TIMESTAMPTZ,
			fecha_cierre TIMESTAMPTZ,
			usuario_creador TEXT NOT NULL,
			usuario_recuperacion TEXT,
			usuario_cierre TEXT,
			observaciones TEXT NOT NULL DEFAULT '',
			fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CHECK (pais_codigo = 'CO'),
			CHECK (tipo IN ('falla_dian','falla_facturador')),
			CHECK (estado IN ('activa','recuperada','cerrada')),
			CHECK ((tipo = 'falla_facturador' AND configuracion_talonario_id IS NOT NULL) OR (tipo = 'falla_dian' AND configuracion_talonario_id IS NULL)),
			UNIQUE (id, empresa_id),
			FOREIGN KEY (configuracion_talonario_id, empresa_id)
				REFERENCES empresa_facturacion_contingencia_configuracion (id, empresa_id)
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_facturacion_contingencia_activa
			ON empresa_facturacion_contingencias (empresa_id, tipo) WHERE estado = 'activa'`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_facturacion_contingencia_historial
			ON empresa_facturacion_contingencias (empresa_id, fecha_inicio DESC, id DESC)`,
		`CREATE TABLE IF NOT EXISTS empresa_facturacion_contingencia_documentos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			contingencia_id BIGINT NOT NULL,
			configuracion_talonario_id BIGINT,
			tipo_documento TEXT NOT NULL,
			documento_codigo TEXT NOT NULL,
			numero_papel TEXT NOT NULL DEFAULT '',
			fecha_expedicion_papel DATE,
			estado_transmision TEXT NOT NULL DEFAULT 'pendiente',
			retry_id BIGINT NOT NULL DEFAULT 0,
			carrito_id BIGINT NOT NULL DEFAULT 0,
			comprobante_codigo TEXT NOT NULL DEFAULT '',
			monto_total NUMERIC(18,2) NOT NULL DEFAULT 0,
			fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (empresa_id, contingencia_id, tipo_documento, documento_codigo),
			CHECK (estado_transmision IN ('pendiente','enviado','aceptado','rechazado','bloqueado')),
			CHECK ((tipo_documento = 'factura_talonario_papel' AND configuracion_talonario_id IS NOT NULL) OR tipo_documento <> 'factura_talonario_papel'),
			FOREIGN KEY (contingencia_id, empresa_id)
				REFERENCES empresa_facturacion_contingencias (id, empresa_id),
			FOREIGN KEY (configuracion_talonario_id, empresa_id)
				REFERENCES empresa_facturacion_contingencia_configuracion (id, empresa_id)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_facturacion_contingencia_documentos_pendientes
			ON empresa_facturacion_contingencia_documentos (empresa_id, contingencia_id, estado_transmision)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_facturacion_contingencia_numero_papel
			ON empresa_facturacion_contingencia_documentos (empresa_id, numero_papel) WHERE numero_papel <> ''`,
	}
	for _, statement := range statements {
		if _, err := execTxSQLCompat(tx, statement); err != nil {
			return err
		}
	}
	return nil
}

func EmpresaFacturacionContingenciasSchemaReady(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("database not available")
	}
	for _, table := range []string{"empresa_facturacion_contingencia_configuracion", "empresa_facturacion_contingencias", "empresa_facturacion_contingencia_documentos"} {
		var name sql.NullString
		if err := queryRowSQLCompat(dbConn, `SELECT to_regclass(?)`, table).Scan(&name); err != nil {
			return err
		}
		if !name.Valid || strings.TrimSpace(name.String) == "" {
			return fmt.Errorf("facturacion contingency schema missing; run pcs-migrate")
		}
	}
	return nil
}

func normalizeFacturacionContingenciaTipo(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case FacturacionContingenciaFallaDIAN:
		return FacturacionContingenciaFallaDIAN
	case FacturacionContingenciaFallaFacturador:
		return FacturacionContingenciaFallaFacturador
	default:
		return ""
	}
}

func validateFacturacionContingenciaConfiguracion(in *EmpresaFacturacionContingenciaConfiguracion) error {
	if in == nil || in.EmpresaID <= 0 {
		return fmt.Errorf("empresa_id es obligatorio")
	}
	in.PaisCodigo = strings.ToUpper(strings.TrimSpace(in.PaisCodigo))
	if in.PaisCodigo == "" {
		in.PaisCodigo = "CO"
	}
	if in.PaisCodigo != "CO" {
		return fmt.Errorf("la contingencia fiscal implementada aplica solo a Colombia")
	}
	in.Prefijo = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(in.Prefijo), " ", ""))
	in.ResolucionNumero = strings.TrimSpace(in.ResolucionNumero)
	if in.Prefijo == "" || in.ResolucionNumero == "" {
		return fmt.Errorf("prefijo y resolucion de talonario o papel son obligatorios")
	}
	from, err := time.Parse("2006-01-02", strings.TrimSpace(in.FechaDesde))
	if err != nil {
		return fmt.Errorf("fecha_desde invalida")
	}
	to, err := time.Parse("2006-01-02", strings.TrimSpace(in.FechaHasta))
	if err != nil || to.Before(from) {
		return fmt.Errorf("fecha_hasta invalida")
	}
	if in.RangoDesde <= 0 || in.RangoHasta < in.RangoDesde || in.ProximoNumero < in.RangoDesde || in.ProximoNumero > in.RangoHasta+1 {
		return fmt.Errorf("rango o proximo numero de contingencia invalido")
	}
	in.Estado = strings.ToLower(strings.TrimSpace(in.Estado))
	if in.Estado != "configurando" && in.Estado != "activo" && in.Estado != "suspendido" {
		return fmt.Errorf("estado de configuracion de contingencia invalido")
	}
	if in.Estado == "activo" {
		bogota := time.FixedZone("America/Bogota", -5*60*60)
		now := time.Now().In(bogota)
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, bogota)
		if today.Before(from) || today.After(to) || in.ProximoNumero > in.RangoHasta {
			return fmt.Errorf("la autorizacion de talonario o papel no esta vigente o no tiene numeracion disponible")
		}
	}
	return nil
}

func UpsertEmpresaFacturacionContingenciaConfiguracion(dbConn *sql.DB, in EmpresaFacturacionContingenciaConfiguracion) (*EmpresaFacturacionContingenciaConfiguracion, error) {
	if err := validateFacturacionContingenciaConfiguracion(&in); err != nil {
		return nil, err
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var existingID, previousNext int64
	previousErr := queryRowTxSQLCompat(tx, `SELECT id,proximo_numero FROM empresa_facturacion_contingencia_configuracion
		WHERE empresa_id=? AND prefijo=? AND resolucion_numero=? FOR UPDATE`, in.EmpresaID, in.Prefijo, in.ResolucionNumero).Scan(&existingID, &previousNext)
	if previousErr != nil && previousErr != sql.ErrNoRows {
		return nil, previousErr
	}
	if previousErr == nil && in.ProximoNumero < previousNext {
		in.ProximoNumero = previousNext
	}
	if in.Estado == "activo" {
		var activeIncidentForOtherConfig bool
		if err = queryRowTxSQLCompat(tx, `SELECT EXISTS (SELECT 1 FROM empresa_facturacion_contingencias
			WHERE empresa_id=? AND tipo='falla_facturador' AND estado='activa'
			AND COALESCE(configuracion_talonario_id,0)<>?)`, in.EmpresaID, existingID).Scan(&activeIncidentForOtherConfig); err != nil {
			return nil, err
		}
		if activeIncidentForOtherConfig {
			return nil, fmt.Errorf("no se puede cambiar la autorizacion durante una contingencia activa del facturador")
		}
		if _, err = execTxSQLCompat(tx, `UPDATE empresa_facturacion_contingencia_configuracion
			SET estado='suspendido',fecha_actualizacion=CURRENT_TIMESTAMP
			WHERE empresa_id=? AND estado='activo' AND id<>?`, in.EmpresaID, existingID); err != nil {
			return nil, err
		}
	}
	var storedID int64
	err = queryRowTxSQLCompat(tx, `INSERT INTO empresa_facturacion_contingencia_configuracion
		(empresa_id,pais_codigo,prefijo,resolucion_numero,fecha_desde,fecha_hasta,rango_desde,rango_hasta,proximo_numero,estado,usuario_creador,observaciones,fecha_actualizacion)
		VALUES (?,?,?,?,?::date,?::date,?,?,?,?,?,?,CURRENT_TIMESTAMP)
		ON CONFLICT (empresa_id,prefijo,resolucion_numero) DO UPDATE SET
			pais_codigo=EXCLUDED.pais_codigo,prefijo=EXCLUDED.prefijo,resolucion_numero=EXCLUDED.resolucion_numero,
			fecha_desde=EXCLUDED.fecha_desde,fecha_hasta=EXCLUDED.fecha_hasta,rango_desde=EXCLUDED.rango_desde,
			rango_hasta=EXCLUDED.rango_hasta,proximo_numero=GREATEST(empresa_facturacion_contingencia_configuracion.proximo_numero,EXCLUDED.proximo_numero),
			estado=EXCLUDED.estado,usuario_creador=EXCLUDED.usuario_creador,observaciones=EXCLUDED.observaciones,fecha_actualizacion=CURRENT_TIMESTAMP
		RETURNING id`, in.EmpresaID, in.PaisCodigo, in.Prefijo, in.ResolucionNumero, in.FechaDesde, in.FechaHasta, in.RangoDesde, in.RangoHasta, in.ProximoNumero, in.Estado, strings.TrimSpace(in.UsuarioCreador), strings.TrimSpace(in.Observaciones)).Scan(&storedID)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return getEmpresaFacturacionContingenciaConfiguracionByID(dbConn, in.EmpresaID, storedID)
}

func GetEmpresaFacturacionContingenciaConfiguracion(dbConn *sql.DB, empresaID int64) (*EmpresaFacturacionContingenciaConfiguracion, error) {
	var id int64
	err := queryRowSQLCompat(dbConn, `SELECT id FROM empresa_facturacion_contingencia_configuracion WHERE empresa_id=?
		ORDER BY (estado='activo') DESC,fecha_actualizacion DESC,id DESC LIMIT 1`, empresaID).Scan(&id)
	if err != nil {
		return nil, err
	}
	return getEmpresaFacturacionContingenciaConfiguracionByID(dbConn, empresaID, id)
}

func getEmpresaFacturacionContingenciaConfiguracionByID(dbConn *sql.DB, empresaID, id int64) (*EmpresaFacturacionContingenciaConfiguracion, error) {
	var item EmpresaFacturacionContingenciaConfiguracion
	err := queryRowSQLCompat(dbConn, `SELECT id,empresa_id,pais_codigo,prefijo,resolucion_numero,fecha_desde::text,fecha_hasta::text,
		rango_desde,rango_hasta,proximo_numero,estado,COALESCE(usuario_creador,''),COALESCE(observaciones,''),fecha_actualizacion::text
		FROM empresa_facturacion_contingencia_configuracion WHERE empresa_id=? AND id=?`, empresaID, id).Scan(
		&item.ID, &item.EmpresaID, &item.PaisCodigo, &item.Prefijo, &item.ResolucionNumero, &item.FechaDesde, &item.FechaHasta,
		&item.RangoDesde, &item.RangoHasta, &item.ProximoNumero, &item.Estado, &item.UsuarioCreador, &item.Observaciones, &item.FechaActualizacion)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func OpenEmpresaFacturacionContingencia(dbConn *sql.DB, empresaID int64, tipo, motivo, evidencia, usuario, observaciones string) (*EmpresaFacturacionContingencia, error) {
	tipo = normalizeFacturacionContingenciaTipo(tipo)
	if empresaID <= 0 || tipo == "" {
		return nil, fmt.Errorf("empresa o tipo de contingencia invalido")
	}
	motivo, evidencia, usuario = strings.TrimSpace(motivo), strings.TrimSpace(evidencia), strings.TrimSpace(usuario)
	if len(motivo) < 20 || len(evidencia) < 8 || usuario == "" {
		return nil, fmt.Errorf("motivo, evidencia y usuario de contingencia son obligatorios")
	}
	var configuracionID interface{}
	if tipo == FacturacionContingenciaFallaFacturador {
		cfg, err := GetEmpresaFacturacionContingenciaConfiguracion(dbConn, empresaID)
		if err != nil {
			return nil, fmt.Errorf("la contingencia del facturador requiere autorizacion de talonario o papel vigente: %w", err)
		}
		if cfg.Estado != "activo" || validateFacturacionContingenciaConfiguracion(cfg) != nil {
			return nil, fmt.Errorf("la autorizacion de talonario o papel no esta activa, vigente o disponible")
		}
		configuracionID = cfg.ID
	}
	_, err := ExecCompat(dbConn, `INSERT INTO empresa_facturacion_contingencias
		(empresa_id,configuracion_talonario_id,pais_codigo,tipo,estado,motivo,evidencia_referencia,usuario_creador,observaciones)
		VALUES (?,?,'CO',?,'activa',?,?,?,?)`, empresaID, configuracionID, tipo, motivo, evidencia, usuario, strings.TrimSpace(observaciones))
	if err != nil {
		return nil, err
	}
	return GetActiveEmpresaFacturacionContingencia(dbConn, empresaID, tipo)
}

func GetActiveEmpresaFacturacionContingencia(dbConn *sql.DB, empresaID int64, tipo string) (*EmpresaFacturacionContingencia, error) {
	tipo = normalizeFacturacionContingenciaTipo(tipo)
	return scanEmpresaFacturacionContingencia(queryRowSQLCompat(dbConn, `SELECT c.id,c.empresa_id,COALESCE(c.configuracion_talonario_id,0),c.pais_codigo,c.tipo,c.estado,c.motivo,c.evidencia_referencia,
		c.fecha_inicio::text,COALESCE(c.fecha_recuperacion::text,''),COALESCE(c.fecha_limite_transmision::text,''),COALESCE(c.fecha_cierre::text,''),
		c.usuario_creador,COALESCE(c.usuario_recuperacion,''),COALESCE(c.usuario_cierre,''),COALESCE(c.observaciones,''),
		(SELECT COUNT(*) FROM empresa_facturacion_contingencia_documentos d WHERE d.empresa_id=c.empresa_id AND d.contingencia_id=c.id AND d.estado_transmision NOT IN ('aceptado'))
		FROM empresa_facturacion_contingencias c WHERE c.empresa_id=? AND c.tipo=? AND c.estado='activa' ORDER BY c.id DESC LIMIT 1`, empresaID, tipo))
}

type facturacionContingenciaScanner interface{ Scan(...interface{}) error }

func scanEmpresaFacturacionContingencia(row facturacionContingenciaScanner) (*EmpresaFacturacionContingencia, error) {
	var item EmpresaFacturacionContingencia
	if err := row.Scan(&item.ID, &item.EmpresaID, &item.ConfiguracionTalonarioID, &item.PaisCodigo, &item.Tipo, &item.Estado, &item.Motivo, &item.EvidenciaReferencia,
		&item.FechaInicio, &item.FechaRecuperacion, &item.FechaLimiteTransmision, &item.FechaCierre, &item.UsuarioCreador,
		&item.UsuarioRecuperacion, &item.UsuarioCierre, &item.Observaciones, &item.DocumentosPendientes); err != nil {
		return nil, err
	}
	return &item, nil
}

func ListEmpresaFacturacionContingencias(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaFacturacionContingencia, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT c.id,c.empresa_id,COALESCE(c.configuracion_talonario_id,0),c.pais_codigo,c.tipo,c.estado,c.motivo,c.evidencia_referencia,
		c.fecha_inicio::text,COALESCE(c.fecha_recuperacion::text,''),COALESCE(c.fecha_limite_transmision::text,''),COALESCE(c.fecha_cierre::text,''),
		c.usuario_creador,COALESCE(c.usuario_recuperacion,''),COALESCE(c.usuario_cierre,''),COALESCE(c.observaciones,''),
		(SELECT COUNT(*) FROM empresa_facturacion_contingencia_documentos d WHERE d.empresa_id=c.empresa_id AND d.contingencia_id=c.id AND d.estado_transmision NOT IN ('aceptado'))
		FROM empresa_facturacion_contingencias c WHERE c.empresa_id=? ORDER BY c.fecha_inicio DESC,c.id DESC LIMIT ?`, empresaID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmpresaFacturacionContingencia, 0)
	for rows.Next() {
		item, err := scanEmpresaFacturacionContingencia(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *item)
	}
	return out, rows.Err()
}

func ListEmpresaFacturacionContingenciaDocumentos(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaFacturacionContingenciaDocumento, error) {
	if dbConn == nil || empresaID <= 0 {
		return nil, fmt.Errorf("conexion empresarial y empresa_id son obligatorios")
	}
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,contingencia_id,COALESCE(configuracion_talonario_id,0),tipo_documento,documento_codigo,
		COALESCE(numero_papel,''),COALESCE(fecha_expedicion_papel::text,''),estado_transmision,retry_id,
		carrito_id,COALESCE(comprobante_codigo,''),monto_total,fecha_actualizacion::text
		FROM empresa_facturacion_contingencia_documentos
		WHERE empresa_id=? ORDER BY fecha_actualizacion DESC,id DESC LIMIT ?`, empresaID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmpresaFacturacionContingenciaDocumento, 0)
	for rows.Next() {
		var item EmpresaFacturacionContingenciaDocumento
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.ContingenciaID, &item.ConfiguracionID, &item.TipoDocumento, &item.DocumentoCodigo,
			&item.NumeroPapel, &item.FechaExpedicionPapel, &item.EstadoTransmision, &item.RetryID,
			&item.CarritoID, &item.ComprobanteCodigo, &item.MontoTotal, &item.FechaActualizacion); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func RecoverEmpresaFacturacionContingencia(dbConn *sql.DB, empresaID, id int64, usuario string) error {
	result, err := ExecCompat(dbConn, `UPDATE empresa_facturacion_contingencias SET estado='recuperada',fecha_recuperacion=CURRENT_TIMESTAMP,
		fecha_limite_transmision=CURRENT_TIMESTAMP + INTERVAL '48 hours',usuario_recuperacion=?,fecha_actualizacion=CURRENT_TIMESTAMP
		WHERE empresa_id=? AND id=? AND estado='activa'`, strings.TrimSpace(usuario), empresaID, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return sql.ErrNoRows
	}
	return nil
}

func CloseEmpresaFacturacionContingencia(dbConn *sql.DB, empresaID, id int64, usuario string) error {
	result, err := ExecCompat(dbConn, `UPDATE empresa_facturacion_contingencias c SET estado='cerrada',fecha_cierre=CURRENT_TIMESTAMP,
		usuario_cierre=?,fecha_actualizacion=CURRENT_TIMESTAMP
		WHERE c.empresa_id=? AND c.id=? AND c.estado='recuperada'
		AND NOT EXISTS (SELECT 1 FROM empresa_facturacion_contingencia_documentos d WHERE d.empresa_id=c.empresa_id AND d.contingencia_id=c.id AND d.estado_transmision<>'aceptado')`,
		strings.TrimSpace(usuario), empresaID, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return fmt.Errorf("contingencia inexistente, activa o con documentos pendientes")
	}
	return nil
}

func RegisterEmpresaFacturacionContingenciaDocumento(dbConn *sql.DB, empresaID int64, tipo, tipoDocumento, documentoCodigo string, retryID int64) (*EmpresaFacturacionContingenciaDocumento, error) {
	incident, err := GetActiveEmpresaFacturacionContingencia(dbConn, empresaID, tipo)
	if err != nil {
		return nil, err
	}
	_, err = ExecCompat(dbConn, `INSERT INTO empresa_facturacion_contingencia_documentos
		(empresa_id,contingencia_id,tipo_documento,documento_codigo,estado_transmision,retry_id)
		VALUES (?,?,?,?,'pendiente',?) ON CONFLICT (empresa_id,contingencia_id,tipo_documento,documento_codigo)
		DO UPDATE SET retry_id=GREATEST(empresa_facturacion_contingencia_documentos.retry_id,EXCLUDED.retry_id),fecha_actualizacion=CURRENT_TIMESTAMP`,
		empresaID, incident.ID, strings.TrimSpace(tipoDocumento), strings.TrimSpace(documentoCodigo), retryID)
	if err != nil {
		return nil, err
	}
	var item EmpresaFacturacionContingenciaDocumento
	err = queryRowSQLCompat(dbConn, `SELECT id,empresa_id,contingencia_id,COALESCE(configuracion_talonario_id,0),tipo_documento,documento_codigo,numero_papel,
		COALESCE(fecha_expedicion_papel::text,''),estado_transmision,retry_id,fecha_actualizacion::text
		FROM empresa_facturacion_contingencia_documentos WHERE empresa_id=? AND contingencia_id=? AND tipo_documento=? AND documento_codigo=?`,
		empresaID, incident.ID, strings.TrimSpace(tipoDocumento), strings.TrimSpace(documentoCodigo)).Scan(
		&item.ID, &item.EmpresaID, &item.ContingenciaID, &item.ConfiguracionID, &item.TipoDocumento, &item.DocumentoCodigo, &item.NumeroPapel,
		&item.FechaExpedicionPapel, &item.EstadoTransmision, &item.RetryID, &item.FechaActualizacion)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// RegisterEmpresaFacturacionTalonarioSale records a paper invoice genuinely
// issued during a facturer outage. It reserves the separately authorized paper
// number atomically and requires an immutable paid-sale source. It deliberately
// does not pretend to generate or transmit the later type-03 XML.
func RegisterEmpresaFacturacionTalonarioSale(dbConn *sql.DB, empresaID, contingenciaID, carritoID int64, comprobanteCodigo, numeroPapel, fechaPapel, usuario string) (*EmpresaFacturacionContingenciaDocumento, error) {
	if dbConn == nil || empresaID <= 0 || contingenciaID <= 0 || carritoID <= 0 {
		return nil, fmt.Errorf("empresa, contingencia y carrito son obligatorios")
	}
	comprobanteCodigo = strings.TrimSpace(comprobanteCodigo)
	numeroPapel = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(numeroPapel), " ", ""))
	fecha, err := time.Parse("2006-01-02", strings.TrimSpace(fechaPapel))
	if err != nil || comprobanteCodigo == "" || numeroPapel == "" {
		return nil, fmt.Errorf("comprobante, numero y fecha de papel son obligatorios")
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var incidentStart time.Time
	var incidentRecovery sql.NullTime
	var configurationID int64
	err = queryRowTxSQLCompat(tx, `SELECT fecha_inicio,fecha_recuperacion,configuracion_talonario_id FROM empresa_facturacion_contingencias
		WHERE empresa_id=? AND id=? AND tipo='falla_facturador' AND estado IN ('activa','recuperada') FOR UPDATE`, empresaID, contingenciaID).Scan(&incidentStart, &incidentRecovery, &configurationID)
	if err != nil {
		return nil, fmt.Errorf("contingencia del facturador no disponible: %w", err)
	}
	if fecha.Before(time.Date(incidentStart.Year(), incidentStart.Month(), incidentStart.Day(), 0, 0, 0, 0, incidentStart.Location())) ||
		(incidentRecovery.Valid && fecha.After(time.Date(incidentRecovery.Time.Year(), incidentRecovery.Time.Month(), incidentRecovery.Time.Day(), 0, 0, 0, 0, incidentRecovery.Time.Location()))) {
		return nil, fmt.Errorf("la fecha de papel no pertenece al periodo de la contingencia")
	}
	var prefix, state string
	var authFrom, authTo time.Time
	var next, end int64
	err = queryRowTxSQLCompat(tx, `SELECT prefijo,fecha_desde,fecha_hasta,proximo_numero,rango_hasta,estado
		FROM empresa_facturacion_contingencia_configuracion WHERE empresa_id=? AND id=? FOR UPDATE`, empresaID, configurationID).Scan(&prefix, &authFrom, &authTo, &next, &end, &state)
	if err != nil {
		return nil, fmt.Errorf("autorizacion de talonario no disponible: %w", err)
	}
	if (state != "activo" && state != "suspendido") || fecha.Before(authFrom) || fecha.After(authTo) || next > end {
		return nil, fmt.Errorf("autorizacion de talonario no vigente o agotada")
	}
	expected := strings.ToUpper(strings.TrimSpace(prefix)) + fmt.Sprintf("%d", next)
	if numeroPapel != expected {
		return nil, fmt.Errorf("numero de papel no coincide con el siguiente autorizado")
	}
	var amount float64
	err = queryRowTxSQLCompat(tx, `SELECT d.monto_total FROM empresa_facturacion_documentos d
		WHERE d.empresa_id=? AND d.tipo_documento='comprobante_pago' AND d.documento_codigo=?
		AND EXISTS (SELECT 1 FROM empresa_facturacion_artefactos a WHERE a.empresa_id=d.empresa_id AND a.tipo_documento=d.tipo_documento
			AND a.documento_codigo=d.documento_codigo AND a.tipo_artefacto=? AND a.estado='activo')
		AND EXISTS (SELECT 1 FROM carritos_compras c WHERE c.empresa_id=d.empresa_id AND c.id=? AND COALESCE(c.pagado_en,'')<>'' AND COALESCE(c.estado,'activo')='inactivo')`,
		empresaID, comprobanteCodigo, EmpresaFacturacionArtefactoTipoFuenteFiscalJSON, carritoID).Scan(&amount)
	if err != nil {
		return nil, fmt.Errorf("la venta pagada no tiene comprobante y fuente fiscal inmutable: %w", err)
	}
	result, err := execTxSQLCompat(tx, `INSERT INTO empresa_facturacion_contingencia_documentos
		(empresa_id,contingencia_id,configuracion_talonario_id,tipo_documento,documento_codigo,numero_papel,fecha_expedicion_papel,estado_transmision,retry_id,carrito_id,comprobante_codigo,monto_total)
		VALUES (?,?,?, 'factura_talonario_papel', ?, ?, ?::date, 'pendiente', 0, ?, ?, ?)
		ON CONFLICT (empresa_id,contingencia_id,tipo_documento,documento_codigo) DO NOTHING`,
		empresaID, contingenciaID, configurationID, comprobanteCodigo, numeroPapel, fechaPapel, carritoID, comprobanteCodigo, amount)
	if err != nil {
		return nil, err
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		return nil, fmt.Errorf("la venta de talonario ya fue registrada")
	}
	counterResult, err := execTxSQLCompat(tx, `UPDATE empresa_facturacion_contingencia_configuracion SET proximo_numero=proximo_numero+1,
		fecha_actualizacion=CURRENT_TIMESTAMP,usuario_creador=? WHERE empresa_id=? AND id=? AND proximo_numero=?`, strings.TrimSpace(usuario), empresaID, configurationID, next)
	if err != nil {
		return nil, err
	}
	counterRows, _ := counterResult.RowsAffected()
	if counterRows != 1 {
		return nil, fmt.Errorf("el consecutivo de talonario cambio durante el registro")
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	var item EmpresaFacturacionContingenciaDocumento
	err = queryRowSQLCompat(dbConn, `SELECT id,empresa_id,contingencia_id,COALESCE(configuracion_talonario_id,0),tipo_documento,documento_codigo,numero_papel,
		fecha_expedicion_papel::text,estado_transmision,retry_id,carrito_id,comprobante_codigo,monto_total,fecha_actualizacion::text
		FROM empresa_facturacion_contingencia_documentos WHERE empresa_id=? AND contingencia_id=? AND tipo_documento='factura_talonario_papel' AND documento_codigo=?`,
		empresaID, contingenciaID, comprobanteCodigo).Scan(&item.ID, &item.EmpresaID, &item.ContingenciaID,
		&item.ConfiguracionID, &item.TipoDocumento, &item.DocumentoCodigo, &item.NumeroPapel, &item.FechaExpedicionPapel, &item.EstadoTransmision, &item.RetryID,
		&item.CarritoID, &item.ComprobanteCodigo, &item.MontoTotal, &item.FechaActualizacion)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func SetEmpresaFacturacionContingenciaDocumentoEstado(dbConn *sql.DB, empresaID int64, tipoDocumento, documentoCodigo, estado string) error {
	estado = strings.ToLower(strings.TrimSpace(estado))
	if estado != "pendiente" && estado != "enviado" && estado != "aceptado" && estado != "rechazado" && estado != "bloqueado" {
		return fmt.Errorf("estado de transmision de contingencia invalido")
	}
	_, err := ExecCompat(dbConn, `UPDATE empresa_facturacion_contingencia_documentos SET estado_transmision=?,fecha_actualizacion=CURRENT_TIMESTAMP
		WHERE empresa_id=? AND tipo_documento=? AND documento_codigo=?`, estado, empresaID, strings.TrimSpace(tipoDocumento), strings.TrimSpace(documentoCodigo))
	return err
}
