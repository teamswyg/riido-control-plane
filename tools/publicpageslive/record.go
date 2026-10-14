package main

import "time"

func newRecord(
	base string,
	now time.Time,
	statusURL string,
	pagesURL string,
	badgeURL string,
	status statusPayload,
	pages pagesPayload,
	badge badgePayload,
) record {
	return record{
		SchemaVersion: schemaVersion, ObservedAt: now.UTC(),
		BaseURL: base, StatusURL: statusURL, PagesStatusURL: pagesURL,
		BadgeURL: badgeURL, HTMLReachable: true,
		RawBodiesIncluded: false, SecretsIncluded: false,
		StatusOverall: status.Overall, StatusVisibility: status.Visibility,
		StatusRawLogsIncluded:    status.RawLogsIncluded,
		StatusSecretsIncluded:    status.SecretsIncluded,
		StatusEndpointDetails:    status.EndpointDetails,
		StatusSourceCommit:       status.SourceCommit,
		StatusSourceRunID:        status.SourceRunID,
		StatusSourceRunURL:       status.SourceRunURL,
		StatusBlockingCategories: status.BlockingCategories,
		PagesStatus:              pages.Status, PagesVisibility: pages.Visibility,
		PagesBuildType:           pages.BuildType,
		PagesSourceCommit:        pages.SourceCommit,
		PagesSourceRunID:         pages.SourceRunID,
		PagesSourceRunURL:        pages.SourceRunURL,
		PagesRawResponseIncluded: pages.RawResponseIncluded,
		PagesSecretsIncluded:     pages.SecretsIncluded,
		BadgeSchemaVersion:       badge.SchemaVersion,
		BadgeLabel:               badge.Label,
		BadgeMessage:             badge.Message,
		BadgeColor:               badge.Color,
	}
}
