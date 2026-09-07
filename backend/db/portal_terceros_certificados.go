package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

type EmpresaPortalTercero struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	TipoTercero        string `json:"tipo_tercero"`
	TipoDocumento      string `json:"tipo_documento"`
	Documento          string `json:"documento"`
	DV                 string `json:"dv,omitempty"`
	RazonSocial        string `json:"razon_social"`
	Email              string `json:"email,omitempty"`
	Telefono           string `json:"telefono,omitempty"`
	Direccion          string `json:"direccion,omitempty"`
	Ciudad             string `json:"ciudad,omitempty"`
	Regimen            string `json:"regimen"`
	Estado             string `json:"estado"`
	AccesoToken        string `json:"acceso_token,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Certificados       int    `json:"certificados,omitempty"`
	Descargas          int    `json:"descargas,omitempty"`
}

type EmpresaCertificadoTributario struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	TerceroID          int64   `json:"tercero_id"`
	TipoCertificado    string  `json:"tipo_certificado"`
	NumeroCertificado  string  `json:"numero_certificado"`
	Anio               int     `json:"anio"`
	PeriodoDesde       string  `json:"periodo_desde"`
	PeriodoHasta       string  `json:"periodo_hasta"`
	Concepto           string  `json:"concepto"`
	BaseValor          float64 `json:"base_valor"`
	RetencionFuente    float64 `json:"retencion_fuente"`
	RetencionIVA       float64 `json:"retencion_iva"`
	RetencionICA       float64 `json:"retencion_ica"`
	OtrosValores       float64 `json:"otros_valores"`
	TotalCertificado   float64 `json:"total_certificado"`
	Moneda             string  `json:"moneda"`
	Estado             string  `json:"estado"`
	FirmaNombre        string  `json:"firma_nombre,omitempty"`
	FirmaCargo         string  `json:"firma_cargo,omitempty"`
	SelloURL           string  `json:"sello_url,omitempty"`
	PublicToken        string  `json:"public_token,omitempty"`
	FechaEmision       string  `json:"fecha_emision,omitempty"`
	FechaEnvio         string  `json:"fecha_envio,omitempty"`
	EnviadoAEmail      string  `json:"enviado_a_email,omitempty"`
	FechaAnulacion     string  `json:"fecha_anulacion,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	TerceroDocumento   string  `json:"tercero_documento,omitempty"`
	TerceroNombre      string  `json:"tercero_nombre,omitempty"`
	TerceroEmail       string  `json:"tercero_email,omitempty"`
	TerceroTipo        string  `json:"tercero_tipo,omitempty"`
}

type EmpresaCertificadoTributarioDescarga struct {
	ID              int64  `json:"id"`
	EmpresaID       int64  `json:"empresa_id"`
	CertificadoID   int64  `json:"certificado_id"`
	TerceroID       int64  `json:"tercero_id"`
	Canal           string `json:"canal"`
	IP              string `json:"ip,omitempty"`
	UserAgent       string `json:"user_agent,omitempty"`
	ValidacionClave string `json:"validacion_clave,omitempty"`
	FechaDescarga   string `json:"fecha_descarga,omitempty"`
}

type EmpresaPortalTercerosCertificadosDashboard struct {
	EmpresaID             int64                                  `json:"empresa_id"`
	TercerosActivos       int                                    `json:"terceros_activos"`
	CertificadosEmitidos  int                                    `json:"certificados_emitidos"`
	CertificadosBorrador  int                                    `json:"certificados_borrador"`
	CertificadosAnulados  int                                    `json:"certificados_anulados"`
	DescargasMes          int                                    `json:"descargas_mes"`
	RetencionesTotal      float64                                `json:"retenciones_total"`
	Terceros              []EmpresaPortalTercero                 `json:"terceros"`
	CertificadosRecientes []EmpresaCertificadoTributario         `json:"certificados_recientes"`
	DescargasRecientes    []EmpresaCertificadoTributarioDescarga `json:"descargas_recientes"`
	Alertas               []string                               `json:"alertas"`
}

func EnsureEmpresaPortalTercerosCertificadosSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_portal_terceros (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			tipo_tercero TEXT DEFAULT 'proveedor',
			tipo_documento TEXT DEFAULT 'NIT',
			documento TEXT NOT NULL,
			dv TEXT,
			razon_social TEXT NOT NULL,
			email TEXT,
			telefono TEXT,
			direccion TEXT,
			ciudad TEXT,
			regimen TEXT DEFAULT 'responsable_iva',
			estado TEXT DEFAULT 'activo',
			acceso_token TEXT,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			UNIQUE(empresa_id, documento)
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_portal_terceros_token ON empresa_portal_terceros(acceso_token)`,
		`CREATE INDEX IF NOT EXISTS ix_portal_terceros_empresa ON empresa_portal_terceros(empresa_id, tipo_tercero, estado)`,
		`CREATE TABLE IF NOT EXISTS empresa_certificados_tributarios (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			tercero_id INTEGER NOT NULL,
			tipo_certificado TEXT DEFAULT 'retencion_fuente',
			numero_certificado TEXT NOT NULL,
			anio INTEGER DEFAULT 0,
			periodo_desde TEXT,
			periodo_hasta TEXT,
			concepto TEXT,
			base_valor REAL DEFAULT 0,
			retencion_fuente REAL DEFAULT 0,
			retencion_iva REAL DEFAULT 0,
			retencion_ica REAL DEFAULT 0,
			otros_valores REAL DEFAULT 0,
			total_certificado REAL DEFAULT 0,
			moneda TEXT DEFAULT 'COP',
			estado TEXT DEFAULT 'borrador',
			firma_nombre TEXT,
			firma_cargo TEXT,
			sello_url TEXT,
			public_token TEXT,
			fecha_emision TEXT,
			fecha_envio TEXT,
			enviado_a_email TEXT,
			fecha_anulacion TEXT,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			UNIQUE(empresa_id, numero_certificado)
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_certificados_tributarios_token ON empresa_certificados_tributarios(public_token)`,
		`CREATE INDEX IF NOT EXISTS ix_certificados_tributarios_empresa ON empresa_certificados_tributarios(empresa_id, tercero_id, anio, estado)`,
		`CREATE TABLE IF NOT EXISTS empresa_certificados_tributarios_descargas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			certificado_id INTEGER NOT NULL,
			tercero_id INTEGER NOT NULL,
			canal TEXT DEFAULT 'portal_publico',
			ip TEXT,
			user_agent TEXT,
			validacion_clave TEXT,
			fecha_descarga TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS ix_certificados_descargas_empresa ON empresa_certificados_tributarios_descargas(empresa_id, certificado_id, fecha_descarga DESC)`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func BuildEmpresaPortalTercerosCertificadosDashboard(dbConn *sql.DB, empresaID int64) (EmpresaPortalTercerosCertificadosDashboard, error) {
	if err := EnsureEmpresaPortalTercerosCertificadosSchema(dbConn); err != nil {
		return EmpresaPortalTercerosCertificadosDashboard{}, err
	}
	terceros, err := ListEmpresaPortalTerceros(dbConn, empresaID, "", 100)
	if err != nil {
		return EmpresaPortalTercerosCertificadosDashboard{}, err
	}
	certs, _ := ListEmpresaCertificadosTributarios(dbConn, empresaID, "", "", 100)
	descargas, _ := ListEmpresaCertificadosTributariosDescargas(dbConn, empresaID, 30)
	d := EmpresaPortalTercerosCertificadosDashboard{EmpresaID: empresaID, Terceros: terceros, CertificadosRecientes: certs, DescargasRecientes: descargas}
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_portal_terceros WHERE empresa_id=? AND estado='activo'`, empresaID).Scan(&d.TercerosActivos)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_certificados_tributarios WHERE empresa_id=? AND estado='emitido'`, empresaID).Scan(&d.CertificadosEmitidos)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_certificados_tributarios WHERE empresa_id=? AND estado='borrador'`, empresaID).Scan(&d.CertificadosBorrador)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_certificados_tributarios WHERE empresa_id=? AND estado='anulado'`, empresaID).Scan(&d.CertificadosAnulados)
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_certificados_tributarios_descargas WHERE empresa_id=? AND substr(COALESCE(fecha_descarga,''),1,7)=?`, empresaID, time.Now().Format("2006-01")).Scan(&d.DescargasMes)
	_ = QueryRowCompat(dbConn, `SELECT COALESCE(SUM(retencion_fuente+retencion_iva+retencion_ica+otros_valores),0) FROM empresa_certificados_tributarios WHERE empresa_id=? AND estado='emitido'`, empresaID).Scan(&d.RetencionesTotal)
	if d.CertificadosBorrador > 0 {
		d.Alertas = append(d.Alertas, "Hay certificados en borrador pendientes por emitir.")
	}
	if d.TercerosActivos == 0 {
		d.Alertas = append(d.Alertas, "No hay terceros activos para certificados tributarios.")
	}
	if d.CertificadosEmitidos > 0 && d.DescargasMes == 0 {
		d.Alertas = append(d.Alertas, "Hay certificados emitidos sin descargas registradas este mes.")
	}
	if len(d.Alertas) == 0 {
		d.Alertas = append(d.Alertas, "Portal tributario sin alertas criticas.")
	}
	return d, nil
}

func UpsertEmpresaPortalTercero(dbConn *sql.DB, row EmpresaPortalTercero) (int64, error) {
	if err := EnsureEmpresaPortalTercerosCertificadosSchema(dbConn); err != nil {
		return 0, err
	}
	row.Documento = normalizePortalTercerosDocumento(row.Documento)
	row.RazonSocial = strings.TrimSpace(row.RazonSocial)
	if row.EmpresaID <= 0 || row.Documento == "" || row.RazonSocial == "" {
		return 0, errors.New("empresa_id, documento y razon_social son requeridos")
	}
	row.TipoTercero = normalizePortalTercerosTipo(row.TipoTercero)
	row.TipoDocumento = firstPortalTerceros(row.TipoDocumento, "NIT")
	row.Regimen = normalizePortalTercerosRegimen(row.Regimen)
	row.Estado = normalizePortalTercerosEstado(row.Estado)
	if row.AccesoToken == "" {
		row.AccesoToken = newPortalTercerosToken("ter")
	}
	if row.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_portal_terceros SET tipo_tercero=?,tipo_documento=?,documento=?,dv=?,razon_social=?,email=?,telefono=?,direccion=?,ciudad=?,regimen=?,estado=?,observaciones=?,fecha_actualizacion=CURRENT_TIMESTAMP,usuario_creador=? WHERE id=? AND empresa_id=?`, row.TipoTercero, row.TipoDocumento, row.Documento, row.DV, row.RazonSocial, row.Email, row.Telefono, row.Direccion, row.Ciudad, row.Regimen, row.Estado, row.Observaciones, row.UsuarioCreador, row.ID, row.EmpresaID)
		return row.ID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_portal_terceros (empresa_id,tipo_tercero,tipo_documento,documento,dv,razon_social,email,telefono,direccion,ciudad,regimen,estado,acceso_token,observaciones,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (empresa_id, documento) DO UPDATE SET tipo_tercero=EXCLUDED.tipo_tercero,tipo_documento=EXCLUDED.tipo_documento,dv=EXCLUDED.dv,razon_social=EXCLUDED.razon_social,email=EXCLUDED.email,telefono=EXCLUDED.telefono,direccion=EXCLUDED.direccion,ciudad=EXCLUDED.ciudad,regimen=EXCLUDED.regimen,estado=EXCLUDED.estado,observaciones=EXCLUDED.observaciones,fecha_actualizacion=CURRENT_TIMESTAMP,usuario_creador=EXCLUDED.usuario_creador`,
		row.EmpresaID, row.TipoTercero, row.TipoDocumento, row.Documento, row.DV, row.RazonSocial, row.Email, row.Telefono, row.Direccion, row.Ciudad, row.Regimen, row.Estado, row.AccesoToken, row.Observaciones, row.UsuarioCreador)
}

func UpsertEmpresaCertificadoTributario(dbConn *sql.DB, row EmpresaCertificadoTributario) (int64, error) {
	if err := EnsureEmpresaPortalTercerosCertificadosSchema(dbConn); err != nil {
		return 0, err
	}
	if row.EmpresaID <= 0 || row.TerceroID <= 0 {
		return 0, errors.New("empresa_id y tercero_id son requeridos")
	}
	row.TipoCertificado = normalizePortalCertificadoTipo(row.TipoCertificado)
	row.Estado = normalizePortalCertificadoEstado(row.Estado)
	row.Moneda = strings.ToUpper(firstPortalTerceros(row.Moneda, "COP"))
	if row.Anio <= 0 {
		row.Anio = time.Now().Year()
	}
	if row.PeriodoDesde == "" {
		row.PeriodoDesde = fmt.Sprintf("%d-01-01", row.Anio)
	}
	if row.PeriodoHasta == "" {
		row.PeriodoHasta = fmt.Sprintf("%d-12-31", row.Anio)
	}
	if row.NumeroCertificado == "" {
		row.NumeroCertificado = fmt.Sprintf("CERT-%d-%d-%s", row.EmpresaID, row.Anio, time.Now().Format("150405"))
	}
	if row.PublicToken == "" {
		row.PublicToken = newPortalTercerosToken("cert")
	}
	if row.FechaEmision == "" && row.Estado == "emitido" {
		row.FechaEmision = time.Now().Format("2006-01-02")
	}
	row.TotalCertificado = roundPortalTerceros(row.RetencionFuente + row.RetencionIVA + row.RetencionICA + row.OtrosValores)
	if row.ID > 0 {
		_, err := ExecCompat(dbConn, `UPDATE empresa_certificados_tributarios SET tercero_id=?,tipo_certificado=?,anio=?,periodo_desde=?,periodo_hasta=?,concepto=?,base_valor=?,retencion_fuente=?,retencion_iva=?,retencion_ica=?,otros_valores=?,total_certificado=?,moneda=?,estado=?,firma_nombre=?,firma_cargo=?,sello_url=?,fecha_emision=?,fecha_envio=?,enviado_a_email=?,fecha_anulacion=?,observaciones=?,fecha_actualizacion=CURRENT_TIMESTAMP,usuario_creador=? WHERE id=? AND empresa_id=?`, row.TerceroID, row.TipoCertificado, row.Anio, row.PeriodoDesde, row.PeriodoHasta, row.Concepto, row.BaseValor, row.RetencionFuente, row.RetencionIVA, row.RetencionICA, row.OtrosValores, row.TotalCertificado, row.Moneda, row.Estado, row.FirmaNombre, row.FirmaCargo, row.SelloURL, row.FechaEmision, row.FechaEnvio, row.EnviadoAEmail, row.FechaAnulacion, row.Observaciones, row.UsuarioCreador, row.ID, row.EmpresaID)
		return row.ID, err
	}
	return insertSQLCompat(dbConn, `INSERT INTO empresa_certificados_tributarios (empresa_id,tercero_id,tipo_certificado,numero_certificado,anio,periodo_desde,periodo_hasta,concepto,base_valor,retencion_fuente,retencion_iva,retencion_ica,otros_valores,total_certificado,moneda,estado,firma_nombre,firma_cargo,sello_url,public_token,fecha_emision,fecha_envio,enviado_a_email,fecha_anulacion,observaciones,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		row.EmpresaID, row.TerceroID, row.TipoCertificado, row.NumeroCertificado, row.Anio, row.PeriodoDesde, row.PeriodoHasta, row.Concepto, row.BaseValor, row.RetencionFuente, row.RetencionIVA, row.RetencionICA, row.OtrosValores, row.TotalCertificado, row.Moneda, row.Estado, row.FirmaNombre, row.FirmaCargo, row.SelloURL, row.PublicToken, row.FechaEmision, row.FechaEnvio, row.EnviadoAEmail, row.FechaAnulacion, row.Observaciones, row.UsuarioCreador)
}

func ListEmpresaPortalTerceros(dbConn *sql.DB, empresaID int64, q string, limit int) ([]EmpresaPortalTercero, error) {
	if err := EnsureEmpresaPortalTercerosCertificadosSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if strings.TrimSpace(q) != "" {
		where += " AND (LOWER(razon_social) LIKE ? OR LOWER(documento) LIKE ? OR LOWER(email) LIKE ?)"
		needle := "%" + strings.ToLower(strings.TrimSpace(q)) + "%"
		args = append(args, needle, needle, needle)
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,COALESCE(tipo_tercero,'proveedor'),COALESCE(tipo_documento,'NIT'),COALESCE(documento,''),COALESCE(dv,''),COALESCE(razon_social,''),COALESCE(email,''),COALESCE(telefono,''),COALESCE(direccion,''),COALESCE(ciudad,''),COALESCE(regimen,'responsable_iva'),COALESCE(estado,'activo'),COALESCE(acceso_token,''),COALESCE(observaciones,''),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,''),COALESCE(usuario_creador,'') FROM empresa_portal_terceros WHERE %s ORDER BY razon_social LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaPortalTercero{}
	for rows.Next() {
		var x EmpresaPortalTercero
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.TipoTercero, &x.TipoDocumento, &x.Documento, &x.DV, &x.RazonSocial, &x.Email, &x.Telefono, &x.Direccion, &x.Ciudad, &x.Regimen, &x.Estado, &x.AccesoToken, &x.Observaciones, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador); err != nil {
			return nil, err
		}
		_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_certificados_tributarios WHERE empresa_id=? AND tercero_id=?`, empresaID, x.ID).Scan(&x.Certificados)
		_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_certificados_tributarios_descargas WHERE empresa_id=? AND tercero_id=?`, empresaID, x.ID).Scan(&x.Descargas)
		out = append(out, x)
	}
	return out, rows.Err()
}

func ListEmpresaCertificadosTributarios(dbConn *sql.DB, empresaID int64, estado, tipo string, limit int) ([]EmpresaCertificadoTributario, error) {
	if err := EnsureEmpresaPortalTercerosCertificadosSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 1000 {
		limit = 300
	}
	args := []interface{}{empresaID}
	where := "c.empresa_id=?"
	if e := normalizePortalCertificadoEstado(estado); e != "" && e != "todos" {
		where += " AND c.estado=?"
		args = append(args, e)
	}
	if t := normalizePortalCertificadoTipo(tipo); strings.TrimSpace(tipo) != "" && t != "todos" {
		where += " AND c.tipo_certificado=?"
		args = append(args, t)
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT c.id,c.empresa_id,c.tercero_id,COALESCE(c.tipo_certificado,''),COALESCE(c.numero_certificado,''),COALESCE(c.anio,0),COALESCE(c.periodo_desde,''),COALESCE(c.periodo_hasta,''),COALESCE(c.concepto,''),COALESCE(c.base_valor,0),COALESCE(c.retencion_fuente,0),COALESCE(c.retencion_iva,0),COALESCE(c.retencion_ica,0),COALESCE(c.otros_valores,0),COALESCE(c.total_certificado,0),COALESCE(c.moneda,'COP'),COALESCE(c.estado,'borrador'),COALESCE(c.firma_nombre,''),COALESCE(c.firma_cargo,''),COALESCE(c.sello_url,''),COALESCE(c.public_token,''),COALESCE(c.fecha_emision,''),COALESCE(c.fecha_envio,''),COALESCE(c.enviado_a_email,''),COALESCE(c.fecha_anulacion,''),COALESCE(c.observaciones,''),COALESCE(c.fecha_creacion,''),COALESCE(c.fecha_actualizacion,''),COALESCE(c.usuario_creador,''),COALESCE(t.documento,''),COALESCE(t.razon_social,''),COALESCE(t.email,''),COALESCE(t.tipo_tercero,'')
		FROM empresa_certificados_tributarios c
		LEFT JOIN empresa_portal_terceros t ON t.empresa_id=c.empresa_id AND t.id=c.tercero_id
		WHERE %s ORDER BY c.anio DESC, c.id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCertificadoTributario{}
	for rows.Next() {
		var x EmpresaCertificadoTributario
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.TerceroID, &x.TipoCertificado, &x.NumeroCertificado, &x.Anio, &x.PeriodoDesde, &x.PeriodoHasta, &x.Concepto, &x.BaseValor, &x.RetencionFuente, &x.RetencionIVA, &x.RetencionICA, &x.OtrosValores, &x.TotalCertificado, &x.Moneda, &x.Estado, &x.FirmaNombre, &x.FirmaCargo, &x.SelloURL, &x.PublicToken, &x.FechaEmision, &x.FechaEnvio, &x.EnviadoAEmail, &x.FechaAnulacion, &x.Observaciones, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador, &x.TerceroDocumento, &x.TerceroNombre, &x.TerceroEmail, &x.TerceroTipo); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func GetEmpresaCertificadoTributarioByToken(dbConn *sql.DB, token string) (EmpresaCertificadoTributario, error) {
	if err := EnsureEmpresaPortalTercerosCertificadosSchema(dbConn); err != nil {
		return EmpresaCertificadoTributario{}, err
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return EmpresaCertificadoTributario{}, sql.ErrNoRows
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT c.id,c.empresa_id,c.tercero_id,COALESCE(c.tipo_certificado,''),COALESCE(c.numero_certificado,''),COALESCE(c.anio,0),COALESCE(c.periodo_desde,''),COALESCE(c.periodo_hasta,''),COALESCE(c.concepto,''),COALESCE(c.base_valor,0),COALESCE(c.retencion_fuente,0),COALESCE(c.retencion_iva,0),COALESCE(c.retencion_ica,0),COALESCE(c.otros_valores,0),COALESCE(c.total_certificado,0),COALESCE(c.moneda,'COP'),COALESCE(c.estado,'borrador'),COALESCE(c.firma_nombre,''),COALESCE(c.firma_cargo,''),COALESCE(c.sello_url,''),COALESCE(c.public_token,''),COALESCE(c.fecha_emision,''),COALESCE(c.fecha_envio,''),COALESCE(c.enviado_a_email,''),COALESCE(c.fecha_anulacion,''),COALESCE(c.observaciones,''),COALESCE(c.fecha_creacion,''),COALESCE(c.fecha_actualizacion,''),COALESCE(c.usuario_creador,''),COALESCE(t.documento,''),COALESCE(t.razon_social,''),COALESCE(t.email,''),COALESCE(t.tipo_tercero,'')
		FROM empresa_certificados_tributarios c
		LEFT JOIN empresa_portal_terceros t ON t.empresa_id=c.empresa_id AND t.id=c.tercero_id
		WHERE c.public_token=? AND c.estado IN ('emitido','enviado') LIMIT 1`, token)
	if err != nil {
		return EmpresaCertificadoTributario{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return EmpresaCertificadoTributario{}, sql.ErrNoRows
	}
	var x EmpresaCertificadoTributario
	err = rows.Scan(&x.ID, &x.EmpresaID, &x.TerceroID, &x.TipoCertificado, &x.NumeroCertificado, &x.Anio, &x.PeriodoDesde, &x.PeriodoHasta, &x.Concepto, &x.BaseValor, &x.RetencionFuente, &x.RetencionIVA, &x.RetencionICA, &x.OtrosValores, &x.TotalCertificado, &x.Moneda, &x.Estado, &x.FirmaNombre, &x.FirmaCargo, &x.SelloURL, &x.PublicToken, &x.FechaEmision, &x.FechaEnvio, &x.EnviadoAEmail, &x.FechaAnulacion, &x.Observaciones, &x.FechaCreacion, &x.FechaActualizacion, &x.UsuarioCreador, &x.TerceroDocumento, &x.TerceroNombre, &x.TerceroEmail, &x.TerceroTipo)
	return x, err
}

func CreateEmpresaCertificadoTributarioDescarga(dbConn *sql.DB, row EmpresaCertificadoTributarioDescarga) (int64, error) {
	if err := EnsureEmpresaPortalTercerosCertificadosSchema(dbConn); err != nil {
		return 0, err
	}
	if row.EmpresaID <= 0 || row.CertificadoID <= 0 || row.TerceroID <= 0 {
		return 0, errors.New("empresa_id, certificado_id y tercero_id son requeridos")
	}
	row.Canal = firstPortalTerceros(row.Canal, "portal_publico")
	return insertSQLCompat(dbConn, `INSERT INTO empresa_certificados_tributarios_descargas (empresa_id,certificado_id,tercero_id,canal,ip,user_agent,validacion_clave) VALUES (?,?,?,?,?,?,?)`, row.EmpresaID, row.CertificadoID, row.TerceroID, row.Canal, row.IP, row.UserAgent, row.ValidacionClave)
}

func ListEmpresaCertificadosTributariosDescargas(dbConn *sql.DB, empresaID int64, limit int) ([]EmpresaCertificadoTributarioDescarga, error) {
	if err := EnsureEmpresaPortalTercerosCertificadosSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,certificado_id,tercero_id,COALESCE(canal,''),COALESCE(ip,''),COALESCE(user_agent,''),COALESCE(validacion_clave,''),COALESCE(fecha_descarga,'') FROM empresa_certificados_tributarios_descargas WHERE empresa_id=? ORDER BY fecha_descarga DESC, id DESC LIMIT %d`, limit), empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaCertificadoTributarioDescarga{}
	for rows.Next() {
		var x EmpresaCertificadoTributarioDescarga
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.CertificadoID, &x.TerceroID, &x.Canal, &x.IP, &x.UserAgent, &x.ValidacionClave, &x.FechaDescarga); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func SeedEmpresaPortalTercerosCertificadosDemo(dbConn *sql.DB, empresaID int64, usuario string) error {
	if err := EnsureEmpresaPortalTercerosCertificadosSchema(dbConn); err != nil {
		return err
	}
	var count int
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_portal_terceros WHERE empresa_id=?`, empresaID).Scan(&count)
	if count > 0 {
		return nil
	}
	terceros := []EmpresaPortalTercero{
		{EmpresaID: empresaID, TipoTercero: "proveedor", TipoDocumento: "NIT", Documento: "900123456", DV: "7", RazonSocial: "Proveedor Servicios Integrales SAS", Email: "contabilidad@proveedor.test", Regimen: "responsable_iva", Ciudad: "Bogota", UsuarioCreador: usuario},
		{EmpresaID: empresaID, TipoTercero: "empleado", TipoDocumento: "CC", Documento: "1010101010", RazonSocial: "Laura Martinez", Email: "laura.martinez@test.local", Regimen: "persona_natural", Ciudad: "Medellin", UsuarioCreador: usuario},
	}
	for _, tercero := range terceros {
		id, err := UpsertEmpresaPortalTercero(dbConn, tercero)
		if err != nil {
			return err
		}
		cert := EmpresaCertificadoTributario{EmpresaID: empresaID, TerceroID: id, TipoCertificado: "retencion_fuente", Anio: time.Now().Year(), Concepto: "Servicios gravados y retenciones practicadas", BaseValor: 12500000, RetencionFuente: 500000, RetencionIVA: 190000, RetencionICA: 82000, Estado: "emitido", FirmaNombre: "Representante legal", FirmaCargo: "Gerencia", EnviadoAEmail: tercero.Email, UsuarioCreador: usuario}
		if tercero.TipoTercero == "empleado" {
			cert.TipoCertificado = "ingresos_retenciones"
			cert.Concepto = "Certificado de ingresos y retenciones laborales"
			cert.BaseValor = 38000000
			cert.RetencionFuente = 1200000
			cert.RetencionIVA = 0
			cert.RetencionICA = 0
		}
		if _, err := UpsertEmpresaCertificadoTributario(dbConn, cert); err != nil {
			return err
		}
	}
	return nil
}

func firstPortalTerceros(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return strings.TrimSpace(v)
}

func normalizePortalTercerosDocumento(v string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(v), ".", ""), " ", ""))
}

func normalizePortalTercerosTipo(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "cliente", "proveedor", "empleado", "contratista", "contador", "accionista", "otro":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "proveedor"
	}
}

func normalizePortalTercerosRegimen(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "responsable_iva", "no_responsable_iva", "simple", "gran_contribuyente", "persona_natural", "no_residente":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "responsable_iva"
	}
}

func normalizePortalTercerosEstado(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "activo", "inactivo", "bloqueado", "archivado":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "activo"
	}
}

func normalizePortalCertificadoTipo(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "retencion_fuente", "retencion_iva", "retencion_ica", "ingresos_retenciones", "certificado_proveedor", "certificado_cliente", "autorretenedores", "otro", "todos":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "retencion_fuente"
	}
}

func normalizePortalCertificadoEstado(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "borrador", "emitido", "enviado", "anulado", "vencido", "todos":
		return strings.ToLower(strings.TrimSpace(v))
	default:
		if strings.TrimSpace(v) == "" {
			return "borrador"
		}
		return "borrador"
	}
}

func newPortalTercerosToken(prefix string) string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return prefix + "_" + hex.EncodeToString(buf)
}

func roundPortalTerceros(v float64) float64 {
	if v == 0 {
		return 0
	}
	return float64(int64(v*100+0.5)) / 100
}
