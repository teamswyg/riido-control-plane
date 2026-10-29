package riidoaiserver

import "time"

// Optional observation metadata must never prevent a liveness update.
func normalizeVersionObservedAt(value, version string, now time.Time) time.Time {
	if version == "" {
		return time.Time{}
	}
	observed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil || observed.IsZero() || observed.After(now.Add(5*time.Minute)) {
		return time.Time{}
	}
	return observed.UTC()
}
