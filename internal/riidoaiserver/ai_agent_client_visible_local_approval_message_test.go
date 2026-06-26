package riidoaiserver

import (
	"strings"
	"testing"
)

func TestClientVisibleTaskThreadTextConvertsLocalApprovalDialog(t *testing.T) {
	t.Parallel()
	for _, input := range []string{
		"승인 방법\n파일 쓰기(Write) — go.mod, main.go 생성\n명령 실행(Bash) — go run 실행\n승인 다이얼로그에서 Allow / 허용을 선택해 주세요.",
		"go 명령 실행에 대한 권한이 필요한 것 같습니다. 파일은 이미 만들어져 있으니, 터미널에서 직접 아래 명령을 실행해 주세요.",
	} {
		got := clientVisibleTaskThreadText(input)
		if got != clientMessageThreadConfirmation {
			t.Fatalf("clientVisibleTaskThreadText approval = %q, want %q", got, clientMessageThreadConfirmation)
		}
		for _, forbidden := range []string{"승인 다이얼로그", "Allow", "Write", "Bash", "터미널에서 직접"} {
			if strings.Contains(got, forbidden) {
				t.Fatalf("approval implementation detail leaked %q in %q", forbidden, got)
			}
		}
	}
}

func TestFailureDiagnosticsConvertsLocalApprovalDialog(t *testing.T) {
	t.Parallel()
	diagnostics := failureDiagnosticsFromAssignmentEvent(nil, "permission dialog requires click Allow")
	if diagnostics == nil {
		t.Fatal("expected diagnostics")
	}
	if diagnostics.Message != clientMessageThreadConfirmation {
		t.Fatalf("diagnostics message = %q, want %q", diagnostics.Message, clientMessageThreadConfirmation)
	}
}
