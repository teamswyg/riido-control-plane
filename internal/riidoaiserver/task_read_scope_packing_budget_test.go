package riidoaiserver

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTaskReadScopePackingBudgetMatchesActualJSONBlocks(t *testing.T) {
	var budget taskReadScopePackingBudget
	var blocks [][]AssignmentContentPage
	sizes := []int{5, 9000, 50000, 9000, 100, 70000, 1}
	for range 12 {
		sizes = append(sizes, 60000)
	}
	sizes = append(sizes, 1)
	for i, size := range sizes {
		data, _ := json.Marshal(map[string]string{"content": strings.Repeat("x", size) + "<한글>\n"})
		page := AssignmentContentPage{ResourceKind: AssignmentContentComponent, SourceID: "work", Part: AssignmentContentEditor, Version: "v", Cursor: "cursor", Offset: i, Status: "available", Data: data, Complete: i%2 == 0, Sources: []AssignmentContextSource{{Kind: "task", ID: "work", Part: "body"}}}
		trial, ok := packingBudgetOracle(t, blocks, page)
		before := budget
		got := budget.appendPage(page)
		if got != ok {
			t.Fatalf("actual JSON fit mismatch index=%d got=%v want=%v", i, got, ok)
		}
		if !got {
			if before != budget {
				t.Fatal("rejected append changed budget")
			}
			continue
		}
		blocks = trial
		encoded, _ := json.Marshal(assignmentContentBlock{Pages: blocks[len(blocks)-1]})
		if budget.blocks != len(blocks) || budget.bytes != len(encoded) || budget.entries != len(blocks[len(blocks)-1]) {
			t.Fatalf("accepted serialized size differs: %+v actual bytes%d blocks%d", budget, len(encoded), len(blocks))
		}
	}
}

func packingBudgetOracle(t *testing.T, blocks [][]AssignmentContentPage, page AssignmentContentPage) ([][]AssignmentContentPage, bool) {
	t.Helper()
	trial := append([][]AssignmentContentPage(nil), blocks...)
	if len(trial) == 0 {
		trial = append(trial, []AssignmentContentPage{})
	}
	last := len(trial) - 1
	candidate := append(append([]AssignmentContentPage(nil), trial[last]...), page)
	encoded, err := json.Marshal(assignmentContentBlock{Pages: candidate})
	if err != nil {
		t.Fatal(err)
	}
	if len(encoded) <= 64<<10 {
		trial[last] = candidate
		return trial, true
	}
	single, _ := json.Marshal(assignmentContentBlock{Pages: []AssignmentContentPage{page}})
	if len(single) > 64<<10 || len(trial) >= 10 {
		return nil, false
	}
	return append(trial, []AssignmentContentPage{page}), true
}
