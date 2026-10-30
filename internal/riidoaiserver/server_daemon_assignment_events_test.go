package riidoaiserver

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDaemonAssignmentEventsScopeClaimAndReconnect(t *testing.T) {
	f := newDeviceEnrollmentHTTPFixture(t)
	f.enrollAndRotateDeviceCredential()
	f.syncCodexRuntimeAndCreateAgent()
	assignments := NewStoreWithConfig(StoreConfig{AgentRegistry: f.aiAgentStore})
	defer assignments.Close()
	server := httptest.NewServer(NewServer(ServerConfig{AIAgentClient: f.aiAgentStore, Assignment: assignments}).Handler())
	defer server.Close()
	connect := func(secret string) (*http.Response, context.CancelFunc) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/v1/daemon/assignment-events", nil)
		req.Header.Set(deviceIDHeader, f.enrollment.DeviceID)
		req.Header.Set(deviceSecretHeader, secret)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			cancel()
			t.Fatal(err)
		}
		return resp, cancel
	}
	rejected, cancel := connect("wrong-secret")
	if rejected.StatusCode != 401 {
		t.Fatalf("invalid secret status=%d", rejected.StatusCode)
	}
	rejected.Body.Close()
	cancel()
	resp, cancel := connect(f.enrollment.DeviceSecret)
	defer cancel()
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	scanner := bufio.NewScanner(resp.Body)
	first := readDaemonNotice(t, scanner)
	if len(first.AgentIDs) != 1 || first.AgentIDs[0] != f.created.Agent.AgentID {
		t.Fatalf("wrong device scope %+v", first)
	}
	a, err := assignments.AssignTask(context.Background(), "stream-task", AssignRequest{ComponentID: "component-1", AgentID: first.AgentIDs[0], RuntimeProvider: "codex", Prompt: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	next := readDaemonNotice(t, scanner)
	if len(next.AgentIDs) != 1 || next.AgentIDs[0] != a.AgentID {
		t.Fatalf("missing assignment notice %+v", next)
	}
	// Disconnect before claiming. Reconnect must include the still-queued work.
	cancel()
	resp.Body.Close()
	again, cancelAgain := connect(f.enrollment.DeviceSecret)
	defer cancelAgain()
	defer again.Body.Close()
	replay := readDaemonNotice(t, bufio.NewScanner(again.Body))
	if len(replay.AgentIDs) != 1 {
		t.Fatalf("missing reconnect reconciliation %+v", replay)
	}
	poll, err := assignments.PollAgent(context.Background(), a.AgentID, PollRequest{DaemonID: first.Bindings[0].DaemonID, DeviceID: f.enrollment.DeviceID, RuntimeID: f.codexRuntimeID})
	if err != nil || poll.Action != PollStart {
		t.Fatalf("notification changed ownership: %+v %v", poll, err)
	}
}

func readDaemonNotice(t *testing.T, scanner *bufio.Scanner) daemonAssignmentEvent {
	t.Helper()
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			var event daemonAssignmentEvent
			if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &event); err != nil {
				t.Fatal(err)
			}
			return event
		}
	}
	t.Fatalf("stream ended: %v", scanner.Err())
	return daemonAssignmentEvent{}
}

func TestDaemonAssignmentNotificationsIgnoreOutputAndNoticeBindingChanges(t *testing.T) {
	before := daemonAssignmentEvent{Bindings: []AgentRuntimeBinding{{AgentID: "a", RuntimeID: "old"}}}
	after := daemonAssignmentEvent{Bindings: []AgentRuntimeBinding{{AgentID: "a", RuntimeID: "new"}, {AgentID: "b", RuntimeID: "new"}}}
	changed := changedDaemonAssignmentAgents(before, after, map[string]string{"a": "same"}, map[string]string{"a": "same", "b": ""})
	if len(changed) != 2 {
		t.Fatalf("binding changes lost: %v", changed)
	}
	a := Assignment{ID: "a", State: AssignmentRunning, UpdatedAt: time.Now()}
	b := a
	b.UpdatedAt = b.UpdatedAt.Add(time.Second)
	if assignmentSignalPart(a) != assignmentSignalPart(b) {
		t.Fatal("heartbeat produced a wakeup")
	}
}
