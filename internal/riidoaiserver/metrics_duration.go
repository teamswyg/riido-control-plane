package riidoaiserver

import (
	"strings"
	"time"
)

func durationMilliseconds(duration time.Duration) int64 {
	if duration <= 0 {
		return 0
	}
	elapsedMS := duration.Milliseconds()
	if elapsedMS == 0 {
		return 1
	}
	return elapsedMS
}

func normalizeHTTPMetricValue(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func (m *httpTransactionLatencyMetrics) observe(elapsedMS int64) {
	m.samplesTotal++
	m.totalMilliseconds += elapsedMS
	if elapsedMS > m.maxMilliseconds {
		m.maxMilliseconds = elapsedMS
	}
	m.lastMilliseconds = elapsedMS
}
