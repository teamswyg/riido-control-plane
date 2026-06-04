package riidoaiserver

import (
	"testing"
	"time"
)

func TestNormalizeProgressLineRendersBareJSONPayload(t *testing.T) {
	line, ok := normalizeProgressLine(AgentThreadProgressLine{
		Seq:        1,
		Message:    `{"code":1102,"args":{"label":"팀 프로젝트","count":3,"representative_title":"프로젝트 Alpha"}}`,
		ObservedAt: time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC),
	})
	if !ok {
		t.Fatal("line should be accepted")
	}

	want := "팀 프로젝트 조회 완료 - 3건(프로젝트 Alpha 외)의 요약을 가져왔습니다. . ."
	if line.Message != want ||
		line.MessageCode != 1102 ||
		line.MessageKey != "tool.collection_completed_count" ||
		line.MessageArgs["count"] != "3" {
		t.Fatalf("line = %+v", line)
	}
}
