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

// FacturacionPanamaChecklist resume los requisitos operativos propios del SFEP/DGI.
type FacturacionPanamaChecklist struct {
	PaisCodigo           string                             `json:"pais_codigo"`
	Ok                   bool                               `json:"ok"`
	Estado               string                             `json:"estado"`
	Modalidad            string                             `json:"modalidad"`
	Ambiente             string                             `json:"ambiente"`
	Faltantes            []string                           `json:"faltantes"`
	Advertencias         []string                           `json:"advertencias"`
	DocumentosSoportados []string                           `json:"documentos_soportados"`
	Items                []FacturacionPanamaChecklistItem   `json:"items"`
	Fuentes              []FacturacionPanamaFuenteNormativa `json:"fuentes"`
}

type FacturacionPanamaChecklistItem struct {
	Clave   string `json:"clave"`
	Titulo  string `json:"titulo"`
	Estado  string `json:"estado"`
	Detalle string `json:"detalle"`
}

type FacturacionPanamaFuenteNormativa struct {
	Titulo string `json:"titulo"`
	URL    string `json:"url"`
}

// FacturacionEcuadorChecklist resume los requisitos operativos propios del SRI Ecuador.
type FacturacionEcuadorChecklist struct {
	PaisCodigo           string                              `json:"pais_codigo"`
	Ok                   bool                                `json:"ok"`
	Estado               string                              `json:"estado"`
	Ambiente             string                              `json:"ambiente"`
	Integracion          string                              `json:"integracion"`
	Faltantes            []string                            `json:"faltantes"`
	Advertencias         []string                            `json:"advertencias"`
	DocumentosSoportados []string                            `json:"documentos_soportados"`
	Items                []FacturacionEcuadorChecklistItem   `json:"items"`
	Fuentes              []FacturacionEcuadorFuenteNormativa `json:"fuentes"`
}

type FacturacionEcuadorChecklistItem struct {
	Clave   string `json:"clave"`
	Titulo  string `json:"titulo"`
	Estado  string `json:"estado"`
	Detalle string `json:"detalle"`
}

type FacturacionEcuadorFuenteNormativa struct {
	Titulo string `json:"titulo"`
	URL    string `json:"url"`
}

// FacturacionDianDocumentoCatalogItem describe documentos y eventos del SFE Colombia.
type FacturacionDianDocumentoCatalogItem struct {
	Codigo               string `json:"codigo"`
	Titulo               string `json:"titulo"`
	Categoria            string `json:"categoria"`
	Alcance              string `json:"alcance"`
	ModuloSugerido       string `json:"modulo_sugerido"`
	EstadoImplementacion string `json:"estado_implementacion"`
	RequiereNumeracion   bool   `json:"requiere_numeracion"`
	RequiereFirma        bool   `json:"requiere_firma"`
	EsEvento             bool   `json:"es_evento"`
	Observacion          string `json:"observacion"`
}

// FacturacionDianObligacionContableItem lista obligaciones que suelen preparar contadores.
type FacturacionDianObligacionContableItem struct {
	Codigo       string `json:"codigo"`
	Titulo       string `json:"titulo"`
	Tipo         string `json:"tipo"`
	Envio        string `json:"envio"`
	Periodicidad string `json:"periodicidad"`
	Descripcion  string `json:"descripcion"`
}

type FacturacionDianFuenteNormativa struct {
	Titulo string `json:"titulo"`
	URL    string `json:"url"`
}

// FacturacionElectronicaPaisConfig define configuración FE por empresa y país.
type FacturacionElectronicaPaisConfig struct {
	ID                            int64  `json:"id"`
	EmpresaID                     int64  `json:"empresa_id"`
	PaisCodigo                    string `json:"pais_codigo"`
	PaisNombre                    string `json:"pais_nombre"`
	BanderaPais                   string `json:"bandera_pais,omitempty"`
	MonedaCodigo                  string `json:"moneda_codigo,omitempty"`
	Proveedor                     string `json:"proveedor,omitempty"`
	Ambiente                      string `json:"ambiente,omitempty"`
	TipoDocumentoEmisor           string `json:"tipo_documento_emisor,omitempty"`
	IdentificadorFiscal           string `json:"identificador_fiscal,omitempty"`
	RazonSocial                   string `json:"razon_social,omitempty"`
	EmailFacturacion              string `json:"email_facturacion,omitempty"`
	EnviarFacturaEmailClienteAuto bool   `json:"enviar_factura_email_cliente_auto"`
	TelefonoFacturacion           string `json:"telefono_facturacion,omitempty"`
	DireccionFiscal               string `json:"direccion_fiscal,omitempty"`
	PrefijoFactura                string `json:"prefijo_factura,omitempty"`
	ResolucionNumero              string `json:"resolucion_numero,omitempty"`
	APIBaseURL                    string `json:"api_base_url,omitempty"`
	CamposPaisJSON                string `json:"campos_pais_json,omitempty"`
	FechaCreacion                 string `json:"fecha_creacion,omitempty"`
	FechaActualizacion            string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador                string `json:"usuario_creador,omitempty"`
	Estado                        string `json:"estado,omitempty"`
	Observaciones                 string `json:"observaciones,omitempty"`
}

// PaisFacturacion representa un país soportado para FE.
type PaisFacturacion struct {
	Codigo  string `json:"codigo"`
	Nombre  string `json:"nombre"`
	Bandera string `json:"bandera"`
	Moneda  string `json:"moneda"`
}

// FacturacionPaisVista metadatos de presentación (UI). No fija lógica de envío;
// Ecuador y Panamá quedan aislados de Colombia (DIAN) y se documenta el enrutamiento.
type FacturacionPaisVista struct {
	EnteFiscal         string `json:"ente_fiscal"`
	NotaIndependencia  string `json:"nota_independencia"`
	ResumenOperativo   string `json:"resumen_operativo"`
	LabelResolucion    string `json:"label_resolucion"`
	LabelIdentificador string `json:"label_identificador"`
	LabelPrefijo       string `json:"label_prefijo"`
	PlaceholderRazon   string `json:"placeholder_razon_social"`
	ModuloDianRuta     string `json:"modulo_dian_ruta,omitempty"`
}

// FacturacionPaisVistaFor retorna textos y etiquetas para la configuración según el país (CO / EC / PA).
func FacturacionPaisVistaFor(codigo string) FacturacionPaisVista {
	switch normalizePaisCodigo(codigo) {
	case "EC":
		return FacturacionPaisVista{
			EnteFiscal:         "Ecuador (SRI): comprobantes de venta electrónicos, retenciones, notas y guías.",
			NotaIndependencia:  "Ecuador (SRI) no utiliza el módulo de la DIAN Colombia. Guarde RUC, establecimiento, punto de emisión y ambiente; la integración con un proveedor o el SRI se configura con proveedor y API base URL, sin afectar Colombia.",
			ResumenOperativo:   "Moneda habitual: USD. En campos estructurados: RUC, establecimiento, punto_emision, ambiente_sri (1=pruebas, 2=producción).",
			LabelResolucion:    "Autorización SRI o referencia (según comprobante)",
			LabelIdentificador: "RUC",
			LabelPrefijo:       "Establecimiento - punto (ej. 001-001)",
			PlaceholderRazon:   "Razón social inscrita en el SRI",
		}
	case "PA":
		return FacturacionPaisVista{
			EnteFiscal:         "Panamá (DGI): facturación electrónica y validación a través de PSE o proveedores homologados.",
			NotaIndependencia:  "Panamá (DGI) no utiliza el módulo de la DIAN Colombia. RUC, DV, folios y conexión con su PSE/ proveedor se definen solo en este perfil, sin tocar resoluciones DIAN.",
			ResumenOperativo:   "Moneda habitual: PAB o USD. En campos estructurados: ruc, dv, (opcionales) folio_inicial, codigo_ubicación según su proveedor.",
			LabelResolucion:    "Autorización o folio (según proveedor DGI / PSE)",
			LabelIdentificador: "RUC y dígito verificador (DV)",
			LabelPrefijo:       "Punto de expedición o prefijo de documento",
			PlaceholderRazon:   "Razón social o nombre fiscal inscrito en DGI",
		}
	case "CR":
		return FacturacionPaisVista{
			EnteFiscal:         "Costa Rica (Ministerio de Hacienda): comprobantes electronicos XML y recepcion mediante API protegida con OAuth/OIDC.",
			NotaIndependencia:  "Costa Rica no utiliza DIAN. El perfil conserva cedula juridica/fisica, sucursal, terminal, actividad economica y credenciales del proveedor o API de Hacienda sin afectar otros paises.",
			ResumenOperativo:   "Moneda habitual: CRC. En campos estructurados: cedula, tipo_identificacion, sucursal, terminal, actividad_economica, version_xml.",
			LabelResolucion:    "Clave/consecutivo o referencia de autorizacion Hacienda",
			LabelIdentificador: "Cedula juridica/fisica",
			LabelPrefijo:       "Sucursal-terminal o prefijo interno",
			PlaceholderRazon:   "Nombre fiscal registrado ante Hacienda",
		}
	case "AR":
		return FacturacionPaisVista{
			EnteFiscal:         "Argentina (ARCA): factura electronica mediante WSFEv1/WSMTXCA/Comprobantes en linea y CAE.",
			NotaIndependencia:  "Argentina no utiliza DIAN. Configure CUIT, punto de venta, condicion frente al IVA, tipo de comprobante y certificado del web service o proveedor homologado.",
			ResumenOperativo:   "Moneda habitual: ARS. En campos estructurados: cuit, punto_venta, condicion_iva, tipo_comprobante, ws_servicio.",
			LabelResolucion:    "CAE/CAEA o referencia ARCA",
			LabelIdentificador: "CUIT",
			LabelPrefijo:       "Punto de venta",
			PlaceholderRazon:   "Razon social registrada ante ARCA",
		}
	case "VE":
		return FacturacionPaisVista{
			EnteFiscal:         "Venezuela (SENIAT): facturacion digital mediante sistema homologado o imprenta digital autorizada.",
			NotaIndependencia:  "Venezuela no utiliza DIAN. Configure RIF, serie, proveedor homologado/imprenta digital y reglas de numeracion fiscal segun la providencia vigente.",
			ResumenOperativo:   "Moneda habitual: VES. En campos estructurados: rif, serie, imprenta_digital, proveedor_homologado, moneda_referencia.",
			LabelResolucion:    "Providencia/autorizacion o referencia SENIAT",
			LabelIdentificador: "RIF",
			LabelPrefijo:       "Serie fiscal",
			PlaceholderRazon:   "Razon social registrada ante SENIAT",
		}
	default: // CO
		return FacturacionPaisVista{
			EnteFiscal:         "Colombia (DIAN UBL 2.1): resolución de numeración, set de pruebas y transmisión en ambientes de habilitación o producción.",
			NotaIndependencia:  "La operación consecutiva, firma, Software ID/PIN, CUFE y el set de pruebas DIAN se configuran en esta sección de facturación electrónica y en la subpágina Pruebas DIAN y documentos; este perfil no se mezcla con Ecuador, Panamá u otros países.",
			ResumenOperativo:   "Moneda: COP. Vincule proveedor o API; en producción use credenciales homologadas con la DIAN.",
			LabelResolucion:    "Número de resolución de facturación / autorización DIAN",
			LabelIdentificador: "NIT (con dígito de verificación)",
			LabelPrefijo:       "Prefijo homologado (p. ej. SEFC)",
			PlaceholderRazon:   "Razón social inscrita ante la DIAN",
			ModuloDianRuta:     "/administrar_empresa/facturacion_electronica_pruebas_dian.html?empresa_id=",
		}
	}
}

// ListFacturacionPaisesConVista retorna el catálogo con metadatos de UI por país.
func ListFacturacionPaisesConVista() []map[string]interface{} {
	out := make([]map[string]interface{}, 0, 6)
	for _, p := range ListPaisesFacturacionDisponibles() {
		out = append(out, map[string]interface{}{
			"codigo":  p.Codigo,
			"nombre":  p.Nombre,
			"bandera": p.Bandera,
			"moneda":  p.Moneda,
			"vista":   FacturacionPaisVistaFor(p.Codigo),
		})
	}
	return out
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

// FacturacionElectronicaRetryItem representa un registro de cola para reintentos de integracion fiscal.
type FacturacionElectronicaRetryItem struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	TipoDocumento      string `json:"tipo_documento"`
	DocumentoCodigo    string `json:"documento_codigo"`
	PaisCodigo         string `json:"pais_codigo"`
	Proveedor          string `json:"proveedor"`
	Ambiente           string `json:"ambiente"`
	EstadoEnvio        string `json:"estado_envio"`
	Intentos           int64  `json:"intentos"`
	MaxIntentos        int64  `json:"max_intentos"`
	ProximoIntento     string `json:"proximo_intento,omitempty"`
	FechaUltimoIntento string `json:"fecha_ultimo_intento,omitempty"`
	UltimoError        string `json:"ultimo_error,omitempty"`
	RespuestaProveedor string `json:"respuesta_proveedor_json,omitempty"`
	ContingenciaActiva bool   `json:"contingencia_activa"`
	FechaContingencia  string `json:"fecha_contingencia,omitempty"`
	ReferenciaExterna  string `json:"referencia_externa,omitempty"`
	NumeroLegal        string `json:"numero_legal,omitempty"`
	CodigoValidacion   string `json:"codigo_validacion,omitempty"`
	FechaEmisionLegal  string `json:"fecha_emision_legal,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// FacturacionElectronicaRetryFilter define filtros para consultar cola de reintentos FE.
type FacturacionElectronicaRetryFilter struct {
	TipoDocumento   string
	EstadoEnvio     string
	DocumentoQuery  string
	SoloVencidos    bool
	IncludeInactive bool
	Limit           int
	Offset          int
}

func supportedPaisesFacturacionMap() map[string]PaisFacturacion {
	return map[string]PaisFacturacion{
		"AR": {Codigo: "AR", Nombre: "Argentina", Bandera: "AR", Moneda: "ARS"},
		"CO": {Codigo: "CO", Nombre: "Colombia", Bandera: "CO", Moneda: "COP"},
		"CR": {Codigo: "CR", Nombre: "Costa Rica", Bandera: "CR", Moneda: "CRC"},
		"EC": {Codigo: "EC", Nombre: "Ecuador", Bandera: "EC", Moneda: "USD"},
		"PA": {Codigo: "PA", Nombre: "Panama", Bandera: "PA", Moneda: "PAB"},
		"VE": {Codigo: "VE", Nombre: "Venezuela", Bandera: "VE", Moneda: "VES"},
	}
}

// ListPaisesFacturacionDisponibles retorna los paises FE soportados.
func ListPaisesFacturacionDisponibles() []PaisFacturacion {
	catalog := supportedPaisesFacturacionMap()
	return []PaisFacturacion{catalog["CO"], catalog["EC"], catalog["PA"], catalog["CR"], catalog["AR"], catalog["VE"]}
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

// EnsureEmpresaFacturacionElectronicaSchema crea/migra tabla FE por pais en PostgreSQL.
func EnsureEmpresaFacturacionElectronicaSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS facturacion_electronica_pais (
			id BIGSERIAL PRIMARY KEY,
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
			enviar_factura_email_cliente_auto INTEGER DEFAULT 0,
			telefono_facturacion TEXT,
			direccion_fiscal TEXT,
			prefijo_factura TEXT,
			resolucion_numero TEXT,
			api_base_url TEXT,
			campos_pais_json TEXT DEFAULT '{}',
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, pais_codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_fe_pais_empresa ON facturacion_electronica_pais(empresa_id, pais_codigo);`,
		`CREATE INDEX IF NOT EXISTS ix_fe_pais_estado ON facturacion_electronica_pais(empresa_id, estado);`,
		`CREATE TABLE IF NOT EXISTS facturacion_electronica_reintentos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			tipo_documento TEXT NOT NULL,
			documento_codigo TEXT NOT NULL,
			pais_codigo TEXT NOT NULL,
			proveedor TEXT,
			ambiente TEXT DEFAULT 'sandbox',
			estado_envio TEXT DEFAULT 'pendiente',
			intentos INTEGER DEFAULT 0,
			max_intentos INTEGER DEFAULT 5,
			proximo_intento TEXT,
			fecha_ultimo_intento TEXT,
			ultimo_error TEXT,
			respuesta_proveedor_json TEXT,
			contingencia_activa INTEGER DEFAULT 0,
			fecha_contingencia TEXT,
			referencia_externa TEXT,
			numero_legal TEXT,
			codigo_validacion TEXT,
			fecha_emision_legal TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, tipo_documento, documento_codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_fe_reintentos_empresa_estado ON facturacion_electronica_reintentos(empresa_id, estado_envio, estado);`,
		`CREATE INDEX IF NOT EXISTS ix_fe_reintentos_proximo_intento ON facturacion_electronica_reintentos(empresa_id, proximo_intento, estado_envio);`,
		`CREATE INDEX IF NOT EXISTS ix_fe_reintentos_documento ON facturacion_electronica_reintentos(empresa_id, tipo_documento, documento_codigo);`,
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
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_pais", "enviar_factura_email_cliente_auto", "INTEGER DEFAULT 0"); err != nil {
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

	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_reintentos", "empresa_id", "INTEGER NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_reintentos", "tipo_documento", "TEXT NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_reintentos", "documento_codigo", "TEXT NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_reintentos", "pais_codigo", "TEXT NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_reintentos", "proveedor", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_reintentos", "ambiente", "TEXT DEFAULT 'sandbox'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_reintentos", "estado_envio", "TEXT DEFAULT 'pendiente'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_reintentos", "intentos", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_reintentos", "max_intentos", "INTEGER DEFAULT 5"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_reintentos", "proximo_intento", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_reintentos", "fecha_ultimo_intento", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_reintentos", "ultimo_error", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_reintentos", "respuesta_proveedor_json", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_reintentos", "contingencia_activa", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_reintentos", "fecha_contingencia", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_reintentos", "referencia_externa", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_reintentos", "numero_legal", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_reintentos", "codigo_validacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_reintentos", "fecha_emision_legal", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_reintentos", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_reintentos", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_reintentos", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "facturacion_electronica_reintentos", "observaciones", "TEXT"); err != nil {
		return err
	}
	if _, err := dbConn.Exec(`CREATE INDEX IF NOT EXISTS ix_fe_reintentos_empresa_estado ON facturacion_electronica_reintentos(empresa_id, estado_envio, estado);`); err != nil {
		return err
	}
	if _, err := dbConn.Exec(`CREATE INDEX IF NOT EXISTS ix_fe_reintentos_proximo_intento ON facturacion_electronica_reintentos(empresa_id, proximo_intento, estado_envio);`); err != nil {
		return err
	}
	if _, err := dbConn.Exec(`CREATE INDEX IF NOT EXISTS ix_fe_reintentos_documento ON facturacion_electronica_reintentos(empresa_id, tipo_documento, documento_codigo);`); err != nil {
		return err
	}

	return nil
}

func defaultFacturacionConfig(empresaID int64, paisCodigo string) FacturacionElectronicaPaisConfig {
	pais := paisFacturacionByCodigo(paisCodigo)
	cfg := FacturacionElectronicaPaisConfig{
		EmpresaID:      empresaID,
		PaisCodigo:     pais.Codigo,
		PaisNombre:     pais.Nombre,
		BanderaPais:    pais.Bandera,
		MonedaCodigo:   pais.Moneda,
		Ambiente:       "sandbox",
		Estado:         "activo",
		CamposPaisJSON: "{}",
	}
	applyFacturacionPaisDefaults(&cfg)
	return cfg
}

func DefaultFacturacionDianDocumentosSoportados() []string {
	items := ListFacturacionDianDocumentosElectronicos()
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.Codigo)
	}
	return out
}

func ListFacturacionDianDocumentosElectronicos() []FacturacionDianDocumentoCatalogItem {
	return []FacturacionDianDocumentoCatalogItem{
		{Codigo: "factura_electronica", Titulo: "Factura electronica de venta", Categoria: "Venta", Alcance: "Venta de bienes o servicios validada previamente por DIAN.", ModuloSugerido: "ventas_simple/carritos", EstadoImplementacion: "base_operativa", RequiereNumeracion: true, RequiereFirma: true},
		{Codigo: "nota_credito", Titulo: "Nota credito electronica", Categoria: "Ajustes de venta", Alcance: "Disminuye, corrige o reversa valores de una factura electronica.", ModuloSugerido: "facturacion_electronica", EstadoImplementacion: "base_operativa", RequiereFirma: true},
		{Codigo: "nota_debito", Titulo: "Nota debito electronica", Categoria: "Ajustes de venta", Alcance: "Aumenta o corrige valores de una factura electronica.", ModuloSugerido: "facturacion_electronica", EstadoImplementacion: "base_operativa", RequiereFirma: true},
		{Codigo: "factura_talonario_contingencia", Titulo: "Reporte de factura de talonario o papel por contingencia", Categoria: "Contingencia", Alcance: "Reporte para validacion posterior cuando hubo inconveniente tecnologico del facturador.", ModuloSugerido: "facturacion_electronica/offline", EstadoImplementacion: "catalogado", RequiereNumeracion: true, RequiereFirma: true},
		{Codigo: "documento_soporte", Titulo: "Documento soporte en adquisiciones a no obligados", Categoria: "Compras", Alcance: "Soporta costos, deducciones o impuestos descontables en compras a sujetos no obligados a facturar.", ModuloSugerido: "compras", EstadoImplementacion: "base_operativa", RequiereNumeracion: true, RequiereFirma: true},
		{Codigo: "nota_ajuste_documento_soporte", Titulo: "Nota de ajuste del documento soporte", Categoria: "Compras", Alcance: "Ajusta o corrige un documento soporte de adquisiciones.", ModuloSugerido: "compras", EstadoImplementacion: "catalogado", RequiereFirma: true},
		{Codigo: "nomina_electronica", Titulo: "Documento soporte de pago de nomina electronica", Categoria: "Nomina", Alcance: "Soporta valores devengados, deducidos y pagados a empleados.", ModuloSugerido: "nomina", EstadoImplementacion: "base_operativa", RequiereFirma: true},
		{Codigo: "nota_ajuste_nomina_electronica", Titulo: "Nota de ajuste de nomina electronica", Categoria: "Nomina", Alcance: "Ajusta o corrige documentos soporte de pago de nomina electronica.", ModuloSugerido: "nomina", EstadoImplementacion: "catalogado", RequiereFirma: true},
		{Codigo: "documento_equivalente_pos", Titulo: "Documento equivalente electronico POS", Categoria: "Documentos equivalentes", Alcance: "Tiquete de maquina registradora con sistema POS transmitido para validacion.", ModuloSugerido: "pos/carritos", EstadoImplementacion: "base_operativa", RequiereNumeracion: true, RequiereFirma: true},
		{Codigo: "nota_ajuste_documento_equivalente", Titulo: "Nota de ajuste del documento equivalente electronico", Categoria: "Documentos equivalentes", Alcance: "Ajusta errores aritmeticos o de contenido en documentos equivalentes electronicos.", ModuloSugerido: "pos/carritos", EstadoImplementacion: "catalogado", RequiereFirma: true},
		{Codigo: "documento_equivalente_servicios_publicos", Titulo: "Documento equivalente electronico de servicios publicos", Categoria: "Documentos equivalentes", Alcance: "Documento electronico aplicable a servicios publicos domiciliarios.", ModuloSugerido: "plantillas/servicios_publicos", EstadoImplementacion: "catalogado", RequiereNumeracion: true, RequiereFirma: true},
		{Codigo: "documento_equivalente_transporte_pasajeros", Titulo: "Documento equivalente electronico de transporte de pasajeros", Categoria: "Documentos equivalentes", Alcance: "Tiquete de transporte de pasajeros cuando aplique la modalidad equivalente.", ModuloSugerido: "plantillas/transporte", EstadoImplementacion: "catalogado", RequiereNumeracion: true, RequiereFirma: true},
		{Codigo: "documento_equivalente_extracto", Titulo: "Documento equivalente electronico extracto", Categoria: "Documentos equivalentes", Alcance: "Extracto reconocido como documento equivalente electronico segun actividad.", ModuloSugerido: "contabilidad", EstadoImplementacion: "catalogado", RequiereNumeracion: true, RequiereFirma: true},
		{Codigo: "documento_equivalente_transporte_aereo", Titulo: "Documento equivalente electronico de transporte aereo", Categoria: "Documentos equivalentes", Alcance: "Tiquete o billete de transporte aereo de pasajeros.", ModuloSugerido: "plantillas/transporte", EstadoImplementacion: "catalogado", RequiereNumeracion: true, RequiereFirma: true},
		{Codigo: "documento_equivalente_juegos_suerte_azar", Titulo: "Documento equivalente electronico de juegos de suerte y azar", Categoria: "Documentos equivalentes", Alcance: "Boleta, fraccion, formulario, carton, billete o instrumento de juegos de suerte y azar no localizados.", ModuloSugerido: "plantillas/juegos", EstadoImplementacion: "catalogado", RequiereNumeracion: true, RequiereFirma: true},
		{Codigo: "documento_equivalente_juegos_localizados", Titulo: "Documento equivalente electronico de juegos localizados", Categoria: "Documentos equivalentes", Alcance: "Documento emitido para juegos localizados segun la actividad.", ModuloSugerido: "plantillas/juegos", EstadoImplementacion: "catalogado", RequiereNumeracion: true, RequiereFirma: true},
		{Codigo: "documento_equivalente_peajes", Titulo: "Documento equivalente electronico de peajes", Categoria: "Documentos equivalentes", Alcance: "Documento expedido para cobro de peajes.", ModuloSugerido: "plantillas/peajes", EstadoImplementacion: "catalogado", RequiereNumeracion: true, RequiereFirma: true},
		{Codigo: "documento_equivalente_bolsa_valores", Titulo: "Comprobante electronico de Bolsa de Valores", Categoria: "Documentos equivalentes", Alcance: "Comprobante de liquidacion de operaciones expedido por la Bolsa de Valores.", ModuloSugerido: "contabilidad/tesoreria", EstadoImplementacion: "catalogado", RequiereNumeracion: true, RequiereFirma: true},
		{Codigo: "documento_equivalente_bolsa_agropecuaria", Titulo: "Documento electronico de bolsa agropecuaria y commodities", Categoria: "Documentos equivalentes", Alcance: "Operaciones de bolsa agropecuaria y de otros commodities.", ModuloSugerido: "contabilidad/tesoreria", EstadoImplementacion: "catalogado", RequiereNumeracion: true, RequiereFirma: true},
		{Codigo: "documento_equivalente_espectaculos", Titulo: "Documento equivalente electronico de espectaculos publicos", Categoria: "Documentos equivalentes", Alcance: "Boleta de ingreso a espectaculos publicos de artes escenicas y otros espectaculos.", ModuloSugerido: "plantillas/eventos", EstadoImplementacion: "catalogado", RequiereNumeracion: true, RequiereFirma: true},
		{Codigo: "documento_equivalente_cine", Titulo: "Documento equivalente electronico de cine", Categoria: "Documentos equivalentes", Alcance: "Boleta de ingreso a cine.", ModuloSugerido: "plantillas/eventos", EstadoImplementacion: "catalogado", RequiereNumeracion: true, RequiereFirma: true},
		{Codigo: "eventos_radian_recepcion", Titulo: "Eventos de recepcion y aceptacion RADIAN", Categoria: "Eventos", Alcance: "Acuse de recibo, recibo de bienes o servicios, aceptacion, reclamo y trazabilidad de factura como titulo valor.", ModuloSugerido: "facturacion_electronica/radian", EstadoImplementacion: "operativo", RequiereFirma: true, EsEvento: true, Observacion: "Disponible como evento documental firmado desde el Centro de habilitacion DIAN mientras la empresa no active produccion; no es una venta nueva."},
	}
}

func ListFacturacionDianObligacionesContadores() []FacturacionDianObligacionContableItem {
	return []FacturacionDianObligacionContableItem{
		{Codigo: "declaraciones_tributarias", Titulo: "Declaraciones tributarias", Tipo: "Obligacion fiscal", Envio: "Servicios informaticos DIAN", Periodicidad: "Segun calendario tributario", Descripcion: "IVA, retencion, renta y otros formularios que no son documentos UBL de factura."},
		{Codigo: "informacion_exogena", Titulo: "Informacion exogena / medios magneticos", Tipo: "Reporte tributario", Envio: "Servicios informaticos DIAN", Periodicidad: "Anual o segun resolucion vigente", Descripcion: "Reportes de terceros, pagos, retenciones y saldos preparados desde la contabilidad."},
		{Codigo: "certificados_retencion", Titulo: "Certificados de retencion", Tipo: "Soporte contable", Envio: "Entrega a terceros y soporte fiscal", Periodicidad: "Anual o por solicitud", Descripcion: "Se preparan desde movimientos contables y retenciones practicadas."},
		{Codigo: "conciliacion_fiscal", Titulo: "Conciliacion fiscal y anexos", Tipo: "Soporte fiscal", Envio: "DIAN o archivo interno segun obligacion", Periodicidad: "Anual", Descripcion: "Relaciona saldos contables y fiscales; no reemplaza la factura electronica."},
	}
}

func ListFacturacionDianFuentesNormativas() []FacturacionDianFuenteNormativa {
	return []FacturacionDianFuenteNormativa{
		{Titulo: "DIAN - Abece Sistema de Factura Electronica", URL: "https://micrositios.dian.gov.co/sistema-de-facturacion-electronica/abece-sistema-de-factura-electronica/"},
		{Titulo: "DIAN - Documento Equivalente Electronico", URL: "https://micrositios.dian.gov.co/sistema-de-facturacion-electronica/documento-equivalente-electronico/"},
		{Titulo: "DIAN - Resolucion 000165 de 2023 compilada", URL: "https://normograma.dian.gov.co/dian/compilacion/docs/resolucion_dian_0165_2023.htm"},
		{Titulo: "DIAN - RADIAN", URL: "https://micrositios.dian.gov.co/sistema-de-facturacion-electronica/radian/"},
	}
}

func defaultCamposPaisJSON(paisCodigo string) string {
	fields := map[string]interface{}{"perfil_auto": true}
	switch normalizePaisCodigo(paisCodigo) {
	case "EC":
		fields["integracion"] = "sri_xml_firmado"
		fields["ruc"] = ""
		fields["ambiente_sri"] = "1"
		fields["establecimiento"] = "001"
		fields["punto_emision"] = "001"
		fields["tipo_emision"] = "normal"
		fields["certificado_firma_ref"] = ""
		fields["certificado_firma_confirmado"] = false
		fields["autorizacion_produccion_sri"] = false
		fields["clave_acceso"] = ""
		fields["numero_autorizacion"] = ""
		fields["ride"] = true
		fields["obligado_contabilidad"] = false
		fields["documentos_soportados"] = []string{"factura", "nota_credito", "nota_debito", "retencion", "guia_remision"}
		fields["fuente_normativa"] = "SRI Ecuador comprobantes electronicos"
	case "PA":
		fields["integracion"] = "dgi_pac_o_facturador"
		fields["ruc"] = ""
		fields["dv"] = ""
		fields["modalidad"] = "pac_o_facturador_gratuito"
		fields["registro_sfep"] = false
		fields["declaracion_jurada_sfep"] = false
		fields["certificado_firma_ref"] = ""
		fields["certificado_firma_confirmado"] = false
		fields["pac_nombre"] = ""
		fields["pac_id"] = ""
		fields["ambiente_sfep"] = "pruebas"
		fields["sucursal"] = "001"
		fields["punto_expedicion"] = "001"
		fields["cafe"] = ""
		fields["cufe"] = ""
		fields["qr_url"] = ""
		fields["documentos_soportados"] = []string{"factura_electronica", "nota_credito", "nota_debito"}
		fields["fuente_normativa"] = "DGI Panama SFEP"
	case "CR":
		fields["integracion"] = "hacienda_api_xml"
		fields["version_xml"] = "4.4"
		fields["sucursal"] = "001"
		fields["terminal"] = "00001"
		fields["actividad_economica"] = ""
	case "AR":
		fields["integracion"] = "arca_wsfev1"
		fields["punto_venta"] = "0001"
		fields["condicion_iva"] = ""
		fields["tipo_comprobante"] = "factura"
		fields["ws_servicio"] = "wsfev1"
	case "VE":
		fields["integracion"] = "seniat_facturacion_digital"
		fields["serie"] = "A"
		fields["imprenta_digital"] = ""
		fields["proveedor_homologado"] = ""
		fields["moneda_referencia"] = "VES"
	default:
		fields["integracion"] = "dian_ubl_2_1"
		fields["documentos_soportados"] = DefaultFacturacionDianDocumentosSoportados()
		fields["documentos_contadores_colombia"] = []string{"declaraciones_tributarias", "informacion_exogena", "certificados_retencion", "conciliacion_fiscal"}
		fields["documentos_dian_catalogo_version"] = "2026-05-20"
		fields["documentos_siigo_referencia"] = []string{"documento_soporte", "nota_credito_ventas", "nota_debito_ventas", "nomina_electronica", "pos_electronico"}
	}
	raw, _ := json.Marshal(fields)
	return string(raw)
}

func applyFacturacionPaisDefaults(cfg *FacturacionElectronicaPaisConfig) {
	if cfg == nil {
		return
	}
	pais := paisFacturacionByCodigo(cfg.PaisCodigo)
	if strings.TrimSpace(cfg.MonedaCodigo) == "" {
		cfg.MonedaCodigo = pais.Moneda
	}
	if strings.TrimSpace(cfg.Proveedor) == "" {
		switch pais.Codigo {
		case "EC":
			cfg.Proveedor = "sri_ecuador"
		case "PA":
			cfg.Proveedor = "dgi_panama_pac"
		case "CR":
			cfg.Proveedor = "hacienda_cr"
		case "AR":
			cfg.Proveedor = "arca_wsfev1"
		case "VE":
			cfg.Proveedor = "seniat_imprenta_digital"
		case "CO":
			cfg.Proveedor = "dian_colombia"
		default:
			cfg.Proveedor = "manual"
		}
	}
	if strings.TrimSpace(cfg.TipoDocumentoEmisor) == "" {
		switch pais.Codigo {
		case "AR":
			cfg.TipoDocumentoEmisor = "CUIT"
		case "CR":
			cfg.TipoDocumentoEmisor = "CEDULA"
		case "EC", "PA":
			cfg.TipoDocumentoEmisor = "RUC"
		case "VE":
			cfg.TipoDocumentoEmisor = "RIF"
		default:
			cfg.TipoDocumentoEmisor = "NIT"
		}
	}
	if strings.TrimSpace(cfg.PrefijoFactura) == "" {
		switch pais.Codigo {
		case "EC":
			cfg.PrefijoFactura = "001-001"
		case "PA":
			cfg.PrefijoFactura = "FE"
		case "CR":
			cfg.PrefijoFactura = "001-00001"
		case "AR":
			cfg.PrefijoFactura = "0001"
		case "VE":
			cfg.PrefijoFactura = "A"
		default:
			cfg.PrefijoFactura = "FE"
		}
	}
	if strings.TrimSpace(cfg.CamposPaisJSON) == "" || strings.TrimSpace(cfg.CamposPaisJSON) == "{}" {
		cfg.CamposPaisJSON = defaultCamposPaisJSON(pais.Codigo)
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
	applyFacturacionPaisDefaults(payload)
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
		enviar_factura_email_cliente_auto,
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
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?, ?)
	ON CONFLICT(empresa_id, pais_codigo) DO UPDATE SET
		pais_nombre = excluded.pais_nombre,
		moneda_codigo = excluded.moneda_codigo,
		proveedor = excluded.proveedor,
		ambiente = excluded.ambiente,
		tipo_documento_emisor = excluded.tipo_documento_emisor,
		identificador_fiscal = excluded.identificador_fiscal,
		razon_social = excluded.razon_social,
		email_facturacion = excluded.email_facturacion,
		enviar_factura_email_cliente_auto = excluded.enviar_factura_email_cliente_auto,
		telefono_facturacion = excluded.telefono_facturacion,
		direccion_fiscal = excluded.direccion_fiscal,
		prefijo_factura = excluded.prefijo_factura,
		resolucion_numero = excluded.resolucion_numero,
		api_base_url = excluded.api_base_url,
		campos_pais_json = excluded.campos_pais_json,
		fecha_actualizacion = CURRENT_TIMESTAMP,
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
		boolToInt(payload.EnviarFacturaEmailClienteAuto),
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
		COALESCE(enviar_factura_email_cliente_auto, 0),
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

	var enviarFacturaEmailClienteAutoInt int
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
		&enviarFacturaEmailClienteAutoInt,
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
	cfg.EnviarFacturaEmailClienteAuto = enviarFacturaEmailClienteAutoInt == 1
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
		COALESCE(enviar_factura_email_cliente_auto, 0),
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
		var enviarFacturaEmailClienteAutoInt int
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
			&enviarFacturaEmailClienteAutoInt,
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
		cfg.EnviarFacturaEmailClienteAuto = enviarFacturaEmailClienteAutoInt == 1
		out = append(out, cfg)
	}
	return out, nil
}

// getPaisFacturacionDesdeLicenciaActiva toma pais_codigo de la licencia activa vinculada a la empresa (señal fuerte de jurisdicción comercial).
func getPaisFacturacionDesdeLicenciaActiva(dbConn *sql.DB, empresaID int64) (string, error) {
	if dbConn == nil || empresaID <= 0 {
		return "", nil
	}
	ok, err := tableExists(dbConn, "licencias")
	if err != nil {
		return "", err
	}
	if !ok {
		return "", nil
	}
	var pais sql.NullString
	err = queryRowSQLCompat(dbConn, `SELECT COALESCE(pais_codigo, '') FROM licencias
		WHERE empresa_id = ? AND COALESCE(activo, 0) = 1
		ORDER BY id DESC
		LIMIT 1`, empresaID).Scan(&pais)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	if !pais.Valid {
		return "", nil
	}
	return strings.TrimSpace(pais.String), nil
}

func detectPaisByTimezone(tz string) string {
	tz = strings.ToLower(strings.TrimSpace(tz))
	switch {
	case strings.Contains(tz, "argentina"), strings.Contains(tz, "buenos_aires"), strings.Contains(tz, "cordoba"), strings.Contains(tz, "mendoza"):
		return "AR"
	case strings.Contains(tz, "costa_rica"):
		return "CR"
	case strings.Contains(tz, "caracas"):
		return "VE"
	case strings.Contains(tz, "panama"):
		return "PA"
	case strings.Contains(tz, "guayaquil"), strings.Contains(tz, "quito"), strings.Contains(tz, "galapagos"):
		return "EC"
	case strings.Contains(tz, "bogota"), strings.Contains(tz, "medellin"), strings.Contains(tz, "cartagena"):
		return "CO"
	default:
		return ""
	}
}

func detectPaisByLanguage(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	switch {
	case strings.HasPrefix(lang, "es-ar"):
		return "AR"
	case strings.HasPrefix(lang, "es-cr"):
		return "CR"
	case strings.HasPrefix(lang, "es-ve"):
		return "VE"
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

		pc, errLic := getPaisFacturacionDesdeLicenciaActiva(dbConn, empresaID)
		if errLic != nil {
			return PaisFacturacion{}, "", errLic
		}
		if pc != "" {
			if p := normalizePaisCodigo(pc); p != "" {
				if _, ok := supportedPaisesFacturacionMap()[p]; ok {
					return paisFacturacionByCodigo(p), "licencia_activa", nil
				}
			}
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
			fecha_actualizacion = CURRENT_TIMESTAMP
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

func normalizeFacturacionRetryEstado(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "pendiente":
		return "pendiente"
	case "fallido":
		return "fallido"
	case "enviado":
		return "enviado"
	case "reconciliado":
		return "reconciliado"
	case "contingencia":
		return "contingencia"
	case "no_aplica":
		return "no_aplica"
	default:
		return "pendiente"
	}
}

func normalizeFacturacionRetryItem(payload *FacturacionElectronicaRetryItem) {
	if payload == nil {
		return
	}
	payload.TipoDocumento = normalizeDocumentoTransaccionalTipo(payload.TipoDocumento, "factura_electronica")
	payload.DocumentoCodigo = normalizeDocumentoTransaccionalCodigo(payload.DocumentoCodigo)
	payload.PaisCodigo = normalizePaisCodigo(payload.PaisCodigo)
	payload.Proveedor = strings.TrimSpace(payload.Proveedor)
	if payload.Proveedor == "" {
		payload.Proveedor = "manual"
	}
	payload.Ambiente = normalizeAmbienteFEFromConfig(payload.Ambiente)
	payload.EstadoEnvio = normalizeFacturacionRetryEstado(payload.EstadoEnvio)
	if payload.Ambiente != "produccion" && payload.EstadoEnvio == "pendiente" {
		payload.EstadoEnvio = "no_aplica"
	}
	if payload.Intentos < 0 {
		payload.Intentos = 0
	}
	if payload.MaxIntentos <= 0 {
		payload.MaxIntentos = 5
	}
	if payload.MaxIntentos > 25 {
		payload.MaxIntentos = 25
	}
	payload.UltimoError = strings.TrimSpace(payload.UltimoError)
	payload.RespuestaProveedor = strings.TrimSpace(payload.RespuestaProveedor)
	payload.ReferenciaExterna = strings.TrimSpace(payload.ReferenciaExterna)
	payload.NumeroLegal = strings.TrimSpace(payload.NumeroLegal)
	payload.CodigoValidacion = strings.TrimSpace(payload.CodigoValidacion)
	payload.FechaEmisionLegal = strings.TrimSpace(payload.FechaEmisionLegal)
	payload.ProximoIntento = strings.TrimSpace(payload.ProximoIntento)
	payload.FechaUltimoIntento = strings.TrimSpace(payload.FechaUltimoIntento)
	payload.FechaContingencia = strings.TrimSpace(payload.FechaContingencia)
	payload.UsuarioCreador = strings.TrimSpace(payload.UsuarioCreador)
	payload.Estado = strings.ToLower(strings.TrimSpace(payload.Estado))
	if payload.Estado != "inactivo" {
		payload.Estado = "activo"
	}
	payload.Observaciones = strings.TrimSpace(payload.Observaciones)

	if payload.ProximoIntento == "" && (payload.EstadoEnvio == "pendiente" || payload.EstadoEnvio == "fallido") {
		payload.ProximoIntento = time.Now().In(time.Local).Format("2006-01-02 15:04:05")
	}

	payload.ContingenciaActiva = payload.ContingenciaActiva || payload.EstadoEnvio == "contingencia"
	if !payload.ContingenciaActiva {
		payload.FechaContingencia = ""
	}
}

func int64ToBoolFE(v int64) bool { return v > 0 }

// GetFacturacionElectronicaRetryByDocumento consulta el estado de integracion fiscal por documento FE.
func GetFacturacionElectronicaRetryByDocumento(dbConn *sql.DB, empresaID int64, tipoDocumento, documentoCodigo string) (*FacturacionElectronicaRetryItem, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	tipoDocumento = normalizeDocumentoTransaccionalTipo(tipoDocumento, "factura_electronica")
	documentoCodigo = normalizeDocumentoTransaccionalCodigo(documentoCodigo)
	if documentoCodigo == "" {
		return nil, fmt.Errorf("documento_codigo es obligatorio")
	}

	item := FacturacionElectronicaRetryItem{}
	var contingenciaActivaRaw int64
	err := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(tipo_documento, ''),
		COALESCE(documento_codigo, ''),
		COALESCE(pais_codigo, ''),
		COALESCE(proveedor, ''),
		COALESCE(ambiente, 'sandbox'),
		COALESCE(estado_envio, 'pendiente'),
		COALESCE(intentos, 0),
		COALESCE(max_intentos, 5),
		COALESCE(proximo_intento, ''),
		COALESCE(fecha_ultimo_intento, ''),
		COALESCE(ultimo_error, ''),
		COALESCE(respuesta_proveedor_json, ''),
		COALESCE(contingencia_activa, 0),
		COALESCE(fecha_contingencia, ''),
		COALESCE(referencia_externa, ''),
		COALESCE(numero_legal, ''),
		COALESCE(codigo_validacion, ''),
		COALESCE(fecha_emision_legal, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM facturacion_electronica_reintentos
	WHERE empresa_id = ? AND tipo_documento = ? AND documento_codigo = ?
	LIMIT 1`, empresaID, tipoDocumento, documentoCodigo).Scan(
		&item.ID,
		&item.EmpresaID,
		&item.TipoDocumento,
		&item.DocumentoCodigo,
		&item.PaisCodigo,
		&item.Proveedor,
		&item.Ambiente,
		&item.EstadoEnvio,
		&item.Intentos,
		&item.MaxIntentos,
		&item.ProximoIntento,
		&item.FechaUltimoIntento,
		&item.UltimoError,
		&item.RespuestaProveedor,
		&contingenciaActivaRaw,
		&item.FechaContingencia,
		&item.ReferenciaExterna,
		&item.NumeroLegal,
		&item.CodigoValidacion,
		&item.FechaEmisionLegal,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	)
	if err != nil {
		return nil, err
	}
	item.ContingenciaActiva = int64ToBoolFE(contingenciaActivaRaw)
	normalizeFacturacionRetryItem(&item)
	return &item, nil
}

// UpsertFacturacionElectronicaRetry crea/actualiza un registro de cola de reintentos FE por documento.
func UpsertFacturacionElectronicaRetry(dbConn *sql.DB, payload FacturacionElectronicaRetryItem) (*FacturacionElectronicaRetryItem, error) {
	if payload.EmpresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	normalizeFacturacionRetryItem(&payload)
	if payload.DocumentoCodigo == "" {
		return nil, fmt.Errorf("documento_codigo es obligatorio")
	}
	if payload.PaisCodigo == "" {
		return nil, fmt.Errorf("pais_codigo es obligatorio")
	}

	stmt := `INSERT INTO facturacion_electronica_reintentos (
		empresa_id,
		tipo_documento,
		documento_codigo,
		pais_codigo,
		proveedor,
		ambiente,
		estado_envio,
		intentos,
		max_intentos,
		proximo_intento,
		fecha_ultimo_intento,
		ultimo_error,
		respuesta_proveedor_json,
		contingencia_activa,
		fecha_contingencia,
		referencia_externa,
		numero_legal,
		codigo_validacion,
		fecha_emision_legal,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?, ?)
	ON CONFLICT(empresa_id, tipo_documento, documento_codigo) DO UPDATE SET
		pais_codigo = excluded.pais_codigo,
		proveedor = excluded.proveedor,
		ambiente = excluded.ambiente,
		estado_envio = excluded.estado_envio,
		intentos = excluded.intentos,
		max_intentos = excluded.max_intentos,
		proximo_intento = excluded.proximo_intento,
		fecha_ultimo_intento = excluded.fecha_ultimo_intento,
		ultimo_error = excluded.ultimo_error,
		respuesta_proveedor_json = excluded.respuesta_proveedor_json,
		contingencia_activa = excluded.contingencia_activa,
		fecha_contingencia = excluded.fecha_contingencia,
		referencia_externa = excluded.referencia_externa,
		numero_legal = CASE WHEN excluded.numero_legal <> '' THEN excluded.numero_legal ELSE facturacion_electronica_reintentos.numero_legal END,
		codigo_validacion = CASE WHEN excluded.codigo_validacion <> '' THEN excluded.codigo_validacion ELSE facturacion_electronica_reintentos.codigo_validacion END,
		fecha_emision_legal = CASE WHEN excluded.fecha_emision_legal <> '' THEN excluded.fecha_emision_legal ELSE facturacion_electronica_reintentos.fecha_emision_legal END,
		fecha_actualizacion = CURRENT_TIMESTAMP,
		usuario_creador = CASE WHEN excluded.usuario_creador <> '' THEN excluded.usuario_creador ELSE facturacion_electronica_reintentos.usuario_creador END,
		estado = excluded.estado,
		observaciones = excluded.observaciones`

	if _, err := dbConn.Exec(stmt,
		payload.EmpresaID,
		payload.TipoDocumento,
		payload.DocumentoCodigo,
		payload.PaisCodigo,
		payload.Proveedor,
		payload.Ambiente,
		payload.EstadoEnvio,
		payload.Intentos,
		payload.MaxIntentos,
		payload.ProximoIntento,
		payload.FechaUltimoIntento,
		payload.UltimoError,
		payload.RespuestaProveedor,
		boolToInt64(payload.ContingenciaActiva),
		payload.FechaContingencia,
		payload.ReferenciaExterna,
		payload.NumeroLegal,
		payload.CodigoValidacion,
		payload.FechaEmisionLegal,
		payload.UsuarioCreador,
		payload.Estado,
		payload.Observaciones,
	); err != nil {
		return nil, err
	}

	return GetFacturacionElectronicaRetryByDocumento(dbConn, payload.EmpresaID, payload.TipoDocumento, payload.DocumentoCodigo)
}

func buildFacturacionRetryQueryPattern(raw string) string {
	raw = strings.ToUpper(strings.TrimSpace(raw))
	raw = strings.ReplaceAll(raw, "%", "")
	raw = strings.ReplaceAll(raw, "_", "")
	if raw == "" {
		return "%"
	}
	return "%" + raw + "%"
}

// ListFacturacionElectronicaRetriesByEmpresa lista la cola de reintentos FE por empresa.
func ListFacturacionElectronicaRetriesByEmpresa(dbConn *sql.DB, empresaID int64, filter FacturacionElectronicaRetryFilter) ([]FacturacionElectronicaRetryItem, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}

	query := `SELECT
		id,
		empresa_id,
		COALESCE(tipo_documento, ''),
		COALESCE(documento_codigo, ''),
		COALESCE(pais_codigo, ''),
		COALESCE(proveedor, ''),
		COALESCE(ambiente, 'sandbox'),
		COALESCE(estado_envio, 'pendiente'),
		COALESCE(intentos, 0),
		COALESCE(max_intentos, 5),
		COALESCE(proximo_intento, ''),
		COALESCE(fecha_ultimo_intento, ''),
		COALESCE(ultimo_error, ''),
		COALESCE(respuesta_proveedor_json, ''),
		COALESCE(contingencia_activa, 0),
		COALESCE(fecha_contingencia, ''),
		COALESCE(referencia_externa, ''),
		COALESCE(numero_legal, ''),
		COALESCE(codigo_validacion, ''),
		COALESCE(fecha_emision_legal, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM facturacion_electronica_reintentos
	WHERE empresa_id = ?`

	args := make([]interface{}, 0, 8)
	args = append(args, empresaID)

	tipoDocumento := normalizeDocumentoTransaccionalTipo(filter.TipoDocumento, "")
	if tipoDocumento != "" {
		query += " AND tipo_documento = ?"
		args = append(args, tipoDocumento)
	}

	estadoEnvio := normalizeFacturacionRetryEstado(filter.EstadoEnvio)
	if strings.TrimSpace(filter.EstadoEnvio) != "" {
		query += " AND estado_envio = ?"
		args = append(args, estadoEnvio)
	}

	if q := strings.TrimSpace(filter.DocumentoQuery); q != "" {
		pattern := buildFacturacionRetryQueryPattern(q)
		query += " AND (UPPER(documento_codigo) LIKE ? OR UPPER(COALESCE(numero_legal, '')) LIKE ? OR UPPER(COALESCE(codigo_validacion, '')) LIKE ?)"
		args = append(args, pattern, pattern, pattern)
	}

	if filter.SoloVencidos {
		query += " AND estado_envio IN ('pendiente','fallido','contingencia') AND (COALESCE(proximo_intento, '') = '' OR proximo_intento <= CURRENT_TIMESTAMP)"
	}

	if !filter.IncludeInactive {
		query += " AND estado = 'activo'"
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 200
	}
	if limit > 500 {
		limit = 500
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query += " ORDER BY CASE estado_envio WHEN 'pendiente' THEN 0 WHEN 'fallido' THEN 1 WHEN 'contingencia' THEN 2 WHEN 'enviado' THEN 3 ELSE 4 END, COALESCE(proximo_intento, ''), id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]FacturacionElectronicaRetryItem, 0)
	for rows.Next() {
		it := FacturacionElectronicaRetryItem{}
		var contingenciaRaw int64
		if err := rows.Scan(
			&it.ID,
			&it.EmpresaID,
			&it.TipoDocumento,
			&it.DocumentoCodigo,
			&it.PaisCodigo,
			&it.Proveedor,
			&it.Ambiente,
			&it.EstadoEnvio,
			&it.Intentos,
			&it.MaxIntentos,
			&it.ProximoIntento,
			&it.FechaUltimoIntento,
			&it.UltimoError,
			&it.RespuestaProveedor,
			&contingenciaRaw,
			&it.FechaContingencia,
			&it.ReferenciaExterna,
			&it.NumeroLegal,
			&it.CodigoValidacion,
			&it.FechaEmisionLegal,
			&it.FechaCreacion,
			&it.FechaActualizacion,
			&it.UsuarioCreador,
			&it.Estado,
			&it.Observaciones,
		); err != nil {
			return nil, err
		}
		it.ContingenciaActiva = int64ToBoolFE(contingenciaRaw)
		normalizeFacturacionRetryItem(&it)
		items = append(items, it)
	}

	return items, nil
}

func facturacionPanamaJSONMap(raw string) map[string]interface{} {
	out := map[string]interface{}{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return out
	}
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func facturacionPanamaString(extra map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value, ok := extra[key]; ok {
			text := strings.TrimSpace(fmt.Sprintf("%v", value))
			if text != "" && text != "<nil>" {
				return text
			}
		}
	}
	return ""
}

func facturacionPanamaBool(extra map[string]interface{}, key string) bool {
	value, ok := extra[key]
	if !ok {
		return false
	}
	switch v := value.(type) {
	case bool:
		return v
	case float64:
		return v != 0
	case int:
		return v != 0
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "1", "true", "si", "sÃ­", "yes", "activo", "ok", "validado":
			return true
		}
	}
	return false
}

func facturacionPanamaDocumentos(extra map[string]interface{}) []string {
	if raw, ok := extra["documentos_soportados"]; ok {
		switch list := raw.(type) {
		case []interface{}:
			out := make([]string, 0, len(list))
			for _, value := range list {
				if text := strings.TrimSpace(fmt.Sprintf("%v", value)); text != "" && text != "<nil>" {
					out = append(out, text)
				}
			}
			if len(out) > 0 {
				return out
			}
		case []string:
			if len(list) > 0 {
				return list
			}
		case string:
			parts := strings.Split(list, ",")
			out := make([]string, 0, len(parts))
			for _, part := range parts {
				if text := strings.TrimSpace(part); text != "" {
					out = append(out, text)
				}
			}
			if len(out) > 0 {
				return out
			}
		}
	}
	return []string{"factura_electronica", "nota_credito", "nota_debito"}
}

func facturacionPanamaItem(ok bool, clave, titulo, detalle string) FacturacionPanamaChecklistItem {
	estado := "pendiente"
	if ok {
		estado = "ok"
	}
	return FacturacionPanamaChecklistItem{Clave: clave, Titulo: titulo, Estado: estado, Detalle: detalle}
}

// BuildFacturacionPanamaChecklist valida el perfil PA sin mezclarlo con DIAN Colombia.
func BuildFacturacionPanamaChecklist(cfg *FacturacionElectronicaPaisConfig) FacturacionPanamaChecklist {
	extra := map[string]interface{}{}
	if cfg != nil {
		extra = facturacionPanamaJSONMap(cfg.CamposPaisJSON)
	}
	modalidad := strings.ToLower(strings.TrimSpace(facturacionPanamaString(extra, "modalidad")))
	if modalidad == "" {
		modalidad = "pac_o_facturador_gratuito"
	}
	ambiente := "sandbox"
	if cfg != nil && strings.TrimSpace(cfg.Ambiente) != "" {
		ambiente = strings.ToLower(strings.TrimSpace(cfg.Ambiente))
	}
	if ambiente != "produccion" {
		ambiente = "sandbox"
	}

	identificador := ""
	razonSocial := ""
	proveedor := ""
	apiBaseURL := ""
	if cfg != nil {
		identificador = strings.TrimSpace(cfg.IdentificadorFiscal)
		razonSocial = strings.TrimSpace(cfg.RazonSocial)
		proveedor = strings.TrimSpace(cfg.Proveedor)
		apiBaseURL = strings.TrimSpace(cfg.APIBaseURL)
	}
	if identificador == "" {
		identificador = facturacionPanamaString(extra, "ruc")
	}
	dv := facturacionPanamaString(extra, "dv", "digito_verificador")
	registroSFEP := facturacionPanamaBool(extra, "registro_sfep")
	declaracionJurada := facturacionPanamaBool(extra, "declaracion_jurada_sfep")
	firmaRef := facturacionPanamaString(extra, "certificado_firma_ref", "firma_ref", "certificado_url")
	firmaConfirmada := facturacionPanamaBool(extra, "certificado_firma_confirmado") || firmaRef != ""
	pacNombre := facturacionPanamaString(extra, "pac_nombre", "pac_id")
	usaPAC := strings.Contains(modalidad, "pac") || strings.TrimSpace(proveedor) != "" && !strings.Contains(modalidad, "gratuito")
	pacOK := !usaPAC || pacNombre != "" || apiBaseURL != "" || strings.TrimSpace(proveedor) != ""

	items := []FacturacionPanamaChecklistItem{
		facturacionPanamaItem(identificador != "", "ruc", "RUC del emisor", "Identificador fiscal inscrito en DGI/e-Tax2.0."),
		facturacionPanamaItem(dv != "", "dv", "Digito verificador", "DV del RUC del contribuyente."),
		facturacionPanamaItem(razonSocial != "", "razon_social", "Razon social", "Nombre fiscal del emisor registrado ante DGI."),
		facturacionPanamaItem(registroSFEP, "registro_sfep", "Registro en SFEP", "Contribuyente registrado en el Sistema de Factura Electronica de Panama."),
		facturacionPanamaItem(declaracionJurada, "declaracion_jurada_sfep", "Declaracion jurada SFEP", "Declaracion jurada/solicitud del sistema completada desde e-Tax2.0."),
		facturacionPanamaItem(firmaConfirmada, "certificado_firma", "Firma electronica", "Certificado o referencia de firma electronica configurado para firmar documentos."),
		facturacionPanamaItem(pacOK, "modalidad", "Modalidad PAC o Facturador Gratuito", "Modalidad operativa seleccionada; si usa PAC debe existir proveedor/API o referencia PAC."),
	}

	faltantes := make([]string, 0)
	for _, item := range items {
		if item.Estado != "ok" {
			faltantes = append(faltantes, item.Clave)
		}
	}
	advertencias := make([]string, 0)
	if ambiente == "produccion" && apiBaseURL == "" && !strings.Contains(modalidad, "gratuito") {
		advertencias = append(advertencias, "En produccion con PAC debe configurar API base URL o referencia de integracion del proveedor.")
	}
	if moneda := ""; cfg != nil {
		moneda = strings.ToUpper(strings.TrimSpace(cfg.MonedaCodigo))
		if moneda != "" && moneda != "PAB" && moneda != "USD" {
			advertencias = append(advertencias, "Panama opera normalmente en PAB o USD; revise la moneda configurada.")
		}
	}

	estado := "listo"
	if len(faltantes) > 0 {
		estado = "pendiente"
	}
	return FacturacionPanamaChecklist{
		PaisCodigo:           "PA",
		Ok:                   len(faltantes) == 0,
		Estado:               estado,
		Modalidad:            modalidad,
		Ambiente:             ambiente,
		Faltantes:            faltantes,
		Advertencias:         advertencias,
		DocumentosSoportados: facturacionPanamaDocumentos(extra),
		Items:                items,
		Fuentes: []FacturacionPanamaFuenteNormativa{
			{Titulo: "DGI Panama - Factura Electronica", URL: "https://dgi.mef.gob.pa/_7facturaelectronica/felectronica"},
			{Titulo: "DGI Panama - e-Tax2.0", URL: "https://etax2.mef.gob.pa/etax2web/Login.aspx"},
		},
	}
}

func facturacionEcuadorJSONMap(raw string) map[string]interface{} {
	return facturacionPanamaJSONMap(raw)
}

func facturacionEcuadorString(extra map[string]interface{}, keys ...string) string {
	return facturacionPanamaString(extra, keys...)
}

func facturacionEcuadorBool(extra map[string]interface{}, key string) bool {
	return facturacionPanamaBool(extra, key)
}

func facturacionEcuadorDocumentos(extra map[string]interface{}) []string {
	if raw, ok := extra["documentos_soportados"]; ok {
		switch list := raw.(type) {
		case []interface{}:
			out := make([]string, 0, len(list))
			for _, value := range list {
				if text := strings.TrimSpace(fmt.Sprintf("%v", value)); text != "" && text != "<nil>" {
					out = append(out, text)
				}
			}
			if len(out) > 0 {
				return out
			}
		case []string:
			if len(list) > 0 {
				return list
			}
		case string:
			parts := strings.Split(list, ",")
			out := make([]string, 0, len(parts))
			for _, part := range parts {
				if text := strings.TrimSpace(part); text != "" {
					out = append(out, text)
				}
			}
			if len(out) > 0 {
				return out
			}
		}
	}
	return []string{"factura", "nota_credito", "nota_debito", "retencion", "guia_remision"}
}

func facturacionEcuadorItem(ok bool, clave, titulo, detalle string) FacturacionEcuadorChecklistItem {
	estado := "pendiente"
	if ok {
		estado = "ok"
	}
	return FacturacionEcuadorChecklistItem{Clave: clave, Titulo: titulo, Estado: estado, Detalle: detalle}
}

// BuildFacturacionEcuadorChecklist valida el perfil EC sin mezclarlo con DIAN Colombia ni DGI Panama.
func BuildFacturacionEcuadorChecklist(cfg *FacturacionElectronicaPaisConfig) FacturacionEcuadorChecklist {
	extra := map[string]interface{}{}
	if cfg != nil {
		extra = facturacionEcuadorJSONMap(cfg.CamposPaisJSON)
	}
	ambiente := "sandbox"
	if cfg != nil && strings.TrimSpace(cfg.Ambiente) != "" {
		ambiente = strings.ToLower(strings.TrimSpace(cfg.Ambiente))
	}
	if ambiente != "produccion" {
		ambiente = "sandbox"
	}
	ambienteSRI := facturacionEcuadorString(extra, "ambiente_sri")
	if ambienteSRI == "" {
		if ambiente == "produccion" {
			ambienteSRI = "2"
		} else {
			ambienteSRI = "1"
		}
	}
	integracion := facturacionEcuadorString(extra, "integracion")
	if integracion == "" {
		integracion = "sri_xml_firmado"
	}

	identificador := ""
	razonSocial := ""
	proveedor := ""
	apiBaseURL := ""
	if cfg != nil {
		identificador = strings.TrimSpace(cfg.IdentificadorFiscal)
		razonSocial = strings.TrimSpace(cfg.RazonSocial)
		proveedor = strings.TrimSpace(cfg.Proveedor)
		apiBaseURL = strings.TrimSpace(cfg.APIBaseURL)
	}
	if identificador == "" {
		identificador = facturacionEcuadorString(extra, "ruc")
	}
	establecimiento := facturacionEcuadorString(extra, "establecimiento", "estab")
	puntoEmision := facturacionEcuadorString(extra, "punto_emision", "pto_emi")
	firmaRef := facturacionEcuadorString(extra, "certificado_firma_ref", "firma_ref", "certificado_url")
	firmaConfirmada := facturacionEcuadorBool(extra, "certificado_firma_confirmado") || firmaRef != ""
	produccionAutorizada := facturacionEcuadorBool(extra, "autorizacion_produccion_sri")
	tieneProveedor := strings.TrimSpace(proveedor) != "" || strings.TrimSpace(apiBaseURL) != "" || strings.Contains(integracion, "facturador_sri")

	items := []FacturacionEcuadorChecklistItem{
		facturacionEcuadorItem(identificador != "", "ruc", "RUC del emisor", "Identificador fiscal registrado ante el SRI."),
		facturacionEcuadorItem(razonSocial != "", "razon_social", "Razon social", "Nombre fiscal del emisor registrado ante el SRI."),
		facturacionEcuadorItem(establecimiento != "", "establecimiento", "Establecimiento", "Codigo de establecimiento para la secuencia del comprobante."),
		facturacionEcuadorItem(puntoEmision != "", "punto_emision", "Punto de emision", "Punto de emision usado en la secuencia del comprobante electronico."),
		facturacionEcuadorItem(firmaConfirmada, "certificado_firma", "Firma electronica", "Certificado de firma electronica vigente para firmar XML."),
		facturacionEcuadorItem(tieneProveedor, "integracion", "Facturador SRI, proveedor o API", "Sistema propio/proveedor configurado para generar, firmar, enviar al SRI y notificar documentos."),
	}
	if ambiente == "produccion" {
		items = append(items, facturacionEcuadorItem(produccionAutorizada, "autorizacion_produccion_sri", "Autorizacion en produccion", "Emisor autorizado para ambiente de produccion en SRI en Linea."))
	}

	faltantes := make([]string, 0)
	for _, item := range items {
		if item.Estado != "ok" {
			faltantes = append(faltantes, item.Clave)
		}
	}
	advertencias := make([]string, 0)
	if ambiente == "produccion" && ambienteSRI != "2" {
		advertencias = append(advertencias, "Para produccion en SRI el ambiente_sri debe ser 2.")
	}
	if ambiente == "sandbox" && ambienteSRI != "1" {
		advertencias = append(advertencias, "Para pruebas en SRI el ambiente_sri debe ser 1.")
	}
	if moneda := ""; cfg != nil {
		moneda = strings.ToUpper(strings.TrimSpace(cfg.MonedaCodigo))
		if moneda != "" && moneda != "USD" {
			advertencias = append(advertencias, "Ecuador usa USD como moneda operativa habitual; revise la moneda configurada.")
		}
	}

	estado := "listo"
	if len(faltantes) > 0 {
		estado = "pendiente"
	}
	return FacturacionEcuadorChecklist{
		PaisCodigo:           "EC",
		Ok:                   len(faltantes) == 0,
		Estado:               estado,
		Ambiente:             ambiente,
		Integracion:          integracion,
		Faltantes:            faltantes,
		Advertencias:         advertencias,
		DocumentosSoportados: facturacionEcuadorDocumentos(extra),
		Items:                items,
		Fuentes: []FacturacionEcuadorFuenteNormativa{
			{Titulo: "SRI Ecuador - Facturacion Electronica", URL: "https://www.sri.gob.ec/facturacion-electronica"},
			{Titulo: "SRI Ecuador - Comprobantes Electronicos", URL: "https://www.sri.gob.ec/comprobantes-electronicos"},
		},
	}
}
