package riidoaiserver

import (
	"encoding/json"
	"time"
)

type AgentThreadProgressLine struct {
	displayMessage string
	Seq            int               `json:"seq"`
	Message        string            `json:"message"`
	MessageCode    int               `json:"message_code,omitempty"`
	MessageKey     string            `json:"message_key,omitempty"`
	MessageArgs    map[string]string `json:"message_args,omitempty"`
	ObservedAt     time.Time         `json:"observed_at,omitempty"`
}

func (line *AgentThreadProgressLine) UnmarshalJSON(data []byte) error {
	line.displayMessage = ""
	var wire struct {
		Seq         int               `json:"seq"`
		Message     string            `json:"message"`
		MessageCode int               `json:"message_code,omitempty"`
		MessageKey  string            `json:"message_key,omitempty"`
		MessageArgs map[string]string `json:"message_args,omitempty"`
		ObservedAt  time.Time         `json:"observed_at,omitempty"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	line.Seq = wire.Seq
	line.Message = wire.Message
	line.MessageCode = wire.MessageCode
	line.MessageKey = wire.MessageKey
	line.MessageArgs = wire.MessageArgs
	line.ObservedAt = wire.ObservedAt
	return nil
}

func (line AgentThreadProgressLine) MarshalJSON() ([]byte, error) {
	wire := struct {
		DisplayMessage string    `json:"display_message,omitempty"`
		MessageKey     string    `json:"message_key,omitempty"`
		Seq            int       `json:"seq"`
		Message        string    `json:"message"`
		ObservedAt     time.Time `json:"observed_at,omitempty"`
	}{
		DisplayMessage: taskProgressDisplayMessage(line),
		Seq:            line.Seq,
		Message:        line.Message,
		ObservedAt:     line.ObservedAt,
	}
	if line.MessageKey == "assistant.partial" || line.MessageKey == "runtime.activity" {
		wire.MessageKey = line.MessageKey
	}
	return json.Marshal(wire)
}
