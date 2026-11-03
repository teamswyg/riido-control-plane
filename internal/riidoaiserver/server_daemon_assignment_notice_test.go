package riidoaiserver

import (
	"bufio"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

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
