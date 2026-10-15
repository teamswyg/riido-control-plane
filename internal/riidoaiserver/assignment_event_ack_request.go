package riidoaiserver

import (
	"encoding/json"
	"strings"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

// Only finite lifecycle acknowledgements are retained here. Streaming progress
// uses its existing sequence dedupe; the independent recovery journal is intact.
func assignmentEventACKRequest(current Assignment, req AgentEventRequest) AgentEventRequest {
	req = normalizeNeedsInputAssignmentEvent(req)
	req.AssignmentID, req.TaskID = strings.TrimSpace(req.AssignmentID), strings.TrimSpace(req.TaskID)
	if req.TaskID == "" {
		req.TaskID = current.TaskID
	}
	req.DaemonID, req.DeviceID, req.RuntimeID = strings.TrimSpace(req.DaemonID), strings.TrimSpace(req.DeviceID), strings.TrimSpace(req.RuntimeID)
	req.ProviderSessionID, req.EventType = strings.TrimSpace(req.ProviderSessionID), strings.TrimSpace(req.EventType)
	if req.EventType == "" {
		if req.State != "" {
			req.EventType = EventAssignmentStateUpdated
		} else {
			req.EventType = EventRiidoLog
		}
	}
	req.Metadata = cloneMetadata(req.Metadata)
	if key := strings.TrimSpace(req.Metadata[metadatakeys.AssignmentEventKey.String()]); key != "" {
		req.Metadata[metadatakeys.AssignmentEventKey.String()] = key
	}
	return req
}

func assignmentEventACKPayloadHash(current Assignment, req AgentEventRequest) (string, error) {
	data, err := json.Marshal(struct {
		AgentID string
		Request AgentEventRequest
	}{current.AgentID, assignmentEventACKRequest(current, req)})
	if err != nil {
		return "", err
	}
	return recoveryHash(string(data)), nil
}
