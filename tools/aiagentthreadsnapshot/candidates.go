package main

import "sort"

func candidateConversations(
	v3 threadCollection,
	v2 threadCollection,
	limit int,
) ([]conversationCandidate, int) {
	source := v3.Threads
	if len(source) == 0 {
		source = v2.Threads
	}
	byID := map[string]*conversationCandidate{}
	for _, thread := range source {
		id := thread.ConversationID
		if id == "" {
			id = thread.ThreadID
		}
		if id == "" {
			continue
		}
		candidate := byID[id]
		if candidate == nil {
			candidate = &conversationCandidate{ConversationID: id}
			byID[id] = candidate
		}
		addThread(candidate, thread)
	}
	candidates := flattenCandidates(byID)
	sort.Slice(candidates, func(i, j int) bool {
		return candidateRank(candidates[i]) > candidateRank(candidates[j])
	})
	if limit > 0 && len(candidates) > limit {
		return candidates[:limit], len(candidates)
	}
	return candidates, len(candidates)
}

func addThread(candidate *conversationCandidate, thread threadRecord) {
	candidate.ThreadCount++
	if isRunning(thread) {
		candidate.RunningCount++
	}
	if thread.WorkStatus == "queued" || thread.AssignmentState == "queued" {
		candidate.QueuedCount++
	}
	if isTerminal(thread.AssignmentState) {
		candidate.TerminalCount++
	}
	if len(thread.ActiveStream) > 0 {
		candidate.ActiveStreams++
	}
	candidate.ThreadID = thread.ThreadID
	candidate.AssignmentID = thread.AssignmentID
	candidate.RunID = thread.RunID
}

func flattenCandidates(byID map[string]*conversationCandidate) []conversationCandidate {
	candidates := make([]conversationCandidate, 0, len(byID))
	for _, candidate := range byID {
		candidates = append(candidates, *candidate)
	}
	return candidates
}

func candidateRank(candidate conversationCandidate) int {
	return candidate.RunningCount*1000 +
		candidate.ActiveStreams*500 +
		candidate.QueuedCount*100 +
		candidate.TerminalCount
}
