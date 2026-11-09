package riidoaiserver

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestDeviceBootstrapPreservesIdentityAndFencesLegacyEnrollment(t *testing.T) {
	f, ctx := newBootstrapFixture(t), context.Background()
	r := f.prepare(t)
	if r.State != "prepared" {
		t.Fatal("missing durable preparation")
	}
	assertEnrollmentCredential(t, f.store, f.credential, true)
	principal := profileColorTestPrincipal()
	if _, err := f.store.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, EnrollDeviceRequest{MachineID: "bootstrap-machine"}); err == nil {
		t.Fatal("legacy rotation bypassed prepared fence")
	}
	if _, err := f.store.advanceDeviceBootstrap(ctx, r.Request.TransitionID, "activate", f.auth(), f.proof(t, r, "activate", r.PreviousBootDigest)); err == nil {
		t.Fatal("same boot activated coverage")
	}
	completed, err := f.store.advanceDeviceBootstrap(ctx, r.Request.TransitionID, "activate", f.auth(), f.proof(t, r, "activate", strings.Repeat("b", 64)))
	if err != nil {
		t.Fatal(err)
	}
	if completed.State != "complete" || completed.Before.DeviceID != completed.After.DeviceID || completed.Before.MachineID != completed.After.MachineID || completed.Before.OwnerPrincipalID != completed.After.OwnerPrincipalID {
		t.Fatal("bootstrap replaced stable identity")
	}
	assertEnrollmentCredential(t, f.store, f.credential, false)
	updated := f.credential
	updated.DeviceSecret = f.nextSecret
	assertEnrollmentCredential(t, f.store, updated, true)
	fresh := openCoverageEnrollmentFixture(t, f.dynamo.server.URL, false)
	replayed, err := fresh.advanceDeviceBootstrap(ctx, r.Request.TransitionID, "activate", bootstrapAuthFor(completed.After), deviceBootstrapProof{})
	if err != nil || !reflect.DeepEqual(replayed, completed) {
		t.Fatal("restart did not return exact committed receipt", err)
	}
	rotated, err := fresh.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, EnrollDeviceRequest{MachineID: "bootstrap-machine"})
	if err != nil || rotated.ExecutionCoverageGeneration != completed.After.ExecutionCoverageGeneration {
		t.Fatal("covered legacy-ID rotation lost generation", err)
	}
	if _, err := fresh.advanceDeviceBootstrap(ctx, r.Request.TransitionID, "activate", bootstrapAuthFor(completed.After), deviceBootstrapProof{}); err == nil {
		t.Fatal("receipt replay resurrected a rotated credential")
	}
}
