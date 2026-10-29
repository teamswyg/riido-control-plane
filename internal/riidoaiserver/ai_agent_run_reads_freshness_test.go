package riidoaiserver

import "testing"

func TestAIAgentRunReadsMetadataFreshness(t *testing.T) {
	const before = "2026-10-26T01:00:00Z"
	const after = "2026-10-27T01:00:00Z"
	for _, test := range []struct {
		name, part, previousBody, currentBody, previousTime, currentTime, readStatus, documentStatus string
		want                                                                                         any
	}{
		{name: "same body", part: "body", previousBody: "body", currentBody: "body", want: false},
		{name: "different body without historical completeness", part: "body", previousBody: "old", currentBody: "new"},
		{name: "newer timestamp", part: "title", previousTime: before, currentTime: after, want: true},
		{name: "same timestamp alone is inconclusive", part: "title", previousTime: before, currentTime: "2026-10-26T10:00:00+09:00"},
		{name: "older timestamp", part: "title", previousTime: after, currentTime: before},
		{name: "invalid previous timestamp", part: "title", previousTime: "bad", currentTime: after},
		{name: "invalid current timestamp", part: "title", previousTime: before, currentTime: "bad"},
		{name: "unknown version", part: "body", currentBody: "body"},
		{name: "historical snapshot bounded", part: "body", previousBody: "bound", currentBody: "bounded body", readStatus: "truncated"},
		{name: "current partial body", part: "body", previousBody: "old", currentBody: "new", documentStatus: "partial"},
		{name: "legacy served partial became full", part: "body", previousBody: "partial", currentBody: "partial content now complete", readStatus: "served"},
		{name: "newer time proves change despite truncation", part: "body", previousBody: "bound", currentBody: "bounded body", readStatus: "truncated", previousTime: before, currentTime: after, want: true},
		{name: "same body newer time", part: "body", previousBody: "body", currentBody: "body", previousTime: before, currentTime: after, want: true},
		{name: "different body same time is inconclusive", part: "body", previousBody: "old", currentBody: "new", previousTime: before, currentTime: before},
		{name: "denied is not a read", part: "body", previousBody: "old", currentBody: "new", readStatus: "denied"},
		{name: "unavailable is not a read", part: "body", previousBody: "old", currentBody: "new", readStatus: "unavailable"},
	} {
		t.Run(test.name, func(t *testing.T) {
			read := AssignmentContextRead{SourceKind: "task", SourceID: "task", SourcePart: test.part, SourceVersion: assignmentContextVersion(test.previousBody), SourceUpdatedAt: test.previousTime}
			server, reader, _ := runReadsTestServer(t, []AssignmentContextRead{read})
			reader.value.Component.UpdatedAt = test.currentTime
			reader.value.Document = AIAgentTaskContextDocument{Content: test.currentBody, Status: test.documentStatus}
			if test.readStatus != "" {
				server.config.AssignmentContext.(*MemoryAssignmentContextStore).reads["assignment"][0].Status = test.readStatus
			}
			rows := runReadsMetadataRows(t, server, "/v1/client/ai-agent/runs/run/reads")
			if len(rows) != 1 || rows[0]["modified_since_read"] != test.want {
				t.Fatalf("reads=%v want modified_since_read=%v", rows, test.want)
			}
		})
	}
}

func TestAIAgentRunReadsMetadataUsesCapturedSourceHashes(t *testing.T) {
	context := AIAgentTaskContext{
		Hierarchy: AIAgentTaskContextHierarchy{Project: AIAgentTaskContextReference{ID: "project", Document: &AIAgentTaskContextDocument{Content: "same project body"}}},
		Extended: &AIAgentTaskContextExtended{
			Comments:        []AIAgentTaskContextComment{{ID: "comment", Body: "same comment"}},
			RelatedMeetings: []AIAgentTaskContextRelatedMeeting{{ID: "meeting", BodyStatus: "available", Document: &AIAgentTaskContextDocument{Content: "same meeting", Status: "available"}}},
		},
	}
	project := assignmentContextReferenceSources("project", context.Hierarchy.Project)[0]
	comment := assignmentContextCommentSources(context.Extended.Comments)[0]
	_, meetingSources, _ := selectAssignmentContextMeeting(context, "meeting")
	meeting := meetingSources[0]
	server, reader, _ := runReadsTestServer(t, []AssignmentContextRead{
		{SourceKind: "project", SourceID: "project", SourcePart: "summary", SourceVersion: project.Version},
		{SourceKind: "comment", SourceID: "comment", SourcePart: "body", SourceVersion: comment.Version},
		{SourceKind: "related_meeting", SourceID: "meeting", SourcePart: "body", SourceVersion: meeting.Version},
	})
	reader.value.Hierarchy, reader.value.Extended = context.Hierarchy, context.Extended
	rows := runReadsMetadataRows(t, server, "/v1/client/ai-agent/runs/run/reads")
	for _, row := range rows {
		if row["modified_since_read"] != false {
			t.Errorf("equal source hashes must be comparable: %v", row)
		}
	}
	reader.value.Hierarchy.Project.Document.Content = "changed project"
	reader.value.Extended.Comments[0].Body = "changed comment"
	reader.value.Extended.RelatedMeetings[0].Document.Content = "changed meeting"
	for _, row := range runReadsMetadataRows(t, server, "/v1/client/ai-agent/runs/run/reads") {
		if _, exists := row["modified_since_read"]; exists {
			t.Errorf("different hashes without historical completeness must stay unknown: %v", row)
		}
	}
}
