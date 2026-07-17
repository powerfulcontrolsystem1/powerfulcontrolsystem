package worker

import (
	"context"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const JobKindPlatformNoop = "platform.noop"

// DefaultRegistry is reserved for isolated runner diagnostics. The production
// pcs-worker builds its registry explicitly and never registers this no-op job.
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
