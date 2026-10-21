package riidoaiserver

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestEnrollmentCredentialOwnershipPreservesInlineLegacyRows(t *testing.T) {
	committed := snapshotTestRecord(fixedSnapshotTestNow())
	committed.DeviceCredentials = []AIAgentClientDeviceCredentialSnapshot{{DeviceID: "legacy-device", SecretHashSHA256: strings.Repeat("a", 64), OwnerPrincipalID: "owner", WorkspaceID: "workspace"}}
	raw, err := json.Marshal(committed)
	if err != nil {
		t.Fatal(err)
	}
	current := map[string]map[string]string{"pk": {"S": dynamoDBAIAgentClientSnapshotPK}, "sk": {"S": dynamoDBAIAgentClientSnapshotSK}, "snapshot_json": {"S": string(raw)}}
	requested := committed
	requested.DeviceCredentials = nil
	if err := mergeCommittedEnrollmentState(context.Background(), []map[string]map[string]string{current}, &requested); err != nil {
		t.Fatal(err)
	}
	if !sameEnrollmentCredential(requested, committed, "legacy-device") {
		t.Fatal("ordinary save discarded legacy credential")
	}
	ctx := withDeviceEnrollmentMutation(context.Background(), committed, "legacy-device")
	requested.DeviceCredentials[0].SecretHashSHA256 = strings.Repeat("b", 64)
	if err := mergeCommittedEnrollmentState(ctx, []map[string]map[string]string{current}, &requested); err != nil {
		t.Fatal(err)
	}
	if requested.DeviceCredentials[0].SecretHashSHA256 != strings.Repeat("b", 64) {
		t.Fatal("scoped enrollment could not rotate legacy credential")
	}
}
