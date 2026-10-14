package main

func refreshWorkflows(loops []loopRecord) map[string]string {
	out := map[string]string{}
	for _, loop := range loops {
		out[loop.ID] = loop.RefreshWorkflow
	}
	return out
}
