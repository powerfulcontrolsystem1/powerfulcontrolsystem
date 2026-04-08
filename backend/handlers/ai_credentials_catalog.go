package handlers

import "strings"

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
