package deploypolicy

func requiredInfraAwarenessPaths() []string {
	return []string{
		"docs/architecture/terraform-authoring.md",
		"deploy/work-units/riid-4825-control-plane-cd-ownership-remodel.riido.json",
		"deploy/work-units/riid-4833-control-plane-cd-public-redaction-hardening.riido.json",
		"deploy/work-units/riid-4835-control-plane-cd-public-export-contract.riido.json",
		"deploy/work-units/riid-4836-control-plane-cd-public-surface-redaction-scan.riido.json",
		"deploy/work-units/riid-4837-cd-ownership-final-guard-public-surface-minimization.riido.json",
		"deploy/work-units/riid-4839-cd-public-config-key-minimization.riido.json",
		"deploy/work-units/riid-4854-control-plane-cd-public-minimization-awareness-no-diff.riido.json",
		"deploy/work-units/riid-4856-codedeploy-activation-gate-awareness-no-diff.riido.json",
		"deploy/work-units/riid-4860-control-plane-cd-ownership-awareness-guard.riido.json",
	}
}
