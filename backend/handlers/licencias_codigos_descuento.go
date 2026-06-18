package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
	Descripcion   string  `json:"descripcion,omitempty"`
	LineaOriginal string  `json:"linea_original,omitempty"`
}

type licenciaDiscountCodeAdminPayload struct {
	Codigo      string  `json:"codigo"`
	OldCodigo   string  `json:"old_codigo,omitempty"`
	Tipo        string  `json:"tipo"`
	Valor       float64 `json:"valor"`
	Spec        string  `json:"spec,omitempty"`
	Activo      bool    `json:"activo"`
	Descripcion string  `json:"descripcion,omitempty"`
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
				http.Error(w, "no se pudieron leer los codigos: "+err.Error(), http.StatusInternalServerError)
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
			item, err := buildLicenciaDiscountCodeAdminItem(payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			items, _, err := readLicenciaDiscountCodeAdminItems(dbSuper)
			if err != nil {
				http.Error(w, "no se pudieron leer los codigos: "+err.Error(), http.StatusInternalServerError)
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
				http.Error(w, "no se pudo guardar el codigo: "+err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "item": item})
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
				http.Error(w, "no se pudieron leer los codigos: "+err.Error(), http.StatusInternalServerError)
				return
			}
			found := false
			for i, existing := range items {
				if strings.EqualFold(existing.Codigo, oldCode) {
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
				http.Error(w, "no se pudo guardar el codigo: "+err.Error(), http.StatusInternalServerError)
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
				http.Error(w, "no se pudieron leer los codigos: "+err.Error(), http.StatusInternalServerError)
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
				http.Error(w, "no se pudo eliminar el codigo: "+err.Error(), http.StatusInternalServerError)
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
	if code == "" || email == "" {
		http.Error(w, "codigo y email son obligatorios", http.StatusBadRequest)
		return
	}
	items, _, err := readLicenciaDiscountCodeAdminItems(dbSuper)
	if err != nil {
		http.Error(w, "no se pudieron leer los codigos: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var item *licenciaDiscountCodeAdminItem
	for i := range items {
		if strings.EqualFold(items[i].Codigo, code) {
			item = &items[i]
			break
		}
	}
	if item == nil {
		http.Error(w, "codigo no encontrado", http.StatusNotFound)
		return
	}
	subject := "Codigo de descuento " + item.Codigo
	status := "Activo"
	if !item.Activo {
		status = "Inactivo"
	}
	body := fmt.Sprintf("Codigo de descuento: %s\nDescuento: %s\nEstado: %s\nUso: checkout de licencias de Powerful Control System.\nRegla: un uso por empresa.\n", item.Codigo, firstNonEmptyString(item.Descripcion, item.Spec), status)
	html := fmt.Sprintf("<html><body><h2>Codigo de descuento</h2><p><strong>Codigo:</strong> %s</p><p><strong>Descuento:</strong> %s</p><p><strong>Estado:</strong> %s</p><p>Uso: checkout de licencias de Powerful Control System. Regla: un uso por empresa.</p></body></html>", htmlEscape(item.Codigo), htmlEscape(firstNonEmptyString(item.Descripcion, item.Spec)), htmlEscape(status))
	metadata := fmt.Sprintf(`{"codigo":%q,"tipo":%q,"valor":%g}`, item.Codigo, item.Tipo, item.Valor)
	if err := sendPCSSystemEmail(dbSuper, email, payload.Nombre, subject, body, html, "licencias_codigo_descuento", metadata, actorEmail); err != nil {
		if isEmpresaUsuarioMailActionableConfigError(err) {
			captureErr := captureEmpresaUsuarioMailNotification(dbSuper, "licencias_codigo_descuento", 0, email, subject, body, "", metadata, actorEmail)
			if captureErr == nil {
				writeJSON(w, http.StatusAccepted, map[string]interface{}{
					"ok":       true,
					"codigo":   item.Codigo,
					"email":    email,
					"sent":     false,
					"captured": true,
					"warning":  "SMTP no disponible; correo capturado en notificaciones de prueba",
				})
				return
			}
		}
		http.Error(w, "no se pudo enviar el correo", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "codigo": item.Codigo, "email": email, "sent": true})
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
		item := licenciaDiscountCodeAdminItem{
			Codigo:        code,
			Spec:          spec,
			Activo:        active,
			LineaOriginal: originalLine,
		}
		item.Tipo, item.Valor, item.Descripcion = describeLicenciaDiscountSpec(spec)
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
	if _, _, ok := parseLicenciaDiscountSpec(spec, 100000); !ok {
		return licenciaDiscountCodeAdminItem{}, fmt.Errorf("descuento invalido; usa porcentaje, valor fijo o gratis")
	}
	item := licenciaDiscountCodeAdminItem{
		Codigo:      code,
		Spec:        spec,
		Activo:      payload.Activo,
		Descripcion: strings.TrimSpace(payload.Descripcion),
	}
	item.Tipo, item.Valor, item.Descripcion = describeLicenciaDiscountSpec(spec)
	return item, nil
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
	spec = strings.TrimSpace(spec)
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
		line := code + "=" + spec
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
