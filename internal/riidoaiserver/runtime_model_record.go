package riidoaiserver

import "slices"

type RuntimeModelRecord struct {
	ModelID                string   `json:"model_id"`
	Label                  string   `json:"label"`
	IsDefault              bool     `json:"is_default"`
	ReasoningEfforts       []string `json:"reasoning_efforts,omitempty"`
	DefaultReasoningEffort string   `json:"default_reasoning_effort,omitempty"`
}

func equalRuntimeModelRecords(a, b []RuntimeModelRecord) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ModelID != b[i].ModelID || a[i].Label != b[i].Label || a[i].IsDefault != b[i].IsDefault ||
			a[i].DefaultReasoningEffort != b[i].DefaultReasoningEffort || !slices.Equal(a[i].ReasoningEfforts, b[i].ReasoningEfforts) {
			return false
		}
	}
	return true
}
