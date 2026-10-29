package riidoaiserver

import (
	"bytes"
	"encoding/json"
	"io"
)

type taskReadWorkMetadata struct {
	Sprints        []json.RawMessage         `json:"sprints"`
	Origin         json.RawMessage           `json:"origin"`
	Details        json.RawMessage           `json:"details"`
	BacklogDetails json.RawMessage           `json:"backlogDetails"`
	IntakeDetails  json.RawMessage           `json:"intakeDetails"`
	Links          []json.RawMessage         `json:"links"`
	GitHub         []json.RawMessage         `json:"github"`
	Repositories   []taskReadScopeRepository `json:"repositories"`
	Figma          json.RawMessage           `json:"figma"`
	Priorities     []json.RawMessage         `json:"priorities"`
	Checklists     []json.RawMessage         `json:"checklists"`
	Attachments    []json.RawMessage         `json:"attachments"`
}

type taskReadMeetingMetadata struct {
	Details json.RawMessage   `json:"details"`
	Labels  []json.RawMessage `json:"labels"`
}

func decodeReadScopeSlot(data []byte, out any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(out) != nil {
		return taskContextReadError{"invalid"}
	}
	if decoder.Decode(new(any)) != io.EOF {
		return taskContextReadError{"invalid"}
	}
	return nil
}

func applyReadScopeWorkMetadata(work *taskReadScopeWork, data []byte) error {
	var value taskReadWorkMetadata
	if decodeReadScopeSlot(data, &value) != nil || value.Links == nil || value.GitHub == nil || value.Repositories == nil || value.Priorities == nil || value.Checklists == nil || value.Attachments == nil || value.Sprints == nil {
		return taskContextReadError{"invalid"}
	}
	candidate := *work
	candidate.Details, candidate.BacklogDetails, candidate.IntakeDetails = value.Details, value.BacklogDetails, value.IntakeDetails
	candidate.Links, candidate.GitHub, candidate.Repositories, candidate.Figma = value.Links, value.GitHub, value.Repositories, value.Figma
	candidate.Priorities, candidate.Checklists, candidate.Attachments = value.Priorities, value.Checklists, value.Attachments
	candidate.Sprints, candidate.Origin = value.Sprints, value.Origin
	candidate.ContentPage = nil
	if _, err := readScopeWorkSummary(&candidate); err != nil {
		return err
	}
	*work = candidate
	completeDeferredMetadata(work.PartsStatus)
	return nil
}

func completeDeferredMetadata(parts []*taskReadPartStatus) {
	for _, part := range parts {
		if part == nil || part.Status != "deferred" {
			continue
		}
		switch part.Name {
		case "comments", "board", "document", "description", "document_alternate":
			continue
		}
		complete := true
		part.Status = "available"
		part.Complete = &complete
	}
}
