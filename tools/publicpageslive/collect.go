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
	badgeURL := base + "status-badge.json"
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
	badgeBody, err := get(client, badgeURL)
	if err != nil {
		return record{}, err
	}
	if err := assertNoSecretMarkers(html, statusBody, pagesBody, badgeBody); err != nil {
		return record{}, err
	}
	status, pages, badge, err := decodeStatus(statusBody, pagesBody, badgeBody)
	if err != nil {
		return record{}, err
	}
	rec := newRecord(base, now, statusURL, pagesURL, badgeURL, status, pages, badge)
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
		(rec.StatusOverall != "degraded" || len(rec.StatusBlockingCategories) > 0) &&
		rec.PagesSourceCommit != "" && rec.PagesSourceRunID != "" &&
		rec.BadgeSchemaVersion == 1 && rec.BadgeLabel != "" && rec.BadgeMessage != "" &&
		!rec.PagesSecretsIncluded && !rec.RawBodiesIncluded && !rec.SecretsIncluded
}
