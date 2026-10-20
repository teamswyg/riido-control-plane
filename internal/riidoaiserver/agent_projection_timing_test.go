package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestProjectionTimingUnknownPhasesAndPrivacy(t *testing.T) {
	req := AgentEventRequest{Metadata: map[string]string{"private": "credential-secret"}, Message: "answer-secret"}
	timing := newAgentProjectionTiming(req, recoveryProjectionEvent())
	fields := timing.fields(context.Background(), false)
	for _, key := range []string{"projection_lock_wait_ms", "projection_reload_ms", "projection_apply_ms", "projection_save_ms"} {
		if fields[key] != int64(-1) {
			t.Fatalf("unattempted %s=%v", key, fields[key])
		}
	}
	if fields["outcome"] != "skipped" || fields["phase"] != "recovery_only" {
		t.Fatal("recovery no-op reported as a saved projection")
	}
	body, err := json.Marshal(fields)
	if err != nil || strings.Contains(string(body), "secret") || strings.Contains(string(body), "task_id") {
		t.Fatal("raw projection metadata leaked")
	}
	if timing.fields(context.Background(), true)["outcome"] != "failed" {
		t.Fatal("validation failure hidden by recovery no-op")
	}
}

func TestProjectionTimingMeasuresFailedPhase(t *testing.T) {
	timing := newAgentProjectionTiming(AgentEventRequest{}, TaskEvent{State: AssignmentCompleted})
	original := errors.New("upstream-private")
	err := measureProjectionPhase(&timing.save, func() error { return original })
	if !errors.Is(err, original) || timing.save < 0 {
		t.Fatal("failed save duration or error identity lost")
	}
	if timing.fields(context.Background(), true)["outcome"] != "failed" {
		t.Fatal("failed save reported as success")
	}
}
