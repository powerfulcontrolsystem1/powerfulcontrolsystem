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

func isAllowedAdminEmpresaCompartidaRole(raw string) bool {
	role := strings.ToLower(strings.TrimSpace(raw))
	switch role {
	case "super_administrador", "superadmin", "super":
		return true
	case "administrador", "admin", "admin_empresa":
		return true
	default:
		return false
	}
}

func newAdminEmpresaCompartidaInvitationTokenAndExpiration() (string, string, string, error) {
	token, err := utils.GenerateSecureToken(32)
	if err != nil {
		return "", "", "", err
	}
	expira := time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339)
	return token, hashAdminEmpresaCompartidaToken(token), expira, nil
}

func adminEmpresaCompartidaAcceptURL(r *http.Request, dbSuper *sql.DB, token string) string {
	baseURL := resolveBaseURLForConfirmation(r, dbSuper)
	return strings.TrimRight(baseURL, "/") + "/super/api/empresas/compartidos/aceptar?token=" + url.QueryEscape(strings.TrimSpace(token))
}

func createAdminEmpresaCompartidaSession(w http.ResponseWriter, r *http.Request, dbSuper *sql.DB, adminEmail string) error {
	token, err := utils.GenerateSecureToken(32)
	if err != nil {
		return err
	}
	if err := dbpkg.CreateSession(dbSuper, adminEmail, r.RemoteAddr, r.UserAgent(), token); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400,
		Secure:   SessionCookieSecure(r),
		SameSite: http.SameSiteLaxMode,
	})
	SetBrowserSessionStateCookie(w, r, true)
	return nil
}

func acceptAdminEmpresaCompartidaInvitationByToken(w http.ResponseWriter, r *http.Request, dbSuper *sql.DB, token string) (map[string]interface{}, int, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, http.StatusBadRequest, fmt.Errorf("token requerido")
	}
	inv, err := dbpkg.GetAdminEmpresaCompartidaInvitacionByTokenHash(dbSuper, hashAdminEmpresaCompartidaToken(token))
	if err != nil || inv == nil {
		return nil, http.StatusNotFound, fmt.Errorf("invitación no válida")
	}
	if isAdminEmpresaCompartidaInvitationExpired(inv) {
		_ = dbpkg.SetAdminEmpresaCompartidaInvitacionEstado(dbSuper, inv.ID, "expirada", strings.TrimSpace(inv.AdminEmail))
		return nil, http.StatusConflict, fmt.Errorf("la invitación expiró")
	}
	estado := strings.ToLower(strings.TrimSpace(inv.Estado))
	if estado == "aceptada" {
		return nil, http.StatusConflict, fmt.Errorf("la invitación ya fue aceptada")
	}
	if estado == "revocada" || estado == "rechazada" {
		return nil, http.StatusConflict, fmt.Errorf("la invitación ya no está disponible")
	}
	adminTarget, err := dbpkg.GetAdminByEmailFull(dbSuper, inv.AdminEmail)
	if err != nil || adminTarget == nil {
		return nil, http.StatusBadRequest, fmt.Errorf("el administrador invitado no existe o no está disponible")
	}
	if !isAllowedAdminEmpresaCompartidaRole(adminTarget.Role) {
		return nil, http.StatusForbidden, fmt.Errorf("solo un usuario administrador o superadministrador puede aceptar una empresa compartida")
	}
	if strings.EqualFold(strings.TrimSpace(adminTarget.Estado), "inactivo") {
		return nil, http.StatusConflict, fmt.Errorf("el administrador invitado está inactivo")
	}

	acceptedAt := time.Now().Format("2006-01-02 15:04:05")
	if _, err := dbpkg.UpsertAdminEmpresaCompartidaAcceso(dbSuper, dbpkg.AdminEmpresaCompartidaAcceso{
		EmpresaID:          inv.EmpresaID,
		AdminEmail:         strings.TrimSpace(inv.AdminEmail),
		CompartidoPorEmail: inv.InvitadoPorEmail,
		InvitacionID:       inv.ID,
		FechaAceptada:      acceptedAt,
		UsuarioCreador:     strings.TrimSpace(inv.AdminEmail),
		Estado:             "activo",
	}); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("no se pudo activar el acceso compartido: %w", err)
	}
	if err := dbpkg.MarkAdminEmpresaCompartidaInvitacionAccepted(dbSuper, inv.ID, acceptedAt, inv.AdminEmail); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("no se pudo cerrar la invitación compartida")
	}
	if err := createAdminEmpresaCompartidaSession(w, r, dbSuper, adminTarget.Email); err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("el acceso fue aceptado, pero no se pudo iniciar la sesión automática")
	}

	return map[string]interface{}{
		"ok":           true,
		"message":      "Acceso compartido aceptado correctamente. La empresa ya está disponible en seleccionar empresa.",
		"empresa_id":   inv.EmpresaID,
		"admin_email":  strings.TrimSpace(adminTarget.Email),
		"redirect_url": "/seleccionar_empresa.html?empresa_id=" + url.QueryEscape(strconv.FormatInt(inv.EmpresaID, 10)) + "&shared_invitation_accepted=1",
	}, http.StatusOK, nil
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
	_ = principalEmail
	owner := adminOwnsEmpresaByCreatorEmail(requesterEmail, empresa.UsuarioCreador)
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
		_ = principalEmail
		owner := adminOwnsEmpresaByCreatorEmail(requesterEmail, empresa.UsuarioCreador)
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
	requesterEmail := strings.ToLower(strings.TrimSpace(adminEmailFromRequest(r)))
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
	owner := adminOwnsEmpresaByCreatorEmail(requesterEmail, empresa.UsuarioCreador)
	return empresa, principalEmail, owner, nil
}

func sendAdminEmpresaCompartidaInvitationEmail(r *http.Request, dbEmp, dbSuper *sql.DB, empresa *dbpkg.Empresa, inviter *dbpkg.Admin, toEmail, toName, token, mensaje string) (string, error) {
	baseURL := resolveBaseURLForConfirmation(r, dbSuper)
	loginURL := strings.TrimRight(baseURL, "/") + "/login.html"
	acceptURL := adminEmpresaCompartidaAcceptURL(r, dbSuper, token)
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

func targetAdminAlreadyHasEmpresaAccess(dbSuper *sql.DB, empresa *dbpkg.Empresa, targetAdminEmail string) (bool, error) {
	if empresa == nil || dbSuper == nil {
		return false, nil
	}
	targetAdminEmail = strings.ToLower(strings.TrimSpace(targetAdminEmail))
	if targetAdminEmail == "" || empresa.EmpresaID <= 0 {
		return false, nil
	}

	// 1) Dueño exacto: solo el email creador real de la empresa.
	creator := strings.ToLower(strings.TrimSpace(empresa.UsuarioCreador))
	if creator != "" && creator == targetAdminEmail {
		return true, nil
	}

	// 2) Acceso compartido activo (solo existe tras aceptación de invitación).
	access, err := dbpkg.GetActiveAdminEmpresaCompartidaAcceso(dbSuper, empresa.EmpresaID, targetAdminEmail)
	if err != nil {
		return false, err
	}
	return access != nil, nil
}

func registrarAuditoriaEmpresaCompartidaNoBloqueante(dbEmp *sql.DB, r *http.Request, empresaID int64, accion string, recurso string, recursoID int64, statusCode int, metadata map[string]interface{}, observaciones string) {
	if dbEmp == nil || r == nil || empresaID <= 0 {
		return
	}
	raw, _ := json.Marshal(metadata)
	resultado := "ok"
	if statusCode >= 400 {
		resultado = "error"
	}
	usuario := strings.TrimSpace(adminEmailFromRequest(r))
	if usuario == "" {
		usuario = "sistema"
	}
	_, _ = dbpkg.CreateEmpresaAuditoriaEvento(dbEmp, dbpkg.EmpresaAuditoriaEvento{
		EmpresaID:      empresaID,
		Modulo:         "empresas_compartidas",
		Accion:         strings.TrimSpace(accion),
		Recurso:        strings.TrimSpace(recurso),
		RecursoID:      recursoID,
		MetodoHTTP:     r.Method,
		Endpoint:       r.URL.Path,
		Resultado:      resultado,
		CodigoHTTP:     int64(statusCode),
		RequestID:      utils.RequestIDFromContext(r.Context()),
		IPOrigen:       strings.TrimSpace(r.RemoteAddr),
		UserAgent:      r.UserAgent(),
		MetadataJSON:   string(raw),
		RetencionDias:  3650,
		UsuarioCreador: usuario,
		Estado:         "activo",
		Observaciones:  strings.TrimSpace(observaciones),
	})
}

func EmpresaCompartidaHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("action")), "pendientes_mias") {
				requesterEmail := strings.ToLower(strings.TrimSpace(adminEmailFromRequest(r)))
				if requesterEmail == "" || requesterEmail == "sistema" {
					http.Error(w, "unauthenticated", http.StatusUnauthorized)
					return
				}
				invitaciones, err := dbpkg.ListPendingAdminEmpresaCompartidaInvitacionesByAdmin(dbSuper, requesterEmail)
				if err != nil {
					http.Error(w, "no se pudieron cargar invitaciones pendientes: "+err.Error(), http.StatusInternalServerError)
					return
				}
				items := make([]map[string]interface{}, 0, len(invitaciones))
				for _, inv := range invitaciones {
					inv = normalizeAdminEmpresaCompartidaInvitation(inv)
					empresaNombre := ""
					empresaEstado := ""
					if dbEmp != nil && inv.EmpresaID > 0 {
						if emp, eerr := dbpkg.GetEmpresaByID(dbEmp, inv.EmpresaID); eerr == nil && emp != nil {
							empresaNombre = strings.TrimSpace(emp.Nombre)
							empresaEstado = strings.TrimSpace(emp.Estado)
						}
					}
					items = append(items, map[string]interface{}{
						"id":             inv.ID,
						"empresa_id":     inv.EmpresaID,
						"empresa_nombre": empresaNombre,
						"empresa_estado": empresaEstado,
						"admin_email":    strings.TrimSpace(inv.AdminEmail),
						"invitado_por":   strings.TrimSpace(inv.InvitadoPorEmail),
						"mensaje":        strings.TrimSpace(inv.Mensaje),
						"expira_en":      strings.TrimSpace(inv.ExpiraEn),
						"fecha_creacion": strings.TrimSpace(inv.FechaCreacion),
						"estado":         strings.TrimSpace(inv.Estado),
					})
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "items": items})
				return
			}
			empresaID, err := parseInt64QueryOptional(r, "empresa_id")
			if err != nil || empresaID <= 0 {
				http.Error(w, "empresa_id invalido", http.StatusBadRequest)
				return
			}
			requesterEmail := strings.ToLower(strings.TrimSpace(adminEmailFromRequest(r)))
			if requesterEmail == "" || requesterEmail == "sistema" {
				http.Error(w, "unauthenticated", http.StatusUnauthorized)
				return
			}
			principalEmail, ok, err := ensureEmpresaInRequesterScope(dbEmp, dbSuper, r, empresaID)
			if err != nil {
				http.Error(w, "no se pudo validar empresa: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if !ok {
				http.Error(w, "no tienes acceso a esta empresa", http.StatusForbidden)
				return
			}
			_, _, owner, ownerErr := ensureEmpresaOwnerAccess(dbEmp, dbSuper, r, empresaID)
			if ownerErr != nil {
				http.Error(w, "no se pudo validar propietario: "+ownerErr.Error(), http.StatusInternalServerError)
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
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":              true,
				"accesos":         accesos,
				"invitaciones":    normalized,
				"is_owner":        owner,
				"requester_email": requesterEmail,
				"principal_email": strings.TrimSpace(principalEmail),
			})
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
			if !isAllowedAdminEmpresaCompartidaRole(adminTarget.Role) {
				http.Error(w, "solo se puede compartir con un usuario administrador o superadministrador", http.StatusForbidden)
				return
			}
			if strings.EqualFold(strings.TrimSpace(adminTarget.Estado), "inactivo") {
				http.Error(w, "el administrador invitado está inactivo", http.StatusConflict)
				return
			}
			if alreadyHasAccess, accessErr := targetAdminAlreadyHasEmpresaAccess(dbSuper, empresa, payload.Email); accessErr != nil {
				http.Error(w, "no se pudo validar acceso actual del administrador invitado", http.StatusInternalServerError)
				return
			} else if alreadyHasAccess {
				http.Error(w, "ese administrador ya tiene acceso a la empresa", http.StatusConflict)
				return
			}
			pending, err := dbpkg.GetPendingAdminEmpresaCompartidaInvitacion(dbSuper, payload.EmpresaID, payload.Email)
			if err != nil {
				http.Error(w, "no se pudo validar invitación pendiente", http.StatusInternalServerError)
				return
			}
			if pending != nil && !isAdminEmpresaCompartidaInvitationExpired(pending) {
				writeJSON(w, http.StatusConflict, map[string]interface{}{
					"ok":            false,
					"code":          "invitation_pending",
					"error":         "Ya existe una invitación pendiente para ese administrador.",
					"invitation_id": pending.ID,
					"empresa_id":    pending.EmpresaID,
					"admin_email":   strings.TrimSpace(pending.AdminEmail),
					"estado":        strings.TrimSpace(pending.Estado),
					"expira_en":     strings.TrimSpace(pending.ExpiraEn),
				})
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
			if action == "aceptar" {
				id, err := parseInt64QueryOptional(r, "id")
				if err != nil || id <= 0 {
					http.Error(w, "id invalido", http.StatusBadRequest)
					return
				}
				requesterEmail := strings.ToLower(strings.TrimSpace(adminEmailFromRequest(r)))
				if requesterEmail == "" || requesterEmail == "sistema" {
					http.Error(w, "unauthenticated", http.StatusUnauthorized)
					return
				}
				inv, err := dbpkg.GetAdminEmpresaCompartidaInvitacionByID(dbSuper, id)
				if err != nil || inv == nil {
					http.Error(w, "invitación no encontrada", http.StatusNotFound)
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
				adminRequester, err := dbpkg.GetAdminByEmailFull(dbSuper, requesterEmail)
				if err != nil || adminRequester == nil {
					http.Error(w, "el administrador autenticado no existe o no está disponible", http.StatusBadRequest)
					return
				}
				if !isAllowedAdminEmpresaCompartidaRole(adminRequester.Role) {
					http.Error(w, "solo un usuario administrador o superadministrador puede aceptar una empresa compartida", http.StatusForbidden)
					return
				}
				if strings.EqualFold(strings.TrimSpace(adminRequester.Estado), "inactivo") {
					http.Error(w, "el administrador autenticado está inactivo", http.StatusConflict)
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
				return
			}

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
				requesterEmail := strings.ToLower(strings.TrimSpace(adminEmailFromRequest(r)))
				if requesterEmail == "" || requesterEmail == "sistema" {
					http.Error(w, "unauthenticated", http.StatusUnauthorized)
					return
				}
				_, principalEmail, owner, ownErr := ensureEmpresaOwnerAccess(dbEmp, dbSuper, r, access.EmpresaID)
				if ownErr != nil {
					http.Error(w, "no se pudo validar empresa: "+ownErr.Error(), http.StatusInternalServerError)
					return
				}
				canRevoke := owner ||
					strings.EqualFold(strings.TrimSpace(access.AdminEmail), requesterEmail) ||
					strings.EqualFold(strings.TrimSpace(access.CompartidoPorEmail), requesterEmail)
				if !canRevoke {
					http.Error(w, "solo el propietario, quien compartio o el administrador receptor pueden revocar este acceso", http.StatusForbidden)
					return
				}
				actorEmail := strings.TrimSpace(principalEmail)
				if actorEmail == "" {
					actorEmail = requesterEmail
				}
				if err := dbpkg.RevokeAdminEmpresaCompartidaAcceso(dbSuper, access.ID, actorEmail); err != nil {
					http.Error(w, "no se pudo revocar acceso compartido", http.StatusInternalServerError)
					return
				}
				registrarAuditoriaEmpresaCompartidaNoBloqueante(dbEmp, r, access.EmpresaID, "revocar_acceso", "admin_empresa_compartida", access.ID, http.StatusOK, map[string]interface{}{
					"admin_email":          strings.TrimSpace(access.AdminEmail),
					"compartido_por_email": strings.TrimSpace(access.CompartidoPorEmail),
					"actor_email":          requesterEmail,
					"actor_es_propietario": owner,
					"invitacion_id":        access.InvitacionID,
					"estado_anterior":      strings.TrimSpace(access.Estado),
					"fecha_aceptada":       strings.TrimSpace(access.FechaAceptada),
				}, "Acceso compartido revocado desde el lapiz de empresa.")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			}
			inv, err := dbpkg.GetAdminEmpresaCompartidaInvitacionByID(dbSuper, id)
			if err != nil || inv == nil {
				http.Error(w, "invitación compartida no encontrada", http.StatusNotFound)
				return
			}
			requesterEmail := strings.ToLower(strings.TrimSpace(adminEmailFromRequest(r)))
			if requesterEmail == "" || requesterEmail == "sistema" {
				http.Error(w, "unauthenticated", http.StatusUnauthorized)
				return
			}
			_, principalEmail, owner, ownErr := ensureEmpresaOwnerAccess(dbEmp, dbSuper, r, inv.EmpresaID)
			if ownErr != nil {
				http.Error(w, "no se pudo validar empresa: "+ownErr.Error(), http.StatusInternalServerError)
				return
			}
			canRevokeInvitation := owner ||
				strings.EqualFold(strings.TrimSpace(inv.AdminEmail), requesterEmail) ||
				strings.EqualFold(strings.TrimSpace(inv.InvitadoPorEmail), requesterEmail)
			if !canRevokeInvitation {
				http.Error(w, "solo el propietario, quien invito o el administrador invitado pueden revocar esta invitacion", http.StatusForbidden)
				return
			}
			actorEmail := strings.TrimSpace(principalEmail)
			if actorEmail == "" {
				actorEmail = requesterEmail
			}
			if err := dbpkg.SetAdminEmpresaCompartidaInvitacionEstado(dbSuper, inv.ID, "revocada", actorEmail); err != nil {
				http.Error(w, "no se pudo revocar invitación compartida", http.StatusInternalServerError)
				return
			}
			registrarAuditoriaEmpresaCompartidaNoBloqueante(dbEmp, r, inv.EmpresaID, "revocar_invitacion", "admin_empresa_compartida_invitaciones", inv.ID, http.StatusOK, map[string]interface{}{
				"admin_email":          strings.TrimSpace(inv.AdminEmail),
				"invitado_por_email":   strings.TrimSpace(inv.InvitadoPorEmail),
				"actor_email":          requesterEmail,
				"actor_es_propietario": owner,
				"estado_anterior":      strings.TrimSpace(inv.Estado),
				"expira_en":            strings.TrimSpace(inv.ExpiraEn),
			}, "Invitacion compartida revocada desde el lapiz de empresa.")
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func EmpresaCompartidaAcceptHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			token := strings.TrimSpace(r.URL.Query().Get("token"))
			result, status, err := acceptAdminEmpresaCompartidaInvitationByToken(w, r, dbSuper, token)
			if err != nil {
				http.Error(w, err.Error(), status)
				return
			}
			redirectURL, _ := result["redirect_url"].(string)
			if strings.TrimSpace(redirectURL) == "" {
				redirectURL = "/seleccionar_empresa.html"
			}
			http.Redirect(w, r, redirectURL, http.StatusFound)
			return

		case http.MethodPost:
			var payload struct {
				Token string `json:"token"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido", http.StatusBadRequest)
				return
			}
			result, status, err := acceptAdminEmpresaCompartidaInvitationByToken(w, r, dbSuper, payload.Token)
			if err != nil {
				http.Error(w, err.Error(), status)
				return
			}
			writeJSON(w, http.StatusOK, result)
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}
