package riidoaiserver

import (
	"testing"

	"github.com/teamswyg/riido-contracts/metadatakeys"
	"github.com/teamswyg/riido-contracts/progressmessage"
)

func TestProgressMetadataKeysConsumeContractsVocabulary(t *testing.T) {
	if progressMessageMetadataCode != metadatakeys.ProgressMessageCode.String() {
		t.Fatalf("progress code key = %q", progressMessageMetadataCode)
	}
	if progressMessageMetadataKey != metadatakeys.ProgressMessageKey.String() {
		t.Fatalf("progress message key = %q", progressMessageMetadataKey)
	}
	if progressMessageMetadataArgPrefix != metadatakeys.ProgressMessageArgPrefix.String() {
		t.Fatalf("progress arg prefix = %q", progressMessageMetadataArgPrefix)
	}
	if metadatakeys.ThreadProgressSeq.String() != "thread_progress_seq" {
		t.Fatalf("thread progress seq key = %q", metadatakeys.ThreadProgressSeq)
	}
}

func TestProgressMessageRenderingConsumesContractsCatalog(t *testing.T) {
	args := map[string]string{
		"label":       "GitHub 조회 중",
		"description": "이슈 목록을 조회 중",
	}
	got, key, ok := renderProgressMessage(1101, args)
	if !ok {
		t.Fatal("renderProgressMessage(1101) returned false")
	}
	want, ok := progressmessage.Render(1101, progressmessage.NormalizeArgsForCode(1101, args), progressmessage.DefaultLocale)
	if !ok {
		t.Fatal("contracts progressmessage.Render(1101) returned false")
	}
	if got != want {
		t.Fatalf("rendered progress = %q, want %q", got, want)
	}
	if key != "tool.collecting" {
		t.Fatalf("progress key = %q, want tool.collecting", key)
	}
}
