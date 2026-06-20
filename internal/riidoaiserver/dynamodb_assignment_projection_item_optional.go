package riidoaiserver

func assignmentProjectionOptionalFields(item map[string]map[string]string, assignment Assignment) {
	if assignment.LeaseToken != "" {
		item["lease_token"] = map[string]string{"S": assignment.LeaseToken}
	}
	if assignment.AgentInstruction != "" {
		item["agent_instruction"] = map[string]string{"S": assignment.AgentInstruction}
	}
	if assignment.ReplacesAssignmentID != "" {
		item["replaces_assignment_id"] = map[string]string{"S": assignment.ReplacesAssignmentID}
	}
	if assignment.BlockedByAssignmentID != "" {
		item["blocked_by_assignment_id"] = map[string]string{"S": assignment.BlockedByAssignmentID}
	}
	if assignment.State.Code() == AssignmentStateCodeQueued {
		item["agent_id"] = map[string]string{"S": assignment.AgentID}
		item["assignment_sort"] = map[string]string{"S": assignmentQueueSort(assignment)}
	}
}
