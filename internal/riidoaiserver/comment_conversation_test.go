package riidoaiserver

import (
	"strings"
	"testing"
	"time"
)

func TestCommentFollowupUsesRootAndOnlySameAgentNativeSession(t *testing.T) {
	for _, name := range []string{"same agent", "legacy root", "other agent", "other root", "busy", "waiting for input"} {
		t.Run(name, func(t *testing.T) {
			s, source, thread, principal, req := nativeResumeScopeFixture(t)
			client := s.aiAgent.(*DevelopmentAIAgentClientStore)
			thread.ConversationID = "comment:root"
			thread.AssignmentState = AgentAssignmentStateCompleted
			thread.WorkStatus = AgentWorkStatusCompleted
			thread.StartedAt = time.Now().Add(-time.Hour)
			client.taskThreads[thread.TaskID] = []AIAgentTaskThreadRecord{thread}
			client.taskThreadMessages[thread.ThreadID] = []AIAgentTaskThreadHistoryMessage{{MessageID: "old", Role: AIAgentTaskThreadMessageRoleAgent, Body: "previous finding", ObservedAt: thread.StartedAt}}
			evidence := verifiedCommentDelivery{RootCommentID: "root", CommentCreatedAt: thread.StartedAt.Add(time.Minute), CommentID: "new-reply", TaskID: thread.TaskID, AgentID: thread.AgentID}
			evidence.ConversationMessages = []verifiedCommentContextMessage{{ID: "human-without-mention", Body: "human context without a mention", CreatedAt: thread.StartedAt}}
			if name == "other agent" {
				evidence.AgentID = "agent-public-openclaw"
				req.AgentID = evidence.AgentID
				req.RuntimeProvider = "openclaw"
			}
			if name == "other root" {
				evidence.RootCommentID = "other"
			}
			if name == "busy" || name == "waiting for input" {
				client.taskThreads[thread.TaskID][0].AssignmentState = AgentAssignmentStateRunning
				client.taskThreads[thread.TaskID][0].WorkStatus = AgentWorkStatusRunning
				source.projection.Assignment.State = AssignmentRunning
				if name == "waiting for input" {
					client.taskThreads[thread.TaskID][0].WorkStatus = AgentWorkStatusWaitingForUser
				}
			}
			if name == "legacy root" {
				client.taskThreads[thread.TaskID][0].ConversationID = thread.ThreadID
				client.taskThreads[thread.TaskID][0].SourceCommentID = "older-reply-outside-context-window"
				evidence.ConversationCommentIDs = []string{"older-reply-outside-context-window"}
			}
			var publication AssignAIAgentTaskRequest
			err := s.prepareCommentConversation(t.Context(), principal, evidence, &publication, &req, "user-1")
			if name == "busy" || name == "waiting for input" {
				if err == nil {
					t.Fatal("forked an active session")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !publication.sourceMessageObservedAt.Equal(evidence.CommentCreatedAt) {
				t.Fatal("trigger timestamp lost outside the context window")
			}
			if publication.conversationID != "comment:"+evidence.RootCommentID {
				t.Fatal("reply became another root")
			}
			if (req.ResumeSessionID == "selected-native") != (name == "same agent" || name == "legacy root") {
				t.Fatal("native scope lost")
			}
			if strings.Contains(req.Prompt, "previous finding") != (name != "other root") {
				t.Fatal("conversation context missing or leaked")
			}
			if !strings.Contains(req.Prompt, "human context without a mention") {
				t.Fatal("ordinary reply context was lost")
			}
			if name == "other agent" && !strings.Contains(req.Prompt, "agent (agent-owned-codex)") {
				t.Fatal("handoff context lost original Agent identity")
			}
		})
	}
}
