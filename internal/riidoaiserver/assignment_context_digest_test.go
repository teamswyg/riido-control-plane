package riidoaiserver

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestContextDigestTracksAllCapturedContent(t *testing.T) {
	resource := AssignmentScopedResource{Kind: AssignmentContentComponent, ID: "work", TeamID: "team", Title: "Current title"}
	part := AssignmentContentManifest{Part: AssignmentContentComments, Version: "parent-v1", Complete: true, Status: "available"}
	pages := []AssignmentContentPage{{Offset: 0, Data: json.RawMessage(`{"id":9007199254740993,"body":"parent"}`)}, {Offset: 1, Complete: true, Data: json.RawMessage(`{"reply":"before"}`)}}
	digest := func(r AssignmentScopedResource, p AssignmentContentManifest, values []AssignmentContentPage) string {
		t.Helper()
		out, err := assignmentCapturedPartDigest(r, p, values)
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	want := digest(resource, part, pages)
	reordered := append([]AssignmentContentPage(nil), pages...)
	reordered[0].Data = json.RawMessage(` { "body": "parent", "id": 9007199254740993 } `)
	reordered[0].Cursor = "random-delivery-cursor"
	part.Cursors, part.NextAction = []string{"different"}, "different command"
	if digest(resource, part, reordered) != want {
		t.Fatal("format or delivery metadata changed digest")
	}
	for name, raw := range map[string]string{"reply edit": `{"reply":"after"}`, "large integer": `{"id":9007199254740992,"body":"parent"}`} {
		t.Run(name, func(t *testing.T) {
			changed := append([]AssignmentContentPage(nil), pages...)
			i := 1
			if name == "large integer" {
				i = 0
			}
			changed[i].Data = json.RawMessage(raw)
			if digest(resource, part, changed) == want {
				t.Fatal("content change missed with same parent version")
			}
		})
	}
	resource.Title = "Renamed"
	if digest(resource, part, pages) == want {
		t.Fatal("title change missed")
	}
	for _, raw := range []string{`{"x":1,"x":2}`, `{} {}`, strings.Repeat("[", 130) + "0" + strings.Repeat("]", 130)} {
		if _, err := canonicalCapturedJSON(json.RawMessage(raw)); err == nil {
			t.Fatal("ambiguous or excessive JSON accepted")
		}
	}
}

func TestContextDigestMissingPartUsesFinalStatus(t *testing.T) {
	b, err := newTaskReadScopeBuilder()
	if err != nil {
		t.Fatal(err)
	}
	resource := AssignmentScopedResource{Kind: AssignmentContentMeeting, ID: "meeting", TeamID: "team"}
	part := b.missingPart(resource, AssignmentContentBoard, "unavailable")
	want, err := assignmentCapturedPartDigest(resource, part, nil)
	if err != nil || part.Digest != want || b.capture.Manifest[0].Digest != want {
		t.Fatal("digest does not describe final captured status")
	}
}
