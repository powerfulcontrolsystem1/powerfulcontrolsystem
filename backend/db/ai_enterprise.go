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

const (
	AIProposalDraft                = "draft"
	AIProposalAwaitingInformation  = "awaiting_information"
	AIProposalAwaitingConfirmation = "awaiting_confirmation"
	AIProposalApproved             = "approved"
	AIProposalExecuting            = "executing"
	AIProposalCompleted            = "completed"
	AIProposalFailed               = "failed"
	AIProposalExpired              = "expired"
	AIProposalCancelled            = "cancelled"
	AIProposalRolledBack           = "rolled_back"
)

// EmpresaAIProposal is a server-owned approval record. PlanJSON never contains credentials.
type EmpresaAIProposal struct {
	ProposalID      string `json:"proposal_id"`
	ConversationID  string `json:"conversation_id"`
	EmpresaID       int64  `json:"empresa_id"`
	UsuarioCreador  string `json:"usuario_creador"`
	ToolName        string `json:"tool_name"`
	RiskLevel       string `json:"risk_level"`
	PlanJSON        string `json:"plan_json"`
	PlanHash        string `json:"plan_hash"`
	EstadoAnterior  string `json:"estado_anterior_json"`
	EstadoEsperado  string `json:"estado_esperado_json"`
	Resumen         string `json:"resumen"`
	RollbackPolicy  string `json:"rollback_policy"`
	Estado          string `json:"estado"`
	IdempotencyKey  string `json:"idempotency_key,omitempty"`
	FechaCreacion   string `json:"fecha_creacion,omitempty"`
	FechaExpiracion string `json:"fecha_expiracion,omitempty"`
	FechaUso        string `json:"fecha_uso,omitempty"`
	ResultadoJSON   string `json:"resultado_json,omitempty"`
}

// EmpresaAIConversation keeps the execution state independent of model text.
// It is deliberately sparse: no prompt, attachment, credential or provider
// response is persisted here.
type EmpresaAIConversation struct {
	ConversationID      string `json:"conversation_id"`
	EmpresaID           int64  `json:"empresa_id"`
	UsuarioID           string `json:"usuario_id"`
	Modo                string `json:"modo"`
	Estado              string `json:"estado"`
	Intencion           string `json:"intencion,omitempty"`
	EntidadesJSON       string `json:"entidades_json,omitempty"`
	CamposFaltantesJSON string `json:"campos_faltantes_json,omitempty"`
	PlanActualJSON      string `json:"plan_actual_json,omitempty"`
	PropuestaPendiente  string `json:"propuesta_pendiente,omitempty"`
	FechaExpiracion     string `json:"fecha_expiracion,omitempty"`
}

// EmpresaAIExecution stores only operational metadata and sanitized result
// summaries. It is not a transcript and never stores provider secrets.
type EmpresaAIExecution struct {
	EmpresaID      int64  `json:"empresa_id"`
	UsuarioID      string `json:"usuario_id"`
	ConversationID string `json:"conversation_id"`
	ProposalID     string `json:"proposal_id,omitempty"`
	ToolName       string `json:"tool_name"`
	Modo           string `json:"modo"`
	RiskLevel      string `json:"risk_level"`
	Resultado      string `json:"resultado"`
	ErrorCategoria string `json:"error_categoria,omitempty"`
	FuentesJSON    string `json:"fuentes_json"`
	CategoriasJSON string `json:"categorias_datos_json"`
	DuracionMS     int64  `json:"duracion_ms"`
}

func EnsureEmpresaAIEnterpriseSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_ai_conversaciones (
			conversation_id TEXT PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			usuario_id TEXT NOT NULL,
			modo TEXT NOT NULL,
			estado TEXT NOT NULL DEFAULT 'active',
			intencion TEXT NOT NULL DEFAULT '',
			entidades_json TEXT NOT NULL DEFAULT '{}',
			campos_faltantes_json TEXT NOT NULL DEFAULT '[]',
			plan_actual_json TEXT NOT NULL DEFAULT '{}',
			propuesta_pendiente TEXT NOT NULL DEFAULT '',
			fecha_expiracion TIMESTAMP NOT NULL,
			fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS empresa_ai_propuestas (
			proposal_id TEXT PRIMARY KEY,
			conversation_id TEXT NOT NULL,
			empresa_id BIGINT NOT NULL,
			usuario_creador TEXT NOT NULL,
			tool_name TEXT NOT NULL,
			risk_level TEXT NOT NULL DEFAULT 'low',
			plan_json TEXT NOT NULL,
			plan_hash TEXT NOT NULL,
			estado_anterior_json TEXT NOT NULL DEFAULT '{}',
			estado_esperado_json TEXT NOT NULL DEFAULT '{}',
			resumen TEXT NOT NULL,
			rollback_policy TEXT NOT NULL DEFAULT 'none',
			estado TEXT NOT NULL DEFAULT 'draft',
			idempotency_key TEXT,
			fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			fecha_expiracion TIMESTAMP NOT NULL,
			fecha_uso TIMESTAMP,
			resultado_json TEXT,
			fecha_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_ai_propuestas_empresa_estado ON empresa_ai_propuestas(empresa_id, estado, fecha_creacion DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_ai_conversaciones_scope ON empresa_ai_conversaciones(empresa_id, usuario_id, estado, fecha_actualizacion DESC);`,
		`CREATE TABLE IF NOT EXISTS empresa_ai_ejecuciones (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			usuario_id TEXT NOT NULL,
			conversation_id TEXT NOT NULL,
			proposal_id TEXT NOT NULL DEFAULT '',
			tool_name TEXT NOT NULL,
			modo TEXT NOT NULL,
			risk_level TEXT NOT NULL,
			resultado TEXT NOT NULL,
			error_categoria TEXT NOT NULL DEFAULT '',
			fuentes_json TEXT NOT NULL DEFAULT '[]',
			categorias_datos_json TEXT NOT NULL DEFAULT '[]',
			duracion_ms BIGINT NOT NULL DEFAULT 0,
			fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_ai_ejecuciones_scope ON empresa_ai_ejecuciones(empresa_id, usuario_id, fecha_creacion DESC);`,
		`CREATE TABLE IF NOT EXISTS empresa_ai_memoria (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			usuario_id TEXT NOT NULL DEFAULT '',
			tipo TEXT NOT NULL,
			clave TEXT NOT NULL,
			valor_json TEXT NOT NULL,
			consentida BOOLEAN NOT NULL DEFAULT FALSE,
			fecha_expiracion TIMESTAMP,
			fecha_creacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(empresa_id, usuario_id, tipo, clave)
		);`,
		`CREATE TABLE IF NOT EXISTS empresa_ai_fuentes_conocimiento (
			document_id TEXT PRIMARY KEY,
			empresa_id BIGINT NOT NULL DEFAULT 0,
			ruta TEXT NOT NULL,
			version_ref TEXT NOT NULL DEFAULT '',
			hash_contenido TEXT NOT NULL,
			modulo TEXT NOT NULL DEFAULT '',
			nivel_confidencialidad TEXT NOT NULL DEFAULT 'internal',
			permisos_requeridos TEXT NOT NULL DEFAULT '',
			estado TEXT NOT NULL DEFAULT 'active',
			fecha_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_ai_propuestas_usuario_estado ON empresa_ai_propuestas(empresa_id, usuario_creador, estado);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_ai_propuestas_idempotency ON empresa_ai_propuestas(empresa_id, usuario_creador, idempotency_key) WHERE idempotency_key IS NOT NULL AND idempotency_key <> '';`,
	}
	for _, stmt := range stmts {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func RecordEmpresaAIExecution(dbConn *sql.DB, in EmpresaAIExecution) error {
	if err := EnsureEmpresaAIEnterpriseSchema(dbConn); err != nil {
		return err
	}
	if in.EmpresaID <= 0 || strings.TrimSpace(in.UsuarioID) == "" || strings.TrimSpace(in.ConversationID) == "" || strings.TrimSpace(in.ToolName) == "" || !json.Valid([]byte(defaultAIJSON(in.FuentesJSON, "[]"))) || !json.Valid([]byte(defaultAIJSON(in.CategoriasJSON, "[]"))) {
		return fmt.Errorf("ejecucion IA incompleta")
	}
	_, err := execSQLCompat(dbConn, `INSERT INTO empresa_ai_ejecuciones (empresa_id,usuario_id,conversation_id,proposal_id,tool_name,modo,risk_level,resultado,error_categoria,fuentes_json,categorias_datos_json,duracion_ms) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`, in.EmpresaID, strings.TrimSpace(in.UsuarioID), strings.TrimSpace(in.ConversationID), strings.TrimSpace(in.ProposalID), strings.TrimSpace(in.ToolName), strings.TrimSpace(in.Modo), strings.TrimSpace(in.RiskLevel), strings.TrimSpace(in.Resultado), strings.TrimSpace(in.ErrorCategoria), defaultAIJSON(in.FuentesJSON, "[]"), defaultAIJSON(in.CategoriasJSON, "[]"), in.DuracionMS)
	return err
}

// EnsureEmpresaAIConversation creates or resumes only a conversation belonging
// to the authenticated user in the current company.
func EnsureEmpresaAIConversation(dbConn *sql.DB, in EmpresaAIConversation, ttl time.Duration) (*EmpresaAIConversation, error) {
	if err := EnsureEmpresaAIEnterpriseSchema(dbConn); err != nil {
		return nil, err
	}
	if in.EmpresaID <= 0 || strings.TrimSpace(in.ConversationID) == "" || strings.TrimSpace(in.UsuarioID) == "" {
		return nil, fmt.Errorf("conversacion IA incompleta")
	}
	if !json.Valid([]byte(defaultAIJSON(in.EntidadesJSON, "{}"))) || !json.Valid([]byte(defaultAIJSON(in.CamposFaltantesJSON, "[]"))) || !json.Valid([]byte(defaultAIJSON(in.PlanActualJSON, "{}"))) {
		return nil, fmt.Errorf("estado de conversacion IA invalido")
	}
	if ttl <= 0 || ttl > 24*time.Hour {
		ttl = 2 * time.Hour
	}
	in.Modo = strings.ToLower(strings.TrimSpace(in.Modo))
	if in.Modo == "" {
		in.Modo = "assisted"
	}
	expires := time.Now().UTC().Add(ttl)
	_, err := execSQLCompat(dbConn, `INSERT INTO empresa_ai_conversaciones (conversation_id,empresa_id,usuario_id,modo,estado,intencion,entidades_json,campos_faltantes_json,plan_actual_json,propuesta_pendiente,fecha_expiracion,fecha_actualizacion) VALUES (?,?,?,?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP) ON CONFLICT (conversation_id) DO UPDATE SET modo=EXCLUDED.modo, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_ai_conversaciones.empresa_id=EXCLUDED.empresa_id AND empresa_ai_conversaciones.usuario_id=EXCLUDED.usuario_id AND empresa_ai_conversaciones.fecha_expiracion>CURRENT_TIMESTAMP`, strings.TrimSpace(in.ConversationID), in.EmpresaID, strings.TrimSpace(in.UsuarioID), in.Modo, "active", strings.TrimSpace(in.Intencion), defaultAIJSON(in.EntidadesJSON, "{}"), defaultAIJSON(in.CamposFaltantesJSON, "[]"), defaultAIJSON(in.PlanActualJSON, "{}"), strings.TrimSpace(in.PropuestaPendiente), expires)
	if err != nil {
		return nil, err
	}
	in.FechaExpiracion = expires.Format(time.RFC3339)
	return &in, nil
}

func defaultAIJSON(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

// EmpresaAIHotelRoomStationSnapshot is the server-derived before-state shown to
// the user. It never exposes the full raw station configuration.
type EmpresaAIHotelRoomStationSnapshot struct {
	EstacionID int64                    `json:"estacion_id"`
	Nombre     string                   `json:"nombre"`
	Tipo       string                   `json:"tipo"`
	Activa     bool                     `json:"activa"`
	Moneda     string                   `json:"moneda"`
	Tarifas    []EmpresaAIHotelRoomRate `json:"tarifas"`
}

func GetEmpresaAIHotelRoomStationSnapshot(dbConn *sql.DB, empresaID, estacionID int64) (*EmpresaAIHotelRoomStationSnapshot, error) {
	if empresaID <= 0 || estacionID <= 0 {
		return nil, fmt.Errorf("contexto de estacion invalido")
	}
	pref, err := GetEmpresaEstacionPref(dbConn, empresaID, 0, "estaciones_config")
	if err != nil || pref == nil {
		return nil, fmt.Errorf("no se encontro la estacion solicitada")
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(pref.Valor), &cfg); err != nil {
		return nil, fmt.Errorf("configuracion de estaciones invalida")
	}
	items, ok := cfg["estaciones"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no se encontro la estacion solicitada")
	}
	snapshot := &EmpresaAIHotelRoomStationSnapshot{EstacionID: estacionID, Tarifas: []EmpresaAIHotelRoomRate{}}
	found := false
	for _, raw := range items {
		station, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		id, _ := station["id"].(float64)
		if int64(id) != estacionID {
			continue
		}
		found = true
		snapshot.Nombre, _ = station["nombre"].(string)
		snapshot.Tipo, _ = station["tipo_estacion"].(string)
		snapshot.Moneda, _ = station["moneda"].(string)
		snapshot.Activa, _ = station["activa"].(bool)
		break
	}
	if !found {
		return nil, fmt.Errorf("la estacion no pertenece a la empresa")
	}
	rates, err := ListEmpresaTarifasPorDia(dbConn, empresaID, EmpresaTarifaPorDiaFilter{EstacionID: estacionID, IncludeInactive: true})
	if err != nil {
		return nil, err
	}
	for _, rate := range rates {
		snapshot.Tarifas = append(snapshot.Tarifas, EmpresaAIHotelRoomRate{Personas: rate.PersonasDesde, Valor: rate.ValorDia})
	}
	return snapshot, nil
}

func normalizeAIProposalState(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case AIProposalDraft, AIProposalAwaitingInformation, AIProposalAwaitingConfirmation, AIProposalApproved, AIProposalExecuting, AIProposalCompleted, AIProposalFailed, AIProposalExpired, AIProposalCancelled, AIProposalRolledBack:
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return AIProposalDraft
	}
}

func CreateEmpresaAIProposal(dbConn *sql.DB, in EmpresaAIProposal, ttl time.Duration) (*EmpresaAIProposal, error) {
	if err := EnsureEmpresaAIEnterpriseSchema(dbConn); err != nil {
		return nil, err
	}
	if in.EmpresaID <= 0 || strings.TrimSpace(in.ProposalID) == "" || strings.TrimSpace(in.ConversationID) == "" || strings.TrimSpace(in.UsuarioCreador) == "" || strings.TrimSpace(in.ToolName) == "" {
		return nil, fmt.Errorf("propuesta IA incompleta")
	}
	if !json.Valid([]byte(in.PlanJSON)) || !json.Valid([]byte(in.EstadoAnterior)) || !json.Valid([]byte(in.EstadoEsperado)) {
		return nil, fmt.Errorf("JSON de propuesta invalido")
	}
	planHash := sha256.Sum256([]byte(in.PlanJSON))
	in.PlanHash = hex.EncodeToString(planHash[:])
	in.Estado = normalizeAIProposalState(in.Estado)
	if in.Estado == AIProposalDraft {
		in.Estado = AIProposalAwaitingConfirmation
	}
	if ttl <= 0 || ttl > 30*time.Minute {
		ttl = 15 * time.Minute
	}
	now := time.Now().UTC()
	expires := now.Add(ttl)
	_, err := execSQLCompat(dbConn, `INSERT INTO empresa_ai_propuestas (
		proposal_id, conversation_id, empresa_id, usuario_creador, tool_name, risk_level, plan_json, plan_hash,
		estado_anterior_json, estado_esperado_json, resumen, rollback_policy, estado, idempotency_key, fecha_expiracion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		strings.TrimSpace(in.ProposalID), strings.TrimSpace(in.ConversationID), in.EmpresaID, strings.TrimSpace(in.UsuarioCreador), strings.TrimSpace(in.ToolName), strings.TrimSpace(in.RiskLevel), in.PlanJSON, in.PlanHash, in.EstadoAnterior, in.EstadoEsperado, strings.TrimSpace(in.Resumen), strings.TrimSpace(in.RollbackPolicy), in.Estado, strings.TrimSpace(in.IdempotencyKey), expires)
	if err != nil {
		return nil, err
	}
	in.FechaCreacion, in.FechaExpiracion = now.Format(time.RFC3339), expires.Format(time.RFC3339)
	return &in, nil
}

func GetEmpresaAIProposal(dbConn *sql.DB, empresaID int64, proposalID string) (*EmpresaAIProposal, error) {
	if err := EnsureEmpresaAIEnterpriseSchema(dbConn); err != nil {
		return nil, err
	}
	if empresaID <= 0 || strings.TrimSpace(proposalID) == "" {
		return nil, fmt.Errorf("empresa_id y proposal_id son obligatorios")
	}
	var p EmpresaAIProposal
	err := queryRowSQLCompat(dbConn, `SELECT proposal_id, conversation_id, empresa_id, usuario_creador, tool_name, risk_level, plan_json, plan_hash, estado_anterior_json, estado_esperado_json, resumen, rollback_policy, estado, COALESCE(idempotency_key,''), COALESCE(CAST(fecha_creacion AS TEXT),''), COALESCE(CAST(fecha_expiracion AS TEXT),''), COALESCE(CAST(fecha_uso AS TEXT),''), COALESCE(resultado_json,'') FROM empresa_ai_propuestas WHERE empresa_id=? AND proposal_id=?`, empresaID, strings.TrimSpace(proposalID)).Scan(&p.ProposalID, &p.ConversationID, &p.EmpresaID, &p.UsuarioCreador, &p.ToolName, &p.RiskLevel, &p.PlanJSON, &p.PlanHash, &p.EstadoAnterior, &p.EstadoEsperado, &p.Resumen, &p.RollbackPolicy, &p.Estado, &p.IdempotencyKey, &p.FechaCreacion, &p.FechaExpiracion, &p.FechaUso, &p.ResultadoJSON)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// BeginEmpresaAIProposalExecution atomically consumes a proposal after all caller-side permission checks.
func BeginEmpresaAIProposalExecution(dbConn *sql.DB, empresaID int64, proposalID, usuario, planHash, idempotencyKey string) (*EmpresaAIProposal, error) {
	if err := EnsureEmpresaAIEnterpriseSchema(dbConn); err != nil {
		return nil, err
	}
	if empresaID <= 0 || strings.TrimSpace(proposalID) == "" || strings.TrimSpace(usuario) == "" || strings.TrimSpace(planHash) == "" || strings.TrimSpace(idempotencyKey) == "" {
		return nil, fmt.Errorf("confirmacion de propuesta incompleta")
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var p EmpresaAIProposal
	err = queryRowTxSQLCompat(tx, `SELECT proposal_id, conversation_id, empresa_id, usuario_creador, tool_name, risk_level, plan_json, plan_hash, estado_anterior_json, estado_esperado_json, resumen, rollback_policy, estado, COALESCE(idempotency_key,'') FROM empresa_ai_propuestas WHERE empresa_id=? AND proposal_id=? FOR UPDATE`, empresaID, proposalID).Scan(&p.ProposalID, &p.ConversationID, &p.EmpresaID, &p.UsuarioCreador, &p.ToolName, &p.RiskLevel, &p.PlanJSON, &p.PlanHash, &p.EstadoAnterior, &p.EstadoEsperado, &p.Resumen, &p.RollbackPolicy, &p.Estado, &p.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	if p.UsuarioCreador != strings.TrimSpace(usuario) {
		return nil, fmt.Errorf("propuesta fuera del alcance del usuario")
	}
	if p.PlanHash != strings.TrimSpace(planHash) {
		return nil, fmt.Errorf("el plan ha cambiado o no corresponde a la propuesta")
	}
	if p.Estado != AIProposalAwaitingConfirmation {
		return nil, fmt.Errorf("la propuesta no esta disponible para confirmacion")
	}
	res, err := execTxSQLCompat(tx, `UPDATE empresa_ai_propuestas SET estado=?, idempotency_key=?, fecha_uso=CURRENT_TIMESTAMP, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND proposal_id=? AND estado=? AND fecha_expiracion>CURRENT_TIMESTAMP`, AIProposalExecuting, strings.TrimSpace(idempotencyKey), empresaID, proposalID, AIProposalAwaitingConfirmation)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return nil, fmt.Errorf("la propuesta vencio, ya fue usada o cambio de estado")
	}
	p.Estado, p.IdempotencyKey = AIProposalExecuting, strings.TrimSpace(idempotencyKey)
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &p, nil
}

func FinishEmpresaAIProposal(dbConn *sql.DB, empresaID int64, proposalID, state, resultJSON string) error {
	if !json.Valid([]byte(resultJSON)) {
		return fmt.Errorf("resultado_json invalido")
	}
	state = normalizeAIProposalState(state)
	if state != AIProposalCompleted && state != AIProposalFailed && state != AIProposalRolledBack {
		return fmt.Errorf("estado final de propuesta invalido")
	}
	res, err := execSQLCompat(dbConn, `UPDATE empresa_ai_propuestas SET estado=?, resultado_json=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND proposal_id=? AND estado=?`, state, resultJSON, empresaID, proposalID, AIProposalExecuting)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return sql.ErrNoRows
	}
	return nil
}

func CancelEmpresaAIProposal(dbConn *sql.DB, empresaID int64, proposalID, usuario string) error {
	res, err := execSQLCompat(dbConn, `UPDATE empresa_ai_propuestas SET estado=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND proposal_id=? AND usuario_creador=? AND estado IN (?,?)`, AIProposalCancelled, empresaID, proposalID, strings.TrimSpace(usuario), AIProposalDraft, AIProposalAwaitingConfirmation)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return sql.ErrNoRows
	}
	return nil
}

// EmpresaAIHotelRoomPlan is the closed input accepted by the first write tool.
// It deliberately models only daily hotel prices; motel/hourly rules remain a separate tool.
type EmpresaAIHotelRoomPlan struct {
	EstacionID             int64                    `json:"estacion_id"`
	NombreHabitacion       string                   `json:"nombre_habitacion"`
	Moneda                 string                   `json:"moneda"`
	HoraCheckIn            string                   `json:"hora_check_in"`
	HoraCheckOut           string                   `json:"hora_check_out"`
	Activa                 bool                     `json:"activa"`
	ConservarConfiguracion bool                     `json:"conservar_configuracion"`
	Tarifas                []EmpresaAIHotelRoomRate `json:"tarifas"`
}

type EmpresaAIHotelRoomRate struct {
	Personas int     `json:"personas"`
	Valor    float64 `json:"valor"`
}

// EmpresaAIProductCreatePlan is deliberately smaller than Producto. It accepts
// only values that are meaningful for a first catalog entry; ownership and the
// creator are always derived from the authenticated execution context.
type EmpresaAIProductCreatePlan struct {
	Nombre             string  `json:"nombre"`
	SKU                string  `json:"sku,omitempty"`
	Descripcion        string  `json:"descripcion,omitempty"`
	CategoriaID        int64   `json:"categoria_id,omitempty"`
	BodegaID           int64   `json:"bodega_id,omitempty"`
	UnidadMedida       string  `json:"unidad_medida,omitempty"`
	Costo              float64 `json:"costo,omitempty"`
	Precio             float64 `json:"precio"`
	ImpuestoPorcentaje float64 `json:"impuesto_porcentaje,omitempty"`
	StockInicial       float64 `json:"stock_inicial,omitempty"`
	StockMinimo        float64 `json:"stock_minimo,omitempty"`
}

func NormalizeEmpresaAIProductCreatePlan(plan *EmpresaAIProductCreatePlan) error {
	if plan == nil {
		return fmt.Errorf("plan de producto invalido")
	}
	plan.Nombre = strings.TrimSpace(plan.Nombre)
	plan.SKU = strings.TrimSpace(plan.SKU)
	plan.Descripcion = strings.TrimSpace(plan.Descripcion)
	plan.UnidadMedida = strings.TrimSpace(plan.UnidadMedida)
	if plan.Nombre == "" || len([]rune(plan.Nombre)) > 160 {
		return fmt.Errorf("nombre de producto invalido")
	}
	if len([]rune(plan.SKU)) > 80 || len([]rune(plan.Descripcion)) > 1000 || len([]rune(plan.UnidadMedida)) > 40 {
		return fmt.Errorf("texto de producto excede el limite permitido")
	}
	if plan.Precio < 0 || plan.Costo < 0 || plan.StockInicial < 0 || plan.StockMinimo < 0 || plan.ImpuestoPorcentaje < 0 || plan.ImpuestoPorcentaje > 100 {
		return fmt.Errorf("valores de producto invalidos")
	}
	if plan.StockInicial > 0 && plan.BodegaID <= 0 {
		return fmt.Errorf("bodega_id es obligatorio cuando se registra stock inicial")
	}
	return nil
}

// FindEmpresaAIProductDuplicates is a bounded, tenant-scoped duplicate check
// used before the agent creates a product proposal.
func FindEmpresaAIProductDuplicates(dbConn *sql.DB, empresaID int64, nombre, sku string) ([]Producto, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa invalida")
	}
	nombre = strings.TrimSpace(nombre)
	sku = strings.TrimSpace(sku)
	if nombre == "" && sku == "" {
		return nil, fmt.Errorf("nombre o sku requerido")
	}
	const selectProductsByName = `SELECT id, empresa_id, COALESCE(sku,''), COALESCE(nombre,''), COALESCE(precio,0), COALESCE(impuesto_porcentaje,0), COALESCE(estado,'activo') FROM productos WHERE empresa_id=? AND LOWER(COALESCE(nombre,'')) = LOWER(?) ORDER BY id DESC LIMIT 10`
	const selectProductsByNameOrSKU = `SELECT id, empresa_id, COALESCE(sku,''), COALESCE(nombre,''), COALESCE(precio,0), COALESCE(impuesto_porcentaje,0), COALESCE(estado,'activo') FROM productos WHERE empresa_id=? AND (LOWER(COALESCE(nombre,'')) = LOWER(?) OR LOWER(COALESCE(sku,'')) = LOWER(?)) ORDER BY id DESC LIMIT 10`
	var rows *sql.Rows
	var err error
	if sku == "" {
		rows, err = dbConn.Query(selectProductsByName, empresaID, nombre)
	} else {
		rows, err = dbConn.Query(selectProductsByNameOrSKU, empresaID, nombre, sku)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Producto, 0)
	for rows.Next() {
		var product Producto
		if err := rows.Scan(&product.ID, &product.EmpresaID, &product.SKU, &product.Nombre, &product.Precio, &product.ImpuestoPorcentaje, &product.Estado); err != nil {
			return nil, err
		}
		out = append(out, product)
	}
	return out, rows.Err()
}

// CreateEmpresaAIProduct reuses the canonical product creation service. The
// latter validates catalog relations and writes inventory/history atomically.
func CreateEmpresaAIProduct(dbConn *sql.DB, empresaID int64, plan EmpresaAIProductCreatePlan, usuario string) (int64, error) {
	if empresaID <= 0 || strings.TrimSpace(usuario) == "" {
		return 0, fmt.Errorf("contexto empresarial invalido")
	}
	if err := NormalizeEmpresaAIProductCreatePlan(&plan); err != nil {
		return 0, err
	}
	duplicates, err := FindEmpresaAIProductDuplicates(dbConn, empresaID, plan.Nombre, plan.SKU)
	if err != nil {
		return 0, err
	}
	if len(duplicates) > 0 {
		return 0, fmt.Errorf("ya existe un producto con el mismo nombre o SKU")
	}
	return CreateProducto(dbConn, Producto{
		EmpresaID: empresaID, BodegaPrincipalID: plan.BodegaID, CategoriaID: plan.CategoriaID,
		SKU: plan.SKU, Nombre: plan.Nombre, Descripcion: plan.Descripcion, UnidadMedida: plan.UnidadMedida,
		Costo: plan.Costo, Precio: plan.Precio, ImpuestoPorcentaje: plan.ImpuestoPorcentaje,
		StockMinimo: plan.StockMinimo, UsuarioCreador: strings.TrimSpace(usuario), Estado: "activo",
		Observaciones: "Creado mediante propuesta IA confirmada",
	}, plan.StockInicial, "IA_PRODUCT_CREATE")
}

func NormalizeEmpresaAIHotelRoomPlan(plan *EmpresaAIHotelRoomPlan) error {
	if plan == nil || plan.EstacionID <= 0 {
		return fmt.Errorf("estacion_id es obligatorio")
	}
	plan.NombreHabitacion = strings.TrimSpace(plan.NombreHabitacion)
	if plan.NombreHabitacion == "" || len([]rune(plan.NombreHabitacion)) > 120 {
		return fmt.Errorf("nombre de habitacion invalido")
	}
	plan.Moneda = normalizeTarifaPorDiaMoneda(plan.Moneda)
	if len(plan.Moneda) != 3 {
		return fmt.Errorf("moneda invalida")
	}
	var err error
	if plan.HoraCheckIn, err = normalizeTarifaPorDiaHora(plan.HoraCheckIn, "15:00"); err != nil {
		return err
	}
	if plan.HoraCheckOut, err = normalizeTarifaPorDiaHora(plan.HoraCheckOut, "12:00"); err != nil {
		return err
	}
	if len(plan.Tarifas) == 0 || len(plan.Tarifas) > 12 {
		return fmt.Errorf("debe indicar entre una y doce tarifas")
	}
	seen := map[int]bool{}
	for i := range plan.Tarifas {
		rate := &plan.Tarifas[i]
		if rate.Personas <= 0 || rate.Personas > 999 || rate.Valor <= 0 || rate.Valor > 1000000000 || seen[rate.Personas] {
			return fmt.Errorf("tarifa por ocupacion invalida")
		}
		seen[rate.Personas] = true
	}
	return nil
}

// ConfigureEmpresaAIHotelRoomStation performs the configuration and tariff creation in one database transaction.
func ConfigureEmpresaAIHotelRoomStation(dbConn *sql.DB, empresaID int64, plan EmpresaAIHotelRoomPlan, usuario string) ([]int64, error) {
	if empresaID <= 0 || strings.TrimSpace(usuario) == "" {
		return nil, fmt.Errorf("contexto empresarial invalido")
	}
	if err := NormalizeEmpresaAIHotelRoomPlan(&plan); err != nil {
		return nil, err
	}
	if err := EnsureEmpresaEstacionPrefsSchema(dbConn); err != nil {
		return nil, err
	}
	if err := EnsureEmpresaTarifasPorDiaSchema(dbConn); err != nil {
		return nil, err
	}
	pref, err := GetEmpresaEstacionPref(dbConn, empresaID, 0, "estaciones_config")
	if err != nil {
		return nil, err
	}
	if pref == nil {
		return nil, fmt.Errorf("no existe configuracion de estaciones para la empresa")
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(pref.Valor), &cfg); err != nil {
		return nil, fmt.Errorf("configuracion de estaciones invalida")
	}
	items, ok := cfg["estaciones"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no se encontro la estacion solicitada")
	}
	stationFound := false
	stationCode := fmt.Sprintf("ESTACION-%03d", plan.EstacionID)
	for _, raw := range items {
		station, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		id, _ := station["id"].(float64)
		if int64(id) != plan.EstacionID {
			continue
		}
		stationFound = true
		station["nombre"] = plan.NombreHabitacion
		station["tipo_estacion"] = "hotel"
		station["activa"] = plan.Activa
		station["moneda"] = plan.Moneda
		stationCode = strings.TrimSpace(fmt.Sprint(station["codigo"]))
		if stationCode == "" || stationCode == "<nil>" {
			stationCode = fmt.Sprintf("ESTACION-%03d", plan.EstacionID)
			station["codigo"] = stationCode
		}
		station["hotel"] = map[string]interface{}{"tarifa_tipo": "por_dia", "hora_check_in": plan.HoraCheckIn, "hora_check_out": plan.HoraCheckOut}
		break
	}
	if !stationFound {
		return nil, fmt.Errorf("la estacion no pertenece a la empresa")
	}
	cfgJSON, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err := execTxSQLCompat(tx, `UPDATE empresa_estacion_prefs SET valor=?, usuario_creador=?, fecha_actualizacion=CURRENT_TIMESTAMP WHERE empresa_id=? AND estacion_id=0 AND clave='estaciones_config'`, string(cfgJSON), strings.TrimSpace(usuario), empresaID); err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(plan.Tarifas))
	for _, rate := range plan.Tarifas {
		payload := EmpresaTarifaPorDia{EmpresaID: empresaID, NombreTarifa: fmt.Sprintf("%s - %d personas", plan.NombreHabitacion, rate.Personas), EstacionID: plan.EstacionID, EstacionCodigo: stationCode, EstacionNombre: plan.NombreHabitacion, ServicioNombre: "hospedaje", ValorDia: rate.Valor, PersonasDesde: rate.Personas, PersonasHasta: rate.Personas, HoraCheckIn: plan.HoraCheckIn, HoraCheckOut: plan.HoraCheckOut, Moneda: plan.Moneda, Prioridad: rate.Personas, AplicarAutomaticamente: true, UsuarioCreador: strings.TrimSpace(usuario), Estado: "activo", Observaciones: "Configurada mediante propuesta IA confirmada"}
		if err := normalizeEmpresaTarifaPorDiaPayload(&payload); err != nil {
			return nil, err
		}
		apply := 1
		id, err := insertTxSQLCompat(tx, `INSERT INTO empresa_tarifas_por_dia (empresa_id,nombre_tarifa,estacion_id,estacion_codigo,estacion_nombre,servicio_nombre,valor_dia,personas_desde,personas_hasta,hora_check_in,hora_check_out,moneda,prioridad,aplicar_automaticamente,usuario_creador,estado,observaciones,fecha_creacion,fecha_actualizacion) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, payload.EmpresaID, payload.NombreTarifa, payload.EstacionID, payload.EstacionCodigo, payload.EstacionNombre, payload.ServicioNombre, payload.ValorDia, payload.PersonasDesde, payload.PersonasHasta, payload.HoraCheckIn, payload.HoraCheckOut, payload.Moneda, payload.Prioridad, apply, payload.UsuarioCreador, payload.Estado, payload.Observaciones)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return ids, nil
}
