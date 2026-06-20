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
