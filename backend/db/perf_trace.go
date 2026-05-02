package db

import (
	"log"
	"os"
	"strings"
)

func PerfTraceEnabled() bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv("PCS_PERF_TRACE")))
	return raw == "1" || raw == "true" || raw == "yes" || raw == "on"
}

func PerfLogf(format string, args ...interface{}) {
	if !PerfTraceEnabled() {
		return
	}
	log.Printf(format, args...)
}
