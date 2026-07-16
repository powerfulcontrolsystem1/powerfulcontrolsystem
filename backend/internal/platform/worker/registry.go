package worker

import (
	"context"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const JobKindPlatformNoop = "platform.noop"

// DefaultRegistry is deliberately small until each business family has been
// moved from API timers to durable handlers. It makes an empty registry a
// startup failure and offers one harmless diagnostic job for staging probes.
func DefaultRegistry() map[string]HandlerSpec {
	return map[string]HandlerSpec{
		JobKindPlatformNoop: {
			Kind:        JobKindPlatformNoop,
			Version:     1,
			Timeout:     15 * time.Second,
			MaxAttempts: 1,
			Enabled:     true,
			Handle: func(context.Context, dbpkg.AsyncJob) error {
				return nil
			},
		},
	}
}

func Kinds(registry map[string]HandlerSpec) map[string]struct{} {
	kinds := make(map[string]struct{}, len(registry))
	for kind := range registry {
		kinds[kind] = struct{}{}
	}
	return kinds
}
