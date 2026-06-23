package main

import "time"

const reportSchemaVersion = "riido-ai-agent-load-evidence.v1"

func (a *aggregator) report(cfg config, host string, startedAt, endedAt time.Time) report {
	success := successCount(a.status)
	total := len(a.all)
	endpoints := make(map[string]endpointSummary, len(a.endpoints))
	for path, ep := range a.endpoints {
		epSuccess := successCount(ep.status)
		epTotal := len(ep.latencies)
		endpoints[path] = endpointSummary{
			Total:        epTotal,
			Success:      epSuccess,
			Failures:     epTotal - epSuccess,
			StatusCounts: cloneStatusCounts(ep.status),
			ErrorCounts:  cloneStatusCounts(ep.errors),
			Latency:      summarizeLatency(ep.latencies),
		}
	}
	return report{
		SchemaVersion: reportSchemaVersion,
		Scenario:      cfg.Scenario,
		BaseHost:      host,
		WorkspaceID:   cfg.WorkspaceID,
		StartedAt:     startedAt,
		EndedAt:       endedAt,
		DurationMs:    endedAt.Sub(startedAt).Milliseconds(),
		Concurrency:   cfg.Concurrency,
		Total:         total,
		Success:       success,
		Failures:      total - success,
		StatusCounts:  cloneStatusCounts(a.status),
		ErrorCounts:   cloneStatusCounts(a.errors),
		Latency:       summarizeLatency(a.all),
		Endpoints:     endpoints,
	}
}

func cloneStatusCounts(source map[string]int) map[string]int {
	out := make(map[string]int, len(source))
	for key, value := range source {
		out[key] = value
	}
	return out
}
