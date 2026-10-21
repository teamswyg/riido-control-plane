package riidoaiserver

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func lifecycleACKFixture(t *testing.T, state AssignmentState) (*conditionalTestOperations, AgentEventRequest) {
	t.Helper()
	now := time.Now().UTC()
	a := Assignment{ID: "ack-assignment", TaskID: "task", AgentID: "agent", State: state, LeaseToken: "lease", CreatedAt: now, UpdatedAt: now}
	return &conditionalTestOperations{projection: AssignmentProjection{Assignment: a, LastEventSeq: 10}}, AgentEventRequest{AssignmentID: a.ID, TaskID: a.TaskID, State: AssignmentFailed, EventType: EventAssignmentFailed, Message: "provider result", Metadata: map[string]string{metadatakeys.AssignmentEventKey.String(): "stable-terminal"}}
}

func ackContainsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

type expiredACKLeaseOperations struct{ *conditionalTestOperations }

func (s *expiredACKLeaseOperations) LoadAgentActiveAssignment(context.Context, string) (AssignmentActiveLease, bool, error) {
	return AssignmentActiveLease{ActiveAssignmentID: s.projection.Assignment.ID, LeaseToken: "lease", LeaseExpiresAt: time.Now().Add(-time.Minute)}, true, nil
}

func (*expiredACKLeaseOperations) RefreshAgentActiveAssignment(context.Context, Assignment, time.Time) error {
	return nil
}
