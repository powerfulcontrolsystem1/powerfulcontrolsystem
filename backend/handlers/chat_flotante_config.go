package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	chatFlotanteChatEnabledKey         = "chat_flotante.chat_enabled"
	chatFlotanteRobotEnabledKey        = "chat_flotante.robot_enabled"
	chatFlotanteVoiceEnabledKey        = "chat_flotante.voice_enabled"
	chatFlotanteRobotVoiceKey          = "chat_flotante.robot_voice"
	chatFlotantePersonalityModeKey     = "chat_flotante.personality_mode"
	chatFlotanteRadioOnlineEnabledKey  = "chat_flotante.radio_online_enabled"
	chatFlotanteRadioCountryKey        = "chat_flotante.radio_country"
	chatFlotanteRadioCustomStationsKey = "chat_flotante.radio_custom_stations"
	chatFlotanteThemeKey               = "chat_flotante.theme"
	chatFlotanteTextSizeKey            = "chat_flotante.text_size"
	chatFlotantePersonalityNormal      = "normal"
	chatFlotantePersonalityRobot       = "robot"
	chatFlotantePersonalitySecretary   = "secretary"
)

type chatFlotanteRadioStationPref struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Tagline     string `json:"tagline"`
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	Genre       string `json:"genre"`
	StreamURL   string `json:"streamUrl"`
	SourceURL   string `json:"sourceUrl,omitempty"`
	Custom      bool   `json:"custom"`
}

func parseChatFlotanteBool(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "on", "activo", "enabled", "si", "yes":
		return true
	default:
		return false
	}
}

func getChatFlotanteEnabledDefaultTrue(dbSuper *sql.DB, key string) bool {
	if dbSuper == nil {
		return true
	}
	raw, _, _, _, err := dbpkg.GetConfigEntry(dbSuper, key)
	if err != nil || strings.TrimSpace(raw) == "" {
		return true
	}
	return parseChatFlotanteBool(raw)
}

func getChatFlotanteChatEnabled(dbSuper *sql.DB) bool {
	return getChatFlotanteEnabledDefaultTrue(dbSuper, chatFlotanteChatEnabledKey)
}

func getChatFlotanteRobotEnabled(dbSuper *sql.DB) bool {
	return false
}

func getChatFlotanteRadioOnlineEnabled(dbSuper *sql.DB) bool {
	return getChatFlotanteEnabledDefaultTrue(dbSuper, chatFlotanteRadioOnlineEnabledKey)
}

func getChatFlotanteVoiceEnabled(dbSuper *sql.DB) bool {
	if dbSuper == nil {
		return false
	}
	raw, _, _, _, err := dbpkg.GetConfigEntry(dbSuper, chatFlotanteVoiceEnabledKey)
	if err != nil {
		return false
	}
	return parseChatFlotanteBool(raw)
}

func normalizeChatFlotanteRobotVoice(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "es-co", "es_co", "colombia", "colombiana", "default", "predeterminada":
		return "es-CO"
	case "es-co-female", "es_co_female", "femenina", "mujer":
		return "es-CO-female"
	case "es-co-male", "es_co_male", "masculina", "hombre":
		return "es-CO-male"
	case "es-mx", "es_mx", "mexico", "mexicana":
		return "es-MX"
	case "es-es", "es_es", "espana", "españa", "castellano":
		return "es-ES"
	default:
		return "es-CO"
	}
}

func getChatFlotanteRobotVoice(dbSuper *sql.DB) string {
	if dbSuper == nil {
		return "es-CO"
	}
	raw, _, _, _, err := dbpkg.GetConfigEntry(dbSuper, chatFlotanteRobotVoiceKey)
	if err != nil {
		return "es-CO"
	}
	return normalizeChatFlotanteRobotVoice(raw)
}

func normalizeChatFlotantePersonalityMode(raw string) string {
	return chatFlotantePersonalityNormal
}

func normalizeChatFlotanteTheme(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "corporativo", "corporate", "rojo_azul", "red_blue":
		return "corporativo"
	case "oceano", "ocean", "azul":
		return "oceano"
	case "esmeralda", "emerald", "verde":
		return "esmeralda"
	case "vino", "wine", "borgona":
		return "vino"
	default:
		return "normal"
	}
}

func normalizeChatFlotanteTextSize(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "pequeno", "small", "s":
		return "pequeno"
	case "grande", "large", "l":
		return "grande"
	default:
		return "mediano"
	}
}

func normalizeChatFlotanteRadioCountry(raw string) string {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "PA", "PANAMA", "PANAMÁ":
		return "PA"
	case "EC", "ECUADOR":
		return "EC"
	default:
		return ""
	}
}

func chatFlotanteLimitText(raw string, maxLen int) string {
	value := strings.TrimSpace(raw)
	if maxLen <= 0 || len(value) <= maxLen {
		return value
	}
	return strings.TrimSpace(value[:maxLen])
}

func chatFlotanteSafeURL(raw string) string {
	value := strings.TrimSpace(raw)
	lower := strings.ToLower(value)
	if len(value) > 512 {
		return ""
	}
	if strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "http://") {
		return value
	}
	return ""
}

func chatFlotanteRadioSlug(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if ok {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "emisora"
	}
	if len(out) > 80 {
		out = strings.Trim(out[:80], "-")
	}
	return out
}

func sanitizeChatFlotanteRadioStations(raw json.RawMessage) ([]chatFlotanteRadioStationPref, string, error) {
	if len(raw) == 0 || strings.EqualFold(strings.TrimSpace(string(raw)), "null") {
		return []chatFlotanteRadioStationPref{}, "[]", nil
	}
	var items []chatFlotanteRadioStationPref
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, "", err
	}
	out := make([]chatFlotanteRadioStationPref, 0, len(items))
	seen := map[string]bool{}
	for _, item := range items {
		name := chatFlotanteLimitText(item.Name, 120)
		streamURL := chatFlotanteSafeURL(item.StreamURL)
		if name == "" || streamURL == "" {
			continue
		}
		countryCode := normalizeChatFlotanteRadioCountry(item.CountryCode)
		id := chatFlotanteLimitText(item.ID, 120)
		if id == "" {
			id = "custom-" + chatFlotanteRadioSlug(name)
		} else {
			id = "custom-" + chatFlotanteRadioSlug(id)
		}
		if seen[id] {
			id = id + "-" + strings.TrimSpace(strings.ReplaceAll(streamURL, "://", "-"))
			id = "custom-" + chatFlotanteRadioSlug(id)
		}
		seen[id] = true
		out = append(out, chatFlotanteRadioStationPref{
			ID:          id,
			Name:        name,
			Tagline:     chatFlotanteLimitText(item.Tagline, 220),
			Country:     chatFlotanteLimitText(item.Country, 80),
			CountryCode: countryCode,
			Genre:       chatFlotanteLimitText(item.Genre, 80),
			StreamURL:   streamURL,
			SourceURL:   chatFlotanteSafeURL(item.SourceURL),
			Custom:      true,
		})
		if len(out) >= 40 {
			break
		}
	}
	encoded, err := json.Marshal(out)
	if err != nil {
		return nil, "", err
	}
	return out, string(encoded), nil
}

func getChatFlotanteRadioCustomStations(dbSuper, dbEmp *sql.DB, empresaID int64) []chatFlotanteRadioStationPref {
	raw := getChatFlotanteStringForEmpresa(dbSuper, dbEmp, empresaID, chatFlotanteRadioCustomStationsKey, "[]")
	items, _, err := sanitizeChatFlotanteRadioStations(json.RawMessage(raw))
	if err != nil {
		return []chatFlotanteRadioStationPref{}
	}
	return items
}

func getChatFlotantePersonalityMode(dbSuper *sql.DB) string {
	if dbSuper == nil {
		return chatFlotantePersonalityNormal
	}
	raw, _, _, _, err := dbpkg.GetConfigEntry(dbSuper, chatFlotantePersonalityModeKey)
	if err != nil {
		return chatFlotantePersonalityNormal
	}
	return normalizeChatFlotantePersonalityMode(raw)
}

func getChatFlotanteEmpresaPref(dbEmp *sql.DB, empresaID int64, key string) (string, bool) {
	if dbEmp == nil || empresaID <= 0 {
		return "", false
	}
	if err := dbpkg.EnsureEmpresaEstacionPrefsSchema(dbEmp); err != nil {
		return "", false
	}
	pref, err := dbpkg.GetEmpresaEstacionPref(dbEmp, empresaID, 0, key)
	if err != nil || pref == nil || strings.TrimSpace(pref.Valor) == "" {
		return "", false
	}
	return strings.TrimSpace(pref.Valor), true
}

func setChatFlotanteEmpresaPref(dbEmp *sql.DB, empresaID int64, key, value, usuario string) error {
	if dbEmp == nil || empresaID <= 0 {
		return nil
	}
	if err := dbpkg.EnsureEmpresaEstacionPrefsSchema(dbEmp); err != nil {
		return err
	}
	_, err := dbpkg.UpsertEmpresaEstacionPref(dbEmp, dbpkg.EmpresaEstacionPref{
		EmpresaID:      empresaID,
		EstacionID:     0,
		Clave:          key,
		Valor:          strings.TrimSpace(value),
		UsuarioCreador: strings.TrimSpace(usuario),
		Estado:         "activo",
		Observaciones:  "Preferencia empresarial del chat flotante y emisora.",
	})
	return err
}

func getChatFlotanteBoolForEmpresa(dbSuper, dbEmp *sql.DB, empresaID int64, key string, globalFallback bool) bool {
	if empresaID > 0 && key == chatFlotanteChatEnabledKey {
		return true
	}
	if raw, ok := getChatFlotanteEmpresaPref(dbEmp, empresaID, key); ok {
		return parseChatFlotanteBool(raw)
	}
	if dbSuper == nil {
		return globalFallback
	}
	raw, _, _, _, err := dbpkg.GetConfigEntry(dbSuper, key)
	if err != nil || strings.TrimSpace(raw) == "" {
		return globalFallback
	}
	return parseChatFlotanteBool(raw)
}

func getChatFlotanteBoolExplicitForEmpresa(dbSuper, dbEmp *sql.DB, empresaID int64, key string, globalFallback bool) bool {
	if empresaID > 0 {
		if raw, ok := getChatFlotanteEmpresaPref(dbEmp, empresaID, key); ok {
			return parseChatFlotanteBool(raw)
		}
		return false
	}
	return getChatFlotanteBoolForEmpresa(dbSuper, dbEmp, empresaID, key, globalFallback)
}

func getChatFlotanteStringForEmpresa(dbSuper, dbEmp *sql.DB, empresaID int64, key, globalFallback string) string {
	if raw, ok := getChatFlotanteEmpresaPref(dbEmp, empresaID, key); ok {
		return raw
	}
	if dbSuper == nil {
		return globalFallback
	}
	raw, _, _, _, err := dbpkg.GetConfigEntry(dbSuper, key)
	if err != nil || strings.TrimSpace(raw) == "" {
		return globalFallback
	}
	return strings.TrimSpace(raw)
}

func chatFlotantePrefsResponse(dbSuper, dbEmp *sql.DB, empresaID int64) map[string]any {
	robotVoice := normalizeChatFlotanteRobotVoice(getChatFlotanteStringForEmpresa(dbSuper, dbEmp, empresaID, chatFlotanteRobotVoiceKey, "es-CO"))
	personalityMode := chatFlotantePersonalityNormal
	radioCountry := normalizeChatFlotanteRadioCountry(getChatFlotanteStringForEmpresa(dbSuper, dbEmp, empresaID, chatFlotanteRadioCountryKey, ""))
	theme := normalizeChatFlotanteTheme(getChatFlotanteStringForEmpresa(dbSuper, dbEmp, empresaID, chatFlotanteThemeKey, "normal"))
	textSize := normalizeChatFlotanteTextSize(getChatFlotanteStringForEmpresa(dbSuper, dbEmp, empresaID, chatFlotanteTextSizeKey, "mediano"))
	return map[string]any{
		"ok":         true,
		"empresa_id": empresaID,
		"scope": func() string {
			if empresaID > 0 {
				return "empresa"
			}
			return "global"
		}(),
		"chat_enabled":          getChatFlotanteBoolForEmpresa(dbSuper, dbEmp, empresaID, chatFlotanteChatEnabledKey, true),
		"robot_enabled":         false,
		"radio_online_enabled":  getChatFlotanteBoolExplicitForEmpresa(dbSuper, dbEmp, empresaID, chatFlotanteRadioOnlineEnabledKey, false),
		"radio_country":         radioCountry,
		"radio_custom_stations": getChatFlotanteRadioCustomStations(dbSuper, dbEmp, empresaID),
		"voice_enabled":         getChatFlotanteBoolForEmpresa(dbSuper, dbEmp, empresaID, chatFlotanteVoiceEnabledKey, false),
		"robot_voice":           robotVoice,
		"personality_mode":      personalityMode,
		"theme":                 theme,
		"text_size":             textSize,
	}
}

func chatFlotanteBoolValue(enabled bool) string {
	if enabled {
		return "1"
	}
	return "0"
}

// ChatFlotantePreferenciasHandler expone preferencias no sensibles del chat IA flotante.
func ChatFlotantePreferenciasHandler(dbSuper, dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !voiceStreamRequireSession(w, r, dbSuper) {
			return
		}

		empresaID, _ := parseInt64QueryOptional(r, "empresa_id")

		switch r.Method {
		case http.MethodGet:
			var ok bool
			empresaID, ok = requireDynamicDocumentEmpresaAccess(w, r, dbEmp, dbSuper, empresaID)
			if !ok {
				return
			}
			writeJSON(w, http.StatusOK, chatFlotantePrefsResponse(dbSuper, dbEmp, empresaID))
			return
		case http.MethodPost, http.MethodPut:
			var payload struct {
				EmpresaID           int64           `json:"empresa_id"`
				ChatEnabled         *bool           `json:"chat_enabled"`
				RobotEnabled        *bool           `json:"robot_enabled"`
				RadioOnlineEnabled  *bool           `json:"radio_online_enabled"`
				RadioCountry        *string         `json:"radio_country"`
				RadioCustomStations json.RawMessage `json:"radio_custom_stations"`
				VoiceEnabled        *bool           `json:"voice_enabled"`
				RobotVoice          string          `json:"robot_voice"`
				PersonalityMode     string          `json:"personality_mode"`
				Theme               string          `json:"theme"`
				TextSize            string          `json:"text_size"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido: "+err.Error(), http.StatusBadRequest)
				return
			}
			if empresaID <= 0 && payload.EmpresaID > 0 {
				empresaID = payload.EmpresaID
			}
			var ok bool
			empresaID, ok = requireDynamicDocumentEmpresaAccess(w, r, dbEmp, dbSuper, empresaID)
			if !ok {
				return
			}
			usuario := adminEmailFromRequest(r)
			if payload.ChatEnabled != nil {
				value := chatFlotanteBoolValue(*payload.ChatEnabled)
				if empresaID > 0 {
					value = chatFlotanteBoolValue(true)
				}
				if empresaID > 0 {
					if err := setChatFlotanteEmpresaPref(dbEmp, empresaID, chatFlotanteChatEnabledKey, value, usuario); err != nil {
						http.Error(w, "No se pudo guardar la configuracion del chat", http.StatusInternalServerError)
						return
					}
				} else if err := dbpkg.SetConfigValue(dbSuper, chatFlotanteChatEnabledKey, value, false); err != nil {
					http.Error(w, "No se pudo guardar la configuracion del chat", http.StatusInternalServerError)
					return
				}
			}
			if payload.RobotEnabled != nil {
				value := chatFlotanteBoolValue(false)
				if empresaID > 0 {
					if err := setChatFlotanteEmpresaPref(dbEmp, empresaID, chatFlotanteRobotEnabledKey, value, usuario); err != nil {
						http.Error(w, "No se pudo guardar la configuracion del chat", http.StatusInternalServerError)
						return
					}
				} else if err := dbpkg.SetConfigValue(dbSuper, chatFlotanteRobotEnabledKey, value, false); err != nil {
					http.Error(w, "No se pudo guardar la configuracion del chat", http.StatusInternalServerError)
					return
				}
			}
			if payload.RadioOnlineEnabled != nil {
				value := chatFlotanteBoolValue(*payload.RadioOnlineEnabled)
				if empresaID > 0 {
					if err := setChatFlotanteEmpresaPref(dbEmp, empresaID, chatFlotanteRadioOnlineEnabledKey, value, usuario); err != nil {
						http.Error(w, "No se pudo guardar la configuracion del chat", http.StatusInternalServerError)
						return
					}
				} else if err := dbpkg.SetConfigValue(dbSuper, chatFlotanteRadioOnlineEnabledKey, value, false); err != nil {
					http.Error(w, "No se pudo guardar la configuracion del chat", http.StatusInternalServerError)
					return
				}
			}
			if payload.RadioCountry != nil {
				value := normalizeChatFlotanteRadioCountry(*payload.RadioCountry)
				if empresaID > 0 {
					if err := setChatFlotanteEmpresaPref(dbEmp, empresaID, chatFlotanteRadioCountryKey, value, usuario); err != nil {
						http.Error(w, "No se pudo guardar la configuracion del chat", http.StatusInternalServerError)
						return
					}
				} else if err := dbpkg.SetConfigValue(dbSuper, chatFlotanteRadioCountryKey, value, false); err != nil {
					http.Error(w, "No se pudo guardar la configuracion del chat", http.StatusInternalServerError)
					return
				}
			}
			if payload.RadioCustomStations != nil {
				_, value, err := sanitizeChatFlotanteRadioStations(payload.RadioCustomStations)
				if err != nil {
					http.Error(w, "emisoras personalizadas invalidas: "+err.Error(), http.StatusBadRequest)
					return
				}
				if empresaID > 0 {
					if err := setChatFlotanteEmpresaPref(dbEmp, empresaID, chatFlotanteRadioCustomStationsKey, value, usuario); err != nil {
						http.Error(w, "No se pudo guardar la configuracion del chat", http.StatusInternalServerError)
						return
					}
				} else if err := dbpkg.SetConfigValue(dbSuper, chatFlotanteRadioCustomStationsKey, value, false); err != nil {
					http.Error(w, "No se pudo guardar la configuracion del chat", http.StatusInternalServerError)
					return
				}
			}
			if payload.VoiceEnabled != nil {
				value := chatFlotanteBoolValue(*payload.VoiceEnabled)
				if empresaID > 0 {
					if err := setChatFlotanteEmpresaPref(dbEmp, empresaID, chatFlotanteVoiceEnabledKey, value, usuario); err != nil {
						http.Error(w, "No se pudo guardar la configuracion del chat", http.StatusInternalServerError)
						return
					}
				} else if err := dbpkg.SetConfigValue(dbSuper, chatFlotanteVoiceEnabledKey, value, false); err != nil {
					http.Error(w, "No se pudo guardar la configuracion del chat", http.StatusInternalServerError)
					return
				}
			}
			if strings.TrimSpace(payload.PersonalityMode) != "" {
				value := chatFlotantePersonalityNormal
				if empresaID > 0 {
					if err := setChatFlotanteEmpresaPref(dbEmp, empresaID, chatFlotantePersonalityModeKey, value, usuario); err != nil {
						http.Error(w, "No se pudo guardar la configuracion del chat", http.StatusInternalServerError)
						return
					}
				} else if err := dbpkg.SetConfigValue(dbSuper, chatFlotantePersonalityModeKey, value, false); err != nil {
					http.Error(w, "No se pudo guardar la configuracion del chat", http.StatusInternalServerError)
					return
				}
			}
			if strings.TrimSpace(payload.RobotVoice) != "" {
				value := normalizeChatFlotanteRobotVoice(payload.RobotVoice)
				if empresaID > 0 {
					if err := setChatFlotanteEmpresaPref(dbEmp, empresaID, chatFlotanteRobotVoiceKey, value, usuario); err != nil {
						http.Error(w, "No se pudo guardar la configuracion del chat", http.StatusInternalServerError)
						return
					}
				} else if err := dbpkg.SetConfigValue(dbSuper, chatFlotanteRobotVoiceKey, value, false); err != nil {
					http.Error(w, "No se pudo guardar la configuracion del chat", http.StatusInternalServerError)
					return
				}
			}
			if strings.TrimSpace(payload.Theme) != "" {
				value := normalizeChatFlotanteTheme(payload.Theme)
				if empresaID > 0 {
					if err := setChatFlotanteEmpresaPref(dbEmp, empresaID, chatFlotanteThemeKey, value, usuario); err != nil {
						http.Error(w, "No se pudo guardar la configuracion visual del chat", http.StatusInternalServerError)
						return
					}
				} else if err := dbpkg.SetConfigValue(dbSuper, chatFlotanteThemeKey, value, false); err != nil {
					http.Error(w, "No se pudo guardar la configuracion visual del chat", http.StatusInternalServerError)
					return
				}
			}
			if strings.TrimSpace(payload.TextSize) != "" {
				value := normalizeChatFlotanteTextSize(payload.TextSize)
				if empresaID > 0 {
					if err := setChatFlotanteEmpresaPref(dbEmp, empresaID, chatFlotanteTextSizeKey, value, usuario); err != nil {
						http.Error(w, "No se pudo guardar la configuracion visual del chat", http.StatusInternalServerError)
						return
					}
				} else if err := dbpkg.SetConfigValue(dbSuper, chatFlotanteTextSizeKey, value, false); err != nil {
					http.Error(w, "No se pudo guardar la configuracion visual del chat", http.StatusInternalServerError)
					return
				}
			}
			writeJSON(w, http.StatusOK, chatFlotantePrefsResponse(dbSuper, dbEmp, empresaID))
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}
