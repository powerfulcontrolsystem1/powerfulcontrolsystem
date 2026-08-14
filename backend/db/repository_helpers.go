package db

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// repositoryCoreCode creates the stable alphanumeric code shared by the
// reservation and delivery repositories. Domain wrappers keep their public
// names while the formatting rule remains single-source.
func repositoryCoreCode(prefix string, parts ...string) string {
	var b strings.Builder
	for _, part := range parts {
		for _, r := range strings.ToUpper(strings.TrimSpace(part)) {
			if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				b.WriteRune(r)
				continue
			}
			if b.Len() > 0 && b.String()[b.Len()-1] != '-' {
				b.WriteRune('-')
			}
		}
		if b.Len() > 0 && b.String()[b.Len()-1] != '-' {
			b.WriteRune('-')
		}
	}
	code := strings.Trim(b.String(), "-")
	if code == "" {
		code = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if len(code) > 42 {
		code = code[:42]
	}
	return strings.Trim(strings.ToUpper(strings.TrimSpace(prefix)), "-") + "-" + strings.Trim(code, "-")
}

func normalizeRepositoryPeriod(value, fallback string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 7 {
		return value[:7]
	}
	if fallback != "" {
		return fallback
	}
	return value
}

func normalizeCurrentRepositoryPeriod(value string) string {
	return normalizeRepositoryPeriod(value, time.Now().Format("2006-01"))
}

func boundedPositiveInt(value, fallback, maximum int) int {
	if value <= 0 {
		return fallback
	}
	if maximum > 0 && value > maximum {
		return maximum
	}
	return value
}

func normalizeRepositoryCurrency(value, fallback string, maximum int) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return fallback
	}
	if maximum > 0 && len(value) > maximum {
		return value[:maximum]
	}
	return value
}

func repositoryISOWeekday(value time.Time) int {
	weekday := int(value.Weekday())
	if weekday == 0 {
		return 7
	}
	return weekday
}

func hashRepositoryKey(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func normalizeActiveArchiveState(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "inactivo" || value == "archivado" {
		return value
	}
	return "activo"
}

func normalizeRoundedPercentage(value float64) float64 {
	if value < 0 {
		value = 0
	}
	if value > 100 {
		value = 100
	}
	return round2(value)
}

func parseRepositoryStationID(externalReference, code string, companyID int64) int64 {
	reference := strings.ToUpper(strings.TrimSpace(externalReference))
	if strings.HasPrefix(reference, "ESTACION_") {
		if value, err := strconv.ParseInt(strings.TrimPrefix(reference, "ESTACION_"), 10, 64); err == nil && value > 0 {
			return value
		}
	}
	prefix := strings.ToUpper(fmt.Sprintf("EST-%d-", companyID))
	code = strings.ToUpper(strings.TrimSpace(code))
	if strings.HasPrefix(code, prefix) {
		if value, err := strconv.ParseInt(strings.TrimPrefix(code, prefix), 10, 64); err == nil && value > 0 {
			return value
		}
	}
	return 0
}

func normalizeRepositorySlug(value, fallback string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "-", "_")
	if value == "" {
		return fallback
	}
	return value
}
