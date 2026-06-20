package riidoaiserver

func (s *DevelopmentAIAgentClientStore) nextClientEventSeqLocked() int64 {
	var maxSeq int64
	for _, event := range s.events {
		if event.Seq > maxSeq {
			maxSeq = event.Seq
		}
	}
	return maxSeq + 1
}
