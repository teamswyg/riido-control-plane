package main

import "time"

const reportSchemaVersion = "riido-ai-agent-load-evidence.v1"

func (a *aggregator) report(cfg config, host string, startedAt, endedAt time.Time, resources resourceDelta) report {
	success := successCount(a.status)
	total := len(a.all)
	duration := endedAt.Sub(startedAt)
	endpoints := make(map[string]endpointSummary, len(a.endpoints))
	for path, ep := range a.endpoints {
		epSuccess := successCount(ep.status)
		epTotal := len(ep.latencies)
		epFailures := epTotal - epSuccess
		endpoints[path] = endpointSummary{
			Total:          epTotal,
			Success:        epSuccess,
			Failures:       epFailures,
			RequestsPerSec: perSecond(epTotal, duration),
			FailureRatePct: failureRatePct(epTotal, epFailures),
			StatusCounts:   cloneStatusCounts(ep.status),
			ErrorCounts:    cloneStatusCounts(ep.errors),
			Latency:        summarizeLatency(ep.latencies),
		}
	}
	r := report{
		SchemaVersion:  reportSchemaVersion,
		Scenario:       cfg.Scenario,
		BaseHost:       host,
		WorkspaceID:    cfg.WorkspaceID,
		StartedAt:      startedAt,
		EndedAt:        endedAt,
		DurationMs:     duration.Milliseconds(),
		Concurrency:    cfg.Concurrency,
		Total:          total,
		Success:        success,
		Failures:       total - success,
		RequestsPerSec: perSecond(total, duration),
		SuccessPerSec:  perSecond(success, duration),
		FailureRatePct: failureRatePct(total, total-success),
		StatusCounts:   cloneStatusCounts(a.status),
		ErrorCounts:    cloneStatusCounts(a.errors),
		Latency:        summarizeLatency(a.all),
		Endpoints:      endpoints,
		Resource:       resources,
	}
	r.Capacity = capacityFromReport(r)
	r.Findings = loadFindings(endpoints, resources)
	return r
}

func cloneStatusCounts(source map[string]int) map[string]int {
	out := make(map[string]int, len(source))
	for key, value := range source {
		out[key] = value
	}
	return out
}
