package main

type taskContextRawEnv struct {
	baseURL     string
	workspaceID string
	teamID      string
	apiKey      string
	timeoutRaw  string
}

func (raw taskContextRawEnv) empty() bool {
	return raw.baseURL == "" &&
		raw.workspaceID == "" &&
		raw.teamID == "" &&
		raw.apiKey == "" &&
		raw.timeoutRaw == ""
}

func (raw taskContextRawEnv) openAPIUnset() bool {
	return raw.workspaceID == "" && raw.teamID == "" && raw.apiKey == ""
}

func (raw taskContextRawEnv) openAPIComplete() bool {
	return raw.workspaceID != "" && raw.teamID != "" && raw.apiKey != ""
}
