package riidoaiserver

import (
	"context"
	"strings"
	"testing"
)

func TestDeviceBootstrapStalePreparationCannotReplaceRotatedAuthority(t *testing.T) {
	f, ctx := newBootstrapFixture(t), context.Background()
	r, err := f.store.beginDeviceBootstrap(ctx, f.request)
	if err != nil {
		t.Fatal(err)
	}
	other := openCoverageEnrollmentFixture(t, f.dynamo.server.URL, false)
	principal := profileColorTestPrincipal()
	rotated, err := other.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, EnrollDeviceRequest{MachineID: "bootstrap-machine"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.store.advanceDeviceBootstrap(ctx, r.Request.TransitionID, "prepare", f.auth(), f.proof(t, r, "prepare", strings.Repeat("a", 64))); err == nil {
		t.Fatal("stale prepare fenced a later native enrollment")
	}
	fresh := openCoverageEnrollmentFixture(t, f.dynamo.server.URL, false)
	assertEnrollmentCredential(t, fresh, rotated, true)
	assertEnrollmentCredential(t, fresh, f.credential, false)
	f.dynamo.mu.Lock()
	item := f.dynamo.items["DEVICE_POLICY#"+r.Before.DeviceID+"\x00REVISION"]
	f.dynamo.mu.Unlock()
	if dynamoDBStringValue(item, "policy_schema") == deviceBootstrapFenceSchema {
		t.Fatal("failed prepare left an orphan fence")
	}
}
