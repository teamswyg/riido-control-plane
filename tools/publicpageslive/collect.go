package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func collect(base string, now time.Time, client *http.Client) (record, error) {
	base = strings.TrimRight(base, "/") + "/"
	htmlURL := base
	statusURL := base + "status.json"
	pagesURL := base + "pages-status.json"
	html, err := get(client, htmlURL)
	if err != nil {
		return record{}, err
	}
	statusBody, err := get(client, statusURL)
	if err != nil {
		return record{}, err
	}
	pagesBody, err := get(client, pagesURL)
	if err != nil {
		return record{}, err
	}
	if err := assertNoSecretMarkers(html, statusBody, pagesBody); err != nil {
		return record{}, err
	}
	status, pages, err := decodeStatus(statusBody, pagesBody)
	if err != nil {
		return record{}, err
	}
	rec := record{
		SchemaVersion: schemaVersion, ObservedAt: now.UTC(),
		BaseURL: base, StatusURL: statusURL, PagesStatusURL: pagesURL,
		HTMLReachable: true, RawBodiesIncluded: false, SecretsIncluded: false,
		StatusOverall: status.Overall, StatusVisibility: status.Visibility,
		StatusRawLogsIncluded: status.RawLogsIncluded,
		StatusSecretsIncluded: status.SecretsIncluded,
		StatusEndpointDetails: status.EndpointDetails,
		StatusSourceCommit:    status.SourceCommit,
		StatusSourceRunID:     status.SourceRunID,
		StatusSourceRunURL:    status.SourceRunURL,
		PagesStatus:           pages.Status, PagesVisibility: pages.Visibility,
		PagesBuildType:           pages.BuildType,
		PagesSourceCommit:        pages.SourceCommit,
		PagesSourceRunID:         pages.SourceRunID,
		PagesSourceRunURL:        pages.SourceRunURL,
		PagesRawResponseIncluded: pages.RawResponseIncluded,
		PagesSecretsIncluded:     pages.SecretsIncluded,
	}
	rec.Passed = liveStatusPassed(rec)
	if !rec.Passed {
		return record{}, fmt.Errorf("public pages live evidence failed validation")
	}
	return rec, nil
}

func liveStatusPassed(rec record) bool {
	return rec.HTMLReachable && rec.StatusVisibility == "public_aggregate" &&
		!rec.StatusRawLogsIncluded && !rec.StatusSecretsIncluded &&
		rec.StatusEndpointDetails == "redacted" && rec.PagesStatus == "published" &&
		rec.PagesBuildType == "workflow" && !rec.PagesRawResponseIncluded &&
		rec.StatusSourceCommit != "" && rec.StatusSourceRunID != "" &&
		rec.PagesSourceCommit != "" && rec.PagesSourceRunID != "" &&
		!rec.PagesSecretsIncluded && !rec.RawBodiesIncluded && !rec.SecretsIncluded
}
