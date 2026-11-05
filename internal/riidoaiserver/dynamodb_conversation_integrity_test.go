package riidoaiserver

import (
	"fmt"
	"testing"
)

func TestConversationDynamoDBRejectsMissingOrCorruptChunk(t *testing.T) {
	for _, corrupt := range []bool{false, true} {
		t.Run(fmt.Sprint(corrupt), func(t *testing.T) {
			f := newSnapshotDynamoDBFixture(t, fixedSnapshotTestNow(), nil, nil)
			defer f.close()
			if err := f.store.SaveAIAgentClientSnapshot(t.Context(), conversationSnapshotForTest(200)); err != nil {
				t.Fatal(err)
			}
			f.mu.Lock()
			for key, item := range f.items {
				if dynamoDBStringValue(item, "pk") == conversationChunkPK {
					if corrupt {
						item["body"]["S"] = "corrupt"
					} else {
						delete(f.items, key)
					}
					break
				}
			}
			f.mu.Unlock()
			if _, _, err := f.store.LoadAIAgentClientSnapshot(t.Context()); err == nil {
				t.Fatal("partial history published")
			}
		})
	}
}
