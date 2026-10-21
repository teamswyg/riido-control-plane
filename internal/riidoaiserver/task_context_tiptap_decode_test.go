package riidoaiserver

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestProviderTiptapDocumentAcceptsStoredSerializedFragmentMap(t *testing.T) {
	for _, body := range []string{`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"latest paragraph"}]}]}`, `{"type":"doc","content":[]}`} {
		for _, raw := range []string{body, `{"default":` + body + `,"other":{"type":"doc","content":[{"type":"text","text":"other fragment must not leak"}]}}`} {
			serialized, err := json.Marshal(raw)
			if err != nil {
				t.Fatal(err)
			}
			for _, input := range [][]byte{[]byte(raw), serialized} {
				got, err := readableTaskContextTiptapJSON(input)
				if err != nil {
					t.Fatal(err)
				}
				want := ""
				if strings.Contains(body, "latest paragraph") {
					want = "latest paragraph"
				}
				if got != want {
					t.Fatalf("stored document=%s converted=%q want=%q", input, got, want)
				}
			}
		}
	}
}
