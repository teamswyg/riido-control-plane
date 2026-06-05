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

func TestNormalizeProgressLineStripsStatefulLabelSuffixes(t *testing.T) {
	tests := []struct {
		name      string
		code      int
		args      map[string]string
		wantText  string
		wantLabel string
	}{
		{
			name:      "running",
			code:      1103,
			args:      map[string]string{"label": "테스트 실행", "description": "Rust 프로젝트에서 cargo test를 다시 실행합니다."},
			wantText:  "테스트 실행 중 - Rust 프로젝트에서 cargo test를 다시 실행합니다.",
			wantLabel: "테스트",
		},
		{
			name:      "completed",
			code:      1104,
			args:      map[string]string{"label": "검증 완료", "summary": "README 마지막 줄과 cargo test 통과 결과를 확인했습니다."},
			wantText:  "검증 완료 - README 마지막 줄과 cargo test 통과 결과를 확인했습니다.",
			wantLabel: "검증",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line, ok := normalizeProgressLine(AgentThreadProgressLine{
				Seq:         1,
				MessageCode: tt.code,
				MessageArgs: tt.args,
				ObservedAt:  time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC),
			})
			if !ok {
				t.Fatal("line should be accepted")
			}
			if line.Message != tt.wantText || line.MessageArgs["label"] != tt.wantLabel {
				t.Fatalf("line = %+v", line)
			}
		})
	}
}
