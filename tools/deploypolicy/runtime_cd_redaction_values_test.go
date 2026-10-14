package deploypolicy

func forbiddenRuntimeCDPublicValues() []string {
	return []string{
		"live URL values",
		"AWS account IDs",
		"ARNs",
		"CodeDeploy deployment IDs",
		"image digests or image URIs",
		"workflow_dispatch input values carrying live URLs",
		"GitHub step outputs carrying live deployment values",
		"task definition JSON",
		"CodeDeploy AppSpec JSON",
		"smoke response payloads",
		"Terraform state",
	}
}
