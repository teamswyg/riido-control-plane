package riidoaiserver

import (
	"errors"
	"slices"
)

func buildAIAgentRunReads(thread AIAgentTaskThreadRecord, reads []AssignmentContextRead, visible map[runReadSourceKey]bool, page int) (AIAgentRunReadsResponse, error) {
	response := AIAgentRunReadsResponse{SchemaVersion: SchemaVersion, RunID: thread.RunID, TaskID: thread.TaskID, ThreadID: thread.ThreadID, AssignmentID: thread.AssignmentID, Page: page, PageSize: aiAgentRunReadsPageSize, Reads: []AIAgentRunRead{}}
	reads = slices.Clone(reads)
	slices.SortStableFunc(reads, func(a, b AssignmentContextRead) int { return a.ServedAt.Compare(b.ServedAt) })
	seen := 0
	for _, read := range reads {
		if read.AssignmentID != thread.AssignmentID || read.TaskID != thread.TaskID || read.AgentID != thread.AgentID {
			return response, errors.New("run read scope mismatch")
		}
		if !visible[runReadSourceKey{read.SourceKind, read.SourceID, read.SourcePart}] {
			continue
		}
		// Compare before multiplication so even the largest accepted page cannot overflow.
		if seen/aiAgentRunReadsPageSize < page-1 {
			seen++
			continue
		}
		if len(response.Reads) == aiAgentRunReadsPageSize {
			response.HasMore = true
			continue
		}
		response.Reads = append(response.Reads, AIAgentRunRead{
			AgentContextReadSource: AgentContextReadSource{Kind: read.SourceKind, ID: read.SourceID, Part: read.SourcePart, Version: read.SourceVersion, UpdatedAt: read.SourceUpdatedAt, Status: read.Status},
			Via:                    read.Via, ServedAt: read.ServedAt.UTC(),
		})
	}
	return response, nil
}
