package riidoaiserver

import "strings"

// recordTaskThreadRequesterLocked remembers who started a run (assignment,
// comment mention, or thread reply). It only fills an empty value so later
// actions on the same run (resume, stop, unassign) never overwrite the
// original requester. Runs created from runtime events alone keep it empty.
func (s *DevelopmentAIAgentClientStore) recordTaskThreadRequesterLocked(taskID, threadID, principalID string) {
	principalID = strings.TrimSpace(principalID)
	threadID = strings.TrimSpace(threadID)
	if principalID == "" || threadID == "" {
		return
	}
	threads := s.taskThreads[taskID]
	for i := range threads {
		if threads[i].ThreadID != threadID {
			continue
		}
		if threads[i].RequestedBy == "" {
			threads[i].RequestedBy = principalID
			s.taskThreads[taskID] = threads
		}
		return
	}
}
