package main

import "time"

type record struct {
	SchemaVersion            string           `json:"schema_version"`
	ObservedAt               time.Time        `json:"observed_at"`
	BaseURL                  string           `json:"base_url"`
	StatusURL                string           `json:"status_url"`
	PagesStatusURL           string           `json:"pages_status_url"`
	BadgeURL                 string           `json:"badge_url"`
	HTMLReachable            bool             `json:"html_reachable"`
	StatusOverall            string           `json:"status_overall"`
	StatusVisibility         string           `json:"status_visibility"`
	StatusRawLogsIncluded    bool             `json:"status_raw_logs_included"`
	StatusSecretsIncluded    bool             `json:"status_secrets_included"`
	StatusEndpointDetails    string           `json:"status_endpoint_details"`
	StatusSourceCommit       string           `json:"status_source_commit"`
	StatusSourceRunID        string           `json:"status_source_run_id"`
	StatusSourceRunURL       string           `json:"status_source_run_url"`
	StatusBlockingCategories []statusCategory `json:"status_blocking_categories"`
	PagesStatus              string           `json:"pages_status"`
	PagesVisibility          string           `json:"pages_visibility"`
	PagesBuildType           string           `json:"pages_build_type"`
	PagesSourceCommit        string           `json:"pages_source_commit"`
	PagesSourceRunID         string           `json:"pages_source_run_id"`
	PagesSourceRunURL        string           `json:"pages_source_run_url"`
	PagesRawResponseIncluded bool             `json:"pages_raw_response_included"`
	PagesSecretsIncluded     bool             `json:"pages_secrets_included"`
	BadgeSchemaVersion       int              `json:"badge_schema_version"`
	BadgeLabel               string           `json:"badge_label"`
	BadgeMessage             string           `json:"badge_message"`
	BadgeColor               string           `json:"badge_color"`
	RawBodiesIncluded        bool             `json:"raw_bodies_included"`
	SecretsIncluded          bool             `json:"secrets_included"`
	Passed                   bool             `json:"passed"`
}
