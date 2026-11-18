package riidoaiserver

import "testing"

func TestThreadMessageApprovesToolApprovalExecutionPhrases(t *testing.T) {
	t.Parallel()
	for _, body := range []string{
		"<p>go 명령 실행도 해줘</p>",
		"실행해줘",
		"직접 실행해 주세요",
		"승인할게 진행해줘",
		`<p data-blockid="approval">APPROVE THIS TOOL REQUEST ONLY.</p>`,
	} {
		if !threadMessageApprovesToolApproval(body) {
			t.Fatalf("threadMessageApprovesToolApproval(%q) = false", body)
		}
	}
}

func TestThreadMessageRejectsGenericFollowupPhrases(t *testing.T) {
	t.Parallel()
	for _, body := range []string{
		"다음 작업을 이어서 진행해 주세요.",
		"이어서 진행해줘",
		"계속 진행해줘",
		"검증과 PR 리뷰 후 승인·저장도 확인해줘.",
		"Follow up on the task and verify the approval flow.",
		mentionHTML("agent-public-openclaw") + "<blockquote>" + mentionHTML("agent-public-openclaw") + "</blockquote> 승인",
		mentionHTML("agent-public-openclaw") + "<code>" + mentionHTML("agent-public-openclaw") + "</code> 승인",
	} {
		if threadMessageApprovesToolApproval(body) {
			t.Fatalf("threadMessageApprovesToolApproval(%q) = true", body)
		}
	}
}

func TestThreadMessageRejectsToolApprovalExecutionNegations(t *testing.T) {
	t.Parallel()
	for _, body := range []string{
		"실행하지마",
		"실행하지 말아줘",
		"명령 실행은 안 해도 돼",
	} {
		if threadMessageApprovesToolApproval(body) {
			t.Fatalf("threadMessageApprovesToolApproval(%q) = true", body)
		}
	}
}
