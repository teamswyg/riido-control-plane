package riidoaiserver

func copyFailureDiagnostics(in *AIAgentTaskThreadFailureDiagnostics) *AIAgentTaskThreadFailureDiagnostics {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}
