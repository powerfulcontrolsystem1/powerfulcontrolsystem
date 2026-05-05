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
	ID                    int64   `json:"id"`
	EmpresaID             int64   `json:"empresa_id"`
	Codigo                string  `json:"codigo"`
	Nombre                string  `json:"nombre"`
	Categoria             string  `json:"categoria"`
	FechaCompra           string  `json:"fecha_compra"`
	Costo                 float64 `json:"costo"`
	ValorResidual         float64 `json:"valor_residual"`
	VidaUtilMeses         int     `json:"vida_util_meses"`
	DepreciacionMensual   float64 `json:"depreciacion_mensual"`
	DepreciacionAcumulada float64 `json:"depreciacion_acumulada"`
	ValorLibros           float64 `json:"valor_libros"`
	CuentaActivo          string  `json:"cuenta_activo"`
	CuentaDepreciacion    string  `json:"cuenta_depreciacion"`
	CuentaGasto           string  `json:"cuenta_gasto"`
	Ubicacion             string  `json:"ubicacion,omitempty"`
	Responsable           string  `json:"responsable,omitempty"`
	Estado                string  `json:"estado"`
	FechaCreacion         string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion    string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador        string  `json:"usuario_creador,omitempty"`
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
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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
	res, err := dbConn.Exec(`INSERT INTO empresa_contabilidad_exogena_formatos
		(empresa_id, formato, version, anio_gravable, concepto, descripcion, periodicidad, estado, usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		x.EmpresaID, x.Formato, x.Version, x.AnioGravable, x.Concepto, x.Descripcion, x.Periodicidad, x.Estado, x.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
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
	res, err := dbConn.Exec(`INSERT INTO empresa_contabilidad_exogena_registros
		(empresa_id, formato_id, tercero_id, tipo_documento, documento, digito_verificacion, razon_social, concepto, cuenta_codigo, base_valor, iva, retencion, total, validaciones, estado, usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		x.EmpresaID, x.FormatoID, x.TerceroID, x.TipoDocumento, x.Documento, x.DigitoVerificacion, x.RazonSocial, x.Concepto, x.CuentaCodigo, x.BaseValor, x.IVA, x.Retencion, x.Total, x.Validaciones, x.Estado, x.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
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
	res, err := dbConn.Exec(`INSERT OR REPLACE INTO empresa_contabilidad_nomina_electronica
		(empresa_id, empleado_id, tipo_documento, documento, nombre, periodo, fecha_pago, salario_base, devengados, deducciones, total, cune, estado_dian, respuesta_dian, json_payload, fecha_actualizacion, usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP,?)`,
		x.EmpresaID, x.EmpleadoID, x.TipoDocumento, x.Documento, x.Nombre, x.Periodo, firstContabilidadValue(x.FechaPago, time.Now().Format("2006-01-02")), x.SalarioBase, x.Devengados, x.Deducciones, x.Total, x.CUNE, x.EstadoDIAN, x.RespuestaDIAN, x.JSONPayload, x.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
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
	res, err := dbConn.Exec(`INSERT INTO empresa_contabilidad_documentos_soporte
		(empresa_id, proveedor_id, tipo_documento, documento, nombre_proveedor, fecha_documento, periodo, concepto, subtotal, iva, retenciones, total, cuds, estado_dian, respuesta_dian, json_payload, usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		x.EmpresaID, x.ProveedorID, x.TipoDocumento, x.Documento, x.NombreProveedor, firstContabilidadValue(x.FechaDocumento, time.Now().Format("2006-01-02")), firstContabilidadValue(x.Periodo, time.Now().Format("2006-01")), x.Concepto, x.Subtotal, x.IVA, x.Retenciones, x.Total, x.CUDS, x.EstadoDIAN, x.RespuestaDIAN, x.JSONPayload, x.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
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
	x.Estado = firstContabilidadValue(x.Estado, "activo")
	x.DepreciacionMensual = roundContabilidad((x.Costo - x.ValorResidual) / float64(x.VidaUtilMeses))
	if x.DepreciacionAcumulada < 0 {
		x.DepreciacionAcumulada = 0
	}
	x.ValorLibros = roundContabilidad(x.Costo - x.DepreciacionAcumulada)
	res, err := dbConn.Exec(`INSERT INTO empresa_contabilidad_activos_fijos
		(empresa_id, codigo, nombre, categoria, fecha_compra, costo, valor_residual, vida_util_meses, depreciacion_mensual, depreciacion_acumulada, valor_libros, cuenta_activo, cuenta_depreciacion, cuenta_gasto, ubicacion, responsable, estado, usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		x.EmpresaID, strings.ToUpper(strings.TrimSpace(x.Codigo)), x.Nombre, x.Categoria, x.FechaCompra, x.Costo, x.ValorResidual, x.VidaUtilMeses, x.DepreciacionMensual, x.DepreciacionAcumulada, x.ValorLibros, x.CuentaActivo, x.CuentaDepreciacion, x.CuentaGasto, x.Ubicacion, x.Responsable, x.Estado, x.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func ListEmpresaActivosFijos(dbConn *sql.DB, empresaID int64, estado string) ([]EmpresaActivoFijo, error) {
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if strings.TrimSpace(estado) != "" {
		where += " AND estado=?"
		args = append(args, estado)
	}
	rows, err := dbConn.Query(`SELECT id, empresa_id, codigo, nombre, categoria, fecha_compra, costo, valor_residual, vida_util_meses,
		depreciacion_mensual, depreciacion_acumulada, valor_libros, cuenta_activo, cuenta_depreciacion, cuenta_gasto,
		COALESCE(ubicacion,''), COALESCE(responsable,''), estado, COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,'')
		FROM empresa_contabilidad_activos_fijos WHERE `+where+` ORDER BY codigo`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaActivoFijo
	for rows.Next() {
		var x EmpresaActivoFijo
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Codigo, &x.Nombre, &x.Categoria, &x.FechaCompra, &x.Costo, &x.ValorResidual, &x.VidaUtilMeses, &x.DepreciacionMensual, &x.DepreciacionAcumulada, &x.ValorLibros, &x.CuentaActivo, &x.CuentaDepreciacion, &x.CuentaGasto, &x.Ubicacion, &x.Responsable, &x.Estado, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
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
	res, err := dbConn.Exec(`INSERT INTO empresa_contabilidad_cartera_cxp
		(empresa_id, tipo, tercero_id, tercero_nombre, documento, fecha_emision, fecha_vencimiento, cuenta_codigo, concepto, valor_original, valor_pagado, saldo, estado, origen_modulo, referencia_externa, usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		x.EmpresaID, x.Tipo, x.TerceroID, x.TerceroNombre, x.Documento, x.FechaEmision, x.FechaVencimiento, x.CuentaCodigo, x.Concepto, x.ValorOriginal, x.ValorPagado, x.Saldo, x.Estado, x.OrigenModulo, x.ReferenciaExterna, x.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
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
