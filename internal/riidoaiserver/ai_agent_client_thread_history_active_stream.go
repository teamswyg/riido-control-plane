package riidoaiserver

import "slices"

func taskThreadHistoryActiveStream(threads []AIAgentTaskThreadHistoryRecord) *AIAgentTaskThreadStreamLink {
	for _, thread := range slices.Backward(threads) {
		if thread.ActiveStream == nil {
			continue
		}
		stream := *thread.ActiveStream
		return &stream
	}
	return nil
}
