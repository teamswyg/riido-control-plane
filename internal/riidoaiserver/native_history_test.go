package riidoaiserver

import (
	"encoding/json"
	"strings"
	"testing"
)

func nativeHistoryTestMetadata(t *testing.T, message nativeHistoryMessage) map[string]string {
	t.Helper()
	data, err := json.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}
	return map[string]string{nativeHistoryMetadataKey: string(data), "lease_token": "lease"}
}

func TestNativeHistoryValidatesResumedAssignmentScope(t *testing.T) {
	assignment := Assignment{State: AssignmentRunning, ResumeSessionID: "session", ProviderSessionID: "session", LeaseToken: "lease"}
	message := nativeHistoryMessage{SessionID: "session", MessageID: "turn:item", Role: "agent", Body: "external answer", Ordinal: 1}
	for _, name := range []string{"valid", "escaped body", "foreign session", "fresh run", "terminal", "lease", "wrong kind", "oversize", "unsupported role"} {
		t.Run(name, func(t *testing.T) {
			a, m := assignment, message
			req := AgentEventRequest{EventType: EventRiidoLog}
			switch name {
			case "escaped body":
				m.Body = strings.Repeat("<", 65536)
			case "foreign session":
				m.SessionID = "foreign"
			case "fresh run":
				a.ResumeSessionID = ""
			case "terminal":
				a.State = AssignmentCompleted
			case "lease":
				a.LeaseToken = "other"
			case "wrong kind":
				req.EventType = EventProviderLog
			case "oversize":
				m.Body = strings.Repeat("x", 65537)
			case "unsupported role":
				m.Role = "tool"
			}
			req.Metadata = nativeHistoryTestMetadata(t, m)
			if err := validateNativeHistoryEvent(a, req); (err == nil) != (name == "valid" || name == "escaped body") {
				t.Fatalf("validation=%v", err)
			}
		})
	}
}
