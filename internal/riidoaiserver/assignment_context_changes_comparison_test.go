package riidoaiserver

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestContextChangesOmissionIsUnknownWithoutOldTitleOrContent(t *testing.T) {
	previous, _ := contextChangesTestBundle(t, contextChangesTestSnapshot("previous"), "old private body")
	previous.Context.ReadScope.Resources[0].Title = "old private title"
	current, _ := contextChangesTestBundle(t, contextChangesTestSnapshot("current"), "current body")
	current.Context.ReadScope.Resources = nil
	current.Content = nil
	out := compareAssignmentContextParts(previous, current)
	raw, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Changes) != 1 || out.Changes[0].Status != "unavailable" || out.Changes[0].Reason != "not_in_current_capture" || out.Changes[0].Current != nil || out.Complete || !out.FullRefreshRequired || out.NextAction != "riido context get" {
		t.Fatal("omission was not unknown", string(raw))
	}
	for _, forbidden := range []string{"private", "deleted", "denied", "title"} {
		if strings.Contains(string(raw), forbidden) {
			t.Fatal("omission leaked old detail or invented cause", string(raw))
		}
	}
}

func TestContextChangesStableAndAddedContentHaveBoundedMeaning(t *testing.T) {
	previous, _ := contextChangesTestBundle(t, contextChangesTestSnapshot("previous"), "body")
	current, _ := contextChangesTestBundle(t, contextChangesTestSnapshot("current"), "body")
	if got := compareAssignmentContextParts(previous, current); len(got.Changes) != 0 || !got.Complete {
		t.Fatal("new delivery cursors counted as content change", got)
	}
	previous.Content = nil
	previous.Context.ReadScope.Resources = nil
	if got := compareAssignmentContextParts(previous, current); len(got.Changes) != 1 || got.Changes[0].Status != "updated" || !got.Complete {
		t.Fatal("new source missing", got)
	}
	current.Truncated = true
	if got := compareAssignmentContextParts(current, current); got.Complete || !got.FullRefreshRequired {
		t.Fatal("partial capture declared reusable", got)
	}
}

func TestContextChangesThreadLineageUsesPublicationIdentity(t *testing.T) {
	assignment := Assignment{ID: "current", TaskID: "work", AgentID: "agent"}
	response := AIAgentTaskActionResponse{AssignmentID: assignment.ID, TaskID: assignment.TaskID, AgentID: assignment.AgentID, ThreadID: "next-thread"}
	selected := AIAgentTaskThreadRecord{AssignmentID: "parent", TaskID: "work", AgentID: "agent", ThreadID: "parent-thread", ConversationID: "conversation"}
	trigger := assignmentContextTrigger{}.withThread(assignment, response, &selected)
	if trigger.PreviousAssignmentID != "parent" || trigger.PreviousThreadID != "parent-thread" || trigger.ConversationID != "conversation" {
		t.Fatal("selected lineage missing")
	}
	response.AssignmentID = "unrelated"
	if got := (assignmentContextTrigger{}).withThread(assignment, response, &selected); got.ConversationID != "" {
		t.Fatal("mismatched publication accepted")
	}
	response.AssignmentID = assignment.ID
	selected.AgentID = "other"
	if got := (assignmentContextTrigger{}).withThread(assignment, response, &selected); got.ConversationID != "" {
		t.Fatal("mismatched selection accepted")
	}
}
