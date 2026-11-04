package riidoaiserver

import (
	"crypto/sha256"
	"encoding/hex"
	"html"
	"strings"
	"time"
)

func (s *DevelopmentAIAgentClientStore) recordNativeHistoryLocked(input assignmentEventInput, message nativeHistoryMessage, req AgentEventRequest) error {
	thread, found := s.taskThreadForAssignmentLocked(input.TaskID, input.AgentID, input.AssignmentID)
	if !found {
		return ErrAIAgentTaskThreadConflict
	}
	hash := sha256.Sum256([]byte(message.SessionID + "\x00" + message.MessageID))
	id := "native:" + hex.EncodeToString(hash[:])
	// Replay may arrive on a later execution after a daemon restart. The source
	// identity is scoped to the conversation and Agent, not the importing run.
	for _, row := range s.taskThreads[input.TaskID] {
		if row.AgentID != input.AgentID || taskThreadConversationID(row) != taskThreadConversationID(thread) {
			continue
		}
		for _, existing := range s.taskThreadMessages[row.ThreadID] {
			if existing.MessageID == id {
				return nil
			}
		}
	}
	// Imported turns precede the request that triggered this execution. Some
	// providers omit timestamps; anchor those turns immediately before that
	// request, with their source ordinal as the stable ordering tie-breaker.
	boundary := thread.StartedAt
	for _, existing := range s.taskThreadMessages[thread.ThreadID] {
		if existing.Role == AIAgentTaskThreadMessageRoleUser && !strings.HasPrefix(existing.MessageID, "native:") && (boundary.IsZero() || existing.ObservedAt.Before(boundary)) {
			boundary = existing.ObservedAt
		}
	}
	observed := message.ObservedAt
	if observed.IsZero() || observed.UnixMilli() >= boundary.UnixMilli() {
		observed = boundary.Add(-time.Millisecond)
	}
	var authorID string
	body := message.Body
	if message.Role == "user" {
		authorID = s.nativeCodexDeviceOwnerLocked(input.AgentID, req)
		body = "<p>" + strings.ReplaceAll(html.EscapeString(body), "\n", "<br>") + "</p>"
	}
	s.appendThreadHistoryMessageLocked(thread.ThreadID, AIAgentTaskThreadHistoryMessage{
		MessageID: id, SourceMessageID: id, Role: AIAgentTaskThreadMessageRole(message.Role),
		AssignmentID: thread.AssignmentID, RunID: thread.RunID, Body: body,
		AuthorPrincipalID: authorID, ObservedAt: observed, Seq: message.Ordinal - 1000001,
	})
	return nil
}

// The event has passed assignment and daemon binding validation. Resolve the
// owner from the registered device, independently of who requested the run.
func (s *DevelopmentAIAgentClientStore) nativeCodexDeviceOwnerLocked(agentID string, req AgentEventRequest) string {
	agent, ok := s.agents[agentID]
	if !ok || req.RuntimeID == "" || agent.RuntimeID != req.RuntimeID || req.DeviceID == "" {
		return ""
	}
	device, runtime, ok := s.deviceRuntimeByRuntimeIDLocked(req.RuntimeID)
	if !ok || runtime.Kind != RuntimeKindCodex || device.DeviceID != req.DeviceID {
		return ""
	}
	return device.OwnerPrincipalID
}
