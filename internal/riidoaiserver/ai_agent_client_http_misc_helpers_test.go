package riidoaiserver

func stringPtr(value string) *string {
	return &value
}

func sameDaemonActions(got, want []DaemonControlAction) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
