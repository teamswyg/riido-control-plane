package riidoaiserver

import (
	"errors"
	"fmt"
	"strings"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-contracts/metadatakeys"
)

// Only finite lifecycle acknowledgements are retained here. Streaming progress
// uses its existing sequence dedupe; the independent recovery journal is intact.
func (s *Store) checkAssignmentEventACK(current Assignment, req AgentEventRequest) (AgentEventResponse, bool, error) {
	key := strings.TrimSpace(req.Metadata[metadatakeys.AssignmentEventKey.String()])
	if key == "" {
		return AgentEventResponse{}, false, nil
	}
	for _, receipt := range s.eventACKFence {
		if receipt.KeyHash != recoveryHash(key) {
			continue
		}
		payloadHash, err := assignmentEventACKPayloadHash(current, req)
		if err != nil {
			return AgentEventResponse{}, false, err
		}
		if receipt.PayloadHash != payloadHash {
			return AgentEventResponse{}, false, errors.New("assignment event key was reused with another payload")
		}
		acknowledged := current
		acknowledged.State, acknowledged.UpdatedAt = receipt.State, receipt.AssignmentUpdatedAt
		acknowledged.ProviderSessionID = receipt.ProviderSessionID
		acknowledged.ConfirmationReason = assignmentcontract.ConfirmationReason(receipt.ConfirmationReason)
		event := TaskEvent{EventID: fmt.Sprintf("%s:%d", current.ID, receipt.Seq), WorkspaceID: current.WorkspaceID, TeamID: current.TeamID, CreatedBy: current.CreatedBy, ConfirmationReason: acknowledged.ConfirmationReason, Seq: receipt.Seq, TaskID: current.TaskID, AssignmentID: current.ID, AgentID: current.AgentID, Type: receipt.Type, State: receipt.State, Message: req.Message, Metadata: cloneMetadata(req.Metadata), At: receipt.At}
		return AgentEventResponse{SchemaVersion: SchemaVersion, Assignment: copyAssignment(acknowledged), Event: event}, true, nil
	}
	if retainedAssignmentEventACK(req) && len(s.eventACKFence) >= maxAssignmentEventACKReceipts {
		return AgentEventResponse{}, false, errors.New("assignment event acknowledgement budget exhausted")
	}
	return AgentEventResponse{}, false, nil
}
