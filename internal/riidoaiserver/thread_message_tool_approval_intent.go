package riidoaiserver

import "strings"

func threadMessageApprovesToolApproval(body string) bool {
	text := strings.ToLower(strings.TrimSpace(body))
	if text == "" || threadMessageRejectsToolApproval(text) {
		return false
	}
	for _, phrase := range []string{
		"승인", "허용", "직접 실행", "실행해줘", "실행해 주세요",
		"실행해", "실행도 해줘", "명령 실행", "go 명령",
		"go run", "approve", "approved", "allow", "allowed",
		"go ahead",
	} {
		if strings.Contains(text, phrase) {
			return true
		}
	}
	return false
}

func threadMessageRejectsToolApproval(text string) bool {
	for _, phrase := range []string{
		"승인하지", "승인 안", "허용하지", "허용 안", "거절", "하지마",
		"하지 마", "하지 말", "안 해", "멈춰", "중단", "deny", "denied",
		"reject", "rejected", "do not",
	} {
		if strings.Contains(text, phrase) {
			return true
		}
	}
	return false
}
