package riidoaiserver

import (
	"context"
	"encoding/json"
	"errors"
)

func (s *DynamoDBAssignmentOperationStore) loadAssignmentContentBlock(ctx context.Context, id string, blockIndex int, ref AssignmentContentBlockRef) (assignmentContentBlock, bool, error) {
	item, err := s.contextRetentionItem(ctx, id, assignmentContentSK(blockIndex))
	if err != nil || len(item) == 0 {
		return assignmentContentBlock{}, false, err
	}
	if expired, err := assignmentContextItemExpired(item, s.now()); expired || err != nil {
		return assignmentContentBlock{}, false, err
	}
	if item["schema"]["S"] != "assignment_context_content_block/v1" {
		return assignmentContentBlock{}, false, errors.New("assignment content schema mismatch")
	}
	decoded, err := gunzipBase64(item["payload_gzip"]["S"])
	if err != nil {
		return assignmentContentBlock{}, false, err
	}
	if assignmentContentDigest(decoded) != ref.Digest {
		return assignmentContentBlock{}, false, errors.New("assignment content block digest mismatch")
	}
	var block assignmentContentBlock
	if err = json.Unmarshal(decoded, &block); err != nil || len(block.Pages) != ref.Count {
		return assignmentContentBlock{}, false, errors.New("assignment content block entry mismatch")
	}
	return block, true, nil
}
