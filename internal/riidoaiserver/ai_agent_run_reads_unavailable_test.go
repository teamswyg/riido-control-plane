package riidoaiserver

import "testing"

func TestAIAgentRunReadsIgnoresUnavailableSectionFallbacks(t *testing.T) {
	current := AIAgentTaskContext{
		Component: AIAgentTaskContextComponent{ID: "task"}, ResourceStatus: "unavailable", Resources: []AIAgentTaskContextResource{{ID: "link"}},
		Hierarchy: AIAgentTaskContextHierarchy{Project: AIAgentTaskContextReference{ID: "project", DocumentStatus: "unavailable", Document: &AIAgentTaskContextDocument{Content: "stale body"}}},
		Extended:  &AIAgentTaskContextExtended{CommentStatus: "unavailable", Comments: []AIAgentTaskContextComment{{ID: "comment"}}, RelatedStatus: "unavailable", RelatedMeetings: []AIAgentTaskContextRelatedMeeting{{ID: "meeting"}}},
	}
	visible := visibleRunReadSources(current)
	for _, key := range []runReadSourceKey{{"link", "link", "metadata"}, {"project", "project", "body"}, {"comment", "comment", "body"}, {"related_meeting", "meeting", "title"}} {
		if visible[key] {
			t.Fatalf("unverified fallback granted: %+v", key)
		}
	}
}
