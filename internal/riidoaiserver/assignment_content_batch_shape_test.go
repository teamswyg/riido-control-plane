package riidoaiserver

import (
	"net/http"
	"strings"
	"testing"
)

func TestAssignmentContentBatchAlwaysReturnsTheRequestedPage(t *testing.T) {
	f := newAssignmentContentBatchFixture(t)
	status, body := readAssignmentContentBatch(t, f, "body-5", 8<<10)
	if status != http.StatusOK || !strings.Contains(body, `"cursor":"body-5"`) || strings.Contains(body, "comments-0") {
		t.Fatalf("status=%d body=%s", status, body)
	}
}

func TestAssignmentContentLegacySinglePageShapeIsUnchanged(t *testing.T) {
	f := newAssignmentContentBatchFixture(t)
	status, body := readAssignmentContentBatch(t, f, "body-0", 0)
	if status != http.StatusOK || !strings.Contains(body, `"cursor":"body-0"`) || !strings.Contains(body, `"next_cursor":"body-1"`) || strings.Contains(body, `"pages"`) {
		t.Fatalf("status=%d body=%s", status, body)
	}
}
