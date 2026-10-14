package riidoaiserver

func clientVisibleTaskThreadResultMessage(thread AIAgentTaskThreadRecord) string {
	if result := clientVisibleTaskThreadText(thread.ResultMessage); result != "" {
		return result
	}
	if taskThreadCarriesResultMessage(thread) {
		return clientVisibleTaskThreadMessage(thread)
	}
	return ""
}
