package deploypolicy

func runtimeCDDocPhrases() []string {
	return []string{
		"Public Export Contract",
		"RIID-4835",
		"Public Surface Scan",
		"RIID-4836",
		"Public Config Key Minimization",
		"RIID-4839",
		"Public Sensitive Surface Guard",
		"RIID-4842",
		"Public Operational Detail Minimization",
		"RIID-4853",
		"RIID-4855",
		"operator/environment gated",
		"infra is the same ownership rule",
		"Image values are deliberately not in that public export set",
	}
}

func runtimeCDBoundaryPhrases() []string {
	return []string{
		"RIID-4835",
		"RIID-4839",
		"RIID-4842",
		"RIID-4845",
		"RIID-4853",
		"aggregate deploy/smoke pass-fail status without live payload values",
		"are not public hand-off artifacts",
	}
}
