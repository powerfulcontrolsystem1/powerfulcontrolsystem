package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/mail"
	"sort"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const superCorreosMasivosMaxDestinatarios = 5000

type superCorreoMasivoPayload struct {
	Categoria     string `json:"categoria"`
	Alcance       string `json:"alcance"`
	Asunto        string `json:"asunto"`
	CuerpoTexto   string `json:"cuerpo_texto"`
	CuerpoHTML    string `json:"cuerpo_html"`
	Confirmar     bool   `json:"confirmar"`
	SoloPreview   bool   `json:"solo_preview"`
	Observaciones string `json:"observaciones"`
}

type superCorreoMasivoRecipient struct {
	Email            string   `json:"email"`
	EmailMasked      string   `json:"email_masked"`
	Nombre           string   `json:"nombre,omitempty"`
	TipoDestinatario string   `json:"tipo_destinatario"`
	EmpresaID        int64    `json:"empresa_id,omitempty"`
	EmpresaNombre    string   `json:"empresa_nombre,omitempty"`
	Rol              string   `json:"rol,omitempty"`
	Fuentes          []string `json:"fuentes,omitempty"`
}

type superCorreoMasivoPreview struct {
	Total              int                          `json:"total"`
	Administradores    int                          `json:"administradores"`
	UsuariosEmpresa    int                          `json:"usuarios_empresa"`
	DuplicadosOmitidos int                          `json:"duplicados_omitidos"`
	InvalidosOmitidos  int                          `json:"invalidos_omitidos"`
	Muestra            []superCorreoMasivoRecipient `json:"muestra"`
}

// SuperCorreosMasivosHandler permite al super administrador enviar comunicados globales.
func SuperCorreosMasivosHandler(dbEmpresas, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminEmail, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper)
		if !ok {
			return
		}
		if dbSuper == nil || dbEmpresas == nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "bases de datos no disponibles"})
			return
		}

		switch r.Method {
		case http.MethodGet:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			alcance := normalizeSuperCorreoMasivoAlcance(r.URL.Query().Get("alcance"))
			preview, err := buildSuperCorreoMasivoPreview(dbEmpresas, dbSuper, alcance)
			if err != nil {
				writeSuperCorreosMasivosPublicError(w, r, http.StatusInternalServerError, "generar vista previa", err, nil)
				return
			}
			if action == "preview" || action == "previsualizar" {
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "preview": preview, "mail_test_mode": isEmpresaUsuarioMailTestMode(dbSuper)})
				return
			}
			historial, err := dbpkg.ListSuperCorreosMasivos(dbSuper, 30)
			if err != nil {
				writeSuperCorreosMasivosPublicError(w, r, http.StatusInternalServerError, "listar historial", err, nil)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":             true,
				"admin_email":    adminEmail,
				"preview":        preview,
				"historial":      historial,
				"mail_test_mode": isEmpresaUsuarioMailTestMode(dbSuper),
				"max_envios":     superCorreosMasivosMaxDestinatarios,
			})
			return

		case http.MethodPost:
			var payload superCorreoMasivoPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "payload invalido"})
				return
			}
			payload = normalizeSuperCorreoMasivoPayload(payload)
			if err := validateSuperCorreoMasivoPayload(payload); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": superCorreoMasivoValidationMessage(err)})
				return
			}

			recipients, preview, err := collectSuperCorreoMasivoRecipients(dbEmpresas, dbSuper, payload.Alcance)
			if err != nil {
				writeSuperCorreosMasivosPublicError(w, r, http.StatusInternalServerError, "resolver destinatarios", err, nil)
				return
			}
			if payload.SoloPreview || !payload.Confirmar {
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "preview": preview, "requires_confirmation": true})
				return
			}
			if len(recipients) == 0 {
				writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "no hay destinatarios validos para el alcance seleccionado"})
				return
			}
			if len(recipients) > superCorreosMasivosMaxDestinatarios {
				writeJSON(w, http.StatusPreconditionFailed, map[string]interface{}{"ok": false, "error": fmt.Sprintf("el lote tiene %d destinatarios y supera el limite operativo de %d", len(recipients), superCorreosMasivosMaxDestinatarios), "preview": preview})
				return
			}

			result, err := sendSuperCorreoMasivoCampaign(r, dbSuper, adminEmail, payload, recipients, preview)
			if err != nil {
				writeSuperCorreosMasivosPublicError(w, r, http.StatusInternalServerError, "ejecutar campana", err, map[string]interface{}{"result": result})
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "result": result})
			return

		default:
			w.Header().Set("Allow", "GET, POST")
			writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"ok": false, "error": "metodo no permitido"})
			return
		}
	}
}

func writeSuperCorreosMasivosPublicError(w http.ResponseWriter, r *http.Request, status int, operation string, err error, extra map[string]interface{}) {
	requestID := resolveAuditoriaRequestID(r)
	log.Printf("[super_correos_masivos] operation=%s request_id=%s error_type=%T", operation, requestID, err)
	payload := map[string]interface{}{
		"ok":    false,
		"code":  "super_correos_masivos_error",
		"error": "No se pudo completar la operacion de correos masivos.",
	}
	for key, value := range extra {
		payload[key] = value
	}
	if requestID != "" {
		payload["request_id"] = requestID
	}
	writeJSON(w, status, payload)
}

func normalizeSuperCorreoMasivoPayload(payload superCorreoMasivoPayload) superCorreoMasivoPayload {
	payload.Categoria = normalizeSuperCorreoMasivoCategoria(payload.Categoria)
	payload.Alcance = normalizeSuperCorreoMasivoAlcance(payload.Alcance)
	payload.Asunto = strings.TrimSpace(payload.Asunto)
	payload.CuerpoTexto = strings.TrimSpace(payload.CuerpoTexto)
	payload.CuerpoHTML = strings.TrimSpace(payload.CuerpoHTML)
	payload.Observaciones = strings.TrimSpace(payload.Observaciones)
	return payload
}

func normalizeSuperCorreoMasivoCategoria(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "politicas", "politica", "nuevas_politicas":
		return "politicas"
	case "actualizaciones", "actualizacion":
		return "actualizaciones"
	case "mantenimiento", "mantenimientos":
		return "mantenimiento"
	case "seguridad":
		return "seguridad"
	case "informacion", "informacion_importante", "":
		return "informacion"
	default:
		return "otro"
	}
}

func normalizeSuperCorreoMasivoAlcance(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "administradores", "admins", "super":
		return "administradores"
	case "usuarios_empresa", "usuarios", "empresas":
		return "usuarios_empresa"
	default:
		return "todos"
	}
}

func validateSuperCorreoMasivoPayload(payload superCorreoMasivoPayload) error {
	if len([]rune(payload.Asunto)) < 6 || len([]rune(payload.Asunto)) > 180 {
		return fmt.Errorf("el asunto debe tener entre 6 y 180 caracteres")
	}
	if len([]rune(payload.CuerpoTexto)) < 20 || len([]rune(payload.CuerpoTexto)) > 20000 {
		return fmt.Errorf("el mensaje debe tener entre 20 y 20000 caracteres")
	}
	if payload.CuerpoHTML != "" && len([]rune(payload.CuerpoHTML)) > 40000 {
		return fmt.Errorf("el HTML del mensaje no puede superar 40000 caracteres")
	}
	return nil
}

func superCorreoMasivoValidationMessage(err error) string {
	switch err.Error() {
	case "el asunto debe tener entre 6 y 180 caracteres":
		return "el asunto debe tener entre 6 y 180 caracteres"
	case "el mensaje debe tener entre 20 y 20000 caracteres":
		return "el mensaje debe tener entre 20 y 20000 caracteres"
	case "el HTML del mensaje no puede superar 40000 caracteres":
		return "el HTML del mensaje no puede superar 40000 caracteres"
	default:
		log.Printf("[super_correos_masivos] validation_error_type=%T", err)
		return "los datos del correo masivo no son validos"
	}
}

func buildSuperCorreoMasivoPreview(dbEmpresas, dbSuper *sql.DB, alcance string) (superCorreoMasivoPreview, error) {
	_, preview, err := collectSuperCorreoMasivoRecipients(dbEmpresas, dbSuper, alcance)
	return preview, err
}

func collectSuperCorreoMasivoRecipients(dbEmpresas, dbSuper *sql.DB, alcance string) ([]superCorreoMasivoRecipient, superCorreoMasivoPreview, error) {
	alcance = normalizeSuperCorreoMasivoAlcance(alcance)
	preview := superCorreoMasivoPreview{}
	byEmail := make(map[string]*superCorreoMasivoRecipient)

	add := func(item superCorreoMasivoRecipient) {
		item.Email = strings.TrimSpace(item.Email)
		if item.Email == "" {
			preview.InvalidosOmitidos++
			return
		}
		parsed, err := mail.ParseAddress(item.Email)
		if err != nil || strings.TrimSpace(parsed.Address) == "" {
			preview.InvalidosOmitidos++
			return
		}
		item.Email = strings.TrimSpace(parsed.Address)
		key := strings.ToLower(item.Email)
		if current, exists := byEmail[key]; exists {
			preview.DuplicadosOmitidos++
			if item.TipoDestinatario != "" && !containsString(current.Fuentes, item.TipoDestinatario) {
				current.Fuentes = append(current.Fuentes, item.TipoDestinatario)
				sort.Strings(current.Fuentes)
			}
			if current.Nombre == "" {
				current.Nombre = item.Nombre
			}
			if current.EmpresaID == 0 {
				current.EmpresaID = item.EmpresaID
				current.EmpresaNombre = item.EmpresaNombre
			}
			return
		}
		item.EmailMasked = maskEmailForSuperCorreoMasivo(item.Email)
		if len(item.Fuentes) == 0 && item.TipoDestinatario != "" {
			item.Fuentes = []string{item.TipoDestinatario}
		}
		byEmail[key] = &item
		switch item.TipoDestinatario {
		case "administrador":
			preview.Administradores++
		case "usuario_empresa":
			preview.UsuariosEmpresa++
		}
	}

	if alcance == "todos" || alcance == "administradores" {
		rows, err := dbpkg.ExecQueryCompat(dbSuper, `SELECT COALESCE(email, ''), COALESCE(name, ''), COALESCE(role, ''), COALESCE(estado, 'activo')
			FROM administradores
			WHERE TRIM(COALESCE(email, '')) <> ''
				AND LOWER(COALESCE(estado, 'activo')) <> 'inactivo'
			ORDER BY id ASC`)
		if err != nil {
			return nil, preview, err
		}
		for rows.Next() {
			var email, name, role, estado string
			if err := rows.Scan(&email, &name, &role, &estado); err != nil {
				_ = rows.Close()
				return nil, preview, err
			}
			_ = estado
			add(superCorreoMasivoRecipient{
				Email:            email,
				Nombre:           name,
				TipoDestinatario: "administrador",
				Rol:              role,
			})
		}
		if err := rows.Close(); err != nil {
			return nil, preview, err
		}
	}

	if alcance == "todos" || alcance == "usuarios_empresa" {
		if err := dbpkg.EnsureEmpresaUsuariosAuthSchema(dbEmpresas); err != nil {
			return nil, preview, err
		}
		rows, err := dbpkg.ExecQueryCompat(dbEmpresas, `SELECT
				COALESCE(u.email, ''),
				COALESCE(u.name, ''),
				COALESCE(u.role, ''),
				COALESCE(u.empresa_id, 0),
				COALESCE((
					SELECT e.nombre
					FROM empresas e
					WHERE e.id = u.empresa_id OR COALESCE(e.empresa_id, e.id) = u.empresa_id
					ORDER BY e.id ASC
					LIMIT 1
				), ''),
				COALESCE(u.estado, 'activo')
			FROM users u
			WHERE TRIM(COALESCE(u.email, '')) <> ''
				AND LOWER(COALESCE(u.estado, 'activo')) NOT IN ('eliminado', 'eliminada', 'borrado')
			ORDER BY u.empresa_id ASC, u.id ASC`)
		if err != nil {
			return nil, preview, err
		}
		for rows.Next() {
			var email, name, role, empresaNombre, estado string
			var empresaID int64
			if err := rows.Scan(&email, &name, &role, &empresaID, &empresaNombre, &estado); err != nil {
				_ = rows.Close()
				return nil, preview, err
			}
			_ = estado
			add(superCorreoMasivoRecipient{
				Email:            email,
				Nombre:           name,
				TipoDestinatario: "usuario_empresa",
				EmpresaID:        empresaID,
				EmpresaNombre:    empresaNombre,
				Rol:              role,
			})
		}
		if err := rows.Close(); err != nil {
			return nil, preview, err
		}
	}

	recipients := make([]superCorreoMasivoRecipient, 0, len(byEmail))
	for _, item := range byEmail {
		recipients = append(recipients, *item)
	}
	sort.Slice(recipients, func(i, j int) bool {
		if recipients[i].TipoDestinatario == recipients[j].TipoDestinatario {
			return strings.ToLower(recipients[i].Email) < strings.ToLower(recipients[j].Email)
		}
		return recipients[i].TipoDestinatario < recipients[j].TipoDestinatario
	})
	preview.Total = len(recipients)
	muestraLimit := 8
	if len(recipients) < muestraLimit {
		muestraLimit = len(recipients)
	}
	preview.Muestra = append(preview.Muestra, recipients[:muestraLimit]...)
	for i := range preview.Muestra {
		preview.Muestra[i].Email = ""
	}
	return recipients, preview, nil
}

func sendSuperCorreoMasivoCampaign(r *http.Request, dbSuper *sql.DB, adminEmail string, payload superCorreoMasivoPayload, recipients []superCorreoMasivoRecipient, preview superCorreoMasivoPreview) (map[string]interface{}, error) {
	testMode := isEmpresaUsuarioMailTestMode(dbSuper)
	codigo := buildSuperCorreoMasivoCodigo(payload.Asunto, adminEmail)
	metadata, _ := json.Marshal(map[string]interface{}{
		"preview":        preview,
		"remote_addr":    r.RemoteAddr,
		"user_agent_len": len(strings.TrimSpace(r.UserAgent())),
	})
	campaignID, err := dbpkg.CreateSuperCorreoMasivo(dbSuper, dbpkg.SuperCorreoMasivo{
		Codigo:             codigo,
		Categoria:          payload.Categoria,
		Alcance:            payload.Alcance,
		Asunto:             payload.Asunto,
		CuerpoTexto:        payload.CuerpoTexto,
		CuerpoHTML:         payload.CuerpoHTML,
		TotalDestinatarios: len(recipients),
		EstadoEnvio:        "en_proceso",
		ModoPrueba:         superCorreoMasivoBoolToInt(testMode),
		MetadataJSON:       string(metadata),
		UsuarioCreador:     adminEmail,
		Estado:             "activo",
		Observaciones:      payload.Observaciones,
	})
	if err != nil {
		return nil, err
	}

	enviados := 0
	fallidos := 0
	omitidos := 0
	for _, recipient := range recipients {
		resultado := "enviado"
		errDetalle := ""
		if testMode {
			resultado = "capturado_modo_prueba"
		} else if err := sendSuperCorreoMasivoSMTP(dbSuper, recipient.Email, recipient.Nombre, payload.Asunto, payload.CuerpoTexto, payload.CuerpoHTML); err != nil {
			resultado = "fallido"
			errDetalle = truncateForSuperCorreoMasivo(err.Error(), 700)
		}

		switch resultado {
		case "enviado", "capturado_modo_prueba":
			enviados++
		case "fallido":
			fallidos++
		default:
			omitidos++
		}
		_, _ = dbpkg.CreateSuperCorreoMasivoDestinatario(dbSuper, dbpkg.SuperCorreoMasivoDestinatario{
			CorreoMasivoID:   campaignID,
			Email:            recipient.Email,
			Nombre:           recipient.Nombre,
			TipoDestinatario: recipient.TipoDestinatario,
			EmpresaID:        recipient.EmpresaID,
			EmpresaNombre:    recipient.EmpresaNombre,
			Rol:              recipient.Rol,
			Resultado:        resultado,
			ErrorDetalle:     errDetalle,
			FechaEnvio:       time.Now().Format("2006-01-02 15:04:05"),
			UsuarioCreador:   adminEmail,
			Estado:           "activo",
			Observaciones:    strings.Join(recipient.Fuentes, ","),
		})
	}

	estado := "enviado"
	if testMode {
		estado = "simulado"
	} else if fallidos > 0 && enviados > 0 {
		estado = "parcial"
	} else if fallidos > 0 && enviados == 0 {
		estado = "fallido"
	}
	if err := dbpkg.UpdateSuperCorreoMasivoResultado(dbSuper, campaignID, enviados, fallidos, omitidos, estado, payload.Observaciones); err != nil {
		return map[string]interface{}{"campaign_id": campaignID, "codigo": codigo, "enviados": enviados, "fallidos": fallidos, "omitidos": omitidos, "estado_envio": estado}, err
	}

	return map[string]interface{}{
		"campaign_id":           campaignID,
		"codigo":                codigo,
		"estado_envio":          estado,
		"modo_prueba":           testMode,
		"total_destinatarios":   len(recipients),
		"enviados_o_capturados": enviados,
		"fallidos":              fallidos,
		"omitidos":              omitidos,
	}, nil
}

func sendSuperCorreoMasivoSMTP(dbSuper *sql.DB, toEmail, toName, subject, bodyPlain, bodyHTML string) error {
	fromName, fromEmail := corporateSystemSenderAddress(dbSuper, "soporte")
	bodyPlain = strings.TrimSpace(bodyPlain)
	if bodyHTML == "" {
		bodyHTML = buildSuperCorreoMasivoHTMLFromText(bodyPlain)
	}
	_ = strings.TrimSpace(toName)
	msg := buildEmpresaUsuarioMultipartMessage(dbSuper, "https://powerfulcontrolsystem.com", fromName, fromEmail, strings.TrimSpace(toEmail), strings.TrimSpace(subject), normalizeMailBodyCRLF(bodyPlain), strings.TrimSpace(bodyHTML))

	return sendEmpresaUsuarioMailuMessage(dbSuper, fromEmail, strings.TrimSpace(toEmail), []byte(msg))
}

func buildSuperCorreoMasivoHTMLFromText(body string) string {
	parts := strings.Split(strings.TrimSpace(body), "\n\n")
	var b strings.Builder
	b.WriteString("<html><body>")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		escaped := html.EscapeString(part)
		escaped = strings.ReplaceAll(escaped, "\n", "<br>")
		b.WriteString("<p>")
		b.WriteString(escaped)
		b.WriteString("</p>")
	}
	b.WriteString("<p>Powerful Control System</p>")
	b.WriteString("</body></html>")
	return b.String()
}

func buildSuperCorreoMasivoCodigo(subject, adminEmail string) string {
	now := time.Now()
	sum := sha256.Sum256([]byte(strings.TrimSpace(subject) + "|" + strings.TrimSpace(adminEmail) + "|" + strconv.FormatInt(now.UnixNano(), 10)))
	return "CM-" + now.Format("20060102-150405") + "-" + strings.ToUpper(hex.EncodeToString(sum[:])[:8])
}

func maskEmailForSuperCorreoMasivo(email string) string {
	email = strings.TrimSpace(email)
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***"
	}
	local := parts[0]
	if local == "" || parts[1] == "" {
		return "***"
	}
	if len(local) <= 2 {
		local = local[:1] + "*"
	} else {
		local = local[:1] + strings.Repeat("*", len(local)-2) + local[len(local)-1:]
	}
	domain := parts[1]
	domainParts := strings.Split(domain, ".")
	if len(domainParts[0]) > 1 {
		domainParts[0] = domainParts[0][:1] + strings.Repeat("*", len(domainParts[0])-1)
	}
	return local + "@" + strings.Join(domainParts, ".")
}

func containsString(items []string, target string) bool {
	target = strings.TrimSpace(target)
	for _, item := range items {
		if strings.TrimSpace(item) == target {
			return true
		}
	}
	return false
}

func superCorreoMasivoBoolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func truncateForSuperCorreoMasivo(raw string, max int) string {
	raw = strings.TrimSpace(raw)
	if max <= 0 || len([]rune(raw)) <= max {
		return raw
	}
	runes := []rune(raw)
	return string(runes[:max])
}

func normalizeMailBodyCRLF(body string) string {
	body = strings.ReplaceAll(body, "\r\n", "\n")
	body = strings.ReplaceAll(body, "\r", "\n")
	return strings.ReplaceAll(body, "\n", "\r\n")
}
