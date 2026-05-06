package db

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

type EmpresaAIUContrato struct {
	ID                    int64               `json:"id"`
	EmpresaID             int64               `json:"empresa_id"`
	Codigo                string              `json:"codigo"`
	Nombre                string              `json:"nombre"`
	ClienteID             int64               `json:"cliente_id,omitempty"`
	ClienteNombre         string              `json:"cliente_nombre,omitempty"`
	Responsable           string              `json:"responsable,omitempty"`
	CentroCosto           string              `json:"centro_costo,omitempty"`
	ModalidadContrato     string              `json:"modalidad_contrato,omitempty"`
	TipoObra              string              `json:"tipo_obra"`
	ModeloAIU             string              `json:"modelo_aiu"`
	BaseIVAModo           string              `json:"base_iva_modo"`
	PorcentajeAdmin       float64             `json:"porcentaje_admin"`
	PorcentajeImprevistos float64             `json:"porcentaje_imprevistos"`
	PorcentajeUtilidad    float64             `json:"porcentaje_utilidad"`
	PorcentajeIVA         float64             `json:"porcentaje_iva"`
	PorcentajeRetFuente   float64             `json:"porcentaje_retencion_fuente"`
	PorcentajeRetICA      float64             `json:"porcentaje_retencion_ica"`
	PorcentajeRetIVA      float64             `json:"porcentaje_retencion_iva"`
	PorcentajeAnticipo    float64             `json:"porcentaje_anticipo"`
	PorcentajeGarantia    float64             `json:"porcentaje_garantia"`
	AvancePorcentaje      float64             `json:"avance_porcentaje"`
	FechaInicio           string              `json:"fecha_inicio,omitempty"`
	FechaFin              string              `json:"fecha_fin,omitempty"`
	Estado                string              `json:"estado"`
	RiesgoNivel           string              `json:"riesgo_nivel,omitempty"`
	CostoDirecto          float64             `json:"costo_directo"`
	ValorAdministracion   float64             `json:"valor_administracion"`
	ValorImprevistos      float64             `json:"valor_imprevistos"`
	ValorUtilidad         float64             `json:"valor_utilidad"`
	AIUTotal              float64             `json:"aiu_total"`
	BaseIVA               float64             `json:"base_iva"`
	ValorIVA              float64             `json:"valor_iva"`
	TotalFactura          float64             `json:"total_factura"`
	ValorRetFuente        float64             `json:"valor_retencion_fuente"`
	ValorRetICA           float64             `json:"valor_retencion_ica"`
	ValorRetIVA           float64             `json:"valor_retencion_iva"`
	ValorAnticipo         float64             `json:"valor_anticipo"`
	ValorGarantia         float64             `json:"valor_garantia"`
	NetoCobrar            float64             `json:"neto_cobrar"`
	DocumentoCodigo       string              `json:"documento_codigo,omitempty"`
	AprobadoPor           string              `json:"aprobado_por,omitempty"`
	FechaAprobacion       string              `json:"fecha_aprobacion,omitempty"`
	UsuarioCreador        string              `json:"usuario_creador,omitempty"`
	FechaCreacion         string              `json:"fecha_creacion,omitempty"`
	FechaActualizacion    string              `json:"fecha_actualizacion,omitempty"`
	Observaciones         string              `json:"observaciones,omitempty"`
	Items                 []EmpresaAIUItem    `json:"items,omitempty"`
	Facturas              []EmpresaAIUFactura `json:"facturas,omitempty"`
}

type EmpresaAIUItem struct {
	ID            int64   `json:"id"`
	EmpresaID     int64   `json:"empresa_id"`
	ContratoID    int64   `json:"contrato_id"`
	Capitulo      string  `json:"capitulo"`
	Descripcion   string  `json:"descripcion"`
	Unidad        string  `json:"unidad"`
	Cantidad      float64 `json:"cantidad"`
	ValorUnitario float64 `json:"valor_unitario"`
	ValorTotal    float64 `json:"valor_total"`
	Estado        string  `json:"estado"`
	FechaCreacion string  `json:"fecha_creacion,omitempty"`
}

type EmpresaAIUFactura struct {
	ID               int64   `json:"id"`
	EmpresaID        int64   `json:"empresa_id"`
	ContratoID       int64   `json:"contrato_id"`
	DocumentoCodigo  string  `json:"documento_codigo"`
	TipoDocumento    string  `json:"tipo_documento"`
	PeriodoContable  string  `json:"periodo_contable,omitempty"`
	Estado           string  `json:"estado"`
	CostoDirecto     float64 `json:"costo_directo"`
	AIUTotal         float64 `json:"aiu_total"`
	BaseIVA          float64 `json:"base_iva"`
	ValorIVA         float64 `json:"valor_iva"`
	TotalFactura     float64 `json:"total_factura"`
	ValorRetenciones float64 `json:"valor_retenciones"`
	ValorAnticipo    float64 `json:"valor_anticipo"`
	ValorGarantia    float64 `json:"valor_garantia"`
	NetoCobrar       float64 `json:"neto_cobrar"`
	FechaDocumento   string  `json:"fecha_documento,omitempty"`
	UsuarioCreador   string  `json:"usuario_creador,omitempty"`
	FechaCreacion    string  `json:"fecha_creacion,omitempty"`
	Observaciones    string  `json:"observaciones,omitempty"`
}

type EmpresaAIUDashboard struct {
	EmpresaID          int64                `json:"empresa_id"`
	ContratosActivos   int                  `json:"contratos_activos"`
	ContratosCerrados  int                  `json:"contratos_cerrados"`
	CostoDirectoTotal  float64              `json:"costo_directo_total"`
	AIUTotal           float64              `json:"aiu_total"`
	BaseIVATotal       float64              `json:"base_iva_total"`
	ValorIVATotal      float64              `json:"valor_iva_total"`
	TotalFacturado     float64              `json:"total_facturado"`
	RetencionesTotal   float64              `json:"retenciones_total"`
	NetoCobrarTotal    float64              `json:"neto_cobrar_total"`
	PendienteFacturar  float64              `json:"pendiente_facturar"`
	ContratosPorEstado map[string]int       `json:"contratos_por_estado"`
	Alertas            []string             `json:"alertas"`
	UltimosContratos   []EmpresaAIUContrato `json:"ultimos_contratos"`
	UltimasFacturas    []EmpresaAIUFactura  `json:"ultimas_facturas"`
}

func EnsureEmpresaAIUConstruccionSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_aiu_contratos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			cliente_id INTEGER DEFAULT 0,
			cliente_nombre TEXT,
			responsable TEXT,
			centro_costo TEXT,
			modalidad_contrato TEXT DEFAULT 'precio_global',
			tipo_obra TEXT DEFAULT 'obra_civil',
			modelo_aiu TEXT DEFAULT 'base_aiu_no_sumada',
			base_iva_modo TEXT DEFAULT 'utilidad',
			porcentaje_admin REAL DEFAULT 10,
			porcentaje_imprevistos REAL DEFAULT 5,
			porcentaje_utilidad REAL DEFAULT 10,
			porcentaje_iva REAL DEFAULT 19,
			porcentaje_retencion_fuente REAL DEFAULT 0,
			porcentaje_retencion_ica REAL DEFAULT 0,
			porcentaje_retencion_iva REAL DEFAULT 0,
			porcentaje_anticipo REAL DEFAULT 0,
			porcentaje_garantia REAL DEFAULT 0,
			avance_porcentaje REAL DEFAULT 0,
			fecha_inicio TEXT,
			fecha_fin TEXT,
			estado TEXT DEFAULT 'borrador',
			riesgo_nivel TEXT DEFAULT 'medio',
			costo_directo REAL DEFAULT 0,
			valor_administracion REAL DEFAULT 0,
			valor_imprevistos REAL DEFAULT 0,
			valor_utilidad REAL DEFAULT 0,
			aiu_total REAL DEFAULT 0,
			base_iva REAL DEFAULT 0,
			valor_iva REAL DEFAULT 0,
			total_factura REAL DEFAULT 0,
			valor_retencion_fuente REAL DEFAULT 0,
			valor_retencion_ica REAL DEFAULT 0,
			valor_retencion_iva REAL DEFAULT 0,
			valor_anticipo REAL DEFAULT 0,
			valor_garantia REAL DEFAULT 0,
			neto_cobrar REAL DEFAULT 0,
			documento_codigo TEXT,
			aprobado_por TEXT,
			fecha_aprobacion TEXT,
			usuario_creador TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_aiu_contratos_empresa_estado ON empresa_aiu_contratos(empresa_id, estado)`,
		`CREATE TABLE IF NOT EXISTS empresa_aiu_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			contrato_id INTEGER NOT NULL,
			capitulo TEXT DEFAULT '',
			descripcion TEXT NOT NULL,
			unidad TEXT DEFAULT 'und',
			cantidad REAL DEFAULT 0,
			valor_unitario REAL DEFAULT 0,
			valor_total REAL DEFAULT 0,
			estado TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS ix_aiu_items_contrato ON empresa_aiu_items(empresa_id, contrato_id)`,
		`CREATE TABLE IF NOT EXISTS empresa_aiu_facturas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			contrato_id INTEGER NOT NULL,
			documento_codigo TEXT NOT NULL,
			tipo_documento TEXT DEFAULT 'factura_electronica',
			periodo_contable TEXT,
			estado TEXT DEFAULT 'emitida',
			costo_directo REAL DEFAULT 0,
			aiu_total REAL DEFAULT 0,
			base_iva REAL DEFAULT 0,
			valor_iva REAL DEFAULT 0,
			total_factura REAL DEFAULT 0,
			valor_retenciones REAL DEFAULT 0,
			valor_anticipo REAL DEFAULT 0,
			valor_garantia REAL DEFAULT 0,
			neto_cobrar REAL DEFAULT 0,
			fecha_documento TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			observaciones TEXT,
			UNIQUE(empresa_id, documento_codigo)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_aiu_facturas_contrato ON empresa_aiu_facturas(empresa_id, contrato_id)`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	contractColumns := map[string]string{
		"responsable":                 "TEXT",
		"centro_costo":                "TEXT",
		"modalidad_contrato":          "TEXT DEFAULT 'precio_global'",
		"porcentaje_retencion_fuente": "REAL DEFAULT 0",
		"porcentaje_retencion_ica":    "REAL DEFAULT 0",
		"porcentaje_retencion_iva":    "REAL DEFAULT 0",
		"porcentaje_anticipo":         "REAL DEFAULT 0",
		"porcentaje_garantia":         "REAL DEFAULT 0",
		"avance_porcentaje":           "REAL DEFAULT 0",
		"riesgo_nivel":                "TEXT DEFAULT 'medio'",
		"valor_retencion_fuente":      "REAL DEFAULT 0",
		"valor_retencion_ica":         "REAL DEFAULT 0",
		"valor_retencion_iva":         "REAL DEFAULT 0",
		"valor_anticipo":              "REAL DEFAULT 0",
		"valor_garantia":              "REAL DEFAULT 0",
		"neto_cobrar":                 "REAL DEFAULT 0",
		"aprobado_por":                "TEXT",
		"fecha_aprobacion":            "TEXT",
	}
	for column, def := range contractColumns {
		if err := ensureColumnIfMissing(dbConn, "empresa_aiu_contratos", column, def); err != nil {
			return err
		}
	}
	facturaColumns := map[string]string{
		"valor_retenciones": "REAL DEFAULT 0",
		"valor_anticipo":    "REAL DEFAULT 0",
		"valor_garantia":    "REAL DEFAULT 0",
		"neto_cobrar":       "REAL DEFAULT 0",
	}
	for column, def := range facturaColumns {
		if err := ensureColumnIfMissing(dbConn, "empresa_aiu_facturas", column, def); err != nil {
			return err
		}
	}
	return nil
}

func UpsertEmpresaAIUContrato(dbConn *sql.DB, item EmpresaAIUContrato) (int64, error) {
	if err := EnsureEmpresaAIUConstruccionSchema(dbConn); err != nil {
		return 0, err
	}
	item = NormalizeEmpresaAIUContrato(item)
	if item.EmpresaID <= 0 || item.Codigo == "" || item.Nombre == "" {
		return 0, errors.New("empresa_id, codigo y nombre son obligatorios")
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_aiu_contratos
		(empresa_id,codigo,nombre,cliente_id,cliente_nombre,responsable,centro_costo,modalidad_contrato,tipo_obra,modelo_aiu,base_iva_modo,porcentaje_admin,porcentaje_imprevistos,porcentaje_utilidad,porcentaje_iva,porcentaje_retencion_fuente,porcentaje_retencion_ica,porcentaje_retencion_iva,porcentaje_anticipo,porcentaje_garantia,avance_porcentaje,fecha_inicio,fecha_fin,estado,riesgo_nivel,observaciones,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (empresa_id,codigo) DO UPDATE SET
			nombre=EXCLUDED.nombre, cliente_id=EXCLUDED.cliente_id, cliente_nombre=EXCLUDED.cliente_nombre,
			responsable=EXCLUDED.responsable, centro_costo=EXCLUDED.centro_costo, modalidad_contrato=EXCLUDED.modalidad_contrato,
			tipo_obra=EXCLUDED.tipo_obra, modelo_aiu=EXCLUDED.modelo_aiu, base_iva_modo=EXCLUDED.base_iva_modo,
			porcentaje_admin=EXCLUDED.porcentaje_admin, porcentaje_imprevistos=EXCLUDED.porcentaje_imprevistos,
			porcentaje_utilidad=EXCLUDED.porcentaje_utilidad, porcentaje_iva=EXCLUDED.porcentaje_iva,
			porcentaje_retencion_fuente=EXCLUDED.porcentaje_retencion_fuente, porcentaje_retencion_ica=EXCLUDED.porcentaje_retencion_ica,
			porcentaje_retencion_iva=EXCLUDED.porcentaje_retencion_iva, porcentaje_anticipo=EXCLUDED.porcentaje_anticipo,
			porcentaje_garantia=EXCLUDED.porcentaje_garantia, avance_porcentaje=EXCLUDED.avance_porcentaje,
			fecha_inicio=EXCLUDED.fecha_inicio, fecha_fin=EXCLUDED.fecha_fin, estado=EXCLUDED.estado, riesgo_nivel=EXCLUDED.riesgo_nivel,
			observaciones=EXCLUDED.observaciones, usuario_creador=EXCLUDED.usuario_creador, fecha_actualizacion=CURRENT_TIMESTAMP`,
		item.EmpresaID, item.Codigo, item.Nombre, item.ClienteID, item.ClienteNombre, item.Responsable, item.CentroCosto, item.ModalidadContrato, item.TipoObra, item.ModeloAIU, item.BaseIVAModo,
		item.PorcentajeAdmin, item.PorcentajeImprevistos, item.PorcentajeUtilidad, item.PorcentajeIVA, item.PorcentajeRetFuente, item.PorcentajeRetICA, item.PorcentajeRetIVA,
		item.PorcentajeAnticipo, item.PorcentajeGarantia, item.AvancePorcentaje, item.FechaInicio, item.FechaFin, item.Estado, item.RiesgoNivel, item.Observaciones, item.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	row, err := GetEmpresaAIUContratoByCodigo(dbConn, item.EmpresaID, item.Codigo)
	if err != nil {
		return id, err
	}
	return row.ID, RecalcularEmpresaAIUContrato(dbConn, item.EmpresaID, row.ID)
}

func CreateEmpresaAIUItem(dbConn *sql.DB, item EmpresaAIUItem) (int64, error) {
	if err := EnsureEmpresaAIUConstruccionSchema(dbConn); err != nil {
		return 0, err
	}
	item = normalizeEmpresaAIUItem(item)
	if item.EmpresaID <= 0 || item.ContratoID <= 0 || item.Descripcion == "" {
		return 0, errors.New("contrato y descripcion son obligatorios")
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_aiu_items
		(empresa_id,contrato_id,capitulo,descripcion,unidad,cantidad,valor_unitario,valor_total,estado)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		item.EmpresaID, item.ContratoID, item.Capitulo, item.Descripcion, item.Unidad, item.Cantidad, item.ValorUnitario, item.ValorTotal, item.Estado)
	if err != nil {
		return 0, err
	}
	return id, RecalcularEmpresaAIUContrato(dbConn, item.EmpresaID, item.ContratoID)
}

func RecalcularEmpresaAIUContrato(dbConn *sql.DB, empresaID, contratoID int64) error {
	row, err := GetEmpresaAIUContrato(dbConn, empresaID, contratoID)
	if err != nil {
		return err
	}
	total := 0.0
	for _, item := range row.Items {
		if strings.EqualFold(item.Estado, "activo") {
			total += item.ValorTotal
		}
	}
	if total > 0 {
		row.CostoDirecto = total
	}
	row = CalculateEmpresaAIUContrato(row)
	_, err = ExecCompat(dbConn, `UPDATE empresa_aiu_contratos SET
		costo_directo=?, valor_administracion=?, valor_imprevistos=?, valor_utilidad=?, aiu_total=?, base_iva=?, valor_iva=?, total_factura=?,
		valor_retencion_fuente=?, valor_retencion_ica=?, valor_retencion_iva=?, valor_anticipo=?, valor_garantia=?, neto_cobrar=?, fecha_actualizacion=CURRENT_TIMESTAMP
		WHERE empresa_id=? AND id=?`,
		row.CostoDirecto, row.ValorAdministracion, row.ValorImprevistos, row.ValorUtilidad, row.AIUTotal, row.BaseIVA, row.ValorIVA, row.TotalFactura,
		row.ValorRetFuente, row.ValorRetICA, row.ValorRetIVA, row.ValorAnticipo, row.ValorGarantia, row.NetoCobrar, empresaID, contratoID)
	return err
}

func GetEmpresaAIUContratoByCodigo(dbConn *sql.DB, empresaID int64, codigo string) (EmpresaAIUContrato, error) {
	rows, err := ListEmpresaAIUContratos(dbConn, empresaID, "", 500)
	if err != nil {
		return EmpresaAIUContrato{}, err
	}
	codigo = strings.ToUpper(strings.TrimSpace(codigo))
	for _, row := range rows {
		if row.Codigo == codigo {
			row.Items, _ = ListEmpresaAIUItems(dbConn, empresaID, row.ID)
			row.Facturas, _ = ListEmpresaAIUFacturas(dbConn, empresaID, row.ID, 100)
			return row, nil
		}
	}
	return EmpresaAIUContrato{}, sql.ErrNoRows
}

func GetEmpresaAIUContrato(dbConn *sql.DB, empresaID, id int64) (EmpresaAIUContrato, error) {
	rows, err := ListEmpresaAIUContratos(dbConn, empresaID, "", 500)
	if err != nil {
		return EmpresaAIUContrato{}, err
	}
	for _, row := range rows {
		if row.ID == id {
			row.Items, _ = ListEmpresaAIUItems(dbConn, empresaID, id)
			row.Facturas, _ = ListEmpresaAIUFacturas(dbConn, empresaID, id, 100)
			return row, nil
		}
	}
	return EmpresaAIUContrato{}, sql.ErrNoRows
}

func ListEmpresaAIUContratos(dbConn *sql.DB, empresaID int64, estado string, limit int) ([]EmpresaAIUContrato, error) {
	return ListEmpresaAIUContratosFiltrados(dbConn, empresaID, EmpresaAIUContratoFiltro{Estado: estado, Limit: limit})
}

type EmpresaAIUContratoFiltro struct {
	Estado string
	Query  string
	Limit  int
}

func ListEmpresaAIUContratosFiltrados(dbConn *sql.DB, empresaID int64, filtro EmpresaAIUContratoFiltro) ([]EmpresaAIUContrato, error) {
	if err := EnsureEmpresaAIUConstruccionSchema(dbConn); err != nil {
		return nil, err
	}
	limit := filtro.Limit
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if strings.TrimSpace(filtro.Estado) != "" {
		where += " AND estado=?"
		args = append(args, normalizeEmpresaAIUEstado(filtro.Estado))
	}
	query := strings.TrimSpace(filtro.Query)
	if query != "" {
		where += " AND (UPPER(codigo) LIKE ? OR UPPER(nombre) LIKE ? OR UPPER(COALESCE(cliente_nombre,'')) LIKE ? OR UPPER(COALESCE(centro_costo,'')) LIKE ?)"
		needle := "%" + strings.ToUpper(query) + "%"
		args = append(args, needle, needle, needle, needle)
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,codigo,nombre,COALESCE(cliente_id,0),COALESCE(cliente_nombre,''),COALESCE(responsable,''),COALESCE(centro_costo,''),COALESCE(modalidad_contrato,'precio_global'),COALESCE(tipo_obra,'obra_civil'),COALESCE(modelo_aiu,'base_aiu_no_sumada'),COALESCE(base_iva_modo,'utilidad'),COALESCE(porcentaje_admin,0),COALESCE(porcentaje_imprevistos,0),COALESCE(porcentaje_utilidad,0),COALESCE(porcentaje_iva,19),COALESCE(porcentaje_retencion_fuente,0),COALESCE(porcentaje_retencion_ica,0),COALESCE(porcentaje_retencion_iva,0),COALESCE(porcentaje_anticipo,0),COALESCE(porcentaje_garantia,0),COALESCE(avance_porcentaje,0),COALESCE(fecha_inicio,''),COALESCE(fecha_fin,''),COALESCE(estado,'borrador'),COALESCE(riesgo_nivel,'medio'),COALESCE(costo_directo,0),COALESCE(valor_administracion,0),COALESCE(valor_imprevistos,0),COALESCE(valor_utilidad,0),COALESCE(aiu_total,0),COALESCE(base_iva,0),COALESCE(valor_iva,0),COALESCE(total_factura,0),COALESCE(valor_retencion_fuente,0),COALESCE(valor_retencion_ica,0),COALESCE(valor_retencion_iva,0),COALESCE(valor_anticipo,0),COALESCE(valor_garantia,0),COALESCE(neto_cobrar,0),COALESCE(documento_codigo,''),COALESCE(aprobado_por,''),COALESCE(fecha_aprobacion,''),COALESCE(usuario_creador,''),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,''),COALESCE(observaciones,'') FROM empresa_aiu_contratos WHERE %s ORDER BY id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaAIUContrato{}
	for rows.Next() {
		var x EmpresaAIUContrato
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Codigo, &x.Nombre, &x.ClienteID, &x.ClienteNombre, &x.Responsable, &x.CentroCosto, &x.ModalidadContrato, &x.TipoObra, &x.ModeloAIU, &x.BaseIVAModo, &x.PorcentajeAdmin, &x.PorcentajeImprevistos, &x.PorcentajeUtilidad, &x.PorcentajeIVA, &x.PorcentajeRetFuente, &x.PorcentajeRetICA, &x.PorcentajeRetIVA, &x.PorcentajeAnticipo, &x.PorcentajeGarantia, &x.AvancePorcentaje, &x.FechaInicio, &x.FechaFin, &x.Estado, &x.RiesgoNivel, &x.CostoDirecto, &x.ValorAdministracion, &x.ValorImprevistos, &x.ValorUtilidad, &x.AIUTotal, &x.BaseIVA, &x.ValorIVA, &x.TotalFactura, &x.ValorRetFuente, &x.ValorRetICA, &x.ValorRetIVA, &x.ValorAnticipo, &x.ValorGarantia, &x.NetoCobrar, &x.DocumentoCodigo, &x.AprobadoPor, &x.FechaAprobacion, &x.UsuarioCreador, &x.FechaCreacion, &x.FechaActualizacion, &x.Observaciones); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func ListEmpresaAIUItems(dbConn *sql.DB, empresaID, contratoID int64) ([]EmpresaAIUItem, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,contrato_id,COALESCE(capitulo,''),COALESCE(descripcion,''),COALESCE(unidad,'und'),COALESCE(cantidad,0),COALESCE(valor_unitario,0),COALESCE(valor_total,0),COALESCE(estado,'activo'),COALESCE(fecha_creacion,'') FROM empresa_aiu_items WHERE empresa_id=? AND contrato_id=? ORDER BY id`, empresaID, contratoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaAIUItem{}
	for rows.Next() {
		var x EmpresaAIUItem
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.ContratoID, &x.Capitulo, &x.Descripcion, &x.Unidad, &x.Cantidad, &x.ValorUnitario, &x.ValorTotal, &x.Estado, &x.FechaCreacion); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func ListEmpresaAIUFacturas(dbConn *sql.DB, empresaID, contratoID int64, limit int) ([]EmpresaAIUFactura, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if contratoID > 0 {
		where += " AND contrato_id=?"
		args = append(args, contratoID)
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,contrato_id,documento_codigo,COALESCE(tipo_documento,'factura_electronica'),COALESCE(periodo_contable,''),COALESCE(estado,'emitida'),COALESCE(costo_directo,0),COALESCE(aiu_total,0),COALESCE(base_iva,0),COALESCE(valor_iva,0),COALESCE(total_factura,0),COALESCE(valor_retenciones,0),COALESCE(valor_anticipo,0),COALESCE(valor_garantia,0),COALESCE(neto_cobrar,0),COALESCE(fecha_documento,''),COALESCE(usuario_creador,''),COALESCE(fecha_creacion,''),COALESCE(observaciones,'') FROM empresa_aiu_facturas WHERE %s ORDER BY id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaAIUFactura{}
	for rows.Next() {
		var x EmpresaAIUFactura
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.ContratoID, &x.DocumentoCodigo, &x.TipoDocumento, &x.PeriodoContable, &x.Estado, &x.CostoDirecto, &x.AIUTotal, &x.BaseIVA, &x.ValorIVA, &x.TotalFactura, &x.ValorRetenciones, &x.ValorAnticipo, &x.ValorGarantia, &x.NetoCobrar, &x.FechaDocumento, &x.UsuarioCreador, &x.FechaCreacion, &x.Observaciones); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func RegistrarEmpresaAIUFactura(dbConn *sql.DB, empresaID, contratoID int64, documentoCodigo, periodo, usuario string) (EmpresaAIUFactura, error) {
	row, err := GetEmpresaAIUContrato(dbConn, empresaID, contratoID)
	if err != nil {
		return EmpresaAIUFactura{}, err
	}
	row = CalculateEmpresaAIUContrato(row)
	documentoCodigo = strings.ToUpper(strings.TrimSpace(documentoCodigo))
	if documentoCodigo == "" {
		documentoCodigo = fmt.Sprintf("AIU-%s-%d", row.Codigo, time.Now().Unix())
	}
	periodo = strings.TrimSpace(periodo)
	if periodo == "" {
		periodo = time.Now().Format("2006-01")
	}
	retenciones := aiuRoundMoney(row.ValorRetFuente + row.ValorRetICA + row.ValorRetIVA)
	obs := fmt.Sprintf("Factura AIU contrato %s: costo directo %.2f, A %.2f, I %.2f, U %.2f, base IVA %.2f, IVA %.2f, retenciones %.2f, anticipo %.2f, garantia %.2f, neto %.2f. Modelo %s.", row.Codigo, row.CostoDirecto, row.ValorAdministracion, row.ValorImprevistos, row.ValorUtilidad, row.BaseIVA, row.ValorIVA, retenciones, row.ValorAnticipo, row.ValorGarantia, row.NetoCobrar, row.ModeloAIU)
	_, err = insertSQLCompat(dbConn, `INSERT INTO empresa_aiu_facturas
		(empresa_id,contrato_id,documento_codigo,tipo_documento,periodo_contable,estado,costo_directo,aiu_total,base_iva,valor_iva,total_factura,valor_retenciones,valor_anticipo,valor_garantia,neto_cobrar,usuario_creador,observaciones)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (empresa_id,documento_codigo) DO UPDATE SET
			contrato_id=EXCLUDED.contrato_id, periodo_contable=EXCLUDED.periodo_contable, estado=EXCLUDED.estado,
			costo_directo=EXCLUDED.costo_directo, aiu_total=EXCLUDED.aiu_total, base_iva=EXCLUDED.base_iva,
			valor_iva=EXCLUDED.valor_iva, total_factura=EXCLUDED.total_factura, valor_retenciones=EXCLUDED.valor_retenciones,
			valor_anticipo=EXCLUDED.valor_anticipo, valor_garantia=EXCLUDED.valor_garantia, neto_cobrar=EXCLUDED.neto_cobrar, usuario_creador=EXCLUDED.usuario_creador,
			observaciones=EXCLUDED.observaciones`,
		empresaID, contratoID, documentoCodigo, "factura_electronica", periodo, "emitida", row.CostoDirecto, row.AIUTotal, row.BaseIVA, row.ValorIVA, row.TotalFactura, retenciones, row.ValorAnticipo, row.ValorGarantia, row.NetoCobrar, usuario, obs)
	if err != nil {
		return EmpresaAIUFactura{}, err
	}
	_, _ = ExecCompat(dbConn, `UPDATE empresa_aiu_contratos SET documento_codigo=?, estado='facturado', fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, documentoCodigo, empresaID, contratoID)
	facturas, err := ListEmpresaAIUFacturas(dbConn, empresaID, contratoID, 1)
	if err != nil || len(facturas) == 0 {
		return EmpresaAIUFactura{}, err
	}
	return facturas[0], nil
}

func BuildEmpresaAIUDashboard(dbConn *sql.DB, empresaID int64) (EmpresaAIUDashboard, error) {
	contratos, err := ListEmpresaAIUContratos(dbConn, empresaID, "", 500)
	if err != nil {
		return EmpresaAIUDashboard{}, err
	}
	facturas, err := ListEmpresaAIUFacturas(dbConn, empresaID, 0, 500)
	if err != nil {
		return EmpresaAIUDashboard{}, err
	}
	ultimosContratos := contratos
	if len(ultimosContratos) > 20 {
		ultimosContratos = ultimosContratos[:20]
	}
	ultimasFacturas := facturas
	if len(ultimasFacturas) > 20 {
		ultimasFacturas = ultimasFacturas[:20]
	}
	out := EmpresaAIUDashboard{
		EmpresaID:          empresaID,
		ContratosPorEstado: map[string]int{},
		UltimosContratos:   ultimosContratos,
		UltimasFacturas:    ultimasFacturas,
	}
	for _, c := range contratos {
		out.ContratosPorEstado[c.Estado] += 1
		if c.Estado == "cerrado" || c.Estado == "facturado" {
			out.ContratosCerrados += 1
		} else {
			out.ContratosActivos += 1
		}
		out.CostoDirectoTotal += c.CostoDirecto
		out.AIUTotal += c.AIUTotal
		out.BaseIVATotal += c.BaseIVA
		out.ValorIVATotal += c.ValorIVA
		if c.Estado != "facturado" && c.Estado != "cerrado" && c.Estado != "anulado" {
			out.PendienteFacturar += c.TotalFactura
		}
		if c.AvancePorcentaje >= 90 && c.Estado != "facturado" && c.Estado != "cerrado" && c.Estado != "anulado" {
			out.Alertas = append(out.Alertas, fmt.Sprintf("%s tiene %.0f%% de avance y aun no esta facturado.", c.Codigo, c.AvancePorcentaje))
		}
		if c.TotalFactura > 0 && c.NetoCobrar < 0 {
			out.Alertas = append(out.Alertas, fmt.Sprintf("%s tiene neto a cobrar negativo por retenciones/anticipos.", c.Codigo))
		}
	}
	for _, f := range facturas {
		out.TotalFacturado += f.TotalFactura
		out.RetencionesTotal += f.ValorRetenciones
		out.NetoCobrarTotal += f.NetoCobrar
	}
	return out, nil
}

func SeedEmpresaAIUDemo(dbConn *sql.DB, empresaID int64, usuario string) error {
	id, err := UpsertEmpresaAIUContrato(dbConn, EmpresaAIUContrato{
		EmpresaID:             empresaID,
		Codigo:                "AIU-DEMO-001",
		Nombre:                "Adecuacion oficina administrativa",
		ClienteNombre:         "Cliente obra demo",
		Responsable:           "Director de obra demo",
		CentroCosto:           "OBRAS-DEMO",
		ModalidadContrato:     "precio_global",
		TipoObra:              "remodelacion",
		ModeloAIU:             "base_aiu_sumada",
		BaseIVAModo:           "utilidad",
		PorcentajeAdmin:       10,
		PorcentajeImprevistos: 5,
		PorcentajeUtilidad:    12,
		PorcentajeIVA:         19,
		PorcentajeRetFuente:   2,
		PorcentajeRetICA:      0.966,
		PorcentajeGarantia:    5,
		AvancePorcentaje:      35,
		Estado:                "en_ejecucion",
		RiesgoNivel:           "medio",
		UsuarioCreador:        usuario,
		Observaciones:         "Contrato demo AIU para arquitectos, constructores y pequenas empresas de obra.",
	})
	if err != nil {
		return err
	}
	if _, err := CreateEmpresaAIUItem(dbConn, EmpresaAIUItem{EmpresaID: empresaID, ContratoID: id, Capitulo: "Preliminares", Descripcion: "Demoliciones y retiro de escombros", Unidad: "global", Cantidad: 1, ValorUnitario: 3500000}); err != nil {
		return err
	}
	if _, err := CreateEmpresaAIUItem(dbConn, EmpresaAIUItem{EmpresaID: empresaID, ContratoID: id, Capitulo: "Acabados", Descripcion: "Piso, pintura e instalaciones menores", Unidad: "m2", Cantidad: 42, ValorUnitario: 185000}); err != nil {
		return err
	}
	return nil
}

func CalculateEmpresaAIUContrato(x EmpresaAIUContrato) EmpresaAIUContrato {
	x = NormalizeEmpresaAIUContrato(x)
	x.ValorAdministracion = aiuRoundMoney(x.CostoDirecto * x.PorcentajeAdmin / 100)
	x.ValorImprevistos = aiuRoundMoney(x.CostoDirecto * x.PorcentajeImprevistos / 100)
	x.ValorUtilidad = aiuRoundMoney(x.CostoDirecto * x.PorcentajeUtilidad / 100)
	x.AIUTotal = aiuRoundMoney(x.ValorAdministracion + x.ValorImprevistos + x.ValorUtilidad)
	switch x.BaseIVAModo {
	case "aiu_total":
		x.BaseIVA = x.AIUTotal
	case "costo_mas_aiu":
		x.BaseIVA = aiuRoundMoney(x.CostoDirecto + x.AIUTotal)
	default:
		x.BaseIVA = x.ValorUtilidad
	}
	x.ValorIVA = aiuRoundMoney(x.BaseIVA * x.PorcentajeIVA / 100)
	if x.ModeloAIU == "base_aiu_sumada" {
		x.TotalFactura = aiuRoundMoney(x.CostoDirecto + x.AIUTotal + x.ValorIVA)
	} else {
		x.TotalFactura = aiuRoundMoney(x.CostoDirecto + x.ValorIVA)
	}
	x.ValorRetFuente = aiuRoundMoney(x.BaseIVA * x.PorcentajeRetFuente / 100)
	x.ValorRetICA = aiuRoundMoney(x.TotalFactura * x.PorcentajeRetICA / 100)
	x.ValorRetIVA = aiuRoundMoney(x.ValorIVA * x.PorcentajeRetIVA / 100)
	x.ValorAnticipo = aiuRoundMoney(x.TotalFactura * x.PorcentajeAnticipo / 100)
	x.ValorGarantia = aiuRoundMoney(x.TotalFactura * x.PorcentajeGarantia / 100)
	x.NetoCobrar = aiuRoundMoney(x.TotalFactura - x.ValorRetFuente - x.ValorRetICA - x.ValorRetIVA - x.ValorAnticipo - x.ValorGarantia)
	return x
}

func NormalizeEmpresaAIUContrato(x EmpresaAIUContrato) EmpresaAIUContrato {
	x.Codigo = strings.ToUpper(strings.TrimSpace(x.Codigo))
	x.Nombre = strings.TrimSpace(x.Nombre)
	x.ClienteNombre = strings.TrimSpace(x.ClienteNombre)
	x.Responsable = strings.TrimSpace(x.Responsable)
	x.CentroCosto = strings.ToUpper(strings.TrimSpace(x.CentroCosto))
	x.ModalidadContrato = normalizeAIUModalidadContrato(x.ModalidadContrato)
	x.TipoObra = normalizeAIUText(x.TipoObra, "obra_civil")
	x.ModeloAIU = normalizeAIUModelo(x.ModeloAIU)
	x.BaseIVAModo = normalizeAIUBaseIVA(x.BaseIVAModo)
	x.PorcentajeAdmin = aiuClampPercent(x.PorcentajeAdmin, 10)
	x.PorcentajeImprevistos = aiuClampPercent(x.PorcentajeImprevistos, 5)
	x.PorcentajeUtilidad = aiuClampPercent(x.PorcentajeUtilidad, 10)
	x.PorcentajeIVA = aiuClampPercent(x.PorcentajeIVA, 19)
	x.PorcentajeRetFuente = aiuClampPercentAllowZero(x.PorcentajeRetFuente)
	x.PorcentajeRetICA = aiuClampPercentAllowZero(x.PorcentajeRetICA)
	x.PorcentajeRetIVA = aiuClampPercentAllowZero(x.PorcentajeRetIVA)
	x.PorcentajeAnticipo = aiuClampPercentAllowZero(x.PorcentajeAnticipo)
	x.PorcentajeGarantia = aiuClampPercentAllowZero(x.PorcentajeGarantia)
	x.AvancePorcentaje = aiuClampPercentAllowZero(x.AvancePorcentaje)
	x.Estado = normalizeEmpresaAIUEstado(x.Estado)
	x.RiesgoNivel = normalizeAIURiesgo(x.RiesgoNivel)
	x.CostoDirecto = aiuRoundMoney(math.Max(0, x.CostoDirecto))
	x.Observaciones = strings.TrimSpace(x.Observaciones)
	return x
}

func UpdateEmpresaAIUContratoEstado(dbConn *sql.DB, empresaID, contratoID int64, estado, usuario, observacion string) (EmpresaAIUContrato, error) {
	if err := EnsureEmpresaAIUConstruccionSchema(dbConn); err != nil {
		return EmpresaAIUContrato{}, err
	}
	row, err := GetEmpresaAIUContrato(dbConn, empresaID, contratoID)
	if err != nil {
		return EmpresaAIUContrato{}, err
	}
	next := normalizeEmpresaAIUEstado(estado)
	if !aiuEstadoTransitionAllowed(row.Estado, next) {
		return EmpresaAIUContrato{}, fmt.Errorf("transicion AIU no permitida: %s -> %s", row.Estado, next)
	}
	observacion = strings.TrimSpace(observacion)
	if observacion != "" {
		if row.Observaciones != "" {
			row.Observaciones += "\n"
		}
		row.Observaciones += fmt.Sprintf("[%s] %s", time.Now().Format("2006-01-02 15:04"), observacion)
	}
	aprobadoPor := row.AprobadoPor
	fechaAprobacion := row.FechaAprobacion
	if next == "aprobado" && aprobadoPor == "" {
		aprobadoPor = strings.TrimSpace(usuario)
		fechaAprobacion = time.Now().Format(time.RFC3339)
	}
	if _, err := ExecCompat(dbConn, `UPDATE empresa_aiu_contratos SET estado=?, aprobado_por=?, fecha_aprobacion=?, observaciones=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`,
		next, aprobadoPor, fechaAprobacion, row.Observaciones, empresaID, contratoID); err != nil {
		return EmpresaAIUContrato{}, err
	}
	return GetEmpresaAIUContrato(dbConn, empresaID, contratoID)
}

func normalizeEmpresaAIUItem(x EmpresaAIUItem) EmpresaAIUItem {
	x.Capitulo = strings.TrimSpace(x.Capitulo)
	x.Descripcion = strings.TrimSpace(x.Descripcion)
	x.Unidad = normalizeAIUText(x.Unidad, "und")
	x.Cantidad = math.Max(0, x.Cantidad)
	x.ValorUnitario = aiuRoundMoney(math.Max(0, x.ValorUnitario))
	x.ValorTotal = aiuRoundMoney(x.Cantidad * x.ValorUnitario)
	if x.ValorTotal == 0 {
		x.ValorTotal = aiuRoundMoney(math.Max(0, x.ValorTotal))
	}
	x.Estado = normalizeAIUText(x.Estado, "activo")
	if x.Estado != "activo" && x.Estado != "inactivo" {
		x.Estado = "activo"
	}
	return x
}

func normalizeAIUModelo(v string) string {
	v = normalizeAIUText(v, "base_aiu_no_sumada")
	switch v {
	case "modelo_2", "sumada", "base_aiu_sumada", "aiu_sumada_total":
		return "base_aiu_sumada"
	default:
		return "base_aiu_no_sumada"
	}
}

func normalizeAIUBaseIVA(v string) string {
	v = normalizeAIUText(v, "utilidad")
	switch v {
	case "aiu", "aiu_total", "administracion_imprevistos_utilidad":
		return "aiu_total"
	case "total", "costo_mas_aiu", "contrato_total":
		return "costo_mas_aiu"
	default:
		return "utilidad"
	}
}

func normalizeEmpresaAIUEstado(v string) string {
	v = normalizeAIUText(v, "borrador")
	switch v {
	case "borrador", "cotizado", "aprobado", "en_ejecucion", "suspendido", "facturado", "cerrado", "anulado":
		return v
	default:
		return "borrador"
	}
}

func normalizeAIUModalidadContrato(v string) string {
	v = normalizeAIUText(v, "precio_global")
	switch v {
	case "precio_global", "administracion_delegada", "costos_reembolsables", "obra_por_unidad", "mantenimiento":
		return v
	default:
		return "precio_global"
	}
}

func normalizeAIURiesgo(v string) string {
	v = normalizeAIUText(v, "medio")
	switch v {
	case "bajo", "medio", "alto", "critico":
		return v
	default:
		return "medio"
	}
}

func aiuEstadoTransitionAllowed(current, next string) bool {
	current = normalizeEmpresaAIUEstado(current)
	next = normalizeEmpresaAIUEstado(next)
	if current == next {
		return true
	}
	if next == "anulado" {
		return current != "cerrado" && current != "facturado"
	}
	allowed := map[string][]string{
		"borrador":     {"cotizado", "aprobado", "anulado"},
		"cotizado":     {"aprobado", "borrador", "anulado"},
		"aprobado":     {"en_ejecucion", "suspendido", "anulado"},
		"en_ejecucion": {"suspendido", "facturado", "cerrado"},
		"suspendido":   {"en_ejecucion", "anulado"},
		"facturado":    {"cerrado"},
		"cerrado":      {},
		"anulado":      {},
	}
	for _, candidate := range allowed[current] {
		if candidate == next {
			return true
		}
	}
	return false
}

func normalizeAIUText(v, fallback string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	v = strings.ReplaceAll(v, "-", "_")
	v = strings.ReplaceAll(v, " ", "_")
	if v == "" {
		return fallback
	}
	return v
}

func aiuClampPercent(v, fallback float64) float64 {
	if v < 0 {
		return 0
	}
	if v == 0 {
		return fallback
	}
	if v > 100 {
		return 100
	}
	return v
}

func aiuClampPercentAllowZero(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func aiuRoundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}
