package riidoaiserver

import (
	"encoding/json"
	"errors"
)

var errReadScopeCaptureIncomplete = errors.New("read scope capture incomplete")

type taskReadEvidenceIndex struct {
	scope    *taskReadScope
	complete bool
	values   map[string]json.RawMessage
	selected map[string]string
}

func newTaskReadEvidenceIndex(evidence *taskReadScopeEvidence, id, team string) (*taskReadEvidenceIndex, error) {
	invalid := taskContextReadError{"invalid"}
	if evidence == nil || evidence.SchemaVersion != "riido-read-scope-capture.v1" || evidence.Scope == nil || evidence.Complete == nil || evidence.BodyPages == nil || evidence.CommentPages == nil || evidence.MeetingPages == nil {
		return nil, invalid
	}
	raw, err := json.Marshal(evidence)
	if err != nil || len(raw) > 1<<20 {
		return nil, invalid
	}
	if err := validateTaskReadScope(*evidence.Scope, id, team); err != nil {
		return nil, err
	}
	index := &taskReadEvidenceIndex{scope: evidence.Scope, complete: *evidence.Complete, values: map[string]json.RawMessage{}, selected: map[string]string{}}
	for _, work := range evidence.Scope.Works {
		index.selected[string(work.Kind)+"\x00"+work.ID] = work.TeamID
	}
	for _, meeting := range evidence.Scope.Meetings {
		index.selected["meeting\x00"+meeting.ID] = meeting.TeamID
	}
	for _, work := range evidence.Scope.Works {
		for _, page := range []*taskReadBodyPage{work.Body, work.DescriptionBody, work.AlternateBody, work.ContentPage} {
			if page != nil {
				if err := index.addBody(work.Kind, page); err != nil {
					return nil, err
				}
			}
		}
		if err := index.addComments(work.Comments); err != nil {
			return nil, err
		}
	}
	for _, captured := range evidence.MeetingPages {
		if err := index.addMeeting(captured); err != nil {
			return nil, err
		}
	}
	for _, captured := range evidence.BodyPages {
		if err := index.addBody(captured.ResourceKind, captured.Page); err != nil {
			return nil, err
		}
	}
	for _, page := range evidence.CommentPages {
		if err := index.addComments(page); err != nil {
			return nil, err
		}
	}
	return index, nil
}
