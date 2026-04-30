package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	chatFlotanteChatEnabledKey       = "chat_flotante.chat_enabled"
	chatFlotanteRobotEnabledKey      = "chat_flotante.robot_enabled"
	chatFlotanteVoiceEnabledKey      = "chat_flotante.voice_enabled"
	chatFlotanteRobotVoiceKey        = "chat_flotante.robot_voice"
	chatFlotantePersonalityModeKey   = "chat_flotante.personality_mode"
	chatFlotantePersonalityNormal    = "normal"
	chatFlotantePersonalityRobot     = "robot"
	chatFlotantePersonalitySecretary = "secretary"
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

// ChatFlotantePreferenciasHandler expone preferencias no sensibles del chat IA flotante.
func ChatFlotantePreferenciasHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !voiceStreamRequireSession(w, r, dbSuper) {
			return
		}

		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, map[string]any{
				"ok":               true,
				"chat_enabled":     getChatFlotanteChatEnabled(dbSuper),
				"robot_enabled":    getChatFlotanteRobotEnabled(dbSuper),
				"voice_enabled":    getChatFlotanteVoiceEnabled(dbSuper),
				"robot_voice":      getChatFlotanteRobotVoice(dbSuper),
				"personality_mode": getChatFlotantePersonalityMode(dbSuper),
			})
			return
		case http.MethodPost, http.MethodPut:
			var payload struct {
				ChatEnabled     *bool  `json:"chat_enabled"`
				RobotEnabled    *bool  `json:"robot_enabled"`
				VoiceEnabled    *bool  `json:"voice_enabled"`
				RobotVoice      string `json:"robot_voice"`
				PersonalityMode string `json:"personality_mode"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido: "+err.Error(), http.StatusBadRequest)
				return
			}
			if payload.ChatEnabled != nil {
				value := "0"
				if *payload.ChatEnabled {
					value = "1"
				}
				if err := dbpkg.SetConfigValue(dbSuper, chatFlotanteChatEnabledKey, value, false); err != nil {
					http.Error(w, "No se pudo guardar chat_flotante.chat_enabled: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if payload.RobotEnabled != nil {
				value := "0"
				if *payload.RobotEnabled {
					value = "1"
				}
				if err := dbpkg.SetConfigValue(dbSuper, chatFlotanteRobotEnabledKey, value, false); err != nil {
					http.Error(w, "No se pudo guardar chat_flotante.robot_enabled: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if payload.VoiceEnabled != nil {
				value := "0"
				if *payload.VoiceEnabled {
					value = "1"
				}
				if err := dbpkg.SetConfigValue(dbSuper, chatFlotanteVoiceEnabledKey, value, false); err != nil {
					http.Error(w, "No se pudo guardar chat_flotante.voice_enabled: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if strings.TrimSpace(payload.PersonalityMode) != "" {
				value := normalizeChatFlotantePersonalityMode(payload.PersonalityMode)
				if err := dbpkg.SetConfigValue(dbSuper, chatFlotantePersonalityModeKey, value, false); err != nil {
					http.Error(w, "No se pudo guardar chat_flotante.personality_mode: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if strings.TrimSpace(payload.RobotVoice) != "" {
				value := normalizeChatFlotanteRobotVoice(payload.RobotVoice)
				if err := dbpkg.SetConfigValue(dbSuper, chatFlotanteRobotVoiceKey, value, false); err != nil {
					http.Error(w, "No se pudo guardar chat_flotante.robot_voice: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"ok":               true,
				"chat_enabled":     getChatFlotanteChatEnabled(dbSuper),
				"robot_enabled":    getChatFlotanteRobotEnabled(dbSuper),
				"voice_enabled":    getChatFlotanteVoiceEnabled(dbSuper),
				"robot_voice":      getChatFlotanteRobotVoice(dbSuper),
				"personality_mode": getChatFlotantePersonalityMode(dbSuper),
			})
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}
