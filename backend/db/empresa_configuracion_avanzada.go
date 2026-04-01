package db

import (
	"database/sql"
	"fmt"
	"strings"
)

// EmpresaConfiguracionAvanzada almacena la configuración empresarial y fiscal
// necesaria para preparar facturación electrónica en Colombia por empresa.
type EmpresaConfiguracionAvanzada struct {
	ID                        int64  `json:"id"`
	EmpresaID                 int64  `json:"empresa_id"`
	TipoDocumentoEmisor       string `json:"tipo_documento_emisor,omitempty"`
	NIT                       string `json:"nit,omitempty"`
	DigitoVerificacion        string `json:"digito_verificacion,omitempty"`
	RazonSocial               string `json:"razon_social,omitempty"`
	NombreComercial           string `json:"nombre_comercial,omitempty"`
	RegimenFiscal             string `json:"regimen_fiscal,omitempty"`
	ResponsabilidadTributaria string `json:"responsabilidad_tributaria,omitempty"`
	EmailFacturacion          string `json:"email_facturacion,omitempty"`
	TelefonoFacturacion       string `json:"telefono_facturacion,omitempty"`
	DireccionFiscal           string `json:"direccion_fiscal,omitempty"`
	Departamento              string `json:"departamento,omitempty"`
	Municipio                 string `json:"municipio,omitempty"`
	PaisCodigo                string `json:"pais_codigo,omitempty"`
	CodigoPostal              string `json:"codigo_postal,omitempty"`
	AmbienteFE                string `json:"ambiente_fe,omitempty"`
	TipoOperacion             string `json:"tipo_operacion,omitempty"`
	PrefijoFactura            string `json:"prefijo_factura,omitempty"`
	ResolucionNumero          string `json:"resolucion_numero,omitempty"`
	ResolucionFechaDesde      string `json:"resolucion_fecha_desde,omitempty"`
	ResolucionFechaHasta      string `json:"resolucion_fecha_hasta,omitempty"`
	ConsecutivoDesde          int64  `json:"consecutivo_desde,omitempty"`
	ConsecutivoHasta          int64  `json:"consecutivo_hasta,omitempty"`
	ProximoConsecutivo        int64  `json:"proximo_consecutivo,omitempty"`
	FormatoImpresion          string `json:"formato_impresion,omitempty"`
	MostrarLogo               bool   `json:"mostrar_logo"`
	LogoURL                   string `json:"logo_url,omitempty"`
	PieFactura                string `json:"pie_factura,omitempty"`
	NotasLegales              string `json:"notas_legales,omitempty"`
	FechaCreacion             string `json:"fecha_creacion,omitempty"`
	FechaActualizacion        string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador            string `json:"usuario_creador,omitempty"`
	Estado                    string `json:"estado,omitempty"`
	Observaciones             string `json:"observaciones,omitempty"`
}

// EnsureEmpresaConfiguracionAvanzadaSchema crea/migra el esquema de configuración avanzada
// por empresa para preparación de facturación electrónica en Colombia.
func EnsureEmpresaConfiguracionAvanzadaSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_configuracion_avanzada (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL UNIQUE,
			tipo_documento_emisor TEXT DEFAULT 'NIT',
			nit TEXT,
			digito_verificacion TEXT,
			razon_social TEXT,
			nombre_comercial TEXT,
			regimen_fiscal TEXT,
			responsabilidad_tributaria TEXT,
			email_facturacion TEXT,
			telefono_facturacion TEXT,
			direccion_fiscal TEXT,
			departamento TEXT,
			municipio TEXT,
			pais_codigo TEXT DEFAULT 'CO',
			codigo_postal TEXT,
			ambiente_fe TEXT DEFAULT 'habilitacion',
			tipo_operacion TEXT DEFAULT '10',
			prefijo_factura TEXT,
			resolucion_numero TEXT,
			resolucion_fecha_desde TEXT,
			resolucion_fecha_hasta TEXT,
			consecutivo_desde INTEGER DEFAULT 1,
			consecutivo_hasta INTEGER DEFAULT 999999,
			proximo_consecutivo INTEGER DEFAULT 1,
			formato_impresion TEXT DEFAULT 'carta',
			mostrar_logo INTEGER DEFAULT 1,
			logo_url TEXT,
			pie_factura TEXT,
			notas_legales TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_config_avanzada_empresa ON empresa_configuracion_avanzada(empresa_id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_config_avanzada_estado ON empresa_configuracion_avanzada(empresa_id, estado);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "tipo_documento_emisor", "TEXT DEFAULT 'NIT'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "nit", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "digito_verificacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "razon_social", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "nombre_comercial", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "regimen_fiscal", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "responsabilidad_tributaria", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "email_facturacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "telefono_facturacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "direccion_fiscal", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "departamento", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "municipio", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "pais_codigo", "TEXT DEFAULT 'CO'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "codigo_postal", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "ambiente_fe", "TEXT DEFAULT 'habilitacion'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "tipo_operacion", "TEXT DEFAULT '10'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "prefijo_factura", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "resolucion_numero", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "resolucion_fecha_desde", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "resolucion_fecha_hasta", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "consecutivo_desde", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "consecutivo_hasta", "INTEGER DEFAULT 999999"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "proximo_consecutivo", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "formato_impresion", "TEXT DEFAULT 'carta'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "mostrar_logo", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "logo_url", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "pie_factura", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "notas_legales", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "observaciones", "TEXT"); err != nil {
		return err
	}
	return nil
}

func defaultConfigAvanzada(empresaID int64) EmpresaConfiguracionAvanzada {
	return EmpresaConfiguracionAvanzada{
		EmpresaID:           empresaID,
		TipoDocumentoEmisor: "NIT",
		PaisCodigo:          "CO",
		AmbienteFE:          "habilitacion",
		TipoOperacion:       "10",
		ConsecutivoDesde:    1,
		ConsecutivoHasta:    999999,
		ProximoConsecutivo:  1,
		FormatoImpresion:    "carta",
		MostrarLogo:         true,
		Estado:              "activo",
	}
}

func defaultFormatoImpresion(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "pos" {
		return "pos"
	}
	return "carta"
}

func defaultAmbienteFE(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "produccion" {
		return "produccion"
	}
	return "habilitacion"
}

func defaultTipoOperacion(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "10"
	}
	return v
}

// GetEmpresaConfiguracionAvanzada obtiene la configuración avanzada por empresa.
// Si no existe registro, retorna valores por defecto para facilitar captura inicial.
func GetEmpresaConfiguracionAvanzada(dbConn *sql.DB, empresaID int64) (*EmpresaConfiguracionAvanzada, error) {
	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(tipo_documento_emisor, 'NIT'),
		COALESCE(nit, ''),
		COALESCE(digito_verificacion, ''),
		COALESCE(razon_social, ''),
		COALESCE(nombre_comercial, ''),
		COALESCE(regimen_fiscal, ''),
		COALESCE(responsabilidad_tributaria, ''),
		COALESCE(email_facturacion, ''),
		COALESCE(telefono_facturacion, ''),
		COALESCE(direccion_fiscal, ''),
		COALESCE(departamento, ''),
		COALESCE(municipio, ''),
		COALESCE(pais_codigo, 'CO'),
		COALESCE(codigo_postal, ''),
		COALESCE(ambiente_fe, 'habilitacion'),
		COALESCE(tipo_operacion, '10'),
		COALESCE(prefijo_factura, ''),
		COALESCE(resolucion_numero, ''),
		COALESCE(resolucion_fecha_desde, ''),
		COALESCE(resolucion_fecha_hasta, ''),
		COALESCE(consecutivo_desde, 1),
		COALESCE(consecutivo_hasta, 999999),
		COALESCE(proximo_consecutivo, 1),
		COALESCE(formato_impresion, 'carta'),
		COALESCE(mostrar_logo, 1),
		COALESCE(logo_url, ''),
		COALESCE(pie_factura, ''),
		COALESCE(notas_legales, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_configuracion_avanzada
	WHERE empresa_id = ?
	LIMIT 1`, empresaID)

	cfg := defaultConfigAvanzada(empresaID)
	var mostrarLogoInt int
	if err := row.Scan(
		&cfg.ID,
		&cfg.EmpresaID,
		&cfg.TipoDocumentoEmisor,
		&cfg.NIT,
		&cfg.DigitoVerificacion,
		&cfg.RazonSocial,
		&cfg.NombreComercial,
		&cfg.RegimenFiscal,
		&cfg.ResponsabilidadTributaria,
		&cfg.EmailFacturacion,
		&cfg.TelefonoFacturacion,
		&cfg.DireccionFiscal,
		&cfg.Departamento,
		&cfg.Municipio,
		&cfg.PaisCodigo,
		&cfg.CodigoPostal,
		&cfg.AmbienteFE,
		&cfg.TipoOperacion,
		&cfg.PrefijoFactura,
		&cfg.ResolucionNumero,
		&cfg.ResolucionFechaDesde,
		&cfg.ResolucionFechaHasta,
		&cfg.ConsecutivoDesde,
		&cfg.ConsecutivoHasta,
		&cfg.ProximoConsecutivo,
		&cfg.FormatoImpresion,
		&mostrarLogoInt,
		&cfg.LogoURL,
		&cfg.PieFactura,
		&cfg.NotasLegales,
		&cfg.FechaCreacion,
		&cfg.FechaActualizacion,
		&cfg.UsuarioCreador,
		&cfg.Estado,
		&cfg.Observaciones,
	); err != nil {
		if err == sql.ErrNoRows {
			return &cfg, nil
		}
		return nil, err
	}
	cfg.MostrarLogo = mostrarLogoInt == 1
	return &cfg, nil
}

// UpsertEmpresaConfiguracionAvanzada crea o actualiza la configuración avanzada por empresa.
func UpsertEmpresaConfiguracionAvanzada(dbConn *sql.DB, payload EmpresaConfiguracionAvanzada) (int64, error) {
	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id requerido")
	}

	payload.TipoDocumentoEmisor = strings.TrimSpace(payload.TipoDocumentoEmisor)
	if payload.TipoDocumentoEmisor == "" {
		payload.TipoDocumentoEmisor = "NIT"
	}
	payload.PaisCodigo = strings.TrimSpace(strings.ToUpper(payload.PaisCodigo))
	if payload.PaisCodigo == "" {
		payload.PaisCodigo = "CO"
	}
	payload.AmbienteFE = defaultAmbienteFE(payload.AmbienteFE)
	payload.TipoOperacion = defaultTipoOperacion(payload.TipoOperacion)
	payload.FormatoImpresion = defaultFormatoImpresion(payload.FormatoImpresion)
	if payload.ConsecutivoDesde <= 0 {
		payload.ConsecutivoDesde = 1
	}
	if payload.ConsecutivoHasta < payload.ConsecutivoDesde {
		payload.ConsecutivoHasta = payload.ConsecutivoDesde
	}
	if payload.ProximoConsecutivo < payload.ConsecutivoDesde || payload.ProximoConsecutivo > payload.ConsecutivoHasta {
		payload.ProximoConsecutivo = payload.ConsecutivoDesde
	}
	if strings.TrimSpace(payload.Estado) == "" {
		payload.Estado = "activo"
	}

	mostrarLogoInt := 0
	if payload.MostrarLogo {
		mostrarLogoInt = 1
	}

	_, err := dbConn.Exec(`INSERT INTO empresa_configuracion_avanzada (
		empresa_id,
		tipo_documento_emisor,
		nit,
		digito_verificacion,
		razon_social,
		nombre_comercial,
		regimen_fiscal,
		responsabilidad_tributaria,
		email_facturacion,
		telefono_facturacion,
		direccion_fiscal,
		departamento,
		municipio,
		pais_codigo,
		codigo_postal,
		ambiente_fe,
		tipo_operacion,
		prefijo_factura,
		resolucion_numero,
		resolucion_fecha_desde,
		resolucion_fecha_hasta,
		consecutivo_desde,
		consecutivo_hasta,
		proximo_consecutivo,
		formato_impresion,
		mostrar_logo,
		logo_url,
		pie_factura,
		notas_legales,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'), ?, ?, ?)
	ON CONFLICT(empresa_id) DO UPDATE SET
		tipo_documento_emisor = excluded.tipo_documento_emisor,
		nit = excluded.nit,
		digito_verificacion = excluded.digito_verificacion,
		razon_social = excluded.razon_social,
		nombre_comercial = excluded.nombre_comercial,
		regimen_fiscal = excluded.regimen_fiscal,
		responsabilidad_tributaria = excluded.responsabilidad_tributaria,
		email_facturacion = excluded.email_facturacion,
		telefono_facturacion = excluded.telefono_facturacion,
		direccion_fiscal = excluded.direccion_fiscal,
		departamento = excluded.departamento,
		municipio = excluded.municipio,
		pais_codigo = excluded.pais_codigo,
		codigo_postal = excluded.codigo_postal,
		ambiente_fe = excluded.ambiente_fe,
		tipo_operacion = excluded.tipo_operacion,
		prefijo_factura = excluded.prefijo_factura,
		resolucion_numero = excluded.resolucion_numero,
		resolucion_fecha_desde = excluded.resolucion_fecha_desde,
		resolucion_fecha_hasta = excluded.resolucion_fecha_hasta,
		consecutivo_desde = excluded.consecutivo_desde,
		consecutivo_hasta = excluded.consecutivo_hasta,
		proximo_consecutivo = excluded.proximo_consecutivo,
		formato_impresion = excluded.formato_impresion,
		mostrar_logo = excluded.mostrar_logo,
		logo_url = excluded.logo_url,
		pie_factura = excluded.pie_factura,
		notas_legales = excluded.notas_legales,
		fecha_actualizacion = datetime('now','localtime'),
		usuario_creador = CASE
			WHEN trim(excluded.usuario_creador) <> '' THEN excluded.usuario_creador
			ELSE empresa_configuracion_avanzada.usuario_creador
		END,
		estado = excluded.estado,
		observaciones = excluded.observaciones`,
		payload.EmpresaID,
		payload.TipoDocumentoEmisor,
		strings.TrimSpace(payload.NIT),
		strings.TrimSpace(payload.DigitoVerificacion),
		strings.TrimSpace(payload.RazonSocial),
		strings.TrimSpace(payload.NombreComercial),
		strings.TrimSpace(payload.RegimenFiscal),
		strings.TrimSpace(payload.ResponsabilidadTributaria),
		strings.TrimSpace(payload.EmailFacturacion),
		strings.TrimSpace(payload.TelefonoFacturacion),
		strings.TrimSpace(payload.DireccionFiscal),
		strings.TrimSpace(payload.Departamento),
		strings.TrimSpace(payload.Municipio),
		payload.PaisCodigo,
		strings.TrimSpace(payload.CodigoPostal),
		payload.AmbienteFE,
		payload.TipoOperacion,
		strings.TrimSpace(payload.PrefijoFactura),
		strings.TrimSpace(payload.ResolucionNumero),
		strings.TrimSpace(payload.ResolucionFechaDesde),
		strings.TrimSpace(payload.ResolucionFechaHasta),
		payload.ConsecutivoDesde,
		payload.ConsecutivoHasta,
		payload.ProximoConsecutivo,
		payload.FormatoImpresion,
		mostrarLogoInt,
		strings.TrimSpace(payload.LogoURL),
		strings.TrimSpace(payload.PieFactura),
		strings.TrimSpace(payload.NotasLegales),
		strings.TrimSpace(payload.UsuarioCreador),
		strings.TrimSpace(payload.Estado),
		strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		return 0, err
	}

	var id int64
	if err := dbConn.QueryRow(`SELECT id FROM empresa_configuracion_avanzada WHERE empresa_id = ? LIMIT 1`, payload.EmpresaID).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}
