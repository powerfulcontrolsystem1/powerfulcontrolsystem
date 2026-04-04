package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

// AIModelsConfigHandler gestiona credenciales IA disponibles en el catalogo backend.
func AIModelsConfigHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			adminEmail := strings.TrimSpace(adminEmailFromRequest(r))
			defs := aiCredentialCatalogModels()
			items := make([]map[string]interface{}, 0, len(defs))
			for _, def := range defs {
				value, enc, _, updatedAt, _ := dbpkg.GetConfigEntry(dbSuper, def.ConfigKey)
				updatedBy, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, def.ConfigKey+".updated_by")
				configured := strings.TrimSpace(value) != ""
				masked := maskSecretValue(value, enc)
				source := "db"
				if !configured {
					envVal := strings.TrimSpace(os.Getenv(def.ApiKeyEnv))
					if envVal != "" {
						configured = true
						masked = maskSecretValue(envVal, false)
						source = "env"
					}
				}
				items = append(items, map[string]interface{}{
					"model_id":         def.ModelID,
					"provider":         def.Provider,
					"display_name":     def.DisplayName,
					"config_key":       def.ConfigKey,
					"api_key_env":      def.ApiKeyEnv,
					"free_plan_note":   def.FreePlanNote,
					"configured":       configured,
					"masked":           masked,
					"source":           source,
					"updated_at":       updatedAt,
					"updated_by":       strings.TrimSpace(updatedBy),
					"google_account":   adminEmail,
					"encryption_state": enc,
				})
			}

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":                   true,
				"google_account":       adminEmail,
				"encryption_available": utils.EncryptionAvailable(),
				"modelos":              items,
			})
			return

		case http.MethodPut, http.MethodPost:
			var payload struct {
				Credentials []struct {
					ModelID string `json:"model_id"`
					APIKey  string `json:"api_key"`
				} `json:"credentials"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
				return
			}

			// Forzar cifrado obligatorio para todas las credenciales sensibles en configuración avanzada.
			if !utils.EncryptionAvailable() {
				http.Error(w, "encryption required: CONFIG_ENC_KEY not set", http.StatusBadRequest)
				return
			}

			adminEmail := strings.TrimSpace(adminEmailFromRequest(r))
			if adminEmail == "" || adminEmail == "sistema" {
				http.Error(w, "unauthenticated", http.StatusUnauthorized)
				return
			}

			defs := aiCredentialByModelID()
			updated := make([]string, 0, len(payload.Credentials))
			for _, item := range payload.Credentials {
				modelID := strings.TrimSpace(item.ModelID)
				apiKey := strings.TrimSpace(item.APIKey)
				if modelID == "" || apiKey == "" {
					continue
				}
				def, ok := defs[modelID]
				if !ok {
					http.Error(w, "model_id no soportado: "+modelID, http.StatusBadRequest)
					return
				}

				// Siempre cifrar el valor antes de persistir
				encVal, err := utils.EncryptString(apiKey)
				if err != nil {
					http.Error(w, "encryption failed: "+err.Error(), http.StatusInternalServerError)
					return
				}

				if err := dbpkg.SetConfigValue(dbSuper, def.ConfigKey, encVal, true); err != nil {
					http.Error(w, "failed to save "+def.ConfigKey+": "+err.Error(), http.StatusInternalServerError)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, def.ConfigKey+".updated_by", adminEmail, false); err != nil {
					http.Error(w, "failed to save updated_by for "+def.ConfigKey+": "+err.Error(), http.StatusInternalServerError)
					return
				}

				providerKey := aiProviderConfigKey(def.Provider)
				if providerKey != "" {
					if err := dbpkg.SetConfigValue(dbSuper, providerKey, encVal, true); err != nil {
						http.Error(w, "failed to save provider key "+providerKey+": "+err.Error(), http.StatusInternalServerError)
						return
					}
				}

				updated = append(updated, modelID)
			}

			if len(updated) == 0 {
				http.Error(w, "debe enviar al menos una credencial valida", http.StatusBadRequest)
				return
			}

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"saved":          true,
				"updated_models": updated,
				"google_account": adminEmail,
			})
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

func readDecryptedConfigOrEmpty(dbSuper *sql.DB, key string) (string, error) {
	v, enc, err := dbpkg.GetConfigValue(dbSuper, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
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

func maskSecretValue(raw string, encrypted bool) string {
	v := strings.TrimSpace(raw)
	if v == "" {
		return ""
	}
	if encrypted {
		return "********"
	}
	if len(v) <= 8 {
		return "****"
	}
	return v[:2] + "****" + v[len(v)-2:]
}
