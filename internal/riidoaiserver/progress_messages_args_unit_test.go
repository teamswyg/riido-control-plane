package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func TestProgressArgStringScalarTypes(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{name: "nil", in: nil, want: ""},
		{name: "string", in: "  hello  ", want: "  hello  "},
		{name: "json-number", in: json.Number("42.5"), want: "42.5"},
		{name: "float64", in: float64(3.25), want: "3.25"},
		{name: "float32", in: float32(1.5), want: "1.5"},
		{name: "int", in: int16(-7), want: "-7"},
		{name: "uint", in: uint8(9), want: "9"},
		{name: "bool", in: true, want: "true"},
		{name: "unsupported", in: []string{"x"}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := progressArgString(tt.in); got != tt.want {
				t.Fatalf("progressArgString(%T) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestMergeProgressArgsKeepsExistingAndDropsBlankValues(t *testing.T) {
	spacedLabel := " label "
	got := mergeProgressArgs(
		map[string]string{"label": "existing"},
		map[string]any{
			spacedLabel: "new",
			"count":     json.Number("3"),
			"blank":     "   ",
			"":          "ignored",
			"list":      []string{"ignored"},
		},
	)
	if got["label"] != "existing" || got["count"] != "3" {
		t.Fatalf("merged args = %+v", got)
	}
	if _, ok := got["blank"]; ok {
		t.Fatalf("blank arg was retained: %+v", got)
	}
}
