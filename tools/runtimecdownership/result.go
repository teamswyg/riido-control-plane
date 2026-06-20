package main

type verifyResult struct {
	Strategies      int
	PublicPolicies  int
	PublicGuards    int
	ForbiddenItems  int
	InfraLinks      int
	LoopFields      int
	StableKeyCount  int
	WorkflowCount   int
	HardeningCount  int
	SupersedesCount int
}
