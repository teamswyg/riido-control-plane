package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestTaskReadScopeMeetingStableMetadataRetainsAuthorityAndDetails(t *testing.T) {
	first := taskReadMeeting{ID: "meeting", TeamID: "team", Status: "partial", Details: json.RawMessage(`{"id":"meeting","title":"Original"}`), Labels: []json.RawMessage{json.RawMessage(`{"id":"label"}`)}}
	want, err := readScopeMeetingStableMetadata(first)
	if err != nil {
		t.Fatal(err)
	}
	completed := first
	completed.Status = "available"
	got, err := readScopeMeetingStableMetadata(completed)
	if err != nil || !bytes.Equal(got, want) {
		t.Fatal("aggregate pagination status changed stable metadata")
	}
	for name, mutate := range map[string]func(*taskReadMeeting){
		"id":      func(v *taskReadMeeting) { v.ID = "other" },
		"team":    func(v *taskReadMeeting) { v.TeamID = "other" },
		"details": func(v *taskReadMeeting) { v.Details = json.RawMessage(`{"id":"meeting","title":"Changed"}`) },
		"labels":  func(v *taskReadMeeting) { v.Labels = []json.RawMessage{json.RawMessage(`{"id":"other"}`)} },
	} {
		t.Run(name, func(t *testing.T) {
			changed := completed
			mutate(&changed)
			actual, err := readScopeMeetingStableMetadata(changed)
			if err != nil || bytes.Equal(actual, want) {
				t.Fatal("non-pagination metadata change was ignored")
			}
		})
	}
}
