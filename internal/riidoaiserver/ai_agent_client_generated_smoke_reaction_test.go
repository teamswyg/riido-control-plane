package riidoaiserver

import (
	"net/http"
	"strings"
	"testing"
)

// assertAIAgentV2SmokeThreadReactions adds and deletes a member reaction on the
// first user message of the task thread history (RIID-5907).
func assertAIAgentV2SmokeThreadReactions(t *testing.T, server http.Handler, base, token, taskID string) {
	t.Helper()
	historyPath := strings.Replace(base, "/v2/", "/v3/", 1) + "/tasks/" + taskID + "/threads"
	var history AIAgentTaskThreadHistoryCollectionResponse
	aiAgentSmokeDecode(t, aiAgentSmokeRequest(t, server, http.MethodGet, historyPath, token, "", http.StatusOK), &history)
	threadID, messageID := "", ""
	for _, thread := range history.Threads {
		for _, message := range thread.Messages {
			if message.Role == AIAgentTaskThreadMessageRoleUser && messageID == "" {
				threadID, messageID = thread.ThreadID, message.MessageID
			}
		}
	}
	if messageID == "" {
		t.Fatalf("v2 smoke history has no user message: %+v", history.Threads)
	}
	reactions := base + "/tasks/" + taskID + "/threads/" + threadID + "/messages/" + messageID + "/reactions"
	var added AIAgentTaskThreadMessageReactionsResponse
	aiAgentSmokeDecode(t, aiAgentSmokeRequest(t, server, http.MethodPost, reactions, token, `{"emoji":"1f44d"}`, http.StatusOK), &added)
	if added.MessageID != messageID || !reactionsHaveMember(added.Reactions, "1f44d", "user-1") {
		t.Fatalf("v2 smoke reaction add = %+v", added)
	}
	var deleted AIAgentTaskThreadMessageReactionsResponse
	aiAgentSmokeDecode(t, aiAgentSmokeRequest(t, server, http.MethodDelete, reactions+"/1f44d", token, "", http.StatusOK), &deleted)
	if reactionsHaveMember(deleted.Reactions, "1f44d", "user-1") {
		t.Fatalf("v2 smoke reaction delete = %+v", deleted)
	}
}

func reactionsHaveMember(reactions []AIAgentMessageReaction, emoji, principalID string) bool {
	for _, reaction := range reactions {
		for _, actor := range reaction.Actors {
			if reaction.Emoji == emoji && actor.Kind == AIAgentMessageReactionActorMember && actor.ID == principalID {
				return true
			}
		}
	}
	return false
}
