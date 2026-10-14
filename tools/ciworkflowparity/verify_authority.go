package main

func verifyRunner(repoRoot string, document manifest) bool {
	runner := document.Runner
	if runner.Provider != "riido-ci" || runner.Revision != "4795cdebfcf4bccaa71ec2344368ad9adf6b1974" ||
		runner.Pipeline != "pipelines/control-plane.local-self-check.riido.json" ||
		runner.PipelineID != "control-plane-local-self-check" || runner.Visibility != "private" ||
		runner.GitHubTokenRequired || runner.CloudCredentialsRequired {
		return false
	}
	raw, err := readRootFile(repoRoot, "tools/riido-ci-local")
	want := []string{
		runner.Revision, "riido-ci checkout must be clean", "-u AWS_ACCESS_KEY_ID",
		"-u AWS_SECRET_ACCESS_KEY", "-u GITHUB_TOKEN", "-u GH_TOKEN",
		"riido-ci evidence must be mode 0600",
	}
	return err == nil && containsAll(string(raw), want)
}

func verifyAuthority(document manifest) bool {
	value := document.Authority
	return !value.WorkflowRetirementAuthorized && value.WorkflowFileEffect == "none" &&
		value.AuthJWTPEPEffect == "none" && value.AWSTerraformDeploymentEffect == "none" &&
		value.RuntimeEffect == "none" && value.FixedCostResourceEffect == "none" &&
		value.RequiresSeparateOwnerReviewedRetirement
}

func verifyRollback(document manifest) bool {
	value := document.Rollback
	return value.BaselineWorkflowPreserved &&
		value.Method == "retain_exact_baseline_until_separate_retirement_parity_review" &&
		value.BlockedToPassedPrivatePolicyPairRequired
}
