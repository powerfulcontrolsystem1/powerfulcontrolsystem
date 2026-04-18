package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/mail"
	"net"
	"net/http"
	"net/smtp"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

// TiposLicenciasHandler placeholder (removed from UI)
func TiposLicenciasHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "tipos_de_licencia API removed", http.StatusNotFound)
	}
}

// LicenciasHandler maneja CRUD de licencias
func LicenciasHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			q := r.URL.Query()
			parseTruthy := func(v string) bool {
				switch strings.ToLower(strings.TrimSpace(v)) {
				case "1", "true", "si", "yes", "activo":
					return true
				default:
					return false
				}
			}

			soloActivas := parseTruthy(q.Get("activo"))
			conEmpresa := parseTruthy(q.Get("con_empresa"))
			usuarioCreador := strings.TrimSpace(q.Get("usuario_creador"))

			// scope=mine permite filtrar por el administrador autenticado sin exponer email en la URL.
			if strings.EqualFold(strings.TrimSpace(q.Get("scope")), "mine") && usuarioCreador == "" {
				c, err := r.Cookie("session_token")
				if err != nil || c == nil || strings.TrimSpace(c.Value) == "" {
					http.Error(w, "unauthenticated", http.StatusUnauthorized)
					return
				}
				s, err := dbpkg.GetSessionByToken(dbSuper, c.Value)
				if err != nil || s == nil {
					http.Error(w, "unauthenticated", http.StatusUnauthorized)
					return
				}
				usuarioCreador = strings.TrimSpace(s.AdminEmail)
			}

			licencias, err := dbpkg.GetLicenciasFiltered(dbSuper, soloActivas, usuarioCreador, conEmpresa)
			if err != nil {
				log.Println("GET /super/api/licencias error:", err)
				http.Error(w, "failed to query licencias: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(licencias)
			return
		case http.MethodPost:
			var payload struct {
				TipoID       int64   `json:"tipo_id"`
				Nombre       string  `json:"nombre"`
				Descripcion  string  `json:"descripcion"`
				Valor        float64 `json:"valor"`
				DuracionDias int     `json:"duracion_dias"`
				ModulosHab   string  `json:"modulos_habilitados"`
				SuperRol     int     `json:"super_rol_habilitado"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}

			log.Printf("POST /super/api/licencias payload: TipoID=%d Nombre=%q", payload.TipoID, payload.Nombre)
			if payload.Nombre == "" {
				http.Error(w, "nombre required", http.StatusBadRequest)
				return
			}
			id, err := dbpkg.CreateLicencia(dbSuper, payload.TipoID, payload.Nombre, payload.Descripcion, payload.Valor, payload.DuracionDias, payload.ModulosHab, payload.SuperRol)
			if err != nil {
				log.Println("POST /super/api/licencias error:", err)
				http.Error(w, "failed to create licencia: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
			return
		case http.MethodPut:
			q := r.URL.Query()
			idStr := q.Get("id")
			if idStr == "" {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				http.Error(w, "invalid id", http.StatusBadRequest)
				return
			}
			// soporte para acción de activar/desactivar vía query param
			if q.Get("action") == "activar" {
				activoStr := q.Get("activo")
				if activoStr == "" {
					http.Error(w, "activo required (0 or 1)", http.StatusBadRequest)
					return
				}
				act, err := strconv.Atoi(activoStr)
				if err != nil || (act != 0 && act != 1) {
					http.Error(w, "invalid activo value", http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetLicenciaActivo(dbSuper, id, act); err != nil {
					log.Println("ACTIVAR /super/api/licencias error:", err)
					http.Error(w, "failed to set activo: "+err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
			// actualización normal (payload JSON)
			var payloadUpdate struct {
				TipoID       int64   `json:"tipo_id"`
				Nombre       string  `json:"nombre"`
				Descripcion  string  `json:"descripcion"`
				Valor        float64 `json:"valor"`
				DuracionDias int     `json:"duracion_dias"`
				ModulosHab   string  `json:"modulos_habilitados"`
				SuperRol     int     `json:"super_rol_habilitado"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payloadUpdate); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if err := dbpkg.UpdateLicencia(dbSuper, id, payloadUpdate.TipoID, payloadUpdate.Nombre, payloadUpdate.Descripcion, payloadUpdate.Valor, payloadUpdate.DuracionDias, payloadUpdate.ModulosHab, payloadUpdate.SuperRol); err != nil {
				log.Println("PUT /super/api/licencias error:", err)
				http.Error(w, "failed to update licencia: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		case http.MethodDelete:
			q := r.URL.Query()
			idStr := q.Get("id")
			if idStr == "" {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				http.Error(w, "invalid id", http.StatusBadRequest)
				return
			}
			if err := dbpkg.DeleteLicencia(dbSuper, id); err != nil {
				log.Println("DELETE /super/api/licencias error:", err)
				http.Error(w, "failed to delete licencia: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

func getDecryptedConfigValue(dbSuper *sql.DB, key string) (string, error) {
	v, enc, err := dbpkg.GetConfigValue(dbSuper, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	if v == "" {
		return "", nil
	}
	if !enc {
		return v, nil
	}
	dec, derr := utils.DecryptString(v)
	if derr != nil {
		return "", derr
	}
	return dec, nil
}

func isApprovedPaymentStatus(status string) bool {
	status = strings.ToLower(strings.TrimSpace(status))
	return status == "approved" || status == "accredited"
}

func parsePaymentPayloadMap(raw string) map[string]interface{} {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil
	}
	if len(payload) == 0 {
		return nil
	}
	return payload
}

func mergePaymentPayloadMaps(base, overlay map[string]interface{}) map[string]interface{} {
	if len(base) == 0 && len(overlay) == 0 {
		return nil
	}
	merged := make(map[string]interface{}, len(base)+len(overlay))
	for key, value := range base {
		merged[key] = value
	}
	for key, value := range overlay {
		srcMap, srcIsMap := value.(map[string]interface{})
		dstMap, dstIsMap := merged[key].(map[string]interface{})
		if srcIsMap && dstIsMap {
			merged[key] = mergePaymentPayloadMaps(dstMap, srcMap)
			continue
		}
		merged[key] = value
	}
	return merged
}

func mergePaymentPayloadJSON(existingRaw, updateRaw string) string {
	existingRaw = strings.TrimSpace(existingRaw)
	updateRaw = strings.TrimSpace(updateRaw)
	if existingRaw == "" {
		return updateRaw
	}
	if updateRaw == "" {
		return existingRaw
	}
	existingPayload := parsePaymentPayloadMap(existingRaw)
	updatePayload := parsePaymentPayloadMap(updateRaw)
	if len(existingPayload) == 0 || len(updatePayload) == 0 {
		return updateRaw
	}
	merged := mergePaymentPayloadMaps(existingPayload, updatePayload)
	if len(merged) == 0 {
		return updateRaw
	}
	mergedBytes, err := json.Marshal(merged)
	if err != nil {
		return updateRaw
	}
	return string(mergedBytes)
}

func parsePaymentTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02",
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func findEpaycoPaymentRecordByCandidates(dbSuper *sql.DB, transactionCandidates []string, referenceCandidates []string) (*dbpkg.EpaycoPaymentRecord, error) {
	seenTransactions := make(map[string]struct{}, len(transactionCandidates))
	for _, candidate := range transactionCandidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		if _, seen := seenTransactions[candidate]; seen {
			continue
		}
		seenTransactions[candidate] = struct{}{}
		rec, err := dbpkg.GetEpaycoPaymentByTransaction(dbSuper, candidate)
		if err != nil {
			return nil, err
		}
		if rec != nil {
			return rec, nil
		}
	}

	seenReferences := make(map[string]struct{}, len(referenceCandidates))
	for _, candidate := range referenceCandidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		if _, seen := seenReferences[candidate]; seen {
			continue
		}
		seenReferences[candidate] = struct{}{}
		rec, err := dbpkg.GetEpaycoPaymentByReference(dbSuper, candidate)
		if err != nil {
			return nil, err
		}
		if rec != nil {
			return rec, nil
		}
	}

	return nil, nil
}

func resolveEpaycoPaymentContextCandidates(dbSuper *sql.DB, lookupCombos [][2]string) (int64, int64, bool) {
	seenCombos := make(map[string]struct{}, len(lookupCombos))
	for _, combo := range lookupCombos {
		txCandidate := strings.TrimSpace(combo[0])
		refCandidate := strings.TrimSpace(combo[1])
		if txCandidate == "" && refCandidate == "" {
			continue
		}
		key := txCandidate + "|" + refCandidate
		if _, seen := seenCombos[key]; seen {
			continue
		}
		seenCombos[key] = struct{}{}

		licenciaID, empresaID, found, err := dbpkg.GetEpaycoPaymentContext(dbSuper, txCandidate, refCandidate)
		if err != nil {
			log.Println("warning: failed to resolve Epayco payment context:", err)
			continue
		}
		if found {
			return licenciaID, empresaID, true
		}
	}
	return 0, 0, false
}

func extractCustomerEmailFromPaymentPayload(payload map[string]interface{}) string {
	if len(payload) == 0 {
		return ""
	}
	if email := strings.TrimSpace(fmt.Sprint(payload["customer_email"])); email != "" && email != "<nil>" {
		return email
	}
	if email := strings.TrimSpace(fmt.Sprint(payload["email"])); email != "" && email != "<nil>" {
		return email
	}
	if billing, ok := payload["billing"].(map[string]interface{}); ok {
		if email := strings.TrimSpace(fmt.Sprint(billing["email"])); email != "" && email != "<nil>" {
			return email
		}
	}
	if data, ok := payload["data"].(map[string]interface{}); ok {
		if email := strings.TrimSpace(fmt.Sprint(data["customer_email"])); email != "" && email != "<nil>" {
			return email
		}
		if email := strings.TrimSpace(fmt.Sprint(data["email"])); email != "" && email != "<nil>" {
			return email
		}
		if billing, ok := data["billing"].(map[string]interface{}); ok {
			if email := strings.TrimSpace(fmt.Sprint(billing["email"])); email != "" && email != "<nil>" {
				return email
			}
		}
	}
	return ""
}

func resolveLicenciaActivationRecipient(dbSuper *sql.DB, empresaID int64, payRec *dbpkg.EpaycoPaymentRecord) (string, error) {
	if payRec == nil || !payRec.RawPayload.Valid {
		return "", fmt.Errorf("payload del pago no disponible para notificar activacion de licencia")
	}
	payload := parsePaymentPayloadMap(payRec.RawPayload.String)
	toEmail := strings.TrimSpace(extractCustomerEmailFromPaymentPayload(payload))
	if toEmail == "" && empresaID > 0 {
		empresa, err := dbpkg.GetEmpresaByScopeID(dbSuper, empresaID)
		if err == nil && empresa != nil {
			toEmail = strings.TrimSpace(empresa.UsuarioCreador)
		}
	}
	if toEmail == "" {
		return "", fmt.Errorf("correo del cliente no disponible para notificar activacion de licencia")
	}
	if _, err := mail.ParseAddress(toEmail); err != nil {
		return "", fmt.Errorf("correo del cliente invalido: %w", err)
	}
	return toEmail, nil
}

func epaycoActivationEmailAlreadySent(payRec *dbpkg.EpaycoPaymentRecord) bool {
	if payRec == nil || !payRec.RawPayload.Valid {
		return false
	}
	payload := parsePaymentPayloadMap(payRec.RawPayload.String)
	if len(payload) == 0 {
		return false
	}
	raw := strings.TrimSpace(fmt.Sprint(payload["licencia_activation_email_sent"]))
	switch strings.ToLower(raw) {
	case "1", "true", "si", "yes":
		return true
	default:
		return false
	}
}

func markEpaycoActivationEmailSent(dbSuper *sql.DB, payRec *dbpkg.EpaycoPaymentRecord, recipient, reference string) error {
	if dbSuper == nil || payRec == nil {
		return nil
	}
	status := strings.TrimSpace(payRec.Status.String)
	if status == "" {
		status = "APPROVED"
	}
	markerRawBytes, err := json.Marshal(map[string]interface{}{
		"licencia_activation_email_sent":    true,
		"licencia_activation_email_to":      strings.TrimSpace(recipient),
		"licencia_activation_email_sent_at": time.Now().Format(time.RFC3339),
		"licencia_activation_email_ref":     strings.TrimSpace(reference),
	})
	if err != nil {
		return err
	}
	mergedPayload := mergePaymentPayloadJSON(payRec.RawPayload.String, string(markerRawBytes))
	transactionID := strings.TrimSpace(payRec.TransactionID.String)
	recordReference := strings.TrimSpace(payRec.Reference.String)
	if transactionID != "" {
		if err := dbpkg.UpdateEpaycoPaymentRecordByTransaction(dbSuper, transactionID, status, mergedPayload); err != nil {
			return err
		}
	}
	if recordReference != "" {
		if err := dbpkg.UpdateEpaycoPaymentRecordByReference(dbSuper, recordReference, status, mergedPayload); err != nil {
			return err
		}
	}
	payRec.RawPayload = sql.NullString{String: mergedPayload, Valid: strings.TrimSpace(mergedPayload) != ""}
	return nil
}

func trySendLicenciaActivationEmail(r *http.Request, dbSuper *sql.DB, empresaID int64, lic *dbpkg.Licencia, payRec *dbpkg.EpaycoPaymentRecord, provider, reference string) error {
	if payRec == nil || lic == nil {
		return nil
	}
	if epaycoActivationEmailAlreadySent(payRec) {
		return nil
	}
	recipient, err := resolveLicenciaActivationRecipient(dbSuper, empresaID, payRec)
	if err != nil {
		return err
	}
	if err := sendLicenciaActivationEmail(r, dbSuper, empresaID, lic, payRec, provider, reference); err != nil {
		return err
	}
	if err := markEpaycoActivationEmailSent(dbSuper, payRec, recipient, reference); err != nil {
		return err
	}
	return nil
}

func buildLicenciaActivationMailBody(empresaNombre string, lic *dbpkg.Licencia, fechaInicio, fechaFin, provider, reference string) string {
	safeEmpresa := strings.TrimSpace(empresaNombre)
	if safeEmpresa == "" {
		safeEmpresa = "tu empresa"
	}
	safeProvider := strings.Title(strings.ToLower(strings.TrimSpace(provider)))
	if safeProvider == "" {
		safeProvider = "la pasarela de pago"
	}
	var builder strings.Builder
	builder.WriteString("Hola,\n\n")
	builder.WriteString("Tu pago fue confirmado correctamente y la licencia ya quedó activa en Powerful Control System.\n\n")
	builder.WriteString("Empresa: ")
	builder.WriteString(safeEmpresa)
	builder.WriteString("\n")
	if lic != nil {
		builder.WriteString("Licencia: ")
		builder.WriteString(strings.TrimSpace(lic.Nombre))
		builder.WriteString("\n")
	}
	if fechaInicio != "" {
		builder.WriteString("Fecha de inicio: ")
		builder.WriteString(fechaInicio)
		builder.WriteString("\n")
	}
	if fechaFin != "" {
		builder.WriteString("Fecha de vencimiento: ")
		builder.WriteString(fechaFin)
		builder.WriteString("\n")
	}
	if reference != "" {
		builder.WriteString("Referencia del pago: ")
		builder.WriteString(reference)
		builder.WriteString("\n")
	}
	builder.WriteString("Pasarela: ")
	builder.WriteString(safeProvider)
	builder.WriteString("\n\n")
	builder.WriteString("Ya puedes ingresar al sistema y continuar con la operación normal de tu empresa.\n\n")
	builder.WriteString("Si no reconoces este movimiento o necesitas ayuda, responde este correo.\n\n")
	builder.WriteString("Powerful Control System\n")
	return builder.String()
}

func sendLicenciaActivationEmail(r *http.Request, dbSuper *sql.DB, empresaID int64, lic *dbpkg.Licencia, payRec *dbpkg.EpaycoPaymentRecord, provider, reference string) error {
	if dbSuper == nil || lic == nil || payRec == nil || !payRec.RawPayload.Valid {
		return nil
	}
	toEmail, err := resolveLicenciaActivationRecipient(dbSuper, empresaID, payRec)
	if err != nil {
		return err
	}

	empresaNombre := ""
	if empresaID > 0 {
		empresa, err := dbpkg.GetEmpresaByScopeID(dbSuper, empresaID)
		if err == nil && empresa != nil {
			empresaNombre = strings.TrimSpace(empresa.Nombre)
		}
	}
	asunto := "Tu licencia ya quedó activa"
	cuerpo := buildLicenciaActivationMailBody(empresaNombre, lic, strings.TrimSpace(lic.FechaInicio), strings.TrimSpace(lic.FechaFin), provider, reference)
	metadataJSON := fmt.Sprintf(`{"provider":%q,"licencia_id":%d,"empresa_id":%d,"reference":%q}`, provider, lic.ID, empresaID, reference)

	if isEmpresaUsuarioMailTestMode(dbSuper) {
		return captureEmpresaUsuarioMailNotification(dbSuper, "licencia_activada_pago", empresaID, toEmail, asunto, cuerpo, reference, metadataJSON, adminEmailFromRequest(r))
	}

	smtpEmail, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_email")
	if err != nil {
		return err
	}
	smtpEmail = strings.TrimSpace(smtpEmail)
	if smtpEmail == "" {
		return fmt.Errorf("gmail.smtp_email no configurado")
	}
	smtpPass, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_app_password")
	if err != nil {
		return err
	}
	smtpPass = strings.TrimSpace(smtpPass)
	if smtpPass == "" {
		return fmt.Errorf("gmail.smtp_app_password no configurado")
	}
	smtpHost, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_host")
	smtpPort, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_port")
	fromName, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_from_name")
	smtpHost = strings.TrimSpace(smtpHost)
	smtpPort = strings.TrimSpace(smtpPort)
	fromName = strings.TrimSpace(fromName)
	if smtpHost == "" {
		smtpHost = "smtp.gmail.com"
	}
	if smtpPort == "" {
		smtpPort = "587"
	}
	if fromName == "" {
		fromName = "Powerful Control System"
	}

	mailHostForAuth := smtpHost
	if h, _, splitErr := net.SplitHostPort(smtpHost); splitErr == nil && strings.TrimSpace(h) != "" {
		mailHostForAuth = h
	}
	auth := smtp.PlainAuth("", smtpEmail, smtpPass, mailHostForAuth)
	addr := net.JoinHostPort(smtpHost, smtpPort)
	msg := "From: " + fromName + " <" + smtpEmail + ">\r\n" +
		"To: " + toEmail + "\r\n" +
		"Subject: " + asunto + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		cuerpo
	return smtp.SendMail(addr, auth, smtpEmail, []string{toEmail}, []byte(msg))
}

func activateLicenciaByIDs(dbSuper *sql.DB, licenciaID, empresaID int64) (bool, error) {
	if licenciaID <= 0 || empresaID <= 0 {
		return false, nil
	}
	lic, err := dbpkg.GetLicenciaByID(dbSuper, licenciaID)
	if err != nil {
		return false, err
	}
	if lic == nil {
		return false, nil
	}
	if lic.EmpresaID == empresaID && lic.Activo == 1 {
		if fechaFin, ok := parsePaymentTime(lic.FechaFin); ok && fechaFin.After(time.Now().Add(-1*time.Minute)) {
			return false, nil
		}
	}
	now := time.Now()
	fechaInicio := now.Format("2006-01-02 15:04:05")
	fechaFin := now.AddDate(0, 0, lic.DuracionDias).Format("2006-01-02 15:04:05")
	if err := dbpkg.ActivateLicenciaForEmpresa(dbSuper, licenciaID, empresaID, fechaInicio, fechaFin); err != nil {
		return false, err
	}
	lic.EmpresaID = empresaID
	lic.Activo = 1
	lic.FechaInicio = fechaInicio
	lic.FechaFin = fechaFin
	return true, nil
}

func extractWompiWebhookPaymentInfo(obj map[string]interface{}) (string, string, string) {
	get := func(v interface{}) string {
		s := strings.TrimSpace(fmt.Sprint(v))
		if s == "<nil>" {
			return ""
		}
		return s
	}

	var transactionID, reference, status string
	data, _ := obj["data"].(map[string]interface{})
	if tx, ok := data["transaction"].(map[string]interface{}); ok {
		transactionID = get(tx["id"])
		reference = get(tx["reference"])
		status = get(tx["status"])
	}
	if transactionID == "" {
		transactionID = get(data["id"])
	}
	if reference == "" {
		reference = get(data["reference"])
	}
	if status == "" {
		status = get(data["status"])
	}
	if transactionID == "" {
		transactionID = get(obj["transaction_id"])
	}
	if reference == "" {
		reference = get(obj["reference"])
	}
	if status == "" {
		status = get(obj["status"])
	}

	status = strings.ToUpper(strings.TrimSpace(status))
	return transactionID, reference, status
}

func parseSignatureCandidates(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	seen := map[string]struct{}{}
	out := make([]string, 0)
	add := func(v string) {
		v = strings.TrimSpace(v)
		v = strings.Trim(v, `"`)
		if v == "" {
			return
		}
		key := strings.ToLower(v)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, v)
	}

	add(raw)
	parts := strings.Split(raw, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "=") {
			split := strings.SplitN(part, "=", 2)
			add(split[1])
			continue
		}
		add(part)
	}

	return out
}

func signatureMatch(candidate, expected string) bool {
	left := []byte(strings.ToLower(strings.TrimSpace(candidate)))
	right := []byte(strings.ToLower(strings.TrimSpace(expected)))
	if len(left) == 0 || len(right) == 0 || len(left) != len(right) {
		return false
	}
	return subtle.ConstantTimeCompare(left, right) == 1
}

func verifyWompiWebhookSignature(dbSuper *sql.DB, r *http.Request, body []byte, obj map[string]interface{}) error {
	integrityKey, err := getDecryptedConfigValue(dbSuper, "wompi.integrity_key")
	if err != nil {
		return err
	}
	integrityKey = strings.TrimSpace(integrityKey)
	if integrityKey == "" {
		return nil
	}

	rawSignature := ""
	headerKeys := []string{"X-Wompi-Signature", "X-Event-Checksum", "X-Signature"}
	for _, hk := range headerKeys {
		if v := strings.TrimSpace(r.Header.Get(hk)); v != "" {
			rawSignature = v
			break
		}
	}
	if rawSignature == "" {
		if sigObj, ok := obj["signature"].(map[string]interface{}); ok {
			rawSignature = strings.TrimSpace(fmt.Sprint(sigObj["checksum"]))
			if rawSignature == "" || rawSignature == "<nil>" {
				rawSignature = strings.TrimSpace(fmt.Sprint(sigObj["signature"]))
			}
		}
	}
	if rawSignature == "" || rawSignature == "<nil>" {
		log.Println("warning: wompi webhook received without signature checksum; skipping strict validation")
		return nil
	}

	candidates := parseSignatureCandidates(rawSignature)
	if len(candidates) == 0 {
		return errors.New("invalid wompi signature format")
	}

	h := hmac.New(sha256.New, []byte(integrityKey))
	h.Write(body)
	hmacHex := hex.EncodeToString(h.Sum(nil))
	hmacB64 := base64.StdEncoding.EncodeToString(h.Sum(nil))

	shaBodyPlus := sha256.Sum256(append(append([]byte{}, body...), []byte(integrityKey)...))
	shaKeyPlus := sha256.Sum256(append([]byte(integrityKey), body...))
	bodyHex := hex.EncodeToString(shaBodyPlus[:])
	keyHex := hex.EncodeToString(shaKeyPlus[:])

	for _, candidate := range candidates {
		if signatureMatch(candidate, hmacHex) || signatureMatch(candidate, hmacB64) || signatureMatch(candidate, bodyHex) || signatureMatch(candidate, keyHex) {
			return nil
		}
	}

	return errors.New("invalid wompi signature")
}

func normalizeWompiMode(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "sandbox", "test", "testing", "sambox", "pruebas":
		return "sandbox"
	case "production", "prod", "live", "real", "reales":
		return "production"
	default:
		return ""
	}
}

func wompiModeFromKeys(publicKey, privateKey string) string {
	if strings.HasPrefix(privateKey, "prv_test_") || strings.HasPrefix(publicKey, "pub_test_") {
		return "sandbox"
	}
	if strings.TrimSpace(publicKey) != "" || strings.TrimSpace(privateKey) != "" {
		return "production"
	}
	return ""
}

func resolveWompiMode(dbSuper *sql.DB, publicKey, privateKey string) (string, string) {
	if configuredMode, _, err := dbpkg.GetConfigValue(dbSuper, "wompi.mode"); err == nil {
		if normalized := normalizeWompiMode(configuredMode); normalized != "" {
			return normalized, "manual"
		}
	}
	if inferred := wompiModeFromKeys(publicKey, privateKey); inferred != "" {
		return inferred, "keys"
	}
	return "sandbox", "default"
}

func wompiBaseURLFromMode(mode string) string {
	if normalizeWompiMode(mode) == "sandbox" {
		return "https://sandbox.wompi.co/v1"
	}
	return "https://production.wompi.co/v1"
}

func fetchWompiAcceptanceInfo(baseURL, publicKey string) (string, string, string, string, error) {
	if strings.TrimSpace(publicKey) == "" {
		return "", "", "", "", fmt.Errorf("wompi.public_key no configurada")
	}
	merchantURL := strings.TrimRight(baseURL, "/") + "/merchants/" + url.PathEscape(publicKey)
	req, err := http.NewRequest("GET", merchantURL, nil)
	if err != nil {
		return "", "", "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+publicKey)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", "", "", "", fmt.Errorf("wompi merchants error %s: %s", resp.Status, string(body))
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(body, &obj); err != nil {
		return "", "", "", "", err
	}
	data, _ := obj["data"].(map[string]interface{})
	presignedAcceptance, _ := data["presigned_acceptance"].(map[string]interface{})
	presignedPersonal, _ := data["presigned_personal_data_auth"].(map[string]interface{})
	acceptanceToken := strings.TrimSpace(fmt.Sprint(presignedAcceptance["acceptance_token"]))
	personalToken := strings.TrimSpace(fmt.Sprint(presignedPersonal["acceptance_token"]))
	acceptancePermalink := strings.TrimSpace(fmt.Sprint(presignedAcceptance["permalink"]))
	personalPermalink := strings.TrimSpace(fmt.Sprint(presignedPersonal["permalink"]))
	if acceptanceToken == "" || acceptanceToken == "<nil>" {
		acceptanceToken = ""
	}
	if personalToken == "" || personalToken == "<nil>" {
		personalToken = ""
	}
	if acceptancePermalink == "<nil>" {
		acceptancePermalink = ""
	}
	if personalPermalink == "<nil>" {
		personalPermalink = ""
	}
	return acceptanceToken, personalToken, acceptancePermalink, personalPermalink, nil
}

func normalizeEpaycoMode(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "sandbox", "sambox", "test", "testing", "pruebas":
		return "sandbox"
	case "production", "prod", "live", "real":
		return "production"
	default:
		return ""
	}
}

func epaycoModeFromKeys(custID, key string) string {
	combined := strings.ToLower(strings.TrimSpace(custID) + " " + strings.TrimSpace(key))
	if strings.Contains(combined, "test") || strings.Contains(combined, "sandbox") || strings.HasPrefix(strings.ToLower(strings.TrimSpace(custID)), "pub_test_") {
		return "sandbox"
	}
	if strings.TrimSpace(custID) != "" || strings.TrimSpace(key) != "" {
		return "production"
	}
	return ""
}

func resolveEpaycoMode(dbSuper *sql.DB, custID, key string) (string, string) {
	if configuredMode, _, err := dbpkg.GetConfigValue(dbSuper, "epayco.mode"); err == nil {
		if normalized := normalizeEpaycoMode(configuredMode); normalized != "" {
			return normalized, "manual"
		}
	}
	if inferred := epaycoModeFromKeys(custID, key); inferred != "" {
		return inferred, "keys"
	}
	return "sandbox", "default"
}

func parseBoolConfigValue(raw string) bool {
	v := strings.ToLower(strings.TrimSpace(raw))
	return v == "1" || v == "true" || v == "si" || v == "yes" || v == "on" || v == "activo"
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func getConfigEntryTrimmed(dbSuper *sql.DB, key string) (string, error) {
	value, _, _, _, err := dbpkg.GetConfigEntry(dbSuper, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(value), nil
}

func resolveEpaycoCredentials(dbSuper *sql.DB) (publicKey, customerID, privateKey string, err error) {
	publicKey, err = getConfigEntryTrimmed(dbSuper, "epayco.public_key")
	if err != nil {
		return "", "", "", err
	}
	customerID, err = getConfigEntryTrimmed(dbSuper, "epayco.customer_id")
	if err != nil {
		return "", "", "", err
	}
	privateKey, err = getDecryptedConfigValue(dbSuper, "epayco.private_key")
	if err != nil {
		return "", "", "", err
	}
	privateKey = strings.TrimSpace(privateKey)

	legacyCustID, err := getConfigEntryTrimmed(dbSuper, "epayco.cust_id")
	if err != nil {
		return "", "", "", err
	}
	legacyKey, err := getConfigEntryTrimmed(dbSuper, "epayco.key")
	if err != nil {
		return "", "", "", err
	}

	if publicKey == "" && looksLikeEpaycoPublicKey(legacyCustID) {
		publicKey = legacyCustID
	}
	if publicKey == "" && looksLikeEpaycoPublicKey(legacyKey) {
		publicKey = legacyKey
	}
	if customerID == "" && legacyCustID != "" && !looksLikeEpaycoPublicKey(legacyCustID) {
		customerID = legacyCustID
	}
	if privateKey == "" && legacyKey != "" && !looksLikeEpaycoPublicKey(legacyKey) {
		privateKey = legacyKey
	}

	return publicKey, customerID, privateKey, nil
}

func resolveEnabledConfigValue(dbSuper *sql.DB, key string, defaultValue bool) (bool, error) {
	raw, _, err := dbpkg.GetConfigValue(dbSuper, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return defaultValue, nil
		}
		return false, err
	}
	if strings.TrimSpace(raw) == "" {
		return defaultValue, nil
	}
	return parseBoolConfigValue(raw), nil
}

type licenciaPaymentMethodStatus struct {
	ID          string `json:"id"`
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion"`
	Enabled     bool   `json:"enabled"`
	Configured  bool   `json:"configured"`
	Available   bool   `json:"available"`
	SortOrder   int    `json:"sort_order"`
}

func loadLicenciaPaymentMethodStatuses(dbSuper *sql.DB) ([]licenciaPaymentMethodStatus, error) {
	epaycoPublicKey, _, epaycoPrivateKey, err := resolveEpaycoCredentials(dbSuper)
	if err != nil {
		return nil, err
	}
	epaycoConfigured := strings.TrimSpace(epaycoPublicKey) != "" && strings.TrimSpace(epaycoPrivateKey) != ""
	epaycoEnabled, err := resolveEnabledConfigValue(dbSuper, "epayco.enabled", false)
	if err != nil {
		return nil, err
	}

	wompiPublicKey, err := getConfigEntryTrimmed(dbSuper, "wompi.public_key")
	if err != nil {
		return nil, err
	}
	wompiPrivateKey, err := getConfigEntryTrimmed(dbSuper, "wompi.private_key")
	if err != nil {
		return nil, err
	}
	wompiIntegrityKey, err := getConfigEntryTrimmed(dbSuper, "wompi.integrity_key")
	if err != nil {
		return nil, err
	}
	wompiConfigured := wompiPublicKey != "" && wompiPrivateKey != "" && wompiIntegrityKey != ""
	wompiEnabled, err := resolveEnabledConfigValue(dbSuper, "wompi.enabled", wompiConfigured)
	if err != nil {
		return nil, err
	}

	return []licenciaPaymentMethodStatus{
		{
			ID:          "epayco",
			Nombre:      "Epayco",
			Descripcion: "Tarjeta, PSE y otros",
			Enabled:     epaycoEnabled,
			Configured:  epaycoConfigured,
			Available:   epaycoEnabled && epaycoConfigured,
			SortOrder:   1,
		},
		{
			ID:          "wompi",
			Nombre:      "Wompi",
			Descripcion: "Nequi / Tarjeta",
			Enabled:     wompiEnabled,
			Configured:  wompiConfigured,
			Available:   wompiEnabled && wompiConfigured,
			SortOrder:   2,
		},
	}, nil
}

func getLicenciaPaymentMethodStatus(dbSuper *sql.DB, methodID string) (licenciaPaymentMethodStatus, error) {
	statuses, err := loadLicenciaPaymentMethodStatuses(dbSuper)
	if err != nil {
		return licenciaPaymentMethodStatus{}, err
	}
	for _, status := range statuses {
		if status.ID == methodID {
			return status, nil
		}
	}
	return licenciaPaymentMethodStatus{}, fmt.Errorf("payment method not found: %s", methodID)
}

func PublicLicenciasPaymentMethodsHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		statuses, err := loadLicenciaPaymentMethodStatuses(dbSuper)
		if err != nil {
			http.Error(w, "failed to load payment methods: "+err.Error(), http.StatusInternalServerError)
			return
		}

		defaultMethod := ""
		for _, status := range statuses {
			if status.Available {
				defaultMethod = status.ID
				break
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"providers":      statuses,
			"default_method": defaultMethod,
		})
	}
}

func resolveRequestScheme(r *http.Request) string {
	if xfp := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); xfp != "" {
		parts := strings.Split(xfp, ",")
		if len(parts) > 0 {
			proto := strings.ToLower(strings.TrimSpace(parts[0]))
			if proto == "http" || proto == "https" {
				return proto
			}
		}
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func resolveRequestHost(r *http.Request) string {
	if xfh := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); xfh != "" {
		parts := strings.Split(xfh, ",")
		if len(parts) > 0 && strings.TrimSpace(parts[0]) != "" {
			return strings.TrimSpace(parts[0])
		}
	}
	return r.Host
}

const canonicalPaymentPublicBaseURL = "https://powerfulcontrolsystem.com"

var (
	epaycoApifyBaseURL           = "https://apify.epayco.co"
	epaycoSmartCheckoutScriptURL = "https://checkout.epayco.co/checkout-v2.js"
)

func splitHostPortLoose(rawHost string) string {
	trimmed := strings.TrimSpace(rawHost)
	if trimmed == "" {
		return ""
	}
	hostOnly, _, err := net.SplitHostPort(trimmed)
	if err == nil {
		return strings.TrimSpace(hostOnly)
	}
	if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
		return strings.Trim(strings.TrimSpace(trimmed), "[]")
	}
	return trimmed
}

func isLoopbackOrLocalHost(rawHost string) bool {
	host := strings.ToLower(splitHostPortLoose(rawHost))
	if host == "" {
		return false
	}
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func normalizeConfiguredBaseURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || strings.TrimSpace(parsed.Host) == "" {
		return ""
	}
	if isLoopbackOrLocalHost(parsed.Host) {
		return ""
	}
	hostOnly := strings.ToLower(splitHostPortLoose(parsed.Host))
	if hostOnly == "www.powerfulcontrolsystem.com" {
		parsed.Host = "powerfulcontrolsystem.com"
	}
	parsed.Scheme = "https"
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return strings.TrimRight(parsed.String(), "/")
}

func resolvePaymentBaseURL(r *http.Request, dbSuper *sql.DB) (string, error) {
	if configured, err := getDecryptedConfigValue(dbSuper, "gmail.confirm_base_url"); err == nil {
		if normalized := normalizeConfiguredBaseURL(configured); normalized != "" {
			return normalized, nil
		}
	}

	if r != nil {
		for _, headerName := range []string{"Origin", "Referer"} {
			if normalized := normalizeConfiguredBaseURL(strings.TrimSpace(r.Header.Get(headerName))); normalized != "" {
				return normalized, nil
			}
		}

		host := strings.TrimSpace(resolveRequestHost(r))
		hostOnly := strings.ToLower(splitHostPortLoose(host))
		if hostOnly != "" {
			if hostOnly == "www.powerfulcontrolsystem.com" {
				host = "powerfulcontrolsystem.com"
			}
			scheme := resolveRequestScheme(r)
			if scheme != "https" {
				scheme = "https"
			}
			if normalized := normalizeConfiguredBaseURL(scheme + "://" + host); normalized != "" {
				return normalized, nil
			}
		}
	}

	return canonicalPaymentPublicBaseURL, nil
}

func looksLikeEpaycoPublicKey(raw string) bool {
	v := strings.ToLower(strings.TrimSpace(raw))
	return strings.HasPrefix(v, "pub_")
}

func maskConfigValue(raw string, visiblePrefix, visibleSuffix int) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	if visiblePrefix < 0 {
		visiblePrefix = 0
	}
	if visibleSuffix < 0 {
		visibleSuffix = 0
	}
	if len(trimmed) <= visiblePrefix+visibleSuffix {
		if len(trimmed) <= 6 {
			return trimmed
		}
		return trimmed[:2] + "..." + trimmed[len(trimmed)-2:]
	}
	return trimmed[:visiblePrefix] + "..." + trimmed[len(trimmed)-visibleSuffix:]
}

func buildLicenciaReturnURL(baseURL, provider, status, reference string, licenciaID, empresaID int64) string {
	trimmedBaseURL := strings.TrimSpace(baseURL)
	query := url.Values{}
	if strings.TrimSpace(provider) != "" {
		query.Set("provider", strings.ToLower(strings.TrimSpace(provider)))
	}
	if strings.TrimSpace(status) != "" {
		query.Set("status", strings.ToLower(strings.TrimSpace(status)))
	}
	if strings.TrimSpace(reference) != "" {
		query.Set("reference", strings.TrimSpace(reference))
	}
	if licenciaID > 0 {
		query.Set("licencia_id", strconv.FormatInt(licenciaID, 10))
	}
	if empresaID > 0 {
		query.Set("empresa_id", strconv.FormatInt(empresaID, 10))
	}

	parsed, err := url.Parse(trimmedBaseURL)
	if err != nil || strings.TrimSpace(parsed.Host) == "" {
		return strings.TrimRight(trimmedBaseURL, "/") + "/pagar_licencia.html?" + query.Encode()
	}
	parsed.Path = "/pagar_licencia.html"
	parsed.RawPath = ""
	parsed.RawQuery = query.Encode()
	parsed.Fragment = ""
	return parsed.String()
}

func buildEpaycoResponseURL(baseURL, status, reference string, licenciaID, empresaID int64) string {
	trimmedBaseURL := strings.TrimSpace(baseURL)
	query := url.Values{}
	query.Set("provider", "epayco")
	if strings.TrimSpace(status) != "" {
		query.Set("status", strings.ToLower(strings.TrimSpace(status)))
	}
	if strings.TrimSpace(reference) != "" {
		query.Set("reference", strings.TrimSpace(reference))
	}
	if licenciaID > 0 {
		value := strconv.FormatInt(licenciaID, 10)
		query.Set("licencia_id", value)
		query.Set("extra1", value)
	}
	if empresaID > 0 {
		value := strconv.FormatInt(empresaID, 10)
		query.Set("empresa_id", value)
		query.Set("extra2", value)
	}

	parsed, err := url.Parse(trimmedBaseURL)
	if err != nil || strings.TrimSpace(parsed.Host) == "" {
		return strings.TrimRight(trimmedBaseURL, "/") + "/epayco/respuesta.html?" + query.Encode()
	}
	parsed.Path = "/epayco/respuesta.html"
	parsed.RawPath = ""
	parsed.RawQuery = query.Encode()
	parsed.Fragment = ""
	return parsed.String()
}

func buildEpaycoCheckoutURL(baseURL, publicKey, customerID, privateKey, reference, licenciaNombre string, licenciaID, empresaID int64, amount float64, customerEmail, mode string) string {
	title := strings.TrimSpace(licenciaNombre)
	if title == "" {
		title = "Licencia"
	}
	trimmedBaseURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	responseURL := buildEpaycoResponseURL(baseURL, "pending", reference, licenciaID, empresaID)
	confirmationURL := trimmedBaseURL + "/epayco/webhook"

	v := url.Values{}
	v.Set("public_key", strings.TrimSpace(publicKey))
	if strings.TrimSpace(customerID) != "" {
		v.Set("p_cust_id_cliente", strings.TrimSpace(customerID))
	}
	if strings.TrimSpace(privateKey) != "" {
		v.Set("p_key", strings.TrimSpace(privateKey))
	}
	v.Set("name", "Licencia "+title)
	v.Set("description", "Pago de licencia "+title)
	v.Set("invoice", reference)
	v.Set("currency", "cop")
	v.Set("amount", strconv.FormatFloat(amount, 'f', 2, 64))
	v.Set("tax", "0")
	v.Set("tax_base", "0")
	v.Set("country", "co")
	v.Set("lang", "es")
	v.Set("external", "false")
	v.Set("response", responseURL)
	v.Set("confirmation", confirmationURL)
	v.Set("test", strconv.FormatBool(normalizeEpaycoMode(mode) == "sandbox"))
	if strings.TrimSpace(customerEmail) != "" {
		v.Set("email_billing", strings.TrimSpace(customerEmail))
	}

	return "https://checkout.epayco.co/checkout.php?" + v.Encode()
}

func fetchEpaycoApifyToken(publicKey, privateKey string) (string, string, error) {
	loginURL := strings.TrimRight(strings.TrimSpace(epaycoApifyBaseURL), "/") + "/login"
	req, err := http.NewRequest(http.MethodPost, loginURL, nil)
	if err != nil {
		return "", "", err
	}
	authToken := base64.StdEncoding.EncodeToString([]byte(strings.TrimSpace(publicKey) + ":" + strings.TrimSpace(privateKey)))
	req.Header.Set("Authorization", "Basic "+authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	rawBody := string(body)
	if resp.StatusCode >= http.StatusBadRequest {
		return "", rawBody, fmt.Errorf("epayco login error: %s", rawBody)
	}

	payload := map[string]interface{}{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &payload); err != nil {
			return "", rawBody, fmt.Errorf("invalid epayco login response: %w", err)
		}
	}
	token := strings.TrimSpace(pickEpaycoField(payload, "token", "access_token"))
	if token == "" {
		return "", rawBody, errors.New("epayco login did not return token")
	}
	return token, rawBody, nil
}

func createEpaycoSmartCheckoutSession(apifyToken string, sessionPayload map[string]interface{}) (string, string, error) {
	endpoint := strings.TrimRight(strings.TrimSpace(epaycoApifyBaseURL), "/") + "/payment/session/create"
	body, err := json.Marshal(sessionPayload)
	if err != nil {
		return "", "", err
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apifyToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	rawBody := string(respBody)
	if resp.StatusCode >= http.StatusBadRequest {
		return "", rawBody, fmt.Errorf("epayco session create error: %s", rawBody)
	}

	payload := map[string]interface{}{}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &payload); err != nil {
			return "", rawBody, fmt.Errorf("invalid epayco session response: %w", err)
		}
	}
	sessionID := strings.TrimSpace(pickEpaycoField(payload, "sessionId", "session_id"))
	if sessionID == "" {
		return "", rawBody, errors.New("epayco session create did not return sessionId")
	}
	return sessionID, rawBody, nil
}

func pickEpaycoField(payload map[string]interface{}, keys ...string) string {
	if payload == nil {
		return ""
	}
	sources := make([]map[string]interface{}, 0, 2)
	if data, ok := payload["data"].(map[string]interface{}); ok && len(data) > 0 {
		sources = append(sources, data)
	}
	sources = append(sources, payload)

	for _, source := range sources {
		for _, key := range keys {
			value, ok := source[key]
			if !ok {
				continue
			}
			s := strings.TrimSpace(fmt.Sprint(value))
			if s != "" && s != "<nil>" {
				return s
			}
		}
	}

	return ""
}

func paymentContextFromInternalReference(values ...string) (int64, int64, bool) {
	for _, raw := range values {
		candidate := strings.ToUpper(strings.TrimSpace(raw))
		if candidate == "" {
			continue
		}
		parts := strings.Split(candidate, "-")
		for idx := 0; idx < len(parts)-3; idx++ {
			if parts[idx] != "LIC" || parts[idx+2] != "EMP" {
				continue
			}
			licenciaID, licErr := strconv.ParseInt(strings.TrimSpace(parts[idx+1]), 10, 64)
			empresaID, empErr := strconv.ParseInt(strings.TrimSpace(parts[idx+3]), 10, 64)
			if licErr == nil && empErr == nil && licenciaID > 0 && empresaID > 0 {
				return licenciaID, empresaID, true
			}
		}
	}
	return 0, 0, false
}

func expectedPaymentContextFromRequest(r *http.Request) (int64, int64, bool) {
	if r == nil || r.URL == nil {
		return 0, 0, false
	}
	licenciaID, licErr := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("licencia_id")), 10, 64)
	empresaID, empErr := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("empresa_id")), 10, 64)
	if licErr != nil || empErr != nil || licenciaID <= 0 || empresaID <= 0 {
		return 0, 0, false
	}
	return licenciaID, empresaID, true
}

func paymentContextMatchesExpected(licenciaID, empresaID, expectedLicenciaID, expectedEmpresaID int64) bool {
	if expectedLicenciaID <= 0 || expectedEmpresaID <= 0 {
		return true
	}
	return licenciaID == expectedLicenciaID && empresaID == expectedEmpresaID
}

func parseEpaycoPaymentStatus(payload map[string]interface{}) string {
	cod := strings.ToUpper(strings.TrimSpace(pickEpaycoField(payload, "x_cod_response", "cod_response", "status_code")))
	switch cod {
	case "1", "APPROVED", "ACCEPTED", "ACEPTADA", "APROBADA":
		return "APPROVED"
	case "2", "DECLINED", "REJECTED", "RECHAZADA":
		return "DECLINED"
	case "3", "PENDING":
		return "PENDING"
	case "4", "ERROR", "FAILED":
		return "ERROR"
	}

	raw := strings.ToLower(strings.TrimSpace(pickEpaycoField(payload, "x_response", "x_transaction_state", "x_respuesta", "status", "state")))
	switch {
	case strings.Contains(raw, "acept"), strings.Contains(raw, "aprobad"), strings.Contains(raw, "approved"), strings.Contains(raw, "accredited"):
		return "APPROVED"
	case strings.Contains(raw, "declin"), strings.Contains(raw, "rechaz"), strings.Contains(raw, "cancel"), strings.Contains(raw, "anulad"):
		return "DECLINED"
	case strings.Contains(raw, "pend"):
		return "PENDING"
	case strings.Contains(raw, "error"), strings.Contains(raw, "fall"), strings.Contains(raw, "failed"):
		return "ERROR"
	default:
		return ""
	}
}

func shouldPreservePendingEpaycoStatus(storedStatus string, payload map[string]interface{}) bool {
	if strings.ToUpper(strings.TrimSpace(storedStatus)) != "PENDING" || len(payload) == 0 {
		return false
	}

	combined := strings.ToLower(strings.TrimSpace(strings.Join([]string{
		pickEpaycoField(payload, "status", "state", "x_response", "x_transaction_state"),
		pickEpaycoField(payload, "description", "message", "descripcion", "mensaje"),
	}, " ")))
	if combined == "" || !strings.Contains(combined, "error") {
		return false
	}

	return strings.Contains(combined, "dato") || strings.Contains(combined, "conex") || strings.Contains(combined, "verifique")
}

func extractEpaycoPaymentInfo(payload map[string]interface{}) (string, string, string) {
	transactionID := strings.TrimSpace(pickEpaycoField(payload, "transaction_id", "x_transaction_id", "id", "tx_id"))
	reference := strings.TrimSpace(pickEpaycoField(payload, "reference", "x_ref_payco", "invoice", "ref_payco"))
	status := strings.ToUpper(strings.TrimSpace(parseEpaycoPaymentStatus(payload)))
	if status == "" {
		status = "PENDING"
	}
	return transactionID, reference, status
}

// WompiConfigHandler gestiona credenciales de Wompi para pagos alternativos con Nequi.
func WompiConfigHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			pub, _, _, pubUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "wompi.public_key")
			prv, _, _, prvUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "wompi.private_key")
			integrity, _, _, intUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "wompi.integrity_key")
			modeRaw, _, _, modeUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "wompi.mode")

			pubSet := pub != ""
			prvSet := prv != ""
			intSet := integrity != ""

			pubMasked := ""
			if pubSet {
				if len(pub) > 16 {
					pubMasked = pub[:8] + "..." + pub[len(pub)-6:]
				} else {
					pubMasked = pub
				}
			}

			prvMasked := ""
			if prvSet {
				prvMasked = "********"
			}

			integrityMasked := ""
			if intSet {
				integrityMasked = "********"
			}

			configuredMode := normalizeWompiMode(modeRaw)
			mode := configuredMode
			modeSource := "manual"
			if mode == "" {
				mode = wompiModeFromKeys(pub, prv)
				if mode != "" {
					modeSource = "keys"
				} else {
					mode = "sandbox"
					modeSource = "default"
				}
			}

			enabled, err := resolveEnabledConfigValue(dbSuper, "wompi.enabled", pubSet && prvSet && intSet)
			if err != nil {
				http.Error(w, "failed to read wompi.enabled: "+err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"public_key_set":        pubSet,
				"public_key_masked":     pubMasked,
				"public_key_updated":    pubUpdated,
				"private_key_set":       prvSet,
				"private_key_masked":    prvMasked,
				"private_key_updated":   prvUpdated,
				"integrity_key_set":     intSet,
				"integrity_key_masked":  integrityMasked,
				"integrity_key_updated": intUpdated,
				"encryption_available":  utils.EncryptionAvailable(),
				"enabled":               enabled,
				"mode":                  mode,
				"mode_set":              configuredMode != "",
				"mode_source":           modeSource,
				"mode_updated":          modeUpdated,
			})
			return

		case http.MethodPost, http.MethodPut:
			var payload struct {
				PublicKey    string `json:"public_key"`
				PrivateKey   string `json:"private_key"`
				IntegrityKey string `json:"integrity_key"`
				Enabled      *bool  `json:"enabled"`
				Mode         string `json:"mode"`
				Encrypt      bool   `json:"encrypt"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
				return
			}
			modeInput := strings.TrimSpace(payload.Mode)
			normalizedMode := normalizeWompiMode(modeInput)
			if modeInput != "" && normalizedMode == "" {
				http.Error(w, "mode inválido: usa sandbox o real", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.PublicKey) == "" && strings.TrimSpace(payload.PrivateKey) == "" && strings.TrimSpace(payload.IntegrityKey) == "" && normalizedMode == "" && payload.Enabled == nil {
				http.Error(w, "at least one value is required (enabled, mode o llaves)", http.StatusBadRequest)
				return
			}

			if payload.PublicKey != "" && !strings.HasPrefix(payload.PublicKey, "pub_") {
				http.Error(w, "public_key inválida: debe iniciar con pub_", http.StatusBadRequest)
				return
			}
			if payload.PrivateKey != "" && !strings.HasPrefix(payload.PrivateKey, "prv_") {
				http.Error(w, "private_key inválida: debe iniciar con prv_", http.StatusBadRequest)
				return
			}
			if payload.IntegrityKey != "" && !strings.Contains(payload.IntegrityKey, "integrity") {
				http.Error(w, "integrity_key inválida: prefijo esperado *_integrity_*", http.StatusBadRequest)
				return
			}

			// Requerir cifrado obligatorio para llaves sensibles.
			if payload.PublicKey != "" {
				if err := dbpkg.SetConfigValue(dbSuper, "wompi.public_key", payload.PublicKey, false); err != nil {
					http.Error(w, "failed to save wompi.public_key: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			saveSensitive := func(key, value string) error {
				if value == "" {
					return nil
				}
				if !utils.EncryptionAvailable() {
					return fmt.Errorf("encryption required: CONFIG_ENC_KEY not set")
				}
				encVal, err := utils.EncryptString(value)
				if err != nil {
					return err
				}
				return dbpkg.SetConfigValue(dbSuper, key, encVal, true)
			}

			if err := saveSensitive("wompi.private_key", payload.PrivateKey); err != nil {
				http.Error(w, "failed to save wompi.private_key: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if err := saveSensitive("wompi.integrity_key", payload.IntegrityKey); err != nil {
				http.Error(w, "failed to save wompi.integrity_key: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if normalizedMode != "" {
				if err := dbpkg.SetConfigValue(dbSuper, "wompi.mode", normalizedMode, false); err != nil {
					http.Error(w, "failed to save wompi.mode: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if payload.Enabled != nil {
				v := "0"
				if *payload.Enabled {
					v = "1"
				}
				if err := dbpkg.SetConfigValue(dbSuper, "wompi.enabled", v, false); err != nil {
					http.Error(w, "failed to save wompi.enabled: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"saved": true, "mode": normalizedMode, "enabled": payload.Enabled})
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

// EpaycoConfigHandler gestiona credenciales de Epayco (public/private key y customer ID opcional) y flag de activación.
func EpaycoConfigHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			publicKeyRaw, _, _, publicKeyUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "epayco.public_key")
			customerIDRaw, _, _, customerIDUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "epayco.customer_id")
			privateKeyRaw, privateKeyEnc, _, privateKeyUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "epayco.private_key")
			legacyCustRaw, _, _, legacyCustUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "epayco.cust_id")
			legacyKeyRaw, legacyKeyEnc, _, legacyKeyUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "epayco.key")
			enabledVal, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, "epayco.enabled")
			modeRaw, _, _, modeUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "epayco.mode")

			publicKey := strings.TrimSpace(publicKeyRaw)
			publicKeyUpdatedAt := publicKeyUpdated
			if publicKey == "" && looksLikeEpaycoPublicKey(legacyCustRaw) {
				publicKey = strings.TrimSpace(legacyCustRaw)
				publicKeyUpdatedAt = legacyCustUpdated
			}
			if publicKey == "" && looksLikeEpaycoPublicKey(legacyKeyRaw) {
				publicKey = strings.TrimSpace(legacyKeyRaw)
				publicKeyUpdatedAt = legacyKeyUpdated
			}

			customerID := strings.TrimSpace(customerIDRaw)
			customerIDUpdatedAt := customerIDUpdated
			if customerID == "" && strings.TrimSpace(legacyCustRaw) != "" && !looksLikeEpaycoPublicKey(legacyCustRaw) {
				customerID = strings.TrimSpace(legacyCustRaw)
				customerIDUpdatedAt = legacyCustUpdated
			}

			privateKey := strings.TrimSpace(privateKeyRaw)
			privateKeyUpdatedAt := privateKeyUpdated
			privateKeyEncrypted := privateKeyEnc
			if privateKey == "" && strings.TrimSpace(legacyKeyRaw) != "" && !looksLikeEpaycoPublicKey(legacyKeyRaw) {
				privateKey = strings.TrimSpace(legacyKeyRaw)
				privateKeyUpdatedAt = legacyKeyUpdated
				privateKeyEncrypted = legacyKeyEnc
			}

			publicKeySet := publicKey != ""
			customerIDSet := customerID != ""
			privateKeySet := privateKey != ""

			enabled := parseBoolConfigValue(enabledVal)

			configuredMode := normalizeEpaycoMode(modeRaw)
			mode := configuredMode
			modeSource := "manual"
			if mode == "" {
				mode = epaycoModeFromKeys(publicKey, privateKey)
				if mode != "" {
					modeSource = "keys"
				} else {
					mode = "sandbox"
					modeSource = "default"
				}
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"public_key_set":      publicKeySet,
				"public_key_masked":   maskConfigValue(publicKey, 4, 4),
				"public_key_updated":  publicKeyUpdatedAt,
				"customer_id_set":     customerIDSet,
				"customer_id_masked":  maskConfigValue(customerID, 2, 3),
				"customer_id_updated": customerIDUpdatedAt,
				"private_key_set":     privateKeySet,
				"private_key_masked": func() string {
					if privateKeySet {
						return "********"
					}
					return ""
				}(),
				"private_key_encrypted": privateKeyEncrypted,
				"private_key_updated":   privateKeyUpdatedAt,
				"cust_id_set":           customerIDSet || publicKeySet,
				"cust_id_masked": func() string {
					if customerIDSet {
						return maskConfigValue(customerID, 2, 3)
					}
					return maskConfigValue(publicKey, 4, 4)
				}(),
				"cust_id_updated": func() string {
					if customerIDSet {
						return customerIDUpdatedAt
					}
					return publicKeyUpdatedAt
				}(),
				"key_set": privateKeySet,
				"key_masked": func() string {
					if privateKeySet {
						return "********"
					}
					return ""
				}(),
				"key_encrypted":        privateKeyEncrypted,
				"key_updated":          privateKeyUpdatedAt,
				"encryption_available": utils.EncryptionAvailable(),
				"enabled":              enabled,
				"mode":                 mode,
				"mode_set":             configuredMode != "",
				"mode_source":          modeSource,
				"mode_updated":         modeUpdated,
			})
			return

		case http.MethodPost, http.MethodPut:
			var payload struct {
				PublicKey  string `json:"public_key"`
				CustomerID string `json:"customer_id"`
				PrivateKey string `json:"private_key"`
				CustID     string `json:"cust_id"`
				Key        string `json:"key"`
				Enabled    *bool  `json:"enabled"`
				Mode       string `json:"mode"`
				Encrypt    bool   `json:"encrypt"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
				return
			}

			normalizedMode := normalizeEpaycoMode(payload.Mode)
			if strings.TrimSpace(payload.Mode) != "" && normalizedMode == "" {
				http.Error(w, "mode invalido: usa sandbox o production", http.StatusBadRequest)
				return
			}

			publicKey := strings.TrimSpace(payload.PublicKey)
			legacyCustID := strings.TrimSpace(payload.CustID)
			if publicKey == "" && looksLikeEpaycoPublicKey(legacyCustID) {
				publicKey = legacyCustID
			}

			customerID := strings.TrimSpace(payload.CustomerID)
			if customerID == "" && legacyCustID != "" && !looksLikeEpaycoPublicKey(legacyCustID) {
				customerID = legacyCustID
			}

			privateKey := strings.TrimSpace(payload.PrivateKey)
			if privateKey == "" {
				privateKey = strings.TrimSpace(payload.Key)
			}

			if publicKey == "" && customerID == "" && privateKey == "" && payload.Enabled == nil && normalizedMode == "" {
				http.Error(w, "at least one of public_key, customer_id, private_key, enabled or mode is required", http.StatusBadRequest)
				return
			}

			if publicKey != "" {
				if strings.ContainsAny(publicKey, " \t\r\n") {
					http.Error(w, "public_key invalida: no puede contener espacios", http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, "epayco.public_key", publicKey, false); err != nil {
					http.Error(w, "failed to save epayco.public_key: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			if customerID != "" {
				if strings.ContainsAny(customerID, " \t\r\n") {
					http.Error(w, "customer_id invalido: no puede contener espacios", http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, "epayco.customer_id", customerID, false); err != nil {
					http.Error(w, "failed to save epayco.customer_id: "+err.Error(), http.StatusInternalServerError)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, "epayco.cust_id", customerID, false); err != nil {
					http.Error(w, "failed to save epayco.cust_id: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			if privateKey != "" {
				if !utils.EncryptionAvailable() {
					http.Error(w, "encryption required: CONFIG_ENC_KEY not set", http.StatusInternalServerError)
					return
				}
				encVal, err := utils.EncryptString(privateKey)
				if err != nil {
					http.Error(w, "failed to encrypt epayco.private_key: "+err.Error(), http.StatusInternalServerError)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, "epayco.private_key", encVal, true); err != nil {
					http.Error(w, "failed to save epayco.private_key: "+err.Error(), http.StatusInternalServerError)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, "epayco.key", encVal, true); err != nil {
					http.Error(w, "failed to save epayco.key: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			if payload.Enabled != nil {
				v := "0"
				if *payload.Enabled {
					v = "1"
				}
				if err := dbpkg.SetConfigValue(dbSuper, "epayco.enabled", v, false); err != nil {
					http.Error(w, "failed to save epayco.enabled: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			if normalizedMode != "" {
				if err := dbpkg.SetConfigValue(dbSuper, "epayco.mode", normalizedMode, false); err != nil {
					http.Error(w, "failed to save epayco.mode: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"saved": true, "mode": normalizedMode})
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

// WompiTermsHandler devuelve links de términos y autorizaciones para cumplimiento de aceptación.
func WompiTermsHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		status, err := getLicenciaPaymentMethodStatus(dbSuper, "wompi")
		if err != nil {
			http.Error(w, "failed to read wompi availability: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if !status.Enabled {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPreconditionFailed)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":    "Wompi no esta activo en configuracion avanzada",
				"provider": "wompi",
			})
			return
		}
		if !status.Configured {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPreconditionFailed)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":    "Wompi no esta configurado completamente",
				"provider": "wompi",
			})
			return
		}
		publicKey, err := getDecryptedConfigValue(dbSuper, "wompi.public_key")
		if err != nil {
			http.Error(w, "failed to read wompi.public_key: "+err.Error(), http.StatusInternalServerError)
			return
		}
		privateKey, _ := getDecryptedConfigValue(dbSuper, "wompi.private_key")
		if strings.TrimSpace(publicKey) == "" {
			http.Error(w, "wompi.public_key not configured", http.StatusInternalServerError)
			return
		}
		mode, modeSource := resolveWompiMode(dbSuper, publicKey, privateKey)
		baseURL := wompiBaseURLFromMode(mode)
		_, _, acceptancePermalink, personalPermalink, ferr := fetchWompiAcceptanceInfo(baseURL, publicKey)
		if ferr != nil {
			http.Error(w, "failed to fetch acceptance tokens: "+ferr.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"provider":                    "wompi",
			"payment_method":              "NEQUI",
			"mode":                        mode,
			"mode_source":                 modeSource,
			"api_base_url":                baseURL,
			"acceptance_permalink":        acceptancePermalink,
			"personal_data_permalink":     personalPermalink,
			"sandbox_phone_approved":      "3991111111",
			"sandbox_phone_declined":      "3992222222",
			"sandbox_phone_error_example": "3993333333",
		})
	}
}

// WompiCreateNequiTransactionHandler crea una transacción Wompi usando método de pago NEQUI.
func WompiCreateNequiTransactionHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			LicenciaID    int64  `json:"licencia_id"`
			EmpresaID     int64  `json:"empresa_id,omitempty"`
			PhoneNumber   string `json:"phone_number"`
			CustomerEmail string `json:"customer_email,omitempty"`
			AcceptTerms   bool   `json:"accept_terms"`
			DiscountCode  string `json:"discount_code,omitempty"`
			AsesorID      string `json:"asesor_id,omitempty"`
			VendedorID    string `json:"vendedor_id,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
			return
		}

		if payload.LicenciaID <= 0 {
			http.Error(w, "licencia_id inválido", http.StatusBadRequest)
			return
		}
		phone := strings.TrimSpace(payload.PhoneNumber)
		if ok, _ := regexp.MatchString(`^3\d{9}$`, phone); !ok {
			http.Error(w, "phone_number inválido: usa 10 dígitos colombianos (ej. 3991111111 en sandbox)", http.StatusBadRequest)
			return
		}
		status, err := getLicenciaPaymentMethodStatus(dbSuper, "wompi")
		if err != nil {
			http.Error(w, "failed to read wompi availability: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if !status.Enabled {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPreconditionFailed)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":    "Wompi no esta activo en configuracion avanzada",
				"provider": "wompi",
			})
			return
		}
		if !status.Configured {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPreconditionFailed)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":    "Wompi no esta configurado completamente",
				"provider": "wompi",
			})
			return
		}

		publicKey, err := getDecryptedConfigValue(dbSuper, "wompi.public_key")
		if err != nil {
			http.Error(w, "failed to read wompi.public_key: "+err.Error(), http.StatusInternalServerError)
			return
		}
		privateKey, err := getDecryptedConfigValue(dbSuper, "wompi.private_key")
		if err != nil {
			http.Error(w, "failed to read wompi.private_key: "+err.Error(), http.StatusInternalServerError)
			return
		}
		integrityKey, err := getDecryptedConfigValue(dbSuper, "wompi.integrity_key")
		if err != nil {
			http.Error(w, "failed to read wompi.integrity_key: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if strings.TrimSpace(publicKey) == "" || strings.TrimSpace(privateKey) == "" || strings.TrimSpace(integrityKey) == "" {
			http.Error(w, "Wompi no configurado: faltan llaves (public/private/integrity)", http.StatusInternalServerError)
			return
		}

		lic, err := dbpkg.GetLicenciaByID(dbSuper, payload.LicenciaID)
		if err != nil || lic == nil {
			http.Error(w, "licencia not found", http.StatusBadRequest)
			return
		}

		amountInCents := int64(math.Round(lic.Valor * 100))
		if amountInCents <= 0 {
			http.Error(w, "valor de licencia inválido para Wompi", http.StatusBadRequest)
			return
		}

		mode, _ := resolveWompiMode(dbSuper, publicKey, privateKey)
		baseURL := wompiBaseURLFromMode(mode)
		acceptanceToken, personalToken, acceptancePermalink, personalPermalink, ferr := fetchWompiAcceptanceInfo(baseURL, publicKey)
		if ferr != nil {
			http.Error(w, "failed to fetch Wompi acceptance data: "+ferr.Error(), http.StatusBadGateway)
			return
		}

		if !payload.AcceptTerms {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":                   "Debes aceptar términos y autorización de datos para continuar con Nequi",
				"acceptance_permalink":    acceptancePermalink,
				"personal_data_permalink": personalPermalink,
			})
			return
		}

		if acceptanceToken == "" || personalToken == "" {
			http.Error(w, "Wompi no devolvió tokens de aceptación válidos", http.StatusBadGateway)
			return
		}

		email := strings.TrimSpace(payload.CustomerEmail)
		if email == "" {
			email = strings.TrimSpace(r.Header.Get("X-Admin-Email"))
		}
		if email == "" {
			http.Error(w, "customer_email requerido para crear la transacción", http.StatusBadRequest)
			return
		}

		reference := fmt.Sprintf("WOMPI-LIC-%d-EMP-%d-%d", payload.LicenciaID, payload.EmpresaID, time.Now().UnixNano())
		signatureSource := fmt.Sprintf("%s%dCOP%s", reference, amountInCents, integrityKey)
		signatureHash := sha256.Sum256([]byte(signatureSource))
		signature := hex.EncodeToString(signatureHash[:])

		paymentBaseURL, err := resolvePaymentBaseURL(r, dbSuper)
		if err != nil {
			http.Error(w, "failed to resolve public base URL for Wompi: "+err.Error(), http.StatusPreconditionFailed)
			return
		}
		redirectURL := buildLicenciaReturnURL(paymentBaseURL, "wompi", "pending", reference, payload.LicenciaID, payload.EmpresaID)

		reqBody := map[string]interface{}{
			"acceptance_token":     acceptanceToken,
			"accept_personal_auth": personalToken,
			"amount_in_cents":      amountInCents,
			"currency":             "COP",
			"customer_email":       email,
			"reference":            reference,
			"signature":            signature,
			"redirect_url":         redirectURL,
			"payment_method": map[string]interface{}{
				"type":         "NEQUI",
				"phone_number": phone,
			},
		}

		bodyBytes, _ := json.Marshal(reqBody)
		apiURL := strings.TrimRight(baseURL, "/") + "/transactions"
		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(bodyBytes))
		if err != nil {
			http.Error(w, "failed to create request: "+err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+privateKey)

		client := &http.Client{Timeout: 20 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "request error: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		if resp.StatusCode >= 400 {
			log.Println("Wompi API error:", resp.Status, string(respBody))
			http.Error(w, "wompi API error: "+string(respBody), http.StatusBadGateway)
			return
		}

		var wompiResp map[string]interface{}
		if err := json.Unmarshal(respBody, &wompiResp); err != nil {
			http.Error(w, "invalid response from wompi: "+err.Error(), http.StatusInternalServerError)
			return
		}

		data, _ := wompiResp["data"].(map[string]interface{})
		transactionID := strings.TrimSpace(fmt.Sprint(data["id"]))
		transactionStatus := strings.ToUpper(strings.TrimSpace(fmt.Sprint(data["status"])))
		respReference := strings.TrimSpace(fmt.Sprint(data["reference"]))
		if transactionID == "" || transactionID == "<nil>" {
			http.Error(w, "wompi response sin transaction id", http.StatusBadGateway)
			return
		}
		if transactionStatus == "" || transactionStatus == "<nil>" {
			transactionStatus = "PENDING"
		}
		if respReference == "" || respReference == "<nil>" {
			respReference = reference
		}

		// compatibilidad: si se envió `vendedor_id` usarlo como `asesor_id` si este último está vacío
		if strings.TrimSpace(payload.AsesorID) == "" && strings.TrimSpace(payload.VendedorID) != "" {
			payload.AsesorID = payload.VendedorID
		}

		if _, err := dbpkg.CreateWompiPaymentRecord(dbSuper, payload.LicenciaID, payload.EmpresaID, transactionID, respReference, transactionStatus, string(respBody), payload.DiscountCode, payload.AsesorID); err != nil {
			log.Println("warning: failed to record Wompi transaction in DB:", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"provider":                "wompi",
			"payment_method":          "NEQUI",
			"mode":                    mode,
			"transaction_id":          transactionID,
			"reference":               respReference,
			"status":                  transactionStatus,
			"acceptance_permalink":    acceptancePermalink,
			"personal_data_permalink": personalPermalink,
			"data":                    data,
		})
	}
}

// recordVendedorComisiones busca el pago por transaction o reference y crea registros de comisión
func recordVendedorComisiones(db *sql.DB, transactionID, reference string, licenciaID, empresaID int64) {
	var payRec *dbpkg.WompiPaymentRecord
	var err error
	if strings.TrimSpace(transactionID) != "" {
		payRec, err = dbpkg.GetWompiPaymentByTransaction(db, transactionID)
		if err != nil {
			log.Println("warning: failed to get pagos_wompi by transaction:", err)
			return
		}
	}
	if payRec == nil && strings.TrimSpace(reference) != "" {
		payRec, err = dbpkg.GetWompiPaymentByReference(db, reference)
		if err != nil {
			log.Println("warning: failed to get pagos_wompi by reference:", err)
			return
		}
	}
	if payRec == nil {
		// No payment record found
		return
	}

	asesorID := ""
	if payRec.AsesorID.Valid {
		asesorID = payRec.AsesorID.String
	}
	if strings.TrimSpace(asesorID) == "" {
		// No vendor code provided
		return
	}

	plan, err := dbpkg.GetAsesorComercialPlanByAsesorID(db, asesorID)
	if err != nil {
		log.Println("warning: failed to load vendedor plan:", err)
		return
	}
	if plan == nil {
		// No plan configured for this vendor
		return
	}

	lic, err := dbpkg.GetLicenciaByID(db, licenciaID)
	if err != nil || lic == nil {
		log.Println("warning: licencia not found when recording comision for vendedor:", err)
		return
	}

	montoTotal := lic.Valor
	porcentaje := plan.ComisionVentaPct
	montoComision := montoTotal * porcentaje / 100.0

	pagoID := int64(0)
	if payRec.ID > 0 {
		pagoID = payRec.ID
	}
	referenciaStr := ""
	if payRec.Reference.Valid {
		referenciaStr = payRec.Reference.String
	}

	obs := fmt.Sprintf("Comisión por venta licencia %d (tx=%s ref=%s)", licenciaID, transactionID, referenciaStr)
	if _, err := dbpkg.CreateAsesorComisionRecord(db, asesorID, empresaID, licenciaID, pagoID, transactionID, montoTotal, porcentaje, montoComision, referenciaStr, obs, "", 0); err != nil {
		log.Println("warning: failed to create asesor_comisiones record:", err)
	}

	// Programar comisiones para meses adicionales si aplica
	meses := 0
	if plan.MesesRenovacion > 0 {
		meses = plan.MesesRenovacion
	}
	if meses > 1 {
		for i := 1; i < meses; i++ {
			scheduledDate := time.Now().AddDate(0, i, 0).Format("2006-01-02")
			obs2 := fmt.Sprintf("Comisión programada %d/%d para vendedor %s", i+1, meses, asesorID)
			if _, err := dbpkg.CreateAsesorComisionRecord(db, asesorID, empresaID, licenciaID, 0, "", montoTotal, porcentaje, montoComision, referenciaStr, obs2, scheduledDate, 0); err != nil {
				log.Println("warning: failed to create scheduled asesor_comisiones record:", err)
			}
		}
	}
}

func recordVendedorComisionesEpayco(db *sql.DB, transactionID, reference string, licenciaID, empresaID int64) {
	var payRec *dbpkg.EpaycoPaymentRecord
	var err error
	if strings.TrimSpace(transactionID) != "" {
		payRec, err = dbpkg.GetEpaycoPaymentByTransaction(db, transactionID)
		if err != nil {
			log.Println("warning: failed to get pagos_epayco by transaction:", err)
			return
		}
	}
	if payRec == nil && strings.TrimSpace(reference) != "" {
		payRec, err = dbpkg.GetEpaycoPaymentByReference(db, reference)
		if err != nil {
			log.Println("warning: failed to get pagos_epayco by reference:", err)
			return
		}
	}
	if payRec == nil {
		return
	}

	asesorID := ""
	if payRec.AsesorID.Valid {
		asesorID = payRec.AsesorID.String
	}
	if strings.TrimSpace(asesorID) == "" {
		return
	}

	plan, err := dbpkg.GetAsesorComercialPlanByAsesorID(db, asesorID)
	if err != nil {
		log.Println("warning: failed to load vendedor plan (epayco):", err)
		return
	}
	if plan == nil {
		return
	}

	lic, err := dbpkg.GetLicenciaByID(db, licenciaID)
	if err != nil || lic == nil {
		log.Println("warning: licencia not found when recording comision epayco:", err)
		return
	}

	montoTotal := lic.Valor
	porcentaje := plan.ComisionVentaPct
	montoComision := montoTotal * porcentaje / 100.0

	pagoID := int64(0)
	if payRec.ID > 0 {
		pagoID = payRec.ID
	}
	referenciaStr := ""
	if payRec.Reference.Valid {
		referenciaStr = payRec.Reference.String
	}

	obs := fmt.Sprintf("Comision por venta licencia %d (epayco tx=%s ref=%s)", licenciaID, transactionID, referenciaStr)
	if _, err := dbpkg.CreateAsesorComisionRecord(db, asesorID, empresaID, licenciaID, pagoID, transactionID, montoTotal, porcentaje, montoComision, referenciaStr, obs, "", 0); err != nil {
		log.Println("warning: failed to create asesor_comisiones record (epayco):", err)
	}

	meses := 0
	if plan.MesesRenovacion > 0 {
		meses = plan.MesesRenovacion
	}
	if meses > 1 {
		for i := 1; i < meses; i++ {
			scheduledDate := time.Now().AddDate(0, i, 0).Format("2006-01-02")
			obs2 := fmt.Sprintf("Comision programada %d/%d para vendedor %s", i+1, meses, asesorID)
			if _, err := dbpkg.CreateAsesorComisionRecord(db, asesorID, empresaID, licenciaID, 0, "", montoTotal, porcentaje, montoComision, referenciaStr, obs2, scheduledDate, 0); err != nil {
				log.Println("warning: failed to create scheduled asesor_comisiones record (epayco):", err)
			}
		}
	}
}

// WompiTransactionStatusHandler consulta estado de la transacción y activa licencia si quedó APPROVED.
func WompiTransactionStatusHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		transactionID := strings.TrimSpace(r.URL.Query().Get("id"))
		if transactionID == "" {
			transactionID = strings.TrimSpace(r.URL.Query().Get("transaction_id"))
		}
		reference := strings.TrimSpace(r.URL.Query().Get("reference"))
		if reference == "" {
			reference = strings.TrimSpace(r.URL.Query().Get("ref"))
		}
		expectedLicenciaID, expectedEmpresaID, hasExpectedContext := expectedPaymentContextFromRequest(r)
		if transactionID == "" && reference != "" {
			rec, lookupErr := dbpkg.GetWompiPaymentByReference(dbSuper, reference)
			if lookupErr != nil {
				http.Error(w, "failed to resolve wompi reference: "+lookupErr.Error(), http.StatusInternalServerError)
				return
			}
			if rec != nil && rec.TransactionID.Valid {
				transactionID = strings.TrimSpace(rec.TransactionID.String)
			}
		}
		if transactionID == "" {
			http.Error(w, "id o reference requerido", http.StatusBadRequest)
			return
		}

		publicKey, err := getDecryptedConfigValue(dbSuper, "wompi.public_key")
		if err != nil {
			http.Error(w, "failed to read wompi.public_key: "+err.Error(), http.StatusInternalServerError)
			return
		}
		privateKey, err := getDecryptedConfigValue(dbSuper, "wompi.private_key")
		if err != nil {
			http.Error(w, "failed to read wompi.private_key: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if strings.TrimSpace(publicKey) == "" {
			http.Error(w, "wompi.public_key not configured", http.StatusInternalServerError)
			return
		}

		mode, _ := resolveWompiMode(dbSuper, publicKey, privateKey)
		baseURL := wompiBaseURLFromMode(mode)
		statusURL := strings.TrimRight(baseURL, "/") + "/transactions/" + url.PathEscape(transactionID)

		fetchStatus := func(authKey string) ([]byte, int, error) {
			req, err := http.NewRequest("GET", statusURL, nil)
			if err != nil {
				return nil, 0, err
			}
			req.Header.Set("Authorization", "Bearer "+authKey)
			client := &http.Client{Timeout: 15 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				return nil, 0, err
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			return body, resp.StatusCode, nil
		}

		respBody, statusCode, err := fetchStatus(publicKey)
		if err != nil {
			http.Error(w, "request error: "+err.Error(), http.StatusBadGateway)
			return
		}
		if statusCode >= 400 && strings.TrimSpace(privateKey) != "" {
			if body2, code2, err2 := fetchStatus(privateKey); err2 == nil {
				respBody = body2
				statusCode = code2
			}
		}
		if statusCode >= 400 {
			http.Error(w, "wompi API error: "+string(respBody), http.StatusBadGateway)
			return
		}

		var wompiResp map[string]interface{}
		if err := json.Unmarshal(respBody, &wompiResp); err != nil {
			http.Error(w, "invalid response from wompi: "+err.Error(), http.StatusInternalServerError)
			return
		}
		data, _ := wompiResp["data"].(map[string]interface{})
		status := strings.ToUpper(strings.TrimSpace(fmt.Sprint(data["status"])))
		if refFromGateway := strings.TrimSpace(fmt.Sprint(data["reference"])); refFromGateway != "" && refFromGateway != "<nil>" {
			reference = refFromGateway
		}

		if err := dbpkg.UpdateWompiPaymentRecordByTransaction(dbSuper, transactionID, status, string(respBody)); err != nil {
			log.Println("warning: failed to update Wompi payment record:", err)
		}

		licenciaID, empresaID, hasContext, ctxErr := dbpkg.GetWompiPaymentContext(dbSuper, transactionID, reference)
		if ctxErr != nil {
			log.Println("warning: failed to resolve Wompi payment context:", ctxErr)
		}
		if !hasContext {
			licenciaID, empresaID, hasContext = paymentContextFromInternalReference(reference, transactionID)
		}
		if hasExpectedContext && hasContext && !paymentContextMatchesExpected(licenciaID, empresaID, expectedLicenciaID, expectedEmpresaID) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":               "El pago consultado no corresponde a la empresa o licencia abierta en esta página.",
				"provider":            "wompi",
				"transaction_id":      transactionID,
				"reference":           reference,
				"status":              status,
				"context_found":       hasContext,
				"context_mismatch":    true,
				"licencia_id":         licenciaID,
				"empresa_id":          empresaID,
				"expected_licencia_id": expectedLicenciaID,
				"expected_empresa_id":  expectedEmpresaID,
			})
			return
		}

		activated := false
		if strings.EqualFold(status, "APPROVED") && hasContext {
			lic, err := dbpkg.GetLicenciaByID(dbSuper, licenciaID)
			if err == nil && lic != nil {
				now := time.Now()
				fechaInicio := now.Format("2006-01-02 15:04:05")
				fechaFin := now.AddDate(0, 0, lic.DuracionDias).Format("2006-01-02 15:04:05")
				if err := dbpkg.ActivateLicenciaForEmpresa(dbSuper, licenciaID, empresaID, fechaInicio, fechaFin); err != nil {
					log.Println("failed to activate licencia from Wompi:", err)
				} else {
					activated = true
				}

				// Registrar comisiones para vendedor_de_licencia si aplica
				go func(txID string, licID, empID int64) {
					if txID == "" {
						return
					}
					recordVendedorComisiones(dbSuper, txID, "", licID, empID)
				}(transactionID, licenciaID, empresaID)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"provider":       "wompi",
			"mode":           mode,
			"transaction_id": transactionID,
			"reference":      reference,
			"status":         status,
			"context_found":  hasContext,
			"licencia_id":    licenciaID,
			"empresa_id":     empresaID,
			"activated":      activated,
			"data":           data,
		})
	}
}

// WompiWebhookHandler procesa notificaciones servidor-servidor de Wompi.
func WompiWebhookHandler(dbSuper *sql.DB, dbEmp ...*sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		var obj map[string]interface{}
		if err := json.Unmarshal(body, &obj); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		if err := verifyWompiWebhookSignature(dbSuper, r, body, obj); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		transactionID, reference, status := extractWompiWebhookPaymentInfo(obj)
		if strings.TrimSpace(transactionID) == "" && strings.TrimSpace(reference) == "" {
			http.Error(w, "transaction_id or reference required", http.StatusBadRequest)
			return
		}
		if status == "" {
			status = "PENDING"
		}

		if transactionID != "" {
			if err := dbpkg.UpdateWompiPaymentRecordByTransaction(dbSuper, transactionID, status, string(body)); err != nil {
				log.Println("warning: failed to update Wompi record by transaction_id:", err)
			}
		}
		if reference != "" {
			if err := dbpkg.UpdateWompiPaymentRecordByReference(dbSuper, reference, status, string(body)); err != nil {
				log.Println("warning: failed to update Wompi record by reference:", err)
			}
		}

		licenciaID, empresaID, hasContext, ctxErr := dbpkg.GetWompiPaymentContext(dbSuper, transactionID, reference)
		if ctxErr != nil {
			log.Println("warning: failed to resolve Wompi payment context:", ctxErr)
		}

		activated := false
		if isApprovedPaymentStatus(status) && hasContext {
			act, actErr := activateLicenciaByIDs(dbSuper, licenciaID, empresaID)
			if actErr != nil {
				log.Println("warning: failed to activate licencia from Wompi webhook:", actErr)
			} else {
				activated = act
			}

			// Registrar comisiones para vendedor_de_licencia si aplica (webhook puede venir sin transaction but with reference)
			go func(txID, ref string, licID, empID int64) {
				recordVendedorComisiones(dbSuper, txID, ref, licID, empID)
			}(transactionID, reference, licenciaID, empresaID)
		}

		ventaDigitalContextFound := false
		ventaDigitalDeliverySent := false
		ventaDigitalDeliveryStage := "not_processed"
		ventaPublicaContextFound := false
		if strings.TrimSpace(status) != "" {
			foundVD, deliveredVD, deliveryStageVD, vdErr := processVentaDigitalPaymentStatusUpdate(r, dbSuper, transactionID, reference, status, string(body))
			ventaDigitalContextFound = foundVD
			if strings.TrimSpace(deliveryStageVD) != "" {
				ventaDigitalDeliveryStage = deliveryStageVD
			}
			if vdErr != nil {
				log.Println("warning: failed to process venta_digital webhook update:", vdErr)
			} else {
				ventaDigitalDeliverySent = deliveredVD
			}
			if len(dbEmp) > 0 && dbEmp[0] != nil {
				foundVP, vpErr := processVentaPublicaPaymentStatusUpdate(dbEmp[0], "wompi", transactionID, reference, status, string(body))
				ventaPublicaContextFound = foundVP
				if vpErr != nil {
					log.Println("warning: failed to process venta_publica webhook update:", vpErr)
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":                           true,
			"provider":                     "wompi",
			"transaction_id":               transactionID,
			"reference":                    reference,
			"status":                       status,
			"context_found":                hasContext,
			"licencia_id":                  licenciaID,
			"empresa_id":                   empresaID,
			"activated":                    activated,
			"venta_digital_context_found":  ventaDigitalContextFound,
			"venta_digital_delivery_sent":  ventaDigitalDeliverySent,
			"venta_digital_delivery_stage": ventaDigitalDeliveryStage,
			"venta_publica_context_found":  ventaPublicaContextFound,
		})
	}
}

// EpaycoCreateTransactionHandler prepara checkout de Epayco y registra transaccion pendiente.
func EpaycoCreateTransactionHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			LicenciaID    int64  `json:"licencia_id"`
			EmpresaID     int64  `json:"empresa_id,omitempty"`
			CustomerEmail string `json:"customer_email,omitempty"`
			DiscountCode  string `json:"discount_code,omitempty"`
			AsesorID      string `json:"asesor_id,omitempty"`
			VendedorID    string `json:"vendedor_id,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
			return
		}
		if payload.LicenciaID <= 0 {
			http.Error(w, "licencia_id invalido", http.StatusBadRequest)
			return
		}

		enabledRaw, _, err := dbpkg.GetConfigValue(dbSuper, "epayco.enabled")
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "failed to read epayco.enabled: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if !parseBoolConfigValue(enabledRaw) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPreconditionFailed)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":    "Epayco no esta activo en configuracion avanzada",
				"provider": "epayco",
			})
			return
		}

		publicKey, _, privateKey, err := resolveEpaycoCredentials(dbSuper)
		if err != nil {
			http.Error(w, "failed to read Epayco credentials: "+err.Error(), http.StatusInternalServerError)
			return
		}
		publicKey = strings.TrimSpace(publicKey)
		privateKey = strings.TrimSpace(privateKey)
		if publicKey == "" {
			http.Error(w, "epayco.public_key no configurada", http.StatusInternalServerError)
			return
		}

		lic, err := dbpkg.GetLicenciaByID(dbSuper, payload.LicenciaID)
		if err != nil || lic == nil {
			http.Error(w, "licencia not found", http.StatusBadRequest)
			return
		}

		email := strings.TrimSpace(payload.CustomerEmail)
		if email == "" {
			email = strings.TrimSpace(r.Header.Get("X-Admin-Email"))
		}

		paymentBaseURL, err := resolvePaymentBaseURL(r, dbSuper)
		if err != nil {
			http.Error(w, "failed to resolve public base URL for Epayco: "+err.Error(), http.StatusPreconditionFailed)
			return
		}

		if privateKey == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPreconditionFailed)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":    "Epayco requiere Public Key y Private Key para Smart Checkout",
				"provider": "epayco",
			})
			return
		}

		mode, modeSource := resolveEpaycoMode(dbSuper, publicKey, privateKey)
		reference := fmt.Sprintf("EPAYCO-LIC-%d-EMP-%d-%d", payload.LicenciaID, payload.EmpresaID, time.Now().UnixNano())
		responseURL := buildEpaycoResponseURL(paymentBaseURL, "pending", reference, payload.LicenciaID, payload.EmpresaID)
		confirmationURL := strings.TrimRight(strings.TrimSpace(paymentBaseURL), "/") + "/epayco/webhook"
		sessionPayload := map[string]interface{}{
			"checkout_version": "2",
			"name":             "Powerful Control System",
			"description":      "Pago de licencia " + strings.TrimSpace(lic.Nombre),
			"currency":         "COP",
			"amount":           lic.Valor,
			"lang":             "ES",
			"invoice":          reference,
			"country":          "CO",
			"taxBase":          0,
			"tax":              0,
			"taxIco":           0,
			"response":         responseURL,
			"confirmation":     confirmationURL,
			"method":           "POST",
			"extras": map[string]interface{}{
				"extra1": strconv.FormatInt(payload.LicenciaID, 10),
				"extra2": strconv.FormatInt(payload.EmpresaID, 10),
				"extra3": reference,
			},
		}
		if email != "" {
			sessionPayload["billing"] = map[string]interface{}{
				"email": email,
			}
		}

		apifyToken, loginRaw, err := fetchEpaycoApifyToken(publicKey, privateKey)
		if err != nil {
			http.Error(w, "failed to authenticate with Epayco Smart Checkout: "+err.Error(), http.StatusBadGateway)
			return
		}
		sessionID, sessionRaw, err := createEpaycoSmartCheckoutSession(apifyToken, sessionPayload)
		if err != nil {
			http.Error(w, "failed to create Epayco Smart Checkout session: "+err.Error(), http.StatusBadGateway)
			return
		}

		if strings.TrimSpace(payload.AsesorID) == "" && strings.TrimSpace(payload.VendedorID) != "" {
			payload.AsesorID = payload.VendedorID
		}

		rawMap := map[string]interface{}{
			"provider":         "epayco",
			"mode":             mode,
			"mode_source":      modeSource,
			"payment_base_url": paymentBaseURL,
			"checkout_type":    "standard",
			"checkout_script":  epaycoSmartCheckoutScriptURL,
			"session_id":       sessionID,
			"response":         responseURL,
			"confirmation":     confirmationURL,
			"license_id":       payload.LicenciaID,
			"empresa_id":       payload.EmpresaID,
			"customer_email":   email,
			"discount_code":    payload.DiscountCode,
			"asesor_id":        payload.AsesorID,
			"created_at":       time.Now().Format(time.RFC3339),
			"integration_flow": "smart_checkout_v2",
			"apify_login_raw":  loginRaw,
			"session_raw":      sessionRaw,
		}
		rawBytes, _ := json.Marshal(rawMap)
		if _, err := dbpkg.CreateEpaycoPaymentRecord(dbSuper, payload.LicenciaID, payload.EmpresaID, reference, reference, "PENDING", string(rawBytes), payload.DiscountCode, payload.AsesorID); err != nil {
			log.Println("warning: failed to record Epayco transaction in DB:", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"provider":       "epayco",
			"payment_method": "SMART_CHECKOUT",
			"mode":           mode,
			"mode_source":    modeSource,
			"transaction_id": reference,
			"reference":      reference,
			"status":         "PENDING",
			"session_id":     sessionID,
			"checkout_session_id": sessionID,
			"checkout_type":      "standard",
			"checkout_script_url": epaycoSmartCheckoutScriptURL,
			"customer_email": email,
			"data": map[string]interface{}{
				"id":         reference,
				"reference":  reference,
				"sessionId":  sessionID,
				"type":       "standard",
				"script_url": epaycoSmartCheckoutScriptURL,
			},
		})
	}
}

// EpaycoTransactionStatusHandler consulta estado por referencia y activa licencia si aplica.
func EpaycoTransactionStatusHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		transactionID := strings.TrimSpace(r.URL.Query().Get("id"))
		if transactionID == "" {
			transactionID = strings.TrimSpace(r.URL.Query().Get("transaction_id"))
		}
		reference := strings.TrimSpace(r.URL.Query().Get("reference"))
		if reference == "" {
			reference = strings.TrimSpace(r.URL.Query().Get("ref"))
		}
		expectedLicenciaID, expectedEmpresaID, hasExpectedContext := expectedPaymentContextFromRequest(r)
		originalTransactionID := transactionID
		originalReference := reference
		if transactionID == "" && reference == "" {
			http.Error(w, "id o reference requerido", http.StatusBadRequest)
			return
		}

		var rec *dbpkg.EpaycoPaymentRecord
		var err error
		if transactionID != "" {
			rec, err = dbpkg.GetEpaycoPaymentByTransaction(dbSuper, transactionID)
			if err != nil {
				http.Error(w, "failed to read pagos_epayco: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if rec == nil && reference != "" {
			rec, err = dbpkg.GetEpaycoPaymentByReference(dbSuper, reference)
			if err != nil {
				http.Error(w, "failed to read pagos_epayco: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		storedStatus := ""
		recordTransactionID := transactionID
		recordReference := reference
		if rec != nil {
			if reference == "" && rec.Reference.Valid {
				reference = strings.TrimSpace(rec.Reference.String)
			}
			if transactionID == "" && rec.TransactionID.Valid {
				transactionID = strings.TrimSpace(rec.TransactionID.String)
			}
			if rec.Status.Valid {
				storedStatus = strings.ToUpper(strings.TrimSpace(rec.Status.String))
			}
			if recordReference == "" && rec.Reference.Valid {
				recordReference = strings.TrimSpace(rec.Reference.String)
			}
			if recordTransactionID == "" && rec.TransactionID.Valid {
				recordTransactionID = strings.TrimSpace(rec.TransactionID.String)
			}
		}
		if recordReference == "" {
			recordReference = recordTransactionID
		}
		if recordTransactionID == "" {
			recordTransactionID = recordReference
		}
		if reference == "" {
			reference = recordReference
		}
		if transactionID == "" {
			transactionID = recordTransactionID
		}

		status := ""
		validationPayload := map[string]interface{}{}
		rawValidation := ""
		gatewayTransactionID := ""
		gatewayReference := ""
		invoiceReference := ""
		if recordReference != "" {
			validationURL := "https://secure.epayco.co/validation/v1/reference/" + url.PathEscape(recordReference)
			req, err := http.NewRequest("GET", validationURL, nil)
			if err == nil {
				client := &http.Client{Timeout: 15 * time.Second}
				resp, reqErr := client.Do(req)
				if reqErr == nil {
					defer resp.Body.Close()
					body, _ := io.ReadAll(resp.Body)
					rawValidation = string(body)
					if resp.StatusCode < 400 {
						if err := json.Unmarshal(body, &validationPayload); err == nil {
							status = parseEpaycoPaymentStatus(validationPayload)
							if status == "ERROR" && shouldPreservePendingEpaycoStatus(storedStatus, validationPayload) {
								status = "PENDING"
							}
							if txFromGateway := strings.TrimSpace(pickEpaycoField(validationPayload, "x_transaction_id", "transaction_id", "id")); txFromGateway != "" {
								gatewayTransactionID = txFromGateway
								transactionID = txFromGateway
							}
							if invoiceFromGateway := strings.TrimSpace(pickEpaycoField(validationPayload, "invoice", "x_id_invoice")); invoiceFromGateway != "" {
								invoiceReference = invoiceFromGateway
								recordReference = invoiceFromGateway
							}
							if refFromGateway := strings.TrimSpace(pickEpaycoField(validationPayload, "x_ref_payco", "reference", "ref_payco")); refFromGateway != "" {
								gatewayReference = refFromGateway
								reference = refFromGateway
							}
						}
					} else {
						log.Printf("warning: epayco validation API returned %s for reference %s: %s", resp.Status, recordReference, string(body))
					}
				} else {
					log.Printf("warning: epayco validation request failed for reference %s: %v", recordReference, reqErr)
				}
			}
		}

		if strings.TrimSpace(status) == "" {
			status = strings.TrimSpace(storedStatus)
		}
		if strings.TrimSpace(status) == "" {
			status = "PENDING"
		}
		status = strings.ToUpper(status)
		if transactionID == "" {
			transactionID = firstNonEmptyString(gatewayTransactionID, recordTransactionID, originalTransactionID)
		}
		if reference == "" {
			reference = firstNonEmptyString(gatewayReference, invoiceReference, recordReference, originalReference)
		}
		if rec == nil {
			rec, err = findEpaycoPaymentRecordByCandidates(dbSuper, []string{recordTransactionID, transactionID, originalTransactionID}, []string{recordReference, invoiceReference, reference, originalReference})
			if err != nil {
				log.Println("warning: failed to reload Epayco payment record before payload merge:", err)
				rec = nil
			}
		}

		payloadToSave := rawValidation
		if strings.TrimSpace(payloadToSave) == "" {
			fallbackPayload, _ := json.Marshal(map[string]interface{}{
				"provider":       "epayco",
				"transaction_id": transactionID,
				"reference":      reference,
				"status":         status,
			})
			payloadToSave = string(fallbackPayload)
		}
		if rec != nil && rec.RawPayload.Valid {
			payloadToSave = mergePaymentPayloadJSON(rec.RawPayload.String, payloadToSave)
		}

		if recordTransactionID != "" {
			if err := dbpkg.UpdateEpaycoPaymentRecordByTransaction(dbSuper, recordTransactionID, status, payloadToSave); err != nil {
				log.Println("warning: failed to update Epayco payment by transaction:", err)
			}
		}
		if recordReference != "" {
			if err := dbpkg.UpdateEpaycoPaymentRecordByReference(dbSuper, recordReference, status, payloadToSave); err != nil {
				log.Println("warning: failed to update Epayco payment by reference:", err)
			}
		}
		if invoiceReference != "" && invoiceReference != recordReference {
			if err := dbpkg.UpdateEpaycoPaymentRecordByReference(dbSuper, invoiceReference, status, payloadToSave); err != nil {
				log.Println("warning: failed to update Epayco payment by invoice reference:", err)
			}
		}

		licenciaID := int64(0)
		empresaID := int64(0)
		hasContext := false
		lookupCombos := [][2]string{
			{recordTransactionID, recordReference},
			{"", recordReference},
			{"", invoiceReference},
			{"", originalReference},
			{originalTransactionID, originalReference},
			{transactionID, reference},
		}
		licenciaID, empresaID, hasContext = resolveEpaycoPaymentContextCandidates(dbSuper, lookupCombos)
		if !hasContext {
			licenciaID, empresaID, hasContext = paymentContextFromInternalReference(recordReference, invoiceReference, reference, transactionID, recordTransactionID, originalReference, originalTransactionID)
		}
		if hasExpectedContext && hasContext && !paymentContextMatchesExpected(licenciaID, empresaID, expectedLicenciaID, expectedEmpresaID) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":                "El pago consultado no corresponde a la empresa o licencia abierta en esta página.",
				"provider":             "epayco",
				"transaction_id":       firstNonEmptyString(transactionID, recordTransactionID, originalTransactionID),
				"reference":            firstNonEmptyString(reference, recordReference, invoiceReference, originalReference),
				"status":               status,
				"context_found":        hasContext,
				"context_mismatch":     true,
				"licencia_id":          licenciaID,
				"empresa_id":           empresaID,
				"expected_licencia_id": expectedLicenciaID,
				"expected_empresa_id":  expectedEmpresaID,
				"data":                 validationPayload,
			})
			return
		}

		activated := false
		if isApprovedPaymentStatus(status) && hasContext {
			act, actErr := activateLicenciaByIDs(dbSuper, licenciaID, empresaID)
			if actErr != nil {
				log.Println("warning: failed to activate licencia from Epayco status:", actErr)
			} else {
				activated = act
				lic, licErr := dbpkg.GetLicenciaByID(dbSuper, licenciaID)
				if licErr != nil {
					log.Println("warning: failed to reload licencia after Epayco activation:", licErr)
				} else {
					payRec, recErr := findEpaycoPaymentRecordByCandidates(dbSuper, []string{recordTransactionID, transactionID, originalTransactionID}, []string{recordReference, invoiceReference, reference, originalReference})
					if recErr != nil {
						log.Println("warning: failed to reload Epayco payment for activation email:", recErr)
					} else if mailErr := trySendLicenciaActivationEmail(r, dbSuper, empresaID, lic, payRec, "epayco", firstNonEmptyString(recordReference, invoiceReference, reference, originalReference)); mailErr != nil {
						log.Println("warning: failed to send licencia activation email for Epayco status:", mailErr)
					}
				}
			}
			go func(txID, ref string, licID, empID int64) {
				recordVendedorComisionesEpayco(dbSuper, txID, ref, licID, empID)
			}(firstNonEmptyString(recordTransactionID, transactionID), firstNonEmptyString(recordReference, invoiceReference, reference), licenciaID, empresaID)
		}

		publicKey, _, privateKey, _ := resolveEpaycoCredentials(dbSuper)
		mode, modeSource := resolveEpaycoMode(dbSuper, publicKey, privateKey)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"provider":       "epayco",
			"mode":           mode,
			"mode_source":    modeSource,
			"transaction_id": firstNonEmptyString(transactionID, recordTransactionID, originalTransactionID),
			"reference":      firstNonEmptyString(reference, recordReference, invoiceReference, originalReference),
			"status":         status,
			"context_found":  hasContext,
			"licencia_id":    licenciaID,
			"empresa_id":     empresaID,
			"activated":      activated,
			"data":           validationPayload,
		})
	}
}

// EpaycoWebhookHandler procesa confirmaciones de Epayco por formulario o JSON.
func EpaycoWebhookHandler(dbSuper *sql.DB, dbEmp ...*sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		payload := map[string]interface{}{}
		rawPayload := ""
		contentType := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))

		if strings.Contains(contentType, "application/json") {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read body", http.StatusBadRequest)
				return
			}
			rawPayload = string(body)
			if len(body) > 0 {
				if err := json.Unmarshal(body, &payload); err != nil {
					http.Error(w, "invalid payload", http.StatusBadRequest)
					return
				}
			}
		} else {
			if err := r.ParseForm(); err == nil {
				for key, values := range r.Form {
					if len(values) == 0 {
						continue
					}
					payload[key] = values[0]
				}
			}
		}

		if len(payload) == 0 {
			q := r.URL.Query()
			for key, values := range q {
				if len(values) == 0 {
					continue
				}
				payload[key] = values[0]
			}
		}

		if strings.TrimSpace(rawPayload) == "" {
			if rawBytes, err := json.Marshal(payload); err == nil {
				rawPayload = string(rawBytes)
			}
		}

		transactionID, reference, status := extractEpaycoPaymentInfo(payload)
		invoiceReference := strings.TrimSpace(pickEpaycoField(payload, "invoice", "x_id_invoice"))
		if transactionID == "" && reference == "" {
			http.Error(w, "transaction_id o reference requerido", http.StatusBadRequest)
			return
		}

		rec, recErr := findEpaycoPaymentRecordByCandidates(dbSuper, []string{transactionID}, []string{reference, invoiceReference})
		if recErr != nil {
			log.Println("warning: failed to load Epayco webhook record before update:", recErr)
		}
		payloadToSave := rawPayload
		if rec != nil && rec.RawPayload.Valid {
			payloadToSave = mergePaymentPayloadJSON(rec.RawPayload.String, rawPayload)
		}

		if transactionID != "" {
			if err := dbpkg.UpdateEpaycoPaymentRecordByTransaction(dbSuper, transactionID, status, payloadToSave); err != nil {
				log.Println("warning: failed to update Epayco webhook by transaction:", err)
			}
		}
		if reference != "" {
			if err := dbpkg.UpdateEpaycoPaymentRecordByReference(dbSuper, reference, status, payloadToSave); err != nil {
				log.Println("warning: failed to update Epayco webhook by reference:", err)
			}
		}
		if invoiceReference != "" && invoiceReference != reference {
			if err := dbpkg.UpdateEpaycoPaymentRecordByReference(dbSuper, invoiceReference, status, payloadToSave); err != nil {
				log.Println("warning: failed to update Epayco webhook by invoice reference:", err)
			}
		}

		licenciaID, empresaID, hasContext := resolveEpaycoPaymentContextCandidates(dbSuper, [][2]string{{transactionID, reference}, {"", reference}, {"", invoiceReference}, {transactionID, invoiceReference}})

		activated := false
		ventaPublicaContextFound := false
		if isApprovedPaymentStatus(status) && hasContext {
			act, actErr := activateLicenciaByIDs(dbSuper, licenciaID, empresaID)
			if actErr != nil {
				log.Println("warning: failed to activate licencia from Epayco webhook:", actErr)
			} else {
				activated = act
				lic, licErr := dbpkg.GetLicenciaByID(dbSuper, licenciaID)
				if licErr != nil {
					log.Println("warning: failed to reload licencia after Epayco webhook activation:", licErr)
				} else {
					payRec, payErr := findEpaycoPaymentRecordByCandidates(dbSuper, []string{transactionID}, []string{reference, invoiceReference})
					if payErr != nil {
						log.Println("warning: failed to reload Epayco payment for webhook activation email:", payErr)
					} else if mailErr := trySendLicenciaActivationEmail(r, dbSuper, empresaID, lic, payRec, "epayco", firstNonEmptyString(invoiceReference, reference)); mailErr != nil {
						log.Println("warning: failed to send licencia activation email for Epayco webhook:", mailErr)
					}
				}
			}
			go func(txID, ref string, licID, empID int64) {
				recordVendedorComisionesEpayco(dbSuper, txID, ref, licID, empID)
			}(transactionID, firstNonEmptyString(invoiceReference, reference), licenciaID, empresaID)
		}
		if len(dbEmp) > 0 && dbEmp[0] != nil {
			foundVP, vpErr := processVentaPublicaPaymentStatusUpdate(dbEmp[0], "epayco", transactionID, firstNonEmptyString(invoiceReference, reference), status, payloadToSave)
			ventaPublicaContextFound = foundVP
			if vpErr != nil {
				log.Println("warning: failed to process venta_publica Epayco webhook update:", vpErr)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":             true,
			"provider":       "epayco",
			"transaction_id": transactionID,
			"reference":      reference,
			"status":         status,
			"context_found":  hasContext,
			"licencia_id":    licenciaID,
			"empresa_id":     empresaID,
			"activated":      activated,
			"venta_publica_context_found": ventaPublicaContextFound,
		})
	}
}

// ActivateLicenciaSinPagoHandler activa una licencia manualmente para avanzar en pruebas internas.
func ActivateLicenciaSinPagoHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			LicenciaID   int64  `json:"licencia_id"`
			EmpresaID    int64  `json:"empresa_id"`
			Motivo       string `json:"motivo,omitempty"`
			DiscountCode string `json:"discount_code,omitempty"`
			AsesorID     string `json:"asesor_id,omitempty"`
			VendedorID   string `json:"vendedor_id,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
			return
		}
		if payload.LicenciaID <= 0 {
			http.Error(w, "licencia_id inválido", http.StatusBadRequest)
			return
		}
		if payload.EmpresaID <= 0 {
			http.Error(w, "empresa_id inválido", http.StatusBadRequest)
			return
		}

		lic, err := dbpkg.GetLicenciaByID(dbSuper, payload.LicenciaID)
		if err != nil || lic == nil {
			http.Error(w, "licencia not found", http.StatusBadRequest)
			return
		}

		now := time.Now()
		fechaInicio := now.Format("2006-01-02 15:04:05")
		fechaFin := now.AddDate(0, 0, lic.DuracionDias).Format("2006-01-02 15:04:05")
		if err := dbpkg.ActivateLicenciaForEmpresa(dbSuper, payload.LicenciaID, payload.EmpresaID, fechaInicio, fechaFin); err != nil {
			http.Error(w, "failed to activate licencia: "+err.Error(), http.StatusInternalServerError)
			return
		}

		log.Printf("Licencia activada sin pago: licencia=%d empresa=%d motivo=%q", payload.LicenciaID, payload.EmpresaID, payload.Motivo)

		// Registrar la activación en pagos_wompi para trazabilidad (provider=MANUAL)
		go func() {
			reference := fmt.Sprintf("MANUAL-LIC-%d-EMP-%d-%d", payload.LicenciaID, payload.EmpresaID, time.Now().UnixNano())

			// compatibilidad: si se envió `vendedor_id` usarlo como `asesor_id` si este último está vacío
			if strings.TrimSpace(payload.AsesorID) == "" && strings.TrimSpace(payload.VendedorID) != "" {
				payload.AsesorID = payload.VendedorID
			}

			rawMap := map[string]interface{}{"motivo": payload.Motivo, "discount_code": payload.DiscountCode, "asesor_id": payload.AsesorID, "vendedor_id": payload.VendedorID}
			rawBytes, _ := json.Marshal(rawMap)
			if _, err := dbpkg.CreateWompiPaymentRecord(dbSuper, payload.LicenciaID, payload.EmpresaID, "", reference, "MANUAL", string(rawBytes), payload.DiscountCode, payload.AsesorID); err != nil {
				log.Println("warning: failed to record manual activation in pagos_wompi:", err)
			}

			// Registrar comisiones si el payload contiene codigo de vendedor
			recordVendedorComisiones(dbSuper, "", reference, payload.LicenciaID, payload.EmpresaID)
		}()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"activated":      true,
			"provider":       "manual",
			"payment_method": "ACTIVAR_SIN_PAGO",
			"licencia_id":    payload.LicenciaID,
			"empresa_id":     payload.EmpresaID,
			"fecha_inicio":   fechaInicio,
			"fecha_fin":      fechaFin,
			"redirect_url":   fmt.Sprintf("/administrar_empresa.html?id=%d", payload.EmpresaID),
		})
	}
}
