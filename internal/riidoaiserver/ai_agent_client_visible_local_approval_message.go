package riidoaiserver

import "strings"

func clientVisibleLocalApprovalMessage(message string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(message))
	if normalized == "" {
		return "", false
	}
	for _, marker := range localApprovalMessageMarkers() {
		if strings.Contains(normalized, marker) {
			return clientMessageThreadConfirmation, true
		}
	}
	return "", false
}

func localApprovalMessageMarkers() []string {
	return []string{
		"approval dialog",
		"allow / 허용",
		"bash) —",
		"write) —",
		"click allow",
		"local tool approval",
		"permission dialog",
		"승인 다이얼로그",
		"승인 방법",
		"권한 승인",
		"권한이 필요",
		"권한을 허용",
		"실행 권한",
		"명령 실행 권한",
		"go 명령 실행",
		"터미널에서 직접",
		"파일 쓰기(write)",
		"명령 실행(bash)",
	}
}
