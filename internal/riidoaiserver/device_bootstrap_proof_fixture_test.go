package riidoaiserver

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"
)

func (f *bootstrapFixture) proof(t *testing.T, r deviceBootstrapRecord, phase, boot string) deviceBootstrapProof {
	t.Helper()
	digest := sha256.Sum256(deviceBootstrapMessage(r, phase, boot))
	signature, err := ecdsa.SignASN1(rand.Reader, f.key, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	return deviceBootstrapProof{IntentID: r.Intent.ID, BootDigest: boot, Signature: base64.RawURLEncoding.EncodeToString(signature)}
}

func (f *bootstrapFixture) prepare(t *testing.T) deviceBootstrapRecord {
	t.Helper()
	r, err := f.store.beginDeviceBootstrap(context.Background(), f.request)
	if err != nil {
		t.Fatal(err)
	}
	r, err = f.store.advanceDeviceBootstrap(context.Background(), r.Request.TransitionID, "prepare", f.auth(), f.proof(t, r, "prepare", strings.Repeat("a", 64)))
	if err != nil {
		t.Fatal(err)
	}
	return r
}
