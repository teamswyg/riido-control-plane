package riidoaiserver

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func heartbeatEncodingSnapshot() AIAgentClientSnapshot {
	snapshot := snapshotTestRecord(fixedSnapshotTestNow())
	rng := rand.New(rand.NewSource(42))
	for i := range 160 {
		body := make([]byte, 4096)
		for j := range body {
			body[j] = byte('a' + rng.Intn(26))
		}
		taskID := fmt.Sprintf("task-%d", i)
		snapshot.TaskThreads[taskID] = []AIAgentTaskThreadRecord{{TaskID: taskID, Message: string(body)}}
	}
	snapshot.Daemons = []DeviceDaemonRecord{{DeviceID: "device-a", DaemonID: "daemon-a", LastSeenAt: snapshot.SavedAt}}
	return snapshot
}

func BenchmarkHeartbeatSnapshotEncoding(b *testing.B) {
	snapshot := heartbeatEncodingSnapshot()
	store := &DynamoDBAIAgentClientSnapshot{}
	_, hashes, _, err := store.encodeSnapshot(snapshot, nil)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		snapshot.Daemons[0].LastSeenAt = snapshot.Daemons[0].LastSeenAt.Add(time.Second)
		_, hashes, _, err = store.encodeSnapshot(snapshot, hashes)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestHeartbeatEncodingCachePreservesCompleteAndDeltaWrites(t *testing.T) {
	snapshot := heartbeatEncodingSnapshot()
	store := &DynamoDBAIAgentClientSnapshot{}
	var committed map[string]string
	for i := range 4 {
		snapshot.Daemons[0].LastSeenAt = snapshot.Daemons[0].LastSeenAt.Add(time.Second)
		if i == 2 {
			snapshot.TaskThreads["task-0"][0].Message = "new result"
		}
		want, wantHashes, wantBytes, err := encodeSplitDynamoDBAIAgentClientSnapshot(snapshot, committed)
		if err != nil {
			t.Fatal(err)
		}
		got, hashes, bytes, err := store.encodeSnapshot(snapshot, committed)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) || !reflect.DeepEqual(hashes, wantHashes) || bytes != wantBytes {
			t.Fatalf("iteration %d: cached encoding changed writes, manifest hashes or byte accounting", i)
		}
		if len(store.partEncodings) != len(dynamoDBAIAgentClientSnapshotPartNames) {
			t.Fatalf("cache must retain exactly one encoding per part, got %d", len(store.partEncodings))
		}
		// A failed write or a different durable writer must not let cached
		// encodings become evidence that a part can be omitted from a retry.
		if i%2 == 0 {
			committed = hashes
		} else {
			committed = nil
		}
	}
}
