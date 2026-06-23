package riidoaiserver

import (
	"strconv"
	"strings"

	"github.com/teamswyg/riido-contracts/metadatakeys"
)

const (
	progressMessageMetadataCode      = string(metadatakeys.ProgressMessageCode)
	progressMessageMetadataKey       = string(metadatakeys.ProgressMessageKey)
	progressMessageMetadataArgPrefix = string(metadatakeys.ProgressMessageArgPrefix)
)

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
