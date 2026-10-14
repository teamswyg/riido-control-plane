package riidoaiserver

import (
	"slices"
	"strings"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func duplicateAssignmentEvent(state *storeState, taskID, assignmentID, agentID, eventType string, metadata map[string]string) (TaskEvent, bool) {
	if existing, ok := duplicateAssignmentEventKey(state, taskID, assignmentID, agentID, eventType, metadata); ok {
		return existing, true
	}
	return duplicateThreadProgressEvent(state, taskID, assignmentID, agentID, eventType, metadata)
}

func duplicateAssignmentEventKey(state *storeState, taskID, assignmentID, agentID, eventType string, metadata map[string]string) (TaskEvent, bool) {
	key := strings.TrimSpace(metadata[metadatakeys.AssignmentEventKey.String()])
	if key == "" {
		return TaskEvent{}, false
	}
	for _, event := range slices.Backward(state.events[taskID]) {
		if event.AssignmentID != assignmentID ||
			event.AgentID != agentID ||
			event.Type != eventType ||
			strings.TrimSpace(event.Metadata[metadatakeys.AssignmentEventKey.String()]) != key {
			continue
		}
		return event, true
	}
	return TaskEvent{}, false
}

func duplicateThreadProgressEvent(state *storeState, taskID, assignmentID, agentID, eventType string, metadata map[string]string) (TaskEvent, bool) {
	if eventType != EventRiidoLog {
		return TaskEvent{}, false
	}
	seq := strings.TrimSpace(metadata[metadatakeys.ThreadProgressSeq.String()])
	if seq == "" {
		return TaskEvent{}, false
	}
	for _, event := range slices.Backward(state.events[taskID]) {
		if event.AssignmentID != assignmentID ||
			event.AgentID != agentID ||
			event.Type != EventRiidoLog ||
			strings.TrimSpace(event.Metadata[metadatakeys.ThreadProgressSeq.String()]) != seq {
			continue
		}
		return event, true
	}
	return TaskEvent{}, false
}
