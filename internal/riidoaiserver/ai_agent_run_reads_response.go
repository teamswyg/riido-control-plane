package riidoaiserver

import (
	"errors"
	"slices"
)

func buildAIAgentRunReads(thread AIAgentTaskThreadRecord, reads []AssignmentContextRead, visible map[runReadSourceKey]runReadMetadata, page int) (AIAgentRunReadsResponse, error) {
	response := AIAgentRunReadsResponse{SchemaVersion: SchemaVersion, RunID: thread.RunID, TaskID: thread.TaskID, ThreadID: thread.ThreadID, AssignmentID: thread.AssignmentID, Page: page, PageSize: aiAgentRunReadsPageSize, Reads: []AIAgentRunRead{}}
	reads = slices.Clone(reads)
	slices.SortStableFunc(reads, func(a, b AssignmentContextRead) int { return a.ServedAt.Compare(b.ServedAt) })
	seen := 0
	listedGroups := map[runReadOwner]bool{}
	for _, read := range reads {
		if read.AssignmentID != thread.AssignmentID || read.TaskID != thread.TaskID || read.AgentID != thread.AgentID {
			return response, errors.New("run read scope mismatch")
		}
		metadata, ok := visible[runReadSourceKey{read.SourceKind, read.SourceID, read.SourcePart}]
		if !ok {
			continue
		}
		listed := runReadListed(read)
		if listed {
			listedGroups[runReadGroup(read, metadata.owner)] = true
		}
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
			Title: metadata.title, ModifiedSinceRead: runReadModifiedSinceRead(read, metadata),
			OwnerKind: metadata.owner.kind, OwnerID: metadata.owner.id, Listed: listed,
		})
	}
	response.SourceTotal = len(listedGroups)
	return response, nil
}

// runReadListed hides prompt evidence that only carried a title or a short
// summary; the read UI lists sources whose content the agent actually received.
func runReadListed(read AssignmentContextRead) bool {
	return read.Via != AssignmentContextViaPrompt || (read.SourcePart != "title" && read.SourcePart != "summary")
}

// runReadGroup is the unit the read UI counts: the owning work, meeting or the
// Agent list for child sources, and the record itself otherwise.
func runReadGroup(read AssignmentContextRead, owner runReadOwner) runReadOwner {
	if owner.kind != "" {
		return owner
	}
	return runReadOwner{read.SourceKind, read.SourceID}
}
