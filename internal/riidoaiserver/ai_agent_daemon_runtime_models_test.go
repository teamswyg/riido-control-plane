package riidoaiserver

import (
	"testing"

	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
)

func TestEqualRuntimeModelRecords(t *testing.T) {
	models := []RuntimeModelRecord{{ModelID: "m1", Label: "M1", IsDefault: true}}
	if !equalRuntimeModelRecords(models, append([]RuntimeModelRecord(nil), models...)) {
		t.Fatalf("matching runtime model records should be equal")
	}
	if equalRuntimeModelRecords(models, nil) {
		t.Fatalf("different runtime model record lengths should not be equal")
	}
	changed := append([]RuntimeModelRecord(nil), models...)
	changed[0].Label = "Changed"
	if equalRuntimeModelRecords(models, changed) {
		t.Fatalf("different runtime model records should not be equal")
	}
	changed[0] = models[0]
	changed[0].ReasoningEfforts = []string{"medium", "high"}
	if equalRuntimeModelRecords(models, changed) {
		t.Fatal("different effort options should invalidate the runtime snapshot")
	}
	models[0].ReasoningEfforts = []string{"medium", "high"}
	changed[0].DefaultReasoningEffort = "medium"
	if equalRuntimeModelRecords(models, changed) {
		t.Fatal("different default effort should invalidate the runtime snapshot")
	}
}

func TestDefaultRuntimeModelByKind(t *testing.T) {
	tests := []struct {
		name    string
		kind    RuntimeKind
		wantID  string
		wantLbl string
	}{
		{"codex", RuntimeKindCodex, providercatalog.DefaultCodexModelID, "Codex 기본 모델"},
		{"claude", RuntimeKindClaudeCode, providercatalog.DefaultClaudeModelID, "Claude Code 기본 모델"},
		{"cursor", RuntimeKindCursor, providercatalog.DefaultCursorModelID, "Cursor Auto"},
		{"openclaw", RuntimeKindOpenClaw, providercatalog.DefaultOpenClawModelID, "OpenClaw 기본 모델"},
		{"unknown", RuntimeKind("other"), providercatalog.DefaultRuntimeModelID, "기본 모델"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := defaultRuntimeModel(tt.kind)
			if got.ModelID != tt.wantID || got.Label != tt.wantLbl || !got.IsDefault {
				t.Fatalf("defaultRuntimeModel(%q) = %#v", tt.kind, got)
			}
		})
	}
}
