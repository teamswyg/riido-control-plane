package riidoaiserver

import (
	"testing"
)

func TestHTTPFailedMembershipPublicationDoesNotBlockLaterExplicitRemoval(t *testing.T) {
	for _, state := range []string{"queued", "running", "completed", "failed"} {
		t.Run(state, func(t *testing.T) { checkFailedMembershipPublicationRecovery(t, state) })
	}
}
