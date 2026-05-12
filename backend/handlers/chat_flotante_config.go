package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	chatFlotanteChatEnabledKey        = "chat_flotante.chat_enabled"
	chatFlotanteRobotEnabledKey       = "chat_flotante.robot_enabled"
	chatFlotanteVoiceEnabledKey       = "chat_flotante.voice_enabled"
	chatFlotanteRobotVoiceKey         = "chat_flotante.robot_voice"
	chatFlotantePersonalityModeKey    = "chat_flotante.personality_mode"
	chatFlotanteRadioOnlineEnabledKey = "chat_flotante.radio_online_enabled"
	chatFlotantePersonalityNormal     = "normal"
	chatFlotantePersonalityRobot      = "robot"
	chatFlotantePersonalitySecretary  = "secretary"
)

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
	return getChatFlotanteEnabledDefaultTrue(dbSuper, chatFlotanteRobotEnabledKey)
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
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case chatFlotantePersonalityRobot, "ejecutivo", "avatar", "3d":
		return chatFlotantePersonalityRobot
	case chatFlotantePersonalitySecretary, "secretaria", "asistente-secretaria", "recepcionista":
		return chatFlotantePersonalitySecretary
	default:
		return chatFlotantePersonalityNormal
	}
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
	personalityMode := normalizeChatFlotantePersonalityMode(getChatFlotanteStringForEmpresa(dbSuper, dbEmp, empresaID, chatFlotantePersonalityModeKey, chatFlotantePersonalityNormal))
	return map[string]any{
		"ok":         true,
		"empresa_id": empresaID,
		"scope": func() string {
			if empresaID > 0 {
				return "empresa"
			}
			return "global"
		}(),
		"chat_enabled":         getChatFlotanteBoolForEmpresa(dbSuper, dbEmp, empresaID, chatFlotanteChatEnabledKey, true),
		"robot_enabled":        getChatFlotanteBoolForEmpresa(dbSuper, dbEmp, empresaID, chatFlotanteRobotEnabledKey, true),
		"radio_online_enabled": getChatFlotanteBoolForEmpresa(dbSuper, dbEmp, empresaID, chatFlotanteRadioOnlineEnabledKey, true),
		"voice_enabled":        getChatFlotanteBoolForEmpresa(dbSuper, dbEmp, empresaID, chatFlotanteVoiceEnabledKey, false),
		"robot_voice":          robotVoice,
		"personality_mode":     personalityMode,
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
			writeJSON(w, http.StatusOK, chatFlotantePrefsResponse(dbSuper, dbEmp, empresaID))
			return
		case http.MethodPost, http.MethodPut:
			var payload struct {
				EmpresaID          int64  `json:"empresa_id"`
				ChatEnabled        *bool  `json:"chat_enabled"`
				RobotEnabled       *bool  `json:"robot_enabled"`
				RadioOnlineEnabled *bool  `json:"radio_online_enabled"`
				VoiceEnabled       *bool  `json:"voice_enabled"`
				RobotVoice         string `json:"robot_voice"`
				PersonalityMode    string `json:"personality_mode"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido: "+err.Error(), http.StatusBadRequest)
				return
			}
			if empresaID <= 0 && payload.EmpresaID > 0 {
				empresaID = payload.EmpresaID
			}
			usuario := adminEmailFromRequest(r)
			if payload.ChatEnabled != nil {
				value := chatFlotanteBoolValue(*payload.ChatEnabled)
				if empresaID > 0 {
					if err := setChatFlotanteEmpresaPref(dbEmp, empresaID, chatFlotanteChatEnabledKey, value, usuario); err != nil {
						http.Error(w, "No se pudo guardar chat_flotante.chat_enabled por empresa: "+err.Error(), http.StatusInternalServerError)
						return
					}
				} else if err := dbpkg.SetConfigValue(dbSuper, chatFlotanteChatEnabledKey, value, false); err != nil {
					http.Error(w, "No se pudo guardar chat_flotante.chat_enabled: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if payload.RobotEnabled != nil {
				value := chatFlotanteBoolValue(*payload.RobotEnabled)
				if empresaID > 0 {
					if err := setChatFlotanteEmpresaPref(dbEmp, empresaID, chatFlotanteRobotEnabledKey, value, usuario); err != nil {
						http.Error(w, "No se pudo guardar chat_flotante.robot_enabled por empresa: "+err.Error(), http.StatusInternalServerError)
						return
					}
				} else if err := dbpkg.SetConfigValue(dbSuper, chatFlotanteRobotEnabledKey, value, false); err != nil {
					http.Error(w, "No se pudo guardar chat_flotante.robot_enabled: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if payload.RadioOnlineEnabled != nil {
				value := chatFlotanteBoolValue(*payload.RadioOnlineEnabled)
				if empresaID > 0 {
					if err := setChatFlotanteEmpresaPref(dbEmp, empresaID, chatFlotanteRadioOnlineEnabledKey, value, usuario); err != nil {
						http.Error(w, "No se pudo guardar chat_flotante.radio_online_enabled por empresa: "+err.Error(), http.StatusInternalServerError)
						return
					}
				} else if err := dbpkg.SetConfigValue(dbSuper, chatFlotanteRadioOnlineEnabledKey, value, false); err != nil {
					http.Error(w, "No se pudo guardar chat_flotante.radio_online_enabled: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if payload.VoiceEnabled != nil {
				value := chatFlotanteBoolValue(*payload.VoiceEnabled)
				if empresaID > 0 {
					if err := setChatFlotanteEmpresaPref(dbEmp, empresaID, chatFlotanteVoiceEnabledKey, value, usuario); err != nil {
						http.Error(w, "No se pudo guardar chat_flotante.voice_enabled por empresa: "+err.Error(), http.StatusInternalServerError)
						return
					}
				} else if err := dbpkg.SetConfigValue(dbSuper, chatFlotanteVoiceEnabledKey, value, false); err != nil {
					http.Error(w, "No se pudo guardar chat_flotante.voice_enabled: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if strings.TrimSpace(payload.PersonalityMode) != "" {
				value := normalizeChatFlotantePersonalityMode(payload.PersonalityMode)
				if empresaID > 0 {
					if err := setChatFlotanteEmpresaPref(dbEmp, empresaID, chatFlotantePersonalityModeKey, value, usuario); err != nil {
						http.Error(w, "No se pudo guardar chat_flotante.personality_mode por empresa: "+err.Error(), http.StatusInternalServerError)
						return
					}
				} else if err := dbpkg.SetConfigValue(dbSuper, chatFlotantePersonalityModeKey, value, false); err != nil {
					http.Error(w, "No se pudo guardar chat_flotante.personality_mode: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if strings.TrimSpace(payload.RobotVoice) != "" {
				value := normalizeChatFlotanteRobotVoice(payload.RobotVoice)
				if empresaID > 0 {
					if err := setChatFlotanteEmpresaPref(dbEmp, empresaID, chatFlotanteRobotVoiceKey, value, usuario); err != nil {
						http.Error(w, "No se pudo guardar chat_flotante.robot_voice por empresa: "+err.Error(), http.StatusInternalServerError)
						return
					}
				} else if err := dbpkg.SetConfigValue(dbSuper, chatFlotanteRobotVoiceKey, value, false); err != nil {
					http.Error(w, "No se pudo guardar chat_flotante.robot_voice: "+err.Error(), http.StatusInternalServerError)
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
