package riidoaiserver

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestLatestProviderDocumentPreservesRealRiidoAtomicNodes(t *testing.T) {
	for _, fixture := range []struct{ node, want string }{
		{`{"type":"riidoMention","attrs":{"mentionId":"member","kind":"member","label":"김유래"}}`, "@김유래"},
		{`{"type":"mentionedComponent","attrs":{"componentId":"task","type":"task","label":"본문 참조"}}`, "@본문 참조"},
		{`{"type":"mentionedComponent","attrs":{"componentId":"task-no-label"}}`, "@task-no-label"},
		{`{"type":"mention","attrs":{"id":"member-no-label"}}`, "[member-no-label]"},
		{`{"type":"imageBlock","attrs":{"src":"https://cdn.example.test/image.png","alt":"설명 이미지"}}`, "[Media: 설명 이미지] (https://cdn.example.test/image.png)"},
		{`{"type":"customVideo","attrs":{"src":"https://cdn.example.test/demo.mp4"}}`, "[Media: ] (https://cdn.example.test/demo.mp4)"},
		{`{"type":"imageBlock","attrs":{"src":"https://cdn.example.test/private?token=secret"}}`, "[Media: ]"},
	} {
		raw := `{"default":{"type":"doc","content":[` + fixture.node + `]}}`
		stored, err := json.Marshal(raw)
		if err != nil {
			t.Fatal(err)
		}
		got, err := readableTaskContextTiptapJSON(stored)
		if err != nil || got != fixture.want || strings.Contains(got, "secret") {
			t.Fatalf("atom=%s got=%q want=%q err=%v", fixture.node, got, fixture.want, err)
		}
	}
}
