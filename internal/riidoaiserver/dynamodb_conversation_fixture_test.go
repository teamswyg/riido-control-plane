package riidoaiserver

import (
	"encoding/base64"
	"fmt"
	"math/rand"
)

func conversationSnapshotForTest(count int) AIAgentClientSnapshot {
	snapshot := snapshotTestRecord(fixedSnapshotTestNow())
	snapshot.Version = 1
	snapshot.TaskThreads["task"] = []AIAgentTaskThreadRecord{{TaskID: "task", ThreadID: "thread", ConversationID: "comment:root", AssignmentID: "run"}}
	messages := make([]AIAgentTaskThreadHistoryMessage, count)
	random := rand.New(rand.NewSource(5891))
	for i := range messages {
		body := make([]byte, 1700)
		_, _ = random.Read(body)
		messages[i] = AIAgentTaskThreadHistoryMessage{MessageID: fmt.Sprintf("native:%d", i), Role: AIAgentTaskThreadMessageRoleUser, Body: base64.StdEncoding.EncodeToString(body), Seq: i + 1}
	}
	snapshot.TaskThreadMessages = map[string][]AIAgentTaskThreadHistoryMessage{"thread": messages}
	return snapshot
}
