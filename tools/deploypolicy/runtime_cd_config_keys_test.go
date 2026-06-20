package deploypolicy

func expectedSecretKeys() []string {
	return []string{"RIIDO_AI_SERVER_DEPLOY_ROLE_ARN", "RIIDO_AI_SERVER_TESTNET_TOKEN"}
}

func expectedRequiredVars() []string {
	return []string{
		"RIIDO_AI_SERVER_AWS_REGION",
		"RIIDO_AI_SERVER_ECR_REPOSITORY",
		"RIIDO_AI_SERVER_ECS_CLUSTER",
		"RIIDO_AI_SERVER_ECS_SERVICE",
		"RIIDO_AI_SERVER_ECS_CONTAINER_NAME",
		"RIIDO_AI_SERVER_TESTNET_BASE_URL",
	}
}

func expectedOptionalVars() []string {
	return []string{
		"RIIDO_AI_SERVER_TESTNET_WORKSPACE_ID",
		"RIIDO_AI_SERVER_CODEDEPLOY_APPLICATION",
		"RIIDO_AI_SERVER_CODEDEPLOY_DEPLOYMENT_GROUP",
	}
}

func expectedAllVars() []string {
	return append(append([]string{}, expectedRequiredVars()...), expectedOptionalVars()...)
}

func expectedCDKeys() []string {
	return append(append([]string{}, expectedSecretKeys()...), expectedAllVars()...)
}
