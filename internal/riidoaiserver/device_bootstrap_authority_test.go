package riidoaiserver

import (
	"context"
	"strings"
	"testing"
)

func TestDeviceBootstrapRejectsAlteredAuthority(t *testing.T) {
	for _, change := range []string{"owner", "intent", "boot", "signature", "source", "next_hash"} {
		t.Run(change, func(t *testing.T) {
			f, ctx := newBootstrapFixture(t), context.Background()
			r, err := f.store.beginDeviceBootstrap(ctx, f.request)
			if err != nil {
				t.Fatal(err)
			}
			proof := f.proof(t, r, "prepare", strings.Repeat("a", 64))
			auth := f.auth()
			switch change {
			case "owner":
				auth.ownerPrincipalID = "other-owner"
			case "source":
				auth.deviceID = "dev_0000000000000000"
			case "intent":
				proof.IntentID = "wrong"
			case "boot":
				proof.BootDigest = strings.Repeat("b", 64)
			case "signature":
				proof.Signature = "invalid"
			case "next_hash":
				request := f.request
				request.NextSecretHash = strings.Repeat("c", 64)
				if _, err := f.store.beginDeviceBootstrap(ctx, request); err == nil {
					t.Fatal("same transition accepted different secret hash")
				}
				return
			}
			if _, err := f.store.advanceDeviceBootstrap(ctx, r.Request.TransitionID, "prepare", auth, proof); err == nil {
				t.Fatal("altered authority accepted")
			}
			assertEnrollmentCredential(t, f.store, f.credential, true)
		})
	}
}
