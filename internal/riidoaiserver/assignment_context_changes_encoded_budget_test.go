package riidoaiserver

import (
	"encoding/json"
	"testing"
)

func TestContextChangesMetadataPreservesEncodedDynamoBudget(t *testing.T) {
	snapshot, pages := contextChangesTestBundle(t, contextChangesTestSnapshot("encoded-boundary"), "body")
	digest := snapshot.Content[0].Digest
	snapshot.Content[0].Digest, snapshot.Context.ReadScope.Resources[0].Parts[0].Digest = "", ""
	clearAssignmentContextLineage(&snapshot)
	// Deterministic, high-entropy JSON-safe text exercises gzip/base64 expansion.
	noise := make([]byte, contextSnapshotItemBytes)
	alphabet := []byte{}
	for b := byte(33); b < 127; b++ {
		if b != 34 && b != 92 && b != 38 && b != 60 && b != 62 {
			alphabet = append(alphabet, b)
		}
	}
	state := uint32(1234567)
	for i := range noise {
		state ^= state << 13
		state ^= state >> 17
		state ^= state << 5
		noise[i] = alphabet[state%uint32(len(alphabet))]
	}
	low, high := 0, len(noise)
	for low+1 < high {
		mid := (low + high) / 2
		snapshot.Context.Document.Content = string(noise[:mid])
		if _, err := assignmentContentItems(snapshot, pages); err == nil {
			low = mid
		} else {
			high = mid
		}
	}
	snapshot.Context.Document.Content = string(noise[:low])
	if err := validateAssignmentContentBundle(snapshot, pages); err != nil {
		t.Fatal("old raw bundle invalid", err)
	}
	if _, err := assignmentContentItems(snapshot, pages); err != nil {
		t.Fatal("old stored bundle invalid", err)
	}
	snapshot.Content[0].Digest, snapshot.Context.ReadScope.Resources[0].Parts[0].Digest = digest, digest
	snapshot.ConversationID, snapshot.ThreadID = "conversation", "thread"
	raw, _ := json.Marshal(snapshot)
	if len(raw) > contextSnapshotItemBytes-128 {
		t.Fatal("fixture must pass raw size gate")
	}
	if _, err := assignmentContentItems(snapshot, pages); err == nil {
		t.Fatal("fixture must exceed encoded budget with new metadata")
	}
	bounded := boundAssignmentComparisonMetadata(snapshot, pages)
	if _, err := assignmentContentItems(bounded, pages); err != nil {
		t.Fatal("comparison metadata discards valid Dynamo capture", err)
	}
}
