package riidoaiserver

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAIAgentRunReadsMetadataStaysOutOfContextReadEvent(t *testing.T) {
	read := AssignmentContextRead{SourceKind: "task", SourceID: "task", SourcePart: "body", SourceUpdatedAt: "2026-10-26T01:00:00Z"}
	server, reader, _ := runReadsTestServer(t, []AssignmentContextRead{read})
	reader.value.Component.Title = "Current title"
	reader.value.Component.UpdatedAt = "2026-10-27T01:00:00Z"
	rows := runReadsMetadataRows(t, server, "/v1/client/ai-agent/runs/run/reads")
	if len(rows) != 1 || rows[0]["title"] != "Current title" || rows[0]["modified_since_read"] != true {
		t.Fatalf("GET metadata missing: %v", rows)
	}
	raw, err := json.Marshal(newAgentContextReadEvent([]AssignmentContextRead{read}))
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"title", "modified_since_read"} {
		if strings.Contains(string(raw), `"`+field+`"`) {
			t.Fatalf("GET-only field %s leaked into SSE: %s", field, raw)
		}
	}
}
