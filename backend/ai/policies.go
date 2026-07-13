package ai

import (
	"regexp"
	"strings"
)

// DataCategory is recorded as metadata only. Raw confidential values must not
// be copied to provider prompts or audit records.
type DataCategory string

const (
	DataPublic               DataCategory = "public"
	DataInternal             DataCategory = "internal"
	DataBusinessConfidential DataCategory = "business_confidential"
	DataPersonal             DataCategory = "personal_data"
	DataFinancial            DataCategory = "financial_data"
	DataTax                  DataCategory = "tax_data"
	DataCredential           DataCategory = "credential"
	DataHighlySensitive      DataCategory = "highly_sensitive"
)

var injectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)ignora\s+(las\s+)?instrucciones`),
	regexp.MustCompile(`(?i)(revela|muestra|dame)\s+(la\s+)?(clave|api[ _-]?key|token|secreto)`),
	regexp.MustCompile(`(?i)(omite|omitir|salta|saltar)\s+(la\s+)?confirmaci[oó]n`),
	regexp.MustCompile(`(?i)(usa|cambia)\s+(la\s+)?empresa\s+\d+`),
	regexp.MustCompile(`(?i)cambia\s+todos\s+los\s+precios`),
}

// IsUntrustedInstruction recognizes common prompt-injection language in data
// retrieved from documents, integrations, products or user attachments. It is
// a defense-in-depth signal; tools remain controlled by server policies.
func IsUntrustedInstruction(value string) bool {
	for _, pattern := range injectionPatterns {
		if pattern.MatchString(value) {
			return true
		}
	}
	return false
}

// ProviderSafeFields retains only explicit allowed keys and drops data that
// may identify credentials, persons or complete fiscal/banking records.
func ProviderSafeFields(values map[string]string, allowed []string) map[string]string {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, key := range allowed {
		allowedSet[strings.ToLower(strings.TrimSpace(key))] = struct{}{}
	}
	out := make(map[string]string)
	for key, value := range values {
		cleanKey := strings.ToLower(strings.TrimSpace(key))
		if _, ok := allowedSet[cleanKey]; !ok || IsSensitiveField(cleanKey) {
			continue
		}
		out[key] = strings.TrimSpace(value)
	}
	return out
}

func IsSensitiveField(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	for _, part := range []string{"password", "secret", "token", "key", "certificate", "cookie", "cvv", "account_number", "bank", "totp"} {
		if strings.Contains(key, part) {
			return true
		}
	}
	return false
}
