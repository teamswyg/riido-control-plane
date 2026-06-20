package riidoaiserver

import (
	"strings"
)

func restoreClientVisibleInlineCode(message string) string {
	if strings.Count(message, "`")%2 == 0 || !strings.Contains(message, "로컬 파일") {
		return message
	}
	segments := strings.Split(message, "`")
	for i := 1; i < len(segments); i += 2 {
		segments[i] = closeInlineCodeAfterLocalFileBeforeKoreanProse(segments[i])
	}
	return strings.Join(segments, "`")
}
