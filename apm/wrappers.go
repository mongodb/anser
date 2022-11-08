package apm

import (
	"context"
	"time"

	"github.com/mongodb/grip"
	"github.com/mongodb/grip/recovery"
)

type loggingMonitor struct {
	interval time.Duration
	Monitor
}

// NewLoggingMonitor wraps an existing logging monitor that
// automatically logs the output the default grip logger on the
// specified interval.
func NewLoggingMonitor(ctx context.Context, dur time.Duration, m Monitor) Monitor {
	impl := &loggingMonitor{
		interval: dur,
		Monitor:  m,
	}
	go impl.flusher(ctx)
	return impl
}

func (m *loggingMonitor) flusher(ctx context.Context) {
	defer recovery.LogStackTraceAndContinue("logging driver apm collector")
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			grip.Info(m.Monitor.Rotate().Message())
		}
	}
}
