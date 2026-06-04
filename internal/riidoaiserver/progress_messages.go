package riidoaiserver

import (
	"regexp"
	"strconv"
	"strings"
)

const (
	progressMessageMetadataCode      = "riido_progress_message_code"
	progressMessageMetadataKey       = "riido_progress_message_key"
	progressMessageMetadataArgPrefix = "riido_progress_message_arg."
)

// Projected from riido-contracts/progressmessage/catalog.ir.riido.json.
// The public client contract still receives rendered message strings.
type progressMessageTemplate struct {
	Code     int
	Key      string
	Template string
}

var progressPlaceholderPattern = regexp.MustCompile(`\{\{([a-zA-Z0-9_]+)\}\}`)

var progressMessageTemplates = []progressMessageTemplate{
	{Code: 1001, Key: "agent.thinking", Template: "생각 중. . ."},
	{Code: 1002, Key: "assignment.queued.agent_busy", Template: "지금은 다른 작업을 처리 중이에요. 현재 작업이 끝나는 대로 바로 시작할게요."},
	{Code: 1003, Key: "assignment.stopped.agent_deleted", Template: "에이전트가 삭제되어 진행 중이던 작업이 중지됐어요."},
	{Code: 1004, Key: "assignment.stopped.by_user", Template: "{{actor_name}}님이 직접 종료하였습니다"},
	{Code: 1101, Key: "tool.collecting", Template: "{{label}} 수집 중 - {{description}}"},
	{Code: 1102, Key: "tool.collection_completed_count", Template: "{{label}} 조회 완료 - {{count}}건({{representative_title}} 외)의 요약을 가져왔습니다. . ."},
	{Code: 1103, Key: "tool.running", Template: "{{label}} 실행 중 - {{description}}"},
	{Code: 1104, Key: "tool.completed", Template: "{{label}} 완료 - {{summary}}"},
	{Code: 1201, Key: "assignment.started", Template: "작업을 시작했어요."},
	{Code: 1202, Key: "assignment.completed", Template: "작업을 완료했어요."},
	{Code: 1203, Key: "assignment.failed", Template: "작업을 계속 진행하지 못했어요."},
}

func normalizeProgressLine(line AgentThreadProgressLine) (AgentThreadProgressLine, bool) {
	line.Message = strings.TrimSpace(line.Message)
	line.MessageKey = strings.TrimSpace(line.MessageKey)
	line.MessageArgs = copyStringMap(line.MessageArgs)
	if rendered, key, ok := renderProgressMessage(line.MessageCode, line.MessageArgs); ok {
		line.Message = rendered
		if line.MessageKey == "" {
			line.MessageKey = key
		}
	}
	if line.Message == "" {
		return AgentThreadProgressLine{}, false
	}
	return line, true
}

func renderProgressMessage(code int, args map[string]string) (string, string, bool) {
	if code <= 0 {
		return "", "", false
	}
	for _, item := range progressMessageTemplates {
		if item.Code != code {
			continue
		}
		return renderProgressTemplate(item.Template, args), item.Key, true
	}
	return "", "", false
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

func renderProgressTemplate(template string, args map[string]string) string {
	return progressPlaceholderPattern.ReplaceAllStringFunc(template, func(match string) string {
		parts := progressPlaceholderPattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		value := strings.TrimSpace(args[parts[1]])
		if value == "" {
			value = "not provided"
		}
		return value
	})
}
