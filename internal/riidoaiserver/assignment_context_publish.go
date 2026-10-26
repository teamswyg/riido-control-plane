package riidoaiserver

import (
	"context"
	"log"
)

// publishAssignmentContextReads streams recorded reads to the client. The
// records are already durable, so a stream failure is only logged.
func (s Server) publishAssignmentContextReads(ctx context.Context, reads []AssignmentContextRead) {
	recorder, ok := s.aiAgent.(AIAgentContextReadRecorder)
	if !ok || len(reads) == 0 {
		return
	}
	if err := recorder.RecordAIAgentContextReads(ctx, reads); err != nil {
		log.Printf("event=assignment_context_publish_failed assignment_id=%q error_type=%T", reads[0].AssignmentID, err)
	}
}
