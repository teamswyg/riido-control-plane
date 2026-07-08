package main

func validStatusJSON() string {
	return `{
		"overall":"degraded",
		"visibility":"public_aggregate",
		"source_commit":"abc123",
		"source_run_id":"456",
		"source_run_url":"https://github.com/teamswyg/riido-control-plane/actions/runs/456",
		"raw_logs_included":false,
		"secrets_included":false,
		"endpoint_details":"redacted",
		"blocking_categories":[{"category":"usability","partial_count":1,"stale_partial_count":1}]
	}`
}

func validPagesJSON() string {
	return `{
		"status":"published",
		"visibility":"public_repository",
		"build_type":"workflow",
		"source_commit":"abc123",
		"source_run_id":"456",
		"source_run_url":"https://github.com/teamswyg/riido-control-plane/actions/runs/456",
		"raw_response_included":false,
		"secrets_included":false
	}`
}

func validBadgeJSON() string {
	return `{
		"schemaVersion":1,
		"label":"riido qa",
		"message":"degraded / 1 categories",
		"color":"orange"
	}`
}
