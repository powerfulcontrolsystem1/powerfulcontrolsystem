package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
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

const asesorComercialMailNotificationType = "asesor_comercial_invitation"

func hashAsesorComercialToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

func currentAdminFromSession(r *http.Request, dbSuper *sql.DB) (*dbpkg.Admin, error) {
	if dbSuper == nil {
		return nil, fmt.Errorf("base super no disponible")
	}
	c, err := r.Cookie("session_token")
	if err != nil || c == nil || strings.TrimSpace(c.Value) == "" {
		return nil, fmt.Errorf("unauthenticated")
	}
	s, err := dbpkg.GetSessionByToken(dbSuper, strings.TrimSpace(c.Value))
	if err != nil || s == nil {
		return nil, fmt.Errorf("unauthenticated")
	}
	admin, err := dbpkg.GetAdminByEmailFull(dbSuper, s.AdminEmail)
	if err != nil || admin == nil {
		return nil, fmt.Errorf("account not found")
	}
	return admin, nil
}

func requireSuperAdmin(r *http.Request, dbSuper *sql.DB) (*dbpkg.Admin, bool, int, string) {
	admin, err := currentAdminFromSession(r, dbSuper)
	if err != nil {
		return nil, false, http.StatusUnauthorized, "unauthenticated"
	}
	if !strings.EqualFold(strings.TrimSpace(admin.Role), "super_administrador") {
		return nil, false, http.StatusForbidden, "solo super administrador"
	}
	return admin, true, http.StatusOK, ""
}

func newAsesorComercialTokenAndExpiration() (string, string, string, error) {
	token, err := utils.GenerateSecureToken(32)
	if err != nil {
		return "", "", "", err
	}
	return token, hashAsesorComercialToken(token), time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339), nil
}

func newAsesorComercialCode(dbSuper *sql.DB) (string, error) {
	for i := 0; i < 12; i++ {
		raw, err := utils.GenerateSecureToken(8)
		if err != nil {
			return "", err
		}
		code := "AC-" + strings.ToUpper(strings.ReplaceAll(raw[:8], "_", "X"))
		code = strings.ReplaceAll(code, "-", "")
		if len(code) > 10 {
			code = code[:10]
		}
		code = "AC-" + code[2:]
		existing, err := dbpkg.GetAsesorComercialByCode(dbSuper, code)
		if err != nil {
			return "", err
		}
		if existing == nil {
			return code, nil
		}
	}
	return "", fmt.Errorf("no se pudo generar codigo unico")
}

func asesorComercialAcceptURL(r *http.Request, dbSuper *sql.DB, token string) string {
	baseURL := resolveBaseURLForConfirmation(r, dbSuper)
	return strings.TrimRight(baseURL, "/") + "/api/asesor_comercial/aceptar?token=" + url.QueryEscape(strings.TrimSpace(token))
}

func sendAsesorComercialInvitationEmail(r *http.Request, dbSuper *sql.DB, item dbpkg.AsesorComercial, token string) (string, error) {
	acceptURL := asesorComercialAcceptURL(r, dbSuper, token)
	loginURL := strings.TrimRight(resolveBaseURLForConfirmation(r, dbSuper), "/") + "/login.html"
	name := strings.TrimSpace(item.AdminNombre)
	if name == "" {
		name = "administrador"
	}
	asunto, cuerpoPlano, cuerpoHTML := asesorComercialEmailContent(name, item.Codigo, item.PorcentajeComision, item.MesesAsociacion, acceptURL, loginURL)
	if isEmpresaUsuarioMailTestMode(dbSuper) {
		metadataJSON := fmt.Sprintf(`{"accept_url":%q,"login_url":%q,"codigo":%q,"mail_mode":"test"}`, acceptURL, loginURL, item.Codigo)
		if err := captureEmpresaUsuarioMailNotification(dbSuper, asesorComercialMailNotificationType, 0, item.AdminEmail, asunto, cuerpoPlano, token, metadataJSON, adminEmailFromRequest(r)); err != nil {
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
	if _, err := mail.ParseAddress(item.AdminEmail); err != nil {
		return acceptURL, fmt.Errorf("correo destino invalido: %w", err)
	}
	hostForAuth := smtpHost
	if strings.Contains(smtpHost, ":") {
		if h, _, splitErr := net.SplitHostPort(smtpHost); splitErr == nil {
			hostForAuth = h
		}
	}
	addr := smtpHost
	if !strings.Contains(addr, ":") {
		addr = smtpHost + ":" + smtpPort
	}
	from := (&mail.Address{Name: fromName, Address: smtpEmail}).String()
	to := (&mail.Address{Name: name, Address: item.AdminEmail}).String()
	boundary := "pcs-asesor-comercial"
	msg := "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + asunto + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/alternative; boundary=" + boundary + "\r\n\r\n" +
		"--" + boundary + "\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" + cuerpoPlano + "\r\n" +
		"--" + boundary + "\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n" + cuerpoHTML + "\r\n" +
		"--" + boundary + "--\r\n"
	auth := smtp.PlainAuth("", smtpEmail, smtpPass, hostForAuth)
	return acceptURL, smtp.SendMail(addr, auth, smtpEmail, []string{item.AdminEmail}, []byte(msg))
}

func asesorComercialEmailContent(name, code string, pct float64, meses int, acceptURL, loginURL string) (string, string, string) {
	if meses <= 0 {
		meses = 6
	}
	subject := "Invitación para activar asesor comercial"
	text := fmt.Sprintf("Hola %s,\n\nPowerful Control System te invitó a ser asesor comercial.\n\nTu código de asesor será: %s\nComisión configurada: %.2f%%\nTiempo de asociación por cliente: %d mes(es)\n\nPara aceptar la invitación, abre este enlace:\n%s\n\nDespués podrás iniciar sesión y ver Mis clientes desde Seleccionar empresa:\n%s\n\nSi no esperabas esta invitación, ignora este mensaje.\n", name, code, pct, meses, acceptURL, loginURL)
	html := fmt.Sprintf("<html><body><p>Hola %s,</p><p>Powerful Control System te invitó a ser <strong>asesor comercial</strong>.</p><p><strong>Código de asesor:</strong> %s<br><strong>Comisión configurada:</strong> %.2f%%<br><strong>Tiempo de asociación por cliente:</strong> %d mes(es)</p><p><a href=\"%s\" style=\"display:inline-block;padding:12px 18px;background:#0f6fcb;color:#fff;text-decoration:none;border-radius:8px;font-weight:700;\">Aceptar invitación</a></p><p>Después podrás iniciar sesión y consultar <strong>Mis clientes</strong> desde Seleccionar empresa.</p><p>Acceso manual: <a href=\"%s\">%s</a></p><p>Si no esperabas esta invitación, ignora este mensaje.</p></body></html>", htmlEscape(name), htmlEscape(code), pct, meses, htmlEscape(acceptURL), htmlEscape(loginURL), htmlEscape(loginURL))
	return subject, text, html
}

func htmlEscape(value string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&#39;")
	return replacer.Replace(value)
}

// AsesorComercialSuperHandler administra asesores, reglas de comision y liquidaciones desde super.
func AsesorComercialSuperHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		admin, ok, status, msg := requireSuperAdmin(r, dbSuper)
		if !ok {
			http.Error(w, msg, status)
			return
		}
		switch r.Method {
		case http.MethodGet:
			if strings.EqualFold(r.URL.Query().Get("action"), "comisiones") {
				items, err := dbpkg.ListAsesorComercialComisiones(dbSuper, "", true)
				if err != nil {
					http.Error(w, "no se pudieron cargar comisiones: "+err.Error(), http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "items": items})
				return
			}
			items, err := dbpkg.ListAsesoresComerciales(dbSuper)
			if err != nil {
				http.Error(w, "no se pudieron cargar asesores: "+err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "items": items})
		case http.MethodPost:
			var payload struct {
				Email              string  `json:"email"`
				PorcentajeComision float64 `json:"porcentaje_comision"`
				MesesAsociacion    int     `json:"meses_asociacion"`
				Observaciones      string  `json:"observaciones"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido", http.StatusBadRequest)
				return
			}
			email := strings.ToLower(strings.TrimSpace(payload.Email))
			if email == "" {
				http.Error(w, "email requerido", http.StatusBadRequest)
				return
			}
			target, err := dbpkg.GetAdminByEmailFull(dbSuper, email)
			if err != nil || target == nil {
				http.Error(w, "el correo debe corresponder a un administrador registrado", http.StatusBadRequest)
				return
			}
			if strings.EqualFold(target.Estado, "inactivo") {
				http.Error(w, "el administrador esta inactivo", http.StatusConflict)
				return
			}
			if payload.PorcentajeComision < 0 || payload.PorcentajeComision > 100 {
				http.Error(w, "porcentaje_comision debe estar entre 0 y 100", http.StatusBadRequest)
				return
			}
			if payload.MesesAsociacion <= 0 {
				payload.MesesAsociacion = 6
			}
			code := ""
			if existing, lookupErr := dbpkg.GetAsesorComercialByEmail(dbSuper, email); lookupErr == nil && existing != nil {
				code = strings.TrimSpace(existing.Codigo)
			}
			if code == "" {
				var codeErr error
				code, codeErr = newAsesorComercialCode(dbSuper)
				if codeErr != nil {
					http.Error(w, "no se pudo generar codigo: "+codeErr.Error(), http.StatusInternalServerError)
					return
				}
			}
			token, tokenHash, expira, err := newAsesorComercialTokenAndExpiration()
			if err != nil {
				http.Error(w, "no se pudo generar invitacion", http.StatusInternalServerError)
				return
			}
			item := dbpkg.AsesorComercial{
				AdminEmail:         email,
				AdminNombre:        target.Name,
				Codigo:             code,
				PorcentajeComision: roundMoney(payload.PorcentajeComision),
				MesesAsociacion:    payload.MesesAsociacion,
				InvitacionExpiraEn: expira,
				InvitadoPorEmail:   admin.Email,
				Observaciones:      payload.Observaciones,
			}
			id, err := dbpkg.CreateAsesorComercial(dbSuper, item, tokenHash)
			if err != nil {
				http.Error(w, "no se pudo guardar asesor comercial: "+err.Error(), http.StatusInternalServerError)
				return
			}
			item.ID = id
			acceptURL, mailErr := sendAsesorComercialInvitationEmail(r, dbSuper, item, token)
			resp := map[string]interface{}{"ok": true, "id": id, "codigo": code, "accept_url": acceptURL}
			if mailErr != nil {
				resp["email_sent"] = false
				resp["message"] = "El asesor fue creado, pero el correo no pudo enviarse. Revisa Gmail SMTP."
				resp["error"] = mailErr.Error()
			} else {
				resp["email_sent"] = true
			}
			writeJSON(w, http.StatusOK, resp)
		case http.MethodPut:
			id, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("id")), 10, 64)
			if id <= 0 {
				http.Error(w, "id requerido", http.StatusBadRequest)
				return
			}
			if strings.EqualFold(r.URL.Query().Get("action"), "marcar_pago") {
				var payload struct {
					Observaciones string `json:"observaciones"`
				}
				_ = json.NewDecoder(r.Body).Decode(&payload)
				if err := dbpkg.MarkAsesorComercialComisionPagada(dbSuper, id, admin.Email, payload.Observaciones); err != nil {
					http.Error(w, "no se pudo marcar pago: "+err.Error(), http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			}
			var payload struct {
				PorcentajeComision float64 `json:"porcentaje_comision"`
				MesesAsociacion    int     `json:"meses_asociacion"`
				Observaciones      string  `json:"observaciones"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido", http.StatusBadRequest)
				return
			}
			if payload.PorcentajeComision < 0 || payload.PorcentajeComision > 100 {
				http.Error(w, "porcentaje_comision debe estar entre 0 y 100", http.StatusBadRequest)
				return
			}
			if payload.MesesAsociacion <= 0 {
				payload.MesesAsociacion = 6
			}
			if err := dbpkg.UpdateAsesorComercial(dbSuper, id, payload.PorcentajeComision, payload.MesesAsociacion, payload.Observaciones, admin.Email); err != nil {
				http.Error(w, "no se pudo actualizar asesor: "+err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
		case http.MethodDelete:
			id, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("id")), 10, 64)
			if id <= 0 {
				http.Error(w, "id requerido", http.StatusBadRequest)
				return
			}
			if err := dbpkg.InactivateAsesorComercial(dbSuper, id, admin.Email); err != nil {
				http.Error(w, "no se pudo desactivar asesor: "+err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func AsesorComercialAcceptHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		if token == "" {
			http.Error(w, "token requerido", http.StatusBadRequest)
			return
		}
		item, err := dbpkg.GetAsesorComercialByTokenHash(dbSuper, hashAsesorComercialToken(token))
		if err != nil || item == nil {
			http.Error(w, "invitacion no valida", http.StatusNotFound)
			return
		}
		if !strings.EqualFold(item.EstadoInvitacion, "pendiente") {
			http.Error(w, "la invitacion ya no esta pendiente", http.StatusConflict)
			return
		}
		if exp, ok := parseAsesorTime(item.InvitacionExpiraEn); ok && time.Now().After(exp) {
			http.Error(w, "la invitacion expiro", http.StatusConflict)
			return
		}
		admin, err := dbpkg.GetAdminByEmailFull(dbSuper, item.AdminEmail)
		if err != nil || admin == nil || strings.EqualFold(admin.Estado, "inactivo") {
			http.Error(w, "administrador no disponible", http.StatusBadRequest)
			return
		}
		acceptedAt := time.Now().Format("2006-01-02 15:04:05")
		if err := dbpkg.AcceptAsesorComercialInvitation(dbSuper, item.ID, acceptedAt, item.AdminEmail); err != nil {
			http.Error(w, "no se pudo aceptar invitacion", http.StatusInternalServerError)
			return
		}
		if err := createAdminEmpresaCompartidaSession(w, r, dbSuper, item.AdminEmail); err != nil {
			http.Error(w, "invitacion aceptada, pero no se pudo iniciar sesion", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/seleccionar_empresa.html?asesor_comercial=aceptado", http.StatusFound)
	}
}

func AsesorComercialMisClientesHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		admin, err := currentAdminFromSession(r, dbSuper)
		if err != nil {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		advisor, err := dbpkg.GetAsesorComercialByEmail(dbSuper, admin.Email)
		if err != nil {
			http.Error(w, "no se pudo validar asesor: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if advisor == nil || !strings.EqualFold(advisor.EstadoInvitacion, "aceptada") {
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "is_asesor": false, "items": []interface{}{}})
			return
		}
		items, err := dbpkg.ListAsesorComercialComisiones(dbSuper, admin.Email, false)
		if err != nil {
			http.Error(w, "no se pudieron cargar clientes: "+err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "is_asesor": true, "asesor": advisor, "items": items})
	}
}

func parseAsesorTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, value); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func roundMoney(value float64) float64 {
	return math.Round(value*100) / 100
}
