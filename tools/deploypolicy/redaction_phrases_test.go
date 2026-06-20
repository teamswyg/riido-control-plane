package deploypolicy

func deployMaskedKeys() []string {
	return []string{
		"AWS_REGION",
		"ECR_REPOSITORY",
		"ECS_CLUSTER",
		"ECS_SERVICE",
		"ECS_CONTAINER_NAME",
		"CODEDEPLOY_APPLICATION",
		"CODEDEPLOY_DEPLOYMENT_GROUP",
		"TESTNET_BASE_URL",
		"TESTNET_WORKSPACE_ID",
	}
}

func workflowForbiddenPhrases() []string {
	return []string{
		"inputs.base_url",
		"description: \"Optional AI Agent testnet base URL",
		"GITHUB_OUTPUT",
		"latest\" >>",
		"latest' >>",
		"task-definition.next.json\" >>",
		"task-definition.current.json\" >>",
		"appspec.json\" >>",
		"deployment-id\" >>",
		"(.devices | length) >= 1",
	}
}
