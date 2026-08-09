package db

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const EmpresaSoporteComprasIAModeloDefault = "openai:gpt-5.5"

func empresaSoportesComprasIATablasCriticas() []string {
	return []string{
		"empresa_soportes_compras_ia",
		"empresa_soportes_compras_ia_eventos",
		"empresa_cuentas_por_pagar",
		"empresa_proveedores",
		"pcs_outbox_events",
	}
}

// EmpresaSoportesComprasIASchemaReady is deliberately read-only. Runtime
// handlers must fail closed when the migration/legacy catalog was not applied
// instead of issuing DDL while a financial conversion is in progress.
func EmpresaSoportesComprasIASchemaReady(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	for _, table := range empresaSoportesComprasIATablasCriticas() {
		var registered sql.NullString
		if err := queryRowSQLCompat(dbConn, `SELECT to_regclass(?)`, table).Scan(&registered); err != nil {
			return fmt.Errorf("verify soportes compras IA table %s: %w", table, err)
		}
		if !registered.Valid || strings.TrimSpace(registered.String) == "" {
			return fmt.Errorf("soportes compras IA table %s is missing; run pcs-migrate before starting the API", table)
		}
	}
	return nil
}

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
	EmpresaID           int64                     `json:"empresa_id"`
	Pendientes          int                       `json:"pendientes"`
	EnRevision          int                       `json:"en_revision"`
	Aprobados           int                       `json:"aprobados"`
	Contabilizados      int                       `json:"contabilizados"`
	Rechazados          int                       `json:"rechazados"`
	Duplicados          int                       `json:"duplicados"`
	TotalPendiente      float64                   `json:"total_pendiente"`
	TotalAprobado       float64                   `json:"total_aprobado"`
	TotalContabilizado  float64                   `json:"total_contabilizado"`
	ConfianzaPromedio   float64                   `json:"confianza_promedio"`
	RequierenRevision   int                       `json:"requieren_revision"`
	InventarioPendiente int                       `json:"inventario_pendiente"`
	PorVencer           int                       `json:"por_vencer"`
	Vencidos            int                       `json:"vencidos"`
	SoportesRecientes   []EmpresaSoporteComprasIA `json:"soportes_recientes"`
	Alertas             []string                  `json:"alertas"`
	ModeloRecomendado   string                    `json:"modelo_recomendado"`
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
			id BIGSERIAL PRIMARY KEY,
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
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_soportes_compras_ia_empresa_estado ON empresa_soportes_compras_ia(empresa_id, estado_soporte, fecha_creacion DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_soportes_compras_ia_hash ON empresa_soportes_compras_ia(empresa_id, archivo_hash);`,
		`CREATE TABLE IF NOT EXISTS empresa_soportes_compras_ia_eventos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			soporte_id INTEGER NOT NULL,
			evento TEXT NOT NULL,
			estado_anterior TEXT,
			estado_nuevo TEXT,
			detalle_json TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
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
	today := time.Now().Format("2006-01-02")
	limitSoon := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	confidenceCount := 0
	lowConfidence := 0
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
			d.TotalAprobado += s.Total
		case "contabilizado":
			d.Contabilizados++
			d.TotalContabilizado += s.Total
		case "rechazado":
			d.Rechazados++
		case "duplicado":
			d.Duplicados++
		}
		if s.RequiereRevisionHumana {
			d.RequierenRevision++
		}
		if s.ImpactaInventario && s.EstadoSoporte != "contabilizado" && s.EstadoSoporte != "rechazado" && s.EstadoSoporte != "duplicado" {
			d.InventarioPendiente++
		}
		if s.ConfianzaIA > 0 {
			d.ConfianzaPromedio += s.ConfianzaIA
			confidenceCount++
			if s.ConfianzaIA < 0.75 {
				lowConfidence++
			}
		}
		if soporteIAEstadoAbierto(s.EstadoSoporte) {
			due := strings.TrimSpace(s.FechaVencimiento)
			if len(due) >= 10 {
				due = due[:10]
				if due < today {
					d.Vencidos++
				} else if due <= limitSoon {
					d.PorVencer++
				}
			}
		}
	}
	if confidenceCount > 0 {
		d.ConfianzaPromedio = soporteIARound(d.ConfianzaPromedio / float64(confidenceCount))
	}
	if d.Pendientes+d.EnRevision > 0 {
		d.Alertas = append(d.Alertas, "Hay soportes pendientes de revision antes de contabilizar.")
	}
	if d.RequierenRevision > 0 {
		d.Alertas = append(d.Alertas, fmt.Sprintf("%d soporte(s) requieren validacion humana por confianza, duplicidad o datos incompletos.", d.RequierenRevision))
	}
	if d.Duplicados > 0 {
		d.Alertas = append(d.Alertas, "Se detectaron soportes duplicados por hash o documento.")
	}
	if d.Vencidos > 0 {
		d.Alertas = append(d.Alertas, fmt.Sprintf("%d soporte(s) tienen vencimiento vencido y siguen abiertos.", d.Vencidos))
	}
	if d.PorVencer > 0 {
		d.Alertas = append(d.Alertas, fmt.Sprintf("%d soporte(s) vencen en los proximos 7 dias.", d.PorVencer))
	}
	if d.InventarioPendiente > 0 {
		d.Alertas = append(d.Alertas, fmt.Sprintf("%d soporte(s) impactan inventario y aun no estan contabilizados.", d.InventarioPendiente))
	}
	if lowConfidence > 0 {
		d.Alertas = append(d.Alertas, fmt.Sprintf("%d extraccion(es) tienen confianza menor al 75%%.", lowConfidence))
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
	if !soporteIARegistroActivo(current.Estado) {
		return EmpresaSoporteComprasIA{}, errors.New("el soporte esta eliminado; recuperalo antes de ejecutar la extraccion IA")
	}
	if !soporteIAEstadoExtraible(current.EstadoSoporte) {
		return EmpresaSoporteComprasIA{}, errors.New("el estado actual no permite ejecutar nuevamente la extraccion IA")
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
	if extracted.ConfianzaIA < 0.85 || extracted.RequiereRevisionHumana {
		extracted.EstadoSoporte = "en_revision"
		extracted.RequiereRevisionHumana = true
	}
	if extracted.ModeloIA == "" {
		extracted.ModeloIA = current.ModeloIA
	}
	extracted = NormalizeEmpresaSoporteComprasIA(extracted)
	dupID := findEmpresaSoporteComprasIADuplicadoExcept(dbConn, empresaID, extracted.ArchivoHash, extracted.DocumentoNumero, soporteID)
	if dupID > 0 {
		extracted.DuplicadoSoporteID = dupID
		extracted.EstadoSoporte = "duplicado"
		extracted.RequiereRevisionHumana = true
	}
	result, err := ExecCompat(dbConn, `UPDATE empresa_soportes_compras_ia SET
		tipo_soporte=?,estado_soporte=?,proveedor_id=?,proveedor_nombre=?,proveedor_nit=?,documento_tipo=?,documento_numero=?,
		fecha_documento=?,fecha_vencimiento=?,subtotal=?,impuesto_iva=?,retencion_fuente=?,retencion_ica=?,retencion_iva=?,
		total=?,moneda=?,categoria_contable=?,centro_costo=?,impacta_inventario=?,confianza_ia=?,modelo_ia=?,extraccion_json=?,
		respuesta_ia=?,duplicado_soporte_id=?,requiere_revision_humana=?,fecha_actualizacion=CURRENT_TIMESTAMP,
		usuario_creador=?,observaciones=? WHERE empresa_id=? AND id=? AND estado_soporte=? AND COALESCE(estado,'activo')='activo'`,
		extracted.TipoSoporte, extracted.EstadoSoporte, extracted.ProveedorID, extracted.ProveedorNombre, extracted.ProveedorNIT, extracted.DocumentoTipo, extracted.DocumentoNumero,
		extracted.FechaDocumento, extracted.FechaVencimiento, extracted.Subtotal, extracted.ImpuestoIVA, extracted.RetencionFuente, extracted.RetencionICA, extracted.RetencionIVA,
		extracted.Total, extracted.Moneda, extracted.CategoriaContable, extracted.CentroCosto, boolToIntSoporteIA(extracted.ImpactaInventario), extracted.ConfianzaIA, extracted.ModeloIA, extracted.ExtraccionJSON,
		extracted.RespuestaIA, extracted.DuplicadoSoporteID, boolToIntSoporteIA(extracted.RequiereRevisionHumana), usuario, extracted.Observaciones, empresaID, soporteID, current.EstadoSoporte)
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return EmpresaSoporteComprasIA{}, errors.New("el soporte cambio de estado durante la extraccion IA")
	}
	_ = InsertEmpresaSoporteComprasIAEvento(dbConn, empresaID, soporteID, "extraer_ia", current.EstadoSoporte, extracted.EstadoSoporte, usuario, map[string]interface{}{"modelo": extracted.ModeloIA, "confianza": extracted.ConfianzaIA})
	return GetEmpresaSoporteComprasIA(dbConn, empresaID, soporteID)
}

func UpdateEmpresaSoporteComprasIAEstado(dbConn *sql.DB, empresaID, soporteID int64, estado, usuario, observaciones string) (EmpresaSoporteComprasIA, error) {
	if empresaID <= 0 || soporteID <= 0 {
		return EmpresaSoporteComprasIA{}, errors.New("empresa_id y soporte_id son obligatorios")
	}
	next := normalizeSoporteIAEstado(estado)
	if next != "aprobado" && next != "rechazado" {
		return EmpresaSoporteComprasIA{}, errors.New("transicion de soporte no permitida")
	}
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = "sistema"
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var currentState, recordState, documentoNumero string
	var proveedorID int64
	var total float64
	err = queryRowTxSQLCompat(tx, `SELECT COALESCE(estado_soporte,'radicado'), COALESCE(estado,'activo'), COALESCE(proveedor_id,0),
		COALESCE(documento_numero,''), COALESCE(total,0)
		FROM empresa_soportes_compras_ia WHERE empresa_id=? AND id=? FOR UPDATE`, empresaID, soporteID).
		Scan(&currentState, &recordState, &proveedorID, &documentoNumero, &total)
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	if !soporteIARegistroActivo(recordState) {
		return EmpresaSoporteComprasIA{}, errors.New("el soporte esta eliminado; recuperalo antes de cambiar su estado")
	}
	idempotent, err := validateSoporteIAStateTransition(currentState, next)
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	if idempotent {
		if err := tx.Commit(); err != nil {
			return EmpresaSoporteComprasIA{}, err
		}
		return GetEmpresaSoporteComprasIA(dbConn, empresaID, soporteID)
	}
	if next == "aprobado" {
		if proveedorID <= 0 || strings.TrimSpace(documentoNumero) == "" || soporteIARound(total) <= 0 {
			return EmpresaSoporteComprasIA{}, errors.New("para aprobar selecciona un proveedor registrado, documento y total mayor que cero")
		}
		var proveedorEstado string
		if err := queryRowTxSQLCompat(tx, `SELECT COALESCE(estado,'activo') FROM proveedores WHERE empresa_id=? AND id=?`, empresaID, proveedorID).Scan(&proveedorEstado); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return EmpresaSoporteComprasIA{}, errors.New("el proveedor no pertenece a la empresa activa")
			}
			return EmpresaSoporteComprasIA{}, err
		}
		if !strings.EqualFold(strings.TrimSpace(proveedorEstado), "activo") {
			return EmpresaSoporteComprasIA{}, errors.New("el proveedor seleccionado no esta activo")
		}
	}
	aprobadoPor := ""
	fechaAprobacion := ""
	if next == "aprobado" {
		aprobadoPor = usuario
		fechaAprobacion = time.Now().Format("2006-01-02 15:04:05")
	}
	result, err := execTxSQLCompat(tx, `UPDATE empresa_soportes_compras_ia SET estado_soporte=?, aprobado_por=?, fecha_aprobacion=?, requiere_revision_humana=?, fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=?, observaciones=? WHERE empresa_id=? AND id=? AND estado_soporte=?`,
		next, aprobadoPor, fechaAprobacion, boolToIntSoporteIA(next != "aprobado"), usuario, strings.TrimSpace(observaciones), empresaID, soporteID, currentState)
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return EmpresaSoporteComprasIA{}, errors.New("el soporte cambio de estado antes de confirmar la accion")
	}
	detail, _ := json.Marshal(map[string]interface{}{"observaciones": strings.TrimSpace(observaciones)})
	if _, err := insertTxSQLCompat(tx, `INSERT INTO empresa_soportes_compras_ia_eventos
		(empresa_id,soporte_id,evento,estado_anterior,estado_nuevo,detalle_json,usuario_creador)
		VALUES (?,?,?,?,?,?,?)`, empresaID, soporteID, "estado", currentState, next, string(detail), usuario); err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	if err := tx.Commit(); err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	return GetEmpresaSoporteComprasIA(dbConn, empresaID, soporteID)
}

func validateSoporteIAStateTransition(current, next string) (bool, error) {
	current = normalizeSoporteIAEstado(current)
	next = normalizeSoporteIAEstado(next)
	if current == next && (next == "aprobado" || next == "rechazado") {
		return true, nil
	}
	switch next {
	case "aprobado":
		switch current {
		case "radicado", "extraido", "en_revision":
			return false, nil
		}
	case "rechazado":
		switch current {
		case "radicado", "extraido", "en_revision", "aprobado":
			return false, nil
		}
	}
	return false, errors.New("el estado actual no permite esta accion")
}

// UpdateEmpresaSoporteComprasIARevision persists the human-reviewed values
// before approval. It never replaces the original file or the raw AI response.
// Editing an already approved support invalidates that approval and returns it
// to review, so the edited values cannot become a CxP without a new explicit
// approval.
func UpdateEmpresaSoporteComprasIARevision(dbConn *sql.DB, empresaID, soporteID int64, revision EmpresaSoporteComprasIA, usuario string) (EmpresaSoporteComprasIA, error) {
	if empresaID <= 0 || soporteID <= 0 {
		return EmpresaSoporteComprasIA{}, errors.New("empresa_id y soporte_id son obligatorios")
	}
	if err := EmpresaSoportesComprasIASchemaReady(dbConn); err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	current, err := GetEmpresaSoporteComprasIA(dbConn, empresaID, soporteID)
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	if !soporteIARegistroActivo(current.Estado) {
		return EmpresaSoporteComprasIA{}, errors.New("el soporte esta eliminado; recuperalo antes de editarlo")
	}
	if !soporteIAEstadoAbierto(current.EstadoSoporte) {
		return EmpresaSoporteComprasIA{}, errors.New("solo se pueden editar soportes pendientes de contabilizar")
	}
	revision = NormalizeEmpresaSoporteComprasIA(revision)
	if revision.ProveedorID > 0 {
		proveedor, err := GetEmpresaProveedorCxP(dbConn, empresaID, revision.ProveedorID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return EmpresaSoporteComprasIA{}, errors.New("el proveedor seleccionado no pertenece a la empresa activa")
			}
			return EmpresaSoporteComprasIA{}, err
		}
		if !strings.EqualFold(strings.TrimSpace(proveedor.Estado), "activo") {
			return EmpresaSoporteComprasIA{}, errors.New("el proveedor seleccionado no esta activo")
		}
		revision.ProveedorNombre = nombreProveedorCxPCanonico(proveedor.RazonSocial, proveedor.NombreComercial)
		revision.ProveedorNIT = strings.TrimSpace(proveedor.NIT)
	}
	duplicateID := findEmpresaSoporteComprasIADuplicadoExcept(dbConn, empresaID, current.ArchivoHash, revision.DocumentoNumero, soporteID)
	if duplicateID > 0 {
		return EmpresaSoporteComprasIA{}, errors.New("el documento o archivo ya fue radicado como soporte #" + strconv.FormatInt(duplicateID, 10))
	}
	nextState := current.EstadoSoporte
	if nextState == "aprobado" {
		nextState = "en_revision"
	}
	result, err := ExecCompat(dbConn, `UPDATE empresa_soportes_compras_ia SET
		tipo_soporte=?, estado_soporte=?, proveedor_id=?, proveedor_nombre=?, proveedor_nit=?, documento_tipo=?, documento_numero=?,
		fecha_documento=?, fecha_vencimiento=?, subtotal=?, impuesto_iva=?, retencion_fuente=?, retencion_ica=?, retencion_iva=?,
		total=?, moneda=?, categoria_contable=?, centro_costo=?, impacta_inventario=?, requiere_revision_humana=1,
		aprobado_por='', fecha_aprobacion='', fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=?, observaciones=?
		WHERE empresa_id=? AND id=? AND COALESCE(estado,'activo')='activo'`,
		revision.TipoSoporte, nextState, revision.ProveedorID, revision.ProveedorNombre, revision.ProveedorNIT, revision.DocumentoTipo, revision.DocumentoNumero,
		revision.FechaDocumento, revision.FechaVencimiento, revision.Subtotal, revision.ImpuestoIVA, revision.RetencionFuente, revision.RetencionICA, revision.RetencionIVA,
		revision.Total, revision.Moneda, revision.CategoriaContable, revision.CentroCosto, boolToIntSoporteIA(revision.ImpactaInventario), strings.TrimSpace(usuario), revision.Observaciones, empresaID, soporteID)
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return EmpresaSoporteComprasIA{}, errors.New("el soporte cambio mientras se guardaba la revision")
	}
	_ = InsertEmpresaSoporteComprasIAEvento(dbConn, empresaID, soporteID, "editar_revision", current.EstadoSoporte, nextState, usuario, map[string]interface{}{"campos": []string{"proveedor", "documento", "fechas", "impuestos", "total", "clasificacion"}})
	return GetEmpresaSoporteComprasIA(dbConn, empresaID, soporteID)
}

func ContabilizarEmpresaSoporteComprasIA(dbConn *sql.DB, empresaID, soporteID int64, usuario string) (EmpresaSoporteComprasIA, error) {
	if empresaID <= 0 || soporteID <= 0 {
		return EmpresaSoporteComprasIA{}, errors.New("empresa_id y soporte_id son obligatorios")
	}
	if err := EmpresaSoportesComprasIASchemaReady(dbConn); err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = "sistema"
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var codigo, estadoSoporte, recordState, documentoTipo, documentoNumero, fechaDocumento, fechaVencimiento, moneda string
	var proveedorID int64
	var total float64
	var convertidoID int64
	err = queryRowTxSQLCompat(tx, `SELECT COALESCE(codigo,''), COALESCE(estado_soporte,'radicado'), COALESCE(estado,'activo'), COALESCE(proveedor_id,0),
		COALESCE(documento_tipo,'factura_compra'), COALESCE(documento_numero,''),
		COALESCE(fecha_documento,''), COALESCE(fecha_vencimiento,''), COALESCE(total,0), COALESCE(moneda,'COP'), COALESCE(convertido_id,0)
		FROM empresa_soportes_compras_ia WHERE empresa_id=? AND id=? FOR UPDATE`, empresaID, soporteID).
		Scan(&codigo, &estadoSoporte, &recordState, &proveedorID, &documentoTipo, &documentoNumero, &fechaDocumento, &fechaVencimiento, &total, &moneda, &convertidoID)
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	if !soporteIARegistroActivo(recordState) {
		return EmpresaSoporteComprasIA{}, errors.New("el soporte esta eliminado; recuperalo antes de contabilizarlo")
	}
	if estadoSoporte == "contabilizado" && convertidoID > 0 {
		if err := tx.Commit(); err != nil {
			return EmpresaSoporteComprasIA{}, err
		}
		return GetEmpresaSoporteComprasIA(dbConn, empresaID, soporteID)
	}
	if estadoSoporte != "aprobado" {
		return EmpresaSoporteComprasIA{}, errors.New("el soporte debe estar aprobado antes de contabilizar")
	}
	if proveedorID <= 0 {
		return EmpresaSoporteComprasIA{}, errors.New("el soporte aprobado debe seleccionar un proveedor registrado de la empresa")
	}
	var proveedorEstado, proveedorNombre string
	if err := queryRowTxSQLCompat(tx, `SELECT COALESCE(estado,'activo'), COALESCE(nombre,'')
		FROM proveedores WHERE empresa_id=? AND id=?`, empresaID, proveedorID).Scan(&proveedorEstado, &proveedorNombre); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EmpresaSoporteComprasIA{}, errors.New("el proveedor no pertenece a la empresa activa")
		}
		return EmpresaSoporteComprasIA{}, err
	}
	if !strings.EqualFold(strings.TrimSpace(proveedorEstado), "activo") {
		return EmpresaSoporteComprasIA{}, errors.New("el proveedor seleccionado no esta activo")
	}
	proveedorNombreCanonico := strings.TrimSpace(proveedorNombre)
	if proveedorNombreCanonico == "" {
		return EmpresaSoporteComprasIA{}, errors.New("el proveedor seleccionado no tiene nombre operativo")
	}
	total = soporteIARound(total)
	if total <= 0 || strings.TrimSpace(documentoNumero) == "" {
		return EmpresaSoporteComprasIA{}, errors.New("el soporte aprobado requiere documento y total mayor que cero")
	}
	codigoCxP := "CXP-" + strings.TrimSpace(codigo)
	if codigoCxP == "CXP-" {
		return EmpresaSoporteComprasIA{}, errors.New("el soporte aprobado no tiene codigo operativo")
	}
	cxpID, err := insertTxSQLCompat(tx, `INSERT INTO empresa_cuentas_por_pagar
		(empresa_id,codigo,proveedor_id,proveedor_nombre,documento_tipo,documento_codigo,fecha_emision,fecha_vencimiento,dias_mora,valor_original,valor_pagado,saldo,estado_cartera,moneda,periodo_contable,usuario_creador,estado,observaciones)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, empresaID, codigoCxP, proveedorID, proveedorNombreCanonico, documentoTipo, documentoNumero,
		fechaDocumento, fechaVencimiento, 0, total, 0, total, "pendiente", strings.ToUpper(strings.TrimSpace(moneda)),
		periodoFromFechaSoporteIA(fechaDocumento), usuario, "activo", "Creado desde captura inteligente de soporte "+codigo)
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	updated, err := execTxSQLCompat(tx, `UPDATE empresa_soportes_compras_ia SET estado_soporte='contabilizado', convertido_tipo='cuenta_por_pagar', convertido_id=?, requiere_revision_humana=0, fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=? WHERE empresa_id=? AND id=? AND estado_soporte='aprobado'`, cxpID, usuario, empresaID, soporteID)
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	if affected, _ := updated.RowsAffected(); affected != 1 {
		return EmpresaSoporteComprasIA{}, errors.New("el soporte cambio de estado antes de contabilizar")
	}
	detalle, _ := json.Marshal(map[string]interface{}{"cuenta_por_pagar_id": cxpID, "proveedor_id": proveedorID, "documento": documentoNumero, "total": total})
	if _, err := insertTxSQLCompat(tx, `INSERT INTO empresa_soportes_compras_ia_eventos
		(empresa_id,soporte_id,evento,estado_anterior,estado_nuevo,detalle_json,usuario_creador)
		VALUES (?,?,?,?,?,?,?)`, empresaID, soporteID, "contabilizar", estadoSoporte, "contabilizado", string(detalle), usuario); err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	outboxPayload, _ := json.Marshal(map[string]interface{}{"soporte_id": soporteID, "cuenta_por_pagar_id": cxpID, "proveedor_id": proveedorID, "total": total})
	if err := InsertOutboxEvent(tx, OutboxEvent{EmpresaID: empresaID, Topic: "cuentas_por_pagar.soporte_ia_contabilizado", PayloadJSON: string(outboxPayload), IdempotencyKey: fmt.Sprintf("soporte-ia-cxp:%d:%d", empresaID, soporteID)}); err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	if err := tx.Commit(); err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	return GetEmpresaSoporteComprasIA(dbConn, empresaID, soporteID)
}

func nombreProveedorCxPCanonico(razonSocial, nombreComercial string) string {
	if razonSocial = strings.TrimSpace(razonSocial); razonSocial != "" {
		return razonSocial
	}
	return strings.TrimSpace(nombreComercial)
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

// GetEmpresaSoporteComprasIAActivo evita que un registro eliminado pueda
// reutilizarse en descargas o acciones operativas sin una restauracion previa.
func GetEmpresaSoporteComprasIAActivo(dbConn *sql.DB, empresaID, id int64) (EmpresaSoporteComprasIA, error) {
	row, err := GetEmpresaSoporteComprasIA(dbConn, empresaID, id)
	if err != nil {
		return row, err
	}
	if !soporteIARegistroActivo(row.Estado) {
		return EmpresaSoporteComprasIA{}, sql.ErrNoRows
	}
	return row, nil
}

func ListEmpresaSoportesComprasIA(dbConn *sql.DB, empresaID int64, estado string, limit int) ([]EmpresaSoporteComprasIA, error) {
	if err := EnsureEmpresaSoportesComprasIASchema(dbConn); err != nil {
		return nil, err
	}
	return listEmpresaSoportesComprasIARegistro(dbConn, empresaID, estado, "activo", limit)
}

func listEmpresaSoportesComprasIA(dbConn *sql.DB, empresaID int64, estado string, limit int) ([]EmpresaSoporteComprasIA, error) {
	return listEmpresaSoportesComprasIARegistro(dbConn, empresaID, estado, "activo", limit)
}

// ListEmpresaSoportesComprasIARegistro permite consultar la bandeja activa o
// la papelera sin mezclar empresas ni exponer estados arbitrarios.
func ListEmpresaSoportesComprasIARegistro(dbConn *sql.DB, empresaID int64, estado, registro string, limit int) ([]EmpresaSoporteComprasIA, error) {
	if err := EnsureEmpresaSoportesComprasIASchema(dbConn); err != nil {
		return nil, err
	}
	return listEmpresaSoportesComprasIARegistro(dbConn, empresaID, estado, registro, limit)
}

func ListEmpresaSoportesComprasIARetencion(dbConn *sql.DB, empresaID int64, retentionDays, limit int) ([]EmpresaSoporteComprasIA, error) {
	if empresaID <= 0 {
		return nil, errors.New("empresa_id es obligatorio")
	}
	if retentionDays < 1 || retentionDays > 3650 {
		return nil, errors.New("retencion_dias debe estar entre 1 y 3650")
	}
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	if err := EnsureEmpresaSoportesComprasIASchema(dbConn); err != nil {
		return nil, err
	}
	cutoff := time.Now().AddDate(0, 0, -retentionDays).Format("2006-01-02 15:04:05")
	rows, err := ExecQueryCompat(dbConn, soporteComprasIASelectSQL()+` WHERE empresa_id=?
		AND COALESCE(estado,'activo')='eliminado'
		AND COALESCE(estado_soporte,'radicado')<>'contabilizado'
		AND COALESCE(convertido_id,0)=0
		AND CASE
			WHEN COALESCE(NULLIF(fecha_actualizacion,''),fecha_creacion) ~ '^\d{4}-\d{2}-\d{2}'
			THEN CAST(COALESCE(NULLIF(fecha_actualizacion,''),fecha_creacion) AS TIMESTAMP)<=CAST(? AS TIMESTAMP)
			ELSE FALSE
		END
		ORDER BY fecha_actualizacion ASC,id ASC LIMIT ?`, empresaID, cutoff, limit)
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

func listEmpresaSoportesComprasIARegistro(dbConn *sql.DB, empresaID int64, estado, registro string, limit int) ([]EmpresaSoporteComprasIA, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	registro = normalizeSoporteIARegistroFiltro(registro)
	where := "empresa_id=? AND COALESCE(estado,'activo')=?"
	args := []interface{}{empresaID, registro}
	if e := normalizeSoporteIAEstadoFiltro(estado); e != "" && e != "todos" {
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

// UpdateEmpresaSoporteComprasIARegistroEstado implementa una papelera
// recuperable. El archivo y los eventos se conservan; un soporte contabilizado
// nunca puede ocultarse porque forma parte de la trazabilidad contable.
func UpdateEmpresaSoporteComprasIARegistroEstado(dbConn *sql.DB, empresaID, soporteID int64, siguiente, usuario, motivo string) (EmpresaSoporteComprasIA, error) {
	if empresaID <= 0 || soporteID <= 0 {
		return EmpresaSoporteComprasIA{}, errors.New("empresa_id y soporte_id son obligatorios")
	}
	siguiente = strings.ToLower(strings.TrimSpace(siguiente))
	if siguiente != "activo" && siguiente != "eliminado" {
		return EmpresaSoporteComprasIA{}, errors.New("estado de registro no permitido")
	}
	motivo = strings.TrimSpace(motivo)
	if motivo == "" {
		return EmpresaSoporteComprasIA{}, errors.New("el motivo es obligatorio")
	}
	if len([]rune(motivo)) > 500 {
		return EmpresaSoporteComprasIA{}, errors.New("el motivo no puede superar 500 caracteres")
	}
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = "sistema"
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var actual, estadoSoporte, archivoHash, documentoNumero string
	var convertidoID int64
	err = queryRowTxSQLCompat(tx, `SELECT COALESCE(estado,'activo'), COALESCE(estado_soporte,'radicado'), COALESCE(convertido_id,0), COALESCE(archivo_hash,''), COALESCE(documento_numero,'')
		FROM empresa_soportes_compras_ia WHERE empresa_id=? AND id=? FOR UPDATE`, empresaID, soporteID).
		Scan(&actual, &estadoSoporte, &convertidoID, &archivoHash, &documentoNumero)
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	actual = strings.ToLower(strings.TrimSpace(actual))
	if actual == "" {
		actual = "activo"
	}
	if actual != "activo" && actual != "eliminado" {
		return EmpresaSoporteComprasIA{}, errors.New("el soporte tiene un estado de registro no reconocido")
	}
	idempotente, err := validateSoporteIARegistroTransition(actual, estadoSoporte, convertidoID, siguiente)
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	if idempotente {
		if err := tx.Commit(); err != nil {
			return EmpresaSoporteComprasIA{}, err
		}
		return GetEmpresaSoporteComprasIA(dbConn, empresaID, soporteID)
	}
	if siguiente == "activo" && (strings.TrimSpace(archivoHash) != "" || strings.TrimSpace(documentoNumero) != "") {
		var duplicateID int64
		err = queryRowTxSQLCompat(tx, `SELECT id FROM empresa_soportes_compras_ia
			WHERE empresa_id=? AND id<>? AND COALESCE(estado,'activo')='activo'
			AND ((?<>'' AND archivo_hash=?) OR (?<>'' AND documento_numero=?))
			ORDER BY id LIMIT 1`, empresaID, soporteID, archivoHash, archivoHash, documentoNumero, documentoNumero).Scan(&duplicateID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return EmpresaSoporteComprasIA{}, err
		}
		if duplicateID > 0 {
			return EmpresaSoporteComprasIA{}, fmt.Errorf("no se puede recuperar: existe el soporte activo #%d con el mismo archivo o documento", duplicateID)
		}
	}
	result, err := execTxSQLCompat(tx, `UPDATE empresa_soportes_compras_ia SET estado=?, fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=? WHERE empresa_id=? AND id=? AND COALESCE(estado,'activo')=?`, siguiente, usuario, empresaID, soporteID, actual)
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return EmpresaSoporteComprasIA{}, errors.New("el soporte cambio mientras se confirmaba la accion")
	}
	evento := "restaurar"
	if siguiente == "eliminado" {
		evento = "eliminar"
	}
	detalle, _ := json.Marshal(map[string]interface{}{"motivo": motivo, "estado_registro_anterior": actual, "estado_registro_nuevo": siguiente})
	if _, err := insertTxSQLCompat(tx, `INSERT INTO empresa_soportes_compras_ia_eventos
		(empresa_id,soporte_id,evento,estado_anterior,estado_nuevo,detalle_json,usuario_creador)
		VALUES (?,?,?,?,?,?,?)`, empresaID, soporteID, evento, estadoSoporte, estadoSoporte, string(detalle), usuario); err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	if err := tx.Commit(); err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	return GetEmpresaSoporteComprasIA(dbConn, empresaID, soporteID)
}

// PurgeEmpresaSoporteComprasIA conserva el contrato transaccional para tareas
// sin archivo: inicia y finaliza la tumba. Los handlers con archivo usan las
// dos fases por separado para coordinar una saga reanudable con el filesystem.
func PurgeEmpresaSoporteComprasIA(dbConn *sql.DB, empresaID, soporteID int64, retentionDays int, confirmation, usuario, motivo string) (EmpresaSoporteComprasIA, error) {
	if _, err := BeginPurgeEmpresaSoporteComprasIA(dbConn, empresaID, soporteID, retentionDays, confirmation, usuario, motivo); err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	return FinalizePurgeEmpresaSoporteComprasIA(dbConn, empresaID, soporteID, usuario)
}

// BeginPurgeEmpresaSoporteComprasIA valida y marca una depuracion pendiente.
// Un reintento del mismo soporte pendiente es idempotente y no recalcula la
// retencion desde la fecha de inicio de la saga.
func BeginPurgeEmpresaSoporteComprasIA(dbConn *sql.DB, empresaID, soporteID int64, retentionDays int, confirmation, usuario, motivo string) (EmpresaSoporteComprasIA, error) {
	if empresaID <= 0 || soporteID <= 0 {
		return EmpresaSoporteComprasIA{}, errors.New("empresa_id y soporte_id son obligatorios")
	}
	if retentionDays < 1 || retentionDays > 3650 {
		return EmpresaSoporteComprasIA{}, errors.New("retencion_dias debe estar entre 1 y 3650")
	}
	motivo = strings.TrimSpace(motivo)
	if motivo == "" {
		return EmpresaSoporteComprasIA{}, errors.New("el motivo es obligatorio")
	}
	if len([]rune(motivo)) > 500 {
		return EmpresaSoporteComprasIA{}, errors.New("el motivo no puede superar 500 caracteres")
	}
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = "sistema"
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var codigo, registro, estadoSoporte, fechaRegistro string
	var convertidoID int64
	err = queryRowTxSQLCompat(tx, `SELECT COALESCE(codigo,''), COALESCE(estado,'activo'),
		COALESCE(estado_soporte,'radicado'), COALESCE(convertido_id,0),
		CAST(COALESCE(NULLIF(fecha_actualizacion,''),fecha_creacion) AS TEXT)
		FROM empresa_soportes_compras_ia WHERE empresa_id=? AND id=? FOR UPDATE`, empresaID, soporteID).
		Scan(&codigo, &registro, &estadoSoporte, &convertidoID, &fechaRegistro)
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	if !strings.EqualFold(strings.TrimSpace(codigo), strings.TrimSpace(confirmation)) {
		return EmpresaSoporteComprasIA{}, errors.New("la confirmacion no coincide con el codigo del soporte")
	}
	registro = strings.ToLower(strings.TrimSpace(registro))
	if registro == "purga_pendiente" || registro == "purgado" {
		if err := tx.Commit(); err != nil {
			return EmpresaSoporteComprasIA{}, err
		}
		return GetEmpresaSoporteComprasIA(dbConn, empresaID, soporteID)
	}
	if registro != "eliminado" {
		return EmpresaSoporteComprasIA{}, errors.New("solo un soporte de la papelera puede depurarse")
	}
	if normalizeSoporteIAEstado(estadoSoporte) == "contabilizado" || convertidoID > 0 {
		return EmpresaSoporteComprasIA{}, errors.New("un soporte contabilizado no puede depurarse porque conserva trazabilidad contable")
	}
	if !soporteComprasIARetentionEligible(fechaRegistro, retentionDays, time.Now()) {
		return EmpresaSoporteComprasIA{}, errors.New("el soporte aun no cumple la retencion configurada")
	}
	result, err := execTxSQLCompat(tx, `UPDATE empresa_soportes_compras_ia
		SET estado='purga_pendiente', fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=?
		WHERE empresa_id=? AND id=? AND COALESCE(estado,'activo')='eliminado'`, usuario, empresaID, soporteID)
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return EmpresaSoporteComprasIA{}, errors.New("el soporte cambio mientras se confirmaba la depuracion")
	}
	detalle, _ := json.Marshal(map[string]interface{}{
		"motivo": motivo, "retencion_dias": retentionDays,
		"estado_registro_anterior": "eliminado", "estado_registro_nuevo": "purga_pendiente",
		"archivo_privado": "cuarentena",
	})
	if _, err := insertTxSQLCompat(tx, `INSERT INTO empresa_soportes_compras_ia_eventos
		(empresa_id,soporte_id,evento,estado_anterior,estado_nuevo,detalle_json,usuario_creador)
		VALUES (?,?,?,?,?,?,?)`, empresaID, soporteID, "purgar_iniciar", estadoSoporte, estadoSoporte, string(detalle), usuario); err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	if err := tx.Commit(); err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	return GetEmpresaSoporteComprasIA(dbConn, empresaID, soporteID)
}

// FinalizePurgeEmpresaSoporteComprasIA invalida la URL solo después de que el
// handler confirma la eliminación del archivo en cuarentena.
func FinalizePurgeEmpresaSoporteComprasIA(dbConn *sql.DB, empresaID, soporteID int64, usuario string) (EmpresaSoporteComprasIA, error) {
	if empresaID <= 0 || soporteID <= 0 {
		return EmpresaSoporteComprasIA{}, errors.New("empresa_id y soporte_id son obligatorios")
	}
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = "sistema"
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	defer func() { _ = tx.Rollback() }()
	var registro, estadoSoporte string
	err = queryRowTxSQLCompat(tx, `SELECT COALESCE(estado,'activo'), COALESCE(estado_soporte,'radicado')
		FROM empresa_soportes_compras_ia WHERE empresa_id=? AND id=? FOR UPDATE`, empresaID, soporteID).Scan(&registro, &estadoSoporte)
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	registro = strings.ToLower(strings.TrimSpace(registro))
	if registro == "purgado" {
		if err := tx.Commit(); err != nil {
			return EmpresaSoporteComprasIA{}, err
		}
		return GetEmpresaSoporteComprasIA(dbConn, empresaID, soporteID)
	}
	if registro != "purga_pendiente" {
		return EmpresaSoporteComprasIA{}, errors.New("el soporte no tiene una depuracion pendiente")
	}
	result, err := execTxSQLCompat(tx, `UPDATE empresa_soportes_compras_ia SET estado='purgado', archivo_url='',
		fecha_actualizacion=CURRENT_TIMESTAMP, usuario_creador=? WHERE empresa_id=? AND id=? AND estado='purga_pendiente'`, usuario, empresaID, soporteID)
	if err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return EmpresaSoporteComprasIA{}, errors.New("el soporte cambio mientras se finalizaba la depuracion")
	}
	detalle, _ := json.Marshal(map[string]interface{}{
		"estado_registro_anterior": "purga_pendiente", "estado_registro_nuevo": "purgado", "archivo_privado": "eliminado",
	})
	if _, err := insertTxSQLCompat(tx, `INSERT INTO empresa_soportes_compras_ia_eventos
		(empresa_id,soporte_id,evento,estado_anterior,estado_nuevo,detalle_json,usuario_creador)
		VALUES (?,?,?,?,?,?,?)`, empresaID, soporteID, "purgar", estadoSoporte, estadoSoporte, string(detalle), usuario); err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	if err := tx.Commit(); err != nil {
		return EmpresaSoporteComprasIA{}, err
	}
	return GetEmpresaSoporteComprasIA(dbConn, empresaID, soporteID)
}

func soporteComprasIARetentionEligible(raw string, retentionDays int, now time.Time) bool {
	if retentionDays < 1 || retentionDays > 3650 || strings.TrimSpace(raw) == "" {
		return false
	}
	parsed, ok := parseSoporteComprasIATimestamp(raw, now.Location())
	if !ok {
		return false
	}
	return !parsed.After(now.AddDate(0, 0, -retentionDays))
}

func parseSoporteComprasIATimestamp(raw string, location *time.Location) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if location == nil {
		location = time.UTC
	}
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05.999999999Z07:00", "2006-01-02 15:04:05Z07:00", "2006-01-02 15:04:05.999999999-07", "2006-01-02 15:04:05-07"} {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return parsed, true
		}
	}
	for _, layout := range []string{"2006-01-02 15:04:05.999999999", "2006-01-02 15:04:05", "2006-01-02"} {
		if parsed, err := time.ParseInLocation(layout, raw, location); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func validateSoporteIARegistroTransition(actual, estadoSoporte string, convertidoID int64, siguiente string) (bool, error) {
	actual = strings.ToLower(strings.TrimSpace(actual))
	siguiente = strings.ToLower(strings.TrimSpace(siguiente))
	if actual == "" {
		actual = "activo"
	}
	if (actual != "activo" && actual != "eliminado") || (siguiente != "activo" && siguiente != "eliminado") {
		return false, errors.New("transicion de registro no permitida")
	}
	if actual == siguiente {
		return true, nil
	}
	if siguiente == "eliminado" {
		if actual != "activo" {
			return false, errors.New("solo un soporte activo puede enviarse a la papelera")
		}
		if normalizeSoporteIAEstado(estadoSoporte) == "contabilizado" || convertidoID > 0 {
			return false, errors.New("un soporte contabilizado no puede eliminarse porque conserva trazabilidad contable")
		}
		return false, nil
	}
	if siguiente == "activo" && actual == "eliminado" {
		return false, nil
	}
	return false, errors.New("transicion de registro no permitida")
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
	row.ArchivoNombre = strings.TrimSpace(row.ArchivoNombre)
	row.ArchivoURL = strings.TrimSpace(row.ArchivoURL)
	row.ArchivoMime = strings.TrimSpace(row.ArchivoMime)
	row.ArchivoHash = strings.TrimSpace(row.ArchivoHash)
	row.ProveedorNombre = strings.TrimSpace(row.ProveedorNombre)
	row.ProveedorNIT = strings.TrimSpace(row.ProveedorNIT)
	row.DocumentoNumero = strings.TrimSpace(row.DocumentoNumero)
	row.FechaDocumento = strings.TrimSpace(row.FechaDocumento)
	row.FechaVencimiento = strings.TrimSpace(row.FechaVencimiento)
	row.CategoriaContable = strings.TrimSpace(row.CategoriaContable)
	row.CentroCosto = strings.TrimSpace(row.CentroCosto)
	row.Observaciones = strings.TrimSpace(row.Observaciones)
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
	row.Subtotal = soporteIANonNegative(row.Subtotal)
	row.ImpuestoIVA = soporteIANonNegative(row.ImpuestoIVA)
	row.RetencionFuente = soporteIANonNegative(row.RetencionFuente)
	row.RetencionICA = soporteIANonNegative(row.RetencionICA)
	row.RetencionIVA = soporteIANonNegative(row.RetencionIVA)
	row.Total = soporteIANonNegative(row.Total)
	if row.Total == 0 && (row.Subtotal != 0 || row.ImpuestoIVA != 0) {
		row.Total = soporteIARound(row.Subtotal + row.ImpuestoIVA - row.RetencionFuente - row.RetencionICA - row.RetencionIVA)
		row.Total = soporteIANonNegative(row.Total)
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
	return findEmpresaSoporteComprasIADuplicadoExcept(dbConn, empresaID, hash, documento, 0)
}

func findEmpresaSoporteComprasIADuplicadoExcept(dbConn *sql.DB, empresaID int64, hash, documento string, excludeID int64) int64 {
	hash = strings.TrimSpace(hash)
	documento = strings.TrimSpace(documento)
	if hash == "" && documento == "" {
		return 0
	}
	where := "empresa_id=? AND COALESCE(estado,'activo')='activo'"
	args := []interface{}{empresaID}
	if excludeID > 0 {
		where += " AND id<>?"
		args = append(args, excludeID)
	}
	if hash != "" && documento != "" {
		where += " AND (archivo_hash=? OR documento_numero=?)"
		args = append(args, hash, documento)
	} else if hash != "" {
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

func normalizeSoporteIAEstadoFiltro(v string) string {
	if strings.TrimSpace(v) == "" {
		return ""
	}
	return normalizeSoporteIAEstado(v)
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
	if v == "eliminado" || v == "purga_pendiente" || v == "purgado" || v == "inactivo" || v == "archivado" {
		return v
	}
	return "activo"
}

func normalizeSoporteIARegistroFiltro(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "eliminado":
		return "eliminado"
	case "purgado":
		return "purgado"
	case "purga_pendiente":
		return "purga_pendiente"
	}
	return "activo"
}

func soporteIARegistroActivo(v string) bool {
	return strings.EqualFold(strings.TrimSpace(v), "activo")
}

func boolToIntSoporteIA(v bool) int {
	if v {
		return 1
	}
	return 0
}

func soporteIAEstadoAbierto(v string) bool {
	switch normalizeSoporteIAEstado(v) {
	case "radicado", "extraido", "en_revision", "aprobado":
		return true
	default:
		return false
	}
}

func soporteIAEstadoExtraible(v string) bool {
	switch normalizeSoporteIAEstado(v) {
	case "radicado", "extraido", "en_revision":
		return true
	default:
		return false
	}
}

func soporteIANonNegative(v float64) float64 {
	if v < 0 {
		return 0
	}
	return v
}

func soporteIARound(v float64) float64 {
	if v < 0.005 && v > -0.005 {
		return 0
	}
	return float64(int64(v*100+0.5)) / 100
}
