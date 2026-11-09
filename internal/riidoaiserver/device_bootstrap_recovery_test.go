package riidoaiserver

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestDeviceBootstrapExpiredIntentRenewsWithoutUnfencing(t *testing.T) {
	f, ctx := newBootstrapFixture(t), context.Background()
	r := f.prepare(t)
	oldProof := f.proof(t, r, "activate", strings.Repeat("b", 64))
	r.Intent.ExpiresAt = time.Now().UTC().Add(-time.Minute)
	action, err := deviceBootstrapRecordAction("fixture", deviceBootstrapMutation{after: r})
	if err != nil {
		t.Fatal(err)
	}
	f.dynamo.mu.Lock()
	f.dynamo.items[membershipDynamoKey(action.Put.Item)] = action.Put.Item
	f.dynamo.mu.Unlock()
	if _, err := f.store.advanceDeviceBootstrap(ctx, r.Request.TransitionID, "activate", f.auth(), f.proof(t, r, "activate", strings.Repeat("b", 64))); !errors.Is(err, errCLIInstallationExpired) {
		t.Fatal("expired intent accepted", err)
	}
	renewed, err := f.store.renewDeviceBootstrap(ctx, r.Request.TransitionID, f.auth())
	if err != nil {
		t.Fatal(err)
	}
	if renewed.Intent.ID == r.Intent.ID || renewed.PreviousBootDigest != r.PreviousBootDigest || renewed.PolicyRevision != r.PolicyRevision || renewed.State != "prepared" || !reflect.DeepEqual(renewed.Before, r.Before) || !reflect.DeepEqual(renewed.After, r.After) {
		t.Fatal("renewal changed prepared authority")
	}
	principal := profileColorTestPrincipal()
	if _, err := f.store.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, EnrollDeviceRequest{MachineID: "bootstrap-machine"}); err == nil {
		t.Fatal("renewal unfenced ordinary enrollment")
	}
	if _, err := f.store.advanceDeviceBootstrap(ctx, r.Request.TransitionID, "activate", f.auth(), oldProof); err == nil {
		t.Fatal("renewal accepted retired challenge")
	}
	if _, err := f.store.advanceDeviceBootstrap(ctx, r.Request.TransitionID, "activate", f.auth(), f.proof(t, renewed, "activate", strings.Repeat("b", 64))); err != nil {
		t.Fatal(err)
	}
}
