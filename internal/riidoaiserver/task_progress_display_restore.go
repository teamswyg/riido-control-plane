package riidoaiserver

import (
	"encoding/json"
	"strings"
	"unicode/utf8"
)

// Records and outbound events are restored from server-owned snapshots. The
// ingress AgentThreadProgressLine decoder deliberately ignores display_message.
func (thread *AIAgentTaskThreadRecord) UnmarshalJSON(raw []byte) error {
	type record AIAgentTaskThreadRecord
	if err := json.Unmarshal(raw, (*record)(thread)); err != nil {
		return err
	}
	return restoreTaskProgressDisplayMessages(raw, thread.Lines)
}

func (event *AgentThreadProgressEvent) UnmarshalJSON(raw []byte) error {
	type progressEvent AgentThreadProgressEvent
	if err := json.Unmarshal(raw, (*progressEvent)(event)); err != nil {
		return err
	}
	return restoreTaskProgressDisplayMessages(raw, event.Lines)
}

func restoreTaskProgressDisplayMessages(raw []byte, lines []AgentThreadProgressLine) error {
	var saved struct {
		Lines []struct {
			DisplayMessage string `json:"display_message"`
		} `json:"lines"`
	}
	if err := json.Unmarshal(raw, &saved); err != nil {
		return err
	}
	for i, line := range saved.Lines {
		text := strings.TrimSpace(line.DisplayMessage)
		if i < len(lines) && !strings.ContainsAny(text, "\r\n") && utf8.RuneCountInString(text) <= 180 {
			lines[i].displayMessage = text
		}
	}
	return nil
}
