package riidoaiserver

import (
	"context"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestDeviceBootstrapConcurrentActivationHasOneAuthority(t *testing.T) {
	f, ctx := newBootstrapFixture(t), context.Background()
	r := f.prepare(t)
	writers := []*PersistentAIAgentClientStore{openCoverageEnrollmentFixture(t, f.dynamo.server.URL, false), openCoverageEnrollmentFixture(t, f.dynamo.server.URL, false)}
	proofs := []deviceBootstrapProof{f.proof(t, r, "activate", strings.Repeat("b", 64)), f.proof(t, r, "activate", strings.Repeat("c", 64))}
	results := make([]deviceBootstrapRecord, 2)
	errs := make([]error, 2)
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := range writers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			results[i], errs[i] = writers[i].advanceDeviceBootstrap(ctx, r.Request.TransitionID, "activate", f.auth(), proofs[i])
		}(i)
	}
	close(start)
	wg.Wait()
	fresh := openCoverageEnrollmentFixture(t, f.dynamo.server.URL, false)
	committed, err := fresh.loadDeviceBootstrap(ctx, r.Request.TransitionID, bootstrapAuthFor(r.After))
	if err != nil || committed.State != "complete" {
		t.Fatal("concurrent activation did not commit one authority", err)
	}
	successes := 0
	for i := range results {
		if errs[i] == nil {
			successes++
			if !reflect.DeepEqual(results[i], committed) {
				t.Fatal("concurrent response invented another completion")
			}
		}
	}
	if successes == 0 {
		t.Fatal("all activation writers failed")
	}
	assertEnrollmentCredential(t, fresh, f.credential, false)
	updated := f.credential
	updated.DeviceSecret = f.nextSecret
	assertEnrollmentCredential(t, fresh, updated, true)
}
