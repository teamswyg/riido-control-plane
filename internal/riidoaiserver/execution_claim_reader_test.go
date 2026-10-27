package riidoaiserver

import (
	"context"
	"testing"
)

func TestExecutionClaimReaderFollowsStrongHistoryPagesAndRejectsIncompleteHistory(t *testing.T) {
	for _, mode := range []string{"complete", "duplicate", "cursor cycle", "foreign cursor", "foreign item", "invalid epoch"} {
		t.Run(mode, func(t *testing.T) {
			f := newContextRetentionFixture(t)
			claim := executionClaim{DeviceID: "device", OwnerPrincipalID: "owner", RequesterPrincipalID: "requester", AssignmentID: "a", AgentID: "agent", LeaseEpoch: leaseEpoch("fixture"), ClaimedAt: f.now}
			first, err := executionClaimItem(claim)
			if err != nil {
				t.Fatal(err)
			}
			claim.AssignmentID = "b"
			second, err := executionClaimItem(claim)
			if err != nil {
				t.Fatal(err)
			}
			cursor := map[string]map[string]string{"pk": first["pk"], "sk": first["sk"]}
			f.queryPages = []contextRetentionQueryPage{{Items: []map[string]map[string]string{first}, LastEvaluatedKey: cursor}, {Items: []map[string]map[string]string{second}}}
			switch mode {
			case "duplicate":
				f.queryPages[1].Items = []map[string]map[string]string{first}
			case "cursor cycle":
				f.queryPages[1] = contextRetentionQueryPage{LastEvaluatedKey: cursor}
			case "foreign cursor":
				cursor["pk"] = map[string]string{"S": "DEVICE_EXECUTION#other"}
			case "foreign item":
				second["pk"] = map[string]string{"S": "DEVICE_EXECUTION#other"}
			case "invalid epoch":
				second["payload"] = map[string]string{"S": `{"device_id":"device","lease_epoch":"raw-token"}`}
			}
			claims, err := f.store.loadDeviceExecutionClaims(context.Background(), "device")
			if mode == "complete" {
				if err != nil || len(claims) != 2 {
					t.Fatalf("complete pages: count=%d err=%v", len(claims), err)
				}
			} else if err == nil || claims != nil {
				t.Fatal("incomplete or invalid history accepted")
			}
		})
	}
}
