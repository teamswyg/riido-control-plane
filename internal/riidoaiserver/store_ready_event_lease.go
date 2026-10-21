package riidoaiserver

import (
	"errors"
	"strings"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func (s *Store) validateReadyEventLease(state *storeState, assignment Assignment, request *AgentEventRequest) error {
	req := *request
	if req.EventType == EventAssignmentReady {
		if assignment.State != AssignmentLeased && assignment.State != AssignmentReady {
			return errors.New("ready acknowledgement belongs to a stopped execution")
		}
		expired, err := s.assignmentActiveLeaseExpired(state, assignment, s.now())
		if err != nil {
			return err
		}
		if expired {
			return errors.New("ready acknowledgement execution lease expired")
		}
	}
	if req.EventType == EventAssignmentReady && (request.Metadata["lease_token"] != "" || strings.HasPrefix(request.Metadata[metadatakeys.AssignmentEventKey.String()], "ready:")) {
		if assignment.LeaseToken == "" || request.Metadata["lease_token"] != assignment.LeaseToken {
			return errors.New("ready acknowledgement execution lease does not match")
		}
		delete(request.Metadata, "lease_token")
	}
	return nil
}
