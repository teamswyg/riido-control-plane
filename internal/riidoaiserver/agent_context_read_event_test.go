package riidoaiserver

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestAgentContextReadEventRestoresTypedForVisibility(t *testing.T) {
	event := newAgentContextReadEvent([]AssignmentContextRead{{AssignmentID: "asg", TaskID: "task", AgentID: "agent", Via: AssignmentContextViaCLI, SourceKind: "task", SourceID: "task", SourcePart: "body", Status: AssignmentContextReadServed, ServedAt: time.Now()}})
	raw, _ := json.Marshal(event)
	restored, err := restoreSnapshotEventPayload(AgentClientEventContextRead, raw)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := restored.(AgentContextReadEvent); !ok || got.AgentID != "agent" {
		t.Fatalf("restored %T, visibility filtering needs the typed event", restored)
	}
	if agentID, ok := eventAgentID(restored); !ok || agentID != "agent" {
		t.Fatal("restored event lost its agent scope")
	}
}

func TestAgentContextReadEventCapsSources(t *testing.T) {
	reads := make([]AssignmentContextRead, 40)
	for i := range reads {
		reads[i] = AssignmentContextRead{AssignmentID: "asg", SourceKind: "link", SourceID: fmt.Sprint(i)}
	}
	if event := newAgentContextReadEvent(reads); len(event.Sources) != agentContextReadEventMaxSources || event.OmittedSources != 8 {
		t.Fatalf("sources=%d omitted=%d", len(event.Sources), event.OmittedSources)
	}
}
