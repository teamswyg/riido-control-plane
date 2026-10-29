package riidoaiserver

import (
	"context"
)

func (s *DynamoDBAssignmentOperationStore) SaveAssignmentContentBundle(ctx context.Context, snapshot AssignmentContextSnapshot, pages []AssignmentContentPage) error {
	if err := validateAssignmentContentBundle(snapshot, pages); err != nil {
		return err
	}
	items, err := assignmentContentItems(snapshot, pages)
	if err != nil {
		return err
	}
	if s.enableAssignmentContextRetention {
		err = s.saveContextPayloadsWithPolicy(ctx, snapshot.AssignmentID, items, true)
	} else {
		actions := append(s.contextPublicationFences(snapshot.AssignmentID, items, true), s.contextPayloadPuts(items, true)...)
		_, err = s.assignmentContextRequest(ctx, dynamoDBTransactWriteTarget, struct{ TransactItems []dynamoDBTransactWriteItem }{actions})
	}
	if err == nil {
		return nil
	}
	// An ambiguous write may have committed. Only the exact original complete
	// bundle proves success; an intent-only or partial payload is never repaired.
	previous, found, readErr := s.LoadAssignmentContextSnapshot(ctx, snapshot.AssignmentID)
	if readErr != nil || !found {
		return err
	}
	oldPages := make([]AssignmentContentPage, 0, len(pages))
	for i := range pages {
		page, found, readErr := s.LoadAssignmentContentPage(ctx, snapshot.AssignmentID, i)
		if readErr != nil || !found {
			return err
		}
		oldPages = append(oldPages, page)
	}
	return equalAssignmentContentBundle(previous, oldPages, snapshot, pages)
}
