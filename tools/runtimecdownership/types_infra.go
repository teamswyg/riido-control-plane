package main

type infraConsumes struct {
	Repo       string   `json:"repo"`
	Paths      []string `json:"paths"`
	LocalScope string   `json:"local_scope"`
}

type infraVisibilityPolicy struct {
	Repo        string   `json:"repo"`
	MustKnow    []string `json:"must_know"`
	MustNotFrom []string `json:"must_not_receive_from_public_workflow"`
}

type infraTopologyContract struct {
	RiidoTask      string   `json:"riido_task"`
	Repo           string   `json:"repo"`
	WorkUnit       string   `json:"terraform_work_unit"`
	RequiredOutput []string `json:"required_outputs"`
	MustNotConsume []string `json:"control_plane_must_not_consume"`
}

type dependencyDirection struct {
	TopDown  string `json:"top_down"`
	BottomUp string `json:"bottom_up"`
}

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
