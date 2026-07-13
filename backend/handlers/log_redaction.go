package handlers

import "strings"

// redactEmailForLog keeps authentication and operational logs useful without
// persisting a customer's or employee's full email address.
func redactEmailForLog(email string) string {
	email = strings.TrimSpace(email)
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
