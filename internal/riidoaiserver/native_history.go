package riidoaiserver

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
)

const nativeHistoryMetadataKey = "native_history_message"

// This wire payload is consumed from authenticated daemon events only, never
// from comment bodies. See thread_history_v3_frontend in docs/20-domain/ai-agent-client-api.riido.json.
type nativeHistoryMessage struct {
	SessionID  string    `json:"session_id"`
	MessageID  string    `json:"message_id"`
	Role       string    `json:"role"`
	Body       string    `json:"body"`
	ObservedAt time.Time `json:"observed_at,omitempty"`
	Ordinal    int       `json:"ordinal"`
}

func decodeNativeHistory(metadata map[string]string) (nativeHistoryMessage, bool, error) {
	raw, found := metadata[nativeHistoryMetadataKey]
	var message nativeHistoryMessage
	if !found {
		return message, false, nil
	}
	if len(raw) > 512<<10 || json.Unmarshal([]byte(raw), &message) != nil || message.SessionID == "" || message.MessageID == "" || len(message.MessageID) > 1024 || len(message.Body) > 65536 || strings.TrimSpace(message.Body) == "" || (message.Role != "user" && message.Role != "agent") || message.Ordinal < 1 || message.Ordinal > 1000000 {
		return message, true, errors.New("invalid native history message")
	}
	return message, true, nil
}

func validateNativeHistoryEvent(assignment Assignment, req AgentEventRequest) error {
	message, found, err := decodeNativeHistory(req.Metadata)
	if err != nil || !found {
		return err
	}
	if req.EventType != EventRiidoLog || isTerminal(assignment.State) || assignment.ResumeSessionID == "" || message.SessionID != assignment.ResumeSessionID || message.SessionID != assignmentcontract.ResumeSessionIDForAssignment(assignment) || (assignment.LeaseToken != "" && req.Metadata["lease_token"] != assignment.LeaseToken) {
		return errors.New("native history does not belong to the active resumed assignment")
	}
	return nil
}
