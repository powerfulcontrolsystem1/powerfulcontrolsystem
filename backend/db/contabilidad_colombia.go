package db

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

type EmpresaContabilidadConfig struct {
	EmpresaID          int64  `json:"empresa_id"`
	NombreSistema      string `json:"nombre_sistema"`
	Moneda             string `json:"moneda"`
	PeriodoActual      string `json:"periodo_actual"`
	PUCVersion         string `json:"puc_version"`
	BaseNIIF           string `json:"base_niif"`
	BloquearCierre     bool   `json:"bloquear_cierre"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
}

type EmpresaContabilidadCuenta struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Codigo             string `json:"codigo"`
	Nombre             string `json:"nombre"`
	Naturaleza         string `json:"naturaleza"`
	TipoCuenta         string `json:"tipo_cuenta"`
	CuentaPadre        string `json:"cuenta_padre,omitempty"`
	AceptaMovimiento   bool   `json:"acepta_movimiento"`
	TerceroRequerido   bool   `json:"tercero_requerido"`
	ImpuestoRequerido  bool   `json:"impuesto_requerido"`
	Estado             string `json:"estado"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
}

type EmpresaContabilidadTercero struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	TipoDocumento      string `json:"tipo_documento"`
	Documento          string `json:"documento"`
	DigitoVerificacion string `json:"digito_verificacion,omitempty"`
	Nombre             string `json:"nombre"`
	TipoTercero        string `json:"tipo_tercero"`
	RegimenFiscal      string `json:"regimen_fiscal"`
	Responsabilidades  string `json:"responsabilidades,omitempty"`
	Email              string `json:"email,omitempty"`
	Telefono           string `json:"telefono,omitempty"`
	Direccion          string `json:"direccion,omitempty"`
	Municipio          string `json:"municipio,omitempty"`
	Estado             string `json:"estado"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
}

type EmpresaContabilidadImpuesto struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	Codigo             string  `json:"codigo"`
	Nombre             string  `json:"nombre"`
	Tipo               string  `json:"tipo"`
	Porcentaje         float64 `json:"porcentaje"`
	CuentaDebito       string  `json:"cuenta_debito"`
	CuentaCredito      string  `json:"cuenta_credito"`
	Estado             string  `json:"estado"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
}

type EmpresaContabilidadAsientoLinea struct {
	ID             int64   `json:"id"`
	EmpresaID      int64   `json:"empresa_id"`
	ComprobanteID  int64   `json:"comprobante_id"`
	CuentaCodigo   string  `json:"cuenta_codigo"`
	CuentaNombre   string  `json:"cuenta_nombre,omitempty"`
	TerceroID      int64   `json:"tercero_id,omitempty"`
	TerceroNombre  string  `json:"tercero_nombre,omitempty"`
	Detalle        string  `json:"detalle"`
	Debito         float64 `json:"debito"`
	Credito        float64 `json:"credito"`
	BaseGravable   float64 `json:"base_gravable"`
	ImpuestoCodigo string  `json:"impuesto_codigo,omitempty"`
	CentroCosto    string  `json:"centro_costo,omitempty"`
}

type EmpresaContabilidadComprobante struct {
	ID                 int64                             `json:"id"`
	EmpresaID          int64                             `json:"empresa_id"`
	Codigo             string                            `json:"codigo"`
	TipoComprobante    string                            `json:"tipo_comprobante"`
	FechaComprobante   string                            `json:"fecha_comprobante"`
	PeriodoContable    string                            `json:"periodo_contable"`
	TerceroID          int64                             `json:"tercero_id,omitempty"`
	TerceroNombre      string                            `json:"tercero_nombre,omitempty"`
	Concepto           string                            `json:"concepto"`
	OrigenModulo       string                            `json:"origen_modulo"`
	ReferenciaExterna  string                            `json:"referencia_externa,omitempty"`
	Estado             string                            `json:"estado"`
	TotalDebito        float64                           `json:"total_debito"`
	TotalCredito       float64                           `json:"total_credito"`
	Diferencia         float64                           `json:"diferencia"`
	Observaciones      string                            `json:"observaciones,omitempty"`
	FechaCreacion      string                            `json:"fecha_creacion,omitempty"`
	FechaActualizacion string                            `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string                            `json:"usuario_creador,omitempty"`
	Lineas             []EmpresaContabilidadAsientoLinea `json:"lineas,omitempty"`
}

type EmpresaContabilidadPeriodo struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	Periodo            string  `json:"periodo"`
	Estado             string  `json:"estado"`
	TotalDebito        float64 `json:"total_debito"`
	TotalCredito       float64 `json:"total_credito"`
	Diferencia         float64 `json:"diferencia"`
	CerradoPor         string  `json:"cerrado_por,omitempty"`
	FechaCierre        string  `json:"fecha_cierre,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
}

type EmpresaContabilidadDashboard struct {
	EmpresaID            int64                            `json:"empresa_id"`
	Config               EmpresaContabilidadConfig        `json:"config"`
	Cuentas              int                              `json:"cuentas"`
	Terceros             int                              `json:"terceros"`
	Impuestos            int                              `json:"impuestos"`
	ComprobantesMes      int                              `json:"comprobantes_mes"`
	ComprobantesBorrador int                              `json:"comprobantes_borrador"`
	TotalDebitoMes       float64                          `json:"total_debito_mes"`
	TotalCreditoMes      float64                          `json:"total_credito_mes"`
	DiferenciaMes        float64                          `json:"diferencia_mes"`
	UltimosComprobantes  []EmpresaContabilidadComprobante `json:"ultimos_comprobantes"`
	Periodos             []EmpresaContabilidadPeriodo     `json:"periodos"`
}

func EnsureEmpresaContabilidadColombiaSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_contabilidad_colombia_config (
			empresa_id INTEGER PRIMARY KEY,
			nombre_sistema TEXT DEFAULT 'Contabilidad Colombia',
			moneda TEXT DEFAULT 'COP',
			periodo_actual TEXT,
			puc_version TEXT DEFAULT 'PUC Colombia base',
			base_niif TEXT DEFAULT 'NIIF pymes',
			bloquear_cierre INTEGER DEFAULT 1,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_contabilidad_colombia_cuentas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			naturaleza TEXT DEFAULT 'debito',
			tipo_cuenta TEXT DEFAULT 'auxiliar',
			cuenta_padre TEXT,
			acepta_movimiento INTEGER DEFAULT 1,
			tercero_requerido INTEGER DEFAULT 0,
			impuesto_requerido INTEGER DEFAULT 0,
			estado TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			UNIQUE(empresa_id,codigo)
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_contabilidad_colombia_terceros (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			tipo_documento TEXT DEFAULT 'NIT',
			documento TEXT NOT NULL,
			digito_verificacion TEXT,
			nombre TEXT NOT NULL,
			tipo_tercero TEXT DEFAULT 'cliente_proveedor',
			regimen_fiscal TEXT DEFAULT 'responsable_iva',
			responsabilidades TEXT,
			email TEXT,
			telefono TEXT,
			direccion TEXT,
			municipio TEXT,
			estado TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			UNIQUE(empresa_id,documento)
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_contabilidad_colombia_impuestos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			tipo TEXT DEFAULT 'iva',
			porcentaje REAL DEFAULT 0,
			cuenta_debito TEXT,
			cuenta_credito TEXT,
			estado TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			UNIQUE(empresa_id,codigo)
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_contabilidad_colombia_comprobantes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			tipo_comprobante TEXT DEFAULT 'nota_contable',
			fecha_comprobante TEXT DEFAULT CURRENT_TIMESTAMP,
			periodo_contable TEXT,
			tercero_id INTEGER DEFAULT 0,
			concepto TEXT NOT NULL,
			origen_modulo TEXT DEFAULT 'manual',
			referencia_externa TEXT,
			estado TEXT DEFAULT 'contabilizado',
			total_debito REAL DEFAULT 0,
			total_credito REAL DEFAULT 0,
			diferencia REAL DEFAULT 0,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			UNIQUE(empresa_id,codigo)
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_contabilidad_colombia_lineas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			comprobante_id INTEGER NOT NULL,
			cuenta_codigo TEXT NOT NULL,
			tercero_id INTEGER DEFAULT 0,
			detalle TEXT,
			debito REAL DEFAULT 0,
			credito REAL DEFAULT 0,
			base_gravable REAL DEFAULT 0,
			impuesto_codigo TEXT,
			centro_costo TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS empresa_contabilidad_colombia_periodos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			periodo TEXT NOT NULL,
			estado TEXT DEFAULT 'abierto',
			total_debito REAL DEFAULT 0,
			total_credito REAL DEFAULT 0,
			diferencia REAL DEFAULT 0,
			cerrado_por TEXT,
			fecha_cierre TEXT,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			UNIQUE(empresa_id,periodo)
		)`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func SeedEmpresaContabilidadColombiaBase(dbConn *sql.DB, empresaID int64, usuario string) error {
	if err := EnsureEmpresaContabilidadColombiaSchema(dbConn); err != nil {
		return err
	}
	cfg := defaultContabilidadConfig(empresaID)
	cfg.UsuarioCreador = usuario
	_ = UpsertEmpresaContabilidadConfig(dbConn, cfg)
	var count int
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_contabilidad_colombia_cuentas WHERE empresa_id=?`, empresaID).Scan(&count)
	if count == 0 {
		for _, c := range defaultPUCCuentas(empresaID, usuario) {
			if _, err := CreateEmpresaContabilidadCuenta(dbConn, c); err != nil {
				return err
			}
		}
	}
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_contabilidad_colombia_impuestos WHERE empresa_id=?`, empresaID).Scan(&count)
	if count == 0 {
		for _, imp := range defaultContabilidadImpuestos(empresaID, usuario) {
			if _, err := CreateEmpresaContabilidadImpuesto(dbConn, imp); err != nil {
				return err
			}
		}
	}
	return nil
}

func defaultContabilidadConfig(empresaID int64) EmpresaContabilidadConfig {
	return EmpresaContabilidadConfig{EmpresaID: empresaID, NombreSistema: "Contabilidad Colombia", Moneda: "COP", PeriodoActual: time.Now().Format("2006-01"), PUCVersion: "PUC Colombia base", BaseNIIF: "NIIF pymes", BloquearCierre: true}
}

func GetEmpresaContabilidadConfig(dbConn *sql.DB, empresaID int64) (EmpresaContabilidadConfig, error) {
	if err := EnsureEmpresaContabilidadColombiaSchema(dbConn); err != nil {
		return EmpresaContabilidadConfig{}, err
	}
	cfg := defaultContabilidadConfig(empresaID)
	var block int
	err := QueryRowCompat(dbConn, `SELECT empresa_id,COALESCE(nombre_sistema,''),COALESCE(moneda,'COP'),COALESCE(periodo_actual,''),COALESCE(puc_version,''),COALESCE(base_niif,''),COALESCE(bloquear_cierre,1),COALESCE(fecha_actualizacion,''),COALESCE(usuario_creador,'') FROM empresa_contabilidad_colombia_config WHERE empresa_id=?`, empresaID).Scan(&cfg.EmpresaID, &cfg.NombreSistema, &cfg.Moneda, &cfg.PeriodoActual, &cfg.PUCVersion, &cfg.BaseNIIF, &block, &cfg.FechaActualizacion, &cfg.UsuarioCreador)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return cfg, nil
		}
		return EmpresaContabilidadConfig{}, err
	}
	cfg.BloquearCierre = block > 0
	return normalizeContabilidadConfig(cfg), nil
}

func UpsertEmpresaContabilidadConfig(dbConn *sql.DB, cfg EmpresaContabilidadConfig) error {
	cfg = normalizeContabilidadConfig(cfg)
	_, err := ExecCompat(dbConn, `INSERT INTO empresa_contabilidad_colombia_config (empresa_id,nombre_sistema,moneda,periodo_actual,puc_version,base_niif,bloquear_cierre,fecha_actualizacion,usuario_creador) VALUES (?,?,?,?,?,?,?,CURRENT_TIMESTAMP,?) ON CONFLICT (empresa_id) DO UPDATE SET nombre_sistema=EXCLUDED.nombre_sistema, moneda=EXCLUDED.moneda, periodo_actual=EXCLUDED.periodo_actual, puc_version=EXCLUDED.puc_version, base_niif=EXCLUDED.base_niif, bloquear_cierre=EXCLUDED.bloquear_cierre, fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=EXCLUDED.usuario_creador`, cfg.EmpresaID, cfg.NombreSistema, cfg.Moneda, cfg.PeriodoActual, cfg.PUCVersion, cfg.BaseNIIF, boolInt(cfg.BloquearCierre), cfg.UsuarioCreador)
	return err
}

func CreateEmpresaContabilidadCuenta(dbConn *sql.DB, x EmpresaContabilidadCuenta) (int64, error) {
	x.Codigo = cleanCode(x.Codigo)
	x.Nombre = strings.TrimSpace(x.Nombre)
	if x.EmpresaID <= 0 || x.Codigo == "" || x.Nombre == "" {
		return 0, errors.New("empresa_id, codigo y nombre son obligatorios")
	}
	x.Naturaleza = firstContabilidadValue(x.Naturaleza, "debito")
	x.TipoCuenta = firstContabilidadValue(x.TipoCuenta, "auxiliar")
	x.Estado = firstContabilidadValue(x.Estado, "activo")
	return insertSQLCompat(dbConn, `INSERT INTO empresa_contabilidad_colombia_cuentas (empresa_id,codigo,nombre,naturaleza,tipo_cuenta,cuenta_padre,acepta_movimiento,tercero_requerido,impuesto_requerido,estado,fecha_creacion,fecha_actualizacion,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,?)`, x.EmpresaID, x.Codigo, x.Nombre, x.Naturaleza, x.TipoCuenta, cleanCode(x.CuentaPadre), boolInt(x.AceptaMovimiento), boolInt(x.TerceroRequerido), boolInt(x.ImpuestoRequerido), x.Estado, x.UsuarioCreador)
}

func ListEmpresaContabilidadCuentas(dbConn *sql.DB, empresaID int64, q string) ([]EmpresaContabilidadCuenta, error) {
	if err := SeedEmpresaContabilidadColombiaBase(dbConn, empresaID, "sistema"); err != nil {
		return nil, err
	}
	where := "empresa_id=?"
	args := []interface{}{empresaID}
	if strings.TrimSpace(q) != "" {
		where += " AND (LOWER(codigo) LIKE ? OR LOWER(nombre) LIKE ?)"
		like := "%" + strings.ToLower(strings.TrimSpace(q)) + "%"
		args = append(args, like, like)
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,codigo,nombre,COALESCE(naturaleza,''),COALESCE(tipo_cuenta,''),COALESCE(cuenta_padre,''),COALESCE(acepta_movimiento,1),COALESCE(tercero_requerido,0),COALESCE(impuesto_requerido,0),COALESCE(estado,'activo'),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,''),COALESCE(usuario_creador,'') FROM empresa_contabilidad_colombia_cuentas WHERE `+where+` ORDER BY codigo LIMIT 500`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaContabilidadCuenta
	for rows.Next() {
		var x EmpresaContabilidadCuenta
		var mov, ter, imp int
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Codigo, &x.Nombre, &x.Naturaleza, &x.TipoCuenta, &x.CuentaPadre, &mov, &ter, &imp, &x.Estado, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		x.AceptaMovimiento, x.TerceroRequerido, x.ImpuestoRequerido = mov > 0, ter > 0, imp > 0
		out = append(out, x)
	}
	return out, rows.Err()
}

func CreateEmpresaContabilidadTercero(dbConn *sql.DB, x EmpresaContabilidadTercero) (int64, error) {
	x.Documento = strings.TrimSpace(x.Documento)
	x.Nombre = strings.TrimSpace(x.Nombre)
	if x.EmpresaID <= 0 || x.Documento == "" || x.Nombre == "" {
		return 0, errors.New("documento y nombre del tercero son obligatorios")
	}
	x.TipoDocumento = strings.ToUpper(firstContabilidadValue(x.TipoDocumento, "NIT"))
	x.TipoTercero = firstContabilidadValue(x.TipoTercero, "cliente_proveedor")
	x.RegimenFiscal = firstContabilidadValue(x.RegimenFiscal, "responsable_iva")
	x.Estado = firstContabilidadValue(x.Estado, "activo")
	return insertSQLCompat(dbConn, `INSERT INTO empresa_contabilidad_colombia_terceros (empresa_id,tipo_documento,documento,digito_verificacion,nombre,tipo_tercero,regimen_fiscal,responsabilidades,email,telefono,direccion,municipio,estado,fecha_creacion,fecha_actualizacion,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,?)`, x.EmpresaID, x.TipoDocumento, x.Documento, strings.TrimSpace(x.DigitoVerificacion), x.Nombre, x.TipoTercero, x.RegimenFiscal, strings.TrimSpace(x.Responsabilidades), strings.TrimSpace(x.Email), strings.TrimSpace(x.Telefono), strings.TrimSpace(x.Direccion), strings.TrimSpace(x.Municipio), x.Estado, x.UsuarioCreador)
}

func ListEmpresaContabilidadTerceros(dbConn *sql.DB, empresaID int64, q string) ([]EmpresaContabilidadTercero, error) {
	if err := EnsureEmpresaContabilidadColombiaSchema(dbConn); err != nil {
		return nil, err
	}
	where := "empresa_id=?"
	args := []interface{}{empresaID}
	if strings.TrimSpace(q) != "" {
		where += " AND (LOWER(documento) LIKE ? OR LOWER(nombre) LIKE ?)"
		like := "%" + strings.ToLower(strings.TrimSpace(q)) + "%"
		args = append(args, like, like)
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,COALESCE(tipo_documento,''),documento,COALESCE(digito_verificacion,''),nombre,COALESCE(tipo_tercero,''),COALESCE(regimen_fiscal,''),COALESCE(responsabilidades,''),COALESCE(email,''),COALESCE(telefono,''),COALESCE(direccion,''),COALESCE(municipio,''),COALESCE(estado,'activo'),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,''),COALESCE(usuario_creador,'') FROM empresa_contabilidad_colombia_terceros WHERE `+where+` ORDER BY nombre LIMIT 300`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaContabilidadTercero
	for rows.Next() {
		var x EmpresaContabilidadTercero
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.TipoDocumento, &x.Documento, &x.DigitoVerificacion, &x.Nombre, &x.TipoTercero, &x.RegimenFiscal, &x.Responsabilidades, &x.Email, &x.Telefono, &x.Direccion, &x.Municipio, &x.Estado, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func CreateEmpresaContabilidadImpuesto(dbConn *sql.DB, x EmpresaContabilidadImpuesto) (int64, error) {
	x.Codigo = strings.ToUpper(strings.TrimSpace(x.Codigo))
	x.Nombre = strings.TrimSpace(x.Nombre)
	if x.EmpresaID <= 0 || x.Codigo == "" || x.Nombre == "" {
		return 0, errors.New("codigo y nombre del impuesto son obligatorios")
	}
	x.Tipo = firstContabilidadValue(x.Tipo, "iva")
	x.Estado = firstContabilidadValue(x.Estado, "activo")
	return insertSQLCompat(dbConn, `INSERT INTO empresa_contabilidad_colombia_impuestos (empresa_id,codigo,nombre,tipo,porcentaje,cuenta_debito,cuenta_credito,estado,fecha_creacion,fecha_actualizacion,usuario_creador) VALUES (?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,?)`, x.EmpresaID, x.Codigo, x.Nombre, x.Tipo, x.Porcentaje, cleanCode(x.CuentaDebito), cleanCode(x.CuentaCredito), x.Estado, x.UsuarioCreador)
}

func ListEmpresaContabilidadImpuestos(dbConn *sql.DB, empresaID int64) ([]EmpresaContabilidadImpuesto, error) {
	if err := SeedEmpresaContabilidadColombiaBase(dbConn, empresaID, "sistema"); err != nil {
		return nil, err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,codigo,nombre,COALESCE(tipo,''),COALESCE(porcentaje,0),COALESCE(cuenta_debito,''),COALESCE(cuenta_credito,''),COALESCE(estado,'activo'),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,''),COALESCE(usuario_creador,'') FROM empresa_contabilidad_colombia_impuestos WHERE empresa_id=? ORDER BY tipo,codigo`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaContabilidadImpuesto
	for rows.Next() {
		var x EmpresaContabilidadImpuesto
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Codigo, &x.Nombre, &x.Tipo, &x.Porcentaje, &x.CuentaDebito, &x.CuentaCredito, &x.Estado, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func CreateEmpresaContabilidadComprobante(dbConn *sql.DB, x EmpresaContabilidadComprobante) (int64, error) {
	if err := SeedEmpresaContabilidadColombiaBase(dbConn, x.EmpresaID, x.UsuarioCreador); err != nil {
		return 0, err
	}
	if x.EmpresaID <= 0 || strings.TrimSpace(x.Concepto) == "" {
		return 0, errors.New("concepto es obligatorio")
	}
	if len(x.Lineas) < 2 {
		return 0, errors.New("un comprobante contable requiere minimo dos lineas")
	}
	x.TipoComprobante = firstContabilidadValue(x.TipoComprobante, "nota_contable")
	x.FechaComprobante = firstContabilidadValue(x.FechaComprobante, time.Now().Format("2006-01-02"))
	x.PeriodoContable = firstContabilidadValue(x.PeriodoContable, periodFromDate(x.FechaComprobante))
	x.Estado = firstContabilidadValue(x.Estado, "contabilizado")
	x.OrigenModulo = firstContabilidadValue(x.OrigenModulo, "manual")
	if err := assertContabilidadPeriodoAbierto(dbConn, x.EmpresaID, x.PeriodoContable); err != nil {
		return 0, err
	}
	var deb, cred float64
	for _, line := range x.Lineas {
		if cleanCode(line.CuentaCodigo) == "" {
			return 0, errors.New("todas las lineas requieren cuenta contable")
		}
		deb += line.Debito
		cred += line.Credito
	}
	deb, cred = roundContabilidad(deb), roundContabilidad(cred)
	diff := roundContabilidad(deb - cred)
	if math.Abs(diff) > 0.009 {
		return 0, fmt.Errorf("comprobante descuadrado: debito %.2f credito %.2f", deb, cred)
	}
	code, err := nextContabilidadComprobanteCode(dbConn, x.EmpresaID, x.TipoComprobante)
	if err != nil {
		return 0, err
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_contabilidad_colombia_comprobantes (empresa_id,codigo,tipo_comprobante,fecha_comprobante,periodo_contable,tercero_id,concepto,origen_modulo,referencia_externa,estado,total_debito,total_credito,diferencia,observaciones,fecha_creacion,fecha_actualizacion,usuario_creador) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,?)`, x.EmpresaID, code, x.TipoComprobante, x.FechaComprobante, x.PeriodoContable, x.TerceroID, strings.TrimSpace(x.Concepto), x.OrigenModulo, strings.TrimSpace(x.ReferenciaExterna), x.Estado, deb, cred, diff, strings.TrimSpace(x.Observaciones), x.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	for _, line := range x.Lineas {
		_, err := insertSQLCompat(dbConn, `INSERT INTO empresa_contabilidad_colombia_lineas (empresa_id,comprobante_id,cuenta_codigo,tercero_id,detalle,debito,credito,base_gravable,impuesto_codigo,centro_costo) VALUES (?,?,?,?,?,?,?,?,?,?)`, x.EmpresaID, id, cleanCode(line.CuentaCodigo), line.TerceroID, strings.TrimSpace(line.Detalle), roundContabilidad(line.Debito), roundContabilidad(line.Credito), roundContabilidad(line.BaseGravable), strings.ToUpper(strings.TrimSpace(line.ImpuestoCodigo)), strings.TrimSpace(line.CentroCosto))
		if err != nil {
			return 0, err
		}
	}
	return id, nil
}

func ListEmpresaContabilidadComprobantes(dbConn *sql.DB, empresaID int64, periodo, estado string, limit int) ([]EmpresaContabilidadComprobante, error) {
	if err := EnsureEmpresaContabilidadColombiaSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 120
	}
	where := "c.empresa_id=?"
	args := []interface{}{empresaID}
	if strings.TrimSpace(periodo) != "" {
		where += " AND c.periodo_contable=?"
		args = append(args, strings.TrimSpace(periodo))
	}
	if strings.TrimSpace(estado) != "" {
		where += " AND LOWER(COALESCE(c.estado,''))=?"
		args = append(args, strings.ToLower(strings.TrimSpace(estado)))
	}
	args = append(args, limit)
	rows, err := ExecQueryCompat(dbConn, `SELECT c.id,c.empresa_id,c.codigo,COALESCE(c.tipo_comprobante,''),COALESCE(c.fecha_comprobante,''),COALESCE(c.periodo_contable,''),COALESCE(c.tercero_id,0),COALESCE(t.nombre,''),COALESCE(c.concepto,''),COALESCE(c.origen_modulo,''),COALESCE(c.referencia_externa,''),COALESCE(c.estado,''),COALESCE(c.total_debito,0),COALESCE(c.total_credito,0),COALESCE(c.diferencia,0),COALESCE(c.observaciones,''),COALESCE(c.fecha_creacion,''),COALESCE(c.fecha_actualizacion,''),COALESCE(c.usuario_creador,'') FROM empresa_contabilidad_colombia_comprobantes c LEFT JOIN empresa_contabilidad_colombia_terceros t ON t.id=c.tercero_id AND t.empresa_id=c.empresa_id WHERE `+where+` ORDER BY c.id DESC LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaContabilidadComprobante
	for rows.Next() {
		var x EmpresaContabilidadComprobante
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Codigo, &x.TipoComprobante, &x.FechaComprobante, &x.PeriodoContable, &x.TerceroID, &x.TerceroNombre, &x.Concepto, &x.OrigenModulo, &x.ReferenciaExterna, &x.Estado, &x.TotalDebito, &x.TotalCredito, &x.Diferencia, &x.Observaciones, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func GetEmpresaContabilidadComprobante(dbConn *sql.DB, empresaID, id int64) (EmpresaContabilidadComprobante, error) {
	rows, err := ListEmpresaContabilidadComprobantes(dbConn, empresaID, "", "", 500)
	if err != nil {
		return EmpresaContabilidadComprobante{}, err
	}
	var out EmpresaContabilidadComprobante
	for _, row := range rows {
		if row.ID == id {
			out = row
			break
		}
	}
	if out.ID == 0 {
		return EmpresaContabilidadComprobante{}, sql.ErrNoRows
	}
	lineRows, err := ExecQueryCompat(dbConn, `SELECT l.id,l.empresa_id,l.comprobante_id,l.cuenta_codigo,COALESCE(c.nombre,''),COALESCE(l.tercero_id,0),COALESCE(t.nombre,''),COALESCE(l.detalle,''),COALESCE(l.debito,0),COALESCE(l.credito,0),COALESCE(l.base_gravable,0),COALESCE(l.impuesto_codigo,''),COALESCE(l.centro_costo,'') FROM empresa_contabilidad_colombia_lineas l LEFT JOIN empresa_contabilidad_colombia_cuentas c ON c.empresa_id=l.empresa_id AND c.codigo=l.cuenta_codigo LEFT JOIN empresa_contabilidad_colombia_terceros t ON t.empresa_id=l.empresa_id AND t.id=l.tercero_id WHERE l.empresa_id=? AND l.comprobante_id=? ORDER BY l.id`, empresaID, id)
	if err != nil {
		return EmpresaContabilidadComprobante{}, err
	}
	defer lineRows.Close()
	for lineRows.Next() {
		var l EmpresaContabilidadAsientoLinea
		if err := lineRows.Scan(&l.ID, &l.EmpresaID, &l.ComprobanteID, &l.CuentaCodigo, &l.CuentaNombre, &l.TerceroID, &l.TerceroNombre, &l.Detalle, &l.Debito, &l.Credito, &l.BaseGravable, &l.ImpuestoCodigo, &l.CentroCosto); err != nil {
			return EmpresaContabilidadComprobante{}, err
		}
		out.Lineas = append(out.Lineas, l)
	}
	return out, lineRows.Err()
}

func CambiarEstadoEmpresaContabilidadComprobante(dbConn *sql.DB, empresaID, id int64, estado, usuario string) error {
	estado = firstContabilidadValue(estado, "anulado")
	res, err := ExecCompat(dbConn, `UPDATE empresa_contabilidad_colombia_comprobantes SET estado=?, fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=COALESCE(NULLIF(?,''),usuario_creador) WHERE empresa_id=? AND id=?`, estado, usuario, empresaID, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func CerrarEmpresaContabilidadPeriodo(dbConn *sql.DB, empresaID int64, periodo, usuario, observaciones string) error {
	periodo = strings.TrimSpace(periodo)
	if periodo == "" {
		return errors.New("periodo es obligatorio")
	}
	var deb, cred float64
	_ = QueryRowCompat(dbConn, `SELECT COALESCE(SUM(total_debito),0), COALESCE(SUM(total_credito),0) FROM empresa_contabilidad_colombia_comprobantes WHERE empresa_id=? AND periodo_contable=? AND estado='contabilizado'`, empresaID, periodo).Scan(&deb, &cred)
	diff := roundContabilidad(deb - cred)
	if math.Abs(diff) > 0.009 {
		return errors.New("no se puede cerrar un periodo descuadrado")
	}
	_, err := ExecCompat(dbConn, `INSERT INTO empresa_contabilidad_colombia_periodos (empresa_id,periodo,estado,total_debito,total_credito,diferencia,cerrado_por,fecha_cierre,observaciones,fecha_creacion,fecha_actualizacion,usuario_creador) VALUES (?,?,?,?,?,?,?,CURRENT_TIMESTAMP,?,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,?) ON CONFLICT (empresa_id,periodo) DO UPDATE SET estado='cerrado', total_debito=EXCLUDED.total_debito, total_credito=EXCLUDED.total_credito, diferencia=EXCLUDED.diferencia, cerrado_por=EXCLUDED.cerrado_por, fecha_cierre=CURRENT_TIMESTAMP, observaciones=EXCLUDED.observaciones, fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=EXCLUDED.usuario_creador`, empresaID, periodo, "cerrado", deb, cred, diff, usuario, strings.TrimSpace(observaciones), usuario)
	if err != nil {
		return err
	}
	t, _ := time.Parse("2006-01", periodo)
	_, _ = UpsertEmpresaCierreFiscalPeriodo(dbConn, EmpresaCierreFiscalPeriodo{
		EmpresaID:           empresaID,
		Periodo:             periodo,
		FechaDesde:          periodo + "-01",
		FechaHasta:          lastDayOfCierreFiscalMonth(t),
		TipoCierre:          "mensual",
		EstadoPeriodo:       "cerrado",
		BloqueaVentas:       false,
		BloqueaCompras:      false,
		BloqueaCaja:         false,
		BloqueaInventario:   false,
		BloqueaContabilidad: true,
		BloqueaFacturacion:  true,
		CerradoPor:          usuario,
		FechaCierre:         time.Now().Format(time.RFC3339),
		Motivo:              strings.TrimSpace(observaciones),
		Observaciones:       "Sincronizado desde cierre de Contabilidad Colombia.",
		UsuarioCreador:      usuario,
	})
	return nil
}

func ReabrirEmpresaContabilidadPeriodo(dbConn *sql.DB, empresaID int64, periodo, usuario, observaciones string) error {
	_, err := ExecCompat(dbConn, `UPDATE empresa_contabilidad_colombia_periodos SET estado='abierto', observaciones=?, fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=COALESCE(NULLIF(?,''),usuario_creador) WHERE empresa_id=? AND periodo=?`, strings.TrimSpace(observaciones), usuario, empresaID, strings.TrimSpace(periodo))
	if err != nil {
		return err
	}
	fiscales, _ := ListEmpresaCierreFiscalPeriodos(dbConn, empresaID, "", 240)
	for _, p := range fiscales {
		if p.Periodo == strings.TrimSpace(periodo) {
			_, _ = CambiarEstadoEmpresaCierreFiscalPeriodo(dbConn, empresaID, p.ID, "abierto", usuario, strings.TrimSpace(observaciones))
			break
		}
	}
	return nil
}

func BuildEmpresaContabilidadDashboard(dbConn *sql.DB, empresaID int64) (EmpresaContabilidadDashboard, error) {
	if err := SeedEmpresaContabilidadColombiaBase(dbConn, empresaID, "sistema"); err != nil {
		return EmpresaContabilidadDashboard{}, err
	}
	cfg, err := GetEmpresaContabilidadConfig(dbConn, empresaID)
	if err != nil {
		return EmpresaContabilidadDashboard{}, err
	}
	out := EmpresaContabilidadDashboard{EmpresaID: empresaID, Config: cfg}
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_contabilidad_colombia_cuentas WHERE empresa_id=?`, empresaID).Scan(&out.Cuentas)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_contabilidad_colombia_terceros WHERE empresa_id=?`, empresaID).Scan(&out.Terceros)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_contabilidad_colombia_impuestos WHERE empresa_id=?`, empresaID).Scan(&out.Impuestos)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*), COALESCE(SUM(total_debito),0), COALESCE(SUM(total_credito),0) FROM empresa_contabilidad_colombia_comprobantes WHERE empresa_id=? AND periodo_contable=? AND estado='contabilizado'`, empresaID, cfg.PeriodoActual).Scan(&out.ComprobantesMes, &out.TotalDebitoMes, &out.TotalCreditoMes)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_contabilidad_colombia_comprobantes WHERE empresa_id=? AND estado='borrador'`, empresaID).Scan(&out.ComprobantesBorrador)
	out.DiferenciaMes = roundContabilidad(out.TotalDebitoMes - out.TotalCreditoMes)
	out.UltimosComprobantes, _ = ListEmpresaContabilidadComprobantes(dbConn, empresaID, "", "", 12)
	out.Periodos, _ = ListEmpresaContabilidadPeriodos(dbConn, empresaID)
	return out, nil
}

func ListEmpresaContabilidadPeriodos(dbConn *sql.DB, empresaID int64) ([]EmpresaContabilidadPeriodo, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,periodo,COALESCE(estado,'abierto'),COALESCE(total_debito,0),COALESCE(total_credito,0),COALESCE(diferencia,0),COALESCE(cerrado_por,''),COALESCE(fecha_cierre,''),COALESCE(observaciones,''),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,''),COALESCE(usuario_creador,'') FROM empresa_contabilidad_colombia_periodos WHERE empresa_id=? ORDER BY periodo DESC LIMIT 36`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaContabilidadPeriodo
	for rows.Next() {
		var x EmpresaContabilidadPeriodo
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Periodo, &x.Estado, &x.TotalDebito, &x.TotalCredito, &x.Diferencia, &x.CerradoPor, &x.FechaCierre, &x.Observaciones, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func assertContabilidadPeriodoAbierto(dbConn *sql.DB, empresaID int64, periodo string) error {
	var estado string
	err := QueryRowCompat(dbConn, `SELECT COALESCE(estado,'abierto') FROM empresa_contabilidad_colombia_periodos WHERE empresa_id=? AND periodo=?`, empresaID, periodo).Scan(&estado)
	if err == nil && strings.ToLower(estado) == "cerrado" {
		return errors.New("periodo contable cerrado")
	}
	return nil
}

func nextContabilidadComprobanteCode(dbConn *sql.DB, empresaID int64, tipo string) (string, error) {
	prefix := strings.ToUpper(strings.ReplaceAll(firstContabilidadValue(tipo, "NC"), "_", "-")) + "-" + time.Now().Format("200601") + "-"
	var count int
	if err := QueryRowCompat(dbConn, `SELECT COUNT(*) FROM empresa_contabilidad_colombia_comprobantes WHERE empresa_id=? AND codigo LIKE ?`, empresaID, prefix+"%").Scan(&count); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%04d", prefix, count+1), nil
}

func normalizeContabilidadConfig(cfg EmpresaContabilidadConfig) EmpresaContabilidadConfig {
	cfg.NombreSistema = firstContabilidadValue(cfg.NombreSistema, "Contabilidad Colombia")
	cfg.Moneda = strings.ToUpper(firstContabilidadValue(cfg.Moneda, "COP"))
	cfg.PeriodoActual = firstContabilidadValue(cfg.PeriodoActual, time.Now().Format("2006-01"))
	cfg.PUCVersion = firstContabilidadValue(cfg.PUCVersion, "PUC Colombia base")
	cfg.BaseNIIF = firstContabilidadValue(cfg.BaseNIIF, "NIIF pymes")
	return cfg
}

func defaultPUCCuentas(empresaID int64, usuario string) []EmpresaContabilidadCuenta {
	rows := []struct {
		c, n, nat, t, p string
		mov, ter, imp   bool
	}{
		{"1105", "Caja", "debito", "mayor", "11", false, false, false},
		{"110505", "Caja general", "debito", "auxiliar", "1105", true, false, false},
		{"1110", "Bancos", "debito", "mayor", "11", false, false, false},
		{"111005", "Bancos moneda nacional", "debito", "auxiliar", "1110", true, false, false},
		{"130505", "Clientes nacionales", "debito", "auxiliar", "1305", true, true, false},
		{"135515", "Retención en la fuente por cobrar", "debito", "auxiliar", "1355", true, true, true},
		{"143505", "Inventarios mercancías no fabricadas", "debito", "auxiliar", "1435", true, false, false},
		{"220505", "Proveedores nacionales", "credito", "auxiliar", "2205", true, true, false},
		{"236505", "Retención en la fuente por pagar", "credito", "auxiliar", "2365", true, true, true},
		{"236805", "Impuesto de industria y comercio retenido", "credito", "auxiliar", "2368", true, true, true},
		{"240805", "IVA generado", "credito", "auxiliar", "2408", true, true, true},
		{"240810", "IVA descontable", "debito", "auxiliar", "2408", true, true, true},
		{"250505", "Salarios por pagar", "credito", "auxiliar", "2505", true, true, false},
		{"413595", "Ingresos por servicios y ventas", "credito", "auxiliar", "4135", true, false, false},
		{"417595", "Devoluciones en ventas", "debito", "auxiliar", "4175", true, false, false},
		{"510506", "Sueldos", "debito", "auxiliar", "5105", true, true, false},
		{"513505", "Servicios", "debito", "auxiliar", "5135", true, true, false},
		{"519595", "Gastos diversos", "debito", "auxiliar", "5195", true, true, false},
		{"613595", "Costo de ventas y prestación de servicios", "debito", "auxiliar", "6135", true, false, false},
	}
	out := make([]EmpresaContabilidadCuenta, 0, len(rows))
	for _, r := range rows {
		out = append(out, EmpresaContabilidadCuenta{EmpresaID: empresaID, Codigo: r.c, Nombre: r.n, Naturaleza: r.nat, TipoCuenta: r.t, CuentaPadre: r.p, AceptaMovimiento: r.mov, TerceroRequerido: r.ter, ImpuestoRequerido: r.imp, Estado: "activo", UsuarioCreador: usuario})
	}
	return out
}

func defaultContabilidadImpuestos(empresaID int64, usuario string) []EmpresaContabilidadImpuesto {
	return []EmpresaContabilidadImpuesto{
		{EmpresaID: empresaID, Codigo: "IVA19", Nombre: "IVA generado 19%", Tipo: "iva", Porcentaje: 19, CuentaDebito: "240810", CuentaCredito: "240805", Estado: "activo", UsuarioCreador: usuario},
		{EmpresaID: empresaID, Codigo: "IVA5", Nombre: "IVA generado 5%", Tipo: "iva", Porcentaje: 5, CuentaDebito: "240810", CuentaCredito: "240805", Estado: "activo", UsuarioCreador: usuario},
		{EmpresaID: empresaID, Codigo: "RETEFUENTE", Nombre: "Retención en la fuente", Tipo: "retencion_fuente", Porcentaje: 2.5, CuentaDebito: "135515", CuentaCredito: "236505", Estado: "activo", UsuarioCreador: usuario},
		{EmpresaID: empresaID, Codigo: "RETEICA", Nombre: "Retención ICA", Tipo: "reteica", Porcentaje: 0.966, CuentaDebito: "135515", CuentaCredito: "236805", Estado: "activo", UsuarioCreador: usuario},
	}
}

func cleanCode(v string) string { return strings.ReplaceAll(strings.TrimSpace(v), " ", "") }
func firstContabilidadValue(v, fallback string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return fallback
	}
	return v
}
func periodFromDate(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 7 {
		return v[:7]
	}
	return time.Now().Format("2006-01")
}
func roundContabilidad(v float64) float64 { return math.Round(v*100) / 100 }
