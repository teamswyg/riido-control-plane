package main

type parityClaim struct {
	AllSourceStepsMapped                              bool `json:"all_source_steps_mapped"`
	RequiredAdapterCount                              int  `json:"required_adapter_count"`
	DependencyFailurePreservesRedactedArtifactAttempt bool `json:"dependency_failure_preserves_redacted_artifact_attempt"`
	LaterOnSuccessTestSkipsAfterPriorFailure          bool `json:"later_on_success_test_skips_after_prior_failure"`
	TestCommandExact                                  bool `json:"test_command_exact"`
	DependencyCommandExact                            bool `json:"dependency_command_exact"`
	SourceWorkflowEdited                              bool `json:"source_workflow_edited"`
	SourceWorkflowExecutedByThisSlice                 bool `json:"source_workflow_executed_by_this_slice"`
}

type authority struct {
	WorkflowRetirementAuthorized            bool   `json:"workflow_retirement_authorized"`
	WorkflowFileEffect                      string `json:"workflow_file_effect"`
	AuthJWTPEPEffect                        string `json:"auth_jwt_pep_effect"`
	AWSTerraformDeploymentEffect            string `json:"aws_terraform_deployment_effect"`
	RuntimeEffect                           string `json:"runtime_effect"`
	FixedCostResourceEffect                 string `json:"fixed_cost_resource_effect"`
	RequiresSeparateOwnerReviewedRetirement bool   `json:"requires_separate_owner_reviewed_retirement"`
}

type rollback struct {
	BaselineWorkflowPreserved                bool   `json:"baseline_workflow_preserved"`
	Method                                   string `json:"method"`
	BlockedToPassedPrivatePolicyPairRequired bool   `json:"blocked_to_passed_private_policy_pair_required"`
}

type classification struct {
	Code         string   `json:"code"`
	Meaning      string   `json:"meaning"`
	DoesNotClaim []string `json:"does_not_claim"`
}
