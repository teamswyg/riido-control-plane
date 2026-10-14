package riidoaiserver

import "strings"

func intentOrientedTaskMarkers() []string {
	return []string{
		"marketing",
		"copywriting",
		"campaign",
		"analysis",
		"research",
		"planning",
		"strategy",
		"마케팅",
		"카피라이트",
		"카피라이팅",
		"분석",
		"리서치",
		"기획",
		"전략",
		"의도",
		"셀링 포인트",
	}
}

func containsAny(text string, markers []string) bool {
	for _, marker := range markers {
		if strings.Contains(text, strings.ToLower(marker)) {
			return true
		}
	}
	return false
}

func looksKorean(text string) bool {
	for _, r := range text {
		if isHangulRune(r) {
			return true
		}
	}
	return false
}
