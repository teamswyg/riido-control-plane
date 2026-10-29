package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
)

func (s *DynamoDBAssignmentOperationStore) LoadAssignmentContentPage(ctx context.Context, id string, index int) (AssignmentContentPage, bool, error) {
	snapshot, found, err := s.LoadAssignmentContextSnapshot(ctx, id)
	if err != nil || !found {
		return AssignmentContentPage{}, false, err
	}
	blockIndex, ref, found, err := assignmentContentBlockFor(snapshot, index)
	if err != nil || !found {
		return AssignmentContentPage{}, false, err
	}
	item, err := s.contextRetentionItem(ctx, id, assignmentContentSK(blockIndex))
	if err != nil || len(item) == 0 {
		return AssignmentContentPage{}, false, err
	}
	if expired, err := assignmentContextItemExpired(item, s.now()); expired || err != nil {
		return AssignmentContentPage{}, false, err
	}
	if item["schema"]["S"] != "assignment_context_content_block/v1" {
		return AssignmentContentPage{}, false, errors.New("assignment content schema mismatch")
	}
	decoded, err := gunzipBase64(item["payload_gzip"]["S"])
	if err != nil {
		return AssignmentContentPage{}, false, err
	}
	if assignmentContentDigest(decoded) != ref.Digest {
		return AssignmentContentPage{}, false, errors.New("assignment content block digest mismatch")
	}
	var block assignmentContentBlock
	if err = json.Unmarshal(decoded, &block); err != nil || len(block.Pages) != ref.Count {
		return AssignmentContentPage{}, false, errors.New("assignment content block entry mismatch")
	}
	page := block.Pages[index-ref.FirstIndex]
	err = validateAssignmentContentPage(page)
	return page, err == nil, err
}
