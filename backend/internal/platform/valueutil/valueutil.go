package valueutil

import (
	"encoding/hex"
	"encoding/json"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"
)

type orderedNumber interface {
	~int | ~int64 | ~float64
}

// FirstNonBlank devuelve el primer texto no vacio ya normalizado.
func FirstNonBlank(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// NonBlankOr aplica un valor predeterminado sin conservar espacios laterales.
func NonBlankOr(value, fallback string) string {
	if value = strings.TrimSpace(value); value != "" {
		return value
	}
	return fallback
}

// FirstForwardedValue extrae el primer valor de un encabezado reenviado.
func FirstForwardedValue(raw string) string {
	parts := strings.Split(strings.TrimSpace(raw), ",")
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}

// HostWithoutPort normaliza hosts IPv4, IPv6 y nombres DNS con puerto opcional.
func HostWithoutPort(rawHost string) string {
	trimmed := strings.TrimSpace(rawHost)
	if trimmed == "" {
		return ""
	}
	hostOnly, _, err := net.SplitHostPort(trimmed)
	if err == nil {
		return strings.TrimSpace(hostOnly)
	}
	if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
		return strings.Trim(strings.TrimSpace(trimmed), "[]")
	}
	return trimmed
}

// IsSafeSQLIdentifier acepta solo identificadores simples, nunca expresiones SQL.
func IsSafeSQLIdentifier(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '_' {
			continue
		}
		return false
	}
	return true
}

// ParsePositiveInt64 convierte un entero estrictamente positivo o devuelve cero.
func ParsePositiveInt64(raw string) int64 {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return 0
	}
	return parsed
}

// Truncate limita texto por runas para no cortar caracteres Unicode.
func Truncate(value string, maximum int) string {
	if maximum <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= maximum {
		return string(runes)
	}
	return string(runes[:maximum])
}

// TrimmedPrefix limits an ASCII-oriented protocol value after removing outer
// whitespace. It is used for fixed-width ISO date/time fields, not free-form
// Unicode text (use Truncate for the latter).
func TrimmedPrefix(value string, maximum int) string {
	value = strings.TrimSpace(value)
	if maximum <= 0 {
		return ""
	}
	if len(value) <= maximum {
		return value
	}
	return value[:maximum]
}

// UniqueSortedNonBlank normalizes a validation-message collection into a
// deterministic set.
func UniqueSortedNonBlank(values []string) []string {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			set[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

// IsHexLength verifies an exact-length hexadecimal identifier such as a
// SHA-384 digest represented by 96 characters.
func IsHexLength(value string, length int) bool {
	value = strings.TrimSpace(value)
	if length <= 0 || len(value) != length {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

// ParseDateTimeLocal reconoce los formatos historicos de fecha del sistema.
func ParseDateTimeLocal(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{"2006-01-02 15:04:05", time.RFC3339, "2006-01-02T15:04:05"} {
		if parsed, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

// NormalizeAllowed devuelve el valor normalizado solo si pertenece al catalogo.
func NormalizeAllowed(raw string, allowed ...string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	for _, candidate := range allowed {
		if value == candidate {
			return value
		}
	}
	return ""
}

// MarshalJSONOr serializa valores internos con una respuesta segura de respaldo.
func MarshalJSONOr(value interface{}, fallback string) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return fallback
	}
	return string(raw)
}

// Clamp conserva un numero dentro del intervalo cerrado indicado.
func Clamp[T orderedNumber](value, minimum, maximum T) T {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

// FirstPositive devuelve el primer numero estrictamente positivo.
func FirstPositive[T orderedNumber](values ...T) T {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	var zero T
	return zero
}
