package riidoaiserver

import (
	"strings"
	"testing"
)

func TestMeetingPromptContainsTitlesAndExplicitReadCommandOnly(t *testing.T) {
	c := meetingContextFixture(t, "available")
	var out strings.Builder
	writeExtendedTaskContext(&out, c.Component, c.Extended, "")
	if strings.Contains(out.String(), "MEETING_BODY_ONLY_MARKER") || !strings.Contains(out.String(), "Linked meeting") || !strings.Contains(out.String(), "riido meeting get <meeting-id>") {
		t.Fatal("meeting prompt must expose metadata and command without captured body")
	}
}
