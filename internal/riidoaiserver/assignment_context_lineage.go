package riidoaiserver

// Only authenticated selection and a successful publication response may seed
// conversation lineage. Neither a CLI reference nor a resume ID is evidence.
func (trigger assignmentContextTrigger) withThread(assignment Assignment, response AIAgentTaskActionResponse, selected *AIAgentTaskThreadRecord) assignmentContextTrigger {
	if response.AssignmentID != assignment.ID || response.TaskID != assignment.TaskID ||
		response.AgentID != assignment.AgentID || response.ThreadID == "" {
		return trigger
	}
	trigger.ThreadID, trigger.ConversationID = response.ThreadID, response.ThreadID
	if selected != nil {
		if selected.TaskID != assignment.TaskID || selected.AgentID != assignment.AgentID || selected.ThreadID == "" {
			return assignmentContextTrigger{Kind: trigger.Kind, ID: trigger.ID, RequesterAdmin: trigger.RequesterAdmin}
		}
		trigger.ConversationID = taskThreadConversationID(*selected)
		trigger.PreviousAssignmentID, trigger.PreviousThreadID = selected.AssignmentID, selected.ThreadID
	}
	return trigger
}

func clearAssignmentContextLineage(snapshot *AssignmentContextSnapshot) {
	snapshot.ConversationID, snapshot.ThreadID = "", ""
	snapshot.PreviousAssignmentID, snapshot.PreviousThreadID = "", ""
	snapshot.SourceStates = nil
}
