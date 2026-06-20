package deploypolicy

func deployWorkflowRequiredPhrases() []string {
	return []string{
		"echo \"::add-mask::$aws_account_id\"",
		"echo \"::add-mask::$registry\"",
		"echo \"::add-mask::$image_uri\"",
		"echo \"::add-mask::$current_task_definition\"",
		"echo \"::add-mask::$next_task_definition\"",
		"redacted=(",
		"echo \"::add-mask::${!name}\"",
		"if [ \"$image_tag\" = \"latest\" ]",
		"workflow_dispatch",
		"deployment_target",
		"ai-agent-development",
		"ai-agent-production",
		"testnet|development|production",
		"tags:",
		"- \"v*\"",
		"TESTNET_BASE_URL: ${{ vars.RIIDO_AI_SERVER_TESTNET_BASE_URL }}",
		"SMOKE_TOKEN_CONFIGURED: ${{ secrets.RIIDO_AI_SERVER_TESTNET_TOKEN != '' }}",
		"if [ \"${SMOKE_TOKEN_CONFIGURED:-false}\" != \"true\" ]",
		"FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: true",
		"CODEDEPLOY_APPLICATION: ${{ vars.RIIDO_AI_SERVER_CODEDEPLOY_APPLICATION }}",
		"CODEDEPLOY_DEPLOYMENT_GROUP: ${{ vars.RIIDO_AI_SERVER_CODEDEPLOY_DEPLOYMENT_GROUP }}",
	}
}
