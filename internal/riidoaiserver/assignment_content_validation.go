package riidoaiserver

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
)

const assignmentContentMaxPageBytes = 64 << 10

func validateAssignmentContentBundle(snapshot AssignmentContextSnapshot, pages []AssignmentContentPage) error {
	if err := validateAssignmentReadScopeManifest(snapshot); err != nil {
		return err
	}
	if strings.TrimSpace(snapshot.AssignmentID) == "" || snapshot.CommentPages != 0 {
		return errors.New("assignment content identity or shared payload budget invalid")
	}
	encoded, err := json.Marshal(snapshot)
	if err != nil || len(encoded) > contextSnapshotItemBytes-128 {
		return errors.New("assignment content manifest budget invalid")
	}
	byCursor := map[string]AssignmentContentPage{}
	for _, page := range pages {
		if err := validateAssignmentContentPage(page); err != nil {
			return err
		}
		if _, exists := byCursor[page.Cursor]; exists {
			return errors.New("duplicate assignment content page")
		}
		byCursor[page.Cursor] = page
	}
	seenParts := map[string]bool{}
	index := 0
	for _, part := range snapshot.Content {
		key := string(part.ResourceKind) + "\x00" + part.SourceID + "\x00" + string(part.Part)
		if part.SourceID == "" || !validAssignmentContentKind(part.ResourceKind) || !validAssignmentContentPart(part.Part) || seenParts[key] {
			return errors.New("assignment content manifest identity invalid")
		}
		seenParts[key] = true
		previous := -1
		for i, cursor := range part.Cursors {
			page, exists := byCursor[cursor]
			if !exists || index >= len(pages) || pages[index].Cursor != cursor || page.ResourceKind != part.ResourceKind || page.SourceID != part.SourceID || page.Part != part.Part || page.Version != part.Version || page.Offset <= previous || page.Complete != (i == len(part.Cursors)-1 && part.Complete) {
				return errors.New("assignment content manifest page mismatch")
			}
			previous = page.Offset
			index++
			delete(byCursor, cursor)
		}
		if part.Complete && len(part.Cursors) == 0 {
			return errors.New("complete assignment content has no captured page")
		}
	}
	if len(byCursor) > 0 {
		return errors.New("assignment content page outside manifest")
	}
	_, refs, err := packAssignmentContentPages(pages)
	if err != nil {
		return err
	}
	if snapshot.ContentBlocks != nil && !slices.Equal(snapshot.ContentBlocks, refs) {
		return errors.New("assignment content block layout mismatch")
	}
	snapshot.ContentBlocks = refs
	encoded, err = json.Marshal(snapshot)
	if err != nil || len(encoded) > contextSnapshotItemBytes-128 {
		return errors.New("assignment content block manifest budget invalid")
	}
	return nil
}

func validAssignmentContentPart(part AssignmentContentPart) bool {
	return part == AssignmentContentMetadata || part == AssignmentContentEditor || part == AssignmentContentDescription || part == AssignmentContentAlternate || part == AssignmentContentComments || part == AssignmentContentBoard
}
