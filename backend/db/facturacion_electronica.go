package db

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// FacturacionElectronicaPaisConfig define configuración FE por empresa y país.
type FacturacionElectronicaPaisConfig struct {
	ID                  int64  `json:"id"`
	EmpresaID           int64  `json:"empresa_id"`
	PaisCodigo          string `json:"pais_codigo"`
	PaisNombre          string `json:"pais_nombre"`
	BanderaPais         string `json:"bandera_pais,omitempty"`
	MonedaCodigo        string `json:"moneda_codigo,omitempty"`
	Proveedor           string `json:"proveedor,omitempty"`
	Ambiente            string `json:"ambiente,omitempty"`
	TipoDocumentoEmisor string `json:"tipo_documento_emisor,omitempty"`
	IdentificadorFiscal string `json:"identificador_fiscal,omitempty"`
	RazonSocial         string `json:"razon_social,omitempty"`
	EmailFacturacion    string `json:"email_facturacion,omitempty"`
	TelefonoFacturacion string `json:"telefono_facturacion,omitempty"`
	DireccionFiscal     string `json:"direccion_fiscal,omitempty"`
	PrefijoFactura      string `json:"prefijo_factura,omitempty"`
	ResolucionNumero    string `json:"resolucion_numero,omitempty"`
	APIBaseURL          string `json:"api_base_url,omitempty"`
	CamposPaisJSON      string `json:"campos_pais_json,omitempty"`
	FechaCreacion       string `json:"fecha_creacion,omitempty"`
	FechaActualizacion  string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador      string `json:"usuario_creador,omitempty"`
	Estado              string `json:"estado,omitempty"`
	Observaciones       string `json:"observaciones,omitempty"`
}

// PaisFacturacion representa un país soportado para FE.
type PaisFacturacion struct {
	Codigo  string `json:"codigo"`
	Nombre  string `json:"nombre"`
	Bandera string `json:"bandera"`
	Moneda  string `json:"moneda"`
}

// FacturacionDocumentoLegal representa los datos legales generados al emitir una factura.
type FacturacionDocumentoLegal struct {
	EmpresaID            int64  `json:"empresa_id"`
	PaisCodigo           string `json:"pais_codigo"`
	Ambiente             string `json:"ambiente"`
	TipoDocumentoEmisor  string `json:"tipo_documento_emisor"`
	IdentificadorFiscal  string `json:"identificador_fiscal"`
	RazonSocial          string `json:"razon_social"`
	PrefijoFactura       string `json:"prefijo_factura"`
	ResolucionNumero     string `json:"resolucion_numero"`
	ConsecutivoAsignado  int64  `json:"consecutivo_asignado"`
	NumeroLegal          string `json:"numero_legal"`
	CodigoValidacion     string `json:"codigo_validacion"`
	FechaEmisionLegal    string `json:"fecha_emision_legal"`
	ResolucionFechaDesde string `json:"resolucion_fecha_desde,omitempty"`
	ResolucionFechaHasta string `json:"resolucion_fecha_hasta,omitempty"`
}

func supportedPaisesFacturacionMap() map[string]PaisFacturacion {
	return map[string]PaisFacturacion{
		"CO": {Codigo: "CO", Nombre: "Colombia", Bandera: "🇨🇴", Moneda: "COP"},
		"PA": {Codigo: "PA", Nombre: "Panamá", Bandera: "🇵🇦", Moneda: "PAB"},
		"EC": {Codigo: "EC", Nombre: "Ecuador", Bandera: "🇪🇨", Moneda: "USD"},
	}
}

// ListPaisesFacturacionDisponibles retorna los países FE soportados.
func ListPaisesFacturacionDisponibles() []PaisFacturacion {
	catalog := supportedPaisesFacturacionMap()
	return []PaisFacturacion{catalog["CO"], catalog["PA"], catalog["EC"]}
}

func normalizePaisCodigo(v string) string {
	return strings.ToUpper(strings.TrimSpace(v))
}

func defaultPaisFacturacion() PaisFacturacion {
	return supportedPaisesFacturacionMap()["CO"]
}

func paisFacturacionByCodigo(codigo string) PaisFacturacion {
	catalog := supportedPaisesFacturacionMap()
	codigo = normalizePaisCodigo(codigo)
	if p, ok := catalog[codigo]; ok {
		return p
	}
	return defaultPaisFacturacion()
}

// EnsureEmpresaFacturacionElectronicaSchema crea/migra tabla FE por país en empresas.db.
func EnsureEmpresaFacturacionElectronicaSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS facturacion_electronica_pais (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			pais_codigo TEXT NOT NULL,
			pais_nombre TEXT NOT NULL,
			moneda_codigo TEXT,
			proveedor TEXT,
			ambiente TEXT DEFAULT 'sandbox',
			tipo_documento_emisor TEXT,
			identificador_fiscal TEXT,
			razon_social TEXT,
			email_facturacion TEXT,
			telefono_facturacion TEXT,
			direccion_fiscal TEXT,
			prefijo_factura TEXT,
			resolucion_numero TEXT,
			api_base_url TEXT,
			campos_pais_json TEXT DEFAULT '{}',
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, pais_codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_fe_pais_empresa ON facturacion_electronica_pais(empresa_id, pais_codigo);`,
		`CREATE INDEX IF NOT EXISTS ix_fe_pais_estado ON facturacion_electronica_pais(empresa_id, estado);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_pais", "empresa_id", "INTEGER NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_pais", "pais_codigo", "TEXT NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_pais", "pais_nombre", "TEXT NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_pais", "moneda_codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_pais", "proveedor", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_pais", "ambiente", "TEXT DEFAULT 'sandbox'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_pais", "tipo_documento_emisor", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_pais", "identificador_fiscal", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_pais", "razon_social", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_pais", "email_facturacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_pais", "telefono_facturacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_pais", "direccion_fiscal", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_pais", "prefijo_factura", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_pais", "resolucion_numero", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_pais", "api_base_url", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_pais", "campos_pais_json", "TEXT DEFAULT '{}' "); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_pais", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_pais", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_pais", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_pais", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

func defaultFacturacionConfig(empresaID int64, paisCodigo string) FacturacionElectronicaPaisConfig {
	pais := paisFacturacionByCodigo(paisCodigo)
	return FacturacionElectronicaPaisConfig{
		EmpresaID:      empresaID,
		PaisCodigo:     pais.Codigo,
		PaisNombre:     pais.Nombre,
		BanderaPais:    pais.Bandera,
		MonedaCodigo:   pais.Moneda,
		Ambiente:       "sandbox",
		Estado:         "activo",
		CamposPaisJSON: "{}",
	}
}

func hydrateFacturacionFromEmpresaConfig(dbConn *sql.DB, cfg *FacturacionElectronicaPaisConfig) error {
	if cfg == nil || cfg.EmpresaID <= 0 {
		return nil
	}

	var tipoDocumentoEmisor string
	var nit string
	var razonSocial string
	var emailFacturacion string
	var telefonoFacturacion string
	var direccionFiscal string
	var prefijoFactura string
	var resolucionNumero string
	var ambienteFE string

	err := dbConn.QueryRow(`SELECT
		COALESCE(tipo_documento_emisor, ''),
		COALESCE(nit, ''),
		COALESCE(razon_social, ''),
		COALESCE(email_facturacion, ''),
		COALESCE(telefono_facturacion, ''),
		COALESCE(direccion_fiscal, ''),
		COALESCE(prefijo_factura, ''),
		COALESCE(resolucion_numero, ''),
		COALESCE(ambiente_fe, 'habilitacion')
	FROM empresa_configuracion_avanzada
	WHERE empresa_id = ?
	LIMIT 1`, cfg.EmpresaID).Scan(
		&tipoDocumentoEmisor,
		&nit,
		&razonSocial,
		&emailFacturacion,
		&telefonoFacturacion,
		&direccionFiscal,
		&prefijoFactura,
		&resolucionNumero,
		&ambienteFE,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	if strings.TrimSpace(cfg.TipoDocumentoEmisor) == "" {
		cfg.TipoDocumentoEmisor = strings.TrimSpace(tipoDocumentoEmisor)
	}
	if strings.TrimSpace(cfg.IdentificadorFiscal) == "" {
		cfg.IdentificadorFiscal = strings.TrimSpace(nit)
	}
	if strings.TrimSpace(cfg.RazonSocial) == "" {
		cfg.RazonSocial = strings.TrimSpace(razonSocial)
	}
	if strings.TrimSpace(cfg.EmailFacturacion) == "" {
		cfg.EmailFacturacion = strings.TrimSpace(emailFacturacion)
	}
	if strings.TrimSpace(cfg.TelefonoFacturacion) == "" {
		cfg.TelefonoFacturacion = strings.TrimSpace(telefonoFacturacion)
	}
	if strings.TrimSpace(cfg.DireccionFiscal) == "" {
		cfg.DireccionFiscal = strings.TrimSpace(direccionFiscal)
	}
	if strings.TrimSpace(cfg.PrefijoFactura) == "" {
		cfg.PrefijoFactura = strings.TrimSpace(prefijoFactura)
	}
	if strings.TrimSpace(cfg.ResolucionNumero) == "" {
		cfg.ResolucionNumero = strings.TrimSpace(resolucionNumero)
	}

	ambienteFE = strings.ToLower(strings.TrimSpace(ambienteFE))
	if ambienteFE == "produccion" {
		cfg.Ambiente = "produccion"
	}

	return nil
}

func normalizeFacturacionConfig(payload *FacturacionElectronicaPaisConfig) {
	if payload == nil {
		return
	}
	pais := paisFacturacionByCodigo(payload.PaisCodigo)
	payload.PaisCodigo = pais.Codigo
	payload.PaisNombre = pais.Nombre
	payload.BanderaPais = pais.Bandera
	if strings.TrimSpace(payload.MonedaCodigo) == "" {
		payload.MonedaCodigo = pais.Moneda
	} else {
		payload.MonedaCodigo = strings.ToUpper(strings.TrimSpace(payload.MonedaCodigo))
	}
	payload.Proveedor = strings.TrimSpace(payload.Proveedor)
	if payload.Proveedor == "" {
		payload.Proveedor = "manual"
	}
	payload.Ambiente = strings.ToLower(strings.TrimSpace(payload.Ambiente))
	if payload.Ambiente != "produccion" {
		payload.Ambiente = "sandbox"
	}
	payload.TipoDocumentoEmisor = strings.TrimSpace(payload.TipoDocumentoEmisor)
	payload.IdentificadorFiscal = strings.TrimSpace(payload.IdentificadorFiscal)
	payload.RazonSocial = strings.TrimSpace(payload.RazonSocial)
	payload.EmailFacturacion = strings.TrimSpace(payload.EmailFacturacion)
	payload.TelefonoFacturacion = strings.TrimSpace(payload.TelefonoFacturacion)
	payload.DireccionFiscal = strings.TrimSpace(payload.DireccionFiscal)
	payload.PrefijoFactura = strings.TrimSpace(payload.PrefijoFactura)
	payload.ResolucionNumero = strings.TrimSpace(payload.ResolucionNumero)
	payload.APIBaseURL = strings.TrimSpace(payload.APIBaseURL)
	payload.Observaciones = strings.TrimSpace(payload.Observaciones)
	payload.UsuarioCreador = strings.TrimSpace(payload.UsuarioCreador)
	payload.Estado = strings.TrimSpace(strings.ToLower(payload.Estado))
	if payload.Estado != "inactivo" {
		payload.Estado = "activo"
	}

	payload.CamposPaisJSON = strings.TrimSpace(payload.CamposPaisJSON)
	if payload.CamposPaisJSON == "" {
		payload.CamposPaisJSON = "{}"
	} else {
		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(payload.CamposPaisJSON), &raw); err != nil {
			payload.CamposPaisJSON = "{}"
		}
	}
}

// UpsertFacturacionElectronicaPaisConfig crea o actualiza configuración por empresa/pais.
func UpsertFacturacionElectronicaPaisConfig(dbConn *sql.DB, payload FacturacionElectronicaPaisConfig) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	if normalizePaisCodigo(payload.PaisCodigo) == "" {
		return 0, fmt.Errorf("pais_codigo es obligatorio")
	}
	normalizeFacturacionConfig(&payload)

	stmt := `INSERT INTO facturacion_electronica_pais (
		empresa_id,
		pais_codigo,
		pais_nombre,
		moneda_codigo,
		proveedor,
		ambiente,
		tipo_documento_emisor,
		identificador_fiscal,
		razon_social,
		email_facturacion,
		telefono_facturacion,
		direccion_fiscal,
		prefijo_factura,
		resolucion_numero,
		api_base_url,
		campos_pais_json,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'), ?, ?, ?)
	ON CONFLICT(empresa_id, pais_codigo) DO UPDATE SET
		pais_nombre = excluded.pais_nombre,
		moneda_codigo = excluded.moneda_codigo,
		proveedor = excluded.proveedor,
		ambiente = excluded.ambiente,
		tipo_documento_emisor = excluded.tipo_documento_emisor,
		identificador_fiscal = excluded.identificador_fiscal,
		razon_social = excluded.razon_social,
		email_facturacion = excluded.email_facturacion,
		telefono_facturacion = excluded.telefono_facturacion,
		direccion_fiscal = excluded.direccion_fiscal,
		prefijo_factura = excluded.prefijo_factura,
		resolucion_numero = excluded.resolucion_numero,
		api_base_url = excluded.api_base_url,
		campos_pais_json = excluded.campos_pais_json,
		fecha_actualizacion = datetime('now','localtime'),
		usuario_creador = excluded.usuario_creador,
		estado = excluded.estado,
		observaciones = excluded.observaciones`

	if _, err := dbConn.Exec(stmt,
		payload.EmpresaID,
		payload.PaisCodigo,
		payload.PaisNombre,
		payload.MonedaCodigo,
		payload.Proveedor,
		payload.Ambiente,
		payload.TipoDocumentoEmisor,
		payload.IdentificadorFiscal,
		payload.RazonSocial,
		payload.EmailFacturacion,
		payload.TelefonoFacturacion,
		payload.DireccionFiscal,
		payload.PrefijoFactura,
		payload.ResolucionNumero,
		payload.APIBaseURL,
		payload.CamposPaisJSON,
		payload.UsuarioCreador,
		payload.Estado,
		payload.Observaciones,
	); err != nil {
		return 0, err
	}

	return getFacturacionElectronicaPaisID(dbConn, payload.EmpresaID, payload.PaisCodigo)
}

func getFacturacionElectronicaPaisID(dbConn *sql.DB, empresaID int64, paisCodigo string) (int64, error) {
	var id int64
	err := dbConn.QueryRow(`SELECT id FROM facturacion_electronica_pais WHERE empresa_id = ? AND pais_codigo = ? LIMIT 1`, empresaID, normalizePaisCodigo(paisCodigo)).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetFacturacionElectronicaPaisConfig obtiene configuración por empresa y país.
func GetFacturacionElectronicaPaisConfig(dbConn *sql.DB, empresaID int64, paisCodigo string) (*FacturacionElectronicaPaisConfig, error) {
	paisCodigo = normalizePaisCodigo(paisCodigo)
	if empresaID <= 0 || paisCodigo == "" {
		return nil, fmt.Errorf("empresa_id y pais_codigo son obligatorios")
	}

	cfg := defaultFacturacionConfig(empresaID, paisCodigo)
	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(pais_codigo, ''),
		COALESCE(pais_nombre, ''),
		COALESCE(moneda_codigo, ''),
		COALESCE(proveedor, ''),
		COALESCE(ambiente, 'sandbox'),
		COALESCE(tipo_documento_emisor, ''),
		COALESCE(identificador_fiscal, ''),
		COALESCE(razon_social, ''),
		COALESCE(email_facturacion, ''),
		COALESCE(telefono_facturacion, ''),
		COALESCE(direccion_fiscal, ''),
		COALESCE(prefijo_factura, ''),
		COALESCE(resolucion_numero, ''),
		COALESCE(api_base_url, ''),
		COALESCE(campos_pais_json, '{}'),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM facturacion_electronica_pais
	WHERE empresa_id = ? AND pais_codigo = ?
	LIMIT 1`, empresaID, paisCodigo)

	if err := row.Scan(
		&cfg.ID,
		&cfg.EmpresaID,
		&cfg.PaisCodigo,
		&cfg.PaisNombre,
		&cfg.MonedaCodigo,
		&cfg.Proveedor,
		&cfg.Ambiente,
		&cfg.TipoDocumentoEmisor,
		&cfg.IdentificadorFiscal,
		&cfg.RazonSocial,
		&cfg.EmailFacturacion,
		&cfg.TelefonoFacturacion,
		&cfg.DireccionFiscal,
		&cfg.PrefijoFactura,
		&cfg.ResolucionNumero,
		&cfg.APIBaseURL,
		&cfg.CamposPaisJSON,
		&cfg.FechaCreacion,
		&cfg.FechaActualizacion,
		&cfg.UsuarioCreador,
		&cfg.Estado,
		&cfg.Observaciones,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if hErr := hydrateFacturacionFromEmpresaConfig(dbConn, &cfg); hErr != nil {
				return nil, hErr
			}
			normalizeFacturacionConfig(&cfg)
			return &cfg, sql.ErrNoRows
		}
		return nil, err
	}

	normalizeFacturacionConfig(&cfg)
	return &cfg, nil
}

// ListFacturacionElectronicaPaisConfigs lista configuraciones FE por empresa.
func ListFacturacionElectronicaPaisConfigs(dbConn *sql.DB, empresaID int64, incluirInactivas bool) ([]FacturacionElectronicaPaisConfig, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	query := `SELECT
		id,
		empresa_id,
		COALESCE(pais_codigo, ''),
		COALESCE(pais_nombre, ''),
		COALESCE(moneda_codigo, ''),
		COALESCE(proveedor, ''),
		COALESCE(ambiente, 'sandbox'),
		COALESCE(tipo_documento_emisor, ''),
		COALESCE(identificador_fiscal, ''),
		COALESCE(razon_social, ''),
		COALESCE(email_facturacion, ''),
		COALESCE(telefono_facturacion, ''),
		COALESCE(direccion_fiscal, ''),
		COALESCE(prefijo_factura, ''),
		COALESCE(resolucion_numero, ''),
		COALESCE(api_base_url, ''),
		COALESCE(campos_pais_json, '{}'),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM facturacion_electronica_pais
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}
	if !incluirInactivas {
		query += " AND estado = 'activo'"
	}
	query += " ORDER BY pais_codigo ASC"

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]FacturacionElectronicaPaisConfig, 0)
	for rows.Next() {
		cfg := FacturacionElectronicaPaisConfig{}
		if err := rows.Scan(
			&cfg.ID,
			&cfg.EmpresaID,
			&cfg.PaisCodigo,
			&cfg.PaisNombre,
			&cfg.MonedaCodigo,
			&cfg.Proveedor,
			&cfg.Ambiente,
			&cfg.TipoDocumentoEmisor,
			&cfg.IdentificadorFiscal,
			&cfg.RazonSocial,
			&cfg.EmailFacturacion,
			&cfg.TelefonoFacturacion,
			&cfg.DireccionFiscal,
			&cfg.PrefijoFactura,
			&cfg.ResolucionNumero,
			&cfg.APIBaseURL,
			&cfg.CamposPaisJSON,
			&cfg.FechaCreacion,
			&cfg.FechaActualizacion,
			&cfg.UsuarioCreador,
			&cfg.Estado,
			&cfg.Observaciones,
		); err != nil {
			return nil, err
		}
		normalizeFacturacionConfig(&cfg)
		out = append(out, cfg)
	}
	return out, nil
}

func detectPaisByTimezone(tz string) string {
	tz = strings.ToLower(strings.TrimSpace(tz))
	switch {
	case strings.Contains(tz, "panama"):
		return "PA"
	case strings.Contains(tz, "guayaquil"), strings.Contains(tz, "quito"):
		return "EC"
	case strings.Contains(tz, "bogota"):
		return "CO"
	default:
		return ""
	}
}

func detectPaisByLanguage(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	switch {
	case strings.HasPrefix(lang, "es-pa"):
		return "PA"
	case strings.HasPrefix(lang, "es-ec"):
		return "EC"
	case strings.HasPrefix(lang, "es-co"):
		return "CO"
	default:
		return ""
	}
}

// DetectFacturacionPais determina país FE para una empresa usando configuración y señales del cliente.
func DetectFacturacionPais(dbConn *sql.DB, empresaID int64, timezone, language string) (PaisFacturacion, string, error) {
	if empresaID > 0 {
		var paisCfg sql.NullString
		err := dbConn.QueryRow(`SELECT COALESCE(pais_codigo, '') FROM empresa_configuracion_avanzada WHERE empresa_id = ? LIMIT 1`, empresaID).Scan(&paisCfg)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return PaisFacturacion{}, "", err
		}
		if paisCfg.Valid && normalizePaisCodigo(paisCfg.String) != "" {
			return paisFacturacionByCodigo(paisCfg.String), "configuracion_avanzada", nil
		}

		var paisFE sql.NullString
		err = dbConn.QueryRow(`SELECT COALESCE(pais_codigo, '') FROM facturacion_electronica_pais WHERE empresa_id = ? ORDER BY fecha_actualizacion DESC, id DESC LIMIT 1`, empresaID).Scan(&paisFE)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return PaisFacturacion{}, "", err
		}
		if paisFE.Valid && normalizePaisCodigo(paisFE.String) != "" {
			return paisFacturacionByCodigo(paisFE.String), "facturacion_electronica", nil
		}
	}

	if codigo := detectPaisByTimezone(timezone); codigo != "" {
		return paisFacturacionByCodigo(codigo), "timezone", nil
	}
	if codigo := detectPaisByLanguage(language); codigo != "" {
		return paisFacturacionByCodigo(codigo), "language", nil
	}
	return defaultPaisFacturacion(), "default", nil
}

func parseFechaISODate(raw string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(raw))
}

func normalizeAmbienteFEFromConfig(ambienteFE string) string {
	ambienteFE = strings.ToLower(strings.TrimSpace(ambienteFE))
	if ambienteFE == "produccion" {
		return "produccion"
	}
	return "sandbox"
}

func buildFacturaCodigoValidacion(empresaID int64, paisCodigo, documentoCodigo, numeroLegal string, montoTotal float64, moneda, identificadorFiscal, resolucionNumero, fechaEmision string) string {
	raw := fmt.Sprintf("%d|%s|%s|%s|%.2f|%s|%s|%s|%s",
		empresaID,
		strings.ToUpper(strings.TrimSpace(paisCodigo)),
		strings.ToUpper(strings.TrimSpace(documentoCodigo)),
		strings.ToUpper(strings.TrimSpace(numeroLegal)),
		montoTotal,
		strings.ToUpper(strings.TrimSpace(moneda)),
		strings.TrimSpace(identificadorFiscal),
		strings.TrimSpace(resolucionNumero),
		strings.TrimSpace(fechaEmision),
	)
	sum := sha256.Sum256([]byte(raw))
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}

// PrepareFacturacionDocumentoLegal valida cumplimiento y reserva consecutivo para emisión legal.
func PrepareFacturacionDocumentoLegal(dbConn *sql.DB, empresaID int64, paisCodigo, documentoCodigo string, montoTotal float64, moneda string) (*FacturacionDocumentoLegal, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	documentoCodigo = strings.ToUpper(strings.TrimSpace(documentoCodigo))
	if documentoCodigo == "" {
		return nil, fmt.Errorf("documento_codigo es obligatorio")
	}
	if montoTotal < 0 {
		montoTotal = 0
	}
	moneda = strings.ToUpper(strings.TrimSpace(moneda))

	paisCodigo = normalizePaisCodigo(paisCodigo)
	if paisCodigo == "" {
		paisDetectado, _, err := DetectFacturacionPais(dbConn, empresaID, "", "")
		if err != nil {
			return nil, err
		}
		paisCodigo = paisDetectado.Codigo
	}

	cfg, err := GetFacturacionElectronicaPaisConfig(dbConn, empresaID, paisCodigo)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if cfg == nil {
		return nil, fmt.Errorf("no existe configuracion de facturacion electronica para el pais solicitado")
	}
	if strings.ToLower(strings.TrimSpace(cfg.Estado)) == "inactivo" {
		return nil, fmt.Errorf("la configuracion de facturacion electronica esta inactiva para %s", cfg.PaisCodigo)
	}

	tx, err := dbConn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var tipoDocumentoEmisor string
	var nit string
	var razonSocial string
	var ambienteFE string
	var prefijoFactura string
	var resolucionNumero string
	var resolucionFechaDesde string
	var resolucionFechaHasta string
	var consecutivoDesde int64
	var consecutivoHasta int64
	var proximoConsecutivo int64

	err = tx.QueryRow(`SELECT
		COALESCE(tipo_documento_emisor, ''),
		COALESCE(nit, ''),
		COALESCE(razon_social, ''),
		COALESCE(ambiente_fe, 'habilitacion'),
		COALESCE(prefijo_factura, ''),
		COALESCE(resolucion_numero, ''),
		COALESCE(resolucion_fecha_desde, ''),
		COALESCE(resolucion_fecha_hasta, ''),
		COALESCE(consecutivo_desde, 1),
		COALESCE(consecutivo_hasta, 999999),
		COALESCE(proximo_consecutivo, 1)
	FROM empresa_configuracion_avanzada
	WHERE empresa_id = ?
	LIMIT 1`, empresaID).Scan(
		&tipoDocumentoEmisor,
		&nit,
		&razonSocial,
		&ambienteFE,
		&prefijoFactura,
		&resolucionNumero,
		&resolucionFechaDesde,
		&resolucionFechaHasta,
		&consecutivoDesde,
		&consecutivoHasta,
		&proximoConsecutivo,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("debe configurar empresa_configuracion_avanzada antes de emitir legalmente")
		}
		return nil, err
	}

	if strings.TrimSpace(cfg.TipoDocumentoEmisor) == "" {
		cfg.TipoDocumentoEmisor = strings.TrimSpace(tipoDocumentoEmisor)
	}
	if strings.TrimSpace(cfg.IdentificadorFiscal) == "" {
		cfg.IdentificadorFiscal = strings.TrimSpace(nit)
	}
	if strings.TrimSpace(cfg.RazonSocial) == "" {
		cfg.RazonSocial = strings.TrimSpace(razonSocial)
	}
	if strings.TrimSpace(cfg.PrefijoFactura) == "" {
		cfg.PrefijoFactura = strings.TrimSpace(prefijoFactura)
	}
	if strings.TrimSpace(cfg.ResolucionNumero) == "" {
		cfg.ResolucionNumero = strings.TrimSpace(resolucionNumero)
	}
	if strings.TrimSpace(cfg.MonedaCodigo) == "" {
		cfg.MonedaCodigo = strings.ToUpper(strings.TrimSpace(moneda))
	}
	if strings.TrimSpace(cfg.Ambiente) == "" {
		cfg.Ambiente = normalizeAmbienteFEFromConfig(ambienteFE)
	}

	if strings.TrimSpace(cfg.TipoDocumentoEmisor) == "" {
		return nil, fmt.Errorf("falta tipo_documento_emisor en configuracion de facturacion")
	}
	if strings.TrimSpace(cfg.IdentificadorFiscal) == "" {
		return nil, fmt.Errorf("falta identificador_fiscal en configuracion de facturacion")
	}
	if strings.TrimSpace(cfg.RazonSocial) == "" {
		return nil, fmt.Errorf("falta razon_social en configuracion de facturacion")
	}
	if strings.TrimSpace(cfg.PrefijoFactura) == "" {
		return nil, fmt.Errorf("falta prefijo_factura en configuracion de facturacion")
	}
	if strings.TrimSpace(cfg.ResolucionNumero) == "" {
		return nil, fmt.Errorf("falta resolucion_numero en configuracion de facturacion")
	}

	now := time.Now().In(time.Local)
	fechaHoy := now.Format("2006-01-02")
	if strings.TrimSpace(resolucionFechaDesde) != "" {
		fechaDesde, err := parseFechaISODate(resolucionFechaDesde)
		if err != nil {
			return nil, fmt.Errorf("resolucion_fecha_desde invalida")
		}
		if fechaHoy < fechaDesde.Format("2006-01-02") {
			return nil, fmt.Errorf("la resolucion de facturacion aun no inicia vigencia")
		}
	}
	if strings.TrimSpace(resolucionFechaHasta) != "" {
		fechaHasta, err := parseFechaISODate(resolucionFechaHasta)
		if err != nil {
			return nil, fmt.Errorf("resolucion_fecha_hasta invalida")
		}
		if fechaHoy > fechaHasta.Format("2006-01-02") {
			return nil, fmt.Errorf("la resolucion de facturacion esta vencida")
		}
	}

	if consecutivoDesde <= 0 {
		consecutivoDesde = 1
	}
	if consecutivoHasta < consecutivoDesde {
		return nil, fmt.Errorf("rango de consecutivos invalido")
	}
	if proximoConsecutivo < consecutivoDesde {
		proximoConsecutivo = consecutivoDesde
	}
	if proximoConsecutivo > consecutivoHasta {
		return nil, fmt.Errorf("rango de consecutivos agotado para facturacion")
	}

	if _, err := tx.Exec(`UPDATE empresa_configuracion_avanzada
		SET proximo_consecutivo = ?,
			fecha_actualizacion = datetime('now','localtime')
		WHERE empresa_id = ?`, proximoConsecutivo+1, empresaID); err != nil {
		return nil, err
	}

	prefix := strings.ToUpper(strings.TrimSpace(cfg.PrefijoFactura))
	prefix = strings.ReplaceAll(prefix, " ", "")
	numeroLegal := fmt.Sprintf("%s-%d", prefix, proximoConsecutivo)
	fechaEmisionLegal := now.Format("2006-01-02 15:04:05")
	if moneda == "" {
		moneda = strings.ToUpper(strings.TrimSpace(cfg.MonedaCodigo))
	}
	if moneda == "" {
		moneda = "COP"
	}
	codigoValidacion := buildFacturaCodigoValidacion(
		empresaID,
		cfg.PaisCodigo,
		documentoCodigo,
		numeroLegal,
		montoTotal,
		moneda,
		cfg.IdentificadorFiscal,
		cfg.ResolucionNumero,
		fechaEmisionLegal,
	)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &FacturacionDocumentoLegal{
		EmpresaID:            empresaID,
		PaisCodigo:           strings.ToUpper(strings.TrimSpace(cfg.PaisCodigo)),
		Ambiente:             normalizeAmbienteFEFromConfig(cfg.Ambiente),
		TipoDocumentoEmisor:  strings.TrimSpace(cfg.TipoDocumentoEmisor),
		IdentificadorFiscal:  strings.TrimSpace(cfg.IdentificadorFiscal),
		RazonSocial:          strings.TrimSpace(cfg.RazonSocial),
		PrefijoFactura:       prefix,
		ResolucionNumero:     strings.TrimSpace(cfg.ResolucionNumero),
		ConsecutivoAsignado:  proximoConsecutivo,
		NumeroLegal:          numeroLegal,
		CodigoValidacion:     codigoValidacion,
		FechaEmisionLegal:    fechaEmisionLegal,
		ResolucionFechaDesde: strings.TrimSpace(resolucionFechaDesde),
		ResolucionFechaHasta: strings.TrimSpace(resolucionFechaHasta),
	}, nil
}
