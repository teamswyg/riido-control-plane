package main

import "encoding/json"

type statusPayload struct {
	Overall         string `json:"overall"`
	Visibility      string `json:"visibility"`
	RawLogsIncluded bool   `json:"raw_logs_included"`
	SecretsIncluded bool   `json:"secrets_included"`
	EndpointDetails string `json:"endpoint_details"`
}

type pagesPayload struct {
	Status              string `json:"status"`
	Visibility          string `json:"visibility"`
	BuildType           string `json:"build_type"`
	RawResponseIncluded bool   `json:"raw_response_included"`
	SecretsIncluded     bool   `json:"secrets_included"`
}

func decodeStatus(statusBody, pagesBody []byte) (statusPayload, pagesPayload, error) {
	var status statusPayload
	if err := json.Unmarshal(statusBody, &status); err != nil {
		return statusPayload{}, pagesPayload{}, err
	}
	var pages pagesPayload
	if err := json.Unmarshal(pagesBody, &pages); err != nil {
		return statusPayload{}, pagesPayload{}, err
	}
	return status, pages, nil
}
