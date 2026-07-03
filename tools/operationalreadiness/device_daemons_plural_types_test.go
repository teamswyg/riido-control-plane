package main

type deviceDaemonsPluralEvidence struct {
	Redacted    bool                            `json:"redacted"`
	Source      deviceDaemonsPluralSource       `json:"source"`
	Deployments []deviceDaemonsPluralDeployment `json:"deployments"`
	Assertions  deviceDaemonsPluralAssertions   `json:"assertions"`
}

type deviceDaemonsPluralSource struct {
	NotionComment string `json:"notion_comment"`
}

type deviceDaemonsPluralDeployment struct {
	Environment   string `json:"environment"`
	DeploymentRun string `json:"deployment_run"`
	HealthzStatus int    `json:"healthz_status"`
	ReadyzStatus  int    `json:"readyz_status"`
	V1Status      int    `json:"v1_status"`
	V2Status      int    `json:"v2_status"`
}

type deviceDaemonsPluralAssertions struct {
	Not404                    bool `json:"not_404"`
	UnauthenticatedProbeOnly  bool `json:"unauthenticated_probe_only"`
	RawResponseBodiesIncluded bool `json:"raw_response_bodies_included"`
	SecretsIncluded           bool `json:"secrets_included"`
}
