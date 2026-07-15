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
	ServicioID            int64               `json:"servicio_id,omitempty"`
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
	Eventos               []EmpresaAIUEvento  `json:"eventos,omitempty"`
}

type EmpresaAIUItem struct {
	ID            int64   `json:"id"`
	EmpresaID     int64   `json:"empresa_id"`
	ContratoID    int64   `json:"contrato_id"`
	ServicioID    int64   `json:"servicio_id,omitempty"`
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
	CarritoID        int64   `json:"carrito_id,omitempty"`
	CarritoItemID    int64   `json:"carrito_item_id,omitempty"`
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

type EmpresaAIUEvento struct {
	ID             int64  `json:"id"`
	EmpresaID      int64  `json:"empresa_id"`
	ContratoID     int64  `json:"contrato_id"`
	Tipo           string `json:"tipo"`
	EstadoAnterior string `json:"estado_anterior,omitempty"`
	EstadoNuevo    string `json:"estado_nuevo,omitempty"`
	Usuario        string `json:"usuario,omitempty"`
	Detalle        string `json:"detalle,omitempty"`
	FechaCreacion  string `json:"fecha_creacion,omitempty"`
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
	UltimosEventos     []EmpresaAIUEvento   `json:"ultimos_eventos"`
}

func EnsureEmpresaAIUConstruccionSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_aiu_contratos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			cliente_id INTEGER DEFAULT 0,
			servicio_id INTEGER DEFAULT 0,
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
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			contrato_id INTEGER NOT NULL,
			servicio_id INTEGER DEFAULT 0,
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
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			contrato_id INTEGER NOT NULL,
			carrito_id INTEGER DEFAULT 0,
			carrito_item_id INTEGER DEFAULT 0,
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
		`CREATE TABLE IF NOT EXISTS empresa_aiu_eventos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			contrato_id INTEGER NOT NULL,
			tipo TEXT NOT NULL,
			estado_anterior TEXT,
			estado_nuevo TEXT,
			usuario TEXT,
			detalle TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS ix_aiu_eventos_contrato ON empresa_aiu_eventos(empresa_id, contrato_id, id DESC)`,
	}
	for _, stmt := range stmts {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	contractColumns := map[string]string{
		"responsable":                 "TEXT",
		"servicio_id":                 "INTEGER DEFAULT 0",
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
		"carrito_id":        "INTEGER DEFAULT 0",
		"carrito_item_id":   "INTEGER DEFAULT 0",
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
	if err := ensureColumnIfMissing(dbConn, "empresa_aiu_items", "servicio_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	for _, stmt := range []string{
		`CREATE INDEX IF NOT EXISTS ix_aiu_contratos_cliente ON empresa_aiu_contratos(empresa_id, cliente_id)`,
		`CREATE INDEX IF NOT EXISTS ix_aiu_contratos_servicio ON empresa_aiu_contratos(empresa_id, servicio_id)`,
		`CREATE INDEX IF NOT EXISTS ix_aiu_items_servicio ON empresa_aiu_items(empresa_id, servicio_id)`,
		`CREATE INDEX IF NOT EXISTS ix_aiu_facturas_carrito ON empresa_aiu_facturas(empresa_id, carrito_id)`,
	} {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
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
	if err := ValidateEmpresaAIUContrato(item); err != nil {
		return 0, err
	}
	clienteID, servicioID, err := prepareAIUContratoCoreRefs(dbConn, item, item.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	item.ClienteID, item.ServicioID = clienteID, servicioID
	previous, previousErr := GetEmpresaAIUContratoByCodigo(dbConn, item.EmpresaID, item.Codigo)
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_aiu_contratos
		(empresa_id,codigo,nombre,cliente_id,servicio_id,cliente_nombre,responsable,centro_costo,modalidad_contrato,tipo_obra,modelo_aiu,base_iva_modo,porcentaje_admin,porcentaje_imprevistos,porcentaje_utilidad,porcentaje_iva,porcentaje_retencion_fuente,porcentaje_retencion_ica,porcentaje_retencion_iva,porcentaje_anticipo,porcentaje_garantia,avance_porcentaje,fecha_inicio,fecha_fin,estado,riesgo_nivel,observaciones,usuario_creador)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (empresa_id,codigo) DO UPDATE SET
			nombre=EXCLUDED.nombre, cliente_id=EXCLUDED.cliente_id, servicio_id=EXCLUDED.servicio_id, cliente_nombre=EXCLUDED.cliente_nombre,
			responsable=EXCLUDED.responsable, centro_costo=EXCLUDED.centro_costo, modalidad_contrato=EXCLUDED.modalidad_contrato,
			tipo_obra=EXCLUDED.tipo_obra, modelo_aiu=EXCLUDED.modelo_aiu, base_iva_modo=EXCLUDED.base_iva_modo,
			porcentaje_admin=EXCLUDED.porcentaje_admin, porcentaje_imprevistos=EXCLUDED.porcentaje_imprevistos,
			porcentaje_utilidad=EXCLUDED.porcentaje_utilidad, porcentaje_iva=EXCLUDED.porcentaje_iva,
			porcentaje_retencion_fuente=EXCLUDED.porcentaje_retencion_fuente, porcentaje_retencion_ica=EXCLUDED.porcentaje_retencion_ica,
			porcentaje_retencion_iva=EXCLUDED.porcentaje_retencion_iva, porcentaje_anticipo=EXCLUDED.porcentaje_anticipo,
			porcentaje_garantia=EXCLUDED.porcentaje_garantia, avance_porcentaje=EXCLUDED.avance_porcentaje,
			fecha_inicio=EXCLUDED.fecha_inicio, fecha_fin=EXCLUDED.fecha_fin, estado=EXCLUDED.estado, riesgo_nivel=EXCLUDED.riesgo_nivel,
			observaciones=EXCLUDED.observaciones, usuario_creador=EXCLUDED.usuario_creador, fecha_actualizacion=CURRENT_TIMESTAMP`,
		item.EmpresaID, item.Codigo, item.Nombre, item.ClienteID, item.ServicioID, item.ClienteNombre, item.Responsable, item.CentroCosto, item.ModalidadContrato, item.TipoObra, item.ModeloAIU, item.BaseIVAModo,
		item.PorcentajeAdmin, item.PorcentajeImprevistos, item.PorcentajeUtilidad, item.PorcentajeIVA, item.PorcentajeRetFuente, item.PorcentajeRetICA, item.PorcentajeRetIVA,
		item.PorcentajeAnticipo, item.PorcentajeGarantia, item.AvancePorcentaje, item.FechaInicio, item.FechaFin, item.Estado, item.RiesgoNivel, item.Observaciones, item.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	row, err := GetEmpresaAIUContratoByCodigo(dbConn, item.EmpresaID, item.Codigo)
	if err != nil {
		return id, err
	}
	eventType := "contrato_creado"
	previousEstado := ""
	if previousErr == nil && previous.ID > 0 {
		eventType = "contrato_actualizado"
		previousEstado = previous.Estado
	}
	_ = RegistrarEmpresaAIUEvento(dbConn, item.EmpresaID, row.ID, eventType, previousEstado, row.Estado, item.UsuarioCreador, "Contrato AIU guardado")
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
	if item.Cantidad <= 0 || item.ValorUnitario <= 0 {
		return 0, errors.New("cantidad y valor unitario deben ser mayores que cero")
	}
	contrato, err := GetEmpresaAIUContrato(dbConn, item.EmpresaID, item.ContratoID)
	if err != nil {
		return 0, err
	}
	servicioID, err := ensureAIUItemServicio(dbConn, item, contrato, contrato.UsuarioCreador)
	if err != nil {
		return 0, err
	}
	item.ServicioID = servicioID
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_aiu_items
		(empresa_id,contrato_id,servicio_id,capitulo,descripcion,unidad,cantidad,valor_unitario,valor_total,estado)
		VALUES (?,?,?,?,?,?,?,?,?,?)`,
		item.EmpresaID, item.ContratoID, item.ServicioID, item.Capitulo, item.Descripcion, item.Unidad, item.Cantidad, item.ValorUnitario, item.ValorTotal, item.Estado)
	if err != nil {
		return 0, err
	}
	_ = RegistrarEmpresaAIUEvento(dbConn, item.EmpresaID, item.ContratoID, "concepto_agregado", "", "", "", fmt.Sprintf("%s: %.2f x %.2f", item.Descripcion, item.Cantidad, item.ValorUnitario))
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
			row.Eventos, _ = ListEmpresaAIUEventos(dbConn, empresaID, row.ID, 100)
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
			row.Eventos, _ = ListEmpresaAIUEventos(dbConn, empresaID, id, 100)
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
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,codigo,nombre,COALESCE(cliente_id,0),COALESCE(servicio_id,0),COALESCE(cliente_nombre,''),COALESCE(responsable,''),COALESCE(centro_costo,''),COALESCE(modalidad_contrato,'precio_global'),COALESCE(tipo_obra,'obra_civil'),COALESCE(modelo_aiu,'base_aiu_no_sumada'),COALESCE(base_iva_modo,'utilidad'),COALESCE(porcentaje_admin,0),COALESCE(porcentaje_imprevistos,0),COALESCE(porcentaje_utilidad,0),COALESCE(porcentaje_iva,19),COALESCE(porcentaje_retencion_fuente,0),COALESCE(porcentaje_retencion_ica,0),COALESCE(porcentaje_retencion_iva,0),COALESCE(porcentaje_anticipo,0),COALESCE(porcentaje_garantia,0),COALESCE(avance_porcentaje,0),COALESCE(fecha_inicio,''),COALESCE(fecha_fin,''),COALESCE(estado,'borrador'),COALESCE(riesgo_nivel,'medio'),COALESCE(costo_directo,0),COALESCE(valor_administracion,0),COALESCE(valor_imprevistos,0),COALESCE(valor_utilidad,0),COALESCE(aiu_total,0),COALESCE(base_iva,0),COALESCE(valor_iva,0),COALESCE(total_factura,0),COALESCE(valor_retencion_fuente,0),COALESCE(valor_retencion_ica,0),COALESCE(valor_retencion_iva,0),COALESCE(valor_anticipo,0),COALESCE(valor_garantia,0),COALESCE(neto_cobrar,0),COALESCE(documento_codigo,''),COALESCE(aprobado_por,''),COALESCE(fecha_aprobacion,''),COALESCE(usuario_creador,''),COALESCE(fecha_creacion,''),COALESCE(fecha_actualizacion,''),COALESCE(observaciones,'') FROM empresa_aiu_contratos WHERE %s ORDER BY id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaAIUContrato{}
	for rows.Next() {
		var x EmpresaAIUContrato
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.Codigo, &x.Nombre, &x.ClienteID, &x.ServicioID, &x.ClienteNombre, &x.Responsable, &x.CentroCosto, &x.ModalidadContrato, &x.TipoObra, &x.ModeloAIU, &x.BaseIVAModo, &x.PorcentajeAdmin, &x.PorcentajeImprevistos, &x.PorcentajeUtilidad, &x.PorcentajeIVA, &x.PorcentajeRetFuente, &x.PorcentajeRetICA, &x.PorcentajeRetIVA, &x.PorcentajeAnticipo, &x.PorcentajeGarantia, &x.AvancePorcentaje, &x.FechaInicio, &x.FechaFin, &x.Estado, &x.RiesgoNivel, &x.CostoDirecto, &x.ValorAdministracion, &x.ValorImprevistos, &x.ValorUtilidad, &x.AIUTotal, &x.BaseIVA, &x.ValorIVA, &x.TotalFactura, &x.ValorRetFuente, &x.ValorRetICA, &x.ValorRetIVA, &x.ValorAnticipo, &x.ValorGarantia, &x.NetoCobrar, &x.DocumentoCodigo, &x.AprobadoPor, &x.FechaAprobacion, &x.UsuarioCreador, &x.FechaCreacion, &x.FechaActualizacion, &x.Observaciones); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func ListEmpresaAIUItems(dbConn *sql.DB, empresaID, contratoID int64) ([]EmpresaAIUItem, error) {
	rows, err := ExecQueryCompat(dbConn, `SELECT id,empresa_id,contrato_id,COALESCE(servicio_id,0),COALESCE(capitulo,''),COALESCE(descripcion,''),COALESCE(unidad,'und'),COALESCE(cantidad,0),COALESCE(valor_unitario,0),COALESCE(valor_total,0),COALESCE(estado,'activo'),COALESCE(fecha_creacion,'') FROM empresa_aiu_items WHERE empresa_id=? AND contrato_id=? ORDER BY id`, empresaID, contratoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaAIUItem{}
	for rows.Next() {
		var x EmpresaAIUItem
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.ContratoID, &x.ServicioID, &x.Capitulo, &x.Descripcion, &x.Unidad, &x.Cantidad, &x.ValorUnitario, &x.ValorTotal, &x.Estado, &x.FechaCreacion); err != nil {
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
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,contrato_id,COALESCE(carrito_id,0),COALESCE(carrito_item_id,0),documento_codigo,COALESCE(tipo_documento,'factura_electronica'),COALESCE(periodo_contable,''),COALESCE(estado,'emitida'),COALESCE(costo_directo,0),COALESCE(aiu_total,0),COALESCE(base_iva,0),COALESCE(valor_iva,0),COALESCE(total_factura,0),COALESCE(valor_retenciones,0),COALESCE(valor_anticipo,0),COALESCE(valor_garantia,0),COALESCE(neto_cobrar,0),COALESCE(fecha_documento,''),COALESCE(usuario_creador,''),COALESCE(fecha_creacion,''),COALESCE(observaciones,'') FROM empresa_aiu_facturas WHERE %s ORDER BY id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaAIUFactura{}
	for rows.Next() {
		var x EmpresaAIUFactura
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.ContratoID, &x.CarritoID, &x.CarritoItemID, &x.DocumentoCodigo, &x.TipoDocumento, &x.PeriodoContable, &x.Estado, &x.CostoDirecto, &x.AIUTotal, &x.BaseIVA, &x.ValorIVA, &x.TotalFactura, &x.ValorRetenciones, &x.ValorAnticipo, &x.ValorGarantia, &x.NetoCobrar, &x.FechaDocumento, &x.UsuarioCreador, &x.FechaCreacion, &x.Observaciones); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func RegistrarEmpresaAIUEvento(dbConn *sql.DB, empresaID, contratoID int64, tipo, estadoAnterior, estadoNuevo, usuario, detalle string) error {
	if empresaID <= 0 || contratoID <= 0 {
		return nil
	}
	tipo = normalizeAIUText(tipo, "evento")
	estadoAnterior = strings.TrimSpace(estadoAnterior)
	estadoNuevo = strings.TrimSpace(estadoNuevo)
	usuario = strings.TrimSpace(usuario)
	detalle = strings.TrimSpace(detalle)
	_, err := ExecCompat(dbConn, `INSERT INTO empresa_aiu_eventos
		(empresa_id,contrato_id,tipo,estado_anterior,estado_nuevo,usuario,detalle)
		VALUES (?,?,?,?,?,?,?)`,
		empresaID, contratoID, tipo, estadoAnterior, estadoNuevo, usuario, detalle)
	return err
}

func ListEmpresaAIUEventos(dbConn *sql.DB, empresaID, contratoID int64, limit int) ([]EmpresaAIUEvento, error) {
	if err := EnsureEmpresaAIUConstruccionSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	args := []interface{}{empresaID}
	where := "empresa_id=?"
	if contratoID > 0 {
		where += " AND contrato_id=?"
		args = append(args, contratoID)
	}
	rows, err := ExecQueryCompat(dbConn, fmt.Sprintf(`SELECT id,empresa_id,contrato_id,COALESCE(tipo,''),COALESCE(estado_anterior,''),COALESCE(estado_nuevo,''),COALESCE(usuario,''),COALESCE(detalle,''),COALESCE(fecha_creacion,'') FROM empresa_aiu_eventos WHERE %s ORDER BY id DESC LIMIT %d`, where, limit), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EmpresaAIUEvento{}
	for rows.Next() {
		var x EmpresaAIUEvento
		if err := rows.Scan(&x.ID, &x.EmpresaID, &x.ContratoID, &x.Tipo, &x.EstadoAnterior, &x.EstadoNuevo, &x.Usuario, &x.Detalle, &x.FechaCreacion); err != nil {
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
	if !aiuContratoPuedeFacturarse(row.Estado) {
		return EmpresaAIUFactura{}, fmt.Errorf("contrato AIU no facturable en estado %s; debe estar aprobado o en ejecucion", row.Estado)
	}
	if len(row.Items) == 0 && row.CostoDirecto <= 0 {
		return EmpresaAIUFactura{}, errors.New("contrato AIU sin costo directo ni conceptos de obra")
	}
	row = CalculateEmpresaAIUContrato(row)
	if row.TotalFactura <= 0 || row.NetoCobrar <= 0 {
		return EmpresaAIUFactura{}, errors.New("contrato AIU sin valor neto positivo para facturar")
	}
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
	_ = RegistrarEmpresaAIUEvento(dbConn, empresaID, contratoID, "factura_generada", row.Estado, "facturado", usuario, fmt.Sprintf("Documento %s por %.2f neto %.2f", documentoCodigo, row.TotalFactura, row.NetoCobrar))
	facturas, err := ListEmpresaAIUFacturas(dbConn, empresaID, contratoID, 1)
	if err != nil || len(facturas) == 0 {
		return EmpresaAIUFactura{}, err
	}
	factura := facturas[0]
	carritoID, itemID, clienteID, servicioID, syncErr := createOrSyncAIUFacturaCarrito(dbConn, row, factura, usuario)
	if syncErr == nil {
		factura.CarritoID, factura.CarritoItemID = carritoID, itemID
		_, _ = ExecCompat(dbConn, `UPDATE empresa_aiu_facturas SET carrito_id=?, carrito_item_id=? WHERE empresa_id=? AND id=?`, nullableID(carritoID), nullableID(itemID), empresaID, factura.ID)
		_, _ = ExecCompat(dbConn, `UPDATE empresa_aiu_contratos SET cliente_id=?, servicio_id=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND id=?`, nullableID(clienteID), nullableID(servicioID), empresaID, contratoID)
	} else {
		_ = RegistrarEmpresaAIUEvento(dbConn, empresaID, contratoID, "integracion_nucleo_observada", "", "", usuario, syncErr.Error())
	}
	return factura, nil
}

func BuildEmpresaAIUDashboard(dbConn *sql.DB, empresaID int64) (EmpresaAIUDashboard, error) {
	contratos, err := ListEmpresaAIUContratos(dbConn, empresaID, "", 500)
	if err != nil {
		return EmpresaAIUDashboard{}, err
	}
	eventos, _ := ListEmpresaAIUEventos(dbConn, empresaID, 0, 20)
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
		UltimosEventos:     eventos,
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
		if c.Estado == "borrador" && c.CostoDirecto > 0 {
			out.Alertas = append(out.Alertas, fmt.Sprintf("%s tiene costos cargados pero sigue en borrador.", c.Codigo))
		}
		if c.RiesgoNivel == "critico" || c.RiesgoNivel == "alto" {
			out.Alertas = append(out.Alertas, fmt.Sprintf("%s esta marcado con riesgo %s.", c.Codigo, c.RiesgoNivel))
		}
		if strings.TrimSpace(c.FechaFin) != "" && c.Estado != "facturado" && c.Estado != "cerrado" && c.Estado != "anulado" {
			if end, err := time.Parse("2006-01-02", c.FechaFin); err == nil && end.Before(time.Now().AddDate(0, 0, -1)) {
				out.Alertas = append(out.Alertas, fmt.Sprintf("%s tiene fecha fin vencida y aun no esta cerrado.", c.Codigo))
			}
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

func prepareAIUContratoCoreRefs(dbConn *sql.DB, contrato EmpresaAIUContrato, usuario string) (int64, int64, error) {
	clienteID, err := ensureAIUClienteCore(dbConn, contrato, usuario)
	if err != nil {
		return 0, 0, err
	}
	servicioID, err := ensureAIUContratoServicio(dbConn, contrato, usuario)
	if err != nil {
		return 0, 0, err
	}
	return clienteID, servicioID, nil
}

func ensureAIUClienteCore(dbConn *sql.DB, contrato EmpresaAIUContrato, usuario string) (int64, error) {
	if contrato.ClienteID > 0 {
		return contrato.ClienteID, nil
	}
	if strings.TrimSpace(contrato.ClienteNombre) == "" {
		return 0, nil
	}
	if err := EnsureEmpresaClientesSchema(dbConn); err != nil {
		return 0, err
	}
	numeroDocumento := aiuCoreCode("AIU-CLI", contrato.Codigo, contrato.ClienteNombre)
	if id, err := findClienteDuplicateID(dbConn, fmt.Sprintf(`SELECT id FROM clientes WHERE empresa_id = ? AND %s = ? LIMIT 1`, clienteDocumentoSQLExpr("numero_documento")), contrato.EmpresaID, normalizeClienteDocumentoValue(numeroDocumento)); err != nil {
		return 0, err
	} else if id > 0 {
		return id, nil
	}
	nombre := strings.TrimSpace(contrato.ClienteNombre)
	if nombre == "" {
		nombre = "Cliente AIU"
	}
	id, err := CreateCliente(dbConn, Cliente{
		EmpresaID:         contrato.EmpresaID,
		TipoDocumento:     "OTRO",
		NumeroDocumento:   numeroDocumento,
		TipoPersona:       "juridica",
		NombreRazonSocial: nombre,
		NombreComercial:   nombre,
		Pais:              "CO",
		UsuarioCreador:    strings.TrimSpace(usuario),
		Estado:            "activo",
		Observaciones:     "Cliente creado/sincronizado desde AIU construccion.",
	})
	if err != nil {
		var dup *ClienteDuplicadoError
		if errors.As(err, &dup) && dup.ClienteID > 0 {
			return dup.ClienteID, nil
		}
		return 0, err
	}
	return id, nil
}

func ensureAIUContratoServicio(dbConn *sql.DB, contrato EmpresaAIUContrato, usuario string) (int64, error) {
	if contrato.ServicioID > 0 {
		return contrato.ServicioID, nil
	}
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		return 0, err
	}
	code := aiuCoreCode("AIU-CTR", contrato.Codigo)
	var id int64
	err := QueryRowCompat(dbConn, `SELECT id FROM servicios WHERE empresa_id=? AND UPPER(TRIM(COALESCE(codigo,'')))=UPPER(TRIM(?)) LIMIT 1`, contrato.EmpresaID, code).Scan(&id)
	if err == nil && id > 0 {
		return id, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	calculado := CalculateEmpresaAIUContrato(contrato)
	precio := calculado.TotalFactura
	if precio <= 0 {
		precio = calculado.CostoDirecto
	}
	nombre := strings.TrimSpace(contrato.Nombre)
	if nombre == "" {
		nombre = "Contrato AIU " + strings.TrimSpace(contrato.Codigo)
	}
	return CreateServicio(dbConn, Servicio{
		EmpresaID:          contrato.EmpresaID,
		Codigo:             code,
		Nombre:             nombre,
		Descripcion:        strings.TrimSpace(contrato.Observaciones),
		Categoria:          "AIU construccion",
		CostoReferencial:   calculado.CostoDirecto,
		Precio:             precio,
		ImpuestoPorcentaje: calculado.PorcentajeIVA,
		UsuarioCreador:     strings.TrimSpace(usuario),
		Estado:             "activo",
		Observaciones:      "Servicio sincronizado desde contrato AIU.",
	})
}

func ensureAIUItemServicio(dbConn *sql.DB, item EmpresaAIUItem, contrato EmpresaAIUContrato, usuario string) (int64, error) {
	if item.ServicioID > 0 {
		return item.ServicioID, nil
	}
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		return 0, err
	}
	code := aiuCoreCode("AIU-ITEM", contrato.Codigo, item.Capitulo, item.Descripcion)
	var id int64
	err := QueryRowCompat(dbConn, `SELECT id FROM servicios WHERE empresa_id=? AND UPPER(TRIM(COALESCE(codigo,'')))=UPPER(TRIM(?)) LIMIT 1`, item.EmpresaID, code).Scan(&id)
	if err == nil && id > 0 {
		return id, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}
	precio := item.ValorTotal
	if precio <= 0 {
		precio = item.ValorUnitario
	}
	categoria := "AIU construccion / conceptos"
	if strings.TrimSpace(item.Capitulo) != "" {
		categoria = "AIU construccion / " + strings.TrimSpace(item.Capitulo)
	}
	return CreateServicio(dbConn, Servicio{
		EmpresaID:          item.EmpresaID,
		Codigo:             code,
		Nombre:             strings.TrimSpace(item.Descripcion),
		Descripcion:        "Concepto de obra del contrato " + strings.TrimSpace(contrato.Codigo),
		Categoria:          categoria,
		CostoReferencial:   item.ValorUnitario,
		Precio:             precio,
		ImpuestoPorcentaje: 0,
		UsuarioCreador:     strings.TrimSpace(usuario),
		Estado:             "activo",
		Observaciones:      "Servicio sincronizado desde concepto AIU.",
	})
}

func createOrSyncAIUFacturaCarrito(dbConn *sql.DB, contrato EmpresaAIUContrato, factura EmpresaAIUFactura, usuario string) (int64, int64, int64, int64, error) {
	if factura.TotalFactura <= 0 {
		return factura.CarritoID, factura.CarritoItemID, contrato.ClienteID, contrato.ServicioID, nil
	}
	if err := EnsureEmpresaCarritosSchema(dbConn); err != nil {
		return 0, 0, 0, 0, err
	}
	clienteID, servicioID, err := prepareAIUContratoCoreRefs(dbConn, contrato, usuario)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	referenciaExterna := fmt.Sprintf("aiu_construccion:factura:%d:%s", factura.ID, strings.TrimSpace(factura.DocumentoCodigo))
	var carritoExistente, itemExistente int64
	err = QueryRowCompat(dbConn, `SELECT id FROM carritos_compras WHERE empresa_id=? AND referencia_externa=? LIMIT 1`, factura.EmpresaID, referenciaExterna).Scan(&carritoExistente)
	if err == nil && carritoExistente > 0 {
		_ = QueryRowCompat(dbConn, `SELECT id FROM carrito_compra_items WHERE empresa_id=? AND carrito_id=? AND referencia_id=? AND tipo_item='servicio' LIMIT 1`, factura.EmpresaID, carritoExistente, servicioID).Scan(&itemExistente)
		return carritoExistente, itemExistente, clienteID, servicioID, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, 0, 0, 0, err
	}
	carritoID, err := CreateCarritoCompra(dbConn, CarritoCompra{
		EmpresaID:         factura.EmpresaID,
		Codigo:            aiuCoreCode("AIU-FAC", factura.DocumentoCodigo),
		Nombre:            "Factura AIU " + strings.TrimSpace(factura.DocumentoCodigo),
		CanalVenta:        "aiu_construccion",
		ClienteID:         clienteID,
		EstadoCarrito:     "abierto",
		Moneda:            "COP",
		ReferenciaExterna: referenciaExterna,
		MetodoPago:        "transferencia_bancaria",
		ReferenciaPago:    factura.DocumentoCodigo,
		UsuarioCreador:    strings.TrimSpace(usuario),
		Observaciones:     "Venta central generada desde factura AIU; impuestos y retenciones se conservan en el documento AIU.",
	})
	if err != nil {
		return 0, 0, 0, 0, err
	}
	itemID, err := CreateCarritoCompraItem(dbConn, CarritoCompraItem{
		EmpresaID:          factura.EmpresaID,
		CarritoID:          carritoID,
		TipoItem:           "servicio",
		ReferenciaID:       servicioID,
		CodigoItem:         aiuCoreCode("AIU-FAC-ITEM", factura.DocumentoCodigo),
		Descripcion:        "Contrato AIU " + strings.TrimSpace(contrato.Codigo),
		UnidadMedida:       "contrato",
		Cantidad:           1,
		PrecioUnitario:     factura.TotalFactura,
		ImpuestoPorcentaje: 0,
		UsuarioCreador:     strings.TrimSpace(usuario),
		Estado:             "activo",
		Observaciones:      factura.Observaciones,
	})
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return carritoID, itemID, clienteID, servicioID, nil
}

func aiuCoreCode(prefix string, parts ...string) string {
	replacer := strings.NewReplacer(
		"á", "A", "é", "E", "í", "I", "ó", "O", "ú", "U", "ñ", "N",
		"Á", "A", "É", "E", "Í", "I", "Ó", "O", "Ú", "U", "Ñ", "N",
		"ä", "A", "ë", "E", "ï", "I", "ö", "O", "ü", "U",
		"Ä", "A", "Ë", "E", "Ï", "I", "Ö", "O", "Ü", "U",
	)
	var b strings.Builder
	for _, part := range parts {
		part = strings.ToUpper(replacer.Replace(strings.TrimSpace(part)))
		for _, r := range part {
			if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				b.WriteRune(r)
				continue
			}
			if b.Len() > 0 && b.String()[b.Len()-1] != '-' {
				b.WriteRune('-')
			}
		}
		if b.Len() > 0 && b.String()[b.Len()-1] != '-' {
			b.WriteRune('-')
		}
	}
	code := strings.Trim(b.String(), "-")
	if code == "" {
		code = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if len(code) > 42 {
		code = code[:42]
	}
	prefixCode := strings.Trim(strings.ToUpper(strings.NewReplacer(" ", "-", "_", "-").Replace(strings.TrimSpace(prefix))), "-")
	if prefixCode == "" {
		prefixCode = "AIU"
	}
	return prefixCode + "-" + strings.Trim(code, "-")
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

func ValidateEmpresaAIUContrato(x EmpresaAIUContrato) error {
	if strings.TrimSpace(x.FechaInicio) != "" {
		if _, err := time.Parse("2006-01-02", strings.TrimSpace(x.FechaInicio)); err != nil {
			return errors.New("fecha_inicio debe tener formato YYYY-MM-DD")
		}
	}
	if strings.TrimSpace(x.FechaFin) != "" {
		if _, err := time.Parse("2006-01-02", strings.TrimSpace(x.FechaFin)); err != nil {
			return errors.New("fecha_fin debe tener formato YYYY-MM-DD")
		}
	}
	if strings.TrimSpace(x.FechaInicio) != "" && strings.TrimSpace(x.FechaFin) != "" {
		inicio, _ := time.Parse("2006-01-02", strings.TrimSpace(x.FechaInicio))
		fin, _ := time.Parse("2006-01-02", strings.TrimSpace(x.FechaFin))
		if fin.Before(inicio) {
			return errors.New("fecha_fin no puede ser anterior a fecha_inicio")
		}
	}
	if x.Estado == "aprobado" && strings.TrimSpace(x.Responsable) == "" {
		return errors.New("un contrato aprobado requiere responsable de obra")
	}
	if (x.Estado == "en_ejecucion" || x.Estado == "facturado" || x.Estado == "cerrado") && strings.TrimSpace(x.CentroCosto) == "" {
		return errors.New("contratos en ejecucion, facturados o cerrados requieren centro de costo")
	}
	return nil
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
	if next == "aprobado" && strings.TrimSpace(row.Responsable) == "" {
		return EmpresaAIUContrato{}, errors.New("para aprobar debes asignar responsable de obra")
	}
	if next == "en_ejecucion" && strings.TrimSpace(row.CentroCosto) == "" {
		return EmpresaAIUContrato{}, errors.New("para iniciar debes asignar centro de costo")
	}
	if next == "facturado" {
		return EmpresaAIUContrato{}, errors.New("usa Generar factura para cambiar un contrato a facturado con documento electronico")
	}
	if next == "cerrado" && row.Estado != "facturado" {
		return EmpresaAIUContrato{}, errors.New("solo puedes cerrar contratos facturados")
	}
	if next == "cerrado" && row.AvancePorcentaje < 100 {
		return EmpresaAIUContrato{}, errors.New("para cerrar el avance debe estar en 100%")
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
	_ = RegistrarEmpresaAIUEvento(dbConn, empresaID, contratoID, "cambio_estado", row.Estado, next, usuario, observacion)
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

func aiuContratoPuedeFacturarse(estado string) bool {
	switch normalizeEmpresaAIUEstado(estado) {
	case "aprobado", "en_ejecucion":
		return true
	default:
		return false
	}
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
