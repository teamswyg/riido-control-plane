package riidoaiserver

import (
	"context"
	"fmt"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func (s Server) recordAgentThreadProgressEvents(ctx context.Context, agentID string, req AgentThreadProgressBatchRequest) error {
	for _, line := range req.Lines {
		metadata := copyStringMap(req.Metadata)
		metadata[metadatakeys.ThreadProgressSeq.String()] = fmt.Sprint(line.Seq)
		metadata = addProgressLineMetadata(metadata, line)
		if _, err := s.assignment.RecordAgentEvent(ctx, agentID, AgentEventRequest{
			AssignmentID: req.AssignmentID,
			TaskID:       req.TaskID,
			DaemonID:     req.DaemonID,
			DeviceID:     req.DeviceID,
			RuntimeID:    req.RuntimeID,
			State:        AssignmentRunning,
			EventType:    EventRiidoLog,
			Message:      line.Message,
			Metadata:     metadata,
		}); err != nil {
			return err
		}
	}
	return nil
}
