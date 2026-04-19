package handlers

import (
	"database/sql"
	"strings"
)

func aiProviderEnabledConfigKey(provider string) string {
	v := strings.ToLower(strings.TrimSpace(provider))
	if v == "" {
		return ""
	}
	return "ai.provider." + v + ".enabled"
}

func isAIProviderEnabled(dbSuper *sql.DB, provider string) bool {
	key := aiProviderEnabledConfigKey(provider)
	if key == "" || dbSuper == nil {
		return true
	}
	value, err := getDecryptedConfigValue(dbSuper, key)
	if err != nil {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "0", "false", "off", "inactivo", "disabled":
		return false
	default:
		return true
	}
}

func uniqueAIProviders() []string {
	seen := map[string]struct{}{}
	providers := make([]string, 0)
	for _, def := range aiCredentialCatalogModels() {
		provider := strings.ToLower(strings.TrimSpace(def.Provider))
		if provider == "" {
			continue
		}
		if _, ok := seen[provider]; ok {
			continue
		}
		seen[provider] = struct{}{}
		providers = append(providers, provider)
	}
	return providers
}

type aiCredentialModelDef struct {
	ModelID      string
	Provider     string
	DisplayName  string
	ApiKeyEnv    string
	ConfigKey    string
	FreePlanNote string
}

func aiCredentialCatalogModels() []aiCredentialModelDef {
	return []aiCredentialModelDef{
		{
			ModelID:      "deepseek:deepseek-chat",
			Provider:     "deepseek",
			DisplayName:  "DeepSeek Chat",
			ApiKeyEnv:    "DEEPSEEK_API_KEY",
			ConfigKey:    "ai.model.deepseek.deepseek_chat.api_key",
			FreePlanNote: "DeepSeek: plan gratuito sujeto a limites y politicas de uso del proveedor.",
		},
		{
			ModelID:      "ollama:ambis",
			Provider:     "ollama",
			DisplayName:  "Ambis Local",
			ApiKeyEnv:    "",
			ConfigKey:    "",
			FreePlanNote: "Ambis Local: usa Ollama en el VPS por loopback, sin exponer credenciales ni acceso publico.",
		},
	}
}

func aiCredentialByModelID() map[string]aiCredentialModelDef {
	defs := aiCredentialCatalogModels()
	out := make(map[string]aiCredentialModelDef, len(defs))
	for _, it := range defs {
		out[it.ModelID] = it
	}
	return out
}

func aiProviderConfigKey(provider string) string {
	v := strings.ToLower(strings.TrimSpace(provider))
	if v == "" {
		return ""
	}
	if v == "ollama" {
		return ""
	}
	return "ai.provider." + v + ".api_key"
}
