package deploypolicy

type currentStrategy struct {
	ID            string   `json:"id"`
	Status        string   `json:"status"`
	CDOwner       string   `json:"cd_owner"`
	TopologyOwner string   `json:"topology_owner"`
	Workflow      string   `json:"workflow"`
	Allowed       []string `json:"allowed_actions"`
}

type optionalWorkflowMode struct {
	ID               string   `json:"id"`
	Status           string   `json:"status"`
	CDOwner          string   `json:"cd_owner"`
	TopologyOwner    string   `json:"topology_owner"`
	Workflow         string   `json:"workflow"`
	ActivationInputs []string `json:"activation_inputs"`
	Allowed          []string `json:"allowed_actions"`
	MustNotOwn       []string `json:"must_not_own"`
}

type futureStrategy struct {
	ID                 string   `json:"id"`
	Status             string   `json:"status"`
	CDOwner            string   `json:"cd_owner"`
	TopologyOwner      string   `json:"topology_owner"`
	ControlPlaneMayOwn []string `json:"control_plane_may_own"`
	InfraMustOwn       []string `json:"infra_must_own"`
}

type codeDeployActivationGate struct {
	RiidoTask              string   `json:"riido_task"`
	CanonicalOwner         string   `json:"canonical_owner"`
	InfraAwarenessOwner    string   `json:"infra_awareness_owner"`
	Status                 string   `json:"status"`
	Rule                   string   `json:"rule"`
	ActivationRequirements []string `json:"activation_requirements"`
	PublicRepoMayKeep      []string `json:"public_repo_may_keep"`
	PublicRepoMustNotKeep  []string `json:"public_repo_must_not_keep"`
	InfraMustKnow          []string `json:"infra_must_know"`
}
