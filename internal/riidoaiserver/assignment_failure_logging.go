package riidoaiserver

import (
	"context"
	"log"
	"strings"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func logAssignmentFailure(state *storeState, event TaskEvent, contexts ...context.Context) {
	if event.Type != EventAssignmentFailed {
		return
	}
	provider := "unknown"
	if assignment, ok := state.assignments[event.AssignmentID]; ok {
		provider = boundedProvider(assignment.RuntimeProvider)
	}
	ctx := context.Background()
	if len(contexts) > 0 {
		ctx = contexts[0]
	}
	traceID, spanID := traceCorrelationFromContext(ctx)
	log.Printf(
		"event=assignment_failed provider=%q result_status=%q failure_category=%q assignment_ref=%q trace_id=%q span_id=%q",
		provider,
		boundedResultStatus(event.Metadata[metadatakeys.AssignmentResultStatus.String()]),
		boundedFailureCategory(event.Metadata[metadatakeys.AssignmentFailureCategory.String()]),
		providerHealthDeviceRef(event.AssignmentID), traceID, spanID,
	)
}

func boundedProvider(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "codex", "claude", "cursor", "openclaw":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "unknown"
	}
}

func boundedResultStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "failed", "blocked", "timeout", "aborted", "cancelled", "needs_input":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "unknown"
	}
}

func boundedFailureCategory(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "provider_limit", "provider_blocked", "process_aborted", "provider_timeout", "provider_result_failed",
		"stale_replay_active_assignment", "stale_replay_queued_assignment", "stale_queued_assignment":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "unknown"
	}
}
