package riidoaiserver

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func TestLogAssignmentFailureIsBounded(t *testing.T) {
	var output bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&output)
	t.Cleanup(func() { log.SetOutput(previous) })

	state := &storeState{assignments: map[string]Assignment{
		"asn-1": {ID: "asn-1", RuntimeProvider: "codex"},
	}}
	logAssignmentFailure(state, TaskEvent{AssignmentID: "asn-1", Type: EventAssignmentFailed, Metadata: map[string]string{
		metadatakeys.AssignmentResultStatus.String():    "blocked",
		metadatakeys.AssignmentFailureCategory.String(): "bearer-secret /Users/private",
	}})

	got := output.String()
	for _, want := range []string{`event=assignment_failed`, `provider="codex"`, `result_status="blocked"`, `failure_category="unknown"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("log %q does not contain %q", got, want)
		}
	}
	if strings.Contains(got, "bearer-secret") || strings.Contains(got, "/Users") {
		t.Fatalf("log leaked raw metadata: %q", got)
	}
}
