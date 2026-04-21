package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

const adminEmpresaCompartidaMailNotificationType = "empresa_admin_share_invitation"

func hashAdminEmpresaCompartidaToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

func newAdminEmpresaCompartidaInvitationTokenAndExpiration() (string, string, string, error) {
	token, err := utils.GenerateSecureToken(32)
	if err != nil {
		return "", "", "", err
	}
	expira := time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339)
	return token, hashAdminEmpresaCompartidaToken(token), expira, nil
}

func parseAdminEmpresaCompartidaTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	layouts := []string{time.RFC3339, "2006-01-02 15:04:05", time.RFC3339Nano}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func isAdminEmpresaCompartidaInvitationExpired(item *dbpkg.AdminEmpresaCompartidaInvitacion) bool {
	if item == nil {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(item.Estado), "pendiente") {
		return false
	}
	expira, ok := parseAdminEmpresaCompartidaTime(item.ExpiraEn)
	if !ok {
		return false
	}
	return time.Now().After(expira)
}

func normalizeAdminEmpresaCompartidaInvitation(item dbpkg.AdminEmpresaCompartidaInvitacion) dbpkg.AdminEmpresaCompartidaInvitacion {
	if isAdminEmpresaCompartidaInvitationExpired(&item) {
		item.Estado = "expirada"
	}
	item.TokenHash = ""
	return item
}

func decorateEmpresaAccessForRequester(dbSuper *sql.DB, requesterEmail, principalEmail string, empresa *dbpkg.Empresa) error {
	if empresa == nil {
		return nil
	}
	owner, err := empresaBelongsToPrincipalScope(dbSuper, principalEmail, empresa.UsuarioCreador)
	if err != nil {
		return err
	}
	if owner {
		empresa.AccessSource = "owner"
		empresa.CompartidaPor = ""
		return nil
	}
	access, err := dbpkg.GetActiveAdminEmpresaCompartidaAcceso(dbSuper, empresa.EmpresaID, requesterEmail)
	if err != nil {
		return err
	}
	if access != nil {
		empresa.AccessSource = "shared"
		empresa.CompartidaPor = strings.TrimSpace(access.CompartidoPorEmail)
	}
	return nil
}

func decorateEmpresasByEffectiveAccess(dbSuper *sql.DB, requesterEmail, principalEmail string, empresas []dbpkg.Empresa) ([]dbpkg.Empresa, error) {
	if len(empresas) == 0 {
		return empresas, nil
	}
	shareMap := map[int64]dbpkg.AdminEmpresaCompartidaAcceso{}
	if strings.TrimSpace(requesterEmail) != "" {
		shares, err := dbpkg.ListActiveAdminEmpresaCompartidaAccesosByAdmin(dbSuper, requesterEmail)
		if err != nil {
			return nil, err
		}
		for _, share := range shares {
			shareMap[share.EmpresaID] = share
		}
	}
	out := make([]dbpkg.Empresa, 0, len(empresas))
	for _, empresa := range empresas {
		owner, err := empresaBelongsToPrincipalScope(dbSuper, principalEmail, empresa.UsuarioCreador)
		if err != nil {
			return nil, err
		}
		if owner {
			empresa.AccessSource = "owner"
			out = append(out, empresa)
			continue
		}
		share, ok := shareMap[empresa.EmpresaID]
		if ok {
			empresa.AccessSource = "shared"
			empresa.CompartidaPor = strings.TrimSpace(share.CompartidoPorEmail)
			out = append(out, empresa)
		}
	}
	return out, nil
}

func ensureEmpresaOwnerAccess(dbEmp, dbSuper *sql.DB, r *http.Request, empresaID int64) (*dbpkg.Empresa, string, bool, error) {
	_, principalEmail, err := resolveRequesterAdminScope(dbSuper, r)
	if err != nil {
		return nil, "", false, err
	}
	if empresaID <= 0 {
		return nil, principalEmail, false, fmt.Errorf("empresa_id invalido")
	}
	empresa, err := dbpkg.GetEmpresaByID(dbEmp, empresaID)
	if err != nil {
		return nil, principalEmail, false, err
	}
	owner, err := empresaBelongsToPrincipalScope(dbSuper, principalEmail, empresa.UsuarioCreador)
	if err != nil {
		return nil, principalEmail, false, err
	}
	return empresa, principalEmail, owner, nil
}

func sendAdminEmpresaCompartidaInvitationEmail(r *http.Request, dbEmp, dbSuper *sql.DB, empresa *dbpkg.Empresa, inviter *dbpkg.Admin, toEmail, toName, token, mensaje string) (string, error) {
	baseURL := resolveBaseURLForConfirmation(r, dbSuper)
	loginURL := strings.TrimRight(baseURL, "/") + "/login.html?shared_invitation_token=" + url.QueryEscape(token)
	acceptURL := loginURL
	companyName := "la empresa"
	if empresa != nil && strings.TrimSpace(empresa.Nombre) != "" {
		companyName = strings.TrimSpace(empresa.Nombre)
	}
	inviterName := "Un administrador"
	if inviter != nil && strings.TrimSpace(inviter.Name) != "" {
		inviterName = strings.TrimSpace(inviter.Name)
	} else if inviter != nil && strings.TrimSpace(inviter.Email) != "" {
		inviterName = strings.TrimSpace(inviter.Email)
	}
	recipientName := strings.TrimSpace(toName)
	if recipientName == "" {
		recipientName = "administrador"
	}
	mensaje = strings.TrimSpace(mensaje)
	asunto, cuerpoPlano, cuerpoHTML, err := applySuperEmailTemplate(dbSuper, superEmailTemplateKeyEmpresaAdminShareInvite, map[string]string{
		"name":                     recipientName,
		"company_name":             companyName,
		"invited_by_name":          inviterName,
		"accept_url":               acceptURL,
		"login_url":                loginURL,
		"admin_message":            mensaje,
		"admin_message_block_text": templateParagraphText("Mensaje del administrador:", mensaje),
		"admin_message_block_html": templateParagraphHTML("Mensaje del administrador:", mensaje),
	})
	if err != nil {
		return acceptURL, err
	}
	if isEmpresaUsuarioMailTestMode(dbSuper) {
		metadataJSON := fmt.Sprintf(`{"accept_url":%q,"login_url":%q,"empresa_id":%d,"mail_mode":"test","mensaje":%q}`, acceptURL, loginURL, empresa.EmpresaID, mensaje)
		if err := captureEmpresaUsuarioMailNotification(dbSuper, adminEmpresaCompartidaMailNotificationType, empresa.EmpresaID, toEmail, asunto, cuerpoPlano, token, metadataJSON, adminEmailFromRequest(r)); err != nil {
			return acceptURL, err
		}
		return acceptURL, nil
	}
	smtpEmail, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_email")
	if err != nil {
		return acceptURL, err
	}
	smtpPass, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_app_password")
	if err != nil {
		return acceptURL, err
	}
	smtpEmail = strings.TrimSpace(smtpEmail)
	smtpPass = strings.TrimSpace(smtpPass)
	if smtpEmail == "" || smtpPass == "" {
		return acceptURL, fmt.Errorf("smtp gmail no configurado")
	}
	smtpHost, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_host")
	smtpPort, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_port")
	fromName, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_from_name")
	smtpHost = strings.TrimSpace(smtpHost)
	if smtpHost == "" {
		smtpHost = "smtp.gmail.com"
	}
	smtpPort = strings.TrimSpace(smtpPort)
	if smtpPort == "" {
		smtpPort = "587"
	}
	fromName = strings.TrimSpace(fromName)
	if fromName == "" {
		fromName = "Powerful Control System"
	}
	mailHostForAuth := smtpHost
	if strings.Contains(smtpHost, ":") {
		if h, _, splitErr := net.SplitHostPort(smtpHost); splitErr == nil {
			mailHostForAuth = h
		}
	}
	addr := smtpHost
	if !strings.Contains(addr, ":") {
		addr = smtpHost + ":" + smtpPort
	}
	auth := smtp.PlainAuth("", smtpEmail, smtpPass, mailHostForAuth)
	boundary := "==PCS_BOUNDARY_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	msg := "From: " + fromName + " <" + smtpEmail + ">\r\n" +
		"To: " + toEmail + "\r\n" +
		"Subject: " + asunto + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/alternative; boundary=\"" + boundary + "\"\r\n\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n\r\n" +
		cuerpoPlano + "\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n\r\n" +
		cuerpoHTML + "\r\n" +
		"--" + boundary + "--\r\n"
	if err := smtp.SendMail(addr, auth, smtpEmail, []string{toEmail}, []byte(msg)); err != nil {
		return acceptURL, err
	}
	return acceptURL, nil
}

func EmpresaCompartidaHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseInt64QueryOptional(r, "empresa_id")
			if err != nil || empresaID <= 0 {
				http.Error(w, "empresa_id invalido", http.StatusBadRequest)
				return
			}
			_, _, owner, err := ensureEmpresaOwnerAccess(dbEmp, dbSuper, r, empresaID)
			if err != nil {
				http.Error(w, "no se pudo validar empresa: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if !owner {
				http.Error(w, "solo el administrador propietario puede gestionar accesos compartidos", http.StatusForbidden)
				return
			}
			accesos, err := dbpkg.ListAdminEmpresaCompartidaAccesosByEmpresa(dbSuper, empresaID)
			if err != nil {
				http.Error(w, "no se pudieron cargar accesos compartidos: "+err.Error(), http.StatusInternalServerError)
				return
			}
			invitaciones, err := dbpkg.ListAdminEmpresaCompartidaInvitacionesByEmpresa(dbSuper, empresaID)
			if err != nil {
				http.Error(w, "no se pudieron cargar invitaciones compartidas: "+err.Error(), http.StatusInternalServerError)
				return
			}
			normalized := make([]dbpkg.AdminEmpresaCompartidaInvitacion, 0, len(invitaciones))
			for _, item := range invitaciones {
				normalized = append(normalized, normalizeAdminEmpresaCompartidaInvitation(item))
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "accesos": accesos, "invitaciones": normalized})
			return

		case http.MethodPost:
			var payload struct {
				EmpresaID int64  `json:"empresa_id"`
				Email     string `json:"email"`
				Mensaje   string `json:"mensaje"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido", http.StatusBadRequest)
				return
			}
			payload.Email = strings.TrimSpace(payload.Email)
			if payload.EmpresaID <= 0 || payload.Email == "" {
				http.Error(w, "empresa_id y email son obligatorios", http.StatusBadRequest)
				return
			}
			if _, err := mail.ParseAddress(payload.Email); err != nil {
				http.Error(w, "email invalido", http.StatusBadRequest)
				return
			}
			empresa, principalEmail, owner, err := ensureEmpresaOwnerAccess(dbEmp, dbSuper, r, payload.EmpresaID)
			if err != nil {
				http.Error(w, "no se pudo validar empresa: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if !owner {
				http.Error(w, "solo el administrador propietario puede invitar acceso compartido", http.StatusForbidden)
				return
			}
			if strings.EqualFold(payload.Email, principalEmail) {
				http.Error(w, "no puedes compartir la empresa contigo mismo", http.StatusConflict)
				return
			}
			adminTarget, err := dbpkg.GetAdminByEmailFull(dbSuper, payload.Email)
			if err != nil || adminTarget == nil {
				http.Error(w, "el administrador invitado debe existir y estar registrado", http.StatusBadRequest)
				return
			}
			if strings.EqualFold(strings.TrimSpace(adminTarget.Estado), "inactivo") {
				http.Error(w, "el administrador invitado está inactivo", http.StatusConflict)
				return
			}
			if canAccess, accessErr := dbpkg.CanAdminAccessEmpresaIA(dbEmp, dbSuper, payload.Email, payload.EmpresaID); accessErr != nil {
				http.Error(w, "no se pudo validar acceso actual del administrador invitado", http.StatusInternalServerError)
				return
			} else if canAccess {
				http.Error(w, "ese administrador ya tiene acceso a la empresa", http.StatusConflict)
				return
			}
			pending, err := dbpkg.GetPendingAdminEmpresaCompartidaInvitacion(dbSuper, payload.EmpresaID, payload.Email)
			if err != nil {
				http.Error(w, "no se pudo validar invitación pendiente", http.StatusInternalServerError)
				return
			}
			if pending != nil && !isAdminEmpresaCompartidaInvitationExpired(pending) {
				http.Error(w, "ya existe una invitación pendiente para ese administrador", http.StatusConflict)
				return
			}
			token, tokenHash, expiraEn, err := newAdminEmpresaCompartidaInvitationTokenAndExpiration()
			if err != nil {
				http.Error(w, "no se pudo generar token de invitación", http.StatusInternalServerError)
				return
			}
			inviter, _ := dbpkg.GetAdminByEmailFull(dbSuper, principalEmail)
			invID, err := dbpkg.CreateAdminEmpresaCompartidaInvitacion(dbSuper, dbpkg.AdminEmpresaCompartidaInvitacion{
				EmpresaID:        payload.EmpresaID,
				AdminEmail:       payload.Email,
				InvitadoPorEmail: principalEmail,
				TokenHash:        tokenHash,
				Mensaje:          payload.Mensaje,
				ExpiraEn:         expiraEn,
				UsuarioCreador:   principalEmail,
				Estado:           "pendiente",
			})
			if err != nil {
				http.Error(w, "no se pudo crear invitación compartida: "+err.Error(), http.StatusInternalServerError)
				return
			}
			acceptURL, mailErr := sendAdminEmpresaCompartidaInvitationEmail(r, dbEmp, dbSuper, empresa, inviter, payload.Email, adminTarget.Name, token, payload.Mensaje)
			response := map[string]interface{}{"ok": true, "id": invID, "accept_url": acceptURL}
			if mailErr != nil {
				response["email_sent"] = false
				response["message"] = "La invitación se creó, pero el correo no pudo enviarse. Puedes reenviarla desde esta misma pantalla."
				response["error"] = mailErr.Error()
				writeJSON(w, http.StatusOK, response)
				return
			}
			response["email_sent"] = true
			response["message"] = "Invitación enviada correctamente al administrador seleccionado."
			writeJSON(w, http.StatusOK, response)
			return

		case http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action != "reenviar" {
				http.Error(w, "accion no soportada", http.StatusBadRequest)
				return
			}
			id, err := parseInt64QueryOptional(r, "id")
			if err != nil || id <= 0 {
				http.Error(w, "id invalido", http.StatusBadRequest)
				return
			}
			inv, err := dbpkg.GetAdminEmpresaCompartidaInvitacionByID(dbSuper, id)
			if err != nil || inv == nil {
				http.Error(w, "invitación no encontrada", http.StatusNotFound)
				return
			}
			empresa, principalEmail, owner, err := ensureEmpresaOwnerAccess(dbEmp, dbSuper, r, inv.EmpresaID)
			if err != nil {
				http.Error(w, "no se pudo validar empresa: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if !owner {
				http.Error(w, "solo el administrador propietario puede reenviar invitaciones", http.StatusForbidden)
				return
			}
			token, tokenHash, expiraEn, err := newAdminEmpresaCompartidaInvitationTokenAndExpiration()
			if err != nil {
				http.Error(w, "no se pudo regenerar token de invitación", http.StatusInternalServerError)
				return
			}
			if err := dbpkg.RefreshAdminEmpresaCompartidaInvitacion(dbSuper, inv.ID, tokenHash, inv.Mensaje, expiraEn, principalEmail); err != nil {
				http.Error(w, "no se pudo actualizar la invitación", http.StatusInternalServerError)
				return
			}
			inviter, _ := dbpkg.GetAdminByEmailFull(dbSuper, principalEmail)
			acceptURL, mailErr := sendAdminEmpresaCompartidaInvitationEmail(r, dbEmp, dbSuper, empresa, inviter, inv.AdminEmail, inv.AdminName, token, inv.Mensaje)
			response := map[string]interface{}{"ok": true, "id": inv.ID, "accept_url": acceptURL}
			if mailErr != nil {
				response["email_sent"] = false
				response["message"] = "La invitación se actualizó, pero el correo no pudo reenviarse."
				response["error"] = mailErr.Error()
				writeJSON(w, http.StatusOK, response)
				return
			}
			response["email_sent"] = true
			response["message"] = "Invitación reenviada correctamente."
			writeJSON(w, http.StatusOK, response)
			return

		case http.MethodDelete:
			kind := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("kind")))
			id, err := parseInt64QueryOptional(r, "id")
			if err != nil || id <= 0 {
				http.Error(w, "id invalido", http.StatusBadRequest)
				return
			}
			if kind == "access" {
				access, err := dbpkg.GetAdminEmpresaCompartidaAccesoByID(dbSuper, id)
				if err != nil || access == nil {
					http.Error(w, "acceso compartido no encontrado", http.StatusNotFound)
					return
				}
				_, principalEmail, owner, ownErr := ensureEmpresaOwnerAccess(dbEmp, dbSuper, r, access.EmpresaID)
				if ownErr != nil {
					http.Error(w, "no se pudo validar empresa: "+ownErr.Error(), http.StatusInternalServerError)
					return
				}
				if !owner {
					http.Error(w, "solo el administrador propietario puede revocar accesos compartidos", http.StatusForbidden)
					return
				}
				if err := dbpkg.RevokeAdminEmpresaCompartidaAcceso(dbSuper, access.ID, principalEmail); err != nil {
					http.Error(w, "no se pudo revocar acceso compartido", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			}
			inv, err := dbpkg.GetAdminEmpresaCompartidaInvitacionByID(dbSuper, id)
			if err != nil || inv == nil {
				http.Error(w, "invitación compartida no encontrada", http.StatusNotFound)
				return
			}
			_, principalEmail, owner, ownErr := ensureEmpresaOwnerAccess(dbEmp, dbSuper, r, inv.EmpresaID)
			if ownErr != nil {
				http.Error(w, "no se pudo validar empresa: "+ownErr.Error(), http.StatusInternalServerError)
				return
			}
			if !owner {
				http.Error(w, "solo el administrador propietario puede revocar invitaciones", http.StatusForbidden)
				return
			}
			if err := dbpkg.SetAdminEmpresaCompartidaInvitacionEstado(dbSuper, inv.ID, "revocada", principalEmail); err != nil {
				http.Error(w, "no se pudo revocar invitación compartida", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func EmpresaCompartidaAcceptHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var payload struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "payload invalido", http.StatusBadRequest)
			return
		}
		payload.Token = strings.TrimSpace(payload.Token)
		if payload.Token == "" {
			http.Error(w, "token requerido", http.StatusBadRequest)
			return
		}
		requesterEmail := strings.ToLower(strings.TrimSpace(adminEmailFromRequest(r)))
		if requesterEmail == "" || requesterEmail == "sistema" {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		inv, err := dbpkg.GetAdminEmpresaCompartidaInvitacionByTokenHash(dbSuper, hashAdminEmpresaCompartidaToken(payload.Token))
		if err != nil || inv == nil {
			http.Error(w, "invitación no válida", http.StatusNotFound)
			return
		}
		if !strings.EqualFold(strings.TrimSpace(inv.AdminEmail), requesterEmail) {
			http.Error(w, "la invitación no corresponde al administrador autenticado", http.StatusForbidden)
			return
		}
		if isAdminEmpresaCompartidaInvitationExpired(inv) {
			_ = dbpkg.SetAdminEmpresaCompartidaInvitacionEstado(dbSuper, inv.ID, "expirada", requesterEmail)
			http.Error(w, "la invitación expiró", http.StatusConflict)
			return
		}
		estado := strings.ToLower(strings.TrimSpace(inv.Estado))
		if estado == "revocada" || estado == "rechazada" {
			http.Error(w, "la invitación ya no está disponible", http.StatusConflict)
			return
		}
		if estado == "aceptada" {
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "message": "La invitación ya había sido aceptada.", "empresa_id": inv.EmpresaID})
			return
		}
		if _, err := dbpkg.GetAdminByEmailFull(dbSuper, requesterEmail); err != nil {
			http.Error(w, "el administrador autenticado no existe o no está disponible", http.StatusBadRequest)
			return
		}
		acceptedAt := time.Now().Format("2006-01-02 15:04:05")
		if _, err := dbpkg.UpsertAdminEmpresaCompartidaAcceso(dbSuper, dbpkg.AdminEmpresaCompartidaAcceso{
			EmpresaID:          inv.EmpresaID,
			AdminEmail:         requesterEmail,
			CompartidoPorEmail: inv.InvitadoPorEmail,
			InvitacionID:       inv.ID,
			FechaAceptada:      acceptedAt,
			UsuarioCreador:     requesterEmail,
			Estado:             "activo",
		}); err != nil {
			http.Error(w, "no se pudo activar el acceso compartido: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := dbpkg.MarkAdminEmpresaCompartidaInvitacionAccepted(dbSuper, inv.ID, acceptedAt, requesterEmail); err != nil {
			http.Error(w, "no se pudo cerrar la invitación compartida", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "message": "Acceso compartido aceptado correctamente.", "empresa_id": inv.EmpresaID})
	}
}