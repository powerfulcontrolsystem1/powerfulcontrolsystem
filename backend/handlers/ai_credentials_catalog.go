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
			ModelID:      "openai:gpt-5.4-mini",
			Provider:     "openai",
			DisplayName:  "OpenAI GPT-5.4 mini",
			ApiKeyEnv:    "OPENAI_API_KEY",
			ConfigKey:    "ai.model.openai.gpt_5_4_mini.api_key",
			FreePlanNote: "OpenAI: requiere API key y se guarda cifrada en configuracion avanzada.",
		},
		{
			ModelID:      "openai:gpt-5.5",
			Provider:     "openai",
			DisplayName:  "OpenAI GPT-5.5",
			ApiKeyEnv:    "OPENAI_API_KEY",
			ConfigKey:    "ai.model.openai.gpt_5_4_mini.api_key",
			FreePlanNote: "Comparte la credencial cifrada del proveedor OpenAI.",
		},
		{
			ModelID:      "openai:gpt-5.6-luna",
			Provider:     "openai",
			DisplayName:  "OpenAI GPT-5.6 Luna",
			ApiKeyEnv:    "OPENAI_API_KEY",
			ConfigKey:    "ai.model.openai.gpt_5_4_mini.api_key",
			FreePlanNote: "Comparte la credencial cifrada del proveedor OpenAI; validar disponibilidad antes de activarlo.",
		},
		{
			ModelID: "openai:gpt-5.6-terra", Provider: "openai", DisplayName: "OpenAI GPT-5.6 Terra", ApiKeyEnv: "OPENAI_API_KEY", ConfigKey: "ai.model.openai.gpt_5_4_mini.api_key",
			FreePlanNote: "Comparte la credencial cifrada del proveedor OpenAI; validar disponibilidad antes de activarlo.",
		},
		{
			ModelID: "openai:gpt-5.6-sol", Provider: "openai", DisplayName: "OpenAI GPT-5.6 Sol", ApiKeyEnv: "OPENAI_API_KEY", ConfigKey: "ai.model.openai.gpt_5_4_mini.api_key",
			FreePlanNote: "Comparte la credencial cifrada del proveedor OpenAI; validar disponibilidad antes de activarlo.",
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
	return "ai.provider." + v + ".api_key"
}
