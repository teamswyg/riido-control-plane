package riidoaiserver

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/teamswyg/riido-contracts/metadatakeys"
	"github.com/teamswyg/riido-contracts/progressmessage"
)

const (
	progressMessageMetadataCode      = string(metadatakeys.ProgressMessageCode)
	progressMessageMetadataKey       = string(metadatakeys.ProgressMessageKey)
	progressMessageMetadataArgPrefix = string(metadatakeys.ProgressMessageArgPrefix)
)

type progressMessagePayload struct {
	Code    int            `json:"code"`
	Key     string         `json:"key,omitempty"`
	Args    map[string]any `json:"args,omitempty"`
	Message string         `json:"message,omitempty"`
}

func normalizeProgressLine(line AgentThreadProgressLine) (AgentThreadProgressLine, bool) {
	line.Message = strings.TrimSpace(line.Message)
	line.MessageKey = strings.TrimSpace(line.MessageKey)
	line.MessageArgs = copyStringMap(line.MessageArgs)
	if line.MessageCode <= 0 {
		if payload, ok := parseProgressMessagePayload(line.Message); ok {
			line.MessageCode = payload.Code
			line.MessageKey = firstNonEmptyProgressValue(line.MessageKey, payload.Key)
			line.MessageArgs = mergeProgressArgs(line.MessageArgs, payload.Args)
			if message := strings.TrimSpace(payload.Message); message != "" {
				line.Message = message
			}
		}
	}
	if rendered, key, ok := renderProgressMessage(line.MessageCode, line.MessageArgs); ok {
		line.Message = rendered
		if line.MessageKey == "" {
			line.MessageKey = key
		}
		line.MessageArgs = progressmessage.NormalizeArgsForCode(line.MessageCode, line.MessageArgs)
	}
	if line.Message == "" {
		return AgentThreadProgressLine{}, false
	}
	return line, true
}

func parseProgressMessagePayload(message string) (progressMessagePayload, bool) {
	message = strings.TrimSpace(message)
	if !strings.HasPrefix(message, "{") {
		return progressMessagePayload{}, false
	}
	var payload progressMessagePayload
	if err := json.Unmarshal([]byte(message), &payload); err != nil || payload.Code <= 0 {
		return progressMessagePayload{}, false
	}
	payload.Key = strings.TrimSpace(payload.Key)
	return payload, true
}

func mergeProgressArgs(base map[string]string, raw map[string]any) map[string]string {
	if len(raw) == 0 {
		return base
	}
	out := copyStringMap(base)
	if out == nil {
		out = map[string]string{}
	}
	for key, value := range raw {
		key = strings.TrimSpace(key)
		rendered := strings.TrimSpace(progressArgString(value))
		if key == "" || rendered == "" {
			continue
		}
		if _, exists := out[key]; !exists {
			out[key] = rendered
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func firstNonEmptyProgressValue(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func progressArgString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case json.Number:
		return v.String()
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, bool:
		return fmt.Sprint(v)
	default:
		return ""
	}
}

func renderProgressMessage(code int, args map[string]string) (string, string, bool) {
	if code <= 0 {
		return "", "", false
	}
	normalizedArgs := progressmessage.NormalizeArgsForCode(code, args)
	rendered, ok := progressmessage.Render(code, normalizedArgs, progressmessage.DefaultLocale)
	if !ok {
		return "", "", false
	}
	return rendered, progressMessageKey(code), true
}

func progressMessageKey(code int) string {
	catalog, err := progressmessage.Catalog()
	if err != nil {
		return ""
	}
	for _, item := range catalog.Messages {
		if item.Code == code {
			return strings.TrimSpace(item.Key)
		}
	}
	return ""
}

func addProgressLineMetadata(metadata map[string]string, line AgentThreadProgressLine) map[string]string {
	if metadata == nil {
		metadata = map[string]string{}
	}
	if line.MessageCode > 0 {
		metadata[progressMessageMetadataCode] = strconv.Itoa(line.MessageCode)
	}
	if strings.TrimSpace(line.MessageKey) != "" {
		metadata[progressMessageMetadataKey] = strings.TrimSpace(line.MessageKey)
	}
	for key, value := range line.MessageArgs {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		metadata[progressMessageMetadataArgPrefix+key] = value
	}
	return metadata
}

func progressLineMetadata(metadata map[string]string) (int, string, map[string]string) {
	code, _ := strconv.Atoi(strings.TrimSpace(metadata[progressMessageMetadataCode]))
	key := strings.TrimSpace(metadata[progressMessageMetadataKey])
	args := map[string]string{}
	for name, value := range metadata {
		if !strings.HasPrefix(name, progressMessageMetadataArgPrefix) {
			continue
		}
		argName := strings.TrimSpace(strings.TrimPrefix(name, progressMessageMetadataArgPrefix))
		value = strings.TrimSpace(value)
		if argName != "" && value != "" {
			args[argName] = value
		}
	}
	if len(args) == 0 {
		args = nil
	}
	return code, key, args
}
