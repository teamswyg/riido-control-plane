package riidoaiserver

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
)

type AssignmentContentBlockRef struct {
	FirstIndex int    `json:"first_index"`
	Count      int    `json:"count"`
	Digest     string `json:"digest"`
}

type assignmentContentBlock struct {
	Pages []AssignmentContentPage `json:"pages"`
}

func assignmentContentDigest(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func packAssignmentContentPages(pages []AssignmentContentPage) ([]assignmentContentBlock, []AssignmentContentBlockRef, error) {
	var blocks []assignmentContentBlock
	var refs []AssignmentContentBlockRef
	current := assignmentContentBlock{}
	first := 0
	flush := func() error {
		if len(current.Pages) == 0 {
			return nil
		}
		if len(blocks) >= assignmentContextMaxCommentPages {
			return errors.New("assignment content physical block budget exceeded")
		}
		encoded, err := json.Marshal(current)
		if err != nil {
			return err
		}
		refs = append(refs, AssignmentContentBlockRef{first, len(current.Pages), assignmentContentDigest(encoded)})
		blocks = append(blocks, current)
		first += len(current.Pages)
		current = assignmentContentBlock{}
		return nil
	}
	for _, page := range pages {
		candidate := assignmentContentBlock{append(append([]AssignmentContentPage(nil), current.Pages...), page)}
		encoded, err := json.Marshal(candidate)
		if err != nil {
			return nil, nil, err
		}
		if len(encoded) > assignmentContentMaxPageBytes {
			if err = flush(); err != nil {
				return nil, nil, err
			}
			candidate = assignmentContentBlock{[]AssignmentContentPage{page}}
			encoded, err = json.Marshal(candidate)
			if err != nil {
				return nil, nil, err
			}
			if len(encoded) > assignmentContentMaxPageBytes {
				return nil, nil, errors.New("assignment content logical page exceeds block budget")
			}
		}
		current = candidate
	}
	err := flush()
	return blocks, refs, err
}
