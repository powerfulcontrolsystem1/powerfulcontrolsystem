package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// EmpresaConfiguracionAvanzada almacena la configuración empresarial y fiscal
// necesaria para preparar facturación electrónica en Colombia por empresa.
type EmpresaConfiguracionAvanzada struct {
	ID                                    int64  `json:"id"`
	EmpresaID                             int64  `json:"empresa_id"`
	ModoDocumentoVenta                    string `json:"modo_documento_venta,omitempty"`
	FacturacionElectronicaActiva          bool   `json:"facturacion_electronica_activa"`
	EnviarEmailVenta                      bool   `json:"enviar_email_venta"`
	EnviarFacturaElectronicaVenta         bool   `json:"enviar_factura_electronica_venta"`
	FacturacionFrecuenciaAutomaticaActiva bool   `json:"facturacion_frecuencia_automatica_activa"`
	FacturacionFrecuenciaCadaNNo          int64  `json:"facturacion_frecuencia_cada_n_no,omitempty"`
	FacturacionFrecuenciaContador         int64  `json:"facturacion_frecuencia_contador,omitempty"`
	TipoDocumentoEmisor                   string `json:"tipo_documento_emisor,omitempty"`
	NIT                                   string `json:"nit,omitempty"`
	DigitoVerificacion                    string `json:"digito_verificacion,omitempty"`
	RazonSocial                           string `json:"razon_social,omitempty"`
	NombreComercial                       string `json:"nombre_comercial,omitempty"`
	RegimenFiscal                         string `json:"regimen_fiscal,omitempty"`
	ResponsabilidadTributaria             string `json:"responsabilidad_tributaria,omitempty"`
	TipoPersonaFiscal                     string `json:"tipo_persona_fiscal,omitempty"`
	NaturalezaJuridica                    string `json:"naturaleza_juridica,omitempty"`
	RegimenTributarioColombia             string `json:"regimen_tributario_colombia,omitempty"`
	IVAResponsabilidad                    string `json:"iva_responsabilidad,omitempty"`
	INCResponsabilidad                    string `json:"inc_responsabilidad,omitempty"`
	ResponsabilidadesRUTJSON              string `json:"responsabilidades_rut_json,omitempty"`
	ObligacionesFiscalesJSON              string `json:"obligaciones_fiscales_json,omitempty"`
	EmailFacturacion                      string `json:"email_facturacion,omitempty"`
	TelefonoFacturacion                   string `json:"telefono_facturacion,omitempty"`
	DireccionFiscal                       string `json:"direccion_fiscal,omitempty"`
	Departamento                          string `json:"departamento,omitempty"`
	Municipio                             string `json:"municipio,omitempty"`
	PaisCodigo                            string `json:"pais_codigo,omitempty"`
	CodigoPostal                          string `json:"codigo_postal,omitempty"`
	AmbienteFE                            string `json:"ambiente_fe,omitempty"`
	TipoOperacion                         string `json:"tipo_operacion,omitempty"`
	PrefijoFactura                        string `json:"prefijo_factura,omitempty"`
	ResolucionNumero                      string `json:"resolucion_numero,omitempty"`
	ResolucionFechaDesde                  string `json:"resolucion_fecha_desde,omitempty"`
	ResolucionFechaHasta                  string `json:"resolucion_fecha_hasta,omitempty"`
	ConsecutivoDesde                      int64  `json:"consecutivo_desde,omitempty"`
	ConsecutivoHasta                      int64  `json:"consecutivo_hasta,omitempty"`
	ProximoConsecutivo                    int64  `json:"proximo_consecutivo,omitempty"`
	FormatoImpresion                      string `json:"formato_impresion,omitempty"`
	ImprimirVenta                         bool   `json:"imprimir_venta"`
	ImprimirFacturaElectronica            bool   `json:"imprimir_factura_electronica"`
	ImprimirCopiaFactura                  bool   `json:"imprimir_copia_factura"`
	MostrarDeducidoImpuestoFactura        bool   `json:"mostrar_deducido_impuesto_factura"`
	ImpresionReciboItemsJSON              string `json:"impresion_recibo_items_json,omitempty"`
	ImpresionCorteItemsJSON               string `json:"impresion_corte_items_json,omitempty"`
	ImpresionFacturaFuentePOS             int64  `json:"impresion_factura_fuente_pos,omitempty"`
	ImpresionFacturaFuenteCarta           int64  `json:"impresion_factura_fuente_carta,omitempty"`
	ImpresionReporteFuentePOS             int64  `json:"impresion_reporte_fuente_pos,omitempty"`
	ImpresionReporteFuenteCarta           int64  `json:"impresion_reporte_fuente_carta,omitempty"`
	MostrarLogo                           bool   `json:"mostrar_logo"`
	MostrarLogoEmpresa                    bool   `json:"mostrar_logo_empresa"`
	MostrarLogoFactura                    bool   `json:"mostrar_logo_factura"`
	MostrarLogoSistema                    bool   `json:"mostrar_logo_sistema"`
	LogoURL                               string `json:"logo_url,omitempty"`
	LogoFacturaURL                        string `json:"logo_factura_url,omitempty"`
	LogoSistemaURL                        string `json:"logo_sistema_url,omitempty"`
	PieFactura                            string `json:"pie_factura,omitempty"`
	NotasLegales                          string `json:"notas_legales,omitempty"`
	ColorCarritoActivo                    string `json:"color_carrito_activo,omitempty"`
	ColorCarritoInactivo                  string `json:"color_carrito_inactivo,omitempty"`
	ColorEstacionDisponible               string `json:"color_estacion_disponible,omitempty"`
	ColorEstacionOcupada                  string `json:"color_estacion_ocupada,omitempty"`
	ColorEstacionSucia                    string `json:"color_estacion_sucia,omitempty"`
	ColorEstacionAlertaTiempo             string `json:"color_estacion_alerta_tiempo,omitempty"`
	MonedaCodigo                          string `json:"moneda_codigo,omitempty"`
	SistemaNumerico                       string `json:"sistema_numerico,omitempty"`
	UsarDecimales                         bool   `json:"usar_decimales"`
	CantidadDecimales                     int64  `json:"cantidad_decimales,omitempty"`
	FechaCreacion                         string `json:"fecha_creacion,omitempty"`
	FechaActualizacion                    string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador                        string `json:"usuario_creador,omitempty"`
	Estado                                string `json:"estado,omitempty"`
	Observaciones                         string `json:"observaciones,omitempty"`
}

const (
	defaultColorCarritoActivo           = "#d9fbe8"
	defaultColorCarritoInactivo         = "#fff9ef"
	defaultColorEstacionDisponible      = "#fff9ef"
	defaultColorEstacionOcupada         = "#d9fbe8"
	defaultColorEstacionSucia           = "#ffe0e0"
	defaultColorEstacionAlertaTiempo    = "#fff3cd"
	defaultMonedaCodigo                 = "COP"
	defaultSistemaNumericoValue         = "latino"
	defaultCantidadDecimales            = int64(2)
	defaultModoDocumentoVenta           = "comprobante_pago"
	defaultFacturacionFrecuenciaCadaNNo = int64(0)
	defaultLogoSistemaURL               = "/img/logo.png"
)

// EnsureEmpresaConfiguracionAvanzadaSchema crea/migra el esquema de configuración avanzada
// por empresa para preparación de facturación electrónica en Colombia.
func EnsureEmpresaConfiguracionAvanzadaSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_configuracion_avanzada (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL UNIQUE,
			modo_documento_venta TEXT DEFAULT 'comprobante_pago',
			enviar_email_venta INTEGER DEFAULT 0,
			enviar_factura_electronica_venta INTEGER DEFAULT 0,
			facturacion_frecuencia_automatica_activa INTEGER DEFAULT 0,
			facturacion_frecuencia_cada_n_no INTEGER DEFAULT 0,
			facturacion_frecuencia_contador INTEGER DEFAULT 0,
			tipo_documento_emisor TEXT DEFAULT 'NIT',
			nit TEXT,
			digito_verificacion TEXT,
			razon_social TEXT,
			nombre_comercial TEXT,
			regimen_fiscal TEXT,
			responsabilidad_tributaria TEXT,
			tipo_persona_fiscal TEXT,
			naturaleza_juridica TEXT,
			regimen_tributario_colombia TEXT,
			iva_responsabilidad TEXT,
			inc_responsabilidad TEXT,
			responsabilidades_rut_json TEXT,
			obligaciones_fiscales_json TEXT,
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
			imprimir_venta INTEGER DEFAULT 0,
			imprimir_factura_electronica INTEGER DEFAULT 0,
			imprimir_copia_factura INTEGER DEFAULT 0,
			mostrar_deducido_impuesto_factura INTEGER DEFAULT 0,
			impresion_recibo_items_json TEXT,
			impresion_corte_items_json TEXT,
			impresion_factura_fuente_pos INTEGER DEFAULT 11,
			impresion_factura_fuente_carta INTEGER DEFAULT 13,
			impresion_reporte_fuente_pos INTEGER DEFAULT 11,
			impresion_reporte_fuente_carta INTEGER DEFAULT 13,
			mostrar_logo INTEGER DEFAULT 1,
			mostrar_logo_empresa INTEGER DEFAULT 1,
			mostrar_logo_factura INTEGER DEFAULT 1,
			mostrar_logo_sistema INTEGER DEFAULT 0,
			logo_url TEXT,
			logo_factura_url TEXT,
			pie_factura TEXT,
			notas_legales TEXT,
			color_carrito_activo TEXT DEFAULT '#d9fbe8',
			color_carrito_inactivo TEXT DEFAULT '#fff9ef',
			color_estacion_disponible TEXT DEFAULT '#fff9ef',
			color_estacion_ocupada TEXT DEFAULT '#d9fbe8',
			color_estacion_sucia TEXT DEFAULT '#ffe0e0',
			color_estacion_alerta_tiempo TEXT DEFAULT '#fff3cd',
			moneda_codigo TEXT DEFAULT 'COP',
			sistema_numerico TEXT DEFAULT 'latino',
			usar_decimales INTEGER DEFAULT 1,
			cantidad_decimales INTEGER DEFAULT 2,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_config_avanzada_empresa ON empresa_configuracion_avanzada(empresa_id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_config_avanzada_estado ON empresa_configuracion_avanzada(empresa_id, estado);`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "modo_documento_venta", "TEXT DEFAULT 'comprobante_pago'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "enviar_email_venta", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "enviar_factura_electronica_venta", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "facturacion_frecuencia_automatica_activa", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "facturacion_frecuencia_cada_n_no", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "facturacion_frecuencia_contador", "INTEGER DEFAULT 0"); err != nil {
		return err
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
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "tipo_persona_fiscal", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "naturaleza_juridica", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "regimen_tributario_colombia", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "iva_responsabilidad", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "inc_responsabilidad", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "responsabilidades_rut_json", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "obligaciones_fiscales_json", "TEXT"); err != nil {
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
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "imprimir_venta", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "imprimir_factura_electronica", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "imprimir_copia_factura", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "mostrar_deducido_impuesto_factura", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "impresion_recibo_items_json", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "impresion_corte_items_json", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "impresion_factura_fuente_pos", "INTEGER DEFAULT 11"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "impresion_factura_fuente_carta", "INTEGER DEFAULT 13"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "impresion_reporte_fuente_pos", "INTEGER DEFAULT 11"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "impresion_reporte_fuente_carta", "INTEGER DEFAULT 13"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "mostrar_logo", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "mostrar_logo_empresa", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "mostrar_logo_factura", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "mostrar_logo_sistema", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "logo_url", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "logo_factura_url", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "pie_factura", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "notas_legales", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "color_carrito_activo", "TEXT DEFAULT '#d9fbe8'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "color_carrito_inactivo", "TEXT DEFAULT '#fff9ef'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "color_estacion_disponible", "TEXT DEFAULT '#fff9ef'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "color_estacion_ocupada", "TEXT DEFAULT '#d9fbe8'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "color_estacion_sucia", "TEXT DEFAULT '#ffe0e0'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "color_estacion_alerta_tiempo", "TEXT DEFAULT '#fff3cd'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "moneda_codigo", "TEXT DEFAULT 'COP'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "sistema_numerico", "TEXT DEFAULT 'latino'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "usar_decimales", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "cantidad_decimales", "INTEGER DEFAULT 2"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_configuracion_avanzada", "fecha_creacion", "TEXT DEFAULT (CURRENT_TIMESTAMP)"); err != nil {
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
	if err := ensureEmpresaConfiguracionAvanzadaFlagColumns(dbConn); err != nil {
		return err
	}
	return nil
}

func ensureEmpresaConfiguracionAvanzadaFlagColumns(dbConn *sql.DB) error {
	if dbConn == nil || !isPostgresDialect() {
		return nil
	}

	flagDefaults := map[string]int{
		"enviar_email_venta":                       0,
		"enviar_factura_electronica_venta":         0,
		"facturacion_frecuencia_automatica_activa": 0,
		"imprimir_venta":                           0,
		"imprimir_factura_electronica":             0,
		"imprimir_copia_factura":                   0,
		"mostrar_deducido_impuesto_factura":        0,
		"mostrar_logo":                             1,
		"mostrar_logo_empresa":                     1,
		"mostrar_logo_factura":                     1,
		"mostrar_logo_sistema":                     0,
		"usar_decimales":                           1,
	}
	for column, defaultValue := range flagDefaults {
		var dataType string
		err := QueryRowCompat(dbConn, `SELECT data_type
			FROM information_schema.columns
			WHERE table_schema = current_schema()
			  AND table_name = 'empresa_configuracion_avanzada'
			  AND column_name = ?
			LIMIT 1`, column).Scan(&dataType)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return err
		}
		if strings.ToLower(strings.TrimSpace(dataType)) != "boolean" {
			continue
		}

		quotedColumn := quotePostgresIdentifier(column)
		stmt := fmt.Sprintf(`ALTER TABLE empresa_configuracion_avanzada
			ALTER COLUMN %s DROP DEFAULT,
			ALTER COLUMN %s TYPE INTEGER USING CASE
				WHEN %s IS NULL THEN NULL
				WHEN lower(trim(%s::text)) IN ('1', 't', 'true', 'yes', 'y', 'on') THEN 1
				ELSE 0
			END,
			ALTER COLUMN %s SET DEFAULT %d`,
			quotedColumn,
			quotedColumn,
			quotedColumn,
			quotedColumn,
			quotedColumn,
			defaultValue,
		)
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func defaultConfigAvanzada(empresaID int64) EmpresaConfiguracionAvanzada {
	return EmpresaConfiguracionAvanzada{
		EmpresaID:                             empresaID,
		ModoDocumentoVenta:                    defaultModoDocumentoVenta,
		FacturacionElectronicaActiva:          false,
		EnviarEmailVenta:                      false,
		EnviarFacturaElectronicaVenta:         false,
		FacturacionFrecuenciaAutomaticaActiva: false,
		FacturacionFrecuenciaCadaNNo:          defaultFacturacionFrecuenciaCadaNNo,
		FacturacionFrecuenciaContador:         0,
		TipoDocumentoEmisor:                   "NIT",
		PaisCodigo:                            "CO",
		AmbienteFE:                            "habilitacion",
		TipoOperacion:                         "10",
		ConsecutivoDesde:                      1,
		ConsecutivoHasta:                      999999,
		ProximoConsecutivo:                    1,
		FormatoImpresion:                      "carta",
		ImprimirVenta:                         false,
		ImprimirFacturaElectronica:            false,
		ImpresionFacturaFuentePOS:             11,
		ImpresionFacturaFuenteCarta:           13,
		ImpresionReporteFuentePOS:             11,
		ImpresionReporteFuenteCarta:           13,
		MostrarLogo:                           true,
		MostrarLogoEmpresa:                    true,
		MostrarLogoFactura:                    true,
		MostrarLogoSistema:                    false,
		LogoSistemaURL:                        defaultLogoSistemaURL,
		ColorCarritoActivo:                    defaultColorCarritoActivo,
		ColorCarritoInactivo:                  defaultColorCarritoInactivo,
		ColorEstacionDisponible:               defaultColorEstacionDisponible,
		ColorEstacionOcupada:                  defaultColorEstacionOcupada,
		ColorEstacionSucia:                    defaultColorEstacionSucia,
		ColorEstacionAlertaTiempo:             defaultColorEstacionAlertaTiempo,
		MonedaCodigo:                          defaultMonedaCodigo,
		SistemaNumerico:                       defaultSistemaNumericoValue,
		UsarDecimales:                         true,
		CantidadDecimales:                     defaultCantidadDecimales,
		Estado:                                "activo",
	}
}

func normalizeHexColor(v string, fallback string) string {
	normalize := func(raw string) string {
		raw = strings.TrimSpace(strings.ToLower(raw))
		if len(raw) != 7 || raw[0] != '#' {
			return ""
		}
		for i := 1; i < len(raw); i++ {
			c := raw[i]
			isDigit := c >= '0' && c <= '9'
			isHexChar := c >= 'a' && c <= 'f'
			if !isDigit && !isHexChar {
				return ""
			}
		}
		return raw
	}

	if out := normalize(v); out != "" {
		return out
	}
	if out := normalize(fallback); out != "" {
		return out
	}
	return defaultColorCarritoActivo
}

func defaultFormatoImpresion(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "pos" {
		return "pos"
	}
	return "carta"
}

func normalizePrintFontSize(v int64, fallback int64, min int64, max int64) int64 {
	if fallback < min {
		fallback = min
	}
	if fallback > max {
		fallback = max
	}
	if v < min || v > max {
		return fallback
	}
	return v
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

func defaultModoDocumentoVentaValue(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "factura_electronica" {
		return "factura_electronica"
	}
	return defaultModoDocumentoVenta
}

func normalizeMonedaCodigo(v string) string {
	v = strings.TrimSpace(strings.ToUpper(v))
	if v == "" {
		return defaultMonedaCodigo
	}
	if len(v) > 10 {
		return defaultMonedaCodigo
	}
	return v
}

func defaultSistemaNumerico(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "internacional" {
		return "internacional"
	}
	return defaultSistemaNumericoValue
}

func normalizeCantidadDecimales(v int64) int64 {
	if v < 0 || v > 6 {
		return defaultCantidadDecimales
	}
	return v
}

func normalizeFrecuenciaCadaNNo(v int64) int64 {
	if v < 0 {
		return 0
	}
	if v > 1000 {
		return 1000
	}
	return v
}

func normalizeFrecuenciaContador(v int64, cadaNNo int64) int64 {
	if v < 0 {
		v = 0
	}
	ciclo := cadaNNo + 1
	if ciclo <= 0 {
		ciclo = 1
	}
	if v >= ciclo {
		return v % ciclo
	}
	return v
}

// GetEmpresaConfiguracionAvanzada obtiene la configuración avanzada por empresa.
// Si no existe registro, retorna valores por defecto para facilitar captura inicial.
func normalizePrintItemsJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 8000 {
		return ""
	}
	var incoming map[string]bool
	if err := json.Unmarshal([]byte(raw), &incoming); err != nil {
		return ""
	}
	clean := make(map[string]bool, len(incoming))
	for key, value := range incoming {
		key = strings.TrimSpace(strings.ToLower(key))
		if key == "" || len(key) > 80 {
			continue
		}
		valid := true
		for i := 0; i < len(key); i++ {
			c := key[i]
			if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
				valid = false
				break
			}
		}
		if valid {
			clean[key] = value
		}
	}
	if len(clean) == 0 {
		return ""
	}
	out, err := json.Marshal(clean)
	if err != nil {
		return ""
	}
	return string(out)
}

func normalizeFiscalToken(v string, allowed map[string]bool) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if allowed[v] {
		return v
	}
	return ""
}

func normalizeResponsabilidadesRUTJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 4000 {
		return ""
	}
	var incoming []string
	if err := json.Unmarshal([]byte(raw), &incoming); err != nil {
		return ""
	}
	allowed := map[string]bool{
		"04": true, "05": true, "06": true, "07": true, "09": true, "13": true, "14": true, "15": true,
		"16": true, "19": true, "20": true, "22": true, "23": true, "24": true, "26": true, "33": true,
		"42": true, "46": true, "47": true, "48": true, "49": true, "50": true, "52": true, "53": true, "55": true,
		"59": true, "60": true,
	}
	seen := map[string]bool{}
	clean := make([]string, 0, len(incoming))
	for _, item := range incoming {
		code := strings.TrimSpace(item)
		if len(code) == 1 {
			code = "0" + code
		}
		if !allowed[code] || seen[code] {
			continue
		}
		seen[code] = true
		clean = append(clean, code)
	}
	if len(clean) == 0 {
		return ""
	}
	out, err := json.Marshal(clean)
	if err != nil {
		return ""
	}
	return string(out)
}

func normalizeObligacionesFiscalesJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 4000 {
		return ""
	}
	var incoming map[string]bool
	if err := json.Unmarshal([]byte(raw), &incoming); err != nil {
		return ""
	}
	allowed := map[string]bool{
		"renta": true, "rst_declaracion_anual": true, "rst_anticipos_bimestrales": true,
		"iva": true, "inc": true, "retencion_fuente": true, "reteiva": true,
		"facturacion_electronica": true, "informacion_exogena": true,
		"beneficiarios_finales": true, "ica": true,
	}
	clean := make(map[string]bool)
	for key, value := range incoming {
		key = strings.TrimSpace(strings.ToLower(key))
		if allowed[key] {
			clean[key] = value
		}
	}
	if len(clean) == 0 {
		return ""
	}
	out, err := json.Marshal(clean)
	if err != nil {
		return ""
	}
	return string(out)
}

func GetEmpresaConfiguracionAvanzada(dbConn *sql.DB, empresaID int64) (*EmpresaConfiguracionAvanzada, error) {
	if err := EnsureEmpresaConfiguracionAvanzadaSchema(dbConn); err != nil {
		return nil, err
	}

	row := QueryRowCompat(dbConn, `SELECT
		id,
		empresa_id,
		COALESCE(modo_documento_venta, 'comprobante_pago'),
		COALESCE(enviar_email_venta, 0),
		COALESCE(enviar_factura_electronica_venta, 0),
		COALESCE(facturacion_frecuencia_automatica_activa, 0),
		COALESCE(facturacion_frecuencia_cada_n_no, 0),
		COALESCE(facturacion_frecuencia_contador, 0),
		COALESCE(tipo_documento_emisor, 'NIT'),
		COALESCE(nit, ''),
		COALESCE(digito_verificacion, ''),
		COALESCE(razon_social, ''),
		COALESCE(nombre_comercial, ''),
		COALESCE(regimen_fiscal, ''),
		COALESCE(responsabilidad_tributaria, ''),
		COALESCE(tipo_persona_fiscal, ''),
		COALESCE(naturaleza_juridica, ''),
		COALESCE(regimen_tributario_colombia, ''),
		COALESCE(iva_responsabilidad, ''),
		COALESCE(inc_responsabilidad, ''),
		COALESCE(responsabilidades_rut_json, ''),
		COALESCE(obligaciones_fiscales_json, ''),
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
		COALESCE(imprimir_venta, 0),
		COALESCE(imprimir_factura_electronica, 0),
		COALESCE(imprimir_copia_factura, 0),
		COALESCE(mostrar_deducido_impuesto_factura, 0),
		COALESCE(impresion_recibo_items_json, ''),
		COALESCE(impresion_corte_items_json, ''),
		COALESCE(impresion_factura_fuente_pos, 11),
		COALESCE(impresion_factura_fuente_carta, 13),
		COALESCE(impresion_reporte_fuente_pos, 11),
		COALESCE(impresion_reporte_fuente_carta, 13),
		COALESCE(mostrar_logo, 1),
		COALESCE(mostrar_logo_empresa, mostrar_logo, 1),
		COALESCE(mostrar_logo_factura, mostrar_logo, 1),
		COALESCE(mostrar_logo_sistema, 0),
		COALESCE(logo_url, ''),
		COALESCE(logo_factura_url, logo_url, ''),
		COALESCE(pie_factura, ''),
		COALESCE(notas_legales, ''),
		COALESCE(color_carrito_activo, '#d9fbe8'),
		COALESCE(color_carrito_inactivo, '#fff9ef'),
		COALESCE(color_estacion_disponible, '#fff9ef'),
		COALESCE(color_estacion_ocupada, '#d9fbe8'),
		COALESCE(color_estacion_sucia, '#ffe0e0'),
		COALESCE(color_estacion_alerta_tiempo, '#fff3cd'),
		COALESCE(moneda_codigo, 'COP'),
		COALESCE(sistema_numerico, 'latino'),
		COALESCE(usar_decimales, 1),
		COALESCE(cantidad_decimales, 2),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_configuracion_avanzada
	WHERE empresa_id = ?
	LIMIT 1`, empresaID)

	cfg := defaultConfigAvanzada(empresaID)
	var imprimirVentaInt int
	var imprimirFacturaElectronicaInt int
	var imprimirCopiaFacturaInt int
	var mostrarDeducidoImpuestoFacturaInt int
	var mostrarLogoInt int
	var mostrarLogoEmpresaInt int
	var mostrarLogoFacturaInt int
	var mostrarLogoSistemaInt int
	var usarDecimalesInt int
	var enviarEmailVentaInt int
	var enviarFacturaElectronicaVentaInt int
	var frecuenciaAutomaticaActivaInt int
	if err := row.Scan(
		&cfg.ID,
		&cfg.EmpresaID,
		&cfg.ModoDocumentoVenta,
		&enviarEmailVentaInt,
		&enviarFacturaElectronicaVentaInt,
		&frecuenciaAutomaticaActivaInt,
		&cfg.FacturacionFrecuenciaCadaNNo,
		&cfg.FacturacionFrecuenciaContador,
		&cfg.TipoDocumentoEmisor,
		&cfg.NIT,
		&cfg.DigitoVerificacion,
		&cfg.RazonSocial,
		&cfg.NombreComercial,
		&cfg.RegimenFiscal,
		&cfg.ResponsabilidadTributaria,
		&cfg.TipoPersonaFiscal,
		&cfg.NaturalezaJuridica,
		&cfg.RegimenTributarioColombia,
		&cfg.IVAResponsabilidad,
		&cfg.INCResponsabilidad,
		&cfg.ResponsabilidadesRUTJSON,
		&cfg.ObligacionesFiscalesJSON,
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
		&imprimirVentaInt,
		&imprimirFacturaElectronicaInt,
		&imprimirCopiaFacturaInt,
		&mostrarDeducidoImpuestoFacturaInt,
		&cfg.ImpresionReciboItemsJSON,
		&cfg.ImpresionCorteItemsJSON,
		&cfg.ImpresionFacturaFuentePOS,
		&cfg.ImpresionFacturaFuenteCarta,
		&cfg.ImpresionReporteFuentePOS,
		&cfg.ImpresionReporteFuenteCarta,
		&mostrarLogoInt,
		&mostrarLogoEmpresaInt,
		&mostrarLogoFacturaInt,
		&mostrarLogoSistemaInt,
		&cfg.LogoURL,
		&cfg.LogoFacturaURL,
		&cfg.PieFactura,
		&cfg.NotasLegales,
		&cfg.ColorCarritoActivo,
		&cfg.ColorCarritoInactivo,
		&cfg.ColorEstacionDisponible,
		&cfg.ColorEstacionOcupada,
		&cfg.ColorEstacionSucia,
		&cfg.ColorEstacionAlertaTiempo,
		&cfg.MonedaCodigo,
		&cfg.SistemaNumerico,
		&usarDecimalesInt,
		&cfg.CantidadDecimales,
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
	cfg.ModoDocumentoVenta = defaultModoDocumentoVentaValue(cfg.ModoDocumentoVenta)
	cfg.FacturacionElectronicaActiva = cfg.ModoDocumentoVenta == "factura_electronica"
	cfg.EnviarEmailVenta = enviarEmailVentaInt == 1
	cfg.EnviarFacturaElectronicaVenta = enviarFacturaElectronicaVentaInt == 1
	cfg.FacturacionFrecuenciaAutomaticaActiva = frecuenciaAutomaticaActivaInt == 1
	cfg.FacturacionFrecuenciaCadaNNo = normalizeFrecuenciaCadaNNo(cfg.FacturacionFrecuenciaCadaNNo)
	cfg.FacturacionFrecuenciaContador = normalizeFrecuenciaContador(cfg.FacturacionFrecuenciaContador, cfg.FacturacionFrecuenciaCadaNNo)
	cfg.ImprimirVenta = imprimirVentaInt == 1
	cfg.ImprimirFacturaElectronica = imprimirFacturaElectronicaInt == 1
	cfg.ImprimirCopiaFactura = imprimirCopiaFacturaInt == 1
	cfg.MostrarDeducidoImpuestoFactura = mostrarDeducidoImpuestoFacturaInt == 1
	cfg.ImpresionFacturaFuentePOS = normalizePrintFontSize(cfg.ImpresionFacturaFuentePOS, 11, 8, 16)
	cfg.ImpresionFacturaFuenteCarta = normalizePrintFontSize(cfg.ImpresionFacturaFuenteCarta, 13, 10, 22)
	cfg.ImpresionReporteFuentePOS = normalizePrintFontSize(cfg.ImpresionReporteFuentePOS, 11, 8, 16)
	cfg.ImpresionReporteFuenteCarta = normalizePrintFontSize(cfg.ImpresionReporteFuenteCarta, 13, 10, 22)
	cfg.MostrarLogoEmpresa = mostrarLogoEmpresaInt == 1
	cfg.MostrarLogoFactura = mostrarLogoFacturaInt == 1
	cfg.MostrarLogoSistema = mostrarLogoSistemaInt == 1
	cfg.MostrarLogo = mostrarLogoInt == 1 && (cfg.MostrarLogoEmpresa || cfg.MostrarLogoFactura || cfg.MostrarLogoSistema)
	cfg.LogoSistemaURL = defaultLogoSistemaURL
	cfg.ColorCarritoActivo = normalizeHexColor(cfg.ColorCarritoActivo, defaultColorCarritoActivo)
	cfg.ColorCarritoInactivo = normalizeHexColor(cfg.ColorCarritoInactivo, defaultColorCarritoInactivo)
	cfg.ColorEstacionDisponible = normalizeHexColor(cfg.ColorEstacionDisponible, defaultColorEstacionDisponible)
	cfg.ColorEstacionOcupada = normalizeHexColor(cfg.ColorEstacionOcupada, defaultColorEstacionOcupada)
	cfg.ColorEstacionSucia = normalizeHexColor(cfg.ColorEstacionSucia, defaultColorEstacionSucia)
	cfg.ColorEstacionAlertaTiempo = normalizeHexColor(cfg.ColorEstacionAlertaTiempo, defaultColorEstacionAlertaTiempo)
	cfg.MonedaCodigo = normalizeMonedaCodigo(cfg.MonedaCodigo)
	cfg.SistemaNumerico = defaultSistemaNumerico(cfg.SistemaNumerico)
	cfg.UsarDecimales = usarDecimalesInt == 1
	cfg.CantidadDecimales = normalizeCantidadDecimales(cfg.CantidadDecimales)
	if cfg.UsarDecimales {
		if cfg.CantidadDecimales <= 0 {
			cfg.CantidadDecimales = defaultCantidadDecimales
		}
	} else {
		cfg.CantidadDecimales = 0
	}
	return &cfg, nil
}

// UpsertEmpresaConfiguracionAvanzada crea o actualiza la configuración avanzada por empresa.
func UpsertEmpresaConfiguracionAvanzada(dbConn *sql.DB, payload EmpresaConfiguracionAvanzada) (int64, error) {
	if err := EnsureEmpresaConfiguracionAvanzadaSchema(dbConn); err != nil {
		return 0, err
	}

	if payload.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id requerido")
	}

	payload.TipoDocumentoEmisor = strings.TrimSpace(payload.TipoDocumentoEmisor)
	payload.ModoDocumentoVenta = defaultModoDocumentoVentaValue(payload.ModoDocumentoVenta)
	payload.FacturacionElectronicaActiva = payload.ModoDocumentoVenta == "factura_electronica"
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
	payload.ImpresionFacturaFuentePOS = normalizePrintFontSize(payload.ImpresionFacturaFuentePOS, 11, 8, 16)
	payload.ImpresionFacturaFuenteCarta = normalizePrintFontSize(payload.ImpresionFacturaFuenteCarta, 13, 10, 22)
	payload.ImpresionReporteFuentePOS = normalizePrintFontSize(payload.ImpresionReporteFuentePOS, 11, 8, 16)
	payload.ImpresionReporteFuenteCarta = normalizePrintFontSize(payload.ImpresionReporteFuenteCarta, 13, 10, 22)
	payload.ColorCarritoActivo = normalizeHexColor(payload.ColorCarritoActivo, defaultColorCarritoActivo)
	payload.ColorCarritoInactivo = normalizeHexColor(payload.ColorCarritoInactivo, defaultColorCarritoInactivo)
	payload.ColorEstacionDisponible = normalizeHexColor(payload.ColorEstacionDisponible, defaultColorEstacionDisponible)
	payload.ColorEstacionOcupada = normalizeHexColor(payload.ColorEstacionOcupada, defaultColorEstacionOcupada)
	payload.ColorEstacionSucia = normalizeHexColor(payload.ColorEstacionSucia, defaultColorEstacionSucia)
	payload.ColorEstacionAlertaTiempo = normalizeHexColor(payload.ColorEstacionAlertaTiempo, defaultColorEstacionAlertaTiempo)
	payload.MonedaCodigo = normalizeMonedaCodigo(payload.MonedaCodigo)
	payload.SistemaNumerico = defaultSistemaNumerico(payload.SistemaNumerico)
	payload.CantidadDecimales = normalizeCantidadDecimales(payload.CantidadDecimales)
	payload.FacturacionFrecuenciaCadaNNo = normalizeFrecuenciaCadaNNo(payload.FacturacionFrecuenciaCadaNNo)
	payload.FacturacionFrecuenciaContador = normalizeFrecuenciaContador(payload.FacturacionFrecuenciaContador, payload.FacturacionFrecuenciaCadaNNo)
	payload.TipoPersonaFiscal = normalizeFiscalToken(payload.TipoPersonaFiscal, map[string]bool{
		"persona_natural":  true,
		"persona_juridica": true,
	})
	payload.NaturalezaJuridica = normalizeFiscalToken(payload.NaturalezaJuridica, map[string]bool{
		"natural_asalariado":    true,
		"natural_independiente": true,
		"natural_comerciante":   true,
		"natural_profesional":   true,
		"sas":                   true,
		"ltda":                  true,
		"sa":                    true,
		"esal":                  true,
		"entidad_publica":       true,
		"sucursal_extranjera":   true,
		"otra_juridica":         true,
	})
	payload.RegimenTributarioColombia = normalizeFiscalToken(payload.RegimenTributarioColombia, map[string]bool{
		"ordinario":           true,
		"simple":              true,
		"especial":            true,
		"ingresos_patrimonio": true,
		"no_declarante_pn":    true,
	})
	payload.IVAResponsabilidad = normalizeFiscalToken(payload.IVAResponsabilidad, map[string]bool{
		"responsable_iva":              true,
		"no_responsable_iva":           true,
		"pj_no_responsable_iva_simple": true,
		"bienes_exentos":               true,
		"servicios_exterior":           true,
	})
	payload.INCResponsabilidad = normalizeFiscalToken(payload.INCResponsabilidad, map[string]bool{
		"no_aplica":                             true,
		"responsable_inc":                       true,
		"no_responsable_inc_restaurantes_bares": true,
	})
	payload.ResponsabilidadesRUTJSON = normalizeResponsabilidadesRUTJSON(payload.ResponsabilidadesRUTJSON)
	payload.ObligacionesFiscalesJSON = normalizeObligacionesFiscalesJSON(payload.ObligacionesFiscalesJSON)
	if strings.TrimSpace(payload.RegimenFiscal) == "" && payload.RegimenTributarioColombia != "" {
		payload.RegimenFiscal = payload.RegimenTributarioColombia
	}
	if strings.TrimSpace(payload.ResponsabilidadTributaria) == "" && payload.IVAResponsabilidad != "" {
		payload.ResponsabilidadTributaria = payload.IVAResponsabilidad
	}
	if payload.UsarDecimales {
		if payload.CantidadDecimales <= 0 {
			payload.CantidadDecimales = defaultCantidadDecimales
		}
	} else {
		payload.CantidadDecimales = 0
	}
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
	payload.LogoURL = strings.TrimSpace(payload.LogoURL)
	payload.LogoFacturaURL = strings.TrimSpace(payload.LogoFacturaURL)
	payload.LogoSistemaURL = defaultLogoSistemaURL
	payload.ImpresionReciboItemsJSON = normalizePrintItemsJSON(payload.ImpresionReciboItemsJSON)
	payload.ImpresionCorteItemsJSON = normalizePrintItemsJSON(payload.ImpresionCorteItemsJSON)
	if payload.MostrarLogo && !payload.MostrarLogoEmpresa && !payload.MostrarLogoFactura && !payload.MostrarLogoSistema {
		payload.MostrarLogoEmpresa = true
		payload.MostrarLogoFactura = true
	}
	payload.MostrarLogo = payload.MostrarLogoEmpresa || payload.MostrarLogoFactura || payload.MostrarLogoSistema

	mostrarLogoInt := 0
	if payload.MostrarLogo {
		mostrarLogoInt = 1
	}
	mostrarLogoEmpresaInt := 0
	if payload.MostrarLogoEmpresa {
		mostrarLogoEmpresaInt = 1
	}
	mostrarLogoFacturaInt := 0
	if payload.MostrarLogoFactura {
		mostrarLogoFacturaInt = 1
	}
	mostrarLogoSistemaInt := 0
	if payload.MostrarLogoSistema {
		mostrarLogoSistemaInt = 1
	}

	imprimirVentaInt := 0
	if payload.ImprimirVenta {
		imprimirVentaInt = 1
	}
	imprimirFacturaElectronicaInt := 0
	if payload.ImprimirFacturaElectronica {
		imprimirFacturaElectronicaInt = 1
	}
	imprimirCopiaFacturaInt := 0
	if payload.ImprimirCopiaFactura {
		imprimirCopiaFacturaInt = 1
	}
	mostrarDeducidoImpuestoFacturaInt := 0
	if payload.MostrarDeducidoImpuestoFactura {
		mostrarDeducidoImpuestoFacturaInt = 1
	}

	usarDecimalesInt := 0
	if payload.UsarDecimales {
		usarDecimalesInt = 1
	}
	enviarEmailVentaInt := 0
	if payload.EnviarEmailVenta {
		enviarEmailVentaInt = 1
	}
	enviarFacturaElectronicaVentaInt := 0
	if payload.EnviarFacturaElectronicaVenta {
		enviarFacturaElectronicaVentaInt = 1
	}
	frecuenciaAutomaticaActivaInt := 0
	if payload.FacturacionFrecuenciaAutomaticaActiva {
		frecuenciaAutomaticaActivaInt = 1
	}

	nowExpr := sqlNowExpr()
	_, err := ExecCompat(dbConn, `INSERT INTO empresa_configuracion_avanzada (
		empresa_id,
		modo_documento_venta,
		enviar_email_venta,
		enviar_factura_electronica_venta,
		facturacion_frecuencia_automatica_activa,
		facturacion_frecuencia_cada_n_no,
		facturacion_frecuencia_contador,
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
		imprimir_venta,
		imprimir_factura_electronica,
		imprimir_copia_factura,
		mostrar_logo,
		logo_url,
		pie_factura,
		notas_legales,
		color_carrito_activo,
		color_carrito_inactivo,
		color_estacion_disponible,
		color_estacion_ocupada,
		color_estacion_sucia,
		color_estacion_alerta_tiempo,
		moneda_codigo,
		sistema_numerico,
		usar_decimales,
		cantidad_decimales,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, `+nowExpr+`, `+nowExpr+`, ?, ?, ?)
	ON CONFLICT(empresa_id) DO UPDATE SET
		modo_documento_venta = excluded.modo_documento_venta,
		enviar_email_venta = excluded.enviar_email_venta,
		enviar_factura_electronica_venta = excluded.enviar_factura_electronica_venta,
		facturacion_frecuencia_automatica_activa = excluded.facturacion_frecuencia_automatica_activa,
		facturacion_frecuencia_cada_n_no = excluded.facturacion_frecuencia_cada_n_no,
		facturacion_frecuencia_contador = excluded.facturacion_frecuencia_contador,
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
		imprimir_venta = excluded.imprimir_venta,
		imprimir_factura_electronica = excluded.imprimir_factura_electronica,
		imprimir_copia_factura = excluded.imprimir_copia_factura,
		mostrar_logo = excluded.mostrar_logo,
		logo_url = excluded.logo_url,
		pie_factura = excluded.pie_factura,
		notas_legales = excluded.notas_legales,
		color_carrito_activo = excluded.color_carrito_activo,
		color_carrito_inactivo = excluded.color_carrito_inactivo,
		color_estacion_disponible = excluded.color_estacion_disponible,
		color_estacion_ocupada = excluded.color_estacion_ocupada,
		color_estacion_sucia = excluded.color_estacion_sucia,
		color_estacion_alerta_tiempo = excluded.color_estacion_alerta_tiempo,
		moneda_codigo = excluded.moneda_codigo,
		sistema_numerico = excluded.sistema_numerico,
		usar_decimales = excluded.usar_decimales,
		cantidad_decimales = excluded.cantidad_decimales,
		fecha_actualizacion = `+nowExpr+`,
		usuario_creador = CASE
			WHEN trim(excluded.usuario_creador) <> '' THEN excluded.usuario_creador
			ELSE empresa_configuracion_avanzada.usuario_creador
		END,
		estado = excluded.estado,
		observaciones = excluded.observaciones`,
		payload.EmpresaID,
		payload.ModoDocumentoVenta,
		enviarEmailVentaInt,
		enviarFacturaElectronicaVentaInt,
		frecuenciaAutomaticaActivaInt,
		payload.FacturacionFrecuenciaCadaNNo,
		payload.FacturacionFrecuenciaContador,
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
		imprimirVentaInt,
		imprimirFacturaElectronicaInt,
		imprimirCopiaFacturaInt,
		mostrarLogoInt,
		strings.TrimSpace(payload.LogoURL),
		strings.TrimSpace(payload.PieFactura),
		strings.TrimSpace(payload.NotasLegales),
		payload.ColorCarritoActivo,
		payload.ColorCarritoInactivo,
		payload.ColorEstacionDisponible,
		payload.ColorEstacionOcupada,
		payload.ColorEstacionSucia,
		payload.ColorEstacionAlertaTiempo,
		payload.MonedaCodigo,
		payload.SistemaNumerico,
		usarDecimalesInt,
		payload.CantidadDecimales,
		strings.TrimSpace(payload.UsuarioCreador),
		strings.TrimSpace(payload.Estado),
		strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	if _, err := ExecCompat(dbConn, `UPDATE empresa_configuracion_avanzada
		SET mostrar_logo_empresa = ?,
			mostrar_logo_factura = ?,
			mostrar_logo_sistema = ?,
			mostrar_logo = ?,
			logo_factura_url = ?,
			mostrar_deducido_impuesto_factura = ?,
			impresion_recibo_items_json = ?,
			impresion_corte_items_json = ?,
			impresion_factura_fuente_pos = ?,
			impresion_factura_fuente_carta = ?,
			impresion_reporte_fuente_pos = ?,
			impresion_reporte_fuente_carta = ?,
			tipo_persona_fiscal = ?,
			naturaleza_juridica = ?,
			regimen_tributario_colombia = ?,
			iva_responsabilidad = ?,
			inc_responsabilidad = ?,
			responsabilidades_rut_json = ?,
			obligaciones_fiscales_json = ?,
			fecha_actualizacion = `+nowExpr+`
		WHERE empresa_id = ?`,
		mostrarLogoEmpresaInt,
		mostrarLogoFacturaInt,
		mostrarLogoSistemaInt,
		mostrarLogoInt,
		strings.TrimSpace(payload.LogoFacturaURL),
		mostrarDeducidoImpuestoFacturaInt,
		payload.ImpresionReciboItemsJSON,
		payload.ImpresionCorteItemsJSON,
		payload.ImpresionFacturaFuentePOS,
		payload.ImpresionFacturaFuenteCarta,
		payload.ImpresionReporteFuentePOS,
		payload.ImpresionReporteFuenteCarta,
		payload.TipoPersonaFiscal,
		payload.NaturalezaJuridica,
		payload.RegimenTributarioColombia,
		payload.IVAResponsabilidad,
		payload.INCResponsabilidad,
		payload.ResponsabilidadesRUTJSON,
		payload.ObligacionesFiscalesJSON,
		payload.EmpresaID,
	); err != nil {
		return 0, err
	}

	var id int64
	if err := QueryRowCompat(dbConn, `SELECT id FROM empresa_configuracion_avanzada WHERE empresa_id = ? LIMIT 1`, payload.EmpresaID).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}
