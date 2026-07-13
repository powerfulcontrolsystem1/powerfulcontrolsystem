package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	aipkg "github.com/you/pos-backend/ai"
	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaAIEnterpriseHandler exposes a closed, server-owned approval flow.
// It is feature-flagged off until a company validates the new UX.
func EmpresaAIEnterpriseHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !enterpriseAIEnabled() {
			writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{"ok": false, "code": "ai_enterprise_disabled", "error": "El orquestador empresarial IA esta desactivado mientras se valida."})
			return
		}
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		user := strings.ToLower(strings.TrimSpace(adminEmailFromRequest(r)))
		if user == "" || user == "sistema" {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		ctx := enterpriseAIExecutionContext(r, empresaID, user)
		if err := ctx.Validate(); err != nil {
			http.Error(w, "contexto IA invalido", http.StatusForbidden)
			return
		}
		if !aipkg.AllowsAgentMode(enterpriseAIAgentModeEnabled(), ctx) {
			writeJSON(w, http.StatusForbidden, map[string]interface{}{"ok": false, "code": "ai_agent_mode_disabled", "error": "El modo agente no esta habilitado para esta empresa."})
			return
		}
		if _, err := dbpkg.EnsureEmpresaAIConversation(dbEmp, dbpkg.EmpresaAIConversation{ConversationID: ctx.ConversationID, EmpresaID: ctx.EmpresaID, UsuarioID: ctx.UserID, Modo: ctx.Mode}, 2*time.Hour); err != nil {
			http.Error(w, "No se pudo preparar la conversacion IA", http.StatusServiceUnavailable)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch r.Method {
		case http.MethodGet:
			switch action {
			case "", "tools":
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "mode": ctx.Mode, "tools": aipkg.Registry(), "write_tools_enabled": enterpriseAIWriteEnabled()})
			case "proposal":
				proposalID := strings.TrimSpace(r.URL.Query().Get("proposal_id"))
				p, err := dbpkg.GetEmpresaAIProposal(dbEmp, empresaID, proposalID)
				if err != nil || p.UsuarioCreador != user {
					http.Error(w, "propuesta no encontrada", http.StatusNotFound)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "proposal": p})
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
			}
		case http.MethodPost:
			switch action {
			case "hotel_room_proposal":
				enterpriseAIHotelProposal(w, r, dbEmp, ctx)
			case "confirm":
				enterpriseAIConfirmProposal(w, r, dbEmp, ctx)
			case "cancel":
				enterpriseAICancelProposal(w, r, dbEmp, ctx)
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
			}
		default:
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
		}
	}
}

func enterpriseAIEnabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("AI_ENTERPRISE_ORCHESTRATOR_ENABLED")), "true")
}
func enterpriseAIWriteEnabled() bool {
	return enterpriseAIEnabled() && strings.EqualFold(strings.TrimSpace(os.Getenv("AI_WRITE_TOOLS_ENABLED")), "true") && strings.EqualFold(strings.TrimSpace(os.Getenv("AI_HOTEL_TOOLS_ENABLED")), "true")
}
func enterpriseAIAgentModeEnabled() bool {
	return enterpriseAIEnabled() && strings.EqualFold(strings.TrimSpace(os.Getenv("AI_AGENT_MODE_ENABLED")), "true")
}

func enterpriseAIExecutionContext(r *http.Request, empresaID int64, user string) aipkg.ExecutionContext {
	role, _ := r.Context().Value("adminRoleEfectivo").(string)
	if role == "" {
		role, _ = r.Context().Value("adminRole").(string)
	}
	requestID := resolveAuditoriaRequestID(r)
	conversationID := strings.TrimSpace(r.Header.Get("X-AI-Conversation-ID"))
	if conversationID == "" {
		conversationID = "conversation-" + requestID
	}
	mode := strings.ToLower(strings.TrimSpace(r.Header.Get("X-AI-Mode")))
	if mode == "" {
		mode = aipkg.ModeAssisted
	}
	return aipkg.ExecutionContext{UserID: user, EmpresaID: empresaID, Role: role, SessionID: strings.TrimSpace(r.Header.Get("X-Session-ID")), ConversationID: conversationID, RequestID: requestID, Mode: mode, AuthorizedScope: []string{"current_company"}, MaxOperations: 1}
}

func decodeEnterpriseJSON(w http.ResponseWriter, r *http.Request, dst interface{}, maxBytes int64) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return http.ErrNotSupported
	}
	return nil
}

type enterpriseAIHotelProposalRequest struct {
	ConversationID string                       `json:"conversation_id"`
	Plan           dbpkg.EmpresaAIHotelRoomPlan `json:"plan"`
}

func enterpriseAIHotelProposal(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, ctx aipkg.ExecutionContext) {
	if !enterpriseAIWriteEnabled() {
		writeJSON(w, http.StatusForbidden, map[string]interface{}{"ok": false, "code": "ai_write_tools_disabled", "error": "Las herramientas de escritura IA permanecen desactivadas."})
		return
	}
	var req enterpriseAIHotelProposalRequest
	if err := decodeEnterpriseJSON(w, r, &req, 128<<10); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if suppliedConversation := strings.TrimSpace(req.ConversationID); suppliedConversation != "" && suppliedConversation != ctx.ConversationID {
		http.Error(w, "conversacion IA no coincide con el contexto autenticado", http.StatusForbidden)
		return
	}
	if err := dbpkg.NormalizeEmpresaAIHotelRoomPlan(&req.Plan); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{"ok": false, "state": dbpkg.AIProposalAwaitingInformation, "missing_or_invalid": err.Error(), "questions": []string{"Indica si la tarifa es por noche, los horarios de check-in/check-out, moneda, activacion y las tarifas por ocupacion."}})
		return
	}
	// Current state is read on the server and is never selected by the model or frontend.
	current, err := dbpkg.GetEmpresaAIHotelRoomStationSnapshot(dbEmp, ctx.EmpresaID, req.Plan.EstacionID)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{"ok": false, "state": dbpkg.AIProposalAwaitingInformation, "missing_or_invalid": "No fue posible validar la estacion de esta empresa."})
		return
	}
	planJSON, _ := json.Marshal(req.Plan)
	beforeJSON, _ := json.Marshal(current)
	expectedJSON, _ := json.Marshal(map[string]interface{}{"tipo_estacion": "hotel", "nombre": req.Plan.NombreHabitacion, "tarifas": req.Plan.Tarifas})
	proposalID, err := aipkg.NewOpaqueID("proposal")
	if err != nil {
		http.Error(w, "No se pudo crear propuesta", http.StatusInternalServerError)
		return
	}
	p, err := dbpkg.CreateEmpresaAIProposal(dbEmp, dbpkg.EmpresaAIProposal{ProposalID: proposalID, ConversationID: ctx.ConversationID, EmpresaID: ctx.EmpresaID, UsuarioCreador: ctx.UserID, ToolName: aipkg.ToolHotelConfigureRoomStation, RiskLevel: "medium", PlanJSON: string(planJSON), EstadoAnterior: string(beforeJSON), EstadoEsperado: string(expectedJSON), Resumen: "Configurar " + req.Plan.NombreHabitacion + " como habitacion hotelera y registrar tarifas por ocupacion.", RollbackPolicy: "transactional_before_commit", Estado: dbpkg.AIProposalAwaitingConfirmation}, 15*time.Minute)
	if err != nil {
		http.Error(w, "No se pudo guardar propuesta", http.StatusInternalServerError)
		return
	}
	registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, ctx.EmpresaID, "centro_ia_empresarial", "propuesta_hotel_creada", "empresa_ai_propuestas", 0, http.StatusCreated, map[string]interface{}{"proposal_id": p.ProposalID, "tool": p.ToolName, "risk": p.RiskLevel}, "propuesta IA creada sin ejecutar cambios")
	writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "proposal": p, "sources": []string{"Configuracion actual de estaciones", "Reglas de tarifas por dia"}})
}

func enterpriseAIConfirmProposal(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, ctx aipkg.ExecutionContext) {
	if !enterpriseAIWriteEnabled() {
		http.Error(w, "herramientas IA de escritura desactivadas", http.StatusForbidden)
		return
	}
	var req struct {
		ProposalID     string `json:"proposal_id"`
		PlanHash       string `json:"plan_hash"`
		IdempotencyKey string `json:"idempotency_key"`
	}
	if err := decodeEnterpriseJSON(w, r, &req, 64<<10); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	p, err := dbpkg.BeginEmpresaAIProposalExecution(dbEmp, ctx.EmpresaID, strings.TrimSpace(req.ProposalID), ctx.UserID, strings.TrimSpace(req.PlanHash), strings.TrimSpace(req.IdempotencyKey))
	if err != nil {
		http.Error(w, "No se pudo confirmar la propuesta", http.StatusConflict)
		return
	}
	if p.ToolName != aipkg.ToolHotelConfigureRoomStation {
		_ = dbpkg.FinishEmpresaAIProposal(dbEmp, ctx.EmpresaID, p.ProposalID, dbpkg.AIProposalFailed, `{"error":"herramienta no habilitada"}`)
		http.Error(w, "herramienta no habilitada", http.StatusBadRequest)
		return
	}
	var plan dbpkg.EmpresaAIHotelRoomPlan
	if err := json.Unmarshal([]byte(p.PlanJSON), &plan); err != nil {
		_ = dbpkg.FinishEmpresaAIProposal(dbEmp, ctx.EmpresaID, p.ProposalID, dbpkg.AIProposalFailed, `{"error":"plan invalido"}`)
		http.Error(w, "plan invalido", http.StatusBadRequest)
		return
	}
	ids, err := dbpkg.ConfigureEmpresaAIHotelRoomStation(dbEmp, ctx.EmpresaID, plan, ctx.UserID)
	if err != nil {
		_ = dbpkg.FinishEmpresaAIProposal(dbEmp, ctx.EmpresaID, p.ProposalID, dbpkg.AIProposalFailed, `{"error":"ejecucion rechazada"}`)
		http.Error(w, "No se pudo aplicar la configuracion", http.StatusConflict)
		return
	}
	result, _ := json.Marshal(map[string]interface{}{"tarifa_ids": ids, "estacion_id": plan.EstacionID, "verified": true})
	if err := dbpkg.FinishEmpresaAIProposal(dbEmp, ctx.EmpresaID, p.ProposalID, dbpkg.AIProposalCompleted, string(result)); err != nil {
		http.Error(w, "La configuracion se aplico pero no se pudo cerrar la propuesta", http.StatusInternalServerError)
		return
	}
	registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, ctx.EmpresaID, "centro_ia_empresarial", "propuesta_hotel_ejecutada", "empresa_ai_propuestas", 0, http.StatusOK, map[string]interface{}{"proposal_id": p.ProposalID, "tool": p.ToolName, "tarifas_creadas": len(ids)}, "configuracion hotelera IA confirmada y ejecutada")
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "proposal_id": p.ProposalID, "status": dbpkg.AIProposalCompleted, "result": json.RawMessage(result)})
}

func enterpriseAICancelProposal(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, ctx aipkg.ExecutionContext) {
	var req struct {
		ProposalID string `json:"proposal_id"`
	}
	if err := decodeEnterpriseJSON(w, r, &req, 32<<10); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if err := dbpkg.CancelEmpresaAIProposal(dbEmp, ctx.EmpresaID, strings.TrimSpace(req.ProposalID), ctx.UserID); err != nil {
		http.Error(w, "No se pudo cancelar la propuesta", http.StatusConflict)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "status": dbpkg.AIProposalCancelled})
}
