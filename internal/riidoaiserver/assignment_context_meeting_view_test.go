package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func meetingContextFixture(t *testing.T, status string) AIAgentTaskContext {
	t.Helper()
	c := assignmentContextTestTaskContext()
	body := `{"related_status":"available","related_meetings":[{"id":"meeting-1","title":"Linked meeting","body_status":"` + status + `","document":{"content":"MEETING_BODY_ONLY_MARKER","contentFormat":"markdown","status":"` + status + `","source":"provider_json_projection"}}]}`
	if err := json.Unmarshal([]byte(body), &c.Extended); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestMeetingBodySelectedOnlyByExplicitView(t *testing.T) {
	c := meetingContextFixture(t, "available")
	for _, view := range []string{"base", "resources", "tree", "meetings"} {
		data, sources, ok := selectAssignmentContextView(c, view, "")
		encoded, _ := json.Marshal(data)
		if !ok || strings.Contains(string(encoded), "MEETING_BODY_ONLY_MARKER") {
			t.Fatalf("view %s exposed a meeting body", view)
		}
		for _, source := range sources {
			if source.Kind == "related_meeting" && source.Part == "body" {
				t.Fatalf("view %s recorded a body read", view)
			}
		}
	}
	data, sources, ok := selectAssignmentContextView(c, "meeting", "meeting-1")
	encoded, _ := json.Marshal(data)
	if !ok || !strings.Contains(string(encoded), "MEETING_BODY_ONLY_MARKER") || len(sources) != 1 || sources[0].Part != "body" || sources[0].Version == "" {
		t.Fatal("explicit meeting read did not serve the captured body and version")
	}
	if _, _, ok := selectAssignmentContextView(c, "meeting", "not-related"); ok {
		t.Fatal("outside meeting was accepted")
	}
}

func TestMeetingBodyHTTPReadRecordsOnlyExplicitBody(t *testing.T) {
	store := NewMemoryAssignmentContextStore()
	f := newAssignmentContextFixtureWith(t, store, meetingContextFixture(t, "available"))
	for _, view := range []string{"base", "resources", "meetings"} {
		code, body := f.read(view, "", f.leased.LeaseToken)
		if code != http.StatusOK || strings.Contains(body, "MEETING_BODY_ONLY_MARKER") {
			t.Fatalf("%s status=%d", view, code)
		}
	}
	code, body := f.read("meeting", "meeting-1", f.leased.LeaseToken)
	if code != http.StatusOK || !strings.Contains(body, "MEETING_BODY_ONLY_MARKER") {
		t.Fatalf("meeting status=%d", code)
	}
	reads, err := store.ListAssignmentContextReads(t.Context(), f.leased.ID)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, read := range reads {
		if read.SourceKind == "related_meeting" && read.SourcePart == "body" {
			count++
			if read.Via != "cli" {
				t.Fatal("prompt recorded unread meeting body")
			}
		}
	}
	if count != 1 {
		t.Fatalf("body reads=%d", count)
	}
}
