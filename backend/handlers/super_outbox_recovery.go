package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	superOutboxRecoveryConfirmation = "REACTIVAR EVENTOS CXP"
)

type superOutboxRecoveryRequest struct {
	EmpresaID    int64   `json:"empresa_id"`
	Topic        string  `json:"topic"`
	EventIDs     []int64 `json:"event_ids"`
	Reason       string  `json:"reason"`
	Confirmation string  `json:"confirmation"`
}

type superOutboxRecoveryItem struct {
	ID                   int64   `json:"id"`
	EmpresaID            int64   `json:"empresa_id"`
	Topic                string  `json:"topic"`
	Status               string  `json:"status"`
	Attempts             int     `json:"attempts"`
	MaxAttempts          int     `json:"max_attempts"`
	AvailableAt          string  `json:"available_at"`
	DeadAt               string  `json:"dead_at"`
	CreatedAt            string  `json:"created_at"`
	LastError            string  `json:"last_error"`
	CuentaPorPagarID     int64   `json:"cuenta_por_pagar_id,omitempty"`
	PagoID               int64   `json:"pago_id,omitempty"`
	MovimientoFinanzasID int64   `json:"movimiento_finanzas_id,omitempty"`
	Monto                float64 `json:"monto,omitempty"`
	RecoveryAction       string  `json:"recovery_action,omitempty"`
}

type superCxPPaymentOutboxPayload struct {
	CuentaPorPagarID     int64   `json:"cuenta_por_pagar_id"`
	PagoID               int64   `json:"pago_id"`
	MovimientoFinanzasID int64   `json:"movimiento_finanzas_id"`
	Monto                float64 `json:"monto"`
}

// SuperOutboxRecoveryHandler previews and explicitly requeues only the enabled
// CxP topic. Authentication, tenant resolution and the exact event IDs are
// mandatory; raw payloads are never returned to the browser.
func SuperOutboxRecoveryHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminEmail, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper)
		if !ok {
			return
		}
		if dbEmp == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{"ok": false, "error": "base empresarial no disponible"})
			return
		}
		if err := dbpkg.VerifyOutboxSchema(dbEmp); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{"ok": false, "error": "outbox no disponible; ejecute la migracion aprobada"})
			return
		}
		if err := dbpkg.VerifyOutboxRecoveryAuditSchema(dbEmp); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{"ok": false, "error": "auditoria de recuperacion no disponible; ejecute la migracion aprobada"})
			return
		}

		switch r.Method {
		case http.MethodGet:
			handleSuperOutboxRecoveryPreview(w, r, dbEmp)
		case http.MethodPost:
			handleSuperOutboxRecoveryExecute(w, r, dbEmp, adminEmail)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"ok": false, "error": "metodo no permitido"})
		}
	}
}

func handleSuperOutboxRecoveryPreview(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("empresa_id")), 10, 64)
	if err != nil || empresaID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "empresa_id es obligatorio"})
		return
	}
	topic := strings.TrimSpace(r.URL.Query().Get("topic"))
	if topic != dbpkg.EmpresaCxPPaymentOutboxTopic {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "topic no permitido para recuperacion"})
		return
	}
	eventIDs, err := parseSuperOutboxEventIDs(r.URL.Query().Get("event_ids"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	limit := dbpkg.MaxOutboxRecoveryEvents
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		parsed, parseErr := strconv.Atoi(raw)
		if parseErr != nil || parsed < 1 || parsed > dbpkg.MaxOutboxRecoveryEvents {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": fmt.Sprintf("limit debe estar entre 1 y %d", dbpkg.MaxOutboxRecoveryEvents)})
			return
		}
		limit = parsed
	}
	empresa, err := dbpkg.GetEmpresaByScopeID(dbEmp, empresaID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "no se pudo validar la empresa"})
		return
	}
	if empresa == nil || empresa.EmpresaID <= 0 {
		writeJSON(w, http.StatusNotFound, map[string]interface{}{"ok": false, "error": "empresa no encontrada"})
		return
	}

	items, err := dbpkg.ListDeadOutboxEventsForRecovery(dbEmp, empresa.EmpresaID, topic, eventIDs, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "no se pudo consultar la cola de recuperacion"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":             true,
		"empresa_id":     empresa.EmpresaID,
		"empresa_nombre": strings.TrimSpace(empresa.Nombre),
		"topic":          topic,
		"confirmation":   superOutboxRecoveryConfirmation,
		"items":          buildSuperOutboxRecoveryItems(items),
		"total":          len(items),
	})
}

func handleSuperOutboxRecoveryExecute(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, adminEmail string) {
	var request superOutboxRecoveryRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "payload invalido"})
		return
	}
	if strings.TrimSpace(request.Topic) != dbpkg.EmpresaCxPPaymentOutboxTopic {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "topic no permitido para recuperacion"})
		return
	}
	if strings.TrimSpace(request.Confirmation) != superOutboxRecoveryConfirmation {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "confirmacion de recuperacion incorrecta"})
		return
	}
	if request.EmpresaID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "empresa_id es obligatorio"})
		return
	}
	if err := validateSuperOutboxRecoveryRequest(request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": err.Error()})
		return
	}
	empresa, err := dbpkg.GetEmpresaByScopeID(dbEmp, request.EmpresaID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "no se pudo validar la empresa"})
		return
	}
	if empresa == nil || empresa.EmpresaID <= 0 {
		writeJSON(w, http.StatusNotFound, map[string]interface{}{"ok": false, "error": "empresa no encontrada"})
		return
	}

	items, err := dbpkg.RequeueDeadOutboxEvents(
		dbEmp,
		empresa.EmpresaID,
		dbpkg.EmpresaCxPPaymentOutboxTopic,
		request.EventIDs,
		request.Reason,
		adminEmail,
	)
	if err != nil {
		log.Printf("[outbox_recovery] operacion rechazada empresa=%d eventos=%d error=%v", empresa.EmpresaID, len(request.EventIDs), err)
		writeJSON(w, http.StatusConflict, map[string]interface{}{"ok": false, "error": "uno o mas eventos ya no estan disponibles dentro del alcance solicitado"})
		return
	}
	recordSuperOutboxEmpresaAudit(dbEmp, r, empresa.EmpresaID, adminEmail, request.Reason, items)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresa.EmpresaID,
		"topic":      dbpkg.EmpresaCxPPaymentOutboxTopic,
		"items":      buildSuperOutboxRecoveryItems(items),
		"requeued":   len(items),
	})
}

func validateSuperOutboxRecoveryRequest(request superOutboxRecoveryRequest) error {
	if len(request.EventIDs) == 0 || len(request.EventIDs) > dbpkg.MaxOutboxRecoveryEvents {
		return fmt.Errorf("seleccione entre 1 y %d eventos", dbpkg.MaxOutboxRecoveryEvents)
	}
	seen := make(map[int64]struct{}, len(request.EventIDs))
	for _, id := range request.EventIDs {
		if id <= 0 {
			return fmt.Errorf("event_ids contiene un id invalido")
		}
		if _, exists := seen[id]; exists {
			return fmt.Errorf("event_ids no puede contener duplicados")
		}
		seen[id] = struct{}{}
	}
	reason := strings.Join(strings.Fields(strings.TrimSpace(request.Reason)), " ")
	if len(reason) < 10 || len(reason) > 500 {
		return fmt.Errorf("la razon debe tener entre 10 y 500 caracteres")
	}
	return nil
}

func parseSuperOutboxEventIDs(raw string) ([]int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	if len(parts) > dbpkg.MaxOutboxRecoveryEvents {
		return nil, fmt.Errorf("maximo %d ids por consulta", dbpkg.MaxOutboxRecoveryEvents)
	}
	ids := make([]int64, 0, len(parts))
	for _, part := range parts {
		id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("event_ids contiene un id invalido")
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func buildSuperOutboxRecoveryItems(events []dbpkg.OutboxRecoveryEvent) []superOutboxRecoveryItem {
	items := make([]superOutboxRecoveryItem, 0, len(events))
	for _, event := range events {
		item := superOutboxRecoveryItem{
			ID:             event.ID,
			EmpresaID:      event.EmpresaID,
			Topic:          event.Topic,
			Status:         event.Status,
			Attempts:       event.Attempts,
			MaxAttempts:    event.MaxAttempts,
			AvailableAt:    event.AvailableAt,
			DeadAt:         event.DeadAt,
			CreatedAt:      event.CreatedAt,
			LastError:      sanitizeSuperAuditoriaString(event.LastError, 300),
			RecoveryAction: event.RecoveryAction,
		}
		var payload superCxPPaymentOutboxPayload
		if json.Unmarshal([]byte(event.PayloadJSON), &payload) == nil {
			item.CuentaPorPagarID = payload.CuentaPorPagarID
			item.PagoID = payload.PagoID
			item.MovimientoFinanzasID = payload.MovimientoFinanzasID
			item.Monto = payload.Monto
		}
		items = append(items, item)
	}
	return items
}

func recordSuperOutboxEmpresaAudit(dbEmp *sql.DB, r *http.Request, empresaID int64, adminEmail, reason string, events []dbpkg.OutboxRecoveryEvent) {
	ids := make([]int64, 0, len(events))
	for _, event := range events {
		ids = append(ids, event.ID)
	}
	metadata, err := json.Marshal(map[string]interface{}{
		"event_ids": ids,
		"topic":     dbpkg.EmpresaCxPPaymentOutboxTopic,
		"count":     len(ids),
	})
	if err != nil {
		metadata = []byte(`{}`)
	}
	if _, err := dbpkg.CreateEmpresaAuditoriaEvento(dbEmp, dbpkg.EmpresaAuditoriaEvento{
		EmpresaID:      empresaID,
		Modulo:         "infraestructura",
		Accion:         "reactivar_outbox_dead",
		Recurso:        "pcs_outbox_events",
		MetodoHTTP:     http.MethodPost,
		Endpoint:       r.URL.Path,
		Resultado:      "ok",
		CodigoHTTP:     http.StatusOK,
		RequestID:      resolveAuditoriaRequestID(r),
		IPOrigen:       resolveAuditoriaIP(r),
		UserAgent:      r.UserAgent(),
		MetadataJSON:   string(metadata),
		UsuarioCreador: adminEmail,
		Observaciones:  sanitizeSuperAuditoriaString(reason, 500),
	}); err != nil {
		log.Printf("[outbox_recovery] no se pudo registrar auditoria empresarial empresa=%d eventos=%d error=%v", empresaID, len(ids), err)
	}
}
