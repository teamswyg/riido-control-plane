package riidoaiserver

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestHTTPThreadReplyExecutesExplicitlyMentionedAgent(t *testing.T) {
	f := newNativeResumeHTTPFixture(t)
	first := f.finishFirst(t, AssignmentCompleted)
	f.client.taskThreadMessages[first.ThreadID] = []AIAgentTaskThreadHistoryMessage{{MessageID: "first-result", Role: AIAgentTaskThreadMessageRoleAgent, Body: "Blog template is in workdir/blog-template.md", AssignmentID: first.AssignmentID, ObservedAt: time.Now()}}
	var next AIAgentTaskActionResponse
	f.post(t, f.base+"/tasks/task-new/threads/"+first.ThreadID+"/messages", CreateAIAgentTaskThreadMessageRequest{
		Body: mentionHTML("agent-public-openclaw") + " 이어서 작업해봐",
	}, http.StatusAccepted, &next)
	projection, found, err := f.store.LoadAssignmentProjection(t.Context(), next.AssignmentID)
	if err != nil || !found {
		t.Fatalf("followup assignment: found=%v err=%v", found, err)
	}
	if projection.Assignment.AgentID != "agent-public-openclaw" {
		t.Fatalf("explicit mention executed %q, want agent-public-openclaw", projection.Assignment.AgentID)
	}
	if next.AgentID != "agent-public-openclaw" || next.AgentSnapshot == nil || next.AgentSnapshot.Name == first.AgentSnapshot.Name || next.ThreadID == first.ThreadID || projection.Assignment.ResumeSessionID != "" {
		t.Fatal("handoff kept the original Agent identity or native session")
	}
	if !strings.Contains(projection.Assignment.Prompt, "workdir/blog-template.md") {
		t.Fatal("handoff lost the source Agent's result")
	}
	root, _ := selectAIAgentThreadForFollowup(AIAgentTaskThreadCollectionResponse{Threads: f.client.taskThreads[first.TaskID]}, first.ThreadID)
	child, _ := selectAIAgentThreadForFollowup(AIAgentTaskThreadCollectionResponse{Threads: f.client.taskThreads[first.TaskID]}, next.ThreadID)
	if taskThreadConversationID(root) != child.ConversationID || child.ParentThreadID != first.ThreadID {
		t.Fatal("handoff created another conversation")
	}
}
