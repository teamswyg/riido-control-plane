package riidoaiserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func TestHTTPTracingUsesAllowlistedRoutePatterns(t *testing.T) {
	recorder := &recordingTraceRecorder{}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/agents/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusAccepted, Health{SchemaVersion: SchemaVersion, Status: "ok"})
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, Health{SchemaVersion: SchemaVersion, Status: "ok"})
	})
	handler := withHTTPTracing(mux, recorder)

	pollReq := httptest.NewRequest(http.MethodPost, "/v1/agents/agent-a/poll", nil)
	pollResp := httptest.NewRecorder()
	handler.ServeHTTP(pollResp, pollReq)
	if pollResp.Code != http.StatusAccepted {
		t.Fatalf("poll status=%d body=%s", pollResp.Code, pollResp.Body.String())
	}
	healthReq := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	healthResp := httptest.NewRecorder()
	handler.ServeHTTP(healthResp, healthReq)
	if healthResp.Code != http.StatusOK {
		t.Fatalf("health status=%d body=%s", healthResp.Code, healthResp.Body.String())
	}

	spans := recorder.snapshot()
	if len(spans) != 1 {
		t.Fatalf("spans = %+v", spans)
	}
	if spans[0].Name != "HTTP POST /v1/agents/{agent_id}/poll" {
		t.Fatalf("span name = %q", spans[0].Name)
	}
	if spans[0].Attributes[metadatakeys.HTTPRoute.String()] != "/v1/agents/{agent_id}/poll" ||
		spans[0].Attributes[metadatakeys.HTTPResponseStatusCode.String()] != "202" ||
		spans[0].Attributes[metadatakeys.HTTPStatusCode.String()] != "202" {
		t.Fatalf("span attributes = %+v", spans[0].Attributes)
	}
}

func TestStoreTracingRecordsOperationSpans(t *testing.T) {
	recorder := &recordingTraceRecorder{}
	store := NewStoreWithConfig(StoreConfig{TraceRecorder: recorder})
	defer store.Close()
	ctx := context.Background()
	assignment, err := store.AssignTask(ctx, "task-a", AssignRequest{
		ComponentID:     "component-a",
		AgentID:         "agent-a",
		RuntimeProvider: "codex",
		Prompt:          "run",
	})
	if err != nil {
		t.Fatalf("AssignTask: %v", err)
	}
	if _, err := store.PollAgent(ctx, assignment.AgentID, PollRequest{DaemonID: "daemon-a", RuntimeID: "runtime-a"}); err != nil {
		t.Fatalf("PollAgent: %v", err)
	}
	if _, err := store.RecordAgentEvent(ctx, assignment.AgentID, AgentEventRequest{
		AssignmentID: assignment.ID,
		TaskID:       assignment.TaskID,
		DaemonID:     "daemon-a",
		RuntimeID:    "runtime-a",
		State:        AssignmentRunning,
		EventType:    EventAssignmentRunning,
	}); err != nil {
		t.Fatalf("RecordAgentEvent: %v", err)
	}

	names := traceSpanNames(recorder.snapshot())
	for _, want := range []string{
		"store.store_create_assignment",
		"store.store_poll_assignment",
		"store.store_append_event",
	} {
		if !slices.Contains(names, want) {
			t.Fatalf("missing span %q in %v", want, names)
		}
	}
}

func TestStoreTracingRecordsLongPollPressure(t *testing.T) {
	recorder := &recordingTraceRecorder{}
	store := NewStoreWithConfig(StoreConfig{TraceRecorder: recorder})
	defer store.Close()
	_, err := store.WaitForAssignment(context.Background(), "agent-a",
		PollRequest{DaemonID: "daemon-a", RuntimeID: "runtime-a", WaitMs: 20}, 20*time.Millisecond, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitForAssignment: %v", err)
	}
	var span recordedTraceSpanSnapshot
	for _, candidate := range recorder.snapshot() {
		if candidate.Name == "store."+StoreOperationWaitAssignment.String() {
			span = candidate
			break
		}
	}
	if span.Name == "" {
		t.Fatal("missing wait assignment span")
	}
	if span.Attributes[riidoPollWaitedKey] != "true" || span.Attributes[metadatakeys.RiidoPollAction.String()] != string(PollNone) {
		t.Fatalf("wait span attrs = %+v", span.Attributes)
	}
	if span.Attributes[riidoPollHoldMsKey] != "20" || span.Attributes[riidoPollTickMsKey] != "10" {
		t.Fatalf("wait pressure attrs = %+v", span.Attributes)
	}
}

func TestHTTPAssignmentTaskContextTracingRecordsDomainOperation(t *testing.T) {
	recorder := &recordingTraceRecorder{}
	store := NewStoreWithConfig(StoreConfig{TraceRecorder: recorder})
	defer store.Close()
	taskContext := &assignmentHTTPTaskContextReader{contextSnapshot: aiAgentTaskContextHTTPFixture()}
	server := NewServer(ServerConfig{
		Assignment:    store,
		TaskContext:   taskContext,
		TraceRecorder: recorder,
		Authorizer:    assignmentHTTPAuthorizer(t, []string{"component-task:task-a:assign"}),
	}).Handler()

	req := httptest.NewRequest(http.MethodPost, "/v1/component-tasks/task-a/assignment", strings.NewReader(`{"component_id":"component-a","agent_id":"agent-a","runtime_provider":"codex"}`))
	req.Header.Set("Authorization", "Bearer assignment-token")
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("assign status=%d body=%s", resp.Code, resp.Body.String())
	}

	spans := recorder.snapshot()
	names := traceSpanNames(spans)
	if !slices.Contains(names, "task_context.task_context_resolve") {
		t.Fatalf("missing task context span in %v", names)
	}
	var taskContextSpan recordedTraceSpanSnapshot
	for _, span := range spans {
		if span.Name == "task_context.task_context_resolve" {
			taskContextSpan = span
			break
		}
	}
	if got := taskContextSpan.Attributes[metadatakeys.RiidoTaskContextOperation.String()]; got != TaskContextOperationResolve.String() {
		t.Fatalf("task context operation attr = %q, want %q; attrs=%+v", got, TaskContextOperationResolve.String(), taskContextSpan.Attributes)
	}
	if got := taskContextSpan.Attributes[metadatakeys.RiidoTraceSurface.String()]; got != "task_context" {
		t.Fatalf("task context surface = %q, want task_context; attrs=%+v", got, taskContextSpan.Attributes)
	}
}

func TestDynamoDBAIAgentClientSnapshotUsesTraceContext(t *testing.T) {
	recorder := &recordingTraceRecorder{}
	fixedNow := time.Date(2026, 6, 16, 1, 2, 3, 0, time.UTC)
	snapshot := AIAgentClientSnapshot{
		SchemaVersion:           AIAgentClientPersistenceSchemaVersion,
		SavedAt:                 fixedNow,
		WorkspaceID:             "workspace-dev",
		TaskThreads:             map[string][]AIAgentTaskThreadRecord{},
		NextDeviceCredentialSeq: 1,
	}
	items := map[string]map[string]map[string]string{}
	var itemsMu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		switch r.Header.Get("X-Amz-Target") {
		case dynamoDBPutItemTarget:
			var payload struct {
				Item map[string]map[string]string `json:"Item"`
			}
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Errorf("decode PutItem payload: %v", err)
			}
			itemsMu.Lock()
			items[payload.Item["sk"]["S"]] = payload.Item
			itemsMu.Unlock()
			_, _ = w.Write([]byte(`{}`))
		case dynamoDBQueryTarget:
			itemsMu.Lock()
			out := make([]map[string]map[string]string, 0, len(items))
			for _, item := range items {
				out = append(out, item)
			}
			itemsMu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{"Items": out})
		default:
			t.Errorf("target = %q", r.Header.Get("X-Amz-Target"))
			_, _ = w.Write([]byte(`{}`))
			return
		}
	}))
	defer server.Close()
	provider, err := NewStaticAWSCredentialsProvider("AKID", "SECRET", "")
	if err != nil {
		t.Fatalf("NewStaticAWSCredentialsProvider: %v", err)
	}
	store, err := NewDynamoDBAIAgentClientSnapshot(DynamoDBAIAgentClientSnapshotConfig{
		Region:              "ap-northeast-2",
		TableName:           "riido-ai-agent-development",
		Endpoint:            server.URL,
		CredentialsProvider: provider,
		Now:                 func() time.Time { return fixedNow },
	})
	if err != nil {
		t.Fatalf("NewDynamoDBAIAgentClientSnapshot: %v", err)
	}
	defer store.Close()
	ctx := WithTraceRecorder(context.Background(), recorder)
	if err := store.SaveAIAgentClientSnapshot(ctx, snapshot); err != nil {
		t.Fatalf("SaveAIAgentClientSnapshot: %v", err)
	}
	if _, ok, err := store.LoadAIAgentClientSnapshot(ctx); err != nil || !ok {
		t.Fatalf("LoadAIAgentClientSnapshot ok=%v err=%v", ok, err)
	}

	names := traceSpanNames(recorder.snapshot())
	for _, want := range []string{aiAgentClientSnapshotSaveTraceName, "aws.dynamodb.PutItem", "aws.dynamodb.Query"} {
		if !slices.Contains(names, want) {
			t.Fatalf("missing span %q in %v", want, names)
		}
	}
	spans := recorder.snapshot()
	spanByName := map[string]recordedTraceSpanSnapshot{}
	for _, span := range spans {
		spanByName[span.Name] = span
	}
	if got := spanByName["aws.dynamodb.PutItem"].Attributes[metadatakeys.RiidoStoreOperation.String()]; got != AIAgentClientSnapshotSave.String() {
		t.Fatalf("PutItem store operation = %q, want %q", got, AIAgentClientSnapshotSave.String())
	}
	if got := spanByName["aws.dynamodb.Query"].Attributes[metadatakeys.RiidoStoreOperation.String()]; got != AIAgentClientSnapshotLoad.String() {
		t.Fatalf("Query store operation = %q, want %q", got, AIAgentClientSnapshotLoad.String())
	}
	saveSpan := spanByName[aiAgentClientSnapshotSaveTraceName]
	wantItemsWritten := strconv.Itoa(len(dynamoDBAIAgentClientSnapshotPartNames) + 1)
	if got := saveSpan.Attributes[riidoSnapshotItemsWrittenKey]; got != wantItemsWritten {
		t.Fatalf("snapshot items_written = %q, want %s; attrs=%+v", got, wantItemsWritten, saveSpan.Attributes)
	}
	if got := saveSpan.Attributes[riidoSnapshotPartsSkippedKey]; got != "0" {
		t.Fatalf("snapshot parts_skipped = %q, want 0; attrs=%+v", got, saveSpan.Attributes)
	}
	if got := saveSpan.Attributes[riidoSnapshotBytesEncodedKey]; got == "" || got == "0" {
		t.Fatalf("snapshot bytes_encoded = %q, want positive; attrs=%+v", got, saveSpan.Attributes)
	}
}

type recordingTraceRecorder struct {
	mu    sync.Mutex
	spans []*recordingTraceSpan
}

type recordingTraceSpan struct {
	mu         sync.Mutex
	Name       string
	Kind       TraceSpanKind
	Attributes map[string]string
	Errors     []string
	Ended      bool
}

type recordedTraceSpanSnapshot struct {
	Name       string
	Kind       TraceSpanKind
	Attributes map[string]string
	Errors     []string
	Ended      bool
}

func (r *recordingTraceRecorder) StartTraceSpan(ctx context.Context, start TraceSpanStart) (context.Context, TraceSpan) {
	span := &recordingTraceSpan{
		Name:       start.Name,
		Kind:       start.Kind,
		Attributes: map[string]string{},
	}
	for _, attr := range start.Attributes {
		span.Attributes[attr.Key] = attr.StringValue()
	}
	r.mu.Lock()
	r.spans = append(r.spans, span)
	r.mu.Unlock()
	return ctx, span
}

func (r *recordingTraceRecorder) snapshot() []recordedTraceSpanSnapshot {
	r.mu.Lock()
	spans := append([]*recordingTraceSpan(nil), r.spans...)
	r.mu.Unlock()
	out := make([]recordedTraceSpanSnapshot, 0, len(spans))
	for _, span := range spans {
		span.mu.Lock()
		attributes := make(map[string]string, len(span.Attributes))
		for key, value := range span.Attributes {
			attributes[key] = value
		}
		out = append(out, recordedTraceSpanSnapshot{
			Name:       span.Name,
			Kind:       span.Kind,
			Attributes: attributes,
			Errors:     append([]string(nil), span.Errors...),
			Ended:      span.Ended,
		})
		span.mu.Unlock()
	}
	return out
}

func (s *recordingTraceSpan) SetAttributes(attributes ...TraceAttribute) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, attr := range attributes {
		s.Attributes[attr.Key] = attr.StringValue()
	}
}

func (s *recordingTraceSpan) RecordError(err error) {
	if err == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Errors = append(s.Errors, err.Error())
}

func (s *recordingTraceSpan) End() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Ended = true
}

func traceSpanNames(spans []recordedTraceSpanSnapshot) []string {
	names := make([]string, 0, len(spans))
	for _, span := range spans {
		names = append(names, span.Name)
	}
	return names
}
