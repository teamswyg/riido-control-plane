package riidoaiserver

import (
	"context"
	"errors"
	"testing"
)

type emptyAssignmentContentStore struct{}

func (emptyAssignmentContentStore) SaveAssignmentContentBundle(context.Context, AssignmentContextSnapshot, []AssignmentContentPage) error {
	return nil
}

func (emptyAssignmentContentStore) LoadAssignmentContentPage(context.Context, string, int) (AssignmentContentPage, bool, error) {
	return AssignmentContentPage{}, false, nil
}

type emptyAssignmentContentRangeStore struct{ emptyAssignmentContentStore }

func (emptyAssignmentContentRangeStore) VisitAssignmentContentPages(context.Context, string, int, int, func(AssignmentContentPage) bool) error {
	return nil
}

func TestAssignmentContentBatchMissingFirstPageIsUnavailableNotPanic(t *testing.T) {
	part := AssignmentContentManifest{ResourceKind: AssignmentContentComponent, SourceID: "work", Part: AssignmentContentDescription, Version: "v1", Cursors: []string{"first"}}
	for _, store := range []AssignmentContentStore{emptyAssignmentContentStore{}, emptyAssignmentContentRangeStore{}} {
		_, _, err := Server{}.assignmentContentBatchView(t.Context(), store, AssignmentContextSnapshot{AssignmentID: "assignment"}, part, 0, 0, 8<<10)
		if !errors.Is(err, errAssignmentContentPageMissing) {
			t.Fatalf("%T err=%v", store, err)
		}
	}
}
