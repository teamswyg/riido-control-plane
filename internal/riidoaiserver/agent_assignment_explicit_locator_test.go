package riidoaiserver

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestExplicitWorkspaceProjectReadsAndAudits(t *testing.T) {
	for _, kind := range []string{"project", "milestone"} {
		t.Run(kind, func(t *testing.T) {
			f := explicitReadFixture(t, func(_ *assignmentContextFixture, v *explicitAssignmentResult, _ http.ResponseWriter, _ *http.Request) bool {
				v.SchemaVersion, v.Kind, v.TeamID = "riido-agent-explicit-read.v2", kind, nil
				return true
			})
			code, body := f.read("task", "target-one", f.leased.LeaseToken)
			var out AgentAssignmentContextResponse
			if code != 200 || json.Unmarshal([]byte(body), &out) != nil || len(out.Sources) != 1 || out.Sources[0].ID != "target-one" {
				t.Fatalf("status=%d body=%s", code, body)
			}
			reads, err := f.contexts.ListAssignmentContextReads(t.Context(), f.leased.ID)
			if err != nil || len(reads) == 0 || reads[len(reads)-1].RequesterPrincipalID != f.leased.CreatedBy {
				t.Fatal("original requester lost", err)
			}
		})
	}
}

func TestExplicitLocatorRejectsAmbiguousTeamContract(t *testing.T) {
	for _, raw := range []string{
		`{"schema_version":"riido-agent-explicit-read.v2","kind":"project"}`,
		`{"schema_version":"riido-agent-explicit-read.v2","kind":"task","team_id":null}`,
		`{"schema_version":"riido-agent-explicit-read.v2","kind":"project","team_id":"fake-team"}`,
		`{"schema_version":"riido-agent-explicit-read.v1","kind":"project","team_id":null}`,
		`{"schema_version":"riido-agent-explicit-read.v3","kind":"project","team_id":null}`,
	} {
		var v explicitAssignmentResult
		if json.Unmarshal([]byte(raw), &v) != nil || validExplicitAssignmentLocator([]byte(raw), v) {
			t.Fatal("invalid team contract accepted", raw)
		}
	}
}
