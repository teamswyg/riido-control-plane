package deploypolicy

func publicRedactionDocAssertions(f redactionFixture) []docAssertion {
	return []docAssertion{
		{f.ClientAPI, "not from a manual"},
		{f.ClientAPI, "The workflow masks both values"},
		{f.Readme, "live URL, AWS account id, ARN, image digest"},
		{f.Boundary, "task-definition ARNs, image digests, live workflow run URLs"},
		{f.Boundary, "must not upload"},
		{f.Boundary, "deployment artifacts from the live"},
		{f.Boundary, "CodeDeploy Handoff"},
		{f.Boundary, "runtime artifact CD execution still belongs"},
		{f.Domain, "live URLs, task-definition ARNs"},
		{f.Domain, "CodeDeploy blue/green"},
		{f.Migration, "RIID-4812 tightens that public boundary"},
		{f.Migration, "RIID-4814"},
		{f.Migration, "RIID-4815"},
		{f.Migration, "RIID-4822"},
		{f.Migration, "RIID-4825"},
		{f.Migration, "RIID-4835"},
	}
}

func publicRedactionBodies(f redactionFixture) map[string]string {
	return map[string]string{
		"README.md":                            f.Readme,
		"runtime-deployment-boundary.md":       f.Boundary,
		"saas-control-plane.md":                f.Domain,
		"ai-agent-client-api.md":               f.ClientAPI,
		"api-client-delivery.md":               f.ClientDelivery,
		"control-plane.md":                     f.Migration,
		"tools/reactquerygen/main.go":          f.Generator,
		"web/generated/aiAgentClient.ts":       f.GeneratedClient,
		"web/generated/aiAgentClient.react.ts": f.GeneratedReactClient,
	}
}
