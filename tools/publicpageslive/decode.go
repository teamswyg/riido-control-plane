package main

import "encoding/json"

type statusPayload struct {
	Overall            string           `json:"overall"`
	Visibility         string           `json:"visibility"`
	RawLogsIncluded    bool             `json:"raw_logs_included"`
	SecretsIncluded    bool             `json:"secrets_included"`
	EndpointDetails    string           `json:"endpoint_details"`
	SourceCommit       string           `json:"source_commit"`
	SourceRunID        string           `json:"source_run_id"`
	SourceRunURL       string           `json:"source_run_url"`
	BlockingCategories []statusCategory `json:"blocking_categories"`
}

type statusCategory struct {
	Category          string `json:"category"`
	PartialCount      int    `json:"partial_count"`
	StalePartialCount int    `json:"stale_partial_count"`
}

type pagesPayload struct {
	Status              string `json:"status"`
	Visibility          string `json:"visibility"`
	BuildType           string `json:"build_type"`
	SourceCommit        string `json:"source_commit"`
	SourceRunID         string `json:"source_run_id"`
	SourceRunURL        string `json:"source_run_url"`
	RawResponseIncluded bool   `json:"raw_response_included"`
	SecretsIncluded     bool   `json:"secrets_included"`
}

type badgePayload struct {
	SchemaVersion int    `json:"schemaVersion"`
	Label         string `json:"label"`
	Message       string `json:"message"`
	Color         string `json:"color"`
}

func decodeStatus(
	statusBody, pagesBody, badgeBody []byte,
) (statusPayload, pagesPayload, badgePayload, error) {
	var status statusPayload
	if err := json.Unmarshal(statusBody, &status); err != nil {
		return statusPayload{}, pagesPayload{}, badgePayload{}, err
	}
	var pages pagesPayload
	if err := json.Unmarshal(pagesBody, &pages); err != nil {
		return statusPayload{}, pagesPayload{}, badgePayload{}, err
	}
	var badge badgePayload
	if err := json.Unmarshal(badgeBody, &badge); err != nil {
		return statusPayload{}, pagesPayload{}, badgePayload{}, err
	}
	return status, pages, badge, nil
}
