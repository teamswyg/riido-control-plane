package riidoaiserver

import "time"

func assignmentContextReadRecords(snapshot AssignmentContextSnapshot, sources []AssignmentContextSource, via, status, requestID string, at time.Time) []AssignmentContextRead {
	reads := make([]AssignmentContextRead, 0, len(sources))
	for _, source := range sources {
		reads = append(reads, AssignmentContextRead{
			AssignmentID:         snapshot.AssignmentID,
			TaskID:               snapshot.TaskID,
			AgentID:              snapshot.AgentID,
			RequesterPrincipalID: snapshot.RequesterPrincipalID,
			TriggerKind:          snapshot.TriggerKind,
			TriggerID:            snapshot.TriggerID,
			SourceKind:           source.Kind,
			SourceID:             source.ID,
			SourcePart:           source.Part,
			SourceVersion:        source.Version,
			SourceUpdatedAt:      source.UpdatedAt,
			Via:                  via,
			Status:               status,
			ServedAt:             at,
			RequestID:            requestID,
		})
	}
	return reads
}
