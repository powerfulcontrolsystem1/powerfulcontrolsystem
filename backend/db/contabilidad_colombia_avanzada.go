package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type EmpresaContabilidadAvanzadaDashboard struct {
	EmpresaID             int64                         `json:"empresa_id"`
	FormatosExogena       int                           `json:"formatos_exogena"`
	RegistrosExogena      int                           `json:"registros_exogena"`
	NominasElectronicas   int                           `json:"nominas_electronicas"`
	DocumentosSoporte     int                           `json:"documentos_soporte"`
	ActivosFijos          int                           `json:"activos_fijos"`
	CarteraCXPendientes   int                           `json:"cartera_cx_pendientes"`
	LibrosDisponibles     []EmpresaLibroOficialResumen  `json:"libros_disponibles"`
	UltimosDocumentosDIAN []EmpresaDocumentoDIANResumen `json:"ultimos_documentos_dian"`
}

type EmpresaDocumentoDIANResumen struct {
	Modulo    string  `json:"modulo"`
	Codigo    string  `json:"codigo"`
	Tercero   string  `json:"tercero"`
	Periodo   string  `json:"periodo"`
	Total     float64 `json:"total"`
	Estado    string  `json:"estado"`
	Fecha     string  `json:"fecha"`
	Respuesta string  `json:"respuesta,omitempty"`
}

type EmpresaExogenaFormato struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Formato            string `json:"formato"`
	Version            string `json:"version"`
	AnioGravable       int    `json:"anio_gravable"`
	Concepto           string `json:"concepto"`
	Descripcion        string `json:"descripcion"`
	Periodicidad       string `json:"periodicidad"`
	Estado             string `json:"estado"`
	UltimaGeneracion   string `json:"ultima_generacion,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
}

type EmpresaExogenaRegistro struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	FormatoID          int64   `json:"formato_id"`
	Formato            string  `json:"formato,omitempty"`
	TerceroID          int64   `json:"tercero_id,omitempty"`
	TipoDocumento      string  `json:"tipo_documento"`
	Documento          string  `json:"documento"`
	DigitoVerificacion string  `json:"digito_verificacion,omitempty"`
	RazonSocial        string  `json:"razon_social"`
	Concepto           string  `json:"concepto"`
	CuentaCodigo       string  `json:"cuenta_codigo"`
	BaseValor          float64 `json:"base_valor"`
	IVA                float64 `json:"iva"`
	Retencion          float64 `json:"retencion"`
	Total              float64 `json:"total"`
	Validaciones       string  `json:"validaciones,omitempty"`
	Estado             string  `json:"estado"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
}

type EmpresaNominaElectronica struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	EmpleadoID         int64   `json:"empleado_id,omitempty"`
	TipoDocumento      string  `json:"tipo_documento"`
	Documento          string  `json:"documento"`
	Nombre             string  `json:"nombre"`
	Periodo            string  `json:"periodo"`
	FechaPago          string  `json:"fecha_pago"`
	SalarioBase        float64 `json:"salario_base"`
	Devengados         float64 `json:"devengados"`
	Deducciones        float64 `json:"deducciones"`
	Total              float64 `json:"total"`
	CUNE               string  `json:"cune,omitempty"`
	EstadoDIAN         string  `json:"estado_dian"`
	RespuestaDIAN      string  `json:"respuesta_dian,omitempty"`
	JSONPayload        string  `json:"json_payload,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
}

type EmpresaDocumentoSoporteElectronico struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	ProveedorID        int64   `json:"proveedor_id,omitempty"`
	TipoDocumento      string  `json:"tipo_documento"`
	Documento          string  `json:"documento"`
	NombreProveedor    string  `json:"nombre_proveedor"`
	FechaDocumento     string  `json:"fecha_documento"`
	Periodo            string  `json:"periodo"`
	Concepto           string  `json:"concepto"`
	Subtotal           float64 `json:"subtotal"`
	IVA                float64 `json:"iva"`
	Retenciones        float64 `json:"retenciones"`
	Total              float64 `json:"total"`
	CUDS               string  `json:"cuds,omitempty"`
	EstadoDIAN         string  `json:"estado_dian"`
	RespuestaDIAN      string  `json:"respuesta_dian,omitempty"`
	JSONPayload        string  `json:"json_payload,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
}

type EmpresaActivoFijo struct {
	ID                          int64   `json:"id"`
	EmpresaID                   int64   `json:"empresa_id"`
	Codigo                      string  `json:"codigo"`
	Nombre                      string  `json:"nombre"`
	Categoria                   string  `json:"categoria"`
	Serial                      string  `json:"serial,omitempty"`
	Placa                       string  `json:"placa,omitempty"`
	FechaCompra                 string  `json:"fecha_compra"`
	Costo                       float64 `json:"costo"`
	ValorResidual               float64 `json:"valor_residual"`
	VidaUtilMeses               int     `json:"vida_util_meses"`
	MetodoDepreciacion          string  `json:"metodo_depreciacion"`
	FechaInicioDepreciacion     string  `json:"fecha_inicio_depreciacion,omitempty"`
	DepreciacionMensual         float64 `json:"depreciacion_mensual"`
	DepreciacionAcumulada       float64 `json:"depreciacion_acumulada"`
	ValorLibros                 float64 `json:"valor_libros"`
	BaseFiscal                  float64 `json:"base_fiscal"`
	VidaUtilFiscalMeses         int     `json:"vida_util_fiscal_meses"`
	MetodoDepreciacionFiscal    string  `json:"metodo_depreciacion_fiscal"`
	DepreciacionMensualFiscal   float64 `json:"depreciacion_mensual_fiscal"`
	DepreciacionAcumuladaFiscal float64 `json:"depreciacion_acumulada_fiscal"`
	ValorFiscal                 float64 `json:"valor_fiscal"`
	DiferenciaNIIFFiscal        float64 `json:"diferencia_niif_fiscal"`
	DeterioroAcumulado          float64 `json:"deterioro_acumulado"`
	ValorRazonable              float64 `json:"valor_razonable"`
	FechaUltimaValoracion       string  `json:"fecha_ultima_valoracion,omitempty"`
	CuentaDeterioro             string  `json:"cuenta_deterioro,omitempty"`
	CuentaActivo                string  `json:"cuenta_activo"`
	CuentaDepreciacion          string  `json:"cuenta_depreciacion"`
	CuentaGasto                 string  `json:"cuenta_gasto"`
	Ubicacion                   string  `json:"ubicacion,omitempty"`
	Responsable                 string  `json:"responsable,omitempty"`
	CentroCosto                 string  `json:"centro_costo,omitempty"`
	Proveedor                   string  `json:"proveedor,omitempty"`
	ValorAsegurado              float64 `json:"valor_asegurado,omitempty"`
	Poliza                      string  `json:"poliza,omitempty"`
	MantenimientoCadaDias       int     `json:"mantenimiento_cada_dias,omitempty"`
	UltimoMantenimiento         string  `json:"ultimo_mantenimiento,omitempty"`
	ProximoMantenimiento        string  `json:"proximo_mantenimiento,omitempty"`
	EstadoOperativo             string  `json:"estado_operativo,omitempty"`
	FechaBaja                   string  `json:"fecha_baja,omitempty"`
	MotivoBaja                  string  `json:"motivo_baja,omitempty"`
	ValorBaja                   float64 `json:"valor_baja,omitempty"`
	Estado                      string  `json:"estado"`
	FechaCreacion               string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion          string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador              string  `json:"usuario_creador,omitempty"`
}

type EmpresaActivoDepreciacion struct {
	ID                    int64   `json:"id"`
	EmpresaID             int64   `json:"empresa_id"`
	ActivoID              int64   `json:"activo_id"`
	ActivoCodigo          string  `json:"activo_codigo,omitempty"`
	ActivoNombre          string  `json:"activo_nombre,omitempty"`
	Periodo               string  `json:"periodo"`
	FechaCalculo          string  `json:"fecha_calculo"`
	Metodo                string  `json:"metodo"`
	BaseDepreciable       float64 `json:"base_depreciable"`
	DepreciacionPeriodo   float64 `json:"depreciacion_periodo"`
	DepreciacionAcumulada float64 `json:"depreciacion_acumulada"`
	ValorLibros           float64 `json:"valor_libros"`
	AsientoContableID     int64   `json:"asiento_contable_id,omitempty"`
	Estado                string  `json:"estado"`
	FechaCreacion         string  `json:"fecha_creacion,omitempty"`
	UsuarioCreador        string  `json:"usuario_creador,omitempty"`
}

type EmpresaActivoEvento struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	ActivoID           int64   `json:"activo_id"`
	ActivoCodigo       string  `json:"activo_codigo,omitempty"`
	ActivoNombre       string  `json:"activo_nombre,omitempty"`
	Tipo               string  `json:"tipo"`
	FechaEvento        string  `json:"fecha_evento"`
	UbicacionOrigen    string  `json:"ubicacion_origen,omitempty"`
	UbicacionDestino   string  `json:"ubicacion_destino,omitempty"`
	ResponsableOrigen  string  `json:"responsable_origen,omitempty"`
	ResponsableDestino string  `json:"responsable_destino,omitempty"`
	Valor              float64 `json:"valor,omitempty"`
	Estado             string  `json:"estado"`
	Detalle            string  `json:"detalle,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
}

type EmpresaActivosFijosAvanzadoResumen struct {
	EmpresaID                  int64                       `json:"empresa_id"`
	ActivosActivos             int                         `json:"activos_activos"`
	ActivosBaja                int                         `json:"activos_baja"`
	CostoTotal                 float64                     `json:"costo_total"`
	ValorLibrosTotal           float64                     `json:"valor_libros_total"`
	ValorFiscalTotal           float64                     `json:"valor_fiscal_total"`
	DiferenciaNIIFFiscalTotal  float64                     `json:"diferencia_niif_fiscal_total"`
	DeterioroAcumuladoTotal    float64                     `json:"deterioro_acumulado_total"`
	DepreciacionAcumuladaTotal float64                     `json:"depreciacion_acumulada_total"`
	DepreciacionPeriodoTotal   float64                     `json:"depreciacion_periodo_total"`
	MantenimientosPendientes   int                         `json:"mantenimientos_pendientes"`
	Depreciaciones             []EmpresaActivoDepreciacion `json:"depreciaciones"`
	UltimosEventos             []EmpresaActivoEvento       `json:"ultimos_eventos"`
}

type EmpresaCarteraCXP struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	Tipo               string  `json:"tipo"`
	TerceroID          int64   `json:"tercero_id,omitempty"`
	TerceroNombre      string  `json:"tercero_nombre"`
	Documento          string  `json:"documento"`
	FechaEmision       string  `json:"fecha_emision"`
	FechaVencimiento   string  `json:"fecha_vencimiento"`
	CuentaCodigo       string  `json:"cuenta_codigo"`
	Concepto           string  `json:"concepto"`
	ValorOriginal      float64 `json:"valor_original"`
	ValorPagado        float64 `json:"valor_pagado"`
	Saldo              float64 `json:"saldo"`
	Estado             string  `json:"estado"`
	OrigenModulo       string  `json:"origen_modulo"`
	ReferenciaExterna  string  `json:"referencia_externa,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
}

type EmpresaCarteraCXPEdadFila struct {
	Tipo       string  `json:"tipo"`
	Rango      string  `json:"rango"`
	Registros  int64   `json:"registros"`
	Saldo      float64 `json:"saldo"`
	Vencido    float64 `json:"vencido"`
	PorVencer  float64 `json:"por_vencer"`
	SaldoMayor float64 `json:"saldo_mayor"`
}

type EmpresaCarteraCXPEdadesResumen struct {
	EmpresaID      int64                       `json:"empresa_id"`
	Tipo           string                      `json:"tipo,omitempty"`
	FechaCorte     string                      `json:"fecha_corte"`
	TotalRegistros int64                       `json:"total_registros"`
	SaldoTotal     float64                     `json:"saldo_total"`
	VencidoTotal   float64                     `json:"vencido_total"`
	PorVencerTotal float64                     `json:"por_vencer_total"`
	Rangos         []EmpresaCarteraCXPEdadFila `json:"rangos"`
}

type EmpresaCarteraCXPAbonoResultado struct {
	Cartera           EmpresaCarteraCXP `json:"cartera"`
	MontoAplicado     float64           `json:"monto_aplicado"`
	SaldoAnterior     float64           `json:"saldo_anterior"`
	SaldoNuevo        float64           `json:"saldo_nuevo"`
	ValorPagadoNuevo  float64           `json:"valor_pagado_nuevo"`
	EstadoAnterior    string            `json:"estado_anterior"`
	EstadoNuevo       string            `json:"estado_nuevo"`
	FechaAplicacion   string            `json:"fecha_aplicacion"`
	ReferenciaPago    string            `json:"referencia_pago,omitempty"`
	EventoContable    string            `json:"evento_contable"`
	DocumentoContable string            `json:"documento_contable"`
}

type EmpresaLibroOficialResumen struct {
	Tipo         string  `json:"tipo"`
	Periodo      string  `json:"periodo"`
	Registros    int     `json:"registros"`
	TotalDebito  float64 `json:"total_debito"`
	TotalCredito float64 `json:"total_credito"`
	Diferencia   float64 `json:"diferencia"`
	Estado       string  `json:"estado"`
}

type EmpresaLibroOficialLinea struct {
	FechaComprobante string  `json:"fecha_comprobante"`
	Codigo           string  `json:"codigo"`
	TipoComprobante  string  `json:"tipo_comprobante"`
	Periodo          string  `json:"periodo"`
	CuentaCodigo     string  `json:"cuenta_codigo"`
	CuentaNombre     string  `json:"cuenta_nombre"`
	TerceroNombre    string  `json:"tercero_nombre,omitempty"`
	Concepto         string  `json:"concepto"`
	Detalle          string  `json:"detalle"`
	Debito           float64 `json:"debito"`
	Credito          float64 `json:"credito"`
	Saldo            float64 `json:"saldo"`
}

func EnsureEmpresaContabilidadColombiaAvanzadaSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_contabilidad_exogena_formatos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			formato TEXT NOT NULL,
			version TEXT DEFAULT 'configurable',
			anio_gravable INTEGER NOT NULL,
			concepto TEXT DEFAULT '',
			descripcion TEXT DEFAULT '',
			periodicidad TEXT DEFAULT 'anual',
			estado TEXT DEFAULT 'activo',
			ultima_generacion TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			UNIQUE(empresa_id, formato, anio_gravable, concepto)
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_contabilidad_exogena_registros (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			formato_id INTEGER NOT NULL,
			tercero_id INTEGER DEFAULT 0,
			tipo_documento TEXT DEFAULT 'NIT',
			documento TEXT NOT NULL,
			digito_verificacion TEXT,
			razon_social TEXT NOT NULL,
			concepto TEXT DEFAULT '',
			cuenta_codigo TEXT DEFAULT '',
			base_valor REAL DEFAULT 0,
			iva REAL DEFAULT 0,
			retencion REAL DEFAULT 0,
			total REAL DEFAULT 0,
			validaciones TEXT,
			estado TEXT DEFAULT 'pendiente',
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_contabilidad_nomina_electronica (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			empleado_id INTEGER DEFAULT 0,
			tipo_documento TEXT DEFAULT 'CC',
			documento TEXT NOT NULL,
			nombre TEXT NOT NULL,
			periodo TEXT NOT NULL,
			fecha_pago TEXT NOT NULL,
			salario_base REAL DEFAULT 0,
			devengados REAL DEFAULT 0,
			deducciones REAL DEFAULT 0,
			total REAL DEFAULT 0,
			cune TEXT,
			estado_dian TEXT DEFAULT 'borrador',
			respuesta_dian TEXT,
			json_payload TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			UNIQUE(empresa_id, documento, periodo)
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_contabilidad_documentos_soporte (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			proveedor_id INTEGER DEFAULT 0,
			tipo_documento TEXT DEFAULT 'NIT',
			documento TEXT NOT NULL,
			nombre_proveedor TEXT NOT NULL,
			fecha_documento TEXT NOT NULL,
			periodo TEXT NOT NULL,
			concepto TEXT NOT NULL,
			subtotal REAL DEFAULT 0,
			iva REAL DEFAULT 0,
			retenciones REAL DEFAULT 0,
			total REAL DEFAULT 0,
			cuds TEXT,
			estado_dian TEXT DEFAULT 'borrador',
			respuesta_dian TEXT,
			json_payload TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_contabilidad_activos_fijos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			categoria TEXT DEFAULT 'equipo',
			fecha_compra TEXT NOT NULL,
			costo REAL DEFAULT 0,
			valor_residual REAL DEFAULT 0,
			vida_util_meses INTEGER DEFAULT 60,
			depreciacion_mensual REAL DEFAULT 0,
			depreciacion_acumulada REAL DEFAULT 0,
			valor_libros REAL DEFAULT 0,
			cuenta_activo TEXT DEFAULT '',
			cuenta_depreciacion TEXT DEFAULT '',
			cuenta_gasto TEXT DEFAULT '',
			ubicacion TEXT,
			responsable TEXT,
			estado TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			UNIQUE(empresa_id, codigo)
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_contabilidad_cartera_cxp (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			tipo TEXT NOT NULL,
			tercero_id INTEGER DEFAULT 0,
			tercero_nombre TEXT NOT NULL,
			documento TEXT NOT NULL,
			fecha_emision TEXT NOT NULL,
			fecha_vencimiento TEXT NOT NULL,
			cuenta_codigo TEXT DEFAULT '',
			concepto TEXT NOT NULL,
			valor_original REAL DEFAULT 0,
			valor_pagado REAL DEFAULT 0,
			saldo REAL DEFAULT 0,
			estado TEXT DEFAULT 'pendiente',
			origen_modulo TEXT DEFAULT 'manual',
			referencia_externa TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}
	return ensureEmpresaActivosFijosAvanzadoSchema(dbConn)
}

func ensureEmpresaActivosFijosAvanzadoSchema(dbConn *sql.DB) error {
	columns := []struct {
		name string
		def  string
	}{
		{"serial", "TEXT"},
		{"placa", "TEXT"},
		{"metodo_depreciacion", "TEXT DEFAULT 'linea_recta'"},
		{"fecha_inicio_depreciacion", "TEXT"},
		{"base_fiscal", "REAL DEFAULT 0"},
		{"vida_util_fiscal_meses", "INTEGER DEFAULT 0"},
		{"metodo_depreciacion_fiscal", "TEXT DEFAULT 'linea_recta'"},
		{"depreciacion_mensual_fiscal", "REAL DEFAULT 0"},
		{"depreciacion_acumulada_fiscal", "REAL DEFAULT 0"},
		{"valor_fiscal", "REAL DEFAULT 0"},
		{"diferencia_niif_fiscal", "REAL DEFAULT 0"},
		{"deterioro_acumulado", "REAL DEFAULT 0"},
		{"valor_razonable", "REAL DEFAULT 0"},
		{"fecha_ultima_valoracion", "TEXT"},
		{"cuenta_deterioro", "TEXT"},
		{"centro_costo", "TEXT"},
		{"proveedor", "TEXT"},
		{"valor_asegurado", "REAL DEFAULT 0"},
		{"poliza", "TEXT"},
		{"mantenimiento_cada_dias", "INTEGER DEFAULT 0"},
		{"ultimo_mantenimiento", "TEXT"},
		{"proximo_mantenimiento", "TEXT"},
		{"estado_operativo", "TEXT DEFAULT 'operativo'"},
		{"fecha_baja", "TEXT"},
		{"motivo_baja", "TEXT"},
		{"valor_baja", "REAL DEFAULT 0"},
	}
	for _, col := range columns {
		if err := addContabilidadColumnIfMissing(dbConn, "empresa_contabilidad_activos_fijos", col.name, col.def); err != nil {
			return err
		}
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_contabilidad_activos_depreciacion (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			activo_id INTEGER NOT NULL,
			periodo TEXT NOT NULL,
			fecha_calculo TEXT NOT NULL,
			metodo TEXT DEFAULT 'linea_recta',
			base_depreciable REAL DEFAULT 0,
			depreciacion_periodo REAL DEFAULT 0,
			depreciacion_acumulada REAL DEFAULT 0,
			valor_libros REAL DEFAULT 0,
			asiento_contable_id INTEGER DEFAULT 0,
			estado TEXT DEFAULT 'generado',
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			UNIQUE(empresa_id, activo_id, periodo)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_activos_depreciacion_empresa_periodo ON empresa_contabilidad_activos_depreciacion(empresa_id, periodo)`,
		`CREATE TABLE IF NOT EXISTS empresa_contabilidad_activos_eventos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			activo_id INTEGER NOT NULL,
			tipo TEXT NOT NULL,
			fecha_evento TEXT NOT NULL,
			ubicacion_origen TEXT,
			ubicacion_destino TEXT,
			responsable_origen TEXT,
			responsable_destino TEXT,
			valor REAL DEFAULT 0,
			estado TEXT DEFAULT 'cerrado',
			detalle TEXT,
			usuario_creador TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS ix_activos_eventos_empresa_activo ON empresa_contabilidad_activos_eventos(empresa_id, activo_id, fecha_evento)`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func addContabilidadColumnIfMissing(dbConn *sql.DB, tableName, columnName, columnDef string) error {
	stmt := fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s %s", tableName, columnName, columnDef)
	if _, err := ExecCompat(dbConn, stmt); err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "already exists") || strings.Contains(msg, "duplicate column") || strings.Contains(msg, "ya existe") {
			return nil
		}
		return err
	}
	return nil
}

func SeedEmpresaContabilidadAvanzadaBase(dbConn *sql.DB, empresaID int64, usuario string, anio int) error {
	if empresaID <= 0 {
		return errors.New("empresa_id requerido")
	}
	if anio <= 0 {
		anio = time.Now().Year()
	}
	formatos := defaultExogenaFormatos(empresaID, usuario, anio)
	for _, f := range formatos {
		_, err := dbConn.Exec(`INSERT OR IGNORE INTO empresa_contabilidad_exogena_formatos
			(empresa_id, formato, version, anio_gravable, concepto, descripcion, periodicidad, estado, usuario_creador)
			VALUES (?,?,?,?,?,?,?,?,?)`,
			f.EmpresaID, f.Formato, f.Version, f.AnioGravable, f.Concepto, f.Descripcion, f.Periodicidad, f.Estado, f.UsuarioCreador)
		if err != nil {
			return err
		}
	}
	return nil
}

func BuildEmpresaContabilidadAvanzadaDashboard(dbConn *sql.DB, empresaID int64) (EmpresaContabilidadAvanzadaDashboard, error) {
	d := EmpresaContabilidadAvanzadaDashboard{EmpresaID: empresaID}
	counts := []struct {
		table string
		dest  *int
		where string
	}{
		{"empresa_contabilidad_exogena_formatos", &d.FormatosExogena, ""},
		{"empresa_contabilidad_exogena_registros", &d.RegistrosExogena, ""},
		{"empresa_contabilidad_nomina_electronica", &d.NominasElectronicas, ""},
		{"empresa_contabilidad_documentos_soporte", &d.DocumentosSoporte, ""},
		{"empresa_contabilidad_activos_fijos", &d.ActivosFijos, " AND estado='activo'"},
		{"empresa_contabilidad_cartera_cxp", &d.CarteraCXPendientes, " AND estado IN ('pendiente','vencido','parcial')"},
	}
	for _, c := range counts {
		_ = dbConn.QueryRow("SELECT COUNT(1) FROM "+c.table+" WHERE empresa_id=?"+c.where, empresaID).Scan(c.dest)
	}
	d.LibrosDisponibles = buildDefaultLibroResumen(dbConn, empresaID, "")
	d.UltimosDocumentosDIAN = listUltimosDocumentosDIAN(dbConn, empresaID)
	return d, nil
}

func CreateEmpresaExogenaFormato(dbConn *sql.DB, x EmpresaExogenaFormato) (int64, error) {
	if x.EmpresaID <= 0 || strings.TrimSpace(x.Formato) == "" || x.AnioGravable <= 0 {
		return 0, errors.New("formato, año gravable y empresa son requeridos")
	}
	x.Formato = strings.ToUpper(strings.TrimSpace(x.Formato))
	x.Version = firstContabilidadValue(x.Version, "configurable")
	x.Periodicidad = firstContabilidadValue(x.Periodicidad, "anual")
	x.Estado = firstContabilidadValue(x.Estado, "activo")
	return insertSQLCompat(dbConn, `INSERT INTO empresa_contabilidad_exogena_formatos
		(empresa_id, formato, version, anio_gravable, concepto, descripcion, periodicidad, estado, usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		x.EmpresaID, x.Formato, x.Version, x.AnioGravable, x.Concepto, x.Descripcion, x.Periodicidad, x.Estado, x.UsuarioCreador)
}

func ListEmpresaExogenaFormatos(dbConn *sql.DB, empresaID int64, anio int) ([]EmpresaExogenaFormato, error) {
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if anio > 0 {
		where += " AND anio_gravable=?"
		args = append(args, anio)
	}
	rows, err := dbConn.Query(`SELECT id, empresa_id, formato, version, anio_gravable, concepto, descripcion, periodicidad, estado,
		COALESCE(ultima_generacion,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'')
		FROM empresa_contabilidad_exogena_formatos WHERE `+where+` ORDER BY anio_gravable DESC, formato`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaExogenaFormato
	for rows.Next() {
		var x EmpresaExogenaFormato
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Formato, &x.Version, &x.AnioGravable, &x.Concepto, &x.Descripcion, &x.Periodicidad, &x.Estado, &x.UltimaGeneracion, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func CreateEmpresaExogenaRegistro(dbConn *sql.DB, x EmpresaExogenaRegistro) (int64, error) {
	if x.EmpresaID <= 0 || x.FormatoID <= 0 || strings.TrimSpace(x.Documento) == "" || strings.TrimSpace(x.RazonSocial) == "" {
		return 0, errors.New("formato, documento y razon social son requeridos")
	}
	x.TipoDocumento = firstContabilidadValue(x.TipoDocumento, "NIT")
	x.Estado = firstContabilidadValue(x.Estado, "pendiente")
	if x.Total == 0 {
		x.Total = x.BaseValor + x.IVA - x.Retencion
	}
	x.Validaciones = validateExogenaRegistro(x)
	return insertSQLCompat(dbConn, `INSERT INTO empresa_contabilidad_exogena_registros
		(empresa_id, formato_id, tercero_id, tipo_documento, documento, digito_verificacion, razon_social, concepto, cuenta_codigo, base_valor, iva, retencion, total, validaciones, estado, usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		x.EmpresaID, x.FormatoID, x.TerceroID, x.TipoDocumento, x.Documento, x.DigitoVerificacion, x.RazonSocial, x.Concepto, x.CuentaCodigo, x.BaseValor, x.IVA, x.Retencion, x.Total, x.Validaciones, x.Estado, x.UsuarioCreador)
}

func ListEmpresaExogenaRegistros(dbConn *sql.DB, empresaID, formatoID int64) ([]EmpresaExogenaRegistro, error) {
	args := []interface{}{empresaID}
	where := "r.empresa_id=?"
	if formatoID > 0 {
		where += " AND r.formato_id=?"
		args = append(args, formatoID)
	}
	rows, err := dbConn.Query(`SELECT r.id, r.empresa_id, r.formato_id, COALESCE(f.formato,''), r.tercero_id, r.tipo_documento, r.documento,
		COALESCE(r.digito_verificacion,''), r.razon_social, r.concepto, r.cuenta_codigo, r.base_valor, r.iva, r.retencion, r.total,
		COALESCE(r.validaciones,''), r.estado, COALESCE(r.fecha_creacion,''), COALESCE(r.usuario_creador,'')
		FROM empresa_contabilidad_exogena_registros r
		LEFT JOIN empresa_contabilidad_exogena_formatos f ON f.id=r.formato_id AND f.empresa_id=r.empresa_id
		WHERE `+where+` ORDER BY r.id DESC LIMIT 1000`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaExogenaRegistro
	for rows.Next() {
		var x EmpresaExogenaRegistro
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.FormatoID, &x.Formato, &x.TerceroID, &x.TipoDocumento, &x.Documento, &x.DigitoVerificacion, &x.RazonSocial, &x.Concepto, &x.CuentaCodigo, &x.BaseValor, &x.IVA, &x.Retencion, &x.Total, &x.Validaciones, &x.Estado, &x.FechaCreacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func GenerateEmpresaExogenaFromAccounting(dbConn *sql.DB, empresaID, formatoID int64, usuario string) (int, error) {
	if empresaID <= 0 || formatoID <= 0 {
		return 0, errors.New("empresa_id y formato_id requeridos")
	}
	rows, err := dbConn.Query(`SELECT COALESCE(l.tercero_id,0), COALESCE(t.tipo_documento,'NIT'), COALESCE(t.documento,'SIN-DOC'),
		COALESCE(t.digito_verificacion,''), COALESCE(t.nombre,'Tercero no identificado'), l.cuenta_codigo,
		SUM(CASE WHEN l.debito>0 THEN l.debito ELSE l.credito END), SUM(COALESCE(l.base_gravable,0))
		FROM empresa_contabilidad_colombia_lineas l
		LEFT JOIN empresa_contabilidad_colombia_terceros t ON t.id=l.tercero_id AND t.empresa_id=l.empresa_id
		WHERE l.empresa_id=?
		GROUP BY COALESCE(l.tercero_id,0), COALESCE(t.tipo_documento,'NIT'), COALESCE(t.documento,'SIN-DOC'), COALESCE(t.digito_verificacion,''), COALESCE(t.nombre,'Tercero no identificado'), l.cuenta_codigo
		ORDER BY l.cuenta_codigo`, empresaID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	created := 0
	for rows.Next() {
		var terceroID int64
		var tipoDoc, doc, dv, nombre, cuenta string
		var total, base float64
		if err := rows.Scan(&terceroID, &tipoDoc, &doc, &dv, &nombre, &cuenta, &total, &base); err != nil {
			return created, err
		}
		reg := EmpresaExogenaRegistro{
			EmpresaID: empresaID, FormatoID: formatoID, TerceroID: terceroID, TipoDocumento: tipoDoc, Documento: doc,
			DigitoVerificacion: dv, RazonSocial: nombre, Concepto: "Generado desde contabilidad", CuentaCodigo: cuenta,
			BaseValor: base, Total: total, Estado: "validado", UsuarioCreador: usuario,
		}
		if _, err := CreateEmpresaExogenaRegistro(dbConn, reg); err != nil {
			return created, err
		}
		created++
	}
	if err := rows.Err(); err != nil {
		return created, err
	}
	_, _ = dbConn.Exec(`UPDATE empresa_contabilidad_exogena_formatos SET ultima_generacion=CURRENT_TIMESTAMP, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, empresaID, formatoID)
	return created, nil
}

func CreateEmpresaNominaElectronica(dbConn *sql.DB, x EmpresaNominaElectronica) (int64, error) {
	if x.EmpresaID <= 0 || strings.TrimSpace(x.Documento) == "" || strings.TrimSpace(x.Nombre) == "" || strings.TrimSpace(x.Periodo) == "" {
		return 0, errors.New("empleado, documento y periodo son requeridos")
	}
	x.TipoDocumento = firstContabilidadValue(x.TipoDocumento, "CC")
	x.EstadoDIAN = firstContabilidadValue(x.EstadoDIAN, "borrador")
	if x.Total == 0 {
		x.Total = x.Devengados - x.Deducciones
	}
	var id int64
	err := queryRowSQLCompat(dbConn, `INSERT INTO empresa_contabilidad_nomina_electronica
		(empresa_id, empleado_id, tipo_documento, documento, nombre, periodo, fecha_pago, salario_base, devengados, deducciones, total, cune, estado_dian, respuesta_dian, json_payload, fecha_actualizacion, usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP,?)
		ON CONFLICT (empresa_id, documento, periodo) DO UPDATE SET
			empleado_id=EXCLUDED.empleado_id,
			tipo_documento=EXCLUDED.tipo_documento,
			nombre=EXCLUDED.nombre,
			fecha_pago=EXCLUDED.fecha_pago,
			salario_base=EXCLUDED.salario_base,
			devengados=EXCLUDED.devengados,
			deducciones=EXCLUDED.deducciones,
			total=EXCLUDED.total,
			cune=EXCLUDED.cune,
			estado_dian=EXCLUDED.estado_dian,
			respuesta_dian=EXCLUDED.respuesta_dian,
			json_payload=EXCLUDED.json_payload,
			fecha_actualizacion=CURRENT_TIMESTAMP,
			usuario_creador=EXCLUDED.usuario_creador
		RETURNING id`,
		x.EmpresaID, x.EmpleadoID, x.TipoDocumento, x.Documento, x.Nombre, x.Periodo, firstContabilidadValue(x.FechaPago, time.Now().Format("2006-01-02")), x.SalarioBase, x.Devengados, x.Deducciones, x.Total, x.CUNE, x.EstadoDIAN, x.RespuestaDIAN, x.JSONPayload, x.UsuarioCreador).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func ListEmpresaNominaElectronica(dbConn *sql.DB, empresaID int64, periodo string) ([]EmpresaNominaElectronica, error) {
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if strings.TrimSpace(periodo) != "" {
		where += " AND periodo=?"
		args = append(args, periodo)
	}
	rows, err := dbConn.Query(`SELECT id, empresa_id, empleado_id, tipo_documento, documento, nombre, periodo, fecha_pago, salario_base,
		devengados, deducciones, total, COALESCE(cune,''), estado_dian, COALESCE(respuesta_dian,''), COALESCE(json_payload,''),
		COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'')
		FROM empresa_contabilidad_nomina_electronica WHERE `+where+` ORDER BY periodo DESC, nombre`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaNominaElectronica
	for rows.Next() {
		var x EmpresaNominaElectronica
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.EmpleadoID, &x.TipoDocumento, &x.Documento, &x.Nombre, &x.Periodo, &x.FechaPago, &x.SalarioBase, &x.Devengados, &x.Deducciones, &x.Total, &x.CUNE, &x.EstadoDIAN, &x.RespuestaDIAN, &x.JSONPayload, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func CreateEmpresaDocumentoSoporte(dbConn *sql.DB, x EmpresaDocumentoSoporteElectronico) (int64, error) {
	if x.EmpresaID <= 0 || strings.TrimSpace(x.Documento) == "" || strings.TrimSpace(x.NombreProveedor) == "" || strings.TrimSpace(x.Concepto) == "" {
		return 0, errors.New("proveedor, documento y concepto son requeridos")
	}
	x.TipoDocumento = firstContabilidadValue(x.TipoDocumento, "NIT")
	x.EstadoDIAN = firstContabilidadValue(x.EstadoDIAN, "borrador")
	if x.Total == 0 {
		x.Total = x.Subtotal + x.IVA - x.Retenciones
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_contabilidad_documentos_soporte
		(empresa_id, proveedor_id, tipo_documento, documento, nombre_proveedor, fecha_documento, periodo, concepto, subtotal, iva, retenciones, total, cuds, estado_dian, respuesta_dian, json_payload, usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		x.EmpresaID, x.ProveedorID, x.TipoDocumento, x.Documento, x.NombreProveedor, firstContabilidadValue(x.FechaDocumento, time.Now().Format("2006-01-02")), firstContabilidadValue(x.Periodo, time.Now().Format("2006-01")), x.Concepto, x.Subtotal, x.IVA, x.Retenciones, x.Total, x.CUDS, x.EstadoDIAN, x.RespuestaDIAN, x.JSONPayload, x.UsuarioCreador)
}

func ListEmpresaDocumentosSoporte(dbConn *sql.DB, empresaID int64, periodo string) ([]EmpresaDocumentoSoporteElectronico, error) {
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if strings.TrimSpace(periodo) != "" {
		where += " AND periodo=?"
		args = append(args, periodo)
	}
	rows, err := dbConn.Query(`SELECT id, empresa_id, proveedor_id, tipo_documento, documento, nombre_proveedor, fecha_documento, periodo,
		concepto, subtotal, iva, retenciones, total, COALESCE(cuds,''), estado_dian, COALESCE(respuesta_dian,''), COALESCE(json_payload,''),
		COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'')
		FROM empresa_contabilidad_documentos_soporte WHERE `+where+` ORDER BY fecha_documento DESC, id DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaDocumentoSoporteElectronico
	for rows.Next() {
		var x EmpresaDocumentoSoporteElectronico
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.ProveedorID, &x.TipoDocumento, &x.Documento, &x.NombreProveedor, &x.FechaDocumento, &x.Periodo, &x.Concepto, &x.Subtotal, &x.IVA, &x.Retenciones, &x.Total, &x.CUDS, &x.EstadoDIAN, &x.RespuestaDIAN, &x.JSONPayload, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func CreateEmpresaActivoFijo(dbConn *sql.DB, x EmpresaActivoFijo) (int64, error) {
	if x.EmpresaID <= 0 || strings.TrimSpace(x.Codigo) == "" || strings.TrimSpace(x.Nombre) == "" {
		return 0, errors.New("codigo y nombre del activo son requeridos")
	}
	if x.VidaUtilMeses <= 0 {
		x.VidaUtilMeses = 60
	}
	x.Categoria = firstContabilidadValue(x.Categoria, "equipo")
	x.FechaCompra = firstContabilidadValue(x.FechaCompra, time.Now().Format("2006-01-02"))
	x.FechaInicioDepreciacion = firstContabilidadValue(x.FechaInicioDepreciacion, x.FechaCompra)
	x.MetodoDepreciacion = normalizeActivoMetodoDepreciacion(x.MetodoDepreciacion)
	x.MetodoDepreciacionFiscal = normalizeActivoMetodoDepreciacion(x.MetodoDepreciacionFiscal)
	if x.VidaUtilFiscalMeses <= 0 {
		x.VidaUtilFiscalMeses = x.VidaUtilMeses
	}
	if x.BaseFiscal <= 0 {
		x.BaseFiscal = x.Costo
	}
	x.EstadoOperativo = firstContabilidadValue(x.EstadoOperativo, "operativo")
	x.Estado = firstContabilidadValue(x.Estado, "activo")
	x.DepreciacionMensual = calcularActivoDepreciacionMensual(x.Costo, x.ValorResidual, x.VidaUtilMeses)
	if x.DepreciacionAcumulada < 0 {
		x.DepreciacionAcumulada = 0
	}
	if x.DepreciacionAcumuladaFiscal < 0 {
		x.DepreciacionAcumuladaFiscal = 0
	}
	if x.DeterioroAcumulado < 0 {
		x.DeterioroAcumulado = 0
	}
	x.DepreciacionMensualFiscal = calcularActivoDepreciacionMensual(x.BaseFiscal, x.ValorResidual, x.VidaUtilFiscalMeses)
	x.ValorLibros = roundContabilidad(x.Costo - x.DepreciacionAcumulada - x.DeterioroAcumulado)
	if x.ValorLibros < 0 {
		x.ValorLibros = 0
	}
	x.ValorFiscal = roundContabilidad(x.BaseFiscal - x.DepreciacionAcumuladaFiscal)
	if x.ValorFiscal < 0 {
		x.ValorFiscal = 0
	}
	x.DiferenciaNIIFFiscal = roundContabilidad(x.ValorLibros - x.ValorFiscal)
	if x.MantenimientoCadaDias > 0 && strings.TrimSpace(x.ProximoMantenimiento) == "" {
		base := x.FechaCompra
		if strings.TrimSpace(x.UltimoMantenimiento) != "" {
			base = x.UltimoMantenimiento
		}
		x.ProximoMantenimiento = addDaysContabilidad(base, x.MantenimientoCadaDias)
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_contabilidad_activos_fijos
		(empresa_id, codigo, nombre, categoria, serial, placa, fecha_compra, costo, valor_residual, vida_util_meses, metodo_depreciacion, fecha_inicio_depreciacion, depreciacion_mensual, depreciacion_acumulada, valor_libros, base_fiscal, vida_util_fiscal_meses, metodo_depreciacion_fiscal, depreciacion_mensual_fiscal, depreciacion_acumulada_fiscal, valor_fiscal, diferencia_niif_fiscal, deterioro_acumulado, valor_razonable, fecha_ultima_valoracion, cuenta_deterioro, cuenta_activo, cuenta_depreciacion, cuenta_gasto, ubicacion, responsable, centro_costo, proveedor, valor_asegurado, poliza, mantenimiento_cada_dias, ultimo_mantenimiento, proximo_mantenimiento, estado_operativo, estado, usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		x.EmpresaID, strings.ToUpper(strings.TrimSpace(x.Codigo)), strings.TrimSpace(x.Nombre), x.Categoria, strings.TrimSpace(x.Serial), strings.TrimSpace(x.Placa), x.FechaCompra, x.Costo, x.ValorResidual, x.VidaUtilMeses, x.MetodoDepreciacion, x.FechaInicioDepreciacion, x.DepreciacionMensual, x.DepreciacionAcumulada, x.ValorLibros, x.BaseFiscal, x.VidaUtilFiscalMeses, x.MetodoDepreciacionFiscal, x.DepreciacionMensualFiscal, x.DepreciacionAcumuladaFiscal, x.ValorFiscal, x.DiferenciaNIIFFiscal, x.DeterioroAcumulado, x.ValorRazonable, x.FechaUltimaValoracion, x.CuentaDeterioro, x.CuentaActivo, x.CuentaDepreciacion, x.CuentaGasto, x.Ubicacion, x.Responsable, x.CentroCosto, x.Proveedor, x.ValorAsegurado, x.Poliza, x.MantenimientoCadaDias, x.UltimoMantenimiento, x.ProximoMantenimiento, x.EstadoOperativo, x.Estado, x.UsuarioCreador)
}

func ListEmpresaActivosFijos(dbConn *sql.DB, empresaID int64, estado string) ([]EmpresaActivoFijo, error) {
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if strings.TrimSpace(estado) != "" {
		where += " AND estado=?"
		args = append(args, estado)
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT id, empresa_id, codigo, nombre, categoria, COALESCE(serial,''), COALESCE(placa,''), fecha_compra, costo, valor_residual, vida_util_meses,
		COALESCE(metodo_depreciacion,'linea_recta'), COALESCE(fecha_inicio_depreciacion,''), depreciacion_mensual, depreciacion_acumulada, valor_libros,
		COALESCE(base_fiscal,0), COALESCE(vida_util_fiscal_meses,0), COALESCE(metodo_depreciacion_fiscal,'linea_recta'), COALESCE(depreciacion_mensual_fiscal,0), COALESCE(depreciacion_acumulada_fiscal,0), COALESCE(valor_fiscal,0), COALESCE(diferencia_niif_fiscal,0), COALESCE(deterioro_acumulado,0), COALESCE(valor_razonable,0), COALESCE(fecha_ultima_valoracion,''), COALESCE(cuenta_deterioro,''),
		cuenta_activo, cuenta_depreciacion, cuenta_gasto,
		COALESCE(ubicacion,''), COALESCE(responsable,''), COALESCE(centro_costo,''), COALESCE(proveedor,''), COALESCE(valor_asegurado,0), COALESCE(poliza,''), COALESCE(mantenimiento_cada_dias,0),
		COALESCE(ultimo_mantenimiento,''), COALESCE(proximo_mantenimiento,''), COALESCE(estado_operativo,'operativo'), COALESCE(fecha_baja,''), COALESCE(motivo_baja,''), COALESCE(valor_baja,0),
		estado, COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'')
		FROM empresa_contabilidad_activos_fijos WHERE `+where+` ORDER BY codigo`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaActivoFijo
	for rows.Next() {
		var x EmpresaActivoFijo
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Codigo, &x.Nombre, &x.Categoria, &x.Serial, &x.Placa, &x.FechaCompra, &x.Costo, &x.ValorResidual, &x.VidaUtilMeses, &x.MetodoDepreciacion, &x.FechaInicioDepreciacion, &x.DepreciacionMensual, &x.DepreciacionAcumulada, &x.ValorLibros, &x.BaseFiscal, &x.VidaUtilFiscalMeses, &x.MetodoDepreciacionFiscal, &x.DepreciacionMensualFiscal, &x.DepreciacionAcumuladaFiscal, &x.ValorFiscal, &x.DiferenciaNIIFFiscal, &x.DeterioroAcumulado, &x.ValorRazonable, &x.FechaUltimaValoracion, &x.CuentaDeterioro, &x.CuentaActivo, &x.CuentaDepreciacion, &x.CuentaGasto, &x.Ubicacion, &x.Responsable, &x.CentroCosto, &x.Proveedor, &x.ValorAsegurado, &x.Poliza, &x.MantenimientoCadaDias, &x.UltimoMantenimiento, &x.ProximoMantenimiento, &x.EstadoOperativo, &x.FechaBaja, &x.MotivoBaja, &x.ValorBaja, &x.Estado, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func GenerarEmpresaActivosDepreciacion(dbConn *sql.DB, empresaID int64, periodo, usuario string) ([]EmpresaActivoDepreciacion, error) {
	periodo = normalizeContabilidadPeriodo(periodo)
	if empresaID <= 0 || periodo == "" {
		return nil, errors.New("empresa_id y periodo son requeridos")
	}
	activos, err := ListEmpresaActivosFijos(dbConn, empresaID, "activo")
	if err != nil {
		return nil, err
	}
	fechaCalculo := periodo + "-28"
	for _, activo := range activos {
		row := calcularEmpresaActivoDepreciacionPeriodo(activo, periodo, fechaCalculo, usuario)
		if row.DepreciacionPeriodo <= 0 && row.DepreciacionAcumulada <= 0 {
			continue
		}
		if _, err := ExecCompat(dbConn, `INSERT INTO empresa_contabilidad_activos_depreciacion
			(empresa_id, activo_id, periodo, fecha_calculo, metodo, base_depreciable, depreciacion_periodo, depreciacion_acumulada, valor_libros, estado, usuario_creador)
			VALUES (?,?,?,?,?,?,?,?,?,?,?)
			ON CONFLICT (empresa_id, activo_id, periodo) DO UPDATE SET
				fecha_calculo=EXCLUDED.fecha_calculo,
				metodo=EXCLUDED.metodo,
				base_depreciable=EXCLUDED.base_depreciable,
				depreciacion_periodo=EXCLUDED.depreciacion_periodo,
				depreciacion_acumulada=EXCLUDED.depreciacion_acumulada,
				valor_libros=EXCLUDED.valor_libros,
				estado=EXCLUDED.estado,
				usuario_creador=EXCLUDED.usuario_creador`,
			row.EmpresaID, row.ActivoID, row.Periodo, row.FechaCalculo, row.Metodo, row.BaseDepreciable, row.DepreciacionPeriodo, row.DepreciacionAcumulada, row.ValorLibros, row.Estado, row.UsuarioCreador); err != nil {
			return nil, err
		}
		if _, err := ExecCompat(dbConn, `UPDATE empresa_contabilidad_activos_fijos
			SET depreciacion_acumulada=?, valor_libros=?, depreciacion_acumulada_fiscal=?, valor_fiscal=?, diferencia_niif_fiscal=?, fecha_actualizacion=CURRENT_TIMESTAMP
			WHERE empresa_id=? AND id=?`,
			row.DepreciacionAcumulada, row.ValorLibros, calcularActivoDepreciacionFiscalAcumulada(activo, periodo), calcularActivoValorFiscalPeriodo(activo, periodo), roundContabilidad(row.ValorLibros-calcularActivoValorFiscalPeriodo(activo, periodo)), empresaID, row.ActivoID); err != nil {
			return nil, err
		}
	}
	return ListEmpresaActivosDepreciacion(dbConn, empresaID, periodo, 1000)
}

func ListEmpresaActivosDepreciacion(dbConn *sql.DB, empresaID int64, periodo string, limit int) ([]EmpresaActivoDepreciacion, error) {
	if limit <= 0 || limit > 2000 {
		limit = 500
	}
	args := []interface{}{empresaID}
	where := "d.empresa_id=?"
	if strings.TrimSpace(periodo) != "" {
		where += " AND d.periodo=?"
		args = append(args, normalizeContabilidadPeriodo(periodo))
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT d.id,d.empresa_id,d.activo_id,COALESCE(a.codigo,''),COALESCE(a.nombre,''),COALESCE(d.periodo,''),COALESCE(d.fecha_calculo,''),COALESCE(d.metodo,'linea_recta'),COALESCE(d.base_depreciable,0),COALESCE(d.depreciacion_periodo,0),COALESCE(d.depreciacion_acumulada,0),COALESCE(d.valor_libros,0),COALESCE(d.asiento_contable_id,0),COALESCE(d.estado,'generado'),COALESCE(d.fecha_creacion,''),COALESCE(d.usuario_creador,'')
		FROM empresa_contabilidad_activos_depreciacion d
		LEFT JOIN empresa_contabilidad_activos_fijos a ON a.empresa_id=d.empresa_id AND a.id=d.activo_id
		WHERE %s ORDER BY d.periodo DESC, a.codigo LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaActivoDepreciacion{}
	for rows.Next() {
		var x EmpresaActivoDepreciacion
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.ActivoID, &x.ActivoCodigo, &x.ActivoNombre, &x.Periodo, &x.FechaCalculo, &x.Metodo, &x.BaseDepreciable, &x.DepreciacionPeriodo, &x.DepreciacionAcumulada, &x.ValorLibros, &x.AsientoContableID, &x.Estado, &x.FechaCreacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func RegistrarEmpresaActivoEvento(dbConn *sql.DB, item EmpresaActivoEvento) (int64, error) {
	item.Tipo = normalizeActivoEventoTipo(item.Tipo)
	item.FechaEvento = firstContabilidadValue(item.FechaEvento, time.Now().Format("2006-01-02"))
	item.Estado = firstContabilidadValue(item.Estado, "cerrado")
	if item.EmpresaID <= 0 || item.ActivoID <= 0 {
		return 0, errors.New("empresa_id y activo_id son requeridos")
	}
	activo, err := GetEmpresaActivoFijo(dbConn, item.EmpresaID, item.ActivoID)
	if err != nil {
		return 0, err
	}
	if item.UbicacionOrigen == "" {
		item.UbicacionOrigen = activo.Ubicacion
	}
	if item.ResponsableOrigen == "" {
		item.ResponsableOrigen = activo.Responsable
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_contabilidad_activos_eventos
		(empresa_id, activo_id, tipo, fecha_evento, ubicacion_origen, ubicacion_destino, responsable_origen, responsable_destino, valor, estado, detalle, usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
		item.EmpresaID, item.ActivoID, item.Tipo, item.FechaEvento, item.UbicacionOrigen, item.UbicacionDestino, item.ResponsableOrigen, item.ResponsableDestino, item.Valor, item.Estado, item.Detalle, item.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	if err := aplicarEmpresaActivoEvento(dbConn, activo, item); err != nil {
		return id, err
	}
	return id, nil
}

func GetEmpresaActivoFijo(dbConn *sql.DB, empresaID, activoID int64) (EmpresaActivoFijo, error) {
	rows, err := ListEmpresaActivosFijos(dbConn, empresaID, "")
	if err != nil {
		return EmpresaActivoFijo{}, err
	}
	for _, row := range rows {
		if row.ID == activoID {
			return row, nil
		}
	}
	return EmpresaActivoFijo{}, sql.ErrNoRows
}

func ListEmpresaActivosEventos(dbConn *sql.DB, empresaID, activoID int64, limit int) ([]EmpresaActivoEvento, error) {
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	args := []interface{}{empresaID}
	where := "e.empresa_id=?"
	if activoID > 0 {
		where += " AND e.activo_id=?"
		args = append(args, activoID)
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT e.id,e.empresa_id,e.activo_id,COALESCE(a.codigo,''),COALESCE(a.nombre,''),COALESCE(e.tipo,''),COALESCE(e.fecha_evento,''),COALESCE(e.ubicacion_origen,''),COALESCE(e.ubicacion_destino,''),COALESCE(e.responsable_origen,''),COALESCE(e.responsable_destino,''),COALESCE(e.valor,0),COALESCE(e.estado,'cerrado'),COALESCE(e.detalle,''),COALESCE(e.usuario_creador,''),COALESCE(e.fecha_creacion,'')
		FROM empresa_contabilidad_activos_eventos e
		LEFT JOIN empresa_contabilidad_activos_fijos a ON a.empresa_id=e.empresa_id AND a.id=e.activo_id
		WHERE %s ORDER BY e.fecha_evento DESC, e.id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaActivoEvento{}
	for rows.Next() {
		var x EmpresaActivoEvento
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.ActivoID, &x.ActivoCodigo, &x.ActivoNombre, &x.Tipo, &x.FechaEvento, &x.UbicacionOrigen, &x.UbicacionDestino, &x.ResponsableOrigen, &x.ResponsableDestino, &x.Valor, &x.Estado, &x.Detalle, &x.UsuarioCreador, &x.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func BuildEmpresaActivosFijosAvanzadoResumen(dbConn *sql.DB, empresaID int64, periodo string) (EmpresaActivosFijosAvanzadoResumen, error) {
	periodo = normalizeContabilidadPeriodo(periodo)
	if periodo == "" {
		periodo = time.Now().Format("2006-01")
	}
	res := EmpresaActivosFijosAvanzadoResumen{EmpresaID: empresaID}
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*),COALESCE(SUM(costo),0),COALESCE(SUM(valor_libros),0),COALESCE(SUM(valor_fiscal),0),COALESCE(SUM(diferencia_niif_fiscal),0),COALESCE(SUM(deterioro_acumulado),0),COALESCE(SUM(depreciacion_acumulada),0) FROM empresa_contabilidad_activos_fijos WHERE empresa_id=? AND estado='activo'`, empresaID).Scan(&res.ActivosActivos, &res.CostoTotal, &res.ValorLibrosTotal, &res.ValorFiscalTotal, &res.DiferenciaNIIFFiscalTotal, &res.DeterioroAcumuladoTotal, &res.DepreciacionAcumuladaTotal)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_contabilidad_activos_fijos WHERE empresa_id=? AND estado IN ('baja','vendido','retirado')`, empresaID).Scan(&res.ActivosBaja)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_contabilidad_activos_fijos WHERE empresa_id=? AND estado='activo' AND COALESCE(proximo_mantenimiento,'')<>'' AND proximo_mantenimiento<=?`, empresaID, time.Now().Format("2006-01-02")).Scan(&res.MantenimientosPendientes)
	_ = QueryRowCompat(dbConn, `SELECT COALESCE(SUM(depreciacion_periodo),0) FROM empresa_contabilidad_activos_depreciacion WHERE empresa_id=? AND periodo=?`, empresaID, periodo).Scan(&res.DepreciacionPeriodoTotal)
	deps, _ := ListEmpresaActivosDepreciacion(dbConn, empresaID, periodo, 80)
	events, _ := ListEmpresaActivosEventos(dbConn, empresaID, 0, 40)
	res.Depreciaciones = deps
	res.UltimosEventos = events
	return res, nil
}

func aplicarEmpresaActivoEvento(dbConn *sql.DB, activo EmpresaActivoFijo, item EmpresaActivoEvento) error {
	switch item.Tipo {
	case "traslado":
		ubicacion := firstContabilidadValue(item.UbicacionDestino, activo.Ubicacion)
		responsable := firstContabilidadValue(item.ResponsableDestino, activo.Responsable)
		_, err := ExecCompat(dbConn, `UPDATE empresa_contabilidad_activos_fijos SET ubicacion=?, responsable=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, ubicacion, responsable, item.EmpresaID, item.ActivoID)
		return err
	case "mantenimiento":
		proximo := activo.ProximoMantenimiento
		if activo.MantenimientoCadaDias > 0 {
			proximo = addDaysContabilidad(item.FechaEvento, activo.MantenimientoCadaDias)
		}
		_, err := ExecCompat(dbConn, `UPDATE empresa_contabilidad_activos_fijos SET ultimo_mantenimiento=?, proximo_mantenimiento=?, estado_operativo='operativo', fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, item.FechaEvento, proximo, item.EmpresaID, item.ActivoID)
		return err
	case "baja", "venta", "retiro":
		estado := "baja"
		if item.Tipo == "venta" {
			estado = "vendido"
		}
		if item.Tipo == "retiro" {
			estado = "retirado"
		}
		_, err := ExecCompat(dbConn, `UPDATE empresa_contabilidad_activos_fijos SET estado=?, fecha_baja=?, motivo_baja=?, valor_baja=?, estado_operativo='retirado', fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, estado, item.FechaEvento, strings.TrimSpace(item.Detalle), item.Valor, item.EmpresaID, item.ActivoID)
		return err
	case "ajuste":
		valorLibros := activo.ValorLibros
		if item.Valor > 0 {
			valorLibros = roundContabilidad(item.Valor)
		}
		_, err := ExecCompat(dbConn, `UPDATE empresa_contabilidad_activos_fijos SET valor_libros=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, valorLibros, item.EmpresaID, item.ActivoID)
		return err
	default:
		return nil
	}
}

func calcularEmpresaActivoDepreciacionPeriodo(activo EmpresaActivoFijo, periodo, fechaCalculo, usuario string) EmpresaActivoDepreciacion {
	metodo := normalizeActivoMetodoDepreciacion(activo.MetodoDepreciacion)
	base := roundContabilidad(activo.Costo - activo.ValorResidual)
	if base < 0 {
		base = 0
	}
	meses := mesesDepreciacionHastaPeriodo(firstContabilidadValue(activo.FechaInicioDepreciacion, activo.FechaCompra), periodo)
	if meses < 0 {
		meses = 0
	}
	if activo.VidaUtilMeses <= 0 {
		activo.VidaUtilMeses = 60
	}
	if meses > activo.VidaUtilMeses {
		meses = activo.VidaUtilMeses
	}
	mensual := calcularActivoDepreciacionMensual(activo.Costo, activo.ValorResidual, activo.VidaUtilMeses)
	depPeriodo := mensual
	acumulada := roundContabilidad(float64(meses) * mensual)
	if metodo == "saldos_decrecientes" {
		previoMeses := meses - 1
		if previoMeses < 0 {
			previoMeses = 0
		}
		tasa := 2 / float64(activo.VidaUtilMeses)
		valorPrevio := activo.Costo
		for i := 0; i < previoMeses; i++ {
			valorPrevio -= roundContabilidad(valorPrevio * tasa)
			if valorPrevio < activo.ValorResidual {
				valorPrevio = activo.ValorResidual
				break
			}
		}
		depPeriodo = roundContabilidad(valorPrevio * tasa)
		if valorPrevio-depPeriodo < activo.ValorResidual {
			depPeriodo = roundContabilidad(valorPrevio - activo.ValorResidual)
		}
		acumulada = roundContabilidad(activo.Costo - (valorPrevio - depPeriodo))
	}
	if acumulada > base {
		acumulada = base
	}
	if depPeriodo < 0 || meses == 0 {
		depPeriodo = 0
	}
	valorLibros := roundContabilidad(activo.Costo - acumulada)
	if valorLibros < activo.ValorResidual {
		valorLibros = activo.ValorResidual
	}
	return EmpresaActivoDepreciacion{
		EmpresaID:             activo.EmpresaID,
		ActivoID:              activo.ID,
		ActivoCodigo:          activo.Codigo,
		ActivoNombre:          activo.Nombre,
		Periodo:               periodo,
		FechaCalculo:          fechaCalculo,
		Metodo:                metodo,
		BaseDepreciable:       base,
		DepreciacionPeriodo:   roundContabilidad(depPeriodo),
		DepreciacionAcumulada: acumulada,
		ValorLibros:           valorLibros,
		Estado:                "generado",
		UsuarioCreador:        usuario,
	}
}

func calcularActivoDepreciacionMensual(costo, residual float64, vidaUtilMeses int) float64 {
	if vidaUtilMeses <= 0 {
		vidaUtilMeses = 60
	}
	base := costo - residual
	if base < 0 {
		base = 0
	}
	return roundContabilidad(base / float64(vidaUtilMeses))
}

func calcularActivoDepreciacionFiscalAcumulada(activo EmpresaActivoFijo, periodo string) float64 {
	baseFiscal := activo.BaseFiscal
	if baseFiscal <= 0 {
		baseFiscal = activo.Costo
	}
	vidaFiscal := activo.VidaUtilFiscalMeses
	if vidaFiscal <= 0 {
		vidaFiscal = activo.VidaUtilMeses
	}
	if vidaFiscal <= 0 {
		vidaFiscal = 60
	}
	meses := mesesDepreciacionHastaPeriodo(firstContabilidadValue(activo.FechaInicioDepreciacion, activo.FechaCompra), periodo)
	if meses < 0 {
		meses = 0
	}
	if meses > vidaFiscal {
		meses = vidaFiscal
	}
	return roundContabilidad(float64(meses) * calcularActivoDepreciacionMensual(baseFiscal, activo.ValorResidual, vidaFiscal))
}

func calcularActivoValorFiscalPeriodo(activo EmpresaActivoFijo, periodo string) float64 {
	baseFiscal := activo.BaseFiscal
	if baseFiscal <= 0 {
		baseFiscal = activo.Costo
	}
	valor := roundContabilidad(baseFiscal - calcularActivoDepreciacionFiscalAcumulada(activo, periodo))
	if valor < 0 {
		return 0
	}
	return valor
}

func mesesDepreciacionHastaPeriodo(fechaInicio, periodo string) int {
	if len(strings.TrimSpace(fechaInicio)) < 7 || len(strings.TrimSpace(periodo)) < 7 {
		return 0
	}
	start, err := time.Parse("2006-01", strings.TrimSpace(fechaInicio)[:7])
	if err != nil {
		return 0
	}
	end, err := time.Parse("2006-01", strings.TrimSpace(periodo)[:7])
	if err != nil {
		return 0
	}
	return (end.Year()-start.Year())*12 + int(end.Month()-start.Month()) + 1
}

func normalizeActivoMetodoDepreciacion(v string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case "saldos_decrecientes", "linea_recta":
		return s
	default:
		return "linea_recta"
	}
}

func normalizeActivoEventoTipo(v string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case "traslado", "mantenimiento", "baja", "venta", "retiro", "ajuste":
		return s
	default:
		return "mantenimiento"
	}
}

func normalizeContabilidadPeriodo(v string) string {
	s := strings.TrimSpace(v)
	if len(s) >= 7 {
		return s[:7]
	}
	return ""
}

func addDaysContabilidad(fecha string, days int) string {
	if days <= 0 {
		return strings.TrimSpace(fecha)
	}
	t, err := time.Parse("2006-01-02", strings.TrimSpace(fecha))
	if err != nil {
		return strings.TrimSpace(fecha)
	}
	return t.AddDate(0, 0, days).Format("2006-01-02")
}

func CreateEmpresaCarteraCXP(dbConn *sql.DB, x EmpresaCarteraCXP) (int64, error) {
	if x.EmpresaID <= 0 || strings.TrimSpace(x.Tipo) == "" || strings.TrimSpace(x.TerceroNombre) == "" || strings.TrimSpace(x.Documento) == "" {
		return 0, errors.New("tipo, tercero y documento son requeridos")
	}
	x.Tipo = strings.ToLower(strings.TrimSpace(x.Tipo))
	if x.Tipo != "cxc" && x.Tipo != "cxp" {
		return 0, errors.New("tipo debe ser cxc o cxp")
	}
	x.FechaEmision = firstContabilidadValue(x.FechaEmision, time.Now().Format("2006-01-02"))
	x.FechaVencimiento = firstContabilidadValue(x.FechaVencimiento, x.FechaEmision)
	x.Estado = firstContabilidadValue(x.Estado, "pendiente")
	x.OrigenModulo = firstContabilidadValue(x.OrigenModulo, "manual")
	if x.Saldo == 0 {
		x.Saldo = x.ValorOriginal - x.ValorPagado
	}
	x.Saldo = roundContabilidad(maxFloat64(x.Saldo, 0))
	x.ValorOriginal = roundContabilidad(maxFloat64(x.ValorOriginal, 0))
	x.ValorPagado = roundContabilidad(maxFloat64(x.ValorPagado, 0))
	x.Estado = normalizeEmpresaCarteraCXPEstado(x.Estado, x.Saldo, x.FechaVencimiento)
	return insertSQLCompat(dbConn, `INSERT INTO empresa_contabilidad_cartera_cxp
		(empresa_id, tipo, tercero_id, tercero_nombre, documento, fecha_emision, fecha_vencimiento, cuenta_codigo, concepto, valor_original, valor_pagado, saldo, estado, origen_modulo, referencia_externa, usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		x.EmpresaID, x.Tipo, x.TerceroID, x.TerceroNombre, x.Documento, x.FechaEmision, x.FechaVencimiento, x.CuentaCodigo, x.Concepto, x.ValorOriginal, x.ValorPagado, x.Saldo, x.Estado, x.OrigenModulo, x.ReferenciaExterna, x.UsuarioCreador)
}

func ListEmpresaCarteraCXP(dbConn *sql.DB, empresaID int64, tipo, estado string) ([]EmpresaCarteraCXP, error) {
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if strings.TrimSpace(tipo) != "" {
		where += " AND tipo=?"
		args = append(args, strings.ToLower(strings.TrimSpace(tipo)))
	}
	if strings.TrimSpace(estado) != "" {
		where += " AND estado=?"
		args = append(args, estado)
	}
	rows, err := dbConn.Query(`SELECT id, empresa_id, tipo, tercero_id, tercero_nombre, documento, fecha_emision, fecha_vencimiento,
		cuenta_codigo, concepto, valor_original, valor_pagado, saldo, estado, origen_modulo, COALESCE(referencia_externa,''),
		COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'')
		FROM empresa_contabilidad_cartera_cxp WHERE `+where+` ORDER BY fecha_vencimiento, id DESC LIMIT 1000`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaCarteraCXP
	for rows.Next() {
		var x EmpresaCarteraCXP
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Tipo, &x.TerceroID, &x.TerceroNombre, &x.Documento, &x.FechaEmision, &x.FechaVencimiento, &x.CuentaCodigo, &x.Concepto, &x.ValorOriginal, &x.ValorPagado, &x.Saldo, &x.Estado, &x.OrigenModulo, &x.ReferenciaExterna, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func GetEmpresaCarteraCXPByID(dbConn *sql.DB, empresaID, id int64) (EmpresaCarteraCXP, error) {
	var x EmpresaCarteraCXP
	if empresaID <= 0 || id <= 0 {
		return x, errors.New("empresa_id e id son obligatorios")
	}
	err := dbConn.QueryRow(`SELECT id, empresa_id, tipo, tercero_id, tercero_nombre, documento, fecha_emision, fecha_vencimiento,
		cuenta_codigo, concepto, valor_original, valor_pagado, saldo, estado, origen_modulo, COALESCE(referencia_externa,''),
		COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'')
		FROM empresa_contabilidad_cartera_cxp WHERE empresa_id=? AND id=? LIMIT 1`, empresaID, id).
		Scan(&x.ID, &x.EmpresaID, &x.Tipo, &x.TerceroID, &x.TerceroNombre, &x.Documento, &x.FechaEmision, &x.FechaVencimiento, &x.CuentaCodigo, &x.Concepto, &x.ValorOriginal, &x.ValorPagado, &x.Saldo, &x.Estado, &x.OrigenModulo, &x.ReferenciaExterna, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador)
	return x, err
}

func AplicarEmpresaCarteraCXPAbono(dbConn *sql.DB, empresaID, id int64, monto float64, fechaAplicacion, referenciaPago, usuario string) (EmpresaCarteraCXPAbonoResultado, error) {
	var result EmpresaCarteraCXPAbonoResultado
	if empresaID <= 0 || id <= 0 {
		return result, errors.New("empresa_id e id son obligatorios")
	}
	monto = roundContabilidad(monto)
	if monto <= 0 {
		return result, errors.New("monto del abono debe ser mayor que cero")
	}
	row, err := GetEmpresaCarteraCXPByID(dbConn, empresaID, id)
	if err != nil {
		return result, err
	}
	saldoAnterior := roundContabilidad(maxFloat64(row.Saldo, 0))
	if saldoAnterior <= 0 {
		return result, errors.New("la cartera seleccionada no tiene saldo pendiente")
	}
	if monto > saldoAnterior {
		monto = saldoAnterior
	}
	fechaAplicacion = firstContabilidadValue(fechaAplicacion, time.Now().Format("2006-01-02"))
	valorPagadoNuevo := roundContabilidad(maxFloat64(row.ValorPagado, 0) + monto)
	saldoNuevo := roundContabilidad(maxFloat64(row.ValorOriginal, 0) - valorPagadoNuevo)
	if saldoNuevo < 0 {
		saldoNuevo = 0
	}
	estadoNuevo := normalizeEmpresaCarteraCXPEstado("", saldoNuevo, row.FechaVencimiento)
	nowExpr := sqlNowExpr()
	_, err = ExecCompat(dbConn, `UPDATE empresa_contabilidad_cartera_cxp
		SET valor_pagado=?, saldo=?, estado=?, referencia_externa=?, fecha_actualizacion=`+nowExpr+`
		WHERE empresa_id=? AND id=?`,
		valorPagadoNuevo,
		saldoNuevo,
		estadoNuevo,
		firstContabilidadValue(referenciaPago, row.ReferenciaExterna),
		empresaID,
		id,
	)
	if err != nil {
		return result, err
	}
	updated, err := GetEmpresaCarteraCXPByID(dbConn, empresaID, id)
	if err != nil {
		return result, err
	}
	evento := "abono_cliente_registrado"
	if strings.EqualFold(row.Tipo, "cxp") {
		evento = "abono_proveedor_registrado"
	}
	result = EmpresaCarteraCXPAbonoResultado{
		Cartera:           updated,
		MontoAplicado:     monto,
		SaldoAnterior:     saldoAnterior,
		SaldoNuevo:        saldoNuevo,
		ValorPagadoNuevo:  valorPagadoNuevo,
		EstadoAnterior:    row.Estado,
		EstadoNuevo:       estadoNuevo,
		FechaAplicacion:   fechaAplicacion,
		ReferenciaPago:    strings.TrimSpace(referenciaPago),
		EventoContable:    evento,
		DocumentoContable: strings.TrimSpace(row.Documento),
	}
	return result, nil
}

func BuildEmpresaCarteraCXPEdades(dbConn *sql.DB, empresaID int64, tipo, fechaCorte string) (EmpresaCarteraCXPEdadesResumen, error) {
	resumen := EmpresaCarteraCXPEdadesResumen{
		EmpresaID:  empresaID,
		Tipo:       strings.ToLower(strings.TrimSpace(tipo)),
		FechaCorte: firstContabilidadValue(fechaCorte, time.Now().Format("2006-01-02")),
		Rangos:     make([]EmpresaCarteraCXPEdadFila, 0),
	}
	if empresaID <= 0 {
		return resumen, errors.New("empresa_id es obligatorio")
	}
	rows, err := ListEmpresaCarteraCXP(dbConn, empresaID, resumen.Tipo, "")
	if err != nil {
		return resumen, err
	}
	byKey := make(map[string]*EmpresaCarteraCXPEdadFila)
	order := []string{"por_vencer", "0_30", "31_60", "61_90", "91_180", "181_mas"}
	for _, key := range order {
		byKey[key] = &EmpresaCarteraCXPEdadFila{Tipo: resumen.Tipo, Rango: key}
	}
	for _, row := range rows {
		if !empresaCarteraCXPAbierta(row.Estado, row.Saldo) {
			continue
		}
		saldo := roundContabilidad(maxFloat64(row.Saldo, 0))
		if saldo <= 0 {
			continue
		}
		key := empresaCarteraCXPEdadRango(row.FechaVencimiento, resumen.FechaCorte)
		fila := byKey[key]
		if fila == nil {
			fila = &EmpresaCarteraCXPEdadFila{Tipo: resumen.Tipo, Rango: key}
			byKey[key] = fila
			order = append(order, key)
		}
		if fila.Tipo == "" {
			fila.Tipo = row.Tipo
		}
		fila.Registros++
		fila.Saldo = roundContabilidad(fila.Saldo + saldo)
		if key == "por_vencer" {
			fila.PorVencer = roundContabilidad(fila.PorVencer + saldo)
			resumen.PorVencerTotal = roundContabilidad(resumen.PorVencerTotal + saldo)
		} else {
			fila.Vencido = roundContabilidad(fila.Vencido + saldo)
			resumen.VencidoTotal = roundContabilidad(resumen.VencidoTotal + saldo)
		}
		if saldo > fila.SaldoMayor {
			fila.SaldoMayor = saldo
		}
		resumen.TotalRegistros++
		resumen.SaldoTotal = roundContabilidad(resumen.SaldoTotal + saldo)
	}
	for _, key := range order {
		if fila := byKey[key]; fila != nil && fila.Registros > 0 {
			resumen.Rangos = append(resumen.Rangos, *fila)
		}
	}
	return resumen, nil
}

func normalizeEmpresaCarteraCXPEstado(estado string, saldo float64, fechaVencimiento string) string {
	estado = strings.ToLower(strings.TrimSpace(estado))
	if saldo <= 0 {
		return "pagado"
	}
	if estado == "anulado" || estado == "castigado" {
		return estado
	}
	if estado == "parcial" || estado == "pendiente" || estado == "vencido" {
		if empresaCarteraDiasVencido(fechaVencimiento, time.Now().Format("2006-01-02")) > 0 {
			return "vencido"
		}
		return estado
	}
	return "pendiente"
}

func empresaCarteraCXPAbierta(estado string, saldo float64) bool {
	if saldo <= 0 {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(estado)) {
	case "pagado", "anulado", "castigado", "inactivo":
		return false
	default:
		return true
	}
}

func empresaCarteraCXPEdadRango(fechaVencimiento, fechaCorte string) string {
	dias := empresaCarteraDiasVencido(fechaVencimiento, fechaCorte)
	switch {
	case dias <= 0:
		return "por_vencer"
	case dias <= 30:
		return "0_30"
	case dias <= 60:
		return "31_60"
	case dias <= 90:
		return "61_90"
	case dias <= 180:
		return "91_180"
	default:
		return "181_mas"
	}
}

func empresaCarteraDiasVencido(fechaVencimiento, fechaCorte string) int {
	vencimiento, err := time.Parse("2006-01-02", strings.TrimSpace(fechaVencimiento))
	if err != nil {
		return 0
	}
	corte, err := time.Parse("2006-01-02", strings.TrimSpace(fechaCorte))
	if err != nil {
		corte = time.Now()
	}
	dias := int(corte.Sub(vencimiento).Hours() / 24)
	if dias < 0 {
		return 0
	}
	return dias
}

func ListEmpresaLibroOficial(dbConn *sql.DB, empresaID int64, tipo, periodo string) ([]EmpresaLibroOficialLinea, error) {
	args := []interface{}{empresaID}
	where := "c.empresa_id=? AND c.estado='contabilizado'"
	if strings.TrimSpace(periodo) != "" {
		where += " AND c.periodo_contable=?"
		args = append(args, periodo)
	}
	orderBy := "c.fecha_comprobante, c.codigo, l.id"
	if strings.EqualFold(tipo, "mayor") || strings.EqualFold(tipo, "auxiliar") {
		orderBy = "l.cuenta_codigo, c.fecha_comprobante, c.codigo, l.id"
	}
	rows, err := dbConn.Query(`SELECT c.fecha_comprobante, c.codigo, c.tipo_comprobante, c.periodo_contable, l.cuenta_codigo,
		COALESCE(cta.nombre,''), COALESCE(t.nombre,''), c.concepto, COALESCE(l.detalle,''), l.debito, l.credito
		FROM empresa_contabilidad_colombia_lineas l
		INNER JOIN empresa_contabilidad_colombia_comprobantes c ON c.id=l.comprobante_id AND c.empresa_id=l.empresa_id
		LEFT JOIN empresa_contabilidad_colombia_cuentas cta ON cta.empresa_id=l.empresa_id AND cta.codigo=l.cuenta_codigo
		LEFT JOIN empresa_contabilidad_colombia_terceros t ON t.empresa_id=l.empresa_id AND t.id=l.tercero_id
		WHERE `+where+` ORDER BY `+orderBy+` LIMIT 3000`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaLibroOficialLinea
	var saldo float64
	for rows.Next() {
		var x EmpresaLibroOficialLinea
		if err := rows.Scan(&x.FechaComprobante, &x.Codigo, &x.TipoComprobante, &x.Periodo, &x.CuentaCodigo, &x.CuentaNombre, &x.TerceroNombre, &x.Concepto, &x.Detalle, &x.Debito, &x.Credito); err != nil {
			return nil, err
		}
		saldo = roundContabilidad(saldo + x.Debito - x.Credito)
		x.Saldo = saldo
		out = append(out, x)
	}
	return out, rows.Err()
}

func buildDefaultLibroResumen(dbConn *sql.DB, empresaID int64, periodo string) []EmpresaLibroOficialResumen {
	rows := buildLibroResumenRows(dbConn, empresaID, periodo)
	if len(rows) == 0 {
		return []EmpresaLibroOficialResumen{
			{Tipo: "diario", Periodo: firstContabilidadValue(periodo, "todos"), Estado: "sin_movimientos"},
			{Tipo: "mayor", Periodo: firstContabilidadValue(periodo, "todos"), Estado: "sin_movimientos"},
			{Tipo: "balance_prueba", Periodo: firstContabilidadValue(periodo, "todos"), Estado: "sin_movimientos"},
			{Tipo: "estado_resultados", Periodo: firstContabilidadValue(periodo, "todos"), Estado: "sin_movimientos"},
			{Tipo: "estado_situacion_financiera", Periodo: firstContabilidadValue(periodo, "todos"), Estado: "sin_movimientos"},
		}
	}
	return rows
}

func buildLibroResumenRows(dbConn *sql.DB, empresaID int64, periodo string) []EmpresaLibroOficialResumen {
	args := []interface{}{empresaID}
	where := "empresa_id=? AND estado='contabilizado'"
	if strings.TrimSpace(periodo) != "" {
		where += " AND periodo_contable=?"
		args = append(args, periodo)
	}
	row := dbConn.QueryRow(`SELECT COUNT(1), COALESCE(SUM(total_debito),0), COALESCE(SUM(total_credito),0) FROM empresa_contabilidad_colombia_comprobantes WHERE `+where, args...)
	var count int
	var deb, cred float64
	if err := row.Scan(&count, &deb, &cred); err != nil {
		return nil
	}
	state := "cuadrado"
	if roundContabilidad(deb-cred) != 0 {
		state = "diferencia"
	}
	return []EmpresaLibroOficialResumen{
		{Tipo: "diario", Periodo: firstContabilidadValue(periodo, "todos"), Registros: count, TotalDebito: deb, TotalCredito: cred, Diferencia: roundContabilidad(deb - cred), Estado: state},
		{Tipo: "mayor", Periodo: firstContabilidadValue(periodo, "todos"), Registros: count, TotalDebito: deb, TotalCredito: cred, Diferencia: roundContabilidad(deb - cred), Estado: state},
		{Tipo: "balance_prueba", Periodo: firstContabilidadValue(periodo, "todos"), Registros: count, TotalDebito: deb, TotalCredito: cred, Diferencia: roundContabilidad(deb - cred), Estado: state},
		{Tipo: "estado_resultados", Periodo: firstContabilidadValue(periodo, "todos"), Registros: count, TotalDebito: deb, TotalCredito: cred, Diferencia: roundContabilidad(deb - cred), Estado: state},
		{Tipo: "estado_situacion_financiera", Periodo: firstContabilidadValue(periodo, "todos"), Registros: count, TotalDebito: deb, TotalCredito: cred, Diferencia: roundContabilidad(deb - cred), Estado: state},
	}
}

func listUltimosDocumentosDIAN(dbConn *sql.DB, empresaID int64) []EmpresaDocumentoDIANResumen {
	out := make([]EmpresaDocumentoDIANResumen, 0, 10)
	if rows, err := dbConn.Query(`SELECT 'nomina_electronica', documento, nombre, periodo, total, estado_dian, fecha_pago, COALESCE(respuesta_dian,'')
		FROM empresa_contabilidad_nomina_electronica WHERE empresa_id=? ORDER BY id DESC LIMIT 5`, empresaID); err == nil {
		defer rows.Close()
		for rows.Next() {
			var x EmpresaDocumentoDIANResumen
			_ = rows.Scan(&x.Modulo, &x.Codigo, &x.Tercero, &x.Periodo, &x.Total, &x.Estado, &x.Fecha, &x.Respuesta)
			out = append(out, x)
		}
	}
	if rows, err := dbConn.Query(`SELECT 'documento_soporte', documento, nombre_proveedor, periodo, total, estado_dian, fecha_documento, COALESCE(respuesta_dian,'')
		FROM empresa_contabilidad_documentos_soporte WHERE empresa_id=? ORDER BY id DESC LIMIT 5`, empresaID); err == nil {
		defer rows.Close()
		for rows.Next() {
			var x EmpresaDocumentoDIANResumen
			_ = rows.Scan(&x.Modulo, &x.Codigo, &x.Tercero, &x.Periodo, &x.Total, &x.Estado, &x.Fecha, &x.Respuesta)
			out = append(out, x)
		}
	}
	return out
}

func defaultExogenaFormatos(empresaID int64, usuario string, anio int) []EmpresaExogenaFormato {
	return []EmpresaExogenaFormato{
		{EmpresaID: empresaID, Formato: "1001", Version: "DIAN configurable", AnioGravable: anio, Concepto: "Pagos o abonos en cuenta", Descripcion: "Pagos, compras, costos, deducciones y retenciones por tercero.", Periodicidad: "anual", Estado: "activo", UsuarioCreador: usuario},
		{EmpresaID: empresaID, Formato: "1003", Version: "DIAN configurable", AnioGravable: anio, Concepto: "Retenciones practicadas", Descripcion: "Retenciones en la fuente practicadas y reportadas por tercero.", Periodicidad: "anual", Estado: "activo", UsuarioCreador: usuario},
		{EmpresaID: empresaID, Formato: "1005", Version: "DIAN configurable", AnioGravable: anio, Concepto: "IVA descontable", Descripcion: "Impuesto sobre las ventas descontable por tercero.", Periodicidad: "anual", Estado: "activo", UsuarioCreador: usuario},
		{EmpresaID: empresaID, Formato: "1006", Version: "DIAN configurable", AnioGravable: anio, Concepto: "IVA generado", Descripcion: "IVA generado en ventas y operaciones gravadas.", Periodicidad: "anual", Estado: "activo", UsuarioCreador: usuario},
		{EmpresaID: empresaID, Formato: "1007", Version: "DIAN configurable", AnioGravable: anio, Concepto: "Ingresos recibidos", Descripcion: "Ingresos facturados o recibidos por tercero.", Periodicidad: "anual", Estado: "activo", UsuarioCreador: usuario},
		{EmpresaID: empresaID, Formato: "1008", Version: "DIAN configurable", AnioGravable: anio, Concepto: "Cuentas por cobrar", Descripcion: "Saldos de cuentas por cobrar al cierre.", Periodicidad: "anual", Estado: "activo", UsuarioCreador: usuario},
		{EmpresaID: empresaID, Formato: "1009", Version: "DIAN configurable", AnioGravable: anio, Concepto: "Cuentas por pagar", Descripcion: "Saldos de cuentas por pagar al cierre.", Periodicidad: "anual", Estado: "activo", UsuarioCreador: usuario},
	}
}

func validateExogenaRegistro(x EmpresaExogenaRegistro) string {
	var issues []string
	if strings.TrimSpace(x.Documento) == "" {
		issues = append(issues, "documento requerido")
	}
	if strings.TrimSpace(x.RazonSocial) == "" {
		issues = append(issues, "razon social requerida")
	}
	if x.Total == 0 && x.BaseValor == 0 {
		issues = append(issues, "valor en cero")
	}
	if len(issues) == 0 {
		return "OK"
	}
	return strings.Join(issues, "; ")
}

func FormatEmpresaDocumentoElectronicoRef(prefix string, empresaID, id int64) string {
	return fmt.Sprintf("%s-%d-%06d", strings.ToUpper(strings.TrimSpace(prefix)), empresaID, id)
}
