package riidoaiserver

import (
	"context"
	"testing"
)

func TestCoverageEnrollmentAtomicFailureAndLostACK(t *testing.T) {
	for _, mode := range []string{"reject", "lost_ack", "lost_ack_missing_policy"} {
		t.Run(mode, func(t *testing.T) {
			fixture := newMembershipSharedDynamoFixture(t)
			endpoint := coverageEnrollmentBoundary(t, fixture, coverageEnrollmentIntercept(t, fixture, mode))
			store := openCoverageEnrollmentFixture(t, endpoint, true)
			principal := profileColorTestPrincipal()
			response, err := store.EnrollDeviceCredential(context.Background(), principal, principal.WorkspaceID, EnrollDeviceRequest{MachineID: "atomic-" + mode, ExecutionCoverageProtocol: executionCoverageProtocol})
			if mode == "lost_ack" {
				if err != nil || response.ExecutionCoverageGeneration == "" {
					t.Fatal("matching committed credential and coverage did not recover lost ACK")
				}
				assertEnrollmentCredential(t, store, response, true)
				return
			}
			if err == nil || response.DeviceSecret != "" {
				t.Fatal("uncommitted or unproven coverage returned credential")
			}
			if mode == "reject" {
				fresh := openCoverageEnrollmentFixture(t, fixture.server.URL, false)
				for _, credential := range fresh.deviceCredentials {
					if credential.machineID == "atomic-reject" {
						t.Fatal("failed birth published credential without coverage")
					}
				}
				fixture.mu.Lock()
				defer fixture.mu.Unlock()
				for _, item := range fixture.items {
					if dynamoDBStringValue(item, "policy_schema") == executionCoverageSchema {
						t.Fatal("failed transaction persisted orphan coverage")
					}
				}
			}
		})
	}
}
