package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"sort"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const licenciaDiscountCodesConfigKey = "licencias.discount_codes"

type licenciaDiscountCodeAdminItem struct {
	Codigo        string  `json:"codigo"`
	Spec          string  `json:"spec"`
	Tipo          string  `json:"tipo"`
	Valor         float64 `json:"valor"`
	Activo        bool    `json:"activo"`
	Nombre        string  `json:"nombre,omitempty"`
	Descripcion   string  `json:"descripcion,omitempty"`
	Email         string  `json:"email,omitempty"`
	Vence         string  `json:"vence,omitempty"`
	EnviadoEmail  bool    `json:"enviado_email,omitempty"`
	UltimoEnvio   string  `json:"ultimo_envio,omitempty"`
	LineaOriginal string  `json:"linea_original,omitempty"`
}

type licenciaDiscountCodeAdminPayload struct {
	Codigo      string  `json:"codigo"`
	OldCodigo   string  `json:"old_codigo,omitempty"`
	Tipo        string  `json:"tipo"`
	Valor       float64 `json:"valor"`
	Spec        string  `json:"spec,omitempty"`
	Activo      bool    `json:"activo"`
	Nombre      string  `json:"nombre,omitempty"`
	Descripcion string  `json:"descripcion,omitempty"`
	Email       string  `json:"email,omitempty"`
	Vence       string  `json:"vence,omitempty"`
}

type licenciaDiscountCodeEmailPayload struct {
	Codigo string `json:"codigo"`
	Email  string `json:"email"`
	Nombre string `json:"nombre,omitempty"`
}

// SuperLicenciasCodigosDescuentoHandler administra los codigos promocionales
// globales que el checkout de licencias ya valida y limita a un uso por empresa.
func SuperLicenciasCodigosDescuentoHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		admin, ok, status, msg := requireSuperAdmin(r, dbSuper)
		if !ok {
			http.Error(w, msg, status)
			return
		}

		switch r.Method {
		case http.MethodGet:
			items, raw, err := readLicenciaDiscountCodeAdminItems(dbSuper)
			if err != nil {
				http.Error(w, "no se pudieron leer los codigos", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":         true,
				"items":      items,
				"raw":        raw,
				"updated_by": strings.TrimSpace(admin.Email),
				"config_key": licenciaDiscountCodesConfigKey,
				"usage_rule": "un_uso_por_empresa",
				"updated_at": time.Now().Format(time.RFC3339),
			})
		case http.MethodPost:
			if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("action")), "enviar_correo") {
				handleLicenciaDiscountCodeEmail(w, r, dbSuper, admin.Email)
				return
			}
			var payload licenciaDiscountCodeAdminPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido", http.StatusBadRequest)
				return
			}
			if normalizeLicenciaDiscountCode(payload.Codigo) == "" {
				generated, err := generateLicenciaDiscountCode()
				if err != nil {
					http.Error(w, "no se pudo generar el codigo", http.StatusInternalServerError)
					return
				}
				payload.Codigo = generated
			}
			item, err := buildLicenciaDiscountCodeAdminItem(payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			items, _, err := readLicenciaDiscountCodeAdminItems(dbSuper)
			if err != nil {
				http.Error(w, "no se pudieron leer los codigos", http.StatusInternalServerError)
				return
			}
			for _, existing := range items {
				if strings.EqualFold(existing.Codigo, item.Codigo) {
					http.Error(w, "ya existe un codigo con ese nombre", http.StatusConflict)
					return
				}
			}
			items = append(items, item)
			if err := saveLicenciaDiscountCodeAdminItems(dbSuper, items, admin.Email); err != nil {
				http.Error(w, "no se pudo guardar el codigo", http.StatusInternalServerError)
				return
			}
			sent := false
			if strings.TrimSpace(item.Email) != "" {
				if err := sendLicenciaDiscountCodeEmail(dbSuper, &item, item.Email, item.Nombre, admin.Email); err == nil {
					item.EnviadoEmail = true
					item.UltimoEnvio = time.Now().Format(time.RFC3339)
					items[len(items)-1] = item
					_ = saveLicenciaDiscountCodeAdminItems(dbSuper, items, admin.Email)
					sent = true
				}
			}
			writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "item": item, "email_sent": sent})
		case http.MethodPut:
			var payload licenciaDiscountCodeAdminPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido", http.StatusBadRequest)
				return
			}
			item, err := buildLicenciaDiscountCodeAdminItem(payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			oldCode := normalizeLicenciaDiscountCode(firstNonEmptyString(payload.OldCodigo, payload.Codigo))
			if oldCode == "" {
				http.Error(w, "old_codigo o codigo es obligatorio", http.StatusBadRequest)
				return
			}
			items, _, err := readLicenciaDiscountCodeAdminItems(dbSuper)
			if err != nil {
				http.Error(w, "no se pudieron leer los codigos", http.StatusInternalServerError)
				return
			}
			found := false
			for i, existing := range items {
				if strings.EqualFold(existing.Codigo, oldCode) {
					if item.Email == "" {
						item.Email = existing.Email
					}
					if item.Nombre == "" {
						item.Nombre = existing.Nombre
					}
					if item.Vence == "" {
						item.Vence = existing.Vence
					}
					item.EnviadoEmail = existing.EnviadoEmail
					item.UltimoEnvio = existing.UltimoEnvio
					items[i] = item
					found = true
					continue
				}
				if strings.EqualFold(existing.Codigo, item.Codigo) {
					http.Error(w, "ya existe otro codigo con ese nombre", http.StatusConflict)
					return
				}
			}
			if !found {
				http.Error(w, "codigo no encontrado", http.StatusNotFound)
				return
			}
			if err := saveLicenciaDiscountCodeAdminItems(dbSuper, items, admin.Email); err != nil {
				http.Error(w, "no se pudo guardar el codigo", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "item": item})
		case http.MethodDelete:
			code := normalizeLicenciaDiscountCode(r.URL.Query().Get("codigo"))
			if code == "" {
				http.Error(w, "codigo es obligatorio", http.StatusBadRequest)
				return
			}
			items, _, err := readLicenciaDiscountCodeAdminItems(dbSuper)
			if err != nil {
				http.Error(w, "no se pudieron leer los codigos", http.StatusInternalServerError)
				return
			}
			next := items[:0]
			found := false
			for _, item := range items {
				if strings.EqualFold(item.Codigo, code) {
					found = true
					continue
				}
				next = append(next, item)
			}
			if !found {
				http.Error(w, "codigo no encontrado", http.StatusNotFound)
				return
			}
			if err := saveLicenciaDiscountCodeAdminItems(dbSuper, next, admin.Email); err != nil {
				http.Error(w, "no se pudo eliminar el codigo", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
		default:
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
		}
	}
}

func handleLicenciaDiscountCodeEmail(w http.ResponseWriter, r *http.Request, dbSuper *sql.DB, actorEmail string) {
	var payload licenciaDiscountCodeEmailPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "payload invalido", http.StatusBadRequest)
		return
	}
	code := normalizeLicenciaDiscountCode(payload.Codigo)
	email := strings.ToLower(strings.TrimSpace(payload.Email))
	if code == "" {
		http.Error(w, "codigo es obligatorio", http.StatusBadRequest)
		return
	}
	items, _, err := readLicenciaDiscountCodeAdminItems(dbSuper)
	if err != nil {
		http.Error(w, "no se pudieron leer los codigos", http.StatusInternalServerError)
		return
	}
	itemIndex := -1
	var item *licenciaDiscountCodeAdminItem
	for i := range items {
		if strings.EqualFold(items[i].Codigo, code) {
			itemIndex = i
			item = &items[i]
			break
		}
	}
	if item == nil {
		http.Error(w, "codigo no encontrado", http.StatusNotFound)
		return
	}
	if email == "" {
		email = strings.TrimSpace(item.Email)
	}
	if email == "" {
		http.Error(w, "email es obligatorio", http.StatusBadRequest)
		return
	}
	if _, err := mail.ParseAddress(email); err != nil {
		http.Error(w, "email invalido", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.Nombre) != "" {
		item.Nombre = strings.TrimSpace(payload.Nombre)
	}
	item.Email = email
	err = sendLicenciaDiscountCodeEmail(dbSuper, item, email, item.Nombre, actorEmail)
	if err != nil {
		subject, body, metadata := licenciaDiscountCodeEmailContent(item)
		captureErr := captureEmpresaUsuarioMailNotification(dbSuper, "licencias_codigo_descuento", 0, email, subject, body, "", metadata, actorEmail)
		writeJSON(w, http.StatusAccepted, map[string]interface{}{
			"ok":            true,
			"codigo":        item.Codigo,
			"email":         email,
			"sent":          false,
			"captured":      captureErr == nil,
			"smtp_warning":  "El correo corporativo no confirmo el envio; revisar Email corporativo Mailu",
			"capture_error": captureErr != nil,
		})
		return
	}
	item.EnviadoEmail = true
	item.UltimoEnvio = time.Now().Format(time.RFC3339)
	if itemIndex >= 0 {
		items[itemIndex] = *item
		_ = saveLicenciaDiscountCodeAdminItems(dbSuper, items, actorEmail)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "codigo": item.Codigo, "email": email, "sent": true})
}

func licenciaDiscountCodeEmailContent(item *licenciaDiscountCodeAdminItem) (string, string, string) {
	if item == nil {
		return "Codigo de descuento Powerful Control System", "", "{}"
	}
	status := "Activo"
	if !item.Activo {
		status = "Inactivo"
	}
	discount := firstNonEmptyString(item.Descripcion, item.Spec)
	expiry := strings.TrimSpace(item.Vence)
	if expiry == "" {
		expiry = "Sin fecha de vencimiento"
	}
	subject := "Tu codigo de descuento PCS: " + item.Codigo
	body := fmt.Sprintf("Hola %s,\n\nTu codigo de descuento para licencias de Powerful Control System es: %s\nDescuento: %s\nVence: %s\nEstado: %s\n\nUsalo en el checkout de licencias. Cada codigo se puede usar una vez por empresa.\n", firstNonEmptyString(item.Nombre, "cliente"), item.Codigo, discount, expiry, status)
	metadata := fmt.Sprintf(`{"codigo":%q,"tipo":%q,"valor":%g,"vence":%q}`, item.Codigo, item.Tipo, item.Valor, item.Vence)
	return subject, body, metadata
}

func sendLicenciaDiscountCodeEmail(dbSuper *sql.DB, item *licenciaDiscountCodeAdminItem, email, toName, actorEmail string) error {
	subject, body, metadata := licenciaDiscountCodeEmailContent(item)
	htmlBody := fmt.Sprintf(`<html><body><h2>Codigo de descuento PCS</h2><p>Hola %s,</p><p>Tu codigo para licencias de Powerful Control System es:</p><p style="font-size:22px;font-weight:800;letter-spacing:.08em">%s</p><p><strong>Descuento:</strong> %s<br><strong>Vence:</strong> %s</p><p>Usalo en el checkout de licencias. Cada codigo se puede usar una vez por empresa.</p></body></html>`,
		htmlEscape(firstNonEmptyString(toName, "cliente")),
		htmlEscape(item.Codigo),
		htmlEscape(firstNonEmptyString(item.Descripcion, item.Spec)),
		htmlEscape(firstNonEmptyString(item.Vence, "Sin fecha de vencimiento")),
	)
	return sendPCSSystemEmail(dbSuper, email, toName, subject, body, htmlBody, "licencias_codigo_descuento", metadata, actorEmail)
}

func readLicenciaDiscountCodeAdminItems(dbSuper *sql.DB) ([]licenciaDiscountCodeAdminItem, string, error) {
	raw, _, err := dbpkg.GetConfigValue(dbSuper, licenciaDiscountCodesConfigKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			raw = ""
		} else {
			return nil, "", err
		}
	}
	if strings.TrimSpace(raw) == "" {
		if legacyRaw, _, legacyErr := dbpkg.GetConfigValue(dbSuper, "licencias.codigos_descuento"); legacyErr == nil {
			raw = legacyRaw
		} else if legacyErr != nil && !errors.Is(legacyErr, sql.ErrNoRows) {
			return nil, "", legacyErr
		}
	}
	items := parseLicenciaDiscountCodeAdminLines(raw)
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Activo != items[j].Activo {
			return items[i].Activo
		}
		return items[i].Codigo < items[j].Codigo
	})
	return items, raw, nil
}

func parseLicenciaDiscountCodeAdminLines(raw string) []licenciaDiscountCodeAdminItem {
	out := make([]licenciaDiscountCodeAdminItem, 0)
	seen := map[string]struct{}{}
	for _, line := range strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n") {
		originalLine := strings.TrimSpace(line)
		if originalLine == "" {
			continue
		}
		active := true
		entry := originalLine
		if strings.HasPrefix(entry, "#") {
			active = false
			entry = strings.TrimSpace(strings.TrimPrefix(entry, "#"))
		}
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			continue
		}
		code := normalizeLicenciaDiscountCode(parts[0])
		if code == "" {
			continue
		}
		if _, exists := seen[code]; exists {
			continue
		}
		seen[code] = struct{}{}
		spec := strings.TrimSpace(parts[1])
		discountSpec, meta := splitLicenciaDiscountSpecMetadata(spec)
		item := licenciaDiscountCodeAdminItem{
			Codigo:        code,
			Spec:          discountSpec,
			Activo:        active,
			LineaOriginal: originalLine,
		}
		item.Tipo, item.Valor, item.Descripcion = describeLicenciaDiscountSpec(discountSpec)
		item.applyMetadata(meta)
		out = append(out, item)
	}
	return out
}

func buildLicenciaDiscountCodeAdminItem(payload licenciaDiscountCodeAdminPayload) (licenciaDiscountCodeAdminItem, error) {
	code := normalizeLicenciaDiscountCode(payload.Codigo)
	if code == "" {
		return licenciaDiscountCodeAdminItem{}, fmt.Errorf("codigo es obligatorio")
	}
	if len(code) < 3 || len(code) > 40 {
		return licenciaDiscountCodeAdminItem{}, fmt.Errorf("codigo debe tener entre 3 y 40 caracteres")
	}
	for _, r := range code {
		if !(r >= 'A' && r <= 'Z') && !(r >= '0' && r <= '9') && r != '-' && r != '_' {
			return licenciaDiscountCodeAdminItem{}, fmt.Errorf("codigo solo puede usar letras, numeros, guion o guion bajo")
		}
	}
	spec := strings.TrimSpace(payload.Spec)
	if spec == "" {
		spec = buildLicenciaDiscountSpecFromPayload(payload)
	}
	if _, _, ok := parseLicenciaDiscountSpec(splitLicenciaDiscountSpecOnly(spec), 100000); !ok {
		return licenciaDiscountCodeAdminItem{}, fmt.Errorf("descuento invalido; usa porcentaje, valor fijo o gratis")
	}
	item := licenciaDiscountCodeAdminItem{
		Codigo:      code,
		Spec:        splitLicenciaDiscountSpecOnly(spec),
		Activo:      payload.Activo,
		Nombre:      strings.TrimSpace(payload.Nombre),
		Descripcion: strings.TrimSpace(payload.Descripcion),
		Email:       strings.ToLower(strings.TrimSpace(payload.Email)),
		Vence:       strings.TrimSpace(payload.Vence),
	}
	if item.Email != "" {
		if _, err := mail.ParseAddress(item.Email); err != nil {
			return licenciaDiscountCodeAdminItem{}, fmt.Errorf("email invalido")
		}
	}
	if item.Vence != "" {
		if _, err := time.Parse("2006-01-02", item.Vence); err != nil {
			return licenciaDiscountCodeAdminItem{}, fmt.Errorf("fecha de vencimiento invalida")
		}
	}
	item.Tipo, item.Valor, _ = describeLicenciaDiscountSpec(item.Spec)
	if item.Descripcion == "" {
		_, _, item.Descripcion = describeLicenciaDiscountSpec(item.Spec)
	}
	return item, nil
}

func generateLicenciaDiscountCode() (string, error) {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "PCS-" + strings.ToUpper(hex.EncodeToString(b[:])), nil
}

func splitLicenciaDiscountSpecOnly(spec string) string {
	discount, _ := splitLicenciaDiscountSpecMetadata(spec)
	return discount
}

func splitLicenciaDiscountSpecMetadata(spec string) (string, map[string]string) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return "", nil
	}
	parts := strings.SplitN(spec, "|", 2)
	discount := strings.TrimSpace(parts[0])
	if len(parts) < 2 {
		return discount, nil
	}
	raw := strings.TrimSpace(parts[1])
	if raw == "" {
		return discount, nil
	}
	meta := map[string]string{}
	if strings.HasPrefix(raw, "{") {
		_ = json.Unmarshal([]byte(raw), &meta)
	}
	return discount, meta
}

func licenciaDiscountCodeExpired(spec string, now time.Time) bool {
	_, meta := splitLicenciaDiscountSpecMetadata(spec)
	if len(meta) == 0 {
		return false
	}
	vence := strings.TrimSpace(meta["vence"])
	if vence == "" {
		return false
	}
	day, err := time.Parse("2006-01-02", vence)
	if err != nil {
		return true
	}
	return now.After(day.Add(24 * time.Hour))
}

func (item *licenciaDiscountCodeAdminItem) applyMetadata(meta map[string]string) {
	if item == nil || len(meta) == 0 {
		return
	}
	item.Nombre = strings.TrimSpace(meta["nombre"])
	if descripcion := strings.TrimSpace(meta["descripcion"]); descripcion != "" {
		item.Descripcion = descripcion
	}
	item.Email = strings.ToLower(strings.TrimSpace(meta["email"]))
	item.Vence = strings.TrimSpace(meta["vence"])
	item.UltimoEnvio = strings.TrimSpace(meta["ultimo_envio"])
	item.EnviadoEmail = parseBoolConfigValue(meta["enviado_email"])
}

func formatLicenciaDiscountCodeLineSpec(item licenciaDiscountCodeAdminItem) string {
	spec := splitLicenciaDiscountSpecOnly(item.Spec)
	meta := map[string]string{}
	if nombre := strings.TrimSpace(item.Nombre); nombre != "" {
		meta["nombre"] = nombre
	}
	if descripcion := strings.TrimSpace(item.Descripcion); descripcion != "" {
		meta["descripcion"] = descripcion
	}
	if email := strings.ToLower(strings.TrimSpace(item.Email)); email != "" {
		meta["email"] = email
	}
	if vence := strings.TrimSpace(item.Vence); vence != "" {
		meta["vence"] = vence
	}
	if item.EnviadoEmail {
		meta["enviado_email"] = "true"
	}
	if ultimoEnvio := strings.TrimSpace(item.UltimoEnvio); ultimoEnvio != "" {
		meta["ultimo_envio"] = ultimoEnvio
	}
	if len(meta) == 0 {
		return spec
	}
	raw, err := json.Marshal(meta)
	if err != nil {
		return spec
	}
	return spec + " | " + string(raw)
}

func buildLicenciaDiscountSpecFromPayload(payload licenciaDiscountCodeAdminPayload) string {
	tipo := strings.ToLower(strings.TrimSpace(payload.Tipo))
	switch tipo {
	case "gratis", "total", "cortesia":
		return "gratis"
	case "valor", "valor_fijo", "monto":
		if payload.Valor < 0 {
			payload.Valor = 0
		}
		return strconv.FormatFloat(payload.Valor, 'f', 2, 64)
	default:
		if payload.Valor < 0 {
			payload.Valor = 0
		}
		if payload.Valor > 100 {
			payload.Valor = 100
		}
		return strconv.FormatFloat(payload.Valor, 'f', 2, 64) + "%"
	}
}

func describeLicenciaDiscountSpec(spec string) (string, float64, string) {
	spec = splitLicenciaDiscountSpecOnly(spec)
	lower := strings.ToLower(spec)
	switch lower {
	case "gratis", "cortesia", "free", "full", "100%", "total0", "total_cero":
		return "gratis", 100, "Descuento total"
	}
	if strings.HasSuffix(lower, "%") {
		pctRaw := strings.TrimSpace(strings.TrimSuffix(lower, "%"))
		pct, _ := strconv.ParseFloat(strings.ReplaceAll(pctRaw, ",", "."), 64)
		if pct < 0 {
			pct = 0
		}
		if pct > 100 {
			pct = 100
		}
		return "porcentaje", pct, fmt.Sprintf("%.2f%%", pct)
	}
	amount, _ := strconv.ParseFloat(strings.ReplaceAll(strings.ReplaceAll(lower, ".", ""), ",", "."), 64)
	if amount < 0 {
		amount = 0
	}
	return "valor", amount, "Valor fijo"
}

func saveLicenciaDiscountCodeAdminItems(dbSuper *sql.DB, items []licenciaDiscountCodeAdminItem, actor string) error {
	sort.SliceStable(items, func(i, j int) bool { return items[i].Codigo < items[j].Codigo })
	lines := make([]string, 0, len(items))
	for _, item := range items {
		code := normalizeLicenciaDiscountCode(item.Codigo)
		spec := strings.TrimSpace(item.Spec)
		if code == "" || spec == "" {
			continue
		}
		line := code + "=" + formatLicenciaDiscountCodeLineSpec(item)
		if !item.Activo {
			line = "# " + line
		}
		lines = append(lines, line)
	}
	if err := dbpkg.SetConfigValue(dbSuper, licenciaDiscountCodesConfigKey, strings.Join(lines, "\n"), false); err != nil {
		return err
	}
	_ = dbpkg.SetConfigValue(dbSuper, licenciaDiscountCodesConfigKey+".updated_by", strings.TrimSpace(actor), false)
	_ = dbpkg.SetConfigValue(dbSuper, licenciaDiscountCodesConfigKey+".updated_at", time.Now().Format(time.RFC3339), false)
	return nil
}
