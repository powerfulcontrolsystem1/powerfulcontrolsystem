package handlers

import (
	"net/http"
	"strings"
	"unicode"
)

// redactEmailForLog keeps authentication and operational logs useful without
// persisting a customer's or employee's full email address.
func redactEmailForLog(email string) string {
	email = normalizeLogValue(email)
	at := strings.LastIndex(email, "@")
	if at <= 0 || at == len(email)-1 {
		return "[redacted]"
	}

	local := []rune(email[:at])
	if len(local) == 1 {
		return string(local) + "***@" + email[at+1:]
	}
	return string(local[:1]) + "***@" + email[at+1:]
}

// redactPersonalDocumentForLog preserves only enough context to correlate a
// failure without writing a complete national ID or tax identifier to logs.
func redactPersonalDocumentForLog(document string) string {
	return redactIdentifierForLog(document, "doc")
}

func redactPhoneForLog(phone string) string {
	return redactIdentifierForLog(phone, "phone")
}

func redactTokenForLog(token string) string {
	if normalizeLogValue(token) == "" {
		return "[redacted-token]"
	}
	return "[redacted-token]"
}

func redactIdentifierForLog(value, kind string) string {
	value = normalizeLogValue(value)
	if value == "" {
		return "[redacted-" + kind + "]"
	}

	visible := make([]rune, 0, len(value))
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			visible = append(visible, r)
		}
	}
	if len(visible) <= 2 {
		return "[redacted-" + kind + "]"
	}
	return string(visible[:1]) + "***" + string(visible[len(visible)-2:])
}

// redactAuthorizationForLog never returns credential material. Callers should
// only use the boolean result to troubleshoot the presence of a header.
func redactAuthorizationForLog(value string) string {
	if normalizeLogValue(value) == "" {
		return "[authorization-absent]"
	}
	return "[authorization-present]"
}

func redactRequestHeadersForLog(headers http.Header) map[string]string {
	redacted := make(map[string]string)
	for key, values := range headers {
		normalizedKey := strings.ToLower(strings.TrimSpace(key))
		switch normalizedKey {
		case "authorization", "cookie", "set-cookie", "x-api-key", "x-auth-token":
			redacted[normalizedKey] = "[redacted]"
		default:
			redacted[normalizedKey] = normalizeLogValue(strings.Join(values, ","))
		}
	}
	return redacted
}

func normalizeLogValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.NewReplacer("\r", "\\r", "\n", "\\n").Replace(value)
	return value
}
