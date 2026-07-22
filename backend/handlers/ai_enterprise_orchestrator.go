package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	aipkg "github.com/you/pos-backend/ai"
	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaAIEnterpriseHandler exposes a closed, server-owned approval flow.
// It is feature-flagged off until a company validates the new UX.
func EmpresaAIEnterpriseHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
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
		ctx, err := enterpriseAIExecutionContext(r, dbEmp, dbSuper, empresaID, user)
		if err != nil {
			http.Error(w, "No se pudo validar permisos de la herramienta IA", http.StatusForbidden)
			return
		}
		if err := ctx.Validate(); err != nil {
			http.Error(w, "contexto IA invalido", http.StatusForbidden)
			return
		}
		if !aipkg.AllowsAgentMode(enterpriseAIAgentModeEnabled(), ctx) {
			writeJSON(w, http.StatusForbidden, map[string]interface{}{"ok": false, "code": "ai_agent_mode_disabled", "error": "El modo agente no esta habilitado para esta empresa."})
			return
		}
		if _, err := dbpkg.CreateOrRefreshEmpresaAIConversation(dbEmp, dbpkg.EmpresaAIConversation{ConversationID: ctx.ConversationID, EmpresaID: ctx.EmpresaID, UsuarioID: ctx.UserID, Modo: ctx.Mode}, 2*time.Hour); err != nil {
			http.Error(w, "No se pudo preparar la conversacion IA", http.StatusServiceUnavailable)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch r.Method {
		case http.MethodGet:
			switch action {
			case "", "tools":
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "mode": ctx.Mode, "tools": enterpriseAIAvailableTools(ctx), "write_tools_enabled": enterpriseAIWriteEnabled()})
			case "proposal":
				proposalID := strings.TrimSpace(r.URL.Query().Get("proposal_id"))
				p, err := dbpkg.GetEmpresaAIProposal(dbEmp, empresaID, proposalID)
				if err != nil || p.UsuarioCreador != user {
					http.Error(w, "propuesta no encontrada", http.StatusNotFound)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "proposal": p})
			case "hotel_room_snapshot":
				if !enterpriseAIRequireTool(ctx, aipkg.ToolHotelInspectRoomStation) {
					http.Error(w, "forbidden: rol sin permiso para la herramienta solicitada", http.StatusForbidden)
					return
				}
				stationID, err := parseOptionalInt64Query(r, "estacion_id")
				if err != nil || stationID <= 0 {
					http.Error(w, "estacion_id invalido", http.StatusBadRequest)
					return
				}
				snapshot, err := dbpkg.GetEmpresaAIHotelRoomStationSnapshot(dbEmp, empresaID, stationID)
				if err != nil {
					http.Error(w, "estacion no encontrada", http.StatusNotFound)
					return
				}
				_ = dbpkg.RecordEmpresaAIExecution(dbEmp, dbpkg.EmpresaAIExecution{EmpresaID: ctx.EmpresaID, UsuarioID: ctx.UserID, ConversationID: ctx.ConversationID, ToolName: aipkg.ToolHotelInspectRoomStation, Modo: aipkg.ModeConsultive, RiskLevel: "low", Resultado: "completed", FuentesJSON: `["Configuracion actual de estaciones","Tarifas por dia"]`, CategoriasJSON: `["internal"]`})
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "snapshot": snapshot, "sources": []string{"Configuracion actual de estaciones", "Tarifas por dia"}})
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
			}
		case http.MethodPost:
			switch action {
			case "hotel_room_proposal":
				enterpriseAIHotelProposal(w, r, dbEmp, ctx)
			case "catalog_search":
				enterpriseAICatalogSearch(w, r, dbEmp, ctx)
			case "product_create_proposal":
				enterpriseAIProductProposal(w, r, dbEmp, ctx)
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
	return enterpriseAIEnabled() && strings.EqualFold(strings.TrimSpace(os.Getenv("AI_WRITE_TOOLS_ENABLED")), "true")
}

func enterpriseAIWriteToolEnabled(tool string) bool {
	if !enterpriseAIWriteEnabled() {
		return false
	}
	switch tool {
	case aipkg.ToolHotelConfigureRoomStation:
		return strings.EqualFold(strings.TrimSpace(os.Getenv("AI_HOTEL_TOOLS_ENABLED")), "true")
	case aipkg.ToolCatalogCreateProduct:
		return strings.EqualFold(strings.TrimSpace(os.Getenv("AI_CATALOG_TOOLS_ENABLED")), "true")
	default:
		return false
	}
}
func enterpriseAIAgentModeEnabled() bool {
	return enterpriseAIEnabled() && strings.EqualFold(strings.TrimSpace(os.Getenv("AI_AGENT_MODE_ENABLED")), "true")
}

func enterpriseAIExecutionContext(r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64, user string) (aipkg.ExecutionContext, error) {
	snapshot, err := getEmpresaPermissionSnapshot(dbEmp, dbSuper, user, empresaID)
	if err != nil || !snapshot.CanAccess {
		return aipkg.ExecutionContext{}, sql.ErrNoRows
	}
	role := snapshot.EffectiveRole
	requestID := resolveAuditoriaRequestID(r)
	conversationID := strings.TrimSpace(r.Header.Get("X-AI-Conversation-ID"))
	if conversationID == "" {
		conversationID = "conversation-" + requestID
	}
	mode := strings.ToLower(strings.TrimSpace(r.Header.Get("X-AI-Mode")))
	if mode == "" {
		mode = aipkg.ModeAssisted
	}
	permissions := make([]string, 0, len(snapshot.RoleModuleActions))
	for permission, allowed := range snapshot.RoleModuleActions {
		if allowed {
			permissions = append(permissions, permission)
		}
	}
	return aipkg.ExecutionContext{UserID: user, EmpresaID: empresaID, Role: role, Permissions: permissions, ConversationID: conversationID, RequestID: requestID, Mode: mode, AuthorizedScope: []string{"current_company"}, MaxOperations: 1}, nil
}

func enterpriseAIRequireTool(ctx aipkg.ExecutionContext, toolName string) bool {
	def, ok := aipkg.Registry()[toolName]
	return ok && aipkg.ToolAllowed(def, ctx.Permissions)
}

func enterpriseAIAvailableTools(ctx aipkg.ExecutionContext) map[string]aipkg.ToolDefinition {
	out := make(map[string]aipkg.ToolDefinition)
	for name, def := range aipkg.Registry() {
		if !aipkg.ToolAllowed(def, ctx.Permissions) {
			continue
		}
		if def.Confirmation == "required" && !enterpriseAIWriteToolEnabled(name) {
			continue
		}
		if def.Confirmation == "none" || enterpriseAIWriteToolEnabled(name) {
			out[name] = def
		}
	}
	return out
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

type enterpriseAIProductProposalRequest struct {
	ConversationID string                           `json:"conversation_id"`
	Plan           dbpkg.EmpresaAIProductCreatePlan `json:"plan"`
}

// enterpriseAICatalogSearch is a bounded, read-only catalog lookup for the
// agent. It intentionally returns identifiers only from the current company,
// so a later write proposal cannot be assembled with foreign IDs.
func enterpriseAICatalogSearch(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, ctx aipkg.ExecutionContext) {
	if !enterpriseAIRequireTool(ctx, aipkg.ToolCatalogSearchProducts) {
		http.Error(w, "forbidden: rol sin permiso para consultar catalogo", http.StatusForbidden)
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	products, err := dbpkg.GetProductosByEmpresa(dbEmp, ctx.EmpresaID, query, "activo", 0, 0, 20, 0)
	if err != nil {
		http.Error(w, "No se pudo consultar el catalogo", http.StatusServiceUnavailable)
		return
	}
	categories, err := dbpkg.GetCategoriasProductoByEmpresa(dbEmp, ctx.EmpresaID, false, query)
	if err != nil {
		http.Error(w, "No se pudo consultar categorias", http.StatusServiceUnavailable)
		return
	}
	warehouses, err := dbpkg.GetBodegasByEmpresa(dbEmp, ctx.EmpresaID, false)
	if err != nil {
		http.Error(w, "No se pudo consultar bodegas", http.StatusServiceUnavailable)
		return
	}
	_ = dbpkg.RecordEmpresaAIExecution(dbEmp, dbpkg.EmpresaAIExecution{EmpresaID: ctx.EmpresaID, UsuarioID: ctx.UserID, ConversationID: ctx.ConversationID, ToolName: aipkg.ToolCatalogSearchProducts, Modo: ctx.Mode, RiskLevel: "read", Resultado: "completed", FuentesJSON: `["Catalogo de productos","Categorias","Bodegas"]`, CategoriasJSON: `["internal"]`})
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "products": products, "categories": categories, "warehouses": warehouses, "sources": []string{"Catalogo de productos", "Categorias", "Bodegas"}})
}

func enterpriseAIProductProposal(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, ctx aipkg.ExecutionContext) {
	if !enterpriseAIWriteToolEnabled(aipkg.ToolCatalogCreateProduct) {
		writeJSON(w, http.StatusForbidden, map[string]interface{}{"ok": false, "code": "ai_write_tools_disabled", "error": "La creacion de productos mediante IA permanece desactivada."})
		return
	}
	if !enterpriseAIRequireTool(ctx, aipkg.ToolCatalogCreateProduct) {
		http.Error(w, "forbidden: rol sin permiso para crear productos", http.StatusForbidden)
		return
	}
	var req enterpriseAIProductProposalRequest
	if err := decodeEnterpriseJSON(w, r, &req, 128<<10); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if suppliedConversation := strings.TrimSpace(req.ConversationID); suppliedConversation != "" && suppliedConversation != ctx.ConversationID {
		http.Error(w, "conversacion IA no coincide con el contexto autenticado", http.StatusForbidden)
		return
	}
	if err := dbpkg.NormalizeEmpresaAIProductCreatePlan(&req.Plan); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{"ok": false, "state": dbpkg.AIProposalAwaitingInformation, "missing_or_invalid": enterpriseAIProposalValidationMessage(r, err)})
		return
	}
	duplicates, err := dbpkg.FindEmpresaAIProductDuplicates(dbEmp, ctx.EmpresaID, req.Plan.Nombre, req.Plan.SKU)
	if err != nil {
		http.Error(w, "No se pudo validar duplicados del catalogo", http.StatusServiceUnavailable)
		return
	}
	if len(duplicates) > 0 {
		writeJSON(w, http.StatusConflict, map[string]interface{}{"ok": false, "state": dbpkg.AIProposalAwaitingInformation, "code": "possible_duplicate", "matches": duplicates, "error": "Existe un producto con el mismo nombre o SKU. Revisa el catalogo antes de continuar."})
		return
	}
	planJSON, _ := json.Marshal(req.Plan)
	beforeJSON := `{"duplicates":0}`
	expectedJSON, _ := json.Marshal(map[string]interface{}{"nombre": req.Plan.Nombre, "precio": req.Plan.Precio, "impuesto_porcentaje": req.Plan.ImpuestoPorcentaje, "stock_inicial": req.Plan.StockInicial, "categoria_id": req.Plan.CategoriaID, "bodega_id": req.Plan.BodegaID})
	proposalID, err := aipkg.NewOpaqueID("proposal")
	if err != nil {
		http.Error(w, "No se pudo crear propuesta", http.StatusInternalServerError)
		return
	}
	p, err := dbpkg.CreateEmpresaAIProposal(dbEmp, dbpkg.EmpresaAIProposal{ProposalID: proposalID, ConversationID: ctx.ConversationID, EmpresaID: ctx.EmpresaID, UsuarioCreador: ctx.UserID, ToolName: aipkg.ToolCatalogCreateProduct, RiskLevel: "medium", PlanJSON: string(planJSON), EstadoAnterior: beforeJSON, EstadoEsperado: string(expectedJSON), Resumen: "Crear producto " + req.Plan.Nombre + " en el catalogo de la empresa actual.", RollbackPolicy: "transactional_before_commit", Estado: dbpkg.AIProposalAwaitingConfirmation}, 15*time.Minute)
	if err != nil {
		http.Error(w, "No se pudo guardar propuesta", http.StatusInternalServerError)
		return
	}
	_ = dbpkg.RecordEmpresaAIExecution(dbEmp, dbpkg.EmpresaAIExecution{EmpresaID: ctx.EmpresaID, UsuarioID: ctx.UserID, ConversationID: ctx.ConversationID, ProposalID: p.ProposalID, ToolName: p.ToolName, Modo: ctx.Mode, RiskLevel: p.RiskLevel, Resultado: "awaiting_confirmation", FuentesJSON: `["Catalogo de productos","Categorias","Bodegas"]`, CategoriasJSON: `["internal"]`})
	registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, ctx.EmpresaID, "centro_ia_empresarial", "propuesta_producto_creada", "empresa_ai_propuestas", 0, http.StatusCreated, map[string]interface{}{"proposal_id": p.ProposalID, "tool": p.ToolName, "risk": p.RiskLevel}, "propuesta IA de producto creada sin ejecutar cambios")
	writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "proposal": p, "sources": []string{"Catalogo de productos", "Categorias", "Bodegas"}})
}

func enterpriseAIHotelProposal(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, ctx aipkg.ExecutionContext) {
	if !enterpriseAIWriteToolEnabled(aipkg.ToolHotelConfigureRoomStation) {
		writeJSON(w, http.StatusForbidden, map[string]interface{}{"ok": false, "code": "ai_write_tools_disabled", "error": "Las herramientas de escritura IA permanecen desactivadas."})
		return
	}
	if !enterpriseAIRequireTool(ctx, aipkg.ToolHotelConfigureRoomStation) {
		http.Error(w, "forbidden: rol sin permiso para configurar habitaciones", http.StatusForbidden)
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
		writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{"ok": false, "state": dbpkg.AIProposalAwaitingInformation, "missing_or_invalid": enterpriseAIProposalValidationMessage(r, err), "questions": []string{"Indica si la tarifa es por noche, los horarios de check-in/check-out, moneda, activacion y las tarifas por ocupacion."}})
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
	_ = dbpkg.RecordEmpresaAIExecution(dbEmp, dbpkg.EmpresaAIExecution{EmpresaID: ctx.EmpresaID, UsuarioID: ctx.UserID, ConversationID: ctx.ConversationID, ProposalID: p.ProposalID, ToolName: p.ToolName, Modo: ctx.Mode, RiskLevel: p.RiskLevel, Resultado: "awaiting_confirmation", FuentesJSON: `["Configuracion actual de estaciones","Reglas de tarifas por dia"]`, CategoriasJSON: `["internal"]`})
	registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, ctx.EmpresaID, "centro_ia_empresarial", "propuesta_hotel_creada", "empresa_ai_propuestas", 0, http.StatusCreated, map[string]interface{}{"proposal_id": p.ProposalID, "tool": p.ToolName, "risk": p.RiskLevel}, "propuesta IA creada sin ejecutar cambios")
	writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "proposal": p, "sources": []string{"Configuracion actual de estaciones", "Reglas de tarifas por dia"}})
}

func enterpriseAIProposalValidationMessage(r *http.Request, err error) string {
	if err == nil {
		return "Los datos de la propuesta no son validos."
	}
	switch err.Error() {
	case "plan de producto invalido", "nombre de producto invalido", "texto de producto excede el limite permitido", "valores de producto invalidos", "bodega_id es obligatorio cuando se registra stock inicial", "estacion_id es obligatorio", "nombre de habitacion invalido", "moneda invalida", "hora invalida", "debe indicar entre una y doce tarifas", "tarifa por ocupacion invalida":
		return err.Error()
	default:
		log.Printf("[ai_enterprise] operation=proposal_validation request_id=%s error_type=%T", resolveAuditoriaRequestID(r), err)
		return "Los datos de la propuesta no son validos."
	}
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
	// Validate the current policy before consuming the one-time proposal. The
	// execution function below locks it and repeats the ownership/hash checks.
	preview, err := dbpkg.GetEmpresaAIProposal(dbEmp, ctx.EmpresaID, strings.TrimSpace(req.ProposalID))
	if err != nil || preview.UsuarioCreador != ctx.UserID {
		http.Error(w, "No se pudo confirmar la propuesta", http.StatusConflict)
		return
	}
	if !enterpriseAIRequireTool(ctx, preview.ToolName) || !enterpriseAIWriteToolEnabled(preview.ToolName) {
		http.Error(w, "La herramienta ya no esta disponible para este usuario", http.StatusForbidden)
		return
	}
	p, err := dbpkg.BeginEmpresaAIProposalExecution(dbEmp, ctx.EmpresaID, strings.TrimSpace(req.ProposalID), ctx.UserID, strings.TrimSpace(req.PlanHash), strings.TrimSpace(req.IdempotencyKey))
	if err != nil {
		http.Error(w, "No se pudo confirmar la propuesta", http.StatusConflict)
		return
	}
	if p.ToolName != aipkg.ToolHotelConfigureRoomStation {
		if p.ToolName != aipkg.ToolCatalogCreateProduct {
			_ = dbpkg.FinishEmpresaAIProposal(dbEmp, ctx.EmpresaID, p.ProposalID, dbpkg.AIProposalFailed, `{"error":"herramienta no habilitada"}`)
			http.Error(w, "herramienta no habilitada", http.StatusBadRequest)
			return
		}
	}
	if !enterpriseAIRequireTool(ctx, p.ToolName) || !enterpriseAIWriteToolEnabled(p.ToolName) {
		_ = dbpkg.FinishEmpresaAIProposal(dbEmp, ctx.EmpresaID, p.ProposalID, dbpkg.AIProposalFailed, `{"error":"permiso o herramienta no disponible"}`)
		http.Error(w, "La herramienta ya no esta disponible para este usuario", http.StatusForbidden)
		return
	}
	if p.ToolName == aipkg.ToolCatalogCreateProduct {
		enterpriseAIConfirmProductProposal(w, r, dbEmp, ctx, p)
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
	_ = dbpkg.RecordEmpresaAIExecution(dbEmp, dbpkg.EmpresaAIExecution{EmpresaID: ctx.EmpresaID, UsuarioID: ctx.UserID, ConversationID: p.ConversationID, ProposalID: p.ProposalID, ToolName: p.ToolName, Modo: ctx.Mode, RiskLevel: p.RiskLevel, Resultado: "completed", FuentesJSON: `["Configuracion actual de estaciones","Tarifas por dia"]`, CategoriasJSON: `["internal"]`})
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "proposal_id": p.ProposalID, "status": dbpkg.AIProposalCompleted, "result": json.RawMessage(result)})
}

func enterpriseAIConfirmProductProposal(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, ctx aipkg.ExecutionContext, p *dbpkg.EmpresaAIProposal) {
	var plan dbpkg.EmpresaAIProductCreatePlan
	if err := json.Unmarshal([]byte(p.PlanJSON), &plan); err != nil {
		_ = dbpkg.FinishEmpresaAIProposal(dbEmp, ctx.EmpresaID, p.ProposalID, dbpkg.AIProposalFailed, `{"error":"plan invalido"}`)
		http.Error(w, "plan invalido", http.StatusBadRequest)
		return
	}
	productID, err := dbpkg.CreateEmpresaAIProduct(dbEmp, ctx.EmpresaID, plan, ctx.UserID)
	if err != nil {
		_ = dbpkg.FinishEmpresaAIProposal(dbEmp, ctx.EmpresaID, p.ProposalID, dbpkg.AIProposalFailed, `{"error":"ejecucion rechazada"}`)
		writeJSON(w, http.StatusConflict, map[string]interface{}{"ok": false, "code": "product_create_rejected", "error": "No se pudo crear el producto. Revisa que no exista un duplicado y que categoria y bodega pertenezcan a la empresa."})
		return
	}
	result, _ := json.Marshal(map[string]interface{}{"product_id": productID, "verified": true})
	if err := dbpkg.FinishEmpresaAIProposal(dbEmp, ctx.EmpresaID, p.ProposalID, dbpkg.AIProposalCompleted, string(result)); err != nil {
		http.Error(w, "El producto se creo pero no se pudo cerrar la propuesta", http.StatusInternalServerError)
		return
	}
	registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, ctx.EmpresaID, "centro_ia_empresarial", "propuesta_producto_ejecutada", "productos", productID, http.StatusOK, map[string]interface{}{"proposal_id": p.ProposalID, "tool": p.ToolName, "product_id": productID}, "producto creado mediante propuesta IA confirmada")
	_ = dbpkg.RecordEmpresaAIExecution(dbEmp, dbpkg.EmpresaAIExecution{EmpresaID: ctx.EmpresaID, UsuarioID: ctx.UserID, ConversationID: p.ConversationID, ProposalID: p.ProposalID, ToolName: p.ToolName, Modo: ctx.Mode, RiskLevel: p.RiskLevel, Resultado: "completed", FuentesJSON: `["Catalogo de productos","Categorias","Bodegas"]`, CategoriasJSON: `["internal"]`})
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
