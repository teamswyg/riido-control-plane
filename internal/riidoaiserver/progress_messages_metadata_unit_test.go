package riidoaiserver

import "testing"

func TestProgressLineMetadataRoundTripTrimsAndDropsBlankArgs(t *testing.T) {
	spacedLabel := " label "
	metadata := addProgressLineMetadata(map[string]string{"keep": "me"}, AgentThreadProgressLine{
		MessageCode: 1201,
		MessageKey:  " tool.running ",
		MessageArgs: map[string]string{
			spacedLabel: " Compile ",
			"empty":     "   ",
			"":          "ignored",
		},
	})
	code, key, args := progressLineMetadata(metadata)
	if metadata["keep"] != "me" {
		t.Fatalf("base metadata was not preserved: %+v", metadata)
	}
	if code != 1201 || key != "tool.running" || args["label"] != "Compile" {
		t.Fatalf("roundtrip = code %d key %q args %+v", code, key, args)
	}
	if _, ok := args["empty"]; ok {
		t.Fatalf("blank arg was retained: %+v", args)
	}
}

func TestProgressLineMetadataReturnsNilArgsWhenEmpty(t *testing.T) {
	code, key, args := progressLineMetadata(map[string]string{
		progressMessageMetadataCode:            "not-a-number",
		progressMessageMetadataKey:             "  ",
		progressMessageMetadataArgPrefix + " ": "ignored",
	})
	if code != 0 || key != "" || args != nil {
		t.Fatalf("metadata = code %d key %q args %+v", code, key, args)
	}
}

func TestAddProgressLineMetadataCreatesMapWhenNil(t *testing.T) {
	metadata := addProgressLineMetadata(nil, AgentThreadProgressLine{
		MessageCode: 7,
	})
	if metadata[progressMessageMetadataCode] != "7" {
		t.Fatalf("metadata = %+v", metadata)
	}
}
