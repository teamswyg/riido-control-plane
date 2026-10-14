package riidoaiserver

func failureDiagnosticsEqual(a, b *AIAgentTaskThreadFailureDiagnostics) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.ResultStatus == b.ResultStatus &&
		a.FailureCategory == b.FailureCategory &&
		a.Message == b.Message
}
