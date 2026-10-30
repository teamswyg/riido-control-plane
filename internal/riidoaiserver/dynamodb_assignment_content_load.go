package riidoaiserver

import "context"

func (s *DynamoDBAssignmentOperationStore) LoadAssignmentContentPage(ctx context.Context, id string, index int) (AssignmentContentPage, bool, error) {
	snapshot, found, err := s.LoadAssignmentContextSnapshot(ctx, id)
	if err != nil || !found {
		return AssignmentContentPage{}, false, err
	}
	blockIndex, ref, found, err := assignmentContentBlockFor(snapshot, index)
	if err != nil || !found {
		return AssignmentContentPage{}, false, err
	}
	block, found, err := s.loadAssignmentContentBlock(ctx, id, blockIndex, ref)
	if err != nil || !found {
		return AssignmentContentPage{}, false, err
	}
	page := block.Pages[index-ref.FirstIndex]
	err = validateAssignmentContentPage(page)
	return page, err == nil, err
}
