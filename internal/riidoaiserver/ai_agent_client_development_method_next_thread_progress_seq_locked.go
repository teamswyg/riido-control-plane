package riidoaiserver

import (
	"strconv"
	"strings"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

func (s *DevelopmentAIAgentClientStore) nextThreadProgressSeqLocked(taskID, threadID string, metadata map[string]string) int {
	if value := strings.TrimSpace(metadata[metadatakeys.ThreadProgressSeq.String()]); value != "" {
		if seq, err := strconv.Atoi(value); err == nil && seq > 0 {
			return seq
		}
	}
	threads := s.taskThreads[taskID]
	for i := range threads {
		if threads[i].ThreadID != threadID {
			continue
		}
		return len(threads[i].Lines) + 1
	}
	return 1
}
