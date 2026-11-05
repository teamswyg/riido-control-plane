package riidoaiserver

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestConversationDynamoDBLargeHistoryRoundTrip(t *testing.T) {
	// Over 4 MiB of poorly compressible messages, including the 2,000-message
	// import boundary. The manifest transaction must stay independently small.
	want := conversationSnapshotForTest(2000)
	f := newSnapshotDynamoDBFixture(t, fixedSnapshotTestNow(), nil, nil)
	defer f.close()
	if err := f.store.SaveAIAgentClientSnapshot(t.Context(), want); err != nil {
		t.Fatal(err)
	}
	got, found, err := f.store.LoadAIAgentClientSnapshot(t.Context())
	if err != nil || !found || !reflect.DeepEqual(got.TaskThreadMessages, want.TaskThreadMessages) {
		t.Fatalf("large history round trip: found=%t err=%v", found, err)
	}
	chunkWrites, chunkReads := 0, 0
	for len(f.requests) > 0 {
		request := <-f.requests
		switch request.header.Get("X-Amz-Target") {
		case dynamoDBPutItemTarget:
			chunkWrites++
			if len(request.body) > 200<<10 {
				t.Fatalf("chunk item too large: %d", len(request.body))
			}
		case dynamoDBTransactWriteTarget:
			if len(request.body) > 400<<10 {
				t.Fatalf("history leaked into transaction: %d", len(request.body))
			}
		case dynamoDBSnapshotTransactGetTarget:
			var input struct {
				TransactItems []struct {
					Get struct{ Key map[string]map[string]string }
				}
			}
			_ = json.Unmarshal(request.body, &input)
			if len(input.TransactItems) == 0 || dynamoDBStringValue(input.TransactItems[0].Get.Key, "pk") != conversationChunkPK {
				continue
			}
			if len(input.TransactItems) > 10 {
				t.Fatal("chunk fetch exceeds response budget")
			}
			chunkReads++
		}
	}
	if chunkWrites < 20 || chunkReads < 3 {
		t.Fatalf("missing partitioned I/O: writes=%d reads=%d", chunkWrites, chunkReads)
	}
}
