package deploypolicy

func generatedClientDocPhrases() []string {
	return []string{
		"create_pr=false",
		"requiring `riido-client` write credentials",
		"This failure is intentional only for `create_pr=true`",
		"Riido `branchName`",
		"secret gates",
	}
}

func generatedClientMigrationPhrases() []string {
	return []string{
		"RIID-4899",
		"legacy delivery workflow",
		"raw `RIIDO_CLIENT_DELIVERY_TOKEN`",
		"synthesize `react-query-*` branch names",
	}
}
