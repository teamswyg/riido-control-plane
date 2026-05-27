package riidoaiserver

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-contracts/hostintegration"
)

func TestStoreSafeRoutingAllowsAvailableProvider(t *testing.T) {
	decision, err := EvaluateStoreSafeRouting(StoreSafeRoutingInput{
		RuntimeProvider: " codex ",
		ProviderStatuses: []ProviderStatusRecord{{
			ProviderKind:  "codex",
			RoutingStatus: hostintegration.ProviderRoutingAvailable,
		}},
	})
	if err != nil {
		t.Fatalf("routing decision failed: %v", err)
	}
	if !decision.Allowed || decision.Reason != "provider available" || decision.ProviderKind != "codex" {
		t.Fatalf("available provider should be assignable: %+v", decision)
	}
}

func TestStoreSafeRoutingBlocksUnavailableProviderStatuses(t *testing.T) {
	cases := map[hostintegration.ProviderRoutingStatus]string{
		hostintegration.ProviderRoutingLoginRequired: "provider login required",
		hostintegration.ProviderRoutingUnsupported:   "provider unsupported",
		hostintegration.ProviderRoutingStoreBlocked:  "provider blocked by store policy",
	}
	for status, reason := range cases {
		t.Run(string(status), func(t *testing.T) {
			decision, err := EvaluateStoreSafeRouting(StoreSafeRoutingInput{
				RuntimeProvider: "codex",
				ProviderStatuses: []ProviderStatusRecord{{
					ProviderKind:  "codex",
					RoutingStatus: status,
				}},
			})
			if err != nil {
				t.Fatalf("routing decision for %s failed: %v", status, err)
			}
			if decision.Allowed || decision.Reason != reason {
				t.Fatalf("%s provider should not be assignable: %+v", status, decision)
			}
		})
	}
}

func TestStoreSafeRoutingBlocksMissingSyncedProvider(t *testing.T) {
	decision, err := EvaluateStoreSafeRouting(StoreSafeRoutingInput{
		RuntimeProvider: "codex",
		ProviderStatuses: []ProviderStatusRecord{{
			ProviderKind:  "claude",
			RoutingStatus: hostintegration.ProviderRoutingAvailable,
		}},
	})
	if err != nil {
		t.Fatalf("routing decision failed: %v", err)
	}
	if decision.Allowed || decision.Reason != "provider status missing" {
		t.Fatalf("missing provider should be blocked: %+v", decision)
	}
}

func TestStoreSafeRoutingKeepsLegacyAssignmentWhenNoStatusWasSynced(t *testing.T) {
	decision, err := EvaluateStoreSafeRouting(StoreSafeRoutingInput{RuntimeProvider: "codex"})
	if err != nil {
		t.Fatalf("routing decision failed: %v", err)
	}
	if !decision.Allowed || decision.Reason != "provider status not synced" {
		t.Fatalf("missing status snapshot should preserve existing assignment behavior: %+v", decision)
	}
}

func TestStoreSafeRoutingRejectsInvalidInput(t *testing.T) {
	_, err := EvaluateStoreSafeRouting(StoreSafeRoutingInput{RuntimeProvider: " "})
	if err == nil || !strings.Contains(err.Error(), "runtime_provider is required") {
		t.Fatalf("missing runtime provider err=%v", err)
	}
	_, err = EvaluateStoreSafeRouting(StoreSafeRoutingInput{
		RuntimeProvider: "codex",
		ProviderStatuses: []ProviderStatusRecord{{
			ProviderKind:  "codex",
			RoutingStatus: "unknown",
		}},
	})
	if err == nil || !strings.Contains(err.Error(), "unknown provider routing status") {
		t.Fatalf("unknown routing status err=%v", err)
	}
}
