package main

type redactionPolicy struct {
	Workflows      []string `json:"public_repo_workflows"`
	MayCommit      []string `json:"public_repo_may_commit"`
	ShouldMinimize []string `json:"public_repo_should_minimize"`
	MustNot        []string `json:"public_repo_must_not_commit_or_upload"`
}

type handoffPolicy struct {
	Scope              string   `json:"scope"`
	AllowedStorage     []string `json:"allowed_storage"`
	RequiredCleanup    []string `json:"required_cleanup"`
	ForbiddenMechanism []string `json:"forbidden_mechanisms"`
}

type publicExportContract struct {
	RiidoTask              string   `json:"riido_task"`
	CanonicalOwner         string   `json:"canonical_owner"`
	InfraAwarenessOwner    string   `json:"infra_awareness_owner"`
	AllowedPublicExports   []string `json:"allowed_public_exports"`
	ForbiddenPublicExports []string `json:"forbidden_public_exports"`
	InfraMustConsumeOnly   []string `json:"infra_must_consume_only"`
	WorkflowMustNotUse     []string `json:"workflow_must_not_use"`
}

type publicSurfaceScanContract struct {
	RiidoTask                  string   `json:"riido_task"`
	CanonicalOwner             string   `json:"canonical_owner"`
	InfraAwarenessOwner        string   `json:"infra_awareness_owner"`
	ScopePaths                 []string `json:"scope_paths"`
	ForbiddenLiterals          []string `json:"forbidden_literals"`
	ForbiddenRegexes           []string `json:"forbidden_regexes"`
	WorkflowForbiddenMechanism []string `json:"workflow_forbidden_mechanisms"`
	AllowedPublicSurface       []string `json:"allowed_public_surface"`
	InfraMustTreatScanAs       string   `json:"infra_must_treat_scan_as"`
}

type operationalDetailMinimization struct {
	RiidoTask             string   `json:"riido_task"`
	CanonicalOwner        string   `json:"canonical_owner"`
	InfraAwarenessOwner   string   `json:"infra_awareness_owner"`
	Rule                  string   `json:"rule"`
	PublicRepoMayKeep     []string `json:"public_repo_may_keep"`
	PublicRepoShouldAvoid []string `json:"public_repo_should_avoid"`
	InfraMustKnow         []string `json:"infra_must_know"`
}
