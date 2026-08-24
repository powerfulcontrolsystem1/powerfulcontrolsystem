package db

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

const empresaDIANDocumentosConfiguracionFingerprint = "empresa-dian-documentos-configuracion:v1:tenant-document-type-numbering"

// EmpresaDIANDocumentosConfiguracionSchemaReady confirms that pcs-migrate ran
// before an API tries to read or change document-family configuration.
func EmpresaDIANDocumentosConfiguracionSchemaReady(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("db connection is nil")
	}
	var table sql.NullString
	if err := QueryRowCompat(dbConn, `SELECT to_regclass(?)`, "empresa_dian_documentos_configuracion").Scan(&table); err != nil {
		return fmt.Errorf("verify DIAN document configuration table: %w", err)
	}
	if !table.Valid || strings.TrimSpace(table.String) == "" {
		return fmt.Errorf("DIAN document configuration table is missing; run pcs-migrate before starting the API")
	}
	return nil
}

// EmpresaDIANDocumentoConfiguracion keeps each DIAN document family separate
// from the invoice range. A DIAN authorization for invoices must never be
// reused for purchase support, payroll, equivalent documents, or RADIAN.
type EmpresaDIANDocumentoConfiguracion struct {
	ID                   int64  `json:"id"`
	EmpresaID            int64  `json:"empresa_id"`
	TipoDocumento        string `json:"tipo_documento"`
	Estado               string `json:"estado"`
	TipoAmbiente         string `json:"tipo_ambiente"`
	ModoOperacionCodigo  string `json:"modo_operacion_codigo,omitempty"`
	TestSetID            string `json:"test_set_id,omitempty"`
	Prefijo              string `json:"prefijo,omitempty"`
	ResolucionNumero     string `json:"resolucion_numero,omitempty"`
	ResolucionFechaDesde string `json:"resolucion_fecha_desde,omitempty"`
	ResolucionFechaHasta string `json:"resolucion_fecha_hasta,omitempty"`
	RangoDesde           int64  `json:"rango_desde"`
	RangoHasta           int64  `json:"rango_hasta"`
	ConsecutivoActual    int64  `json:"consecutivo_actual"`
	URLDIANOverride      string `json:"url_dian_override,omitempty"`
	Observaciones        string `json:"observaciones,omitempty"`
	UsuarioCreador       string `json:"usuario_creador,omitempty"`
	FechaCreacion        string `json:"fecha_creacion,omitempty"`
	FechaActualizacion   string `json:"fecha_actualizacion,omitempty"`
}

func applyEmpresaDIANDocumentosConfiguracionTx(ctx context.Context, tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	for _, statement := range []string{
		`CREATE TABLE empresa_dian_documentos_configuracion (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			tipo_documento TEXT NOT NULL,
			estado TEXT NOT NULL DEFAULT 'inactivo',
			tipo_ambiente TEXT NOT NULL DEFAULT 'habilitacion',
			modo_operacion_codigo TEXT NOT NULL DEFAULT '',
			test_set_id TEXT NOT NULL DEFAULT '',
			prefijo TEXT NOT NULL DEFAULT '',
			resolucion_numero TEXT NOT NULL DEFAULT '',
			resolucion_fecha_desde DATE,
			resolucion_fecha_hasta DATE,
			rango_desde BIGINT NOT NULL DEFAULT 0,
			rango_hasta BIGINT NOT NULL DEFAULT 0,
			consecutivo_actual BIGINT NOT NULL DEFAULT 0,
			url_dian_override TEXT NOT NULL DEFAULT '',
			observaciones TEXT NOT NULL DEFAULT '',
			usuario_creador TEXT NOT NULL DEFAULT '',
			fecha_creacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT uq_empresa_dian_documento_configuracion UNIQUE (empresa_id, tipo_documento),
			CONSTRAINT ck_empresa_dian_documento_tipo CHECK (tipo_documento ~ '^[a-z0-9_]{3,120}$'),
			CONSTRAINT ck_empresa_dian_documento_estado CHECK (estado IN ('inactivo', 'configurando', 'habilitacion', 'activo', 'suspendido')),
			CONSTRAINT ck_empresa_dian_documento_ambiente CHECK (tipo_ambiente IN ('habilitacion', 'produccion')),
			CONSTRAINT ck_empresa_dian_documento_rango CHECK (rango_desde >= 0 AND rango_hasta >= rango_desde AND consecutivo_actual >= 0)
		)`,
		`CREATE INDEX ix_empresa_dian_documentos_configuracion_estado
			ON empresa_dian_documentos_configuracion (empresa_id, tipo_documento, estado)`,
	} {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("migrate empresa_dian_documentos_configuracion: %w", err)
		}
	}
	return nil
}

func normalizeEmpresaDIANDocumentoConfiguracion(item *EmpresaDIANDocumentoConfiguracion) error {
	if item == nil || item.EmpresaID <= 0 {
		return fmt.Errorf("empresa_id es obligatorio")
	}
	item.TipoDocumento = strings.ToLower(strings.TrimSpace(item.TipoDocumento))
	item.Estado = strings.ToLower(strings.TrimSpace(item.Estado))
	item.TipoAmbiente = strings.ToLower(strings.TrimSpace(item.TipoAmbiente))
	item.ModoOperacionCodigo = strings.TrimSpace(item.ModoOperacionCodigo)
	item.TestSetID = strings.TrimSpace(item.TestSetID)
	item.Prefijo = strings.ToUpper(strings.TrimSpace(item.Prefijo))
	item.ResolucionNumero = strings.TrimSpace(item.ResolucionNumero)
	item.ResolucionFechaDesde = strings.TrimSpace(item.ResolucionFechaDesde)
	item.ResolucionFechaHasta = strings.TrimSpace(item.ResolucionFechaHasta)
	item.URLDIANOverride = strings.TrimSpace(item.URLDIANOverride)
	item.Observaciones = strings.TrimSpace(item.Observaciones)
	item.UsuarioCreador = strings.TrimSpace(item.UsuarioCreador)
	if item.Estado == "" {
		item.Estado = "inactivo"
	}
	if item.TipoAmbiente == "" {
		item.TipoAmbiente = "habilitacion"
	}
	if item.TipoDocumento == "" || len(item.TipoDocumento) > 120 {
		return fmt.Errorf("tipo_documento es obligatorio")
	}
	for _, r := range item.TipoDocumento {
		if !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && r != '_' {
			return fmt.Errorf("tipo_documento invalido")
		}
	}
	switch item.Estado {
	case "inactivo", "configurando", "habilitacion", "activo", "suspendido":
	default:
		return fmt.Errorf("estado de documento DIAN invalido")
	}
	if item.TipoAmbiente != "habilitacion" && item.TipoAmbiente != "produccion" {
		return fmt.Errorf("tipo_ambiente de documento DIAN invalido")
	}
	if item.RangoDesde < 0 || item.RangoHasta < item.RangoDesde || item.ConsecutivoActual < 0 {
		return fmt.Errorf("rango o consecutivo de documento DIAN invalido")
	}
	for field, value := range map[string]struct {
		value   string
		maximum int
	}{
		"modo_operacion_codigo": {item.ModoOperacionCodigo, 120},
		"test_set_id":           {item.TestSetID, 300},
		"prefijo":               {item.Prefijo, 64},
		"resolucion_numero":     {item.ResolucionNumero, 120},
		"url_dian_override":     {item.URLDIANOverride, 2048},
		"observaciones":         {item.Observaciones, 2000},
		"usuario_creador":       {item.UsuarioCreador, 320},
	} {
		if len(value.value) > value.maximum {
			return fmt.Errorf("%s supera la longitud permitida", field)
		}
	}
	for field, value := range map[string]string{
		"resolucion_fecha_desde": strings.TrimSpace(item.ResolucionFechaDesde),
		"resolucion_fecha_hasta": strings.TrimSpace(item.ResolucionFechaHasta),
	} {
		if value == "" {
			continue
		}
		if _, err := time.Parse("2006-01-02", value); err != nil {
			return fmt.Errorf("%s invalida", field)
		}
	}
	if item.ResolucionFechaDesde != "" && item.ResolucionFechaHasta != "" && item.ResolucionFechaHasta < item.ResolucionFechaDesde {
		return fmt.Errorf("rango de vigencia de resolucion invalido")
	}
	if err := validateDIANDocumentoURLOverride(item.URLDIANOverride); err != nil {
		return err
	}
	return nil
}

func validateDIANDocumentoURLOverride(raw string) error {
	if raw == "" {
		return nil
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
		return fmt.Errorf("url_dian_override invalida")
	}
	host := parsed.Hostname()
	if host == "" || strings.EqualFold(host, "localhost") {
		return fmt.Errorf("url_dian_override no permite destinos locales")
	}
	if ip := net.ParseIP(host); ip != nil && (ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsUnspecified()) {
		return fmt.Errorf("url_dian_override no permite destinos privados")
	}
	return nil
}

// GetEmpresaDIANDocumentoConfiguracionContext reads only the configuration of
// the requested tenant and document family. A missing row is not synthesized:
// emitting code must treat it as an explicit, incomplete configuration.
func GetEmpresaDIANDocumentoConfiguracionContext(ctx context.Context, dbConn *sql.DB, empresaID int64, tipoDocumento string) (*EmpresaDIANDocumentoConfiguracion, error) {
	if dbConn == nil {
		return nil, fmt.Errorf("db connection is nil")
	}
	lookup := &EmpresaDIANDocumentoConfiguracion{EmpresaID: empresaID, TipoDocumento: tipoDocumento}
	if err := normalizeEmpresaDIANDocumentoConfiguracion(lookup); err != nil {
		return nil, err
	}
	var out EmpresaDIANDocumentoConfiguracion
	err := QueryRowCompatContext(ctx, dbConn, `SELECT
		id, empresa_id, tipo_documento, estado, tipo_ambiente,
		COALESCE(modo_operacion_codigo, ''), COALESCE(test_set_id, ''),
		COALESCE(prefijo, ''), COALESCE(resolucion_numero, ''),
		COALESCE(resolucion_fecha_desde::TEXT, ''), COALESCE(resolucion_fecha_hasta::TEXT, ''),
		COALESCE(rango_desde, 0), COALESCE(rango_hasta, 0), COALESCE(consecutivo_actual, 0),
		COALESCE(url_dian_override, ''), COALESCE(observaciones, ''),
		COALESCE(usuario_creador, ''), COALESCE(fecha_creacion::TEXT, ''), COALESCE(fecha_actualizacion::TEXT, '')
	FROM empresa_dian_documentos_configuracion
	WHERE empresa_id = ? AND tipo_documento = ?
	LIMIT 1`, lookup.EmpresaID, lookup.TipoDocumento).Scan(
		&out.ID, &out.EmpresaID, &out.TipoDocumento, &out.Estado, &out.TipoAmbiente,
		&out.ModoOperacionCodigo, &out.TestSetID, &out.Prefijo, &out.ResolucionNumero,
		&out.ResolucionFechaDesde, &out.ResolucionFechaHasta, &out.RangoDesde, &out.RangoHasta,
		&out.ConsecutivoActual, &out.URLDIANOverride, &out.Observaciones, &out.UsuarioCreador,
		&out.FechaCreacion, &out.FechaActualizacion,
	)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ListEmpresaDIANDocumentosConfiguracionContext lists configuration records
// scoped to a single tenant. It never reads another company's DIAN data.
func ListEmpresaDIANDocumentosConfiguracionContext(ctx context.Context, dbConn *sql.DB, empresaID int64) ([]EmpresaDIANDocumentoConfiguracion, error) {
	if dbConn == nil {
		return nil, fmt.Errorf("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	rows, err := querySQLCompatContext(ctx, dbConn, `SELECT
		id, empresa_id, tipo_documento, estado, tipo_ambiente,
		COALESCE(modo_operacion_codigo, ''), COALESCE(test_set_id, ''),
		COALESCE(prefijo, ''), COALESCE(resolucion_numero, ''),
		COALESCE(resolucion_fecha_desde::TEXT, ''), COALESCE(resolucion_fecha_hasta::TEXT, ''),
		COALESCE(rango_desde, 0), COALESCE(rango_hasta, 0), COALESCE(consecutivo_actual, 0),
		COALESCE(url_dian_override, ''), COALESCE(observaciones, ''),
		COALESCE(usuario_creador, ''), COALESCE(fecha_creacion::TEXT, ''), COALESCE(fecha_actualizacion::TEXT, '')
	FROM empresa_dian_documentos_configuracion
	WHERE empresa_id = ?
	ORDER BY tipo_documento ASC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]EmpresaDIANDocumentoConfiguracion, 0)
	for rows.Next() {
		var item EmpresaDIANDocumentoConfiguracion
		if err := rows.Scan(
			&item.ID, &item.EmpresaID, &item.TipoDocumento, &item.Estado, &item.TipoAmbiente,
			&item.ModoOperacionCodigo, &item.TestSetID, &item.Prefijo, &item.ResolucionNumero,
			&item.ResolucionFechaDesde, &item.ResolucionFechaHasta, &item.RangoDesde, &item.RangoHasta,
			&item.ConsecutivoActual, &item.URLDIANOverride, &item.Observaciones, &item.UsuarioCreador,
			&item.FechaCreacion, &item.FechaActualizacion,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// UpsertEmpresaDIANDocumentoConfiguracionContext writes one configuration per
// tenant and document family. The unique key makes an administrative retry
// idempotent and does not allow an update to cross empresa_id boundaries.
func UpsertEmpresaDIANDocumentoConfiguracionContext(ctx context.Context, dbConn *sql.DB, item EmpresaDIANDocumentoConfiguracion) (int64, error) {
	if dbConn == nil {
		return 0, fmt.Errorf("db connection is nil")
	}
	if err := normalizeEmpresaDIANDocumentoConfiguracion(&item); err != nil {
		return 0, err
	}
	var id int64
	err := QueryRowCompatContext(ctx, dbConn, `INSERT INTO empresa_dian_documentos_configuracion (
		empresa_id, tipo_documento, estado, tipo_ambiente, modo_operacion_codigo, test_set_id,
		prefijo, resolucion_numero, resolucion_fecha_desde, resolucion_fecha_hasta,
		rango_desde, rango_hasta, consecutivo_actual, url_dian_override, observaciones,
		usuario_creador, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?, '')::DATE, NULLIF(?, '')::DATE, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT (empresa_id, tipo_documento) DO UPDATE SET
		estado = EXCLUDED.estado,
		tipo_ambiente = EXCLUDED.tipo_ambiente,
		modo_operacion_codigo = EXCLUDED.modo_operacion_codigo,
		test_set_id = EXCLUDED.test_set_id,
		prefijo = EXCLUDED.prefijo,
		resolucion_numero = EXCLUDED.resolucion_numero,
		resolucion_fecha_desde = EXCLUDED.resolucion_fecha_desde,
		resolucion_fecha_hasta = EXCLUDED.resolucion_fecha_hasta,
		rango_desde = EXCLUDED.rango_desde,
		rango_hasta = EXCLUDED.rango_hasta,
		consecutivo_actual = EXCLUDED.consecutivo_actual,
		url_dian_override = EXCLUDED.url_dian_override,
		observaciones = EXCLUDED.observaciones,
		usuario_creador = EXCLUDED.usuario_creador,
		fecha_actualizacion = CURRENT_TIMESTAMP
	RETURNING id`,
		item.EmpresaID, item.TipoDocumento, item.Estado, item.TipoAmbiente, item.ModoOperacionCodigo, item.TestSetID,
		item.Prefijo, item.ResolucionNumero, item.ResolucionFechaDesde, item.ResolucionFechaHasta,
		item.RangoDesde, item.RangoHasta, item.ConsecutivoActual, item.URLDIANOverride, item.Observaciones,
		item.UsuarioCreador,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
