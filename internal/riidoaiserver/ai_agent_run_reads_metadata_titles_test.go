package riidoaiserver

import "testing"

func TestAIAgentRunReadsMetadataCurrentTitles(t *testing.T) {
	for _, path := range []string{"/v1/client/ai-agent/runs/run/reads", "/v2/client/workspaces/" + defaultAIAgentClientWorkspaceID + "/ai-agent/runs/run/reads"} {
		t.Run(path, func(t *testing.T) {
			reads := []AssignmentContextRead{
				{SourceKind: "task", SourceID: "task", SourcePart: "body"},
				{SourceKind: "project", SourceID: "project", SourcePart: "summary"},
				{SourceKind: "milestone", SourceID: "goal", SourcePart: "title"},
				{SourceKind: "parent_task", SourceID: "parent", SourcePart: "body"},
				{SourceKind: "link", SourceID: "link", SourcePart: "metadata"},
				{SourceKind: "attachment", SourceID: "file", SourcePart: "metadata"},
				{SourceKind: "related_task", SourceID: "related", SourcePart: "title"},
				{SourceKind: "related_meeting", SourceID: "meeting", SourcePart: "body"},
				{SourceKind: "comment", SourceID: "comment", SourcePart: "body"},
			}
			server, reader, _ := runReadsTestServer(t, reads)
			reader.value.Component.Title = " Current task "
			reader.value.Hierarchy = AIAgentTaskContextHierarchy{
				Project:    AIAgentTaskContextReference{ID: "project", Title: "Current project", Document: &AIAgentTaskContextDocument{Content: "project body"}},
				Milestone:  AIAgentTaskContextReference{ID: "goal", Title: "Current goal"},
				ParentTask: AIAgentTaskContextReference{ID: "parent", Title: "Current parent", Document: &AIAgentTaskContextDocument{Content: "parent body"}},
			}
			reader.value.Resources = []AIAgentTaskContextResource{{ID: "link", Kind: "link", Title: "Current link"}, {ID: "file", Kind: "file", Title: "Current file"}}
			reader.value.Extended = &AIAgentTaskContextExtended{
				RelatedWork:     []AIAgentTaskContextRelatedWork{{ID: "related", Kind: "task", Title: "Current related"}},
				RelatedMeetings: []AIAgentTaskContextRelatedMeeting{{ID: "meeting", Title: "Current meeting", BodyStatus: "available", Document: &AIAgentTaskContextDocument{Content: "meeting body", Status: "available"}}},
				Comments:        []AIAgentTaskContextComment{{ID: "comment", Body: "Comment is never a title"}},
			}
			rows := runReadsMetadataRows(t, server, path)
			wantTitles := []string{"Current task", "Current project", "Current goal", "Current parent", "Current link", "Current file", "Current related", "Current meeting", ""}
			if len(rows) != len(wantTitles) {
				t.Fatalf("reads=%d want=%d", len(rows), len(wantTitles))
			}
			for i, title := range wantTitles {
				if title == "" {
					if _, ok := rows[i]["title"]; ok {
						t.Fatalf("comment title leaked: %v", rows[i])
					}
				} else if rows[i]["title"] != title {
					t.Errorf("read %d title=%v want=%s", i, rows[i]["title"], title)
				}
			}
			if len(reader.requests) != 1 {
				t.Fatalf("current context requests=%d want=1", len(reader.requests))
			}
		})
	}
}
