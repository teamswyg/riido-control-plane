package riidoaiserver

func copyQueueDiagnostics(in *AIAgentTaskThreadQueueDiagnostics) *AIAgentTaskThreadQueueDiagnostics {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}
