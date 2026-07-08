package riidoaiserver

import (
	"testing"

	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-contracts/provider/capability"
)

func TestCloneProviderStatusRecordsEmpty(t *testing.T) {
	t.Parallel()
	if got := cloneProviderStatusRecords(nil); got != nil {
		t.Fatalf("nil clone = %+v, want nil", got)
	}
	if got := cloneProviderStatusRecords([]ProviderStatusRecord{}); got != nil {
		t.Fatalf("empty clone = %+v, want nil", got)
	}
}

func TestCloneProviderStatusRecordsCopiesSlice(t *testing.T) {
	t.Parallel()
	in := []ProviderStatusRecord{{
		ProviderKind:  capability.ProviderKind("codex"),
		RoutingStatus: hostintegration.ProviderRoutingAvailable,
	}}
	got := cloneProviderStatusRecords(in)
	if len(got) != 1 || got[0] != in[0] {
		t.Fatalf("clone = %+v, want %+v", got, in)
	}
	got[0].RoutingStatus = hostintegration.ProviderRoutingLoginRequired
	if in[0].RoutingStatus != hostintegration.ProviderRoutingAvailable {
		t.Fatalf("clone mutation leaked into source: %+v", in)
	}
}
