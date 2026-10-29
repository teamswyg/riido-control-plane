package riidoaiserver

import (
	"errors"
	"strings"
	"testing"
)

func TestContextReadStreamFailsClosed(t *testing.T) {
	for _, mode := range []string{"replay", "live"} {
		for _, failure := range []string{"task", "upstream", "empty", "wrong-task", "all-hidden", "agent", "malformed"} {
			t.Run(mode+"/"+failure, func(t *testing.T) {
				s, reader, auth, event := contextReadStreamServer(t)
				read := event.Payload.(AgentContextReadEvent)
				switch failure {
				case "task":
					auth.fail = true
				case "upstream":
					reader.err = errors.New("private upstream details")
				case "empty":
					reader.value = AIAgentTaskContext{}
				case "wrong-task":
					reader.value.Component.ID = "other-task"
				case "all-hidden":
					read.Sources = read.Sources[1:]
				case "agent":
					store := s.aiAgent.(*DevelopmentAIAgentClientStore)
					agent := store.agents[read.AgentID]
					agent.Visibility = AgentVisibilityPrivate
					store.agents[read.AgentID] = agent
				case "malformed":
					read.TaskID = ""
				}
				event.Payload = read
				body := contextReadStreamRequest(s, mode, event).Body.String()
				if body != "" {
					t.Fatalf("denied source existence/error leaked: %q", body)
				}
			})
		}
	}
}

func TestContextReadStreamKeepsCurrentDirectoryVisibility(t *testing.T) {
	s, _, _, event := contextReadStreamServer(t)
	read := event.Payload.(AgentContextReadEvent)
	read.Sources = append(read.Sources,
		AgentContextReadSource{Kind: "agent", ID: "agent-public-openclaw", Part: "profile"},
		AgentContextReadSource{Kind: "agent", ID: "agent-private-cursor", Part: "profile", Version: "private-hash"})
	event.Payload = read
	body := contextReadStreamRequest(s, "replay", event).Body.String()
	if !strings.Contains(body, "agent-public-openclaw") || strings.Contains(body, "private-hash") || strings.Contains(body, "agent-private-cursor") {
		t.Fatalf("directory visibility changed: %s", body)
	}
	if len(event.Payload.(AgentContextReadEvent).Sources) != 4 {
		t.Fatal("delivery mutated the shared source list")
	}
}
