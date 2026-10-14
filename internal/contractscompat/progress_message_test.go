package contractscompat

import (
	"testing"

	"github.com/teamswyg/riido-contracts/progressmessage"
)

func TestProgressMessageBaseline(t *testing.T) {
	rendered, ok := progressmessage.Render(1101, progressmessage.NormalizeArgsForCode(1101, map[string]string{
		"label":       "GitHub 조회 중",
		"description": "이슈 목록",
	}), progressmessage.DefaultLocale)
	if !ok || rendered != "GitHub 수집 중 - 이슈 목록" {
		t.Fatalf("progressmessage.Render = %q, %v", rendered, ok)
	}
}
