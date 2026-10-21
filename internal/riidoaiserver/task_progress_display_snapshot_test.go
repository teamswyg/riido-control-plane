package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func TestContextualProgressSnapshotRestoresReplaySummary(t *testing.T) {
	line := AgentThreadProgressLine{Seq: 1, Message: "rendered progress", MessageCode: 1103, MessageArgs: map[string]string{"description": "답글 전송 권한을 점검하고 있어요."}}
	event := AgentThreadProgressEvent{ThreadID: "thread", Lines: []AgentThreadProgressLine{line}}
	eventJSON, _ := json.Marshal(event)
	snapshot := AIAgentClientSnapshot{
		TaskThreads: map[string][]AIAgentTaskThreadRecord{"task": {{ThreadID: "thread", Lines: event.Lines}}},
		Events:      []AIAgentClientEventSnapshot{{EventType: AgentClientEventThreadProgress, Payload: eventJSON}},
	}
	raw, _ := json.Marshal(snapshot)
	var restored AIAgentClientSnapshot
	if err := json.Unmarshal(raw, &restored); err != nil {
		t.Fatal(err)
	}
	if got := taskProgressDisplayMessage(restored.TaskThreads["task"][0].Lines[0]); got != taskProgressDisplayMessage(line) {
		t.Fatalf("snapshot thread lost summary: %q", got)
	}
	payload, err := restoreSnapshotEventPayload(restored.Events[0].EventType, restored.Events[0].Payload)
	if err != nil {
		t.Fatal(err)
	}
	replay := payload.(AgentThreadProgressEvent)
	if len(replay.Lines) != 1 || taskProgressDisplayMessage(replay.Lines[0]) != taskProgressDisplayMessage(line) {
		t.Fatalf("snapshot replay lost summary: %+v", replay)
	}
}
