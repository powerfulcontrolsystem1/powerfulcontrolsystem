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

func EnsureEmpresaAIEnterpriseSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	stmts := []string{
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
