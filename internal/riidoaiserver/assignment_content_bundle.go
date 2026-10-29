package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"errors"
)

var errAssignmentContentConflict = errors.New("assignment content bundle conflict")

type assignmentContentBundle struct {
	Snapshot AssignmentContextSnapshot
	Pages    []AssignmentContentPage
}

func cloneAssignmentContentBundle(snapshot AssignmentContextSnapshot, pages []AssignmentContentPage) (assignmentContentBundle, error) {
	_, refs, err := packAssignmentContentPages(pages)
	if err != nil {
		return assignmentContentBundle{}, err
	}
	snapshot.ContentBlocks = refs
	encoded, err := json.Marshal(assignmentContentBundle{snapshot, pages})
	if err != nil {
		return assignmentContentBundle{}, err
	}
	var out assignmentContentBundle
	err = json.Unmarshal(encoded, &out)
	return out, err
}

func equalAssignmentContentBundle(previous AssignmentContextSnapshot, oldPages []AssignmentContentPage, snapshot AssignmentContextSnapshot, pages []AssignmentContentPage) error {
	old, err := json.Marshal(assignmentContentBundle{previous, oldPages})
	if err != nil {
		return err
	}
	canonical, err := cloneAssignmentContentBundle(snapshot, pages)
	if err != nil {
		return err
	}
	next, err := json.Marshal(canonical)
	if err != nil {
		return err
	}
	if !bytes.Equal(old, next) {
		return errAssignmentContentConflict
	}
	return nil
}
