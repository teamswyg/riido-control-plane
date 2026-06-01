package riidoaiserver

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAssignmentAPIJSONShapes(t *testing.T) {
	now := time.Date(2026, 5, 27, 11, 0, 0, 0, time.UTC)
	assignment := Assignment{
		ID:                    "asn-000001",
		TaskID:                "task-a",
		ComponentID:           "component-1",
		AgentID:               "agent-a",
		RuntimeProvider:       "codex",
		Prompt:                "run tests",
		AgentInstruction:      "act as QA",
		State:                 AssignmentLeased,
		LeaseToken:            "lease-1",
		ReplacesAssignmentID:  "asn-old",
		BlockedByAssignmentID: "asn-blocker",
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	assertJSON(t, "assign request", AssignRequest{
		ComponentID:      "component-1",
		AgentID:          "agent-a",
		RuntimeProvider:  "codex",
		Prompt:           "run tests",
		AgentInstruction: "act as QA",
		CreatedBy:        "user-a",
	}, `{"component_id":"component-1","agent_id":"agent-a","runtime_provider":"codex","prompt":"run tests","agent_instruction":"act as QA","created_by":"user-a"}`)
	assertJSON(t, "poll request", PollRequest{
		DaemonID:  "daemon-a",
		DeviceID:  "device-a",
		RuntimeID: "daemon-a:agent:agent-a:codex",
	}, `{"daemon_id":"daemon-a","device_id":"device-a","runtime_id":"daemon-a:agent:agent-a:codex"}`)
	assertJSON(t, "poll response", PollResponse{
		SchemaVersion: SchemaVersion,
		Action:        PollStart,
		Assignment:    &assignment,
	}, `{"schema_version":"riido-ai-server.v1","action":"start","assignment":{"assignment_id":"asn-000001","task_id":"task-a","component_id":"component-1","agent_id":"agent-a","runtime_provider":"codex","prompt":"run tests","agent_instruction":"act as QA","state":"leased","lease_token":"lease-1","replaces_assignment_id":"asn-old","blocked_by_assignment_id":"asn-blocker","created_at":"2026-05-27T11:00:00Z","updated_at":"2026-05-27T11:00:00Z"}}`)
	assertJSON(t, "heartbeat request", AgentHeartbeatRequest{
		DaemonID:            "daemon-a",
		DeviceID:            "device-a",
		RuntimeID:           "daemon-a:agent:agent-a:codex",
		RunningTaskIDs:      []string{"task-a"},
		ActiveAssignmentIDs: []string{"asn-000001"},
	}, `{"daemon_id":"daemon-a","device_id":"device-a","runtime_id":"daemon-a:agent:agent-a:codex","running_task_ids":["task-a"],"active_assignment_ids":["asn-000001"]}`)
	assertJSON(t, "heartbeat response", AgentHeartbeatResponse{
		SchemaVersion:        SchemaVersion,
		RefreshedAssignments: []Assignment{assignment},
	}, `{"schema_version":"riido-ai-server.v1","refreshed_assignments":[{"assignment_id":"asn-000001","task_id":"task-a","component_id":"component-1","agent_id":"agent-a","runtime_provider":"codex","prompt":"run tests","agent_instruction":"act as QA","state":"leased","lease_token":"lease-1","replaces_assignment_id":"asn-old","blocked_by_assignment_id":"asn-blocker","created_at":"2026-05-27T11:00:00Z","updated_at":"2026-05-27T11:00:00Z"}]}`)
	assertJSON(t, "agent event request", AgentEventRequest{
		AssignmentID: "asn-000001",
		TaskID:       "task-a",
		DaemonID:     "daemon-a",
		DeviceID:     "device-a",
		RuntimeID:    "daemon-a:agent:agent-a:codex",
		State:        AssignmentRunning,
		EventType:    EventRiidoLog,
		Message:      "working",
		Metadata:     map[string]string{"step": "test"},
	}, `{"assignment_id":"asn-000001","task_id":"task-a","daemon_id":"daemon-a","device_id":"device-a","runtime_id":"daemon-a:agent:agent-a:codex","state":"running","event_type":"riido_log","message":"working","metadata":{"step":"test"}}`)
	event := TaskEvent{
		Seq:          1,
		TaskID:       "task-a",
		AssignmentID: "asn-000001",
		AgentID:      "agent-a",
		Type:         EventAssignmentRunning,
		State:        AssignmentRunning,
		Message:      "running",
		Metadata:     map[string]string{"step": "run"},
		At:           now,
	}
	assertJSON(t, "agent event response", AgentEventResponse{
		SchemaVersion: SchemaVersion,
		Assignment:    &assignment,
		Event:         event,
	}, `{"schema_version":"riido-ai-server.v1","assignment":{"assignment_id":"asn-000001","task_id":"task-a","component_id":"component-1","agent_id":"agent-a","runtime_provider":"codex","prompt":"run tests","agent_instruction":"act as QA","state":"leased","lease_token":"lease-1","replaces_assignment_id":"asn-old","blocked_by_assignment_id":"asn-blocker","created_at":"2026-05-27T11:00:00Z","updated_at":"2026-05-27T11:00:00Z"},"event":{"seq":1,"task_id":"task-a","assignment_id":"asn-000001","agent_id":"agent-a","type":"assignment_running","state":"running","message":"running","metadata":{"step":"run"},"at":"2026-05-27T11:00:00Z"}}`)
	assertJSON(t, "health", Health{
		SchemaVersion: SchemaVersion,
		Status:        "ok",
	}, `{"schema_version":"riido-ai-server.v1","status":"ok"}`)
	assertJSON(t, "metrics", MetricsSnapshot{
		SchemaVersion:                       MetricsSchemaVersion,
		GeneratedAt:                         now,
		TasksTotal:                          2,
		AssignmentsTotal:                    3,
		AssignmentsByState:                  map[AssignmentState]int{AssignmentQueued: 1, AssignmentRunning: 2},
		PollRequestsTotal:                   4,
		PollActionsTotal:                    map[PollAction]int64{PollNone: 2, PollStart: 1, PollActive: 1},
		AgentEventsTotal:                    5,
		TaskEventsTotal:                     6,
		SSESubscribers:                      7,
		OutboxErrorsTotal:                   8,
		EventAppendLatencySamplesTotal:      9,
		EventAppendLatencyTotalMilliseconds: 10,
		EventAppendLatencyMaxMilliseconds:   11,
		EventAppendLatencyLastMilliseconds:  12,
	}, `{"schema_version":"riido-ai-server-metrics.v1","generated_at":"2026-05-27T11:00:00Z","tasks_total":2,"assignments_total":3,"assignments_by_state":{"queued":1,"running":2},"poll_requests_total":4,"poll_actions_total":{"active":1,"none":2,"start":1},"agent_events_total":5,"task_events_total":6,"sse_subscribers":7,"outbox_errors_total":8,"event_append_latency_samples_total":9,"event_append_latency_total_ms":10,"event_append_latency_max_ms":11,"event_append_latency_last_ms":12}`)
}

func assertJSON(t *testing.T, name string, value any, want string) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal %s: %v", name, err)
	}
	if got := string(data); got != want {
		t.Fatalf("%s JSON = %s, want %s", name, got, want)
	}
}
