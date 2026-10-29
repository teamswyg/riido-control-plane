package riidoaiserver

import (
	"testing"
	"time"
)

func TestAIAgentRunReadsSourceTotalSkipsTitleOnlyPromptGroups(t *testing.T) {
	thread := AIAgentTaskThreadRecord{TaskID: "task", RunID: "run", ThreadID: "thread", AssignmentID: "assignment", AgentID: "agent"}
	at := time.Date(2026, 10, 29, 0, 0, 0, 0, time.UTC)
	read := func(kind, id, part, via string, i int) AssignmentContextRead {
		return AssignmentContextRead{AssignmentID: "assignment", TaskID: "task", AgentID: "agent", SourceKind: kind, SourceID: id, SourcePart: part, Via: via, Status: "served", ServedAt: at.Add(time.Duration(i) * time.Second)}
	}
	reads := []AssignmentContextRead{
		read("task", "task", "body", AssignmentContextViaPrompt, 0),
		read("related_task", "title-only", "title", AssignmentContextViaPrompt, 1),
		read("related_task", "opened", "title", AssignmentContextViaPrompt, 2),
		read("related_task", "opened", "body", AssignmentContextViaCLI, 3),
		read("project", "project", "summary", AssignmentContextViaPrompt, 4),
		read("comment", "trigger", "body", AssignmentContextViaPrompt, 5),
	}
	visible := map[runReadSourceKey]runReadMetadata{}
	for _, r := range reads {
		visible[runReadSourceKey{r.SourceKind, r.SourceID, r.SourcePart}] = runReadMetadata{}
	}
	visible[runReadSourceKey{"comment", "trigger", "body"}] = runReadMetadata{owner: runReadOwner{"task", "task"}}
	got, err := buildAIAgentRunReads(thread, reads, visible, 1)
	if err != nil {
		t.Fatal(err)
	}
	listed := map[string]bool{}
	for _, r := range got.Reads {
		listed[r.Kind+":"+r.ID+":"+r.Part] = r.Listed
	}
	if listed["related_task:title-only:title"] || listed["related_task:opened:title"] || listed["project:project:summary"] {
		t.Fatalf("title or summary prompt evidence listed: %v", listed)
	}
	if !listed["task:task:body"] || !listed["related_task:opened:body"] || !listed["comment:trigger:body"] {
		t.Fatalf("content reads must stay listed: %v", listed)
	}
	if got.SourceTotal != 2 {
		t.Fatalf("source_total=%d want task (with its trigger comment) and the opened related task", got.SourceTotal)
	}
}

func TestReadScopePromptSourcesMatchReadScopePrompt(t *testing.T) {
	c := AIAgentTaskContext{
		Component: AIAgentTaskContextComponent{ID: "task"},
		Document:  AIAgentTaskContextDocument{Content: "preview"},
		Hierarchy: AIAgentTaskContextHierarchy{Project: AIAgentTaskContextReference{ID: "project", Title: "Project", Document: &AIAgentTaskContextDocument{Content: "project body"}}},
		Resources: []AIAgentTaskContextResource{{ID: "link", Kind: "link"}},
		Extended:  &AIAgentTaskContextExtended{RelatedWork: []AIAgentTaskContextRelatedWork{{ID: "related", Kind: "task", Title: "Related"}}},
	}
	comments := []AIAgentTaskContextComment{{ID: "trigger", Body: "do it", Priority: TaskContextCommentPriorityTrigger}, {ID: "older", Body: "old"}}
	var got []string
	for _, source := range readScopePromptSources(c, comments) {
		got = append(got, source.Kind+":"+source.ID+":"+source.Part)
	}
	want := []string{"task:task:body", "project:project:title", "comment:trigger:body"}
	if len(got) != len(want) {
		t.Fatalf("prompt sources=%v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("prompt sources=%v want %v", got, want)
		}
	}
}
