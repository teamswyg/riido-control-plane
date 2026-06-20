package riidoaiserver

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func closeInlineCodeAfterLocalFileBeforeKoreanProse(segment string) string {
	const placeholder = "로컬 파일"
	index := strings.LastIndex(segment, placeholder)
	if index < 0 {
		return segment
	}
	afterIndex := index + len(placeholder)
	after := segment[afterIndex:]
	if after == "" || !unicode.IsSpace([]rune(after)[0]) {
		return segment
	}
	trimmed := strings.TrimLeftFunc(after, unicode.IsSpace)
	if trimmed == "" {
		return segment
	}
	first, _ := utf8.DecodeRuneInString(trimmed)
	if !isHangulRune(first) {
		return segment
	}
	return segment[:afterIndex] + "`" + after
}
