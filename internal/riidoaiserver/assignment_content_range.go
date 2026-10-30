package riidoaiserver

import (
	"context"
	"errors"
)

// VisitAssignmentContentPages reads the snapshot once and each physical block
// once, verifying the block digest exactly as a single-page load does. A page
// that cannot be read after the first one ends the visit without an error.
func (s *DynamoDBAssignmentOperationStore) VisitAssignmentContentPages(ctx context.Context, id string, first, count int, visit func(AssignmentContentPage) bool) error {
	snapshot, found, err := s.LoadAssignmentContextSnapshot(ctx, id)
	if err != nil || !found {
		return err
	}
	cache := assignmentContentBlockCache{loaded: -1}
	var firstErr error
	for index := first; index < first+count; index++ {
		page, readErr := s.cachedAssignmentContentPage(ctx, id, snapshot, index, &cache)
		if readErr != nil {
			if index == first {
				firstErr = readErr
			}
			break
		}
		if !visit(page) {
			break
		}
	}
	return firstErr
}

type assignmentContentBlockCache struct {
	loaded int
	block  assignmentContentBlock
	ref    AssignmentContentBlockRef
}

var errAssignmentContentPageMissing = errors.New("assignment content page unavailable")

func (s *DynamoDBAssignmentOperationStore) cachedAssignmentContentPage(ctx context.Context, id string, snapshot AssignmentContextSnapshot, index int, cache *assignmentContentBlockCache) (AssignmentContentPage, error) {
	blockIndex, ref, found, err := assignmentContentBlockFor(snapshot, index)
	if err != nil {
		return AssignmentContentPage{}, err
	}
	if found && blockIndex != cache.loaded {
		cache.block, found, err = s.loadAssignmentContentBlock(ctx, id, blockIndex, ref)
		if err != nil {
			return AssignmentContentPage{}, err
		}
		cache.loaded, cache.ref = blockIndex, ref
	}
	if !found {
		return AssignmentContentPage{}, errAssignmentContentPageMissing
	}
	page := cache.block.Pages[index-cache.ref.FirstIndex]
	return page, validateAssignmentContentPage(page)
}
