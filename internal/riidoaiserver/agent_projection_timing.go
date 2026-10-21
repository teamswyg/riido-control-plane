package riidoaiserver

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/metadatakeys"
)

type agentProjectionTiming struct {
	started                        time.Time
	phase                          string
	lock, reload, projection, save int64
}

func newAgentProjectionTiming(req AgentEventRequest, event TaskEvent) *agentProjectionTiming {
	phase := ""
	switch {
	case assignment.RecoveryContextFromMetadata(event.Metadata) != nil:
		phase = "recovery_only"
	case req.Metadata[metadatakeys.ProgressMessageKey.String()] == "assistant.partial":
		phase = "partial"
	case isTerminal(event.State):
		phase = "terminal"
	}
	return &agentProjectionTiming{started: time.Now(), phase: phase, lock: -1, reload: -1, projection: -1, save: -1}
}

func measureProjectionPhase(value *int64, run func() error) error {
	started := time.Now()
	err := run()
	*value = time.Since(started).Milliseconds()
	return err
}

func (t *agentProjectionTiming) fields(ctx context.Context, failed bool) map[string]any {
	outcome := "saved"
	if t.phase == "recovery_only" {
		outcome = "skipped"
	}
	if failed {
		outcome = "failed"
	}
	traceID, spanID := traceCorrelationFromContext(ctx)
	return map[string]any{
		"event": "agent_projection_timing", "phase": t.phase, "outcome": outcome,
		"projection_total_ms":     time.Since(t.started).Milliseconds(),
		"projection_lock_wait_ms": t.lock, "projection_reload_ms": t.reload,
		"projection_apply_ms": t.projection, "projection_save_ms": t.save,
		"trace_id": traceID, "span_id": spanID,
	}
}

func (t *agentProjectionTiming) finish(ctx context.Context, failed bool) {
	if t.phase == "" {
		return
	}
	if body, err := json.Marshal(t.fields(ctx, failed)); err == nil {
		log.Print(string(body))
	}
}
