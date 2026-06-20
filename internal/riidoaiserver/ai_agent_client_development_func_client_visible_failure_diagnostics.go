package riidoaiserver

func clientVisibleFailureDiagnostics(in *AIAgentTaskThreadFailureDiagnostics) *AIAgentTaskThreadFailureDiagnostics {
	out := copyFailureDiagnostics(in)
	if out == nil {
		return nil
	}
	out.Message = clientVisibleTaskThreadText(out.Message)
	return out
}
