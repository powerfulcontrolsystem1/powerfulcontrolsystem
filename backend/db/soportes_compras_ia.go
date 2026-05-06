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

const EmpresaSoporteComprasIAModeloDefault = "openai:gpt-5.5"

type EmpresaSoporteComprasIA struct {
	ID                     int64   `json:"id"`
	EmpresaID              int64   `json:"empresa_id"`
	Codigo                 string  `json:"codigo"`
	TipoSoporte            string  `json:"tipo_soporte"`
	EstadoSoporte          string  `json:"estado_soporte"`
	Origen                 string  `json:"origen"`
	ArchivoNombre          string  `json:"archivo_nombre"`
	ArchivoURL             string  `json:"archivo_url"`
	ArchivoMime            string  `json:"archivo_mime"`
	ArchivoHash            string  `json:"archivo_hash"`
	ProveedorID            int64   `json:"proveedor_id"`
	ProveedorNombre        string  `json:"proveedor_nombre"`
	ProveedorNIT           string  `json:"proveedor_nit"`
	DocumentoTipo          string  `json:"documento_tipo"`
	DocumentoNumero        string  `json:"documento_numero"`
	FechaDocumento         string  `json:"fecha_documento"`
	FechaVencimiento       string  `json:"fecha_vencimiento"`
	Subtotal               float64 `json:"subtotal"`
	ImpuestoIVA            float64 `json:"impuesto_iva"`
	RetencionFuente        float64 `json:"retencion_fuente"`
	RetencionICA           float64 `json:"retencion_ica"`
	RetencionIVA           float64 `json:"retencion_iva"`
	Total                  float64 `json:"total"`
	Moneda                 string  `json:"moneda"`
	CategoriaContable      string  `json:"categoria_contable"`
	CentroCosto            string  `json:"centro_costo"`
	ImpactaInventario      bool    `json:"impacta_inventario"`
	ConfianzaIA            float64 `json:"confianza_ia"`
	ModeloIA               string  `json:"modelo_ia"`
	ExtraccionJSON         string  `json:"extraccion_json"`
	RespuestaIA            string  `json:"respuesta_ia"`
	DuplicadoSoporteID     int64   `json:"duplicado_soporte_id"`
	RequiereRevisionHumana bool    `json:"requiere_revision_humana"`
	AprobadoPor            string  `json:"aprobado_por"`
	FechaAprobacion        string  `json:"fecha_aprobacion"`
	ConvertidoTipo         string  `json:"convertido_tipo"`
	ConvertidoID           int64   `json:"convertido_id"`
	Usuario                string  `json:"usuario_creador"`
	Estado                 string  `json:"estado"`
	Observaciones          string  `json:"observaciones"`
	FechaCreacion          string  `json:"fecha_creacion"`
	FechaActualizacion     string  `json:"fecha_actualizacion"`
}

type EmpresaSoporteComprasIAEvento struct {
	ID             int64  `json:"id"`
	EmpresaID      int64  `json:"empresa_id"`
	SoporteID      int64  `json:"soporte_id"`
	Evento         string `json:"evento"`
	EstadoAnterior string `json:"estado_anterior"`
	EstadoNuevo    string `json:"estado_nuevo"`
	DetalleJSON    string `json:"detalle_json"`
	Usuario        string `json:"usuario_creador"`
	FechaCreacion  string `json:"fecha_creacion"`
}

type EmpresaSoporteComprasIADashboard struct {
	EmpresaID          int64                     `json:"empresa_id"`
	Pendientes         int                       `json:"pendientes"`
	EnRevision         int                       `json:"en_revision"`
	Aprobados          int                       `json:"aprobados"`
	Contabilizados     int                       `json:"contabilizados"`
	Rechazados         int                       `json:"rechazados"`
	Duplicados         int                       `json:"duplicados"`
	TotalPendiente     float64                   `json:"total_pendiente"`
	TotalContabilizado float64                   `json:"total_contabilizado"`
	SoportesRecientes  []EmpresaSoporteComprasIA `json:"soportes_recientes"`
	Alertas            []string                  `json:"alertas"`
	ModeloRecomendado  string                    `json:"modelo_recomendado"`
}

func EnsureEmpresaSoportesComprasIASchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if err := EnsureEmpresaModulosFaltantesSchema(dbConn); err != nil {
		return err
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_soportes_compras_ia (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			tipo_soporte TEXT DEFAULT 'gasto',
			estado_soporte TEXT DEFAULT 'radicado',
			origen TEXT DEFAULT 'manual',
			archivo_nombre TEXT,
			archivo_url TEXT,
			archivo_mime TEXT,
			archivo_hash TEXT,
			proveedor_id INTEGER DEFAULT 0,
			proveedor_nombre TEXT,
			proveedor_nit TEXT,
			documento_tipo TEXT DEFAULT 'factura_compra',
			documento_numero TEXT,
			fecha_documento TEXT,
			fecha_vencimiento TEXT,
			subtotal REAL DEFAULT 0,
			impuesto_iva REAL DEFAULT 0,
			retencion_fuente REAL DEFAULT 0,
			retencion_ica REAL DEFAULT 0,
			retencion_iva REAL DEFAULT 0,
			total REAL DEFAULT 0,
			moneda TEXT DEFAULT 'COP',
			categoria_contable TEXT,
			centro_costo TEXT,
			impacta_inventario INTEGER DEFAULT 0,
			confianza_ia REAL DEFAULT 0,
			modelo_ia TEXT DEFAULT 'openai:gpt-5.5',
			extraccion_json TEXT,
			respuesta_ia TEXT,
			duplicado_soporte_id INTEGER DEFAULT 0,
			requiere_revision_humana INTEGER DEFAULT 1,
			aprobado_por TEXT,
			fecha_aprobacion TEXT,
			convertido_tipo TEXT,
			convertido_id INTEGER DEFAULT 0,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_soportes_compras_ia_empresa_estado ON empresa_soportes_compras_ia(empresa_id, estado_soporte, fecha_creacion DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_soportes_compras_ia_hash ON empresa_soportes_compras_ia(empresa_id, archivo_hash);`,
		`CREATE TABLE IF NOT EXISTS empresa_soportes_compras_ia_eventos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			soporte_id INTEGER NOT NULL,
			evento TEXT NOT NULL,
			estado_anterior TEXT,
			estado_nuevo TEXT,
			detalle_json TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_soportes_compras_ia_eventos_soporte ON empresa_soportes_compras_ia_eventos(empresa_id, soporte_id, fecha_creacion DESC);`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func BuildEmpresaSoportesComprasIADashboard(dbConn *sql.DB, empresaID int64) (EmpresaSoporteComprasIADashboard, error) {
	if err := EnsureEmpresaSoportesComprasIASchema(dbConn); err != nil {
		return EmpresaSoporteComprasIADashboard{}, err
	}
	rows, err := listEmpresaSoportesComprasIA(dbConn, empresaID, "", 200)
	if err != nil {
		return EmpresaSoporteComprasIADashboard{}, err
	}
	d := EmpresaSoporteComprasIADashboard{EmpresaID: empresaID, ModeloRecomendado: EmpresaSoporteComprasIAModeloDefault, SoportesRecientes: rows}
	if len(d.SoportesRecientes) > 12 {
		d.SoportesRecientes = d.SoportesRecientes[:12]
	}
	for _, s := range rows {
		switch s.EstadoSoporte {
		case "radicado", "extraido":
			d.Pendientes++
			d.TotalPendiente += s.Total
		case "en_revision":
			d.EnRevision++
			d.TotalPendiente += s.Total
		case "aprobado":
			d.Aprobados++
		case "contabilizado":
			d.Contabilizados++
			d.TotalContabilizado += s.Total
		case "rechazado":
			d.Rechazados++
		case "duplicado":
			d.Duplicados++
		}
	}
	if d.Pendientes+d.EnRevision > 0 {
		d.Alertas = append(d.Alertas, "Hay soportes pendientes de revision antes de contabilizar.")
	}
	if d.Duplicados > 0 {
		d.Alertas = append(d.Alertas, "Se detectaron soportes duplicados por hash o documento.")
	}
	if len(rows) == 0 {
		d.Alertas = append(d.Alertas, "No hay soportes radicados. Carga una foto o PDF para iniciar.")
	}
	return d, nil
}

func CreateEmpresaSoporteComprasIA(dbConn *sql.DB, row EmpresaSoporteComprasIA) (EmpresaSoporteComprasIA, error) {
	if err := EnsureEmpresaSoportesComprasIASchema(dbConn); err != nil {
		return row, err
	}
	row = NormalizeEmpresaSoporteComprasIA(row)
	if row.EmpresaID <= 0 {
		return row, errors.New("empresa_id es obligatorio")
	}
	if row.Codigo == "" {
		row.Codigo = nextSoporteComprasIACode(dbConn, row.EmpresaID)
	}
	if row.ModeloIA == "" {
		row.ModeloIA = EmpresaSoporteComprasIAModeloDefault
	}
	if row.ArchivoHash != "" {
		row.DuplicadoSoporteID = findEmpresaSoporteComprasIADuplicado(dbConn, row.EmpresaID, row.ArchivoHash, row.DocumentoNumero)
		if row.DuplicadoSoporteID > 0 {
			row.EstadoSoporte = "duplicado"
			row.RequiereRevisionHumana = true
		}
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_soportes_compras_ia (
		empresa_id,codigo,tipo_soporte,estado_soporte,origen,archivo_nombre,archivo_url,archivo_mime,archivo_hash,
		proveedor_id,proveedor_nombre,proveedor_nit,documento_tipo,documento_numero,fecha_documento,fecha_vencimiento,
		subtotal,impuesto_iva,retencion_fuente,retencion_ica,retencion_iva,total,moneda,categoria_contable,centro_costo,
		impacta_inventario,confianza_ia,modelo_ia,extraccion_json,respuesta_ia,duplicado_soporte_id,requiere_revision_humana,
		usuario_creador,estado,observaciones
	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		row.EmpresaID, row.Codigo, row.TipoSoporte, row.EstadoSoporte, row.Origen, row.ArchivoNombre, row.ArchivoURL, row.ArchivoMime, row.ArchivoHash,
		row.ProveedorID, row.ProveedorNombre, row.ProveedorNIT, row.DocumentoTipo, row.DocumentoNumero, row.FechaDocumento, row.FechaVencimiento,
		row.Subtotal, row.ImpuestoIVA, row.RetencionFuente, row.RetencionICA, row.RetencionIVA, row.Total, row.Moneda, row.CategoriaContable, row.CentroCosto,
		boolToIntSoporteIA(row.ImpactaInventario), row.ConfianzaIA, row.ModeloIA, row.ExtraccionJSON, row.RespuestaIA, row.DuplicadoSoporteID, boolToIntSoporteIA(row.RequiereRevisionHumana),
		row.Usuario, row.Estado, row.Observaciones)
	if err != nil {
		return row, err
	}
	row.ID = id
	_ = InsertEmpresaSoporteComprasIAEvento(dbConn, row.EmpresaID, row.ID, "radicar", "", row.EstadoSoporte, row.Usuario, map[string]interface{}{"archivo": row.ArchivoNombre, "modelo": row.ModeloIA})
	return row, nil
}

func UpdateEmpresaSoporteComprasIAExtraccion(dbConn *sql.DB, empresaID, soporteID int64, extracted EmpresaSoporteComprasIA, usuario string) (EmpresaSoporteComprasIA, error) {
	if err := EnsureEmpresaSoportesComprasIASchema(dbConn); err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	current, err := GetEmpresaSoporteComprasIA(dbConn, empresaID, soporteID)
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	extracted.EmpresaID = empresaID
	extracted.ID = soporteID
	extracted.Codigo = current.Codigo
	extracted.ArchivoNombre = current.ArchivoNombre
	extracted.ArchivoURL = current.ArchivoURL
	extracted.ArchivoMime = current.ArchivoMime
	extracted.ArchivoHash = current.ArchivoHash
	extracted.Usuario = usuario
	extracted.Estado = current.Estado
	extracted.EstadoSoporte = "extraido"
	if extracted.ConfianzaIA < 0.85 {
		extracted.EstadoSoporte = "en_revision"
		extracted.RequiereRevisionHumana = true
	}
	if extracted.ModeloIA == "" {
		extracted.ModeloIA = current.ModeloIA
	}
	extracted = NormalizeEmpresaSoporteComprasIA(extracted)
	dupID := findEmpresaSoporteComprasIADuplicado(dbConn, empresaID, extracted.ArchivoHash, extracted.DocumentoNumero)
	if dupID > 0 && dupID != soporteID {
		extracted.DuplicadoSoporteID = dupID
		extracted.EstadoSoporte = "duplicado"
		extracted.RequiereRevisionHumana = true
	}
	_, err = ExecCompat(dbConn, `UPDATE empresa_soportes_compras_ia SET
		tipo_soporte=?,estado_soporte=?,proveedor_id=?,proveedor_nombre=?,proveedor_nit=?,documento_tipo=?,documento_numero=?,
		fecha_documento=?,fecha_vencimiento=?,subtotal=?,impuesto_iva=?,retencion_fuente=?,retencion_ica=?,retencion_iva=?,
		total=?,moneda=?,categoria_contable=?,centro_costo=?,impacta_inventario=?,confianza_ia=?,modelo_ia=?,extraccion_json=?,
		respuesta_ia=?,duplicado_soporte_id=?,requiere_revision_humana=?,fecha_actualizacion=datetime('now','localtime'),
		usuario_creador=?,observaciones=? WHERE empresa_id=? AND id=?`,
		extracted.TipoSoporte, extracted.EstadoSoporte, extracted.ProveedorID, extracted.ProveedorNombre, extracted.ProveedorNIT, extracted.DocumentoTipo, extracted.DocumentoNumero,
		extracted.FechaDocumento, extracted.FechaVencimiento, extracted.Subtotal, extracted.ImpuestoIVA, extracted.RetencionFuente, extracted.RetencionICA, extracted.RetencionIVA,
		extracted.Total, extracted.Moneda, extracted.CategoriaContable, extracted.CentroCosto, boolToIntSoporteIA(extracted.ImpactaInventario), extracted.ConfianzaIA, extracted.ModeloIA, extracted.ExtraccionJSON,
		extracted.RespuestaIA, extracted.DuplicadoSoporteID, boolToIntSoporteIA(extracted.RequiereRevisionHumana), usuario, extracted.Observaciones, empresaID, soporteID)
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	_ = InsertEmpresaSoporteComprasIAEvento(dbConn, empresaID, soporteID, "extraer_ia", current.EstadoSoporte, extracted.EstadoSoporte, usuario, map[string]interface{}{"modelo": extracted.ModeloIA, "confianza": extracted.ConfianzaIA})
	return GetEmpresaSoporteComprasIA(dbConn, empresaID, soporteID)
}

func UpdateEmpresaSoporteComprasIAEstado(dbConn *sql.DB, empresaID, soporteID int64, estado, usuario, observaciones string) (EmpresaSoporteComprasIA, error) {
	current, err := GetEmpresaSoporteComprasIA(dbConn, empresaID, soporteID)
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	next := normalizeSoporteIAEstado(estado)
	if next == "contabilizado" {
		return EmpresaSoporteComprasIA{}, errors.New("usa la accion contabilizar para convertir el soporte")
	}
	aprobadoPor := current.AprobadoPor
	fechaAprobacion := current.FechaAprobacion
	if next == "aprobado" {
		aprobadoPor = strings.TrimSpace(usuario)
		fechaAprobacion = time.Now().Format("2006-01-02 15:04:05")
	}
	_, err = ExecCompat(dbConn, `UPDATE empresa_soportes_compras_ia SET estado_soporte=?, aprobado_por=?, fecha_aprobacion=?, requiere_revision_humana=?, fecha_actualizacion=datetime('now','localtime'), usuario_creador=?, observaciones=? WHERE empresa_id=? AND id=?`,
		next, aprobadoPor, fechaAprobacion, boolToIntSoporteIA(next != "aprobado"), usuario, observaciones, empresaID, soporteID)
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	_ = InsertEmpresaSoporteComprasIAEvento(dbConn, empresaID, soporteID, "estado", current.EstadoSoporte, next, usuario, map[string]interface{}{"observaciones": observaciones})
	return GetEmpresaSoporteComprasIA(dbConn, empresaID, soporteID)
}

func ContabilizarEmpresaSoporteComprasIA(dbConn *sql.DB, empresaID, soporteID int64, usuario string) (EmpresaSoporteComprasIA, error) {
	if err := EnsureEmpresaSoportesComprasIASchema(dbConn); err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	row, err := GetEmpresaSoporteComprasIA(dbConn, empresaID, soporteID)
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	if row.EstadoSoporte != "aprobado" {
		return EmpresaSoporteComprasIA{}, errors.New("el soporte debe estar aprobado antes de contabilizar")
	}
	codigo := "CXP-" + strings.TrimSpace(row.Codigo)
	if codigo == "CXP-" {
		codigo = nextSoporteComprasIACode(dbConn, empresaID)
	}
	cxpID, err := insertSQLCompat(dbConn, `INSERT INTO empresa_cuentas_por_pagar (empresa_id,codigo,proveedor_id,proveedor_nombre,documento_tipo,documento_codigo,fecha_emision,fecha_vencimiento,dias_mora,valor_original,valor_pagado,saldo,estado_cartera,moneda,periodo_contable,usuario_creador,estado,observaciones) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		empresaID, codigo, row.ProveedorID, row.ProveedorNombre, row.DocumentoTipo, row.DocumentoNumero, row.FechaDocumento, row.FechaVencimiento, 0, row.Total, 0, row.Total, "pendiente", row.Moneda, periodoFromFechaSoporteIA(row.FechaDocumento), usuario, "activo", "Creado desde captura inteligente de soporte "+row.Codigo)
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	_, err = ExecCompat(dbConn, `UPDATE empresa_soportes_compras_ia SET estado_soporte='contabilizado', convertido_tipo='cuenta_por_pagar', convertido_id=?, requiere_revision_humana=0, fecha_actualizacion=datetime('now','localtime'), usuario_creador=? WHERE empresa_id=? AND id=?`, cxpID, usuario, empresaID, soporteID)
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	_ = InsertEmpresaSoporteComprasIAEvento(dbConn, empresaID, soporteID, "contabilizar", row.EstadoSoporte, "contabilizado", usuario, map[string]interface{}{"cuenta_por_pagar_id": cxpID})
	return GetEmpresaSoporteComprasIA(dbConn, empresaID, soporteID)
}

func GetEmpresaSoporteComprasIA(dbConn *sql.DB, empresaID, id int64) (EmpresaSoporteComprasIA, error) {
	rows, err := ExecQueryCompat(dbConn, soporteComprasIASelectSQL()+` WHERE empresa_id=? AND id=?`, empresaID, id)
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return EmpresaSoporteComprasIA{}, sql.ErrNoRows
	}
	return scanEmpresaSoporteComprasIA(rows)
}

func ListEmpresaSoportesComprasIA(dbConn *sql.DB, empresaID int64, estado string, limit int) ([]EmpresaSoporteComprasIA, error) {
	if err := EnsureEmpresaSoportesComprasIASchema(dbConn); err != nil {
		return nil, err
	}
	return listEmpresaSoportesComprasIA(dbConn, empresaID, estado, limit)
}

func listEmpresaSoportesComprasIA(dbConn *sql.DB, empresaID int64, estado string, limit int) ([]EmpresaSoporteComprasIA, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	where := "empresa_id=? AND COALESCE(estado,'activo')='activo'"
	args := []interface{}{empresaID}
	if e := normalizeSoporteIAEstado(estado); e != "" && e != "todos" {
		where += " AND estado_soporte=?"
		args = append(args, e)
	}
	args = append(args, limit)
	rows, err := ExecQueryCompat(dbConn, soporteComprasIASelectSQL()+` WHERE `+where+` ORDER BY fecha_creacion DESC, id DESC LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaSoporteComprasIA{}
	for rows.Next() {
		row, err := scanEmpresaSoporteComprasIA(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func ListEmpresaSoportesComprasIAEventos(dbConn *sql.DB, empresaID, soporteID int64, limit int) ([]EmpresaSoporteComprasIAEvento, error) {
	if limit <= 0 || limit > 300 {
		limit = 100
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,soporte_id,COALESCE(evento,''),COALESCE(estado_anterior,''),COALESCE(estado_nuevo,''),COALESCE(detalle_json,''),COALESCE(usuario_creador,''),COALESCE(fecha_creacion,'') FROM empresa_soportes_compras_ia_eventos WHERE empresa_id=? AND soporte_id=? ORDER BY fecha_creacion DESC, id DESC LIMIT ?`, empresaID, soporteID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaSoporteComprasIAEvento{}
	for rows.Next() {
		var row EmpresaSoporteComprasIAEvento
		if err := rows.Scan(&row.ID, &row.EmpresaID, &row.SoporteID, &row.Evento, &row.EstadoAnterior, &row.EstadoNuevo, &row.DetalleJSON, &row.Usuario, &row.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func InsertEmpresaSoporteComprasIAEvento(dbConn *sql.DB, empresaID, soporteID int64, evento, anterior, nuevo, usuario string, detalle map[string]interface{}) error {
	if dbConn == nil || empresaID <= 0 || soporteID <= 0 {
		return nil
	}
	raw, _ := json.Marshal(detalle)
	_, err := insertSQLCompat(dbConn, `INSERT INTO empresa_soportes_compras_ia_eventos (empresa_id,soporte_id,evento,estado_anterior,estado_nuevo,detalle_json,usuario_creador) VALUES (?,?,?,?,?,?,?)`, empresaID, soporteID, strings.TrimSpace(evento), strings.TrimSpace(anterior), strings.TrimSpace(nuevo), string(raw), strings.TrimSpace(usuario))
	return err
}

func NormalizeEmpresaSoporteComprasIA(row EmpresaSoporteComprasIA) EmpresaSoporteComprasIA {
	row.Codigo = strings.ToUpper(strings.TrimSpace(row.Codigo))
	row.TipoSoporte = normalizeSoporteIATipo(row.TipoSoporte)
	row.EstadoSoporte = normalizeSoporteIAEstado(row.EstadoSoporte)
	row.Origen = normalizeSoporteIAOrigen(row.Origen)
	row.DocumentoTipo = normalizeSoporteIADocumentoTipo(row.DocumentoTipo)
	row.Moneda = strings.ToUpper(strings.TrimSpace(row.Moneda))
	if row.Moneda == "" {
		row.Moneda = "COP"
	}
	row.ModeloIA = strings.TrimSpace(row.ModeloIA)
	if row.ModeloIA == "" {
		row.ModeloIA = EmpresaSoporteComprasIAModeloDefault
	}
	row.Estado = normalizeSoporteIAEstadoRegistro(row.Estado)
	if row.ConfianzaIA < 0 {
		row.ConfianzaIA = 0
	}
	if row.ConfianzaIA > 1 {
		row.ConfianzaIA = 1
	}
	if row.Total == 0 && (row.Subtotal != 0 || row.ImpuestoIVA != 0) {
		row.Total = soporteIARound(row.Subtotal + row.ImpuestoIVA - row.RetencionFuente - row.RetencionICA - row.RetencionIVA)
	}
	if row.EstadoSoporte == "" {
		row.EstadoSoporte = "radicado"
	}
	return row
}

func EmpresaSoporteComprasIAHashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func soporteComprasIASelectSQL() string {
	return `SELECT id,empresa_id,COALESCE(codigo,''),COALESCE(tipo_soporte,'gasto'),COALESCE(estado_soporte,'radicado'),COALESCE(origen,'manual'),COALESCE(archivo_nombre,''),COALESCE(archivo_url,''),COALESCE(archivo_mime,''),COALESCE(archivo_hash,''),COALESCE(proveedor_id,0),COALESCE(proveedor_nombre,''),COALESCE(proveedor_nit,''),COALESCE(documento_tipo,'factura_compra'),COALESCE(documento_numero,''),COALESCE(fecha_documento,''),COALESCE(fecha_vencimiento,''),COALESCE(subtotal,0),COALESCE(impuesto_iva,0),COALESCE(retencion_fuente,0),COALESCE(retencion_ica,0),COALESCE(retencion_iva,0),COALESCE(total,0),COALESCE(moneda,'COP'),COALESCE(categoria_contable,''),COALESCE(centro_costo,''),COALESCE(impacta_inventario,0),COALESCE(confianza_ia,0),COALESCE(modelo_ia,'openai:gpt-5.5'),COALESCE(extraccion_json,''),COALESCE(respuesta_ia,''),COALESCE(duplicado_soporte_id,0),COALESCE(requiere_revision_humana,1),COALESCE(aprobado_por,''),COALESCE(fecha_aprobacion,''),COALESCE(convertido_tipo,''),COALESCE(convertido_id,0),COALESCE(usuario_creador,''),COALESCE(estado,'activo'),COALESCE(observaciones,''),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,'') FROM empresa_soportes_compras_ia`
}

func scanEmpresaSoporteComprasIA(rows *sql.Rows) (EmpresaSoporteComprasIA, error) {
	var row EmpresaSoporteComprasIA
	var inventario, revision int
	err := rows.Scan(&row.ID, &row.EmpresaID, &row.Codigo, &row.TipoSoporte, &row.EstadoSoporte, &row.Origen, &row.ArchivoNombre, &row.ArchivoURL, &row.ArchivoMime, &row.ArchivoHash, &row.ProveedorID, &row.ProveedorNombre, &row.ProveedorNIT, &row.DocumentoTipo, &row.DocumentoNumero, &row.FechaDocumento, &row.FechaVencimiento, &row.Subtotal, &row.ImpuestoIVA, &row.RetencionFuente, &row.RetencionICA, &row.RetencionIVA, &row.Total, &row.Moneda, &row.CategoriaContable, &row.CentroCosto, &inventario, &row.ConfianzaIA, &row.ModeloIA, &row.ExtraccionJSON, &row.RespuestaIA, &row.DuplicadoSoporteID, &revision, &row.AprobadoPor, &row.FechaAprobacion, &row.ConvertidoTipo, &row.ConvertidoID, &row.Usuario, &row.Estado, &row.Observaciones, &row.FechaCreacion, &row.FechaActualizacion)
	row.ImpactaInventario = inventario != 0
	row.RequiereRevisionHumana = revision != 0
	return row, err
}

func findEmpresaSoporteComprasIADuplicado(dbConn *sql.DB, empresaID int64, hash, documento string) int64 {
	hash = strings.TrimSpace(hash)
	documento = strings.TrimSpace(documento)
	if hash == "" && documento == "" {
		return 0
	}
	where := "empresa_id=? AND COALESCE(estado,'activo')='activo'"
	args := []interface{}{empresaID}
	if hash != "" {
		where += " AND archivo_hash=?"
		args = append(args, hash)
	} else {
		where += " AND documento_numero=?"
		args = append(args, documento)
	}
	var id int64
	_ = QueryRowCompat(dbConn, `SELECT COALESCE(id,0) FROM empresa_soportes_compras_ia WHERE `+where+` ORDER BY id ASC LIMIT 1`, args...).Scan(&id)
	return id
}

func nextSoporteComprasIACode(dbConn *sql.DB, empresaID int64) string {
	var count int
	_ = QueryRowCompat(dbConn, `SELECT COUNT(1) FROM empresa_soportes_compras_ia WHERE empresa_id=?`, empresaID).Scan(&count)
	return fmt.Sprintf("SCI-%04d", count+1)
}

func periodoFromFechaSoporteIA(fecha string) string {
	fecha = strings.TrimSpace(fecha)
	if len(fecha) >= 7 {
		return fecha[:7]
	}
	return time.Now().Format("2006-01")
}

func normalizeSoporteIATipo(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "compra", "gasto", "documento_soporte", "factura_compra", "recibo", "servicio":
		return v
	default:
		return "gasto"
	}
}

func normalizeSoporteIAEstado(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "radicado", "extraido", "en_revision", "aprobado", "rechazado", "contabilizado", "duplicado", "todos":
		return v
	default:
		return "radicado"
	}
}

func normalizeSoporteIAOrigen(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "foto", "pdf", "xml", "email", "manual", "api":
		return v
	default:
		return "manual"
	}
}

func normalizeSoporteIADocumentoTipo(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "factura_compra", "documento_soporte", "cuenta_cobro", "recibo_caja", "gasto", "orden_compra", "otro":
		return v
	default:
		return "factura_compra"
	}
}

func normalizeSoporteIAEstadoRegistro(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "inactivo" || v == "archivado" {
		return v
	}
	return "activo"
}

func boolToIntSoporteIA(v bool) int {
	if v {
		return 1
	}
	return 0
}

func soporteIARound(v float64) float64 {
	if v < 0.005 && v > -0.005 {
		return 0
	}
	return float64(int64(v*100+0.5)) / 100
}
