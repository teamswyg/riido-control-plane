package riidoaiserver

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

type fakeActiveLeaseStore struct {
	AssignmentOperationStore
	lease  AssignmentActiveLease
	exists bool
}

func (f fakeActiveLeaseStore) LoadAgentActiveAssignment(context.Context, string) (AssignmentActiveLease, bool, error) {
	return f.lease, f.exists, nil
}

func (f fakeActiveLeaseStore) RefreshAgentActiveAssignment(context.Context, Assignment, time.Time) error {
	return nil
}

func TestAssignmentLeaseCurrentUsesDurableExpiry(t *testing.T) {
	now := time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC)
	stale := Assignment{ID: "asg", AgentID: "agent", UpdatedAt: now.Add(-time.Minute)}
	for _, tc := range []struct {
		name  string
		store fakeActiveLeaseStore
		want  bool
	}{
		{"expired", fakeActiveLeaseStore{lease: AssignmentActiveLease{ActiveAssignmentID: "asg", LeaseExpiresAt: now.Add(-time.Second)}, exists: true}, false},
		{"other-assignment", fakeActiveLeaseStore{lease: AssignmentActiveLease{ActiveAssignmentID: "next", LeaseExpiresAt: now.Add(time.Minute)}, exists: true}, false},
		{"missing", fakeActiveLeaseStore{}, false},
		{"held", fakeActiveLeaseStore{lease: AssignmentActiveLease{ActiveAssignmentID: "asg", LeaseExpiresAt: now.Add(time.Minute)}, exists: true}, true},
	} {
		store := &Store{now: func() time.Time { return now }, activeLeaseDuration: 20 * time.Second, operationStore: tc.store}
		if got, err := store.AssignmentLeaseCurrent(context.Background(), stale); err != nil || got != tc.want {
			t.Fatalf("%s: got %v err=%v", tc.name, got, err)
		}
	}
	fresh := &Store{now: func() time.Time { return now }, activeLeaseDuration: 20 * time.Second}
	if got, _ := fresh.AssignmentLeaseCurrent(context.Background(), Assignment{UpdatedAt: now}); !got {
		t.Fatal("a just-updated assignment holds its lease")
	}
}

func TestPromptRelatedReadsMatchRenderedItems(t *testing.T) {
	c := extendedAssignmentContextForTest(0)
	for i := range 80 {
		c.Extended.RelatedWork = append(c.Extended.RelatedWork, AIAgentTaskContextRelatedWork{ID: fmt.Sprintf("w%02d", i), Kind: "task", Title: strings.Repeat("긴 제목 ", 6), Relation: "related"})
	}
	prompt, _ := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{TaskID: "task-private", Context: c})
	lines, omitted := promptRelatedLines(c.Extended)
	if omitted == 0 || !strings.Contains(prompt.Prompt, fmt.Sprintf("%d more related item(s) not shown", omitted)) {
		t.Fatalf("omitted=%d prompt lacks the overflow note", omitted)
	}
	for _, source := range assignmentContextPromptSources(c, nil) {
		if strings.HasPrefix(source.Kind, "related_") && !strings.Contains(prompt.Prompt, "[id: "+source.ID+"]") {
			t.Fatalf("read record for %s, which the prompt does not contain", source.ID)
		}
	}
	if len(lines)+omitted != 82 {
		t.Fatalf("lines=%d omitted=%d", len(lines), omitted)
	}
}

func TestPromptResourceReadsMatchRenderedEntries(t *testing.T) {
	c := assignmentContextTestTaskContext()
	c.Resources = nil
	for i := range 40 {
		c.Resources = append(c.Resources, AIAgentTaskContextResource{ID: fmt.Sprintf("r%02d", i), Kind: "link", Title: strings.Repeat("링크 ", 20), URL: "https://example.com/" + strings.Repeat("p", 60)})
	}
	c.ResourceStatus = "available"
	for _, extended := range []*AIAgentTaskContextExtended{nil, extendedAssignmentContextForTest(0).Extended} {
		c.Extended = extended
		prompt, _ := ComposeAIAgentAssignmentPrompt(AIAgentAssignmentPromptInput{TaskID: "task-private", Context: c})
		recorded := 0
		for _, source := range assignmentContextPromptSources(c, nil) {
			if source.Kind != "link" {
				continue
			}
			recorded++
			if !strings.Contains(prompt.Prompt, "[id: "+source.ID+"]") {
				t.Fatalf("extended=%v: %s recorded but not in the prompt", extended != nil, source.ID)
			}
		}
		if recorded == 0 || recorded == 40 || !strings.Contains(prompt.Prompt, "more link(s) or attachment(s) not shown") {
			t.Fatalf("extended=%v: recorded=%d", extended != nil, recorded)
		}
	}
}
