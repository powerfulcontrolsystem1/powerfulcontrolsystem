package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

const superConfigBackupVersion = "super-config-backup.v1"

func superConfigSensitiveSecretKeys() map[string]struct{} {
	keys := map[string]struct{}{
		"wompi.private_key":       {},
		"wompi.integrity_key":     {},
		"gmail.smtp_app_password": {},
	}
	for _, def := range aiCredentialCatalogModels() {
		if key := strings.TrimSpace(def.ConfigKey); key != "" {
			keys[key] = struct{}{}
		}
		if providerKey := strings.TrimSpace(aiProviderConfigKey(def.Provider)); providerKey != "" {
			keys[providerKey] = struct{}{}
		}
	}
	return keys
}

func superConfigKeyRequiresEncryption(key string) bool {
	key = strings.TrimSpace(key)
	if key == "" {
		return false
	}
	_, ok := superConfigSensitiveSecretKeys()[key]
	return ok
}

// EnsureSensitiveSuperConfigEncrypted normaliza credenciales secretas legacy para que queden cifradas.
func EnsureSensitiveSuperConfigEncrypted(dbSuper *sql.DB) error {
	sensitiveKeys := superConfigSensitiveSecretKeys()
	if len(sensitiveKeys) == 0 {
		return nil
	}

	for key := range sensitiveKeys {
		value, encrypted, _, _, err := dbpkg.GetConfigEntry(dbSuper, key)
		if err != nil {
			return fmt.Errorf("read config key %s: %w", key, err)
		}
		if strings.TrimSpace(value) == "" || encrypted {
			continue
		}

		if !utils.EncryptionAvailable() {
			return fmt.Errorf("sensitive key %s is plaintext and CONFIG_ENC_KEY is not available", key)
		}

		encVal, encErr := utils.EncryptString(value)
		if encErr != nil {
			return fmt.Errorf("encrypt sensitive key %s: %w", key, encErr)
		}
		if err := dbpkg.SetConfigValue(dbSuper, key, encVal, true); err != nil {
			return fmt.Errorf("persist encrypted sensitive key %s: %w", key, err)
		}
	}

	return nil
}

type superConfigBackupItem struct {
	Key        string `json:"key"`
	Value      string `json:"value"`
	Encrypted  bool   `json:"encrypted"`
	Configured bool   `json:"configured"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}

type superConfigBackupPayload struct {
	Version   string                  `json:"version"`
	Scope     string                  `json:"scope"`
	CreatedAt string                  `json:"created_at"`
	CreatedBy string                  `json:"created_by,omitempty"`
	Items     []superConfigBackupItem `json:"items"`
}

func superConfigCriticalKeys() []string {
	keys := []string{
		"wompi.public_key",
		"wompi.private_key",
		"wompi.integrity_key",
		"wompi.mode",
		"gmail.smtp_email",
		"gmail.smtp_app_password",
		"gmail.smtp_from_name",
		"gmail.smtp_host",
		"gmail.smtp_port",
		"gmail.confirm_base_url",
		"gmail.restart_alert_to",
		"gmail.restart_alert_enabled",
		"gmail.smtp_test_mode",
		"usuarios.password_min_length",
		"usuarios.password_require_uppercase",
		"usuarios.password_require_lowercase",
		"usuarios.password_require_digit",
		"usuarios.password_require_symbol",
		"usuarios.password_rotation_days",
	}

	for _, def := range aiCredentialCatalogModels() {
		if k := strings.TrimSpace(def.ConfigKey); k != "" {
			keys = append(keys, k)
			keys = append(keys, k+".updated_by")
		}
		if providerKey := strings.TrimSpace(aiProviderConfigKey(def.Provider)); providerKey != "" {
			keys = append(keys, providerKey)
		}
	}

	uniq := map[string]struct{}{}
	out := make([]string, 0, len(keys))
	for _, key := range keys {
		normalized := strings.TrimSpace(key)
		if normalized == "" {
			continue
		}
		if _, exists := uniq[normalized]; exists {
			continue
		}
		uniq[normalized] = struct{}{}
		out = append(out, normalized)
	}
	sort.Strings(out)
	return out
}

func buildSuperConfigBackupPayload(dbSuper *sql.DB, adminEmail string) (*superConfigBackupPayload, error) {
	items := make([]superConfigBackupItem, 0)
	for _, key := range superConfigCriticalKeys() {
		value, encrypted, _, updatedAt, err := dbpkg.GetConfigEntry(dbSuper, key)
		if err != nil {
			return nil, err
		}

		if superConfigKeyRequiresEncryption(key) && strings.TrimSpace(value) != "" && !encrypted {
			if !utils.EncryptionAvailable() {
				return nil, fmt.Errorf("sensitive key %s is plaintext and CONFIG_ENC_KEY is not available", key)
			}
			encVal, encErr := utils.EncryptString(value)
			if encErr != nil {
				return nil, fmt.Errorf("encrypt sensitive key %s for backup: %w", key, encErr)
			}
			if saveErr := dbpkg.SetConfigValue(dbSuper, key, encVal, true); saveErr != nil {
				return nil, fmt.Errorf("persist encrypted sensitive key %s for backup: %w", key, saveErr)
			}
			value = encVal
			encrypted = true
		}

		items = append(items, superConfigBackupItem{
			Key:        key,
			Value:      value,
			Encrypted:  encrypted,
			Configured: strings.TrimSpace(value) != "",
			UpdatedAt:  strings.TrimSpace(updatedAt),
		})
	}

	return &superConfigBackupPayload{
		Version:   superConfigBackupVersion,
		Scope:     "super_config_critica",
		CreatedAt: time.Now().Format(time.RFC3339),
		CreatedBy: strings.TrimSpace(adminEmail),
		Items:     items,
	}, nil
}

// SuperConfigBackupHandler exporta/restaura configuraciones críticas de super administrador.
func SuperConfigBackupHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			payload, err := buildSuperConfigBackupPayload(dbSuper, adminEmailFromRequest(r))
			if err != nil {
				log.Printf("GET /super/api/config/backup error: %v", err)
				http.Error(w, "No se pudo generar el respaldo", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Disposition", "attachment; filename=super_config_backup.json")
			if err := json.NewEncoder(w).Encode(payload); err != nil {
				log.Printf("GET /super/api/config/backup write error: %v", err)
			}
			return

		case http.MethodPost, http.MethodPut:
			var payload superConfigBackupPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON de respaldo invalido", http.StatusBadRequest)
				return
			}

			version := strings.TrimSpace(payload.Version)
			if version != "" && version != superConfigBackupVersion {
				http.Error(w, "version de respaldo no soportada", http.StatusBadRequest)
				return
			}

			allowed := map[string]struct{}{}
			for _, key := range superConfigCriticalKeys() {
				allowed[key] = struct{}{}
			}

			restored := make([]string, 0)
			skipped := make([]string, 0)
			for _, item := range payload.Items {
				key := strings.TrimSpace(item.Key)
				if key == "" {
					continue
				}
				if _, ok := allowed[key]; !ok {
					skipped = append(skipped, key)
					continue
				}
				if !item.Configured && strings.TrimSpace(item.Value) == "" {
					skipped = append(skipped, key)
					continue
				}

				valueToSave := item.Value
				encryptedToSave := item.Encrypted
				if superConfigKeyRequiresEncryption(key) && strings.TrimSpace(valueToSave) != "" {
					if !encryptedToSave {
						if !utils.EncryptionAvailable() {
							http.Error(w, "No se puede restaurar secreto en texto plano: CONFIG_ENC_KEY no disponible", http.StatusBadRequest)
							return
						}
						encVal, encErr := utils.EncryptString(valueToSave)
						if encErr != nil {
							http.Error(w, "No se pudo cifrar la clave "+key+": "+encErr.Error(), http.StatusInternalServerError)
							return
						}
						valueToSave = encVal
						encryptedToSave = true
					}
				}

				if err := dbpkg.SetConfigValue(dbSuper, key, valueToSave, encryptedToSave); err != nil {
					log.Printf("PUT /super/api/config/backup restore key=%s error: %v", key, err)
					http.Error(w, "No se pudo restaurar la clave "+key, http.StatusInternalServerError)
					return
				}
				restored = append(restored, key)
			}

			if len(restored) == 0 {
				http.Error(w, "No se encontraron claves configuradas para restaurar", http.StatusBadRequest)
				return
			}

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":             true,
				"restored_count": len(restored),
				"restored_keys":  restored,
				"skipped_keys":   skipped,
			})
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}
