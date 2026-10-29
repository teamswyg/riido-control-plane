package riidoaiserver

import "errors"

func assignmentContentBlockFor(snapshot AssignmentContextSnapshot, index int) (int, AssignmentContentBlockRef, bool, error) {
	total := 0
	for _, part := range snapshot.Content {
		total += len(part.Cursors)
	}
	if index < 0 || index >= total {
		return 0, AssignmentContentBlockRef{}, false, nil
	}
	first, selected := 0, -1
	if len(snapshot.ContentBlocks) > assignmentContextMaxCommentPages {
		return 0, AssignmentContentBlockRef{}, false, errors.New("assignment content block count invalid")
	}
	for i, ref := range snapshot.ContentBlocks {
		if ref.FirstIndex != first || ref.Count <= 0 || ref.Count > total-first || len(ref.Digest) != len("sha256:")+64 {
			return 0, AssignmentContentBlockRef{}, false, errors.New("assignment content block layout invalid")
		}
		if index >= first && index < first+ref.Count {
			selected = i
		}
		first += ref.Count
	}
	if first != total || selected < 0 {
		return 0, AssignmentContentBlockRef{}, false, errors.New("assignment content block layout incomplete")
	}
	return selected, snapshot.ContentBlocks[selected], true, nil
}
