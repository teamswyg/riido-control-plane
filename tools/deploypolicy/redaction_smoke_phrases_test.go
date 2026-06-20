package deploypolicy

func smokeWorkflowRequiredPhrases() []string {
	return []string{
		"profile-thumbnails/uploads",
		"profile_thumbnail_url",
		"TESTNET_BASE_URL: ${{ vars.RIIDO_AI_SERVER_TESTNET_BASE_URL }}",
		"echo \"::add-mask::$TESTNET_BASE_URL\"",
		"echo \"::add-mask::$TESTNET_TOKEN\"",
		"umask 077",
		"trap 'rm -f \"$replay\"' EXIT",
	}
}
