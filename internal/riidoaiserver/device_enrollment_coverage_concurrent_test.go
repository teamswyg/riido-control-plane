package riidoaiserver

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestCoverageEnrollmentConcurrentSameMachineHasOneBirth(t *testing.T) {
	fixture := newMembershipSharedDynamoFixture(t)
	var arriving atomic.Int32
	both := make(chan struct{})
	endpoint := coverageEnrollmentBoundary(t, fixture, func(_ http.ResponseWriter, r *http.Request, body []byte) bool {
		if !coverageBirthRequest(r, body) {
			return false
		}
		switch arriving.Add(1) {
		case 1:
			select {
			case <-both:
			case <-r.Context().Done():
			}
		case 2:
			close(both)
		}
		return false
	})
	first := openCoverageEnrollmentFixture(t, endpoint, true)
	second := openCoverageEnrollmentFixture(t, endpoint, true)
	principal := profileColorTestPrincipal()
	type result struct {
		response EnrollDeviceResponse
		err      error
	}
	results := make(chan result, 2)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, store := range []*PersistentAIAgentClientStore{first, second} {
		go func() {
			response, err := store.EnrollDeviceCredential(ctx, principal, principal.WorkspaceID, EnrollDeviceRequest{MachineID: "same-covered-machine", ExecutionCoverageProtocol: executionCoverageProtocol})
			results <- result{response, err}
		}()
	}
	a, b := <-results, <-results
	if (a.err == nil) == (b.err == nil) {
		t.Fatal("concurrent proposed namespaces did not admit exactly one initial birth")
	}
	fresh := openCoverageEnrollmentFixture(t, fixture.server.URL, false)
	count := 0
	for _, record := range fresh.deviceCredentials {
		if record.machineID == "same-covered-machine" {
			count++
		}
	}
	if count != 1 {
		t.Fatal("same machine published multiple credential namespaces")
	}
	fixture.mu.Lock()
	defer fixture.mu.Unlock()
	count = 0
	for _, item := range fixture.items {
		if dynamoDBStringValue(item, "policy_schema") == executionCoverageSchema {
			count++
		}
	}
	if count != 1 {
		t.Fatal("concurrent failure left orphan policy birth")
	}
}
