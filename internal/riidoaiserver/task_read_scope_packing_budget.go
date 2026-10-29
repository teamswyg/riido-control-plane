package riidoaiserver

import "encoding/json"

type taskReadScopePackingBudget struct{ blocks, bytes, entries int }

// Encoding a pages array is additive: one envelope, encoded pages and commas.
// The persisted bundle is independently packed and verified before publication.
func (b *taskReadScopePackingBudget) appendPage(page AssignmentContentPage) bool {
	encoded, err := json.Marshal(page)
	if err != nil {
		return false
	}
	single := len(`{"pages":[]}`) + len(encoded)
	if single > assignmentContentMaxPageBytes {
		return false
	}
	if b.blocks == 0 {
		b.blocks, b.bytes, b.entries = 1, single, 1
		return true
	}
	if candidate := b.bytes + 1 + len(encoded); candidate <= assignmentContentMaxPageBytes {
		b.bytes = candidate
		b.entries++
		return true
	}
	if b.blocks >= assignmentContextMaxCommentPages {
		return false
	}
	b.blocks++
	b.bytes, b.entries = single, 1
	return true
}
