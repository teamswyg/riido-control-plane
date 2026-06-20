package riidoaiserver

import (
	"strings"
)

func clientVisibleTaskThreadText(message string) string {
	message = stripRiidoLogBlocks(message)
	message = clientVisibleMarkdownLocalLinkPattern.ReplaceAllString(message, "$1")
	message = clientVisibleAngleLocalPathPattern.ReplaceAllString(message, "로컬 파일")
	message = clientVisibleApplicationSupportLocalPathPattern.ReplaceAllString(message, "로컬 파일")
	message = clientVisibleLocalPathPattern.ReplaceAllString(message, "로컬 파일")
	message = restoreClientVisibleInlineCode(message)
	return strings.TrimSpace(message)
}
