package riidoaiserver

import (
	"context"
	"fmt"
	"testing"
)

func TestConversationOwnedSaveRejectsMissingOrCorruptThreadChunk(t *testing.T) {
	for _, corrupt := range []bool{false, true} {
		t.Run(fmt.Sprint(corrupt), func(t *testing.T) {
			f := newSnapshotDynamoDBFixture(t, fixedSnapshotTestNow(), nil, nil)
			defer f.close()
			want := chunkedThreadSnapshotForTest(t)
			if err := f.store.SaveAIAgentClientSnapshot(t.Context(), want); err != nil {
				t.Fatal(err)
			}
			f.mu.Lock()
			part := f.items[dynamoDBAIAgentClientSnapshotPartSK(dynamoDBAIAgentClientSnapshotPartTaskThreads)]
			hashes, err := conversationChunkHashes(part)
			f.mu.Unlock()
			if err != nil || len(hashes) == 0 {
				t.Fatalf("missing thread chunks: %v", err)
			}
			f.mu.Lock()
			var chunkKey string
			var chunk map[string]map[string]string
			for key, item := range f.items {
				if dynamoDBStringValue(item, "pk") == conversationChunkPK && dynamoDBStringValue(item, "sk") == hashes[0] {
					chunkKey, chunk = key, item
					break
				}
			}
			body := dynamoDBStringValue(chunk, "body")
			if corrupt {
				chunk["body"]["S"] = "corrupt"
			} else {
				delete(f.items, chunkKey)
			}
			f.mu.Unlock()
			drainSnapshotRequests(f.requests)
			ctx := context.WithValue(t.Context(), taskAgentPreserveMembershipContextKey{}, true)
			if err := f.store.SaveAIAgentClientSnapshot(ctx, want); err == nil {
				t.Fatal("owned save accepted incomplete committed history")
			}
			for len(f.requests) > 0 {
				request := <-f.requests
				if target := request.header.Get("X-Amz-Target"); target == dynamoDBPutItemTarget || target == dynamoDBTransactWriteTarget {
					t.Fatal("failed history verification wrote state")
				}
			}
			f.mu.Lock()
			chunk["body"]["S"] = body
			f.items[chunkKey] = chunk
			f.mu.Unlock()
			if err := f.store.SaveAIAgentClientSnapshot(ctx, want); err != nil {
				t.Fatalf("retry after restoring the original chunk: %v", err)
			}
		})
	}
}
