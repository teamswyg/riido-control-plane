package riidoaiserver

import "encoding/json"

func readScopeWorkSummary(work *taskReadScopeWork) (AssignmentScopedResource, error) {
	if work.ContentPage != nil {
		if work.Summary == nil || work.Summary.ID != work.ID || work.Summary.Kind != work.Kind {
			return AssignmentScopedResource{}, taskContextReadError{"invalid"}
		}
		key := ""
		if work.Summary.KeyNumber != nil {
			key = *work.Summary.KeyNumber
		}
		return AssignmentScopedResource{Kind: work.Kind, ID: work.ID, TeamID: work.TeamID, Title: work.Summary.Title, KeyNumber: key, UpdatedAt: work.Summary.UpdatedAt, Reasons: work.Reasons}, nil
	}
	details := work.Details
	if work.Kind == AssignmentContentBacklog {
		details = work.BacklogDetails
	}
	if work.Kind == AssignmentContentIntake {
		details = work.IntakeDetails
	}
	var value struct {
		ID        string
		Title     string
		KeyNumber string
		UpdatedAt string
	}
	if json.Unmarshal(details, &value) != nil || value.ID != work.ID {
		return AssignmentScopedResource{}, taskContextReadError{"invalid"}
	}
	return AssignmentScopedResource{Kind: work.Kind, ID: work.ID, TeamID: work.TeamID, Title: value.Title, KeyNumber: value.KeyNumber, UpdatedAt: value.UpdatedAt, Reasons: work.Reasons}, nil
}

func readScopeStatus(status string) bool {
	return status == "available" || status == "empty" || status == "partial" || status == "unavailable"
}
